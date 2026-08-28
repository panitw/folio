---
baseline_commit: 592e1b8
story_key: 5-10-preview-the-exact-production-document
status: in-progress
created: 2026-08-28
---

# Story 5.10: Preview the exact production document

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Covers:** FR34, FR35 · NFR5 · AD-16, AD-18 · UX-DR9, UX-DR15, UX-DR23

**Standing delivery decision:** D-000.4 requires focused unit/contract/component tests, typecheck, lint, production/offline build, Go tests/vet/native and js/wasm builds, format/diff checks, and compiled e2e source on this story. Playwright execution and the four-target hash matrix remain written/compiled but not run until the Epic 5 boundary after Story 5.12. This is not a D-000.4 matrix override: it consumes the existing production renderer; it must prove native/wasm exact-byte parity in focused tests instead of changing renderer determinism.

## In plain terms

Preview is not a browser interpretation of the canvas and not a remote service. It is the PDF bytes produced by `folio-go` in the one existing wasm worker, from a fresh canonical serialization of the current engine-owned revision plus raw local data and raw author-edited parameters. `pdfjs-dist` renders only those returned PDF bytes to a controlled local canvas surface.

The browser never parses a `.folio`, computes layout, regenerates a PDF, or substitutes a current-looking preview after its inputs change. The next story owns persistent stale-preview identity/invalidation policy, but this story must already make render correlation and cancellation safe: an older, aborted, failed, or superseded render can neither install PDF bytes nor overwrite the current preview state.

## Story

**As a** template author,  
**I want** to see the real PDF my developer's service will produce,  
**So that** I am the first person to see the document rather than the last.

## Acceptance criteria

### AC1 — Preview is a full-surface exact-production mode (FR34, FR35; UX-DR9/15/23)

**Given** a loaded workspace,  
**When** the author activates Preview through the document-bar mode switch or its keyboard shortcut,  
**Then** the canvas work surface is replaced by the PDF viewer as a mode switch, not added as a panel, iframe, embed, browser-native PDF control, new route, account feature, or cloud-looking destination.

**And** the active Preview chip uses the existing cyan authority grammar; Design remains visibly approximate and Preview is labelled as the exact local production PDF without a tutorial or any claim of server rendering.

**And** selection and design-canvas scroll/viewport state are retained while switching modes where meaningful. Viewer page/scroll state is independently retained while Preview remains current. Returning to Design restores the design surface without a command, revision, dirty-baseline, serialized-byte, or selection change.

**And** the Preview surface is responsive as a viewer: it scrolls/pans within the central work surface at the desktop minimum, preserves readable controls and visible focus, and never introduces a responsive document layout or changes canonical page geometry. The existing precision-shell rails may remain fixed/overflow rather than collapsing into a mobile editor.

**Proof / red proof.** Component/e2e-source tests prove that Preview replaces rather than coexists with canvas paint; focusable mode controls have accessible names and cyan visible focus; mode-only changes send no engine command and preserve revision/dirty/selection. A source guard rejects `<iframe>`, `<embed>`, object/browser viewer usage, browser print/PDF navigation APIs as the viewer, and “server/cloud/account/upload/sync” Preview copy or request paths.

### AC2 — one worker renders exactly the bytes production would render (AD-16; NFR5)

**Given** the author requests Preview,  
**When** the render is admitted by the existing FIFO `EngineClient`/worker channel,  
**Then** the worker first obtains a fresh copy of its authoritative canonical serialization for the current revision and passes those exact opaque `.folio` bytes to a render-only Go wasm operation.

**And** that operation takes exactly three separate opaque byte inputs: canonical template bytes, raw sample-data JSON bytes, and the raw editable parameter JSON bytes. It parses the template through the production public boundary and calls the production `folio.Render(template, folio.Data(data), folio.Params(params), fonts.Shipped())` path (or one shared, directly parity-tested production seam); it returns the resulting PDF bytes and diagnostics as bounded transport data.

**And** it never receives a live `*Template`, TypeScript `.folio` model, canvas projection, browser measurements, raster, DOM, URL, file handle, network response, or a caller-provided font set. TypeScript must not parse, inspect, normalize, merge, stringify, or synthesize the `.folio` bytes. The raw 5.10 parameter editor is deliberately a JSON buffer; Story 6.3 alone upgrades it into a parameter-aware UI. It must not hard-code `{}`.

