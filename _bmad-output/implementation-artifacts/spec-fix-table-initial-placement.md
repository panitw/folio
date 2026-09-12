---
title: 'Start a newly placed table at full content width'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
baseline_commit: '987fdf1708b1d0fab23c16fd00f87213fc67f472'
review_loop_iteration: 0
context: ['/Users/panitw/.codex/RTK.md']
---

<frozen-after-approval reason="user-requested behavior">

## Intent

**Problem:** Dropping a table on the canvas creates a zero-width, 12-point-high box. It is difficult to see or use, and its horizontal position follows the pointer even though the user expects a table to span the document content.

**Approach:** Newly created tables start at band-relative x=0, retain the dropped vertical position under the existing snap setting, and span the content width. Give each new table one blank starter column of the full available width so that its real, saved geometry matches its canvas geometry.

## Boundaries & Constraints

**Always:** Go owns document geometry and creation defaults. Derive full width from the engine's band projection, which excludes page margins. Table width remains the sum of absolute column widths, with no independent table width field. Apply the same default to page-coordinate drops and band-relative creation used on later content sheets and keyboard placement. Preserve existing snap, rejection, transactional, save/load, and history semantics. Existing authored tables remain unchanged.

**Ask First:** Changes to persisted schema or automatic resizing of existing tables following page setup edits are outside this request.

**Never:** Add CSS-only minimum width, browser geometry calculations, synthetic data bindings or labels, or independently stored table width. Do not change unrelated component placement defaults or invent data rows.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Pointer drop | Table dropped away from left edge inside content, snap off | x=0; y equals drop offset within content; one blank column spans the entire band | Existing located errors remain available |
| Snapped drop | Table dropped near right edge with snap on | x=0; normal snapped y; exact band width even when not a grid multiple | Do not reject due to the pointer's horizontal location |
| Later-sheet creation | Table created with content-relative y beyond page one | x=0; retained content-relative y under snapping; full band width | Preserve continuous content coordinates |
| Width source | Nondefault page orientation/margins | Width matches current engine band width, not paper width or a hardcoded default | Normal page setup validation |
| Persistence/history | Newly placed table serialized/reloaded or undone/redone | Stable column ID and exact geometry; one creation history entry | Refusals do not mutate bytes or consume IDs |

</frozen-after-approval>

## Code Map

