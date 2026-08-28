---
baseline_commit: 49df7d3
story_key: 6-1-load-sample-data-and-browse-its-paths
status: done
created: 2026-08-28
---

# Story 6.1: Load sample data and browse its paths

**Epic:** 6 — A template author can bind a report to data and build the golden report  
**Covers:** FR9 · UX-DR7, UX-DR13

## In plain terms (read this first if you just want the gist)

This story shipped a docked local JSON discovery panel. An author can load or replace a valid sample, inspect its bounded object/collection/scalar tree with a keyboard, and see exact escaped examples without leaving the authoring workspace. The panel names the selected local file, explains that discovery paths are not component bindings, and keeps normal canvas, save, and history controls available.

The accepted file bytes are retained unchanged for Preview; a separate bounded parser validates one JSON document and creates only the display projection. Large files are rejected before reading, BOM-prefixed files are rejected to match the Go decoder, and long/deep/wide content is visibly truncated rather than made into an unbounded browser object or DOM tree. Invalid, cancelled, unreadable, and oversized replacements leave the prior accepted sample untouched.

Sample state remains a transient local companion: it is not stored in `.folio`, commands, history, dirty/save state, targets, storage, telemetry, or network traffic. Replacement and document lifecycle are authority-scoped so a late picker result or Preview operation cannot revive cleared data across Open or Start blank. Story 6.1 deliberately does not bind a component, define a data-path grammar, edit tables, or redesign parameter entry.

## Story

As a template author,  
I want to see the shape of the data my developer will supply,  
so that I am binding to something real rather than to a name I invented.

## Acceptance criteria

### AC1 — The empty binding panel is explicit and does not block authoring

**Given** the workspace before sample data has been loaded,  
**When** its docked Data/Binding panel is displayed,  
**Then** it states that binding is unavailable because no sample data is loaded and provides a keyboard-operable local-file load action.

**And** the panel remains part of the workspace beside the existing authoring surfaces, never a route, modal destination, or separate screen; component placement, selection, properties, save, undo/redo, and Design/Preview mode controls remain available.

**And** unavailable future binding actions are absent or truthfully disabled with a stated reason. The panel must not imply that typing a path, editing raw data, or binding a component is available in this story.

**Proof / red proof.** Component and compiled e2e-source coverage proves the initial named empty state, keyboard access to the load action, docked placement, and continued canvas authoring. Red-prove a dead/unnamed load action, a separate destination, or a disabled canvas caused solely by absent sample data.

### AC2 — A valid local JSON sample becomes a bounded, navigable path tree

**Given** the author selects one valid JSON document from their local machine,  
**When** loading succeeds,  
**Then** the docked panel names the selected local file and presents its paths as an expandable, keyboard-operable tree containing nested objects, scalar leaves, explicit nulls, and collections.

**And** object nodes expose their children; scalar and null leaves show a bounded, escaped example value and type; collection nodes use the established `[]` path notation, show a bounded item count/example shape, and preserve source ordering for visible children. The tree is a local discovery projection only: it does not parse or mutate the template, invent a binding, choose a component, or declare table-only compatibility before Story 6.2.

**And** the browser retains the exact accepted raw sample bytes for Preview identity/render input while using a separate bounded local inspection representation for display. It must not normalize, reserialize, coerce numeric literals, or send a browser-derived document model to the engine.

**And** hostile or unusually deep/wide input is bounded before it can make the panel unresponsive: enforce the existing worker-payload byte ceiling and explicit depth/node/string-preview/item-preview limits, disclose truncation in the UI, and keep all displayed content escaped.

**Proof / red proof.** Test nested objects, arrays of objects, arrays of scalars, empty arrays/objects, null, booleans, long strings, large exact decimal literals, key ordering, expansion/collapse and keyboard tree semantics. Red-prove raw HTML execution, a tree that changes raw preview bytes, unbounded traversal, float coercion, or treating an array member as a component binding.

