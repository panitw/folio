---
baseline_commit: 3134ac7
story_key: 6-3-supply-parameters-for-preview
status: done
created: 2026-08-28
---

# Story 6.3: Supply parameters for preview

**Epic:** 6 — A template author can bind a report to data and build the golden report
**Covers:** FR21 · AD-16

## In plain terms (read this first if you just want the gist)

Some report values are deliberately not part of the report data. A service may provide the report date, a branch name, or another run-specific value at the moment it produces a statement. This story lets the author provide those values locally before looking at the exact preview. When a template asks for a report date, the author can see that request and enter a value for it, so the preview reflects a realistic statement instead of a blank or guessed date.

These values are a temporary companion to the preview, not an edit to the template. They are not saved in the report file, do not make the report need saving, and cannot be undone or redone. Reopening a report, starting a blank report, or replacing the local companion inputs must not quietly carry old values into a different report. Changing a value does make any earlier preview out of date immediately, because it is a different set of inputs to the production result.

The author is shown only the names the template actually asks for. The app does not guess names by reading template text itself; the local production engine supplies the bounded list. Editing remains a local parameter document supplied alongside the template and sample data to the same exact preview. There is no new preview route, no server connection, and no attempt to turn these temporary values into saved report data. If a requested value is absent, the existing production error remains the source of truth: it identifies both the missing parameter and the report element that asked for it.

## Story

As a template author,
I want to provide the runtime values my developer's service will supply,
so that I can preview a statement whose generated date is real rather than blank.

## Acceptance criteria

### AC1 — The engine discovers the template's referenced parameter names, and the author can edit the transient parameter document

**Given** a loaded canonical template that references one or more `params` values, including `{{params.reportDate}}`,
**When** the author opens Preview inputs,
**Then** the UI receives a bounded, engine-derived projection of the distinct referenced parameter names and presents an accessible local parameter-document editor for those references.

**And** Go/wasm exclusively parses template expressions and determines which `params.<name>` references exist. TypeScript may retain the raw parameter-document bytes/text and paint the supplied reference list, but it must not scan template bytes, infer references from canvas/binding paint data, manufacture a template-expression grammar, or become a second template authority. Discovery excludes data-root paths, row aliases, table/column/footer configuration, and any unreferenced name.

**And** editing a referenced value changes only transient Preview input state. It neither changes canonical `.folio` bytes/revision/history/saved-revision state nor creates an engine document command. The document is a valid bounded JSON object for the existing parameter channel; invalid local draft text is stated honestly and is not sent as a substitute value. An empty object remains a valid parameter document so missing-reference diagnostics can be produced by the engine.

**Proof / red proof.** Prove engine-derived discovery for repeated references and for `reportDate`; a keyboard-operable named editor; valid value replacement; invalid draft containment; and no document revision, dirty state, undo/redo, or save-byte change. Red-prove browser template parsing, a parameter name invented from UI state, sample data being used as parameter discovery, an unreferenced editable field, or a parameter edit reaching the committed command/history route.

### AC2 — Parameter bytes remain a distinct third Preview and identity input, and every accepted change invalidates stale output

**Given** accepted sample-data bytes, a canonical template revision, and an accepted transient parameter document,
**When** Preview identity is checked and the exact PDF is rendered,
**Then** the same raw parameter bytes are passed unchanged as the distinct third render input and as the identity input alongside raw sample bytes; no alternate render path, live template object, browser rendering, or second wasm instance is introduced.

**And** an accepted parameter-document change synchronously cancels/revokes older Preview work, marks existing output stale, and schedules/requires the current input set under the existing FIFO, token, document-generation, revision, and identity admission rules. A late identity/render response for prior parameter bytes cannot become current. The canvas remains usable while only Preview performs work.

**And** parameters remain local, nonpersisted, nonundoable, and nontelemetry state. They do not enter canonical template bytes, save targets, browser persistence, sample data, binding commands, history metadata, network traffic, or the canvas's document projection. Clearing/replacing a template or companion inputs must scope or clear their editor state so values cannot be attributed to a new document.

