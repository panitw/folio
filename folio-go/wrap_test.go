package folio

import (
	"testing"

	"github.com/panitw/folio/folio-go/internal/bind"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/text"
)

// shippedChain is the fallback chain the shipped-face tests use, in
// AD-8's declared order.
func shippedChain() []string { return []string{"Noto Sans", "Noto Sans Thai", "Noto Sans SC"} }

func segmentsFor(t *testing.T, s string) []faceSegment {
	t.Helper()
	segs, _, err := shapeSegments("e-test", shippedChain(), s, testShippedFontSet(), newFontCache())
	if err != nil {
		t.Fatalf("shapeSegments(%q): %v", s, err)
	}
	if len(segs) == 0 {
		t.Fatalf("shapeSegments(%q) produced no segments", s)
	}
	return segs
}

// TestMeasureUsesShapedAdvancesNotHmtx is AC9, and its expected values
// are LITERALS hand-derived from an independent oracle rather than
// recomputed the way production computes them — the same discipline
// TestFaceSegmentOriginsUseShapedAdvances applies one layer down.
//
// Provenance (D-000.26 — the subject of the measurement, not only its
// result):
//
//	face      folio-go/fonts/notosans/NotoSans-Regular.ttf, unitsPerEm 1000
//	oracle    hb-shape (HarfBuzz) 14.2.0, --output-format=json
//	fontSize  16 pt, so millipoints = units1000 * 16
//
// PRESENCE PRECONDITION, AND IT IS THE POINT: the subject must be a run
// whose shaped advance is provably DIFFERENT from its hmtx sum. "AV"
// kerns -40 (shaped A = 599 against an hmtx 639); "Ada" does not kern at
// all. A test written only against "Ada" cannot tell the two hypotheses
// apart and would pass with fontset.AdvanceForRune on the measurement
// path — which is precisely the defect AC9 forbids. Both are asserted,
// and the non-kerning case is the negative control: it must measure the
// SAME under either hypothesis, so rejecting the wrong answer is not
// achieved by rejecting all answers.
func TestMeasureUsesShapedAdvancesNotHmtx(t *testing.T) {
	const fontSize = geom.Length(16000) // 16 pt; geom.Length is MILLIPOINTS

	cases := []struct {
		text        string
		wantShaped  geom.Length // the correct answer
		naiveHmtx   geom.Length // what a per-rune hmtx sum would give
		discriminat bool        // does this input distinguish the two?
		note        string
	}{
		{
			text: "AV ", wantShaped: 23344, naiveHmtx: 23984, discriminat: true,
			note: "A/V kerns -40: shaped 599+600+260=1459, hmtx 639+600+260=1499",
		},
		{
			text: "Wo. ", wantShaped: 32688, naiveHmtx: 33008, discriminat: true,
			note: "W/o kerns -20: shaped 910+605+268+260=2043, hmtx 930+605+268+260=2063",
		},
		{
			text: "Ada ", wantShaped: 33200, naiveHmtx: 33200, discriminat: false,
			note: "NEGATIVE CONTROL: no kern pair, so both hypotheses agree at 2075*16",
		},
	}

	var discriminating int
	for _, tc := range cases {
		segs := segmentsFor(t, tc.text)
		n := len([]rune(tc.text))
		got := measureRuneRange(segs, 0, n, fontSize)

		if tc.discriminat {
			discriminating++
			if tc.wantShaped == tc.naiveHmtx {
				t.Fatalf("%q is marked discriminating but its two hypotheses are equal (%d) — the table is wrong, not the code", tc.text, tc.wantShaped)
			}
			if got == tc.naiveHmtx {
				t.Errorf("%q measured %d millipoints, which is the PER-RUNE hmtx sum — measurement must come from the shaped advances that are drawn (%s)", tc.text, got, tc.note)
				continue
			}
		}
		if got != tc.wantShaped {
			t.Errorf("%q measured %d millipoints, want %d (%s)", tc.text, got, tc.wantShaped, tc.note)
		}
	}

	if discriminating == 0 {
		t.Fatal("vacuity: no case in this table can distinguish shaped advances from hmtx advances, so it asserts nothing about AC9")
	}
	t.Logf("AC9: %d of %d cases discriminate shaped advances from hmtx; all measured from the shaped glyphs that are drawn", discriminating, len(cases))
}

