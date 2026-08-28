---
baseline_commit: e314c14
story_key: 6-7-the-round-trip-closes
status: done
created: 2026-08-28
---

# Story 6.7: The round trip closes

**Epic:** 6 — A template author can bind a report to data and build the golden report  
**Covers:** S5, S7, S8, S9 · FR7–10, FR21, FR35, FR38, FR41 · NFR1, NFR5  
**Governing rulings:** AD-8, AD-9, AD-14–18, AD-23 · D-000.4

## In plain terms (read this first if you just want the gist)

In one real browser session, an author starts with the Go-owned empty starter and creates a Customer Account Statement through normal controls: embedded logo, customer/account/period framing, mixed Latin/Thai/CJK transaction data, five Date/Description/Debit/Credit/Balance columns, generated-date and Page X/Y footer, Preview, and Save As. A fixture is never opened or re-saved as browser evidence. The durable manifest retains the initial `initialize`, exact command/request history, raw render inputs, and hashes; disposable browser PDFs and native binaries stay ignored.

The saved `.folio` bytes are the handoff. They must be exactly the canonical bytes returned by the engine at the saved revision—not a TypeScript reconstruction, normalization, or second serialization. Preview must render those template bytes with the same raw sample-data and parameter bytes. Native production rendering consumes the saved file and byte-identical data/parameter files with the shipped FontSet, producing the PDF bytes admitted by the browser viewer. Compare bytes and SHA-256: equal hashes alone do not establish byte equality.

The proof needs a second, structurally different browser-authored template. It also needs a canonical hand-edited template never opened by the designer: native load/validate/render must succeed. Enumerate the actual closed Go diagnostic registry at test time and exercise every registered error code through its production path. Every case yields a stable code, actionable message, and location whenever applicable; warnings remain successful renders.

Keep one authority throughout: Go owns parse, canonical serialization, document state, validation, diagnostics, identity, and rendering; the one worker owns one Wasm instance; React only drives controls and observes immutable results; PDF.js only admits/displays the returned PDF. This story is inherently browser/native integration-shaped, so it is the explicit D-000.4 exception: run the necessary real compiled Playwright/browser and native integration evidence here. The full four-target matrix remains the Epic 6 boundary gate, and no release tag is cut before that boundary succeeds.

## Story

As a solo builder,  
I want a template authored in the browser to render byte-identically through the Go library,  
so that the claim the entire product rests on is demonstrated end to end.

## Acceptance criteria

### AC1 — An authored browser session closes over saved, Preview, and native bytes

**Given** a fresh browser document and local golden-report sample/parameter files,  
**When** a real compiled-browser test authors the golden report through the normal palette, binding tree, parameter, table-editor, and save interactions (rather than opening/initializing/re-saving a fixture),  
**Then** the test records non-vacuous command/session evidence, saves the exact canonical byte array returned by `serialize` at the authored revision, and proves the downloaded/local saved `.folio` is byte-for-byte that array.

**And** Preview is admitted by the existing one-worker/Wasm/PDF.js route from that exact saved canonical template byte array and the exact raw data and parameter bytes; the native `folio-go` production entry point reads those same three byte arrays with `fonts.Shipped()` and its PDF bytes are byte-for-byte equal to the browser-admitted PDF.

**And** the proof records byte length and SHA-256 for the saved template, data, params, browser PDF, and native PDF, but fails on a direct byte mismatch even if any digest assertion is altered. It proves the browser viewer admitted the candidate before comparing it; a worker-produced but unadmitted PDF is insufficient.

**Proof / red proof.** Use a fresh test download/output directory; fail if authorship uses `load`, `initialize` with fixture bytes, a prepared `.folio`, or an unchanged fixture re-save. Red-prove one-byte changes to the saved template/data/params, a native render fed a reconstructed/normalized value, a PDF compared only by hash, and a second renderer/worker/browser layout path.

### AC2 — A second authored shape and a portable hand-edited template work

