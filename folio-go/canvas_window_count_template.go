package folio

// Story 7.5 — the FIRST multi-page CANVAS fixtures.
//
// WHY THEY HAD TO BE NEW. No canvas or projection fixture in this repository
// was multi-page, and none COULD be: containComponent's one-window cap made a
// template that placed anything below the foot of page one unauthorable, and
// multi_page_fixture_test.go, statement_fixture_test.go and
// page_count_matrix_test.go never call Canvas or CanvasWithTextPaint at all.
// Every canvas assertion about pagination was therefore vacuous by
// construction, which is how the closed form these two fixtures exist to
// refuse could have survived unnoticed.
//
// WHY fixtures/page-count-{1,5,20,50} COULD NOT BE REUSED. Their elements sit
// exactly one window apart, which makes `ceil(lowestBottom / H)` and the true
// slide count COINCIDE at every N. They exercise the code without
// discriminating it. Discriminating needs a GAP.
//
// THE GEOMETRY, shared by both constants and identical to the page-count
// matrix's: A4 portrait, margins {top 30, right 54, bottom 42, left 36}, page
// header 18, page footer 24. So one content window is
//
//	841890 − 30000 − 42000 − 18000 − 24000 = 727890 millipoints (727.89pt)
//
// which internal/layout's ContentHeight derives and CanvasProjection reports
// as contentWindowHeight.

// canvasWindowCountGapTemplateJSON is the DISCRIMINATING fixture: one line of
// text at the top of the column, and a second element declared TEN WINDOWS
// BELOW it with nothing in between.
//
// The right answer is TWO. The window advances to the top of the first item
// that did not fit, so the gap costs nothing: page one holds the text, page
// two starts at the rect. The closed form paginate.go forbids by name answers
// ELEVEN — nine of them blank pages that no document has — so this fixture
// red-proves that exact spelling rather than merely exercising the count.
//
// The rect carries NO STYLE on purpose. The render path skips unstyled
// element boxes, correctly, because an unstyled rect prints nothing; the
// canvas paints every component's box, and this count is a claim about the
// column as the canvas paints it. An unstyled element that still occupies a
// window is that difference, asserted.
const canvasWindowCountGapTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 20, "value": "Opening balance", "style": {"fontFamily": "body", "fontSize": 12}},
        {"id": "e2", "type": "rect", "x": 0, "y": 7280, "width": 200, "height": 20}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 24
    },
    "pageHeader": {
      "elements": [],
      "height": 18
    }
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 42, "left": 36, "right": 54, "top": 30}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// canvasWindowCountControlTemplateJSON is the NEGATIVE CONTROL: three
// elements one window apart, at y = 0, 728 and 1456 points.
//
// Both routes answer THREE here — the slide and the closed form agree,
// exactly as they agree on every page-count-N fixture. It is in the suite for
// two reasons: without it the discriminating test could pass by always
// answering two, and with it the record says out loud WHY the forbidden
// spelling survived this long. One-window spacing hides the defect.
const canvasWindowCountControlTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 20, "value": "Window one", "style": {"fontFamily": "body", "fontSize": 12}},
        {"id": "e2", "type": "text", "x": 0, "y": 728, "width": 200, "height": 20, "value": "Window two", "style": {"fontFamily": "body", "fontSize": 12}},
        {"id": "e3", "type": "text", "x": 0, "y": 1456, "width": 200, "height": 20, "value": "Window three", "style": {"fontFamily": "body", "fontSize": 12}}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 24
    },
    "pageHeader": {
      "elements": [],
      "height": 18
    }
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 42, "left": 36, "right": 54, "top": 30}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// canvasWindowCountOversizedTemplateJSON declares a content component TALLER
// THAN ONE WINDOW — 900pt in a 727.89pt column. Story 7.5 newly makes such a
// component authorable, so layout.Paginate's OverflowError is newly reachable
// from the canvas, and the projection must degrade the COUNT rather than fail
// and blank the canvas.
const canvasWindowCountOversizedTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "rect", "x": 0, "y": 0, "width": 200, "height": 900, "style": {"background": "#eeeeee"}}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 24
    },
    "pageHeader": {
      "elements": [],
      "height": 18
    }
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 2,
  "page": {"margin": {"bottom": 42, "left": 36, "right": 54, "top": 30}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// canvasWindowCountBoundTableTemplateJSON is the FLOOR, recorded rather than
// discovered later: a table bound to a collection the canvas has never been
// given. projectedSize returns a table's header height and no rows, because
// the canvas has no data — so the column the canvas paints is one header tall
// however many hundred rows the finished statement runs to.
const canvasWindowCountBoundTableTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "as": "row", "bind": "transactions[]", "headerHeight": 16, "columns": [{"id": "e2", "label": "Date", "width": 80, "align": "left", "bind": "{{row.date}}"}], "style": {"fontFamily": "body", "fontSize": 8}}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 24
    },
    "pageHeader": {
      "elements": [],
      "height": 18
    }
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 42, "left": 36, "right": 54, "top": 30}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// canvasWindowCountUnshapedTextTemplateJSON is the floor's THIRD cause, the
// one neither a table nor an over-tall box produces: a content text element
// that HAS a value and no style.fontFamily to resolve a chain from. fontChain
// refuses it, addCanvasTextPaint degrades that one element to an empty paint
// — the pre-existing, deliberate behaviour, because such a document is
// structurally valid and still loadable — and the extents its lines would
// have contributed are simply missing from the column the count measures.
//
// The second element is shaped normally, so the fixture discriminates: the
// column is counted, it is just counted short.
const canvasWindowCountUnshapedTextTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 20, "value": "No chain resolves this"},
        {"id": "e2", "type": "text", "x": 0, "y": 40, "width": 200, "height": 20, "value": "Shaped normally", "style": {"fontFamily": "body", "fontSize": 12}}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 24
    },
    "pageHeader": {
      "elements": [],
      "height": 18
    }
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 42, "left": 36, "right": 54, "top": 30}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
