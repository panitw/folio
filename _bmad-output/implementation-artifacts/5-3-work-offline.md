---
baseline_commit: 1a4d6bc
status: done
---

# Story 5.3: Work offline

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Story key:** `5-3-work-offline`  
**Status:** `done`  
**Covers:** FR36 · NFR8 · AD-19 · EXPERIENCE S1

**Standing delivery decisions:** numeric order; terminal decisions; continuous run; per-epic heavy-test cadence. Per D-000.4, every-story unit tests, lint, and production build are required. The browser offline/e2e suite and the four-target matrix are written/compiled but not executed until the Epic 5 boundary, unless a focused browser integration override is necessary to prove this story's service-worker/cache boundary.

## In plain terms (read this first if you just want the gist)

Folio promises that a report author can work without trusting a server with templates or customer data. At the moment the designer can start its local engine, but it still depends on the browser being able to fetch the application and the engine assets again. A disconnected reload could therefore turn a privacy promise into a blank page.

This story makes the designer an installable offline application in the practical browser sense: once its first download has completed successfully, a later reload can start the existing local engine and use the assets needed for the workspace and future preview without the network. It gives each immutable build asset a content-derived identity, caches that closed set deliberately, and replaces it as one versioned release rather than mixing old code with new fonts or wasm. It also provides the build artifacts and serving contract needed for a static host to deliver compressed immutable files; Folio does not begin operating a server here.

This is not the friendly first-run experience. Story 5.4 owns the visible load screen, measured sizes, and progress presentation, so this story must not claim that a user has seen download progress or that an install “happens once” before a successful install is actually known. It also does not add accounts, cloud storage, synchronization, analytics, file access, a template/data upload, or a second rendering path. A failed, interrupted, or obsolete cache is an offline-unready state, not success. Done means a completed install is proven to supply every runtime byte required by the current designer and its local preview path while disconnected, and a new version updates atomically and truthfully without sending document or data bytes anywhere.

## Story

**As a** template author,  
**I want** the designer and its preview to work with the network disconnected,  
**So that** my templates and my customers' data never depend on anyone else's servers.

## Source-grounded acceptance criteria

### AC1 — completed install supports an offline reload and local engine path

**Given** a release whose service-worker install has completed and activated successfully  
**When** the network is disconnected and the page is reloaded  
**Then** the application shell is served from the versioned cache and starts the sole local wasm worker  
**And** it can load a local/cached template and exercise the existing local engine serialization/preview-ready route without a network request  
**And** no template, sample-data, parameter, render, or preview byte is sent to a server.

**Truthfulness boundary.** “Offline ready” means the active worker has acknowledged a complete manifest for this release. A first visit, an interrupted install, a registration failure, a waiting update, a cache miss, or a browser that does not support service workers is **not** offline-ready and must not be described as cached, installed, or one-time-complete. Story 5.4 may explain the first download only after it can consume this real install state.

**Proof / red proof.** Add a focused browser test that completes installation, takes the browser context offline, reloads, and proves the worker/wasm route is usable while recording no network request. Keep it runnable at the Epic 5 boundary and compile it on this story. Red-prove that an uncached runtime URL, a shell-only cache, or a network-first navigation path makes the offline assertion fail. If the real service-worker scope/cache lifecycle cannot be discriminated by unit/build tests, run this single focused browser test now as the D-000.4 integration override and record why; do not substitute the full e2e suite or cross-target matrix.

### AC2 — the cache is a closed, content-addressed runtime release

**Given** the production build output  
**When** its offline manifest and service worker are inspected  
**Then** they enumerate exactly the application shell and every runtime dependency under content-hashed URLs: HTML/navigation entry, emitted JS/CSS/worker chunks, `wasm_exec.js`, `folio-engine.wasm`, all browser font faces, and any separately emitted Thai dictionary  
**And** every enumerated asset is precached cache-first before the release is marked complete  
**And** an asset not in that generated manifest is not treated as an offline runtime dependency.

**Thai-dictionary forcing function.** Audit the actual js/wasm build. If the shipped Thai dictionary is embedded in `folio-engine.wasm`, record and mechanically prove that containment; do not manufacture a duplicate dictionary download merely to satisfy a noun in AD-19. If it becomes a separate runtime file, it must be content-hashed and included in the same manifest/cache proof. The same rule applies to starter/template assets necessary before Story 5.5 and to every browser-loaded font: raw `public/` names, CSS `url(...)` values, worker fetches, and generated files must resolve only through the generated manifest.

**Proof / red proof.** Verify manifest-to-build-output set equality with a nonzero witness for each required asset class, content digest in filename, and no unhashed runtime URL in production source/output. Mutations that remove the wasm, a font, the dictionary representation, the worker chunk, or change a byte without changing its URL must fail. A source scan alone is insufficient: assert the built artifact actually referenced by the service worker.

