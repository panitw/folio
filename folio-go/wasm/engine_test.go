package wasm

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	folio "github.com/panitw/folio/folio-go"
	"github.com/panitw/folio/folio-go/fonts"
)

func TestEngineLoadAndSerializeRoundTripsCanonicalBytes(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	nonCanonical := append([]byte("\n  "), input...)
	snapshot, err := engine.Load(nonCanonical)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.DocumentState != "loaded" || snapshot.ByteLength != len(input) {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	got, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, input) {
		t.Fatal("worker engine serialization did not come from the Go serializer")
	}
	got[0] ^= 1
	again, _, err := engine.Serialize()
	if err != nil || !bytes.Equal(again, input) {
		t.Fatal("Serialize exposed aliased authoritative bytes")
	}
}

func TestEngineParameterReferencesAreARevisionCorrelatedProjection(t *testing.T) {
	engine := NewEngine()
	input, err := os.ReadFile("../testdata/example/first-pdf.folio")
	if err != nil {
		t.Fatal(err)
	}
	input = bytes.Replace(input, []byte("{{customer.name}}"), []byte("{{params.reportDate}} {{params.reportDate}}"), 1)
	if _, err := engine.Load(input); err != nil {
		t.Fatal(err)
	}
	references, revision, err := engine.ParameterReferences()
	if err != nil || !reflect.DeepEqual(references, []string{"reportDate"}) || revision != engine.Snapshot().Revision {
		t.Fatalf("parameter references = %#v/%d, err=%v", references, revision, err)
	}
}

func TestEngineEmptyParameterReferencesRemainAnArrayForWorkerTransport(t *testing.T) {
	engine := NewEngine()
	input, err := os.ReadFile("../testdata/example/first-pdf.folio")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Load(input); err != nil {
		t.Fatal(err)
	}
	references, _, err := engine.ParameterReferences()
	if err != nil {
		t.Fatal(err)
	}
	if references == nil || len(references) != 0 {
		t.Fatalf("empty parameter references must be a non-nil array for JSON transport, got %#v", references)
	}
}

