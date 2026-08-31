package folio

import (
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/template"
)

// This file pins `chainFaceNames` (render.go) — THE one boundary between the
// document's chain and the render path's face-name list.
//
// IT LIVES HERE, IN `package folio`, BECAUSE NOWHERE ELSE CAN. The function's
// consequences are consequences at Render and Validate, and both are exported
// from this package; internal/template is a different package that cannot call
// either, so its tests can pin the FORMAT's behaviour and nothing about the
// render path's. Story 8.3's first draft of the comment on chainFaceNames
// claimed these properties were "pinned by test (fonts_embedded_test.go)" —
// they were not, and could not have been. Deleting the `Embedded()` arm
// outright left the whole Go suite green.
//
// WHAT THE BOUNDARY DOES CHANGED AT STORY 8.4, AND SO DID THESE PINS. Until
// 8.4 an embedded entry was DROPPED here; now it is mapped to the reserved
// face name embeddedFaceName derives from its ASSET KEY, and the fontCache
// resolves that name from the document's own assets. The two mutation pins
// below are re-pointed at the new boundary rather than retired, because the
// defects they catch are the same two defects — they have only swapped which
// side of the boundary is correct.
//
// TWO MUTATIONS, because they are caught in different places and knowing which
// is which is the difference between coverage and the feeling of it:
//   - DROPPING the embedded arm — a straight regression to the 8.3 boundary —
//     shortens the chain. The unit half below catches it on length and order;
//     the end-to-end halves catch it because the document then has no face
//     that covers Thai at all.
//   - Emitting the BARE `entry.AssetKey` instead of the derived name — a
//     64-character digest reaching the render path as an unqualified face name,
//     which is what would let a caller's FontSet shadow a document's carried
//     face — is caught by TestEmbeddedFaceNameIsDerivedFromTheAssetKey and by
//     TestEmbeddedFaceWinsOverACollidingFontSetKey.

// embeddedChainDoc is a one-text-element document whose `body` chain is
// `chain` verbatim, carrying the same font asset fixtures/embedded-font/ does
// so an `{"asset": …}` entry has something real to name.
func embeddedChainDoc(t *testing.T, chain string) string {
	t.Helper()
	source := embeddedFontTemplateJSON()
	// Spliced between the two keys that BRACKET the fonts block in canonical
	// (sorted) order — `bands` before it, `locale` after it — rather than by
	// counting braces. Brace counting over a document carrying ~47 KB of
	// base64 is a second JSON parser written badly; the sort order is a
	// property AD-9 guarantees.
	const open, next = "  \"fonts\": {", "  \"locale\":"
	start := strings.Index(source, open)
	end := strings.Index(source, next)
	if start < 0 || end < 0 || end < start {
		t.Fatalf("fixture assumption violated: the generated document's fonts block is not bracketed by %q and %q", open, next)
	}
	return source[:start] + "  \"fonts\": {\"body\": " + chain + "},\n" + source[end:]
}

// embeddedChainDocText is embeddedChainDoc with the drawn text substituted
// too. The fixture's own text is pure Thai (Story 8.4, D-8.4.4b), which is the
// SHARP witness for "the carried face is the only one that can draw this" and
// is exactly the wrong input for an ORDER assertion — the carried face covers
// both scripts, so a chain's order only shows up in mixed text.
func embeddedChainDocText(t *testing.T, chain, text string) string {
	t.Helper()
	const quoted = `"value": "สัญญา"`
	source := embeddedChainDoc(t, chain)
	if !strings.Contains(source, quoted) {
		t.Fatalf("fixture assumption violated: the generated document does not carry %s", quoted)
	}
	return strings.Replace(source, quoted, `"value": "`+text+`"`, 1)
}

// nonFontAssetKey and nonFontAssetData are the three-by-two-pixel PNG several
// tests in this package already carry, under the key the format's own rule
// produces for it. It is reused rather than minted so no new binary — not even
// an inline one — enters the repository for DW-83's sake.
const nonFontAssetKey = "5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078"

const nonFontAssetData = `"iVBORw0KGgoAAAANSUhEUgAAAAMAAAACCAIAAAASFvFNAAAAGElEQVR42mL6z8DAAMZMEOo/AwMg", "AAD//zwUBf/NjsW5AAAAAElFTkSuQmCC"`

