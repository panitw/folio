package folio

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

// STORY 12.1: THE PAGE HEADER AND PAGE FOOTER TAKE THE HEIGHT THE AUTHOR SETS.
//
// Band.Height had a loader, two consumers and NO WRITER. Everything below is
// about the writer: what it accepts, what it refuses, whether the refusal is
// located, and whether the document is byte-unchanged afterwards.
//
// THE FIXTURES ARE WRITTEN OUT HERE rather than taken from the golden corpus
// because every row of this story's matrix is a statement about a SPECIFIC
// (header, footer, element) triple, and a fixture that drifts turns "refused at
// 79" into "refused at whatever the corpus says today".
//
// NUMBERS ARE READ BACK AS RAW JSON LITERALS, never decoded into a Go numeric
// type: internal/arch_test.go's TestNoFloat64UnderModule scans _test.go files
// too, and `json.Unmarshal` into `any` is exactly how a float64 gets into a
// test that never spells one.

// A4 portrait at 36pt margins: 841.89 − 36 − 36 = 769.89pt of printable column,
// which is the number every content-window row below is measured against.
// A4 portrait at 36pt margins: 841.89 - 36 - 36 = 769.89pt of printable column,
// which is the number every content-window row below is measured against. IT IS
// SPELLED IN POINTS because that is what a refusal quotes: the author types
// points into a box labelled "(pt)" and every derived quantity in the sentence
// they read back is in the same units (Story 12.1 review, patch 6).
const bandHeightInnerColumn = "769.89pt"

// bandHeightDocumentWith is the one fixture builder, and it takes BOTH element
// lists because every claim in this file has to be made twice — once of the
// page header and once of the page footer. A suite that only ever populated the
// header let a rotation that made a pageFooter command write PageHeader through
// green (see the rotation red-proof at the foot of this file).
func bandHeightDocumentWith(headerHeight, footerHeight string, headerElements, footerElements []string) []byte {
	return []byte(`{"version":"1.0","locale":"th","utcOffset":"+07:00",` +
		`"page":{"margin":{"top":36,"right":36,"bottom":36,"left":36},"orientation":"portrait","size":"A4"},` +
		`"fonts":{},"bands":{"content":{"elements":[]},` +
		`"pageFooter":{"elements":[` + strings.Join(footerElements, ",") + `],"height":` + footerHeight + `},` +
		`"pageHeader":{"elements":[` + strings.Join(headerElements, ",") + `],"height":` + headerHeight + `}},` +
		`"assets":{},"nextId":9}`)
}

func bandHeightDocument(headerHeight, footerHeight string, headerElements ...string) []byte {
	return bandHeightDocumentWith(headerHeight, footerHeight, headerElements, nil)
}

// bandHeightFooterDocument is bandHeightDocument's mirror: the same page, the
// same two heights, and the elements in the PAGE FOOTER instead.
func bandHeightFooterDocument(headerHeight, footerHeight string, footerElements ...string) []byte {
	return bandHeightDocumentWith(headerHeight, footerHeight, nil, footerElements)
}

// bandHeightRect is a rect, deliberately: it carries its own box and needs no
// font chain, so a fixture element never brings a second refusal path with it.
func bandHeightRect(id, y, height string) string {
	return `{"height":` + height + `,"id":"` + id + `","type":"rect","width":100,"x":0,"y":` + y + `}`
}

func bandHeightTemplate(t *testing.T, doc []byte) *Template {
	t.Helper()
	tpl, err := ParseTemplate(doc)
	if err != nil {
		t.Fatalf("fixture precondition: the document must load before the command can be measured: %v", err)
	}
	return tpl
}

// bandHeightRefusal asserts the two halves of every refusal row at once: the
// command was refused, and the document is byte-identical afterwards (AC3, and
// the "atomic" clause of the Boundaries section).
func bandHeightRefusal(t *testing.T, tpl *Template, command string) *ComponentCommandError {
	t.Helper()
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if _, applyErr := ApplyComponentCommand(tpl, []byte(command)); applyErr == nil {
		t.Fatalf("command unexpectedly succeeded: %s", command)
	} else {
		after, serErr := SerializeTemplate(tpl)
		if serErr != nil {
			t.Fatal(serErr)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("a refused band-height command mutated the document:\n%s\nbefore: %s\nafter:  %s", command, before, after)
		}
		var failure *ComponentCommandError
		if !errors.As(applyErr, &failure) {
			t.Fatalf("refusal for %s is %T (%v), want *ComponentCommandError — an unlocated refusal reaches the author as the fixed page-setup sentence", command, applyErr, applyErr)
		}
		return failure
	}
	return nil
}

// bandHeightKey returns the RAW JSON literal a band's `height` key serializes
// to, and whether the key is present at all. Raw bytes rather than a decoded
// number: the assertion is about what is IN the file.
func bandHeightKey(t *testing.T, canonical []byte, band string) (string, bool) {
	t.Helper()
	var document map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &document); err != nil {
		t.Fatal(err)
	}
	var bands map[string]json.RawMessage
	if err := json.Unmarshal(document["bands"], &bands); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(bands[band], &fields); err != nil {
		t.Fatalf("no %s band in the serialized document: %v", band, err)
	}
	height, ok := fields["height"]
	return string(height), ok
}

