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

// assertWindowOriginsAreWellFormed states the three properties the browser
// protocol independently re-checks, so a fixture that satisfied the exact
// values below by accident — a stale slice, a duplicated shift — still fails
// here. A projection that fails any of them is not merely wrong: it is
// rejected by parseInbound, which discards the WHOLE snapshot and blanks the
// canvas with no attributable error.
func assertWindowOriginsAreWellFormed(t *testing.T, name string, projection CanvasProjection) {
	t.Helper()
	origins := projection.ContentWindowOrigins
	if origins == nil {
		t.Fatalf("%s: contentWindowOrigins is nil, which marshals to JSON null and blanks the canvas", name)
	}
	if int64(len(origins)) != projection.ContentWindowCount {
		t.Fatalf("%s: %d origins for %d windows; there must be exactly one origin per window", name, len(origins), projection.ContentWindowCount)
	}
	if len(origins) == 0 || origins[0] != 0 {
		t.Fatalf("%s: origins = %v, want a first window beginning at column offset 0", name, origins)
	}
	for i := 1; i < len(origins); i++ {
		if origins[i] <= origins[i-1] {
			t.Fatalf("%s: origins = %v are not strictly increasing at index %d", name, origins, i)
		}
	}
}

// TestCanvasProjectsWhereEachWindowBegins is Story 7.6's AC2, and it
// red-proves the forbidden closed form a SECOND time — more sharply than the
// count can.
//
// The count only distinguishes `index * contentWindowHeight` from the engine's
// answer where the spacing is uneven: on the CONTROL fixture, elements a round
// 728pt apart, both routes answer three. The ORIGINS distinguish it there too,
// because the window height is 727890 and the elements sit at 728000 — so the
// closed form answers [0, 727890, 1455780] where the engine answers
// [0, 728000, 1456000], adrift by 110 millipoints per window. That is small
// enough to survive a casual eye on screen and large enough to assert exactly.
func TestCanvasProjectsWhereEachWindowBegins(t *testing.T) {
	gap := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountGapTemplateJSON))
	assertWindowOriginsAreWellFormed(t, "gap", gap)
	// The gap fixture's second element is declared TEN windows below the
	// text, and the window advances to the top of the first item that did not
	// fit — so window two begins at that element's own Y, not nine windows
	// earlier and not at a multiple of anything.
	if got := gap.ContentWindowOrigins; len(got) != 2 || got[0] != 0 || got[1] != 7280000 {
		t.Fatalf("gap contentWindowOrigins = %v, want [0 7280000]", got)
	}

	control := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON))
	assertWindowOriginsAreWellFormed(t, "control", control)
	closedForm := []int64{0, control.ContentWindowHeight, 2 * control.ContentWindowHeight}
	if got := control.ContentWindowOrigins; len(got) != 3 || got[0] != 0 || got[1] != 728000 || got[2] != 1456000 {
		t.Fatalf("control contentWindowOrigins = %v, want [0 728000 1456000] (the forbidden closed form answers %v)", got, closedForm)
	}
	// Stated rather than left implicit: the two answers really do differ on
	// this fixture, so the assertion above is a discrimination and not a
	// coincidence.
	if closedForm[1] != 727890 || closedForm[2] != 1455780 {
		t.Fatalf("the closed form on this geometry is %v; this test's red proof assumed [0 727890 1455780]", closedForm)
	}
}

