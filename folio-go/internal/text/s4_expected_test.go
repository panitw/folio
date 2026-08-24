package text

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// expectedBreakItem is one hand-authored label in S4's frozen
// expected-break fixture.
//
// Words is the LABEL — the segmentation a reader would make — and
// ExpectedBreaks is its arithmetic. A reviewer checks Words; the numbers
// follow from it. That is the shape that makes "hand-checked once" a
// thing a person can actually do (acceptance.md:65).
type expectedBreakItem struct {
	ID             string   `json:"id"`
	Text           string   `json:"text"`
	Words          []string `json:"words"`
	ExpectedBreaks []int    `json:"expectedBreaks"`
	Gloss          string   `json:"gloss"`
}

type expectedBreakFixture struct {
	Items []expectedBreakItem `json:"items"`
}

func loadExpectedBreaks(t *testing.T) expectedBreakFixture {
	t.Helper()
	path := filepath.Join(repoRootFromTest(t), "fixtures", "expected-breaks", "expected_breaks.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read expected_breaks.json: %v", err)
	}
	var f expectedBreakFixture
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("unmarshal expected_breaks.json: %v", err)
	}
	return f
}

// TestS4ExpectedBreaksAreLabelsNotEngineOutput is the fixture's
// SELF-CONSISTENCY check, and it runs before the conformance assertion
// because it is what makes the labels auditable: every item's words must
// concatenate to its text, and its expected positions must be exactly
// the cumulative rune lengths of all but the last word.
//
// This is the property that keeps the file hand-checkable. If the
// numbers could drift from the segmentation, a reviewer reading the
// words would be certifying something other than what the test asserts.
// It also makes "regenerated to make a test pass" visible: you cannot
// quietly move a number without moving a word boundary a human can see.
func TestS4ExpectedBreaksAreLabelsNotEngineOutput(t *testing.T) {
	f := loadExpectedBreaks(t)
	if len(f.Items) == 0 {
		t.Fatal("V-S4: the expected-break fixture is empty")
	}
	multiWord := 0
	for _, it := range f.Items {
		if len(it.Words) >= 2 {
			multiWord++
		}
		joined := ""
		for _, w := range it.Words {
			joined += w
		}
		if joined != it.Text {
			t.Errorf("%s: words %v concatenate to %q, but text is %q", it.ID, it.Words, joined, it.Text)
			continue
		}
		var want []int
		acc := 0
		for _, w := range it.Words[:max(0, len(it.Words)-1)] {
			acc += len([]rune(w))
			want = append(want, acc)
		}
		if len(want) == 0 {
			want = nil
		}
		got := it.ExpectedBreaks
		if len(got) == 0 {
			got = nil
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s (%q): expectedBreaks %v does not match its own segmentation %v, which implies %v — the numbers must follow the words", it.ID, it.Text, it.ExpectedBreaks, it.Words, want)
		}
	}

	// NON-DEGENERACY GUARD for BOTH assertions above, which are the same
	// conservation shape D-000.33 names: "the parts reconstruct the
	// whole". A single-word item satisfies `concat(words) == text`
	// trivially — it is `whole == whole` — and derives an EMPTY expected
	// list, so a fixture of nothing but single-word items would certify
	// neither the concatenation nor the cumulative-length arithmetic.
	// At least one item must genuinely be partitioned.
	if multiWord == 0 {
		t.Fatalf("vacuity: all %d fixture items carry a single word, so `words concatenate to text` reduces to whole == whole and the cumulative-length derivation has nothing to derive", len(f.Items))
	}
	t.Logf("V-S4 self-consistency: %d of %d items are genuinely partitioned (>= 2 words), so the concatenation and the cumulative-length derivation are non-degenerate", multiWord, len(f.Items))
}

// TestS4ExpectedBreaksMatchTheEngine is AC14: the engine's break
// opportunities must equal the frozen, hand-authored labels exactly
// (epics.md:836, "results match the fixture exactly").
//
// THE FIXTURE IS THE ORACLE, NOT THE OUTPUT. If this test fails, the
// engine is wrong or the label is wrong — and which one is a question
// for a person. It is NEVER resolved by regenerating the fixture.
// There is no generator for it, deliberately: nothing in this repository
// can write fixtures/expected-breaks/expected_breaks.json.
//
// This is distinct from TestAC10ComputedBreaksMatchS4Basis, which
// compares the engine against ITS OWN recorded output across targets and
// whose own README says in terms that it is "a cross-target REGRESSION
// ANCHOR ONLY, never a correctness oracle". That one answers "does this
// target agree with the others"; this one answers "is the answer right".
//
// Every item is free of whitespace, so each opportunity is zero-width
// (LineEnd == NextStart) and comparing positions compares everything.
// That is asserted rather than assumed.
//
// VACUITY GUARDS, all of which must hold for the comparison to mean
// anything:
//   - the fixture is non-empty;
//   - at least one item expects a NON-ZERO number of breaks (otherwise an
//     engine returning nothing would pass everything);
//   - at least one item expects ZERO (otherwise an engine returning a
//     break at every position would not be caught by the corpus of
//     compounds alone);
//   - the total number of expected positions is non-zero.
func TestS4ExpectedBreaksMatchTheEngine(t *testing.T) {
	f := loadExpectedBreaks(t)
	dict := Dictionary()

	if len(f.Items) == 0 {
		t.Fatal("V-S4: the expected-break fixture is empty")
	}
	var withBreaks, withoutBreaks, totalExpected int
	for _, it := range f.Items {
		if len(it.ExpectedBreaks) > 0 {
			withBreaks++
			totalExpected += len(it.ExpectedBreaks)
		} else {
			withoutBreaks++
		}
	}
	if withBreaks == 0 {
		t.Fatal("V-S4: no item expects a break — an engine that never proposes one would pass this fixture")
	}
	if withoutBreaks == 0 {
		t.Fatal("V-S4: no item expects ZERO breaks — the atomic/uncoverable polarity is untested")
	}
	if totalExpected == 0 {
		t.Fatal("V-S4: the fixture expects no break positions at all")
	}

	var mismatches int
	for _, it := range f.Items {
		ops := Opportunities(dict, it.Text, nil)

		got := make([]int, 0, len(ops))
		for _, o := range ops {
			if o.LineEnd != o.NextStart {
				t.Errorf("%s (%q): opportunity %+v consumes runes, but every fixture item is whitespace-free and must break zero-width", it.ID, it.Text, o)
			}
			got = append(got, o.LineEnd)
		}
		if len(got) == 0 {
			got = nil
		}
		want := it.ExpectedBreaks
		if len(want) == 0 {
			want = nil
		}
		if !reflect.DeepEqual(got, want) {
			mismatches++
			t.Errorf("S4 MISMATCH: %s (%q, %s): engine proposes %v, the frozen label expects %v (segmentation: %v).\n"+
				"    DO NOT regenerate the fixture to close this. Either the engine is wrong or the label is wrong, and that is a question for a person.",
				it.ID, it.Text, it.Gloss, got, want, it.Words)
		}
	}

	if mismatches == 0 {
		t.Logf("AC14: the engine reproduces all %d hand-authored S4 labels exactly (%d items expecting breaks, %d expecting none, %d positions in total)",
			len(f.Items), withBreaks, withoutBreaks, totalExpected)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
