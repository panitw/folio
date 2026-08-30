package folio

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
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

func TestPlacedImageStartsEmptyAndSurvivesTheRoundTripAndRender(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	placed, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"image","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	// The box is placed but unfilled: nothing to paint, and no reason given,
	// because nothing has gone wrong — the author simply has not chosen a
	// file. That is what tells the designer to draw its empty placeholder
	// rather than one of ImageUnavailable's two failure texts.
	component := newProjectedComponent(t, before, placed)
	if component.Image != nil || component.ImageUnavailable != nil {
		t.Fatalf("placed image projected image=%#v unavailable=%s, want an empty box", component.Image, describeStringPtr(component.ImageUnavailable))
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	// No stowaway: an empty box embeds no bytes anywhere in the document.
	if !bytes.Contains(canonical, []byte(`"asset": null`)) {
		t.Fatalf("canonical bytes do not declare the empty asset: %s", canonical)
	}
	if !bytes.Contains(canonical, []byte(`"assets": {}`)) {
		t.Fatalf("placing an empty image added a document asset: %s", canonical)
	}
	// The canonical bytes still load, and the element is still an image.
	reloaded, err := ParseTemplate(canonical)
	if err != nil {
		t.Fatalf("a document with an empty image box did not load: %v", err)
	}
	again, err := SerializeTemplate(reloaded)
	if err != nil || !bytes.Equal(canonical, again) {
		t.Fatalf("empty image box did not round-trip: err=%v", err)
	}
	// Render draws nothing for it and completes without a caveat: an unfilled
	// box is an authoring state, not a defect in the document.
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "statement-1", "data.json"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := Render(reloaded, Data(data), Params(`{"generatedDate":"2026-08-27"}`), testShippedFontSet())
	if err != nil {
		t.Fatalf("a document with an empty image box did not render: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("empty image box reported diagnostics: %#v", result.Diagnostics)
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
	if err != nil || len(view.Columns) != 1 || view.Columns[0].Width != 72000 || view.Columns[0].Align != "left" || !view.Columns[0].RowFieldEditable {
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

func TestSetComponentBoundsMovesOriginAndSizeInOneCommand(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	createdProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":36,"y":36,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	created := newProjectedComponent(t, before, createdProjection)
	// A north-west drag: the origin and the size move together, which is the
	// whole reason this command exists next to move and resize.
	bounded, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setComponentBounds","version":1,"id":"`+created.ID+`","x":24.005,"y":12.006,"width":84.007,"height":48.008,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	component := newProjectedComponent(t, before, bounded)
	if component.X != 24005 || component.Y != 12006 || component.Width != 84007 || component.Height != 48008 {
		t.Fatalf("bounds units = (%d,%d,%d,%d), want (24005,12006,84007,48008)", component.X, component.Y, component.Width, component.Height)
	}
	// A TALL CONTENT Y, ACCEPTED. This seam moves origin and size together
	// and its refusal map below carries no y or height overflow probe at
	// all, so the clause Story 7.5 lifted had zero coverage here on either
	// side of the change. 2400pt is roughly three and a half windows down a
	// content band 679.89pt tall.
	tall, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setComponentBounds","version":1,"id":"`+created.ID+`","x":0,"y":2400,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatalf("bounds three windows below the band foot were refused: %v", err)
	}
	if component = newProjectedComponent(t, before, tall); component.Y != 2400000 {
		t.Fatalf("tall bounds y = %d, want 2400000", component.Y)
	}
	if roundTripped := reloadedComponent(t, tpl, created.ID); roundTripped.Y != 2400000 {
		t.Fatalf("canonical bytes carried y = %d, want 2400000", roundTripped.Y)
	}
	canonical, _ := SerializeTemplate(tpl)
	for name, command := range map[string]string{
		"missing height":  `{"kind":"setComponentBounds","version":1,"id":"` + created.ID + `","x":0,"y":0,"width":72,"snap":false}`,
		"zero width":      `{"kind":"setComponentBounds","version":1,"id":"` + created.ID + `","x":0,"y":0,"width":0,"height":24,"snap":false}`,
		"outside theband": `{"kind":"setComponentBounds","version":1,"id":"` + created.ID + `","x":-1,"y":0,"width":72,"height":24,"snap":false}`,
		"past band width": `{"kind":"setComponentBounds","version":1,"id":"` + created.ID + `","x":0,"y":0,"width":100000,"height":24,"snap":false}`,
		"unknown id":      `{"kind":"setComponentBounds","version":1,"id":"nope","x":0,"y":0,"width":72,"height":24,"snap":false}`,
	} {
		if _, err := ApplyComponentCommand(tpl, []byte(command)); err == nil {
			t.Fatalf("%s unexpectedly succeeded", name)
		}
		if after, _ := SerializeTemplate(tpl); !bytes.Equal(canonical, after) {
			t.Fatalf("rejected %s changed canonical bytes", name)
		}
	}
}

func TestSnapDoesNotPushAnEdgeDragOutOfItsBand(t *testing.T) {
	tpl := componentTemplate(t)
	before, _ := Canvas(tpl)
	createdProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	created := newProjectedComponent(t, before, createdProjection)
	var band CanvasBand
	for _, candidate := range createdProjection.Bands {
		if candidate.Name == "content" {
			band = candidate
		}
	}
	literal := func(millipoints int64) string {
		return fmt.Sprintf("%d.%03d", millipoints/1000, millipoints%1000)
	}
	// The far edge of the band, to the millipoint. Nearest-grid snapping can
	// round this away from the band, and rounding alone must not turn a legal
	// drag into a refusal the designer shows as a bounce back.
	edgeX, edgeY := band.Width-created.Width, band.Height-created.Height
	moved, err := ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+created.ID+`","x":`+literal(edgeX)+`,"y":`+literal(edgeY)+`,"snap":true}`))
	if err != nil {
		t.Fatalf("edge move with snapping was refused: %v", err)
	}
	component := newProjectedComponent(t, before, moved)
	if component.X%GridIncrement != 0 || component.Y%GridIncrement != 0 {
		t.Fatalf("pulled-back origin = (%d,%d), want grid multiples", component.X, component.Y)
	}
	// THE X HALF ONLY, since Story 7.5. The pull-back is a rescue of the
	// grid's own rounding on the axis that still has an edge to be pulled
	// back to; the content band has no bottom edge any more, so a snapped Y
	// past `edgeY` is the position the author asked for and stays there.
	if component.X+component.Width > band.Width {
		t.Fatalf("pulled-back geometry (%d,%d,%d,%d) leaves band width %d", component.X, component.Y, component.Width, component.Height, band.Width)
	}
	if edgeX-component.X >= GridIncrement {
		t.Fatalf("pull-back moved x %d further than one grid step from %d", component.X, edgeX)
	}
	// Same for a bounds drag that lands its far edges exactly on the band.
	bounded, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setComponentBounds","version":1,"id":"`+created.ID+`","x":`+literal(edgeX)+`,"y":`+literal(edgeY)+`,"width":`+literal(created.Width)+`,"height":`+literal(created.Height)+`,"snap":true}`))
	if err != nil {
		t.Fatalf("edge bounds with snapping was refused: %v", err)
	}
	component = newProjectedComponent(t, before, bounded)
	if component.X+component.Width > band.Width {
		t.Fatalf("pulled-back bounds (%d,%d,%d,%d) leave band width %d", component.X, component.Y, component.Width, component.Height, band.Width)
	}
	// The pull-back rescues the grid's own rounding and nothing else: a
	// caller asking for geometry a whole grid step outside is still refused.
	canonical, _ := SerializeTemplate(tpl)
	for name, command := range map[string]string{
		"a grid step past the right edge": `{"kind":"moveComponent","version":1,"id":"` + created.ID + `","x":` + literal(edgeX+GridIncrement) + `,"y":0,"snap":true}`,
		"far past the right edge":         `{"kind":"setComponentBounds","version":1,"id":"` + created.ID + `","x":` + literal(edgeX+100*GridIncrement) + `,"y":0,"width":72,"height":24,"snap":true}`,
	} {
		if _, err := ApplyComponentCommand(tpl, []byte(command)); err == nil {
			t.Fatalf("%s unexpectedly succeeded", name)
		}
		if after, _ := SerializeTemplate(tpl); !bytes.Equal(canonical, after) {
			t.Fatalf("rejected %s changed canonical bytes", name)
		}
	}
	// A grid step past what used to be the bottom edge, which was in that
	// refusal map until Story 7.5 and is now an ordinary placement. The Y arm
	// of the pull-back assertions above went vacuous when the cap lifted —
	// nothing pulls a content Y back any more — so THIS is what carries the
	// Y axis's discriminating power now: far below the foot of page one, the
	// drag is accepted, still snaps to the grid, and is what the canonical
	// bytes carry back on the next load.
	far := band.Height*3 + 4321
	dropped, err := ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+created.ID+`","x":0,"y":`+literal(far)+`,"snap":true}`))
	if err != nil {
		t.Fatalf("a content drag three windows below the band was refused: %v", err)
	}
	component = newProjectedComponent(t, before, dropped)
	if component.Y%GridIncrement != 0 {
		t.Fatalf("snapped y = %d, want a grid multiple", component.Y)
	}
	if component.Y <= band.Height {
		t.Fatalf("snapped y = %d, want a position below the one-window band height %d", component.Y, band.Height)
	}
	roundTripped := reloadedComponent(t, tpl, created.ID)
	if roundTripped.Y != component.Y {
		t.Fatalf("canonical bytes carried y = %d, want the placed %d", roundTripped.Y, component.Y)
	}
}

