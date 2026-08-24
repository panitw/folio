package text

import "unicode"

// Span is a half-open range of RUNE indices [Start, End).
//
// Rune indices, not byte offsets, because every position this package
// and its callers exchange is a rune index: AD-25's break positions, the
// corpus's proper-noun spans, and GlyphInfo.Cluster (D-2.3.2, pinned by
// TestClusterValuesAreRuneIndices). Mixing the two units is the defect
// that story warned about, so there is only one unit here.
type Span struct {
	Start, End int
}

// Opportunity is one position at which a line may break.
//
// A line that breaks here ends at LineEnd and the next line begins at
// NextStart. The two differ only where the break CONSUMES text — today
// exactly one case, a run of whitespace, which is drawn on neither
// line. For every other break the two are equal.
//
// Modelling the consumed range explicitly is what keeps "the trailing
// space does not count toward the line's width" a property of the break
// rather than a special case at the measuring site.
type Opportunity struct {
	LineEnd   int
	NextStart int
}

// isCJKIdeographOrKana reports whether r is a Han ideograph or a kana
// character, using the standard library's own Unicode script tables.
//
// IT IS A CATEGORY TEST, NOT A CHARACTER LIST, AND THAT IS BINDING
// (D-000.23: a guard written for a defect covers the defect, not its
// class; a denylist is never coverage). unicode.Han, unicode.Hiragana
// and unicode.Katakana are Unicode's own script definitions, shipped
// with the toolchain and updated with it — no table is generated here
// and no module is added, so wantModuleGraph's "exactly two modules"
// still holds.
//
// NOTE, AND IT IS A REAL NARROWING: the engine has no CJK rune
// classification of its own to reuse. Face resolution is COVERAGE-based
// (resolveRuneFace asks each face in the chain whether it has a glyph),
// so it never classifies a rune by script at all. This function is
// therefore new, and it is the only rune-script classification for CJK
// in the codebase.
//
// Fullwidth punctuation and fullwidth digits are deliberately NOT
// included. They are neither Han nor kana, and deciding whether a line
// may end before "，" or begin with it is exactly the line-start /
// line-end prohibition (kinsoku) this engine does not implement — see
// the package doc. Adding them here would be a guess about that
// question wearing a script test's clothes.
func isCJKIdeographOrKana(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r)
}

// Opportunities returns every position at which s may break, in
// ascending LineEnd order, with the positions strictly interior to any
// span in atomic removed.
//
// Three sources, one per script family:
//
//   - LATIN AND EVERYTHING ELSE — a break exists AFTER each maximal run
//     of Unicode White_Space, and the run is consumed. Nothing else in
//     Latin breaks: no hyphenation, no break at "-", no contextual pair
//     rules. THIS IS NOT UAX #14 and must never be described as it —
//     full UAX #14 needs the LineBreak property for every code point,
//     which is either a new module (forbidden: go list -m all stays at
//     exactly two) or a generated table nobody has budgeted. See the
//     package doc.
//
//   - CJK — a break exists between any two adjacent Han-or-kana runes.
//     Kinsoku (line-start / line-end punctuation prohibition) is NOT
//     implemented, as a category.
//
//   - THAI — the dictionary engine, ComputeBreaks in constrained mode,
//     under both of AD-25's absolutes plus Story 2.4's
//     both-sides-coverable filter.
//
// ATOMIC IS A PARAMETER, NOT AN IMPORT (D-000.16, verbatim: "the signal
// rides on the value, never through an import ... the breaking API
// accepts it as a parameter. Stages communicate by what they pass, not
// by what they import"). internal/text does not know that a span came
// from a bound value, from a template declaration, or from a caller's
// own reasoning, and it must not learn.
//
// Removal is by CONSTRUCTION over the span set — every span, whatever
// its script and whatever produced it — never by matching sample
// strings.
func Opportunities(dict *BytesTrie, s string, atomic []Span) []Opportunity {
	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return nil
	}

	// byLineEnd keys on the position the CURRENT line ends at, so a
	// whitespace-consuming break and a zero-width break proposed at the
	// same place collapse to one, with the consuming form winning
	// (it is the one that does not draw the space).
	byLineEnd := map[int]Opportunity{}
	add := func(o Opportunity) {
		if prev, ok := byLineEnd[o.LineEnd]; ok && prev.NextStart >= o.NextStart {
			return
		}
		byLineEnd[o.LineEnd] = o
	}

	// 1. Whitespace. A run at the very start or very end of s yields no
	// interior opportunity — there would be no text on one side of it.
	for i := 0; i < n; {
		if !unicode.IsSpace(runes[i]) {
			i++
			continue
		}
		j := i
		for j < n && unicode.IsSpace(runes[j]) {
			j++
		}
		if i > 0 && j < n {
			add(Opportunity{LineEnd: i, NextStart: j})
		}
		i = j
	}

	// 2. CJK, between adjacent Han-or-kana runes.
	for i := 1; i < n; i++ {
		if isCJKIdeographOrKana(runes[i-1]) && isCJKIdeographOrKana(runes[i]) {
			add(Opportunity{LineEnd: i, NextStart: i})
		}
	}

	// 3. Thai, from the dictionary engine. A position adjacent to
	// whitespace is skipped here and left to rule 1: ComputeBreaks
	// reports a script-span boundary at the START of a following space
	// run (isThaiScript excludes U+0020), which would break BEFORE the
	// whitespace and draw it at the head of the next line — the exact
	// polarity AC1 forbids.
	if dict != nil {
		breaks, _ := ComputeBreaks(dict, s, false)
		for _, b := range breaks {
			if b <= 0 || b >= n {
				continue
			}
			if unicode.IsSpace(runes[b-1]) || unicode.IsSpace(runes[b]) {
				continue
			}
			add(Opportunity{LineEnd: b, NextStart: b})
		}
	}

	// Collected by ascending LineEnd directly rather than by ranging
	// byLineEnd: a map range is forbidden under internal/ wherever its
	// order can reach an output byte (D-1.3.5, NFR1.d, enforced by
	// lint's map-range rule), and these opportunities decide where lines
	// break, which is about as far into the output bytes as an ordering
	// can reach. Sorting afterwards would also have been deterministic,
	// but the rule is deliberately syntactic — it does not ask whether
	// this particular range happened to be safe.
	ops := make([]Opportunity, 0, len(byLineEnd))
	for i := 1; i < n; i++ {
		o, ok := byLineEnd[i]
		if !ok {
			continue
		}
		if spansCover(atomic, o) {
			continue
		}
		ops = append(ops, o)
	}
	return ops
}

// spansCover reports whether o falls strictly inside any span in
// atomic. Both ends are tested: a break that ends a line inside a
// declared span would split it just as surely as one that starts the
// next line inside it.
func spansCover(atomic []Span, o Opportunity) bool {
	for _, sp := range atomic {
		if o.LineEnd > sp.Start && o.LineEnd < sp.End {
			return true
		}
		if o.NextStart > sp.Start && o.NextStart < sp.End {
			return true
		}
	}
	return false
}
