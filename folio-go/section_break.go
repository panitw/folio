package folio

import (
	"fmt"
	"strings"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/template"
)

// This file is the engine half of spec-section-break (CAP-2 to CAP-5): the
// content band's optional `sectionBreak`, the load checks it needs, the
// keepTogether split it forces, and the section-aware pagination both of
// buildPageModel's passes run.
//
// THE RULE. Elements declared at or below the break form the section. The
// items above it are paginated on their own first. If they end at or above
// the line on their last page, the section is drawn on that page; otherwise
// on a page added after it. Either way every section element sits at its
// declared y, so the section's first page carries its own shift (zero), and
// a section that does not fit continues under the ordinary rules.
//
// THE BYTE-IDENTITY FAST PATH. When the above-line items fit one page and end
// at or above the line, the document is paginated exactly as it is without
// the key — one layout pass over every item — so its PDF is unchanged (CAP-3).

// sectionBreakDataPath locates a band-level section-break diagnostic.
const sectionBreakDataPath = "bands.content.sectionBreak"

// sectionBreakSplit is one document's section break, resolved for pagination.
// The zero value is "no break", and paginates exactly as before.
type sectionBreakSplit struct {
	// line is the break in page space: the content origin plus the offset.
	line geom.Length
	// members holds the ids of the content-band elements declared at or below
	// the break. Looked up, never ranged.
	members map[string]bool
}

// sectionBreakOf resolves t's section break against g. A document without a
// break, or whose section is empty, gets the zero value.
func sectionBreakOf(t *Template, g layout.PageGeometry) sectionBreakSplit {
	offset, ok := declaredSectionBreak(t)
	if !ok {
		return sectionBreakSplit{}
	}
	out := sectionBreakSplit{line: layout.Origins(g).Content + offset}
	for _, el := range t.doc.Bands.Content.Elements {
		if el.Y < offset {
			continue
		}
		if out.members == nil {
			out.members = map[string]bool{}
		}
		out.members[string(el.ID)] = true
	}
	return out
}

// declaredSectionBreak returns the content band's break offset, if it has one.
func declaredSectionBreak(t *Template) (geom.Length, bool) {
	if t == nil || t.doc == nil {
		return 0, false
	}
	sb := t.doc.Bands.Content.SectionBreak
	if !sb.Set || sb.Null {
		return 0, false
	}
	return sb.Value, true
}

// validateSectionBreak refuses, at load, a break the content band cannot
// honour (SECTION_BREAK_INVALID, located at the band) and an element whose
// declared box lies on both sides of it (SECTION_BREAK_STRADDLED, located at
// the element). It lives here, not in internal/template, because the range
// check needs layout.ContentHeight — validateTableMinHeights' reason.
func validateSectionBreak(t *Template) error {
	offset, ok := declaredSectionBreak(t)
	if !ok {
		return nil
	}
	if offset <= 0 {
		return newRenderError(DiagCodeSectionBreakInvalid, "", sectionBreakDataPath,
			fmt.Errorf("folio: bands.content.sectionBreak: %spt is at or above the content band's top — a section break must lie inside the content band, greater than 0", template.FormatPoints(offset)))
	}
	if g, err := pageGeometryOf(t); err == nil {
		if height := layout.ContentHeight(g); offset >= height {
			return newRenderError(DiagCodeSectionBreakInvalid, "", sectionBreakDataPath,
				fmt.Errorf("folio: bands.content.sectionBreak: %spt is at or below the bottom of the content band (a content height of %spt) — move the break up, or give the content band more room", template.FormatPoints(offset), template.FormatPoints(height)))
		}
	}
	for _, el := range t.doc.Bands.Content.Elements {
		top, bottom := sectionBreakDeclaredBox(el)
		if top < offset && bottom > offset {
			return newRenderError(DiagCodeSectionBreakStraddled, string(el.ID), "",
				fmt.Errorf("folio: element %s: its declared box runs from %spt to %spt, across the section break at %spt — every element must lie wholly above or wholly below the break; move or resize the element, or move the break", el.ID, template.FormatPoints(top), template.FormatPoints(bottom), template.FormatPoints(offset)))
		}
	}
	return nil
}