// TestSetBandHeightWritesTheHeightTheAuthorSet is the first row of the matrix
// and the whole point of the story: a header height that nothing in the product
// could set before, set, projected, and serialized.
func TestSetBandHeightWritesTheHeightTheAuthorSet(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("20", "30"))
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80}`))
	if err != nil {
		t.Fatalf("a header height with nothing in the band was refused: %v", err)
	}
	if projection.Bands[0].Name != bandPageHeader || projection.Bands[0].Height != 80000 {
		t.Fatalf("projected pageHeader = %#v, want name pageHeader and height 80000", projection.Bands[0])
	}
	// The CONTENT band's derived height moved with it, which is the half a
	// test asserting only "some byte changed" would not have noticed: the one
	// derivation (layout.ContentHeight) was re-run over the new geometry.
	if projection.Bands[1].Name != bandContent || projection.Bands[2].Name != bandPageFooter {
		t.Fatalf("band order changed: %#v", projection.Bands)
	}
	if projection.Bands[1].Height != 769890-80000-30000 {
		t.Fatalf("derived content height = %d, want %d", projection.Bands[1].Height, 769890-80000-30000)
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if literal, ok := bandHeightKey(t, canonical, "pageHeader"); !ok || literal != "80" {
		t.Fatalf("serialized pageHeader.height = %q (present %v), want 80 — AC5's already-shipped writeBand, testable for the first time", literal, ok)
	}
	// AC5's other half in the same breath: the footer was not touched and the
	// content band did not GAIN a height key from a command aimed at a sibling.
	if literal, ok := bandHeightKey(t, canonical, "pageFooter"); !ok || literal != "30" {
		t.Fatalf("serialized pageFooter.height = %q (present %v), want 30", literal, ok)
	}
	if literal, ok := bandHeightKey(t, canonical, "content"); ok {
		t.Fatalf("the content band gained a height key %q from a pageHeader command", literal)
	}
}

// TestSetBandHeightAcceptsTheLowestOccupiedEdge is the BOUNDARY, and it is a
// separate test from the strand because the two are one millipoint apart and a
// guard written with the wrong comparison passes every other row in the matrix.
func TestSetBandHeightAcceptsTheLowestOccupiedEdge(t *testing.T) {
	// e1 occupies y=50 through y=80. A band exactly 80pt tall contains it:
	// y + height == height is INSIDE, not outside.
	tpl := bandHeightTemplate(t, bandHeightDocument("100", "30", bandHeightRect("e1", "50", "30")))
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80}`)); err != nil {
		t.Fatalf("shortening to exactly the lowest occupied edge was refused: %v", err)
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if literal, _ := bandHeightKey(t, canonical, "pageHeader"); literal != "80" {
		t.Fatalf("serialized pageHeader.height = %q, want 80", literal)
	}
}

// TestSetBandHeightRefusesToStrandAComponent is the refusal one millipoint
// below the boundary above, and it asserts the LOCATED fields rather than only
// that something was refused: an unlocated refusal reaches the author as
// pageSetupDiagnostic's fixed sentence and the height they typed is never named.
func TestSetBandHeightRefusesToStrandAComponent(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("100", "30", bandHeightRect("e1", "50", "30")))
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":79}`)
	if failure.ElementID != "e1" {
		t.Errorf("strand refusal ElementID = %q, want e1 — the element that would be stranded is the located half", failure.ElementID)
	}
	if failure.DataPath != "bands.pageHeader.height" {
		t.Errorf("strand refusal DataPath = %q, want bands.pageHeader.height", failure.DataPath)
	}
	// THE REFUSAL NAMES THE ACT, NOT THE OBJECT. containComponent's own
	// sentence answers "you moved a component", which is not what happened.
	if strings.Contains(failure.Message, "component geometry must stay within") {
		t.Errorf("the refusal borrowed containComponent's sentence: %q", failure.Message)
	}
	if !strings.Contains(failure.Message, "79") || !strings.Contains(failure.Message, "e1") {
		t.Errorf("strand refusal = %q, want it to name the height 79 and the element e1", failure.Message)
	}
	// UNITS (patch 6): the reach is the element's lowest edge, 80pt, and it is
	// spelled in the points the author typed rather than as 80000mp.
	if !strings.Contains(failure.Message, "80pt") {
		t.Errorf("strand refusal = %q, want the reach spelled as 80pt", failure.Message)
	}
	assertPointsNotMillipoints(t, failure.Message)
}

// TestSetBandHeightRefusesToStrandAComponentThatIsNotTheFirst is THE VACUITY
// GUARD the lead asked for. A strand check that examined only band.Elements[0]
// passes every other row in this file: e1 at y=0 h=10 fits a 40pt band, so a
// first-element-only check accepts a height that strands e2.
func TestSetBandHeightRefusesToStrandAComponentThatIsNotTheFirst(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("100", "30",
		bandHeightRect("e1", "0", "10"),
		bandHeightRect("e2", "50", "30"),
	))
	// Precondition, stated rather than assumed: the FIRST element really does
	// fit the proposed height, so this test can only pass by looking past it.
	if _, err := ApplyComponentCommand(bandHeightTemplate(t, bandHeightDocument("100", "30", bandHeightRect("e1", "0", "10"))), []byte(`{"kind":"setBandHeight","version":1,"band":"pageHeader","height":40}`)); err != nil {
		t.Fatalf("fixture precondition: e1 alone must FIT a 40pt band, else the not-the-first case is not being measured: %v", err)
	}
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":40}`)
	if failure.ElementID != "e2" {
		t.Errorf("strand refusal ElementID = %q, want e2 — the predicate must evaluate EVERY element in the band", failure.ElementID)
	}
	if !strings.Contains(failure.Message, "e2") {
		t.Errorf("strand refusal = %q, want it to name e2", failure.Message)
	}
}