func TestEngineTableColumnsAreRevisionCorrelatedAndHistoryOwned(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	if _, err := engine.Load(input); err != nil {
		t.Fatal(err)
	}
	created, err := engine.Apply([]byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	var tableID string
	for _, component := range created.Canvas.Components {
		if component.Type == "table" {
			tableID = component.ID
		}
	}
	if tableID == "" {
		t.Fatal("table was not projected")
	}
	before := engine.Snapshot()
	after, err := engine.Apply([]byte(`{"kind":"addTableColumn","version":1,"id":"` + tableID + `","index":0}`))
	if err != nil || after.Revision != before.Revision+1 || !after.CanUndo {
		t.Fatalf("add history = %#v, err=%v", after, err)
	}
	projection, err := engine.TableColumns(tableID)
	if err != nil || projection.Revision != after.Revision || len(projection.Table.Columns) != 1 {
		t.Fatalf("projection = %#v, err=%v", projection, err)
	}
	if _, err := engine.Undo(); err != nil {
		t.Fatal(err)
	}
	undone, err := engine.TableColumns(tableID)
	if err != nil || len(undone.Table.Columns) != 0 || undone.Revision == projection.Revision {
		t.Fatalf("undo projection = %#v, err=%v", undone, err)
	}
	beforeRejected := engine.Snapshot()
	if _, err := engine.Apply([]byte(`{"kind":"addTableColumn","version":1,"id":"` + tableID + `","index":9}`)); err == nil {
		t.Fatal("invalid table command succeeded")
	}
	if afterRejected := engine.Snapshot(); !reflect.DeepEqual(afterRejected, beforeRejected) {
		t.Fatalf("rejection changed history/revision: %#v != %#v", afterRejected, beforeRejected)
	}
}

func TestEngineRenderMatchesTheNativeProductionPathByteForByte(t *testing.T) {
	fixtures := []struct{ name, template, data, params string }{
		{"simple", "../testdata/example/first-pdf.folio", `{"customer":{"name":"Ada"}}`, `{"preview":null}`},
		// This is a genuine five-page, table/text, multi-script shipped-font
		// document. A one-page ASCII fixture cannot detect pagination or font
		// path divergence, so both inputs remain required parity subjects.
		{"multipage-text-font", "../../fixtures/statement-5/input.folio", "../../fixtures/statement-5/data.json", "../../fixtures/statement-5/params.json"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			input, err := os.ReadFile(fixture.template)
			if err != nil {
				t.Fatal(err)
			}
			data, params := []byte(fixture.data), []byte(fixture.params)
			if fixture.name != "simple" {
				data, err = os.ReadFile(fixture.data)
				if err != nil {
					t.Fatal(err)
				}
				params, err = os.ReadFile(fixture.params)
				if err != nil {
					t.Fatal(err)
				}
			}
			engine := NewEngine()
			if _, err := engine.Load(input); err != nil {
				t.Fatal(err)
			}
			canonical, snapshot, err := engine.Serialize()
			if err != nil {
				t.Fatal(err)
			}
			wantTemplate, err := folio.ParseTemplate(canonical)
			if err != nil {
				t.Fatal(err)
			}
			want, err := folio.Render(wantTemplate, folio.Data(data), folio.Params(params), fonts.Shipped())
			if err != nil {
				t.Fatal(err)
			}
			got, evidence, err := engine.Render(canonical, data, params)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want.Bytes) || evidence.Revision != snapshot.Revision {
				t.Fatalf("wasm render diverged from production: revision=%d want=%d", evidence.Revision, snapshot.Revision)
			}
			wantDigest := sha256.Sum256(want.Bytes)
			if evidence.PDFSHA256 != fmt.Sprintf("%x", wantDigest) {
				t.Fatalf("digest = %q", evidence.PDFSHA256)
			}
			identity, identityRevision, err := engine.PreviewIdentity(data, params)
			// This is deliberately not a same-helper comparison: the engine must
			// supply precisely the complete shipped set to the public identity
			// contract. Omitting any production face in Engine.PreviewIdentity
			// therefore disagrees with this independently assembled expectation.
			wantIdentity := folio.PreviewIdentity(canonical, folio.Data(data), folio.Params(params), fonts.Shipped())
			if err != nil || identity != wantIdentity || evidence.Identity != wantIdentity || identityRevision != snapshot.Revision {
				t.Fatalf("identity evidence = %q/%d, render = %q, want=%q, err=%v", identity, identityRevision, evidence.Identity, wantIdentity, err)
			}
			for face, program := range fonts.Shipped() {
				changed := append([]byte(nil), program...)
				changed[0] ^= 1
				mutated := fonts.Shipped()
				mutated[face] = changed
				if folio.PreviewIdentity(canonical, folio.Data(data), folio.Params(params), mutated) == wantIdentity {
					t.Fatalf("shipped face %q did not affect preview identity", face)
				}
			}
			if _, _, err := engine.Render(append(canonical, ' '), []byte(`{}`), []byte(`{}`)); err == nil {
				t.Fatal("stale/noncanonical template render unexpectedly succeeded")
			}
		})
	}
}

func TestEngineRejectedLoadIsTransactional(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	before, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	beforeBytes, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	invalid := bytes.Replace(input, []byte(`"top": 36`), []byte(`"top": 800`), 1)
	if _, err := engine.Load(invalid); err == nil {
		t.Fatal("invalid projection load unexpectedly succeeded")
	}
	after := engine.Snapshot()
	afterBytes, _, err := engine.Serialize()
	if err != nil || !reflect.DeepEqual(after, before) || !bytes.Equal(afterBytes, beforeBytes) {
		t.Fatalf("rejected load changed engine: before=%#v after=%#v", before, after)
	}
}

