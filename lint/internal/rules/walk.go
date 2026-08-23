package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// fileVisit is called once per parsed .go file discovered by walkGoFiles.
type fileVisit func(rel string, file *ast.File, fset *token.FileSet) error

// walkGoFiles walks root and calls visit for every .go file, with its
// parsed AST and a path relative to root (AC4). It skips any directory
// literally named "testdata", at any depth, by directory name (AC2) —
// not by path prefix and not by a list of specific paths. It returns the
// first parse error verbatim (AC5, D-1.3.3 amended): a target directory
// that cannot be read must never be silently treated as "zero findings".
func walkGoFiles(root string, visit fileVisit) error {
	fset := token.NewFileSet()
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if perr != nil {
			return perr
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		return visit(rel, file, fset)
	})
}
