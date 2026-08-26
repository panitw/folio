package expr

import (
	"fmt"
	"strings"
)

// Check walks e and reports every statically decidable defect: an
// unknown function name (AC11), a wrong arity (AC10), or a literal
// argument of the wrong kind (Decision 3's "literal-argument kind"
// half — AC10's own examples, sum("hello") and formatNumber(x, 123)).
// It never resolves a path against data and never asks what a PATH
// argument would resolve to (Decision 3, FLAG-2: that half is
// explicitly NOT this story's obligation, and is owed at evaluation by
// each function's own implementing story).
//
// Check is meant to run once per parsed expression, at LOAD time (R3:
// "syntax and arity at load; execution at evaluation") — folio.
// ParseTemplate calls Parse then Check over every "{{ }}" binding in
// the document (excluding AD-4's reserved page/pages tokens, which
// never reach a parser at all). It is also safe, and cheap, to call
// again before evaluation as a defence-in-depth measure.
func Check(e Expr) error {
	switch v := e.(type) {
	case *CallExpr:
		return checkCall(v)
	case *NumberLit:
		return checkNumberLit(v)
	default:
		// A bare path or string literal — at the top level, or nested
		// one level deeper, reached via the recursion below — has
		// nothing left to check: a path is checked at evaluation
		// (AD-14, against real data), and a string literal is
		// trivially well-formed once Parse accepted it.
		return nil
	}
}

// checkNumberLit is R3's own numeric-literal half (QA Finding 7,
// Major): a number literal's bounds (maxDecimalCoefficientDigits,
// maxDecimalExponentMagnitude — decimal.go) are decidable with NO
// data at all, exactly like arity, so R3 ("syntax and arity at load;
// execution at evaluation") puts them at Check, not evaluation. Before
// this fix a template carrying an impossible literal —
// "12345678901234567890123456789", or "1e999999999999999999999" —
// loaded clean and died mid-render: the inverse of what F3 forced for
// unimplemented FUNCTIONS, which must reach evaluation because the
// canonical golden needs them to; a literal's own shape is not that
// case, and NewDecimal is the one place that already knows how to
// reject it, so Check calls the exact function eval.go calls, never a
// re-derived rule.
func checkNumberLit(n *NumberLit) error {
	if _, err := NewDecimal(n.Literal); err != nil {
		return fmt.Errorf("invalid number literal %s: %s", n.Raw, err)
	}
	return nil
}

func checkCall(call *CallExpr) error {
	entry, ok := lookupFunc(call.Name)
	if !ok {
		return fmt.Errorf(
			"unknown function %q — the eight legal names are %s: %s",
			call.Name, strings.Join(LegalFunctionNames(), ", "), call.Raw,
		)
	}
	if len(call.Args) != entry.arity {
		return fmt.Errorf(
			"%s() takes %d argument(s), got %d: %s",
			entry.name, entry.arity, len(call.Args), call.Raw,
		)
	}
	for i, arg := range call.Args {
		if err := checkArgKind(entry, i, arg, call.Raw); err != nil {
			return err
		}
		// Recurse: a nested call (AC3's nesting case) is checked with
		// the same rigour as a top-level one.
		if err := Check(arg); err != nil {
			return err
		}
	}

	// AC10/F3: a pattern literal's own grammar is decidable with NO
	// data at all, exactly like arity and a number literal's bounds
	// (checkNumberLit, *Do not re-open* item 9) — so it belongs here,
	// at Check (load time), not at Eval. checkArgKind above has
	// already confirmed call.Args[1] is a *StringLit for both
	// functions (argStringLiteral); this is the SAME grammar
	// parseDatePattern/validateNumberPattern apply at evaluation —
	// one implementation, never two that could drift.
	switch call.Name {
	case "formatDate":
		lit, ok := call.Args[1].(*StringLit)
		if !ok {
			return fmt.Errorf("expr: %s(): internal: pattern argument was not a string literal after checkArgKind: %s", call.Name, call.Raw)
		}
		if _, err := parseDatePattern(lit.Value); err != nil {
			return err
		}
	case "formatNumber":
		lit, ok := call.Args[1].(*StringLit)
		if !ok {
			return fmt.Errorf("expr: %s(): internal: pattern argument was not a string literal after checkArgKind: %s", call.Name, call.Raw)
		}
		if _, err := validateNumberPattern(lit.Value); err != nil {
			return err
		}
	}
	return nil
}

func checkArgKind(entry funcEntry, index int, arg Expr, callRaw string) error {
	kind := entry.args[index]
	switch kind {
	case argAny:
		return nil
	case argNotLiteral:
		switch arg.(type) {
		case *StringLit, *NumberLit:
			return fmt.Errorf(
				"%s(): argument %d must not be a literal (expected a data path or a nested call), got %s: %s",
				entry.name, index+1, arg.Text(), callRaw,
			)
		}
		return nil
	case argStringLiteral:
		if _, ok := arg.(*StringLit); !ok {
			return fmt.Errorf(
				"%s(): argument %d must be a string literal pattern, got %s: %s",
				entry.name, index+1, arg.Text(), callRaw,
			)
		}
		return nil
	default:
		return nil
	}
}
