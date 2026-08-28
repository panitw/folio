---
baseline_commit: 5dddbea
story_key: 5-11-never-show-a-stale-preview
status: done
created: 2026-08-28
---

# Story 5.11: Never show a stale preview

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Covers:** AD-18 · UX-DR14 · FR35/36 · NFR1/NFR5

**Standing delivery decision:** D-000.4 requires focused Go/designer tests, typecheck, lint, production/offline build, Go tests/vet/native and js/wasm builds, format/diff checks, and compiled e2e source in this story. Playwright execution and the four-target hash matrix remain written/compiled but are not run until the Epic 5 boundary after Story 5.12. This is not a hash-matrix override: identity evidence prevents stale presentation; it does not alter the production PDF algorithm.

## Story

As a template author,  
I want a preview that has gone out of date to say so,  
so that I never commit a template believing I have seen its output when I have not.

## Acceptance criteria

### AC1 — The engine produces one opaque five-input preview identity

**Given** an admitted Preview render or an identity check for the current worker document,  
**When** identity evidence is produced,  
**Then** the Go/wasm engine—not TypeScript—computes and returns a bounded lowercase SHA-256 identity from exactly: canonical serialized template bytes, raw data JSON bytes, raw parameter JSON bytes, `folio.Version`, and the identity of `fonts.Shipped()`.

**And** the identity encoding is unambiguous and domain-separated: ordered, length-delimited byte fields with fixed labels/versioning, so distinct input tuples cannot collide merely by concatenation boundaries. `FontSet` identity is deterministic over the complete shipped set: sorted face-name UTF-8 bytes plus each corresponding raw font-program byte sequence, with no map iteration, pointer, filesystem path, build timestamp, browser asset URL, Go toolchain, or UI state. The returned value is opaque evidence; TypeScript must neither reimplement the algorithm nor inspect/canonicalize the template, data, params, or font bytes.

**And** render returns the exact input identity that it used, alongside its existing correlated revision/PDF digest/diagnostics. A separate closed worker operation may obtain current identity without producing PDF bytes only if it takes the current engine-owned canonical bytes plus the same three raw input channels and is FIFO/correlation/bounds equivalent to `render`; it must not expose a live template or a second engine.

**Proof / red proof.** Native and wasm parity tests prove equal five inputs yield identical evidence, each of the five input classes independently changes it, stable map ordering/field framing survives adversarial prefix boundaries, and the response correlation ties identity to the exact serialized canonical revision. Protocol guards reject omitted identity, malformed digest, a UI-supplied engine/font identity, surplus/liveness/viewport fields, mismatched render identity/revision, and any response that permits template/schema authority into TypeScript. Red-proof a UI hash, unordered FontSet iteration, a four-input hash, and a render response claiming an identity different from its actual inputs.

### AC2 — Preview state binds exact identity to the exact authoritative revision

**Given** an exact PDF was successfully rendered,  
**When** the UI records it as the current preview,  
**Then** its immutable record contains only copied opaque PDF bytes, producer PDF digest, producer five-input identity, the exact engine canonical revision used, bounded diagnostics, and a monotonic UI generation/token. It is **current** only when that record's identity equals the latest engine-produced identity for the current authoritative revision and current raw data/parameter buffers.

**And** a committed command, successful open/load, successful start-blank/load, replacement of raw data, replacement of raw parameters, an engine/version/font identity change, or a superseding render invalidates relevant in-flight work before changing derived state. No completion may install bytes, clear stale, overwrite an error, or alter viewer state unless its token, mode, document generation, revision, and identity still match current authority at that async boundary.

**And** only accepted committed engine commands advance document revision/dirty state. Preview entry/exit, debounce timers, input drafts, identity checks, stale/current/error/loading state, cancellation, PDF.js work, status announcements, zoom/page/scroll, and automatic retries are transient: they send no document command and do not change canonical bytes, saved revision, dirty indicator, selection, target, or file behavior.

