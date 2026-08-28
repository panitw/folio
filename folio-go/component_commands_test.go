package folio

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func componentTemplate(t *testing.T) *Template {
	t.Helper()
	b, err := os.ReadFile("testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := ParseTemplate(b)
	if err != nil {
		t.Fatal(err)
	}
	return tpl
}

func newProjectedComponent(t *testing.T, before, after CanvasProjection) CanvasComponent {
	t.Helper()
	known := map[string]bool{}
	for _, component := range before.Components {
		known[component.ID] = true
	}
	for _, component := range after.Components {
		if !known[component.ID] {
			return component
		}
	}
	t.Fatal("command projection did not add a component")
	return CanvasComponent{}
}

func pointLiteral(value int64) string {
	whole, fraction := value/1000, value%1000
	if fraction == 0 {
		return strconv.FormatInt(whole, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + strings.TrimRight(fmt.Sprintf("%03d", fraction), "0")
}

func TestComponentCommandsCreateAllClosedKindsAndKeepOrder(t *testing.T) {
	tpl := componentTemplate(t)
	for _, kind := range []string{"text", "image", "table", "line", "rect"} {
		before, err := Canvas(tpl)
		if err != nil {
			t.Fatal(err)
		}
		projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"`+kind+`","band":"content","x":12,"y":12,"width":72,"height":24,"snap":false}`))
		if err != nil {
			t.Fatalf("create %s: %v", kind, err)
		}
		component := newProjectedComponent(t, before, projection)
		if component.Type != kind || component.Band != "content" || component.ID == "" {
			t.Fatalf("create %s projection = %#v", kind, component)
		}
		if kind == "table" && (component.Resizable || component.Width != 0 || component.Height != 12000) {
			t.Fatalf("table projection must be derived and non-resizable: %#v", component)
		}
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"widget","band":"content","x":12,"y":12,"width":72,"height":24,"snap":false}`)); err == nil {
		t.Fatal("sixth component kind unexpectedly succeeded")
	}
}

func TestComponentCommandsSnapContainAndFailureAreTransactional(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	beforeProjection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"text","band":"content","x":3,"y":3,"width":72,"height":24,"snap":true}`)); err != nil {
		t.Fatal(err)
	}
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	created := newProjectedComponent(t, beforeProjection, projection)
	if created.X != 6000 || created.Y != 6000 {
		t.Fatalf("snap = %#v, want 6000 millipoints", created)
	}
	afterCreate, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, afterCreate) {
		t.Fatal("successful component command did not change canonical bytes")
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+created.ID+`","x":999999,"y":0,"snap":false}`)); err == nil {
		t.Fatal("out-of-band move unexpectedly succeeded")
	}
	afterFailure, err := SerializeTemplate(tpl)
	if err != nil || !bytes.Equal(afterCreate, afterFailure) {
		t.Fatalf("failed command changed canonical bytes: %v", err)
	}
}

func TestComponentCommandsRejectTableResizeAndPreserveTableGeometry(t *testing.T) {
	tpl := componentTemplate(t)
	beforeProjection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	table := newProjectedComponent(t, beforeProjection, projection)
	before, _ := SerializeTemplate(tpl)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"resizeComponent","version":1,"id":"`+table.ID+`","width":72,"height":24,"snap":false}`)); err == nil || !strings.Contains(err.Error(), "derived geometry") {
		t.Fatalf("table resize error = %v", err)
	}
	after, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, after) || strings.Contains(string(after), `"width": 72`) {
		t.Fatalf("table resize changed canonical state: %s", after)
	}
}

func TestDropComponentUsesGoHalfOpenBandHitTesting(t *testing.T) {
	tpl := componentTemplate(t)
	canvas, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []struct {
		band string
		y    int64
	}{
		{"pageHeader", canvas.Bands[0].Y},
		{"content", canvas.Bands[1].Y},
		{"pageFooter", canvas.Bands[2].Y},
	} {
		before, _ := Canvas(tpl)
		command := []byte(`{"kind":"dropComponent","version":1,"type":"text","x":36,"y":` + pointLiteral(want.y) + `,"snap":false}`)
		after, err := ApplyComponentCommand(tpl, command)
		if err != nil {
			t.Fatalf("drop %s: %v", want.band, err)
		}
		if got := newProjectedComponent(t, before, after).Band; got != want.band {
			t.Fatalf("drop at %d resolved %s, want %s", want.y, got, want.band)
		}
	}
	before, _ := SerializeTemplate(tpl)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"dropComponent","version":1,"type":"text","x":35.999,"y":36,"snap":false}`)); err == nil {
		t.Fatal("drop on the left page edge unexpectedly succeeded")
	}
	after, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected drop changed canonical bytes")
	}
}

func TestComponentMoveResizeDeleteAreExactAndMonotonic(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	createdProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	created := newProjectedComponent(t, before, createdProjection)
	moved, err := ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+created.ID+`","x":1.001,"y":2.002,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	component := newProjectedComponent(t, before, moved)
	if component.X != 1001 || component.Y != 2002 {
		t.Fatalf("move units = (%d,%d), want (1001,2002)", component.X, component.Y)
	}
	resized, err := ApplyComponentCommand(tpl, []byte(`{"kind":"resizeComponent","version":1,"id":"`+created.ID+`","width":73.003,"height":25.004,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	component = newProjectedComponent(t, before, resized)
	if component.Width != 73003 || component.Height != 25004 {
		t.Fatalf("resize units = (%d,%d), want (73003,25004)", component.Width, component.Height)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"deleteComponent","version":1,"id":"`+created.ID+`"}`)); err != nil {
		t.Fatal(err)
	}
	beforeNext, _ := Canvas(tpl)
	next, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"line","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	if newProjectedComponent(t, beforeNext, next).ID == created.ID {
		t.Fatal("component id was reused after deletion")
	}
}
