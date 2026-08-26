package folio_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	folio "github.com/panitw/folio/folio-go"
)

// Story 2.3a, AC6 and AC7. This file asserts the two claims the story
// makes about what it did NOT change, in the form D-000.21 (sharpened)
// requires: assert on the artifact that carries the property, and PROVE
// it carries it.
//
// The claims are separable, and both are needed:
//
//   - AC6, byte-neutrality. Story 2.3a swapped a fractional vendor
//     accessor for the integer the same table carries. Measured over all
//     37,312 glyphs of the four committed faces, the two agree with 0
//     mismatches (internal/fontset/vendor-boundary.md, Table 4), so no
//     width, no /W array, no content stream and no golden moves.
//   - AC7, no third Epic 2 gate obligation. The gate owed exactly two
//     things at 431a6a5 — the four-target matrix legs, and D-2.3.5's
//     Thai sign-off — and owes exactly those two now.
//
// WHERE THE RENDERING HALF LIVES. Each fixture's own test already
// renders it and compares the produced bytes against its committed
// digest (fixture_test.go for minimal-rect, font-text, image-embed and
// multi-script-fallback; shaped_fixture_test.go for shaped-text). This
// file does NOT duplicate that work. What it adds is the half those
// tests cannot provide on their own: that the digests they compare
// against are still the ones this story inherited. A re-recording that
// moved a golden AND its expected.json together would leave every one of
// those tests green.

// goldenDigestRecord IS THE LIST (D-000.47). It declares, ONCE and
// declaratively, every fixture's committed sha256 AND every site in the
// repository that records it — and the guard below READS the list rather
// than re-stating it.
//
// WHY THIS SHAPE, measured rather than assumed. Story 2.5a grepped each
// committed digest across the whole repository and found a golden's
// digest living at FOUR kinds of site:
//
//	the artifact       fixtures/<f>/expected.pdf     — the bytes themselves
//	the normative hash fixtures/<f>/expected.json    — what matrixDocuments reads
//	the second literal this file's declaration       — deliberately independent
//	the documentation  fixtures/<f>/README.md prose  — two fixtures quote it
//
// FOUR SITES FOUND BY MEASUREMENT ARE FOUR SITES A FUTURE RE-RECORD CAN
// MISS, and the previous shape of this file knew about exactly one of
// them. So the completeness half below does not trust this table either:
// it scans the declared search scope for each digest VALUE and asserts
// the set of files carrying it equals the declared set exactly. Adding a
// README that quotes a digest without declaring it here is then a
// failure rather than a silent fifth site.
//
// This is the same derive-from-a-declarative-spec move that fixed the
// per-face assertion set (D-000.23) and the gate-obligation count
// (D-2.5.1).
//
// THE SECOND LITERAL IS STILL A SECOND LITERAL (D-2.3.4). The sha256
// below is written out here rather than read from the file it checks.
// That is deliberate and it is the whole mechanism: the cheapest
// available response to a hash mismatch is to re-record, and a
// re-recording is invisible in a diff when only one copy of the number
// exists. With two, a re-record must change this file too — and that is
// exactly the conversation the second literal is for.
//
// fixtures/shaped-text's digest is the load-bearing one. D-2.3.5's Thai
// semantic READING sign-off will be bound to THESE EXACT BYTES: the
// sign-off record must name this digest. It has not been given yet, and
// Story 2.5a moved these bytes precisely so that it is not asked for and
// then invalidated (D-000.41).
type goldenDigestSite struct {
	// kind is one of "expected.json", "second-literal", "readme",
	// "signoff". The artifact itself is not listed: it is what gets
	// hashed.
	kind string
	// relPath is repo-root-relative, or "" for "second-literal", which
	// is this very declaration.
	relPath string
}

