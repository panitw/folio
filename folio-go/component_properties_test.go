package folio

import (
	"bytes"
	"strings"
	"testing"
)

func TestUpdateComponentPropertiesIsClosedCanonicalAndAtomic(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"x":{"op":"set","value":12.125},"fontSize":{"op":"set","value":10},"visibleIf":{"op":"set","value":"customer.active"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if projection.Components[0].ID == "" {
		t.Fatal("missing property projection")
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) || !strings.Contains(string(after), `"x": 12.125`) || !strings.Contains(string(after), `"visibleIf": "customer.active"`) {
		t.Fatalf("canonical property change missing: %s", after)
	}
	if _, err := ParseTemplate(after); err != nil {
		t.Fatalf("property command did not preserve load validation: %v", err)
	}
}

func TestUpdateComponentPropertiesRejectsTableGeometryAndRollsBackBatch(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e2"],"changes":{"width":{"op":"set","value":72}}}`)); err == nil || !strings.Contains(err.Error(), "not editable") {
		t.Fatalf("table width error = %v", err)
	}
	afterTable, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, afterTable) {
		t.Fatal("rejected table width changed canonical bytes")
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1","missing"],"changes":{"x":{"op":"set","value":12}}}`)); err == nil {
		t.Fatal("bad multi-target unexpectedly succeeded")
	}
	afterBatch, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, afterBatch) {
		t.Fatal("bad multi-target partially changed canonical bytes")
	}
}

func TestUpdateComponentPropertiesRejectsStylePlaceholderAndClearIsAbsent(t *testing.T) {
	tpl := componentTemplate(t)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"{{customer.color}}"}}}`)); err == nil {
		t.Fatal("style placeholder unexpectedly succeeded")
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"fontSize":{"op":"clear"}}}`)); err != nil {
		t.Fatal(err)
	}
	after, _ := SerializeTemplate(tpl)
	if strings.Contains(string(after), `"fontSize": 12`) {
		t.Fatalf("clear did not serialize absent: %s", after)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"visibleIf":{"op":"null"}}}`)); err != nil {
		t.Fatal(err)
	}
	afterNull, _ := SerializeTemplate(tpl)
	if !strings.Contains(string(afterNull), `"visibleIf": null`) {
		t.Fatalf("explicit null was not preserved: %s", afterNull)
	}
}

func TestUpdateComponentPropertiesValidatesColourFontAndNullWithoutMutation(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"not-a-colour"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"borderColor":{"op":"set","value":"#12ZZ00"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"fontFamily":{"op":"set","value":"missing-chain"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"fontFamily":{"op":"null"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"bold":{"op":"null"}}}`,
	} {
		if _, err := ApplyComponentCommand(tpl, []byte(command)); err == nil {
			t.Fatalf("invalid property command succeeded: %s", command)
		}
		after, err := SerializeTemplate(tpl)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("invalid command mutated bytes: %v\n%s", err, after)
		}
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"#aBc123"},"fontFamily":{"op":"set","value":"body"}}}`)); err != nil {
		t.Fatal(err)
	}
}

func TestComponentPropertyProjectionIsBoundedAndTypeCoherent(t *testing.T) {
	tpl := componentTemplate(t)
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	for _, component := range projection.Components {
		if component.Type != "text" && component.Type != "table" && (component.FontFamily != nil || component.FontSize != nil || component.Bold != nil || component.Italic != nil || component.Align != nil || component.Valign != nil) {
			t.Fatalf("non-text/table typography leaked into projection: %#v", component)
		}
		if component.Type == "table" && component.TableBind == nil {
			t.Fatalf("table bind missing from display-only projection: %#v", component)
		}
	}
	long := strings.Repeat("x", maxCanvasPropertyString+1)
	bad, err := ParseTemplate([]byte(strings.Replace(string(mustSerialize(t, tpl)), `"fontSize": 12`, `"fontFamily": "`+long+`", "fontSize": 12`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Canvas(bad); err == nil {
		t.Fatal("over-bound fontFamily reached projection")
	}
}

func mustSerialize(t *testing.T, tpl *Template) []byte {
	t.Helper()
	value, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
