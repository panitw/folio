package rules

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// absenceCheck is one artifact this guard asserts is absent — on
// purpose, per D-1.3.4: the licence check's Go half ships complete at
// this story, but two dependent halves have nothing to check yet.
// Asserting their absence, rather than a conditional "check it if
// present", means each one goes red the day it lands, forcing the
// matching half to be wired before the build can pass again — "a
// conditional check starts silently passing the moment the directory
// arrives" (D-1.3.4).
//
// Story 2.1 (D-2.1.x) adds a second check KIND alongside the original
// path-existence one: a content check, which fails when a forbidden
// literal string appears anywhere in a .go file under a scoped
// directory, rather than when a path exists. This followed a measured
// false positive: the original "folio-go/cmd/ must be absent" row
// (below) was keyed on a PATH as a proxy for its real purpose — forcing
// AD-7's params-date wiring to be settled when the CLI that reads
// SOURCE_DATE_EPOCH arrives (DW-10) — and fired on Story 2.1's own
// build-time tooling (cmd/gentrie, cmd/gencorpus), which is a
// legitimate second tenant of cmd/ with nothing to do with DW-10. The
// general rule this produced: key a guard on its purpose, not on a
// proxy for its purpose — where the key is broader than the purpose,
// the gap is where false positives live, and a false positive in a
// guard invites exactly the workaround (weakening the guard) that
// erodes it fastest.
type absenceCheckKind int

const (
	// absenceKindPath fails when relPath exists under the scanned root.
	absenceKindPath absenceCheckKind = iota
	// absenceKindContent fails when forbidden appears in any .go file
	// under scopeRelDir (recursively, "testdata" subtrees excluded —
	// the same exclusion ScanEmbedFont/ResolveAssets use, so a lint
	// fixture exercising this very check does not also trip the
	// production scan of the real tree it lives inside).
	absenceKindContent
)

type absenceCheck struct {
	kind        absenceCheckKind
	relPath     string // kind == absenceKindPath: the path that must not exist
	scopeRelDir string // kind == absenceKindContent: directory to scan for .go files
	forbidden   string // kind == absenceKindContent: literal substring that must not appear
	rule        string
	desc        string
}

// absenceChecks names the artifacts DW-2 defers, each with its owning
// story (deferred-work.md). Both are keyed on a DIRECTORY, not a guessed
// filename (Finding 8, this story's QA review — a Major, proved by
// construction): the original checks matched exactly
// "folio-designer/package-lock.json" and "folio-go/fonts/OFL.txt",
// which a pnpm-lock.yaml, a yarn.lock, or an OFL-1.1.txt would each pass
// straight through with zero findings — D-1.3.4's own rejected hazard
// ("the guard reports success precisely when it stops covering
// anything") arriving through a side door the exact filename never
// anticipated. folio-designer/ absent catches ANY artifact landing
// under it — including the package-lock.json shape and the
// third-party-notices/pdfjs-dist/NOTICE shape the original third check
// existed for separately; both live inside folio-designer/, so the
// directory-level check is strictly broader, not narrower, than the two
// checks it replaces. folio-go/fonts/ absent is the same move for the
// OFL licence text: Story 2.2 cannot ship a single face without
// creating that directory first.
//
// The exact real-world path each future story lands its artifacts at
// beneath these directories is that story's own call; this guard's job
// is only to fail loudly the day either directory appears at all.
var absenceChecks = []absenceCheck{
	{
		relPath: "folio-designer",
		rule:    "absence-designer-project",
		desc:    "folio-designer/ must be absent until Story 5.1 creates the project and wires the JS licence half (DW-2)",
	},
	{
		relPath: "folio-go/fonts",
		rule:    "absence-fonts-dir",
		desc:    "folio-go/fonts/ must be absent until Story 2.2 ships faces and wires the OFL 1.1 licence text (DW-2)",
	},
	// The two entries below are Story 1.4's tripwires (AC48, AC49,
	// D-1.4.2), added AFTER AC50/D-1.4.11 closed ScanAbsences' own
	// coverage-witness gap (T2's ordering is load-bearing — "adding
	// entries to a list whose emptiness passes is exactly how a
	// tripwire becomes decorative"). Keyed on the DIRECTORY, per DW-2's
	// recorded correction (an exact filename guess is a false-pass
	// hazard a directory-level check does not have).
	{
		relPath: "folio-go/internal/expr",
		rule:    "absence-expr-package",
		desc:    "folio-go/internal/expr/ must be absent until Story 3.2 derives columns[].footerOf from bind (DW-5)",
	},
	{
		relPath: "folio-go/internal/diag",
		rule:    "absence-diag-package",
		desc:    "folio-go/internal/diag/ must be absent until Story 3.6 mints TABLE_FOOTER_SOURCE_UNRESOLVED/TABLE_FOOTER_SOURCE_FORBIDDEN and AD-14's other diagnostic codes (DW-6, D-1.4.2)",
	},
	// Story 1.7's tripwire (AC25, D-1.7.7), RE-KEYED by Story 2.1
	// (D-2.1.x, this row's disposition recorded in deferred-work.md's
	// DW-10 entry): originally keyed on the folio-go/cmd/ PATH existing
	// at all, as a proxy for "the CLI that reads SOURCE_DATE_EPOCH has
	// arrived". That key was broader than the purpose — cmd/ has more
	// than one legitimate tenant, and Story 2.1's own build-time tooling
	// (cmd/gentrie, cmd/gencorpus) tripped it despite having nothing to
	// do with AD-7 or params-date wiring. Re-keyed on the trigger DW-10
	// actually cares about: SOURCE_DATE_EPOCH appearing in Go source
	// under folio-go/ at all — which is what the CLI Story 3.7 builds
	// will actually introduce, regardless of what its cmd/ subpackage
	// is named. Still forces AD-7's params-date wiring to be settled
	// rather than remembered (DW-10, owner Story 3.7).
	{
		kind:        absenceKindContent,
		scopeRelDir: "folio-go",
		forbidden:   "SOURCE_DATE_EPOCH",
		rule:        "absence-source-date-epoch",
		desc:        "SOURCE_DATE_EPOCH must not appear in any Go source under folio-go/ until Story 3.7 builds the CLI and settles AD-7's params-date wiring (DW-10)",
	},
}

