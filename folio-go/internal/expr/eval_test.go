package expr

import (
	"fmt"
	"strings"
	"testing"
)

// mapResolver is a minimal, in-memory Resolver for these tests: a
// missing key is Absent (a located error, mirroring AD-14), an
// explicit Value{Kind: KindNull} entry is Null (a legal, non-error
// value), anything else is Present.
type mapResolver map[string]Value

func (r mapResolver) Resolve(path []string) (Value, error) {
	key := strings.Join(path, ".")
	v, ok := r[key]
	if !ok {
		return Value{}, fmt.Errorf("test: path %q is absent from the test resolver", key)
	}
	return v, nil
}

// CollectionLength/ProjectCollection: mapResolver is a flat map keyed
// by dotted path — it has no notion of a collection at all, so both
// methods simply report "absent", the same outcome Resolve gives any
// other unmapped path. Tests that need real collection behaviour use
// internal/bind's exprResolver directly (bind_test.go-adjacent tests)
// or sliceResolver (below), never mapResolver.
func (r mapResolver) CollectionLength(path []string) (int, error) {
	return 0, fmt.Errorf("test: collection path %q is absent from the test resolver", strings.Join(path, "."))
}

func (r mapResolver) ProjectCollection(path []string) ([]Value, error) {
	return nil, fmt.Errorf("test: collection path %q is absent from the test resolver", strings.Join(path, "."))
}

func mustEval(t *testing.T, src string, resolver Resolver) Value {
	t.Helper()
	e, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	if err := Check(e); err != nil {
		t.Fatalf("Check(%q): %v", src, err)
	}
	v, _, err := Eval(e, resolver, "e1")
	if err != nil {
		t.Fatalf("Eval(%q): unexpected error: %v", src, err)
	}
	return v
}

// --- AC12: upper()/lower() ---

func TestUpperLowerBasic(t *testing.T) {
	r := mapResolver{"x": {Kind: KindString, Str: "Hello"}}
	if v := mustEval(t, "upper(x)", r); v.Str != "HELLO" {
		t.Errorf("upper: got %q", v.Str)
	}
	if v := mustEval(t, "lower(x)", r); v.Str != "hello" {
		t.Errorf("lower: got %q", v.Str)
	}
}

func TestUpperLowerEmptyString(t *testing.T) {
	r := mapResolver{"x": {Kind: KindString, Str: ""}}
	if v := mustEval(t, "upper(x)", r); v.Str != "" {
		t.Errorf("upper(\"\") = %q, want empty", v.Str)
	}
}

// TestUpperLowerThaiIsNoOp: the golden report is a Thai bank statement
// — Thai has no case distinction, so case mapping must leave the bytes
// unchanged.
func TestUpperLowerThaiIsNoOp(t *testing.T) {
	const thai = "ศรีสุข"
	r := mapResolver{"x": {Kind: KindString, Str: thai}}
	if v := mustEval(t, "upper(x)", r); v.Str != thai {
		t.Errorf("upper(Thai) must be a no-op, got %q", v.Str)
	}
	if v := mustEval(t, "lower(x)", r); v.Str != thai {
		t.Errorf("lower(Thai) must be a no-op, got %q", v.Str)
	}
}

func TestUpperLowerCJK(t *testing.T) {
	const cjk = "日本語"
	r := mapResolver{"x": {Kind: KindString, Str: cjk}}
	if v := mustEval(t, "upper(x)", r); v.Str != cjk {
		t.Errorf("upper(CJK) must be a no-op, got %q", v.Str)
	}
}

// TestUpperLowerCombiningMarks: a string with a combining mark must
// round-trip byte-for-byte through case mapping (strings.ToUpper on an
// ASCII base plus a combining mark leaves the mark alone).
func TestUpperLowerCombiningMarks(t *testing.T) {
	const withMark = "élan" // "e" + COMBINING ACUTE ACCENT + "lan"
	r := mapResolver{"x": {Kind: KindString, Str: withMark}}
	got := mustEval(t, "upper(x)", r)
	if !strings.Contains(got.Str, "́") {
		t.Errorf("upper() must preserve the combining mark, got %q", got.Str)
	}
}

