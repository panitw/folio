// Command genmanifest regenerates lint/MANIFEST.md (AC19) from the live
// module graphs of all three of the repo's Go modules. Run it from the
// repo root: `go run ./lint/cmd/genmanifest`. TestManifestUpToDate in
// lint/internal/manifest asserts the committed file matches what this
// command would produce, so a new dependency that changes the graph
// must be regenerated and committed, not silently drift.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/panitw/folio/lint/internal/manifest"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "genmanifest:", err)
		os.Exit(1)
	}
	rows, err := manifest.Generate(repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "genmanifest:", err)
		os.Exit(1)
	}
	out := filepath.Join(repoRoot, "lint", "MANIFEST.md")
	if err := os.WriteFile(out, []byte(manifest.Render(rows)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "genmanifest:", err)
		os.Exit(1)
	}
	fmt.Println("wrote", out)
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "folio-go")); err == nil && info.IsDir() {
			if info2, err := os.Stat(filepath.Join(dir, "lint")); err == nil && info2.IsDir() {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root (a directory containing both folio-go/ and lint/) walking up from %s", dir)
		}
		dir = parent
	}
}