**And** no second worker, wasm instance, `Render` implementation, PDF library encoder, or Go renderer is introduced. The existing single worker may queue render after commands/serialization; a 50-page render may occupy that worker, but only Preview's surface is blocking and the design UI remains usable to switch back or continue local interaction.

**And** the response carries the exact production PDF byte path: copied PDF bytes plus an engine-computed SHA-256 digest of those exact bytes, correlated to the canonical serialization revision used for the call. The UI treats both as opaque display/result evidence; it does not recompute a competing production hash. The same template/data/params/font inputs produce byte-for-byte identical PDF output and digest from the wasm request and a native `folio-go` render. Add discriminating fixtures covering an empty/simple page and a multi-page/text/font case; compare full bytes/digest, not extracted PDF text or a screenshot. Do not re-record a golden merely to make this pass.

**Proof / red proof.** Go/wasm tests prove `render` is closed, byte-only, has three typed inputs, rejects missing/malformed/oversized input, and agrees byte-for-byte and digest-for-digest with native `Render`. Protocol tests reject an operation that accepts no three-part envelope, a live-document/snapshot/viewport field, surplus keys, unsafe bounds, a response without correlation, PDF bytes beyond the explicit limit, a malformed/mismatched digest, or diagnostics outside the existing bounded diagnostic shape. A mutation to use template bytes from an earlier revision, `CanvasProjection`, a second `NewEngine`, or any non-production PDF routine must make a guard/parity test red.

### AC3 — the viewer renders only opaque returned PDF bytes with local PDF.js assets

**Given** a successful correlated render response,  
**When** Preview displays it,  
**Then** a controlled `pdfjs-dist` canvas viewer receives a copied `Uint8Array` of those returned opaque PDF bytes and renders pages locally. PDF.js is a viewer only: it has no template/data/parameter/engine authority and never creates, modifies, or hashes the production PDF.

**And** use the approved `pdfjs-dist` 6.2.x package line named in the decision log, with its worker/standard-font/cmap assets resolved from local, immutable build output and included in the service-worker release manifest. No CDN, dynamic import from an external origin, `eval`, remote font/CMap fetch, or fallback to the browser PDF viewer is permitted.

**And** add the package's Apache-2.0 license/NOTICE material to the repository's license/notice mechanism and extend the existing AD-26 dependency/license guard so the declared package, lockfile resolution, emitted local assets, and attribution agree. Do not change or remove Folio's shipped Noto-font NOTICE files.

**And** object URLs, `PDFDocumentLoadingTask`, page render tasks, worker/document proxies, canvases, and copied byte buffers have explicit ownership. On replacement, mode exit, abort, unmount, viewer error, or failed render: cancel/destroy outstanding PDF.js work, revoke only URLs created by this viewer, clear references, and never revoke an unrelated file/download URL. Prefer direct bytes where PDF.js permits; any object URL must be single-owner and revoked exactly once after it is no longer reachable.

**Proof / red proof.** Unit/component tests spy on PDF.js adapter seams and URL lifecycle for success, replacement, cancellation, failure, unmount, and repeated Preview entry; assert no URL leak, double revoke, stale canvas, or render after destroy. Static/offline tests prove all PDF.js runtime assets are in the immutable release set and forbid http(s), CDN, iframe/embed/object viewer APIs, remote CMap/font configuration, fetch/XHR/WebSocket/EventSource outside existing offline boot paths, and browser-storage/cloud APIs in preview code. Red-prove one missing PDF.js asset/NOTICE, a CDN URL, an unreleased object URL, and an iframe fallback.

### AC4 — rendering has correlated, cancellable, failure-honest state

**Given** Preview begins rendering a document that may be 50 pages,  
**When** the render is pending,  
**Then** only the Preview surface shows a determinate page/progress count when the engine/viewer can honestly provide one, otherwise an explicitly indeterminate numeric-free rendering state; it exposes a keyboard-operable Cancel/Return to Design control and does not disable the design canvas, file controls, or committed editing solely because rendering is pending.

