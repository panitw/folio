package manifest

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/panitw/folio/lint/internal/licence"
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
// right now. A dependency added to a Go module graph or the designer lockfile
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

func TestDesignerManifestClassifiesRuntimeAndBuildDependencies(t *testing.T) {
	rows, err := Generate(repoRootFromTest(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"react": "shipped", "react-dom": "shipped", "vite": "build-time-only"}
	for module, shippedBy := range want {
		found := false
		for _, row := range rows {
			if row.Serves == "folio-designer" && row.Module == module {
				found = true
				if row.ShippedBy != shippedBy {
					t.Errorf("%s classification = %q, want %q", module, row.ShippedBy, shippedBy)
				}
			}
		}
		if !found {
			t.Errorf("designer manifest is missing %s", module)
		}
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
	// Story 8.4h (task 8): this fixture carried the string "a licence",
	// which MEASURES as (FamilyUnknown, "") — harmless while an
	// unclassifiable licence produced a "SEE NOTICE" row, fatal the
	// moment it became a build failure (AC1). Upgraded to a
	// classifiable, ALLOWLISTED marker so this test keeps testing what
	// its name says — the untracked-directory exclusion — rather than
	// tripping over the new refusal. This is repairing a fixture whose
	// premise the gate changed, not weakening a claim: the subject
	// under test is git's index, not licence classification.
	if err := os.WriteFile(licensePath, []byte("SPDX-License-Identifier: OFL-1.1\n"), 0o644); err != nil {
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

// ---------------------------------------------------------------------------
// Story 8.4h (AC1, AC2, AC3, D-8.5.13): THE POPULATION-INDEPENDENT RED-PROOFS.
//
// WHY SYNTHETIC AND NOT THE REAL TREE. lint/internal/rules/fontsassets_test.go
// red-proves its guards by MUTATING the live repository's own files and
// restoring them in t.Cleanup. That shape proves the guard fires for the faces
// committed TODAY, and stops proving anything the moment the population
// changes — a new face, a moved directory, a story that ships a catalogue.
// These proofs instead build a NEW, OTHERWISE-EMPTY directory in a fresh
// scratch repository, so the claim they establish is about the GATE, not about
// the tree: proved this way, any real face in any directory is covered by
// construction (D-8.5.4, D-8.5.8b).
//
// EACH ARM ASSERTS A SUBSTRING UNIQUE TO ITS OWN MESSAGE. "Something failed"
// is not a proof of anything — a test that reds on a neighbouring guard (the
// missing-LICENSE refusal, the missing-NOTICE refusal, the D-3.6.5 scan-error
// floor) would keep passing after the refusal it claims to prove was deleted.
// The three font-site refusals are deliberately worded so no substring below
// matches more than one of them.
// ---------------------------------------------------------------------------

// writeSyntheticFontDirectory builds one tracked, otherwise-complete font
// directory — a font binary, a LICENSE carrying licenceText, and a NOTICE with
// a copyright line — inside an already-initialised scratch repository, and git
// ADDS all three. `git add` only, never a commit: gitTrackedFileCount reads
// `git ls-files`, which reads the INDEX.
//
// The font "binary" is the same 27 bytes the DW-19 tests use. The walk keys on
// EXTENSION alone, never on content, so no real sfnt is needed and using one
// would make the fixture claim more than the gate reads.
func writeSyntheticFontDirectory(t *testing.T, root, name, licenceText string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", name, err)
	}
	files := map[string]string{
		"Synthetic.ttf": "not a real font, just bytes",
		"LICENSE.txt":   licenceText,
		"NOTICE":        "Copyright 2026 Test\n",
	}
	var paths []string
	for base, content := range files {
		full := filepath.Join(dir, base)
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
		paths = append(paths, full)
	}
	sort.Strings(paths)
	gitAdd(t, root, paths...)
	return dir
}

// scratchRepoWithFontDirectory is the whole recipe: a fresh temp repository
// holding exactly ONE font directory, carrying licenceText. Nothing else in
// the tree is font-extensioned, so whatever ResolveAssets says is about THIS
// directory and cannot be an artifact of a neighbour.
func scratchRepoWithFontDirectory(t *testing.T, licenceText string) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	writeSyntheticFontDirectory(t, root, "synthetic-fonts", licenceText)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	gitAdd(t, root, "README.md")
	return root
}

// TestResolveAssetsRefusesAnUnclassifiableFontLicence is AC2's first arm and
// the hole D-8.5.13's charter names directly: a licence text that classifies
// to nothing used to fall through to the literal label "SEE NOTICE" and
// produce a clean row on a clean build.
//
// "a licence" is the measured unclassifiable input — the same string this
// file's own DW-19 fixture carried until Story 8.4h repaired it.
func TestResolveAssetsRefusesAnUnclassifiableFontLicence(t *testing.T) {
	root := scratchRepoWithFontDirectory(t, "a licence")

	rows, err := ResolveAssets(root)
	if err == nil {
		t.Fatalf("expected ResolveAssets to refuse an unclassifiable font licence, got nil and %d row(s): %v", len(rows), rows)
	}
	// THIS ARM'S OWN MESSAGE. Not shared with the copyleft arm, not shared
	// with the off-allowlist arm, and not produced by any neighbouring guard.
	if !strings.Contains(err.Error(), "could not be classified") {
		t.Errorf("expected the UNCLASSIFIABLE refusal, got: %v", err)
	}
	if !strings.Contains(err.Error(), "synthetic-fonts") {
		t.Errorf("expected the error to LOCATE the directory (AD-14), got: %v", err)
	}
	if rows != nil {
		t.Errorf("a refused directory must produce NO row, got: %v", rows)
	}
}

// TestResolveAssetsRefusesACopyleftFontLicence is AC2's second arm, and the
// hole the charter did NOT name (Design Note 1, measured): a GPL-3.0 licence
// text never reached the "SEE NOTICE" fall-through at all. It classifies
// perfectly well, and shipped a clean row labelled "GPL-3.0" on a clean build,
// because both asset sites DISCARDED the classifier's Family return and
// nothing compared the id to any list.
//
// It must red on the COPYLEFT message specifically, not on the
// off-allowlist one: a copyleft licence is refused BY NAME, not merely by
// absence from the owner's four ids.
func TestResolveAssetsRefusesACopyleftFontLicence(t *testing.T) {
	root := scratchRepoWithFontDirectory(t, "GNU GENERAL PUBLIC LICENSE\nVersion 3, 29 June 2007\n")

	rows, err := ResolveAssets(root)
	if err == nil {
		t.Fatalf("expected ResolveAssets to refuse a GPL-3.0 font licence, got nil and %d row(s): %v", len(rows), rows)
	}
	if !strings.Contains(err.Error(), "a copyleft licence") {
		t.Errorf("expected the COPYLEFT refusal, got: %v", err)
	}
	// Explicitly NOT the unclassifiable arm's message — this is the
	// assertion that keeps the two proofs independent.
	if strings.Contains(err.Error(), "could not be classified") {
		t.Errorf("a GPL-3.0 text classifies fine; it must not be refused as unclassifiable: %v", err)
	}
	if !strings.Contains(err.Error(), `"GPL-3.0"`) {
		t.Errorf("expected the error to NAME the classified identifier, got: %v", err)
	}
	if !strings.Contains(err.Error(), "synthetic-fonts") {
		t.Errorf("expected the error to LOCATE the directory (AD-14), got: %v", err)
	}
	if rows != nil {
		t.Errorf("a refused directory must produce NO row, got: %v", rows)
	}
}

// TestResolveAssetsRefusesAPermissiveButOffAllowlistFontLicence is the arm
// that proves the font site enforces THE OWNER'S FOUR IDS and not merely
// "permissive". ISC is in licence.permissiveSPDX and would pass any
// family-only or IsPermissiveSPDX-based check — so if a later change routes
// the font path through the dependency-side permissive set (the collapse
// Design Note 6 warns about, in its other direction), this test is what reds.
func TestResolveAssetsRefusesAPermissiveButOffAllowlistFontLicence(t *testing.T) {
	root := scratchRepoWithFontDirectory(t, "SPDX-License-Identifier: ISC\n")

	rows, err := ResolveAssets(root)
	if err == nil {
		t.Fatalf("expected ResolveAssets to refuse an off-allowlist font licence, got nil and %d row(s): %v", len(rows), rows)
	}
	if !strings.Contains(err.Error(), "not one of the licences permitted") {
		t.Errorf("expected the OFF-ALLOWLIST refusal, got: %v", err)
	}
	if strings.Contains(err.Error(), "could not be classified") || strings.Contains(err.Error(), "a copyleft licence") {
		t.Errorf("ISC classifies as permissive; it must be refused on its own message, not a neighbouring arm's: %v", err)
	}
	if !strings.Contains(err.Error(), `"ISC"`) {
		t.Errorf("expected the error to NAME the classified identifier, got: %v", err)
	}
	if rows != nil {
		t.Errorf("a refused directory must produce NO row, got: %v", rows)
	}
}

// TestResolveAssetsAcceptsEveryAllowlistedFontLicence is the green arm — the
// one that stops the three refusals above from being satisfiable by a gate
// that simply refuses everything. All FOUR of the owner's ids are exercised,
// including Ubuntu-font-1.0, which has no asset in this repository and
// therefore no other live witness (Design Note 8).
func TestResolveAssetsAcceptsEveryAllowlistedFontLicence(t *testing.T) {
	for _, c := range []struct{ name, licenceText, wantID string }{
		{"OFL-1.1 by marker", "SIL OPEN FONT LICENSE\nVersion 1.1 - 26 February 2007\n", "OFL-1.1"},
		{"Apache-2.0 by marker", "Apache License\nVersion 2.0, January 2004\n", "Apache-2.0"},
		{"MIT by marker", "MIT License\n\nPermission is hereby granted, free of charge\n", "MIT"},
		{"Ubuntu-font-1.0 by marker", "UBUNTU FONT LICENCE Version 1.0\n", "Ubuntu-font-1.0"},
		{"Ubuntu-font-1.0 by SPDX line", "SPDX-License-Identifier: Ubuntu-font-1.0\n", "Ubuntu-font-1.0"},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := scratchRepoWithFontDirectory(t, c.licenceText)

			rows, err := ResolveAssets(root)
			if err != nil {
				t.Fatalf("an allowlisted licence must resolve without error, got: %v", err)
			}
			if len(rows) != 1 {
				t.Fatalf("expected exactly one row, got %d: %v", len(rows), rows)
			}
			if rows[0].Licence != c.wantID {
				t.Errorf("row licence = %q, want %q", rows[0].Licence, c.wantID)
			}
			if rows[0].Licence == "SEE NOTICE" {
				t.Error(`"SEE NOTICE" is unreachable as a passing outcome after Story 8.4h (AC4)`)
			}
		})
	}
}

// TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist is AC3, and
// Design Note 6's THE-COLLAPSE-HAS-HAPPENED detector. It is stated as a test
// rather than as a principle because a shared string is not the tell — the
// tell is that Site B starts refusing CC0-1.0.
//
// CC0-1.0 is permissive, is the Thai dictionary's real and legitimate licence
// (Story 2.1, D-2.1.3), and is NOT one of the owner's four font ids. So a
// future change that "tidies" the two sites onto one constant reds HERE, on a
// test whose name says why, instead of reddening the whole build on a shipped,
// owner-unobjected asset — or, in the other direction, admitting CC0 fonts.
func TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist(t *testing.T) {
	// The wordlist path has no git-tracked filter — it is gated only on
	// os.Stat of words_th.txt — so a bare temp tree is the whole fixture.
	scratchWordlist := func(t *testing.T, licenceText string) string {
		t.Helper()
		root := t.TempDir()
		dir := filepath.Join(root, filepath.FromSlash(wordlistAssetDir))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		for base, content := range map[string]string{
			"words_th.txt":        "คำ\n",
			"LICENSE-CC0-1.0.txt": licenceText,
			"NOTICE":              "Copyright: none asserted\n",
		} {
			if err := os.WriteFile(filepath.Join(dir, base), []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", base, err)
			}
		}
		return root
	}

	t.Run("CC0-1.0 passes — it is NOT one of the four font ids", func(t *testing.T) {
		root := scratchWordlist(t, "Creative Commons Legal Code\n\nCC0 1.0 Universal\n")

		row, ok, err := resolveWordlistAssetRow(root)
		if err != nil {
			t.Fatalf("the wordlist's own CC0 legal code must resolve without error — "+
				"refusing it is the signature of Site B having been collapsed onto the font "+
				"allowlist (Design Note 6). Got: %v", err)
		}
		if !ok {
			t.Fatal("expected a wordlist row")
		}
		if row.Licence != "CC0-1.0" {
			t.Errorf("wordlist row licence = %q, want %q", row.Licence, "CC0-1.0")
		}
	})

	t.Run("an unclassifiable wordlist licence is refused, in this function's own voice", func(t *testing.T) {
		root := scratchWordlist(t, "a licence")

		_, ok, err := resolveWordlistAssetRow(root)
		if err == nil {
			t.Fatalf("expected an unclassifiable wordlist licence to be refused, got nil (ok=%v)", ok)
		}
		if !strings.Contains(err.Error(), "wordlist licence text could not be classified") {
			t.Errorf("expected Site B's OWN unclassifiable refusal, got: %v", err)
		}
		// AC9, not AC25: the wordlist site cites its own criterion.
		if !strings.Contains(err.Error(), "(AC9, AD-26)") {
			t.Errorf("expected the wordlist function's own citation voice, got: %v", err)
		}
		if !strings.Contains(err.Error(), wordlistAssetDir) {
			t.Errorf("expected the error to LOCATE the wordlist (AD-14), got: %v", err)
		}
	})

	t.Run("a copyleft wordlist licence is refused", func(t *testing.T) {
		root := scratchWordlist(t, "GNU GENERAL PUBLIC LICENSE\nVersion 3\n")

		if _, _, err := resolveWordlistAssetRow(root); err == nil {
			t.Fatal("expected a GPL wordlist licence to be refused, got nil")
		} else if !strings.Contains(err.Error(), "a copyleft licence") {
			t.Errorf("expected Site B's copyleft refusal, got: %v", err)
		}
	})
}

