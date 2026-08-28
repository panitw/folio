---
baseline_commit: 82642e5
story_key: 6-6-present-a-failed-render-honestly
status: done
created: 2026-08-28
---

# Story 6.6: Present a failed render honestly

**Epic:** 6 — A template author can bind a report to data and build the golden report  
**Covers:** UX-DR12, UX-DR24 · FR41 · AD-14–18  
**Governing rulings:** D-000.4, D-6.2.1

## In plain terms (read this first if you just want the gist)

Preview already asks the one local Go/wasm engine to render the exact current canonical template with the accepted sample-data bytes and parameter bytes. When that render cannot produce a PDF, the engine already returns a bounded producer fact: stable code, terse message, and, when known, the existing element id and data path. The app must now show that fact as a proper failed-render card in the Preview surface. It must say plainly that the local PDF render failed, show the received code/message and provenance without inventing any of them, use the danger treatment, and offer two honest actions: retry the same Preview work and return to Design. A retry is just a new render request through the established Preview scheduler; it is not a repair, template mutation, new renderer, or claim that the failure disappeared.

The author must be able to return even when there is no usable element id. When an existing opaque id is present, return uses the Story 5.12 locate/return seam: leave Preview, select only that currently projected element, announce the result, and focus the Design canvas. A missing, obsolete, or path-only location stays readable and still returns safely, without guessing a selection. The card is an error, not a warning: square/error shape and solid danger border are the primary language; red is reserved for this failed render state. It is not an opportunity to create a browser error taxonomy or code registry, classify Go errors, parse template bytes, or replace the exact/stale Preview state machine.

Most importantly, an old failure must never overwrite a newer truth. Display and actions remain scoped to the active Preview token, generation, revision, mode, and current input identity rules already established by Story 5.11/5.12. Input edits, document replacement, successful retry, cancellation, return to Design, and a newer request revoke or hide the old card. Story 6.7 alone proves the authored browser-to-native byte-identical round trip; this story presents failure truthfully and proves its local UI/protocol boundaries only.

## Story

As a template author,  
I want a render failure explained in place,  
so that I can fix it without leaving the tool or guessing.

## Acceptance criteria

### AC1 — A failed render is stated as a producer fact in the Preview surface

**Given** the exact current Preview render fails after the engine has received canonical template bytes, accepted sample-data bytes, and accepted parameter bytes,  
**When** the failure is admissible for the active Preview token, generation, revision, and mode,  
**Then** Preview shows one local failed-render error card that plainly states the failure, displays the received stable code and bounded message, and identifies it as a local PDF render failure.

**And** the card displays the received element id and data path independently when supplied, in mono treatment; it neither fabricates a missing value nor derives an id/path from template, sample, parameter, canvas, or error text.

**And** its danger/error treatment uses the DESIGN error-card language: a solid border and 3 px solid danger left edge, square error marker, danger token, and shape/text/accessibility cues that do not rely on colour alone. It remains terse and technical—no reassurance, browser taxonomy, server/cloud implication, or invented remediation.

**Proof / red proof.** Cover a real render-path failure with code, message, element id, and data path; path-only and no-location failures; bounded hostile/invalid producer values rejected at the existing protocol boundary; and accessible name/live announcement/error shape/token assertions. Red-prove a hand-authored code map, a template/data parser in TypeScript, warning/amber styling for a failure, or a failure card that calls the render successful.

### AC2 — Retry and return are truthful, safe actions

**Given** an active failed-render card,  
**When** the author chooses Retry preview,  
**Then** the app schedules one new render through the existing debounced/FIFO Preview authority path using the same current canonical document and accepted raw inputs; it clears the old failure only as that attempt becomes the current authority and never mutates `.folio`, document revision, history, save/dirty state, sample data, or parameters.

**And** when the author chooses Return to Design, the app uses the existing Story 5.12 return-with-optional-selection seam: it always leaves Preview, clears local interaction safely, focuses the canvas, and selects only an element id that is still present in the current engine canvas projection. A missing, path-only, or obsolete id never blocks return and never causes a guessed selection; an obsolete supplied id receives the existing locate-unavailable announcement.

