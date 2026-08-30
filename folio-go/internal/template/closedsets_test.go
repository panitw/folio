package template

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file is AC11's mandated second half (D-1.8.1, D-1.4.12; Finding
// 5, Story 1.8 review): "mediaType must not be added to
// internal/template/closedsets.go, AND A TEST ASSERTS IT IS ABSENT from
// that file's set inventory." Before this file, the first half held —
// grepping closedsets.go for "mediaType" found nothing — but nothing in
// the tree enforced it or would even notice a future agent adding it:
// the property held because nobody had added it, the exact "unexercised,
// not enforced" pattern AC17b names for AD-24. Under D-1.4.12, adding
// mediaType to this file's closed-set inventory is a permanent MAJOR
// version bump on the format — the one-way door this test exists to
// guard.
//
// Enumerated by AST, same technique as drift_test.go's extractGoKeys:
// no checked-in list of set names to fall out of date.

// closedSetsSource is the file whose package-level `closed*` map
// declarations are this package's closed-set inventory.
const closedSetsSource = "closedsets.go"

// scanClosedSets parses path and returns the name of every package-level
// var whose declared name starts with "closed" (this package's
// closed-set naming convention — closedLocales, closedElementTypes, and
// so on), together with every string literal key found inside each such
// var's map-literal initializer (so a mediaType set added under an
// EXISTING, more generically-named var would still be caught by its
// contents, not just by a new var's name).
func scanClosedSets(path string) (setNames []string, allKeys []string, err error) {
	fset := token.NewFileSet()
	file, perr := parser.ParseFile(fset, path, nil, 0)
	if perr != nil {
		return nil, nil, perr
	}
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 {
				continue
			}
			name := vs.Names[0].Name
			if !strings.HasPrefix(name, "closed") {
				continue
			}
			setNames = append(setNames, name)
			for _, val := range vs.Values {
				cl, ok := val.(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, elt := range cl.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					if lit, ok := kv.Key.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						unquoted := strings.Trim(lit.Value, `"`)
						allKeys = append(allKeys, unquoted)
					}
				}
			}
		}
	}
	return setNames, allKeys, nil
}

// TestClosedSetsNeverIncludeMediaType is AC11's mechanical half.
func TestClosedSetsNeverIncludeMediaType(t *testing.T) {
	setNames, allKeys, err := scanClosedSets(closedSetsSource)
	if err != nil {
		t.Fatalf("scan %s: %v", closedSetsSource, err)
	}
	if len(setNames) == 0 {
		t.Fatal("vacuity guard (D-000.9): found zero closed* set declarations in closedsets.go")
	}
	for _, name := range setNames {
		if strings.Contains(strings.ToLower(name), "mediatype") {
			t.Fatalf("AC11 violation: closedsets.go declares %q — mediaType must never join this "+
				"file's closed-set inventory (D-1.4.12: doing so is a permanent MAJOR version bump)", name)
		}
	}
	for _, key := range allKeys {
		if strings.HasPrefix(key, "image/") || strings.HasPrefix(key, "application/") {
			t.Fatalf("AC11 violation: closedsets.go contains a media-type-shaped key %q inside one of "+
				"its closed sets (D-1.4.12)", key)
		}
	}
}