**Given** a second fresh browser session,  
**When** it authors and saves a template structurally different from the golden report (different bands/component mix and no reuse of the golden `.folio` bytes),  
**Then** the same saved-byte → admitted Preview → native production render chain is byte-identical and deterministic across repeated browser/native renders with unchanged inputs.

**Given** a separately hand-edited, canonical `.folio` that the designer never opens,  
**When** the native CLI/library loads, validates, and renders it with explicit data and parameters,  
**Then** it succeeds and produces a well-formed deterministic PDF; the test proves no browser file-open or serialization route touched that template.

**Proof / red proof.** Assert structural difference from the golden document using Go-owned parsed/template facts (not TypeScript schema parsing), and use distinct authored component ids/command history. Red-prove a copied golden fixture, a browser-created alternate, a byte-only non-canonical hand edit, or host fonts/environment-dependent inputs.

### AC3 — The diagnostic registry has a complete, actionable production census

**Given** the constructed closed registry in `folio-go/internal/diag`,  
**When** the census suite runs,  
**Then** it derives its expected set from the registry value at runtime and has one real production trigger for every registered error-severity code—no hard-coded subset, ignored code, name-pattern exclusion, or synthetic TypeScript error.

**And** each trigger returns the exact registered code, non-empty actionable message, and the applicable element id and/or data path through `folio.Render`/CLI/Wasm producer transport. The census explicitly distinguishes valid warnings (successful PDF plus diagnostics) from render-stopping errors, preserving the one `Diagnostic` authority.

**Proof / red proof.** The test fails for an unexercised newly registered code, duplicate/missing registry membership, blank message, changed code meaning, or an error rendered as a warning/current PDF. Preserve the existing diagnostic registry and all settled meanings; add no browser registry or parallel taxonomy.

## Tasks / subtasks

- [x] **1. Build an integration harness around existing authorities** (AC: 1–2)
  - Inspect and reuse `folio-designer/src/App.tsx`, `engine-client.ts`, `engine.worker.ts`, `preview/pdf-viewer.tsx`, `folio-go/wasm/engine.go`, `wasm/cmd/engine/main.go`, and `cmd/folio/main.go`; do not add a renderer, engine instance, template parser, or document model in TypeScript.
  - Add a fresh-directory orchestration seam that captures browser downloads and raw bytes without mutating tracked fixtures/goldens. Native rendering must use the public library/CLI production path and `fonts.Shipped()`, not an internal helper.

- [x] **2. Author the golden report through real browser controls** (AC: 1)
  - Drive a compiled Playwright browser through real UI interactions to create a fresh report and demonstrate the complete Epic 6 workflow. Capture command evidence, snapshot revision, `serialize` output, accepted raw data/params, Preview admission, downloaded file, and PDF bytes.
  - Make authorship discriminating: the harness must fail when its command path is replaced by fixture Open/Load/Initialize or a no-op save. Do not silently seed document state or inspect `.folio` fields in TypeScript.

- [x] **3. Compare the exact channels against native production** (AC: 1–2)
  - Persist the captured canonical template/data/params bytes verbatim, run native render twice from those files, and compare every browser/native/repeat PDF with `bytes.Equal` plus recorded SHA-256/length evidence.
  - Add the structurally distinct authored session and independently authored canonical hand-edit witness. Let Go parse/inspect structure for the distinction; do not duplicate the schema in the browser test.

- [x] **4. Add the full diagnostic-registry census** (AC: 3)
  - Extend Go production-level tests near `render_error_test.go`, `diagnostic.go`, or an appropriately focused new test to enumerate the constructed registry. Map every error code to a real malformed-template, binding, expression, glyph/content, or later registered condition and assert location/actionability.
  - Treat unregistered/dormant/duplicate mapping as a test failure. Exercise producer transport only where it proves the existing browser boundary; never fabricate a diagnostic in React/TypeScript.

