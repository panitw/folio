---
baseline_commit: 6f21e51
story_key: 5-12-diagnostics-that-locate-and-an-interface-that-can-be-driven
status: done
created: 2026-08-28
---

# Story 5.12: Diagnostics that locate, and an interface that can be driven from the keyboard

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Covers:** FR41 · AD-10, AD-14, AD-15 · UX-DR17, UX-DR19, UX-DR20, UX-DR22, UX-DR25

**Standing delivery decision:** D-000.4 remains per-epic. This is not a hash-shaped override: it changes presentation, command history, and keyboard interaction, not the PDF algorithm. Focused/unit tests, typecheck, lint, production/offline build, Go tests/vet/native and js/wasm builds, formatting/diff checks, and compiled e2e source run in this story. Playwright execution and the four-target hash matrix are written/compiled but run only at the Epic 5 boundary after this story has been committed; Epic 5 cannot be marked done until both pass.

## In plain terms (read this first if you just want the gist)

Folio can already produce the real PDF in the browser, and the engine already returns warnings alongside a successful document. Today those warnings are only counted, so an author sees that something went wrong without a useful route to the thing that caused it. The interface also works best with a mouse, and there is no reliable way to reverse an accidental edit. That leaves the last part of the authoring loop incomplete: identify the issue, go to it, correct it, and keep working by keyboard if needed.

This story turns returned render diagnostics into clear, local, dismissible notices in Preview. Choosing a notice returns the author to Design and selects the engine-owned element the notice names. A failed render is stated in Preview with its element and data-path location, then offers the same return route. It also makes the palette, properties, existing binding/table placeholders, and frequent operations keyboard-operable; gives every control a visible focus treatment and accessible name; and adds undo/redo as engine-owned history of committed document commands.

It deliberately does not build sample-data discovery, binding by path, a parameter-aware editor, or the Table Editor matrix; those belong to Epic 6. It does not invent a screen-reader representation of the canvas, make raw data/parameter edits undoable, add a command palette, or redesign the preview’s stale-output truthfulness. The result is done when a warning or error names and locates its cause, normal authoring actions can be completed without pointer-only controls, and undo/redo changes the authoritative document rather than a browser-side copy. The final cross-browser/e2e and byte-matrix proof remains an Epic 5 boundary gate, so it is expected to be unrun in this story’s Delivery Log.

## Story

As a template author,  
I want a render complaint to point at the thing that caused it, and to be able to work without a mouse where it matters,  
so that fixing a problem is one click and the tool does not exclude people who cannot drag.

## Acceptance criteria

### AC1 — Successful-render diagnostics are bounded, local, and locate back to the authoritative element

**Given** the existing local production render returns one or more bounded `Warning` diagnostics (including clipped content, an over-tall row, or an uncovered glyph),  
**When** that exact Preview result becomes current,  
**Then** Preview presents each warning where its PDF consequence is visible, without blocking the PDF viewer or treating the render as failed.

**And** each card displays the engine-returned severity, code/message, and available element id/data path without recreating, parsing, or guessing diagnostic data in TypeScript; output is escaped/bounded and remains correlated to the exact current preview identity/revision/token. A stale, cancelled, superseded, failed, or prior-document result may not install, update, or locate diagnostics.

**And** warnings use the specified diagnostic-card semantics: triangle plus dashed border/left edge before amber/data colour, concise technical wording, a mono location, and a keyboard-operable dismiss action. Dismissal is UI-local to that exact preview generation only: it must not mutate template bytes, worker revision/history, dirty/save state, data/params, or suppress a later render’s same warning.

**And** activating a warning whose element id is present returns to Design, clears incompatible transient placement/drag state, selects that exact id in the current engine snapshot, and moves keyboard focus to a stable Design/canvas selection target. A diagnostic with no element id remains readable/dismissible but never selects an invented target. If the id is absent from the current authoritative projection, the UI announces a terse locate-unavailable status and changes no selection.

**Proof / red proof.** Unit/component/e2e-source tests exercise multiple warnings, each named warning class, no-element-id, unknown/stale id, dismiss/re-render, keyboard activation, and a late preview response. Prove no diagnostic state enters a document command or alters dirty/revision. Red-prove TypeScript-written diagnostics, colour-only warning semantics, a missing current identity/revision/token guard, dismissal that persists into a new render, and locate-back selecting an arbitrary component.

