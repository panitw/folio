---
baseline_commit: 7bfb076
story_key: 6-2-bind-a-component-by-picking-a-path
status: done
created: 2026-08-28
---

# Story 6.2: Bind a component by picking a path

**Epic:** 6 — A template author can bind a report to data and build the golden report
**Covers:** FR7 · UX-DR7

## In plain terms (read this first if you just want the gist)

Template authors can now connect a selected text field to an offered path from local sample discovery with one committed Go/wasm command. The picker offers observed primitive root-scope leaves as a convenience; Go owns the command grammar and writes only the accepted binding expression. There is no binding-path text box and typed placeholder syntax is rejected through the ordinary property command. Once connected, the report field clearly shows that it receives data, while the ordinary selected outline and handles continue to mean only that the field is selected.

The connection is part of the saved report, so it behaves like every other real edit: it can be undone or redone, makes the report need saving, and is accepted or rejected by the same local engine that owns the report. The data sample itself remains a temporary companion; choosing from it does not save the sample file or its location. The tree can show a currently chosen item, but it cannot silently decide that a connection is valid. The engine makes that decision and returns the updated view only after the whole edit succeeds.

This is deliberately the simple, report-wide case. It supports ordinary scalar fields used in the golden report. It does not yet connect a table to a collection, bind table columns, introduce row-relative choices, edit parameters, or change Preview’s existing input and freshness rules.

## Story

As a template author,
I want to bind a field by choosing it from the tree,
so that I never mistype a path and discover it at render time.

## Acceptance criteria

### AC1 — A selected component can receive a picked root scalar binding through one engine command

**Given** a loaded sample, one selected component, and an offered primitive root-scope path chosen from the docked Data tree,
**When** the author connects them,
**Then** the browser sends one opaque, committed binding command to the existing Go/wasm document owner, and only a successful response installs the next snapshot.

**And** the command atomically stores the engine-defined root-path binding in the canonical template, advances the document revision and engine-side undo history exactly once when bytes change, invalidates Preview through the existing authority path, and participates in save/undo/redo like another document mutation.

**And** Go/wasm owns command-envelope legality, target eligibility, root-path grammar, reserved namespaces, bounds, canonical mutation, and rejection diagnostics. Runtime value kind remains AD-14 validate/render work against the actual input; the browser may filter its offered primitive leaves from the active sample as discovery affordance, but must not become a TypeScript grammar, compatibility matrix, template model, or alternate validator. A rejected command leaves canonical bytes, revision, undo/redo branches, selection truth, and displayed binding state unchanged.

**Proof / red proof.** Prove one successful picked bind reaches the real worker command channel, serializes canonically, survives save/load, and is undoable/redone as one mutation. Prove a syntactically valid collection-shaped root path remains command-legal without sample evidence and that AD-14 reports its incompatible runtime value. Red-prove a browser-only state change, two commands for one connect action, a malformed/unsupported command that partially installs, a command response accepted after document replacement, or sample-tree observation becoming command legality.

### AC2 — Bound and selected states are distinct, paint-only, and accessible

**Given** a component whose binding command succeeded,
**When** the canvas and binding panel paint the current engine snapshot,
**Then** the snapshot exposes only the bounded paint state needed to identify that component’s binding; it is not a browser-owned `.folio` model or editable document mirror.

**And** the component and panel show the bound path as data using the amber bind treatment and machine-value typography. A bound component that is also selected keeps cyan selection border and handles; amber never replaces, recolours, or impersonates selection/focus. The indication is not colour-only: it has a concise accessible name/text equivalent.

**And** the Data tree remains keyboard-operable. Selecting a candidate and connecting it to the current component is keyboard reachable with named controls and visible cyan focus; unavailable connection states state the concrete reason. Discovery expansion, active-tree styling, and a pending choice remain transient UI interaction, not document state.

**Proof / red proof.** Cover unbound, bound, selected, and bound-plus-selected render states; keyboard tree-to-connect flow; named controls; focus visibility; and snapshot protocol rejection of a malformed/over-broad binding projection. Red-prove amber selection handles, colour-only binding, a stale worker response repainting a different document, or an editable TypeScript binding mirror.

### AC3 — Golden-report scalar fields are authored by picking, with later binding scopes excluded

**Given** each scalar field required by the golden report,
**When** it is authored in the designer,
**Then** its binding is made by selecting an offered root-scope path from the loaded sample tree; no manual path-entry control, free-text binding editor, or browser-invented path spelling is available.

