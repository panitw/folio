// This file implements text-element "{{ }}" binding (D-1.6.5, replaced
// at Story 3.2 by the expression language, D-1.6.5/D-3.2.1: "this
// story's parser REPLACES 1.6's matcher — deleted, not kept
// alongside"). Scope (AC20, unchanged since 1.6): callers apply
// BindText only to text-element `value` interpolation.
package bind

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/panitw/folio/folio-go/internal/expr"
)

// Substitution records one placeholder's contribution to BindTextSpans'
// output: which data path produced it, and which part of the result it
// occupies.
//
// Start and End are RUNE indices into the returned string, half-open.
// Rune indices, because every position the breaking stage works in is a
// rune index — AD-25's break positions and GlyphInfo.Cluster alike
// (D-2.3.2) — and a byte offset that happened to be equal for ASCII
// would diverge silently on the first Thai character.
//
// A placeholder that resolved to JSON null contributes an EMPTY span
// (Start == End) rather than no span at all. It is reported because the
// caller asked what each path produced, and "nothing, at this position"
// is the honest answer; it can never affect breaking, since an empty
// span has no interior.
//
// THIS TYPE CARRIES NO OPINION ABOUT LINE BREAKING, and internal/bind
// has none. It reports what it substituted and where. Whether a span is
// unbreakable is the DOCUMENT's declaration, matched against Path by
// the caller, and handed to internal/text as a parameter (D-000.16:
// "the signal rides on the value, never through an import").
type Substitution struct {
	// Path is the bare dotted path exactly as authored, including a
	// leading "params" segment when it resolved against the params
	// root — or, for an expression that is not a bare path (a function
	// call, Story 3.2), the expression's own raw source text.
	Path string

	// Start, End are rune indices into BindTextSpans' returned string.
	Start, End int
}

// AD-4's page-number slots (D-1.6.5, AC18) are owned by Story 2.7,
// never resolved from data and never an error at 1.6. A data document
// containing a top-level "page" key must not change what {{page}}
// renders (AC19) — this reservation is checked BEFORE any data lookup,
// or any parse attempt (AC4, Story 3.2), so it can never be shadowed.
//
// Resolve (below) gets this for free from expr.ScanPlaceholders'
// Placeholder.Reserved field: reservation is decided exactly once, by
// internal/expr.IsReserved (AD-13: one definition, not two)
// — this file no longer needs, and no longer keeps, a local alias.
// (QA Finding 5, Major: this file previously hand-rolled a byte-
// identical second "{{"/"}}" search rather than calling
// ScanPlaceholders, which its own doc comment claimed nothing did.)

// declaredResolutionRoots is Story 2.7's AC3 closure, in the shape
// internal/template/closedsets.go and declaredEpic2GateObligations
// already use for "this absence must stay absent" (D-2.5.1): the
// DECLARED set of resolution roots — the values BindTextSpans passes as
// lookupBound's rootName argument — compared, both directions, against
// the OBSERVED set internal/bind_arch_test.go's
// TestBindResolutionRootsAreClosed collects by scanning this file's own
// lookupBound call sites.
//
// A "page" NAMESPACE is precisely a THIRD entry here (BindText's own
// doc comment below carries the fence, verbatim: "conflating the two
// is how 'page' would eventually acquire a namespace, which AD-4
// forbids forever" — cited by enclosing function, not by line number:
// this story's review, Finding 7, found the line-number pointer here
// already stale once, from this story's own edit). Adding
// one is a one-line, VISIBLE diff to this list, and the guard catches a
// third lookupBound call site whether or not this list is edited to
// match it.
//
// Story 3.1 / D-3.1.1 widens this to a THIRD entry, "row": the row
// scope AD-11 declares, authorised as a resolution root by the same
// mechanism D-1.7.4 used for "params" — a root, not a reserved token.
// "row" is the ROOT CLASS, never the author's own alias spelling (see
// Scope's doc comment, scope.go) — the guard's own Fatalf on a
// non-literal rootName is exactly why the alias can never be passed
// here directly. TestBindResolutionRootsClosureRedProof was re-run
// after this widening and still reddens (D-3.1.1), proving AD-4's
// "page" fence survived it: "page"/"pages" remain reserved tokens,
// resolved from neither root.
var declaredResolutionRoots = []string{"data", "params", "row"}

