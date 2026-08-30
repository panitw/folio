package folio

import (
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// alignedTemplate is one band holding four text elements that differ ONLY in
// their committed alignment: same value, same face, same size, same box. Any
// difference in where they are drawn is therefore the alignment rule and
// nothing else.
const alignedTemplateJSON = `{"version":"1.0","page":{"size":"A4","orientation":"portrait","margin":{"top":36,"right":36,"bottom":36,"left":36}},"bands":{"pageHeader":{"height":20,"elements":[]},"content":{"elements":[
{"id":"e1","type":"text","x":10,"y":0,"width":200,"height":40,"value":"Total","style":{"fontFamily":"body","fontSize":12}},
{"id":"e2","type":"text","x":10,"y":50,"width":200,"height":40,"value":"Total","style":{"fontFamily":"body","fontSize":12,"align":"center"}},
{"id":"e3","type":"text","x":10,"y":100,"width":200,"height":40,"value":"Total","style":{"fontFamily":"body","fontSize":12,"align":"right"}},
{"id":"e4","type":"text","x":10,"y":150,"width":200,"height":40,"value":"Total","style":{"fontFamily":"body","fontSize":12,"valign":"bottom"}},
{"id":"e5","type":"text","x":10,"y":200,"width":200,"height":40,"value":"Total","style":{"fontFamily":"body","fontSize":12,"valign":"middle"}}
]},"pageFooter":{"height":20,"elements":[]}},"fonts":{"body":["Roboto-Regular"]},"locale":"en","utcOffset":"+00:00","assets":{},"nextId":6}`

func TestTextAlignmentDistributesSlackOnly(t *testing.T) {
	for _, c := range []struct {
		name                  string
		align                 string
		box, line, wantOffset geom.Length
	}{
		{"unset is the start edge", "", 200_000, 50_000, 0},
		{"left is the start edge", "left", 200_000, 50_000, 0},
		{"right takes all the slack", "right", 200_000, 50_000, 150_000},
		{"center halves the slack", "center", 200_000, 50_000, 75_000},
		{"an exact half rounds to even, down", "center", 200_000, 199_999, 0},
		{"an exact half rounds to even, up", "center", 200_000, 199_997, 2},
		{"a line that exactly fills has no slack", "center", 200_000, 200_000, 0},
		{"an overflowing line keeps the start edge", "right", 200_000, 260_000, 0},
		{"an undeclared width has no box to align in", "right", 0, 50_000, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := textAlignOffset(c.align, c.box, c.line); got != c.wantOffset {
				t.Errorf("textAlignOffset(%q, %d, %d) = %d, want %d", c.align, c.box, c.line, got, c.wantOffset)
			}
		})
	}
	for _, c := range []struct {
		name                   string
		valign                 string
		box, block, wantOffset geom.Length
	}{
		{"unset is the top edge", "", 40_000, 14_000, 0},
		{"top is the top edge", "top", 40_000, 14_000, 0},
		{"bottom takes all the slack", "bottom", 40_000, 14_000, 26_000},
		{"middle halves the slack", "middle", 40_000, 14_000, 13_000},
		{"an exact half rounds to even, down", "middle", 40_000, 39_999, 0},
		{"an exact half rounds to even, up", "middle", 40_000, 39_997, 2},
		{"a block taller than its box keeps the top edge", "middle", 40_000, 60_000, 0},
		{"an undeclared height has no box to align in", "bottom", 0, 14_000, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := textValignOffset(c.valign, c.box, c.block); got != c.wantOffset {
				t.Errorf("textValignOffset(%q, %d, %d) = %d, want %d", c.valign, c.box, c.block, got, c.wantOffset)
			}
		})
	}
}

// TestTextBlockHeightIsTheRuledModelsOwnExtent pins the one derivation valign
// depends on: the spans the packer and the page-splitter already use, never a
// re-measurement of the text.
func TestTextBlockHeightIsTheRuledModelsOwnExtent(t *testing.T) {
	vm := verticalMetrics{FirstBaseline: 11_000, Advance: 14_000, LastDescent: 3_000}
	for _, c := range []struct {
		lines int
		want  geom.Length
	}{{0, 0}, {1, 14_000}, {2, 28_000}, {3, 42_000}} {
		if got := textBlockHeight(c.lines, vm); got != c.want {
			t.Errorf("textBlockHeight(%d) = %d, want %d", c.lines, got, c.want)
		}
	}
}

