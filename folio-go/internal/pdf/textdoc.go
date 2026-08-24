package pdf

import (
	"maps"
	"slices"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// EmbeddedFace is one font's resolved, subsetted embedding, ready to be
// written into a PDF (AC5). internal/pdf never touches a vendor font
// type: everything here is plain data that already crossed
// internal/fontset's boundary (AC17a) — this package's job is only to
// spell it out as PDF objects, per AD-3's routing rules.
type EmbeddedFace struct {
	Name           string // PDF resource name (the caller's FontSet key)
	PostScriptName string // the embedded program's OWN name — see baseFont below
	Program        []byte // FontFile2 payload (a standalone TrueType font)
	Tag            string // AC6's six-letter subset tag
	NumGlyphs      int
	GlyphForRune   map[rune]uint16 // rune -> CID (Identity-H: CID == subset glyph id)
	WidthForGlyph  map[uint16]int64

	Ascent, Descent, CapHeight             int64
	BBoxXMin, BBoxYMin, BBoxXMax, BBoxYMax int64
	HeadCreated, HeadModified              int64 // AC11a: copied verbatim into the embedded head table
}

// TextRun is one literal run of text on one page, placed at a
// PROVISIONAL origin (AC28). AD-24 makes element x/y band-relative once
// internal/layout exists (Story 2.5); until then this package treats
// X/Y as an offset from the page's top-left printable corner (inside
// the margins), Y increasing DOWNWARD — the opposite sense from PDF's
// native bottom-up user space, chosen to match how X/Y read in the
// `.folio` document model. TestProvisionalBandOriginIsPinned (in
// folio-go, module root) fails the moment internal/layout exists,
// forcing this convention to be revisited rather than quietly becoming
// permanent.
type TextRun struct {
	Face     string
	Text     string
	X, Y     geom.Length
	FontSize geom.Length
}

// ImagePlacement is one image element's resolved, DRAWN placement on a
// page (AD-24: scaled to fit, centred, computed in integer millipoints —
// the fit/centre computation itself lives in package folio's render.go;
// this struct carries only the RESULT). X, Y are the top-left corner of
// the DRAWN box (already centred within the element's declared box,
// i.e. offset by the centring math), in the same page-local, Y-down
// convention as TextRun.X/Y.
type ImagePlacement struct {
	ResourceName          string // key into the images map passed to SerializeTextDocument
	X, Y                  geom.Length
	DrawWidth, DrawHeight geom.Length
}

// ImageXObject is one distinct embedded image asset, ready to be
// written as a PDF XObject stream (AC9, AC10): the Stream bytes are
// ALREADY COMPRESSED (the source file's own JPEG or PNG IDAT bytes,
// passed through unchanged — D-1.8.1) and are never decoded or
// re-encoded by this package. Embedded exactly once per document
// regardless of how many elements place it (AC8's dedup-by-key
// definition, mirrored here the same way EmbeddedFace is embedded once
// per document). Stream/Filter here are exactly what
// internal/template.DecodedImage produced (AC10a: see that type's own
// Stream field for the full explanation of why this is never
// re-compressed — carried risk R4, D-1.8.1, AD-22).
type ImageXObject struct {
	Width, Height    int64 // intrinsic pixel dimensions (/Width, /Height — appendInt, D-1.1.b)
	ColorSpace       string
	BitsPerComponent int64
	Filter           string // "DCTDecode" or "FlateDecode", no leading slash
	HasDecodeParms   bool
	PredictorColors  int64
	PredictorBPC     int64
	PredictorColumns int64
	Stream           []byte
}

// TextPage is one page's provisional content plus the page geometry
// needed to place it (AC28's provisional-origin math needs the page
// height and margins to flip Y into PDF's bottom-up space).
type TextPage struct {
	Runs                  []TextRun
	Images                []ImagePlacement
	Width, Height         geom.Length
	MarginTop, MarginLeft geom.Length
}

// SerializeTextDocument writes pages as a classic (uncompressed) PDF 1.7
// document, embedding every distinct face named in faces EXACTLY ONCE
// regardless of how many pages or runs reference it (AC9: "one subset
// per font per document... never per page, never per element" — this
// function is the "never per page" half; the "one subsetting call"
// half is internal/fontset.Font.Subset's caller's job, upstream of
// here). It follows Story 1.1/1.4's determinism rules: every geometric
// number goes through appendLength, every count/offset/object number
// through appendInt/appendIntPadded (AD-3), and /ID is derived from the
// document's own content (AD-7) — no compression, no /Info dictionary,
// no /CreationDate or /ModDate.
func SerializeTextDocument(pages []TextPage, faces map[string]EmbeddedFace, images map[string]ImageXObject) ([]byte, error) {
	b := newBuilder()

	catalogID := b.reserve()
	pagesID := b.reserve()

	// Images are emitted in SORTED resource-name order (same
	// ScanMapRange-compliant idiom as faces below): deterministic
	// object numbers and resource-dict key order (AC18a: asset keys are
	// 64 lowercase hex characters, a strict subset of pdfNameEscape's
	// kept set, so distinct keys cannot collide as resource names —
	// asserted directly in TestAssetKeyEscapeIsIdentity).
	imageNames := slices.Sorted(maps.Keys(images))
	imageIDs := make(map[string]int64, len(imageNames))
	for _, name := range imageNames {
		imageIDs[name] = b.reserve()
	}

	// Faces are emitted in SORTED name order (ScanMapRange-compliant:
	// the escape hatch idiom, not a raw map range) — this reaches
	// output bytes (object numbers, the resource dictionary's key
	// order), same reasoning as AC6/AC7's tag derivation.
	faceNames := slices.Sorted(maps.Keys(faces))

	// Finding 23 (QA review): pdfNameEscape drops every character
	// outside [A-Za-z0-9_-], so two distinct face names can reduce to
	// the same escaped resource name (e.g. "Roboto Regular" and
	// "RobotoRegular" both become "RobotoRegular"). Detect that
	// collision across the WHOLE document's face set up front and fail
	// loudly, rather than silently rendering one face's runs in
	// another's glyphs.
	seenResourceNames := make(map[string]string, len(faceNames))
	for _, name := range faceNames {
		escaped := pdfNameEscape(name)
		if other, collided := seenResourceNames[escaped]; collided {
			return nil, &resourceNameCollisionError{a: other, b: name, resourceName: escaped}
		}
		seenResourceNames[escaped] = name
	}

	type faceObjIDs struct {
		type0, cidFont, descriptor, fontFile, toUnicode int64
	}
	faceIDs := make(map[string]faceObjIDs, len(faceNames))
	for _, name := range faceNames {
		faceIDs[name] = faceObjIDs{
			type0:      b.reserve(),
			cidFont:    b.reserve(),
			descriptor: b.reserve(),
			fontFile:   b.reserve(),
			toUnicode:  b.reserve(),
		}
	}

	pageIDs := make([]int64, len(pages))
	contentIDs := make([]int64, len(pages))
	for i := range pages {
		pageIDs[i] = b.reserve()
		contentIDs[i] = b.reserve()
	}

	// --- Catalog ---
	b.begin(catalogID)
	b.write([]byte("<< /Type /Catalog /Pages "))
	b.writeRef(pagesID)
	b.write([]byte(" >>"))
	b.end()

	// --- Pages (parent) ---
	b.begin(pagesID)
	b.write([]byte("<< /Type /Pages /Kids ["))
	for _, pid := range pageIDs {
		b.writeRef(pid)
	}
	b.write([]byte("] /Count "))
	b.writeInt(int64(len(pages)))
	b.write([]byte(" >>"))
	b.end()

	// --- Per-page objects ---
	for i, page := range pages {
		content, cerr := buildTextContentStream(page, faces)
		if cerr != nil {
			return nil, cerr
		}
		content, cerr = appendImageContentStream(content, page, imageIDs)
		if cerr != nil {
			return nil, cerr
		}

		b.begin(pageIDs[i])
		b.write([]byte("<< /Type /Page /Parent "))
		b.writeRef(pagesID)
		b.write([]byte(" /MediaBox [0 0 "))
		b.writeLength(page.Width)
		b.write([]byte(" "))
		b.writeLength(page.Height)
		b.write([]byte("] /Resources << /Font <<"))
		for _, name := range faceNames {
			b.write([]byte(" /"))
			b.write([]byte(pdfNameEscape(name)))
			b.write([]byte(" "))
			b.writeRef(faceIDs[name].type0)
		}
		b.write([]byte(" >>"))
		if len(imageNames) > 0 {
			b.write([]byte(" /XObject <<"))
			for _, name := range imageNames {
				b.write([]byte(" /"))
				b.write([]byte(pdfNameEscape(name)))
				b.write([]byte(" "))
				b.writeRef(imageIDs[name])
			}
			b.write([]byte(" >>"))
		}
		b.write([]byte(" >> /Contents "))
		b.writeRef(contentIDs[i])
		b.write([]byte(" >>"))
		b.end()

		b.begin(contentIDs[i])
		b.write([]byte("<< /Length "))
		b.writeInt(int64(len(content)))
		b.write([]byte(" >>\nstream\n"))
		b.write(content)
		b.write([]byte("endstream"))
		b.end()
	}

	// --- Per-face font objects ---
	for _, name := range faceNames {
		face := faces[name]
		ids := faceIDs[name]
		// ISO 32000-1 Table 117 (CIDFontType2): /BaseFont "shall be the
		// value of the CIDFontName entry in the CIDFont program",
		// prefixed with the six-letter subset tag (§9.6.4). Before
		// Story 2.2 folio spelled this from the FontSet KEY — the name
		// the caller happened to file the face under — so the declared
		// name and the embedded program disagreed (`NotoSansSC` vs
		// `NotoSansSC-Regular`). PDF/A validators flag that, and it is
		// the name a viewer falls back to when the embedded program
		// fails to load.
		//
		// The two halves are INDEPENDENT, which is what keeps the tag
		// derivation non-circular (AC6): the tag hashes the final
		// program bytes, while PostScriptName is read off the supplied
		// face's own `name` table. Neither is an input to the other, and
		// the embedded program carries no `name` table for a tag to be
		// written into (PDF §9.9.2 sanctions the reduced table set for
		// an embedded CIDFontType2, since /CIDToGIDMap does the
		// mapping). The semantic acceptance assertion pins the composed
		// result to ^[A-Z]{6}\+<name6>$ per face, which catches both a
		// stale name and a double-applied tag.
		//
		// Falls back to the FontSet key — the pre-2.2 behaviour — only
		// when the supplied program declares no name record at all.
		psName := face.PostScriptName
		if psName == "" {
			psName = name
		}
		baseFont := face.Tag + "+" + pdfNameEscape(psName)

		toUnicodeCMap := buildToUnicodeCMap(face)

		b.begin(ids.type0)
		b.write([]byte("<< /Type /Font /Subtype /Type0 /BaseFont /"))
		b.write([]byte(baseFont))
		b.write([]byte(" /Encoding /Identity-H /DescendantFonts ["))
		b.writeRef(ids.cidFont)
		b.write([]byte("] /ToUnicode "))
		b.writeRef(ids.toUnicode)
		b.write([]byte(" >>"))
		b.end()

		b.begin(ids.cidFont)
		b.write([]byte("<< /Type /Font /Subtype /CIDFontType2 /BaseFont /"))
		b.write([]byte(baseFont))
		b.write([]byte(" /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >>"))
		b.write([]byte(" /FontDescriptor "))
		b.writeRef(ids.descriptor)
		b.write([]byte(" /DW "))
		b.writeInt(1000)
		b.write([]byte(" /W ["))
		b.writeInt(0)
		b.write([]byte(" ["))
		for newGID := 0; newGID < face.NumGlyphs; newGID++ {
			if newGID > 0 {
				b.write([]byte(" "))
			}
			b.writeInt(face.WidthForGlyph[uint16(newGID)])
		}
		b.write([]byte("]] /CIDToGIDMap /Identity >>"))
		b.end()

		b.begin(ids.descriptor)
		b.write([]byte("<< /Type /FontDescriptor /FontName /"))
		b.write([]byte(baseFont))
		b.write([]byte(" /Flags "))
		b.writeInt(4) // bit 3 (Symbolic) — this is a CID-keyed subset, not one of the 14 standard fonts.
		b.write([]byte(" /FontBBox ["))
		b.writeInt(face.BBoxXMin)
		b.write([]byte(" "))
		b.writeInt(face.BBoxYMin)
		b.write([]byte(" "))
		b.writeInt(face.BBoxXMax)
		b.write([]byte(" "))
		b.writeInt(face.BBoxYMax)
		b.write([]byte("] /ItalicAngle "))
		// Finding 22 (QA review): route through writeInt rather than a
		// bare "0" literal — D-1.1.b names /ItalicAngle explicitly as a
		// "convert to thousandths, decimal route" value. It is a
		// constant 0 today (no italic faces are tested), but the shape
		// is already correct for the day the value stops being constant.
		b.writeInt(0)
		b.write([]byte(" /Ascent "))
		b.writeInt(face.Ascent)
		b.write([]byte(" /Descent "))
		b.writeInt(face.Descent)
		b.write([]byte(" /CapHeight "))
		b.writeInt(face.CapHeight)
		b.write([]byte(" /StemV "))
		b.writeInt(80)
		b.write([]byte(" /FontFile2 "))
		b.writeRef(ids.fontFile)
		b.write([]byte(" >>"))
		b.end()

		b.begin(ids.fontFile)
		b.write([]byte("<< /Length "))
		b.writeInt(int64(len(face.Program)))
		b.write([]byte(" /Length1 "))
		b.writeInt(int64(len(face.Program)))
		b.write([]byte(" >>\nstream\n"))
		b.write(face.Program)
		b.write([]byte("endstream"))
		b.end()

		b.begin(ids.toUnicode)
		b.write([]byte("<< /Length "))
		b.writeInt(int64(len(toUnicodeCMap)))
		b.write([]byte(" >>\nstream\n"))
		b.write(toUnicodeCMap)
		b.write([]byte("endstream"))
		b.end()
	}

	// --- Per-image XObject objects ---
	for _, name := range imageNames {
		img := images[name]
		b.begin(imageIDs[name])
		b.write([]byte("<< /Type /XObject /Subtype /Image /Width "))
		b.writeInt(img.Width)
		b.write([]byte(" /Height "))
		b.writeInt(img.Height)
		b.write([]byte(" /ColorSpace /"))
		b.write([]byte(img.ColorSpace))
		b.write([]byte(" /BitsPerComponent "))
		b.writeInt(img.BitsPerComponent)
		b.write([]byte(" /Filter /"))
		b.write([]byte(img.Filter))
		if img.HasDecodeParms {
			b.write([]byte(" /DecodeParms << /Predictor 15 /Colors "))
			b.writeInt(img.PredictorColors)
			b.write([]byte(" /BitsPerComponent "))
			b.writeInt(img.PredictorBPC)
			b.write([]byte(" /Columns "))
			b.writeInt(img.PredictorColumns)
			b.write([]byte(" >>"))
		}
		b.write([]byte(" /Length "))
		b.writeInt(int64(len(img.Stream)))
		b.write([]byte(" >>\nstream\n"))
		b.write(img.Stream)
		b.write([]byte("endstream"))
		b.end()
	}

	return b.finish(), nil
}

// pdfNameEscape returns name with characters outside a conservative safe
// set (ASCII letters, digits, hyphen, underscore) replaced with nothing
// — enough for the face names and tags this story constructs; PDF name
// escaping (#xx) is not needed for the identifiers this story produces.
func pdfNameEscape(name string) string {
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '_':
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return "F"
	}
	return string(out)
}

