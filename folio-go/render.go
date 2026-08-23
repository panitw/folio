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
	case doc.doc.Page.SizeIsName && doc.doc.Page.SizeName == "A4":
		width, height = pageWidthA4, pageHeightA4
	case !doc.doc.Page.SizeIsName:
		width, height = doc.doc.Page.SizeCustom.Width, doc.doc.Page.SizeCustom.Height
	default:
		return 0, 0, fmt.Errorf(
			"folio: Render: page.size names %q, which this version does not implement dimensions "+
				"for (only \"A4\" or a custom width/height)", doc.doc.Page.SizeName,
		)
	}
	if doc.doc.Page.Orientation == "landscape" {
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

// bandWithOrigin pairs one of the document's three bands with the
// PROVISIONAL vertical offset (AC28) at which its own element-relative
// Y=0 sits — the placement math both collectTextRuns and
// collectImageRuns need, factored once so the two element kinds agree
// on where a band starts.
type bandWithOrigin struct {
	band   template.Band
	origin geom.Length
}

// documentBands resolves pageDimensions and returns the three bands
// with their origins, in authored (header, content, footer) order.
func documentBands(doc *Template) ([]bandWithOrigin, error) {
	_, height, err := pageDimensions(doc)
	if err != nil {
		return nil, err
	}
	usableHeight := height - doc.doc.Page.Margin.Top - doc.doc.Page.Margin.Bottom

	var headerHeight, footerHeight geom.Length
	if doc.doc.Bands.PageHeader.Height.Set && !doc.doc.Bands.PageHeader.Height.Null {
		headerHeight = doc.doc.Bands.PageHeader.Height.Value
	}
	if doc.doc.Bands.PageFooter.Height.Set && !doc.doc.Bands.PageFooter.Height.Null {
		footerHeight = doc.doc.Bands.PageFooter.Height.Value
	}

	return []bandWithOrigin{
		{doc.doc.Bands.PageHeader, 0},
		{doc.doc.Bands.Content, headerHeight},
		{doc.doc.Bands.PageFooter, usableHeight - footerHeight},
	}, nil
}

// imageRunSource is one image element found while walking the
// document's bands: its declared BOX (before fit/centre), and the asset
// key it references.
type imageRunSource struct {
	elementID  string
	assetKey   string
	x, y       geom.Length
	boxW, boxH geom.Length
}

// collectImageRuns walks every band in authored order and returns one
// imageRunSource per image element (AD-24, source AC3/AC4). It does not
// decode or validate the referenced asset — that is
// resolveImagePlacement's job (renderDocument), called once the union
// of images the whole document uses is known, mirroring collectTextRuns/
// resolveFace's split.
func collectImageRuns(doc *Template) ([]imageRunSource, error) {
	bands, err := documentBands(doc)
	if err != nil {
		return nil, err
	}

	var runs []imageRunSource
	for _, b := range bands {
		for _, el := range b.band.Elements {
			if el.Type != template.ElementImage {
				continue
			}
			if !el.Width.Set || !el.Height.Set {
				return nil, fmt.Errorf("folio: Render: element %s: image element has no declared width/height box", el.ID)
			}
			if !el.Asset.Set {
				// M-2: parse_bands.go already makes a missing asset on
				// an image element a load error, so this is
				// unreachable for any successfully parsed Document —
				// handled rather than assumed.
				return nil, fmt.Errorf("folio: Render: element %s: image element has no asset", el.ID)
			}
			runs = append(runs, imageRunSource{
				elementID: string(el.ID),
				assetKey:  el.Asset.Value,
				x:         el.X,
				y:         b.origin + el.Y,
				boxW:      el.Width.Value,
				boxH:      el.Height.Value,
			})
		}
	}
	return runs, nil
}

// resolveImagePlacement computes AC13/AC14's fit-and-centre geometry for
// one image run: the binding axis is chosen by CROSS-MULTIPLICATION
// (bw*H vs bh*W — exact integer comparison, no division, no float), then
// exactly one geom.ScaleRound call computes the free axis, and the
// centring offsets are a SECOND ScaleRound call each
// (ScaleRound(box-drawn, 1, 2)) rather than "/2" — D-1.8.4: "(bw-dw)/2
// truncates when the difference is odd... route it through the same
// function, so round-half-to-even applies and the program keeps exactly
// one rounding mode in exactly one function."
func resolveImagePlacement(run imageRunSource, img template.DecodedImage) (drawX, drawY, drawW, drawH geom.Length) {
	bw, bh := run.boxW, run.boxH
	w, h := int64(img.Width()), int64(img.Height())

	// AC13: compare bw*H against bh*W, exact integer, no division.
	if bw*geom.Length(h) <= bh*geom.Length(w) {
		// width binds
		drawW = bw
		drawH = geom.ScaleRound(bw, h, w)
	} else {
		// height binds
		drawH = bh
		drawW = geom.ScaleRound(bh, w, h)
	}

	offsetX := geom.ScaleRound(bw-drawW, 1, 2)
	offsetY := geom.ScaleRound(bh-drawH, 1, 2)

	drawX = run.x + offsetX
	drawY = run.y + offsetY
	return drawX, drawY, drawW, drawH
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
func collectTextRuns(doc *Template, data, params bind.Value, fs FontSet) ([]textRunSource, error) {
	// bandOrigin is PROVISIONAL (AC28): the vertical offset, from the
	// page's top printable edge, at which each band's own (element-
	// relative) Y=0 sits — pageHeader at the top, content directly below
	// it, pageFooter flush against the bottom margin. Real band
	// composition (stacking bands that actually grow with content, page
	// breaks, etc.) is Story 2.5's (internal/layout) job; this is only
	// enough to keep this story's fixtures from overlapping.
	// (documentBands, factored above so collectImageRuns agrees on the
	// same band origins.)
	bands, err := documentBands(doc)
	if err != nil {
		return nil, err
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
			boundText, berr := bind.BindText(el.Value.Value, data, params, string(el.ID))
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
// names a fallback CHAIN (doc.doc.Fonts[chain] is an ordered list of face
// names); the first face name present as a key in fs wins. The engine
// never queries the host — a chain with no member present in fs, or an
// element with no fontFamily at all, is a located error naming the
// element (AC2, AC4).
func resolveFace(doc *Template, el template.Element, fs FontSet) (string, error) {
	if !el.Style.Set || el.Style.Null || !el.Style.Value.FontFamily.Set || el.Style.Value.FontFamily.Null {
		return "", fmt.Errorf("has text but no style.fontFamily to resolve a font from")
	}
	chainName := el.Style.Value.FontFamily.Value
	chain, ok := doc.doc.Fonts[chainName]
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
func renderDocument(t *Template, data, params bind.Value, fs FontSet) ([]byte, error) {
	runs, err := collectTextRuns(t, data, params, fs)
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

	imageRuns, ierr := collectImageRuns(t)
	if ierr != nil {
		return nil, ierr
	}

	// AC9's "one XObject per asset per document" (the same shape as
	// fonts' "one subset per font per document"): dedup by asset key —
	// decode each DISTINCT referenced asset exactly once, regardless of
	// how many elements place it.
	pdfImages := make(map[string]pdf.ImageXObject, len(imageRuns))
	decodedByKey := make(map[string]template.DecodedImage, len(imageRuns))
	assetKeys := make([]string, 0, len(imageRuns))
	// firstElementIDByAssetKey carries FIRST (in imageRuns' authored
	// order) referencing element id alongside each distinct asset key,
	// so a render-time error can name the element that caused it (D-1.8.1
	// amended's binding verdict-table clause: "Located, naming element
	// id, asset key and media type" — Finding 4, Story 1.8 review: this
	// was previously hard-coded to "", producing a visible hole in the
	// error message instead of the element id).
	firstElementIDByAssetKey := make(map[string]string, len(imageRuns))
	seenAssetKey := map[string]bool{}
	for _, r := range imageRuns {
		if seenAssetKey[r.assetKey] {
			continue
		}
		seenAssetKey[r.assetKey] = true
		assetKeys = append(assetKeys, r.assetKey)
		firstElementIDByAssetKey[r.assetKey] = r.elementID
	}
	slices.Sort(assetKeys)
	for _, key := range assetKeys {
		asset, ok := t.doc.Assets[key]
		if !ok {
			return nil, fmt.Errorf("folio: Render: an image element references asset %q, which is not present in the document's assets map", key)
		}
		raw, derr := template.DecodeAssetBytes(asset)
		if derr != nil {
			return nil, fmt.Errorf("folio: Render: asset %q: %w", key, derr)
		}
		img, derr := template.DecodeImageForRender(asset.MediaType, raw, key, firstElementIDByAssetKey[key])
		if derr != nil {
			return nil, fmt.Errorf("folio: Render: %w", derr)
		}
		decodedByKey[key] = img
		pdfImages[key] = pdf.ImageXObject{
			Width:            img.Width(),
			Height:           img.Height(),
			ColorSpace:       img.ColorSpace,
			BitsPerComponent: img.BitsPerComponent,
			Filter:           img.Filter,
			HasDecodeParms:   img.HasDecodeParms,
			PredictorColors:  img.PredictorColors,
			PredictorBPC:     img.PredictorBPC,
			PredictorColumns: img.PredictorColumns,
			Stream:           img.Stream,
		}
	}

	pdfPlacements := make([]pdf.ImagePlacement, len(imageRuns))
	for i, r := range imageRuns {
		img := decodedByKey[r.assetKey]
		drawX, drawY, drawW, drawH := resolveImagePlacement(r, img)
		pdfPlacements[i] = pdf.ImagePlacement{
			ResourceName: r.assetKey,
			X:            drawX,
			Y:            drawY,
			DrawWidth:    drawW,
			DrawHeight:   drawH,
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
		Images:     pdfPlacements,
		Width:      width,
		Height:     height,
		MarginTop:  t.doc.Page.Margin.Top,
		MarginLeft: t.doc.Page.Margin.Left,
	}

	return pdf.SerializeTextDocument([]pdf.TextPage{page}, embedded, pdfImages)
}
