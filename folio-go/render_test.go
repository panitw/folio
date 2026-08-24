package folio

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/fontset"
	"github.com/panitw/folio/folio-go/internal/pdf"
)

// subprocessEnvVar, when set to "1", makes TestMain render the FONTLESS
// document to stdout and exit instead of running the test suite. This is
// how the determinism test in this file re-executes the test binary as a
// fresh OS process rather than comparing two calls inside one process (a
// same-process comparison would pass on shared memoised state, which is
// exactly what AC8 rules out).
//
// This selector keeps rendering Story 1.1's fixed rectangle via
// internal/pdf.Serialize() directly, NOT via Render — Render now
// requires a non-nil template (AC14b), and fixtures/minimal-rect/ is
// reclassified (AC14a, D-1.5.9) as pinning internal/pdf's fontless
// emission path specifically. AC10a: "the subprocess seam KEEPS the
// fontless path" — so fixtures/minimal-rect/ does not lose its
// four-target verification.
const subprocessEnvVar = "FOLIO_SUBPROCESS_RENDER"

// subprocessFontEnvVar is AC10a's SECOND selector: it renders the
// template+font document (fontTestTemplateJSON, testRobotoFontBytes)
// through the public Render path and writes the bytes to stdout. The
// cross-target matrix (matrix_test.go) drives both selectors and
// verifies AC10b's vacuity guard before comparing any hash: the child's
// output must actually contain a FontFile2, or the whole exercise
// certifies nothing (F-6).
const subprocessFontEnvVar = "FOLIO_SUBPROCESS_RENDER_FONT"

// subprocessImageEnvVar is AC22's THIRD selector (D-1.8.6, joining the
// two above — it replaces neither): it renders imageTestTemplateJSON
// (an embedded, real, supported PNG asset drawn by one image element)
// through the public Render path and writes the bytes to stdout. The
// matrix harness (matrix_test.go) verifies AC23's vacuity guard — the
// child's output must actually contain an image XObject
// (/Subtype /Image) — before comparing any hash, the same shape AC10b
// already established for the font path.
const subprocessImageEnvVar = "FOLIO_SUBPROCESS_RENDER_IMAGE"

// subprocessMultiScriptEnvVar is Story 2.2's FOURTH selector (AC8,
// joining the three above — it replaces none of them): it renders
// multiScriptTestTemplateJSON, a document whose ONE text element mixes
// Latin, Thai and CJK runes against a three-member fallback chain,
// through the public Render path, using the REAL shipped face set
// (`github.com/panitw/folio/folio-go/fonts`.Shipped()) rather than a
// synthetic FontSet — this is the fixture that actually exercises
// AC7's per-(face,pinned-instance) goldens across all four targets. The
// matrix harness verifies its OWN feature guard on every leg (AC8,
// V6): the captured stream must contain THREE distinct, INSTANCED
// (fvar-free) embedded programs, not merely "a FontFile2" — the same
// reasoning AC10b/AC23 already apply to the font and image paths.
const subprocessMultiScriptEnvVar = "FOLIO_SUBPROCESS_RENDER_MULTISCRIPT"