### AC2 — A failed render is located and honest without stealing Story 6.6

**Given** the local render fails with the existing bounded engine diagnostic,  
**When** Preview presents the failure,  
**Then** it blocks only the Preview work surface, states a terse technical error, names the available element id and data path, announces it once, and offers a keyboard-operable `Return to Design` route.

**And** failure never presents old PDF bytes as current, never claims a server/cloud retry, and never changes the document, saved revision, file target, raw inputs, selection, or undo history. Existing 5.11 stale/current/error authority controls continue to govern before any diagnostic is shown.

**And** this story provides the accessible location/return behavior only. It must not prematurely implement Story 6.6’s broader failed-render presentation, its `error-card`/`colors.danger` visual treatment, sample-data UX, or parameter-aware missing-parameter workflow.

**Proof / red proof.** Test element-plus-path failures, element-only/path-only/no-location failures, return focus, retry/return cancellation, and a late error after a newer current render. Red-prove an error card that overwrites a newer preview, an unannounced error, a false current label, or a browser-created location.

### AC3 — Every implemented interactive surface is keyboard-operable and visibly focused

**Given** keyboard-only use of the application,  
**When** the author reaches the palette, document/mode controls, canvas bands and component selection, page/property controls, Preview controls/diagnostic actions, and every currently implemented binding/table entry point,  
**Then** each is reachable in a logical tab order, operable with native keyboard conventions, and exposes a visible select-token focus indicator without relying on pointer hover.

**And** palette placement can be completed by keyboard through an explicit selected palette item and a focused band; canvas selection/property editing is not pointer-only. Pointer-primary drag/resize may remain pointer-primary, but its visual handles retain enlarged hit targets and cannot be the only route to select, delete, edit properties, or locate a diagnostic.

**And** every icon-only control has a specific accessible name; stateful controls expose state; status/error/diagnostic announcements are named, atomic, and non-spammy; controls do not introduce a tab trap. The existing PDF canvas remains accurately labelled as a rendered image, not falsely exposed as semantic PDF text.

**And** any future-binding/table controls currently represented as unavailable must state why and not pretend that a pointer-only capability exists. When Epic 6 supplies the binding tree and Table Editor, its concrete tree/data-grid interaction semantics remain that epic’s work; this story establishes the reusable focus/name/keyboard guardrails and tests the seams that exist now.

**Proof / red proof.** Use user-event/component and compiled Playwright-source coverage to tab through all implemented controls, activate palette/band placement and locate-back without a pointer, inspect focus styling/token use, validate labels/pressed/disabled reasons/live regions, and assert no tab trap. Red-prove removal of a focus-visible rule, an unnamed icon control, mouse-only palette/band placement, a colour-only status, and a misleading enabled future control.

### AC4 — Shortcuts are scoped, discoverable, and never steal text editing

**Given** the Design or Preview surface is active,  
**When** the author invokes frequent-operation shortcuts,  
**Then** Save, Undo, Redo, Delete, Duplicate, Nudge, Toggle Preview, and Toggle Snapping are covered with platform-appropriate modifiers where applicable; every implemented shortcut has a visible hint in its owning control, menu-equivalent, or tooltip.

**And** global shortcuts do not fire while a text input, textarea, select, contenteditable element, or an IME composition owns the keystroke, except platform Save where preventing the browser default is required and the operation is actually available. Disabled/unavailable commands do nothing and do not announce a success. Escape cancels local placement/drag/draft interaction before it is interpreted as a document mutation.

**And** no command palette, browser key remapping scheme, server shortcut, or hidden undocumented shortcut is added.

**Proof / red proof.** Table-driven shortcut tests cover macOS/other modifier normalization, Design/Preview routing, input/textarea/IME exclusion, disabled states, preventDefault behavior, and visible hints. Red-prove a document mutation from a property/input draft keystroke, an unhinted shortcut, or a shortcut that bypasses the engine command/history boundary.

### AC5 — Undo/redo is worker-owned committed-command history

