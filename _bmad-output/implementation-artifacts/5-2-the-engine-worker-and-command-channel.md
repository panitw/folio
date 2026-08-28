---
baseline_commit: 05c8c70
story_key: 5-2-the-engine-worker-and-command-channel
status: done
created: 2026-08-28
---

# Story 5.2: The engine worker and command channel

**Epic:** 5 — A template author can lay out a report and see the real PDF
**Story key:** `5-2-the-engine-worker-and-command-channel`
**Status:** `done`
**Covers:** **AD-15, AD-16** · the canonical document boundary for all remaining designer stories

**Standing delivery decisions:** numeric order; terminal decision channel; continuous run; unit/lint/build on every story; browser e2e and the cross-target matrix at Epic 5 close under **D-000.4**. This is not a hash-shaped-story override. The Delivery Log must name deferred suites explicitly and must never describe a compile-only check as an executed browser test.

## In plain terms

The designer now has a visual shell, but it must never become a second document editor that merely happens to export a `.folio` file. This story puts the real Go engine in one dedicated Web Worker and makes that worker the sole owner of the document. The React UI receives read-only snapshots to paint, and it asks the worker to do every durable operation through one asynchronous command channel. There is one engine, one document, one serializer, and no competing TypeScript representation of Folio's schema.

That boundary is deliberately small but load-bearing. A future drag may be previewed in React, and an in-progress property field may remain local, but neither can change the document until it becomes a committed command. Later file saving therefore obtains bytes from this same engine; later preview renders those serialized bytes through this same engine; later canvas work asks this same engine for measurement. This story does not build any of those user-facing capabilities early.

The implementation must be honest about its lifecycle. Startup creates exactly one Worker and exactly one wasm instance inside it, serializes all requests through one correlated request/response protocol, and disposes/rejects outstanding requests predictably when the worker terminates or initialization fails. It must not instantiate wasm in React, in a helper imported by React, or once per call. The UI must remain responsive: every call is a Promise/message round trip, never a synchronous wasm call on the main thread.

## Story

**As a** template author,
**I want** the application to hold exactly one copy of my document,
**So that** what I see, what I save and what renders can never disagree.

## Acceptance criteria

### AC1 — exactly one wasm engine runs in exactly one dedicated Worker

**Given** the designer at runtime
**When** the engine is instantiated
**Then** exactly one wasm module instance exists, in one dedicated Worker
**And** no React component, main-thread helper, preview helper, or test convenience path instantiates another engine
**And** initialization is idempotent at the application boundary: repeated consumers share the same client/initialization promise rather than creating another Worker
**And** worker initialization, failures, and termination are observable through typed protocol outcomes rather than uncaught `message` events or permanently pending Promises.

**Proof / red proof.** Unit-test the client factory/lifecycle with a Worker and wasm-host seam: two callers obtain one worker/one initialization; a second constructor/import path is rejected by a structural ownership test; an initialization failure and a terminated worker reject all relevant requests once with useful errors. Red-prove the singleton assertion by temporarily creating a second worker/client and prove the test fails. A test double is allowed only at the Worker boundary; it must not reimplement parsing, serialization, or document mutation in TypeScript.

### AC2 — the Go/wasm engine is the only `.folio` document authority

**Given** a template is loaded, initialized, mutated, validated, or serialized
**When** ownership is inspected
**Then** the wasm engine parses, holds, mutates, validates, and serializes it
**And** no TypeScript model of a `.folio` document exists anywhere in `folio-designer/`
**And** the worker returns only a deliberately limited, immutable UI snapshot and serialized byte results—not a mutable schema mirror or an internal Go document handle
**And** malformed input and engine diagnostics cross the protocol as data with stable codes/messages, not as a TypeScript attempt to reproduce Go validation.

**Proof / red proof.** Add a real js/wasm host test that loads canonical fixture bytes through the worker and proves its returned serialization is the engine's canonical serialization. Add a repository structural test that scans production designer TypeScript for prohibited `.folio` schema declarations/field mirrors while allowing the explicitly named command-envelope, response, snapshot, and transient-UI types. The scan must have a nonzero relevant-file witness and must inspect nested production source, not only a hand-maintained file list. Red-prove it by introducing a representative forbidden document interface/type alias and by bypassing engine serialization with a UI-owned byte/model path.

### AC3 — the UI paints immutable snapshots and commits only commands on one channel

**Given** the user interface
**When** it paints
**Then** it holds an immutable snapshot from the worker and sends every committed mutation as a command over one channel
**And** snapshots are replaced, never mutated in place by the UI
**And** each request has a monotonically allocated correlation id, exactly one terminal response, and an explicit protocol-version/type discriminator
**And** unknown, malformed, duplicate, or out-of-order protocol messages fail closed without changing document state
**And** command execution is serialized in arrival order inside the sole worker, so a later snapshot cannot describe a mutation that did not succeed.

