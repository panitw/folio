package bind

// Story 2.7, AC3: "no page namespace exists, and none can be added" —
// enforced structurally, over the subject that CAN express the
// violation (D-000.50). The expression language does not exist yet
// (Epic 3), so there is nothing to grep for "page" IN — a guard over an
// empty set cannot fail, which is not coverage. The subject that DOES
// exist today is BindTextSpans' resolution-root dispatch itself
// (the fence in BindText's own doc comment, text.go — cited by
// enclosing function rather than line number, this story's review,
// Finding 7: the line-number pointer here went stale once already,
// from this story's own edit): a "page" namespace comes into
// existence precisely when "page" becomes a THIRD lookupBound rootName,
// alongside "data" and "params".
//
// THE MECHANISM, same shape as internal/template/closedsets.go's
// closedAligns and byte_neutrality_test.go's declaredEpic2GateObligations
// (D-2.5.1): a DECLARED set (declaredResolutionRoots, text.go) compared,
// BOTH DIRECTIONS, against an OBSERVED set collected by AST-scanning
// this package's own source for every lookupBound call's rootName
// string literal. Never a count (D-2.5.1: "never a count, never in a
// test name").

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// packageSourceFiles lists every non-test .go file in this package's
// own directory (this story's review, Finding 2: the scan previously
// hardcoded "text.go" by name, so a third resolution root introduced in
// any OTHER file of package bind was invisible to it). Each is parsed
// and scanned SEPARATELY by observedResolutionRootsInFiles — no source
// concatenation, so nothing depends on file boundaries or on there
// being exactly one "package" clause.
func packageSourceFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, name)
	}
	if len(files) == 0 {
		t.Fatal("presence precondition: no non-test .go files found in this package directory")
	}
	slices.Sort(files) // deterministic order, D-1.3.5
	return files
}

// observedResolutionRoots AST-scans src (Go source) for every call
// site of a function named funcName, and collects the string-literal
// value of the ARGUMENT AT POSITION argIndex (0-based) from each call
// — text.go's lookupBound(root, subPath, fullPath, elementID, rootName,
// rootDesc), rootName at index 4.
//
// A call whose relevant argument is NOT a string literal (e.g. a
// variable) is reported as a finding rather than silently skipped —
// the exact shape a THIRD root smuggled in through a computed name
// would take, and this guard must not wave it through.
func observedResolutionRoots(t *testing.T, filename string, src []byte, funcName string, argIndex int) (roots []string, dynamicCallSites int) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != funcName {
			return true
		}
		if len(call.Args) <= argIndex {
			return true
		}
		lit, ok := call.Args[argIndex].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			dynamicCallSites++
			return true
		}
		// lit.Value carries the Go source quoting; strconv.Unquote would
		// be the general tool, but every root name here is a plain
		// double-quoted ASCII identifier, so trimming the quotes is
		// exact and avoids importing strconv for one call.
		v := lit.Value
		if len(v) >= 2 {
			v = v[1 : len(v)-1]
		}
		roots = append(roots, v)
		return true
	})
	return roots, dynamicCallSites
}

// observedResolutionRootsAcrossPackage scans EVERY non-test .go file in
// this package's directory (packageSourceFiles), aggregating roots and
// dynamicCallSites across all of them — this story's review, Finding 2,
// first evasion: the scan used to hardcode "text.go" by name, so a
// third root introduced in any OTHER file of package bind was invisible
// to it.
//
// overrideFile/overrideSrc substitute IN-MEMORY content for the one
// named file (used by the red-proof below to scan a MUTATED copy of
// text.go without ever writing it to disk); every other file is read
// fresh. Pass "" / nil to scan the package exactly as committed.
func observedResolutionRootsAcrossPackage(t *testing.T, funcName string, argIndex int, overrideFile string, overrideSrc []byte) (roots []string, dynamicCallSites int) {
	t.Helper()
	for _, name := range packageSourceFiles(t) {
		src := overrideSrc
		if name != overrideFile || overrideSrc == nil {
			var err error
			src, err = os.ReadFile(name)
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
		}
		r, d := observedResolutionRoots(t, name, src, funcName, argIndex)
		roots = append(roots, r...)
		dynamicCallSites += d
	}
	return roots, dynamicCallSites
}

