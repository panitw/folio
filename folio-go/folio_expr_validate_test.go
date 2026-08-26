package folio

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/template"
)

// footerDocJSON builds a minimal single-column table document whose
// column requests a "sum" footer with footerOf OMITTED, so ParseTemplate
// must attempt D-1.4.1's derivation. bindExpr is the column's raw
// "bind" value (already including the surrounding "{{ }}"), alias is
// the table's declared row alias ("as"; omitted when empty, defaulting
// to "row" per D-3.1.1).
func footerDocJSON(alias, bindExpr string) string {
	asField := ""
	if alias != "" {
		asField = `"as": "` + alias + `", `
	}
	return `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e2", "type": "table", "x": 0, "y": 0, ` + asField + `"bind": "transactions[]", "headerHeight": 14,
          "columns": [
            {"id": "e3", "label": "Amount", "width": 80, "bind": "` + bindExpr + `", "footer": "sum"}
          ]}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
}

// TestParseTemplateDerivesFooterOfShape1 is AC21's derivable arm,
// shape 1: a bare row-scoped path. Asserts the RESOLVED VALUE, not
// merely that the document loaded (D-000.59's own anti-vacuity
// clause).
func TestParseTemplateDerivesFooterOfShape1(t *testing.T) {
	tpl, err := ParseTemplate([]byte(footerDocJSON("row", `{{row.amount}}`)))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	got, ok := tpl.derivedFooters["e3"]
	if !ok {
		t.Fatal("expected a derived footerOf for column e3")
	}
	if got.FooterOf != "transactions.amount" {
		t.Errorf("FooterOf = %q, want transactions.amount", got.FooterOf)
	}
	if got.HasFooterFormat {
		t.Errorf("shape 1 must not derive a footerFormat, got %+v", got)
	}
}

// TestParseTemplateDerivesFooterOfShape2 is AC21's derivable arm,
// shape 2, using the canonical golden's own alias/pattern (D-1.4.10):
// footerFormat also resolves to the derived default.
func TestParseTemplateDerivesFooterOfShape2(t *testing.T) {
	tpl, err := ParseTemplate([]byte(footerDocJSON("transaction", `{{formatNumber(transaction.amount, \"#,##0.00\")}}`)))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	got, ok := tpl.derivedFooters["e3"]
	if !ok {
		t.Fatal("expected a derived footerOf for column e3")
	}
	if got.FooterOf != "transactions.amount" {
		t.Errorf("FooterOf = %q, want transactions.amount", got.FooterOf)
	}
	if !got.HasFooterFormat || got.FooterFormat != "#,##0.00" {
		t.Errorf("expected footerFormat to default to #,##0.00, got %+v", got)
	}
}

// TestParseTemplateRejectsNonDerivableFooterBind is AC21's rejection
// arm: a footer requested with footerOf omitted, over a bind that is
// neither derivable shape, is a load error naming the column id.
func TestParseTemplateRejectsNonDerivableFooterBind(t *testing.T) {
	_, err := ParseTemplate([]byte(footerDocJSON("row", `{{if(row.x, row.a, row.b)}}`)))
	if err == nil {
		t.Fatal("expected a load error: bind is not one of D-1.4.1's two derivable shapes")
	}
	if !strings.Contains(err.Error(), "e3") {
		t.Errorf("error must name the column id, got: %v", err)
	}
}

// TestParseTemplateFooterOfDerivationNeverWritesBack is R2/D-1.4.3's
// P3 fixed point: Serialize(Parse(b)) == b must still hold, byte for
// byte, for a document whose footer derives footerOf rather than
// declaring it — the derived value is resolved ALONGSIDE the document,
// never written into it.
//
// QA Finding 2 (Blocker): the earlier version of this test compared
// SerializeDocument(doc) against itself, and separately inspected two
// documents ParseTemplate's derivation never touched — a fresh
// template.ParseDocument(original) before the derivation ran, and
// another fresh template.ParseDocument(original) after. A write-back
// inside validateTableColumns lands in the *Document ParseTemplate
// itself parsed and walked (validateAndDeriveExpressions receives that
// exact pointer, and TableExt.Columns is a slice — a value receiver
// still shares its backing array), which neither comparison could ever
// see. Fixed by serializing tpl.doc — the actual document the
// derivation walked over — and comparing it against an independently
// derived canonical form.
func TestParseTemplateFooterOfDerivationNeverWritesBack(t *testing.T) {
	original := []byte(footerDocJSON("row", `{{row.amount}}`))

	// canonical is P3's fixed point, established independently of the
	// derivation path: an ordinary parse+serialize of original, which
	// never calls expr.DeriveFooterOf and so can never write anything
	// back. This also pins the fixture's own starting shape.
	canonicalDoc, err := template.ParseDocument(original)
	if err != nil {
		t.Fatalf("template.ParseDocument: %v", err)
	}
	canonical, err := template.SerializeDocument(canonicalDoc)
	if err != nil {
		t.Fatalf("SerializeDocument (canonical): %v", err)
	}
	if strings.Contains(string(canonical), `"footerOf"`) {
		t.Fatalf("fixture must omit footerOf by construction, got:\n%s", canonical)
	}

	// tpl.doc is the ACTUAL *template.Document ParseTemplate's
	// derivation walked over — the one place a write-back would show.
	tpl, err := ParseTemplate(original)
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	derived, err := template.SerializeDocument(tpl.doc)
	if err != nil {
		t.Fatalf("SerializeDocument (tpl.doc): %v", err)
	}
	if string(derived) != string(canonical) {
		t.Fatalf("Serialize(ParseTemplate(b).doc) != Serialize(Parse(b)) — P3 violated by a write-back (R2):\nderived:   %s\ncanonical: %s", derived, canonical)
	}
	if strings.Contains(string(derived), `"footerOf"`) {
		t.Fatalf("derivation must NEVER be written back into the document (R2): serialized output contains \"footerOf\":\n%s", derived)
	}
}

// TestParseTemplateRejectsSyntaxErrorInTextValue is AC19 through the
// public API: a syntax error inside "{{ }}" fails at LOAD (ParseTemplate),
// naming the element id and the offending text.
func TestParseTemplateRejectsSyntaxErrorInTextValue(t *testing.T) {
	const tplJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "{{a + b}}", "style": {"fontFamily": "body", "fontSize": 14}}
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
	_, err := ParseTemplate([]byte(tplJSON))
	if err == nil {
		t.Fatal("expected a load error for a syntax-invalid expression")
	}
	if !strings.Contains(err.Error(), "e1") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	// QA Finding 14 (Minor): AC19 requires the offending expression
	// text VERBATIM, not merely "an error occurred somewhere in e1" —
	// this is the half of the epic's third Given that makes the
	// diagnostic useful to a template author.
	if !strings.Contains(err.Error(), "a + b") {
		t.Errorf("error must carry the offending expression text %q, got: %v", "a + b", err)
	}
}

// TestParseTemplateRejectsSyntaxErrorInVisibleIf is QA Finding 12
// (Minor) through the public API: visibleIf is a BARE expression (no
// "{{ }}" — folio-format.md's field table, and fixtures_test.go:298's
// "visibleIf": "customer.hasTransactions"), never walked by
// validateAndDeriveExpressions before this fix, so a malformed
// visibleIf loaded clean. It must now fail at LOAD, naming the element
// id, exactly like a malformed text value.
func TestParseTemplateRejectsSyntaxErrorInVisibleIf(t *testing.T) {
	const tplJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "static text", "visibleIf": "a + b", "style": {"fontFamily": "body", "fontSize": 14}}
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
	_, err := ParseTemplate([]byte(tplJSON))
	if err == nil {
		t.Fatal("expected a load error: visibleIf's grammar has no operators, same as any other expression")
	}
	if !strings.Contains(err.Error(), "e1") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	if !strings.Contains(err.Error(), "visibleIf") {
		t.Errorf("error should identify visibleIf as the offending field, got: %v", err)
	}
}

// TestParseTemplateAcceptsBareVisibleIfExpression confirms the
// positive case: an ordinary bare-path visibleIf, exactly the shape
// fixtures_test.go's own golden already uses, must still load.
func TestParseTemplateAcceptsBareVisibleIfExpression(t *testing.T) {
	const tplJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "static text", "visibleIf": "customer.hasTransactions", "style": {"fontFamily": "body", "fontSize": 14}}
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
	if _, err := ParseTemplate([]byte(tplJSON)); err != nil {
		t.Fatalf("expected the bare-path visibleIf to load, got: %v", err)
	}
}