// TestClosedSetsMediaTypeRedProof is this test's own red-proof: add a
// closedMediaTypes map (the suggested resolution's own example) to a
// scratch copy of the real closedsets.go and confirm the scanner names
// it, per D-000.13.
func TestClosedSetsMediaTypeRedProof(t *testing.T) {
	real, err := os.ReadFile(closedSetsSource)
	if err != nil {
		t.Fatalf("read %s: %v", closedSetsSource, err)
	}
	mutated := string(real) + "\n" +
		"var closedMediaTypes = map[string]bool{\n\t\"image/png\": true, \"image/jpeg\": true,\n}\n"

	dir := t.TempDir()
	scratch := filepath.Join(dir, "closedsets.go")
	if err := os.WriteFile(scratch, []byte(mutated), 0o644); err != nil {
		t.Fatalf("write scratch closedsets.go: %v", err)
	}

	setNames, allKeys, err := scanClosedSets(scratch)
	if err != nil {
		t.Fatalf("scan mutated closedsets.go: %v", err)
	}
	if len(setNames) == 0 {
		t.Fatal("vacuity guard: mutated scratch file produced zero set names")
	}

	named := false
	for _, name := range setNames {
		if strings.Contains(strings.ToLower(name), "mediatype") {
			named = true
		}
	}
	if !named {
		t.Fatalf("RP: expected closedMediaTypes to be named among %v", setNames)
	}
	foundKey := false
	for _, key := range allKeys {
		if key == "image/png" {
			foundKey = true
		}
	}
	if !foundKey {
		t.Fatalf("RP: expected the media-type-shaped key \"image/png\" to be extracted, got %v", allKeys)
	}
}

// alignDocWithStyle is one text element carrying `style` verbatim, so a
// case can put any align spelling — legal or not — at the STYLE
// attachment point.
func alignDocWithStyle(style string) []byte {
	return []byte(`{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 40, "value": "v", "style": {` + style + `}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "2.0"
}
`)
}

// alignDocWithTable is one table carrying a headerStyle, a single
// column and — since Story 7.8 — its own `style` block, so a case can
// put an align spelling at any of the THREE table attachment points and
// see which one rejected it. The table's own `style` was absent from
// this helper until Story 7.8, which is why nothing in the repository
// pinned a table's `style.align: "justify"` while it was silently
// accepted (DW-29).
func alignDocWithTable(tableStyle, headerStyle, columnExtra string) []byte {
	return []byte(`{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "rows[]", "as": "row", "headerHeight": 20,
          "columns": [{"id": "e2", "label": "L", "width": 100, "bind": "{{row.v}}"` + columnExtra + `}],
          "style": {` + tableStyle + `},
          "headerStyle": {` + headerStyle + `}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "2.0"
}
`)
}

