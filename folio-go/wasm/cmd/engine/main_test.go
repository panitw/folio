//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
	"os"
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