**Proof / red proof.** Test command ordering with delayed worker responses; concurrent UI calls must resolve to snapshots reflecting the engine's serialized command order, not response-arrival order. Test that a response cannot settle a different request, a duplicate response cannot settle twice, and an unknown protocol version/kind yields a rejected request with no state transition. Test `Object.isFrozen`/equivalent observable immutability at the client boundary and verify the UI replaces state on a new snapshot. Red-prove an in-place snapshot write, a duplicate response, and a command sent outside the sole client channel.

### AC4 — transient interaction state remains transient and all engine calls remain asynchronous

**Given** transient interaction state — a drag in flight, a resize preview, or an uncommitted property keystroke
**When** it occurs
**Then** it lives only in UI state and never enters the document
**And** only an explicit commit produces a command and a new engine snapshot
**And** every engine call is asynchronous over the request/response channel; no UI render or event handler blocks on a wasm call
**And** the protocol supports cancellation/abandonment of a caller without corrupting worker state; it does not invent engine-side rollback for a command already committed.

**Proof / red proof.** Use a component/client integration test with a controlled worker response: changing a draft/transient value produces no outgoing command and leaves the latest engine snapshot byte-identical; committing it produces exactly one command and replaces the snapshot only after the response. Assert the UI event loop can process a second interaction while an engine request remains pending. Red-prove by routing draft state into a command or mutating the displayed snapshot before a successful response.

## Required protocol and lifecycle invariants

1. **One owner, not one cache.** The worker owns the live engine document. React state is a read-only projection suitable for painting, not an editable document cache. Do not copy Go's `Document`, `Template`, element, band, style, or asset schema into TypeScript.
2. **One channel.** All engine operations—including initial document setup/load, snapshot retrieval, committed mutation, validation, serialization, future measure, and future render—use one versioned request/response envelope and one client. Do not introduce direct `postMessage` calls outside the client or a special preview/file channel.
3. **Opaque command payloads.** The TypeScript side may model protocol metadata and named UI intents, but document-field semantics, validation, id allocation, canonicalization, and mutation interpretation belong to Go/wasm. Keep any command payload opaque/engine-defined at the transport boundary; do not encode a second `.folio` JSON schema in UI types.
4. **Serialized worker execution.** Process requests deterministically in FIFO arrival order. The client must correlate replies by request id and expose only the successful resulting snapshot; stale or duplicate replies cannot overwrite newer UI state.
5. **Immutable snapshots.** Deep-freeze or otherwise make the public snapshot boundary observably non-mutable in development/test. Snapshot identity changes only on a successful engine result. A caller cannot mutate nested snapshot data and thereby change what another consumer paints.
6. **Lifecycle is finite and fail-closed.** Define `starting → ready → failed | terminated`; reject requests before ready, after failure, and after termination. Termination rejects each outstanding request and removes listeners. Never silently recreate a worker after a failure: recovery is an explicit application-level action and must still preserve the singleton rule.
7. **No document mutation from transients.** Drag geometry, resize previews, focus, selection, hover, input drafts, and an abandoned request stay in UI state. Only the later explicit commit/blur/Enter semantics specified by Stories 5.7–5.8 issue a durable command. Undo/redo will be engine-side history over those committed commands; sample-data loading is not a command and is not undoable.
8. **Bytes remain authoritative.** Serialization is emitted by the engine. Reserve the render API for Story 5.10, where it must receive serialized `.folio` bytes from the same serialization path plus data and parameters—never a live document pointer or UI snapshot.
9. **No hidden I/O.** This worker boundary must not add a network client, account state, server endpoint, browser storage, service worker, file API, or host-font lookup. Story 5.3 owns offline asset caching; Story 5.5 owns file access.

## Tasks / subtasks

- [x] **1. Establish the Go js/wasm host as the sole document engine** (AC: 1, 2)
  - [x] Added the build-constrained js/wasm host and a public opaque-template serializer; the pure core and canonical rendering path remain untouched.
  - [x] Added engine-owned initialize/load/snapshot/validate/serialize/opaque-command operations; unsupported commands fail closed rather than pre-implementing later editor semantics.
  - [x] Converted host failures to coded protocol results and keep the worker-owned wasm callback inside its disposable worker realm.
- [x] **2. Build the dedicated worker and singleton client seam** (AC: 1, 3, 4)
  - [x] Added one discoverable Worker entry and one composition-root singleton client; React shell components do not construct workers.
  - [x] Added a versioned, discriminated envelope with monotonic request ids, explicit lifecycle outcomes, and byte-only opaque payloads.
  - [x] Added FIFO worker execution, idempotent singleton initialization, pending cleanup, termination/error handling, and fail-closed duplicate/unknown response handling.
- [x] **3. Expose immutable, UI-safe snapshots without recreating the document schema** (AC: 2, 3)
  - [x] Added the minimal Go-derived `documentState`/revision/byte-length projection, with no TypeScript document fields.
  - [x] The client copies bytes, deep-freezes snapshots/results, and settles only the correlated successful request.
  - [x] Kept the existing shell honest and inert; the narrow transient controller is test-only boundary support, not a premature editor UI.