var goldenDigestRecord = []struct {
	dir    string
	sha256 string
	sites  []goldenDigestSite
}{
	{
		dir:    "minimal-rect",
		sha256: "0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/minimal-rect/expected.json"},
			{kind: "second-literal"},
		},
	},
	{
		// Re-recorded by Story 2.5a (DW-15 + D-2.4.2 amended). Its
		// baselines moved UP: Roboto's hhea ascent is 928, BELOW the
		// 1000-unit em, the only fixture whose baselines move that way.
		dir:    "font-text",
		sha256: "a69a665331e7f0d31619f48179b54c7b9cb7a90ae013ed9c7c79daa128612181",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/font-text/expected.json"},
			{kind: "second-literal"},
		},
	},
	{
		dir:    "image-embed",
		sha256: "e5778eb872c98ec4a3c3c89466a8313cf52931b896701de8a43f3506abe689fc",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/image-embed/expected.json"},
			{kind: "second-literal"},
		},
	},
	{
		// Re-recorded by Story 2.5a.
		dir:    "multi-script-fallback",
		sha256: "4699c8d710724ea544cc26bb3ee2b96af7a333f3dddd4462c0c846f7790480b0",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/multi-script-fallback/expected.json"},
			{kind: "second-literal"},
		},
	},
	{
		// Re-recorded by Story 2.5a. THE SIGN-OFF-BINDING FIXTURE.
		//
		// fixtures/shaped-text/thai-signoff.json was ADDED as a site here
		// on the Epic 2 gate correction that followed the owner's
		// sign-off (D-000.47): the record names this exact digest so a
		// future re-record invalidates it by construction
		// (shaped_signoff_matrix_test.go's assertSignOffMatchesFrozenHash
		// enforces the SAME binding under -tags=matrix; this entry is
		// what makes the completeness half of THIS guard, which runs in
		// the ordinary suite, aware the site exists at all).
		dir:    "shaped-text",
		sha256: "6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/shaped-text/expected.json"},
			{kind: "second-literal"},
			{kind: "signoff", relPath: "fixtures/shaped-text/thai-signoff.json"},
		},
	},
	{
		// Re-recorded by Story 2.5a. Quotes its own digest in its README.
		dir:    "three-band-page",
		sha256: "746efcbcfb5be30a06caaaefae25e3eaba1962c3fa47a74da10af6d0885372bf",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/three-band-page/expected.json"},
			{kind: "second-literal"},
			{kind: "readme", relPath: "fixtures/three-band-page/README.md"},
		},
	},
	{
		// RECORDED by Story 2.6 — the FIRST multi-page document in the
		// repository, and the only artifact on which the ruled pagination
		// model is observable at all. Its 29 content lines partition 22/7
		// across two pages; every pre-2.6 fixture is single-page, so none
		// of them can express a pagination defect (D-000.50).
		dir:    "multi-page",
		sha256: "66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/multi-page/expected.json"},
			{kind: "second-literal"},
			{kind: "readme", relPath: "fixtures/multi-page/README.md"},
		},
	},
	{
		// RECORDED by Story 2.7 — the first document carrying a resolved
		// {{page}}/{{pages}} construct, and the only artifact spanning
		// the page-9-to-page-10 digit-count boundary (finding 8, story
		// creation: no other fixture reaches 10 pages).
		dir:    "page-count-20",
		sha256: "b32fa1c5babb8327b09b5c2bc0a11628b8c8885b9c5661c0262ec24920c5150f",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/page-count-20/expected.json"},
			{kind: "second-literal"},
			{kind: "readme", relPath: "fixtures/page-count-20/README.md"},
		},
	},
	{
		// Re-recorded by Story 2.5a. THE ONLY FIXTURE WITH MULTI-LINE
		// ELEMENTS, so the only artifact in the repository on which
		// D-2.4.2's AMENDED advance (1511 -> 1610 units on the Noto x3
		// chain) is observable at all. Its README quotes the digest on
		// all four matrix-target lines.
		dir:    "wrapped-text",
		sha256: "07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6",
		sites: []goldenDigestSite{
			{kind: "expected.json", relPath: "fixtures/wrapped-text/expected.json"},
			{kind: "second-literal"},
			{kind: "readme", relPath: "fixtures/wrapped-text/README.md"},
		},
	},
}

// goldenDigestSearchScope is where the completeness half looks for a
// digest occurrence. Declared rather than "everywhere" for one reason,
// stated so it is not mistaken for laziness: _bmad-output/ story files
// record digests as PAST-TENSE MEASUREMENTS ("recorded at 17f5f7a the
// digest was X"). Those are history and must NOT be rewritten when a
// golden moves — rewriting them would destroy the record of what
// actually produced which bytes, which is the thing AD-21 exists to
// keep. So the scope is the live artifacts and this file.
//
// NAMED EXCEPTION (Story 2.6 finisher, Finding 11): boundaryGateDocuments
// — e.g. epic-2-boundary-gate.md — also quote a golden's digest, and
// theirs is a LIVE gate binding, not history, yet they sit under
// _bmad-output/ and are OUT of this scope like every story file. That is
// deliberate (the gate doc is append-only and dated, so a re-record
// appends a new line rather than staling an old one, and re-scoping to it
// would drag every story file back in), but it means this test's own log
// line — "the search scope carries no undeclared occurrence" — is true of
// the SCOPE, not of the repository. boundaryGateDocuments' digests are
// covered separately, by SHAPE rather than by value, in
// TestBoundaryGateDigestsAreWellFormed below.
var goldenDigestSearchScope = []string{"fixtures", filepath.Join("folio-go", "byte_neutrality_test.go")}