**Proof / red proof.** Prove byte-for-byte forwarding of a nontrivial valid parameter document to both identity and render; identity changes when only parameters change; stale status appears immediately; and a delayed old render/identity result is rejected. Red-prove two-input rendering, parsed/re-serialized parameter transport, a parameter change leaving a current preview visible, parameter bytes in save output, and a parameter value surviving a document-generation replacement.

### AC3 — Missing referenced parameters keep the production engine's located diagnostic

**Given** a template element referencing `{{params.reportDate}}` and an accepted parameter document that omits `reportDate`,
**When** the author requests Preview,
**Then** the production render fails through the existing AD-14 diagnostic channel, locating the element id and naming the missing `params.reportDate` path; the UI presents that returned fact without inventing a parallel parameter-error taxonomy.

**And** explicit JSON `null`, present compatible values, and wrong-kind values retain their established engine meanings. This story must not add browser coercion, defaulting, schema inference, required-field validation that preempts render, or an automatic value based on the wall clock, locale, time zone, sample data, or environment.

**And** parameter discovery/editing is deliberately not scalar component binding, manual binding authoring, a sample-data binding change, table collection binding, row-scope column binding, footer aggregation, failed-render redesign, a golden-report fixture, or final native/browser byte comparison. Stories 6.4–6.7 retain those responsibilities.

**Proof / red proof.** Cover absent `reportDate` with returned element and path provenance, then a supplied compatible date that reaches a successful exact Preview. Cover explicit null and wrong-kind through their existing renderer outcomes. Red-prove a browser-generated missing-field message, an implicit current date/default, an unlocated error, a parameter alias shadowing `params`, or table/row/column/footer behavior added here.

## Tasks / subtasks

- [x] **1. Add the narrow engine-owned parameter-reference projection** (AC: 1, 3)
  - Reuse the canonical engine expression/template analysis seam to collect distinct bounded `params` reference names from the loaded canonical document; preserve exact namespace semantics and stable ordering.
  - Return only the smallest display/protocol projection needed for the editor, with strict Go/wasm and TypeScript admission, deep-freeze, malformed/oversized/unknown-field rejection, and document-generation/revision correlation. Do not expose a template/schema mirror or parsed expression tree.
  - Keep missing/null/wrong-kind resolution in the existing production renderer and AD-14 diagnostic registry. Do not add a parameter-validation command, new diagnostic code taxonomy, or a TypeScript compatibility checker.

- [x] **2. Replace the provisional generic parameter edit with referenced-parameter document UX** (AC: 1, 3)
  - Build on the existing raw Preview-parameter state and exact byte channel. Present engine-discovered references with named, keyboard-operable controls and visible cyan focus; retain an accessible bounded parameter-document representation where needed to preserve arbitrary valid JSON values exactly.
  - Make draft parsing/acceptance explicit: only an accepted JSON object becomes Preview input; invalid drafts receive local input feedback and do not alter the last accepted bytes. Preserve valid `{}` as the deliberate way to request engine missing-parameter evidence.
  - Treat all parameter editor state as transient and scope it to the active document/companion-input lifecycle. It is never a component property, binding UI state, save artifact, undo entry, or persisted browser setting.

- [x] **3. Thread accepted raw parameter bytes through existing Preview freshness authority** (AC: 2)
  - Reuse the dedicated worker, `identity`, `serialize`, and `render` protocol operations. Pass the same accepted raw parameter bytes separately to identity and render; do not change canonical serialization or add a render/template authority.
  - On accepted parameter changes, invoke the existing invalidation/cancellation/coalescing path and preserve its token, FIFO, document-generation, revision, and identity checks. Ensure replacement/clear transitions revoke pending work and cannot install prior parameter results.
  - Keep Preview blocked only within its surface and preserve existing exact-PDF evidence/freshness semantics.

