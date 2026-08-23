package rules

import (
	"fmt"

	"github.com/panitw/folio/lint/internal/licence"
)

// RuleLicence is this guard's stable rule id (AC4, AC29).
const RuleLicence = "licence"

// ScanLicenceGraph is the AC1 pure checker for AD-26's licence boundary
// (AC18, D-1.3.9): given a Go module's root directory, it resolves the
// module's full dependency graph hermetically (GOPROXY=off, RP-12) and
// reports a finding for every dependency that either carries a forbidden
// licence family (GPL, LGPL, AGPL, SSPL, or a commercial EULA — AD-26's
// Rule) or whose licence cannot be resolved at all (AC19, AC29's
// unknown/ graph — "a silent pass on an unidentifiable licence is the
// realistic failure mode", D-1.3.8). It never warns; every violation is
// a finding, and the caller fails the build on any finding (AC18).
//
// moduleDir IS the target directory here (AC1's seam applied to a module
// graph rather than a source tree): the production caller points it at
// each of the repo's three real module roots in turn; the fixture caller
// points it at lint/testdata/licence/{copyleft,permissive,unknown}/.
func ScanLicenceGraph(moduleDir string) ([]Finding, error) {
	modules, err := licence.ResolveGraph(moduleDir)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	for _, m := range modules {
		text, ok := licence.ReadLicenceText(m.Dir)
		if !ok {
			findings = append(findings, Finding{
				Path: m.Path, Rule: RuleLicence,
				Message: fmt.Sprintf("module %s: licence unresolvable — no LICENSE file found (AC19, AC29)", m.Path),
			})
			continue
		}
		family, spdx := licence.ClassifyLicenceText(text)
		switch family {
		case licence.FamilyCopyleft:
			findings = append(findings, Finding{
				Path: m.Path, Rule: RuleLicence,
				Message: fmt.Sprintf("module %s: forbidden licence %s (copyleft; AD-26 bans GPL, LGPL, AGPL, SSPL, or a commercial EULA at any depth)", m.Path, spdx),
			})
		// A commercial EULA has no dedicated case (Finding 9, this
		// story's QA review): the classifier has no marker table for
		// one, so it falls through to FamilyUnknown below and fails the
		// build there — "licence unresolvable" — rather than through a
		// case that was never actually reachable.
		case licence.FamilyUnknown:
			findings = append(findings, Finding{
				Path: m.Path, Rule: RuleLicence,
				Message: fmt.Sprintf("module %s: licence unresolvable — could not classify licence text (AC19, AC29)", m.Path),
			})
		case licence.FamilyPermissive:
			// fine; no finding.
		}
	}
	return findings, nil
}
