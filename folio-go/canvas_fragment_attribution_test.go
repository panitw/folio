package folio

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

// Story 8.4a: the canvas projection carries, per PAINT FRAGMENT, which of the
// document's own assets the engine resolved that fragment's face to.
//
// WHY THIS IS A SEPARATE FILE. canvas_embedded_face_test.go is Story 8.4's
// AD-17 measurement pin and is left byte-for-byte unmodified by this story —
// the claim that the canvas MEASURES with the carried face must keep being
// made by the test that made it, not by an edited descendant of it. What is
// new here is ATTRIBUTION: the same resolution, reported to the browser so it
// can rasterize with the same face. No metric, no advance and no line break
// moves; the fragment's x is still the engine's.
//
// WHAT MUTATION THIS IS RED-PROVED AGAINST: drop `AssetKey: carried` at the
// fragment append in page_setup.go and TestACarriedFragmentIsAttributedToIts-
// AssetKey fails on its first fragment. Drop the index check in
// fontCache.carriedAssetKey and TestAShippedFaceCarriesNoAssetKey still
// passes (a shipped name has no prefix to read), which is why the round-trip
// below asserts the index and the name agree rather than either alone.

// carriedFaceProjection projects a template that carries a font face, through
// the SHIPPED font set — the same one canvas_embedded_face_test.go uses, so
// the chain's first entry ("Noto Sans") really is resolvable and the fall
// through to the carried face is a coverage answer rather than an absence.
func carriedFaceProjection(t *testing.T, source string) CanvasProjection {
	t.Helper()
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	projection, err := CanvasWithTextPaint(tpl, testShippedFontSet())
	if err != nil {
		t.Fatalf("CanvasWithTextPaint: %v", err)
	}
	// A DEGRADED CHAIN IS HOW THIS GOES VACUOUS: addCanvasTextPaint disposes
	// of an element it cannot shape and carries on, so a projection that never
	// saw the carried face is still well formed and has nothing in it to
	// compare. This document has no table and no conditional visibility, so a
	// degraded chain is the only thing that can make the count inexact.
	if !projection.ContentWindowCountIsExact {
		t.Fatal("the engine reports an inexact window count for a document with no table and no conditional visibility — its font chain degraded, so no fragment below is attributed to anything")
	}
	return projection
}

func projectedFragments(projection CanvasProjection) []CanvasTextFragment {
	var out []CanvasTextFragment
	for _, component := range projection.Components {
		if component.TextPaint == nil {
			continue
		}
		for _, line := range component.TextPaint.Lines {
			out = append(out, line.Fragments...)
		}
	}
	return out
}

func TestACarriedFragmentIsAttributedToItsAssetKey(t *testing.T) {
	projection := carriedFaceProjection(t, embeddedFontTemplateJSON())
	fragments := projectedFragments(projection)
	if len(fragments) == 0 {
		t.Fatal("presence precondition: the carried-face document projected no paint fragments, so nothing below is asserted")
	}
	want := embeddedFontAssetKey()
	for _, fragment := range fragments {
		if fragment.AssetKey != want {
			t.Fatalf("fragment %q is attributed to asset key %q, want %q — this document's chain names a shipped Latin face first and PURE THAI text, so every fragment can only have been drawn with the face the document carries", fragment.Text, fragment.AssetKey, want)
		}
	}
	// AND IT IS THE KEY, NOT THE MINTED NAME. The browser derives a CSS family
	// from this string and hands it straight back to the `asset` operation, so
	// an engine-internal namespace reaching the wire would be a second
	// spelling of embedded_face.go's derivation in a second language.
	for _, fragment := range fragments {
		if strings.Contains(fragment.AssetKey, embeddedFaceNamePrefix) {
			t.Fatalf("fragment %q carries %q, which spells the reserved face-name prefix — the wire carries the ASSET KEY", fragment.Text, fragment.AssetKey)
		}
	}
}

