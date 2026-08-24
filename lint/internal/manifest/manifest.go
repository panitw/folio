// Package manifest generates AD-26's third-party licence manifest
// (AC19, AC20): every module in each of the repo's three Go module
// graphs, with a resolved licence, labelled with the module it serves
// and whether that module is shipped or build-time-only.
package manifest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/panitw/folio/lint/internal/licence"
)

// servedModule is one of the repo's three real Go modules, and whether
// what it produces is shipped to consumers or exists only to build/test
// the repo (D-1.3.9's shipped/build-time-only label — it does not soften
// AC18's uniform ban; it exists so a future "this copyleft tool is only
// used at build time" argument has to be made explicitly, to the owner).
type servedModule struct {
	dirName   string // relative to repo root
	name      string // as it appears in the manifest's "serves" column
	shippedBy string // "shipped" or "build-time-only"
}

// servedModules is deliberately not derived from a directory listing:
// AD-26 binds all three named modules explicitly (D-1.3.9), and a
// silently-discovered fourth module must not silently join the ban's
// scope without a developer noticing it in a diff.
var servedModules = []servedModule{
	{dirName: "folio-go", name: "folio-go", shippedBy: "shipped"},
	{dirName: "hashmatrix", name: "hashmatrix", shippedBy: "build-time-only"},
	{dirName: "lint", name: "lint", shippedBy: "build-time-only"},
}

// Row is one manifest entry: a resolved dependency, the licence it
// classified as, which repo module it serves, and whether that module
// ships to consumers.
type Row struct {
	Module    string
	Version   string
	Licence   string
	Serves    string
	ShippedBy string
}

// Generate resolves all three of the repo's Go module graphs and
// returns the manifest as a sorted set of rows (AC19). A module whose
// licence cannot be resolved is still included, with Licence set to
// "UNRESOLVED" — Generate itself never fails the build; that is AC18's
// job (rules.ScanLicenceGraph), kept as a separate concern so the
// manifest can always be regenerated and inspected even while red.
func Generate(repoRoot string) ([]Row, error) {
	var rows []Row
	for _, sm := range servedModules {
		modDir := filepath.Join(repoRoot, sm.dirName)
		modules, err := licence.ResolveGraph(modDir)
		if err != nil {
			return nil, fmt.Errorf("resolve %s graph: %w", sm.dirName, err)
		}
		for _, m := range modules {
			row := Row{Module: m.Path, Version: m.Version, Serves: sm.name, ShippedBy: sm.shippedBy}
			text, ok := licence.ReadLicenceText(m.Dir)
			if !ok {
				row.Licence = "UNRESOLVED"
			} else if _, spdx := licence.ClassifyLicenceText(text); spdx != "" {
				row.Licence = spdx
			} else {
				row.Licence = "UNRESOLVED"
			}
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Serves != rows[j].Serves {
			return rows[i].Serves < rows[j].Serves
		}
		return rows[i].Module < rows[j].Module
	})
	return rows, nil
}

// fontExtensions are the binary font formats this scan recognises.
var fontExtensions = []string{".ttf", ".otf", ".ttc"}

// assetServesLabel derives AC25's manifest "Serves" label from an
// asset's directory, relative to repoRoot (forward-slash form). Finding
// 9 (QA review): the previous implementation scanned only a fixed,
// two-entry directory list, so a font committed ANYWHERE else — a
// typo'd path, a stray copy — was invisible to the accounting even
// though AC25 reads "ANY committed font binary". This function still
// gives the two known locations their specific, already-published
// labels; anything else still gets accounted for (LICENSE*/NOTICE*
// still required, still fails the build if missing), under a label
// that names where it was actually found rather than silently matching
// one of the two known locations.
func assetServesLabel(relDir string) string {
	switch {
	case relDir == "folio-go/testdata/fonts" || strings.HasPrefix(relDir, "folio-go/testdata/fonts/"):
		return "folio-go test fixture"
	case relDir == "folio-go/fonts" || strings.HasPrefix(relDir, "folio-go/fonts/"):
		return "folio-go shipped"
	default:
		return "committed asset (" + relDir + ")"
	}
}

