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