// reloadedComponent serializes tpl, parses the bytes back and returns the
// named component as the fresh projection sees it — a full round trip through
// the canonical form, which is the only way to show that a placement PERSISTS
// rather than merely being accepted by one command.
func reloadedComponent(t *testing.T, tpl *Template, id string) CanvasComponent {
	t.Helper()
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := ParseTemplate(canonical)
	if err != nil {
		t.Fatalf("canonical bytes did not load back: %v", err)
	}
	projection, err := Canvas(reloaded)
	if err != nil {
		t.Fatal(err)
	}
	for _, component := range projection.Components {
		if component.ID == id {
			return component
		}
	}
	t.Fatalf("component %q is absent from the reloaded projection", id)
	return CanvasComponent{}
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

// projectedBands is the three CanvasBands of a template, by name, so a test
// can probe containComponent with the same rectangles the command path hands
// it rather than with numbers it invented.
func projectedBands(t *testing.T, tpl *Template) map[string]CanvasBand {
	t.Helper()
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	bands := make(map[string]CanvasBand, len(projection.Bands))
	for _, band := range projection.Bands {
		bands[band.Name] = band
	}
	if len(bands) != 3 {
		t.Fatalf("projection carries %d bands, want 3", len(bands))
	}
	return bands
}

// TestBandContainmentRefusalsCarryTheirExactMessages CHARACTERIZES the one
// message every band-extent refusal in the designer command path produces,
// by FULL-STRING equality, before Story 7.5 splits the predicate that
// produces it.
//
// Until this test existed, `grep "must stay within" --include="*_test.go"`
// returned nothing: "the message is unchanged" was a claim no test could
// contradict. A story that splits first and asserts afterwards is asserting
// whatever it happened to produce, so this lands first and on purpose.
//
// WHY FULL-STRING EQUALITY AND NEVER strings.Contains. This same file's
// unrelated column refusal reads "footerOf must stay within the table
// collection" — a substring assertion on "must stay within" is satisfied by
// a message about table collections, which is the opposite of a
// characterization.
func TestBandContainmentRefusalsCarryTheirExactMessages(t *testing.T) {
	tpl := componentTemplate(t)
	bands := projectedBands(t, tpl)
	// Spelled out here rather than assembled from the production format
	// string: a characterization that reuses the code it characterizes
	// asserts nothing at all.
	want := map[string]string{
		"pageHeader": "folio: component geometry must stay within pageHeader",
		"content":    "folio: component geometry must stay within content",
		"pageFooter": "folio: component geometry must stay within pageFooter",
	}
	exactly := func(name, probe string, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s in %s was accepted; want the refusal %q", probe, name, want[name])
		}
		if err.Error() != want[name] {
			t.Fatalf("%s in %s = %q, want exactly %q", probe, name, err.Error(), want[name])
		}
	}
	type probe struct {
		name                string
		x, y, width, height geom.Length
	}
	// REPRESENTATIONAL refusals. A negative coordinate is not a statement
	// about how tall a band is, so it is refused in every band and stays
	// refused in every band.
	for _, name := range []string{"pageHeader", "content", "pageFooter"} {
		for _, p := range []probe{
			{"negative x", -1, 0, 6000, 6000},
			{"negative y", 0, -1, 6000, 6000},
			{"negative width", 0, 0, -1, 6000},
			{"negative height", 0, 0, 6000, -1},
		} {
			exactly(name, p.name, containComponent(bands[name], p.x, p.y, p.width, p.height))
		}
	}
	// The HORIZONTAL cap, in the content band: a column is unbounded
	// vertically, never horizontally.
	content := bands["content"]
	exactly("content", "x past the band width", containComponent(content, geom.Length(content.Width)+1, 0, 6000, 6000))
	exactly("content", "width past the band width", containComponent(content, 0, 0, geom.Length(content.Width)+1, 6000))
	// The BAND-CAPACITY refusals, in the two repeating bands. A page header
	// is exactly one page tall because that is what repeating means.
	for _, name := range []string{"pageHeader", "pageFooter"} {
		band := bands[name]
		exactly(name, "y past the band height", containComponent(band, 0, geom.Length(band.Height)+1, 6000, 6000))
		exactly(name, "height past the band height", containComponent(band, 0, 0, 6000, geom.Length(band.Height)+1))
	}
	// And the clause Story 7.5 LIFTED, asserted from the other side. This
	// test landed with these two lines reading `exactly(...)`, red-proved,
	// and one commit later they read this — which is the whole reason they
	// were written first. The content band is a COLUMN: a Y past one
	// window's worth of height is a position on a later page, not geometry
	// outside the document.
	if err := containComponent(content, 0, geom.Length(content.Height)*11, 6000, 6000); err != nil {
		t.Fatalf("a content y eleven windows down = %v, want acceptance", err)
	}
	if err := containComponent(content, 0, 0, 6000, geom.Length(content.Height)*11); err != nil {
		t.Fatalf("a content box eleven windows tall = %v, want acceptance", err)
	}
}

