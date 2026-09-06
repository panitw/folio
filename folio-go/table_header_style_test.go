package folio

// STORY 12.3 — the three table properties the engine has always rendered and
// no command could write: headerHeight, altRowBackground and the headerStyle
// block. One test per row of the story's own I/O matrix, plus the writer
// census that keeps the claim about WHO authors headerStyle true.
//
// This file is declared in table_behaviour_suite_test.go's
// declaredTableBehaviourSuite: it is named `table_*_test.go`, and that guard
// compares its literal list against `filepath.Glob("table_*_test.go")` in BOTH
// directions, so a new file of that name reds the guard on the commit that
// adds it unless the list moves with it.

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/template"
)

// theWorkedExampleTable is the table this file edits: `e2` in the content band
// of the worked example, which already declares `style.fontFamily: body` and
// `style.fontSize: 8` and NO headerStyle at all. That combination is what makes
// the fallback observable — a cleared header field has somewhere real to fall
// through to.
const theWorkedExampleTable = "e2"

func headerStyleFixture(t *testing.T) *Template {
	t.Helper()
	tpl := componentTemplate(t)
	view, err := TableColumns(tpl, theWorkedExampleTable)
	if err != nil {
		t.Fatalf("fixture precondition: the worked example must project %s as a table: %v", theWorkedExampleTable, err)
	}
	if view.HeaderFontSize != 0 || view.HeaderFontSizeResolved != 8000 || view.HeaderFontFamilyResolved != "body" {
		t.Fatalf("fixture precondition: %s must declare no headerStyle and a table style of 8pt body, got %#v", theWorkedExampleTable, view)
	}
	return tpl
}

func applyToTable(t *testing.T, tpl *Template, command string) error {
	t.Helper()
	_, err := ApplyComponentCommand(tpl, []byte(command))
	return err
}

func mustApplyToTable(t *testing.T, tpl *Template, command string) {
	t.Helper()
	if err := applyToTable(t, tpl, command); err != nil {
		t.Fatalf("apply %s: %v", command, err)
	}
}

func canonicalBytes(t *testing.T, tpl *Template) []byte {
	t.Helper()
	encoded, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	return encoded
}

func projectTable(t *testing.T, tpl *Template) TableColumnsProjection {
	t.Helper()
	view, err := TableColumns(tpl, theWorkedExampleTable)
	if err != nil {
		t.Fatalf("project %s: %v", theWorkedExampleTable, err)
	}
	return view
}

// refusalLeavesTheDocumentAlone is the shape every refusal row shares: the
// command errors, the error is LOCATED (an element id and a data path, not a
// bare sentence), and the canonical bytes are byte-identical afterwards.
func refusalLeavesTheDocumentAlone(t *testing.T, tpl *Template, command, wantPath string) {
	t.Helper()
	before := canonicalBytes(t, tpl)
	err := applyToTable(t, tpl, command)
	if err == nil {
		t.Fatalf("apply %s: expected a refusal", command)
	}
	located, ok := err.(*ComponentCommandError)
	if !ok {
		t.Fatalf("apply %s: error is %T, not a located *ComponentCommandError — an unlocated refusal reaches the author as ENGINE_REJECTED with no field to look at", command, err)
	}
	if located.DataPath != wantPath {
		t.Errorf("apply %s: refusal is located at %q, want %q", command, located.DataPath, wantPath)
	}
	if after := canonicalBytes(t, tpl); !bytes.Equal(before, after) {
		t.Errorf("apply %s: a refused command changed the canonical bytes", command)
	}
}

// refusalSaysWhy is refusalLeavesTheDocumentAlone plus the REASON.
//
// A refusal whose MESSAGE is unasserted can be deleted outright and replaced by
// an unrelated fallback that happens to share its DataPath, with every test
// still green. That is not hypothetical: deleting setTableHeaderHeight's `op`
// guard at this tree left both of its tests passing, because the command then
// fell through to "headerHeight must be a positive length" — a different
// refusal, for a different reason, located at the same table.headerHeight.
// Pinning the sentence is what makes the guard falsifiable.
func refusalSaysWhy(t *testing.T, tpl *Template, command, wantPath, wantMessage string) {
	t.Helper()
	refusalLeavesTheDocumentAlone(t, tpl, command, wantPath)
	located, ok := applyToTable(t, tpl, command).(*ComponentCommandError)
	if !ok {
		t.Fatalf("apply %s: expected a located refusal", command)
	}
	if located.Message != wantMessage {
		t.Errorf("apply %s: refused with %q, want %q — the guard this story added must be the one that fired, not a fallback that shares its DataPath", command, located.Message, wantMessage)
	}
}

// ---------------------------------------------------------------------------
// The I/O matrix, one test per row.
// ---------------------------------------------------------------------------