// TestSetBandHeightRefusesAHeightThatLeavesNoContentWindow is the command's
// half of the shared predicate. Its sibling is
// TestCanvasRefusesBandsThatLeaveNoContentWindow below: two audiences, two
// sentences, ONE bandsLeaveContentWindow.
func TestSetBandHeightRefusesAHeightThatLeavesNoContentWindow(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("20", "400"))
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":400}`)
	if failure.ElementID != "" {
		t.Errorf("a content-window refusal is not addressed to an element, got ElementID %q", failure.ElementID)
	}
	if failure.DataPath != "bands.pageHeader.height" {
		t.Errorf("content-window refusal DataPath = %q, want bands.pageHeader.height", failure.DataPath)
	}
	if !strings.Contains(failure.Message, "400") {
		t.Errorf("content-window refusal = %q, want it to name the height 400 that was refused", failure.Message)
	}
	if !strings.Contains(failure.Message, bandHeightInnerColumn) {
		t.Errorf("content-window refusal = %q, want it to name the %s of printable column available", failure.Message, bandHeightInnerColumn)
	}
	// UNITS (patch 6). The author types points and must read points: a bound
	// quoted as 369889mp beside a height quoted as 400pt is four digits wider
	// than the number it is about, and it is the one quantity in the sentence
	// they would have to type back.
	assertPointsNotMillipoints(t, failure.Message)
	if !strings.Contains(failure.Message, "400pt") {
		t.Errorf("content-window refusal = %q, want the OTHER band's 400pt height spelled in points too", failure.Message)
	}
}

// assertPointsNotMillipoints is the units assertion every refusal in this file
// makes. "mp" cannot appear in a legal band name or in a points literal, so its
// presence anywhere in a message is a millipoint quantity that escaped.
func assertPointsNotMillipoints(t *testing.T, message string) {
	t.Helper()
	if strings.Contains(message, "mp") {
		t.Errorf("refusal quotes a millipoint quantity: %q — every number the author could type back must be in the points they typed", message)
	}
}

// TestSetBandHeightNamesTheBoundThePredicateAcceptsUpTo is patch 2's assertion,
// and it is deliberately made through the COMMAND rather than by calling the
// predicate twice: the message is parsed for the bound it names, and that exact
// number is then sent back as a height. If the sentence and the check ever
// disagree by one millipoint, one of the two commands below flips.
func TestSetBandHeightNamesTheBoundThePredicateAcceptsUpTo(t *testing.T) {
	failure := bandHeightRefusal(t, bandHeightTemplate(t, bandHeightDocument("20", "400")), `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":400}`)
	bound := boundNamedIn(t, failure.Message)
	// 769.89 - 400 - 0.001. Stated so the parse below cannot pass by finding
	// some other number in the sentence.
	if bound != "369.889" {
		t.Fatalf("the refusal names %q as the largest legal header height; the predicate accepts up to 369.889", bound)
	}
	if _, err := ApplyComponentCommand(bandHeightTemplate(t, bandHeightDocument("20", "400")), []byte(`{"kind":"setBandHeight","version":1,"band":"pageHeader","height":`+bound+`}`)); err != nil {
		t.Fatalf("the height the refusal named as legal was refused: %v", err)
	}
	// And one millipoint MORE is refused, so the named bound is the LARGEST
	// accepted value and not merely some accepted value.
	bandHeightRefusal(t, bandHeightTemplate(t, bandHeightDocument("20", "400")), `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":369.89}`)
}

// boundNamedIn extracts the number the content-window refusal names as the
// ceiling, from the message's own words. It reads the sentence the author
// reads; nothing here re-derives the arithmetic.
func boundNamedIn(t *testing.T, message string) string {
	t.Helper()
	const marker = "must be at most "
	at := strings.Index(message, marker)
	if at < 0 {
		t.Fatalf("the content-window refusal names no ceiling at all: %q", message)
	}
	rest := message[at+len(marker):]
	end := strings.Index(rest, "pt")
	if end < 0 {
		t.Fatalf("the ceiling in %q is not spelled in points", message)
	}
	return rest[:end]
}

// TestSetBandHeightRefusesBandsThatExactlyFillTheColumn: content must be
// STRICTLY positive. 369.89 + 400 == 769.89 exactly, and a document whose
// content band is zero millipoints tall has nowhere for content to be.
func TestSetBandHeightRefusesBandsThatExactlyFillTheColumn(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("20", "400"))
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":369.89}`)
	if failure.DataPath != "bands.pageHeader.height" {
		t.Errorf("exact-fill refusal DataPath = %q, want bands.pageHeader.height", failure.DataPath)
	}
	// One millipoint less IS accepted, which is what makes the row above a
	// statement about strictness rather than about arithmetic being wrong.
	fresh := bandHeightTemplate(t, bandHeightDocument("20", "400"))
	if _, err := ApplyComponentCommand(fresh, []byte(`{"kind":"setBandHeight","version":1,"band":"pageHeader","height":369.889}`)); err != nil {
		t.Fatalf("one millipoint short of exactly filling the column was refused: %v", err)
	}
}

