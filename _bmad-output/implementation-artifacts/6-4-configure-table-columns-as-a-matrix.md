---
baseline_commit: 2f2c0b3
story_key: 6-4-configure-table-columns-as-a-matrix
status: done
created: 2026-08-28
---

# Story 6.4: Configure table columns as a matrix

**Epic:** 6 — A template author can bind a report to data and build the golden report  
**Covers:** FR10 · UX-DR8, UX-DR25 · AD-13, AD-15–17

## In plain terms (read this first if you just want the gist)

This story gives an author one focused Table Editor for the selected Table. It is a dense matrix: one row per ordered column, with cells for header label, fixed width, and cell alignment. The empty editor makes Add column unmistakable; once columns exist, the author can add, remove, reorder, and change those three values without leaving the surface. It is not five copies of a single-column form, a wizard, or a browser-side table model.

The saved table remains owned by Go/wasm. Every add, removal, move, header-label change, width change, or alignment change is one versioned command accepted, checked, canonically serialized, reparsed, installed, and recorded by the engine as one history step. TypeScript may keep local focus and draft text, but it paints only the admitted snapshot and sends opaque command bytes; it must never calculate a saved table width, allocate column IDs, rewrite columns optimistically, or construct `.folio` JSON.

Column width is authoritative. A table’s displayed and saved width is only the sum of its ordered column widths; no independent table-width field, scale-to-fit behavior, or browser measurement may appear. Header labels and cell alignment are the configuration in scope. Row collection binding, field binding in row scope, and footer aggregates remain deliberately absent until Story 6.5, even when sample data is loaded. The Table Editor is keyboard-operable as an accessible data grid, with visible cyan focus and named controls; amber remains reserved for data, which this story does not configure.

**Completion actual.** The matrix now admits every committed engine snapshot independently of selection, while its own projection is revoked by an editor-session token. Its dialog restores the invoking control on close, traps keyboard focus, and supports Escape; the grid reports the header plus each data row and retains the nearest logical cell after responses. Public Go command callers receive the same serialize/reparse transaction as wasm, and the closed projection cap is enforced before the 129th column can commit.

## Story

As a template author,  
I want to configure all my columns in one grid,  
so that setting up five columns is one task rather than five repetitions of a form.

## Acceptance criteria

### AC1 — A selected Table opens a focused matrix editor

**Given** exactly one Table component is selected,  
**When** the author invokes Table Editor,  
**Then** a focused surface opens over the canvas, preserving the selected table and leaving the canvas as the surrounding authoring context.

**And** only a selected Table can invoke it; no selection, multiple selection, or a non-Table selection has no misleading editor route.

**And** the editor paints a bounded, revision-correlated, Go-derived projection of the selected table’s ordered columns. TypeScript does not parse template bytes or keep a parallel template/table model.

**Proof / red proof.** Cover selected/non-table/multi-selection entry, editor close/return, document replacement, undo/redo, and delayed command/snapshot responses. Red-prove an editor opening for the wrong table, a response installing after its document/selection/revision is revoked, or a browser-created table document.

### AC2 — Empty and populated states are a true matrix, not repeated forms

**Given** the selected Table has no columns,  
**When** the editor is displayed,  
**Then** it states the empty table condition and Add column is the only meaningful configuration action and is unmistakable.

**Given** the Table has columns,  
**When** they are displayed,  
**Then** there is one matrix row per ordered column, with column-oriented controls for header label, width, cell alignment, reorder, and removal; it is never rendered as repeated single-column forms or a wizard.

**And** add, remove, and reorder preserve the engine-defined ordered column list and identifiers. A new column receives its canonical ID from Go/wasm. Removing or moving a column cannot leave a partial order, duplicate ID, or browser-only intermediate result.

**Proof / red proof.** Cover empty, one, and five-column states; add after a middle column; remove first/middle/last; reorder both directions; undo/redo; canonical save/reload; duplicate/malformed/out-of-range command rejection. Red-prove a repeated form, duplicate/unstable IDs, partial mutation, or an optimistic browser order surviving a rejected command.

