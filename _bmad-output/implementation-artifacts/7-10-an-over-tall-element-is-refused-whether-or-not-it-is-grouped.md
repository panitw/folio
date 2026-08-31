---
title: 'Story 7.10: An over-tall element is refused whether or not it is grouped'
type: 'bugfix'
created: '2026-08-31'
status: 'ready-for-dev'
baseline_revision: '9844e6dc84672302513b3d4460ca86071bb786bd'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
warnings: ['oversized']
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative. The intent contract below governs; where the two differ, the contract wins.*

If you declare a box too tall for the page, Folio refuses your document and tells you which element
is wrong. But if you also tag that element into a keep-together group, the refusal quietly becomes a
warning: Folio prints the document, cuts the element off at the bottom of the page, and carries on.
So a feature that has nothing to do with heights can switch off a hard error. This story makes the
refusal follow **what** is over-tall rather than **whether** it happens to be grouped.

<intent-contract>

## Intent

**Problem:** A content-band element the author declared taller than one content window is a located,
fatal `OverflowError` when untagged and a clip-and-warn when tagged into a keep-together group. The
tag launders authorship: an unrelated declaration switches a hard error into a warning. Story 7.7's
contract matrix says both things (rows 3 and 5) and they collide (DW-47, DW-50).

**Approach:** Split the disposition on **what is over-tall**, not on whether it is grouped
(D-7.7.9). An individually over-tall element is a located fatal `OverflowError` whether or not it
carries a group tag; a keep-together group that exceeds a window **only in aggregate** — every
member fitting, the sum not — keeps Story 4.6's clip-and-warn unchanged. Table rows are untouched:
D-4.6.2's ratio is that leniency follows **authorship**, and a row's height comes from data the
author cannot fix. Correct the three documents the ruling makes stale in the same commit as the
behaviour, and give the fatal message a word that is true of the element it names.

## Boundaries & Constraints

**Always:**
- The discriminator is **what is over-tall**, never **whether it is grouped** (D-7.7.9).
- **Leniency follows authorship (D-4.6.2).** A table row's height is derived from data the author
  cannot audit and stays clip-and-warn. A loose element's declared box is the author's own and is
  fatal, per D-2.6.1.
- **A group of one is a no-op**: it must be indistinguishable from no group. An author must not be
  able to escape a fatal error by declaring an unrelated feature.
- **Both arms are asserted against ONE fixture subject** — one geometry, one element set, differing
  in exactly which element is tagged. Either arm alone proves nothing about the discriminator,
  because either arm alone is **also consistent with the rule being replaced**: an isolated fatal
  case is consistent with "untagged elements are fatal", and an isolated clip case is consistent
  with "grouped things are clipped". Only the pair, held against a common base, measures the
  discriminator itself.
- Story 7.7's third acceptance criterion, `ARCHITECTURE-SPINE.md`'s AD-14 carve-out and
  `folio-format.md`'s over-tall table are amended **in the same commit as the behaviour**. A spine
  running ahead of the code is the same defect as one lagging it (D-7.7.14).
- This story **narrows what renders** and is therefore free only before `folio-go/v0.1.0`
  (D-7.7.13, D-7.8.3, AD-22). The spec must enumerate which currently-rendering documents stop
  rendering.
- All **22** `goldenDigestRecord` digests are verified unmoved. If any shipped fixture contains a
  tagged, individually over-tall element, this story turns its bytes into a fatal error — **say so
  and HALT**; never re-record.
- `fixtures/statement-signoff.json` is a human attestation. No agent writes `reader`, `date` or
  `examined`.

- **THE INTENT GAP IS RESOLVED — six rulings, 2026-08-31 (D-7.10.1 … D-7.10.6). Implement them; do
  not re-derive them.** The halt was correct: D-7.7.9's mechanism did not reach multi-line text,
  because it reasoned in **elements** while the paginator groups **line items**.

- **RULING 1 — A GROUP'S MEMBERS ARE THE TEMPLATE ELEMENTS CARRYING THE TAG, not the column items
  they decompose into.** The tag is declared on an element; the split into one item per shaped line
  is how the paginator **implements** keeping things together — an internal decomposition, not a
  statement about what the author grouped. **Reading the discriminator in items lets an
  implementation detail decide a product behaviour**, which is exactly how a one-element group came
  to be classified "aggregate-only" and DW-50 fell through. Consequences:
  - **a tagged element whose OWN extent exceeds the window → FATAL** (DW-50 is reached and fixed);
  - **a tagged multi-element group, each element fitting, the sum not → clip-and-warn** (Story 7.7's
    AC preserved unchanged).

- **RULING 2 — why fatal is right, and the chain matters because the obvious counter-argument does
  not apply.** (1) **D-2.6.1 governs height overflow**: an item that fits in *no* window is a
  located template error, fatal — that is the default for anything too tall. (2) **AD-25's clip
  precedent is WIDTH-only** — FR44/Story 2.8 clip against the declared *width*, and D-2.6.1
  explicitly **excluded** page-edge overflow from FR44, so "the author declared this atomic and it
  does not fit, therefore clip" **does not transfer** from a value overflowing its box to an element
  overflowing a page. (3) **Story 4.6's clip is the exception to D-2.6.1**, justified specifically
  because a table row's height is **data-driven and the author cannot fix it**. (4) **A tagged
  element always has a fix in the author's hands: remove the tag.** The grouping exists only because
  they asked for it, so 4.6's exception does not reach it — **regardless of whether the element's
  height is data-driven, because the TAG never is.**
  **The real discriminator, sharper than D-7.7.9's: WHO CREATED THE GROUPING.** Engine-created
  (table rows, atomic by construction) → lenient. Author-created (a keep-together tag) → strict,
  because the author can dissolve it.

