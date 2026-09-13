---
title: 'Canvas keyboard shortcuts: copy, paste, delete, select all'
type: 'feature'
created: '2026-09-13'
status: 'done'
baseline_commit: 'd20bb16414076b732e20492f8f0fc27d4f30a812'
route: 'dispatch'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The design canvas has no copy/paste or select-all, and Delete only works when keyboard focus sits on a single selected element, so a drag-selected group cannot be deleted from the keyboard at all.

**Approach:** Extend the window `shortcut` handler in `App.tsx` with Copy (Cmd-C on Mac / Ctrl-C elsewhere), Paste (Cmd-V / Ctrl-V), Delete (Delete or Backspace) and Select All (Cmd-A / Ctrl-A), backed by two new engine commands — `deleteComponents` and `duplicateComponents` — so a multi-element delete or paste is ONE undo entry.

## Boundaries & Constraints

**Always:** Shortcuts fire only in Design mode, never while the target is editable (`isEditableTarget` / IME composing), never while the Table Editor is open, never while a pointer gesture or `fileBusy` owns the canvas. Mac/Windows modifier through `primaryModifier`. Copy never mutates the document. Paste and delete each produce exactly one history entry; paste selects the pasted copies; delete clears selection, column selection and the table editor claim as `deleteSelection` does today. A pasted copy follows `duplicateComponent` rules: same band as its source, new ids (table columns too), no keep-together tag, 6pt offset falling back to the source position when it would not fit, snapped when Snap is on.

**Never:** No cut (Cmd-X), no system/OS clipboard, no paste across browser tabs. No change to Cmd-D duplicate, arrow nudges or undo/redo. No TypeScript mirror of engine fit/offset rules — Go decides placement.

**Decisions (owner, 2026-09-13):**
- *Paste copies live originals.* The clipboard holds ids; paste duplicates those elements as they are now. Ids no longer in the document are dropped; if none remain, nothing is sent.
- *Repeated paste stair-steps.* After a successful paste the clipboard ids become the pasted copies, so successive pastes land +6, +12, +18pt.
- *Focus area = last band touched.* The band of the most recent canvas click, drag-select start, keyboard focus or selection change; Content before any band has been touched. Select All selects every component in that band.
- *Spec kept whole* at ~2,000 tokens: the four shortcuts share one handler, one guard and the new engine commands.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Delete group | 3 selected across bands, press Delete | all 3 removed, one undo restores all 3 | engine refusal → `commitError`, selection kept |
| Delete nothing | no selection, Delete/Backspace | nothing sent, default not prevented | N/A |
| Copy+paste group | 2 selected, Cmd-C, Cmd-V | 2 copies offset 6pt, copies selected, one undo removes both | N/A |
| Paste empty clipboard | nothing copied, Cmd-V | nothing sent | N/A |
| Copied id gone | copied ids later deleted/undone away | remaining ids pasted | missing ids dropped; none left → nothing sent |
| Repeat paste | Cmd-V ×3 after one Cmd-C | copies at +6, +12, +18pt | N/A |
| Select all | Cmd-A with focus area = header | exactly the header components selected | empty band → empty selection |
| Typing | focus in a property textbox | browser's native copy/paste/select-all/delete | N/A |

</frozen-after-approval>

## Code Map

