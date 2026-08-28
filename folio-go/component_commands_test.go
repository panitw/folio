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

func TestBindComponentScalarOwnsRootExpressionAndPaintProjection(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}`))
	if err != nil {
		t.Fatal(err)
	}
	component := componentByID(t, projection, "e1")
	if component.Binding == nil || *component.Binding != "customer.name" {
		t.Fatalf("binding paint = %#v, want customer.name", component.Binding)
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil || bytes.Equal(before, canonical) || !bytes.Contains(canonical, []byte(`"value": "{{customer.name}}"`)) {
		t.Fatalf("canonical scalar binding = %s, err=%v", canonical, err)
	}
	for _, command := range [][]byte{
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["params","name"]}`),
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["page"]}`),
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e2","segments":["customer","name"]}`),
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["not-valid"]}`),
	} {
		if _, err := ApplyComponentCommand(tpl, command); err == nil {
			t.Fatalf("invalid scalar bind unexpectedly succeeded: %s", command)
		}
		after, err := SerializeTemplate(tpl)
		if err != nil || !bytes.Equal(canonical, after) {
			t.Fatalf("rejected scalar bind changed canonical bytes: %s, err=%v", command, err)
		}
	}
}

func TestBindComponentScalarPreservesDecodedSegmentsAndRejectsTypedBindings(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	// This is legal command grammar even though a picker would withhold an
	// observed collection. D-6.2.1 keeps sample runtime kind out of command
	// legality; AD-14 reports any incompatible runtime value later.
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["items"]}`)); err != nil {
		t.Fatalf("sample-independent collection-shaped path was rejected: %v", err)
	}
	collectionCanonical, err := SerializeTemplate(tpl)
	if err != nil || !bytes.Contains(collectionCanonical, []byte(`"value": "{{items}}"`)) {
		t.Fatalf("collection-shaped path was not canonically retained: %s, err=%v", collectionCanonical, err)
	}
	_, renderErr := Render(tpl, Data(`{"items":[],"transactions":[]}`), Params(`{}`), testShippedFontSet())
	if renderErr == nil || !strings.Contains(renderErr.Error(), "element e1") || !strings.Contains(renderErr.Error(), `"items"`) || !strings.Contains(renderErr.Error(), "array") {
		t.Fatalf("runtime collection kind must use the AD-14 located diagnostic, got %v", renderErr)
	}
	for _, command := range [][]byte{
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["a.b"]}`),
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["line\\nbreak"]}`),
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["สวัสดี"]}`),
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":[""]}`),
		[]byte(`{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["params","name"]}`),
		[]byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"{{customer.name}}"}}}`),
	} {
		if _, err := ApplyComponentCommand(tpl, command); err == nil {
			t.Fatalf("ambiguous or typed binding unexpectedly succeeded: %s", command)
		}
		after, err := SerializeTemplate(tpl)
		if err != nil || !bytes.Equal(collectionCanonical, after) {
			t.Fatalf("rejected command changed canonical bytes: %s, err=%v", command, err)
		}
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"literal text"}}}`)); err != nil {
		t.Fatalf("literal text edit was rejected: %v", err)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil || bytes.Equal(before, after) || !bytes.Contains(after, []byte(`"value": "literal text"`)) {
		t.Fatalf("literal edit was not canonical: %s, err=%v", after, err)
	}
}

