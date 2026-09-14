package wasm

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/panitw/folio/folio-go/internal/template"
)

// SPEC-multi-pages story 1: engine commands that address an element by id
// work on a multi-page document. Each of update, move and delete on a page-2
// element changes only that element, and one undo restores the bytes.
func TestIdAddressedCommandsWorkOnALaterPageElementAndUndo(t *testing.T) {
	input, err := os.ReadFile("../../fixtures/multi-page-statement/input.folio")
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngine()
	if _, err := engine.Load(input); err != nil {
		t.Fatal(err)
	}
	if saved, _, err := engine.Serialize(); err != nil || !bytes.Equal(saved, input) {
		t.Fatalf("the multi-page file does not save back byte-for-byte (err %v)", err)
	}
	parse := func(b []byte) *template.Document {
		t.Helper()
		d, err := template.ParseDocument(b)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	const target = "ee"
	indexOf := func(d *template.Document) int {
		for i, el := range d.Pages[1].Elements {
			if string(el.ID) == target {
				return i
			}
		}
		return -1
	}
	for _, c := range []struct {
		command string
		check   func(before, after *template.Document)
	}{
		{`{"kind":"updateComponentProperties","version":1,"ids":["ee"],"changes":{"value":{"op":"set","value":"1. Changed terms."}}}`, func(before, after *template.Document) {
			if got := after.Pages[1].Elements[indexOf(after)].Value.Value; got != "1. Changed terms." {
				t.Errorf("value = %q", got)
			}
		}},
		{`{"kind":"moveComponent","version":1,"id":"ee","x":0,"y":60,"snap":false}`, func(before, after *template.Document) {
			el := after.Pages[1].Elements[indexOf(after)]
			if el.X != 0 || el.Y != 60000 {
				t.Errorf("moved to %d,%d", el.X, el.Y)
			}
		}},
		{`{"kind":"deleteComponent","version":1,"id":"ee"}`, func(before, after *template.Document) {
			if indexOf(after) != -1 || len(after.Pages[1].Elements) != len(before.Pages[1].Elements)-1 {
				t.Error("the element was not deleted from page 2")
			}
		}},
	} {
		before, _, err := engine.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := engine.Apply([]byte(c.command)); err != nil {
			t.Fatalf("%s: %v", c.command, err)
		}
		after, _, err := engine.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		b, a := parse(before), parse(after)
		if a.PageCount() != 2 || len(a.Bands.Content.Elements) != 0 {
			t.Fatalf("%s: the document left the multi-page shape", c.command)
		}
		c.check(b, a)
		// Nothing else changed: page 1, the bands and page 2's other elements.
		if !reflect.DeepEqual(b.Pages[0], a.Pages[0]) || !reflect.DeepEqual(b.Bands, a.Bands) {
			t.Errorf("%s changed page 1 or the bands", c.command)
		}
		others := func(d *template.Document) []template.Element {
			var out []template.Element
			for _, el := range d.Pages[1].Elements {
				if string(el.ID) != target {
					out = append(out, el)
				}
			}
			return out
		}
		if !reflect.DeepEqual(others(b), others(a)) {
			t.Errorf("%s changed another page-2 element", c.command)
		}
		if _, err := engine.Undo(); err != nil {
			t.Fatal(err)
		}
		if undone, _, _ := engine.Serialize(); !bytes.Equal(undone, before) {
			t.Fatalf("one undo did not restore the bytes before %s", c.command)
		}
		// Re-apply so the next command runs on the changed document, except
		// the delete, which is last.
		if _, err := engine.Redo(); err != nil {
			t.Fatal(err)
		}
	}
}