- [x] **4. Prove ownership, ordering, transient isolation, and failure behavior** (AC: 1–4)
  - [x] Added Go engine and js/wasm-host compile tests, designer client/transient tests, and a deferred browser integration package.
  - [x] Added non-vacuous structural guards for one Worker construction, one wasm instantiation, no main-thread channel bypass, and no `.folio` schema mirror.
  - [x] Captured red-proof targets for a second worker, schema mirror, duplicate/unknown response, snapshot mutation, and draft-to-command leak.
- [x] **5. Verify and record the correct cadence** (AC: 1–4)
  - [x] Ran the required focused Go/designer unit, type, lint, build, vet, and js/wasm compile gates; the pre-existing full Go-suite red is named below.
  - [x] Compiled the worker browser/integration/e2e suite without executing Playwright.
  - [x] Deferred the four-target hash matrix to the Epic 5 boundary; no rendering algorithm changed and no D-000.4 override is warranted.

## Implementation guidance and boundaries

### Existing seams to reuse

- `folio-go/folio.go` exposes the opaque `folio.Template` boundary and `ParseTemplate`; it deliberately prevents callers from constructing or inspecting `internal/template.Document` field-by-field. Preserve that opacity.
- `folio-go/render_entry.go`, `folio-go/render.go`, `folio-go/fontset.go`, and `folio-go/fonts/` embody the existing engine/font boundary. The wasm adapter must call the public engine path rather than translating the document into a JavaScript model.
- `folio-designer/src/App.tsx` is a deliberate shell with disabled, honestly labelled later features. Thread the engine client through a narrow composition seam; do not retrofit fake file, page, palette, property, preview, or canvas behavior.
- `folio-designer` already has Vitest, strict TypeScript, oxlint, Vite, Playwright, and `test:e2e:compile`. Extend those local conventions rather than introducing a second frontend toolchain.

### Security, determinism, and worker safety

- Treat every worker message and incoming `.folio` byte payload as untrusted. Validate envelope shape/version before dispatch, size-bound payloads before copying/decoding, reject unknown operations, and do not use dynamic evaluation, arbitrary method dispatch, prototype-sensitive object merging, or exception text as a protocol type.
- Preserve AD-1/AD-9: canonicalization, validation, ids, integer geometry, exact decimals, sorting, and serialization stay in Go. A worker adapter must not read clock, locale, environment, filesystem, network, host fonts, or browser layout metrics to alter engine output.
- Do not transfer/mutate a buffer that the UI still treats as its authoritative snapshot. Define ownership for `ArrayBuffer`/`Uint8Array` at each crossing and copy where needed to prevent detached or aliased client state. Never log template/data/parameter bytes in production diagnostics.
- Use a single explicit protocol version. A mismatch fails closed; it must not fall back to loosely typed JSON or silently ignore fields.
- The worker may be asynchronous without becoming nondeterministic: operation ordering comes from its queue, not promise race timing or map iteration. Tests must control delay ordering and prove state follows command order.

### Explicit non-goals and forward seams

- No service worker/cache headers/first-run screen (Stories 5.3–5.4); no open/save or unsaved indicator implementation (5.5); no page setup/grid/canvas/palette/property editing (5.6–5.8); no browser text measurement (5.9); no PDF preview/pdf.js/staleness (5.10–5.11); and no diagnostics presentation/keyboard feature completion (5.12).
- Do not add a second render-only wasm instance. The architecture records it only as a future measured trade-off because duplicating the CJK font set is materially expensive; it is not an optimization available to this story.
- The UI snapshot must leave room for later selection/transient state without absorbing it into document state. Selection, scroll, hover, draft values, and drag previews are UI state; future committed edits enter via this same command channel.
- No new deferred-work entry is due: DW-2 was discharged by Story 5.1; DW-20 is mechanically owned by its existing Go test before the later release tag, not by Story 5.2. Continue to preserve all other open deferred work and its named owners.

## Test strategy and D-000.4 treatment

Every-story evidence is required: focused Go adapter tests; designer unit/component/protocol tests; `npm run typecheck`, `npm run lint`, `npm test`, `npm run build`; appropriate Go `test`, `vet`, lint, and js/wasm compile commands; and red proofs listed in AC1–AC4. The focused worker browser/integration/e2e package must compile on this story.

The browser e2e suite and the four-target determinism matrix are **deferred, not waived**: run both at the Epic 5 boundary after Story 5.12 and before `epic-5` becomes `done`. This story is not one of D-000.4's hash-shaped overrides because it must not modify canonical rendering. If implementation reveals a real worker-runtime incompatibility that cannot be proven by compilation and unit seams, run the smallest relevant browser integration/e2e test immediately, state the D-000.4 override rationale, and record the actual command/result. Do not run or report the full cross-target matrix without such a rendering-related reason.

