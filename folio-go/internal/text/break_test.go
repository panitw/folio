package text

import "testing"

// TestComputeBreaksKnownWords is a sanity baseline: a run of two known,
// unrelated dictionary words yields a break at the boundary between
// them (not vacuously "no breaks anywhere").
func TestComputeBreaksKnownWords(t *testing.T) {
	d := Dictionary()
	// "ประเทศ" (country) + "ไทย" (Thailand) - two well-known dictionary words.
	breaks, runs := ComputeBreaks(d, "ประเทศไทย", false)
	if len(breaks) == 0 {
		t.Fatal("expected at least one break opportunity between two concatenated known words")
	}
	if len(runs) == 0 {
		t.Fatal("expected a non-empty run decomposition")
	}
}

// TestComputeBreaksNoBreakInsideCluster is AC7/P1 exercised through the
// full pipeline, not just ClusterBoundaries in isolation.
func TestComputeBreaksNoBreakInsideCluster(t *testing.T) {
	d := Dictionary()
	runes := []rune("เก็บ") // เ-ก-็-บ: a leading vowel + consonant + above mark + consonant
	breaks, _ := ComputeBreaks(d, "เก็บ", false)
	for _, b := range breaks {
		// b must never fall strictly inside the เก็ cluster (index 1..2).
		if b == 1 || b == 2 {
			t.Errorf("break reported at %d, inside a Thai character cluster in %q (runes=%q)", b, "เก็บ", runes)
		}
	}
}

// TestComputeBreaksAtomicUnknownRun is AC6/P2: a nonsense Thai string
// with no dictionary coverage at all must yield NO interior breaks and
// must be reported as a single RunUnknownThai span.
func TestComputeBreaksAtomicUnknownRun(t *testing.T) {
	d := Dictionary()
	// "ฅงฌฎ": rare/obsolete Thai consonants strung together. Verified
	// (probe, this story's dev record) to contain no dictionary match
	// at any substring length, including single characters — unlike ฆ
	// and ฏ, which the wordlist itself lists as standalone entries.
	nonsense := "ฅงฌฎ"
	breaks, runs := ComputeBreaks(d, nonsense, false)
	if len(breaks) != 0 {
		t.Errorf("expected zero interior breaks in an uncoverable run, got %v", breaks)
	}
	if len(runs) != 1 || runs[0].Kind != RunUnknownThai {
		t.Errorf("expected a single RunUnknownThai span, got %+v", runs)
	}
}

// TestComputeBreaksMixedScript is P6d's sanity check: mixed Thai/Latin
// content produces a break at the script boundary.
func TestComputeBreaksMixedScript(t *testing.T) {
	d := Dictionary()
	breaks, runs := ComputeBreaks(d, "ประเทศABC123", false)
	if len(breaks) == 0 {
		t.Fatal("expected a break at the Thai/Latin script boundary")
	}
	var sawNonThai bool
	for _, r := range runs {
		if r.Kind == RunNonThai {
			sawNonThai = true
		}
	}
	if !sawNonThai {
		t.Error("expected at least one RunNonThai span for the Latin/digit tail")
	}
}

// TestUnconstrainedVsConstrainedSwitchActuallyToggles is V11's guard:
// the unconstrained/constrained switch must provably change behaviour
// on at least one real input, or P6f/P6g measure nothing.
func TestUnconstrainedVsConstrainedSwitchActuallyToggles(t *testing.T) {
	d := Dictionary()
	// "ดอเลาะ": a genuine Thai-Malay regional surname (this story's
	// corpus, name-116/name-117 population — re-pointed from the
	// pre-rebuild name-101/name-102 ids, this story's second QA review,
	// Nit 1). Verified (probe, this
	// story's dev record) to be a case where the two modes ACTUALLY
	// differ: unconstrained proposes no break at all (nothing in it
	// matches anything, so nothing is ever proposed — "nothing to
	// override"), while the constrained engine's atomic-run
	// resumption scan finds a short, spurious legal match partway
	// through and proposes an interior break there. The original
	// example ("บ้านรถ") looked plausible but measured IDENTICALLY
	// under both modes (this story's reopening finding) — a compound
	// of two morphemes that are BOTH cleanly recognised, at a position
	// that is also a clean cluster boundary, gives greedy matching no
	// reason to behave differently whether or not AD-25's constraints
	// are active, so it never actually red-proved anything.
	compound := "ดอเลาะ"

	unconBreaks, _ := ComputeBreaks(d, compound, true)
	conBreaks, _ := ComputeBreaks(d, compound, false)

	// V11's actual requirement (the reopening's finding: a version of
	// this test asserting only "unconstrained found >=1 break" passes
	// even when `unconstrained` is hard-coded to false at the top of
	// ComputeBreaks — it never compared the two modes at all). This
	// MUST assert the two RESULTS actually differ for this input, not
	// merely that one of them is non-empty.
	sameLen := len(unconBreaks) == len(conBreaks)
	same := sameLen
	if sameLen {
		for i := range unconBreaks {
			if unconBreaks[i] != conBreaks[i] {
				same = false
				break
			}
		}
	}
	if same {
		t.Fatalf("V11 VIOLATION: unconstrained and constrained produced IDENTICAL break sets (%v) for %q — the switch is not proven to toggle anything; injecting `unconstrained = false` at the top of ComputeBreaks must make this test fail, and it would not if this assertion were absent", unconBreaks, compound)
	}

	t.Logf("compound=%q unconstrained breaks=%v constrained breaks=%v (differ, as required)", compound, unconBreaks, conBreaks)
}