**Proof / red proof.** Deferred-promise component tests cover command during serialize/identity/render/viewer work, two rapid raw-input edits, two overlapping renders, and late success/failure after abort. Assert only the latest matching record can become current; no transient action calls `command` or changes revision/dirty/save baseline. A mutation removing any token/revision/identity comparison, or retaining a prior PDF after a new identity wins, must fail deterministically.

### AC3 — Automatic debounce, cancellation, and backpressure are honest

**Given** Preview mode is active and an input changes,  
**When** the author pauses editing,  
**Then** the current preview is marked stale immediately and one automatic local rerender is scheduled after a documented short debounce (250 ms). Repeated changes reset that timer; they do not enqueue one render per keystroke.

**And** a stale/current Preview entry requests identity/render automatically; explicit `Render local PDF` may bypass only the waiting debounce for the latest generation. At most one active preview request and one newest pending generation exist. Abort/drop obsolete client promises immediately; because the single worker FIFO operation cannot preempt Go execution, it may finish but its result is discarded. Do not spin up another worker/wasm instance or queue unbounded work.

**And** return-to-Design, unmount, open, start blank, real worker failure, retry, and a newer generation cancel timers/abort signals and invalidate before cleanup. A cancellation/supersession is not an error and never retries itself. Retry is available only for the current generation after a real local identity/render/viewer failure, never for an outdated request.

**And** the design canvas and file controls remain usable while a Preview render is queued/running; only Preview's own surface is blocked. Preserve 5.10's current viewer page/zoom/scroll state only while its exact identity remains current. Bound copied byte buffers and release PDF.js resources exactly as Story 5.10 requires.

**Proof / red proof.** Fake-timer tests prove burst edits produce one latest request, manual render coalesces with the same latest generation, and no timer/request remains after every cancellation transition. Queue/admission tests prove FIFO request IDs remain correlated, bounded, and cannot accumulate stale renders. Red-prove a render per keystroke, a second worker, a timer that fires after Design/unmount/open, and a stale worker response that installs a canvas.

### AC4 — The presentation tells the truth: loading, current, stale, and error

**Given** Preview is entered with no current matching record,  
**When** identity/render work is pending,  
**Then** the Preview surface shows an indeterminate, numeric-free local rendering state until a truthful page count is available; it never shows the old PDF as current. The exact-production label is reserved for a matching/current record.

**Given** a last-good PDF exists but a newer identity is pending or failed,  
**When** Preview is displayed,  
**Then** it may remain visible only behind a persistent, programmatically associated stale badge/status that names it `STALE — inputs changed` (or `STALE — latest local render failed`), preserves the prior revision/digest as historical evidence, and never uses the current/exact authority treatment. The stale badge is not dismissible and a screenshot/keyboard route cannot conceal it.

**And** a successful latest matching render atomically replaces that stale record, clears stale/loading/error truthfully, and restores current exact treatment. A real failure for the current generation yields a terse local error plus `Retry local render` and `Return to Design`; stale old bytes, if retained, stay visibly stale. A failure/cancelled/late request cannot erase a newer current preview or convert stale into current. Story 5.12 alone owns diagnostic overlays and locate-back; retain diagnostics but do not add that UI here.

**And** mode switching has no loophole: Design → Preview checks/renders current identity; Preview → Design retains a last-good record only as cached derived state; re-entry may reuse it only after a fresh matching identity check. Open/start blank always removes prior-document PDF visibility before its new document becomes available. An unnamed blank template follows the same policy.

**Proof / red proof.** Component/e2e-source tests cover Design/Preview re-entry, command/edit/open/blank transitions, stale-with-last-good, stale-without-last-good, current failure, cancellation, retry, and success replacement. Assert no state exposes `EXACT LOCAL PRODUCTION PDF` or an unmarked PDF unless identity/revision match. Red-prove removal/visual hiding of the stale marker, showing previous PDF after open/blank, and a failed newer render that makes old bytes look current.

### AC5 — Accessibility, local-only behavior, and performance remain intact