// TestTheTwoAssetSitesDoNotShareAPolicy states the two-populations invariant
// directly, as a property of the two lists rather than of one input: there is
// an id each site accepts and the other refuses, in BOTH directions. A shared
// constant, a shared helper or a shared list cannot satisfy this test.
func TestTheTwoAssetSitesDoNotShareAPolicy(t *testing.T) {
	// CC0-1.0: legitimate for the wordlist, forbidden for a font.
	if fontAssetLicenceAllowed["CC0-1.0"] {
		t.Error(`CC0-1.0 must NEVER be in the four-id font asset allowlist — adding it to fix a ` +
			`wordlist scoping error would amend an owner decision (D-8.5.3)`)
	}
	if !licence.IsPermissiveSPDX("CC0-1.0") {
		t.Error("CC0-1.0 must remain in the permissive SPDX set — the shipped Thai wordlist is under it (D-2.1.3)")
	}
	// Ubuntu-font-1.0: on the font allowlist AND permissive; ISC: permissive
	// but NOT an acceptable font licence. The second is the direction that
	// proves the font list is not merely "the permissive set".
	if !fontAssetLicenceAllowed["Ubuntu-font-1.0"] {
		t.Error("Ubuntu-font-1.0 is the owner's fourth id (D-8.5.3) and must be on the font allowlist")
	}
	if fontAssetLicenceAllowed["ISC"] {
		t.Error("ISC is permissive but is not one of the owner's four ids; the font site is not the permissive set")
	}
	if !licence.IsPermissiveSPDX("ISC") {
		t.Error("ISC must remain in the permissive SPDX set")
	}
}