// AssetRow is one committed non-code redistributed asset (AD-26,
// D-1.5.6, AC25): a font binary, the directory it was found in, and
// (once resolvable) the licence and copyright line its accompanying
// notice files carry.
type AssetRow struct {
	Path      string // relative to repo root
	Licence   string
	Copyright string
	Serves    string
}

// ResolveAssets walks the WHOLE repository (Finding 9, QA review — a
// fixed two-entry directory list missed a font committed anywhere else,
// proved by construction: a font at folio-go/assets/Stray.ttf, or at
// folio-go/testdata/fonts/extra/Second.ttf, both stayed green under the
// old implementation) for committed font binaries and requires each
// directory containing one to carry both a licence text file
// (LICENSE*) and a notice file (NOTICE*) naming a copyright line
// (AC25's binding property: "Any committed font binary — fixture or
// shipped — travels with its licence text and copyright line"). A font
// binary present without both is a build failure, not a silent gap —
// mirroring AC18/AC19's "unresolved licence fails the build" treatment
// of the Go-module half of this same rule (AD-26).
//
// The walk excludes ".git" (not a source tree) and
// "*/testdata/lint/**" (lint's OWN guard fixtures — e.g. ScanEmbedFont's
// evading-shape fixtures under folio-go/testdata/lint/embed-font/,
// which deliberately carry .ttf-extensioned files with no real font
// content and no LICENSE/NOTICE, because they exist to test detection,
// not to be redistributed). Every other font-extensioned file in the
// repository, at any depth, is in scope.
func ResolveAssets(repoRoot string) ([]AssetRow, error) {
	fontsByDir := map[string][]string{} // relative dir (slash form) -> font file basenames
	var dirOrder []string

	walkErr := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			if d.Name() == "lint" && filepath.Base(filepath.Dir(path)) == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		isFont := false
		for _, fe := range fontExtensions {
			if ext == fe {
				isFont = true
				break
			}
		}
		if !isFont {
			return nil
		}
		rel, rerr := filepath.Rel(repoRoot, path)
		if rerr != nil {
			return rerr
		}
		rel = filepath.ToSlash(rel)
		dir := filepath.ToSlash(filepath.Dir(rel))
		if _, seen := fontsByDir[dir]; !seen {
			dirOrder = append(dirOrder, dir)
		}
		fontsByDir[dir] = append(fontsByDir[dir], filepath.Base(rel))
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s for redistributed font assets: %w", repoRoot, walkErr)
	}

	sort.Strings(dirOrder)

	var rows []AssetRow
	for _, dir := range dirOrder {
		fontFiles := append([]string(nil), fontsByDir[dir]...)
		sort.Strings(fontFiles)

		absDir := filepath.Join(repoRoot, filepath.FromSlash(dir))
		entries, err := os.ReadDir(absDir)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", absDir, err)
		}

		var licenceText, noticeText string
		var haveLicence, haveNotice bool
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, "LICENSE") {
				b, rerr := os.ReadFile(filepath.Join(absDir, name))
				if rerr != nil {
					return nil, fmt.Errorf("read %s: %w", filepath.Join(absDir, name), rerr)
				}
				licenceText = string(b)
				haveLicence = true
			}
			if strings.HasPrefix(name, "NOTICE") {
				b, rerr := os.ReadFile(filepath.Join(absDir, name))
				if rerr != nil {
					return nil, fmt.Errorf("read %s: %w", filepath.Join(absDir, name), rerr)
				}
				noticeText = string(b)
				haveNotice = true
			}
		}

		if !haveLicence {
			return nil, fmt.Errorf("%s: contains a committed font binary but no LICENSE* file (AC25, AD-26)", dir)
		}
		if !haveNotice {
			return nil, fmt.Errorf("%s: contains a committed font binary but no NOTICE* file naming its copyright line (AC25, AD-26)", dir)
		}

		copyrightLine, ok := extractCopyrightLine(noticeText)
		if !ok {
			return nil, fmt.Errorf("%s: NOTICE file does not contain a line starting with \"Copyright\" (AC25, AD-26)", dir)
		}

		licenceLabel := "SEE NOTICE"
		if _, spdx := licence.ClassifyLicenceText(licenceText); spdx != "" {
			licenceLabel = spdx
		}

		serves := assetServesLabel(dir)
		for _, ff := range fontFiles {
			rows = append(rows, AssetRow{
				Path:      dir + "/" + ff,
				Licence:   licenceLabel,
				Copyright: copyrightLine,
				Serves:    serves,
			})
		}
	}

	// The CC0 wordlist (Story 2.1, AC9, AD-26) is a second, distinct
	// declared asset location — NOT font-extensioned, so the walk above
	// never sees it (that gap is exactly what motivated AC9's own
	// separate fail-closed guard, lint/internal/rules/wordlistassets.go
	// — D-1.8.11's premise, confirmed with a second live instance).
	// Resolved explicitly here, rather than widening fontExtensions
	// (which would be the forbidden fail-open shortcut for a completely
	// different reason: fontExtensions is keyed on FONT files, and a
	// wordlist is not one).
	if wordlistRow, ok, err := resolveWordlistAssetRow(repoRoot); err != nil {
		return nil, err
	} else if ok {
		rows = append(rows, wordlistRow)
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	return rows, nil
}

