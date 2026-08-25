package expr

// Expr is one node of a parsed expression tree (AC1, AC3): a bare
// dotted path, a function call, a string literal, or a number literal.
// It carries no evaluation logic of its own — Eval (eval.go) walks the
// tree and Check (check.go) validates it statically; a node's only
// job is to remember what the author wrote and where.
type Expr interface {
	// Text returns the exact source text this node was parsed from
	// (AC19's "the offending expression text, verbatim, as the author
	// wrote it"; AC15's "the offending expression text").
	Text() string

	exprNode()
}

// PathExpr is a bare dotted path (AC3): "ident ( '.' ident )*",
// preserving 1.6's grammar verbatim (D-1.6.5). Segments is the path
// split on ".", in order.
type PathExpr struct {
	Segments []string
	Raw      string
}

func (p *PathExpr) Text() string { return p.Raw }
func (*PathExpr) exprNode()      {}

// StringLit is a double-quoted string literal (AC3). Value is the
// literal's content with the surrounding quotes removed; Raw is the
// exact source text including the quotes.
type StringLit struct {
	Value string
	Raw   string
}

func (s *StringLit) Text() string { return s.Raw }
func (*StringLit) exprNode()      {}

// NumberLit is a JSON number literal (AC3): sign, digits, optional
// fraction, optional exponent — the same shape
// internal/template.SplitJSONNumber accepts, since NewDecimal is what
// eventually consumes it (eval.go). Literal is the literal exactly as
// written; Raw is identical to Literal (a number literal is never
// wrapped in anything else).
type NumberLit struct {
	Literal string
	Raw     string
}

func (n *NumberLit) Text() string { return n.Raw }
func (*NumberLit) exprNode()      {}

// CallExpr is a function call over comma-separated arguments (AC3),
// nesting to any depth (AC3's minimum is one level;
// formatNumber(sum(t.amount), "#,##0.00") parses). Name is the bare
// function identifier as written — Check (check.go) is what resolves
// it against the closed eight-entry table (table.go); Parse itself
// has no notion of which names are legal, so the grammar and the
// closed set stay two separably-testable properties (AC1 vs AC5-AC11).
type CallExpr struct {
	Name string
	Args []Expr
	Raw  string
}

func (c *CallExpr) Text() string { return c.Raw }
func (*CallExpr) exprNode()      {}

// Kind is the discriminant of a resolved Value (evaluation-time
// result), and separately of a Presence-carrying lookup outcome.
type Kind int

const (
	KindNull Kind = iota
	KindString
	KindNumber
	KindBool
)

func (k Kind) String() string {
	switch k {
	case KindNull:
		return "null"
	case KindString:
		return "string"
	case KindNumber:
		return "number"
	case KindBool:
		return "bool"
	default:
		return "unknown"
	}
}

// Value is an expression's evaluated result: a scalar, deliberately —
// this story implements no function that produces or consumes a
// collection value (sum/count/avg are registered but unimplemented,
// AC15). AD-14's null case is a Kind, not a separate sentinel: an
// explicit JSON null is a legal Value (KindNull), never itself an
// evaluation error — only an ABSENT path is (see Resolver).
type Value struct {
	Kind Kind
	Str  string
	Num  Decimal
	Bool bool
}

// Resolver looks up a dotted path against whatever data tree the
// caller owns (evaluation time only — Parse/Check never resolve a
// path against data). internal/expr never sees internal/bind's own
// Value tree: bind (stage rank 4) imports expr (rank 3) to reuse this
// interface, never the reverse (D-1.6.1, F2) — so Resolver is exactly
// the seam a caller's own data model crosses through, defined here at
// the lower rank and implemented at the higher one.
//
// AD-14's three cases are split across Resolve's two return values,
// deliberately asymmetric: an ABSENT path is reported as a non-nil
// err — already the caller's own fully-worded, located error (e.g.
// bind's "data path %q is absent from the report data"), which Eval
// propagates completely unchanged, because only the resolver's own
// root (data/params/row, in bind's case) knows how to word it. An
// explicit JSON NULL is not an error at all: it is the ordinary
// Value{Kind: KindNull}, err == nil — the owner's if() ruling (D-000.x,
// this story) depends on that: a null CONDITION is a legal value that
// evaluates to the else branch, silently, and can only do that if
// resolving it did not already fail.
type Resolver interface {
	Resolve(path []string) (Value, error)
}