// TestGoldenDigestAgreesAtEveryDeclaredSite is the digest guard, in the
// form D-000.47 requires: the list is declared once, above, and this
// reads it.
//
// It asserts three things, and the third is the one the previous shape
// of this file could not:
//
//  1. The ARTIFACT hashes to the declared digest. (The previous version
//     never re-hashed anything — it compared expected.json to a literal,
//     so a re-record that moved the PDF and its JSON together would have
//     left it green.)
//  2. Every DECLARED site records that digest.
//  3. NO UNDECLARED site records it. This is the completeness half, and
//     it is what makes "four sites" a checked number rather than a
//     comment.
//
// Presence comes first at every level, because "the digest matches" is
// trivially true of a fixture that is not there.
func TestGoldenDigestAgreesAtEveryDeclaredSite(t *testing.T) {
	root := repoRootForByteNeutrality(t)

	if len(goldenDigestRecord) == 0 {
		t.Fatal("vacuity guard: the digest declaration is empty, so this test asserts nothing")
	}

	// VACUITY GUARD ON THE SCOPE ITSELF. If the search scope does not
	// exist, the completeness half below finds zero occurrences of
	// everything and certifies that nothing is undeclared — the exact
	// shape of a guard that passes because it looked nowhere.
	for _, rel := range goldenDigestSearchScope {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("presence precondition: the declared search scope %s does not exist (%v), so the completeness assertion below would pass by looking nowhere", rel, err)
		}
	}

	checkedArtifacts, checkedSites := 0, 0
	for _, fx := range goldenDigestRecord {
		dir := filepath.Join(root, "fixtures", fx.dir)
		info, derr := os.Stat(dir)
		if derr != nil || !info.IsDir() {
			t.Errorf("presence precondition: fixture directory %s is missing (%v) — a digest claim about a fixture that is not there is vacuous", dir, derr)
			continue
		}
		if len(fx.sites) == 0 {
			t.Errorf("%s declares no recording sites, so nothing about it is checked", fx.dir)
			continue
		}
		if len(fx.sha256) != 64 || strings.ToLower(fx.sha256) != fx.sha256 {
			t.Errorf("%s's declared digest %q is not 64 lowercase hex characters", fx.dir, fx.sha256)
			continue
		}

		// (1) THE ARTIFACT ITSELF.
		pdfPath := filepath.Join(dir, "expected.pdf")
		pdf, perr := os.ReadFile(pdfPath)
		if perr != nil {
			t.Errorf("presence precondition: %s could not be read: %v", pdfPath, perr)
			continue
		}
		if len(pdf) == 0 {
			t.Errorf("presence precondition: %s is empty — two empty files are byte-identical (Story 1.1)", pdfPath)
			continue
		}
		sum := sha256.Sum256(pdf)
		if got := hex.EncodeToString(sum[:]); got != fx.sha256 {
			t.Errorf("fixtures/%s/expected.pdf hashes to %s, but the declared digest is %s.\n%s", fx.dir, got, fx.sha256, goldenDigestRemedy)
		}
		checkedArtifacts++

		// (2) EVERY DECLARED SITE.
		for _, site := range fx.sites {
			switch site.kind {
			case "second-literal":
				// This declaration IS the site. Nothing to read.
				checkedSites++
			case "expected.json":
				path := filepath.Join(root, site.relPath)
				body, rerr := os.ReadFile(path)
				if rerr != nil {
					t.Errorf("presence precondition: %s could not be read: %v", path, rerr)
					continue
				}
				if len(body) == 0 {
					t.Errorf("presence precondition: %s is empty", path)
					continue
				}
				var raw map[string]any
				if jerr := json.Unmarshal(body, &raw); jerr != nil {
					t.Errorf("presence precondition: %s is not valid JSON: %v", path, jerr)
					continue
				}
				field, present := raw["sha256"]
				if !present {
					t.Errorf("presence precondition: %s carries no \"sha256\" field — the property this test asserts does not live in this artifact", path)
					continue
				}
				got, isString := field.(string)
				if !isString {
					t.Errorf("presence precondition: %s's \"sha256\" is %T, not a JSON string — a widened per-target object would make a single-value comparison meaningless", path, field)
					continue
				}
				if len(got) != 64 || strings.ToLower(got) != got {
					t.Errorf("presence precondition: %s's \"sha256\" is %q, which is not 64 lowercase hex characters", path, got)
					continue
				}
				if got != fx.sha256 {
					t.Errorf("%s's digest is %s, but the second literal declares %s.\n%s", site.relPath, got, fx.sha256, goldenDigestRemedy)
				}
				checkedSites++
			case "readme":
				path := filepath.Join(root, site.relPath)
				body, rerr := os.ReadFile(path)
				if rerr != nil {
					t.Errorf("presence precondition: %s could not be read: %v", path, rerr)
					continue
				}
				if !strings.Contains(string(body), fx.sha256) {
					t.Errorf("%s does not quote fixtures/%s's digest %s. A fixture's own documentation stating a digest the fixture no longer has is a silently lying artifact.\n%s", site.relPath, fx.dir, fx.sha256, goldenDigestRemedy)
				}
				checkedSites++
			case "signoff":
				// A human sign-off record — e.g.
				// fixtures/shaped-text/thai-signoff.json — names, in its
				// own "sha256" field, the exact digest the reader looked
				// at. Read the same way as "expected.json" above (both
				// are flat JSON objects with a top-level "sha256"
				// string), but reported as what it is: a human record
				// going stale, not a machine-derived fixture drifting.
				path := filepath.Join(root, site.relPath)
				body, rerr := os.ReadFile(path)
				if rerr != nil {
					t.Errorf("presence precondition: %s could not be read: %v", path, rerr)
					continue
				}
				if len(body) == 0 {
					t.Errorf("presence precondition: %s is empty", path)
					continue
				}
				var raw map[string]any
				if jerr := json.Unmarshal(body, &raw); jerr != nil {
					t.Errorf("presence precondition: %s is not valid JSON: %v", path, jerr)
					continue
				}
				field, present := raw["sha256"]
				if !present {
					t.Errorf("presence precondition: %s carries no \"sha256\" field — the property this test asserts does not live in this artifact", path)
					continue
				}
				got, isString := field.(string)
				if !isString {
					t.Errorf("presence precondition: %s's \"sha256\" is %T, not a JSON string", path, field)
					continue
				}
				if len(got) != 64 || strings.ToLower(got) != got {
					t.Errorf("presence precondition: %s's \"sha256\" is %q, which is not 64 lowercase hex characters", path, got)
					continue
				}
				if got != fx.sha256 {
					t.Errorf("%s names digest %s, but fixtures/%s's declared digest is %s. The sign-off is STALE: the fixture moved since it was signed off, and D-2.3.5's anti-rot condition says that must invalidate the record, not survive it.", site.relPath, got, fx.dir, fx.sha256)
				}
				checkedSites++
			default:
				t.Errorf("%s declares a site of unknown kind %q", fx.dir, site.kind)
			}
		}
	}

	if checkedArtifacts == 0 || checkedSites == 0 {
		t.Fatalf("vacuity: %d artifacts and %d recording sites were checked", checkedArtifacts, checkedSites)
	}

	// (3) THE COMPLETENESS HALF. For each declared digest, the set of
	// files inside the search scope that CONTAIN it must equal the set
	// declared above. This is what turns "a digest lives at four sites"
	// from a comment into a checked property: a fifth site added later
	// fails here instead of silently going stale at the next re-record.
	occurrences := map[string][]string{}
	for _, rel := range goldenDigestSearchScope {
		base := filepath.Join(root, rel)
		werr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			// expected.pdf is the artifact, not a recording OF the
			// digest; and it cannot contain its own hash.
			if filepath.Base(path) == "expected.pdf" {
				return nil
			}
			body, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			text := string(body)
			for _, fx := range goldenDigestRecord {
				if strings.Contains(text, fx.sha256) {
					relp, _ := filepath.Rel(root, path)
					occurrences[fx.sha256] = append(occurrences[fx.sha256], filepath.ToSlash(relp))
				}
			}
			return nil
		})
		if werr != nil {
			t.Fatalf("walking the declared search scope %s: %v", rel, werr)
		}
	}

	for _, fx := range goldenDigestRecord {
		declared := map[string]bool{}
		for _, site := range fx.sites {
			switch site.kind {
			case "second-literal":
				declared[filepath.ToSlash(filepath.Join("folio-go", "byte_neutrality_test.go"))] = true
			default:
				declared[filepath.ToSlash(site.relPath)] = true
			}
		}
		observed := map[string]bool{}
		for _, p := range occurrences[fx.sha256] {
			observed[p] = true
		}
		for p := range observed {
			if !declared[p] {
				t.Errorf("%s records fixtures/%s's digest %s, but goldenDigestRecord does not declare it as a site. Add it to the declaration — an undeclared site is one a future re-record will miss (D-000.47).", p, fx.dir, fx.sha256)
			}
		}
		for p := range declared {
			if !observed[p] {
				t.Errorf("goldenDigestRecord declares %s as a site recording fixtures/%s's digest %s, but that file does not contain it.", p, fx.dir, fx.sha256)
			}
		}
	}

	t.Logf("D-000.47: %d artifacts re-hashed; %d declared recording sites agree; the search scope %v carries no undeclared occurrence.", checkedArtifacts, checkedSites, goldenDigestSearchScope)
}

