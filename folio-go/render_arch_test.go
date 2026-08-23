package folio

// AC4/AC4a (D-1.7.2): "exactly one function produces the document
// bytes, and both Render and RenderTo route through it" is asserted
// STRUCTURALLY, by AST — never as a behavioural byte comparison
// between Render's output and RenderTo's output. Such a comparison
// would be exactly the vacuous test D-1.7.2 refuses to write: a shared
// core is the only correct design (two emitters would be D-1.4.14's
// drift hazard), which is precisely why the comparison could never
// fail. Byte identity holds as a THEOREM of this structural shape, not
// as a test that cannot fail (D-1.7.2, verbatim: "Delete the vacuous
// behavioural comparison rather than keeping it for comfort — a green
// test nobody can fail reads as coverage.").
//
// "Producer" is defined precisely (M-2): a top-level function,
// declared in a non-test file directly under folio-go/ (package
// folio's own entry points — NOT internal/pdf, which exports two
// byte-producing functions of its own and would make this find three
// producers and redden for the wrong reason, AC26 Q3), whose body
// calls into internal/pdf AT ALL — any selector resolved through the
// internal/pdf import alias, not just SerializeTextDocument. "Routes
// through it" means Render/RenderTo reach that producer via a
// direct-call graph built over package folio's own top-level
// functions.
//
// This story's review, Finding 1 (Major): the original selector match
// named exactly "SerializeTextDocument", which is the ONE selector
// RP-2's own mutation happened to use — so RP-2 proved the guard fires
// on the shape it was written for, not that the property holds. A
// second byte-producing function calling internal/pdf's OTHER exported
// entry point, Serialize() (M-2 names it explicitly as the fixture
// document's producer), built and passed undetected. Matching on "any
// call through the pdf alias" closes that: internal/pdf currently has
// exactly one call site from package folio (renderDocument's call to
// SerializeTextDocument), so this is not a broadening in practice
// today — it is a narrowing of what future code can get away with.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// funcSet is a package-folio-scoped call graph: for each top-level
// function declared in a non-test file directly under folio-go/, the
// set of other top-level (same-scope) functions it calls directly,
// plus whether its body calls into internal/pdf at all.
type funcCallInfo struct {
	callsProducerPkg bool            // body calls into internal/pdf directly (any selector)
	calls            map[string]bool // names of same-package functions called directly
}

// buildFolioCallGraph parses every non-test .go file directly under
// dir (the module-root package "folio") and returns, per top-level
// function name, its direct-call info.
func buildFolioCallGraph(t *testing.T, dir string) map[string]funcCallInfo {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	fset := token.NewFileSet()
	graph := map[string]funcCallInfo{}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, perr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if perr != nil {
			t.Fatalf("parse %s: %v", path, perr)
		}

		// Resolve which local import alias, if any, names
		// internal/pdf, so the producer-detecting selector match
		// resolves the import path rather than matching literal text
		// "pdf." (AC26 Q3's point applied here too).
		pdfAlias := ""
		for _, imp := range file.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasSuffix(p, "/internal/pdf") {
				continue
			}
			alias := p[strings.LastIndex(p, "/")+1:]
			if imp.Name != nil {
				alias = imp.Name.Name
			}
			pdfAlias = alias
		}

		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || fd.Body == nil {
				continue
			}
			info := funcCallInfo{calls: map[string]bool{}}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				switch fn := call.Fun.(type) {
				case *ast.Ident:
					// A same-package function call — the fact that
					// matters for "routes through it".
					info.calls[fn.Name] = true
				case *ast.SelectorExpr:
					// Finding 1: match ANY call resolved through the
					// internal/pdf import alias, not a single named
					// selector — a whitelisted function name is a
					// property about that one function, not about
					// "produces the document bytes".
					if pkgIdent, ok := fn.X.(*ast.Ident); ok && pkgIdent.Name == pdfAlias && pdfAlias != "" {
						info.callsProducerPkg = true
					}
				}
				return true
			})
			graph[fd.Name.Name] = info
		}
	}
	return graph
}

// reaches reports whether starting function fn can reach target via
// zero or more direct same-package calls recorded in graph.
func reaches(graph map[string]funcCallInfo, fn, target string, visited map[string]bool) bool {
	if fn == target {
		return true
	}
	if visited[fn] {
		return false
	}
	visited[fn] = true
	info, ok := graph[fn]
	if !ok {
		return false
	}
	for callee := range info.calls {
		if reaches(graph, callee, target, visited) {
			return true
		}
	}
	return false
}

// TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt
// is AC4/AC4a's structural guard.
func TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve module root: %v", err)
	}
	graph := buildFolioCallGraph(t, root)

	if len(graph) == 0 {
		t.Fatalf("vacuity guard: scanned zero top-level functions under the module root — AC4's property is unassertable (D-000.9)")
	}

	var producers []string
	for name, info := range graph {
		if info.callsProducerPkg {
			producers = append(producers, name)
		}
	}

	// AC4a: zero producers found is a hard failure, never "nothing to
	// check, pass".
	if len(producers) == 0 {
		t.Fatalf("AC4a: found ZERO functions in package folio that call internal/pdf.SerializeTextDocument — the document-byte producer is unassertable")
	}
	if len(producers) != 1 {
		t.Fatalf("AC4: expected EXACTLY ONE document-byte producer, found %d: %v", len(producers), producers)
	}
	producer := producers[0]

	entryPoints := []string{"Render", "RenderTo"}
	var missing []string
	for _, ep := range entryPoints {
		if _, ok := graph[ep]; !ok {
			t.Fatalf("vacuity guard: entry point %q not found among package folio's top-level functions", ep)
		}
		if !reaches(graph, ep, producer, map[string]bool{}) {
			missing = append(missing, ep)
		}
	}
	// AC4a: zero routing entry points found is a hard failure too.
	if len(missing) == len(entryPoints) {
		t.Fatalf("AC4a: NEITHER Render nor RenderTo routes through the producer %q — routing is unassertable", producer)
	}
	if len(missing) > 0 {
		t.Fatalf("AC4: %v does not route through the single document-byte producer %q — Render and RenderTo must both route through it", missing, producer)
	}
}
