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
)

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
	var out strings.Builder
	rest := text
	for {
		start := strings.Index(rest, "{{")
		if start < 0 {
			out.WriteString(rest)
			break
		}
		out.WriteString(rest[:start])
		afterOpen := rest[start+2:]
		end := strings.Index(afterOpen, "}}")
		if end < 0 {
			return "", fmt.Errorf("bind: element %s: unterminated \"{{\" (missing closing \"}}\")", elementID)
		}
		inner := afterOpen[:end]
		rest = afterOpen[end+2:]

		trimmed := strings.TrimSpace(inner)
		if reservedPlaceholders[trimmed] {
			// AC18: reserved for Story 2.7 — left byte-for-byte
			// unchanged, never resolved from data, never an error.
			out.WriteString("{{")
			out.WriteString(inner)
			out.WriteString("}}")
			continue
		}

		path, perr := parseBindingPath(inner)
		if perr != nil {
			return "", fmt.Errorf(
				"bind: element %s: %q is not a supported binding path (%s) — expression syntax is not yet supported (Epic 3)",
				elementID, trimmed, perr,
			)
		}

		// AC12/AC13/D-1.7.4: the FIRST segment "params" always selects
		// the params root, decided at the path-segment level (not on
		// the trimmed literal string) so the whitespace-tolerant
		// spelling "{{ params.x }}" is caught identically (M-6).
		if path[0] == "params" {
			if len(path) == 1 {
				// AC17: "{{params}}" bare, no dot — params names a
				// namespace, not a value.
				return "", fmt.Errorf("bind: element %s: %q is a namespace, not a value", elementID, trimmed)
			}
			resolved, err := lookupBound(params, path[1:], path, elementID, "params", "the supplied params")
			if err != nil {
				return "", err
			}
			if resolved != nil {
				out.WriteString(*resolved)
			}
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
			return "", err
		}
		if resolved != nil {
			out.WriteString(*resolved)
		}
	}
	return out.String(), nil
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