### AC3 — Invalid, cancelled, and replacement loads are truthful and non-destructive

**Given** a selection is cancelled, unreadable, oversized, or not one valid JSON document,  
**When** the author attempts to load it,  
**Then** the panel gives a concise local error and retains the previously accepted sample and tree, if any; cancellation is quiet and leaves state unchanged.

**And** selecting a new valid sample replaces the old sample atomically: its file label, discovery tree, raw Preview input, preview identity authority, and any relevant local status change together. No partial tree, prior bytes, or stale success can become authoritative.

**And** loading is local-only and transient. The sample bytes, path, handle, name, discovery tree, expansion state, errors, and UI status do not enter `.folio` bytes, engine command/history, dirty/save state, file target, browser persistence, telemetry, or network traffic. Loading is expressly not undoable; successful reload/start-blank clears the sample and returns the panel to its empty state.

**Proof / red proof.** Cover cancelled picker, malformed/trailing JSON, read failure, byte-limit rejection, replacement while Preview work is pending, reload/start-blank clearing, and no-data undo/redo. Red-prove replacement that retains prior bytes/tree, a document revision/dirty/history change, sample metadata in serialized template bytes, or network/storage use.

## Tasks / subtasks

- [x] **1. Establish one bounded local sample-data controller and file boundary** (AC: 1–3)
  - Reuse or narrowly extend the established local-file capability/fallback pattern without giving the template file abstraction knowledge of data structure.
  - Define immutable transient sample state: exact accepted bytes, local display name, bounded parsed discovery projection, and explicitly local error/status. Clear it on successful document replacement and worker reinitialization; never serialise it or send it as a command.
  - Validate one complete JSON document before replacing state. Preserve numeric lexemes for the raw engine input and keep the display parser independent of the `.folio` model.

- [x] **2. Build the docked empty and populated Data panel** (AC: 1, 2)
  - Compose the panel with the existing workspace/property surface; add a clear, keyboard-operable load/replace action and honest empty/error/cancelled states.
  - Implement a semantic tree with native disclosure/tree keyboard behavior, visible select-token focus, named controls, expanded/collapsed state, bounded previews, and collection notation. Follow DESIGN’s tree-node, mono machine-value, bind accent, and shape-before-colour rules.
  - Keep Story 6.2’s connect-to-selected-component action, bound canvas marking, type compatibility, and Story 6.4’s data-grid/table controls out of the surface.

- [x] **3. Integrate accepted bytes once with existing Preview authority** (AC: 2, 3)
  - Replace the temporary hand-edited sample-data route with the accepted local sample bytes while preserving Story 5.11’s engine-owned five-input identity, cancellation, token/generation/revision checks, stale truthfulness, one worker, and one PDF.js owner.
  - A missing sample must remain a visible unavailable Preview-input state, not silently become fabricated `{}` data. Keep raw parameter editing until Story 6.3.

- [x] **4. Prove bounds, lifecycle, and non-document ownership** (AC: 1–3)
  - Add focused controller/parser, component/accessibility, file-boundary, preview-authority, and static ownership tests; add compiled Playwright-source coverage for local loading, tree navigation, and replacement/cancellation.
  - Run focused designer tests, `npm run typecheck`, `npm run lint`, `npm run build`, offline verification/red verification, `npm run test:e2e:compile`; run relevant Go/wasm tests, `go vet ./...`, native and js/wasm builds, formatting/diff checks, and applicable lint/hashmatrix static gates.
  - Do not run compiled Playwright or the four-target hash matrix in this story. Record both as deferred under D-000.4 for the Epic 6 boundary, unless implementation proves this story inherently hash/integration-shaped.

## Developer guardrails