func TestSetHeaderStyleFieldWritesItAndProjectsBothMembers(t *testing.T) {
	tpl := headerStyleFixture(t)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"fontSize","op":"set","value":14}`)
	view := projectTable(t, tpl)
	if view.HeaderFontSize != 14000 || view.HeaderFontSizeResolved != 14000 {
		t.Errorf("committed/resolved fontSize = %d/%d, want 14000/14000", view.HeaderFontSize, view.HeaderFontSizeResolved)
	}
	if !bytes.Contains(canonicalBytes(t, tpl), []byte(`"headerStyle"`)) {
		t.Error("the document does not carry a headerStyle key after one was authored")
	}
}

// THE ROW THIS WHOLE STORY TURNS ON. A cleared field's RESOLVED member must
// read the table's own style — 8pt here, from the fixture — and it must do so
// because the ENGINE cascaded it, not because anything recomputed it. The
// committed member goes back to absent in the same breath, which is what keeps
// "clearable back to absent" meaningful.
func TestClearHeaderStyleFieldFallsBackToTheTablesOwnStyle(t *testing.T) {
	tpl := headerStyleFixture(t)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"fontSize","op":"set","value":14}`)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"background","op":"set","value":"#101010"}`)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"fontSize","op":"clear"}`)
	view := projectTable(t, tpl)
	if view.HeaderFontSize != 0 {
		t.Errorf("committed fontSize = %d after a clear, want 0 (absent)", view.HeaderFontSize)
	}
	if view.HeaderFontSizeResolved != 8000 {
		t.Errorf("resolved fontSize = %d, want the table's own style.fontSize of 8000", view.HeaderFontSizeResolved)
	}
	encoded := canonicalBytes(t, tpl)
	if bytes.Contains(encoded, []byte(`"fontSize": 14`)) {
		t.Error("the cleared fontSize is still in the canonical bytes")
	}
	// The zero Presence removes the KEY. An explicit null would still be a key
	// in the file — different bytes, an undo entry, and a raised format version.
	if bytes.Contains(encoded, []byte(`"fontSize": null`)) {
		t.Error("clearing wrote an explicit null rather than removing the key")
	}
	// And the sibling that was NOT cleared is untouched, so this is a clear of
	// one field rather than of the block.
	if view.HeaderBackground != "#101010" {
		t.Errorf("committed background = %q, want the sibling field left alone", view.HeaderBackground)
	}
}

// AND WITH NO TABLE STYLE TO FALL BACK TO, the resolved member is the format's
// documented default rather than nothing.
func TestClearedHeaderStyleFieldFallsBackToTheDocumentedDefault(t *testing.T) {
	tpl := headerStyleFixture(t)
	mustApplyToTable(t, tpl, `{"kind":"updateComponentProperties","version":1,"ids":["`+theWorkedExampleTable+`"],"changes":{"fontSize":{"op":"clear"}}}`)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"fontSize","op":"set","value":14}`)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"fontSize","op":"clear"}`)
	if got := projectTable(t, tpl).HeaderFontSizeResolved; got != int64(defaultFontSizePt) {
		t.Errorf("resolved fontSize with no table style = %d, want the engine's default %d", got, defaultFontSizePt)
	}
}

func TestClearingTheLastHeaderStyleFieldRemovesTheBlock(t *testing.T) {
	tpl := headerStyleFixture(t)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"align","op":"set","value":"center"}`)
	if !bytes.Contains(canonicalBytes(t, tpl), []byte(`"headerStyle"`)) {
		t.Fatal("precondition: the block must exist before the clear that removes it")
	}
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"align","op":"clear"}`)
	encoded := canonicalBytes(t, tpl)
	if bytes.Contains(encoded, []byte(`"headerStyle"`)) {
		t.Errorf("clearing the last field left a headerStyle key behind:\n%s", encoded)
	}
	// An empty object is the failure this row exists for, and it is a different
	// failure from a leftover key with a value in it.
	if bytes.Contains(encoded, []byte(`"headerStyle": {}`)) {
		t.Error("clearing the last field left an empty headerStyle object")
	}
}

func TestSetAndClearTheAlternatingRowBackground(t *testing.T) {
	tpl := headerStyleFixture(t)
	mustApplyToTable(t, tpl, `{"kind":"setTableAltRowBackground","version":1,"id":"`+theWorkedExampleTable+`","op":"set","value":"#DDEEFF"}`)
	if got := projectTable(t, tpl).AltRowBackground; got != "#DDEEFF" {
		t.Errorf("altRowBackground = %q, want #DDEEFF", got)
	}
	if !bytes.Contains(canonicalBytes(t, tpl), []byte(`"altRowBackground": "#DDEEFF"`)) {
		t.Error("altRowBackground did not reach the canonical bytes")
	}
	mustApplyToTable(t, tpl, `{"kind":"setTableAltRowBackground","version":1,"id":"`+theWorkedExampleTable+`","op":"clear"}`)
	if got := projectTable(t, tpl).AltRowBackground; got != "" {
		t.Errorf("altRowBackground = %q after a clear, want absent", got)
	}
	if encoded := canonicalBytes(t, tpl); bytes.Contains(encoded, []byte(`"altRowBackground"`)) {
		t.Errorf("clearing left the key in the document:\n%s", encoded)
	}
}

func TestSetTheHeaderHeightInMillipoints(t *testing.T) {
	tpl := headerStyleFixture(t)
	mustApplyToTable(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+theWorkedExampleTable+`","height":18}`)
	if got := projectTable(t, tpl).HeaderHeight; got != 18000 {
		t.Errorf("headerHeight = %d, want 18000 millipoints", got)
	}
	if !bytes.Contains(canonicalBytes(t, tpl), []byte(`"headerHeight": 18`)) {
		t.Error("headerHeight did not reach the canonical bytes in points")
	}
}

// headerHeight IS REQUIRED, so a clear is refused rather than honoured. No
// clear affordance is rendered anywhere, which makes this reachable only from a
// hand-built command — and that is exactly why the engine has to refuse it: the
// document that came back would not load.
func TestTheHeaderHeightCannotBeCleared(t *testing.T) {
	tpl := headerStyleFixture(t)
	refusalSaysWhy(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+theWorkedExampleTable+`","op":"clear"}`, "table.headerHeight", "headerHeight is required: it accepts neither a clear nor a null")
}