- [x] **5. Run and record the integration exception honestly** (AC: 1–3)
  - Run the focused/full designer gates, real compiled Playwright execution, native Go tests/build/vet/formatting, Wasm build, and the new browser/native harness. Record commands, platform, exact results, output locations/digests, and the known P6g corpus-floor red without weakening it.
  - This story may run its necessary real browser/native integration under D-000.4. Keep the full `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` matrix as the Epic 6 boundary task; do not claim it green or create a release tag here.

## Developer guardrails

1. **Authored means authored this session.** No fixture reopen/resave, seeded canonical bytes, hidden command, or snapshot restore can satisfy AC1. Prove the controls produced committed commands and a changed fresh engine document.
2. **Save the exact canonical bytes.** `Engine.Serialize` remains the sole source. Downloaded bytes must equal that buffer exactly; no JS `JSON.stringify`, decoder/re-encoder, newline repair, or browser-owned canonicalizer.
3. **Keep all three input channels exact and distinct.** Template, data, and params travel as raw byte copies. Native must receive the captured files unchanged; retain exact-decimal numeric lexemes and never parse JSON in TypeScript for render comparison.
4. **One renderer and one viewer.** Production `folio.Render`/Wasm renders; PDF.js admits/displays. Canvas, screenshots, text extraction, hashes alone, or a browser approximation cannot establish PDF byte identity.
5. **Use shipped-font identity.** Browser/Wasm and native use the same repository FontSet. No host font, locale, wall clock, environment, network, or unpinned date input may enter the proof.
6. **Census the registry, not a hand list.** Derive expected codes from `internal/diag`'s constructed registry and keep an explicit real trigger map whose coverage is exact. Stable code meanings are additive-only.
7. **Warnings are not errors.** A successful PDF with warnings remains a successful render; render-stopping diagnostic codes must be located/actionable. Do not change Story 6.6’s honest failure provenance/UI split.
8. **Boundary discipline.** The real browser/native round trip is allowed here because this story demands it. Do not run or claim the four-target matrix as complete, alter fixtures/goldens/fonts for convenience, cut a tag, commit, or touch unrelated dirty `_bmad`/`.agents`/research work.

## Project structure notes

- Browser/Preview authority: `folio-designer/src/App.tsx`, `engine-client.ts`, `engine-protocol.ts`, `engine.worker.ts`, and `src/preview/{freshness,pdf-viewer}.ts(x)` already carry immutable buffers, FIFO worker requests, producer identity, and PDF.js admission.
- Go/Wasm/native production seams: `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go`, `folio-go/render_entry.go`, `folio-go/preview_identity.go`, `folio-go/cmd/folio/main.go`, and `folio-go/fonts/`.
- Canonical template and diagnostics authority: `folio-go/internal/template`, `folio-go/diagnostic.go`, `folio-go/internal/diag/diag.go`, and existing `render_error_test.go` / `matrix_test.go` conventions.
- Reuse completed Epic 6 behavior from Stories 6.1–6.6; no product UI change is presumed. Expected delivery is integration fixtures/harness/tests and a precise evidence record, with production changes only if the real proof exposes a narrow existing-authority defect.

## References

- `_bmad-output/planning-artifacts/epics.md` — Story 6.7 AC source and Epic 6 end state.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR35, FR38, FR41, NFR1, NFR5.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-8, AD-9, AD-14–18, AD-23.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — exact Preview, stale authority, local/offline behavior.
- `_bmad-output/implementation-artifacts/epic-6-context.md` and `folio-mvp-decision-log.md` — Epic dependencies, D-000.4 integration exception/boundary discipline.
- `_bmad-output/implementation-artifacts/6-1-load-sample-data-and-browse-its-paths.md` through `6-6-present-a-failed-render-honestly.md` — accepted data/parameter inputs, binding/table workflow, Preview and diagnostic authority.

## Delivery log

### Story creation — 2026-08-28