- [x] **4. Prove parameter authority, transport, lifecycle, and diagnostic boundaries** (AC: 1–3)
  - Add focused Go expression/reference/projection and wasm/protocol tests; designer editor/a11y/draft/lifecycle tests; Preview identity/render byte-forwarding and stale-response tests; and static authority checks preventing browser template parsing or parameter persistence.
  - Add compiled Playwright-source coverage for discovering `reportDate`, supplying it, seeing stale Preview, and receiving the engine-located missing-parameter failure. Keep real Playwright execution deferred unless the implementation proves this story intrinsically integration- or hash-shaped.
  - Run focused unit suites plus `npm run typecheck`, `npm run lint`, `npm run build`, `npm run test:e2e:compile`, relevant Go/wasm tests, `go vet ./...`, native and js/wasm builds, formatting/diff checks, and applicable lint/hashmatrix static gates. Record actual commands/results and red proofs during implementation.

## Developer guardrails

1. **Parameters are inputs, never document edits.** They stay raw, separate Preview/identity input; transient, nonpersisted, nonundoable, non-saveable, and outside committed commands. Do not add them to `.folio`, snapshots as document data, browser persistence, or history.
2. **One engine decides references.** Go/wasm discovers canonical `params` references. TypeScript must not parse template bytes, mine canvas text, define parameter expression grammar, infer requirements/types, or create a template model.
3. **Preserve the settled three-input transport.** Identity and render consume canonical serialized template bytes, raw sample data, and the same raw parameter document. Do not add an alternate render route, parsed/re-serialized transport, a live-document render, or another worker/wasm instance.
4. **AD-14 remains diagnostic authority.** Missing referenced parameters retain the renderer's located parameter path and element id. Null and wrong-kind behavior stays unchanged; no browser default, coercion, current-time value, or second error taxonomy.
5. **Preview freshness is non-negotiable.** An accepted parameter change invalidates older output immediately. Retain cancellation, FIFO correlation, generation/revision/identity guards, and rejection of late responses; Preview alone may be busy.
6. **Constrain the editor to references without inventing schema.** Show names actually referenced by the canonical template, but do not claim that the current sample defines legal types or that the UI can validate runtime compatibility. Values remain JSON input to production render.
7. **Keep Epic 6 scope ordered.** Scalar binding remains Story 6.2. Table matrix, collection/row-scope column binding, and aggregates belong to Stories 6.4–6.5; failure presentation is Story 6.6; authored golden fixture and native/browser byte proof are Story 6.7.
8. **D-000.4 cadence is unchanged.** Unit/lint/build and deferred Playwright-source compile run for this story. Real Playwright and the four-target matrix remain Epic 6 boundary evidence unless a demonstrated intrinsic integration/hash exception is recorded before implementation.
9. **Preserve unrelated work.** Do not modify `_bmad` configuration/manifest churn, `.agents/`, planning research, fixtures/goldens/fonts, or unrelated user changes.

## Project structure notes

