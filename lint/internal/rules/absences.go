package rules

import (
	"os"
	"path/filepath"
)

// absenceCheck is one artifact this guard asserts is absent — on
// purpose, per D-1.3.4: the licence check's Go half ships complete at
// this story, but a dependent half has nothing to check yet. Asserting
// its absence, rather than a conditional "check it if present", means
// it goes red the day the directory lands, forcing the matching half
// to be wired before the build can pass again — "a conditional check
// starts silently passing the moment the directory arrives" (D-1.3.4).
//
// Story 3.7 (D-000.67 part 1, AC13) removed this file's SECOND check
// kind — a content check that failed when a forbidden literal string
// appeared in a .go file under a scoped directory — together with its
// own witness field (AbsencesStats.ContentFilesScanned) and both
// fixtures that existed only to exercise it. It was never relaxed to
// reach zero tenants: it was REMOVED, in the same commit as the
// positive assertions that replace the one obligation it still carried
// (DW-10's part (a), "cmd/folio render reads SOURCE_DATE_EPOCH and
// passes it in as a parameter" — see main_subprocess_test.go's
// TestRenderReadsSourceDateEpochFromEnvironment, a genuine subprocess
// test, which is what makes it the valid replacement; QA Finding 9,
// this story's review, corrected this citation from "main_test.go").
// D-000.67 part 1,
// ruled canonically from this exact removal: "A mechanism can carry
// more than one presence precondition, each keyed on a different
// population. Check every witness the mechanism reports, not the one
// the roadmap made you think of." The content check's population was
// tracked by TWO witnesses — ChecksEvaluated (rows) and
// ContentFilesScanned (files opened) — and only the roadmap's own
// forward note tracked the first; removing the row drove the second to
// zero HERE, at Story 3.7, one epic before the recorded 3→2→1→0
// schedule below said it would.
//
// THE SCHEDULE: this list started at 3 entries, Story 3.6 dropped it to
// 2, Story 3.7 drops it to 1 (this commit), and Story 5.1 must drop it
// to 0 by REMOVING ScanAbsences and its precondition TOGETHER, never by
// decrementing absenceChecks to an empty slice and leaving the
// mechanism running: absences_test.go's own
// TestAbsencesZeroWitnessIsCaught proves that an empty absenceChecks
// yields zero findings and ChecksEvaluated == 0 — a GREEN pass, not a
// red one. The tripwire goes SLACK at zero, not LOUD, so "just delete
// the remaining row" is not a valid way to reach zero; Story 5.1 owns
// removing the scanner itself. The kind removed by this story is NOT
// kept "for later" — an empty check kind is precisely the
// zero-candidate scan D-000.67 exists to prevent, and no tenant is
// scheduled for it.
type absenceCheck struct {
	relPath string // the path that must not exist
	rule    string
	desc    string
}

// absenceChecks names the artifacts DW-2 defers, each with its owning
// story (deferred-work.md). Keyed on a DIRECTORY, not a guessed filename
// (Finding 8, Story 1.4's QA review — a Major, proved by construction):
// the original check matched exactly
// "folio-designer/package-lock.json", which a pnpm-lock.yaml or a
// yarn.lock would have passed straight through with zero findings —
// D-1.3.4's own rejected hazard ("the guard reports success precisely
// when it stops covering anything") arriving through a side door the
// exact filename never anticipated. folio-designer/ absent catches ANY
// artifact landing under it — including the package-lock.json shape and
// the third-party-notices/pdfjs-dist/NOTICE shape a separate check
// originally existed for; both live inside folio-designer/, so the
// directory-level check is strictly broader, not narrower, than the two
// checks it replaces.
//
// Story 2.2 (AC5) removed this list's other original entry,
// "absence-fonts-dir" (folio-go/fonts/ required to be ABSENT): Story 2.2
// shipped faces there, so that tripwire's job is done, and it is
// replaced by ScanFontsAssets (fontsassets.go) — a fail-closed guard
// with the OPPOSITE polarity, requiring folio-go/fonts/ to exist and
// hold only recognised shapes, now that it is expected to be present.
//
// "absence-diag-package" (folio-go/internal/diag/ absent) was
// DISCHARGED BY REPLACEMENT by Story 3.6 (R6, AC2, D-000.59): deleted
// in the same commit as the positive assertion that replaces it
// (internal/diag's own TestRegistryIsAdditiveOnly). "absence-cmd-dir",
// re-keyed by Story 2.1 to the content check "absence-source-date-epoch",
// was itself DISCHARGED BY REPLACEMENT by Story 3.7 (AC13, D-000.67
// part 1, this commit) alongside the entire content-check mechanism —
// see this file's own package-level doc comment above.
//
// The exact real-world path each future story lands its artifacts at
// beneath these directories is that story's own call; this guard's job
// is only to fail loudly the day the directory appears at all.
var absenceChecks = []absenceCheck{
	{
		relPath: "folio-designer",
		rule:    "absence-designer-project",
		desc:    "folio-designer/ must be absent until Story 5.1 creates the project and wires the JS licence half (DW-2)",
	},
}

// AbsencesStats reports what ScanAbsences actually examined, from the
// scanner's own execution (Story 1.4, AC50, D-1.4.11) — see
// MapRangeStats' doc comment for why a second, independently-derived
// walk cannot be trusted as a vacuity guard. ChecksEvaluated is
// incremented once per entry in absenceChecks actually iterated, so a
// caller that (by mistake, or by a future refactor) never reaches the
// loop body reports zero and the production test fails loudly instead
// of passing on identical output to a healthy run.
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
// rule id for the artifact that fired.
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