// goldenDigestRemedy is the failure message's advice, written once
// because every branch above needs it and a remedy that drifts between
// branches is worse than one that is merely wrong.
//
// D-000.37, AND WHY THIS TEXT CHANGED IN STORY 2.5a'S OWN COMMIT. What
// stood here said, verbatim: "Do not update this literal to make the
// test pass — that is the move AD-21 and D-000.22 exist to prevent."
// That was true when only Story 2.3a's byte-neutrality premise was at
// stake, and Story 2.5a had to update these literals — so a TRUE guard
// was carrying a remedy that FORBADE THE CORRECT ACTION. A tripwire's
// failure message is executed by a human; a stale remedy is worse than
// a stale comment, and the ruling is that it is corrected in the same
// commit as the movement it fails to anticipate.
//
// The prohibition is kept, and it is now stated as the RULE it always
// was rather than as one story's premise: updating a digest is legal
// only as the deliberate, attributable re-recording of a golden, never
// as the way to make a red test go green.
const goldenDigestRemedy = "" +
	"A GOLDEN'S DIGEST MOVED. There are exactly two possibilities and they are not interchangeable.\n" +
	"\n" +
	"  (a) THIS IS A DELIBERATE RE-RECORDING. A story intended to change these bytes, said so, and\n" +
	"      can name ONE cause for the movement. Then every site in goldenDigestRecord is updated\n" +
	"      TOGETHER — expected.pdf, expected.json, this second literal, and any README quoting the\n" +
	"      digest — in that story's own commit, with a semantic acceptance step read off the NEW\n" +
	"      artifact (D-000.22, and D-000.44: a re-recording is a recording, so the step is owed\n" +
	"      again). Two authorised movements have happened so far: Story 2.3a re-derived the subset\n" +
	"      tag, and Story 2.5a corrected the vertical placement model (DW-15 + D-2.4.2 amended),\n" +
	"      moving five goldens as one attributable cause.\n" +
	"\n" +
	"  (b) NOTHING INTENDED TO CHANGE THESE BYTES. Then this is the regression the guard exists to\n" +
	"      catch, and the digest is the evidence. FIND THE CAUSE.\n" +
	"\n" +
	"DO NOT UPDATE A DIGEST TO MAKE A TEST GO GREEN. That is the move AD-21 and D-000.22 exist to\n" +
	"prevent, and it is not made legal by case (a) — under (a) the digest changes because a story\n" +
	"decided to re-record, and the test going green is the consequence, not the reason."