// nonFontChainDoc adds that PNG to the document's assets map and hands back a
// document whose `body` chain is `chain` — so a chain entry can name it.
//
// The asset is a REAL, VALID PNG under a REAL image media type. Rewriting the
// carried FONT's mediaType to `image/png` would not do: image/png is a
// recognised image type, so the loader sniffs the bytes and refuses the
// document at load — which is a different defect entirely, and not the one
// DW-83 is about. DW-83's document is valid: it carries a perfectly good
// image and names it where a face belongs.
func nonFontChainDoc(t *testing.T, chain, text string) string {
	t.Helper()
	source := embeddedChainDocText(t, chain, text)
	const anchor = "  \"assets\": {\n"
	at := strings.Index(source, anchor)
	if at < 0 {
		t.Fatal("fixture assumption violated: the generated document has no assets block")
	}
	// Inserted FIRST because "5a05…" sorts before the font asset's "c945…",
	// keeping the document in the canonical key order AD-9 guarantees.
	png := "    \"" + nonFontAssetKey + "\": {\n      \"data\": [" + nonFontAssetData + "],\n      \"mediaType\": \"image/png\"\n    },\n"
	return source[:at+len(anchor)] + png + source[at+len(anchor):]
}

// TestChainFaceNamesMapsEveryEntryAndKeepsOrder is the unit-level half: the
// function itself, over every mixture, asserting BOTH that an embedded entry
// becomes its derived name and that nothing else moves.
//
// The order assertion is not decoration. A converter written as "walk the
// entries" is trivially order-preserving; one written as "partition, then
// concatenate" is not, and the two are indistinguishable from a chain whose
// embedded entry happens to sit last. Every case here puts one somewhere else.
func TestChainFaceNamesMapsEveryEntryAndKeepsOrder(t *testing.T) {
	face := template.FaceEntry
	asset := template.AssetEntry
	k1, k2 := embeddedFaceName("k1"), embeddedFaceName("k2")
	for _, tc := range []struct {
		name  string
		chain []template.FontChainEntry
		want  []string
	}{
		{"no entries", nil, []string{}},
		{"named faces only", []template.FontChainEntry{face("A"), face("B"), face("C")}, []string{"A", "B", "C"}},
		{"embedded first", []template.FontChainEntry{asset("k1"), face("A"), face("B")}, []string{k1, "A", "B"}},
		{"embedded in the middle", []template.FontChainEntry{face("A"), asset("k1"), face("B")}, []string{"A", k1, "B"}},
		{"embedded last", []template.FontChainEntry{face("A"), face("B"), asset("k1")}, []string{"A", "B", k1}},
		{"two embedded, interleaved", []template.FontChainEntry{asset("k1"), face("A"), asset("k2"), face("B")}, []string{k1, "A", k2, "B"}},
		{"embedded only", []template.FontChainEntry{asset("k1"), asset("k2")}, []string{k1, k2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := chainFaceNames(tc.chain)
			if len(got) != len(tc.want) {
				t.Fatalf("chainFaceNames(%v) = %v, want %v", tc.chain, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("chainFaceNames(%v) = %v, want %v — the document's authored order is not preserved", tc.chain, got, tc.want)
				}
			}
			// The BARE asset key must never appear as a face name. If it did,
			// the render path would carry a 64-character digest in the same
			// namespace a caller's FontSet keys live in, and the two could
			// shadow one another — the collision AD-8 keeps a document's
			// carried face and a caller's supplied face out of.
			for _, name := range got {
				if name == "k1" || name == "k2" {
					t.Fatalf("an embedded entry's BARE asset key reached the render path as a face name: %v", got)
				}
			}
		})
	}
}

