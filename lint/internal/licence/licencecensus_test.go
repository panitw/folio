package licence

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// censusVerdict is one pinned population verdict: a licence text this
// repository relies on, and the (family, id) the classifier MUST give
// it.
type censusVerdict struct {
	where  string
	family Family
	id     string
}

// pinnedCensus is the WHOLE licence population of this repository with
// its verdict written down — 48 committed LICENSE*/COPYING files plus
// the 9 dependency licences the three Go module graphs resolve to.
//
// IT GREW BY 22 AT STORY 8.5: the catalogue's 21 faces, and the compound
// fixture D-8.4j.2 requires. That is the largest single addition this
// table has taken, and every row is written out by hand for the reason
// the paragraph below gives — a new redistributed licence is exactly the
// event that must not land unrecorded.
//
// THIS TABLE IS THE OBSERVATION. It replaces a differential that stopped
// observing anything the moment ClassifyLicenceText's body became a call
// to classifyByAllSignals: from that commit the census compared a
// function to ITSELF, so its "0 of 35 changed verdict" line was true by
// construction and could not have reported otherwise. MEASURED: with the
// SIL OPEN FONT LICENSE entry's requiredVersion set to "VERSION 9.9" —
// which reclassifies all eleven committed OFL files to (unknown, "") and
// fails the shipped asset gate outright — the differential census still
// printed "0 change verdict" and PASSED. A guard whose comment claims a
// safety it does not have is this story's own subject (D-8.0.1), and it
// reached instance four here, in the test written to catch instance
// three.
//
// The table is TEST-OWNED and written out longhand, on D-8.4i.3's
// reasoning about the allowlist: an expectation DERIVED from the thing
// it checks passes any edit to that thing. Adding a dependency or a
// committed licence file therefore requires recording its verdict here.
// That is not friction to be engineered away — under AD-26 a new
// redistributed licence is exactly the event that must not land
// unrecorded.
var pinnedCensus = []censusVerdict{
	{"LICENSE", FamilyPermissive, "MIT"},
	// STORY 8.5'S CATALOGUE, 21 NEW COMMITTED LICENCE TEXTS, PINNED ONE BY
	// ONE. They are the first faces this repository has ever put through the
	// fail-closed asset gate in bulk, and two of them —
	// ubuntusans/ and ubuntusansmono/ — are the first Ubuntu-font-1.0 ASSET
	// this repository has ever carried, closing the gap classify.go:167-171
	// records ("there is no analogue of TestCommittedOFLTextClassifiesAsOFL11
	// to write until Story 8.5 lands a face under it"). Written out longhand,
	// on D-8.4i.3's reasoning: an expectation derived from the thing it
	// checks passes any edit to that thing.
	{"folio-designer/public/fonts/cascadiacode/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/cascadiamono/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/cousine/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/firacode/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/geist/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/geistmono/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/ibmplexmono/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/ibmplexsans/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/ibmplexsansthai/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/intelonemono/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/inter/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/interdisplay/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/jetbrainsmono/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/literata/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/notosans/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/notosanssc/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/notosansthai/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/notosansthailooped/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/notoserif/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/notoserifthai/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/roboto/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/sourcecodepro/LICENSE-OFL.md", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/sourcesans3/LICENSE-OFL.md", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/sourceserif4/LICENSE-OFL.md", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/spacegrotesk/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-designer/public/fonts/ubuntusans/LICENSE-UFL.txt", FamilyPermissive, "Ubuntu-font-1.0"},
	{"folio-designer/public/fonts/ubuntusansmono/LICENSE-UFL.txt", FamilyPermissive, "Ubuntu-font-1.0"},
	{"folio-designer/third-party-notices/pdfjs-dist/LICENSE-APACHE-2.0", FamilyPermissive, "Apache-2.0"},
	{"folio-designer/third-party-notices/pdfjs-dist/LICENSE-CMAPS", FamilyPermissive, "BSD-3-Clause"},
	{"folio-designer/third-party-notices/pdfjs-dist/LICENSE-LIBERATION", FamilyPermissive, "OFL-1.1"},
	{"folio-go/fonts/notosans/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-go/fonts/notosanssc/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-go/fonts/notosansthai/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-go/internal/text/wordlist/LICENSE-CC0-1.0.txt", FamilyPermissive, "CC0-1.0"},
	{"folio-go/testdata/fonts/LICENSE-Roboto.txt", FamilyPermissive, "Apache-2.0"},
	{"folio-go/testdata/fonts/notosansthai-variable-testonly/LICENSE-OFL.txt", FamilyPermissive, "OFL-1.1"},
	{"folio-go/testdata/lint/wordlist-assets/compliant/folio-go/internal/text/wordlist/LICENSE-CC0-1.0.txt", FamilyUnknown, ""},
	{"folio-go/testdata/lint/wordlist-assets/violating/folio-go/internal/text/wordlist/LICENSE-CC0-1.0.txt", FamilyUnknown, ""},
	{"lint/testdata/licence/copyleft/example.test/agpl-lib/LICENSE", FamilyCopyleft, "AGPL-3.0-only"},
	{"lint/testdata/licence/copyleft/example.test/gpl-lib/LICENSE", FamilyCopyleft, "GPL-3.0-only"},
	{"lint/testdata/licence/copyleft/example.test/lgpl-lib/LICENSE", FamilyCopyleft, "LGPL-3.0-only"},
	{"lint/testdata/licence/copyleft/example.test/sspl-lib/LICENSE", FamilyCopyleft, "SSPL-1.0"},
	{"lint/testdata/licence/permissive/example.test/apache-lib/LICENSE", FamilyPermissive, "Apache-2.0"},
	{"lint/testdata/licence/permissive/example.test/bsd-lib/LICENSE", FamilyPermissive, "BSD-3-Clause"},
	// D-8.4j.2'S COMPOUND CASE, seeded as a fixture because no procurable
	// font carries one the four-id allowlist admits (Design Note 3, measured
	// before the decision per D-000.12: Hack is MIT + Bitstream Vera and
	// Public Sans is OFL-1.1 + CC0-1.0, and BOTH must fail the gate). The id
	// pinned here is the WHOLE EXPRESSION, which is what Story 8.4j's
	// whole-line capture returns and what per-term admission then tests; a
	// regression to the one-token capture reports "MIT" and reds this row.
	{"lint/testdata/licence/permissive/example.test/compound-lib/LICENSE", FamilyPermissive, "MIT OR Apache-2.0"},
	{"lint/testdata/licence/permissive/example.test/mit-lib/LICENSE", FamilyPermissive, "MIT"},
	{"lint/testdata/licence/permissive/example.test/ufl-lib/LICENSE", FamilyPermissive, "Ubuntu-font-1.0"},
	{"dep folio-go -> github.com/boxesandglue/textshape", FamilyPermissive, "MIT"},
	{"dep lint -> github.com/google/go-cmp", FamilyPermissive, "BSD-3-Clause"},
	{"dep lint -> github.com/yuin/goldmark", FamilyPermissive, "MIT"},
	{"dep lint -> golang.org/x/mod", FamilyPermissive, "BSD-3-Clause"},
	{"dep lint -> golang.org/x/net", FamilyPermissive, "BSD-3-Clause"},
	{"dep lint -> golang.org/x/sync", FamilyPermissive, "BSD-3-Clause"},
	{"dep lint -> golang.org/x/sys", FamilyPermissive, "BSD-3-Clause"},
	{"dep lint -> golang.org/x/telemetry", FamilyPermissive, "BSD-3-Clause"},
	{"dep lint -> golang.org/x/tools", FamilyPermissive, "BSD-3-Clause"},
}

// TestLicenceSignalCensus is Story 8.4i's task 1 — D-8.4i.1's
// population census, and D-8.4i.6's hard constraint that the gate must
// not become fatal in the same commit that first measures the
// population.
//
// AS COMMITTED AT 2e9365e it was report-only and DIFFERENTIAL: the
// collect-all-signals rule ran beside the then-shipped first-match
// switch over the whole population, and 0 of 35 texts changed verdict.
// That measurement is a fact about THAT COMMIT, where the two
// classifiers were genuinely different code, and it is what licensed
// the next commit to make refusals fatal. It is recorded in the story's
// Delivery Log and in DW-117's closing note, attributed there.
//
// AS IT STANDS NOW the differential is meaningless — both columns would
// be the same function — so the test PINS the population's verdicts
// against the table above instead. What it asserts today: every licence
// text this repository redistributes or depends on classifies exactly as
// recorded, no population member is missing from the table, and no table
// entry has silently lost its file.
//
// WHY THE CENSUS EXISTS AT ALL. ClassifyLicenceText is shared between
// the ASSET path (manifest.go, 12 files, visible) and the DEPENDENCY
// path (licencegraph.go, 9 files, not visible in this repository at
// all). If a change reds something legitimate here, the answer is to
// HALT and route it to the engineering lead — never to weaken the rule,
// narrow the population or exempt a file. Moving the bar to fit the
// instrument (D-8.5.10) is the failure this whole thread exists to stop.
func TestLicenceSignalCensus(t *testing.T) {
	root := repoRootForCensus(t)

	want := make(map[string]censusVerdict, len(pinnedCensus))
	for _, v := range pinnedCensus {
		if _, dup := want[v.where]; dup {
			t.Fatalf("pinnedCensus lists %q twice", v.where)
		}
		want[v.where] = v
	}
	seen := map[string]bool{}
	measured := 0

	record := func(where, text string) {
		measured++
		family, id := ClassifyLicenceText(text)
		seen[where] = true
		pin, ok := want[where]
		if !ok {
			t.Errorf("%s: classifies as (%v, %q) but is NOT in pinnedCensus. A licence text this "+
				"repository redistributes or depends on must have its verdict RECORDED (AD-26): add it "+
				"to the table — do not delete it from the population.", where, family, id)
			return
		}
		if family != pin.family || id != pin.id {
			t.Errorf("%s: classifies as (%v, %q), pinned as (%v, %q). If the new verdict is CORRECT, the "+
				"table is the place that records the change and says who decided it. If it is not, this is "+
				"a HALT and a finding for the engineering lead — never a reason to weaken the rule, narrow "+
				"the population or exempt a file (D-8.5.10).",
				where, family, id, pin.family, pin.id)
			return
		}
		t.Logf("    %-72s (%v, %q)", where, family, id)
	}

	// --- population A: every committed LICENSE*/COPYING* file ---
	//
	// Deliberately a SUPERSET of the 12 asset files the font/wordlist
	// gate classifies and the 8 lint fixtures: a census that enumerated
	// only the files it expected to find would not be a census. The
	// third-party notice texts and the repository's own LICENSE are
	// carried too, at no cost.
	committed := committedLicenceFiles(t, root)
	if len(committed) < 20 {
		t.Fatalf("census found only %d committed licence files — expected the 12 asset files, the 8 lint "+
			"fixtures and more besides; this measurement would be vacuous", len(committed))
	}
	for _, rel := range committed {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read committed licence file %s: %v", rel, err)
		}
		record(rel, string(data))
	}

	// --- population B: every dependency LICENSE the Go module graphs
	// resolve to ---
	//
	// This is the half nobody can see by reading the repository, and the
	// half D-8.4i.1's Block If was written about. hashmatrix resolves
	// zero non-main modules; its scan is vacuous, which is itself worth
	// recording rather than assuming.
	deps := 0
	for _, mod := range []string{"folio-go", "lint", "hashmatrix"} {
		modules, err := ResolveGraph(filepath.Join(root, mod))
		if err != nil {
			t.Fatalf("resolve %s module graph: %v", mod, err)
		}
		t.Logf("module graph %-10s resolves %d non-main module(s)", mod, len(modules))
		for _, m := range modules {
			text, ok := ReadLicenceText(m.Dir)
			if !ok {
				record("dep "+mod+" -> "+m.Path+" (NO LICENCE FILE)", "")
				continue
			}
			deps++
			record("dep "+mod+" -> "+m.Path, text)
		}
	}
	if deps == 0 {
		t.Fatal("census resolved no dependency licence text at all — the population this rule is least " +
			"able to inspect by eye is exactly the one that must not go unmeasured (D-8.4i.1)")
	}

	// --- the record ---
	for _, v := range pinnedCensus {
		if !seen[v.where] {
			t.Errorf("%s is pinned in pinnedCensus but was NOT found in the population. A licence file "+
				"disappearing silently is how a census stops being a census — if it was removed "+
				"deliberately, remove its pin in the same change.", v.where)
		}
	}
	t.Logf("CENSUS: %d licence texts measured (%d committed files + %d dependency licences), all matching "+
		"their pinned verdicts", measured, len(committed), deps)
}

// committedLicenceFiles lists every tracked file whose basename names a
// licence text. Tracked, not walked: the population under audit is what
// the repository redistributes.
func committedLicenceFiles(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files in %s: %v", root, err)
	}
	licenceBasename := regexp.MustCompile(`^(LICENSE|LICENCE|COPYING)`)
	var rels []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		base := path.Base(line)
		if !licenceBasename.MatchString(base) {
			continue
		}
		if strings.HasSuffix(base, ".go") {
			continue
		}
		rels = append(rels, line)
	}
	return rels
}

// repoRootForCensus walks up from this package's directory to the
// repository root, identified by the .git directory rather than by a
// hard-coded number of parent hops.
func repoRootForCensus(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, ".git")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no repository root above %s", dir)
		}
		dir = parent
	}
}