- **RULING 3 — AC1 IS REPLACED. "The same as untagged" was a PROXY, never the rule** — a proxy for
  *the author declared this and can fix it*. It was true for the population in front of the ruling
  (rects, images) and false for the one that was not (multi-line text). **Third proxy-versus-purpose
  instance this run.** The replacement states the difference from untagged rather than hiding it:

  > **Given** a keep-together group whose single tagged element's own extent exceeds the content
  > window
  > **When** the document paginates
  > **Then** it is refused with a located fatal `OverflowError` naming that element — because the
  > author declared an atomic block that cannot fit, which is **unsatisfiable** rather than merely
  > degraded, and removing the tag is the author's fix
  > **And** this differs **deliberately** from the untagged case, where the same element's lines
  > split across windows and render cleanly; **the tag is what makes it unsatisfiable**

  **BLOCK IF the story asserts message-equality with the untagged case.** Under this ruling the
  untagged document **renders**, so there is no untagged error to be equal to. Such an assertion
  would be false, and writing one is the most likely way an implementer "reconciles" the old AC.

- **RULING 4 — the aggregate-only case stays exactly as Story 7.7 shipped it. 7.10 does not touch
  it.** Named honestly rather than dressed up: Ruling 2's fixability argument, pushed all the way,
  would make that case fatal too — the author can untag it just the same. **7.7 chose clip by
  importing 4.6's TREATMENT without importing its REASON**, and the lead confirmed it at the time.
  It is left because it is shipped, deliberate and outside this story's subject — **not because it
  is obviously right.** *What would reopen it:* a real document losing content that way, as DW-50
  came from a real case. **If anyone sees one, raise it NOW** — one ruling would cover both, and it
  is cheaper before the tag than after.

- **RULING 5 — the `paginate.go` `Kind` defect is IN SCOPE, both halves.** Its comment enumerates
  *"exactly one of three sites"* and claims `Kind == "table"` is *"NO LONGER PRODUCED FROM PACKAGE
  FOLIO"*; `collectElementBoxRects`, added by Epic 9, is a **fourth, ungrouped source**, so the
  claim is **false at HEAD**. It is in scope because it is **load-bearing for this story, not merely
  adjacent**: AC1 requires the message to name the element, and today an over-tall **rect** is told
  it is a **table** — *"element e1: table is taller than the content window"*. **7.10 is the story
  that asserts that message for the first time, and asserting a wrong message pins the defect.**
  **Guardrail: `Kind` is derived — fix the DERIVATION, never special-case the string** — and correct
  the comment to state what is actually produced.

- **RULING 6 — the re-pointing obligation discharges by ADDING, and BOTH ARMS ARE REQUIRED.** There
  is nothing to re-point: today's clip-and-warn for a tagged over-tall element is **asserted by
  nothing** (seventh instance this run). The two arms **are the discriminator, not a demonstration**:
  the **tagged single element** → fatal, message naming the element and its **true kind**; the
  **tagged multi-element aggregate group** → still clipped and warned, unchanged. **Without the
  second arm the change reads as "tagging makes things fatal", and the next story generalises it in
  exactly the wrong direction.**
  **The two-document shape is confirmed:** AD-14 makes an `Error` abort and a `Warning` accompany
  success, so no single rendered document can yield both arms — **a property of the contract, not a
  gap in the fixture**. Follow the shipped `keepTogetherTemplateJSON` / `keepTogetherUngroupedTemplateJSON`
  pair, and hold the new pair to that pair's own standard: *"IT IS A DISCRIMINATOR, NOT A
  DEMONSTRATION."*

- **DW-64 and DW-65 ride here.** The implementation touches `internal/layout`, and **Epic 7's fence
  is closed and no longer applies** — the epic is `done` as of `9844e6d`.

**Block If:**
- **`multi-line text reach` — the ruling's mechanism does not reach the element kind it names.**
  See Design Notes, "The population D-7.7.9 does not reach". D-7.7.9's mechanism is per-member
  extent, but both pagination passes emit **one column item per shaped line**, so every member of a
  tagged multi-line text element fits and only the union exceeds. Under the ruling as written,
  DW-50's document is an *aggregate-only* group and is left exactly as it is today. Three
  dispositions differ observably and permanently and the contract selects none. HALT.
- A shipped golden's bytes would move, or a `goldenDigestRecord` digest changes.
- The discriminator, as implemented, reddens any table-row / header-row / footer-row clip test —
  that means it is unscoped and would reverse Story 4.6.
- Story 7.7's amended AC, AD-14 or `folio-format.md` cannot be corrected without contradicting a
  ratified invariant.

**Never:**
- Never touch `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` (the mandated red) or its
  drift twin.
- **Zero paths under `folio-designer/`.** This is engine work; no designer-side grouping exists.
- Never widen into Epic 8's font stories. Never re-open Epic 7's `done` status.
- Never re-record a golden, never write a sign-off record, never modify the repo-root `README.md`,
  never `git add -A` / `git add .`, never branch, never push.