// TestAlignSetsAreThreeSetsPinnedAgainstTheirMaps is Story 7.3's
// structural guard, WIDENED at Story 7.8 from two sets to three. It
// exists because NOTHING enumerated the shared map's members before the
// first split — TestClosedSetsNeverIncludeMediaType scans this file's
// inventory but only for media-type shapes, so a `justify` added to the
// shared map would have passed every test in the repository while
// silently legalising justified table cells.
//
// It pins each ordered slice against its map in BOTH directions and as
// an exact literal SEQUENCE (LocaleTags' own AC4a shape): a set that
// gains a member without gaining it in the slice ships a rejection
// message that lies about what is legal, and a slice reordered by
// accident reorders the message every load error prints.
//
// THE ASSERTION THIS TEST USED TO CARRY WAS WRONG, and its name said
// so. `len(StyleAlignTokens) == len(ColumnAlignTokens)+1` states "there
// are exactly two sets and they differ by `justify`", which was true of
// the KEY-PATH partition Story 7.3 shipped and false of the CONSUMER
// partition that actually holds. The replacement below states the
// consumer invariant instead: `justify` is admitted by exactly one of
// the three sets, and the two TABLE sets are the same three values
// because one alignFallback consumes both.
func TestAlignSetsAreThreeSetsPinnedAgainstTheirMaps(t *testing.T) {
	for _, c := range []struct {
		name   string
		tokens []string
		set    map[string]bool
		want   []string
	}{
		{"style", StyleAlignTokens, closedStyleAligns, []string{"left", "center", "right", "justify"}},
		{"table-style", TableStyleAlignTokens, closedTableStyleAligns, []string{"left", "center", "right"}},
		{"column", ColumnAlignTokens, closedColumnAligns, []string{"left", "center", "right"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if len(c.tokens) != len(c.want) {
				t.Fatalf("%s tokens = %v, want exactly %v", c.name, c.tokens, c.want)
			}
			for i, w := range c.want {
				if c.tokens[i] != w {
					t.Errorf("%s token %d is %q, want %q — the order IS the message's order", c.name, i, c.tokens[i], w)
				}
			}
			// slice -> map
			for _, tok := range c.tokens {
				if !c.set[tok] {
					t.Errorf("%s: %q is in the ordered slice but the map does not admit it — the message would name a value the loader refuses", c.name, tok)
				}
			}
			// map -> slice
			if len(c.set) != len(c.tokens) {
				t.Errorf("%s: the map has %d members and the slice %d — the loader would admit a value the message never names", c.name, len(c.set), len(c.tokens))
			}
		})
	}

	// THE PARTITION ITSELF, stated as the facts D-7.3.1 and Story 7.8
	// turn on. `justify` belongs to exactly ONE of the three sets: the
	// one whose consumer is the paragraph justifier.
	if !closedStyleAligns[AlignJustify] {
		t.Error("a non-table element's style.align must admit justify (FR47)")
	}
	justifying := 0
	for _, set := range []map[string]bool{closedStyleAligns, closedTableStyleAligns, closedColumnAligns} {
		if set[AlignJustify] {
			justifying++
		}
	}
	if justifying != 1 {
		t.Fatalf("%d of the three align sets admit justify, want exactly 1 — the paragraph justifier's set alone. A table's style.align, its headerStyle.align and its columns[].align are read into ONE alignFallback and drawn by ONE set of switches, so any of them admitting justify re-opens DW-29", justifying)
	}
	if closedTableStyleAligns[AlignJustify] {
		t.Fatal("a table's style.align/headerStyle.align must NEVER admit justify — it cascades into the same cell switches columns[].align feeds, and the document would pay a MAJOR version bump for a render identical to `left` (DW-29, D-7.3.1's partition corrected at Story 7.8)")
	}
	if closedColumnAligns[AlignJustify] {
		t.Fatal("columns[].align must NEVER admit justify — justified table cells are a separate scope decision, not a side effect of a map edit (D-7.3.1)")
	}

	// THE TWO TABLE SETS ARE THE SAME SET OF VALUES, and that is the
	// consumer partition made visible: one alignFallback, one
	// vocabulary. They are separate DECLARATIONS (separately documented
	// closed sets, either free to move) but a divergence between them
	// would mean the same switch could receive a value it cannot draw.
	if len(TableStyleAlignTokens) != len(ColumnAlignTokens) {
		t.Errorf("the table-style set has %d members and the column set %d — both feed the same alignFallback, so a value legal at one and not the other is a value the cell switches cannot account for", len(TableStyleAlignTokens), len(ColumnAlignTokens))
	}
	for i, tok := range TableStyleAlignTokens {
		if i < len(ColumnAlignTokens) && ColumnAlignTokens[i] != tok {
			t.Errorf("table-style token %d is %q but column token %d is %q", i, tok, i, ColumnAlignTokens[i])
		}
	}

	// And no set is another by subtraction at a call site: every table
	// value is also a legal non-table style value, so the three sets
	// have not diverged in the wrong direction.
	for _, tok := range TableStyleAlignTokens {
		if !closedStyleAligns[tok] {
			t.Errorf("%q is a legal table align but not a legal style align — the sets have diverged in the wrong direction", tok)
		}
	}
	for _, tok := range ColumnAlignTokens {
		if !closedStyleAligns[tok] {
			t.Errorf("%q is a legal column align but not a legal style align — the sets have diverged in the wrong direction", tok)
		}
	}

	// The exported predicates are the single source the property-command
	// path validates through; each must agree with the map it wraps, and
	// they must disagree with each other on exactly `justify`.
	for _, tok := range StyleAlignTokens {
		if !IsStyleAlign(tok) {
			t.Errorf("IsStyleAlign(%q) = false", tok)
		}
	}
	if IsStyleAlign("middle") || IsStyleAlign("") {
		t.Error("IsStyleAlign admits a value the closed set does not")
	}
	for _, tok := range TableStyleAlignTokens {
		if !IsTableStyleAlign(tok) {
			t.Errorf("IsTableStyleAlign(%q) = false", tok)
		}
	}
	if IsTableStyleAlign(AlignJustify) {
		t.Error("IsTableStyleAlign admits justify — the property-command path would write onto a table exactly the value the loader refuses")
	}
	if IsTableStyleAlign("middle") || IsTableStyleAlign("") {
		t.Error("IsTableStyleAlign admits a value the closed set does not")
	}
}