func TestUpperLowerNonStringOperandIsLocatedError(t *testing.T) {
	r := mapResolver{"x": {Kind: KindNumber}}
	// QA Finding 15 (Minor): fail the parse explicitly, as mustEval
	// already does elsewhere — a discarded parse error means a
	// regressed Parse (returning a nil Expr) would fall through to
	// Eval's own "unrecognised expression node" error, which still
	// contains "e7", so this test would keep passing for the wrong
	// reason.
	e, perr := Parse("upper(x)")
	if perr != nil {
		t.Fatalf(`Parse("upper(x)"): unexpected syntax error: %v`, perr)
	}
	_, _, err := Eval(e, r, "e7")
	if err == nil {
		t.Fatal("expected a located error for a non-string operand")
	}
	if !strings.Contains(err.Error(), "e7") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	if !strings.Contains(err.Error(), "string") {
		t.Errorf("error must name the expected kind (string), got: %v", err)
	}
}

// --- AC13/AC14: if() ---

func TestIfTrueBranch(t *testing.T) {
	r := mapResolver{"cond": {Kind: KindBool, Bool: true}, "a": {Kind: KindString, Str: "A"}, "b": {Kind: KindString, Str: "B"}}
	if v := mustEval(t, "if(cond, a, b)", r); v.Str != "A" {
		t.Errorf("got %q, want A", v.Str)
	}
}

func TestIfFalseBranch(t *testing.T) {
	r := mapResolver{"cond": {Kind: KindBool, Bool: false}, "a": {Kind: KindString, Str: "A"}, "b": {Kind: KindString, Str: "B"}}
	if v := mustEval(t, "if(cond, a, b)", r); v.Str != "B" {
		t.Errorf("got %q, want B", v.Str)
	}
}

// TestIfAbsentConditionIsLocatedError: an ABSENT path as condition is
// a LOCATED ERROR carrying the path — never false. This is the
// deliberate contrast with TestIfNullConditionIsSilentlyFalse below;
// the pair is tested together so the distinction cannot drift.
func TestIfAbsentConditionIsLocatedError(t *testing.T) {
	r := mapResolver{"a": {Kind: KindString, Str: "A"}, "b": {Kind: KindString, Str: "B"}}
	// QA Finding 15 (Minor): fail the parse explicitly.
	e, perr := Parse("if(cond, a, b)")
	if perr != nil {
		t.Fatalf(`Parse("if(cond, a, b)"): unexpected syntax error: %v`, perr)
	}
	_, _, err := Eval(e, r, "e9")
	if err == nil {
		t.Fatal("an absent condition path must be a located Error, never false")
	}
	if !strings.Contains(err.Error(), "cond") {
		t.Errorf("error must carry the condition's path, got: %v", err)
	}
}

// TestIfNullConditionIsSilentlyFalse is THE dedicated, findable test
// for the owner's ruling: if(condition, then, else) with an explicit
// JSON null condition takes the ELSE branch, SILENTLY — no error, no
// diagnostic, no warning. This is the one behaviour in the engine that
// produces no signal in its output at all (a reader cannot distinguish
// a hidden section from one that was never there); this test and
// folio-format.md's stated-behaviour entry are the only two places
// this fact exists, per the owner's own implementation obligation.
//
// Paired deliberately with TestIfAbsentConditionIsLocatedError above:
// explicit null is silently false, an absent path is a located error
// — never the same outcome, and this pair is what keeps that
// distinction from drifting apart under a future edit.
func TestIfNullConditionIsSilentlyFalse(t *testing.T) {
	r := mapResolver{
		"cond": {Kind: KindNull},
		"a":    {Kind: KindString, Str: "A"},
		"b":    {Kind: KindString, Str: "B"},
	}
	v := mustEval(t, "if(cond, a, b)", r)
	if v.Str != "B" {
		t.Fatalf("explicit null condition must silently take the else branch, got %q", v.Str)
	}
}

func TestIfEmptyStringConditionIsLocatedErrorNoTruthiness(t *testing.T) {
	r := mapResolver{"cond": {Kind: KindString, Str: ""}, "a": {Kind: KindString, Str: "A"}, "b": {Kind: KindString, Str: "B"}}
	// QA Finding 15 (Minor): fail the parse explicitly. Before this
	// fix, a regressed Parse (rejecting "if(cond, a, b)" and returning
	// a nil Expr) would still make this test pass: Eval(nil, …) falls
	// to its own "unrecognised expression node" default-case error,
	// which is still non-nil — the assertion below never inspected the
	// message, so it could not tell "no truthiness, as intended" from
	// "the parser broke". Now it asserts the eval error names the
	// condition's actual wrong KIND, which only the intended
	// no-truthiness arm can produce.
	e, perr := Parse("if(cond, a, b)")
	if perr != nil {
		t.Fatalf(`Parse("if(cond, a, b)"): unexpected syntax error: %v`, perr)
	}
	_, _, err := Eval(e, r, "e1")
	if err == nil {
		t.Fatal("an empty-string condition must be a located error — no truthiness")
	}
	if !strings.Contains(err.Error(), "string") {
		t.Errorf(`error must name the condition's actual kind ("string"), got: %v`, err)
	}
}