- Never make an over-tall **table row** fatal. That reverses Story 4.6 and D-4.6.2.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Tagged over-tall declared box | One `rect`/`image`/box element, declared height > one content window, carrying `keepTogether` | Located fatal `OverflowError`, identical to the untagged element's | Error, never bytes, never a Warning |
| Untagged over-tall declared box (control) | Same element, no tag | Located fatal `OverflowError` — unchanged from HEAD | Error, unchanged |
| Aggregate-only group | Two elements, each fitting one window, union exceeding it, one tag | Clipped to a page of its own, one `TABLE_ROW_CLIPPED_HEIGHT` **Warning** beside the bytes, located at the group's first member, naming the group | Warning, never fatal |
| Over-tall table row | A data/header/footer row whose extent exceeds a window | Clip-and-warn, **unchanged** — leniency follows authorship (D-4.6.2) | Warning, never fatal |
| Fatal message wording | Untagged over-tall element-box `rect` | Message names the element's actual kind; the word "table" never describes a non-table | Error, message accurate |
| Single-member group of a multi-line text element | One text element, tagged alone, whose shaped lines sum past one window | **OPEN — see Block If `multi-line text reach`** | Undetermined |

</intent-contract>

## Code Map

**Every anchor verified by reading the file at baseline `9844e6d`.** The four behavioural claims the
dispatch supplied were re-measured at that revision with a throwaway probe outside the repo; results
are in Design Notes, "Re-verified at `9844e6d`".

### The deciding site — `folio-go/internal/layout/paginate.go`

- **`:548`** — `Paginate`. `height = ContentHeight(g)` at `:550`.
- ⚠ **`:790-791` — THE CLIP BRANCH, and the whole defect:**
  ```go
  if it.Group.Present {
      if itemHeight := effectiveBottom - effectiveTop; itemHeight > height {
  ```
  It keys on exactly two things: `ItemGroup.Present` (a bare boolean, nothing about kind) and a
  height comparison. `effectiveTop`/`effectiveBottom` are set at `:762-768` and for a grouped item
  are the **merged group union extent** from the `groups` pre-pass (`:641-659`) — never the item's
  own extent.
- `:781-789` — the comment defending that key (*"Keyed on Group.Present, and NOT on the item's kind
  … A branch keyed on Kind == \"table\" would clip nothing (D-4.6.2)"*). Still true of table rows;
  **becomes false as an unconditional statement** once the discriminator lands.
- Inside the branch: cut line `:899`; member sweep truncating rects / marking drops `:900-931`
  (`RectClip` `:925-928`, `dropped[j]=true` `:930`); the record `clipped = append(clipped,
  TableRowClipped{…})` `:935-941`; the drop at emission `:1138-1147` (`if !dropped[i]` gates
  `ContentRuns` and `ContentImages`; `ContentRects` always emits).
- ⚠ **`:1075-1088` — the fatal arm, reached ONLY by ungrouped items** (grouped ones `continue` out
  at `:951`):
  ```go
  if itemHeight := effectiveBottom - effectiveTop; itemHeight > height {
      kind := "line"
      if len(it.Images) > 0 { kind = "image" }
      if len(it.Rects) > 0 { kind = "table" }
      return Pagination{}, &OverflowError{...}
  }
  ```
  `Kind` is a free-form `string`. **`"table"` is a variable, not a literal** — this is the wrong
  word the epic flags.
- `:250-256` — `OverflowError`; `:266-274` — its `Error()`, a **string concatenation**, not a format
  string: `"element " + e.ElementID + ": " + e.Kind + " is taller than the content window (" + …`.
- ⚠ **`:241-249` — A COMMENT ASSERTING A NEGATIVE THAT IS FALSE AT HEAD.** It claims `Kind`
  `"table"` is *"NO LONGER PRODUCED FROM PACKAGE FOLIO"* because every `tableRectSource` is built at
  one of three table constructors, all carrying a present `ItemGroup`. `collectElementBoxRects`
  (`folio-go/element_box.go:52`, appended `render.go:1653-1658`) is a **fourth** source and
  deliberately leaves the group zero (`element_box.go:103-107`). Measured: an untagged over-tall box
  element is a fatal error reading *"element e1: table is taller than the content window"*.
- `:143` `ColumnItem.Group`; `:157-168` `ItemGroup` (`Present` `:161`, `Key` `:167`); `:175-178`
  `ItemGroupKey`; `:500` `PageAssignment.Shift`. ⚠ `Key` is a **map key** at `:735`, `:851`, `:1094`
  — a new discriminator field belongs on `ItemGroup`, never on `ItemGroupKey`.
- `:335-341`, `:351-353` (`Pagination.Clipped` doc: *"An UNGROUPED over-tall item is not here"*) —
  both become stale.
- ⚠ **DW-64 sites in this file:** `:35-44` (the paraphrase; **`:42` reads *"the exception is scoped
  to grouped items"*, the exact assertion this story narrows**) and `:223-232` (the verbatim
  superseded AD-14 quote at `:224-226`).

### The keep-together namespace — `folio-go/render.go` (package `folio`)

- `:2006` `const keepTogetherKeyPrefix = "keepTogether:"`; `:2094-2099` `keepTogetherTagOf`.
  `:1961-2005` is an essay on why the prefix is a correctness device **precisely so `internal/layout`
  stays ignorant of it**. Read this before choosing where the discriminator lives.
- `:2017` `keepTogetherTags(t *Template) keepTogetherIndex`; `:2049` `keepTogetherGroup`; `:2069`
  `orKeepTogether` (guard `:2070-2072`, `if g.Present { return g }` — what leaves table rows alone).
- `:2120` `paginateDocument`. Appends rects → text → images (`:2158`, `:2196`, `:2256`); tags at
  `:2183`, `:2237`, `:2279`. ⚠ **`:2196-2250` builds ONE `ColumnItem` PER SHAPED LINE**, grouped by
  `(band, elementID, lineIndex)`, with `Top/Bottom` the **line's** extent and the group stamped on
  every line item.