// TestSetBandHeightRefusesTheContentBand is what keeps `height` absent from
// every serialized content band: writeBand emits the key on Set alone, so a
// content band written even once would carry one forever.
func TestSetBandHeightRefusesTheContentBand(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("20", "30"))
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"content","height":100}`)
	if failure.DataPath != "bands.content.height" {
		t.Errorf("content-band refusal DataPath = %q, want bands.content.height", failure.DataPath)
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if literal, ok := bandHeightKey(t, canonical, "content"); ok {
		t.Fatalf("the content band gained a height key %q from a refused command", literal)
	}
}

// TestSetBandHeightRefusesTheRemainingMatrixRows covers the rows whose whole
// content is "refused, and located": a negative height, a band name that is not
// a band, and the door's own arity and duplicate-key gates.
func TestSetBandHeightRefusesTheRemainingMatrixRows(t *testing.T) {
	for _, probe := range []struct {
		name     string
		command  string
		dataPath string
	}{
		{
			name:     "a negative height",
			command:  `{"kind":"setBandHeight","version":1,"band":"pageFooter","height":-5}`,
			dataPath: "bands.pageFooter.height",
		},
		{
			name:     "a band name that is not a band",
			command:  `{"kind":"setBandHeight","version":1,"band":"footer","height":20}`,
			dataPath: "bands.footer.height",
		},
		{
			name:     "an empty band name",
			command:  `{"kind":"setBandHeight","version":1,"band":"","height":20}`,
			dataPath: "bands",
		},
		{
			name:     "a height that is not a number",
			command:  `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":"80"}`,
			dataPath: "bands.pageHeader.height",
		},
		{
			name:     "a height with four decimal places",
			command:  `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80.0001}`,
			dataPath: "bands.pageHeader.height",
		},
		{
			// THE EMPTIED BOX. band-height-command.test.ts pins that a row the
			// author cleared reaches the wire as `"height":null` rather than as
			// a 0 nobody typed; this is the other end of that wire, so the
			// browser's claim is pinned against the engine and not only against
			// itself.
			name:     "a null height, the shape an emptied box sends",
			command:  `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":null}`,
			dataPath: "bands.pageHeader.height",
		},
	} {
		t.Run(probe.name, func(t *testing.T) {
			failure := bandHeightRefusal(t, bandHeightTemplate(t, bandHeightDocument("20", "30")), probe.command)
			if failure.DataPath != probe.dataPath {
				t.Errorf("DataPath = %q, want %q", failure.DataPath, probe.dataPath)
			}
			if failure.Message == "" {
				t.Error("a located refusal with no message is the fixed page-setup sentence with extra steps")
			}
		})
	}
}

// TestSetBandHeightIsRefusedByTheDoorsExistingGates keeps the arity and
// duplicate-key rows honest: they are refused by componentFields and
// refuseDuplicateCommandKeys, which are NOT ComponentCommandErrors, so they
// cannot be asserted through bandHeightRefusal's located-ness check.
func TestSetBandHeightIsRefusedByTheDoorsExistingGates(t *testing.T) {
	for _, probe := range []struct {
		name    string
		command string
	}{
		{"three keys", `{"kind":"setBandHeight","version":1,"band":"pageHeader"}`},
		{"five keys", `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80,"snap":false}`},
		{"a repeated band", `{"kind":"setBandHeight","version":1,"band":"content","height":80,"band":"pageHeader"}`},
		{"a repeated height", `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80,"height":9999}`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			tpl := bandHeightTemplate(t, bandHeightDocument("20", "30"))
			before, err := SerializeTemplate(tpl)
			if err != nil {
				t.Fatal(err)
			}
			if _, applyErr := ApplyComponentCommand(tpl, []byte(probe.command)); applyErr == nil {
				t.Fatalf("the door accepted %s", probe.command)
			}
			after, err := SerializeTemplate(tpl)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("a command the door refused still mutated the document")
			}
		})
	}
}

