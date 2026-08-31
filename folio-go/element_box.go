// This file is Story 9.1's whole engine obligation: an element's
// `style.background` and `style.border` finally reach the page model.
//
// Before this story they reached NOTHING outside a table. The format
// carried them, parse_bands.go decoded them, folio_expr_validate.go
// checked them for placeholders, serialize.go wrote them back — and the
// only consumer on any render path was table_render.go, which resolves
// the TABLE element's own style into per-cell chrome (resolveHeaderStyle
// / resolveBodyStyle, `base = el.Style.Value`). A text, image, rect or
// line element's background and border were inert, and rect and line
// reached the page model not at all (render_visibility.go's "LATENT, not
// broken": their verdicts were computed and unconsumed).
//
// The shape of the fix is deliberately NOT new machinery. A box is a
// rect group with a band, an element id and a vertical extent — which is
// exactly tableRectSource, the carrier Story 4.1 already built and that
// paginateDocument, contentColumnItems and the page assembler already
// know how to place, repeat, shift and clip. So this file produces
// tableRectSource values through buildCellRectWithBackgroundField, the
// SAME builder that draws a table cell's chrome, and adds no second
// implementation of "what does style.border mean" (D-000.38's rule,
// applied to geometry rather than to placeholders).
//
// What keeps AD-21's corpus a witness rather than a casualty: an element
// with neither background nor border produces no source at all, so a
// document that declares none — every golden in the corpus — reaches
// paginateDocument with a byte-identical rect slice.
package folio

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/pagemodel"
	"github.com/panitw/folio/folio-go/internal/template"
)

// collectElementBoxRects walks every band's elements in document order
// and returns one tableRectSource per element that declares a box.
//
// Order is bands in documentBands' authored order (header, content,
// footer), then elements in declaration order within a band — the same
// order collectBandTableRuns walks — because that order is the emitted
// byte order.
//
// Four kinds are eligible: text, image, rect and line. A TABLE is
// excluded by name (not by omission): its style.background and
// style.border are already consumed, as the cell chrome Epic 4 draws, so
// painting an element box for it too would draw the same declaration
// twice.
func collectElementBoxRects(bands []bandWithOrigin, visible visibilityVerdicts) ([]tableRectSource, error) {
	var sources []tableRectSource
	for bandIndex, b := range bands {
		for _, el := range b.band.Elements {
			if el.Type == template.ElementTable {
				continue
			}
			if !isVisible(visible, el.ID) {
				// AD-24's Visibility clause, the same reading
				// collectBandTableRuns applies to a hidden table: absent
				// from the page model entirely, and no sibling moves.
				continue
			}
			style, hasBackground, hasBorder := elementBoxDeclaration(el)
			if !hasBackground && !hasBorder {
				// The corpus's own path: no declaration, no source, no
				// change to the emitted bytes.
				continue
			}
			w, h, sized := declaredBox(el)
			if !sized {
				// A box needs a rectangle with area. parse_bands.go makes
				// width and height REQUIRED on every non-table element, so
				// the absent case here is defensive; the reachable one is a
				// zero or negative rectangle, which the loader accepts and
				// which has no box to draw. Nothing is drawn and the render
				// is otherwise unaffected — the null-asset image's
				// precedent, where the element is present and the paint is
				// absent.
				continue
			}
			rect, err := buildCellRectWithBackgroundField(
				string(el.ID),
				el.X, layout.PlaceInBand(b.origin, el.Y), w, h,
				hasBackground, style.Background.Value, "style.background",
				hasBorder, style.Border.Value,
			)
			if err != nil {
				return nil, err
			}
			sources = append(sources, tableRectSource{
				band:      bandIndex,
				elementID: string(el.ID),
				top:       layout.PlaceInBand(b.origin, el.Y),
				bottom:    layout.PlaceInBand(b.origin, el.Y) + h,
				rects:     []pagemodel.Rect{rect},
				// isDataRow/isHeaderRow/isFooterRow stay false: an element
				// box is not a table row, so chromeRowGroup returns the
				// zero ItemGroup and this source paginates as its own
				// ungrouped content item — and rectIsDataRow[ref] is false
				// downstream, so no row displacement is ever applied to it.
			})
		}
	}
	return sources, nil
}