- `:2186-2198` — a table data row's chrome rect spans the row's **entire** extent. ⚠ So for every
  over-tall table row *"some member is individually over-tall"* is **always true**
  (`paginate.go:611-616` records the equality as a construction invariant). **An unscoped
  any-member-over-tall rule reverses Story 4.6 wholesale.**
- `:2303-2306`, `:2325-2327` — D-4.2.2: the result is built *"straight from Paginate's OWN decision —
  never a second, independent re-run of the fit arithmetic"*.
- `:2461-2499` `clippedRowDiagnostic`; four role arms `:2477-2489`, keep-together arm `:2478-2480`.
  `c.ItemHeight` is the group **union** (`paginate.go:937`), so the surviving aggregate arm keeps a
  truthful message with no builder change.
- `:1691-1694` (PHASE A) and `:2298-2301` (PHASE B) both funnel through
  `paginateWithFooterOrphanFix` (`folio-go/table_footer.go:208`) and both `wrapOverflowError`.

### The other two pagination consumers

- `folio-go/page_number.go:87` `contentColumnItems` — PHASE A, the page **count**. ⚠ **`:89-110`
  emits one item per shaped line too**, `Group:` at `:104`. Both passes must agree.
- ⚠ `folio-go/page_setup.go:842` — the **canvas** calls `layout.Paginate` **directly** and
  *degrades* on any error (`:843-859`), setting `ContentWindowCountIsExact` false. A discriminator
  placed inside `Paginate` reaches the canvas automatically; one placed in
  `paginateWithFooterOrphanFix` does **not**, and the canvas would then confidently draw a clipped
  layout for a document the renderer refuses — reintroducing the canvas/render divergence Story 7.9
  just closed.

### Error plumbing — read-only

- `folio-go/render_error.go:132-141` `wrapOverflowError` (a direct type assertion, not `errors.As`).
- `CONTENT_UNLAYOUTABLE` minted `folio-go/internal/diag/diag.go:109`, re-exported
  `folio-go/diagnostic.go:207`. `TABLE_ROW_CLIPPED_HEIGHT` minted `internal/diag/diag.go:220`,
  re-exported `diagnostic.go:341`, emitted `render.go:2329`.

### The AST tripwire — `folio-go/table_row_clip_test.go`

- `:1081` `TestAPresentItemGroupIsATableRowOrAKeepTogetherGroup`; allowlist `:1096-1100`
  (`chromeRowGroup`, `lineRowGroup`, `keepTogetherGroup`); vacuity floor 3 at `:1186-1191`.
- It scans package-`folio` non-test sources for (a) any `*ast.AssignStmt` whose LHS is a selector
  named `Present` (`:1130-1147` — fires on the **field name alone**, any type, even assigning
  `false`) and (b) any `layout.ItemGroup{…}` literal setting `Present` to other than literal `false`
  (`:1149-1160`). **Reads are invisible to both.** Safe untag spelling if ever needed:
  `it.Group = layout.ItemGroup{}`.
- ⚠ **`:152-159` quotes AD-14 verbatim** — a fourth stale site DW-64 does not list, if AD-14's
  existing sentence is reworded rather than extended.
- `:1063-1080` asserts in prose that keep-together groups are *"a population D-4.6.2 has ruled
  clippable"* — true only of aggregate-over-tall groups afterwards.

### Existing tests — what moves and what must NOT

**The population this story moves has ZERO existing tests.** Searched `keep_together_fixture_test.go`,
`table_row_clip_test.go`, `table_footer_test.go`, `table_pagination_test.go`, `matrix_test.go`,
`canvas_window_count_test.go`, `byte_neutrality_test.go`, `render_test.go`, `render_overflow_test.go`
and `internal/layout/*_test.go`: **nothing asserts that an individually over-tall *tagged* element is
clipped.** DW-47 was reproduced ad hoc at Story 7.7's close and never captured. So the guard
obligation is discharged by **adding** the departed population under its new treatment, plus
re-pointing the **prose** listed below — there is no green assertion to invert.

**Aggregate-arm tests — UNAFFECTED, and they are the tripwires that catch an unscoped fix:**
- `keep_together_fixture_test.go:396` `TestKeepTogetherOverTallGroupIsClippedWithAWarning` — members
  ~15pt at y0 and y900; union 915pt vs 729.890pt. Aggregate-only.
- `keep_together_fixture_test.go:883` `TestKeepTogetherOverTallGroupDropsATaggedImage` — text 16pt +
  image 40pt at y900. Aggregate-only.
- `keep_together_fixture_test.go:273` `TestKeepTogetherSingleMemberChangesNothing` — member fits;
  ⚠ its header comment `:239-247` encodes matrix row 5's over-broad claim.

**Table-row family — must stay green; a red here means the discriminator is unscoped:**
`internal/layout/paginate_group_test.go:340`, `:556`; `table_footer_test.go:1341`;
`table_row_clip_test.go:186, 263, 363, 473, 561, 619, 762, 899, 1010`;
`table_pagination_test.go:662`. Every one has a member that is individually over-tall.

**Unaffected either way:** `table_row_clip_test.go:973` (pure unit test on `clippedRowDiagnostic`);
`table_footer_test.go:1404` `TestOverTallSingleRowStillOverflows` (ungrouped, already fatal).

### The documents to amend

- **`_bmad-output/planning-artifacts/epics.md:2302-2305`** — Story 7.7's third AC. `:2302` is
  exactly `**Given** a group taller than one window`; `aggregate` inserts after `window`. The 7.7
  section (`:2284-2311`) contains **no** matrix.
- **`_bmad-output/implementation-artifacts/7-7-keep-a-signature-block-together-across-a-page-break.md`**
  — the colliding matrix lives here, not in epics.md: `:177` row 3, `:179` row 5.