func TestEngineRejectsUnknownCommandWithoutChangingDocument(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	before, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Apply([]byte(`{"kind":"unknown"}`)); err == nil {
		t.Fatal("unknown command unexpectedly succeeded")
	}
	after := engine.Snapshot()
	if after != before {
		t.Fatalf("unknown command changed state: before=%#v after=%#v", before, after)
	}
}

func TestEngineCommitsComponentChangesThroughGoOwnedCommandChannel(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	before, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	after, err := engine.Apply([]byte(`{"kind":"createComponent","version":1,"type":"text","band":"content","x":12,"y":12,"width":72,"height":24,"snap":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != before.Revision+1 || after.ByteLength <= before.ByteLength || after.Canvas == nil || len(after.Canvas.Components) == 0 {
		t.Fatalf("commit snapshot = %#v, before = %#v", after, before)
	}
}

func TestEnginePageSetupRevisionAndProjectionChangeTogether(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	before, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	after, err := engine.Apply([]byte(`{"kind":"pageSetup","version":1,"preset":"Letter","orientation":"landscape","width":1,"height":1,"margin":{"top":36,"right":36,"bottom":36,"left":36}}`))
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != before.Revision+1 || after.Canvas == nil || after.Canvas.Width != 792000 || after.Canvas.Height != 612000 {
		t.Fatalf("page setup snapshot = %#v", after)
	}
}

func TestEnginePropertyBatchAdvancesOneRevisionOrLeavesEverythingUntouched(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	before, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	after, err := engine.Apply([]byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1","e5"],"changes":{"y":{"op":"set","value":12}}}`))
	if err != nil || after.Revision != before.Revision+1 || after.Canvas == nil {
		t.Fatalf("property batch = %#v, %v", after, err)
	}
	bytesBeforeFailure, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Apply([]byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1","missing"],"changes":{"y":{"op":"set","value":18}}}`)); err == nil {
		t.Fatal("bad target unexpectedly succeeded")
	}
	bytesAfterFailure, snapshotAfterFailure, err := engine.Serialize()
	if err != nil || snapshotAfterFailure.Revision != after.Revision || !bytes.Equal(bytesBeforeFailure, bytesAfterFailure) {
		t.Fatal("rejected property batch changed engine state")
	}
}

func TestEngineUndoRedoOwnsCommittedCanonicalHistoryAndResetsOnLoad(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	loaded, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	committed, err := engine.Apply([]byte(`{"kind":"createComponent","version":1,"type":"text","band":"content","x":12,"y":12,"width":72,"height":24,"snap":true}`))
	if err != nil {
		t.Fatal(err)
	}
	after, _, err := engine.Serialize()
	if err != nil || bytes.Equal(before, after) {
		t.Fatal("accepted command did not create a distinct canonical state")
	}
	undone, err := engine.Undo()
	if err != nil || undone.Revision != committed.Revision+1 {
		t.Fatalf("undo = %#v, %v", undone, err)
	}
	got, _, err := engine.Serialize()
	if err != nil || !bytes.Equal(got, before) {
		t.Fatal("undo did not restore engine canonical bytes")
	}
	redone, err := engine.Redo()
	if err != nil || redone.Revision != undone.Revision+1 {
		t.Fatalf("redo = %#v, %v", redone, err)
	}
	got, _, err = engine.Serialize()
	if err != nil || !bytes.Equal(got, after) {
		t.Fatal("redo did not restore engine canonical bytes")
	}
	if _, err := engine.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Apply([]byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":100,"y":100,"width":72,"height":24,"snap":true}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Redo(); !errors.Is(err, ErrNoRedo) {
		t.Fatalf("divergent command retained redo branch: %v", err)
	}
	if _, err := engine.Load(input); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Undo(); !errors.Is(err, ErrNoUndo) {
		t.Fatalf("load retained undo history: %v", err)
	}
	if engine.Snapshot().Revision <= loaded.Revision {
		t.Fatal("revision was not monotonic")
	}
}