- Created from baseline `e314c14` after grounding in the complete Epic 6 source/context, sprint tracker, PRD, architecture spine, UX DESIGN/EXPERIENCE, decision log, Stories 6.1–6.6, and current Go/Wasm/designer production seams.
- Confirmed `wasm.Engine` already owns canonical template bytes and renders only a caller-supplied byte-identical current template with raw data/params; the worker transports copied opaque buffers and Preview admission remains separate from canvas approximation.
- Confirmed the native CLI and public `folio.Render` with `fonts.Shipped()` are the production comparison paths, and `internal/diag` is the constructed closed registry requiring census rather than a duplicated list.
- D-000.4’s ordinary cadence is overridden only for this intrinsically integration-shaped Story 6.7: real compiled browser/native round-trip evidence is required now. The complete four-target matrix and any release tag remain Epic 6 boundary work.
- Implementation has not started. No task is checked, no test has been claimed, no implementation/commit is authorized by this story record, and unrelated dirty workspace changes remain out of scope.

### Integration delivery — 2026-08-29

- Added the compiled Playwright `browser-native-roundtrip` proof. In two distinct fresh browser contexts it forces the testable fallback picker, loads raw local data, authors through palette/binding/tree/table/editor/footer/save controls, records worker operations, saves the engine `serialize` bytes, requires the visible PDF.js `EXACT LOCAL PRODUCTION PDF` admission, and hands the captured files to the public native CLI and Go library witness. It never opens a prepared `.folio` in either browser session.
- The 2026-08-29 Darwin/arm64 run used Node `v24.16.0` and Chromium `1228` at `/Users/panitw/Library/Caches/ms-playwright/chromium-1228/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing`; `browser-native-roundtrip.spec.ts` passed in `1.6m`. Retained evidence is `folio-designer/test-results/browser-native-roundtrip-f-d5f43-ed-Preview-and-native-Folio/browser-native-roundtrip/evidence.json`: golden template `1499`/`27500753a2e4ace159d7905aa405e6b5e3741cccb51a329abd3f32e9d54592eb`, data `239`/`a1a10748ed4a30b8ce111ea255949945949cdadf0c31a288e47bb527af0534e8`, params `75`/`27dfa599a1d2c58266d3318a6a19a1c45f239fe0e62fa907e262e4f7eda11d9f`, and browser/native PDF `61614`/`71c0fa867cfdf895c80d43b48ccef17943422e8a12b40e009d157ba5d3851de7`; alternate template `973`/`f7b8cdc4d8007b5496576d6931be0f386f90826ea340a2c26cbd229f62660395` and browser/native PDF `52768`/`e06d726c0ce9e5c65092ac5fb40f37165d423f10b983d99d3cd7910f37f02a8a`. The witness also parses/renders the separate canonical hand-edited `testdata/template/golden/worked-example.json` only in native Go.
- The proof exposed and fixed the Story 6.3 worker-protocol regression: `wasm.Engine.ParameterReferences` copied an empty result with `append([]string(nil), ...)`, which JSON-marshaled to `null`; the closed TypeScript protocol requires `parameterReferences.names` to be an array, so the client failed before identity/render. It now returns a non-nil empty slice, guarded by `TestEngineEmptyParameterReferencesRemainAnArrayForWorkerTransport`. No tracing or product telemetry remains.
- The proof also exposed the normal Table Editor's missing-new-column usability seam: a new column with an empty bind now projects as `rowFieldEditable`, allowing the existing Go command boundary to receive its first row field. The authored golden keeps a string `amountDisplay` cell binding and a separate numeric `transactions.amount` footer source, preserving both strict text-binding and numeric aggregation semantics.
- Added a dynamic Go diagnostic registry census. It derives the registered error set from `internal/diag` at runtime, demands a real production trigger for every error code, asserts code/message/location transport, and separately proves warning diagnostics retain successful PDF bytes.
- Verification: `npm run lint` passed with only four pre-existing Fast Refresh export warnings; `npm test` passed `27` files / `162` tests; `npm run build`, `npm run verify:offline`, and `npm run test:e2e:compile` passed; focused Wasm protocol/parity and registry tests passed; `go vet ./...` and `git diff --check` passed. `go test ./...` has the mandated existing P6g corpus-floor red only: opaque names `7`, required `>=20` (`1252` passed, `2` failed, `5` skipped). The Epic 6 four-target matrix remains deliberately unrun and unclaimed.