// TestJavaScriptSafeGeometryBoundRefusalsCarryTheirExactMessages is the
// SEPARATE, SURVIVING upper bound: it is a different path, with a different
// message, upstream of containComponent, and Story 7.5's split does not
// touch it. After the lift it is the only thing bounding a content
// component's Y, so it is characterized here for the same reason as the
// band messages — nothing asserted it before.
func TestJavaScriptSafeGeometryBoundRefusalsCarryTheirExactMessages(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	createdProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	created := newProjectedComponent(t, before, createdProjection)
	// One millipoint past Number.MAX_SAFE_INTEGER, as the command's own
	// three-decimal point literal.
	const past = "9007199254740.992"
	_, err = ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+created.ID+`","x":0,"y":`+past+`,"snap":false}`))
	if want := "folio: component.y: y exceeds the JavaScript-safe geometry bound"; err == nil || err.Error() != want {
		t.Fatalf("move past the JavaScript-safe bound = %v, want exactly %q", err, want)
	}
	// The projection-time sibling of the same bound, which is a different
	// message and must stay one: the command path refuses first, so this is
	// the backstop that keeps an unrepresentable coordinate from reaching
	// the JSON/JS boundary at all.
	stranded := componentTemplate(t)
	if len(stranded.doc.Bands.Content.Elements) == 0 {
		t.Fatal("fixture has no content element to strand")
	}
	stranded.doc.Bands.Content.Elements[0].X = geom.Length(MaxCanvasMillipoints) + 1
	_, err = Canvas(stranded)
	if want := "folio: component exceeds the JavaScript-safe geometry bound"; err == nil || err.Error() != want {
		t.Fatalf("projection of an unrepresentable coordinate = %v, want exactly %q", err, want)
	}
}