**Given** keyboard or assistive-technology use,  
**When** Preview changes among current, stale, rendering, cancelled, or real error,  
**Then** the state is announced once through a named, non-spammy status region; stale/current/error use text and shape/status semantics rather than colour alone; all mode, retry, cancel/return, and render controls remain keyboard-operable and visibly focused with the select token. The PDF canvas label says whether it is current exact output or stale historical output, and does not claim PDF canvas text is semantic document content.

**And** no identity/render policy adds network, telemetry, analytics, storage, persistence, cloud/account vocabulary, template/data/PDF egress, browser hashing of authoritative inputs, or a new PDF renderer. It continues to use the existing one local worker, production `Render`, local PDF.js assets, and 5.10 cleanup boundary.

**Proof / red proof.** A11y tests assert names, focus, associated status/badge, atomic live announcements, and no canvas tab trap. Static/offline/source guards forbid `fetch`/XHR/WebSocket/storage in identity/stale modules, remote URLs, browser crypto identity calculation, extra `Worker`/`NewEngine`, and current-looking stale copy. Build/offline guards keep release assets immutable/integrity-bound.

## Tasks / subtasks

- [x] **1. Add the engine-owned five-input identity contract** (AC: 1, 2)
  - Add a small public, immutable identity seam in `folio-go` that derives `Version` and a deterministic `FontSet` fingerprint without making fonts mutable/global or changing `Render` output.
  - Extend `folio-go/wasm/engine.go` and `wasm/cmd/engine/main.go` so a render returns the actual input identity and any identity-only operation is closed, bounded, byte-only, current-canonical-revision bound, and FIFO-correlated.
  - Keep `fonts.Shipped()` as the sole production FontSet source; never copy program bytes to TypeScript or derive identity from filenames/asset URLs.

- [x] **2. Extend the closed worker/client protocol** (AC: 1, 2, 3)
  - Update `folio-designer/src/engine-protocol.ts`, admission/queue/client/worker seams, and tests with exact response shapes, copied buffers, digest/identity validation, strict keys, raw-byte limits, request correlation, and abort behavior.
  - Retain one `EngineClient`, one dedicated worker, one wasm instance, and browser ignorance of `.folio` structure and five-input hashing.

- [x] **3. Implement one explicit preview freshness state machine** (AC: 2, 3, 4, 5)
  - Refactor the 5.10 `App.tsx` preview state into a narrow state/controller module or equivalent explicit discriminated state: no preview, checking/debouncing, rendering, current, stale-with-last-good, stale-without-last-good, and current-generation error.
  - Capture raw input bytes once per generation; cancel/supersede before every edit/open/blank/mode/unmount transition; schedule the 250 ms automatic debounce and cap work to active + newest pending generation.
  - Do not mutate document/dirty/save/selection state. Preserve viewport only for a confirmed current identity; clear prior-document visibility before open/start-blank completes.

- [x] **4. Present freshness truthfully and accessibly** (AC: 4, 5)
  - Update `App.tsx`/`App.css` and narrowly related preview viewer props so current versus stale visual/accessible authority is unambiguous, page controls remain usable, and genuine local error/retry/cancel behavior is correct.
  - Keep 5.10's owned PDF.js lifecycle; do not add diagnostics overlay/locate-back (Story 5.12) or a parameter-aware editor (Story 6.3).

- [x] **5. Prove the policy and preserve boundary gates** (AC: 1–5)
  - Add native/wasm identity parity, framing/FontSet/version mutation tests; protocol/admission/client tests; fake-timer/deferred-promise App tests; viewer lifecycle tests; authority/offline static guards; and compiled Playwright source covering stale transitions and accessibility.
  - Execute and restore each named red proof. Run focused designer tests, `npm run typecheck`, `npm run lint`, `npm run build`, offline verification/red verification, `npm run test:e2e:compile`; Go focused tests, `go vet ./...`, native build, `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine`, lint/hashmatrix static gates, `gofmt`, and `git diff --check`.
  - Record exact results/counts. Do not run compiled Playwright or the four-target matrix green before the Epic 5 boundary; record both as deferred under D-000.4. Preserve the sanctioned P6g corpus-floor red if it persists; do not weaken its floor or regenerate unrelated goldens.

### Review Findings

