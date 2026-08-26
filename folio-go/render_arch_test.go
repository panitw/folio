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

// funcSet is a package-folio-scoped call graph: for each callable node
// declared in a non-test file directly under folio-go/, the set of
// other same-package callable nodes it calls directly, plus whether
// its body calls into internal/pdf at all.
type funcCallInfo struct {
	callsProducerPkg bool            // body calls into internal/pdf directly (any selector)
	calls            map[string]bool // keys of same-package callable nodes called directly
}

// buildFolioCallGraph parses every non-test .go file directly under
// dir (the module-root package "folio") and returns, per callable
// node, its direct-call info.
//
// THREE kinds of top-level callable are nodes (D-3.7.9(a), QA Finding
// 2, superseding this story's original "state the residual" fix — the
// engineering lead ruled the walker must FAIL CLOSED, not merely admit
// what it misses):
//
//   - a receiverless function, keyed by its own name (unchanged since
//     Story 1.7);
//   - a top-level var initialised directly to a function literal
//     (`var v = func(...) {...}`), keyed by the var's name — closes
//     the "func-typed variable" escape route the review measured
//     (a package-level `sneakyEmit` reaching internal/pdf was
//     previously a call edge to a node that DID NOT EXIST, which
//     `reaches`/the walk in TestValidateNeverReachesRenderOrInternalPDF
//     silently treated as "nothing further to see" rather than as an
//     unresolved, and therefore suspect, call target);
//   - a METHOD (fd.Recv != nil), keyed by "."+methodName, WITHOUT
//     regard to its receiver type. An AST-only scan cannot resolve a
//     selector expression's static receiver type without go/types (the
//     lead's own note: the COMPLETE version of this check is a lint
//     rule over go/types, deferred with a real trigger — before the
//     folio-go/v0.1.0 tag, since that is when Validate's contract
//     freezes publicly). Until then, a call `x.Foo()` is treated as
//     reaching EVERY method named Foo declared anywhere in package
//     folio — "name-matching over-approximates, which is the safe
//     direction: a spurious edge only makes the guard stricter"
//     (D-3.7.9(a), verbatim). This closes the "method value" escape
//     route the review measured (a `sneakyRenderer{}.emit()` call was
//     previously invisible: methods were not graph nodes at all,
//     fd.Recv != nil was skipped outright), and the same reasoning
//     covers interface dispatch (`var r Renderer = sneakyRenderer{};
//     r.emit()`) for free, because the match is on the SELECTOR NAME
//     alone — X's type, static or dynamic, is never consulted. Nodes
//     sharing a method name are MERGED (union of calls and
//     callsProducerPkg), because the over-approximation cannot tell
//     them apart anyway.
//
// A selector call fn.Sel.Name that matches neither the internal/pdf
// import alias nor any locally-declared method name is, BY
// CONSTRUCTION, not a call to anything package folio declares: it is
// necessarily some other package's exported function or method (fmt,
// strings, a stdlib type, another internal/* package, ...), which is
// why no separate "unresolved call" bucket is needed to fail the walk
// closed. Every call site directly under folio-go/ resolves to exactly
// one of: (a) a same-package function, (b) a same-package func-typed
// var, (c) a same-package method (by name), (d) an internal/pdf call
// (tracked separately as callsProducerPkg), or (e)
// definitionally-external — case (e) being safe by this reasoning, not
// by omission. (At HEAD, measured: package folio declares exactly
// seven methods — Severity.String, RenderError.Error,
// RenderError.Unwrap, fontCache.get, and faceSegment's
// segmentLocal/glyphRangeForRunes/advance1000 — none of which call
// into internal/pdf, and zero top-level func-typed vars, so this
// change does not alter TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt's
// producer count.)
func buildFolioCallGraph(t *testing.T, dir string) map[string]funcCallInfo {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	fset := token.NewFileSet()
	type parsedFile struct {
		file     *ast.File
		pdfAlias string
	}
	var files []parsedFile

	// Pass 1: parse every file and collect the COMPLETE, package-wide
	// set of method names — the selector-resolution step below needs
	// the full set, not just whatever the current file happens to
	// declare.
	methodNames := map[string]bool{}
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
		files = append(files, parsedFile{file: file, pdfAlias: pdfAlias})

		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Recv != nil {
				methodNames[fd.Name.Name] = true
			}
		}
	}

	inspectBody := func(body *ast.BlockStmt, pdfAlias string) funcCallInfo {
		info := funcCallInfo{calls: map[string]bool{}}
		ast.Inspect(body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				// A same-package function, or func-typed var, call —
				// the fact that matters for "routes through it".
				info.calls[fn.Name] = true
			case *ast.SelectorExpr:
				// Finding 1 (Story 1.7's review): match ANY call
				// resolved through the internal/pdf import alias, not
				// a single named selector — a whitelisted function
				// name is a property about that one function, not
				// about "produces the document bytes".
				if pkgIdent, ok := fn.X.(*ast.Ident); ok && pkgIdent.Name == pdfAlias && pdfAlias != "" {
					info.callsProducerPkg = true
					return true
				}
				// D-3.7.9(a): over-approximate any selector whose
				// name matches a locally-declared method, regardless
				// of X's type (see the doc comment above).
				if methodNames[fn.Sel.Name] {
					info.calls["."+fn.Sel.Name] = true
				}
			}
			return true
		})
		return info
	}

	graph := map[string]funcCallInfo{}
	merge := func(key string, info funcCallInfo) {
		existing, ok := graph[key]
		if !ok {
			graph[key] = info
			return
		}
		existing.callsProducerPkg = existing.callsProducerPkg || info.callsProducerPkg
		for c := range info.calls {
			existing.calls[c] = true
		}
		graph[key] = existing
	}

	// Pass 2: build the graph. Every FuncDecl (function or method) and
	// every top-level `var v = func(...) {...}` becomes a node.
	for _, pf := range files {
		for _, decl := range pf.file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Body == nil {
					continue
				}
				info := inspectBody(d.Body, pf.pdfAlias)
				if d.Recv != nil {
					merge("."+d.Name.Name, info)
				} else {
					merge(d.Name.Name, info)
				}
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
						continue
					}
					lit, ok := vs.Values[0].(*ast.FuncLit)
					if !ok {
						continue
					}
					merge(vs.Names[0].Name, inspectBody(lit.Body, pf.pdfAlias))
				}
			}
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