// TestTheColumnLiftIsExercisedAtTheCommandSurface closes the gap between the
// clause Story 7.5 moved and the surface the story is ABOUT.
//
// Two separate weaknesses live here, and neither shows up as a red test:
//
//  1. TestBandContainmentRefusalsCarryTheirExactMessages calls
//     containComponent DIRECTLY. That is the right shape for characterizing a
//     predicate, but the story's expectations are written at the command
//     surface ("setComponentBounds / moveComponent / createComponent"), and
//     the refusal an author actually sees is a *ComponentCommandError whose
//     Message and DataPath are set by the eleven call sites, not by the
//     predicate. Only the pageHeader property path asserted that wrapping.
//
//  2. containEdgeY has FOUR call sites and only one of them — moveComponent's
//     y — was reachable from any test. Reverting the other three (
//     dropComponent's y, and setComponentBounds' height and y) to the
//     unconditional containEdge leaves the ENTIRE Go suite green, which was
//     measured, not assumed. Those three are the sites whose pre-clamp gate
//     the lift WIDENS: a drag that was refused outright before now passes the
//     probe, so a containEdge left behind would quietly pull the component
//     back onto page one instead of refusing it — and in the bounds case
//     would also collapse its height to zero, because the positivity guard
//     has already run by then.
func TestTheColumnLiftIsExercisedAtTheCommandSurface(t *testing.T) {
	tpl := componentTemplate(t)
	bands := projectedBands(t, tpl)
	content, header, footer := bands["content"], bands["pageHeader"], bands["pageFooter"]
	before, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	refusedAs := func(name string, err error, wantMessage string) {
		t.Helper()
		var failure *ComponentCommandError
		if !errors.As(err, &failure) {
			t.Fatalf("%s = %v, want a component command failure", name, err)
		}
		if failure.Message != wantMessage {
			t.Fatalf("%s message = %q, want exactly %q", name, failure.Message, wantMessage)
		}
		if failure.DataPath != "component.geometry" {
			t.Fatalf("%s data path = %q, want component.geometry", name, failure.DataPath)
		}
	}

	// MATRIX ROW 1 NAMES createComponent, and nothing created a tall content
	// component: the largest content y anywhere else in these tests is 40pt.
	tall := content.Height*4 + 4321
	createdProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":`+pointLiteral(tall)+`,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatalf("createComponent four windows below the content top was refused: %v", err)
	}
	created := newProjectedComponent(t, before, createdProjection)
	if created.Y != tall {
		t.Fatalf("created y = %d, want the requested %d", created.Y, tall)
	}
	if roundTripped := reloadedComponent(t, tpl, created.ID); roundTripped.Y != tall {
		t.Fatalf("canonical bytes carried y = %d, want %d", roundTripped.Y, tall)
	}

	// THE COMMAND-SURFACE MESSAGES, by full-string equality, for the two
	// bands the property-panel test does not reach. The column is unbounded
	// vertically and never horizontally, so content still refuses on x.
	_, err = ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+created.ID+`","x":`+pointLiteral(content.Width+1000)+`,"y":0,"snap":false}`))
	refusedAs("a content x past the band width", err, "folio: component geometry must stay within content")

	footerProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"pageFooter","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	inFooter := newProjectedComponent(t, createdProjection, footerProjection)
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+inFooter.ID+`","x":0,"y":`+pointLiteral(footer.Height+1000)+`,"snap":false}`))
	refusedAs("a pageFooter y past the band height", err, "folio: component geometry must stay within pageFooter")
	if after, _ := SerializeTemplate(tpl); !bytes.Equal(canonical, after) {
		t.Fatal("a refused pageFooter move changed the canonical bytes")
	}

	// dropComponent's Y PULL-BACK, at a page point one point above the foot
	// of the content band. hitTestBand's rectangle is still one page tall, so
	// this is the lowest a drop can land — and the dropped box hangs past the
	// band foot, which is the whole point. With containEdgeY the component
	// keeps the grid position the author dropped it at; with the
	// unconditional containEdge it is pulled back inside page one instead.
	dropPageY := content.Y + content.Height - 1000
	beforeDrop, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	droppedProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"dropComponent","version":1,"type":"rect","x":`+pointLiteral(content.X)+`,"y":`+pointLiteral(dropPageY)+`,"snap":true}`))
	if err != nil {
		t.Fatalf("a snapped drop at the foot of the content band was refused: %v", err)
	}
	dropped := newProjectedComponent(t, beforeDrop, droppedProjection)
	if dropped.Y%GridIncrement != 0 {
		t.Fatalf("dropped y = %d, want a grid multiple", dropped.Y)
	}
	if dropped.Y+dropped.Height <= content.Height {
		t.Fatalf("dropped box (y %d + height %d) was pulled back inside the one-window band height %d", dropped.Y, dropped.Height, content.Height)
	}

	// setComponentBounds' HEIGHT and Y pull-backs, on a tall content
	// component with snapping on and no coordinate on the grid. The mutant
	// here is the expensive one: containEdge(height, band.Height-y) sees a
	// negative limit and floorToGrid returns 0, so the component is committed
	// with no height at all — the positivity guard at the top of the command
	// ran long before.
	boundsY := content.Height*3 + 500
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setComponentBounds","version":1,"id":"`+created.ID+`","x":0.5,"y":`+pointLiteral(boundsY)+`,"width":72.5,"height":24.5,"snap":true}`)); err != nil {
		t.Fatalf("a snapped bounds drag three windows down was refused: %v", err)
	}
	bounded := reloadedComponent(t, tpl, created.ID)
	if bounded.Height <= 0 {
		t.Fatalf("snapped bounds committed height = %d; the pull-back collapsed the component", bounded.Height)
	}
	if bounded.Height%GridIncrement != 0 || bounded.Y%GridIncrement != 0 {
		t.Fatalf("snapped bounds (y %d, height %d) are not grid multiples", bounded.Y, bounded.Height)
	}
	if bounded.Y <= content.Height {
		t.Fatalf("snapped bounds y = %d was pulled back inside the one-window band height %d", bounded.Y, content.Height)
	}

	// AND THE OTHER DIRECTION, which nothing exercised either: in a band that
	// DOES cap, containEdgeY must still clamp. A pageHeader component whose
	// far edge is off the grid snaps PAST the band foot, and the pull-back is
	// what rescues it — a containEdgeY that returned its input unchanged in
	// every band would make this command a refusal instead.
	beforeHeader, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	headerProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"pageHeader","x":0,"y":0,"width":72,"height":25,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	inHeader := newProjectedComponent(t, beforeHeader, headerProjection)
	headerEdgeY := header.Height - inHeader.Height
	if headerEdgeY%GridIncrement == 0 {
		t.Fatalf("this test needs an off-grid edge to say anything; edgeY %d is already a grid multiple", headerEdgeY)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"moveComponent","version":1,"id":"`+inHeader.ID+`","x":0,"y":`+pointLiteral(headerEdgeY)+`,"snap":true}`)); err != nil {
		t.Fatalf("an edge move inside the pageHeader was refused; the pull-back did not rescue the grid's rounding: %v", err)
	}
	movedInHeader := reloadedComponent(t, tpl, inHeader.ID)
	if movedInHeader.Y+movedInHeader.Height > header.Height {
		t.Fatalf("pageHeader component (y %d + height %d) left its band height %d; containEdgeY did not clamp", movedInHeader.Y, movedInHeader.Height, header.Height)
	}
	if headerEdgeY-movedInHeader.Y >= GridIncrement {
		t.Fatalf("pageHeader pull-back moved y to %d, more than one grid step from the edge %d", movedInHeader.Y, headerEdgeY)
	}
}
