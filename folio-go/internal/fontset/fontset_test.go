package fontset

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/boxesandglue/textshape/ot"
)

// testFontPath is Story 1.5's one small Latin test face (AC26), committed
// with its licence text and copyright line (AC25) at
// folio-go/testdata/fonts/Roboto-Regular.ttf.
func testFontBytes(t *testing.T) []byte {
	t.Helper()
	// internal/fontset -> internal -> folio-go
	path := filepath.Join("..", "..", "testdata", "fonts", "Roboto-Regular.ttf")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read test font %s: %v", path, err)
	}
	return data
}

// patchUnitsPerEm returns a COPY of a well-formed sfnt with the head
// table's unitsPerEm field (offset 18 within head, per ot/metrics.go)
// overwritten to value. It locates the head table via the sfnt table
// directory (12-byte header, then 16-byte records) rather than assuming
// a fixed offset, so it survives a different test font being substituted
// later.
func patchUnitsPerEm(t *testing.T, data []byte, value uint16) []byte {
	t.Helper()
	out := make([]byte, len(data))
	copy(out, data)

	numTables := binary.BigEndian.Uint16(out[4:6])
	for i := 0; i < int(numTables); i++ {
		rec := out[12+i*16 : 12+i*16+16]
		if string(rec[0:4]) != "head" {
			continue
		}
		headOffset := binary.BigEndian.Uint32(rec[8:12])
		binary.BigEndian.PutUint16(out[headOffset+18:headOffset+20], value)
		return out
	}
	t.Fatal("test font has no head table")
	return nil
}

// TestNewRejectsZeroUnitsPerEm is AC19's first retained fixture: a font
// with unitsPerEm == 0 produces a located error, not a panic. D-000.9:
// what would this print if it were unable to run at all? A test harness
// panic (not a t.Fatal) — which this test would surface immediately as a
// crashed test binary, distinguishable from a clean failure.
func TestNewRejectsZeroUnitsPerEm(t *testing.T) {
	patched := patchUnitsPerEm(t, testFontBytes(t), 0)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("New panicked instead of returning a located error (AC19): %v", r)
		}
	}()

	_, err := New("zero-upem-face", patched)
	if err == nil {
		t.Fatal("expected an error for unitsPerEm == 0, got nil")
	}
	if !strings.Contains(err.Error(), "zero-upem-face") {
		t.Errorf("error does not name the font: %v", err)
	}
	if !strings.Contains(err.Error(), "0") {
		t.Errorf("error does not name the value: %v", err)
	}
}

// TestNewRejectsUnitsPerEmAboveMax is AC19's second retained fixture: a
// font with unitsPerEm > 16384 produces a located error.
func TestNewRejectsUnitsPerEmAboveMax(t *testing.T) {
	patched := patchUnitsPerEm(t, testFontBytes(t), 20000) // > 16384, representable in uint16 (F-4)

	_, err := New("too-large-upem-face", patched)
	if err == nil {
		t.Fatal("expected an error for unitsPerEm == 20000 (> 16384), got nil")
	}
	if !strings.Contains(err.Error(), "too-large-upem-face") || !strings.Contains(err.Error(), "20000") {
		t.Errorf("error does not name both the font and the value: %v", err)
	}
}

