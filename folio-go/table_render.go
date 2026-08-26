// This file is Story 4.1's header-row rendering: a table's column
// geometry, its header cells' vector chrome (border/padding/
// background), and its column labels — the header LABEL is the ONLY
// text this story produces (AC9: zero data rows).
//
// SCOPE FENCE, stated once here rather than at every function: rows,
// cell text wrapping, pagination across pages, the repeated header on
// continuation pages, footer aggregates and alternating row shading are
// explicitly NOT this file's job (4.2-4.8, per the story's own AC9
// table). This file produces exactly one thing per table: its header
// row, once, wherever the table's own Y places it.
package folio

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/bind"
	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/pagemodel"
	"github.com/panitw/folio/folio-go/internal/template"
	"github.com/panitw/folio/folio-go/internal/text"
)

// tableRectSource is one table's header-row vector chrome (Story 4.1,
// R1/AC3/AC6): one pagemodel.Rect per column cell, ALWAYS populated
// when the table has at least one column — even a fully style-less
// table's cells are represented, carrying HasFill==false and
// HasStroke==false, so the header row still occupies its own page
// space (D-2.6.5's own rule: an item that occupies space must not be
// empty) without drawing anything (R6).
//
// It mirrors imageRunSource's shape (band + a page-absolute extent),
// deliberately: this is a SECOND, PARALLEL content kind alongside text
// runs and image placements, not a variant of either.
type tableRectSource struct {
	band        int
	elementID   string
	top, bottom geom.Length
	rects       []pagemodel.Rect

	// isDataRow / rowIndex — Story 4.2, DECISION-2 (owner ruling): a
	// carried identity for "which bound-collection row produced this
	// chrome group", so Story 4.3 ("a row moves whole to the next
	// page") can group this story's output by DIRECT FIELD LOOKUP
	// rather than by reconstructing membership from element ids,
	// extents or emission order — exactly the reconstruction the
	// ruling names as where a wrapped row silently becomes two.
	//
	// isDataRow is false (and rowIndex meaningless) for the header's
	// own chrome group (Story 4.1, unchanged): a header is not a row.
	// For a data row's chrome group (Story 4.2), isDataRow is true and
	// rowIndex is the row's 0-based position in the bound collection.
	//
	// This story asserts only that the identity itself is correct
	// (TestDataRowIdentityIsConsistentAndDistinct) — it asserts nothing
	// about pagination (AC7). Story 4.3 is this field's one named
	// consumer.
	isDataRow bool
	rowIndex  int

	// isHeaderRow — Story 4.3: the header's OWN grouping identity,
	// parallel to isDataRow/rowIndex above but for the one header row a
	// table has. AC5 extends "a row moves whole to the next page" to the
	// header without special-casing it: the header's chrome and its
	// column labels are one group, exactly as a data row's chrome and
	// its physical lines are one group. Set true ONLY on the header's own
	// tableRectSource (isDataRow stays false there, unchanged); never set
	// alongside isDataRow.
	isHeaderRow bool
}

// chromeRowGroup derives this rect source's layout.ItemGroup — Story 4.3's
// grouping identity — by DIRECT FIELD LOOKUP from isDataRow/isHeaderRow/
// rowIndex, never by reconstruction (D-4.2.2, R3). Not grouped
// (layout.ItemGroup{}) for a table with neither: unreachable in practice
// (every tableRectSource is either the header's own or a data row's,
// never neither), but stated rather than assumed.
func (r tableRectSource) chromeRowGroup() layout.ItemGroup {
	switch {
	case r.isHeaderRow:
		return layout.ItemGroup{Present: true, Key: layout.ItemGroupKey{ElementID: r.elementID, IsHeader: true}}
	case r.isDataRow:
		return layout.ItemGroup{Present: true, Key: layout.ItemGroupKey{ElementID: r.elementID, Index: r.rowIndex}}
	default:
		return layout.ItemGroup{}
	}
}