### Finisher completion — 2026-08-29

- Closed all six review findings. The compiled browser proof now starts with `initialize` on an empty Go-owned starter, retains every command request/payload in both session histories, and rejects any later fixture `load`/`initialize` route.
- The primary document is a browser-authored Customer Account Statement with embedded logo, customer/account/period expressions, Thai/CJK data, five table columns, generated-date/confidentiality and `Page {{page}} of {{pages}}` footer. Go parses its saved bytes and asserts those facts before native render.
- Preview copies the exact worker render template/data/params before transfer, compares them directly with the saved canonical/download and accepted raw inputs, repeats a forced browser render, then gives those exact bytes to CLI and native witness. The separately authored alternate uses a different session identity, command topology, bands, and repeated parity chain.
- `evidence/story-6.7-roundtrip-manifest.json` is the committed bounded retention record: raw small inputs, hashes, startup and exact command histories. `folio-designer/test-results/` is ignored, so PDFs and the transient native executable remain disposable.
- Gates: explicit Chromium full Playwright **20/20 passed**; focused statement proof **1/1 passed**; Vitest **162/162 passed**; typecheck, lint (four pre-existing Fast Refresh warnings), E2E compile, build, offline verification, offline Wasm witness, offline red proof, `go vet`, `gofmt -d`, and `git diff --check` passed. `go test ./...` is the sanctioned known red only: **1,259 passed, 2 failed, 5 skipped**, P6g opaque names **7 < 20**. Epic 6 remains `in-progress`; the four-target matrix remains its separate boundary gate.

### Review Findings

- [x] [Review][Patch][Blocker] **Author the defined Customer Account Statement, not a two-component toy, before claiming the golden-report round trip.** [`folio-designer/e2e/browser-native-roundtrip.spec.ts:7`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L7) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:91`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L91) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:203`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L203)
  - **Category:** AC1 · S5 · Epic 6 end state · acceptance-report identity
  - **Observation:** The browser session authors one `customer.name` text element and a two-column table, both in the Content band. It never authors the acceptance report defined by PRD §9.1 / Story 4.7: logo, customer and account blocks, statement period, five-column Date/Description/Debit/Credit/Balance table, confidentiality footer, visible `params.reportDate`, `Page X of Y`, mixed Latin/Thai/CJK, wrapped content, and an embedded image. The supplied `reportDate` is unreferenced; only reserved `documentDate` can affect the result. The retained template confirms the text and table also occupy the same origin rather than forming the report.
  - **Impact:** The green browser/native equality closes S5 over a small surrogate, not “the golden report authored in this session.” The central Epic 6 acceptance claim remains unmeasured even though the story marks AC1 complete.
  - **Required patch:** Drive the real golden Customer Account Statement through normal browser controls from the empty starter, retain Go-owned facts proving its required structure/bindings/params use, and run the exact saved-byte → admitted PDF.js Preview → native CLI/library byte comparison on that authored report. Do not seed or reopen the Story 4.7 fixture.

- [x] [Review][Patch][High] **Make browser authorship reject a prepared startup document and retain auditable command evidence.** [`folio-designer/e2e/browser-native-roundtrip.spec.ts:18`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L18) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:203`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L203) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:252`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L252)
  - **Category:** AC1 red proof · non-fixture authorship · generated evidence integrity
  - **Observation:** Each session sets its baseline only after startup and slices away the initial `initialize`/`load` traffic. It then proves only that at least 8/3 opaque `command` operations occurred; command payloads, revision transitions, initial Go-owned empty shape, and resulting command history are neither asserted nor written to `evidence.json`. The retained evidence records total request and serialization counts only.
  - **Impact:** Replacing the starter with prepared report bytes and applying enough unrelated or no-op commands can keep the proof green. The evidence file cannot later demonstrate which controls built which document, directly missing AC1’s non-vacuous session evidence and its explicit initialize/fixture red proof.
  - **Required patch:** Assert the initial Go projection is the intended empty starter, inspect the captured startup operation instead of excluding it, capture the exact committed command payload/order and revision advances, assert those commands account for the required authored facts, and persist a bounded auditable command/session manifest with the evidence.

