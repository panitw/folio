package rules

import (
	"path/filepath"
	"testing"
)

// TestAbsencesProductionScan is AC21's production caller: at the real
// repo root, the remaining DW-2 directory does not exist yet, so this
// must report zero findings today — and go red the day it lands.
//
// Story 1.4 (AC50, D-1.4.11) added the coverage-witness assertion:
// before that story, "zero findings" was this test's entire success
// signal, which is exactly what a scanner that silently evaluated
// nothing would also print (M-5's live D-000.9 exposure). Failing on
// stats.ChecksEvaluated == 0 makes the "I could not look" case
// distinguishable from "I looked and found nothing" — see
// TestAbsencesZeroWitnessIsCaught below for the constructed red proof.
func TestAbsencesProductionScan(t *testing.T) {
	root := repoRootFromTest(t)
	findings, stats, err := ScanAbsences(root)
	if err != nil {
		t.Fatalf("scan repo root: %v", err)
	}
	if stats.ChecksEvaluated == 0 {
		t.Fatalf("ScanAbsences evaluated zero absence checks — coverage witness failed (AC50, D-1.4.11): a scanner that looked at nothing must not report the same 'zero findings' as a healthy run")
	}
	// Story 1.7 (AC25b): the STRONGER form M-12 offered — ChecksEvaluated
	// equals the row count exactly, not merely non-zero — closes the gap
	// a future `continue` guard inside the scan loop would otherwise
	// leave open (today the loop body is unconditional, so the two forms
	// are equivalent; this pins the stronger one so that stops being
	// true only on purpose).
	if stats.ChecksEvaluated != len(absenceChecks) {
		t.Fatalf("AC25b: ChecksEvaluated (%d) must equal the row count (%d) — a skipped row would still read as a healthy non-zero witness otherwise", stats.ChecksEvaluated, len(absenceChecks))
	}
	if len(findings) > 0 {
		t.Fatalf("DW-2 artifact(s) found present at the repo root — wire the matching licence half (AC21):\n%v", findings)
	}
}

