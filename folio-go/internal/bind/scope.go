package bind

// Scope carries every resolution root reachable while resolving one
// placeholder (AC0): DATA and PARAMS are always present; ROW is
// optional, set only while resolving inside a repeating region's row
// scope (AD-11, Story 3.1). BindText/BindTextSpans (text.go) construct
// a Scope with no row set — the existing, pre-3.1 behaviour, byte-
// identical (AC0/AC7). A future story that generates rows (4.2) calls
// WithRow to activate one.
//
// The row's ALIAS (the author's "as" spelling, or the literal "row"
// when the region omits it, AC2) is carried on the Scope as ordinary
// data used for dispatch and for error text — never as a lookupBound
// rootName. TestBindResolutionRootsAreClosed
// (resolution_roots_arch_test.go) asserts every rootName is a string
// literal; an author's alias is data-dependent and would trip that
// guard's non-literal Fatalf. D-3.1.1 records this: the root-CLASS
// name passed to lookupBound is always the literal "row".
type Scope struct {
	data, params Value

	rowSet   bool
	row      Value
	rowAlias string
}

// NewScope builds a Scope with no row active — the "wrapper" scope
// AC0 requires BindText/BindTextSpans to construct.
func NewScope(data, params Value) Scope {
	return Scope{data: data, params: params}
}

// WithRow returns a copy of s with a row scope active, addressed by
// alias (the region's declared "as", or "row" when omitted — callers
// apply that default, not this function). Selection inside the
// resolver is by first-path-segment equality against alias, evaluated
// after params and before the data root (D-3.1.1).
func (s Scope) WithRow(row Value, alias string) Scope {
	s.rowSet = true
	s.row = row
	s.rowAlias = alias
	return s
}