1. **Sample data is not a template feature.** It is a local, transient companion input. Never write its bytes, name, path, handle, or tree state into `.folio`, document commands, engine history, saved revision, or browser persistence. It is not undoable.
2. **One authority per concern.** Go/wasm remains authoritative for template parsing, mutations, render identity, and rendering. TypeScript may parse bounded raw sample data solely for a local discovery projection; it must not become a second template schema, binding evaluator, data coercer, or engine identity calculator.
3. **Raw bytes remain exact.** Retain the accepted file bytes unchanged for Preview. The display projection may never reserialize JSON or narrow numbers; valid display parsing must not alter the engine input.
4. **Bound before display.** Enforce byte, depth, node, child/item preview, key, and string-preview ceilings. Escape all values and make truncation visible. No arbitrary HTML, unbounded recursion, unbounded rendering, or source-data mutation.
5. **Accessibility is functional.** Use named native controls and visible focus. The tree must support a logical keyboard route; disabled/unavailable states state why. Do not make the Data panel pointer-only or use colour alone to signal a bindable-looking node.
6. **Preserve Preview truth.** A sample replacement invalidates Preview through the existing opaque engine identity path. Keep one worker, one wasm instance, canonical template serialization, raw three-input render, cancellation, and stale-result rejection.
7. **Do not pull forward later Epic 6 work.** No component binding, manual path entry, type-match/connect action, bound styling, row-scope chooser, parameter-aware UX, Table Editor matrix, failed-render redesign, golden report, or final round-trip in this story.
8. **Preserve unrelated work.** Do not modify `.agents/`, `_bmad` configuration/manifest churn, planning research, fixture/golden/font sources, or unrelated user changes.

## Project structure notes

- Inspect the existing designer file-access capability and fallback before adding a narrow sample-read seam; data selection must not contaminate template open/save target ownership.
- The present `App` holds a temporary raw sample-data string for Preview. Replace it with a named transient sample-data controller/state module and a focused docked Data-panel/tree component rather than growing more unstructured state in the application shell.
- The existing engine protocol already limits payloads and carries raw byte inputs for identity/render. Reuse those limits and preserve its strict request/correlation rules; Story 6.1 should not add a data-tree operation to wasm.
- Existing Preview freshness code owns response authority. Wire accepted raw bytes into that seam once; display-tree expansion and load errors are UI-only and must not affect document state.
- Existing Story 5.12 keyboard/focus work is the baseline. Extend it for the concrete Data tree, but do not reinterpret its honest table/binding placeholders as implemented binding.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 6 and Story 6.1 acceptance criteria; Stories 6.2–6.7 scope boundaries.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-14 through AD-18, especially AD-15 (engine-owned document; sample loading not undoable) and AD-16 (one worker/three raw preview inputs).
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR7, FR9, FR14, FR41 and Q8; sample JSON is separate from the template.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — S3 Binding Panel, empty-data state, tree-node semantics, interaction/accessibility requirements, and KF-1.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` and `mockups/Binding.dc.html` — docked panel composition, tree-node, binding-chip, type, focus, and accent rules.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.4; settled 6.1/6.2 ruling that sample JSON path is not persisted; 5.10/6.3 preview-input split.
- `_bmad-output/implementation-artifacts/5-11-never-show-a-stale-preview.md` and `5-12-diagnostics-that-locate-and-an-interface-that-can-be-driven.md` — current raw-input/preview authority, accessibility, keyboard, and non-undoable sample-data seams.
- `folio-designer/src/App.tsx`, `file/`, `engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, and `preview/` — current local-file, raw-input, one-worker, and stale-preview seams to preserve.

## Delivery log

### Story creation — 2026-08-28

- Created against baseline commit `49df7d3` from the Epic 6 source, current sprint tracker, architecture spine, PRD, UX EXPERIENCE/DESIGN and Binding mockup, decision/deferred logs, completed Epic 5 records (especially Story 5.12), and current designer/Go source.
- Confirmed the material design policy is already settled, not open: sample JSON is a separate local companion and its path is not persisted in `.folio`. No owner decision is required for Story 6.1.
- Confirmed the current designer already has one worker/wasm instance, engine-owned opaque preview identity, transient raw data/parameter inputs, local file capability seams, stale-result guards, and keyboard/focus foundations. It has no dedicated local sample-file controller or navigable path tree.
- D-000.4 cadence remains unchanged: run unit/lint/build and compile deferred e2e source every story; compiled Playwright and the four-target matrix are deferred to the Epic 6 boundary unless implementation proves an inherent integration/hash-shaped exception. Record actual commands, results, red proofs, and any real exception during development.
- Implementation has not started. No task is pre-checked and no test result, implementation file list, or completion claim is recorded here.