// TestValidateNeverReachesRenderOrInternalPDF is Story 3.7's AC1: "no
// render is attempted" is a STRUCTURAL property of Validate, asserted
// by AST over package folio's own top-level functions — never inferred
// from the absence of returned bytes, which a Validate that renders
// and throws the bytes away would also satisfy. Reuses
// buildFolioCallGraph (this file, D-000.42: no second call-graph
// builder) rather than writing a new one.
//
// Two independent checks, both required:
//
//  1. Validate's call graph never reaches the two NAMED functions that
//     can produce document bytes today — "renderDocument" and
//     "buildPageModel" — by same-package call. (predictDocument, the
//     function Validate DOES call and shares with buildPageModel, is
//     deliberately NOT on this forbidden list: it is the shared
//     derivation D-000.42 requires, and it is covered by check 2
//     instead, which is the stronger, name-independent property.)
//  2. Nothing reachable from Validate — including Validate itself —
//     calls into internal/pdf at all (any selector through the pdf
//     import alias), per funcCallInfo.callsProducerPkg. This is the
//     real, name-independent form of "no render is attempted": it
//     catches a render call added under ANY name, not just the two
//     named above.
//
// QA Finding 2 (this story's review, Major) / D-3.7.9(a): check 2's
// claim was TRUE OF THE INTENT but not yet of the CODE — the review
// measured that buildFolioCallGraph skipped every method declaration
// (fd.Recv != nil) outright and only ever resolved a bare identifier
// call, so a same-package METHOD (`sneakyRenderer{}.emit()`) or a
// package-level FUNC-TYPED VARIABLE (`sneakyEmit()`) reaching
// internal/pdf from inside Validate built, ran, and left this test
// GREEN. buildFolioCallGraph now resolves both (see its own doc
// comment: methods become graph nodes, matched by name across every
// receiver type, deliberately over-approximating; func-typed vars
// become graph nodes too) — "under ANY name" is now true BY
// CONSTRUCTION, not by claim. The one acknowledged residual is
// go/types-precision receiver resolution, deferred to a lint rule
// before the folio-go/v0.1.0 tag (see buildFolioCallGraph's comment);
// it can only produce SPURIOUS (over-approximating) edges, never miss
// a real one, so it cannot make this guard pass when it should fail.
func TestValidateNeverReachesRenderOrInternalPDF(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve module root: %v", err)
	}
	graph := buildFolioCallGraph(t, root)

	if _, ok := graph["Validate"]; !ok {
		t.Fatalf("vacuity guard: Validate not found among package folio's top-level functions")
	}

	for _, forbidden := range []string{"renderDocument", "buildPageModel"} {
		if _, ok := graph[forbidden]; !ok {
			t.Fatalf("vacuity guard: %q not found among package folio's top-level functions — AC1's property is unassertable", forbidden)
		}
		if reaches(graph, "Validate", forbidden, map[string]bool{}) {
			t.Errorf("AC1: Validate's call graph reaches %q — Validate must never attempt a render", forbidden)
		}
	}

	// Check 2: walk Validate's own reachable set and assert none of
	// them call into internal/pdf.
	visited := map[string]bool{}
	var walk func(fn string)
	var offenders []string
	walk = func(fn string) {
		if visited[fn] {
			return
		}
		visited[fn] = true
		info, ok := graph[fn]
		if !ok {
			return
		}
		if info.callsProducerPkg {
			offenders = append(offenders, fn)
		}
		for callee := range info.calls {
			walk(callee)
		}
	}
	walk("Validate")
	if len(offenders) > 0 {
		t.Errorf("AC1: Validate's call graph reaches internal/pdf through %v — no render may be attempted", offenders)
	}
}
