package folio_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	folio "github.com/panitw/folio/folio-go"
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

// declaredEpic2GateObligations IS THE LIST. It is the single, explicit
// declaration of everything the Epic 2 boundary gate owes, and the guard
// below asserts the OBSERVED obligation set equals it exactly.
//
// D-2.5.1, which replaced this guard's previous shape:
//
//	"TestStory23aAddedNoThirdEpic2GateObligation encodes a COUNT IN ITS
//	NAME, and that name rots on every future obligation. Replace it with
//	an assertion that the registered obligation set equals an explicit
//	declared list. Adding one then becomes a ONE-LINE DIFF to that list —
//	visible and reviewable — instead of a rename."
//
// The same derive-from-a-declarative-spec move that fixed the per-face
// assertion set in D-000.23's consequent obligation. It retires the
// counting problem PERMANENTLY rather than deferring it one story: the
// old name asserted "no THIRD", then had to mean "no FOURTH" while still
// reading "Third", which is a false statement compiled into the suite.
//
// Two kinds of obligation are observable in the tree, and both are
// listed here so neither can grow unnoticed:
//
//	matrix-file: <path>      a //go:build matrix file — a gate-run test.
//	                         A deliberately-red sign-off record is one of
//	                         these, and so is the matrix harness itself.
//	matrix-document: <slug>  a document registered in matrixDocuments
//	                         whose four legs the gate must run and
//	                         compare.
//
// Each entry names the story and the ruling that authorised it. Adding
// one without a ruling is exactly what this guard exists to stop.
var declaredEpic2GateObligations = []string{
	// The gate-run test files.
	"matrix-file: fontgen_matrix_test.go",                 // Story 2.2 — shipped-face instancing
	"matrix-file: matrix_test.go",                         // Story 1.2 — the four-target legs themselves
	"matrix-file: shaped_signoff_matrix_test.go",          // Story 2.3 — Thai READING sign-off (D-2.3.5)
	"matrix-file: expected_breaks_signoff_matrix_test.go", // Story 2.4 — Thai BREAK sign-off (D-2.4.3)

	// The documents whose four legs the gate runs and compares.
	"matrix-document: minimal-rect",          // Story 1.1
	"matrix-document: font-text",             // Story 1.5
	"matrix-document: image-embed",           // Story 1.8
	"matrix-document: multi-script-fallback", // Story 2.2
	"matrix-document: shaped-text",           // Story 2.3
	"matrix-document: wrapped-text",          // Story 2.4 — legs RUN in-story (D-000.4 override)
	"matrix-document: three-band-page",       // Story 2.5 — legs DEFERRED to the gate (D-2.5.1; D-000.4 override criterion DECLINED)
}

// TestEpic2GateObligationsMatchTheDeclaredSet asserts, mechanically
// rather than as an intention, that what the Epic 2 boundary gate owes
// is exactly what declaredEpic2GateObligations says it owes — no more,
// and no fewer.
//
// The two Thai sign-offs are distinct human judgments — "does this Thai
// READ correctly" and "do these BREAK POINTS fall correctly" — and each
// is tracked as its own deliberately-red gate-run test
// (TestShapedTextThaiSemanticSignOffIsRecorded,
// TestExpectedBreaksHumanSignOffIsRecorded). Each stays red until its
// own sign-off file names a reader, a date, what they examined, and the
// digest it certifies. No story creates either file — see (a).
//
// D-000.26 (refined) is why they are two records and not one: a sign-off
// binds to the artifact expressing the property judged. The reading
// judgment binds to fixtures/shaped-text/expected.pdf's digest; the
// break judgment binds to the break-opportunity VECTOR, which a uniform
// baseline shift would leave untouched.
func TestEpic2GateObligationsMatchTheDeclaredSet(t *testing.T) {
	root := repoRootForByteNeutrality(t)

	// (a) The sign-off obligations are still OUTSTANDING. If either file
	//     appears, a human judgment has been given — a real event that
	//     must be recorded deliberately by its owner, never as a side
	//     effect of another story (D-000.28: a claim written before the
	//     event it asserts is false from birth, and reads identically to
	//     a true one).
	signoff := filepath.Join(root, "fixtures", "shaped-text", "thai-signoff.json")
	if _, err := os.Stat(signoff); err == nil {
		t.Errorf("%s exists. No story may create it: it is D-2.3.5's outstanding gate obligation, and the record must name a real reader and a real date", signoff)
	}

	// (b) The OBSERVED obligation set, gathered from the tree itself.
	moduleRoot := filepath.Join(root, "folio-go")
	got := map[string]bool{}
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
			got["matrix-file: "+filepath.ToSlash(rel)] = true
		}
		return nil
	})
	if werr != nil {
		t.Fatalf("walk %s: %v", moduleRoot, werr)
	}
	if scanned == 0 {
		t.Fatal("vacuity guard: the walk read 0 Go files, so \"the obligation set is unchanged\" says nothing")
	}

	// The registered matrix documents. matrix_test.go is itself
	// matrix-tagged and therefore not compiled into this binary, so its
	// slugs are read from SOURCE — the same mechanism
	// matrix_registration_test.go uses (which separately pins these
	// slugs against .github/workflows/matrix.yml; this guard pins them
	// against the declared obligation list, a different question).
	//
	// Both guards call folio.MatrixDocumentSlugsFromSource and nothing
	// else. They previously shared a REGEX (`(?m)^\s*slug:\s*"…"`), which
	// made them one guard wearing two names: a single-line composite
	// literal is gofmt-clean, compiles under -tags matrix, and was
	// measured invisible to BOTH at once (Story 2.5 review, Finding 1).
	// The shared reader parses the Go grammar, where the literal has no
	// second spelling, and FAILS on any entry it cannot read rather than
	// dropping it from the set.
	slugs, elements, serr := folio.MatrixDocumentSlugsFromSource(filepath.Join(moduleRoot, "matrix_test.go"))
	if serr != nil {
		t.Fatalf("read matrixDocuments from matrix_test.go: %v", serr)
	}
	if len(slugs) == 0 || elements == 0 {
		t.Fatal("vacuity guard: no matrixDocuments entries found in matrix_test.go — the document half of the obligation set would be silently empty")
	}
	// N-of-N witness: every literal element yielded a slug. Asserted here
	// as well as inside the reader, so this guard does not take the
	// reader's word for its own completeness.
	if len(slugs) != elements {
		t.Fatalf("vacuity guard: read %d slugs from %d matrixDocuments entries — a registered document is missing from the observed obligation set", len(slugs), elements)
	}
	t.Logf("obligation witness — %d matrix documents read from %d matrixDocuments literal entries, %d Go files walked", len(slugs), elements, scanned)
	for _, slug := range slugs {
		got["matrix-document: "+slug] = true
	}

	// (c) Set equality against the declared list, in both directions.
	want := map[string]bool{}
	for _, o := range declaredEpic2GateObligations {
		want[o] = true
	}
	for o := range want {
		if !got[o] {
			t.Errorf("declared Epic 2 gate obligation %q is NOT present in the tree — obligations must not disappear silently either. Either restore it, or remove its line from declaredEpic2GateObligations with the ruling that discharged it.", o)
		}
	}
	for o := range got {
		if !want[o] {
			t.Errorf(
				"%q is an Epic 2 gate obligation and is NOT in declaredEpic2GateObligations.\n\n"+
					"An obligation may not be added without a ruling that says so explicitly. Record it as a "+
					"ONE-LINE addition to that list, naming the story and the decision that authorised it "+
					"(D-2.5.1). Do not rename this test, and do not encode a count anywhere.",
				o)
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
