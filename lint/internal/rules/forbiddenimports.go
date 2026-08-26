package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

// RuleForbiddenImports is this guard's stable rule id for a banned
// import path (AC4). Split from the math-selector rule below (Finding
// 10, this story's QA review): one checker previously emitted a single
// generic rule id for two structurally different violations, which
// degraded AC1/AC4's "by file and rule" fixture assertion to "by file
// alone" whenever two rules could plausibly hit the same file.
const RuleForbiddenImports = "forbidden-imports"

// RuleMathSelector is this guard's stable rule id for a math-selector
// violation — a disallowed `math` function call, or a non-call
// reference to a float-valued or unrecognised `math` member (AC10,
// AC12).
const RuleMathSelector = "math-selector"

// RuleDotImport is this guard's stable rule id for a dot import — `.
// "pkg"` — under internal/ (Finding 15, Story 3.4's QA review, ruled
// wider than the finding itself by the engineering lead).
//
// WHY THIS EXISTS AND WHY IT IS NOT SCOPED TO `math`: RuleMathSelector
// above (D-1.3.10) resolves a call's package by requiring its
// *ast.SelectorExpr's X to be a bare *ast.Ident, looked up in
// importAliases — the exact shape a dot import defeats, because
// `. "math"` makes `Pow(10, 2)` a plain, unqualified call with no
// SelectorExpr at all. That same aliasing assumption is what every
// OTHER selector-keyed ban in this file — and in effect every path ban
// in bannedImportPaths that a caller might reference by selector after
// aliasing it away — rests on. Patching mathAllowedCalls alone would
// leave time/os/net/math-rand equally exposed to the identical
// spelling and fix only the one instance a finding happened to name
// (D-000.23's failure: a guard written for the defect rather than its
// class). One predicate on the import spec's own local name being "."
// closes all of them at once, and closes every future selector-keyed
// rule this file gains, by construction. There is no legitimate
// dot-import under internal/ — none exists in this tree today (D-000.50:
// measured, not assumed) and Go style guidance discourages the form
// universally outside small a test-assertion-library idiom this
// codebase does not use anywhere.
const RuleDotImport = "dot-import"

// RuleRuntimeCaller is Story 2.1's addition (AC2, V1): AD-1's render
// path must never depend on the filesystem or on inspecting its own
// call stack (both are exactly what would make `internal/text`'s
// embedded-trie loader behave differently on js/wasm than natively).
// "os" is already a banned import path; `runtime.Caller` is a
// SELECTOR ban on an otherwise-unbanned package, exactly like
// RuleMathSelector's treatment of `math` — importing "runtime" itself
// (e.g. for runtime.GOOS, runtime.Version) is not banned; only the
// Caller selector is.
const RuleRuntimeCaller = "runtime-caller-selector"

// bannedImportPaths is AD-1's Rule, verbatim: "Render-path code may not
// import time, os, math/rand, [or] net". math is deliberately absent —
// it is judged by selector, not by import path (AC12, D-1.3.10: AD-1
// bans four package paths, then names a class of functions, and its next
// sentence enumerates seven tolerated math functions — a rule cannot
// both ban a package and list which of its functions are permitted).
var bannedImportPaths = map[string]bool{
	"time":      true,
	"os":        true,
	"math/rand": true,
	"net":       true,
}

// isBannedImportPath matches a banned path OR any of its subpackages
// (Finding 6, this story's QA review): the original exact-match test let
// `math/rand/v2`, `net/http`, `net/url` and `os/exec` all pass with zero
// findings — `math/rand/v2` shipped in Go 1.22, this module pins Go
// 1.26, and it is the modern spelling of the exact package AD-1 bans.
// AD-1's Prevents line is "determinism eroding one reasonable-looking
// commit at a time"; an exact-match deny-list is exactly that erosion
// path. This mirrors D-1.3.10's own reasoning for the `math` rule,
// applied to import paths: a rule must stay fail-safe against a
// plausible variant it did not anticipate.
func isBannedImportPath(path string) bool {
	for banned := range bannedImportPaths {
		if path == banned || strings.HasPrefix(path, banned+"/") {
			return true
		}
	}
	return false
}

