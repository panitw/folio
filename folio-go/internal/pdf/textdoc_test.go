package pdf

import (
	"bytes"
	"testing"
)

// fakeFace is a minimal, synthetic EmbeddedFace for exercising
// SerializeTextDocument's structure without needing a real subset —
// internal/fontset's own tests (fontset_test.go) already exercise real
// subsetting; this file's job is the PDF assembly layer alone.
func fakeFace(name string) EmbeddedFace {
	return EmbeddedFace{
		Name:      name,
		Program:   []byte("FAKE-TRUETYPE-PROGRAM-BYTES-" + name),
		Tag:       "ABCDEF",
		NumGlyphs: 2,
		GlyphForRune: map[rune]uint16{
			'A': 1,
			'B': 2,
		},
		WidthForGlyph: map[uint16]int64{0: 0, 1: 500, 2: 600},
	}
}

func onePageWithText(face string) TextPage {
	return TextPage{
		Runs: []TextRun{
			{Face: face, Text: "AB", X: 0, Y: 0, FontSize: 12000},
		},
		Width:      595276,
		Height:     841890,
		MarginTop:  36000,
		MarginLeft: 36000,
	}
}

// TestSerializeTextDocumentEmbedsOneFontFileForTwoPages is AC9,
// asserted BEHAVIOURALLY: "a two-page document using one face embeds
// exactly one FontFile2." internal/pdf has no notion of "the same
// template rendered across multiple pages" (no internal/layout exists
// until Story 2.5, so real pagination is out of this story's scope) —
// this test exercises the mechanism this property actually depends on:
// SerializeTextDocument receives the resolved face set ONCE and embeds
// each entry exactly once regardless of how many TextPages reference
// it, which is the "never per page" half of AC9. The "one subsetting
// call" half — that internal/fontset.Font.Subset is called exactly
// once per face, over the document-wide union of runes, not per
// element — is asserted behaviourally at the folio package level by
// TestRenderSubsetsOneFacePerDocumentNotPerElement (render_test.go),
// which renders a real two-element, two-band document through the
// public Render path (Finding 6, QA review: the previous version of
// this comment claimed that half was "asserted via code review", which
// is not an assertion; it never reached renderDocument at all — the
// two-PAGE half genuinely remains out of reach until Story 2.5's
// pagination exists, and stays deferred here on that basis, honestly).
func TestSerializeTextDocumentEmbedsOneFontFileForTwoPages(t *testing.T) {
	faces := map[string]EmbeddedFace{"Body": fakeFace("Body")}
	pages := []TextPage{onePageWithText("Body"), onePageWithText("Body")}

	out, err := SerializeTextDocument(pages, faces)
	if err != nil {
		t.Fatalf("SerializeTextDocument: %v", err)
	}

	count := bytes.Count(out, []byte("/FontFile2 "))
	if count != 1 {
		t.Fatalf("two-page document using one face embeds %d FontFile2 references, want exactly 1 (AC9)", count)
	}
	if bytes.Count(out, []byte("/Type /Page")) < 2 {
		t.Fatalf("expected at least two /Type /Page objects (a two-page document), got fewer — test setup is broken")
	}
	if !bytes.Contains(out, []byte(fakeFace("Body").Program)) {
		t.Fatal("output does not contain the face's program bytes at all")
	}
	// The program bytes must appear EXACTLY ONCE — embedding it per page
	// would duplicate the (potentially large) font program, defeating
	// AC9's whole point.
	if got := bytes.Count(out, []byte(fakeFace("Body").Program)); got != 1 {
		t.Fatalf("face program bytes appear %d times, want exactly 1 (AC9: never per page)", got)
	}
}

// TestSerializeTextDocumentMissingFaceIsLocatedError exercises the
// content-stream builder's own guard: a run naming a face absent from
// the face set is a located error, never a silently-emitted .notdef
// stream.
func TestSerializeTextDocumentMissingFaceIsLocatedError(t *testing.T) {
	pages := []TextPage{onePageWithText("DoesNotExist")}
	_, err := SerializeTextDocument(pages, map[string]EmbeddedFace{})
	if err == nil {
		t.Fatal("expected an error for a run naming a face absent from the face set")
	}
}

// TestSerializeTextDocumentUnmappedRuneIsLocatedError exercises the
// per-rune guard in appendHexCIDString: a rune with no entry in
// GlyphForRune is a located error.
func TestSerializeTextDocumentUnmappedRuneIsLocatedError(t *testing.T) {
	face := fakeFace("Body")
	// "Z" is not in fakeFace's GlyphForRune map.
	pages := []TextPage{{
		Runs:   []TextRun{{Face: "Body", Text: "Z", X: 0, Y: 0, FontSize: 12000}},
		Width:  595276,
		Height: 841890,
	}}
	_, err := SerializeTextDocument(pages, map[string]EmbeddedFace{"Body": face})
	if err == nil {
		t.Fatal("expected an error for a rune absent from the face's GlyphForRune map")
	}
}

// TestAppendXrefGeneralHandlesArbitraryObjectCounts is a basic sanity
// check on the generalised xref writer (document.go's appendXref stays
// fixed at exactly Story 1.1's four objects, untouched per AC14a).
func TestAppendXrefGeneralHandlesArbitraryObjectCounts(t *testing.T) {
	offsets := []int{0, 10, 200, 3000, 40000, 500000, 6000000}
	out := appendXrefGeneral(nil, offsets)
	if !bytes.HasPrefix(out, []byte("xref\n0 7\n")) {
		t.Fatalf("unexpected xref header: %q", out[:20])
	}
	lines := bytes.Split(bytes.TrimSuffix(out[len("xref\n0 7\n"):], nil), []byte("\n"))
	// 7 offsets -> 7 entries, each 20 bytes (19 chars + trailing \n via Split).
	if len(lines) < 7 {
		t.Fatalf("expected at least 7 xref entry lines, got %d", len(lines))
	}
}