func TestEngineScalarBindingIsOneCanonicalUndoableMutation(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	loaded, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	bound, err := engine.Apply([]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}`))
	if err != nil || bound.Revision != loaded.Revision+1 || !bound.CanUndo || bound.Canvas == nil {
		t.Fatalf("scalar bind snapshot = %#v, err=%v", bound, err)
	}
	component := canvasComponentByID(t, bound.Canvas, "e1")
	if component.Binding == nil || *component.Binding != "customer.name" {
		t.Fatalf("engine binding projection = %#v", component.Binding)
	}
	after, _, err := engine.Serialize()
	if err != nil || bytes.Equal(before, after) || !bytes.Contains(after, []byte(`"value": "{{customer.name}}"`)) {
		t.Fatalf("accepted bind was not canonical: %s, err=%v", after, err)
	}
	undone, err := engine.Undo()
	if err != nil || undone.Revision != bound.Revision+1 {
		t.Fatalf("undo scalar bind = %#v, err=%v", undone, err)
	}
	got, _, err := engine.Serialize()
	if err != nil || !bytes.Equal(got, before) {
		t.Fatalf("undo did not restore pre-bind bytes: %v", err)
	}
	redone, err := engine.Redo()
	if err != nil || redone.Revision != undone.Revision+1 {
		t.Fatalf("redo scalar bind = %#v, err=%v", redone, err)
	}
	got, _, err = engine.Serialize()
	if err != nil || !bytes.Equal(got, after) {
		t.Fatalf("redo did not restore scalar binding: %v", err)
	}
	stable := engine.Snapshot()
	if _, err := engine.Apply([]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["params","name"]}`)); err == nil {
		t.Fatal("params scalar bind unexpectedly succeeded")
	}
	if engine.Snapshot().Revision != stable.Revision {
		t.Fatal("rejected scalar bind advanced revision")
	}
}

func TestEngineScalarBindingRoundTripsAndNoOpPreservesHistoryBranches(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	if _, err := engine.Load(input); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Apply([]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Apply([]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["account","number"]}`)); err != nil {
		t.Fatal(err)
	}
	undone, err := engine.Undo()
	if err != nil || !undone.CanRedo {
		t.Fatalf("undo before no-op = %#v, err=%v", undone, err)
	}
	beforeNoOp, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	noOp, err := engine.Apply([]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}`))
	if err != nil || noOp.Revision != undone.Revision || !noOp.CanRedo {
		t.Fatalf("same binding must preserve revision and redo: %#v, err=%v", noOp, err)
	}
	afterNoOp, _, err := engine.Serialize()
	if err != nil || !bytes.Equal(beforeNoOp, afterNoOp) {
		t.Fatalf("same binding changed canonical bytes: %v", err)
	}
	if _, err := engine.Redo(); err != nil {
		t.Fatalf("no-op binding discarded redo: %v", err)
	}
	canonical, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	reloaded := NewEngine()
	loaded, err := reloaded.Load(canonical)
	if err != nil || loaded.Canvas == nil || canvasComponentByID(t, loaded.Canvas, "e1").Binding == nil {
		t.Fatalf("saved canonical binding did not survive load: %#v, err=%v", loaded, err)
	}
}

func canvasComponentByID(t *testing.T, canvas *folio.CanvasProjection, id string) folio.CanvasComponent {
	t.Helper()
	for _, component := range canvas.Components {
		if component.ID == id {
			return component
		}
	}
	t.Fatalf("component %q is absent from wasm canvas", id)
	return folio.CanvasComponent{}
}

