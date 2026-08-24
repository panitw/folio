package folio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/pdf"
)

// expectedFixture mirrors fixtures/minimal-rect/expected.json.
type expectedFixture struct {
	FolioGoVersion string `json:"folioGoVersion"`
	GoToolchain    string `json:"goToolchain"`
	SHA256         string `json:"sha256"`
}

// repoRootFromTest finds the folio repository root by walking up from the
// current working directory until it finds a directory containing both
// folio-go/ and fixtures/. Duplicated (rather than shared) with
// internal/pdf's copy because Go test helpers do not cross package
// boundaries and internal/pdf is not importable from here in reverse; both
// copies exist only in _test.go files, which are exempt from AD-1's os
// ban (D-000.5 §Consequences).
func repoRootFromTest(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if isRepoRootDir(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root (a directory containing both folio-go/ and fixtures/) walking up from %s", dir)
		}
		dir = parent
	}
}

// isSHA256HexString reports whether s is exactly 64 lower-case hex
// characters — the JSON-string shape AC16 requires of expected.json's
// sha256 field.
func isSHA256HexString(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func isRepoRootDir(dir string) bool {
	folioGo, err1 := os.Stat(filepath.Join(dir, "folio-go"))
	fixtures, err2 := os.Stat(filepath.Join(dir, "fixtures"))
	return err1 == nil && folioGo.IsDir() && err2 == nil && fixtures.IsDir()
}

// checkFixtureShape is the AC1/AC6 pure checker underlying DW-1: it reads
// and validates the fixture at the given path — parses as JSON, carries
// both version fields, and has a sha256 field that is a JSON string of
// exactly 64 lower-case hex characters (AC16), never a per-target map,
// array or object (D-1.2.2). It takes the fixture's location as a
// parameter rather than a hard-coded relative path (AC6: the seam
// applied generally here, not patched into this one call site as a
// one-off), so it can be red-proved against a scratch copy without ever
// touching the real golden fixture (RP-4, DW-1).
//
// Blocker 3 (this story's QA review): this function, loadExpectedFixture
// and TestFixtureShapeCheckRedProof used to live in matrix_test.go,
// which carries the "matrix" build tag — so DW-1's entire red-proof
// executed in zero gates (the tagged suite is deferred to the Epic 1
// boundary, D-000.4; CI's only `-tags=matrix` steps are `go build` and
// `go vet`, neither of which runs a test). They live here instead,
// untagged, in package folio's ordinary `go test ./...` path — the same
// home isSHA256HexString already has — so DW-1's closure actually runs
// on every story. matrix_test.go's TestCrossTargetByteIdentity and
// TestTargetRenderHash still call loadExpectedFixture below; being in a
// different file of the same package changes nothing about that call.
func checkFixtureShape(path string) (expectedFixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return expectedFixture{}, fmt.Errorf("read fixture: %w", err)
	}
	var f expectedFixture
	if err := json.Unmarshal(data, &f); err != nil {
		return expectedFixture{}, fmt.Errorf("parse fixture JSON: %w", err)
	}
	if f.GoToolchain == "" {
		return expectedFixture{}, fmt.Errorf("fixture is missing goToolchain")
	}
	if f.SHA256 == "" {
		return expectedFixture{}, fmt.Errorf("fixture is missing sha256")
	}
	if !isSHA256HexString(f.SHA256) {
		return expectedFixture{}, fmt.Errorf("fixture sha256 %q is not a JSON string of exactly 64 lower-case hex characters (AC16)", f.SHA256)
	}
	return f, nil
}

