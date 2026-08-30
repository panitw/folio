package folio

import (
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
)

// A text element's style.align and style.valign were parsed, checked against
// their closed sets and serialized long before anything drew with them: both
// producers placed every line at the element's own top-left corner, so an
// authored "align": "right" was accepted, round-tripped, shown in the
// inspector — and then ignored on the page. These functions are the whole of
// the placement rule, and they are shared by the PDF producer (render.go) and
// the canvas paint projection (page_setup.go) so that the two cannot drift:
// the canvas exists to show what prints.
//
// SLACK ONLY. Alignment distributes the space an element's declared box has
// left over and nothing else: a line wider than its declared width, or a
// block taller than its declared height, keeps the element's own start edge,
// where FR44's existing clip-and-warn applies unchanged. One rule for both
// producers, and every canvas coordinate stays inside the non-negative,
// band-relative bound the paint plan is checked against.
//
// The arithmetic is integer millipoints throughout, halved by geom.ScaleRound
// — the same round-half-to-even rule every other derived span in the producer
// is rounded by — so all four targets produce identical bytes.

// elementAlignment reads one element's committed alignment. An unset style,
// an explicit null, and "left"/"top" all mean the same thing here: no offset.
func elementAlignment(el template.Element) (align string, valign string) {
	if !el.Style.Set || el.Style.Null {
		return "", ""
	}
	st := el.Style.Value
	if st.Align.Set && !st.Align.Null {
		align = st.Align.Value
	}
	if st.Valign.Set && !st.Valign.Null {
		valign = st.Valign.Value
	}
	return align, valign
}

// textAlignOffset is one packed line's horizontal offset inside the element's
// declared width. A centred line splits the slack with geom.ScaleRound, whose
// exact-half case rounds to even.
func textAlignOffset(align string, boxWidth, lineWidth geom.Length) geom.Length {
	slack := boxWidth - lineWidth
	if boxWidth <= 0 || slack <= 0 {
		return 0
	}
	switch align {
	case "right":
		return slack
	case "center":
		return geom.ScaleRound(slack, 1, 2)
	default: // "left", and an unset align: the element's own start edge
		return 0
	}
}

// textValignOffset is the whole packed block's vertical offset inside the
// element's declared height. It moves the block, never the spacing within it:
// the inter-baseline advance is a property of the font chain and the size.
func textValignOffset(valign string, boxHeight, blockHeight geom.Length) geom.Length {
	slack := boxHeight - blockHeight
	if boxHeight <= 0 || slack <= 0 {
		return 0
	}
	switch valign {
	case "bottom":
		return slack
	case "middle":
		return geom.ScaleRound(slack, 1, 2)
	default: // "top", and an unset valign: the element's own top edge
		return 0
	}
}

// textBlockHeight is the vertical extent of `lines` packed lines under one
// vertical model: the inter-baseline advance for every line after the first,
// plus the first line's ascent and the last line's descent — the same two
// spans positionSegments and the page-splitting item extents already use.
func textBlockHeight(lines int, vm verticalMetrics) geom.Length {
	if lines <= 0 {
		return 0
	}
	return geom.Length(int64(lines-1))*vm.Advance + vm.FirstBaseline + vm.LastDescent
}
