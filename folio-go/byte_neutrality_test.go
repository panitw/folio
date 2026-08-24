package folio_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Story 2.3a, AC6 and AC7. This file asserts the two claims the story
// makes about what it did NOT change, in the form D-000.21 (sharpened)
// requires: assert on the artifact that carries the property, and PROVE
// it carries it.
//
// The claims are separable, and both are needed:
//
//   - AC6, byte-neutrality. Story 2.3a swapped a fractional vendor
//     accessor for the integer the same table carries. Measured over all
//     37,312 glyphs of the four committed faces, the two agree with 0
//     mismatches (internal/fontset/vendor-boundary.md, Table 4), so no
//     width, no /W array, no content stream and no golden moves.
//   - AC7, no third Epic 2 gate obligation. The gate owed exactly two
//     things at 431a6a5 — the four-target matrix legs, and D-2.3.5's
//     Thai sign-off — and owes exactly those two now.
//
// WHERE THE RENDERING HALF LIVES. Each fixture's own test already
// renders it and compares the produced bytes against its committed
// digest (fixture_test.go for minimal-rect, font-text, image-embed and
// multi-script-fallback; shaped_fixture_test.go for shaped-text). This
// file does NOT duplicate that work. What it adds is the half those
// tests cannot provide on their own: that the digests they compare
// against are still the ones this story inherited. A re-recording that
// moved a golden AND its expected.json together would leave every one of
// those tests green.

// fixtureDigests pins each fixture's committed sha256 as a SECOND,
// INDEPENDENT literal, written out here rather than read from the file
// this test checks. That is deliberate and is D-2.3.4's discipline: the
// cheapest available response to a hash mismatch is to re-record, and
// re-recording is invisible in a diff when only one copy of the number
// exists. With two, a re-record must change this file too — and if a
// golden is ever legitimately re-recorded, that is exactly the
// conversation the second literal forces.
//
// fixtures/shaped-text's digest is the load-bearing one. D-2.3.5's Thai
// semantic sign-off is bound to THESE EXACT BYTES: the sign-off record
// must name this digest, and re-recording would invalidate a human
// sign-off that has not even been given yet.
var fixtureDigests = []struct {
	dir    string
	sha256 string
}{
	{"minimal-rect", "0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c"},
	{"font-text", "5b60997e232762e452ccae39d652120396339a84a0eb8724c3e226763ebf338e"},
	{"image-embed", "e5778eb872c98ec4a3c3c89466a8313cf52931b896701de8a43f3506abe689fc"},
	{"multi-script-fallback", "20f3388a87216bc9731b46952a2abb00f6a93c6857f26df64bcee660eb15dbae"},
	{"shaped-text", "5964aad0e696010c6e3f34a48d0775af6ae527a6cbe2f5c6319158f43c92e00f"},
}

// TestStory23aMovedNoGoldenDigest is AC6's assertion. Presence comes
// first, at three levels, because "the digest matches" is trivially true
// of a fixture that is not there: the directory must exist, its
// expected.json must exist and parse, and it must carry a sha256 field
// that is a JSON string of the right shape. Only then is the value
// compared. A test that silently passed over a missing fixture would
// prove nothing at all.
func TestStory23aMovedNoGoldenDigest(t *testing.T) {
	root := repoRootForByteNeutrality(t)
	fixturesDir := filepath.Join(root, "fixtures")

	if len(fixtureDigests) == 0 {
		t.Fatal("vacuity guard: the pinned digest list is empty, so this test asserts nothing")
	}

	for _, fx := range fixtureDigests {
		dir := filepath.Join(fixturesDir, fx.dir)
		info, derr := os.Stat(dir)
		if derr != nil || !info.IsDir() {
			t.Errorf("presence precondition: fixture directory %s is missing (%v) — a byte-neutrality claim about a fixture that is not there is vacuous", dir, derr)
			continue
		}

		path := filepath.Join(dir, "expected.json")
		body, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Errorf("presence precondition: %s could not be read: %v", path, rerr)
			continue
		}
		if len(body) == 0 {
			t.Errorf("presence precondition: %s is empty", path)
			continue
		}

		var raw map[string]any
		if jerr := json.Unmarshal(body, &raw); jerr != nil {
			t.Errorf("presence precondition: %s is not valid JSON: %v", path, jerr)
			continue
		}
		field, present := raw["sha256"]
		if !present {
			t.Errorf("presence precondition: %s carries no \"sha256\" field — the property this test asserts does not live in this artifact", path)
			continue
		}
		got, isString := field.(string)
		if !isString {
			t.Errorf("presence precondition: %s's \"sha256\" is %T, not a JSON string — a widened per-target object would make a single-value comparison meaningless", path, field)
			continue
		}
		if len(got) != 64 || strings.ToLower(got) != got {
			t.Errorf("presence precondition: %s's \"sha256\" is %q, which is not 64 lowercase hex characters", path, got)
			continue
		}

		if got != fx.sha256 {
			t.Errorf(
				"fixtures/%s/expected.json's digest is %s, but Story 2.3a inherited %s.\n"+
					"Story 2.3a asserts it changed NO output byte: the fractional vendor accessor it removed and the "+
					"integer table read that replaced it agree on all 37,312 glyphs of the four committed faces "+
					"(internal/fontset/vendor-boundary.md, Table 4).\n"+
					"If this digest moved, either that premise is false or a golden was re-recorded. Do not update this "+
					"literal to make the test pass — that is the move AD-21 and D-000.22 exist to prevent.",
				fx.dir, got, fx.sha256)
		}
	}
}