// TestMeasureIsAdditiveAcrossASplit asserts the property the line
// breaker depends on: measuring a range is the same as measuring its
// two halves and adding them, when the split is at a break opportunity.
//
// This is what makes "the line I measured is the line I draw" true. If
// slicing lost or double-counted a glyph, this would fail even though
// each individual measurement looked plausible.
//
// ADDITIVITY IS NECESSARY, NOT SUFFICIENT — AND THE EXACT LIMIT IS
// STATED HERE, MEASURED, so nobody credits this test with more than it
// has. Additivity is conserved by ANY monotone boundary function: a
// slicing rule that puts the cut at the wrong glyph but puts BOTH sides'
// glyphs somewhere still sums to the whole. What it can catch is a
// boundary that SATURATES — one side collapsing to no glyphs, leaving
// `whole + 0 == whole`, which is true of every slicing function correct
// or not. AC10's hand-derived boundary table
// (TestSliceAtRuneBoundaryUsesRuneIndices) is the discriminator for the
// shifted-but-conserving case, and this test does not duplicate it.
//
// # Two holes, both closed here, both previously open
//
//  1. NO NON-DEGENERACY GUARD. The saturating case above was not
//     excluded at all, so the assertion could have been arithmetic about
//     nothing. D-000.33 was applied to AC10's guard when it was found
//     there and not carried back to this one — the class, not the
//     instance (D-000.23).
//
//  2. NO SPLIT INTERIOR TO A FACE SEGMENT. Measured, not assumed:
//     "ณัฐวุฒิ เกิด กรุงเทพ" — the subject this test's own comment called
//     the discriminating one — segments as
//     [Thai 0,7) [Latin 7,8) [Thai 8,12) [Latin 12,13) [Thai 13,20) and
//     breaks at 7 and 12, which are EXACTLY the segment boundaries.
//     Cluster is segment-local, so at a segment boundary the rune-index
//     and byte-offset readings BOTH give 0 and coincide. Every split in
//     the original subject set was of that kind, which is why the
//     byte-offset mutation of glyphRangeForRunes left this test green.
//     "ณัฐวุฒิเกิดกรุงเทพ" is one whitespace-free Thai segment breaking at
//     runes 7 and 11 — strictly inside it — so a segment-local
//     misreading has somewhere to show.
//
// With both closed, the byte-offset mutation now reddens this test:
// at rune 7 of that subject the byte offset is 21, past the segment's
// largest cluster, the right side collapses to zero glyphs, and the
// non-degeneracy witness fires.
//
// Preconditions, all asserted below:
//   - the subject measures to a non-zero width;
//   - it has at least one break opportunity, so a split exists;
//   - EVERY subject contributes at least one NON-DEGENERATE split, in
//     which both sides measure strictly greater than zero;
//   - at least one split falls STRICTLY INSIDE a face segment.
func TestMeasureIsAdditiveAcrossASplit(t *testing.T) {
	const fontSize = geom.Length(16000) // 16 pt; geom.Length is MILLIPOINTS
	dict := text.Dictionary()

	subjects := []string{
		"Ada ก 汉",
		"ณัฐวุฒิ เกิด กรุงเทพ",
		"结算单，共３页",
		"Page footer 0123456789",
		// Whitespace-free Thai: ONE face segment, dictionary breaks at
		// runes 7 and 11, both strictly interior to it. See the second
		// hole in the doc comment above — without this subject every
		// split in this table sits on a face-segment boundary, where a
		// segment-local index and a segment-local byte offset are both
		// zero and cannot be told apart.
		"ณัฐวุฒิเกิดกรุงเทพ",
	}

	var checked, nonDegenerate, interior int
	perSubject := make(map[string]int, len(subjects))
	splitsPerSubject := make(map[string]int, len(subjects))
	for _, s := range subjects {
		segs := segmentsFor(t, s)
		n := len([]rune(s))
		whole := measureRuneRange(segs, 0, n, fontSize)
		if whole <= 0 {
			t.Fatalf("precondition: %q measured %d millipoints", s, whole)
		}
		ops := text.Opportunities(dict, s, nil)
		if len(ops) == 0 {
			t.Fatalf("precondition: %q has no break opportunity, so no split can be tested", s)
		}
		for _, op := range ops {
			left := measureRuneRange(segs, 0, op.LineEnd, fontSize)
			right := measureRuneRange(segs, op.NextStart, n, fontSize)
			consumed := measureRuneRange(segs, op.LineEnd, op.NextStart, fontSize)
			checked++
			splitsPerSubject[s]++
			// NON-DEGENERACY, measured on the QUANTITY BEING CONSERVED
			// (the width) and not merely on the rune indices: a side
			// that measures zero contributes nothing to the sum, and
			// `whole + 0 == whole` is true of any slicing function.
			if left > 0 && right > 0 {
				nonDegenerate++
				perSubject[s]++
			}
			for _, sg := range segs {
				if op.LineEnd > sg.runeStart && op.LineEnd < sg.runeEnd {
					interior++
					break
				}
			}
			if left+consumed+right != whole {
				t.Errorf("%q split at %+v: left(%d) + consumed(%d) + right(%d) = %d, but the whole measures %d",
					s, op, left, consumed, right, left+consumed+right, whole)
			}
		}
	}
	if checked == 0 {
		t.Fatal("vacuity: no split was checked")
	}

	// N-OF-N WITNESS. Every subject must carry its own non-degenerate
	// split, not just the population as a whole — otherwise one wide
	// Latin subject could certify the Thai one, which is the subject
	// this test exists for.
	witnessed := 0
	for _, s := range subjects {
		if perSubject[s] == 0 {
			t.Errorf("vacuity: every one of %q's %d splits is DEGENERATE (a side measures 0 millipoints), so for this subject the check reduces to whole+0 == whole and asserts nothing about slicing", s, splitsPerSubject[s])
			continue
		}
		witnessed++
	}
	if witnessed != len(subjects) {
		t.Fatalf("vacuity: only %d of %d subjects contributed a non-degenerate split; additivity is necessary and not sufficient, and a degenerate partition satisfies it for any slicing function", witnessed, len(subjects))
	}

	// The second precondition: at least one split must sit STRICTLY
	// INSIDE a face segment. GlyphInfo.Cluster is SEGMENT-LOCAL, so a
	// split on a segment boundary reads 0 under a rune-index and a
	// byte-offset interpretation alike; a subject set made only of those
	// cannot see a segment-local misreading at all.
	if interior == 0 {
		t.Fatalf("vacuity: not one of the %d splits swept falls strictly inside a face segment — every split sits on a segment boundary, where a segment-local rune index and a segment-local byte offset are both 0. This subject set cannot distinguish the two readings, whatever it sums to", checked)
	}
	t.Logf("measurement is additive across all %d break opportunities in %d subjects; %d splits are NON-DEGENERATE (both sides > 0 mp), witnessed in %d of %d subjects: %v; %d splits fall strictly inside a face segment",
		checked, len(subjects), nonDegenerate, witnessed, len(subjects), perSubject, interior)
}