func TestAShippedFaceCarriesNoAssetKey(t *testing.T) {
	projection := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON))
	fragments := projectedFragments(projection)
	if len(fragments) == 0 {
		t.Fatal("presence precondition: the shipped-face control document projected no paint fragments, so the absence below is asserted over nothing")
	}
	for _, fragment := range fragments {
		if fragment.AssetKey != "" {
			t.Fatalf("fragment %q drawn with a SHIPPED face is attributed to asset key %q — absence is the wire's statement that this fragment is a shipped face, and a present key would make the browser ask for a family nothing registers", fragment.Text, fragment.AssetKey)
		}
	}
}

// TestAMixedScriptElementIsAttributedFragmentByFragment is AC3. One element,
// one chain, two faces: the Latin runs through the shipped entry and the Thai
// falls through it to the face the document carries. Attributing at the
// COMPONENT would hand one of the two the other's glyphs, and the browser
// would paint at the engine's x with the wrong advances — the reported defect
// this whole line of work exists for.
func TestAMixedScriptElementIsAttributedFragmentByFragment(t *testing.T) {
	const thai = `"value": "สัญญา"`
	source := embeddedFontTemplateJSON()
	if !strings.Contains(source, thai) {
		t.Fatalf("fixture precondition: the carried-face document no longer spells %s, so this test cannot build the mixed-script variant from it", thai)
	}
	mixed := strings.Replace(source, thai, `"value": "Deed สัญญา"`, 1)
	fragments := projectedFragments(carriedFaceProjection(t, mixed))
	if len(fragments) < 2 {
		t.Fatalf("presence precondition: the mixed-script element projected %d fragments, and the whole point is that it produces more than one", len(fragments))
	}
	carried := 0
	shipped := 0
	for _, fragment := range fragments {
		switch fragment.AssetKey {
		case "":
			shipped++
		case embeddedFontAssetKey():
			carried++
		default:
			t.Fatalf("fragment %q is attributed to asset key %q, which this document does not carry", fragment.Text, fragment.AssetKey)
		}
	}
	if carried == 0 || shipped == 0 {
		t.Fatalf("the mixed-script element produced %d carried-face fragments and %d shipped-face ones; both must be present or the per-fragment claim is asserted over a uniform line", carried, shipped)
	}
}