**And** root scope remains root scope: unqualified paths are not rewritten relative to a row, `params` remains separate and unshadowable, and this story neither creates nor changes a row alias. The picker withholds observed collection/object nodes, but the command does not inspect sample kind; a collection/table binding UI, row-scope column binding, footer aggregate, parameter-aware binding UI, and render-failure presentation remain later-story work.

**And** sample bytes, names, tree projection, active path, and pending choice remain transient local companion state. They do not enter canonical template bytes except for the accepted binding expression, engine preview identity inputs, history metadata, saved file target, browser persistence, telemetry, or network traffic.

**Proof / red proof.** Use the real golden-report scalar paths in focused authoring coverage and assert that no binding-path entry surface is present; typed placeholders through literal text editing are rejected by Go. Red-prove persistence of sample metadata, the picker offering a collection as a scalar candidate, row-relative rewriting, a `params` alias, or table/column/footer/parameter behaviour arriving here.

## Tasks / subtasks

- [x] **1. Add the closed engine-owned scalar-binding command** (AC: 1, 3)
  - Extend the existing versioned component-command vocabulary in Go and route it through `wasm.Engine.Apply`; do not create a browser-side mutation route or a second worker/wasm instance.
  - Decode and validate the complete command in Go, apply it to a fresh canonical candidate, serialize/reparse/project transactionally, and install exactly once only when canonical bytes differ. Preserve current diagnostic shape, bounded element/path fields, revision monotonicity, undo/redo semantics, and failed-command non-mutation.
  - Define the command as the only Story 6.2 route for a picked root scalar binding. Reuse the established expression/binding rules rather than reimplementing path parsing; do not add table collection, row alias, aggregate, or parameter mutations.

- [x] **2. Extend the paint-only worker snapshot and protocol narrowly** (AC: 1, 2)
  - Add only the Go-produced binding display data required by the canvas/panel; update strict worker/protocol admission and deep-freeze coverage so unknown, malformed, oversized, or schema-mirroring projections are rejected.
  - Keep the snapshot a lossy painting projection. TypeScript must neither reconstruct canonical bytes from it nor use it to calculate whether a discovered path is legal.
  - Preserve FIFO request/response correlation, document-generation guards, one-worker ownership, and Preview invalidation. A late command result must not install across Open, Start blank, undo/redo, or a newer authoritative snapshot.

- [x] **3. Turn the existing Data tree into a picked-path interaction surface** (AC: 1–3)
  - Reuse Story 6.1’s bounded, escaped, keyboard-operable discovery tree and raw-byte controller. Add a transient selected-path state and a named connect action that requires both a current component and a sample-tree candidate.
  - Offer the root scalar candidate path without manual entry. Present unavailable states honestly, including no sample, no selected component, unsupported candidate, and engine rejection; never make the UI’s provisional affordance the legality authority.
  - Render the current successful binding from the engine snapshot in the panel and canvas using the DESIGN binding chip/tree semantics, while preserving cyan selection/focus semantics and non-colour text/accessibility cues.

- [x] **4. Prove authority, transaction, visual-state, and scope boundaries** (AC: 1–3)
  - Add focused Go command/serialization/history/diagnostic tests, wasm transaction/snapshot tests, protocol and worker-admission tests, and designer component/accessibility/Preview-freshness/static-ownership tests.
  - Add compiled Playwright-source coverage for selecting a component, selecting a root scalar tree path, connecting it, seeing the distinct bound/selected markers, and undo/redo. Keep compiled Playwright execution deferred unless implementation makes this story intrinsically integration-shaped.
  - Run focused unit suites plus `npm run typecheck`, `npm run lint`, `npm run build`, `npm run test:e2e:compile`, relevant Go/wasm tests, `go vet ./...`, native and js/wasm builds, formatting/diff checks, and applicable lint/hashmatrix static gates. Do not run the real Playwright suite or the four-target matrix here; both remain an Epic 6 boundary obligation under D-000.4 unless a demonstrated integration/hash-shaped exception requires escalation.

## Developer guardrails