func componentByID(t *testing.T, projection CanvasProjection, id string) CanvasComponent {
	t.Helper()
	for _, component := range projection.Components {
		if component.ID == id {
			return component
		}
	}
	t.Fatalf("component %q is absent from projection", id)
	return CanvasComponent{}
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

func TestTableColumnCommandsAreClosedCanonicalAndDerived(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	table := newProjectedComponent(t, before, projection)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":0}`)); err != nil {
		t.Fatal(err)
	}
	view, err := TableColumns(tpl, table.ID)
	if err != nil || len(view.Columns) != 1 || view.Columns[0].Width != 72000 || view.Columns[0].Align != "left" {
		t.Fatalf("column projection = %#v, err=%v", view, err)
	}
	column := view.Columns[0]
	for _, command := range [][]byte{
		[]byte(`{"kind":"updateTableColumn","version":1,"id":"` + table.ID + `","columnId":"` + column.ID + `","field":"header","value":"Amount"}`),
		[]byte(`{"kind":"updateTableColumn","version":1,"id":"` + table.ID + `","columnId":"` + column.ID + `","field":"width","value":96}`),
		[]byte(`{"kind":"updateTableColumn","version":1,"id":"` + table.ID + `","columnId":"` + column.ID + `","field":"align","value":"right"}`),
		[]byte(`{"kind":"addTableColumn","version":1,"id":"` + table.ID + `","index":1}`),
	} {
		if _, err := ApplyComponentCommand(tpl, command); err != nil {
			t.Fatalf("apply %s: %v", command, err)
		}
	}
	view, err = TableColumns(tpl, table.ID)
	if err != nil || len(view.Columns) != 2 || view.Columns[0].Header != "Amount" || view.Columns[0].Width != 96000 || view.Columns[0].Align != "right" {
		t.Fatalf("edited projection = %#v, err=%v", view, err)
	}
	canvas, err := Canvas(tpl)
	if err != nil || componentByID(t, canvas, table.ID).Width != 168000 {
		t.Fatalf("derived table width = %#v, err=%v", canvas, err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"moveTableColumn","version":1,"id":"`+table.ID+`","columnId":"`+view.Columns[1].ID+`","toIndex":0}`)); err != nil {
		t.Fatal(err)
	}
	view, _ = TableColumns(tpl, table.ID)
	if view.Columns[0].ID == column.ID {
		t.Fatal("move did not preserve engine ordered columns")
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil || bytes.Contains(canonical, []byte(`"tableWidth"`)) {
		t.Fatalf("canonical table geometry leaked a width: %s, err=%v", canonical, err)
	}
}

func TestTableColumnRejectionsDoNotMutate(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	table := newProjectedComponent(t, before, projection)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":0}`)); err != nil {
		t.Fatal(err)
	}
	view, _ := TableColumns(tpl, table.ID)
	canonical, _ := SerializeTemplate(tpl)
	for _, command := range [][]byte{
		[]byte(`{"kind":"updateTableColumn","version":1,"id":"` + table.ID + `","columnId":"` + view.Columns[0].ID + `","field":"width","value":0}`),
		[]byte(`{"kind":"updateTableColumn","version":1,"id":"` + table.ID + `","columnId":"` + view.Columns[0].ID + `","field":"align","value":"justify"}`),
		[]byte(`{"kind":"removeTableColumn","version":1,"id":"` + table.ID + `","columnId":"missing"}`),
		[]byte(`{"kind":"addTableColumn","version":1,"id":"` + table.ID + `","index":3}`),
	} {
		if _, err := ApplyComponentCommand(tpl, command); err == nil {
			t.Fatalf("invalid command succeeded: %s", command)
		}
		after, _ := SerializeTemplate(tpl)
		if !bytes.Equal(canonical, after) {
			t.Fatalf("rejection mutated canonical bytes: %s", command)
		}
	}
}

func TestTableColumnCommandsAreTransactionalAtThePublicSeam(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":500,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	table := newProjectedComponent(t, before, projection)
	canonical, _ := SerializeTemplate(tpl)
	// addTableColumn reaches containment only after adding the candidate column;
	// direct callers must still retain their original canonical template.
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":0}`)); err == nil {
		t.Fatal("out-of-band add unexpectedly succeeded")
	}
	after, _ := SerializeTemplate(tpl)
	if !bytes.Equal(canonical, after) {
		t.Fatal("rejected add mutated the public caller template")
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":0}`)); err == nil {
		t.Fatal("a rejected candidate consumed an id or changed later command behavior")
	}
}

func TestTableColumnProjectionCapRejectsThe129thCommandWithoutMutation(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	table := newProjectedComponent(t, before, projection)
	for index := 0; index < maxTableColumns; index++ {
		if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":`+strconv.Itoa(index)+`}`)); err != nil {
			t.Fatalf("add %d: %v", index+1, err)
		}
		view, err := TableColumns(tpl, table.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateTableColumn","version":1,"id":"`+table.ID+`","columnId":"`+view.Columns[index].ID+`","field":"width","value":0.001}`)); err != nil {
			t.Fatalf("shrink %d: %v", index+1, err)
		}
	}
	view, err := TableColumns(tpl, table.ID)
	if err != nil || len(view.Columns) != maxTableColumns {
		t.Fatalf("128-column projection = %#v, err=%v", view, err)
	}
	canonical, _ := SerializeTemplate(tpl)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":128}`)); err == nil {
		t.Fatal("129th editor-unprojectable column unexpectedly succeeded")
	}
	after, _ := SerializeTemplate(tpl)
	if !bytes.Equal(canonical, after) {
		t.Fatal("129th column rejection mutated canonical bytes")
	}
}

