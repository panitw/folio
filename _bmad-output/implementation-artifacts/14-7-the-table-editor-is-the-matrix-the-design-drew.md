---
title: 'The table editor is the matrix the design drew'
type: 'feature'
created: '2026-09-10'
status: 'done'
baseline_commit: 'c747023886e09971e7ee766847ae1341f13752ad'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

Configuring a table's columns today means filling in eleven fields per column, in a grid so wide it
scrolls sideways. This rebuilds it into the six-column matrix the design draws — number, header,
bound field, width, alignment, footer aggregate — with reorder and remove as small affordances on
each row instead of four more columns. The alignment picker becomes the same three-segment control
the inspector already uses, so there is one alignment control in the product. The three separate
footer fields become one, with source and format appearing only when the aggregate needs them. And
the dialog gains two read-outs it should always have had: a width budget that says whether the
columns fit, and a header saying which collection, how many sample rows and which band you are
editing. Nothing about the document, the file format or a rendered byte changes.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The table editor is an eleven-column form — Header, Width, Cell alignment, Row field,
Footer, Footer source, Footer format, Move earlier, Move later, Remove, Add after — whose grid is
`min-width: 1320px; overflow-x: auto`, so the densest surface in the product (UX-DR8) is reached by
horizontal scrolling. Four of those columns are row actions wearing column headers, three are one
concept spread across three fields, and alignment is a `<select>` here while the inspector two
panels away uses a segmented control for the same concept. The author cannot see whether the columns
fit the page until the render fails, and cannot see which collection, how many sample rows or which
band the table belongs to at all.

**Approach:** Rebuild the matrix as the six columns the design draws — `#`, HEADER LABEL, BOUND FIELD
· row scope, WIDTH, ALIGN, FOOTER AGGREGATE — with reorder and remove as unlabelled row affordances
and a single Add column control below the grid. Extract the inspector's existing segmented control
into a module both surfaces consume, so the ALIGN cell is literally the same control. Collapse the
three footer fields into one aggregate control that reveals source and format only when the chosen
aggregate needs them, changing no engine rule. Add a read-only scope header, a width budget and a
footer summary, all derived from data the browser already holds. This is presentation only: no
format change, no version increment, no new engine surface, no new projection field, and not one
rendered byte moves.

## Boundaries & Constraints

**Always:**

- **The display unit is points.** `D-14.2.Q3` ruled this product-wide and it is settled: widths
  render in **pt**, one decimal. The mockup's `mm` is mockup fidelity against zero millimetres in the
  product. The stored value is millipoints either way and is untouched.
- **Every new read-out is derived browser-side from data already in scope.** The item count is
  `SampleNode.count`; the band is `CanvasProjection.components[].band`; the available width is
  `band.width − table.x`, which is the engine's own rule in `containComponent`, and a table's
  projected width is already Σ column widths. **No projection field is added and no Go file is
  touched.** (`epic-14-context.md` claimed this story needed a sample-item-count projection field; the
  claim was measured false and corrected at `c747023`. Going through Go would have cost a six-file
  `hasExactKeys` wire dance whose failure mode is a silently terminated worker.)
- **The collection and the row alias stay editable inside this dialog.** `DW-351` records this as the
  only editable site for a table's collection in the entire product, and Story 14.6 has just started
  routing authors here. They are restyled as a labelled group; they are not moved, and they do not
  become read-only.
- **Commit-on-blur is unchanged**, and every refusal stays live and located where it is raised.
- **The keyboard must not lose reach.** Every control the eleven-column matrix could reach by arrow
  keys and Home/End must remain reachable by arrow keys and Home/End (UX-DR25).
- **One alignment control in the product.** The table's ALIGN is the inspector's segmented control,
  not a copy of it.
- **Column align is `left`/`center`/`right` — never `justify`.** `ColumnAlignTokens` excludes it
  while `StyleAlignTokens` includes it; the table must never receive the four-segment array.
- **The segmented control must not inherit `.matrix-row button`.** That rule is `(0,1,1)` and sets
  `border`, `border-radius`, `background`, `color` and `font: var(--type-mono)`; `.property-segment`
  is `(0,1,0)` and loses, while `.property-segment[aria-pressed="true"]` is `(0,2,0)` and still wins
  on background — the pressed fill inside the wrong border, half-broken and invisible to jsdom. This
  is a must-handle, and it is proved in a browser spec, not by reading the stylesheet.