// resolvedHeaderStyle is a table's header-row style, cascaded ONCE per
// table (AC3/AC4, and this story's owner-ruled headerStyle addition):
// for every field, `headerStyle.<field>` wins when the table declares
// one AND that field is set within it; otherwise the table's own
// `style.<field>` wins when set; otherwise the field's documented
// default (folio-format.md's Style table, unchanged by this story
// except fontFamily — see the story's Delivery Log). `columns[].align`
// is resolved separately, per column, and still wins over both
// (AC4's own grounds, extended one level: the column's field is the
// most specific of the three).
type resolvedHeaderStyle struct {
	hasFontFamily bool
	fontFamily    string
	fontSize      geom.Length

	hasBorder bool
	border    template.Border

	hasBackground bool
	background    string

	padding template.Padding

	valign string // "top", "middle" or "bottom" — never empty

	// alignFallback is what a column without its OWN `align` uses
	// (AC4): headerStyle.align, else style.align, else "left".
	alignFallback string
}

// resolveHeaderStyle cascades el's table style and headerStyle exactly
// once (called once per table element).
func resolveHeaderStyle(el template.Element) resolvedHeaderStyle {
	var base, header template.Style
	if el.Style.Set && !el.Style.Null {
		base = el.Style.Value
	}
	hasHeader := false
	if el.Table.Set && el.Table.Value.HeaderStyle.Set && !el.Table.Value.HeaderStyle.Null {
		header = el.Table.Value.HeaderStyle.Value
		hasHeader = true
	}

	r := resolvedHeaderStyle{valign: "top", alignFallback: "left"}

	switch {
	case hasHeader && header.FontFamily.Set && !header.FontFamily.Null:
		r.hasFontFamily, r.fontFamily = true, header.FontFamily.Value
	case base.FontFamily.Set && !base.FontFamily.Null:
		r.hasFontFamily, r.fontFamily = true, base.FontFamily.Value
	}

	r.fontSize = defaultFontSizePt
	switch {
	case hasHeader && header.FontSize.Set && !header.FontSize.Null:
		r.fontSize = header.FontSize.Value
	case base.FontSize.Set && !base.FontSize.Null:
		r.fontSize = base.FontSize.Value
	}

	switch {
	case hasHeader && header.Border.Set && !header.Border.Null:
		r.hasBorder, r.border = true, header.Border.Value
	case base.Border.Set && !base.Border.Null:
		r.hasBorder, r.border = true, base.Border.Value
	}

	switch {
	case hasHeader && header.Background.Set && !header.Background.Null:
		r.hasBackground, r.background = true, header.Background.Value
	case base.Background.Set && !base.Background.Null:
		r.hasBackground, r.background = true, base.Background.Value
	}

	switch {
	case hasHeader && header.Padding.Set && !header.Padding.Null:
		r.padding = header.Padding.Value
	case base.Padding.Set && !base.Padding.Null:
		r.padding = base.Padding.Value
	}

	switch {
	case hasHeader && header.Valign.Set && !header.Valign.Null:
		r.valign = header.Valign.Value
	case base.Valign.Set && !base.Valign.Null:
		r.valign = base.Valign.Value
	}

	switch {
	case hasHeader && header.Align.Set && !header.Align.Null:
		r.alignFallback = header.Align.Value
	case base.Align.Set && !base.Align.Null:
		r.alignFallback = base.Align.Value
	}

	return r
}