// TestEmbeddedFaceNameIsDerivedFromTheAssetKey pins D-8.4.1 at the derivation
// itself: the name comes from the ASSET KEY and from nothing else.
//
// It is a separate test from the mapping above because the mapping would pass
// under a derivation that read `font.family` — the entry would still be
// converted, still in order, and still not the bare key. That derivation is
// precisely the one AD-8 forbids: a document's "Inter" and a caller's "Inter"
// would collide in one namespace.
func TestEmbeddedFaceNameIsDerivedFromTheAssetKey(t *testing.T) {
	key := embeddedFontAssetKey()
	name := embeddedFaceName(key)
	if !strings.Contains(name, key) {
		t.Fatalf("the embedded face name %q does not carry its asset key %q", name, key)
	}
	if name == key {
		t.Fatal("the embedded face name is the bare asset key — it must be qualified, or it shares a namespace with the caller's FontSet keys")
	}
	// The document's own `font.family` is "Noto Sans Thai", and a shipped face
	// of that exact name is in the FontSet. The derived name must contain
	// neither, or the two could substitute for one another.
	for _, family := range []string{"Noto Sans Thai", "Noto Sans"} {
		if strings.Contains(name, family) {
			t.Errorf("the embedded face name %q is derived from font.family %q — AD-8/D-8.4.1: the ASSET KEY decides", name, family)
		}
	}
	// Length first: under the pre-8.4 boundary this is EMPTY, and indexing it
	// would panic and take the whole test binary down with it — which hides
	// every other failure the same mutation causes.
	got := chainFaceNames([]template.FontChainEntry{template.AssetEntry(key)})
	if len(got) != 1 || got[0] != name {
		t.Errorf("the boundary and the derivation disagree about an embedded entry's face name: chainFaceNames = %v, want [%s]", got, name)
	}
}

