---
title: 'Make table columns bindable and addable in the editor'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
baseline_commit: '5ad83f626f55fb10dd17b875ebd1cf7a4bc5024e'
review_loop_iteration: 0
context: ['/Users/panitw/.codex/RTK.md']
---

<frozen-after-approval reason="user-requested bug fix">

## Intent

**Problem:** The table editor shows an empty, uneditable bound-field cell. Its Add column command also fails after the full-width starter column because the default new width overflows the band. Authors cannot finish configuring a table in the dialog.

**Approach:** Restore a row-field input with sample-derived suggestions inside each column. Let the engine fit added columns by splitting an existing column when the normal added width does not fit.

## Boundaries & Constraints

**Always:** Go owns canonical expressions, validation, dimensions, IDs, and history. Reuse the existing binding command and fenced editor commit path. Keep the main-window DATA binding workflow. Preserve arbitrary expressions without silently overwriting them. All controls are named and keyboard reachable; projected values must replace stale DOM values after edits or refusals. Add and any accompanying resize are one undo step.

**Ask First:** Persisted schema changes or unrelated product behavior are outside this request.

**Never:** Calculate authored widths or build binding expressions in browser code; change existing tables on load; change labels, bindings, or formatting as a side effect of Add; weaken unrelated contract checks.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Bind | Blank/simple field with or without sample | Enter relative field such as date or customer.name; Go applies current alias; suggestions only from current collection | Invalid path reports error and retains committed binding |
| Clear | Simple bound field emptied | Binding clears in one command; undo restores it | No-op clears add no history |
| Complex binding | Existing expression outside simple row path | Show expression and explain why simple field editing is unavailable | No accidental replacement |
| Add with room | At least 72pt available | Keep existing widths and add normal blank 72pt column | Existing limits remain |
| Add at width budget | Full-width starter or later crowded table | Split widest splittable column into two positive widths, keeping exact total and other columns unchanged; first widest wins ties | No splittable column, ID/cap/index/containment failure refuses atomically |
| Reprojection/history | Automatic resize, bind, rejection, undo/redo, Cancel | Controls, canvas, and saved bytes agree; Cancel restores this editor session | No stale values or partial history |

</frozen-after-approval>

## Code Map

- `folio-designer/src/TableEditor.tsx`: Props, CELL.bound=4, moveFocus/matrixCell, boxKey and binding-display cells. The old absence comments and help text intentionally forbid editing; this request supersedes that design. Reuse scoped candidates; keep the six visible columns and the engine binding output.
- `folio-designer/src/App.tsx`: commitTableColumn owns busy/revision/session fencing and Cancel edit counting; wire a binding callback through it. Existing updateTableColumnBindingCommand in table-column-command.ts escapes field input.
- `folio-go/component_commands.go`: addTableColumn adds 72000 then containment rejects; applyTableColumnCommand supplies a transactional clone. updateTableColumnBinding validates a relative path and constructs alias expression; extend empty field to clear. Use engine band remaining width and exact millipoints.
- `folio-designer/src/TableEditor.test.tsx`, `src/App.test.tsx`: mock projections, history/refusal checks, and pinned keyboard walks; replace obsolete absence assertions with positive authoring coverage. Search other consumers for new required callback.
- `folio-go/component_commands_test.go`, `folio-go/wasm/engine_test.go`: command admission, boundary, canonical persistence, history coverage.
- `folio-designer/e2e/table-editor.spec.ts`, `table-matrix-layout.spec.ts`: existing browser flows assume field absence/manual narrowing; update relevant traversal and Add assertions. Existing DATA coverage is in table-column-binding.spec.ts.
- Architecture AD-2/13/15/17: exact engine geometry, absolute column widths, single document owner. No schema or CSS geometry workaround.

## Tasks & Acceptance

**Execution:**
- [x] Restore field authoring, scoped suggestions, current binding display and accurate help in TableEditor.tsx; wire App.tsx callback.
- [x] Implement engine-owned Add fit and clear binding semantics in component_commands.go, with unit and WASM regressions.
- [x] Update frontend component/contract/traversal tests and affected existing browser specs without weakening unrelated assertions.
- [x] Coordinator: add a new real-browser regression in e2e/table-column-authoring.spec.ts and run production browser validation, including save/reopen and Cancel/history.
- [x] Review changes and run appropriate final checks.