func TestIfZeroConditionIsLocatedErrorNoTruthiness(t *testing.T) {
	r := mapResolver{"cond": {Kind: KindNumber, Num: Decimal{Coefficient: 0, Exponent: 0}}, "a": {Kind: KindString, Str: "A"}, "b": {Kind: KindString, Str: "B"}}
	// QA Finding 15 (Minor): see the sibling test above — same shape.
	e, perr := Parse("if(cond, a, b)")
	if perr != nil {
		t.Fatalf(`Parse("if(cond, a, b)"): unexpected syntax error: %v`, perr)
	}
	_, _, err := Eval(e, r, "e1")
	if err == nil {
		t.Fatal("a zero condition must be a located error — 0 is not false (no truthiness)")
	}
	if !strings.Contains(err.Error(), "number") {
		t.Errorf(`error must name the condition's actual kind ("number"), got: %v`, err)
	}
}

// TestIfShortCircuitsUnselectedBranch is AC14: the unselected branch
// calls an unimplemented function; the expression must still succeed
// because that branch is never evaluated at all.
func TestIfShortCircuitsUnselectedBranch(t *testing.T) {
	r := mapResolver{"cond": {Kind: KindBool, Bool: true}}
	v := mustEval(t, `if(cond, "a", sum(t.x))`, r)
	if v.Str != "a" {
		t.Fatalf("got %q, want \"a\" — a non-short-circuiting implementation would have errored on sum()", v.Str)
	}
}

// TestIfShortCircuitsAbsentPathInUnselectedBranch is QA Finding 16(a)
// (Minor): the DOCUMENTED motivating example (folio-format.md's
// "### Expressions" section) is {{if(hasDiscount, discount.amount,
// "N/A")}} on a row where discount is absent entirely.
// TestIfShortCircuitsUnselectedBranch above exercises the mechanism
// with an unimplemented function in the untaken branch; this is the
// actual documented shape, with an ABSENT PATH in the untaken branch
// instead — a case nothing exercised before this fix.
func TestIfShortCircuitsAbsentPathInUnselectedBranch(t *testing.T) {
	r := mapResolver{"hasDiscount": {Kind: KindBool, Bool: false}}
	v := mustEval(t, `if(hasDiscount, discount.amount, "N/A")`, r)
	if v.Str != "N/A" {
		t.Fatalf(`got %q, want "N/A" — a non-short-circuiting implementation would have errored resolving the absent "discount.amount"`, v.Str)
	}
}

// --- AC15/AC18/AC30: the two functions this story leaves unimplemented ---
//
// AC30 (D-000.59's shape): sum/count/avg are DROPPED from this table —
// they are Story 3.3's own positive assertions now (below, and
// aggregate_test.go) — in the SAME commit that adds the three of them
// to TestImplementedEntriesMatchEvalCallSwitch's derived set
// (table_derivational_test.go). formatDate/formatNumber remain, still
// proving the located-error arm Story 3.4 will one day retire in turn.

func TestUnimplementedFunctionsAreLocatedErrors(t *testing.T) {
	r := mapResolver{"x": {Kind: KindString, Str: "irrelevant"}}
	cases := []struct {
		src   string
		name  string
		story string
	}{
		{`formatDate(x, "yyyy-MM-dd")`, "formatDate", "3.4"},
		{`formatNumber(x, "#,##0.00")`, "formatNumber", "3.4"},
	}
	seen := 0
	for _, c := range cases {
		seen++
		e, err := Parse(c.src)
		if err != nil {
			t.Fatalf("Parse(%q): %v", c.src, err)
		}
		_, _, err = Eval(e, r, "e5")
		if err == nil {
			t.Fatalf("%s: expected a located error, got none", c.src)
		}
		if !strings.Contains(err.Error(), "e5") {
			t.Errorf("%s: error must name the element id, got: %v", c.src, err)
		}
		if !strings.Contains(err.Error(), c.story) {
			t.Errorf("%s: error must name the owning story %s, got: %v", c.src, c.story, err)
		}
		// QA Finding 14 (Minor): AC15 also requires the function name
		// and the offending expression text — the assertions above
		// left both unpinned, so dropping call.Raw or the function
		// name from the message would have been a silent, green
		// change.
		if !strings.Contains(err.Error(), c.name) {
			t.Errorf("%s: error must name the function %q, got: %v", c.src, c.name, err)
		}
		if !strings.Contains(err.Error(), c.src) {
			t.Errorf("%s: error must carry the offending expression text verbatim, got: %v", c.src, err)
		}
	}
	if seen != 2 {
		t.Fatalf("AC30: expected exactly 2 assertions (formatDate, formatNumber), ran %d", seen)
	}
}