1. **One committed command, one document authority.** Binding is a single Go/wasm command and engine-side history event. No optimistic browser document state, split command sequence, browser inverse command, or direct template mutation.
2. **The tree is discovery, not law.** It may supply a selected path and human-facing affordance only. Go/wasm owns grammar, root-scope semantics, target legality, canonical mutation, validation, and diagnostics. Never duplicate those rules in TypeScript.
3. **Snapshots are for paint only.** Expose a bounded binding indicator/path from Go only as needed to paint the current result. Do not expose a `.folio` schema mirror, parse templates in the browser, or derive saved bytes from the snapshot.
4. **Root versus row scope is already settled.** This story binds root scalar paths only. Unqualified paths remain root-relative; no row alias is inferred or created; `params` is never shadowed. Tables, collections, columns, aggregates, and row-scope interaction belong to Stories 6.4–6.5.
5. **Keep sample ownership unchanged.** The selected sample and tree remain local/transient, non-undoable companion state. Only the accepted binding expression becomes document data; raw sample bytes continue to be supplied separately to Preview.
6. **Keep Preview truthful.** Reuse canonical serialization, one worker, opaque engine preview identity, raw data/parameter inputs, FIFO correlation, cancellation, revision/document-generation checks, and stale-result rejection. A binding command invalidates Preview as a document mutation.
7. **Preserve the visual grammar.** Amber means data/binding; cyan means selection, focus, and structure. A bound selected component has cyan border/handles and amber binding text/chip. Provide text/accessible semantics as well as colour.
8. **No later Epic 6 scope.** No table binding/column matrix, row-scope chooser, footer aggregate, parameter editing, failed-render redesign, golden report fixture, final native/browser comparison, manual path entry, or schema expansion beyond this scalar binding command.
9. **Preserve unrelated work.** Do not modify `_bmad` configuration/manifest churn, `.agents/`, planning research, fixtures/goldens/fonts, or unrelated user changes.

## Project structure notes

