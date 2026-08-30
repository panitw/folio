package folio

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/template"
)

func TestUpdateComponentPropertiesIsClosedCanonicalAndAtomic(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"x":{"op":"set","value":12.125},"fontSize":{"op":"set","value":10},"visibleIf":{"op":"set","value":"customer.active"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if projection.Components[0].ID == "" {
		t.Fatal("missing property projection")
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) || !strings.Contains(string(after), `"x": 12.125`) || !strings.Contains(string(after), `"visibleIf": "customer.active"`) {
		t.Fatalf("canonical property change missing: %s", after)
	}
	if _, err := ParseTemplate(after); err != nil {
		t.Fatalf("property command did not preserve load validation: %v", err)
	}
}

func TestCreatedTextAdoptsADeclaredFontChainSoItCanRender(t *testing.T) {
	tpl := componentTemplate(t)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"text","band":"content","x":40,"y":40,"width":200,"height":24,"snap":false}`)); err != nil {
		t.Fatal(err)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	// Render resolves a face through style.fontFamily and refuses text without
	// one, so a freshly placed element must already name a declared chain.
	if !strings.Contains(string(after), `"fontFamily"`) {
		t.Fatalf("created text names no font chain: %s", after)
	}
	if _, err := ParseTemplate(after); err != nil {
		t.Fatalf("created text did not preserve load validation: %v", err)
	}
}

func TestCreatedTextLeavesFontFamilyAbsentWhenNoChainIsDeclared(t *testing.T) {
	// Nothing to adopt is not the same as adopting something invented: the
	// element stays exactly as it was before this default existed.
	tpl, err := ParseTemplate([]byte(`{"assets":{},"bands":{"content":{"elements":[]},"pageFooter":{"elements":[],"height":40},"pageHeader":{"elements":[],"height":60}},"fonts":{},"locale":"en","nextId":1,"page":{"margin":{"bottom":36,"left":36,"right":36,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"1.0"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"text","band":"content","x":40,"y":40,"width":200,"height":24,"snap":false}`)); err != nil {
		t.Fatal(err)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(after), `"fontFamily"`) {
		t.Fatalf("created text invented a font chain: %s", after)
	}
}

func TestUpdateComponentPropertiesRejectsTableGeometryAndRollsBackBatch(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e2"],"changes":{"width":{"op":"set","value":72}}}`)); err == nil || !strings.Contains(err.Error(), "not editable") {
		t.Fatalf("table width error = %v", err)
	}
	afterTable, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, afterTable) {
		t.Fatal("rejected table width changed canonical bytes")
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1","missing"],"changes":{"x":{"op":"set","value":12}}}`)); err == nil {
		t.Fatal("bad multi-target unexpectedly succeeded")
	}
	afterBatch, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, afterBatch) {
		t.Fatal("bad multi-target partially changed canonical bytes")
	}
}

