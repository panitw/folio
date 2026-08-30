package folio

import (
	"fmt"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/layout"
)

func parseWindowCountTemplate(t *testing.T, source string) *Template {
	t.Helper()
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse window-count fixture: %v", err)
	}
	return tpl
}

func projectWithPaint(t *testing.T, tpl *Template) CanvasProjection {
	t.Helper()
	projection, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatalf("CanvasWithTextPaint: %v", err)
	}
	return projection
}

// renderPathWindows is the RENDER path's own answer, built the way
// paginateDocument's page-count pass builds it: document bands, the band's
// shaped text runs, contentColumnItems, layout.Paginate. It is a different
// route to the same integer, and Story 7.5's projection must agree with it
// rather than assume it does — the projection reads the CANVAS paint plan's
// extents, which is a second consumer of one shaping, not a second shaping.
func renderPathWindows(t *testing.T, tpl *Template) int {
	t.Helper()
	data := emptyBindValue(t)
	bands, err := documentBands(tpl)
	if err != nil {
		t.Fatalf("documentBands: %v", err)
	}
	runs, _, _, err := collectBandTextRuns(tpl, bands, contentBandIndex, data, data, testFontSet(), newFontCache(), contentBandResolver, nil)
	if err != nil {
		t.Fatalf("collectBandTextRuns: %v", err)
	}
	plan, err := layout.Paginate(mustPageGeometry(t, tpl), contentColumnItems(runs, nil, nil, nil))
	if err != nil {
		t.Fatalf("layout.Paginate: %v", err)
	}
	return len(plan.Pages)
}

// TestCanvasReportsTheWindowHeightAndTheWindowCount is AC4's discriminating
// assertion, and the whole reason canvas_window_count_template.go exists.
//
// The gap fixture declares one line of text at the top of the column and a
// rect TEN WINDOWS BELOW IT. Two is the answer: the window advances to the
// top of the first item that did not fit, so an element declared far below
// the text STARTS the next window rather than generating blank pages before
// it. `ceil(lowestBottom / contentWindowHeight)` answers eleven, so this test
// is red under the one spelling internal/layout/paginate.go forbids by name —
// which the page-count-N fixtures, spaced exactly one window apart, cannot
// say, because there the two routes agree at every N.
func TestCanvasReportsTheWindowHeightAndTheWindowCount(t *testing.T) {
	gap := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountGapTemplateJSON))
	// The window height is ContentHeight's number, and the content band
	// rectangle is one window, so the two must be the same integer: a
	// divergence would have the canvas and the engine drawing different
	// pages while agreeing on every byte.
	if gap.ContentWindowHeight != 727890 {
		t.Fatalf("contentWindowHeight = %d, want 727890", gap.ContentWindowHeight)
	}
	if band := gap.Bands[1]; band.Name != "content" || band.Height != gap.ContentWindowHeight {
		t.Fatalf("content band %q height %d, want it equal to the window height %d", band.Name, band.Height, gap.ContentWindowHeight)
	}
	if gap.ContentWindowCount != 2 {
		closedForm := (int64(7280000+20000) + gap.ContentWindowHeight - 1) / gap.ContentWindowHeight
		t.Fatalf("contentWindowCount = %d, want 2 (the forbidden closed form answers %d here)", gap.ContentWindowCount, closedForm)
	}
	// The negative control: one window apart, both routes answer three. It
	// is here so the test above cannot pass by always answering two, and so
	// the record says why the forbidden spelling survived this long.
	control := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON))
	if control.ContentWindowCount != 3 {
		t.Fatalf("control contentWindowCount = %d, want 3", control.ContentWindowCount)
	}
	if control.ContentWindowHeight != gap.ContentWindowHeight {
		t.Fatalf("the two fixtures share a geometry but report windows of %d and %d", control.ContentWindowHeight, gap.ContentWindowHeight)
	}
}

