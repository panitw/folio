package rules

import (
	"os"
	"path/filepath"
)

// absenceCheck is one artifact this guard asserts is absent — on
// purpose, per D-1.3.4: the licence check's Go half ships complete at
// this story, but two dependent halves have nothing to check yet.
// Asserting their absence, rather than a conditional "check it if
// present", means each one goes red the day it lands, forcing the
// matching half to be wired before the build can pass again — "a
// conditional check starts silently passing the moment the directory
// arrives" (D-1.3.4).
type absenceCheck struct {
	relPath string
	rule    string
	desc    string
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
	// Story 1.7's tripwire (AC25, D-1.7.7): registered AFTER re-
	// confirming live that ScanAbsences' coverage witness (AbsencesStats
	// .ChecksEvaluated, TestAbsencesProductionScan's hard failure on
	// ChecksEvaluated == 0, TestAbsencesZeroWitnessIsCaught) is still
	// wired — D-1.4.11's sequencing precondition. Keyed on the
	// DIRECTORY, per DW-2's recorded correction. cmd/folio — the only
	// component NFR1.f names as reading SOURCE_DATE_EPOCH — cannot be
	// created without tripping this, which is exactly the mechanical
	// trigger D-1.7.7 relies on to force AD-7's params-date wiring to
	// be settled rather than remembered (DW-10, owner Story 3.7).
	{
		relPath: "folio-go/cmd",
		rule:    "absence-cmd-dir",
		desc:    "folio-go/cmd/ must be absent until Story 3.7 builds the CLI and settles AD-7's params-date wiring (DW-10)",
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
		full := filepath.Join(root, filepath.FromSlash(c.relPath))
		if _, err := os.Stat(full); err == nil {
			findings = append(findings, Finding{
				Path: c.relPath, Rule: c.rule,
				Message: c.relPath + ": must be absent — " + c.desc,
			})
		}
	}
	return findings, stats, nil
}