// TestCanvasSaysWhenTheWindowCountIsAFloor pins the flag to its three
// documented causes and, just as importantly, to its ABSENCE where none of
// them holds. A flag that were always true would satisfy every positive case
// here and say nothing.
func TestCanvasSaysWhenTheWindowCountIsAFloor(t *testing.T) {
	for _, exact := range []struct {
		name   string
		source string
	}{
		{"gap", canvasWindowCountGapTemplateJSON},
		{"control", canvasWindowCountControlTemplateJSON},
		// The SAME unshapeable text as case (c) below, in the page header
		// instead of the content band. The flag is a statement about the
		// content column, and this column is counted exactly — so the
		// degradation site's `band.name == bandContent` guard is what keeps
		// this false. Deleting that guard left the whole Go suite green.
		{"unshaped header", canvasWindowCountUnshapedHeaderTemplateJSON},
	} {
		projection := projectWithPaint(t, parseWindowCountTemplate(t, exact.source))
		if projection.ContentWindowCountIsFloor {
			t.Fatalf("%s: contentWindowCountIsFloor is true for a column with no table, no degradation and no unshaped text", exact.name)
		}
	}

	// (a) A bound table: the canvas has the header and none of the rows.
	table := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountBoundTableTemplateJSON))
	assertWindowOriginsAreWellFormed(t, "bound table", table)
	if table.ContentWindowCount != 1 || len(table.ContentWindowOrigins) != 1 || !table.ContentWindowCountIsFloor {
		t.Fatalf("bound table: count %d, origins %v, floor %v; want 1, [0] and true", table.ContentWindowCount, table.ContentWindowOrigins, table.ContentWindowCountIsFloor)
	}

	// (b) The Ruling C degradation: a component taller than one window.
	oversized := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountOversizedTemplateJSON))
	assertWindowOriginsAreWellFormed(t, "oversized", oversized)
	if oversized.ContentWindowCount != 1 || len(oversized.ContentWindowOrigins) != 1 || !oversized.ContentWindowCountIsFloor {
		t.Fatalf("oversized: count %d, origins %v, floor %v; want 1, [0] and true", oversized.ContentWindowCount, oversized.ContentWindowOrigins, oversized.ContentWindowCountIsFloor)
	}

	// (c) A content text element whose chain would not resolve contributes no
	// extents, so the column is counted short. The discriminating half is the
	// precondition: the element really is present and really painted nothing.
	unshaped := projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountUnshapedTextTemplateJSON))
	assertWindowOriginsAreWellFormed(t, "unshaped text", unshaped)
	if paint := paintOf(t, unshaped, "e1"); paint == nil || len(paint.Lines) != 0 {
		t.Fatalf("precondition: e1 must degrade to an empty paint for this case to say anything: %#v", paint)
	}
	if paint := paintOf(t, unshaped, "e2"); paint == nil || len(paint.Lines) == 0 {
		t.Fatalf("precondition: e2 must shape normally, so the column is counted and merely counted SHORT: %#v", paint)
	}
	if !unshaped.ContentWindowCountIsFloor {
		t.Fatal("unshaped text: contentWindowCountIsFloor is false, but one content element's lines are missing from the column that was counted")
	}
}

// TestCanvasOriginsForAnEmptyColumnAndForTheShapelessEntryPoint covers the two
// ends of the range: a column with nothing in it, and the entry point that
// cannot shape at all.
func TestCanvasOriginsForAnEmptyColumnAndForTheShapelessEntryPoint(t *testing.T) {
	empty := projectWithPaint(t, parseWindowCountTemplate(t, `{"assets":{},"bands":{"content":{"elements":[]},"pageFooter":{"elements":[],"height":24},"pageHeader":{"elements":[],"height":18}},"fonts":{"body":["Roboto-Regular"]},"locale":"en","nextId":1,"page":{"margin":{"bottom":42,"left":36,"right":54,"top":30},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"1.0"}`))
	assertWindowOriginsAreWellFormed(t, "empty column", empty)
	if len(empty.ContentWindowOrigins) != 1 || empty.ContentWindowCountIsFloor {
		t.Fatalf("empty column: origins %v, floor %v; want [0] and false — nothing about an empty column is a floor", empty.ContentWindowOrigins, empty.ContentWindowCountIsFloor)
	}

	// Canvas has no FontSet, cannot shape, and says so in both fields: one
	// window beginning at zero, DECLARED a floor. It never reaches the
	// browser, but the struct is shared and its values must be honest.
	bare, err := Canvas(componentTemplate(t))
	if err != nil {
		t.Fatal(err)
	}
	assertWindowOriginsAreWellFormed(t, "Canvas", bare)
	if len(bare.ContentWindowOrigins) != 1 || !bare.ContentWindowCountIsFloor {
		t.Fatalf("Canvas: origins %v, floor %v; want [0] and true", bare.ContentWindowOrigins, bare.ContentWindowCountIsFloor)
	}
}

