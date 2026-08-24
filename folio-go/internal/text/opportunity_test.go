package text

import "testing"

// runesOf is a readability helper: the tests below state positions in
// rune indices, and the inputs are multi-byte, so a byte-indexed
// mistake would be invisible without it.
func runesOf(s string) []rune { return []rune(s) }

// TestLatinBreaksAfterWhitespaceRunAndConsumesIt is AC1.
//
// Presence precondition: the opportunity set asserted on is required to
// be NON-EMPTY, so an implementation returning nil for every input
// cannot pass. The exact positions are asserted, not merely the count,
// because a count assertion cannot tell "after the space" from "before
// the space" — which is the whole polarity AC1 fixes.
//
// Red-proof: change Opportunities to emit {LineEnd: j, NextStart: j}
// (break AFTER the run but draw the space on the previous line) or
// {LineEnd: i, NextStart: i} (break BEFORE the run and draw the space
// on the next line); either moves "Page footer 0123456789" off the
// asserted pair and this test fails naming the input.
func TestLatinBreaksAfterWhitespaceRunAndConsumesIt(t *testing.T) {
	dict := Dictionary()

	const subject = "Page footer 0123456789"
	ops := Opportunities(dict, subject, nil)
	if len(ops) == 0 {
		t.Fatalf("precondition: %q yielded no break opportunities at all — the assertions below would be vacuous", subject)
	}

	// "Page footer 0123456789": spaces at rune 4 and rune 11.
	want := []Opportunity{{LineEnd: 4, NextStart: 5}, {LineEnd: 11, NextStart: 12}}
	if len(ops) != len(want) {
		t.Fatalf("%q: got %d opportunities %v, want %d %v", subject, len(ops), ops, len(want), want)
	}
	for i := range want {
		if ops[i] != want[i] {
			t.Errorf("%q: opportunity %d = %+v, want %+v — LineEnd is where the line ENDS (before the space) and NextStart where the next line BEGINS (after it)", subject, i, ops[i], want[i])
		}
	}

	// The consumed run is genuinely consumed: the whitespace belongs to
	// neither side.
	r := runesOf(subject)
	for _, o := range ops {
		if o.NextStart == o.LineEnd {
			t.Errorf("%q: opportunity %+v consumes nothing, but it sits at a whitespace run", subject, o)
			continue
		}
		for k := o.LineEnd; k < o.NextStart; k++ {
			if r[k] != ' ' {
				t.Errorf("%q: opportunity %+v claims to consume rune %d = %q, which is not whitespace", subject, o, k, string(r[k]))
			}
		}
	}
}

// TestLatinDoesNotBreakAnywhereElse is AC1's negative half, stated as a
// category rather than as a list of samples: for a Latin input with no
// whitespace at all, there must be ZERO opportunities — no hyphenation,
// no break at "-", no break inside a word.
func TestLatinDoesNotBreakAnywhereElse(t *testing.T) {
	dict := Dictionary()
	for _, subject := range []string{
		"co-operative",
		"antidisestablishmentarianism",
		"office",
		"AV",
		"0123456789",
		"e-mail@example.com",
	} {
		ops := Opportunities(dict, subject, nil)
		if len(ops) != 0 {
			t.Errorf("%q: got %d break opportunities %v, want 0 — Latin breaks at whitespace and nowhere else (this is NOT UAX #14)", subject, len(ops), ops)
		}
	}
}

