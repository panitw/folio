//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/wasm"
)

// This is compiled for the real js/wasm host every story. Browser execution is
// owned by the D-000.4 Epic 5 boundary cadence; the native wasm.Engine test
// covers the same canonical fixture on every ordinary Go run.
func TestWasmHostRoundTripsCanonicalFixture(t *testing.T) {
	input, err := os.ReadFile("../../../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := wasm.NewEngine()
	nonCanonical := append([]byte("\n  "), input...)
	loaded := dispatch(engine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString(nonCanonical)})
	if !loaded.OK || loaded.Snapshot.DocumentState != "loaded" {
		t.Fatalf("load response = %#v", loaded)
	}
	serialized := dispatch(engine, request{Operation: "serialize"})
	if !serialized.OK {
		t.Fatalf("serialize response = %#v", serialized)
	}
	got, err := base64.StdEncoding.DecodeString(serialized.BytesBase64)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, input) {
		t.Fatal("wasm host serialization bypassed canonical Go serialization")
	}
}

func TestWasmHostSanitizesTemplateDiagnostics(t *testing.T) {
	engine := wasm.NewEngine()
	malicious := `{"version":"1.0","page":"` + string(bytes.Repeat([]byte("x"), 2048)) + `"}`
	got := dispatch(engine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(malicious))})
	if got.OK || got.DiagnosticCode == "" || got.Message != "The template could not be processed" {
		t.Fatalf("diagnostic = %#v", got)
	}
	if len(got.Message) > 512 || bytes.Contains([]byte(got.Message), []byte("xxx")) {
		t.Fatalf("unsafe message = %q", got.Message)
	}
}

func TestWasmHostReportsEngineAuthoredRenderMessages(t *testing.T) {
	starter, err := os.ReadFile("../../../../folio-designer/public/templates/starter.folio")
	if err != nil {
		t.Fatal(err)
	}
	encode := func(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
	place := func(changes string) (*wasm.Engine, []byte) {
		engine := wasm.NewEngine()
		if loaded := dispatch(engine, request{Operation: "load", PayloadBase64: encode(starter)}); !loaded.OK {
			t.Fatalf("load = %#v", loaded)
		}
		created := dispatch(engine, request{Operation: "command", PayloadBase64: encode([]byte(`{"kind":"createComponent","version":1,"type":"text","band":"content","x":40,"y":40,"width":200,"height":24,"snap":false}`))})
		if !created.OK {
			t.Fatalf("create = %#v", created)
		}
		if changes != "" {
			changed := dispatch(engine, request{Operation: "command", PayloadBase64: encode([]byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":` + changes + `}`))})
			if !changed.OK {
				t.Fatalf("change = %#v", changed)
			}
		}
		serialized := dispatch(engine, request{Operation: "serialize"})
		if !serialized.OK {
			t.Fatalf("serialize = %#v", serialized)
		}
		out, err := base64.StdEncoding.DecodeString(serialized.BytesBase64)
		if err != nil {
			t.Fatal(err)
		}
		return engine, out
	}
	data := encode([]byte(`{"amount":10000}`))
	params := encode([]byte(`{}`))

	// A wrong-kind binding carries its own diagnostic code; the message must
	// name the element and the reason rather than a fixed placeholder.
	boundEngine, bound := place(`{"expression":{"op":"set","value":"{{amount}}"}}`)
	got := dispatch(boundEngine, request{Operation: "render", TemplateBase64: encode(bound), DataBase64: data, ParamsBase64: params})
	if got.OK || got.DiagnosticCode == "" || !strings.Contains(got.Message, "not a string") || got.ElementID != "e1" {
		t.Fatalf("bind render = %#v", got)
	}

	// A render failure with no diagnostic code of its own still reports what
	// the engine said instead of collapsing to "the engine rejected".
	unfontedEngine, unfonted := place(`{"fontFamily":{"op":"clear"}}`)
	got = dispatch(unfontedEngine, request{Operation: "render", TemplateBase64: encode(unfonted), DataBase64: data, ParamsBase64: params})
	if got.OK || got.DiagnosticCode != "ENGINE_REJECTED" || !strings.Contains(got.Message, "style.fontFamily") {
		t.Fatalf("unfonted render = %#v", got)
	}
	if len(got.Message) > 512 {
		t.Fatalf("unbounded message = %d bytes", len(got.Message))
	}
}

func TestTableColumnsRequestRequiresTheExactSelectionEnvelope(t *testing.T) {
	engine := wasm.NewEngine()
	payload := base64.StdEncoding.EncodeToString([]byte(`{"id":"e7"}`))
	for _, in := range []request{
		{Operation: "table-columns", TemplateBase64: base64.StdEncoding.EncodeToString([]byte("template")), PayloadBase64: payload},
		{Operation: "table-columns", DataBase64: base64.StdEncoding.EncodeToString([]byte("sample")), PayloadBase64: payload},
		{Operation: "table-columns", ParamsBase64: base64.StdEncoding.EncodeToString([]byte("params")), PayloadBase64: payload},
		{Operation: "table-columns", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(`{"id":"e7","bind":"row.amount"}`))},
		{Operation: "table-columns", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(`{"id":"e7","footer":"sum"}`))},
		{Operation: "table-columns", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(`{"id":"e7","sample":"data"}`))},
		{Operation: "table-columns", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(`{"id":"e7"}{"id":"e8"}`))},
		{Operation: "table-columns", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(`{"id":"` + string(bytes.Repeat([]byte("e"), 129)) + `"}`))},
	} {
		out := dispatch(engine, in)
		if out.OK || out.DiagnosticCode != "WASM_INPUT_INVALID" {
			t.Fatalf("table envelope = %#v", out)
		}
	}
}