// TestCommittedAssetPopulationClassifiesCleanly is AC6, and the assertion that
// makes "the gate does not land red" a fact the build RE-CHECKS rather than a
// claim made once at implementation time. A gate that lands red is not a gate;
// a gate proved green only by a human running a command once is a gate that
// silently rots.
//
// Deliberately NOT a cardinality assertion over the population (D-8.5.4): a
// count written next to the thing it counts stops being true the moment the
// thing grows, and this population grows by design at Story 8.5. What is
// asserted instead is a PROPERTY of every row — which stays true however many
// rows there are, and reds the moment one of them is not permitted.
func TestCommittedAssetPopulationClassifiesCleanly(t *testing.T) {
	root := repoRootFromTest(t)

	rows, err := ResolveAssets(root)
	if err != nil {
		t.Fatalf("the committed asset population must classify cleanly IN THE SAME COMMIT that makes "+
			"classification fatal — halt and report the directory rather than weakening the gate or "+
			"adding an exemption (D-8.5.13). Got: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("ResolveAssets returned no rows at all — this assertion would be vacuous")
	}

	const wordlistPath = wordlistAssetDir + "/words_th.txt"
	sawWordlist := false
	for _, row := range rows {
		if row.Licence == "SEE NOTICE" {
			t.Errorf("%s: no generated manifest row may read %q after Story 8.4h (AC4)", row.Path, "SEE NOTICE")
		}
		if row.Path == wordlistPath {
			sawWordlist = true
			// The wordlist is under the OTHER policy, deliberately.
			if row.Licence != "CC0-1.0" {
				t.Errorf("wordlist row licence = %q, want %q", row.Licence, "CC0-1.0")
			}
			continue
		}
		if !fontAssetLicenceAllowed[row.Licence] {
			t.Errorf("%s: font row licence %q is not one of the owner's four ids %v (AC6, D-8.5.3)",
				row.Path, row.Licence, fontAssetLicenceAllowlist)
		}
	}
	if !sawWordlist {
		t.Errorf("expected a row for %s — its absence would make the two-policy assertion above vacuous", wordlistPath)
	}
}
