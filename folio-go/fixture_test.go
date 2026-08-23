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

	// Render and validate structurally first, unconditionally — see the
	// ordering note above (Major 5).
	b, err := Render(nil)
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