// TestSliceAtRuneBoundaryUsesRuneIndices is AC10.
//
// Presence precondition, and it names which inputs can tell the two
// readings apart: GlyphInfo.Cluster is a RUNE index (D-2.3.2, pinned by
// TestClusterValuesAreRuneIndices). "office" is ASCII and CANNOT
// distinguish a rune index from a byte offset; "ณัฐวุฒิ" can, because
// byte offsets there would be multiples of 3. The Thai case is therefore
// the subject, and the ASCII case is included only to show it does not
// discriminate.
//
// Red-proof: treat Cluster as a byte offset in glyphRangeForRunes (scale
// localFrom/localTo by the UTF-8 length); the Thai case slices at the
// wrong glyph and the emitted line's width stops equalling the measured
// line's width — caught by THE HAND-DERIVED BOUNDARY TABLE below, at
// rune 1 and at every boundary after it.
//
// NOT by the additivity leg, and this correction is deliberate. The
// paragraph beside the table already records that additivity SURVIVED
// this mutation; an earlier version of this sentence credited it with
// the catch anyway, and the two statements could not both be true.
// Re-measured at the finish commit: under the byte-offset mutation the
// boundary-table assertions fail and the additivity leg does not, since
// "ณัฐวุฒิ"'s byte offsets run past its largest cluster and every split
// collapses to whole + 0.
func TestSliceAtRuneBoundaryUsesRuneIndices(t *testing.T) {
	const fontSize = geom.Length(16000) // 16 pt; geom.Length is MILLIPOINTS

	// "ณัฐวุฒิ" shapes to clusters 0,0,2,3,3,5,5 — non-contiguous, and
	// multi-byte, so a byte-offset reading diverges immediately.
	const subject = "ณัฐวุฒิ"
	segs := segmentsFor(t, subject)
	if len(segs) != 1 {
		t.Fatalf("precondition: %q split into %d face segments, want 1", subject, len(segs))
	}

	clusters := make([]int, 0, len(segs[0].glyphs))
	for _, g := range segs[0].glyphs {
		clusters = append(clusters, g.Cluster)
	}
	n := len([]rune(subject))
	maxCluster := 0
	for _, c := range clusters {
		if c > maxCluster {
			maxCluster = c
		}
	}
	if maxCluster >= n {
		t.Fatalf("precondition failed: %q has %d runes but reports a cluster of %d — these look like BYTE offsets, and every assertion below would be measuring the wrong thing (clusters: %v)", subject, n, maxCluster, clusters)
	}
	if len(subject) == n {
		t.Fatalf("precondition failed: %q is single-byte, so it cannot distinguish a rune index from a byte offset — pick a multi-byte subject", subject)
	}
	t.Logf("subject %q: %d runes, %d bytes, clusters %v — multi-byte, so rune indices and byte offsets are distinguishable", subject, n, len(subject), clusters)

	// THE BOUNDARY ITSELF IS ASSERTED, AS LITERALS. An earlier version of
	// this test asserted only that the two halves' widths sum to the
	// whole — and that assertion SURVIVED the byte-offset mutation,
	// because reading Cluster as a byte offset makes every boundary past
	// rune 1 collapse to "all glyphs / no glyphs", and whole + 0 still
	// equals whole. Additivity is necessary and not sufficient; it cannot
	// see a degenerate slice. So the boundaries below are hand-derived
	// from the cluster vector and written out, and additivity is kept as
	// a second, weaker check rather than the only one.
	//
	// Derivation, from clusters [0 0 2 3 3 5 5] and the rule "the first
	// glyph whose Cluster >= r":
	//
	//	r=0 -> 0   r=1 -> 2 (first cluster >= 1 is the 2 at index 2)
	//	r=2 -> 2   r=3 -> 3
	//	r=4 -> 5 (first cluster >= 4 is the 5 at index 5)
	//	r=5 -> 5   r=6 -> 7 (none; past the end)   r=7 -> 7
	//
	// Under a byte-offset reading these become 0,3,5,7,7,7,7,7 — different
	// at r=1 and at every boundary after it.
	wantBoundary := []int{0, 2, 2, 3, 5, 5, 7, 7}
	if len(wantBoundary) != n+1 {
		t.Fatalf("the hand-derived boundary table has %d entries for a %d-rune subject", len(wantBoundary), n)
	}
	whole := measureRuneRange(segs, 0, n, fontSize)
	nonDegenerate := 0
	for r := 0; r <= n; r++ {
		lo, hi := segs[0].glyphRangeForRunes(0, r)
		if lo != 0 {
			t.Errorf("prefix [0,%d) starts at glyph %d, want 0", r, lo)
		}
		if hi != wantBoundary[r] {
			t.Errorf("prefix [0,%d) ends at glyph %d, want %d — the slice boundary is the first glyph whose Cluster >= %d, and Cluster is a RUNE index (clusters: %v)", r, hi, wantBoundary[r], r, clusters)
		}
		left, right := measureRuneRange(segs, 0, r, fontSize), measureRuneRange(segs, r, n, fontSize)
		if left+right != whole {
			t.Errorf("slicing at rune %d: %d + %d != %d — the slice lost or duplicated a glyph", r, left, right, whole)
		}
		if left > 0 && right > 0 {
			nonDegenerate++
		}
	}
	// The additivity leg gets the same non-degeneracy precondition its
	// sibling now carries. r=0 and r=n are degenerate BY CONSTRUCTION
	// and are swept deliberately (they check the endpoints), so this
	// says the loop also visits interior splits where both sides carry
	// glyphs — without which this leg would be `whole+0 == whole` at
	// every position and would add nothing to the boundary table above.
	if nonDegenerate == 0 {
		t.Fatalf("vacuity: every one of the %d splits of %q is degenerate (a side measures 0 mp), so the additivity leg reduces to whole+0 == whole and is not a second check on the boundary table at all", n+1, subject)
	}

	// The ASCII counterexample, asserted as NOT discriminating so nobody
	// later "simplifies" the subject above to it.
	const ascii = "office"
	if len(ascii) != len([]rune(ascii)) {
		t.Fatalf("%q was expected to be single-byte", ascii)
	}
}