For every deferred suite, the Delivery Log must say **written/compiled, not run — D-000.4 Epic 5 boundary cadence**. Report skipped tests by name, never only by count (D-000.81).

## Project structure notes

- New wasm host/entry code belongs under the existing `folio-go/` module in a clearly build-constrained `wasm/`-oriented package/command; it must not leak `syscall/js` into the pure core.
- New Worker/client/protocol/snapshot code belongs under `folio-designer/src/`, with browser/integration tests under the existing designer test configuration. Keep worker construction discoverable so the singleton structural test can scan it.
- Generated wasm artifacts belong in an explicit build output/public asset path, never hand-edited and never duplicated under source control as a second engine copy.
- Do not alter `_bmad/_config/files-manifest.csv`, `_bmad/_config/manifest.yaml`, `_bmad/bmm/config.yaml`, `_bmad/core/config.yaml`, or `.agents/`; all were pre-existing unrelated working-tree changes.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5, Story 5.2 and adjacent ownership boundaries.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-1, AD-8–10, AD-14–20, especially AD-15/AD-16.
- `_bmad-output/specs/spec-folio/SPEC.md`, `acceptance.md`, and `folio-format.md` — canonical `.folio` ownership, serialization, validation, and deterministic artifact contract.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — two-user handoff, local/offline product constraints, FR8/FR12/FR35/NFR5.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` and `DESIGN.md` — desktop shell, transient interaction semantics, later workspace/preview states.
- `_bmad-output/implementation-artifacts/5-1-the-design-system-and-application-shell.md` and commit `05c8c70` — shell baseline and its explicit 5.2 boundary handoff.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.4/D-000.81 and Epic 5 canonical-worker direction.
- `_bmad-output/implementation-artifacts/deferred-work.md` — active owner audit; no item is newly due in Story 5.2.

## Delivery Log

### Story creation — 2026-08-28

- Created from the complete Epic 5 requirements, Story 5.1 completed baseline/commit `05c8c70`, architecture, SPEC/acceptance/format, PRD, UX experience/design sources, sprint tracker, decision log, deferred-work register, and current repository seams.
- Confirmed the Story 5.1 handoff: preserve the strict Vite/React shell and its honest deferred controls; do not introduce a TypeScript `.folio` model, fake canvas/file/preview behavior, or a second frontend toolchain.
- Owner/forcing-function audit: no deferred-work item is owned by Story 5.2 or becomes due here. DW-2 is discharged; DW-20 remains tied to its existing mechanical Go-test trigger before the later `folio-go/v0.1.0` release event.
- D-000.4 recorded: per-story unit/lint/build and wasm/e2e compilation are required; browser e2e and the cross-target matrix are deferred to the Epic 5 boundary unless implementation establishes a documented, worker-runtime-specific override.
- No implementation, test execution, commit, push, or deferred-work mutation performed during creation.

### Implementation and verification — 2026-08-28

- Initial implementation evidence is superseded by the finisher evidence below. It was not runtime proof: the original production build lacked wasm assets, UI was not wired to snapshots, and the original red proofs missed realistic violations.

### Adversarial code review — 2026-08-28

- Fresh review against baseline `05c8c70`, the complete Story 5.2 implementation, Epic 5/architecture/SPEC/PRD/UX sources, Story 5.1, D-000.4/D-000.81, project invariants, and the actual test paths found **5 Blockers, 6 Majors, 0 Minors**. All eleven are open `PATCH` items for the finisher; the reviewer made no implementation fixes.
- Measured gates: designer Vitest **17/17 passed**; typecheck, Oxlint, production build, and Playwright TypeScript compile passed. Focused `folio-go` **569 passed** plus vet/build and js/wasm build/test compilation. Full `folio-go` reported **1198 passed, 4 skipped**, with only the sanctioned P6g child and parent failures (`7 < 20`). Full `lint` reported **116 passed, 1 failed**: `TestManifestUpToDate`.
- Exact Story 5.1 attribution: regenerating `lint/MANIFEST.md` from the committed `05c8c70` tree adds exactly the three Story 5.1 designer font rows (`NotoSans`, `NotoSansSC`, `NotoSansThai`). Therefore HEAD itself is stale and Story 5.1's recorded `lint` **117 passed** result cannot have been measured from its committed final tree. This is an inherited-integrity blocker, not an unrelated waiver; the Story 5.2 finisher must repair it before `done`.
- Review mutations were restored: a realistic `DesignerFileState` `.folio` mirror and a second `WebAssembly.instantiate` production path both left all four ownership tests green; bypassing `folio.SerializeTemplate` inside the worker engine left both wasm-engine tests green; and a late response to an abandoned request drove the client from `ready` to `failed` exactly as the source predicts.
- Browser Playwright execution and the four-target matrix remain intentionally unrun under D-000.4. The browser test source was reviewed for executability/teeth and is itself a finding: it never observes the worker or wasm, while the production build contains neither required wasm runtime asset.

## File List

- `_bmad-output/implementation-artifacts/5-2-the-engine-worker-and-command-channel.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `folio-go/serialize_template.go`
- `folio-go/serialize_template_test.go`
- `folio-go/wasm/engine.go`
- `folio-go/wasm/engine_test.go`
- `folio-go/wasm/cmd/engine/main.go`
- `folio-go/wasm/cmd/engine/main_test.go`
- `folio-designer/package.json`
- `folio-designer/scripts/build-wasm.mjs`
- `folio-designer/public/templates/starter.folio`
- `folio-designer/src/App.tsx`
- `folio-designer/src/main.tsx`
- `folio-designer/src/engine-protocol.ts`
- `folio-designer/src/engine-client.ts`
- `folio-designer/src/engine.worker.ts`
- `folio-designer/src/engine-worker-admission.ts`
- `folio-designer/src/engine-worker-queue.ts`
- `folio-designer/src/transient-interaction.ts`
- `folio-designer/src/engine-client.test.ts`
- `folio-designer/src/transient-interaction.test.ts`
- `folio-designer/src/engine-ownership-contract.test.ts`
- `folio-designer/src/engine-worker-queue.test.ts`
- `folio-designer/src/engine-worker-admission.test.ts`
- `folio-designer/e2e/engine-worker.spec.ts`
- `lint/MANIFEST.md` *(inherited Story 5.1 manifest repair)*