// buildToUnicodeCMap builds a minimal ToUnicode CMap stream (bfchar
// entries, one per glyph actually used) mapping each embedded glyph's
// CID back to the Unicode codepoint it represents — this is what an
// independent reader (AC13) uses to extract correct text from the
// rendered PDF, and what this story's independent-reader validation
// checks. Basic Multilingual Plane only (runes <= 0xFFFF, always true
// for this story's Latin test face); a supplementary-plane rune would
// need a UTF-16 surrogate pair, out of scope here.
//
// Entries are emitted in ascending CID order — reached by sorting the
// (rune -> CID) map's VALUES via a plain, non-map slice sort, so no
// map-range site is introduced here beyond the one already declared
// ScanMapRange-compliant via slices.Sorted(maps.Keys(...)).
func buildToUnicodeCMap(face EmbeddedFace) []byte {
	type entry struct {
		cid uint16
		r   rune
	}
	entries := make([]entry, 0, len(face.GlyphForRune))
	for _, r := range slices.Sorted(maps.Keys(face.GlyphForRune)) {
		entries = append(entries, entry{cid: face.GlyphForRune[r], r: r})
	}
	slices.SortFunc(entries, func(a, b entry) int { return int(a.cid) - int(b.cid) })

	var c []byte
	c = append(c, "/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n"...)
	c = append(c, "/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n"...)
	c = append(c, "/CMapName /Adobe-Identity-UCS def\n/CMapType 2 def\n"...)
	c = append(c, "1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n"...)

	c = appendInt(c, int64(len(entries)))
	c = append(c, " beginbfchar\n"...)
	for _, e := range entries {
		c = append(c, '<')
		c = appendHex4(c, e.cid)
		c = append(c, "> <"...)
		c = appendHex4(c, uint16(e.r))
		c = append(c, ">\n"...)
	}
	c = append(c, "endbfchar\n"...)
	c = append(c, "endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n"...)
	return c
}