- **`ARCHITECTURE-SPINE.md:311-324`** (AD-14). Rule bullet `:316-324`; the carve-out wraps across
  `:319-320`: *"Rows and author-declared keep-together groups too tall for one content window (FR25,
  FR51), and clipped content (FR44), are `Warning`s returned alongside PDF bytes, never silent and
  never fatal."* File is **722 lines**. Discriminator clause confirmed **absent** (zero hits for
  "discriminator", "individually over-tall", "regardless of tagging").
- **Guards over the spine:** `lint/internal/rules/stagerank_test.go:253`
  `TestSpineStageLadderMatchesStageRankTable` reads only the region between
  `<!-- stage-rank-table:begin/end -->` at spine `:77`/`:91` — far above AD-14; no line numbers, no
  length assertion. `.claude/skills/bmad-architecture/scripts/lint_spine.py` parses AD headings and
  flags `TBD|TODO|FIXME|XXX`, `similar to AD-\d+` and `{template_token}`. **Adding a sentence inside
  AD-14's Rule bullet is safe against both.**
- **`_bmad-output/specs/spec-folio/folio-format.md`** — `:215` (the responsibility table row saying
  *any* declared keep-together group is clipped-and-warned), `:219-230` (the prose, whose `:223-224`
  states the rule unconditionally) and `:257` (*"A group taller than a whole content window is
  clipped, not refused"*). All three become false for the individually-over-tall case.
- **DW-64 sites:** `folio-go/diagnostic.go:343-348` (stale clause `:344`, *"over-tall rows (FR25, not
  yet built)"* — both halves wrong); `internal/layout/paginate.go:35-44` and `:223-232`;
  `internal/layout/paginate_group_test.go:325-339` (quote at `:331-332`).
- **DW-65 gate fires here:** `folio-go/internal/diag/diag.go:8-14` cites `ARCHITECTURE-SPINE.md:613`;
  spine `:613` is `COLUMN ||--o| AGGREGATE : "footer only"`, inside a mermaid `erDiagram` fence
  (`:602-615`). Its owner is *"the next story that edits `ARCHITECTURE-SPINE.md`"* — this story. The
  prescribed fix is to cite the **AD number** (AD-14 begins at `:311`), not to re-count lines.

### Golden surface — measured, nothing at risk

- `folio-go/byte_neutrality_test.go:92-515` `goldenDigestRecord`: **22 entries**, each pinning
  `fixtures/<dir>/expected.pdf` (hashed `:596-606`), declaring **64 recording sites** (22
  `expected.json`, 22 `second-literal`, 15 `readme`, 5 `signoff`).
- `grep -rl keepTogether fixtures/` returns **exactly two files**, both in `fixtures/keep-together/`
  (`input.folio`, `README.md`). Its three tagged elements are `e2` h=**16**, `e3` h=**1**, `e4`
  h=**16** against a **729.890 pt** window. The only tall element, `e1` h=660, carries **no tag**.
- The over-tall tagged constructions in the tree (`keepTogetherOverTallTemplate`
  `keep_together_fixture_test.go:367`, `keepTogetherOverTallImageTemplate` `:860`) are **test-only Go
  consts producing no digest**, and both are aggregate-only.
- ⚠ **727890 mp is a DIFFERENT geometry** (margins 30/42, header 18, footer 24) used by multi-page,
  three-band-page and `canvas_window_count_template.go:23` — not keep-together's window. Do not mix
  the two figures.

## Tasks & Acceptance

⚠ **This story is HALTED at its plan gate on the `multi-line text reach` Block If.** Task 1 below is
undetermined until the lead rules. Tasks 2-9 are settled and unaffected by that ruling.

**Execution:**

1. **BLOCKED** — `folio-go/internal/layout/paginate.go` -- narrow the clip branch at `:790-791` so a
   keep-together group containing an individually over-tall member is refused rather than clipped,
   while a table row and an aggregate-only group are untouched. -- The exact predicate depends on the
   `multi-line text reach` ruling; see Design Notes.
2. `folio-go/internal/layout/paginate.go` -- add a non-`Key` field to `ItemGroup` (`:157-168`)
   distinguishing a clip-eligible group from a keep-together group, set by the three existing
   derivations. -- ⚠ It must go on `ItemGroup`, **not** `ItemGroupKey`, which is a map key at `:735`,
   `:851`, `:1094`. This keeps the fit arithmetic inside `Paginate` (D-4.2.2) and reaches the canvas
   path automatically, which a pre-pass in `paginateWithFooterOrphanFix` would not.
3. `folio-go/internal/layout/paginate.go` -- give the fatal arm's `Kind` derivation (`:1076-1082`) a
   fourth value for an element box, so a `rect` is never announced as a "table". -- The epic flags
   this; DW-47 records it; measured verbatim at baseline.
4. `folio-go/internal/layout/paginate.go:241-249` -- correct the comment claiming `Kind` `"table"` is
   *"NO LONGER PRODUCED FROM PACKAGE FOLIO"*. -- It is false at HEAD: `collectElementBoxRects`
   (`element_box.go:52`) is a fourth, ungrouped `tableRectSource`. **A comment asserting a negative
   carries a test's evidentiary burden and must name the population it measured** (Epic 7 boundary
   obligation) — so name it, and assert it.
5. `folio-go/internal/layout/paginate.go:35-44`, `:223-232`, `:335-341`, `:351-353`, `:781-789` and
   `folio-go/internal/layout/paginate_group_test.go:325-339`, `folio-go/diagnostic.go:343-348` --
   re-point every stale AD-14 paraphrase. -- **DW-64 rides here**: this story is the one with the
   `internal/layout` fence lifted, which is the only cheap moment to bring all sites into line.
