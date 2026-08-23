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
func ScanAbsences(root string) ([]Finding, error) {
	var findings []Finding
	for _, c := range absenceChecks {
		full := filepath.Join(root, filepath.FromSlash(c.relPath))
		if _, err := os.Stat(full); err == nil {
			findings = append(findings, Finding{
				Path: c.relPath, Rule: c.rule,
				Message: c.relPath + ": must be absent — " + c.desc,
			})
		}
	}
	return findings, nil
}