- `folio-go/component_commands.go:242-317` -- `ApplyComponentCommand` switch; add `deleteComponents`, `duplicateComponents` cases.
- `folio-go/component_commands.go:2259-2348` -- `deleteComponent`/`duplicateComponent`; extract duplicate's clone body into a helper sharing one `ids` counter; do NOT change single-id behaviour.
- `folio-go/group_movement.go:86-235` -- pattern for `ids` parsing (non-empty, unique, all found, `componentFields` count) and working-copy-then-install atomicity.
- `folio-go/wasm/engine.go:298-354` -- `Engine.Apply` already gives one history entry per command; no change expected.
- Tests to mirror: `folio-go/component_commands_test.go:1199,1542,3875`, `folio-go/group_movement_test.go:71`, `folio-go/wasm/engine_test.go:745,1142`.
- `folio-designer/src/component-command.ts` -- add `deleteComponentsCommand(ids)`, `duplicateComponentsCommand(ids, snap)` via `commandBytes`/`jsonArray`.
- `folio-designer/src/App.tsx:1142-1158` -- `commitComponent` returns an id only when exactly one component is added; paste needs the full added-id list (return it or compute in `after`).
- `folio-designer/src/App.tsx:1159,1498` -- `installSelection` / `select`: route select-all and post-paste selection through `installSelection`.
- `folio-designer/src/App.tsx:1585` -- `deleteSelection` single-id only; generalize to any selection size.
- `folio-designer/src/App.tsx:2653-2706` -- window `shortcut` handler; add the four arms after the `editing || tableEditor` guard, using `selectedRef`/`modeRef`/`snapshotRef`.
- `folio-designer/src/App.tsx:1168,2870,5191` -- band `onPointerDown`/component press: where to record the last-touched focus-area band (also component focus and `installSelection`). Band names: `pageHeader`, `content`, `pageFooter`.
- `folio-designer/src/App.tsx:5191` -- component-level Delete/Backspace `onKeyDown` with `stopPropagation`; keep it calling the generalized `deleteSelection` so the window arm does not double-fire.
- `folio-designer/src/shortcuts.ts` -- add `copy`, `paste`, `selectAll` hints.
- `folio-designer/e2e/component-nudging.spec.ts` -- Playwright shape to mirror.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/component_commands.go` -- add `deleteComponents{ids}` and `duplicateComponents{ids,snap}`, atomic (all ids validated before any mutation, rollback on `Canvas` error) -- one command, one undo entry
- [x] `folio-go/component_commands_test.go`, `folio-go/wasm/engine_test.go` -- cover multi-band delete/duplicate, empty/duplicate/unknown ids refused with no mutation, table column ids fresh, keep-together cleared, single history entry
- [x] `folio-designer/src/component-command.ts` (+ test) -- the two command builders
- [x] `folio-designer/src/shortcuts.ts` -- hints
- [x] `folio-designer/src/App.tsx` -- clipboard ref, focus-area ref, generalized `deleteSelection`, copy/paste/select-all arms, clipboard reset when a new document generation installs
- [x] `folio-designer/src/App.test.tsx` -- guard cases: editable target, preview mode, table editor open, Mac vs Windows modifier
- [x] `folio-designer/e2e/canvas-keyboard-shortcuts.spec.ts` -- copy/paste group with undo, Delete group with undo, select-all per band

**Acceptance Criteria:**
- Given Mac, when Ctrl-C/Ctrl-V is pressed, then nothing is copied or pasted (and Cmd on Windows likewise does nothing).
- Given a pasted group, when Undo is pressed once, then every pasted copy is gone and the originals are unchanged.
- Given Cmd-A in Design mode on the canvas, then the browser's page text selection does not occur.

## Design Notes

The clipboard is a `useRef` of `{ generation, ids }`; paste ignores it when `documentGeneration` differs. Paste sends `duplicateComponents` with the ids still present in the snapshot, then selects the components the snapshot added.

## Verification

**Commands:**
- `cd folio-go && go test -count=1 ./ ./wasm/...` -- expected: green
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: green
- `cd folio-designer && npm run test:e2e -- canvas-keyboard-shortcuts component-nudging placed-component-selection selection-group-properties` -- expected: green

## Implementation Notes

- Group engine tests live in new files `folio-go/component_group_commands_test.go` and `folio-go/wasm/group_commands_test.go` rather than the files named in Tasks.
- The canvas-region `onKeyDown` Delete arm was removed; the window handler owns Delete for the canvas, the component-level handler remains.
- Step-03 audit added App tests for partial copied-id survival and a refused group delete keeping selection.

## Spec Change Log

## Review Triage Log

| # | Source | Finding | Verdict | Evidence | Route |
|---|--------|---------|---------|----------|-------|
| 1 | verification-gap, blind | Undo/redo empties the clipboard, so copy → undo → paste sends nothing | high | `applyHistory` (`App.tsx:2587`) calls `setCurrentSnapshot(…, true)`, which bumps `documentGeneration` and clears `clipboardRef`; contradicts matrix row "undone away → remaining ids pasted". Design Note keyed the clipboard on the wrong counter; code fix is a load-only identity ref | patch |
| 2 | blind | Paste → undo → paste does nothing because the clipboard now holds only the undone copies | medium | clipboard replaced by `added`; after undo none survive; user-visible stall | patch |
| 3 | blind, edge | Delete/Backspace/Cmd-C/V/A arms fire from any non-editable focus (toolbar, document bar, dialogs) and swallow page select-all / OS copy | medium | removed arm required `target === currentTarget`; new window arm has no target check; TableEditor/FontBrowser render outside `<main>` (`App.tsx:3192-3193`) | patch |
| 4 | blind, verification-gap | Component-level Delete bypasses the `canvasOwned` guard and now deletes groups mid-gesture | medium | `CanvasComponent` `onKeyDown` stops propagation and calls `deleteSelection` directly | patch |
| 5 | edge | Key auto-repeat or a second press before the commit resolves double-pastes or deletes gone ids | medium | `selectedRef` cleared only in `after`; clipboard updated only on resolve; `priorIds` captured per call | patch |
| 6 | edge | Shift+Delete/Backspace deletes | low | `plain` omits `shiftKey`; direct correction | patch |
| 7 | blind | `deleteComponents` pre-check loop is redundant | low | second loop fails before `installComponentCopy`; deletion only | patch |
| 8 | verification-gap | No tests for `canvasOwned` guard, toolbar Delete on a group, undo/doc-switch clipboard | medium | grep of tests finds none | patch |
| 9 | blind | Empty-band Select All clears errors | false | same `installSelection` path as clicking empty canvas; intended selection change | reject |
| 10 | blind | Focus band from last id of a cross-band rectangle | low | "most recent" selection change is later than drag start; rare; fix adds branches | reject |
| 11 | blind | New shortcut hints unused | low | spec task only adds them; no user harm | reject |
| 12 | blind | Stair-step stops when offset does not fit | false | spec Always rule mandates source-position fallback | reject |
| 13 | blind | Go helpers duplicate `moveComponents` internals / error path wording | low | refactor of untouched code; no divergence demonstrated | reject |
| 14 | edge | Paste resolving after a newer snapshot selects stale ids | low | pre-existing `commitComponent` pattern shared by create; rare; adds guard | reject |
| 15 | blind, edge | Spec names different test files / removed arm unmentioned | low | fix edits this build's spec; recorded in Implementation Notes | reject |
| 16 | blind | No UI test for cross-band group or Shift/Alt+V ignored | low | engine test covers cross-band; rare; more tests than direct fix | reject |
| 17 | post-patch verification | Delete pressed right after releasing a component drag/resize is dropped | high | `canvasKeyAllowed` blocks on `drag !== undefined`, which stays set until the commit settles; `e2e/component-manipulation.spec.ts:19` fails with the component still present at revision 4 | patch |