// multiScriptTestTemplateJSON is Story 2.2's AC8 fixture document: one
// text element, one three-member fallback chain naming all three
// shipped faces in AD-8's declared order (Latin, Thai, CJK), and a
// value mixing all three scripts in a single run — "Ada" (Latin,
// resolves to Noto Sans), "ก" (Thai, resolves to Noto Sans Thai, since
// Noto Sans has no glyph for it) and "汉" (a Han ideograph, resolves to
// Noto Sans SC). This is D-1.8.6's precedent: a NEW document, added
// beside fontTestTemplateJSON/imageTestTemplateJSON, never replacing
// either (D-2.2-D5).
const multiScriptTestTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 500, "height": 20, "value": "Ada ก 汉", "style": {"fontFamily": "body", "fontSize": 14}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {
    "body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"]
  },
  "locale": "en",
  "nextId": 2,
  "page": {
    "margin": {"bottom": 36, "left": 36, "right": 36, "top": 36},
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// subprocessShapedTextEnvVar is Story 2.3's FIFTH selector (joining the
// four above — it replaces none of them): it renders
// shapedTextTemplateJSON, the document whose text was chosen
// specifically so that a WRONG shaping answer is a different hash,
// through the public Render path against the real shipped face set.
//
// AC11 is why it exists at all. Every pre-2.3 golden was measured to be
// blind to shaping: its text shapes to itself, so a correct
// implementation and one that never calls the shaper produce identical
// bytes. A fifth fixture recorded over equally blind text would look
// like coverage in every report while being unable to fail when this
// story regresses — D-000.9's exact shape.
const subprocessShapedTextEnvVar = "FOLIO_SUBPROCESS_RENDER_SHAPEDTEXT"

// subprocessWrappedTextEnvVar is Story 2.4's SIXTH selector (joining the
// five above, replacing none): fixtures/wrapped-text/, the line-breaking
// document.
//
// It needs its own selector for the same reason shaped-text did. Every
// earlier fixture FITS its box — measured, all ten text elements, the
// widest at 26% of its width — so a matrix comparing four legs of any of
// them would agree perfectly while certifying nothing about wrapping: a
// build that never broke a line would produce identical bytes. This is
// the only registered document whose elements are measured to overflow
// their boxes, and therefore the only one whose legs can disagree if
// line breaking is not reproducible across targets.
const subprocessWrappedTextEnvVar = "FOLIO_SUBPROCESS_RENDER_WRAPPEDTEXT"

// shapedTextTemplateJSON is Story 2.3's AC10 fixture document
// (fixtures/shaped-text/input.folio, kept byte-identical to it by
// TestShapedTextGoldenFixture). Every element is chosen for what it
// makes observable:
//
//	e1  Thai tall consonants with marks — GSUB selects a LOWERED mark
//	    form; the naive answer collides with the ascender. THE defect.
//	e2  A real Thai name, plus two plain-Thai negative controls.
//	e3  "น้ำ า" — a merged SARA AM cluster AND a standalone SARA AA in
//	    ONE document. Both end in the same glyph and need different
//	    /ToUnicode answers, which is what forces CIDs to be allocated
//	    per (glyph, text) rather than per glyph.
//	e4  Latin ligatures (ffi, fi) and kern pairs (AV, Wo, To).
//	e5  CJK, the negative control: identical glyphs, all advances 1000.
//	e6  A mixed run, so shaping is exercised per FACE-SEGMENT (AC8) —
//	    the fallback chain still decides the face, shaping runs inside
//	    what it chose, never across a boundary.
//	e7  "AV ก" — a KERNING Latin segment followed by another face's
//	    segment. Added by Story 2.3's finisher (Blocker 1). e4 is
//	    shape-observable but is the LAST segment of its element, so its
//	    cursor is never consumed; e6 is a mixed run but "Ada " carries no
//	    kern, so it agreed by accident. Between them the golden was green
//	    over a live defect — text drawn kerned and placed unkerned. e7 is
//	    the case neither could produce: the Thai segment's origin moves by
//	    the A/V kern, so a regression to unkerned segment cursors is a
//	    different hash.
const shapedTextTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 500, "height": 24, "value": "ปั ฟั ที่ ป้ำ", "style": {"fontFamily": "body", "fontSize": 16}},
        {"id": "e2", "type": "text", "x": 0, "y": 28, "width": 500, "height": 24, "value": "ณัฐวุฒิ เกิด กรุงเทพ", "style": {"fontFamily": "body", "fontSize": 16}},
        {"id": "e3", "type": "text", "x": 0, "y": 56, "width": 500, "height": 24, "value": "น้ำ า", "style": {"fontFamily": "body", "fontSize": 16}},
        {"id": "e4", "type": "text", "x": 0, "y": 84, "width": 500, "height": 24, "value": "office fi AV Wo. To,", "style": {"fontFamily": "body", "fontSize": 16}},
        {"id": "e5", "type": "text", "x": 0, "y": 112, "width": 500, "height": 24, "value": "结算单，共３页", "style": {"fontFamily": "body", "fontSize": 16}},
        {"id": "e6", "type": "text", "x": 0, "y": 140, "width": 500, "height": 24, "value": "Ada ปั 结", "style": {"fontFamily": "body", "fontSize": 16}},
        {"id": "e7", "type": "text", "x": 0, "y": 168, "width": 500, "height": 24, "value": "AV ก", "style": {"fontFamily": "body", "fontSize": 16}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {
    "body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"]
  },
  "locale": "th",
  "nextId": 8,
  "page": {
    "margin": {"bottom": 36, "left": 36, "right": 36, "top": 36},
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// imageTestAssetKey is the SHA-256 (lowercase hex) of imageTestPNGBytes
// below — generated by Go's image/png encoder and independently
// cross-checked with Python's hashlib/base64 during story creation.
const imageTestAssetKey = "5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078"

// imageTestTemplateJSON is a `.folio` document with one image element
// in content, referencing a real, supported, non-square 3x2 8-bit RGB
// PNG asset (AC27a: non-square, so it exercises AC13's binding-axis
// choice) — its data is the canonical 76-column base64 split of
// imageTestPNGBytes.
const imageTestTemplateJSON = `{
  "assets": {
    "` + imageTestAssetKey + `": {
      "data": [
        "iVBORw0KGgoAAAANSUhEUgAAAAMAAAACCAIAAAASFvFNAAAAGElEQVR42mL6z8DAAMZMEOo/AwMg",
        "AAD//zwUBf/NjsW5AAAAAElFTkSuQmCC"
      ],
      "mediaType": "image/png"
    }
  },
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "image", "asset": "` + imageTestAssetKey + `", "x": 0, "y": 0, "width": 100, "height": 60}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {},
  "locale": "en",
  "nextId": 2,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// fontTestTemplateJSON is a `.folio` document with one text element in
// each of pageHeader, content and pageFooter, all resolving to the same
// face via one fallback chain ("body" -> ["Roboto-Regular"]) — enough to
// exercise AC9's "union of glyphs the whole document uses, collected
// once" across more than one element and more than one band.
const fontTestTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "Hello, World!", "style": {"fontFamily": "body", "fontSize": 14}}
      ]
    },
    "pageFooter": {
      "elements": [
        {"id": "e2", "type": "text", "x": 0, "y": 0, "width": 400, "height": 16, "value": "Page footer 0123456789", "style": {"fontFamily": "body", "fontSize": 9}}
      ],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {
    "body": ["Roboto-Regular"]
  },
  "locale": "en",
  "nextId": 3,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// testFontSet returns a FontSet carrying Story 1.5's one small Latin
// test face (AC26), keyed to match fontTestTemplateJSON's fonts chain.
func testFontSet() FontSet {
	return FontSet{"Roboto-Regular": testRobotoFontBytes}
}

// parseFontTestTemplate parses fontTestTemplateJSON, failing the test on
// any parse error.
func parseFontTestTemplate(t *testing.T) *Template {
	t.Helper()
	tpl, err := ParseTemplate([]byte(fontTestTemplateJSON))
	if err != nil {
		t.Fatalf("parse fontTestTemplateJSON: %v", err)
	}
	return tpl
}

// TestRenderSubsetsOneFacePerDocumentNotPerElement is Finding 6's fix
// (QA review): AC9's "one subsetting call" half was previously asserted
// only "via code review" in a synthetic textdoc_test.go fixture that
// never reached renderDocument. This test exercises the REAL public
// path — Render, on fontTestTemplateJSON, whose content and pageFooter
// bands each carry a text element referencing the SAME face ("body" ->
// "Roboto-Regular") — and asserts the property behaviourally two ways:
//
//  1. Exactly one /FontFile2 stream in the output, even though two
//     elements in two different bands reference the face.
//  2. The embedded subset's tag is the DOCUMENT-WIDE union closure's
//     tag, not either element's own smaller, per-element closure — a
//     regression that moved subsetting inside the per-run loop (the
//     exact hazard AC9 names) would still embed exactly one FontFile2
//     per RUN, so check 1 alone would not catch it; the tag comparison
//     does, because a per-element subset's closure (and therefore its
//     tag) differs from the union's.
func TestRenderSubsetsOneFacePerDocumentNotPerElement(t *testing.T) {
	tpl := parseFontTestTemplate(t)
	fs := testFontSet()

	out, err := Render(tpl, Data("{}"), nil, fs)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !containsFontFile2(out) {
		t.Fatal("rendered document has no /FontFile2 — vacuity guard failed before this test's real assertions could mean anything (AC10b)")
	}

	n := bytes.Count(out, []byte("/FontFile2 "))
	if n != 1 {
		t.Fatalf("AC9: expected exactly one /FontFile2 reference for one face used by two elements in two bands, got %d", n)
	}

	font, ferr := fontset.New("Roboto-Regular", testRobotoFontBytes)
	if ferr != nil {
		t.Fatalf("fontset.New: %v", ferr)
	}
	// Story 2.3, AC5: the subset is keyed on the glyphs the renderer
	// DRAWS, so the comparison sets are built by shaping each element's
	// text, exactly as renderDocument does.
	elem1Sub, err := font.Subset(shapedGlyphIDsForTest(t, font, "Hello, World!"))
	if err != nil {
		t.Fatalf("Subset(elem1 alone): %v", err)
	}
	elem2Sub, err := font.Subset(shapedGlyphIDsForTest(t, font, "Page footer 0123456789"))
	if err != nil {
		t.Fatalf("Subset(elem2 alone): %v", err)
	}
	unionSub, err := font.Subset(shapedGlyphIDsForTest(t, font, "Hello, World!Page footer 0123456789"))
	if err != nil {
		t.Fatalf("Subset(union): %v", err)
	}
	if elem1Sub.Tag == unionSub.Tag || elem2Sub.Tag == unionSub.Tag {
		t.Fatalf(
			"test fixture assumption violated: a per-element subset's tag must differ from the union's for this "+
				"assertion to be discriminating (elem1=%q elem2=%q union=%q)",
			elem1Sub.Tag, elem2Sub.Tag, unionSub.Tag,
		)
	}

	baseFontUnion := "/BaseFont /" + unionSub.Tag + "+"
	if !bytes.Contains(out, []byte(baseFontUnion)) {
		t.Fatalf("rendered document's embedded /BaseFont does not carry the document-wide union tag %q — AC9's "+
			"'one subsetting call per document, over the union of glyphs the whole document uses' is violated", unionSub.Tag)
	}
	for _, perElement := range []struct {
		label string
		sub   *fontset.Subset
	}{{"elem1", elem1Sub}, {"elem2", elem2Sub}} {
		needle := "/BaseFont /" + perElement.sub.Tag + "+"
		if bytes.Contains(out, []byte(needle)) {
			t.Fatalf("rendered document's embedded /BaseFont carries %s's OWN per-element tag %q instead of the "+
				"document-wide union tag — subsetting is happening per element, not once per document (AC9)",
				perElement.label, perElement.sub.Tag)
		}
	}
}

// shapedGlyphIDsForTest returns the source glyph ids font draws for
// text, deduplicated by first occurrence — the input shape Subset takes
// after Story 2.3 re-keyed it off runes (AC5).
func shapedGlyphIDsForTest(t *testing.T, font *fontset.Font, text string) []uint16 {
	t.Helper()
	glyphs, err := font.Shaper().Shape(text)
	if err != nil {
		t.Fatalf("Shape(%q): %v", text, err)
	}
	if len(glyphs) == 0 {
		t.Fatalf("Shape(%q) produced no glyphs", text)
	}
	seen := map[uint16]bool{}
	var out []uint16
	for _, g := range glyphs {
		if seen[g.GlyphID] {
			continue
		}
		seen[g.GlyphID] = true
		out = append(out, g.GlyphID)
	}
	return out
}

// containsFontFile2 is AC10b's binding vacuity guard, applied wherever
// this file or matrix_test.go needs to confirm a rendered document
// actually embeds a font before comparing any bytes: "the child render
// on the font path must be asserted to actually contain a FontFile2
// before any byte comparison runs... same output ⇒ instance" (D-000.9).
func containsFontFile2(b []byte) bool {
	return bytes.Contains(b, []byte("/FontFile2"))
}

// containsImageXObject is AC23's binding vacuity guard (D-1.8.6), the
// image-path mirror of containsFontFile2 above: a rendered document
// must be confirmed to actually contain an image XObject
// (/Subtype /Image) before any byte comparison runs.
func containsImageXObject(b []byte) bool {
	return bytes.Contains(b, []byte("/Subtype /Image"))
}

// toolchainEnvVar, when set to "1", makes TestMain write the toolchain
// that built this binary to stdout and exit, instead of running the test
// suite. Story 1.2's matrix harness (AC3, D-1.2.4) uses this to witness
// each target binary's toolchain before comparing any hash: the host's
// `go version` describes the machine, not the artifact, and in CI the
// four legs are built on three different runners — only
// runtime.Version() compiled into the binary itself is an honest
// witness. Same seam as subprocessEnvVar, one more env-gated mode, both
// exiting before m.Run() is ever reached so neither touches os/exec or
// pipes — which is what keeps this branch wasm-safe.
const toolchainEnvVar = "FOLIO_SUBPROCESS_TOOLCHAIN"

func TestMain(m *testing.M) {
	if os.Getenv(toolchainEnvVar) == "1" {
		os.Stdout.WriteString(runtime.Version())
		os.Exit(0)
	}
	if os.Getenv(subprocessEnvVar) == "1" {
		writeToStdoutOrDie(pdf.Serialize())
	}
	if os.Getenv(subprocessFontEnvVar) == "1" {
		tpl, err := ParseTemplate([]byte(fontTestTemplateJSON))
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		b, err := Render(tpl, Data("{}"), nil, testFontSet())
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		writeToStdoutOrDie(b)
	}
	if os.Getenv(subprocessImageEnvVar) == "1" {
		tpl, err := ParseTemplate([]byte(imageTestTemplateJSON))
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		b, err := Render(tpl, Data("{}"), nil, nil)
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		writeToStdoutOrDie(b)
	}
	if os.Getenv(subprocessShapedTextEnvVar) == "1" {
		tpl, err := ParseTemplate([]byte(shapedTextTemplateJSON))
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		b, err := Render(tpl, Data("{}"), nil, testShippedFontSet())
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		writeToStdoutOrDie(b)
	}
	if os.Getenv(subprocessWrappedTextEnvVar) == "1" {
		tpl, err := ParseTemplate([]byte(wrappedTextTemplateJSON))
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		b, err := Render(tpl, Data(wrappedTextDataJSON), nil, testShippedFontSet())
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		writeToStdoutOrDie(b)
	}
	if os.Getenv(subprocessMultiScriptEnvVar) == "1" {
		tpl, err := ParseTemplate([]byte(multiScriptTestTemplateJSON))
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		b, err := Render(tpl, Data("{}"), nil, testShippedFontSet())
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		writeToStdoutOrDie(b)
	}
	os.Exit(m.Run())
}

// writeToStdoutOrDie writes b to stdout in full and exits 0, or reports a
// short/failed write and exits 1. A short write here would otherwise be
// indistinguishable from a real determinism failure: the parent process
// would see truncated bytes and report a byte-offset divergence that
// points at the renderer, when the actual fault is this pipe (this
// story's QA review, Nit 25).
func writeToStdoutOrDie(b []byte) {
	n, werr := os.Stdout.Write(b)
	if werr != nil {
		os.Stderr.WriteString("write to stdout: " + werr.Error())
		os.Exit(1)
	}
	if n != len(b) {
		os.Stderr.WriteString("short write to stdout: wrote " + strconv.Itoa(n) + " of " + strconv.Itoa(len(b)) + " bytes")
		os.Exit(1)
	}
	os.Exit(0)
}

// assertWellFormedPDF performs toolchain-independent structural
// validation of the classic (non-stream) PDF this story's document is:
// it re-derives the document's self-described structure from its own
// bytes — no PDF-reading dependency, consistent with AC2 — and checks
// internal consistency, rather than checking for a handful of literal
// substrings.
//
// This story's QA review measured that the previous version (four
// checks: non-empty, header prefix, %%EOF suffix, exactly one
// "/Type /Page " substring) accepted a 118-byte stub with no catalog, no
// page tree, no content stream, no xref table and no trailer — every test
// that used it, including AC5's own validity test, passed on that stub
// (Major 4). It is also this function that AC8's vacuity guard 1 depends
// on to keep two identical failures from comparing equal, and — with
// Major 4's fix — it now doubles as Minor 18's PDF self-consistency
// assertions and Nit 26's fix for a page-count check that used to depend
// on an incidental trailing space rather than an actual dictionary
// boundary.
func assertWellFormedPDF(t *testing.T, label string, b []byte) {
	t.Helper()

	if len(b) == 0 {
		t.Fatalf("%s: output is empty", label)
	}
	if !bytes.HasPrefix(b, []byte("%PDF-1.7")) {
		t.Fatalf("%s: output does not start with %%PDF-1.7", label)
	}
	if !bytes.HasSuffix(b, []byte("%%EOF\n")) {
		t.Fatalf("%s: output does not end with %%%%EOF", label)
	}

	pageCount := countPageObjects(b)
	if pageCount != 1 {
		t.Fatalf("%s: expected exactly one /Type /Page object, found %d", label, pageCount)
	}

	if !bytes.Contains(b, []byte("/Type /Catalog")) {
		t.Fatalf("%s: no /Type /Catalog object found", label)
	}

	xrefOffset := assertStartxrefPointsAtXref(t, label, b)
	assertXrefEntriesPointAtTheirObjects(t, label, b, xrefOffset)
	assertStreamLengthsAreExact(t, label, b)
}

// countPageObjects counts occurrences of "/Type /Page" that are not the
// start of "/Type /Pages" — a substring match alone (as the previous
// version of this file relied on) cannot tell "/Type /Page " from
// "/Type /Page>>" apart from "/Type /Pages", except by accident of a
// trailing space (Nit 26).
func countPageObjects(b []byte) int {
	needle := []byte("/Type /Page")
	count := 0
	idx := 0
	for {
		rel := bytes.Index(b[idx:], needle)
		if rel == -1 {
			return count
		}
		pos := idx + rel
		end := pos + len(needle)
		if end < len(b) && b[end] == 's' { // "/Type /Pages", not a page object
			idx = end
			continue
		}
		count++
		idx = end
	}
}

// assertStartxrefPointsAtXref parses "startxref\n<offset>\n" and checks
// that offset genuinely indexes to the start of the "xref" keyword, and
// returns that offset.
func assertStartxrefPointsAtXref(t *testing.T, label string, b []byte) int {
	t.Helper()

	const kw = "startxref\n"
	six := bytes.LastIndex(b, []byte(kw))
	if six == -1 {
		t.Fatalf("%s: no startxref keyword found", label)
	}
	rest := b[six+len(kw):]
	eol := bytes.IndexByte(rest, '\n')
	if eol == -1 {
		t.Fatalf("%s: startxref value is not newline-terminated", label)
	}
	offsetStr := string(rest[:eol])
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		t.Fatalf("%s: startxref value %q is not an integer: %v", label, offsetStr, err)
	}
	if offset < 0 || offset >= len(b) || !bytes.HasPrefix(b[offset:], []byte("xref")) {
		t.Fatalf("%s: startxref offset %d does not point at the xref keyword", label, offset)
	}
	return offset
}

// assertXrefEntriesPointAtTheirObjects parses the classic xref table
// starting at xrefOffset and checks that every in-use entry's offset
// genuinely lands on "N 0 obj" for its own object number.
func assertXrefEntriesPointAtTheirObjects(t *testing.T, label string, b []byte, xrefOffset int) {
	t.Helper()

	const xrefKW = "xref\n"
	xrefBody := b[xrefOffset:]
	if !bytes.HasPrefix(xrefBody, []byte(xrefKW)) {
		t.Fatalf("%s: xref section does not start with %q", label, xrefKW)
	}
	xrefBody = xrefBody[len(xrefKW):]

	headerEnd := bytes.IndexByte(xrefBody, '\n')
	if headerEnd == -1 {
		t.Fatalf("%s: xref subsection header is not newline-terminated", label)
	}
	fields := strings.Fields(string(xrefBody[:headerEnd]))
	if len(fields) != 2 {
		t.Fatalf("%s: xref subsection header %q does not have two fields", label, xrefBody[:headerEnd])
	}
	start, err1 := strconv.Atoi(fields[0])
	count, err2 := strconv.Atoi(fields[1])
	if err1 != nil || err2 != nil || start != 0 || count < 1 {
		t.Fatalf("%s: unexpected xref subsection header %q", label, xrefBody[:headerEnd])
	}

	entries := xrefBody[headerEnd+1:]
	for i := 0; i < count; i++ {
		entryStart := i * 20
		if entryStart+20 > len(entries) {
			t.Fatalf("%s: xref entry %d is truncated", label, i)
		}
		entry := entries[entryStart : entryStart+20]
		if entry[10] != ' ' || entry[16] != ' ' || entry[18] != ' ' || entry[19] != '\n' {
			t.Fatalf("%s: xref entry %d %q is not exactly 20 bytes in the expected shape", label, i, entry)
		}
		offStr := string(entry[0:10])
		genStr := string(entry[11:16])
		kind := entry[17]
		off, oerr := strconv.Atoi(offStr)
		_, gerr := strconv.Atoi(genStr)
		if oerr != nil || gerr != nil {
			t.Fatalf("%s: xref entry %d %q has a non-numeric offset or generation field", label, i, entry)
		}
		switch kind {
		case 'f':
			if i != 0 {
				t.Fatalf("%s: xref entry %d is a free entry, but only object 0 should be", label, i)
			}
		case 'n':
			if i == 0 {
				t.Fatalf("%s: xref entry 0 must be the free entry, got in-use", label)
			}
			want := strconv.Itoa(i) + " 0 obj"
			if off < 0 || off+len(want) > len(b) || string(b[off:off+len(want)]) != want {
				t.Fatalf("%s: xref entry %d claims offset %d, but %q was not found there", label, i, off, want)
			}
		default:
			t.Fatalf("%s: xref entry %d has an unrecognised entry kind %q", label, i, string(kind))
		}
	}
}

// assertStreamLengthsAreExact finds every "/Length N" declaration
// followed by a stream body and checks that the body is exactly N bytes
// long, ending at a literal "endstream".
func assertStreamLengthsAreExact(t *testing.T, label string, b []byte) {
	t.Helper()

	found := 0
	idx := 0
	for {
		const kw = "/Length "
		rel := bytes.Index(b[idx:], []byte(kw))
		if rel == -1 {
			break
		}
		pos := idx + rel + len(kw)
		end := pos
		for end < len(b) && b[end] >= '0' && b[end] <= '9' {
			end++
		}
		if end == pos {
			t.Fatalf("%s: /Length at offset %d is not followed by digits", label, pos)
		}
		declared, err := strconv.Atoi(string(b[pos:end]))
		if err != nil {
			t.Fatalf("%s: /Length value %q is not an integer", label, b[pos:end])
		}

		const streamKW = "stream\n"
		streamRel := bytes.Index(b[end:], []byte(streamKW))
		if streamRel == -1 || streamRel > 40 {
			t.Fatalf("%s: no %q found shortly after /Length %d", label, streamKW, declared)
		}
		bodyStart := end + streamRel + len(streamKW)
		bodyEnd := bodyStart + declared
		if bodyEnd > len(b) {
			t.Fatalf("%s: /Length %d overruns the document", label, declared)
		}
		if !bytes.HasPrefix(b[bodyEnd:], []byte("endstream")) {
			t.Fatalf("%s: stream body does not end with endstream where /Length %d says it should", label, declared)
		}
		found++
		idx = bodyEnd
	}
	if found == 0 {
		t.Fatalf("%s: no stream object found to validate /Length against", label)
	}
}

// renderInSubprocess runs this same test binary as a fresh OS process with
// FOLIO_SUBPROCESS_RENDER=1 set, and the given GOMAXPROCS, and returns
// what it wrote to stdout.
func renderInSubprocess(t *testing.T, gomaxprocs string) []byte {
	t.Helper()

	// -test.run=^$ explicitly selects no test function to run; the
	// subprocessEnvVar check inside TestMain above is what actually
	// routes the child process. TestMain is not itself a test function,
	// so "-test.run=^TestMain$" (the previous flag here) matched nothing
	// either — harmlessly, since the env-var short-circuit runs before
	// m.Run() either way — but it named the wrong mechanism (this
	// story's QA review, Nit 24).
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(),
		subprocessEnvVar+"=1",
		"GOMAXPROCS="+gomaxprocs,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("subprocess render (GOMAXPROCS=%s) failed: %v\nstderr: %s", gomaxprocs, err, stderr.String())
	}
	return stdout.Bytes()
}

// firstDivergence returns the byte offset of the first byte at which a and
// b differ, and a short hex window around it, for diagnosing determinism
// failures.
func firstDivergence(a, b []byte) (offset int, window string) {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i, hexWindow(a, b, i)
		}
	}
	if len(a) != len(b) {
		return n, hexWindow(a, b, n)
	}
	return -1, ""
}

func hexWindow(a, b []byte, at int) string {
	const radius = 8
	start := at - radius
	if start < 0 {
		start = 0
	}
	end := func(x []byte) int {
		e := at + radius
		if e > len(x) {
			e = len(x)
		}
		return e
	}
	var sb strings.Builder
	sb.WriteString("a=")
	sb.WriteString(hexOf(a[start:end(a)]))
	sb.WriteString(" b=")
	sb.WriteString(hexOf(b[start:end(b)]))
	return sb.String()
}

func hexOf(b []byte) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 0, len(b)*2)
	for _, c := range b {
		out = append(out, hexDigits[c>>4], hexDigits[c&0x0f])
	}
	return string(out)
}

// TestRenderIsByteIdenticalAcrossTwoProcesses is AC8: the same input
// rendered by two independently started OS processes produces
// byte-identical output. Child A runs with GOMAXPROCS=1 and child B with
// GOMAXPROCS=4, so the comparison also catches order-dependence introduced
// by concurrency (e.g. an accidental map iteration reaching an output
// byte, NFR1.d).
func TestRenderIsByteIdenticalAcrossTwoProcesses(t *testing.T) {
	if runtime.GOOS == "js" {
		// js/wasm has no os/exec and no pipes (measured, Story 1.2 F-3):
		// exec.Command fails at build-select time and cmd.Run() would
		// fail at "pipe: not implemented on js" regardless. This is not
		// a determinism gap — Story 1.2's cross-target matrix harness
		// (folio-go/matrix_test.go, //go:build matrix) covers the same
		// property across all four targets, including js/wasm, via the
		// FOLIO_SUBPROCESS_RENDER seam driven from outside the process
		// rather than via os/exec inside it. The skip keys on
		// runtime.GOOS == "js" specifically, never on "an exec call
		// returned an error" — a genuine exec failure on a Linux leg
		// must stay a failure (AC15).
		t.Skip("js/wasm has no os/exec or pipes; covered by the cross-target matrix harness (folio-go/matrix_test.go)")
	}
	a := renderInSubprocess(t, "1")
	b := renderInSubprocess(t, "4")

	assertWellFormedPDF(t, "child A (GOMAXPROCS=1)", a)
	assertWellFormedPDF(t, "child B (GOMAXPROCS=4)", b)

	if !bytes.Equal(a, b) {
		offset, window := firstDivergence(a, b)
		t.Fatalf("subprocess outputs differ (len a=%d, len b=%d); first divergence at byte offset %d; %s",
			len(a), len(b), offset, window)
	}
}

// TestRenderProducesAValidPDF exercises AC5 directly in-process: the
// output is non-empty, has exactly one page, and passes the well-formed
// structural checks above. Independent-reader validation (qpdf --check)
// is recorded in the Delivery Log rather than wired into this test, since
// this module intentionally has no PDF-reading dependency (AC2).
func TestRenderProducesAValidPDF(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	b, err := Render(tpl, Data("{}"), nil, FontSet{})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	assertWellFormedPDF(t, "Render()", b)
}

// TestRenderWithNoDateInParamsOmitsCreationAndModDate is AC9, RE-SCOPED
// at Story 1.7 as AC24 (D-1.7.7) requires: neither key appears anywhere
// in the output, and no /Info dictionary is emitted at all. It also
// checks the neighbouring metadata/compression keys the Dev Notes and
// fixture README promise are absent but that nothing previously
// asserted: /Producer, /Creator and /Metadata (XMP) are the next places
// a timestamp, tool name or machine name could leak into the bytes —
// the same AD-7 hazard AC9 exists to close, one key over — and /Filter
// is the R4 compression risk this story's Dev Notes explicitly remove
// from the variable set (Story 1.6's QA review, Minor 20).
//
// Finding 12 (Story 1.6's QA review): the previous version swept ONLY
// the fontless document (an empty FontSet). F-7 explicitly flagged
// /Filter as the compression risk on the FONT path ("if this story
// compresses the FontFile2, that assertion goes red") — but as wired,
// it could never go red, because it never saw that path: the document
// with the large binary stream, where a future /Filter or /Producer
// would actually be tempting, had no metadata guard at all. This is
// table-driven over both documents.
//
// AC24 (D-1.7.7, this story): before Params existed, this test's
// "unconditional" absence assertion was true SOLELY because nothing
// could supply a date — a comment about intent, not a statement the
// test actually pinned. Now that Params exists, the property this test
// asserts is precisely AD-7's negative case, verbatim from
// epics.md:1098-1100 ("Given no date supplied by any route … Then
// /CreationDate and /ModDate are omitted") — Story 3.7's fourth AC,
// which owns the POSITIVE case (a date supplied through params DOES
// appear). Left unconditional and unrenamed, 3.7's developer would meet
// a red test whose cheapest fix is to weaken it — exactly the erosion
// this run has spent six stories preventing (D-1.7.7, binding
// rationale). The two new cases below cover both "no params at all"
// and "params supplied, but no date key in them" — 1.7 emits no PDF
// date and builds no date type; wiring an actual date through params
// into /CreationDate is Story 3.7's, not this test's.
func TestRenderWithNoDateInParamsOmitsCreationAndModDate(t *testing.T) {
	forbidden := []string{"/CreationDate", "/ModDate", "/Info", "/Producer", "/Creator", "/Metadata", "/Filter"}

	cases := []struct {
		label  string
		tpl    *Template
		fs     FontSet
		params Params
	}{
		{label: "fontless, no params", tpl: mustParseMinimalTemplate(t), fs: FontSet{}, params: nil},
		{label: "font-text, no params", tpl: parseFontTestTemplate(t), fs: testFontSet(), params: nil},
		// AC24: params NON-EMPTY, but carrying no date key at all — a
		// render whose params supply no date, deliberately, not merely
		// a render with no params argument at all.
		{label: "fontless, params present but no date key", tpl: mustParseMinimalTemplate(t), fs: FontSet{}, params: Params(`{"reportTitle": "Q3 Statement"}`)},
	}
	for _, c := range cases {
		b, err := Render(c.tpl, Data("{}"), c.params, c.fs)
		if err != nil {
			t.Fatalf("%s: Render() error: %v", c.label, err)
		}
		for _, key := range forbidden {
			if bytes.Contains(b, []byte(key)) {
				t.Errorf("%s: output contains %s", c.label, key)
			}
		}
	}
}

// mustParseMinimalTemplate parses minimalTemplateJSON, failing the test
// on any parse error — the fontless counterpart to parseFontTestTemplate.
func mustParseMinimalTemplate(t *testing.T) *Template {
	t.Helper()
	tpl, err := ParseTemplate([]byte(minimalTemplateJSON))
	if err != nil {
		t.Fatalf("parse minimalTemplateJSON: %v", err)
	}
	return tpl
}

// TestRenderIDEntriesAreIdentical is AC10's identity half: both /ID array
// entries are byte-identical to each other and 16 bytes each (32 hex
// characters).
func TestRenderIDEntriesAreIdentical(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	b, err := Render(tpl, Data("{}"), nil, FontSet{})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	idx := bytes.Index(b, []byte("/ID ["))
	if idx == -1 {
		t.Fatal("output does not contain an /ID array")
	}
	rest := b[idx+len("/ID ["):]

	first, rest, ok := extractAngleBracketed(rest)
	if !ok {
		t.Fatal("could not parse first /ID entry")
	}
	second, _, ok := extractAngleBracketed(rest)
	if !ok {
		t.Fatal("could not parse second /ID entry")
	}

	if len(first) != 32 {
		t.Errorf("first /ID entry is %d hex characters, want 32 (16 bytes)", len(first))
	}
	if len(second) != 32 {
		t.Errorf("second /ID entry is %d hex characters, want 32 (16 bytes)", len(second))
	}
	if first != second {
		t.Errorf("/ID entries differ: %q != %q", first, second)
	}
}

func extractAngleBracketed(b []byte) (content string, rest []byte, ok bool) {
	if len(b) == 0 || b[0] != '<' {
		return "", b, false
	}
	end := bytes.IndexByte(b, '>')
	if end == -1 {
		return "", b, false
	}
	return string(b[1:end]), b[end+1:], true
}

// letterSizeTemplateJSON is minimalTemplateJSON with "size": "Letter" —
// a legally loadable value (internal/template's closedPageSizeNames
// already accepts it, Story 1.4) that this story never built real
// dimensions for. Finding 17 (QA review), retained fixture: Render must
// return a located error, not silently substitute A4 (Story 1.3's
// standing rule: a deliberately violating fixture proves it fires, and
// the fixture is retained).
const letterSizeTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": []
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {},
  "locale": "en",
  "nextId": 1,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "Letter"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// TestRenderUnrecognisedNamedPageSizeIsLocatedError is Finding 17's fix
// (QA review): a document naming a page size other than "A4" (here,
// the schema-legal but dimensionally-unimplemented "Letter") must fail
// loudly rather than silently rendering as A4.
func TestRenderUnrecognisedNamedPageSizeIsLocatedError(t *testing.T) {
	tpl, err := ParseTemplate([]byte(letterSizeTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, err = Render(tpl, Data("{}"), nil, FontSet{})
	if err == nil {
		t.Fatal("expected a located error for page.size \"Letter\" (Finding 17), got nil")
	}
	if !strings.Contains(err.Error(), "Letter") {
		t.Errorf("error does not name the unrecognised size: %v", err)
	}
}

// TestRenderNilTemplateIsLocatedError is AC14b: Render(nil, f) returns a
// located error, not silently ignoring its argument the way the
// PROVISIONAL Story 1.1-1.4 signature did — "better than a public entry
// point documented as ignoring its argument" (D-1.5.9).
func TestRenderNilTemplateIsLocatedError(t *testing.T) {
	_, err := Render(nil, Data("{}"), nil, FontSet{})
	if err == nil {
		t.Fatal("expected a located error for Render(nil, ...), got nil")
	}
}

// TestRenderEmbedsSubsetFontAsType0Identity is AC5: a document using a
// subset of a Latin face is embedded as a Type0/Identity-H composite
// font carrying a FontFile2 stream and a ToUnicode CMap.
func TestRenderEmbedsSubsetFontAsType0Identity(t *testing.T) {
	tpl := parseFontTestTemplate(t)
	b, err := Render(tpl, Data("{}"), nil, testFontSet())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	assertWellFormedPDF(t, "font document", b)

	for _, want := range []string{"/Subtype /Type0", "/Encoding /Identity-H", "/FontFile2", "/ToUnicode", "/Subtype /CIDFontType2"} {
		if !bytes.Contains(b, []byte(want)) {
			t.Errorf("output does not contain %q", want)
		}
	}
	if !containsFontFile2(b) {
		t.Fatal("output does not contain a FontFile2 (AC10b vacuity guard)")
	}
}

// TestProvisionalBandOriginIsPinned is AC28: the provisional band-origin
// convention must be revisited the day internal/layout exists (Story
// 2.5) rather than quietly becoming permanent. This fails the moment
// that package appears, forcing a human to look at this file again.
//
// Finding 13 (QA review): the previous failure message named only
// internal/pdf/textdoc.go (buildTextContentStream) — which does hold a
// provisional convention (the Y-flip into PDF user space) — but AC28's
// actual subject, the BAND origin (pageHeader at 0, content below the
// header band, pageFooter at usableHeight-footerHeight), is implemented
// in folio-go/render.go's collectTextRuns, which the old message never
// mentioned. Both sites must be named, or the band-origin half can be
// missed — exactly the "provisional quietly becomes real" outcome AC28
// exists to prevent, arriving through a stale pointer.
func TestProvisionalBandOriginIsPinned(t *testing.T) {
	if _, err := os.Stat(filepath.Join("internal", "layout")); err == nil {
		t.Fatal(
			"internal/layout now exists — AD-24's real band-relative placement has arrived. " +
				"TWO provisional sites must be replaced, and this pinning test deleted " +
				"(AC28, D-1.5.5's self-retiring-assertion pattern): " +
				"(1) folio-go/render.go's collectTextRuns — the PROVISIONAL band-origin convention " +
				"(pageHeader at y=0, content below the header band, pageFooter flush against the " +
				"bottom margin); " +
				"(2) folio-go/internal/pdf/textdoc.go's buildTextContentStream — the PROVISIONAL " +
				"Y-flip into PDF's bottom-up user space.",
		)
	}
}

// TestRenderWithFontIsByteIdenticalAcrossTwoProcesses is the font-path
// counterpart to TestRenderIsByteIdenticalAcrossTwoProcesses above,
// exercised locally (not just in the matrix): two independent OS
// processes rendering the same template+font document produce
// byte-identical output, and — AC10b's vacuity guard, applied here too
// — the output is confirmed to actually contain a FontFile2 before any
// comparison runs.
func TestRenderWithFontIsByteIdenticalAcrossTwoProcesses(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skip("js/wasm has no os/exec or pipes; covered by the cross-target matrix harness (folio-go/matrix_test.go)")
	}
	a := renderFontInSubprocess(t, "1")
	b := renderFontInSubprocess(t, "4")

	assertWellFormedPDF(t, "child A (GOMAXPROCS=1, font)", a)
	assertWellFormedPDF(t, "child B (GOMAXPROCS=4, font)", b)

	if !containsFontFile2(a) {
		t.Fatal("child A output does not contain a FontFile2 (AC10b vacuity guard) — a match below would prove nothing")
	}
	if !containsFontFile2(b) {
		t.Fatal("child B output does not contain a FontFile2 (AC10b vacuity guard) — a match below would prove nothing")
	}

	if !bytes.Equal(a, b) {
		offset, window := firstDivergence(a, b)
		t.Fatalf("subprocess font-document outputs differ (len a=%d, len b=%d); first divergence at byte offset %d; %s",
			len(a), len(b), offset, window)
	}
}

func renderFontInSubprocess(t *testing.T, gomaxprocs string) []byte {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(),
		subprocessFontEnvVar+"=1",
		"GOMAXPROCS="+gomaxprocs,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("subprocess font render (GOMAXPROCS=%s) failed: %v\nstderr: %s", gomaxprocs, err, stderr.String())
	}
	return stdout.Bytes()
}

// TestEmbeddedHeadTimestampsEqualSourceExactly is AC11a: the embedded
// FontFile2's head table created (offset 20) and modified (offset 28)
// equal the SOURCE face's values exactly, byte-for-byte — equality
// alone, no "or zero" latitude and no clock read (D-1.5.10 withdrew
// both from D-1.5.4). F-5's measured values for the committed test face:
// created=3304067374, modified=3573633780.
func TestEmbeddedHeadTimestampsEqualSourceExactly(t *testing.T) {
	tpl := parseFontTestTemplate(t)
	b, err := Render(tpl, Data("{}"), nil, testFontSet())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !containsFontFile2(b) {
		t.Fatal("output does not contain a FontFile2 — cannot proceed (AC10b vacuity guard)")
	}

	embeddedProgram := extractFontFile2Program(t, b)
	embCreated, embModified := headTimesFromSFNT(t, embeddedProgram)
	srcCreated, srcModified := headTimesFromSFNT(t, testRobotoFontBytes)

	if embCreated != srcCreated {
		t.Errorf("embedded head.created = %d, want source's %d (F-5, AC11a)", embCreated, srcCreated)
	}
	if embModified != srcModified {
		t.Errorf("embedded head.modified = %d, want source's %d (F-5, AC11a)", embModified, srcModified)
	}
	if srcCreated != 3304067374 {
		t.Errorf("source head.created = %d, want the measured 3304067374 (F-5) — has testdata/fonts/Roboto-Regular.ttf changed?", srcCreated)
	}
	if srcModified != 3573633780 {
		t.Errorf("source head.modified = %d, want the measured 3573633780 (F-5) — has testdata/fonts/Roboto-Regular.ttf changed?", srcModified)
	}
}

// extractFontFile2Program locates the FIRST FontFile2 stream in a
// rendered PDF and returns its raw program bytes, by locating the
// object it references and reading exactly /Length1 bytes after
// "stream\n". This is a minimal, purpose-built extractor — not a
// general PDF parser — consistent with AC2/AC13: folio-go has no
// PDF-reading dependency, so verification code that needs to read one
// back is written by hand, the same way assertWellFormedPDF is.
func extractFontFile2Program(t *testing.T, pdfBytes []byte) []byte {
	t.Helper()
	idx := bytes.Index(pdfBytes, []byte("/FontFile2 "))
	if idx == -1 {
		t.Fatal("no /FontFile2 key found")
	}
	rest := pdfBytes[idx+len("/FontFile2 "):]
	sp := bytes.IndexByte(rest, ' ')
	if sp == -1 {
		t.Fatal("could not parse /FontFile2 object reference")
	}
	objNum, err := strconv.Atoi(string(rest[:sp]))
	if err != nil {
		t.Fatalf("could not parse /FontFile2 object number: %v", err)
	}

	marker := []byte(strconv.Itoa(objNum) + " 0 obj\n")
	objIdx := bytes.Index(pdfBytes, marker)
	if objIdx == -1 {
		t.Fatalf("could not find object %d 0 obj", objNum)
	}
	objBody := pdfBytes[objIdx:]

	const lenKW = "/Length1 "
	lenIdx := bytes.Index(objBody, []byte(lenKW))
	if lenIdx == -1 {
		t.Fatal("FontFile2 object has no /Length1")
	}
	numStart := lenIdx + len(lenKW)
	numEnd := numStart
	for numEnd < len(objBody) && objBody[numEnd] >= '0' && objBody[numEnd] <= '9' {
		numEnd++
	}
	length, err := strconv.Atoi(string(objBody[numStart:numEnd]))
	if err != nil {
		t.Fatalf("could not parse /Length1: %v", err)
	}

	const streamKW = "stream\n"
	streamIdx := bytes.Index(objBody[numEnd:], []byte(streamKW))
	if streamIdx == -1 {
		t.Fatal("FontFile2 object has no stream keyword")
	}
	bodyStart := numEnd + streamIdx + len(streamKW)
	if bodyStart+length > len(objBody) {
		t.Fatal("FontFile2 stream overruns the document")
	}
	return objBody[bodyStart : bodyStart+length]
}

// extractAllFontFile2Programs is Story 2.2's multi-face extension of
// extractFontFile2Program above: the multi-script fixture (AC8) embeds
// THREE distinct FontFile2 streams, one per shipped face actually used
// — this walks every "/FontFile2 N 0 R" occurrence (by object number,
// deduplicated, since a Type0/CIDFontType2 pair can each reference the
// FontDescriptor that carries it) and returns each program's bytes in
// ascending object-number order (deterministic — never map iteration
// order).
func extractAllFontFile2Programs(t *testing.T, pdfBytes []byte) [][]byte {
	t.Helper()

	seen := map[int]bool{}
	var objNums []int
	rest := pdfBytes
	for {
		idx := bytes.Index(rest, []byte("/FontFile2 "))
		if idx == -1 {
			break
		}
		after := rest[idx+len("/FontFile2 "):]
		sp := bytes.IndexByte(after, ' ')
		if sp == -1 {
			t.Fatal("could not parse /FontFile2 object reference")
		}
		objNum, err := strconv.Atoi(string(after[:sp]))
		if err != nil {
			t.Fatalf("could not parse /FontFile2 object number: %v", err)
		}
		if !seen[objNum] {
			seen[objNum] = true
			objNums = append(objNums, objNum)
		}
		rest = after[sp:]
	}
	if len(objNums) == 0 {
		t.Fatal("no /FontFile2 key found")
	}
	sort.Ints(objNums) // deterministic order — never rely on map/seen iteration.

	programs := make([][]byte, 0, len(objNums))
	for _, objNum := range objNums {
		marker := []byte(strconv.Itoa(objNum) + " 0 obj\n")
		objIdx := bytes.Index(pdfBytes, marker)
		if objIdx == -1 {
			t.Fatalf("could not find object %d 0 obj", objNum)
		}
		objBody := pdfBytes[objIdx:]

		const lenKW = "/Length1 "
		lenIdx := bytes.Index(objBody, []byte(lenKW))
		if lenIdx == -1 {
			t.Fatalf("object %d has no /Length1", objNum)
		}
		numStart := lenIdx + len(lenKW)
		numEnd := numStart
		for numEnd < len(objBody) && objBody[numEnd] >= '0' && objBody[numEnd] <= '9' {
			numEnd++
		}
		length, err := strconv.Atoi(string(objBody[numStart:numEnd]))
		if err != nil {
			t.Fatalf("object %d: could not parse /Length1: %v", objNum, err)
		}

		const streamKW = "stream\n"
		streamIdx := bytes.Index(objBody[numEnd:], []byte(streamKW))
		if streamIdx == -1 {
			t.Fatalf("object %d has no stream keyword", objNum)
		}
		bodyStart := numEnd + streamIdx + len(streamKW)
		if bodyStart+length > len(objBody) {
			t.Fatalf("object %d: FontFile2 stream overruns the document", objNum)
		}
		programs = append(programs, objBody[bodyStart:bodyStart+length])
	}
	return programs
}

// sfntHasTable reports whether a raw sfnt program's table directory
// names tag (a 4-byte OpenType table tag, e.g. "fvar"). Used by AC8's
// feature guard to assert an embedded program is genuinely INSTANCED —
// a variable font retains fvar/gvar/avar; PinAllAxesToDefault's whole
// purpose (Story 2.2, AC7) is to make the SUBSETTED, embedded program
// carry NONE of them.
func sfntHasTable(t *testing.T, data []byte, tag string) bool {
	t.Helper()
	if len(data) < 12 {
		t.Fatal("sfnt data too short to have a table directory")
	}
	numTables := int(data[4])<<8 | int(data[5])
	for i := 0; i < numTables; i++ {
		recStart := 12 + i*16
		if recStart+4 > len(data) {
			t.Fatal("sfnt table directory overruns data")
		}
		if string(data[recStart:recStart+4]) == tag {
			return true
		}
	}
	return false
}

// headTimesFromSFNT reads an sfnt's head table created/modified fields
// (offsets 20 and 28, LONGDATETIME) directly from its table directory —
// the same technique fontset_test.go's patchUnitsPerEm helper uses, kept
// as an independent, from-scratch read here rather than importing
// internal/fontset (this test verifies the PUBLIC Render output, not
// the internal seam).
func headTimesFromSFNT(t *testing.T, data []byte) (created, modified int64) {
	t.Helper()
	if len(data) < 12 {
		t.Fatal("sfnt data too short")
	}
	numTables := int(data[4])<<8 | int(data[5])
	for i := 0; i < numTables; i++ {
		recStart := 12 + i*16
		if recStart+16 > len(data) {
			t.Fatal("sfnt table directory overruns data")
		}
		rec := data[recStart : recStart+16]
		if string(rec[0:4]) != "head" {
			continue
		}
		off := int(rec[8])<<24 | int(rec[9])<<16 | int(rec[10])<<8 | int(rec[11])
		if off+54 > len(data) {
			t.Fatal("head table overruns data")
		}
		head := data[off : off+54]
		created = beInt64(head[20:28])
		modified = beInt64(head[28:36])
		return created, modified
	}
	t.Fatal("sfnt has no head table")
	return 0, 0
}

func beInt64(b []byte) int64 {
	var v uint64
	for _, c := range b {
		v = v<<8 | uint64(c)
	}
	return int64(v)
}
