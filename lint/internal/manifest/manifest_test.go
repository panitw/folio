package manifest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repoRootFromTest mirrors rules.repoRootFromTest (duplicated: this is a
// different package, and it lives only in a _test.go file).
func repoRootFromTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		folioGo, err1 := os.Stat(filepath.Join(dir, "folio-go"))
		lintDir, err2 := os.Stat(filepath.Join(dir, "lint"))
		if err1 == nil && folioGo.IsDir() && err2 == nil && lintDir.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root walking up from %s", dir)
		}
		dir = parent
	}
}

// TestManifestUpToDate is AC19's completeness assertion (RP-9): the
// committed lint/MANIFEST.md must match what Generate/Render produce
// right now. A dependency added to any of the three module graphs
// without regenerating and committing the manifest fails this test —
// `cd lint && go run ./cmd/genmanifest` (lint has its own go.mod — run
// from inside lint/, not the repo root) regenerates it (Finding 16, QA
// review: the previously-documented command failed with "cannot find
// main module").
func TestManifestUpToDate(t *testing.T) {
	root := repoRootFromTest(t)

	rows, err := Generate(root)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assetRows, err := ResolveAssets(root)
	if err != nil {
		t.Fatalf("ResolveAssets: %v", err)
	}
	want := Render(rows) + RenderAssets(assetRows)

	committedPath := filepath.Join(root, filepath.FromSlash(CommittedRelPath))
	got, err := os.ReadFile(committedPath)
	if err != nil {
		t.Fatalf("read committed manifest %s: %v (run `cd lint && go run ./cmd/genmanifest`)", committedPath, err)
	}

	if string(got) != want {
		t.Fatalf("lint/MANIFEST.md is out of date — run `cd lint && go run ./cmd/genmanifest` and commit the result")
	}
}

// TestResolveAssetsIncludesWordlist is Story 2.1's addition (AC9): the
// CC0 wordlist at folio-go/internal/text/wordlist/ must appear in
// ResolveAssets' output — a font-extension-only walk would never see
// it (that gap is exactly what motivated AC9's separate guard); this
// asserts the manifest generator ALSO accounts for it, not just the
// lint guard.
func TestResolveAssetsIncludesWordlist(t *testing.T) {
	root := repoRootFromTest(t)
	rows, err := ResolveAssets(root)
	if err != nil {
		t.Fatalf("ResolveAssets: %v", err)
	}
	const wantPath = "folio-go/internal/text/wordlist/words_th.txt"
	var found *AssetRow
	for i := range rows {
		if rows[i].Path == wantPath {
			found = &rows[i]
		}
	}
	if found == nil {
		t.Fatalf("expected an AssetRow for %s, got %d rows: %v", wantPath, len(rows), rows)
	}
	if found.Licence != "CC0-1.0" {
		t.Errorf("wordlist AssetRow.Licence = %q, want %q", found.Licence, "CC0-1.0")
	}
	if found.Copyright == "" {
		t.Error("wordlist AssetRow.Copyright is empty")
	}
}

// initGitRepo creates a fresh, empty git repository at dir — DW-19's
// fix consults git's own index, so a synthetic test subject needs a
// REAL repository, not a bare directory tree.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func gitAdd(t *testing.T, dir string, paths ...string) {
	t.Helper()
	args := append([]string{"add"}, paths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add %v: %v\n%s", paths, err, out)
	}
}

