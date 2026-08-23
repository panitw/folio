package pdf

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isForbiddenStrconvSelector reports whether name is one of the strconv
// functions D-1.1.b's guardrail names: the Format* family (including
// FormatFloat), Itoa, and the Append* family (including AppendFloat — the
// exact call this story's QA review used to demonstrate the hole: a text
// list that enumerated strconv.FormatFloat but not strconv.AppendFloat let
// a float64 conversion plus float formatting reach an output byte with
// every guard, the golden hash, vet and gofmt all silent — Blocker 2(a)).
func isForbiddenStrconvSelector(name string) bool {
	return name == "Itoa" || strings.HasPrefix(name, "Format") || strings.HasPrefix(name, "Append")
}

// isForbiddenFmtSelector reports whether name is a fmt function that
// formats or prints a value — "any fmt formatting call", D-1.1.b's
// guardrail text verbatim. There is no verb allow-list to keep in step
// with new fmt.* additions: none of these calls belong in internal/pdf
// outside numbers.go at all, so every formatting/printing entry point is
// forbidden regardless of which verbs it is given.
func isForbiddenFmtSelector(name string) bool {
	switch name {
	case "Sprint", "Sprintf", "Sprintln",
		"Fprint", "Fprintf", "Fprintln",
		"Print", "Printf", "Println",
		"Errorf", "Appendf", "Append":
		return true
	default:
		return false
	}
}

// importAliases maps each import's local name (its explicit alias, or the
// last path segment when unaliased) to its full import path, so a call
// can be matched by the package it actually resolves to rather than by
// literal source text.
func importAliases(file *ast.File) map[string]string {
	aliases := map[string]string{}
	for _, imp := range file.Imports {
		p := strings.Trim(imp.Path.Value, `"`)
		name := p
		if idx := strings.LastIndex(p, "/"); idx != -1 {
			name = p[idx+1:]
		}
		if imp.Name != nil {
			name = imp.Name.Name
		}
		aliases[name] = p
	}
	return aliases
}

func itoa10(n int) string {
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

// TestNumberFormattingIsConfinedToNumbersGo is AC6's source-level guard,
// merged with D-1.1.b's file-scoped guardrail. D-1.1.b's guardrail text
// was written for Story 1.3's lint; it is enforced here, in Story 1.1,
// because this story is where the hole it closes actually exists (this
// story's QA review, Blocker 2) — it is brought forward rather than left
// for 1.3.
//
// Matching resolves each call's import alias and its selector name via
// go/ast on the parsed call expression, not literal source text or
// regular expressions: a text/regex scan that enumerates
// strconv.FormatFloat but not strconv.AppendFloat is exactly the
// incomplete guard Blocker 2(a) measured, and comments (including this
// package's own doc comments, which name these forbidden calls to explain
// why they are forbidden) are never part of the AST's call expressions,
// so no comment-stripping step is needed to avoid the guard tripping on
// its own documentation.
//
// Two rules, applied in one walk:
//   - Anywhere under internal/pdf, in every file except numbers.go
//     (including _test.go files — no exemption): no strconv.Format*,
//     strconv.Itoa, strconv.Append*, or fmt formatting call. This is
//     D-1.1.b's guardrail, verbatim.
//   - Everywhere else under folio-go/ (outside _test.go files): the same
//     forbidden set, per AC6's broader "nowhere under folio-go/ outside
//     _test.go files" wording. Nothing outside internal/pdf writes an
//     output byte at all (AD-5), so this half is belt-and-braces on files
//     that have no legitimate reason to need it.
func TestNumberFormattingIsConfinedToNumbersGo(t *testing.T) {
	root := repoRootFromTest(t)
	folioGo := filepath.Join(root, "folio-go")
	pdfDir := filepath.Join(folioGo, "internal", "pdf")

	fset := token.NewFileSet()
	filesChecked := 0
	var findings []string

	err := filepath.WalkDir(folioGo, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		inPDFDir := filepath.Dir(path) == pdfDir
		isTest := strings.HasSuffix(path, "_test.go")
		isNumbersGo := inPDFDir && filepath.Base(path) == "numbers.go"

		if isNumbersGo {
			return nil // the one file this guardrail exempts
		}
		if !inPDFDir && isTest {
			return nil // AC6's broader rule exempts _test.go outside internal/pdf
		}

		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		filesChecked++

		aliases := importAliases(file)
		rel, _ := filepath.Rel(root, path)

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			pkgPath, known := aliases[pkgIdent.Name]
			if !known {
				return true
			}
			pos := fset.Position(call.Pos())
			switch {
			case pkgPath == "strconv" && isForbiddenStrconvSelector(sel.Sel.Name):
				findings = append(findings, rel+":"+itoa10(pos.Line)+": strconv."+sel.Sel.Name)
			case pkgPath == "fmt" && isForbiddenFmtSelector(sel.Sel.Name):
				findings = append(findings, rel+":"+itoa10(pos.Line)+": fmt."+sel.Sel.Name)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk folio-go/: %v", err)
	}

	// Vacuity guard: this must actually have looked at more than zero
	// files (document.go and emit_source_test.go itself, at minimum, are
	// both in internal/pdf and are not numbers.go).
	if filesChecked == 0 {
		t.Fatal("vacuity guard: checked 0 files")
	}

	if len(findings) > 0 {
		t.Fatalf("forbidden strconv/fmt formatting call(s) found outside numbers.go (AC6, D-1.1.b):\n%s", strings.Join(findings, "\n"))
	}
}