// A TABLE'S PROJECTED HEIGHT IS ITS headerHeight (projectedSize), so growing it
// can push the table out of a band that caps vertically. The check is
// containComponent — the same predicate updateTableColumn's width case asks,
// one new call site.
func TestAHeaderHeightThatOverflowsItsBandIsRefused(t *testing.T) {
	tpl := componentTemplate(t)
	before, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	created, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"table","band":"pageHeader","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatalf("create a table in the page header: %v", err)
	}
	table := newProjectedComponent(t, before, created)
	mustApplyToTable(t, tpl, `{"kind":"addTableColumn","version":1,"id":"`+table.ID+`","index":0}`)
	// The page header is 60pt tall in this fixture, so 40pt fits and 100pt
	// cannot. Both arms, so the refusal is not simply "every height is refused".
	if err := applyToTable(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+table.ID+`","height":40}`); err != nil {
		t.Fatalf("a header height that fits its band must be accepted: %v", err)
	}
	encoded := canonicalBytes(t, tpl)
	err = applyToTable(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+table.ID+`","height":100}`)
	located, ok := err.(*ComponentCommandError)
	if !ok {
		t.Fatalf("a header height taller than its band produced %T, want a located refusal", err)
	}
	if located.DataPath != "table.headerHeight" || located.ElementID != table.ID {
		t.Errorf("refusal located at %q/%q, want table.headerHeight on %s", located.ElementID, located.DataPath, table.ID)
	}
	if after := canonicalBytes(t, tpl); !bytes.Equal(encoded, after) {
		t.Error("a refused header height changed the canonical bytes")
	}
}

func TestAMalformedAlternatingRowColourIsRefused(t *testing.T) {
	tpl := headerStyleFixture(t)
	refusalLeavesTheDocumentAlone(t, tpl, `{"kind":"setTableAltRowBackground","version":1,"id":"`+theWorkedExampleTable+`","op":"set","value":"not-a-colour"}`, "table.altRowBackground")
}

func TestAnUnknownHeaderStyleFieldIsRefused(t *testing.T) {
	tpl := headerStyleFixture(t)
	// `padding` and `border` are refused by the SAME gate as a nonsense name:
	// the arm's closed set is the only door, and neither is in it.
	//
	// ⚠ `bold` and `italic` LEFT THIS LIST AT STORY 11.2 and are in the
	// closed set now — resolveHeaderStyle cascades both, so a header style
	// declaring either is read by something that draws (FR57, AC2). They are
	// asserted below instead, where a NON-BOOLEAN value for them is refused
	// at the field itself rather than at the gate.
	for _, field := range []string{"paddingTop", "border", "padding", "notAField"} {
		refusalLeavesTheDocumentAlone(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"`+field+`","op":"set","value":"x"}`, "table.headerStyle")
	}
	// The two booleans take their own arm, so `"x"` is refused by the VALUE
	// check and locates at the field — a different door from the gate above,
	// and asserting it is what keeps the two doors distinguishable.
	for _, field := range []string{"bold", "italic"} {
		refusalLeavesTheDocumentAlone(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"`+field+`","op":"set","value":"x"}`, "table.headerStyle."+field)
	}
	// And the nine that ARE in it are all reachable, so the closed set is not
	// simply refusing everything.
	for _, field := range tableHeaderStyleFields {
		if err := applyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"`+field+`","op":"clear"}`); err != nil {
			t.Errorf("clearing the declared field %q was refused: %v", field, err)
		}
	}
}

// TestTheHeaderStyleBooleanArmActuallyStoresWhatItWasGiven is P4: the
// two booleans Story 11.2 added to the closed set take their OWN arm
// above the string default, and nothing observed what that arm WROTE.
// Measured by mutation: inverting the stored value (`Value: !flag`) left
// the whole suite green — every other assertion about them checks only
// that the field is reachable or that a bad value is refused.
//
// It reads the value back out of the RE-PARSED CANONICAL BYTES rather
// than off the in-memory template, so the assertion covers the command,
// the serializer and the loader in one statement — and `false` is
// asserted as well as `true`, because an arm that stored a constant
// would satisfy either one alone.
func TestTheHeaderStyleBooleanArmActuallyStoresWhatItWasGiven(t *testing.T) {
	for _, tc := range []struct {
		field string
		value string
		read  func(template.Style) template.Presence[bool]
	}{
		{"bold", "true", func(st template.Style) template.Presence[bool] { return st.Bold }},
		{"bold", "false", func(st template.Style) template.Presence[bool] { return st.Bold }},
		{"italic", "true", func(st template.Style) template.Presence[bool] { return st.Italic }},
		{"italic", "false", func(st template.Style) template.Presence[bool] { return st.Italic }},
	} {
		t.Run(tc.field+"="+tc.value, func(t *testing.T) {
			tpl := headerStyleFixture(t)
			mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"`+tc.field+`","op":"set","value":`+tc.value+`}`)
			reloaded, err := ParseTemplate(canonicalBytes(t, tpl))
			if err != nil {
				t.Fatalf("the command produced bytes that do not load: %v", err)
			}
			var found bool
			for _, el := range reloaded.doc.Bands.Content.Elements {
				if string(el.ID) != theWorkedExampleTable {
					continue
				}
				found = true
				hs := el.Table.Value.HeaderStyle
				if !hs.Set || hs.Null {
					t.Fatalf("the command left no headerStyle block to carry %s", tc.field)
				}
				got := tc.read(hs.Value)
				if !got.Set || got.Null {
					t.Fatalf("headerStyle.%s did not survive the round trip: %+v", tc.field, got)
				}
				if want := tc.value == "true"; got.Value != want {
					t.Fatalf("headerStyle.%s round-tripped as %v, want %v — the command's own arm is storing something other than what it was given", tc.field, got.Value, want)
				}
			}
			if !found {
				t.Fatalf("fixture precondition: element %s is not in the content band, so nothing above was asserted", theWorkedExampleTable)
			}
		})
	}
}

func TestNullIsRefusedByAllThreeArms(t *testing.T) {
	tpl := headerStyleFixture(t)
	refusalLeavesTheDocumentAlone(t, tpl, `{"kind":"setTableAltRowBackground","version":1,"id":"`+theWorkedExampleTable+`","op":"null"}`, "table.altRowBackground")
	refusalLeavesTheDocumentAlone(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"color","op":"null"}`, "table.headerStyle.color")
	refusalSaysWhy(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+theWorkedExampleTable+`","op":"null"}`, "table.headerHeight", "headerHeight is required: it accepts neither a clear nor a null")
}