// declaredEpic2GateObligations IS THE LIST. It is the single, explicit
// declaration of everything the Epic 2 boundary gate owes, and the guard
// below asserts the OBSERVED obligation set equals it exactly.
//
// D-2.5.1, which replaced this guard's previous shape:
//
//	"TestStory23aAddedNoThirdEpic2GateObligation encodes a COUNT IN ITS
//	NAME, and that name rots on every future obligation. Replace it with
//	an assertion that the registered obligation set equals an explicit
//	declared list. Adding one then becomes a ONE-LINE DIFF to that list —
//	visible and reviewable — instead of a rename."
//
// The same derive-from-a-declarative-spec move that fixed the per-face
// assertion set in D-000.23's consequent obligation. It retires the
// counting problem PERMANENTLY rather than deferring it one story: the
// old name asserted "no THIRD", then had to mean "no FOURTH" while still
// reading "Third", which is a false statement compiled into the suite.
//
// Two kinds of obligation are observable in the tree, and both are
// listed here so neither can grow unnoticed:
//
//	matrix-file: <path>      a //go:build matrix file — a gate-run test.
//	                         A deliberately-red sign-off record is one of
//	                         these, and so is the matrix harness itself.
//	matrix-document: <slug>  a document registered in matrixDocuments
//	                         whose four legs the gate must run and
//	                         compare.
//
// Each entry names the story and the ruling that authorised it. Adding
// one without a ruling is exactly what this guard exists to stop.
var declaredEpic2GateObligations = []string{
	// The gate-run test files.
	"matrix-file: fontgen_matrix_test.go",                 // Story 2.2 — shipped-face instancing
	"matrix-file: matrix_test.go",                         // Story 1.2 — the four-target legs themselves
	"matrix-file: shaped_signoff_matrix_test.go",          // Story 2.3 — Thai READING sign-off (D-2.3.5)
	"matrix-file: expected_breaks_signoff_matrix_test.go", // Story 2.4 — Thai BREAK sign-off (D-2.4.3)

	// The documents whose four legs the gate runs and compares.
	"matrix-document: minimal-rect",          // Story 1.1
	"matrix-document: font-text",             // Story 1.5
	"matrix-document: image-embed",           // Story 1.8
	"matrix-document: multi-script-fallback", // Story 2.2
	"matrix-document: shaped-text",           // Story 2.3
	"matrix-document: wrapped-text",          // Story 2.4 — legs RUN in-story (D-000.4 override)
	"matrix-document: three-band-page",       // Story 2.5 — legs DEFERRED to the gate (D-2.5.1; D-000.4 override criterion DECLINED)
	"matrix-document: multi-page",            // Story 2.6 — the gate's FIFTH obligation, SANCTIONED by D-2.6.2; legs DEFERRED to the gate (D-000.4 override criterion DECLINED)
	"matrix-document: page-count-20",         // Story 2.7 — the gate's SIXTH obligation, SANCTIONED by D-2.7.4 (D-2.6.2's criterion: FR31 had no cross-target artifact before this entry); legs DEFERRED to the gate (D-000.4 override criterion DECLINED, D-2.7.4)
	"matrix-document: hidden-image",          // Story 3.5 finisher (Finding 1 / Blocker) — D-000.54: native leg (host target) RUN by this story; the other three DEFERRED to the Epic 3 boundary gate (D-000.4 override criterion DECLINED — integer/set work, no new source of cross-target divergence)
}