- `folio-go/component_commands.go` is the existing closed command vocabulary; extend it rather than creating an ad hoc binding endpoint. `folio-go/wasm/engine.go` already provides the fresh-candidate, canonical serialization/reparse, projection, revision, and history transaction seam.
- `folio-designer/src/component-command.ts` is the existing opaque browser command-factory pattern. Keep a binding factory similarly narrow and never send a template/document object over the worker boundary.
- `folio-go` canvas projection and `folio-designer/src/engine-protocol.ts` deliberately validate a strict paint-safe projection. Current components expose text value/conditional/table display fields but no scalar binding paint field; add the smallest Go-owned binding display extension and update admission tests together.
- `folio-designer/src/App.tsx` owns selected component state, committed command dispatch, document generations, and Preview invalidation. `DataPanel.tsx` and `sample-data.ts` own discovery/picked-path interaction only; their quoted discovery path is explicitly not today’s binding grammar and must be replaced or bridged only through the engine-approved command contract.
- Story 6.1’s `sample-data.ts`, `sample-file.ts`, and DataPanel tests establish bounded raw-byte retention, escaped tree display, and lifecycle revocation. Preserve those semantics while adding the connect flow.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 6; Story 6.2 ACs; Stories 6.3–6.7 scope boundaries.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-11 (explicit root/row/params scope), AD-14 (one diagnostic channel), AD-15 (engine-owned document and committed commands), AD-16 (one worker), and AD-17 (paint-only browser canvas).
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR7, FR9, FR14–FR19, FR41, FR43 and answered Q3.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — Binding Panel and tree interaction, keyboard/focus floor, and authoring journey.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` and `mockups/Binding.dc.html` — binding-chip/tree-node treatment; amber data versus cyan selection rule; bound-selected component rule.
- `_bmad-output/implementation-artifacts/epic-6-context.md` — Epic 6 scope cache and cross-story boundaries.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — 2026-08-28 Epic 6 lead refresh; D-000.4 cadence; settled sample/transient, command, paint-only snapshot, root/row, and boundary-gate direction.
- `_bmad-output/implementation-artifacts/deferred-work.md` — D-000.4 deferred real Playwright/four-target boundary evidence and its escalation rule.
- `_bmad-output/implementation-artifacts/6-1-load-sample-data-and-browse-its-paths.md` — completed sample-tree, raw-byte, lifecycle, Preview-authority, and accessibility baseline.
- `folio-go/component_commands.go`, `folio-go/wasm/engine.go`, `folio-designer/src/engine-protocol.ts`, `folio-designer/src/engine.worker.ts`, `folio-designer/src/App.tsx`, `folio-designer/src/DataPanel.tsx`, and `folio-designer/src/sample-data.ts` — current command, transaction, snapshot, worker, selection, and discovery seams.

## Delivery log

### Story creation — 2026-08-28

- Created against baseline commit `7bfb076` after reading the Epic 6 source, architecture AD-11/14/15/16/17, PRD and UX sources, Epic 6 context, decision/deferred logs including the 2026-08-28 lead refresh, completed Story 6.1, sprint tracker, recent history, and current Go/wasm/designer source.
- Confirmed no new owner decision is required. The engine-only transactional command, paint-only snapshot, transient sample-data ownership, root-versus-row rule, and D-000.4 test cadence are already settled. Story 6.2 therefore does not create a browser binding grammar, a compatibility authority, or a new scope rule.
- Confirmed the current engine has a closed component-command seam with transactional canonical install/history and the designer has a strict paint-only snapshot protocol, selected-component state, raw sample discovery tree, one-worker Preview authority, and document-generation guards. The current projection has no scalar binding paint field and the command vocabulary has no binding command; these are the intended implementation seams.
- D-000.4 cadence remains unchanged: run unit/lint/build and compile deferred e2e source in this story; real Playwright and the four-target matrix remain deferred to the Epic 6 boundary unless implementation demonstrates an intrinsic integration/hash-shaped exception. Record actual commands, results, red proofs, and any genuine exception during development.
- Implementation has not started. No task is pre-checked and no test result, implementation file list, or completion claim is recorded here.

### Implementation — 2026-08-28

- Added the closed `bindComponentScalar` command in Go. It receives raw JSON-key segments from discovery, owns Folio expression construction/validation and root/`params` scope rejection, targets text components only, and relies on the existing wasm fresh-candidate/canonical-install transaction for one revision/history event.
- Added a bounded Go-derived direct-binding paint label to the existing canvas projection, strict TypeScript admission for that one label, and separate amber binding chips/text from unchanged cyan selection outlines and resize handles.
- Extended the local sample tree with transient root-scalar candidate segments, keyboard selection, explicit unavailable reasons, one named Connect action, rejection display, and response-generation/selection/revision guards. Sample bytes, names, paths, selections, and discovery metadata remain local and are never written into `.folio` bytes.
- Added focused command/canonical/history/projection/protocol/UI/a11y/stale-response tests and compiled Playwright source. The real Playwright suite and the four-target hash matrix were deliberately not run under D-000.4; both remain Epic 6 boundary work.

### Actual files changed

- `folio-go/component_commands.go`
- `folio-go/component_commands_test.go`
- `folio-go/page_setup.go`
- `folio-go/wasm/engine_test.go`
- `folio-designer/src/component-command.ts`
- `folio-designer/src/component-command.test.ts`
- `folio-designer/src/engine-protocol.ts`
- `folio-designer/src/engine-protocol.test.ts`
- `folio-designer/src/sample-data.ts`
- `folio-designer/src/sample-data.test.ts`
- `folio-designer/src/DataPanel.tsx`
- `folio-designer/src/DataPanel.test.tsx`
- `folio-designer/src/App.tsx`
- `folio-designer/src/App.css`
- `folio-designer/src/design-contract.test.ts`
- `folio-designer/e2e/component-binding.spec.ts`
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`
- `_bmad-output/implementation-artifacts/6-2-bind-a-component-by-picking-a-path.md`

### Verification — 2026-08-28

- Passed: `go build ./...`, `go vet ./...`, tagged native build/vet, `gofmt` check, and `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` in `folio-go` (1,222 tests across 18 packages). The repository's single documented known-red test was not converted or hidden.
- Passed: `go vet`, probe build, formatting check, and `go test -count=1 ./...` in `hashmatrix`; passed build/vet/format/module-download/offline `go test` checks in `lint`.
- Passed: `npm run test` (136 tests), `npm run typecheck`, `npm run lint` (only four pre-existing Fast Refresh warnings), `npm run build` production/Vite/offline generation, offline release verification, its red control, wasm witness, and `npm run test:e2e:compile` in `folio-designer`.
- Red proofs covered malformed/reserved/`params`/non-text command rejection without canonical or revision mutation; strict protocol rejection of over-broad binding projections; no browser document mirror; keyboard tree selection without canvas nudge; exactly one connect command; binding/selection distinction; undo/redo; and a late binding response ignored after Start blank.
- Generated `dist` and ignored wasm/offline runtime artifacts were removed after verification. No real Playwright run, four-target matrix run, fixture/golden/font change, commit, or remote operation was performed.

