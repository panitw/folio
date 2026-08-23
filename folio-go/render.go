package folio

import (
	"fmt"
	"maps"
	"slices"

	"github.com/panitw/folio/folio-go/internal/bind"
	"github.com/panitw/folio/folio-go/internal/fontset"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/pdf"
	"github.com/panitw/folio/folio-go/internal/template"
)

// A4 page dimensions in millipoints, matching internal/pdf's Story 1.1
// fixture constants (595.276pt x 841.89pt) — package folio does not
// import those unexported constants; it restates the one named size
// this story resolves. A custom page size is read directly from the
// template instead.
const (
	pageWidthA4  geom.Length = 595276
	pageHeightA4 geom.Length = 841890
)

// defaultFontSizePt is the provisional default text size (12pt) used
// when a text element's style does not specify one (AC28: provisional,
// pending real style inheritance/defaults, which are not this story's
// concern).
const defaultFontSizePt geom.Length = 12000

// pageDimensions resolves a Document's page geometry (AC28's placement
// math needs page height and margins). An unrecognised named size is a
// located error, not a silent A4 substitution (Finding 17, QA review:
// the previous version fell back to A4 for ANY named size other than
// "A4" — a document declaring "size": "Letter" rendered as A4 with no
// error, silent wrong output on a valid-looking input, in a module
// whose entire premise is predictable bytes). Note that
// internal/template's closedPageSizeNames (Story 1.4) already accepts
// "Letter" as a load-time-valid value — this story only ever built A4
// dimensions (this story's scope is the font-embedding mechanism, not a
// named-page-size table), so "Letter" is a legally loadable document
// this version of Render cannot yet produce bytes for; failing loudly
// here is more honest than a silent A4 substitution, and adding real
// Letter dimensions is out of this story's scope.
func pageDimensions(doc *Template) (width, height geom.Length, err error) {
	switch {
	case doc.Page.SizeIsName && doc.Page.SizeName == "A4":
		width, height = pageWidthA4, pageHeightA4
	case !doc.Page.SizeIsName:
		width, height = doc.Page.SizeCustom.Width, doc.Page.SizeCustom.Height
	default:
		return 0, 0, fmt.Errorf(
			"folio: Render: page.size names %q, which this version does not implement dimensions "+
				"for (only \"A4\" or a custom width/height)", doc.Page.SizeName,
		)
	}
	if doc.Page.Orientation == "landscape" {
		width, height = height, width
	}
	return width, height, nil
}

// textRunSource is one text element found while walking the document's
// bands, together with the resolved face name it needs.
type textRunSource struct {
	face     string
	text     string
	x, y     geom.Length
	fontSize geom.Length
}

// collectTextRuns walks every band (content, pageHeader, pageFooter) in
// authored order and returns one textRunSource per non-empty text
// element, resolving each element's face via its style's fontFamily
// against the document's font fallback chains and the caller's FontSet
// — the first chain entry present in fs wins. AC9's "union of glyphs
// the document uses" is exactly the set of runes across everything this
// function returns.
//
// Each text element's authored value is first resolved against data
// via internal/bind.BindText (AC15-AC21, D-1.6.5): this is the ONLY
// site that calls BindText, so AC20's field-scope fence — bind text
// interpolation applies to text-element `value` only, never
// table.bind/columns[].bind — holds by construction rather than by a
// document-wide scan that would also visit those other fields (M-4:
// the canonical golden fixture's `columns[].bind` contains an
// expression-shaped `{{formatNumber(...)}}` that is deliberately not
// this story's business).
func collectTextRuns(doc *Template, data bind.Value, fs FontSet) ([]textRunSource, error) {
	_, height, err := pageDimensions(doc)
	if err != nil {
		return nil, err
	}
	usableHeight := height - doc.Page.Margin.Top - doc.Page.Margin.Bottom

	var headerHeight, footerHeight geom.Length
	if doc.Bands.PageHeader.Height.Set && !doc.Bands.PageHeader.Height.Null {
		headerHeight = doc.Bands.PageHeader.Height.Value
	}
	if doc.Bands.PageFooter.Height.Set && !doc.Bands.PageFooter.Height.Null {
		footerHeight = doc.Bands.PageFooter.Height.Value
	}

	// bandOrigin is PROVISIONAL (AC28): the vertical offset, from the
	// page's top printable edge, at which each band's own (element-
	// relative) Y=0 sits — pageHeader at the top, content directly below
	// it, pageFooter flush against the bottom margin. Real band
	// composition (stacking bands that actually grow with content, page
	// breaks, etc.) is Story 2.5's (internal/layout) job; this is only
	// enough to keep this story's fixtures from overlapping.
	bands := []struct {
		band   template.Band
		origin geom.Length
	}{
		{doc.Bands.PageHeader, 0},
		{doc.Bands.Content, headerHeight},
		{doc.Bands.PageFooter, usableHeight - footerHeight},
	}

	var runs []textRunSource
	for _, b := range bands {
		for _, el := range b.band.Elements {
			if el.Type != template.ElementText {
				continue
			}
			if !el.Value.Set || el.Value.Null || el.Value.Value == "" {
				continue
			}
			boundText, berr := bind.BindText(el.Value.Value, data, string(el.ID))
			if berr != nil {
				return nil, fmt.Errorf("folio: Render: %w", berr)
			}
			// QA Finding 5 (this story's review, Major): resolveFace
			// must run BEFORE the AC9 empty-text short-circuit below,
			// not after it. The previous ordering let boundText == ""
			// skip font-chain validation entirely, so an element with
			// an unresolvable style.fontFamily chain (Story 1.5
			// AC2/AC4's located error) rendered successfully whenever
			// its bound value happened to be null or "" — the SAME
			// broken template passing or failing depending on which
			// report it was handed. AC9 only requires that a null
			// binding "renders as empty, and is not an error"; it does
			// not license skipping the element's own validation.
			face, err := resolveFace(doc, el, fs)
			if err != nil {
				return nil, fmt.Errorf("folio: Render: element %s: %w", el.ID, err)
			}
			if boundText == "" {
				// AC9: a placeholder resolving to explicit JSON null
				// renders as empty — nothing left to draw for this run
				// — but the element's own validation above still ran.
				continue
			}
			fontSize := defaultFontSizePt
			if el.Style.Set && !el.Style.Null && el.Style.Value.FontSize.Set && !el.Style.Value.FontSize.Null {
				fontSize = el.Style.Value.FontSize.Value
			}
			runs = append(runs, textRunSource{
				face:     face,
				text:     boundText,
				x:        el.X,
				y:        b.origin + el.Y,
				fontSize: fontSize,
			})
		}
	}
	return runs, nil
}

