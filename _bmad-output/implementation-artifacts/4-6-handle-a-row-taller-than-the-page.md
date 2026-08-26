# Story 4.6: Handle a row taller than the page

Status: done

Epic: 4 — Tables and pagination
Sprint key: `4-6-handle-a-row-taller-than-the-page`
Baseline: `45cf812` (Story 4.5 finisher), tree clean at creation.
Covers: FR25, FR41 · AD-14, AD-24, AD-13, AD-5, AD-1

---

## In plain terms (read this first if you just want the gist)

A table row is as tall as its contents make it. If one record carries a wall of text, that row can
end up taller than the whole printable area of a page, and no page could hold it. Folio used to
refuse: the render stopped and no document came out at all. One bad record in a hundred thousand
took down an entire statement run.

That is fixed. Such a row now starts at the top of a fresh page, is drawn as far down as the page
has room for, and the rest is left out; the render finishes and produces a document. A warning travels
back alongside the bytes naming the table, which row was cut, how tall it was, and the height it was
measured against.

Review changed one thing substantially. As first written, the cut page quietly lost its repeated
column headers — the one page in a document that can least afford it, since a trimmed row is hard
enough to read with its labels intact. The headers now repeat, and the row is cut to fit beneath
them. Only when repeating them would leave room for not even one line are they dropped — and
that is reported too.

What this deliberately does not do: it does not grow the page, shrink the text, or continue the row
onto the next page. A picture sized past the page, or a stray line of text outside any table, still
stops the render with an error. That asymmetry is the point — a row's height comes from data the
author never saw; a picture's size is something the author typed.

What will look wrong later and is not: several older tests that insisted this case must fail were
rewritten here on purpose, as earlier stories said they would be. One test elsewhere fails by
design and belongs to different work; three slow tests run only on request. And two of this story's
own acceptance criteria claim two independently checkable outcomes each where only one really is —
recorded rather than reduced, and left for review to settle.

---

## Story

As an integrating Go developer,
I want a pathological row to produce a diagnostic rather than an infinite loop or a dead render,
So that one bad record cannot take down my statement service.

---

## Premise verification (D-000.81 / this story's creator as reviewer of its own brief)

The brief asked to be checked rather than complied with. Four things are confirmed, three are
corrected or extended. Everything below was measured at `45cf812`, with throwaway probes that
were deleted; the tree is clean.

### Confirmed

1. **`TestOverTallSingleRowStillOverflows` exists** (`folio-go/table_footer_test.go:1365`) and
   asserts `*layout.OverflowError`. It passes at HEAD.
2. **Mechanism is probe-then-merge at the `folio` package level.** `layout.Paginate` has exactly
   three call sites, all inside `paginateWithFooterOrphanFix` (`folio-go/table_footer.go:183`,
   `:222`, `:252`). Pass 1 at `:183` runs on unmerged items, so an over-tall row's error escapes
   before any merge is attempted. `internal/layout/paginate.go`'s general behaviour is untouched
   by 4.5.
3. **HEAD's behaviour for this story's subject is a located error, and the render produces
   nothing.** Measured through the public `Render`, on `tooTallRowDoc()`:
   `Result.Bytes` length **0**; error `*folio.RenderError` with `Diagnostic.Code =
   "CONTENT_UNLAYOUTABLE"`, wrapping
   `*layout.OverflowError{ElementID:"e1", Kind:"line", ItemHeight:272400, ContentHeight:110000}`.
4. **The footer-alone-too-tall case was routed to this story, not closed.**
   `TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs`
   (`folio-go/table_footer_test.go:1316-1324`) says so in its own words: D-4.5.2's two acceptable
   answers were "route it to FR44/2.8's clip" or "declare it 4.6's subject"; 4.5 took the second,
   guaranteeing only termination via the existing `OverflowError`. **It is inside this story's
   scope.** `internal/layout/paginate.go:753` agrees independently: *"Story 4.6 owns clipping this
   case to a fresh page; this story changes nothing about it."*

### Corrected — the brief's characterisation of the fence is narrower than the fence's stated intent

D-4.5.4's invariant reads *"an over-tall **group** which is not a footer-plus-row still produces
`OverflowError`"*. The test that discharges it constructs an item with **no `Group` field at all**:

```go
items := []layout.ColumnItem{
    {ElementID: "e9", Top: 0, Bottom: 999_999_000, Runs: []layout.TextRunRef{0}},
}
```

That is an **ungrouped** item. The fence as shipped guards the ungrouped path, which is not the
path its own words name. It did red-proof (widening the bypass reddens it), so it is not hollow —
but it is not a test of grouped over-tall items, and a reader taking D-4.5.4 at its word would
believe otherwise. **Consequence for this story, and it is a favourable one:** under the
recommended scope below this test stays green **unchanged**, because the ungrouped path is not
what this story changes. Flagged rather than reconciled.

### Extended — the brief omits the one thing that settles the whole story, and HEAD violates it

**AD-14, verbatim** (`ARCHITECTURE-SPINE.md:311`):

> Over-tall rows (FR25) and clipped content (FR44) are `Warning`s returned alongside PDF bytes,
> never silent and never fatal.

The brief invited a design fork on "clipped, split, allowed to overflow, or errors". **There is no
fork on the over-tall table row.** The spine already rules it, the epic's ACs already state it, and
**HEAD is currently in violation**: the case returns an `Error` with zero bytes. This story is
bringing the code to the spine, not choosing between arms. The fork that *does* exist is one of
scope, and it is framed below.

### Extended — the located error does not name the row, and calls a table row a line

Two defects inherited with the pin, both measured:

- **`OverflowError.ElementID` is `"e1"` — the table's id, not the row's.** There is no row index
  anywhere in the error type. The epic's AC requires the diagnostic to name *"the row and the
  element id"*. That is achievable but is **new data**, not a relabelling: the row index lives in
  `layout.ItemGroupKey.Index` and must be carried into the diagnostic.
- **`Kind` is `"line"`, not `"table"`.** Story 4.3 routed this to us in writing
  (`table_pagination_test.go:660`: *"the mislabelling … predates this story and belongs to 4.6"*).
  The cause is ordering: the shipped path appends text before rects, the row's first line ties the
  chrome on `Top`, and the stable sort visits the line first.

---

## The measurement that shapes the implementation

**`Kind` cannot discriminate a table row from a plain text element, and `Group.Present` can.**

Measured through `Render` at HEAD, on two documents:

| subject | `Result.Bytes` | error code | `OverflowError.Kind` |
| --- | --- | --- | --- |
| table whose only data row is 272,400mp tall | 0 | `CONTENT_UNLAYOUTABLE` | `"line"` |
| plain text element 272,400mp tall, outside any table | 0 | `CONTENT_UNLAYOUTABLE` | `"line"` |

The two are **indistinguishable at the public API today**. An implementation that keys its clip on
`Kind == "table"` will miss the shipped path entirely and clip nothing.

The grouping, measured on the same over-tall-row document via the real collectors:

```
run[0]  el="e1" top=10000 bot=20000   group={Present:true Key:{ElementID:e1 IsHeader:true  Index:0}}
run[1]  el="e1" top=20000 bot=292400  group={Present:true Key:{ElementID:e1 IsHeader:false Index:0}}
rect[0] el="e1" top=10000 bot=20000   isData=false group={Present:true Key:{ElementID:e1 IsHeader:true  Index:0}}
rect[1] el="e1" top=20000 bot=292400  isData=true  group={Present:true Key:{ElementID:e1 IsHeader:false Index:0}}
```

So: the over-tall row's items **are** grouped, and `Key.Index` **is** the row index the epic asks
for. A plain text element's items carry the zero-value `ItemGroup` (`render.go:222`, the `default:`
arm). **`ItemGroup.Present` is the discriminator; `Kind` is not.**

Corollary, verified: every `tableRectSource` is constructed at exactly one of three sites —
`table_render.go:707` (header), `:931` (data row), `:1113` (footer) — and `chromeRowGroup` maps all
three to a present group. **No table rect is ungrouped.** Under the recommended scope, `Kind ==
"table"` therefore becomes unreachable from package `folio`, though still constructible by a direct
caller of `internal/layout`. See Task 8.

---

## Decision to surface — the scope of the clip (arms + recommendation)

D-4.5.2's own generalisation applies to us: *a relocation arm needs a floor case*. Ours is: **which
over-tall items clip, and which keep erroring?**

**Arm A (recommended) — clip only items whose `ItemGroup.Present` is true.**
Table header rows, data rows and footer groups clip. Ungrouped items — a plain text element's line,
an image's declared box — keep `layout.OverflowError` exactly as today.
*What changes:* the three tests listed in "blast radius" invert. `TestOverTallSingleRowStillOverflows`,
`internal/layout/paginate_test.go`'s line/image table rows, and `render_overflow_test.go` all stay
green **unchanged**. D-2.6.1's image ruling (*"a template error … rather than a render-time surprise"*)
survives. AD-14's own wording is honoured literally: it scopes the never-fatal guarantee to
*over-tall **rows** (FR25)*, and FR25 is tables.

**Arm B — clip every over-tall item.**
*What changes:* D-2.6.1 is reversed; `OverflowError`'s doc contract (*"NEVER A STRADDLE AND NEVER A
SILENT CLIP: both are what this error exists to prevent"*) is reversed; `OverflowError` becomes
unreachable and its type, its `Kind` field, its four tests and `CONTENT_UNLAYOUTABLE`'s overflow arm
all become dead. D-4.5.4's fence reddens and must be deleted one story after it was placed. An
author who typed a 400pt image box gets a silently truncated picture instead of being told the box
is wrong.

**Recommendation: Arm A.** The distinction is not arbitrary — it is AD-13's own line. A table row's
height is **derived from data** and the author may never have seen the offending record; an image
box and a font size are **declared by the author** and are the author's own typo. Deriving a
failure from data that the author cannot audit in advance is precisely the case that must not be
fatal; a typo is precisely the case that should be.

