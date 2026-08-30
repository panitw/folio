//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	folio "github.com/panitw/folio/folio-go"
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

// TestWasmHostSanitizesTemplateDiagnostics carries BOTH arms of the
// boundary rule, because Story 7.8 moved a population across it.
//
// ARM ONE — the destruction rule itself, on the population that still
// carries it. TEMPLATE_MALFORMED's message is replaced wholesale
// because that message may quote the offending document back, and a
// large or hostile one would be reflected instead of described. The
// fixture is re-pointed, not the rule: it used to be a well-formed
// object with a 2048-byte value, which reached here as an UNCODED
// *LoadError and so was bucketed under TEMPLATE_MALFORMED along with
// every located field error in the format. D-7.8.1 stopped bucketing
// LoadErrors there, so a document that is merely WRONG no longer
// exercises this rule; a document that is not a document does.
//
// ARM TWO — and it exists because arm one alone would have been a green
// test measuring the residue. When a change moves a population OUT of a
// guard's scope, the guard's test must be re-pointed at a member still
// in scope AND the departed population asserted under its new
// treatment. The departed population is the parseable-but-invalid
// document, and its new treatment is D-7.8.5's: the message SURVIVES,
// so the author is finally told which element and which field — and the
// fragment of their own document quoted back inside it is bounded IN
// RUNES and visibly elided, which is what makes surviving safe.
func TestWasmHostSanitizesTemplateDiagnostics(t *testing.T) {
	t.Run("unparseable-bytes-are-described-not-reflected", func(t *testing.T) {
		engine := wasm.NewEngine()
		malicious := `{"version":"1.0","page":"` + string(bytes.Repeat([]byte("x"), 2048)) + `"`
		got := dispatch(engine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(malicious))})
		if got.OK || got.DiagnosticCode != folio.DiagCodeTemplateMalformed || got.Message != "The template could not be processed" {
			t.Fatalf("diagnostic = %#v", got)
		}
		if len(got.Message) > 512 || bytes.Contains([]byte(got.Message), []byte("xxx")) {
			t.Fatalf("unsafe message = %q", got.Message)
		}
	})

	t.Run("parseable-but-invalid-keeps-its-message-bounded-and-elided", func(t *testing.T) {
		// A WELL-FORMED document whose `style` key holds 2048 Thai
		// characters — the exact shape that regressed when D-7.8.1
		// landed alone. Thai and not "xxx" on purpose: a bound that
		// counted BYTES would pass on ASCII while handing this author a
		// third of the budget, and would split a rune on the way out.
		const thai = "\u0e01"
		value := strings.Repeat(thai, 2048)
		doc := `{"assets":{},"bands":{"content":{"elements":[{"id":"e1","type":"text","x":0,"y":0,"width":200,"height":40,"value":"v","style":"` + value + `"}]},"pageFooter":{"elements":[],"height":20},"pageHeader":{"elements":[],"height":20}},"fonts":{"body":["Noto Sans"]},"locale":"en","nextId":2,"page":{"margin":{"bottom":36,"left":36,"right":36,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"1.0"}`
		engine := wasm.NewEngine()
		got := dispatch(engine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(doc))})
		if got.OK {
			t.Fatalf("a non-object style must fail the load: %#v", got)
		}
		if got.DiagnosticCode != folio.DiagCodeTemplateFieldInvalid {
			t.Fatalf("code = %q, want %q", got.DiagnosticCode, folio.DiagCodeTemplateFieldInvalid)
		}
		if got.Message == "The template could not be processed" {
			t.Fatal("the message was destroyed: a parseable document's located refusal must reach its author (D-7.8.1)")
		}
		for _, want := range []string{"style", "e1"} {
			if !strings.Contains(got.Message, want) {
				t.Fatalf("the message must name %q, got %q", want, got.Message)
			}
		}
		reflected := strings.Count(got.Message, thai)
		if reflected > 84 {
			t.Fatalf("the host reported %d runes of the author's own document; the bound is 84 (D-7.8.5)", reflected)
		}
		if reflected <= 28 {
			t.Fatalf("only %d runes survived: the bound is counting BYTES, not runes (D-7.8.5)", reflected)
		}
		if !strings.Contains(got.Message, "\u2026") {
			t.Fatalf("the elided message must say so, got %q", got.Message)
		}
		if !utf8.ValidString(got.Message) {
			t.Fatalf("the reported message is not valid UTF-8 — a bound split a rune: %q", got.Message)
		}
		t.Logf("bounded reflection at the host: %d bytes, %d author runes, code=%s", len(got.Message), reflected, got.DiagnosticCode)
	})

	t.Run("a-multi-fragment-runaway-still-fits-the-host-window", func(t *testing.T) {
		// The leg above spends ONE fragment budget. This one spends
		// three at once: claimID passes a raw element id as ElementID,
		// again as Value, and again inside Reason via
		// validateElementID's %q. Four per-fragment rune bounds share
		// no budget, so the assembled sentence measured 1142 bytes —
		// and bounded() here is a raw BYTE slice value[:512], which
		// delivered it to the author cut mid-rune with no elision
		// marker. Thai on purpose, for the same reason as above.
		const thai = "\u0e01"
		id := strings.Repeat(thai, 2048)
		doc := `{"assets":{},"bands":{"content":{"elements":[{"id":"` + id + `","type":"text","x":0,"y":0,"width":200,"height":40,"value":"v"}]},"pageFooter":{"elements":[],"height":20},"pageHeader":{"elements":[],"height":20}},"fonts":{"body":["Noto Sans"]},"locale":"en","nextId":2,"page":{"margin":{"bottom":36,"left":36,"right":36,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"1.0"}`
		engine := wasm.NewEngine()
		got := dispatch(engine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(doc))})
		if got.OK {
			t.Fatalf("an id that is not an id must fail the load: %#v", got)
		}
		if len(got.Message) > 512 {
			t.Fatalf("the host reported %d bytes, past its own 512-byte window", len(got.Message))
		}
		if !utf8.ValidString(got.Message) {
			t.Fatalf("the reported message is not valid UTF-8 — the host's byte slice split a rune: %q", got.Message)
		}
		if !strings.Contains(got.Message, "\u2026") {
			t.Fatalf("a message this long was truncated somewhere and must say so, got %q", got.Message)
		}
		t.Logf("multi-fragment runaway at the host: %d bytes, code=%s", len(got.Message), got.DiagnosticCode)
	})
}