// TestSetBandHeightResentUnchangedLeavesTheBytesIdentical is the row the
// designer's difference test exists to make unnecessary, asserted on the Go
// side anyway: re-sending the height already in force writes the same value, so
// the canonical bytes do not move. wasm/engine.go's own short-circuit compares
// exactly these bytes, which is what makes it "no history entry" there.
func TestSetBandHeightResentUnchangedLeavesTheBytesIdentical(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("20", "30"))
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setBandHeight","version":1,"band":"pageHeader","height":20}`)); err != nil {
		t.Fatalf("re-sending the height already in force was refused: %v", err)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("re-sending the current height moved the bytes:\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestUneditedDocumentSerializesByteIdentically is AC5 in its honest form (see
// the spec's Design Notes): `height` is REQUIRED on the two capping bands and
// FORBIDDEN on content, so the only key that can be absent is content's, and
// what keeps it absent is the refusal above. What is left to assert here is the
// other half — that opening and serializing a document changes nothing.
func TestUneditedDocumentSerializesByteIdentically(t *testing.T) {
	// A document that ALREADY STRANDS a component, because that is the case a
	// new check is most likely to have made unloadable: e1 runs to y=80 in a
	// 60pt band. Nothing may refuse it at load; only the WRITER refuses.
	document := bandHeightDocument("60", "30", bandHeightRect("e1", "50", "30"))
	tpl := bandHeightTemplate(t, document)
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	reparsed := bandHeightTemplate(t, canonical)
	again, err := SerializeTemplate(reparsed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonical, again) {
		t.Fatalf("an unedited document did not round-trip:\nfirst:  %s\nsecond: %s", canonical, again)
	}
	if _, ok := bandHeightKey(t, canonical, "content"); ok {
		t.Fatal("the content band carries a height key it was never given")
	}
	// And the already-stranded document still PROJECTS, which is what lets the
	// designer open it at all — the strand is a writer's refusal, not a
	// loader's. If this ever fails, the panel's difference test (App.tsx) is
	// protecting a document nobody can open.
	if _, err := Canvas(tpl); err != nil {
		t.Fatalf("an already-stranded document lost its projection: %v", err)
	}
}

// TestCanvasRefusesBandsThatLeaveNoContentWindow is the OTHER caller of
// bandsLeaveContentWindow — the hand-edit backstop Q3 kept. It guards LOADING
// and owes only a refusal to project; the command guards AUTHORING and owes a
// located message. This test and
// TestSetBandHeightRefusesAHeightThatLeavesNoContentWindow are the pair that go
// red together when the shared predicate is changed, which is the evidence that
// the invariant has one implementation and not two.
func TestCanvasRefusesBandsThatLeaveNoContentWindow(t *testing.T) {
	for _, probe := range []struct {
		name           string
		header, footer string
		// sentence is the refusal expected, or "" for a document that projects.
		// It is written out per row rather than assumed, because two of these
		// five are refused by a DIFFERENT gate — see the negative rows.
		sentence string
	}{
		{"a header and footer that overrun the column", "400", "400", "page setup leaves no positive content region"},
		{"a header and footer that exactly fill it", "369.89", "400", "page setup leaves no positive content region"},
		{"one millipoint short of exactly filling it", "369.889", "400", ""},
		// MEASURED, AND RECORDED BECAUSE IT IS NOT WHAT IT LOOKS LIKE: the
		// geometry-bound loop above the band gate tests `v < 0` over the header
		// and footer too, so a negative band height never reaches
		// bandsLeaveContentWindow through Canvas at all. Its `header >= 0`
		// terms are carried unchanged from the expression they were extracted
		// from — Q3's condition is that the predicate move as one piece, not
		// that every term of it be reachable from both callers. The command
		// refuses a negative height with its own located sentence before it
		// asks the predicate anything.
		{"a negative header", "-1", "30", "page setup exceeds the JavaScript-safe geometry bound"},
		{"a negative footer", "20", "-1", "page setup exceeds the JavaScript-safe geometry bound"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			// The loader accepts all five: parse_bands.go has no positivity
			// check and no upper bound, which is why Canvas has to hold one.
			tpl := bandHeightTemplate(t, bandHeightDocument(probe.header, probe.footer))
			_, err := Canvas(tpl)
			if probe.sentence == "" {
				if err != nil {
					t.Fatalf("Canvas refused a projectable document: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Canvas projected a document with no content window")
			}
			if !strings.Contains(err.Error(), probe.sentence) {
				t.Fatalf("Canvas refusal = %q, want %q — its own bare sentence, because the loader's audience is not the author's", err, probe.sentence)
			}
			var located *ComponentCommandError
			if errors.As(err, &located) {
				t.Fatal("Canvas's refusal became located; the two audiences share a predicate, not a sentence")
			}
		})
	}
}

// TestSetBandHeightNamesTheFieldForAnEmptiedBox is patch 9's other half: the
// wire shape is pinned in band-height-command.test.ts, and what the ENGINE says
// about it is pinned here. An emptied row must come back naming the field, not
// naming the engine's own parser.
func TestSetBandHeightNamesTheFieldForAnEmptiedBox(t *testing.T) {
	failure := bandHeightRefusal(t, bandHeightTemplate(t, bandHeightDocument("20", "30")), `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":null}`)
	if failure.DataPath != "bands.pageHeader.height" {
		t.Errorf("null-height refusal DataPath = %q, want bands.pageHeader.height", failure.DataPath)
	}
	// The message the engine ACTUALLY produces, pinned. "height must be a
	// number" names the field the author emptied; a message about the engine's
	// own parser (`invalid numeric literal ""`) would not.
	if failure.Message != "height must be a number" {
		t.Errorf("null-height refusal = %q, want %q", failure.Message, "height must be a number")
	}
}

// ---------------------------------------------------------------------------
// THE PAGE FOOTER. Every claim above is made again here, of the OTHER band.
//
// This section is not symmetry for its own sake. Deleting the band-pointer swap
// in setBandHeight — so a pageFooter command writes PageHeader — left the whole
// suite green while every accepting, strand, boundary and content-window test
// used pageHeader and the only pageFooter command in the file was the negative
// row, which returns above the swap. The rotation red-proof at the foot of this
// file is what these four tests exist to make bite.

// TestSetBandHeightWritesTheFooterHeightTheAuthorSet asserts BOTH halves of the
// swap: the footer's serialized literal moved, and the header's did not.
func TestSetBandHeightWritesTheFooterHeightTheAuthorSet(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("20", "30"))
	projection, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setBandHeight","version":1,"band":"pageFooter","height":80}`))
	if err != nil {
		t.Fatalf("a footer height with nothing in the band was refused: %v", err)
	}
	if projection.Bands[2].Name != bandPageFooter || projection.Bands[2].Height != 80000 {
		t.Fatalf("projected pageFooter = %#v, want name pageFooter and height 80000", projection.Bands[2])
	}
	if projection.Bands[0].Name != bandPageHeader || projection.Bands[0].Height != 20000 {
		t.Fatalf("a pageFooter command moved the pageHeader: %#v", projection.Bands[0])
	}
	if projection.Bands[1].Height != 769890-20000-80000 {
		t.Fatalf("derived content height = %d, want %d", projection.Bands[1].Height, 769890-20000-80000)
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if literal, ok := bandHeightKey(t, canonical, "pageFooter"); !ok || literal != "80" {
		t.Fatalf("serialized pageFooter.height = %q (present %v), want 80", literal, ok)
	}
	// THE HALF THE ROTATION HID. "A byte moved" is true of a command that wrote
	// the wrong band; "the footer is 80 AND the header is still 20" is not.
	if literal, ok := bandHeightKey(t, canonical, "pageHeader"); !ok || literal != "20" {
		t.Fatalf("serialized pageHeader.height = %q (present %v), want the untouched 20 — a pageFooter command wrote the header", literal, ok)
	}
}

// TestSetBandHeightAcceptsTheLowestOccupiedEdgeOfTheFooter is the boundary
// acceptance, in the footer.
func TestSetBandHeightAcceptsTheLowestOccupiedEdgeOfTheFooter(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightFooterDocument("30", "100", bandHeightRect("e7", "50", "30")))
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setBandHeight","version":1,"band":"pageFooter","height":80}`)); err != nil {
		t.Fatalf("shortening the footer to exactly its lowest occupied edge was refused: %v", err)
	}
	canonical, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if literal, _ := bandHeightKey(t, canonical, "pageFooter"); literal != "80" {
		t.Fatalf("serialized pageFooter.height = %q, want 80", literal)
	}
	if literal, _ := bandHeightKey(t, canonical, "pageHeader"); literal != "30" {
		t.Fatalf("serialized pageHeader.height = %q, want the untouched 30", literal)
	}
}