// TestPackLinesWrapsWithinTheDeclaredWidth is AC11's positive half.
//
// Presence precondition: the subject is FIRST measured to be wider than
// the box it is given, so "it wrapped" is not satisfied by an input that
// fitted all along. The box is then set to force at least two lines, and
// that is asserted rather than assumed.
func TestPackLinesWrapsWithinTheDeclaredWidth(t *testing.T) {
	const fontSize = geom.Length(16000) // 16 pt; geom.Length is MILLIPOINTS
	dict := text.Dictionary()

	subjects := []struct{ name, s string }{
		{"latin", "Page footer 0123456789 and more words to wrap"},
		{"cjk", "结算单结算单结算单结算单"},
		{"thai", "ณัฐวุฒิ เกิด กรุงเทพ ประเทศไทย"},
	}

	for _, sub := range subjects {
		t.Run(sub.name, func(t *testing.T) {
			segs := segmentsFor(t, sub.s)
			n := len([]rune(sub.s))
			full := measureRuneRange(segs, 0, n, fontSize)
			box := full / 3
			if box <= 0 {
				t.Fatalf("precondition: %q measured %d millipoints", sub.s, full)
			}
			if full <= box {
				t.Fatalf("precondition: %q measures %d, which already fits the %d box — wrapping would not be exercised", sub.s, full, box)
			}

			ops := text.Opportunities(dict, sub.s, nil)
			lines := packLines(segs, ops, n, fontSize, box)
			if len(lines) < 2 {
				t.Fatalf("%q in a %d-millipoint box produced %d line(s); the input measures %d so at least two are required", sub.s, box, len(lines), full)
			}

			// Every line fits, except where a single atomic unit cannot.
			for i, ln := range lines {
				if ln.width > box {
					// Permitted only if the line is one indivisible unit:
					// no break opportunity lies strictly inside it.
					for _, op := range ops {
						if op.LineEnd > ln.from && op.LineEnd < ln.to {
							t.Errorf("%q line %d [%d,%d) is %d wide, over the %d box, and it CONTAINS a break opportunity at %d — it should have been broken there",
								sub.s, i, ln.from, ln.to, ln.width, box, op.LineEnd)
							break
						}
					}
				}
			}

			// The lines reconstruct the text, in order, losing only the
			// whitespace the breaks consume.
			//
			// NON-DEGENERACY FIRST (D-000.33 as a class). "First line
			// starts at 0 and last line ends at n" is satisfied
			// trivially by a partition whose parts are empty, so every
			// line is required to draw something before the coverage
			// assertions below mean anything.
			for i, ln := range lines {
				if ln.to <= ln.from {
					t.Errorf("%q line %d is the EMPTY range [%d,%d): a line that draws no rune makes the reconstruction checks below hold vacuously", sub.s, i, ln.from, ln.to)
				}
			}
			if lines[0].from != 0 {
				t.Errorf("first line starts at rune %d, want 0", lines[0].from)
			}
			if last := lines[len(lines)-1]; last.to != n {
				t.Errorf("last line ends at rune %d, want %d — text was dropped", last.to, n)
			}
			for i := 1; i < len(lines); i++ {
				if lines[i].from < lines[i-1].to {
					t.Errorf("line %d starts at %d, before line %d ended at %d — text is duplicated", i, lines[i].from, i-1, lines[i-1].to)
				}
			}
		})
	}
}

