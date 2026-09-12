package folio

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
)

func groupFixture(t *testing.T) (*Template, []string) {
	t.Helper()
	tpl := componentTemplate(t)
	tpl.doc.Bands.PageHeader.Elements = nil
	tpl.doc.Bands.PageFooter.Elements = nil
	tpl.doc.Bands.Content.Elements = nil
	ids := []string{}
	for index, kind := range []string{"text", "rect", "image", "line", "table"} {
		before, _ := Canvas(tpl)
		after, err := ApplyComponentCommand(tpl, []byte(fmt.Sprintf(`{"kind":"createComponent","version":1,"type":%q,"band":"content","x":%d.125,"y":%d.225,"width":24,"height":12,"snap":false}`, kind, 13+index*40, 10+index*30)))
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, newProjectedComponent(t, before, after).ID)
	}
	return tpl, ids
}

func moveIntent(ids []string, ref string, dx, dy string, snap bool) []byte {
	list := ""
	for i, id := range ids {
		if i > 0 {
			list += ","
		}
		list += fmt.Sprintf("%q", id)
	}
	return []byte(fmt.Sprintf(`{"kind":"moveComponents","version":1,"ids":[%s],"referenceId":%q,"dx":%s,"dy":%s,"snap":%t,"expectedRevision":1}`, list, ref, dx, dy, snap))
}

func TestBulkBorderEdgeClearRemovesNullAndRetainsOtherProperties(t *testing.T) {
	tpl, ids := groupFixture(t)
	for _, id := range ids[:2] {
		_, _, _, element, _ := findComponent(tpl, id)
		element.Style = template.Presence[template.Style]{Set: true, Value: template.Style{
			Border:     template.Presence[template.Border]{Set: true, Null: true},
			Background: template.Presence[string]{Set: true, Value: "#123456"},
		}}
	}
	command := []byte(fmt.Sprintf(`{"kind":"updateComponentProperties","version":1,"ids":[%q,%q],"changes":{"borderEdges":{"op":"clear"}}}`, ids[0], ids[1]))
	projection, err := ApplyComponentCommand(tpl, command)
	if err != nil {
		t.Fatal(err)
	}
	for _, component := range projection.Components[:2] {
		if component.Authored.BorderEdges.State != "absent" || component.Authored.Background.Value == nil || *component.Authored.Background.Value != "#123456" {
			t.Fatalf("clear did not remove null border while preserving background: %+v", component.Authored)
		}
	}
}

func TestGroupMoveAllKindsCommonDeltaAndReadOnlyPreview(t *testing.T) {
	for _, snap := range []bool{false, true} {
		tpl, ids := groupFixture(t)
		before, _ := Canvas(tpl)
		original, _ := SerializeTemplate(tpl)
		command := moveIntent(ids, ids[0], "1.126", "2.227", snap)
		preview, err := PreviewComponentMove(tpl, command)
		if err != nil {
			t.Fatal(err)
		}
		untouched, _ := SerializeTemplate(tpl)
		if !bytes.Equal(original, untouched) {
			t.Fatal("preview mutated template")
		}
		want := ComponentMove{1126, 2227}
		if snap {
			want = ComponentMove{-1125, 1775}
		}
		if preview != want {
			t.Fatalf("snap=%t preview=%+v want=%+v", snap, preview, want)
		}
		after, err := ApplyComponentCommand(tpl, command)
		if err != nil {
			t.Fatal(err)
		}
		for i, old := range before.Components {
			moved := after.Components[i]
			if moved.X-old.X != preview.DX || moved.Y-old.Y != preview.DY {
				t.Fatalf("different member delta: %+v -> %+v", old, moved)
			}
			moved.X, moved.Y = old.X, old.Y
			if !reflect.DeepEqual(moved, old) {
				t.Fatalf("movement changed non-position properties of %s", old.ID)
			}
		}
	}
}

func TestGroupMoveAtomicRefusalsAndZero(t *testing.T) {
	tpl, ids := groupFixture(t)
	before, _ := SerializeTemplate(tpl)
	for _, command := range [][]byte{
		moveIntent(append(append([]string{}, ids...), "ezmissing"), ids[0], "5", "5", false),
		moveIntent([]string{ids[0], ids[0]}, ids[0], "5", "5", false),
		moveIntent(ids, "ezmissing", "5", "5", false),
		moveIntent(ids, ids[0], "9007199254740.992", "5", false),
		bytes.Replace(moveIntent(ids, ids[0], "5", "5", false), []byte(`"expectedRevision":1`), []byte(`"expectedRevision":null`), 1),
	} {
		if _, err := ApplyComponentCommand(tpl, command); err == nil {
			t.Fatalf("invalid group accepted: %s", command)
		}
		after, _ := SerializeTemplate(tpl)
		if !bytes.Equal(before, after) {
			t.Fatal("refusal partially mutated group")
		}
	}
	if _, err := ApplyComponentCommand(tpl, moveIntent(ids, ids[0], "0", "0", true)); err != nil {
		t.Fatal(err)
	}
	after, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, after) {
		t.Fatal("zero gesture snapped off-grid originals")
	}
}

