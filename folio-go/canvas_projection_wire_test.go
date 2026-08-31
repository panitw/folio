package folio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// This file guards the ONE SEAM neither side's own tests can see: the JSON
// key set CanvasProjection actually puts on the wire, against the exact key
// set the browser will accept.
//
// WHY IT IS NEEDED, and it is not hypothetical — Story 7.9 renamed a field
// across this seam. Go's tests read the STRUCT FIELD
// (`projection.ContentWindowCountIsExact`), so they never see the json tag;
// the designer's tests hand-author their canvas fixtures as object literals
// in their own files, so they never see Go. Both sides can therefore agree
// with themselves while disagreeing with each other, and a one-sided rename
// is green twice over.
//
// WHAT THE FAILURE LOOKS LIKE IF IT SHIPS, which is why a merely mechanical
// check is worth this much prose. engine-protocol.ts's `hasOnly` is an
// EXACT-KEY check: a key Go sends that the list does not name makes isCanvas
// return false, parseInbound discards the WHOLE SNAPSHOT, and the canvas
// blanks with no attributable error — no diagnostic, no console line, nothing
// naming the field. That is the same class of silent-blank failure
// canvasWindowOrigins refuses a bad origin sequence to avoid, one layer up.
//
// The shape is internal/template/drift_test.go's: neither side is trusted as
// the definition. The Go side is EXTRACTED from what encoding/json actually
// emits (not read off the struct tags, which is the same file a rename would
// have edited), the TypeScript side is EXTRACTED from the guard's own key
// list in engine-protocol.ts, and the recorded literal below is a third,
// independent statement that both must equal — so a rename applied to both
// halves in one edit still reddens here until the record is updated
// deliberately.

// canvasProjectionWireKeys is the recorded key set, sorted. Every entry is a
// name the designer's isCanvas guard lists and Go's CanvasProjection emits.
// Changing this list is a PROTOCOL CHANGE: both the Go json tag and
// engine-protocol.ts's hasOnly list have to move with it, in the same commit.
var canvasProjectionWireKeys = []string{
	"bands",
	"commandHeight",
	"commandWidth",
	"components",
	"contentWindowCount",
	"contentWindowCountIsExact",
	"contentWindowHeight",
	"contentWindowOrigins",
	"defaultFontSize",
	"fontChains",
	"fontFamilies",
	"gridIncrement",
	"height",
	"marginBottom",
	"marginLeft",
	"marginRight",
	"marginTop",
	"orientation",
	"preset",
	"width",
}

// canvasFontChainWireKeys is the recorded key set of the NESTED object, sorted.
// fontChains is the first nested object this projection carries, and the
// browser checks it with hasExactKeys — a check that is BOTH ways, so a field
// added to CanvasFontChain in Go and not to engine-protocol.ts makes isCanvas
// return false, and the symptom is the blank canvas this file's header
// describes, one level down where the top-level key list cannot see it.
var canvasFontChainWireKeys = []string{"entries", "name"}

// canvasFontChainEntryWireKeys is the recorded key set of the ENTRY object —
// one level below the chain — and it closes a gap that was MEASURED rather
// than supposed.
//
// The chain-level record above pins key NAMES only. It says nothing about an
// entry's value TYPE or nesting depth, so Story 8.3's change — CanvasFontChain
// .Entries going from []string to a slice of four-field structs — would have
// left that record green while the browser's own entry guard rejected every
// snapshot: isCanvas false, parseInbound undefined, engine-client terminates
// the worker, and the canvas is permanently blank with no element id and
// nothing to attribute it to. A key added to CanvasFontChainEntry in Go and
// not to engine-protocol.ts's entry guard does exactly the same thing, two
// levels down where neither existing record can see it.
//
// The browser's entry guard is hasExactKeys, so it rejects in BOTH
// directions: a key Go stops sending fails it as surely as a key Go starts
// sending.
var canvasFontChainEntryWireKeys = []string{"assetKey", "face", "family", "style"}