// resolveFace resolves one text element's face name: style.fontFamily
// names a fallback CHAIN (doc.Fonts[chain] is an ordered list of face
// names); the first face name present as a key in fs wins. The engine
// never queries the host — a chain with no member present in fs, or an
// element with no fontFamily at all, is a located error naming the
// element (AC2, AC4).
func resolveFace(doc *Template, el template.Element, fs FontSet) (string, error) {
	if !el.Style.Set || el.Style.Null || !el.Style.Value.FontFamily.Set || el.Style.Value.FontFamily.Null {
		return "", fmt.Errorf("has text but no style.fontFamily to resolve a font from")
	}
	chainName := el.Style.Value.FontFamily.Value
	chain, ok := doc.Fonts[chainName]
	if !ok || len(chain) == 0 {
		return "", fmt.Errorf("style.fontFamily %q names a chain with no entries in the document's fonts map", chainName)
	}
	for _, face := range chain {
		if _, present := fs[face]; present {
			return face, nil
		}
	}
	return "", fmt.Errorf("no face in chain %q (%v) is present in the supplied FontSet", chainName, chain)
}

// renderDocument is Render's implementation once t is known non-nil
// (AC14b). It resolves every text element's face, subsets each distinct
// face EXACTLY ONCE over the union of runes the whole document uses
// (AC9), and hands the result to internal/pdf.SerializeTextDocument.
func renderDocument(t *Template, data bind.Value, fs FontSet) ([]byte, error) {
	runs, err := collectTextRuns(t, data, fs)
	if err != nil {
		return nil, err
	}

	// Union of runes per face, across the WHOLE document (AC9) — built
	// by ranging `runs` (a slice), never a map, at the collection site.
	runesByFace := map[string][]rune{}
	for _, r := range runs {
		runesByFace[r.face] = append(runesByFace[r.face], []rune(r.text)...)
	}

	faceNames := slices.Sorted(maps.Keys(runesByFace)) // ScanMapRange-compliant: sorted, deterministic object order.

	embedded := make(map[string]pdf.EmbeddedFace, len(faceNames))
	for _, name := range faceNames {
		data, ok := fs[name]
		if !ok {
			return nil, fmt.Errorf("folio: Render: face %q was resolved from a fallback chain but is missing from the FontSet", name)
		}
		font, ferr := fontset.New(name, data)
		if ferr != nil {
			return nil, fmt.Errorf("folio: Render: %w", ferr)
		}
		// ONE subsetting call per font per document (AC9), over the
		// union of runes collected above.
		sub, serr := font.Subset(runesByFace[name])
		if serr != nil {
			return nil, fmt.Errorf("folio: Render: %w", serr)
		}
		metrics := font.Metrics()
		created, modified := font.HeadTimes()
		embedded[name] = pdf.EmbeddedFace{
			Name:          name,
			Program:       sub.Program,
			Tag:           sub.Tag,
			NumGlyphs:     sub.NumGlyphs,
			GlyphForRune:  sub.GlyphForRune,
			WidthForGlyph: sub.WidthForGlyph,
			Ascent:        metrics.Ascent,
			Descent:       metrics.Descent,
			CapHeight:     metrics.CapHeight,
			BBoxXMin:      metrics.BBoxXMin,
			BBoxYMin:      metrics.BBoxYMin,
			BBoxXMax:      metrics.BBoxXMax,
			BBoxYMax:      metrics.BBoxYMax,
			HeadCreated:   created,
			HeadModified:  modified,
		}
	}

	width, height, perr := pageDimensions(t)
	if perr != nil {
		return nil, perr
	}

	pdfRuns := make([]pdf.TextRun, len(runs))
	for i, r := range runs {
		pdfRuns[i] = pdf.TextRun{Face: r.face, Text: r.text, X: r.x, Y: r.y, FontSize: r.fontSize}
	}

	page := pdf.TextPage{
		Runs:       pdfRuns,
		Width:      width,
		Height:     height,
		MarginTop:  t.Page.Margin.Top,
		MarginLeft: t.Page.Margin.Left,
	}

	return pdf.SerializeTextDocument([]pdf.TextPage{page}, embedded)
}