**The honest cost of Arm A, stated rather than hidden:** the two symptoms are byte-identical at the
public API today (the table above). An author with a too-tall row gets a document and a warning; an
author with a too-tall paragraph gets no document and an error; and nothing in the error explains
which side of the line they landed on. This story does not change the ungrouped message. If the lead
prefers, a sentence naming the asymmetry could be added to `OverflowError.Error()` — that is a
separable and cheap follow-on, deliberately not taken here.

**Flagged for the lead.** If Arm B is ruled, ACs 1–4 are unchanged in substance but AC7's
byte-neutrality claim, the blast-radius table and Task 8 all change, and D-4.5.4's fence must be
retired with an entry saying so.

---

## Decision to surface — the diagnostic code (D-4.5.1's test applied, not copied)

D-4.5.1's discriminator, applied rather than a precedent copied in either direction:

> Two conditions share a code only if the author would take the same action **AND** the same thing
> happened to their document. Same remedy is not sufficient.

**Against `TEXT_CLIPPED_WIDTH` (Story 2.8) — different.** D-2.8.1 scoped that code to the
**horizontal** axis explicitly, and `render.go:809` records that the check *"never reads
el.Height"*. Ours is vertical, at a page edge rather than a box edge, and it destroys **whole lines
of a row** rather than the tail of a line. The investigations differ: a clipped-width author sees
text stop mid-word inside a visible box and knows the box is too narrow; a clipped-height author
sees a row end at the bottom of a page, where a missing continuation is exactly what a normal page
break looks like. **Do not widen.**

**Against `TABLE_HEADER_REPEAT_SUPPRESSED` (4.4, drops) and `TABLE_FOOTER_ORPHAN_SUPPRESSED` (4.5,
relocates) — different.** This condition **destroys** content. Nothing is dropped-but-recoverable
and nothing is moved; bytes that the data contained are absent from the document forever. That is a
third thing, and it is the one of the three that a reader can suffer without noticing. **Mint.**

**Recommended code: `TABLE_ROW_CLIPPED_HEIGHT`**, minted in `internal/diag/diag.go` and bridged as
`folio.DiagCodeTableRowClippedHeight`, following the established two-constant pattern and D-1.4.2
(mint where the condition ships, never ahead of it).

**And the test cutting the other way, which is the part that proves it is a test and not a reflex:**
the **footer-alone-too-tall** case (routed to us by 4.5) and the **over-tall header row** case
**share this code**. Same thing happened — a table group taller than the window, clipped, content
destroyed — and same action — shorten it, or enlarge the page. Both limbs of D-4.5.1 are satisfied,
so one code covers all three group roles, with the role and the row index carried in the message.
A header row and a footer row are rows in domain language and in FR25's own wording.

---

## Governing decisions read

D-4.5.1 (code discriminator), D-4.5.2 (the floor case; routes the footer-alone case here),
D-4.5.4 (the fence, and its narrower-than-stated discharge), D-4.5.5 (a default needs a
discriminating fixture), D-000.79 (whoever names the mutation must not have written the code;
Class A/B), D-000.80 (part (a): remove the accidental cause, the test must still pass),
D-000.82, D-000.83 + amendment (report the set, never the scalar), D-000.84 (the four-story test,
concluded "partially"), D-000.85 (the per-observable screen; **this story is n=1 of the new
pre-registered count**), D-000.67 part 2 + amendment (a mechanism carries more than one presence
precondition), D-000.68, AD-1, AD-5, AD-13, AD-14, AD-23, AD-24.

---

## Acceptance criteria

Every output-producing AC carries its **observable count** and a **named deletion-mutation per
observable** (D-000.85). Where the mutation list is shorter than the observable count, that is a
defect in this story, visible here rather than at review.

### AC1 — the render completes and returns bytes

**Given** a table whose single data row is taller than the content window
**When** `Render` runs
**Then** it returns a nil error and a non-empty `Result.Bytes`
**And** the returned bytes are a structurally valid PDF (the existing structural-validity helper).

- **Observables: 2** — (i) the error is nil where HEAD returns `*RenderError`; (ii) the byte slice
  is non-empty and valid where HEAD returns length 0.
- **Deletion (i):** remove the clip branch so the over-tall group falls through to the existing
  `OverflowError` return. AC1's error assertion reddens.
- **Deletion (ii):** keep the clip branch but have it drop the whole group's refs from every page
  (clip to zero height rather than to the window). The error stays nil and AC1(i) stays **green**;
  the structural-validity/non-empty assertion must redden on its own. That green is the
  discrimination — if (i) also reddens here, the two observables are not separated.
- **D-000.80 part (a):** not accidentally true at HEAD. Measured directly: HEAD returns 0 bytes and
  a `*RenderError`. No accidental cause to remove.

### AC2 — the clipped row starts at the top of a fresh page, and the partition is asserted as a set

**Given** a table whose rows fill one page and whose next row is taller than the content window
**When** pagination runs
**Then** the over-tall row's chrome and lines are assigned to a page containing **no other table
row**, and that page is **later** than the page holding the preceding rows
**And** the assertion names **which row indices land on which page**, never a page count or a total.

- **Observables: 2** — (i) the over-tall row's page contains no other row (it is alone); (ii) the
  preceding rows are undisturbed and still land where they did (the clip does not reflow the rest
  of the document).
- **Deletion (i):** remove the "start a fresh page for the over-tall group" step and let the group
  be clipped in place at whatever offset the window had reached. The alone-on-its-page assertion
  reddens; (ii)'s untouched-predecessors assertion stays **green**.
- **Deletion (ii):** keep the fresh page but also advance the window for every *subsequent* item by
  the group's full untruncated height. (i) stays green; the predecessor/successor partition
  reddens.
- **D-000.68 / D-000.83:** the criterion is the **set** `{page → row indices}`, not a page count. A
  wrong arrangement produces the same count. Do not write `len(pages) == 2`.
- **D-000.80 part (a), and this one is load-bearing:** "at the top of a fresh page" is
  **accidentally true** for any fixture whose over-tall row is the first thing in the content band —
  `tooTallRowDoc()` is exactly that shape, and page 0 is always a fresh page. **Part (a): the
  fixture must place at least one full page of ordinary rows before the over-tall row**, so that
  "fresh page" means page N>0 and is not satisfied by "page 0 exists". With that fixture, remove the
  fresh-page step and the test must redden; restore it and, with the *preceding rows deleted from
  the data* (the accidental cause removed), the test must **still pass**.

### AC3 — the clip keeps what fits and drops what does not, whole lines only

**Given** an over-tall row whose lines are individually shorter than the content window
**When** it is clipped
**Then** every line whose extent lies entirely within the page's content window is present on that
page, and every line whose extent crosses or lies beyond the content bottom is absent
**And** no line is drawn partially — never a straddle
**And** the row's chrome rectangle is truncated at the content bottom rather than drawn to its
untruncated height.

- **Observables: 3** — (i) the kept-line set; (ii) the dropped-line set, including the straddling
  line; (iii) the chrome rect's bottom coordinate.
- **Deletion (i):** drop every line of the group, keeping only the chrome. (ii) and (iii) stay
  green; the kept-set assertion reddens.
- **Deletion (ii):** keep every line of the group, including those beyond the content bottom. (i)
  and (iii) stay green; the dropped-set assertion reddens.
- **Deletion (iii):** leave the chrome rect at its untruncated `bottom`. (i) and (ii) stay green;
  the chrome-extent assertion reddens on its own.
- **D-4.5.5, and this is the story's own discriminating-fixture obligation.** `tooTallRowDoc()`'s
  row is **one physical line 272,400mp tall** against a 110,000mp window. Clipping it keeps **zero**
  lines — so against that fixture alone, "clip to the window" and "drop all the row's text" render
  **identically**, and a wrong clip boundary is invisible. **The story must add a fixture whose
  over-tall row has several lines and whose window falls between two of them**, so the kept/dropped
  partition is a proper non-empty split of a non-empty set. Name the boundary line explicitly in the
  test. The single-line fixture is kept as the degenerate case, not as the only case.

### AC4 — a located Warning naming the table and the row, on the bytes channel

**Given** the clip has occurred
**When** the render returns
**Then** `Result.Diagnostics` carries exactly one `Diagnostic` for it, with `Severity ==
SeverityWarning`, `Code == DiagCodeTableRowClippedHeight`, `ElementID` naming the table, and a
message naming **the row's index within the bound collection**, the row's own height, and the
content height it was measured against
**And** the diagnostic travels on `Result.Diagnostics`, never wrapped in a `*RenderError` (AD-14,
D-3.6.3 arm A).