// loadExpectedFixture reads fixtures/minimal-rect/expected.json — the
// same recorded golden Story 1.1's TestRenderMatchesGoldenFixture reads
// (AC7, vacuity guard 3: no second golden) — via checkFixtureShape, and
// fails the test on any shape violation.
func loadExpectedFixture(t *testing.T, path string) expectedFixture {
	t.Helper()
	f, err := checkFixtureShape(path)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// TestFixtureShapeCheckRedProof is DW-1's closure (AC6, RP-4): it
// red-proves AC16's shape check against scratch copies of
// fixtures/minimal-rect/expected.json, and confirms the real fixture
// file was never touched. Story 1.2's finisher deferred exactly this
// (DW-1) because the check previously had no way to be shown going red
// without mutating the real golden fixture.
//
// Finding 12 (this story's QA review): the fixture inventory promises
// two scratch shapes — "a scratch expected.json whose sha256 is a
// per-target object, and one whose value is upper-case or 63
// characters" — but only the first existed, so the isSHA256HexString
// branch of checkFixtureShape (as opposed to the earlier JSON-unmarshal
// failure) had no red-proof at all. Both now run as subtests.
func TestFixtureShapeCheckRedProof(t *testing.T) {
	root := repoRootFromTest(t)
	realPath := filepath.Join(root, "fixtures", "minimal-rect", "expected.json")

	before, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatalf("read real fixture: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(before, &raw); err != nil {
		t.Fatalf("parse real fixture: %v", err)
	}

	writeScratch := func(t *testing.T, mutated map[string]any) string {
		t.Helper()
		widened, err := json.MarshalIndent(mutated, "", "  ")
		if err != nil {
			t.Fatalf("marshal scratch fixture: %v", err)
		}
		scratchPath := filepath.Join(t.TempDir(), "expected.json")
		if err := os.WriteFile(scratchPath, widened, 0o644); err != nil {
			t.Fatalf("write scratch fixture: %v", err)
		}
		return scratchPath
	}

	t.Run("sha256 widened into a per-target object", func(t *testing.T) {
		mutated := map[string]any{}
		for k, v := range raw {
			mutated[k] = v
		}
		// The exact D-1.2.2 hazard AC16 exists to block.
		mutated["sha256"] = map[string]string{
			"darwin-arm64": "0000000000000000000000000000000000000000000000000000000000000000",
		}
		if _, err := checkFixtureShape(writeScratch(t, mutated)); err == nil {
			t.Fatal("expected checkFixtureShape to fail on a widened per-target sha256 object, got nil error")
		}
	})

	t.Run("sha256 wrong length or case", func(t *testing.T) {
		cases := []struct {
			name string
			sha  string
		}{
			{"63 characters", strings.Repeat("a", 63)},
			{"upper-case", strings.ToUpper(strings.Repeat("a", 64))},
		}
		for _, c := range cases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				mutated := map[string]any{}
				for k, v := range raw {
					mutated[k] = v
				}
				mutated["sha256"] = c.sha
				if _, err := checkFixtureShape(writeScratch(t, mutated)); err == nil {
					t.Fatalf("expected checkFixtureShape to fail on sha256 = %q, got nil error", c.sha)
				}
			})
		}
	})

	after, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatalf("re-read real fixture: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("real fixtures/minimal-rect/expected.json was mutated by this red-proof")
	}
}