- **A control not traceable to a field the engine consumes is absent, not disabled-and-mysterious.**
- **Anything unpickable, unavailable or refused states its reason.** No bare grey-out, no empty frame
  that reads as a rendering failure.
- Design tokens only. Square corners, 1px borders, mono for machine-read values, sans for prose.

**Ask First:**

- Any change to a Go file, the wire projection, the `.folio` format, or a command's shape.
- Adding drag-and-drop reordering, or any interaction the story does not name.
- Any change that removes an author's ability to set a value the engine accepts.
- Any finding that the frozen intent above is wrong.

**Never:**

- **Never build the Cancel / Done discard.** It is Story 14.7b under `D-14.7.1`. **Build the footer
  bar** — it carries the `5 columns · 2 aggregates` summary on the left and **today's `Close Table
  Editor` button** on the right, moved down out of the heading; 14.7b replaces that one button with
  the `Cancel` / `Done` pair. Do not add a counter, do not issue an undo, do not touch `applyHistory`.
- **Never fix `DW-368`** (global Cmd+Z reaching the document behind the open modal). It belongs with
  the counter it would desynchronise, in 14.7b.
- Never build the mockup's HEADER / CELLS / BORDERS sections (Story 14.8), its `Show header row` or
  `Repeat on continuation pages` checkboxes, its `Padding` field (an owner ruling forbids the panel
  authoring padding on any element kind), its `Row height` select or its borders preset. None has a
  format field behind it.
- Never write a literal hex, `rgb(`, `hsl(` or a non-token `border-radius` into `App.css`.
- Never let the browser measure text or read layout — no `getBoundingClientRect`, no `offset*`,
  `client*`, `scroll*`, no `ResizeObserver`, no `measureText`.
- Never claim a browser spec is covered by CI. Say it was compiled, or give the counts from a run.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Six columns, no more | A table with columns | The grid carries exactly the six labelled headers `#`, `HEADER LABEL`, `BOUND FIELD · row scope`, `WIDTH`, `ALIGN`, `FOOTER AGGREGATE`; no `Move earlier`, `Move later`, `Remove` or `Add after` columnheader exists | N/A |
| Reorder and remove | Any row | Reachable as row affordances carrying accessible names, not as labelled columns; both keyboard-operable | Disabled at the ends, with the end stated |
| Add a column | Add column pressed | One control below the grid appends a column; position is reached by reorder | Engine refusal rendered in the dialog's alert |
| Alignment | A column's ALIGN cell | The same three-segment control the inspector uses, `left`/`center`/`right` only | N/A |
| Aggregate `none` | `column.footer === ''` | One control reading `none`; **no** source field and **no** format field exists in the row | N/A |
| Aggregate `count` | `column.footer === 'count'` | Source is absent (count takes none); format is revealed | N/A |
| Aggregate `sum`/`avg` | `column.footer === 'sum'`/`'avg'` | Source and format both revealed | Engine refusal rendered in the alert |
| A revealed cell vanishes under focus | Focus in the source field, aggregate set to `none` | Focus lands on a surviving cell in the same row, never on `document.body` | N/A |
| Budget exact | Σ widths === available | `Σ 174.0 of 174.0 available` with the `exact` badge | N/A |
| Budget under | Σ widths < available | The sum, the available figure, and the shortfall stated; no `exact` badge | N/A |
| Budget over | Σ widths > available (a loaded file — the loader does not bound it) | The overflow stated in the author's terms before rendering, not after | N/A |
| Scope header | Sample loaded, count known | `transactions[] · 34 items in sample · band: content`, read-only | N/A |
| Scope header, no sample | `sampleAvailable === false` | The collection and band still stated; the count says it is unknown rather than showing `0` | N/A |
| Scope header, truncated node | Collection node has no `count` | The count says it is unknown rather than fabricating a number | N/A |
| Footer summary | Any column set | `5 columns · 2 aggregates`, aggregates counting columns with a non-empty footer | N/A |
| No columns | `columns.length === 0` | Today's empty state survives, with its stated reason and its Add column control | N/A |
| Keyboard reach | Any row | Arrow keys reach every enabled control in the row from its first cell; Home/End reach the row's first and last enabled controls; ArrowUp/ArrowDown move between rows | Absent and disabled cells are skipped, not landed on |

</frozen-after-approval>

## Code Map

Anchors re-measured at `cee83bf`. Nothing here was carried from the epic without checking.

- `folio-designer/src/TableEditor.tsx` (259 lines) — **the subject.** `cellCount = 11` (`:12`); the
  eleven `<span role="columnheader">` (`:198`); the row's eleven cells with `aria-colindex={1..11}`
  and `matrixCell(index, 0..10)` (`:199-210`); the scope inputs and the `local discovery hints only`
  note (`:194`); `focusCell` (`:39-49`); `moveFocus` (`:56-72`); `trapDialog` (`:73-81`);
  `dispatchOnce` (`:106-109`); `matrixCell` (`:191`). Its `HEADER AND ROWS` section (`:230+`) and its
  header-style controls belong to Epic 12 and must survive byte-compatible in behaviour.
- **The absent-cell semantics — reuse this, it is why revealed-on-demand fields are cheap.**
  `moveFocus`'s `enabled()` is `…querySelector(…)?.matches(':disabled') === false`, so an **absent**
  cell yields `undefined === false` → `false` and is skipped exactly like a disabled one. **But
  `focusCell` is NOT symmetric:** `!preferred?.matches(':disabled')` is `true` when `preferred` is
  `undefined`, so `target` is `undefined` and it returns early instead of falling back. A revealed
  cell that disappears while focused therefore strands focus on `document.body`. That is the
  I/O matrix's "revealed cell vanishes under focus" row, and it will ship green if untested.
- `folio-designer/src/App.tsx` — `SegmentedProperty` (`:4342-4356`, module scope, **not exported**);
  `SegmentSpec` (`:2918`); `AlignIcon` (`:2925-2927`); `alignGlyphs` (`:2921`); `alignSegments`
  (`:2929`); `justifySegment` (`:2935`); the three/four widening at `:2986-2990`; both invocations on
  `:3018`. **`App.tsx:33` imports `TableEditor`, so importing back is a cycle** — the control must
  move to a new module. Every consumer today is inside `App.tsx`, so the move touches one file.
  Also: `tableSampleCandidates` (`:62-82`, slices to 50 and drops `count`); the `<TableEditor>` call
  site (`:2551`), which already has `sampleData` and the canvas projection in scope.
- `folio-designer/src/sample-data.ts` — `SampleNode.count` (`:12`) is the **true** item count, not
  the truncated one (`:113-127`: `count++` runs on every item while only `SAMPLE_LIMITS.items` = 5
  are kept as children). It is `undefined` on a truncated node — absence means unknown, never zero.
  The `⚠` comment at `:14-33` records that a collection's `segments` has a second reader here; do not
  disturb it.
- `folio-designer/src/engine-protocol.ts` — `TableColumn` (`:214`), `TableColumns` (`:256`),
  `isTableColumns` (`:449-497`, `hasExactKeys` in **both** directions). Read-only: touching it means
  touching Go, which this story does not do. `canvas.bands[]` and `canvas.components[]` (`:341-342`)
  carry `width`, `x` and `band` — the width budget and the band come from here.
- `folio-go/component_commands.go:2115-2124` `containComponent` — **the authority for the budget**:
  `width > geom.Length(band.Width)-x`. With `projectedSize` (`:2032-2041`) making a table's width Σ
  column widths. **Read-only evidence; do not edit.** Available width is `band.width − table.x`, not
  `band.width`.
- `folio-designer/src/App.css:657` — `.table-matrix { min-width: 1320px; overflow-x: auto }` and the
  **eleven-track** `.matrix-header, .matrix-row { grid-template-columns: … }`. **And the hazard:**
  `.matrix-row button` is `(0,1,1)` and sets `border`, `border-radius`, `background`, `color` and
  `font: var(--type-mono)`. `.property-segment` is `(0,1,0)` and **loses**, while
  `.property-segment[aria-pressed="true"]` is `(0,2,0)` and still wins on background — giving the
  pressed fill inside the wrong border. `App.css:35-37` records this exact defect class and the
  idiom for fixing it. jsdom cannot see any of it.
- `folio-designer/src/App.css:633-637` — `.property-segmented`, `.property-segment`,
  `.property-segment[aria-pressed="true"]`, `:disabled`, `.segment-icon`.
- `folio-go/internal/template/closedsets.go:104,119,124` — `StyleAlignTokens` (four, with `justify`),
  `TableStyleAlignTokens` (three), `ColumnAlignTokens` (three). Read-only evidence.
- **Tests that will red, with why:**
  - `folio-designer/src/App.test.tsx:294` and `:705` — `aria-colcount '11'`. The test at `:701` also
    says "eleven columns" in its **name**.
  - `folio-designer/src/App.test.tsx:336` — asserts the immediate DOM successor of the active matrix
    cell is `Header height in points`. Any enabled control added between the last row-1 cell and the
    HEADER AND ROWS section reds it. `:341`/`:343` assert both ends of the trap wrap (D-12.3.2).
  - `folio-designer/e2e/table-editor.spec.ts:16` (`aria-colcount '11'`) and `:19-31` (the roving walk
    asserting Row field at index 3, Footer aggregate at 4, Add after at 10, Remove at 9).
  - `folio-designer/e2e/browser-native-roundtrip.spec.ts:207-264` — **the golden round trip authors
    through this dialog**: `Header for column N`, `Row field for column N`, `Footer aggregate for
    column 5` (`selectOption('sum')`), `Footer source for column 5` (`.fill(…)`). If the reveal
    changes when the source field exists, this must be resequenced, not deleted.
  - `folio-designer/src/control-vocabulary-contract.test.tsx` — the four floors at `:485-489`
    (`PER_STATE_CONTROL_FLOOR 15`, `CONTROL_FLOOR 170`, `CLASS_FAMILY_FLOOR 17`,
    `GROUP_INSTANCE_FLOOR 25`) and `CHECKED_GROUPS` at `:461-474`.
- **`control-vocabulary-contract.test.tsx` — three mechanical facts that decide how the new state is
  added.** (1) `states` (`:309-354`) declares `open: () => Element`, **synchronous**, and
  `sweepEveryState()` is called inside each sync `it`. Opening this dialog needs an awaited
  `table-columns` reply, so the state's `open` and its callers must become async. (2) The file's
  `engine()` mock (`:302`) answers `{ snapshot: { revision: 2 } }` with no canvas and no
  `tableColumns`; it must be taught the full 20-member + `columns` projection at a **matching
  revision** or the dialog never mounts and the state sweeps nothing. (3) ⚠ **`:761` is
  `sweepStates(states.slice(0, -1))`** and asserts an exact message list naming `PDF navigation` and
  `Render actions` — it drops the **last** state, meaning `preview`. **Insert the Table Editor state
  BEFORE `preview`.** Appending it would make that red proof drop the new state instead and prove
  something other than what it claims — the comment at `:756-760` warns about precisely this, having
  been bitten once already at Story 14.2.
- **Tests that will NOT red, checked so nobody chases them:** `property-prose-height.test.ts` (every
  regex anchored to `.property-value-prose`); `design-contract.test.ts` (reads App.css but names no
  table selector — only its whole-file hex and radius rules apply); `command-json-soleness.test.ts`
  (`:151` whitelists the `data-matrix-cell` template as a negative control — keep that spelling);
  `binding-vocabulary.test.tsx` (inspector-side only); the population floors in
  `canvas-authority-contract.test.ts:762-766` (they count **files**, so adding modules and a test
  file is free; only deleting `TableEditor.tsx` outright would red `:765`).

## Tasks & Acceptance

**Operational constraints for the implementer (D-14.4.1 — these are binding):**

- **Do not `git add`, commit, stash, checkout, reset, revert, restore, `git clean`, push or branch.**
  Reading git state is fine.
- **Do not edit** `sprint-status.yaml`, `deferred-work.md`, `epics.md`, `DESIGN.md`, or anything under
  `fixtures/declared-variants/`. Do not run `code -r`.
- Shell is **zsh**: `${PIPESTATUS[0]}` is empty. Use `cmd > log 2>&1; echo $?`. Never `$?` after a
  pipe. Quote every glob and variable.
- `find` wraps `bfs` and is non-deterministic here (`D-14.6.1`) — use `git ls-files`. A `find` miss is
  never evidence of absence.
- `App.tsx` holds two NUL bytes near line 3997: plain `diff` prints "Binary files differ" with zero
  changed lines. Use `diff -a` and `cmp`.
- **Build the ArrowUp/ArrowDown and disabled-skip proofs FIRST, against the eleven-column matrix, and
  watch them pass — then rebuild.** Today every fixture in the repo has exactly one column, so the
  vertical loop cannot step and the skip branch is never exercised. "The navigation survived" is
  otherwise a comparison against nothing.
- **Re-measure the four vocabulary floors and the new census by EXECUTING the sweep and reading what
  it reports.** Do not hand-write either. Same for `:761`'s expected message list.

**Execution:**

- [x] `folio-designer/src/TableEditor.test.tsx` -- new dedicated test file, built **before** the
      rebuild -- mirror the fixture at `App.test.tsx:284-293` and the `queries`-counter shape at
      `:389-399`, but with a **multi-column** table (at least three columns, with differing footers),
      and prove ArrowUp/ArrowDown and the disabled-cell skip against the current matrix. This is
      `DW-351`'s actual content.
- [x] `folio-designer/src/segmented-control.tsx` -- new module holding the presentational segmented
      control plus the align glyphs and segment data -- because `App.tsx:33` imports `TableEditor`, so
      the reverse is a cycle. **State in the module's own comment that a table column takes the
      three-value array and never `justify`** — not only in this spec.
- [x] `folio-designer/src/App.tsx` -- consume the extracted control, leaving `SegmentedProperty`'s
      press-again-to-clear semantics on the inspector side where they belong (a column's align is not
      clearable) -- keep both invocations at `:3018` behaviourally identical.
- [x] `folio-designer/src/TableEditor.tsx` -- rebuild the matrix to six labelled columns with reorder
      and remove as row affordances, one Add column control below the grid, the ALIGN segmented
      control, the collapsed FOOTER AGGREGATE with its revealed source and format, the read-only scope
      header, the width budget and the footer summary; restyle the collection/alias group; **fix the
      `:194` note**, which still calls the candidates "local discovery hints only" — wording written
      when nothing directed authors here, and 14.6's context bar now does.
- [x] `folio-designer/src/App.css` -- six-track grid, the affordance tracks, the budget and summary
      rules -- **and neutralise the `.matrix-row button` (0,1,1) hazard against the segmented
      control**, following the `App.css:35-37` idiom (scope the exclusion or add a same-or-higher
      specificity rule after `:657`; do not bump `.property-segment` globally).
- [x] `folio-designer/src/control-vocabulary-contract.test.tsx` -- add a Table Editor render state
      **before** `preview`, make `open` and its callers async, teach the mock a valid `table-columns`
      reply, add the new group to `CHECKED_GROUPS`, and re-baseline all four floors and the census
      from an executed sweep.
- [x] `folio-designer/src/App.test.tsx` -- update `aria-colcount`, the `:701` test name, and the
      D-12.3.2 successor/wrap assertions at `:336`/`:341`/`:343` to the new shape. ⚠ **Moving `Close
      Table Editor` out of the heading and into the footer bar re-orders `trapDialog`'s focusable
      list deliberately, exactly as D-12.3.2 re-ordered it once before.** `Close` stops being first
      and becomes last, so the forward wrap is now from `Close` to the dialog's new first control and
      the backward wrap is its inverse. Re-derive both ends from the DOM as `:335` already does, and
      **say in the test that the new order is intended** — an unexplained re-ordering reads as a
      regression to the next author.
- [x] `folio-designer/e2e/table-editor.spec.ts` -- rewrite the roving walk for the six-column lattice.
- [x] `folio-designer/e2e/table-matrix-layout.spec.ts` -- new browser spec for **what jsdom cannot
      see**: that the six-track grid no longer needs horizontal scrolling at the dialog's width, that
      the budget meter reflects the sum, and that each ALIGN segment renders as one joined segmented
      control rather than three bordered mono buttons (the specificity hazard, observed).
- [x] `folio-designer/e2e/browser-native-roundtrip.spec.ts` -- resequence `authorTableWithFooter`
      (`:207-264`) for the revealed source field. **Do not weaken it** — it is the only end-to-end
      proof that authoring reaches the engine and survives the round trip.

**Acceptance Criteria:**

- Given the rebuilt dialog, when the grid renders, then exactly six `columnheader` elements exist,
  spelled as the design spells them, and no columnheader named `Move earlier`, `Move later`, `Remove`
  or `Add after` exists — **proved by adding each of those four back, at container, row and cell
  position in turn, and showing the fence reds at every position** (`D-14.5.1`; a `within(x)` sweep
  cannot see a violation **on** `x`, and role queries exclude `aria-hidden` subtrees by default).
- Given the ALIGN cell, when it renders, then it is the same segmented control module the inspector
  renders, offering exactly `left`/`center`/`right` — **proved by adding a `justify` segment and
  showing it reds**, not by reverting the control.
- Given a column whose aggregate is `none`, when the row renders, then no footer source and no footer
  format control exists anywhere in that row — **proved by adding each back at every position**.
- Given a column whose aggregate is `count`, when the row renders, then a format control exists and a
  source control does not.
- Given focus in a revealed source field, when the aggregate is set to `none` so that cell is
  removed, then focus lands on a surviving cell in the same row, not on `document.body`.
- Given any row, when arrow keys are driven from its first cell, then every enabled control in the row
  is reachable; Home and End reach its first and last enabled controls; ArrowUp and ArrowDown move
  between rows; and absent and disabled cells are skipped rather than landed on.
- Given columns summing to less than, exactly, and more than `band.width − table.x`, when the budget
  renders, then each of the three states is stated in points, and only the exact case carries the
  `exact` badge.
- Given a loaded sample whose collection node carries a count, when the scope header renders, then it
  states the collection, that count and the band; and given no sample or a truncated node, it says the
  count is unknown rather than showing a number.
- Given the collection and row alias, when the dialog renders, then both are still editable and still
  commit through `configureTableBindingCommand` — **proved by a test that reads the committed value
  back**, not by counting dispatches.
- Given the dialog's footer bar, when it renders, then it states `{n} columns · {m} aggregates` with
  `m` counting columns whose footer is non-empty, and carries `Close Table Editor`; and the dialog's
  focus trap still wraps at both ends, over the re-ordered list, with the invoking control restored
  on Escape.
- Given the rebuild, when the suite runs, then the Epic 12 HEADER AND ROWS section behaves exactly as
  before, including its busy-restore, its `badInput` refusal and its one-command-per-burst property.

## Spec Change Log

## Design Notes

**Why revealed-on-demand fields are nearly free, and where the trap is.** `moveFocus`'s `enabled()`
already treats an **absent** cell exactly like a disabled one, so a fixed cell lattice covering the
maximal row shape needs no per-row arithmetic — a row with no footer simply has holes, and the scan
walks past them. The asymmetry is in `focusCell`, which returns early on an absent cell instead of
falling back the way it does for a disabled one. That single difference is the entire "revealed cell
vanishes under focus" hazard, and it is invisible to any test that does not remove a cell while it
holds focus.

**The reveal predicate is today's disabled predicate, unchanged.** Source is live when the aggregate
is set and is not `count`; format is live when the aggregate is set. Turning those from `disabled`
into presence is the whole of AC3 — the engine's rules do not move, and this story only stops
presenting three fields where the design presents one.

**Where the design is silent, and it is silent rather than wrong.** The mockup draws no revealed
source or format at all — its aggregate cell is a closed chip whose only drawn values are `none` and
`sum` — and it draws only the `exact` budget state. Both are needed and neither is transcribable, so
both are designed here and recorded as silence. The mockup also draws the collection as read-only
plain text and draws **no row alias anywhere** — so AC5's original "edited where the design puts them"
had no referent, and the AC was corrected at `599ae62` to "stay editable in a labelled, restyled
group", which is what this spec builds. Add-after leaves the row: a
single Add column appends, and a position is reached by reorder — a step longer, not a capability
lost. Reorder stays keyboard-operable buttons rather than the mockup's drag handle, because the
accessibility floor is a hard constraint and drag-and-drop is not in any AC.

**One place the mockup loses rather than is silent**, for the record alongside the three the epic
already carries and the Cancel/Apply one D-14.7.1 records as the fourth: it labels widths in `mm`.
`D-14.2.Q3` settled the display unit as points product-wide.

## Verification

**Commands** (D-000.33's per-story cadence — nothing heavier without an explicit override):

- `npx vitest run` -- expected: 0 failures. Baseline measured by me: **75 files / 1285 tests / 0
  failures, exit 0**. Test names must be diffed as a **multiset**, not by total — report GONE and NEW
  sets, not a delta.

  *Provenance:* the four baselines were measured at `fd3198d`/`cee83bf`, and `baseline_commit` is
  `c747023`. Every commit between them (`cee83bf`, `599ae62`, `c747023`) touches **only** files under
  `_bmad-output/` — verified with `git diff --name-only cee83bf..HEAD` — so no code moved and the
  baselines stand. Re-measure anyway before trusting them; five builders in this run have corrected
  anchors handed to them.
- `npx tsc -b --force` -- expected: exit 0, no output. `--force` is mandatory.
- `npx oxlint` -- expected: exit 0, **0 errors**, and the warning **set** below. Pre-story baseline:
  exactly 4 `react(only-export-components)`. **It is now 7, ruled deliberately — see below.**

  **The baseline is a SET, not the integer 4** (`D-14.7.2`, ruled at review):

  | file | count | why |
  |---|---|---|
  | `preview/pdf-viewer.tsx` | 2 | pre-existing |
  | `App.tsx` | 2 | pre-existing (line numbers move whenever `App.tsx` moves — they are not the pin) |
  | `segmented-control.tsx` | 3 | `alignGlyphs`, `alignSegments`, `justifySegment` |

  **Why 7 was accepted rather than restored to 4**, with the alternative's cost measured rather than
  estimated:
  1. The rule is a **Fast-Refresh/HMR developer-experience heuristic**. It changes nothing in the
     shipped bundle and flags no defect. `oxlint` exits **0**; these are warnings, and the error count
     — which is the correctness signal — is unchanged at 0.
  2. **Restoring 4 costs three modules where the extraction needed one.** The naive split is a
     **cycle**: `alignSegments` holds `<AlignIcon/>`, and `AlignIcon` reads `alignGlyphs`, so a single
     constants module would import the component while the component imports it back — the exact cycle
     the extraction existed to avoid. Avoiding it needs `align-glyphs.ts` → `segmented-control.tsx` →
     `align-segments.tsx`, three files for an HMR nicety.
  3. That split would **separate the ⚠ never-`justify` comment from the `justifySegment` constant it
     guards** — the one piece of prose in that module doing safety work, and the rule that keeps a
     column from being offered a value `ColumnAlignTokens` refuses.
  4. `App.tsx` and `pdf-viewer.tsx` already carry exactly this shape, so the new module is consistent
     with the codebase rather than newly deviant.
  5. **The integer was a proxy; the set is the predicate.** "Exactly 4" only ever meant "no unexplained
     export shape appeared". A per-file set says that directly, and a new warning anywhere shows up as
     a new key rather than as an integer a future builder might simply re-baseline.

  **A future story must re-measure the set and explain any new key.** Do not re-baseline silently.
- `npm run test:e2e:compile` -- expected: exit 0. **This is `tsc --noEmit` only. It is not a browser
  run and must never be reported as one.**

**What these prove, and what they do not — state this plainly in the report:**

- **Proven by execution:** the six columnheaders and the four absence fences; the align vocabulary
  fence; the footer reveal arms; the focus-restoration hazard; the arrow/Home/End reach; the budget's
  three states as computed values; the scope header's three states; the collection/alias read-back;
  the Epic 12 section's behaviour; the vocabulary contract's floors and census.
- **NOT proven, because jsdom applies no stylesheet and computes no layout:** that the six-track grid
  actually fits without horizontal scrolling; that the budget meter renders; and — most importantly —
  that the ALIGN control is not wearing `.matrix-row button`'s border, radius and mono font. Story
  14.3 shipped a padded hit region that passed every unit test and did not work in a browser at all.
  These live in `e2e/table-matrix-layout.spec.ts`, which is **compiled here and executed by CI on the
  push that closes this story**. `folio-designer-e2e` (`.github/workflows/ci.yml`) runs
  `npm run test:e2e` on every push with no `continue-on-error`; the reason a browser spec is normally
  uncovered mid-epic is that this pipeline pushes only at epic boundaries, not that the job is absent.
  **For this story the coordinator will push and wait for CI green before Story 14.8 is dispatched**,
  so the run is real rather than notional — but it is **still not covered until a run log shows
  counts**, and no report may call it covered before then.
- **Did not run, in these words:** the browser suite, the Go suites, the matrix legs, `npm run build`
  as a gate, the `verify:offline*` chain, and the font-host scans.

## Suggested Review Order

**The shared control — the AC that made this a refactor, not a restyle**

- The extracted control both surfaces render, so ALIGN is the same code, not a lookalike.
  [`segmented-control.tsx:60`](../../folio-designer/src/segmented-control.tsx#L60)

- Why a column never receives the four-segment array: `ColumnAlignTokens` has three members.
  [`segmented-control.tsx:15`](../../folio-designer/src/segmented-control.tsx#L15)

- The table's call site, which relabels each segment per column and cannot reach `justify`.
  [`TableEditor.tsx:296`](../../folio-designer/src/TableEditor.tsx#L296)

**The lattice — six labelled columns over twelve addressable cells**

- Cell addresses derive from `alignSegments.length`, so a fourth segment cannot collide.
  [`TableEditor.tsx:36`](../../folio-designer/src/TableEditor.tsx#L36)

- The grid itself: six columnheaders, reorder and remove as row affordances.
  [`TableEditor.tsx:370`](../../folio-designer/src/TableEditor.tsx#L370)

**Focus, which is where the real defects were**

- `focusCell` now treats an absent cell like a disabled one and prefers the same row.
  [`TableEditor.tsx:115`](../../folio-designer/src/TableEditor.tsx#L115)

- Every control addressed exactly once — the property, not the constant, is asserted.
  [`TableEditor.test.tsx:609`](../../folio-designer/src/TableEditor.test.tsx#L609)

**The read-outs, all derived browser-side — no projection field was added**

- The budget's four arms, including a table that starts outside its band.
  [`TableEditor.tsx:318`](../../folio-designer/src/TableEditor.tsx#L318)

- Available width is `band.width − table.x`, the engine's own `containComponent` rule.
  [`App.tsx:2297`](../../folio-designer/src/App.tsx#L2297)

- The sample item count, matched separately from counted so a sibling cannot answer.
  [`App.tsx:119`](../../folio-designer/src/App.tsx#L119)

- The read-only scope line the design draws in the sheet header.
  [`TableEditor.tsx:342`](../../folio-designer/src/TableEditor.tsx#L342)

**CSS — the hazard jsdom cannot see**

- Six fixed-or-`minmax` tracks: no content-sized track, so header and rows resolve alike.
  [`App.css:686`](../../folio-designer/src/App.css#L686)

- `:not(.property-segment)` scopes the exclusion instead of bumping the shared control.
  [`App.css:687`](../../folio-designer/src/App.css#L687)

**Peripherals — guards and browser specs**

- The vocabulary sweep's new state, inserted before `preview` so `slice(0, -1)` still drops `preview`.
  [`control-vocabulary-contract.test.tsx:348`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L348)

- Coverage names normalised, so a multi-column fixture cannot orphan a checked group.
  [`control-vocabulary-contract.test.tsx:605`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L605)

- The cascade proof: equality against the inspector's own Align group, not a magic height.
  [`table-matrix-layout.spec.ts:28`](../../folio-designer/e2e/table-matrix-layout.spec.ts#L28)

- Narrow viewport, the only place the overflow containment is reachable.
  [`table-matrix-layout.spec.ts:164`](../../folio-designer/e2e/table-matrix-layout.spec.ts#L164)
