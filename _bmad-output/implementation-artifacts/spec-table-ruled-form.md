---
title: 'A table prints as a ruled form'
type: 'feature'
created: '2026-09-13'
status: 'done'
baseline_commit: '6b181eb40b0a5aafefee99f776df981f8ec9be48'
review_loop_iteration: 1
context: ['{project-root}/_bmad-output/specs/spec-table-rules/format-changes.md']
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A statement form — a frame, a rule at every column boundary running the full height of a
fixed ruled area, a header fill with a rule under it, no rules between rows, bilingual two-line
headings — cannot be authored. A table's `style.border` is consumed as chrome on every cell, so the
frame *is* the grid, each interior line is stroked twice, and the frame's weight follows the row
count; there is no field for interior lines; a table has no height at all, so nothing can rule an
empty area; and a column label is shaped single-line, so a `\n` draws as a glyph no font covers.

**Approach:** Split the perimeter from the interior, **per page**. A table's
`style.border`/`style.background` paint the table's own frame, closed on every page the table
occupies. A new `table.rules` block draws the interior lines, addressed by *boundary*
(`between: ["columns","rows"]`), each drawn once and never on the table's own edge. A new
`table.minHeight` floors each page's slice, so column rules run through empty space on every page.
Column labels go through the packer a data cell already uses.

## Boundaries & Constraints

**Always:**
- Go owns all geometry; the designer never allocates it (AD-13, AD-15).
- Reuse the existing builders — the body cell's `packLines`, the element box's rect builder — never a
  second implementation of the same rule.
- A document that declares no table `border`/`background`, no `rules`, no `minHeight`, and whose
  labels hold no `\n`, must serialize byte-identically and render to the same PDF hash. The golden
  corpus is the witness.
- New keys round-trip (`decodeTableExt` `consumed` map + `writeElement`) and raise the document
  version.
- **The frame closes per page** (owner-ruled). Each page's slice of a table carries its own complete
  frame. **The repeated header lies inside it** (owner-ruled): on a continuation page the frame's top
  edge sits at the top of the repeated header, exactly as on the first page, and the column rules run
  through that header. On every page the frame's bottom edge sits at that slice's bottom.
- **`minHeight` floors every page's slice** (owner-ruled), not the table's total.
- **The perimeter belongs to the frame.** Where the frame strokes an edge, no header or cell edge is
  stroked on that same line.
- **Version: a minor bump** (owner-ruled). `SupportedVersion` becomes `3.1`; `SupportedMajor` stays `3`.
  Old documents load and render with the new frame semantics.
- **`between` is a closed set** `{columns, rows}` (owner-ruled). An unknown value is a load error.

**Ask First:**
- Any change to the inspector's BOX controls or to `ComponentBox` — investigation shows both are
  already correct for a table and must be left alone.

**Never:**
- No migration, no legacy arm, no version-gated second behaviour for the old `style.border` meaning.
  Owner-ruled: the two meanings cannot both be the default.
- No authored table `height`. AD-13's rule stands; `minHeight` is an input to a derivation, never a
  stored copy of a computed value.
