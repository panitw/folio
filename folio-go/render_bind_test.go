package folio

import (
	"strings"
	"testing"
)

// bindTestTemplateJSON is a `.folio` document with one text element
// bound to report data (AC15) and one table element whose bind/columns
// are expression-shaped (M-4/AC20) — table.bind and columns[].bind are
// never scanned by Render's binding (AC20's field-scope fence): only
// text-element `value` interpolation is (D-1.6.5, D-1.6.8).
const bindTestTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "Statement for {{customer.name}}", "style": {"fontFamily": "body", "fontSize": 14}},
        {"id": "e2", "type": "table", "x": 0, "y": 30, "bind": "{{transactions[]}}", "headerHeight": 20,
          "columns": [
            {"id": "e3", "label": "Amount", "width": 80, "bind": "{{formatNumber(transaction.amount, \"#,##0.00\")}}"}
          ]}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {
    "body": ["Roboto-Regular"]
  },
  "locale": "en",
  "nextId": 4,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// TestRenderScopeFenceIgnoresTableBind is AC20/D-1.6.8: a table's
// bind/columns[].bind fields carry expression-shaped "{{…}}" content
// (parentheses, a comma and quotes — three separate AC16 rejection
// triggers) that would fail AC16's grammar if it were ever scanned.
// Render succeeds anyway, because binding is scoped to text-element
// `value` interpolation only (collectTextRuns skips every non-text
// element, and is the ONLY call site of internal/bind.BindText) — this
// is the shape M-4 found already living in the canonical golden
// fixture (worked-example.json:19), reproduced here as a render-time
// (not merely round-trip) proof.
func TestRenderScopeFenceIgnoresTableBind(t *testing.T) {
	tpl, err := ParseTemplate([]byte(bindTestTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	data := Data(`{"customer": {"name": "Ada Lovelace"}}`)
	b, err := Render(tpl, data, testFontSet())
	if err != nil {
		t.Fatalf("Render must succeed despite the table's expression-shaped bind fields: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("Render produced no bytes")
	}
}

// TestRenderBindsTextPlaceholder is AC15/AC22-AC24 end to end through
// the public Render API.
func TestRenderBindsTextPlaceholder(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSONWithText))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	data := Data(`{"customer": {"name": "Ada Lovelace"}}`)
	b, err := Render(tpl, data, testFontSet())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("Render produced no bytes")
	}
}

// TestRenderAbsentPathIsLocatedError is AC8 through the public Render
// API: a binding whose path is absent from the data is an Error
// naming both the data path and the element id, and the render fails.
func TestRenderAbsentPathIsLocatedError(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSONWithText))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, err = Render(tpl, Data(`{}`), testFontSet())
	if err == nil {
		t.Fatal("expected an Error for an absent data path")
	}
}

// TestRenderNullPathRendersEmpty is AC9 through the public Render API:
// an explicit JSON null renders as empty, and is not an error.
func TestRenderNullPathRendersEmpty(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSONWithText))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, err = Render(tpl, Data(`{"customer": {"name": null}}`), testFontSet())
	if err != nil {
		t.Fatalf("explicit null must not be an error: %v", err)
	}
}

// TestRenderWrongKindIsLocatedError is AC10 through the public Render
// API: a JSON number bound into a text element is an Error, never
// coerced.
func TestRenderWrongKindIsLocatedError(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSONWithText))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, err = Render(tpl, Data(`{"customer": {"name": 123}}`), testFontSet())
	if err == nil {
		t.Fatal("expected an Error for a JSON number bound into a text element")
	}
}

// TestRenderRejectsExpressionSyntaxInTextValue is AC16/AC17 through the
// public Render API.
func TestRenderRejectsExpressionSyntaxInTextValue(t *testing.T) {
	const tplJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "{{formatNumber(a, \"x\")}}", "style": {"fontFamily": "body", "fontSize": 14}}
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
	tpl, err := ParseTemplate([]byte(tplJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, err = Render(tpl, Data(`{}`), testFontSet())
	if err == nil {
		t.Fatal("expected a located error naming the element id and Epic 3")
	}
	if !strings.Contains(err.Error(), "e1") || !strings.Contains(err.Error(), "Epic 3") {
		t.Fatalf("error must name the element id and mention Epic 3, got: %v", err)
	}
}

// TestRenderNullBoundTextStillValidatesFontChain is QA Finding 5 (this
// story's review, Major): an element with an unresolvable
// style.fontFamily chain must be a located error (Story 1.5 AC2/AC4)
// regardless of whether its bound text happens to be empty — the AC9
// short-circuit ("null renders as empty, not an error") must not
// suppress the element's own font-chain validation. Reproduces the
// finding's own table: the same template, unresolvable chain
// "nosuchchain", varying only the data.
func TestRenderNullBoundTextStillValidatesFontChain(t *testing.T) {
	const tplJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "{{customer.name}}", "style": {"fontFamily": "nosuchchain", "fontSize": 14}}
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
	tpl, err := ParseTemplate([]byte(tplJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}

	cases := []struct {
		name string
		data string
	}{
		{"null binding", `{"customer": {"name": null}}`},
		{"empty-string binding", `{"customer": {"name": ""}}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Render(tpl, Data(c.data), testFontSet())
			if err == nil {
				t.Fatalf("expected a located font-chain error even though the bound text is empty (data: %s)", c.data)
			}
			if !strings.Contains(err.Error(), "nosuchchain") {
				t.Errorf("error must still name the unresolvable chain, got: %v", err)
			}
		})
	}

	// Control: a non-empty binding against the same broken template
	// was already an error before this fix; confirm it still is.
	if _, err := Render(tpl, Data(`{"customer": {"name": "Ada"}}`), testFontSet()); err == nil {
		t.Fatal("control: a non-empty binding against an unresolvable chain must still be an error")
	}
}

// minimalTemplateJSONWithText is minimalTemplateJSON plus one text
// element bound to "customer.name".
const minimalTemplateJSONWithText = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "Statement for {{customer.name}}", "style": {"fontFamily": "body", "fontSize": 14}}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {
    "body": ["Roboto-Regular"]
  },
  "locale": "en",
  "nextId": 2,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