// AND THE GATE THE SEVEN COLUMN ARMS SHARE, repeated here because these three
// repeat it: not found, and not a table.
func TestTheThreeArmsShareTheTableGate(t *testing.T) {
	tpl := headerStyleFixture(t)
	for _, command := range []string{
		`{"kind":"setTableHeaderHeight","version":1,"id":"nope","height":18}`,
		`{"kind":"setTableAltRowBackground","version":1,"id":"nope","op":"clear"}`,
		`{"kind":"updateTableHeaderStyle","version":1,"id":"nope","field":"align","op":"clear"}`,
	} {
		refusalLeavesTheDocumentAlone(t, tpl, command, "table.id")
	}
	// AND A MALFORMED COMMAND AGAINST A MISSING ID IS STILL "table.id".
	//
	// Well-formed ops alone could not see this: the three arms reached the
	// shared gate at DIFFERENT points, so `{"id":"nope","op":"bogus"}` was
	// refused at table.altRowBackground — a refusal naming a field on an
	// element that does not exist, pointing the author at the wrong thing. The
	// ordering is now the same in all three arms, and this row is what holds it
	// there.
	for _, command := range []string{
		`{"kind":"setTableHeaderHeight","version":1,"id":"nope","op":"bogus"}`,
		`{"kind":"setTableAltRowBackground","version":1,"id":"nope","op":"bogus"}`,
		`{"kind":"updateTableHeaderStyle","version":1,"id":"nope","field":"align","op":"bogus"}`,
		// A field name outside the seven, on an element that is not there:
		// still the element, not the field.
		`{"kind":"updateTableHeaderStyle","version":1,"id":"nope","field":"notAField","op":"clear"}`,
		// And a surplus key, on an element that is not there.
		`{"kind":"setTableHeaderHeight","version":1,"id":"nope","height":18,"surplus":1}`,
	} {
		refusalLeavesTheDocumentAlone(t, tpl, command, "table.id")
	}
	// A component that exists and is not a table takes the second sentence.
	before, err := Canvas(tpl)
	if err != nil {
		t.Fatal(err)
	}
	created, err := ApplyComponentCommand(tpl, []byte(`{"kind":"createComponent","version":1,"type":"rect","band":"content","x":0,"y":0,"width":72,"height":24,"snap":false}`))
	if err != nil {
		t.Fatal(err)
	}
	rect := newProjectedComponent(t, before, created)
	refusalLeavesTheDocumentAlone(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+rect.ID+`","height":18}`, "table.id")
}

// A DOCUMENT NOBODY EDITED SERIALIZES TO THE SAME BYTES. This story adds no
// key, removes none, and raises no version, so opening the editor on a table
// and touching nothing must be free.
func TestATableWhoseHeaderIsNotEditedSerializesIdentically(t *testing.T) {
	tpl := headerStyleFixture(t)
	before := canonicalBytes(t, tpl)
	if _, err := TableColumns(tpl, theWorkedExampleTable); err != nil {
		t.Fatal(err)
	}
	if after := canonicalBytes(t, tpl); !bytes.Equal(before, after) {
		t.Error("projecting a table for the editor changed the document")
	}
	// And the fixture on disk is unchanged by the round trip, which is what
	// "no corpus digest moves" means for this one document.
	source, err := os.ReadFile(filepath.Join("testdata", "template", "golden", "worked-example.json"))
	if err != nil {
		t.Fatal(err)
	}
	reparsed, err := ParseTemplate(source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonicalBytes(t, reparsed), before) {
		t.Error("the fixture no longer round-trips to the same bytes")
	}
}

// RE-SETTING A VALUE TO WHAT IT ALREADY IS IS A SILENT SUCCESS, not an error.
// The canonical-bytes short-circuit that turns it into "no revision, no undo
// entry" lives one layer up in wasm/engine.go's Apply, which runs it BEFORE
// pushUndo and install; what is asserted here is the half this layer owns —
// the command is accepted and the bytes do not move.
func TestReSettingAHeaderStyleFieldToItsCurrentValueIsASilentNoOp(t *testing.T) {
	tpl := headerStyleFixture(t)
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"fontSize","op":"set","value":14}`)
	settled := canonicalBytes(t, tpl)
	if err := applyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"fontSize","op":"set","value":14}`); err != nil {
		t.Fatalf("re-setting a value to itself must be a silent success, got %v", err)
	}
	if after := canonicalBytes(t, tpl); !bytes.Equal(settled, after) {
		t.Error("re-setting a value to itself moved the bytes")
	}
	// Clearing an already-absent field is the same shape from the other side.
	mustApplyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"color","op":"clear"}`)
	if after := canonicalBytes(t, tpl); !bytes.Equal(settled, after) {
		t.Error("clearing an already-absent field moved the bytes")
	}
}

// EVERY FIELD VALIDATES THROUGH THE SAME PREDICATE THE LOADER ASKS. The
// consequence of getting this wrong is not a bad message: a command door that
// admitted what the file door refuses would stamp out a document the designer
// cannot reopen.
func TestHeaderStyleValuesAreJudgedByTheLoadersOwnPredicates(t *testing.T) {
	tpl := headerStyleFixture(t)
	for _, refused := range []struct{ field, value, path string }{
		{"align", `"justify"`, "table.headerStyle.align"},
		{"valign", `"centre"`, "table.headerStyle.valign"},
		{"color", `"#GGGGGG"`, "table.headerStyle.color"},
		{"background", `"rebeccapurple"`, "table.headerStyle.background"},
		{"fontFamily", `"no-such-chain"`, "table.headerStyle.fontFamily"},
		{"fontSize", `0`, "table.headerStyle.fontSize"},
		{"fontSize", `-3`, "table.headerStyle.fontSize"},
		{"lineSpacing", `"1.5"`, "table.headerStyle.lineSpacing"},
	} {
		refusalLeavesTheDocumentAlone(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"`+refused.field+`","op":"set","value":`+refused.value+`}`, refused.path)
	}
	// The accepting arm for each, so the refusals above are about the VALUE and
	// not about the field.
	for _, accepted := range []struct{ field, value string }{
		{"align", `"center"`},
		{"valign", `"middle"`},
		{"color", `"#112233"`},
		{"background", `"#445566"`},
		{"fontFamily", `"body"`},
		{"fontSize", `14`},
		{"lineSpacing", `1.5`},
	} {
		if err := applyToTable(t, tpl, `{"kind":"updateTableHeaderStyle","version":1,"id":"`+theWorkedExampleTable+`","field":"`+accepted.field+`","op":"set","value":`+accepted.value+`}`); err != nil {
			t.Errorf("a legal %s value was refused: %v", accepted.field, err)
		}
	}
	view := projectTable(t, tpl)
	if view.HeaderAlign != "center" || view.HeaderValign != "middle" || view.HeaderColor != "#112233" || view.HeaderBackground != "#445566" || view.HeaderFontFamily != "body" || view.HeaderFontSize != 14000 || view.HeaderLineSpacing != 1500 {
		t.Errorf("the seven accepted values did not all land: %#v", view)
	}
	// And the resolved twin of a SET field is that field: the cascade prefers
	// headerStyle over the table's own style, which is the whole point of it.
	if view.HeaderAlignResolved != "center" || view.HeaderFontSizeResolved != 14000 {
		t.Errorf("resolved members do not follow a committed one: %#v", view)
	}
}

