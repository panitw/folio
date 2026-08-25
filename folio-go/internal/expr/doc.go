// Package expr is Story 3.2's expression language: a hand-written
// recursive-descent parser (AD-9, D-3.2.2 — no generator, no
// third-party dependency, no general-purpose expression library) over
// the small grammar `.folio` bindings use inside "{{ }}" — a bare
// dotted path, a function call over comma-separated arguments, a
// double-quoted string literal, or a number literal — plus the eight
// named functions FR18 promises (sum, count, avg, formatDate,
// formatNumber, upper, lower, if), of which only upper/lower/if are
// implemented at this story; the other five are registered, parse and
// derive successfully, and fail loudly at evaluation, naming the
// story that implements them (AC15-AC18).
//
// internal/expr is rank 3 in the stage-rank table
// (lint/internal/rules/stagerank.go): it may import internal/template
// (rank 2), which it does for exactly one thing — SplitJSONNumber, the
// module's one JSON-number-literal splitter (D-1.6.1) — and it is
// imported BACK by internal/bind (rank 4), never the reverse. This
// direction is forced, not chosen (D-3.2.1, D-1.6.1): DW-8 moves
// Decimal here because Decimal is built on SplitJSONNumber, so expr
// must outrank template, and bind — which needs Decimal for its own
// report-data value tree — must therefore sit below expr.
//
// Decimal, NewDecimal, SumDecimals and AvgDecimals (decimal.go,
// reduce.go) moved here from internal/bind UNCHANGED in behaviour
// (D-3.2.1): this is the one and only "Decimal" type declaration in
// the module (AC22), and D-3.1a.3's reducer-inventory tripwire — which
// asserts SumDecimals/AvgDecimals live in the SAME package as the
// Decimal declaration — is RELATIONAL and follows this move with zero
// edits to that guard.
//
// Two-phase load/evaluate split (R3, forced by F3 — the canonical
// golden binds a still-unimplemented formatNumber(...) call, so an
// unimplemented function cannot be a load error): Parse produces an
// AST from raw grammar alone, with no knowledge of the eight-function
// table; Check walks that AST against the table and reports every
// statically decidable defect — a syntax error, a wrong arity, an
// unknown function name, or a literal argument of the wrong kind
// (Decision 3: arity and literal-argument-kind are both decidable
// without data, so both are 3.2's to check; a PATH argument's runtime
// kind is not decidable without data and is explicitly NOT 3.2's
// obligation — owed at evaluation by Story 3.3 (sum/count/avg) and
// Story 3.4 (formatDate/formatNumber)). Eval walks the AST against a
// Resolver (evaluation-time data lookup) and actually computes a
// result, calling only the branch of an if() that was selected
// (AC14's short-circuit) and reporting a registered-but-unimplemented
// function as a LOCATED error, never a plausible value (AC15,
// guarding directly against F6/AC17's SumDecimals(nil) == {0,0}
// hazard: the "sum" table entry is never wired to SumDecimals in this
// story at all).
package expr