func TestUpdateComponentPropertiesRejectsStylePlaceholderAndClearIsAbsent(t *testing.T) {
	tpl := componentTemplate(t)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"{{customer.color}}"}}}`)); err == nil {
		t.Fatal("style placeholder unexpectedly succeeded")
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"fontSize":{"op":"clear"}}}`)); err != nil {
		t.Fatal(err)
	}
	after, _ := SerializeTemplate(tpl)
	if strings.Contains(string(after), `"fontSize": 12`) {
		t.Fatalf("clear did not serialize absent: %s", after)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"visibleIf":{"op":"null"}}}`)); err != nil {
		t.Fatal(err)
	}
	afterNull, _ := SerializeTemplate(tpl)
	if !strings.Contains(string(afterNull), `"visibleIf": null`) {
		t.Fatalf("explicit null was not preserved: %s", afterNull)
	}
}

func TestUpdateComponentPropertiesValidatesColourFontAndNullWithoutMutation(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"not-a-colour"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"borderColor":{"op":"set","value":"#12ZZ00"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"fontFamily":{"op":"set","value":"missing-chain"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"fontFamily":{"op":"null"}}}`,
		`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"bold":{"op":"null"}}}`,
	} {
		if _, err := ApplyComponentCommand(tpl, []byte(command)); err == nil {
			t.Fatalf("invalid property command succeeded: %s", command)
		}
		after, err := SerializeTemplate(tpl)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("invalid command mutated bytes: %v\n%s", err, after)
		}
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"#aBc123"},"fontFamily":{"op":"set","value":"body"}}}`)); err != nil {
		t.Fatal(err)
	}
}

func TestComponentPropertyProjectionIsBoundedAndTypeCoherent(t *testing.T) {
	tpl := componentTemplate(t)
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	// THE LEAK PROBE. The worked example carries only text and table
	// elements, so the disjunction below had nothing to be false about: it
	// asserted that a non-text/table component carries no typography over a
	// projection with no such component in it. A RECT element with every
	// typography key set — `lineSpacing` among them, projected for the first
	// time by Story 7.4 through the same type gate in applyCanvasStyle — is
	// added so removing that gate turns this test red.
	leaky := bodyTextDocument(t, "short", `{"fontFamily":"body","fontSize":12}`)
	element := &leaky.doc.Bands.Content.Elements[0]
	element.Type = template.ElementRect
	element.Value = template.Presence[string]{}
	element.Style.Value.LineSpacing = template.Presence[int64]{Set: true, Value: 1500}
	element.Style.Value.Bold = template.Presence[bool]{Set: true, Value: true}
	element.Style.Value.Italic = template.Presence[bool]{Set: true, Value: true}
	element.Style.Value.Align = template.Presence[string]{Set: true, Value: "center"}
	element.Style.Value.Valign = template.Presence[string]{Set: true, Value: "middle"}
	leakyProjection, err := Canvas(leaky)
	if err != nil {
		t.Fatalf("the leak probe was refused a projection: %v", err)
	}
	probed := 0
	for _, component := range append(append([]CanvasComponent{}, projection.Components...), leakyProjection.Components...) {
		if component.Type != "text" && component.Type != "table" {
			probed++
			if component.FontFamily != nil || component.FontSize != nil || component.Bold != nil || component.Italic != nil || component.Align != nil || component.Valign != nil || component.LineSpacing != nil {
				t.Fatalf("non-text/table typography leaked into projection: %#v", component)
			}
		}
		if component.Type == "table" && component.TableBind == nil {
			t.Fatalf("table bind missing from display-only projection: %#v", component)
		}
	}
	if probed == 0 {
		t.Fatal("presence precondition: no non-text/table component was projected, so the typography-leak assertion said nothing")
	}
	long := strings.Repeat("x", maxCanvasPropertyString+1)
	bad, err := ParseTemplate([]byte(strings.Replace(string(mustSerialize(t, tpl)), `"fontSize": 12`, `"fontFamily": "`+long+`", "fontSize": 12`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Canvas(bad); err == nil {
		t.Fatal("over-bound fontFamily reached projection")
	}
}

func mustSerialize(t *testing.T, tpl *Template) []byte {
	t.Helper()
	value, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

// TestStyleAlignPropertyValidatesAgainstItsConsumersSet is Story 7.3's
// guard on the one place the alignment vocabularies could still be
// conflated, RENAMED and EXTENDED at Story 7.8.
//
// This arm previously set style.align to WHATEVER STRING ARRIVED, with no
// closed-set check at all — harmless only while one shared set served both
// `style.align` and `columns[].align`. Story 7.3 gave it the STYLE set's
// exported predicate, so a component could be justified and a nonsense
// value was refused, and the COLUMN arm (updateTableColumn,
// TestTableColumnRejectionsDoNotMutate) still refused `justify`.
//
// The name it carried — "…AgainstTheStyleSetOnly" — became FALSE at
// Story 7.8, and so did its coverage. There is no single style set: a
// table's style.align is consumed by the cell renderer and admits three
// values, a text element's is consumed by the paragraph justifier and
// admits four. This test ran only on `e1`, a TEXT element, so it stayed
// green while the table half went unwritten — and IsStyleAlign's own doc
// comment obliges this path to validate against the same single source
// the loader does. The table leg below is that obligation, and it is the
// last door by which a table document could still reach format 2.0.
func TestStyleAlignPropertyValidatesAgainstItsConsumersSet(t *testing.T) {
	tpl := componentTemplate(t)
	for _, align := range []string{"left", "center", "right", "justify"} {
		cmd := []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"align":{"op":"set","value":"` + align + `"}}}`)
		if _, err := ApplyComponentCommand(tpl, cmd); err != nil {
			t.Errorf("style.align %q must be accepted: %v", align, err)
		}
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(canonical, []byte(`"align": "justify"`)) {
		t.Fatalf("the last accepted value did not reach the canonical bytes:\n%s", canonical)
	}
	for _, align := range []string{"middle", "top", "flush", "JUSTIFY", ""} {
		cmd := []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"align":{"op":"set","value":"` + align + `"}}}`)
		_, err := ApplyComponentCommand(tpl, cmd)
		if err == nil {
			t.Errorf("style.align %q must be refused — this arm used to accept any string at all", align)
			continue
		}
		if !strings.Contains(err.Error(), "left, center, right, justify") {
			t.Errorf("the rejection must name the STYLE set's own members, got: %v", err)
		}
		after, serr := SerializeTemplate(tpl)
		if serr != nil {
			t.Fatal(serr)
		}
		if !bytes.Equal(canonical, after) {
			t.Errorf("rejecting %q mutated the canonical bytes", align)
		}
	}

	// THE TABLE LEG (Story 7.8). `e2` in this fixture is a table, and
	// `align` is offered on a table exactly as it is on text
	// (applyPropertyChanges' allowed set) — so before this story the
	// engine accepted `justify` here, styleFor wrote it onto the
	// table's style, and the very next SerializeTemplate stamped the
	// document 2.0. The designer could author, through its own engine, a
	// file it could not reopen. Story 7.4 closed the UI door only.
	//
	// A FRESH template: the text legs above deliberately left `e1`
	// justified, which raises this document to 2.0 for a legitimate
	// reason, and that would confound the version assertion below.
	tpl = componentTemplate(t)
	for _, align := range []string{"left", "center", "right"} {
		cmd := []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e2"],"changes":{"align":{"op":"set","value":"` + align + `"}}}`)
		if _, err := ApplyComponentCommand(tpl, cmd); err != nil {
			t.Errorf("a table's style.align %q must be accepted: %v", align, err)
		}
	}
	tableCanonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	cmd := []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e2"],"changes":{"align":{"op":"set","value":"justify"}}}`)
	if _, err := ApplyComponentCommand(tpl, cmd); err == nil {
		t.Fatal("a TABLE's style.align must refuse \"justify\" — the loader refuses it, and this path is required to validate against the same single source (closedsets.go's IsStyleAlign doc comment)")
	} else {
		if !strings.Contains(err.Error(), "left, center, right") {
			t.Errorf("the rejection must name the TABLE set's own members, got: %v", err)
		}
		if strings.Contains(err.Error(), "justify") {
			t.Errorf("the rejection must not name justify among a table's legal values, got: %v", err)
		}
	}
	afterTable, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tableCanonical, afterTable) {
		t.Error("refusing a table's justify mutated the canonical bytes")
	}
	if bytes.Contains(afterTable, []byte(`"version": "2.0"`)) {
		t.Error("the refused command still raised the serialized document to 2.0 — the whole point of refusing it is that the author never pays a MAJOR for a value nothing draws")
	}
}