// appendHex4 encodes v as a 4-hex-digit big-endian byte pair. D-1.1.b,
// verbatim, on why this is not routed through numbers.go's appendInt:
// "Glyph ids under Identity-H take neither route — they are a
// big-endian hex pair inside a string literal, i.e. a byte encoding
// like /ID, not a number. Same for the six-letter subset tag. This
// must be said in a code comment or a later agent will 'unify' it."
func appendHex4(dst []byte, v uint16) []byte {
	return append(dst,
		hexDigits[(v>>12)&0xf], hexDigits[(v>>8)&0xf],
		hexDigits[(v>>4)&0xf], hexDigits[v&0xf],
	)
}

// buildTextContentStream renders one page's text runs as a content
// stream: for each run, select the resource font, position with an
// absolute text matrix (Tm — avoids accumulating relative Td offsets
// across runs), and show its text as a hex string of 2-byte CIDs
// (Identity-H: CID == subset glyph id).
//
// Provisional origin (AC28): run.X/run.Y are read as an offset from the
// page's top-left printable corner, Y increasing DOWNWARD; converted to
// PDF's bottom-up user space as
//
//	pdfX = marginLeft + run.X
//	pdfY = pageHeight - marginTop - run.Y - run.FontSize
//
// placing the text baseline approximately run.FontSize below the run's
// top-left corner. This is a stand-in for AD-24's real band-relative
// placement, which arrives with internal/layout in Story 2.5 —
// TestProvisionalBandOriginIsPinned (folio-go, module root) fails the
// day that package exists, forcing this function to be revisited.
func buildTextContentStream(page TextPage, faces map[string]EmbeddedFace) ([]byte, error) {
	var c []byte
	for _, run := range page.Runs {
		if run.Text == "" {
			continue
		}
		face, ok := faces[run.Face]
		if !ok {
			return nil, &missingFaceError{face: run.Face}
		}

		pdfX := page.MarginLeft + run.X
		pdfY := flipY(page.Height, page.MarginTop, run.Y, run.FontSize)

		c = append(c, "BT\n/"...)
		c = append(c, pdfNameEscape(run.Face)...)
		c = append(c, ' ')
		c = appendLength(c, run.FontSize)
		c = append(c, " Tf\n1 0 0 1 "...)
		c = appendLength(c, pdfX)
		c = append(c, ' ')
		c = appendLength(c, pdfY)
		c = append(c, " Tm\n<"...)
		hex, herr := appendHexCIDString(c, run.Text, face)
		if herr != nil {
			return nil, herr
		}
		c = hex
		c = append(c, "> Tj\nET\n"...)
	}
	return c, nil
}