func TestTableDataBindingAndFooterCommandsAreCanonicalAndTransactional(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	table := newProjectedComponent(t, before, projection)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":0}`)); err != nil {
		t.Fatal(err)
	}
	view, err := TableColumns(tpl, table.ID)
	if err != nil {
		t.Fatal(err)
	}
	column := view.Columns[0]
	for _, command := range [][]byte{
		[]byte(`{"kind":"configureTableBinding","version":1,"id":"` + table.ID + `","collection":"transactions[]","alias":"transaction"}`),
		[]byte(`{"kind":"updateTableColumnBinding","version":1,"id":"` + table.ID + `","columnId":"` + column.ID + `","field":"amount"}`),
		[]byte(`{"kind":"updateTableColumnFooter","version":1,"id":"` + table.ID + `","columnId":"` + column.ID + `","footer":"sum","footerOf":"","footerFormat":""}`),
	} {
		if _, err := ApplyComponentCommand(tpl, command); err != nil {
			t.Fatalf("apply %s: %v", command, err)
		}
	}
	view, err = TableColumns(tpl, table.ID)
	if err != nil || view.Collection != "transactions[]" || view.Alias != "transaction" || view.Columns[0].Binding != "{{transaction.amount}}" || view.Columns[0].Footer != "sum" {
		t.Fatalf("data projection = %#v, err=%v", view, err)
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(canonical, []byte(`"as": "transaction"`)) || !bytes.Contains(canonical, []byte(`"bind": "{{transaction.amount}}"`)) || !bytes.Contains(canonical, []byte(`"footer": "sum"`)) {
		t.Fatalf("canonical data table omitted configured fields: %s", canonical)
	}
	if _, err := ParseTemplate(canonical); err != nil {
		t.Fatalf("canonical reload: %v", err)
	}
	for _, rejected := range [][]byte{
		[]byte(`{"kind":"configureTableBinding","version":1,"id":"` + table.ID + `","collection":"params.items[]","alias":"row"}`),
		[]byte(`{"kind":"updateTableColumnBinding","version":1,"id":"` + table.ID + `","columnId":"` + column.ID + `","field":"bare row"}`),
		[]byte(`{"kind":"updateTableColumnFooter","version":1,"id":"` + table.ID + `","columnId":"` + column.ID + `","footer":"count","footerOf":"transactions.amount","footerFormat":""}`),
	} {
		if _, err := ApplyComponentCommand(tpl, rejected); err == nil {
			t.Fatalf("rejected command succeeded: %s", rejected)
		}
		after, _ := SerializeTemplate(tpl)
		if !bytes.Equal(canonical, after) {
			t.Fatalf("rejection changed canonical bytes: %s", rejected)
		}
	}
}

func TestTableAliasMigrationReservedRootsAndStrictEnvelope(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	table := newProjectedComponent(t, before, projection)
	for index := 0; index < 3; index++ {
		if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":`+strconv.Itoa(index)+`}`)); err != nil {
			t.Fatal(err)
		}
	}
	view, _ := TableColumns(tpl, table.ID)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"configureTableBinding","version":1,"id":"`+table.ID+`","collection":"transactions[]","alias":"transaction"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateTableColumnBinding","version":1,"id":"`+table.ID+`","columnId":"`+view.Columns[0].ID+`","field":"params.value"}`)); err != nil {
		t.Fatal(err)
	}
	// A formatNumber row expression is a legal persisted source and must migrate
	// without a browser parser; params remains a distinct root and stays put.
	_, _, _, tableElement, err := findComponent(tpl, table.ID)
	if err != nil {
		t.Fatal(err)
	}
	tableElement.Table.Value.Columns[1].Bind = `{{formatNumber(transaction.amount, "#,##0.00")}}`
	tableElement.Table.Value.Columns[2].Bind = `{{params.value}}`
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"configureTableBinding","version":1,"id":"`+table.ID+`","collection":"transactions[]","alias":"item"}`)); err != nil {
		t.Fatal(err)
	}
	view, err = TableColumns(tpl, table.ID)
	if err != nil || view.Alias != "item" || view.Columns[0].Binding != "{{item.params.value}}" || view.Columns[1].Binding != `{{formatNumber(item.amount, "#,##0.00")}}` || view.Columns[2].Binding != "{{params.value}}" {
		t.Fatalf("migrated projection = %#v, err=%v", view, err)
	}
	canonical, _ := SerializeTemplate(tpl)
	for _, alias := range []string{"params", "page", "pages"} {
		if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"configureTableBinding","version":1,"id":"`+table.ID+`","collection":"transactions[]","alias":"`+alias+`"}`)); err == nil {
			t.Fatalf("reserved alias %q succeeded", alias)
		}
		after, _ := SerializeTemplate(tpl)
		if !bytes.Equal(canonical, after) {
			t.Fatalf("reserved alias %q mutated bytes", alias)
		}
	}
	if _, err := ApplyComponentCommand(tpl, append([]byte(`{"kind":"configureTableBinding","version":1,"id":"`+table.ID+`","collection":"transactions[]","alias":"sale"}`), []byte(` {}`)...)); err == nil {
		t.Fatal("concatenated command succeeded")
	}
	if _, err := ParseTemplate(canonical); err != nil {
		t.Fatalf("migrated reload: %v", err)
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
