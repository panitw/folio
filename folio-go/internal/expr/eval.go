package expr

import (
	"fmt"
	"strings"
)

// Eval walks e against resolver, computing a Value (AC12-AC18), and
// alongside it every Caveat (Story 3.3, DECISION-5) the walk produced
// — non-error conditions the render survives (avg()-on-empty, R9).
// caveats is nil whenever none occurred; it is never a non-nil empty
// slice (D-2.8.6's "empty is nil, one representation", applied here so
// a caller can compare without special-casing the length-0 case).
// elementID names the binding site in every located error this
// produces, matching AD-14's convention throughout the rest of the
// codebase.
//
// Eval assumes e already passed Check (arity, unknown-function-name
// and literal-argument-kind are NOT re-derived here — a defensive
// unknown-function/arity mismatch is still handled below, cheaply, in
// case a caller ever reaches Eval without calling Check first, but the
// authoritative, located version of those errors is Check's).
func Eval(e Expr, resolver Resolver, elementID string) (Value, []Caveat, error) {
	switch n := e.(type) {
	case *PathExpr:
		v, err := resolver.Resolve(n.Segments)
		return v, nil, err
	case *StringLit:
		return Value{Kind: KindString, Str: n.Value}, nil, nil
	case *NumberLit:
		d, err := NewDecimal(n.Literal)
		if err != nil {
			return Value{}, nil, fmt.Errorf("expr: element %s: invalid number literal %q: %w", elementID, n.Literal, err)
		}
		return Value{Kind: KindNumber, Num: d}, nil, nil
	case *CallExpr:
		return evalCall(n, resolver, elementID)
	default:
		return Value{}, nil, fmt.Errorf("expr: element %s: internal: unrecognised expression node %T", elementID, e)
	}
}

func evalCall(call *CallExpr, resolver Resolver, elementID string) (Value, []Caveat, error) {
	entry, ok := lookupFunc(call.Name)
	if !ok {
		return Value{}, nil, fmt.Errorf(
			"expr: element %s: unknown function %q — the eight legal names are %s: %s",
			elementID, call.Name, strings.Join(LegalFunctionNames(), ", "), call.Raw,
		)
	}
	if len(call.Args) != entry.arity {
		return Value{}, nil, fmt.Errorf(
			"expr: element %s: %s() takes %d argument(s), got %d: %s",
			elementID, entry.name, entry.arity, len(call.Args), call.Raw,
		)
	}

	if !entry.implemented {
		// AC15/AC17/AC18: a registered-but-unimplemented function is a
		// LOCATED error, never a plausible value — this is the guard
		// against F6's hazard. Neither its arguments nor any other
		// unimplemented entry's are evaluated at all before this error
		// is returned.
		return Value{}, nil, fmt.Errorf(
			"expr: element %s: function %q is not yet implemented (coming in Story %s): %s",
			elementID, entry.name, entry.owningStory, call.Raw,
		)
	}

	switch entry.name {
	case "upper", "lower":
		v, err := evalUpperLower(entry.name, call, resolver, elementID)
		return v, nil, err
	case "if":
		return evalIf(call, resolver, elementID)
	case "sum":
		return evalSum(call, resolver, elementID)
	case "count":
		v, err := evalCount(call, resolver, elementID)
		return v, nil, err
	case "avg":
		return evalAvg(call, resolver, elementID)
	default:
		// Unreachable given functionTable's own entries (table.go):
		// every implemented entry is handled above. Kept as a located
		// error, not a panic, per AC20's "plain Go errors" discipline
		// and AD-14's "never a panic" (AC11).
		return Value{}, nil, fmt.Errorf("expr: element %s: internal: %q is marked implemented but has no evaluator", elementID, entry.name)
	}
}

// evalUpperLower is AC12: upper()/lower() evaluate per Go's
// strings.ToUpper/ToLower. A non-string operand — including a value
// resolved from data of the wrong kind, a number literal, or a null —
// is a located error, never a coerced stringification (AD-14's
// wrong-kind case, never a coercion, AD-14 verbatim).
func evalUpperLower(name string, call *CallExpr, resolver Resolver, elementID string) (Value, error) {
	v, _, err := Eval(call.Args[0], resolver, elementID)
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
//
// A Caveat from cond's OWN evaluation (e.g. avg() used, unusually, as
// a condition and landing on the empty-average caveat before erroring
// on its non-boolean kind) still propagates — evalIf never discards a
// caveat it collected on the way to a result, selected branch or not.
func evalIf(call *CallExpr, resolver Resolver, elementID string) (Value, []Caveat, error) {
	condVal, condCaveats, err := Eval(call.Args[0], resolver, elementID)
	if err != nil {
		return Value{}, nil, err
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
		return Value{}, nil, fmt.Errorf(
			"expr: element %s: if() condition must be a boolean, got %s (no truthiness — AD-14): %s",
			elementID, condVal.Kind, call.Raw,
		)
	}
	branchVal, branchCaveats, err := Eval(branch, resolver, elementID)
	if err != nil {
		return Value{}, nil, err
	}
	return branchVal, appendCaveats(condCaveats, branchCaveats), nil
}

// appendCaveats concatenates a and b, preserving D-2.8.6's "empty is
// nil, one representation": nil in, nil out, never a non-nil empty
// slice manufactured along the way.
func appendCaveats(a, b []Caveat) []Caveat {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	out := make([]Caveat, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}