// TestPackLinesOverflowsRatherThanSplittingAnAtomicUnit is AC11's other
// half, and the one that guards the same invariant as AC7 from the
// opposite side: when a declared unbreakable value does not fit its box,
// the engine overflows visibly rather than re-breaking at a guess.
//
// Presence precondition: the declared value is first measured to be
// WIDER than the box, so the assertion is not made about something that
// fitted anyway.
func TestPackLinesOverflowsRatherThanSplittingAnAtomicUnit(t *testing.T) {
	const fontSize = geom.Length(16000) // 16 pt; geom.Length is MILLIPOINTS
	dict := text.Dictionary()

	// "ศรีสุข" is D-2.1.6's worked case: both halves are dictionary
	// words, so the engine WOULD break inside it if not told otherwise.
	const value = "ศรีสุข"
	const subject = "ผู้รับ " + value
	runes := []rune(subject)
	valueStart := len([]rune("ผู้รับ "))
	span := text.Span{Start: valueStart, End: len(runes)}

	segs := segmentsFor(t, subject)
	n := len(runes)

	// Precondition A: undeclared, the value really does carry an
	// interior break opportunity.
	undeclared := text.Opportunities(dict, subject, nil)
	interior := 0
	for _, op := range undeclared {
		if op.LineEnd > span.Start && op.LineEnd < span.End {
			interior++
		}
	}
	if interior == 0 {
		t.Fatalf("precondition: %q carries no interior break in %+v even undeclared — the assertion below would be vacuous", subject, span)
	}

	// Precondition B: the value alone is wider than the box we give it.
	valueWidth := measureRuneRange(segs, span.Start, span.End, fontSize)
	box := valueWidth - 1
	if box <= 0 {
		t.Fatalf("precondition: the declared value measured %d millipoints", valueWidth)
	}

	ops := text.Opportunities(dict, subject, []text.Span{span})
	lines := packLines(segs, ops, n, fontSize, box)

	// No line boundary falls inside the declared span.
	for i, ln := range lines {
		if ln.from > span.Start && ln.from < span.End {
			t.Errorf("line %d starts at rune %d, strictly inside the declared unbreakable span %+v", i, ln.from, span)
		}
		if ln.to > span.Start && ln.to < span.End {
			t.Errorf("line %d ends at rune %d, strictly inside the declared unbreakable span %+v", i, ln.to, span)
		}
	}

	// And the overflow is real, not absorbed: some line exceeds the box.
	overflowed := false
	for _, ln := range lines {
		if ln.width > box {
			overflowed = true
		}
	}
	if !overflowed {
		t.Errorf("the declared value measures %d against a %d box, so some line MUST overflow visibly; got lines %+v — an atomic unit was silently squeezed or re-broken", valueWidth, box, lines)
	}
}

