package folio

import (
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/panitw/folio/folio-go/internal/bind"
	"github.com/panitw/folio/folio-go/internal/fontset"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/pagemodel"
	"github.com/panitw/folio/folio-go/internal/pdf"
	"github.com/panitw/folio/folio-go/internal/template"
	"github.com/panitw/folio/folio-go/internal/text"
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

// pageDimensions resolves a Document's page geometry (band composition
// needs page height and margins; see pageGeometryOf, which is the only
// caller that reaches internal/layout with them). An unrecognised named size is a
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
// bands, together with the resolved face name it needs and the SHAPED
// answer for its text.
//
// glyphs/clusterTexts are populated by positionSegments, in the same
// pass that computes x, and that co-location is the point rather than a
// convenience: x is derived from the very glyphs carried alongside it
// (positionSegments' cursor sums faceSegment.advance1000 over the same
// slice it hands to the run), so a run cannot be drawn from one shaping
// answer and positioned from another. Story 2.3's finisher, Blocker 1 — the previous arrangement
// shaped in renderDocument and summed raw `hmtx` advances here, which
// drew kerned text at unkerned origins.
type textRunSource struct {
	face     string
	text     string
	x, y     geom.Length
	fontSize geom.Length

	// baselineOffset is top-of-line -> baseline for this run: the ruled
	// model's first span, max(hhea ascent) over the DECLARED chain,
	// scaled to fontSize (D-2.4.2 as amended). It is a LAYOUT quantity
	// and it is resolved here, in package folio, because AD-5 keeps
	// placement decisions out of every renderer: internal/pdf must not
	// re-derive it from faces[run.Face].
	//
	// It is identical for every run of one element — every face segment
	// and every line — because it is a function of the chain and the
	// size and NOT of what was drawn. Deriving it per-run from the
	// resolved face would make it content-dependent, which AD-24 rules
	// out for exactly the reason it rules it out for the advance:
	// adding one CJK character would reflow the element.
	baselineOffset geom.Length

	// glyphs is the shaper's answer for text, in FONT UNITS and drawing
	// order. clusterTexts is its per-glyph /ToUnicode source text
	// (text.ClusterTexts), parallel to glyphs.
	glyphs       []text.ShapedGlyph
	clusterTexts []string

	// --- Story 2.6: which band, and which atomic column item ---
	//
	// band is the index into documentBands' authored order: 0 pageHeader,
	// 1 content, 2 pageFooter. Only the CONTENT band paginates; the other
	// two are repeated verbatim on every page (AC3), which is what makes
	// page 34 as complete as page 1.
	band int

	// elementID names the element this run came from, for the located
	// overflow diagnostic ONLY (layout.OverflowError). No geometry is
	// derived from it.
	elementID string

	// lineIndex is the run's line within its element. Together with
	// elementID it identifies the atomic COLUMN ITEM this run belongs to:
	// one line's runs are contiguous in the slice and share both values,
	// so grouping is a scan for a change of key rather than a map.
	lineIndex int

	// itemTop / itemBottom are the LINE's vertical extent, page-absolute:
	// `baseline − max(ascent)` .. `baseline + max(descent)` over the
	// element's DECLARED chain (D-2.4.2 as amended). Carried rather than
	// re-derived downstream — a second derivation of this number is
	// precisely what that amendment exists to prevent.
	itemTop, itemBottom geom.Length

	// --- Story 2.7: AD-4's late-bound page-number slot ---
	//
	// pageSlots mirrors pagemodel.PageNumberSlot's GlyphLo/GlyphHi/
	// DigitsY, in the FONT-UNIT glyph slice this run carries before
	// buildShapedPDFRuns converts it (renderDocument attaches the
	// pagemodel.PageNumberSlot values themselves, once CIDs and
	// 1000-em advances exist). Populated only by positionSegments, when
	// it is handed a non-empty slots argument — every other call site
	// leaves this nil, so this is additive and changes no existing run.
	//
	// A SLICE (this story's review, Blocker 1): a run may carry more
	// than one {{page}} occurrence — see pagemodel.TextRun.PageSlots'
	// doc comment for why a scalar field here was a silent mis-render.
	pageSlots []textRunPageSlot

	// --- Story 2.8: FR44's clip, D-2.8.1 ---
	//
	// clipToBox marks that this run belongs to a text element whose
	// widest packed line exceeds its declared WIDTH (detectWidthOverflow,
	// below) — never its declared height, which D-2.8.1 rules is not a
	// clip bound at all. clipX/clipWidth are that element's declared box
	// left edge and width, PAGE-ABSOLUTE exactly like x above.
	// internal/pdf uses them to wrap this run's drawing operators in a
	// PDF clip path restricted to the box's HORIZONTAL extent only — the
	// vertical clip bound it uses is the full page, never anything
	// derived from an element's declared height (AC3).
	clipToBox        bool
	clipX, clipWidth geom.Length
}

// textRunPageSlot is one {{page}} reservation's glyph range within the
// carrying textRunSource — positionSegments' own coordinates (local to
// this run, before buildShapedPDFRuns reindexes into the document's
// final CID space). Mirrors pagemodel.PageNumberSlot's GlyphLo/GlyphHi/
// DigitsY exactly; the CID table and per-digit advance are attached
// later, in renderDocument, once pass one has allocated them.
type textRunPageSlot struct {
	glyphLo, glyphHi int
	digitsY          int
}

// bandWithOrigin pairs one of the document's three bands with the
// PAGE-ABSOLUTE vertical offset at which its own element-relative Y=0
// sits — the placement both collectTextRuns and collectImageRuns need,
// factored once so the two element kinds agree on where a band starts.
//
// The origin itself is NOT computed here. Under AD-24 "bands are placed
// on the page by internal/layout alone", so this file resolves the page
// SETUP and hands it to internal/layout, which answers with the three
// origins. Package folio computes no band origin
// (TestNoBandOriginArithmeticInPackageFolio, internal/bandcomposition_arch_test.go).
type bandWithOrigin struct {
	band   template.Band
	origin geom.Length
}

// pageGeometryOf reads the document's page setup into internal/layout's
// closed input struct. Every geometric input band composition is allowed
// to see passes through here, and nothing else does: an element, or a
// measurement of one, has no route into PageGeometry, which is AD-24's
// "nothing negotiates" holding structurally rather than by convention.
//
// A band's height key that is absent or explicitly null reads as zero —
// the same treatment the pre-2.5 code gave it, unchanged.
func pageGeometryOf(doc *Template) (layout.PageGeometry, error) {
	width, height, err := pageDimensions(doc)
	if err != nil {
		return layout.PageGeometry{}, err
	}
	g := layout.PageGeometry{
		Width:        width,
		Height:       height,
		MarginTop:    doc.doc.Page.Margin.Top,
		MarginBottom: doc.doc.Page.Margin.Bottom,
		MarginLeft:   doc.doc.Page.Margin.Left,
		MarginRight:  doc.doc.Page.Margin.Right,
	}
	if doc.doc.Bands.PageHeader.Height.Set && !doc.doc.Bands.PageHeader.Height.Null {
		g.PageHeaderHeight = doc.doc.Bands.PageHeader.Height.Value
	}
	if doc.doc.Bands.PageFooter.Height.Set && !doc.doc.Bands.PageFooter.Height.Null {
		g.PageFooterHeight = doc.doc.Bands.PageFooter.Height.Value
	}
	return g, nil
}

// documentBands returns the three bands with their origins, in authored
// (header, content, footer) order, asking internal/layout for the
// origins rather than deriving them.
func documentBands(doc *Template) ([]bandWithOrigin, error) {
	g, err := pageGeometryOf(doc)
	if err != nil {
		return nil, err
	}
	origins := layout.Origins(g)
	return []bandWithOrigin{
		{doc.doc.Bands.PageHeader, origins.PageHeader},
		{doc.doc.Bands.Content, origins.Content},
		{doc.doc.Bands.PageFooter, origins.PageFooter},
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

	// band is documentBands' authored index: 0 pageHeader, 1 content,
	// 2 pageFooter. Story 2.6 paginates the content band alone.
	band int
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
	for bandIndex, b := range bands {
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
				y:         layout.PlaceInBand(b.origin, el.Y),
				boxW:      el.Width.Value,
				boxH:      el.Height.Value,
				band:      bandIndex,
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
func collectTextRuns(doc *Template, data, params bind.Value, fs FontSet, cache *fontCache) ([]textRunSource, error) {
	// documentBands asks internal/layout where each band's own
	// (element-relative) Y=0 sits on the page: pageHeader at the top,
	// content directly below it, pageFooter starting exactly where the
	// content band ends. Every element's page Y below is
	// layout.PlaceInBand(origin, el.Y) — a TRANSLATION, never an
	// inversion (D-2.0.4, AD-24).
	bands, err := documentBands(doc)
	if err != nil {
		return nil, err
	}

	var runs []textRunSource
	for bandIndex := range bands {
		// Diagnostics discarded here: collectTextRuns is the pre-2.7
		// legacy, all-bands walker kept alive only for a handful of
		// composition tests (collect_text_runs_composition_test.go,
		// shaped_fixture_test.go) that assert on run geometry, not on
		// overflow. renderDocument (below) is the one path that
		// actually surfaces Story 2.8's Diagnostics through Render.
		bandRuns, _, _, berr := collectBandTextRuns(doc, bands, bandIndex, data, params, fs, cache, passthroughResolver)
		if berr != nil {
			return nil, berr
		}
		runs = append(runs, bandRuns...)
	}
	return runs, nil
}

// elementTokenResolver decides what a text element's bound text — ALREADY
// through bind.BindTextSpans, so the only "{{…}}" tokens still literal
// are {{page}} and {{pages}} (internal/bind/text.go:45-50's reservation)
// — becomes before shaping. It is the seam Story 2.7 uses to give the
// content band and the two repeated bands different answers to "what
// does {{page}} mean here" without duplicating collectBandTextRuns'
// shaping/packing/positioning body (D-000.42).
type elementTokenResolver func(elementID, boundText string, subs []bind.Substitution) (resolvedText string, resolvedSubs []bind.Substitution, slots []pageSlotSpan, err error)

// passthroughResolver is collectTextRuns' resolver: {{page}}/{{pages}}
// pass through exactly as bind.BindTextSpans left them, unchanged. This
// is BYTE-FOR-BYTE today's pre-Story-2.7 behaviour — collectTextRuns
// itself never learns a page count and never needs to.
func passthroughResolver(_, boundText string, subs []bind.Substitution) (string, []bind.Substitution, []pageSlotSpan, error) {
	return boundText, subs, nil, nil
}

// contentBandResolver is D-2.7.3's fence: {{page}}/{{pages}} resolve
// ONLY in the page-header and page-footer bands. In the content band
// the construct is a located template error naming the element — not a
// silent literal — because content-band Y depends on content-band
// layout, which depends on this construct's width: a fixed point AD-24
// forbids negotiating.
func contentBandResolver(elementID, boundText string, subs []bind.Substitution) (string, []bind.Substitution, []pageSlotSpan, error) {
	if name, found := firstReservedPageToken(boundText, subs); found {
		return "", nil, nil, fmt.Errorf(
			"folio: Render: element %s: {{%s}} resolves only in the page header and page footer bands "+
				"(D-2.7.3) — this element is in the content band, where the page count the construct "+
				"needs depends on this band's own layout, which AD-24 does not permit resolving by "+
				"negotiation",
			elementID, name,
		)
	}
	return boundText, subs, nil, nil
}

// headerFooterResolver is D-2.7.2's reservation, built once pageCount
// (Y) is known: {{pages}} becomes Y's exact digits (no reservation — Y
// is the same value on every page) and {{page}} becomes a
// digits(Y)-wide filler reservation, reported as a pageSlotSpan for
// positionSegments to mark.
func headerFooterResolver(pageCount int) elementTokenResolver {
	return func(elementID, boundText string, subs []bind.Substitution) (string, []bind.Substitution, []pageSlotSpan, error) {
		if _, found := firstReservedPageToken(boundText, subs); !found {
			return boundText, subs, nil, nil
		}
		resolved, slots, repl := resolvePageTokens(boundText, pageCount, subs)
		return resolved, shiftSubstitutions(subs, repl), slots, nil
	}
}

// collectBandTextRuns is collectTextRuns' body, generalised over ONE
// band and one elementTokenResolver — the single implementation
// collectTextRuns (legacy, all bands, pass-through) and Story 2.7's
// two-phase pipeline (content-only, then header/footer-once-Y-is-known)
// both drive, so the shaping/packing/positioning sequence exists in
// exactly one place (D-000.42).
//
// It returns, alongside the band's runs, one pendingPageSlot per
// {{page}} occurrence resolved in this call — indices LOCAL to the
// returned runs slice; the caller shifts them once the final combined
// run order is known.
func collectBandTextRuns(
	doc *Template,
	bands []bandWithOrigin,
	bandIndex int,
	data, params bind.Value,
	fs FontSet,
	cache *fontCache,
	resolve elementTokenResolver,
) ([]textRunSource, []pendingPageSlot, []Diagnostic, error) {
	b := bands[bandIndex]
	var runs []textRunSource
	var pending []pendingPageSlot
	// diags accumulates in ELEMENT DECLARATION ORDER within this one
	// band — the `for _, el := range b.band.Elements` loop below walks
	// the authored `.folio` document in order, never a map, so this
	// slice is already the order D-2.8.6's Result.Diagnostics doc
	// comment requires WITHIN one band. renderDocument concatenates
	// this band's diags after the header band's and before the footer
	// band's to get full document order across all three.
	var diags []Diagnostic

	for _, el := range b.band.Elements {
		if el.Type != template.ElementText {
			continue
		}
		if !el.Value.Set || el.Value.Null || el.Value.Value == "" {
			continue
		}
		boundText, subs, berr := bind.BindTextSpans(el.Value.Value, data, params, string(el.ID))
		if berr != nil {
			return nil, nil, nil, fmt.Errorf("folio: Render: %w", berr)
		}

		resolvedText, resolvedSubs, slots, rerr := resolve(string(el.ID), boundText, subs)
		if rerr != nil {
			return nil, nil, nil, rerr
		}
		boundText, subs = resolvedText, resolvedSubs

		// QA Finding 5 (this story's review, Major): the fontFamily
		// chain must be validated BEFORE the AC9 empty-text
		// short-circuit below, not after it. The previous ordering
		// let boundText == "" skip font-chain validation entirely,
		// so an element with an unresolvable style.fontFamily chain
		// (Story 1.5 AC2/AC4's located error) rendered successfully
		// whenever its bound value happened to be null or "" — the
		// SAME broken template passing or failing depending on
		// which report it was handed. AC9 only requires that a null
		// binding "renders as empty, and is not an error"; it does
		// not license skipping the element's own validation.
		chain, err := fontChain(doc, el)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("folio: Render: element %s: %w", el.ID, err)
		}
		if boundText == "" {
			// AC9: a placeholder resolving to explicit JSON null
			// renders as empty — nothing left to draw for this run
			// — but the element's fontFamily chain still validated
			// above. There is no text, so there is nothing to check
			// coverage against (Story 2.2, AC4).
			continue
		}
		fontSize := defaultFontSizePt
		if el.Style.Set && !el.Style.Null && el.Style.Value.FontSize.Set && !el.Style.Value.FontSize.Null {
			fontSize = el.Style.Value.FontSize.Value
		}
		// Story 2.2, AC4: COVERAGE-based resolution, per rune,
		// across the chain — never "first chain member present in
		// fs" (the pre-Story-2.2 reading). May split one element
		// into several face segments, one per contiguous run of
		// runes sharing the same resolved face. Shaped ONCE here;
		// every line below is a SLICE of these glyphs, never a
		// re-shape of a shorter string (Story 2.4, AC10).
		segs, serr := shapeSegments(chain, boundText, fs, cache)
		if serr != nil {
			return nil, nil, nil, fmt.Errorf("folio: Render: element %s: %w", el.ID, serr)
		}
		totalRunes := len([]rune(boundText))

		// Story 2.4: where may this element break, and what may it
		// not break inside? The atomic spans are the document's
		// declared unbreakableValues matched against the rune spans
		// bind.BindTextSpans reported — handed to internal/text as a
		// PARAMETER (D-000.16), never through an import.
		atomic := atomicSpansFor(doc.doc.UnbreakableValues, subs)
		ops := text.Opportunities(text.Dictionary(), boundText, atomic)

		boxWidth := geom.Length(0)
		if el.Width.Set && !el.Width.Null {
			boxWidth = el.Width.Value
		}
		lines := packLines(segs, ops, totalRunes, fontSize, boxWidth)

		// Story 2.8, AC1/D-2.8.1: does this element's widest packed
		// line exceed its declared WIDTH? Computed from the SAME
		// wrappedLine.width the packer already measured — never a
		// second measurement of the text — and against boxWidth alone,
		// never el.Height (D-2.8.1: a text element's declared height is
		// not a clip bound and this function never reads el.Height).
		overflow, overflows := detectWidthOverflow(string(el.ID), lines, boxWidth)
		if overflows {
			diags = append(diags, Diagnostic{
				Severity:  SeverityWarning,
				Code:      DiagCodeTextClippedWidth,
				ElementID: overflow.elementID,
				Message: fmt.Sprintf(
					"element %s: the widest laid-out line is %s wide, exceeding the element's declared "+
						"width of %s; the overflowing content is clipped at the box's left/right edges, "+
						"never reflowed and never dropped (FR44)",
					overflow.elementID, millipoints(overflow.measuredWidth), millipoints(overflow.declaredWidth),
				),
			})
		}

		// ONE vertical model for the element, from its DECLARED
		// chain (D-2.4.2 as amended): computed once, outside the
		// loop, from ONE walk of the chain, because every span of it
		// is a function of the chain and the size and of nothing on
		// any individual line.
		//
		// UNCONDITIONAL, where the superseded code computed the
		// advance only when len(lines) > 1. The first-baseline
		// offset is needed by EVERY element with at least one line,
		// so this call now runs for single-line elements too — which
		// widens the set of inputs that can reach verticalModel's two
		// error paths. That widening is measured rather than assumed:
		// see TestVerticalModelErrorPathsAreUnreachableThroughRender.
		vm, serr := chainVerticalModel(chain, fontSize, fs, cache)
		if serr != nil {
			return nil, nil, nil, fmt.Errorf("folio: Render: element %s: %w", el.ID, serr)
		}

		elementY := layout.PlaceInBand(b.origin, el.Y)
		startPending := len(pending)
		for i, ln := range lines {
			lineY := elementY + geom.Length(int64(i))*vm.Advance
			placed, poserr := positionSegments(segs, ln.from, ln.to, el.X, lineY, fontSize, vm.FirstBaseline, slots)
			if poserr != nil {
				return nil, nil, nil, fmt.Errorf("folio: Render: element %s: %w", el.ID, poserr)
			}
			// Story 2.6: the LINE's extent, computed here from the
			// vertical model that is already in hand — `lineY` IS
			// `baseline − max(ascent)` because positionSegments
			// places the baseline at lineY + vm.FirstBaseline, and
			// vm.FirstBaseline IS max(ascent) scaled. The bottom
			// adds max(descent). Same numbers, no re-derivation.
			for j := range placed {
				placed[j].band = bandIndex
				placed[j].elementID = string(el.ID)
				placed[j].lineIndex = i
				placed[j].itemTop = lineY
				placed[j].itemBottom = lineY + vm.FirstBaseline + vm.LastDescent
				// Story 2.8, AC6: every run of an overflowing element
				// carries the SAME clip box (the element's declared
				// left edge and width) — clipping is a property of the
				// ELEMENT, not of any one line or face segment within
				// it, so a multi-line or multi-face-segment overflow
				// clips uniformly across all its runs.
				if overflows {
					placed[j].clipToBox = true
					placed[j].clipX = el.X
					placed[j].clipWidth = boxWidth
				}
				// Story 2.7 review, Blocker 1: ONE pendingPageSlot per
				// {{page}} occurrence this run carries, not one per
				// run — a run may carry more than one.
				for _, ps := range placed[j].pageSlots {
					pending = append(pending, pendingPageSlot{
						runIndex: len(runs) + j,
						glyphLo:  ps.glyphLo,
						glyphHi:  ps.glyphHi,
						digitsY:  ps.digitsY,
					})
				}
			}
			runs = append(runs, placed...)
		}

		// Story 2.7: this element used a {{page}} slot, so its face's
		// ten digits need CIDs allocated even though only ONE filler
		// digit was actually shaped above (finding 3, story creation:
		// digit identity never affects width, but substitution needs
		// EVERY digit's CID, since any of 0-9 may be a page's own).
		if len(slots) > 0 && len(pending) > startPending {
			dt, dterr := digitTableRun(chain, fontSize, fs, cache)
			if dterr != nil {
				return nil, nil, nil, dterr
			}
			dtIndex := len(runs)
			runs = append(runs, dt)
			for k := startPending; k < len(pending); k++ {
				pending[k].digitTableIndex = dtIndex
			}
		}
	}
	return runs, pending, diags, nil
}

// widthOverflow is Story 2.8, AC1's per-element horizontal overflow
// record: the element id, the declared bound and the measured extent
// (the widest packed line's width). The axis is always "width" —
// D-2.8.1 fences the vertical axis out entirely, so there is no axis
// field to carry. detectWidthOverflow is the SINGLE place this record
// is computed, driving both the emitted Diagnostic (AC1/AC7) and the
// clip decision (AC6): an element that does not overflow never reaches
// either.
type widthOverflow struct {
	elementID     string
	declaredWidth geom.Length
	measuredWidth geom.Length
}

// detectWidthOverflow reports whether lines' widest member exceeds
// boxWidth (D-2.8.1: the declared WIDTH is FR44's only clip bound; a
// text element's declared HEIGHT is read by nothing here, exactly as
// finding 1 measured it is read by nothing in the pre-2.8 renderer). A
// zero-or-negative boxWidth means "no declared width" — packLines'
// own convention (wrap.go: "el.Width.Set && !el.Width.Null" gates
// boxWidth above; an absent width is always boxWidth == 0) — and never
// overflows: there is no bound to measure against.
func detectWidthOverflow(elementID string, lines []wrappedLine, boxWidth geom.Length) (widthOverflow, bool) {
	if boxWidth <= 0 {
		return widthOverflow{}, false
	}
	widest := geom.Length(0)
	for _, ln := range lines {
		if ln.width > widest {
			widest = ln.width
		}
	}
	if widest <= boxWidth {
		return widthOverflow{}, false
	}
	return widthOverflow{elementID: elementID, declaredWidth: boxWidth, measuredWidth: widest}, true
}

// millipoints spells a geom.Length for a HUMAN-READABLE Diagnostic
// message (Story 2.8). Not an output-format emitter: nothing in a PDF,
// a page model or a golden passes through here, so AD-3's "one number
// emitter" (internal/pdf's numbers.go) is untouched — this mirrors
// internal/layout/paginate.go's own millipoints, kept package-local
// exactly as that one is, because the property it serves (a readable
// diagnostic) is local to each package that needs one.
func millipoints(v geom.Length) string {
	return strconv.FormatInt(int64(v), 10) + "mp"
}

// fontChain resolves one text element's style.fontFamily to its ordered
// fallback chain of face names (AD-8's Rule; AC3). It does not touch
// coverage — resolveRuneFace does that, per rune, against this chain.
func fontChain(doc *Template, el template.Element) ([]string, error) {
	if !el.Style.Set || el.Style.Null || !el.Style.Value.FontFamily.Set || el.Style.Value.FontFamily.Null {
		return nil, fmt.Errorf("has text but no style.fontFamily to resolve a font from")
	}
	chainName := el.Style.Value.FontFamily.Value
	chain, ok := doc.doc.Fonts[chainName]
	if !ok || len(chain) == 0 {
		return nil, fmt.Errorf("style.fontFamily %q names a chain with no entries in the document's fonts map", chainName)
	}
	return chain, nil
}

// fontCache parses a face's bytes into a *fontset.Font at most once per
// distinct face name, across the whole render (coverage checks touch
// every candidate face in a chain, not just the one ultimately used,
// so without this a long document would re-parse the same face
// repeatedly). Looked up and written only by key — NEVER ranged
// (ScanMapRange, D-2.2.3's whole-module scan).
type fontCache struct {
	byName map[string]*fontset.Font
}

func newFontCache() *fontCache {
	return &fontCache{byName: map[string]*fontset.Font{}}
}

// get parses and caches fs[name] on first use. A face NAMED in a chain
// but ABSENT from fs is reported here, once, the first time that chain
// entry is actually consulted — not a document-wide upfront validation
// pass, matching this package's existing "validate at the point of use"
// shape (resolveFace's prior behaviour).
func (c *fontCache) get(name string, fs FontSet) (*fontset.Font, error) {
	if f, ok := c.byName[name]; ok {
		return f, nil
	}
	data, ok := fs[name]
	if !ok {
		return nil, fmt.Errorf("face %q is not present in the supplied FontSet", name)
	}
	f, err := fontset.New(name, data)
	if err != nil {
		return nil, err
	}
	c.byName[name] = f
	return f, nil
}

// resolveRuneFace is AC4's COVERAGE-based resolution: walk chain in
// order and return the first face name that is BOTH present in fs AND
// whose cmap actually contains a glyph for r (never a proxy such as
// "the locale is ja" or "the face is not the preferred one for this
// script" — D-2.2-D4). A chain member absent from fs is skipped, not an
// error by itself — only "no member of the chain covers r" is.
func resolveRuneFace(chain []string, r rune, fs FontSet, cache *fontCache) (string, error) {
	for _, name := range chain {
		if _, present := fs[name]; !present {
			continue
		}
		f, err := cache.get(name, fs)
		if err != nil {
			return "", err
		}
		if f.HasGlyph(r) {
			return name, nil
		}
	}
	return "", fmt.Errorf("no font in chain %v has a glyph for rune %U — a located failure, not a blank box (AC4)", chain, r)
}

// shapeSegments performs Story 2.2's per-rune coverage resolution and
// Story 2.3's per-face-segment shaping, ONCE, and returns the result
// without positioning it.
//
// It is separated from positioning so that Story 2.4's line breaker
// measures and slices the SAME shaped glyphs that are ultimately drawn.
// Shaping once and slicing is not an optimisation — it is the
// correctness property: re-shaping a line's shorter text can
// legitimately produce different glyphs at the new boundary, and a
// second derivation of the same quantity is exactly what Story 2.3's
// Blocker 1 removed.
func shapeSegments(chain []string, elementText string, fs FontSet, cache *fontCache) ([]faceSegment, error) {
	type segment struct {
		face  string
		runes []rune
	}
	var segments []segment
	for _, r := range elementText {
		face, err := resolveRuneFace(chain, r, fs, cache)
		if err != nil {
			return nil, err
		}
		if n := len(segments); n > 0 && segments[n-1].face == face {
			segments[n-1].runes = append(segments[n-1].runes, r)
			continue
		}
		segments = append(segments, segment{face: face, runes: []rune{r}})
	}

	out := make([]faceSegment, 0, len(segments))
	runeStart := 0
	for _, seg := range segments {
		f, err := cache.get(seg.face, fs)
		if err != nil {
			return nil, err
		}

		// AC1/AC8: one buffer per FACE-SEGMENT, never per element and
		// never per document — GuessSegmentProperties derives script and
		// direction from the buffer's own contents, so a mixed-script
		// string would let one script's rules govern another's runes.
		// The segmentation above is Story 2.2's, unchanged.
		segText := string(seg.runes)
		glyphs, serr := f.Shaper().Shape(segText)
		if serr != nil {
			return nil, serr
		}
		texts, terr := text.ClusterTexts(segText, glyphs)
		if terr != nil {
			return nil, fmt.Errorf("face %q: %w", seg.face, terr)
		}

		out = append(out, faceSegment{
			face:         seg.face,
			segText:      segText,
			runeStart:    runeStart,
			runeEnd:      runeStart + len(seg.runes),
			glyphs:       glyphs,
			clusterTexts: texts,
			unitsPerEm:   int64(f.UnitsPerEm()),
		})
		runeStart += len(seg.runes)
	}
	return out, nil
}

// positionSegments turns the element-global rune range [from, to) of a
// shaped element into drawable runs, laying them out left to right from
// x at baseline y.
//
// Positioning: only the FIRST sub-run keeps x; each later sub-run's x is
// the previous one's x plus that sub-run's SHAPED total advance, scaled
// by fontSize. The cursor is advanced by measureRuneRange over exactly
// the glyphs the run carries — the same function the line breaker
// decides with — so a run's drawn width and its contribution to the
// cursor cannot disagree.
// slots is Story 2.7's addition: element-global rune spans of a
// {{page}} reservation within [from,to). Empty for every caller but the
// page-header/page-footer collection path, in which case this function
// is byte-for-byte what it was before this story — no extra allocation,
// no extra comparison beyond the nil check on the outer loop.
//
// A run's pageSlots is a SLICE (this story's review, Blocker 1): more
// than one {{page}} occurrence can land in one face segment on one
// line — "Page {{page}} of {{pages}} / {{page}}" is entirely ASCII, so
// it is a single run, and each matching slot is APPENDED, never
// overwritten.
//
// A slot that would straddle two face segments or two lines cannot be
// expressed as one contiguous glyph range and is therefore a located
// error naming the run's rune range, not a panic — this story's review,
// Finding 10: everywhere else on this path a structural impossibility
// is a located error (D-2.6.5's precedent; digitTableRun and
// buildPageNumberSlot in page_number.go both return one), and a public
// entry point (Render) must not let an internal panic cross it
// uncaught. Unreachable through the shipped set, where the construct is
// entirely ASCII (finding 5, story creation), but checked rather than
// assumed.
func positionSegments(segs []faceSegment, from, to int, x, y, fontSize, baselineOffset geom.Length, slots []pageSlotSpan) ([]textRunSource, error) {
	runs := make([]textRunSource, 0, len(segs))
	cursor := x
	for _, s := range segs {
		if to <= s.runeStart || from >= s.runeEnd {
			continue
		}
		elemLo := maxInt(from, s.runeStart)
		elemHi := minInt(to, s.runeEnd)
		lo, hi := s.glyphRangeForRunes(elemLo, elemHi)
		if hi <= lo {
			continue
		}
		runeLo := elemLo - s.runeStart
		runeHi := elemHi - s.runeStart
		run := textRunSource{
			face:           s.face,
			text:           string([]rune(s.segText)[runeLo:runeHi]),
			x:              cursor,
			y:              y,
			fontSize:       fontSize,
			baselineOffset: baselineOffset,
			glyphs:         s.glyphs[lo:hi],
			clusterTexts:   s.clusterTexts[lo:hi],
		}
		for _, sl := range slots {
			if sl.to <= elemLo || sl.from >= elemHi {
				continue // no overlap with this segment's contribution to this line
			}
			if sl.from < elemLo || sl.to > elemHi {
				return nil, fmt.Errorf(
					"folio: Render: internal error: a {{page}} reservation [%d,%d) straddles a "+
						"face-segment or line boundary at [%d,%d) — Story 2.7 requires the construct to "+
						"resolve to one face segment on one line",
					sl.from, sl.to, elemLo, elemHi,
				)
			}
			slotLo, slotHi := s.glyphRangeForRunes(sl.from, sl.to)
			run.pageSlots = append(run.pageSlots, textRunPageSlot{
				glyphLo: slotLo - lo,
				glyphHi: slotHi - lo,
				digitsY: sl.digitsY,
			})
		}
		runs = append(runs, run)
		cursor += geom.ScaleRound(geom.Length(s.advance1000(lo, hi)), int64(fontSize), 1000)
	}
	return runs, nil
}

// renderDocument is Render's implementation once t is known non-nil
// (AC14b). It resolves every text element's face, subsets each distinct
// face EXACTLY ONCE over the union of runes the whole document uses
// (AC9), and hands the result to internal/pdf.SerializeTextDocument.
//
// ERROR ORDERING ON A MULTI-DEFECT DOCUMENT (this story's review,
// Finding 9). This story's two-phase restructure — geometry, then
// images, then content-band text (D-2.7.3's fence), then
// layout.Paginate (D-2.6.5's OverflowError), then header/footer text
// (D-2.7.2's reservation) — changed WHICH located error a document with
// MORE THAN ONE simultaneous defect reports, in four ways relative to
// the pre-2.7 single-pass collection order. This is NOT a contractual
// promise: no AC and no golden fixes an ordering among unrelated,
// simultaneous defects, and none of the eight pre-existing goldens
// (single-outcome inputs) can observe it. It is recorded here, as this
// story's review asked, so a future restructure changes it
// KNOWINGLY rather than as an unnoticed by-product of where a phase
// boundary happens to fall — not so a caller can depend on it.
func renderDocument(t *Template, data, params bind.Value, fs FontSet) ([]byte, []Diagnostic, error) {
	// cache is shared between collection (coverage checks, AC4) and
	// embedding (subsetting) below, so a face is ever parsed at most
	// once per render regardless of how many chain members or runes
	// consult it. Only ever looked up by key, never ranged.
	cache := newFontCache()

	bands, bandsErr := documentBands(t)
	if bandsErr != nil {
		return nil, nil, bandsErr
	}
	geometry, gerr := pageGeometryOf(t)
	if gerr != nil {
		return nil, nil, gerr
	}
	imageRuns, ierr := collectImageRuns(t)
	if ierr != nil {
		return nil, nil, ierr
	}

	// Story 2.7, PHASE A: the content band ALONE, under D-2.7.3's fence
	// (contentBandResolver errors on {{page}}/{{pages}} rather than
	// passing them through). This is the ONLY input layout.Paginate
	// needs — internal/layout/band.go's ContentHeight takes page
	// geometry alone (finding 2, story creation) — so it is legal to
	// learn Y here, BEFORE the page-header/page-footer text exists,
	// without becoming pass two's job: this is still pass one, just
	// reordered within it.
	contentRuns, _, contentDiags, cterr := collectBandTextRuns(t, bands, contentBandIndex, data, params, fs, cache, contentBandResolver)
	if cterr != nil {
		return nil, nil, cterr
	}
	contentPlan, plerr := layout.Paginate(geometry, contentColumnItems(contentRuns, imageRuns))
	if plerr != nil {
		return nil, nil, fmt.Errorf("folio: Render: %w", plerr)
	}
	pageCount := len(contentPlan.Pages)

	// PHASE B: the two repeated bands, now that D-2.7.2's reservation
	// (digits(Y)) is computable. Collected in documentBands' authored
	// order — header, then footer — and combined below in that same
	// order relative to content (header, content, footer) so a
	// document using NO page slot produces the identical run sequence,
	// and therefore the identical CID allocation order
	// (buildShapedPDFRuns is order-sensitive by design, AC7/D-2.3-Q1),
	// this story's own change produced for every pre-existing document.
	headerRuns, headerPending, headerDiags, herr := collectBandTextRuns(t, bands, pageHeaderBandIndex, data, params, fs, cache, headerFooterResolver(pageCount))
	if herr != nil {
		return nil, nil, herr
	}
	footerRuns, footerPending, footerDiags, ferr := collectBandTextRuns(t, bands, pageFooterBandIndex, data, params, fs, cache, headerFooterResolver(pageCount))
	if ferr != nil {
		return nil, nil, ferr
	}

	// Story 2.8, D-2.8.6: Result.Diagnostics is DOCUMENT ORDER — band
	// order, then element declaration order within a band — never map
	// order (there is no map here) and never collection order (PHASE A
	// collects content BEFORE header/footer above, but band order is
	// header, content, footer, so this concatenation is NOT the order
	// the three collectBandTextRuns calls above happened to run in).
	// Each *Diags slice is already in element-declaration order within
	// its own band (collectBandTextRuns' own doc comment on `diags`).
	var diags []Diagnostic
	diags = append(diags, headerDiags...)
	diags = append(diags, contentDiags...)
	diags = append(diags, footerDiags...)

	headerOffset := 0
	contentOffset := len(headerRuns)
	footerOffset := contentOffset + len(contentRuns)

	runs := make([]textRunSource, 0, len(headerRuns)+len(contentRuns)+len(footerRuns))
	runs = append(runs, headerRuns...)
	runs = append(runs, contentRuns...)
	runs = append(runs, footerRuns...)

	var pending []pendingPageSlot
	for _, p := range headerPending {
		p.runIndex += headerOffset
		p.digitTableIndex += headerOffset
		pending = append(pending, p)
	}
	for _, p := range footerPending {
		p.runIndex += footerOffset
		p.digitTableIndex += footerOffset
		pending = append(pending, p)
	}

	// Story 2.3, AC1/AC8: each face-segment is shaped ONCE, with its own
	// buffer, by shapeSegments; positionSegments then derives the
	// segment cursor from those same shaped glyphs. So there is exactly
	// ONE shaping answer per segment and the drawn glyphs and the next
	// segment's origin are derived from it. Re-shaping here would
	// reintroduce the second derivation Blocker 1 was; these two slices
	// are views onto what shapeSegments already produced, built by
	// ranging a SLICE (D-1.3.5).
	shapedRuns := make([][]text.ShapedGlyph, len(runs))
	clusterTexts := make([][]string, len(runs))
	for i, r := range runs {
		shapedRuns[i] = r.glyphs
		clusterTexts[i] = r.clusterTexts
	}

	// Union of SHAPED GLYPH IDS per face, across the WHOLE document
	// (AC9, AC5) — built by ranging `runs` (a slice), never a map.
	//
	// This is the change AC5 is about: the subset input is the set of
	// glyphs the renderer actually draws, not the set of runes the
	// author typed. They are measurably different — shaping "office"
	// draws the `ffi` ligature, which no rune maps to — and building
	// the subset from the runes would leave the drawn glyph
	// unaddressable (D-1.5.8, one level up).
	glyphsByFace := map[string][]uint16{}
	for i, r := range runs {
		for _, g := range shapedRuns[i] {
			glyphsByFace[r.face] = append(glyphsByFace[r.face], g.GlyphID)
		}
	}

	faceNames := slices.Sorted(maps.Keys(glyphsByFace)) // ScanMapRange-compliant: sorted, deterministic object order.

	subsets := make(map[string]*fontset.Subset, len(faceNames))
	embedded := make(map[string]pdf.EmbeddedFace, len(faceNames))
	for _, name := range faceNames {
		font, ferr := cache.get(name, fs)
		if ferr != nil {
			return nil, nil, fmt.Errorf("folio: Render: face %q was resolved from a fallback chain but is missing from the FontSet: %w", name, ferr)
		}
		// ONE subsetting call per font per document (AC9), over the
		// union of shaped glyph ids collected above.
		sub, serr := font.Subset(glyphsByFace[name])
		if serr != nil {
			return nil, nil, fmt.Errorf("folio: Render: %w", serr)
		}
		subsets[name] = sub
		metrics := font.Metrics()
		created, modified := font.HeadTimes()
		embedded[name] = pdf.EmbeddedFace{
			Name: name,
			// The face's OWN identity, read off the supplied font
			// program's `name` table — not this map key (ISO 32000-1
			// Table 117; see internal/pdf's baseFont comment).
			PostScriptName: font.PostScriptName(),
			Program:        sub.Program,
			Tag:            sub.Tag,
			NumGlyphs:      sub.NumGlyphs,
			WidthForGlyph:  sub.WidthForGlyph,
			Ascent:         metrics.Ascent,
			Descent:        metrics.Descent,
			CapHeight:      metrics.CapHeight,
			BBoxXMin:       metrics.BBoxXMin,
			BBoxYMin:       metrics.BBoxYMin,
			BBoxXMax:       metrics.BBoxXMax,
			BBoxYMax:       metrics.BBoxYMax,
			HeadCreated:    created,
			HeadModified:   modified,
		}
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
			return nil, nil, fmt.Errorf("folio: Render: an image element references asset %q, which is not present in the document's assets map", key)
		}
		raw, derr := template.DecodeAssetBytes(asset)
		if derr != nil {
			return nil, nil, fmt.Errorf("folio: Render: asset %q: %w", key, derr)
		}
		img, derr := template.DecodeImageForRender(asset.MediaType, raw, key, firstElementIDByAssetKey[key])
		if derr != nil {
			return nil, nil, fmt.Errorf("folio: Render: %w", derr)
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

	pdfPlacements := make([]pagemodel.ImagePlacement, len(imageRuns))
	for i, r := range imageRuns {
		img := decodedByKey[r.assetKey]
		drawX, drawY, drawW, drawH := resolveImagePlacement(r, img)
		pdfPlacements[i] = pagemodel.ImagePlacement{
			AssetKey:   r.assetKey,
			X:          drawX,
			Y:          drawY,
			DrawWidth:  drawW,
			DrawHeight: drawH,
		}
	}

	pdfRuns, cerr := buildShapedPDFRuns(runs, shapedRuns, clusterTexts, subsets, embedded, cache, fs)
	if cerr != nil {
		return nil, nil, cerr
	}

	// Story 2.7, AC2's between-passes attachment point: buildShapedPDFRuns
	// has just allocated CIDs for every glyph in the document, INCLUDING
	// each page-slot's digit table, so this is the first point at which
	// buildPageNumberSlot can read the ten pre-shaped, pre-CID'd digits a
	// {{page}} occurrence will select among. Nothing here shapes
	// anything — it reads what pass one already measured.
	//
	// APPENDED, not assigned (this story's review, Blocker 1): `pending`
	// carries one entry per {{page}} OCCURRENCE, and more than one can
	// share a runIndex — a run's PageSlots is the ordered collection of
	// every reservation it carries, not its last one.
	for _, ps := range pending {
		slot, serr := buildPageNumberSlot(pdfRuns[ps.runIndex].Face, pdfRuns[ps.digitTableIndex], ps)
		if serr != nil {
			return nil, nil, serr
		}
		pdfRuns[ps.runIndex].PageSlots = append(pdfRuns[ps.runIndex].PageSlots, *slot)
	}

	// internal/layout produces the page model (AD-5); package folio hands
	// it to a renderer and does nothing else with it.
	//
	// Story 2.6: this used to be `[]pagemodel.Page{layout.ComposePage(...)}`
	// — a ONE-ELEMENT SLICE, which was the entire defect. Everything
	// upstream already produced page-absolute content and everything
	// downstream already handled N pages: pdf.SerializeTextDocument
	// reserves a page/content object pair per page and writes len(pages)
	// into /Count. Only the middle produced one. Content taller than the
	// content band was still DRAWN — below the bottom edge of the sheet,
	// with no error and no warning.
	pages, perr := paginateDocument(geometry, runs, imageRuns, pdfRuns, pdfPlacements)
	if perr != nil {
		return nil, nil, perr
	}

	b, serr := pdf.SerializeTextDocument(pages, embedded, pdfImages)
	if serr != nil {
		return nil, nil, serr
	}
	return b, diags, nil
}

// contentBandIndex is documentBands' authored index for the content band —
// the one band Story 2.6 paginates. The page header and page footer are
// repeated verbatim on every page instead.
const (
	pageHeaderBandIndex = 0
	contentBandIndex    = 1
	pageFooterBandIndex = 2
)

// paginateDocument turns one document's finished, page-absolute content into
// N pages, by asking internal/layout where the window boundaries fall.
//
// AD-4 IS THE POINT OF THIS FUNCTION'S SHAPE. Everything it receives is
// already laid out: pdfRuns and pdfPlacements carry final page-absolute
// coordinates, and this function only decides WHICH PAGE each belongs to and
// subtracts that page's window shift. It measures nothing, breaks nothing and
// re-positions nothing relative to anything else — pass one already did all
// of it. The page model that leaves here is finished, which is what lets
// internal/pdf lay nothing out (internal/passtwo_arch_test.go).
//
// THE PER-PAGE ORDER IS load-bearing and is NOT incidental: page header runs,
// then that page's content runs in AUTHORED order, then page footer runs —
// exactly the order documentBands walks the three bands, and therefore
// exactly the order the pre-2.6 code produced. That is what makes a document
// which fits on one page emit the SAME BYTES as before this story, which
// every one of the six existing goldens depends on.
func paginateDocument(
	geometry layout.PageGeometry,
	runs []textRunSource,
	imageRuns []imageRunSource,
	pdfRuns []pagemodel.TextRun,
	pdfPlacements []pagemodel.ImagePlacement,
) ([]pagemodel.Page, error) {
	// The two repeated bands, and the content column's atomic items.
	var header, footer layout.BandContent
	var items []layout.ColumnItem

	// Text. One line's runs are CONTIGUOUS in `runs` and share
	// (band, elementID, lineIndex) — positionSegments emits them together —
	// so the grouping below is a scan for a change of key, never a map
	// (D-1.3.5 / ScanMapRange: a map range would make the item order, and
	// therefore the emitted byte order, non-deterministic).
	for i := 0; i < len(runs); i++ {
		switch runs[i].band {
		case pageHeaderBandIndex:
			header.Runs = append(header.Runs, layout.TextRunRef(i))
			continue
		case pageFooterBandIndex:
			footer.Runs = append(footer.Runs, layout.TextRunRef(i))
			continue
		case digitTableBandIndex:
			// Story 2.7: exists only to carry a face's ten digit CIDs
			// through buildShapedPDFRuns (see digitTableRun) — never
			// drawn, never assigned to a page, never a content-band
			// item. Explicitly skipped, by name, rather than falling
			// into the "unrecognised band" internal error below: that
			// error guards documentBands' three-band ENUMERATION, and
			// this band is deliberately outside it.
			continue
		}
		// Content band: gather this whole line.
		//
		// This ASSUMES runs[i].band == contentBandIndex, since the switch
		// above already `continue`d past header and footer. That is true
		// today because documentBands (see documentBands, this file)
		// enumerates exactly three bands — but the assumption used to be
		// implicit: if it ever stopped holding (a fourth band), the loop
		// below matches ZERO runs at j (its condition fails at j==i), so
		// item.Runs stays empty, i = j-1 restores i, and the outer i++
		// puts i right back where it started — an infinite loop appending
		// an empty ColumnItem every iteration until OOM. Story 2.6
		// finisher, Finding 10: made an explicit internal error instead of
		// relying on an unasserted enumeration invariant in another
		// function.
		if runs[i].band != contentBandIndex {
			return nil, fmt.Errorf("folio: internal error: paginateDocument: run %d has band %d, which is neither the page-header, page-footer, nor content band — documentBands' three-band enumeration invariant no longer holds", i, runs[i].band)
		}
		j := i
		item := layout.ColumnItem{
			ElementID: runs[i].elementID,
			Top:       runs[i].itemTop,
			Bottom:    runs[i].itemBottom,
		}
		for j < len(runs) &&
			runs[j].band == contentBandIndex &&
			runs[j].elementID == runs[i].elementID &&
			runs[j].lineIndex == runs[i].lineIndex {
			item.Runs = append(item.Runs, layout.TextRunRef(j))
			j++
		}
		items = append(items, item)
		i = j - 1
	}

	// Images. Each is its own atomic item, and its extent is its DECLARED
	// BOX (r.y .. r.y+boxH), not the drawn box centred inside it: AD-24
	// already scaled the image to fit the box, so "does it fit on a page" is
	// a question about the BOX the template declared (D-2.6.1, rule 4).
	for i, r := range imageRuns {
		switch r.band {
		case pageHeaderBandIndex:
			header.Images = append(header.Images, layout.ImageRef(i))
		case pageFooterBandIndex:
			footer.Images = append(footer.Images, layout.ImageRef(i))
		default:
			items = append(items, layout.ColumnItem{
				ElementID: r.elementID,
				Top:       r.y,
				Bottom:    r.y + r.boxH,
				Images:    []layout.ImageRef{layout.ImageRef(i)},
			})
		}
	}

	plan, err := layout.Paginate(geometry, items)
	if err != nil {
		return nil, fmt.Errorf("folio: Render: %w", err)
	}

	pages := make([]pagemodel.Page, 0, len(plan.Pages))
	for pageIdx, assigned := range plan.Pages {
		// Story 2.7, AC2's between-passes step: pageNum is THIS page's
		// own number, 1-based. resolvePageRunForPage is a no-op for
		// every run but the ones carrying a PageSlots entry (empty for
		// every document that declares no {{page}} construct, which is
		// what keeps every pre-2.7 golden byte-identical: len(PageSlots)
		// == 0 returns run unchanged, verbatim, exactly as this loop
		// already copied it).
		pageNum := pageIdx + 1
		pageRuns := make([]pagemodel.TextRun, 0, len(header.Runs)+len(assigned.ContentRuns)+len(footer.Runs))
		for _, ref := range header.Runs {
			pageRuns = append(pageRuns, resolvePageRunForPage(pdfRuns[ref], pageNum))
		}
		for _, ref := range assigned.ContentRuns {
			// The window shift, and it is the ONLY transformation
			// pagination applies. Every content item on one page shares it,
			// so no item can be displaced relative to another — the column
			// itself is never mutated.
			run := pdfRuns[ref]
			run.Y -= assigned.Shift
			pageRuns = append(pageRuns, run)
		}
		for _, ref := range footer.Runs {
			pageRuns = append(pageRuns, resolvePageRunForPage(pdfRuns[ref], pageNum))
		}

		pageImages := make([]pagemodel.ImagePlacement, 0, len(header.Images)+len(assigned.ContentImages)+len(footer.Images))
		for _, ref := range header.Images {
			pageImages = append(pageImages, pdfPlacements[ref])
		}
		for _, ref := range assigned.ContentImages {
			img := pdfPlacements[ref]
			img.Y -= assigned.Shift
			pageImages = append(pageImages, img)
		}
		for _, ref := range footer.Images {
			pageImages = append(pageImages, pdfPlacements[ref])
		}

		pages = append(pages, layout.ComposePage(geometry, pageRuns, pageImages))
	}
	return pages, nil
}

// cidKey identifies one allocated CID: a subset glyph together with the
// text it extracts as. Two entries sharing a glyph but differing in text
// are two CIDs pointing at one glyph (Story 2.3, D-2.3-Q1 as ruled) —
// see pdf.EmbeddedFace.ExtraCIDs for the measured case that forces it.
type cidKey struct {
	glyph uint16
	text  string
}

// buildShapedPDFRuns turns the document's shaped runs into the CID-level
// runs internal/pdf emits, and — as the same pass, because the two
// cannot be computed independently — allocates each face's CID space and
// its /ToUnicode entries.
//
// Three things happen here and each is an acceptance criterion:
//
//   - AC4: every CID a content stream will emit originates in a shaped
//     glyph run. There is no other route: this is the only function that
//     produces a pagemodel.ShapedGlyph, and the only input it reads is the
//     shaper's output mapped through the subset plan. The old
//     rune -> GlyphForRune -> CID path does not exist any more, so the
//     property is structural rather than asserted by a denylist.
//   - AC6/AD-2: every position is scaled from FONT UNITS to the PDF's
//     1000-unit em exactly once, here, through geom.ScaleRound — this
//     module's one scaling function with its one documented rounding
//     mode. internal/pdf receives numbers already in the output's unit
//     and scales nothing.
//   - AC7/D-2.3-Q1: CIDs are allocated per (subset glyph, cluster text)
//     pair, in first-encounter order over runs in document order — a
//     deterministic order derived by ranging SLICES only (D-1.3.5).
//
// AD-23 holds trivially here: ot.GlyphPos is int16 throughout and
// geom.Length is int64, so nothing on this path is or becomes a float.
// Advances come from the shaper (which includes GPOS kerning), never
// from ot.Face.HorizontalAdvance (which returns float32 and omits it).
func buildShapedPDFRuns(
	runs []textRunSource,
	shaped [][]text.ShapedGlyph,
	clusterTexts [][]string,
	subsets map[string]*fontset.Subset,
	embedded map[string]pdf.EmbeddedFace,
	cache *fontCache,
	fs FontSet,
) ([]pagemodel.TextRun, error) {
	type faceCIDs struct {
		byKey       map[cidKey]uint16
		baseClaimed map[uint16]bool
		extras      []uint16
		entries     []pdf.CIDText
	}
	alloc := map[string]*faceCIDs{}

	pdfRuns := make([]pagemodel.TextRun, len(runs))
	for i, r := range runs {
		sub, ok := subsets[r.face]
		if !ok {
			return nil, fmt.Errorf("folio: Render: face %q has shaped runs but no subset", r.face)
		}
		font, ferr := cache.get(r.face, fs)
		if ferr != nil {
			return nil, fmt.Errorf("folio: Render: face %q: %w", r.face, ferr)
		}
		upem := int64(font.UnitsPerEm())

		state, seen := alloc[r.face]
		if !seen {
			state = &faceCIDs{byKey: map[cidKey]uint16{}, baseClaimed: map[uint16]bool{}}
			alloc[r.face] = state
		}

		glyphs := make([]pagemodel.ShapedGlyph, 0, len(shaped[i]))
		for gi, g := range shaped[i] {
			newGID, retained := sub.GlyphForSource[g.GlyphID]
			if !retained {
				// AC5: a shaped glyph the plan did not retain is a
				// located error naming the face and the glyph id, never
				// a silent .notdef.
				return nil, fmt.Errorf(
					"folio: Render: face %q: shaped glyph id %d was not retained by the subset plan",
					r.face, g.GlyphID,
				)
			}

			key := cidKey{glyph: newGID, text: clusterTexts[i][gi]}
			cid, allocated := state.byKey[key]
			if !allocated {
				switch {
				case !state.baseClaimed[newGID]:
					// The BASE block: CID == subset glyph id, exactly
					// as every folio PDF worked before this story.
					cid = newGID
					state.baseClaimed[newGID] = true
				default:
					// This glyph already carries a different text at its
					// base CID, so its second meaning needs a second CID
					// pointing at the same glyph.
					//
					// Identity-H's CID is TWO BYTES. Past 65535 the
					// conversion below wraps silently and the extra CID
					// collides with the base block, producing both a
					// wrong glyph and a wrong /ToUnicode entry — a silent
					// wrap where every other limit in this codebase is a
					// located error (Story 2.3 finisher, Finding 12).
					// Unreachable in practice (it needs a subset near the
					// 65535-glyph ceiling PLUS context-distinct CIDs on
					// top of it), which is exactly why it would never be
					// noticed if it did happen.
					next := sub.NumGlyphs + len(state.extras)
					if next > 0xFFFF {
						return nil, fmt.Errorf(
							"folio: Render: face %q: CID space exhausted — the subset has %d glyphs and this "+
								"document needs %d additional CIDs for glyphs carrying more than one source "+
								"text, which exceeds Identity-H's two-byte CID ceiling of 65535",
							r.face, sub.NumGlyphs, len(state.extras)+1,
						)
					}
					cid = uint16(next)
					state.extras = append(state.extras, newGID)
				}
				state.byKey[key] = cid
				state.entries = append(state.entries, pdf.CIDText{CID: cid, Text: key.text})
			}

			glyphs = append(glyphs, pagemodel.ShapedGlyph{
				CID:      cid,
				XAdvance: int64(geom.ScaleRound(geom.Length(int64(g.XAdvance)), 1000, upem)),
				XOffset:  int64(geom.ScaleRound(geom.Length(int64(g.XOffset)), 1000, upem)),
				YOffset:  int64(geom.ScaleRound(geom.Length(int64(g.YOffset)), 1000, upem)),
			})
		}

		pdfRuns[i] = pagemodel.TextRun{
			Face:           r.face,
			Glyphs:         glyphs,
			SourceText:     r.text,
			X:              r.x,
			Y:              r.y,
			FontSize:       r.fontSize,
			BaselineOffset: r.baselineOffset,
			ClipToBox:      r.clipToBox,
			ClipX:          r.clipX,
			ClipWidth:      r.clipWidth,
		}
	}

	// Write each face's allocated CID space back into its EmbeddedFace.
	// Ranges `runs` (a slice) to reach the face names, never `alloc`
	// (a map) — D-1.3.5, and the reason the entries end up in a
	// deterministic order at all.
	written := map[string]bool{}
	for _, r := range runs {
		if written[r.face] {
			continue
		}
		written[r.face] = true
		state := alloc[r.face]
		face := embedded[r.face]
		face.ExtraCIDs = state.extras
		entries := slices.Clone(state.entries)
		// Ascending CID order — the order buildToUnicodeCMap emits and
		// the order the pre-2.3 CMap already used, so a document that
		// needs no extra CIDs produces byte-identical /ToUnicode.
		slices.SortFunc(entries, func(a, b pdf.CIDText) int { return int(a.CID) - int(b.CID) })
		face.ToUnicode = entries
		embedded[r.face] = face
	}

	return pdfRuns, nil
}
