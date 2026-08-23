package folio

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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

func isRepoRootDir(dir string) bool {
	folioGo, err1 := os.Stat(filepath.Join(dir, "folio-go"))
	fixtures, err2 := os.Stat(filepath.Join(dir, "fixtures"))
	return err1 == nil && folioGo.IsDir() && err2 == nil && fixtures.IsDir()
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

	// Render and validate structurally first, unconditionally — see the
	// ordering note above (Major 5).
	b, err := Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
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