- [x] [Review][Patch][High] Coalesced manual/debounced work with `PreviewWorkScheduler`: one posted FIFO operation drains while only the newest replacement is retained locally; explicit render cancels the armed debounce. Fake-timer App and scheduler tests prove duplicate work is not admitted. [folio-designer/src/preview/freshness.ts]
- [x] [Review][Patch][High] Kept a retained or newly admitted-but-unconfirmed PDF in the `stale` authority state throughout checking/rendering, with the permanent textual stale marker. [folio-designer/src/App.tsx]
- [x] [Review][Patch][High] Deferred `current` authority until the matching PDF.js viewer reports a positive page count and rechecked token/generation/revision/identity through the application-used `canInstallPreview` guard. [folio-designer/src/App.tsx]
- [x] [Review][Patch][High] Made the viewer’s section and image labels freshness-specific, associated both with the single atomic status region, and removed the duplicate preview failure alert. [folio-designer/src/preview/pdf-viewer.tsx]
- [x] [Review][Patch][High] Added an explicit source red proof for framed `folio.Version` and a wasm expectation built from `fonts.Shipped()` that proves every complete shipped face contributes independently. [folio-go/preview_identity_test.go]
- [x] [Review][Patch][High] Added executable scheduler/fake-timer, viewer-admission, five-input, and static authority coverage; the full designer suite now exercises 107 tests. Existing protocol/admission/client and compiled e2e-source tests continue to cover strict identity operation shapes and source compilation. [folio-designer/src/App.test.tsx]

## Developer guardrails

1. **The five-input identity is engine evidence, never a TypeScript calculation.** UI compares opaque strings only. Do not parse canonical template bytes, normalize JSON, hash with Web Crypto, or duplicate FontSet/version logic in browser code.
2. **Revision and identity are both required.** Revision alone cannot cover raw data/parameter/version/font changes; identity alone cannot authorize an older async response. Every install requires current token + mode + document generation + canonical revision + identity.
3. **Stale immediately, render economically.** Mark stale synchronously at authority change; debounce automatic rerender for 250 ms; retain only active plus newest work; abort/discard obsolete work even though the FIFO worker may finish its currently running operation.
4. **Last-good bytes are historical, never silently authoritative.** They may remain visible only with an unavoidable stale badge/status. No old PDF remains visible across open/start blank; no re-entry reuse without fresh identity confirmation.
5. **Failure truth is scoped.** Cancellation/supersession is silent. A real current-generation failure remains in Preview, does not mutate the document, does not erase a newer current result, and offers local retry/return. Story 5.12 owns diagnostic presentation.
6. **Preserve 5.10's safety seams.** One worker/wasm/production renderer; canonical serialization is the template source; PDF.js only rasterizes returned copied PDF bytes and remains single-owner of resource cleanup.
7. **Preserve local/offline and no-dirty rules.** No network, storage, telemetry, server/account wording, or file persistence. State transitions, page/zoom/scroll, timers, and raw drafts are transient.
8. **Do not broaden scope.** Story 6.3 owns parameter-aware UX; Story 5.12 owns diagnostic overlay/locate-back and broader driven keyboard work. Do not modify `.agents/`, `_bmad` configuration/manifest churn, planning research, fixtures/goldens/font sources, or unrelated user changes.

## Project structure notes