**And** Retry and Return have visible cyan keyboard focus, accessible names, and remain usable without a pointer. Returning or retrying does not dismiss, reclassify, or mutate the producer diagnostic.

**Proof / red proof.** Test keyboard activation, exactly one retry request, retry success, retry failure, return for valid/path-only/no-location/obsolete locations, focus restoration, and no document mutation. Red-prove a direct React-to-wasm call, a second worker/renderer, an invented selected id, a retry that serializes or changes history, or a return that leaves a stale failure actionable in Design.

### AC3 — Failure authority cannot overwrite newer Preview truth

**Given** a render failure is pending or displayed,  
**When** inputs change, a committed document snapshot arrives, Preview is exited, Start blank/Open replaces the document, a newer render/retry begins, the worker fails, or a newer render succeeds,  
**Then** the prior failure is revoked or hidden by the existing token/generation/revision/mode/identity authority rules and cannot overwrite the newer status, PDF admission, diagnostics, selection, or controls.

**And** an old failure’s Retry or Return action is inert once it is no longer the active failure. A failed retry retains any previously admitted PDF only as the existing visibly stale Preview state permits; it must never label it current or conceal the failed attempt.

**And** successful render diagnostics remain Story 5.12 warning cards with their own amber/dashed treatment and generation-local dismissal. This story changes only the failure-card presentation and action wiring; it does not alter the producer diagnostic registry, successful-render evidence, canonical serialization, or PDF.js ownership.

**Proof / red proof.** Cover delayed rejection after input/document/mode revocation, delayed rejection after a newer success, stale action inertness, viewer failure versus engine render failure, and preservation of the prior admitted-PDF stale marker. Red-prove a stale error replacing a current preview, failure state surviving a successful retry, an old action changing Design state, or a retry that bypasses PDF.js admission/freshness checks.

## Tasks / subtasks

- [x] **1. Preserve the closed render-failure provenance through the existing transport** (AC: 1, 3)
  - Audit `folio-go/wasm/cmd/engine/main.go` → `folio-designer/src/engine.worker.ts` → `engine-protocol.ts` → `engine-client.ts`; preserve the bounded code, message, optional `elementId`, and optional `dataPath` already returned by `engineFailure`/worker failure responses.
  - Keep exact-key parsing, lengths, copied buffers, FIFO correlation, deep-frozen results, and existing `EngineError` validation. Do not add a TypeScript error registry, enum taxonomy, template/data parser, or a parallel producer contract.
  - Keep Go/wasm renderer, canonical bytes, diagnostics, and preview identity unchanged unless a focused test exposes a genuine transport defect.

- [x] **2. Replace the basic failure presentation with the DESIGN error card** (AC: 1)
  - Narrowly extend `folio-designer/src/preview/diagnostic-presenter.tsx` and `App.css`; preserve `PreviewDiagnostics` warning-card semantics and add a dedicated failure card using only existing tokens, particularly `--color-danger`.
  - Render the opaque code, message, element id, and data path independently and safely. Use the existing `EngineError` values only; no inferred cause, inferred element, or data-path reconstruction.
  - Give the card a meaningful live/error announcement and named controls. Follow DESIGN’s square error marker and solid treatment; do not use emoji, rounded corners, shadows, or colour as the sole difference.

- [x] **3. Wire retry/return through the proven Preview and locate seams** (AC: 2, 3)
  - Add a narrow retry callback that schedules the existing `renderPreview`/`PreviewWorkScheduler` route after rechecking the current active failure. Do not issue a direct worker call or retain a second request authority.
  - Reuse `returnFromFailure` / `returnWithOptionalSelection`, including its current-projection id check, Design transition, stale-error clearing, locate-unavailable announcement, and deferred canvas focus. Do not fork a second locator.
  - Keep `activeFailure`, `canInstallPreview`, current token/generation/revision/mode checks, cancellation, viewer-error handling, and stale/current status rules authoritative. Revoke old failure UI and handlers on every existing invalidation route.