// TestLineSpacingPropertyCommandDecodesThroughTheOneLoaderValidator is the
// coverage the `lineSpacing` command arm shipped without: Story 7.2 landed
// the arm, and neither component_commands_test.go nor this file carried a
// single case for it. Story 7.4 gives the inspector a control that sends it,
// which is the point at which the gap stops being theoretical.
//
// THE UNIT IS THE POINT. The wire carries the author's own dimensionless
// ratio as a RAW, UNQUOTED JSON number; template.DecodeLineSpacingRaw
// performs the ×1000 to thousandths itself, exactly as it does for a value
// written in a `.folio` file — which is what makes "a value refused in a file
// is refused in the inspector, for the same reason" (D-7.2.3) true by
// construction. Sending thousandths directly is refused, and this test pins
// that so nobody "fixes" the designer into sending 1500.
func TestLineSpacingPropertyCommandDecodesThroughTheOneLoaderValidator(t *testing.T) {
	for _, accepted := range []struct {
		literal     string
		thousandths int64
	}{{"1.5", 1500}, {"2", 2000}, {"0.001", 1}, {"1000", 1000000}} {
		tpl := componentTemplate(t)
		cmd := []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"set","value":` + accepted.literal + `}}}`)
		if _, err := ApplyComponentCommand(tpl, cmd); err != nil {
			t.Fatalf("lineSpacing %s must be accepted: %v", accepted.literal, err)
		}
		style := styleOfElement(t, tpl, "e1")
		if !style.LineSpacing.Set || style.LineSpacing.Value != accepted.thousandths {
			t.Fatalf("lineSpacing %s committed %d thousandths, want %d", accepted.literal, style.LineSpacing.Value, accepted.thousandths)
		}
		// And it comes back out on the projection, in the same unit the
		// document carries — the field an inspector control reads back.
		projection, err := Canvas(tpl)
		if err != nil {
			t.Fatal(err)
		}
		if got := componentOfProjection(t, projection, "e1").LineSpacing; got == nil || *got != accepted.thousandths {
			t.Fatalf("projection did not carry lineSpacing %d back: %#v", accepted.thousandths, got)
		}
	}

	// THE REFUSAL MESSAGE IS ASSERTED, not merely `err != nil`. This Message
	// field is the exact text that crosses the channel — wasm/cmd/engine
	// hands a *ComponentCommandError's Message straight to the browser, and
	// the panel prints it under the field — so a test that only checked for
	// non-nil would let the wording drift away from the mock the designer's
	// own test renders (App.test.tsx, "shows the engine's located
	// line-spacing refusal"). Note the DOUBLED key: applyPropertyChanges
	// prefixes the command key, and template's validator words its reason in
	// terms of the same key. That is what the engine really emits.
	for _, refused := range []struct {
		literal string
		message string
	}{
		{"0", "lineSpacing: lineSpacing must be between 1 and 1000000 thousandths (0.001 to 1000); 0 is outside that range"},
		{"1500", "lineSpacing: lineSpacing must be between 1 and 1000000 thousandths (0.001 to 1000); 1500000 is outside that range"},
		{"1000.5", "lineSpacing: lineSpacing must be between 1 and 1000000 thousandths (0.001 to 1000); 1000500 is outside that range"},
		{`"1.5"`, `lineSpacing: template: expected a JSON number, got a JSON string "1.5" — never coerced`},
		{"null", "lineSpacing: lineSpacing does not accept null; omit the key to inherit the leading the declared font chain itself rules"},
		{"-1", "lineSpacing: lineSpacing must be between 1 and 1000000 thousandths (0.001 to 1000); -1000 is outside that range"},
	} {
		tpl := componentTemplate(t)
		before, err := SerializeTemplate(tpl)
		if err != nil {
			t.Fatal(err)
		}
		cmd := []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"set","value":` + refused.literal + `}}}`)
		_, err = ApplyComponentCommand(tpl, cmd)
		if err == nil {
			t.Fatalf("lineSpacing %s must be refused", refused.literal)
		}
		var located *ComponentCommandError
		if !errors.As(err, &located) {
			t.Fatalf("lineSpacing %s was refused with an unlocated error (%T): %v", refused.literal, err, err)
		}
		if located.Message != refused.message {
			t.Fatalf("lineSpacing %s was refused with\n got %q\nwant %q", refused.literal, located.Message, refused.message)
		}
		if located.DataPath != "component.lineSpacing" || located.ElementID != "e1" {
			t.Fatalf("lineSpacing %s was refused at %q on element %q, want component.lineSpacing on e1", refused.literal, located.DataPath, located.ElementID)
		}
		after, err := SerializeTemplate(tpl)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("refusing lineSpacing %s mutated the canonical bytes", refused.literal)
		}
	}

	// Clear returns the element to the leading the declared chain rules.
	tpl := componentTemplate(t)
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"set","value":1.5}}}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"clear"}}}`)); err != nil {
		t.Fatal(err)
	}
	if styleOfElement(t, tpl, "e1").LineSpacing.Set {
		t.Fatal("clear left lineSpacing committed")
	}
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if componentOfProjection(t, projection, "e1").LineSpacing != nil {
		t.Fatal("a cleared lineSpacing is still projected")
	}
}

// styleOfElement and componentOfProjection address by ID rather than by
// position: the worked-example fixture's element order is not this test's to
// depend on.
func styleOfElement(t *testing.T, tpl *Template, id string) template.Style {
	t.Helper()
	for _, band := range [][]template.Element{tpl.doc.Bands.PageHeader.Elements, tpl.doc.Bands.Content.Elements, tpl.doc.Bands.PageFooter.Elements} {
		for _, element := range band {
			if string(element.ID) == id {
				return element.Style.Value
			}
		}
	}
	t.Fatalf("element %s is not in the document", id)
	return template.Style{}
}

func componentOfProjection(t *testing.T, projection CanvasProjection, id string) CanvasComponent {
	t.Helper()
	for _, component := range projection.Components {
		if component.ID == id {
			return component
		}
	}
	t.Fatalf("component %s is not in the projection", id)
	return CanvasComponent{}
}

// TestMultiParagraphValueCommitsAndReProjectsAsSeveralCanvasLines is AC1's
// engine half, end to end and with a REAL clause rather than a constant: the
// property command accepts a value far past the 512 bytes that used to reject
// the edit outright (Canvas runs inside the command's own transaction), the
// value round-trips with its mandatory breaks intact, and the canvas paints
// one CanvasTextLine per break.
func TestMultiParagraphValueCommitsAndReProjectsAsSeveralCanvasLines(t *testing.T) {
	clause := "1. The Supplier shall deliver the Goods to the Buyer at the address set out in Schedule 1, on the dates given there.\n" +
		"2. Risk in the Goods passes to the Buyer on delivery; title passes on payment in full and not before.\n" +
		"3. The Buyer shall pay each invoice within thirty days of the date of that invoice, without set-off.\n" +
		"4. Either party may terminate this agreement on thirty days' written notice to the other party.\n" +
		"5. Neither party is liable for any failure to perform caused by an event beyond its reasonable control.\n" +
		"6. This agreement is governed by the laws of the jurisdiction named in Schedule 2."
	if len(clause) <= maxCanvasPropertyString {
		t.Fatalf("fixture precondition: the clause is %d bytes and must exceed the old 512-byte cap", len(clause))
	}
	tpl := bodyTextDocument(t, "placeholder", `{"fontFamily":"body","fontSize":12}`)
	encoded, err := json.Marshal(clause)
	if err != nil {
		t.Fatal(err)
	}
	cmd := []byte(`{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":` + string(encoded) + `}}}`)
	// ACCEPTED — not merely "the canvas survived". This is the assertion the
	// story exists for: before the split, this command failed.
	projection, err := ApplyComponentCommand(tpl, cmd)
	if err != nil {
		t.Fatalf("a six-paragraph clause was rejected by the property command: %v", err)
	}
	if got := tpl.doc.Bands.Content.Elements[0].Value.Value; got != clause {
		t.Fatalf("the value did not round-trip:\n got %q\nwant %q", got, clause)
	}
	if value := componentOfProjection(t, projection, "e1").Value; value == nil || *value != clause {
		t.Fatal("the projection did not carry the committed clause back to the panel")
	}

	painted, err := CanvasWithTextPaint(tpl, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	paint := paintOf(t, painted, "e1")
	if paint == nil || paint.Truncated {
		t.Fatalf("a six-paragraph clause must project whole: %#v", paint)
	}
	// SIX numbered paragraphs, so FIVE mandatory breaks, so at least six
	// painted lines — that is what the fixture guarantees, independently of
	// how any of them wrap. (Measured at this element's 500pt width and 12pt,
	// five of the six also wrap, giving ten; the count is not asserted,
	// because wrapping is not what this test is about.)
	if len(paint.Lines) < 6 {
		t.Fatalf("projected %d canvas lines; six paragraphs must each start at least one line", len(paint.Lines))
	}
	// And each paragraph STARTS a line of its own, which is the actual claim
	// — a count alone would be satisfied by six lines of wrapped paragraph
	// one. Asserted by paragraph opening rather than by line index, so it
	// says the same thing at any column width.
	starts := make([]string, 0, len(paint.Lines))
	for _, line := range paint.Lines {
		if len(line.Fragments) > 0 {
			starts = append(starts, line.Fragments[0].Text)
		}
	}
	for _, opening := range []string{"1. The Supplier", "2. Risk in", "3. The Buyer", "4. Either party", "5. Neither party", "6. This agreement"} {
		if !slices.ContainsFunc(starts, func(start string) bool { return strings.HasPrefix(start, opening) }) {
			t.Fatalf("no painted line begins %q; each mandatory break must start a new line. Line openings: %q", opening, starts)
		}
	}
	if !strings.HasPrefix(starts[0], "1. The Supplier") {
		t.Fatalf("the first painted line is not the clause's first paragraph: %q", starts[0])
	}

	// A CRLF pair is ONE mandatory break, not two — the engine folds it, so
	// a clause pasted from a Windows word processor does not double-space.
	crlf := bodyTextDocument(t, "placeholder", `{"fontFamily":"body","fontSize":12}`)
	lf := bodyTextDocument(t, "placeholder", `{"fontFamily":"body","fontSize":12}`)
	crlf.doc.Bands.Content.Elements[0].Value.Value = "One.\r\nTwo.\r\nThree."
	lf.doc.Bands.Content.Elements[0].Value.Value = "One.\nTwo.\nThree."
	crlfPaint, err := CanvasWithTextPaint(crlf, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	lfPaint, err := CanvasWithTextPaint(lf, testFontSet())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(paintOf(t, crlfPaint, "e1").Lines), len(paintOf(t, lfPaint, "e1").Lines); got != want {
		t.Fatalf("CRLF projected %d lines against LF's %d; the pair must fold to one break", got, want)
	}
}

// TestUpdateComponentPropertiesBandBoundsFollowTheColumnLift covers the
// property-edit surface's SOLE band bound.
//
// applyPropertyChanges enforces only positivity on x and y; the single
// containComponent call in updateComponentPropertiesInPlace is the whole of
// the band checking for every property this panel can edit. Before Story 7.5
// this file had no band-bounds coverage at all — every component it created
// sat at a safe interior position — so the path was untested on both sides of
// the clause that moved.
func TestUpdateComponentPropertiesBandBoundsFollowTheColumnLift(t *testing.T) {
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
	// The content band is 679.89pt of window on this fixture's A4 page, so
	// 2400pt is a placement on the fourth sheet — a schedule's later rows, or
	// a signature block at the end of a long statement.
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["`+created.ID+`"],"changes":{"y":{"op":"set","value":2400}}}`)); err != nil {
		t.Fatalf("a content y three windows down was refused by the property panel: %v", err)
	}
	if roundTripped := reloadedComponent(t, tpl, created.ID); roundTripped.Y != 2400000 {
		t.Fatalf("canonical bytes carried y = %d, want 2400000", roundTripped.Y)
	}
	// The horizontal bound is untouched: the column is unbounded vertically,
	// never horizontally.
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["`+created.ID+`"],"changes":{"x":{"op":"set","value":2400}}}`)); err == nil {
		t.Fatal("a content x past the band width was accepted by the property panel")
	}
	// And a repeating band still caps, in the same words. A page header is
	// exactly one page tall because that is what repeating means.
	headerProjection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"pageHeader","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	inHeader := newProjectedComponent(t, createdProjection, headerProjection)
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ApplyComponentCommand(tpl, []byte(`{"kind":"updateComponentProperties","version":1,"ids":["`+inHeader.ID+`"],"changes":{"y":{"op":"set","value":600}}}`))
	var failure *ComponentCommandError
	if !errors.As(err, &failure) {
		t.Fatalf("a pageHeader y past the band = %v, want a component command failure", err)
	}
	if want := "folio: component geometry must stay within pageHeader"; failure.Message != want {
		t.Fatalf("pageHeader refusal message = %q, want exactly %q", failure.Message, want)
	}
	if failure.DataPath != "component.geometry" {
		t.Fatalf("pageHeader refusal data path = %q, want component.geometry", failure.DataPath)
	}
	if after, _ := SerializeTemplate(tpl); !bytes.Equal(canonical, after) {
		t.Fatal("a refused property change moved the canonical bytes")
	}
}