// elementBoxDeclaration is THE ONE READING of "this element declares a
// box", and the only place in the module that spells it. A box exists
// where a PRESENT, non-null `style.background` or `style.border` does —
// nothing else, and in particular nothing about the element's kind, its
// size or whether the canvas can see it.
//
// It returns the style beside the two flags because the builder needs
// all three, and splitting them would put the same two Presence tests in
// two places. Callers that only need the verdict use elementDeclaresBox.
func elementBoxDeclaration(el template.Element) (style template.Style, hasBackground, hasBorder bool) {
	s, ok := elementStyle(el)
	if !ok {
		return template.Style{}, false, false
	}
	return s, s.Background.Set && !s.Background.Null, s.Border.Set && !s.Border.Null
}

// elementDeclaresBox is that rule as a predicate, for the SECOND caller
// Story 7.9 gives it: page_setup.go's window count, which must contribute
// a column item exactly where this file contributes a rect source and
// nowhere else.
//
// It exists so that "styled means background or border" is written down
// once. The canvas used to place every non-text component whatever its
// style, so an unstyled rect occupied a window the printed document does
// not have — invisible while the two answers were both ungrouped, and a
// confidently wrong count the moment a group made the canvas's partition
// matter. A second copy of the rule in page_setup.go would have been the
// same defect waiting on the next divergence.
func elementDeclaresBox(el template.Element) bool {
	_, hasBackground, hasBorder := elementBoxDeclaration(el)
	return hasBackground || hasBorder
}

// elementStyle returns el's style block, and whether it has one at all.
// A present-null style (`"style": null`) is no style, exactly as every
// other reader of a Presence treats it.
func elementStyle(el template.Element) (template.Style, bool) {
	if !el.Style.Set || el.Style.Null {
		return template.Style{}, false
	}
	return el.Style.Value, true
}

// declaredBox returns el's declared width and height, and whether BOTH
// are present, non-null and positive. A box is drawn only from a
// rectangle the author actually declared; nothing here infers one from
// measured content, which would make the box a function of the data —
// and a text element's declared height is read by nothing else in the
// renderer (D-2.8.1: the declared WIDTH is FR44's only clip bound), so
// the box neither gains nor loses a clip bound by being drawn.
func declaredBox(el template.Element) (w, h geom.Length, ok bool) {
	if !el.Width.Set || el.Width.Null || !el.Height.Set || el.Height.Null {
		return 0, 0, false
	}
	if el.Width.Value <= 0 || el.Height.Value <= 0 {
		return 0, 0, false
	}
	return el.Width.Value, el.Height.Value, true
}

// elementInk resolves a style block's `color` into page-model channels:
// Story 10.1's text ink, the one style colour that paints glyphs rather
// than a rectangle. It is validated at RENDER, through the module's one
// hex parser, exactly as style.background and style.border.color already
// are (buildCellRectWithBackgroundField) — the format checks a colour
// string for a placeholder at load and for nothing else, by design.
//
// A style that is absent, null, or carries no `color` returns hasInk
// false, and every producer then emits no colour operator at all: the
// state every document rendered in before this field existed.
func elementInk(style template.Presence[template.Style], elementID, fieldPath string) (pagemodel.Color, bool, error) {
	if !style.Set || style.Null {
		return pagemodel.Color{}, false, nil
	}
	return styleInk(style.Value, elementID, fieldPath)
}

// styleInk is elementInk over an already-resolved style block — the form
// a table's cascade (resolveHeaderStyle / resolveBodyStyle) hands over,
// where the Presence wrapper has already been unwrapped.
func styleInk(st template.Style, elementID, fieldPath string) (pagemodel.Color, bool, error) {
	if !st.Color.Set || st.Color.Null {
		return pagemodel.Color{}, false, nil
	}
	c, ok := parseHexColor(st.Color.Value)
	if !ok {
		return pagemodel.Color{}, false, newRenderError(DiagCodeStyleColorInvalid, elementID, "",
			fmt.Errorf("folio: Render: element %s: %s %q is not a #RRGGBB colour", elementID, fieldPath, st.Color.Value))
	}
	return c, true, nil
}
