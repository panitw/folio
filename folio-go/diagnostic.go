// This file declares AD-14's diagnostic payload and D-2.8.3's result
// shape — the seam Story 2.8 was blocked on until the owner ruled OD-1.
//
// AD-14 already fixed the VALUE (one Diagnostic carrying Severity, a
// stable string code from a closed registry, an optional element id, an
// optional data path, and a message) and the DISPOSITION (a Warning
// accompanies a successful render, never silent, never fatal). Only the
// SEAM that carries it out of Render was open, and D-2.8.3 rules it: a
// result struct, not a third return value, not a sink parameter, not a
// sibling entry point. See the decision log's D-2.8.3 through D-2.8.6
// for the grounds — they are not re-argued here.
package folio

// Severity classifies a Diagnostic (AD-14, verbatim): "Error aborts the
// render, Warning accompanies a successful one." A Diagnostic with
// SeverityError is never returned from Render/RenderTo alongside bytes
// — an Error is reported as Go's ordinary error return instead, and
// today's render path (this story) produces only SeverityWarning
// values. The type exists now because AD-14 defines both severities as
// one closed set, and a caller matching on severity must have both
// cases to match against, even before the engine emits the second one.
type Severity int

const (
	// SeverityWarning accompanies a successful render: the PDF bytes
	// are complete and correct for what they contain, but something the
	// author asked for could not be honoured exactly as declared (FR44:
	// content clipped at its box boundary).
	SeverityWarning Severity = iota

	// SeverityError is reserved for a future Diagnostic-shaped error
	// path (AD-14's other half). Nothing in this story constructs one;
	// today's located errors (e.g. layout.OverflowError, D-2.6.5) stay
	// plain Go errors, unchanged, and are not routed through this type
	// here (see "Do not re-open" in Story 2.8).
	SeverityError
)

// String reports s as the closed-registry word AD-14 names it by
// ("Warning" or "Error"), never a bare integer — a Diagnostic's
// Severity is presented to a human, and an unlabelled 0/1 would defeat
// that purpose the first time someone printed one.
func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "Warning"
	case SeverityError:
		return "Error"
	default:
		return "Severity(" + itoa(int(s)) + ")"
	}
}

// itoa is a tiny, dependency-free int->string helper (avoids pulling
// strconv into this file for one Severity.String() fallback branch that
// only fires for a value never constructed inside this module).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// DiagCodeTextClippedWidth is the closed registry's one code this story
// mints (AD-14: "a stable string code from a closed registry"; D-2.8.1:
// FR44's only subject is the horizontal axis). It names a text
// element's widest packed line exceeding its declared width, clipped at
// the box boundary rather than reflowed or dropped (AC1/AC2/AC6).
//
// Additive only (AD-14, verbatim: "changing a code's meaning is a
// breaking change"): once shipped, this string's meaning is permanent.
const DiagCodeTextClippedWidth = "TEXT_CLIPPED_WIDTH"

// DiagCodeEmptyAverage is the closed registry's second code (Story
// 3.3, DECISION-5/R9): avg() evaluated over a present-but-empty
// collection. D-3.1a.2's kernel has no identity element for an average
// of zero operands and refuses honestly; Story 4.2's own AC requires
// an empty-collection table to render successfully, so the render
// proceeds — the aggregate resolves to empty — and this Warning is how
// a reader learns why a total column shows nothing, rather than the
// render simply failing (which would make Story 4.2's AC
// unsatisfiable) or the blank column looking unremarkable (which it is
// not: an ALL-NULL collection is a real average and renders a number,
// R7.3 — only a genuinely EMPTY collection reaches this code).
//
// Minted here, in the module-root folio package, following
// DiagCodeTextClippedWidth's own precedent (D-2.8.1): D-1.4.2 forbids
// minting a code AHEAD of the condition it names, never minting one in
// this package before internal/diag exists (Story 3.6) — the condition
// ships in this story, which is exactly when 2.8 minted.
//
// Additive only (AD-14, verbatim: "changing a code's meaning is a
// breaking change"): once shipped, this string's meaning is permanent.
const DiagCodeEmptyAverage = "AGGREGATE_EMPTY_AVERAGE"

// Diagnostic is AD-14's one diagnostic/error value. Every failure mode
// AD-14 names — over-tall rows (FR25, not yet built) and clipped
// content (FR44, this story) — is expressed as a value of this type,
// never a bespoke per-area type: the whole point of AD-14 is that a
// caller (a CLI, a designer, this library's own tests) checks one shape
// for every kind of thing that can go wrong with one element.
type Diagnostic struct {
	// Severity is Warning for every Diagnostic this story constructs
	// (see Severity's doc comment).
	Severity Severity

	// Code is a stable string from the closed registry above. Never
	// prose, never localized, never sentence-shaped — a caller
	// programmatically dispatching on the kind of problem (a CLI
	// choosing an icon, a designer highlighting a category) matches
	// this field, not Message.
	Code string

	// ElementID is the element the Diagnostic concerns, when it
	// concerns one (AD-10: "every diagnostic and every error that
	// concerns a template element carries its id"). Empty only for a
	// Diagnostic that names no single element — no such Diagnostic
	// exists in this story.
	ElementID string

	// DataPath is the bound data path the Diagnostic concerns, when it
	// concerns one. Empty for FR44's clip: a width overflow is a
	// property of the LAID-OUT content against its declared box, not of
	// one JSON value at one path.
	DataPath string

	// Message is a human-readable sentence, safe to print directly to a
	// designer or a CLI. It is never parsed by this library or by any
	// test asserting on a Diagnostic's Code/ElementID instead (D-000.21).
	Message string
}

// Result is Render's return value (D-2.8.3, the owner's decision,
// verbatim):
//
//	type Result struct {
//	    Bytes       []byte
//	    Diagnostics []Diagnostic
//	}
//	func Render(t *Template, d Data, p Params, f FontSet) (Result, error)
//
// Bytes is the complete, valid PDF document — present whenever error is
// nil, exactly as Render's pre-2.8 []byte return was. Diagnostics is
// every Warning the render produced, alongside those bytes; it is never
// consulted to decide whether Bytes is valid — a non-nil error is the
// only signal a caller needs for that, unchanged from before this
// story.
//
// Diagnostics ordering and emptiness are BOTH determinism guarantees,
// stated here because nothing else in this module checks a non-byte
// output surface (D-2.8.6):
//
//   - ORDER IS DOCUMENT ORDER: band order (page header, then content,
//     then page footer), and within one band, element DECLARATION order
//     — the order elements appear in the authored `.folio` document.
//     NEVER map order, and never an order that depends on which
//     goroutine or which pass happened to finish detecting an overflow
//     first.
//   - EMPTY IS nil, ONE REPRESENTATION: a render that produced no
//     Diagnostic ever sets Diagnostics to a non-nil empty slice. Two
//     renders of the same document therefore compare equal under
//     reflect.DeepEqual exactly when they produce identical bytes —
//     never differing only in whether Diagnostics is nil or []Diagnostic{}.
type Result struct {
	Bytes       []byte
	Diagnostics []Diagnostic
}