// TestTheCarriedFaceNameAndTheAssetKeyAreOneDerivation asserts what
// fontCache.carriedAssetKey's doc comment claims by construction: the index is
// keyed by the minted name and records the same key the name was minted from,
// so reading the prefix back off the name and reading the index agree. It also
// pins the ORDER — the index is asked first — by showing a FontSet face whose
// name merely looks minted is not reported as one the document carries.
func TestTheCarriedFaceNameAndTheAssetKeyAreOneDerivation(t *testing.T) {
	const key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	name := embeddedFaceName(key)
	back, ok := embeddedFaceAssetKey(name)
	if !ok || back != key {
		t.Fatalf("embeddedFaceAssetKey(%q) = %q, %v; want %q, true", name, back, ok, key)
	}
	if _, ok := embeddedFaceAssetKey("Noto Sans"); ok {
		t.Fatal("a shipped face name reads as a minted one — the prefix is no longer a namespace")
	}

	tpl, err := ParseTemplate([]byte(embeddedFontTemplateJSON()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cache := newDocumentFontCache(tpl)
	minted := embeddedFaceName(embeddedFontAssetKey())
	got, ok := cache.carriedAssetKey(minted)
	if !ok || got != embeddedFontAssetKey() {
		t.Fatalf("carriedAssetKey(%q) = %q, %v; want the document's own asset key and true", minted, got, ok)
	}
	if got, ok := cache.carriedAssetKey("Noto Sans"); ok || got != "" {
		t.Fatalf("carriedAssetKey(%q) = %q, %v; a shipped face is carried by nothing", "Noto Sans", got, ok)
	}
	// THE ORDER. This name is in the mint's namespace by SHAPE and is absent
	// from the document's index, so the index's answer — not the string's — is
	// what must come back. A caller's FontSet may legally file an entry under
	// any string at all, and a shipped face that merely looks minted must not
	// be reported as a face the document carries.
	if got, ok := cache.carriedAssetKey(embeddedFaceName("deadbeef")); ok || got != "" {
		t.Fatalf("carriedAssetKey read the prefix off a name this document's index does not hold and returned %q, %v — the index is the authority, and the prefix is only read after it answers", got, ok)
	}
}

// TestANonMintedFaceNameYieldsNoAssetKey pins the one thing
// strings.CutPrefix does NOT do for free: on a miss it returns the whole
// input, not the empty string. embeddedFaceAssetKey therefore empties the
// value itself, so a caller that reads only the first result can never put a
// FACE NAME where an ASSET KEY belongs.
//
// It is a wire claim, not a style one. Story 8.4a projects this value as a
// fragment's `assetKey`, and the browser admits that field only as 64
// lowercase hex characters; anything else fails the whole projection's guard,
// which raises PROTOCOL_INVALID, terminates the worker and rejects every
// pending request — the session is dead until reload. Safe by construction
// today only because every name reaching the accessor came from the mint;
// this asserts the accessor is safe even when that stops being true.
func TestANonMintedFaceNameYieldsNoAssetKey(t *testing.T) {
	for _, name := range []string{
		"Noto Sans",
		"Noto Sans Thai",
		"",
		"asset",                     // the prefix without its colon
		"ASSET:0123456789abcdef",    // the namespace is case-sensitive
		" asset:0123456789abcdef",   // and is a PREFIX, not a substring
		"family:asset:deadbeefcafe", // ditto, one level in
	} {
		key, ok := embeddedFaceAssetKey(name)
		if ok {
			t.Fatalf("embeddedFaceAssetKey(%q) reported true; %q is not in the mint's namespace", name, name)
		}
		if key != "" {
			t.Fatalf("embeddedFaceAssetKey(%q) = %q on a miss; a miss must be the EMPTY string and never the name itself, which would reach the wire as a fragment's assetKey and fail the browser's 64-hex admission", name, key)
		}
	}
	// And the hit is unchanged: the key comes back whole, with true.
	const key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if got, ok := embeddedFaceAssetKey(embeddedFaceName(key)); !ok || got != key {
		t.Fatalf("embeddedFaceAssetKey(embeddedFaceName(%q)) = %q, %v; want %q, true", key, got, ok, key)
	}
}

// twoCarriedFacesTemplateJSON is the I/O matrix's "two embedded entries in one
// chain" row, built rather than shipped.
//
// NO NEW FIXTURE, NO NEW GOLDEN, NO NEW SIGN-OFF. A `fixtures/` directory
// would need an `expected.pdf`, which is a human-attested artifact under
// AD-21/D-4.7.1 and is above a builder's authority to mint. This document is
// never rendered: TestTwoCarriedFacesInOneChainAreAttributedToTheirOwnKeys
// projects it with CanvasWithTextPaint and reads the projection, so no page,
// no digest and nothing to attest are produced. And no binary enters the
// repository — the second face is the SHIPPED NotoSans-Regular the module
// already commits, embedded as an asset the same way the Thai one already is.
//
// THE CHAIN NAMES NO SHIPPED FACE AT ALL, which is the point: with
// ["Noto Sans", <carried Thai>] the Latin half is drawn by the FontSet and
// carries no key, so the two-carried case is never exercised. Here BOTH
// entries are assets the document carries, the Latin face covers the Latin
// run and covers none of the Thai, and each fragment must name its OWN key.
func twoCarriedFacesTemplateJSON(t *testing.T) (string, string, string) {
	t.Helper()
	source := embeddedFontTemplateJSON()
	thaiKey := embeddedFontAssetKey()
	latinRaw := testShippedNotoSans
	latinSum := sha256.Sum256(latinRaw)
	latinKey := fmt.Sprintf("%x", latinSum)
	if latinKey == thaiKey {
		t.Fatal("the two faces hash to one key, so the document carries one asset and the distinctness below is asserted over nothing")
	}
	var data strings.Builder
	for i, line := range base64Wrapped76(latinRaw) {
		if i > 0 {
			data.WriteString(",\n")
		}
		data.WriteString("        \"" + line + "\"")
	}
	const assetTail = "      \"mediaType\": \"font/ttf\"\n    }\n  },"
	if strings.Count(source, assetTail) != 1 {
		t.Fatalf("fixture precondition: the carried-face document no longer ends its one asset with %q, so a second asset cannot be spliced in beside it", assetTail)
	}
	source = strings.Replace(source, assetTail, "      \"mediaType\": \"font/ttf\"\n    },\n    \""+latinKey+"\": {\n      \"data\": [\n"+data.String()+"\n      ],\n      \"font\": {\n        \"family\": \"Noto Sans\",\n        \"licence\": \"SIL Open Font License 1.1\",\n        \"source\": \"folio-go/fonts/notosans/NotoSans-Regular.ttf — the shipped static Regular instance\",\n        \"style\": \"Regular\"\n      },\n      \"mediaType\": \"font/ttf\"\n    }\n  },", 1)
	const shippedHead = "    \"body\": [\n      \"Noto Sans\",\n      {"
	if strings.Count(source, shippedHead) != 1 {
		t.Fatalf("fixture precondition: the carried-face document no longer opens its chain with %q, so the shipped entry cannot be replaced by a carried one", shippedHead)
	}
	source = strings.Replace(source, shippedHead, "    \"body\": [\n      {\n        \"asset\": \""+latinKey+"\"\n      },\n      {", 1)
	const thai = "\"value\": \"สัญญา\""
	if strings.Count(source, thai) != 1 {
		t.Fatalf("fixture precondition: the carried-face document no longer spells %s, so the mixed-script variant cannot be built from it", thai)
	}
	source = strings.Replace(source, thai, "\"value\": \"Deed สัญญา\"", 1)
	return source, latinKey, thaiKey
}

// TestTwoCarriedFacesInOneChainAreAttributedToTheirOwnKeys closes the last
// unproved row of this story's I/O matrix at the GO ATTRIBUTION surface. The
// browser half was already proved twice — embedded-face-family.test.ts shows
// two distinct keys derive two distinct families, and
// embedded-face-registry.test.ts registers both — but the engine half, "each
// fragment names its OWN asset key", rested on a document that carries exactly
// one font. A projection that attributed every carried fragment to whichever
// key it found first would have passed every existing test in this file.
func TestTwoCarriedFacesInOneChainAreAttributedToTheirOwnKeys(t *testing.T) {
	source, latinKey, thaiKey := twoCarriedFacesTemplateJSON(t)
	fragments := projectedFragments(carriedFaceProjection(t, source))
	if len(fragments) < 2 {
		t.Fatalf("presence precondition: the mixed-script element projected %d fragments, and the claim is about more than one", len(fragments))
	}
	seen := map[string]string{}
	for _, fragment := range fragments {
		switch fragment.AssetKey {
		case latinKey, thaiKey:
			seen[fragment.AssetKey] = fragment.Text
		case "":
			t.Fatalf("fragment %q carries no asset key, but this chain names NO shipped face — every fragment can only have been drawn by a face the document carries", fragment.Text)
		default:
			t.Fatalf("fragment %q is attributed to asset key %q, which this document does not carry", fragment.Text, fragment.AssetKey)
		}
	}
	if len(seen) != 2 {
		t.Fatalf("the two carried entries produced %d distinct asset keys (%v); each fragment must name its OWN, or the browser paints one script with the other's face", len(seen), seen)
	}
}