// TestAlignClosedSetsRejectAtTheRightSiteWithTheirOwnMessage is the
// partition's BEHAVIOURAL red-proof: the same word, `justify`, must load
// at a TEXT element's style.align and be a located load error at all
// three of a table's attachment points — and each message must list
// exactly the members of the set that rejected it, not another set's.
//
// Legs (b) and (b2) are Story 7.8's inversion. (b) asserted that
// `headerStyle.align: "justify"` MUST LOAD, which was correct against
// the contract as Story 7.3 wrote it and is the shipped behaviour
// DW-29 records as the defect. It now asserts the refusal. (b2) is new:
// a table's OWN style.align was pinned by nothing at all, because the
// helper emitted no table `style` block.
func TestAlignClosedSetsRejectAtTheRightSiteWithTheirOwnMessage(t *testing.T) {
	// (a) justify LOADS at a TEXT element's style.align.
	if _, err := ParseDocument(alignDocWithStyle(`"align": "justify"`)); err != nil {
		t.Errorf("style.align: \"justify\" on a text element must load: %v", err)
	}
	// (b) justify is REFUSED at a table's headerStyle.align, located at
	//     the element and the field.
	_, berr := ParseDocument(alignDocWithTable("", `"align": "justify"`, ""))
	if berr == nil {
		t.Fatal("headerStyle.align: \"justify\" must be a load error — a header cell is a table cell, and it takes the same alignFallback columns[].align feeds (DW-29)")
	}
	assertLocatedTableAlignRefusal(t, berr, "headerStyle.align")

	// (b2) justify is REFUSED at a table's OWN style.align. This is the
	//      door D-7.3.1's key-path partition left open: the guard was
	//      written for `columns[]` versus `style`, but a table's own
	//      style cascades into every cell that declares no column align.
	_, b2err := ParseDocument(alignDocWithTable(`"align": "justify"`, "", ""))
	if b2err == nil {
		t.Fatal("a TABLE's own style.align: \"justify\" must be a load error — it cascades into every header, body and footer cell through alignFallback (DW-29)")
	}
	assertLocatedTableAlignRefusal(t, b2err, "style.align")

	// (c) justify is REFUSED at columns[].align, located at the column.
	_, err := ParseDocument(alignDocWithTable("", "", `, "align": "justify"`))
	if err == nil {
		t.Fatal("columns[].align: \"justify\" must be a load error — justified table cells are not in scope")
	}
	msg := err.Error()
	if !strings.Contains(msg, "e2") {
		t.Errorf("the column rejection must be LOCATED at the column that carries it, got: %v", err)
	}
	if !strings.Contains(msg, "align") {
		t.Errorf("the column rejection must name the field, got: %v", err)
	}
	if !strings.Contains(msg, "not one of the closed set left, center, right") {
		t.Errorf("the column rejection must list exactly the COLUMN set, got: %v", err)
	}
	if strings.Contains(msg, "justify,") || strings.Contains(msg, ", justify") {
		t.Errorf("the column rejection must not name justify as legal, got: %v", err)
	}

	// (d) An illegal value on a TEXT element's style is rejected with
	//     the STYLE set, which includes justify. This is the honesty
	//     half: before the split both sites printed one hand-written
	//     literal, so extending either set would have shipped a message
	//     that lied.
	_, serr := ParseDocument(alignDocWithStyle(`"align": "middle"`))
	if serr == nil {
		t.Fatal("style.align: \"middle\" must be a load error")
	}
	smsg := serr.Error()
	if !strings.Contains(smsg, "not one of the closed set left, center, right, justify") {
		t.Errorf("the style rejection must list exactly the STYLE set, justify included, got: %v", serr)
	}
	if !strings.Contains(smsg, "style.align") {
		t.Errorf("the style rejection must be located at style.align, got: %v", serr)
	}

	// (e) …and at a table's headerStyle it names headerStyle, not its
	//     sibling — and it does NOT name justify, because that element
	//     can never carry it. A type-guard bolted above the two shipped
	//     sets would pass every other leg here and fail this one.
	_, herr := ParseDocument(alignDocWithTable("", `"align": "middle"`, ""))
	if herr == nil {
		t.Fatal("headerStyle.align: \"middle\" must be a load error")
	}
	if !strings.Contains(herr.Error(), "headerStyle.align") {
		t.Errorf("the headerStyle rejection must be located at headerStyle.align, got: %v", herr)
	}
	if !strings.Contains(herr.Error(), closedSetMessage(TableStyleAlignTokens)) {
		t.Errorf("a table's headerStyle rejection must list the TABLE set, got: %v", herr)
	}
	if strings.Contains(herr.Error(), AlignJustify) {
		t.Errorf("a table's rejection message must never name justify as legal — no table element can carry it, got: %v", herr)
	}

	// (e2) …and at a table's OWN style.align, the same. This is the
	//      I/O matrix's "table style middle" row: the message must
	//      list left, center, right and must NOT name justify.
	_, terr := ParseDocument(alignDocWithTable(`"align": "middle"`, "", ""))
	if terr == nil {
		t.Fatal("a table's style.align: \"middle\" must be a load error")
	}
	if !strings.Contains(terr.Error(), "style.align") || !strings.Contains(terr.Error(), "e1") {
		t.Errorf("the table-style rejection must be located at the element and style.align, got: %v", terr)
	}
	if !strings.Contains(terr.Error(), closedSetMessage(TableStyleAlignTokens)) {
		t.Errorf("a table's style rejection must list the TABLE set, got: %v", terr)
	}
	if strings.Contains(terr.Error(), AlignJustify) {
		t.Errorf("a table's rejection message must never name justify as legal, got: %v", terr)
	}

	// (f) THE MESSAGES ARE DERIVED, not restated: each one contains the
	//     exact rendering of its own ordered slice. A future edit to any
	//     set changes both the enforcement and the message, or this
	//     fails.
	if !strings.Contains(msg, closedSetMessage(ColumnAlignTokens)) {
		t.Errorf("the column message is not derived from ColumnAlignTokens: %v", err)
	}
	if !strings.Contains(smsg, closedSetMessage(StyleAlignTokens)) {
		t.Errorf("the style message is not derived from StyleAlignTokens: %v", serr)
	}
	if !strings.Contains(b2err.Error(), closedSetMessage(TableStyleAlignTokens)) {
		t.Errorf("the table-style message is not derived from TableStyleAlignTokens: %v", b2err)
	}
}