func TestGroupMoveIntersectedBoundsAndGridFallback(t *testing.T) {
	tpl, ids := groupFixture(t)
	// Establish a narrow feasible interval using opposite horizontal edges.
	_, band, _, first, _ := findComponent(tpl, ids[0])
	first.X = 125
	_, _, _, last, _ := findComponent(tpl, ids[1])
	lastWidth, _ := projectedSize(*last)
	last.X = geom.Length(band.Width) - lastWidth - 125
	_, _, _, reference, _ := findComponent(tpl, ids[2])
	reference.X = 10125
	move, err := PreviewComponentMove(tpl, moveIntent(ids, ids[2], "999", "0", true))
	if err != nil || move.DX != 125 {
		t.Fatalf("no feasible grid fallback: %+v, %v", move, err)
	}
	reference.X = 11900
	move, err = PreviewComponentMove(tpl, moveIntent(ids, ids[2], "-999", "0", true))
	if err != nil || move.DX != 100 {
		t.Fatalf("feasible grid point must win over fallback: %+v, %v", move, err)
	}
	// Header height is the first vertical limit; content remains a column.
	before, _ := Canvas(tpl)
	after, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"pageHeader","x":10,"y":4,"width":24,"height":12,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	header := newProjectedComponent(t, before, after)
	_, headerBand, _, _, _ := findComponent(tpl, header.ID)
	move, err = PreviewComponentMove(tpl, moveIntent(append(ids, header.ID), ids[0], "0", "999999", false))
	if err != nil || move.DY != headerBand.Height-header.Y-header.Height {
		t.Fatalf("header intersection: %+v, %v", move, err)
	}
	move, err = PreviewComponentMove(tpl, moveIntent(ids, ids[0], "0", "999999", false))
	if err != nil || move.DY != 999999000 {
		t.Fatalf("content was vertically capped: %+v, %v", move, err)
	}
}