// BindText resolves every "{{…}}" placeholder in text against data and
// params (Story 3.2's expression language, D-1.6.5, D-1.7.4), scoped to
// one text element (elementID names it in every error). "{{page}}"
// and "{{pages}}" pass through unchanged (AC4).
//
// AD-14's three cases (AC12/AC13, Story 3.2's if() ruling): an absent
// path is an error naming both the data path and elementID; an
// explicit JSON null renders as empty and is not an error; a value of
// the wrong kind for a text binding (anything but a string) is an
// error, never coerced.
//
// "params" is a SECOND resolution root, not a reserved token (Story
// 1.7, D-1.7.4, verbatim): "params is a namespace — a second
// resolution root resolved at bind time, alongside the data root." A
// placeholder whose FIRST path segment is "params" ALWAYS resolves
// against params, never against data — regardless of whether data
// itself happens to carry a top-level "params" key: that key is
// ordinary, unreachable caller JSON. This is deliberately NOT
// implemented by extending expr.IsReserved (page/pages are
// reserved whole TOKENS, resolved from neither root and owned by Story
// 2.7; params is a NAMESPACE, resolved from its own root — conflating
// the two is how "page" would eventually acquire a namespace, which
// AD-4 forbids forever).
func BindText(text string, data, params Value, elementID string) (string, error) {
	s, _, err := BindTextSpans(text, data, params, elementID)
	return s, err
}

// BindTextSpans is BindText plus the report of WHAT it substituted
// WHERE: one Substitution per resolved placeholder, in output order,
// with rune spans into the returned string.
//
// It is the same traversal, not a second one. BindText delegates here,
// so there is exactly one implementation of the binding grammar and the
// spans cannot drift from the text they describe — the same "only one
// derivation" discipline Story 2.3's Blocker 1 imposed on advances.
//
// {{page}} and {{pages}} produce NO Substitution. They are reserved
// tokens owned by Story 2.7, resolved from neither root, and they name
// no data path — so there is nothing a document could declare about
// them.
//
// This is a thin wrapper over Resolve, constructing a Scope with NO row
// set — the byte-identical pre-3.1 behaviour when no row is active. A
// future row-scoped caller (Story 4.2) calls Resolve directly with a
// Scope built through NewScope(...).WithRow(...).
func BindTextSpans(text string, data, params Value, elementID string) (string, []Substitution, error) {
	return Resolve(text, NewScope(data, params), elementID)
}