// TestCJKBreaksBetweenAdjacentIdeographs is AC2.
//
// Presence precondition: the subject is the CJK element of the
// COMMITTED fixtures/shaped-text/ template (e5), whose glyph coverage
// against the shipped Noto Sans SC face is already proven by that
// fixture's own tests — so this is asserted on content the engine
// demonstrably renders, not on a string invented for this test.
//
// Red-proof: restrict isCJKIdeographOrKana to kana only (drop
// unicode.Han) and the subject loses every opportunity, failing the
// non-empty precondition; restrict it to Han only and the kana subject
// below loses its.
func TestCJKBreaksBetweenAdjacentIdeographs(t *testing.T) {
	dict := Dictionary()

	// fixtures/shaped-text/input.folio element e5.
	const subject = "结算单，共３页"
	ops := Opportunities(dict, subject, nil)
	if len(ops) == 0 {
		t.Fatalf("precondition: %q yielded no break opportunities — AC2's assertion would be vacuous", subject)
	}

	// 结 算 单 ， 共 ３ 页 — adjacent Han pairs are (结,算) and (算,单)
	// only. "，" (U+FF0C) and "３" (U+FF13) are neither Han nor kana, so
	// no opportunity touches them: that is kinsoku's territory and
	// kinsoku is not implemented.
	want := []Opportunity{{LineEnd: 1, NextStart: 1}, {LineEnd: 2, NextStart: 2}}
	if len(ops) != len(want) {
		t.Fatalf("%q: got %d opportunities %v, want %d %v", subject, len(ops), ops, len(want), want)
	}
	for i := range want {
		if ops[i] != want[i] {
			t.Errorf("%q: opportunity %d = %+v, want %+v", subject, i, ops[i], want[i])
		}
	}

	// The kana half of the range, so dropping either script from the
	// classifier is caught. "こんにちは" is the same Japanese sample
	// TestJapaneseTextThroughPanCJKFaceDoesNotFireDiagnostic renders
	// through the Pan-CJK face.
	const kana = "こんにちは"
	kops := Opportunities(dict, kana, nil)
	if len(kops) != len(runesOf(kana))-1 {
		t.Errorf("%q: got %d opportunities %v, want %d (between every adjacent kana pair)", kana, len(kops), kops, len(runesOf(kana))-1)
	}
}

// TestCJKZeroWidthBreaksConsumeNothing pins the other half of the
// Opportunity contract: a CJK break draws every rune, on one side or
// the other. Only whitespace is consumed.
func TestCJKZeroWidthBreaksConsumeNothing(t *testing.T) {
	dict := Dictionary()
	const subject = "结算单"
	ops := Opportunities(dict, subject, nil)
	if len(ops) == 0 {
		t.Fatalf("precondition: %q yielded no opportunities", subject)
	}
	for _, o := range ops {
		if o.LineEnd != o.NextStart {
			t.Errorf("%q: CJK opportunity %+v consumes runes [%d,%d); no CJK break may drop text", subject, o, o.LineEnd, o.NextStart)
		}
	}
}

// TestThaiOpportunitiesComeFromTheDictionaryEngine is AC3: the Thai
// positions Opportunities reports are exactly ComputeBreaks'
// constrained answer, minus the ones adjacent to whitespace (which rule
// 1 owns, at the opposite polarity).
//
// Presence precondition: a Thai subject measured to carry at least one
// interior dictionary break.
func TestThaiOpportunitiesComeFromTheDictionaryEngine(t *testing.T) {
	dict := Dictionary()

	const subject = "ประเทศไทย"
	engine, _ := ComputeBreaks(dict, subject, false)
	if len(engine) == 0 {
		t.Fatalf("precondition: %q carries no dictionary break at all", subject)
	}
	ops := Opportunities(dict, subject, nil)
	if len(ops) != len(engine) {
		t.Fatalf("%q: Opportunities reported %v, ComputeBreaks reported %v", subject, ops, engine)
	}
	for i, b := range engine {
		if ops[i].LineEnd != b || ops[i].NextStart != b {
			t.Errorf("%q: opportunity %d = %+v, want a zero-width break at %d", subject, i, ops[i], b)
		}
	}

	// AD-25's atomic-unknown-run absolute still holds through this API:
	// a Thai run the dictionary cannot cover proposes no interior break.
	// "ชัยวัฒน์" is the [0,8) uncoverable span of corpus name-021.
	const uncoverable = "ชัยวัฒน์"
	if got := Opportunities(dict, uncoverable, nil); len(got) != 0 {
		t.Errorf("%q is not dictionary-tileable, so AD-25 requires zero interior breaks; got %v", uncoverable, got)
	}
}

