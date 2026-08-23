package rules

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// TestForbiddenImportsProductionScan is the AC1 production caller: it
// scans the real folio-go/internal/ tree and asserts zero findings,
// failing on a scan error separately from, and before, the zero-findings
// assertion (AC5, RP-3b). Non-vacuous per AC12's requirement that "the
// production scan over the real tree is green on the shipped suite" —
// internal/pdf/numbers_test.go and internal/geom/scale_test.go both
// import math today (F-6), so a scan that visited nothing would not
// exercise the selector logic at all. The vacuity guard reads the
// scanner's OWN reported ForbiddenImportsStats (Major 5, this story's
// QA review), not a second, independently-derived walk — see
// MapRangeStats' doc comment for why a second walk cannot be trusted:
// injecting a dead first statement into ScanForbiddenImports would zero
// out its own reported stats but never touch an unrelated walk built the
// old way.
func TestForbiddenImportsProductionScan(t *testing.T) {
	root := repoRootFromTest(t)
	internalDir := filepath.Join(root, "folio-go", "internal")

	findings, stats, err := ScanForbiddenImports(internalDir)
	if err != nil {
		t.Fatalf("scan folio-go/internal/: %v", err)
	}

	if stats.FilesSeen == 0 {
		t.Fatal("vacuity guard: scanner's own stats report 0 files seen under internal/")
	}
	filesWithMathImport := countFilesImporting(t, internalDir, "math")
	if filesWithMathImport == 0 {
		t.Fatal("vacuity guard: expected at least one file under internal/ importing \"math\" (F-6) — the production scan did not exercise the selector logic")
	}

	if len(findings) > 0 {
		var msgs []string
		for _, f := range findings {
			msgs = append(msgs, f.Message)
		}
		t.Fatalf("forbidden imports/selectors found under internal/ (AD-1):\n%s", strings.Join(msgs, "\n"))
	}
}

func countFilesImporting(t *testing.T, root, importPath string) int {
	t.Helper()
	count := 0
	err := walkGoFiles(root, func(rel string, file *ast.File, fset *token.FileSet) error {
		for _, imp := range file.Imports {
			if strings.Trim(imp.Path.Value, `"`) == importPath {
				count++
				break
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return count
}

// TestForbiddenImportsFixtureScan is the AC1 fixture caller, red-proving
// AC13's nine fixtures at folio-go/testdata/lint/forbidden-imports/
// (never under folio-go/internal/, F-10): a non-test file and a
// `_test.go` file each importing time (D-1.3.1: two, not one); a call to
// a math function outside the seven; a reference to math.Pi; and a file
// importing subpackages of three banned paths (Finding 6, this story's
// QA review: math/rand/v2, net/http, net/url, os/exec) — all reported —
// plus a `_test.go` file using the exemption's exact four imports, a
// file using math.MaxInt64, a file calling math.Abs, and a file whose
// comments name math.Round/math.MinInt64 (reproducing
// internal/geom/scale.go:31's shape) — none reported.
func TestForbiddenImportsFixtureScan(t *testing.T) {
	root := repoRootFromTest(t)
	fixtureRoot := filepath.Join(root, "folio-go", "testdata", "lint", "forbidden-imports")

	got, _, err := ScanForbiddenImports(fixtureRoot)
	if err != nil {
		t.Fatalf("scan fixture tree %s: %v", fixtureRoot, err)
	}

	want := []Finding{
		{Path: "violating_time_import.go", Rule: RuleForbiddenImports},
		{Path: "violating_time_import_test.go", Rule: RuleForbiddenImports},
		{Path: "violating_subpackage_imports.go", Rule: RuleForbiddenImports},
		{Path: "violating_math_call.go", Rule: RuleMathSelector},
		{Path: "violating_math_pi.go", Rule: RuleMathSelector},
	}
	assertExactFindings(t, got, want)
}

// TestForbiddenImportsMessageContent is AC10's message-content
// assertion (Finding 11, this story's QA review): AC10 requires the
// failure to name the file, the offending import or selector, and AD-1's
// allow-listed numeric surface, but no test previously inspected
// Finding.Message or Finding.Line for this rule — assertExactFindings
// discards both. Each literal below is written independently of the
// production constants it is checked against (the same discipline
// Finding 2's fix established for EscapeHatch), so a drift in either the
// message template or allowedNumericSurface reddens this test.
func TestForbiddenImportsMessageContent(t *testing.T) {
	const wantSurface = "+ - * /, comparison, and Sqrt, Floor, Ceil, Round, Trunc, Abs, Mod"
	if allowedNumericSurface != wantSurface {
		t.Fatalf("allowedNumericSurface drifted from the literal this test independently spells out:\n got:  %q\n want: %q", allowedNumericSurface, wantSurface)
	}

	root := repoRootFromTest(t)
	fixtureRoot := filepath.Join(root, "folio-go", "testdata", "lint", "forbidden-imports")

	got, _, err := ScanForbiddenImports(fixtureRoot)
	if err != nil {
		t.Fatalf("scan fixture tree %s: %v", fixtureRoot, err)
	}

	messageFor := func(path, rule string) string {
		t.Helper()
		for _, f := range got {
			if f.Path == path && f.Rule == rule {
				return f.Message
			}
		}
		t.Fatalf("no finding for path=%s rule=%s", path, rule)
		return ""
	}

	// Banned import: names the file, the import, and the allow-listed surface.
	bannedMsg := messageFor("violating_time_import.go", RuleForbiddenImports)
	if !strings.Contains(bannedMsg, "violating_time_import.go") {
		t.Errorf("banned-import message missing file name: %q", bannedMsg)
	}
	if !strings.Contains(bannedMsg, `"time"`) {
		t.Errorf("banned-import message missing offending import: %q", bannedMsg)
	}
	if !strings.Contains(bannedMsg, wantSurface) {
		t.Errorf("banned-import message missing allow-listed numeric surface: %q", bannedMsg)
	}

	// math call outside the seven: names the file, the selector, and the surface.
	callMsg := messageFor("violating_math_call.go", RuleMathSelector)
	if !strings.Contains(callMsg, "violating_math_call.go") {
		t.Errorf("math-call message missing file name: %q", callMsg)
	}
	if !strings.Contains(callMsg, wantSurface) {
		t.Errorf("math-call message missing allow-listed numeric surface: %q", callMsg)
	}

	// math.Pi non-call reference: this exact clause was unmet before
	// Finding 11's fix — the non-call message previously omitted the
	// allow-listed numeric surface entirely.
	piMsg := messageFor("violating_math_pi.go", RuleMathSelector)
	if !strings.Contains(piMsg, "violating_math_pi.go") {
		t.Errorf("math.Pi message missing file name: %q", piMsg)
	}
	if !strings.Contains(piMsg, "math.Pi") {
		t.Errorf("math.Pi message missing offending selector: %q", piMsg)
	}
	if !strings.Contains(piMsg, wantSurface) {
		t.Errorf("math.Pi message missing allow-listed numeric surface (Finding 11): %q", piMsg)
	}
}
