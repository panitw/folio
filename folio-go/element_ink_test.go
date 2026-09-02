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

// TestAMalformedColorIsRefusedWhateverTheDataSays is E10-1's proof: the
// colour check sat ~120 lines below the AC9 empty-text short-circuit and
// the hidden-element skip, so `style.color: "red"` was a located error when
// the element had a visible value and NO diagnostic at all in each of the
// three cases below. The same broken template passed or failed on the data
// it was handed.
//
// The fix is a HOIST — elementInk now sits beside fontChain, whose comment
// exists to say that nothing may go below that line. The fontFamily control
// runs beside each case here so "refused" is measured against the sibling
// this defect is a repeat of, rather than against an expectation.
func TestAMalformedColorIsRefusedWhateverTheDataSays(t *testing.T) {
	// Each case is the SAME element with the same broken style block,
	// differing only in what makes the render path skip it. The three are
	// the three that were MEASURED to render clean.
	//
	// Note the empty case binds to "" rather than declaring `"value": ""`:
	// a literally empty template value is dropped by a short-circuit ABOVE
	// fontChain's line, which this repair does not move — the hoist is a
	// move to fontChain's position, not past it.
	cases := map[string]string{
		"value binds to empty": `"value": "{{customer.blank}}"`,
		"value binds to null":  `"value": "{{customer.absent}}"`,
		"element hidden":       `"value": "Inked", "visibleIf": "customer.flag"`,
	}
	docFor := func(element, style string) string {
		return `{
  "assets": {},
  "bands": {
    "content": {"elements": [{"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 30, ` + element + `, "style": {` + style + `}}]},
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
	const data = `{"customer":{"flag":false,"blank":"","absent":null}}`
	renderIt := func(t *testing.T, source string) error {
		t.Helper()
		tpl, err := ParseTemplate([]byte(source))
		if err != nil {
			t.Fatalf("ParseTemplate: %v", err)
		}
		_, _, _, _, err = buildPageModel(tpl, mustDecodeData(t, data), mustDecodeParams(t), testFontSet())
		return err
	}
	for name, element := range cases {
		t.Run(name, func(t *testing.T) {
			err := renderIt(t, docFor(element, `"color": "red", "fontFamily": "body", "fontSize": 12`))
			if err == nil {
				t.Fatal("a non-#RRGGBB style.color rendered clean — the template's validity must not depend on the report it was handed")
			}
			var re *RenderError
			if !errors.As(err, &re) {
				t.Fatalf("error is %T (%v), want a located *RenderError", err, err)
			}
			if re.Diagnostic.ElementID != "e1" || re.Diagnostic.Code != DiagCodeStyleColorInvalid {
				t.Errorf("error = {element:%q code:%q}, want {e1 %s}", re.Diagnostic.ElementID, re.Diagnostic.Code, DiagCodeStyleColorInvalid)
			}

			// THE CONTROL, on the same document in the same shape: an
			// unresolvable fontFamily chain was hoisted for exactly this
			// reason and already refuses here. style.color must match it.
			if cerr := renderIt(t, docFor(element, `"fontFamily": "nosuchchain", "fontSize": 12`)); cerr == nil {
				t.Fatal("the fontFamily control rendered clean — this test's premise no longer holds")
			}

			// And the well-formed twin still renders, so "refused" above is
			// evidence about the colour rather than about the fixture.
			if okErr := renderIt(t, docFor(element, `"color": "#c81e1e", "fontFamily": "body", "fontSize": 12`)); okErr != nil {
				t.Fatalf("the valid-colour twin failed to render: %v", okErr)
			}
		})
	}
}

// cascadeTableDoc builds a one-column bound table whose own style declares
// `color: #c81e1e` and whose headerStyle is headerStyleJSON verbatim.
func cascadeTableDoc(headerStyleJSON string) string {
	return `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "rows[]", "as": "row", "headerHeight": 20,
          "style": {"color": "#c81e1e", "background": "#eeeeee", "border": {}, "fontFamily": "body", "fontSize": 10},
          ` + headerStyleJSON + `
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
}

// TestANullHeaderColorFallsThroughToTheTableColour is E10-2's proof.
//
// resolveHeaderStyle's Color arm was the ONLY one of nine reading `.Set`
// without `.Null`, so `headerStyle: {"color": null}` WON the cascade with a
// null and stopped the fall-through: measured at 2 inked runs where an
// absent or empty headerStyle gave 3, with the header printing black. The
// background and border arms on the SAME table are the controls — both
// already fall through on an explicit null — and folio-format.md already
// states the correct rule, so the code was the outlier and neither document
// changes.
func TestANullHeaderColorFallsThroughToTheTableColour(t *testing.T) {
	inkOps := func(t *testing.T, headerStyleJSON string) (inked, black int) {
		t.Helper()
		tpl, err := ParseTemplate([]byte(cascadeTableDoc(headerStyleJSON)))
		if err != nil {
			t.Fatalf("ParseTemplate: %v", err)
		}
		pages, _, _, _, err := buildPageModel(tpl, mustDecodeData(t, `{"rows":[{"a":"one"},{"a":"two"}]}`), mustDecodeParams(t), testFontSet())
		if err != nil {
			t.Fatalf("buildPageModel: %v", err)
		}
		for _, run := range pages[0].Runs {
			if run.HasColor && run.Color == (pagemodel.Color{R: 0xc8, G: 0x1e, B: 0x1e}) {
				inked++
			} else {
				black++
			}
		}
		return inked, black
	}

	absent, absentBlack := inkOps(t, ``)
	empty, emptyBlack := inkOps(t, `"headerStyle": {},`)
	null, nullBlack := inkOps(t, `"headerStyle": {"color": null},`)
	if absent != 3 || absentBlack != 0 {
		t.Fatalf("headerStyle absent gives %d inked / %d black, want 3 / 0 — the fixture moved and this test's controls no longer apply", absent, absentBlack)
	}
	if empty != absent || emptyBlack != 0 {
		t.Errorf("headerStyle {} gives %d inked / %d black, want %d / 0", empty, emptyBlack, absent)
	}
	if null != absent || nullBlack != 0 {
		t.Errorf("headerStyle {\"color\": null} gives %d inked / %d black, want %d / 0 — an explicit null must fall through to style.color, exactly as headerStyle.background and headerStyle.border already do", null, nullBlack, absent)
	}

	// The controls, on the same table and the same document: the background
	// and border arms already fall through on an explicit null, and this
	// counts their ink so "3" above is measured against siblings rather than
	// against a number.
	for _, control := range []struct{ name, headerStyle string }{
		{"background", `"headerStyle": {"background": null},`},
		{"border", `"headerStyle": {"border": null},`},
	} {
		tpl, err := ParseTemplate([]byte(cascadeTableDoc(control.headerStyle)))
		if err != nil {
			t.Fatalf("ParseTemplate: %v", err)
		}
		pages, _, _, _, err := buildPageModel(tpl, mustDecodeData(t, `{"rows":[{"a":"one"},{"a":"two"}]}`), mustDecodeParams(t), testFontSet())
		if err != nil {
			t.Fatalf("buildPageModel: %v", err)
		}
		fills, strokes := 0, 0
		for _, rect := range pages[0].Rects {
			if rect.HasFill {
				fills++
			}
			if rect.HasStroke {
				strokes++
			}
		}
		if fills != 3 || strokes != 3 {
			t.Errorf("%s control: %d fills / %d stroke groups, want 3 / 3 — an explicit null on this arm falls through to the table's own style", control.name, fills, strokes)
		}
	}
}

// TestAHeaderColourDiagnosticNamesTheArmItTook is E10-5b's proof. The error
// site emitted the literal "headerStyle.color/style.color", which names both
// fields and identifies neither: AD-14's callers match on the CODE, so
// nothing was broken — but AD-14's other half is that a diagnostic LOCATES,
// and the cascade already knew which arm it took.
//
// It asserts the LOCATED PATH, never the surrounding wording.
func TestAHeaderColourDiagnosticNamesTheArmItTook(t *testing.T) {
	for _, tc := range []struct {
		name, headerStyle, table, wantPath, wantAbsent string
	}{
		{
			name:        "headerStyle wins",
			headerStyle: `"headerStyle": {"color": "red"},`,
			wantPath:    "headerStyle.color",
			wantAbsent:  "style.color/",
		},
		{
			name:        "style is what cascaded",
			headerStyle: ``,
			table:       "bad-base",
			wantPath:    "style.color",
			wantAbsent:  "headerStyle.color",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := cascadeTableDoc(tc.headerStyle)
			if tc.table == "bad-base" {
				doc = strings.Replace(doc, `"color": "#c81e1e"`, `"color": "red"`, 1)
			}
			tpl, err := ParseTemplate([]byte(doc))
			if err != nil {
				t.Fatalf("ParseTemplate: %v", err)
			}
			_, _, _, _, err = buildPageModel(tpl, mustDecodeData(t, `{"rows":[{"a":"one"}]}`), mustDecodeParams(t), testFontSet())
			if err == nil {
				t.Fatal("a malformed cascaded header colour rendered without error")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.wantPath) {
				t.Errorf("the diagnostic does not name %q: %s", tc.wantPath, msg)
			}
			if strings.Contains(msg, tc.wantAbsent) {
				t.Errorf("the diagnostic still names the arm it did NOT take (%q): %s", tc.wantAbsent, msg)
			}
		})
	}
}