// ---------------------------------------------------------------------------
// The writer census.
// ---------------------------------------------------------------------------

// declaredHeaderStyleAssigners is every non-test function in the module ROOT
// (package folio) that ASSIGNS to a HeaderStyle-rooted expression, or takes a
// writable pointer into one.
//
// THE PHRASING MATTERS AND BOTH OBVIOUS PHRASINGS ARE FALSE. "Nothing writes
// headerStyle" was already false before this story: renameFontChain rewrites
// HeaderStyle.FontFamily when a font chain is renamed, and it is the ONE such
// site — verified at D-12.3.0's own tree with `git show 71627a5`, where the
// ruling's "two sites" was an over-count when it was written rather than drift
// since. And "two writers" is false in the other direction now that 12.3 exists.
// What is true, and what this list says, is that the set is CLOSED and every
// member is named.
var declaredHeaderStyleAssigners = []string{
	"cleanupEmptyHeaderStyle", // 12.3 — drops a block whose last field was cleared
	"headerStyleFor",          // 12.3 — materialises the block and hands out the pointer
	"renameFontChain",         // 8.1 — rewrites HeaderStyle.FontFamily when a chain is renamed
}

// declaredHeaderStyleCommands is the far narrower claim the story actually
// makes: exactly ONE command AUTHORS the header style, and it is the arm added
// for it. renameFontChain is not in this list because it authors nothing — it
// follows a rename the author made somewhere else.
var declaredHeaderStyleCommands = []string{"updateTableHeaderStyle"}

func TestOnlyTheDeclaredSitesWriteAHeaderStyle(t *testing.T) {
	assigners, callers := map[string]bool{}, map[string]bool{}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	scanned := 0
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		fset := token.NewFileSet()
		file, perr := parser.ParseFile(fset, name, nil, 0)
		if perr != nil {
			t.Fatalf("parse %s: %v", name, perr)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.AssignStmt:
					for _, lhs := range n.Lhs {
						if namesHeaderStyle(lhs) {
							assigners[fn.Name.Name] = true
						}
					}
				case *ast.UnaryExpr:
					if n.Op == token.AND && namesHeaderStyle(n.X) {
						assigners[fn.Name.Name] = true
					}
				case *ast.CallExpr:
					if ident, ok := n.Fun.(*ast.Ident); ok && ident.Name == "headerStyleFor" {
						callers[fn.Name.Name] = true
					}
				}
				return true
			})
		}
	}
	if scanned == 0 {
		t.Fatal("vacuity guard: the census parsed ZERO non-test files in the module root, so it proves nothing")
	}
	if got := sortedFunctionNames(assigners); !reflect.DeepEqual(got, declaredHeaderStyleAssigners) {
		t.Errorf("the functions assigning to a HeaderStyle are\n\t%v\nand the declared set is\n\t%v — a new writer is a new place a header style can change, and this list is where that becomes a reviewed line rather than a silent one", got, declaredHeaderStyleAssigners)
	}
	// headerStyleFor is the only function that hands out a writable pointer into
	// the block, so its caller set IS the set of commands that author one.
	if got := sortedFunctionNames(callers); !reflect.DeepEqual(got, declaredHeaderStyleCommands) {
		t.Errorf("the functions taking a writable header style are\n\t%v\nand the declared set is\n\t%v", got, declaredHeaderStyleCommands)
	}
	// Non-vacuity: the matcher finds a write in a synthetic source, so an empty
	// result above would have been a finding rather than a pass.
	if !matcherFindsAHeaderStyleWrite(t, "package p\nfunc w(e *E) { e.Table.Value.HeaderStyle = x }\n") {
		t.Error("the census matcher does not recognise a plain HeaderStyle assignment")
	}
	if matcherFindsAHeaderStyleWrite(t, "package p\nfunc r(e *E) bool { return e.Table.Value.HeaderStyle.Set }\n") {
		t.Error("the census matcher counts a READ as a write, which would make the declared set meaningless")
	}
}

// namesHeaderStyle reports whether expr is rooted at a `.HeaderStyle` selector
// — `x.HeaderStyle`, `x.HeaderStyle.Value`, `x.HeaderStyle.Value.FontFamily`
// and so on all count, because every one of them writes into the block.
func namesHeaderStyle(expr ast.Expr) bool {
	for {
		sel, ok := expr.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if sel.Sel.Name == "HeaderStyle" {
			return true
		}
		expr = sel.X
	}
}

func matcherFindsAHeaderStyleWrite(t *testing.T, source string) bool {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "synthetic.go", source, 0)
	if err != nil {
		t.Fatalf("parse the synthetic source: %v", err)
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if assign, ok := node.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if namesHeaderStyle(lhs) {
					found = true
				}
			}
		}
		return true
	})
	return found
}

