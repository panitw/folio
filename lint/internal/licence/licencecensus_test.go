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

// TestLicenceSignalCensus is Story 8.4i's task 1: D-8.4i.1's REPORT-ONLY
// population census, and D-8.4i.6's hard constraint that the gate must
// not become fatal in the same commit that first measures the
// population.
//
// It runs the collect-all-signals rule (licencesignals.go) over the
// WHOLE existing licence population — every committed LICENSE* file in
// the repository, plus every dependency LICENSE file the three Go module
// graphs resolve to — and RECORDS what it would say beside what the
// shipped classifier says today. It asserts nothing about the
// difference: in the commit that introduces it, no verdict changes and
// no refusal becomes fatal.
//
// It is kept after task 2 flips ClassifyLicenceText onto the new rule,
// where it becomes a standing witness that the two agree on the
// committed population — and, more usefully, a located record of what
// every licence file in the repository classifies as.
//
// WHY THE CENSUS IS THE FIRST TASK AND NOT A NOTE. ClassifyLicenceText
// is shared between the ASSET path (manifest.go, 12 files, visible) and
// the DEPENDENCY path (licencegraph.go, 9 files, not visible in this
// repository at all). If the new rule reds something legitimate, the
// answer is to HALT and route it to the engineering lead — never to
// weaken the rule, narrow its population or exempt a file. Moving the
// bar to fit the instrument (D-8.5.10) is the failure this whole thread
// exists to stop.
func TestLicenceSignalCensus(t *testing.T) {
	root := repoRootForCensus(t)

	type verdict struct {
		where             string
		oldFamily, newFam Family
		oldID, newID      string
		differs           bool
	}
	var verdicts []verdict
	record := func(where, text string) {
		oldFamily, oldID := ClassifyLicenceText(text)
		newFam, newID := classifyByAllSignals(text)
		verdicts = append(verdicts, verdict{
			where: where, oldFamily: oldFamily, newFam: newFam,
			oldID: oldID, newID: newID,
			differs: oldFamily != newFam || oldID != newID,
		})
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
	changed := 0
	for _, v := range verdicts {
		marker := "   "
		if v.differs {
			marker = ">> "
			changed++
		}
		t.Logf("%s%-72s shipped=(%v,%q) collect-all=(%v,%q)", marker, v.where, v.oldFamily, v.oldID, v.newFam, v.newID)
	}
	t.Logf("CENSUS: %d licence texts measured (%d committed files + %d dependency licences); %d change verdict",
		len(verdicts), len(committed), deps, changed)

	// REPORT-ONLY. Nothing here fails on `changed`. If a verdict does
	// change on a LEGITIMATE file, that is a HALT and a finding for the
	// engineering lead — not a reason to soften the rule.
	if changed != 0 {
		t.Logf("NOTE: %d verdict(s) differ. Per D-8.4i.1 this is a finding to route, never a reason to "+
			"weaken the rule, narrow the population, or exempt a file.", changed)
	}
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