func TestAuthoredInspectorRetainsNullFalseZeroAndHiddenBorders(t *testing.T) {
	tpl, ids := groupFixture(t)
	_, _, _, element, _ := findComponent(tpl, ids[1])
	element.VisibleIf = template.Presence[string]{Set: true, Value: ""}
	element.Style = template.Presence[template.Style]{Set: true, Value: template.Style{
		Background: template.Presence[string]{Set: true, Null: true},
		FontSize:   template.Presence[geom.Length]{Set: true, Value: -1000},
		Border: template.Presence[template.Border]{Set: true, Value: template.Border{
			Width: template.Presence[geom.Length]{Set: true, Value: 0}, Color: template.Presence[string]{Set: true, Value: "#123456"}, Edges: template.Presence[[]string]{Set: true, Value: []string{}},
		}},
	}}
	_, _, _, text, err := findComponent(tpl, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	text.Style.Value.Bold = template.Presence[bool]{Set: true, Value: false}
	projected, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	rect := projected.Components[1]
	if rect.BorderColor != nil || rect.BorderWidth != nil {
		t.Fatal("nonpainting border leaked into paint")
	}
	if rect.Authored.VisibleIf.Value == nil || *rect.Authored.VisibleIf.Value != "" || rect.Authored.FontSize.State != "absent" || rect.Authored.BorderWidth.Value == nil || *rect.Authored.BorderWidth.Value != 0 || *rect.Authored.BorderColor.Value != "#123456" || rect.Authored.BorderEdges.Value == nil || len(*rect.Authored.BorderEdges.Value) != 0 || rect.Authored.Background.State != "null" {
		t.Fatalf("lost authored evidence: %+v", rect.Authored)
	}
	if projected.Components[0].Authored.Bold.Value == nil || *projected.Components[0].Authored.Bold.Value {
		t.Fatal("explicit false was lost")
	}
}

func windowMoveIntent(ids []string, ref, dx, dy string, snap bool) []byte {
	return bytes.Replace(moveIntent(ids, ref, dx, dy, snap), []byte(`"expectedRevision":1`), []byte(`"expectedRevision":1,"constrainToWindow":true`), 1)
}

func TestWindowMoveFullBoxSnapAndSharedDelta(t *testing.T) {
	for _, count := range []int{1, 5} {
		for _, snap := range []bool{false, true} {
			for _, travel := range []string{"-999999", "999999"} {
				tpl, ids := groupFixture(t)
				ids = ids[:count]
				before, _ := Canvas(tpl)
				original, _ := SerializeTemplate(tpl)
				command := windowMoveIntent(ids, ids[0], "0", travel, snap)
				preview, err := PreviewComponentMove(tpl, command)
				if err != nil {
					t.Fatal(err)
				}
				referenceY := before.Components[0].Y
				expected := -referenceY
				if travel == "999999" {
					expected = before.ContentWindowHeight
					for _, component := range before.Components[:count] {
						expected = min(expected, before.ContentWindowHeight-component.Y-component.Height)
					}
					if snap {
						expected = ((referenceY+expected)/int64(GridIncrement))*int64(GridIncrement) - referenceY
					}
				}
				if preview.DY != expected {
					t.Fatalf("count=%d snap=%t travel=%s: accepted %d want exact edge %d", count, snap, travel, preview.DY, expected)
				}
				unchanged, _ := SerializeTemplate(tpl)
				if !bytes.Equal(original, unchanged) {
					t.Fatal("preview mutated authored geometry")
				}
				after, err := ApplyComponentCommand(tpl, command)
				if err != nil {
					t.Fatal(err)
				}
				for i, old := range before.Components {
					moved := after.Components[i]
					if i >= count {
						if moved.X != old.X || moved.Y != old.Y {
							t.Fatal("unselected component moved")
						}
						continue
					}
					if moved.Y < 0 || moved.Y+moved.Height > before.ContentWindowHeight {
						t.Fatalf("box crossed window: %+v", moved)
					}
					if moved.X-old.X != preview.DX || moved.Y-old.Y != preview.DY {
						t.Fatal("preview/commit or shared delta diverged")
					}
				}
			}
		}
	}
}

func TestWindowMoveLaterPageAndExistingOverflow(t *testing.T) {
	tpl := parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON)
	initial, _ := CanvasWithTextPaint(tpl, testFontSet())
	created, err := ApplyComponentCommand(tpl, []byte(fmt.Sprintf(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":13.125,"y":%s,"width":24,"height":12,"snap":false}`, pointLiteral(initial.ContentWindowHeight*2+12000))))
	if err != nil {
		t.Fatal(err)
	}
	id := newProjectedComponent(t, initial, created).ID
	ids := []string{"", id}
	_, _, _, element, _ := findComponent(tpl, id)
	before, _ := CanvasWithTextPaint(tpl, testFontSet())
	members, _ := groupMemberIndex(tpl, testFontSet())
	member := members[id]
	if member.windowOrigin == 0 {
		t.Fatal("fixture did not paginate later component")
	}

	for _, travel := range []string{"-999999", "999999"} {
		move, err := PreviewComponentMove(tpl, windowMoveIntent([]string{ids[1]}, ids[1], "0", travel, false), testFontSet())
		if err != nil {
			t.Fatal(err)
		}
		y := int64(element.Y) + move.DY
		_, height := projectedSize(*element)
		if y < int64(member.windowOrigin) || y+int64(height) > int64(member.windowOrigin+member.windowHeight) {
			t.Fatalf("left later window: %d in %+v", y, before.ContentWindowOrigins)
		}
	}
	element.Y = 10000
	element.Height = template.Presence[geom.Length]{Set: true, Value: geom.Length(before.ContentWindowHeight + 10000)}
	for _, snap := range []bool{false, true} {
		move, err := PreviewComponentMove(tpl, windowMoveIntent([]string{ids[1]}, ids[1], "10", "999999", snap), testFontSet())
		if err != nil || move.DX == 0 || move.DY > 0 {
			t.Fatalf("overflow increased or horizontal movement blocked: %+v %v", move, err)
		}
		zero, err := PreviewComponentMove(tpl, windowMoveIntent([]string{ids[1]}, ids[1], "0", "0", snap), testFontSet())
		if err != nil || zero != (ComponentMove{}) {
			t.Fatalf("zero changed existing geometry: %+v %v", zero, err)
		}
	}
}

func TestWindowMoveStopsAtOverlappingNextOrigin(t *testing.T) {
	tpl := parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON)
	tpl.doc.Bands.Content.Elements = nil
	var lineID string
	for _, command := range []string{
		`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":0,"width":24,"height":12,"snap":false}`,
		`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":715,"width":24,"height":24,"snap":false}`,
		`{"kind":"createComponent","version":1,"type":"line","band":"content","x":40,"y":700,"width":24,"height":1,"snap":false}`,
	} {
		before, _ := Canvas(tpl)
		after, err := ApplyComponentCommand(tpl, []byte(command))
		if err != nil {
			t.Fatal(err)
		}
		lineID = newProjectedComponent(t, before, after).ID
	}
	projection, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.ContentWindowOrigins) < 2 || projection.ContentWindowOrigins[1] >= projection.ContentWindowHeight {
		t.Fatalf("fixture needs overlapping window: %+v", projection.ContentWindowOrigins)
	}
	for _, snap := range []bool{false, true} {
		command := windowMoveIntent([]string{lineID}, lineID, "0", "99999", snap)
		preview, err := PreviewComponentMove(tpl, command, testFontSet())
		if err != nil {
			t.Fatal(err)
		}
		if 700000+preview.DY+1000 > projection.ContentWindowOrigins[1] {
			t.Fatalf("crossed display seam: %+v", preview)
		}
	}
}