- `folio-go/component_commands.go`: `dropComponent` hit-tests page coordinates and snaps using provisional 72x24 geometry; `createComponent` accepts band-relative coordinates. Both reach `createComponentInBand`, which currently seeds no columns and forces width=0/height=12000. `projectedSize` sums columns. `addTableColumn` supplies established column ID/default conventions.
- `folio-go/page_setup.go`: `bandByName` and band projection supply the content width after margins. `containComponent` enforces component bounds.
- `folio-go/component_commands_test.go`: covers all five defaults, table command lifecycle, rejection transactions, binding, and column caps. Several fixtures intentionally construct empty tables via create; adapt those fixture setups explicitly so their original behavior remains tested.
- `folio-go/wasm/engine_test.go`: command history, snapshots and table projection coverage.
- `folio-designer/src/component-command.ts` and `src/App.tsx`: read-only integration references; first-sheet palette placements use dropComponent, later sheets use createComponent. The browser should continue sending intent.
- `folio-designer/e2e/table-editor.spec.ts` and `table-matrix-layout.spec.ts`: existing real-browser tests assume an empty newly dropped table and then add a column. Adapt to the starter column while preserving their assertions. Search other real creation consumers for the same assumption.
- Architecture spine AD-13/15/17: authoritative columns, engine-owned mutation, canvas paints projected geometry. Existing UX describes empty-table editing; that state remains possible by removing all columns or loading an existing empty table.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/component_commands.go` -- implement shared table defaults with exact band width and two safe allocated IDs; ensure drop snapping uses the actual table geometry.
- [x] `folio-go/component_commands_test.go` and relevant Go test fixtures -- cover matrix behaviors, ID allocation/rejection and canonical persistence; preserve coverage for deliberately empty authored tables.
- [x] `folio-go/wasm/engine_test.go` -- verify a table creation and its starter column undo/redo atomically.
- [x] `folio-designer/e2e/table-editor.spec.ts`, `table-matrix-layout.spec.ts` and a focused placement browser test -- adjust changed default assumptions and verify visible full-width placement with real WASM.
- [x] Review: `folio-go/component_commands.go`, `component_commands_test.go`, and `wasm/engine_test.go` -- duplicate every table column with a fresh ID and preserve canonical, independent-edit, rejection, and history behavior.
- [x] Review: `folio-go/component_commands.go` and `component_commands_test.go` -- apply the existing default font chain to new tables and verify populated-item and edited-header rendering.

**Acceptance Criteria:**
- Given a blank document, when a table is placed by pointer, then the visible selected table begins at the content area's left edge and spans its width at the intended vertical position.
- Given the table editor for a fresh table, when opened, then its single blank column reports the full available width and remains editable.
- Given existing non-table placement and existing authored table behavior, when their relevant regression checks run, then they remain valid.

## Spec Change Log

- 2026-09-12, review fixes: a table with a starter column exposed the existing Duplicate command's reused column IDs. Duplication now clones the column storage, allocates a fresh ID for every copied column, and preflights the entire allocation before mutation. New tables also adopt the existing default font chain, as new text does, so populated items and an edited header render successfully. The frozen placement intent, column labels/bindings, persisted schema, and existing authored tables are unchanged.

## Design Notes

One blank starter column is the minimum real table geometry satisfying this request without duplicating width authority. Do not introduce labels or infer data. Adjust old empty-table test fixtures through explicit removal of the starter column or loading authored empty tables, rather than weakening their contract assertions. No new dependencies are required. Implementer may refine file placement while preserving these behaviors.

The starter column consumes the full width budget. Adding a column retains the existing 72-point default and containment rules: the author first narrows the starter column to make room. This manual sizing tradeoff is intentional; Add does not redistribute or shrink existing columns.

Review regression coverage includes immediate duplication of a fresh table, duplication of multiple authored columns, independent source/copy edits, canonical reload, ID-exhaustion refusal without mutation, atomic duplicate undo/redo, and rendering populated items both before and after editing the starter header with the shipped font set.

## Verification

**Commands:**
- `rtk go test ./... -run 'Test(Component|Table|Drop|Create)'` from `folio-go` -- targeted command and table coverage passes; select additional specifically named regressions if necessary.
- `rtk go test ./wasm/...` from `folio-go` -- command/history tests pass.
- `rtk npm run test:e2e:compile` from `folio-designer` -- browser tests typecheck.
- `rtk npx playwright test e2e/table-placement.spec.ts e2e/table-editor.spec.ts e2e/table-matrix-layout.spec.ts` from `folio-designer` -- real production build and browser placement/editor tests pass.

**Review verification results (2026-09-12):**
- `rtk go test . -run 'Test(TableDuplicate|TableCreationUsesDefaultFont)' -count=1` -- all 7 focused native cases passed.
- `rtk go test ./wasm/... -run TestEngineFreshTableDuplicateHistory -count=1` -- the fresh-table duplicate/history regression passed.
- `rtk proxy go test -count=1 .` from `folio-go` -- the full native root-package suite passed in 11.357s, including all changed fixtures and existing authored-table/non-table regressions.
- `rtk go test ./wasm/... -count=1` -- all 39 WASM engine tests passed.
- `rtk npx playwright test e2e/table-placement.spec.ts e2e/table-editor.spec.ts e2e/table-matrix-layout.spec.ts e2e/component-properties.spec.ts` from `folio-designer` -- all 7 real-browser tests passed after the final production Go/WASM build (164.881s).
- Initial placement verification also passed the required targeted Go command, browser-spec typecheck and all 28 canvas-authority contract tests. No browser production code or schema was changed.
- The placement browser run saved snap-on and snap-off full-width screenshots under `folio-designer/test-results/`; the selected table visibly spans the content band. `git diff --check` is clean.

## Suggested Review Order

- Set full-width starter geometry and inherit the document font.
  [component_commands.go:1727](../../folio-go/component_commands.go#L1727)

- Use actual table geometry for drop snapping and containment.
  [component_commands.go:1654](../../folio-go/component_commands.go#L1654)

- Keep duplicated table columns independent with fresh document IDs.
  [component_commands.go:2054](../../folio-go/component_commands.go#L2054)

- Verify placement, exact widths, persistence, rendering, and safe duplication.
  [component_commands_test.go:3520](../../folio-go/component_commands_test.go#L3520)

- Verify creation and its starter column undo and redo together.
  [engine_test.go:135](../../folio-go/wasm/engine_test.go#L135)

- Observe real browser placement, dimensions, editor values, and history.
  [table-placement.spec.ts:3](../../folio-designer/e2e/table-placement.spec.ts#L3)
