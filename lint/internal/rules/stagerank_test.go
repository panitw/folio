package rules

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestStageRankProductionScan is D-1.3.3's PRODUCTION caller: it points
// the pure checker at the real folio-go/internal/ tree and asserts ZERO
// findings, failing on a scan error separately from, and BEFORE, the
// zero-findings assertion (D-1.3.3 amended, RP-3b — a tree that cannot
// be read must never be silently treated as "zero findings").
//
// D-000.16's own stated precondition for the table is "green today".
// This test is that precondition, executed rather than asserted in
// prose.
//
// VACUITY GUARD (D-000.9), and it reads ScanStageRank's OWN returned
// stats rather than re-walking the tree a second, independent way
// (Major 5's precedent): a dead scanner produces exactly the
// zero-findings all-clear a healthy one does, and a second walk built
// the old way would keep reporting healthy numbers after
// `if true { return nil, stats, nil }` were injected into the checker.
// The guard names the five packages that make this scan non-trivial —
// `layout` and `pagemodel` are the packages this story CREATES, and the
// arrow between `text` and `fontset` is the one D-000.16's ranks were
// corrected for — plus a non-zero count of first-party import edges
// actually compared. `diag` was added by Story 3.6's finisher (Finding
// 14, QA review): AC1 requires `internal/diag`'s zero-first-party-import
// property to be asserted by TWO independent guards (an in-package AST
// scan plus this lint scan); without `diag` named here, `ScanStageRank`
// reporting a violation there was verified to redden, but nothing
// asserted the scan had ever actually ENTERED the package — a future
// change that caused the walk to skip it would go slack silently.
func TestStageRankProductionScan(t *testing.T) {
	root := repoRootFromTest(t)
	internalDir := filepath.Join(root, "folio-go", "internal")

	findings, stats, err := ScanStageRank(internalDir)
	if err != nil {
		t.Fatalf("scan folio-go/internal/: %v", err)
	}

	for _, pkg := range []string{"layout", "pagemodel", "pdf", "text", "fontset", "diag"} {
		if !slices.Contains(stats.PackagesVisited, pkg) {
			t.Fatalf("vacuity guard: the scanner's own stats do not report visiting package %q — a scan that never entered it reports the same zero findings a healthy scan does (D-000.9). Visited: %v",
				pkg, stats.PackagesVisited)
		}
	}
	if stats.FilesSeen == 0 {
		t.Fatal("vacuity guard: the scanner's own stats report 0 .go files parsed under folio-go/internal/")
	}
	if stats.FirstPartyImports == 0 {
		t.Fatal("vacuity guard: the scanner's own stats report 0 first-party internal import edges examined — the rank comparison never ran even once, so zero findings proves nothing")
	}

	if len(findings) > 0 {
		var msgs []string
		for _, f := range findings {
			msgs = append(msgs, f.Message)
		}
		t.Fatalf("stage-rank violations under folio-go/internal/ (D-000.16):\n%s", strings.Join(msgs, "\n"))
	}
}

// TestStageRankFixtureScan is D-1.3.3's FIXTURE caller over the RETAINED
// VIOLATING FIXTURE at folio-go/testdata/lint/stage-rank/ (never under
// folio-go/internal/, F-10). It asserts exactly the named findings BY
// FILE AND RULE, never by count (AC1, RP-3c) — matching neither a subset
// nor a superset, so deleting an expected finding fails on the "expected
// not reported" half and inventing one fails on the "unexpected
// reported" half.
//
// The fixture covers four violation shapes and three compliant shapes:
//
//	layout/       -> pdf        HIGHER rank — AD-5's OWN ARROW, named
//	                            explicitly. D-000.23 read the other way:
//	                            a guard written for a CLASS must still be
//	                            shown to cover the INSTANCE it exists for.
//	expr/         -> bind       HIGHER rank — D-1.6.1's pre-commitment.
//	pagemodel/    -> diag       EQUAL rank — the branch a plain
//	                            "forward-only" reading would miss.
//	unrankedstage/              NO rank at all — the fail-safe default,
//	                            firing with no import present.
//	template/     -> geom       lower rank: NOT reported.
//	pdf/          -> pagemodel  lower rank: NOT reported. This is the
//	                            arrow Story 2.5 actually adds, so the
//	                            table must not accidentally forbid it.
//	(root file)                 rankNoStage with no first-party import:
//	                            NOT reported.
func TestStageRankFixtureScan(t *testing.T) {
	root := repoRootFromTest(t)
	fixtureRoot := filepath.Join(root, "folio-go", "testdata", "lint", "stage-rank")

	got, stats, err := ScanStageRank(fixtureRoot)
	if err != nil {
		t.Fatalf("scan fixture tree %s: %v", fixtureRoot, err)
	}
	if stats.FilesSeen == 0 {
		t.Fatalf("vacuity guard: the scanner's own stats report 0 files parsed under %s — the retained fixture is missing, and an empty tree produces the same empty finding set a compliant one does", fixtureRoot)
	}

	want := []Finding{
		{Path: filepath.Join("layout", "violating_pdf_import.go"), Rule: RuleStageRank},
		{Path: filepath.Join("expr", "violating_bind_import.go"), Rule: RuleStageRank},
		{Path: filepath.Join("pagemodel", "violating_equal_rank.go"), Rule: RuleStageRank},
		{Path: filepath.Join("unrankedstage", "violating_unranked.go"), Rule: RuleStageRank},
	}
	assertExactFindings(t, got, want)
}