6. `folio-go/keep_together_fixture_test.go` -- add the two-arm discriminator: one geometry, one
   element set, two documents differing in exactly which element is tagged. Assert the fatal arm as
   an **error assertion** and the aggregate arm as bytes-plus-Warning, in adjacent tests that name
   each other. -- Follows the shipped `keepTogetherTemplateJSON` / `keepTogetherUngroupedTemplateJSON`
   pair, whose own doc comment calls it *"A DISCRIMINATOR, NOT A DEMONSTRATION"*. See Design Notes
   for why this is not a `fixtures/` directory.
7. `folio-go/table_row_clip_test.go:152-159`, `:1063-1080` and
   `folio-go/keep_together_fixture_test.go:239-247` -- re-point the prose that asserts the old,
   broader rule. -- Assertions stay green; the comments become false.
8. `_bmad-output/planning-artifacts/epics.md:2302` -- insert `in aggregate` after `taller than one
   window`. `_bmad-output/implementation-artifacts/7-7-...md:177,:179` -- tighten matrix rows 3 and 5.
   -- D-7.7.9: correcting the text is part of the ruling, not follow-up.
9. `ARCHITECTURE-SPINE.md` AD-14 Rule bullet (`:316-324`) -- append the discriminator clause.
   `_bmad-output/specs/spec-folio/folio-format.md:215`, `:219-230`, `:257` -- qualify the
   keep-together row and prose with "in aggregate". `folio-go/internal/diag/diag.go:10` -- cite
   **AD-14** instead of `ARCHITECTURE-SPINE.md:613` (**DW-65**, whose owner is the next story to edit
   the spine — this one). -- DW-49(b); half (a) already landed with Story 7.9 and must not be redone.

**Acceptance Criteria:**
- Given a content-band element with a declared box taller than one content window, tagged into a
  keep-together group, when the document is rendered, then it is refused with the same located fatal
  `OverflowError` the untagged element receives — no bytes, no `TABLE_ROW_CLIPPED_HEIGHT`.
- Given a two-member keep-together group whose members each fit one window but whose union does not,
  when the document is rendered, then Story 4.6's clip-and-warn applies unchanged — bytes plus one
  `TABLE_ROW_CLIPPED_HEIGHT` Warning located at the group's first member.
- Given an over-tall table data, header or footer row, when the document is rendered, then it is
  clipped and warned exactly as at baseline — no table-row test changes behaviour.
- Given the untagged over-tall element box, when the render is refused, then the message names the
  element's actual kind and never calls a non-table a "table".
- Given the amended documents, when Story 7.7's AC, AD-14, `folio-format.md` and the four AD-14
  paraphrases are read, then none contradicts the shipped discriminator, and all land in the same
  commit as the behaviour.
- Given the full `goldenDigestRecord`, when all 22 `expected.pdf` digests are recomputed, then every
  one is identical to baseline `9844e6d`.

## Spec Change Log

### 2026-08-31 — planned at `9844e6d`; halted at the plan gate on `intent gap`

The four behavioural claims the dispatch supplied were re-verified at `9844e6d` rather than taken on
trust (see Design Notes). Three of the dispatch's watch-items were settled and are recorded above;
the fourth — the reach of the ruling's mechanism over multi-line text — is the halt.

## Review Triage Log

## Design Notes

### Re-verified at `9844e6d` (the dispatch required this)

A throwaway module outside the repo, calling only exported API, then deleted; tree confirmed clean.
Byte counts differ from the register's because the probe documents are not byte-identical to the
ad-hoc ones — every **qualitative** claim reproduced exactly:

| case | result at `9844e6d` |
|---|---|
| over-tall `rect` (900pt), **tagged** | **578 bytes** + one `TABLE_ROW_CLIPPED_HEIGHT` Warning, group height 900000mp vs 729890mp |
| same `rect`, **untagged** | fatal `*folio.RenderError`: *"element e1: **table** is taller than the content window (900000mp against a content height of 729890mp)"* |
| ~60-clause text, **untagged** | **61,850 bytes, no diagnostic** — it flows cleanly |
| same text, **tagged alone** | **51,570 bytes** + Warning, group height **1005576mp** — content lost |
| two-member aggregate-only group | 21,186 bytes + Warning, group height 912892mp |

⚠ **An element box needs `style.background` or `style.border` to reach pagination at all**
(`element_box.go:52`, predicate `elementDeclaresBox`). A `rect` with neither renders 530 bytes and no
diagnostic however tall it is declared. Any fixture for this story must declare a box.

### The population D-7.7.9 does not reach — the halt

D-7.7.9's mechanism is *"every member fitting, the sum not"*. Measured above, a tagged multi-line
text element's group height is **1005576mp** — the union of its **lines**, because both pagination
passes emit **one `ColumnItem` per shaped line** (`render.go:2196-2250`, `page_number.go:89-110`).
Every line is ~15pt against a 729.890pt window, so **every member fits**. Applied literally, the
ruling classifies DW-50's document as *aggregate-only* and leaves it exactly as it is today — 51,570
bytes with content silently dropped — even though D-7.7.9 names DW-50 in its own heading and the
dispatch cites it as *"the same contract row on a second element kind"*.

The ruling's stated premise, *"a group of one is a no-op — there is nothing to keep it together
with"*, is **false for a multi-line text element**: there is something to keep together, namely its
own lines. That is precisely what DW-50 records, and DW-50 says it *"wants a ruling rather than a
patch"*.

Three dispositions, all defensible, differing observably and permanently:

- **(A) Member = column item.** Aggregate arm applies → clipped, content lost, unchanged from HEAD.
  DW-50 closes as "no change needed". But this contradicts *"a group of one must be indistinguishable
  from no group"* and matrix row 5, and leaves the story with no effect on the text kind at all.
- **(B) Member = template element.** A group spanning exactly one `ElementID` whose union exceeds the
  window is fatal → the tagged long text is refused. This is the literal reading of *"an over-tall
  individual element is fatal, tagged or not"*, it **narrows what renders** (so it fits the
  before-the-tag set that D-7.7.13 and D-7.8.3 put this story in), and it is implementable: the
  distinct-`ElementID` count is available on `ColumnItem` at the deciding site. But it refuses a
  document that untagged renders perfectly, and the element's height comes from its **content**, not a
  declared box — so D-4.6.2's authorship ratio arguably points the other way.
- **(C) A single-member group is dropped before pagination.** Then AC1's own wording — *"the same
  located fatal `OverflowError` **the untagged element receives**"* — holds literally in both cases:
  the over-tall box is fatal, the long text flows cleanly. This satisfies row 5 and the no-op
  sentence and fixes both DW-47 and DW-50. But it **widens** what renders (a document that clips
  today would flow, changing bytes), which is the opposite of the narrowing this story is scheduled
  before the tag to perform — and it would also disable keep-together as a way to hold one element's
  lines together, which is a real thing to want.

**Recommendation: (B)**, on the ground that the intent classifies this story by what it does — *"it
changes when a render fails: documents that render today — clipped, with a true Warning — will fail
fatally afterwards"* — and only (B) does that for the text kind. But (B) contradicts AC1's
untagged-equivalence wording, so a builder choosing it would be silently amending the ruling. The
lead reasoned about exactly this hazard (the no-op sentence), so correcting its premise may change
the ruling. **Halting rather than picking.**

### Why the discriminator must be scoped, and where it must live

For a table data row the chrome rect spans the row's **entire** extent (`render.go:2186-2198`;
`paginate.go:611-616` records the equality as a construction invariant), so *"some member is
individually over-tall"* is **always true** of an over-tall row. An unscoped rule therefore reverses
Story 4.6 wholesale. The discriminator must additionally require that the group is a **keep-together**
group.

That knowledge is a package-`folio` concept (`keepTogetherKeyPrefix`, `render.go:2006`), and
`render.go:1961-2005` argues at length that `internal/layout` must stay ignorant of it. Hence task 2:
carry the distinction as a **field on `ItemGroup`** set by the three existing derivations, rather than
having layout sniff the key prefix or having package `folio` re-run the fit arithmetic before calling
`Paginate`. A folio-side pre-pass was considered and rejected on two grounds: it is *"a second,
independent re-run of the fit arithmetic"*, which `render.go:2303-2306` forbids (D-4.2.2); and the
canvas calls `layout.Paginate` **directly** (`page_setup.go:842`), so a pre-pass would leave the
canvas confidently drawing a clipped layout for a document the renderer refuses — reintroducing the
canvas/render divergence Story 7.9 closed one story ago.

**This is why the implementation touches `internal/layout`, and therefore why DW-64 rides here.**

### Why "one fixture carrying both arms" is not a `fixtures/` directory

The guardrail is mandatory, and the reasoning behind it is that either arm alone is also consistent
with the rule being replaced. But **one rendered document cannot carry both arms**: AD-14 makes an
`Error` abort the render and a `Warning` accompany a successful one, so a document that fatally
refuses produces no bytes and cannot also produce a clip Warning. That is not a preference — it is
arithmetic on AD-14's own definitions.

The fixture machinery agrees: every `expected.json` in the repo carries exactly three keys
(`folioGoVersion`, `goToolchain`, `sha256`) and three separate sites hard-require a non-empty,
page-tree-resolving `expected.pdf` for anything in `goldenDigestRecord`
(`byte_neutrality_test.go:596-612`, `golden_structural_validity_test.go:100-110`, the per-fixture
test). Representing a refusal as a fixture would mean extending that schema — new machinery, not
registration. And the repo has already ruled on exactly this pathology:
`folio-go/render_overflow_test.go:5-12`, citing D-2.6.5's guardrail and D-000.50's finding, says of an
over-tall content-band element that *"no committed fixture can express this case, and none should be
distorted to. … The pathology belongs here, where it can be stated as an error assertion rather than
as bytes."*

So "one fixture" is satisfied as **one fixture subject**: one geometry and one element set, two
documents differing in exactly which element is tagged, asserted in adjacent tests. The shipped
`keepTogetherTemplateJSON` / `keepTogetherUngroupedTemplateJSON` pair is the established idiom, and
its own doc comment states the purpose in the same terms the guardrail uses — *"IT IS A
DISCRIMINATOR, NOT A DEMONSTRATION"*, with a twin *"that makes that difference measurable rather than
asserted"*. **This ruling is recorded so the lead can reverse it cheaply if it disagrees.**

### On `multiple-goals` — considered and declined

The behaviour change, the two-arm assertion and the five document amendments share one subject: the
disposition of an over-tall element. D-7.7.14 states half (b) *"is not a second subject"*, and
D-7.7.12 confirms 7.10 in Epic 7 as a repair. The route-out test this project uses requires the
routed work to carry a decision above a builder's authority; DW-64 and DW-65 carry none — both are
already ruled, and both are gates that fire precisely because this story lifts the `internal/layout`
fence and edits the spine. Declined, recorded rather than silent.

### Which currently-rendering documents stop rendering