## Completion status

Finisher validation complete; all review findings resolved.
**Status:** `done`

## QA Review

### Review Summary

- Outcome: **resolved by finisher**; status advanced to `done` after the validation record below.
- Findings: **5 Blockers, 6 Majors, 0 Minors**.
- Original disposition: **11 PATCH, 0 DECISION, 0 DEFER, 0 DISMISS**. Final disposition: **11 FIX, 0 DECISION, 0 DEFER, 0 DISMISS**.
- Reviewer scope: complete uncommitted Story 5.2 implementation against baseline `05c8c70`; governing Story 5.2/Epic 5/architecture/SPEC/PRD/UX contracts; Story 5.1; D-000.4/D-000.81; project invariants; source, tests, runtime assets, and delivery records. Pre-existing `_bmad` configuration churn, `.agents/`, and `_bmad-output/planning-artifacts/research/` were excluded and preserved.
- File List audit: the Story 5.2 implementation paths are present and no Story 5.2 source/test path was found missing from the list. The material record defect is in the Delivery Log's runtime/proof claims, captured below.
- Finisher gate: resolve or explicitly dismiss every item, repair the inherited Story 5.1 manifest red, rerun all ordinary unit/lint/build/js-wasm gates and the named red proofs, retain the explicit D-000.4 browser/matrix deferral, refresh the Delivery Log, and only then consider `done`.

### Review Findings

- [x] [Review][Fix][Blocker] **The production designer cannot start the wasm engine because neither required runtime asset is built or shipped.**
  - **Category:** AC1/AC2 / AD-15/AD-16 / runtime reachability / build integrity
  - **Location:** `folio-designer/src/engine.worker.ts:20-31`; `folio-designer/package.json:9-17`; `folio-designer/e2e/engine-worker.spec.ts:3-8`
  - **Observation:** the worker unconditionally loads `/wasm/wasm_exec.js` and fetches `/wasm/folio-engine.wasm`, but `folio-designer/public/` contains only fonts, the npm build has no Go/wasm build-or-copy step, and the measured production bundle contains only the JS worker, CSS, HTML, and app JS. The e2e test checks only that the Story 5.1 document bar is visible, so the guaranteed worker boot failure is invisible.
  - **Impact:** every real application load reaches `WASM_INITIALIZATION_FAILED`; there is zero wasm instance and zero Go-owned document, so the central Story 5.2 boundary does not exist at runtime despite compilation passing.
  - **Resolution:** **FIXED.** Reproducible `build:wasm`, served-URL proof, and the focused real-worker e2e source now cover the asset/runtime path.

- [x] [Review][Fix][Blocker] **No production UI path initializes, paints, or successfully commits against an engine snapshot.**
  - **Category:** AC2/AC3/AC4 / AD-15 / functional completeness
  - **Location:** `folio-designer/src/main.tsx:4-8`; `folio-designer/src/App.tsx:13-27`; `folio-designer/src/transient-interaction.ts:5-15`; `folio-go/wasm/engine.go:76-80`
  - **Observation:** `main.tsx` merely constructs the client and discards the result; `App` never receives a client or snapshot and continues to render the static “NO TEMPLATE LOADED” shell. `TransientInteraction` is not used by production UI, while its `commit()` sends arbitrary draft bytes to the only real engine command implementation, which unconditionally rejects every command.
  - **Impact:** the acceptance path “immutable snapshot paints UI; explicit commit produces exactly one command and replaces the snapshot after success” exists only in a fake-worker test. Even after wasm assets are supplied, no real user/runtime path exercises document ownership or a successful committed command.
  - **Resolution:** **FIXED.** The production composition/UI path loads, serializes, paints, and commits through the one Go owner.

