package folio

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParameterReferencesUseOnlyCanonicalNonTableExpressions(t *testing.T) {
	input, err := os.ReadFile("testdata/example/first-pdf.folio")
	if err != nil {
		t.Fatal(err)
	}
	input = bytes.Replace(input, []byte("{{customer.name}}"), []byte("{{params.reportDate}} {{formatDate(params.reportDate, \\\"yyyy\\\")}} {{params.branch}} {{params.statement.reportDate}}"), 1)
	tpl, err := ParseTemplate(input)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParameterReferences(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"branch", "reportDate", "statement"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("references = %#v, want %#v", got, want)
	}
}

func TestParameterReferencesIncludeNestedAndTableVisibleIfButNotTableBindings(t *testing.T) {
	input, err := os.ReadFile("testdata/example/first-pdf.folio")
	if err != nil {
		t.Fatal(err)
	}
	input = bytes.Replace(input, []byte(`"nextId": 2`), []byte(`"nextId": 4`), 1)
	table := `,
        {"id": "e2", "type": "table", "x": 0, "y": 30, "bind": "transactions[]", "headerHeight": 12, "visibleIf": "if(params.showTable, if(params.nestedGate, \"shown\", \"hidden\"), \"hidden\")", "columns": [{"id": "e3", "label": "Ignored", "width": 80, "bind": "{{params.columnOnly}}"}]}`
	input = bytes.Replace(input, []byte("      ]\n    },"), []byte(table+"\n      ]\n    },"), 1)
	tpl, err := ParseTemplate(input)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParameterReferences(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"nestedGate", "showTable"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("references = %#v, want %#v", got, want)
	}
}

func TestParameterReferencesKeepsVisibleIfMissingParameterLocatedAtItsElement(t *testing.T) {
	input, err := os.ReadFile("testdata/example/first-pdf.folio")
	if err != nil {
		t.Fatal(err)
	}
	input = bytes.Replace(input, []byte(`"value": "Hello, {{customer.name}}!"`), []byte(`"visibleIf": "params.reportDate", "value": "Hello, {{customer.name}}!"`), 1)
	tpl, err := ParseTemplate(input)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParameterReferences(tpl)
	if err != nil || !reflect.DeepEqual(got, []string{"reportDate"}) {
		t.Fatalf("references = %#v, err=%v", got, err)
	}
	_, err = Render(tpl, Data(`{}`), Params(`{}`), FontSet{})
	if err == nil || !strings.Contains(err.Error(), "params.reportDate") || !strings.Contains(err.Error(), "e1") {
		t.Fatalf("missing parameter diagnostic = %v; want engine-located e1 params.reportDate provenance", err)
	}
}

func TestParameterReferencesRejectsNameLongerThanProtocolBound(t *testing.T) {
	name := strings.Repeat("a", MaxParameterReferenceNameLength+1)
	input, err := os.ReadFile("testdata/example/first-pdf.folio")
	if err != nil {
		t.Fatal(err)
	}
	input = bytes.Replace(input, []byte("{{customer.name}}"), []byte("{{params."+name+"}}"), 1)
	tpl, err := ParseTemplate(input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParameterReferences(tpl)
	if err == nil || !strings.Contains(err.Error(), "128-character") {
		t.Fatalf("long parameter name error = %v", err)
	}
}