// resolvedBodyStyle is a table's DATA-ROW style, cascaded ONCE per
// table (Story 4.2, AC5; D-000.76's ruling that headerStyle governs
// the header row ONLY): for every field, the table's own
// `style.<field>` wins when set, otherwise the field's documented
// default. There is deliberately NO headerStyle arm here — that is the
// whole point of AC5's fence. `columns[].align` is resolved separately
// per column and still wins over `alignFallback`, exactly as it does
// for the header (AC4, extended one level).
//
// fontFamily/fontSize are NOT carried here: a data row's font chain
// cascades from `style` alone, which is exactly fontChain's own
// cascade (render.go) — reused verbatim so a data cell's "no
// resolvable fontFamily" error is the SAME message the existing
// font-resolution failure produces, never a third spelling (AC5's own
// grounds, D-000.65).
type resolvedBodyStyle struct {
	hasBorder bool
	border    template.Border

	hasBackground bool
	background    string

	padding template.Padding

	valign string // "top", "middle" or "bottom" — never empty

	alignFallback string
}

// resolveBodyStyle cascades el's table style ALONE (never headerStyle)
// exactly once per table (Story 4.2, AC5).
func resolveBodyStyle(el template.Element) resolvedBodyStyle {
	var base template.Style
	if el.Style.Set && !el.Style.Null {
		base = el.Style.Value
	}

	r := resolvedBodyStyle{valign: "top", alignFallback: "left"}

	if base.Border.Set && !base.Border.Null {
		r.hasBorder, r.border = true, base.Border.Value
	}
	if base.Background.Set && !base.Background.Null {
		r.hasBackground, r.background = true, base.Background.Value
	}
	if base.Padding.Set && !base.Padding.Null {
		r.padding = base.Padding.Value
	}
	if base.Valign.Set && !base.Valign.Null {
		r.valign = base.Valign.Value
	}
	if base.Align.Set && !base.Align.Null {
		r.alignFallback = base.Align.Value
	}
	return r
}

// paddingEdges returns the four padding insets, each independently
// defaulting to zero when its own field is absent (AC3, R6).
func paddingEdges(p template.Padding) (top, right, bottom, left geom.Length) {
	if p.Top.Set && !p.Top.Null {
		top = p.Top.Value
	}
	if p.Right.Set && !p.Right.Null {
		right = p.Right.Value
	}
	if p.Bottom.Set && !p.Bottom.Null {
		bottom = p.Bottom.Value
	}
	if p.Left.Set && !p.Left.Null {
		left = p.Left.Value
	}
	return
}

// parseHexColor decodes a `#RRGGBB` string into page-model channels
// (AC5/D2). It is the ONLY hex parser in the module: no colour was ever
// consumed on any render path before this story (D1), so this is also
// the first place a malformed one can be discovered.
func parseHexColor(s string) (pagemodel.Color, bool) {
	if len(s) != 7 || s[0] != '#' {
		return pagemodel.Color{}, false
	}
	var out [3]uint8
	for i := 0; i < 3; i++ {
		hi, ok1 := hexDigit(s[1+i*2])
		lo, ok2 := hexDigit(s[2+i*2])
		if !ok1 || !ok2 {
			return pagemodel.Color{}, false
		}
		out[i] = hi<<4 | lo
	}
	return pagemodel.Color{R: out[0], G: out[1], B: out[2]}, true
}

func hexDigit(b byte) (uint8, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, true
	default:
		return 0, false
	}
}

// buildHeaderCellRect builds one column cell's vector primitive (AC3,
// R1) from the table's cascaded header style. hs.hasBackground/
// hasBorder absent means "draw nothing for this half" (R6) — never a
// hardcoded header treatment (the owner ruling this story carries: the
// AUTHOR expresses the look via style/headerStyle, this function
// invents none of it).
func buildHeaderCellRect(elementID string, x, y, w, h geom.Length, hs resolvedHeaderStyle) (pagemodel.Rect, error) {
	return buildCellRect(elementID, x, y, w, h, hs.hasBackground, hs.background, hs.hasBorder, hs.border)
}