// canvasTextFragmentWireKeys is the recorded key set of the PAINT FRAGMENT —
// three levels below the projection's top level, inside components ->
// textPaint -> lines -> fragments — and the level at which Story 8.4a's
// attribution travels.
//
// IT IS THE ACCEPTED SET, NOT THE EMITTED ONE, and that is the difference
// from the three records above. `assetKey` is OPTIONAL by design: it is
// present exactly when the engine resolved that fragment to a face the
// document CARRIES, and absent — omitempty — for every fragment drawn with a
// shipped face, which is the wire's own precise statement of "this fragment is
// a shipped face". The browser's fragment guard is `hasOnly`, a SUBSET check,
// so the accepted list is what this record pins on the TypeScript side, and
// the Go side is pinned in both directions: a carried fragment emits the whole
// set, a shipped one emits it minus exactly the optional key.
//
// WHY THIS LEVEL NEEDED A RECORD AT ALL, measured rather than supposed. The
// three records above descend into `fontChains` and stop; NOTHING in this file
// or anywhere else in Go descended into `components`, `textPaint`, `lines` or
// `fragments`. So a fragment field added in Go and not in engine-protocol.ts
// reddened NOTHING here — and on the browser it is the sharpest failure this
// protocol has: `hasOnly` rejects the unlisted key, isCanvasTextPaint fails,
// isCanvas fails, parseInbound returns undefined, and engine-client raises
// PROTOCOL_INVALID, which TERMINATES THE WORKER, rejects the ready promise and
// rejects every pending request. That is not a blank canvas — the session is
// dead until reload, with no edit, save, undo or preview possible.
var canvasTextFragmentWireKeys = []string{"assetKey", "text", "x"}

// marshalledCanvasKeys is the Go side, taken from the bytes rather than from
// the struct: whatever encoding/json puts on the wire for this value, sorted.
func marshalledCanvasKeys(t *testing.T, projection CanvasProjection) []string {
	t.Helper()
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatalf("marshal CanvasProjection: %v", err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatalf("unmarshal CanvasProjection: %v", err)
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// TestCanvasProjectionWireKeysAreTheRecordedSet is the Go half. It marshals a
// REAL projection — one built from a template, through the entry point that
// reaches the browser — and, separately, the zero value, because the two can
// only differ if a field acquires an `omitempty` that would silently drop a
// key from a document that happens not to set it.
func TestCanvasProjectionWireKeysAreTheRecordedSet(t *testing.T) {
	projected := marshalledCanvasKeys(t, projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON)))
	if !reflect.DeepEqual(projected, canvasProjectionWireKeys) {
		t.Errorf("CanvasProjection marshals the keys\n\t%v\nand the recorded protocol set is\n\t%v", projected, canvasProjectionWireKeys)
	}
	zero := marshalledCanvasKeys(t, CanvasProjection{})
	if !reflect.DeepEqual(zero, projected) {
		t.Errorf("a zero CanvasProjection marshals\n\t%v\nand a projected one\n\t%v — a key that appears only sometimes is a key the browser's exact-key guard will reject only sometimes", zero, projected)
	}
	// And the NESTED object, from the same bytes. A field added to
	// CanvasFontChain is invisible to the check above — the top-level key set
	// does not change — and the browser rejects it just as hard.
	chainKeys := marshalledObjectKeys(t, chainFromProjectionBytes(t, projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON))))
	if !reflect.DeepEqual(chainKeys, canvasFontChainWireKeys) {
		t.Errorf("CanvasFontChain marshals the keys\n\t%v\nand the recorded protocol set is\n\t%v", chainKeys, canvasFontChainWireKeys)
	}
	zeroChain := marshalledObjectKeys(t, mustMarshal(t, CanvasFontChain{}))
	if !reflect.DeepEqual(zeroChain, chainKeys) {
		t.Errorf("a zero CanvasFontChain marshals\n\t%v\nand a projected one\n\t%v — an omitempty here drops a key the browser's exact-key guard requires", zeroChain, chainKeys)
	}
	// And the ENTRY, one level below that. Same reason, same failure mode:
	// invisible to both records above, and a blank canvas either way.
	entryKeys := marshalledObjectKeys(t, entryFromProjectionBytes(t, projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON))))
	if !reflect.DeepEqual(entryKeys, canvasFontChainEntryWireKeys) {
		t.Errorf("CanvasFontChainEntry marshals the keys\n\t%v\nand the recorded protocol set is\n\t%v", entryKeys, canvasFontChainEntryWireKeys)
	}
	zeroEntry := marshalledObjectKeys(t, mustMarshal(t, CanvasFontChainEntry{}))
	if !reflect.DeepEqual(zeroEntry, entryKeys) {
		t.Errorf("a zero CanvasFontChainEntry marshals\n\t%v\nand a projected one\n\t%v — an entry key that appears only for SOME entries (a named face carries no family; an embedded one does) is one the browser's exact-key guard rejects only for some documents", zeroEntry, entryKeys)
	}
	// And the PAINT FRAGMENT, three levels down inside components, which none
	// of the three checks above can see. Both halves of the optional key are
	// asserted, from two documents that differ in exactly one thing: whether
	// the face the engine resolved is one the document carries.
	carriedKeys := marshalledObjectKeys(t, fragmentFromProjectionBytes(t, carriedFaceProjection(t, embeddedFontTemplateJSON())))
	if !reflect.DeepEqual(carriedKeys, canvasTextFragmentWireKeys) {
		t.Errorf("a CARRIED-face CanvasTextFragment marshals the keys\n\t%v\nand the recorded protocol set is\n\t%v — a fragment field added on one side only terminates the designer's worker, which is a harder failure than the blank canvas the records above describe", carriedKeys, canvasTextFragmentWireKeys)
	}
	shippedKeys := marshalledObjectKeys(t, fragmentFromProjectionBytes(t, projectWithPaint(t, parseWindowCountTemplate(t, canvasWindowCountControlTemplateJSON))))
	var withoutOptional []string
	for _, key := range canvasTextFragmentWireKeys {
		if key != "assetKey" {
			withoutOptional = append(withoutOptional, key)
		}
	}
	if !reflect.DeepEqual(shippedKeys, withoutOptional) {
		t.Errorf("a SHIPPED-face CanvasTextFragment marshals the keys\n\t%v\nand the recorded set minus its one optional key is\n\t%v — the optional key must be omitted for a shipped face (its absence IS the statement) and every other key must always be present", shippedKeys, withoutOptional)
	}
}

