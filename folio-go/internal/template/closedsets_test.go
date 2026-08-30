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

// alignDocWithTable is one table carrying a headerStyle and a single
// column, so a case can put an align spelling at EITHER of the two
// remaining attachment points — headerStyle (a STYLE set member) and
// columns[] (a COLUMN set member) — and see which one rejected it.
func alignDocWithTable(headerStyle, columnExtra string) []byte {
	return []byte(`{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "rows[]", "as": "row", "headerHeight": 20,
          "columns": [{"id": "e2", "label": "L", "width": 100, "bind": "{{row.v}}"` + columnExtra + `}],
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

// TestAlignSetsAreTwoSetsPinnedAgainstTheirMaps is Story 7.3's structural
// guard, and it exists because NOTHING enumerated closedAligns' members
// before the split — TestClosedSetsNeverIncludeMediaType scans this
// file's inventory but only for media-type shapes, so a `justify` added
// to the shared map would have passed every test in the repository while
// silently legalising justified table cells.
//
// It pins each ordered slice against its map in BOTH directions and as
// an exact literal SEQUENCE (LocaleTags' own AC4a shape): a set that
// gains a member without gaining it in the slice ships a rejection
// message that lies about what is legal, and a slice reordered by
// accident reorders the message every load error prints.
func TestAlignSetsAreTwoSetsPinnedAgainstTheirMaps(t *testing.T) {
	for _, c := range []struct {
		name   string
		tokens []string
		set    map[string]bool
		want   []string
	}{
		{"style", StyleAlignTokens, closedStyleAligns, []string{"left", "center", "right", "justify"}},
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

	// THE SPLIT ITSELF, stated as the two facts D-7.3.1 turns on.
	if !closedStyleAligns[AlignJustify] {
		t.Error("style.align must admit justify (FR47)")
	}
	if closedColumnAligns[AlignJustify] {
		t.Fatal("columns[].align must NEVER admit justify — justified table cells are a separate scope decision, not a side effect of a map edit (D-7.3.1)")
	}
	// And neither set is the other by subtraction: every column value is
	// a style value, and the style set is exactly one larger.
	for _, tok := range ColumnAlignTokens {
		if !closedStyleAligns[tok] {
			t.Errorf("%q is a legal column align but not a legal style align — the two sets have diverged in the wrong direction", tok)
		}
	}
	if len(StyleAlignTokens) != len(ColumnAlignTokens)+1 {
		t.Errorf("the style set has %d members and the column set %d — after Story 7.3 they differ by exactly one, `justify`", len(StyleAlignTokens), len(ColumnAlignTokens))
	}

	// IsStyleAlign is the exported single source the property-command
	// path validates through; it must agree with the map it wraps.
	for _, tok := range StyleAlignTokens {
		if !IsStyleAlign(tok) {
			t.Errorf("IsStyleAlign(%q) = false", tok)
		}
	}
	if IsStyleAlign("middle") || IsStyleAlign("") {
		t.Error("IsStyleAlign admits a value the closed set does not")
	}
}

// TestAlignClosedSetsRejectAtTheRightSiteWithTheirOwnMessage is the
// split's BEHAVIOURAL red-proof: the same word, `justify`, must load at
// the two style attachment points and be a located load error at the
// column one — and each message must list exactly the members of the set
// that rejected it, not the other set's.
func TestAlignClosedSetsRejectAtTheRightSiteWithTheirOwnMessage(t *testing.T) {
	// (a) justify LOADS at style.align.
	if _, err := ParseDocument(alignDocWithStyle(`"align": "justify"`)); err != nil {
		t.Errorf("style.align: \"justify\" must load: %v", err)
	}
	// (b) justify LOADS at headerStyle.align — the second style
	//     attachment point, which decodeStyle serves through the same
	//     fieldPrefix parameter.
	if _, err := ParseDocument(alignDocWithTable(`"align": "justify"`, "")); err != nil {
		t.Errorf("headerStyle.align: \"justify\" must load: %v", err)
	}
	// (c) justify is REFUSED at columns[].align, located at the column.
	_, err := ParseDocument(alignDocWithTable("", `, "align": "justify"`))
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

	// (d) An illegal STYLE value is rejected with the STYLE set, which
	//     includes justify. This is the honesty half: before the split
	//     both sites printed one hand-written literal, so extending
	//     either set would have shipped a message that lied.
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

	// (e) …and at headerStyle it names headerStyle, not its sibling.
	_, herr := ParseDocument(alignDocWithTable(`"align": "middle"`, ""))
	if herr == nil {
		t.Fatal("headerStyle.align: \"middle\" must be a load error")
	}
	if !strings.Contains(herr.Error(), "headerStyle.align") {
		t.Errorf("the headerStyle rejection must be located at headerStyle.align, got: %v", herr)
	}

	// (f) THE MESSAGES ARE DERIVED, not restated: each one contains the
	//     exact rendering of its own ordered slice. A future edit to
	//     either set changes both the enforcement and the message, or
	//     this fails.
	if !strings.Contains(msg, closedSetMessage(ColumnAlignTokens)) {
		t.Errorf("the column message is not derived from ColumnAlignTokens: %v", err)
	}
	if !strings.Contains(smsg, closedSetMessage(StyleAlignTokens)) {
		t.Errorf("the style message is not derived from StyleAlignTokens: %v", serr)
	}
}