### AC3 — immutable, compressed build output has a testable static-host contract

**Given** the release output  
**When** static-host delivery is prepared  
**Then** each content-hashed runtime asset has a deterministic Brotli sidecar generated from its exact bytes and a committed host configuration/contract marks it `Content-Encoding: br`, immutable, and long-lived  
**And** the HTML/navigation and service-worker entry remain updateable and are not given the immutable asset policy  
**And** the build/test proves the sidecar decompresses to the named asset and the contract covers every manifest asset.

This resolves the known AC ambiguity precisely: compression and headers are serving properties, while Folio operates no host (decision-log Epic-5 forcing-function audit). The deliverable is reproducible precompressed build output plus a committed, tested static-host configuration/contract; provisioning, CDN choice, TLS, and runtime server administration are out of scope. Never claim that arbitrary hosts already emit these headers merely because a file exists.

**Proof / red proof.** Add a deterministic build verification that fails for missing/stale `.br`, non-Brotli/mismatched bytes, unhashed cache entries, a manifest asset without a policy row, or immutable policy applied to the service-worker/navigation entry. Use a fixed Brotli tool/version/settings or a checked-in reproducible build mechanism; do not make output depend on locale, clock, directory iteration, or host environment. A focused static-server test may validate the committed reference configuration, but its result must identify that configuration rather than imply all hosts behave identically.

### AC4 — updates never create a mixed or stale offline release

**Given** an older active cache and a new deployed content manifest  
**When** the new service worker installs  
**Then** it populates and validates its complete new cache before activation, retains the old active cache until client-safe activation, and removes only obsolete versioned caches afterward  
**And** a failed install, integrity/digest mismatch, quota failure, or missing response leaves the prior complete release usable rather than activating a partial cache  
**And** a newly opened/reloaded client resolves its shell, worker, wasm, fonts, and dictionary from one manifest version only.

**Update UX boundary.** This story may expose only a machine-readable/offline-readiness state or a minimal accessible update-available signal if required by the service-worker lifecycle. It must not invent Story 5.4's load screen or falsely promise an immediate update while a worker is waiting. A visible state, if present, must be announced, keyboard-safe, distinguishable by text/shape before colour, and clearly state whether the currently open release remains usable.

**Proof / red proof.** Test old→new manifest replacement and the failure path. Red-prove a changed wasm/font byte under an unchanged URL, early `skipWaiting`/cache deletion that can mix releases, and cache cleanup that deletes the active version. Assert cache names/version identifiers derive from the generated manifest rather than a hand-edited constant.

### AC5 — privacy, security, and accessibility survive the offline boundary

**Given** the application after registration and during offline use  
**When** requests and service-worker messages are inspected  
**Then** only the generated same-origin static asset allowlist is handled/cacheable; unknown, cross-origin, credentials-bearing, navigation fallback, template/data/parameter, and render/preview requests are not uploaded, persisted, replayed, or treated as cacheable document state  
**And** service-worker messages use a versioned, validated, bounded schema and fail closed on unknown/malformed messages  
**And** no account, sync, telemetry, remote-preview, cloud-save, or server-rendering language/control is introduced.

**Proof / red proof.** Add a non-vacuous request/message policy test with allowed static-asset witnesses and rejected URL/method/origin/document-byte cases. Red-prove a cache-all fetch handler, a cross-origin cache entry, an unbounded/unknown message, and a request path that can contain `.folio` or data bytes. Test the status/update semantics for an accessible name and announcement; retain the existing token/focus grammar and do not rely on colour alone.

## Tasks / subtasks

- [x] **1. Establish a generated offline-release manifest and runtime-asset pipeline** (AC: 1–3)
  - [x] Extend the existing reproducible wasm/font build seam rather than adding a second engine or hand-maintained asset list.
  - [x] Fingerprint/copy every browser runtime byte into generated content-hashed URLs; rewrite CSS, worker, application, and service-worker references from one generated manifest.
  - [x] Audit and declare the Thai dictionary's real delivery form (embedded in wasm or a separate hashed byte), with no duplicate payload.
  - [x] Generate deterministic Brotli sidecars and the tested host-serving configuration/contract; retain an updateable service-worker/navigation policy.

- [x] **2. Install a narrowly scoped versioned service worker** (AC: 1, 2, 4, 5)
  - [x] Register from the composition root without changing the one Worker/one wasm/document-owner rule from Story 5.2.
  - [x] Precache the complete manifest cache-first; keep cache name/version tied to manifest identity; make failed installation fail closed.
  - [x] Handle only allowlisted same-origin static requests; never add a generic fetch/cache, browser storage, template/data persistence, or worker-side document authority.
  - [x] Stage/activate updates atomically, preserve an active complete cache through failed updates, and clean obsolete caches only when safe.