// fragmentFromProjectionBytes digs the first painted fragment out of the
// marshalled projection, by the same rule the two helpers above use: from the
// bytes, never from the struct. The fixture must actually paint something, or
// the check asserts nothing.
func fragmentFromProjectionBytes(t *testing.T, projection CanvasProjection) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(mustMarshal(t, projection), &object); err != nil {
		t.Fatalf("unmarshal CanvasProjection: %v", err)
	}
	var components []json.RawMessage
	if err := json.Unmarshal(object["components"], &components); err != nil {
		t.Fatalf("unmarshal components: %v", err)
	}
	for _, raw := range components {
		var component map[string]json.RawMessage
		if err := json.Unmarshal(raw, &component); err != nil {
			t.Fatalf("unmarshal a component: %v", err)
		}
		paint, ok := component["textPaint"]
		if !ok {
			continue
		}
		var textPaint map[string]json.RawMessage
		if err := json.Unmarshal(paint, &textPaint); err != nil {
			t.Fatalf("unmarshal a textPaint: %v", err)
		}
		var lines []json.RawMessage
		if err := json.Unmarshal(textPaint["lines"], &lines); err != nil {
			t.Fatalf("unmarshal a textPaint's lines: %v", err)
		}
		for _, rawLine := range lines {
			var line map[string]json.RawMessage
			if err := json.Unmarshal(rawLine, &line); err != nil {
				t.Fatalf("unmarshal a paint line: %v", err)
			}
			var fragments []json.RawMessage
			if err := json.Unmarshal(line["fragments"], &fragments); err != nil {
				t.Fatalf("unmarshal a paint line's fragments: %v", err)
			}
			if len(fragments) > 0 {
				return fragments[0]
			}
		}
	}
	t.Fatal("fixture precondition: the wire-key fixture projected no paint fragment, so the fragment-level key check asserts nothing")
	return nil
}