// TestPackLinesWithNoDeclaredWidthProducesOneLine pins the property
// every existing golden depends on: an element with no usable box is one
// line, exactly as before this story.
func TestPackLinesWithNoDeclaredWidthProducesOneLine(t *testing.T) {
	const fontSize = geom.Length(16000) // 16 pt; geom.Length is MILLIPOINTS
	dict := text.Dictionary()
	const subject = "Page footer 0123456789"
	segs := segmentsFor(t, subject)
	n := len([]rune(subject))
	lines := packLines(segs, text.Opportunities(dict, subject, nil), n, fontSize, 0)
	if len(lines) != 1 || lines[0].from != 0 || lines[0].to != n {
		t.Fatalf("with no declared width, got %+v, want exactly one line [0,%d)", lines, n)
	}
}

// TestAtomicSpansForIsDrivenByTheDeclarationSet is AC7's construction
// property at the seam where the document's declaration meets the
// element's substitutions: every substituted span whose path is
// declared, and no others — never a list of sample values.
func TestAtomicSpansForIsDrivenByTheDeclarationSet(t *testing.T) {
	subs := []bind.Substitution{
		{Path: "customer.name", Start: 3, End: 9},
		{Path: "customer.city", Start: 12, End: 19},
		{Path: "amount", Start: 22, End: 30},
	}
	got := atomicSpansFor([]string{"customer.name", "amount"}, subs)
	want := []text.Span{{Start: 3, End: 9}, {Start: 22, End: 30}}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("span %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if s := atomicSpansFor(nil, subs); len(s) != 0 {
		t.Errorf("no declaration must yield no atomic spans, got %+v", s)
	}
	if s := atomicSpansFor([]string{"nothing.matches"}, subs); len(s) != 0 {
		t.Errorf("a declaration matching no substitution must yield no spans, got %+v", s)
	}
}

