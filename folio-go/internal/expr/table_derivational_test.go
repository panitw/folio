package expr

// This file replaces TestThreeImplementedFunctions/
// TestFiveUnimplementedFunctions (Story 3.1/3.2's hard-coded name
// lists, F11/D-3.1a.3, AC29): "a guard whose expected value must be
// edited is one that gets edited wrongly" — those two tests named
// {"upper","lower","if"} and {"sum","count","avg","formatDate",
// "formatNumber"} verbatim, so Story 3.3's own edit (three of the five
// flip) and Story 3.4's (the remaining two flip) each required editing
// the guard in the same diff as the thing it guards, exactly the shape
// D-3.1a.3 warns against.
//
// Restated DERIVATIONALLY instead (AC29): every entry with
// implemented == false names an owningStory, and every entry with
// implemented == true has an evalCall (eval.go) switch branch —
// checked by AST-scanning evalCall's own "switch entry.name" for the
// set of names its case clauses match, never by a literal list here.
// So stated, this survives 3.3's edit (three names move across) and
// 3.4's (two more) with NO edit to this file at all.

import (
	"go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"os"
	"testing"
)

// evalCallSwitchNames AST-scans eval.go's evalCall function for its
// "switch entry.name { case "x", "y": … }" statement, returning every
// string literal any case clause matches (excluding "default", which
// carries no string literal at all and is evalCall's own
// unreachable-in-practice safety net, eval.go's own comment).
func evalCallSwitchNames(t *testing.T) map[string]bool {
	t.Helper()
	src, err := os.ReadFile("eval.go")
	if err != nil {
		t.Fatalf("read eval.go: %v", err)
	}
	fset := gotoken.NewFileSet()
	f, err := goparser.ParseFile(fset, "eval.go", src, 0)
	if err != nil {
		t.Fatalf("parse eval.go: %v", err)
	}

	names := map[string]bool{}
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "evalCall" {
			return true
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			sw, ok := n.(*ast.SwitchStmt)
			if !ok {
				return true
			}
			tag, ok := sw.Tag.(*ast.SelectorExpr)
			if !ok || tag.Sel.Name != "name" {
				return true // not the "switch entry.name" statement
			}
			found = true
			for _, stmt := range sw.Body.List {
				cc, ok := stmt.(*ast.CaseClause)
				if !ok || cc.List == nil {
					continue // cc.List == nil is the "default" clause
				}
				for _, expr := range cc.List {
					lit, ok := expr.(*ast.BasicLit)
					if !ok || lit.Kind != gotoken.STRING {
						t.Fatalf("evalCall's switch matches a non-literal case expression %#v — this scan cannot derive the implemented set from it", expr)
					}
					v := lit.Value
					if len(v) >= 2 {
						v = v[1 : len(v)-1]
					}
					names[v] = true
				}
			}
			return false
		})
		return true
	})
	if !found {
		t.Fatal("presence precondition (D-000.9): evalCall's \"switch entry.name\" statement was not found — the scan found nothing to disagree with")
	}
	return names
}

// TestImplementedEntriesMatchEvalCallSwitch is AC29's derivational
// restatement: functionTable's implemented==true set must equal
// evalCall's own switch-case set, in both directions — an entry
// flipped to implemented without a branch is caught (evalCall would
// hit its own "has no evaluator" internal error at runtime, but this
// test catches it at build time instead), and a stray switch branch
// for an entry NOT marked implemented is caught too (dead code the
// implemented-guard above it can never reach).
func TestImplementedEntriesMatchEvalCallSwitch(t *testing.T) {
	switchNames := evalCallSwitchNames(t)

	tableImplemented := map[string]bool{}
	for _, e := range functionTable {
		if e.implemented {
			tableImplemented[e.name] = true
		} else if e.owningStory == "" {
			t.Errorf("%s: implemented=false entry must name its owning story", e.name)
		}
	}

	for name := range tableImplemented {
		if !switchNames[name] {
			t.Errorf("%s: functionTable marks this implemented, but evalCall's switch has no case for it", name)
		}
	}
	for name := range switchNames {
		if !tableImplemented[name] {
			t.Errorf("%s: evalCall's switch has a case for this, but functionTable does not mark it implemented", name)
		}
	}
	if len(tableImplemented) == 0 {
		t.Fatal("presence precondition (D-000.9): zero implemented entries found — the comparison above would pass vacuously")
	}
}

// TestUnimplementedEntriesHaveNoEvalCallBranch is AC29's restatement,
// CORRECTED (Story 3.3 finisher pass, Finding 6): this test used to be
// TestExactlyTwoUnimplementedRemainAfterThisStory, hard-coding `if n !=
// 2` — re-introducing, inside the very test AC29 exists to make
// edit-proof, the hard-coded expected value D-3.1a.3 names as the
// hazard ("a guard whose expected value must be edited is one that
// gets edited wrongly"). Story 3.4 flipping formatDate/formatNumber to
// implemented=true makes n==0, and the count assertion above would
// have failed, forcing 3.4 to edit this guard in the same diff as the
// thing it guards — the exact defeat this file's own top comment
// claims does not happen. (The test's own doc comment additionally
// named a non-existent function,
// "TestExactlyThreeUnimplementedRemainAfterThisStory", and used the
// word "count" in a test name, which D-2.5.1 — quoted two files away in
// resolution_roots_arch_test.go — forbids on both counts.)
//
// Restated derivationally instead, as the INVERSE of
// TestImplementedEntriesMatchEvalCallSwitch above: every entry NOT
// marked implemented must have NO evalCall switch branch. This holds
// for any table state — 3.3's edit, 3.4's, and whatever comes after —
// with no edit to this file.
func TestUnimplementedEntriesHaveNoEvalCallBranch(t *testing.T) {
	switchNames := evalCallSwitchNames(t)

	unimplemented := 0
	for _, e := range functionTable {
		if e.implemented {
			continue
		}
		unimplemented++
		if switchNames[e.name] {
			t.Errorf("%s: functionTable marks this NOT implemented, but evalCall's switch has a case for it — dead code the implemented-guard above it can never reach", e.name)
		}
	}
	if unimplemented == 0 {
		t.Fatal("presence precondition (D-000.9): zero unimplemented entries found — the comparison above would pass vacuously")
	}
}