// TestSetBandHeightRefusesToStrandAComponentInTheFooter is the strand, one
// millipoint below the boundary above, with the fixture element in the FOOTER.
//
// A rotation that made this command read the header's elements would find an
// EMPTY header band and accept the height: the refusal is the assertion.
func TestSetBandHeightRefusesToStrandAComponentInTheFooter(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightFooterDocument("30", "100", bandHeightRect("e7", "50", "30")))
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"pageFooter","height":79}`)
	if failure.ElementID != "e7" {
		t.Errorf("strand refusal ElementID = %q, want e7", failure.ElementID)
	}
	if failure.DataPath != "bands.pageFooter.height" {
		t.Errorf("strand refusal DataPath = %q, want bands.pageFooter.height", failure.DataPath)
	}
	if !strings.Contains(failure.Message, "pageFooter") || !strings.Contains(failure.Message, "79") {
		t.Errorf("strand refusal = %q, want it to name the pageFooter and the height 79", failure.Message)
	}
	assertPointsNotMillipoints(t, failure.Message)
}

// TestSetBandHeightRefusesAFooterHeightThatLeavesNoContentWindow is the
// content-window refusal asked of the footer, which is the direction that
// exercises the header/footer swap in the predicate's argument order.
func TestSetBandHeightRefusesAFooterHeightThatLeavesNoContentWindow(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightDocument("400", "20"))
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"pageFooter","height":400}`)
	if failure.ElementID != "" {
		t.Errorf("a content-window refusal is not addressed to an element, got ElementID %q", failure.ElementID)
	}
	if failure.DataPath != "bands.pageFooter.height" {
		t.Errorf("content-window refusal DataPath = %q, want bands.pageFooter.height", failure.DataPath)
	}
	if !strings.Contains(failure.Message, bandHeightInnerColumn) {
		t.Errorf("content-window refusal = %q, want it to name the %s of printable column", failure.Message, bandHeightInnerColumn)
	}
	assertPointsNotMillipoints(t, failure.Message)
	// The ceiling is symmetric: the bound on the FOOTER beside a 400pt header
	// is the same 369.889pt, and the sentence must say so for the footer too.
	if bound := boundNamedIn(t, failure.Message); bound != "369.889" {
		t.Fatalf("the footer refusal names %q as its ceiling, want 369.889", bound)
	}
	if _, err := ApplyComponentCommand(bandHeightTemplate(t, bandHeightDocument("400", "20")), []byte(`{"kind":"setBandHeight","version":1,"band":"pageFooter","height":369.889}`)); err != nil {
		t.Fatalf("the footer height the refusal named as legal was refused: %v", err)
	}
}

// TestSetBandHeightRefusesToStrandAComponentThatIsNotTheFirstInTheFooter is the
// vacuity guard, mirrored. Restricting the strand loop to the first element
// must red this as well as its header twin.
func TestSetBandHeightRefusesToStrandAComponentThatIsNotTheFirstInTheFooter(t *testing.T) {
	tpl := bandHeightTemplate(t, bandHeightFooterDocument("30", "100",
		bandHeightRect("e7", "0", "10"),
		bandHeightRect("e8", "50", "30"),
	))
	failure := bandHeightRefusal(t, tpl, `{"kind":"setBandHeight","version":1,"band":"pageFooter","height":40}`)
	if failure.ElementID != "e8" {
		t.Errorf("strand refusal ElementID = %q, want e8 — the predicate must evaluate EVERY element in the footer", failure.ElementID)
	}
}

// ---------------------------------------------------------------------------
// THE GUARDS THE REVIEW ADDED.