// TestAPropertyCommandCannotStampATableDocumentAt2_0 is Story 7.8's
// version half on the IN-MEMORY path, end to end: command in, canonical
// bytes out, and those bytes loaded again.
//
// The file path closes by construction — a loader refusal means no
// *Document ever reaches versionRequiredByContent, which runs only at
// save.
//
// The property-command path holds an already-loaded document, so the
// loader cannot reach it directly. It is closed TWICE over, and the
// distinction is worth stating because a green here does not by itself
// prove the align arm is doing anything: updateComponentProperties (and
// wasm.Engine.Apply) serialize and RE-PARSE before installing, so a
// document the loader refuses cannot be installed whatever the arm
// does. The arm is what makes the refusal LOCATED rather than a generic
// "component properties did not pass format validation" — which is what
// TestStyleAlignPropertyValidatesAgainstItsConsumersSet's table leg
// pins, and that leg is the live red-proof of the arm itself.
//
// THE TWO LEGS TOGETHER ARE THE DISCRIMINATOR. The text leg is what
// distinguishes this from a blanket ban on `justify`, which Story 7.3,
// Story 7.4 and two shipped goldens forbid.
func TestAPropertyCommandCannotStampATableDocumentAt2_0(t *testing.T) {
	command := func(id, align string) []byte {
		return []byte(`{"kind":"updateComponentProperties","version":1,"ids":["` + id + `"],"changes":{"align":{"op":"set","value":"` + align + `"}}}`)
	}

	// THE TABLE: refused, and the document it would have raised stays
	// where it was — and stays loadable by a 1.x reader.
	tableTpl := componentTemplate(t)
	before, err := SerializeTemplate(tableTpl)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(before, []byte(`"version": "1.0"`)) {
		t.Fatalf("fixture precondition: worked-example.json must start at 1.0, got:\n%s", before)
	}
	if _, err := ApplyComponentCommand(tableTpl, command("e2", "justify")); err == nil {
		t.Fatal("the property command must refuse justify on a table — otherwise the designer's own engine authors a 2.0 document whose value nothing draws and which the loader then refuses to reopen (AC1)")
	} else {
		t.Logf("refused: %v", err)
	}
	after, err := SerializeTemplate(tableTpl)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("the refused command mutated the document")
	}
	if !bytes.Contains(after, []byte(`"version": "1.0"`)) {
		t.Errorf("the serialized document is no longer 1.0 after a REFUSED command:\n%s", after)
	}
	if _, err := ParseTemplate(after); err != nil {
		t.Errorf("the document the engine still holds must remain loadable: %v", err)
	}

	// THE TEXT ELEMENT: accepted exactly as today, and it DOES raise the
	// document to 2.0 — which is the honest cost of FR47 and is not what
	// this story removes.
	textTpl := componentTemplate(t)
	if _, err := ApplyComponentCommand(textTpl, command("e1", "justify")); err != nil {
		t.Fatalf("justify on a TEXT element must still be accepted: %v", err)
	}
	textBytes, err := SerializeTemplate(textTpl)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(textBytes, []byte(`"align": "justify"`)) {
		t.Errorf("the accepted value did not reach the canonical bytes:\n%s", textBytes)
	}
	if !bytes.Contains(textBytes, []byte(`"version": "2.0"`)) {
		t.Errorf("a justified TEXT element must still raise the document to 2.0 — narrowing by consumer is not a ban:\n%s", textBytes)
	}
}