// TestCanvasWindowCountIsIndependentOfPaintTruncation re-asserts D-7.4.2 §5
// FROM THE PROJECTION SIDE. canvas_body_text_bounds_test.go's
// TestPaginationIsIndependentOfCanvasPaintTruncation guards the RENDER path's
// oracle and stays exactly as it is; this one guards the number Story 7.5
// actually ships, which is computed by a different route.
//
// The discrimination is the second half: a document containing only the
// PAINTED PREFIX occupies visibly fewer windows, so a count that had read the
// paint's line list would be wrong here by a wide margin rather than
// coincidentally right.
func TestCanvasWindowCountIsIndependentOfPaintTruncation(t *testing.T) {
	line := "clause"
	long := strings.TrimSuffix(strings.Repeat(line+"\n", maxCanvasBodyTextLines+400), "\n")
	tpl := bodyTextDocument(t, long, `{"fontFamily":"body","fontSize":12}`)
	projection := projectWithPaint(t, tpl)
	paint := paintOf(t, projection, "e1")
	if paint == nil || !paint.Truncated {
		t.Fatalf("presence precondition: this fixture must TRUNCATE for the test to say anything: %#v", paint)
	}
	if want := renderPathWindows(t, tpl); projection.ContentWindowCount != int64(want) {
		t.Fatalf("the canvas counts %d windows and the render path %d for the same document", projection.ContentWindowCount, want)
	}
	// What the paint alone would have implied.
	prefix := strings.TrimSuffix(strings.Repeat(line+"\n", len(paint.Lines)), "\n")
	painted := projectWithPaint(t, bodyTextDocument(t, prefix, `{"fontFamily":"body","fontSize":12}`))
	if painted.ContentWindowCount >= projection.ContentWindowCount {
		t.Fatalf("the painted prefix occupies %d windows and the whole document %d; the two must differ for this test to discriminate", painted.ContentWindowCount, projection.ContentWindowCount)
	}
}

// TestCanvasWindowCountAgreesWithTheRenderPathOracle measures the agreement
// the story must not assume. The projection builds its ColumnItems from the
// canvas paint plan's extents; the render path builds its own from shaped
// runs carrying glyphs, faces and CIDs. Same shaping, same vertical model,
// two builders — so equal counts are a property to check, not a given.
func TestCanvasWindowCountAgreesWithTheRenderPathOracle(t *testing.T) {
	// The one-window case is the control with its two later elements removed,
	// so the pair differs in exactly the thing under test.
	oneWindow := canvasWindowCountControlTemplateJSON
	for _, drop := range []string{
		`,
        {"id": "e2", "type": "text", "x": 0, "y": 728, "width": 200, "height": 20, "value": "Window two", "style": {"fontFamily": "body", "fontSize": 12}}`,
		`,
        {"id": "e3", "type": "text", "x": 0, "y": 1456, "width": 200, "height": 20, "value": "Window three", "style": {"fontFamily": "body", "fontSize": 12}}`,
	} {
		if !strings.Contains(oneWindow, drop) {
			t.Fatal("the control fixture moved; this test's edit no longer applies to it")
		}
		oneWindow = strings.Replace(oneWindow, drop, "", 1)
	}
	for name, source := range map[string]string{
		"one window":    oneWindow,
		"three windows": canvasWindowCountControlTemplateJSON,
	} {
		tpl := parseWindowCountTemplate(t, source)
		projection := projectWithPaint(t, tpl)
		if want := renderPathWindows(t, tpl); projection.ContentWindowCount != int64(want) {
			t.Fatalf("%s: canvas counts %d, render path counts %d", name, projection.ContentWindowCount, want)
		}
	}
}

