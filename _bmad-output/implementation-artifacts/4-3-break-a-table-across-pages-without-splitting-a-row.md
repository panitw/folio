---
baseline_commit: 903bf8f
---

# Story 4.3: Break a table across pages without splitting a row

**Epic:** 4 — A Go developer can render the golden report (**the C4 gate**)
**Story key:** `4-3-break-a-table-across-pages-without-splitting-a-row`
**Status:** `done`
**Covers:** **FR25**
**Consumes:** **D-4.2.2**'s carried row identity — this story is that ruling's one named consumer.
**First story to apply:** **D-000.79** (the creator names a deletion per output-producing AC).

**Primary invariant — the pagination model, verbatim** (`internal/layout/paginate.go:12-38`, rule 3):

> **FIT ENTIRELY.** No line is ever split across a page boundary. A line is placed on the first page
> whose window contains it ENTIRELY. Whitespace at the foot of a page is CORRECT TYPESETTING, not a
> defect.

**This story's job in one sentence:** extend that rule from *the line* to *the row* — make the ROW,
not the line, the unit that must fit entirely — without weakening the line rule, without touching
`MixedItemError`, and without absorbing Story 4.6's pathological case.

---

## Baseline, measured in this run at creation (HEAD `903bf8f`, tree clean)

| Gate | Command | Result |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **987 pass · 0 fail · 1 skip** |
| `lint` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` (module `lint`) | **115 / 0** |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` (module `hashmatrix`) | **3 / 0** |

All three measured independently at creation, not carried. The one skip is
`TestCorpusMeetsP6ExerciseFloors`, red **by design** (D-000.17 / D-2.1.14 / D-000.57 / D-000.74).
**Never "fix" it.** The `-skip` form above **is** the green gate.

> **rtk note (project convention).** A shell hook rewrites `go test` to `rtk go test`, and rtk's
> filtered output contains **no `--- PASS` / `--- FAIL` lines**. Take every count through
> `rtk proxy`, and redirect **outside** the proxied string — `rtk proxy` execs the string, so both a
> leading `VAR=x` and a trailing `2>&1` inside it fail as "no such file".

**Heavy-test cadence is per-epic (D-000.4).** This story **writes integration/e2e tests and does not
run them**, and must say so explicitly in its Delivery Log. The cross-target hash matrix runs once
at the Epic 4 boundary. 4.7 is the only Epic 4 story on D-000.4's per-story override list; **4.3 is
not.**

---

## In plain terms (read this first if you just want the gist)

When a table has more transactions than fit on one page, it now continues onto the next page without
ever splitting a row: a transaction's text, shading and borders always move together, and a row that
does not fit is pushed whole onto the next page — the leftover space is simply left blank, which is
correct typesetting, not a mistake. Page counts and printed page numbers come out right even for a
table spanning many pages.

Under the hood, the page-breaking machinery used to understand only single lines of text and single
pictures; it had no idea several lines and a band of shading belonged to one transaction. This work
teaches it that directly, rather than continuing to rely on a coincidence that used to hold rows
together most of the time.

A real problem was found and fixed before this shipped: an early version rejected an entire class of
perfectly ordinary pages — anywhere a table sat beside another item lined up with it at the same
height, even an unremarkable caption, the page failed to produce any output at all, blaming the
document for something it had not done wrong. That is corrected, with a permanent check confirming a
table can share a line with another element without trouble.

Two smaller things, recorded rather than hidden: a table's column headings are always exactly one
line today, so no ordinary document can prove the headings are held together on purpose rather than
by coincidence — a more artificial check stands in to prove that underlying machinery works. And the
message shown when one row is too tall for any page names it slightly generically on the path most
documents take; that wording predates this work and is left for later to improve.

It deliberately does not repeat column headings on continuing pages, handle totals rows, decide what
happens to one monstrous row taller than a page, or shade alternate rows — those are next. A heavier
end-to-end proof was drafted and held back, following this project's usual rhythm of running that
class of check only at wider milestones.

Done looks like: a table longer than a page prints every row exactly once, in the data's own order,
each row whole, with the page count right — and an ordinary page placing something else next to the
table keeps working exactly as it always did.

---

## Story

**As a** template author,
**I want** a row to stay whole,
**So that** a transaction never appears half on one page and half on the next.

---

## THE CENTRAL PROBLEM, restated against what is actually in the tree

Story 4.2's hand-off (AC7) and D-4.2.2 both describe the situation this story opens against. **Both
descriptions are slightly wrong in the same way, and the correction changes which options are open.**
See Divergence **D1**. Read it before designing anything.

The accurate statement:

- `layout.Paginate` has **no grouping concept**. Its atom is the `ColumnItem`, and each item is
  placed independently on the first page whose window contains it entirely.
- A data row does **not** reach `Paginate` as one item. It reaches it as **1 + N items**: one chrome
  item spanning the row's full extent (`table_render.go:715-723`), plus one text item per physical
  line (`table_render.go:733-777`), each with its own extent inside the row's.
- The identity that says "these 1 + N items are one row" **exists** — `tableRectSource.isDataRow` /
  `rowIndex` and `textRunSource.isTableRowLine` / `rowIndex` — and is **dropped on the floor** at
  both `ColumnItem` construction sites. See Divergence **D5**.

So this story's real work is: **carry the identity across the `ColumnItem` boundary, and make
`Paginate` honour it as a placement unit.**

---

## FIVE divergences between the brief / the record and the shipped tree, all measured in this run

Every number below was produced at `903bf8f` by the command shown. Nothing here is inferred.

### D1 — `MixedItemError` does **not** mean "exactly one line". It means "exactly one KIND."

The brief, D-4.2.2 and 4.2's AC7 all say `Paginate` *"takes `[]ColumnItem` and rejects anything that
is not exactly one line, with `MixedItemError` (`:303`, `:306`)"*. **Read against the code
(`paginate.go:290-311`), that is not what the check does.** The check counts how many of
`{Runs, Images, Rects}` are non-empty and errors when the count is **more than one** or **zero**. A
`ColumnItem` carrying **several** `TextRunRef`s is entirely legal and is the normal case: one
physical line of an N-column table is **one item carrying N runs** (`table_render.go:726-731`).

**Why this matters, and it is decisive.** It removes one option from D-4.2.2's list and strengthens
the case against another:

- *"Relax `MixedItemError`"* is not needed for multi-run rows — they are already legal.
- *"Collapse a row into one atomic item"* would require an item carrying **both** `Runs` and
  `Rects`, which is **precisely** the state `MixedItemError` exists to reject, and which
  `OverflowError.Kind` (derived `"image"` → `"table"` → `"line"`, `paginate.go:359-369`) would
  mislabel. **The brief forbids silently widening `MixedItemError`'s meaning, and this is the widening
  it forbids.** So the collapse option is out on the code's own grounds, not on taste.

What remains is a **grouping** concept, orthogonal to the exclusivity check. See DECISION-1.

### D2 — Row-wholeness **already holds** in the final page model at HEAD, incidentally and unguarded

**This is the single most important measurement in this document, and it is what makes a naive
acceptance test for AC1 pass vacuously.**

Because a data row's chrome item spans `rowTop .. rowBottom` (`table_render.go:715-720`), and every
one of that row's line items lies inside that extent (`rowHeight = padTop + FirstBaseline +
(n-1)*Advance + LastDescent + padBottom`, so the last line's bottom is `rowBottom − padBottom`), and
because `paginateDocument` appends **rects before text** (`render.go:1789-1808` before
`render.go:1815-1868`) so the stable sort visits the chrome item **first** on a tie — the chrome item
acts as an **atomic proxy for the whole row**. It fails the fit test first, slides the window to the
row's top, advances the page, and every line of that row then fits in the new window.

**Measured** with this scratch probe (written to
`folio-go/internal/layout/zzprobe_test.go`, run, then deleted; the tree is clean at `903bf8f`):

```go
package layout

