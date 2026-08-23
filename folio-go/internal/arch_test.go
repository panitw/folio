// Package arch holds architectural fitness tests that apply across every
// package under folio-go/internal/, rather than to any one of them. It has
// no non-test files: it exists purely to run assertions like "no float64
// anywhere under internal/" that a single package's own test file cannot
// see past its own directory.
package arch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findFloatOccurrences walks a parsed file for any appearance at all of
// the identifier "float64" or "float32", plus any untyped floating-point
// literal (token.FLOAT). Flagging the bare identifier wherever it occurs
// in the AST — rather than only at specific declaration positions —
// subsumes every type-position shape in one pass: func parameters,
// results and type parameters; struct fields, including nested,
// array/slice/map/pointer/variadic and anonymous-struct forms; interface
// method signatures; type declarations and aliases; and explicit
// conversions such as float64(x), whose callee is itself an *ast.Ident
// named "float64".
//
// This replaces an earlier version that type-asserted expr.(*ast.Ident)
// only at f.Type / decl.Type for a fixed set of declaration kinds, and so
// caught only the four narrowest shapes: this story's own QA review
// (Blocker 2(b) / Major 7) measured nine further shapes that walk missed
// entirely, including a bare float64(x) conversion — which is exactly
// what let a float64 conversion plus float formatting reach an output
// byte with this guard staying green.
func findFloatOccurrences(fset *token.FileSet, file *ast.File, path string) []string {
	var findings []string
	ast.Inspect(file, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.Ident:
			if v.Name == "float64" || v.Name == "float32" {
				p := fset.Position(v.Pos())
				findings = append(findings, path+":"+itoa(p.Line)+": identifier "+v.Name)
			}
		case *ast.BasicLit:
			if v.Kind == token.FLOAT {
				p := fset.Position(v.Pos())
				findings = append(findings, path+":"+itoa(p.Line)+": untyped floating-point literal "+v.Value)
			}
		}
		return true
	})
	return findings
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// TestNoFloat64UnderInternal parses every package under folio-go/internal/
// and fails if any declaration, signature or literal mentions float64 or
// float32 (AC7 / AD-23). Guard against vacuity (guard 2): it asserts it
// actually visited the "geom" and "pdf" package directories **by name**,
// and counts declarations only from files in a real package directory
// (excluding this test's own directory, "."), so this cannot pass by
// walking zero files, and — the review's Minor 16 — cannot pass after
// deleting either real package while the other survives: the previous
// version's guard unconditionally counted this file's own directory and
// its own declarations, so it required only one real package to stay
// green, not two.
func TestNoFloat64UnderInternal(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve internal/ root: %v", err)
	}

	packagesVisited := map[string]bool{}
	declCount := 0
	var findings []string

	fset := token.NewFileSet()
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if perr != nil {
			return perr
		}

		rel, _ := filepath.Rel(root, path)
		pkgDir := filepath.Dir(rel)
		packagesVisited[pkgDir] = true

		if pkgDir != "." {
			ast.Inspect(file, func(n ast.Node) bool {
				switch n.(type) {
				case *ast.FuncDecl, *ast.TypeSpec, *ast.ValueSpec:
					declCount++
				}
				return true
			})
		}

		findings = append(findings, findFloatOccurrences(fset, file, rel)...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/: %v", err)
	}

	if !packagesVisited["geom"] || !packagesVisited["pdf"] {
		t.Fatalf("vacuity guard: expected to visit package directories \"geom\" and \"pdf\" by name, got %v", packagesVisited)
	}
	if declCount == 0 {
		t.Fatalf("vacuity guard: visited 0 declarations across real package directories (excluding this test's own directory)")
	}

	if len(findings) > 0 {
		t.Fatalf("float64/float32 found under internal/ (forbidden by AD-23):\n%s", strings.Join(findings, "\n"))
	}
}
