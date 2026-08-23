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