// sectionBreakDeclaredBox is the vertical extent of el's declared box, the
// extent the straddle rule reads. A table's box is its header row — its rows
// and its minHeight floor may run past the line. An element with no usable
// declared height is a point at its y.
func sectionBreakDeclaredBox(el template.Element) (top, bottom geom.Length) {
	top, bottom = el.Y, el.Y
	if el.Type == template.ElementTable {
		if el.Table.Set && !el.Table.Null {
			bottom = el.Y + el.Table.Value.HeaderHeight
		}
		return top, bottom
	}
	if el.Height.Set && !el.Height.Null && el.Height.Value > 0 {
		bottom = el.Y + el.Height.Value
	}
	return top, bottom
}

// sectionBreakSplitTags returns, in authored order of first appearance, the
// keepTogether tags whose members lie on both sides of the break, with the
// first member of each.
func sectionBreakSplitTags(t *Template) (tags []string, firstMember map[string]string) {
	offset, ok := declaredSectionBreak(t)
	if !ok {
		return nil, nil
	}
	above := map[string]bool{}
	below := map[string]bool{}
	firstMember = map[string]string{}
	var order []string
	for _, el := range t.doc.Bands.Content.Elements {
		if !el.KeepTogether.Set || el.KeepTogether.Null || el.KeepTogether.Value == "" {
			continue
		}
		tag := el.KeepTogether.Value
		if _, seen := firstMember[tag]; !seen {
			firstMember[tag] = string(el.ID)
			order = append(order, tag)
		}
		if el.Y >= offset {
			below[tag] = true
		} else {
			above[tag] = true
		}
	}
	for _, tag := range order {
		if above[tag] && below[tag] {
			tags = append(tags, tag)
		}
	}
	return tags, firstMember
}

// sectionBreakSplitDiagnostics is one SECTION_BREAK_SPLITS_KEEP_TOGETHER
// Warning per split group, located at the group's first member. Emitted on
// every render, whether or not the section moves, because the grouping the
// author declared is not the grouping that is honoured.
func sectionBreakSplitDiagnostics(t *Template) []Diagnostic {
	tags, first := sectionBreakSplitTags(t)
	var out []Diagnostic
	for _, tag := range tags {
		out = append(out, Diagnostic{
			Severity:  SeverityWarning,
			Code:      DiagCodeSectionBreakSplitsKeepTogether,
			ElementID: first[tag],
			Message: fmt.Sprintf("folio: Render: element %s: keepTogether group %q has members on both sides of the section break, so it is split at the line — the members above the break are kept together where they are, and the members below it are kept together with the section. Move the group's members to one side of the break, or move the break, to keep the whole group together",
				first[tag], tag),
		})
	}
	return out
}

// sectionPlan is a pagination with the section break applied. Pages above
// the break shift by Pages[p].Shift; section elements shift by
// sectionShift[p]. With no break section is nil and every element shifts by
// Pages[p].Shift, exactly as before.
type sectionPlan struct {
	layout.Pagination
	section      map[string]bool
	sectionShift []geom.Length
}

// shiftFor is the window shift for elementID's content on page.
func (p sectionPlan) shiftFor(page int, elementID string) geom.Length {
	if p.section[elementID] {
		return p.sectionShift[page]
	}
	return p.Pages[page].Shift
}

// paginateWithSectionBreak is the ONE pagination both of buildPageModel's
// passes run, so the page count {{pages}} prints and the pages rendered
// always agree. With no break, or an empty section, it is exactly
// paginateWithFooterOrphanFix over every item.
func paginateWithSectionBreak(g layout.PageGeometry, items []layout.ColumnItem, sb sectionBreakSplit) (sectionPlan, []Diagnostic, error) {
	var above, below []layout.ColumnItem
	if len(sb.members) > 0 {
		for _, it := range items {
			if sb.members[it.ElementID] {
				below = append(below, it)
			} else {
				above = append(above, it)
			}
		}
	}
	if len(below) == 0 {
		plan, diags, err := paginateWithFooterOrphanFix(g, items, footerOrphanTargetsFrom(items))
		return sectionPlan{Pagination: plan}, diags, err
	}

	planA, pagesA, diagsA, err := paginateWithFooterOrphanFixPages(g, above, footerOrphanTargetsFrom(above))
	if err != nil {
		return sectionPlan{}, nil, err
	}
	shared := aboveLineEndsAtOrAbove(g, above, planA, pagesA, sb.line)
	if shared && len(planA.Pages) == 1 {
		// CAP-3's fast path: nothing moves, so paginate exactly as the
		// document without the key does.
		plan, diags, err := paginateWithFooterOrphanFix(g, items, footerOrphanTargetsFrom(items))
		return sectionPlan{Pagination: plan}, diags, err
	}

	planS, _, diagsS, err := paginateWithFooterOrphanFixPages(g, below, footerOrphanTargetsFrom(below))
	if err != nil {
		return sectionPlan{}, nil, err
	}

	pages := append([]layout.PageAssignment(nil), planA.Pages...)
	shifts := make([]geom.Length, len(pages))
	for i := range pages {
		shifts[i] = pages[i].Shift
	}
	offset := len(pages)
	first := 0
	if shared {
		offset = len(pages) - 1
		pages[offset] = mergePageAssignments(pages[offset], planS.Pages[0])
		shifts[offset] = planS.Pages[0].Shift
		first = 1
	}
	for _, pa := range planS.Pages[first:] {
		pages = append(pages, pa)
		shifts = append(shifts, pa.Shift)
	}

	out := sectionPlan{
		Pagination: layout.Pagination{
			Pages:      pages,
			Suppressed: append([]layout.TableHeaderSuppressed(nil), planA.Suppressed...),
			Clipped:    append([]layout.TableRowClipped(nil), planA.Clipped...),
		},
		section:      sb.members,
		sectionShift: shifts,
	}
	for _, s := range planS.Suppressed {
		s.Page += offset
		out.Suppressed = append(out.Suppressed, s)
	}
	for _, c := range planS.Clipped {
		c.Page += offset
		out.Clipped = append(out.Clipped, c)
	}
	return out, append(diagsA, diagsS...), nil
}