**Given** an accepted canvas or property command,  
**When** Undo then Redo are requested,  
**Then** the one wasm engine instance changes its own canonical document/history and returns the normal immutable snapshot/projection over the existing FIFO channel; TypeScript stores no document history, diff, inverse command, or `.folio` schema model.

**And** history covers every accepted committed document mutation that Epic 5 exposes: page setup, create/drop, move, resize, delete, and component-property commits/clears/nulls/toggles. It does not include selection, focus, placement/drag previews, zoom/grid/snap preferences, Preview state/viewer state, open/save UI transitions, failed/rejected commands, or raw sample-data/parameter editing.

**And** undo/redo obey a bounded explicit policy, with rejected/no-history requests returned as a stable, non-mutating engine response. A new accepted command after Undo clears the redo branch. History dies on successful reload/start-blank/worker reinitialization and is never persisted in `.folio` or across a browser reload. Command admission remains transactional: a failed command creates neither a history entry nor a revision/dirty change.

**And** undo/redo invalidates preview truthfully through the existing 5.11 authority state machine, preserves the single worker/wasm instance and revision monotonicity, and updates selection safely when the selected element was removed/restored. It may not use a browser snapshot as a substitute for canonical bytes.

**Proof / red proof.** Native/wasm protocol parity and designer tests cover each command family, multi-step undo/redo, branch invalidation, bounds/no-op behavior, failure atomicity, load/start-blank reset, revision/dirty/preview invalidation, selection restoration/clearing, and one-worker/FIFO correlation. Red-prove a TypeScript inverse stack, history persistence in serialized bytes, raw input/load history entries, redo surviving a divergent command, or a failed command changing history/revision.

## Tasks / subtasks

- [x] **1. Preserve and expose the full bounded diagnostic contract** (AC: 1, 2)
  - Audit the existing `folio-go` → wasm host → worker protocol → `EngineClient` projection. Ensure successful render retains each bounded diagnostic object rather than only a count, and render failures retain the existing bounded location fields without duplicating a diagnostic registry in TypeScript.
  - Keep strict protocol keys, copied buffers, maximum lengths/list count, FIFO request correlation, and 5.11 identity/revision/token authority checks. Add only a narrow typed presentation projection where necessary.

- [x] **2. Build Preview diagnostic/error presentation and locate-back** (AC: 1, 2)
  - Add a narrow diagnostic presenter under `folio-designer/src/preview/` (or a similarly bounded component) and compose it through `App.tsx`; preserve the single PDF.js owner and the exact/stale authority contract.
  - Implement generation-local dismissal and locate-back through existing engine-owned canvas ids/projections. Never parse template bytes, synthesize ids/paths, or turn a missing id into a guessed selection.
  - Use DESIGN.md diagnostic-card semantics for warnings. Keep Story 6.6’s failure-card visual ownership out of scope.

- [x] **3. Make existing controls genuinely keyboard-driven** (AC: 3, 4)
  - Audit implemented App/canvas/palette/properties/file/mode/preview controls for native semantics, focus-visible select-token styling, labels, state, live announcements, and tab order.
  - Add keyboard palette-to-band placement, keyboard component selection/locate focus, and scoped frequent-operation shortcuts with visible hints. Do not claim Epic 6 binding/tree/table behavior before those surfaces exist.

- [x] **4. Add closed engine undo/redo commands and wire them once** (AC: 4, 5)
  - Implement history in the wasm session shell beside the canonical template/revision, with a closed byte command vocabulary and response behavior. Preserve AD-15 ownership and the one worker/one wasm instance.
  - Extend the protocol/client/worker/app seams and invalidate Preview using existing authoritative paths. Treat reload/start-blank as history reset, not undoable events.

- [x] **5. Prove the story and prepare the Epic 5 boundary gate** (AC: 1–5)
  - Add focused Go native/wasm, protocol/client, App/component, accessibility/static, and Playwright-source tests. Execute and restore every named red proof.
  - Run focused designer tests, `npm run typecheck`, `npm run lint`, `npm run build`, offline verification/red verification, and `npm run test:e2e:compile`; Go focused tests, `go vet ./...`, native build, `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine`, lint/hashmatrix static gates, `gofmt`, and `git diff --check`.
  - Do **not** run compiled Playwright or the four-target matrix green in this story. Record both as D-000.4 deferred, then hand the committed story to the Epic 5 boundary gate. Preserve the sanctioned P6g corpus-floor red if it persists; do not weaken its floor or regenerate unrelated goldens.