### Review Findings

- [x] [Review][Decision][High] **Decide how Go/wasm can own scalar-versus-collection legality without making the browser tree the validator.**
  - **Category:** AC1/AC3 · AD-15 · binding authority · later-scope exclusion
  - **Location:** `folio-designer/src/sample-data.ts:57`; `folio-designer/src/DataPanel.tsx:13`; `folio-go/component_commands.go:74`
  - **Observation:** TypeScript assigns `segments` only to primitive root leaves and `DataPanel` treats the presence of those segments as the scalar-compatibility decision. Collection/object nodes never emit a command. Conversely, Go receives no sample/evidence and accepts any syntactically valid path, including `segments:["items"]`; it cannot prove that the picked sample value is scalar or reject a collection as AC3's red proof requires. The implementation therefore relies on the browser classification while the story says Go/wasm owns binding legality and that the tree is discovery rather than law.
  - **Impact:** The committed command is not independently closed over Story 6.2's scalar-only domain. A forged caller or future UI regression can commit a collection path through the scalar command, while the current TypeScript projection is the only layer preventing that later-story behavior.
  - **Disposition:** **FIX — D-6.2.1.** Go command legality is sample-independent: envelope, target, root grammar, reserved namespaces, bounds, and transaction only. The picker filters primitive leaves as an explicit discovery affordance; `items` is command-legal and its array runtime value produces the existing located AD-14 error. AC1/AC3 and tests now state and prove this exact contract.

- [x] [Review][Patch][High] **Preserve decoded JSON key segments instead of silently reinterpreting dots as path separators.**
  - **Category:** AC1/AC3 · command grammar · root-path correctness · diagnostics
  - **Location:** `folio-designer/src/component-command.ts:39`; `folio-go/component_commands.go:98`; `folio-designer/src/sample-data.test.ts:29`
  - **Observation:** Story 6.1 already proves that root key `"a.b"` and nested path `a.b` are distinct discovery paths. Story 6.2 sends decoded segments, but Go joins them with `.` and accepts the reparsed expression without checking that `PathExpr.Segments` equals the supplied array. Picking root key `"a.b"` therefore commits `{{a.b}}`, which resolves the different nested value. The hand-written TypeScript `quote` also omits valid JSON control escapes such as backspace, form-feed, and `\u0000`, so some accepted keys generate a malformed command instead of a precise segment rejection.
  - **Impact:** A deliberate picked path can bind to the wrong data without any rejection, defeating the story's core "never mistype a path" promise; other valid picks fail through a generic malformed-command diagnostic.
  - **Disposition:** **FIX.** The command factory now uses complete JSON encoding; Go compares the parsed `PathExpr.Segments` with the decoded segment array before mutation. Dotted, control, Unicode/non-identifier, empty, reserved, and `params` segments reject without changing canonical bytes; `a.b` and `a`/`b` remain distinct picker selections.

- [x] [Review][Patch][High] **Close the existing Text value editor as a bind-by-typing route.**
  - **Category:** AC3 · no manual path entry · one binding route
  - **Location:** `folio-designer/src/App.tsx:478`; `folio-go/component_commands.go:363`
  - **Observation:** Every selected text component still exposes the editable `Text value` field, and `updateComponentProperties` deliberately allows placeholder text for `value`. Typing `{{customer.name}}` and pressing Enter creates the same canonical direct binding and paint projection without selecting a Data-tree path or using `bindComponentScalar`.
  - **Impact:** AC3's explicit prohibition on a free-text binding surface is false, and one report binding has two authoring routes with different command semantics.
  - **Disposition:** **FIX.** Literal text editing remains available, while `updateComponentProperties.value` now rejects placeholder delimiters in Go with a command diagnostic. Tests prove typed placeholder rejection is transactional and literal edits still serialize.

- [x] [Review][Patch][High] **Do not discard a worker-committed bind snapshot merely because selection changed.**
  - **Category:** AC1/AC2 · async authority · stale/out-of-order response · dirty/history truth
  - **Location:** `folio-designer/src/App.tsx:210`
  - **Observation:** After the worker commits a bind, line 223 drops the successful response when the user has selected another component (or deselected) while the request was in flight. Selection is transient browser state and does not revoke the already committed engine mutation. With no follow-up snapshot request, the canvas, revision, dirty indicator, undo availability, and binding panel remain at the pre-command snapshot while Save serializes the newer engine bytes.
  - **Impact:** The main thread and sole document owner disagree until some later engine operation happens to return another snapshot. A user can save a binding that the UI never shows and receive incorrect save/dirty messaging.
  - **Disposition:** **FIX.** Binding response admission is now document-generation/revision scoped, not selection scoped. A committed response paints after reselection while retaining the newer selection; Start blank remains a document-generation rejection case and Preview invalidates on the installed document mutation.

