package expr

import (
	"fmt"
	"strings"
)

// Eval walks e against resolver, computing a Value (AC12-AC18).
// elementID names the binding site in every located error this
// produces, matching AD-14's convention throughout the rest of the
// codebase.
//
// Eval assumes e already passed Check (arity, unknown-function-name
// and literal-argument-kind are NOT re-derived here — a defensive
// unknown-function/arity mismatch is still handled below, cheaply, in
// case a caller ever reaches Eval without calling Check first, but the
// authoritative, located version of those errors is Check's).
func Eval(e Expr, resolver Resolver, elementID string) (Value, error) {
	switch n := e.(type) {
	case *PathExpr:
		return resolver.Resolve(n.Segments)
	case *StringLit:
		return Value{Kind: KindString, Str: n.Value}, nil
	case *NumberLit:
		d, err := NewDecimal(n.Literal)
		if err != nil {
			return Value{}, fmt.Errorf("expr: element %s: invalid number literal %q: %w", elementID, n.Literal, err)
		}
		return Value{Kind: KindNumber, Num: d}, nil
	case *CallExpr:
		return evalCall(n, resolver, elementID)
	default:
		return Value{}, fmt.Errorf("expr: element %s: internal: unrecognised expression node %T", elementID, e)
	}
}

func evalCall(call *CallExpr, resolver Resolver, elementID string) (Value, error) {
	entry, ok := lookupFunc(call.Name)
	if !ok {
		return Value{}, fmt.Errorf(
			"expr: element %s: unknown function %q — the eight legal names are %s: %s",
			elementID, call.Name, strings.Join(LegalFunctionNames(), ", "), call.Raw,
		)
	}
	if len(call.Args) != entry.arity {
		return Value{}, fmt.Errorf(
			"expr: element %s: %s() takes %d argument(s), got %d: %s",
			elementID, entry.name, entry.arity, len(call.Args), call.Raw,
		)
	}

	if !entry.implemented {
		// AC15/AC17/AC18: a registered-but-unimplemented function is a
		// LOCATED error, never a plausible value — this is the guard
		// against F6's hazard: the "sum" entry above declares its
		// Decimal-typed signature (AC9, table.go) but is never wired
		// to SumDecimals here. Neither sum's nor any of the other four
		// unimplemented entries' arguments are evaluated at all before
		// this error is returned, so a mistyped or absent path buried
		// in an unimplemented call's argument is not additionally
		// reported (that would be Story 3.3/3.4's own evaluation, once
		// implemented).
		return Value{}, fmt.Errorf(
			"expr: element %s: function %q is not yet implemented (coming in Story %s): %s",
			elementID, entry.name, entry.owningStory, call.Raw,
		)
	}

	switch entry.name {
	case "upper", "lower":
		return evalUpperLower(entry.name, call, resolver, elementID)
	case "if":
		return evalIf(call, resolver, elementID)
	default:
		// Unreachable given functionTable's own entries (table.go):
		// every implemented entry is handled above. Kept as a located
		// error, not a panic, per AC20's "plain Go errors" discipline
		// and AD-14's "never a panic" (AC11).
		return Value{}, fmt.Errorf("expr: element %s: internal: %q is marked implemented but has no evaluator", elementID, entry.name)
	}
}

// evalUpperLower is AC12: upper()/lower() evaluate per Go's
// strings.ToUpper/ToLower. A non-string operand — including a value
// resolved from data of the wrong kind, a number literal, or a null —
// is a located error, never a coerced stringification (AD-14's
// wrong-kind case, never a coercion, AD-14 verbatim).
func evalUpperLower(name string, call *CallExpr, resolver Resolver, elementID string) (Value, error) {
	v, err := Eval(call.Args[0], resolver, elementID)
	if err != nil {
		return Value{}, err
	}
	if v.Kind != KindString {
		return Value{}, fmt.Errorf(
			"expr: element %s: %s() operand must be a string, got %s (never coerced): %s",
			elementID, name, v.Kind, call.Raw,
		)
	}
	s := v.Str
	if name == "upper" {
		s = strings.ToUpper(s)
	} else {
		s = strings.ToLower(s)
	}
	return Value{Kind: KindString, Str: s}, nil
}

// evalIf is AC13/AC14 and the owner's ruling on if(null, …): if(cond,
// then, else), arity exactly 3 (table.go). cond must resolve to a
// JSON boolean — NO truthiness in any form (AD-14: a wrong-kind value
// is an Error, never a coercion). Two of AD-14's three presence cases
// diverge deliberately, and the divergence IS the point, tested as a
// pair so it cannot drift apart silently:
//
//   - An ABSENT path as cond is a LOCATED ERROR carrying the path —
//     this falls out of resolver.Resolve's own contract (ast.go) with
//     no special-casing here at all: Eval(call.Args[0], …) below
//     simply returns whatever error the resolver produced for an
//     absent path, unchanged.
//   - An EXPLICIT JSON null as cond takes the ELSE branch, SILENTLY —
//     no error, no diagnostic, no warning (OWNER DECISION, this
//     story). The trade-off (a reader cannot distinguish a hidden
//     section from one that was never there) was presented and chosen
//     deliberately, over the alternative of a Warning. This is the one
//     behaviour in the engine that produces no signal at all; it is
//     documented in folio-format.md and has its own dedicated,
//     findable test (TestIfNullConditionIsSilentlyFalse, eval_test.go)
//     for exactly that reason.
//
// AC14: only the SELECTED branch is evaluated — an unimplemented
// function called from the branch that was NOT selected must not
// surface at all. This is not a special case here either: the
// unselected call.Args[1]/[2] element is simply never passed to Eval.
func evalIf(call *CallExpr, resolver Resolver, elementID string) (Value, error) {
	condVal, err := Eval(call.Args[0], resolver, elementID)
	if err != nil {
		return Value{}, err
	}

	var branch Expr
	switch condVal.Kind {
	case KindBool:
		if condVal.Bool {
			branch = call.Args[1]
		} else {
			branch = call.Args[2]
		}
	case KindNull:
		branch = call.Args[2] // OWNER RULING: silent false, no diagnostic.
	default:
		return Value{}, fmt.Errorf(
			"expr: element %s: if() condition must be a boolean, got %s (no truthiness — AD-14): %s",
			elementID, condVal.Kind, call.Raw,
		)
	}
	return Eval(branch, resolver, elementID)
}