// Resolve is the one implementation of the binding grammar's
// resolution-root traversal, taking a Scope rather than bare
// data/params values, so BindText/BindTextSpans and any future
// row-scoped caller share one traversal, never two.
//
// Story 3.2 (D-1.6.5, D-3.2.1): every non-reserved "{{ }}" occurrence
// is parsed and checked by internal/expr (expr.Parse, expr.Check —
// syntax, arity, unknown-function-name, literal-argument-kind), then
// evaluated (expr.Eval) against a Resolver built from scope
// (exprResolver, below) that dispatches each path to the data, params
// or row root exactly as 1.6/3.1 did — the SAME lookupBound helper as
// before, now returning a generic expr.Value rather than a
// string-only result, so it can serve a path resolved from ANYWHERE
// in an expression (a bare top-level path, or nested inside a function
// argument), not only the top-level substitution case.
func Resolve(text string, scope Scope, elementID string) (string, []Substitution, error) {
	var out strings.Builder
	// runesWritten tracks the rune length of out, maintained alongside
	// every write rather than recomputed, so a Substitution's bounds are
	// recorded at the moment the text is appended.
	runesWritten := 0
	var subs []Substitution
	write := func(s string) {
		out.WriteString(s)
		runesWritten += utf8.RuneCountInString(s)
	}

	// The "{{ }}" search itself is expr.ScanPlaceholders — AD-13's ONE
	// tokenizer for this syntax (QA Finding 5, Major: this loop used to
	// hand-roll its own byte-identical copy of ScanPlaceholders' scan,
	// which scan.go's own doc comment claimed did not exist). literal[i]
	// is the plain text immediately preceding placeholders[i]; trailing
	// is whatever follows the last placeholder (the whole string, if
	// there were none) — exactly the segmentation this loop needs to
	// reassemble its output.
	literal, placeholders, trailing, serr := expr.ScanPlaceholders(text)
	if serr != nil {
		return "", nil, fmt.Errorf("bind: element %s: %s", elementID, serr)
	}
	for i, ph := range placeholders {
		write(literal[i])

		trimmed := strings.TrimSpace(ph.Inner)
		if ph.Reserved {
			// AC4: reserved for Story 2.7 — left byte-for-byte
			// unchanged, never resolved from data, never an error. It
			// names no data path, so it yields no Substitution.
			write("{{")
			write(ph.Inner)
			write("}}")
			continue
		}

		astExpr, perr := expr.Parse(ph.Inner)
		if perr != nil {
			return "", nil, fmt.Errorf(
				"bind: element %s: %q is not a valid expression: %s",
				elementID, trimmed, perr,
			)
		}
		if cerr := expr.Check(astExpr); cerr != nil {
			return "", nil, fmt.Errorf("bind: element %s: %s", elementID, cerr)
		}

		resolver := exprResolver{scope: scope, elementID: elementID}
		val, everr := expr.Eval(astExpr, resolver, elementID)
		if everr != nil {
			return "", nil, everr
		}

		from := runesWritten
		switch val.Kind {
		case expr.KindNull:
			// AD-14: an explicit JSON null renders as empty, never an
			// error.
		case expr.KindString:
			write(val.Str)
		default:
			return "", nil, fmt.Errorf(
				"bind: element %s: %q resolved to a %s, not a string — text bindings are never coerced",
				elementID, trimmed, val.Kind,
			)
		}
		subs = append(subs, Substitution{Path: substitutionPathFor(astExpr, trimmed), Start: from, End: runesWritten})
	}
	write(trailing)
	return out.String(), subs, nil
}

// substitutionPathFor reports the dotted path a resolved Substitution
// names: the joined segments for a bare path (the common case, and the
// only case before Story 3.2), or the expression's own raw source text
// for anything else (a function call) — there is no single "path" a
// general expression names, and the raw text is the most honest
// available answer.
func substitutionPathFor(e expr.Expr, trimmed string) string {
	if p, ok := e.(*expr.PathExpr); ok {
		return strings.Join(p.Segments, ".")
	}
	return trimmed
}

// exprResolver adapts a Scope to expr.Resolver (ast.go): it is the ONE
// place root dispatch (data/params/row) happens for the WHOLE
// expression tree being evaluated — including a path nested inside a
// function argument, e.g. upper(params.name) — not only a bare
// top-level path, which is why dispatch cannot be decided once before
// calling into expr.Eval and must instead live inside Resolve itself.
type exprResolver struct {
	scope     Scope
	elementID string
}