- No per-cell border field, and no field to opt out of label wrapping.
- Do not make a rule spannable by naming a cell edge — `edges` vocabulary is not reintroduced.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Frame, not grid | table `style.border: {width:1}` | one stroked rect around the table's slice; no cell carries it | N/A |
| Ruled columns only | `rules: {between:["columns"]}`, 3 columns | 2 vertical lines, at the 2 interior boundaries; none on the table's left/right edge; no horizontal lines | N/A |
| Rules reach the floor | above + `minHeight: 300`, 2 rows, one page | the 2 verticals run from the slice top to the floored bottom, through the empty area | N/A |
| Line drawn once | `rules: {between:["columns","rows"]}` | each interior boundary carries exactly one stroke, not one per adjacent cell | N/A |
| Header boundary | `rules.between` includes `rows` and `headerStyle.border.edges` includes `bottom` | one line at that boundary — the header's border wins, the rules skip it | N/A |
| Perimeter owned by the frame | `style.border` + `headerStyle.border` with `top`, `left`, `right` | the table's outer edges are stroked once, by the frame | N/A |
| Floor below content | `minHeight: 10`, rows totalling 400 | slice is 400; `minHeight` never shrinks or clips | N/A |
| Floor too tall | `minHeight` exceeds a content window | **load error** naming the element | `TABLE_MIN_HEIGHT_UNPLACEABLE` |
| **Framed table across pages** | `style.border`, rows spanning 2 pages | page count **equal to the same table without a border**; each page's slice has its own closed frame, and on **every** page — continuation pages included — the header lies inside it (frame top = header top) | no `TABLE_ROW_CLIPPED_HEIGHT`, no near-empty first page |
| **Rules across pages** | `rules.between:["columns"]`, 2 pages | each page's column rules run from that slice's top — the top of its header, repeated or not — to that slice's bottom | N/A |
| **Floor on every page** | `minHeight: 300`, table ending partway down its last page | the last page's frame and column rules reach at least 300pt below that slice's top | N/A |
| **Floor capped by the page** | a slice whose top leaves less than `minHeight` above the content bottom | the ruled area runs to the content bottom; the table never moves page and never overflows | N/A |
| **Floor pushes content below** | a sibling element placed below a floored table on the same page | the sibling starts below the floored bottom | N/A |
| Two-line label | `"label": "วันที่\nDATE"` | two lines, centred per `align`; header row grows if `headerHeight` is smaller | no `TEXT_MISSING_GLYPH` |
| Over-wide label | label wider than its column | wraps; header grows | no silent clip |
| Mixed-script header | one Thai label and one English label in the same header | every column's lines use line metrics that fit every column's face — no overlap, no mis-spacing | N/A |
| Untouched document | none of the new keys, no `\n` | byte-identical serialization, identical PDF hash | N/A |

</frozen-after-approval>

## Code Map

**Go — format**
- `folio-go/internal/template/model.go:464` — `TableExt`; add `Rules Presence[TableRules]`, `MinHeight Presence[geom.Length]`. `Border` :537, `Style` :500, `Presence` in `presence.go:17`.
- `folio-go/internal/template/parse_bands.go:336` — `decodeTableExt`. New keys MUST join the `consumed` map or they round-trip opaquely as Extra. Precedents: `altRowBackground` :388, `headerHeight` :378 (`decodePointsRaw`), `headerStyle` :399 (nested block, null vs value), `decodeBorder` :788 (nested object shape to copy). Colours are validated at **render**, not load (`DiagCodeStyleColorInvalid`, `diagnostic.go:261`) — `rules.color` follows that precedent.
- `folio-go/internal/template/closedsets.go` — `between`'s closed set, declared **separately** from `BorderEdgeTokens` (a boundary is not an edge).
- `folio-go/internal/template/serialize.go:354` — `writeElement`, table branch :400-427. `writeBorder` :551, `writePoints` :107, `writeStringArray` :75.
- `folio-go/internal/template/version.go:262` — `versionRequiredByContent`; ranks ~:315, `versionForRank` ~:335 (guarded by `TestVersionForRankIsStrictlyAscending`); `SupportedVersion` ~:88.

**Go — pagination (the per-page slice; the heart of this change)**
- `folio-go/internal/layout/paginate.go:643` `Paginate`. Internally knows `pageOf` :777, `groupPage` :816, `headerPageOf` :833, but `PageAssignment` (:569-628) exposes only flat `ContentRects`, `HeaderRepeats`, `RowDisplacement`, `ClippedRects`. **Nothing today says "table T occupies page N from Y1 to Y2."**
- Space reservation precedent: the repeated header's `reservation` :792 and `ceilingFor` :1112 lower the window by `hh` per page. A per-(table, page) floor is reserved the same way, so a sibling below a floored slice is pushed rather than overlapped.
- Header repeat: normal path :1148-1202, clipped-row path :972-1029; `Shift` :1173. Over-tall group path :920-1093 (fresh page :935, clip at `contentBottom`, `TableRowClipped` :1075).
- `folio-go/render.go:3098` `paginateDocument` → `paginateWithFooterOrphanFix` (`table_footer.go:208-327`, re-runs Paginate, re-keys grouping, never touches output rects). The page loop **`render.go:3133-3231`** is the only code turning assignments into rects: shift/displacement/clip at :3194-3228. Rect refs flattened at :2963-2975 with parallel `rectElementID`/`rectIsDataRow`.
- **Frame and rules are emitted here, per slice**: insert immediately after table T's last shifted content rect and before `footer.Rects` (`render.go:3226`). Rects paint before text (`internal/pdf/textdoc.go:232-238`), so the frame lands over row and repeated-header fills and under all text.
- `folio-go/render.go:2469` `predictDocument` (items from `page_number.go:87` `contentColumnItems`) runs the same pagination for the page count. **It must agree with `paginateDocument`**, so any floor reservation applies to both.

