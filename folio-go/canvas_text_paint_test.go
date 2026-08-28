package folio

import (
	"bytes"
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
)

func TestCanvasTextPaintUsesPackedProductionLineGeometryWithoutMutation(t *testing.T) {
	tpl := parseFontTestTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Components) == 0 || projection.Components[0].TextPaint == nil {
		t.Fatalf("missing text paint projection: %#v", projection.Components)
	}
	paint := projection.Components[0].TextPaint
	if len(paint.Lines) != 1 || paint.Lines[0].Width <= 0 || paint.Lines[0].Baseline < paint.Lines[0].Top || len(paint.Lines[0].Fragments) == 0 {
		t.Fatalf("invalid text paint line: %#v", paint.Lines)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("text paint mutated canonical template: %v", err)
	}
}

func TestCanvasTextPaintCarriesEngineOverflowInsteadOfRebreaking(t *testing.T) {
	tpl, err := ParseTemplate([]byte(`{"version":"1.0","page":{"size":"A4","orientation":"portrait","margin":{"top":36,"right":36,"bottom":36,"left":36}},"bands":{"pageHeader":{"height":20,"elements":[]},"content":{"elements":[{"id":"e1","type":"text","x":0,"y":0,"width":1,"height":20,"value":"unbreakable","style":{"fontFamily":"body","fontSize":12}}]},"pageFooter":{"height":20,"elements":[]}},"fonts":{"body":["Roboto-Regular"]},"locale":"en","utcOffset":"+00:00","assets":{},"nextId":2}`))
	if err != nil {
		t.Fatal(err)
	}
	projection, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	paint := projection.Components[0].TextPaint
	if paint == nil || !paint.Overflow || len(paint.Lines) != 1 || paint.Lines[0].Width <= projection.Components[0].Width {
		t.Fatalf("overflow paint = %#v", paint)
	}
}

func TestCanvasTextPaintExactlyMatchesTheShippingRunPath(t *testing.T) {
	tpl := parseFontTestTemplate(t)
	projection, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	data := emptyBindValue(t)
	runs, err := collectTextRuns(tpl, data, data, testFontSet(), newFontCache())
	if err != nil {
		t.Fatalf("shipping run collection: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("presence precondition: shipping path produced no text runs")
	}
	byElement := make(map[string][]textRunSource)
	for _, run := range runs {
		byElement[run.elementID] = append(byElement[run.elementID], run)
	}
	bandY := make(map[string]int64)
	for _, band := range projection.Bands {
		bandY[band.Name] = band.Y
	}
	shippingFonts := newFontCache()
	for _, component := range projection.Components {
		if component.Type != "text" || component.TextPaint == nil {
			continue
		}
		paint := component.TextPaint
		wantRuns := byElement[component.ID]
		if len(wantRuns) == 0 {
			t.Fatalf("component %s has paint but no shipping runs", component.ID)
		}
		for lineIndex, line := range paint.Lines {
			var lineRuns []textRunSource
			for _, run := range wantRuns {
				if run.lineIndex == lineIndex {
					lineRuns = append(lineRuns, run)
				}
			}
			if len(lineRuns) != len(line.Fragments) {
				t.Fatalf("component %s line %d fragments = %d, want shipping run count %d", component.ID, lineIndex, len(line.Fragments), len(lineRuns))
			}
			var shippingWidth geom.Length
			for fragmentIndex, run := range lineRuns {
				runWidth, err := shippingRunWidth(run, testFontSet(), shippingFonts)
				if err != nil {
					t.Fatal(err)
				}
				shippingWidth += runWidth
				bandOrigin := bandY[component.Band] - projection.MarginTop
				if line.Top != int64(run.y)-bandOrigin || line.Baseline != int64(run.y+run.baselineOffset)-bandOrigin {
					t.Errorf("component %s line %d origin/baseline diverges from shipping run: %#v versus y=%d baseline=%d", component.ID, lineIndex, line, run.y, run.y+run.baselineOffset)
				}
				fragment := line.Fragments[fragmentIndex]
				if fragment.Text != run.text || fragment.X != int64(run.x) {
					t.Errorf("component %s line %d fragment %d = %#v, want shipping run text=%q x=%d", component.ID, lineIndex, fragmentIndex, fragment, run.text, run.x)
				}
			}
			if line.Width != int64(shippingWidth) {
				t.Errorf("component %s line %d width = %d, want shipping shaped width %d", component.ID, lineIndex, line.Width, shippingWidth)
			}
		}
	}
}

func shippingRunWidth(run textRunSource, fs FontSet, cache *fontCache) (geom.Length, error) {
	face, err := cache.get(run.face, fs)
	if err != nil {
		return 0, err
	}
	var advance1000 geom.Length
	for _, glyph := range run.glyphs {
		advance1000 += geom.ScaleRound(geom.Length(glyph.XAdvance), 1000, int64(face.UnitsPerEm()))
	}
	return geom.ScaleRound(advance1000, int64(run.fontSize), 1000), nil
}

func TestCanvasTextPaintRejectsDerivedCoordinatesOutsideTheJSRange(t *testing.T) {
	if _, err := canvasLineTop(geom.Length(MaxCanvasMillipoints), 1, 1); err == nil {
		t.Fatal("overflowing later-line origin was accepted")
	}
	if _, err := canvasDerivedSum(geom.Length(MaxCanvasMillipoints), 1); err == nil {
		t.Fatal("overflowing baseline was accepted")
	}
}
