package fonts

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestShippedRobotoMatchesDesignerCatalogue is Story 16.8's machine-checked
// form of "THERE IS EXACTLY ONE ROBOTO" — its own Always clause, restated
// here as a run rather than a comment.
//
// WHY THIS TEST EXISTS, NAMED PLAINLY. Two cuts of Roboto already sit in
// this repository under one name, 180 KB apart: the designer's own
// catalogue face (`folio-designer/public/fonts/roboto/Roboto-Regular.ttf`,
// `font-catalogue.json`'s `"roboto"` entry, sha256 e688a215…, 347.6 KB) and
// a test fixture (`folio-go/testdata/fonts/Roboto-Regular.ttf`, sha256
// 79e85140…, 167.7 KB) used only to exercise the Apache-2.0 branch of
// internal/fontset's licence-signature table. Shipping the wrong one would
// not fail a build — it would render, silently, and every document naming
// "Roboto" would disagree pixel-for-pixel depending on which binary
// produced it. That is worse than a build failure, so it is checked by a
// run rather than left to a comment ("a comment is not a measurement").
//
// WHY THE COMPARISON IS AGAINST THE DESIGNER'S FILE, READ FRESH, RATHER
// THAN A LITERAL DIGEST RESTATED HERE. A hardcoded sha256 on this side
// could be "kept honest" by editing one string the moment either file
// changed, which is exactly the laundering font-binary-identity.test.ts's
// own digest ties (on the designer side) are written to resist. Reading
// the designer's committed file at test time means a byte in either copy
// moving independently is what turns this test red — not a copy of a
// number moving in lockstep with it.
func TestShippedRobotoMatchesDesignerCatalogue(t *testing.T) {
	shipped, ok := Shipped()["Roboto"]
	if !ok {
		t.Fatal(`fonts.Shipped() carries no "Roboto" key — Story 16.8 requires the engine to ship a fourth face, ` +
			`named exactly "Roboto", alongside the three Story 2.2 Noto faces`)
	}

	// THE DESIGNER'S CATALOGUE FILE, READ AS THE SOURCE OF TRUTH. Story
	// 16.8's Code Map anchors this exact path and its recorded digest
	// (e688a215e0841b6e4edb…, 347.6 KB) as "the cut to ship" — it is what
	// `font-catalogue.json`'s `"roboto"` entry resolves to and what
	// `scripts/build-wasm.mjs` fingerprints into
	// `catalogue-roboto.<hash>.ttf` for the browser.
	catalogueRobotoPath := filepath.Join("..", "..", "folio-designer", "public", "fonts", "roboto", "Roboto-Regular.ttf")
	catalogueBytes, err := os.ReadFile(catalogueRobotoPath)
	if err != nil {
		t.Fatalf("could not read the designer's catalogue Roboto at %s: %v", catalogueRobotoPath, err)
	}

	shippedDigest := sha256.Sum256(shipped)
	catalogueDigest := sha256.Sum256(catalogueBytes)

	if shippedDigest != catalogueDigest {
		t.Fatalf(
			"folio-go's embedded \"Roboto\" (sha256 %s, %d bytes) is NOT byte-identical to the designer catalogue's "+
				"Roboto at %s (sha256 %s, %d bytes). THERE IS EXACTLY ONE ROBOTO: two cuts under one name would make "+
				"\"fontFamily\": \"Roboto\" render differently depending on which build produced the document. If "+
				"this fired, the embedded bytes were copied from the wrong source — re-copy them byte-for-byte from "+
				"the designer's committed catalogue file rather than re-downloading or re-deriving them.",
			hex.EncodeToString(shippedDigest[:]), len(shipped), catalogueRobotoPath, hex.EncodeToString(catalogueDigest[:]), len(catalogueBytes),
		)
	}

	// NON-VACUITY: a size floor so a truncated or empty read on either side
	// cannot compare two near-empty slices and call it identity. 300 KB
	// sits comfortably below the recorded 347.6 KB and above the kind of
	// truncation a broken embed or a broken read would produce.
	const minPlausibleBytes = 300_000
	if len(shipped) < minPlausibleBytes {
		t.Fatalf("fonts.Shipped()[\"Roboto\"] is only %d bytes, far short of the recorded 347.6 KB — the embed is reading something other than the shipped face", len(shipped))
	}

	// THE NEGATIVE CONTROL, NAMED IN THE INTENT CONTRACT ITSELF:
	// folio-go/testdata/fonts/Roboto-Regular.ttf IS A DIFFERENT CUT and must
	// stay different. This is not a defect to fix — it is a fixture for
	// internal/fontset's Apache-2.0 licence-signature test — but if it ever
	// became byte-identical to the shipped face, the sentence above stating
	// they differ would be quietly false, so that premise is checked too.
	testFixturePath := filepath.Join("..", "testdata", "fonts", "Roboto-Regular.ttf")
	testFixtureBytes, err := os.ReadFile(testFixturePath)
	if err != nil {
		t.Fatalf("could not read the test-fixture Roboto at %s: %v", testFixturePath, err)
	}
	testFixtureDigest := sha256.Sum256(testFixtureBytes)
	if testFixtureDigest == shippedDigest {
		t.Fatalf(
			"folio-go/testdata/fonts/Roboto-Regular.ttf (the Apache-2.0 licence-signature fixture) is now "+
				"byte-identical to the shipped face. This test's own header comment says the two are different "+
				"cuts by design (%s vs the fixture) — if that has changed, the comment above needs correcting "+
				"along with this assertion, not the other way round.",
			hex.EncodeToString(shippedDigest[:]),
		)
	}
}