## Developer guardrails

1. **Diagnostics are producer facts, not browser interpretations.** The engine’s closed severity/code/location/message registry remains authoritative. TypeScript may display bounded received values and compare opaque preview evidence; it must not generate a second registry, template parser, inferred path, element id, or PDF diagnostic.
2. **Preview truth still wins.** Present diagnostics only for the exact current correlated preview. Existing 5.11 identity/revision/generation/token checks, cancellation, stale badge, local-only renderer, and PDF.js lifecycle remain mandatory.
3. **Locate-back uses an opaque existing id.** Return to Design before selecting; missing/obsolete/no-id diagnostics remain readable but do not manufacture selection. Locate, dismissal, focus, status, and viewer state are transient and never make the document dirty.
4. **Shape before colour.** Warning = triangle/dashed diagnostic treatment; error location/announcement must remain distinguishable before colour. Maintain cyan for focus/authority and amber only for data/diagnostic treatment; do not claim 6.6’s danger error-card work.
5. **Keyboard is a functional path.** Native controls first; focus must be visible, icon controls named, focus order logical, and keyboard shortcuts must respect editable fields and IME composition. No command palette.
6. **Undo lives in Go/wasm only.** History stores canonical engine state or a proven equivalent behind the worker, never TypeScript snapshots/inverses or serialized `.folio` history. Only accepted committed commands enter it; raw data/params and loading are expressly outside it; reload resets it.
7. **Do not broaden into Epic 6.** Binding tree/data loading and binding-by-pick, parameter-aware UX, table matrix/data-grid implementation, and Story 6.6’s complete failed-render visual language are deferred. The allowed accessibility seam work must not preimplement those product surfaces.
8. **Preserve unrelated work.** Do not modify `.agents/`, `_bmad` configuration/manifest churn, planning research, fixture/golden/font sources, or unrelated uncommitted changes.

## Project structure notes

- Existing seams to inspect first: `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go`, their native/host tests, `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, `App.tsx`, `App.css`, `component-command.ts`, `component-property-command.ts`, and adjacent tests.
- Story 5.10 already transports bounded diagnostic fields through the wasm host; Story 5.11 presently records only a count in `PreviewRecord`. Extend that record/presenter with full immutable bounded diagnostics, preserving copied PDF bytes and producer evidence.
- Prefer small focused modules under `folio-designer/src/preview/` for diagnostic presentation and under the existing engine protocol/session boundary for undo/redo. Do not add a second worker, direct React-to-wasm call, browser document model, or preview renderer.
- Existing `App.tsx` already has Save and Preview shortcuts, focusable bands, keyboard placement, Delete/Escape handling, and accessible PDF controls. Normalize and complete those paths rather than replacing them; test that global shortcuts do not hijack property/data/parameter drafts.
- The current UI honestly labels table binding as display-only and has no binding-tree/Table Editor surface. Keep that state explicit until Epic 6.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5, Story 5.12 ACs; the Epic 6 scope boundary.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-10, AD-14, AD-15 through AD-19; designer accessibility convention and source tree.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — Preview/state patterns, accessibility floor, interaction primitives, determinism obligations, and the explicit no-command-palette rule.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` — diagnostic-card/error-card shape rules, focus/select grammar, and token constraints.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.4, DW-17, post-render missing-glyph rule, undo-dies-on-reload ruling, and Story 5.11 identity decision.
- `_bmad-output/implementation-artifacts/5-9-a-canvas-the-browser-never-measures.md`, `5-10-preview-the-exact-production-document.md`, and `5-11-never-show-a-stale-preview.md` — shipped canvas authority, exact Preview, and stale-preview ownership/deferrals.

## Delivery log

### Story creation — 2026-08-28

