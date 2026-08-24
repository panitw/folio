// This file implements AC15-AC21 (D-1.6.5): the accepted binding
// grammar, its loud rejection of anything expression-shaped, and the
// {{page}}/{{pages}} reservation. Scope (AC20): callers apply BindText
// only to text-element `value` interpolation — table.bind and
// columns[].bind belong to the table stories and Epic 3, and are never
// passed here.
package bind

import (
	"fmt"
	"strings"
	"unicode/utf8"
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
	// leading "params" segment when it resolved against the params root.
	Path string

	// Start, End are rune indices into BindTextSpans' returned string.
	Start, End int
}

// reservedPlaceholders are AD-4's page-number slots (D-1.6.5, AC18):
// owned by Story 2.7, never resolved from data and never an error at
// 1.6. A data document containing a top-level "page" key must not
// change what {{page}} renders (AC19) — this reservation is checked
// BEFORE any data lookup is attempted, so it can never be shadowed.
var reservedPlaceholders = map[string]bool{
	"page":  true,
	"pages": true,
}

// BindText resolves every "{{…}}" placeholder in text against data and
// params (AC15-AC21, AC12-AC17, D-1.6.5, D-1.7.4), scoped to one text
// element (elementID names it in every error, AC8/AC16/AC17). "{{page}}"
// and "{{pages}}" pass through unchanged (AC18). Anything else that is
// not a bare dotted path ("{{" ws? ident ("." ident)* ws? "}}", AC15) is
// a located error naming elementID and mentioning Epic 3 (AC17) — this
// is the mechanism that keeps 1.6 from becoming a second, partial
// expression implementation (AC16).
//
// AD-14's three cases (D-1.6.2, AC8-AC11): an absent path is an error
// naming both the data path and elementID; an explicit JSON null
// renders as empty and is not an error; a value of the wrong kind for
// a text binding (anything but a string) is an error, never coerced.
//
// "params" is a SECOND resolution root, not a reserved token (Story
// 1.7, AC12-AC17, D-1.7.4, verbatim): "params is a namespace — a
// second resolution root resolved at bind time, alongside the data
// root." A placeholder whose FIRST path segment is "params" ALWAYS
// resolves against params, never against data (AC13) — regardless of
// whether data itself happens to carry a top-level "params" key
// (AC14/AC15): that key is ordinary, unreachable caller JSON. This is
// deliberately NOT implemented by extending reservedPlaceholders below
// (AC12): page/pages are reserved whole TOKENS, resolved from neither
// root and owned by Story 2.7; params is a NAMESPACE, resolved from
// its own root — conflating the two is how "page" would eventually
// acquire a namespace, which AD-4 forbids forever.
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
// Story 2.4 (D-2.1.10) needs this because a template may declare that
// certain DATA PATHS are never to be split across lines, and honouring
// that requires knowing which stretch of the bound string came from
// which path. folio-format.md:130 lets a text `value` mix literal text
// with bindings — both of the specification's own worked examples do
// ("Statement for {{customer.name}}") — so "the whole element" is not a
// usable answer: it would forbid breaking between "Statement" and
// "for". The span is the unit, and this is where the span is known.
//
// {{page}} and {{pages}} produce NO Substitution. They are reserved
// tokens owned by Story 2.7, resolved from neither root, and they name
// no data path — so there is nothing a document could declare about
// them.
func BindTextSpans(text string, data, params Value, elementID string) (string, []Substitution, error) {
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
	rest := text
	for {
		start := strings.Index(rest, "{{")
		if start < 0 {
			write(rest)
			break
		}
		write(rest[:start])
		afterOpen := rest[start+2:]
		end := strings.Index(afterOpen, "}}")
		if end < 0 {
			return "", nil, fmt.Errorf("bind: element %s: unterminated \"{{\" (missing closing \"}}\")", elementID)
		}
		inner := afterOpen[:end]
		rest = afterOpen[end+2:]

		trimmed := strings.TrimSpace(inner)
		if reservedPlaceholders[trimmed] {
			// AC18: reserved for Story 2.7 — left byte-for-byte
			// unchanged, never resolved from data, never an error. It
			// names no data path, so it yields no Substitution.
			write("{{")
			write(inner)
			write("}}")
			continue
		}

		path, perr := parseBindingPath(inner)
		if perr != nil {
			return "", nil, fmt.Errorf(
				"bind: element %s: %q is not a supported binding path (%s) — expression syntax is not yet supported (Epic 3)",
				elementID, trimmed, perr,
			)
		}

		// record wraps one resolved placeholder's write, capturing the
		// rune span it occupies. Used by both roots below, so neither
		// can acquire a different notion of where a substitution sits.
		record := func(resolved *string) {
			from := runesWritten
			if resolved != nil {
				write(*resolved)
			}
			subs = append(subs, Substitution{Path: strings.Join(path, "."), Start: from, End: runesWritten})
		}

		// AC12/AC13/D-1.7.4: the FIRST segment "params" always selects
		// the params root, decided at the path-segment level (not on
		// the trimmed literal string) so the whitespace-tolerant
		// spelling "{{ params.x }}" is caught identically (M-6).
		if path[0] == "params" {
			if len(path) == 1 {
				// AC17: "{{params}}" bare, no dot — params names a
				// namespace, not a value.
				return "", nil, fmt.Errorf("bind: element %s: %q is a namespace, not a value", elementID, trimmed)
			}
			resolved, err := lookupBound(params, path[1:], path, elementID, "params", "the supplied params")
			if err != nil {
				return "", nil, err
			}
			record(resolved)
			continue
		}

		// This story's review, Finding 3 (Major): D-1.7.4's binding
		// clause is "the same code path as absent data" — the data
		// branch now routes through lookupBound, the SAME helper the
		// params branch above uses, rather than a second inline copy
		// of AD-14's three-case switch. The two copies were output-
		// equivalent (rootName "data", rootDesc "the report data"
		// reproduce the original messages byte-for-byte), so this is a
		// pure de-duplication: AD-14's Null/wrong-kind cases and
		// D-1.6.6's bound-value check now hold on BOTH roots by
		// construction, and TestAD14Triple (which already exercises
		// this branch for the data root) covers lookupBound directly
		// rather than the abandoned inline copy.
		resolved, err := lookupBound(data, path, path, elementID, "data", "the report data")
		if err != nil {
			return "", nil, err
		}
		record(resolved)
	}
	return out.String(), subs, nil
}