- **Observables: 3** — (i) a Warning with this code exists at all; (ii) it carries the row index
  (data HEAD's error type does not have); (iii) its severity/channel is Warning-on-`Diagnostics`,
  not Error-on-`RenderError`.
- **Deletion (i):** delete only the `Diagnostic` construction, leaving the clip itself intact.
  **AC1, AC2 and AC3 must all stay green**; only AC4 reddens. This is the Story 4.5 finisher's
  demonstrated shape and the reason D-000.85 exists — if a layout AC also reddens here, the
  observables have not been separated.
- **Deletion (ii):** construct the Warning with the code and element id but format the message from
  the element id alone, omitting the row index (i.e. do not thread `ItemGroupKey.Index` through).
  (i) and (iii) stay **green**; only the row-index assertion reddens. This one matters: the row
  index is the epic's explicit requirement and is the observable most likely to ride along
  unwitnessed.
- **Deletion (iii):** emit the same Diagnostic at `SeverityError` wrapped in a `*RenderError`. (ii)
  stays green in content; (i)'s presence-on-`Diagnostics` and (iii) redden.
- **D-000.80 part (a):** not accidentally true — no code with this name exists at HEAD
  (`internal/diag/diag.go`'s registry, checked).
- **AD-14 registry obligations:** the new code is additive, appears in `internal/diag`'s exported
  code slice, and is asserted as a **literal string** in `internal/diag/diag_test.go`'s table, per
  the existing bridge convention (a bridge that alters one byte is a breaking change wearing a
  refactor).

### AC5 — the footer-alone-too-tall case, routed here by 4.5, is closed under the same code

**Given** a table whose footer group is by itself taller than the content window
**When** pagination runs
**Then** it is clipped by the same mechanism, on its own page, and reports
`TABLE_ROW_CLIPPED_HEIGHT` with the group identified as the footer
**And** `TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs` is **rewritten** to assert
this, not deleted, and its comment records that D-4.5.2's deferral is now discharged.

- **Observables: 2** — (i) the footer group clips rather than erroring; (ii) the diagnostic
  identifies it as the footer group rather than as data row index -1 leaking the sentinel.
- **Deletion (i):** restrict the clip branch to `Group.Key.Index >= 0`. The footer case reverts to
  `OverflowError` and (i) reddens; the data-row ACs stay **green**.
- **Deletion (ii):** format the message with the raw `Index` value. (i) stays green; the
  footer-naming assertion reddens on the sentinel `-1` appearing in author-facing text.
- **Note:** `footerGroupIndex` is `-1`. A message that prints it verbatim is a wire value reaching a
  reader. Render the role, not the sentinel.

### AC6 — termination, bounded

**Given** any of the over-tall fixtures
**When** pagination runs
**Then** it returns within a bounded wall-clock limit — never a hang, never an unbounded loop
**And** the bound is enforced by the test, following 4.5's `select`/`time.After` precedent.

- **Observables: 1** — the call returns. (A hang is not a red; it is a stuck test. The bound is what
  makes it an observable at all.)
- **Deletion:** make the clip branch `continue` without advancing the window or marking the group
  placed, so the sweep revisits it. The bounded-return assertion reddens as a timeout **failure**,
  not as a hung suite.

### AC7 — no existing golden moves

**Given** the full byte-neutrality and matrix suites
**When** they run at this story's HEAD
**Then** every recorded digest is unchanged.

> **AC7's stated reason was corrected at finish** (this story's reviewer, Finding 10, verified
> independently). It read *"no document in the corpus contains an over-tall group, so this story's
> branch is not reached by any of them."* The true statement is **wider and worse**: **no document in
> the recorded corpus contains a table at all.** None of the nine committed golden PDFs (`font-text`,
> `image-embed`, `minimal-rect`, `multi-page`, `multi-script-fallback`, `page-count-20`,
> `shaped-text`, `three-band-page`, `wrapped-text`) has a `"type": "table"` element; the only one
> under `testdata`/`fixtures` is `folio-go/testdata/template/golden/worked-example.json`, which is a
> **template** round-trip golden, not a rendered-PDF golden.
>
> So AC7 is not a measurement of this story's branch being unreached — it is a measurement of a
> corpus that **cannot exercise any of Epic 4**, five stories' worth. The AC still holds and its
> deletion screen below still discriminates; what changes is that the recorded reason now matches the
> measured one. Routed to Story 4.7 as D-4.7.0's three obligations.

- **Observables: 1** — the digest set.
- **Deletion:** make the clip branch fire unconditionally rather than only on `itemHeight >
  height`. The byte-neutrality suite reddens across many documents. This is the presence check that
  the neutrality claim is a measurement and not an absence of coverage.
- **This is a Class-B-shaped guard by construction** (it asserts the *absence* of an effect). Record
  it as such in the Class A ledger rather than counting it as a witness.

### AC8 — the inherited `Kind` mislabelling is discharged by construction, and said so

**Given** the clip covers every grouped item and every table rect is grouped (measured: three
construction sites, all grouped)
**When** an over-tall item still reaches `OverflowError`
**Then** it is ungrouped, and `Kind` is therefore correct for it — `"line"` for a line, `"image"`
for an image
**And** the story records that `Kind == "table"` is no longer reachable from package `folio`, in
`OverflowError`'s own doc comment, so a future reader does not treat a dead branch as live.

- **Observables: 1** — the doc-comment/reachability record is prose and **has no deletion screen.**
  Stated rather than papered over (D-000.84's named third class). It is the reviewer's catch.
- **Do not** write a comment-only `if` block for this (D-000.84, Story 4.5's AC3 shape). If it
  cannot be asserted, it is a comment on an existing declaration, not a statement in the code path.

---

## Blast radius — the existing assertions this story inverts

Reported as a **set**, not a count (D-000.83 amendment).

**Must be rewritten (grouped subjects):**

| test | file | what changes |
| --- | --- | --- |
| `TestPaginateGroupTallerThanWindowReturnsLocatedOverflow` | `internal/layout/paginate_group_test.go:333` | asserts `*OverflowError` Kind `"table"` for a grouped item; becomes the clip assertion |
| `TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow` | `table_pagination_test.go:570` | PHASE B arm |
| `TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender` | `table_pagination_test.go:646` | the public-`Render` arm 4.3 named as *"the pin 4.6 must build its clip-and-diagnose behaviour against"* |
| `TestFooterAloneTooTallForTheWindowTerminatesRatherThanHangs` | `table_footer_test.go:1325` | AC5 |

**Must stay green, unchanged (ungrouped subjects) — verify, do not assume:**

| test | file |
| --- | --- |
| `TestOverTallSingleRowStillOverflows` (D-4.5.4's fence — ungrouped item, see Premise Verification) | `table_footer_test.go:1365` |
| the line/image rows of `internal/layout/paginate_test.go`'s overflow table | `internal/layout/paginate_test.go:325` |
| `render_overflow_test.go`'s `Render`-level line/image cases | `render_overflow_test.go:150` |
| `render_error_test.go`'s `CONTENT_UNLAYOUTABLE` case | `render_error_test.go:205` |

Any redness in the second table under Arm A is a scope violation, not a fixture problem.

---

## Tasks

1. [x] **Re-measure the baseline before writing code.** Reproduce the two probes recorded above (the
   over-tall row and the over-tall plain text element through `Render`) and confirm the HEAD figures
   in Premise Verification. Delete the probes. If any figure differs, stop and report — do not build
   on this story's stated baseline without re-confirming it.
2. [x] **Mint the code.** `internal/diag/diag.go`: add `CodeTableRowClippedHeight Code =
   "TABLE_ROW_CLIPPED_HEIGHT"`, add it to the exported registry slice, add its literal row to
   `internal/diag/diag_test.go`'s table. Bridge it as `folio.DiagCodeTableRowClippedHeight` in
   `diagnostic.go`, following the existing constants' doc-comment convention including the
   additive-only AD-14 note.
3. [x] **Carry the clip decision in `layout.Pagination`, following 4.4's `Suppressed` precedent
   exactly.** Add `Clipped []TableRowClipped` carrying at minimum: `ElementID`, the group key (so
   the row index and header/footer role are available without re-derivation), `ItemHeight`,
   `ContentHeight`, and the page it landed on. The **decision** lives in `internal/layout`; the
   **message** is built in package `folio` (`render.go`'s existing `Suppressed` loop is the model).
   Do not build a message string inside `internal/layout`.
4. [x] **Implement the clip at `internal/layout/paginate.go:753`.** Replace the unconditional
   `OverflowError` return with: if `it.Group.Present`, start a fresh page for the group, assign the
   group's refs whose extents fit within that page's window, omit the rest, truncate the chrome
   rect's bottom to the content bottom, record a `TableRowClipped`, and advance. Otherwise return
   `OverflowError` as today. Key the branch on `Group.Present` — **not** on `Kind`, for the reason
   measured above.
5. [x] **Build the Warning in `render.go`,** alongside the existing `Suppressed`/`footerOrphanDiags`
   loops, naming the table, the row index (or the header/footer role), the row's height, the content
   height, and the same three author-facing levers the sibling messages use (font size / cell
   padding, margins, page-header/page-footer heights) — D-000.37's "executable by a human".
6. [x] **Write the multi-line discriminating fixture** required by AC3 and the
   preceding-full-page fixture required by AC2's part (a). These are the two fixtures without which
   this story's central assertions are vacuous; write them before the assertions.
7. [x] **Rewrite the four tests in the blast-radius table**, each with a comment naming this story and
   the earlier story that placed it, so the inversion reads as planned rather than as a regression.
8. [x] **Record the reachability change** in `OverflowError`'s doc comment: amend *"NEVER A STRADDLE AND
   NEVER A SILENT CLIP"* to say that the clip introduced here is neither silent (it carries a
   Warning) nor applicable to ungrouped items, and note that `Kind == "table"` is no longer produced
   from package `folio`. Amend the `"ONE OVERFLOW RULE, TWO SUBJECTS"` paragraph in the same pass.
9. [x] **Run the full suite plus the byte-neutrality and matrix suites.** Report every figure with its
   units and, where the set is small, report the **set** (D-000.83 amendment). Report skips **by
   name** (D-000.81).
10. [x] **Report the Class A ledger split by sub-cause** — compound-observable / prose-claim / other —
    per D-000.85. A flat count is not acceptable. This story is **n=1 of a new pre-registered
    series**; report the number that occurs, not a favourable one.
11. [x] **Do not commit.** This story ends at status `review`.

---

## Dev notes

- **AD-1.** `internal/layout` is render path: no `time`, no map ranged into output. The clip's
  record slice is appended in sweep order — a slice, never a map ranged for emission (R5, the same
  rule `Suppressed` follows).
- **AD-5.** The clip is a page-model decision. `internal/layout` must not learn anything about PDF.
  Truncating a chrome rect's bottom is a coordinate change on an existing rect; **no PDF clip path
  is needed and none should be added.** Dropping a line is omitting its ref from
  `PageAssignment.ContentRuns`, which is already a per-page subset.
- **AD-24.** The content band does not grow, nothing negotiates, and a line is never split. The
  clip drops whole lines; it never draws half of one.
- **AD-23.** Nothing here touches money-path numbers. Heights are `geom.Length` fixed-point
  throughout; do not introduce a float to compute a proportion of a row.
- **D-000.79.** Whoever names the mutation must not have written the code. The mutations above were
  named at creation, before any implementation existed. If the implementation's shape makes one of
  them impossible to perform as written, **say so and name a replacement** — do not silently
  substitute a mutation that the code you just wrote happens to survive.
- **D-000.67 part 2 / its amendment.** Check every witness the clip mechanism reports, not the one
  the ACs made you think of. If `Clipped` grows a second population (say a per-page tally as well
  as a per-group one), enumerate both.

---

## Delivery Log

> **Provenance, stated first (D-000.12/13, and the D-67.8.5 three-tier precedent).** This section did
> not exist at review — Tasks 9 and 10 were ticked with nothing behind them, which is this story's
> Finding 5. **Everything below was measured by the FINISHER, at the finisher's own final baseline,
> and is not a reconstruction of what the developer ran.** Where a figure the developer reported is
> quoted, it is quoted to be corrected. Nothing here is presented as contemporaneous with development.

### Gates

Every figure is `total (pass / fail / skip)` with the invocation beside it (D-000.89). All runs
`CGO_ENABLED=0 GOWORK=off`, from `folio-go/`, counts taken from raw `test2json` (`go test -json`),
never from a wrapper's summary line.

| invocation | commit / state | top-level | incl. subtests |
| --- | --- | --- | --- |
| `go test -count=1 ./...` | `45cf812` (baseline) | **704** (699 / 1 / 4) | **1061** (1055 / 2 / 4) |
| `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | `45cf812` | 703 (699 / 0 / 4) | 1053 (1049 / 0 / 4) |
| `go test -count=1 ./...` | story HEAD, at review | **712** (707 / 1 / 4) | **1082** (1076 / 2 / 4) |
| `go test -count=1 ./...` | **this finisher's close** | **716** (711 / 1 / 4) | **1094** (1088 / 2 / 4) |

**The reported figures were wrong, and the correction is Finding 6's, reproduced here independently.**
The story reported *"707 / 1076 (from 699 / 1055)"*. Those are **`rtk`'s PASSED counts labelled as
TOTALS** — the failures and the skips are exactly what a passed-count hides. And the 1049↔1055 drift
was **not** a subtest-counting convention: Story 4.5's 699/1049 is exactly reproducible **under its
own `-skip` invocation** (verified above, in a detached worktree at `45cf812`), while this story
paired a `-skip` figure (699) with a no-`-skip` figure (1055). `1055 − 1049 = 6` is precisely the six
**passing subtests** of `TestCorpusMeetsP6ExerciseFloors` that `-skip` removes. Same suite, two
invocations, one row.

**The +4 / +12 delta is attributed BY NAME**, by differencing the two runs' test-name sets rather than
inferring from totals:

- **Added (5 top-level):** `TestClippedRowsPageStillCarriesItsRepeatedHeader` (+3 subtests),
  `TestClippedRowSuppressesTheRepeatOnlyWhenItBuysNothingAndRecordsIt`,
  `TestClippedRowsChromeIsTruncatedInTheRenderedDocument` (+5 subtests),
  `TestPaginateOverTallHeaderGroupClipsAndIsRecordedAsTheHeader`,
  `TestEveryPublicDiagCodeBridgeIsPinnedToALiteral`.
- **Removed (1 top-level):** `TestDiagCodeBridgePreservesExactStrings` — renamed, not deleted; the
  census replaces it.

**Skips: the expected set of 4, unchanged, BY NAME (D-000.81).** No new skips.

- `TestXrefEntriesRejectsMalformedSubprocess`
- `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`
- `TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`
- `TestTwoTablesWithPageCountFooterRenderConsistently`

**Red by design, named separately and never folded into the arithmetic:**
`TestCorpusMeetsP6ExerciseFloors` (`internal/text`) — P6g has 7, needs ≥20. It is the 1 fail at
top level and 2 including subtests, at every commit in the table above, and it is the deliverable of
the Thai break-opportunity spike rather than a regression of this story.

**Heavy tests, run both directions** (this story touches `paginate.go` and `render.go`):

| test | `FOLIO_HEAVY` unset | `FOLIO_HEAVY=1` |
| --- | --- | --- |
| `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability` | `--- SKIP` | `--- PASS` |
| `TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable` | `--- SKIP` | `--- PASS` |
| `TestTwoTablesWithPageCountFooterRenderConsistently` | `--- SKIP` | `--- PASS` |

Invocation: `[FOLIO_HEAVY=1] go test -count=1 -timeout 900s -v -run '^(…)$' .`

**Other gates.** `gofmt -l /Users/panitw/Projects/folio/folio-go` — clean, no output.
`go vet ./...` — clean. **Cross-target hash matrix not run — per-epic cadence (D-000.4);** this
story's deliverable is not hash-shaped.

**AC7, confirmed by measurement rather than assumed** (D-4.6.4 predicted the clip's extent change
would be free on hashes and required it be checked): `git status --porcelain fixtures/
folio-go/testdata/` is **empty** — no golden moved. All nine `expected.pdf` digests recomputed live
and unchanged. Every golden/byte-neutrality test green by name, including
`TestGoldenDigestAgreesAtEveryDeclaredSite`, `TestEveryGoldenPDFResolvesItsPageTree` and the seven
per-fixture golden tests. The prediction held.

### The per-observable deletion screen (D-000.85), re-run in full at the final baseline

Method: each mutation applied to a pristine tree one at a time, `go vet ./...` first to reject any
mutation that does not compile, then the whole suite; the tree restored from a pre-saved copy and
**verified by SHA-256** after every run; a final all-files SHA check at the end. Failures are reported
as the **difference** against the baseline failure set, so the red-by-design test never appears as a
result. Tests named below are the ones that changed state — "stayed GREEN" means checked explicitly,
not assumed.

| # | mutation | reddened | stayed GREEN |
| --- | --- | --- | --- |
| AC1 (i) | remove the clip branch; the group falls through to `OverflowError` | 20 entries incl. `TestOverTallRowRendersBytesRatherThanFailing` (both subtests) | — |
| AC1 (ii) | clip to zero height — the group contributes nothing to any page | AC3's `kept lines`, D-4.6.4's, the chrome tests | **`TestOverTallRowRendersBytesRatherThanFailing` — BOTH subtests. See Class A #1.** |
| AC2 (i) | remove the fresh-page step | `TestClippedRowLandsAloneOnAFreshPageAndItsNeighboursDoNotMove`, `TestClippingOneRowPushesNothingOffItsPage`, `TestClippedRowsPageStillCarriesItsRepeatedHeader` | — (see Class A #2) |
| AC2 (ii) | `reserved = hh` → `reserved = 0`: repeat the header but do not narrow the cut by it | `TestClippingOneRowPushesNothingOffItsPage`, AC3's chrome + dropped-set, the render-document test | **`TestClippedRowLandsAloneOnAFreshPageAndItsNeighboursDoNotMove`**, `TestClippedRowSuppressesTheRepeatOnlyWhenItBuysNothingAndRecordsIt`, D-4.6.4's repeat subtest |
| AC3 (i) | drop every line of the group, keep only the chrome | `…TruncatesItsChrome/kept lines`, D-4.6.4's below-the-repeat subtest | the dropped-set and chrome subtests |
| AC3 (ii) | keep every line, including those past the content bottom | `…TruncatesItsChrome/dropped lines including the straddler`, `TestClippingOneRowPushesNothingOffItsPage` | the kept-set and chrome subtests |
| AC3 (iii-plan) | never emit the `RectClip` | `…/chrome truncated at the content bottom` + all three of the new render-document subtests | the kept/dropped subtests |
| **AC3 (iii-render)** | **delete the render-side application of the bound (`render.go`)** | **`TestClippedRowsChromeIsTruncatedInTheRenderedDocument` + 3 subtests** | `…/chrome truncated at the content bottom` (the PLAN-side sibling), `TestClippedRowLandsAlone…`, `TestOverTallSingleRowStillOverflows` |
| AC4 (i) | delete ONLY the `Diagnostic` construction, clip intact | `TestClippedRowReportsALocatedWarningNamingTheRow` (+ presence subtest), `TestRowTallerThanContentWindowIsClippedThroughRender` | **AC1, AC2 and AC3 all green** |
| AC4 (ii) | format the message from the element id alone | `…/the message names the row index`, `…NamesTheRoleNeverTheSentinel/data row` | the presence and channel subtests |
| AC4 (iii) | emit the same Diagnostic at `SeverityError` | `…NamesTheRoleNeverTheSentinel` (all three roles), `…/a warning with this code exists, exactly once` | the row-index subtest's content |
| AC5 (i) | restrict the clip to `Key.Index >= 0` | `TestFooterAloneTooTallForTheWindowIsClippedRatherThanFatal` — **and nothing else** | every data-row AC |
| AC5 (ii) | print the footer's raw `-1` sentinel | `…NamesTheRoleNeverTheSentinel/footer row` — **and nothing else** | the data-row and header-row arms |
| AC6 | `continue` without marking the group placed, so the sweep revisits it | `TestOverTallGroupPaginationTerminatesWithinABound` (both subtests) as a **bounded failure**, not a hung suite, + 21 more entries | — |
| AC7 | make the clip fire unconditionally | **57 entries = 44 top-level + 13 subtests**, all in the table behaviour suite | **not one recorded golden** — see AC7's corrected reason |
| D-4.6.4 (i) | delete only the `HeaderRepeats` append in the clip branch | `TestClippedRowsPageStillCarriesItsRepeatedHeader/the clipped row's page repeats the header` — **and nothing else** | AC3, AC2(i), the suppression-floor test |
| D-4.6.4 (ii) | (same mutation as AC2 (ii)) | as AC2 (ii) | as AC2 (ii) |
| D-4.6.4 (iii) | always repeat, never fall to 4.4's arm (c) | `TestClippedRowSuppressesTheRepeatOnlyWhenItBuysNothingAndRecordsIt` — **and nothing else** | D-4.6.4's own repeat test, AC3, AC2(i) |
| F7-A | `g.Present = true` via field assignment in a non-test source | `TestAPresentItemGroupIsAlwaysATableRow` | — |
| F7-B | `layout.ItemGroup{Present: ok}` with a non-literal value | `TestAPresentItemGroupIsAlwaysATableRow` | — |
| F8 | truncate the tripwire's walk so it enters no files | `TestAPresentItemGroupIsAlwaysATableRow` (the vacuity guard fires) | — |
| D-4.5.5 control | revert the fixture's cell fill and border, change nothing else | `…ChromeIsTruncatedInTheRenderedDocument/the fixture's cells actually paint` + the content-stream subtest | — |

**The census's own limbs, proved separately** (each reddens `TestEveryPublicDiagCodeBridgeIsPinnedToALiteral`):
mint a new `DiagCode*` constant with no pin; add a pin naming a constant that does not exist; make the
enumeration match nothing (the zero-population guard).

**And the seven previously-unguarded bridges, re-bisected after the census landed:** all seven now
redden. See Finding 4's resolution for the measured set.

### The Class A ledger — declared 15, audited 14, Class A 6

Per **D-000.88**, three numbers rather than one ratio, and the declared-vs-audited delta is the
anti-gaming metric. **This story is n=1 of D-000.85's pre-registered series, its first data point came
back wrong, and the correction was REVIEWER-FORCED, not self-caught.** That is recorded here as a
property of the instrument, per D-000.88's ruling that the point is kept rather than dropped.

- **Declared: 15.** AC1 2 · AC2 2 · AC3 3 · AC4 3 · AC5 2 · AC6 1 · AC7 1 · AC8 1.
- **Audited: 14.** All 15 less AC8, which is prose and has no deletion screen by construction.
- **Class A: 6.** The story recorded **3**. The reviewer measured **at least 5**. Measuring every one
  at the final baseline gives **6**.

**Split by sub-cause (Task 10 — a flat count is not acceptable):**

*Compound-observable — an observable with no mutation that reddens it while its siblings stay green:*

1. **AC1 (ii)** — "the bytes are non-empty and structurally valid". **NEWLY FOUND BY THE FINISHER.**
   AC1's own named deletion (clip to zero height) leaves both of `TestOverTallRowRendersBytesRatherThanFailing`'s
   subtests **green**: a document that destroyed one row is still a well-formed document with bytes and
   a resolving page tree. Observable (ii) cannot fail while (i) passes.
2. **AC2 (i)** — "the over-tall row's page contains no other row". Reviewer-found (Finding 2), and
   confirmed here against **two** mutations: deleting the fresh-page step, and hand-placing the
   preceding row onto the clipped page. Both also redden band containment, because both change the
   page's `Shift`. **AC2 (ii) is now separably witnessed** (see the screen table) — that mutation
   exists only because D-4.6.4 introduced the reservation, so the reviewer's finding was correct
   about the code as it stood.

*Prose-claim — no deletion screen exists by construction:*

3. **AC8** — the reachability record in `OverflowError`'s doc comment. Declared as such by the story
   itself (D-000.84's named third class); carried, not counted as a witness.

*Unwitnessed mechanism — the code was live and nothing observed it. **All three were UNDECLARED**, and
all three are now CLOSED:*

4. **AC3 (iii)'s render-side half.** Deleting `render.go`'s application of the clip bound produced
   **0 new failures across 1082 test entries** and a **byte-identical** document. Reviewer-found
   (Blocker 1). **Closed** — the same deletion now reddens 3 named subtests.
5. **D-4.6.4's header repeat on a clipped row's page.** The mechanism was not merely unwitnessed, it
   was **absent**, and no assertion in the story could have seen it. Reviewer-found off-list.
   **Closed** — 3 observables, each separably red-proved.
6. **The `DiagCode` bridge census (D-000.87).** Seven exported constants with no witness of any kind.
   Reviewer-found (Finding 4). **Closed.**

*Class B, recorded as such rather than counted as a witness:* **AC7**, which asserts the **absence** of
an effect. Unchanged from the story's own declaration.

> **PROPOSAL, not a correction — D-000.88 forbids the measured party from moving its own denominator.**
> On the evidence above, **AC1 and AC2 each declare two observables where only one is separably
> witnessed.** A corrected declaration would read **13** rather than 15. **This is put to the reviewer
> to adjudicate and is deliberately NOT applied**: the story's declared count stays 15 in this
> document, and the delta stays visible.

---

## Review Summary

- Reviewed by: bmad-code-reviewer (fresh, adversarial; did not write this code, fixed nothing)
- Date: 2026-08-27
- Story Status Recommendation: **Changes Requested**
- Blockers: 2 · Majors: 4 · Minors: 4 · Nits: 2

**Method.** Every claim below was measured, not read. Twelve source mutations were applied to the
working tree one at a time, each reverted by writing back a pre-saved pristine copy and re-verified by
SHA-256; the tree at the end of the review is byte-identical to the tree at the start (`render.go`
`f4c3d20d…`, `table_footer.go` `abdbd9d2…`, `internal/layout/paginate.go` `be677d36…`,
`table_row_clip_test.go` `a92a9668…`), and a final full-suite run reproduces the entry state exactly.
All figures below come from raw `test2json` output via `rtk proxy`, never from `rtk`'s summary line.

**Verified and sound** (recorded so the finisher does not re-litigate them):

- **AC4's per-observable screen passes cleanly.** Deleting *only* the `clippedRowDiagnostic` append
  (`render.go:2000-2002`) reddens exactly `TestClippedRowReportsALocatedWarningNamingTheRow` (+ its
  presence subtest) and `TestRowTallerThanContentWindowIsClippedThroughRender` — **AC1, AC2 and AC3
  all stay green.** That green is the discrimination D-000.85 asks for, and it is real.
- **Story 4.5's ruled tie-revert genuinely still holds, and is witnessed three ways.** Deleting the
  clipped-footer stand-down (`table_footer.go:239-241`) reddens
  `TestFooterAloneTooTallForTheWindowIsClippedRatherThanFatal`; deleting the `anyTieWasClipped` guard
  reddens `TestUnsatisfiableFooterTieRecordsTheOrphanSuppressedWarning`,
  `TestUnsatisfiableFooterTieStillRendersWithTheFooterPlacedAlone` and
  `TestUnsatisfiableTieForOneTableDoesNotAbandonAnothersTie`; collapsing the per-table `case err == nil`
  arm reddens the same three. DECISION-2(b) is held under the new signal, not merely green.
- **`TestOverTallSingleRowStillOverflows` is untouched** (`git diff` shows no hunk near it) and passes.
  Its fixture is ungrouped and correct. Not widened.
- **The four inverted tests all assert substance**, not renamed vacuity. The union-height assertion
  survived on the first (`paginate_group_test.go`: `c.ItemHeight != 150000` — "the GROUP's union height,
  not any one member's own"), and each new arm reddens under the mutations below.
- **AD-1 / AD-5 / AD-23 clean.** No float or `math.` in the diff; no `time`/`os`/`math/rand` under
  `internal/`; no PDF clip path added (`RectClip` is a column-space bound applied as
  `min(rect bottom, bound)`); every `range` in the new layout code is over a slice, never a map.
- **AC7's headline figure is exactly right.** The unconditional-clip mutation reddens **57 test entries
  = 46 top-level + 11 subtests**, all in the table behaviour suite. The reported "46" reproduces.
- **The `emptyWalkFloor` correction (60 measured → 30 floor) is the right call and is *not* a
  conservation assertion in disguise** — 2× headroom, and its stated reasoning (a legitimate change to
  the clip boundary must not trip a vacuity guard) is correct. Contrast Finding 8.

**Gate discipline.** Skips are the expected set of **4**, by name:
`TestXrefEntriesRejectsMalformedSubprocess`, `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`,
`TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`, `TestTwoTablesWithPageCountFooterRenderConsistently`.
No new skips. Red by design, named separately: `TestCorpusMeetsP6ExerciseFloors` (P6g 7, need 20).

---

## Review Findings

### Finding 1: the chrome truncation never reaches the document — deleting the render-side clip changes nothing, in tests or in bytes
- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/render.go:2091-2093`; witness at `folio-go/table_row_clip_test.go:446-466`
- **Observation**: Deleting the entire application of the clip bound —
  `if bottom, clip := rectClipBottomFor(assigned.ClippedRects, ref); clip && r.Y+r.H > bottom { r.H = bottom - r.Y }`
  — produced **0 new failures across all 1082 test entries**. Stronger: the rendered document is
  **byte-identical** with and without it (62,482 bytes, identical rolling hash, on `overTallRowDoc()`).
  An instrumented probe confirms the branch *does* fire in production (page 2, refs 26 and 27,
  `H 152544 → 110000`), so the code is live — it is simply unobserved. Two independent causes:
  (a) AC3(iii) asserts against `layout.Pagination.ClippedRects` in `itemsForTest`'s **synthetic** RectRef
  space (`table_pagination_test.go:104-112`, one ref per `tableRectSource`), which is not production's
  `pdfRects` space (`render.go:1850-1858` flattens `ts.rects...`, so one source yields N refs); and
  (b) neither `overTallRowDoc()` nor `tooTallRowDoc()` gives its table any cell fill or border, so the
  truncated rectangle emits no marks at all. `TestClippingOneRowPushesNothingOffItsPage` cannot see it
  either — it re-applies the bound itself from the plan (`table_row_clip_test.go:307-309`).
- **Impact**: AC3's observable (iii) is the only one of the three that reaches a reader, and it has no
  witness at the site that produces it. A regression here draws a table row's rectangle past the bottom
  of the sheet — exactly the harm AC3(iii)'s own failure message names — and the whole suite stays green.
  This is D-000.67 part 2: the mechanism carries two presence preconditions (the plan records a bound;
  the renderer applies it) and only the first was mutated.
- **Suggested Resolution**: Give one fixture a visible cell chrome (background or border) so the
  truncation is byte-observable, and assert the clipped row's rectangle height through the **public
  `Render` / page-model path**, not through `itemsForTest`. Then re-run the deletion at
  `render.go:2091-2093` and confirm it reddens. Declare this a Class A observable if it cannot be closed.
- **Related AC**: AC3 (observable iii)

### Finding 2: AC2's two observables are not separated, and the code's recorded justification for the replacement is false as measured
- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_row_clip_test.go:274-292` (the claim) and `:293-342` (the test);
  mutation site `folio-go/internal/layout/paginate.go:786-790`
- **Observation**: AC2's Deletion (i) states: *"The alone-on-its-page assertion reddens; (ii)'s
  untouched-predecessors assertion stays **green**."* Removing the fresh-page step
  (`if pageHasItem { page++; pages = append(...); pageHasItem = false }`) from the clip branch reddens
  **both** `TestClippedRowLandsAloneOnAFreshPageAndItsNeighboursDoNotMove` **and**
  `TestClippingOneRowPushesNothingOffItsPage`. The test file explicitly anticipates this hazard,
  discloses that it replaced the story's original AC2(ii) formulation because that one "cannot stay
  green under AC2's own first deletion", and asserts of the replacement: *"This one can, and does."*
  **It does not.** Nor is there any other mutation that reddens (ii) alone: the AC3(iii) mutation
  (no `RectClip` emitted) also takes it down, because the test reads the untruncated chrome bottom
  (303,296mp) as outside the band.
- **Impact**: AC2 has one observable wearing two test names, so the redness of the alone-on-its-page
  assertion is not attributable. This story pre-registered itself as **n=1 of D-000.85's new series**;
  a recorded screen outcome that measurement contradicts corrupts the first data point of the series
  the programme is steering by, and it under-counts the Class A ledger.
- **Suggested Resolution**: Either find a mutation that reddens the band-containment property alone
  (e.g. advance the window by the group's untruncated height *while keeping* the fresh page — the
  story's own Deletion (ii)) and record its measured result, or declare AC2 a Class A
  compound-observable and correct the comment at `:286-292`. Do not leave the false claim in the source.
- **Related AC**: AC2 (observables i and ii)

### Finding 3: the clipped row's page silently loses its repeated table header — no repeat, and no suppression record
- **Severity**: Major
- **Category**: Correctness / Convention
- **Location**: `folio-go/internal/layout/paginate.go:781-845` (the clip branch's `continue` at `:844`)
- **Observation**: Measured on this story's own fixture (`overTallRowDoc()` / `overTallRowFixtureData()`):
  page 1 → `headerRepeats=1`, page 3 → `headerRepeats=1`, **page 2 (the clipped row) → `headerRepeats=0`**,
  and `plan.Suppressed` is `[]` for the whole document. The clip branch `continue`s before Story 4.4's
  entire DECISION-2/DECISION-3 block, so neither the repeat nor the `TableHeaderSuppressed` record that
  Story 4.4 requires when a repeat is dropped is ever produced.
- **Impact**: Story 4.4's DECISION-2 arm (c) is explicit in this same file: *"Suppress the repeat on THIS
  page only … **Recorded, never silent.**"* AD-14 says over-tall and clipped cases are *"never silent"*.
  This story's own new `folio-format.md` prose asserts *"Nothing is silent in either direction."* All
  three are now false for exactly one page per clipped row: the reader gets an unlabelled table row on a
  continuation page and no `TABLE_HEADER_REPEAT_SUPPRESSED` warning. No test can see it —
  `rowPartition` reads `ContentRects` only, and header repeats travel on the separate `HeaderRepeats`
  channel. (Note the over-tall **header** arm does *not* have this defect: probed directly, it clips on
  its own page and the next page's repeat is suppressed **and recorded**.)
- **Suggested Resolution**: Append a `TableHeaderSuppressed` record in the clip branch when the clipped
  group is a data/footer row of a table whose header sits on an earlier page (the same condition the
  `table != ""` block computes), or route the clip branch through that block. Add an assertion on
  `plan.Suppressed` / the emitted diagnostics for the clipped page. Suppressing the repeat is the right
  behaviour; the silence is the defect.
- **Related AC**: AC2, AC4 (AD-14's never-silent limb)

### Finding 4: D-000.87's census guard was not built — the bridge ships a hand-listed 5-of-15, and 7 constants have no guard of any kind
- **Severity**: Major
- **Category**: Convention / Tests
- **Location**: `folio-go/diag_bridge_test.go:22-49`; population at `folio-go/diagnostic.go:128-296`
- **Observation**: The bridge-pin finding itself **reproduces and is real** — Stories 4.4 and 4.5 did
  reach this file with no literal pin, and all three Epic 4 codes are pinned now. But D-000.87, ruled
  during this story, requires a **census**: *"Enumerate every public `DiagCode*` constant and require
  each to have a literal pin. The guard fires the moment a code is minted without one,"* plus *"the
  enumeration must fail loudly if it finds ZERO constants."* What shipped is a hand-written 5-entry
  table with no enumeration and no vacuity guard, over a population of **15** exported `DiagCode*`
  constants. Measured by bisect (each bridge rewritten to a wrong literal, `go test . ./cmd/folio`),
  **7 constants redden nothing at all**: `DiagCodeTableFooterSourceForbidden`, `DiagCodeTemplateMalformed`,
  `DiagCodeBindingPathAbsent`, `DiagCodeExpressionInvalid`, **`DiagCodeContentUnlayoutable`**,
  `DiagCodeDocumentDateInvalid`, `DiagCodeStyleColorInvalid`. Three of the ruling's "ten currently
  unpinned" are in fact caught indirectly (`TextMissingGlyph` → 4 tests, `TableFooterSourceUnresolved` → 1,
  `InternalUnhandledCaveat` → 1, via `TestAllProducedDiagnosticsCarryARegisteredCode` and the cmd tests),
  so the ruling's own "ten" is an unmeasured figure and the true unguarded population is **7**.
- **Impact**: `DiagCodeContentUnlayoutable` is the exact constant this story's AC8 and D-4.6.2 asymmetry
  rest on — *"an over-tall UNGROUPED item still produces `CONTENT_UNLAYOUTABLE`"* — and its bridge can be
  changed to an arbitrary string with the suite fully green. The ruling names **Story 4.6's finisher** as
  owner, so this is correctly routed rather than concealed; it is logged so it is discharged rather than
  absorbed, which is the precise failure mode D-000.87 was written about.
- **Suggested Resolution**: Build the census as ruled (reflection or an AST scan of `diagnostic.go`'s
  `const DiagCode… = string(diag.Code…)` declarations, cross-checked against a test-owned literal table,
  with a non-zero-population guard), and use the measured **7**, not the ruling's unmeasured 10.
- **Related AC**: AC4 (registry obligations), AC8

### Finding 5: Tasks 9 and 10 are ticked, but the story file carries no Delivery Log / Dev Agent Record
- **Severity**: Major
- **Category**: Maintainability / AC Conformance
- **Location**: `_bmad-output/implementation-artifacts/4-6-handle-a-row-taller-than-the-page.md` (ends at
  line 509, "Dev notes"; tasks at `:480-485`)
- **Observation**: Task 9 requires *"Report every figure with its units … report the **set** … Report
  skips **by name**"*; Task 10 requires *"Report the Class A ledger split by sub-cause … A flat count is
  not acceptable."* Both are marked `[x]`. The story file contains no such section: no suite figures, no
  skip set, no per-observable mutation record, no Class A ledger. The decision log carries the AC7
  measurement and the D-4.6.* rulings, but none of the delivery record.
- **Impact**: The 14 named deletions, the "none survived" claim and the "3 Class A of 15" split are
  unauditable from the artifact. Findings 1 and 2 above show at least one unnamed site survives and at
  least one named screen outcome is misreported — precisely what a written ledger exists to expose at
  review rather than at a gate.
- **Suggested Resolution**: Add the Delivery Log with per-observable mutation results (mutation applied,
  tests reddened by name, tests that stayed green), the suite figures with units per Finding 6, the skip
  set by name, and the Class A ledger split by sub-cause — including the two undeclared entries this
  review found.
- **Related AC**: Tasks 9, 10

### Finding 6: the reported suite figures are `rtk`'s *passed* counts labelled as totals, and the 1049↔1055 drift is an invocation difference, not a counting convention
- **Severity**: Major
- **Category**: Correctness (measurement) / Convention
- **Location**: reported figures for `folio-go`; cf. `4-5-render-footer-aggregates-without-orphaning-them.md:1084`
- **Observation**: Measured from raw `test2json` via `rtk proxy`:

  | invocation | commit | top-level total | top passed | incl. subtests total | passed |
  | --- | --- | --- | --- | --- | --- |
  | `go test ./...` | this HEAD | **712** | 707 | **1082** | 1076 |
  | `go test ./...` | `45cf812` | **704** | 699 | **1061** | 1055 |
  | `go test -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | this HEAD | 711 | 707 | 1074 | 1070 |
  | `go test -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | `45cf812` | 703 | **699** | 1053 | **1049** |

  The reported "707 / 1076 (from 699 / 1055)" are the **passed** counts, not totals — `rtk`'s own summary
  line at this HEAD reads verbatim `1076 passed, 2 failed, 4 skipped`, which is the number that was read.
  On the drift: Story 4.5's **699 / 1049 is exactly reproducible and correct for its own invocation**
  (with `-skip`). This story paired 699 (a `-skip` figure) with 1055 (a no-`-skip` figure) — two different
  invocations in one baseline row. `1055 − 1049 = 6` is exactly the six *passing* subtests of
  `TestCorpusMeetsP6ExerciseFloors` that `-skip` removes. Nothing to do with how nested subtests are counted.
- **Impact**: D-000.83 exists for unexplained count drift; an explanation that names the wrong cause
  closes the question falsely. Passed-counts labelled as totals also hide the 2 failures and 4 skips
  inside the headline number.
- **Suggested Resolution**: Report `total (passed / failed / skipped)` with the exact invocation beside
  it, from raw `test2json` (`rtk proxy`), never from `rtk`'s summary line. Correct the 1055 → 1049
  attribution: same-invocation baselines, or state the invocation per figure.
- **Related AC**: Task 9

### Finding 7: the D-4.6.2 tripwire catches one syntactic spelling of the property it guards
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_row_clip_test.go:665-757` (`setsPresentTrue` at `:741-757`)
- **Observation**: Red-proved three ways, in package `folio`'s own non-test sources:
  - `func zzMutA() layout.ItemGroup { return layout.ItemGroup{Present: true} }` → **reddens correctly**.
  - `var g layout.ItemGroup; g.Present = true` → **passes silently**.
  - `layout.ItemGroup{Present: ok, Key: …}` (a bool variable) → **passes silently**.
  `setsPresentTrue` requires an `*ast.KeyValueExpr` whose value is the identifier `true`, so field
  assignment and any non-literal value evade it. Neither evasion is flagged by `go vet` either.
- **Impact**: D-4.6.2 required this tripwire precisely because `Group.Present ⟺ table row` is *"accidentally,
  not by construction"* true, and the failure mode is silent widening of the clip. A guard shaped to one
  spelling is an AST whitelist, not the property.
- **Suggested Resolution**: Also flag assignments to a `.Present` field of a `layout.ItemGroup`-typed
  value, and treat a non-`true` literal value as suspicious rather than absent (or use `go/types` to
  resolve the type rather than matching the `layout.ItemGroup` selector spelling).
- **Related AC**: D-4.6.2's tripwire requirement

### Finding 8: the tripwire's vacuity floor is pinned to its exact measured population, contradicting the same file's own correction 300 lines above
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_row_clip_test.go:720-724`; contrast `:327-341`
- **Observation**: `const measuredAtStory46 = 6` with `if constructions < 6 { t.Fatalf(...) }`, against a
  population measured at exactly 6 (`render.go:206,219,221`; `table_render.go:246,251,253`). The same
  file's `emptyWalkFloor` deliberately sits at 30 against a measurement of 60, with an explicit and
  correct rationale for not pinning a floor to its measurement. The methodology correction was applied to
  one floor and not the other.
- **Impact**: A legitimate refactor that merges any two of the six arms produces a false red with a
  vacuity message. The upper direction is genuinely guarded (a 7th construction elsewhere fires the
  `t.Errorf`), so this is brittleness, not a hole.
- **Suggested Resolution**: Drop the floor to a value that guards vacuity without asserting conservation
  (e.g. 2 — one arm per derivation), or state explicitly that 6 is a conservation assertion and why.
- **Related AC**: D-4.6.2's tripwire requirement

### Finding 9: the header-row arm of the shared code is never exercised through pagination
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_row_clip_test.go:571-600`
- **Observation**: `TestClippedGroupDiagnosticNamesTheRoleNeverTheSentinel` exercises all three roles by
  calling `clippedRowDiagnostic` **directly** with a synthetic `layout.TableRowClipped`. The data-row arm
  additionally has four end-to-end tests and the footer arm has
  `TestFooterAloneTooTallForTheWindowIsClippedRatherThanFatal`; the **header arm has nothing that runs it
  through `Paginate`**. I probed it directly: the behaviour is correct (clipped alone on page 0, and the
  next page's repeat correctly suppressed **and recorded** —
  `Suppressed=[{ElementID:e1 Page:1 RowHeight:10000 Available:110000 HeaderHeight:200000}]`).
- **Impact**: D-4.6.3's headline is *"ONE CODE FOR ALL THREE GROUP ROLES"*, and that is the ruling's
  novelty (the discriminator merging rather than splitting). Two of three roles are witnessed end to end;
  the third rests on a direct call to the message builder, which cannot see a pagination-level regression.
- **Suggested Resolution**: Add one `layout.Paginate` case with an over-tall header group asserting
  `plan.Clipped[0].Key.IsHeader` — cheap, and it closes the ruling's third arm.
- **Related AC**: AC5, D-4.6.3

### Finding 10: AC7's byte-neutrality claim is true for a materially wider reason than AC7 states
- **Severity**: Minor
- **Category**: Tests / AC Conformance
- **Location**: story `:388-399`; corpus at `fixtures/*/expected.pdf`
- **Observation**: Verified independently and confirmed. The unconditional-clip mutation reddens 46
  top-level tests + 11 subtests, all in the table behaviour suite, and **not one recorded golden**.
  Cause verified directly: none of the nine committed golden PDF fixtures (`font-text`, `image-embed`,
  `minimal-rect`, `multi-page`, `multi-script-fallback`, `page-count-20`, `shaped-text`,
  `three-band-page`, `wrapped-text`) contains a table element; the only `"type": "table"` under
  `testdata`/`fixtures` is `folio-go/testdata/template/golden/worked-example.json`, a **template**
  round-trip golden, not a rendered-PDF golden.
- **Impact**: AC7 reads *"no document in the corpus contains an over-tall group, so this story's branch is
  not reached by any of them."* The true statement is stronger and more consequential: **no document in
  the corpus contains a table at all**, so no behaviour of Epic 4 — five stories' worth — is byte-pinned
  anywhere. AC7 is therefore not a measurement of this story's branch; it is a measurement of a corpus
  that cannot exercise the epic. This is properly surfaced and routed in D-4.7.0 and it bears directly on
  Story 4.7; stated here as instructed regardless of severity.
- **Suggested Resolution**: None for this story — carry D-4.7.0's three obligations into 4.7's brief, and
  reword AC7 so the recorded reason matches the measured one.
- **Related AC**: AC7

### Finding 11: `dropped[]` cannot drop the runs of a group member that carries both `Rects` and `Runs`
- **Severity**: Nit
- **Category**: Correctness (latent)
- **Location**: `folio-go/internal/layout/paginate.go:809-822`
- **Observation**: `if len(items[j].Rects) > 0 { …; continue }` short-circuits before the
  `if items[j].Bottom > contentBottom { dropped[j] = true }` assignment. An item carrying both would have
  its rect bounded but its runs kept, drawn past the content bottom.
- **Impact**: Unreachable today — both `paginateDocument` (`render.go:1869/1923/1963`) and `itemsForTest`
  build rect-items and run-items disjointly — but it is a silent-content-retention trap for any future
  construction that carries both, and the drop path is the one that destroys content.
- **Suggested Resolution**: Handle `Rects` and `Runs` independently rather than as an either/or, or
  assert the disjointness the code assumes.
- **Related AC**: AC3

### Finding 12: `AssertPDFPageTreeResolves` gates on the whole subtest's failure state, not its own
- **Severity**: Nit
- **Category**: Tests
- **Location**: `folio-go/pdf_page_tree_check_test.go` (`if t.Failed() { return false }`, and the final
  `return !t.Failed()`)
- **Observation**: Both gates read `t.Failed()`, which is true if *anything* in the calling (sub)test has
  already failed, not only this helper. Its doc comment says it *"reports whether the document passed"*.
- **Impact**: Harmless at both current call sites (each is preceded only by `t.Fatalf`s), but the
  extraction is explicitly justified as the single shared oracle, so a future caller with a prior
  non-fatal failure gets a false "document invalid" verdict.
- **Suggested Resolution**: Track this helper's own failures in a local bool and return that.
- **Related AC**: AC1

---

## Finding Resolutions (finisher)

**Triage: 12 FIX · 0 DISMISS · 0 DEFER.** Every finding had a concrete, in-scope, cheaply-verifiable
resolution and every one reproduced when re-measured. One Major was **escalated to Blocker** by the
orchestrator and then **re-ruled by the engineering lead** (D-4.6.4) in a direction opposite to the
escalation's own instruction — see Finding 3.

> **A note on the counts in the finisher's brief.** The brief described the header-loss defect as "one
> unlisted finding … escalating to Blocker", but it is **Finding 3**, already in the review as a Major.
> The review's own tally (2 Blockers · 4 Majors · 4 Minors · 2 Nits = 12) is correct and was verified by
> counting `**Severity**` lines directly. There are 12 findings, not 13.

| # | Severity | Decision | Files touched |
| --- | --- | --- | --- |
| 1 | Blocker | **FIX** | `table_row_clip_test.go` |
| 2 | Blocker | **FIX** | `table_row_clip_test.go` (+ the story's Class A ledger) |
| 3 | Major → **Blocker** | **FIX** (re-ruled: repeat, don't record) | `internal/layout/paginate.go`, `table_row_clip_test.go` |
| 4 | Major | **FIX** | `diag_bridge_test.go` |
| 5 | Major | **FIX** | the story file (`## Delivery Log`) |
| 6 | Major | **FIX** | the story file (gate figures) |
| 7 | Minor | **FIX** | `table_row_clip_test.go` |
| 8 | Minor | **FIX** | `table_row_clip_test.go` |
| 9 | Minor | **FIX** | `internal/layout/paginate_group_test.go` |
| 10 | Minor | **FIX** | the story file (AC7's stated reason) |
| 11 | Nit | **FIX** | `internal/layout/paginate.go` |
| 12 | Nit | **FIX** | `pdf_page_tree_check_test.go` |

### Finding 1 — the chrome truncation never reaches the document · FIX

Reproduced exactly: deleting `render.go`'s application of the clip bound gave **0 new failures across
1082 entries**. Both causes had to be closed or the assertion re-inerts, and both were.
**(a) Wrong ref space:** the shipped AC3(iii) reads `ClippedRects` in `itemsForTest`'s synthetic
one-ref-per-source space. The new `TestClippedRowsChromeIsTruncatedInTheRenderedDocument` asserts in
production's **flattened `pdfRects` space**, reached through `buildPageModel`, and pins the flattening
itself (2 columns → 2 rects per row) so the two spaces cannot silently diverge again.
**(b) Nothing painted:** `overTallRowDoc()` now declares a cell `background` **and** a `border`, and
the test asserts that precondition rather than assuming it. The truncation is now asserted in the
**bytes** too, both directions — the fill operator `10 20 60 100 re f` must be present and the
untruncated height `152.544` must appear nowhere in the document.
**Re-run:** the same deletion now reddens 3 named subtests, while the plan-side sibling
`…/chrome truncated at the content bottom` stays **green** — which is D-000.67 part 2's two presence
preconditions finally getting two proofs. A control mutation reverting only the fixture's chrome
confirms cause (b) was load-bearing.

### Finding 2 — AC2's two observables are not separated · FIX

Confirmed: removing the fresh-page step reddens **both** AC2 tests, so the recorded claim *"This one
can, and does"* was false as measured. **The false claim is removed from the source**, replaced by
what was actually measured — including the fact that it was the reviewer who found it.
**AC2 (ii) is now separably witnessed:** the mutation `reserved = hh` → `reserved = 0` (repeat the
header, but do not narrow the cut by it) reddens `TestClippingOneRowPushesNothingOffItsPage` with
`TestClippedRowLandsAloneOnAFreshPageAndItsNeighboursDoNotMove` **green**. That mutation exists only
because Finding 3's fix introduced the reservation, so the finding was correct about the code as it
stood.
**AC2 (i) is NOT separably witnessed** — two mutations tried, both redden containment too. It is
carried as **Class A #2**. Per **D-000.88** the corrected observable count is **proposed, not
applied**: the finisher does not move its own denominator.
Also fixed here: `TestClippingOneRowPushesNothingOffItsPage` was reading each item's drawn extent
using the page `Shift` **only**, ignoring the per-table `RowDisplacement` — which under-reports every
row of a repeating table by the header's height, precisely the direction that hides content pushed off
the **bottom**. With the displacement included it can see the harm it names, and that is what makes
AC2 (ii)'s separating mutation observable at all.

### Finding 3 — the clipped row's page silently loses its repeated header · FIX, re-ruled

Reproduced: `headerRepeats` was 1 on pages 1 and 3, **0 on page 2**, with `plan.Suppressed` empty.
**The brief instructed "fix the recording". The engineering lead overturned that mid-task (D-4.6.4)
and the implemented answer is the opposite one: the header REPEATS.** Story 4.4's suppression has a
stated trigger — *reserving the header leaves no room for a data row* — and that trigger was never met
here, because there **is** a row on the page. 4.4's remedy was being applied where 4.4's condition
never fired, and recording it would have documented a suppression that should not happen at all.
The clip branch now composes the two rules instead of short-circuiting one: **repeat the header, then
clip the row into the space below it** (`contentBottom` narrows by exactly the header's height). The
floor is 4.4's own — if the reservation leaves room for **not one line**, DECISION-2 arm (c) fires on
its own terms and is recorded through the existing `TableHeaderSuppressed` channel.
**Both arms are witnessed by real documents, differing in exactly the property the branch keys on:**
`overTallRowDoc()` repeats (9 of its 14 lines survive under the reservation); `tooTallRowDoc()`'s
single 272,400mp line survives neither cut, so it suppresses **and records**, and the
`TABLE_HEADER_REPEAT_SUPPRESSED` warning reaches the reader.
**Three separably red-proved observables** (see the screen table). And the falsification is
**disposed of rather than reconciled**: with the header repeating there is no silent suppression, so
Story 4.4's *"recorded, never silent"*, AD-14's *never silent*, and this story's own `folio-format.md`
sentence *"Nothing is silent in either direction"* are all **true**. The documentation was right and
the code was wrong; **the documentation was not weakened to match it.**
Confirmed free on hashes, as D-4.6.4 predicted and required be checked, not assumed.

### Finding 4 — D-000.87's census was not built · FIX

What shipped was a hand-written 5-entry table over a population of 15 — the partial sweep, not the
fix. **The real number was measured here, not inherited.** Bisecting every one of the 15 bridges in
turn (rewrite to a wrong literal, run the whole suite) gives **7 constants that redden nothing at
all**, reproducing the reviewer's set exactly:

`DiagCodeTableFooterSourceForbidden` · `DiagCodeTemplateMalformed` · `DiagCodeBindingPathAbsent` ·
`DiagCodeExpressionInvalid` · **`DiagCodeContentUnlayoutable`** · `DiagCodeDocumentDateInvalid` ·
`DiagCodeStyleColorInvalid`

**Seven, not the ruling's ten.** That figure was arithmetic over a name-grep, and *"not literally
pinned"* and *"reddens nothing under mutation"* are different properties — only the second is what a
census is for. Three of the ten (`TextMissingGlyph`, `TableFooterSourceUnresolved`,
`InternalUnhandledCaveat`) are caught indirectly; confirmed by the same bisect.
**And this is not tidying.** `DiagCodeContentUnlayoutable` is the code naming the **refused** half of
this story's own authorship asymmetry — rows forgiven because their height came from data, typed
content refused. It could have been repointed at any string with the whole suite green.
`TestEveryPublicDiagCodeBridgeIsPinnedToALiteral` enumerates the population by **parsing package
folio's own non-test sources** (not a filename, not a remembered list), requires a **test-owned
literal** per constant — never a comparison against the constant itself, which is the defect — and
**fails loudly on a zero population**. All three limbs red-proved separately, and all seven formerly
unguarded bridges now redden.

### Finding 5 — Tasks 9 and 10 ticked with no Delivery Log · FIX

`## Delivery Log` written, with the per-observable screen results (mutation, tests reddened by name,
tests that stayed green), the gate figures with units and invocations, the skip set by name, and the
Class A ledger split by sub-cause — including the **three undeclared entries**. Its provenance is
stated at the top: it is the finisher's own measurement, not a reconstruction of the developer's run.

### Finding 6 — passed counts labelled as totals · FIX

Reproduced exactly, including the drift's true cause. The corrected table is in the Delivery Log, with
`total (pass / fail / skip)` and the exact invocation on every row. `45cf812`'s two invocations were
re-measured in a detached worktree; `1055 − 1049 = 6` is the six passing subtests of the red-by-design
test that `-skip` removes.

### Finding 7 — the tripwire has teeth for one spelling · FIX

Both evasions closed and both red-proved live in package `folio`'s own non-test sources.
`g.Present = true` (field assignment) is now flagged by the **field being written**, deliberately
over-approximating rather than resolving the receiver's type with `go/types` — over-approximating
costs a comment on a false positive, under-approximating loses the property silently, and the package
contains no such assignment today so it costs nothing. `Present: ok` (a non-literal value) is closed
by inverting the test: everything counts **except** the literal `false`, because a value the scan
cannot evaluate is a value that might be true.

### Finding 8 — the tripwire's floor is pinned to its measurement · FIX

Dropped from 6 to **2** — one arm per derivation, the least that can exist while both derivations do —
with the rationale `emptyWalkFloor` already carried 300 lines above. The methodology fix now applies to
both floors. The upper direction is where this test's teeth are and is untouched. The guard still fires
on a truncated walk, red-proved by making the scan enter no files.

### Finding 9 — the header arm never runs through pagination · FIX

`TestPaginateOverTallHeaderGroupClipsAndIsRecordedAsTheHeader` added: an over-tall header group through
the real `Paginate`, asserting `plan.Clipped[0].Key.IsHeader`. It also pins the one thing the header arm
does differently, and it is why D-4.6.4's repeat block is guarded on `!IsHeader` — a table's header is
never repeated above **itself** (DECISION-1) — while the next page, which cannot reserve a header far
taller than its window, falls to DECISION-2 arm (c) and is recorded there. D-4.6.3's third role is now
witnessed end to end.

### Finding 10 — AC7 is true for a wider reason than it states · FIX

Verified independently: the unconditional-clip mutation reddens **57 entries (44 top-level + 13
subtests)**, all in the table behaviour suite, and **not one recorded golden**. AC7's stated reason is
corrected in place — the true statement is that **no document in the recorded corpus contains a table
at all**, so no behaviour of Epic 4 is byte-pinned anywhere. Routed to Story 4.7 via D-4.7.0.

### Finding 11 — `dropped[]` cannot drop the runs of a member carrying both · FIX

The short-circuit is gone. A member's rects and its runs are now answered **independently**, each
getting the treatment its own nature admits: a rectangle is a coordinate and is **truncated**; a line
is atomic and is **dropped whole** (AD-24 forbids drawing half of one). `dropped[i]` now gates runs and
images only, and the emission loop is explicit that rects still emit because a rect is never dropped.
Behaviour is identical for every input this module builds today — nothing carries both kinds — so this
has **no behavioural red-proof and none is claimed**; it is stated as a latent-trap fix, which is what
the finding called it.

### Finding 12 — `AssertPDFPageTreeResolves` gates on the whole subtest · FIX

Both gates and the return now read a **local** failure flag, set by a small recorder the helper's own
`Errorf` calls go through. The doc comment says "this check" and now means it. Harmless at both current
call sites; wrong the moment a third arrives with a prior non-fatal failure — which is the point of
extracting it as the single shared oracle.
