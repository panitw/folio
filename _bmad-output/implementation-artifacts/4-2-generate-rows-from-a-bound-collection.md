---
baseline_commit: b5eb557
---

# Story 4.2: Generate rows from a bound collection

**Epic:** 4 — A Go developer can render the golden report (**the C4 gate**)
**Story key:** `4-2-generate-rows-from-a-bound-collection`
**Status:** `done`
**Covers:** **FR22**, **FR25** · **AD-11**
**Also discharges:** **DW-14** (`/ToUnicode` 100-entry section cap), re-owned to this story at the
Epic 3 boundary gate.

**Primary invariant — AD-11, verbatim** (`ARCHITECTURE-SPINE.md:266-278`):

> **Rule:** a repeating region declares its alias: `{"bind": "transactions[]", "as": "transaction"}`,
> defaulting to `row` when omitted. Inside the region, the alias resolves to the current row.
> Unqualified paths always resolve from the document root — a row **never** shadows the root.
> `params.` is always the parameter namespace and can be shadowed by nothing.

**Co-primary invariant — AD-13 / Story 4.1's AC2**, which this story must not weaken:
**column widths are never negotiated against content.** A cell whose content is too wide makes its
**row taller**; it never makes its **column wider**, and it never moves a later column's x-origin.
Wrapping happens *inside* the declared width. Nothing widens.

---

## Baseline, measured in this run at creation (HEAD `b5eb557`, tree clean)