// sheetOf answers which drawn sheet a column offset lands on, the way the
// designer's own model does: the last window that begins at or above it.
func sheetOf(origins []int64, y int64) int {
	sheet := 0
	for i := 1; i < len(origins); i++ {
		if origins[i] <= y {
			sheet = i
		}
	}
	return sheet
}

// TestAComponentAuthoredWindowsDownTheColumnLandsOnItsOwnSheet closes Story
// 7.6's loop END TO END rather than asserting that a conditional changed: a
// component is CREATED windows below the top of the column through the same
// band-aware command the designer's later-sheet placement sends, the template
// is serialized to its canonical bytes and parsed back, and the projection
// taken from those bytes is asked which sheet the component is on. Then it is
// MOVED further down and asked again.
//
// It is deliberately built from the command surface and the canonical bytes,
// not from a fixture literal: what the story claims is that an author can put
// something on sheet three and have it stay there.
func TestAComponentAuthoredWindowsDownTheColumnLandsOnItsOwnSheet(t *testing.T) {
	// The control fixture is the base because its font chain is one
	// testFontSet supplies: this test is about where a component LANDS, and a
	// document the projection cannot shape would never reach the question.
	tpl := parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON)
	bands := projectedBands(t, tpl)
	content := bands["content"]
	before, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	deep := content.Height*2 + 5000
	created, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":`+pointLiteral(deep)+`,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatalf("createComponent two windows down the column was refused: %v", err)
	}
	component := newProjectedComponent(t, before, created)

	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := ParseTemplate(canonical)
	if err != nil {
		t.Fatalf("the canonical bytes did not parse back: %v", err)
	}
	projection := projectWithPaint(t, reloaded)
	assertWindowOriginsAreWellFormed(t, "authored deep", projection)
	if projection.ContentWindowCountIsFloor {
		t.Fatalf("this document has no table, no degradation and no unshaped text; floor = %v", projection.ContentWindowCountIsFloor)
	}
	sheet := sheetOf(projection.ContentWindowOrigins, component.Y)
	if sheet == 0 {
		t.Fatalf("a component at column offset %d landed on sheet one of %d; origins %v", component.Y, projection.ContentWindowCount, projection.ContentWindowOrigins)
	}
	// The window it landed in really does contain it — the sheet the canvas
	// draws it on shows it, rather than merely being the last one that starts
	// above it.
	origin := projection.ContentWindowOrigins[sheet]
	if component.Y < origin || component.Y >= origin+projection.ContentWindowHeight {
		t.Fatalf("component y %d is not inside window %d, which spans [%d, %d)", component.Y, sheet, origin, origin+projection.ContentWindowHeight)
	}

	// AND FURTHER DOWN, through the ordinary opaque move the drag commits —
	// a COLUMN coordinate, not a pin to a sheet.
	deeper := content.Height*5 + 5000
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+component.ID+`","x":0,"y":`+pointLiteral(deeper)+`,"snap":false}`)); err != nil {
		t.Fatalf("a move five windows down the column was refused: %v", err)
	}
	movedBytes, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	movedTemplate, err := ParseTemplate(movedBytes)
	if err != nil {
		t.Fatal(err)
	}
	moved := projectWithPaint(t, movedTemplate)
	assertWindowOriginsAreWellFormed(t, "moved deeper", moved)
	if landed := sheetOf(moved.ContentWindowOrigins, deeper); landed <= sheet {
		t.Fatalf("moving from column offset %d to %d left the component on sheet %d, no later than sheet %d; origins %v", component.Y, deeper, landed, sheet, moved.ContentWindowOrigins)
	}
	if reloadedComponent(t, tpl, component.ID).Y != deeper {
		t.Fatalf("the canonical bytes did not carry the column coordinate %d", deeper)
	}
}
