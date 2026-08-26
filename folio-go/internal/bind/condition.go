// This file is Story 3.5's own obligation (R4): a bare-expression
// evaluator over the SAME resolution seam Resolve (text.go) already
// uses, for a caller that needs the raw expr.Value a condition resolved
// to — never coerced to a string — rather than a bound TEXT result.
//
// Every OTHER bind.* entry point (BindText, BindTextSpans, Resolve) is
// text-span shaped: it scans for "{{ }}" placeholders and coerces
// whatever each one resolves to into a string (text.go's Resolve,
// `case default: … not a string — text bindings are never coerced`).
// A visibility condition is a BARE expression with no "{{ }}" wrapping
// (folio_expr_validate.go's checkVisibleIfExpression exists separately
// for exactly this reason: routing it through the text path "would
// scan for {{ }} occurrences inside it and find none, passing silently
// no matter what the string said"), and its result must stay a JSON
// boolean/null all the way back to the caller (D-3.2.3) rather than
// being forced through the string coercion every text-shaped entry
// point applies.
package bind

import (
	"fmt"
	"strings"

	"github.com/panitw/folio/folio-go/internal/expr"
)

// EvaluateCondition parses and statically checks src — a bare
// expression, not a "{{ }}"-wrapped interpolation — then evaluates it
// against scope, returning the resolved expr.Value UNCOERCED, plus any
// Caveat the walk produced.
//
// This is a new function on bind, not a new method on expr.Resolver
// (D-3.3.1/S7: the seam is exactly Resolve + CollectionLength +
// ProjectCollection, closed and enforced by
// lint/internal/rules/resolvermethodset.go) — it reuses the SAME
// exprResolver adapter Resolve (text.go) already builds, so a
// condition and a text placeholder dispatch to data/params/row through
// one traversal, never two.
func EvaluateCondition(src string, scope Scope, fc expr.FormatContext, elementID string) (expr.Value, []expr.Caveat, error) {
	e, perr := parseAndCheck(src, elementID)
	if perr != nil {
		return expr.Value{}, nil, perr
	}
	// D-3.5.1's own tripwire, applied here too (Story 3.5 finisher
	// review, Finding 10 / Minor): "a third condition slot appearing
	// that does not route through the hoisted predicate." EvaluateCondition
	// IS a condition slot. Today the only caller is checkVisibleIfExpression
	// (folio_expr_validate.go, at LOAD) followed by render_visibility.go
	// (at RENDER, on a value that already passed the load check) — so
	// this can never fire through the public API. It is enforced here
	// as well, not only at load, so a FUTURE second caller (e.g. a
	// programmatically-built *Template, or a designer-side evaluator)
	// cannot silently bypass D-3.5.1 by reaching this function directly.
	if expr.IsLiteralExpr(e) {
		return expr.Value{}, nil, fmt.Errorf(
			"bind: element %s: %q must not be a literal (expected a data path or a call) — "+
				"the grammar has no boolean literal, so a bare literal can never be a boolean",
			elementID, strings.TrimSpace(src),
		)
	}
	resolver := exprResolver{scope: scope, elementID: elementID}
	return expr.Eval(e, resolver, fc, elementID)
}

// parseAndCheck is the parse + static-check half every bare-expression
// entry point in this file shares (D-000.9: one traversal, one error
// shape, never two spellings of the same two steps).
func parseAndCheck(src, elementID string) (expr.Expr, error) {
	e, perr := expr.Parse(src)
	if perr != nil {
		return nil, fmt.Errorf(
			"bind: element %s: %q is not a valid expression: %s",
			elementID, strings.TrimSpace(src), perr,
		)
	}
	if cerr := expr.Check(e); cerr != nil {
		return nil, fmt.Errorf("bind: element %s: %s", elementID, cerr)
	}
	return e, nil
}

// EvaluateValue is this file's second bare-expression entry point, and
// it is the general one this file's own header describes: "a caller that
// needs the raw expr.Value a[n expression] resolved to — never coerced
// to a string". It parses, statically checks and evaluates src against
// scope through the SAME exprResolver seam Resolve (text.go) and
// EvaluateCondition already use, and returns the resolved expr.Value
// uncoerced, plus any Caveat the walk produced.
//
// It deliberately does NOT carry EvaluateCondition's D-3.5.1
// literal tripwire: that tripwire is a property of a CONDITION slot
// ("the grammar has no boolean literal"), not of bare evaluation, and
// applying it here would make this function a second condition slot —
// exactly what that tripwire exists to prevent.
//
// Its one caller today is Story 4.5's table footer (folio's
// table_render.go, footerCellExprText): D-1.4.1 rules that a footer
// whose footerFormat is "absent and underived" renders UNFORMATTED, and
// the only way to render a Decimal at its own scale through the closed
// pattern grammar is to know that scale — which means evaluating the
// aggregate, as a NUMBER, before the display expression is synthesised.
// Routing that through this function rather than through a private
// evaluator is what keeps the footer's value on the one aggregate
// evaluation (Story 4.5's AC4/DW-7) instead of manufacturing a second.
func EvaluateValue(src string, scope Scope, fc expr.FormatContext, elementID string) (expr.Value, []expr.Caveat, error) {
	e, perr := parseAndCheck(src, elementID)
	if perr != nil {
		return expr.Value{}, nil, perr
	}
	resolver := exprResolver{scope: scope, elementID: elementID}
	return expr.Eval(e, resolver, fc, elementID)
}