// TestSetBandHeightSeesEveryBandItMaySet is what makes setBandHeight's
// "band missing from the projection" refusal unreachable, stated rather than
// assumed. If Canvas ever stops projecting one of the two settable bands, the
// strand check would silently examine a zero CanvasBand whose Name is in no
// list — every element would "fit" every height — so the handler refuses
// instead, and this test is the reason that branch cannot be reached today.
func TestSetBandHeightSeesEveryBandItMaySet(t *testing.T) {
	projection, err := Canvas(bandHeightTemplate(t, bandHeightDocument("20", "30")))
	if err != nil {
		t.Fatal(err)
	}
	projected := map[string]bool{}
	for _, band := range projection.Bands {
		projected[band.Name] = true
	}
	for _, name := range bandsCappingVertically {
		if !projected[name] {
			t.Fatalf("Canvas does not project %s; setBandHeight's strand check would examine a zero band and pass every element", name)
		}
	}
}

// TestSetBandHeightIgnoresHorizontalOverflow is patch 11. containComponent also
// enforces the horizontal bound, and a HEIGHT cannot change whether a component
// hangs off the side of its band — so an element wider than the printable
// column must not make every band-height change refused under a sentence about
// a vertical reach it never measured.
func TestSetBandHeightIgnoresHorizontalOverflow(t *testing.T) {
	// The printable column is 595.276 - 36 - 36 = 523.276pt wide. This rect is
	// 600pt wide and 10pt tall at y=0: it overflows HORIZONTALLY and fits any
	// height at all vertically.
	wide := `{"height":10,"id":"e1","type":"rect","width":600,"x":0,"y":0}`
	tpl := bandHeightTemplate(t, bandHeightDocument("100", "30", wide))
	// Precondition, measured rather than assumed: the element really is outside
	// its band horizontally, so this test is about the axis it claims to be.
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatalf("fixture precondition: the document must project: %v", err)
	}
	if containComponent(projection.Bands[0], 0, 0, 600_000, 10_000) == nil {
		t.Fatalf("fixture precondition: a 600pt-wide rect must overflow a %dmp band", projection.Bands[0].Width)
	}
	if _, err := ApplyComponentCommand(tpl, []byte(`{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80}`)); err != nil {
		t.Fatalf("a horizontally overflowing element refused a band-height change it cannot be affected by: %v", err)
	}
	// And the VERTICAL cap is not weakened by scoping to it: the same element,
	// at the same width, still strands when the height goes below its foot.
	failure := bandHeightRefusal(t, bandHeightTemplate(t, bandHeightDocument("100", "30", `{"height":30,"id":"e1","type":"rect","width":600,"x":0,"y":50}`)), `{"kind":"setBandHeight","version":1,"band":"pageHeader","height":79}`)
	if failure.ElementID != "e1" {
		t.Errorf("scoping to the vertical axis lost the strand: ElementID = %q, want e1", failure.ElementID)
	}
}

// TestSetBandHeightRefusalDataPathSurvivesTheHostsCut is patch 10. The comment
// on bandHeightPath claims it is "bounded the way fontChainPath is bounded";
// fontChainPath has TestFontChainRefusalDataPathSurvivesTheHostsCut and this is
// the band's equivalent. `band` arrives as a free string on the wire, the host
// cuts DataPath at maxComponentDataPathBytes BY BYTES, and a path split through
// the middle of a rune locates nothing.
func TestSetBandHeightRefusalDataPathSurvivesTheHostsCut(t *testing.T) {
	for _, probe := range []struct {
		name string
		band string
	}{
		// Single-byte runes, comfortably past the 256-byte cut: the plain
		// over-long case, where the trim must simply happen.
		{"a long ASCII band name", strings.Repeat("a", maxComponentDataPathBytes+64)},
		// Two-byte runes, so a cut taken by BYTES lands inside a character
		// unless it is walked back to a rune start.
		{"a long multi-byte band name", strings.Repeat("é", maxComponentDataPathBytes)},
		// Four-byte runes, the widest UTF-8 encodes: "bands." is 6 bytes, so
		// 256-6 = 250 is not a multiple of 4 and the cut MUST be walked back.
		{"a long four-byte-rune band name", strings.Repeat("𝔅", maxComponentDataPathBytes)},
	} {
		t.Run(probe.name, func(t *testing.T) {
			failure := bandHeightRefusal(t, bandHeightTemplate(t, bandHeightDocument("20", "30")), `{"kind":"setBandHeight","version":1,"band":"`+probe.band+`","height":20}`)
			if len(failure.DataPath) > maxComponentDataPathBytes {
				t.Fatalf("DataPath is %d bytes; the host cuts at %d and would take this one apart itself", len(failure.DataPath), maxComponentDataPathBytes)
			}
			if !utf8.ValidString(failure.DataPath) {
				t.Fatalf("DataPath is not valid UTF-8: %q", failure.DataPath)
			}
			if !strings.HasPrefix(failure.DataPath, "bands.") {
				t.Fatalf("DataPath = %q, want the bands.<name> shape even trimmed", failure.DataPath)
			}
		})
	}
	// NON-VACUITY: a band name that FITS keeps its whole path, so the trim above
	// is a measurement of the bound rather than of a guard that always fires.
	if failure := bandHeightRefusal(t, bandHeightTemplate(t, bandHeightDocument("20", "30")), `{"kind":"setBandHeight","version":1,"band":"footer","height":20}`); failure.DataPath != "bands.footer.height" {
		t.Fatalf("an ordinary refusal lost its path: %q", failure.DataPath)
	}
}