func sortedFunctionNames(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	slices.Sort(out)
	return out
}

// ---------------------------------------------------------------------------
// REVIEW PATCH SET (2026-09-05). The four findings below are all the same
// shape: a claim the story makes that nothing in the tree could falsify.
// ---------------------------------------------------------------------------

// headerCascadeDocument is a hand-authored table that declares exactly the two
// style blocks a row asks it to and nothing else, so a resolved member can be
// watched moving from level to level. It is hand-authored rather than driven
// off the worked example because two of the things these rows need — a SECOND
// font chain, and a headerStyle carrying a field no command can author — are
// reachable only from the file door.
//
// tableStyle and headerStyle are whole JSON fragments including their trailing
// comma, or "" for "the block is not there at all".
func headerCascadeDocument(t *testing.T, tableStyle, headerStyle string) *Template {
	t.Helper()
	doc := `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "rows[]", "as": "row", "headerHeight": 20,
          ` + tableStyle + headerStyle + `
          "columns": [{"id": "e2", "label": "A", "width": 100, "align": "left", "bind": "{{row.a}}"}]}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"], "display": ["Roboto-Bold"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	tpl, err := ParseTemplate([]byte(doc))
	if err != nil {
		t.Fatalf("parse the cascade fixture: %v", err)
	}
	return tpl
}

// THE FOUR RESOLVED MEMBERS NOTHING ASSERTED, AND THE THREE THAT WERE.
//
// Measured before this test existed: HeaderBackgroundResolved, HeaderColorResolved,
// HeaderLineSpacingResolved and HeaderValignResolved appeared in exactly one
// production file and in ZERO `_test.go` files, while HeaderFontSizeResolved,
// HeaderAlignResolved and HeaderFontFamilyResolved each appeared in one. Those
// four carry AC3 — *the editor shows the value the document will actually use,
// because the engine sent it* — and wiring any of them straight to the
// COMMITTED value, or to a constant, left the whole suite green.
//
// Each row walks BOTH legs of the cascade so a member wired to either end is
// caught: put the value on the TABLE's own style and the resolved twin must
// follow it; then put a different value on `headerStyle` and the resolved twin
// must follow THAT instead. A member wired to the committed field fails the
// first leg; one wired to the table's own style fails the second.
func TestEveryResolvedProjectionMemberFollowsBothLegsOfTheCascade(t *testing.T) {
	rows := []struct {
		field       string
		member      string
		tableStyle  string
		headerValue string
		wantTable   string
		wantHeader  string
		read        func(TableColumnsProjection) string
	}{
		{"fontFamily", "HeaderFontFamilyResolved", `"fontFamily": "body"`, `"display"`, "body", "display",
			func(v TableColumnsProjection) string { return v.HeaderFontFamilyResolved }},
		{"fontSize", "HeaderFontSizeResolved", `"fontSize": 9`, `14`, "9000", "14000",
			func(v TableColumnsProjection) string { return strconv.FormatInt(v.HeaderFontSizeResolved, 10) }},
		{"lineSpacing", "HeaderLineSpacingResolved", `"lineSpacing": 1.5`, `2`, "1500", "2000",
			func(v TableColumnsProjection) string { return strconv.FormatInt(v.HeaderLineSpacingResolved, 10) }},
		{"background", "HeaderBackgroundResolved", `"background": "#eeeeee"`, `"#101010"`, "#eeeeee", "#101010",
			func(v TableColumnsProjection) string { return v.HeaderBackgroundResolved }},
		{"color", "HeaderColorResolved", `"color": "#1b2a4a"`, `"#c81e1e"`, "#1b2a4a", "#c81e1e",
			func(v TableColumnsProjection) string { return v.HeaderColorResolved }},
		{"valign", "HeaderValignResolved", `"valign": "middle"`, `"bottom"`, "middle", "bottom",
			func(v TableColumnsProjection) string { return v.HeaderValignResolved }},
		{"align", "HeaderAlignResolved", `"align": "center"`, `"right"`, "center", "right",
			func(v TableColumnsProjection) string { return v.HeaderAlignResolved }},
		// STORY 11.3's TWO. Leg one puts `true` on the table's own style and
		// leg two puts `false` on headerStyle, so the second leg is the one a
		// member wired to `style.bold` cannot pass — the direction matters for
		// a bool, because a leg-two value of `true` would also be produced by
		// a member that simply echoed leg one.
		{"bold", "HeaderBoldResolved", `"bold": true`, `false`, "true", "false",
			func(v TableColumnsProjection) string { return strconv.FormatBool(v.HeaderBoldResolved) }},
		{"italic", "HeaderItalicResolved", `"italic": true`, `false`, "true", "false",
			func(v TableColumnsProjection) string { return strconv.FormatBool(v.HeaderItalicResolved) }},
	}
	// THE TIE DW-240 WAS MISSING, AND THE DURABLE HALF OF ITS FIX. Nothing
	// related `tableHeaderStyleFields` to this projection's member list, which
	// is exactly why nobody noticed that the command layer accepted nine
	// header-style fields while the projection carried seven pairs: a weight
	// an author could write and could not read back. `rows` above is a
	// hand-maintained list and was itself the second copy of that asymmetry,
	// so it is pinned to the engine's own closed set here rather than counted.
	// A tenth field added to `tableHeaderStyleFields` now reds THIS test until
	// its pair and its row exist.
	fields := make([]string, 0, len(rows))
	for _, row := range rows {
		fields = append(fields, row.field)
	}
	slices.Sort(fields)
	want := slices.Clone(tableHeaderStyleFields)
	slices.Sort(want)
	if !slices.Equal(fields, want) {
		t.Errorf("this test walks the header-style fields\n\t%v\nand the engine's closed set is\n\t%v — a field a command can author with no row here is a field whose resolved projection member nothing follows, which is DW-240 repeating", fields, want)
	}
	for _, row := range rows {
		t.Run(row.member, func(t *testing.T) {
			// LEG ONE: the value lives on the table's own style and there is no
			// headerStyle at all, so the resolved twin must read the table's.
			tpl := headerCascadeDocument(t, `"style": {`+row.tableStyle+`},`, "")
			view, err := TableColumns(tpl, "e1")
			if err != nil {
				t.Fatalf("project the table: %v", err)
			}
			if got := row.read(view); got != row.wantTable {
				t.Errorf("%s = %q with the value on the table's own style, want %q — the resolved member is not reading the engine's cascade", row.member, got, row.wantTable)
			}
			// And the COMMITTED twin beside it is still absent, which is what
			// makes the two members different members rather than one repeated.
			// "false" JOINS "" AND "0" AS A SPELLING OF ABSENT, and that is
			// the disclosed collapse rather than a hole in this check: a bool
			// on a wire whose key set is pinned exactly in both directions has
			// no third value to put an absence in, so committed-absent and
			// committed-`false` are the same member value
			// (TableColumnsProjection's own comment states the limit). Leg two
			// below is where the two booleans are still discriminating, and it
			// sets headerStyle to `false` over a table style of `true`.
			if committed := committedTwin(t, view, row.field); committed != "" && committed != "0" && committed != "false" {
				t.Errorf("committed %s = %q with nothing on headerStyle, want absent", row.field, committed)
			}
			// LEG TWO: headerStyle now declares a DIFFERENT value, so the
			// resolved twin must abandon the table's style for it.
			if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateTableHeaderStyle","version":1,"id":"e1","field":"`+row.field+`","op":"set","value":`+row.headerValue+`}`)); err != nil {
				t.Fatalf("set headerStyle.%s: %v", row.field, err)
			}
			view, err = TableColumns(tpl, "e1")
			if err != nil {
				t.Fatalf("re-project the table: %v", err)
			}
			if got := row.read(view); got != row.wantHeader {
				t.Errorf("%s = %q with headerStyle.%s declared, want %q — headerStyle must beat the table's own style", row.member, got, row.field, row.wantHeader)
			}
			// AND BACK: clearing it returns the resolved twin to the table's
			// own value, which is the row of the I/O matrix this whole story
			// turns on, asserted here once per member rather than once total.
			if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateTableHeaderStyle","version":1,"id":"e1","field":"`+row.field+`","op":"clear"}`)); err != nil {
				t.Fatalf("clear headerStyle.%s: %v", row.field, err)
			}
			view, err = TableColumns(tpl, "e1")
			if err != nil {
				t.Fatalf("re-project the table after the clear: %v", err)
			}
			if got := row.read(view); got != row.wantTable {
				t.Errorf("%s = %q after the clear, want the table's own %q back", row.member, got, row.wantTable)
			}
		})
	}
}