// aboveLineEndsAtOrAbove reports whether the above-line content ends at or
// above line on its last page: the lowest page-space bottom of any item on
// that page, counting a table row's displacement, a floor push and a table's
// floored slice, is at or above the line. A group clipped on that page ran
// past the content bottom, so it crosses.
func aboveLineEndsAtOrAbove(g layout.PageGeometry, items []layout.ColumnItem, plan layout.Pagination, itemPages []int, line geom.Length) bool {
	last := len(plan.Pages) - 1
	for _, c := range plan.Clipped {
		if c.Page == last {
			return false
		}
	}
	pa := plan.Pages[last]
	end := layout.Origins(g).Content
	for i, it := range items {
		if itemPages[i] != last {
			continue
		}
		bottom := it.Bottom - pa.Shift + elementPushFor(pa.ElementPush, it.ElementID)
		if it.Group.Present && !it.Group.Key.IsHeader && !strings.HasPrefix(it.Group.Key.ElementID, keepTogetherKeyPrefix) {
			bottom += rowDisplacementFor(pa.RowDisplacement, it.ElementID)
		}
		if bottom > end {
			end = bottom
		}
	}
	for _, sl := range pa.TableSlices {
		if sl.Bottom > end {
			end = sl.Bottom
		}
	}
	return end <= line
}

// mergePageAssignments puts the section's first page onto the above-line
// content's last page. Content refs are merged in ascending order, which is
// authored emission order (both passes build items so that refs ascend within
// each kind). Every other per-page list is keyed by element or carries its
// own shift, so it is concatenated.
func mergePageAssignments(a, s layout.PageAssignment) layout.PageAssignment {
	out := a
	out.ContentRuns = mergeAscending(a.ContentRuns, s.ContentRuns)
	out.ContentImages = mergeAscending(a.ContentImages, s.ContentImages)
	out.ContentRects = mergeAscending(a.ContentRects, s.ContentRects)
	out.HeaderRepeats = append(append([]layout.TableHeaderRepeat(nil), a.HeaderRepeats...), s.HeaderRepeats...)
	out.RowDisplacement = append(append([]layout.TableRowDisplacement(nil), a.RowDisplacement...), s.RowDisplacement...)
	out.ClippedRects = append(append([]layout.RectClip(nil), a.ClippedRects...), s.ClippedRects...)
	out.TableSlices = append(append([]layout.TableSlice(nil), a.TableSlices...), s.TableSlices...)
	out.ElementPush = append(append([]layout.ElementPush(nil), a.ElementPush...), s.ElementPush...)
	return out
}

// mergeAscending merges two ascending ref lists into one ascending list. On a
// tie the above-line ref goes first.
func mergeAscending[T ~int](a, b []T) []T {
	if len(b) == 0 {
		return a
	}
	out := make([]T, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if b[j] < a[i] {
			out = append(out, b[j])
			j++
		} else {
			out = append(out, a[i])
			i++
		}
	}
	out = append(out, a[i:]...)
	return append(out, b[j:]...)
}