Documents containing a content-band element that **declares a box** (`style.background` or
`style.border`) taller than one content window **and** carries a `keepTogether` tag. Enumerated
against the repo at `9844e6d`: **none.** No shipped fixture, no fixture-backed Go const and no
existing test contains that construction — the only `keepTogether` tags in `fixtures/` are
`fixtures/keep-together/`'s three elements at heights 16, 1 and 16 pt against a 729.890 pt window,
and the two over-tall Go consts are aggregate-only and produce no digest. The affected population is
real but **entirely unrepresented in the repository**, which is why all 22 goldens survive and why
this change costs nothing today and would cost every downstream suite after the tag. Whether DW-50's
long-text document joins this list is the open question above.

## Verification

**Commands** (the D-R7.1 heavy set; from the repository root unless stated):

- `cd folio-go && go test -count=1 ./...` -- expected: exactly **one** failure, the mandated red
  `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` (got 7, need >=20). Report measured
  pass/fail counts, never "green".
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `gofmt -l folio-go` -- expected: no output. (`lint/…/licencegraph_test.go` has a known gofmt
  finding, DW-23, outside this path.)
- `cd folio-go && FOLIO_MATRIX_TARGET=<t> go test -tags=matrix -count=1 -run TestTargetRenderHash -v .`
  -- once per target for `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`, each **exported**.
  ⚠ **Plus an unset control** confirming it is a deliberate no-op — that control is what proves the
  four legs were not no-ops. Say which legs ran.
  `TestShippedFacesReproduceFromUpstream` fails/skips without `fontTools`; environmental.
- `cd folio-go && go test -tags=matrix -count=1 -run TestCrossTargetByteIdentity -v .` -- expected: pass.
- `cd lint && go test -count=1 ./...` -- ⚠ **`-count=1` is MANDATORY.** The `rules` package walks
  `folio-go` as a directory and Go's cache does not track `ReadDir`, so a new file never invalidates
  it; this gate returned a **cached** green while two tests were red (D-7.9.5). Expected: 4 packages
  `ok`. A cached green is no measurement at all.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` --
  expected: typecheck clean; lint with **exactly 4** pre-existing `only-export-components` warnings;
  vitest **284 tests / 34 files** as the arithmetic baseline — a number to report, asserted by no
  in-tree guard; e2e compile-only (D-000.4). **This story changes zero designer paths.**
- `shasum -a 256` over all **22** `expected.pdf` paths in `goldenDigestRecord`
  (`byte_neutrality_test.go:92-515`) -- expected: **every digest identical to baseline `9844e6d`**.
  Any movement is a HALT, never a re-record.

**Mandatory proofs** (report each verbatim; "red-proved" as prose is not evidence):

- **Deletion is the cheapest screen, and the implementer must not be the only one choosing the
  mutation** (Epic 7 boundary obligation). Delete the discriminator outright and confirm the new
  fatal-arm test reddens; separately delete the **scoping** clause and confirm the table-row family
  reddens. Each in isolation — a combined mutant hides which half is covered.
- **Both arms, both directions.** The fatal arm must redden if the discriminator is neutralised; the
  aggregate arm must redden if the discriminator is over-broadened to catch it.
- **The table-row tripwire.** `internal/layout/paginate_group_test.go:340`, `:556`,
  `table_footer_test.go:1341` and the `table_row_clip_test.go` family must stay green **unmodified**.
  A red there means the discriminator is unscoped and reverses Story 4.6.
- **PHASE A in isolation** (`page_number.go`). Per the standing finding, removing a `contentColumnItems`
  substitution has left the entire suite green before, because almost every fixture declares empty
  `pageHeader`/`pageFooter` so nothing reads the page count. Mutate the PHASE A path alone and report
  what reddens; if nothing does, say so.
- **The `Kind` fix.** Re-run the untagged over-tall element-box probe and quote the message; the word
  "table" must be gone. Assert it, do not only observe it.
- ⚠ **Assert the mutation landed before concluding anything from it.** Count occurrences of the
  substring being replaced — a doc comment carrying the same expression has silently absorbed a
  mutation in this repo before.

**Manual checks:**

- `git status --porcelain` never shows `README.md`; its md5 stays `078d7d80d518d54af2fc04fb270d46b8`
  at 8,470 bytes. Stage explicit paths only — never `git add -A` or `git add .`.
- `fixtures/statement-signoff.json` unmodified; `reader`, `date`, `examined` untouched.
- `git diff --stat` shows **zero** paths under `folio-designer/`.
- The behaviour change and all five document amendments are in the **same commit** (D-7.7.14).
- `ARCHITECTURE-SPINE.md` still parses: `<!-- stage-rank-table:begin/end -->` at `:77`/`:91` intact,
  and `cd lint && go test -count=1 ./...` green.

## Auto Run Result

### Dispatch 1 — 2026-08-31, plan only (`Halt after planning.`)

Status: **blocked**
Blocking condition: **intent gap**

Planned at baseline `9844e6d`; no implementation code written, nothing committed. The spec's Code Map,
Tasks, Design Notes and Verification are complete and were derived from the code at that revision,
with the dispatch's four behavioural claims re-measured rather than assumed.

Settled at this gate and recorded above: the golden impact (none — all 22 digests safe), the
`internal/layout` question (yes, touched, so DW-64 rides here, and DW-65's gate fires too), the
scoping trap that would otherwise reverse Story 4.6, and the reading of "one fixture carrying both
arms".

Halted on the one question the intent does not settle: whether D-7.7.9's per-member mechanism reaches
a tagged multi-line text element, whose members are its individual lines. See Design Notes, "The
population D-7.7.9 does not reach", for the measurement, the three dispositions and a recommendation.