// TestTheProjectionCarriesAPairForEveryHeaderStyleFieldACommandCanWrite is the
// OTHER half of DW-240's tie, and the durable one: it reads the projection's
// member list off the STRUCT rather than off a hand-maintained list, so it
// holds even for a field nobody remembered to add a cascade row for.
//
// DW-240 WAS AN ASYMMETRY NOTHING COULD SEE. `tableHeaderStyleFields` (the
// authoring half) went to nine at Story 11.2 while TableColumnsProjection kept
// seven Header*/Header*Resolved pairs, and no test in this repository related
// the two — so a header weight an author could write was a header weight the
// panel could not read back, silently, for a whole story. The two members were
// the cheap half of the fix; this is the half that stops the tenth field
// repeating it.
//
// IT IS READ FROM THE JSON TAGS, NOT THE Go NAMES, because the tags are what
// the browser's exact-key guard pins and therefore what "a member" means on
// this seam. `headerHeight` is the one `header*` key with no `Resolved` twin,
// and that is asserted rather than skipped: the struct's own comment calls the
// singleton deliberate, so the test states which key it is instead of letting
// any unpaired key through.
func TestTheProjectionCarriesAPairForEveryHeaderStyleFieldACommandCanWrite(t *testing.T) {
	tags := map[string]bool{}
	value := reflect.TypeOf(TableColumnsProjection{})
	for i := 0; i < value.NumField(); i++ {
		tag, _, _ := strings.Cut(value.Field(i).Tag.Get("json"), ",")
		if tag != "" {
			tags[tag] = true
		}
	}
	// POSITIVE CONTROL: the reflection actually read something, and read the
	// key set the wire record already pins. A typo'd tag name would otherwise
	// make every absence below vacuous.
	if len(tags) != len(tableColumnsProjectionWireKeys) {
		t.Fatalf("read %d json tags off TableColumnsProjection and the recorded wire set has %d — this test is measuring the wrong struct", len(tags), len(tableColumnsProjectionWireKeys))
	}
	var paired, unpaired []string
	for tag := range tags {
		field, ok := strings.CutPrefix(tag, "header")
		// A tag spelled EXACTLY `header` would take `field[:1]` on an empty
		// string and PANIC — a test that crashes instead of reporting, on the
		// one input this loop cannot name. It is not reachable today and it is
		// guarded rather than argued: the guard costs a line, and "unreachable"
		// is the claim that ages worst in this repository.
		if !ok || field == "" || strings.HasSuffix(tag, "Resolved") {
			continue
		}
		name := strings.ToLower(field[:1]) + field[1:]
		if tags[tag+"Resolved"] {
			paired = append(paired, name)
			continue
		}
		unpaired = append(unpaired, name)
	}
	// BOTH LISTS ARE SORTED, because both are built by ranging a MAP and Go
	// randomises that order deliberately. `paired` was sorted and `unpaired` was
	// not, so a second unpaired key would have made this test's verdict depend
	// on the run — green sometimes, red sometimes, over the same code.
	slices.Sort(paired)
	slices.Sort(unpaired)
	want := slices.Clone(tableHeaderStyleFields)
	slices.Sort(want)
	if !slices.Equal(paired, want) {
		t.Errorf("TableColumnsProjection carries a committed/resolved pair for\n\t%v\nand the fields a command may author are\n\t%v — a field on one side only is a value an author can write and cannot read back (DW-240), or a member the engine sends that no command can produce", paired, want)
	}
	if !slices.Equal(unpaired, []string{"height"}) {
		t.Errorf("the header keys with no resolved twin are %v, want exactly [height] — headerHeight is the one deliberate singleton (it is required by the format, so it is never absent and committed IS resolved); any other unpaired header key is a missing pair", unpaired)
	}
}