// testExemptImportPaths is D-1.3.1's `_test.go` exemption, and nothing
// else is ever added to it (AC11): os, testing, path/filepath and embed
// are permitted in `_test.go` files. time, math/rand and net stay banned
// in tests too — only "os" is actually on bannedImportPaths, so this set
// exists to keep the exemption's own scope explicit and closed rather
// than implicit in "whatever isn't banned". This set stays exact-match
// (Finding 6's fix is scoped to the ban, not to this closed exemption).
var testExemptImportPaths = map[string]bool{
	"os":            true,
	"testing":       true,
	"path/filepath": true,
	"embed":         true,
}

// mathAllowedCalls is AD-1's closed allow-list of tolerated math
// functions (AC12, D-1.3.10): "Sqrt, Floor, Ceil, Round, Trunc, Abs,
// Mod". Any other math function call is a violation — an allow-list,
// not a deny-list, so a future Go release adding a transcendental is
// banned by default rather than silently permitted.
var mathAllowedCalls = map[string]bool{
	"Sqrt": true, "Floor": true, "Ceil": true, "Round": true,
	"Trunc": true, "Abs": true, "Mod": true,
}

// mathAllowedIntConstants is AC12's "integer-limit constants…
// permitted" half of the value-kind test for non-call math references.
// Everything not on this list — including math.Pi, math.E,
// math.MaxFloat64 and math.SmallestNonzeroFloat64 — is a violation
// (fail-safe default, matching the closed-allow-list shape of the call
// rule above).
var mathAllowedIntConstants = map[string]bool{
	"MaxInt": true, "MinInt": true,
	"MaxInt8": true, "MinInt8": true,
	"MaxInt16": true, "MinInt16": true,
	"MaxInt32": true, "MinInt32": true,
	"MaxInt64": true, "MinInt64": true,
	"MaxUint8": true, "MaxUint16": true, "MaxUint32": true, "MaxUint64": true,
}

// allowedNumericSurface is quoted verbatim in every failure message
// (AC10): AD-1's "allow-listed numeric surface".
const allowedNumericSurface = "+ - * /, comparison, and Sqrt, Floor, Ceil, Round, Trunc, Abs, Mod"

// importAliases maps each import's local name (its explicit alias, or
// the last path segment when unaliased) to its full import path, so a
// call or reference is matched by the package it actually resolves to,
// never by literal source text (AC12: "never a regex, never the literal
// text `math.`").
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

// ForbiddenImportsStats reports what ScanForbiddenImports actually
// examined, from the scanner's own execution (Major 5, this story's QA
// review) — see MapRangeStats' doc comment for why a second,
// independently-derived walk cannot be trusted as a vacuity guard.
//
// FilesSeenNames (Finding 14, Story 3.4's QA review) makes a BY-NAME
// coverage witness possible directly from the scanner's own record:
// FilesSeen alone is an int, so a by-name witness built against it had
// to re-derive its own, independent walk (walkGoFiles) to get names —
// exactly the "a vacuity guard built by re-deriving a fact a different
// way cannot see a scanner that silently does nothing" failure
// MapRangeStats' own doc comment warns against. A future early-skip
// added to ScanForbiddenImports now shows up here directly.
type ForbiddenImportsStats struct {
	DirsVisited    []string
	FilesSeen      int
	FilesSeenNames []string
}