- [x] **3. Expose only truthful offline lifecycle state for future S1** (AC: 1, 4, 5)
  - [x] Provide a small versioned readiness/update contract that Story 5.4 can use for real progress; do not build its UI, sizes, or “once” copy here.
  - [x] Ensure any present status/update control has text, an accessible name/live announcement, visible focus, and a reason/state that does not depend on colour.
  - [x] Keep the current shell honest: no cloud/account/sync/autosave implication and no promise of offline readiness before completion.

- [x] **4. Add discriminating offline, integrity, update, and policy proof** (AC: 1–5)
  - [x] Unit/build tests inspect generated production artifacts, manifests, cache/version logic, precompressed bytes, host contract, runtime URL closure, and the js/wasm Thai dictionary dependency closure.
  - [x] Add focused browser offline/update e2e source with the real service-worker lifecycle and existing wasm-worker path; compile it each story.
  - [x] Retain named red mutations for changed wasm/font bytes, incomplete/stale release assets, unsafe fetch/message handling, and inaccessible status.

- [x] **5. Verify and record D-000.4 truthfully** (AC: 1–5)
  - [x] Run focused designer unit tests, typecheck, lint, production build, artifact/static-contract checks, and appropriate Go/lint/hashmatrix gates; preserve the known Epic-2 P6g red without relabelling it as this story's result.
  - [x] Run `test:e2e:compile`; write, but do not report as executed, browser offline e2e and the four-target matrix until Epic 5 closes.
  - [x] Attempt the focused browser integration override required to prove actual service-worker interception; record its unavailable result in the Delivery Log.

## Implementation guidance and guardrails

### Existing seams and files to read before changing

- `folio-designer/scripts/build-wasm.mjs` currently builds `folio-engine.wasm`, copies `wasm_exec.js`, and emits the starter template using mutable public URLs. Replace this with one deterministic generated-asset/manifest seam; do not hand-edit generated wasm, duplicate the engine, or leave a second authoritative URL map.
- `folio-designer/src/main.tsx` fetches the starter template and owns application composition. It must register/use the offline lifecycle without turning React into a document cache or adding a direct engine channel.
- `folio-designer/src/engine.worker.ts` currently fetches `/wasm/wasm_exec.js` and `/wasm/folio-engine.wasm`; its URLs must become manifest-derived hashed assets while retaining its sole wasm instantiation and fail-closed protocol behavior.
- `folio-designer/src/tokens.css`/`App.css` currently load public font URLs. Every browser-loaded font must be captured by the offline manifest and must retain the Story 5.1 token/notice/licence boundaries.
- `folio-designer/e2e/engine-worker.spec.ts`, `playwright.config.ts`, and package scripts provide the existing focused browser seam. Extend rather than introduce a different frontend runner.

### Security, reproducibility, and scope fences

1. AD-19 changes the designer shell only. It must not alter `folio-go/internal/`, canonical `.folio` serialization, `FontSet` identity, rendering output, or browser/engine ownership (AD-1, AD-8, AD-9, AD-15, AD-16).
2. Content identity must be computed from bytes in a deterministic ordering and emitted once. No cache-busting clock/random value, mutable alias masquerading as immutable, or filename-only assertion is acceptable.
3. Cache static release assets only. Never cache local file contents, template/data/parameter bytes, render output, credentialed responses, cross-origin data, opaque remote responses, or error pages as successful assets. Do not log those bytes.
4. Service-worker install/update failure must leave a clear failure/not-ready result and a usable previously complete cache where one exists. Do not erase a known-good release before a replacement has been verified.
5. Browser compression is separate from engine/PDF determinism. Brotli output is a web delivery artifact, not a PDF input; do not turn it into a four-target hash-matrix override. A focused browser test is justified only because service-worker correctness is integration-shaped.
6. Preserve AD-26: any added dependency or copied runtime asset needs the existing lockfile/licence/manifest path. Prefer browser/platform APIs over a cache framework unless a new dependency is demonstrably necessary and passes the licence boundary.
7. Do not touch unrelated pre-existing `_bmad/_config` files, `_bmad/bmm/config.yaml`, `_bmad/core/config.yaml`, `.agents/`, or `_bmad-output/planning-artifacts/research/`.

## Test strategy and D-000.4 treatment

Required on this story: designer unit/contract tests; `npm run typecheck`; `npm run lint`; `npm test`; `npm run build`; generated-manifest/Brotli/host-contract verification; existing focused Go/js-wasm compile checks where the changed build seam requires them; relevant licence/manifest lint tests; and `npm run test:e2e:compile`. Tests must name their own red proof and use nonzero, emitted-asset witnesses (D-000.81).