**And** every render has a monotonically unique UI request token plus the worker FIFO request id/revision captured at serialization. Completion may install PDF bytes only if its token is still current, it was not aborted, the mode remains Preview, and the snapshot/revision it represents is still the current authoritative revision. Later response, cancellation, worker failure, PDF.js failure, component unmount, and return-to-Design each invalidate the token before cleanup.

**And** cancellation is silent/no-error for the document state and marks no dirty state; a real render failure remains contained in Preview, gives a terse technical local-error message, preserves the previous design document and its ability to return to Design, and does not imply retrying a server. Diagnostics returned with successful PDF bytes are retained as data for Story 5.12's presentation work; do not implement a diagnostic overlay early.

**And** this story does not present a stale result as current. Until Story 5.11 installs the full AD-18 identity key, render admission/install must at minimum reject any response whose captured canonical revision is no longer current, clear the exact/current marker on a committed command, and never repaint older PDF bytes over a newer result. Story 5.11 extends this to data, params, engine version, and `FontSet` identity; do not duplicate its final identity algorithm in UI code.

**Proof / red proof.** Deterministic deferred-promise tests cover two overlapping renders, command while render waits, return-to-Design, abort before worker settlement, worker failure, PDF.js failure, unmount, and a 50-page progress sequence. Assert the old bytes/canvas never install, no completion changes canonical snapshot/revision/dirty baseline, and a later valid response wins. A red proof that removes token/revision checking or abort cleanup must show a stale PDF install and fail.

### AC5 — local-only security, accessibility, and neutrality remain true

**Given** Preview and the raw parameter editor,  
**When** an author uses keyboard, assistive technology, or an offline browser,  
**Then** Preview/Design, return/cancel/retry and page navigation/zoom controls are reachable, named, and visibly focused; progress/status changes are announced without repeatedly reading every page canvas; the PDF canvas has an accurate accessible label and no false claim that PDF content has become a semantic document. Preserve the current keyboard/accessibility floor; Story 5.12 owns its broader driven-interface work.

**And** all template/data/parameter/PDF bytes remain process-local to the browser and the one local worker. Preview adds no telemetry, analytics, remote diagnostics, upload, account, OPFS/IndexedDB copy, file persistence, or egress. It works after the existing verified offline cache is ready.

**And** opening/closing Preview, typing an uncommitted parameter draft, viewer scrolling/zoom, progress, cancellation, PDF.js cache, URL lifecycle, and render failure are transient UI state. They neither send a document command nor alter canonical bytes, engine revision, saved revision, dirty indicator, file target, selection, or save/open behavior. Only accepted committed engine commands change the document; they must invalidate active preview work before derived state is refreshed.

**Proof / red proof.** A11y tests verify control names/focus/live status and no line/page canvas tab trap. Contract tests verify no network/storage/command path and that every transient preview action has zero command/revision/dirty effect. Offline-release verification must prove PDF.js assets are cache-listed, immutable, Brotli sidecar-covered, and integrity-bound. Mutate Preview to call `fetch`/storage, mark dirty on mode/render state, or leave an unlabelled control; each must red.

## Tasks / subtasks

- [x] **1. Add the closed production-render wasm seam** (AC: 2, 4)
  - Extend `folio-go/wasm/engine.go` and `folio-go/wasm/cmd/engine/main.go` with a render-only byte request/result; reuse `ParseTemplate`, `Render`, `fonts.Shipped`, `Data`, and `Params` rather than duplicating production logic.
  - Define bounded render response bytes/progress/diagnostics, preserve canonical session bytes and revision, and expose only an immutable response projection.
  - Add native-vs-wasm byte-parity fixtures/tests plus malformed/bounds/correlation tests.

- [x] **2. Extend the existing closed worker protocol, not its ownership model** (AC: 2, 4, 5)
  - Add the render operation/envelope and exact validator changes to `folio-designer/src/engine-protocol.ts`, admission/queue/client/worker seams and their tests. Use copied `ArrayBuffer`s, strict allowed keys, FIFO correlation, abort signal, and explicit byte/list bounds.
  - Keep one construction site in `engine-client.ts`; no React direct wasm call or browser-visible template schema.