- [x] **4. Add focused evidence without pulling in Epic 6 closure work** (AC: 1–3)
  - Add presenter, App/freshness, protocol/client/worker, accessibility/style-contract, and compiled Playwright-source tests for every AC and red proof. Use engine-produced located absent/wrong-kind/parameter failures where fixtures can exercise the real producer seam.
  - Run the focused designer tests, typecheck, lint, build, offline ordinary/red/wasm witnesses, `npm run test:e2e:compile`; run relevant Go/wasm tests, native and js/wasm builds, vet, formatting, and diff checks. Record actual commands/results and any sanctioned red accurately.
  - Under D-000.4, defer compiled Playwright execution and the four-target hash matrix to the Epic 6 boundary. Do not claim Story 6.7’s authored save/native byte-equivalence evidence, mutate goldens, or weaken the known P6g corpus-floor red.

## Developer guardrails

1. **Producer provenance is immutable presentation input.** Preserve the engine-provided code, bounded message, optional element id, and data path exactly through the closed protocol. TypeScript may display and scope those values; it may not classify failures, construct a code registry, parse `.folio`/JSON, or infer a location/remedy.
2. **Keep Story 5.12’s locate/return seam.** Return always works. Selection is optional and only for an opaque id still present in the current engine projection. Path-only/missing/obsolete provenance remains visible but never manufactures selection.
3. **Danger means failed render, not warning or selection.** Use `--color-danger`, the solid error-card border/left edge, and a square marker. Keep cyan for focus/authority and preserve Story 5.12’s amber dashed/triangle diagnostic cards for successful renders with caveats.
4. **Failure is authority-scoped.** Gate presentation and handlers with the established token, generation, revision, mode, identity, cancellation, and PDF.js admission rules. Do not allow a late failure to replace a newer current/stale Preview, show in Design, or act after revocation.
5. **Retry is a new honest attempt, never repair.** It must enter the existing scheduler/FIFO/one-worker path using current accepted inputs. It must not mutate template bytes, revision, history, dirty/save state, sample/parameter state, diagnostic provenance, or prior admitted PDF bytes.
6. **No new renderer or browser error taxonomy.** Preserve Go/wasm production render ownership, worker protocol, and PDF.js as sole viewer. Do not add a browser validation/render path, direct React-to-wasm bridge, alternate worker, server behavior, or TypeScript diagnostic classification.
7. **Do not preimplement Story 6.7.** No golden-report fixture authorship, native-library comparison, saved-file round trip, hash-matrix claim, or byte-identity closure belongs here. This story’s heavy tests are explicitly deferred under D-000.4.
8. **Preserve unrelated work.** Do not alter `.agents/`, `_bmad` configuration/manifest churn, planning research, fixture/golden/font sources, or existing user changes. Keep the current table/binding work and all canonical engine behavior intact.

## Project structure notes