// TestRenderMatchesGoldenFixture is AC11: the live render's hash must
// match the recorded golden fixture, and the fixture must carry both
// version fields (vacuity guard: a fixture missing either field fails
// before any hash comparison runs — this is RP-8's target).
//
// Ordering matters here, and this story's QA review (Major 5) measured
// why: the toolchain-mismatch check used to run, and t.Fatalf, before
// Render() was ever called. On any machine not running the exact
// recorded toolchain — every contributor who is not the recording
// machine, and every leg Story 1.2 is about to add — that meant *nothing
// in the module observed the output bytes at all*, because the golden
// hash was the only assertion that did. Render() now runs unconditionally
// and its output is checked structurally (assertWellFormedPDF, which this
// story's review also strengthened — Major 4/18) before the
// toolchain-gated hash comparison, so a machine on a different toolchain
// still observes the document's self-consistency instead of nothing.
func TestRenderMatchesGoldenFixture(t *testing.T) {
	root := repoRootFromTest(t)
	fixturePath := filepath.Join(root, "fixtures", "minimal-rect", "expected.json")

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture expectedFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture JSON: %v", err)
	}

	// Vacuity guard: the fixture must carry both version fields before
	// any hash comparison runs, and the hash field must be present too.
	// This does NOT require fixture.FolioGoVersion == Version: AC11
	// requires the fixture to *record* the version, not to match the
	// module's current one, and asserting equality would fail the golden
	// test on the very first legitimate version bump even though the
	// rendered bytes are unchanged — training the "just regenerate it"
	// reflex vacuity guard 4 exists to suppress, on a false positive
	// (this story's QA review, Major 11).
	if fixture.FolioGoVersion == "" {
		t.Fatal("fixture is missing folioGoVersion")
	}
	if fixture.GoToolchain == "" {
		t.Fatal("fixture is missing goToolchain (RP-8: this must fail before the hash comparison runs)")
	}
	if fixture.SHA256 == "" {
		t.Fatal("fixture is missing sha256")
	}
	// AC16: sha256 must be a JSON string of exactly 64 lower-case hex
	// characters — never a map, array, or per-target object. A developer
	// meeting a red matrix arm (Story 1.2) who starts "fixing" it by
	// converting this field into a per-target map must hit a red here
	// before they can reach a green (D-1.2.2's escalation rule: a real
	// cross-target divergence is an owner decision, never a developer's
	// to paper over by recording what each target produced).
	if !isSHA256HexString(fixture.SHA256) {
		t.Fatalf("fixture sha256 %q is not a JSON string of exactly 64 lower-case hex characters "+
			"(AC16: never a per-target map, array, or object — see D-1.2.2)", fixture.SHA256)
	}

	// Serialize and validate structurally first, unconditionally — see
	// the ordering note above (Major 5).
	//
	// AC14a, D-1.5.9: this fixture is RETAINED UNCHANGED and
	// RECLASSIFIED as pinning internal/pdf's fontless emission path —
	// AD-3 number emission, xref arithmetic, content-derived /ID — not
	// the public Render API, which D-1.1.c designed to change every
	// story and which now requires a non-nil template (AC14b). No
	// `.folio` template can reproduce these exact 547 bytes (F-8: the
	// golden's content stream has no colour operator, and any
	// template-authored rectangle's style.background emits one via
	// D-1.1.b's colour rule) — so this test calls internal/pdf.Serialize()
	// directly, exactly as it did before Render accepted a template at
	// all.
	b := pdf.Serialize()
	assertWellFormedPDF(t, "golden fixture render", b)

	// expected.pdf is kept for human diffing, but nothing previously read
	// it, so it could silently drift from the normative hash in
	// expected.json without any test noticing (this story's QA review,
	// Minor 19).
	expectedPDFPath := filepath.Join(root, "fixtures", "minimal-rect", "expected.pdf")
	expectedPDFBytes, err := os.ReadFile(expectedPDFPath)
	if err != nil {
		t.Fatalf("read expected.pdf: %v", err)
	}
	expectedPDFSum := sha256.Sum256(expectedPDFBytes)
	expectedPDFHex := hex.EncodeToString(expectedPDFSum[:])
	if expectedPDFHex != fixture.SHA256 {
		t.Fatalf(
			"fixtures/minimal-rect/expected.pdf's own sha256 (%s) does not match "+
				"expected.json's recorded sha256 (%s) — the fixture's two halves have drifted apart",
			expectedPDFHex, fixture.SHA256,
		)
	}

	// A toolchain bump is a versioned breaking change under AD-22 (C6):
	// fail loudly and instructively rather than silently comparing
	// against a hash measured under a different compiler. This gate now
	// runs after the structural checks above, not before Render() is
	// even called (Major 5).
	if runtime.Version() != fixture.GoToolchain {
		t.Fatalf(
			"toolchain mismatch: running under %s, fixture was recorded under %s. "+
				"A Go toolchain bump is a versioned breaking change under AD-22 — "+
				"the golden hash must be re-measured deliberately, not regenerated "+
				"automatically, and the change must be recorded as an intended, "+
				"versioned event (C6).",
			runtime.Version(), fixture.GoToolchain,
		)
	}

	sum := sha256.Sum256(b)
	gotHex := hex.EncodeToString(sum[:])

	if gotHex != fixture.SHA256 {
		t.Fatalf(
			"golden fixture mismatch: got sha256 %s, want %s (fixtures/minimal-rect). "+
				"Under AD-21/AD-22 this is a defect until proven to be an intended, "+
				"versioned change — see fixtures/minimal-rect/README.md. Do not "+
				"regenerate the fixture to make this pass.",
			gotHex, fixture.SHA256,
		)
	}
}