// TestResolveAssetsExcludesUntrackedDirectoryWithoutError is DW-19's
// own regression guard (Story 3.6, AC10, D-3.6.5): a directory holding
// a real font file on disk but ZERO files git tracks is a local,
// untracked scratch folder — excluded from ResolveAssets' output
// entirely (no row, no error), never mistaken for a licensing
// violation. Anchor: git's own index (a fact this test's synthetic
// repository controls directly).
//
// A second, TRACKED and fully compliant font directory sits alongside
// the untracked one — the real repository's own shape (`.font-sources`
// untracked, `folio-go/fonts/*` tracked) and, since the D-3.6.5
// amendment (Finding 1, QA review, Blocker), a REQUIRED precondition:
// "candidates exist and every one is untracked" is now its own scan
// error (see TestResolveAssetsAllDirectoriesUntrackedIsAScanError),
// so a test proving the untracked-exclusion behaviour in isolation
// must not itself BE that all-untracked case.
func TestResolveAssetsExcludesUntrackedDirectoryWithoutError(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	untrackedDir := filepath.Join(root, "scratch-fonts")
	if err := os.MkdirAll(untrackedDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// A real font-extensioned file, present on disk but NEVER git-added
	// — the exact shape .font-sources/ has on a developer's machine.
	if err := os.WriteFile(filepath.Join(untrackedDir, "Untracked.ttf"), []byte("not a real font, just bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	compliantDir := filepath.Join(root, "tracked-fonts")
	if err := os.MkdirAll(compliantDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	fontPath := filepath.Join(compliantDir, "Tracked.ttf")
	if err := os.WriteFile(fontPath, []byte("not a real font, just bytes"), 0o644); err != nil {
		t.Fatalf("write font: %v", err)
	}
	licensePath := filepath.Join(compliantDir, "LICENSE.txt")
	if err := os.WriteFile(licensePath, []byte("a licence"), 0o644); err != nil {
		t.Fatalf("write licence: %v", err)
	}
	noticePath := filepath.Join(compliantDir, "NOTICE")
	if err := os.WriteFile(noticePath, []byte("Copyright 2026 Test\n"), 0o644); err != nil {
		t.Fatalf("write notice: %v", err)
	}
	gitAdd(t, root, fontPath, licensePath, noticePath)

	// A trivial tracked file so the repository has at least one commit
	// worth of history to query against.
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	gitAdd(t, root, "README.md")

	rows, err := ResolveAssets(root)
	if err != nil {
		t.Fatalf("ResolveAssets must not error over an untracked scratch directory (DW-19), got: %v", err)
	}
	for _, r := range rows {
		if strings.HasPrefix(r.Path, "scratch-fonts/") {
			t.Errorf("ResolveAssets produced a row for the untracked directory: %+v", r)
		}
	}
}

// TestResolveAssetsStillReportsATrackedViolation is AC10's own named
// mutation: pointing the resolver at a TRACKED directory carrying a
// real violation (a committed font binary with no LICENSE* file) must
// still report it — proving DW-19's fix excludes only what git never
// tracked, and does not "trade one blind spot for another" by
// swallowing a real, committed finding.
//
// Finding 7 (QA review): the original version of this test built a
// synthetic repo holding ONLY the violating tracked directory, so the
// masking scenario AC10's mutation exists to rule out — an untracked
// directory sorting ahead of the real violation — was never exercised.
// This version co-locates an untracked "scratch-fonts" directory that
// sorts alphabetically BEFORE "committed-fonts" (s < c is false, so
// rename it to sort first below) to prove the exclusion added for the
// untracked case cannot swallow a later, genuine finding.
func TestResolveAssetsStillReportsATrackedViolation(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	// Sorts alphabetically AHEAD of "committed-fonts" (D-3.6.5's own
	// masking scenario: ".font-sources" < a real asset directory).
	untrackedDir := filepath.Join(root, "aaa-scratch-fonts")
	if err := os.MkdirAll(untrackedDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(untrackedDir, "Untracked.ttf"), []byte("not a real font, just bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Deliberately never git-added — this is the co-present untracked
	// directory Finding 7 says the previous version omitted.

	violatingDir := filepath.Join(root, "committed-fonts")
	if err := os.MkdirAll(violatingDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	fontPath := filepath.Join(violatingDir, "Committed.ttf")
	if err := os.WriteFile(fontPath, []byte("not a real font, just bytes"), 0o644); err != nil {
		t.Fatalf("write font: %v", err)
	}
	// Deliberately NO LICENSE* file alongside it.
	gitAdd(t, root, fontPath)

	_, err := ResolveAssets(root)
	if err == nil {
		t.Fatal("expected ResolveAssets to report the missing LICENSE* file in a TRACKED directory, got nil error")
	}
	if !strings.Contains(err.Error(), "no LICENSE* file") {
		t.Errorf("expected error to mention %q, got: %v", "no LICENSE* file", err)
	}
}

// TestResolveAssetsAllDirectoriesUntrackedIsAScanError is the D-3.6.5
// amendment's "Required" floor (Finding 1, QA review, Blocker):
// dirOrder non-empty (candidate font-bearing directories exist on
// disk) but EVERY one of them resolves to zero git-tracked files is a
// SCAN ERROR — "could not look" — never silently reported as "nothing
// to report" (D-000.9: the two must not produce identical output).
// Anchor: git's own index. Both directories here hold a real
// font-extensioned file on disk and neither is ever git-added — the
// exact shape a directory rename, a wrong repoRoot, or a layout
// refactor would produce.
func TestResolveAssetsAllDirectoriesUntrackedIsAScanError(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	for _, name := range []string{"scratch-one", "scratch-two"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "Untracked.ttf"), []byte("not a real font, just bytes"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	// A trivial tracked file so the repository has at least one commit
	// worth of history to query against, and so dirOrder's two entries
	// are the ONLY font-bearing candidates.
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	gitAdd(t, root, "README.md")

	_, err := ResolveAssets(root)
	if err == nil {
		t.Fatal("expected ResolveAssets to return a scan error when every candidate directory is untracked, got nil")
	}
	if !strings.Contains(err.Error(), "scan error") {
		t.Errorf("expected error to mention %q, got: %v", "scan error", err)
	}
}

// TestResolveAssetsNoCandidateDirectoriesIsNotAScanError pins the
// precision boundary the D-3.6.5 amendment draws explicitly: dirOrder
// being EMPTY (no font-extensioned file exists anywhere on disk) is a
// legitimately empty result, never the scan error above. Without this
// test, TestResolveAssetsAllDirectoriesUntrackedIsAScanError's fix
// could be over-implemented as "zero rows is always an error," which
// would break this repository's own history before any font shipped.
func TestResolveAssetsNoCandidateDirectoriesIsNotAScanError(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	gitAdd(t, root, "README.md")

	rows, err := ResolveAssets(root)
	if err != nil {
		t.Fatalf("ResolveAssets must not error when no candidate font directory exists at all, got: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected zero asset rows, got %d: %v", len(rows), rows)
	}
}