// assertLocatedTableAlignRefusal is AC1's "located" clause, checked the
// same way at both of a table's style attachment points: the element id,
// the field, and a legal-values list that is the TABLE set and names no
// value that element cannot carry.
func assertLocatedTableAlignRefusal(t *testing.T, err error, field string) {
	t.Helper()
	le, ok := err.(*LoadError)
	if !ok {
		t.Fatalf("the refusal must be a *LoadError so the field and element travel as DATA, got %T: %v", err, err)
	}
	if le.ElementID != "e1" {
		t.Errorf("LoadError.ElementID = %q, want %q — the refusal must name the element", le.ElementID, "e1")
	}
	if le.Field != field {
		t.Errorf("LoadError.Field = %q, want %q", le.Field, field)
	}
	if le.Value != AlignJustify {
		t.Errorf("LoadError.Value = %q, want %q", le.Value, AlignJustify)
	}
	msg := le.Error()
	for _, want := range []string{"e1", field, closedSetMessage(TableStyleAlignTokens)} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal message must contain %q, got: %s", want, msg)
		}
	}
	if strings.Contains(msg, AlignJustify+",") || strings.Contains(msg, ", "+AlignJustify) {
		t.Errorf("the refusal must not name justify among the legal values for a table, got: %s", msg)
	}
}