// pdfEscapedEmbeddedFaceName is the derived face name as internal/pdf spells
// it in a resource dictionary and a content stream. pdfNameEscape keeps
// [A-Za-z0-9_-] and drops everything else, so the derivation's separator goes
// and the asset key's 64 hex characters stay — which is what makes the
// produced PDF itself a witness to WHICH face drew the page, by identity and
// never by a count of embedded programs.
func pdfEscapedEmbeddedFaceName(assetKey string) string {
	var b strings.Builder
	for _, r := range embeddedFaceName(assetKey) {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// renderEmbeddedChainDoc parses and renders one of embeddedChainDoc's
// variants against the SHIPPED FontSet alone, and returns the bytes.
func renderEmbeddedChainDoc(t *testing.T, source string) []byte {
	t.Helper()
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := Render(tpl, Data(`{}`), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return res.Bytes
}

// TestEmbeddedEntryRendersFromTheDocumentsOwnBytes is the end-to-end half of
// the same claim, and it is the pin that reddens if the embedded arm of
// chainFaceNames is dropped.
//
// The document's text is Thai, the FontSet is the SHIPPED set, and the chain
// is ["Noto Sans", <the carried face>]. Measured from the shipped cmaps:
// NotoSans-Regular covers ZERO codepoints in U+0E00–U+0E7F. So the only face
// that can draw this text is the one the document CARRIES — and if the
// embedded entry contributed nothing, every rune would be uncovered, would be
// omitted from the output, and would raise a missing-glyph Warning.
//
// That is the assertion: a clean render, no diagnostics, and drawn glyphs.
// "It rendered" alone is satisfied by a render that dropped every rune.
func TestEmbeddedEntryRendersFromTheDocumentsOwnBytes(t *testing.T) {
	source := embeddedChainDoc(t, `["Noto Sans", {"asset": "`+embeddedFontAssetKey()+`"}]`)
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := Render(tpl, Data(`{}`), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("a chain whose only covering entry is the carried face must render: %v", err)
	}
	if len(res.Diagnostics) != 0 {
		t.Fatalf("rendering from the carried face produced diagnostics — every rune fell through the chain uncovered: %+v", res.Diagnostics)
	}
	// The face the page is drawn with is the DOCUMENT's, named by its derived
	// key, and that key reaches the produced PDF as the font resource name.
	if !strings.Contains(string(res.Bytes), pdfEscapedEmbeddedFaceName(embeddedFontAssetKey())) {
		t.Error("the produced PDF names no font resource derived from the carried asset's key — the page was not drawn with the document's own face")
	}
	// And the shipped Latin face the chain names FIRST drew nothing, because
	// it covers none of this text — so it was never subset and never embedded.
	if strings.Contains(string(res.Bytes), "+NotoSans-Regular") {
		t.Error("the shipped Latin face reached the page, though it covers none of the drawn text")
	}
}

// TestEmbeddedFaceWinsOverACollidingFontSetKey pins the PRECEDENCE rule the
// derived name's namespace rests on: a caller cannot shadow a document's
// carried face by filing bytes under the same string.
//
// The FontSet handed in below carries the shipped LATIN Noto Sans under the
// embedded face's own derived name. If the FontSet arm were consulted first —
// or at all — the Thai text would be drawn with a Latin face that covers none
// of it, and every rune would raise a missing-glyph Warning.
func TestEmbeddedFaceWinsOverACollidingFontSetKey(t *testing.T) {
	source := embeddedChainDoc(t, `[{"asset": "`+embeddedFontAssetKey()+`"}]`)
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	fs := FontSet{}
	for name, data := range testShippedFontSet() {
		fs[name] = data
	}
	fs[embeddedFaceName(embeddedFontAssetKey())] = fs["Noto Sans"]
	res, err := Render(tpl, Data(`{}`), nil, fs)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if len(res.Diagnostics) != 0 {
		t.Fatalf("a caller's FontSet entry shadowed the document's own carried face: %+v", res.Diagnostics)
	}
}

// TestChainResolvesInTheDeclaredOrder is the ORDER half at the render
// surface, and since Story 8.4 it has teeth it could not have before: the
// carried face covers BOTH scripts, so which entry sits first decides which
// face draws the Latin.
//
// ["Noto Sans", <carried Thai>] draws Latin with the shipped face and Thai
// with the carried one. [<carried Thai>, "Noto Sans"] draws BOTH with the
// carried one, because the first entry that covers a rune wins (AD-8's Rule).
// The two must therefore differ — and a converter that dropped everything
// after the first embedded entry, or that sorted the chain, would make them
// agree.
func TestChainResolvesInTheDeclaredOrder(t *testing.T) {
	const mixed = "A สัญญา"
	shippedFirst := renderEmbeddedChainDoc(t, embeddedChainDocText(t, `["Noto Sans", {"asset": "`+embeddedFontAssetKey()+`"}]`, mixed))
	carriedFirst := renderEmbeddedChainDoc(t, embeddedChainDocText(t, `[{"asset": "`+embeddedFontAssetKey()+`"}, "Noto Sans"]`, mixed))
	if string(shippedFirst) == string(carriedFirst) {
		t.Fatal("reversing the chain changed nothing — coverage resolution is not walking the document's authored order")
	}
	// The witness that the difference is the one claimed: with the carried
	// face first, the shipped Latin face is not embedded at all.
	if strings.Contains(string(carriedFirst), "+NotoSans-Regular") {
		t.Error("the shipped Latin face reached the page even though the carried face precedes it and covers every rune")
	}
	if !strings.Contains(string(shippedFirst), "+NotoSans-Regular") {
		t.Error("the shipped Latin face did NOT reach the page even though it precedes the carried face — the Latin text was drawn with the wrong face")
	}
}

// TestNonFontAssetIsAcceptedAtLoad is DW-83's first half, and it is the one
// this story must NOT "fix". A chain entry may name an asset that is not a
// font: the loader checks that the key is present in the assets map and
// nothing more, because D-1.8.1 as amended preserves an unrecognised or
// wrong-kind media type at load and errors at render.
func TestNonFontAssetIsAcceptedAtLoad(t *testing.T) {
	if _, err := ParseTemplate([]byte(nonFontChainDoc(t, `["Noto Sans", {"asset": "`+nonFontAssetKey+`"}]`, "สัญญา"))); err != nil {
		t.Fatalf("a chain entry naming a non-font asset must LOAD (D-1.8.1 as amended): %v", err)
	}
}

// TestNonFontAssetDrawnErrorsAtRenderAndAtValidate is DW-83's second and third
// halves, and the third is the one D-8.3.5 warns is the one that disappears.
//
// The document's text is Thai and the chain's other entry is the shipped Latin
// Noto Sans, which covers none of it — so coverage resolution MUST reach the
// entry naming the image, and that is the moment the renderer is asked to draw
// with something it cannot read.
//
// THE ERROR MUST LOCATE ITSELF at all three coordinates AC4 names: the chain,
// the entry index, and the asset key. A message naming only the asset key
// cannot be acted on when one chain is shared by many elements.
//
// AND VALIDATE MUST RETURN THE IDENTICAL ERROR — asserted by string equality
// and by the diagnostic count, so removing the Validate arm reddens THIS test
// on its own. This is chain_face_names_test.go's own shape, kept deliberately
// rather than the image precedent's (render_image_test.go asserts Render only,
// which is exactly the omission D-8.3.5 names).
func TestNonFontAssetDrawnErrorsAtRenderAndAtValidate(t *testing.T) {
	source := nonFontChainDoc(t, `["Noto Sans", {"asset": "`+nonFontAssetKey+`"}]`, "สัญญา")

	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("the document must LOAD: %v", err)
	}
	_, rerr := Render(tpl, Data(`{}`), nil, testShippedFontSet())
	if rerr == nil {
		t.Fatal("a chain entry naming a non-font asset must fail at Render once something must draw from it (DW-83)")
	}
	for _, want := range []string{
		"element e1",                // the element that needed the face
		`font chain "body" entry 1`, // WHICH entry of WHICH chain
		"asset " + nonFontAssetKey,  // and which asset
		`cannot render font media type "image/png"`,
		"the document is valid", // a capability limit, never a format error
	} {
		if !strings.Contains(rerr.Error(), want) {
			t.Errorf("the render error does not locate itself (%q missing): %v", want, rerr)
		}
	}

	diags, verr := Validate([]byte(source), Data(`{}`), nil, testShippedFontSet())
	if verr == nil {
		t.Fatal("Validate must PREDICT Render (D-1.8.1 amended): Render refuses this document, so Validate must too")
	}
	if verr.Error() != rerr.Error() {
		t.Errorf("Validate and Render disagree about the same document:\n\tValidate: %v\n\tRender:   %v", verr, rerr)
	}
	if len(diags) != 0 {
		t.Errorf("Validate returned %d diagnostic(s) alongside a hard error: %+v", len(diags), diags)
	}
}

// TestNonFontAssetNeverDrawnRendersClean is the OTHER side of the same rule,
// and it is what makes the error's placement a decision rather than an
// accident: the decode is lazy, at the point of use, so an entry no rune ever
// resolves to costs the render nothing and complains about nothing.
//
// The text here is LATIN, which the chain's first entry covers completely, so
// coverage resolution never reaches the entry naming the image. The chain is
// still walked for the vertical model — chainLineMetrics visits every entry —
// and the tolerance there (fontCache.metricsFace) is what keeps this clean.
func TestNonFontAssetNeverDrawnRendersClean(t *testing.T) {
	source := nonFontChainDoc(t, `["Noto Sans", {"asset": "`+nonFontAssetKey+`"}]`, "Latin only")
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := Render(tpl, Data(`{}`), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("a chain entry naming a non-font asset that nothing draws from must render clean: %v", err)
	}
	if len(res.Diagnostics) != 0 {
		t.Fatalf("an entry nothing draws from produced diagnostics: %+v", res.Diagnostics)
	}
	if _, verr := Validate([]byte(source), Data(`{}`), nil, testShippedFontSet()); verr != nil {
		t.Fatalf("Validate must predict this clean render, got: %v", verr)
	}
}

// TestChainOfOnlyUnusableEntriesProducesTheExistingLocatedError keeps the pin
// Story 8.3 wrote, re-pointed at the condition that can still reach it.
//
// A chain of nothing but embedded entries is now perfectly renderable, so the
// document that used to produce this error no longer can. What still can is a
// chain whose every entry is a face NOTHING supplies — the pre-existing
// empty-metrics path, deliberately not widened by this story. The assertion is
// here so that if a later story does widen it, the change is deliberate.
func TestChainOfOnlyUnusableEntriesProducesTheExistingLocatedError(t *testing.T) {
	source := embeddedChainDoc(t, `["No Such Face"]`)

	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("a chain naming an unsupplied face must LOAD: %v", err)
	}
	_, rerr := Render(tpl, Data(`{}`), nil, testShippedFontSet())
	if rerr == nil {
		t.Fatal("a chain with no usable entry must fail at Render — nothing can draw the text")
	}
	for _, want := range []string{"element e1", "none of the fallback chain's faces"} {
		if !strings.Contains(rerr.Error(), want) {
			t.Errorf("the render error does not locate itself (%q missing): %v", want, rerr)
		}
	}

	diags, verr := Validate([]byte(source), Data(`{}`), nil, testShippedFontSet())
	if verr == nil {
		t.Fatal("Validate must PREDICT Render (D-1.8.1 amended): Render refuses this document, so Validate must too")
	}
	if verr.Error() != rerr.Error() {
		t.Errorf("Validate and Render disagree about the same document:\n\tValidate: %v\n\tRender:   %v", verr, rerr)
	}
	if len(diags) != 0 {
		t.Errorf("Validate returned %d diagnostic(s) alongside a hard error: %+v", len(diags), diags)
	}
}