### Developer record — 2026-08-28

- Added transient bounded sample loading, exact raw-byte Preview input, a docked discovery panel, and a local-file capability seam without putting sample state in `.folio`, commands, history, or save state.
- Added parser, panel, Preview authority, static ownership, and compiled e2e-source coverage; later-epic binding/table/parameter work remains absent.

### Finisher record — 2026-08-28

- Fixed all eight independent review findings: document-scoped sample-load/Preview revocation, complete roving tree keyboard semantics, size-before-read and non-materializing validation bounds, bounded unambiguous display projections, BOM compatibility, missing focused witnesses, chronological records, and object property wording.
- Confirmed `epic-6-context.md` is an accurate, legitimate Epic 6 scope cache and included it with this Story 6.1 delivery. Story 6.1 is done; Epic 6 remains in progress.

### Actual file list

- `folio-designer/src/sample-data.ts`, `folio-designer/src/sample-file.ts`, `folio-designer/src/DataPanel.tsx`, `folio-designer/src/App.tsx`, `folio-designer/src/main.tsx`, `folio-designer/src/file/capability.ts`, `folio-designer/src/App.css`
- `folio-designer/src/sample-data.test.ts`, `folio-designer/src/sample-file.test.ts`, `folio-designer/src/DataPanel.test.tsx`, `folio-designer/src/App.test.tsx`, `folio-designer/src/engine-ownership-contract.test.ts`, `folio-designer/src/file/file-access-contract.test.ts`, `folio-designer/src/canvas-authority-contract.test.ts`
- `folio-designer/e2e/sample-data.spec.ts`, `folio-designer/e2e/application-shell.spec.ts`
- `_bmad-output/implementation-artifacts/6-1-load-sample-data-and-browse-its-paths.md`, `epic-6-context.md`, `sprint-status.yaml`, and the accurate Epic 6 lead-grounding refresh in `folio-mvp-decision-log.md`

### Verification

- Measured finisher gates are recorded in the Review Summary below. Compiled Playwright execution and the four-target hash matrix remain explicitly deferred under D-000.4 to the Epic 6 boundary; neither was run for this non-integration/non-hash-shaped story.

## Suggested Review Order

**Transient sample ownership**