// TestAbsencesZeroWitnessIsCaught is AC50's red-proof, by construction:
// it substitutes an empty absenceChecks list (restored via t.Cleanup, so
// no other test in this package observes it) and asserts that the
// resulting AbsencesStats.ChecksEvaluated is 0 — precisely the state
// TestAbsencesProductionScan above now fails loudly on. This is not a
// reverted mutation of production code (D-000.9 obligation 2's forbidden
// shape): it swaps a package-level test double for the duration of one
// test, in a gate that runs by default, and the assertion is on the
// witness's own honesty rather than on any finding.
//
// This test is UNCHANGED by Story 3.7's removal of the content-check
// mechanism (D-000.67 part 1): it is about ChecksEvaluated, the row
// count, which Story 5.1 still owns bringing to zero by removing
// ScanAbsences itself — never by emptying absenceChecks alone (that
// state is exactly what this test proves reads as a healthy green
// pass, not a red one).
func TestAbsencesZeroWitnessIsCaught(t *testing.T) {
	saved := absenceChecks
	absenceChecks = nil
	t.Cleanup(func() { absenceChecks = saved })

	root := repoRootFromTest(t)
	findings, stats, err := ScanAbsences(root)
	if err != nil {
		t.Fatalf("scan repo root with empty check list: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("an empty absenceChecks must produce zero findings, got %v", findings)
	}
	if stats.ChecksEvaluated != 0 {
		t.Fatalf("expected ChecksEvaluated == 0 with an empty absenceChecks list, got %d", stats.ChecksEvaluated)
	}
	// This is exactly the condition TestAbsencesProductionScan's
	// coverage-witness check now rejects — proving the guard is wired,
	// not decorative.
}

// TestAbsencesChecksIncludeTheRemainingEntry closes the shrunk-list gap
// D-1.4.11 warned about and Story 1.4's finisher review (Finding 5,
// Major) confirmed was still open: the witness above proves
// ChecksEvaluated tracks the scanner's OWN loop, but nothing previously
// pinned WHICH check the list holds — deleting the sole remaining entry
// would still leave ChecksEvaluated == 0, which the OTHER test above
// already covers, but a swap-for-a-different-rule-id regression would
// not be caught by either. Renamed a third time as the list shrank
// 5 -> 4 -> 3 -> 2 -> 1 (DW-10's stale-name defect, D-000.37, fixed in
// this same commit): Story 1.7 added a fifth entry as a path check,
// "absence-cmd-dir"; Story 2.1 (D-2.1.x) re-keyed it to a content
// check, "absence-source-date-epoch"; Story 2.2 (AC5) removed
// "absence-fonts-dir"; Story 3.2/3.6 (D-000.59) discharged
// "absence-expr-package" and "absence-diag-package" by replacement;
// and Story 3.7 (AC13, D-000.67 part 1) discharged
// "absence-source-date-epoch" by replacement, removing the entire
// content-check mechanism it was the sole tenant of — bringing this
// list down to its final ONE entry. Per the schedule recorded in
// absences.go's own comment: 3 -> 2 -> 1 (this story) -> 0 (Story 5.1,
// which must remove ScanAbsences and its precondition TOGETHER —
// TestAbsencesZeroWitnessIsCaught above proves the mechanism goes
// SLACK, not LOUD, at an empty list, so reaching zero by decrementing
// this list alone is not a valid path).
func TestAbsencesChecksIncludeTheRemainingEntry(t *testing.T) {
	want := []string{
		"absence-designer-project",
	}
	if len(absenceChecks) != len(want) {
		t.Fatalf("absenceChecks has %d entries, want %d (%v) — an entry was added or removed without updating this pin", len(absenceChecks), len(want), want)
	}
	for _, w := range want {
		found := false
		for _, c := range absenceChecks {
			if c.rule == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("absenceChecks is missing rule id %q — a tripwire was silently deleted (D-1.4.11's shrunk-list hazard)", w)
		}
	}
}

// TestAbsencesFixtureScan is AC21's fixture caller, red-proving RP-10 on
// retained fixture roots at folio-go/testdata/lint/absences/ (never at
// the artifacts' real paths, per AC21's own text: "a fixture root
// containing those files red-proves them without creating them at their
// real paths"). The violating/ fixture's folio-designer/ subtree
// contains both a package-lock.json and a nested
// third-party-notices/pdfjs-dist/NOTICE — Finding 8's fix (Story 1.4's
// QA review) keys the check on the folio-designer/ directory itself, so
// both are caught by the SAME finding; want reflects that (one finding).
//
// Story 3.7 (AC13, D-000.67 part 1) removed this fixture's second
// finding, folio-go/internal/paramsdate/placeholder.go, along with the
// content-check mechanism it existed solely to exercise.
func TestAbsencesFixtureScan(t *testing.T) {
	root := repoRootFromTest(t)
	base := filepath.Join(root, "folio-go", "testdata", "lint", "absences")

	t.Run("violating", func(t *testing.T) {
		got, stats, err := ScanAbsences(filepath.Join(base, "violating"))
		if err != nil {
			t.Fatalf("scan violating/: %v", err)
		}
		if stats.ChecksEvaluated == 0 {
			t.Fatalf("expected a non-zero coverage witness (AC50)")
		}
		want := []Finding{
			{Path: "folio-designer", Rule: "absence-designer-project"},
		}
		assertExactFindings(t, got, want)
	})

	t.Run("compliant", func(t *testing.T) {
		got, stats, err := ScanAbsences(filepath.Join(base, "compliant"))
		if err != nil {
			t.Fatalf("scan compliant/: %v", err)
		}
		if stats.ChecksEvaluated == 0 {
			t.Fatalf("expected a non-zero coverage witness (AC50)")
		}
		if len(got) > 0 {
			t.Fatalf("compliant/ must report zero findings, got %v", got)
		}
	})
}