// TestRenderMatchesFontTextGoldenFixture is AC14: fixtures/font-text/,
// the NEW golden fixture beside fixtures/minimal-rect/ (AC14a: that one
// is retained unchanged, never replaced). Same shape and same rules as
// TestRenderMatchesGoldenFixture above, applied to the font-embedding
// path through the public Render API instead of internal/pdf.Serialize()
// directly.
func TestRenderMatchesFontTextGoldenFixture(t *testing.T) {
	root := repoRootFromTest(t)
	fixturePath := filepath.Join(root, "fixtures", "font-text", "expected.json")

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture expectedFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture JSON: %v", err)
	}

	if fixture.FolioGoVersion == "" {
		t.Fatal("fixture is missing folioGoVersion")
	}
	if fixture.GoToolchain == "" {
		t.Fatal("fixture is missing goToolchain (RP-8: this must fail before the hash comparison runs)")
	}
	if fixture.SHA256 == "" {
		t.Fatal("fixture is missing sha256")
	}
	if !isSHA256HexString(fixture.SHA256) {
		t.Fatalf("fixture sha256 %q is not a JSON string of exactly 64 lower-case hex characters (AC16)", fixture.SHA256)
	}

	// Finding 14 (QA review): fixtures/font-text/input.folio's own README
	// stated it is "kept in sync by hand — the fixture is not read at
	// test time", with nothing asserting that claim. AD-21 makes the
	// fixture the normative record of what produced the hash; if
	// fontTestTemplateJSON were edited and input.folio were not, the
	// fixture would document a document that does not produce its own
	// recorded hash — a silently lying artifact. This costs nothing new:
	// the test already reads fixtures/ (out-of-module) under -count=1.
	inputFolioPath := filepath.Join(root, "fixtures", "font-text", "input.folio")
	inputFolioBytes, err := os.ReadFile(inputFolioPath)
	if err != nil {
		t.Fatalf("read %s: %v", inputFolioPath, err)
	}
	if string(inputFolioBytes) != fontTestTemplateJSON {
		t.Fatalf(
			"%s has drifted from folio-go/fontTestTemplateJSON (render_test.go) — the two are "+
				"supposed to be byte-identical (kept in sync by hand, per the fixture's own "+
				"README); update input.folio to match, or this fixture no longer documents what "+
				"actually produced its recorded hash",
			inputFolioPath,
		)
	}

	tpl := parseFontTestTemplate(t)
	b, err := Render(tpl, Data("{}"), nil, testFontSet())
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	assertWellFormedPDF(t, "font-text golden fixture render", b)

	// AC10b's vacuity guard, applied here too: confirm the render
	// actually contains a FontFile2 before comparing any hash — a
	// fontless render matching a fontless golden would prove nothing
	// about font byte-stability (D-000.9).
	if !containsFontFile2(b) {
		t.Fatal("font-text golden fixture render does not contain a FontFile2 — the fixture would certify nothing (AC10b)")
	}

	// Story 2.2 (AC6/AC6a, D-2.2.2 (superseded)): the six-letter subset
	// tag re-derivation moved this fixture's whole-document hash (the
	// tag changed at its three sites, and the content-derived /ID moved
	// as a consequence — see fixtures/font-text/README.md's "Re-recorded
	// at Story 2.2" section for the measured, tag-only delta). AC8's
	// original "bytes unchanged" clause is corrected by this same
	// commit into a narrower, still-meaningful one: the EMBEDDED FONT
	// PROGRAM itself — independent of its tag, and independent of the
	// PDF's other content-derived bytes — must still be byte-identical.
	// Pinning this constant is what makes a FUTURE change that moves
	// the program (not just the tag) distinguishable from one that
	// doesn't, which the whole-document hash alone cannot tell apart.
	const wantFontTextProgramSHA256 = "2ef52cca3f6bdb76de8cf5a4c73ca3f7e0f9154bb7bb0b1122425127419db4de"
	programSum := sha256.Sum256(extractFontFile2Program(t, b))
	if gotProgram := hex.EncodeToString(programSum[:]); gotProgram != wantFontTextProgramSHA256 {
		t.Fatalf(
			"font-text's embedded FontFile2 PROGRAM bytes changed: got sha256 %s, want %s — "+
				"this is the assertion that replaced AC8's original whole-fixture \"bytes unchanged\" "+
				"clause (Story 2.2): the program itself must stay stable even though the fixture's "+
				"whole-document hash was deliberately re-recorded for the tag-derivation change",
			gotProgram, wantFontTextProgramSHA256,
		)
	}

	expectedPDFPath := filepath.Join(root, "fixtures", "font-text", "expected.pdf")
	expectedPDFBytes, err := os.ReadFile(expectedPDFPath)
	if err != nil {
		t.Fatalf("read expected.pdf: %v", err)
	}
	expectedPDFSum := sha256.Sum256(expectedPDFBytes)
	expectedPDFHex := hex.EncodeToString(expectedPDFSum[:])
	if expectedPDFHex != fixture.SHA256 {
		t.Fatalf(
			"fixtures/font-text/expected.pdf's own sha256 (%s) does not match "+
				"expected.json's recorded sha256 (%s) — the fixture's two halves have drifted apart",
			expectedPDFHex, fixture.SHA256,
		)
	}

	if runtime.Version() != fixture.GoToolchain {
		t.Fatalf(
			"toolchain mismatch: running under %s, fixture was recorded under %s. "+
				"A Go toolchain bump is a versioned breaking change under AD-22 — "+
				"the golden hash must be re-measured deliberately (C6).",
			runtime.Version(), fixture.GoToolchain,
		)
	}

	sum := sha256.Sum256(b)
	gotHex := hex.EncodeToString(sum[:])
	if gotHex != fixture.SHA256 {
		t.Fatalf(
			"golden fixture mismatch: got sha256 %s, want %s (fixtures/font-text). "+
				"Under AD-21/AD-22 this is a defect until proven to be an intended, "+
				"versioned change — see fixtures/font-text/README.md.",
			gotHex, fixture.SHA256,
		)
	}
}