// TestParseTemplateRejectsOversizedNumberLiteralAtLoad is R3/QA
// Finding 7 (Major) through the public API: a numeric literal whose
// bounds NewDecimal would reject must fail at LOAD (ParseTemplate),
// not survive to Render. Before this fix, this exact document loaded
// with no error at all — the failure mode R3/F3 were written to keep
// apart.
func TestParseTemplateRejectsOversizedNumberLiteralAtLoad(t *testing.T) {
	const tplJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "{{upper(12345678901234567890123456789)}}", "style": {"fontFamily": "body", "fontSize": 14}}
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
	_, err := ParseTemplate([]byte(tplJSON))
	if err == nil {
		t.Fatal("expected a load error: a 29-digit coefficient exceeds NewDecimal's own bound and must not survive to render")
	}
	if !strings.Contains(err.Error(), "e1") {
		t.Errorf("error must name the element id, got: %v", err)
	}
}

// TestParseTemplateRejectsWrongArityAtLoad is AC10 through the public
// API: sum(a, b) is a located parse error at LOAD, not an
// unimplemented-function error at evaluation.
func TestParseTemplateRejectsWrongArityAtLoad(t *testing.T) {
	const tplJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "{{sum(a, b)}}", "style": {"fontFamily": "body", "fontSize": 14}}
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
	_, err := ParseTemplate([]byte(tplJSON))
	if err == nil {
		t.Fatal("expected a load error: sum() takes 1 argument")
	}
	if !strings.Contains(err.Error(), "e1") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	// QA Finding 14 (Minor): the offending text, verbatim.
	if !strings.Contains(err.Error(), "sum(a, b)") {
		t.Errorf("error must carry the offending expression text %q, got: %v", "sum(a, b)", err)
	}
}

