package expr

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// tokenKind is one lexical token kind. The lexer never skips interior
// whitespace (only Parse's own leading/trailing trim, mirroring
// 1.6's "ws? ident ws?" — D-1.6.5): any whitespace character found
// mid-expression is simply not part of any legal token, so it
// surfaces as an "unexpected character" syntax error at the point it
// appears — the same shape 1.6 gave "{{a b}}" (F10), now generalised
// to the whole grammar rather than reimplemented for it.
type tokenKind int

const (
	tokEOF tokenKind = iota
	tokIdent
	tokString
	tokNumber
	tokDot
	tokLParen
	tokRParen
	tokComma
)

type token struct {
	kind       tokenKind
	text       string // the token's own source text (identifier name, quoted string incl. quotes, number literal)
	start, end int    // byte offsets into the parser's src
}

// parser is a hand-written recursive-descent parser (AC1, AD-9: no
// generator, no third-party dependency, no regex engine standing in
// for the grammar) over one expression's source text — the trimmed
// content between "{{" and "}}", never the surrounding braces
// themselves.
type parser struct {
	src   string
	pos   int
	depth int // nested function-call depth; see maxCallDepth
}

// maxCallDepth bounds the parsePrimary -> parseIdentLed -> parseCall
// -> parsePrimary recursion cycle (QA Finding 3, Blocker). It is a
// declared property of the parser, stated once — not fitted to any
// one attack input: no legitimate .folio expression nests function
// calls more than a handful deep (AC3's own example,
// "formatNumber(sum(t.amount), ...)", nests two), so 64 leaves
// generous headroom for real authoring while keeping the parser's own
// goroutine stack usage a small, fixed multiple of 64 frames —
// nowhere near the runtime's fatal stack-overflow threshold, however
// large the input. Without a bound, a single text element whose value
// is "{{" + "a(" repeated hundreds of thousands of times drives
// unbounded recursion into a goroutine stack exhaustion: an
// unrecoverable runtime "fatal error: stack overflow" (a throw, not a
// panic — recover() cannot catch it), reachable from author-supplied
// content at both load (folio.ParseTemplate) and render
// (bind.BindText). Exceeding the bound is instead an ordinary located
// syntax error, exactly like every other rejected form.
const maxCallDepth = 64

// Parse parses src — the raw content between "{{" and "}}", NOT yet
// trimmed — against AC3's grammar: a bare dotted path, a function call
// over comma-separated arguments, a double-quoted string literal, or a
// number literal, nesting to any depth. It performs no semantic check
// at all (no arity, no unknown-function check, no literal-kind check —
// Check, check.go, does all of that): Parse is grammar only, so a
// syntax error and a semantic error stay two separably-testable
// properties (R3: "syntax and arity at load; execution at
// evaluation" — Parse is the syntax half).
//
// AC19's cases all surface here: unbalanced parenthesis, trailing
// comma, unterminated string literal, empty expression, a bare
// operator ("{{a + b}}" — there are no operators in this grammar), and
// a path segment starting with a digit (rejected because the lexer's
// own ident rule, mirroring 1.6's [A-Za-z_][A-Za-z0-9_]*, never
// produces an identifier token starting with a digit — a numeric
// token appears instead, and the parser rejects it wherever an
// identifier was expected).
func Parse(src string) (Expr, error) {
	trimmed := strings.TrimSpace(src)
	if trimmed == "" {
		return nil, fmt.Errorf("empty expression")
	}
	p := &parser{src: trimmed}
	e, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	// p.src was already trimmed of leading/trailing whitespace above,
	// and the grammar has no operators (F11), so anything left over —
	// a bare operator like "a + b", an extra ")", stray text after a
	// complete call — is trailing garbage: AC19's "bare operator" case.
	p.skipWS()
	if p.pos != len(p.src) {
		return nil, fmt.Errorf("unexpected trailing content %q at position %d", p.src[p.pos:], p.pos)
	}
	return e, nil
}

