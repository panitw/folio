package rules

import (
	"path/filepath"
	"testing"
)

// TestAbsencesProductionScan is AC21's production caller: at the real
// repo root, neither DW-2 directory exists yet, so this must report
// zero findings today — and go red the day either one lands.
func TestAbsencesProductionScan(t *testing.T) {
	root := repoRootFromTest(t)
	findings, err := ScanAbsences(root)
	if err != nil {
		t.Fatalf("scan repo root: %v", err)
	}
	if len(findings) > 0 {
		t.Fatalf("DW-2 artifact(s) found present at the repo root — wire the matching licence half (AC21):\n%v", findings)
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
// one per directory-level check, not three).
func TestAbsencesFixtureScan(t *testing.T) {
	root := repoRootFromTest(t)
	base := filepath.Join(root, "folio-go", "testdata", "lint", "absences")

	t.Run("violating", func(t *testing.T) {
		got, err := ScanAbsences(filepath.Join(base, "violating"))
		if err != nil {
			t.Fatalf("scan violating/: %v", err)
		}
		want := []Finding{
			{Path: "folio-designer", Rule: "absence-designer-project"},
			{Path: "folio-go/fonts", Rule: "absence-fonts-dir"},
		}
		assertExactFindings(t, got, want)
	})

	t.Run("compliant", func(t *testing.T) {
		got, err := ScanAbsences(filepath.Join(base, "compliant"))
		if err != nil {
			t.Fatalf("scan compliant/: %v", err)
		}
		if len(got) > 0 {
			t.Fatalf("compliant/ must report zero findings, got %v", got)
		}
	})
}
