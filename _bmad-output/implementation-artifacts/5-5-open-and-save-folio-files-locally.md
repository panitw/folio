---
baseline_commit: fd47e6d
story_key: 5-5-open-and-save-folio-files-locally
status: done
created: 2026-08-28
---

# Story 5.5: Open and save `.folio` files locally

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Story key:** `5-5-open-and-save-folio-files-locally`  
**Status:** `done`  
**Covers:** FR8 · AD-9, AD-15, AD-20 · UX-DR13, UX-DR16 · CAP-4 / S5

**Standing delivery decisions:** numeric order; terminal decision channel; continuous run; unit/lint/build on every story. Under D-000.4, browser/e2e execution and the four-target hash matrix are Epic 5 boundary gates, not per-story green claims. This is not a hash-shaped override: it changes browser-local file transport, not canonical rendering. Every Delivery Log must name written/compiled-but-unrun suites by name.

## In plain terms

Folio is a local-file tool, not a hosted editor. This story turns the shell's honest placeholders into real Open and Save controls: a Chromium-class browser keeps an approved file handle and writes the same file; Firefox/Safari use a normal file chooser to open and a download to save. The choice happens once when the app starts. Everything after that calls one file-access interface and must not scatter browser-capability branches through React.

The browser never interprets a `.folio` document. It supplies opaque bytes to the one Go/wasm worker, which parses, validates, owns, and serializes canonical bytes. Saving writes exactly those serializer bytes. TypeScript may track file identity, a save target/handle, errors, and a saved revision/byte baseline; it must never define the `.folio` schema, parse its JSON, repair it, or serialize it.

There is no cloud, account, recent-files list, OPFS document copy, IndexedDB/localStorage template cache, autosave, upload, telemetry, or server round trip. A canceled picker is a no-op, not an error. A failed or denied write is surfaced tersely and leaves the current document dirty. An unnamed local starter remains a usable blank workspace with both **Open** and **Start blank**, never a dead canvas.

## Story

**As a** template author,  
**I want** to work on files on my own laptop with no account,  
**So that** my templates are mine and live in my own Git repository.

## Source-grounded acceptance criteria

### AC1 — one local-only file-access adapter is selected once

**Given** application startup  
**When** browser file capability is detected  
**Then** the composition root selects exactly one `FileAccess` implementation once for that browser session: the File System Access tier only when the needed picker/write capabilities are present, otherwise the download tier  
**And** the rest of the application depends only on the shared interface and never rechecks capabilities or branches by browser/vendor  
**And** adapter methods are asynchronous and are invoked from user gestures where browser activation requires it.

**Given** a browser implementing the File System Access API  
**When** the author opens a `.folio` and later saves it  
**Then** the chosen adapter retains the approved `FileSystemFileHandle` only in memory for this session and writes through that handle in place  
**And** it does not create a second browser-storage copy or silently choose another path.

**Given** Firefox or Safari, which do not provide this API for Folio's use  
**When** the author opens and later saves a `.folio`  
**Then** open uses an accept-filtered `<input type="file">` and save creates a local `.folio` download  
**And** the fallback does not claim it overwrote the originally selected file or held permission it cannot hold.

**Proof / red proof.** Unit-test capability selection with complete/incomplete API seams and prove it is evaluated once. Add structural coverage with nonzero production-source witnesses that rejects capability tests outside the adapter/composition seam. Red-prove a second capability branch and a fallback that calls a File System Access API.

### AC2 — opening is Go-owned parsing/validation and preserves honest identity

**Given** the author invokes Open in either tier  
**When** they choose one `.folio` file  
**Then** the adapter returns only its bytes and a display-safe basename/identity record; the engine client's existing `load` operation receives the bytes  
**And** the Go/wasm engine alone parses, validates, canonicalizes, and returns the immutable snapshot/diagnostic result  
**And** TypeScript does not parse JSON, validate fields, construct a document, or implement a `.folio` schema.

**Given** load succeeds  
**When** the workspace is displayed  
**Then** its title is the selected file's basename (not a path), its identity is the current adapter target, and its saved baseline is the successfully loaded engine revision/canonical serialization  
**And** a subsequently committed engine mutation makes the persistent unsaved indicator dirty until a successful save establishes a new baseline.

**Given** the author cancels an Open picker  
**When** the picker resolves/rejects with its cancellation signal  
**Then** the current document, identity/title, saved baseline, and dirty state remain unchanged and no alert/error is announced  
**And** any other read/decode/load/validation failure is announced with the bounded engine/adapter diagnostic, preserves the prior session, and does not manufacture a replacement document.

**Proof / red proof.** Test success, malformed/unsupported bytes, explicit cancellation, same-file reselection in the input fallback, and an open failure after a dirty document exists. Prove a failed open cannot change title, identity, snapshot, or dirty baseline. The ownership guard from Story 5.2 must reject a representative TypeScript document interface/JSON parse/save path while allowing file-adapter metadata only.