// ScanForbiddenImports is the AC1 pure checker for AD-1's import and
// math-selector lint (AC10–AC13). It bans time/os/math/rand/net imports
// (and any subpackage of one, Finding 6) in non-test files under the
// target directory, applies D-1.3.1's `_test.go`-suffix-keyed exemption,
// and matches `math` by selector on the AST — resolving the import
// alias, never a regex (AC12) — as a closed allow-list of calls plus a
// value-kind test on non-call references. Any directory named testdata
// is excluded (AC2).
func ScanForbiddenImports(root string) ([]Finding, ForbiddenImportsStats, error) {
	var findings []Finding
	var stats ForbiddenImportsStats
	dirsSeen := map[string]bool{}
	err := walkGoFiles(root, func(rel string, file *ast.File, fset *token.FileSet) error {
		stats.FilesSeen++
		stats.FilesSeenNames = append(stats.FilesSeenNames, filepath.ToSlash(rel))
		dir := filepath.ToSlash(filepath.Dir(rel))
		if !dirsSeen[dir] {
			dirsSeen[dir] = true
			stats.DirsVisited = append(stats.DirsVisited, dir)
		}

		isTest := strings.HasSuffix(rel, "_test.go")
		aliases := importAliases(file)

		// Banned import paths, and their subpackages (AC10, AC11, Finding 6).
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if !isBannedImportPath(path) {
				continue
			}
			if isTest && testExemptImportPaths[path] {
				continue
			}
			pos := fset.Position(imp.Pos())
			findings = append(findings, Finding{
				Path: rel, Rule: RuleForbiddenImports, Line: pos.Line,
				Message: fmt.Sprintf("%s:%d: forbidden import %q (AD-1's allow-listed numeric surface: %s)",
					rel, pos.Line, path, allowedNumericSurface),
			})
		}

		// Dot imports (RuleDotImport, Finding 15): unconditional, no
		// test exemption, no allow-list — every selector-keyed ban in
		// this file (math, runtime.Caller) resolves its package by
		// requiring a bare *ast.Ident on the left of the dot, which a
		// dot import removes entirely by making the imported package's
		// exports unqualified identifiers.
		for _, imp := range file.Imports {
			if imp.Name == nil || imp.Name.Name != "." {
				continue
			}
			pos := fset.Position(imp.Pos())
			findings = append(findings, Finding{
				Path: rel, Rule: RuleDotImport, Line: pos.Line,
				Message: fmt.Sprintf("%s:%d: forbidden dot import %s — it removes every selector-keyed ban's ability to resolve the imported package (AD-1)",
					rel, pos.Line, imp.Path.Value),
			})
		}

		// math selector matching: distinguish call sites from non-call
		// references by first collecting which *ast.SelectorExpr nodes
		// are a CallExpr's Fun.
		callSelectors := map[*ast.SelectorExpr]bool{}
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					callSelectors[sel] = true
				}
			}
			return true
		})

		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			pkgPath, known := aliases[pkgIdent.Name]
			if !known || pkgPath != "math" {
				return true
			}
			pos := fset.Position(sel.Pos())
			if callSelectors[sel] {
				if !mathAllowedCalls[sel.Sel.Name] {
					findings = append(findings, Finding{
						Path: rel, Rule: RuleMathSelector, Line: pos.Line,
						Message: fmt.Sprintf("%s:%d: forbidden math.%s call — only %s are allowed (AD-1)",
							rel, pos.Line, sel.Sel.Name, allowedNumericSurface),
					})
				}
				return true
			}
			if !mathAllowedIntConstants[sel.Sel.Name] {
				findings = append(findings, Finding{
					Path: rel, Rule: RuleMathSelector, Line: pos.Line,
					Message: fmt.Sprintf("%s:%d: forbidden non-call reference math.%s (float-valued or unrecognised; only integer-limit constants are permitted, AC12; AD-1's allow-listed numeric surface: %s)",
						rel, pos.Line, sel.Sel.Name, allowedNumericSurface),
				})
			}
			return true
		})

		// runtime.Caller selector matching (AC2, V1, Story 2.1): same
		// alias-resolved shape as the math selector above, but a
		// simpler ban — ANY reference to runtime.Caller (call or not)
		// is a finding; no allow-list, because there is no legitimate
		// render-path use of it (D-000.9: measured zero occurrences
		// repo-wide before this story, so this is a genuine, not
		// vacuous, new check).
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			pkgPath, known := aliases[pkgIdent.Name]
			if !known || pkgPath != "runtime" || sel.Sel.Name != "Caller" {
				return true
			}
			pos := fset.Position(sel.Pos())
			findings = append(findings, Finding{
				Path: rel, Rule: RuleRuntimeCaller, Line: pos.Line,
				Message: fmt.Sprintf("%s:%d: forbidden runtime.Caller reference — render-path code must never inspect its own call stack or depend on filesystem-shaped introspection (AD-1, AC2)", rel, pos.Line),
			})
			return true
		})
		return nil
	})
	return findings, stats, err
}