- [x] [Review][Fix][Blocker] **Malformed, mismatched, and duplicate worker requests do not fail closed and can hang or mutate twice.**
  - **Category:** AC1/AC3 / protocol security / correlation / exactly-once execution
  - **Location:** `folio-designer/src/engine.worker.ts:36-43`; `folio-designer/src/engine.worker.ts:45-67`; `folio-go/wasm/engine.go:33-45`
  - **Observation:** the worker silently returns for a non-object, wrong `kind`, wrong protocol version, or missing/non-string id, and it casts the remaining object without validating `operation`, payload type, or a bounded id. It keeps no seen/in-flight/terminal id set. Posting the same valid `load` envelope twice therefore queues and executes it twice, incrementing the engine revision twice; malformed requests receive no terminal response at all.
  - **Impact:** a buggy or hostile message can leave a Promise pending forever, and a duplicate committed command can change the sole document more than once. This directly violates the typed-outcome, one-terminal-response, duplicate-message, and no-state-transition clauses.
  - **Resolution:** **FIXED.** Full bounded request validation, correlated typed errors, lifecycle failure, and seen-id at-most-once admission are covered.

- [x] [Review][Fix][Blocker] **Abandoning a caller predictably terminates the sole engine when its already-committed response arrives.**
  - **Category:** AC4 / lifecycle / cancellation semantics
  - **Location:** `folio-designer/src/engine-client.ts:31-47`; `folio-designer/src/engine-client.ts:68-82`; `folio-designer/src/engine-client.test.ts:70-79`
  - **Observation:** abort deletes the request from `#pending`, but the worker is intentionally not cancelled. When its legitimate terminal response later arrives, `#settle` classifies the now-untracked id as unknown/duplicate and calls `#fail`, terminating the only worker. The shipped test stops immediately after the abort and never emits the required late response. A review mutation that did emit it observed `client.state === 'failed'`, not `ready`.
  - **Impact:** the advertised abandonment mechanism corrupts the application lifecycle after any in-flight command is abandoned, contradicting the explicit rule that caller cancellation must not corrupt worker state or invent rollback.
  - **Resolution:** **FIXED.** Bounded tombstones consume the one late response and the new red proof retains unknown/duplicate failure behavior.

- [x] [Review][Fix][Blocker] **The committed Story 5.1 tree leaves the mandatory lint gate red, so Story 5.2 cannot finish on a truthful green baseline.**
  - **Category:** inherited integrity / every-story lint gate / AD-26 manifest truthfulness
  - **Location:** `lint/MANIFEST.md:269-282`; `_bmad-output/implementation-artifacts/5-1-the-design-system-and-application-shell.md:244`; `_bmad-output/implementation-artifacts/5-2-the-engine-worker-and-command-channel.md:178-180`
  - **Observation:** `go test -count=1 ./...` in `lint` reports 116 pass / 1 fail at `TestManifestUpToDate`. Regenerating from an isolated exact `05c8c70` Story 5.1 tree adds precisely three rows for `folio-designer/public/fonts/{notosans,notosanssc,notosansthai}`. The same commit includes those font binaries but omits those rows while its story claims lint 117 passed.
  - **Impact:** HEAD is objectively stale; the Story 5.1 pass claim cannot describe its committed final tree, most plausibly because generation/staging happened before the final font set. Treating it as “unrelated” would allow a second story to finish with a known mandatory gate failure and would make the pipeline's integrity record cumulative fiction.
  - **Resolution:** **FIXED.** The three missing Story 5.1 font rows are committed in this scoped repair; final lint is 117 passing.

- [x] [Review][Fix][Major] **The singleton promise represents construction, not initialization, and lifecycle cleanup is incomplete.**
  - **Category:** AC1 / lifecycle / singleton correctness
  - **Location:** `folio-designer/src/engine-client.ts:23-27`; `folio-designer/src/engine-client.ts:50-55`; `folio-designer/src/engine-client.ts:78-100`; `folio-designer/src/main.tsx:8`
  - **Observation:** `createEngineClientSingleton` resolves immediately to a client still in `starting`; a later worker `failed` lifecycle does not reject that shared promise, so `main.tsx`'s `.catch` cannot observe initialization failure. If `createWorker()` throws synchronously, the factory throws before any Promise exists. Both termination paths leave `onmessage`/`onerror` installed despite the story's explicit listener-removal invariant.
  - **Impact:** consumers cannot await one idempotent ready/failed initialization outcome, startup failure is hidden, and retained handlers permit late events/references after the finite lifecycle should be closed.
  - **Resolution:** **FIXED.** Ready/failed singleton semantics and full listener cleanup are unit-tested.