// entryFromProjectionBytes digs the first projected chain's first ENTRY out of
// the marshalled projection, by the same rule chainFromProjectionBytes uses:
// from the bytes, never from the struct. The fixture must declare a chain with
// at least one entry, or the check asserts nothing.
func entryFromProjectionBytes(t *testing.T, projection CanvasProjection) []byte {
	t.Helper()
	var chain map[string]json.RawMessage
	if err := json.Unmarshal(chainFromProjectionBytes(t, projection), &chain); err != nil {
		t.Fatalf("unmarshal the first font chain: %v", err)
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(chain["entries"], &entries); err != nil {
		t.Fatalf("unmarshal the first font chain's entries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("fixture precondition: the wire-key fixture's first font chain must declare at least one entry, or the entry-level key check asserts nothing")
	}
	return entries[0]
}

func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	return encoded
}

// marshalledObjectKeys is marshalledCanvasKeys' half that reads an object's
// keys, over raw bytes, so a NESTED object can be read the same way the
// top-level one is: from what encoding/json actually emitted, never from the
// struct tags a rename would have edited in the same breath.
func marshalledObjectKeys(t *testing.T, object []byte) []string {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(object, &fields); err != nil {
		t.Fatalf("unmarshal object %s: %v", object, err)
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// chainFromProjectionBytes digs the first projected font chain out of the
// marshalled projection. The fixture must declare one: a projection with no
// chains would let this whole check pass while saying nothing.
func chainFromProjectionBytes(t *testing.T, projection CanvasProjection) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(mustMarshal(t, projection), &object); err != nil {
		t.Fatalf("unmarshal CanvasProjection: %v", err)
	}
	var chains []json.RawMessage
	if err := json.Unmarshal(object["fontChains"], &chains); err != nil {
		t.Fatalf("unmarshal fontChains: %v", err)
	}
	if len(chains) == 0 {
		t.Fatal("fixture precondition: the wire-key fixture must declare at least one font chain, or the nested key check asserts nothing")
	}
	return chains[0]
}

// canvasGuardKeyList extracts the key list engine-protocol.ts's isCanvas
// guard passes to hasOnly. It is deliberately the guard's OWN list and not a
// second copy of it kept here: a copy would be one more thing to forget to
// rename, which is the defect this file exists for.
var canvasGuardKeyList = regexp.MustCompile(`(?s)const isCanvas = .*?hasOnly\(value, \[(.*?)\]\)`)

// canvasGuardChainKeyList extracts the nested object's key list from the same
// guard, for the same reason and by the same rule: the guard's OWN list, never
// a copy of it kept here.
var canvasGuardChainKeyList = regexp.MustCompile(`hasExactKeys\(chain, \[(.*?)\]\)`)

// canvasGuardChainEntryKeyList extracts the ENTRY object's key list from the
// guard that checks it, for the same reason and by the same rule: the guard's
// OWN list, never a copy of it kept here. Anchored on isFontChainEntry's own
// name because `hasExactKeys(value, [...])` is the whole file's idiom and an
// unanchored match would read some other guard's list and compare it to this
// record, which is a green test asserting the wrong thing.
var canvasGuardChainEntryKeyList = regexp.MustCompile(`(?s)const isFontChainEntry = .*?hasExactKeys\(value, \[(.*?)\]\)`)

// canvasGuardFragmentKeyList extracts the PAINT FRAGMENT's key list from the
// guard that checks it, by the same rule: the guard's OWN list, never a copy
// of it kept here. Anchored on the `fragment` parameter name because
// `hasOnly(value, [...])` is the file's idiom and an unanchored match would
// read some other guard's list and compare it to this record — a green test
// asserting the wrong thing.
var canvasGuardFragmentKeyList = regexp.MustCompile(`hasOnly\(fragment, \[(.*?)\]\)`)

// TestCanvasProjectionWireKeysAreTheOnesTheDesignerAccepts is the TypeScript
// half, and the one that ties the two languages together. `hasOnly` rejects
// any key not on this list, so a Go-side rename that stops here — or a
// designer-side rename that Go never hears about — blanks the canvas.
func TestCanvasProjectionWireKeysAreTheOnesTheDesignerAccepts(t *testing.T) {
	path := filepath.Join(repoRootFromTest(t), "folio-designer", "src", "engine-protocol.ts")
	source, err := os.ReadFile(path)
	if err != nil {
		// Not a skip. The guard on the other side of this seam is part of the
		// protocol, and a missing one is a finding, not an excuse.
		t.Fatalf("read the designer's protocol guard: %v", err)
	}
	match := canvasGuardKeyList.FindSubmatch(source)
	if match == nil {
		t.Fatal("engine-protocol.ts no longer has an isCanvas guard whose hasOnly list this test can read; if the guard was restructured, re-derive this extraction rather than deleting the check")
	}
	keys := extractedKeyList(string(match[1]))
	if !reflect.DeepEqual(keys, canvasProjectionWireKeys) {
		t.Errorf("the designer's isCanvas guard accepts the keys\n\t%v\nand the recorded protocol set is\n\t%v — one side of this seam has been renamed and the other has not, and the symptom is a blank canvas with nothing to attribute it to", keys, canvasProjectionWireKeys)
	}
	// The NESTED object, whose guard is hasExactKeys rather than hasOnly and
	// so rejects in both directions: a key Go stops sending fails it as surely
	// as a key Go starts sending. Nothing else in either language sees this
	// pair — Go's own tests read the struct field, and the designer's read
	// hand-authored fixtures.
	chainMatch := canvasGuardChainKeyList.FindSubmatch(source)
	if chainMatch == nil {
		t.Fatal("engine-protocol.ts no longer checks the projected font chain's keys where this test can read it; if the guard was restructured, re-derive this extraction rather than deleting the check")
	}
	chainKeys := extractedKeyList(string(chainMatch[1]))
	if !reflect.DeepEqual(chainKeys, canvasFontChainWireKeys) {
		t.Errorf("the designer's font-chain guard accepts the keys\n\t%v\nand the recorded protocol set is\n\t%v — fontChains is the first NESTED object on this projection, and a field added to CanvasFontChain on one side only blanks the canvas exactly as a top-level one would", chainKeys, canvasFontChainWireKeys)
	}

	// The ENTRY object, one level below the chain — Story 8.3's addition, and
	// the level the two records above are blind to. They pin key NAMES only,
	// so an entry's shape could change from a string to an object with neither
	// of them noticing, while the browser rejected every snapshot.
	entryMatch := canvasGuardChainEntryKeyList.FindSubmatch(source)
	if entryMatch == nil {
		t.Fatal("engine-protocol.ts no longer checks the projected font chain ENTRY's keys where this test can read it; if the guard was restructured, re-derive this extraction rather than deleting the check")
	}
	entryKeys := extractedKeyList(string(entryMatch[1]))
	if !reflect.DeepEqual(entryKeys, canvasFontChainEntryWireKeys) {
		t.Errorf("the designer's font-chain ENTRY guard accepts the keys\n\t%v\nand the recorded protocol set is\n\t%v — a field added to CanvasFontChainEntry on one side only blanks the canvas exactly as a chain-level or a top-level one would, and neither of this file's other two records can see it", entryKeys, canvasFontChainEntryWireKeys)
	}

	// The PAINT FRAGMENT, three levels down inside components — Story 8.4a's
	// addition, and the level every record above is blind to. Its guard is
	// `hasOnly`, so what is pinned here is the ACCEPTED set: a key Go starts
	// sending that this list does not name terminates the designer's worker.
	fragmentMatch := canvasGuardFragmentKeyList.FindSubmatch(source)
	if fragmentMatch == nil {
		t.Fatal("engine-protocol.ts no longer checks the projected paint FRAGMENT's keys where this test can read it; if the guard was restructured, re-derive this extraction rather than deleting the check")
	}
	fragmentKeys := extractedKeyList(string(fragmentMatch[1]))
	if !reflect.DeepEqual(fragmentKeys, canvasTextFragmentWireKeys) {
		t.Errorf("the designer's paint-FRAGMENT guard accepts the keys\n\t%v\nand the recorded protocol set is\n\t%v — a fragment key added in Go and not here does not blank the canvas, it raises PROTOCOL_INVALID and terminates the worker, and the session is dead until reload", fragmentKeys, canvasTextFragmentWireKeys)
	}
}

// extractedKeyList turns a TypeScript array literal's inner text into the
// sorted key set it names.
func extractedKeyList(literal string) []string {
	var keys []string
	for _, field := range strings.Split(literal, ",") {
		key := strings.Trim(strings.TrimSpace(field), "'\"")
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