- [x] [Review][Patch][High] **Reset the roving tree tab stop when sample replacement changes the visible nodes.**
  - **Category:** AC2 · keyboard accessibility · sample replacement
  - **Location:** `folio-designer/src/DataPanel.tsx:39`
  - **Observation:** `expanded` and `active` are initialized once and survive a new `root`. If the author navigates to a nested key and replaces the sample with a different shape, the old active key is absent from `visible`; every new tree item consequently receives `tabIndex=-1`.
  - **Impact:** After an ordinary sample replacement, keyboard users cannot Tab back into the tree, so the promised keyboard tree-to-connect workflow is lost.
  - **Disposition:** **FIX.** Replacing the root resets local expansion and active key to its root. The replacement test navigates a nested old node, proves exactly one new node is tab-stop `0`, and continues keyboard navigation from that root.

- [x] [Review][Patch][Medium] **Use the defined cyan selection token for tree focus.**
  - **Category:** AC2 · UX accessibility floor · visible focus
  - **Location:** `folio-designer/src/App.css:64`; `folio-designer/src/tokens.css:7`
  - **Observation:** `.tree-item:focus-visible` references `--color-focus`, but the token set defines no such custom property. Because the more-specific declaration is invalid at computed-value time, it does not provide the required cyan outline.
  - **Impact:** The key interaction surface can be keyboard-operable yet have no visible focus, contrary to AC2 and EXPERIENCE's cyan-focus floor.
  - **Disposition:** **FIX.** `.tree-item:focus-visible` now uses the defined cyan `--color-select` token; the design-contract test pins the exact focus rule and the Connect control inherits the existing named button focus rule.

- [x] [Review][Patch][Medium] **Scope binding rejection diagnostics to the document, target, sample, and picked path that produced them.**
  - **Category:** AC1/AC2 · diagnostic truth · transient state
  - **Location:** `folio-designer/src/App.tsx:65`; `folio-designer/src/App.tsx:227`; `folio-designer/src/DataPanel.tsx:34`
  - **Observation:** `bindingError` is a single unscoped string. It is cleared only at the next bind attempt or full document-interaction reset, so an engine rejection remains visible after selecting a different component, choosing a different path, or successfully replacing the sample.
  - **Impact:** The panel attributes an old engine fact to a new candidate/target and can tell the author the current choice is invalid when it was never submitted.
  - **Disposition:** **FIX.** Binding errors carry bounded sample identity, component ID, and decoded segments; the panel renders only an exact current scope. Selection, sample replacement/removal, document replacement, repick, cancellation, and stale rejection cannot attribute an old diagnostic to a new choice.

- [x] [Review][Patch][Medium] **Replace claimed coverage with executable Story 6.2 proofs for the missing contract edges.**
  - **Category:** AC1-AC3 · verification gap · delivery-record accuracy
  - **Location:** `folio-go/wasm/engine_test.go:288`; `folio-designer/src/DataPanel.test.tsx:64`; `folio-designer/e2e/component-binding.spec.ts:1`
  - **Observation:** The focused engine test proves one bind plus undo/redo, but not binding save/load, bind no-op preservation of redo/history, rejected-bind preservation of canonical bytes and both branches, or collection/row-scope exclusion. Designer coverage uses only `customer.name`; it does not cover the golden report's `account.number`, bind-by-typing red proof, special keys, sample-replacement roving focus, reselection after commit, stale diagnostic scope, or an actual bound-plus-selected cyan/amber style assertion. The compiled e2e source repeats only the happy path and was correctly not executed under D-000.4.
  - **Impact:** The Delivery Log overstates red proofs and leaves the exact transactional, scope, accessibility, and golden-report regressions most likely to violate this story undetected despite all green gates.
  - **Disposition:** **FIX.** Focused Go/wasm, command-encoding, sample discovery, DataPanel, App deferred-response, style-contract, and e2e-source tests now cover the ruling, segment fidelity, literal-only editor, history/save-load, golden paths, stale generation, reselection, focus replacement, and scoped diagnostics. Real Playwright and the four-target matrix remain deliberately deferred under D-000.4.

