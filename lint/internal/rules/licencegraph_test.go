package rules

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestLicenceGraphProductionScan is AC18's production caller: it walks
// each of the repo's three real Go modules — folio-go, hashmatrix, lint
// itself — and asserts zero findings, uniformly (D-1.3.9: AD-26's scope
// line is "Binds: all"; a checker covering its own dependency is
// correct, not a bootstrap problem, D-1.3.6/D-1.3.9).
func TestLicenceGraphProductionScan(t *testing.T) {
	root := repoRootFromTest(t)
	for _, mod := range []string{"folio-go", "hashmatrix", "lint"} {
		mod := mod
		t.Run(mod, func(t *testing.T) {
			findings, err := ScanLicenceGraph(filepath.Join(root, mod))
			if err != nil {
				t.Fatalf("scan %s module graph: %v", mod, err)
			}
			if len(findings) > 0 {
				var msgs []string
				for _, f := range findings {
					msgs = append(msgs, f.Message)
				}
				t.Fatalf("forbidden or unresolvable licence(s) found in %s's module graph (AD-26):\n%s", mod, strings.Join(msgs, "\n"))
			}
		})
	}
}

// TestLicenceGraphFixtureScan is AC29's fixture caller: copyleft/ fails
// naming the module and licence for each of its four sibling stubs;
// permissive/ passes; unknown/ — no LICENSE file at all — fails as
// unresolvable (RP-13, V11).
func TestLicenceGraphFixtureScan(t *testing.T) {
	root := repoRootFromTest(t)
	base := filepath.Join(root, "lint", "testdata", "licence")

	t.Run("copyleft", func(t *testing.T) {
		findings, err := ScanLicenceGraph(filepath.Join(base, "copyleft"))
		if err != nil {
			t.Fatalf("scan copyleft/ graph: %v", err)
		}
		want := []Finding{
			{Path: "example.test/gpl-lib", Rule: RuleLicence},
			{Path: "example.test/lgpl-lib", Rule: RuleLicence},
			{Path: "example.test/agpl-lib", Rule: RuleLicence},
			{Path: "example.test/sspl-lib", Rule: RuleLicence},
		}
		assertExactFindings(t, findings, want)
	})

	t.Run("permissive", func(t *testing.T) {
		findings, err := ScanLicenceGraph(filepath.Join(base, "permissive"))
		if err != nil {
			t.Fatalf("scan permissive/ graph: %v", err)
		}
		if len(findings) > 0 {
			t.Fatalf("permissive/ graph must PASS with zero findings, got %v", findings)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		findings, err := ScanLicenceGraph(filepath.Join(base, "unknown"))
		if err != nil {
			t.Fatalf("scan unknown/ graph: %v", err)
		}
		want := []Finding{
			{Path: "example.test/mystery-lib", Rule: RuleLicence},
		}
		assertExactFindings(t, findings, want)
	})
}
