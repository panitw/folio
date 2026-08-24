//go:build matrix

// fontgen_matrix_test.go is Story 2.2's REGENERATION test (D-2.2.4).
//
// The three shipped faces under folio-go/fonts/ are static, Regular-only
// instances DERIVED from upstream variable builds, and the derived files
// are committed rather than produced at build time. That choice is what
// keeps the shipped font from being a function of the build environment
// — a different fontTools produces a different font, which produces a
// different PDF, which is AD-22's drift class reintroduced at the asset
// layer. The cost of the choice is that nothing in an ordinary build
// re-checks that the committed bytes are still what the recorded
// derivation produces. This file is that check.
//
// WHY IT CARRIES THE `matrix` BUILD TAG. Same reasoning as
// matrix_test.go, and the same pair of properties:
//
//   - It is COMPILED on every story — CI runs `go build -tags=matrix
//     ./...` and `go vet -tags=matrix ./...` in the folio-go job — so it
//     cannot silently rot while nobody is looking at it.
//   - It is RUN at the epic gate, alongside the four-target byte-identity
//     harness, because a full run needs the upstream sources and a pinned
//     fontTools that an ordinary `go test ./...` has no business
//     requiring.
//
// IT DOES NOT SKIP WHEN ITS INPUTS ARE ABSENT. A t.Skip here would make
// "the sources were not present" indistinguishable from "the faces
// reproduce", which is D-000.9's absence-reads-as-success in the one
// place it would be least visible. When the sources are missing the test
// FAILS, and the failure message names every URL and sha256 needed to
// obtain them.
package folio

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fontSourcesDir is where the upstream VARIABLE builds are expected.
// Overridable so the epic gate can point at a cache; the default matches
// the Makefile's FONT_SOURCES.
func fontSourcesDir(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("FOLIO_FONT_SOURCES"); dir != "" {
		return dir
	}
	return filepath.Join(repoRootFromTest(t), ".font-sources")
}

// TestShippedFacesReproduceFromUpstream re-runs the recorded derivation
// and asserts the committed faces come back byte-for-byte.
//
// The generator does the comparison itself, in both directions — it
// verifies each SOURCE's sha256 before instancing and each PRODUCED
// file's sha256 after, refuses to compare an operand it did not
// successfully read, and prints an explicit "N of N" coverage witness.
// That last part is not ceremony: an earlier reproducibility check in
// this story's own history printed IDENTICAL for four faces that were
// never generated, because two failed hash computations compared empty
// to empty.
func TestShippedFacesReproduceFromUpstream(t *testing.T) {
	root := repoRootFromTest(t)
	sources := fontSourcesDir(t)

	if _, err := os.Stat(sources); err != nil {
		t.Fatalf(
			"the upstream variable font sources are not at %s, so the shipped faces' derivation cannot be "+
				"replayed.\n\n"+
				"This test FAILS rather than skipping on purpose: a skip would make \"the sources were not "+
				"present\" read exactly like \"the faces reproduce\".\n\n"+
				"Obtain them (each face's NOTICE.md carries the release URL and the source sha256):\n"+
				"  NotoSans-VF.ttf      bfb7bb691513f12e734dc346c03a03f784912432d7e3fa8e56efcf906fe86b3d\n"+
				"  NotoSansThai-VF.ttf  5a1c559bb539583c8a1fd99d1c5b9491e5e14478c9cd2bd0970d5c3096cc9ef8\n"+
				"  NotoSansSC-VF.ttf    a3041811a78c361b1de50f953c805e0244951c21c5bd412f7232ef0d899af0da\n\n"+
				"then either set FOLIO_FONT_SOURCES=<dir> or place them in %s and re-run.\n"+
				"stat error: %v",
			sources, sources, err,
		)
	}

	python := os.Getenv("FOLIO_FONTGEN_PYTHON")
	if python == "" {
		python = "python3"
	}
	script := filepath.Join(root, "tools", "fontgen", "instance_faces.py")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("the generator %s is missing, so the derivation is no longer recorded anywhere: %v", script, err)
	}

	cmd := exec.Command(python, script,
		"--sources", sources,
		"--repo-root", root,
		"--verify-only",
	)
	out, err := cmd.CombinedOutput()
	text := string(out)
	t.Logf("fontgen --verify-only output:\n%s", text)

	if err != nil {
		t.Fatalf(
			"the committed shipped faces do NOT reproduce from their recorded derivation.\n\n"+
				"Under AD-21/AD-22 this is a defect until proven to be an intended, versioned change: the "+
				"shipped font is an input to every font-bearing golden in this repo, so a face that has "+
				"moved moves all of them.\n\n"+
				"Check the toolchain first — the derivation is pinned to Python 3.12.13 / fontTools 4.63.0, "+
				"and fontTools is NOT byte-deterministic across versions.\n\nexit: %v",
			err,
		)
	}

	// Assert the generator's own coverage witness, rather than trusting a
	// zero exit status (D-000.13: assert on the reported result, never on
	// the exit code alone).
	const wantWitness = "derived and compared 3 of 3 faces"
	if !strings.Contains(text, wantWitness) {
		t.Fatalf(
			"the generator exited 0 but its output does not contain the coverage witness %q. "+
				"A run that compared nothing exits 0 too.\ngot:\n%s",
			wantWitness, text,
		)
	}
	if !strings.Contains(text, "fontgen: OK") {
		t.Fatalf("the generator exited 0 without printing its success line.\ngot:\n%s", text)
	}
}