**Go — table render**
- `folio-go/table_render.go:726` `collectBandTableRuns` — builds rows pre-pagination. Running extent: `tableBottom` :766, `rowHeight`/`rowBottom` :1112, `rowTop = rowBottom` :1254, footer :1346.
- `folio-go/table_render.go:672` `buildCellRectWithBackgroundField` / :598 `buildCellRect` / :464 `resolveBodyStyle` — must stop consuming the element's own `style.border`/`background`. Header keeps `resolveHeaderStyle` :357.
- `folio-go/table_render.go:842` — header label shaped `breaksAreDrawn`, never packed; one-line vertical model :880; glyphs pinned to line 0 :900; silent clip :908-921. Body packer to reuse: :1066 + `packLines` :1077; body clip diagnostic :1084.
- `folio-go/table_render.go:243` `chromeRowGroup` — row/header/footer grouping keys.
- `folio-go/element_box.go:52` `collectElementBoxRects` — **keep its `ElementTable` skip** (:57). It draws a *declared* rectangle and a table declares no height; the frame is built from measured slices instead.
- `folio-go/internal/pdf/rectdoc.go:57` — edge stroking.

**Go — designer boundary**
- `folio-go/page_setup.go:288` — `CanvasComponent` wire fields; `applyCanvasStyle` :2042, border projection :2106.
- `folio-go/table_columns_projection.go` — table projection to the designer.
- `folio-go/component_commands.go` — `setTableHeaderHeight` ~:2947 is the template for new setters; label bound is **256 bytes** :523.

**Designer**
- `folio-designer/src/engine-protocol.ts:335` `TableColumns` + **:617 `isTableColumns` uses `hasExactKeys`** — new projection keys must be registered or the worker terminates on `PROTOCOL_INVALID`. Column `header` bound :684 counts `.length` (UTF-16 units).
- `folio-designer/src/table-style-command.ts` — command builders; encoders from `command-json.ts`.
- `folio-designer/src/TableEditor.tsx` — `Props` :12; `styleNumber` :467, `styleColour` :438, `boxKey` :385; header height :717; sections `HEADER` :716 / `CELLS` :756 / `BORDERS` :773. Header label `<input>` :642.
- `folio-designer/src/App.tsx:3145` — the single `<TableEditor>` render; callbacks wrap `commitTableColumn` :1222.
- `folio-designer/src/App.tsx:5334` `TablePaint`, header cell :5365. The canvas must paint the engine's *packed* label lines, never let the browser wrap — `canvas-authority-contract.test.ts:46` bans browser text layout in `App.css`.
- **Read-only — already correct:** inspector `borderFields` `App.tsx:3473`, `BorderEdgesProperty` :5057, BOX render :3604, `ComponentBox` :5479.