// TestEpic2GateObligationsMatchTheDeclaredSet asserts, mechanically
// rather than as an intention, that what the Epic 2 boundary gate owes
// is exactly what declaredEpic2GateObligations says it owes — no more,
// and no fewer.
//
// The two Thai sign-offs are distinct human judgments — "does this Thai
// READ correctly" and "do these BREAK POINTS fall correctly" — and each
// is tracked as its own deliberately-red gate-run test
// (TestShapedTextThaiSemanticSignOffIsRecorded,
// TestExpectedBreaksHumanSignOffIsRecorded). Each stayed red until its
// own sign-off file named a reader, a date, what they examined, and the
// digest it certifies — see (a) for what this guard now does about it.
//
// D-000.26 (refined) is why they are two records and not one: a sign-off
// binds to the artifact expressing the property judged. The reading
// judgment binds to fixtures/shaped-text/expected.pdf's digest; the
// break judgment binds to the break-opportunity VECTOR, which a uniform
// baseline shift would leave untouched.
func TestEpic2GateObligationsMatchTheDeclaredSet(t *testing.T) {
	root := repoRootForByteNeutrality(t)

	// (a) THE SIGN-OFF OBLIGATIONS ARE NOW DISCHARGED (Epic 2 gate
	//     correction, following the owner's two sign-offs dated
	//     2026-08-25). This block used to assert the OPPOSITE — that
	//     fixtures/shaped-text/thai-signoff.json did not exist, because
	//     no story is permitted to fabricate a human judgment
	//     (D-000.28: a claim written before the event it asserts is
	//     false from birth). That guard did its job: it stayed green
	//     until a real sign-off landed, then failed the instant one did
	//     — which was CORRECT, not a bug, because at that instant the
	//     guard's premise ("this file must not exist yet") had just
	//     become false. Its job is finished now that the premise it
	//     policed no longer holds, and per the ruling binding this
	//     correction — an absence guard is a placeholder for a presence
	//     guard, and deleting it without landing the successor is how a
	//     deferral is lost — it is REPLACED, not deleted, by a positive
	//     assertion that each record is real and still binds.
	//
	//     "Still binds" is not "is present". A blank template or a
	//     record left over a re-recorded fixture would both satisfy a
	//     bare os.Stat and both are indistinguishable from nobody having
	//     looked. So this checks the same three things
	//     shaped_signoff_matrix_test.go and
	//     expected_breaks_signoff_matrix_test.go check under
	//     -tags=matrix — all four fields populated, the date parses as
	//     ISO-8601, and the record's digest still equals what its
	//     fixture hashes to NOW — in the ordinary suite, so a sign-off
	//     going stale between gate runs is caught continuously rather
	//     than only when the matrix build happens to run
	//     (TestExpectedBreaksVectorIsPinnedInTheOrdinarySuite already
	//     does this same "pin it in the ordinary suite, not only behind
	//     the matrix tag" move for expected_breaks.json's digest itself,
	//     D-2.6.4).
	//
	//     FINDING (D-000.42): the doc comment above this function used
	//     to read "No story creates either file — see (a)". (a) never
	//     covered both files — it only ever stat'd thai-signoff.json.
	//     fixtures/expected-breaks/break-signoff.json was never checked
	//     by this guard in either direction, which is exactly why its
	//     arrival tripped nothing here: the comment overclaimed this
	//     guard's reach for the entire time both sign-offs were
	//     outstanding. Both records are checked below, symmetrically,
	//     now that the gap is found.
	assertSignOffIsRealAndStillBinding(t, root,
		filepath.Join("fixtures", "shaped-text", "thai-signoff.json"),
		liveShapedTextSignOffDigest(t, root),
	)
	assertSignOffIsRealAndStillBinding(t, root,
		filepath.Join("fixtures", "expected-breaks", "break-signoff.json"),
		liveExpectedBreaksDigest(t, root),
	)

	// (b) The OBSERVED obligation set, gathered from the tree itself.
	moduleRoot := filepath.Join(root, "folio-go")
	got := map[string]bool{}
	scanned := 0
	werr := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || (path != moduleRoot && strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		scanned++
		if hasMatrixBuildConstraint(string(src)) {
			rel, _ := filepath.Rel(moduleRoot, path)
			got["matrix-file: "+filepath.ToSlash(rel)] = true
		}
		return nil
	})
	if werr != nil {
		t.Fatalf("walk %s: %v", moduleRoot, werr)
	}
	if scanned == 0 {
		t.Fatal("vacuity guard: the walk read 0 Go files, so \"the obligation set is unchanged\" says nothing")
	}

	// The registered matrix documents. matrix_test.go is itself
	// matrix-tagged and therefore not compiled into this binary, so its
	// slugs are read from SOURCE — the same mechanism
	// matrix_registration_test.go uses (which separately pins these
	// slugs against .github/workflows/matrix.yml; this guard pins them
	// against the declared obligation list, a different question).
	//
	// Both guards call folio.MatrixDocumentSlugsFromSource and nothing
	// else. They previously shared a REGEX (`(?m)^\s*slug:\s*"…"`), which
	// made them one guard wearing two names: a single-line composite
	// literal is gofmt-clean, compiles under -tags matrix, and was
	// measured invisible to BOTH at once (Story 2.5 review, Finding 1).
	// The shared reader parses the Go grammar, where the literal has no
	// second spelling, and FAILS on any entry it cannot read rather than
	// dropping it from the set.
	slugs, elements, serr := folio.MatrixDocumentSlugsFromSource(filepath.Join(moduleRoot, "matrix_test.go"))
	if serr != nil {
		t.Fatalf("read matrixDocuments from matrix_test.go: %v", serr)
	}
	if len(slugs) == 0 || elements == 0 {
		t.Fatal("vacuity guard: no matrixDocuments entries found in matrix_test.go — the document half of the obligation set would be silently empty")
	}
	// N-of-N witness: every literal element yielded a slug. Asserted here
	// as well as inside the reader, so this guard does not take the
	// reader's word for its own completeness.
	if len(slugs) != elements {
		t.Fatalf("vacuity guard: read %d slugs from %d matrixDocuments entries — a registered document is missing from the observed obligation set", len(slugs), elements)
	}
	t.Logf("obligation witness — %d matrix documents read from %d matrixDocuments literal entries, %d Go files walked", len(slugs), elements, scanned)
	for _, slug := range slugs {
		got["matrix-document: "+slug] = true
	}

	// (c) Set equality against the declared list, in both directions.
	want := map[string]bool{}
	for _, o := range declaredEpic2GateObligations {
		want[o] = true
	}
	for o := range want {
		if !got[o] {
			t.Errorf("declared Epic 2 gate obligation %q is NOT present in the tree — obligations must not disappear silently either. Either restore it, or remove its line from declaredEpic2GateObligations with the ruling that discharged it.", o)
		}
	}
	for o := range got {
		if !want[o] {
			t.Errorf(
				"%q is an Epic 2 gate obligation and is NOT in declaredEpic2GateObligations.\n\n"+
					"An obligation may not be added without a ruling that says so explicitly. Record it as a "+
					"ONE-LINE addition to that list, naming the story and the decision that authorised it "+
					"(D-2.5.1). Do not rename this test, and do not encode a count anywhere.",
				o)
		}
	}
}

