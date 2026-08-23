package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// repoRootFromTest finds the folio repository root by walking up from
// the current working directory (Go test binaries run with cwd set to
// the package directory) until it finds a directory containing both
// folio-go/ and lint/ — the same D-000.5/AD-21 pattern folio-go's own
// repoRootFromTest helpers use, duplicated here because it lives only in
// a _test.go file and lint/ is a separate module.
func repoRootFromTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if isRepoRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root (a directory containing both folio-go/ and lint/) walking up from %s", dir)
		}
		dir = parent
	}
}

func isRepoRoot(dir string) bool {
	folioGo, err1 := os.Stat(filepath.Join(dir, "folio-go"))
	lintDir, err2 := os.Stat(filepath.Join(dir, "lint"))
	return err1 == nil && folioGo.IsDir() && err2 == nil && lintDir.IsDir()
}

// assertExactFindings implements AC1's "by file and rule, never by
// count" fixture assertion (RP-3c): a scan that finds the right *number*
// of wrong things, but the wrong ones, must still fail. It compares
// distinct (path, rule) pairs, not raw finding counts, since a single
// file may legitimately trip a rule at more than one AST site.
func assertExactFindings(t *testing.T, got []Finding, want []Finding) {
	t.Helper()
	gotSet := map[[2]string]bool{}
	for _, f := range got {
		gotSet[[2]string{f.Path, f.Rule}] = true
	}
	wantSet := map[[2]string]bool{}
	for _, f := range want {
		wantSet[[2]string{f.Path, f.Rule}] = true
	}
	for k := range wantSet {
		if !gotSet[k] {
			t.Errorf("expected finding not reported: file=%s rule=%s", k[0], k[1])
		}
	}
	for k := range gotSet {
		if !wantSet[k] {
			t.Errorf("unexpected finding reported: file=%s rule=%s", k[0], k[1])
		}
	}
	if len(gotSet) != len(wantSet) {
		t.Errorf("distinct (file,rule) pair count mismatch: got %d, want %d", len(gotSet), len(wantSet))
	}
}