func TestEngineNoOpDoesNotChangeHistoryRevisionOrRedo(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	if _, err := engine.Load(input); err != nil {
		t.Fatal(err)
	}
	command := []byte(`{"kind":"pageSetup","version":1,"preset":"custom","orientation":"landscape","width":300.125,"height":400.5,"margin":{"top":10,"right":11.5,"bottom":12,"left":13}}`)
	changed, err := engine.Apply(command)
	if err != nil || !changed.CanUndo || changed.CanRedo {
		t.Fatalf("first command = %#v, %v", changed, err)
	}
	undone, err := engine.Undo()
	if err != nil || undone.CanUndo || !undone.CanRedo {
		t.Fatalf("undo = %#v, %v", undone, err)
	}
	stable, err := engine.Apply([]byte(`{"kind":"pageSetup","version":1,"preset":"A4","orientation":"portrait","width":0,"height":0,"margin":{"top":36,"right":36,"bottom":36,"left":36}}`))
	if err != nil {
		t.Fatal(err)
	}
	if stable.Revision != undone.Revision || stable.CanUndo != undone.CanUndo || stable.CanRedo != undone.CanRedo {
		t.Fatalf("no-op changed history evidence: before=%#v after=%#v", undone, stable)
	}
	redone, err := engine.Redo()
	if err != nil || redone.Revision != stable.Revision+1 || !redone.CanUndo || redone.CanRedo {
		t.Fatalf("no-op cleared redo or changed revision: %#v, %v", redone, err)
	}
}