- [x] **3. Build a dedicated local PDF.js viewer boundary** (AC: 1, 3, 4, 5)
  - Add a narrow `folio-designer/src/preview/` adapter/component/state module with ownership-tested PDF.js loading/render/teardown. Keep PDF.js types/configuration at that edge.
  - Make `App.tsx` compose Design and Preview as exclusive work-surface modes, preserve meaningful selection/scroll, and retain existing local-file/dirty/canvas interactions.
  - Add raw parameter-buffer editing local to Preview. Its bytes are explicit render input and are not a `.folio` model, persisted data, or Story 6.3's parameter-aware UI.

- [x] **4. Package and prove offline/local-only delivery** (AC: 3, 5)
  - Add the approved dependency and lockfile, NOTICE/license attribution, local worker/support assets, generated runtime asset references, release-manifest/service-worker integration, and strict verification/red proofs.
  - Audit the new dependency under AD-26 before adding it; do not modify the existing shipped-font asset provenance or notices.

- [x] **5. Verify honestly under D-000.4** (AC: 1–5)
  - Run designer focused tests, `npm run typecheck`, `npm run lint`, `npm run build`, offline release/witness/red verification, and `npm run test:e2e:compile`; Go focused tests, `go vet ./...`, native build, `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine`, existing lint/hashmatrix static gates, `gofmt`, and `git diff --check`.
  - Execute and restore each named red proof. Record exact counts and the sanctioned P6g corpus-floor red if it persists. Do not call compiled Playwright/e2e or the cross-target matrix green; both are due at Epic 5 boundary.

## Developer guardrails

1. **Canonical bytes are the only render template input.** Obtain them from the same worker serialization/save path at admission time. Never render `Engine.template`, a snapshot, a TS object, saved file input that differs from canonical bytes, or a cached prior serialization.
2. **One engine, one renderer of PDF bytes.** The new wasm call must invoke the existing production `Render` path with `fonts.Shipped()`. PDF.js rasterizes returned bytes only; it cannot become a fallback generator.
3. **Three inputs are real, separate, opaque byte channels.** Data and Params retain exact-decimal JSON semantics in Go. Keep raw 5.10 parameter editing narrow; no hard-coded `{}`, auto-discovered parameter form, local persistence, or data/template merge.
4. **No stale-install loophole.** Check a render token and captured revision at every async boundary. Abort/invalidate before cleanup. A late worker/PDF.js callback must be a no-op and must not alter dirty, revision, selection, saved baseline, or viewer state.
5. **PDF.js has a single resource owner.** Centralize loading/rendering/cleanup; copy inputs; destroy loading/render tasks/proxies; revoke only owned URLs once. Do not use a browser viewer, iframe, embed, object, data URL, or CDN.
6. **Offline/no-egress is functional, not copy.** All PDF.js support assets must be emitted and precached like wasm/fonts. No preview-time remote request, telemetry, storage, upload, account, or server wording.
7. **Preserve prior work.** Existing singleton worker, FIFO protocol, Go-owned command history, opaque file adapter, canonical dirty baseline, engine-measured canvas, offline startup, visual token grammar, and local selection/focus behavior remain authoritative.
8. **Keep future boundaries intact.** Story 5.11 owns the final five-input engine-side preview identity and stale presentation policy; Story 5.12 owns diagnostic overlays/locate-back and the broad keyboard interface; Story 6.3 owns parameter-aware UX. Provide their seams; do not pre-complete their features.
9. **Preserve unrelated work.** Do not modify `.agents/`, `_bmad` config/manifest churn, planning research, fixture/golden/font source files, or unrelated uncommitted changes.

## Project structure notes