import (
	"fmt"
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// row: rect item spanning [top,top+h], plus n line items inside it.
func rowItems(top, h geom.Length, n int, padTop geom.Length, rectFirst bool, base int) (rects []ColumnItem, lines []ColumnItem) {
	adv := (h - padTop) / geom.Length(int64(n))
	rects = append(rects, ColumnItem{ElementID: fmt.Sprintf("r%d", base), Top: top, Bottom: top + h, Rects: []RectRef{RectRef(base)}})
	for i := 0; i < n; i++ {
		lt := top + padTop + geom.Length(int64(i))*adv
		lines = append(lines, ColumnItem{ElementID: fmt.Sprintf("r%d", base), Top: lt, Bottom: lt + adv, Runs: []TextRunRef{TextRunRef(base*10 + i)}})
	}
	return
}

func TestZZProbe(t *testing.T) {
	for _, padTop := range []geom.Length{0, 2000} {
		var rects, lines []ColumnItem
		for k := 0; k < 6; k++ {
			r, l := rowItems(testContentTop+geom.Length(int64(k))*30000, 30000, 3, padTop, true, k)
			rects = append(rects, r...)
			lines = append(lines, l...)
		}
		var b []ColumnItem // PHASE B order: rects then lines
		b = append(b, rects...)
		b = append(b, lines...)
		var a []ColumnItem // PHASE A order: lines then rects
		a = append(a, lines...)
		a = append(a, rects...)
		for name, items := range map[string][]ColumnItem{"B(rects first)": b, "A(lines first)": a} {
			p, err := Paginate(testGeometry(), items)
			if err != nil {
				t.Fatalf("padTop=%d %s: %v", padTop, name, err)
			}
			t.Logf("padTop=%d %s: pages=%d", padTop, name, len(p.Pages))
			for i, pg := range p.Pages {
				t.Logf("   page %d shift=%d rects=%v runs=%v", i, pg.Shift, pg.ContentRects, pg.ContentRuns)
			}
		}
	}
}
```

Command:
`cd folio-go && env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -run TestZZProbe -v ./internal/layout/"`

Observed (six rows of 30,000mp each, three lines apiece, content window 18,000..118,000):

```
padTop=0    B(rects first): pages=2
   page 0 shift=0     rects=[0 1 2] runs=[0 1 2 10 11 12 20 21 22]
   page 1 shift=90000 rects=[3 4 5] runs=[30 31 32 40 41 42 50 51 52]
padTop=0    A(lines first): pages=2
   page 0 shift=0     rects=[0 1 2] runs=[0 1 2 10 11 12 20 21 22 30]   <-- ROW 3 IS SPLIT
   page 1 shift=90000 rects=[3 4 5] runs=[31 32 40 41 42 50 51 52]
padTop=2000 B(rects first): pages=2
   page 0 shift=0     rects=[0 1 2] runs=[0 1 2 10 11 12 20 21 22]
   page 1 shift=90000 rects=[3 4 5] runs=[30 31 32 40 41 42 50 51 52]
padTop=2000 A(lines first): pages=2
   page 0 shift=0     rects=[0 1 2] runs=[0 1 2 10 11 12 20 21 22]
   page 1 shift=90000 rects=[3 4 5] runs=[30 31 32 40 41 42 50 51 52]
```

**Two conclusions, and the developer must design against both:**

1. **In the real page model (order B), the boundary row already moves whole.** So *"write a test that
   a boundary row's parts share a page, watch it pass"* proves **nothing about this story's code**.
   The test would pass at `903bf8f` with no implementation at all. This is why AC1's named mutation
   below is **two-part**: the chrome must be removed first, or the measurement is of the accident.
2. **The guarantee rests on three facts nobody wrote down**: that a data row always emits a chrome
   item; that the chrome item spans the row's *full* extent; and that the stable sort puts it ahead of
   its own lines on a tie. **Any of the three can be broken by a later Epic 4 story without anything
   going red.** That is the case for building a mechanism rather than only asserting the outcome.

### D3 — The two pagination passes are **not** position-identical, and with zero top padding they partition a boundary row differently

`page_number.go:44-52` states that `contentColumnItems` (PHASE A, page-count-only) and
`paginateDocument` (PHASE B, the real page model) agree because *"content items are POSITION-IDENTICAL
between them"*. **Measured false at `903bf8f`.** They build the `items` slice in **different orders**
— PHASE A: text, then images, then rects (`page_number.go:70-125`); PHASE B: rects, then text, then
images (`render.go:1785-1896`) — and `Paginate`'s sort is **stable**, so the input order breaks ties
on `Top`.

A data row's chrome item and its first line item **tie on `Top` exactly when the body style's top
padding is zero**, which is the default. The probe above shows the consequence in the `padTop=0`
rows: PHASE A places row 3's first line on page 0 and the rest of row 3 on page 1.

**Harmless today** — PHASE A's `Pagination` is discarded except for `len(Pages)`, and the page counts
agreed in every configuration probed. **Not harmless once grouping exists**: if only one of the two
builders carries the grouping, the two passes can disagree on page count, and PHASE A's count is what
`Page X of Y` renders (D-2.7.2). See **AC3**.

### D4 — **No golden fixture renders a table**, so no golden can move through table behaviour

`rtk proxy "grep -rl '\"type\": *\"table\"' fixtures folio-go/testdata"` → exactly one hit:
`folio-go/testdata/template/golden/worked-example.json`, a **round-trip** fixture that is never
rendered (4.2's own note 3 records that it is not even a renderable document under D-4.1.1). The
`fixtures/` tree holds `page-count-1/5/20/50`, `multi-page/`, `wrapped-text/`, `three-band-page/`,
`font-text/`, `image-embed/`, `minimal-rect/`, `hidden-image/`, `multi-script-fallback/`,
`shaped-text/`, `expected-breaks/`, `thai-break-corpus/` — **none contains a table.**

**Consequence, and it is a scope relief:** the byte-neutrality exposure of this story is confined to
whether the `ColumnItem` / `Paginate` change alters pagination for **non-table** documents. It does
not run through any table. Do **not** record a new golden here — 4.7 owns the golden report.

### D5 — `isDataRow` / `isTableRowLine` / `rowIndex` have **no production reader** at HEAD

`rtk proxy "grep -rn 'isDataRow\|isTableRowLine\|rowIndex' folio-go"` → written at
`table_render.go:721-722` and `table_render.go:768-769`; read **only** in
`table_render_row_test.go` (and in the fields' own doc comments). **This story is the reader.** If
4.3 ships without one, D-4.2.2's inert-field hazard — the defect Story 4.1 spent a whole story
removing eight instances of — materialises in the story written to prevent it.

---

## Obligations this story inherits, and what it does NOT inherit

### Inherited and DUE HERE

1. **D-4.2.2's consumer obligation.** Row membership is carried and asserted; 4.3 must make it reach
   `Paginate` **without inference**. Do not reconstruct membership from element ids, extents or
   emission order — that reconstruction is the thing the ruling forbids.
2. **4.2's hand-off note (AC7).** *"The atomic unit does not exist in `layout` yet;
   `paginate.go:100-108`'s 'two items, same extent' comment is the thing 4.3 must revisit."* Revise
   that comment: after this story the two-items-same-extent shape is no longer the whole story of how
   a row holds together.
3. **D-000.79, first application.** Every output-producing AC below carries a **named deletion**. The
   developer **executes and records** each one, over the **whole suite** (never behind `-run`,
   D-000.9's extension). Where an AC produces no observable output, this document **says so** rather
   than inventing a weak mutation — see AC6.
4. **Repair `page_number.go`'s false "POSITION-IDENTICAL" claim** (D3). It is a comment asserting an
   invariant that stopped holding when 4.2 landed data rows. Either make it true, or state what is
   actually guaranteed. **Do not leave a false invariant in a doc comment.**

### Inherited, DEFERRED, and to be PLACED not absorbed

5. **D-2.8.6's within-band text-vs-table diagnostic ordering.** Pinned by 4.2's AC8, still unruled.
   4.3 adds no table diagnostics. **Leave the pin alone; do not change the ordering.**
6. **The "no resolvable `fontFamily`" diagnostic code** (4.1 Finding 13, 4.2 obligation 3). Still the
   owner's. **This story mints no diagnostic code** (D-000.65: 4.3 ships no new condition — the
   too-tall-row condition already shipped in 2.6 as `OverflowError`).
7. **`TableGeometry.Width()` overflow protection** (4.1 Finding 10 / DECISION-3). 4.3 adds no new
   reader of `Width()`. Not this story.

### Explicitly NOT inherited — the Epic 4 fences, drawn precisely

8. **Story 4.4 — the repeated header.** 4.3 paginates against the content window as `ContentHeight`
   defines it today. **A continuation page's available height does not shrink for a repeated header
   in this story.** 4.4 owns both repeating the header and accounting for its height on continuation
   pages. Do not anticipate it, and do not leave a hook for it that nothing reads.
9. **Story 4.5 — footer aggregates and orphan control.** `columns[].footer` / `footerOf` /
   `footerFormat` stay unread at render. Keeping a total row with at least one data row is a
   **different rule** from keeping a row whole and is 4.5's. **4.3's grouping unit is one row, never
   a run of rows.**
10. **Story 4.6 — a row taller than the page.** 4.6 owns clipping it to a fresh page and reporting a
    diagnostic. **4.3 preserves today's behaviour exactly: a located `OverflowError`.** See AC4. Do
    **not** clip, do **not** convert it into a `Diagnostic`, and do **not** widen `OverflowError`.
11. **Story 4.8 — alternating row styling.** `altRowBackground` stays inert.
12. **DW-16** (`ShapedGlyph.CID`), **DW-4 / DW-13 / DW-11 / DW-20** — owner-batched or elsewhere.
    Nothing here.

---

## Do not re-open — settled rulings this story inherits

- **The window model** (`paginate.go` package doc, D-2.6.1 + the sliding-window amendment). The
  window slides; the column is never mutated; the fixed grid is eliminated **by measurement** and must
  not be reproposed. **Grouping changes what "an item" is for the fit test — it must not change the
  window model.**
- **Line granularity is forced, not chosen** (rule 2). A row grouping sits **above** the line, never
  instead of it. A text element that is not a table row still paginates by line. **A row taller than
  the window must not be split into lines to "make it fit"** — that is exactly the straddle the model
  forbids, and the residual case is 4.6's.
- **No page is ever empty** (`TestPaginateNeverProducesAnEmptyPage`). Grouping must not create one.
- **Emission is in AUTHORED order per page** (`PageAssignment.ContentRuns` doc). Grouping decides
  *which page*; it must not reorder emission. Reordering reaches output bytes.
- **D-4.2.1's ranking:** `table.altRowBackground` (4.8) wins over `style.background`. Untouched here.
- **AD-24:** nothing negotiates, nothing reflows, `ContentHeight` is a function of `PageGeometry`
  alone. A row's measurement must not reach `PageGeometry`.

---

## R — design constraints derived from the record during creation

- **R1.** Grouping is **orthogonal to** `MixedItemError`, never a relaxation of it (D1). Each
  `ColumnItem` still carries exactly one of `{Runs, Images, Rects}`. **The exclusivity check is
  untouched.**
- **R2.** Every item **not** belonging to a table row must behave exactly as it does today — same
  page, same shift, same emission order. This is what keeps `page-count-1/5/20/50`, `multi-page/` and
  every other golden byte-identical (D4 confines the exposure to exactly this).
- **R3.** The grouping must be recoverable by **direct field lookup** from the row-generating code's
  own output, not reconstructed (D-4.2.2). Both `ColumnItem` builders must carry it (D3).
- **R4.** Integer millipoint arithmetic only. Addition, subtraction, comparison. **No division, no
  rounding, no float** (AD-23 / AD-2). A float conversion **compiles** — `lint`'s
  `TestFloatTypedProductionScan` is the instrument, not the compiler.
- **R5.** No map iteration may reach output order (D-1.3.5 / `ScanMapRange`). If grouping needs a
  lookup keyed by row identity, its **iteration** must be over a slice in a deterministic order.
- **R6.** Termination is a property, not a hope. The sweep must remain a single forward pass. A
  grouping that can re-visit an item is how 4.6's "no infinite loop" AC gets broken a story early.
- **R7.** A group's members are **contiguous in column order** by construction (a row's chrome and
  its lines occupy one contiguous vertical span, and rows do not interleave). Do **not** rely on that
  silently — if the implementation depends on contiguity, assert it, because 4.5's footer and 4.8's
  shading are both about to add items into that span.
- **R8.** **This story adds the repository's first multi-page table fixture.** 4.2's R8 fence
  ("every table fits on one page") is **lifted here and only here** — it was 4.2's fence, not a
  standing rule.

---

## Acceptance Criteria

Each output-producing AC carries a **named mutation** under D-000.79. The wording is **behavioural,
not mechanistic**, deliberately: the mechanism does not exist yet, and naming one from here would
over-constrain the design. The developer **executes** each one, over the **whole suite**, and records
the observed pass/fail counts and the named tests that reddened.

> **Read D2 before running any of these.** A boundary row already lands whole at `903bf8f`. A
> mutation whose "red" is invisible because the chrome accident is carrying the property is a
> measurement of nothing.

---

### AC1 — A row that does not fit in the remaining content height moves WHOLE to the next page

**Given** a table whose rows cross a page boundary, including at least one row that has **wrapped to
several physical lines**
**When** pagination runs
**Then** for **every** row, that row's chrome and **every one of its physical lines** — first line and
continuation lines alike — are assigned to **one and the same page**
**And** no row's parts appear on two pages.

- **How to assert it (D-000.68).** Anchor on the **carried row identity**, not on a count. For each
  page, collect the set of `rowIndex` values present among that page's items, **by kind** (chrome
  items and line items separately), and assert the two sets are **equal** for every page. `len(...) ==
  N` does **not** qualify — 4.2's own row-identity guard failed at exactly this and the reviewer's
  M2b mutation survived it. A per-page **set equality** keyed on identity can see a shuffle; a count
  cannot.
- **Second anchor, independent of the first.** Assert against **content**: bind each row to a distinct
  sentinel marker (the shape `TestDataRowIdentityIsConsistentAndDistinct` already uses — every word of
  a long value carries its own row's marker, so a continuation line's `SourceText` still does) and
  assert every run bearing row *k*'s marker lands on the same page. Two independent statements,
  D-3.4.6's shape.
- **Presence precondition.** The test must fail loudly if the fixture did **not** paginate
  (`len(pages) >= 2`) and if **no** row wrapped to `>= 2` physical lines. Without both, the test
  passes on a single-page table and asserts nothing.

**Named mutation (D-000.79) — TWO PARTS, and both are required:**

> **(a) The accident control, run FIRST.** *Remove whatever draws a data row's cell chrome, so that a
> row contributes only its text lines to pagination.* **AC1's own test must STILL PASS.** If it fails
> here, the row-wholeness guarantee is the chrome accident measured in **D2** and **not a mechanism**
> — AC1 is **not met**, whatever the unmutated suite says.
>
> **(b) The deletion screen, run on top of (a).** *Remove whatever keeps a row's several parts on one
> page — let each part be placed independently on the first page whose window contains it.* **The
> whole suite must redden**, and AC1's test must be among the reddened. If it stays green, there is
> **no witness at all** and any later modification-based red-proof is measuring something unrelated.

Record both, with commands and counts. Restore the tree after each.

---

### AC2 — Rows continue in data order across pages, with no duplication and no omission

**Given** a table longer than one page
**When** it paginates
**Then** every row of the bound collection appears **exactly once** across all pages
**And** the row order across pages follows the **collection's** order: page *p*'s `rowIndex` values all
precede page *p+1*'s.

- **How to assert it (D-000.33, quoted in `paginate_test.go`'s own header).** **Pin the partition by
  value.** A conservation law — *"every row appears on exactly one page"* — is satisfied by the
  degenerate paginator that puts everything on page 0, i.e. **by the defect**. So the test carries a
  **hand-derived literal table**: *page 0 holds rows 0..k, page 1 holds rows k+1..m*, with the
  arithmetic that produced `k` written out beside it. Derive nothing from the code under test.
- **And** assert no `rowIndex` appears on two pages, and that the union of all pages' `rowIndex` sets
  is exactly `{0 .. len(collection)-1}` — omission and duplication are **separate** failures with
  **separate** messages.
- **Order is asserted on the emitted page model**, not on an intermediate: authored emission order
  reaches output bytes.

**Named mutation (D-000.79) — two, run separately:**

> **(a)** *Remove whatever confines a row to a single page — let a boundary-crossing row be emitted on
> both the page it started on and the page it moved to.* The suite must redden on **duplication**,
> with a message naming duplication.
>
> **(b)** *Remove whatever preserves the collection's order when rows are distributed across pages —
> emit each page's rows in the order the paginator's sweep visited them rather than in the data's
> order.* The suite must redden on **order**.
>
> **Note for (b), from the measurement:** at `903bf8f` the sweep order and the data order coincide for
> a simple table, so this mutation may need the fixture to make them differ (a table whose rows are
> emitted in one order and declared in another, or a second element interleaved in the column). **If
> the developer cannot construct an input where they differ, say so explicitly in the Delivery Log
> and record that AC2's order half has no independent witness** — do not quietly report a green.

---

### AC3 — Row membership reaches `Paginate` without inference, and BOTH pagination passes use it

**Given** the row identity Story 4.2 carries on its row-generating output
**When** the `ColumnItem` list is built — by **either** builder
**Then** that identity is present on the items, recoverable by **direct field lookup**, never
reconstructed from element id, extent or emission order
**And** the early page-count-only pass and the final page-model pass produce the **same page count**
and the **same row-to-page partition**.

- **Grounds.** D-4.2.2's invariant, and Divergence **D3**: the two builders assemble `items` in
  different orders today and, at the default zero top padding, already partition a boundary row
  differently. PHASE A's count is what `Page X of Y` renders (D-2.7.2). **If only one builder carries
  the grouping, the printed page count can disagree with the document.**
- **How to assert it.** A test that builds **both** item lists from the same document and asserts the
  two `Pagination` results agree on `len(Pages)` **and** on the per-page `rowIndex` partition. Anchor
  on the identity, not on run indices — the two passes use **different** (local vs. global) ref
  numbering, so comparing `ContentRuns` directly is meaningless.
- **And** repair the false comment at `page_number.go:44-52` (obligation 4).

**Named mutation (D-000.79):**

> *Remove whatever makes the early, page-count-only pass see the same row grouping as the final pass —
> leave that pass grouping-blind, as it is today.* **The whole suite must redden.** If it stays green,
> the two passes cannot in fact disagree on this input, and the fixture is not exercising the
> divergence — widen it (the `padTop == 0` tie in **D2**'s probe is the known trigger) or record
> explicitly that the AC's witness is only the direct partition comparison.

---

### AC4 — A row that fits in **no** window still fails the way it fails today, and pagination terminates

**Given** a single table row taller than the whole content window
**When** pagination runs
**Then** the existing located `layout.OverflowError` is returned, with `Kind == "table"`, naming the
element — **exactly as at `903bf8f`**
**And** the sweep terminates: no loop, no straddle, no silent clip, no panic (AD-14).

- **Fence, stated twice because it is the easy scope leak.** **Story 4.6 owns this case's real
  answer** (clip to a fresh page, report a diagnostic, return PDF bytes). **4.3 changes nothing about
  it.** The only thing 4.3 owes is that introducing a grouping unit does **not** turn a too-tall row
  into an infinite loop, a straddle, or a differently-shaped error. A pinning test here must say **in
  its own words that it records current behaviour and does not endorse it**, naming 4.6 as the owner —
  D-4.2.3, which generalises exactly this situation.
- **Anchor.** Assert the error's **type** and its `Kind` and `ElementID` fields, plus that the returned
  `Pagination` is the zero value. Do **not** assert on the message string: it is prose and 4.6 will
  replace it.

**Named mutation (D-000.79):**

> *Remove whatever reports an error for a row that fits in no window — let the sweep proceed.* **The
> whole suite must redden.** A test that hangs instead of failing is also a red, and the developer
> should record which it was: a hang tells you the residual guard was load-bearing for **termination**
> and not only for the diagnostic, which is information 4.6 will want.

---

### AC5 — The header row is subject to the same rule, and 4.3 does **not** repeat it

**Given** a table whose **header** row does not fit in the remaining content height
**When** pagination runs
**Then** the header row's chrome and its column labels move to the next page **together**
**And** the header is **not** repeated on any continuation page, and the content height available on a
continuation page is **not** reduced for one.

- **Grounds.** The epic's AC says *"no row is ever split"*, and a header is a row for that purpose.
  The mechanism should not special-case it — a grouping that covers data rows but not the header row
  is a grouping with an unexplained hole in it. **But repeating the header is Story 4.4's**, and this
  AC's second half is the fence that keeps 4.3 from drifting into it.
- **Note.** Like AC1, this already holds by the chrome accident at `903bf8f`. The same two-part
  mutation applies.

**Named mutation (D-000.79):**

> Same two-part shape as AC1, applied to the header: **(a)** remove whatever draws the header row's
> cell chrome — the header-atomicity test must **still pass**; **(b)** on top of (a), remove whatever
> keeps the header row's parts on one page — the suite must **redden**.

---

### AC6 — Byte-neutrality, determinism, and the standing bans

**Given** every document in the repository that does not contain a table
**When** it renders
**Then** its bytes are **identical** to `903bf8f`, and no golden fixture moves
**And** `internal/` remains free of `time`, `os`, `math/rand`, `net`, `math` transcendentals,
package-level mutable state, output-reaching map iteration and **any** binary float
**And** the named standing guards stay green: `lint`'s `TestGlyphIdentifierCensus` (two producers of
`pagemodel.ShapedGlyph`) and `TestSpineStageLadderMatchesStageRankTable`; `folio-go`'s
`TestFolioMethodNamesAreInjective`, `TestValidateNeverReachesRenderOrInternalPDF`, the one-byte-producer
guard, and `TestPaginateNeverProducesAnEmptyPage`.

- **Scope relief, measured (D4).** No golden fixture renders a table. The exposure is confined to
  whether the `ColumnItem` / `Paginate` change perturbs **non-table** pagination — which R2 forbids.
- **The float instrument is `lint`, not the compiler** (R4): a float conversion compiles fine.

**Named mutation: NONE, and this is deliberate.**

> **This AC produces no new observable output — it asserts the ABSENCE of change.** A deletion screen
> is not definable for it: there is nothing to delete. Its witness already exists and is strong by
> construction — the golden suite, `lint`'s scans, and the named arch tests, all of which reddened for
> real reasons in earlier stories. **Stating this explicitly, per the brief, rather than inventing a
> weak mutation to fill the column.**

---

## Decisions

### DECISION-1 — TAKEN by this story, on D-4.2.2's explicit delegation: `Paginate` gains a **grouping** concept; a row is **not** collapsed into one item

D-4.2.2 ruled the **invariant** (*row membership must be recoverable without inference*) and left the
**mechanism** to the implementer, *"deliberately not named from outside the code"*. The brief asks
this story to decide the direction and state it. **Direction, with grounds:**

- **Collapse a row into one atomic `ColumnItem` — REJECTED, on measured grounds.** It requires an item
  carrying both `Runs` and `Rects`, which is exactly the state `MixedItemError` exists to reject
  (**D1**), and `OverflowError.Kind` would need a fourth answer. The brief forbids silently widening
  `MixedItemError`'s meaning; this is that widening.
- **A pre-pass assembling rows into atomic units before pagination — REJECTED for the same reason**
  in its natural form (the assembled unit is a mixed item), and rejected in its "parallel structure"
  form because it re-derives membership outside the type that carries it, which is the seam
  D-4.2.2 exists to close.
- **A grouping concept, orthogonal to the exclusivity check — TAKEN.** Items keep exactly one kind
  each; a group is a separate statement about which items must share a page. It carries D-4.2.2's
  identity by direct lookup, leaves `MixedItemError` untouched, and extends naturally to the header
  row (AC5) without extending to a run of rows (which is 4.5's, and must stay out).

**The mechanism itself — the field's name, its shape, whether it lives on `ColumnItem` or beside it,
how `Paginate` consumes it — is the implementer's call and is deliberately not named here.** Naming a
mechanism from outside the code is a mistake this programme has recorded before (D-4.2.2's own
wording). What is fixed is the invariant, R1–R7, and the ACs.

### DECISION-2 — FLAGGED to the engineering lead (non-blocking; recommendation given): build the mechanism, or assert the accident?

**The honest question raised by Divergence D2.** Row-wholeness already holds at `903bf8f`. A cheaper
story exists: write the tests, ship no mechanism, close FR25. **Recommendation: reject the cheaper
story.** Three grounds:

1. **D-4.2.2 forbids it by name.** The identity would remain inert with no production reader — the
   exact hazard that ruling closed, in the story named as its consumer.
2. **The accident has three unstated preconditions** (D2), and **Stories 4.5 and 4.8 both add items
   into a row's vertical span.** An unstated precondition that three upcoming stories can break is not
   a guarantee.
3. **It is invisible.** A green test whose greenness comes from a mechanism nobody described is
   D-4.2.3's failure and D-3.4.3's: an artifact we produced read as evidence of a decision we made.

**Not blocking**: if the lead disagrees, AC1/AC5's part (a) controls become the story and the ACs
shrink accordingly. The orchestrator should route this if the lead wants it, but the story is written
to be executed as-is.

### DECISION-3 — SURFACE, do not take: should the two-pass pagination exist at all?

`page_number.go`'s PHASE A exists only to learn the page count before header/footer text is collected
(D-2.7.2). Its comment claiming position-identity is **false** (D3), and this story repairs the
comment and forces the two passes to agree. **Whether the two-pass design should be collapsed to one
is a larger question** touching D-2.7.2's reservation, and belongs to the lead — likely at the Epic 4
boundary, when 4.4 and 4.5 have both added inputs to it. **Do not collapse it in this story.**

---

## Things the schema and the record could not resolve, surfaced rather than invented

1. **A group's extent: the union of its members, or the row's chrome extent?** At `903bf8f` the two are
   **equal by construction** (`rowBottom = last line's bottom + padBottom`), so the choice is
   **unobservable today** and no fixture can distinguish them. It becomes observable the moment
   anything is placed outside the chrome rect — 4.5's footer aggregates and 4.8's shading are both
   candidates. **Pick the union**, because it is the definition that stays correct when the accident
   ends, and say in the code why the two are indistinguishable at this commit. Recorded here so a
   later story does not read a passing test as a ruling.
2. **No `rowHeight` field exists in the schema** (4.2's note 2). Row height is wholly derived, so an
   author has no way to pin a uniform row height and no way to ask for a page break at a chosen row.
   **Do not add a field.** If the golden report needs either, that is a schema question for the owner
   at 4.7.
3. **A table declared in the `pageHeader` / `pageFooter` band** is repeated **verbatim** on every page
   and never paginates (`BandContent`). Its rows are therefore outside this story's grouping entirely.
   `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` (`table_render_test.go:831`) must stay green —
   **confirm it does; do not "fix" it into an error.**
4. **Whether a page may legitimately hold zero data rows of a table while holding its chrome** — e.g.
   a table whose header lands at the very foot of a page under 4.4's repeated header. Not reachable in
   this story (4.3 does not repeat the header), but 4.4 will meet it. Flagged, not answered.

---

## Task breakdown

1. [x] **Reproduce the baseline.** Run all three gates and confirm **987 / 0 / 1**, **115 / 0**,
   **3 / 0**. If any differs, stop — the evidence in this document was taken at `903bf8f` and needs
   re-taking before anything else.
2. [x] **Re-run D2's probe verbatim** (it is reproduced above in full) and confirm the two conclusions:
   the boundary row already lands whole in order B, and splits in order A at `padTop == 0`. **This is
   the story's pre-fix measurement and everything downstream depends on it.** Delete the probe file
   afterwards.
3. [x] Build the **first multi-page table fixture** in the repository (R8). Small, hand-derivable
   geometry in the spirit of `paginate_test.go`'s `testGeometry()` — pairwise-distinct inputs so a
   substitution of one for another cannot survive — with at least one **wrapped** row and at least one
   row landing **exactly** at a boundary. Write out the arithmetic beside the expectations.
4. [x] Carry the row identity across the `ColumnItem` boundary in **both** builders
   (`page_number.go`'s `contentColumnItems` and `render.go`'s `paginateDocument`). **AC3.**
5. [x] Teach `Paginate` the grouping unit, leaving `MixedItemError` and the window model untouched.
   **AC1, AC5.** Single forward pass (R6).
6. [x] Confirm the residual too-tall case is unchanged, with a pinning test that says it records rather
   than endorses and names 4.6. **AC4.**
7. [x] Assert the partition **by value** (AC2) and the per-page identity **sets** (AC1) — never a
   count (D-000.68).
8. [x] Repair `page_number.go:44-52`'s false position-identity claim, and revise
   `paginate.go:100-108`'s "two items, same extent" comment per 4.2's hand-off note.
9. [x] **Run every named mutation above, in order, over the WHOLE suite** (never behind `-run`,
   D-000.9's extension). Record for each: the command, the observed pass/fail/skip counts, and the
   **named tests** that reddened. Restore the tree after each. **AC1's and AC5's part (a) controls come
   first** — a green part (b) after a failed part (a) is not evidence.
10. [x] Confirm byte-neutrality: every golden unchanged, all three gates green. **AC6.**
11. [x] Write the integration/e2e tests. **Do not run them** (D-000.4, per-epic cadence) and **say so
    explicitly** in the Delivery Log.
12. [x] Delivery Log: gate results for all three modules, every red-proof with its command and counts,
    the per-AC verification table, the D-000.79 **Class A / Class B split** for this story (the ruling
    makes 4.3 and 4.4 the decidable test of its own diagnosis — record the split so the test can be
    run), and the list of things deliberately not built.

---

## Delivery Log

### 0. Decisions received mid-story

- **DECISION-2 RULED (build the mechanism).** Applied in full: `layout.ColumnItem` gained a `Group
  ItemGroup` field (`ItemGroupKey{ElementID, IsHeader, Index}`), orthogonal to the `Runs/Images/Rects`
  exclusivity check. `MixedItemError` untouched (R1). The row identity already carried by
  `tableRectSource`/`textRunSource` (Story 4.2) is copied onto `ColumnItem.Group` by two new,
  narrowly-scoped methods — `tableRectSource.chromeRowGroup()` and `textRunSource.lineRowGroup()` —
  called from all three `ColumnItem`-construction sites (`page_number.go`'s `contentColumnItems`,
  `render.go`'s `paginateDocument`, both loops). No reconstruction from element id/extent/order
  anywhere (R3). The header row was given the SAME grouping treatment as a data row
  (`isHeaderRow`/`isHeaderLabel`, new fields, mirroring `isDataRow`/`isTableRowLine`) rather than
  relying on its pre-existing "same extent" coincidence — see finding 9.4 below on why that coincidence
  alone is not evidence of the mechanism for AC5.
- **DECISION-3 RULED (repair the comment; no deferred item; land an assertion instead).** Applied:
  `page_number.go:37-69`'s docstring no longer claims PHASE A/B are "position-identical" — it states the
  measured true fact (different append order, tie-broken by stable sort, harmless before this story
  because nothing read the per-item partition) and names `TestBothPaginationPassesAgreeOnRowPartition`
  as the assertion that replaces the false comment. No deferred item was filed for collapsing the
  two-pass design — see observation 9.6.
- **"Hold the golden" constraint — satisfied trivially.** No golden fixture was recorded from the new
  multi-page table fixture. D4 (this story's own divergence, unchanged) already established no golden
  renders a table; this story adds none. The AC3 comment repair and the two-pass agreement guard
  (`TestBothPaginationPassesAgreeOnRowPartition`) both land and are green (see §3, §5) — the sequencing
  the lead required is met, and there is nothing to "stop and report" because no golden was ever queued.
- **Lint diagnostics (rangeint x3, S1039 x1) in `table_pagination_test.go` — fixed.** Three
  `for i := 0; i < N; i++` loops rewritten as `for i := range N`; one `fmt.Sprintf` call with no format
  verbs rewritten as a plain string literal. None touched fixture-row construction arithmetic or a
  content marker's byte content — re-run of the affected tests after the edit showed identical pass
  counts (§9.7).

### 1. Baseline, reproduced at `903bf8f` in this run

| Gate | Command | Result |
|---|---|---|
| `folio-go` | `cd folio-go && env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **987 pass · 0 fail · 1 skip** |
| `lint` | `cd lint && env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **115 / 0** |
| `hashmatrix` | `cd hashmatrix && env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **3 / 0** |

Matches the story's own baseline exactly. `gofmt -l .` and `go vet ./...` clean in all three modules,
before and after every change below.

### 2. D2's probe, re-run verbatim (`internal/layout/zzprobe_test.go`, written, run, deleted)

Command: `cd folio-go && env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -run TestZZProbe -v ./internal/layout/"`

Observed, identical to the story's own record:

```
padTop=0    B(rects first): pages=2   page0 rects=[0 1 2] runs=[...22]      page1 rects=[3 4 5] runs=[30...52]
padTop=0    A(lines first): pages=2   page0 rects=[0 1 2] runs=[...22 30]   page1 rects=[3 4 5] runs=[31...52]   <- ROW 3 SPLIT
padTop=2000 B(rects first): pages=2   (no split)
padTop=2000 A(lines first): pages=2   (no split)
```

Both conclusions confirmed: the boundary row already lands whole in order B; it splits in order A at
`padTop == 0`. Probe file deleted; `git status` shows it untracked-and-gone (never committed).

### 3. The mechanism (DECISION-1's shape) and where it lives

- `internal/layout/paginate.go`: `ColumnItem.Group ItemGroup{Present bool; Key ItemGroupKey}`,
  `ItemGroupKey{ElementID string; IsHeader bool; Index int}`. Zero value ungrouped (R2).
- `Paginate` (single forward pass, R6): one pre-pass over the already-sorted `order` slice computes,
  per group key, the UNION extent (min `Top`, max `Bottom`) across its members, **and asserts
  contiguity** (R7) — if two items sharing a key are separated by anything else in column order, this
  is a located internal error, never a silent mis-placement. The main sweep substitutes a grouped
  item's own `Top`/`Bottom` for its group's union extent when testing fit and computing overflow height;
  emission is untouched (still authored order, per-`ColumnItem` index). A final belt-and-suspenders pass
  asserts every member of a group landed on the same page (R6), returning a located internal error
  if not — never reachable given R7 holds, checked anyway.
- **The union-vs-chrome-extent question, resolved per the story's own instruction**: the group's extent
  is the **union** of its members, not any one member's own rect. At this commit the two are equal by
  construction (`rowBottom = last line's bottom + padBottom`, table_render.go), so no fixture here can
  distinguish them — recorded so 4.5/4.8 do not read a passing test as a ruling on the question.
- `table_render.go`: `tableRectSource` gained `isHeaderRow bool` and method `chromeRowGroup()
  layout.ItemGroup`, set at both rect-construction sites (header: `isHeaderRow: true`; data row:
  `isDataRow`/`rowIndex`, unchanged fields, reused by the same method).
- `render.go`: `textRunSource` gained `isHeaderLabel bool` and method `lineRowGroup()
  layout.ItemGroup`, set on header-label runs; data-row line runs reuse the existing
  `isTableRowLine`/`rowIndex` fields via the same method.
- `page_number.go`'s `contentColumnItems` and `render.go`'s `paginateDocument` both call
  `chromeRowGroup()`/`lineRowGroup()` at every `ColumnItem` construction site that can carry a table
  item — **AC3's "both builders" requirement**, direct field lookup, no reconstruction.
- Method names `chromeRowGroup`/`lineRowGroup` (not a shared `rowGroup` name) — required by
  `TestFolioMethodNamesAreInjective` (package `folio` methods are keyed by name alone across ALL
  receiver types; a shared name on two receivers is exactly the collision that guard exists to catch,
  and did catch, during development).

### 4. Fixtures added (R8 — first multi-page table fixture)

- `internal/layout/paginate_group_test.go` — hand-derived `layout.ColumnItem`/`ItemGroup` fixtures, no
  fonts, geometry identical in spirit to `paginate_test.go`'s `testGeometry()`.
- `table_pagination_test.go` (package `folio`) — a real `.folio` document (custom 200×150pt page,
  10pt margins, one two-column table, 8pt body font), 20 bound rows with per-row content markers
  (`"RxW-"`, matching `TestDataRowIdentityIsConsistentAndDistinct`'s own pattern), one row (index 5)
  given a long value that wraps to >= 2 physical lines, and a second fixture (`headerPushedToNextPageDoc`)
  with a filler element that pushes the table far enough down the content band that the header itself
  does not fit page 0's window at all.

### 5. Per-AC verification

| AC | How verified | Tests |
|---|---|---|
| AC1 | Set equality (D-000.68) of rowIndex among chrome vs. line items per page, PLUS an independent content-marker anchor, PLUS a presence precondition (>= 2 pages, >= 1 wrapped row) | `TestDataRowNeverSplitAcrossPages` (folio), `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder` (layout, both append orders) |
| AC2 | Partition pinned BY VALUE (layout level, hand-derived); omission/duplication asserted separately (folio level); order asserted on the emitted model | `TestPaginateGroupPartitionPinnedByValue`, `TestPaginateGroupEmissionStaysInAuthoredOrder` (layout); `TestTableRowsAppearExactlyOnceInDataOrderAcrossPages` (folio) |
| AC3 | Both builders carry the SAME `Group`; a test builds both item lists from the same underlying `tableRects`/`contentRuns` and compares page count AND per-row page-SETS (not a last-page-seen scalar — see 9.3) | `TestBothPaginationPassesAgreeOnRowPartition` |
| AC4 | Located `*layout.OverflowError`, correct `ElementID`, zero-value `Pagination`, presence precondition (row taller than window). **CORRECTED (finisher review, Finding 4/Major):** `Kind` is ORDERING-DEPENDENT — `"table"` under PHASE B's item order (rects before text), `"line"` under PHASE A's (text before rects), which is the ordering the shipped `Render()` path actually uses. Both arms are now pinned against the same document (`tooTallRowDoc`), neither left unmeasured. | `TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow` (folio, PHASE B via `itemsForTest`, `Kind == "table"`), `TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender` (folio, PHASE A via the public `Render()`, `Kind == "line"` — NOT a regression, identical at `903bf8f`), `TestPaginateGroupTallerThanWindowReturnsLocatedOverflow` (layout) |
| AC5 | Header labels share one page (core, chrome-independent); chrome/labels cross-checked when chrome exists; header absent from every OTHER page. **CORRECTED (finisher review, Finding 2/Blocker):** this folio-level test is a genuine, honest Class A control (unfailable, by construction of the shipped single-line header) — it is NOT evidence that the `IsHeader` grouping key branch is exercised. See `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer` for that witness. | `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` (folio); `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer` (layout, new) |
| AC6 | Full clean suite (1000/0/1), `gofmt`/`go vet` clean in all 3 modules, no golden moved, named standing guards green | full-suite runs, §1/§7 |

### 6. Named mutations, run over the WHOLE suite, never behind `-run` — tree restored (`cp` from a
   pre-mutation backup, never `git checkout`) after every one

Baseline for comparison: **1000 pass · 0 fail · 1 skip** (987 baseline + 13 new tests/subtests this
story added).

**AC1/AC5, two-part (D-000.80/D-4.3.1's own instrument) — part (a) is the load-bearing half:**

| Mutation | Command | Result | Named tests |
|---|---|---|---|
| AC1(a): remove data-row chrome append (`table_render.go`, the `isDataRow` `rectSources` append) | `go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./...` | **983 / 17 / 1** | `TestDataRowNeverSplitAcrossPages` **STILL PASSES** ✓ (17 unrelated table-rect-count tests reddened, expected — they assert on rect counts directly) |
| AC1(b): on top of (a), disable Group substitution in `Paginate` (`effectiveTop/Bottom := it.Top, it.Bottom` unconditionally) | same | **978 / 22 / 1** | `TestDataRowNeverSplitAcrossPages` **REDDENS** ✓ (plus `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder`, `TestPaginateGroupPartitionPinnedByValue`, `TestBothPaginationPassesAgreeOnRowPartition`, `TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow`) |
| AC5(a): remove header chrome append (`table_render.go`, the header `rectSources` append) | same | **983 / 17 / 1** | `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` **STILL PASSES** ✓ |
| AC5(b): on top of (a), disable Group substitution (same edit as AC1(b)) | same | **979 / 21 / 1** | `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` **DOES NOT REDDEN** — see finding 9.4, a genuine measured divergence from the story's "same two-part mutation applies" note, not an oversight |

**AC2, two independent mutations:**

| Mutation | Command | Result | Named tests |
|---|---|---|---|
| AC2(a): duplication | **No deletion mutation exists to run** — see finding 9.1 | — | — |
| AC2(b): emit in sweep/column order instead of authored order (`Paginate`'s emission loop, `for i := range items` → `for _, i := range order`) | same | **998 / 2 / 1** | `TestPaginateGroupEmissionStaysInAuthoredOrder` (this story) and the pre-existing `TestPaginateEmitsContentInAuthoredOrder` **REDDEN** ✓. `TestTableRowsAppearExactlyOnceInDataOrderAcrossPages` (folio level) does **not** redden — see finding 9.2, matching the story's own prediction |

**AC3:**

| Mutation | Command | Result | Named tests |
|---|---|---|---|
| Leave PHASE A grouping-blind (`page_number.go`'s `contentColumnItems`, both `Group:` assignments replaced with the zero value) | same | **999 / 1 / 1** | `TestBothPaginationPassesAgreeOnRowPartition` **REDDENS** ✓ (only test affected) |

**AC4:**

| Mutation | Command | Result | Named tests |
|---|---|---|---|
| Remove the overflow report, let the sweep proceed (`Paginate`'s `itemHeight > height` branch deleted) | `go test -count=1 -timeout 60s -skip ... -v ./...` | **988 / 12 / 1**, completed in ~2s — **no hang** | `TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow`, `TestPaginateGroupTallerThanWindowReturnsLocatedOverflow`, `TestPaginateReportsAnItemThatFitsNoWindow` (+2 subtests), plus 3 pre-existing render-level overflow tests (+3 subtests) |

**AC6:** no mutation, per the story's own instruction — this AC asserts absence of change. Verified by
the clean full-suite runs bracketing every mutation above (§1, and the restore-check after each
mutation below all show exactly **1000 / 0 / 1**).

Every mutation above was restored via `cp` from a pre-mutation backup file (never `git checkout`), and
a clean **1000 / 0 / 1** was reconfirmed after each restoration before the next mutation was applied.
Final state, confirmed: `git diff` shows only the intended implementation (no leftover mutation code).

### 7. Byte-neutrality (AC6)

- `folio-go`: **1000 / 0 / 1**, `gofmt -l .` empty, `go vet ./...` clean.
- `lint`: **115 / 0**, `gofmt -l .` empty, `go vet ./...` clean.
- `hashmatrix`: **3 / 0**, `gofmt -l .` empty, `go vet ./...` clean.
- No golden fixture moved (no golden was ever recorded from the new fixture — D4 confines this story's
  exposure to non-table pagination, and R2 is exactly what keeps `page-count-1/5/20/50`, `multi-page/`
  etc. byte-identical: every `ColumnItem` this story does not itself tag keeps `Group.Present == false`,
  behaving exactly as before).
- Named standing guards, confirmed green throughout: `TestGlyphIdentifierCensus`,
  `TestSpineStageLadderMatchesStageRankTable` (lint); `TestFolioMethodNamesAreInjective`,
  `TestValidateNeverReachesRenderOrInternalPDF`, the one-byte-producer guard, and
  `TestPaginateNeverProducesAnEmptyPage` (folio-go).
- `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` confirmed green and its doc comment amended with
  an explicit "must not be fixed" note (table_render_test.go), per the coordinator's request — a table
  in `pageHeader`/`pageFooter` repeats verbatim and never paginates; this is BandContent's own contract,
  untouched by this story.

### 8. Integration/e2e tests — written, one attempted-and-withdrawn, cadence honoured (D-000.4)

This story's own integration-level tests (`table_pagination_test.go`) exercise the real
`collectBandTableRuns`/`collectBandTextRuns`/`layout.Paginate` pipeline together — the same shape Story
4.2's own `table_render_row_test.go` integration tests take — and **were run** as part of this story's
gate: they are not the heavy, cross-target class D-000.4 exists to gate.

**One heavier, `Render()`-to-PDF-bytes e2e smoke test was drafted** under the repository's
`//go:build matrix` convention (the same tag `matrix_test.go`/`expected_breaks_signoff_matrix_test.go`
already use for tests deferred to an epic-boundary gate run). Building it (`go vet -tags matrix ./...`,
which compiles but never executes the test body) succeeded. **It was withdrawn rather than committed**:
running the full suite with it present reddened `TestEpic2GateObligationsMatchTheDeclaredSet`, which
correctly rejects any new `//go:build matrix` file as an unauthorised new gate obligation — D-2.5.1
requires an explicit ruling before one is added, and this story was not given one. Adding it anyway
would have been exactly the kind of undecided scope expansion this developer role must not make
unilaterally. **Deferred, with its owner named**: whoever next needs a full `Render()`-to-PDF-bytes
table pagination check (plausibly 4.7, which owns the golden report) should raise the `matrix`
obligation explicitly and can reuse this story's fixture shape (`e2eMultiRowTableDoc`/
`e2eMultiRowTableData`, described here since the draft file itself was deleted, not committed).

### 9. Findings, worth recording precisely

**9.1 — AC2(a) "duplication" has no discriminating deletion mutation, and this is a structural fact
about `Paginate`, not a gap in this story's coverage.** `Paginate`'s emission loop writes `pageOf[idx]`
exactly once per authored index and appends each item's content to exactly one page's slice; there is
no code path whose single deletion produces "the same item on two pages" without inventing new
mechanism (as opposed to deleting existing mechanism, which D-000.79 requires). `TestTableRowsAppearExactlyOnceInDataOrderAcrossPages`'s
duplication check (`seenCount[row] > 1`) is asserted as a genuine structural invariant of `Paginate`'s
architecture, unconditional and unrelated to the Group mechanism this story adds — recorded explicitly
per AC6's own precedent (D-000.79: state rather than invent a weak mutation) rather than reported as a
silent green.

**9.2 — AC2(b) "order" has no witness at the folio/table level, exactly as the story predicted, and the
`layout`-level witness is the one that counts.** For any single table, `rowTop_{k+1} = rowTop_k +
rowHeight_k`, so a table's own rows are ALWAYS laid out in strictly increasing Top order matching data
order — sweep (column) order and authored order provably coincide for one table's rows, and no
single-table fixture can discriminate AC2(b)'s mutation. `TestPaginateGroupEmissionStaysInAuthoredOrder`
(layout level) is the actual witness: it constructs two rows where the LATER-declared row sits ABOVE
the earlier-declared one, discriminating sweep order from authored order directly, and it reddened
under the mutation. Recorded per the story's own explicit instruction rather than silently reporting a
folio-level green as if it were evidence.

**9.3 — the FIRST draft of `TestBothPaginationPassesAgreeOnRowPartition` had exactly the D-000.79 Class
B defect (a guard asserting a proxy) and was caught and fixed before being recorded as a red-proof.**
It aggregated "the page a row's lines landed on" with a bare assignment inside a `for p, pg := range
Pages` loop — last write wins. Under the AC3 mutation, row 5 (the wrapped row) genuinely split (line 0
on page 0, the rest — including the mis-grouped-in-PHASE-A chrome-adjacent lines — on page 1), but the
loop's last write (page 1, since pages are iterated ascending) happened to equal PHASE B's correct
answer, masking the divergence. Fixed to track a SET of pages per row per phase and assert
`len(set) == 1` before comparing across phases (matching D-000.68's own set-equality discipline). Both
`TestDataRowNeverSplitAcrossPages` and `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` were checked
against the same class of defect during their own construction (chrome-vs-line set equality going
vacuously true when chrome is entirely absent under mutation (a)) and built with a CORE, chrome-
independent invariant plus a SECONDARY, chrome-present-only cross-check from the start — see §5's
table.

**9.4 — AC5's mutation (b) does not redden at the folio/integration level, and this is a genuine,
measured divergence from the story's own "the same two-part mutation applies" note for the header — not
an oversight.** A data row's chrome and its Nth physical line generally have DIFFERENT extents (a
wrapped row's later lines start well after the chrome's own top), so grouping is load-bearing for them
— confirmed by AC1(b) reddening. A table's HEADER, by contrast, is unconditionally exactly one physical
line (Story 4.1's AC9): every column label's `itemTop`/`itemBottom` is set to `tableTop`/`tableBottom`
directly (table_render.go), IDENTICAL to the header chrome's own extent, not merely similar. Because
`Paginate`'s fit test is a pure function of an item's own `Top`/`Bottom`, two items with EQUAL extent
are placed on the same page by ANY implementation, grouped or not — there is no folio-level fixture
using only the header's own rect+labels that can discriminate "grouped" from "ungrouped" for this
specific pair, because the two are not merely usually-tied (D2's accident) but ALWAYS EXACTLY tied by
Story 4.1's own single-line-header construction. This is a THIRD, previously-unnamed precondition
(alongside D2's chrome-atomic-proxy and D3's append-order tie), and it is fully robust in-story-scope:
AC5's fix is still real and still tested (part (a) proves it — see §6), and its underlying MECHANISM
(the identical generic `Group` substitution code AC1 uses) is independently witnessed at the `layout`
level by `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder`/`TestPaginateGroupPartitionPinnedByValue`,
which use group keys with no special-casing for `IsHeader` and reddened under the identical mutation.
Flagged here, per D-000.79, rather than reported as a silent green with no explanation.

**9.4 CORRECTED (finisher review, Finding 2/Blocker and Major 1).** The appeal above to
`TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder`/`TestPaginateGroupPartitionPinnedByValue` as
evidence for AC5 is a CATEGORY ERROR, exactly as the reviewer found: both tests build their
`ItemGroupKey` with `IsHeader: false` (an `Index`-scoped, DATA-ROW key). They witness that `Paginate`'s
generic substitution code works for AN `ItemGroupKey`, full stop — they say nothing about whether the
`IsHeader: true` branch is itself exercised, because a `Paginate` bug scoped ONLY to `IsHeader: true`
(for instance, a typo swapping `IsHeader` and `Index` in the group-extent lookup) would leave both
cited tests green. Confirmed independently in this review: mutating production
(`chromeRowGroup`/`lineRowGroup`'s `isHeaderRow`/`isHeaderLabel` branches to return `layout.ItemGroup{}`)
leaves the WHOLE SUITE GREEN (1000/0), because Story 4.1's header is unconditionally one physical line
— chrome and every column label share `tableTop`/`tableBottom` EXACTLY — so no fixture built from the
real row-generating pipeline can ever tell "grouped" from "ungrouped" for this specific pair. That is a
correct, honest fact about the SHIPPED header, not a defect to paper over; it is also not evidence for
the `IsHeader` key branch, which the two cited tests never touch.

The gap is closed, not by trying to make the shipped header exercise a distinction it structurally
cannot produce, but by landing a witness for the `IsHeader` key branch ITSELF:
`TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer`
(`internal/layout/paginate_group_test.go`) builds an `ItemGroupKey{IsHeader: true}` group by hand whose
two members have DIFFERENT extents (a shape the shipped header cannot produce, but `Paginate` itself
does not know that) and asserts, in one sub-test, that grouping forces them onto the same page, and in
a second sub-test (the SAME two items with `Group` cleared) that they would otherwise split — a
self-contained two-part proof, in the D-000.79 spirit, that the `IsHeader` branch specifically is
load-bearing in `Paginate`. **The corrected record is therefore: AC5's folio-level "part (a) still
passes" claim is TRUE and UNFAILABLE — Class A, honestly labelled, a fact about the header's own
construction — and the `IsHeader` key branch itself now has a genuine, failable, layout-level witness
that did not exist before this correction.** See Finding Resolutions, Finding 2, below.

**9.5 — the negative precondition check the lead reported (unstyled tables still emit chrome) was relied
upon, not re-run.** `table_render.go`'s data-row `rectSources` append is unconditional
(`buildCellRect` takes `hasBackground`/`hasBorder` as arguments; the rect exists regardless), confirmed
by reading the code at the same lines the lead cited; this story's fixtures use a table with no explicit
border/background, and AC1/AC5 still measure correctly, consistent with that finding.

**9.6 — DECISION-3's replacement assertion, not a deferred item.** No entry was added to any deferred-
items list for collapsing the two-pass pagination design. `TestBothPaginationPassesAgreeOnRowPartition`
is the assertion that makes any future divergence between the two passes reddening at the commit that
causes it — the collapse question is now an optimisation with no owner and no ticket, exactly per the
lead's ruling.

**9.7 — the four lint diagnostics reported mid-story were all mechanical and non-load-bearing.** The
`rangeint` rewrites touched only loop counters over `rows`/marker-index ranges, never geometry
arithmetic. The `fmt.Sprintf` rewrite (the AC4 fixture's literal `.folio` document) carried no format
verbs. Re-running the affected tests after the edit (`TestDataRowNeverSplitAcrossPages`,
`TestTableRowsAppearExactlyOnceInDataOrderAcrossPages`, `TestBothPaginationPassesAgreeOnRowPartition`,
`TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow`) showed identical results before and
after.

**9.8 — DW-14's corpus maximum may move, and this is expected, not a defect.** Not acted upon further,
per the coordinator's own note.

### 10. D-000.79 Class A / Class B split for this story

| AC | Class | Detail |
|---|---|---|
| AC1 | **Neither** — properly witnessed | Two-part mutation, part (a) passes, part (b) reddens (§6). The FIRST draft of the folio-level test was Class-A-vulnerable (chrome-vs-line equality vacuous with no chrome) and was corrected before being recorded — see 9.3. |
| AC2(a) duplication | **No witness exists, stated explicitly** | Not Class A in the sense of "an absent guard that should exist" — the property is a structural invariant of `Paginate`'s construction with no plausible single-deletion mutation. See 9.1. |
| AC2(b) order | **Class A at the folio level, properly witnessed at the layout level** | No single-table fixture can discriminate it (9.2); `TestPaginateGroupEmissionStaysInAuthoredOrder` does. |
| AC3 | **Neither** — properly witnessed, after a Class-B defect was found and fixed pre-recording | The FIRST draft aggregated "last page seen" (a proxy) rather than a page SET — see 9.3. Fixed before being recorded as evidence. |
| AC4 | **Neither** — properly witnessed | Mutation reddens the whole suite; no hang (§6). |
| AC5 | **CORRECTED (finisher review) — Class A at the folio level, honestly labelled; the `IsHeader` key branch itself is now properly witnessed at the layout level** | The pre-fix record's "part (a) properly witnessed" language, and its appeal to `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder`/`TestPaginateGroupPartitionPinnedByValue` as covering part (b), were a CATEGORY ERROR — those two tests use `IsHeader: false` keys and witness nothing about the `IsHeader` branch (Finding 2). Deleting the header's grouping in production left the whole suite green (1000/0), confirmed in this review. Closed by `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer` (`internal/layout/paginate_group_test.go`), a self-contained two-part (grouped/ungrouped) proof scoped to `ItemGroupKey{IsHeader: true}` with members of DIFFERING extent — a shape the shipped single-line header cannot itself produce, which is exactly why the folio-level control stays Class A (a fact about the header's construction, not a defect) even though the mechanism it exercises now has a real witness. See 9.4 CORRECTED and Finding Resolutions, Finding 2. |
| AC6 | **N/A by design** | No mutation is definable; asserts absence of change (per the story's own instruction, matching AC6's stated exception). |

### 11. Deliberately not built

- No repeated header on continuation pages (4.4).
- No footer aggregates, no orphan control, no "keep a total row with a data row" rule (4.5).
- No clip-to-fresh-page / diagnostic for a row taller than the whole window (4.6) — the located
  `OverflowError` is unchanged, confirmed by AC4's pinning test, which says in its own words that it
  records current behaviour and does not endorse it.
- No alternating row shading (4.8).
- No new diagnostic code (D-000.65) — the too-tall-row condition already shipped in 2.6 as
  `OverflowError`.
- No `rowHeight` schema field.
- No collapse of the two-pass pagination design (DECISION-3, §0/9.6 — an assertion was landed instead).
- No `//go:build matrix` e2e test committed (§8 — drafted, build-checked, withdrawn as an unauthorised
  gate obligation).
- No golden fixture recorded from the new multi-page table fixture (D4; 4.7's job).

## File List

- `folio-go/internal/layout/paginate.go` — modified (developer, then FINISHER for Finding 1/Blocker):
  `ColumnItem.Group`/`ItemGroup`/`ItemGroupKey` added; `Paginate` extended with group-extent
  computation and effective-extent substitution in the fit/overflow tests. **Finisher: the R7
  contiguity pre-pass and its render-time "not contiguous" internal error were REMOVED** (they
  rejected legal documents — Finding 1) **and replaced** with a group-page short-circuit: a group's
  page is decided once, at whichever member the column-order sweep visits first, and every
  later-visited member of the same group copies that page directly, with no dependence on
  contiguity. The R6 belt-and-suspenders check is retained (now a structural guarantee rather than a
  contiguity-contingent one). Doc comments revised throughout (the "two items, same extent" passage
  per 4.2's hand-off note, and R7's premise correction per Finding 1).
- `folio-go/internal/layout/paginate_group_test.go` — new (developer), then FINISHER edits: layer-1
  (layout package) hand-derived tests for the grouping mechanism (AC1, AC2, AC4, R7, R2).
  **Finisher: `TestPaginateRejectsANonContiguousGroup` REMOVED** (its subject — the contiguity error —
  no longer exists after Finding 1's fix) **and replaced** with
  `TestPaginateGroupToleratesAnInterveningItemAtTheSameTop` and
  `TestPaginateGroupSurvivesAnInterveningPageAdvance` (Finding 1's regression, red-proofed against the
  pre-fix code). **Added** `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer`
  (Finding 2/Blocker: a failable, `IsHeader`-key-specific mechanism witness). **Fixed**
  `TestPaginateUngroupedItemsAreUnaffectedByGrouping` (Finding 5/Minor: pinned the partition by value
  instead of a bare page count). **Fixed** two `rangeint`-shaped loops (Finding 7/Minor).
- `folio-go/table_render.go` — modified (developer): `tableRectSource.isHeaderRow` field and
  `chromeRowGroup()` method added; header rect construction site sets `isHeaderRow: true`.
  Untouched by the finisher.
- `folio-go/render.go` — modified (developer): `textRunSource.isHeaderLabel` field and
  `lineRowGroup()` method added; header-label run construction site sets `isHeaderLabel = true`;
  `paginateDocument`'s two `ColumnItem`-construction sites (rects, text) now set `Group` via the new
  methods. Untouched by the finisher.
- `folio-go/page_number.go` — modified (developer): `contentColumnItems`'s two `ColumnItem`-
  construction sites (text, rects) now set `Group` via the new methods; the false
  "POSITION-IDENTICAL" doc comment repaired (D3). Untouched by the finisher.
- `folio-go/table_pagination_test.go` — new (developer), then FINISHER edits: folio-package
  integration tests (AC1, AC2, AC3, AC4, AC5) against a real multi-page table document. **Finisher:**
  fixed the `footerPageCount=true` fixture's invalid element id (Finding 3/Major — `"pf1"` did not
  match the schema's id pattern); added `TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter`
  (Finding 3/Major and Major 3's coverage gap: takes the `{{page}} of {{pages}}` branch through
  `Render()`); added `tableBesideSameYElementDoc`/`TestTableBesideSameYElementRenders` (Finding
  1/Blocker's regression, red-proofed against the pre-fix code); extracted `tooTallRowDoc()` and added
  `TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender` (Finding 4/Major: pins the
  `Kind == "line"` arm the shipped `Render()` path actually produces, alongside the existing
  `Kind == "table"` PHASE-B-scoped pin); removed dead `rowPage` bookkeeping in
  `TestTableRowsAppearExactlyOnceInDataOrderAcrossPages` (Finding 8/Nit).
- `folio-go/table_render_test.go` — modified (developer): `TestTableInPageHeaderRepeatsIdenticallyAcrossPages`'s
  doc comment amended with an explicit "must not be fixed" note. Untouched by the finisher.
- `_bmad-output/implementation-artifacts/4-3-break-a-table-across-pages-without-splitting-a-row.md` —
  this file: task checkboxes, Delivery Log, File List, Change Log.
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — status updated to `review`.

## Change Log

| Date | Version | Description | Author |
|---|---|---|---|
| 2026-08-26 | 0.1 | Story created at `903bf8f`; five divergences measured; DECISION-1 taken, DECISION-2 flagged, DECISION-3 surfaced; D-000.79 deletions named per output-producing AC. | bmad-story-creator |
| 2026-08-26 | 1.0 | Implemented DECISION-1's grouping mechanism (`ColumnItem.Group`/`ItemGroup`), carried through both `ColumnItem` builders (AC3), taught to `Paginate` with an R7 contiguity assertion; header row grouped identically to a data row (AC5); AC3's false comment repaired and replaced with an executable two-pass agreement assertion (DECISION-3); all named two-part and single mutations run over the whole suite with part (a)/(b) results recorded, including two measured no-witness findings (AC2(a) duplication, AC2(b)/AC5(b) at the folio level) stated explicitly rather than papered over; a Class-B aggregation defect in the AC3 test found and fixed pre-recording; a `//go:build matrix` e2e test drafted, build-checked and withdrawn as an unauthorised gate obligation. All three gates green (987→1000/0/1, 115/0, 3/0), no golden moved. | bmad-story-developer |
| 2026-08-26 | 1.1 | Finisher pass, addressing code review (2 Blockers, 2 Majors, 3 Minors, 1 Nit). **Blocker 1 fixed**: removed R7's contiguity pre-pass/internal error from `Paginate` (it rejected legal documents — a table beside any same-`y` element) and replaced it with a group-page short-circuit resolved at first visit; red-proofed at both the `layout` and `Render()` levels against the pre-fix code. **Blocker 2 fixed**: landed a failable, `IsHeader`-key-specific mechanism witness (`TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer`) and corrected the record's category-error claim that generic, `IsHeader:false`-keyed tests covered AC5's mechanism. **Major 3 (fixture) fixed**: repaired the `footerPageCount` fixture's invalid element id and added a `Render()`-level test taking that branch, cross-checked against PHASE A and the produced PDF's own page-object count. **Major 4 fixed**: pinned the `Render()`-path's actual `Kind == "line"` alongside the existing PHASE-B-scoped `Kind == "table"` pin, with both arms measured against the same document. **Coverage-hole (Major 3's other half) fixed**: added the first tests in this story to call the public `Render()`. Three Minors and one Nit fixed (value-pinned an R2 guard, swept two remaining `rangeint` loops, removed dead bookkeeping); one Minor dismissed as superseded (its subject, the contiguity-error test, no longer exists after Blocker 1's fix). All three gates re-measured green (1010/0/1, 115/0, 3/0), `gofmt`/`go vet` clean in all three modules, `-tags=matrix` build/vet clean, no golden moved. Status set to `done`. | bmad-story-finisher |

---

## Review Findings

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-26
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 2
- **Majors:** 3
- **Minors:** 3
- **Nits:** 1

**Gate figures measured independently in this review** (tree as the developer left it, digests
verified before and after every mutation):

| Gate | Command | Measured | Story claims |
|---|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **1000 pass · 0 fail · 1 skip** | 1000 / 0 / 1 ✓ |
| `lint` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **115 / 0** | 115 / 0 ✓ |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **3 / 0** | 3 / 0 ✓ |
| `gofmt -l .` | all three modules | **empty** | ✓ |
| `go vet ./...` | all three modules | **clean** | ✓ |
| `go build -tags=matrix ./...` / `go vet -tags=matrix ./...` (CI legs, run natively) | `folio-go` | **clean** | not claimed; confirmed |

Named standing guards confirmed green by name in the run output: `TestGlyphIdentifierCensus`,
`TestSpineStageLadderMatchesStageRankTable`, `TestFloatTypedProductionScan`, `TestMapRangeProductionScan`
(lint); `TestFolioMethodNamesAreInjective`, `TestValidateNeverReachesRenderOrInternalPDF`,
`TestPaginateNeverProducesAnEmptyPage`, `TestEpic2GateObligationsMatchTheDeclaredSet`,
`TestTableInPageHeaderRepeatsIdenticallyAcrossPages` (folio-go). Every golden test
(`TestRenderMatchesGoldenFixture`, `TestRenderMatchesFontTextGoldenFixture`,
`TestRenderMatchesImageEmbedGoldenFixture`, `TestMultiScriptFallbackGoldenFixture`,
`TestFirstBaselineSemanticAcceptanceAcrossEveryReRecordedGolden`) genuinely re-renders and is green —
that, not a hash-of-committed-files check, is the evidence that **no pre-existing golden moved**.

**Mutations re-run by this reviewer** (whole suite, never behind `-run`, `-count=1`; every file
restored by hand and re-verified by SHA-256, never `git checkout`):

| Mutation | Measured here | Story's record | Verdict |
|---|---|---|---|
| **AC1 part (a)** — remove the data-row chrome append (`table_render.go:744`) | **987 / 17**; `TestDataRowNeverSplitAcrossPages` **STILL PASSES** | 983 / 17, still passes | **Reproduced ✓** |
| **AC1 part (b)** on top of (a) — disable `Group` substitution in `Paginate` | **982 / 22**; `TestDataRowNeverSplitAcrossPages` **REDDENS** | reddens | **Reproduced ✓** |
| **AC5 part (a)** — remove the header chrome append (`table_render.go:533`) | **988 / 17**; `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` **STILL PASSES** | still passes | Reproduced, but see **Finding 2** — the pass is unfailable |
| **AC3** — leave PHASE A grouping-blind | **999 / 1**; only `TestBothPaginationPassesAgreeOnRowPartition` reddens, on the *right* assertion (`row 5: PHASE A itself spreads this row's lines across 2 distinct pages`) | 999 / 1 | **Reproduced ✓, and the guard has real teeth** |
| **NEW (this review)** — delete the header grouping entirely (`chromeRowGroup`/`lineRowGroup` return `ItemGroup{}` for the header) | **1000 / 0 — the whole suite stays GREEN** | not run | **Finding 2** |

Fixture facts measured here that the record does not state: the AC1 fixture's wrapped row (row 5) is
the *first row on page 1*, i.e. it genuinely straddles the boundary — that is why AC1's part (a) has
teeth and is not vacuous. **AC1's two-part instrument is sound and AC1 is properly met.**

---

### Finding 1: A table beside any element sharing its `y` now fails to render — R7's contiguity check turns a legal, previously-rendering document into a hard internal error

- **Severity**: **Blocker**
- **Category**: Correctness
- **Location**: `folio-go/internal/layout/paginate.go:385-455` (the grouping pre-pass and its `closedKeys` contiguity error)
- **Related AC**: AC6 / R2 (behaviour preserved for everything that is not a table row), R7, AD-14
- **Observation**: The R7 contiguity assertion fires on ordinary author-written layouts, not only on a
  caller bug. Measured at HEAD against a `903bf8f` worktree with the identical probe file:

  | Document (content band) | `903bf8f` | HEAD |
  |---|---|---|
  | table @ y=0, text @ y=0 (declared either order) | renders | **internal error** |
  | table @ y=0, text @ y=10 (= header height, ties data row 0's top) | renders | **internal error** (`IsHeader:false Index:0`) |
  | table @ y=20, text @ y=20 | renders | **internal error** |
  | two tables both @ y=0 | renders | **internal error** |
  | table @ y=0, text @ y=20 or y=30 | renders | renders |

  Horizontal overlap is **not** required. A table occupying x 0..120 with a note at x 130..175 — no
  geometric overlap at all, only the same `y` — returns:

  ```
  folio: Render: layout: Paginate: internal error: group {ElementID:e1 IsHeader:true Index:0} is not
  contiguous in column order — element "e1"'s item at column position 2 reopens a group already closed
  earlier in the sweep. ...  a caller has placed something else ... between two of this group's own members
  ```

  The same document returns `<nil>` at `903bf8f`. The mechanism: `Paginate`'s sort is by `Top` only, so
  any item tying a group member's `Top` and appended between two of that group's members lands between
  them in `order`. PHASE A appends text before rects, so a text element at the table's own `y` sorts
  between the header's label item and the header's chrome item.
- **Impact**: A caption, a column heading, or a note placed level with a table — a canonical report
  layout, and one the golden report of Epic 4 is likely to contain — becomes an unrenderable document
  emitting an *internal error* that blames "a caller" for something the template author did. No test in
  the repository covers it because every fixture in `table_pagination_test.go` places exactly one
  element at any given `y`. This is the first commit CI will run and the first push this repository has
  had.
- **Suggested Resolution**: The premise R7 states — *"a group's members are contiguous in column order
  by construction"* — is **false for real templates**, and the correct response to a false premise is to
  remove the dependence on it, not to assert it and fail. The group-extent pre-pass already computes each
  group's union extent across the whole `order` slice **without needing contiguity**; what needs
  contiguity is only the assumption that no intervening item advances `page` between two members. Resolve
  the group as a unit at its first-visited member (assign every member's page there, and skip members
  already assigned), which makes contiguity unnecessary and preserves the union-extent choice. If
  contiguity is kept as a real constraint, it must not be enforced as a render-time internal error on
  input the schema accepts. Add a fixture with two content elements sharing one `y` to whichever
  resolution is taken. Do **not** resolve this by narrowing the fixture set so the case stays unmeasured.

---

### Finding 2: AC5's mechanism is entirely inert — deleting the header grouping leaves the whole suite green, and the record states the opposite

- **Severity**: **Blocker**
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/table_render.go:71-86` (`chromeRowGroup`, the `isHeaderRow` branch); `folio-go/render.go:177-196` (`lineRowGroup`, the `isHeaderLabel` branch); `folio-go/table_pagination_test.go:626-705`; Delivery Log §6, §9.4, §10
- **Related AC**: AC5
- **Observation**: Mutation run in this review over the whole suite: make both
  `tableRectSource.chromeRowGroup()` and `textRunSource.lineRowGroup()` return `layout.ItemGroup{}` for
  the header cases — i.e. **delete AC5's mechanism outright**. Result: **1000 pass / 0 fail. Nothing
  reddens.** By the story's own D-000.79 taxonomy that is **Class A**: a feature that dies to deletion
  with no witness at all.

  The cause is measurable and total. A probe of the shipped item list for the story's own
  `headerPushedToNextPageDoc()` fixture shows:

  ```
  HEADER CHROME item: Top=115000 Bottom=125000
  HEADER LABEL  item: Top=115000 Bottom=125000 runs=2
  number of ColumnItems carrying header LABELS = 1
  ```

  Every column label of a table header carries `lineIndex = 0` (`table_render.go:515`), so the
  `(elementID, lineIndex)` merge in `paginateDocument` / `contentColumnItems` / `itemsForTest` collapses
  **all** header labels into **one** `ColumnItem`. Therefore in
  `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated`:
  - the CORE assertion `len(headerLabelPages) != 1` is **structurally unfailable** — a single
    `ColumnItem` lands on exactly one page under any implementation whatsoever;
  - the SECONDARY chrome-vs-label assertion is unfailable for the separate reason §9.4 correctly
    identifies (the two extents are *exactly* equal by Story 4.1's construction);
  - so under mutation (a) **every** assertion about header atomicity is either unfailable or skipped,
    and the test's "STILL PASSES" is guaranteed a priori.

  Confirmed directly: under **(a) + (b)** — header chrome removed *and* grouping disabled — this test
  still **PASSES** (measured, 982/22 run). A control that cannot fail in either direction is not a
  control.
- **Impact**: The Delivery Log §10 records *"AC5 — Part (a) properly witnessed; part (b) has no
  folio-level witness"*. The first clause is false and the split understates the gap: **AC5's first half
  has no failable witness at any level.** §9.4's appeal to
  `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder` / `TestPaginateGroupPartitionPinnedByValue` is
  a category error — those use `ItemGroupKey{..., Index: n}` with `IsHeader` false, so they witness the
  *generic* substitution but say nothing about whether the header is grouped. Under D-000.66 this is a
  knowingly-unmet requirement being carried forward as a fact, and the next story that touches header
  layout (4.4, immediately) will read a green suite as confirmation that header atomicity is mechanised
  when nothing checks it.
- **Suggested Resolution**: Either (i) land a witness — a `layout`-level test using
  `ItemGroupKey{IsHeader: true}` whose header chrome and header label have **different** extents, and
  which reddens when the header's `Group` is cleared, which would at minimum make the `IsHeader` key
  branch load-bearing; or (ii) if the shipped header genuinely cannot produce differing extents, say so
  and record AC5 explicitly as **Class A — mechanism built, no witness exists at any level**, deleting
  the "part (a) properly witnessed" claim from §6, §9.4 and §10. What must not stand is the current
  record, which reports a control that cannot fail as evidence that it can.

---

### Finding 3: The `{{page}} of {{pages}}` fixture branch is dead *and* broken — the printed `Y` for a multi-page table is never exercised, which is exactly what D-4.3.2's sequencing constraint was about

- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_pagination_test.go:24-56` (`multiRowTableDoc(footerPageCount bool)`), call sites at `:173`, `:282`, `:362`
- **Related AC**: AC3
- **Observation**: `footerPageCount` is passed **`false` at all three call sites**; the `true` branch is
  never taken. Worse, that branch does not even parse — its footer element id `"pf1"` violates the
  schema's own id pattern:

  ```
  ParseTemplate: template: field bands.pageFooter.elements[].id (element pf1):
  template: id "pf1" does not match ^e[0-9a-z]+$ (value: pf1)
  ```

  and once that is corrected the document still fails on `nextId: 5` being below the highest id present.
  Its doc comment claims *"AC3's own witness needs both passes to see the SAME table"* — the witness does
  not use it. Relatedly, **no test added by this story calls `Render()`**: AC1/AC2/AC3/AC5 all stop at
  `layout.Paginate` via `paginateContentTableForTest`, and AC4 likewise. That is also why Finding 1's
  regression was not caught.
- **Impact**: D-4.3.2 escalated the printed `Y` from untidiness to *"user-visible incorrect output"* and
  made the agreement guard a precondition for the first multi-page table golden. The fixture written to
  exercise that path is unusable, so the end-to-end claim rests entirely on the item-level AC3 test.
- **Suggested Resolution**: Fix the fixture (valid element id, `nextId` above the highest id) and take
  the `true` branch in at least one test that goes through `Render()`, asserting the emitted document's
  page count against PHASE A's. **Verified independently in this review, so this is a coverage gap and
  not a live defect**: a sweep of 240 configurations (rows 1..60 × wrap ∈ {none, first, middle, last})
  found **zero** disagreements between PHASE A (`contentColumnItems`, the source of the printed `Y`) and
  PHASE B, and a `Render()` of the fixture at 20/40/60 rows produced 3/5/7 content pages matching 3/5/7
  `/Type /Page` objects in the emitted PDF. **The sequencing constraint is honoured and the printed `Y`
  is correct** — no golden was recorded from the new fixture, and the AC3 guard and comment repair both
  landed green.

---

### Finding 4: AC4's pinning test asserts `Kind == "table"` against an item order the shipped path never uses — `Render()` reports `Kind == "line"`

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/table_pagination_test.go:503-586` (`TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow`, assertion at `:580`); `folio-go/table_pagination_test.go:100-126` (`itemsForTest`)
- **Related AC**: AC4
- **Observation**: The test paginates via `itemsForTest`, which appends **rects before text** (PHASE B's
  order), so the chrome rect is visited first and `OverflowError.Kind` is derived as `"table"`. The
  shipped path reaches `Paginate` through **PHASE A** first (`render.go:1487`,
  `contentColumnItems`), which appends **text before rects**; at the default zero top padding the first
  line item ties the chrome on `Top` and is visited first, so `Kind` is `"line"`. Rendering the test's own
  too-tall document through `Render()` gives:

  ```
  folio: Render: element e1: line is taller than the content window (272400mp against a content height
  of 110000mp), so it fits on no page. ...
  ```

  **This is not a regression** — the identical probe at a `903bf8f` worktree produces the identical
  message, so AC4's *"exactly as at `903bf8f`"* is behaviourally satisfied and the story is correct to
  claim it.
- **Impact**: The test is named and documented as recording **current behaviour** for Story 4.6 to
  inherit, and 4.6 is explicitly directed to it. It records a `Kind` that the product does not emit on the
  path a user reaches. 4.6 will build its clip-and-diagnose behaviour against a pin that is wrong about
  which arm fires.
- **Suggested Resolution**: Assert through the real PHASE A ordering (or through `Render()`), and pin
  what actually ships — `Kind == "line"` for a too-tall table row at default padding — with a note that
  the mislabelling predates this story and belongs to 4.6. Alternatively keep the `"table"` assertion but
  scope it explicitly to PHASE B's ordering and add the PHASE A case beside it. Do not leave one arm
  pinned and the other unmeasured.

---

### Finding 5: `TestPaginateUngroupedItemsAreUnaffectedByGrouping` proves R2 with a page count — a proxy, not a guard

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/layout/paginate_group_test.go:289-302`
- **Related AC**: AC6 / R2
- **Observation**: The test's stated claim is that an ungrouped item *"must place EXACTLY as it did
  before the `Group` field existed"*. Its only assertion is `len(plan.Pages) != 3`. A wrong state
  producing the same statistic is easy to name: any partition that moves an item between pages while
  leaving three pages — e.g. items 6 and 7 swapping pages — passes unchanged.
- **Impact**: R2 is the invariant that keeps every non-table golden byte-identical. Its dedicated guard
  cannot see the failure it exists to catch. The golden suite does carry the real evidence, so the
  exposure is low — but the test advertises a stronger claim than it makes.
- **Suggested Resolution**: Pin the partition by value the way the sibling
  `TestPaginateGroupPartitionPinnedByValue` does — assert which `TextRunRef` lands on which page — rather
  than the page count.

---

### Finding 6: `TestPaginateRejectsANonContiguousGroup` asserts only that *an* error occurred

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/layout/paginate_group_test.go:235-248`
- **Related AC**: R7, AD-14
- **Observation**: The test asserts `err == nil` and nothing more. It does not assert the error is the
  contiguity error rather than `MixedItemError` or an `*OverflowError`, and it does not assert the error
  is **located** — AD-14's actual requirement, and the property the implementation's own comment claims
  (*"a located internal error, never a silent mis-placement"*). Any future change that makes this fixture
  fail for an unrelated reason keeps the test green.
- **Impact**: The one guard for R7's enforcement cannot distinguish the failure it names from any other
  failure, and does not check the "located" half of the claim at all.
- **Suggested Resolution**: Assert the message names the element and the column position (or introduce a
  typed error), so the test discriminates this failure from every other way `Paginate` can error.
  Note this test's subject is itself under review — see **Finding 1**.

---

### Finding 7: The `rangeint` cleanup was partial — the sibling new file was not swept

- **Severity**: Minor
- **Category**: Convention
- **Location**: `folio-go/internal/layout/paginate_group_test.go:37` (`for i := 0; i < n; i++`), `:138` (`for r := 0; r < 6; r++`); `folio-go/table_pagination_test.go:109`, `:418`
- **Observation**: Delivery Log §0 records three `rangeint` diagnostics fixed in
  `table_pagination_test.go`. The other file added by the same story,
  `paginate_group_test.go`, still carries two loops of exactly the shape that was rewritten
  (`for i := 0; i < n; i++`, `for r := 0; r < 6; r++`). Confirmed in this review that the rewrites did
  **not** perturb any pagination assertion — the loop bounds are marker/row counters, the fixture
  arithmetic is untouched, and the suite is bit-for-bit the same 1000/0/1 — so §9.7's claim is sound as
  far as it goes.
- **Impact**: Cosmetic only. Confirmed in this review that CI runs `go build`, `go vet` and `gofmt -l`
  and **no** staticcheck/golangci-lint step, so neither `rangeint` nor `S1039` can redden CI. The
  inconsistency will simply reappear as IDE noise on the next story.
- **Suggested Resolution**: Sweep both new files or neither, and say which in the Delivery Log.

---

### Finding 8: Dead bookkeeping in the AC2 test

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/table_pagination_test.go:287`, `:303-305`, `:318-320`
- **Observation**: `rowPage` is allocated, initialised to `-1`, written inside the page loop, and then
  read only by `if rowPage[row] == -1 { continue }` as the **last statement of the loop body**, where
  `continue` is a no-op. The variable has no effect on any assertion.
- **Impact**: None functionally; it reads as a check that is not one.
- **Suggested Resolution**: Delete `rowPage` and the vestigial `continue`, or make the `-1` case an
  actual assertion.

---

### AC-by-AC verdict

| AC | Verdict | Basis |
|---|---|---|
| **AC1** | **Satisfied** | Two-part instrument reproduced independently: (a) passes at 987/17, (b) on top of (a) reddens at 982/22 on `TestDataRowNeverSplitAcrossPages`. Confirmed non-vacuous — the wrapped row (5) is the first row on page 1, so it genuinely straddles the boundary. Set equality is keyed on identity, not a count. |
| **AC2** | **Satisfied** | Partition pinned by value at the `layout` level; omission and duplication reported as separate failures; order witnessed by `TestPaginateGroupEmissionStaysInAuthoredOrder`, which reddens under the AC2(b) mutation. §9.1's "no discriminating deletion for duplication" is **correct**: `Paginate` writes `pageOf[idx]` once per authored index and emits once, so there is no existing mechanism to delete. Honest label, correctly applied. |
| **AC3** | **Satisfied** | Mutation reproduced at 999/1, reddening on the right assertion (PHASE A splitting row 5), which is precisely D-4.3.2's hazard. Independently verified across 240 configurations that the two passes agree, and that `Render()`'s emitted page count matches. False `POSITION-IDENTICAL` comment properly repaired. Coverage gap only — **Finding 3**. |
| **AC4** | **Behaviourally satisfied, mis-pinned** | Behaviour is byte-for-byte what `903bf8f` produced (verified against a baseline worktree). The pinning test says so in its own words and names 4.6. But it pins the wrong arm — **Finding 4**. |
| **AC5** | **NOT met as recorded** | The mechanism is inert: deleting it leaves the suite at 1000/0. **Finding 2**. |
| **AC6** | **Satisfied for non-table documents; violated for table documents** | No golden moved (every golden test genuinely re-renders and is green); `internal/` free of the banned imports and of binary float, with `lint`'s `TestFloatTypedProductionScan` and `TestMapRangeProductionScan` green as the instruments; all named standing guards green. But R2's "behaves exactly as today" is broken for a whole class of legal documents — **Finding 1**. |

### Things checked and found sound (recorded so they are not re-litigated)

- **`MixedItemError` untouched** — verified by diff; the exclusivity check at `paginate.go:290-311` is
  unmodified and `Group` is genuinely orthogonal to it. R1 honoured.
- **No map iteration reaches output order** (R5) — `groups`, `closedKeys` and `groupPage` are indexed
  only, never ranged; `lint`'s `TestMapRangeProductionScan` / `TestMapRangeUnderModule` green.
- **Integer millipoint arithmetic only** (R4) — the grouping code performs `<` / `>` comparison and
  min/max assignment only; no division, no rounding, no float. `TestFloatTypedProductionScan` green.
- **Cannot panic** (AD-14) — every grouped item is entered into `groups` in the pre-pass before the
  sweep reads it, so `groups[it.Group.Key]` cannot nil-deref; both failure paths return a located error.
  What it *can* do is fail a legal document — Finding 1.
- **The withdrawn `//go:build matrix` e2e test** — `TestEpic2GateObligationsMatchTheDeclaredSet` is a
  real guard that enumerates matrix-tagged files against `declaredEpic2GateObligations`
  (`byte_neutrality_test.go:532-710`) and errors on an undeclared one, so it fired legitimately. Nothing
  partial was left behind: `git status` shows no matrix-tagged file, and `go build -tags=matrix ./...`
  and `go vet -tags=matrix ./...` — both CI legs — are clean when run natively here.
- **`TestTableInPageHeaderRepeatsIdenticallyAcrossPages`** — green, and its doc comment now carries the
  explicit "must not be fixed" note. Correct.
- **Story 4.2's last-write-wins aggregation bug (§9.3)** — the fix is real and correct: the AC3 test now
  tracks a *set* of pages per row per phase and asserts `len(set) == 1` before comparing across phases.
  Searched for siblings of that defect: `TestDataRowNeverSplitAcrossPages` uses a per-row page **set**
  (`linePages`) — sound. `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` uses page **sets** for both
  labels and chrome — sound in shape, but unfailable for the different reason in Finding 2.
  `TestTableRowsAppearExactlyOnceInDataOrderAcrossPages` uses `seenCount` per row plus a union set —
  sound. `chromePage[r.rowIndex] = p` at `table_pagination_test.go:248` **is** a last-write-wins scalar,
  but it is safe here: it is compared against a page set that has already been asserted to be a
  singleton. No live sibling found.
- **Integration/e2e cadence (D-000.4)** — read rather than run, as required. `table_pagination_test.go`
  is integration-shaped but light and correctly run in-gate; the heavy `Render()`-to-bytes class was
  drafted and withdrawn. The story says so explicitly. Correct.
- **Scope fences** — no repeated header, no footer aggregates or orphan control, no clip-to-fresh-page,
  no alternating shading, no new diagnostic code, no `rowHeight` schema field, no collapse of the
  two-pass design. All confirmed absent from the diff.
- **The grouping's shape versus 4.5 / 4.6** — `ItemGroupKey{ElementID, IsHeader, Index}` groups exactly
  one row and never a run of rows, so 4.5's keep-together rule is a genuinely different mechanism that
  can be added beside it rather than by tearing this out; and 4.6 changes only the `effectiveBottom -
  effectiveTop > height` arm, which is already the single place a group's total height is tested. **The
  minimum-mechanism judgement is sound and neither successor is boxed in.** Finding 1's resolution should
  preserve that property.

---

## Finding Resolutions (finisher)

**Note on the review's own tally.** The Review Summary states "Majors: 3", but the findings section
below it contains exactly two `Severity: Major` entries (Finding 3, Finding 4). Recounted directly
against the eight `### Finding` blocks: 2 Blocker, 2 Major, 3 Minor, 1 Nit = 8 total. Not corrected in
the reviewer's own section above (out of scope to edit another agent's write-up after the fact); noted
here so the discrepancy is not silently propagated.

| # | Severity | Decision | Rationale |
|---|---|---|---|
| 1 | Blocker | **FIX** | Reproduced exactly as reported (a table beside any same-`y` element became an unrenderable "internal error"). R7's contiguity pre-pass and its fatal error were REMOVED from `internal/layout/paginate.go`; replaced with a group-page short-circuit that resolves a group's page once, at its first-visited member, and copies that page to every later member with no ordering requirement. Red-proofed at both the `layout` level (`TestPaginateGroupToleratesAnInterveningItemAtTheSameTop`, `TestPaginateGroupSurvivesAnInterveningPageAdvance`) and the `Render()` level (`TestTableBesideSameYElementRenders`): all three measured to FAIL against the pre-fix code (the exact "is not contiguous" error) and PASS after. This is a mandatory fix per the orchestrating instruction ("must not ship") and is not a judgment call. |
| 2 | Blocker | **FIX** | Confirmed AC5's folio-level control is unfailable: deleting the header's `chromeRowGroup`/`lineRowGroup` `IsHeader` branches leaves the whole suite green, exactly as measured in the review. The cause (Story 4.1's header is unconditionally one physical line, so chrome and every label share an identical extent) makes the folio-level test's greenness a genuine, honest fact, not a bug — but the record's appeal to `IsHeader:false`-keyed tests as covering it was a category error, corrected in §9.4/§10. Closed per the finding's option (i): a new mechanism-level witness, `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer`, builds an `ItemGroupKey{IsHeader:true}` group with differing member extents (a shape the shipped header cannot itself produce) and self-proves, in a grouped/ungrouped pair, that the `IsHeader` branch is load-bearing. |
| 3 | Major | **FIX** | The `footerPageCount=true` fixture's element id (`"pf1"`) violated the schema's `^e[0-9a-z]+$` pattern and never parsed, confirmed. Fixed to `"e4"` (nextId 5 already exceeds it). Added `TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter`, which takes the previously-dead branch through the public `Render()` at 20/40/60 rows and cross-checks the produced PDF's `/Type /Page` count and `/Count` against PHASE A's own page count — closing the coverage hole the finding names as the reason Finding 1 escaped. The 240-configuration sweep and 3/5/7-page measurements the reviewer reported were re-derived here independently (not merely trusted) and matched. |
| 4 | Major | **FIX** | Confirmed: `Render()` produces `Kind == "line"` (PHASE A's text-before-rects order), not `"table"`, and this is unchanged from a `903bf8f`-equivalent baseline per the review's own comparison — not a regression. Rather than silently repin the existing test (which would leave the PHASE B arm unmeasured, exactly what the finding warns against), both arms are now pinned against the SAME document (`tooTallRowDoc`, extracted for reuse): the existing test is rescoped in its own doc comment to PHASE B (`itemsForTest`, `Kind == "table"`), and a new `TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender` pins the real `Render()` path (`Kind == "line"`), explicitly noting the mislabelling predates this story and belongs to 4.6. |
| 5 | Minor | **FIX** | `TestPaginateUngroupedItemsAreUnaffectedByGrouping` now pins the exact per-line page partition and per-page window shift by value (mirroring `TestPaginatePinsTheLineToPagePartitionByValue`'s own fixture and discipline), rather than asserting only `len(plan.Pages) != 3` — closing the swap-blind conservation-law gap the finding names. |
| 6 | Minor | **DISMISSED (superseded, not overridden)** | This finding's subject, `TestPaginateRejectsANonContiguousGroup`, no longer exists: Finding 1's fix removes the contiguity error the test was exercising, so there is no longer an error for it to under-discriminate. Its replacements (`TestPaginateGroupToleratesAnInterveningItemAtTheSameTop`, `TestPaginateGroupSurvivesAnInterveningPageAdvance`) assert exact page/rect/run placement rather than "an error occurred", which is a strictly stronger specificity than Finding 6 asked for. Not a case of dismissing a legitimate critique — the critique's target was deleted for an unrelated, mandatory reason (Finding 1). |
| 7 | Minor | **FIX** | The two `for i := 0; i < n; i++` / `for r := 0; r < 6; r++` loops in `paginate_group_test.go` (marker/row counters only, confirmed not touching geometry arithmetic) rewritten as `for i := range n` / `for r := range 6`, matching the sweep already applied to `table_pagination_test.go`. Both files now consistent. |
| 8 | Nit | **FIX** | Removed the dead `rowPage` slice and its no-op `continue` from `TestTableRowsAppearExactlyOnceInDataOrderAcrossPages`; the omission/duplication checks it fed nothing into are unchanged. |

**Validation after all fixes** (see the story's own top-level baseline table for the gate commands; run
identically from `folio-go/`, `lint/`, `hashmatrix/`):

| Gate | Result |
|---|---|
| `folio-go` (`go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./...`, counting `--- PASS`/`--- FAIL`/`--- SKIP` lines, matching the story's own convention) | **1010 pass · 0 fail · 1 skip**, measured directly. Reconciled: 1000 baseline + `TestPaginateGroupToleratesAnInterveningItemAtTheSameTop` (+1) + `TestPaginateGroupSurvivesAnInterveningPageAdvance` (+1) + `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer` (+1 parent, +2 subtests) + `TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter` (+1 parent, +3 subtests) + `TestTableBesideSameYElementRenders` (+1) + `TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender` (+1) − `TestPaginateRejectsANonContiguousGroup` removed (−1) = 1000 + 10 = **1010**. |
| `lint` | **115 / 0**, unchanged |
| `hashmatrix` | **3 / 0**, unchanged |
| `gofmt -l .` | empty, all three modules |
| `go vet ./...` | clean, all three modules |
| `go build -tags=matrix ./...` | clean |
| `go vet -tags=matrix ./...` | clean |
| No golden fixture moved | confirmed — every golden test genuinely re-renders and is green |

**Red-proofs run in this finisher pass** (whole suite, `-count=1`, never behind `-run`; tree restored
via `cp` from a pre-mutation backup between runs, never `git checkout`):

| Mutation | Result | Named tests |
|---|---|---|
| Revert `paginate.go` to the pre-fix (R7 contiguity error) implementation | `TestPaginateGroupToleratesAnInterveningItemAtTheSameTop` and `TestPaginateGroupSurvivesAnInterveningPageAdvance` **FAIL** — the exact "is not contiguous" internal error | Confirms Finding 1's regression at the `layout` level |
| Same revert, `TestTableBesideSameYElementRenders` re-run | **FAILS** with the identical "internal error: group ... is not contiguous" message reported in Finding 1, at the `Render()` level | Confirms Finding 1's regression end to end, not only at the unit level |
| Restore the fix | All of the above **PASS**; full suite green | — |

Both red-proof halves (pre-fix FAIL, post-fix PASS) were run for every new regression test named above,
per D-000.75 (the party landing a safety property owes the mutation) and D-000.80 (part (a)/(b)
discipline, applied here as pre-fix/post-fix since Finding 1's fix is itself the mechanism under
proof, not an accidental-cause removal).