// committedTwin reads the COMMITTED member beside a resolved one, as a string,
// so the test above can assert the pair are genuinely two members.
func committedTwin(t *testing.T, view TableColumnsProjection, field string) string {
	t.Helper()
	switch field {
	case "fontFamily":
		return view.HeaderFontFamily
	case "fontSize":
		return strconv.FormatInt(view.HeaderFontSize, 10)
	case "lineSpacing":
		return strconv.FormatInt(view.HeaderLineSpacing, 10)
	case "background":
		return view.HeaderBackground
	case "color":
		return view.HeaderColor
	case "valign":
		return view.HeaderValign
	case "align":
		return view.HeaderAlign
	case "bold":
		return strconv.FormatBool(view.HeaderBold)
	case "italic":
		return strconv.FormatBool(view.HeaderItalic)
	}
	t.Fatalf("no committed twin for %q", field)
	return ""
}

// cleanupEmptyHeaderStyle's PROTECTION FOR THE FIELDS NO COMMAND CAN AUTHOR.
//
// The predicate consults Bold, Italic, Border, Padding and Extra as well as the
// seven authorable fields, and nothing tested that. Shortening it to the seven
// left every test green while an author who had hand-written a header border
// lost it the moment they touched header alignment — the block would go from
// "declares align and border" to "declares border" to, wrongly, absent.
//
// Precedent for the shape: line_spacing_test.go's
// TestLineSpacingClearedFromTheOnlyStyleFieldStrandsNoStyleBlock, which pins
// the same predicate's other direction on element.Style.
func TestClearingTheLastAuthorableHeaderFieldKeepsAHandAuthoredBorder(t *testing.T) {
	tpl := headerCascadeDocument(t, "", `"headerStyle": {"align": "center", "border": {"edges": ["bottom"], "color": "#112233", "width": 1}},`)
	if got := projectHeaderCascade(t, tpl).HeaderAlign; got != "center" {
		t.Fatalf("precondition: the hand-authored align must project, got %q", got)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateTableHeaderStyle","version":1,"id":"e1","field":"align","op":"clear"}`)); err != nil {
		t.Fatalf("clear headerStyle.align: %v", err)
	}
	encoded, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	// The cleared field is gone.
	if bytes.Contains(encoded, []byte(`"align": "center"`)) {
		t.Errorf("the cleared align survived:\n%s", encoded)
	}
	// AND THE BLOCK IS STILL THERE, because it still declares something.
	if !bytes.Contains(encoded, []byte(`"headerStyle"`)) {
		t.Errorf("clearing the last AUTHORABLE field dropped a headerStyle that still declares a border:\n%s", encoded)
	}
	if !bytes.Contains(encoded, []byte(`"border"`)) {
		t.Errorf("the hand-authored header border was deleted by a command that was asked about alignment:\n%s", encoded)
	}
	// And the document still loads, which is what "survives in the canonical
	// bytes" has to mean for a block the loader will read back.
	if _, err := ParseTemplate(encoded); err != nil {
		t.Fatalf("the surviving document no longer loads: %v", err)
	}
}

func projectHeaderCascade(t *testing.T, tpl *Template) TableColumnsProjection {
	t.Helper()
	view, err := TableColumns(tpl, "e1")
	if err != nil {
		t.Fatalf("project e1: %v", err)
	}
	return view
}

// A CLEAR AGAINST AN EXPLICITLY-NULL headerStyle IS A NO-OP, and a no-op must
// not move the bytes. headerStyleFor replaces the null with a real block and
// cleanupEmptyHeaderStyle then removes the key outright, so `"headerStyle":
// null` became no headerStyle at all — a byte change and an undo entry for a
// command that removed a field which was already absent.
func TestClearingAFieldOfAnExplicitlyNullHeaderStyleMovesNoBytes(t *testing.T) {
	tpl := headerCascadeDocument(t, "", `"headerStyle": null,`)
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !bytes.Contains(before, []byte(`"headerStyle": null`)) {
		t.Fatalf("precondition: the fixture must carry an explicit null headerStyle:\n%s", before)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"updateTableHeaderStyle","version":1,"id":"e1","field":"align","op":"clear"}`)); err != nil {
		t.Fatalf("clear headerStyle.align on a null block: %v", err)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatalf("re-serialize: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("a clear of an already-absent field changed the document:\nbefore\n%s\nafter\n%s", before, after)
	}
}

// setTableHeaderHeight's ARITY REFUSAL IS LOCATED, like every other refusal in
// this change. It returned componentFields' bare fmt.Errorf, which reaches the
// author as ENGINE_REJECTED with no element and no field to look at — the one
// unlocated refusal across the three arms, and the exact thing
// refusalLeavesTheDocumentAlone's own comment warns about.
func TestSetTheHeaderHeightLocatesItsArityRefusal(t *testing.T) {
	tpl := headerStyleFixture(t)
	const want = "setTableHeaderHeight takes exactly kind, version, id and height"
	// A surplus key.
	refusalSaysWhy(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+theWorkedExampleTable+`","height":18,"surplus":1}`, "table.headerHeight", want)
	// And a missing one.
	refusalSaysWhy(t, tpl, `{"kind":"setTableHeaderHeight","version":1,"id":"`+theWorkedExampleTable+`"}`, "table.headerHeight", want)
}