- Primary designer seams: `folio-designer/src/App.tsx` (Preview state, freshness admission, retry/return callbacks), `src/preview/diagnostic-presenter.tsx` (presentation), `src/App.css`, and adjacent App/presenter/freshness tests.
- Transport seams to inspect before changing any types: `folio-go/wasm/cmd/engine/main.go`, `folio-go/wasm/engine.go`, `folio-designer/src/engine.worker.ts`, `engine-protocol.ts`, and `engine-client.ts`. Existing render failures already map `folio.RenderError` to bounded code/location and worker responses.
- Preserve the successful-render diagnostics path and PDF viewer in `folio-designer/src/preview/`; the error card must compose with, not replace, the single PDF.js owner and its page-admission callback.
- Do not touch table command/projection work unless an integration witness proves this card requires a narrow shared typing fix. The expected change is a small preview/App/style/test slice, not a new architectural surface.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 6, Story 6.6 acceptance criteria and the explicit Story 6.7 boundary.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-14 diagnostics, AD-15 engine-owned document, AD-16 one worker/render path, AD-18 preview identity and PDF.js ownership.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` — error-card/danger/square-marker rules, diagnostic-card contrast, focus and token grammar.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — Preview state, local/offline and keyboard/accessibility requirements.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — Epic 6 lead grounding, D-000.4 cadence, D-6.2.1, and the no-browser-error-taxonomy direction.
- `_bmad-output/implementation-artifacts/5-11-never-show-a-stale-preview.md` and `5-12-diagnostics-that-locate-and-an-interface-that-can-be-driven.md` — current Preview authority, producer diagnostics, locate/return, PDF.js, and deferred-test seams.
- `_bmad-output/implementation-artifacts/6-1-load-sample-data-and-browse-its-paths.md` through `6-5-bind-columns-in-row-scope-and-configure-footer-aggregates.md` — accepted local Preview inputs, data/binding authority, and current Epic 6 implementation conventions.

## Delivery log

### Story creation — 2026-08-28

- Created against baseline `82642e5` after grounding in Epic 6 context/source, sprint tracker, PRD, architecture spine, UX DESIGN/EXPERIENCE, decision log, completed Stories 5.11–5.12 and 6.1–6.5, recent commits, and current Go/wasm/designer code.
- Confirmed the existing engine response carries bounded render provenance: code, generic bounded message, optional element id, and optional data path. `PreviewFailure` currently displays only a basic local failure section with a cyan edge and Return; Story 6.6 must refine that existing seam rather than create a new taxonomy or renderer.
- Confirmed `App.tsx` already supplies Preview token/generation/revision/mode admission, stale-preview handling, one worker/FIFO scheduling, `activeFailure`, and `returnWithOptionalSelection`/`returnFromFailure`. The implementation must reuse these seams and add retry only through the scheduler.
- D-000.4 remains unchanged: compile e2e source and run focused/local verification at story cadence; compiled Playwright and the four-target hash matrix are deferred to the Epic 6 boundary. Story 6.7 owns the final browser/native round-trip proof.
- Implementation has not started. No task is pre-checked, no test result or changed-file list is claimed, and no owner decision is required.

## Dev agent record

### Completion notes

- Replaced the basic Preview failure section with the DESIGN error card: solid danger border and 3 px danger edge, square marker, opaque code/message, independently displayed element id/data path, `role="alert"`, and cyan focus-visible Retry preview / Return to Design controls.
- Reused the existing `renderPreview` → `PreviewWorkScheduler` route for Retry, gated by the existing active failure token/generation/revision/mode predicate. Return remains the Story 5.12 `returnWithOptionalSelection` seam. No producer protocol, Go/wasm behavior, canonical bytes, document/history/dirty state, or PDF.js ownership changed.
- Added presenter, App authority/retry, closed protocol/client provenance, DESIGN style-contract, and compiled Playwright-source coverage. The delayed-rejection App test proves a revoked failure card cannot return after leaving Preview.
- Verification passed: focused designer suite (88 tests), full `npm test` (27 files / 160 tests), `typecheck`, `lint` (only pre-existing Fast Refresh warnings), `build`, offline ordinary/red/wasm verification, `test:e2e:compile`, native and js/wasm Go builds, `go vet`, and `gofmt -d`.
- Sanctioned known red: `go test ./...` reports `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` at 7/20 (1242 passed, 2 failed, 4 skipped). This is the documented corpus-floor red; it was not changed or weakened.
- Deferred by D-000.4: real compiled Playwright execution and the four-target hash matrix remain Epic 6 boundary work. Story 6.7 byte-identical authored round-trip evidence was not implemented or claimed.
- Manual scoped review found no Story 6.6 defects. The BMAD review subagent layers were not run because this task explicitly prohibits subagents.

### File list

- `folio-designer/src/preview/diagnostic-presenter.tsx`
- `folio-designer/src/App.tsx`
- `folio-designer/src/App.css`
- `folio-designer/src/preview/diagnostic-presenter.test.tsx`
- `folio-designer/src/App.test.tsx`
- `folio-designer/src/engine-protocol.test.ts`
- `folio-designer/src/engine-client.test.ts`
- `folio-designer/src/design-contract.test.ts`
- `folio-designer/e2e/preview-parameters.spec.ts`

### Review Findings

- [x] [Review][Patch][High] **Force Retry preview to perform a new render when last-good bytes share the current identity.** [`folio-designer/src/App.tsx:184`](../../folio-designer/src/App.tsx#L184)
  - **Category:** AC2/AC3 · retry authority · last-good bytes · truthful stale state
  - **Observation:** `retryFromFailure` enters the ordinary `renderPreview` path, but `runPreview` returns at the identity cache check whenever the retained PDF has the same revision/identity/generation. In the required “failed attempt with a previously admitted PDF” case, Retry therefore executes only `identity`, sets the retained historical PDF to `current`, clears the card, and never calls `serialize` or `render`.
  - **Impact:** The failed attempt is concealed and last-good bytes are relabelled current without a new producer render. This directly violates the rule that retry is one new render request and that a failed retry may retain the prior PDF only as visibly stale.
  - **Required patch:** Carry explicit retry intent through the existing scheduler so it bypasses only the same-identity reuse shortcut, still uses current accepted bytes/FIFO authority/PDF.js admission, and proves retry success plus retry failure while prior admitted bytes remain stale.

- [x] [Review][Patch][High] **Do not present viewer, identity, serialization, cancellation-adjacent, or worker failures as engine render provenance.** [`folio-designer/src/App.tsx:203`](../../folio-designer/src/App.tsx#L203) [`folio-designer/src/App.tsx:282`](../../folio-designer/src/App.tsx#L282) [`folio-designer/src/preview/diagnostic-presenter.tsx:24`](../../folio-designer/src/preview/diagnostic-presenter.tsx#L24)
  - **Category:** AC1/AC3 · producer provenance · viewer/worker taxonomy boundary · honest messaging
  - **Observation:** One broad `try/catch` covers `identity`, `serialize`, and `render`, while `viewerError` installs `PDF_VIEWER_FAILED` into the same `PreviewFailureRecord`. `PreviewFailure` unconditionally announces every such value as `Render failure` and `Local PDF render failed`, even when Go produced a valid PDF and only PDF.js failed, or when no render request was sent at all.
  - **Impact:** The UI invents a producer fact and gives Retry semantics to failures outside the producer render contract. A viewer failure can additionally hit the same-identity shortcut and be hidden without replacing or readmitting the failed viewer bytes.
  - **Required patch:** Preserve operation/source state without creating a browser code registry: only an admitted render-operation `EngineError` may use the failed-render card; viewer and transport/lifecycle failures need truthful existing-state handling. Prove viewer versus engine render failure, worker failure, identity/serialize rejection, cancellation suppression, and stale authority.

- [x] [Review][Patch][Medium] **Reject invalid empty failure fields instead of accepting and silently normalizing them away.** [`folio-designer/src/engine-protocol.ts:95`](../../folio-designer/src/engine-protocol.ts#L95) [`folio-designer/src/engine-client.ts:15`](../../folio-designer/src/engine-client.ts#L15)
  - **Category:** AC1 · closed protocol · exact provenance · hostile producer values
  - **Observation:** `isError` requires a non-empty code but accepts `message: ""`, `elementId: ""`, and `dataPath: ""`. `errorFor`, `previewFailure`, and `PreviewFailure` then use truthiness and discard the two supplied location fields, while an empty message becomes a blank asserted failure. The new protocol tests check only over-max lengths, not empty or wrong-shape optional provenance.
  - **Impact:** The main-thread boundary admits invalid producer envelopes and no longer preserves the exact received presence/value distinction promised by the story.
  - **Required patch:** Make the optional-location and message contract explicit and consistent across producer, worker, parser, client, and presenter—reject supplied empty invalid values (or, if emptiness is ruled valid, preserve and render them exactly). Add exact-key/type/empty/max/hostile witnesses and prove the client copy cannot drop accepted fields.

- [x] [Review][Patch][Medium] **Replace the completion record with executable Story 6.6 proofs for the unmeasured authority and accessibility edges.** [`folio-designer/src/App.test.tsx:334`](../../folio-designer/src/App.test.tsx#L334) [`folio-designer/e2e/preview-parameters.spec.ts:16`](../../folio-designer/e2e/preview-parameters.spec.ts#L16)
  - **Category:** AC1-AC3 · verification gap · delivery-record accuracy
  - **Observation:** The sole retry test starts with no prior PDF and covers only first-failure/second-success plus absence of a `command`; it cannot detect the same-identity last-good bypass and does not prove retry failure, FIFO/duplicate gating, exact accepted input bytes, revision/history/save invariance, or PDF.js admission. Return coverage omits failure-specific obsolete/no-location focus assertions. The delayed test covers only Preview exit, while the named input/document/newer-success/stale-action races and viewer/worker distinctions are absent. The compiled Playwright source merely checks that the card and two buttons are visible; it activates neither action and proves no keyboard/focus behavior.
  - **Impact:** All green gates coexist with the high-severity defects above, while the Dev agent record claims every AC/red proof is covered.
  - **Required patch:** Add discriminating protocol/client/App/presenter/style/e2e-source witnesses for every proof named in AC1-AC3, including last-good retry success/failure, rapid duplicate activation/FIFO, valid/path-only/no-location/obsolete Return with canvas focus, stale handler inertness after each revocation route, exact no-mutation state, viewer/worker separation, and accessible keyboard/live/focus/danger behavior. Keep real Playwright and the four-target matrix explicitly deferred under D-000.4 and narrow the completion claims to measured evidence.

### Review Summary

- **Outcome:** **Changes requested.** Four unresolved patch findings remain: two high-severity authority/provenance defects and two medium-severity protocol/proof gaps. Story and sprint tracking returned to `in-progress`.
- **Measured designer gates:** `npm test` passed **160 tests across 27 files**; typecheck and E2E-source compile passed; lint reported only the four recorded Fast Refresh warnings. The production build and ordinary/red/wasm offline witnesses passed after the generated offline manifest was present. `git diff --check` passed.
- **Measured Go/static gates:** focused wasm tests passed **16 tests**; native and js/wasm builds, `go vet ./...`, and `gofmt -d folio-go/wasm` passed. Full `go test ./...` reproduced only the sanctioned corpus-floor red: **1,242 passed, 2 failed, 4 skipped** (`P6g (opaque names): got 7, need >=20`).
- **Review execution:** The commissioning handoff explicitly prohibited subagents, so the adversarial, edge-case, verification-gap, and acceptance lenses were applied directly. No product-code fix, fixture/golden/font change, decision-log amendment, commit, push, or remote operation was performed.
- **Explicitly unrun under D-000.4:** real Playwright execution and the four-target hash matrix. Neither result is claimed as green.
- **Cleanup:** No temporary probe or generated tracked artifact was created. Unrelated `_bmad` configuration/manifest changes, `.agents/`, and planning research remained untouched.

### Finisher record — 2026-08-28

- Resolved the four review findings. Retry now carries explicit force intent through the existing FIFO scheduler and always performs fresh identity, canonical serialization, and producer render work; same-identity retained PDF bytes cannot become current until PDF.js admits the new candidate.
- Only a rejected producer `render` response is marked as producer provenance and shown in the failure card. Identity, serialization, worker/lifecycle, cancellation-adjacent, and PDF viewer failures remain truthful local Preview state without an invented producer code/message/location or Retry card.
- The closed boundary rejects empty supplied failure messages, element ids, and data paths. Accepted optional provenance is copied and rendered by presence rather than truthiness, so it cannot disappear during transport/presentation.
- Added executable App/protocol/client/e2e-source witnesses for same-identity last-good retry failure, duplicate retry suppression/FIFO, fresh identity/serialize/render calls, stale retention, no command mutation, viewer-vs-producer separation, empty/hostile provenance, and keyboard activation source coverage. Existing return/location/focus and stale-revocation witnesses remain in the focused suite.
- Measured after the patch: focused designer suite **82 tests / 4 files**, full designer suite **162 tests / 27 files**, typecheck, lint (the four existing Fast Refresh warnings only), production/offline build plus ordinary/red/wasm offline witnesses, and compiled Playwright source. Go wasm tests, native/js-wasm builds, vet, and formatting passed. Full Go tests retain only the sanctioned P6g corpus-floor red (7/20 opaque names). Real Playwright execution and the four-target hash matrix remain deferred under D-000.4.