// TestBindResolutionRootsAreClosed is AC3's structural guard.
//
// A NARROWER, secondary check alongside
// TestBindTextDottedPageAndPagesPathsAreOrdinaryDataPaths
// (text_test.go), which is now AC3's PRIMARY guard: this scan is keyed
// on lookupBound call sites — a proxy this story's review (Blocker 2)
// showed can be defeated by an early-return dispatch that never calls
// lookupBound at all. Left in place because the narrower shape it DOES
// catch (an extra lookupBound root) is real, and because a structural
// signal that fires at the exact call site is more useful for a future
// reader than the behavioural test alone.
//
// PRESENCE PRECONDITION (D-000.9/D-000.21): the scan must find at least
// one lookupBound call site, or "the observed set equals the declared
// set" is satisfied vacuously by an empty scan finding nothing to
// disagree with.
func TestBindResolutionRootsAreClosed(t *testing.T) {
	observed, dynamic := observedResolutionRootsAcrossPackage(t, "lookupBound", 4, "", nil)
	if dynamic > 0 {
		t.Fatalf("%d lookupBound call site(s) pass a NON-LITERAL rootName — a computed root name is exactly "+
			"how a third resolution root would be smuggled past a literal scan; make it a literal or extend "+
			"this guard to resolve it", dynamic)
	}
	if len(observed) == 0 {
		t.Fatal("presence precondition (D-000.9): zero lookupBound call sites found — \"the observed set " +
			"equals the declared set\" would be satisfied vacuously by an empty scan")
	}

	declaredSet := map[string]bool{}
	for _, r := range declaredResolutionRoots {
		declaredSet[r] = true
	}
	observedSet := map[string]bool{}
	for _, r := range observed {
		observedSet[r] = true
	}

	for _, d := range declaredResolutionRoots {
		if !observedSet[d] {
			t.Errorf("declared resolution root %q is never observed at a lookupBound call site — either "+
				"remove it from declaredResolutionRoots or restore its call site", d)
		}
	}
	for r := range observedSet {
		if !declaredSet[r] {
			t.Errorf("OBSERVED resolution root %q (a lookupBound call site's rootName) is not in "+
				"declaredResolutionRoots %v — a THIRD resolution root has appeared. If this is deliberate, "+
				"it is a direction change under AD-4 (\"no page namespace exists, and none can be added\") "+
				"and needs the engineering lead's ruling, not a one-line edit to this list",
				r, declaredResolutionRoots)
		}
	}

	wantSorted := append([]string(nil), declaredResolutionRoots...)
	slices.Sort(wantSorted)
	gotSorted := append([]string(nil), observed...)
	slices.Sort(gotSorted)
	t.Logf("declared resolution roots: %v; observed at %d lookupBound call site(s): %v", wantSorted, len(observed), gotSorted)
}

// TestBindResolutionRootsClosureRedProof is D-000.52: a structural
// claim about a guard is worth exactly as much as its demonstration.
// This introduces the violation the comment above describes — a THIRD
// lookupBound call site naming "page" as its root — into a SCRATCH
// COPY of text.go's source (never the committed file) and asserts
// TestBindResolutionRootsAreClosed's own scan catches it.
func TestBindResolutionRootsClosureRedProof(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(".", "text.go"))
	if err != nil {
		t.Fatalf("read text.go: %v", err)
	}

	mutated := injectThirdResolutionRoot(t, src)

	observed, dynamic := observedResolutionRootsAcrossPackage(t, "lookupBound", 4, "text.go", mutated)
	if dynamic > 0 {
		t.Fatalf("mutation produced a non-literal rootName call site — the mutation itself is malformed")
	}

	observedSet := map[string]bool{}
	for _, r := range observed {
		observedSet[r] = true
	}
	if !observedSet["page"] {
		t.Fatalf("presence precondition: the mutation was supposed to introduce a lookupBound(..., \"page\", ...) "+
			"call site, but the scan over the mutated source did not observe \"page\" among %v — the mutation "+
			"did not take", observed)
	}

	declaredSet := map[string]bool{}
	for _, r := range declaredResolutionRoots {
		declaredSet[r] = true
	}
	reddened := false
	for r := range observedSet {
		if !declaredSet[r] {
			reddened = true
		}
	}
	if !reddened {
		t.Fatal("RED-PROOF FAILED: a mutated source declaring a third lookupBound root \"page\" did not " +
			"trip the closure check — TestBindResolutionRootsAreClosed would pass on a document that " +
			"had actually acquired a page namespace")
	}
	t.Logf("red-proof: injecting lookupBound(..., \"page\", ...) is observed as %v, outside declared %v — the guard reddens", observed, declaredResolutionRoots)
}

// injectThirdResolutionRoot appends a syntactically valid extra
// lookupBound call naming "page" as its rootName to a COPY of src, in
// a throwaway function the parser accepts but nothing ever calls — the
// scan is over lookupBound CALL SITES textually, so it does not need to
// compile or execute, only to parse (matching this guard's own AST-scan
// mechanism exactly).
func injectThirdResolutionRoot(t *testing.T, src []byte) []byte {
	t.Helper()
	addition := "\n\nfunc zzRedProofThirdResolutionRoot() {\n" +
		"\t_, _ = lookupBound(Value{}, nil, nil, \"\", \"page\", \"a smuggled page namespace\")\n" +
		"}\n"
	out := make([]byte, 0, len(src)+len(addition))
	out = append(out, src...)
	out = append(out, addition...)
	return out
}