// TestAtomicSpansRemoveInteriorOpportunities is AC7's mechanism,
// asserted at the breaking API where the parameter arrives.
//
// PRESENCE PRECONDITION, AND IT IS THE POINT: each case first proves
// that the UNDECLARED form of the same input DOES carry an interior
// opportunity in that span. Asserting only the declared case would be
// vacuous — it could not distinguish an honoured declaration from a
// value that never had a break opportunity to begin with. Both
// polarities, every case.
//
// Coverage is by construction over the span set — the test drives the
// spans, not a list of sample strings (D-000.23).
func TestAtomicSpansRemoveInteriorOpportunities(t *testing.T) {
	dict := Dictionary()

	cases := []struct {
		name    string
		text    string
		atomic  []Span
		wantIn  []Span // spans that must end up break-free
		wantOut int    // opportunities expected to survive outside them
	}{
		{
			// D-2.1.6's worked case: "ศรีสุข" as a surname is
			// byte-identical to "ศรี" + "สุข", both in the dictionary.
			name:    "thai surname, whole string declared",
			text:    "ศรีสุข",
			atomic:  []Span{{Start: 0, End: 6}},
			wantIn:  []Span{{Start: 0, End: 6}},
			wantOut: 0,
		},
		{
			name:    "thai surname inside a literal sentence, only the value declared",
			text:    "ผู้รับ ศรีสุข",
			atomic:  []Span{{Start: 7, End: 13}},
			wantIn:  []Span{{Start: 7, End: 13}},
			wantOut: 1, // the whitespace break after "ผู้รับ" survives
		},
		{
			// Two placeholders: a filter that marked only the FIRST
			// substituted span atomic passes the one-span cases above
			// and fails here. That is this test's red-proof.
			name:    "two declared spans, both must be honoured",
			text:    "ศรีสุข ศรีสุข",
			atomic:  []Span{{Start: 0, End: 6}, {Start: 7, End: 13}},
			wantIn:  []Span{{Start: 0, End: 6}, {Start: 7, End: 13}},
			wantOut: 1,
		},
		{
			// The mechanism is script-agnostic by construction.
			name:    "CJK declared span",
			text:    "结算单",
			atomic:  []Span{{Start: 0, End: 3}},
			wantIn:  []Span{{Start: 0, End: 3}},
			wantOut: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Polarity 1 — UNDECLARED. Prove the interior break exists.
			undeclared := Opportunities(dict, tc.text, nil)
			for _, sp := range tc.wantIn {
				interior := 0
				for _, o := range undeclared {
					if o.LineEnd > sp.Start && o.LineEnd < sp.End {
						interior++
					}
				}
				if interior == 0 {
					t.Fatalf("PRECONDITION FAILED: %q carries no interior break opportunity in span %+v even when UNDECLARED (%v) — asserting its absence when declared would assert nothing", tc.text, sp, undeclared)
				}
			}

			// Polarity 2 — DECLARED. The interior breaks are gone.
			declared := Opportunities(dict, tc.text, tc.atomic)
			for _, sp := range tc.wantIn {
				for _, o := range declared {
					if o.LineEnd > sp.Start && o.LineEnd < sp.End {
						t.Errorf("%q: declared span %+v still carries an interior break %+v", tc.text, sp, o)
					}
					if o.NextStart > sp.Start && o.NextStart < sp.End {
						t.Errorf("%q: declared span %+v still has a line STARTING inside it: %+v", tc.text, sp, o)
					}
				}
			}

			// Literal text outside the declared spans is unaffected.
			if len(declared) != tc.wantOut {
				t.Errorf("%q: %d opportunities survived outside the declared spans (%v), want %d — a declaration must not suppress breaks in literal text around it", tc.text, len(declared), declared, tc.wantOut)
			}
		})
	}
}