// TestNewAcceptsAValidFace is AC19's third retained fixture: a valid
// font (unitsPerEm == 2048, measured F-5) loads and scales correctly.
func TestNewAcceptsAValidFace(t *testing.T) {
	f, err := New("roboto", testFontBytes(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if f.UnitsPerEm() != 2048 {
		t.Fatalf("UnitsPerEm() = %d, want 2048 (F-5 measurement)", f.UnitsPerEm())
	}
}

// TestHeadTimesMatchMeasuredValues is F-5's measured values, pinned:
// SOURCE head.created=3304067374 head.modified=3573633780.
func TestHeadTimesMatchMeasuredValues(t *testing.T) {
	f, err := New("roboto", testFontBytes(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	created, modified := f.HeadTimes()
	if created != 3304067374 {
		t.Errorf("created = %d, want 3304067374 (F-5)", created)
	}
	if modified != 3573633780 {
		t.Errorf("modified = %d, want 3573633780 (F-5)", modified)
	}
}

// TestSubsetContainsRequestedGlyphs is a basic sanity check: subsetting
// "AB" produces a subset in which both runes resolve to a glyph id.
func TestSubsetContainsRequestedGlyphs(t *testing.T) {
	f, err := New("roboto", testFontBytes(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sub, err := f.Subset([]rune("AB"))
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}
	if len(sub.Program) == 0 {
		t.Fatal("Subset produced no program bytes")
	}
	if _, ok := sub.GlyphForRune['A']; !ok {
		t.Error("subset has no glyph mapping for 'A'")
	}
	if _, ok := sub.GlyphForRune['B']; !ok {
		t.Error("subset has no glyph mapping for 'B'")
	}
	if len(sub.Tag) != 6 {
		t.Fatalf("tag %q is not exactly six characters", sub.Tag)
	}
	for _, c := range sub.Tag {
		if c < 'A' || c > 'Z' {
			t.Fatalf("tag %q contains a non A-Z character", sub.Tag)
		}
	}
}

// TestRepeatInvarianceWithinOneProcess is AC8's repeat-invariance proof:
// calling Subset N>=16 times in one process on the same rune set produces
// byte-identical output every time (D-1.5.7). Go randomises map
// iteration per range statement, not once per process — the lead
// measured 8 distinct iteration orders ranging an 8-element map 200
// times in one process — so this reddens immediately if a future
// textshape (or a regression in this package) drops its internal
// determinism.
func TestRepeatInvarianceWithinOneProcess(t *testing.T) {
	f, err := New("roboto", testFontBytes(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const n = 32 // >= 16, AC8.
	runes := []rune("Hello, World! 0123456789")

	first, err := f.Subset(runes)
	if err != nil {
		t.Fatalf("Subset (call 1): %v", err)
	}
	// Finding 18 (QA review): close this test's own vacuity rather than
	// relying on TestSubsetContainsRequestedGlyphs running first in the
	// same package, which is a property of test scheduling (declaration
	// order), not a dependency Go guarantees — -run filtering or a file
	// split breaks it (D-000.9).
	if len(first.Program) == 0 {
		t.Fatal("Subset (call 1) produced no program bytes — repeat-invariance would pass vacuously")
	}
	for i := 1; i < n; i++ {
		got, err := f.Subset(runes)
		if err != nil {
			t.Fatalf("Subset (call %d): %v", i+1, err)
		}
		if !bytes.Equal(first.Program, got.Program) {
			t.Fatalf("call %d diverged from call 1 (AC8 repeat-invariance, N=%d)", i+1, n)
		}
		if first.Tag != got.Tag {
			t.Fatalf("call %d tag %q diverged from call 1 tag %q", i+1, got.Tag, first.Tag)
		}
	}
}

// TestPermutationInvariance is AC8's permutation-invariance proof:
// distinct orderings of the SAME rune set produce byte-identical
// output. This is only meaningful because Subset does NOT sort its
// input (D-1.5.7, AC8a) — a sort here would make permuted-then-sorted
// inputs the same input, and this test would pass vacuously regardless
// of whether the property actually held.
func TestPermutationInvariance(t *testing.T) {
	f, err := New("roboto", testFontBytes(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ascending := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	reversed := make([]rune, len(ascending))
	for i, r := range ascending {
		reversed[len(ascending)-1-i] = r
	}
	shuffled := []rune("MZAQBWCXDVEYFUGTHSRIJKLNOP")
	// Finding 19 (QA review): assert SET equality, not just length. A
	// length-only check would still pass if the literal above were
	// edited to introduce a duplicate plus an omission — which would
	// silently compare subsets of DIFFERENT rune sets and read as either
	// a false determinism failure or a coincidental pass.
	sortedAscending := slices.Clone(ascending)
	slices.Sort(sortedAscending)
	sortedShuffled := slices.Clone(shuffled)
	slices.Sort(sortedShuffled)
	if !slices.Equal(sortedAscending, sortedShuffled) {
		t.Fatalf("test bug: shuffled %q is not a permutation of ascending %q", shuffled, ascending)
	}

	base, err := f.Subset(ascending)
	if err != nil {
		t.Fatalf("Subset(ascending): %v", err)
	}
	// Finding 18 (QA review): close this test's own vacuity — see the
	// identical rationale in TestRepeatInvarianceWithinOneProcess.
	if len(base.Program) == 0 {
		t.Fatal("Subset(ascending) produced no program bytes — permutation-invariance would pass vacuously")
	}
	rev, err := f.Subset(reversed)
	if err != nil {
		t.Fatalf("Subset(reversed): %v", err)
	}
	shuf, err := f.Subset(shuffled)
	if err != nil {
		t.Fatalf("Subset(shuffled): %v", err)
	}

	if !bytes.Equal(base.Program, rev.Program) {
		t.Fatal("Subset(ascending) and Subset(reversed) produced different bytes (AC8 permutation-invariance)")
	}
	if !bytes.Equal(base.Program, shuf.Program) {
		t.Fatal("Subset(ascending) and Subset(shuffled) produced different bytes (AC8 permutation-invariance)")
	}
	if base.Tag != rev.Tag || base.Tag != shuf.Tag {
		t.Fatalf("tags diverged across permutations: ascending=%s reversed=%s shuffled=%s", base.Tag, rev.Tag, shuf.Tag)
	}
}

// TestDeriveTagUsesClosureSetNotOutputNumbering is AC7b's retained
// fixture, killing the "output numbering" rejected reading permanently
// (D-1.5.8): two documents whose glyph sets differ but are the SAME SIZE
// must produce DIFFERENT tags. Under the rejected output-numbering
// reading (hash of {0..n-1}), both would collide because
// createCompactMapping always produces a dense {0..n-1} range regardless
// of which source glyphs those are — this test is red under that reading
// and green under the ruled one (source-font numbering), and it is not
// vacuous: the two glyph sets below are deliberately chosen to be
// same-size, different-content.
func TestDeriveTagUsesClosureSetNotOutputNumbering(t *testing.T) {
	f, err := New("roboto", testFontBytes(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Two 3-rune sets, same size, disjoint content, chosen so their
	// glyph closures differ (Latin letters map to distinct, non-composite
	// glyphs in Roboto — no shared components to accidentally collapse
	// the closures to the same set).
	setA := []rune("ABC")
	setB := []rune("DEF")

	subA, err := f.Subset(setA)
	if err != nil {
		t.Fatalf("Subset(setA): %v", err)
	}
	subB, err := f.Subset(setB)
	if err != nil {
		t.Fatalf("Subset(setB): %v", err)
	}

	if subA.NumGlyphs != subB.NumGlyphs {
		// Finding 2 (QA review): a skip here is a green build, and it
		// would silently retire the only permanent guard against the
		// catastrophic output-numbering reading the day a textshape
		// upgrade changes GSUB/composite closure for these two sets.
		// The fixture's precondition failing means the FIXTURE is
		// broken and needs a human to re-choose it — never a pass.
		t.Fatalf("test fixture assumption violated: NumGlyphs differ (%d vs %d) — not a same-size case; choose two rune sets whose closures are the same size", subA.NumGlyphs, subB.NumGlyphs)
	}
	if subA.Tag == subB.Tag {
		t.Fatalf(
			"two same-size, different-content glyph sets produced the SAME tag %q — this is exactly the "+
				"catastrophic output-numbering reading AC7b exists to kill (D-1.5.8)",
			subA.Tag,
		)
	}
}

// TestDeriveTagRedProofAgainstOutputNumbering is Finding 1's rewrite
// (QA review): the shipped version asserted only `rejectedTag(8) !=
// rejectedTag(8)` — the same argument twice, never calling deriveTag,
// never comparing two DIFFERENT glyph sets, and unfailable by
// construction. Proved by construction that it was worthless: swapping
// fontset.go's `deriveTag(plan.GlyphSet())` for the rejected
// {0..n-1} derivation reddened TestDeriveTagUsesClosureSetNotOutputNumbering
// (collided on tag "BQJKPS") while the old version of this test stayed
// green.
//
// This version exercises the REAL production path (f.Subset, which
// calls deriveTag internally) on the same two same-size, different-
// content sets AC7b uses, and separately reproduces what the REJECTED
// output-numbering derivation would have produced over the identical
// NumGlyphs. The two computations are then compared against each
// other: the rejected derivation MUST collide (it is a pure function of
// count alone, by construction of createCompactMapping's always-dense
// {0..n-1} output numbering), and the real, shipped tags MUST NOT — so
// this test fails the moment deriveTag's call site is swapped for the
// rejected derivation, exactly the mutation the reviewer used.
func TestDeriveTagRedProofAgainstOutputNumbering(t *testing.T) {
	f, err := New("roboto", testFontBytes(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	setA := []rune("ABC")
	setB := []rune("DEF")

	subA, err := f.Subset(setA)
	if err != nil {
		t.Fatalf("Subset(setA): %v", err)
	}
	subB, err := f.Subset(setB)
	if err != nil {
		t.Fatalf("Subset(setB): %v", err)
	}
	if subA.NumGlyphs != subB.NumGlyphs {
		t.Fatalf("test fixture precondition violated: NumGlyphs differ (%d vs %d) — not a same-size case", subA.NumGlyphs, subB.NumGlyphs)
	}

	// rejectedTag reproduces the REJECTED output-numbering derivation
	// directly: a hash over {0..numGlyphs-1}, using the SAME digest/tag
	// folding as production's fnv64aOverGIDs + deriveTag, so the only
	// variable is which glyph-id set is hashed.
	rejectedTag := func(numGlyphs int) string {
		gids := make([]ot.GlyphID, numGlyphs)
		for i := range gids {
			gids[i] = ot.GlyphID(i)
		}
		digest := fnv64aOverGIDs(gids)
		const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		var tag [6]byte
		v := digest
		for i := 0; i < 6; i++ {
			tag[i] = letters[v%26]
			v /= 26
		}
		return string(tag[:])
	}

	rejectedA := rejectedTag(subA.NumGlyphs)
	rejectedB := rejectedTag(subB.NumGlyphs)
	if rejectedA != rejectedB {
		t.Fatalf("test bug: rejected output-numbering derivation must collide for equal NumGlyphs (got %q vs %q)", rejectedA, rejectedB)
	}

	// The assertion that actually exercises production and fails under
	// the mutation this test is named for: if fontset.go's deriveTag
	// call site used the output-numbering derivation instead of
	// Plan.GlyphSet(), subA.Tag would equal subB.Tag here, just like the
	// rejected derivation does above.
	if subA.Tag == subB.Tag {
		t.Fatalf(
			"deriveTag collided (%q) for two same-size, different-content glyph sets — indistinguishable from the "+
				"rejected output-numbering reading, which always collides for equal NumGlyphs (%q)",
			subA.Tag, rejectedA,
		)
	}
}
