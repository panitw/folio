# Story 4.5: Render footer aggregates without orphaning them

Status: done
Epic: 4 — A Go developer can render the golden report (**the C4 gate**)
Covers: **FR25, FR27**
Created: 2026-08-27, at HEAD `530a47b` (Story 4.4, ledger commit), tree clean

---

## Baseline, measured in this run at creation (HEAD `530a47b`)

Every figure below was taken in this run, not carried from the brief. Counting convention matches
Stories 4.1–4.4: `--- PASS` / `--- FAIL` / `--- SKIP` lines **anywhere** in `-v` output, so subtests
are counted. Report the convention with the number.

| Gate | Command (run from that module's directory) | Measured |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **1022 pass · 0 fail · 3 skip** |
| `lint` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **115 pass · 0 fail** |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **3 pass · 0 fail** |

The three skips, named rather than counted:

```
--- SKIP: TestXrefEntriesRejectsMalformedSubprocess
--- SKIP: TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable      (DW-21, FOLIO_HEAVY=1)
--- SKIP: TestTwoTablesWithPageCountFooterRenderConsistently          (DW-21, FOLIO_HEAVY=1)
```

`TestCorpusMeetsP6ExerciseFloors` is excluded by `-skip` and therefore prints no line at all. It is
**red by design** (D-000.17 / D-2.1.14 / D-000.57 / D-000.74) and is never to be "fixed".

---

## Creation probe — what a footer-configured table actually does at HEAD

A probe test was written into `folio-go`, run, and **deleted**; the tree was re-verified clean at
`530a47b` afterwards. It drove the real `ParseTemplate` → `collectBandTableRuns` → `layout.Paginate`
path with a two-column table whose second column declares `"footer": "sum"` and whose first declares
`"footer": "count"`, over a 200×150pt page with 10pt margins and 10pt header/footer bands.

**Finding 0 — a numeric column bind must be wrapped.** The first probe used `"bind": "{{row.b}}"`
over numeric data and failed at render with *"resolved to a number, not a string — text bindings are
never coerced"* (AD-14, and `bind`'s own rule). **Every realistic footer column is therefore
derivation shape 2** — `{{formatNumber(row.b, "#,##0.00")}}` — which also derives `footerFormat`. The
fixture for this story must use shape 2, or it does not resemble any document a user can write.

**Finding 1 — the derivation is live and its result is consumed by nothing.**

```
derivedFooters = map[e3:{FooterOf:items.b FooterFormat:#,##0.00 HasFooterFormat:true}]
```

`count` derives nothing (correct — D-1.4.1: `count` never names a per-column numeric source). At HEAD
`Template.derivedFooters` is written by `ParseTemplate` and **read by nothing except three tests**
(`folio_expr_validate_test.go`). **This story is its first consumer.**

**Finding 2 — no footer is rendered, and there is no look-alike for one.** For a 20-row table the
probe collected **21 table rect sources** — one header, twenty data rows, nothing after row 19 — and
**42 runs, of which zero were neither a header label nor a data-row line.** FR27 is genuinely unmet;
AC1 is not accidentally true; and unlike Story 4.4 there is **no geometric twin already sitting where
the new thing goes**. The trap in this story is elsewhere (see Finding 4).

**Finding 3 — measured geometry, for the arithmetic every AC below is anchored on.**

```
contentTop = 10,000mp    contentHeight = 110,000mp    ceiling = 120,000mp
header row: top 10,000  bottom 20,000    height 10,000mp  (declared headerHeight)
data row i: top 20,000 + i*10,896        height 10,896mp  (one line, default padding)
```

Capacity, hand-derived and then confirmed by running the probe at three row counts:

| rows | pages | partition | why it matters here |
|---|---|---|---|
| 20 | 3 | 9 / 9 / 2 | matches Story 4.4's post-fix prediction exactly; page 1 and 2 carry a repeated header and `RowDisplacement {e1 10,000}` |
| 18 | 2 | 9 / 9 | page 1's rows end at 216,128 against a reserved ceiling of 218,064 — **1,936mp left** |
| **9** | **1** | **9** | last row ends at **118,064** against a ceiling of **120,000** — **1,936mp left** |

**The 9-row case is this story's orphan fixture, and it is the robust one.** It needs no continuation
page and no reservation to create the squeeze: a footer row roughly a data row tall (≈10,896mp)
cannot fit in 1,936mp. The 18-row case produces the same squeeze *only because Story 4.4 reserves
10,000mp on the continuation page* (the raw slack there is 11,936mp, which a footer **would** fit) —
a fixture whose orphan-ness is a side effect of another story's mechanism is the wrong fixture to
pin this rule on. **Re-derive this arithmetic during development rather than copying it.**

### Finding 4 — the two-part trap in this story, and where it is NOT

Probed rather than reasoned about:

- **The orphan property does NOT already hold.** With a naive footer (its own group, or none), the
  9-row fixture puts rows 0..8 on page 0 and the footer **alone** on page 1. Correct behaviour puts
  rows 0..7 on page 0 and **row 8 plus the footer** on page 1.
- **The page COUNT does not move: 2 before, 2 after.** Only the partition moves, 9/0 → 8/1.
  D-000.68's *"a count is a lossy set"* has a concrete instance here. **A page-count assertion
  cannot carry AC5.**
- **The accident that IS live is Story 4.3's, and it applies to the footer's own chrome.** D-000.80's
  founding case: a row's chrome rect is one `ColumnItem` spanning the whole row, it fails the fit test
  first, and it drags the row's lines along as an *atomic proxy*. A footer row built the same way
  inherits that accident — which is why **AC5 carries a part (a) that deletes the footer's chrome and
  requires the test to still pass.**

---

## In plain terms (read this first if you just want the gist)

A statement's transaction table usually ends with a totals line. Until this story the template format
could declare one and the loader would check it was declared correctly, but nothing ever drew it. Now
it is drawn: a sum, a count or an average, in the column that asks for it, aligned the way that column
is aligned, computed by the same arithmetic the rest of the product already uses, and covering every
transaction in the data rather than only the ones printed on the last sheet. It never lands alone
either — if the totals line will not fit at the foot of a page, the last transaction line comes across
with it, so a page never opens with a stranded "Total" and nothing above it.

Two things about what shipped are worth knowing, because neither was right first time.

The first was a rounding bug on the money path. A totals line whose column did not say how to format
itself was being printed with no decimal places at all, so a true total of 30.85 appeared as 31. It
now prints at its own precision. Why nothing noticed is the part worth keeping: the only column in the
test data that took that path was counting things, and a count has no decimal places, so the wrong
answer and the right one looked identical. The test data now includes a column that would look
different — that, rather than the code fix, is what stops the same mistake returning.

The second was that several claims this story made about itself had no test behind them: that the
totals reach the shared arithmetic rather than a private copy of it, that a warning is raised when the
totals line genuinely cannot be kept with its data, and that the zeros an empty table shows were
computed rather than assumed. Each now has its own test, and each test was proved by deliberately
breaking the thing it watches and confirming it fails.

One limitation is deliberate and named rather than hidden: an average carries more decimal places than
the formatting language can express, so it prints at the widest precision that language allows. That
ceiling belongs to the format, not to this story. Alternating row shading and a row too tall for a
page remain later stories' work, as planned.

---

## Story

**As a template author,**
**I want the sum row to stay with its data,**
**So that a total never appears alone at the top of a page with nothing above it.**

---

## The pre-ruled obligation that is due here — established, and it is NOT what the brief said

The brief handed this story an obligation to mint `TABLE_FOOTER_SOURCE_UNRESOLVED` /
`TABLE_FOOTER_SOURCE_FORBIDDEN`. **Measured against the tree and the record, that obligation does not
exist. Both codes shipped at Story 3.6 and DW-6 is retired.** Evidence, all at `530a47b`:

- `folio-go/internal/diag/diag.go` declares `CodeTableFooterSourceUnresolved` and
  `CodeTableFooterSourceForbidden`; `diag_test.go`'s registry test pins both strings; `folio-go`'s
  public `DiagCodeTableFooterSourceUnresolved` / `DiagCodeTableFooterSourceForbidden` bridge them.
- `UNRESOLVED` is attached and reachable: `folio_expr_validate.go`'s `validateTableColumns`,
  `!derivable` branch. `FORBIDDEN` is attached and reachable at two sites in
  `internal/template/parse_bands.go` (`footerOf` alongside `footer: "count"`; `footerOf`/
  `footerFormat` with no `footer`).
- `deferred-work.md`'s DW-6 header reads **"RETIRED by Story 3.6 (R6, R8, AC2)"**, by replacement,
  in the same commit.

**What D-1.4.2's table actually routes to Story 4.5 is the other row of the same table:** *"Footer
uses the same aggregate evaluation as `{{sum(...)}}` — No — nothing renders a table until 4.5."* That
is **DW-7**, whose owner is *"**Story 4.5**, by name"* and whose anti-rot mechanism is, in its own
words, **"none possible for the footer half"**. **That is this story's inherited debt, and it is
AC4.** D-000.65 is not in tension with any of this: it says a code is minted when its condition first
ships, and both these conditions shipped at Story 3.2 and were coded at 3.6.

**So: mint no diagnostic code in this story, and stretch none.** D-1.4.1 already ruled the residuals
by name — *"**Mint no others:** a non-numeric value at `footerOf` is AD-14's existing wrong-kind
Error, and `avg` over an empty collection is Story 3.3's existing diagnostic. Reuse both."* Story
4.2's own footer-adjacent precedent is in the tree with the ruling quoted at the call site:
`table_render.go` reuses `DiagCodeBindingPathAbsent` for a column's unresolvable binding with the
comment *"(D-000.65: reuse, mint nothing)"*. Follow it. The **one** place where minting could become
legitimate here is DECISION-2 below, and only if the lead rules that arm into existence.

---

## What is actually in the tree, and what 4.5 must move

`collectBandTableRuns` (`folio-go/table_render.go`) emits, per content-band table: one header
`tableRectSource` (`isHeaderRow`) plus its column-label runs, then **one `tableRectSource` per data
row** (`isDataRow`, `rowIndex`) and one run group per physical line (`isTableRowLine`, `rowIndex`).
Row identity is carried, never reconstructed (D-4.2.2), and turned into `layout.ItemGroup` by
`tableRectSource.chromeRowGroup` / `textRunSource.lineRowGroup`. Nothing is emitted after the last
data row (probe Finding 2).

Both pagination passes consume those same producers: `contentColumnItems` (`page_number.go`, PHASE A,
page count only) and `paginateDocument` (`render.go`, PHASE B, the real pages). **They append in
different orders and rely on the carried group identity to agree** (D-4.3.2, and that function's own
"CORRECTED CLAIM" comment). A footer emitted by the one producer therefore reaches both passes for
free — *provided* it carries a group identity in both, which is AC7's subject.

`layout.Paginate` (`internal/layout/paginate.go`) slides one window down an immutable column. Story
4.3's `ItemGroup`/`ItemGroupKey{ElementID, IsHeader, Index}` keeps a group's members on one page by
testing the group's **union extent**; Story 4.4 layered a per-table displacement channel on top —
`PageAssignment.HeaderRepeats` / `RowDisplacement`, `TableHeaderRepeat`, `TableRowDisplacement`,
`TableHeaderSuppressed` — deliberately **not** folded into the page-wide `Shift`, precisely so that
4.5 and 4.6 would not have to tear out a cheaper shape mid-gating-epic. **The channel is there for
you. Use it; do not rebuild it and do not contort it.**

Note what `paginate.go`'s own grouping comment already anticipated at 4.3:

> *"the union rather than any one member's own chrome rect … Today the union and a data row's own
> chrome rect ARE equal by construction, so no fixture at this commit can distinguish the two
> choices — this picks the one that keeps working when that stops being true (4.5/4.8 are both about
> to add items into a row's span)."*

**This story is the fixture that can distinguish them.** If a footer joins an existing group, the
union extent and any single member's chrome rect diverge for the first time in the programme.

### The aggregate evaluation that already exists, and must not be duplicated

`internal/expr/aggregate.go`'s `evalSum` / `evalCount` / `evalAvg` are Story 3.3's one
implementation. `sum`/`avg` route through `SumDecimals`/`AvgDecimals` (`reduce.go`) and
`routing_arch_test.go` asserts that routing structurally, red-proved by
`TestSumRoutingRedProofInlineAccumulator`. That file's own header names this story:

> *"the SAME kernel … DW-7/D-1.4.1 requires every aggregate caller — this expression, and **Story
> 4.5's table footer** — to share"*

`count` calls `CollectionLength` and nothing else, and is a property of the collection, not of a
projection — so a footer `count` counts the table's own bound collection (its `bind` with `[]`
stripped), and `footerOf` is a load error alongside it. `expr.Eval` is reached from `bind.Resolve`,
which is the **one** display-text path every other cell already uses, including `formatNumber`
(D-1.4.1: *"`footerFormat` is data handed to the **existing** `formatNumber` … Absent and underived,
the footer renders unformatted through the one display-text function in `internal/bind`"*).

**Whether the footer's value is produced by synthesising an expression through that one path, or by
another route that provably reaches the same evaluation, is the developer's call** (D-000.79 §2 — a
mechanism named from outside the code over-constrains the design). What is *not* optional is AC4's
proof that it did.

---

## R — design constraints this story inherits (outcomes, not mechanisms)

- **R1 — `MixedItemError` is untouched and is not widened.** Exactly one of `{Runs, Images, Rects}`
  per item. A footer row is chrome **and** value text: two kinds, so at least two items — never
  collapsed into one, and a row is never collapsed into a single item.
- **R2 — the content column is never mutated.** No gap inserted, no sibling's declared `y` changed
  (AD-24: *"siblings never move, because nothing in a band ever reflows"*).
- **R3 — one aggregate implementation.** `SumDecimals`/`AvgDecimals` in `internal/expr` are the one
  kernel; D-3.1a.3's reducer inventory guards it. **No summation in `internal/layout`, none in
  `package folio`.** D-1.4.1 names the failure mode: *"that is exactly how the golden report's total
  drifts from a hand-computed one."*
- **R4 — exact decimal only, and a float conversion compiles.** AD-23/AD-1: no `float32`, `float64`,
  `math/big.Float`, no `math` transcendentals, no `time`, `os`, `math/rand`, `net`, no package-level
  mutable state. `lint`'s `TestFloatTypedProductionScan` is the instrument. `avg` uses Story 3.3's
  declared scale (max operand scale + `avgExtraScale`) and round-half-to-even; **no second rounding
  mode and no second scale decision at a call site** (D-3.1a.2).
- **R5 — no map iteration may reach output order** (D-1.3.5; `lint`'s `TestMapRangeProductionScan`).
  Which page the footer lands on, and the order its cells and value runs are emitted in, is decided
  by a slice walk. A map **lookup** is fine; a **range** is not.
- **R6 — no new glyphs producer.** `buildShapedPDFRuns` stays the only `pagemodel.ShapedGlyph`
  producer (D-16). The footer's value is new text and will legitimately shape new glyphs — through
  that one producer, via the same `shapeSegments` path every cell uses.
- **R7 — AD-14 holds.** Never a panic. Absent = Error; explicit `null` = empty and not an error;
  wrong kind = Error and never coercion. Reuse existing codes (see the section above).
- **R8 — invent no schema field.** `columns[].footer`, `columns[].footerOf`, `columns[].footerFormat`
  are the whole surface. There is **no `footerStyle`** and **no `footerHeight`** — see "Things the
  schema and the record could not resolve".
- **R9 — no golden fixture may move.** A new fixture is fine. Verify rather than trust: at creation
  the only repository template carrying a `table` element is
  `testdata/template/golden/worked-example.json`, which is a **template round-trip** golden, never
  rendered to a PDF golden — and it **does** declare a `sum` footer on column `e4` (proved by
  `folio_expr_validate_test.go:389`). Re-confirm nothing renders it.
- **R10 — no new gate obligation.** `TestEpic2GateObligationsMatchTheDeclaredSet` correctly refused
  an unauthorised one in 4.3. **No new `//go:build matrix` file** (DW-21's own grounds).
- **R11 — these guards stay green:** `lint`'s `TestGlyphIdentifierCensus` and
  `TestSpineStageLadderMatchesStageRankTable`; `folio-go`'s `TestFolioMethodNamesAreInjective`,
  `TestValidateNeverReachesRenderOrInternalPDF`, the one-byte-producer guard,
  `TestEpic2GateObligationsMatchTheDeclaredSet`, and
  `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` (a table in the `pageHeader` band is repeated
  verbatim and never paginates — a different feature with a confusingly similar name; **it is not a
  bug to be "fixed" into this story's mechanism**).
- **R12 — heavy tests are written with a REAL BODY and a CONDITIONAL gate.** D-000.4's per-epic
  cadence, and Story 4.4's Blocker 1: two heavy tests shipped as **empty bodies under unconditional
  `t.Skip`**, reported as *"written, not run"*. DW-21 is the shipped pattern — `FOLIO_HEAVY=1`, an
  ordinary env var, never a build tag. **Whatever heavy test this story writes must be appended to
  DW-21's recorded command in `deferred-work.md`, in the same commit**, or the Epic 4 gate will not
  run it.

---

## Acceptance Criteria

Every AC below that produces observable output carries a **named deletion**, stated **behaviourally,
never as a mechanism**. D-000.79: the creator names it because the creator has not written the
implementation. **Deletion is the screen, not the replacement** — if deleting the behaviour outright
does not redden the suite, stop; there is no witness at all, and any modification-based red-proof run
afterwards is measuring something unrelated.

Every red-proof runs the **WHOLE suite**, never behind `-run` (D-000.9 extension), records its
command, its counts and the **named** tests that reddened, and restores the tree by hand between runs.

**Anchoring, once, for all ten (D-000.68).** The fixture's data must be chosen so that the rendered
**sum, count and average are pairwise distinct strings, and none of them equals the rendered text of
any data cell in the table**. A `count` of `9` in a table whose amounts include `9.00` is a witness
that proves nothing. Anchor on text the test owns; never on rect geometry alone (every footer cell
shares a data row's `x`/`w` by construction); never on a count of rects, runs or pages.

---

### AC1 — sum, count and average render per the column's configuration

**Given** a table with per-column footer aggregates
**When** it renders
**Then** a footer row is produced carrying, for each column that declares `footer`, that column's
aggregate value — `sum`, `count` or `avg` as declared
**And** the value is placed in that column's own cell, honouring **that column's `align`**
(D-1.4.1: alignment inherits `columns[].align`; there is no fourth field)
**And** a column that declares no `footer` carries no value (see DECISION-3 for its chrome).

**Deletion — part (a), the test of the test (D-000.80).**
> **Remove whatever renders a DATA row's cell text**, so the only text in the table body is the
> footer's. **AC1's test must STILL PASS.**

Why it is needed, stated so it is not skipped as ceremony: a footer's rendered value is a short
numeric string in a table full of short numeric strings. Part (a) is the only thing that distinguishes
*"the test observes the footer's value"* from *"the test observes a data cell that happens to carry
the same characters."* **Part (a) passing is the load-bearing half.**

**Deletion — part (b), on top of (a).**
> **Remove whatever emits the footer row's values. The suite must REDDEN.**

A green (b) after a failed (a) is not evidence. Run (a) first.

---

### AC2 — `footerFormat` is applied, through the one existing formatter

**Given** a column whose `footerFormat` is set — explicitly, or derived from a shape-2 `bind`
**When** the footer value renders
**Then** it is formatted by that pattern
**And** the pattern is handed to the **existing** `formatNumber`, not to a second formatter.

Anchor: choose a pattern and a value whose formatted and unformatted spellings differ by more than
whitespace (measured shape: `#,##0.00` over a total of `1234.5` → `1,234.50`, versus `1234.5`). Cover
**both** arrivals of the pattern: explicit `footerFormat`, and the shape-2 derivation the probe
measured (`{{formatNumber(row.b, "#,##0.00")}}` → `FooterFormat:#,##0.00 HasFooterFormat:true`).

**Deletion.**
> **Remove whatever applies the column's `footerFormat` to the computed footer value, leaving the
> value rendered. The suite must REDDEN.**

**No part (a) is claimed**, and the reason is stated rather than assumed: unformatted and formatted
differ by construction in the fixture, so there is no accidental cause to remove. If during
development the two spellings coincide for some value, that is a fixture defect, not a passing test.

---

### AC3 — the aggregate covers the whole collection, matching Story 3.3

**Given** the aggregate values
**When** they are computed
**Then** they cover the **whole** bound collection, never the rows placed on any one page
**And** they equal, exactly, the value the same aggregate written as an ordinary `{{ }}` expression
over the same path produces for the same data.

**In-test presence preconditions, asserted rather than assumed** (this is what keeps the AC from
being vacuous, and it is why no part (a) is claimed):

- the fixture **paginates** — assert `len(Pages) >= 2` and fail loudly if not;
- the rows on the footer's **own page** do **not** aggregate to the same value as the whole
  collection — assert that inequality from literals the test owns, so a per-page implementation
  cannot coincide with a whole-collection one.

**Deletion.**
> **Remove whatever computes the footer's value from the bound collection. The suite must REDDEN** —
> at AC3's own assertion, not only at AC1's.

**Precision mutation, run in addition** (D-000.79: deletion screens, modification measures):
> **Make the footer aggregate only the rows placed on the footer's own page. The suite must REDDEN.**
> With the preconditions above in place this cannot pass; without them it silently can.

---

### AC4 — the footer uses the SAME aggregate evaluation as `{{sum(...)}}` / `{{avg(...)}}` / `{{count(...)}}` — DW-7, due here by name

**Given** a footer aggregate
**When** its value is computed
**Then** it is produced by Story 3.3's one aggregate evaluation, routing `sum`/`avg` through
`SumDecimals`/`AvgDecimals`
**And** no second summation, averaging or counting implementation exists anywhere in the module.

**This is the whole of DW-7, whose owner is this story by name and whose anti-rot mechanism is, in
its own words, "none possible for the footer half".** Nothing will fire if this is skipped — which is
exactly why it is an AC and not a note.

**Retire DW-7 by REPLACEMENT, in the same commit** — DW-6's own discipline under D-000.59, quoted
there: *"Replace, never merely delete."* The replacement is a **positive** assertion for the footer
half, in the same shape Story 3.3 shipped for the `{{ }}` half
(`TestSumRoutesThroughSumDecimals`/`TestAvgRoutesThroughAvgDecimals`, `routing_arch_test.go`) — plus
the behavioural equality AC3 asserts. Update DW-7's entry to say what landed and where; do not delete
the entry.

**Deletion** (this AC's screen is a mutation, and the reason is stated):
> **Replace the footer's route into the shared aggregate evaluation with an inline accumulator that
> produces an equal-looking answer. The suite must REDDEN.**

A pure deletion of the footer's value is AC1's deletion and screens nothing here: the defect this AC
exists to catch is a **second implementation that agrees on easy data and drifts on hard data**.
`TestSumRoutingRedProofInlineAccumulator` is the shipped precedent for exactly this mutation — and
note the **stated bound** recorded in `aggregate.go`: an AST routing scan proves the call is
*present*, not that its *result is used*. Do not let the footer half be weaker than the behavioural
equality in AC3.

---

### AC5 — a footer that does not fit moves to the next page **together with at least one preceding data row**

**Given** a footer aggregate row that does not fit in the remaining space
**When** pagination runs
**Then** it moves to the next page **together with at least one preceding data row**
**And** the data row that moves with it is the row that immediately precedes it in the collection
**And** no page ever carries a footer row as its only content from that table.

**This is the first orphan rule in the programme, and its shape is a relationship between two groups,
not a property of one.** Story 4.3 built row atomicity ("the footer moves whole"); that is **not**
this rule and satisfying it does not satisfy this.

**Anchoring, by value, from the measured fixture (re-derive it, do not copy it):** the 9-row fixture
above paginates to **one** page at HEAD and to **two** with a footer, under **either** implementation.
Naive: page 0 = rows 0..8, page 1 = footer alone. Correct: page 0 = rows 0..7, page 1 = **row 8** +
footer. **Two pages either way** — assert the **per-page row-index partition and the footer's page by
value**, and assert the **identity** (`rowIndex == 8`) of the row that came with it. "Some data row is
on the footer's page" is not a witness; on a wider fixture it passes for unrelated reasons.

**In-test presence precondition:** assert that the preceding data row **fits its original page with
room to spare** (measured: 1,936mp of slack) — i.e. that it would not have moved on its own. Without
this the AC is satisfiable by a footer and a row that both simply failed to fit.

**Deletion — part (a), the test of the test (D-000.80).**
> **Remove whatever draws the footer row's own cell chrome, leaving its value runs.** **AC5's test
> must STILL PASS.**

This is D-000.80's founding accident, aimed at the one place it can still bite: a row's chrome rect is
a single item spanning the whole row, it fails the fit test first, and it drags the row's text along
as an *atomic proxy*. A footer row built like a data row inherits that accident exactly. Part (a) is
the only thing that distinguishes *"the footer is tied to the preceding row"* from *"the footer's
chrome happens to be an atomic proxy for something."* **Part (a) passing is the load-bearing half.**

**Deletion — part (b), on top of (a).**
> **Remove whatever keeps the footer with a preceding data row when the footer does not fit. The
> suite must REDDEN** — at the partition assertion and at the moved row's identity, **not** at a page
> count.

---

### AC6 — the footer travels with its own table's displacement, and with nothing else's

**Given** a footer row that lands on a page where its table's header repeats (Story 4.4)
**When** that page renders
**Then** the repeated header is above that page's first table row, that row is above the footer, and
**nothing overlaps**
**And** the footer receives the **same** per-table displacement that table's rows receive on that page
**And** no element belonging to any other table, and no non-table sibling on that page, moves at all.

**Grounds, and they are one story old.** Story 4.4 shipped three Class B defects, each proven by a
mutation that left the suite green: a repeat displaced by 37pt with nothing pinning its position; a
displacement that leaked onto **every** element on the page; and a whole diagnostic construction block
deleted unnoticed. Two of those shapes are directly reachable here. **Assert positions by value —
where things actually landed — never that a bookkeeping field was populated**, and include a
**negative** witness (an unrelated sibling's drawn `y`, unchanged) in the same test.

**Deletion — the positive half.**
> **Remove whatever applies the table's per-page row displacement to the footer, leaving it applied
> to that table's data rows. The suite must REDDEN** — with the footer overlapping the last data row
> or the repeated header, observed as a position.

**Deletion — the negative half.**
> **Remove whatever scopes that displacement to the one table, so it applies to every element on the
> page. The suite must REDDEN** at the unrelated sibling's position.

---

### AC7 — both pagination passes agree, so `Page X of Y` stays true

**Given** a document with a paginating table carrying a footer, and a `{{page}} of {{pages}}` construct
in a repeated band
**When** it renders
**Then** the page count the page-count-only pass computes equals the count the final pass produces
**And** the per-page partition — data rows **and the footer's page** — is identical between the two.

The fixture must be one where the orphan rule **changes the partition** (the 9-row fixture is such a
fixture), otherwise the guard is satisfied by two passes that were going to agree anyway.
`TestBothPaginationPassesAgreeOnRowPartition` (`page_number_test.go`) is the existing instrument and
the right place to extend. D-4.3.2: **two passes silently disagreeing is the defect.** The two
builders append in different orders and agree only because both carry the same group identity — a
footer that carries an identity in one builder and not the other makes `Y` wrong.

**Deletion.**
> **Remove whatever makes the page-count-only pass see the footer, leaving the final pass seeing it.
> The suite must REDDEN**, and on the **partition**, not merely on a count.

---

### AC8 — it holds through the public `Render()`, and this AC's LAYER is pinned

**Given** the same multi-page table document
**When** it is rendered through the public `Render()` entry point
**Then** the produced PDF carries the footer's formatted value text on the page the footer landed on,
and on no other page
**And** the document's page objects and `/Count` agree with the page count the page-count-only pass
produced.

**This assertion lives at the `Render()` layer and may not later be narrowed to a layout-level check.**
The failure mode is narrowing, not deletion. **Grounds:** Story 4.3 shipped a **live regression** — a
table beside any element sharing its `y` became an unrenderable internal error — and it escaped review
because **no test in that story called `Render()`**.
`TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter` and
`TestTableBesideSameYElementRenders` are the shape to follow.

**Deletion.**
> **Remove whatever carries the footer row from the pagination result into the composed page model.
> The suite must REDDEN at the `Render()` level** — a layout-level test reddening is not a substitute.

This is a *separate* deletion from AC1's on purpose: it screens the plumbing independently of the
computation.

---

### AC9 — a table bound to an empty collection still renders, and says so honestly

**Given** a table bound to an **empty** collection whose columns declare footers
**When** the document renders
**Then** the render **succeeds** and returns PDF bytes
**And** a `sum` footer renders `0` and a `count` footer renders `0` — Story 3.3's values, at Story
3.3's scales, **not** harmonised with each other (D-3.1a.2: *"a developer who 'harmonises' the two
zeros has re-decided the scale at a call site"*)
**And** an `avg` footer produces the **existing** `DiagCodeEmptyAverage` **Warning** and an empty
cell — never an Error, never a new code, never a panic (AD-14, R9 of Story 3.3, and Story 4.2's own
empty-collection AC)
**And** pagination terminates.

This AC's shape depends on **DECISION-3**; if the lead rules that a footer row does not render at all
for an empty collection, the first three clauses change with it and the last two stand regardless.

**Deletion.**
> **Remove whatever produces a footer value when the bound collection is empty. The suite must
> REDDEN** — and it must redden on the rendered `0`s and on the presence of the Warning, not on an
> error being absent.

---

### AC10 — byte-neutrality: a table with no footer changes in no way

**Given** any table declaring `footer` on **no** column
**When** the document renders
**Then** it produces exactly what it produces at `530a47b` — same items, same partition, same bytes
**And** every committed golden is byte-identical
**And** `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` and every guard in R11 stays green.

Evidence must be a **re-render comparison**, not a hash of committed files — Story 4.4's review
Finding 7 caught exactly that substitution.

**No deletion is claimed for AC10, and the owner is named** rather than left as an untriggerable
"belongs to whoever built it": the behaviour this AC protects is **Story 4.1**'s (header chrome and
labels) and **Story 4.2**'s (data-row chrome and cell text), and each is screened by that story's own
deletions. A deletion here would delete another story's feature, which measures that story, not this
one.

---

## Decisions

### DECISION-1 — SURFACED, not taken: what shape carries "moves together with at least one preceding data row"?

The outcome is fixed by AC5. **The mechanism is the developer's call** (D-000.79 §2). Two shapes are
known to be viable against the measured pipeline; both are recorded with their trap, so neither has to
be rediscovered:

**(i) The footer joins the preceding data row's existing group.** Zero new machinery: `Paginate`
already tests a group's **union extent** and slides the window to the group's earliest `Top`, which is
precisely "both move together". It also makes the footer count as one of that table's rows for Story
4.4's reservation, so the repeated header appears above them — which AC6 wants. **Trap:** the union
extent and a single member's chrome rect diverge here for the first time (`paginate.go`'s own 4.3
comment predicted this story as the case), and a group taller than the window becomes an
`OverflowError` where the row alone would have fitted — that is DECISION-2.

**(ii) A "keep-with-previous" relation tested at the footer's own fit time.** More expressive (it can
relax when the tie is unsatisfiable) and more machinery, added inside the gating epic.

Whichever is chosen: **`MixedItemError` stays untouched**, a row is never collapsed into one item
(R1), and no map is ranged for emission order (R5).

### DECISION-2 — SURFACED: what happens when the preceding row and the footer together exceed the content window?

The orphan rule is then unsatisfiable. Three arms, and the lead's ruling is needed before development:

- **(a)** report the existing overflow error. No document renders a footer today, so this regresses
  nothing — but it turns a document that *would* render into a hard failure for a geometry the author
  cannot easily see.
- **(b)** honour the fit and place the footer alone, **recording** the fact — Story 4.4's DECISION-2
  precedent exactly (suppress on that one page, record it as data a caller turns into a located
  Warning, never silent, always terminating). **This is the arm this story recommends**, on that
  precedent and on AD-14's "never a panic, always terminate".
- **(c)** something else the lead prefers.

**If (b) is ruled, it is the one place in this story where minting a diagnostic code is legitimate** —
under D-000.65, because the condition first occurs here and nowhere earlier. It would be a **new**
condition, not `TABLE_FOOTER_SOURCE_*`, which name load-time source-derivation failures and must not
be stretched to cover a layout outcome: stretching is worse than minting, because it corrupts both
meanings. Story 4.4's `DiagCodeTableHeaderRepeatSuppressed` is the shape and the precedent.

### DECISION-3 — SURFACED: does a footer row render chrome for a column that declares no `footer`, and does it render at all for an empty collection?

D-1.4.1 says *"Omitted means no footer cell for that column"*, which is unambiguous about the **value**
and silent about the **chrome**. Every data row emits one rect per column **unconditionally**
(`table_render.go`, and D-000.80's own negative check confirmed it) — a footer row that emits chrome
only under the columns declaring an aggregate would draw a broken-looking partial row in any bordered
table. **Recommendation: chrome for every column, value only where `footer` is declared**, mirroring
the data-row precedent; and **the footer row renders for an empty collection too**, since it is a
property of the column's configuration rather than of the data, and Story 4.2's empty-collection AC
already requires the render to succeed. AC9 is written to this recommendation and moves with the
ruling.

### DECISION-4 — ANSWERED here, on grounds: the footer row's style and alignment

No ruling is needed and no field may be invented. **Alignment** inherits `columns[].align` — D-1.4.1
ruled it explicitly (*"no fourth field"*). **Style** cascades from the table's own `style`, exactly as
a data row's does: `headerStyle` governs the header row **exclusively** (D-000.76, Story 4.1, and the
schema's own text), so there is no `headerStyle` arm for a footer any more than there is for a data
row. There is **no `footerStyle`** and none is to be added: extending the schema is a MINOR-additive
question for the format, not a thing a story does mid-epic.

---

## Things the schema and the record could not resolve, surfaced rather than invented

1. **There is no `footerHeight`, and `headerHeight` exists.** The header's height is declared; a data
   row's is derived (padding + the body font's vertical model). The record never says which the footer
   follows. **Derive it like a data row** — that is the only option that invents no field — and note
   that this makes the footer's height data-dependent in exactly the way a data row's already is.
   Measured for the fixture above: a one-line body row is 10,896mp.
2. **The out-of-collection `footerOf` load error is UNCODED at HEAD.** D-1.4.1 says
   `TABLE_FOOTER_SOURCE_UNRESOLVED` covers *"underivable **or out-of-collection** source"*, but
   `internal/template/parse_bands.go`'s prefix check (`footerOf` must start with the table's
   collection path) returns a plain `newLoadError`, not `newLoadErrorCoded(..., CodeTableFooterSource
   Unresolved)` — unlike the two `FORBIDDEN` checks beside it, which are coded. **This is a Story 3.6
   absorption gap, not this story's work**, and it is recorded here rather than quietly fixed or
   quietly ignored. Recommend the lead route it (a deferred-work entry, or the Epic 4 boundary gate).
3. **`Template.derivedFooters` is written and read by nothing but tests at HEAD.** This story is its
   first consumer. If the developer finds it does not carry what the render path needs, that is a real
   gap in the 3.2 derivation's shape, not a licence to re-derive: re-derivation is what D-4.2.2 and
   D-1.4.1 both forbid.
4. **A `count` footer's operand.** `footerOf` is a load error alongside `count`, so the operand is the
   table's own `bind` with `[]` stripped. That is inference from two rulings, not a sentence anyone
   wrote; it is the only reading consistent with both.

---

## Do not re-open — settled rulings this story inherits

- **The two footer diagnostic codes are minted, DW-6 is retired, and nothing is minted here** except
  possibly under DECISION-2(b). See the dedicated section above.
- **Story 4.4's per-element displacement channel stays.** It was built at 4.4 precisely because 4.5
  and 4.6 need it. Do not fold it into the page-wide `Shift` (DECISION-3 of 4.4, ruled) and do not
  rebuild it.
- **`MixedItemError` is untouched** and a row is never one item (Story 4.3, R1).
- **`avg`'s scale and rounding mode are Story 3.3's and are not re-decided at a call site**
  (D-3.1a.2, D-3.3.x).
- **A table in the `pageHeader` band is a different feature** and is left alone (R11).
- **The heavy-test gate is an env var, never a build tag** (DW-21, on
  `TestEpic2GateObligationsMatchTheDeclaredSet`'s own refusal at 4.3).
- **Story 4.6 owns the row taller than the page; Story 4.8 owns alternating shading.** Neither is
  built here, and this story's shape must not force either to tear it out (4.4's R7, still binding).

---

## Task breakdown

1. **Reproduce the baseline.** Run all three gates, record the figures and the named skips with the
   counting convention. Confirm the tree is clean at `530a47b`.
2. **Re-run the creation probe.** Re-derive the 9-row arithmetic (slack 1,936mp against a footer row
   of ≈10,896mp) rather than copying it, and confirm at HEAD that a footer-configured table emits no
   footer item of any kind.
3. **Take DECISION-2 to the engineering lead before writing the mechanism**, and DECISION-3 with it.
   Record the rulings in the Delivery Log verbatim.
4. **Build the footer row** — its items, its group identity carried (never reconstructed), through
   **both** pagination builders. AC1, AC2, AC7.
5. **Wire the value to the one aggregate evaluation** and land DW-7's replacement assertion; update
   DW-7's entry in `deferred-work.md` in the same commit. AC3, AC4.
6. **Build the orphan rule** to the shape chosen under DECISION-1. AC5.
7. **Cover the 4.4 interaction** — displacement reaches the footer and nothing else. AC6.
8. **Assert through `Render()`**, and pin the layer in the test's own doc comment. AC8.
9. **Cover the empty collection.** AC9.
10. **Prove byte-neutrality by re-render**, not by hashing committed files. AC10.
11. **Write the heavy test with a real body under `FOLIO_HEAVY=1`**, confirm it SKIPs unset and PASSes
    set, and **append it to DW-21's recorded command**. Do not run it as part of the routine gate;
    say so explicitly in the Delivery Log (D-000.4).
12. **Run every named deletion above, whole-suite, restoring the tree by hand between runs**, and
    record for each: the command, the pass/fail/skip counts, and the **named** tests that reddened.
    Part (a) first, every time.
13. **Re-run the three gates**, record the final figures, and set the story's status to `review`.

---

## Delivery Log

_To be completed by the developer. It must contain, at minimum: the reproduced baseline; the
re-derived probe arithmetic; the rulings received for DECISION-1/2/3 quoted verbatim; per-AC
verification; every named deletion with its command, counts and reddened test names (part (a) results
included, and named as such); the heavy test's both-directions confirmation; the byte-neutrality
evidence; and an explicit statement of what was deliberately not built._

---

## Delivery Log

### 1. Baseline, reproduced at HEAD `530a47b` (tree clean apart from the story file and `sprint-status.yaml`)

Counting convention as the story states it: `--- PASS` / `--- FAIL` / `--- SKIP` lines anywhere in `-v`
output, subtests included. **Skips reported BY NAME, never by count (D-000.81).**

| Gate | Measured at baseline |
|---|---|
| `folio-go` | **672 pass · 0 fail · 3 skip** |
| `lint` | **77 pass · 0 fail · 0 skip** |
| `hashmatrix` | **3 pass · 0 fail · 0 skip** |

The three baseline skips, named:

```
--- SKIP: TestXrefEntriesRejectsMalformedSubprocess
--- SKIP: TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable      (DW-21, FOLIO_HEAVY=1)
--- SKIP: TestTwoTablesWithPageCountFooterRenderConsistently          (DW-21, FOLIO_HEAVY=1)
```

**WITHDRAWN BY THE FINISHER (review Finding 7). The paragraph that stood here claimed the story's
recorded `lint` figure of 115 was a discrepancy against a measured 77.** It is not a discrepancy; it
is a UNIT MISMATCH, and the developer's own declared convention above is the one that was not
followed. Re-measured in the finisher pass under the exact command the story gives:

| Gate | top-level (`^--- PASS`) | including subtests (`--- PASS` anywhere) |
|---|---|---|
| `lint` | **77** | **115** |
| `folio-go` (at HEAD `530a47b`) | **672** | **1022** |

Both figures are correct and count different things. The story's table counted subtests-inclusive —
its own stated convention — and the developer counted top-level while declaring the other. Nothing
moved in the tree, and no correction was owed. **Every figure in §9 and §14 below was likewise counted
TOP-LEVEL, and is now labelled as such rather than left bare.** Per the sharpened D-000.83 guidance:
where the set is small enough to print (skips, red-by-design), the SET is reported and no count is
used at all; a units-labelled scalar appears only where the set is too large to print.

`TestCorpusMeetsP6ExerciseFloors` is excluded by `-skip` and prints no line at all. **Named separately
per D-000.81:** it is red by design (D-000.17 / D-2.1.14 / D-000.57 / D-000.74) and was re-measured
unchanged by this story — `P6g (opaque names) floor not met: got 7, need >=20`, identical to baseline.
It is never to be "fixed".

### 2. Re-derived probe arithmetic (not copied)

A throwaway probe was written, run, and deleted. It confirmed at HEAD that a footer-configured table
emits **no footer item of any kind**, and re-derived the geometry independently:

```
header row: top 10,000  bottom 20,000     (declared headerHeight)
data row i: top 20,000 + i*10,896         height 10,896mp
9 rows:     last row ends at 118,064 against a ceiling of 120,000 — 1,936mp slack
```

A footer row of ≈10,896mp cannot fit in 1,936mp, so the 9-row fixture is the orphan case, needing no
continuation page and no Story 4.4 reservation to create the squeeze. **Confirmed: 2 pages either
way; only the partition moves, 9/0 → 8/1.** Page count therefore carries nothing here, exactly as the
story predicted, which is why AC5 asserts the partition and the moved row's identity by value.

### 3. Rulings received, and what they changed

- **D-4.5.1 — mint, two codes.** The discriminator ruled: *"Two conditions share a code only if the
  author would take the same action AND the same thing happened to their document. Same remedy is not
  sufficient."* Applied: Story 4.4 **drops** a declared element; this story **relocates** one. Minted
  `TABLE_FOOTER_ORPHAN_SUPPRESSED`. **This overturned my stated pre-ruling default** (let the existing
  `OverflowError` arise, mint nothing) — corrected mid-story, as the ruling requires.
- **D-4.5.2 — the footer-alone hole must be named, not left undefined.** Answered below (§6).
- **D-4.5.3 — zero by decision, with a test that fails if it is zero by fallthrough.** Answered below (§7).
- **D-4.5.4 — the story that carves the exception fences it, in 4.5, red-proofed.** Answered below (§5).

### 4. DECISION-1 — RE-DERIVED, not defended. The "zero new machinery" reason is void.

**Shape chosen: neither (i) nor (ii) as written — a PROBE-THEN-MERGE at the `folio` package level**
(`folio-go/table_footer.go`, `paginateWithFooterOrphanFix`). `internal/layout/paginate.go` is
**untouched by this story**, which the coordinator correctly inferred from its mtime; stating it
explicitly as asked.

**How it works.** `layout.Paginate` runs once as a probe. If — and only if — that probe would orphan
the footer (place it on a later page than the row immediately preceding it), `Paginate` runs a second
time with the footer's group Key temporarily redirected to the preceding row's Key, for that one call
only. The footer's own group identity (`ItemGroupKey.Index == footerGroupIndex`, i.e. `-1`) is never
permanently merged.

**Which reason carried it NOW.** My first pass did pick shape (i) on the "zero new machinery" grounds.
**D-4.5.1's "never error" ruling voided that reason outright**, and this is the re-derivation:

> Under an unconditional merge, an over-tall joined group reaches `Paginate`'s `OverflowError` with only
> two exits, and the ruling closes both. Erroring is ruled out. Bypassing `OverflowError` generically
> inside the grouping path would **also swallow Story 4.6's over-tall single data row** — manufacturing
> in 4.5 exactly the accidental cause 4.6 would then measure against, which is D-000.80's founding
> accident reproduced one story after building the screen for it.

The probe-then-merge shape keeps the exception **keyed on footer-ness by construction**: the merged
grouping is something only `table_footer.go` ever asks `Paginate` for, so the never-error catch is
reachable only from this file's own merge attempt. An over-tall single data row never enters that
second call and behaves exactly as at HEAD. The common case costs one extra `Paginate` call and only
when a real orphan is detected — the rare case, not the common one.

**Cost accepted and stated:** one additional `Paginate` pass in the orphan case, against a permanently
narrower blast radius for the carve-out.

### 5. Fact 2 / the D-4.5.4 fence — never-error stays keyed on footer-ness. **NOT trivially green.**

`TestOverTallSingleRowStillOverflows` asserts an over-tall item that is not a footer-plus-row still
produces `layout.OverflowError`.

**Red-proofed by widening the bypass**, exactly as D-4.5.4 requires — the catch was broadened to match
any `OverflowError` rather than only one from this file's own merge attempt:

```
env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors$ -v ./...
→ 681 pass · 6 fail
--- FAIL: TestOverTallSingleRowStillOverflows                       (the fence itself)
--- FAIL: TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs
--- FAIL: TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender   (Story 4.3's own)
--- FAIL: TestRenderReportsAnItemThatFitsOnNoPage
--- FAIL: TestFourErrorModesCarrySeverityErrorDiagnostics
--- FAIL: TestMessageRewriteDoesNotAffectCodeRecovery
```

**The fence is load-bearing, not trivially green.** Restored; whole suite green afterwards.

### 6. D-4.5.2 — the footer-alone hole, named explicitly

**Declared Story 4.6's subject; 4.5 terminates without placing.** When the footer BY ITSELF exceeds the
content window, pass 1 of `paginateWithFooterOrphanFix` returns `layout.OverflowError` directly, before
any merge is attempted — the existing FR44/`CONTENT_UNLAYOUTABLE` route, never a hang and never an
undefined state. Asserted with a 5-second timeout guard by
`TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs`. This is stated in code
(`table_footer.go`'s package comment) as well as here, per the ruling.

### 7. D-4.5.3 — zero BY DECISION, with a test that fails if it is zero by fallthrough

`sum`→`0.00` and `count`→`0` over an empty collection come from Story 3.3's real evaluation
(`SumDecimals(nil)` / `CollectionLength`), at their OWN scales, deliberately not harmonised with each
other (D-3.1a.2). `avg` produces the **existing** `DiagCodeEmptyAverage` Warning and an empty cell.

**The anti-fallthrough guardrail**, which is the ruling's actual point:
`TestFooterOverEmptyCollectionRendersAndTerminates` asserts the empty-average Warning carries
`DataPath == "items.c"` — **the column's real DERIVED `footerOf`**, which only an implementation that
actually evaluated the expression can know. A hardcoded `0`-returning fallthrough produces a
plausible-looking `0.00` but cannot produce that path.

**Mechanism note, recorded because it is load-bearing and non-obvious.** `avg` over an empty collection
resolves to `expr.KindNull`, and `evalFormatNumber` hard-errors on a non-number operand. Wrapping it
would have turned AC9's Warning into a render-aborting Error — the exact AD-14 violation AC9 forbids.
So for the empty-collection `avg` case alone, the synthesised expression omits `formatNumber` and
resolves bare, letting `bind.Resolve`'s existing `KindNull` → empty-string path carry it. `sum`/`count`
are unaffected: both yield a real `KindNumber` zero and always go through `formatNumber`.

### 8. Fact 1 — Story 4.8's alternating row background does NOT see the footer as a row

**Answered by reading what 4.8 will key striping off** (`epics.md`, Story 4.8, verbatim): *"the
alternation follows **row index in the collection**, so it does not reset per page."*

Row index in the collection is carried by the **row-type tags** `isDataRow`/`rowIndex`
(`tableRectSource`) and `isTableRowLine`/`rowIndex` (`textRunSource`) — **not** by
`layout.ItemGroup.Key`, which is a pagination-time grouping concept. This story gives the footer
**distinct row-type tags** (`isFooterRow` / `isFooterLine`), never setting `isDataRow` or
`isTableRowLine` on it, in **both** cases:

- the non-orphan common case (footer keeps its own group Key, `Index -1`);
- the orphan case (the Key is redirected to the preceding row's Key **for one `Paginate` call only**,
  while `isFooterRow`/`isFooterLine` never change).

So 4.8 cannot mistake the footer for a row under either case. **Measurement backing it:**
`TestFooterlessTableCarriesNoFooterRow` asserts a footer-less table produces no `isFooterRow` source
and zero orphan targets; the AC6 test reads row 8 and the footer as separately-identified rows at
distinct measured Y values (20,000 vs 30,896).

### 9. Per-AC verification, and every named deletion with its result

Every red-proof ran the **WHOLE suite**, never behind `-run`, restoring the tree between runs
(verified with `git diff` / a pristine backup each time).

> **UNITS (finisher, review Finding 7): every pass/fail figure in the table below is counted
> TOP-LEVEL (`^--- PASS`), against a top-level baseline of 688 (1038 including subtests).** The
> reddened test NAMES, not the counts, are the durable record — a count moves whenever the file's
> test count moves, and this story's finisher pass added ten tests. **§18 below re-runs every
> deletion this story now depends on against the FINAL baseline**, so the two tables are not to be
> compared figure-for-figure. Where a row below was neither re-run nor invalidated (AC2, AC5, AC6,
> AC8, AC10), the reddened names still stand.

| AC | How proven | Named deletion | Result |
|---|---|---|---|
| **AC1** | `TestFooterAggregatesRenderPerColumnConfiguration` (sum/count/avg all render), `TestFooterHonoursColumnAlign` (drawn X differs left vs right, by value) | **(a)** remove data-row cell text | **668 pass · 18 fail — AC1's test STILL PASSED** ✅ load-bearing half |
| | | **(b)** on top of (a), remove footer value emission | **661 pass · 25 fail** — `TestFooterAggregatesRenderPerColumnConfiguration` reddened |
| **AC2** | `TestFooterFormatAppliesThroughExistingFormatter` — both arrivals: explicit `footerFormat` (e3) and shape-2 derived (e4) | apply `footerFormat` to nothing, leave value rendered | **678 pass · 8 fail** — `TestFooterFormatAppliesThroughExistingFormatter` reddened |
| **AC3** | `TestFooterAggregateCoversWholeCollectionNotJustItsOwnPage`, with `len(Pages)>=2` asserted and the footer-page inequality anchored on owned literals | compute footer from nothing | **679 pass · 7 fail** — AC3's own test reddened |
| | | **precision:** aggregate only the footer's own page | **678 pass · 8 fail** — AC3's own test reddened |
| **AC4** | `TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression` — footer output byte-identical to an independently-evaluated author-written `{{sum/count/avg}}` | add a rival inline accumulator shaped `[]Decimal → (Decimal, error)` | **685 pass · 1 fail** — `TestDecimalReducerInventoryIsExactlySumAndAvg` reddened (a third reducer where exactly two are permitted) |
| **AC5** | `TestFooterOrphanTieMovesWithImmediatelyPrecedingRow` — partition and `rowIndex==8` identity by value; `TestFooterOrphanTiePresencePrecondition` — row 8 fits page 0 unaided | **(a)** delete the footer's own chrome | **686 pass · 0 fail — STILL PASSED** ✅ the tie is not a chrome-as-atomic-proxy artifact |
| | | **(b)** on top of (a), remove the tie | **685 pass · 1 fail** — reddened at *"row 8 … not moved with the footer"* and *"row 8 is not on page 1"*, **not** at a page count (2 either way) |
| **AC6** | `TestFooterTravelsWithItsOwnTablesDisplacementAndNothingElses` — header 10,000 / row 8 20,000 / footer 30,896 / sibling 12,832, all by value, plus non-overlap on chrome extents | **positive:** footer excluded from displacement | **686 pass · 1 fail** — footer Y=**20896** vs want 30896, i.e. *overlapping row 8*, observed as a position |
| | | **negative:** displacement applied page-wide | **685 pass · 2 fail** — sibling Y=**22832** vs want 12832; Story 4.4's own `TestSiblingPositionUnaffectedByTableHeaderRepeatThroughPageModel` also reddened |
| **AC7** | `TestBothPaginationPassesAgreeOnFooterPartition` | remove the footer from PHASE A | **686 pass · 1 fail** — *"page counts disagree: PHASE A=1 PHASE B=2"* |
| **AC8** | `TestFooterHoldsThroughPublicRender` — through `Render()`, footer text on its page and no other, `/Type /Page` count and `/Count` agree; layer pinned in the test's own doc comment | remove the footer's carry into the composed page model | **678 pass · 9 fail** — `TestFooterHoldsThroughPublicRender` reddened; **`TestBothPaginationPassesAgreeOnFooterPartition` and `TestFooterOrphanTiePresencePrecondition` stayed GREEN** ✅ discriminates as required |
| **AC9** | `TestFooterOverEmptyCollectionRendersAndTerminates` — bytes returned, `0`/`0.00` at own scales, `DiagCodeEmptyAverage` Warning with the real derived DataPath | remove footer value production over an empty collection | **686 pass · 1 fail** — reddened on the rendered `0`s **and** on the Warning's presence, not on an error being absent |
| **AC10** | `TestFooterlessTableCarriesNoFooterRow` — no `isFooterRow` source, zero orphan targets, and a **re-render** byte comparison | none claimed (owner named: Stories 4.1/4.2, per the AC) | n/a |

**AC6's honest note.** AC6's positive mutation was run **before** its witness existed and left the
suite green — i.e. the feature was briefly built with no witness at all, which is Story 4.4's Blocker 3
shape. The test above was written in response, and the mutation re-run against it. Recorded rather
than presented as first-time-right.

**AC7's honest note.** The first AC7 mutation (remove only the footer's *chrome* from PHASE A) left the
suite green, because the footer's group identity is carried on **both** its rects and its runs. That is
a correct property, not a hole — the mutation was aimed wrongly. Re-aimed to remove the footer from
PHASE A entirely; it then reddened on the partition as required.

**One design change that AC5(a) forced.** `footerOrphanTargetsFrom` originally read footer-ness out of
`tableRects` (the chrome producer). AC5's part (a) is aimed at exactly that, so the function was
rewritten to read the targets off the `layout.ColumnItem`s themselves, keyed on `Group.Key` — which both
the chrome item and the value-line items carry identically. The tie now survives the footer's chrome
being deleted, which is what made part (a) pass honestly rather than by luck.

### 10. Heavy test — both directions confirmed

`TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability` (`folio-go/table_footer_test.go`): a
**real body** — 500 rows through public `Render()`, footer sum appears exactly once and on the last page
only, byte-stable across two renders, 60s bound. Gated by `os.Getenv("FOLIO_HEAVY") != "1"`, an env var,
**never a build tag** (R10/DW-21: a new `//go:build matrix` file would register as an unauthorised gate
obligation).

```
unset          → --- SKIP: TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability
FOLIO_HEAVY=1  → --- PASS: TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability (0.09s)
```

**Appended to DW-21's recorded command in the same change**, in both places that carry it:
`deferred-work.md`'s DW-21 entry and `table_header_repeat_test.go`'s `heavyTestGateEnvVar` doc comment.
**Per D-000.4 it was NOT run as part of the routine gate** — it reports `--- SKIP` in the final figures
above, and that is deliberate.

**Unrun suites, named explicitly (D-000.4's per-epic cadence):**

```
TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable                (Story 4.4, DW-21)
TestTwoTablesWithPageCountFooterRenderConsistently                    (Story 4.4, DW-21)
TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability        (Story 4.5, DW-21 — this story)
TestXrefEntriesRejectsMalformedSubprocess                             (pre-existing, unrelated)
TestCorpusMeetsP6ExerciseFloors                                       (red by design, -skip excluded)
```

### 11. DW-7 retired BY REPLACEMENT

`deferred-work.md`'s DW-7 entry is **appended to, never edited in place** (D-000.29/D-3.1.1). The
replacement is a positive routing assertion for the footer half in the same shape Story 3.3 shipped for
the `{{...}}` half, plus AC3's behavioural equality, plus the captured red-proof in the table above.
The entry now records what landed, where, and that no further owner is due.

### 12. D-000.67 part 2 sweep of `parse_bands.go` — **count of sites examined**

**61 load-error sites examined** in `internal/template/parse_bands.go`:

- **3 coded** — the two pre-existing `TABLE_FOOTER_SOURCE_FORBIDDEN` sites (Story 3.6), **plus one
  swept in by this story**: the out-of-collection `footerOf` prefix check at line 450, which D-1.4.1
  says `TABLE_FOOTER_SOURCE_UNRESOLVED` covers (*"underivable **or out-of-collection** source"*) but
  which Story 3.6 left as a plain `newLoadError` while coding the two checks beside it. A genuine
  absorption gap, closed rather than re-recorded. Pinned by
  `TestOutOfCollectionFooterOfCarriesTheUnresolvedCode`.
- **58 uncoded**, of which **4 are footer-related**. **All four are correctly uncoded, but they are
  NOT all the same kind of failure, and the reason recorded here first was wrong for one of them
  (review Finding 8, corrected by the finisher):**
  - **Three are TYPE failures** — the `decodeStringRaw` guards on `footer`, `footerOf` and
    `footerFormat` (`must be a string: …`). Cited by the declaration they sit in rather than by line
    number, which went stale once already in this programme.
  - **One is a CLOSED-SET / value-domain failure** — the `closedFooterKinds` check, which rejects a
    well-formed string carrying a footer KIND outside `{sum, count, avg}`. The ground for leaving it
    uncoded is not "it is a type failure" (it is not): it is that `TABLE_FOOTER_SOURCE_*` names a
    failure of the footer's numeric SOURCE — which collection path its value comes from — and a bad
    KIND is a different statement, so coding it with either would corrupt both meanings. The
    asymmetry the sweep should have surfaced is that its immediate neighbour, the
    `footerOf`/`footerFormat`-present-with-no-`footer` check, IS coded `TABLE_FOOTER_SOURCE_FORBIDDEN`
    — because that one really is about the source.
  - The ground is now stated **at the call site in `parse_bands.go` as well as here**, so the next
    D-000.67 sweep does not re-litigate it.
  All four surface as `TEMPLATE_MALFORMED` through `wrapTemplateError`, like the other 54.

Three now-stale doc comments naming "two sites" were corrected (`errors.go` ×3, `render_error.go` ×1).

### 13. Byte-neutrality (AC10) — evidence is a RE-RENDER, not a hash of committed files

- `TestFooterlessTableCarriesNoFooterRow` renders a footer-less table **twice** and compares bytes;
  it also asserts no `isFooterRow` source and zero orphan targets are produced for it.
- **No committed golden moved.** `git status --short -- folio-go/testdata fixtures hashmatrix` is
  **empty**. Every golden test is itself a re-render comparison and all ran green, including
  `TestMultiPageGoldenMatchesTheCommittedArtifact`, `TestRenderMatchesGoldenFixture`,
  `TestThreeBandPageGoldenFixture`, `TestGoldenDigestAgreesAtEveryDeclaredSite`,
  `TestEveryGoldenPDFResolvesItsPageTree`, and both cross-process byte-identity tests.
- R11's guards all green: `TestGlyphIdentifierCensus`, `TestSpineStageLadderMatchesStageRankTable`,
  `TestFolioMethodNamesAreInjective`, `TestValidateNeverReachesRenderOrInternalPDF`, the
  one-byte-producer guard, `TestEpic2GateObligationsMatchTheDeclaredSet`,
  `TestTableInPageHeaderRepeatsIdenticallyAcrossPages`.

### 14. Gates as the DEVELOPER measured them — SUPERSEDED by §18

> **Superseded (finisher).** The figures below are TOP-LEVEL counts (review Finding 7's unit
> correction applies to all three rows). §18 carries the gates as they stand at completion, after
> this story's finisher pass. The skip SET below is still correct.

### 14. Final gates — all three modules, skips BY NAME

| Gate | Measured |
|---|---|
| `folio-go` — `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -v ./...` | **688 pass · 0 fail · 4 skip** |
| `lint` — `go test -count=1 -v ./...` | **77 pass · 0 fail · 0 skip** |
| `hashmatrix` — `go test -count=1 -v ./...` | **3 pass · 0 fail · 0 skip** |

The four `folio-go` skips, named:

```
--- SKIP: TestXrefEntriesRejectsMalformedSubprocess
--- SKIP: TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability   (DW-21, this story's)
--- SKIP: TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable           (DW-21)
--- SKIP: TestTwoTablesWithPageCountFooterRenderConsistently               (DW-21)
```

Skip count moved 3 → 4, and the one added name is this story's own heavy test. Red-by-design
`TestCorpusMeetsP6ExerciseFloors` named separately above (§1), re-measured unchanged.

Also green: bare `gofmt -l` (no output across all three modules), `go vet ./...` (all three), and both
`-tags=matrix` builds (`folio-go`, `lint`).

### 15. AD-23 / AD-1 compliance

No `time`, `os`, `math/rand`, `net`, `math` transcendentals, package-level mutable state, or
output-reaching map iteration was added under `internal/`. **No binary float anywhere** — aggregates
stay exact decimal through `SumDecimals`/`AvgDecimals`. `lint`'s `TestFloatTypedProductionScan` and
`TestNoFloat64UnderModule` are green; the latter **caught two real float uses in my own first-draft
test fixtures** (a `float64` JSON struct and a `Sscanf("%f")` position parser) and both were rewritten
to integer arithmetic — recorded because the guard doing its job is the evidence, not the absence of a
finding. `os.Getenv` appears only in the heavy test's gate (test code, DW-21's shipped pattern).

### 16. What was deliberately NOT built, and what is deferred

- **Story 4.6's over-tall single row** — untouched; behaves exactly as at HEAD, fenced and red-proofed
  (§5). The D-4.5.1 discriminator goes to 4.6's creator **as a test to apply**, not a precedent to copy.
- **Story 4.8's alternating shading** — not built. Fact-checked (§8) so 4.8 is not handed a
  pre-satisfied property.
- **No `footerStyle` / `footerHeight`** — no schema field invented (R8). Footer height is derived like a
  data row's; style cascades from the table's own `style`; alignment inherits `columns[].align`
  (DECISION-4, already answered in the story).
- **`internal/layout/paginate.go` unmodified** — the per-element displacement channel from Story 4.4 was
  used, not rebuilt and not folded into the page-wide `Shift`.

### 17. Files changed by the DEVELOPER pass (the finisher's are in §18.9)

```
folio-go/table_footer.go                      (new)  — the orphan-tie mechanism
folio-go/table_footer_test.go                 (new)  — AC1-AC10 + fence + heavy test
folio-go/table_render.go                             — footer row emission, aggregate expression synthesis
folio-go/render.go                                   — isFooterLine, footer displacement, both Paginate call sites
folio-go/page_number.go                              — (unchanged in the end; PHASE A already carried the group)
folio-go/diagnostic.go                               — DiagCodeTableFooterOrphanSuppressed
folio-go/internal/diag/diag.go                       — CodeTableFooterOrphanSuppressed (minted, D-4.5.1)
folio-go/internal/diag/diag_test.go                  — registry pin for the new code
folio-go/internal/template/parse_bands.go            — D-000.67 part 2 sweep: UNRESOLVED coded
folio-go/internal/template/errors.go                 — doc comments corrected (two sites -> three)
folio-go/render_error.go                             — doc comment corrected
folio-go/table_header_repeat_test.go                 — DW-21 recorded command extended
_bmad-output/implementation-artifacts/deferred-work.md    — DW-7 retired by replacement; DW-21 amended
_bmad-output/implementation-artifacts/sprint-status.yaml  — 4-5 -> review
```

---

### 18. FINISHER PASS — what changed, and every gate re-measured

Everything in §18 was run by the finisher after the review, at the tree this story commits.

#### 18.1 Gates at completion — the SET where a set is printable, a units-labelled scalar where it is not

| Gate | Command (from that module's directory) | top-level `^--- PASS` | incl. subtests `--- PASS` | FAIL |
|---|---|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -v ./...` | **699** | **1049** | **0** |
| `lint` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **77** | **115** | **0** |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **3** | **3** | **0** |

Baseline at the start of the finisher pass, re-measured under the identical command: **688 top-level ·
1038 incl. subtests · 0 fail** for `folio-go` (matching the review exactly), **77 / 115** for `lint`,
**3 / 3** for `hashmatrix`. `folio-go` grew by **eleven top-level tests**, all added by this pass.

**`folio-go`'s skips, as a SET rather than a count** (D-000.83, sharpened: a set cannot be ambiguous
about its units, and a subtest restructuring cannot move it without moving names) — the set is
unchanged from the review's:

```
TestXrefEntriesRejectsMalformedSubprocess                             (pre-existing, unrelated)
TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability        (DW-21, FOLIO_HEAVY=1, this story's)
TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable                (DW-21, FOLIO_HEAVY=1, Story 4.4's)
TestTwoTablesWithPageCountFooterRenderConsistently                    (DW-21, FOLIO_HEAVY=1, Story 4.4's)
```

`lint` and `hashmatrix` skip **nothing** — the empty set, stated rather than reported as "0".

**The red-by-design test, named separately and never folded into any arithmetic above:**
`TestCorpusMeetsP6ExerciseFloors` is `-skip`-excluded and therefore prints no line at all. Re-measured
by the finisher and unchanged from baseline — `P6g (opaque names) floor not met: got 7, need >=20`. It
is red by design (D-000.17 / D-2.1.14 / D-000.57 / D-000.74) and is never to be "fixed".

**Heavy test, both directions, re-confirmed after touching the footer path:**

```
unset          -> --- SKIP: TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability
FOLIO_HEAVY=1  -> --- PASS: TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability (0.08s)
```

Also green: bare `gofmt -l` across all three modules (no output), `go vet ./...` (all three), and both
`-tags=matrix` builds (`folio-go`, `lint`). `git status --short -- folio-go/testdata fixtures
hashmatrix` is **empty**: no golden and no fixture moved.

#### 18.2 The count instability, recorded once so the numbers above are read correctly

A red-proof's pass/fail COUNT tracks the file's test count, not the mutation. This pass added eleven
tests, so no figure in §9 is comparable to a figure in §18.3 — the reddened test **names** are.
One concrete instance, kept because it looks like an error and is not: `PARTA_no_footer_chrome` below
reports **698 pass · 1 fail**, which is 699 — the same 699 as the green baseline, with one test moved
from the pass column to the fail column.

#### 18.3 Every deletion this story now depends on, re-run at the FINAL baseline

Whole suite, `-count=1`, never behind `-run`; tree restored between runs and **verified by SHA-256**
against a pristine backup each time. Command for all rows:
`env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -v ./...`.

| # | Mutation | Aimed at | Top-level PASS·FAIL | Reddened, BY NAME |
|---|---|---|---|---|
| **M1** | the underived-`footerFormat` pattern back to the fixed lossy `"0"` that shipped | Blocker 1 | 698 · 1 | `TestFooterWithNoFooterFormatRendersUnformattedNotRounded` |
| **M1b** | the underived pattern to a fixed maximum-precision `"0.0000"` | Blocker 1's *other* wrong default | 692 · 7 | incl. `TestFooterWithNoFooterFormatRendersUnformattedNotRounded`, `TestFooterEmptyCollectionCountZeroComesFromTheRealEvaluation`, `TestFooterAggregatesRenderPerColumnConfiguration` |
| **M2** | the review's own: `sum` route replaced by the literal `"13,500.00"` | Blocker 2 / AC4 / DW-7 | 695 · 4 | `TestFooterCellExpressionNamesTheSharedAggregateFunctions`, `TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression` (+2 collateral) |
| **M3a** | the whole DECISION-2(b) carve-out deleted (`return …, err2`) | Blocker 3, **layout** observable | 696 · 3 | `TestUnsatisfiableFooterTieStillRendersWithTheFooterPlacedAlone`, `TestUnsatisfiableFooterTieRecordsTheOrphanSuppressedWarning`, `TestUnsatisfiableTieForOneTableDoesNotAbandonAnothersTie` |
| **M3b** | suppression kept, **only** the `Diagnostic` construction deleted | Blocker 3, **diagnostic** observable | 697 · 2 | `TestUnsatisfiableFooterTieRecordsTheOrphanSuppressedWarning`, `TestUnsatisfiableTieForOneTableDoesNotAbandonAnothersTie` — and `…StillRendersWithTheFooterPlacedAlone` **stayed GREEN**, which is the discrimination |
| **M4a** | empty-collection `sum` hardcoded `"0.00"` | Blocker 4, **sum** observable | 698 · 1 | `TestFooterEmptyCollectionSumZeroComesFromTheRealEvaluation` |
| **M4b** | empty-collection `count` hardcoded `"0"` | Blocker 4, **count** observable | 698 · 1 | `TestFooterEmptyCollectionCountZeroComesFromTheRealEvaluation` |
| **M4c** | empty-collection `avg` hardcoded `""` | Blocker 4, **avg** observable | 698 · 1 | `TestFooterEmptyCollectionAvgWarningComesFromTheRealEvaluation` |
| **M5** | AC3's own PRECISION mutation: the footer aggregates only its own page (renders `"1,900.00"`) | Major 5 / AC3 | 688 · 11 | incl. `TestFooterAggregateCoversWholeCollectionNotJustItsOwnPage`, reddening at **both** of its own assertions — *"the footer's own row does not carry the whole-collection sum"* and *"the footer's own row carries `1,900.00`"* |
| **M6** | PHASE A stops seeing the footer, PHASE B still does | Major 6 / AC7 | 698 · 1 | `TestBothPaginationPassesAgreeOnFooterPartition`, reddening **on the partition first** — *"the two passes disagree on group {e1 Index:8}: PHASE A=0 PHASE B=1"* and *"group {e1 Index:-1} is carried by PHASE B and by PHASE A not at all"*, with the page-count line last |
| **M7** | merge reverted to all-or-nothing (one table's failure warns for every table) | Minor 9 | 698 · 1 | `TestUnsatisfiableTieForOneTableDoesNotAbandonAnothersTie` |
| **M8** | the merge rewrites a NON-footer group key as well | Minor 10 | 696 · 3 | `TestFooterMergeRewritesOnlyFooterGroupKeys` (+2) |

**D-000.80 part (a), the tests of the tests, run FIRST in each pair:**

| # | Part (a) mutation | Must STILL PASS | Result |
|---|---|---|---|
| **A1** | remove whatever renders a DATA row's cell text | AC1's `TestFooterAggregatesRenderPerColumnConfiguration` **and** Blocker 1's new `TestFooterWithNoFooterFormatRendersUnformattedNotRounded` | 676 · 23 — **both STILL PASSED** ✅ |
| **A2** | remove whatever draws the footer row's own cell CHROME | AC5's `TestFooterOrphanTieMovesWithImmediatelyPrecedingRow` and `TestFooterOrphanTiePresencePrecondition` | 698 · 1 — **both STILL PASSED** ✅ (the one failure is AC6's test, which reads chrome-rect extents by value on purpose) |

Two tests are named as **deliberately failing under A1**, so their appearance in that run's failure
list is not a defect: `TestFooterUnderivedFixtureDataCellsCannotCollideWithFooterValues` and
`TestFooterAggregateCoversWholeCollectionNotJustItsOwnPage` both assert that DATA cells are drawn —
they are anchoring/precondition instruments, not the AC's witness, and part (a) is not claimed for
either. That split is why the Blocker 1 fixture's anchoring lives in its own test rather than inside
the witness.

#### 18.4 One observable, one deletion — the counting this pass added

The engineering lead's diagnosis of Blockers 3 and 4 was that they share a root: *an AC that produces
two observables passes the deletion screen when deleting it reddens on EITHER; the screen's result is
a boolean standing in for a set.* Recorded per AC touched:

| AC / ruling | Observables | Deletion per observable |
|---|---|---|
| DECISION-2(b) / D-4.5.1 (Blocker 3) | **2** — the layout outcome (render succeeds, footer placed alone) and the `TABLE_FOOTER_ORPHAN_SUPPRESSED` Warning | M3a and **M3b**; M3b is the one that matters, because it leaves the layout half green |
| AC9 / D-4.5.3 (Blocker 4) | **3** — `sum`'s zero, `count`'s zero, `avg`'s Warning | M4a, M4b, M4c — each reddens **exactly one** test, and only its own |
| AC4 / DW-7 (Blocker 2) | **2** — the value tracks the data (behavioural) and the route names the shared aggregate (structural) | M2 reddens both, and the two tests are separately named so a future weakening of either is visible |
| AC1/AC2's underived arm (Blocker 1) | **1** observable, **2** wrong defaults | M1 (lossy) and M1b (fixed-max-precision) — the fixture discriminates against both |

#### 18.5 Blocker 1 — the FIXTURE gap was closed, not just the code path

Stated explicitly because shipping only the code fix would let the same bug return the next time
anyone touches the pattern default: **a default is only tested by an input that would render
differently under a wrong default.** Before this pass, the only column in the whole corpus reaching
D-1.4.1's "absent and underived" arm was `footerFixtureDocCount`'s `count` column — integral, so the
lossy default and a correct one produced the identical string. The corpus now carries
`footerFixtureDocUnderivedFormat` + `footerFixtureDataNonIntegral`: three columns that ALL reach that
arm (each declares a footer over a numeric source while its own cell bind is shape 1 over a STRING
field, so no format is explicit and none is derivable), over data whose exact total is **30.85**.
Under the shipped default that rendered **31**; under a fixed maximum-precision default it renders
**30.8500**; both are pinned as forbidden spellings alongside the correct one.

#### 18.6 Blocker 1's mechanism, and the one residual, stated

"Unformatted" is not a pattern the grammar can SPELL — `validateNumberPattern`'s fraction part is
drawn from `'0'` alone, never `'#'` — so it is COMPUTED from the value instead:
`expr.UnformattedPattern` derives the pattern from the aggregate's own `Decimal` scale (no grouping,
no zero-padding), and `footerCellExprText` learns that scale by evaluating the aggregate once through
`bind.EvaluateValue` — a new, narrow bare-expression entry point beside the existing
`EvaluateCondition`, sharing the SAME `exprResolver` seam, so AC4/DW-7 is untouched (no second
arithmetic, only a second traversal of the one that already exists). `EvaluateCondition`'s D-3.5.1
literal tripwire is deliberately NOT inherited: that tripwire is a property of a condition slot, and
inheriting it would make this a second condition slot — the exact thing it exists to prevent.

**The one residual, named rather than hidden:** the derived pattern is clamped to the grammar's own
`maxPatternFractionDigits` (`== avgExtraScale`), which is the widest pattern
`validateNumberPattern` accepts and therefore the product's declared display ceiling — it binds an
explicit author-written `footerFormat` exactly as hard. `avg` reaches that clamp **routinely** (its
scale is maximum operand scale + `avgExtraScale` by construction), so this is the ordinary case for
`avg`, not an exotic one; the fixture pins `15.4250` by value so a change to the ceiling is visible
here rather than silent. Widening the ceiling is a change to the pattern grammar's own bound (AC17)
and belongs to whoever owns that bound. Recorded as a follow-up below, not fixed here.

#### 18.7 D-4.5.4's fence, re-aimed (review Finding 10)

The review's suggested mirror — *a document reaching the second `Paginate` where the overflow is
genuinely non-footer in origin* — is **not buildable at this commit, and the reason is the reviewer's
own verdict on Claim 2**: `layout.Paginate` raises `OverflowError` only when an item or group is
taller than a **full** content window (a window-position-independent test), pass 1 already succeeded,
and pass 2 differs from pass 1 **only** in footer group keys — so any pass-2 `OverflowError` can only
come from a merged footer group. Rather than assert that by construction a second time, the property
it actually rests on is now executable: `applyFooterMerge` was extracted from
`paginateWithFooterOrphanFix`, and `TestFooterMergeRewritesOnlyFooterGroupKeys` asserts that every key
the merge rewrites was a footer key, that nothing else is touched, and that the input slice is not
mutated. **M8 is its red-proof.** `TestOverTallSingleRowStillOverflows` keeps its own doc-comment
concession (it fences pass 1's passthrough, true by construction) — it is no longer the only thing
standing between the carve-out and an unnoticed widening.

#### 18.8 What the finisher did NOT change

- **`internal/layout/paginate.go` still unmodified** by this story.
- **No golden or fixture moved**; no new gate obligation; no new `//go:build matrix` file.
- **No diagnostic code minted** beyond D-4.5.1's `TABLE_FOOTER_ORPHAN_SUPPRESSED`, which now has a
  behavioural witness for the first time.
- **`MixedItemError` untouched**; no `footerStyle`/`footerHeight` invented; alignment still inherits
  `columns[].align`.
- **AC2, AC5, AC6, AC8 and AC10 were left as the developer built them** — the review found all five
  satisfied, and A2 above re-confirms AC5's part (a) at the new baseline.

#### 18.9 Files changed by the finisher pass

```
folio-go/internal/expr/numberpattern.go     UnformattedPattern (new)
folio-go/internal/bind/condition.go         parseAndCheck extracted; EvaluateValue (new)
folio-go/table_render.go                    footerCellExprText: derive the underived pattern from the value's own scale
folio-go/table_footer.go                    per-table merge retry (Finding 9); applyFooterMerge extracted (Finding 10)
folio-go/internal/template/parse_bands.go   closedFooterKinds check: the real ground for leaving it uncoded (Finding 8)
folio-go/table_footer_test.go               11 new/rewritten tests (see Finding Resolutions)
_bmad-output/implementation-artifacts/deferred-work.md    DW-7 corrected and re-discharged
```

---

## Finding Resolutions

**Triage: 11 findings — 11 FIX, 0 DISMISS, 0 DEFER.** One follow-up is raised (not a deferral of a
finding; a bound this story measured and does not own — §18.6).

Severity cross-check against the review's own header (4 Blockers · 2 Majors · 4 Minors · 1 Nit):
counted directly from the `Severity:` lines rather than from the summary — **4 · 2 · 4 · 1**, agreeing.

---

### Finding 1 — a `sum`/`avg` footer with no `footerFormat` silently rounds — **FIX** (Blocker)

**Reproduced before fixing**, by probe: a `sum` footer over `10.55 + 20.30` rendered `"31"` against a
true total of `30.85`. D-1.4.1 settles this ("absent and underived, the footer renders unformatted"),
so no ruling was sought.

**Changed:** `internal/expr/numberpattern.go` gains `UnformattedPattern(Decimal) string`;
`internal/bind/condition.go` gains `EvaluateValue` (sharing a new `parseAndCheck` with
`EvaluateCondition`); `table_render.go`'s `footerCellExprText` takes `scope`/`fc` and, on the
underived arm only, evaluates the aggregate once to learn its scale and derives the pattern from it.
Mechanism and the clamp residual: §18.6.

**Fixture gap closed, which is the half that stops the bug returning:** §18.5.

**Proved by:** M1 (the shipped lossy default → reddens), M1b (a fixed maximum-precision default →
reddens), A1 part (a) (data-row cell text removed → the witness STILL PASSES). Plus
`TestUnformattedPatternTracksTheValuesOwnScale` as the unit-level witness that the pattern is a
function of the value and never a literal.

---

### Finding 2 — AC4/DW-7 has no footer-side witness — **FIX** (Blocker)

**Reproduced:** the review's mutation (`return "13,500.00"`) left the suite green.

**Changed:** `TestFooterCellExpressionNamesTheSharedAggregateFunctions` (new, structural, by AST —
the footer-side analogue of `routing_arch_test.go`'s two routing assertions) and
`TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression` (rewritten to run over TWO
datasets, with an asserted anti-constant precondition that their aggregates differ).

**DW-7 ends this story DISCHARGED — on new evidence, and the old evidence is withdrawn in the
ledger.** `deferred-work.md`'s DW-7 entry carries an appended finisher block naming the [[D-4.2.4]]
defect explicitly: the retirement had been recorded against `TestDecimalReducerInventoryIsExactlySumAndAvg`,
a Story 3.1a module-wide guard that was not measuring the footer.

**Proved by:** M2 — the review's own mutation now reddens both new witnesses.

---

### Finding 3 — `TABLE_FOOTER_ORPHAN_SUPPRESSED` has zero behavioural witness — **FIX** (Blocker)

**Changed:** a real fixture (`footerFixtureDocUnsatisfiableTie`, geometry derived by throwaway probe,
not by hand) in which the footer row and its immediately preceding data row EACH fit the content
window alone but not together, driven through the public `Render()`; and **two** tests, one per
observable — `TestUnsatisfiableFooterTieStillRendersWithTheFooterPlacedAlone` (render succeeds, row 0
stays on page 0, the footer's values land alone on page 1) and
`TestUnsatisfiableFooterTieRecordsTheOrphanSuppressedWarning` (`Result.Diagnostics` carries exactly
one `SeverityWarning` `DiagCodeTableFooterOrphanSuppressed` naming `e1`).

**Proved by:** M3a and **M3b**. M3b is the load-bearing one: it deletes only the `Diagnostic`
construction, and the layout test stays GREEN while the diagnostic test reddens. §18.4 records why
two were needed.

---

### Finding 4 — AC9's zero is not proven to be by decision — **FIX** (Blocker)

**Changed:** the single `TestFooterOverEmptyCollectionRendersAndTerminates` was split so each reducer
carries its own provenance witness —
`TestFooterEmptyCollectionSumZeroComesFromTheRealEvaluation`,
`…CountZeroComesFromTheRealEvaluation`, `…AvgWarningComesFromTheRealEvaluation`. The sum and count
witnesses assert **structurally** (by AST) that the expression the footer hands to `bind.Resolve` is a
`formatNumber()` over the shared `sum`/`count` on the right path — which a hardcoded zero cannot
produce — **and** behaviourally that the drawn text equals what resolving that same expression
independently yields, which catches a hardcode placed downstream of the synthesis. The avg witness
keeps the Caveat `DataPath` guardrail. The original test retains AC9's other clauses (bytes returned,
one page, the two zeros deliberately unharmonised per D-3.1a.2).

**Proved by:** M4a, M4b, M4c — three mutations, three tests, each reddening **exactly its own**.

---

### Finding 5 — AC3's precondition is an empty `if`; its assertion is self-implied — **FIX** (Major)

**Changed:** `TestFooterAggregateCoversWholeCollectionNotJustItsOwnPage` rewritten against the
composed page model and asserted POSITIONALLY. Row 8's Y comes from its own `R8` marker; the footer's
Y is the next distinct run Y below it — **neither is located by the value under test**, which is what
kept the old version circular. It then asserts: `1,900.00` present at row 8's Y (the presence
precondition that stops the next line proving nothing), `13,500.00` present at the footer's Y, and
`1,900.00` **absent** at the footer's Y.

Per the lead: AC3's defect is explicitly **not** the two-observables/one-boolean shape — prose cannot
be deleted — so no screen was invented for it; it is simply fixed.

**Proved by:** M5, the AC's own named precision mutation, which reddens at **both** of the test's own
assertions.

---

### Finding 6 — AC7 never compares the data-row partition — **FIX** (Major)

**Changed:** `TestBothPaginationPassesAgreeOnFooterPartition` now builds a full group→page map for
each pass (`groupPagePartition`, new) and compares them **in both directions**, so a key present in
one pass and missing from the other fails rather than passing unnoticed. Presence preconditions are
taken against PHASE B (the pass that decides what exists) rather than PHASE A — asserting them
against A would abort the comparison instead of reporting it, since A is where the defect lives. The
page-count check moved to the END and from `Fatalf` to `Errorf`, so the failure lands on the
partition rather than on a count that happened to move with it.

**Proved by:** M6 — reddens on the partition first, page count last (transcript in §18.3).

---

### Finding 7 — §1's "correction" is a unit mismatch — **FIX** (Minor)

**Changed:** the false correction is WITHDRAWN in §1, with both units tabulated; §9 and §14 are
labelled top-level; §14 is marked superseded by §18.1, which carries both units for every gate. Per
the sharpened D-000.83 guidance, skips are now reported as a SET (and `lint`/`hashmatrix` as the
EMPTY set) rather than as a count, and the red-by-design test is named separately.

---

### Finding 8 — §12 mischaracterises 1 of the 4 uncoded footer sites — **FIX** (Minor)

**Changed:** §12 now distinguishes the three `decodeStringRaw` TYPE failures from the one
`closedFooterKinds` CLOSED-SET failure and states the actual ground for leaving the latter uncoded
(a footer KIND is not a statement about the footer's numeric SOURCE), together with the neighbouring
asymmetry the sweep should have surfaced. The ground is also stated **at the call site** in
`parse_bands.go` so the next D-000.67 sweep does not re-litigate it. Sites are cited by declaration
rather than by line number.

---

### Finding 9 — one unfixable table reverts and warns for every merged table — **FIX** (Minor)

**Changed:** `paginateWithFooterOrphanFix` no longer discards the whole merged plan when any one
merged group overflows. It retries each candidate on top of the ones already accepted, so a table
whose tie fits keeps it, and emits a Warning naming only the tables that genuinely could not fit. The
fast path is unchanged (one extra `Paginate` in the common orphan case); the walk is reached only on
the rare unsatisfiable path, and `toMerge` is a slice walked in first-appearance order, so its outcome
is deterministic (R5).

**Changed (test):** `TestUnsatisfiableTieForOneTableDoesNotAbandonAnothersTie`, at the layout level,
with two tables of which exactly one has an unsatisfiable tie. Its geometry was derived by throwaway
probe after a first hand-derived attempt was **wrong** — extents are page-absolute, so the first
window runs from contentTop, and the sweep slides its window *within* a page. Its presence
precondition now asserts, against the un-fixed `layout.Paginate`, that **both** footers really are
orphaned, so "only one Warning" is a statement about scoping rather than an artifact of there being
only one candidate.

**Proved by:** M7.

---

### Finding 10 — the fence's subject is pass 1's passthrough, not the carve-out — **FIX** (Minor)

The suggested mirror is not buildable at this commit, and the reason is the review's own Claim 2
verdict. Re-aimed instead at the property the carve-out's scope actually rests on. Full reasoning and
the new instrument: §18.7.

**Proved by:** M8.

---

### Finding 11 — the heavy test's name promises the orphan tie; its body never asserts it — **FIX** (Nit)

**Changed:** `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability` now asserts that the
last data row (`R499`) is on the final page alongside the footer — by the moved row's IDENTITY, the
way AC5's own test does, not as "some row is there". Confirmed in both directions (§18.1).

---

### Follow-up raised (not a finding)

**The pattern grammar's fraction-digit ceiling bounds what "unformatted" can mean.**
`maxPatternFractionDigits == avgExtraScale`, so a footer value whose scale exceeds it renders rounded
half-to-even at that ceiling. `avg` reaches it by construction. This is a property of the pattern
grammar (AC17's bound), it binds explicit author-written `footerFormat` identically, and it is not
introduced by this story — the ceiling is measured, pinned by value in the new fixture (`15.4250`),
and documented on `UnformattedPattern` itself. Widening it belongs to whoever owns that bound.

---

## Review Findings

### Review Summary

- **Reviewed by:** bmad-code-reviewer (fresh, adversarial; did not write this code)
- **Date:** 2026-08-27
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 4 · **Majors:** 2 · **Minors:** 4 · **Nits:** 1

**Aim, per the brief:** budget reallocated to **Class B** (D-000.79) — *name a WRONG state that produces the same summary statistic as the RIGHT one* — with every red-proof checked for **D-000.80 part (a)**.

**Gate re-measured independently by this review (units stated, D-000.81):**

| Gate | Top-level `^--- PASS` | Incl. subtests `--- PASS` | Fail |
|---|---|---|---|
| `folio-go` | **688** | **1038** | 0 |
| `lint` | **77** | **115** | 0 |

Command: `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors$ -v ./..."`.
`folio-go` skips, **BY NAME** (4): `TestXrefEntriesRejectsMalformedSubprocess`,
`TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`,
`TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`,
`TestTwoTablesWithPageCountFooterRenderConsistently`.
`TestCorpusMeetsP6ExerciseFloors` is **red by design**, `-skip`-excluded, named separately and never folded into arithmetic.
Bare `gofmt -l` across `folio-go`: no output. Tree verified pristine by SHA-256 before and after every mutation below.

**Verdict on the developer's own claims (verified by execution, not by reading):**

| # | Claim | Verdict |
|---|---|---|
| 1 | The fence reddens 6 tests incl. Story 4.3's | **REPRODUCED** exactly — but see Minor 10 for what it actually fences |
| 2 | Never-error carve-out keyed on footer-ness | **HOLDS.** Pass 1 returns an over-tall row's `OverflowError` before any merge; pass 2 differs from pass 1 only in footer group keys, so its `OverflowError` can only come from the merged group. **Story 4.6's subject is not pre-empted.** |
| 3 | Footer-alone terminates by mechanism | **HOLDS** — pass-1 early return, not fixture size (it reddens under a pass-1 mutation) |
| 4 | Zero **by decision** for the empty collection | **FALSE — Blocker 4** |
| 5 | AC5 part (a) now passes honestly | **HOLDS.** `footerOrphanTargetsFrom` genuinely keys on `ColumnItem.Group.Key`, which the value runs carry independently of the chrome; `pageOfGroup` falls back to `Runs`. Chrome-independent by construction. |
| 6 | AC6's witness discriminates position | **HOLDS** — asserts exact `Y` by value (30,896 vs the mutation's 20,896); presence alone would not distinguish them |
| 7 | `parse_bands.go` sweep, 4 footer sites correctly uncoded | Conclusion defensible; **stated reason wrong for 1 of 4 — Minor 8** |
| 8 | New diagnostic code fully wired | Registry wiring **complete** (declared, `allCodes`, pinned, bridged) — but its construction block is **never executed — Blocker 3** |

---

### Finding 1: A `sum`/`avg` footer with no `footerFormat` silently rounds the total — a wrong bank-statement figure

- **Severity**: **Blocker**
- **Category**: Correctness / AC Conformance
- **Location**: `folio-go/table_render.go`, `footerCellExprText` — `pattern := "0"` (the "absent and underived" default)
- **Observation**: When `footerFormat` is neither explicit nor derivable (a shape-1 `bind` on the footer column), the synthesised expression is `{{formatNumber(<agg>, "0")}}`. `"0"` parses to `fracDigits: 0` (`internal/expr/numberpattern.go`), so the total is rendered rounded to whole units. Measured by probe (a `sum` footer, `footerOf: "items.b"`, data `10.55 + 20.30`):

  ```
  PROBE page 0 texts: ["A" "R0" "R1" "31"]
  ```

  The true total is **30.85**; the document renders **31**.
- **Impact**: A money total is silently altered at the display boundary. This contradicts **D-1.4.1**, quoted verbatim in this story's own "aggregate evaluation" section — *"Absent and underived, the footer renders **unformatted** through the one display-text function in `internal/bind`"* — and it contradicts the story's own plain-terms promise: *"Money is counted exactly, never with the rounding a computer does by default."* The arithmetic is exact; the rendering is not. A ruled behaviour was replaced by a lossy default without a decision being surfaced.
- **Why no test caught it — this is the Class B**: the only column in the entire fixture set that reaches the `"0"` default is the **`count`** column (`footerFixtureDoc`'s `e2`, shape-1 bind `{{row.a}}`). `count` is integral, so a formatter with zero fraction digits produces **the same rendered string** as a correct one. The wrong state and the right state are indistinguishable on the only data class that exercises the path. `e3` carries an explicit `footerFormat` and `e4` derives one, so neither touches the default.
- **Suggested Resolution**: Take this back to the engineering lead as a decision, since D-1.4.1 already ruled it and the ruling was not followed. Either honour "renders unformatted" (a decimal-to-text path that preserves the value's own scale), or — if the pattern grammar genuinely admits no such pattern — surface that as a DECISION rather than defaulting to a lossy one. Whatever is chosen, add a fixture column that reaches the default with a **non-integral** value, so the path stops being witnessed only by `count`.
- **Related AC**: AC1, AC2 (and AC3's "matching Story 3.3" equality, which the default breaks for any non-integral total)

---

### Finding 2: AC4 / DW-7 — the footer half has no witness; the entire aggregate route can be a hardcoded constant

- **Severity**: **Blocker**
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/table_footer_test.go:399` `TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression`; Delivery Log §9 (AC4 row) and §11
- **Observation**: Red-proof run by this review (**M4**) — the footer's `sum` route replaced by a literal that never touches `SumDecimals`, `internal/expr`, or any aggregate evaluation, and which agrees with the fixture:

  ```go
  if col.Footer.Value == "sum" && !collectionEmpty {
      return "13,500.00", nil   // a rival that agrees on this story's data
  }
  ```

  Whole suite, `-count=1`, not behind `-run`: **688 pass · 0 fail — completely green.** `TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression` did **not** redden.
- **Why it cannot redden**: the test builds its "independent" expectation by calling `bind.Resolve` on a hand-written `{{formatNumber(sum(items.b), "#,##0.00")}}` and compares the resulting **string** to the rendered text. It is a value comparison on easy data (integers, `.00`, an `avg` that terminates at two decimals). Any rival that agrees on that data passes. This is exactly the defect AC4 says it exists to catch — *"a second implementation that agrees on easy data and drifts on hard data"* — and exactly what AC4 warned against: *"Do not let the footer half be weaker than the behavioural equality in AC3."*
- **Corroborated by the developer's own record**: the AC4 row in §9 reports the named mutation reddening **only** `TestDecimalReducerInventoryIsExactlySumAndAvg` — a **Story 3.1a** structural inventory guard that fires for a third reducer *anywhere in the module*, with no relation to the footer. **DW-7 was retired by replacement against a witness that belongs to another story.** DW-7's own recorded anti-rot mechanism is *"none possible for the footer half"*, which is precisely why the AC demanded a positive footer-side assertion.
- **Impact**: DW-7 is marked discharged in `deferred-work.md` while the obligation it names is unguarded. A future refactor can introduce a second summation for the footer and the gate will not notice.
- **Suggested Resolution**: Add a footer-side assertion with structural teeth in the shape Story 3.3 shipped for the `{{ }}` half (`routing_arch_test.go`'s `TestSumRoutesThroughSumDecimals` / `TestAvgRoutesThroughAvgDecimals`), **or** make the behavioural equality discriminating by anchoring it on data where an inline accumulator provably drifts (scales that force `SumDecimals`' exact-decimal path and round-half-to-even to disagree with a naive accumulation). Do not re-retire DW-7 until the mutation above reddens a Story 4.5 test.
- **Related AC**: AC4

---

### Finding 3: DECISION-2(b)'s ruled carve-out and the minted `TABLE_FOOTER_ORPHAN_SUPPRESSED` have zero behavioural witness

- **Severity**: **Blocker**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_footer.go:246-268` (the `errors.As` catch, the `Diagnostic` construction loop, `return plan, diags, nil`)
- **Observation**: Red-proof run by this review (**M1**) — the entire carve-out deleted outright (`return layout.Pagination{}, nil, err2` instead of suppress-and-record). Whole suite: **688 pass · 0 fail — completely green.** No test observes the never-error behaviour, the footer-placed-alone fallback, or the Warning.

  Corroborating grep — `DiagCodeTableFooterOrphanSuppressed` / `CodeTableFooterOrphanSuppressed` appear **only** at: `internal/diag/diag.go` (declaration, `allCodes`), `internal/diag/diag_test.go:43` (registry string pin), `diagnostic.go` (bridge), and `table_footer.go:261` (the single construction site). **No test asserts the code is ever produced**, and no test constructs a fixture in which the preceding row and the footer together exceed the content window.
- **Impact**: This is **Story 4.4's Blocker 3 shape reproduced one story later** — *"a whole diagnostic construction block deleted unnoticed"* — which **this story's own AC6 cites as its grounds**. A code was minted under D-4.5.1 against AD-14's closed registry (permanent, additive-only) for a condition that nothing demonstrates can occur. The ruled DECISION-2(b) outcome ("never error, place alone, record it, always terminate") is asserted in prose and in comments but is unmeasured. Note that `TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs` does **not** cover this: it exercises the **footer alone** exceeding the window, which returns from **pass 1** before any merge — a different branch.
- **Suggested Resolution**: Build the fixture DECISION-2(b) actually describes — a table whose **footer plus its immediately preceding row** exceed the content window while each fits alone — and assert, by value: the render **succeeds**, the footer is placed alone on its page, and `Result.Diagnostics` carries a `SeverityWarning` `DiagCodeTableFooterOrphanSuppressed` naming that element. Then re-run M1 and confirm it reddens.
- **Related AC**: AC5 (its unsatisfiable arm), DECISION-2(b), D-4.5.1

---

### Finding 4: AC9 / D-4.5.3 — the empty-collection zero is NOT proven to be zero by decision

- **Severity**: **Blocker**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_footer_test.go:752` `TestFooterOverEmptyCollectionRendersAndTerminates`; Delivery Log §7
- **Observation**: Red-proof run by this review (**M2**) — both non-`avg` empty-collection aggregates replaced by hardcoded literals that evaluate nothing at all:

  ```go
  if col.Footer.Value == "sum"   && collectionEmpty { return "0.00", nil }
  if col.Footer.Value == "count" && collectionEmpty { return "0",    nil }
  ```

  Whole suite: **688 pass · 0 fail — completely green.**
- **Why the stated guardrail does not discriminate — this is the Class B**: §7 claims the anti-fallthrough guard is the empty-average Warning's `DataPath == "items.c"`, *"which only an implementation that actually evaluated the expression can know."* That is true **of the `avg` column** — and only of it. The `sum` and `count` zeros are asserted purely as rendered strings (`pageContains(pages, 0, "0")` / `"0.00"`). A hardcoded fallthrough for `sum`/`count` leaves `avg` on the real path, so the Warning and its `DataPath` still appear exactly as before. **The guard certifies a different column than the ones whose zeros it is claimed to certify.** D-4.5.3 asked for *"a test that would fail if it were zero by fallthrough"*; M2 is that fallthrough and the test passes.
- **Impact**: This is the ruling's own named failure mode — *"a reducer returning its zero value because it never entered its loop produces an identical PDF; this is a wrong bank-statement total"* — shipped unguarded, and recorded in the Delivery Log as satisfied.
- **Suggested Resolution**: Give `sum` and `count` their own provenance witness, not a value witness. Options: assert that the empty-collection `sum` renders at **`SumDecimals`' own scale** and `count` at **its** scale, distinguishably (D-3.1a.2's deliberately-unharmonised two zeros already give you two different strings — pin both to the scale the kernel produces, not to literals the test chose); or route a caveat/diagnostic that only a real evaluation can emit. Re-run M2 and require it to redden.
- **Related AC**: AC9, D-4.5.3

---

### Finding 5: AC3's mandated non-vacuity precondition is an empty `if` block, and its remaining assertion is implied by its own setup

- **Severity**: **Major**
- **Category**: Tests
- **Location**: `folio-go/table_footer_test.go:372-397` `TestFooterAggregateCoversWholeCollectionNotJustItsOwnPage`
- **Observation**: Two defects in one test.

  (a) The AC's explicitly mandated precondition — *"the rows on the footer's own page do not aggregate to the same value as the whole collection — **assert that inequality** from literals the test owns"* — is written as an `if` statement whose **body contains only a comment**:

  ```go
  if pageContains(pages, footerPage, "1,900.00") {
      // fine: row 8's own cell also renders here, that's expected —
      // the precondition is that the WHOLE-COLLECTION answer ...
  }
  ```

  It executes nothing and can never fail. The AC states this mechanism *"is what keeps the AC from being vacuous, and it is why no part (a) is claimed."*

  (b) The one surviving assertion is a tautology — `footerPage` is **defined** as the page containing `"13,500.00"`, and is then asserted to contain `"13,500.00"`:

  ```go
  footerPage, ok := anyPageContains(pages, "13,500.00")
  ...
  if !pageContains(pages, footerPage, "13,500.00") { t.Fatal(...) }
  ```

  Nothing can make this redden that has not already tripped the `!ok` fatal three lines above.
- **Impact**: AC3 reduces to *"there are ≥ 2 pages and the string `13,500.00` appears somewhere"* — which `TestFooterAggregatesRenderPerColumnConfiguration` (AC1) already asserts. The AC's own stated protection against a per-page implementation coinciding with a whole-collection one is not in the tree. The §9 red-proofs for AC3 redden on the **presence** check (a per-page sum renders `1,900.00`, so `13,500.00` is absent) — i.e. on AC1's assertion wearing AC3's name, not on the AC's own mechanism. It happens to hold for this fixture, which is luck, not measurement.
- **Suggested Resolution**: Replace the empty `if` with a real assertion that the footer-page-only aggregate value (`"1,900.00"`, count `"1"`, avg `"1.00"`) is **present as a data cell but absent as the footer's value**, and drop or rewrite the tautological check. Per D-000.79, a mechanism a ruling names must be built, not narrated.
- **Related AC**: AC3

---

### Finding 6: AC7's test implements half its own AC text — the data-row partition is never compared between passes

- **Severity**: **Major**
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/table_footer_test.go:643` `TestBothPaginationPassesAgreeOnFooterPartition`
- **Observation**: AC7 requires *"the per-page partition — **data rows** and the footer's page — is identical between the two."* The test asserts only three things: `len(phaseAPlan.Pages) == len(phaseBPlan.Pages)`, that the count is 2, and that **the footer's** group lands on the same page in both passes. It never compares where rows 0..8 landed in PHASE A against PHASE B.
- **Impact**: The defect D-4.3.2 names is *"two passes silently disagreeing."* A disagreement confined to the data rows — the larger half of the partition, and the half whose builders *"append in different orders"* — passes this guard untouched. The footer's page agreeing is a necessary but not sufficient condition, and the AC said so explicitly.
- **Note**: this is a **weaker** instrument than the existing `TestBothPaginationPassesAgreeOnRowPartition` the AC named as *"the existing instrument and the right place to extend"* — that test was not extended; a narrower new one was added beside it.
- **Suggested Resolution**: Compare the full per-page group→page map between the two plans (all `Index >= 0` groups plus `IsHeader` plus the footer), not just the footer's entry — or extend `TestBothPaginationPassesAgreeOnRowPartition` as the AC directed.
- **Related AC**: AC7

---

### Finding 7: Delivery Log §1's "correction to the story's baseline" is a unit mismatch, not a discrepancy — and contradicts the convention the same paragraph declares

- **Severity**: **Minor**
- **Category**: Maintainability / Process
- **Location**: Delivery Log §1 (this file), lines ~710-731
- **Observation**: §1 states the convention as *"`--- PASS` / `--- FAIL` / `--- SKIP` lines **anywhere** in `-v` output, subtests included"*, then reports `lint` at **77 pass** and flags the story's **115** as *"a discrepancy between the story's recorded figure and the tree"*. Measured by this review under the exact command:

  | Gate | `^--- PASS` (top-level) | `--- PASS` (incl. subtests) |
  |---|---|---|
  | `lint` | **77** | **115** |
  | `folio-go` | **688** | **1038** |

  **Both figures are correct and measure different units.** The story counted subtests-inclusive (its own stated convention); the developer counted **top-level only** while declaring the subtests-inclusive convention. Nothing moved in the tree.
- **Impact**: A false correction is now recorded against the story's baseline table, and the same unit error propagates through every count in §9's red-proof table (which is why its figures are internally consistent but not comparable to the story's). This is the failure D-000.68 names: a count without its unit is a lossy measurement.
- **Suggested Resolution**: Withdraw the "correction", restate every figure in §1/§9/§14 with the unit it was counted in, and pick one convention and hold it.
- **Related AC**: none (process / D-000.81)

---

### Finding 8: §12's sweep mischaracterises 1 of the 4 footer-related sites; the neighbouring asymmetry is real

- **Severity**: **Minor**
- **Category**: Convention / Maintainability
- **Location**: `folio-go/internal/template/parse_bands.go:414`; Delivery Log §12
- **Observation**: §12 states the 4 uncoded footer-related sites *"are malformed-template/**type** failures, not statements about a footer's numeric SOURCE."* Verified: lines **411**, **433**, **459** are indeed `"must be a string: "` type failures. Line **414** is not:

  ```go
  if !closedFooterKinds[s] {
      return Column{}, newLoadError("footer", string(id), s, "not one of the closed set sum, count, avg")
  }
  ```

  That is a **closed-set / value-domain violation**, not a type failure — a well-typed string carrying a footer *kind* outside `{sum, count, avg}`. The asymmetry the sweep should have surfaced: its immediate neighbour at line **425** (`footerOf`/`footerFormat` present with no `footer`) is a pure field-presence check and **is** coded `TABLE_FOOTER_SOURCE_FORBIDDEN`.
- **Impact**: The **conclusion** (leave 414 uncoded) is defensible — `TABLE_FOOTER_SOURCE_*` names a failure of the footer's numeric *source*, and a bad footer *kind* is a different statement — but the **stated reason is wrong**, so the judgement recorded as "auditable" does not in fact audit. D-000.67 part 2 asks for the sites-examined figure *and* a reason that survives inspection.
- **Suggested Resolution**: Correct §12 to distinguish the three type failures from the one closed-set failure, and state the actual ground for leaving 414 uncoded (kind ≠ source), so the next sweep does not re-litigate it.
- **Related AC**: none (D-000.67 part 2)

---

### Finding 9: A single unfixable table reverts and warns for every merged table in the document

- **Severity**: **Minor**
- **Category**: Correctness
- **Location**: `folio-go/table_footer.go`, `paginateWithFooterOrphanFix` — the `plan2`/`err2` handling and the `for _, c := range toMerge` diagnostic loop
- **Observation**: `toMerge` may hold candidates for several tables. All are merged into one `merged` slice and one `layout.Paginate` call. If **any** one merged group overflows, `err2` is non-nil, the whole of `plan2` is discarded, and the function returns pass 1's `plan` — so tables whose orphan tie **would** have succeeded are silently left orphaned. The diagnostic loop then emits a `TABLE_FOOTER_ORPHAN_SUPPRESSED` Warning for **every** candidate in `toMerge`, including those that were never the cause.
- **Impact**: On a multi-table document, one pathological table degrades the orphan rule for all of them, and the Warnings misattribute the cause. Not reachable by any current test (see Finding 3), so it is latent rather than live.
- **Suggested Resolution**: Merge and re-paginate per table, or on overflow retry with the offending table's candidate removed, so the suppression and its Warning are scoped to the table that actually could not fit.
- **Related AC**: AC5 (unsatisfiable arm), DECISION-2(b)

---

### Finding 10: The fence's subject is pass 1's passthrough, not the carve-out it is named for

- **Severity**: **Minor**
- **Category**: Tests
- **Location**: `folio-go/table_footer_test.go:982` `TestOverTallSingleRowStillOverflows`
- **Observation**: Claim 1 **reproduced** — widening the catch to any `OverflowError` reddens 6 top-level tests, including Story 4.3's `TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender`:

  ```
  --- FAIL: TestFourErrorModesCarrySeverityErrorDiagnostics
  --- FAIL: TestMessageRewriteDoesNotAffectCodeRecovery
  --- FAIL: TestRenderReportsAnItemThatFitsOnNoPage
  --- FAIL: TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs
  --- FAIL: TestOverTallSingleRowStillOverflows
  --- FAIL: TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender
  ```

  So the fence is **not trivially green** and the story's claim stands. However, the test calls `paginateWithFooterOrphanFix(g, items, **nil**)` with an item that overflows in **pass 1** — it returns before `len(targets) == 0` is even consulted and never reaches the second `Paginate` or the carve-out. Its own doc comment concedes the point: *"this assertion is TRUE BY CONSTRUCTION rather than by a runtime bypass condition."* The mutation that reddens it lands on **pass 1's error return**, a different site from the carve-out at pass 2. `TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs` reddens under the same pass-1 mutation and is, for fencing purposes, the same assertion with a different fixture.
- **Impact**: Pass 1's passthrough is well fenced. The carve-out's **scope** — the property D-4.5.4 exists to protect — is fenced only by construction, and Finding 3 shows the carve-out has no witness of any kind. The two findings compound: nothing in the suite executes that branch, and nothing constrains what it would do if widened.
- **Suggested Resolution**: Once Finding 3's fixture exists, add its mirror — a document reaching the second `Paginate` where the overflow is genuinely non-footer in origin — and assert it still errors.
- **Related AC**: AC5, D-4.5.4

---

### Finding 11: The heavy test's name promises the orphan tie; its body never asserts it

- **Severity**: **Nit**
- **Category**: Maintainability
- **Location**: `folio-go/table_footer_test.go:850` `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`
- **Observation**: The body is **real** (not an empty body under an unconditional skip — Story 4.4's Blocker 1 is **not** repeated here; the gate is `os.Getenv(heavyTestGateEnvVar) != "1"`, an env var, correctly appended to DW-21). It asserts byte stability across two renders, ≥ 50 pages, and that the sum appears exactly once on the last page. It never asserts that the last data row accompanies the footer — the "orphan tie" its name claims to hold.
- **Impact**: Cosmetic today; a future reader may credit this test with coverage it does not carry.
- **Suggested Resolution**: Either assert `R499` on the final page alongside the sum, or rename to match what it measures.
- **Related AC**: AC5 (heavy cadence, R12)

---

### AC-by-AC disposition

| AC | Disposition |
|---|---|
| **AC1** | Satisfied for the formatted path; **Blocker 1** for the underived-format path (values silently rounded). Align clause is well witnessed by drawn X. |
| **AC2** | Both arrivals (explicit + shape-2 derived) genuinely covered. **Blocker 1** applies to the third, undeclared arrival the AC did not enumerate. |
| **AC3** | Behaviourally correct in the tree (the expression is root-scoped and cannot be per-page), but **Major 5** — the AC's own non-vacuity mechanism is an empty `if` and the surviving assertion is self-implied. |
| **AC4** | **Blocker 2** — no footer-side witness; DW-7 retired against another story's guard. |
| **AC5** | Main rule satisfied and well anchored (partition + `rowIndex == 8` by value; part (a) verified honest, Claim 5). Its unsatisfiable arm is **Blocker 3**. |
| **AC6** | **Satisfied.** Positive and negative halves both assert positions by value in one test; the honest note about the missing-witness ordering is credited. |
| **AC7** | **Major 6** — implements half its own text; data-row partition never compared. |
| **AC8** | **Satisfied.** Asserted at the `Render()` layer, layer pinned in the doc comment, `/Type /Page` and `/Count` both checked, and the developer's discrimination evidence (AC7/AC5 tests staying green under AC8's deletion) is sound. |
| **AC9** | **Blocker 4** — zero is not proven to be by decision. Render-succeeds, Warning-present and terminate clauses are satisfied. |
| **AC10** | **Satisfied.** Evidence is a genuine re-render byte comparison, not a hash of committed files; no golden moved; R11 guards green (re-verified: 688 pass · 0 fail). |

**Restoration note:** every mutation above (M1, M2, M4, M5) and the one probe file were reverted **by hand** and the tree re-verified by SHA-256 (`table_footer.go` `68e8ae80…`, `table_render.go` `d5e46803…`, `table_footer_test.go` `4aa9e918…`) with a final whole-suite run at **688 pass · 0 fail** and bare `gofmt -l` clean. No production code, test, or file other than this story file was changed by this review.