- Existing update files: `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go`, `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, `App.tsx`, `App.css`, `package.json`, generated offline asset/release scripts, and adjacent tests/e2e source.
- New preview code belongs under `folio-designer/src/preview/`, with narrow PDF.js adapter tests beside it. Keep application composition in `App.tsx`; keep wasm ownership in the existing worker/Go seams.
- `folio-go/fonts/fonts.go` is the shipped `FontSet` source and its font notices are existing immutable provenance. Use it; do not copy font programs into designer preview code.
- The current `CanvasProjection` is display-only and must not become a render input. The current `serialize` operation returns canonical copied bytes and is the only permitted template source.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5; Stories 5.1–5.12; Story 5.10 ACs; Story 5.11 stale boundary.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-15 through AD-21, especially AD-16/18/19/26.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — S5 Preview, state patterns, determinism obligations, mode switch, keyboard/accessibility floor.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` — mode-switch, cyan authority, progress, focus, responsive token grammar.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR34–36 and local/offline PDF requirements.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.4; 5.10 raw parameter ruling; `pdfjs-dist` 6.2.x + Apache notice decision.
- `_bmad-output/implementation-artifacts/5-1-the-design-system-and-application-shell.md` through `5-9-a-canvas-the-browser-never-measures.md` — shipped shell, worker, offline, startup, local files, canvas, component/property, and browser-authority constraints.
- `folio-go/render_entry.go`, `folio-go/fonts/fonts.go`, `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go` — actual production render and current wasm boundary.
- `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, `App.tsx`, offline release scripts — existing protocol, composition, and offline asset seams.

## Delivery log

### Story creation — 2026-08-28

- Created from the complete Epic 5/Story 5.10 source, PRD, architecture spine, UX DESIGN/EXPERIENCE, D-000.4 and the decision log's raw-parameter/PDF.js rulings, current sprint tracker, recent Story 5.5–5.9 implementation records, current designer worker/protocol/app/offline source, and Go render/wasm/font seams.
- Confirmed current `serialize` returns copied canonical worker bytes; the existing worker is singleton/FIFO; Go `Render` already takes typed raw `Data` and `Params` plus `FontSet`; and the current UI has only a disabled “PREVIEW · later” marker. The story therefore extends one closed byte protocol and replaces that marker with an exclusive viewer mode.
- Confirmed Story 5.10 must ship the raw editable parameter buffer despite Story 6.3 following later, while final five-input preview identity/stale UX stays owned by Story 5.11.
- Implementation has not started. The developer must record real commands/results, red proofs, deferred D-000.4 suites, completion notes, and actual file list; do not pre-check tasks or claim test execution.

## Dev Agent Record

### Agent Model Used

GPT-5.6 (direct implementation; no subagent requested).

### Debug Log References

- `go test ./wasm/...` — PASS.
- `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine` — PASS (the generated `folio-go/engine` artifact was removed).
- `go test ./...` — expected repository red: `internal/text.TestCorpusMeetsP6ExerciseFloors/P6g`, 7 opaque names where the standing floor is 20; all other packages passed.
- `go vet ./...` and native `go build ./...` — PASS.
- `npm run typecheck`, `npm test` (19 files / 95 tests), `npm run lint` (two pre-existing Fast Refresh warnings), and `npm run test:e2e:compile` — PASS.
- `GOPROXY=off go test -count=1 ./...` in `lint` — PASS after regenerating `lint/MANIFEST.md` for the locked `pdfjs-dist` graph.
- `npm run build` — PASS. It emits the local PDF.js worker, four CMap assets, four standard-font assets, immutable sidecars, release manifest, and service worker.
- `npm run verify:offline`, `npm run verify:offline:red`, `npm run verify:offline:wasm` — PASS. The red corpus includes missing PDF.js worker/CMap/standard-font assets.

### Completion Notes List

- Added a closed `render` request: three bounded opaque byte inputs, current-canonical revision enforcement, production `ParseTemplate` + `Render(..., fonts.Shipped())`, PDF SHA-256 evidence, and bounded diagnostics.
- Added FIFO-correlated render protocol/client/worker transport with copied buffers and stale/cancel-safe Preview admission in the application shell.
- Added an exclusive local Preview work surface, raw transient data/parameter buffers, cancel/return/retry, keyboard mode shortcut, accessible controls, and a narrow PDF.js canvas owner that destroys tasks/proxies/canvases on replacement or unmount.
- Added locked `pdfjs-dist` 6.2.108, its local worker plus generated CMap/standard-font runtime asset graph, Apache attribution, and the regenerated AD-26 manifest.
- Focused lifecycle tests cover copied direct bytes (no object URL), replacement/unmount cleanup, cancellation, late-document suppression, PDF.js error containment, and stale render suppression after return to Design.
- Ready for independent review. Playwright execution and the four-target hash matrix remain deliberately deferred to the Epic 5 boundary under D-000.4.

### File List

#### Expected implementation scope

- `_bmad-output/implementation-artifacts/5-10-preview-the-exact-production-document.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go`, and focused render/wasm tests
- `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, adjacent tests
- `folio-designer/src/preview/` (new viewer/adapter/state modules and tests)
- `folio-designer/src/App.tsx`, `App.css`, `App.test.tsx`, relevant e2e source
- `folio-designer/package.json`, lockfile, generated offline asset/release verification only as required for local PDF.js assets and NOTICE/license compliance