func (r exprResolver) Resolve(path []string) (expr.Value, error) {
	if len(path) == 0 {
		return expr.Value{}, fmt.Errorf("bind: element %s: empty path", r.elementID)
	}

	// AC12/AC13/D-1.7.4: the FIRST segment "params" always selects the
	// params root, decided at the path-segment level (not on a
	// trimmed literal string) so whitespace-tolerant spellings are
	// caught identically.
	if path[0] == "params" {
		if len(path) == 1 {
			// AC17 (1.7): "{{params}}" bare, no dot — params names a
			// namespace, not a value.
			return expr.Value{}, fmt.Errorf("bind: element %s: %q is a namespace, not a value", r.elementID, "params")
		}
		return lookupBound(r.scope.params, path[1:], path, r.elementID, "params", "the supplied params")
	}

	// Story 3.1/D-3.1.1: when a row scope is active and the FIRST
	// segment equals the region's declared alias, the row root is
	// selected — evaluated AFTER params (params "can be shadowed by
	// nothing") and BEFORE the data root (a row never shadows the
	// document root). The rootName passed to lookupBound is the
	// LITERAL "row" — the root CLASS, never scope.rowAlias, which is
	// document data and would trip TestBindResolutionRootsAreClosed's
	// non-literal Fatalf.
	if r.scope.rowSet && path[0] == r.scope.rowAlias {
		if len(path) == 1 {
			return expr.Value{}, fmt.Errorf("bind: element %s: %q is a namespace, not a value", r.elementID, path[0])
		}
		return lookupBound(r.scope.row, path[1:], path, r.elementID, "row", "the current row")
	}

	return lookupBound(r.scope.data, path, path, r.elementID, "data", "the report data")
}

// lookupBound resolves subPath against root — data, params or row
// (internal/bind/resolution_roots_arch_test.go's structural guard
// scans this exact call shape, rootName at argument index 4, and
// D-3.1.1's declaredResolutionRoots above must observe exactly the
// three literal strings passed here) — and converts a Present value to
// a general expr.Value (Story 3.2: no longer string-only, so a path
// resolved from inside a function argument — e.g. if(row.active, …) —
// can carry a bool, not only text). fullPath is the ORIGINAL
// placeholder path (the full dotted path, including a leading "params"
// segment when root is the params tree) used only for error text.
func lookupBound(root Value, subPath, fullPath []string, elementID, rootName, rootDesc string) (expr.Value, error) {
	val, presence := root.Lookup(subPath)
	switch presence {
	case Absent:
		return expr.Value{}, fmt.Errorf("bind: element %s: %s path %q is absent from %s", elementID, rootName, strings.Join(fullPath, "."), rootDesc)
	case Null:
		return expr.Value{Kind: expr.KindNull}, nil
	case Present:
		switch val.Kind {
		case KindString:
			return expr.Value{Kind: expr.KindString, Str: val.Str}, nil
		case KindBool:
			return expr.Value{Kind: expr.KindBool, Bool: val.Bool}, nil
		case KindNumber:
			d, derr := val.AsDecimal()
			if derr != nil {
				return expr.Value{}, fmt.Errorf(
					"bind: element %s: %s path %q: %w",
					elementID, rootName, strings.Join(fullPath, "."), derr,
				)
			}
			return expr.Value{Kind: expr.KindNumber, Num: d}, nil
		default:
			return expr.Value{}, fmt.Errorf(
				"bind: element %s: %s path %q is a %s, not a scalar value usable in an expression",
				elementID, rootName, strings.Join(fullPath, "."), val.Kind,
			)
		}
	}
	// Unreachable given Presence's own closed three-value enum
	// (Absent/Null/Present, all three handled above). QA Finding (Nit):
	// this used to fall through to "return expr.Value{}, nil" — a
	// silent KindNull with no error at all. evalIf reads a null
	// condition as SILENTLY false (owner ruling), so if a fourth
	// Presence state were ever added, this used to be the quietest
	// possible failure mode: a whole section could disappear from a
	// rendered document with no diagnostic anywhere. Kept as a located
	// error instead, naming the unhandled value, so a future enum
	// widening fails loudly here rather than silently downstream.
	return expr.Value{}, fmt.Errorf(
		"bind: element %s: internal: %s path %q resolved to an unhandled presence value %v",
		elementID, rootName, strings.Join(fullPath, "."), presence,
	)
}