### AC3 — Save and Save As have explicit, non-fictional target semantics

**Given** a document has a current File System Access handle  
**When** Save is invoked  
**Then** the app asks the engine client for `serialize` bytes, creates a writable on that current handle, writes those bytes, and closes it successfully before clearing dirty state  
**And** the visible title/identity remain that file's basename/current handle.

**Given** a document has no writable in-place target — an unnamed starter, a fallback-tier open, or a prior target that cannot be written  
**When** Save is invoked  
**Then** it follows the selected adapter's explicit save-target route: File System Access prompts for a target; fallback downloads a `.folio` file  
**And** it never reports that the original file was overwritten.

**Given** Save As is invoked  
**When** a target is selected or a download is issued  
**Then** it always obtains fresh engine `serialize` bytes and uses a new File System Access target or a named fallback download  
**And** after a successful File System Access Save As the new handle/basename becomes the current identity; after a fallback download the UI identifies the saved/downloaded filename without claiming a retained handle  
**And** Save As must not mutate document content, revision, or canonical bytes.

**Given** saving is canceled, permission is denied/revoked, a write/close fails, quota/media/lock errors occur, or serialization fails  
**When** the operation settles  
**Then** cancellation leaves state unchanged without an error; other failures present a concise, accessible local-save error, retain the prior identity, and leave dirty state true  
**And** no partial-success claim is made before `close()`/download creation succeeds.

**Proof / red proof.** Use adapter/worker seams to prove operation order is `serialize → write → close → clean`; mutate/throw each stage and verify the dirty baseline remains. Cover Save versus Save As identity transitions, fallback download naming, cancellation, denied permission, stale/revoked handles, write/close errors, and duplicate save clicks (one in-flight save per document). A test double may stand in only for picker/handle/download mechanics; it must consume opaque engine bytes and never synthesize template bytes.

### AC4 — byte fidelity and determinism stay with the engine

**Given** a file is opened, changed through a committed engine command, and saved  
**When** bytes handed to the adapter are inspected  
**Then** they are byte-for-byte the result of the Go engine's current `serialize` response  
**And** no TypeScript JSON stringify, formatting, key ordering, newline conversion, schema migration, metadata injection, or blob text conversion can alter them.

**Given** canonical `.folio` bytes  
**When** they are opened and immediately saved through either adapter seam  
**Then** the bytes handed to the target equal the engine serialization exactly  
**And** the canonical guarantees remain those of AD-9: sorted keys, two-space indentation, LF endings, no trailing whitespace, and one trailing newline.

**Given** noncanonical but valid input  
**When** it is opened  
**Then** normalization is solely the Go parser/serializer's documented behavior; the UI neither compares/parses text nor hides a resulting canonical save as an in-place byte-preserving no-op.

**Proof / red proof.** Add a real js/wasm-host or controlled EngineClient integration test using fixture bytes that captures adapter write/download bytes and compares them exactly to `serialize`. Red-prove a UI `JSON.stringify`, CRLF/text conversion, and a stale/pre-command byte cache; each must fail a structural or behavioral guard. This is byte-fidelity evidence, not a reason to run the full cross-target rendering matrix under D-000.4.

### AC5 — the workspace tells the truth about blank, dirty, local, and offline use

**Given** the application has no user-opened template  
**When** the Workspace is displayed  
**Then** it is titled as an unnamed blank template and provides working, keyboard-operable **Open** and **Start blank** actions  
**And** Start blank uses only Go-owned starter/blank bytes through the worker; it creates no TypeScript template object and leaves the new document unsaved.

**Given** a document is clean or dirty  
**When** the document bar/status surface is displayed  
**Then** a persistent, quiet, always-visible text-backed unsaved indicator communicates the state in both file tiers; the amber dot may support it but cannot be the only semantic signal  
**And** there is no autosave in either tier.

**Given** Open, Save, or Save As is available  
**When** it is reached by pointer or keyboard  
**Then** it has a visible cyan focus treatment, an accessible name, and concise local-file wording; frequent Save has its documented shortcut and only prevents the browser default while Folio can handle it  
**And** busy/error/success status changes are announced without relying on colour or a decorative icon.

**Given** a completed first-load cache and a disconnected network  
**When** the author opens or saves a local file  
**Then** the interaction uses only browser-local picker/handle/download and the existing local worker/cache path, with no network request, remote save, account, sync, recent-file service, or cloud wording/control.

**Proof / red proof.** Component tests cover names, keyboard activation/focus, shortcut behavior, text status/dirty transitions, unnamed blank Open/Start blank routes, and no-autosave. Compile the existing browser specs and add focused source for local open/save/offline coverage; do not claim Playwright execution unless it actually ran. Structural/policy tests must reject account, sync, collaborator, share, cloud-save, recent-file, OPFS-document-storage, localStorage/IndexedDB template persistence, and template/data upload routes with nonzero allowed local-file witnesses.

