package folio

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestMatrixDocumentSlugsAreRegisteredInCI closes a drift that has
// already happened once.
//
// matrix_test.go carries `//go:build matrix`, so its matrixDocuments
// list executes only at an epic boundary gate. .github/workflows/
// matrix.yml carries a SECOND, hand-written copy of the same list — the
// `docs="..."` line and the per-target upload-artifact paths. Nothing
// connected the two, so a document registered in one and not the other
// was invisible: Story 2.2 added multi-script-fallback to matrixDocuments
// and never to the workflow, and CI went on publishing four documents'
// worth of legs while comparing three, reporting green. That is Story
// 1.2's own Finding 8 recurring one story later, which is what a
// hand-maintained duplicate list does.
//
// This test is deliberately UNTAGGED — it runs in the ordinary
// `go test ./...` every story runs, not at the gate — and it reads both
// lists as SOURCE TEXT, because the tagged list is not compiled into
// this binary. Reading source is the weaker mechanism; running in every
// story is what makes it worth having.
func TestMatrixDocumentSlugsAreRegisteredInCI(t *testing.T) {
	root := repoRootFromTest(t)

	harness, err := os.ReadFile(filepath.Join(root, "folio-go", "matrix_test.go"))
	if err != nil {
		t.Fatalf("read matrix_test.go: %v", err)
	}
	slugRe := regexp.MustCompile(`(?m)^\s*slug:\s*"([^"]+)"`)
	var declared []string
	for _, m := range slugRe.FindAllStringSubmatch(string(harness), -1) {
		declared = append(declared, m[1])
	}
	if len(declared) == 0 {
		t.Fatal("vacuity guard: found no slug: entries in matrix_test.go — this test would pass on an empty comparison")
	}

	workflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "matrix.yml"))
	if err != nil {
		t.Fatalf("read matrix.yml: %v", err)
	}
	docsRe := regexp.MustCompile(`(?m)^\s*docs="([^"]+)"`)
	dm := docsRe.FindStringSubmatch(string(workflow))
	if dm == nil {
		t.Fatal("could not find the docs=\"...\" list in .github/workflows/matrix.yml")
	}
	inCI := strings.Fields(dm[1])
	if len(inCI) == 0 {
		t.Fatal("vacuity guard: the workflow's docs list is empty")
	}

	if !equalSets(declared, inCI) {
		t.Fatalf(
			"matrixDocuments and .github/workflows/matrix.yml's docs list disagree.\n"+
				"  matrix_test.go: %v\n  matrix.yml:     %v\n"+
				"A document in one list and not the other is a matrix leg nobody compares, reported as green.",
			sorted(declared), sorted(inCI),
		)
	}

	// ...and every document must have an upload-artifact path on every
	// target, or its hashes never reach the comparison job at all.
	targets := []string{"darwin-arm64", "linux-amd64", "linux-arm64", "js-wasm"}
	for _, slug := range declared {
		for _, target := range targets {
			want := fmt.Sprintf("hash.%s.%s.txt", target, slug)
			if !strings.Contains(string(workflow), want) {
				t.Errorf("matrix.yml declares no upload-artifact path %s — that leg's hash never reaches the comparison job", want)
			}
		}
	}

	t.Logf("matrix registration witness — %d documents registered in both matrixDocuments and matrix.yml, across %d targets", len(declared), len(targets))
}

func equalSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := sorted(a), sorted(b)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
