# Reviewer lens — adversarial: two compliant units that still diverge

Method: construct pairs of epics that obey every AD to the letter and still produce
incompatible builds. Each surviving pair is a hole.

Verdict: **CHANGES REQUIRED** — 2 critical, 4 high, 4 medium. All applied at Finalize.

## Critical

**A1 · JSON numbers are `float64`, and nothing said otherwise.**
Report data is JSON. Go's `encoding/json` decodes every number to `float64` by default. AD-2
removes floats from *geometry* but says nothing about *data values*, so `sum(transactions.amount)`
over a bank statement runs in binary floating point. Two failures at once: `0.1 + 0.2` money
arithmetic in a product whose acceptance test is a **customer account statement**, and a
compiler free to fuse a multiply-add in an aggregate helper on arm64 but not on wasm — exactly
the NFR1.a hazard AD-2 was written to close, re-entering through the data door.
The bind epic and the expression epic can each obey every AD and still pick different numeric
types. → **AD-23**.

**A2 · No one owns the coordinate origin.**
Elements carry x/y. Nothing says whether that is page-absolute or band-relative, and nothing
says where the flip from PageModel space (top-left, Y down) to PDF user space (bottom-left,
Y up) happens. The designer epic and the layout epic can each be internally correct and
disagree by the height of the page header. The flip is also a per-call-site sign error waiting
to happen if more than one function performs it. → **AD-24**.

## High

**A3 · Aggregate scope is undefined.** `{{sum(transactions.amount)}}` in a page footer: the
whole collection, or the rows on this page? FR27 implies whole-collection; per-page subtotals
are not in scope. Unstated, the expression epic and the pagination epic pick differently and
the 50-page golden report shows it. → amended AD-11.

**A4 · An expression could depend on pagination.** AD-4 makes pass two layout-free and handles
page numbers as late-bound slots, but never forbids an expression referencing pagination. The
moment one does, evaluation has to move after layout and the two-pass structure collapses.
→ amended AD-4.

**A5 · Missing versus null data is undefined.** FR41 makes unresolvable bindings an error, but
JSON `null` and an absent key are different things, and so is a present-but-wrong-typed value.
Two areas will choose differently. Also unstated: the scalar-to-text conversion used when a
bound value is not a string — a second, accidental number formatter competing with
`formatNumber`. → amended AD-14 and conventions.

**A6 · Conditional visibility has no layout semantics.** FR20 puts visibility in scope. Is a
hidden element absent from the page or an empty box? Does hiding one reflow anything? Can a
condition hide a *table row*, which would change pagination? → **AD-24**.

## Medium

**A7 · Subset scope unstated** — per document or per page. Per page produces different bytes
for the same glyphs. → convention.

**A8 · Preview has no parameter source.** The golden report's generated date must arrive
through `params` (FR43), but no UX surface supplies parameters during preview, so S5 cannot be
demonstrated. Gap between EXPERIENCE and the PRD, not caught by either. → amended AD-16.

**A9 · Image fit rule unstated** — stretch, fit, or fill, and rounded how. Named as a
determinism control point in addendum §E and then dropped. → **AD-24**.

**A10 · Two unstated technology forks** — the expression parser (hand-written vs generated)
and template validation (the Go parser vs a JSON Schema dependency). A JSON Schema library
arriving as a second source of truth would directly undermine AD-9. → conventions.

## Pairs tested and found already closed

Font bytes diverging between native and browser (AD-8) · designer and engine holding two
`.folio` models (AD-15) · preview rendered from a live object graph rather than saved bytes
(AD-16) · canvas wrapping text by browser rules (AD-17) · stale preview shown as production
(AD-18) · table width stored twice (AD-13) · page-number fixups re-measuring in pass two
(AD-4) · PDF types leaking into the page model (AD-5).