- `folio-go/wasm/engine.go` already owns preview identity and render with distinct non-empty raw data/parameter byte channels; `folio-go/wasm/cmd/engine/main.go` preserves those base64 transport fields. Add parameter-reference discovery at the canonical engine seam, not in the browser.
- `folio-designer/src/engine-protocol.ts`, `engine.worker.ts`, `engine-worker-admission.ts`, and `engine-client.ts` already enforce a strict request/response boundary. Any new reference projection must be narrow, versioned through that boundary, and admitted as paint/editor metadata rather than a template model.
- `folio-designer/src/App.tsx` currently holds the provisional raw `previewParams` text and supplies encoded bytes to identity and render. Its existing `invalidatePreview`, scheduler, tokens, generations, and `canInstallPreview` logic are the intended freshness seam; the generic editor is the intended replacement surface.
- `folio-designer/src/preview/freshness.ts` and current App/engine-protocol tests establish the stale-result authority model. Extend them for a parameter-only input change rather than creating a second Preview state machine.
- `folio-go/internal/bind` and existing render diagnostics already resolve the separate unshadowable `params` namespace. Preserve their missing/null/wrong-kind semantics; Story 6.3 is UI/protocol discovery and transient input work, not a renderer or expression-language change.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 6, Story 6.3 ACs, FR21, and Stories 6.4–6.7 boundaries.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-11 (unshadowable `params`), AD-14 (one located diagnostic channel), AD-15 (engine-owned document), AD-16 (one worker and three Preview inputs), and AD-17 (paint-only browser).
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR21, FR35–FR36, FR40–FR43, NFR1/NFR5/NFR8, and answered Q3/Q11.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` and `DESIGN.md` — exact Preview, stale-output, local/offline, focus, and cyan/amber semantics.
- `_bmad-output/implementation-artifacts/epic-6-context.md` — Epic 6 cache, transient companion-input direction, and cross-story dependencies.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — 2026-08-28 Epic 6 lead refresh, D-000.4 cadence, and D-6.2.1 sample-shaped affordance versus runtime authority ruling.
- `_bmad-output/implementation-artifacts/deferred-work.md` — D-000.4 deferred real Playwright/four-target boundary evidence and escalation rule.
- `_bmad-output/implementation-artifacts/6-1-load-sample-data-and-browse-its-paths.md` and `6-2-bind-a-component-by-picking-a-path.md` — completed raw sample, Preview authority, picker-boundary, accessibility, and D-6.2.1 baselines.
- `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go`, `folio-designer/src/App.tsx`, `folio-designer/src/engine-protocol.ts`, `folio-designer/src/engine.worker.ts`, and `folio-designer/src/preview/freshness.ts` — current parameter transport and Preview-admission seams.

## File list

- `folio-go/parameter_references.go`
- `folio-go/parameter_references_test.go`
- `folio-go/wasm/engine.go`
- `folio-go/wasm/engine_test.go`
- `folio-go/wasm/cmd/engine/main.go`
- `folio-designer/src/App.tsx`
- `folio-designer/src/App.test.tsx`
- `folio-designer/src/engine-protocol.ts`
- `folio-designer/src/engine-protocol.test.ts`
- `folio-designer/src/engine-client.ts`
- `folio-designer/src/engine-client.test.ts`
- `folio-designer/src/engine.worker.ts`
- `folio-designer/src/engine-ownership-contract.test.ts`
- `folio-designer/e2e/preview-parameters.spec.ts`

## Delivery log

### Story creation — 2026-08-28

- Created against baseline commit `3134ac7` after reading the Epic 6 source; architecture AD-11/14/15/16/17; PRD and UX sources; Epic 6 context; decision/deferred logs including the 2026-08-28 lead refresh and D-6.2.1; completed Stories 6.1–6.2; sprint tracker; and current Go/wasm/designer source.
- Confirmed no new owner decision is required. The parameter document is already settled as separate raw Preview/identity input: transient, nonpersisted, nonundoable, and outside engine document commands. D-6.2.1 further confirms that transient input shape is not canonical runtime authority; missing referenced parameters retain existing AD-14 renderer diagnostics.
- Confirmed the current implementation already sends raw `previewParams` bytes as the third input to identity and render through one worker, while the generic Preview textarea lacks engine-owned parameter-reference discovery and reference-focused editing. Those are the intended Story 6.3 seams; no new render path, template authority, or wasm instance is warranted.
- D-000.4 cadence remains unchanged: run unit/lint/build and compile deferred e2e source in this story; real Playwright and the four-target matrix remain deferred to the Epic 6 boundary unless implementation demonstrates an intrinsic integration/hash-shaped exception. Record actual commands, results, red proofs, and any genuine exception during development.
- Implementation has not started. No task is pre-checked and no test result, implementation file list, completion claim, commit, or remote operation is recorded here.

### Implementation — 2026-08-28

- Added `parameter-references`, a small revision-correlated Go/wasm projection that walks the already-admitted canonical expression AST, deduplicates and stably sorts direct `params.<name>` references, and deliberately excludes tables, columns, footers, data roots, and browser-derived candidates. The TypeScript closed protocol rejects malformed, unordered, oversized, unknown-field, or revision-mismatched metadata and deep-freezes admitted names.
- Replaced the generic Preview textarea with engine-named keyboard controls plus an accessible raw JSON-object editor. Draft and accepted parameter documents are distinct: malformed/non-object text keeps the accepted raw bytes untouched and states that fact locally; `{}` remains accepted for production AD-14 missing-parameter evidence. All editor state is transient and is cleared on sample/template replacement.
- Preserved the established raw three-input `identity`/`serialize`/`render` flow. Accepted parameter input alone invokes the existing invalidation/scheduling path; no command, document revision, undo/redo state, save output, browser persistence, or template parser was added. Tests cover exact forwarded bytes, stale rejection, named editor/a11y, invalid-draft containment, the protocol boundary, and authority static checks. Added compiled-only Playwright source covering discovered `reportDate`, engine-located absent input, and a stale supplied value.
- Passed: `npm run test` (142 tests), `npm run typecheck`, `npm run test:e2e:compile`, `npm run lint` (only existing Fast Refresh warnings), production `npm run build`, offline Wasm witness `npm run verify:offline:wasm`, offline red proof `npm run verify:offline:red`, focused `go test . ./wasm` (602), `go vet ./...`, native `go build ./...`, JS/Wasm `go build ./wasm/cmd/engine`, `lint` module test/vet/build (117), `hashmatrix` module test/vet/probe build (3), `gofmt -l`, and `git diff --check`.
- Recorded boundary deferrals: real Playwright and the four-target matrix remain Epic 6 evidence per D-000.4. The full `go test ./...` has the pre-existing `internal/text` P6g corpus-floor failure (7 opaque names vs required 20); all affected local Story packages pass. A direct JS/Wasm `go test` compile attempt cannot execute on this host (`exec format error`); JS/Wasm build passed.

### Review Findings

- [x] [Review][Patch][High] **Discover nested parameter roots and common `visibleIf` references instead of silently dropping them.**
  - **Category:** AC1/AC3 · Go-owned reference completeness · nested paths · table-scope boundary
  - **Location:** `folio-go/parameter_references.go:28-40`; `folio-go/parameter_references.go:67-72`; `folio-go/folio_expr_validate.go:43-50`
  - **Observation:** `collectParameterPaths` accepts only paths with exactly two segments, so a valid expression such as `{{params.statement.reportDate}}` contributes no editable parameter even though the production resolver traverses that nested params path. The element loop also skips a table before checking the common element-level `visibleIf`; `visibleIf: "params.showTable"` is therefore omitted even though `visibleIf` is valid on every element kind and is not table column/footer configuration.
  - **Impact:** The engine projection can say there are no editable references while exact production render requires one, preventing the author from satisfying a valid template through the named surface and violating the promise that every referenced parameter is discoverable.
  - **Resolution:** **FIX.** Traverse every production expression site in the same order as canonical validation, process common element `visibleIf` before excluding table-specific configuration, and define nested params paths as a stable top-level parameter-name projection (or another explicitly compatible editor projection). Add direct, nested, repeated, table-`visibleIf`, table-column/footer exclusion, data-root, row-alias, and reserved-token tests.

- [x] [Review][Patch][High] **Enforce the same parameter-name bound in Go that TypeScript admits.**
  - **Category:** AC1 · bounded projection · Go/wasm protocol compatibility
  - **Location:** `folio-go/parameter_references.go:56-64`; `folio-designer/src/engine-protocol.ts:95`
  - **Observation:** Go bounds only the number of distinct names. Its expression grammar accepts an ASCII identifier longer than 128 characters, but the TypeScript protocol rejects every name longer than 128. A valid loaded template can therefore produce a Go success response that the main-thread protocol classifies as invalid.
  - **Impact:** Opening Preview for that template terminates the sole engine client instead of returning a bounded engine error, taking identity/render and the rest of the local document session down with the discovery request.
  - **Resolution:** **FIX.** Make the Go projection enforce the protocol's declared count and per-name byte bounds before wasm transport, keep the constants mechanically aligned, and prove the exact 128/129 boundaries plus a many-distinct-names overflow without emitting an inadmissible success envelope.

- [x] [Review][Patch][High] **Do not parse and re-serialize the whole accepted parameter document for a named edit.**
  - **Category:** AC1/AC2 · raw-byte fidelity · number lexemes · precision
  - **Location:** `folio-designer/src/App.tsx:198-204`; `folio-designer/src/App.tsx:534-539`
  - **Observation:** The named editor runs the accepted document and the entered value through `JSON.parse`, spreads the resulting JavaScript object, and `JSON.stringify`s everything. A document containing `9007199254740993` and `1e+00` becomes `9007199254740992` and `1` when an unrelated `reportDate` field is edited; the entered value itself is normalized the same way.
  - **Impact:** A local convenience edit silently changes untouched runtime values and can corrupt an exact integer before the raw bytes reach the production decoder. That defeats the deliberately precision-preserving params channel and AC2's parsed/re-serialized transport red proof.
  - **Resolution:** **FIX.** Retain lossless raw JSON member/value tokens and patch only the selected top-level member (or constrain the named surface to an interaction that cannot rewrite arbitrary values), while keeping the raw editor authoritative. Prove whitespace/key-order preservation where unchanged, exponent/trailing-zero lexemes, integers beyond JavaScript's safe range, nested values, explicit null, and byte equality at both identity and render calls.

- [x] [Review][Patch][High] **Treat valid Object-prototype parameter names as own names.**
  - **Category:** AC1 · reserved/special names · editor correctness
  - **Location:** `folio-designer/src/App.tsx:523-539`; `folio-designer/src/engine-protocol.ts:95`
  - **Observation:** The engine/protocol validly admit names such as `constructor`, `toString`, and `__proto__`, but `ParameterEditor` reads `values[name]` from a normal prototype-bearing object. When the accepted document omits such a key, the lookup returns an inherited function/object instead of `undefined`; `useState` may even invoke the inherited function as a lazy initializer.
  - **Impact:** Valid engine-discovered parameters can open with fabricated values such as `[object Object]`, fail JSON editing, or behave differently from ordinary names, so the author cannot reliably supply them.
  - **Resolution:** **FIX.** Use a `Map`, a null-prototype dictionary, or explicit own-property lookup end to end. Add empty/present cases for `constructor`, `toString`, `valueOf`, `hasOwnProperty`, and `__proto__`, and prove they remain ordinary JSON member names without prototype mutation or inherited fallback.

- [x] [Review][Patch][High] **Keep keyboard focus when a named value becomes valid.**
  - **Category:** AC1 · accessibility · keyboard operation
  - **Location:** `folio-designer/src/App.tsx:523-531`
  - **Observation:** Each `ParameterValueInput` key includes its accepted value. As soon as typed text becomes valid JSON, acceptance changes that key and React unmounts/remounts the input. A number is valid after its first digit; `null` and a quoted string remount at completion, dropping focus in the middle of an ordinary keyboard workflow.
  - **Impact:** The named editor is not continuously keyboard-operable: multi-digit numbers require refocusing after the first digit, and every successful value completion unexpectedly ejects focus.
  - **Resolution:** **FIX.** Key inputs only by stable engine reference identity and reconcile external accepted-value changes without remounting the focused control. Add real focus assertions while typing multi-digit numbers, strings, null, invalid-to-valid transitions, raw-editor reconciliation, and document replacement.

- [x] [Review][Patch][Medium] **Reject parameter-reference metadata on render and identity responses.**
  - **Category:** AC1/AC2 · strict protocol admission · operation correlation
  - **Location:** `folio-designer/src/engine-protocol.ts:180-188`; `folio-designer/src/engine-client.ts:86-102`
  - **Observation:** `parseInbound` admits any success envelope containing any allowed optional payload, and the client mismatch check rejects `parameterReferences` for serialize/other operations but not for render or identity. A buggy/hostile render or identity response can therefore carry an unrelated reference projection and still resolve successfully.
  - **Impact:** The operation-specific closed protocol asserted by the story is no longer closed; unrelated metadata crosses an authority boundary and is returned to a caller that did not request it, creating a future accidental-consumption path.
  - **Resolution:** **FIX.** Require `parameterReferences` to be absent for every operation except `parameter-references`, preferably through one exhaustive operation-to-success-shape admission table. Add client-level red tests for extra references on render, identity, serialize, snapshot, and command responses.

- [x] [Review][Patch][High] **Refresh engine-derived references after Undo/Redo changes the active revision in Preview.**
  - **Category:** AC1/AC2 · revision freshness · lifecycle · undo/redo
  - **Location:** `folio-designer/src/App.tsx:215-233`; `folio-designer/src/App.tsx:432-443`; `folio-designer/src/App.tsx:500-505`
  - **Observation:** References load on entry to Preview and after successful Open/Start blank, but `applyHistory` changes the canonical revision without reloading them. Undo and Redo remain enabled in Preview, so a history step can add/remove params expressions while the editor continues painting the prior revision's names.
  - **Impact:** The UI attributes stale engine facts to the current canonical document: required fields can be missing, removed fields remain editable, and the next exact render uses a different template than the displayed reference list.
  - **Resolution:** **FIX.** On every admitted canonical revision change while Preview is active, revoke the prior projection and request a newly revision-correlated one; token the request so late same-document results cannot overwrite newer names. Prove undo-add, undo-remove, redo, delayed old projection, equal-number document replacement, and parameter values remaining transient/non-history state.

- [x] [Review][Patch][Medium] **Represent reference loading and failure honestly instead of painting both as an engine-confirmed empty list.**
  - **Category:** AC1 · error handling · UX truthfulness
  - **Location:** `folio-designer/src/App.tsx:69`; `folio-designer/src/App.tsx:215-226`; `folio-designer/src/App.tsx:523-525`
  - **Observation:** `parameterReferences` starts as `[]`; the catch path also sets `[]`; and the renderer labels either state "The engine found no editable parameter references." There is no distinction among not requested, loading, admitted empty, stale-revoked, or failed discovery.
  - **Impact:** The app states an engine fact it has not received and, after a real failure, tells the author the template has no references even while production render may report a missing one.
  - **Resolution:** **FIX.** Model pending/ready-empty/ready-nonempty/error states explicitly, show the no-reference claim only for an admitted correlated empty projection, and provide bounded local failure/retry copy without inventing parameter diagnostics or guessing names.

- [x] [Review][Patch][High] **Replace the Delivery Log's coverage claims with executable Story 6.3 proofs.**
  - **Category:** AC1-AC3 · verification gap · delivery-record accuracy
  - **Location:** `folio-go/parameter_references_test.go:10-27`; `folio-designer/src/App.test.tsx:94-113`; `folio-designer/src/engine-protocol.test.ts:123-133`; `folio-designer/e2e/preview-parameters.spec.ts:11-25`; Delivery Log §Implementation
  - **Observation:** The only new Go discovery test covers repeated direct text references. The App test checks that identity bytes merely contain one string, starts from an already-unsaved document, and never proves exact render bytes, identity change, save/revision/history invariance, late-result rejection, replacement/cancellation, number lexemes, special names, focus retention, null, or wrong-kind behavior. The protocol test does not exercise client operation matching. The compiled E2E source checks the missing path but not the required element id, and does not prove a successful current exact Preview after supplying the value. Several concrete bugs above survive all 142 green designer tests and the focused Go suite.
  - **Impact:** The recorded proof claims are materially stronger than the executable evidence, so the highest-risk authority, precision, lifecycle, diagnostic, and accessibility regressions can ship behind an all-green gate report.
  - **Resolution:** **FIX.** Add focused Go/wasm, client, App, static-authority, and compiled E2E-source tests for every named proof/red-proof clause, including both exact third-input call sites and located element/path provenance. Keep real Playwright and the four-target matrix explicitly deferred under D-000.4; do not claim their results.

### Review Summary

- **Outcome:** **Resolved.** All **9** persisted findings are **FIX**: **7 High** and **2 Medium**; 0 decision-needed, 0 deferred, and 0 dismissed.
- **Authority/lifecycle audit:** Go/wasm remains the only template-expression parser; no TypeScript template parser, alternate render path, second worker/wasm instance, document command, persistence, telemetry, or save-byte route was added for parameters. The settled three-input identity/render flow remains in place. The blocking gaps are incomplete/bounds-incompatible discovery, lossy named editing, special-name/focus failures, non-exhaustive response admission, stale revision projection, and dishonest projection state.
- **Review execution:** The commissioning handoff prohibited subagents, so the adversarial, edge-case, verification-gap, and acceptance lenses were applied directly. Unrelated `_bmad` configuration/manifest churn, `.agents/`, and planning research were preserved and remain outside the Story 6.3 file list.
- **Measured gates:** Designer Vitest passed **142 tests across 26 files**; typecheck, E2E-source compile, lint (four pre-existing Fast Refresh warnings only), production build/ordinary offline verification, wasm witness, and offline red proof passed. Focused `folio-go`/wasm passed **602 tests across 2 packages** plus vet, native build, and js/wasm build. `lint` passed **117 tests across 4 packages** plus vet/build; `hashmatrix` passed **3 tests across 2 packages** plus vet/build. `gofmt -l` and `git diff --check` were clean. Full `folio-go go test ./...` reproduced only the documented P6g corpus-floor red: **1,232 passed, 2 failed, 4 skipped across 18 packages**.
- **Explicitly unrun:** Real Playwright execution and the real four-target hash matrix were not run under unchanged D-000.4 per-epic cadence. Their source/static gates were inspected; neither result is claimed as green.
- **Cleanup:** Reviewer-generated `folio-designer/dist` and `folio-go/engine` outputs were removed. No implementation fix, fixture/golden/font edit, commit, push, or remote operation was performed.

### Finisher — 2026-08-28

- Resolved all nine review findings. Go now includes direct and nested `params.<member>...` paths and every element's `visibleIf`, including a table's common visibility expression, while deliberately excluding table columns/footers. Its collection walk carries element/field provenance on any projection failure and enforces the same 128-character direct-name bound as the browser protocol.
- The named editor is a raw-token replacement surface, not a JSON serializer: it validates JSON input, replaces only the selected top-level value, and leaves other whitespace, key order, nested members, exact integer spelling, exponent form, trailing zeroes, and `null` bytes untouched. Its null-prototype value dictionary treats `constructor` and `__proto__` as ordinary own JSON keys. Stable name-only React keys preserve focus after accepted input.
- Reference discovery now has explicit pending, ready-empty, ready-nonempty, and failed states. It never calls an unavailable or stale result an empty engine finding, and an active Preview re-queries the revision-correlated Go projection after Undo or Redo. The client closes operation envelopes by rejecting stray parameter-reference metadata on render, identity, serialize, and ordinary snapshot/command responses.
- Executable witnesses now cover table/nested discovery, table-column exclusion, 128/129 bounds, located missing `params.reportDate` provenance at `e1`, exact identity bytes after a named edit preserving `1.00e+2`, `-0`, and `123.4500`, special names, focus, pending/failed/empty truthfulness, Undo refresh, and client response red admission. The compiled E2E source remains included; real Playwright and the four-target matrix remain explicitly deferred under D-000.4.
- Final scoped verification is recorded below. The full Go suite still has only the established unrelated `internal/text` P6g corpus-floor red (7 opaque names, floor 20); Story packages and all targeted witnesses pass. No unrelated configuration, research, fixture/golden/font, or `.agents/` changes are included.
- Final gates: `npm test` passed **147 tests across 26 files**; `npm run typecheck`, `npm run lint` (four existing Fast Refresh warnings only), `npm run test:e2e:compile`, `npm run build`, `npm run verify:offline`, `npm run verify:offline:wasm`, and `npm run verify:offline:red` passed. `go test ./... -run TestParameterReferences`, `go vet ./...`, native `go build ./...`, and `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine` passed; `lint` passed **117 tests** plus vet/build and `hashmatrix` passed **3 tests** plus vet/probe build. `gofmt -l` and `git diff --check` are clean. Real Playwright/four-target execution remains deferred.