## Tasks / subtasks

- [x] **1. Define the narrow two-tier file boundary under `folio-designer/src/file/`** (AC: 1, 3)
  - [x] Create a capability-selected-once `FileAccess` interface plus File System Access and input/download implementations; keep picker, permission, handle, and download details out of `App.tsx`.
  - [x] Model only transport metadata: opaque `ArrayBuffer` bytes, display basename, current-target capability, cancellation versus bounded failure, and optional in-memory handle. Do not add `.folio` field types, JSON helpers, OPFS, or durable browser storage.
  - [x] Define explicit Save and Save As target rules, single-flight saving, accept/filter and suggested `.folio` names, and testable write/close completion semantics.

- [x] **2. Integrate file actions with the sole engine client and session state** (AC: 2–4)
  - [x] Reuse `EngineClient.request('load', bytes)` and `request('serialize')`; extend only the existing protocol/Go host if a safe opaque command is genuinely missing. Do not route bytes around the worker.
  - [x] Replace the shell's disabled fake Open/Save controls with real actions, title/identity state, error/busy feedback, and saved-revision/bytes baseline semantics.
  - [x] Preserve the current worker singleton, immutable snapshot discipline, request correlation, byte-copy ownership, and failure lifecycle from Story 5.2. A file action must not make a second worker/wasm engine.

- [x] **3. Implement the governed empty/dirty/accessibility UX** (AC: 5)
  - [x] Treat the existing Go-owned starter asset as an unnamed blank workspace until a real local file is opened; make Start blank explicit and keep it unsaved.
  - [x] Provide quiet, persistent textual clean/unsaved state; render local errors with the EXPERIENCE terse technical voice, accessible announcement, and visible focus.
  - [x] Implement keyboard activation and Save shortcut without making unsupported actions silently no-op; preserve the governed SVG/token styling and document-bar landmarks.

- [x] **4. Prove capability, ownership, local-only behavior, and byte fidelity** (AC: 1–5)
  - [x] Add adapter, component, and engine-client tests for both tiers, cancellation, permissions, stale handles, write/close errors, Save/Save As transitions, dirty baselines, duplicate saves, and offline-local interaction.
  - [x] Add non-vacuous structural guards for one capability decision, no TS `.folio` schema/serialization, no file-access bypass, and prohibited cloud/recent-file/browser-storage fiction. Include red mutants and witnesses.
  - [x] Capture actual bytes at the adapter boundary and compare with worker serialization; retain the Go format round-trip tests as the canonical parser/serializer proof.

- [x] **5. Verify at the correct D-000.4 cadence and record reality** (AC: 1–5)
  - [x] Run focused designer unit/component/contract tests, `npm run typecheck`, `npm run lint`, `npm run build`, relevant Go/js-wasm tests/build/vet, and `npm run test:e2e:compile`.
  - [x] If a real browser-only picker/permission behavior cannot be discriminated through the seams and compilation, run the smallest relevant browser integration test, state why it is a D-000.4 override, and record its actual result.
  - [x] Otherwise record browser e2e and the named four-target matrix as written/compiled, not run — due at Epic 5 close. Name every skip/red by test name, never only a count (D-000.81).

## Developer guardrails

### Required architecture and ownership rules

1. **Go owns `.folio` bytes and meaning.** `internal/template` remains the only parser/validator/serializer and AD-9 remains the sole canonicalization contract. File paths/handles never reach `internal/`; browser code passes bytes to `folio.ParseTemplate` only through the existing wasm engine boundary.
2. **No TypeScript schema and no JSON Schema library.** The permitted TS shapes are request/response envelopes, immutable snapshots, UI transients, and file-access metadata. No document field mirror, JSON parse/stringify, canonicalizer, migration, validator, or text serializer is allowed.
3. **One worker, one channel.** Open uses `load`; Save/Save As use `serialize`; both await the existing asynchronous client. Do not direct-call wasm, expose a live document, add a second worker, or create a file-specific `postMessage` channel.
4. **Files remain local.** Handle objects stay in memory only. Do not persist handles or document bytes to OPFS, IndexedDB, Cache Storage, localStorage, sessionStorage, a service worker, a server, or telemetry. Story 5.3's service worker caches only generated static runtime assets, never document state.
5. **Permission/cancellation is state, not a crash.** Treat expected picker abort as silent no-op. Treat denial/revocation/write/close/read errors as bounded local errors; preserve the previous successful engine snapshot and dirty baseline. Never clear dirty until a complete successful save target operation.
6. **Identity is honest.** Display only basename/title, never filesystem path. A File System Access target can be current/save-in-place; the download tier cannot. Do not introduce a recent-files list, "synced", "saved to cloud", or overwrite claim after a fallback download.
7. **Saved baseline is engine-based.** Tie clean/dirty to the last successful engine serialization/revision written to the selected target, not to React object equality, timestamps, file size, or optimistic button clicks. Future committed mutations must make it dirty; transient UI work must not.
8. **Deterministic bytes survive transport.** Use `ArrayBuffer`/`Uint8Array` as opaque binary data; ensure buffer transfer/copy ownership cannot detach or mutate the engine bytes before write. Never use text decoding or Blob-from-string to create a `.folio` save.