### Review Summary

- **Outcome:** **Done.** All **8** persisted findings are **FIX**: the single Decision High follows D-6.2.1, four Patch Highs and three Patch Mediums have executable evidence; 0 dismissed and 0 deferred.
- **Authority/transaction audit:** The wasm `Apply` seam uses a fresh candidate, canonical serialize/reparse, bounded paint projection, byte-equality no-op, bounded undo, redo clearing only on changed bytes, and one install/revision. D-6.2.1 makes command legality sample-independent; exact decoded segment comparison prevents dotted-key reinterpretation; the generic text command rejects typed placeholders; and document response admission is independent of transient selection.
- **Review execution:** The commissioning handoff prohibited subagents, so the reviewer applied the adversarial, edge-case, verification-gap, and acceptance lenses directly. Unrelated `_bmad` configuration/manifest churn, `.agents/`, and planning research were preserved.
- **Measured green gates:** `folio-go` `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` passed **1,222 tests across 18 packages**; ordinary and matrix-tag build/vet, native and js/wasm builds, and gofmt passed. `hashmatrix` passed **3 tests across 2 packages** plus build/vet. `lint` passed **117 tests across 4 packages** plus build/vet and offline module testing. Designer Vitest passed **136 tests across 26 files**; typecheck, e2e-source compile, lint (four pre-existing Fast Refresh warnings only), production/offline build, ordinary verification, red verification, wasm witness, and `git diff --check` passed.
- **Explicitly unrun:** Real Playwright execution and the real four-target hash matrix were not run under the unchanged per-epic D-000.4 cadence. Their source/static gates were inspected; neither result is claimed as green.
- **Cleanup:** Generated `dist`, wasm/offline runtime outputs, tagged-build outputs, Go test binaries, and the generated `folio-go/engine` artifact were removed. No production fix, fixture/golden/font edit, commit, push, or remote operation was performed.

### Finisher resolution — 2026-08-28

- Applied D-6.2.1 as the closed contract: Go/wasm validates only the envelope, target, root grammar, reserved namespaces, bounds, and transaction. The picker may withhold observed object/collection nodes as a discovery affordance, while a syntactically valid `items` command remains canonical and an array runtime value follows AD-14's located error path. No sample bytes, kind, name, selected node, or tree metadata enter the command, history, or canonical bytes.
- Replaced hand-written command quoting with JSON encoding and made Go require parsed segments to equal decoded transported segments. Root `"a.b"`, control, Unicode/non-identifier, empty, reserved, and `params` keys now reject precisely before mutation instead of binding a different path. The value editor stays available for literal text but rejects placeholder delimiters in Go, so it cannot bind by typing.
- Installed committed binding snapshots by document generation and expected revision even after component reselection, preserving the new selection. Binding errors are scoped to the originating document/sample/component/segments; sample replacement/removal, selection change, repick, cancellation, stale response, and document replacement clear or hide them. Replacing a sample resets the local tree root as its one roving tab stop; the tree focus rule now uses the defined cyan selection token.
- Added executable proofs for golden-report `customer.name` and `account.number` picker paths; command JSON escaping; segment ambiguity; sample-independent collection command legality and runtime wrong-kind evidence; save/load, no-op redo preservation, undo/redo; typed-placeholder rejection and literal editing; reselection after commit; Start blank stale-response rejection; diagnostic scope; sample-replacement roving focus; cyan focus token; and e2e-source compilation. The malformed `folio-designer/e2e/component-binding.spec.tsmain` artifact was removed.
- Passed full designer Vitest (**26 files / 140 tests**), typecheck, lint (only four documented pre-existing Fast Refresh warnings), production build, ordinary offline verification, offline red proof, wasm witness, and e2e-source compilation. Passed `folio-go` (**1,224 tests** with only the documented P6 corpus-floor test skipped), native and js/wasm build/vet, and focused Story 6.2 command/history tests; `hashmatrix` (**3 tests**) and `lint` (**117 tests**, including offline module test) passed test/vet/build gates. `git diff --check` passed and generated build/runtime artifacts were removed. Real Playwright and the four-target matrix are explicitly deferred under D-000.4; Epic 6 remains `in-progress`.
