package folio

import (
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/template"
)

// This file is the engine's one reading of SPEC-multi-pages' designed pages.
// Every content-element reader goes through contentElements, contentPageIndex
// or firstContentBand, never through Bands.Content: a multi-page document
// holds its content in Document.Pages and leaves Bands.Content empty.

// contentPagesSplit is the document's designed pages, resolved for
// pagination: which page each content item belongs to, and each page's own
// section break.
type contentPagesSplit struct {
	// pageOf maps a content element id to its page. Nil for a one-page
	// document, whose every item is page 0.
	pageOf map[string]int
	// breaks holds one section break per designed page; the zero value is
	// "no break". Only page 1 may declare one in this story.
	breaks []sectionBreakSplit
}

// contentPagesOf resolves t's designed pages against g.
func contentPagesOf(t *Template, g layout.PageGeometry) contentPagesSplit {
	split := contentPagesSplit{breaks: make([]sectionBreakSplit, t.doc.PageCount())}
	split.breaks[0] = sectionBreakOf(t, g)
	if len(split.breaks) > 1 {
		split.pageOf = contentPageIndex(t)
	}
	return split
}

// partition splits items by designed page, keeping each page's items in
// their incoming order. A one-page document's items are returned as they are.
func (s contentPagesSplit) partition(items []layout.ColumnItem) [][]layout.ColumnItem {
	perPage := make([][]layout.ColumnItem, len(s.breaks))
	if len(perPage) == 1 {
		perPage[0] = items
		return perPage
	}
	for _, it := range items {
		page := s.pageOf[it.ElementID]
		perPage[page] = append(perPage[page], it)
	}
	return perPage
}

// contentPagesPlan is the whole document's pagination: each designed page
// paginated as its own column (with its own section break) and the output
// pages concatenated in page order. Every page starts a new output page after
// the previous page's last one — in this story every page renders as if its
// Page Break were on.
type contentPagesPlan struct {
	// Pagination holds the concatenated output pages, with Suppressed and
	// Clipped page numbers restated in the document's output page numbering.
	layout.Pagination
	plans        []sectionPlan
	designedPage []int
	localPage    []int
}

// outputShiftFor is the window shift for elementID's content on output page.
func (p contentPagesPlan) outputShiftFor(page int, elementID string) geom.Length {
	return p.plans[p.designedPage[page]].shiftFor(p.localPage[page], elementID)
}

// paginateContentPages is the ONE pagination both of buildPageModel's passes
// run, so {{pages}} sums every designed page's output pages exactly as they
// are rendered. A one-page document is paginated exactly as before.
func paginateContentPages(g layout.PageGeometry, items []layout.ColumnItem, split contentPagesSplit) (contentPagesPlan, []Diagnostic, error) {
	var out contentPagesPlan
	var diags []Diagnostic
	for page, pageItems := range split.partition(items) {
		plan, pageDiags, err := paginateWithSectionBreak(g, pageItems, split.breaks[page])
		if err != nil {
			return contentPagesPlan{}, nil, err
		}
		offset := len(out.Pages)
		for local := range plan.Pages {
			out.designedPage = append(out.designedPage, page)
			out.localPage = append(out.localPage, local)
		}
		out.Pages = append(out.Pages, plan.Pages...)
		for _, s := range plan.Suppressed {
			s.Page += offset
			out.Suppressed = append(out.Suppressed, s)
		}
		for _, c := range plan.Clipped {
			c.Page += offset
			out.Clipped = append(out.Clipped, c)
		}
		out.plans = append(out.plans, plan)
		diags = append(diags, pageDiags...)
	}
	return out, diags, nil
}

// contentElements returns every designed page's content elements, in page
// order. For a one-page document it is that page's own slice.
func contentElements(t *Template) []template.Element {
	bands := t.doc.ContentBands()
	if len(bands) == 1 {
		return bands[0].Elements
	}
	var out []template.Element
	for _, band := range bands {
		out = append(out, band.Elements...)
	}
	return out
}

// firstContentBand is page 1's content band: the band commands that create
// or drop into `content` target, and the one band a section break may be
// declared on in this story.
func firstContentBand(t *Template) *template.Band {
	return t.doc.ContentBands()[0]
}

// contentPageIndex maps every content element's id to its designed page's
// index. Ids are unique document-wide, so the map is total over the content
// elements. Looked up, never ranged.
func contentPageIndex(t *Template) map[string]int {
	out := map[string]int{}
	for page, band := range t.doc.ContentBands() {
		for _, el := range band.Elements {
			out[string(el.ID)] = page
		}
	}
	return out
}