### AC3 — Header and cell configuration are canonical transactional commands

**Given** an existing column,  
**When** the author commits a header label, fixed width, or cell alignment,  
**Then** Go/wasm validates the closed command vocabulary, target Table/column, bounds, IDs, allowed alignment values, and geometry; it canonically serializes and reparses the candidate before one atomic install and one undo-history event.

**And** header label and cell alignment are distinct persisted column attributes. Width is an absolute positive engine length; the canvas/editor presents only engine-derived values and never uses browser text/layout measurement to decide it.

**And** validation failure leaves canonical bytes, revision, history, ordered columns, and the currently admitted snapshot unchanged, while returning a bounded location/message suitable for the existing local command-error seam.

**Proof / red proof.** Test each command independently and as a sequence; successful canonical round trip; one revision/history step per accepted mutation; undo/redo; invalid table/column IDs, invalid alignment, zero/negative/overflowing width, over-wide table geometry, malformed envelope, and stale response. Red-prove direct template mutation, split header/cell updates, revision/history change on rejection, or a command that accepts an independently stored table width.

### AC4 — Geometry and scope boundaries remain explicit

**Given** columns are added, removed, moved, or resized,  
**When** the Table projection and canonical bytes are inspected,  
**Then** table width equals the ordered sum of column widths and is never stored or separately editable.

**And** this story does not bind a Table to a collection, bind a column to row-scope data, infer fields from a sample, configure footer aggregates, or alter renderer/footer schema. Loaded sample shape remains a discovery affordance only under D-6.2.1; it is not command legality or runtime type authority.

**Proof / red proof.** Prove width changes only through columns and survives save/reload, including reorder with an unchanged total. Red-prove a persisted table width, scale-to-fit behavior, a row collection/column binding/footer field in a Story 6.4 command or UI, or sample bytes/tree metadata entering canonical commands or history.

### AC5 — The matrix is operable as a data grid

**Given** the Table Editor is open,  
**When** the author uses keyboard navigation and controls,  
**Then** it behaves as a data grid: the active row/cell is perceivable, focus movement and editing are keyboard-operable, every add/remove/reorder/action control has an accessible name, and focus remains visible with the select (cyan) token.

**And** focus, selection, disabled/unavailable state, and errors are communicated by semantic/visible means, not colour alone. The dense matrix uses the existing zero-radius, dark-chrome/light-canvas, `matrix-row` token contract; amber is not used to imply structure or selection.

**Proof / red proof.** Add component and compiled e2e-source coverage for keyboard entry, row/cell movement, editing, add/remove/reorder, focus retention after a committed update and a rejected update, accessible names, and empty-state action. Red-prove a mouse-only control, missing/unnamed focus target, colour-only state, or focus loss that prevents a multi-cell keyboard flow.

## Tasks / subtasks

- [x] **1. Define the Go-owned closed table-column command vocabulary and transaction seam** (AC: 2–4)
  - Extend `ApplyComponentCommand` rather than adding a side channel. Define versioned opaque commands for add, remove, move/reorder, and one-field header/cell configuration; use canonical engine column IDs and deterministic bounds/order.
  - Apply every candidate through the existing fresh clone → canonical serialize → reparse → install/history path. Rejections must be non-mutating and located through the established command-error channel.
  - Reuse `template.TableExt.Columns`, `template.Column`, and `projectedSize`; never add a table `width` field or duplicate geometry rule.

- [x] **2. Extend only the paint-safe engine/worker projection needed by the editor** (AC: 1–4)
  - Add a bounded selected-table/column projection and strict Go/wasm response transport; keep the TypeScript protocol closed, revision-correlated, and deep-frozen.
  - Admit no template bytes, raw column schema model, sample tree, or independently calculated dimensions in the browser. Revoke late responses across command, undo/redo, open/start-blank, worker failure, and selection changes.