// TestCanvasWindowCountDegradesRatherThanFailingTheProjection is Story 7.5's
// Ruling C. Lifting the content band's cap makes a component taller than one
// window authorable, so layout.Paginate's OverflowError becomes reachable
// from the canvas for the first time. Turning it into a projection failure
// would make a canvas bound into a document validity rule and would blank the
// canvas with no attributable error; reporting ONE window is Paginate's own
// answer for a column it cannot place.
func TestCanvasWindowCountDegradesRatherThanFailingTheProjection(t *testing.T) {
	tpl := parseWindowCountTemplate(t, canvasWindowCountOversizedTemplateJSON)
	projection, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatalf("an over-tall content component failed the projection: %v", err)
	}
	if projection.ContentWindowCount != 1 {
		t.Fatalf("over-tall contentWindowCount = %d, want the degraded floor of 1", projection.ContentWindowCount)
	}
	if len(projection.Components) != 1 || projection.Components[0].Height != 900000 {
		t.Fatalf("the component itself must survive the degradation: %#v", projection.Components)
	}
	// The precondition, stated rather than trusted: the geometry really is
	// past one window, so the degradation branch really is the one taken.
	if geomHeight := projection.Components[0].Height; geomHeight <= projection.ContentWindowHeight {
		t.Fatalf("component height %d does not exceed one window %d", geomHeight, projection.ContentWindowHeight)
	}
}

// TestCanvasWindowCountIsAFloorForABoundTable records, rather than lets Story
// 7.6 discover by drawing one sheet for a fifty-page statement, that the
// count describes THE COLUMN AS THE CANVAS PAINTS IT.
//
// projectedSize returns a table's header height and no rows, because the
// canvas has never been given the data. The behaviour is pre-existing (Epic 6
// shipped it) and irreparable inside this story — reusing the render/bind
// machinery is Story 13.4, in another epic — so what 7.5 owes is that the
// number says what it is a number about, in the projection's own comment and
// here.
func TestCanvasWindowCountIsAFloorForABoundTable(t *testing.T) {
	tpl := parseWindowCountTemplate(t, canvasWindowCountBoundTableTemplateJSON)
	projection := projectWithPaint(t, tpl)
	if projection.ContentWindowCount != 1 {
		t.Fatalf("bound-table contentWindowCount = %d, want 1", projection.ContentWindowCount)
	}
	if len(projection.Components) != 1 {
		t.Fatalf("want one projected component, got %d", len(projection.Components))
	}
	// The floor's mechanism, asserted so a future change to projectedSize
	// cannot quietly turn this count into a prediction: the projected table
	// is exactly its header, with no row contributing a millipoint.
	if height := projection.Components[0].Height; height != 16000 {
		t.Fatalf("projected table height = %d, want the header height 16000 and no rows", height)
	}
	if projection.Components[0].Height >= projection.ContentWindowHeight {
		t.Fatal("the fixture must fit one window for the floor to be the interesting fact about it")
	}
}

// TestCanvasReportsOneWindowForAnEmptyColumn pins internal/layout's own rule
// at the projection seam: a document with no content items is ONE page, not
// zero. A zero here would reach Story 7.6 as a canvas with no sheets on it.
func TestCanvasReportsOneWindowForAnEmptyColumn(t *testing.T) {
	tpl := componentTemplate(t)
	bare, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	// Canvas has no FontSet and therefore cannot shape; it reports the
	// documented floor of one window and never reaches the browser, because
	// every projection that does is a CanvasWithTextPaint.
	if bare.ContentWindowCount != 1 {
		t.Fatalf("Canvas contentWindowCount = %d, want the documented floor of 1", bare.ContentWindowCount)
	}
	if bare.ContentWindowHeight != bare.Bands[1].Height {
		t.Fatalf("Canvas window height %d does not match its own content band %d", bare.ContentWindowHeight, bare.Bands[1].Height)
	}
	empty := parseWindowCountTemplate(t, fmt.Sprintf(`{"assets":{},"bands":{"content":{"elements":[]},"pageFooter":{"elements":[],"height":24},"pageHeader":{"elements":[],"height":18}},"fonts":{"body":["Roboto-Regular"]},"locale":"en","nextId":1,"page":{"margin":{"bottom":42,"left":36,"right":54,"top":30},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"%s"}`, "1.0"))
	if projection := projectWithPaint(t, empty); projection.ContentWindowCount != 1 {
		t.Fatalf("empty-column contentWindowCount = %d, want 1", projection.ContentWindowCount)
	}
}
