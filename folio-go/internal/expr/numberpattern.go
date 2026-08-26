package expr

import (
	"fmt"
	"strings"
)

// This file is AC10b: formatNumber's pattern grammar is a SEPARATE
// closed set from the date grammar (datepattern.go) — it is not the
// same grammar with different letters, and it owes AC17's own forcing
// assertion (below).
//
// maxPatternFractionDigits is AC17's named bound, cited BY SYMBOL
// everywhere (AC15): the maximum number of fraction digits any
// accepted formatNumber pattern may request is avgExtraScale
// (reduce.go, D-3.1a.1) — never the literal 4. avg() always produces
// AT LEAST avgExtraScale fractional digits (its own scale is maximum
// operand scale + avgExtraScale, and operand scale is always >= 0), so
// bounding the pattern validator's acceptance at avgExtraScale is what
// keeps "no formatNumber pattern can request more fractional digits
// than avg produces" true DERIVATIONALLY rather than by two
// declarations that merely happen to agree today.
const maxPatternFractionDigits = avgExtraScale

// numberPatternSpec is a validated formatNumber pattern's shape:
// enough to drive formatNumber's own rendering (numberformat.go)
// without re-parsing the pattern string a second time.
type numberPatternSpec struct {
	// minIntDigits is the count of '0' characters in the pattern's
	// integer part — the minimum digit count the integer part is
	// zero-padded to. Always >= 1 (a pattern with none is rejected).
	minIntDigits int

	// fracDigits is the count of '0' characters after the pattern's
	// single '.', or 0 when the pattern has no '.'. Bounded by
	// maxPatternFractionDigits (AC17).
	fracDigits int

	// grouping reports whether the pattern's integer part contains a
	// ',' — formatNumber groups digits by the locale's own groupSize
	// (locale.go) only when this is true.
	grouping bool
}

// validateNumberPattern parses and validates a formatNumber pattern
// against AC10b's closed grammar: an integer part drawn from
// {'#', '0', ','}, containing at least one '0'; an optional single
// '.' followed by a fraction part drawn from {'0'} only, bounded by
// maxPatternFractionDigits. Anything else is a located LOAD error
// naming the pattern and the offending character (consistent with
// AC10a's date-pattern error shape).
func validateNumberPattern(pattern string) (numberPatternSpec, error) {
	intPart := pattern
	fracPart := ""
	hasFrac := false

	if i := strings.IndexByte(pattern, '.'); i >= 0 {
		intPart, fracPart, hasFrac = pattern[:i], pattern[i+1:], true
		if strings.IndexByte(fracPart, '.') >= 0 {
			return numberPatternSpec{}, fmt.Errorf(
				"expr: number pattern %q: more than one '.' — the fraction separator may appear at most once",
				pattern,
			)
		}
	}

	if intPart == "" {
		return numberPatternSpec{}, fmt.Errorf("expr: number pattern %q: the integer part is empty", pattern)
	}

	minInt := 0
	grouping := false
	for _, c := range intPart {
		switch c {
		case '#':
		case '0':
			minInt++
		case ',':
			grouping = true
		default:
			return numberPatternSpec{}, fmt.Errorf(
				"expr: number pattern %q: %q is not part of the closed number-pattern grammar ('#', '0', ',' and one '.') — AC10b",
				pattern, string(c),
			)
		}
	}
	if minInt == 0 {
		return numberPatternSpec{}, fmt.Errorf(
			"expr: number pattern %q: the integer part must contain at least one '0'", pattern,
		)
	}

	if hasFrac {
		for _, c := range fracPart {
			if c != '0' {
				return numberPatternSpec{}, fmt.Errorf(
					"expr: number pattern %q: the fraction part must consist only of '0' (got %q) — AC10b",
					pattern, string(c),
				)
			}
		}
	}

	fracDigits := len(fracPart)
	if fracDigits > maxPatternFractionDigits {
		return numberPatternSpec{}, fmt.Errorf(
			"expr: number pattern %q: %d fraction digit(s) exceeds the maximum %d (avgExtraScale, AC17) that avg() is guaranteed to produce",
			pattern, fracDigits, maxPatternFractionDigits,
		)
	}

	return numberPatternSpec{minIntDigits: minInt, fracDigits: fracDigits, grouping: grouping}, nil
}