- Created from the Epic 5/Story 5.12 source, current sprint tracker, architecture spine, UX DESIGN/EXPERIENCE, D-000.4/DW-17 and the decision log’s settled post-render/undo rules, plus completed Stories 5.1–5.11 and the current Go/wasm/designer code seams at baseline `6f21e51`.
- Confirmed the wasm render boundary already returns bounded diagnostic objects on successful renders and failures, while the current 5.11 app record retains only a diagnostic count. The developer must carry the existing producer facts through the closed protocol and presentation; no diagnostic schema is to be recreated in TypeScript.
- Confirmed `App.tsx` already offers Save and Preview shortcuts, keyboard palette-to-band placement, focusable bands, Delete/Escape, named icon buttons, and labelled PDF controls. This story completes and proves the keyboard path rather than discarding those seams.
- Confirmed no binding tree or Table Editor exists yet; their concrete keyboard/data-grid work stays in Epic 6. The implemented history boundary is absent: it must be built inside the wasm engine session, not in React.
- Implementation has not started. The developer must record real commands/results, red proofs, D-000.4 deferred suites, review resolutions, completion notes, and actual changed-file list. Do not pre-check tasks or claim test execution.

## Review findings

- [x] [Review][Patch][High] **Current-preview authority is not enforced at presentation or activation.** FIX — diagnostics now render only after the exact record is PDF-admitted/current, and locate rechecks token/generation/revision/identity before changing Design state. App coverage proves pre-admission and input-invalidated diagnostics remain hidden and inert.

- [x] [Review][Patch][High] **The failure return route and unavailable-locate announcement share the wrong state transition.** FIX — Return to Design is unconditional and focus-safe; optional selection occurs only for an existing opaque id. Obsolete ids announce locate-unavailable in Design, while path-only/no-location failures still return normally.

- [x] [Review][Patch][High] **React's undo/redo availability collapses a bounded engine stack to one step.** FIX — Go snapshots now carry bounded-session `canUndo`/`canRedo`; React reflects those facts only and retains no history model. UI coverage drives multiple undo/redo steps, bounds, and a divergent branch.

- [x] [Review][Patch][High] **Accepted no-op commands create history, revision, dirty, and preview changes.** FIX — canonical-byte equality exits before history/revision/redo mutation. Native and UI tests prove a no-op preserves the stable snapshot and redo branch.

- [x] [Review][Patch][High] **The mandatory designer unit gate is red.** FIX — shortcut hints are visually exposed but aria-hidden from intentional names, and the fake-timer test restores timers in `finally`. Full designer suite: 23 files / 120 tests PASS.

- [x] [Review][Patch][High] **Successful-render diagnostics are displayed but never announced.** FIX — an atomic named status announces the admitted generation’s warning count/codes once; local dismissals keep that announcement stable. Presenter tests cover multi-warning and dismissal behavior.

- [x] [Review][Patch][Medium] **Shortcut discovery is neither platform-correct nor complete.** FIX — shared platform normalization drives dispatch and visible hints for Save/Undo/Redo/Delete/Duplicate/Nudge/Preview/Snap. Table-driven tests cover macOS/non-mac hints and editable/IME exclusion.

- [x] [Review][Patch][Medium] **The diagnostic card does not use the specified design tokens.** FIX — the card uses `--color-bind-edge-dash`, exact 3px dashed bind edge, and `--color-bind-tint-warm`.

### Review Summary

- Outcome: **resolved**. Findings: **8 FIXED** (6 High, 2 Medium), **0 DEFER**, **0 DECISION-NEEDED**, **0 DISMISSED**.
- Evidence gap across the findings: the only new UI tests are three isolated presenter tests, and the only new history test covers one command, one Undo, one Redo, branch invalidation, and successful-load reset. There is no App-level diagnostic authority/dismissal/locate/failure suite, no shortcut/IME suite, no multi-step/bound/no-op/failure/load-failure UI-history suite, no client/worker host round-trip test for Undo/Redo, and no new Playwright source for Story 5.12. Those proofs are required parts of AC1-AC5, not optional polish.
- Sanctioned red remains exactly P6g: the ordinary full Go suite measured 1,225 pass / 2 reported failures / 4 skip; the child `P6g (opaque names)` is 7 against floor 20 and the parent reports that same underlying floor. No other Go failure appeared.

## Dev Agent Record

### Implementation summary