// AbsencesStats reports what ScanAbsences actually examined, from the
// scanner's own execution (Story 1.4, AC50, D-1.4.11) — see
// MapRangeStats' doc comment for why a second, independently-derived
// walk cannot be trusted as a vacuity guard. Before this story
// ScanAbsences returned only ([]Finding, error): with absenceChecks
// empty (its state before Story 1.3's DW-2 entries existed, and — the
// live D-000.9 exposure M-5 measured — its state again the moment a
// caller ever passed an empty list) TestAbsencesProductionScan asserted
// only len(findings) == 0, which is satisfied identically whether the
// scanner considered two checks or zero. ChecksEvaluated closes that:
// it is incremented once per entry in absenceChecks actually iterated,
// so a caller that (by mistake, or by a future refactor) never reaches
// the loop body reports zero and the production test — updated in this
// same story, AC50 — fails loudly instead of passing on identical
// output to a healthy run.
type AbsencesStats struct {
	ChecksEvaluated int
	// ContentFilesScanned is Story 2.1's reopening addition (Finding 13):
	// ChecksEvaluated counts ROWS, not WORK. A content-kind row whose
	// scope directory does not exist (or exists but holds zero .go
	// files) returns zero findings identically to a healthy scan of a
	// populated tree — exactly the D-000.9 shape WordlistAssetsStats
	// already guards against with LocationExists/FilesSeen. This counts
	// every .go file any content-kind row actually opened and read,
	// across all rows, so "zero content findings" is distinguishable
	// from "the content check never scanned anything".
	ContentFilesScanned int
}

// ScanAbsences is the AC1 pure checker for AC21's asserted absences. It
// takes the repository root as its target directory (real, or a fixture
// root standing in for it) and reports a finding for every artifact
// found present. A conditional "check it if present" is explicitly
// rejected (D-1.3.4) — this checker never has a "skip if missing"
// branch; every check always runs and either finds nothing (the
// expected state today) or fails. Each finding carries the specific
// rule id for the artifact that fired (Finding 10, this story's QA
// review): the field was previously declared per check and then
// discarded, with every finding emitted under one generic rule id
// instead, degrading AC1/AC4's "by file and rule" assertion to "by file
// alone".
func ScanAbsences(root string) ([]Finding, AbsencesStats, error) {
	var findings []Finding
	var stats AbsencesStats
	for _, c := range absenceChecks {
		stats.ChecksEvaluated++
		switch c.kind {
		case absenceKindContent:
			found, filesScanned, err := scanForbiddenContent(root, c)
			if err != nil {
				return nil, stats, err
			}
			stats.ContentFilesScanned += filesScanned
			findings = append(findings, found...)
		default:
			full := filepath.Join(root, filepath.FromSlash(c.relPath))
			if _, err := os.Stat(full); err == nil {
				findings = append(findings, Finding{
					Path: c.relPath, Rule: c.rule,
					Message: c.relPath + ": must be absent — " + c.desc,
				})
			}
		}
	}
	return findings, stats, nil
}

// scanForbiddenContent implements absenceKindContent: it walks
// root/c.scopeRelDir for .go files and reports a Finding for each one
// containing c.forbidden as a literal substring. "testdata" subtrees
// are excluded — the same exclusion ScanEmbedFont/ResolveAssets apply —
// so a lint fixture that deliberately red-proves this very check (e.g.
// folio-go/testdata/lint/absences/violating/) does not also trip the
// production scan of the real tree it is nested inside. A scope
// directory that does not exist in the given root (a minimal fixture
// root with no such subtree) is not an error — there is simply nothing
// to scan.
func scanForbiddenContent(root string, c absenceCheck) ([]Finding, int, error) {
	scopeDir := filepath.Join(root, filepath.FromSlash(c.scopeRelDir))
	if _, err := os.Stat(scopeDir); err != nil {
		return nil, 0, nil
	}

	var findings []Finding
	filesScanned := 0
	walkErr := filepath.WalkDir(scopeDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		filesScanned++
		if strings.Contains(string(b), c.forbidden) {
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				return rerr
			}
			rel = filepath.ToSlash(rel)
			findings = append(findings, Finding{
				Path: rel, Rule: c.rule,
				Message: rel + ": contains " + c.forbidden + " — " + c.desc,
			})
		}
		return nil
	})
	if walkErr != nil {
		return nil, filesScanned, walkErr
	}
	return findings, filesScanned, nil
}