- Validate once, retain raw bytes, and create only a bounded display projection.
  [`sample-data.ts:32`](../../folio-designer/src/sample-data.ts#L32)

- Keep adversarial nesting out of the browser JSON parser before full validation.
  [`sample-data.ts:167`](../../folio-designer/src/sample-data.ts#L167)

- Reuse local-picker capability without giving template file access sample semantics.
  [`sample-file.ts:5`](../../folio-designer/src/sample-file.ts#L5)

**Preview authority and lifecycle**

- Feed Preview a transport copy of accepted raw bytes, never the discovery projection.
  [`App.tsx:116`](../../folio-designer/src/App.tsx#L116)

- Replace samples atomically and preserve prior accepted state on local failures.
  [`App.tsx:254`](../../folio-designer/src/App.tsx#L254)

**Docked authoring surface**

- Show an honest empty state and native, keyboard-operable discovery tree.
  [`DataPanel.tsx:3`](../../folio-designer/src/DataPanel.tsx#L3)

- Keep Data docked beside existing workspace and Preview-input surfaces.
  [`App.tsx:415`](../../folio-designer/src/App.tsx#L415)

**Proof**

- Cover byte fidelity, numeric lexemes, malformed input, and projection bounds.
  [`sample-data.test.ts:8`](../../folio-designer/src/sample-data.test.ts#L8)

- Cover empty authoring, replacement/cancellation, and engine raw-byte authority.
  [`DataPanel.test.tsx:20`](../../folio-designer/src/DataPanel.test.tsx#L20)

## Review Findings

### Finding 1: A pending sample load can survive document replacement and install stale sample/Preview authority

- **Severity:** Blocker
- **Category:** Async state / Preview freshness / lifecycle isolation
- **Location:** `folio-designer/src/App.tsx:253-268`, `:269-289`, `:316-328`
- **Observation:** `loadSample` has no request generation, cancellation token, or document-generation guard. `open` and `startBlank` invalidate Preview only when they begin, then clear sample state only after their worker work succeeds. A sample picker started before either operation can resolve after that invalidation and install into the replacement document; conversely, it can resolve during the replacement, schedule identity/render work, and then outlive `clearSampleData`, which does not revoke that newer Preview generation. Because document revisions can restart at the same numeric value, the existing revision check is not a document-identity guard for this interleaving.
- **Impact:** A successful open/start-blank can finish with sample data repopulated from the prior document interaction, or the panel can say no sample is loaded while a PDF produced from those cleared bytes becomes current. This violates AC3's atomic replacement, successful document-replacement clearing, and never-stale Preview authority.
- [x] **Resolution — FIXED:** `sampleLoadGeneration` now scopes picker results, is synchronously revoked when Open/Start blank begins and when samples clear, and `runPreview` also checks `documentGeneration`. `clearSampleData` invalidates Preview authority itself. `DataPanel.test.tsx` holds a picker promise across an equal-revision Start blank and proves its late bytes never install or reach identity.
- **Related:** AC3; Developer guardrails 1 and 6.

### Finding 2: The rendered ARIA tree is not keyboard-navigable as a tree

- **Severity:** Major
- **Category:** Accessibility / tree semantics
- **Location:** `folio-designer/src/DataPanel.tsx:13-17`
- **Observation:** `role="treeitem"` is placed on an unfocusable `<li>`, while the focusable `<summary>` has neither treeitem semantics nor `aria-expanded`. Scalar/null leaves have no focusable descendant at all. Native `<details>` provides Enter/Space disclosure for the currently focused summary, but it does not implement tree-item focus, Up/Down sibling navigation, Left/Right parent/child navigation, or a roving tab stop.
- **Impact:** Keyboard users cannot traverse or inspect every path, and assistive technology is not told which treeitems expand. The load button focus assertion does not prove AC2's keyboard-operable navigable tree.
- [x] **Resolution — FIXED:** The panel now uses one roving `treeitem` tab stop, `aria-expanded` on branches, and Arrow Up/Down/Left/Right, Home, End, Enter, and Space behavior. Every scalar/null leaf is focusable. `DataPanel.test.tsx` drives real keys through a branch and scalar leaf and asserts focus plus expanded state; compiled e2e source follows the same route.
- **Related:** AC1, AC2; EXPERIENCE accessibility floor.

### Finding 3: The byte ceiling is enforced only after unbounded file allocation and full JSON materialization

- **Severity:** Major
- **Category:** Resource bounds / hostile input
- **Location:** `folio-designer/src/sample-file.ts:14-22`, `:30-42`; `folio-designer/src/sample-data.ts:32-42`
- **Observation:** Both file adapters call `file.arrayBuffer()` before checking `File.size`, so an arbitrarily large local file is fully allocated before the 8 MiB engine ceiling runs. After decoding, `acceptSampleData` calls `JSON.parse(text)` and materializes the complete object/array graph even though that result is discarded; only the second parser's retained projection is bounded.
- **Impact:** A wide valid JSON file can consume memory proportional to its complete source graph, and a huge selected file can allocate far beyond the declared payload ceiling. The implementation therefore does not bound hostile input before it can make the panel unresponsive as AC2 requires.
- [x] **Resolution — FIXED:** Both sample-file capability tiers reject `File.size` before `arrayBuffer()`. `DiscoveryParser` is now the one-document grammar validator and bounded projection builder; `JSON.parse` is absent. `sample-file.test.ts` proves an oversized File's read method is never called, while parser tests cover malformed, trailing, deep, and wide input.
- **Related:** AC2, AC3; Developer guardrail 4.

### Finding 4: Numeric examples and flattened paths bypass the projection's display limits and can misdescribe keys

- **Severity:** Major
- **Category:** Projection bounds / discovery fidelity
- **Location:** `folio-designer/src/sample-data.ts:83-115`, `:134-139`, `:162-163`; `folio-designer/src/DataPanel.tsx:15-17`
- **Observation:** String values and labels are bounded, but number lexemes are stored and rendered in full, and `childPath` concatenates the original unbounded key rather than its bounded display form. One valid number or key can therefore contribute nearly the whole 8 MiB payload to a DOM text node. Flattening raw keys with `.` and appending `[]` also makes structurally different inputs indistinguishable (`{"a.b":1}` and `{"a":{"b":1}}` both present `a.b`; keys already ending in `[]` collide with collection notation).
- **Impact:** The UI violates AC2's bounded scalar/key previews and can present a false data shape—the exact opposite of discovery's purpose. It also starts inventing an ambiguous browser-owned binding grammar before Story 6.2.
- [x] **Resolution — FIXED:** Display paths use bounded, JSON-quoted inspection segments joined by `›`, with `[]` reserved for collection shape; they are explicitly labelled non-binding display paths. Number, string, key, label, and path text are capped with visible ellipses. `sample-data.test.ts` covers long number/key plus dotted, bracket, and empty keys and proves distinct structures remain distinct.
- **Related:** AC2; Developer guardrails 2, 4, and 7.

### Finding 5: UTF-8 BOM input is accepted by discovery but rejected by the production Go decoder

- **Severity:** Major
- **Category:** Raw-byte validation / engine compatibility
- **Location:** `folio-designer/src/sample-data.ts:35-42`; `folio-go/internal/bind/value.go:160-184`
- **Observation:** The default `TextDecoder('utf-8')` strips a leading UTF-8 BOM, so browser validation sees `{}` and accepts the sample while retaining the original BOM-prefixed bytes. Those exact bytes later reach Go's `json.Decoder`, which does not strip the BOM and rejects them as invalid JSON.
- **Impact:** The panel can truthfully claim a sample was accepted and show its tree, yet exact Preview fails on the same accepted bytes. Local discovery and the production input contract disagree at the raw-byte boundary.
- [x] **Resolution — FIXED:** `acceptSampleData` rejects UTF-8 BOM bytes before decoding or state replacement, preserving raw-byte/Go compatibility without changing engine ownership. `sample-data.test.ts` witnesses the rejection; Go's existing `bind.DecodeData` remains the production decoder and its BOM behavior was rechecked with the Go suite.
- **Related:** AC2, AC3; Developer guardrail 3.

### Finding 6: The claimed proof omits the story's highest-risk behaviors

- **Severity:** Major
- **Category:** Verification gap
- **Location:** `folio-designer/src/DataPanel.test.tsx:19-40`; `folio-designer/src/sample-data.test.ts:7-34`; `folio-designer/e2e/sample-data.spec.ts:3-13`; Delivery Log §Verification
- **Observation:** The new Playwright source is non-vacuous—it loads a real file and observes the docked tree—but it never drives tree keys, cancellation, invalid/oversized input, replacement, reload/start-blank clearing, or Preview replacement. The component test focuses only the load button, asserts cancellation only by seeing the old filename, and never covers read failure, malformed replacement, history/dirty/save isolation, persistence/network absence, render-time raw bytes, pending Preview work, or document-replacement races. There is no focused test of either `SampleFileAccess` implementation. Nevertheless the Delivery Log claims parser/controller, accessibility, file-boundary, replacement/cancellation, and Preview-authority coverage.
- **Impact:** The passing 128-test suite and e2e-source compilation do not exercise the mechanisms behind Findings 1–5, so the proof obligations in AC1–AC3 remain materially unmet despite a green gate record.
- [x] **Resolution — FIXED:** Focused parser, file-boundary, panel, and App race tests now witness byte fidelity, parser bounds/BOM, no oversized read, keyboard focus, invalid/cancelled preservation, exact identity data, and equal-revision Start blank clearing. Compiled e2e source covers local load, key traversal, and malformed replacement preservation. The Delivery Log now distinguishes measured gates from Epic-boundary deferrals and does not claim unexecuted Playwright/matrix results.
- **Related:** AC1–AC3 proof clauses; Task 4.

### Finding 7: Delivery Log entries are reverse chronological

- **Severity:** Minor
- **Category:** Process / record integrity
- **Location:** This story, `## Delivery log`
- **Observation:** `Developer record — 2026-08-28` appears before the older `Story creation — 2026-08-28` entry. The run-dev-cycle record is required to be append-only in oldest-to-newest order.
- **Impact:** Future finisher/reviewer entries cannot be read as a reliable delivery sequence, and appending more entries would preserve an already-corrupt chronology.
- [x] **Resolution — FIXED:** The Delivery Log below is now ordered oldest-to-newest: Story creation, developer implementation, then this finisher record. Future records append after it.

### Finding 8: Object property counts are announced as item counts

- **Severity:** Minor
- **Category:** UI truthfulness
- **Location:** `folio-designer/src/DataPanel.tsx:15`
- **Observation:** The description formats every `count` as `<n> items`, including object nodes, so an object is announced as, for example, `object · 3 items` rather than three properties/fields. The collection-specific wording leaks onto a different node kind.
- **Impact:** The discovery tree gives a small but repeated false description of data shape, including to screen-reader users.
- [x] **Resolution — FIXED:** Tree descriptions now say `properties` for object nodes and `items` for collections, including zero counts. The keyboard/panel fixture covers populated objects and an empty collection; parser coverage includes an empty object.
- **Related:** AC2.

## Review Summary

- **Reviewed by:** fresh isolated BMAD code reviewer
- **Date:** 2026-08-28
- **Recommendation:** **Accepted after finisher fixes**
- **Story status:** `done`
- **Findings:** 8 fixed (1 Blocker, 5 Majors, 2 Minors); 0 dismissed, 0 deferred, 0 decision-needed.
- **Designer gates:** `npm test` passed **26 files / 133 tests**; `npm run typecheck`, `npm run build`, ordinary offline verification, red-proof verification, wasm-witness verification, `npm run test:e2e:compile`, and `git diff --check` passed after the fixes. Lint's only output is four established Fast Refresh warnings in pre-existing component-export files.
- **Go/static gates:** `folio-go` produced **1,226 passed, 2 known P6g corpus-floor failures, 4 skipped**; vet, native build, and js/wasm build passed. `lint` passed **117 tests** plus vet/build; `hashmatrix` passed **3 tests** plus vet/build. Generated binaries/test results were removed.
- **Explicitly unrun under unchanged D-000.4 cadence:** compiled Playwright execution (`npm run test:e2e`) and the real four-target hash matrix. This story is neither integration-shaped nor hash-shaped, so both remain deferred to the Epic 6 boundary.
- **Cached context audit:** `_bmad-output/implementation-artifacts/epic-6-context.md` is a legitimate required cache produced/consumed by `bmad-build`'s epic-context step, is newer than every planning artifact, follows the required heading/section shape, and accurately distills Epic 6 without story-level implementation drift. It should remain in the scoped Story 6.1 delivery.
- **Scope/isolation audit:** production code does not write sample bytes/name/tree into `.folio`, commands, history, browser storage, telemetry, or network calls, and it exposes no Story 6.2 binding action or later table/parameter UX. The accepted sample remains exact raw Preview input; discovery is bounded and local only.