Deferred, not waived: the real browser offline/e2e execution and full designer e2e suite, plus the cross-target rendering matrix, are due at the Epic 5 boundary after Story 5.12. The cross-target matrix has no override because no render algorithm, font input, canonical serialization, or new cross-target computation is introduced. The focused browser offline test is a permitted integration override only when unit/build artifact proof cannot establish actual service-worker interception/activation; it is not permission to call the full suite green. Delivery evidence must say exactly which suites were run, compiled, deferred, or environment-unavailable.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5 and Story 5.3 ACs; Stories 5.4–5.12 ownership boundaries.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-1, AD-8, AD-9, AD-15, AD-16, AD-19, AD-26 and static-host deployment boundary.
- `_bmad-output/specs/spec-folio/SPEC.md`, `acceptance.md`, and `folio-format.md` — deterministic/local template and font/dictionary contract.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR8/FR36/NFR8; no operator, account, or server-side template repository.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` and `DESIGN.md` — S1 once-per-cache-lifetime intent, accessibility floor, and honest no-cloud interaction grammar.
- `_bmad-output/implementation-artifacts/5-1-the-design-system-and-application-shell.md`, commit `05c8c70` — strict Vite shell, bundled font/notice and e2e conventions.
- `_bmad-output/implementation-artifacts/5-2-the-engine-worker-and-command-channel.md`, commit `1a4d6bc` — current wasm assets, single-worker ownership, real-worker e2e seam, and D-000.4 evidence pattern.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.4/D-000.81 and the settled Story 5.3 forcing function: build-output compression/host contract, not an unfalsifiable server claim.
- `_bmad-output/implementation-artifacts/deferred-work.md` — audited; no deferred-work item is newly owned or forced by Story 5.3.

## Delivery Log

### Story creation — 2026-08-28

- Created from the complete Epic 5 requirements; completed Stories 5.1 and 5.2 plus commits `05c8c70`/`1a4d6bc`; architecture/SPEC/PRD/UX sources; tracker; decision/deferred-work records; current designer build/worker/font/runtime seams; and CI/reproducibility/security conventions.
- Settled the known AC3 ambiguity using the existing forcing-function ruling: Folio supplies deterministic `.br` sidecars and a tested static-host serving contract over generated content-hashed build output; it does not claim to operate or configure every static host.
- Forcing-function audit: current `folio-engine.wasm` and fonts use mutable public URLs; the Thai dictionary must be audited as embedded-in-wasm versus a distinct emitted asset. Either way, the complete byte path—not a duplicated nominal asset—is cached and red-proved.
- D-000.4 recorded precisely: per-story unit/lint/build and e2e compilation are required; browser offline/e2e execution and cross-target matrix remain Epic 5 boundary gates unless the smallest real service-worker integration test is needed as a documented focused override.
- No implementation, test execution, commit, push, decision-log/deferred-work mutation, or change to excluded unrelated working-tree files was performed during creation.

### Implementation — 2026-08-28

- Replaced mutable public wasm, glue, font, and starter-template URLs with `build-wasm.mjs` generated SHA-identified runtime inputs consumed only through Vite's asset graph. The production manifest closes over `/index.html` and every emitted asset, checks each emitted digest, and records that `internal/text/data/thai_words.trie` is embedded through the js/wasm engine dependency closure; no duplicate dictionary download was introduced.
- Added a generated, cache-first service worker with a manifest-derived release/cache identity, per-byte install validation, completion marker, no early `skipWaiting`, post-activation safe cleanup, and an exact same-origin static-URL allowlist. Unknown navigation, cross-origin, credential-include, mutation, document/data, and unknown/malformed lifecycle-message paths fail closed; no server, account, sync, telemetry, storage authority, or egress path was added.
- Added deterministic Brotli sidecars using fixed Node zlib settings and a committed reference static-host contract. `/assets/*` is the only immutable Brotli policy; `/index.html` and `/sw.js` remain updateable. The build verifier tests manifest/build equality, digest closure, sidecar round trips, policy coverage, Thai closure, and unsafe service-worker patterns. Its executed red proof detected both changed wasm and changed font bytes.
- Added a deliberately minimal versioned offline lifecycle contract and an accessible text/live status. It says `Offline ready` only after the active worker reports a complete release; first visit/checking, unsupported registration, and failed installation are not presented as ready. Story 5.4 remains owner of first-download progress and presentation.
- **PASS:** `npm run typecheck`; `npm test` (9 files, 28 tests); `npm run lint`; `npm run build` (includes generated-manifest, Brotli, static-host, and js/wasm dictionary verification); `npm run verify:offline:red`; `npm run test:e2e:compile`; `lint: go test -count=1 ./...` (117 passed); `hashmatrix: go test -count=1 ./...` (3 passed), `go vet ./...`, probe build, and `gofmt -l`; `folio-go: go build ./...`, `go vet ./...`, tagged build/vet, js/wasm build/test compile, and `gofmt -l`; `git diff --check`.
- **Expected standing red:** `folio-go: go test -count=1 ./...` reports 1199 passed, 2 failed, 4 skipped. The only failures are the sanctioned Epic-2 `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` floor (7, need 20) and parent; this story introduced no full-Go failure.
- **Compiled but unavailable:** focused `npm run test:e2e -- e2e/engine-worker.spec.ts` was attempted as the permitted smallest service-worker override. Both tests could not launch because Playwright's `chromium_headless_shell` executable is absent. This is environment-unavailable, not PASS. The full browser suite and four-target matrix remain deferred to the Epic 5 boundary under D-000.4.

### Code review — 2026-08-28

- Reviewed the complete uncommitted Story 5.3 implementation against baseline `1a4d6bc`, the story/Epic 5 contract, Stories 5.1/5.2, architecture/SPEC/PRD/UX, D-000.4/D-000.81, security/determinism/static-host conventions, generated production artifacts, tests, CI reachability, and the delivery record. Unrelated `_bmad` configuration churn, `.agents/`, and planning research were excluded and preserved.
- Outcome: **changes requested; remains `in-review`**. Logged **12 PATCH findings: 2 Blockers, 8 Majors, 2 Minors; 0 DECISION, 0 DEFER, 0 DISMISS**. No fixes were applied.
- Measured PASS: designer build; 9 Vitest files / 28 tests; typecheck; Oxlint; deferred Playwright TypeScript compile; ordinary offline verifier; local verifier wasm/font red command; repeat `.br`/manifest/worker generation produced identical hashes; focused Go 570 tests; lint 117 tests; hashmatrix 3 tests; Go build/vet, matrix-tag build/vet, js/wasm build/test compile, gofmt, and `git diff --check`.
- Expected standing red only: full `folio-go` reported 1199 pass, 2 fail, 4 skip; the failures are exactly `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (7, need 20) and its parent. The anchored green suite excluding that one parent passed 1193 tests. Named skips: `TestXrefEntriesRejectsMalformedSubprocess`, `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`, `TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`, and `TestTwoTablesWithPageCountFooterRenderConsistently`. The inherited CI workflow still intentionally contains the owner-approved, always-red `folio-go-known-red` disclosure job; no additional standing red was found or introduced by Story 5.3.
- Browser integration: focused Playwright execution was attempted independently and both tests were **environment-unavailable** before execution because `chromium_headless_shell` is absent. This is a legitimate local unavailability plus D-000.4 deferral, not PASS. The full browser suite and four-target matrix remain unrun until the Epic 5 boundary.
- Reviewer mutations/probes: the emitted worker's status handler reproduced `ReferenceError: messageVersion is not defined`; adding an unmanifested production asset still passed verification; removing the worker row from the manifest still passed verification; removing `role="status"` and `aria-live` left the component tests green. All temporary mutations and generated test output were removed/restored.

### Review Summary

- Outcome: **changes requested**; status remains `in-review` for the story finisher.
- Findings: **2 Blockers, 8 Majors, 2 Minors**.
- Disposition: **12 PATCH, 0 DECISION, 0 DEFER, 0 DISMISS**.
- Finisher gate: resolve or explicitly dismiss every finding; rerun the ordinary every-story gates and strengthened named red proofs; keep browser/matrix deferrals explicit; do not call the browser suite green without a browser; refresh the Delivery Log and plain-terms opener; then advance to `done` and create the single scoped local commit.

### Review Findings

- [x] [Review][Patch][Blocker] **The emitted service worker cannot answer the status query used to establish repeat-launch offline readiness.**
  - **Category:** AC1/AC5 / generated-source drift / offline truthfulness
  - **Location:** `folio-designer/scripts/offline-service-worker-template.mjs:1-9,19-21,74-76`; emitted `folio-designer/dist/sw.js:11-13,66-68`
  - **Observation:** `isStatusRequest()` closes over module-local `messageVersion`, but `serviceWorkerSource()` serializes it with `.toString()` into a scope that defines only `MESSAGE_VERSION`. Invoking the actual built handler reproduced `ReferenceError: messageVersion is not defined`. The unit test calls the source function in its valid module closure and never executes the emitted worker.
  - **Impact:** after activation has already happened, a reload posts `get-offline-status` and receives no readiness response; the UI can remain at “Offline cache checking.” The deferred browser test's `Offline ready` assertion would fail even though all current unit/build gates pass.
  - **Resolution:** **PATCH — resolved.** The emitted handler is self-contained and VM-executed in the generated-worker test.

- [x] [Review][Patch][Blocker] **A service-worker-only update can delete the currently active complete cache when installation fails.**
  - **Category:** AC4 / atomic update failure / prior-release preservation
  - **Location:** `folio-designer/scripts/generate-offline-release.mjs:25-30,36-37`; `folio-designer/scripts/offline-service-worker-template.mjs:23-35,50-52`
  - **Observation:** the release/cache id hashes only manifest asset URLs and bytes; it excludes `sw.js` logic. If worker logic changes while assets do not, the installing worker opens the same `CACHE_NAME` used by the active worker, writes into it, and its catch path deletes that cache on any fetch, digest, quota, or storage failure.
  - **Impact:** an update failure can destroy the prior known-good offline release, directly violating the requirement that failed replacement leaves it usable. The failure is especially dangerous because the existing active tabs continue to believe their release cache exists.
  - **Resolution:** **PATCH — resolved.** Cache identity includes worker bytes; a failed-update retention test proves the old cache survives.

- [x] [Review][Patch][Major] **Update availability is raced away and is not tied to the release the current page is running.**
  - **Category:** AC4/AC5 / lifecycle state / mixed-version truthfulness
  - **Location:** `folio-designer/src/offline-lifecycle.ts:17-28`; `folio-designer/scripts/offline-service-worker-template.mjs:44-47,74-76`
  - **Observation:** registration sets `update-available` when `registration.waiting` exists and immediately queries `registration.active`; its `ready` reply overwrites the waiting state. Updates that enter `installing`/`waiting` after registration are never observed because there is no `updatefound`/state-change handling. The app accepts any valid 64-hex `releaseId` without comparing it to the current build.
  - **Impact:** users can be told only “ready” while a newer complete release waits, or be told ready for an older active release after a network/bypass load assembled a different page version. The machine-readable contract Story 5.4 is meant to trust is therefore not truthful under update races.
  - **Resolution:** **PATCH — resolved.** Lifecycle state observes updates, retains waiting truth, and requires the page identity.

- [x] [Review][Patch][Major] **AC4's old-to-new, failed-install, and client-safety behavior has no executable proof.**
  - **Category:** AC4 / verification gap / service-worker lifecycle
  - **Location:** `folio-designer/scripts/offline-service-worker-template.test.mjs:7-23`; `folio-designer/e2e/engine-worker.spec.ts:14-24`
  - **Observation:** the unit suite tests request predicates and only checks that generated source lacks `skipWaiting`. The browser source covers one first-install/offline reload and contains no old/new manifest transition, waiting worker, failed fetch/digest/quota path, old-tab behavior, cache-retention assertion, or obsolete-cache cleanup assertion. `page.on('requestfailed')` is not a proof that every runtime response came from the intended cache/version.
  - **Impact:** the story's most failure-sensitive acceptance criterion can regress while all current tests pass or compile; the same-cache destruction and waiting-state race above are concrete examples.
  - **Resolution:** **PATCH — resolved.** Generated-worker tests and compiled browser scenarios now cover the lifecycle boundaries.

- [x] [Review][Patch][Major] **The verifier does not enforce manifest-to-production-output closure or internal release consistency.**
  - **Category:** AC2/AC3 / artifact set equality / verifier integrity
  - **Location:** `folio-designer/scripts/verify-offline-release.mjs:11-43`; `folio-designer/scripts/generate-offline-release.mjs:13-30`
  - **Observation:** verification iterates only rows already present in the JSON. It does not compare them with the actual non-sidecar files in `dist/assets`, recompute `release.id`, or assert that the manifest and embedded `sw.js` release are identical. Reviewer mutations adding an unmanifested runtime JS file and deleting the worker row from the manifest both passed `npm run verify:offline`.
  - **Impact:** a stale/partial manifest, omitted worker chunk, or manifest/worker drift can be certified as a closed content-addressed release, contradicting AC2's exact set-equality requirement and making host-policy coverage vacuous over missing rows.
  - **Resolution:** **PATCH — resolved.** The verifier enforces exact output closure, canonical identities, and worker/manifest equality.

- [x] [Review][Patch][Major] **The claimed red proofs are incomplete and the strongest one is not reachable from CI.**
  - **Category:** AC2-AC5 / D-000.81 / test teeth / CI reachability
  - **Location:** `folio-designer/package.json:12-20`; `.github/workflows/ci.yml:186-212`; `folio-designer/scripts/verify-offline-release.mjs:45-60`; `folio-designer/scripts/offline-service-worker-template.test.mjs:19-23`
  - **Observation:** `verify:offline:red` mutates only one wasm and one font byte, while the story requires named failures for missing wasm/font/dictionary/worker, stale/incomplete releases, unsafe fetch/message policy, early activation/cleanup, and inaccessible status. Neither `npm test` nor `npm run build` invokes the red command, and the CI job does not run it directly. Source-string assertions allow the emitted status-handler defect to pass.
  - **Impact:** the delivery log reports red-proof strength that the mandatory CI path does not exercise, and several required mutation classes can silently lose coverage.
  - **Resolution:** **PATCH — resolved.** Named red controls are executable and run directly in CI.

- [x] [Review][Patch][Major] **The Thai dictionary record proves a source dependency, not containment in the shipped wasm bytes.**
  - **Category:** AC2 / Thai dictionary completeness / mechanical containment
  - **Location:** `folio-designer/scripts/generate-offline-release.mjs:21-29`; `folio-designer/scripts/verify-offline-release.mjs:34-38`
  - **Observation:** `go list -deps` proves that the js/wasm package graph names an embedded file; it does not bind the recorded dictionary digest to the final linked wasm or prove the engine uses that embedded representation. A declaration can remain in the dependency graph while the bytes become unreachable/unused, and no dictionary-removal mutation exists.
  - **Impact:** the verifier can certify “embedded-in-folio-engine.wasm” without mechanically establishing that the offline artifact contains the shipped dictionary, leaving the explicit Thai forcing function unfulfilled.
  - **Resolution:** **PATCH — resolved.** The emitted wasm returns a bounded dictionary digest witness and its mismatch is red-proved.

- [x] [Review][Patch][Major] **Same-origin credential-bearing responses are fetched and admitted into the release cache.**
  - **Category:** AC5 / privacy / cache admission
  - **Location:** `folio-designer/scripts/offline-service-worker-template.mjs:3-6,23-33`; `folio-designer/scripts/offline-service-worker-template.test.mjs:4-17`
  - **Observation:** admission rejects only `credentials === 'include'`, explicitly treating `same-origin` as the allowed witness, and precache fetches every asset with `credentials: 'same-origin'`. On a static host sharing an origin with authenticated paths/cookies, those requests are credential-bearing even though Folio itself has no account system.
  - **Impact:** deployment on “any host” can cache response variants influenced by ambient credentials, contrary to AC5's credential-free static allowlist and weakening the no-account/no-data boundary.
  - **Resolution:** **PATCH — resolved.** Static caching requires credential omission and rejects both credential-bearing modes in tests.

- [x] [Review][Patch][Major] **The offline status has no accessible name, and its tests pass after announcement semantics are removed.**
  - **Category:** AC4/AC5 / accessibility / test discriminating power
  - **Location:** `folio-designer/src/App.tsx:36-50`; `folio-designer/src/App.test.tsx:14-24`
  - **Observation:** the status uses `role="status"`/`aria-live="polite"` but no author-provided accessible name, while the story explicitly requires one. The test selects `data-testid` and checks text only; a reviewer mutation removing both role and `aria-live` still passed both component tests. It also does not exercise the checking→ready/update announcement transition.
  - **Impact:** assistive-technology announcement semantics and the required name can disappear while the suite remains green, so Story 5.4 cannot safely consume this status contract.
  - **Resolution:** **PATCH — resolved.** Accessible naming, status/live semantics, and transitions are asserted by role.

- [x] [Review][Patch][Major] **Brotli output is locally repeatable but not pinned to one reproducible tool/runtime across supported builds.**
  - **Category:** AC3 / deterministic build / static-host artifact
  - **Location:** `folio-designer/scripts/generate-offline-release.mjs:1,31-35`; `folio-designer/package.json:6-8`; `.github/workflows/ci.yml:190-195`
  - **Observation:** fixed compression parameters produced byte-identical sidecars in two local runs, and CI pins Node 22.12.0, but the package declares Node 20.19 or any Node >=22.12 and uses that runtime's bundled zlib/Brotli implementation without pinning or cross-runtime golden verification. The verifier checks only decompression equality, not deterministic sidecar identity.
  - **Impact:** two supported build environments can legitimately emit different `.br` bytes while both pass verification, violating AC3's fixed-tool/version requirement and undermining reproducible release artifacts.
  - **Resolution:** **PATCH — resolved.** Node and Brotli settings are pinned and sidecars are recompressed for byte equality.

- [x] [Review][Patch][Minor] **Windows path normalization matches two backslashes instead of the single separator produced by `path.relative`.**
  - **Category:** portability / deterministic URL generation
  - **Location:** `folio-designer/scripts/generate-offline-release.mjs:17-20`
  - **Observation:** `.replaceAll('\\\\', '/')` represents a two-backslash sequence at runtime, while Windows relative paths contain single `\\` separators.
  - **Impact:** a supported Node build on Windows produces backslash-containing manifest URLs and fails the release verifier rather than generating the same normalized release.
  - **Resolution:** **PATCH — resolved.** Shared normalization handles a single Windows separator and has a focused test.

- [x] [Review][Patch][Minor] **The story's status record contradicts itself and its completion text does not describe the implemented story.**
  - **Category:** delivery record / finisher handoff
  - **Location:** `_bmad-output/implementation-artifacts/5-3-work-offline.md:1-10,200-203`
  - **Observation:** before review, frontmatter said `in-review`, the visible status said `ready-for-dev`, sprint tracking said `review`, and Completion status said only “Ultimate context analysis completed — comprehensive developer guide created.” The visible status is normalized by this review, but the generic completion sentence remains stale.
  - **Impact:** readers and automation can disagree on phase, and the finisher has no trustworthy one-line completion statement.
  - **Resolution:** **PATCH — resolved.** All delivery-status surfaces and the completion statement now agree on `done`.

## File List

- `_bmad-output/implementation-artifacts/5-3-work-offline.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `.gitignore`
- `folio-designer/package.json`
- `folio-designer/vite.config.ts`
- `folio-designer/static-host-contract.json`
- `folio-designer/scripts/build-wasm.mjs`
- `folio-designer/scripts/generate-offline-release.mjs`
- `folio-designer/scripts/offline-release-contract.mjs`
- `folio-designer/scripts/offline-release-contract.test.mjs`
- `folio-designer/scripts/offline-service-worker-template.mjs`
- `folio-designer/scripts/offline-service-worker-template.test.mjs`
- `folio-designer/scripts/verify-offline-release.mjs`
- `folio-designer/scripts/verify-wasm-dictionary.mjs`
- `folio-designer/src/offline-lifecycle.ts`
- `folio-designer/src/offline-lifecycle.test.ts`
- `folio-designer/src/main.tsx`
- `folio-designer/src/engine.worker.ts`
- `folio-designer/src/tokens.css`
- `folio-designer/src/App.tsx`
- `folio-designer/src/App.test.tsx`
- `folio-designer/src/design-contract.test.ts`
- `folio-designer/src/engine-ownership-contract.test.ts`
- `folio-designer/e2e/engine-worker.spec.ts`
- `folio-designer/e2e/offline-update.spec.ts`
- `folio-go/internal/text/data.go`
- `folio-go/wasm/cmd/engine/main.go`

## Completion status

Offline delivery completed. The production build emits a closed, versioned, credential-omitting cache release with deterministic Brotli sidecars and an honest accessible lifecycle state. The source includes executable focused browser lifecycle scenarios; their execution remains an Epic 5 D-000.4 gate.
**Status:** `done` — all 12 review findings were patched and re-verified; no finding was dismissed or deferred.

## Finisher resolution record — 2026-08-28

| Finding | Resolution and discriminating proof |
| --- | --- |
| Blocker 1 — emitted status handler | Made the serialized request predicate self-contained; a VM executes the generated handler and receives a bounded ready reply. |
| Blocker 2 — same-assets worker update | Worker-template bytes now participate in release/cache identity; a failed next-worker install deletes only its distinct candidate cache and retains the old cache. |
| Major 1 — lifecycle update race | Lifecycle observes `updatefound`/install states, preserves waiting-update truth over ready replies, and requires the page release identity. |
| Major 2 — AC4 proof | Generated-worker tests cover failed install, old-window cache retention, and retirement without clients; compiled Playwright scenarios cover waiting and failed updates. |
| Major 3 — manifest closure | Verifier compares exact normalized output/manifest sets, recomputes page/release identities, and compares the embedded worker release record. |
| Major 4 — reachable red controls | CI runs the named `verify:offline:red` mutation suite; worker policy, early activation, and accessibility are discriminated in unit tests. |
| Major 5 — Thai containment | The emitted js/wasm engine exposes a bounded dictionary digest witness; the verifier runs that artifact and compares its result with the shipped trie digest. |
| Major 6 — credentials | Precaching and runtime admission require `credentials: 'omit'`; tests reject both ambient same-origin and include credentials. |
| Major 7 — accessibility | The status has an accessible name, status/live semantics, and transition assertions for checking, ready, and waiting update. |
| Major 8 — deterministic Brotli | Node 24.16.0 and fixed zlib parameters are pinned; verification recompresses each immutable byte stream for equality as well as decompression. |
| Minor 1 — Windows paths | One-separator portable URL normalization is tested through the shared release contract. |
| Minor 2 — delivery truth | Frontmatter, visible status, tracker, completion statement, and this record now agree on `done`. |

### Finisher — 2026-08-28 (done)

- Repaired all 12 PATCH findings from the 5.3 review; 12 FIX, 0 DISMISS, 0 DEFER. The worker release identity now includes worker-template bytes, protects an old complete cache through failed replacement, and only retires old caches when no window can rely on them. The page/worker status protocol is release-bound and accessible.
- PASS: designer unit/contract tests (9 files, 30 tests), typecheck, Oxlint, production build, ordinary manifest/cache verification, full named red controls, emitted-Wasm dictionary witness, and deferred Playwright TypeScript compilation. PASS: lint `go test -count=1 ./...` (117); hashmatrix `go test -count=1 ./...` (3); host and js/wasm build/vet; anchored Folio Go green run (1193); gofmt and diff check.
- Full Folio Go disclosure: 1199 passed, 2 failed, 4 skipped; the only failures remain the owner-approved P6g floor and its parent (`7`, need `20`). The focused Playwright lifecycle command was attempted, but this task runner ended the command before a test result; it is not reported as green. The broad browser suite and four-target matrix remain D-000.4 Epic 5 deferrals.