- [x] [Review][Patch][High] **Capture the exact render-request template, data, and parameter bytes before native comparison.** [`folio-designer/e2e/browser-native-roundtrip.spec.ts:32`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L32) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:49`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L49) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:239`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L239)
  - **Category:** AC1 · exact raw channels · same-input proof · red-proof discrimination
  - **Observation:** The Worker observer captures only request operation/id and response bytes. Native input files are then written from the test’s original `sample`/`params` constants and saved download, not from the actual browser `render` request payload. No assertion proves the browser sent that saved template array or byte-identical accepted data/parameter arrays.
  - **Impact:** A browser-side normalization, reconstruction, stale input, or wrong template buffer can render one PDF while native receives different test-authored files; the harness would report only the resulting PDF mismatch, and can even pass when the changed channel is semantically irrelevant. The required one-byte template/data/params red proofs and exact-decimal raw-input claim are absent.
  - **Required patch:** Copy the three opaque `render` payload buffers in the observer before transfer, compare them directly against the saved download and accepted picker/editor bytes, persist their lengths/hashes, feed those captured files to both native CLI and library witnesses, and add discriminating one-byte/reconstruction mutations that fail on direct bytes rather than digest alone.

- [x] [Review][Patch][Medium] **Prove a repeated, independently authored alternate shape rather than a shared skeleton with a type swap.** [`folio-designer/e2e/browser-native-roundtrip.spec.ts:220`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L220) [`folio-go/browser_roundtrip_witness_test.go:100`](../../folio-go/browser_roundtrip_witness_test.go#L100)
  - **Category:** AC2 · S7 · structural independence · deterministic browser repeat
  - **Observation:** Both sessions use the same page, bands, data, params, `customer.name` text at the same origin, and overlapping fresh-document ids (`e1`/`e2`). The alternate merely replaces the golden table with a rectangle. Go reduces “shape” to a histogram of `band/type`, command evidence is only a count, and each browser document is rendered once; only the native library is repeated.
  - **Impact:** The proof does not satisfy AC2’s distinct component-id/command-history discriminator or repeated browser/native-render clause, and it gives weak evidence that a genuinely different report topology crosses the same full chain.
  - **Required patch:** Author a second report with materially different Go-owned band/component topology and deliberately non-overlapping authored ids/history, retain that history, assert structural facts beyond a type-count histogram, and compare repeated admitted browser renders as well as repeated native renders with unchanged captured inputs.

- [x] [Review][Patch][Medium] **Make the diagnostic census exact for every registry member and every applicable location.** [`folio-go/internal/diag/diag.go:251`](../../folio-go/internal/diag/diag.go#L251) [`folio-go/diagnostic_registry_census_test.go:17`](../../folio-go/diagnostic_registry_census_test.go#L17) [`folio-go/diagnostic_registry_census_test.go:79`](../../folio-go/diagnostic_registry_census_test.go#L79) [`folio-go/diagnostic_registry_census_test.go:107`](../../folio-go/diagnostic_registry_census_test.go#L107)
  - **Category:** AC3 · dynamic full-registry census · warning-success semantics · actionable locations
  - **Observation:** `ErrorCodes` derives coverage from a second hard-coded `dispositions` map, so a newly registered code classified as Warning needs no production trigger. Seven warning-classified codes exist, but the suite exercises only `TEXT_MISSING_GLYPH`. Error messages need only one non-space byte, and the generic “element or path” check lets `BINDING_PATH_ABSENT` pass without its applicable data path; no exact per-code location contract is asserted.
  - **Impact:** A dormant new warning, a wrong disposition, a meaningless message, or lost binding path can all leave the census green. That falls short of the story’s runtime-derived complete registry sweep and its red proofs for unexercised membership, changed meaning, location, and warning success.
  - **Required patch:** Key an exact trigger/expectation map against every value returned by `diag.All()`, execute every warning through successful `Render` with PDF bytes as well as every stopping error, assert each code’s exact applicable element/path and meaningful stable message contract, and make set equality fail for any untriggered registry member without a separately maintained severity escape hatch.

- [x] [Review][Patch][Medium] **Give retained integration evidence a durable, bounded lifecycle outside disposable Playwright output.** [`folio-designer/playwright.config.ts:7`](../../folio-designer/playwright.config.ts#L7) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:200`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L200) [`folio-designer/e2e/browser-native-roundtrip.spec.ts:245`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L245)
  - **Category:** evidence retention · repository hygiene · reproducible release record
  - **Observation:** `preserveOutput: 'always'` is already Playwright’s default and does not make `test-results` durable: the next Playwright invocation clears/replaces that directory. The output is unignored and currently retains templates, data, params, duplicate PDFs, an absolute workstation path, and a roughly 20 MB native CLI binary, while only the small JSON manifest is attached.
  - **Impact:** The delivery log calls this a retained release record, but it is both easy to erase accidentally and expensive/noisy to keep wholesale. Subsequent focused runs can silently replace the claimed evidence, while an accidental add sweeps a platform binary and transient artifacts into the story.
  - **Required patch:** Define an explicit retention policy: persist only a minimal immutable manifest and any deliberately selected byte witnesses in a stable scoped location (or CI artifact), exclude the built CLI/transient duplicates/absolute paths, ignore disposable `test-results`, and verify the retained manifest against current files when present.

### Review Summary

- **Outcome:** **Changes requested.** Six unresolved patch findings remain: one blocker, two high-severity proof gaps, and three medium-severity coverage/evidence gaps. Story and sprint tracking returned to `in-progress`.
- **Real browser verdict:** The focused `browser-native-roundtrip.spec.ts` passed with the explicitly available Chromium 1228 binary in **1.6 minutes**; the browser step itself completed in about **6 seconds**. Independent `cmp`/SHA-256 checks reproduced exact browser/native CLI equality for both current authored shapes. PDF.js admission reached `EXACT LOCAL PRODUCTION PDF` in both sessions.
- **Measured designer gates:** `npm test` passed **162 tests across 27 files**; typecheck, E2E-source compile, production build, ordinary offline verification, offline red proof, and Wasm offline witness passed. Lint reported only the four recorded Fast Refresh warnings. An additional full real-browser audit passed **13/20** scenarios and exposed seven older compiled-only boundary failures (five capability-picker waits under this Chromium, one stale assertion, one ambiguous locator); these are reported here, not misrepresented as a green full-browser gate.
- **Measured Go/static gates:** focused registry census passed **8 error subtests plus its warning witness**; the browser-authored public-library witness and focused Wasm protocol/parity tests passed. Native and js/wasm builds, `go vet ./...`, `gofmt -d .`, and `git diff --check` passed. Full `go test ./...` reproduced only the sanctioned corpus-floor red: **1,252 passed, 2 failed, 5 skipped** (`P6g opaque names: 7`, required `>=20`).
- **Review execution:** The commissioning handoff explicitly prohibited subagents, so the adversarial, edge-case, verification-gap, and acceptance lenses were applied directly. No product-code fix, fixture/golden/font change, decision-log amendment, commit, push, or remote operation was performed.
- **Boundary status:** The four-target Epic 6 matrix remains deliberately unrun and unclaimed. The extra full-browser audit is red and must not be described as an Epic 6 boundary pass.
- **Cleanup:** Reviewer-created temporary native/js-wasm build outputs and extra Playwright failure output are disposable and removed/restored to the focused witness before handoff. Unrelated `_bmad` configuration/manifest changes, `.agents/`, and planning research remain untouched.