// wordlistAssetDir is the same declared location AC9's fail-closed
// guard scans (lint/internal/rules/wordlistassets.go's
// wordlistAssetLocation) — duplicated here deliberately rather than
// imported, matching this codebase's existing precedent of duplicating
// small helpers (e.g. repoRootFromTest) across packages rather than
// creating a shared-utility import just to avoid one constant.
const wordlistAssetDir = "folio-go/internal/text/wordlist"

// resolveWordlistAssetRow produces the wordlist's AssetRow if the
// declared location exists and carries its wordlist, licence text and
// NOTICE (AC9's OWN guard is what enforces "exists and is complete" —
// this function reports "not present" rather than erroring when the
// location is simply absent, e.g. at a revision before Story 2.1).
func resolveWordlistAssetRow(repoRoot string) (AssetRow, bool, error) {
	absDir := filepath.Join(repoRoot, filepath.FromSlash(wordlistAssetDir))
	if _, err := os.Stat(filepath.Join(absDir, "words_th.txt")); err != nil {
		return AssetRow{}, false, nil
	}

	licenceText, err := os.ReadFile(filepath.Join(absDir, "LICENSE-CC0-1.0.txt"))
	if err != nil {
		return AssetRow{}, false, fmt.Errorf("%s: wordlist present but licence text unreadable (AC9, AD-26): %w", wordlistAssetDir, err)
	}
	noticeText, err := os.ReadFile(filepath.Join(absDir, "NOTICE"))
	if err != nil {
		return AssetRow{}, false, fmt.Errorf("%s: wordlist present but NOTICE unreadable (AC9, AD-26): %w", wordlistAssetDir, err)
	}

	copyrightLine, ok := extractCopyrightLine(string(noticeText))
	if !ok {
		return AssetRow{}, false, fmt.Errorf("%s: NOTICE file does not contain a line starting with \"Copyright\" (AC25-equivalent for AC9, AD-26)", wordlistAssetDir)
	}

	licenceLabel := "SEE NOTICE"
	if _, spdx := licence.ClassifyLicenceText(string(licenceText)); spdx != "" {
		licenceLabel = spdx
	}

	return AssetRow{
		Path:      wordlistAssetDir + "/words_th.txt",
		Licence:   licenceLabel,
		Copyright: copyrightLine,
		Serves:    "folio-go shipped (embedded dictionary)",
	}, true, nil
}