#### Actual implementation scope

- `folio-go/wasm/engine.go`, `folio-go/wasm/engine_test.go`, `folio-go/wasm/cmd/engine/main.go`
- `folio-designer/src/engine-protocol.ts`, `engine-protocol.test.ts`, `engine-client.ts`, `engine.worker.ts`, `transient-interaction.test.ts`
- `folio-designer/src/preview/pdf-viewer.tsx`, `folio-designer/src/preview/pdfjs-dist.d.ts`
- `folio-designer/src/App.tsx`, `App.css`, `App.test.tsx`, `scripts/build-wasm.mjs`, `src/generated/pdfjs-assets.ts`
- `folio-designer/package.json`, `package-lock.json`, `third-party-notices/pdfjs-dist/NOTICE`, `lint/MANIFEST.md`

### Finisher verification — 2026-08-28

- Closed all nine persisted review findings: zero diagnostics are `[]`; PDF.js receives local packed-CMap and standard-font bases; late proxies are destroyed; viewer callbacks are token-correlated; Preview e2e source is current; parity covers simple plus five-page multi-script/text/font inputs; transport limits are raw-byte preallocation guards; current viewer state persists through a Design round-trip; and Apache/CMap/Liberation license material is redistributed and guarded.
- Passing gates: designer `npm test` (21 files, 100 tests), `npm run typecheck`, `npm run test:e2e:compile`, `npm run lint` (warnings only), `npm run build`, native `go test ./wasm`, `go vet ./...`, `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine`, and lint-module `go test ./...` (117 tests). Offline release and PDF.js support-asset red proofs pass as part of `npm run build`.
- D-000.4: compiled Playwright and the four-target matrix remain deliberately deferred to the Epic 5 boundary. The complete native Go suite was run and has only the sanctioned standing P6g corpus-floor red: 7 opaque names versus the 20 floor; all other 1,222 tests passed (2 failures, 4 skipped).

## Completion status

Story 5.10 is complete and ready for review.

## Completion status

Ultimate context engine analysis completed — comprehensive developer guide created. Status: `in-progress`.

### Review Findings

- [x] [Review][Patch][High] Preserve empty diagnostics in the wasm response. `boundedDiagnostics` now encodes the zero case as `[]` on the closed render response. [folio-go/wasm/cmd/engine/main.go]
- [x] [Review][Patch][High] Wire the emitted PDF.js CMap and standard-font URLs into `getDocument`. [folio-designer/src/preview/pdf-viewer.tsx]
- [x] [Review][Patch][High] Destroy a document proxy that resolves after cancellation/unmount. [folio-designer/src/preview/pdf-viewer.tsx]
- [x] [Review][Patch][High] Correlate PDF.js callbacks with the active preview token. [folio-designer/src/App.tsx]
- [x] [Review][Patch][High] Replace the stale Epic 5 e2e assertion and add compiled Story 5.10 Preview coverage. [folio-designer/e2e/application-shell.spec.ts]
- [x] [Review][Patch][High] Add the required discriminating parity/protocol/lifecycle mutations. [folio-go/wasm/engine_test.go, folio-designer/src/engine-protocol.test.ts]
- [x] [Review][Patch][Medium] Enforce transport limits at the producer boundary and in raw-byte units. [folio-go/wasm/cmd/engine/main.go, folio-designer/src/engine.worker.ts]
- [x] [Review][Patch][Medium] Retain viewer page/zoom/scroll state while an exact Preview remains current. [folio-designer/src/preview/pdf-viewer.tsx, folio-designer/src/App.tsx]
- [x] [Review][Patch][High] Redistribute and guard the actual license materials for copied PDF.js support assets. [folio-designer/third-party-notices/pdfjs-dist/]