// TestStory23aAddedNoThirdEpic2GateObligation is AC7's assertion, stated
// mechanically rather than as an intention.
//
// The Epic 2 gate owes exactly two things and must not be given a third:
// the four-target matrix legs, and D-2.3.5's Thai sign-off. The sign-off
// is tracked as a DELIBERATELY RED gate-run test
// (TestShapedTextThaiSemanticSignOffIsRecorded, //go:build matrix),
// which stays red until fixtures/shaped-text/thai-signoff.json names a
// reader, a date, what they examined, and the digest above. Story 2.3a
// neither creates that file nor touches that test.
//
// Under the per-epic heavy-test cadence (D-000.4) this story writes no
// matrix leg and runs none, and there is nothing to register: AC6
// asserts every output byte is unchanged, so the four-target byte
// identity Story 2.3 established still holds by construction and a new
// leg would be a gate obligation with no question behind it.
func TestStory23aAddedNoThirdEpic2GateObligation(t *testing.T) {
	root := repoRootForByteNeutrality(t)

	// (a) The sign-off obligation is still outstanding. If this file
	//     appears, the Thai sign-off has been given — which is a real
	//     event that must be recorded deliberately, by D-2.3.5's owner,
	//     never as a side effect of another story.
	signoff := filepath.Join(root, "fixtures", "shaped-text", "thai-signoff.json")
	if _, err := os.Stat(signoff); err == nil {
		t.Errorf("%s exists. Story 2.3a must not create it: it is D-2.3.5's outstanding gate obligation, and the record must name a real reader and a real date (D-000.28: a claim written before the event it asserts is false from birth)", signoff)
	}

	// (b) The set of matrix-tagged FILES is exactly the three that
	//     carried the tag at 431a6a5. Asserted as an inventory, by name,
	//     so that ADDING a matrix-tagged file — which is what "a third
	//     gate obligation" would look like in the tree — fails here.
	//     Removing one fails too.
	wantMatrixFiles := map[string]bool{
		"fontgen_matrix_test.go":        true,
		"matrix_test.go":                true,
		"shaped_signoff_matrix_test.go": true,
	}

	moduleRoot := filepath.Join(root, "folio-go")
	gotMatrixFiles := map[string]bool{}
	scanned := 0
	werr := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || (path != moduleRoot && strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		scanned++
		if hasMatrixBuildConstraint(string(src)) {
			rel, _ := filepath.Rel(moduleRoot, path)
			gotMatrixFiles[filepath.ToSlash(rel)] = true
		}
		return nil
	})
	if werr != nil {
		t.Fatalf("walk %s: %v", moduleRoot, werr)
	}
	if scanned == 0 {
		t.Fatal("vacuity guard: the walk read 0 Go files, so \"the matrix-tagged set is unchanged\" says nothing")
	}

	for name := range wantMatrixFiles {
		if !gotMatrixFiles[name] {
			t.Errorf("matrix-tagged file %s is no longer tagged — the Epic 2 gate's obligations must not shrink silently either", name)
		}
	}
	for name := range gotMatrixFiles {
		if !wantMatrixFiles[name] {
			t.Errorf(
				"%s is matrix-tagged and was not at 431a6a5. That is a THIRD Epic 2 gate obligation. "+
					"The gate owes exactly two things — the four-target matrix legs and D-2.3.5's Thai sign-off — "+
					"and a third may not be added without saying so explicitly.",
				name)
		}
	}
}

// hasMatrixBuildConstraint reports whether src carries an actual
// //go:build constraint naming the matrix tag — not merely a mention of
// one.
//
// The distinction is load-bearing and was found by this test failing on
// its own source. A substring search over the whole file reported SIX
// matrix-tagged files where there are three, because it also matched
// every file that DISCUSSES the tag in a comment — including this one.
// A guard that counts discussions of a constraint as instances of it
// would fire on any story that documented the gate, which is precisely
// the story that must not be blocked.
//
// So the constraint is recognised where the go tool recognises it: on
// its own line, in the build-constraint region, which ends at the
// package clause.
func hasMatrixBuildConstraint(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			return false
		}
		if !strings.HasPrefix(trimmed, "//go:build ") {
			continue
		}
		for _, tok := range strings.FieldsFunc(strings.TrimPrefix(trimmed, "//go:build "), func(r rune) bool {
			return r == ' ' || r == '(' || r == ')' || r == '!' || r == '&' || r == '|'
		}) {
			if tok == "matrix" {
				return true
			}
		}
	}
	return false
}

// repoRootForByteNeutrality walks up until it finds a directory holding
// both folio-go/ and lint/ — the same D-000.5/AD-21 pattern the other
// root-finders in this module use.
func repoRootForByteNeutrality(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		folioGo, e1 := os.Stat(filepath.Join(dir, "folio-go"))
		lintDir, e2 := os.Stat(filepath.Join(dir, "lint"))
		if e1 == nil && folioGo.IsDir() && e2 == nil && lintDir.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find the repo root (a directory containing both folio-go/ and lint/) walking up from %s", dir)
		}
		dir = parent
	}
}
