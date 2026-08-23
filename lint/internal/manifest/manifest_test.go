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
// `go run ./lint/cmd/genmanifest` from the repo root regenerates it.
func TestManifestUpToDate(t *testing.T) {
	root := repoRootFromTest(t)

	rows, err := Generate(root)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := Render(rows)

	committedPath := filepath.Join(root, "lint", "MANIFEST.md")
	got, err := os.ReadFile(committedPath)
	if err != nil {
		t.Fatalf("read committed manifest %s: %v (run `go run ./lint/cmd/genmanifest`)", committedPath, err)
	}

	if string(got) != want {
		t.Fatalf("lint/MANIFEST.md is out of date — run `go run ./lint/cmd/genmanifest` from the repo root and commit the result")
	}
}