### Existing seams and likely files

- `folio-designer/src/engine-client.ts`, `engine-protocol.ts`, and `engine.worker.ts` already provide the singleton asynchronous `load`/`serialize` byte boundary. Reuse them; do not recreate their lifecycle/protocol logic.
- `folio-go/wasm/engine.go` and `folio-go/wasm/cmd/engine/main.go` already parse on load and return copied canonical bytes on serialize. Preserve this Go-owned behavior and coded failure route.
- `folio-designer/src/App.tsx` is the current document-bar consumer: it has deliberately disabled Open/Save placeholders and an unnamed starter label. Replace that honesty placeholder with scoped file UX, not later canvas/preview/binding work.
- New file adapter code belongs in `folio-designer/src/file/`, per the architecture source tree. Tests belong beside it and in the existing Vitest/compiled Playwright conventions.
- `folio-designer/src/main.tsx` is the appropriate composition boundary for one capability decision and one file-access instance; preserve its engine/offline startup behavior.
- `folio-designer/src/App.css`, `tokens.css`, and existing component tests define the UI grammar. Keep existing landmarks, icon names, focus treatment, dark chrome/light page separation, and accessibility semantics.

### Scope boundaries

- Do not implement page setup/grid/canvas/palette/property mutation (Stories 5.6–5.9), preview/pdf.js/stale preview (5.10–5.11), or diagnostics feature completion (5.12).
- Do not settle UX2 (sample JSON persistence) or UX4 (undo across save/reload); neither authorizes schema or persistence changes here.
- Do not modify canonical format rules, Go rendering behavior, fixtures, fonts, cross-target hashes, service-worker release contents, or account/network architecture.
- Do not modify unrelated working-tree churn: `_bmad/_config/files-manifest.csv`, `_bmad/_config/manifest.yaml`, `_bmad/bmm/config.yaml`, `_bmad/core/config.yaml`, `.agents/`, or planning research. Preserve `deferred-work.md` unless implementation discovers a genuinely new deferred item with an explicit owner.

## Test strategy and D-000.4 treatment

Every-story evidence is required: focused adapter/client/component/structural tests; typecheck; lint; production build; relevant js/wasm and Go unit/build/vet tests; and compiled browser source. Tests must make cancellation, fallback behavior, byte capture, and dirty-baseline preservation observable rather than merely checking that a button was clicked.

The designer browser/e2e suite and the four-target matrix (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) remain **deferred, not waived**, until after Story 5.12 and before `epic-5` is marked done. This story introduces no new rendering/cross-target divergence source, so it does not meet D-000.4's override criterion. If a real picker/permission runtime behavior cannot be proved with the required seams, execute only the smallest focused browser integration test and record the exact rationale/result. Do not promote compilation to browser execution or report a suite green without evidence.

## Project structure notes

