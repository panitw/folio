package wasm

import (
	"bytes"
	"os"
	"reflect"
	"testing"
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