// TestRenderMatchesImageEmbedGoldenFixture is AC22/AC25 (D-1.8.6): the
// THIRD matrix document, joining minimal-rect and font-text — same
// shape as TestRenderMatchesFontTextGoldenFixture above, applied to the
// image-embedding path.
func TestRenderMatchesImageEmbedGoldenFixture(t *testing.T) {
	root := repoRootFromTest(t)
	fixturePath := filepath.Join(root, "fixtures", "image-embed", "expected.json")

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture expectedFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture JSON: %v", err)
	}

	if fixture.FolioGoVersion == "" {
		t.Fatal("fixture is missing folioGoVersion")
	}
	if fixture.GoToolchain == "" {
		t.Fatal("fixture is missing goToolchain (RP-8: this must fail before the hash comparison runs)")
	}
	if fixture.SHA256 == "" {
		t.Fatal("fixture is missing sha256")
	}
	if !isSHA256HexString(fixture.SHA256) {
		t.Fatalf("fixture sha256 %q is not a JSON string of exactly 64 lower-case hex characters (AC16)", fixture.SHA256)
	}

	// AC25a: input.folio is byte-identical to the Go constant that
	// renders it — the same obligation font-text/input.folio carries
	// (Finding 14, Story 1.5's QA review), inherited here.
	inputFolioPath := filepath.Join(root, "fixtures", "image-embed", "input.folio")
	inputFolioBytes, err := os.ReadFile(inputFolioPath)
	if err != nil {
		t.Fatalf("read %s: %v", inputFolioPath, err)
	}
	if string(inputFolioBytes) != imageTestTemplateJSON {
		t.Fatalf(
			"%s has drifted from folio-go/imageTestTemplateJSON (render_test.go) — the two are "+
				"supposed to be byte-identical; update input.folio to match, or this fixture no "+
				"longer documents what actually produced its recorded hash",
			inputFolioPath,
		)
	}

	tpl, err := ParseTemplate([]byte(imageTestTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	b, err := Render(tpl, Data("{}"), nil, nil)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	assertWellFormedPDF(t, "image-embed golden fixture render", b)

	// AC23's vacuity guard, applied here too: confirm the render
	// actually contains an image XObject before comparing any hash — a
	// render that silently dropped the image would certify nothing
	// (D-000.9).
	if !containsImageXObject(b) {
		t.Fatal("image-embed golden fixture render does not contain an image XObject — the fixture would certify nothing (AC23)")
	}

	expectedPDFPath := filepath.Join(root, "fixtures", "image-embed", "expected.pdf")
	expectedPDFBytes, err := os.ReadFile(expectedPDFPath)
	if err != nil {
		t.Fatalf("read expected.pdf: %v", err)
	}
	expectedPDFSum := sha256.Sum256(expectedPDFBytes)
	expectedPDFHex := hex.EncodeToString(expectedPDFSum[:])
	if expectedPDFHex != fixture.SHA256 {
		t.Fatalf(
			"fixtures/image-embed/expected.pdf's own sha256 (%s) does not match "+
				"expected.json's recorded sha256 (%s) — the fixture's two halves have drifted apart",
			expectedPDFHex, fixture.SHA256,
		)
	}

	if runtime.Version() != fixture.GoToolchain {
		t.Fatalf(
			"toolchain mismatch: running under %s, fixture was recorded under %s. "+
				"A Go toolchain bump is a versioned breaking change under AD-22 — "+
				"the golden hash must be re-measured deliberately (C6).",
			runtime.Version(), fixture.GoToolchain,
		)
	}

	sum := sha256.Sum256(b)
	gotHex := hex.EncodeToString(sum[:])
	if gotHex != fixture.SHA256 {
		t.Fatalf(
			"golden fixture mismatch: got sha256 %s, want %s (fixtures/image-embed). "+
				"Under AD-21/AD-22 this is a defect until proven to be an intended, "+
				"versioned change — see fixtures/image-embed/README.md.",
			gotHex, fixture.SHA256,
		)
	}
}

// TestMultiScriptFallbackGoldenFixture is Story 2.2's AC8 fourth
// fixture: renders multiScriptTestTemplateJSON (fixtures/multi-script-
// fallback/input.folio) through the public Render path against the
// REAL shipped face set, and compares against the recorded golden —
// the same shape TestRenderMatchesFontTextGoldenFixture already
// established for font-text, extended with AC7's per-(face,pinned-
// instance) program digests (V7): the whole-document hash proves the
// DOCUMENT is stable; the three per-face digests below prove each
// enumerated PAIR individually is, which is what "a golden for EVERY
// (face, pinned instance) pair — not one sample" asks for.
func TestMultiScriptFallbackGoldenFixture(t *testing.T) {
	root := repoRootFromTest(t)
	fixturePath := filepath.Join(root, "fixtures", "multi-script-fallback", "expected.json")

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture expectedFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture JSON: %v", err)
	}
	if fixture.FolioGoVersion == "" {
		t.Fatal("fixture is missing folioGoVersion")
	}
	if fixture.GoToolchain == "" {
		t.Fatal("fixture is missing goToolchain (RP-8: this must fail before the hash comparison runs)")
	}
	if fixture.SHA256 == "" {
		t.Fatal("fixture is missing sha256")
	}
	if !isSHA256HexString(fixture.SHA256) {
		t.Fatalf("fixture sha256 %q is not a JSON string of exactly 64 lower-case hex characters (AC16)", fixture.SHA256)
	}

	inputFolioPath := filepath.Join(root, "fixtures", "multi-script-fallback", "input.folio")
	inputFolioBytes, err := os.ReadFile(inputFolioPath)
	if err != nil {
		t.Fatalf("read %s: %v", inputFolioPath, err)
	}
	if string(inputFolioBytes) != multiScriptTestTemplateJSON {
		t.Fatalf(
			"%s has drifted from folio-go/multiScriptTestTemplateJSON (render_test.go) — the two are "+
				"supposed to be byte-identical (kept in sync by hand, per font-text's precedent)",
			inputFolioPath,
		)
	}

	tpl, err := ParseTemplate([]byte(multiScriptTestTemplateJSON))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	b, err := Render(tpl, Data("{}"), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	assertWellFormedPDF(t, "multi-script-fallback golden fixture render", b)

	if !containsFontFile2(b) {
		t.Fatal("multi-script-fallback golden fixture render does not contain a FontFile2 — the fixture would certify nothing (AC10b)")
	}
	programs := extractAllFontFile2Programs(t, b)
	if len(programs) != len(shippedFaceSpecs) {
		t.Fatalf(
			"expected exactly %d embedded FontFile2 programs (one per shipped face actually used), got %d",
			len(shippedFaceSpecs), len(programs),
		)
	}

	// D-000.22's semantic acceptance step, on the EMBEDDED programs:
	// static, glyf, and Regular. This is the assertion whose absence let
	// Story 2.2 record a golden with Simplified Chinese at
	// usWeightClass=100 while four-target byte-identity, the program
	// digest and the missing-glyph diagnostic all reported success.
	assertEmbeddedProgramsAreStaticRegular(t, "multi-script-fallback golden", programs, 400)

	// ...and on the produced PDF: /BaseFont must carry the embedded
	// program's own PostScript name behind exactly one six-letter tag
	// (ISO 32000-1 Table 117), not the FontSet key.
	wantPS := map[string]bool{}
	for _, spec := range shippedFaceSpecs {
		wantPS[spec.PostScriptName] = true
	}
	assertBaseFontNames(t, "multi-script-fallback golden", b, wantPS)

	// V5a's exclusivity half: every occurrence of each subset tag in
	// the whole PDF must sit in a name position.
	assertSubsetTagsAppearOnlyInNames(t, "multi-script-fallback golden", b)

	// AC7's per-(face, instance) goldens (V7): one per shipped face.
	// Object-number order in the rendered PDF is Noto Sans, Noto Sans
	// SC, Noto Sans Thai (measured, matches alphabetical face-name
	// sort — render.go's faceNames := slices.Sorted(maps.Keys(...))).
	//
	// RE-RECORDED by Story 2.2's finisher, and this is a versioned
	// change under AD-22, not a drift. The shipped faces are no longer
	// the upstream VARIABLE builds instanced at render time to their axis
	// defaults; they are static Regular instances derived ahead of the
	// build (D-2.2.4). The old Noto Sans SC program here recorded
	// OS/2.usWeightClass = 100 — Thin — because that face's `wght` axis
	// defaults to 100 and "pin every axis to its default" faithfully
	// produced it. Every guard agreed the artifact was correct; none was
	// asked whether the value meant what its name implied (D-000.21).
	// assertEmbeddedProgramsAreStaticRegular above is the assertion that
	// was missing, and it now runs before these hashes are compared.
	wantProgramSHA256 := []string{
		"51620d3ae7f511c3b537439f40737d277907c44b0b2cc9d643529daffcd9ccd6", // Noto Sans
		"4a2c1e286e0628124e90931afd0b017f16eea3fd38b5c9ff978d3d82315a495d", // Noto Sans SC
		"beca96084cf1277a6b234727a373c66f8fbc2671a189cf6cea387e7b17702a28", // Noto Sans Thai
	}
	faceLabels := []string{"Noto Sans", "Noto Sans SC", "Noto Sans Thai"}

	// V7's guard, REBUILT. It previously read
	//
	//     if len(wantProgramSHA256) != 3 { ... }
	//
	// which compares a slice literal's length to a hard-coded 3
	// declared six lines above it. That cannot detect the hazard V7 is
	// named for, and was red-proved open: adding a fourth face to
	// fonts.Shipped() left the whole root package green, with four
	// shipped (face, instance) pairs against three goldens.
	//
	// The count is now asserted against the ENUMERATION rather than
	// narrated beside it (D-000.14). testShippedFontSet() is chained to
	// fonts.Shipped() by TestShippedSpecCoversEverythingShipped and
	// TestFontsShippedMatchesExpectedFaceSet, so a face added to the
	// shipped set and nowhere else now fails HERE.
	if len(wantProgramSHA256) != len(shippedFaceSpecs) {
		t.Fatalf(
			"AC7/V7: %d per-instance goldens are recorded, but %d (face, pinned instance) pairs ship. "+
				"The matrix would be monitoring a subset of the risk surface and reporting green over the "+
				"uncovered ones — D-2.2.1's exact hazard.",
			len(wantProgramSHA256), len(shippedFaceSpecs),
		)
	}
	if len(faceLabels) != len(wantProgramSHA256) {
		t.Fatalf("test bug: %d face labels for %d goldens", len(faceLabels), len(wantProgramSHA256))
	}
	for i, prog := range programs {
		sum := sha256.Sum256(prog)
		got := hex.EncodeToString(sum[:])
		if got != wantProgramSHA256[i] {
			t.Fatalf(
				"%s's embedded, instanced program changed: got sha256 %s, want %s (AC7 per-instance golden). "+
					"Under AD-21/AD-22 this is a defect until proven to be an intended, versioned change.",
				faceLabels[i], got, wantProgramSHA256[i],
			)
		}
	}

	expectedPDFPath := filepath.Join(root, "fixtures", "multi-script-fallback", "expected.pdf")
	if expectedPDFBytes, err := os.ReadFile(expectedPDFPath); err == nil {
		expectedPDFSum := sha256.Sum256(expectedPDFBytes)
		expectedPDFHex := hex.EncodeToString(expectedPDFSum[:])
		if expectedPDFHex != fixture.SHA256 {
			t.Fatalf(
				"fixtures/multi-script-fallback/expected.pdf's own sha256 (%s) does not match "+
					"expected.json's recorded sha256 (%s) — the fixture's two halves have drifted apart",
				expectedPDFHex, fixture.SHA256,
			)
		}
	}

	if runtime.Version() != fixture.GoToolchain {
		t.Fatalf(
			"toolchain mismatch: running under %s, fixture was recorded under %s. "+
				"A Go toolchain bump is a versioned breaking change under AD-22 — "+
				"the golden hash must be re-measured deliberately (C6).",
			runtime.Version(), fixture.GoToolchain,
		)
	}

	sum := sha256.Sum256(b)
	gotHex := hex.EncodeToString(sum[:])
	if gotHex != fixture.SHA256 {
		t.Fatalf(
			"golden fixture mismatch: got sha256 %s, want %s (fixtures/multi-script-fallback). "+
				"Under AD-21/AD-22 this is a defect until proven to be an intended, "+
				"versioned change — see fixtures/multi-script-fallback/README.md.",
			gotHex, fixture.SHA256,
		)
	}
}