// TestStageRankMessageNamesAD5sArrow is the message half of the fixture
// caller: a finding that fires but says nothing useful is a tripwire
// whose remedy nobody can execute (D-000.37 — "a tripwire's failure
// message is executable by a human"). The layout -> pdf finding must
// name both endpoints, both ranks, and the rule that a stage passes what
// a later stage needs rather than importing it.
func TestStageRankMessageNamesAD5sArrow(t *testing.T) {
	root := repoRootFromTest(t)
	fixtureRoot := filepath.Join(root, "folio-go", "testdata", "lint", "stage-rank")

	got, _, err := ScanStageRank(fixtureRoot)
	if err != nil {
		t.Fatalf("scan fixture tree %s: %v", fixtureRoot, err)
	}

	wantPath := filepath.Join("layout", "violating_pdf_import.go")
	var msg string
	for _, f := range got {
		if f.Path == wantPath {
			msg = f.Message
			break
		}
	}
	if msg == "" {
		t.Fatalf("presence precondition: no finding reported for %s — the message assertions below would pass vacuously on an empty string", wantPath)
	}
	for _, want := range []string{`"layout"`, `"pdf"`, "rank 7", "rank 8", "STRICTLY LOWER", "rides on the VALUE"} {
		if !strings.Contains(msg, want) {
			t.Errorf("layout -> pdf finding message does not contain %q; got:\n%s", want, msg)
		}
	}
}

// TestStageRankUnrankedMessageNamesASymbolThatExists is the sibling of
// TestStageRankMessageNamesAD5sArrow, for the OTHER message this rule
// emits.
//
// It exists because the unranked-package message shipped naming
// `stageRanks` — a symbol that appears nowhere in the repository. The
// table is `stageRankTable`. The file path was right, so the cost was
// small, but D-000.37 is exactly the ruling this violated: "a tripwire's
// failure message is executable by a human" and "the remedy text is
// maintained as code, not as prose." Only the layout -> pdf message had
// its content held by a test; this branch had none, so a wrong
// identifier in it was invisible.
//
// So this asserts more than a spelling. It reads stagerank.go's own
// source and requires that the identifier the remedy names is actually
// DECLARED there. A rename of the table that leaves the message behind
// fails here, which a string-equality assertion would not have done —
// it would merely have moved the stale name from the message into the
// test.
func TestStageRankUnrankedMessageNamesASymbolThatExists(t *testing.T) {
	root := repoRootFromTest(t)
	fixtureRoot := filepath.Join(root, "folio-go", "testdata", "lint", "stage-rank")

	got, _, err := ScanStageRank(fixtureRoot)
	if err != nil {
		t.Fatalf("scan fixture tree %s: %v", fixtureRoot, err)
	}

	wantPath := filepath.Join("unrankedstage", "violating_unranked.go")
	var msg string
	for _, f := range got {
		if f.Path == wantPath {
			msg = f.Message
			break
		}
	}
	if msg == "" {
		t.Fatalf("presence precondition: no finding reported for %s — the message assertions below would pass vacuously on an empty string", wantPath)
	}

	// The remedy must say WHERE to act, WHAT to edit, and that silence is
	// not an option.
	const wantSymbol = "stageRankTable"
	for _, want := range []string{
		`"unrankedstage"`,
		"lint/internal/rules/stagerank.go",
		wantSymbol,
		"never a pass",
		"known ranks:",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("unranked-package finding message does not contain %q; got:\n%s", want, msg)
		}
	}

	// ...and the symbol it names must exist. This is the half that would
	// have caught the shipped defect: `stageRanks` satisfied every
	// substring check a human would think to write, and resolved to
	// nothing.
	srcPath := filepath.Join(root, "lint", "internal", "rules", "stagerank.go")
	src, rerr := os.ReadFile(srcPath)
	if rerr != nil {
		t.Fatalf("read %s: %v", srcPath, rerr)
	}
	if !strings.Contains(string(src), "var "+wantSymbol+" = ") {
		t.Fatalf("the remedy names %q, but %s declares no such variable — a tripwire whose remedy points at a symbol that does not exist is not executable by a human (D-000.37)", wantSymbol, srcPath)
	}
}
