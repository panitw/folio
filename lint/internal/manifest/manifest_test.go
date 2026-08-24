package manifest

import (
	"os"
	"path/filepath"
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

	committedPath := filepath.Join(root, "lint", "MANIFEST.md")
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