**Acceptance Criteria:**
- Given the screenshot's fresh table, when Add column is clicked, then a second editable column appears without manual width changes and remains inside the band.
- Given the editor with a table bound to transactions[], when date is entered in its field input, then the binding persists as the current alias's date expression and can also still be changed from DATA.
- Given a configured table, when reopened after saving or undone/redone, then widths and bindings match the engine and remain editable.

## Spec Change Log

- Review added full-expression length checks, including alias migration, so accepted edits remain projectable.
- Real-browser testing exposed blur swallowing a following Add click; pending field edits now complete before Add or Done/Escape, while Cancel discards them.
- Review preserved caret shortcuts and focus after refusals, documented implicit footer-source constraints, and updated the native roundtrip test's obsolete missing-input assertion.
- Recorded the existing file-save/editor interlock gap in deferred-work.md; ordinary authoring, save/reopen, and session revocation are covered here.

## Design Notes

Use a native input/datalist matching Root collection; field labels are `Row field for column N`. Leave ordinary matrix navigation intact and provide a native suggestion keyboard route. Disable the simple input for complex expressions, with visible explanation. Empty simple fields clear through the same command. Width boxes must refresh when Add resizes a previously edited column. If no standard 72pt column fits, splitting the widest column makes the full-width starter become two equal columns without resizing unrelated columns. Assign any odd millipoint deterministically; accept the engine's existing minimum positive width.

A blank added column means unbound; retain the existing `Column N` header label. Clearing a binding that supplies the implicit source of a sum/avg footer keeps the engine’s transactional refusal until the author supplies an explicit `footerOf`; never alter the aggregate as a side effect.

The implementation subagent owns production, Go/unit tests, and updates to existing browser specs. The coordinator owns the new browser regression file and production browser runs. The implementation subagent does not commit or push; the coordinator records the reviewed local commit. Implementation details may be refined using local evidence within these constraints.

## Verification

- In folio-go: `rtk proxy go test -count=1 . ./wasm/...`.
- In folio-designer: focused Vitest suites for TableEditor, App, table binding, command encoding and ownership contracts; TypeScript and lint checks.
- Coordinator: `rtk npm run test:e2e:compile`, then `rtk npm run test:e2e -- e2e/table-column-authoring.spec.ts e2e/table-editor.spec.ts e2e/table-matrix-layout.spec.ts e2e/table-column-binding.spec.ts e2e/table-placement.spec.ts e2e/browser-native-roundtrip.spec.ts`. This performs production/WASM/offline build; avoid concurrent preview servers.
- `rtk git diff --check`; inspect a regression screenshot of the completed editor.

## Verification Results

- Final full Vitest run: 80 files, 1,573 tests passed. The implementation review pass also ran seven focused suites with 631 passing tests.
- Go root and WASM suites passed with cache disabled.
- Production browser run: all 16 tests passed, including three new authoring regressions, existing DATA binding, placement, and byte-identical browser/native PDF roundtrip.
- Browser coverage confirms three columns, exact width preservation, scoped suggestions, binding without sample data, direct field-edit-then-Add, save/reopen, DATA rebinding, undo/redo, invalid input refusal, Escape, and Cancel.
- TypeScript application build and E2E compilation passed. Lint has eight existing Fast Refresh warnings; no new errors. Diff whitespace check passed.
- Inspected the completed editor screenshot: three visible field controls, canonical bindings, readable controls, and exact width budget.
- Native datalist option content and the Alt+Down picker invocation are tested. Selecting the operating-system popup itself was not observable under automation, including an isolated native datalist; no claim of automated popup-selection coverage is made.

## Suggested Review Order

**Field authoring and action ordering**

- Expose editable row fields while preserving canonical binding output and complex expressions.
  [TableEditor.tsx:530](../../folio-designer/src/TableEditor.tsx#L530)

- Complete pending field edits before actions and retain existing session fences.
  [TableEditor.tsx:96](../../folio-designer/src/TableEditor.tsx#L96)

- Return admission success only after the matching editor projection arrives.
  [App.tsx:1164](../../folio-designer/src/App.tsx#L1164)

**Engine geometry and validation**

- Split the widest column atomically only when a normal addition cannot fit.
  [component_commands.go:345](../../folio-go/component_commands.go#L345)

- Validate relative paths and full expression lengths; clear bindings through the existing command.
  [component_commands.go:670](../../folio-go/component_commands.go#L670)

**Regression evidence**

- Exercise authoring, persistence, exact widths, history, and pending-action behavior in the production browser.
  [table-column-authoring.spec.ts:56](../../folio-designer/e2e/table-column-authoring.spec.ts#L56)
