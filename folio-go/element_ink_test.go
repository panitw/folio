// Story 10.1's own suite: style.color — the ink a text-bearing element
// prints in. The format carried no such field before this story; text
// took the PDF's own initial fill and there was nothing to set.
package folio

import (
	"errors"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/pagemodel"
)

func inkTemplateJSON(styleFields string) string {
	return `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 30, "value": "Inked", "style": {` + styleFields + `"fontFamily": "body", "fontSize": 12}},
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

func TestStyleColorInksEveryRunOfItsElementAndNoOther(t *testing.T) {
	pages := boxPages(t, inkTemplateJSON(`"color": "#c81e1e", `))
	inked, plain := 0, 0
	for _, run := range pages[0].Runs {
		switch {
		case run.HasColor:
			inked++
			if run.Color != (pagemodel.Color{R: 0xc8, G: 0x1e, B: 0x1e}) {
				t.Errorf("inked run carries %+v, want #c81e1e", run.Color)
			}
		default:
			plain++
		}
	}
	if inked == 0 || plain == 0 {
		t.Fatalf("runs: %d inked, %d plain — want e1's runs inked and e2's left alone", inked, plain)
	}
}

// An element that declares no colour must carry NO colour, not black:
// the difference is a colour operator in the content stream, and it is
// what keeps every document written before this field byte-identical.
func TestAnUndeclaredColorIsAbsentRatherThanBlack(t *testing.T) {
	for _, run := range boxPages(t, inkTemplateJSON(``))[0].Runs {
		if run.HasColor {
			t.Fatalf("a run carries a colour (%+v) on a document that declares none", run.Color)
		}
	}
}

func TestANullColorIsNoColor(t *testing.T) {
	for _, run := range boxPages(t, inkTemplateJSON(`"color": null, `))[0].Runs {
		if run.HasColor {
			t.Errorf("an explicitly null color inked a run %+v — present-null means no declaration", run.Color)
		}
	}
}

func TestAMalformedColorIsALocatedRenderError(t *testing.T) {
	tpl, err := ParseTemplate([]byte(inkTemplateJSON(`"color": "red", `)))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, _, _, _, err = buildPageModel(tpl, mustDecodeData(t, `{"customer":{"flag":false}}`), mustDecodeParams(t), testFontSet())
	if err == nil {
		t.Fatal("a non-#RRGGBB style.color rendered without error")
	}
	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("error is %T (%v), want a located *RenderError", err, err)
	}
	if re.Diagnostic.ElementID != "e1" || re.Diagnostic.Code != DiagCodeStyleColorInvalid {
		t.Errorf("error = {element:%q code:%q}, want {e1 %s}", re.Diagnostic.ElementID, re.Diagnostic.Code, DiagCodeStyleColorInvalid)
	}
}

// style.color is a string-valued style field, so Story 3.5's fence
// covers it by construction: no colour-by-data, in any style field.
func TestAPlaceholderInStyleColorIsALoadError(t *testing.T) {
	_, err := ParseTemplate([]byte(inkTemplateJSON(`"color": "{{customer.brand}}", `)))
	if err == nil {
		t.Fatal("a {{ }} placeholder in style.color loaded successfully")
	}
	if !strings.Contains(err.Error(), "style.color") {
		t.Errorf("error %q does not name style.color", err)
	}
}

// A table's cells take the table's own cascade, exactly as their
// background and padding already do.
func TestATableCascadesItsInkToHeaderAndCells(t *testing.T) {
	doc := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "rows[]", "as": "row", "headerHeight": 20,
          "style": {"color": "#1b2a4a", "fontFamily": "body", "fontSize": 10},
          "headerStyle": {"color": "#c81e1e"},
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
	header, body := 0, 0
	for _, run := range pages[0].Runs {
		if !run.HasColor {
			t.Fatalf("a table run carries no ink: %q", run.SourceText)
		}
		switch run.Color {
		case pagemodel.Color{R: 0xc8, G: 0x1e, B: 0x1e}:
			header++
		case pagemodel.Color{R: 0x1b, G: 0x2a, B: 0x4a}:
			body++
		default:
			t.Errorf("run %q carries %+v, which is neither the header's nor the body's ink", run.SourceText, run.Color)
		}
	}
	if header == 0 || body == 0 {
		t.Errorf("header runs = %d, body runs = %d — headerStyle.color must win for the header and style.color for the cells", header, body)
	}
}
