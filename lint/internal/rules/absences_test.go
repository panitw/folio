package rules

import (
	"path/filepath"
	"testing"
)

// TestAbsencesProductionScan is AC21's production caller: at the real
// repo root, neither DW-2 directory exists yet, so this must report
// zero findings today — and go red the day either one lands.
//
// Story 1.4 (AC50, D-1.4.11) adds the coverage-witness assertion: before
// this story, "zero findings" was this test's entire success signal,
// which is exactly what a scanner that silently evaluated nothing would
// also print (M-5's live D-000.9 exposure). Failing on
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

// TestAbsencesChecksIncludeAllFiveEntries closes the shrunk-list gap
// D-1.4.11 warned about and this story's finisher review (Finding 5,
// Major) confirmed was still open: the witness above proves
// ChecksEvaluated tracks the scanner's OWN loop, but nothing previously
// pinned WHICH checks the list holds — deleting just the two Story 1.4
// tripwires left ChecksEvaluated == 2 (still non-zero) and the whole
// suite green. This asserts the specific rule ids, so removing any one
// entry — not only emptying the whole list — fails loudly. Story 1.7
// (AC25, D-1.7.7) adds the fifth entry, "absence-cmd-dir".
func TestAbsencesChecksIncludeAllFiveEntries(t *testing.T) {
	want := []string{
		"absence-designer-project",
		"absence-fonts-dir",
		"absence-expr-package",
		"absence-diag-package",
		"absence-cmd-dir",
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
// third-party-notices/pdfjs-dist/NOTICE — Finding 8's fix (this story's
// QA review) keys the check on the folio-designer/ directory itself, so
// both are caught by the SAME finding; want reflects that (two findings,
// one per directory-level check, not three). folio-go/cmd/placeholder.go
// is Finding 10's fix (this story's review, Nit): before it existed,
// only the one-shot RP-8 mutation ever proved absence-cmd-dir fires;
// this makes that proof permanent, alongside the two rows above.
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
			{Path: "folio-go/fonts", Rule: "absence-fonts-dir"},
			{Path: "folio-go/cmd", Rule: "absence-cmd-dir"},
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