// TestLineAdvanceIsTheMaxOverTheDeclaredChain is the ruled
// inter-baseline rule (D-2.4.2 as AMENDED), asserted with LITERALS
// derived from the committed faces' hhea tables rather than recomputed
// the way lineAdvance computes it.
//
// Provenance (D-000.26). Read via font.TableData(ot.TagHhea) +
// ot.ParseHhea, scaled to the 1000-unit em, from the faces as committed
// under folio-go/fonts/:
//
//	face             asc    desc   gap   A - D + gap
//	Noto Sans        1069   -293   0     1362
//	Noto Sans Thai   1061   -450   0     1511    <- wins DESCENT
//	Noto Sans SC     1160   -288   0     1448    <- wins ASCENT
//
// The ruled advance maximises each axis INDEPENDENTLY:
//
//	max(A) + max(|D|) + max(gap) = 1160 + 450 + 0 = 1610
//
// TWO HYPOTHESES ARE ASSERTED AGAINST, NOT MERELY ABSENT.
//
//   - firstFace: "the chain's first member governs". Varying the chain
//     ORDER proves the rule is a maximum and not a position.
//   - superseded: "max(A - D + gap) over the chain", the worst SINGLE
//     FACE — the rule as D-2.4.2 was FIRST ruled, before the amendment.
//     It gives 1511 where the ruled answer is 1610, a 99-unit shortfall
//     that is a potential ink overlap between a Thai line's below-vowels
//     and the next line's ideograph ascenders.
//
// AC14 / D-000.34 — WHAT THE AMENDMENT DID TO THIS TABLE'S OWN
// DISCRIMINATING POWER, recorded because it is exactly the failure mode
// D-000.34 names. The "Thai first" row used to be this table's NEGATIVE
// CONTROL, labelled "the one order where (a) and (b) agree": under the
// superseded rule a Thai-first chain gave 1511 by BOTH hypotheses. The
// amendment destroys that property — no single-face-first ordering can
// produce 1610, because 1610 requires TWO faces — so that row silently
// became a discriminating case and the table would have been left with
// its negative control gone and nobody told. The negative control is now
// carried explicitly by the single-face rows, where a single face cannot
// fail to supply both axes and all three hypotheses agree BY
// CONSTRUCTION.
func TestLineAdvanceIsTheMaxOverTheDeclaredChain(t *testing.T) {
	const fontSize = geom.Length(16000) // 16 pt

	// units1000 -> millipoints at 16 pt is a plain multiply: 16000/1000.
	mp := func(units int64) geom.Length { return geom.Length(units * 16) }

	cases := []struct {
		name       string
		chain      []string
		wantUnits  int64
		firstFace  int64 // what the first-face hypothesis would produce
		superseded int64 // what max(A - D + gap) over the chain would produce
		note       string
	}{
		{
			name:      "shipped chain, Latin first",
			chain:     []string{"Noto Sans", "Noto Sans Thai", "Noto Sans SC"},
			wantUnits: 1610, firstFace: 1362, superseded: 1511,
			note: "1160 (SC) + 450 (Thai) governs — neither the first face's 1362 nor any single face's 1511",
		},
		{
			name:      "same faces, CJK first — order must not matter",
			chain:     []string{"Noto Sans SC", "Noto Sans", "Noto Sans Thai"},
			wantUnits: 1610, firstFace: 1448, superseded: 1511,
			note: "a maximum is order-independent; a first-face rule would give 1448 here",
		},
		{
			name:      "Thai first — was the negative control BEFORE the amendment",
			chain:     []string{"Noto Sans Thai", "Noto Sans", "Noto Sans SC"},
			wantUnits: 1610, firstFace: 1511, superseded: 1511,
			note: "under the SUPERSEDED rule both hypotheses gave 1511 here and this row was the negative control; the amendment made it discriminating, which is why the control moved to the single-face rows (D-000.34)",
		},
		{
			name:      "NEGATIVE CONTROL: Latin-only chain pays only for Latin",
			chain:     []string{"Noto Sans"},
			wantUnits: 1362, firstFace: 1362, superseded: 1362,
			note: "one face supplies BOTH axes, so all three hypotheses agree by construction — rejecting the others is not achieved by rejecting every answer",
		},
		{
			name:      "NEGATIVE CONTROL: a chain member absent from the FontSet does not constrain",
			chain:     []string{"Noto Sans", "Not In The FontSet"},
			wantUnits: 1362, firstFace: 1362, superseded: 1362,
			note: "a face the caller did not supply cannot appear in the element",
		},
	}

	fs := testShippedFontSet()
	var vsFirstFace, vsSuperseded, controls int
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lineAdvance(tc.chain, fontSize, fs, newFontCache())
			if err != nil {
				t.Fatalf("lineAdvance(%v): %v", tc.chain, err)
			}
			want := mp(tc.wantUnits)
			if tc.wantUnits != tc.firstFace && got == mp(tc.firstFace) {
				t.Errorf("chain %v produced %d millipoints, which is the FIRST FACE's leading — the rule is a MAXIMUM over the declared chain, not a position (%s)", tc.chain, got, tc.note)
				return
			}
			if tc.wantUnits != tc.superseded && got == mp(tc.superseded) {
				t.Errorf("chain %v produced %d millipoints, which is max(hhea A - D + gap) over the chain — the SUPERSEDED worst-single-face rule. The amended rule maximises each axis INDEPENDENTLY, giving %d (%s)", tc.chain, got, want, tc.note)
				return
			}
			if got != want {
				t.Errorf("chain %v produced %d millipoints, want %d (%s)", tc.chain, got, want, tc.note)
			}
		})
		switch {
		case tc.wantUnits == tc.firstFace && tc.wantUnits == tc.superseded:
			controls++
		default:
			if tc.wantUnits != tc.firstFace {
				vsFirstFace++
			}
			if tc.wantUnits != tc.superseded {
				vsSuperseded++
			}
		}
	}
	if vsFirstFace == 0 {
		t.Fatal("vacuity: no case distinguishes the maximum from the first face, so this table asserts nothing about that half of the ruling")
	}
	if vsSuperseded == 0 {
		t.Fatal("vacuity: no case distinguishes max(A)+max(|D|)+max(gap) from the SUPERSEDED max(A-D+gap), so this table would pass unchanged against the rule the amendment replaced")
	}
	if controls == 0 {
		t.Fatal("vacuity: every case rejects both rival hypotheses, so rejecting them is achieved by rejecting EVERY answer — this table needs at least one case where all three agree")
	}
	t.Logf("D-2.4.2 (amended): %d of %d cases distinguish the ruled answer from the first-face rule, %d from the superseded worst-single-face rule, and %d are negative controls where all three agree",
		vsFirstFace, len(cases), vsSuperseded, controls)
}

// TestLineAdvanceIsNotContentDependent is constraint 1, asserted
// directly: leading is a function of (declared chain, font size) and of
// nothing that is drawn.
//
// The strongest form of the claim is that lineAdvance CANNOT see the
// content — it takes no text argument at all — so this test asserts the
// observable consequence: the same chain and size yield the same
// advance, and adding a CJK character to an element does not change it.
func TestLineAdvanceIsNotContentDependent(t *testing.T) {
	const fontSize = geom.Length(16000)
	chain := shippedChain()
	fs := testShippedFontSet()

	base, err := lineAdvance(chain, fontSize, fs, newFontCache())
	if err != nil {
		t.Fatalf("lineAdvance: %v", err)
	}
	if base <= 0 {
		t.Fatalf("precondition: lineAdvance returned %d", base)
	}
	// Repeated with a fresh cache, and against every element of the
	// shaped-text fixture's very different scripts: nothing about the
	// text can reach it.
	for _, content := range []string{"Ada", "Ada 汉", "ณัฐวุฒิ", "结算单，共３页", ""} {
		again, err := lineAdvance(chain, fontSize, fs, newFontCache())
		if err != nil {
			t.Fatalf("lineAdvance: %v", err)
		}
		if again != base {
			t.Errorf("leading changed to %d while considering content %q — leading must not depend on what appears on a line (AD-24)", again, content)
		}
	}
	// It DOES scale with size, which is the other half of constraint 1.
	half, err := lineAdvance(chain, fontSize/2, fs, newFontCache())
	if err != nil {
		t.Fatalf("lineAdvance: %v", err)
	}
	if half >= base {
		t.Errorf("leading at half the font size is %d, not less than %d — leading must be a function of font size", half, base)
	}
}