- [x] **3. Build the focused accessible matrix surface** (AC: 1, 2, 5)
  - Invoke it only from one selected Table; make the empty state and Add column action clear.
  - Render one `matrix-row` per engine-ordered column, never a repeated form. Use named add/remove/reorder and header/width/alignment controls, data-grid keyboard semantics, cyan focus, and the existing design tokens.
  - Keep browser drafts local until commit; send the smallest opaque command only. Do not introduce a TypeScript `.folio`/table model, browser measurement, or optimistic canonical state.

- [x] **4. Preserve document/Preview ownership and prove the contract** (AC: 1–5)
  - Accepted commands invalidate Preview through the existing canonical revision/freshness path; rejected commands do not. Preserve one worker, canonical serialization, undo/redo, local-only/offline operation, and existing sample/parameter ownership.
  - Add focused Go, wasm, protocol/client, App/component/accessibility, ownership/static, and compiled Playwright-source tests for all acceptance and red-proof clauses.
  - Run focused unit suites, `npm run test`, `npm run typecheck`, `npm run lint`, `npm run build`, `npm run test:e2e:compile`, relevant Go/wasm tests, `go vet ./...`, native and js/wasm builds, formatting/diff checks, and applicable lint/hashmatrix static gates. Record actual results and any genuine exception during implementation.

## Developer guardrails

1. **Go/wasm is the only document authority.** Table-column changes are canonical, transactional commands through the existing worker and engine history seam. No TypeScript document/table model, direct `.folio` mutation, browser inverse command, or optimistic saved state.
2. **Use ordered columns as the only width authority.** `table.width` must not be added, stored, sent, edited, or inferred. The table width is the checked sum of canonical ordered column widths; resizing/scaling to fit is prohibited.
3. **Do not pull Story 6.5 forward.** No collection binding, row alias/field picker, column `bind`, footer, `footerOf`, `footerFormat`, aggregate chooser, aggregate rendering, or sample-derived row legality. Existing schema fields remain untouched unless required solely to preserve their absence.
4. **D-6.2.1 still governs sample input.** The sample is transient discovery input, not table command authority. Its bytes, filename, tree metadata, observed kinds, and row shapes never enter canonical commands, document bytes, history, save state, or runtime compatibility checks.
5. **Keep accessibility functional.** The editor is an accessible data grid, not visually grid-like controls. Keyboard navigation/editing/reorder and visible cyan focus are required; named controls and shape/text cues supplement colour.
6. **Respect the visual grammar.** Use `DESIGN.md` tokens, zero-radius geometry, dark chrome/light canvas, `matrix-row` focus treatment, cyan for focus/structure/selection, and amber only for data. Do not hard-code colours.
7. **Preserve Preview truth.** Every accepted document mutation follows existing invalidation, generation, revision, FIFO, and stale-result admission rules. Canvas remains approximate; exact Preview stays serialized canonical bytes plus existing raw data/parameter inputs.
8. **Per-epic heavy evidence is deferred, not skipped.** Under D-000.4, real Playwright and the four-target hash matrix remain required Epic 6 boundary evidence. Do not run or claim them for this story unless implementation demonstrates a genuine integration/hash-shaped exception and the escalation is recorded before doing so.
9. **Preserve unrelated work.** Do not modify existing `_bmad` configuration/manifest changes, `.agents/`, planning research, fixtures/goldens/fonts, or unrelated user changes.

## Project structure notes