- [x] [Review][Fix][Major] **The structural ownership guards are name/API regexes that accept realistic second authorities.**
  - **Category:** AC1/AC2 / test discriminating power / no-TypeScript-schema invariant
  - **Location:** `folio-designer/src/engine-ownership-contract.test.ts:6-35`
  - **Observation:** the schema guard rejects only a short list of exact declaration names, not document-field mirrors. A temporary production `DesignerFileState` containing version/page/bands/elements/assets passed all four tests. The wasm guard scans only `engine.worker.ts` for `WebAssembly.instantiateStreaming`; a second production `WebAssembly.instantiate(...)` path in another file also passed. The Worker count likewise recognizes only literal `new Worker(` spelling.
  - **Impact:** the load-bearing “one Go authority / one wasm instance” proof stays green under the realistic violations it exists to prevent; the developer's representative red proofs are too syntactic to protect the architecture.
  - **Resolution:** **FIXED.** AST-wide semantic guards retain the realistic schema and second-wasm mutation controls.

- [x] [Review][Fix][Major] **Canonical serialization and the real js/wasm Worker path are not independently proven.**
  - **Category:** AC2 / AD-9 / round-trip semantics / test vacuity / browser reachability
  - **Location:** `folio-go/wasm/engine_test.go:9-33`; `folio-go/wasm/cmd/engine/main_test.go:14-37`; `folio-designer/e2e/engine-worker.spec.ts:3-8`
  - **Observation:** both Go tests feed an already-canonical fixture, so caching the caller bytes instead of calling `folio.SerializeTemplate` produces the same expected bytes. A review mutation making exactly that bypass left both wasm-engine tests green. The js/wasm test is compiled but directly calls Go `dispatch`; it never runs `wasm_exec`, a browser Worker, transfer semantics, fetch paths, or callback registration. The Playwright source only checks Story 5.1 shell visibility.
  - **Impact:** a serializer bypass, host-registration incompatibility, detached-buffer error, or unreachable wasm asset can ship while every claimed Story 5.2 proof passes/compiles. D-000.4 permits deferring execution, not substituting an unrelated smoke test or a vacuous canonical-in/canonical-out assertion.
  - **Resolution:** **FIXED.** Non-canonical native/js-wasm tests and the focused browser source prove the intended route; browser execution remains environment-unavailable.

- [x] [Review][Fix][Major] **Engine diagnostics lose their stable Go codes and reflect raw parser text through an unbounded generic error channel.**
  - **Category:** AC2 / AD-14 / security / error sanitization
  - **Location:** `folio-go/wasm/cmd/engine/main.go:65-99`; `folio-go/render_error.go:40-53`; `folio-go/internal/template/parse.go:70-97`; `folio-designer/src/engine-client.ts:72`
  - **Observation:** `folio.ParseTemplate` already returns `*folio.RenderError` with a stable `Diagnostic.Code`, element id/path metadata, and a message. The wasm adapter discards that structure, labels all engine failures `ENGINE_REJECTED`, and forwards `err.Error()` verbatim. Many parser errors deliberately include the offending raw JSON value, and the client forwards the message unchanged.
  - **Impact:** the UI cannot match the stable diagnostic registry required by AD-14, while attacker-controlled template content can be reflected into protocol/UI errors at large size. Codes/messages are neither stable nor meaningfully sanitized as the story promises.
  - **Resolution:** **FIXED.** Stable diagnostic codes/location fields cross with bounded safe text; hostile parser content has a dedicated test.

- [x] [Review][Fix][Major] **The client accepts out-of-order success settlement and the ordering test masks it with `Promise.all`.**
  - **Category:** AC3 / correlation / FIFO / stale-response safety
  - **Location:** `folio-designer/src/engine-client.ts:68-75`; `folio-designer/src/engine-client.test.ts:27-44`
  - **Observation:** the test deliberately emits response 2 before response 1, but asserts only the final `[first, second]` values after `Promise.all`; it never measures which Promise settled first or which snapshot a UI setter would apply last. The client immediately resolves whichever id arrives. A consumer attaching `setSnapshot` to both calls can therefore paint revision 2 and then be overwritten by stale revision 1 if messages reorder or a worker lifecycle changes.
  - **Impact:** the exact stale/reordered-response failure named by AC3 is accepted by the proof suite. Today's worker loop is FIFO, but the client boundary neither enforces nor detects monotonic snapshot application and offers no protection against future async/restart changes.
  - **Resolution:** **FIXED.** Controlled queue and client out-of-order proofs enforce FIFO settlement.

