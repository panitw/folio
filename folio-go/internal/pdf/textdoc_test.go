package pdf

import "github.com/panitw/folio/folio-go/internal/pagemodel"

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
		Name:          name,
		Program:       []byte("FAKE-TRUETYPE-PROGRAM-BYTES-" + name),
		Tag:           "ABCDEF",
		NumGlyphs:     3,
		WidthForGlyph: map[uint16]int64{0: 0, 1: 500, 2: 600},
		ToUnicode: []CIDText{
			{CID: 1, Text: "A"},
			{CID: 2, Text: "B"},
		},
	}
}

func onePageWithText(face string) pagemodel.Page {
	return pagemodel.Page{
		Runs: []pagemodel.TextRun{
			{Face: face, SourceText: "AB", X: 0, Y: 0, FontSize: 12000, Glyphs: []pagemodel.ShapedGlyph{
				{CID: 1, XAdvance: 500},
				{CID: 2, XAdvance: 600},
			}},
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
// each entry exactly once regardless of how many pagemodel.Pages reference
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
	pages := []pagemodel.Page{onePageWithText("Body"), onePageWithText("Body")}

	out, err := SerializeTextDocument(pages, faces, nil, nil)
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
	pages := []pagemodel.Page{onePageWithText("DoesNotExist")}
	_, err := SerializeTextDocument(pages, map[string]EmbeddedFace{}, nil, nil)
	if err == nil {
		t.Fatal("expected an error for a run naming a face absent from the face set")
	}
}

// TestSerializeTextDocumentUnknownCIDIsLocatedError exercises the
// per-glyph guard in appendShapedRun: a CID with no width in the
// embedded face is a located error, never a silently-emitted .notdef.
// Story 2.3 re-keyed this from runes onto CIDs, because there is no
// rune -> CID route into a content stream any more (AC4).
func TestSerializeTextDocumentUnknownCIDIsLocatedError(t *testing.T) {
	face := fakeFace("Body")
	// CID 9 is beyond fakeFace's base block and its (empty) extras.
	pages := []pagemodel.Page{{
		Runs: []pagemodel.TextRun{{Face: "Body", SourceText: "Z", X: 0, Y: 0, FontSize: 12000, Glyphs: []pagemodel.ShapedGlyph{
			{CID: 9, XAdvance: 500},
		}}},
		Width:  595276,
		Height: 841890,
	}}
	_, err := SerializeTextDocument(pages, map[string]EmbeddedFace{"Body": face}, nil, nil)
	if err == nil {
		t.Fatal("expected an error for a CID with no width in the embedded face")
	}
}

// TestShapedRunFailsClosedOnYOffset is AC6's fail-closed branch, tested
// DIRECTLY — which is still worth doing, but no longer because the
// branch is unreachable any other way.
//
// This comment used to say: "Measured across all three shipped faces at
// Story 2.3: YOffset is 0 for every glyph of every sample, so no
// production input triggers this and the branch has NO available
// red-proof through a rendered document." That was false. Story 2.3
// measured its own samples and reported on the shipped set, and
// ordinary Thai stacking two marks over one base reaches this branch
// through the public entry point on the shipped Noto Sans Thai — see
// thai_mark_stacking_test.go, which pins the message an author actually
// receives and holds the branch to a real document (DW-28).
//
// The synthetic run keeps its own value: it exercises the branch
// without depending on a face, so it survives any change to the shipped
// set. What it must no longer be credited with is being the only
// possible proof.
func TestShapedRunFailsClosedOnYOffset(t *testing.T) {
	face := fakeFace("Body")
	run := pagemodel.TextRun{Face: "Body", SourceText: "A", FontSize: 12000, Glyphs: []pagemodel.ShapedGlyph{
		{CID: 1, XAdvance: 500, YOffset: 37},
	}}
	if _, err := appendShapedRun(nil, run, face); err == nil {
		t.Fatal("a glyph carrying a non-zero YOffset must be a located error, not a silently dropped offset")
	}

	// ...and the same run with YOffset zeroed must succeed, so the test
	// above cannot pass for an unrelated reason.
	run.Glyphs[0].YOffset = 0
	if _, err := appendShapedRun(nil, run, face); err != nil {
		t.Fatalf("the same run with YOffset 0 must emit cleanly, got: %v", err)
	}
}

// TestZeroAdjustmentRunEmitsTj is AC6/AC14's compatibility clause,
// asserted on the emitted bytes rather than assumed: a run in which
// every adjustment computes to zero emits a bare Tj hex string —
// exactly the bytes folio emitted before Story 2.3 — and a run carrying
// a real offset or a kerned advance emits a TJ array instead.
func TestZeroAdjustmentRunEmitsTj(t *testing.T) {
	face := fakeFace("Body")

	plain := pagemodel.TextRun{Face: "Body", SourceText: "AB", FontSize: 12000, Glyphs: []pagemodel.ShapedGlyph{
		{CID: 1, XAdvance: 500},
		{CID: 2, XAdvance: 600},
	}}
	got, err := appendShapedRun(nil, plain, face)
	if err != nil {
		t.Fatalf("appendShapedRun: %v", err)
	}
	if string(got) != "<00010002> Tj\n" {
		t.Fatalf("a zero-adjustment run must emit today's exact Tj bytes, got %q", got)
	}

	kerned := pagemodel.TextRun{Face: "Body", SourceText: "AB", FontSize: 12000, Glyphs: []pagemodel.ShapedGlyph{
		{CID: 1, XAdvance: 460}, // /W says 500: GPOS kerned it by -40
		{CID: 2, XAdvance: 600},
	}}
	got, err = appendShapedRun(nil, kerned, face)
	if err != nil {
		t.Fatalf("appendShapedRun (kerned): %v", err)
	}
	if !bytes.HasPrefix(got, []byte("[<0001>40<0002>")) {
		t.Fatalf("a kerned run must emit a TJ array carrying the +40 adjustment, got %q", got)
	}

	offset := pagemodel.TextRun{Face: "Body", SourceText: "AB", FontSize: 12000, Glyphs: []pagemodel.ShapedGlyph{
		{CID: 1, XAdvance: 500},
		{CID: 2, XAdvance: 600, XOffset: -29},
	}}
	got, err = appendShapedRun(nil, offset, face)
	if err != nil {
		t.Fatalf("appendShapedRun (offset): %v", err)
	}
	if !bytes.Contains(got, []byte("29<0002>")) {
		t.Fatalf("an x-offset run must emit the +29 pre-glyph adjustment, got %q", got)
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

// TestAssetKeyEscapeIsIdentity is AC18a: asset keys are 64 lowercase hex
// characters, a strict subset of pdfNameEscape's kept set
// ([A-Za-z0-9_-]), so escaping is the identity and distinct keys stay
// distinct — asserted directly rather than left to be re-derived (the
// font path's resource-name collision hazard, Story 1.5 QA Finding 23,
// cannot arise here).
func TestAssetKeyEscapeIsIdentity(t *testing.T) {
	key := "5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078"
	if got := pdfNameEscape(key); got != key {
		t.Fatalf("pdfNameEscape(%q) = %q, want the identity (AC18a)", key, got)
	}
}

// TestColouredRunBracketsItsInk is Story 10.1's byte-level assertion: a
// run carrying an ink emits a fill-colour operator inside its own q/Q
// bracket, so the colour cannot leak into whatever draws next; a run
// carrying none emits no colour operator at all, which is what leaves
// every document written before style.color existed byte-identical.
func TestColouredRunBracketsItsInk(t *testing.T) {
	faces := map[string]EmbeddedFace{"Body": fakeFace("Body")}
	run := pagemodel.TextRun{Face: "Body", SourceText: "A", FontSize: 12000, X: 1000, Y: 2000, Glyphs: []pagemodel.ShapedGlyph{{CID: 1, XAdvance: 500}}}

	plain, err := buildTextContentStream(pagemodel.Page{Width: 600000, Height: 800000, Runs: []pagemodel.TextRun{run}}, faces)
	if err != nil {
		t.Fatalf("buildTextContentStream: %v", err)
	}
	if bytes.Contains(plain, []byte(" rg")) || bytes.Contains(plain, []byte("q\n")) {
		t.Errorf("an uncoloured run emitted a colour operator or a bracket: %q", plain)
	}

	run.HasColor, run.Color = true, pagemodel.Color{R: 255, G: 0, B: 0}
	coloured, err := buildTextContentStream(pagemodel.Page{Width: 600000, Height: 800000, Runs: []pagemodel.TextRun{run}}, faces)
	if err != nil {
		t.Fatalf("buildTextContentStream (coloured): %v", err)
	}
	if !bytes.HasPrefix(coloured, []byte("q\n1 0 0 rg\nBT\n")) {
		t.Errorf("a coloured run must set its fill inside its own bracket, before BT, got %q", coloured)
	}
	if !bytes.HasSuffix(coloured, []byte("ET\nQ\n")) {
		t.Errorf("a coloured run must close its bracket after ET, got %q", coloured)
	}
	// The ONE difference between the two streams is the bracket and the
	// operator: the text itself is emitted identically.
	if got := bytes.ReplaceAll(bytes.ReplaceAll(coloured, []byte("q\n1 0 0 rg\n"), nil), []byte("ET\nQ\n"), []byte("ET\n")); !bytes.Equal(got, plain) {
		t.Errorf("colouring a run changed more than its ink:\n coloured-minus-ink %q\n plain              %q", got, plain)
	}
}