- `folio-go/component_commands.go` is the closed authoring command entry point. It already has transactional property mutation, `findComponent`, `projectedSize`, and containment checks; extend this seam rather than creating a table endpoint.
- `folio-go/internal/template/model.go` already owns ordered `TableExt.Columns`; each `Column` has canonical ID, label, absolute width, optional alignment, and later-only binding/footer fields. `template.AllocateElementID` and parser/serializer validation remain authoritative.
- `folio-go/component_commands.go:909` derives table paint width from columns. Keep that function/rule authoritative and test overflow/containment without any stored table width.
- `folio-go/wasm/engine.go`, `folio-designer/src/engine.worker.ts`, `engine-client.ts`, and `engine-protocol.ts` are the one-worker closed transport. Add only a bounded paint projection and operation-specific admission needed for the matrix.
- `folio-designer/src/App.tsx` owns selection, document generation, committed command dispatch, undo/redo, and Preview invalidation. Add the focused surface there or as a narrow component; drafts remain view state only.
- `folio-designer/src/App.tsx` currently truthfully says table size, binding, columns, and header style are not editable in the general property panel. Replace only the columns/configuration portion through the dedicated editor; table binding and header style remain out of scope.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 6 and Story 6.4 ACs; FR10; Stories 6.5–6.7 scope boundary.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-13 (derived table geometry), AD-14 (one diagnostic channel), AD-15 (engine-owned document), AD-16 (one worker), AD-17 (paint-only browser), and AD-21/22 determinism evidence.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR10, FR22–27, FR35–36, FR41–44, NFR1/NFR2/NFR5/NFR8.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — S4 Table Editor, `matrix-row`, empty table state, interaction model, and accessibility floor.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` and `mockups/TableEditor.dc.html` — matrix-row and focus token contract, dark-chrome/light-canvas and cyan/amber grammar.
- `_bmad-output/implementation-artifacts/epic-6-context.md` — Epic 6 scope cache, local/offline workflow, command authority, and dependencies.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — Epic 6 lead refresh, D-000.4 cadence, and D-6.2.1 sample-shaped affordance versus canonical/runtime authority.
- `_bmad-output/implementation-artifacts/deferred-work.md` — deferred boundary work and escalation/reporting obligations.
- `_bmad-output/implementation-artifacts/6-1-load-sample-data-and-browse-its-paths.md`, `6-2-bind-a-component-by-picking-a-path.md`, and `6-3-supply-parameters-for-preview.md` — completed transient sample, command transaction, snapshot, accessibility, Preview, and lifecycle patterns.
- `folio-go/component_commands.go`, `folio-go/internal/template/model.go`, `folio-go/internal/template/parse.go`, `folio-go/internal/template/serialize.go`, `folio-go/wasm/engine.go`, `folio-designer/src/App.tsx`, `folio-designer/src/engine-protocol.ts`, `folio-designer/src/engine-client.ts`, and `folio-designer/src/engine.worker.ts` — current implementation seams.

## Delivery log

### Story creation — 2026-08-28

- Created against baseline commit `2f2c0b3` after reading Epic 6 and Story 6.4; FR10 and table requirements; architecture AD-13–17; PRD and UX S4/matrix/accessibility sources; Epic 6 context; decision and deferred-work logs including D-6.2.1; completed Stories 6.1–6.3; sprint tracker; recent history; and current Go/wasm/designer table-command, schema, projection, worker, and App seams.
- No new owner decision is required. The existing ordered `TableExt.Columns` model, engine-only canonical mutation/history, paint-only browser snapshot, one-worker transport, AD-13 derived width rule, and D-6.2.1 separation of sample affordance from canonical/runtime authority settle this story's direction.
- Story 6.4 owns only structural columns: add/remove/reorder plus header label, absolute width, and cell alignment. Collection/row-scope column binding and footer aggregates are intentionally deferred to Story 6.5; no table width, browser template model, or renderer/schema expansion is warranted.
- Per D-000.4, real Playwright and the four-target hash matrix are deferred to the Epic 6 boundary unless a genuine integration/hash-shaped exception is recorded before implementation. Routine implementation must record actual focused test/build/lint results, red proofs, and any exception; it must not claim boundary evidence.
- Implementation has not started. No task is checked, no test result, implementation file list, completion claim, commit, remote operation, or decision-log amendment is recorded here.

### Implementation — 2026-08-28

- Added the closed Go/wasm command vocabulary for `addTableColumn`, `removeTableColumn`, `moveTableColumn`, and one-field `updateTableColumn`. New column IDs come from the engine allocator; each accepted command travels through the existing clone → canonical serialize/reparse → install/history transaction. Width stays the ordered column sum and containment is rechecked after structural/width mutations.
- Added a bounded, revision-correlated selected-table projection (`tableId`, ordered `id/header/width/align` only) across wasm, worker, protocol validation, and frozen client result. The browser never receives template bytes/schema, bindings, footer data, or an independent table width.
- Added a focused Table Editor matrix with empty state, accessible grid/row/cell semantics, named add/remove/move controls, keyboard cell movement, local input drafts, cyan focus styling, and command-only commits. Selection/document/revision checks revoke stale table projection and command results.
- Tests added: Go command/rejection/derived-width and wasm revision/history tests; protocol admission/red-field tests; App keyboard/accessibility test; compiled Playwright-source test `folio-designer/e2e/table-editor.spec.ts` (not executed per D-000.4).
- Verification passed: focused `go test ./folio-go ./folio-go/wasm`; designer Vitest (`149` tests); TypeScript typecheck; Oxlint (four pre-existing Fast Refresh warnings only); compiled Playwright source; offline red and wasm verification; native and js/wasm Go build/vet; diff check. `go test ./...` retains the pre-existing Epic 2 P6g corpus-floor failure (`7`, required `20`) and is not caused by this story. Per D-000.4, real Playwright and four-target hash matrix were intentionally not run.

### File list

- `folio-go/component_commands.go`
- `folio-go/table_columns_projection.go`
- `folio-go/component_commands_test.go`
- `folio-go/wasm/engine.go`
- `folio-go/wasm/engine_test.go`
- `folio-go/wasm/cmd/engine/main.go`
- `folio-go/wasm/cmd/engine/main_test.go`
- `folio-designer/src/table-column-command.ts`
- `folio-designer/src/table-column-command.test.ts`
- `folio-designer/src/TableEditor.tsx`
- `folio-designer/src/App.tsx`
- `folio-designer/src/App.css`
- `folio-designer/src/engine-protocol.ts`
- `folio-designer/src/engine-protocol.test.ts`
- `folio-designer/src/engine-client.ts`
- `folio-designer/src/engine-client.test.ts`
- `folio-designer/src/engine-ownership-contract.test.ts`
- `folio-designer/src/engine.worker.ts`
- `folio-designer/src/App.test.tsx`
- `folio-designer/e2e/table-editor.spec.ts`

### Review Findings

- [x] [Review][Patch][High] **Make the public Go table-column command seam transactional on every rejection.** [`folio-go/component_commands.go:79`](../../folio-go/component_commands.go#L79)
  - **Category:** AC2/AC3 · canonical transaction · rejection invariance
  - **Observation:** The new helpers mutate the caller's live `Template` before their final fallible checks. `addTableColumn` appends/inserts before containment, while `updateTableColumn` assigns the new label/width/alignment before containment; remove/move likewise mutate before the final `Canvas` call. The wasm `Engine.Apply` clone protects the worker path, but `ApplyComponentCommand` is public and the story explicitly extends that Go seam. A direct over-wide width command returned an error while `SerializeTemplate` proved the caller's canonical bytes had changed.
  - **Impact:** A rejected Go command can leave canonical state different from the revision/history/snapshot reported to its caller, contradicting the story's atomic no-mutation guarantee and making direct callers behave differently from wasm.
  - **Required patch:** Apply the whole table vocabulary to a fresh clone or complete every fallible validation/projection step before mutating the caller, then add over-wide/add-containment and post-mutation projection-failure witnesses that compare canonical bytes before/after rejection.

- [x] [Review][Patch][High] **Reject the 129th column before committing an editor-unprojectable document.** [`folio-go/component_commands.go:98`](../../folio-go/component_commands.go#L98)
  - **Category:** AC1/AC2 · projection bounds · history truth
  - **Observation:** `TableColumns` and the TypeScript protocol cap a selected table at 128 columns, but `addTableColumn` has no matching command bound. A measured engine sequence kept each existing column at `0.001` pt, accepted column 129, advanced to revision 259, and then failed `Engine.TableColumns` with `table has too many columns for editor projection`.
  - **Impact:** A legal editor command can commit/history-record a state the focused editor can no longer paint or repair. Undo is the only recovery even though the add response itself was reported successful.
  - **Required patch:** Enforce the same closed column-count ceiling in Go before ID allocation/mutation and prove the exact 128/129 boundary, unchanged canonical bytes/revision/history on rejection, and recovery after undo/redo.

- [x] [Review][Patch][High] **Install a worker-committed table snapshot independently of transient selection.** [`folio-designer/src/App.tsx:304`](../../folio-designer/src/App.tsx#L304)
  - **Category:** AC1/AC3 · async authority · selection/table switching · dirty/history truth
  - **Observation:** After the worker has committed a table command, `commitTableColumn` returns before `setCurrentSnapshot` whenever selection changed or cleared in flight. Selection is transient UI state and cannot revoke an already-installed engine mutation. Story 6.2 fixed this same bug class for binding, but the new table path reintroduces it.
  - **Impact:** The worker/save path can hold newer canonical bytes while canvas width/order, revision, dirty state, undo/redo availability, and selected-component properties remain stale. A user can save a column change the UI never shows.
  - **Required patch:** Admit the committed document snapshot by document generation and expected prior revision, preserving whatever selection is now active; independently scope only the editor projection/error to the still-current selected table. Add reselection, deselection, and switch-to-another-table races.

- [x] [Review][Patch][Medium] **Give the Table Editor its own session token so Close revokes delayed re-projection.** [`folio-designer/src/App.tsx:296`](../../folio-designer/src/App.tsx#L296)
  - **Category:** AC1 · close/return · stale responses · local editor scope
  - **Observation:** Close only clears `tableEditor`. A command already in flight retains the old projection in its closure; after commit it issues `table-columns`, and line 309 can call `setTableEditor` again because document, selection, ID, and revision still match. No editor-open epoch/current-session check exists.
  - **Impact:** Pressing Close during a commit is not final: the delayed query can reopen the modal and reinstall a surface the author explicitly revoked.
  - **Required patch:** Token open/close/table-switch sessions and require the exact live session before installing projection or error state; committed document admission must remain independent as described above. Prove close during command and close during re-query.

- [x] [Review][Patch][High] **Implement the matrix as a truthful data grid with retained focus.** [`folio-designer/src/TableEditor.tsx:11`](../../folio-designer/src/TableEditor.tsx#L11)
  - **Category:** AC5 · ARIA grid semantics · keyboard navigation · focus retention
  - **Observation:** The grid renders a header plus N data rows but declares `aria-rowcount=N`; data rows start at `aria-rowindex=1`, colliding with the header's position. `active` changes only through this component's arrow handler, so Tab/pointer focus can leave row 1 marked selected while another row is active. The later-move button has neither a matrix-cell target nor grid-key handling. Busy commits disable every focused cell, and removal deletes the focused row, yet there is no restore/next-cell policy after accepted or rejected updates, add/remove, or reorder.
  - **Impact:** Assistive technology receives inconsistent coordinates/selection, arrow navigation does not cover every action, and ordinary multi-cell keyboard flows can lose their only focus target—directly violating AC5's data-grid and focus-retention clauses.
  - **Required patch:** Define correct header/data row indices and roving cell semantics, update active row on every focus path, include both reorder actions in the keyboard model, and deterministically restore the corresponding/nearest cell after success and rejection (including last-column removal). Prove Tab, arrows, Home/End, both reorder controls, accepted/rejected edits, and structural mutations with real focus assertions.

- [x] [Review][Patch][High] **Complete the focused dialog's keyboard lifecycle and return focus to its invoker.** [`folio-designer/src/TableEditor.tsx:8`](../../folio-designer/src/TableEditor.tsx#L8)
  - **Category:** AC1/AC5 · modal focus · keyboard-only operation
  - **Observation:** The component labels itself `aria-modal=true` but neither traps focus nor makes the canvas/chrome inert, handles Escape, or records/restores the `Configure columns` invoker. Its `opener` ref actually points to the Close button; closing removes the focused node and leaves no deterministic return target.
  - **Impact:** Keyboard users can tab into supposedly inert background controls, change application modes beneath the editor, and lose their place when they close it. The claimed focused surface is not behaviorally modal.
  - **Required patch:** Implement modal containment/Escape and focus restoration to the invoking table control (or the nearest surviving selected-table control), then prove open, wraparound, close, removed-table/document-replacement fallback, and return-to-canvas behavior.

- [x] [Review][Patch][Medium] **Use complete JSON encoding for table command strings.** [`folio-designer/src/table-column-command.ts:3`](../../folio-designer/src/table-column-command.ts#L3)
  - **Category:** AC3 · command envelope · local drafts
  - **Observation:** The hand-written `quote` escapes only backslash, quote, newline, carriage return, and tab. Valid JavaScript strings containing other JSON control characters (for example `U+0000`, backspace, or form feed) produce malformed command bytes. A measured `A\u0000B` header failed `JSON.parse` with a bad-control-character error. Story 6.2 already replaced this exact quoting pattern with complete JSON encoding.
  - **Impact:** A locally accepted header draft can fail as a generic malformed envelope rather than reach Go's bounded header validation, and command factories have inconsistent string semantics.
  - **Required patch:** Encode all string values with the platform JSON encoder while preserving the fixed closed field order/shape; add control-character, slash/quote, Unicode, and malformed-number factory tests against Go decoding.

- [x] [Review][Patch][Medium] **Reject stray `tableColumns` metadata on render and identity responses.** [`folio-designer/src/engine-client.ts:97`](../../folio-designer/src/engine-client.ts#L97)
  - **Category:** AC1/AC4 · strict operation correlation · protocol surplus fields
  - **Observation:** Operation mismatch checks exclude `parameterReferences` from render/identity, but the new `tableColumns` field was not added to those two branches. A shape-valid render or identity response carrying unrelated table metadata passes and is returned to a caller that did not request it.
  - **Impact:** The closed response channel is again operation-surplus, recreating the Story 6.3 protocol defect and leaving an accidental-consumption path for selected-table state during Preview work.
  - **Required patch:** Use one exhaustive operation-to-success-payload table and reject table metadata everywhere except `table-columns`; add client-level red tests for render, identity, serialize, parameter references, snapshot, command, undo, and redo.

- [x] [Review][Patch][Medium] **Make the wasm `table-columns` request envelope exact, not merely ID-bearing.** [`folio-go/wasm/cmd/engine/main.go:109`](../../folio-go/wasm/cmd/engine/main.go#L109)
  - **Category:** AC1/AC4 · closed transport · later-scope exclusion
  - **Observation:** Unlike `parameter-references`, the new dispatch arm does not reject `templateBase64`, `dataBase64`, or `paramsBase64`. Its payload is unmarshaled into `{id}` without rejecting unknown members, so `{id, bind, footer, sample...}` is silently accepted even though those fields are explicitly outside Story 6.4.
  - **Impact:** The semantic boundary claims a strict selected-table query while tolerating surplus byte channels and later-story/data fields, weakening the no-row/footer/sample authority fence.
  - **Required patch:** Require exactly one bounded `id` member and no other byte inputs, with native dispatch witnesses for every surplus channel/member and malformed/oversized IDs.

- [x] [Review][Patch][High] **Replace the Delivery Log's broad proof claims with executable Story 6.4 witnesses.** [`folio-designer/src/App.test.tsx:34`](../../folio-designer/src/App.test.tsx#L34)
  - **Category:** AC1-AC5 · verification gap · delivery-record accuracy
  - **Observation:** New evidence consists of two direct Go tests, one wasm test, one parser-shape test, one App happy-path arrow assertion, and a 19-line compiled-only Playwright happy path. It does not execute empty/one/five columns; add-middle; remove first/middle/last; both move directions; every command's no-op/redo/history behavior; canonical save/reload; malformed/duplicate/non-table/overflow/over-wide/cap/surplus rejection; selection/document/close races; command encoding; derived-width persistence; row/footer/sample absence; or focus retention after accepted/rejected/structural updates. The Delivery Log nevertheless claims focused command/rejection, protocol/client, App/accessibility, ownership/static, and compiled-source coverage for all acceptance/red-proof clauses.
  - **Impact:** All nine concrete defects above survive 149 green designer tests and the focused Go suites, so the recorded evidence is materially stronger than what ran.
  - **Required patch:** Add focused Go/wasm, dispatch/client, command-factory, App/component/accessibility, authority/static, and compiled e2e-source witnesses for every named proof/red-proof clause. Keep real Playwright and the four-target matrix explicitly deferred under D-000.4 and do not claim either result.

### Review Summary

- **Outcome:** **Resolved.** All **10** patch findings are closed: **6 High**, **4 Medium**; 0 decision-needed, 0 deferred, and 0 dismissed.
- **Story status:** **`done`**; `sprint-status.yaml` is synchronized to `done` while Epic 6 remains `in-progress`.
- **Authority/scope audit:** The production projection itself is bounded to `tableId` plus ordered `id/header/width/align`, and the new UI/commands contain no row binding, footer aggregate, sample evidence, template parser, saved table-width field, browser measurement, or second TypeScript `.folio` model. The defects are transactional, bound/correlation, async admission, command encoding, accessibility, and proof gaps—not Story 6.5 scope leakage in canonical state.
- **Review execution:** The commissioning handoff prohibited subagents, so the adversarial, edge-case, verification-gap, and acceptance lenses were applied directly in this fresh isolated reviewer. Unrelated `_bmad` configuration/manifest changes, `.agents/`, and planning research remained outside the review/edit scope.
- **Measured designer gates:** Vitest passed **154 tests across 27 files**; typecheck, E2E-source compile, production build plus ordinary offline verification, offline wasm witness, offline red proof, and `git diff --check` passed. Oxlint emitted only the four established Fast Refresh warnings in `src/preview/pdf-viewer.tsx` and `src/App.tsx`.
- **Measured Go/static gates:** Full unskipped Go reproduced only the established P6g corpus-floor red: **1,240 passed, 2 failed, 4 skipped**. Focused table/wasm tests, `go vet ./...`, native build, js/wasm host test/build, and gofmt check passed. `lint` passed **117 tests across 4 packages** plus vet/build; `hashmatrix` passed **3 tests across 2 packages** plus vet/build.
- **Resolution witnesses:** Direct public-command rollback and the exact 128/129 cap are covered in Go; native wasm dispatch rejects every extra byte channel and envelope field; client tests reject `tableColumns` on every non-table operation; command-factory tests cover JSON controls/quotes/Unicode/non-finite values; App tests cover focus trap/Escape/return and commit-after-close; compiled e2e source covers the seven-cell grid, named actions, keyboard movement, and Escape.
- **Explicitly unrun under D-000.4:** Real Playwright execution and the real four-target hash matrix. Their absence is intentional at story cadence; the compiled Playwright source was inspected and neither result is claimed as green.
- **Cleanup:** Temporary repro sources/binaries and the generated `folio-go/engine` build artifact were removed. No fixture/golden/font edit, decision-log change, push, or remote operation was performed.

### Finisher implementation — 2026-08-28

- Wrapped public table commands in a clone → serialize → reparse → canvas → install transaction and added the shared 128-column cap before ID allocation.
- Added independent document-snapshot admission and session-scoped editor projection/error admission, including Close, selection, document replacement, and Escape revocation paths.
- Corrected modal and matrix semantics: header-inclusive row count, unique data row indices, seven separately navigable columns, both move actions, active-cell updates on focus, retained focus, and invoker/canvas return.
- Replaced partial string escaping with `JSON.stringify`, made success payload admission exhaustive, and made the wasm table query accept only an exact `{id}` payload with no byte side channels.
- Real Playwright and the four-target hash matrix remain intentionally unrun under D-000.4; only compiled Playwright source is claimed.