// extractCopyrightLine returns the first line of text that contains the
// word "Copyright", trimmed of Markdown emphasis markers.
func extractCopyrightLine(text string) (string, bool) {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "Copyright") {
			return strings.Trim(strings.TrimSpace(line), "*"), true
		}
	}
	return "", false
}

// RenderAssets formats AssetRows as a second Markdown table, appended
// after the module-dependency table (AC25).
func RenderAssets(rows []AssetRow) string {
	var b strings.Builder
	b.WriteString("\n## Redistributed non-code assets\n\n")
	b.WriteString("AD-26, verbatim: \"Redistributed non-code assets keep their own terms and\n")
	b.WriteString("their notices.\" Every committed font binary — fixture or shipped — appears\n")
	b.WriteString("below with the licence and copyright line its accompanying LICENSE*/NOTICE*\n")
	b.WriteString("files carry (AC25, D-1.5.6). A font binary committed without both files is a\n")
	b.WriteString("build failure (`ResolveAssets`), not a silent gap.\n\n")
	if len(rows) == 0 {
		b.WriteString("_No redistributed non-code assets are committed at this commit._\n")
		return b.String()
	}
	b.WriteString("| Path | Licence | Copyright | Serves |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", r.Path, r.Licence, r.Copyright, r.Serves)
	}
	return b.String()
}

// Render formats rows as the committed Markdown manifest (AC19: "a
// committed, reviewable file"). The header states explicitly that a
// checker covering its own dependency is correct, not a bootstrap
// problem (AC20, D-1.3.6, D-1.3.9), so nobody reading it mistakes
// lint's own row for a mistake.
func Render(rows []Row) string {
	var b strings.Builder
	b.WriteString("# Third-party licence manifest\n\n")
	b.WriteString("Generated by `lint/internal/manifest` (AC19, AD-26). Every module in the\n")
	b.WriteString("resolved dependency graph of each of this repository's three Go modules —\n")
	b.WriteString("`folio-go`, `hashmatrix` and `lint` — appears below with its resolved\n")
	b.WriteString("licence, which module it serves, and whether that module ships to\n")
	b.WriteString("consumers or exists only to build/test the repo. An unresolved or\n")
	b.WriteString("unrecognised licence fails the build (AC19) rather than appearing here\n")
	b.WriteString("silently; a forbidden licence (GPL, LGPL, AGPL, SSPL, or a commercial\n")
	b.WriteString("EULA) fails the build uniformly across all three modules, with no\n")
	b.WriteString("build-time-only relaxation (AC18, D-1.3.9).\n\n")
	b.WriteString("**`lint`'s own dependency, if any, is covered by this same table below.\n")
	b.WriteString("A checker covering its own dependency is correct, not a bootstrap\n")
	b.WriteString("problem** (AC20, D-1.3.6, D-1.3.9) — it must not be \"fixed\" by excluding\n")
	b.WriteString("lint's own module: AD-26 binds all, and a licence manifest that omits the\n")
	b.WriteString("one module doing the checking is exactly the self-exemption this rule\n")
	b.WriteString("exists to prevent.\n\n")
	if len(rows) == 0 {
		b.WriteString("_No external dependencies exist in any of the three module graphs at\n")
		b.WriteString("this commit (F-8). This table is not empty theatre: it is complete for\n")
		b.WriteString("what exists, and AC19 makes an unresolvable licence a build failure so it\n")
		b.WriteString("stays complete as the graph grows._\n")
		return b.String()
	}
	b.WriteString("| Module | Version | Licence | Serves | Shipped / build-time-only |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", r.Module, r.Version, r.Licence, r.Serves, r.ShippedBy)
	}
	return b.String()
}
