package expr

// This file is FR18/AD-9/C1's closed table: the single package-level
// literal, in one file, of the eight and ONLY eight functions the
// expression language will ever have (AC5). Adding a ninth requires
// editing both this literal and AC7's expected-set guard, in the same
// diff, or CI is red (AC8, C1) — that is the whole point of counting
// them here, in one place, rather than letting each function register
// itself.
//
// AC6: no exported function anywhere in internal/expr adds to,
// mutates, or replaces this table — closed at COMPILE time, not by
// convention. functionTable is unexported, is never returned by
// reference from an exported function, and there is no exported
// "Register" of any kind (internal/expr_arch_test.go asserts this by
// AST, over internal/expr's own exported declarations).

// argKind is a per-argument constraint Check (check.go) can decide
// WITHOUT any data — Decision 3's "literal-argument kind" half, never
// its "path-argument value kind" half (that is explicitly NOT this
// story's obligation; it is owed at evaluation by whichever story
// implements the function — see FuncEntry.OwningStory).
type argKind int

const (
	// argAny places no static constraint on this argument at all.
	argAny argKind = iota
	// argNotLiteral forbids a bare string/number literal in this
	// argument position: sum/count/avg's single operand must name a
	// collection, which no literal production in the grammar can be
	// (AC10's `sum("hello")` example), and if()'s condition must be a
	// JSON boolean, which — since this grammar has no boolean literal
	// — no literal expression can ever be (F11/FLAG-3: absent and
	// explicit null are AD-14's OWN distinct cases, decided at
	// evaluation against real data, never here).
	argNotLiteral
	// argStringLiteral requires this argument to be, syntactically, a
	// string literal — never a path, a call, or a number literal.
	// formatDate/formatNumber's pattern argument (D-1.4.1) is always a
	// literal pattern, never data-dependent (AC10's
	// `formatNumber(x, 123)` example: the pattern is present but is
	// the WRONG kind of literal).
	argStringLiteral
)

// returnKind marks a table entry's declared RETURN type (AC9: "each of
// the three aggregation entries declares a Decimal-typed signature …
// declaring the table honestly requires referencing Decimal — a
// stringly-typed table is a defect on its own merits"). Each variant
// embeds the real Go type it names, rather than a string label, so the
// table's own declarations are checkable by the Go compiler, not by
// convention: sum/count/avg's entries below literally construct a
// returnDecimal{}, which only type-checks because Decimal (decimal.go)
// exists and is spelled correctly.
type returnKind interface{ isReturnKind() }

type returnDecimal struct{ zero Decimal } // sum, count, avg
type returnString struct{}                // formatDate, formatNumber, upper, lower
type returnAny struct{}                   // if — whichever branch is selected

func (returnDecimal) isReturnKind() {}
func (returnString) isReturnKind()  {}
func (returnAny) isReturnKind()     {}

// FuncEntry is one row of the closed eight-entry table.
type funcEntry struct {
	name  string
	arity int
	args  []argKind // len(args) == arity
	ret   returnKind

	// implemented reports whether Eval actually computes a result for
	// this function (AC12-AC14: upper/lower/if) or whether every call
	// is a located "not yet implemented" error (AC15-AC18: the other
	// five).
	implemented bool

	// owningStory names the story that implements this function, used
	// only in the unimplemented-function error message (AC15).
	// Meaningless (and unused) when implemented is true.
	owningStory string
}

// functionTable is FR18's closed eight, keyed by name (AC5). Ordering
// here is source order, not significant to any guard.
var functionTable = [8]funcEntry{
	{name: "sum", arity: 1, args: []argKind{argNotLiteral}, ret: returnDecimal{}, implemented: true},
	{name: "count", arity: 1, args: []argKind{argNotLiteral}, ret: returnDecimal{}, implemented: true},
	{name: "avg", arity: 1, args: []argKind{argNotLiteral}, ret: returnDecimal{}, implemented: true},
	{name: "formatDate", arity: 2, args: []argKind{argAny, argStringLiteral}, ret: returnString{}, implemented: false, owningStory: "3.4"},
	{name: "formatNumber", arity: 2, args: []argKind{argAny, argStringLiteral}, ret: returnString{}, implemented: false, owningStory: "3.4"},
	{name: "upper", arity: 1, args: []argKind{argAny}, ret: returnString{}, implemented: true},
	{name: "lower", arity: 1, args: []argKind{argAny}, ret: returnString{}, implemented: true},
	{name: "if", arity: 3, args: []argKind{argNotLiteral, argAny, argAny}, ret: returnAny{}, implemented: true},
}

// LegalFunctionNames returns the eight legal names, in table order —
// used only to compose AC11's "the eight legal names" error text, so
// that text can never drift from the table it names (AD-13).
func LegalFunctionNames() []string {
	names := make([]string, 0, len(functionTable))
	for _, e := range functionTable {
		names = append(names, e.name)
	}
	return names
}

// lookupFunc finds name's table entry.
func lookupFunc(name string) (funcEntry, bool) {
	for _, e := range functionTable {
		if e.name == name {
			return e, true
		}
	}
	return funcEntry{}, false
}