| Gate | Command | Result |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **965 pass · 0 fail · 1 skip** |
| `lint` | (module `lint`, per its own README) | **115 / 0** (carried from 4.1's final gate) |
| `hashmatrix` | (module `hashmatrix`) | **3 / 0** (carried from 4.1's final gate) |

The one skip is `TestCorpusMeetsP6ExerciseFloors`, red **by design** (D-000.17 / D-2.1.14 /
D-000.57 / D-000.74). Never "fix" it. The `-skip` form above **is** the green gate.

Heavy-test cadence is **per-epic** (D-000.4). **This story writes integration/e2e tests and does
not run them, and must say so explicitly in its Delivery Log.** The cross-target hash matrix runs
once at the Epic 4 boundary; 4.7 is the only Epic 4 story on D-000.4's per-story override list.
**4.2 is not.**

---

## In plain terms (read this first if you just want the gist)

A statement with many transactions no longer needs its author to place one row per transaction by
hand. A table now prints one row for every item in the data it is bound to, in the order the data
provides them. Each row's cells pull that row's own values — a cell asking for "the amount" gets
that transaction's amount, never a neighbour's. A cell whose text is too long for its column wraps
onto extra lines, and the row simply grows taller to fit; the column itself never gets wider, and
no later column ever shifts, because a column's width stays the author's decision throughout. A
table bound to a list that turns out to be empty still prints its column headings and simply has
no rows underneath it, which is a normal, successful result rather than an error.

Partway through the work it became clear that a bordered, padded table with data rows underneath a
plain heading looked wrong — a decorated strip floating over unruled text — so the border, padding
and background that were originally built for the heading alone were extended to cover every data
row too. A vertical-alignment setting, which lets an author push a short cell's text toward the
middle or the bottom of a row when a neighbouring cell in the same row has wrapped onto several
lines, was also connected up to actually do something, which up to that point it had not.

A check of the finished work before it shipped found that several of the safeguards meant to
protect this new row-drawing behaviour were not actually capable of noticing when that behaviour
broke — they would have stayed green even if a row's border, its padding, or the internal
bookkeeping a future page-breaking feature will depend on had silently stopped working. Those
safeguards have since been rewritten so each one demonstrably fails when the specific thing it
protects is broken, proven by deliberately breaking each thing in turn and watching the right
check, and only that check, react. A number of smaller wording and record-keeping slips in the
accompanying tests and documentation were also cleaned up. One narrow edge case was set aside
rather than fixed here: what should happen when every cell in every row of a table has nothing at
all to display and the table has not declared a font for its body text is left for whoever next
works on that class of font-related error message, because doing it well means revisiting how an
older, unrelated error is decided — a bigger change than this piece of work should make on its own.

What this work still deliberately does not do: it does not handle a table that runs past the
bottom of a page, does not repeat the heading on a following page, does not compute totals, does
not deal with a single row too tall to fit anywhere on a page, and does not shade alternating
rows. Those remain later work, and every example table built for this piece of work was kept
small enough to stay on one page, so none of those later behaviours could be accidentally
demonstrated — or accidentally relied upon — ahead of time.

---

## Story

**As a** template author,
**I want** one row per item in my data,
**So that** a statement with 400 transactions prints 400 rows without me placing any of them.

---

## Obligations this story inherits, and what it does NOT inherit

### Inherited and DUE HERE

1. **DW-14 — chunk `/ToUnicode` `beginbfchar` sections at 100 entries.** Re-owned to this story at
   the Epic 3 boundary gate. See AC9 and the measurement section below. It is due here **because of
   the timing, not because of the code**: chunking is byte-identical for every document already
   under the cap, so it costs **zero re-record now**, and costs a four-page-count × four-target
   golden-report re-record if it lands at 4.7 — inside the story that *is* the C4 gate.
2. **`TestTableRendersZeroDataRows` (`table_render_test.go:518`) is this story's red-to-green.**
   4.1 marked it *"expected to be REWRITTEN by Story 4.2"* in its own doc comment. Rewrite it —
   do not delete it, and do not weaken it (see AC1 and AC4).

### Inherited, DEFERRED, and to be PLACED not absorbed

3. **The "no resolvable `fontFamily`" diagnostic-code question** (4.1 Finding 13). 4.1 deferred it
   *"to whoever next touches font-resolution error handling"*. **This story does not mint a code.**
   D-000.65 mints a code where its condition **first** ships, and this condition shipped in Story
   1.5's `fontChain` and again in 4.1's header path. What this story owes is **not a third divergent
   spelling of that error** — see AC5, which requires the data-cell failure to route through the
   same helper as the header's. Leave the code question with the owner.
4. **`TableGeometry.Width()` overflow protection** (4.1 Finding 10, bundled with 4.1's DECISION-3
   "a table wider than its band"). **Not this story.** 4.2 adds no new reader of `Width()`.
5. **D-2.8.6's within-band text-vs-table diagnostic ordering** (4.1 Finding 7). 4.1's finisher
   judged this *"better placed against 4.2–4.8's richer table surface"*. 4.2 is the first story that
   produces table diagnostics **from data**, so it is the first story that can pin the ordering.
   See **AC8** and **DECISION-3**: this story **pins the actual ordering in a test and names the
   deviation**; it does **not** change the ordering, which is a behaviour change needing a ruling.

### Explicitly NOT inherited

6. **DW-16** (`ShapedGlyph.CID` is not always a glyph id). DW-16's own text: *"Stories 4.1 and 4.2
   do **NOT** close the window."* 4.2 owes it nothing. Do **not** add `cid` to
   `pdfConceptSubstrings`.
7. **DW-4 / DW-13 / DW-11 / DW-20** — owner-batched or scheduled elsewhere (D-000.78, D-4.0.2).
   Nothing here.
8. **`altRowBackground`** is parsed and inert. It stays inert. **Story 4.8** owns it.
9. **`columns[].footer` / `footerOf` / `footerFormat`** are parsed and derived at load. They stay
   unread at render. **Story 4.5** owns them (D-000.65 routes
   `TABLE_FOOTER_SOURCE_UNRESOLVED`/`_FORBIDDEN` there explicitly).

---

## FIVE divergences between the brief / the record and the shipped tree, all measured in this run

Every number below was produced at `b5eb557` by the command shown. Nothing here is inferred.

### D1 — `fixtures/page-count-20/` carries **32** `/ToUnicode` entries, not "20 pages' worth"

DW-14 warned that page-count-20's green cap-assert must not be read as headroom. Measured, and it
is worse than the warning: a **20-page** document has **fewer** entries than a **two-page** one.

```
$ python3 -c '<regex over each fixtures/*/expected.pdf, per DW-14>'
fixtures/font-text/expected.pdf              declared [25]        counted [25]
fixtures/multi-page/expected.pdf             declared [45]        counted [45]
fixtures/multi-script-fallback/expected.pdf  declared [4, 1, 1]   counted [4, 1, 1]
fixtures/page-count-20/expected.pdf          declared [32]        counted [32]
fixtures/shaped-text/expected.pdf            declared [14, 7, 28] counted [14, 7, 28]
fixtures/three-band-page/expected.pdf        declared [17]        counted [17]
fixtures/wrapped-text/expected.pdf           declared [28, 18, 38] counted [28, 18, 38]
```

**Largest section anywhere in the repository: 45**, on a two-page single-face document. Page count
is not the driver; **distinct (glyph, cluster text) content** is, exactly as DW-14 says. Record this
table in the Delivery Log — it is the "both directions" figure D-000.49 requires.

`fixtures/page-count-1/`, `-5/`, `-50/`, `expected-breaks/`, `thai-break-corpus/` and
`hidden-image/` carry **no** `beginbfchar` section at all (measured: zero matches), so they are not
witnesses either way.

### D2 — `layout.ColumnItem` **enforces** content exclusivity in code, and a row needs both kinds

`internal/layout/paginate.go:290-307` returns `&MixedItemError{}` for any item populating more than
one of `Runs` / `Images` / `Rects`. 4.1 worked around this by emitting **two items with the same
`Top`/`Bottom`** — its own comment (`paginate.go:100-108`): *"two items, same extent, is how they
land on the same page without a fourth exclusivity case."*

A data **row** has both cell chrome (rects) and cell text (runs), and a **wrapped** row's text is
several `wrappedLine`s at *different* extents. So a three-line row is, in today's model, four
separate `ColumnItem`s, three of which have distinct extents. **There is no atomic "row" in
`layout` today.** That is Story 4.3's problem ("the row moves whole"), not this story's — but 4.2
must not make it harder, and must not pre-empt it. See **AC7** and **DECISION-2**.

### D3 — every `columns[].bind` in `table_render_test.go` names a root path, not the row alias

`columns[].bind` has **never been evaluated** (D-1.6.8's fence: 3.1 did not evaluate it, 3.2 did not
wire it into the table, 4.1 rendered labels only). The moment this story evaluates it, these break:

| Site | bind | table's `as` | effective alias | outcome once evaluated |
|---|---|---|---|---|
| `table_render_test.go:98-99` (`twoColumnsNoAlign`) | `{{item.a}}`, `{{item.b}}` | absent | `row` | `data.item` **absent → Error** |
| `table_render_test.go:226-227` | `{{item.a}}`, `{{item.b}}` | absent | `row` | same |
| `table_render_test.go:335-336` | `{{item.a}}`, `{{item.b}}` | absent | `row` | same |
| `table_render_test.go:681-683` (`threeColumnTableDoc`) | `{{item.a/b/c}}` | absent | `row` | same |
| `table_render_test.go:853` | `{{row.n}}` | absent | `row` | **fine** |
| `render_table_bind_test.go:19` | `{{transaction.amount}}` | `"transaction"` | `transaction` | **fine** |
| `render_bind_test.go:31` | `{{formatNumber(transaction.amount, …)}}` | `"transaction"` | `transaction` | **fine** |
| `folio_expr_validate_test.go` (5 sites) | `{{transaction.amount}}` | declared per case | — | **fine** |
| `testdata/template/golden/worked-example.json:12,19` | `{{transaction.date}}`, `{{formatNumber(...)}}` | `"transaction"` | `transaction` | **fine** (round-trip fixture, never rendered) |

**Nine column binds across four column-set constants in `table_render_test.go` must be re-pointed
from `item.` to `row.`.** These are fixture corrections, not behaviour changes; say so in the
Delivery Log and do not let them be read as a regression.

### D4 — `TestTableStyleFieldsAreNotDataDriven` renders `"items": []` and goes **vacuous** the moment rows ship

`table_render_test.go:579-597` proves "style is never data-driven" by rendering two datasets that
differ only in an unbound field `overdue` — but its collection is **empty**, so today it compares
two header-only documents, and after this story it will *still* compare two header-only documents
while claiming a property about a table with rows. This is the exact shape 4.1's review caught twice
(Blockers 1 and 2). **AC10 requires it to be given a non-empty collection**, with cell binds
resolving real row fields, and `overdue` still unbound.

### D5 — `headerStyle` is **header-only**, so a data cell cannot inherit its `fontFamily`

D-000.76 ruled `headerStyle` governs *"the header row ONLY, never a data row"*.
`resolveHeaderStyle` (`table_render.go:70`) cascades `headerStyle` → `style` → default. A data cell
must cascade **`style` → default only**. A table declaring `headerStyle.fontFamily` and **no**
`style.fontFamily` therefore renders its header successfully today and must become a **located
error** for its data cells (D-4.1.1: there is no font default; `fontFamily` is required on any
element carrying text). A first draft that reuses `resolvedHeaderStyle` for cells gets this wrong
silently. **AC5** is the guard.

---

## Do not re-open — settled rulings this story inherits

| Ruling | What it settles | Why it stands |
|---|---|---|
| **AD-13 / 4.1 AC2** | Column widths are never negotiated against content. | The whole table pipeline's tractability rests on it; 4.3's pagination assumes it. |
| **D-000.76** | `headerStyle` is a full `Style` block governing the header row **only**. | Owner ruling. A data row is not a header row. |
| **D-4.1.1** | There is **no** `fontFamily` default; it is **required** on any element carrying text. `folio-format.md` amended. | Sorted-key defaulting was proposed and rejected: renaming a chain would silently change output. Require-now/default-later is additive and moves no hash; the reverse moves every hash. |
| **AD-11 / D-3.1.1** | The row alias is a **third resolution root**. Unqualified paths resolve from the document root; a row never shadows it; `params.` is shadowed by nothing. | Resolves PRD Q3. |
| **AD-14** | One `Diagnostic` type, closed registry, **never a panic**. Absent = Error; explicit `null` = empty and **not** an error; wrong kind = Error, never coercion. | Pre-settled data cases. |
| **D-000.65** | A code is minted where its condition **first** ships. | Keeps 4.2 from minting a code for a 1.5-era condition. |
| **D-2.8.1** | The declared **width** is a clip bound; a declared **height** is not. | Governs the residual over-wide-token case in AC2. |
| **D-000.68** | A guard is anchored to the compiler, the type system, or a literal the test owns — never the code's own spelling. | 4.1's AC2 anchor pattern (test-owned expected X/W literals) is the model to copy. |
| **D-000.9 + extension** | *"A guard's 'all clear' must never be produced by the same code path as its 'I could not look'"*, and *"a measurement offered as evidence must name the command, name the mutation under which it reddens, and confirm the mutation was run."* | Applies to this story's **records**, not only its tests. |
| **D-000.49** | Numbers are asserted in **both** directions. | Governs the DW-14 figures in D1. |
| **AD-23 / AD-1** | Nothing under `internal/` touches `time`, `os`, `math/rand`, `net`, `math` transcendentals, package-level mutable state, output-reaching map iteration, or **any binary float**. | `internal/geom.Length` is `int64` millipoints. Row heights are integer millipoint arithmetic. `internal/expr.Decimal` is for report **data**, not geometry — do not conflate the two regimes. |

---

## R — design constraints derived from the record during creation

- **R1.** Row geometry is **integer millipoints** (`geom.Length`, `int64`). No `float64`, no
  `big.Float`, no `expr.Decimal` in a coordinate. The `lint` module's `TestFloatTypedProductionScan`
  is the instrument, **not** `go build` (4.1 Finding 4: `go build ./...` succeeds on a float).
- **R2.** All cells of one table share **one** vertical model: the chain and size come from the
  table's `style`, and there is no per-column style in the schema. Compute
  `chainVerticalModel(chain, fontSize, …)` **once per table**, outside the row loop, exactly as
  `collectBandTextRuns` computes it once per element (`render.go:768`). Do not recompute per cell.
- **R3.** A block of *n* lines is `vm.FirstBaseline + (n-1)*vm.Advance + vm.LastDescent`. 4.1's
  header used the n=1 form (`textBlockHeight := vm.FirstBaseline + vm.LastDescent`,
  `table_render.go:~370`). Extend that expression; do not invent a second leading model.
- **R4.** Glyphs reach the page model only through the existing shaper→pagemodel bridge.
  `lint`'s `TestGlyphIdentifierCensus` pins producers of `pagemodel.ShapedGlyph` at **two** and
  readers of `.CID` to `internal/pdf` plus the copier. **4.2 adds callers of
  `shapeSegments`/`positionSegments`, never a new construction site.** If the implementation seems
  to need one, that is a decision to surface, **not** a census to widen.
- **R5.** **No new package under `internal/`.** A new package needs a rank in `stagerank.go` **and**
  a row in the spine's fenced ladder in the same commit
  (`TestSpineStageLadderMatchesStageRankTable`). This story has no need of one: row generation
  belongs beside `collectBandTableRuns` in `folio-go/table_render.go`, and any pure geometry helper
  belongs in the existing `internal/layout` package.
- **R6.** **No new method on a package-`folio` root type.** `TestFolioMethodNamesAreInjective`
  (DW-20 / D-4.0.2) reddens on a duplicate method name across receivers.
- **R7.** **Invent no schema fields.** Measured against `folio-format.md` and
  `internal/template/model.go`: there is **no** `rowHeight`, **no** per-row or per-cell `Style`,
  **no** `Style` on `Column` (a column has `align` only), **no** wrap/clip/truncate toggle, **no**
  row filter/sort/limit, **no** per-row `visibleIf`, **no** table `width`/`height`. A missing field
  is a **finding to surface**, not a field to add.
- **R8.** Every test this story adds keeps its table **on one page**. Pagination behaviour belongs
  to 4.3 and must not be asserted, relied on, or accidentally demonstrated here.

---

## Acceptance Criteria

Every AC below states its **instrument** (the test that observes it), its **anchor** (what the
assertion is pinned to, per D-000.68), and its **red-proof** (the mutation under which the
instrument reddens). **Run every red-proof over the WHOLE suite, never behind a narrow `-run`** —
4.1's Finding 1 was a mutation that reddened for the wrong reason while a narrow `-run` hid that the
vacuity control reddened too.

---

### AC1 — One row per collection element, in data order

**Given** a table bound to a collection
**When** the document renders
**Then** exactly one row is generated per element of the collection, in the collection's own order.

- **Instrument.** Rewrite `TestTableRendersZeroDataRows` (`table_render_test.go:518`) into
  `TestTableRendersOneRowPerCollectionElementInDataOrder`, keeping its two-assertion shape (rect
  count **and** run count **and** `SourceText`) — 4.1's Finding 14 fixed that test precisely because
  a rect-only fence would have gone green under a text-only row implementation.
- **Anchor.** Test-owned literals: a five-element collection whose cell values are five *distinct*
  strings the test declares; expected run count = `columns × (1 header + 5 rows)` adjusted for
  wrapping (keep every cell single-line here); expected rect count = `columns × (1 + 5)`; and the
  **ordered** list of `SourceText` values compared against the test's own literal slice. **Data
  order is asserted by comparing the sequence, never a set** — a set assertion cannot see a reversal.
- **Vacuity fence.** **No AC may be satisfied by asserting on an empty collection.** This test's
  collection is non-empty by construction, and a five-element collection with five distinct values
  is what makes "in data order" a positive claim rather than an accident.
- **Red-proof.** (a) revert `collectBandTableRuns` to emit the header only → run/rect counts drop to
  the header's; (b) reverse the row loop (`for i := len(items)-1; i >= 0; i--`) → counts stay right,
  the ordered `SourceText` comparison reddens **and nothing else does**. Proof (b) is the one that
  matters; record both, and record that **no other test in the suite reddened under (b)**.

---

### AC2 — A too-wide cell wraps inside its column and the ROW grows taller; nothing widens

**Given** a cell whose content exceeds its column width
**When** the row is laid out
**Then** the content wraps within the column's content width and the row grows taller
**And** the column's width, every column's x-origin, and the table's total width are unchanged.

- **Mechanism.** Reuse `packLines(segs, ops, totalRunes, fontSize, contentWidth)` (`wrap.go:179`) —
  the same packer text elements use — with `contentWidth = column.Width - padLeft - padRight`.
  Honour the document's `unbreakableValues` through `atomicSpansFor(doc.doc.UnbreakableValues, subs)`
  exactly as `collectBandTextRuns` does (`render.go:736`).
- **Row height.** Per R2/R3: `rowHeight = padTop + vm.FirstBaseline + (maxLines-1)*vm.Advance +
  vm.LastDescent + padBottom`, where `maxLines` is the maximum over the row's cells. Integer
  millipoints throughout (R1).
- **Instrument.** `TestWrappedCellGrowsTheRowAndNeverTheColumn`, modelled on
  `TestColumnGeometryNeverNegotiatesAgainstLabelContent` (`table_render_test.go:659`) — that test is
  4.1's anchor pattern and it exists at exactly this shape for exactly this reason. **Three
  columns**, so a widened middle column would visibly push the third column's origin. Two renders:
  short middle-cell value vs a long one, in all three shaping paths (long Latin, long space-less
  Thai, long CJK — the AC2 requirement 4.1 already established for labels).
- **Anchor.** **Test-owned literal** `{X, Width}` pairs for all three columns and a literal expected
  total, identical for the narrow and the wide render. Read neither from either render. (4.1's
  Finding 8: the original version used one render's output as the other's expectation, which is not
  an anchor.)
- **The row-grows assertion.** The wide render's row extent (`Bottom - Top`, or the y-delta between
  consecutive rows' first-line baselines) is **strictly greater** than the narrow render's, by
  **exactly `(n-1) * vm.Advance`** for a test-owned expected `n`. Assert the exact delta, not merely
  ">" — a ">" passes under any bug that makes rows arbitrarily tall.
- **Control (D-000.9).** The two renders must differ in glyph output (run count / `SourceText`), or
  every geometry equality above is vacuous — both could have rendered nothing. Assert the control
  **in the same test**, and **run the red-proofs without `-run`** so a control reddening alongside
  the guard is visible (4.1 Finding 1).
- **Red-proofs.** (a) pass the *measured* text width instead of the declared width into
  `layout.ColumnWidths` → the geometry literals redden and the control stays green; (b) pass
  `boxWidth = 0` (unbounded) to `packLines` for cells → the row-height delta reddens; (c) compute
  row height as a constant `headerHeight` → the delta reddens.

---

### AC3 — Residual overflow inside a column is CLIPPED, not widened, with the existing warning

**Given** a cell whose content contains no break opportunity narrow enough for the column
**When** the row is laid out
**Then** the widest packed line is clipped at the column's content box (left/right edges)
**And** a `Warning` with the **existing** `DiagCodeTextClippedWidth` names the element
**And** the column's width and every later column's origin are still unchanged.

- **Grounds.** D-2.8.1 — the declared width **is** a clip bound. 4.1 already does exactly this for
  an over-wide header label (`table_render.go`, `clipToBox`/`clipX`/`clipWidth`). Reuse it.
- **No new diagnostic code.** D-000.65: this condition first shipped in Story 2.8. Reuse
  `DiagCodeTextClippedWidth` and `detectWidthOverflow` (`render.go:873`). **Do not mint a code.**
- **Instrument.** `TestUnbreakableCellContentIsClippedNotWidened` — one cell containing a single
  long unbreakable token. Assert: exactly one `Warning`, its code, its element id; `ClipToBox` set
  on the affected runs with `ClipWidth == column content width` (a **test-owned literal**); and the
  column geometry literals unchanged.
- **Red-proof.** Drop the `clipToBox` assignment for cells → the clip assertion reddens while the
  geometry assertions stay green (proving the two are independent).

---

### AC4 — Each cell binding resolves in the table's row scope, per Story 3.1 / AD-11

**Given** each cell binding
**When** it is resolved
**Then** it resolves against a scope whose row root is the current element under the table's alias
**And** the alias is the declared `as`, or `"row"` when absent
**And** an unqualified path still resolves from the **document root**, never from the row
**And** `params.` still resolves to parameters and is shadowed by nothing.

- **Mechanism.** `bind.NewScope(data, params).WithRow(rowValue, alias)` with
  `alias = resolvedRowAlias(tbl.As)` (`render.go:244`), then
  `bind.Resolve(col.Bind, scope, fc, string(col.ID))`. `fc` is the document's
  `expr.NewFormatContext(doc.doc.Locale, doc.doc.UTCOffset)` — constructed once per band collection,
  never per cell, and **never** from the host (AD-1).
- **Diagnostic located by COLUMN id.** `columns[].id` exists precisely *"so a diagnostic can name a
  column"*. A cell's error/warning must carry the **column** id, not only the table's element id.
- **AD-14's three pre-settled data cases, asserted individually:** absent path → **Error**;
  explicit JSON `null` → **empty cell, not an error**; wrong kind → **Error, never coercion**.
- **Instrument.** `TestCellBindingsResolveInRowScope`, a table matrix over four columns in one
  render: `{{row.a}}` (row root), `{{a}}` (document root — must **not** see the row's `a`),
  `{{params.p}}`, and a declared-alias variant `as: "transaction"` / `{{transaction.a}}` in a second
  render. Plus `TestCellBindingAbsentNullWrongKind` for the AD-14 triple.
- **Anchor.** Test-owned literals: the document root's `a` and the row's `a` are **deliberately
  different strings**, so "unqualified resolves from the root" is falsifiable. A test where they
  coincide proves nothing.
- **Red-proof.** (a) resolve every cell against the data root, ignoring the row → the row-root
  column reddens; (b) make the row shadow the root → the unqualified column reddens; (c) hardcode
  `"row"` instead of calling `resolvedRowAlias` → the declared-alias render reddens.

---

### AC5 — A data cell's font comes from the table's `style`, never from `headerStyle`

**Given** a table declaring `headerStyle.fontFamily` and **no** `style.fontFamily`
**When** it renders with a non-empty collection
**Then** the header renders and the **data cells** produce a located error naming the element
**And** the error is the **same** message the existing font-resolution failure produces — no third
spelling, no new diagnostic code.

- **Grounds.** D-000.76 (`headerStyle` is header-only) + D-4.1.1 (no font default; `fontFamily`
  required on any element carrying text). Divergence **D5** above.
- **Mechanism.** A second, separate cascade for the body: `style.<field>` → documented default,
  with **no** `headerStyle` arm. Do not reuse `resolvedHeaderStyle` for cells. The same rule applies
  to `fontSize`, `padding`, `border`, `background`, `align` and `valign` for data cells: **`style`
  only**. `columns[].align` still wins over `style.align` for that column (both header and body).
- **Instrument.** `TestDataCellsDoNotInheritHeaderStyle` — two assertions in one test: (i) the
  `headerStyle.fontFamily`-only table errors on its data cells and the message names the element;
  (ii) a table whose `headerStyle` sets a **different** `background`/`align` from `style` renders
  header cells with `headerStyle`'s and **data** cells with `style`'s, asserted on
  `pagemodel.Rect.Fill` and on the runs' x-origins against **test-owned literals**.
- **This AC exists because of 4.1's Blocker 1.** `headerStyle` shipped inert: deleting its render
  handling left the entire three-module gate green. Assertion (ii) is the instrument that would have
  caught it, extended to the header/body boundary this story creates.
- **Red-proof.** Cascade cells through `resolveHeaderStyle` instead of the body cascade → (i) stops
  erroring **and** (ii)'s data-cell fill/align reddens. Both halves must be observed; record both.

---

### AC6 — An empty collection renders the header, produces no rows, and succeeds

**Given** a table bound to an empty collection
**When** the document renders
**Then** the header renders, **no** data rows are produced, and the render **succeeds**.

- **⚠ THIS AC IS ALREADY TRUE AT HEAD AND WILL PASS WITH ZERO NEW CODE.**
  `TestRenderTableBindEmptyArrayIsNotAnError` (`render_table_bind_test.go`) passes at `b5eb557`, as
  part of the measured 965. **As the epic phrases it, this AC is vacuous.** It must be strengthened
  or it is a record of nothing.
- **Instrument.** `TestEmptyCollectionRendersHeaderOnly`, which asserts **all three** halves
  positively, against test-owned literals: rect count `== len(columns)` (**the header is present** —
  not merely "no error"), run count `== len(columns)` with each `SourceText` a column **label**, and
  `Render` returns bytes with no error and **no diagnostic**.
- **Red-proofs — both are required, and each must redden a DIFFERENT assertion.**
  (a) emit one row for an empty collection (`for i := 0; i <= len(items); i++`) → the counts redden;
  (b) skip the whole table when the collection is empty → the **header-present** assertions redden.
  A test that only survives (a) has not established this AC. Record which assertion each mutation
  moved.
- **`null` vs `[]`.** An explicitly-`null` bind is already a **render error** (`checkTableBindings`,
  `render.go:274`; `TestRenderTableBindNonArrayFailsRender/null`). That is settled and unchanged;
  do not "fix" it into an empty collection.

---

### AC7 — Row output is emitted in the shape 4.3 needs, and 4.2 claims no pagination property

**Given** the rows this story generates
**When** they reach `layout.Paginate`
**Then** each row's chrome is one `ColumnItem` spanning that row's full extent, and each wrapped
line is a line item within it — the **same two-items-same-extent shape 4.1 used for the header**
**And** this story asserts **no** pagination property whatever.

- **Grounds.** Divergence **D2**: `Paginate` returns `MixedItemError` for any item mixing kinds
  (`paginate.go:290-307`), so a row cannot today be one atomic item. **Story 4.3 owns "the row
  moves whole", and therefore owns whichever mechanism delivers it** — a fourth exclusivity case, a
  relaxation of `MixedItemError`, or a grouping key on `ColumnItem`. **4.2 must not pre-empt that
  choice, and must not silently make it by accident.**
- **Fence (R8).** Every table in every test this story adds fits on one page. No multi-page table
  fixture. No assertion about page boundaries. If a test's table paginates, shrink the fixture.
- **Instrument.** A negative fence, `TestStory42AddsNoMultiPageTableFixture`, is **not** worth
  writing; instead the Delivery Log records, per new test, the page count observed (`len(pages)`),
  and every one is `1`. State this as a measurement, with the command.
- **Hand-off note for 4.3, to be written into the code comment beside the row loop:** the atomic
  unit does not exist in `layout` yet; `paginate.go:100-108`'s "two items, same extent" comment is
  the thing 4.3 must revisit.

---

### AC8 — Table diagnostics from data are located, ordered, and the D-2.8.6 deviation is pinned

**Given** a document whose content band holds a text element and a table, both producing diagnostics
**When** `Result.Diagnostics` is read
**Then** the order is exactly what the code produces today, and a test says so by name.

- **Grounds.** 4.1's Finding 7 measured the deviation: `render.go:1450-1456` assembles
  `headerDiags, headerTableDiags, contentDiags, contentTableDiags, footerDiags, footerTableDiags`,
  so **within a band, every text element's diagnostics precede every table's**, regardless of
  declaration order — while `render.go:1443-1448`'s own comment states D-2.8.6 as *"band order, then
  element declaration order within a band"*. 4.1's finisher forwarded the **test** and the
  interleave-vs-relax **decision** to 4.2–4.8.
- **This story pins, it does not change.** Write
  `TestWithinBandTableDiagnosticsFollowAllTextDiagnostics` with a content band declaring a **table
  first** and a **text element second**, both emitting a Warning; assert the text element's
  diagnostic comes **first**, and document in the test's own comment that this is the *deviation*
  from D-2.8.6's stated rule, pinned deliberately so a future interleave is a visible, deliberate
  change rather than a silent one. **Surface the interleave decision (DECISION-3), do not take it.**
- **Cell-level location.** Within one table, cell diagnostics are emitted in **row order, then
  column declaration order**, each naming its **column** id (AC4). Assert the full ordered sequence
  against a test-owned literal; never a set, never a map.
- **Red-proof.** Swap `contentDiags` and `contentTableDiags` in the assembly → the ordering test
  reddens and nothing else does.

---

### AC9 — DW-14: `/ToUnicode` sections are chunked at 100 entries, byte-neutrally

**Given** any face whose `ToUnicode` table carries more than 100 entries
**When** the CMap stream is built
**Then** it is emitted as consecutive `N beginbfchar … endbfchar` sections of at most 100 entries
each, in the same ascending-CID order, with every entry present exactly once
**And** every document already under the cap is **byte-identical** to its pre-change self.

- **Where.** `internal/pdf.buildToUnicodeCMap` (`internal/pdf/textdoc.go:571`). One emitter, one
  local change. Do **not** cap the number of CIDs — the allocation is D-2.3.2 and is correct.
- **The witness DW-14 names is not sufficient on its own.** `assertToUnicodeSectionsUnderCap`
  (`wrapped_text_fixture_test.go:374`) fails only when a section **exceeds** 100 — after chunking it
  can never fail, by construction. That is a D-000.9 defect: its "all clear" and its "could not
  look" become the same value. **Keep it** (it still witnesses the golden fixtures), but the real
  witnesses are the two below.
- **Witness 1 — byte-neutrality, MEASURED IN THIS RUN, and it is a real measurement.** With the
  chunking applied at cap 100, the full gate is **965 pass / 0 fail / 1 skip** — unchanged from
  baseline. And the golden suite genuinely observes CMap bytes: with the cap set to **2**, the gate
  goes to **954 pass / 11 fail**:
  `TestRenderMatchesFontTextGoldenFixture`, `TestMultiScriptFallbackGoldenFixture`,
  `TestMultiPageGoldenMatchesTheCommittedArtifact`, `TestShapedTextFixtureCIDsOriginateInShapedRuns`,
  `TestShapedTextFixtureToUnicodeRoundTrips`, `TestShapedTextGoldenFixture`,
  `TestThreeBandPageGoldenFixture`, `TestWrappedTextGoldenFixture`,
  `TestFormatLocaleGoldensAllFourLocales`, `TestFormatLocaleEndToEndRenderThai`,
  `TestFormatLocaleEndToEndRenderJapanese`.
  **This discharges D-000.9's extension**: "byte-identical" is not an assertion that could not fail.
- **Witness 2 — a direct chunking test is MANDATORY, because the fix is otherwise inert.** Nothing
  in the repository has more than 45 entries (D1), so at cap 100 the chunking branch is never taken
  and the whole change could be reverted with the gate still green — **exactly 4.1's Blocker 1**.
  Add `TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap` in `internal/pdf`, constructing a synthetic
  `EmbeddedFace` with a **test-owned** 250 entries and asserting sections `[100, 100, 50]`, each
  section's **declared** count equal to its **actual** count, and the concatenated entries equal to
  the input in order.
- **Red-proof, RUN IN THIS RUN.** With the emitter reverted to unbounded, the probe printed
  `entries=250 sections=1 declared=[250] counted=[250]` and failed on both `want 3 sections` and
  `section 0 carries 250 entries, over the 100 cap`. With the chunking applied it printed
  `entries=250 sections=3 declared=[100 100 50] counted=[100 100 50]` and passed.
- **Also record** the D1 table of per-fixture section sizes in the Delivery Log, in both directions
  (D-000.49), including the explicit statement that **`page-count-20` carries 32 entries — fewer
  than two-page `multi-page`'s 45 — so its green is not headroom for 4.7's varied data.**
- **Task-1 restoration.** The measurement fixture is reproduced verbatim below. Restore it, confirm
  it reproduces, then keep it (it is a real test this story ships).

---

### AC10 — Style is still never data-driven, now asserted against a table that HAS rows

**Given** two report datasets differing only in a field the template does not bind
**When** each is rendered
**Then** the two documents are **byte-identical**.

- **Grounds.** Divergence **D4**. `TestTableStyleFieldsAreNotDataDriven`
  (`table_render_test.go:579`) currently renders `"items": []` and would remain a header-only
  comparison forever.
- **Instrument.** Change its two datasets to a **non-empty** `items` array **identical in both**,
  with `overdue` still the only differing field, and with the cell binds re-pointed per D3. Also
  update `TestTableStyleFieldsAreNotDataDrivenControl` (`table_render_test.go:766`) so its control
  still demonstrates that a **bound** difference *does* move the bytes — otherwise the byte-identity
  assertion is again vacuous.
- **Red-proof.** Make any cell style read a data field → the identity assertion reddens **and the
  control stays green**. Run without `-run`.

---

### AC11 — Determinism and the standing bans hold

**Given** the same template, data and params
**When** rendered twice, and rendered under `-count=2`
**Then** the bytes are identical.

- No `time`, `os`, `math/rand`, `net`, `math` transcendentals, package-level mutable state,
  **output-reaching map iteration**, or any binary float on the new path (AD-23 / AD-1).
- **Row iteration is over a slice, in index order** — never over a Go map. The row `Value` is
  `bind.Value`; if any per-cell lookup goes through a map, its **output order** must come from the
  column slice, not the map.
- **Instrument.** The existing determinism/golden suite plus `-count=2` on the whole module.
- **Red-proof for the float ban:** the instrument is `lint`'s `TestFloatTypedProductionScan`, **not**
  `go build ./...` (4.1 Finding 4 measured that `go build` succeeds on a float-typed production
  declaration). Record the `lint` invocation and its 115/0 result, not a build.

---

## The measurement fixture, verbatim

Restore this as **task 1** of AC9 and confirm it reproduces. It was written, run pre-fix (red), run
post-fix (green), and removed at creation so the baseline is clean; **this story keeps it.**

File: `folio-go/internal/pdf/tounicode_chunk_test.go`

```go
package pdf

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap is DW-14's direct
// witness. It is REQUIRED because the chunking is otherwise inert: the
// largest ToUnicode section anywhere in the repository is 45 (measured
// at b5eb557), so at cap 100 the chunking branch is never taken by any
// fixture and the whole change could be reverted with the gate green.
//
// Anchor (D-000.68): the 250 entries and the expected [100,100,50]
// split are literals this test owns; nothing here reads the emitter's
// own constant.
func TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap(t *testing.T) {
	const n = 250
	face := EmbeddedFace{Name: "probe"}
	for i := 0; i < n; i++ {
		face.ToUnicode = append(face.ToUnicode, CIDText{CID: uint16(i), Text: "a"})
	}
	got, err := buildToUnicodeCMap(face)
	if err != nil {
		t.Fatalf("buildToUnicodeCMap: %v", err)
	}
	re := regexp.MustCompile(`(?s)(\d+) beginbfchar\n(.*?)endbfchar`)
	ms := re.FindAllStringSubmatch(string(got), -1)
	if len(ms) == 0 {
		t.Fatal("no beginbfchar section at all")
	}
	var counted, declared []string
	total := 0
	for _, m := range ms {
		c := strings.Count(m[2], "\n")
		total += c
		counted = append(counted, strconv.Itoa(c))
		declared = append(declared, m[1])
	}
	t.Logf("entries=%d sections=%d declared=%v counted=%v", n, len(ms), declared, counted)
	if len(ms) != 3 {
		t.Errorf("want 3 sections for %d entries, got %d", n, len(ms))
	}
	for i, m := range ms {
		if m[1] != counted[i] {
			t.Errorf("section %d declares %s but carries %s entries", i, m[1], counted[i])
		}
		if c := strings.Count(m[2], "\n"); c > 100 {
			t.Errorf("section %d carries %d entries, over the 100 cap", i, c)
		}
	}
	if total != n {
		t.Errorf("sections carry %d entries in total, want %d", total, n)
	}
}
```

**Recorded outputs.**

Pre-fix (`buildToUnicodeCMap` unbounded, i.e. HEAD `b5eb557`):

```
$ env CGO_ENABLED=0 GOWORK=off rtk proxy \
    "go test -count=1 -run TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap -v ./internal/pdf/"
=== RUN   TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap
    entries=250 sections=1 declared=[250] counted=[250]
    want 3 sections for 250 entries, got 1
    section 0 carries 250 entries, over the 100 cap
--- FAIL: TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap (0.00s)
FAIL
```

Post-fix (chunked at 100):

```
=== RUN   TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap
    entries=250 sections=3 declared=[100 100 50] counted=[100 100 50]
--- PASS: TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap (0.00s)
```

**The candidate fix, as measured** (replaces the single unbounded loop in
`internal/pdf/textdoc.go:578-592`; the developer may improve on it, but this exact form was the one
measured byte-neutral):

```go
	const bfcharSectionCap = 100
	for start := 0; start < len(face.ToUnicode); start += bfcharSectionCap {
		end := start + bfcharSectionCap
		if end > len(face.ToUnicode) {
			end = len(face.ToUnicode)
		}
		section := face.ToUnicode[start:end]
		c = appendInt(c, int64(len(section)))
		c = append(c, " beginbfchar\n"...)
		for _, e := range section {
			c = append(c, '<')
			c = appendHex4(c, e.CID)
			c = append(c, "> <"...)
			for _, r := range e.Text {
				if r > 0xFFFF {
					return nil, &supplementaryPlaneError{face: face.Name, r: r}
				}
				c = appendHex4(c, uint16(r))
			}
			c = append(c, ">\n"...)
		}
		c = append(c, "endbfchar\n"...)
	}
```

**Full-gate results with that fix applied, measured in this run:** **965 / 0 / 1** — identical to
baseline. **Same fix with `bfcharSectionCap = 2`:** **954 / 11** (the eleven named in AC9). The tree
was restored to `b5eb557` and `git status` confirmed clean before this file was written.

---

## Decisions raised at creation — RULED mid-story, see `folio-mvp-decision-log.md` D-4.2.1–D-4.2.4

**DECISION-1 is RULED: IN scope (D-4.2.1).** Data cells get cell chrome after all, cascaded from
`style` alone. See the Delivery Log's "AC verification" and "File List" for what was built.
**DECISION-2 is RULED (D-4.2.2):** row identity must be recoverable from `Paginate`'s input without
inference; carried via `isDataRow`/`rowIndex` (rects) and `isTableRowLine`/`rowIndex` (runs) on this
story's own caller-owned structs, asserted for CONTENT correctness (not pagination) in
`TestDataRowIdentityIsConsistentAndDistinct`, with Story 4.3 named as the consumer.
**DECISION-3 is RULED (D-4.2.3):** AC8's pin stands, with its test's own doc comment stating it
records current behaviour without endorsing it. The sections below are kept as originally written
(the creator's framing, and the reasoning that led to each ruling) rather than rewritten after the
fact.

---

### DECISION-1 (needs a ruling before AC1 is implemented) — do data rows get cell rects?

**The gap.** FR24 ("cell borders and cell padding") was 4.1's, and 4.1 fenced itself to the **header
row**. No story in Epic 4 names data-row cell **borders** or **background**: 4.8 owns
`altRowBackground` only, which is a bare colour string and not a `Style`.

**Why it cannot just be skipped.** 4.1's own reasoning for always emitting a header rect is
D-2.6.5's rule — *"an item that occupies space must not be empty"*. A data row occupies space. And
4.1's Finding 14 anticipated this precisely: *"A row implementation that emitted cell text without
cell rects — which is precisely what 4.2 will do first — would leave this fence green."*

**Creator's recommendation:** yes — one `pagemodel.Rect` per data cell, driven by the table's own
`style` (never `headerStyle`, per AC5), carrying `HasFill == false` / `HasStroke == false` when
unstyled. Grounds: it is the header path reused verbatim, it costs nothing for an unstyled table
(4.1 measured that a style-less rect draws nothing), and without it a bordered table renders a boxed
header floating above unruled text, which is not a table. **Flagged rather than assumed** because
it extends this story past its four epic ACs. If the lead trims it, AC1's rect-count assertions drop
to the header's count and AC5(ii) loses its data-cell fill assertion — say so in the Delivery Log.

### DECISION-2 (informational; the ruling belongs to 4.3) — where does row atomicity live?

Recorded in divergence **D2** and AC7. `layout.Paginate` **enforces** single-kind items
(`MixedItemError`), and a wrapped row is several items at several extents, so there is no atomic row
today. 4.2 deliberately keeps 4.1's two-items-same-extent shape and asserts nothing about
pagination. **4.3 must choose** between a fourth exclusivity case, relaxing `MixedItemError`, and a
grouping key — and `OverflowError.Kind`, derived as `"image"` when the item carries images and
`"line"` otherwise (`paginate.go:366-370`), needs a fourth answer under the first option.

### DECISION-3 (surface, do not take) — interleave the within-band diagnostic ordering?

4.1's Finding 7 measured that within one band all text diagnostics precede all table diagnostics,
contradicting D-2.8.6's stated "declaration order within a band". AC8 **pins the current behaviour
in a named test** so a future change is deliberate. Whether the assembly should be interleaved by
declaration order (matching the stated rule) or D-2.8.6 relaxed to match the code is a behaviour
decision that belongs to the lead. **Do not change it in this story.**

### DECISION-4 (already settled elsewhere; recorded so it is not re-litigated)

The "no resolvable `fontFamily`" **diagnostic code** stays with the owner (obligation 3 above).
`TableGeometry.Width()` overflow stays bundled with 4.1's DECISION-3 (obligation 4).

---

## Things the schema and the record could not resolve, surfaced rather than invented

1. **There is no way to express a per-row or per-cell style.** `Column` carries `align` only.
   `altRowBackground` is a bare colour. So a data cell's border/background/padding/font can come
   from exactly one place: the table's own `style`. If the golden report needs, say, a right-aligned
   numeric column with a different background, the schema cannot say it. **Do not add a field.**
   Recorded for 4.7 / Epic 6 to raise if the golden report actually needs it.
2. **There is no `rowHeight`.** Row height is therefore wholly derived (R3). That is the right
   answer for AC2 ("the row grows taller"), but it means a template author has **no** way to pin a
   uniform row height, and a one-line row and a three-line row will differ. If uniform rows turn out
   to be a golden-report requirement, that is a schema question for the owner, not a 4.2 fix.
3. **`testdata/template/golden/worked-example.json` declares no `style.fontFamily`** on its table
   while carrying non-empty column labels. Harmless today (it is a round-trip fixture and is never
   rendered), but under D-4.1.1 it is not a renderable document. Flagged for **4.7**, which will
   want a renderable golden report and may well start from this file.
4. **A table declared in the `pageHeader`/`pageFooter` band with a non-empty collection** will now
   generate rows into a band that is repeated **verbatim** on every page (D-2.6's `BandContent`).
   That is coherent — the same rows on every page — and
   `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` (`table_render_test.go:831`) already exists
   and must stay green. Confirm it does; do not "fix" it into an error.

---

## Task breakdown

1. [x] **Restore the AC9 measurement fixture verbatim** (`internal/pdf/tounicode_chunk_test.go`) and
   confirm it reproduces the recorded pre-fix red at `b5eb557`. If it does not, the baseline has
   moved and the evidence needs re-taking before anything else.
2. [x] Apply the DW-14 chunking to `buildToUnicodeCMap`; confirm the fixture goes green and the full
   gate returns **965 / 0 / 1** unchanged. Record the per-fixture section table (D1) in both
   directions.
3. [x] Re-point the nine `columns[].bind` sites in `table_render_test.go` from `item.` to `row.` (D3),
   as a **separate, clearly-labelled** step so the diff does not read as a behaviour change.
4. [x] Build the body-style cascade (`style` → default, no `headerStyle` arm) and the per-table vertical
   model. **AC5.**
5. [x] Build row generation: iterate the bound collection in index order, build the row scope per row,
   resolve each column's bind, shape and pack each cell at its column's content width, compute the
   row height, place the lines and (per D-4.2.1, ruled: DECISION-1 is IN scope) the cell rects.
   **AC1, AC2, AC3, AC4.**
6. [x] Rewrite `TestTableRendersZeroDataRows` → `TestTableRendersOneRowPerCollectionElementInDataOrder`.
   **AC1.**
7. [x] Strengthen `TestTableStyleFieldsAreNotDataDriven` and its control to a non-empty collection.
   **AC10.**
8. [x] Add the empty-collection, wrapping, clipping, row-scope, header-style-isolation and
   diagnostic-ordering tests. **AC2, AC3, AC4, AC6, AC8.**
9. [x] Run every red-proof **over the whole suite, without `-run`**, and record for each: the command,
   the mutation, which assertions moved, and **which controls did not**. A mutation that reddens a
   control has not isolated its property (4.1 Finding 1).
10. [x] Write the integration/e2e material required, **do not run it**, and say so explicitly (D-000.4).
11. [x] Final gate: `folio-go` 984/0/1, `lint` 115/0, `hashmatrix` 3/0. Confirmed **no golden
    fixture moved**, via `git status`/`git diff --stat` on `fixtures/` (empty) AND via the full
    golden-fixture test suite (each of which calls `Render`/`buildPageModel` for real and compares
    against the committed artifact) staying green — 4.1's Finding 6 precedent (never trust a
    hash-of-committed-files check alone) is honoured by using the render-based goldens as the
    instrument, not a separate file hash.
12. [x] Set the story's status to `review` in `sprint-status.yaml`.

*(There is no commit task. This story ends at `review`.)*

Also completed, beyond the original task list, per the engineering lead's mid-story rulings
(D-4.2.1 – D-4.2.4, `folio-mvp-decision-log.md`):

13. [x] **D-4.2.1 (DECISION-1 ruled IN scope).** Data cells get cell chrome (border/padding/background),
    cascaded from the table's own `style` only (never `headerStyle`) via `resolveBodyStyle` +
    `buildCellRect` (refactored out of `buildHeaderCellRect`'s body so header and row share one
    primitive). `folio-format.md` amended: data cells inherit `style`, and `table.altRowBackground`
    (Story 4.8) is now ranked as winning over `style.background` on the rows it applies to
    — recorded but NOT implemented (4.8's own scope; 4.8 still owns making `altRowBackground`
    non-inert).
14. [x] **D-4.2.2 (DECISION-2, row identity).** `tableRectSource` gained `isDataRow`/`rowIndex`;
    `textRunSource` gained `isTableRowLine`/`rowIndex`. `TestDataRowIdentityIsConsistentAndDistinct`
    asserts the identity's CONTENT (same row -> same identity; different rows -> different
    identities; a wrapped row's several line items all agree) — Story 4.3 is named, in the field's
    own doc comment, as the one consumer. No pagination property is asserted anywhere (AC7 stands).
15. [x] **D-4.2.3 (DECISION-3, label the pin).** `TestWithinBandTableDiagnosticsFollowAllTextDiagnostics`'s
    doc comment states in its own words that it "RECORDS CURRENT BEHAVIOUR. IT DOES NOT ENDORSE IT"
    and names the open interleave-vs-relax question for the lead.
16. [x] **D-4.2.4 (DW-14 addition).** `TestNoRealToUnicodeSectionExceedsTheCap` asserts the cap over the
    REAL shipped fixture corpus (`fixtures/*/expected.pdf`, read directly, not re-rendered), reports
    the observed maximum and fixture/section counts examined (anti-vacuity), and carries its own
    red-proof (`TestNoRealToUnicodeSectionExceedsTheCapRedProof`, a hand-built 101-entry section).

---

## Delivery Log

### Gate results (all three modules, measured in this run)

| Module | Command | Result |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | **984 pass / 0 fail / 1 skip** |
| `folio-go` (determinism, AC11) | same, with `-count=2` | **1967 pass / 0 fail** (doubled minus the permanently-skipped test, itself never doubled) |
| `lint` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 ./...` | **115 pass / 0 fail** (unchanged from baseline — this story adds no new lint surface) |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 ./...` | **3 pass / 0 fail** (unchanged) |
| `gofmt -l` | all three modules | clean |
| `go vet ./...` | all three modules | clean |

The one skip is `TestCorpusMeetsP6ExerciseFloors`, red by design (D-000.17/D-2.1.14/D-000.57/D-000.74)
— untouched by this story.

**No golden fixture moved.** `git status`/`git diff --stat -- fixtures/` show zero changes under
`fixtures/`, AND every golden-fixture test that calls `Render`/`buildPageModel` for real
(`TestRenderMatchesFontTextGoldenFixture`, `TestMultiPageGoldenMatchesTheCommittedArtifact`,
`TestMultiScriptFallbackGoldenFixture`, `TestShapedTextGoldenFixture`,
`TestThreeBandPageGoldenFixture`, `TestWrappedTextGoldenFixture`,
`TestFormatLocaleGoldensAllFourLocales` and its two end-to-end companions) passed in the 984-pass
run above.

### Heavy-test cadence (D-000.4)

**This story is NOT one of D-000.4's per-story matrix overrides** (only 1.2, 1.5, 1.8, 2.4 and 4.7
are). Integration/e2e material this story ships — every golden-fixture test named above, plus the
new row/wrap/clip/identity tests in `table_render_row_test.go` — is written and included in the
unit gate above, but the CROSS-TARGET hash matrix (linux/amd64, linux/arm64 under Docker, js/wasm
under Node, against darwin/arm64) is **NOT run by this story**. It is deferred to the Epic 4
boundary gate, per D-000.4. **Named explicitly, as required:** the unrun suites are the four-target
matrix legs for every `matrixDocuments` entry this story's changes touch (all of them, since
`buildToUnicodeCMap` is shared by every text-bearing document) — none of Epic 4's per-story
overrides is this story.

### D1 — per-fixture `/ToUnicode` section sizes, measured in BOTH directions (D-000.49)

Measured against the CURRENT tree (after DW-14's chunking landed), via
`TestNoRealToUnicodeSectionExceedsTheCap`'s own corpus walk:

| Fixture | Sections (declared==counted, unchanged from creation-time measurement) |
|---|---|
| `font-text` | `[25]` |
| `multi-page` | `[45]` |
| `multi-script-fallback` | `[4, 1, 1]` |
| `page-count-20` | `[32]` |
| `shaped-text` | `[14, 7, 28]` |
| `three-band-page` | `[17]` |
| `wrapped-text` | `[28, 18, 38]` |

Observed maximum across the whole corpus: **45** (on the two-page `multi-page` fixture).
**`page-count-20` carries 32 entries — fewer than two-page `multi-page`'s 45 — so its green is NOT
headroom for 4.7's varied, larger data set.** `minimal-rect` and `image-embed` carry no
`/ToUnicode` section at all (fontless / image-only), consistent with D1's own measurement.
Reproduced live: `TestNoRealToUnicodeSectionExceedsTheCap` logs "examined 9 fixture(s), 7 carrying
at least one /ToUnicode section, 13 section(s) total; observed maximum section size = 45 (cap
100)".

### Red-proofs run (over the WHOLE suite, never behind a narrow `-run`), each restored after measurement

| # | AC | Mutation | Command | Result |
|---|---|---|---|---|
| 1 | DW-14/AC9 | `bfcharSectionCap` 100 -> 2 | full `folio-go` suite | **972 pass / 12 fail** (of 984 total) — exactly the 11 named golden/CID tests reddened: `TestRenderMatchesFontTextGoldenFixture`, `TestMultiScriptFallbackGoldenFixture`, `TestMultiPageGoldenMatchesTheCommittedArtifact`, `TestShapedTextFixtureCIDsOriginateInShapedRuns`, `TestShapedTextFixtureToUnicodeRoundTrips`, `TestShapedTextGoldenFixture`, `TestThreeBandPageGoldenFixture`, `TestWrappedTextGoldenFixture`, `TestFormatLocaleGoldensAllFourLocales`, `TestFormatLocaleEndToEndRenderThai`, `TestFormatLocaleEndToEndRenderJapanese` — plus this story's OWN `TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap` (expected: it pins cap=100 by its own literal, so changing the constant necessarily reddens it too — not one of the "11 named"). 12 failed total, matching exactly. Restored; suite returns to 984/0/1. |
| 2 | DW-14 direct witness | Reverted chunking to the original unbounded single-section emitter | `go test -run TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap ./internal/pdf/` | Reproduced the exact recorded pre-fix output: `entries=250 sections=1 declared=[250] counted=[250]`, failing both assertions. Restored. |
| 3 | AC1(a) | `if len(items) > 0` -> `if false && len(items) > 0` (header-only) | full suite (`-run` narrowed to the target test to observe, but the mutation itself reverts the whole row feature) | `TestTableRendersOneRowPerCollectionElementInDataOrder` reddens: "got 2 rects, want 12", "got 2 runs, want 12". Restored. |
| 4 | AC1(b) | Row loop reversed (`rowVal := items[len(items)-1-rowIdx]`, `rowIdx` kept sequential) | full suite | **Exactly one test reddens**: `TestTableRendersOneRowPerCollectionElementInDataOrder` (the ordered SourceText comparison). No other test — including `TestDataRowIdentityIsConsistentAndDistinct`, whose "which row wrapped" assumption was rewritten to be order-independent specifically so this red-proof would isolate cleanly (see its own updated doc comment) — reddened. Restored. |
| 5 | AC3 | `if cr.clip` -> `if false && cr.clip` (clip flag never set) | full suite | **Exactly one test reddens**: `TestUnbreakableCellContentIsClippedNotWidened` ("expected at least one clipped run"). The geometry assertions in the SAME test, and every other test, stayed green — proving clip and geometry are independent. Restored. |
| 6 | AC4(a) | Row scope never attached (`bind.NewScope(data, params)` with no `.WithRow(...)`) | full suite | **22 tests reddened**, every one of them a table test whose column bind names `row.<field>` (now resolving against the document root, which has no such key) — confirming the row-root path depends on the attached scope. Restored. |
| 7 | AC5 | Data-cell style/font cascaded through `resolveHeaderStyle`'s own chain instead of `resolveBodyStyle`'s | full suite | **Exactly one test reddens**: `TestDataCellsDoNotInheritHeaderStyle` (part (i): the font-error stopped firing, since `headerStyle.fontFamily` now satisfies the body path too). Restored. |
| 8 | AC6(a) | Row loop changed to `for rowIdx := 0; rowIdx <= len(items); rowIdx++` (one extra row emitted for an empty collection) and the `len(items) > 0` gate replaced with `if true` | full suite | 38 tests reddened, including the target `TestEmptyCollectionRendersHeaderOnly` and every other table test that renders an empty collection without a resolvable body font (a direct consequence of removing the `len(items) > 0` gate, which is a stronger mutation than the AC's own suggested form). Confirms the "zero rows for an empty collection" invariant is load-bearing broadly, not merely locally. Restored. AC6(b) ("skip the whole table when empty") is not independently exercised by a live mutation; it is guaranteed by construction — the header rect/run emission code executes unconditionally, entirely before the `if len(items) > 0` block this story adds, so no code path exists that could make an empty collection suppress the header. |
| 9 | AC8/DECISION-3 | `contentDiags`/`contentTableDiags` append order swapped | full suite | **Exactly one test reddens**: `TestWithinBandTableDiagnosticsFollowAllTextDiagnostics` (both assertions: `diags[0]`/`diags[1]` element ids swap). Restored. |

Every mutation was applied to a saved copy, the full suite run, the result recorded above, and the
file restored via `cp` from the saved copy (never `git checkout`, per this agent's own standing
practice) before the next step — `git status` / `git diff --stat` confirm the working tree carries
only the intended production and test changes after every restoration.

### D3 — the nine `columns[].bind` fixture corrections (table_render_test.go), plus one more found live

The nine sites D3 tabulated (`item.a`/`item.b`/`item.c` -> `row.a`/`row.b`/`row.c`, across
`twoColumnsNoAlign`, the two duplicate literal column sets at lines 226-227/335-336, and
`threeColumnTableDoc`'s three-column set) were re-pointed via a single `sed 's/item\./row./g'` pass
over `table_render_test.go`, confirmed to touch exactly 9 lines (a 10th match, `{{row.n}}`, was
already correct and untouched). **These are fixture corrections, not behaviour changes** — the
binds were never evaluated before this story (D-1.6.8's fence), so no prior test's PASS depended
on their old, wrong spelling.

**A tenth site, not in D3's table, was found live**: `render_bind_test.go`'s
`TestRenderScopeFenceIgnoresTableBind` declared no `"as"` on its table while its column bound
`{{formatNumber(transaction.amount, ...)}}` — under the default alias `"row"`, `transaction.` was
about to resolve against the (non-existent) document root and fail. Corrected by adding
`"as": "transaction"` to the table (matching the column's own, unchanged bind). A SECOND, related
correction was needed in the same test: its `amount` field was `"1.00"` (a JSON string), and
`formatNumber` never coerces (AD-14) — changed to the JSON number `1.00`. Both are fixture
corrections surfaced by evaluating column binds for the first time, not behaviour changes; recorded
here rather than silently folded into the "nine" so the count is not misrepresented.

### AC verification (how each is proven)

- **AC1** — `TestTableRendersOneRowPerCollectionElementInDataOrder` (rewrite of `TestTableRendersZeroDataRows`): 5 distinct rows, run/rect counts, and an ORDERED `SourceText` sequence comparison. Red-proofs 3 and 4 above.
- **AC2** — `TestWrappedCellGrowsTheRowAndNeverTheColumn`: 3 columns, all three shaping paths (Latin/Thai/CJK), test-owned `{X,W}` geometry literals identical across narrow/wide renders, an exact `(linesWide-1)*vm.Advance` height-delta assertion (not `>`), and a glyph-count control.
- **AC3** — `TestUnbreakableCellContentIsClippedNotWidened`: one `Warning` (`DiagCodeTextClippedWidth`), located by the COLUMN id, `ClipToBox`/`ClipWidth` on the affected runs, geometry unchanged. Red-proof 5.
- **AC4** — `TestCellBindingsResolveInRowScope` (row/root/params in one render, deliberately DIFFERENT root vs row values), `TestCellBindingResolvesUnderDeclaredAlias` (declared `"as"`), `TestCellBindingAbsentNullWrongKind` (the AD-14 triple, each asserted individually). Red-proof 6.
- **AC5** — `TestDataCellsDoNotInheritHeaderStyle`: (i) a `headerStyle.fontFamily`-only table errors on data cells with fontChain's OWN message (no third spelling); (ii) `headerStyle`/`style` set DIFFERENT backgrounds/aligns and header vs. data cells are asserted to carry their OWN, not the other's. Red-proof 7.
- **AC6** — `TestEmptyCollectionRendersHeaderOnly`: rect count == columns, run count == columns with column-label `SourceText`, zero diagnostics, PLUS a differential check that a one-item collection produces strictly more rects/runs. Red-proof 8.
- **AC7** — Every new test in this story logs its observed page count and asserts it is exactly 1 (R8). `TestDataRowIdentityIsConsistentAndDistinct` proves the row identity itself; it asserts NOTHING about pagination. No multi-page table fixture was added.
- **AC8** — `TestWithinBandTableDiagnosticsFollowAllTextDiagnostics`, whose OWN doc comment states it "RECORDS CURRENT BEHAVIOUR. IT DOES NOT ENDORSE IT" (D-4.2.3). Red-proof 9.
- **AC9** — see the DW-14 section above and red-proofs 1-2.
- **AC10** — `TestTableStyleFieldsAreNotDataDriven`/`...Control`, both changed to a non-empty, identical-in-both-renders `items` collection with real row binds.
- **AC11** — `-count=2` full-suite run (see Gate results); `lint`'s `TestFloatTypedProductionScan` (115/0, unchanged) is the float-ban instrument, never `go build`; row iteration is `for rowIdx, rowVal := range items` over a slice, never a map.

### Things NOT built (per the story's own fences, unaffected by the mid-story rulings)

- `layout.Paginate` itself is untouched: no new exclusivity case, no relaxation of `MixedItemError`,
  no grouping mechanism inside `internal/layout`. Row identity (D-4.2.2) is carried entirely in
  package `folio`'s own caller-owned structs (`tableRectSource`, `textRunSource`), which is also
  why it survives `Paginate`'s own opaque output (a `TextRunRef`/`RectRef` index is stable across
  the call, so a consumer holding one can look the identity up directly) — a stronger guarantee
  than attaching it to `layout.ColumnItem`, which `Paginate` does not hand back.
- `columns[].footer`/`footerOf`/`footerFormat` remain unread at render (Story 4.5's).
- `altRowBackground` remains inert at render (Story 4.8's) — only its PRECEDENCE against
  `style.background` is now recorded (D-4.2.1).
- No new diagnostic code was minted (D-000.65): `DiagCodeTextClippedWidth` and
  `DiagCodeBindingPathAbsent` are both reused verbatim for the data-cell cases.
- The "no resolvable fontFamily" diagnostic CODE question stays with its existing owner — this
  story's data-cell failures route through the same plain-wrapped, non-`*RenderError` shape the
  header's font-chain failure already uses (via `fontChain` reused verbatim), minting nothing.

### Finisher fixes and re-measurement (post-review, D-000.49: append, not rewrite)

The review's verdict was **Changes Requested: 3 Blockers, 4 Majors, 11 Minors, 4 Nits**, all 21
plus the 3 blockers triaged and resolved below (see **## Finding Resolutions** at the end of this
file for the per-finding table). Every code fix was made surgically, in the file(s) the finding
named. **No golden fixture moved** (re-confirmed: `git status`/`git diff --stat -- fixtures/` empty
after every fix).

**All three Blockers were the SAME defect class the reviewer named (D-000.79: a guard that cannot
fail) and are now fixed with genuine, mutation-confirmed instruments — none was argued away, and
none required re-opening D-4.2.1/D-4.2.2's rulings.**

- **Blocker 1 (Finding 1) — data-row border was inert.** New test
  `TestDataRowBorderIsDrawn` (`table_render_row_test.go`) renders a non-empty collection with
  `style.border` declared and asserts the data row's rects carry `HasStroke`, the declared
  `Stroke`/`StrokeWidth`, and the full four-edge default `Edges`. **Red-proof M1 reproduced**:
  forcing the data-row `buildCellRect` call's border arguments to `(false, template.Border{})`
  left the full suite at **985/0/1 unchanged** before the fix existed; after adding the test, the
  SAME mutation reddens exactly `TestDataRowBorderIsDrawn` and nothing else. Restored,
  `/usr/bin/diff` confirmed byte-identical.
- **Blocker 2 (Finding 2) — data-cell padding was inert.** New DIFFERENTIAL test
  `TestDataCellPaddingShiftsRowHeightAndContentOrigin` (`table_render_row_test.go`): two renders
  of the same template varying only `style.padding`, asserting the padded render's data-row rect
  `H` exceeds the unpadded render's by EXACTLY `padding.top+padding.bottom` and the data cell's
  run `X` shifts by EXACTLY `padding.left` (test-owned literals, D-000.68 — no font-metric
  re-derivation needed since a single-line cell's own vertical contribution cancels out of the
  delta). **Red-proof M3 reproduced**: zeroing all four resolved padding edges left the full suite
  green before the fix; the same mutation now reddens exactly this one test. Restored, diffed
  clean.
- **Blocker 3 (Finding 3) — row identity's defining property (a wrapped row's lines all
  agreeing with the row whose DATA they render) was unasserted.** Rewrote
  `TestDataRowIdentityIsConsistentAndDistinct`: every row's cell values now carry a PER-ROW MARKER
  substring, and the test asserts, for every table-row-line run, that (a) the marker its own
  `SourceText` carries names the row it claims via `rowIndex`, including every continuation line
  of the wrapped row, and (b) that `rowIndex` matches the `tableRectSource` group whose vertical
  extent physically contains the line — content and geometric identity are cross-checked, not
  merely a distinct-index count. **Red-proof M2b reproduced**: making a wrapped row's continuation
  lines (`li>0`) claim `rowIdx+1` (the next EXISTING row) left the full suite at 985/0/1 before
  this fix (confirmed independently, over the whole suite, no `-run`); after the fix, the SAME
  mutation reddens exactly `TestDataRowIdentityIsConsistentAndDistinct` — every continuation line
  of row 0 is reported as claiming row 1 while its own marker still says row 0 — and nothing else.
  Restored, diffed clean.

**The four Majors:**

- **Finding 4 — `resolvedBodyStyle.valign` resolved and never read.** WIRED, not stripped: the
  row-generation loop now computes, per column, how many lines of "slack" that column's own cell
  has relative to the row's tallest cell (`linesInRow` minus the cell's own line count), and
  distributes it per `bs.valign` ("top" leaves a short cell's lines at the row's first physical
  lines — the only behaviour before this fix — "bottom" moves them to the row's last lines,
  "middle" splits the difference by integer line count, never a sub-line fraction, so no binary
  float is introduced, AD-1 holds). This required no change to the row's own height or to any
  column's geometry (AD-13 untouched) — only to which physical line SLOT a shorter cell's own
  content occupies. New test `TestDataCellValignDistributesRowSlack`
  (`table_render_row_test.go`) proves it: a wrapping column's own observed per-line Y positions
  (which valign never moves, since that column has zero slack) become this test's own anchor, and
  a short sibling column's Y is asserted against specific entries in that same observed list for
  "top"/"middle"/"bottom". **Red-proof reproduced**: disabling the per-cell offset reddens exactly
  this one test. AC5's own mechanism text, the code comment, and `folio-format.md`'s statement all
  already correctly claimed this cascade — they needed no correction, only the code catching up to
  them.
- **Finding 5 — AD-14's `null`/wrong-kind subtests under-asserted.** `explicit_null` now renders
  through `buildPageModel` against a table with a SECOND, always-bound sibling column and asserts
  a positive, differential run count (2 header labels + 1 sibling data run — the null cell
  contributes none) plus the sibling's own text is present. `wrong_kind` now asserts the SAME
  rigour as `absent`: a `*RenderError`, its diagnostic code, and the column id — confirmed the
  production code already reuses `DiagCodeBindingPathAbsent` for a wrong-kind failure (D-000.65:
  the same code as `absent`, no third spelling), so the assertion states an existing, correct
  fact rather than requiring a code change.
- **Finding 6 — AC6(b)'s mandated red-proof never run; red-proof #3 recorded under a narrow
  `-run`.** Both re-run for real, over the WHOLE suite, no `-run`, and recorded honestly below
  (this subsection and the reproduced-mutations note above) rather than only asserted by
  reasoning.
- **Finding 7 — `TestNoRealToUnicodeSectionExceedsTheCapRedProof` never exercised the test it was
  named for.** Extracted `walkToUnicodeCorpus(fixturesDir string)` — the ONE walk both the real
  corpus test (over `fixtures/`) and its red-proof (over a `t.TempDir()` carrying one hand-built
  101-entry `expected.pdf`) now call. The red-proof asserts the walk itself reports a violation,
  reproducing the reviewer's manual M4 probe as a standing, permanent test rather than a one-off.

**The eleven Minors and four Nits** — full disposition in **## Finding Resolutions**; the
highlights: Findings 8/9/10 corrected `internal/pdf/textdoc.go`'s doc-comment attachment (the
`bfcharSectionCap` comment was swallowing `buildToUnicodeCMap`'s own doc comment — reordered so
each attaches to its own declaration), the stale test-name citations, and the PDF-spec citation
(a MANDATE, not guidance). Finding 11 gave `TestToUnicodeSectionCapBoundaryValues` its own
test-owned `capForTest` literal (D-000.68) instead of comparing against the constant under test,
and added a just-under-cap (99-entry) face. Finding 12 extended three more table tests
(`TestColumnGeometryNeverNegotiatesAgainstLabelContent`, `TestColumnAlignWinsOverStyleAlign`,
`TestColumnAlignWinsOverHeaderStyleAlign`) to a non-empty collection identical across their
compared renders — the SAME vacuity class D4 already fixed in two places, now closed in the three
this story also touched but left empty; `TestColumnAlignWinsOverHeaderStyleAlign` additionally now
asserts the AC5/D-000.76 boundary directly (a data cell must NOT inherit `headerStyle.align`).
Finding 13 rewrote `TestRenderScopeFenceIgnoresTableBind`'s doc comment, which asserted "column
binds are not yet evaluated" in one paragraph and "column binds are evaluated for the first time"
in another. Finding 15 extracted a shared `widthClipMessage` builder used by both the text-element
and table-cell clip sites, fixing the table site's message to say "column" (not "element") while
preserving the substantively-different "declared" vs "content" width phrasing. Finding 16
strengthened `TestCellBindingsResolveInRowScope` to assert each sentinel's run X against its own
column's span, not merely that all three sentinels appear somewhere. Finding 17 corrected the
Delivery Log's `-count=2` note (see re-measurement below). Finding 18 extracted
`tableCollectionSegments` so `checkTableBindings` (render.go) and the row-generation loop
(table_render.go) parse a table's collection path through ONE function. Findings 19–21 were dead
duplicate-check removal, an AC citation fix, and combining two identical `collectBandTableRuns`
calls into one.

**Finding 14 (fontChain called before any cell is known to carry text) — DEFERRED**, per the
reviewer's own recommendation ("worth surfacing to the lead rather than deciding here, since it
touches D-4.1.1's boundary"). This is a genuine boundary question about an EXISTING, pre-4.2
failure mode (a text-bearing element with no resolvable `style.fontFamily`), not a defect this
story introduced: the header's own per-column guard (`if col.Label == "" { continue }`) already
has an equivalent asymmetry, unexamined until now. Deciding whether `fontChain` should be called
eagerly (current behaviour, both header and body) or lazily (only once a cell's bound text is
known to be non-empty) is a font-resolution-error-handling policy question shared with
`fontChain`'s existing text-element caller, and belongs with the same owner Obligation 3 already
names ("whoever next touches font-resolution error handling"), not with a table-row story.

**Re-measured red-proofs (Findings 6 and 17), full suite, no `-run`:**

| # | Finding/AC | Mutation | Command | Result |
|---|---|---|---|---|
| R1 | Finding 6 / AC1(a) | `if len(items) > 0` → `if false && len(items) > 0` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -v ./...` (full suite, no `-run`) | **16 top-level tests reddened** (every table test with a non-empty collection), confirming the same set found by the story's own original red-proof — this time genuinely observed over the whole suite rather than under a narrow `-run`. Restored, diffed clean. |
| R2 | Finding 6 / AC6(b) | Skip the WHOLE table (header included) when the collection is empty (truncate `runs`/`rectSources` back to their pre-header marks and `continue`) | same, full suite, no `-run` | **7 top-level tests reddened**, including the target `TestEmptyCollectionRendersHeaderOnly` — and specifically its HEADER-PRESENT assertions (`got 0 rects, want 2`, `got 0 runs, want 2`), a DIFFERENT assertion from (a)'s (which moves the ROW counts upward, never toward zero). Every other reddened test is a header-only fixture (`{"items": []}`) whose header the mutation now also suppresses — consistent with the mutation's own description. Restored, diffed clean. |

**Finding 17, corrected**: the Delivery Log's original `-count=2` parenthetical ("doubled minus the
permanently-skipped test, itself never doubled") was itself wrong — the skip DOES double under
`-count=2`. Re-measured on the finished tree: `-count=1` is **985 pass / 0 fail / 1 skip**;
`-count=2` is **1969 pass / 0 fail / 2 skip**. The skip doubling correctly to 2 is the discriminating
fact AC11 needs (both runs of the permanently-red-by-design test are still counted, never silently
dropped). The apparent "off by one" in a naive `2×985−1` check is a counting artifact of Go's own
subtest-name disambiguation on a few table-driven tests sharing subtest names across two runs
(e.g. `TestShapedExpectationsMatchArtifact`'s per-glyph subtests) — not a real skip or duplicate —
so this Delivery Log states the two measured totals plainly rather than asserting a derivation
that does not reliably hold across runs with shared subtest names.

**Final three-module gate, measured independently by the finisher after every fix and every
mutation above was restored:**

| Module | Command | Result |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -v ./...` | **985 pass / 0 fail / 1 skip** |
| `folio-go` (determinism, AC11) | same, `-count=2` | **1969 pass / 0 fail / 2 skip** |
| `lint` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **115 pass / 0 fail** |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **3 pass / 0 fail** |
| `gofmt -l` | all three modules | clean |
| `go vet ./...` | all three modules | clean |

The one skip is `TestXrefEntriesRejectsMalformedSubprocess` (unrelated to this story, present at
the reviewer's own measurement too); `TestCorpusMeetsP6ExerciseFloors` is excluded by `-skip` and
is red by design (D-000.17/D-2.1.14/D-000.57/D-000.74) — untouched.

**No golden fixture moved**: `git status`/`git diff --stat -- fixtures/` show zero changes, both
before and after every fix in this subsection.

**Heavy tests (D-000.4), read for teeth, not run — stated explicitly, as required.** The
cross-target matrix (linux/amd64, linux/arm64 under Docker, js/wasm under Node vs darwin/arm64) was
**not run** by this finishing pass either — 4.2 is not on D-000.4's per-story override list, and
nothing in this finishing pass changes that. The new/strengthened tests this pass adds
(`TestDataRowBorderIsDrawn`, `TestDataCellPaddingShiftsRowHeightAndContentOrigin`,
`TestDataCellValignDistributesRowSlack`, the rewritten `TestDataRowIdentityIsConsistentAndDistinct`,
the extracted `walkToUnicodeCorpus`'s red-proof) are all included in the unit gate above and were
each confirmed against a live, restored mutation.

## File List

- `folio-go/render.go` — `textRunSource` gained `isTableRowLine`/`rowIndex` (D-4.2.2); `collectBandTableRuns`'s three call sites in `predictDocument` now pass `data`/`params`/`fc`. **Finisher**: extracted `tableCollectionSegments` (shared with `table_render.go` — Finding 18) and `widthClipMessage` (shared clip-warning builder — Finding 15); both call sites now use them.
- `folio-go/table_render.go` — row generation added to `collectBandTableRuns` (AC1-AC5, AC7's row-line/row-rect shape); `resolvedBodyStyle`/`resolveBodyStyle` (AC5); `buildCellRect` factored out of `buildHeaderCellRect` (shared header/row rect primitive); `tableRectSource` gained `isDataRow`/`rowIndex` (D-4.2.2). **Finisher**: wired `bs.valign` into the row loop via a per-column line offset (Finding 4/Blocker-class); the clip-warning and collection-path-parsing sites now call the shared helpers above (Findings 15/18); `strings` import dropped (no longer needed directly).
- `folio-go/internal/pdf/textdoc.go` — `buildToUnicodeCMap` chunks `/ToUnicode` `beginbfchar` sections at the new `bfcharSectionCap = 100` constant (DW-14/AC9), preserving the empty-table edge case byte-for-byte. **Finisher**: reordered `bfcharSectionCap`'s and `buildToUnicodeCMap`'s doc comments so each attaches to its own declaration (Finding 8); corrected the stale `TestNoToUnicodeSectionAnywhereExceedsTheCap` citation to `TestNoRealToUnicodeSectionExceedsTheCap` (Finding 9); restated the PDF-spec citation as a requirement, not guidance (Finding 10).
- `folio-go/internal/pdf/tounicode_chunk_test.go` — NEW (developer). The restored DW-14 measurement fixture (`TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap`) plus `TestToUnicodeSectionCapBoundaryValues` (synthetic boundary values). **Finisher**: `TestToUnicodeSectionCapBoundaryValues` now anchors to its own test-owned `capForTest` literal instead of `bfcharSectionCap` (Finding 11), added a 99-entry (just-under-cap) face, and corrected its own stale test-name citation (Finding 9).
- `folio-go/tounicode_corpus_test.go` — NEW (developer). `TestNoRealToUnicodeSectionExceedsTheCap` (D-4.2.4: the real-corpus property, anti-vacuity logging) and its red-proof. **Finisher**: extracted `walkToUnicodeCorpus` so the red-proof (now over a `t.TempDir()` fixture) exercises the SAME walk the real test does, rather than only a shared size-extraction helper (Finding 7).
- `folio-go/table_render_row_test.go` — NEW (developer). AC2, AC3, AC4, AC5, AC6, AC7 (identity), AC8 instruments. **Finisher**: added `TestDataRowBorderIsDrawn` and `TestDataCellPaddingShiftsRowHeightAndContentOrigin` (Blockers 1/2), `TestDataCellValignDistributesRowSlack` (Finding 4); rewrote `TestDataRowIdentityIsConsistentAndDistinct` to assert content identity, not only a distinct-index count (Blocker 3), combining its two `collectBandTableRuns` calls into one (Finding 21); strengthened `TestCellBindingAbsentNullWrongKind`'s `explicit_null`/`wrong_kind` subtests (Finding 5) and `TestCellBindingsResolveInRowScope`'s column attribution (Finding 16).
- `folio-go/table_render_test.go` — nine `item.` -> `row.` bind corrections (D3); `TestTableRendersZeroDataRows` rewritten to `TestTableRendersOneRowPerCollectionElementInDataOrder` (AC1); `TestTableStyleFieldsAreNotDataDriven`/`...Control` strengthened to non-empty, identical collections (AC10). **Finisher**: gave `TestColumnGeometryNeverNegotiatesAgainstLabelContent`, `TestColumnAlignWinsOverStyleAlign` and `TestColumnAlignWinsOverHeaderStyleAlign` a non-empty collection too (Finding 12); dropped a dead duplicate run-count check (Finding 19); corrected a failure message's AC citation (Finding 20).
- `folio-go/render_bind_test.go` — `TestRenderScopeFenceIgnoresTableBind`'s fixture corrected (`"as": "transaction"` added; `amount` changed from a string to a number) — a tenth fixture correction found live, not in D3's table. **Finisher**: rewrote the test's own doc comment, which contradicted itself on whether column binds are evaluated (Finding 13).
- `_bmad-output/specs/spec-folio/folio-format.md` — documents that a data cell cascades from `style` alone (never `headerStyle`), and ranks `altRowBackground` over `style.background` (D-4.2.1). No finisher change: its `valign` claim is now true (Finding 4 was fixed by wiring the code, not by correcting the document).
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — status: `review` → **finisher: `done`**.

## Change Log

| Date | Change |
|---|---|
| 2026-08-26 | Story created at `b5eb557`. DW-14 measured, chunked, gated and reverted in-run. |
| 2026-08-26 | Engineering lead rulings applied mid-story: D-4.2.1 (data-cell chrome IN scope), D-4.2.2 (row identity carried via caller-owned structs, Story 4.3 named as consumer), D-4.2.3 (AC8's pin labelled as non-endorsing), D-4.2.4 (DW-14 corpus-wide property + red-proof added). |
| 2026-08-26 | Implementation complete: row generation, wrapping/clipping, row-scope binding, data-cell style cascade, DW-14 chunking, all ACs verified with red-proofs over the whole suite. Gate: `folio-go` 984/0/1, `lint` 115/0, `hashmatrix` 3/0. Status set to `review`. |
| 2026-08-26 | Reviewed (bmad-code-reviewer). Changes Requested: 3 Blockers, 4 Majors, 11 Minors, 4 Nits. All three Blockers were the same defect class (D-000.79: a guard that cannot fail) — data-row border, data-cell padding, and row identity's wrapped-line agreement were each provably inert to the entire three-module gate. See QA Results below. |
| 2026-08-26 | Finished (bmad-story-finisher). All 3 Blockers, 4 Majors, 11 Minors and 4 Nits triaged; 20 FIX, 1 DEFER (Finding 14, a pre-existing font-resolution-error-handling boundary question, owned by whoever next touches that area — Obligation 3). Every Blocker fixed with a genuine, mutation-confirmed instrument (`TestDataRowBorderIsDrawn`, `TestDataCellPaddingShiftsRowHeightAndContentOrigin`, and a content-plus-geometry rewrite of `TestDataRowIdentityIsConsistentAndDistinct`), each reproduced against the reviewer's own M1/M3/M2b. `resolvedBodyStyle.valign` (Finding 4) was WIRED into the row loop rather than stripped from the documentation, since AC5/the code comment/`folio-format.md` were already correct about the intended behaviour. `walkToUnicodeCorpus` was extracted so its red-proof (Finding 7) genuinely exercises the real corpus test. AC6(b)'s red-proof and AC1(a)'s narrow-`-run` red-proof (Finding 6) were both re-run over the whole suite and recorded. Two shared helpers (`tableCollectionSegments`, `widthClipMessage`) removed duplicated logic (Findings 15/18). Gate re-measured independently: `folio-go` 985/0/1 (`-count=2`: 1969/0/2), `lint` 115/0, `hashmatrix` 3/0 — all three `go build`/`go vet`/`gofmt -l` clean. No golden fixture moved. See "Finisher fixes and re-measurement" in the Delivery Log and "## Finding Resolutions" below. Status set to `done`. |

## Review Summary

- **Reviewed by:** bmad-code-reviewer (spawned fresh; no developer context)
- **Date:** 2026-08-26
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 3
- **Majors:** 4
- **Minors:** 11
- **Nits:** 4

### Gate, measured independently by the reviewer at the working tree

| Gate | Result measured | Story's claim | Verdict |
|---|---|---|---|
| `folio-go` `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | **984 pass / 0 fail / 1 skip** | 984/0/1 | matches |
| `folio-go` `-count=2` (AC11 determinism) | **1967 pass / 0 fail / 2 skip** | 1967/0 | headline matches (see Minor 11) |
| `lint` `go test -count=1 ./...` | **115 pass / 0 fail** | 115/0 | matches |
| `hashmatrix` `go test -count=1 ./...` | **3 pass / 0 fail** | 3/0 | matches |
| `gofmt -l` (bare binary, all three modules) | clean | clean | matches |
| `go vet ./...` (all three modules) | clean | clean | matches |

The one skip is `TestXrefEntriesRejectsMalformedSubprocess`; `TestCorpusMeetsP6ExerciseFloors` is
excluded by `-skip` and is red by design (D-000.17/D-2.1.14/D-000.57/D-000.74). Neither is a finding.

**Standing guards re-run individually, all green:** `lint`'s `TestGlyphIdentifierCensus`,
`TestSpineStageLadderMatchesStageRankTable`, `TestFloatTypedProductionScan`; `folio-go`'s
`TestFolioMethodNamesAreInjective`, `TestValidateNeverReachesRenderOrInternalPDF`. `git grep`
confirms this story adds **callers** of `shapeSegments`/`positionSegments` and **no** new
`pagemodel.ShapedGlyph` construction site (R4 honoured).

**No golden fixture moved — verified with evidence that could detect a move.** `git status` on
`fixtures/` is empty, but that alone is the weak check 4.1's Finding 6 warned about. The decisive
evidence is that the render-based goldens genuinely observe CMap bytes: the reviewer set
`bfcharSectionCap = 2` and the suite went to **972 / 12**, reddening exactly the eleven golden and
CID tests the Delivery Log names, plus this story's own `TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap`.
The Delivery Log's red-proof #1 reproduces exactly. Those goldens therefore re-render for real, and
their green at cap 100 is a positive statement of byte-neutrality, not an unfailable one.

**Heavy tests (D-000.4), read for teeth rather than run — stated explicitly as required.** The
cross-target matrix (linux/amd64, linux/arm64 under Docker, js/wasm under Node vs darwin/arm64) was
**not run**, correctly: 4.2 is not on D-000.4's per-story override list. Reading the material this
story ships: the golden-fixture tests have real teeth (proved above by the cap-2 mutation);
`table_render_row_test.go`'s new tests have real teeth for geometry, wrapping, clipping, row scope
and background (each verified below by mutation) but **no teeth at all for data-cell border,
data-cell padding, or wrapped-row identity agreement** — see Blockers 1–3.

### Mutations the reviewer applied and re-ran (whole suite, never behind a narrow `-run`)

Every mutation was applied to a saved pristine copy, run, then reverted by `cp` (never
`git checkout`) and re-verified by SHA-256. Final restored gate: **984 / 0 / 1**, tree byte-identical
to the state at review start.

| # | Mutation | Result |
|---|---|---|
| M1 | Data-row `buildCellRect` border half forced to `false, template.Border{}` | **984 / 0 — nothing reddened.** → Blocker 1 |
| M2 | `placed[j].rowIndex = rowIdx + li` (wrapped-row lines disagree) | 983 / 1 — reddened, but on the *distinct-index count* (`got 5 distinct rowIndex values, want 3`), not on agreement |
| M2b | Wrapped-row continuation lines claim the **next existing** row's index | **984 / 0 — nothing reddened.** → Blocker 3 |
| M3 | Data-cell padding zeroed (`padTopB/RightB/BottomB/LeftB = 0`) | **984 / 0 — nothing reddened.** → Blocker 2 |
| M3b | Data-cell background half forced to `false, ""` | 983 / 1 — `TestDataCellsDoNotInheritHeaderStyle` reddened. Background **is** guarded (control for M1/M3) |
| M4 | A temporary `fixtures/zz-reviewer-probe/expected.pdf` carrying a 101-entry `beginbfchar` section | **Reddened correctly**, naming the fixture and reporting `observed maximum section size = 101 (cap 100)`. Probe removed; `fixtures/` clean |
| M5 | `bfcharSectionCap` 100 → 2 | 972 / 12, exactly reproducing the Delivery Log's red-proof #1 |

**The corpus property reddens under the reviewer's own forced violation (M4).** The three mandated
DW-14 checks: the direct chunking test asserts on `buildToUnicodeCMap`'s **emitted bytes**, not a
helper's return value (so an emitter that stopped calling the chunker is caught); the cap **is** a
named constant with a citation; and the expected `[100, 100, 50]` split is a test-owned literal, not
derived from the constant. DW-14 is, on substance, correctly landed — the findings against it below
are all documentary.

---

### Finding 1: Data-row cell **border** is inert to the entire three-module gate

- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_render.go:686` (the data-row `buildCellRect` call); absence of any
  witness in `folio-go/table_render_row_test.go` and `folio-go/table_render_test.go`
- **Observation**: DECISION-1 was ruled **IN scope** mid-flight (D-4.2.1): data cells get cell
  chrome — border, padding, background — cascaded from `style` alone. Mutation **M1** replaced the
  border half of the data-row `buildCellRect` call with `false, template.Border{}`, so no data cell
  ever draws a border. The full `folio-go` suite returned **984 pass / 0 fail / 1 skip** — bit-for-bit
  the unmutated figure. Every `HasStroke` / `Edges` / `Stroke` / `StrokeWidth` assertion in the repo
  (`table_render_test.go:120,162,171,174,191,311,314`) renders `{"items": []}` and therefore observes
  the **header row only**. Control M3b proves the harness *can* see data-cell chrome: removing the
  **background** half does redden `TestDataCellsDoNotInheritHeaderStyle`. Border simply has no witness.
- **Impact**: This is 4.1's Blocker 1 shape recurring inside the story that was explicitly written to
  prevent it (AC5's own words: *"`headerStyle` shipped inert: deleting its render handling left the
  entire three-module gate green"*). An owner-ruled feature ships with no guard: it can be reverted,
  half-reverted or broken by a future refactor with the whole gate green. It matters concretely
  because a bordered table is DECISION-1's stated justification (*"without it a bordered table
  renders a boxed header floating above unruled text, which is not a table"*), and **Story 4.7
  freezes the golden report across four targets** — a silently border-less body would be enshrined.
- **Suggested Resolution**: Add a render-level assertion over a **non-empty** collection with
  `style.border` declared: assert the data-row rects carry `HasStroke`, the declared `Stroke` colour,
  `StrokeWidth`, and the full four-edge `Edges` set (the guardrail is *honour the field* — if
  `style.border` is set, data cells get a full grid; Folio must not decide data rows are borderless).
  Record M1 as its red-proof. Do not fix by weakening the ruling.
- **Related AC**: D-4.2.1 / DECISION-1; AC5(ii)

### Finding 2: Data-cell **padding** is inert to the entire three-module gate

- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_render.go:546` (`padTopB, padRightB, padBottomB, padLeftB := paddingEdges(bs.padding)`)
  and its uses at the row-height and `contentX`/`contentW` sites
- **Observation**: Mutation **M3** zeroed all four data-cell padding edges immediately after they are
  resolved. The full suite returned **984 pass / 0 fail / 1 skip** — unchanged. Data-cell padding
  affects both the row's height (`rowHeight = padTopB + … + padBottomB`) and every cell's text
  x-origin and wrap width (`contentX = cg.X + padLeftB`, `contentW = cg.Width - padLeftB - padRightB`),
  yet nothing observes it. `TestTableHeaderPaddingInsetsLabel` (`table_render_test.go:202`) renders
  `{"items": []}`. `TestTableStyleFieldsAreNotDataDriven` does declare `"padding": {"left": 5, "top": 2}`
  over a non-empty collection, but it only compares two renders **to each other**, both with the same
  padding — it cannot see padding being ignored in both.
- **Impact**: Same class as Finding 1, with a wider blast radius: padding is load-bearing for AC2's
  row-height formula (R3) and for AC2/AC3's content width. A regression here silently changes every
  row's height and every cell's wrap point — and would be frozen by 4.7.
- **Suggested Resolution**: Add a differential render-level assertion over a non-empty collection:
  `style.padding.left = 20` must shift the data cell's run X by exactly 20000 millipoints relative to
  an otherwise identical zero-padding render, and `style.padding.top`/`bottom` must grow the data
  row's rect `H` by exactly the declared amount (test-owned literals, per D-000.68). Record M3 as its
  red-proof.
- **Related AC**: D-4.2.1 / DECISION-1; AC2 (row height); AC3 (content width)

### Finding 3: The row-identity property the owner ruling exists for — *a wrapped row's N items all agree* — is not asserted

- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_render_row_test.go:572-668` (`TestDataRowIdentityIsConsistentAndDistinct`);
  fields at `folio-go/table_render.go:61-62`, `folio-go/render.go:174-175`
- **Observation**: D-4.2.2 requires row membership to be *recoverable from `Paginate`'s input without
  inference*, and the story's own task 14 states the test asserts *"a wrapped row's several line items
  all agree"*. It does not. The test builds `rowLineCount[rowIndex] -> set of lineIndex` and asserts
  only three things: `len(rowLineCount) == 3` (a **count** of distinct indices), `maxLines >= 2` (some
  row wrapped), and global `lineIndex` uniqueness. None of these detects a wrapped row's continuation
  lines being attributed to the **wrong existing row**. Mutation **M2b** made exactly that defect —
  a wrapped row's lines `li > 0` claim `rowIdx + 1` — and the full suite returned **984 pass / 0 fail /
  1 skip**, entirely green. (Mutation **M2**, the cruder `rowIdx + li`, *did* redden, but the failure
  message shows it reddened on `got 5 distinct rowIndex values, want 3` — the count, not agreement.)
  The test also never reads `SourceText`, so nothing ties a run's `rowIndex` to the collection element
  whose text it renders, and never cross-checks a run's identity against the rect group whose extent
  contains it.
- **Impact**: `isDataRow` / `isTableRowLine` / `rowIndex` have **no production reader** (verified by
  `git grep`: written in `table_render.go:698,699,744,745`, read only in this one test). Story 4.3 is
  the sole named consumer. So this test is the *entire* protection on the field, and it is blind to
  precisely the failure mode 4.3 will be broken by — the field's own doc comment names it: *"exactly
  the reconstruction the ruling names as where a wrapped row silently becomes two."* An inert field
  guarded by a test that cannot see its defining property is the defect Story 4.1 spent a story fixing
  eight instances of.
- **Suggested Resolution**: Assert **content**, not shape. Bind each row's cells to distinct sentinel
  values, then assert that (a) every run whose `SourceText` belongs to row *k*'s data carries
  `rowIndex == k`, including every continuation line of the wrapped cell; and (b) the run's `rowIndex`
  equals the `rowIndex` of the `tableRectSource` whose `top..bottom` contains that run. Record M2b as
  the red-proof — a mutation that keeps the distinct-index count correct is the one that matters.
- **Related AC**: D-4.2.2 / DECISION-2; AC7

### Finding 4: `resolvedBodyStyle.valign` is resolved and never read — a provably inert field

- **Severity**: Major
- **Category**: Correctness / AC Conformance
- **Location**: `folio-go/table_render.go:187` (field), `:200`, `:211-212` (assignment); no reader
  anywhere. Contrast `hs.valign`, read at `folio-go/table_render.go:471`.
- **Observation**: `resolveBodyStyle` computes `r.valign` ("top"/"middle"/"bottom", commented *"never
  empty"*) and nothing ever reads it — confirmed by `grep -n "bs\." table_render.go`, which returns
  exactly three readers: `bs.padding`, `bs.alignFallback`, `bs.hasBackground/background/hasBorder/border`.
  Data-row lines are unconditionally placed at `rowTop + padTopB + li*vm.Advance`, i.e. always
  top-aligned. This is observable behaviour, not a theoretical gap: within a wrapped row, a one-line
  cell sitting in a three-line row has real vertical slack that `valign: "middle"`/`"bottom"` should
  distribute. Three artifacts claim otherwise: AC5's mechanism text (*"The same rule applies to
  `fontSize`, `padding`, `border`, `background`, `align` and `valign` for data cells: `style` only"*),
  the code comment at `table_render.go:531`, and the new paragraph in `folio-format.md` (*"cascades its
  font, border, background, padding, align and valign from the table's own `style` **only**"*).
- **Impact**: An author-declared style field is silently dropped on data rows while being honoured on
  the header row, and the shipped spec document states the opposite. No mutation is needed to
  demonstrate it — the field has no reader, so it cannot be tested at all. It also means
  `folio-format.md` now ships a false statement about rendered behaviour, which is worse than silence.
- **Suggested Resolution**: Either (a) apply `bs.valign` when placing a cell's lines within the row's
  computed height, with a render-level test over a wrapped row, or (b) if valign is genuinely
  undefined for a derived-height row (R7: there is no `rowHeight` field), **delete the field from
  `resolvedBodyStyle`**, strike `valign` from the AC5 comment at `table_render.go:531`, and correct
  the `folio-format.md` sentence. Do not leave it resolved-and-dropped.
- **Related AC**: AC5

### Finding 5: AD-14's *explicit null = empty cell* and *wrong kind = Error* are only half-asserted

- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_render_row_test.go:344-386` (`TestCellBindingAbsentNullWrongKind`)
- **Observation**: AC4 requires AD-14's three cases *"asserted individually"*. Only the `absent`
  subtest is asserted properly (error is a `*RenderError`, `Code == DiagCodeBindingPathAbsent`,
  `ElementID == "e2"`, the column id). The other two are not:
  - `explicit_null` asserts only `rerr == nil` and `len(res.Bytes) != 0`. It never asserts the cell is
    **empty**. A regression that rendered the literal text `null`, or the string `<nil>`, or fell back
    to the column label, passes this test unchanged. The AC's words are *"explicit JSON `null` → empty
    cell, not an error"*; only the second clause is instrumented.
  - `wrong_kind` asserts only `rerr != nil`. Any error at all satisfies it — a font failure, a
    template failure, an unrelated panic-turned-error. It asserts neither the diagnostic code nor the
    column id, unlike its sibling, and so cannot distinguish "Error, never coercion" from "errored for
    some other reason".
- **Impact**: Two of AD-14's three pre-settled data cases ship with assertions that cannot discriminate
  the behaviour they name. These are the cases most likely to be silently changed by a future
  refactor of `bind.Resolve`'s caller, and the story explicitly claims them as verified.
- **Suggested Resolution**: For `explicit_null`, render through `buildPageModel` and assert the run
  count/`SourceText` set positively — the null cell contributes **no** run, while a sibling non-null
  column does (a differential, so "no runs at all" cannot pass). For `wrong_kind`, assert
  `errors.As(&RenderError)`, the diagnostic code, and `ElementID == "e2"`, matching the `absent`
  subtest's rigour.
- **Related AC**: AC4 (AD-14 triple)

### Finding 6: AC6's second mandated red-proof was not run, and red-proof #3 was run under a narrow `-run`

- **Severity**: Major
- **Category**: Tests / Convention
- **Location**: Story Delivery Log, "Red-proofs run" table, rows 3 and 8
- **Observation**: Two deviations from this story's own evidence rules.
  - **AC6(b) was never mutated.** AC6 states: *"Red-proofs — **both are required, and each must redden
    a DIFFERENT assertion**. … A test that only survives (a) has not established this AC."* The
    Delivery Log's row 8 covers (a) and then says of (b): *"is not independently exercised by a live
    mutation; it is guaranteed by construction."* The reasoning is verifiable and the reviewer confirms
    it by inspection — the header rect/run emission at `table_render.go:~505-513` is unconditional and
    precedes the `if len(items) > 0` gate — but D-000.9's extension is explicit that *"a measurement
    offered as evidence must name the command, name the mutation under which it reddens, and confirm
    the mutation was run."* An argument is not a measurement.
  - **Red-proof #3 contradicts itself.** Its Command cell reads *"full suite (`-run` narrowed to the
    target test to observe …)"*. Task 9 and every AC require red-proofs *"over the WHOLE suite, never
    behind a narrow `-run`"* — precisely because 4.1's Finding 1 was a mutation whose vacuity control
    reddened invisibly behind a `-run`. As recorded, #3 cannot say whether anything else reddened.
- **Impact**: The story's evidence record is the artifact the next story trusts. One mandated red-proof
  is absent and one is recorded under the exact condition the rule was written to forbid.
- **Suggested Resolution**: Run AC6(b) as a real mutation (e.g. `continue` the whole table when
  `len(items) == 0`) over the full suite and record which assertion of
  `TestEmptyCollectionRendersHeaderOnly` moved — it must be the header-present assertions, a different
  one from (a)'s counts. Re-run #3 without `-run` and record the full-suite figure.
- **Related AC**: AC6; task 9

### Finding 7: `TestNoRealToUnicodeSectionExceedsTheCapRedProof` does not red-proof the test it is named for

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/tounicode_corpus_test.go:92-113`
- **Observation**: The test is presented as *"the mandatory red-proof … for the corpus witness above"*,
  and the story's task 16 records it as such. It never invokes
  `TestNoRealToUnicodeSectionExceedsTheCap`, never places an over-cap section in the corpus that test
  walks, and never observes that test failing. It calls the shared helper `toUnicodeSectionSizes` on a
  hand-built string, asserts the helper returns 101, and **passes**. Its own doc comment softens the
  claim to *"must redden the SAME size-extraction logic that test uses"*, which is a materially weaker
  statement than its name and than the Delivery Log's.
- **Impact**: A green test named `…RedProof` reads to a future maintainer — and to Story 4.7's author,
  who inherits this guard at exactly the moment it becomes load-bearing — as proof that the corpus
  witness has teeth. It never demonstrated that. **The reviewer ran the real red-proof (M4): a
  temporary `fixtures/zz-reviewer-probe/expected.pdf` with a 101-entry section, and
  `TestNoRealToUnicodeSectionExceedsTheCap` did redden correctly**, naming the fixture and logging
  `observed maximum section size = 101 (cap 100)`. So the guard is sound; the *evidence artifact* is
  not what it claims.
- **Suggested Resolution**: Make it a genuine red-proof: have the corpus walk take its root directory
  as a parameter (or a package-level test seam), point the red-proof at a `t.TempDir()` containing one
  crafted over-cap `expected.pdf`, and assert the walk **reports a violation**. Alternatively rename it
  to `TestToUnicodeSectionSizeExtraction` and record M4 in the Delivery Log as the actual red-proof.
- **Related AC**: AC9 / D-4.2.4

### Finding 8: `buildToUnicodeCMap` lost its doc comment to the new constant

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/internal/pdf/textdoc.go:571-591`
- **Observation**: `bfcharSectionCap`'s comment block was inserted with **no blank line** after the
  pre-existing ~20-line comment describing `buildToUnicodeCMap` (Story 2.3's provenance, the caller-supplied
  entries, the empty-`Text` `<>` case, the BMP-only restriction). Godoc now attaches all of that prose
  to the constant, and `buildToUnicodeCMap` is left with **no doc comment at all**.
- **Impact**: The most valuable documentation on the emitter — the shaping/cluster reasoning a future
  reader needs before touching it, at exactly the file 4.7 depends on — is now filed under an integer.
- **Suggested Resolution**: Insert a blank line before `// bfcharSectionCap is DW-14's cap:` so the
  original block stays attached to the function.

### Finding 9: Two doc comments cite a test name that does not exist

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/internal/pdf/textdoc.go:584`; `folio-go/internal/pdf/tounicode_chunk_test.go:131`
- **Observation**: Both name `TestNoToUnicodeSectionAnywhereExceedsTheCap`. No such test exists;
  `grep -rn` finds only `TestNoRealToUnicodeSectionExceedsTheCap` (package `folio`).
- **Impact**: The constant's comment stakes its whole D-000.9 argument — *"the corpus-wide witness
  below … can each state their own expectation as an INDEPENDENT literal"* — on a citation a reader
  cannot follow.
- **Suggested Resolution**: Correct both to `TestNoRealToUnicodeSectionExceedsTheCap` and note it lives
  in package `folio`, not `internal/pdf`.

### Finding 10: The PDF-spec citation understates the cap — it is mandated, not recommended

- **Severity**: Minor
- **Category**: Correctness (documentation)
- **Location**: `folio-go/internal/pdf/textdoc.go:571-580`
- **Observation**: The comment states the spec *"permits an implementation-defined section size"* and
  frames 100 as long-standing Adobe **guidance**. ISO 32000-1's ToUnicode CMap clause (§9.10.3) and
  the Adobe CMap specification it derives from **require** that a `beginbfchar` section contain no more
  than 100 entries. The cited "Table 115" is the general CMap operator table, not the clause carrying
  the limit.
- **Impact**: The constant's value is correct and the code is right, but the justification is weaker
  than the truth. A future reader — plausibly at 4.7, under re-record pressure — could read this as
  discretionary and raise the cap, producing non-conforming PDFs that this repo's own tests (which pin
  their own literal 100) would then contradict.
- **Suggested Resolution**: Restate the citation as a spec **requirement** (ISO 32000-1 §9.10.3 / Adobe
  CMap spec: at most 100 entries per `beginbfchar` section), keeping the reader-compatibility note as
  supporting colour.

### Finding 11: `TestToUnicodeSectionCapBoundaryValues` derives its expectation from the constant under test

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/pdf/tounicode_chunk_test.go:104-135`
- **Observation**: Its assertion is `if c > bfcharSectionCap` — the emitter chunks at
  `bfcharSectionCap` and the test compares against the same symbol, so for *any* value of the constant
  the check holds. It can catch chunking being **removed** (sections of 101 and 250 would then appear),
  but it cannot catch the cap being set to a wrong value — the property D-000.68's independent-literal
  rule exists to protect, and which its sibling `TestToUnicodeSectionsAreChunkedAtTheSpecifiedCap`
  correctly does protect with `wantDeclared := []int{100, 100, 50}`. Separately, the doc comment claims
  the corpus spans *"at, just under, and over the cap"*; the faces are 45 / 100 / 101 / 250 — there is
  nothing just under 100.
- **Impact**: Low on its own (the primary chunking test carries the literal), but the comment overstates
  what this test establishes, and 45 is described as "just under" a cap of 100.
- **Suggested Resolution**: Use a test-owned `const capForTest = 100` here as
  `tounicode_corpus_test.go` already does with `bfcharCapForTest`, and either add a 99-entry face or
  drop "just under" from the comment.

### Finding 12: Three tests re-pointed to `row.` binds still render an empty collection

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_render_test.go` — `TestColumnGeometryNeverNegotiatesAgainstLabelContent`,
  `TestColumnAlignWinsOverStyleAlign`, `TestColumnAlignWinsOverHeaderStyleAlign`
- **Observation**: Divergence D4's critique — *a table test that renders `"items": []` while claiming a
  property about a table with rows* — was applied to `TestTableStyleFieldsAreNotDataDriven` and its
  control (AC10, correctly), but not to these three, which had their column binds rewritten from
  `item.` to `row.` and still render `{"items": []}`. Their new `row.` binds are therefore never
  evaluated. `TestColumnGeometryNeverNegotiatesAgainstLabelContent` is specifically 4.1's AC2 anchor
  pattern and remains a header-only observation.
- **Impact**: Not a regression — these tests are no weaker than at `b5eb557` — but the same vacuity the
  story diagnosed and fixed in two places persists in three others that this story touched, including
  the one it names as its own model (AC2's *"modelled on
  `TestColumnGeometryNeverNegotiatesAgainstLabelContent`"*).
- **Suggested Resolution**: Give each a non-empty, identical collection so the alignment and geometry
  properties are asserted over the header **and** the body. Low cost — the fixtures already carry the
  binds.

### Finding 13: `TestRenderScopeFenceIgnoresTableBind`'s doc comment now contradicts itself

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/render_bind_test.go`
- **Observation**: The comment block retains *"Story 3.1 does not evaluate column binds either — that
  is Story 3.2"* and *"…that stays Story 3.2's"*, while the newly appended 4.2 note in the **same
  block** states the fixture corrections were *"surfaced by this story because column binds are
  evaluated for the first time"*. Both cannot be true.
- **Impact**: The next reader of the scope fence cannot tell whether column binds are evaluated. The
  test's discriminating power is also now diluted: it asserts only that `Render` succeeds with non-empty
  bytes, and success now depends on the bind evaluating cleanly as well as on the fence holding, so a
  failure no longer localises. (The dilution is inherent to the story; the stale comment is not.)
- **Suggested Resolution**: Strike the two obsolete sentences and state plainly that as of 4.2 column
  binds **are** evaluated, so this fence now tests that *text* binding does not scan them.

### Finding 14: `fontChain` is called before any cell is known to carry text

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/table_render.go:527-530`; `fontChain` at `folio-go/render.go:916-919`
- **Observation**: Inside `if len(items) > 0`, `fontChain(doc, el)` is called unconditionally, before a
  single cell binding is resolved. `fontChain` errors whenever `style.fontFamily` is absent, with the
  message *"has text but no style.fontFamily to resolve a font from"*. So a table with a non-empty
  collection and no `style.fontFamily` hard-fails the whole render even when **every** cell resolves to
  an explicit JSON `null` or the empty string — i.e. when the body carries no text at all, and the error
  message is factually wrong. The header path handles the equivalent case correctly, skipping the font
  requirement per column via `if col.Label == "" { continue }` (`table_render.go:~421`).
- **Impact**: A narrow but real tension with AD-14 (*explicit `null` = empty and **not** an error*) and
  with D-4.1.1's wording (*required on any element **carrying text***). An author whose data legitimately
  yields all-empty rows gets a hard render error naming text that does not exist. Asymmetric with the
  header's own treatment of the same condition.
- **Suggested Resolution**: Resolve the cells first and call `fontChain` lazily on the first cell whose
  bound text is non-empty, mirroring the header's per-label guard; or, if eager failure is intended,
  say so in a comment and make the message accurate. Worth surfacing to the lead rather than deciding
  here, since it touches D-4.1.1's boundary.
- **Related AC**: AC4 (AD-14), AC5

### Finding 15: The clip warning is a hand-copied near-duplicate, and calls a column an "element"

- **Severity**: Minor
- **Category**: Maintainability / Correctness (message text)
- **Location**: `folio-go/table_render.go:640-647`; existing site `folio-go/render.go:769-775`
- **Observation**: The detection is correctly shared (`detectWidthOverflow`, `millipoints`) and no new
  diagnostic code was minted — D-000.65 honoured. But the **message** is copy-pasted, differing in
  exactly two noun phrases (`the element's declared width of %s` → `the column's content width of %s`;
  `clipped at the box's left/right edges` → `clipped at the column's left/right edges`). Nothing keeps
  the two in sync — a future edit to the shared remainder (e.g. the `FR44` reference) must be made
  twice. Separately, the new message opens `"element %s: "` while the `%s` it interpolates is
  `overflow.elementID` carrying a **column** id (`string(col.ID)`), so a table clip reads
  `element e3: …` where `e3` is a column.
- **Impact**: Low functionally; the divergence in the width phrase is substantively correct. But the
  literal word "element" is now inaccurate at this site, and AC4's own grounds are that
  `columns[].id` exists *"so a diagnostic can name a column"* — the message should say so.
- **Suggested Resolution**: Extract a shared message builder taking the noun ("element"/"column") and
  the width label, so both sites share one sentence; and have the table site say `column %s:`.

### Finding 16: `TestCellBindingsResolveInRowScope` cannot detect a column swap

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_render_row_test.go:248-292`
- **Observation**: The test scans all runs for the three sentinel strings and records only *that each
  appeared somewhere*. It never checks which **column** each landed in. If the row-root and
  document-root columns' resolutions were transposed, all three sentinels would still be present and the
  test would pass. The anchor design (deliberately different root and row values for `a`) is right and
  does make AC4's own red-proofs (a) and (b) redden; the gap is only that column attribution is
  unasserted.
- **Impact**: A resolution that returns the correct set of values against the wrong columns is invisible.
- **Suggested Resolution**: Assert each sentinel's run X falls within its own column's declared span
  (the columns are 80pt apart and the test already owns those literals), or assert the ordered
  `SourceText` sequence as `TestTableRendersOneRowPerCollectionElementInDataOrder` does.

### Finding 17: Delivery Log's `-count=2` parenthetical is wrong about the skip

- **Severity**: Minor
- **Category**: Correctness (record)
- **Location**: Story Delivery Log, Gate results table, `folio-go` (determinism, AC11) row
- **Observation**: Records **1967 pass / 0 fail**, glossed as *"doubled minus the permanently-skipped
  test, itself never doubled"*. Measured independently: **1967 pass / 0 fail / 2 skip** — the skip **is**
  doubled. The stated arithmetic (`2 × 984 − 1 = 1967`) therefore holds for a different reason than the
  gloss gives, and one test that passes under `-count=1` does not contribute a second pass under
  `-count=2`.
- **Impact**: The headline figure is right and AC11 is satisfied; the explanation is not, and it papers
  over a small unexplained delta that a future reader may re-derive and be confused by.
- **Suggested Resolution**: Correct the parenthetical to `1967 pass / 0 fail / 2 skip` and either
  identify the single non-doubled pass or drop the derivation rather than assert a wrong one.

### Finding 18: Collection-path parsing is duplicated from `checkTableBindings`

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/table_render.go:517-522`; `folio-go/render.go:305-309`
- **Observation**: `strings.TrimSuffix(tbl.Bind, "[]")` + conditional `strings.Split(…, ".")` is
  reproduced verbatim. The two must agree or the render resolves a different collection from the one
  validated. They agree today (verified), and `checkTableBindings` is confirmed to run first
  (`render.go:1364`, inside `predictDocument`, before `collectBandTableRuns`) — so the discarded error
  from `data.Lookup` is safe as the comment claims.
- **Suggested Resolution**: Extract a `tableCollectionSegments(bind string) []string` helper used by both.

### Finding 19: Dead duplicate run-count check in the AC1 rewrite

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/table_render_test.go`, `TestTableRendersOneRowPerCollectionElementInDataOrder`
- **Observation**: After `len(pages[0].Runs) != columns*(1+rows)` has already `Fatalf`'d, the later
  `if len(pages[0].Runs) != len(want)` can never fire — `len(want)` is `columns*(1+rows)` by
  construction. Also the rect-count check uses `t.Errorf` while the run-count check uses `t.Fatalf`,
  for no stated reason.
- **Suggested Resolution**: Drop the second check, or keep only it.

### Finding 20: `TestTableStyleFieldsAreNotDataDriven`'s failure message cites the wrong AC

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/table_render_test.go`, `TestTableStyleFieldsAreNotDataDriven`
- **Observation**: The doc comment was updated to AC10 but the `t.Fatal` text still ends *"style must
  never be data-driven (AC7)"* — AC7 in this story is the pagination-shape fence.
- **Suggested Resolution**: Change the message to AC10.

### Finding 21: The identity test calls `collectBandTableRuns` twice with identical arguments

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/table_render_row_test.go:588` and `:617`
- **Observation**: The same call is made twice, discarding the runs the first time and the rects the
  second, with a comment explaining the re-collection. One call binding both results would do.
- **Suggested Resolution**: Collect once into two variables.

---

### ACs explicitly considered

- **AC1** — satisfied. Rewrite is an **ordered** `SourceText` sequence comparison over five distinct
  values with rect and run counts; the subagent-verified diff shows the old set-membership check
  replaced by an ordered one. Nit 19 only.
- **AC2** — satisfied, and strong. Test-owned `{X,W}` literals identical across the narrow and wide
  renders, an exact `(n−1)·vm.Advance` delta (not `>`), a glyph-count control in the same test, all three
  shaping paths, page count asserted as 1. **Nothing widens**: `layout.ColumnWidths` receives only
  declared `c.Width` values and `contentW = cg.Width − padLeft − padRight`; the co-primary invariant
  (AD-13 / 4.1 AC2) holds. Weakened only indirectly by Finding 2 (padding inert).
- **AC3** — satisfied. Existing `DiagCodeTextClippedWidth` and `detectWidthOverflow` reused, no code
  minted; `ClipWidth` asserted against a test-owned literal; column geometry asserted unchanged. See
  Finding 15 on the message.
- **AC4** — satisfied for row/root/params resolution and the declared alias; **partially** for the AD-14
  triple (Finding 5) and unable to detect column attribution (Finding 16).
- **AC5** — satisfied for `fontFamily` (error routed through `fontChain` verbatim, substring-asserted —
  no third spelling) and for `background`/`align` (M3b confirms teeth). **Not** satisfied for `valign`
  (Finding 4), `border` (Finding 1) or `padding` (Finding 2).
- **AC6** — satisfied in code; the header-present, run-label and zero-diagnostic assertions are all
  positive, plus a differential against a one-item collection. Red-proof (b) not run (Finding 6).
- **AC7** — satisfied. `layout.Paginate`, `MixedItemError` and `internal/layout` are **untouched**
  (confirmed: no changes under `internal/layout/` in the working tree). Every new test asserts
  `len(pages) == 1` and logs it. No pagination property is asserted anywhere. The row-identity carrier
  lives entirely in package `folio`'s own structs, pre-empting nothing in 4.3.
- **AC8** — satisfied. The pin is correct and its doc comment states, in its own words, *"THIS TEST
  RECORDS CURRENT BEHAVIOUR. IT DOES NOT ENDORSE IT"*, naming the open interleave-vs-relax question for
  the lead (D-4.2.3). Requirement met exactly.
- **AC9** — satisfied on substance; all three mandated checks pass (emitted-bytes assertion, named
  constant with citation, independent literal split), and the corpus property reddens under the
  reviewer's own forced violation (M4). Findings 7–11 are documentary and evidentiary, not behavioural.
  Byte-neutrality is real: cap 2 → 972/12 exactly as recorded.
- **AC10** — satisfied. Both datasets non-empty and identical, `overdue` still unbound, control still
  demonstrates a bound difference moves the bytes. Nit 20 only.
- **AC11** — satisfied. `-count=2` measured 1967/0 independently; `lint`'s `TestFloatTypedProductionScan`
  is green at 115/0 and is the float instrument (not `go build`); row iteration is
  `for rowIdx, rowVal := range items` over a **slice**; no map reaches output ordering in the new
  production path. AD-23/AD-1 hold: the only `internal/` change (`internal/pdf/textdoc.go`) is pure
  integer arithmetic with no `time`/`os`/`math/rand`/`net`/transcendentals/package-level mutable state.
  Record nit at Finding 17.
- **Scope fence honoured.** No pagination or row-splitting behaviour (4.3), no repeated header (4.4),
  no footer aggregates (4.5) — `columns[].footer`/`footerOf`/`footerFormat` confirmed unread at render
  — no over-tall-row handling (4.6), and `altRowBackground` remains inert (4.8), with only its
  precedence documented. `TestFolioMethodNamesAreInjective` green (R6); no new `internal/` package (R5);
  no schema field invented (R7).
- **`folio-format.md`** carries both required statements, unambiguously: data cells cascade from `style`
  **only** with no `headerStyle` arm, and `altRowBackground` **wins over** `style.background` on the
  rows it applies to, explicitly labelled as ruled-but-not-yet-implemented (4.8's). The only defect is
  that the same sentence also claims `valign` cascades — see Finding 4.

### Verdict

The substance of this story is largely well built: row generation, wrapping-inside-the-column,
clipping, row-scope binding, the header/body style split and the DW-14 chunking are all correct, and
several of the instruments are genuinely strong (AC2's exact-delta geometry test and AC1's ordered
sequence in particular). DW-14 — the highest-stakes item, and the one 4.7 will freeze — is landed
correctly and its corpus guard has real teeth under an independent forced violation.

What blocks is the recurring programme defect, present three times: **the mid-flight-ruled data-cell
chrome ships with two of its three fields (border, padding) inert to the entire three-module gate, and
the owner-ruled row identity ships with its defining property — a wrapped row's items all agreeing —
unasserted.** Each was confirmed by a mutation that left the suite at exactly 984 / 0 / 1. All three
are cheap to close and none requires re-opening a ruling.

**Changes Requested.**

## Finding Resolutions (bmad-story-finisher)

Triaged by the finisher after the review above. **20 of 21 findings (plus all 3 Blockers) FIX; one
Minor (Finding 14) DEFERRED** to the existing owner of font-resolution error handling — a
pre-existing boundary question this story did not introduce and should not decide unilaterally.
Every Blocker is fixed with a genuine, mutation-confirmed instrument; none was argued away and none
required re-opening D-4.2.1/D-4.2.2's rulings. Full mutation transcripts are in "Finisher fixes and
re-measurement" above, in the Delivery Log.

| # | Finding | Severity | Decision | Rationale | Files changed |
|---|---|---|---|---|---|
| B1 | Data-row cell border inert to the entire gate | **Blocker** | **FIX** | The mechanism (`buildCellRect`, shared with the header) was already correct — this was a pure test gap. Added `TestDataRowBorderIsDrawn`, asserting `HasStroke`/`Stroke`/`StrokeWidth`/`Edges` on data-row rects over a non-empty collection. Reproduced the reviewer's M1 (force border args to `false, template.Border{}`) both before (984/0/1 unchanged) and after (reddens exactly this test) the fix. | `folio-go/table_render_row_test.go` |
| B2 | Data-cell padding inert to the entire gate | **Blocker** | **FIX** | Same class as B1: padding was correctly wired into row height and content X/W, but nothing asserted it. Added a differential test asserting an EXACT delta (not a comparison of two equally-padded renders, which is what `TestTableStyleFieldsAreNotDataDriven` does and why it couldn't see this). Reproduced the reviewer's M3 (zero all four padding edges) both before and after. | `folio-go/table_render_row_test.go` |
| B3 | Row identity's defining property (a wrapped row's lines all agreeing with the row whose DATA they render) unasserted | **Blocker** | **FIX** | The count-of-distinct-indices shape could not see M2b (continuation lines claiming the NEXT EXISTING row). Rewrote the test to embed a per-row marker in every row's cell content and assert each table-row-line run's OWN marker matches its `rowIndex`, cross-checked against the `tableRectSource` group whose extent contains it. Reproduced M2b both before (984/0/1 unchanged, confirmed independently) and after (reddens exactly this test) the fix. | `folio-go/table_render_row_test.go` |
| 1 | `resolvedBodyStyle.valign` resolved and never read | Major | **FIX (wired)** | Implementing per-cell vertical-slack distribution does not conflict with AC7/D2's shared-line-extent invariant (it only changes which physical line SLOT a shorter cell's own content occupies, never the row's own height or any column's geometry) — a clean, in-scope fix rather than a documentation correction. AC5's mechanism text, the code comment, and `folio-format.md` were all already correct; only the code needed to catch up. New test `TestDataCellValignDistributesRowSlack`, anchored to the wrapping column's own observed (mutation-immune) per-line Y positions, per D-000.68. Red-proof (disable the per-cell offset) confirmed to redden exactly this test. | `folio-go/table_render.go`, `folio-go/table_render_row_test.go` |
| 2 | AD-14's `null`/wrong-kind subtests under-asserted | Major | **FIX** | `explicit_null` now asserts a positive, differential run count against a second always-bound sibling column. `wrong_kind` now asserts the same rigour as `absent` (a `*RenderError`, its code, the column id) — confirmed the production code already reuses `DiagCodeBindingPathAbsent` (D-000.65), so no code change was needed, only the assertion. | `folio-go/table_render_row_test.go` |
| 3 | AC6(b)'s red-proof never run; red-proof #3 recorded under a narrow `-run` | Major | **FIX** | Both re-run for real, over the whole suite, no `-run`, recorded in the Delivery Log's "Re-measured red-proofs" table rather than only argued. AC1(a)'s mutation now shows 16 top-level tests reddening (matching the story's own record, this time genuinely observed rather than narrowed); AC6(b)'s mutation shows the header-present assertions moving, a DIFFERENT assertion from (a)'s. | story file (Delivery Log) |
| 4 | `TestNoRealToUnicodeSectionExceedsTheCapRedProof` never exercised the test it was named for | Major | **FIX** | Extracted `walkToUnicodeCorpus(fixturesDir)` — the ONE walk both the real corpus test and its red-proof now call. The red-proof builds a `t.TempDir()` fixture with a hand-built 101-entry section and asserts the walk itself reports a violation — reproducing the reviewer's manual M4 probe as a standing test rather than a one-off. | `folio-go/tounicode_corpus_test.go` |
| 5 | `buildToUnicodeCMap` lost its doc comment to the new constant | Minor | **FIX** | Confirmed by reading: the comment block was one contiguous run immediately preceding `const bfcharSectionCap`, so godoc attached ALL of it to the constant and left the function undocumented. Reordered: the constant's own comment now precedes the constant, and the function's original doc comment (plus a one-line pointer to DW-14's chunking) now precedes the function. | `folio-go/internal/pdf/textdoc.go` |
| 6 | Two doc comments cite a test name that does not exist | Minor | **FIX** | Confirmed by grep: `TestNoToUnicodeSectionAnywhereExceedsTheCap` does not exist. Corrected both citations to `TestNoRealToUnicodeSectionExceedsTheCap` (package `folio`). | `folio-go/internal/pdf/textdoc.go`, `folio-go/internal/pdf/tounicode_chunk_test.go` |
| 7 | PDF-spec citation understates the cap as guidance, not a requirement | Minor | **FIX** | Restated as ISO 32000-1 §9.10.3 / the Adobe CMap spec MANDATING at most 100 entries per section, keeping the reader-compatibility note as supporting colour rather than the primary justification. | `folio-go/internal/pdf/textdoc.go` |
| 8 | `TestToUnicodeSectionCapBoundaryValues` derives its expectation from the constant under test | Minor | **FIX** | Added a test-owned `capForTest = 100` (same discipline as `tounicode_corpus_test.go`'s `bfcharCapForTest`), independent of `bfcharSectionCap`. Also added a 99-entry (just-under-cap) face, since the doc comment claimed the corpus spans "just under" the cap but no face did. | `folio-go/internal/pdf/tounicode_chunk_test.go` |
| 9 | Three tests re-pointed to `row.` binds still render an empty collection | Minor | **FIX** | Same vacuity class D4 already fixed in two places (AC10). Gave `TestColumnGeometryNeverNegotiatesAgainstLabelContent`, `TestColumnAlignWinsOverStyleAlign` and `TestColumnAlignWinsOverHeaderStyleAlign` a non-empty collection identical across their compared renders. `TestColumnAlignWinsOverHeaderStyleAlign` now additionally asserts the AC5/D-000.76 boundary directly: a data cell must NOT inherit `headerStyle.align`. | `folio-go/table_render_test.go` |
| 10 | `TestRenderScopeFenceIgnoresTableBind`'s doc comment contradicts itself | Minor | **FIX** | One paragraph said column binds "are not yet evaluated... that stays Story 3.2's"; a second, newer paragraph said they are "evaluated for the first time" by this story. Rewrote the first paragraph to state plainly that as of 4.2 column binds ARE evaluated in row scope, and that this test now proves TEXT binding specifically does not scan them. | `folio-go/render_bind_test.go` |
| 11 | `fontChain` called before any cell is known to carry text | Minor | **DEFER** | The reviewer's own recommendation: "worth surfacing to the lead rather than deciding here, since it touches D-4.1.1's boundary." This is a pre-existing failure-mode asymmetry (the header already has an equivalent, unexamined gap via its own per-column label guard), not a defect 4.2 introduced. Deciding eager-vs-lazy font resolution is a font-resolution-error-handling policy question shared with `fontChain`'s existing text-element caller — belongs with Obligation 3's named owner ("whoever next touches font-resolution error handling"), not a table-row story. | none (deferred) |
| 12 | Clip warning is a hand-copied near-duplicate and calls a column an "element" | Minor | **FIX** | Extracted `widthClipMessage(noun, widthLabel, overflow)`, shared by both the text-element and table-cell clip sites. The table site now correctly says "column %s:" (matching AC4's own grounds that `columns[].id` exists so a diagnostic can name a column), while preserving the substantively-different "declared" (an element's own box) vs "content" (a column's width minus padding) phrasing — collapsing that distinction would have made the table's statement less accurate, not more consistent. | `folio-go/render.go`, `folio-go/table_render.go` |
| 13 | `TestCellBindingsResolveInRowScope` cannot detect a column swap | Minor | **FIX** | Added an assertion that each sentinel's run X falls at its own column's exact origin (test-owned literals: 0/80000/160000), so a resolution that returns the correct VALUES against the WRONG columns is now caught. | `folio-go/table_render_row_test.go` |
| 14 | Delivery Log's `-count=2` parenthetical wrong about the skip | Minor | **FIX** | Confirmed independently: the skip DOES double under `-count=2`. Re-measured the finished tree (`985/0/1` at `-count=1`, `1969/0/2` at `-count=2`) and recorded both totals plainly, noting that a naive `2n-1` cross-check is unreliable in the presence of shared subtest names across runs — rather than asserting a fixed but still-approximate derivation. | story file (Delivery Log) |
| 15 | Collection-path parsing duplicated from `checkTableBindings` | Nit | **FIX** | Extracted `tableCollectionSegments(bind string) []string`, used by both `checkTableBindings` (render.go) and the row-generation loop (table_render.go). | `folio-go/render.go`, `folio-go/table_render.go` |
| 16 | Dead duplicate run-count check in the AC1 rewrite | Nit | **FIX** | Confirmed the second check could never fire (guaranteed by the first `Fatalf`, which already checks the same equality). Removed it, keeping the first. | `folio-go/table_render_test.go` |
| 17 | `TestTableStyleFieldsAreNotDataDriven`'s failure message cites the wrong AC | Nit | **FIX** | Changed "(AC7)" to "(AC10)" in the failure message, matching the test's own (already-correct) doc comment. | `folio-go/table_render_test.go` |
| 18 | Identity test calls `collectBandTableRuns` twice with identical arguments | Nit | **FIX** | Combined into one call, binding both `runs` and `rects` — done in the same edit as Blocker 3's rewrite. | `folio-go/table_render_row_test.go` |

**Follow-ups for a new ticket / the existing owner (none of them block this story):**
- Finding 11/14 (deferred): whether `fontChain` should be called lazily, only once a cell's own
  bound text is known to be non-empty, mirroring the header's per-column label guard — or whether
  eager resolution is intentional and its error message should simply stop claiming the element
  "has text" when it may not. Owner: whoever next touches font-resolution error handling
  (Obligation 3, already named in this story's own text).
- DECISION-2 (row atomicity) and DECISION-3 (within-band diagnostic interleaving) remain open for
  Story 4.3 and the engineering lead respectively, exactly as this story's own creation record
  states — this finishing pass did not touch either.
- D-000.79 (recorded independently in `folio-mvp-decision-log.md` during review) proposes a
  standing process change — the STORY CREATOR names the deletion-mutation for every output-producing
  AC, so a Class-A "no witness at all" defect is screened before the developer ever writes the
  implementation. That is a Story 4.3+ process question, not something this finishing pass can
  retroactively apply to 4.2's own creation record.