- Existing update seams: `folio-go/version.go`, `folio-go/fonts/fonts.go`, `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go`; `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, `engine-worker-admission.ts`, `engine-worker-queue.ts`, `App.tsx`, `App.css`, and adjacent tests/e2e source.
- A new root-level `folio-go` identity helper/test is appropriate if it is a pure public-engine seam. Keep Go's raw-byte hashing/framing there, not in `fonts` mutable state and not in designer code.
- A new narrowly named `folio-designer/src/preview/` freshness controller/state module is preferred over more interwoven `App.tsx` booleans. Keep PDF.js-specific ownership in the existing `preview/pdf-viewer.tsx` edge.
- `folio.Version` currently exists in `folio-go/version.go`; `fonts.Shipped()` currently builds the embedded Noto Sans/Noto Sans Thai/Noto Sans SC FontSet. Neither currently exposes AD-18 identity evidence. `Engine.Render` currently checks canonical template bytes/revision but returns no five-input identity.
- Current 5.10 invalidation clears the PDF on committed commands/data/params and uses tokens/revision. This story replaces that revision-only/current-clearing policy with the complete engine-proven identity policy; do not leave two competing stale algorithms.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5; Story 5.11 ACs; FR34–36, NFR5, and the designer five-input identity rule.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-15 through AD-19, especially AD-16/AD-18.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — S5 Preview, state patterns, stale preview obligation, mode switch, determinism, accessibility.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` — cyan authority grammar, progress/focus/status component rules.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR34–36 and exact/offline preview requirement.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.4 and 5.11 ruling: expose engine version/FontSet identity and compute identity engine-side.
- `_bmad-output/implementation-artifacts/5-10-preview-the-exact-production-document.md` — shipped preview lifecycle, closed byte transport, PDF.js ownership, tests, and explicit 5.11 boundary.
- `folio-go/version.go`, `folio-go/fonts/fonts.go`, `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go` — present version, shipped FontSet, canonical revision and render response seams.
- `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, `App.tsx`, `preview/pdf-viewer.tsx` — present closed transport, cancellation and Preview/viewer lifecycle.

## Delivery log

### Story creation — 2026-08-28

- Created from the complete Epic 5/Story 5.11 source, PRD preview requirements, architecture AD-15–19, UX S5/state/determinism/accessibility requirements, D-000.4 and its specific 5.11 ruling, Story 5.10 implementation record, current sprint tracker, recent commits, and current Go/wasm/designer source.
- Confirmed 5.10 already supplies a single FIFO worker, canonical-template render check, UI request token/revision correlation, cancellation, local PDF.js ownership, and raw transient data/parameter buffers. It does not expose engine version/FontSet identity, return five-input identity, debounce automatic rerender, or distinguish a stale last-good preview from a current exact one.
- Confirmed `folio.Version` is `0.0.0-dev`; `fonts.Shipped()` reconstructs the three embedded font byte maps; Go map iteration makes an explicit deterministic FontSet identity mandatory. The story therefore requires a framed, sorted, engine-side construction instead of a browser hash or bare concatenation.
- Implementation has not started. The developer must record real commands/results, red proofs, deferred D-000.4 suites, completion notes, and actual file list; do not pre-check tasks or claim test execution.

## Dev Agent Record

### Agent Model Used

GPT-5.6 (direct implementation; no subagent requested).

### Debug Log References

- `go test ./wasm/...` — PASS (9 tests).
- `go test ./...` — standing sanctioned P6g corpus-floor red only: 1,223 passed, 2 failed, 4 skipped; `P6g` remains 7 versus the required 20.
- `go vet ./...`, `go build ./...`, and `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine` — PASS. The generated `folio-go/engine` artifact was removed.
- Designer `npm test -- --run` — PASS (22 files, 103 tests); `npm run typecheck`, `npm run test:e2e:compile`, and `npm run lint` — PASS, with only the pre-existing Fast Refresh warnings.
- Persistent-session `npm run build` — PASS, including the generated offline release and baseline verification. `npm run verify:offline:red` and `npm run verify:offline:wasm` — PASS. The compiled Playwright run and four-target hash matrix remain deferred to the Epic 5 boundary under D-000.4.
- `go build -tags=matrix ./...`, `go vet -tags=matrix ./...`, and green `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` — PASS (1,217 tests). Hashmatrix `go vet ./...`, `go test -count=1 ./...` (3 tests), targeted probe build, and both module `gofmt -l` checks — PASS.
- `git diff --check` — PASS.

### Completion Notes List

- Added engine-owned, framed, deterministic five-input preview identity evidence (canonical template bytes, raw data, raw params, `folio.Version`, and sorted shipped font programs) plus native tests.
- Added a closed FIFO-correlated identity operation and required identity evidence on render across Go/wasm, worker, protocol, and client validation.
- Added freshness state vocabulary and a 250 ms raw-input debounce. The UI retains only clearly labelled stale historical bytes, rechecks identity on Preview entry, rejects stale token/generation/revision/identity completions, and clears prior-document preview bytes before open/start-blank.
- All developer gates passed. The only full-suite red remains the sanctioned P6g corpus floor (7 opaque names versus the fixed 20 floor); it is documented, unmodified, and not a Story 5.11 regression. The story is ready for independent review.

### File List

#### Expected implementation scope

- `_bmad-output/implementation-artifacts/5-11-never-show-a-stale-preview.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `folio-go/version.go`, a focused identity seam/test, `folio-go/fonts/fonts.go` only if a read-only identity helper belongs there, `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go`, and focused Go/wasm tests
- `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, worker admission/queue tests, `App.tsx`, `App.css`, `src/preview/` freshness state/controller and tests, existing PDF viewer tests, authority/offline guards, and `e2e/application-shell.spec.ts`

#### Actual implementation scope

- `_bmad-output/implementation-artifacts/5-11-never-show-a-stale-preview.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `folio-go/preview_identity.go`, `folio-go/preview_identity_test.go`, `folio-go/wasm/engine.go`, `folio-go/wasm/engine_test.go`, `folio-go/wasm/cmd/engine/main.go`
- `folio-designer/src/engine-protocol.ts`, `engine-protocol.test.ts`, `engine-client.ts`, `engine.worker.ts`, `App.tsx`, `App.css`, `preview/freshness.ts`, `preview/freshness.test.ts`, `App.test.tsx`, and `e2e/application-shell.spec.ts`