- [x] [Review][Fix][Major] **The Delivery Log overstates runtime coverage and red-proof execution.**
  - **Category:** auditability / Delivery Log truthfulness / finisher handoff
  - **Location:** `_bmad-output/implementation-artifacts/5-2-the-engine-worker-and-command-channel.md:175-180`
  - **Observation:** the record says the composition root initializes the singleton, lifecycle cleanup is complete, red proofs execute, a focused worker browser/integration package exists, and no worker-runtime incompatibility was found. The source and review mutations disprove those statements: construction is not readiness, listeners remain, several representative violations pass, the e2e never observes a worker, and the built app cannot fetch either wasm runtime asset.
  - **Impact:** the finisher and Epic 5 boundary owner would trust compile-only or unrelated evidence as runtime coverage, obscuring why the mandatory catch-up suite cannot currently exercise Story 5.2.
  - **Resolution:** **FIXED.** The finisher validation section is the authoritative evidence and names exact executed, standing-red, skipped, and unavailable gates.

### Finisher resolution and validation — 2026-08-28

All **11/11** findings are **FIX** (5 Blockers, 6 Majors); none is waived, deferred, or dismissed. Routine decisions logged here: generated runtime assets remain ignored build output rather than a second checked-in wasm copy; the sole supported opaque command is Go-owned `commit`, which advances its engine snapshot without inventing later editor schema; and the browser-focused test is an evidence override because this story’s core boundary is a real Worker/wasm path, while the full browser suite and four-target matrix remain Epic 5 boundary work.

| Finding | Resolution and proof |
| --- | --- |
| B1 runtime assets | `build:wasm` builds `folio-engine.wasm`, copies matching `wasm_exec.js`, and creates a deliberately non-canonical starter asset. Production build output and a served preview returned HTTP 200 for all three URLs. |
| B2 reachable UI | The composition root awaits ready, loads and serializes opaque bytes, then paints the Go snapshot. The reachable button sends the Go-owned opaque commit and replaces React state only after success. |
| B3 malformed/duplicate | Full envelope/id/payload validation returns correlated typed failures, uncorrelated invalid input fails lifecycle, and seen ids prevent a second execution. |
| B4 abandonment | The client retains one bounded tombstone per abandoned in-flight id; its later FIFO response is consumed without changing `ready`. |
| B5 manifest | Regenerated `lint/MANIFEST.md`; exactly the inherited Story 5.1 `NotoSans`, `NotoSansSC`, and `NotoSansThai` rows were added. Final lint is green. |
| M1 singleton lifecycle | The singleton resolves only after `ready` or rejects typed startup failure, catches synchronous construction failure, and clears both listeners on failed/terminated. |
| M2 ownership guards | TypeScript AST guards scan all production source for Worker/SharedWorker construction, both wasm instantiate forms, and realistic three-field document mirrors; both review mutations are direct red controls. |
| M3 canonical/runtime path | Native engine and js/wasm host use whitespace-prefixed non-canonical bytes and require canonical output. The focused browser source loads, serializes, commits, and checks served runtime assets; its attempted execution is recorded below. |
| M4 diagnostics | `RenderError.Diagnostic` code/location crosses as bounded typed data, while template/parser text is replaced by a bounded safe message; the js/wasm host test injects 2 KiB hostile content. |
| M5 FIFO settlement | Worker queue has a controlled-delay FIFO unit proof; client rejects out-of-order terminal responses rather than allowing stale paint. |
| M6 delivery truth | This section supersedes the prior false runtime claims and separates executed native/build evidence from unavailable/deferred browser/matrix work. |

**Executed PASS:** `npm test` — **7 files / 23 tests** (including malformed/version-mismatch and duplicate admission, duplicate/unknown response, late-abandonment response, FIFO queue delay, non-vacuous ownership mutations); `npm run typecheck`; `npm run lint`; `npm run build` (runs actual wasm build/copy); `npm run test:e2e:compile`; production preview URL proof for `/wasm/folio-engine.wasm` (HTTP 200, `application/wasm`, 5,732,992 bytes), `/wasm/wasm_exec.js` (HTTP 200), and `/templates/starter.folio` (HTTP 200); `go test . ./wasm` — **570 passed / 2 packages**; `go vet ./...`; `go build ./...`; `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine`; `GOOS=js GOARCH=wasm go test -c ./wasm/cmd/engine`; `gofmt -l` (no output); `git diff --check`; `lint: go test -count=1 ./...` — **117 passed / 4 packages**.

**Expected standing red (not waived):** `folio-go: go test -count=1 ./...` — **1199 passed, 2 failed, 4 skipped / 18 packages**. Only the sanctioned Epic 2 P6g red remains: `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (7, floor 20) and its parent. It is unrelated to Story 5.2; no new full-Go failure was introduced.

**Skipped/unavailable:** focused `npm run test:e2e -- e2e/engine-worker.spec.ts` was attempted as the smallest worker-runtime integration override, but Playwright has no installed Chromium executable (`chromium_headless_shell` missing); it therefore did not execute a browser and is **environment-unavailable**, not reported as PASS. The test source compiles and is retained for Epic 5. The remaining full Playwright browser suite and four-target hash matrix are **written/compiled or otherwise prepared, not run — D-000.4 Epic 5 boundary cadence**.