// assertSignOffIsRealAndStillBinding is (a)'s successor guard. It
// asserts three things about the sign-off file at relPath, and reports
// which one failed rather than a bare "invalid":
//
//  1. PRESENCE — the file exists and is valid JSON.
//  2. COMPLETENESS — reader, date, examined and sha256 are all
//     non-empty, and date parses as ISO-8601 (YYYY-MM-DD). A record
//     missing any field is indistinguishable from no sign-off at all
//     (the same bar shaped_signoff_matrix_test.go and
//     expected_breaks_signoff_matrix_test.go hold their records to).
//  3. STILL BINDING — the record's sha256 equals wantDigest, computed
//     from the LIVE fixture at call time. A re-record moves wantDigest
//     and this fails by construction (D-2.3.5's anti-rot condition),
//     which is the entire reason the sign-off is checked against a
//     recomputed value and not merely against its own past existence.
func assertSignOffIsRealAndStillBinding(t *testing.T, root, relPath, wantDigest string) {
	t.Helper()
	path := filepath.Join(root, relPath)

	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Errorf("%s does not exist. Its own matrix-gated test (shaped_signoff_matrix_test.go or expected_breaks_signoff_matrix_test.go) still tracks that obligation as outstanding; this guard only asserts the record once it exists, so its absence here is not itself a failure of THIS guard — but if you expected it to be present, it is not.", relPath)
		return
	}
	if err != nil {
		t.Errorf("read %s: %v", relPath, err)
		return
	}

	var rec struct {
		Reader   string `json:"reader"`
		Date     string `json:"date"`
		Examined string `json:"examined"`
		SHA256   string `json:"sha256"`
	}
	if uerr := json.Unmarshal(raw, &rec); uerr != nil {
		t.Errorf("%s is not valid JSON: %v", relPath, uerr)
		return
	}

	for _, f := range []struct{ name, value string }{
		{"reader", rec.Reader},
		{"date", rec.Date},
		{"examined", rec.Examined},
		{"sha256", rec.SHA256},
	} {
		if strings.TrimSpace(f.value) == "" {
			t.Errorf("%s has an empty %q. A sign-off missing any field is indistinguishable from no sign-off at all.", relPath, f.name)
		}
	}
	if _, derr := time.Parse("2006-01-02", rec.Date); derr != nil {
		t.Errorf("%s's \"date\" %q does not parse as ISO-8601 (YYYY-MM-DD): %v", relPath, rec.Date, derr)
	}
	if rec.SHA256 != wantDigest {
		t.Errorf(
			"%s names digest %s, but the fixture it certifies now hashes to %s.\n\n"+
				"The sign-off is STALE: the fixture moved since it was signed off, so the record applies to "+
				"bytes nobody looked at. This is the anti-rot condition D-2.3.5 and D-2.4.3 both require — "+
				"re-read the changed fixture and update both the digest and what was examined. Do not simply "+
				"paste the new digest in.",
			relPath, rec.SHA256, wantDigest)
	}
}

// liveShapedTextSignOffDigest returns the digest fixtures/shaped-text's
// Thai READING sign-off must currently name — read from
// expected.json's own "sha256" field, the same value
// shaped_signoff_matrix_test.go's assertSignOffMatchesFrozenHash binds
// to. It is deliberately NOT sha256(expected.json)'s own bytes: the
// field is the fixture's self-declared digest of expected.pdf, which is
// the artifact the human sign-off is actually about.
func liveShapedTextSignOffDigest(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "fixtures", "shaped-text", "expected.json")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc struct {
		SHA256 string `json:"sha256"`
	}
	if uerr := json.Unmarshal(body, &doc); uerr != nil {
		t.Fatalf("%s is not valid JSON: %v", path, uerr)
	}
	if strings.TrimSpace(doc.SHA256) == "" {
		t.Fatalf("%s carries no sha256 to bind the sign-off to", path)
	}
	return doc.SHA256
}

// liveExpectedBreaksDigest returns the digest fixtures/expected-breaks's
// BREAK sign-off must currently name — the sha256 of
// expected_breaks.json's own bytes, computed fresh rather than read from
// expectedBreaksDigest's literal in expected_breaks_digest_test.go. Two
// independently-computed values agreeing is a stronger presence claim
// than one value asserted twice, and the two guards exist for different
// reasons (D-2.6.4 pins the fixture; this pins the sign-off to it).
func liveExpectedBreaksDigest(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "fixtures", "expected-breaks", "expected_breaks.json")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// hasMatrixBuildConstraint reports whether src carries an actual
// //go:build constraint naming the matrix tag — not merely a mention of
// one.
//
// The distinction is load-bearing and was found by this test failing on
// its own source. A substring search over the whole file reported SIX
// matrix-tagged files where there are three, because it also matched
// every file that DISCUSSES the tag in a comment — including this one.
// A guard that counts discussions of a constraint as instances of it
// would fire on any story that documented the gate, which is precisely
// the story that must not be blocked.
//
// So the constraint is recognised where the go tool recognises it: on
// its own line, in the build-constraint region, which ends at the
// package clause.
func hasMatrixBuildConstraint(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			return false
		}
		if !strings.HasPrefix(trimmed, "//go:build ") {
			continue
		}
		for _, tok := range strings.FieldsFunc(strings.TrimPrefix(trimmed, "//go:build "), func(r rune) bool {
			return r == ' ' || r == '(' || r == ')' || r == '!' || r == '&' || r == '|'
		}) {
			if tok == "matrix" {
				return true
			}
		}
	}
	return false
}