// TestAlignedTextElementsMoveInsideTheirDeclaredBox is the production-path
// assertion: the same string, in the same box, drawn at three horizontal and
// three vertical alignments, lands where the slack rule says it lands.
func TestAlignedTextElementsMoveInsideTheirDeclaredBox(t *testing.T) {
	tpl, err := ParseTemplate([]byte(alignedTemplateJSON))
	if err != nil {
		t.Fatal(err)
	}
	data := emptyBindValue(t)
	runs, err := collectTextRuns(tpl, data, data, testFontSet(), newFontCache())
	if err != nil {
		t.Fatal(err)
	}
	byElement := make(map[string]textRunSource, len(runs))
	for _, run := range runs {
		if _, seen := byElement[run.elementID]; seen {
			t.Fatalf("element %s produced more than one run; this fixture is meant to be one line, one face", run.elementID)
		}
		byElement[run.elementID] = run
	}
	if len(byElement) != 5 {
		t.Fatalf("presence precondition: got runs for %d elements, want 5", len(byElement))
	}
	width, err := shippingRunWidth(byElement["e1"], testFontSet(), newFontCache())
	if err != nil {
		t.Fatal(err)
	}
	const box = geom.Length(200_000)
	slack := box - width
	if slack <= 0 {
		t.Fatalf("presence precondition: %q measures %d in a %d box, leaving no slack to distribute", "Total", width, box)
	}
	left := byElement["e1"].x
	if got, want := byElement["e2"].x, left+geom.ScaleRound(slack, 1, 2); got != want {
		t.Errorf("centred x = %d, want %d", got, want)
	}
	if got, want := byElement["e3"].x, left+slack; got != want {
		t.Errorf("right-aligned x = %d, want %d", got, want)
	}

	// Vertical: e1/e4/e5 sit at declared y of 0/150/200 in the same band, so
	// each one's own top is its declared origin plus its valign offset.
	vm, err := chainVerticalModel([]string{"Roboto-Regular"}, 12_000, testFontSet(), newFontCache())
	if err != nil {
		t.Fatal(err)
	}
	block := textBlockHeight(1, vm)
	const boxHeight = geom.Length(40_000)
	if got, want := byElement["e4"].y, byElement["e1"].y+150_000+(boxHeight-block); got != want {
		t.Errorf("bottom-aligned y = %d, want %d", got, want)
	}
	if got, want := byElement["e5"].y, byElement["e1"].y+200_000+geom.ScaleRound(boxHeight-block, 1, 2); got != want {
		t.Errorf("middle-aligned y = %d, want %d", got, want)
	}
}

// TestCanvasPaintMatchesTheShippingRunPathUnderAlignment is the parity claim
// that matters for the designer: the canvas draws an aligned element exactly
// where the PDF producer draws it, from the same committed style, with no
// second alignment rule in the browser.
func TestCanvasPaintMatchesTheShippingRunPathUnderAlignment(t *testing.T) {
	tpl, err := ParseTemplate([]byte(alignedTemplateJSON))
	if err != nil {
		t.Fatal(err)
	}
	projection, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	data := emptyBindValue(t)
	runs, err := collectTextRuns(tpl, data, data, testFontSet(), newFontCache())
	if err != nil {
		t.Fatal(err)
	}
	shipping := make(map[string]textRunSource, len(runs))
	for _, run := range runs {
		shipping[run.elementID] = run
	}
	bandY := make(map[string]int64)
	for _, band := range projection.Bands {
		bandY[band.Name] = band.Y
	}
	checked := 0
	for _, component := range projection.Components {
		paint := component.TextPaint
		if paint == nil || len(paint.Lines) == 0 {
			t.Fatalf("component %s has no paint to compare", component.ID)
		}
		run, ok := shipping[component.ID]
		if !ok {
			t.Fatalf("component %s has paint but no shipping run", component.ID)
		}
		bandOrigin := bandY[component.Band] - projection.MarginTop
		if got, want := paint.Lines[0].Fragments[0].X, int64(run.x); got != want {
			t.Errorf("component %s canvas x = %d, want the shipping run's %d", component.ID, got, want)
		}
		if got, want := paint.Lines[0].Top, int64(run.y)-bandOrigin; got != want {
			t.Errorf("component %s canvas top = %d, want the shipping run's %d", component.ID, got, want)
		}
		checked++
	}
	if checked != 5 {
		t.Fatalf("compared %d components, want 5", checked)
	}
}

// TestCanvasProjectionCarriesTheDeclaredFamiliesAndTheDefaultSize covers the
// two values the designer's typography controls are built from: the closed
// set style.fontFamily may name in this document, and the size the producer
// uses when an element commits none.
func TestCanvasProjectionCarriesTheDeclaredFamiliesAndTheDefaultSize(t *testing.T) {
	tpl, err := ParseTemplate([]byte(`{"version":"1.0","page":{"size":"A4","orientation":"portrait","margin":{"top":36,"right":36,"bottom":36,"left":36}},"bands":{"pageHeader":{"height":20,"elements":[]},"content":{"elements":[]},"pageFooter":{"height":20,"elements":[]}},"fonts":{"heading":["Roboto-Regular"],"body":["Roboto-Regular"]},"locale":"en","utcOffset":"+00:00","assets":{},"nextId":1}`))
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.FontFamilies) != 2 || projection.FontFamilies[0] != "body" || projection.FontFamilies[1] != "heading" {
		t.Errorf("font families = %#v, want the declared chains, sorted", projection.FontFamilies)
	}
	for _, name := range projection.FontFamilies {
		if !knownFontFamily(tpl, name) {
			t.Errorf("projected family %q is not one a fontFamily command would accept", name)
		}
	}
	if projection.DefaultFontSize != int64(defaultFontSizePt) {
		t.Errorf("default font size = %d, want the producer's own %d", projection.DefaultFontSize, defaultFontSizePt)
	}
}