// TestWasmHostReportsTheTableJustifyRefusalIntact is Story 7.8's AC4,
// asserted at the only place it is actually decided: the wasm boundary.
//
// A designer author opens a hand-edited `.folio` whose table carries
// `style.align: "justify"`. Before this story that document LOADED,
// paid a MAJOR version bump and drew every cell at the start edge with
// no diagnostic at all. Refusing it at load is only half the fix — an
// UNCODED refusal would have arrived here as TEMPLATE_MALFORMED, whose
// message reportableMessage replaces with "The template could not be
// processed", and the author would be told nothing about which element
// or which field. Every Go-side assertion would still have been green.
//
// This is why D-7.8.1 had to be settled before the story could be
// written, and it is the shape of
// TestWasmHostReportsTheLineSpacingRefusalIntact above.
func TestWasmHostReportsTheTableJustifyRefusalIntact(t *testing.T) {
	const doc = `{"assets":{},"bands":{"content":{"elements":[{"id":"e1","type":"table","x":0,"y":0,"bind":"rows[]","as":"row","headerHeight":20,"columns":[{"id":"e2","label":"L","width":100,"bind":"{{row.v}}"}],"style":{"fontFamily":"body","fontSize":11,"align":"justify"}}]},"pageFooter":{"elements":[],"height":20},"pageHeader":{"elements":[],"height":20}},"fonts":{"body":["Noto Sans"]},"locale":"en","nextId":3,"page":{"margin":{"bottom":36,"left":36,"right":36,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"2.0"}`
	engine := wasm.NewEngine()
	got := dispatch(engine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(doc))})
	if got.OK {
		t.Fatalf("a table carrying style.align: \"justify\" must fail the load: %#v", got)
	}
	if got.DiagnosticCode != folio.DiagCodeTemplateFieldInvalid {
		t.Fatalf("code = %q, want %q", got.DiagnosticCode, folio.DiagCodeTemplateFieldInvalid)
	}
	if got.Message == "The template could not be processed" {
		t.Fatal("the message was replaced by the malformed-template placeholder — the author is told neither which element nor which field")
	}
	for _, want := range []string{"e1", "style.align", "left, center, right"} {
		if !strings.Contains(got.Message, want) {
			t.Fatalf("the message must name %q, got %q", want, got.Message)
		}
	}
	if strings.Contains(got.Message, "right, justify") {
		t.Fatalf("the message must not name justify among a table's legal values, got %q", got.Message)
	}

	// The SAME value on a TEXT element still loads. Without this leg the
	// assertions above are equally consistent with a blanket ban, which
	// Story 7.3, Story 7.4 and two shipped goldens forbid.
	const textDoc = `{"assets":{},"bands":{"content":{"elements":[{"id":"e1","type":"text","x":0,"y":0,"width":200,"height":40,"value":"v","style":{"fontFamily":"body","fontSize":11,"align":"justify"}}]},"pageFooter":{"elements":[],"height":20},"pageHeader":{"elements":[],"height":20}},"fonts":{"body":["Noto Sans"]},"locale":"en","nextId":2,"page":{"margin":{"bottom":36,"left":36,"right":36,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"2.0"}`
	textEngine := wasm.NewEngine()
	textGot := dispatch(textEngine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(textDoc))})
	if !textGot.OK {
		t.Fatalf("a TEXT element's style.align: \"justify\" must still load (FR47): %#v", textGot)
	}
	t.Logf("table refusal at the host: code=%s message=%q; the same value on a text element loaded", got.DiagnosticCode, got.Message)
}