func TestEngineDuplicateIsACommittedGoCommand(t *testing.T) {
	input, err := os.ReadFile("../testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	before, err := engine.Load(input)
	if err != nil {
		t.Fatal(err)
	}
	after, err := engine.Apply([]byte(`{"kind":"duplicateComponent","version":1,"id":"e1","snap":true}`))
	if err != nil || after.Revision != before.Revision+1 || after.Canvas == nil || len(after.Canvas.Components) != len(before.Canvas.Components)+1 {
		t.Fatalf("duplicate = %#v, %v", after, err)
	}
	if after.Canvas.Components[len(after.Canvas.Components)-1].ID == "e1" {
		t.Fatal("duplicate retained its source opaque id")
	}
}

// ---------------------------------------------------------------------------
// STORY 8.1: THE FONT-CHAIN COMMANDS AS ENGINE TRANSACTIONS.
//
// fontChainEngineDocJSON declares THREE chains and names two of them from
// elements in different bands, including a table's headerStyle. It is the
// repo's first byte-level pin on multi-chain emission (Design Notes R2).
const fontChainEngineDocJSON = `{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e7", "type": "text", "x": 0, "y": 0, "width": 100, "height": 20, "value": "content", "style": {"fontFamily": "body"}},
      {"id": "e9", "type": "table", "x": 0, "y": 30, "bind": "items[]", "headerHeight": 20,
        "columns": [{"id": "e10", "label": "Date", "width": 100, "bind": "{{row.a}}"}],
        "headerStyle": {"fontFamily": "body"}}
    ]},
    "pageFooter": {"elements": [{"id": "e12", "type": "text", "x": 0, "y": 0, "width": 100, "height": 20, "value": "footer", "style": {"fontFamily": "heading"}}], "height": 20},
    "pageHeader": {"elements": [{"id": "e2", "type": "text", "x": 0, "y": 0, "width": 100, "height": 20, "value": "header", "style": {"fontFamily": "body"}}], "height": 20}
  },
  "fonts": {"body": ["Noto Sans", "Noto Sans Thai"], "heading": ["Noto Sans"], "unused": ["Noto Sans SC"]},
  "locale": "en",
  "nextId": 39,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}`

func fontChainEngine(t *testing.T) *Engine {
	t.Helper()
	engine := NewEngine()
	if _, err := engine.Load([]byte(fontChainEngineDocJSON)); err != nil {
		t.Fatal(err)
	}
	return engine
}

func fontChainFamilies(t *testing.T, snapshot Snapshot) []string {
	t.Helper()
	if snapshot.Canvas == nil {
		t.Fatal("snapshot carries no canvas")
	}
	return snapshot.Canvas.FontFamilies
}

// TestEngineFontChainCommandsAdvanceExactlyOneRevisionAndOneUndoStep is AC1:
// every accepted chain command is ONE committed mutation, by construction —
// Apply's single pushUndo — and a refused one commits nothing.
func TestEngineFontChainCommandsAdvanceExactlyOneRevisionAndOneUndoStep(t *testing.T) {
	for _, command := range []string{
		`{"kind":"addFontChain","version":1,"name":"caption","entries":["Noto Sans"]}`,
		`{"kind":"renameFontChain","version":1,"name":"body","to":"brand"}`,
		`{"kind":"deleteFontChain","version":1,"name":"unused"}`,
		`{"kind":"addFontChainEntry","version":1,"name":"body","index":0,"face":"Noto Sans SC"}`,
		`{"kind":"moveFontChainEntry","version":1,"name":"body","from":0,"to":1}`,
		`{"kind":"removeFontChainEntry","version":1,"name":"body","index":0}`,
	} {
		engine := fontChainEngine(t)
		before := engine.Snapshot()
		after, err := engine.Apply([]byte(command))
		if err != nil {
			t.Fatalf("%s: %v", command, err)
		}
		if after.Revision != before.Revision+1 || !after.CanUndo || after.CanRedo || after.Canvas == nil {
			t.Fatalf("%s snapshot = %#v", command, after)
		}
		undone, err := engine.Undo()
		if err != nil {
			t.Fatalf("%s undo: %v", command, err)
		}
		if undone.CanUndo {
			t.Fatalf("%s pushed more than one undo entry", command)
		}
		restored, _, err := engine.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		fresh := fontChainEngine(t)
		original, _, err := fresh.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(restored, original) {
			t.Fatalf("%s: one undo did not restore the document", command)
		}
	}
}

// TestEngineFontChainRenameUndoesTheMapAndTheElementsTogether is AC2: the map
// key and all four references move in one entry, so one undo restores every
// one of them. A rename that pushed twice, or that carried the elements in a
// second transaction, would fail here rather than at render.
func TestEngineFontChainRenameUndoesTheMapAndTheElementsTogether(t *testing.T) {
	engine := fontChainEngine(t)
	before, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	renamed, err := engine.Apply([]byte(`{"kind":"renameFontChain","version":1,"name":"body","to":"brand"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := fontChainFamilies(t, renamed); !reflect.DeepEqual(got, []string{"brand", "heading", "unused"}) {
		t.Fatalf("families after rename = %#v", got)
	}
	after, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(after, []byte(`"body"`)) {
		t.Fatal("the old chain name survives somewhere in the renamed document")
	}
	if bytes.Count(after, []byte(`"fontFamily": "brand"`)) != 3 {
		t.Fatalf("renamed document carries %d brand references, want the three style/headerStyle bearers", bytes.Count(after, []byte(`"fontFamily": "brand"`)))
	}
	undone, err := engine.Undo()
	if err != nil || undone.CanUndo {
		t.Fatalf("undo = %#v, %v — a rename is ONE history entry", undone, err)
	}
	restored, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, restored) {
		t.Fatal("one undo did not restore the fonts map AND the elements together")
	}
}

func TestEngineRefusedFontChainCommandLeavesByteRevisionAndHistoryUntouched(t *testing.T) {
	engine := fontChainEngine(t)
	committed, err := engine.Apply([]byte(`{"kind":"addFontChain","version":1,"name":"caption","entries":["Noto Sans"]}`))
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	for _, refused := range []string{
		`{"kind":"addFontChain","version":1,"name":"caption","entries":["Noto Sans"]}`,
		`{"kind":"deleteFontChain","version":1,"name":"body"}`,
		`{"kind":"removeFontChainEntry","version":1,"name":"heading","index":0}`,
		`{"kind":"renameFontChain","version":1,"name":"body","to":"heading"}`,
	} {
		if _, err := engine.Apply([]byte(refused)); err == nil {
			t.Fatalf("%s unexpectedly succeeded", refused)
		}
		after, snapshot, err := engine.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) || snapshot.Revision != committed.Revision || snapshot.CanUndo != committed.CanUndo || snapshot.CanRedo != committed.CanRedo {
			t.Fatalf("%s changed engine state: %#v", refused, snapshot)
		}
	}
}

// TestEngineFontChainRenameOutAndBackIsByteIdentical is Design Notes R2's
// claim, measured on a MULTI-CHAIN document: a chain's emitted position is a
// total function of its key and its entries are the slice, so renaming out and
// back must restore the canonical bytes exactly — fonts map AND bands, since a
// rename that failed to restore a style.fontFamily would move the bands bytes
// instead. Two commands, two revisions, two undo steps.
func TestEngineFontChainRenameOutAndBackIsByteIdentical(t *testing.T) {
	engine := fontChainEngine(t)
	loaded := engine.Snapshot()
	original, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	out, err := engine.Apply([]byte(`{"kind":"renameFontChain","version":1,"name":"body","to":"zbrand"}`))
	if err != nil {
		t.Fatal(err)
	}
	moved, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(original, moved) {
		t.Fatal("the rename did not move the canonical bytes; the round trip would prove nothing")
	}
	back, err := engine.Apply([]byte(`{"kind":"renameFontChain","version":1,"name":"zbrand","to":"body"}`))
	if err != nil {
		t.Fatal(err)
	}
	final, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, final) {
		t.Fatal("rename out and back did not restore the canonical bytes exactly")
	}
	if out.Revision != loaded.Revision+1 || back.Revision != out.Revision+1 {
		t.Fatalf("revisions = %d, %d, %d — two commands are two revisions", loaded.Revision, out.Revision, back.Revision)
	}
	if _, err := engine.Undo(); err != nil {
		t.Fatal(err)
	}
	second, err := engine.Undo()
	if err != nil || second.CanUndo {
		t.Fatalf("second undo = %#v, %v — two commands are two undo steps", second, err)
	}
}

// TestEngineFontChainNoOpChangesNothing is the bytes.Equal short-circuit at
// the chain path: a move that reorders nothing is valid and commits nothing.
func TestEngineFontChainNoOpChangesNothing(t *testing.T) {
	engine := fontChainEngine(t)
	committed, err := engine.Apply([]byte(`{"kind":"addFontChain","version":1,"name":"caption","entries":["Noto Sans"]}`))
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	stable, err := engine.Apply([]byte(`{"kind":"moveFontChainEntry","version":1,"name":"body","from":1,"to":1}`))
	if err != nil {
		t.Fatal(err)
	}
	after, _, err := engine.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || stable.Revision != committed.Revision || stable.CanUndo != committed.CanUndo || stable.CanRedo != committed.CanRedo {
		t.Fatalf("a no-op chain command changed engine state: %#v", stable)
	}
}

// TestEngineProjectsTheChainsThemselvesNotOnlyTheirNames is AC5: the canvas
// projection carries the entries, so a move or a remove is observable to the
// designer at all. FontChains[i].Name == FontFamilies[i] is asserted, because
// the browser's guard drops the whole snapshot if it stops holding.
func TestEngineProjectsTheChainsThemselvesNotOnlyTheirNames(t *testing.T) {
	engine := fontChainEngine(t)
	snapshot, err := engine.Apply([]byte(`{"kind":"moveFontChainEntry","version":1,"name":"body","from":0,"to":1}`))
	if err != nil {
		t.Fatal(err)
	}
	chains := snapshot.Canvas.FontChains
	names := make([]string, 0, len(chains))
	for _, chain := range chains {
		names = append(names, chain.Name)
	}
	if !reflect.DeepEqual(names, fontChainFamilies(t, snapshot)) {
		t.Fatalf("FontChains names = %#v, FontFamilies = %#v", names, snapshot.Canvas.FontFamilies)
	}
	if !reflect.DeepEqual(chains[0].Entries, []string{"Noto Sans Thai", "Noto Sans"}) {
		t.Fatalf("projected body chain = %#v, want the reordered entries", chains[0].Entries)
	}
}