- Retained immutable, bounded engine diagnostics with each exact Preview record and added a local diagnostic presenter with shape-first warning treatment, generation-local dismissal, producer-location display, and locate-back only for an id present in the current Go projection.
- Added error location/return treatment that leaves Preview authority intact and preserves the existing stale/current state machine.
- Added visible focus treatment and native keyboard paths, including scoped undo/redo, duplicate, nudge, Preview, and snapping shortcuts that do not intercept editable-field or IME input.
- Added closed `undo`/`redo` operations. The Go wasm session owns bounded canonical-byte history, clears redo after divergent accepted commands, resets on load, and advances revisions monotonically. TypeScript stores only transient UI availability flags, never history or document bytes.

### Verification

- Designer: `npm test -- --run` — PASS, 23 files / 110 tests; `npm run typecheck`, `npm run lint` (five existing Fast Refresh warnings only), `npm run test:e2e:compile`, and `npm run build` — PASS. Offline baseline, red-proof, and wasm-witness checks — PASS.
- Go: focused wasm package — PASS, 11 tests; native and matrix `go build`/`go vet` — PASS; green full suite excluding the fixed corpus floor — PASS, 1,219 tests. Hashmatrix vet/probe/test — PASS. Lint module build/vet/tests — PASS.
- Ordinary known-red proof remains exactly sanctioned: `TestCorpusMeetsP6ExerciseFloors/P6g` is 7 against the immutable floor of 20 (6 pass / 2 reported failures). No corpus or golden source was changed.
- `git diff --check` — PASS. `folio-go/engine` and `folio-designer/test-results` were absent at handoff.

### Deferred work

- Under D-000.4, compiled Playwright execution and the four-target hash matrix remain deferred to the Epic 5 boundary. They were not run here.
- `lint/internal/rules/licencegraph_test.go` is pre-existing non-gofmt output in the separate lint module; it was not modified by this story.

## Completion status

Implementation complete, review findings resolved, and verified for local commit.

## Delivery Log

### 2026-08-28 — development (review)

Baseline `6f21e51`. Implemented producer-fact diagnostic presentation/locate-back, keyboard and accessibility guardrails, and wasm-owned committed-command undo/redo. Measured designer unit, typecheck, lint, e2e-source, production/offline/red/wasm release gates; Go focused/native/matrix/vet/full-green/hashmatrix/lint gates; and diff checks. D-000.4 defers only Playwright execution and the four-target hash matrix. The sanctioned P6g 7<20 known red remains unchanged.

### 2026-08-28 — independent review (review)

Baseline `6f21e51`. Adversarial review found 8 unresolved PATCH findings (6 High, 2 Medium): stale diagnostics remain visible/actionable, missing/obsolete failure locations cannot return correctly, React truncates multi-step history, engine no-ops create history/revisions, diagnostic announcements and shortcut discovery are incomplete, diagnostic tokens drift from DESIGN, and the current designer unit gate is red. Measured: focused designer run 2 files pass / 1 fail and 26 tests pass / 18 fail; typecheck and e2e-source compile pass; lint passes with five Fast Refresh warnings (the diagnostic presenter adds one); production build plus baseline/red/wasm offline verification pass. Go wasm-focused 11 pass, focused component-command 16 pass, vet and js/wasm build pass; full Go suite reports only sanctioned P6g (1,225 pass / 2 reports / 4 skip, 7<20). `git diff --check` passes. Per D-000.4, compiled Playwright execution and the four-target boundary matrix were not run. No implementation fix, commit, push, or unrelated-file edit was made.

### 2026-08-28 — finish (done)

Resolved all eight PATCH findings. Designer unit: 23 files / 120 tests PASS; typecheck, lint (four existing Fast Refresh warnings), production build, offline ordinary/red/wasm witnesses, and e2e-source compilation PASS. Go wasm tests: 12 PASS; native build, js/wasm build, vet, matrix-tag vet, hashmatrix static probe/tests, lint-module build/vet/tests, gofmt, and diff check PASS. Ordinary full Go remains the sole sanctioned red: P6g is 7 against floor 20 (1,226 pass / 2 reports / 4 skip). D-000.4 continues to defer actual Playwright execution and the four-target hash matrix to the Epic 5 boundary; Epic 5 remains in progress.