// boundaryGateDocuments are the human-facing epic-boundary gate files.
// Unlike a story file, these record digests as FORWARD-LOOKING
// INSTRUCTIONS — "the deferred legs will compare X" — so a human reads
// one and acts on it. Nothing else in this repository checks them:
// goldenDigestSearchScope deliberately excludes _bmad-output/ (see its
// comment), and that exclusion is right, because story files record
// past-tense measurements that must never be rewritten.
//
// WHY A SHAPE CHECK RATHER THAN A VALUE CHECK. A value check would have
// to re-derive which digest each row is ABOUT, which is prose, and it
// would drag _bmad-output/ back into a scope that is excluded on
// purpose. But a sha256 has a checkable shape independent of its value,
// and Story 2.5a shipped a 65-character one here — one 'f' too many, in
// the single row whose matrix legs were deferred and therefore never
// ran. A digest of the wrong LENGTH cannot be right for anything, so it
// is detectable with no knowledge of what it should have been.
var boundaryGateDocuments = []string{
	filepath.Join("_bmad-output", "implementation-artifacts", "epic-1-boundary-gate.md"),
	filepath.Join("_bmad-output", "implementation-artifacts", "epic-2-boundary-gate.md"),
}

// digestShapeMinRun is the length at which a delimited hex run stops
// being plausibly anything else (an abbreviated commit hash, a byte
// offset, a hex colour) and starts being a digest that got mistyped.
// Measured over both gate documents at the time of writing: 647
// delimited hex runs in total, of which the longest below this
// threshold is 8 characters, so nothing legitimate is caught by it.
const digestShapeMinRun = 32

// TestBoundaryGateDigestsAreWellFormed asserts that every digest-shaped
// token in a boundary-gate document is exactly 64 hex characters.
//
// A run is DELIMITED: neither of its neighbouring characters may be
// alphanumeric, so a hex substring of a longer identifier is not
// mistaken for a digest, and — the case that matters — a 65-character
// run is seen as 65 rather than as a valid 64 followed by a stray
// character.
func TestBoundaryGateDigestsAreWellFormed(t *testing.T) {
	root := repoRootForByteNeutrality(t)

	if len(boundaryGateDocuments) == 0 {
		t.Fatal("vacuity guard: no boundary-gate document is declared, so this test asserts nothing")
	}

	wellFormed, scanned := 0, 0
	for _, rel := range boundaryGateDocuments {
		p := filepath.Join(root, rel)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("presence precondition: the declared boundary-gate document %s could not be read (%v), so the shape assertion below would pass by looking nowhere", rel, err)
		}
		if len(b) == 0 {
			t.Fatalf("presence precondition: %s is empty", rel)
		}
		scanned++
		for _, run := range delimitedHexRuns(string(b)) {
			if len(run) < digestShapeMinRun {
				continue
			}
			if len(run) != 64 {
				t.Errorf("%s records %q as a digest: that is %d hex characters, and a sha256 is 64. A digest of the wrong length cannot be correct for any artifact, and this file is read by a HUMAN executing a deferred gate leg, so nothing downstream will catch it (Story 2.5a, D-000.48).", rel, run, len(run))
				continue
			}
			wellFormed++
		}
	}

	// VACUITY GUARD. If no document carried a digest at all, every
	// assertion above was skipped and this test would certify a set it
	// never looked at.
	if wellFormed == 0 {
		t.Fatalf("vacuity guard: %d boundary-gate documents were scanned and not one 64-character digest was found, so the shape assertion compared nothing", scanned)
	}
	t.Logf("D-000.48: %d boundary-gate documents scanned; %d digest-shaped runs of >=%d characters, all exactly 64.", scanned, wellFormed, digestShapeMinRun)
}

// delimitedHexRuns returns every maximal run of hex characters in s
// whose neighbours are not alphanumeric.
func delimitedHexRuns(s string) []string {
	isHex := func(c byte) bool {
		return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
	}
	isAlnum := func(c byte) bool {
		return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
	}
	var out []string
	for i := 0; i < len(s); {
		if !isHex(s[i]) {
			i++
			continue
		}
		j := i
		for j < len(s) && isHex(s[j]) {
			j++
		}
		beforeOK := i == 0 || !isAlnum(s[i-1])
		afterOK := j == len(s) || !isAlnum(s[j])
		if beforeOK && afterOK {
			out = append(out, s[i:j])
		}
		i = j
	}
	return out
}

// repoRootForByteNeutrality walks up until it finds a directory holding
// both folio-go/ and lint/ — the same D-000.5/AD-21 pattern the other
// root-finders in this module use.
func repoRootForByteNeutrality(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		folioGo, e1 := os.Stat(filepath.Join(dir, "folio-go"))
		lintDir, e2 := os.Stat(filepath.Join(dir, "lint"))
		if e1 == nil && folioGo.IsDir() && e2 == nil && lintDir.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find the repo root (a directory containing both folio-go/ and lint/) walking up from %s", dir)
		}
		dir = parent
	}
}
