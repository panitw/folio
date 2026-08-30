// Story 9.1's own suite: an element's style.background and style.border
// reach the page model. Before this story they reached nothing outside a
// table's cell chrome — see element_box.go's doc comment for the trace.
package folio

import (
	"errors"
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/pagemodel"
)

// boxTemplateJSON builds a three-band document whose content element e1
// carries styleFields verbatim (a style block's inner fields, e.g.
// `"background": "#112233", `). e2 is an unstyled sibling, so every
// assertion about "the boxes" is also an assertion that an element
// declaring no box contributes none.
func boxTemplateJSON(styleFields string) string {
	return `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 12, "y": 20, "width": 200, "height": 30, "value": "Boxed", "style": {` + styleFields + `"fontFamily": "body", "fontSize": 12}},
        {"id": "e2", "type": "text", "x": 0, "y": 100, "width": 200, "height": 20, "value": "Plain", "style": {"fontFamily": "body", "fontSize": 12}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
}

// boxPages renders doc through the SAME function Render() calls
// internally, so what these tests read is the page model that becomes
// the PDF, not a parallel derivation of it.
func boxPages(t *testing.T, doc string) []pagemodel.Page {
	t.Helper()
	tpl, err := ParseTemplate([]byte(doc))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	pages, _, _, _, err := buildPageModel(tpl, mustDecodeData(t, `{"customer":{"flag":false}}`), mustDecodeParams(t), testFontSet())
	if err != nil {
		t.Fatalf("buildPageModel: %v", err)
	}
	return pages
}

// contentBandOrigin re-derives where the content band starts from the
// template's own geometry, independently of the collector under test.
func contentBandOrigin(t *testing.T, doc string) geom.Length {
	t.Helper()
	tpl, err := ParseTemplate([]byte(doc))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	g, err := pageGeometryOf(tpl)
	if err != nil {
		t.Fatalf("pageGeometryOf: %v", err)
	}
	return layout.Origins(g).Content
}

func TestElementBackgroundFillsTheDeclaredBox(t *testing.T) {
	doc := boxTemplateJSON(`"background": "#112233", `)
	pages := boxPages(t, doc)
	if len(pages) != 1 {
		t.Fatalf("pages = %d, want 1", len(pages))
	}
	if len(pages[0].Rects) != 1 {
		t.Fatalf("page rects = %d, want exactly 1 (e1's box; e2 declares none)", len(pages[0].Rects))
	}
	rect := pages[0].Rects[0]
	want := pagemodel.Rect{
		X: 12000, Y: contentBandOrigin(t, doc) + 20000, W: 200000, H: 30000,
		HasFill: true, Fill: pagemodel.Color{R: 0x11, G: 0x22, B: 0x33},
	}
	if rect != want {
		t.Errorf("e1's box = %+v, want %+v — the box is the element's DECLARED rectangle, placed in its band", rect, want)
	}
}

func TestElementBorderStrokesWithTheTableCellDefaults(t *testing.T) {
	// No width, no colour, no edges: the defaults a table cell already
	// resolves through this same builder — 0.5pt, #000000, all four edges.
	rect := boxPages(t, boxTemplateJSON(`"border": {}, `))[0].Rects[0]
	if rect.HasFill {
		t.Errorf("a border-only box has a fill (%+v) — a declaration the element never made", rect.Fill)
	}
	if !rect.HasStroke || rect.Stroke != (pagemodel.Color{}) || rect.StrokeWidth != 500 {
		t.Errorf("border defaults = {stroke:%v color:%+v width:%d}, want {true #000000 500}", rect.HasStroke, rect.Stroke, rect.StrokeWidth)
	}
	if (rect.Edges != pagemodel.RectEdges{Top: true, Right: true, Bottom: true, Left: true}) {
		t.Errorf("border edges = %+v, want all four", rect.Edges)
	}

	declared := boxPages(t, boxTemplateJSON(`"border": {"color": "#c81e1e", "width": 2, "edges": ["bottom"]}, `))[0].Rects[0]
	if !declared.HasStroke || declared.Stroke != (pagemodel.Color{R: 0xc8, G: 0x1e, B: 0x1e}) || declared.StrokeWidth != 2000 {
		t.Errorf("declared border = {stroke:%v color:%+v width:%d}, want {true #c81e1e 2000}", declared.HasStroke, declared.Stroke, declared.StrokeWidth)
	}
	if (declared.Edges != pagemodel.RectEdges{Bottom: true}) {
		t.Errorf("declared edges = %+v, want bottom alone", declared.Edges)
	}
}

func TestElementBoxCarriesBothFillAndStrokeAtOnce(t *testing.T) {
	rect := boxPages(t, boxTemplateJSON(`"background": "#0b1120", "border": {"color": "#ffffff"}, `))[0].Rects[0]
	if !rect.HasFill || !rect.HasStroke {
		t.Fatalf("box = {fill:%v stroke:%v}, want both — one rectangle carries both declarations", rect.HasFill, rect.HasStroke)
	}
}

func TestHiddenElementDrawsNoBox(t *testing.T) {
	doc := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 30, "value": "Hidden", "visibleIf": "customer.flag", "style": {"background": "#112233", "fontFamily": "body", "fontSize": 12}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	// mustDecodeData supplies customer.flag = false, so e1 is hidden.
	if rects := boxPages(t, doc)[0].Rects; len(rects) != 0 {
		t.Errorf("a hidden element contributed %d rect(s): %+v — AD-24 makes it absent from the page model, box included", len(rects), rects)
	}
}

// parse_bands.go makes width and height required on every non-table
// element, so "no rectangle" reaches the renderer only as a zero or
// negative one — which the loader accepts.
func TestBoxWithoutADeclaredRectangleDrawsNothing(t *testing.T) {
	doc := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 0, "value": "Unsized", "style": {"background": "#112233", "fontFamily": "body", "fontSize": 12}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	if rects := boxPages(t, doc)[0].Rects; len(rects) != 0 {
		t.Errorf("an element whose declared rectangle has no area drew %d rect(s) — there is no rectangle to draw", len(rects))
	}
}

func TestRectAndLineElementsDrawTheirBox(t *testing.T) {
	doc := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "rect", "x": 0, "y": 0, "width": 120, "height": 40, "style": {"background": "#1b2a4a", "border": {"color": "#ffffff", "width": 1}}},
        {"id": "e2", "type": "line", "x": 0, "y": 60, "width": 300, "height": 1, "style": {"background": "#000000"}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	rects := boxPages(t, doc)[0].Rects
	if len(rects) != 2 {
		t.Fatalf("rect+line drew %d rect(s), want 2 — before Story 9.2 neither kind reached the page model at all", len(rects))
	}
	if !rects[0].HasFill || !rects[0].HasStroke {
		t.Errorf("the rect element = {fill:%v stroke:%v}, want both", rects[0].HasFill, rects[0].HasStroke)
	}
	if !rects[1].HasFill || rects[1].H != 1000 || rects[1].W != 300000 {
		t.Errorf("the line element = {fill:%v w:%d h:%d}, want a filled bar 300000 x 1000 — its declared height IS its thickness", rects[1].HasFill, rects[1].W, rects[1].H)
	}
}

func TestHeaderAndFooterBoxesRepeatOnEveryPage(t *testing.T) {
	// Two content elements far enough apart to force a second page, so the
	// repetition is observed rather than assumed.
	doc := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e3", "type": "text", "x": 0, "y": 0, "width": 200, "height": 20, "value": "First", "style": {"fontFamily": "body", "fontSize": 12}},
        {"id": "e4", "type": "text", "x": 0, "y": 900, "width": 200, "height": 20, "value": "Second", "style": {"fontFamily": "body", "fontSize": 12}}
      ]
    },
    "pageFooter": {"elements": [{"id": "e2", "type": "rect", "x": 0, "y": 4, "width": 400, "height": 2, "style": {"background": "#888888"}}], "height": 20},
    "pageHeader": {"elements": [{"id": "e1", "type": "rect", "x": 0, "y": 4, "width": 400, "height": 2, "style": {"background": "#444444"}}], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 5,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	pages := boxPages(t, doc)
	if len(pages) < 2 {
		t.Fatalf("pages = %d, want at least 2 — the repetition is untested on a one-page document", len(pages))
	}
	for i, page := range pages {
		if len(page.Rects) != 2 {
			t.Errorf("page %d carries %d rect(s), want the page-header box and the page-footer box on every page", i+1, len(page.Rects))
			continue
		}
		if page.Rects[0].Fill != (pagemodel.Color{R: 0x44, G: 0x44, B: 0x44}) || page.Rects[1].Fill != (pagemodel.Color{R: 0x88, G: 0x88, B: 0x88}) {
			t.Errorf("page %d fills = %+v, %+v, want the header's then the footer's", i+1, page.Rects[0].Fill, page.Rects[1].Fill)
		}
	}
}

func TestATableStyleIsNeverPaintedTwice(t *testing.T) {
	// e1's style is a TABLE's style: Epic 4 already paints it as the cell
	// chrome. Story 9.1 must not add an element box on top of it.
	doc := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "rows[]", "as": "row", "headerHeight": 20,
          "style": {"background": "#112233", "fontFamily": "body", "fontSize": 10},
          "columns": [{"id": "e2", "label": "A", "width": 100, "align": "left", "bind": "{{row.a}}"}]}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	tpl, err := ParseTemplate([]byte(doc))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	pages, _, _, _, err := buildPageModel(tpl, mustDecodeData(t, `{"rows":[{"a":"one"}],"customer":{"flag":false}}`), mustDecodeParams(t), testFontSet())
	if err != nil {
		t.Fatalf("buildPageModel: %v", err)
	}
	// One header cell + one body cell for the single column and single row.
	// An element box would be a THIRD rect spanning the table's own extent.
	if got := len(pages[0].Rects); got != 2 {
		t.Errorf("a styled table produced %d rect(s), want 2 (header cell, body cell) — its style must paint as cell chrome only, never a second time as an element box", got)
	}
}

func TestAnElementBoxPaintsUnderItsOwnText(t *testing.T) {
	// Page assembly draws every rect before every run, so a background can
	// never cover the text it sits behind. Asserted on the page model
	// rather than on the PDF bytes: pagemodel.Page is where the order is
	// decided (AD-5).
	page := boxPages(t, boxTemplateJSON(`"background": "#112233", `))[0]
	if len(page.Rects) == 0 || len(page.Runs) == 0 {
		t.Fatalf("page has rects=%d runs=%d, want both", len(page.Rects), len(page.Runs))
	}
}

func TestAMalformedElementColourIsALocatedRenderError(t *testing.T) {
	tpl, err := ParseTemplate([]byte(boxTemplateJSON(`"background": "red", `)))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, _, _, _, err = buildPageModel(tpl, mustDecodeData(t, `{"customer":{"flag":false}}`), mustDecodeParams(t), testFontSet())
	if err == nil {
		t.Fatal("a non-#RRGGBB element background rendered without error")
	}
	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("error is %T (%v), want a located *RenderError naming the element and the field", err, err)
	}
	if re.Diagnostic.ElementID != "e1" || re.Diagnostic.Code != DiagCodeStyleColorInvalid {
		t.Errorf("error = {element:%q code:%q}, want {e1 %s}", re.Diagnostic.ElementID, re.Diagnostic.Code, DiagCodeStyleColorInvalid)
	}
}

// Story 9.2: a placed line and a placed rect are visible without the
// author styling them first. Both assertions read the PROJECTION — the
// same values the canvas paints from and the inspector shows — so a
// default that existed only in the document, or only on the canvas,
// fails here. Millipoints throughout: AD-23 forbids a float anywhere
// under the module, tests included.
func TestAPlacedLineAndRectArriveVisible(t *testing.T) {
	for _, tc := range []struct {
		kind       string
		wantFill   string
		wantEdge   string
		wantWidth  int64
		wantHeight int64
	}{
		{kind: "line", wantFill: "#000000", wantHeight: 1000},
		{kind: "rect", wantEdge: "#000000", wantWidth: 1000, wantHeight: 24000},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			tpl := componentTemplate(t)
			before, err := Canvas(tpl)
			if err != nil {
				t.Fatal(err)
			}
			after, err := ApplyComponentCommand(tpl, []byte(`{"kind":"dropComponent","version":1,"type":"`+tc.kind+`","x":72,"y":180,"snap":false}`))
			if err != nil {
				t.Fatal(err)
			}
			placed := newProjectedComponent(t, before, after)
			if got := derefString(placed.Background); got != tc.wantFill {
				t.Errorf("%s background = %q, want %q", tc.kind, got, tc.wantFill)
			}
			if got := derefString(placed.BorderColor); got != tc.wantEdge {
				t.Errorf("%s border colour = %q, want %q", tc.kind, got, tc.wantEdge)
			}
			if got := derefLength(placed.BorderWidth); got != tc.wantWidth {
				t.Errorf("%s border width = %d millipoints, want %d", tc.kind, got, tc.wantWidth)
			}
			if placed.Height != tc.wantHeight {
				t.Errorf("%s dropped %d millipoints tall, want %d — a rule's declared height is its thickness", tc.kind, placed.Height, tc.wantHeight)
			}
		})
	}
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefLength(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