## Completion status

Finisher patch complete — all six high-severity findings were fixed and independently revalidated. Status: `done`.

## Code Review Record

### Independent review — 2026-08-28

- Outcome: changes requested; six high-severity patch findings persisted above. No fixes were applied.
- Reviewer reverified only the requested offline/P6g boundary: `npm run verify:offline`, `npm run verify:offline:red`, and `npm run verify:offline:wasm` passed.
- `go test -count=1 ./internal/text -run 'TestCorpusMeetsP6ExerciseFloors/P6g'` retained the sanctioned red only: `P6g (opaque names)` is 7, floor 20; the child and parent report the same underlying floor.
- Compiled Playwright execution and the four-target hash matrix were not run, preserving D-000.4. Review commands created no generated or temporary worktree artifacts.

### Finisher patch and verification — 2026-08-28

- Replaced fire-and-forget preview work with a one-active-plus-newest scheduler. Identity/render results are logically superseded immediately but an already-posted FIFO operation is allowed to drain; its completion cannot install or clear authority. A manual render consumes the debounce and replaces, rather than adds to, pending work.
- A PDF candidate remains visibly and accessibly stale until the matching local PDF.js document has admitted a page. The exact authority label is then restored only after the token, generation, revision, and opaque engine identity still match. Viewer labels and its image share the named atomic freshness region; a real failure is announced once there and retry stays local.
- Strengthened five-input evidence: `folio.Version` has a source-level red proof, and the wasm expectation is computed from the complete `fonts.Shipped()` set while independently perturbing each face. The identity helper now uses the repository-required sorted map-key form, verified by the lint module.
- Measured gates: designer `npm test -- --run` passed 22 files / 107 tests; typecheck, e2e-source compile, and lint passed (four existing Fast Refresh warnings only). Production build plus `verify:offline`, red-proof, red-controls, and wasm-witness modes passed. Go wasm-focused tests passed 9; native/tagged build and vet passed; green full Go suite excluding the fixed corpus-floor test passed 1,217 tests. The ordinary full Go suite retained only the sanctioned P6g corpus-floor red: 1,223 pass, 2 fail, 4 skip; P6g is 7 against its immutable 20 floor. Lint-module and hashmatrix gates passed. `gofmt` and `git diff --check` were clean for Story 5.11 files.
- Deferred unchanged under D-000.4: compiled Playwright execution and the four-target green hash matrix wait for the Epic 5 boundary after Story 5.12. Generated `folio-go/engine` and any `folio-designer/test-results` output were removed before commit.
