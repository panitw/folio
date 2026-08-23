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

// BindText resolves every "{{…}}" placeholder in text against data
// (AC15-AC21), scoped to one text element (elementID names it in every
// error, AC8/AC16/AC17). "{{page}}" and "{{pages}}" pass through
// unchanged (AC18). Anything else that is not a bare dotted path
// ("{{" ws? ident ("." ident)* ws? "}}", AC15) is a located error
// naming elementID and mentioning Epic 3 (AC17) — this is the
// mechanism that keeps 1.6 from becoming a second, partial expression
// implementation (AC16).
//
// AD-14's three cases (D-1.6.2, AC8-AC11): an absent path is an error
// naming both the data path and elementID; an explicit JSON null
// renders as empty and is not an error; a value of the wrong kind for
// a text binding (anything but a string) is an error, never coerced.
func BindText(text string, data Value, elementID string) (string, error) {
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

		val, presence := data.Lookup(path)
		switch presence {
		case Absent:
			return "", fmt.Errorf("bind: element %s: data path %q is absent from the report data", elementID, strings.Join(path, "."))
		case Null:
			// AC9: renders as empty, not an error — nothing appended.
		case Present:
			// QA Finding 2 (this story's review, Major): AD-23's "each
			// literal is converted to an exact scaled-integer decimal"
			// must happen on a real production path, not only inside
			// internal/bind's own unit tests — and AC3/AC4's mandated
			// "located error naming the data path and the element id"
			// must actually exist somewhere. A number is always wrong
			// kind for a text binding (AC10), but before saying so we
			// still validate it via AsDecimal so a coefficient- or
			// exponent-out-of-bound literal in REPORT DATA is reported
			// as such — located — rather than silently passing as an
			// unremarkable "wrong kind" and never touching Decimal at
			// all. Coercion is still never attempted (AC10): a
			// well-formed number is still rejected below.
			if val.Kind == KindNumber {
				if _, derr := val.AsDecimal(); derr != nil {
					return "", fmt.Errorf(
						"bind: element %s: data path %q: %w",
						elementID, strings.Join(path, "."), derr,
					)
				}
			}
			if val.Kind != KindString {
				return "", fmt.Errorf(
					"bind: element %s: data path %q is a %s, not a string — text bindings are never coerced",
					elementID, strings.Join(path, "."), val.Kind,
				)
			}
			out.WriteString(val.Str)
		}
	}
	return out.String(), nil
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