- Architecture assigns FR8 to `folio-designer/src/file/`; keep the two-tier implementation there and expose one narrow interface to the app composition root.
- Keep browser API types at the adapter edge. Test shims must be explicit and must not leak document structure into TS.
- Generated starter/wasm/font assets remain generated runtime inputs. They are not user-document persistence and must not become a second template store.
- No new dependency is justified for pickers, JSON schema, storage, sync, or downloads. Use browser APIs behind the adapter and audit any unavoidable package through AD-26 before adding it.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5, Story 5.5, FR8, UX-DR13/16, and interaction requirements.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-9, AD-15, AD-16, AD-19, AD-20; source tree and capability map.
- `_bmad-output/specs/spec-folio/SPEC.md`, `folio-format.md`, and `acceptance.md` — local lifecycle, one engine-owned format, S5 browser-save/Go-render handoff, canonical bytes.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — local machine/no account/no server template ownership and FR8/FR11–13.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` and `DESIGN.md` — workspace, empty/unsaved state, explicit local save, shortcuts, focus/accessibility, and cloud-fiction prohibition.
- `_bmad-output/implementation-artifacts/5-1-the-design-system-and-application-shell.md` through `5-4-an-honest-first-run-load-screen.md` — shipped shell, singleton worker, offline/static-only boundary, starter-load behavior, test cadence, and inherited P6g disclosure.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.4, D-000.81, Story 5 sequencing, and the 5.5 AD-20 ownership note.
- `_bmad-output/implementation-artifacts/deferred-work.md` — reviewed for existing owners; no file-access item is newly due.

## Delivery Log

### Story creation — 2026-08-28

- Created from complete Epic 5 requirements; shipped Stories 5.1–5.4 and commits through `fd47e6d`; PRD; architecture spine; SPEC/format/acceptance; UX DESIGN/EXPERIENCE; current worker/shell source; tracker; decision log; and deferred-work register.
- Confirmed the delivered engine seam already owns `load` and `serialize`; this story must compose it with browser-local file transport rather than recreate parsing, validation, serialization, or a document model in TypeScript.
- Confirmed the architecture assigns FR8 to `folio-designer/src/file/`, AD-20 mandates two tiers selected by one startup capability check, and D-000.4 defers browser e2e/four-target matrix to Epic 5 close. No per-story matrix override is justified.
- Implementation has not started. Required every-story gates, exact commands/results, named skips, any focused browser override, completion notes, and completion file list must be recorded by the developer/finisher; do not pre-check tasks or claim a test ran.

### Developer implementation — 2026-08-28

- Added the once-selected `FileAccess` composition boundary. Chromium-class capability supplies an in-memory File System Access handle for in-place Save; the incomplete-capability fallback uses an accept-filtered input and byte download only. No handles, bytes, paths, or document state are persisted.
- Open sends opaque adapter bytes only to `EngineClient.request('load', bytes)` and establishes its clean revision only after `serialize`. Save and Save As serialize immediately before target I/O; File System Access reports success only after `write` then `close`, and a fallback reports a download rather than an overwrite. Cancellation is silent; other local failures leave the session dirty.
- Replaced unavailable controls with keyboard-operable Open, Save, Save As, and Start blank actions, persistent textual dirty state, terse accessible outcomes, and Ctrl/Cmd+S. Start blank reloads already Go-serialized starter bytes through the worker and remains unnamed/unsaved.
- Added adapter/component/ownership/local-only tests and compiled browser source. `file-access.test.ts` captures exact binary write/download bytes, cancellation, same-file input selection, Save/Save As, and denied/stale/write/close errors; structural red proofs reject second capability branches, JSON document ownership, browser storage/cloud state, and uploads.
- Passed: `npm test` (15 files, 55 tests), `npm run typecheck`, `npm run lint`, `npx vite build`, `npm run build:offline`, `npm run verify:offline`, `npm run verify:offline:wasm`, `npm run verify:offline:red`, and `npm run test:e2e:compile`. The composite `npm run build` was attempted but the host’s 30-second command window ended during Brotli generation; every constituent build/verification stage then passed independently.
- Passed Go gates: `go test -skip 'TestCorpusMeetsP6ExerciseFloors' ./...` in `folio-go` (1193), `go vet ./...`, and `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine`; `lint` test/vet/build (117); `hashmatrix` test/vet/build (3). The sanctioned red `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` remains red at 7 < 20; no new red was introduced.
- Not run under D-000.4: Playwright execution for `application-shell.spec.ts`, `engine-worker.spec.ts`, `offline-update.spec.ts`, and `local-file-actions.spec.ts` (all compile-covered); the four-target hash matrix `darwin/arm64`, `linux/amd64`, `linux/arm64`, and `js/wasm`. No focused browser override was warranted because picker, permission, byte-order, and fallback behavior are discriminated through explicit seams.

### Code review — 2026-08-28

- Reviewed the complete Story 5.5 working tree from baseline `fd47e6d` against the story artifact, Epic 5, shipped Stories 5.1–5.4, the architecture/SPEC/PRD/UX contracts, D-000.4, and the worker/offline/file ownership boundaries. Unrelated `_bmad` configuration, `.agents/`, and planning research were excluded.
- Outcome: **1 Blocker · 4 Major · 3 Minor**, all open for the finisher. The blocker is a reproducible stale-save race that leaves the Go engine at a newer revision while React paints an older revision as clean. No finding requires a new product decision; all route to **FIX**.
- Independent gates passed: designer unit/component/contract suite **15 files / 55 tests**; typecheck; lint; full production/offline build; ordinary, wasm-witness, and red-proof offline verification; and Playwright source compilation. Folio Go with the sanctioned floor excluded passed **1193** tests; Go vet and the js/wasm engine build passed. Lint passed test/build/vet (**117**); hashmatrix passed test/build/vet (**3**).
- The ordinary Folio Go suite has only the sanctioned red: **1199 passed, 2 failed, 4 skipped**; the failures are `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (`7 < 20`) and its parent. Named skips: `TestXrefEntriesRejectsMalformedSubprocess`, `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`, `TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`, and `TestTwoTablesWithPageCountFooterRenderConsistently`.
- Reviewer red proofs were temporary and restored: (1) noncanonical input bytes differing from engine serialization still painted **Saved local file**; (2) revision 3 committed during a pending revision 2 save was painted back as revision 2/clean when that save settled; (3) a forbidden second capability branch passed the structural test; (4) an OPFS + `TextDecoder`/CRLF/Blob-string rewrite passed both structural ownership tests; and (5) a post-object-URL download failure leaked the URL. `git diff --check` passed after restoration.
- Deferred exactly under D-000.4, not claimed green: Playwright execution for `application-shell.spec.ts`, `engine-worker.spec.ts`, `offline-update.spec.ts`, and `local-file-actions.spec.ts`; and the four-target matrix (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`). Story 5.5 remains a non-override because it introduces no new cross-target rendering divergence source.

### Finisher — 2026-08-28 (done)

- Resolved all review findings as **FIX**. Save now acquires an activation-gated target before its asynchronous engine serialization, then writes/closes fresh opaque bytes. Save completion never replaces the displayed snapshot; it records a clean revision only when that exact written revision is still current. The controlled pending-save/pending-commit proof retains revision 3 and its unsaved state after revision 2 settles.
- Open compares opaque source and engine serialization bytes without inspecting document text. Canonical-equal input establishes a baseline; valid noncanonical input remains explicitly unsaved until the canonical bytes are written. Cancellation clears temporary busy wording without changing identity, baseline, dirty state, or emitting an error.
- Strengthened red controls now reject a second capability decision in production code, OPFS access, file-upload fetches, byte-to-text/CRLF/Blob-string conversion, and a narrow two-field `.folio` schema mirror. Adapter tests cover activation order, save failure cleanup, exact opaque bytes, write/close failure, and object-URL/anchor cleanup on every exit.
- Replaced the visibility-only browser source with deferred-but-executable native and fallback tests: a real worker opens fixture bytes, fallback Save uses the keyboard and captures a download while offline; native Save As captures opaque worker bytes through an injected picker seam; cancellation preserves the local identity. Playwright execution remains deferred under D-000.4, while this source is compile-checked.
- Added text-backed operation announcements and application `aria-busy`, while retaining the persistent dirty text. The offline generator now replaces a previous generated bootstrap before regeneration and allows its self-referential S1 size to settle, avoiding stale/nested release metadata during repeated normal verification.
- Measured: designer unit/component/contract suite **15 files / 60 tests**; typecheck; lint; production Vite build; offline release generation plus ordinary, wasm-witness, and red-proof verification; e2e source compilation; Go focused suite **1193** passes; Go vet; js/wasm engine build; `lint` test/vet/build (**117**); `hashmatrix` test/vet/build (**3**); gofmt and `git diff --check`. The ordinary Go suite remains intentionally red only at `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (**7 < 20**) and its parent, with four named existing skips.
- Deferred without a green claim: execution of `application-shell.spec.ts`, `engine-worker.spec.ts`, `offline-update.spec.ts`, and the strengthened `local-file-actions.spec.ts`; and the `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` matrix. Both remain D-000.4 Epic 5 boundary work; no focused override is justified.

## Dev Agent Record

### Agent Model Used

GPT-5.6 Sol — direct implementation (subagent dispatch prohibited by task instruction).

### Debug Log References

- `folio-designer/src/file/file-access.test.ts`
- `folio-designer/src/file/file-access-contract.test.ts`
- `folio-designer/src/App.test.tsx`

### Completion Notes List

- [x] Implemented Story 5.5 local-only open/save workflow.
- [x] Verified focused design, offline, Go, lint, and compile gates; retained the sanctioned P6g red disclosure.
- [x] Deferred actual browser picker/e2e execution and target matrix only under D-000.4.

### File List

#### Expected implementation scope

- `_bmad-output/implementation-artifacts/5-5-open-and-save-folio-files-locally.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `folio-designer/src/file/` adapter/interface modules and focused tests
- `folio-designer/src/App.tsx`, `App.css`, and focused component tests
- `folio-designer/src/main.tsx` only for composition-root capability selection/injection
- `folio-designer/scripts/generate-offline-release.mjs` only to make repeated offline verification replace, rather than nest, generated release metadata
- Existing engine client/protocol/worker and Go wasm host tests only if a narrowly necessary opaque-byte seam is absent
- `folio-designer/e2e/` source only for compiled deferred local-file coverage

#### Completion file list

- `_bmad-output/implementation-artifacts/5-5-open-and-save-folio-files-locally.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `folio-designer/e2e/application-shell.spec.ts`
- `folio-designer/e2e/local-file-actions.spec.ts`
- `folio-designer/src/App.css`
- `folio-designer/src/App.test.tsx`
- `folio-designer/src/App.tsx`
- `folio-designer/src/engine-ownership-contract.test.ts`
- `folio-designer/src/main.tsx`
- `folio-designer/scripts/generate-offline-release.mjs`
- `folio-designer/src/startup-sequence.ts`
- `folio-designer/src/file/capability.ts`
- `folio-designer/src/file/file-access.ts`
- `folio-designer/src/file/file-access.test.ts`
- `folio-designer/src/file/file-access-contract.test.ts`
- `folio-designer/src/file/file-system-access.ts`
- `folio-designer/src/file/input-download.ts`

## Completion status

Finished after review remediation. Status: `done`.

## QA Review

### Review Summary

- Outcome: **changes required**; **1 Blocker · 4 Major · 3 Minor** remain open. Status stays `review`; the finisher must resolve every finding and re-run the measured gates before `done`.
- The native File System Access and input/download tiers are composed once and the Go worker remains the sole `.folio` parser/serializer. The review failures are in dirty-baseline truthfulness, async operation ordering, browser activation, verification teeth, fallback cleanup, accessibility, and artifact truthfulness—not in canonical Go ownership or D-000.4 cadence.
- Review-layer fan-out was skipped because the commissioning handoff explicitly prohibited subagents. The reviewer performed the adversarial, edge-case, verification-gap, and acceptance passes directly and ran controlled red mutations.

### Review Findings

- [x] [Review][Resolved][Blocker] **A save that settles after a newer engine commit rolls the UI back and falsely marks the document clean.**
  - **Category:** AC2/AC3 / concurrent operations / stale completion / Story 5.2 one-document invariant
  - **Location:** `folio-designer/src/App.tsx:31-40,61-75,108-119`
  - **Observation:** `fileBusy` disables only file-action buttons; the committed engine action remains enabled. Save captures a serialized snapshot, awaits target I/O, then unconditionally writes that old snapshot and revision into React. The reviewer held a revision 2 save open, committed revision 3, and then released the save: the engine remained at revision 3 while the UI reverted to revision 2 and displayed `Saved local file`.
  - **Impact:** the newest engine-owned mutation is unsaved but Folio says it is saved; what the UI paints, what the engine owns, and what the file contains disagree. A subsequent edit/save begins from a misleading baseline and can conceal data loss.
  - **Required resolution:** **FIX.** Serialize/write completion must establish only the revision actually written; it must never replace a newer UI snapshot or clean a newer engine revision. Gate committed mutations during file operations or make completions revision/operation-token aware, and add a controlled stale-completion test.

- [x] [Review][Resolved][Major] **Opening noncanonical valid bytes immediately claims the file is saved although canonical engine bytes were never written to its target.**
  - **Category:** AC2/AC4 / canonical opaque bytes / dirty-baseline truthfulness
  - **Location:** `folio-designer/src/App.tsx:47-55,108-109`; `folio-designer/src/App.test.tsx:50-60`
  - **Observation:** Open asks Go to load and serialize, but it discards the returned canonical bytes and always sets `savedRevision` to the canonical revision. The reviewer supplied valid input bytes different from the serialize response; the UI painted `Saved local file`. The existing test uses identical three-byte buffers and therefore cannot discriminate this case.
  - **Impact:** the in-place file still contains the noncanonical source while the UI says the canonical engine state is saved. This violates the engine-based baseline rule and hides the normalization that the next Save will perform.
  - **Required resolution:** **FIX.** Compare opaque byte arrays without parsing/interpreting them, or model an equivalent engine-owned saved-byte baseline; mark the open dirty whenever current canonical bytes have not actually been written. Add canonical-equal and canonical-different tests for both tiers.

- [x] [Review][Resolved][Major] **Save As and first Save request the browser picker only after an asynchronous worker round trip, so required transient user activation can expire.**
  - **Category:** AC1/AC3 / native File System Access / permission and activation timing
  - **Location:** `folio-designer/src/App.tsx:61-67,90-95`; `folio-designer/src/file/file-system-access.ts:30-42`
  - **Observation:** the click/shortcut handler awaits `engine.request('serialize')` before `FileSystemAccess.save()` reaches `showSaveFilePicker()`. For Save As or a document without a current handle, the activation-gated picker is therefore not requested at the gesture boundary; any worker queue delay, large serialization, or browser activation expiry turns a valid action into `Could not save local file`.
  - **Impact:** the primary Chromium Save As/unnamed Save route can fail based on timing rather than permission, despite the story explicitly requiring activation-sensitive adapter methods to be invoked from user gestures.
  - **Required resolution:** **FIX.** Acquire a fresh target while transient activation is guaranteed, then serialize fresh bytes and perform write/close; preserve cancellation and identity rules. Add an ordering/activation seam test that fails if picker acquisition is deferred behind engine work.

- [x] [Review][Resolved][Major] **The structural policy tests stay green for realistic forbidden capability, OPFS, and byte-rewrite mutations.**
  - **Category:** AC1/AC4/AC5 / structural guards / red-proof validity
  - **Location:** `folio-designer/src/file/file-access-contract.test.ts:6-35`; `folio-designer/src/engine-ownership-contract.test.ts:11-27,58-78`
  - **Observation:** `extraCapabilityChecks()` reports filenames, so `file-system-access.ts` is already an allowed hit; adding a second `window.showSaveFilePicker` branch inside that file did not change the expected array and passed. `forbiddenFileState()` does not recognize `navigator.storage.getDirectory()`/OPFS. A temporary production module that wrote to OPFS and decoded bytes through `TextDecoder`, converted LF to CRLF, and constructed a Blob from the resulting string passed both contract suites. The schema guard also triggers only when at least three exact field names appear in one declaration.
  - **Impact:** the tests claimed as non-vacuous enforcement do not prevent the exact second capability authority, durable document copy, or TypeScript byte transformation that the story says they red-prove.
  - **Required resolution:** **FIX.** Guard semantic branch/call counts rather than allowed filenames; cover OPFS APIs and opaque-byte-to-text/Blob-string paths; broaden the no-schema proof; and retain each reviewer mutant as a named red control.

- [x] [Review][Resolved][Major] **The emitted Story 5.5 browser test does not exercise open, save, fallback, or offline behavior.**
  - **Category:** AC5 / production composition / deferred e2e source teeth
  - **Location:** `folio-designer/e2e/local-file-actions.spec.ts:3-13`; `_bmad-output/implementation-artifacts/5-5-open-and-save-folio-files-locally.md:233-236`
  - **Observation:** the test name claims local open/save routes “offline,” but the body only loads the page, checks four controls are visible, and focuses Start blank. It never disconnects the context, selects or injects a file, captures a download, invokes Save/Save As, exercises the real worker, or asserts absence of network/template upload. Compiling this source proves only its TypeScript syntax.
  - **Impact:** the Epic 5 D-000.4 browser catch-up run can be green while Story 5.5's browser-local behavior remains entirely broken; the developer record overstates focused offline/open/save coverage.
  - **Required resolution:** **FIX.** Emit executable browser coverage with controlled picker/download seams for both tiers, real worker serialization bytes, keyboard Save, cancellation/error state, and an actually offline local action. Keep actual execution deferred until the Epic 5 gate unless a new override reason emerges.

- [x] [Review][Resolved][Minor] **A fallback download failure after object-URL creation leaks the document Blob for the page lifetime.**
  - **Category:** AC3/AC5 / object URL cleanup / local document privacy
  - **Location:** `folio-designer/src/file/input-download.ts:39-55`; `folio-designer/src/file/file-access.test.ts:66-78`
  - **Observation:** revocation is scheduled only on the success path. If anchor creation, DOM append, or `click()` throws after `createObjectURL()`, the catch converts the error but has no reference cleanup. A reviewer test forcing `click()` to throw observed zero `revokeObjectURL` calls.
  - **Impact:** repeated blocked/failed saves retain opaque local document bytes in Blob storage until reload and leave cleanup behavior unproved.
  - **Required resolution:** **FIX.** Revoke in a failure-safe `finally`/owned cleanup path and remove any appended anchor on every exit; test success and each post-creation failure stage.

- [x] [Review][Resolved][Minor] **Open/Save/Save As busy states are not announced to assistive technology.**
  - **Category:** AC5 / accessibility / async status
  - **Location:** `folio-designer/src/App.tsx:43-45,61-64,78-80,112-119`
  - **Observation:** entering an operation clears the prior status and disables the file buttons, but renders no `Opening…`, `Saving…`, or busy status and exposes no `aria-busy`. Only terminal success/error messages are announced.
  - **Impact:** keyboard and screen-reader users receive no explanation for the suddenly disabled controls during a picker/read/serialize/write operation, contrary to AC5's explicit busy-status announcement.
  - **Required resolution:** **FIX.** Expose a concise text-backed operation status/`aria-busy` without replacing the persistent dirty indicator, and cover it in component accessibility tests.

- [x] [Review][Resolved][Minor] **The story record contradicts its own status and overstates the structural proof.**
  - **Category:** Delivery Log / status and evidence truthfulness
  - **Location:** `_bmad-output/implementation-artifacts/5-5-open-and-save-folio-files-locally.md:4,12,228-236,296-298`
  - **Observation:** frontmatter, sprint status, and completion say `review`, while the prominent body field still says `ready-for-dev`. The developer log says the red proofs reject second capability branches, browser storage, and uploads even though reviewer mutations above stayed green; it also describes the focused browser source as local open/save/offline coverage when it is visibility/focus only.
  - **Impact:** the pipeline handoff and future audit cannot rely on the story's headline status or claimed verification scope without re-reading the code.
  - **Required resolution:** **FIX.** Reconcile the body status during finishing and amend—not erase—the record with the actual proof scope and final measured gates.
