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

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/pagemodel"
	"github.com/panitw/folio/folio-go/internal/template"
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
	rect := pagemodel.Rect{X: x, Y: y, W: w, H: h}

	if hs.hasBackground {
		c, ok := parseHexColor(hs.background)
		if !ok {
			return pagemodel.Rect{}, newRenderError(DiagCodeStyleColorInvalid, elementID, "",
				fmt.Errorf("folio: Render: element %s: style.background/headerStyle.background %q is not a #RRGGBB colour", elementID, hs.background))
		}
		rect.HasFill = true
		rect.Fill = c
	}

	if hs.hasBorder {
		width := geom.Length(500) // folio-format.md's documented default: 0.5pt
		if hs.border.Width.Set && !hs.border.Width.Null {
			width = hs.border.Width.Value
		}
		colorHex := "#000000"
		if hs.border.Color.Set && !hs.border.Color.Null {
			colorHex = hs.border.Color.Value
		}
		c, ok := parseHexColor(colorHex)
		if !ok {
			return pagemodel.Rect{}, newRenderError(DiagCodeStyleColorInvalid, elementID, "",
				fmt.Errorf("folio: Render: element %s: style.border.color/headerStyle.border.color %q is not a #RRGGBB colour", elementID, colorHex))
		}
		edges := pagemodel.RectEdges{Top: true, Right: true, Bottom: true, Left: true}
		if hs.border.Edges.Set && !hs.border.Edges.Null {
			edges = pagemodel.RectEdges{}
			for _, e := range hs.border.Edges.Value {
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
			band:      bandIndex,
			elementID: string(el.ID),
			top:       tableTop,
			bottom:    tableBottom,
			rects:     rects,
		})
	}

	return runs, rectSources, diags, nil
}