// ---------------------------------------------------------------------------
// RED-PROOFS (D-000.14). Each was run by MUTATING THE SUBJECT — the production
// code in component_commands.go and page_setup.go — never the expectation, and
// each mutated line is recorded here so a later reader can repeat it exactly.
//
//  1. DELETE THE STRAND CHECK. Removing the `for _, element := range
//     band.Elements` loop from setBandHeight (replaced by `_ = candidate`):
//     FIVE tests FAIL — TestSetBandHeightRefusesToStrandAComponent,
//     …ThatIsNotTheFirst, …InTheFooter, …ThatIsNotTheFirstInTheFooter and
//     TestSetBandHeightIgnoresHorizontalOverflow (whose second half is a
//     strand). No other test moves.
//
//  2. EXAMINE ONLY THE FIRST ELEMENT. Replacing the loop's range with
//     `for _, element := range band.Elements[:min(1, len(band.Elements))] {`
//     (the `min` is only so an EMPTY band does not panic instead of passing,
//     which would redden the wrong tests):
//     TestSetBandHeightRefusesToStrandAComponentThatIsNotTheFirst and its
//     …InTheFooter twin FAIL, and the two single-element strand tests still
//     pass — which is the whole reason the not-the-first case is a separate
//     test. This is the vacuity the lead asked to be checked, and it is
//     checked by a fixture whose first element FITS.
//
//  3. DELETE THE COMMAND'S CONTENT-WINDOW CALL. Replacing the
//     `if !bandsLeaveContentWindow(header, footer, innerH) {` block in
//     setBandHeight with `_, _, _ = header, footer, innerH`:
//     TestSetBandHeightRefusesAHeightThatLeavesNoContentWindow,
//     TestSetBandHeightNamesTheBoundThePredicateAcceptsUpTo,
//     TestSetBandHeightRefusesBandsThatExactlyFillTheColumn and
//     TestSetBandHeightRefusesAFooterHeightThatLeavesNoContentWindow FAIL, and
//     TestCanvasRefusesBandsThatLeaveNoContentWindow still passes — the two
//     call sites are independent.
//
//     WHAT THEY FAIL ON IS THE POINT, and it is why this red-proof was worth
//     running: the command still REFUSES, because the Canvas(t) at the end of
//     the handler asks the same predicate and rolls the write back. What is
//     lost is LOCATEDNESS — the refusal arrives as
//     `*errors.errorString (folio: page setup leaves no positive content
//     region)` instead of a *ComponentCommandError, which is exactly the shape
//     pageSetupDiagnostic discards for its fixed sentence. The command's own
//     check is not a second guard against a bad document; it is the only thing
//     that tells the author which number they must change.
//
//  4. CHANGE THE SHARED PREDICATE. Rewriting page_setup.go's
//     `return innerH - other - 1` in bandContentWindowCeiling as
//     `return innerH - other` (the predicate is written in terms of that
//     ceiling, so this is the one place the strictness lives):
//     TestCanvasRefusesBandsThatLeaveNoContentWindow AND three command tests —
//     TestSetBandHeightNamesTheBoundThePredicateAcceptsUpTo,
//     TestSetBandHeightRefusesBandsThatExactlyFillTheColumn and
//     TestSetBandHeightRefusesAFooterHeightThatLeavesNoContentWindow — go red
//     TOGETHER, from ONE edit in ONE place. That is the assertion Q3's
//     condition asks for and no assertion in this file can make on its own: a
//     second implementation would have let one of the two sides stay green.
//
//  5. THE ROTATION (2026-09-05 review, patch 1). Deleting the band-pointer
//     swap in setBandHeight —
//     `band, otherName := &t.doc.Bands.PageHeader, bandPageFooter` kept and the
//     `if name == bandPageFooter { band, otherName = &t.doc.Bands.PageFooter,
//     bandPageHeader }` block removed — so a pageFooter command writes
//     PageHeader.
//
//     BEFORE THIS REVIEW THAT LEFT THE ENTIRE GO SUITE GREEN except the
//     mandated P6g red: every accepting, strand, boundary and content-window
//     test used pageHeader, and the file's one pageFooter command was the
//     negative-height row, which returns above the swap. It now FAILS FIVE
//     tests, all of them the mirrored ones added for it:
//     TestSetBandHeightWritesTheFooterHeightTheAuthorSet,
//     TestSetBandHeightAcceptsTheLowestOccupiedEdgeOfTheFooter,
//     TestSetBandHeightRefusesToStrandAComponentInTheFooter,
//     TestSetBandHeightRefusesAFooterHeightThatLeavesNoContentWindow and
//     TestSetBandHeightRefusesToStrandAComponentThatIsNotTheFirstInTheFooter.
//
//  6. RE-DERIVE THE BOUND IN THE MESSAGE (patch 2). Replacing
//     `template.FormatPoints(bandContentWindowCeiling(other, innerH))` in the
//     content-window refusal with `template.FormatPoints(innerH-other)` — the
//     second spelling the predicate exists to prevent, and one the CHECK does
//     not notice, because it still refuses exactly the right heights:
//     TestSetBandHeightNamesTheBoundThePredicateAcceptsUpTo and
//     TestSetBandHeightRefusesAFooterHeightThatLeavesNoContentWindow FAIL, on
//     the height the message itself named as legal being refused.
//
//  7. UN-SCOPE THE STRAND CHECK FROM THE VERTICAL AXIS (patch 11). Putting
//     `containComponent(candidate, element.X, element.Y, width, height)` back
//     in place of `containComponent(candidate, 0, element.Y, 0, height)`:
//     TestSetBandHeightIgnoresHorizontalOverflow FAILS — an element that hangs
//     off the SIDE of its band makes every band-height change refused, under a
//     sentence naming a vertical reach nothing measured. Nothing else moves,
//     which is the other half of the claim: the vertical cap is unweakened.