// lookupBound resolves subPath against root — EITHER root, data or
// params (this story's review, Finding 3: D-1.7.4's binding text is
// "the same code path as absent data", so this is now the ONE
// implementation of AD-14's three-case switch, called from both the
// data branch and the params branch of BindText above, not a params-
// only helper with an abandoned inline twin) — and returns the text to
// append, or nil for a null value that renders as empty (AC9's rule).
// fullPath is the ORIGINAL placeholder path (the full dotted path,
// including a leading "params" segment when root is the params tree)
// used only for error text, so an absent-params error names
// "params.reportDate" (AC16) rather than the report-data phrasing
// (M-6: "absent from the report data" is actively misleading for a
// value that was never sought there at all) — rootName/rootDesc carry
// that distinct wording ("data"/"the report data" or "params"/"the
// supplied params").
func lookupBound(root Value, subPath, fullPath []string, elementID, rootName, rootDesc string) (*string, error) {
	val, presence := root.Lookup(subPath)
	switch presence {
	case Absent:
		return nil, fmt.Errorf("bind: element %s: %s path %q is absent from %s", elementID, rootName, strings.Join(fullPath, "."), rootDesc)
	case Null:
		return nil, nil
	case Present:
		if val.Kind == KindNumber {
			if _, derr := val.AsDecimal(); derr != nil {
				return nil, fmt.Errorf(
					"bind: element %s: %s path %q: %w",
					elementID, rootName, strings.Join(fullPath, "."), derr,
				)
			}
		}
		if val.Kind != KindString {
			return nil, fmt.Errorf(
				"bind: element %s: %s path %q is a %s, not a string — text bindings are never coerced",
				elementID, rootName, strings.Join(fullPath, "."), val.Kind,
			)
		}
		s := val.Str
		return &s, nil
	}
	return nil, nil
}

// parseBindingPath parses inner (the raw text between "{{" and "}}",
// not yet trimmed) against AC15's grammar:
//
//	binding := "{{" ws? ident ( "." ident )* ws? "}}"
//	ident   := [A-Za-z_][A-Za-z0-9_]*
//
// Only leading/trailing whitespace is tolerated (ws?); any interior
// whitespace, or any character outside [A-Za-z0-9_.], is rejected
// (AC16): "(", ")", "[", "]", ",", quotes, operators and interior
// whitespace inside the path.
func parseBindingPath(inner string) ([]string, error) {
	s := strings.TrimSpace(inner)
	if s == "" {
		return nil, fmt.Errorf("empty binding")
	}

	segments := strings.Split(s, ".")
	for _, seg := range segments {
		if !isValidIdent(seg) {
			return nil, fmt.Errorf("%q is not a valid path segment", seg)
		}
	}
	return segments, nil
}

// isValidIdent reports whether s matches AC15's ident grammar:
// [A-Za-z_][A-Za-z0-9_]*.
func isValidIdent(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c == '_':
			// legal anywhere
		case c >= '0' && c <= '9':
			if i == 0 {
				return false // ident may not START with a digit
			}
		default:
			return false
		}
	}
	return true
}