// buildCellRect builds ONE cell's vector primitive (header or data row
// alike, Story 4.1/4.2) from a resolved has*/value pair for background
// and border — never a hardcoded treatment (the owner ruling both
// stories carry: the AUTHOR expresses the look via style, this
// function invents none of it). hasBackground/hasBorder absent means
// "draw nothing for this half" — a fully style-less cell still gets a
// Rect (HasFill==false, HasStroke==false), so it still occupies its
// own page space (D-2.6.5: an item that occupies space must not be
// empty) without drawing anything.
func buildCellRect(elementID string, x, y, w, h geom.Length, hasBackground bool, background string, hasBorder bool, border template.Border) (pagemodel.Rect, error) {
	rect := pagemodel.Rect{X: x, Y: y, W: w, H: h}

	if hasBackground {
		c, ok := parseHexColor(background)
		if !ok {
			return pagemodel.Rect{}, newRenderError(DiagCodeStyleColorInvalid, elementID, "",
				fmt.Errorf("folio: Render: element %s: style.background/headerStyle.background %q is not a #RRGGBB colour", elementID, background))
		}
		rect.HasFill = true
		rect.Fill = c
	}

	if hasBorder {
		width := geom.Length(500) // folio-format.md's documented default: 0.5pt
		if border.Width.Set && !border.Width.Null {
			width = border.Width.Value
		}
		colorHex := "#000000"
		if border.Color.Set && !border.Color.Null {
			colorHex = border.Color.Value
		}
		c, ok := parseHexColor(colorHex)
		if !ok {
			return pagemodel.Rect{}, newRenderError(DiagCodeStyleColorInvalid, elementID, "",
				fmt.Errorf("folio: Render: element %s: style.border.color/headerStyle.border.color %q is not a #RRGGBB colour", elementID, colorHex))
		}
		edges := pagemodel.RectEdges{Top: true, Right: true, Bottom: true, Left: true}
		if border.Edges.Set && !border.Edges.Null {
			edges = pagemodel.RectEdges{}
			for _, e := range border.Edges.Value {
				switch e {
				case "top":
					edges.Top = true
				case "right":
					edges.Right = true
				case "bottom":
					edges.Bottom = true
				case "left":
					edges.Left = true
				}
			}
		}
		rect.HasStroke = true
		rect.Stroke = c
		rect.StrokeWidth = width
		rect.Edges = edges
	}

	return rect, nil
}