// TestSumOverEmptyOperandIsNeverASilentZero is AC17: THE
// VENDOR-DEFAULT RED-PROOF (F6, D-000.25). Evaluating sum() over an
// ABSENT collection must return an error, and specifically must never
// return a Decimal of value zero — the hazard being guarded against is
// SumDecimals(nil) == Decimal{0,0} (D-3.1a.2, reduce.go): a
// ProjectCollection that silently reported "no elements" for an absent
// collection (rather than the located Error R8 requires) would feed
// SumDecimals nothing and get that exact identity back, a plausible-
// looking zero total on a bank statement whose collection was never
// there at all.
func TestSumOverEmptyOperandIsNeverASilentZero(t *testing.T) {
	r := mapResolver{"transactions": {Kind: KindString, Str: "irrelevant — never reached"}}
	e, err := Parse("sum(transactions.amount)")
	if err != nil {
		t.Fatalf("unexpected syntax error: %v", err)
	}
	v, _, err := Eval(e, r, "e17")
	// QA Finding 8 (Minor): the value check must live on the SUCCESS
	// path. t.Fatal calls runtime.Goexit, so the old ordering (value
	// check unconditionally after the err==nil check) could only ever
	// be reached once err was already known non-nil — on which every
	// error path returns the zero Value{} (KindNull, never KindNumber)
	// anyway, so the clause was unreachable in both directions. Moving
	// it inside err == nil makes it the thing that actually
	// distinguishes "erroring, as required" from "returning the exact
	// SumDecimals(nil) hazard value" — AC17's "specifically" clause.
	if err == nil {
		if v.Kind == KindNumber && v.Num == (Decimal{Coefficient: 0, Exponent: 0}) {
			t.Fatal("AC17 RED-PROOF FAILED: sum() returned the exact SumDecimals(nil) identity {0,0} — the vendor-default hazard shipped")
		}
		t.Fatal("AC17: sum() must return an error, not a plausible value")
	}
}

// TestUnimplementedAndUnknownFunctionErrorsAreDistinguishable is AC18.
func TestUnimplementedAndUnknownFunctionErrorsAreDistinguishable(t *testing.T) {
	r := mapResolver{"x": {Kind: KindString, Str: "v"}}

	// QA Finding 15 (Minor): fail each parse explicitly.
	// AC30: sum() is implemented as of this story — formatDate() is
	// now the unimplemented example (owned by Story 3.4).
	e1, perr1 := Parse(`formatDate(x, "yyyy-MM-dd")`)
	if perr1 != nil {
		t.Fatalf(`Parse("formatDate(...)"): unexpected syntax error: %v`, perr1)
	}
	_, _, err1 := Eval(e1, r, "e1")

	e2, perr2 := Parse("frobnicate(x)")
	if perr2 != nil {
		t.Fatalf(`Parse("frobnicate(x)"): unexpected syntax error: %v`, perr2)
	}
	_, _, err2 := Eval(e2, r, "e1")

	if err1 == nil || err2 == nil {
		t.Fatalf("expected both to error, got err1=%v err2=%v", err1, err2)
	}
	if err1.Error() == err2.Error() {
		t.Fatal("AC18: the unimplemented-function error and the unknown-function error must not collapse to the same message")
	}
	if !strings.Contains(err1.Error(), "3.4") {
		t.Errorf("unimplemented error should name Story 3.4, got: %v", err1)
	}
	if strings.Contains(err2.Error(), "3.4") {
		t.Errorf("unknown-function error must not claim an owning story, got: %v", err2)
	}
}

// --- number/string literal evaluation, and path evaluation shape ---

func TestEvalStringLiteral(t *testing.T) {
	v := mustEval(t, `"hello"`, mapResolver{})
	if v.Kind != KindString || v.Str != "hello" {
		t.Errorf("got %#v", v)
	}
}

func TestEvalPathPropagatesResolverError(t *testing.T) {
	// QA Finding 15 (Minor): fail the parse explicitly.
	e, perr := Parse("missing.path")
	if perr != nil {
		t.Fatalf(`Parse("missing.path"): unexpected syntax error: %v`, perr)
	}
	_, _, err := Eval(e, mapResolver{}, "e1")
	if err == nil {
		t.Fatal("expected the resolver's own absent error to propagate")
	}
	if !strings.Contains(err.Error(), "missing.path") {
		t.Errorf("expected the resolver's own wording to propagate unchanged, got: %v", err)
	}
}