func (p *parser) peekByte() (byte, bool) {
	if p.pos >= len(p.src) {
		return 0, false
	}
	return p.src[p.pos], true
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isIdentCont(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// skipWS advances past ASCII spaces and tabs. Called only at specific
// structural boundaries inside a function call's argument list — right
// after "(", right after ",", and right before the closing ")" — so
// that "formatNumber(t.amount, \"#,##0.00\")" (the canonical golden's
// own spelling, worked-example.json:19, with a space after the comma)
// parses, while the bare-path grammar stays exactly as strict as 1.6's
// (D-1.6.5): "a . b" and "a b" are never tolerated, because
// parseIdentLed's own dot-continuation check (below) never calls this.
func (p *parser) skipWS() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

// nextToken lexes exactly one token starting at p.pos, without
// consuming it (callers advance p.pos themselves via the returned
// token's end, through the parse* methods below).
func (p *parser) nextToken() (token, error) {
	c, ok := p.peekByte()
	if !ok {
		return token{kind: tokEOF, start: p.pos, end: p.pos}, nil
	}
	switch {
	case c == '.':
		return token{kind: tokDot, text: ".", start: p.pos, end: p.pos + 1}, nil
	case c == '(':
		return token{kind: tokLParen, text: "(", start: p.pos, end: p.pos + 1}, nil
	case c == ')':
		return token{kind: tokRParen, text: ")", start: p.pos, end: p.pos + 1}, nil
	case c == ',':
		return token{kind: tokComma, text: ",", start: p.pos, end: p.pos + 1}, nil
	case c == '"':
		return p.lexString()
	case isIdentStart(c):
		return p.lexIdent()
	case isDigit(c) || c == '-':
		return p.lexNumber()
	default:
		// QA Finding 13 (Minor): decode a RUNE, not a byte — string(c)
		// on a lone byte of a multi-byte UTF-8 sequence names a
		// character that is not in the source at all (Parse("é") used
		// to report "Ã", not "é"). On a Thai bank-statement product
		// this is the common case, not the exotic one.
		r, _ := utf8.DecodeRuneInString(p.src[p.pos:])
		if r == '\n' || r == '\r' {
			// A newline is legal inside a JSON string but never inside
			// this grammar's own token stream (skipWS only tolerates
			// ' '/'\t', deliberately — D-1.6.5). Named explicitly
			// rather than left to read as an arbitrary "unexpected
			// character".
			return token{}, fmt.Errorf("unexpected newline at position %d (only spaces and tabs are permitted here)", p.pos)
		}
		return token{}, fmt.Errorf("unexpected character %q at position %d", r, p.pos)
	}
}

func (p *parser) lexIdent() (token, error) {
	start := p.pos
	i := start
	for i < len(p.src) && isIdentCont(p.src[i]) {
		i++
	}
	return token{kind: tokIdent, text: p.src[start:i], start: start, end: i}, nil
}

// lexString lexes a double-quoted string literal (AC3/AC19). No
// escape sequences: `.folio` string bindings and formatNumber/
// formatDate patterns never need one (D-1.4.1's own patterns are plain
// ASCII), and an unclosed quote before the input ends is AC19's
// "unterminated string literal" case. A quote therefore cannot appear
// INSIDE a string literal at all — including an escaped one — so an
// author who writes \" gets a diagnostic that says so directly (QA
// Finding 11, Minor), rather than the lexer silently stopping at the
// bogus closing quote and letting a LATER, unrelated stage (trailing
// content, or "expected ',' or ')'") misreport the actual cause. Any
// OTHER backslash is untouched: it survives literally into Value/Raw,
// exactly as before this fix.
func (p *parser) lexString() (token, error) {
	start := p.pos
	i := start + 1 // past the opening quote
	for i < len(p.src) && p.src[i] != '"' {
		if p.src[i] == '\\' && i+1 < len(p.src) && p.src[i+1] == '"' {
			return token{}, fmt.Errorf(`string literals do not support escape sequences (found \" at position %d): a quote cannot appear inside a string literal`, i)
		}
		i++
	}
	if i >= len(p.src) {
		return token{}, fmt.Errorf("unterminated string literal starting at position %d", start)
	}
	end := i + 1 // past the closing quote
	return token{kind: tokString, text: p.src[start:end], start: start, end: end}, nil
}

// lexNumber lexes a JSON number literal: -? digits (.digits)?
// ([eE][+-]?digits)? — the same shape
// internal/template.SplitJSONNumber accepts (NewDecimal, decimal.go,
// is what eventually consumes it). The integer part follows JSON's own
// grammar exactly (RFC 8259: "0" or a non-zero digit followed by more
// digits — no other leading zero is legal): "01"/"007"/"-01" are
// rejected here (QA Finding 7, Major), rather than being silently
// accepted by the parser and then silently normalised by
// SplitJSONNumber downstream (which explicitly trusts encoding/json's
// own upstream grammar check and performs none of its own), which is
// what ast.go's own doc comment on NumberLit already claims this
// grammar is.
func (p *parser) lexNumber() (token, error) {
	start := p.pos
	i := start
	if i < len(p.src) && p.src[i] == '-' {
		i++
	}
	digitsStart := i
	for i < len(p.src) && isDigit(p.src[i]) {
		i++
	}
	if i == digitsStart {
		return token{}, fmt.Errorf("invalid number literal at position %d", start)
	}
	if i-digitsStart > 1 && p.src[digitsStart] == '0' {
		return token{}, fmt.Errorf("invalid number literal (leading zero not allowed) at position %d", start)
	}
	if i < len(p.src) && p.src[i] == '.' {
		i++
		fracStart := i
		for i < len(p.src) && isDigit(p.src[i]) {
			i++
		}
		if i == fracStart {
			return token{}, fmt.Errorf("invalid number literal (digits expected after '.') at position %d", start)
		}
	}
	if i < len(p.src) && (p.src[i] == 'e' || p.src[i] == 'E') {
		i++
		if i < len(p.src) && (p.src[i] == '+' || p.src[i] == '-') {
			i++
		}
		expStart := i
		for i < len(p.src) && isDigit(p.src[i]) {
			i++
		}
		if i == expStart {
			return token{}, fmt.Errorf("invalid number literal (digits expected in exponent) at position %d", start)
		}
	}
	return token{kind: tokNumber, text: p.src[start:i], start: start, end: i}, nil
}

// parsePrimary parses one full primary expression: a path, a call, a
// string literal, or a number literal (AC3). It is also the entry
// point parseArgs uses for each argument, which is how nesting works
// (AC3: formatNumber(sum(t.amount), "#,##0.00") — the first argument
// is itself a full primary expression, a CallExpr).
func (p *parser) parsePrimary() (Expr, error) {
	tok, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	switch tok.kind {
	case tokString:
		p.pos = tok.end
		return &StringLit{Value: tok.text[1 : len(tok.text)-1], Raw: tok.text}, nil
	case tokNumber:
		p.pos = tok.end
		return &NumberLit{Literal: tok.text, Raw: tok.text}, nil
	case tokIdent:
		return p.parseIdentLed(tok)
	case tokEOF:
		return nil, fmt.Errorf("unexpected end of expression at position %d", tok.start)
	default:
		return nil, fmt.Errorf("unexpected token %q at position %d", tok.text, tok.start)
	}
}

// parseIdentLed continues parsing after an identifier has already been
// lexed: either it is immediately followed by "(" (a call — AC3) or it
// begins a dotted path (AC3, D-1.6.5's grammar preserved verbatim).
// Function calls are never qualified by a leading path (no
// "a.b.sum(x)" — FR18's eight names are always bare), so this
// disambiguation happens exactly once, at the first identifier.
func (p *parser) parseIdentLed(first token) (Expr, error) {
	p.pos = first.end
	if c, ok := p.peekByte(); ok && c == '(' {
		return p.parseCall(first)
	}
	segments := []string{first.text}
	for {
		c, ok := p.peekByte()
		if !ok || c != '.' {
			break
		}
		dotPos := p.pos
		p.pos++ // consume "."
		next, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if next.kind != tokIdent {
			got := next.text
			if next.kind == tokEOF {
				// QA Finding 13 (Minor): next.text is "" for tokEOF,
				// so this used to print `got ""` — technically
				// accurate, practically unreadable. Parse("a.") now
				// reads "…got end of expression", like the analogous
				// case in parsePrimary.
				got = "end of expression"
			} else {
				got = fmt.Sprintf("%q", got)
			}
			return nil, fmt.Errorf("expected an identifier after '.' at position %d, got %s", dotPos+1, got)
		}
		p.pos = next.end
		segments = append(segments, next.text)
	}
	return &PathExpr{Segments: segments, Raw: p.src[first.start:p.pos]}, nil
}

// parseCall parses a function call's "(" args ")" tail, given that
// name has already been lexed and the next byte is known to be "(".
// It is the sole recursion point for nested calls (each argument is
// parsed via parsePrimary, which may itself be another call), so it
// is where maxCallDepth is enforced (QA Finding 3, Blocker).
func (p *parser) parseCall(name token) (Expr, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxCallDepth {
		return nil, fmt.Errorf("expression nests function calls too deeply (max %d) at position %d", maxCallDepth, name.start)
	}
	p.pos++ // consume "("
	p.skipWS()
	var args []Expr
	if c, ok := p.peekByte(); !ok || c != ')' {
		for {
			arg, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			p.skipWS()
			c, ok := p.peekByte()
			if !ok {
				// QA Finding 13 (Minor): this message used to carry no
				// position at all, unlike every other message in this
				// file — Parse("sum(a") gave no hint WHERE the missing
				// ')' was expected. Now names both the opening "(" and
				// the point reached at end of input.
				return nil, fmt.Errorf("unbalanced parenthesis: %q is missing its closing ')' — opened at position %d, reached end of input at position %d", name.text, name.start, p.pos)
			}
			if c == ',' {
				p.pos++ // consume ","
				p.skipWS()
				continue
			}
			if c == ')' {
				break
			}
			return nil, fmt.Errorf("expected ',' or ')' at position %d, got %q", p.pos, string(c))
		}
	}
	// QA Finding 13 (Minor): a second, independent peekByte check used
	// to live here, expecting to re-report "unbalanced parenthesis" —
	// but the block above only ever completes with the next byte
	// already confirmed to be ')' (either the "if" condition above was
	// false, meaning peekByte already saw ')', or the loop's own
	// "if c == ')' { break }" did). That made the second check
	// unreachable dead code: it could never observe anything other
	// than ')'. Deleted; ")" is consumed directly.
	p.pos++ // consume ")"
	return &CallExpr{Name: name.text, Args: args, Raw: p.src[name.start:p.pos]}, nil
}