// collectBandTableRuns walks one band's table elements and returns the
// header-row label runs (through the SAME shape/position pipeline
// every text element uses — D-16's forcing function: buildShapedPDFRuns
// stays the ONLY producer of pagemodel.ShapedGlyph) plus one
// tableRectSource per VISIBLE table with at least one column.
//
// A HIDDEN table (AC8, AD-24's Visibility clause) contributes nothing
// at all: no label run, no rect — checked before any geometry is even
// computed, exactly as collectBandTextRuns already does for a hidden
// text element.
func collectBandTableRuns(
	doc *Template,
	bands []bandWithOrigin,
	bandIndex int,
	data, params bind.Value,
	fc expr.FormatContext,
	fs FontSet,
	cache *fontCache,
	visible visibilityVerdicts,
) ([]textRunSource, []tableRectSource, []Diagnostic, error) {
	b := bands[bandIndex]
	var runs []textRunSource
	var rectSources []tableRectSource
	var diags []Diagnostic

	for _, el := range b.band.Elements {
		if el.Type != template.ElementTable {
			continue
		}
		if !isVisible(visible, el.ID) {
			// AD-24's Visibility clause: absent from the page model
			// entirely — no header row, no borders, no background —
			// and siblings do not move (nothing here ever adjusts
			// another element's position; band composition is a
			// TRANSLATION, never a negotiation, per internal/layout's
			// own doc comment).
			continue
		}
		if !el.Table.Set {
			continue // structurally unreachable (a table element always carries TableExt), defensive
		}
		tbl := el.Table.Value
		if len(tbl.Columns) == 0 {
			continue // nothing to lay out or draw
		}

		hs := resolveHeaderStyle(el)
		tableTop := layout.PlaceInBand(b.origin, el.Y)
		tableBottom := tableTop + tbl.HeaderHeight

		widths := make([]geom.Length, len(tbl.Columns))
		for i, c := range tbl.Columns {
			widths[i] = c.Width
		}
		geometry := layout.ColumnWidths(el.X, widths)

		rects := make([]pagemodel.Rect, len(tbl.Columns))
		padTop, padRight, padBottom, padLeft := paddingEdges(hs.padding)

		var chain []string
		if hs.hasFontFamily {
			var ok bool
			chain, ok = doc.doc.Fonts[hs.fontFamily]
			if !ok || len(chain) == 0 {
				// Mirrors fontChain's own error, verbatim in shape
				// (render.go) — a text element with the same defect
				// fails the same way, plain-wrapped, no *RenderError:
				// this story does not widen that existing behaviour.
				return nil, nil, nil, fmt.Errorf("folio: Render: element %s: style.fontFamily %q names a chain with no entries in the document's fonts map", el.ID, hs.fontFamily)
			}
		}

		for i, col := range tbl.Columns {
			cg := geometry.Columns[i]
			rect, rerr := buildHeaderCellRect(string(el.ID), cg.X, tableTop, cg.Width, tbl.HeaderHeight, hs)
			if rerr != nil {
				return nil, nil, nil, rerr
			}
			rects[i] = rect

			if col.Label == "" {
				continue
			}
			if !hs.hasFontFamily {
				// Same failure mode a text element with no
				// style.fontFamily already has (R6, amended, and
				// fontChain's own error text): a non-empty label
				// needs a resolvable font, and no default exists
				// (this story's Delivery Log) — plain-wrapped, exactly
				// as fontChain's caller wraps it for a text element.
				return nil, nil, nil, fmt.Errorf("folio: Render: element %s: has a column label but no style.fontFamily (nor headerStyle.fontFamily) to resolve a font from", el.ID)
			}

			align := hs.alignFallback
			if col.Align.Set && !col.Align.Null {
				align = col.Align.Value
			}

			contentX := cg.X + padLeft
			contentW := cg.Width - padLeft - padRight
			contentY := tableTop + padTop
			contentH := tbl.HeaderHeight - padTop - padBottom

			segs, glyphDiags, serr := shapeSegments(string(col.ID), chain, col.Label, fs, cache)
			if serr != nil {
				return nil, nil, nil, serr
			}
			diags = append(diags, glyphDiags...)
			totalRunes := len([]rune(col.Label))

			vm, verr := chainVerticalModel(chain, hs.fontSize, fs, cache)
			if verr != nil {
				return nil, nil, nil, verr
			}

			measured := measureRuneRange(segs, 0, totalRunes, hs.fontSize)

			var textX geom.Length
			switch align {
			case "right":
				textX = contentX + contentW - measured
			case "center":
				textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
			default: // "left" and any value the load-time closed-set check already rejected otherwise
				textX = contentX
			}

			textBlockHeight := vm.FirstBaseline + vm.LastDescent
			var lineTopY geom.Length
			switch hs.valign {
			case "bottom":
				lineTopY = contentY + contentH - textBlockHeight
			case "middle":
				lineTopY = contentY + geom.ScaleRound(contentH-textBlockHeight, 1, 2)
			default: // "top"
				lineTopY = contentY
			}

			placed, perr := positionSegments(segs, 0, totalRunes, textX, lineTopY, hs.fontSize, vm.FirstBaseline, nil)
			if perr != nil {
				return nil, nil, nil, perr
			}

			overflows := measured > contentW
			for j := range placed {
				placed[j].band = bandIndex
				placed[j].elementID = string(el.ID)
				placed[j].lineIndex = 0
				placed[j].itemTop = tableTop
				placed[j].itemBottom = tableBottom
				placed[j].isHeaderLabel = true
				if overflows {
					// AC2: the wide-label render's header text is
					// clipped, per its declared (padded) box — it
					// NEVER widens the column. Same mechanism Story
					// 2.8 already gives text elements (D-2.8.1),
					// reused rather than re-invented.
					placed[j].clipToBox = true
					placed[j].clipX = contentX
					placed[j].clipWidth = contentW
				}
			}
			runs = append(runs, placed...)
		}

		rectSources = append(rectSources, tableRectSource{
			band:        bandIndex,
			elementID:   string(el.ID),
			top:         tableTop,
			bottom:      tableBottom,
			rects:       rects,
			isHeaderRow: true,
		})

		// --- Story 4.2: data rows ---
		//
		// checkTableBindings (render.go) already ran, before this
		// function is ever called, and already proved tbl.Bind resolves
		// to a KindArray value (or failed the whole render) — so
		// data.Lookup below is safe to read .Arr from unconditionally.
		alias := resolvedRowAlias(tbl.As)
		collectionVal, _ := data.Lookup(tableCollectionSegments(tbl.Bind))
		items := collectionVal.Arr

		if len(items) > 0 {
			// AC5: a data cell's font/border/background/padding/align/
			// valign cascade from the table's own `style` ONLY — never
			// `headerStyle` (D-000.76: header-only). fontChain is
			// reused VERBATIM (not re-implemented) so a data cell's
			// "no resolvable fontFamily" failure is the SAME message
			// the existing font-resolution failure produces, never a
			// third spelling (AC5's own grounds, D-000.65).
			bodyChain, cerr := fontChain(doc, el)
			if cerr != nil {
				return nil, nil, nil, fmt.Errorf("folio: Render: element %s: %w", el.ID, cerr)
			}
			bodyFontSize := defaultFontSizePt
			if el.Style.Set && !el.Style.Null && el.Style.Value.FontSize.Set && !el.Style.Value.FontSize.Null {
				bodyFontSize = el.Style.Value.FontSize.Value
			}
			bs := resolveBodyStyle(el)
			padTopB, padRightB, padBottomB, padLeftB := paddingEdges(bs.padding)

			// R2: ONE vertical model for the WHOLE table's body, computed
			// once outside the row loop — never recomputed per cell or
			// per row.
			vm, verr := chainVerticalModel(bodyChain, bodyFontSize, fs, cache)
			if verr != nil {
				return nil, nil, nil, fmt.Errorf("folio: Render: element %s: %w", el.ID, verr)
			}

			type cellResult struct {
				lines     []wrappedLine
				segs      []faceSegment
				align     string
				clip      bool
				clipX     geom.Length
				clipWidth geom.Length
			}

			rowTop := tableBottom
			// nextLineIndex starts at 1: the header's own label line
			// above always uses lineIndex 0 (unchanged), so this counter
			// — monotonically increasing across EVERY physical line of
			// EVERY row of this table — never collides with it and never
			// collides between two different rows, which is exactly what
			// keeps contentColumnItems/paginateDocument's
			// (elementID, lineIndex) grouping from merging two distinct
			// physical lines into one ColumnItem (D2).
			nextLineIndex := 1

			for rowIdx, rowVal := range items {
				scope := bind.NewScope(data, params).WithRow(rowVal, alias)

				cellResults := make([]cellResult, len(tbl.Columns))
				maxLines := 0
				for ci, col := range tbl.Columns {
					cg := geometry.Columns[ci]
					contentW := cg.Width - padLeftB - padRightB

					align := bs.alignFallback
					if col.Align.Set && !col.Align.Null {
						align = col.Align.Value
					}

					// AC4: the column's bind resolves in the table's
					// ROW SCOPE — bind.Resolve, never bind.BindTextSpans
					// — so an unqualified path still resolves from the
					// document root (a row never shadows it) and
					// `params.` still resolves to the parameters,
					// shadowed by nothing (AD-11). AC4's own three AD-14
					// cases (absent -> Error, explicit null -> empty and
					// not an error, wrong kind -> Error never coerced)
					// are ALL already implemented inside bind.Resolve
					// itself — nothing here re-implements any of them.
					//
					// Diagnostic located by COLUMN id (AC4's own
					// grounds: "columns[].id exists precisely so a
					// diagnostic can name a column") — the SAME
					// DiagCodeBindingPathAbsent code collectBandTextRuns
					// already uses for a text element's own unresolvable
					// binding (D-000.65: reuse, mint nothing).
					boundText, subs, caveats, berr := bind.Resolve(col.Bind, scope, fc, string(col.ID))
					if berr != nil {
						return nil, nil, nil, newRenderError(DiagCodeBindingPathAbsent, string(col.ID), "", fmt.Errorf("folio: Render: %w", berr))
					}
					for _, c := range caveats {
						diags = append(diags, diagnosticFromCaveat(string(col.ID), c))
					}
					if boundText == "" {
						cellResults[ci] = cellResult{align: align}
						continue
					}

					segs, glyphDiags, serr := shapeSegments(string(col.ID), bodyChain, boundText, fs, cache)
					if serr != nil {
						return nil, nil, nil, serr
					}
					diags = append(diags, glyphDiags...)
					totalRunes := len([]rune(boundText))

					atomic := atomicSpansFor(doc.doc.UnbreakableValues, subs)
					ops := text.Opportunities(text.Dictionary(), boundText, atomic)
					// AC2: wrap INSIDE the column's own content width —
					// packLines is the SAME packer text elements use.
					lines := packLines(segs, ops, totalRunes, bodyFontSize, contentW)

					// AC3: residual overflow (no break opportunity
					// narrow enough) is CLIPPED at the column's content
					// box, with the EXISTING DiagCodeTextClippedWidth —
					// D-2.8.1's own precedent, D-000.65: no new code.
					overflow, overflows := detectWidthOverflow(string(col.ID), lines, contentW)
					if overflows {
						diags = append(diags, Diagnostic{
							Severity:  SeverityWarning,
							Code:      DiagCodeTextClippedWidth,
							ElementID: overflow.elementID,
							Message:   widthClipMessage("column", "content", overflow),
						})
					}

					if len(lines) > maxLines {
						maxLines = len(lines)
					}
					cellResults[ci] = cellResult{
						lines: lines, segs: segs, align: align,
						clip: overflows, clipX: cg.X + padLeftB, clipWidth: contentW,
					}
				}

				// R3: a block of n lines is
				// FirstBaseline + (n-1)*Advance + LastDescent. n is at
				// least 1 even when every cell in the row is empty — a
				// data row, unlike an empty text element, is not itself
				// optional (it is one element of the bound collection),
				// so it still occupies one blank line's worth of height.
				linesInRow := maxLines
				if linesInRow < 1 {
					linesInRow = 1
				}
				rowHeight := padTopB + vm.FirstBaseline + geom.Length(int64(linesInRow-1))*vm.Advance + vm.LastDescent + padBottomB
				rowBottom := rowTop + rowHeight

				// AC5 (Story 4.2 review Finding 4): bs.valign
				// distributes a CELL's own vertical slack —
				// linesInRow minus that cell's own line count — WITHIN
				// the row, the body's analogue of the header's own
				// valign (R2/R3's shared vertical model: same three-
				// way switch, applied to a whole-line COUNT here
				// rather than a sub-line pixel remainder, since a
				// row's height is already an exact multiple of
				// vm.Advance). It never changes the row's own height
				// (still exactly linesInRow, computed above) or any
				// column's geometry (AD-13) — it only decides which of
				// the row's physical line SLOTS a shorter cell's own
				// lines occupy: "top" (the default, and the ONLY
				// behaviour before this fix) leaves them at the row's
				// first slots; "bottom" shifts them to the last
				// slots; "middle" splits the remainder, rounding down
				// (an integer LINE count, not a Length — no
				// geom.ScaleRound/binary-float concern, AD-1 holds).
				lineOffsets := make([]int, len(tbl.Columns))
				for ci := range tbl.Columns {
					slack := linesInRow - len(cellResults[ci].lines)
					if slack < 0 {
						slack = 0
					}
					switch bs.valign {
					case "bottom":
						lineOffsets[ci] = slack
					case "middle":
						lineOffsets[ci] = slack / 2
					default: // "top"
						lineOffsets[ci] = 0
					}
				}

				// DECISION-1 (ruled): a data row gets cell chrome too,
				// cascaded from `style` alone (never headerStyle,
				// never a data-driven decision — bs was resolved ONCE,
				// above, from the template alone). One tableRectSource
				// per ROW — reusing the EXACT same downstream handling
				// contentColumnItems/paginateDocument already give the
				// header's own tableRectSource, so a data row's chrome
				// becomes its own ColumnItem, spanning the row's full
				// extent (AC7), with no changes needed to either of
				// those two functions.
				cellRects := make([]pagemodel.Rect, len(tbl.Columns))
				for ci := range tbl.Columns {
					cg := geometry.Columns[ci]
					rect, rerr := buildCellRect(string(el.ID), cg.X, rowTop, cg.Width, rowHeight, bs.hasBackground, bs.background, bs.hasBorder, bs.border)
					if rerr != nil {
						return nil, nil, nil, rerr
					}
					cellRects[ci] = rect
				}
				rectSources = append(rectSources, tableRectSource{
					band:      bandIndex,
					elementID: string(el.ID),
					top:       rowTop,
					bottom:    rowBottom,
					rects:     cellRects,
					isDataRow: true,
					rowIndex:  rowIdx,
				})

				// One physical line at a time: ALL columns' cell content
				// at line li share the SAME lineIndex (assigned once per
				// li, below) and the SAME vertical extent, exactly as
				// the header's several column labels already share
				// lineIndex 0 — this is what keeps a multi-column line
				// grouped into ONE ColumnItem (D2's "two items, same
				// extent" shape, extended to N columns on one physical
				// line rather than only two content kinds).
				for li := 0; li < linesInRow; li++ {
					lineTopY := rowTop + padTopB + geom.Length(int64(li))*vm.Advance
					lineBottom := lineTopY + vm.FirstBaseline + vm.LastDescent
					for ci := range tbl.Columns {
						cr := cellResults[ci]
						cellLi := li - lineOffsets[ci]
						if cellLi < 0 || cellLi >= len(cr.lines) {
							continue
						}
						ln := cr.lines[cellLi]
						cg := geometry.Columns[ci]
						contentX := cg.X + padLeftB
						contentW := cg.Width - padLeftB - padRightB
						measured := ln.width

						var textX geom.Length
						switch cr.align {
						case "right":
							textX = contentX + contentW - measured
						case "center":
							textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
						default: // "left"
							textX = contentX
						}

						placed, poserr := positionSegments(cr.segs, ln.from, ln.to, textX, lineTopY, bodyFontSize, vm.FirstBaseline, nil)
						if poserr != nil {
							return nil, nil, nil, poserr
						}
						for j := range placed {
							placed[j].band = bandIndex
							placed[j].elementID = string(el.ID)
							placed[j].lineIndex = nextLineIndex
							placed[j].itemTop = lineTopY
							placed[j].itemBottom = lineBottom
							placed[j].isTableRowLine = true
							placed[j].rowIndex = rowIdx
							if cr.clip {
								placed[j].clipToBox = true
								placed[j].clipX = cr.clipX
								placed[j].clipWidth = cr.clipWidth
							}
						}
						runs = append(runs, placed...)
					}
					nextLineIndex++
				}

				rowTop = rowBottom
			}
		}
	}

	return runs, rectSources, diags, nil
}