**Evidence**
- **No fixture declares `style.border`** — no shipped golden moves from the semantic change alone.
- Fixtures with tables: `alignment-rounding`, `alternating-rows`, `line-spacing`, `statement-1/5/20/50`. `fixtures/statement-*` declare `headerHeight: 28`, whose single 8pt line with 8pt padding already measures 28.88pt.
- `folio-go/internal/template/fixtures_test.go:220` asserts a declared-key count on the maximal document.
- `folio-go/internal/text` `TestCorpusMeetsP6ExerciseFloors/P6g` fails at baseline (P6g=7, needs 20) — pre-existing, verified with every change stashed.
- The reverted iteration-0 implementation is preserved at the session scratchpad `impl-iter0/` (reverse-applies cleanly against the baseline).

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/internal/template/model.go` -- add `TableRules` and `Rules`/`MinHeight` to `TableExt` -- the format's new vocabulary
- [x] `folio-go/internal/template/closedsets.go` -- add the `between` closed set, separate from `BorderEdgeTokens` -- owner-ruled closed; a boundary is not an edge
- [x] `folio-go/internal/template/parse_bands.go` -- decode both in `decodeTableExt`, add to `consumed`, refuse a non-positive `minHeight`, a negative `rules.width`, and an unknown `between` value -- unconsumed keys round-trip opaquely
- [x] `folio-go/internal/template/serialize.go` -- write both in the table branch -- AD-9 round-trip
- [x] `folio-go/internal/template/version.go` -- rank the new keys; `SupportedVersion` → `3.1`, `SupportedMajor` stays `3` -- owner-ruled minor bump
- [x] `folio-go/table_render.go` -- stop body/footer cell chrome consuming the element's own `style.border`/`background`; emit **no** box or rule source pre-pagination -- a whole-table item is what broke multi-page pagination
- [x] `folio-go/table_render.go` -- pack labels through `breaksAreConsumed` + `packLines`; place glyphs at their line index; take line metrics as the **maximum across columns**, never the last column's; guard `contentW <= 0`; grow the header row only when a label packs to more than one line -- two-line headings, mixed scripts, the golden invariant
- [x] `folio-go/internal/layout/paginate.go` -- expose each table's per-page slice (element, page, top, bottom, continuation) on `PageAssignment`; reserve each slice's floor — `max(content bottom, min(slice top + minHeight, content bottom of the window))` — so siblings below are pushed -- the floor must be known where fit is decided, or it overlaps content
- [x] `folio-go/render.go` -- in the page loop, emit per slice the frame rect and the interior rules (verticals top→bottom; horizontals at row boundaries within the slice, skipping the header boundary when the header strokes its bottom), inserted after the table's last content rect; suppress header/cell edges that coincide with a stroked frame edge -- frame closes per page, perimeter owned by the frame
- [x] `folio-go/render.go` -- keep `predictDocument` and `paginateDocument` agreeing on page count with the floor reservation -- two pagination callers
- [x] `folio-go/folio.go` + new `table_min_height.go` -- refuse `minHeight` taller than the content window at `ParseTemplate`, naming the element; register `TABLE_MIN_HEIGHT_UNPLACEABLE` in `internal/diag` -- unsatisfiable floor
- [x] `folio-go/table_columns_projection.go` + `page_setup.go` -- project `rules`, `minHeight`, and each header label's **packed lines** -- the canvas paints the engine's answer
- [x] `folio-go/component_commands.go` -- add `setTableMinHeight` and `updateTableRules` (canonical `between` order, block collapse on the last clear, explicit `null` preserved); count the label bound in **Unicode code points** -- 256 bytes is ~85 Thai characters
- [x] `folio-designer/src/engine-protocol.ts` -- register the new projection keys in `TableColumns`/`isTableColumns`; count the column `header` bound in **code points** (`Array.from`), matching Go -- `hasExactKeys` terminates the worker; mismatched units refuse the whole reply
- [x] `folio-designer/src/table-style-command.ts` -- add the two command builders -- the editor's route to the engine
- [x] `folio-designer/src/TableEditor.tsx` -- author `rules` and `minHeight`; header label control accepts a line feed; disable rule width/colour until a boundary is ticked -- a rules block with no boundary draws nothing yet bumps the version
- [x] `folio-designer/src/App.tsx` + `App.css` -- wire callbacks; paint the projected packed label lines -- never browser wrapping
- [x] `folio-go` tests -- every I/O matrix row; multi-page framed and ruled tables (page count equal to the unframed table, one frame per page, no clip diagnostic); a one-line label taller than `headerHeight` leaves the row unchanged; both new commands (ordering, collapse, refusals, result still loads); projection **values**, not just key names; the code-point label bound -- the matrix is the contract
- [x] `folio-designer/src` tests -- exact command bytes for both new kinds; the ruled-area controls; the line-feed label control -- mock-call assertions alone let an encoder drift ship green
- [x] `_bmad-output/specs/spec-folio/folio-format.md` -- fold in the field changes, per-page frame and floor, `3.1`, and label semantics -- the canonical contract

**Acceptance Criteria:**
- Given a document that declares none of the new keys and no `\n` in a label, when it is loaded and serialized, then the bytes are identical and the rendered PDF hash is unchanged.
- Given any table with a frame or rules, when it renders, then its page count equals the same table rendered without them.
- Given a table whose label is too wide for its column, when it renders, then the label wraps and no diagnostic is silently withheld.
- Given the Table Editor and the engine, when a label of 256 code points containing astral characters is committed, then both accept it and the editor still opens.

## Spec Change Log

**2026-09-13 — implementation notes (three departures from the letter of the spec, each recorded
with its reason).**

1. **`headerHeight` grows the header row only when the packed label needs MORE THAN ONE LINE.**
   The context doc says the header row is `max(headerHeight, packed height + padding)`
   unconditionally. The frozen Boundaries block says a document with no `\n` in a label must render
   to the same PDF hash, with the golden corpus as the witness. The two conflict on the corpus that
   exists: `fixtures/statement-*` declare `headerHeight: 28` with 8pt labels and 8pt padding each
   side, whose ONE line already measures 28.88pt, so an unconditional `max` moved all four
   signed-off goldens by 0.88pt for a change that is about wrapping. The frozen constraint won; the
   narrowing costs the feature nothing, since the floor exists so a heading that NEEDS extra lines
   gets the room for them. Recorded in `table_render.go` at the site.

2. **`element_box.go`'s `ElementTable` skip was NOT removed.** The Code Map named it as "the gate to
   remove", but `collectElementBoxRects` draws a DECLARED rectangle (`declaredBox`: width and height
   both present and positive) and a table declares no height — AD-13 refuses one, and `minHeight` is
   a floor under a derivation, not a height. So a table's box height is only knowable where the rows
   have been measured. The box is emitted there instead (`table_render.go`'s `tableBoxAndRules`),
   through the SAME carrier and the SAME builder `collectElementBoxRects` uses — one box painter's
   worth of behaviour, at the one site that can compute the rectangle. Both files carry the reason.

3. **The table's box is a GROUPED pagination item (sentinel Index -3), not an ungrouped one.** A box
   spans the table's whole extent, so a bordered table taller than one content window is an item
   taller than the window — and an ungrouped one of those is D-2.6.1's hard refusal, which would
   fail the whole render for a document the author can otherwise print. Story 4.6 already settled
   what to do with an ENGINE-created group that cannot fit: clip it at the page's content bottom and
   say so (`TABLE_ROW_CLIPPED_HEIGHT`, with its own arm in `clippedRowDiagnostic` naming the box).

**2026-09-13 — review loop 1: intent_gap loopback, code reverted.**

- **Triggering finding (verified by reproduction, not taken on report):** a table declaring
  `style.border` whose rows span pages rendered **3 pages instead of 2**, with page 0 holding only the
  3 header cells, one frame clipped on page 1, and spurious `TABLE_HEADER_REPEAT_SUPPRESSED` and
  `TABLE_ROW_CLIPPED_HEIGHT`. Cause: entry 3 above. One grouped item as tall as the whole table
  forces pagination's over-tall-group path, which gives the group a fresh page and clips it. No test
  caught it because every table fixture carrying a border fit on one page.
- **Also triggering:** the implementer decided both Ask First gates — the major-version bump and the
  closed set — rather than halting.
- **Owner rulings (now frozen):** the frame closes per page; `minHeight` floors every page's slice;
  a minor bump to `3.1`; `between` is a closed set.
- **Amended:** frozen Intent, Always and Ask First; six new matrix rows (multi-page frame and rules,
  per-page floor, floor capped by the page, floor pushing content, mixed-script header, perimeter
  ownership); Code Map gains the pagination anchors; tasks re-derived.
- **Known-bad state avoided:** any frame, box or rule emitted as a single whole-table pagination item,
  grouped or ungrouped. Entry 3 is **superseded** — the frame is built per page slice after pagination.
- **Two decisions made in planning, not by the owner, flagged at the checkpoint:** a slice that cannot
  fit `minHeight` below its top runs to the content bottom (the table never changes page, never
  overflows); a sibling below a floored slice is pushed down, as rows already push it.
- **KEEP:** the format layer (`TableRules`, `decodeTableRules` with `consumed`, the closed set declared
  apart from `BorderEdgeTokens`, `writeTableRules`, the `3.1` rank, `maximalFixture` declaring both
  keys, the raised key count); entries 1 and 2 above; label packing via `breaksAreConsumed` +
  `packLines`; the `TABLE_MIN_HEIGHT_UNPLACEABLE` wiring; the `hasExactKeys` registration; the
  command builders; the Table Editor's ruled-area section and line-feed label control.
- **DO NOT KEEP:** `tableBoxAndRules` as a pre-pagination source; the sentinel `-3` group and its
  `clippedRowDiagnostic` arm; `headerVM` taken from the last column's metrics; the browser label guard
  counting UTF-16 units against Go counting runes; canvas label lines left to CSS.

**2026-09-13 — review loop 1, during step 3: where the repeated header sits in the frame.**

- **Finding (measured by reproduction, not taken on report):** on a continuation page the frame began
  *below* the repeated header, so that header sat outside the frame with no column rules through it.
  Page 1 measured `frameTop == headerTop` (inside); page 2 measured `frameTop == headerBottom`
  (outside). Every continuation page looked different from the first.
- **Cause: the frozen text, not the code.** The planner wrote "on a continuation page its top edge sits
  under the repeated header" and carried it into the option the owner approved. The implementer built
  it exactly as written and flagged the consequence as a risk.
- **Owner ruling (now frozen):** the repeated header lies inside the frame. On a continuation page the
  frame's top edge sits at the top of the repeated header, and the column rules run through it, so
  every page matches page 1.
- **Amended:** the frozen Always bullet on the per-page frame, and the two multi-page matrix rows.
- **Resolved in step 3 rather than left to review.** A frozen-block ambiguity surfaced in step 4 is an
  intent gap, and an intent gap reverts the whole implementation. This one was found during step 3's
  own verification, so it was put to the owner before review ran.
- **KEEP:** the whole loop-1 implementation. The change is confined to where a continuation slice's top
  is taken from, and to the tests that pin it.

## Design Notes

**The boundary vocabulary.** `edges` describes a cell's four sides, so it cannot say "rule between
columns but not between rows" without every cell owning an edge its neighbour also owns — why
interior lines are stroked twice today. `between` addresses the *boundary*, so a line has one owner:

```json
"style":  { "border": { "width": 1 } },
"rules":  { "width": 0.5, "between": ["columns"] },
"minHeight": 300
```

**Why the frame is built after pagination.** A frame's bottom is its slice's bottom, and a slice's
extent is only known once pages are assigned. Built before, the only item that could carry it is one
as tall as the whole table — which pagination correctly treats as a group that fits no page.

**The per-page floor**, for a slice with top `t` and content bottom `c` in a window ending at `w`:

```
floored bottom = max(c, min(t + minHeight, w))
```

It never lifts `c`, never passes `w`, and is reserved inside `Paginate` exactly as the repeated
header's height is, so whatever follows the table on that page starts below it.

**`headerHeight` is a floor only for multi-line labels** (Change Log entry 1): the golden invariant
forbids growing a row whose single line already overruns it, which the statement fixtures do.

## Verification

**Commands:**
- `cd folio-go && go test ./...` -- expected: all pass except the pre-existing `internal/text` `TestCorpusMeetsP6ExerciseFloors/P6g`
- `cd folio-go && go vet ./... && gofmt -l .` -- expected: clean, no files listed
- `cd folio-designer && npx tsc -b` -- expected: clean
- `cd folio-designer && npx vitest run` -- expected: all pass
- `cd folio-designer && npx playwright test` -- expected: 113 pass; a new failure is this change's
- `cd folio-designer && npx oxlint src` -- expected: 8 warnings, all pre-existing `only-export-components`

**Manual checks:**
- Place a table, set a 1pt border in the inspector BOX section, open Preview: the PDF shows a frame, not a grid.
- Bind the same table to enough rows to span two pages: two pages, each with its own closed frame, and no near-empty first page.

## Suggested Review Order

**Per-page slices: where frame, rules and floor are decided**

- Entry point: each table's slice per page, floored and capped; the design in one file.
  [`table_slice.go:203`](../../folio-go/internal/layout/table_slice.go#L203)

- Floor push moves only horizontally overlapping elements on the table's last page.
  [`table_slice.go:139`](../../folio-go/internal/layout/table_slice.go#L139)

- Paginate reserves the push where fit is decided, so floors never overlap content.
  [`paginate.go:1156`](../../folio-go/internal/layout/paginate.go#L1156)

- Page-count pass sends the same slice request, so page slots agree with the render.
  [`page_number.go:162`](../../folio-go/page_number.go#L162)

**Drawing the frame and interior rules**

- Frame and rules built once from style and rules; closed per page, not per cell.
  [`table_frame.go:50`](../../folio-go/table_frame.go#L50)

- Rects emitted after pagination per slice; perimeter edges suppressed under the frame.
  [`table_frame.go:166`](../../folio-go/table_frame.go#L166)

- Band tables get one floored slice, capped at the band's bottom.
  [`render.go:3015`](../../folio-go/render.go#L3015)

- Pushed elements move by their page's push: runs, images, rects.
  [`render.go:3224`](../../folio-go/render.go#L3224)

**Multi-line headers**

- Labels packed through the body cell's packer; row grows only past one line.
  [`table_render.go:1634`](../../folio-go/table_render.go#L1634)

- Canvas projection carries the engine's packed lines, never browser wrapping.
  [`page_setup.go:368`](../../folio-go/page_setup.go#L368)

**Format, load refusals and version**

- Decode `minHeight` and `rules`; closed boundary set, duplicates refused.
  [`parse_bands.go:444`](../../folio-go/internal/template/parse_bands.go#L444)

- Minor bump to 3.1; SupportedMajor stays 3.
  [`version.go:88`](../../folio-go/internal/template/version.go#L88)

- Load refuses a floor taller than the content window; commands refuse stranding one.
  [`table_min_height.go:36`](../../folio-go/table_min_height.go#L36)

- Round-trip writer for the rules block.
  [`serialize.go:592`](../../folio-go/internal/template/serialize.go#L592)

**Commands and designer**

- Min height command: located refusal above the window.
  [`component_commands.go:3020`](../../folio-go/component_commands.go#L3020)

- Rules command: canonical order; clearing `between` collapses the block.
  [`component_commands.go:3074`](../../folio-go/component_commands.go#L3074)

- Protocol guard mirrors Go's label-line bounds; unknown keys terminate the worker.
  [`engine-protocol.ts:94`](../../folio-designer/src/engine-protocol.ts#L94)

- Header label textarea: line feeds, rows follow lines, arrows respect the caret.
  [`TableEditor.tsx:721`](../../folio-designer/src/TableEditor.tsx#L721)

- Ruled-area controls; width and colour wait for a boundary.
  [`TableEditor.tsx:971`](../../folio-designer/src/TableEditor.tsx#L971)

- Canvas paints packed label lines.
  [`App.tsx:5365`](../../folio-designer/src/App.tsx#L5365)

**Tests**

- Framed table across pages: same page count, header inside every page's frame.
  [`table_ruled_form_test.go:113`](../../folio-go/table_ruled_form_test.go#L113)

- Window-changing commands refuse to strand a floor.
  [`table_ruled_form_test.go:592`](../../folio-go/table_ruled_form_test.go#L592)

- Push moves a sibling below, not one beside.
  [`table_ruled_form_test.go:727`](../../folio-go/table_ruled_form_test.go#L727)

- Each interior boundary stroked exactly once.
  [`table_rules_test.go:146`](../../folio-go/table_rules_test.go#L146)