type resourceNameCollisionError struct{ a, b, resourceName string }

func (e *resourceNameCollisionError) Error() string {
	return "internal/pdf: face names " + e.a + " and " + e.b +
		" both escape to the PDF resource name " + e.resourceName + " (Finding 23, QA review)"
}

type missingFaceError struct{ face string }

func (e *missingFaceError) Error() string {
	return "internal/pdf: text run names face " + e.face + ", not present in the document's face set"
}

type missingGlyphError struct {
	face string
	r    rune
}

func (e *missingGlyphError) Error() string {
	return "internal/pdf: face " + e.face + " has no subset glyph for a rune in its own text run (document assembly did not include it in the subsetting union)"
}

const hexDigits = "0123456789abcdef"

// appendHexCIDString appends text as a PDF hex string body (without the
// surrounding angle brackets): one 2-byte big-endian CID per rune,
// looked up in face.GlyphForRune (Identity-H: CID == subset glyph id).
// A rune absent from the map is a located error, not a silently emitted
// .notdef — every text element's runes are included in the union handed
// to Font.Subset before this function ever runs (folio.go), so a miss
// here means document assembly and content-stream assembly disagreed
// about what the document uses, which must fail loudly.
//
// D-1.1.b, verbatim, on why these CIDs are not routed through
// numbers.go's appendInt: "Glyph ids under Identity-H take neither
// route — they are a big-endian hex pair inside a string literal, i.e.
// a byte encoding like /ID, not a number. Same for the six-letter
// subset tag. This must be said in a code comment or a later agent
// will 'unify' it."
func appendHexCIDString(dst []byte, text string, face EmbeddedFace) ([]byte, error) {
	for _, r := range text {
		cid, ok := face.GlyphForRune[r]
		if !ok {
			return nil, &missingGlyphError{face: face.Name, r: r}
		}
		b := [2]byte{byte(cid >> 8), byte(cid)}
		for _, x := range b {
			dst = append(dst, hexDigits[x>>4], hexDigits[x&0x0f])
		}
	}
	return dst, nil
}