// TestWasmHostReportsTheLineSpacingRefusalIntact is Story 7.2's AC7 at
// the only place it is actually decided. reportableMessage replaces a
// diagnostic's message with "The template could not be processed" for
// TEMPLATE_MALFORMED and for that code ALONE — so an UNCODED lineSpacing
// load error would be destroyed here, before the author ever saw which
// element or which range it was about, and every Go-side assertion would
// still be green. Minting STYLE_LINE_SPACING_INVALID is what makes the AC
// reachable; this is where that is observable.
func TestWasmHostReportsTheLineSpacingRefusalIntact(t *testing.T) {
	engine := wasm.NewEngine()
	doc := `{"assets":{},"bands":{"content":{"elements":[{"id":"e1","type":"text","x":0,"y":0,"width":200,"height":40,"value":"v","style":{"fontFamily":"body","fontSize":11,"lineSpacing":1000.001}}]},"pageFooter":{"elements":[],"height":20},"pageHeader":{"elements":[],"height":20}},"fonts":{"body":["Noto Sans"]},"locale":"en","nextId":2,"page":{"margin":{"bottom":36,"left":36,"right":36,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"1.0"}`
	got := dispatch(engine, request{Operation: "load", PayloadBase64: base64.StdEncoding.EncodeToString([]byte(doc))})
	if got.OK {
		t.Fatalf("an out-of-range lineSpacing must fail the load: %#v", got)
	}
	if got.DiagnosticCode != folio.DiagCodeStyleLineSpacingInvalid {
		t.Fatalf("code = %q, want %q", got.DiagnosticCode, folio.DiagCodeStyleLineSpacingInvalid)
	}
	if got.Message == "The template could not be processed" {
		t.Fatal("the message was replaced by the malformed-template placeholder — the author is told nothing about which element or which range")
	}
	for _, want := range []string{"e1", "lineSpacing"} {
		if !strings.Contains(got.Message, want) {
			t.Fatalf("the message must name %q, got %q", want, got.Message)
		}
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
