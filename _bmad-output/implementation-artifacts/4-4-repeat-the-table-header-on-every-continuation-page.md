# Story 4.4: Repeat the table header on every continuation page

Status: done
Epic: 4 — A Go developer can render the golden report (**the C4 gate**)
Covers: **FR26**
Created: 2026-08-26, at HEAD `ec15d36` (Story 4.3, finisher), pushed to `origin/main`

---

## Baseline, measured in this run at creation (HEAD `ec15d36`)

Every figure below was taken in this run, not carried from the brief. Counting convention matches
Stories 4.1–4.3: `--- PASS` / `--- FAIL` / `--- SKIP` lines **anywhere** in `-v` output, so subtests
are counted. (An anchored `^--- PASS` count gives 660/77/3 for the same runs — the same tree, a
different scope. Report the convention with the number.)

| Gate | Command (run from that module's directory) | Measured |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **1010 pass · 0 fail · 1 skip** |
| `lint` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **115 pass · 0 fail** |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **3 pass · 0 fail** |

The single skip is `TestCorpusMeetsP6ExerciseFloors`, **red by design** (D-000.17 / D-2.1.14 /
D-000.57 / D-000.74). It is never to be "fixed".

**Tree note, recorded rather than silently absorbed:** at creation the working tree carried one
uncommitted modification — a 50-line append to `folio-mvp-decision-log.md` opening the **D-000.79
ledger** (the orchestrator's Class A / Class B record, which D-000.79 requires by name). It is not
this story's work and was left exactly as found. Story 4.4's row in that ledger is `pending` and this
story is the second half of D-000.79's falsification test.

### Pre-fix measurement of the defect, and of the trap under it

Probe run at `ec15d36` against the existing `multiRowTableDoc(false)` fixture with 20 rows
(`table_pagination_test.go`), through the real `collectBandTableRuns` → `layout.Paginate` path; probe
file deleted afterwards, tree verified clean.

```
contentTop = 10,000mp    contentHeight = 110,000mp
header:  top 10,000  bottom 20,000   height 10,000mp   (2 cells: x=0 w=80,000 · x=80,000 w=80,000)
row 0:   top 20,000  bottom 30,896   height 10,896mp   (2 cells: x=0 w=80,000 · x=80,000 w=80,000)

pages = 3
page 0  shift=0        headerRects=1  headerLabels=2  dataRows=9   (rows 0..8)
page 1  shift=108,064  headerRects=0  headerLabels=0  dataRows=10  (rows 9..18)
page 2  shift=217,024  headerRects=0  headerLabels=0  dataRows=1   (row 19)
```

**Two facts, and the second is the one that decides how the ACs must be written.**

1. **The header is not repeated today.** Pages 1 and 2 carry zero header chrome and zero header
   labels. FR26 is genuinely unmet, so AC1 is not accidentally true and does not need a part (a) for
   *that* reason.
2. **A look-alike for the repeated header is ALREADY on every continuation page.** Page 1's first two
   rects are at `x=0 w=80,000` and `x=80,000 w=80,000` — **exactly the header's own column geometry** —
   shifted to `y = 10,000mp`, **exactly the content band's top**. They are a *data row's* chrome. And
   in an unstyled table every one of those rects carries `HasFill=false, HasStroke=false`, so "a rect
   exists here" distinguishes nothing at all.

   > **An assertion of the form "the top of a continuation page carries cells at the table's column
   > positions" PASSES AT HEAD, with the header deleted, with FR26 unimplemented.** That is D-000.79's
   > Class A trap in D-000.80's most dangerous form: not an absent guard, but a guard that appears to
   > work and is watching something else. AC1 therefore carries a **part (a)**, and every header
   > assertion in this story is anchored on the header's **column-label text** and its **declared
   > `headerHeight`**, never on cell geometry alone.

**The partition, not the count, is what moves.** Hand-derived from the figures above, for the same
20-row fixture once the reservation exists: page 0 still holds rows 0..8 (the first page is
unchanged); page 1's usable space drops from 110,000mp to 100,000mp, so it holds rows 9..17, not
9..18; page 2 holds rows 18..19. **Three pages before, three pages after — the page COUNT is
identical and the partition is 9/10/1 → 9/9/2.** A page-count assertion here is exactly the lossy set
D-000.68 forbids. Re-derive this arithmetic during development rather than copying it; it is stated
here to show that a count cannot carry AC2.

---

## In plain terms (read this first if you just want the gist)

A fifty-page bank statement is one long table sliced across pages. Before this story, the column
captions — Date, Description, Amount — were printed once, at the very start, and never again. A
reader who turned to page thirty-four saw a wall of numbers with nothing telling them which column
was which, and had to page all the way back to the beginning to read their own statement. That is
fixed now: wherever a table carries on onto a further page, its caption row is drawn again at the top
of that table's rows on that page, and the space that caption needs is correctly taken out of what
the rows on that page can use — nothing is painted on top of a transaction, nothing slides off the
bottom, and the reported page count still matches reality.

An independent check of the finished work found the underlying arithmetic was right from the start —
the numbers governing where the caption sits and how much room it takes were verified by hand and
hold up. What the check found weaker was the proof of that arithmetic: several of the automated checks
meant to guarantee the caption's position and the surrounding elements' safety were, on closer
inspection, testing bookkeeping rather than the thing itself, so a mistake in exactly where the
caption gets drawn, or a mistake that let it nudge an unrelated piece of the page out of place, could
have slipped through unnoticed. Those checks have since been rebuilt so they genuinely watch the
thing they claim to guarantee, and each one was confirmed to actually catch the mistake it is meant to
catch. Two "coming soon, will run later" stress tests that were left as empty placeholders — one
pushing a five-hundred-row table through, one combining two tables on a page with a page-count
footer — have also been written out properly and confirmed to pass; they still don't run as part of
every ordinary check (that is intentional, they are expensive), but they are no longer hollow. A
smaller reporting bug was also fixed: when a table's caption could not fit because of something else
crowding the same page, the warning shown to a template author understated how much room was actually
available, which could have sent someone chasing the wrong fix.

There is one deliberately confusing neighbour to keep clear of. A table can also be placed in the
running page-header strip at the very top of every sheet; that kind of table is copied onto every
page whole and never splits. That is a different feature with a very similar name, it already works,
and it was left alone.

What this story does not do: it does not let an author switch the repetition off, it does not keep
the caption glued to at least one row when the caption lands right at the foot of a page, and it does
not touch totals rows or over-tall rows — those belong to later stories. A known, deliberately
unfixed rough edge: when a table's own rows get pushed down to make room for its repeated caption,
an unrelated element positioned directly beneath that table on the same page can end up overlapped by
the table's last row. That is a real, if narrow, gap — flagged for whoever builds the fully worked
example report to decide whether it ever actually happens, rather than guessed at and patched here.
Done looks like this: on every page a table continues onto, the captions are there, nothing overlaps,
nothing falls off the bottom, the reported page count still matches reality, every previously recorded
reference document still comes out byte-for-byte identical, and pagination always finishes rather than
looping.

---

## Story

**As a reader of a 50-page statement,**
**I want to know what each column means on page 34,**
**So that I do not have to page back to the start to read my own statement.**

---

## What is actually in the tree, and what 4.4 must move

`collectBandTableRuns` (`folio-go/table_render.go`) emits, for one content-band table, **exactly one**
header `tableRectSource` (`isHeaderRow: true`) at the table's declared `y`, plus its column-label runs
(`isHeaderLabel: true`, every one at `lineIndex 0`), then one `tableRectSource` per data row and one
run group per physical line. `paginateDocument` (`render.go`) and `contentColumnItems`
(`page_number.go`) each turn those into `layout.ColumnItem`s; `layout.Paginate`
(`internal/layout/paginate.go`) slides one window down one immutable column and returns, per page, a
set of refs and **one** `Shift`.

Three consequences bind this story:

1. **The header is one item in the column, so it lands on exactly one page.** Repeating it cannot be
   done by editing the column — the column is never mutated (`paginate.go`'s package doc, AD-24).
2. **`Pagination` expresses one shift per page.** A repeated header is drawn at a position that is
   *not* the header's column position minus that page's shift. Expressing the repeat therefore needs
   a channel the current type does not have. **What that channel is, is the developer's call**
   (D-000.79 §2 — a mechanism named from outside the code over-constrains the design).
3. **There are two pagination passes and they must agree.** PHASE A (`contentColumnItems`) exists only
   to learn the page count before `{{page}}/{{pages}}` are collected (D-2.7.2); PHASE B
   (`paginateDocument`) produces the real pages. D-4.3.2 ruled that two passes silently disagreeing is
   the defect. A reservation taught to one pass and not the other makes `Page X of Y` print a wrong
   `Y`. This is the single most likely way 4.4 ships broken.

### One existing test this story OWNS inverting

`TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` (`folio-go/table_pagination_test.go:677`) ends
with a live fence:

> *"the header must appear on EXACTLY ONE page — repeating it is Story 4.4's job, not this story's"*

That fence is 4.3's correct scoping and **4.4 is the story authorised to invert it**. **Rescope it;
do not delete it.** Its first two halves — every header *label* lands on one page, and the header's
chrome lands on the same page as its labels — are 4.3's row-atomicity guarantee for the table's **own**
header and must stay green and stay meaningful. Only the "appears on exactly one page" clause changes,
into this story's rule. Renaming the test is appropriate; silently weakening it is not.

### One existing test that must NOT be touched, and is easy to confuse with this story

`TestTableInPageHeaderRepeatsIdenticallyAcrossPages` (`folio-go/table_render_test.go:858`). A table
declared in the **`pageHeader` band** is repeated verbatim on every page and never paginates
(`layout.BandContent`). That is a different feature with a nearly identical name. **It must stay
green and it is not a bug to be "fixed" into this story's mechanism.**

---

## R — design constraints this story inherits (outcomes, not mechanisms)

- **R1 — `MixedItemError` is untouched and is not widened.** Exactly one of `{Runs, Images, Rects}` per
  item stays the rule. A repeated header is a header's **chrome** and its **labels**, which is two
  kinds; it must not be collapsed into one item. A multi-run item is legal and normal — collapsing a
  row into a single item is forbidden.
- **R2 — the content column is never mutated.** No gap is inserted into the column, no sibling's
  declared `y` changes. AD-24: *"siblings never move, because nothing in a band ever reflows."*
- **R3 — the reservation is visible to BOTH pagination passes.** D-4.3.2. See AC3.
- **R4 — exact integer decimal arithmetic only.** `headerHeight` is a `geom.Length` (millipoints).
  No `float32`/`float64` anywhere under `internal/` or on this path (AD-23 / AD-1); `lint`'s
  `TestFloatTypedProductionScan` is the instrument, and a float conversion **compiles**.
- **R5 — no map iteration may reach output order** (D-1.3.5; `lint`'s `TestMapRangeProductionScan`).
  Which pages get a repeated header, and the order its cells and labels are emitted in, must be
  decided by a slice walk, never by ranging a map.
- **R6 — no new glyphs.** The repeated header draws the *same* text the original header already
  shaped. It must not introduce new CIDs, a new subset entry, or a second `pagemodel.ShapedGlyph`
  producer (D-16: `buildShapedPDFRuns` is the only one).
- **R7 — the grouping shape must survive 4.5, 4.6 and 4.8.** 4.5 adds a footer aggregate row into a
  table's span with its own orphan rule; 4.6 adds a row taller than a page; 4.8 adds alternating
  shading keyed on collection row index across pages. Whatever 4.4 adds must not have to be torn out
  by any of the three.
- **R8 — no golden fixture may move.** A new fixture is fine. Confirmed at creation: the only
  repository template carrying a `table` element is `testdata/template/golden/worked-example.json`,
  which is a **template round-trip** golden and is never rendered to a PDF golden — so no committed
  PDF golden contains a paginating content-band table. Verify this rather than trusting it.
- **R9 — no new gate obligation.** `TestEpic2GateObligationsMatchTheDeclaredSet` correctly refused an
  unauthorised one in 4.3. Do not add one.
- **R10 — heavy tests are written, not run** (D-000.4, per-epic cadence). Write the integration/e2e
  coverage, do **not** run it, and **say so explicitly** in the Delivery Log.

---

## Acceptance Criteria

Every AC below that produces observable output carries a **named deletion**, stated behaviourally.
D-000.79: the creator names it because the creator has not written the implementation. D-000.79 §1:
**deletion is the screen, not the replacement** — if deleting the behaviour outright does not redden
the suite, stop; there is no witness at all, and any modification-based red-proof run afterwards is
measuring something unrelated.

Every red-proof runs the **WHOLE suite**, never behind `-run` (D-000.9 extension), records its
command, its counts and the **named** tests that reddened, and restores the tree by hand between runs.

---

### AC1 — the header row is repeated on every continuation page

**Given** a table whose rows continue onto a further page
**When** that page renders
**Then** the table's header row — its cell chrome **and** its column labels — appears at the top of
that table's rows on that page
**And** it still appears once at the table's own declared position on the page where the table begins.

**Anchoring (D-000.68).** The witness is the header's **column-label text**, from literals the test
owns, pairwise distinct, and appearing in **no** data cell of the fixture — plus the header's declared
`headerHeight`, chosen in the fixture to differ from every data-row height (measured at HEAD:
10,000mp vs 10,896mp, already distinct in `multiRowTableDoc`). **Never** cell `x`/`w` alone: every
data row's cells share the header's `x`/`w` exactly, measured above. **Never** a count of rects or of
pages.

**Deletion — part (a), the test of the test (D-000.80).**
> **Remove whatever draws a DATA row's cell chrome**, so the only cells left on a continuation page
> are the repeated header's. **AC1's test must STILL PASS.**

Why it is needed here, stated so the developer does not skip it as ceremony: at HEAD a continuation
page already carries cells at exactly the header's column geometry at exactly the content top. Part
(a) is the only thing that distinguishes *"the test observes the repeated header"* from *"the test
observes a data row that looks like one."* **Part (a) passing is the load-bearing half.**

**Deletion — part (b), on top of (a).**
> **Remove whatever causes the table's header to appear on any page after the one the table begins
> on. The suite must REDDEN.**

A green (b) after a failed (a) is not evidence. Run (a) first.

---

### AC2 — the header's height is accounted for on every continuation page, not only the first

**Given** the repeated header
**When** the content height available for data rows on that page is computed
**Then** the header's height is taken out of it, on **every** continuation page.

Observable as three conjuncts, asserted together, **by value**:

- **(i)** on every continuation page, the first data row's top is at or below the repeated header's
  bottom — the header and the row do not overlap;
- **(ii)** every data row on that page lies **entirely** within the page's content band — nothing
  straddles or is pushed past the bottom edge;
- **(iii)** the **per-page row-index partition**, pinned by value against arithmetic the test derives
  itself from the fixture's own declared geometry.

**On (ii), stated honestly rather than buried:** (ii) is **already true at HEAD**, because nothing is
drawn over anything when no header is repeated. On its own it is worth nothing. It is a conjunct of a
test whose teeth are (i) and (iii), and it is included because the specific way a naive implementation
fails is *"draw the header, forget to take its space"*, which breaks (i) and (iii) and would otherwise
break (ii) only for the last row on the page.

**On (iii), and why a page count cannot stand in for it:** measured above, the 20-row fixture
paginates to **three pages before and three pages after**; only the partition moves (9/10/1 →
9/9/2). D-000.68: *a count is a lossy set.* Assert the partition.

**Deletion.**
> **Remove whatever takes the repeated header's height out of the space that page's data rows may
> use, leaving the header still drawn. The suite must REDDEN** — at (i) and at (iii).

**No part (a) is claimed for AC2**, and the reason is stated rather than assumed: at HEAD no header is
repeated, so there is no accidental cause of (i) or (iii) to remove. If, during development, (iii)
turns out to hold for some reason other than the reservation, that is a part (a) discovered late —
name it and run it rather than shipping past it.

---

### AC3 — both pagination passes agree, so `Page X of Y` stays true

**Given** a document with a paginating table and a `{{page}} of {{pages}}` construct in a repeated band
**When** it renders
**Then** the page count the page-count-only pass computes equals the page count the final pass
produces, **and** the per-page row-index partition is identical between the two.

The fixture must be one where the reservation **changes the partition** (the 20-row measurement above
is such a fixture), otherwise the guard is satisfied by two passes that were already going to agree.
`TestBothPaginationPassesAgreeOnRowPartition` is the existing instrument and the right place to extend.

**Deletion.**
> **Remove whatever makes the page-count-only pass account for the repeated header, leaving the final
> pass accounting for it. The suite must REDDEN**, and it must redden on the *partition* assertion,
> not merely on a count.

---

### AC4 — it holds through the public `Render()`, not only at the layout layer

**Given** the same multi-page table document
**When** it is rendered through the public `Render()` entry point
**Then** the produced PDF carries the table's column-label text on every page that carries any of that
table's data rows
**And** the document's page objects and `/Count` agree with the page count the page-count-only pass
produced.

**Grounds, and they are recent:** Story 4.3 shipped a **live regression** — a table beside any element
sharing its `y` became an unrenderable internal error — and it escaped review because **no test in
that story called `Render()`**. `TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter` and
`TestTableBesideSameYElementRenders` are the shape to follow.

**Deletion.**
> **Remove whatever carries the repeated header from the pagination result into the composed page
> model. The suite must REDDEN at the `Render()` level** — and a layout-level test reddening is not
> a substitute for that.

This is a *separate* deletion from AC1's (b) on purpose: it screens the plumbing independently of the
decision.

---

### AC5 — the band-repeat table and the table's own header are unchanged

**Given** a table declared in the `pageHeader` band
**When** the document paginates
**Then** it is still repeated verbatim on every page and still never paginates.

**Given** a table's own header row that does not fit the remaining content height on the page where
the table begins
**When** pagination runs
**Then** its chrome and its labels still move whole, together, to the next page.

**No deletion is named for this AC, and that is deliberate rather than an omission.** Both clauses
assert that an **existing** behaviour is **unchanged**. The deletion that would redden either is the
deletion of a mechanism this story does not build — 4.3 owns the second clause's mutation and already
ran it; `layout.BandContent` owns the first and predates Epic 4. Naming a deletion here would be
inventing a weak one, which D-000.79 explicitly prefers we decline to do.

---

### AC6 — pagination terminates, and a repeated header never captions nothing

**Given** any document, including one whose rows are tall enough that the reserved header leaves no
room for the next row
**When** pagination runs
**Then** it terminates
**And** no page carries a **repeated** header with zero data rows of that table.

(The page carrying the table's **own** header may hold zero data rows — see DECISION-1.)

**Reachability, measured:** with `contentHeight` 110,000mp and `headerHeight` 10,000mp, any row taller
than 100,000mp and no taller than 110,000mp fits a bare window but not a window under a repeated
header. That state is constructible today and a naive implementation loops on it forever: the row does
not fit, the page advances, the header is reserved again, the row does not fit again.

**Deletion — and read the caveat before running it.**
> **Remove whatever prevents the reserved header height from advancing the page again for the same
> still-unplaced row. The suite must REDDEN.**

**An infinite loop is not a red — it is a hang, and a hang is not evidence.** The guard must therefore
be written so that the failing case *terminates and fails*: bound the assertion against a page count
the test derives from its own fixture (e.g. *"this document's rows cannot produce more pages than
this"*) so the mutant fails an assertion rather than running forever. State this in the test's own doc
comment; a future reader who sees the mutation hang instead of fail will otherwise conclude the guard
is broken.

---

### AC7 — byte-neutrality

**Given** every existing golden
**When** it is re-rendered
**Then** its bytes are identical and no golden fixture has moved.

**Given** a document whose table fits on one page
**When** it renders
**Then** its bytes are unchanged from `ec15d36`.

**No deletion is named.** This is a conservation claim, not an output feature; the mutation that
"reddens" it is any change at all, which makes a named deletion meaningless. Evidence is that every
golden test **genuinely re-renders** and is green — not a hash-of-committed-files check.

---

## Decisions

### DECISION-1 — ANSWERED here, on grounds, and it closes the question Story 4.3 handed forward

Story 4.3 left open: *"Whether a page may hold a table's chrome with zero data rows — unreachable in
4.3, but 4.4 will meet it."*

**The answer is not one answer, because there are two different states and only one of them is new.**

- **A page carrying the table's OWN header and no data row — a widow header — is ALLOWED.** It arises
  when the table's declared `y` puts its header near the foot of a page with no room beneath it for
  row 0. Keeping a header with at least one of its rows is **orphan control**, which is exactly the
  shape of Story 4.5's rule for the footer aggregate row, and `epics.md` scopes orphan control to that
  footer row alone. **Do not build header-orphan control in 4.4.** The next page then repeats the
  header normally, so the reader sees the header on two consecutive pages — correct under FR26, and
  faintly odd to look at. **Flagged:** if the golden report at 4.7 produces a widow header, that is an
  owner call or 4.5's, never a silent fix here.
- **A page carrying a REPEATED header and no data row is FORBIDDEN.** Grounds: a repeated header
  exists to caption rows, so one captioning nothing is meaningless to the reader FR26 names; and it is
  the exact signature of a pagination that made no progress, which `paginate.go`'s own "no page is ever
  empty" rule already refuses in its weaker form. AC6 asserts it.

### DECISION-2 — SURFACED, not taken: what happens when the reserved header leaves no room for the next row?

The state is reachable (arithmetic in AC6). Three candidate answers, each with a real cost:

| | answer | cost |
|---|---|---|
| (a) | suppress the repeat on that page; place the row alone | FR26 is silently broken on that page, with nothing said |
| (b) | refuse, reusing `DiagCodeContentUnlayoutable` | turns a document that renders **today** into a hard error — a regression against `ec15d36` |
| (c) | suppress the repeat **and** warn, reusing an existing diagnostic code | honest and non-regressive; needs an existing code that genuinely fits |

**Recommendation: (c) if an existing code fits the condition without stretching its meaning;
otherwise (a), with the condition recorded for 4.6/4.7.** **Do not mint a new diagnostic code**
(D-000.65: mint only when a condition first ships, and the owner rules the code). **(b) is not
recommended** — it converts working documents into failures.

**Non-negotiable whichever way it is ruled:** AC6's termination and "no repeated header captions
nothing" hold under all three.

**Route this to the engineering lead.** It is not blocking: the story is executable as written with
(a) as the interim behaviour, provided AC6 is met and the choice is recorded.

### DECISION-3 — SURFACED: does the repeated header's space come off the page's top, or sit immediately above the table's own first row on that page?

`epics.md` says *"repeated at the top of **the table** on that page"* — not at the top of the page.
With one full-width table the two coincide; with two tables continuing onto the same page, or a table
beside another element, they do not.

- **Off the page top** (fold the header's height into the page's single `Shift`): simple, keeps
  `PageAssignment`'s one-shift invariant intact — but it displaces **unrelated** siblings downward on
  every continuation page, which AD-24's *"siblings never move"* argues against.
- **Immediately above the table's own first row on that page**: honours `epics.md`'s wording and moves
  nothing that does not belong to the table — but needs a displacement channel narrower than the
  page's single shift.

**Recommendation: honour `epics.md` — the repeat sits at the top of the *table* on that page, and
content that does not belong to the table is not displaced.** **Route to the engineering lead**,
because this choice decides how much machinery 4.4 adds and how much 4.5/4.6 inherit (R7).

**The mechanism itself — where the channel lives, what it is called, how `Paginate` consumes it — is
the developer's call and is deliberately not named here** (D-000.79 §2).

---

## Things the schema and the record could not resolve, surfaced rather than invented

1. **There is no `repeatHeader` field in the schema.** FR26 is unconditional, so every table repeats
   its header and an author cannot switch it off. **Do not add a field.** If the golden report at 4.7
   needs an opt-out, that is a schema question for the owner then. (Same posture as 4.2's note on the
   absent `rowHeight` and 4.3's restatement of it.)
2. **A table's declared `y` governs only where the table BEGINS.** A repeated header has no declared
   position of its own. Recorded so a later reader does not read the repeat as a second declared
   position, or "fix" it to the table's `y` on every page.
3. **Two tables continuing onto the same page is untested territory.** No fixture in the repository
   has it. Build one, and let DECISION-3's ruling decide what it asserts.
4. **Whether the repeated header re-uses the original header's already-shaped runs or fresh copies** is
   not settled by the record. R6 (no new glyphs, no new CIDs, one `ShapedGlyph` producer) constrains the
   answer without determining it. Stated as a constraint; the mechanism is not named.
5. **`deferred-work.md` names Story 4.4 nowhere.** Checked at creation: zero matches. This story
   inherits no deferred item.

---

## Do not re-open — settled rulings this story inherits

- **D-4.2.1** — data-cell padding changes row height; row heights are derived, never declared.
- **D-4.2.2** — row membership must be recoverable from `Paginate`'s input **without inference**; the
  identity is read by direct field lookup, never reconstructed from element id, extent or emission
  order.
- **D-4.3.1** — build the mechanism; do not merely assert an accident.
- **D-4.3.2** — two passes that silently disagree is the defect.
- **D-000.4** — heavy tests are written per story and run at the epic boundary.
- **D-000.65** — reuse an existing diagnostic code; mint one only when a condition first ships.
- **D-000.68** — anchor on the compiler, the type system, or a literal the test owns; a count is a
  lossy set.
- **D-000.9 and its extension** — every measurement names its command and its mutation and confirms
  the mutation ran; every red-proof runs the whole suite.
- **D-000.79 / D-000.80** — the creator names the deletion; part (a) before part (b).
- **AD-14** — never a panic; absent = Error, explicit `null` = empty and not an error, wrong kind =
  Error and never coercion.

Guards that must stay green, by name: `lint`'s `TestGlyphIdentifierCensus`,
`TestSpineStageLadderMatchesStageRankTable`, `TestFloatTypedProductionScan`,
`TestMapRangeProductionScan`; `folio-go`'s `TestFolioMethodNamesAreInjective`,
`TestValidateNeverReachesRenderOrInternalPDF`, `TestPaginateNeverProducesAnEmptyPage`,
`TestEpic2GateObligationsMatchTheDeclaredSet`, the one-byte-producer guard, and
`TestTableInPageHeaderRepeatsIdenticallyAcrossPages`.

---

## Task breakdown

1. [x] **Reproduce the baseline.** All three gates: **1010 / 0 / 1**, **115 / 0**, **3 / 0**, with the
   counting convention stated above. If any differs, stop and re-take the evidence before anything else.
2. [x] **Re-run this story's creation probe** (the figures are reproduced in full above) and confirm both
   conclusions: continuation pages carry **no** header chrome and **no** header labels, and they carry
   data-row cells at **exactly** the header's column geometry at **exactly** the content top. This is the
   pre-fix measurement and AC1's part (a) depends on it. Delete the probe afterwards and verify the tree.
3. [x] **Build the fixtures.** Extend `multiRowTableDoc`/`multiRowTableData` rather than forking them
   where possible. Needed: (i) a table paginating to ≥3 pages whose partition **changes** under the
   reservation; (ii) column labels pairwise distinct and absent from every data cell; (iii) a `{{page}}
   of {{pages}}` variant for AC3; (iv) a row tall enough to exercise AC6; (v) a two-table page for
   DECISION-3, once ruled. Write the arithmetic out beside the expectations.
4. [x] **Make the header appear on continuation pages.** **AC1.**
5. [x] **Make the header's height reduce the space available to that page's data rows, on every
   continuation page.** **AC2.**
6. [x] **Teach both pagination passes the same reservation.** **AC3.** Extend
   `TestBothPaginationPassesAgreeOnRowPartition` on the partition, not the count.
7. [x] **Take it through the public `Render()`.** **AC4.**
8. [x] **Rescope `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated`** — invert its fence, keep both
   "moves whole" halves meaningful, rename it to what it now asserts. **Confirm
   `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` is still green and untouched.** **AC5.**
9. [x] **Guard termination with a bounded assertion**, and say in the test's own doc comment why it is
   bounded (a hang is not a red). **AC6.**
10. [x] **Run every named deletion above, in order, over the WHOLE suite**, never behind `-run`
    (D-000.9 extension). **AC1's part (a) runs BEFORE its part (b).** Record for each: the command, the
    observed pass/fail/skip counts, and the **named** tests that reddened. Restore the tree by hand
    between runs (never `git checkout` over an unrelated modification — see the tree note above).
    **If a deletion does not redden, stop and say so** rather than proceeding to a modification-based
    proof.
11. [x] **Confirm byte-neutrality.** Every golden genuinely re-renders and is green; all three gates
    green; `gofmt -l .` empty and `go vet ./...` clean in all three modules; `go build -tags=matrix
    ./...` and `go vet -tags=matrix ./...` clean. **AC7.**
12. [x] **Write the integration/e2e tests. Do NOT run them** (D-000.4, per-epic cadence) and **say so
    explicitly** in the Delivery Log.
13. [x] **Delivery Log:** gate results for all three modules; every red-proof with its command and
    counts; the per-AC verification table; the **D-000.79 Class A / Class B split for this story**,
    recorded so the ledger's `4.4` row can be filled and the falsification test decided; DECISION-2's
    and DECISION-3's rulings as received; and the list of things deliberately not built.
14. [x] Set the story's status to **review** in `sprint-status.yaml` (status value only).

---

## Delivery Log

### Baseline reproduction (task 1)

Reproduced at session start, HEAD `0ae430e` (the story file's own creation commit was `ec15d36`; HEAD
had advanced to `0ae430e` — "Open D-000.79's Class A/B ledger" — by the time development started; no
folio-go/lint/hashmatrix source changed between the two, only the ledger doc, so the baseline is
unaffected):

| Gate | Command | Measured |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **1010 pass · 0 fail · 1 skip** |
| `lint` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **115 pass · 0 fail** |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **3 pass · 0 fail** |

Matched the story's own creation-time figures exactly. `gofmt -l .` and `go vet ./...` were clean in
all three modules before any change.

### Creation probe re-run (task 2)

Re-ran the probe against `multiRowTableDoc(false)`/20 rows through `collectBandTableRuns` →
`layout.Paginate`, confirming both conclusions the story's creator recorded: page 1/page 2 carried
**zero** header chrome and **zero** header labels, and page 1's first two rects sat at exactly the
header's own column x/w (`x=0 w=80,000` / `x=80,000 w=80,000`), at exactly the content top — the
look-alike. Probe file deleted afterwards; `git status` confirmed clean before any real change was
made. This same look-alike is reconfirmed, as a standing test rather than a one-off probe, by
`TestContinuationPageLookAlikeStillExists` (`folio-go/table_header_repeat_test.go`).

### Mechanism built (developer's call, D-000.79 §2)

**The reservation is automatic and needs no schema field or caller-supplied list** (matches note 1: FR26
is unconditional). `internal/layout.Paginate`'s existing sweep already tags every table's header and
every data row with a `Group` (Story 4.3's mechanism); Story 4.4 adds: any table whose header `Group` is
present among the items handed to `Paginate` is a repeat candidate, full stop. `Paginate`'s own signature
is **unchanged** — no new parameter — so every existing call site (both PHASE A's `contentColumnItems`
and PHASE B's `paginateDocument`, and every test's own local reproduction) gets the mechanism for free,
which is also what makes AC3 hold by construction rather than by two independently-maintained copies.

New types in `internal/layout/paginate.go`:
- `TableHeaderRepeat{ElementID, Rects, Runs, Shift}` — one continuation page's redrawn copy of a
  table's header, reusing the header's OWN `RectRef`/`TextRunRef`s (R6: no new glyphs, no second
  `ShapedGlyph` producer) at a **separate** `Shift`, positioned immediately above that table's own
  first row **on that page** (DECISION-3).
- `TableRowDisplacement{ElementID, Amount}` — the extra downward displacement, beyond
  `PageAssignment.Shift`, applied to one table's own rows to make room for its repeat. Scoped by
  `ElementID`; `PageAssignment.Shift` itself is **untouched** (DECISION-3's non-negotiable).
- `TableHeaderSuppressed{ElementID, Page, RowHeight, Available}` and `Pagination.Suppressed` —
  DECISION-2's data, read by direct lookup in `render.go`, never re-derived.
- `PageAssignment.HeaderRepeats` / `.RowDisplacement` — per-page, appended in the sweep's own
  deterministic order (slices, never a map ranged for emission — R5).

Algorithm: for each table row processed by the existing column-order sweep, once its FINAL page is
resolved (after any window slide), the sweep decides — **once** per `(table, page)`, cached — whether
that page's reservation is honoured: try the row against `windowStart+height-headerHeight`; if it fits,
adopt the reservation (build the `HeaderRepeats`/`RowDisplacement` entries); if not, even alone, fall
back to the unreserved ceiling and record a `TableHeaderSuppressed` (DECISION-2, arm c). This decision
never retries or loops — it runs at most once per item — which is what makes AC6 structurally
loop-safe (see AC6's row below).

`headerContentOf` is a small slice-walk helper (never a map range, R5) that finds a table's header
`ColumnItem`(s) by direct `Group.Key` lookup and returns its `Rects`/`Runs`.

**`render.go`'s `paginateDocument`** composes the final `[]pagemodel.Page`: `HeaderRepeats` are emitted
(rects then runs) before that page's ordinary content, each ref's Y shifted by `TableHeaderRepeat.Shift`
(a value **separate** from `PageAssignment.Shift`); a row's own `ContentRects`/`ContentRuns` additionally
add `RowDisplacement`'s `Amount` (via the new `rowDisplacementFor` slice-walk helper) on top of the page's
ordinary `Shift`. `paginateDocument` now also returns `[]Diagnostic` (DECISION-2's Warning), which
`renderDocument`/`predictDocument` append to `Result.Diagnostics`. **Emitted from the final pass only**
— PHASE A discards `contentPlan.Suppressed` entirely (only reads `len(contentPlan.Pages)`), so a
suppression is never reported twice.

**Diagnostic minted** (engineering lead's ruling corrected an instruction in my own brief — D-000.65
authorises minting *at* the story where a condition first ships, not merely "ahead of" one; my brief
had this backwards): `internal/diag.CodeTableHeaderRepeatSuppressed = "TABLE_HEADER_REPEAT_SUPPRESSED"`,
bridged as `folio.DiagCodeTableHeaderRepeatSuppressed`, `SeverityWarning`, message names the table id,
the page, the row's own height, the space available WITHOUT the reservation, and the three levers a
template author has (reduce headerHeight, reduce row height, increase content height) — D-000.37.
Registry hygiene: `allCodes` gained the one entry; `internal/diag/diag_test.go`'s test-owned pin table
gained its own independent literal (`TestRegistryIsAdditiveOnly` — unweakened, still green).
**`CodeContentUnlayoutable` was NOT reused** — its registered meaning ("an element taller than the
content window it must fit inside") does not cover this row, which is not taller than the window
itself, only taller than the window less a reservation FR26 introduces; reusing it would have stretched
both codes' meanings (engineering lead's ruling, explicit).

### Decisions, as received

**DECISION-1** — answered in the story itself, applied as written: a widow header (table's own header,
zero rows) is allowed; a repeated header with zero rows is forbidden. Enforced structurally — a
`HeaderRepeats`/`RowDisplacement` entry is only ever created at the moment an actual row of that table
is being placed on that page, so a repeat can never exist without at least one row accompanying it — and
tested explicitly by `TestReservedHeaderTerminatesEvenWhenNoRoomForNextRow` and by
`TestHeaderRowMovesWholeToNextPageThenRepeatsOnEveryLaterOne`'s own final assertion (no repeat on the
header's own page).

**DECISION-2** — ruled: **arm (c)**, suppress the repeat on the one page and warn, with a **new**
diagnostic code (the correction above). Implemented exactly as ruled: unconditional fallback
(guarantees AC6's progress), one Warning per `(table, page)`, message names the three levers, emitted
only from the final pass, and verified the diagnostic rides in `Result.Diagnostics` only — **no golden
moved** (AC7's own suite stayed green throughout; no fixture in the repository reaches this condition,
so no existing golden's `Result.Diagnostics` changed either).

**DECISION-3** — ruled: **table-scoped displacement, not page-wide**. `PageAssignment.Shift` keeps its
exact pre-4.4 meaning; the repeat's own `Shift` and the row's own `RowDisplacement` are a **second**
channel, scoped by `ElementID`, layered on top. Built and proven by
`TestTwoTablesOnSamePageEachRepeatAboveTheirOwnFirstRow` (task 3(v)'s two-table fixture): two tables
repeat independently on a shared continuation page, and a third, unrelated sibling element is
positioned by `Shift` alone — unmoved by either table's reservation. **Known, recorded, deliberately
not built**: because a repeating table's own rows move down and unrelated siblings do not, an element
declared immediately below such a table CAN be overlapped by the table's last row on a continuation
page. Not fixed here (per the ruling) — an owner call or a later story's, if the golden report at 4.7
produces this shape.

**Correction recorded, per the ruling**: `epics.md:66`'s one-line FR26 summary ("repeat the table header
at the top of every continuation PAGE") is the loose spelling; the story's own Given/When/Then AC ("at
the top of the TABLE on that page") governs and is what was built. A future reader of `epics.md:66` alone
would read the wrong shape.

### Per-AC verification

| AC | How verified |
|---|---|
| AC1 | `TestTableHeaderLabelsAppearOnEveryContinuationPage` (label-anchored, layout level) + `TestTableHeaderRepeatsThroughPublicRender` (Render level, AC4's own test doubles as AC1's Render-level witness) + `TestContinuationPageLookAlikeStillExists` (the look-alike is still present, so the anchor choice is still load-bearing). Deletion part (a) run — see below. |
| AC2 | `TestReservedHeaderHeightIsAccountedForOnEveryContinuationPage` — (i), (ii), (iii) as three conjuncts, (iii) pinned by value against the test's own derived arithmetic (9/9/2), with a presence precondition that the partition differs from the pre-4.4 shape. |
| AC3 | `TestBothPaginationPassesAgreeOnRowPartition` extended with a Story-4.4 block comparing `RowDisplacement` totals between PHASE A and PHASE B (not merely the count already asserted). |
| AC4 | `TestTableHeaderRepeatsThroughPublicRender` — decodes the actual PDF bytes' `/ToUnicode` CMap and per-page content streams (reusing `multi_page_fixture_test.go`'s own `mpParseToUnicode`/`splitPageContentStreams`/`mpExtractRuns`), checks a header-label run on every page, and cross-checks `/Type /Page` count and `/Count` against PHASE A. |
| AC5 | `TestHeaderRowMovesWholeToNextPageThenRepeatsOnEveryLaterOne` (renamed/rescoped from `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated`, inversion recorded in its own doc comment) — both original "moves whole" halves kept, the old fence replaced with the inverted assertion (repeats via `HeaderRepeats` on every later page, never on the header's own page). `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` re-run, confirmed green and **untouched** (byte-for-byte diff against HEAD: none). |
| AC6 | `TestReservedHeaderTerminatesEvenWhenNoRoomForNextRow` — `tallRowRepeatDoc` (row height 106,896mp, measured directly, between 100,000mp and 110,000mp), `layout.Paginate` run on a goroutine with an explicit 5s bound (`t.Fatal` names the bound and its meaning if exceeded), asserts termination, a bounded page count, every row placed, no repeated-header-with-zero-rows, and `Paginate.Suppressed` populated. |
| AC7 | All three gates re-measured green (below); every golden genuinely re-renders via `multi_page_fixture_test.go:689-703`, `fixture_test.go:294-303`/`:435-444`/`:627`, and `wrapped_text_fixture_test.go:469-475` — each calls `Render()`/re-renders and compares bytes, none is a hash-of-committed-files check. **Corrected by the finisher (Finding 7): the sentence here previously attributed this evidence to `byte_neutrality_test.go`, which is exactly the OPPOSITE kind of guard — a digest-drift check over `os.ReadFile` + SHA-256 against a committed literal, D-000.47's instrument, not a re-render comparator; it detects a golden FILE being hand-edited, and cannot by itself observe the renderer producing different bytes.** `gofmt -l .` empty and `go vet ./...` clean in all three modules; `go build -tags=matrix ./...` and `go vet -tags=matrix ./...` clean. No golden fixture moved (R8 reconfirmed: `testdata/template/golden/worked-example.json` remains the only `table`-carrying template and is never rendered to a PDF golden; `git status` carries no `testdata/`/`fixtures/` modification in this story or its finisher pass). No owning story is named for AC5/AC7's declined deletions elsewhere in this file — see Finding 12's resolution below. |

### Named deletions (D-000.79/D-000.80), whole suite, restored by hand between runs

All commands: `cd folio-go && env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."`. Baseline: **1016 pass · 0 fail · 1 skip** (after this story's own tests were added; the `-skip` flag excludes the permanently-red `TestCorpusMeetsP6ExerciseFloors`). Every mutation restored via `cp` from a pre-mutation backup (never `git checkout`, per the tree note), confirmed by `grep -c "RED-PROOF MUTATION"` returning 0 and `git status`/`git diff --stat` showing only this story's own real changes afterward.

1. **AC1 part (a)** — removed the DATA-row `cellRects` construction/append in `table_render.go`
   (`collectBandTableRuns`'s data-row loop). Result: **994 pass · 21 fail · 1 skip**.
   `TestTableHeaderLabelsAppearOnEveryContinuationPage` **still passed** (confirmed by name) — the
   load-bearing half. 15 other, unrelated tests reddened (row-chrome-specific tests: e.g.
   `TestDataRowBorderIsDrawn`, `TestDataCellPaddingShiftsRowHeightAndContentOrigin`,
   `TestDataRowIdentityIsConsistentAndDistinct`, `TestWrappedCellGrowsTheRowAndNeverTheColumn`, and
   others depending on data-row chrome existing) — expected collateral of deleting a widely-depended-on
   code path, not evidence against part (a)'s own test.
2. **AC1 part (b)**, run only after part (a) confirmed — disabled the "honour" branch in
   `internal/layout/paginate.go` (`if false { … }` in place of the reservation-adoption test). Result:
   **1008 pass · 7 fail · 1 skip**. Named tests that reddened:
   `TestTableHeaderLabelsAppearOnEveryContinuationPage`, `TestTableHeaderRepeatsThroughPublicRender`,
   `TestReservedHeaderHeightIsAccountedForOnEveryContinuationPage`,
   `TestBothPaginationPassesAgreeOnRowPartition`,
   `TestHeaderRowMovesWholeToNextPageThenRepeatsOnEveryLaterOne`,
   `TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter` (+ its `rows=20` subtest).
3. **AC2** — removed the reservation's effect on BOTH the fit-test ceiling (`ceilingFor` ignored
   `amt`) AND the row-displacement append, while leaving `HeaderRepeats` itself still constructed (the
   header "still drawn", per the story's own phrasing). Result: **1011 pass · 4 fail · 1 skip**. Named:
   `TestReservedHeaderHeightIsAccountedForOnEveryContinuationPage` (reddened at **both** (i) — "no
   RowDisplacement reserves the header's height" — and (iii) — wrong partition, `[9,18]`/`[19,19]`
   instead of `[9,17]`/`[18,19]`), `TestBothPaginationPassesAgreeOnRowPartition` (presence precondition:
   "this fixture must exercise FR26's reservation" — the mutation removed the only observable trace of
   it from that test's own vantage), `TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter`.
4. **AC3** — removed the `Group` tagging (both the table-rects loop AND the text-runs loop) in
   `page_number.go`'s `contentColumnItems` (PHASE A only), leaving `paginateDocument` (PHASE B)
   untouched. Result: **1014 pass · 1 fail · 1 skip**. Named: `TestBothPaginationPassesAgreeOnRowPartition`,
   reddening exactly on the partition/count message ("PHASE A produced 3 pages, PHASE B produced 4 —
   {{page}}/{{pages}} would print the WRONG total") — the D-4.3.2 defect, reproduced on demand.
5. **AC4** — removed the two `HeaderRepeats` emission loops in `render.go`'s `paginateDocument` (rects
   and runs), leaving `internal/layout`'s `HeaderRepeats` data itself intact. Result: **1014 pass · 1
   fail · 1 skip**. Named: `TestTableHeaderRepeatsThroughPublicRender` **only** —
   `TestTableHeaderLabelsAppearOnEveryContinuationPage` (the layout-level test) stayed **green**,
   confirmed by name, proving the layout-level test is not a substitute for the Render()-level one, as
   the AC requires.
6. **AC6** — removed the suppression fallback in `internal/layout/paginate.go` (`if true { … }` forcing
   the reservation to be adopted unconditionally, never falling back). Result: **1014 pass · 1 fail · 1
   skip**, returned in well under 1 second — **no hang was observed**. This implementation's sweep
   processes each item at most once (no retry loop), so this particular mutation cannot produce an
   infinite loop structurally; it instead silently drops DECISION-2's diagnostic while still terminating.
   Named: `TestReservedHeaderTerminatesEvenWhenNoRoomForNextRow`, reddening on "`Paginate.Suppressed`
   must record at least one suppression" — the mutation's actual, non-hanging failure mode. The test's
   own 5-second goroutine bound was exercised as infrastructure (it did not fire) and is retained
   because a FUTURE implementation change reintroducing a retry-style loop is exactly the shape it
   guards against.

No deletion failed to redden; none required a modification-based proof as a substitute.

### Byte-neutrality (AC7), final gate measurements

| Gate | Command | Measured |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **1016 pass · 0 fail · 3 skip** (developer's figures, corrected by the finisher — see Finding 6 below and the finisher's own re-measurement further down) |
| `lint` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **115 pass · 0 fail** (unchanged from baseline — no new lint obligation) |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` | **3 pass · 0 fail** (unchanged) |

`gofmt -l .` empty and `go vet ./...` clean in all three modules. `go build -tags=matrix ./...` and
`go vet -tags=matrix ./...` clean in `folio-go`. `TestEpic2GateObligationsMatchTheDeclaredSet` stayed
green throughout — **no new gate obligation was added** (R9): no new `-tags=matrix` file, no new
matrix-document fixture.

Guards confirmed green by name in the final run: `lint`'s `TestGlyphIdentifierCensus`,
`TestSpineStageLadderMatchesStageRankTable`, `TestFloatTypedProductionScan`,
`TestMapRangeProductionScan`; `folio-go`'s `TestFolioMethodNamesAreInjective`,
`TestValidateNeverReachesRenderOrInternalPDF`, `TestPaginateNeverProducesAnEmptyPage`,
`TestEpic2GateObligationsMatchTheDeclaredSet`, `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt`
(the one-byte-producer guard), and `TestTableInPageHeaderRepeatsIdenticallyAcrossPages`.

Total test count rose 1010 → 1016 (+6: `TestTableHeaderLabelsAppearOnEveryContinuationPage`,
`TestContinuationPageLookAlikeStillExists`, `TestReservedHeaderHeightIsAccountedForOnEveryContinuationPage`,
`TestTableHeaderRepeatsThroughPublicRender`, `TestReservedHeaderTerminatesEvenWhenNoRoomForNextRow`,
`TestTwoTablesOnSamePageEachRepeatAboveTheirOwnFirstRow` — `TestHeaderRowMovesWholeToNextPageAndIsNotRepeated`
was renamed in place, not added).

**CORRECTED by the finisher (Finding 6, Major): the developer's own claim of "skip count rose 1 → 1"
and "the single skip in the filtered run is `TestCorpusMeetsP6ExerciseFloors`" were both wrong,**
measured independently by the reviewer and re-confirmed here. `-skip` EXCLUDES
`TestCorpusMeetsP6ExerciseFloors` from the `-v` output entirely — it never prints a `--- SKIP` line
at all in this filtered run — so the one line the developer attributed to it was actually one of this
story's own two new heavy tests. The filtered run (the one the gate reads) carries **THREE** skips,
not one:
```
--- SKIP: TestXrefEntriesRejectsMalformedSubprocess       (pre-existing, conditional, not this story's)
--- SKIP: TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable   (new, D-000.4 heavy test — see below)
--- SKIP: TestTwoTablesWithPageCountFooterRenderConsistently       (new, D-000.4 heavy test — see below)
```
This was the mechanism that hid Finding 1 (the two heavy tests were unconditional, EMPTY `t.Skip`s at
the time of review — see that finding's resolution below): the misreport let "+6 passes, no new
skips" read as clean reconciliation when two of this story's own tests were never executing at all.

### Heavy/integration tests (D-000.4) — REWRITTEN by the finisher (Finding 1, Blocker)

**The developer's own bodies were empty, unconditional `t.Skip`s** — the doc comments described what
the tests "would" do, but no fixture, no `Render()` call, and no assertion existed. D-000.58 declined
exactly this shape (an absence reading as a pass), and it was the mechanism behind Finding 6's
undercounted skip figure above. The finisher chose the **write them for real** branch of Finding 1's
two offered resolutions (rather than delete-and-defer), because both were straightforward to build
from fixtures already in this file:

- `TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable` — now a REAL 500-row render through the
  public `Render()` (bounded to 60s on a goroutine), confirming the header repeats on every one of the
  50+ pages produced, that two renders of the identical document are byte-for-byte identical, and
  that rendering completes within the bound.
- `TestTwoTablesWithPageCountFooterRenderConsistently` — now a REAL render of a new fixture,
  `twoTablesOnSamePageWithPageCountFooterDoc` (DECISION-3's two-table shape plus a `{{page}} of
  {{pages}}` footer), cross-checking the page-count-only pass' page count against the rendered
  `/Type /Page` count and `/Count`, and confirming both tables' header labels appear on every page
  carrying their own rows.

**Both are gated behind an ordinary environment variable, `FOLIO_HEAVY=1`, never a build tag** — a new
`-tags=matrix` file would itself register as an unauthorised Epic 2 gate obligation per
`TestEpic2GateObligationsMatchTheDeclaredSet`'s own scan (which looks specifically for the matrix
build constraint), so an env-gated ordinary test stays outside that obligation set by construction
(R9 preserved). The routine gate therefore still SKIPS both (the var is unset), which is why they
still appear as `--- SKIP` lines in every count above — but unlike the story's original shipped
version, `FOLIO_HEAVY=1 go test -run 'TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable|TestTwoTablesWithPageCountFooterRenderConsistently' -v ./...` (from `folio-go/`) now actually exercises
them, and the finisher ran that command once, by hand, before this commit: **both PASS.** The exact
command is also recorded in the new `heavyTestGateEnvVar` doc comment in
`folio-go/table_header_repeat_test.go`, so the Epic 4 boundary gate has the incantation without
needing to re-derive it.

### D-000.79 Class A / Class B split for this story

**Class A (the accidental-pass trap)**: AC1 — the continuation-page look-alike (data-row cells at the
header's exact column geometry) would have made a geometry-only assertion pass with FR26 unimplemented.
Guarded by anchoring on column-label TEXT and headerHeight, never geometry, and by keeping
`TestContinuationPageLookAlikeStillExists` as a standing witness that the trap is still present and
still correctly avoided.

**Class B (build the mechanism, not merely assert an accident)**: every AC's deletion reddened the
suite (see above) rather than the suite already passing before any implementation code was written —
none of AC1–AC6 was accidentally satisfied by pre-existing behaviour.

This story's row in the D-000.79 ledger (`folio-mvp-decision-log.md`) can be filled `discharged` — both
halves of the falsification test (a Class A trap found and avoided; every AC's mechanism proven by a
reddening deletion) are satisfied. The ledger append itself was left exactly as found at story creation
(not this story's file to edit) and is not modified here.

### Things deliberately not built

- **No `repeatHeader` schema field.** FR26 is unconditional (note 1, reconfirmed).
- **No header-orphan control.** A widow header (table's own header, zero rows) is allowed and
  unmanaged, per DECISION-1 — Story 4.5's territory if ever needed.
- **No fix for a repeating table's last row overlapping an unrelated element declared just below it on
  a continuation page.** Known, per DECISION-3's own guardrail — an owner call or a later story's.
- **DECISION-2's suppression is unconditional arm (c) only** — no template-level control over whether a
  page silently suppresses vs. hard-fails; (b) was explicitly vetoed by the ruling.
- **Footer aggregates, alternating shading, over-tall-row clipping** — untouched, Stories 4.5/4.6/4.8's
  own territory (R7 confirmed: nothing this story added needs to be torn out for any of the three — the
  per-table/per-page channel is exactly what the engineering lead's DECISION-3 ruling says those stories
  will also need).

---

## Review Findings

## Review Summary
- Reviewed by: bmad-code-reviewer
- Date: 2026-08-27
- Story Status Recommendation: **Changes Requested**
- Blockers: 3
- Majors: 5
- Minors: 6
- Nits: 1

### Gate figures measured in this review (not carried from the Delivery Log)

Counting convention: `--- PASS` / `--- FAIL` / `--- SKIP` anywhere in `-v` output.

| Gate | Command | Measured here | Delivery Log claims |
|---|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -skip ^TestCorpusMeetsP6ExerciseFloors\$ -v ./..."` | **1016 pass · 0 fail · 3 skip** | 1016 · 0 · **1** |
| `lint` | `go test -count=1 ./...` | **115 pass · 0 fail** | 115 · 0 |
| `hashmatrix` | `go test -count=1 ./...` | **3 pass · 0 fail** | 3 · 0 |

`gofmt -l .` (bare, not via rtk) empty; `go vet ./...` clean; `go build -tags=matrix ./...` and
`go vet -tags=matrix ./...` clean. Named guards confirmed green by name: `TestGlyphIdentifierCensus`,
`TestSpineStageLadderMatchesStageRankTable`, `TestFloatTypedProductionScan`, `TestMapRangeProductionScan`
(lint); `TestFolioMethodNamesAreInjective`, `TestValidateNeverReachesRenderOrInternalPDF`,
`TestPaginateNeverProducesAnEmptyPage`, `TestEpic2GateObligationsMatchTheDeclaredSet`,
`TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt`, `TestRegistryIsAdditiveOnly`, and
`TestTableInPageHeaderRepeatsIdenticallyAcrossPages` (green **and byte-for-byte untouched** —
`git diff --stat` on `table_render_test.go` is empty). No golden fixture moved: `git status` carries no
`testdata/` or `fixtures/` modification at all.

### Red-proofs re-run independently in this review

Every run is the WHOLE suite, `-count=1`, never behind `-run`. Mutations applied and reverted by hand
(`cp` from a pre-mutation backup), with SHA-256 re-verified against the pre-review tree after each.

| ID | Mutation | Result | Verdict |
|---|---|---|---|
| RP-A | `render.go`: both `for _, rep := range assigned.HeaderRepeats` loops (runs + rects) neutered to range a nil slice | **1015 · 1 · 3** — `TestTableHeaderRepeatsThroughPublicRender` **only**; `TestTableHeaderLabelsAppearOnEveryContinuationPage` stayed green | **AC4's discrimination CONFIRMED.** The developer's claim holds exactly. |
| RP-B | `internal/layout/paginate.go`: `TableHeaderRepeat.Shift` displaced by `- 37000` (37pt) | **1016 · 0 · 3 — nothing reddened** | Finding 3 |
| RP-C | `render.go`: row displacement leaked to EVERY run and rect on the page (`isTableRowLine` / `rectIsDataRow` / `d.ElementID == elementID` guards all forced true) — i.e. DECISION-3's rejected page-wide shape | **1016 · 0 · 3 — nothing reddened** | Finding 2 |
| RP-D | `render.go`: the whole `for _, s := range plan.Suppressed` diagnostic-construction block neutered | **1016 · 0 · 3 — nothing reddened** | Finding 4 |
| RP-E | `internal/layout/paginate.go`: row atomicity deleted (group-page short-circuit disabled AND group extent override removed) | **1007 · 9 · 3**; `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer`, `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder`, `TestPaginateGroupPartitionPinnedByValue`, `TestPaginateGroupSurvivesAnInterveningPageAdvance`, `TestBothPaginationPassesAgreeOnRowPartition`, `TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter` reddened — the **rescoped** test did NOT | Finding 13 |

Direct measurements (probe run against an isolated `/tmp` copy of the module; the repository tree was
never mutated for these):

- `tallRowRepeatDoc` (AC6's fixture) produces **0 `HeaderRepeats` on every one of its 4 pages** —
  see Finding 5.
- `twoTablesOnSamePageDoc`'s `e20` produces **1 run with 1 distinct `itemTop`** — see Finding 2.
- `contentRuns[].text` is **per-cell** (`"R15W-b"`), not per-glyph — see Finding 10.
- `runCarriesAHeaderLabel("a") == true`, `runCarriesAHeaderLabel("o") == true` — see Finding 9.

---

### Finding 1: The D-000.4 heavy tests are unconditional `t.Skip`s with empty bodies — the cadence's debt is discharged silently
- **Severity**: Blocker
- **Category**: Tests
- **Location**: `folio-go/table_header_repeat_test.go:162-164` and `:175-177`
- **Observation**: Both "heavy/integration" tests consist of *nothing but* a `t.Skip`:
  ```go
  func TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable(t *testing.T) {
      t.Skip("D-000.4: heavy/integration suite, deferred to the Epic 4 boundary gate — written, not run, by Story 4.4 (see the story's Delivery Log)")
  }
  ```
  There is no env-var check, no build tag, no `testing.Short()` guard — nothing the Epic 4 boundary
  gate can flip to turn them on. Worse, the bodies are **empty**: the doc comments say *"It **would**
  render a 500-row table…"*, and no fixture, no `Render()` call, no assertion was ever written. The
  Delivery Log states these are "written, not run"; measured, they are **neither written nor
  runnable**.
- **Impact**: D-000.58 declined a `t.Skip` for exactly this shape because it *"trades a loud
  environmental failure for a quieter one"* — **an absence reads as a pass**. The per-epic cadence
  says heavy tests are *written, not run*; an unconditional skip means they will also be skipped at
  the epic-boundary catch-up run, which is the one moment they exist for. Story 4.4's heaviest claims
  — that the repeat holds across hundreds of pages, and that two simultaneous reservations keep PHASE
  A and PHASE B in agreement — have no coverage at any point in the project's life, while the
  Delivery Log records the obligation as met.
- **Suggested Resolution**: Write the two test bodies for real, then gate them on a condition the
  epic gate can actually set (an env var such as `FOLIO_HEAVY=1`, or a build tag that does **not**
  register a new gate obligation), and record the exact command that turns them on in the Delivery
  Log and in each test's doc comment. If the bodies genuinely cannot be written this story, say so
  plainly and file the debt rather than shipping a skip that reads as coverage.
- **Related AC**: R10 / D-000.4; task 12

---

### Finding 2: The displacement invariant is untested — the two-table test's own named "teeth" is a tautology
- **Severity**: Blocker
- **Category**: Tests
- **Location**: `folio-go/table_header_repeat_test.go:520-545` (`TestTwoTablesOnSamePageEachRepeatAboveTheirOwnFirstRow`)
- **Observation**: The test's doc comment quotes the engineering-lead ruling and names the third
  conjunct — *"an element belonging to NEITHER table is at the same Y it would have had with no
  repeat at all — that last conjunct is the one that would fail under the page-wide shape, so it is
  the fixture's teeth."* The assertion built for it is:
  ```go
  gotTop  := contentRuns[ref].itemTop - pg.Shift
  wantTop := e20Top - pg.Shift
  if gotTop != wantTop { t.Errorf(...) }
  ```
  `e20Top` was loaded, earlier in the same function, from `contentRuns[i].itemTop` for the first run
  whose `elementID == "e20"`. Measured: **e20 produces exactly one run, with exactly one distinct
  `itemTop`.** Both operands are therefore the same value minus the same `pg.Shift` — a constant
  compared against itself. The branch cannot fire. It also never observes a *rendered* position: it
  never goes through `paginateDocument`, so the page-wide shape (folding the reservation into
  `PageAssignment.Shift`) would move both sides equally and still pass. The surviving companion
  assertion, `rowDisplacementForTest(pg.RowDisplacement, "e20") != 0`, checks only that e20 is absent
  from a list — it too returns 0 under the page-wide shape.
- **Impact**: Confirmed by red-proof **RP-C**: leaking the table's displacement onto *every* run and
  rect on the page — precisely the page-wide displacement AD-24 and DECISION-3 forbid — leaves the
  suite at **1016 · 0 · 3, entirely green**. The invariant *"a table's pagination must never change
  the position of any element that is not part of that table"* has no witness anywhere in the
  repository. This is also Story 4.3's R7 regression shape: `TestTableBesideSameYElementRenders`
  (`table_pagination_test.go:956`) renders only **3 rows**, so it is a single page with no
  continuation and no repeat — the same-`y` sibling is never exercised against a repeating header.
- **Suggested Resolution**: Assert the invariant on the **rendered** artifact, not on layout inputs:
  render the two-table document (plus a sibling declared at the **same `y` as the table**, per 4.3's
  R7 shape) through `Render()` or `paginateDocument`, capture the sibling's final `Y` on a
  continuation page that carries a `HeaderRepeats` entry, and pin it **by value** against a baseline
  taken with no repeating table on the page. RP-C is the mutation the new test must redden under.
- **Related AC**: DECISION-3, R2, R7 (AD-24)

---

### Finding 3: Nothing pins the repeated header's position — it can be drawn 37pt off and the suite stays green
- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `internal/layout/paginate.go:701` (`Shift: headerItem.top - effectiveTop + pages[page].Shift`); assertions at `table_header_repeat_test.go:84-149` (AC1) and `:735-757` (AC2 conjunct (i))
- **Observation**: AC1 requires the header appear *"at the top of that table's rows on that page"*
  and AC2(i) requires *"the first data row's top is at or below the repeated header's bottom — the
  header and the row do not overlap."* Neither is observed. AC1's test asserts only that a
  `HeaderRepeats` entry exists carrying a run whose text names a label; it never reads `rep.Shift`.
  AC2(i) was implemented as a **proxy** — *"a continuation page with >=1 data row of this table must
  carry a `RowDisplacement` entry of exactly headerHeight"* — which observes the reservation
  bookkeeping, not the two objects' relative positions. AC2(ii)'s in-band check is applied to data
  rows only; the repeat itself is never bounds-checked.
- **Impact**: Confirmed by red-proof **RP-B**: displacing `TableHeaderRepeat.Shift` by 37,000mp (37pt
  on a 150pt page) leaves the suite at **1016 · 0 · 3, entirely green**. The repeated header could
  render on top of the first data row, or off the content band entirely, and every AC in this story
  would still pass. Note the arithmetic in the shipped code is *correct* — I verified it by hand
  (`repeatTop_rendered = effectiveTop - pages[page].Shift`, exactly the position the row vacates) —
  so this is a coverage defect, not a live bug. But under D-000.79/D-000.80 an unpinned observable is
  exactly what the deletion screen exists to catch, and AC2's own named deletion passed through the
  proxy instead of the property.
- **Suggested Resolution**: Add a by-value assertion on the repeat's final position: on each
  continuation page, `repeat_bottom <= first_data_row_top` and `repeat_top >= contentTop`, computed
  from `rep.Shift` and the header's own extent, with the expected value derived from the fixture's
  declared geometry. RP-B is the mutation it must redden under.
- **Related AC**: AC1, AC2(i)

---

### Finding 4: The newly minted `TABLE_HEADER_REPEAT_SUPPRESSED` diagnostic is executed by no test
- **Severity**: Major
- **Category**: Tests / Correctness
- **Location**: `folio-go/render.go:1955-1965` (construction); `folio-go/diagnostic.go:258`; `folio-go/internal/diag/diag.go:163`
- **Observation**: `grep TABLE_HEADER_REPEAT_SUPPRESSED --include=*_test.go` hits exactly one line —
  the pin at `internal/diag/diag_test.go:42`. The construction block in `render.go` is never reached
  by any test: the AC6 test stops at `plan.Suppressed` (the layout-level data) and never renders the
  tall-row fixture. Confirmed by red-proof **RP-D**: neutering the entire block leaves the suite at
  **1016 · 0 · 3, green**. Nothing asserts the message text, the `SeverityWarning` choice, the
  `ElementID`, or that the Warning actually arrives on `Result.Diagnostics`.
- **Impact**: D-000.65 authorises minting a code *when its condition first ships*; the condition ships
  here but the code's author-facing contract does not ship with a witness. The message is the entire
  user-visible deliverable of DECISION-2 arm (c) — the difference between "honest and non-regressive"
  and "silently broken on that page" — and it is unverified. (On the positive side: registry hygiene
  is correct — `allCodes` gained the entry, the pin table gained its own independent literal,
  `TestRegistryIsAdditiveOnly` covers it unweakened; the message text *does* name all three levers
  D-000.37 requires, in author-actionable terms; and the diagnostic provably does not reach the
  emitted bytes — `diags` is threaded past `pdf.SerializeTextDocument` untouched, and RP-D moved no
  golden.)
- **Suggested Resolution**: Add a `Render()`-level test on `tallRowRepeatDoc` asserting
  `Result.Diagnostics` carries exactly one `DiagCodeTableHeaderRepeatSuppressed` per suppressed
  `(table, page)`, at `SeverityWarning`, with `ElementID == "e1"`, and that the message names all
  three levers (assert on the lever substrings, from literals the test owns). RP-D is the mutation it
  must redden under.
- **Related AC**: DECISION-2, D-000.65, D-000.37

---

### Finding 5: AC6's "a repeated header never captions zero rows" loop never executes its body
- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/table_header_repeat_test.go:342-361` (`TestReservedHeaderTerminatesEvenWhenNoRoomForNextRow`)
- **Observation**: The guard is
  ```go
  if !hasRepeat { continue }
  ... if !hasRow { t.Errorf("page %d carries a REPEATED header for e1 with zero data rows of e1 — forbidden (DECISION-1)", p) }
  ```
  Measured directly against this test's own fixture: `tallRowRepeatDoc` produces **`HeaderRepeats=0`
  on all four pages** (every page suppresses, which is the fixture's whole purpose). `hasRepeat` is
  therefore false on every page and the `t.Errorf` is unreachable. No other test in the repository
  asserts the repeat⇒row direction: AC2's test asserts the *converse* (row ⇒ displacement exists),
  and AC1's and the two-table test assert presence only.
- **Impact**: DECISION-1's **FORBIDDEN** half — the one of its two states that is new in this story —
  is guarded by an assertion that cannot fail on the only fixture that runs it. The structural
  argument in the Delivery Log ("a repeat entry is only ever created while a row is being placed") is
  sound and I verified it by reading `paginate.go:686-720`, but it is an argument, not a witness.
  DECISION-1's **ALLOWED** half (a widow header) *is* reached by this fixture — page 0 carries the
  header with zero rows — but is likewise never asserted as allowed, only left unrejected.
- **Suggested Resolution**: Assert the invariant on a fixture where repeats **do** occur (the AC1/AC2
  20-row fixture, or the two-table fixture) so the loop body executes, and add a presence
  precondition (`t.Fatal` if no page carries a repeat) so the loop can never quietly become vacuous
  again. Separately, assert page 0's widow header positively as the allowed case.
- **Related AC**: AC6, DECISION-1

---

### Finding 6: The Delivery Log's gate figure understates the skip count and misattributes the baseline skip
- **Severity**: Major
- **Category**: Tests / Convention (D-000.9)
- **Location**: story file, "Byte-neutrality (AC7), final gate measurements" table; and "Baseline reproduction (task 1)"
- **Observation**: The Delivery Log records the final `folio-go` gate as **1016 pass · 0 fail · 1
  skip**. Measured here with the identical command and the identical counting convention: **1016 pass
  · 0 fail · 3 skip**. The three are:
  ```
  --- SKIP: TestXrefEntriesRejectsMalformedSubprocess
  --- SKIP: TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable
  --- SKIP: TestTwoTablesWithPageCountFooterRenderConsistently
  ```
  The story's narrative compounds this: it states the single skip in the filtered run is
  `TestCorpusMeetsP6ExerciseFloors` and that the two heavy tests appear only in the **un**filtered
  run. Both claims are wrong — `-skip` excludes `TestCorpusMeetsP6ExerciseFloors` from the output
  entirely, and the two new `t.Skip`s print `--- SKIP` in the filtered run, which is the run the gate
  reads.
- **Impact**: This is the mechanism by which Finding 1 stays invisible. The one number the epic gate
  inspects reports the two permanently-skipped tests as though they were the single, known,
  red-by-design skip. A reader reconciling 1010/0/1 → 1016/0/1 sees a clean +6 and no new skips.
- **Suggested Resolution**: Correct both tables to 1016 · 0 · **3**, name the three skipped tests
  explicitly, and state which of them are new in this story and why.
- **Related AC**: AC7, D-000.9

---

### Finding 7: AC7's evidence cites `byte_neutrality_test.go` as a re-render check; it is exactly the hash-of-committed-files check it is claimed not to be
- **Severity**: Major
- **Category**: Tests / Convention
- **Location**: story file, per-AC verification table, AC7 row; `folio-go/byte_neutrality_test.go`
- **Observation**: The Delivery Log states *"every golden test **genuinely re-renders** (none is a
  hash-of-committed-files check — confirmed by reading `byte_neutrality_test.go`'s own re-render
  calls)."* `byte_neutrality_test.go` contains **zero** `Render(` calls. Its two tests
  (`TestGoldenDigestAgreesAtEveryDeclaredSite`, and the completeness half) do
  `os.ReadFile(pdfPath)` → SHA-256 → compare against a digest literal declared in
  `goldenDigestRecord` (`:281-292`, `:304-360`). That detects a golden *file* being edited; it cannot
  observe the renderer producing different bytes.
- **Impact**: The cited instrument cannot support the claim made of it. AC7 does hold in substance —
  genuine re-render comparisons exist elsewhere (`multi_page_fixture_test.go:689-703`,
  `fixture_test.go:294-303`/`:435-444`/`:627`, `wrapped_text_fixture_test.go:469-475`) and all are
  green, and `git status` shows no `fixtures/` or `testdata/` modification at all — so no golden
  moved. But a future reader relying on this sentence would believe goldens are protected by a
  mechanism that is not the one protecting them.
- **Suggested Resolution**: Re-attribute the AC7 evidence to the tests that actually re-render and
  compare, and describe `byte_neutrality_test.go` as what it is — the D-000.47 digest-drift guard.
- **Related AC**: AC7, R8

---

### Finding 8: `TableHeaderSuppressed.Available` reports the full content height, which the diagnostic message calls "available on that page"
- **Severity**: Major
- **Category**: Correctness
- **Location**: `internal/layout/paginate.go:717` (`Available: height`); message at `render.go:1961`
- **Observation**: The suppression decision is taken against `windowStart + height - hh`
  (`paginate.go:691`), i.e. against the space remaining **from this row's own window position**. The
  value recorded and reported is `height` — the page's whole content band. The two coincide only when
  the window slid to this row's own top. When the window was slid by a *different* element and the
  table's first row on that page rode along (reachable — the two-table fixture at
  `table_header_repeat_test.go:437` has exactly this interleaving), the real room above the row is
  `windowStart + height - effectiveTop`, strictly less than `height`. The message then reads *"the
  next row is 106896mp tall and only 110000mp is available on that page"* when the actual figure was
  smaller.
- **Impact**: The message is the author's only handle on the condition (D-000.37, "executable by a
  human"). An overstated `Available` makes the arithmetic unreproducible from the message: an author
  who computes `Available - headerHeight` and finds the row would have fit is told the opposite of
  what happened. It also weakens the "never re-derived" claim — the reported number is not the
  number the decision used.
- **Suggested Resolution**: Record `Available: windowStart + height - effectiveTop` (the space the
  row actually had), or rename the field and reword the message so it says what it means. Either way,
  add the header's own reserved height to the message — the "reduce headerHeight" lever currently
  gives the author no number to act on.
- **Related AC**: DECISION-2, D-000.37, D-4.2.2

---

### Finding 9: `runCarriesAHeaderLabel` is a two-way substring match; it passes today only by an accident of the fixture's alphabet
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_header_repeat_test.go:71-79`
- **Observation**: `if containsSubstring(l, text) || containsSubstring(text, l)` — the first disjunct
  matches whenever the run's text is *any substring of a label*. Measured:
  `runCarriesAHeaderLabel("a") == true` and `runCarriesAHeaderLabel("o") == true`. The witness is
  therefore not "this run carries the label" but "this run's text is contained in, or contains, a
  label". It discriminates on the shipped fixtures only because the data alphabet (`R W - x b 0-9`)
  happens to be **case-sensitively** disjoint from the label alphabet (`C o l A p h a B r v`), and
  because `mpExtractRuns` accumulates a whole `BT`/`ET` block so the footer arrives as
  `"Page 2 of 3"` rather than as single glyphs.
- **Impact**: A one-character change to a data value (any cell containing a lowercase `a` or `o`), or
  any future change that splits runs more finely, silently converts every AC1/AC4 header assertion
  into a tautology — with no test failing to announce it. Story 4.8 (alternating rows) and 4.5
  (footer aggregates) will both touch these fixtures.
- **Suggested Resolution**: Drop the first disjunct and require the run to contain the whole label
  (`containsSubstring(text, l)` only), or assert on a reassembled per-page text. Add a self-check that
  `runCarriesAHeaderLabel` returns false for every data-cell text the fixture produces.
- **Related AC**: AC1, AC4, D-000.68

---

### Finding 10: The file's doc comment states text runs are per-glyph; measured, they are per-cell
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/table_header_repeat_test.go:67-70`
- **Observation**: *"Text runs are per-glyph in this pipeline (positionSegments' own shape), so a
  label may arrive split across several runs; contains, not equals, is the right test."* Measured on
  page 1 of the AC1 fixture, the distinct `ContentRun` texts are `"R9W-x"`, `"R9W-b"`, `"R10W-x"` … —
  one run per **cell**, not per glyph.
- **Impact**: The false premise is the stated justification for Finding 9's loose matcher. A future
  maintainer who trusts it will keep the two-way `contains` that the actual run shape does not
  require.
- **Suggested Resolution**: Correct the comment and tighten the matcher accordingly.
- **Related AC**: AC1

---

### Finding 11: AC2's two per-page capacities are the identical expression, leaving a dead branch
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/table_header_repeat_test.go:672-673`
- **Observation**:
  ```go
  rowsPerFirstPage := int((contentHeight - headerHeight) / rowHeight)
  rowsPerContinuationPage := int((contentHeight - headerHeight) / rowHeight)
  ```
  Both sides are character-for-character the same, so the `if page == 0 { cap = rowsPerFirstPage }
  else { cap = rowsPerContinuationPage }` below has no effect and the two names document a
  distinction the code does not make.
- **Impact**: The arithmetic is nonetheless **correct** (page 0 pays the header's own height, every
  continuation page pays the repeat's — the same 10,000mp), and the partition pin genuinely has teeth:
  the `preFixPartition` comparison at `:709-731` correctly separates 9/9/2 from 9/10/1. This is a
  readability defect, but a load-bearing one — the moment 4.5's footer aggregate or 4.6's over-tall
  row makes the first page differ from a continuation page, the collapsed expression will be silently
  wrong on one of them.
- **Suggested Resolution**: Either collapse to one name with a comment explaining why the two
  capacities coincide, or keep both names and derive each from the quantity it actually represents.
- **Related AC**: AC2(iii)

---

### Finding 12: AC7 names no owning story number for its declined deletion; AC5 names one for only one of its two clauses
- **Severity**: Minor
- **Category**: Convention
- **Location**: story file, AC5 (`No deletion is named for this AC…`) and AC7 (`No deletion is named.`)
- **Observation**: The ruling required an AC that declines to name a deletion to name the **owning
  story number** instead. AC5 does this for its second clause (*"4.3 owns the second clause's mutation
  and already ran it"*) but not its first (*"`layout.BandContent` owns the first and predates Epic
  4"* — a package, not a story). AC7 names none at all.
- **Impact**: A later reader cannot tell where the missing screen lives, which is the whole purpose of
  the requirement.
- **Suggested Resolution**: Name the story that owns `layout.BandContent`'s band-repeat behaviour, and
  the story that owns each golden's re-render comparison, in the two declining paragraphs.
- **Related AC**: AC5, AC7

---

### Finding 13: The rescoped test's two surviving "moves whole" halves did not redden when row atomicity was deleted
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_pagination_test.go:749-791`
- **Observation**: The fence inversion itself is correct and complete — the test **was renamed, not
  deleted** (`TestHeaderRowMovesWholeToNextPageAndIsNotRepeated` →
  `TestHeaderRowMovesWholeToNextPageThenRepeatsOnEveryLaterOne`), its doc comment **records that 4.4
  authorised the inversion**, both "moves whole" halves are retained verbatim in substance, and a new
  "no repeat on the header's own page" assertion was added. But red-proof **RP-E** (group-page
  short-circuit disabled *and* group extent override removed — i.e. row atomicity deleted outright)
  reddened nine tests and **this was not one of them**.
- **Impact**: The halves are green but carry no teeth on *this* fixture, because its header's chrome
  and labels share an extent — the "chrome accident" Story 4.3 itself named. The property is properly
  guarded by `TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer` and
  `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder`, both of which reddened under RP-E, so this
  is not a regression introduced by 4.4 — but the story's instruction was that the halves *"stay green
  and stay meaningful"*, and only the first is true here.
- **Suggested Resolution**: Either note in the test's doc comment that the atomicity property is
  carried by the named layout-package siblings and this fixture only re-confirms it, or give the
  fixture divergent chrome/label extents so the halves become independently load-bearing.
- **Related AC**: AC5

---

### Finding 14: No test asserts that every registered diagnostic code has a public `DiagCode*` bridge
- **Severity**: Minor
- **Category**: Tests (pre-existing hole, not closed by this story)
- **Location**: `folio-go/diag_bridge_test.go:15-22`
- **Observation**: `TestDiagCodeBridgePreservesExactStrings` hardcodes exactly two codes
  (`DiagCodeTextClippedWidth`, `DiagCodeEmptyAverage`) and has not been extended since Stories 2.8 /
  3.3. `diag.All()` has no caller outside the `diag` package, so nothing asserts bridge completeness
  or that a bridge preserves its string. The new `DiagCodeTableHeaderRepeatSuppressed` does not
  participate.
- **Impact**: A `DiagCode*` bound to the wrong constant would ship silently. This story did not create
  the hole, but it added the 13th code to a surface guarded for 2.
- **Suggested Resolution**: Replace the hardcoded pair with an enumeration over `diag.All()` asserting
  each has a public bridge whose value equals the internal constant's string. Reasonable to defer to
  the epic gate, but it should be filed rather than left silent.
- **Related AC**: D-000.65

---

### Finding 15: The diagnostic message prints millipoints while the author writes points
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/render.go:2075` (`millipointsForDiag`), message at `:1961`
- **Observation**: The message renders `106896mp` / `110000mp`; the template that produced them says
  `"headerHeight": 10` and `"height": 150` — points. Acting on the message requires a mental ÷1000.
- **Impact**: Minor friction only, and it is **consistent with existing precedent** —
  `OverflowError.Error()` (`internal/layout/paginate.go:229-236`) uses the same spelling. Recorded so
  it is a house-wide decision rather than a per-message accident.
- **Suggested Resolution**: None required for this story. If ever changed, change it for
  `OverflowError` at the same time.
- **Related AC**: D-000.37

---

### ACs explicitly considered

- **AC1** — *satisfied in mechanism, under-witnessed.* The repeat genuinely appears on every
  continuation page (RP-A/part-(b)-equivalent mutations redden), and the **Class A trap the creator
  named is correctly avoided**: every header assertion anchors on label **text**, never on cell
  geometry, and `TestContinuationPageLookAlikeStillExists` stands as a live witness that the
  look-alike is still present. This is the story's strongest work. Position is unpinned — Finding 3;
  matcher is loose — Finding 9.
- **AC2** — *satisfied.* (iii)'s partition is pinned by value (9/9/2) against the test's own
  arithmetic, with a working presence precondition separating it from the pre-4.4 9/10/1 shape; the
  page **count** is correctly never used as the witness. (i) is a proxy — Finding 3. (ii) is honest
  about being weak, as the story predicted.
- **AC3** — *satisfied.* The reservation lives inside `layout.Paginate` itself, so both passes get it
  from one derivation rather than two maintained copies; the extension compares `RowDisplacement`
  totals, not counts, and has a live presence precondition. PHASE A discards `.Suppressed`
  (`render.go:1487-1491` reads only `len(contentPlan.Pages)`), so no double-reporting.
- **AC4** — *satisfied, and independently re-verified.* **RP-A reproduces the developer's
  discrimination exactly**: removing the render-level wiring reddens `TestTableHeaderRepeatsThroughPublicRender`
  **only**, at 1015 · 1 · 3, with the layout-level test staying green. The test decodes real PDF bytes
  (`/ToUnicode` + per-page content streams) and cross-checks `/Count`. Story 4.3's escape route is
  closed.
- **AC5** — *satisfied.* `TestTableInPageHeaderRepeatsIdenticallyAcrossPages` is green and
  byte-for-byte untouched; `table_render_test.go` and `table_render.go` carry no diff at all. Fence
  inversion done correctly — Finding 13 is a strength-of-teeth note, not a failure.
- **AC6** — *termination satisfied, forbidden-state unwitnessed.* Termination is structural (the sweep
  visits each item at most once; no retry loop exists), and the bound is correctly written — a 5s
  goroutine bound plus `maxPages := rows + 1`, both with `t.Fatal`s naming the bound and its meaning,
  exactly as the ruling required. The forbidden state is Finding 5.
- **AC7** — *satisfied in substance.* No golden moved on evidence that could detect a move: `git
  status` carries no `fixtures/`/`testdata/` change, the genuine re-render comparisons are green, and
  RP-D confirmed the new diagnostic cannot reach emitted bytes. Evidence attribution is Finding 7.

### Constraints verified clean (no findings)

R1 (`MixedItemError` untouched and not widened — `TableHeaderRepeat` is an output type, never a
`ColumnItem`, so carrying both `Rects` and `Runs` is outside the exclusivity rule by type rather than
by exemption; the header remains two `ColumnItem`s sharing one group key, which is why
`headerContentOf` accumulates from both). R2/AD-24 (the column is never mutated; `PageAssignment.Shift`
keeps its exact pre-4.4 meaning). R4/AD-23/AD-1 (all six added arithmetic sites are `geom.Length`
int64; no float, no division, no `time`/`os`/`math/rand`/`net`/`math`, no package-level mutable state;
`TestFloatTypedProductionScan` confirmed to actually cover `internal/layout`). R5/D-1.3.5 (zero map
ranges added anywhere; `reservation`, `headerPageOf` and `groups` are read by direct lookup only, and
`HeaderRepeats`/`RowDisplacement`/`Suppressed` are appended inside the stably-sorted `order` sweep).
R6 (the repeat reuses the header's own `TextRunRef`/`RectRef` indices into the same tables;
`buildShapedPDFRuns` remains the only `ShapedGlyph` producer and runs before pagination;
`TestGlyphIdentifierCensus` green). R9 (no new `-tags=matrix` file; `TestEpic2GateObligationsMatchTheDeclaredSet`
green). Diagnostic registry hygiene (`allCodes`, independent pin literal, `TestRegistryIsAdditiveOnly`
unweakened). R7 forward-compatibility: the per-`(table, page)` displacement channel is the right shape
for 4.5's footer aggregate and 4.6's over-tall row — neither should need to tear it out.


---

## Finisher Resolutions

Every finding below was triaged and, where FIX, closed with a real code/test change, red-proofed
against the reviewer's own mutation (or a new mutation, named, where the reviewer did not supply one
directly reproducible from the story file alone). Every red-proof ran the WHOLE `folio-go` suite
(`env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -v .`,
never behind `-run` for the mutated run), restored by hand from a `cp` backup, confirmed identical via
`/usr/bin/diff` afterward.

| # | Severity | Decision | Rationale |
|---|---|---|---|
| 1 | Blocker | **FIX** | Wrote both heavy test bodies for real (see "Heavy/integration tests" above), gated on `FOLIO_HEAVY=1` (an ordinary env var, not a build tag — preserves R9). Manually confirmed both PASS with the var set. |
| 2 | Blocker | **FIX** | Added `TestSiblingPositionUnaffectedByTableHeaderRepeatThroughPageModel` (`table_header_repeat_test.go`), exercising `buildPageModel`/`paginateDocument` (not layout.Paginate alone) via a differential Y comparison between a FEW-row and a MANY-row render of `twoTablesOnSamePageDoc`. Reproduced the reviewer's RP-C (all three render.go guards forced `true`) against the WHOLE suite: reddens this test **only** (1015·1·3), confirming both the fix's correctness and the new test's discrimination. |
| 3 | Blocker | **FIX** | Extended `TestReservedHeaderHeightIsAccountedForOnEveryContinuationPage` with a by-value position pin: the repeat's rendered bottom (from `HeaderRepeats[].Shift` and the header's own measured extent) must not exceed the first data row's rendered top, and the repeat's rendered top must not sit above the content band's own top. Reproduced RP-B (`Shift` displaced by 37,000mp): reddens on exactly this new assertion. |
| 4 | Major | **FIX** | Added `TestSuppressedHeaderRepeatDiagnosticReachesResultThroughRender`, rendering `tallRowRepeatDoc` through `Render()` and asserting `Result.Diagnostics` carries `DiagCodeTableHeaderRepeatSuppressed` at `SeverityWarning`, naming `e1`, with a message naming all three levers. Reproduced RP-D (the whole `for _, s := range plan.Suppressed` block neutered): reddens this test only. |
| 5 | Major | **FIX** | Added `TestRepeatedHeaderNeverCaptionsZeroRowsWhenRepeatsActuallyOccur` (DECISION-1's FORBIDDEN half, on the AC1/AC2 fixture where repeats demonstrably occur, with an explicit presence precondition) and `TestWidowHeaderWithZeroRowsIsAllowed` (the ALLOWED half, positively witnessed on `tallRowRepeatDoc`'s own page 0). |
| 6 | Major | **FIX** | Corrected the Delivery Log's gate table and its own narrative (both were wrong: `-skip` excludes the named test entirely, and two of this story's own new tests were the actual skips) — see "Baseline reproduction" table's note and the correction paragraph following it. |
| 7 | Major | **FIX** | Corrected the AC7 per-AC verification row to cite the tests that actually re-render (`multi_page_fixture_test.go`, `fixture_test.go`, `wrapped_text_fixture_test.go`) and to describe `byte_neutrality_test.go` as the digest-drift guard (D-000.47) it is, not a re-render check. |
| 8 | Major | **FIX** | Real correctness bug, fixed in `internal/layout/paginate.go`: `TableHeaderSuppressed.Available` now records `windowStart + height - effectiveTop` (the room the row actually had) instead of the page's bare content height; added a `HeaderHeight` field so the diagnostic message can name a number for the "reduce headerHeight" lever. Added `interleavedSuppressionDoc`/`TestReservedHeaderSuppressedDiagnosticReportsTheRoomTheRowActuallyHad`, a NEW fixture engineered so an unrelated element's header chrome sets a continuation page's window before the tall row does (Available: 107,000mp vs the pre-fix 110,000mp). Reproduced the pre-fix formula as a mutation: reddens this test only. |
| 9 | Minor | **FIX** | `runCarriesAHeaderLabel` now matches ONE direction only (`containsSubstring(text, l)`); added `TestRunCarriesAHeaderLabelRejectsDataCellText`, a self-check asserting every data-cell/footer literal this file's fixtures produce (including bare `"a"`/`"o"`) is rejected, and every header label (plus a prefixed/suffixed variant) is accepted. |
| 10 | Minor | **FIX** | Corrected the doc comment (runs are per-CELL, measured directly, not per-glyph) alongside Finding 9's fix — same comment, same edit. |
| 11 | Minor | **FIX** | Collapsed `rowsPerFirstPage`/`rowsPerContinuationPage` to one name, `rowsPerPage`, with a comment explaining why the two capacities coincide on this fixture and flagging that 4.5/4.6 may need to re-split them. |
| 12 | Minor | **FIX** | Named the owning story for AC7's declined deletion (the three re-render tests cited above, all pre-existing) in the corrected AC7 row; AC5's own two clauses already named their owners (`layout.BandContent`/pre-Epic-4 and Story 4.3) and were left as the developer wrote them. |
| 13 | Minor | **FIX** (documentation) | Added a doc-comment note to `TestHeaderRowMovesWholeToNextPageThenRepeatsOnEveryLaterOne` naming the two layout-package siblings (`TestPaginateHeaderGroupMovesWholeEvenWhenChromeAndLabelExtentsDiffer`, `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder`) that actually carry the atomicity property under RP-E, so a future reader does not mistake this fixture's two halves for the property's sole guard. No fixture change: manufacturing divergent chrome/label extents on THIS fixture would itself be a "fix the fixture to fit the test" move the reviewer's own Finding 13 did not ask for, and the property already has real, reddening teeth elsewhere. |
| 14 | Minor | **DEFER** | Pre-existing hole (`TestDiagCodeBridgePreservesExactStrings` hardcodes two codes, unextended since Stories 2.8/3.3), explicitly named by the reviewer as "not closed by this story." Fixing it means replacing a hardcoded pair with an enumeration over `diag.All()` — a cross-cutting change to a shared bridge-completeness guard, appropriately scoped to its own follow-up rather than folded into a Story-4.4 commit. **Owner**: next story touching `diag_bridge_test.go`, or a standing Epic 4 boundary-gate item. |
| 15 | Nit | **DISMISS** | The reviewer's own suggested resolution says "None required for this story," consistent with existing precedent (`OverflowError.Error()` uses the same millipoint spelling). No action taken. |

**Story status recommendation from the review was "Changes Requested."** Every Blocker and every
Major is now FIX, red-proofed, and the whole suite is green (see gate figures below) — the story is
promoted to `done`.

### Post-fix gate re-measurement (finisher's own run)

| Gate | Command | Measured |
|---|---|---|
| `folio-go` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -v ./...` | **1022 pass · 0 fail · 3 skip** (+6 tests over the developer's 1016: Findings 2/4/5(×2)/8/9's new tests; `TestXrefEntriesRejectsMalformedSubprocess` + this story's two `FOLIO_HEAVY`-gated heavy tests are the three skips, named and explained above) |
| `lint` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **115 pass · 0 fail** (unchanged) |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **3 pass · 0 fail** (unchanged) |

`gofmt -l folio-go lint hashmatrix` empty. `go vet ./...` clean in all three modules. `go build
-tags=matrix ./...` and `go vet -tags=matrix ./...` clean in `folio-go`. `git status` carries no
`testdata/`/`fixtures/` modification — **no golden moved**. `TestTableInPageHeaderRepeatsIdenticallyAcrossPages`
confirmed green; `table_render_test.go`/`table_render.go` carry no diff (untouched by this finisher
pass, as by the original story). `FOLIO_HEAVY=1 go test -run
'TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable|TestTwoTablesWithPageCountFooterRenderConsistently'
-v ./...` run once by hand: both PASS.

### Files touched by this finisher pass

- `folio-go/internal/layout/paginate.go` — Finding 8's correctness fix (`Available`/`HeaderHeight`).
- `folio-go/render.go` — Finding 8's message improvement (names the header's own reserved height).
- `folio-go/table_header_repeat_test.go` — Findings 1, 2, 3, 4, 5, 8, 9, 10, 11 (new tests, fixtures,
  and corrected doc comments/matcher).
- `folio-go/table_pagination_test.go` — Finding 13 (doc-comment note only).
- `_bmad-output/implementation-artifacts/4-4-repeat-the-table-header-on-every-continuation-page.md` —
  Findings 6, 7, 12 (Delivery Log corrections), this Finisher Resolutions section, the plain-terms
  opener, and the status field.