// TestLineAdvanceDeclinesTheSubstitutingAccessors pins the leading of
// each shipped face at the CHAIN layer: it equals that face's own hhea
// numbers, scaled.
//
// WHAT THIS TEST IS AND IS NOT, STATED SO IT IS NOT MISCOUNTED AS
// COVERAGE (D-000.24). The `substitutedUnits` branch below is an
// UNREACHABLE VENDOR-CONTRACT ASSERTION, in the sense
// internal/fontset's ot.NewFace comment already uses. (*ot.Face).Ascender
// substitutes 800 and Descender -200 when hhea is absent, a pair giving
// exactly 1000 units for every face alike — but requireReadableTables
// makes an absent hhea a LOAD ERROR, so no *Font whose accessors
// substitute can reach here, and for a face whose hhea IS present the
// accessors return the table's own numbers. The branch therefore cannot
// fire, and rerouting fontset.LineMetrics through the accessors leaves
// this test green. It is kept as a labelled forward guard for the day a
// substituting face becomes constructible, not as constraint 2's proof.
//
// D-2.4.2 CONSTRAINT 2'S REACHABLE GUARD IS ELSEWHERE, and it is
// red-proved against exactly that reroute:
// internal/fontset.TestLineMetricsReadsTheHheaTableNotTheVendorAccessors.
// What this test does assert, and what nothing else does, is that
// lineAdvance's per-face answer is the face's own scaled leading rather
// than any shared constant.
func TestLineAdvanceDeclinesTheSubstitutingAccessors(t *testing.T) {
	const fontSize = geom.Length(16000)
	fs := testShippedFontSet()

	// The vendor's substituted pair, 800 - (-200) = 1000 units.
	const substitutedUnits = int64(1000)

	perFace := map[string]int64{
		"Noto Sans":      1362,
		"Noto Sans Thai": 1511,
		"Noto Sans SC":   1448,
	}
	measured := make(map[string]geom.Length, len(perFace))
	for _, name := range shippedChain() {
		want := perFace[name]
		got, err := lineAdvance([]string{name}, fontSize, fs, newFontCache())
		if err != nil {
			t.Fatalf("lineAdvance(%q): %v", name, err)
		}
		measured[name] = got
		if got == geom.Length(substitutedUnits*16) {
			t.Errorf("%q produced %d millipoints, exactly the vendor's substituted 800/-200 default — leading must be read from the hhea TABLE, never through (*ot.Face).Ascender/Descender", name, got)
			continue
		}
		if got != geom.Length(want*16) {
			t.Errorf("%q produced %d millipoints, want %d (hhea ascent-descent+lineGap = %d units at 16 pt)", name, got, want*16, want)
		}
	}

	// VACUITY PRECONDITION, ANCHORED TO WHAT WAS MEASURED AND NOT TO THE
	// TABLE ABOVE. An earlier version of this precondition looped over
	// `perFace` — the same literal map that supplies `want` — so
	// `differing` could never be zero and the precondition could never
	// fail: it inspected its own expectations rather than the artifact.
	// It now inspects the values lineAdvance actually returned, so a
	// change that collapsed every shipped face onto the substituted
	// constant is reported here instead of being certified.
	differing := 0
	for _, got := range measured {
		if got != geom.Length(substitutedUnits*16) {
			differing++
		}
	}
	if len(measured) == 0 {
		t.Fatal("vacuity: no shipped face was measured at all")
	}
	if differing == 0 {
		t.Fatalf("vacuity: all %d measured shipped faces produced exactly %d millipoints, the vendor's substituted default, so this test cannot tell a real leading from a substituted one (measured: %v)",
			len(measured), substitutedUnits*16, measured)
	}
	t.Logf("%d of %d shipped faces measured a leading distinguishable from the vendor's substituted %d units: %v", differing, len(measured), substitutedUnits, measured)
}

// TestLineAdvanceRefusesAnUnsatisfiableChain: a chain with no member in
// the FontSet is a located error, never a silently substituted default.
func TestLineAdvanceRefusesAnUnsatisfiableChain(t *testing.T) {
	_, err := lineAdvance([]string{"Nope", "Also Nope"}, geom.Length(16000), testShippedFontSet(), newFontCache())
	if err == nil {
		t.Fatal("a chain with no face present in the FontSet must be an error, not a default line height")
	}
}