// TestParseTemplateRejectsExcessiveCallNestingAtLoad is QA Finding 3's
// (Blocker) load-entry-point half: reproduces the reviewer's real
// trigger — a single text element whose value nests ~800,000 function
// calls, ~1.6MB, an unremarkable size for a .folio file — through the
// actual public API, folio.ParseTemplate, load's real entry point.
// Before internal/expr's maxCallDepth existed this input drove
// unbounded recursion into an unrecoverable runtime stack overflow;
// now it must surface as an ordinary located load error naming the
// element id, like every other rejected form.
func TestParseTemplateRejectsExcessiveCallNestingAtLoad(t *testing.T) {
	deepValue, err := json.Marshal("{{" + strings.Repeat("a(", 800000) + "}}")
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	tplJSON := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": ` + string(deepValue) + `, "style": {"fontFamily": "body", "fontSize": 14}}
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
	_, err = ParseTemplate([]byte(tplJSON))
	if err == nil {
		t.Fatal("expected a load error: excessive call nesting must be rejected, not overflow the stack")
	}
	if !strings.Contains(err.Error(), "e1") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	if !strings.Contains(err.Error(), "nests function calls too deeply") {
		t.Errorf("error must name the depth limit specifically, got: %v", err)
	}
}

// TestParseTemplateAcceptsCanonicalGolden is F3's forcing case, proved
// directly: the canonical worked example binds a formatNumber(...)
// call (D-1.4.1 shape 2, now fully implemented as of Story 3.4) and
// must still load, deriving its footer as before.
func TestParseTemplateAcceptsCanonicalGolden(t *testing.T) {
	tpl, err := LoadTemplate("testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatalf("the canonical golden must still load under Story 3.2's load-time expression validation: %v", err)
	}
	got, ok := tpl.derivedFooters["e4"]
	if !ok {
		t.Fatal("expected the golden's own footer column (e4) to derive footerOf")
	}
	if got.FooterOf != "transactions.amount" {
		t.Errorf("FooterOf = %q, want transactions.amount", got.FooterOf)
	}
	if !got.HasFooterFormat || got.FooterFormat != "#,##0.00" {
		t.Errorf("expected footerFormat to default to #,##0.00, got %+v", got)
	}
}
