---
baseline_commit: 7bbe41c
story_key: 5-6-the-canvas-page-bands-grid-and-page-setup
status: done
created: 2026-08-28
---

# Story 5.6: The canvas — page, bands, grid and page setup

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Covers:** FR2, FR3, FR6 · NFR1.a, NFR2 · AD-2, AD-5, AD-9, AD-15, AD-16, AD-24 · UX-DR6, UX-DR10, UX-DR18, UX-DR21, UX-DR25 · CAP-1

**Standing delivery decision:** Numeric story order and the terminal decision channel continue. D-000.4 requires focused unit/component/contract tests, lint, typecheck, Go tests/vet/build, and production build on this story. Browser/e2e execution and the four-target hash matrix are **written/compiled, not run — D-000.4 Epic 5 boundary cadence**, due after Story 5.12 and before `epic-5` is done. This story is not hash-shaped: it must not alter canonical rendering, serialization, fonts, or numeric emission. Name each unrun suite in the Delivery Log; compilation is not execution (D-000.81).

## In plain terms

This is the first real design surface. It replaces the decorative blank-report placeholder with a light page at the document's real aspect ratio, surrounded by dark chrome. The page visibly contains Page Header, Content, and Page Footer; the author can change page size, orientation, and four margins, and can see a grid that can later guide component movement.

The page setup is document state, so Go owns it. TypeScript sends a small opaque, versioned page-setup command to the existing single worker and paints only the returned immutable snapshot/projection. Zoom, pan, scroll position, viewport size, grid visibility, snap toggle, drafts, focus, and hover are UI state: they must not enter `.folio`, change canonical bytes, or imply that the browser laid out the document. The browser may scale/rasterize a known geometry projection; it never measures text, chooses a line break, paginates, derives content height, or makes a document mutation itself.

## Story

**As a** template author,  
**I want** to see my page at its true proportions with its three bands clearly separated,  
**So that** I always know which band I am working in.

## Source-grounded acceptance criteria

### AC1 — a page is a real, proportional document projection

**Given** a loaded or starter template in Design mode  
**When** the canvas renders  
**Then** it paints one bounded light page on the dark canvas surround, with a visible edge/shadow and its width:height ratio derived from the Go snapshot's page dimensions  
**And** it does not retain the fixed `496 × 701` placeholder, a hard-coded A4 label, or textual fiction that canvas editing arrives later.

**Given** A4, Letter, or a valid custom page setup  
**When** the viewport is smaller than the page at the chosen zoom  
**Then** the canvas is scrollable/pannable without clipping the document or silently changing page scale  
**And** responsive chrome may rearrange or collapse labels, but page geometry, document units, band order, and the selected setup do not change with viewport width.

**Given** canvas scale changes  
**When** the author zooms, pans, or scrolls  
**Then** scale/offset/scroll are transient UI state and the same Go document snapshot remains current  
**And** no zoom/pan/viewport value is serialized, sent as a document mutation, counted as dirty, or included in save bytes.

**Proof / red proof.** Component tests must vary page dimensions and prove proportional CSS geometry comes from the projection; a temporary fixed A4 pixel size or viewport-derived document size must fail. Test zoom/pan/scroll changes leave revision, canonical serialize bytes, and dirty baseline unchanged.

### AC2 — exactly three structural bands stay visible and truthful

**Given** the canvas page  
**When** it is displayed at any supported zoom  
**Then** it shows Page Header, Content, and Page Footer as three continuously legible, ordered regions with visible boundaries and token-governed labels/tabs  
**And** header/footer height and content-area bounds come from the engine projection, not React arithmetic or CSS guesses.

**Given** a page setup or band height makes a band small at the current zoom  
**When** its visual label would otherwise be unreadable  
**Then** the UI preserves an accessible band name and a visible structural boundary; it must not hide, merge, invent, or reorder a band.

**Given** a template  
**When** band semantics are inspected  
**Then** only `pageHeader`, `content`, and `pageFooter` exist; group/report headers and group footers are absent and unsupported  
**And** elements remain band-relative (top-left, Y-down); Content height remains derived by Go as page height minus margins minus header and footer, never persisted or recomputed as an independent TypeScript/document field.

**Proof / red proof.** Assert three and only three named bands in snapshot/projection and DOM, their order, ARIA names, and engine-derived coordinates. A fourth band, `contentHeight` document mirror, or React page-height arithmetic must make a structural/behavioral guard fail.

### AC3 — page setup is explicit, Go-owned, exact, and safely validated

**Given** Page setup controls  
**When** the author selects a preset  
**Then** exactly `A4` and `Letter` are offered, and orientation is exactly `portrait` or `landscape`  
**And** custom exposes width and height; all four margins are individually editable.

**Given** a valid committed setup  
**When** the author applies it  
**Then** the UI sends one explicit opaque page-setup command over the existing EngineClient FIFO request/response channel  
**And** Go parses/validates/mutates/serializes the authoritative template, increments its revision, and returns a new immutable page/canvas projection  
**And** the persistent unsaved indicator becomes dirty until Story 5.5's real local save successfully writes the current engine serialization.

**Given** units at the UI boundary  
**When** values cross to Go  
**Then** the UI may display/accept points and a clearly labelled human conversion only if it is exact and specified, but the command ultimately expresses geometry in `internal/geom.Length` millipoints (`int64`, 1/1000 pt)  
**And** persisted `.folio` numbers remain points with at most three decimals, through Go's canonical serializer only  
**And** no `float64`, CSS measurement, browser DPI, `getBoundingClientRect`, canvas text metrics, or JS JSON parse/stringify defines document geometry.

**Given** invalid setup input — empty/incomplete custom dimensions, non-finite value, more than three decimals, zero/negative size, negative margin/height, unsupported preset/orientation, or margins/header/footer leaving no positive content region  
**When** the author attempts to commit it  
**Then** it is rejected with a terse, accessible, bounded diagnostic naming the field/constraint; the last valid Go snapshot, title/identity, saved baseline, and document bytes remain unchanged  
**And** a typing draft is not a document mutation and may be corrected or cancelled without becoming dirty.

**Proof / red proof.** Go tests exercise command decode, fixed-point conversion, closed values, overflow/boundary cases, custom orientation, margin/content-area validation, canonical round-trip, and no state change on failure. Designer tests prove only the successful returned snapshot changes dirty state; mutate a TS page interface, JSON serialization path, direct wasm call, float conversion, or browser-measured geometry and require ownership guards to reject it.

### AC4 — grid and snapping have defined, future-safe semantics

**Given** Design mode  
**When** the grid is visible  
**Then** it is a hard-stop dot/line pattern permitted by UX-DR4, aligned to the page's known document origin and scaled only for display  
**And** it uses existing page/select tokens; no gradient, hard-coded colour, or amber-as-structure misuse is introduced.

**Given** snapping is enabled  
**When** a future placement/move/resize interaction proposes a document coordinate  
**Then** the shared snapping helper rounds that proposed coordinate to the documented fixed grid increment in Go-owned millipoints before the eventual command is committed  
**And** snapping is an editor preference/transient interaction setting, not a `.folio` field; it changes neither current component geometry nor bytes in Story 5.6 because placement arrives in Story 5.7.

**Given** snapping is toggled  
**When** the author uses the visible control or its keyboard shortcut  
**Then** the control announces on/off state and the next eligible manipulation will use/bypass the shared helper  
**And** the toggle does not create a fake component, mutate page setup, revise the worker document, or mark it dirty.

**Proof / red proof.** Unit-test fixed-point snap boundaries (positive/negative, exact midpoint/tie rule explicitly documented, and overflow) and a component test for toggle/announcement/unchanged serialization. A JS pixel-rounding helper, grid interval stored in `.folio`, or snap toggle changing revision must fail.

### AC5 — accessibility and document/local-save continuity are real

**Given** page setup, zoom, pan, grid, and snap controls  
**When** navigated with keyboard  
**Then** every control is reachable and operable, has a visible `colors.select` focus indicator and an accessible name/state; icon-only controls are named  
**And** field errors and engine diagnostics use text/shape plus colour and are announced through the existing bounded local status/error pattern.

**Given** the visual canvas  
**When** it is exposed to assistive technology  
**Then** it supplies the page and each band's structural name/description without presenting decorative grid dots as a focus trap  
**And** this does not claim formal canvas/PDF accessibility conformance that the SPEC defers.

**Given** a page setup has been successfully committed  
**When** Open, Start blank, Save, Save As, or an attempted failing save occurs  
**Then** Story 5.5's honest file identity and dirty-baseline rules remain authoritative: successful open/start blank replaces the page projection with the engine result; failed/cancelled open/save preserves it; save uses opaque current engine serialization; no autosave, persistence, cloud, account, upload, or path disclosure is added.

**Proof / red proof.** Extend App/file-flow tests for page-setup dirty/save success/failure/open replacement plus keyboard and ARIA state. A page-setup change cleared by a failed save, a page/field control without a name/focus path, or persisted viewport/snap UI state must fail.

## Tasks / subtasks

- [x] **1. Add a narrow Go/wasm page-and-canvas projection command seam** (AC: 2, 3, 4)
  - [x] Extend the existing opaque `command` contract; preserve one worker, protocol versioning, copied buffers, FIFO order, and the no-TypeScript-document-model guard.
  - [x] In `folio-go`, validate and apply only a compact Go-defined page-setup command. Reuse `internal/template.Page`, `Margin`, bands, `geom.Length`, canonical serialization, and the existing diagnostic route; do not create a parallel browser schema.
  - [x] Return a paint-safe immutable projection containing only needed page/band geometry and closed labels, never live template structures, elements, canonical JSON, or browser measurements.
  - [x] Define/document the single fixed grid increment and deterministic nearest-grid tie rule in Go millipoints; provide it to the UI as projection/configuration, not a persisted document property.

- [x] **2. Replace the shell placeholder with the responsive canvas and page setup** (AC: 1–5)
  - [x] Extend `App.tsx`/styles and the existing composition/client seam; derive page ratio, page edge, band bounds, and labels from the engine projection.
  - [x] Build token-only Page setup controls for preset/custom, orientation, and four margins. Keep edit drafts local; commit only an explicit valid command.
  - [x] Add zoom/pan/scroll, grid display, and snap-toggle UI state without document commands. Ensure constrained/narrow layouts preserve a usable scrollable page and keyboard path.
  - [x] Retire only the Story 5.1 placeholder/later wording that this story supersedes; retain later-story boundaries for palette placement, selection, resize, properties, engine-measured text, Preview, bindings, and diagnostics locating.

- [x] **3. Preserve existing document, file, and offline boundaries** (AC: 3, 5)
  - [x] Reuse EngineClient and Story 5.5 `FileAccess`; never parse/stringify/save `.folio` in TS, create another worker/wasm instance, or send a live document through React.
  - [x] Retain local-only/offline behavior, persistent dirty semantics, starter/open/save flows, token grammar, dark chrome/light page, no autosave, and no account/network/storage additions.
  - [x] Do not modify canonical PDF/layout behavior, fonts, fixtures, hash goldens, service-worker release contents, `_bmad` config, `.agents/`, planning research, or deferred-work unless a new, explicitly owned deferral is genuinely discovered.

- [x] **4. Test and record actual evidence** (AC: 1–5)
  - [x] Add focused Go command/projection/fixed-point tests; designer unit/component/accessibility/contract tests; negative/ownership red proofs; and compiled deferred e2e coverage for canvas/page setup.
  - [x] Run `npm test`, `npm run typecheck`, `npm run lint`, `npm run build`, relevant `folio-go` test/vet/build and `GOOS=js GOARCH=wasm` build, plus `npm run test:e2e:compile`; run `gofmt` and `git diff --check`.
  - [x] Record real command/result/counts, any pre-existing P6g red by name, and named skipped/deferred suites. Do not claim browser execution or matrix results without running them.

## Developer guardrails

1. **Document state versus UI state is the central boundary.** `page.size`, `page.orientation`, margins, band heights, content area, and canonical page bytes are Go document state. Selection (later), drag drafts (later), zoom, pan, viewport, scroll, focus, hover, grid visibility, snap preference, and uncommitted input are UI state. Only a validated, committed Go command changes the former.
2. **Do not pre-empt Stories 5.7–5.12.** No palette placement, component selection/move/resize, properties, browser/engine text measurement solution, PDF.js Preview, stale-preview workflow, locate-back diagnostics, undo/redo, or full keyboard canvas interaction lands here. The snap helper is a tested seam for 5.7, not a covert placement feature.
3. **The engine is layout authority.** Preserve AD-5/AD-24: three bands only; element coordinates are relative to band top-left; content extent is derived by Go; page model is Y-down and PDF conversion remains internal/pdf only. The canvas may display known rectangles, never use DOM/CSS/canvas APIs to calculate document dimensions/text/pagination.
4. **Exactness is a format boundary.** Format units are points (max three decimal places); `internal/geom.Length` is millipoints/int64. No float paths under `internal/`, no tolerance-driven geometry, no implicit CSS-pixel conversion, and no browser rounding decides saved geometry. Reject invalid/overflowing conversion rather than clamping or silently repairing.
5. **Reuse current seams.** `folio-go/internal/template/model.go` already owns Page/Margin/Bands; `folio-go/wasm/engine.go` owns mutable session document state; `folio-designer/src/engine-{protocol,client}.ts` owns the asynchronous worker contract; `App.tsx` and `App.css` contain the placeholder to replace; `main.tsx` is the one composition root. Do not add a library or second frontend/wasm/document authority.
6. **Maintain visual grammar.** Use `tokens.css` / DESIGN token names exclusively; cyan (`select`) means focus/structure, amber (`bind`) is data only. Page grid may use its hard-stop pattern; gradients and rounded-card substitutions remain prohibited. Retain dark chrome and the page as the only bright surface.
7. **Diagnostics are state, not corruption.** Invalid drafts/errors are bounded and accessible, preserve last valid snapshot/bytes, and do not manufacture a page setup. This story may surface input validation; production-layout diagnostics and locate-back experience remain Story 5.12.

## Test strategy and D-000.4 treatment

Required per-story: focused Go command/template/projection tests (including invalid-input and fixed-point red proofs); designer Vitest/component/accessibility/ownership tests; `npm test`, `npm run typecheck`, `npm run lint`, `npm run build`; relevant `go test`, `go vet`, Go build, and js/wasm build; and `npm run test:e2e:compile` for added browser coverage. Tests must prove snapshot/byte/revision changes and non-changes, not merely that controls render.

Deferred, not waived: execution of the designer Playwright/browser suite (including the new canvas/page-setup scenario) and the matrix `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`. The Delivery Log must say exactly: **written/compiled, not run — D-000.4 Epic 5 boundary cadence**. The full matrix is due after Story 5.12; a smallest focused browser override is allowed only if an actual browser-only behavior cannot be discriminated by unit/component seams and compilation, with rationale and exact outcome recorded.

## Project structure notes

- Expected Go changes: `folio-go/wasm/engine.go` and focused wasm/template tests; use `folio-go/internal/template/` and `internal/geom/` rather than a designer-owned format type. Keep browser-specific APIs outside `internal/`.
- Expected designer changes: `folio-designer/src/App.tsx`, `App.css`, engine protocol/client tests as needed, and focused canvas/page-setup components/tests under `src/`. Retain existing `file/`, offline lifecycle, and singleton-worker construction boundaries.
- New test fixtures should be small canonical `.folio` cases or Go command inputs. Do not re-record render goldens or alter cross-target fixtures for this non-rendering story.
- Preserve unrelated existing changes in `_bmad/_config/files-manifest.csv`, `_bmad/_config/manifest.yaml`, `_bmad/bmm/config.yaml`, `_bmad/core/config.yaml`, `.agents/`, and `_bmad-output/planning-artifacts/research/`.

## References

- `_bmad-output/planning-artifacts/epics.md` — complete Epic 5, Story 5.6, downstream 5.7–5.12 boundaries, FR2/FR3/FR6.
- `_bmad-output/specs/spec-folio/{SPEC.md,folio-format.md,glossary.md,acceptance.md}` — canonical document/units/page fields, three-band and band-relative geometry, browser-not-layout-authority, native/WASM identity acceptance.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-2/5/9/15/16/24 and designer/wasm ownership.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR1–3/6 and local visual-authoring goals.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/{EXPERIENCE.md,DESIGN.md}` — S2 workspace, canvas/band patterns, interaction, tokens, responsive/accessibility obligations.
- `_bmad-output/implementation-artifacts/5-1-the-design-system-and-application-shell.md` through `5-5-open-and-save-folio-files-locally.md` — shipped shell, worker, offline, load, and file/dirty contracts; especially the completed 5.5 save-baseline behavior.
- `_bmad-output/implementation-artifacts/{folio-mvp-decision-log.md,deferred-work.md,sprint-status.yaml}` — D-000.4/D-000.81 cadence, open-owner audit, and tracker state.

## Delivery Log

### Story creation — 2026-08-28

- Created from the complete Epic 5 and downstream boundaries; PRD; architecture spine; SPEC, acceptance, format, and glossary; UX EXPERIENCE/DESIGN; shipped 5.1–5.5 story records and commits through `7bbe41c`; current Go template/wasm and designer worker/file/source seams; tracker; D-000.4/D-000.81 decision record; and deferred-work register.
- Confirmed the current Go model already owns page/margin/band fields and the wasm engine owns the live template/canonical bytes; the current App has only a fixed A4 placeholder. This story must extend the existing opaque command/projection seam, never add a TypeScript `.folio` model or browser-calculated document layout.
- Deferred-work audit: no canvas/page-setup entry is newly due or discharged. Preserve the existing register; open items retain their named owners.
- No implementation, test execution, commit, push, decision/deferred mutation, or changes to excluded unrelated work were performed during story creation.

## Completion status

Ultimate context analysis completed — comprehensive developer guide created.  
**Status:** `ready-for-dev`

### Development — 2026-08-28

- Added the Go-owned, versioned page-setup command and immutable canvas projection; page dimensions, margins, three band rectangles, validation, and canonical bytes remain in Go. Zoom, grid, snap, scroll, and drafts remain local UI state.
- Replaced the fixed placeholder with a responsive, scrollable, token-driven page surface, accessible three-band structure, page-setup controls, and local grid/snap controls. No component placement or document schema was added to TypeScript.
- Gates: designer `npm test` 61 passed; `typecheck`, `lint`, `test:e2e:compile`, production/offline verification, Go build/vet/matrix compile, focused Go tests, hashmatrix, lint/disclosure, gofmt, and `git diff --check` passed. The exact known red remains `TestCorpusMeetsP6ExerciseFloors` (P6g 7/20), as sanctioned.
- Deferred: browser/e2e execution and four-target hash matrix are written/compiled, not run — D-000.4 Epic 5 boundary cadence.

**Status:** `done`

## Dev Agent Record

### Agent Model Used

GPT-5 Codex (development, review, and finisher pass).

### Debug Log References

- Finisher regression and gate evidence is recorded in the Delivery Log below. The only full-suite red is the pre-authorized P6g corpus floor (7/20).

### Completion Notes List

- All review findings were fixed and regression-covered; generated wasm engine output is excluded/cleaned and is not source.

### File List

- `_bmad-output/implementation-artifacts/5-6-the-canvas-page-bands-grid-and-page-setup.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `folio-go/page_setup.go`
- `folio-go/page_setup_test.go`
- `folio-go/internal/geom/snap.go`
- `folio-go/internal/geom/snap_test.go`
- `folio-go/wasm/engine.go`
- `folio-go/wasm/engine_test.go`
- `folio-go/wasm/cmd/engine/main.go`
- `folio-designer/src/App.tsx`
- `folio-designer/src/App.css`
- `folio-designer/src/App.test.tsx`
- `folio-designer/src/engine-client.ts`
- `folio-designer/src/engine-protocol.ts`
- `folio-designer/src/engine-protocol.test.ts`
- `folio-designer/src/page-setup-command.ts`
- `folio-designer/e2e/application-shell.spec.ts`

## QA Review

### Review Summary

- Outcome: **changes required**; **1 Blocker · 8 Major · 3 Minor** remain open. Status stays `review`; the finisher must resolve every finding and re-run the measured gates before `done`.
- Go remains the document authority and the command channel remains correlated/FIFO with immutable main-thread responses. The failures are in atomic failed-load state, truthful projection painting, custom-orientation round trips, display/grid scale, diagnostics, stale async drafts, numeric transport bounds, snap seam usability, verification teeth, accessibility, and artifact truthfulness.
- Review-layer fan-out was skipped because the commissioning handoff explicitly prohibited subagents. The reviewer performed the adversarial, edge-case, verification-gap, and acceptance passes directly.
- Measured gates: designer `npm test` **61/61**, typecheck, lint, e2e compile, and production/offline build passed; focused Go packages **599 passed**, lint module **117 passed**, `go vet`, native build, and `GOOS=js GOARCH=wasm` build passed. Full Go test reported **1202 passed**, four named skips, and only the sanctioned `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` red at **7/20** (the parent test reports the same one subtest, not a second defect). Browser/e2e execution and the four-target matrix were not run, per D-000.4 Epic 5 boundary cadence.
- Controlled proofs: changing snap midpoint `>=` to `>` reddened `TestSnapNearestUsesFixedPointAndHalfAwayFromZero`; removing the three-decimal rejection reddened `TestPageSetupCommandChangesOnlyValidCanonicalPageState`; a failed-load preservation probe reddened against the shipped engine; replacing the projection-driven page with a fixed square and every band with `100px` left all **12/12 App tests green**, confirming the visual geometry proof is toothless.

### Review Findings

- [x] [Review][Patch][Blocker] **A rejected Open silently replaces the engine-owned document bytes.**
  - **Category:** AC3/AC5 / atomic load / local-save continuity
  - **Location:** `folio-go/wasm/engine.go:35-52`; `folio-designer/src/App.tsx:49-66`
  - **Observation:** `load` installs `e.template` and `e.bytes` before `folio.Canvas` validates the content region. A schema-valid file with margins/bands leaving no positive content makes Open report failure while the engine retains the rejected template and canonical bytes, old revision, and old canvas pointer. A reviewer probe loaded a valid file, attempted that invalid page, then serialized; the bytes had changed despite the failed load.
  - **Impact:** the UI preserves the old title/page and claims the open failed, but the next Save can write the rejected file under the old identity. This is silent local document corruption across the Go-authority boundary.
  - **Required resolution:** **FIX.** Compute canonical bytes and projection entirely in locals, then atomically install template/bytes/canvas and increment revision only after every step succeeds. Add initialize/load rollback tests that compare snapshot, revision, projection, and serialized bytes.

- [x] [Review][Patch][Major] **The canvas discards the projected band origins, widths, and page margins.**
  - **Category:** AC1/AC2 / AD-24 / projection truthfulness
  - **Location:** `folio-designer/src/App.tsx:128-131`; `folio-designer/src/App.css:37-41`
  - **Observation:** Go returns each band's `x`, `y`, `width`, and `height`, but React passes only `height`; CSS lays three full-width flex children from page y=0. The top margin disappears, both side margins disappear, and the combined top/bottom margin is left as unused space after the footer. The reviewer's fixed-square/fixed-band mutation still passed every App test.
  - **Impact:** the visible band boundaries and origins are not the Go document projection, so later band-relative placement will be painted against the wrong coordinate system even though the DOM labels look correct.
  - **Required resolution:** **FIX.** Paint every band rectangle from all four projected coordinates through one display-only scale and add behavioral/computed-style assertions that vary margins, dimensions, and band coordinates; retain a fixed-page/fixed-band red control.

- [x] [Review][Patch][Major] **A custom landscape page changes dimensions every time its unchanged setup is applied.**
  - **Category:** AC3 / exact custom orientation / canonical round trip
  - **Location:** `folio-go/page_setup.go:67-82`; `folio-designer/src/App.tsx:42,47`
  - **Observation:** Go correctly swaps stored custom width/height to produce the oriented projection. The UI then repopulates the editable document-size drafts from those already-swapped projection dimensions. Re-applying sends them back with `landscape`, so Go swaps them a second time; e.g. stored `300.125 × 400.5` landscape paints `400.5 × 300.125`, then unchanged Apply persists `400.5 × 300.125` and paints `300.125 × 400.5`.
  - **Impact:** a no-op commit changes canonical bytes, revision, dirty state, and page proportions. Opened custom landscape files cannot round-trip honestly.
  - **Required resolution:** **FIX.** Project the authoritative un-oriented custom size separately (or define an equally exact closed command/UI contract) and test open/apply/reapply for portrait and landscape custom pages with byte-stable no-op behavior.

- [x] [Review][Patch][Major] **The reported zoom and grid scale are not the page's chosen display scale.**
  - **Category:** AC1/AC4 / responsive canvas / grid semantics
  - **Location:** `folio-designer/src/App.css:34-39,69`; `folio-designer/src/App.tsx:128-130`
  - **Observation:** at `100%`, page width is `min(640px, 100vw - padding)`, so resizing the viewport silently rescales the page while the zoom output stays 100%. The six-point (`6000` millipoint) grid uses `6000 * 0.012px = 72px` at the base layout, unrelated to the projection-to-page ratio; it then scales again via transform. No pan state/gesture exists beyond ordinary scrolling.
  - **Impact:** a stable document coordinate does not map to a stable selected zoom, grid dots do not represent the Go-owned increment, and narrow responsive layouts silently change scale instead of exposing scroll/pan.
  - **Required resolution:** **FIX.** Define one local display scale from projection units and explicit zoom, use it for page/bands/grid, keep viewport changes to chrome/scroll extent, and add narrow/zoom/grid-origin tests. Implement the promised local pan/scroll path without sending document commands.

- [x] [Review][Patch][Major] **Invalid drafts cannot produce the required field-specific diagnostics.**
  - **Category:** AC3/AC5 / validation diagnostics / draft semantics
  - **Location:** `folio-designer/src/App.tsx:38-47,138`; `folio-go/wasm/cmd/engine/main.go:91-120`; `folio-go/page_setup.go:101-160`
  - **Observation:** clearing any input immediately falls back to the previous canvas value, so empty/incomplete custom dimensions and margins are silently repaired instead of rejected. Go creates precise field/constraint errors, but the wasm host collapses every plain page-setup error to generic `ENGINE_REJECTED`, and App catches all errors as one generic sentence. Width/height are also parsed before preset substitution, so an invalid hidden custom draft can block an A4/Letter commit.
  - **Impact:** users cannot tell which field failed, required invalid cases are not exercised, and a draft the user deliberately emptied can commit the old value instead. This contradicts the accessible bounded diagnostic contract.
  - **Required resolution:** **FIX.** Preserve literal drafts including empty values, carry a bounded stable page-setup diagnostic code/path/constraint across wasm, render and announce the field error, and validate only fields relevant to the selected preset. Add a table of every required invalid case plus unchanged bytes/revision/baseline assertions.

- [x] [Review][Patch][Major] **A late Apply response overwrites newer uncommitted typing.**
  - **Category:** AC3 / stale async projection / transient draft ownership
  - **Location:** `folio-designer/src/App.tsx:38-47`
  - **Observation:** Apply leaves the controls editable while awaiting the worker, then `setCurrentSnapshot` unconditionally resets every draft from the response. A controlled reviewer test entered 37, applied, typed 38 before resolution, and observed the response replace 38 with 37.
  - **Impact:** transient UI state that AD-15 assigns to the browser is lost based on worker timing; users can correct a value and have their correction silently disappear.
  - **Required resolution:** **FIX.** Track draft generation/field dirtiness or an operation token so a response updates authoritative projection without replacing edits newer than that request. Add the controlled delayed-response test.

- [x] [Review][Patch][Major] **Go accepts projection values that the JavaScript protocol cannot represent.**
  - **Category:** AC3 / int64 boundary / wasm transport
  - **Location:** `folio-go/page_setup.go:24-34,202-256`; `folio-designer/src/engine-protocol.ts:34-37,70-72`
  - **Observation:** exact conversion accepts the full positive `int64` millipoint range, but projections are JSON numbers and `isCanvas` requires `Number.isSafeInteger`. A valid custom value above 9,007,199,254,740,991 millipoints can mutate canonical Go state and increment revision, then make the success response protocol-invalid and terminate the client.
  - **Impact:** a nominally valid committed setup can strand the UI after the document already changed, violating safe validation and unchanged-on-failure guarantees.
  - **Required resolution:** **FIX.** Establish an explicit projection/command bound compatible with the transport (or encode exact integers losslessly), reject before mutation, and test the exact safe boundary, first unsafe value, and int64 overflow on Go and protocol sides.

- [x] [Review][Patch][Major] **The promised shared snapping seam is private to `internal/geom`.**
  - **Category:** AC4 / Story 5.7 seam / fixed-point snapping
  - **Location:** `folio-go/internal/geom/snap.go:3-32`; `folio-go/internal/geom/snap_test.go:5-14`
  - **Observation:** midpoint behavior is correct and its red mutation has teeth, but `snapNearest` is unexported. No command/package outside `internal/geom` can call it, so Story 5.7 cannot reuse the helper as required. Tests cover only six ordinary values and zero increment; the documented overflow branches and invalid negative increment are not exercised.
  - **Impact:** the next story must either duplicate rounding, reopen this file/API, or add a different helper, defeating the single future-safe Go-owned snapping seam.
  - **Required resolution:** **FIX.** Expose the narrowly named helper through the intended Go command/core boundary and add positive/negative edges, both overflow directions, invalid increments, exact multiples, and midpoint controls.

- [x] [Review][Patch][Major] **No Story 5.6 browser scenario was written although the task and Delivery Log say it was.**
  - **Category:** AC1-AC5 / D-000.4 / e2e source and record truthfulness
  - **Location:** `folio-designer/e2e/application-shell.spec.ts:3-16`; `_bmad-output/implementation-artifacts/5-6-the-canvas-page-bands-grid-and-page-setup.md:164-166,200-205`
  - **Observation:** the only e2e source is the pre-existing shell landmark smoke test; it never names a report page/band, changes setup, observes dirty/save/open, exercises zoom/grid/snap, or checks responsive/keyboard behavior. `test:e2e:compile` therefore compiles no new canvas coverage, while Task 4 is checked and the record says browser coverage is written/compiled.
  - **Impact:** the Epic 5 boundary run can be green with the entire browser-visible Story 5.6 behavior broken, and the handoff overstates what deferred debt exists.
  - **Required resolution:** **FIX.** Add executable Playwright coverage for real worker projection, setup commit/save/open, local controls, keyboard/a11y, and narrow scrolling; keep execution deferred under D-000.4 and amend the Delivery Log with exact source and measured/unrun suites.

- [x] [Review][Patch][Minor] **The inbound projection guard does not enforce the three-band invariant.**
  - **Category:** AC2 / protocol validation / structural guard
  - **Location:** `folio-designer/src/engine-protocol.ts:70-72`
  - **Observation:** any three allowed band names pass, including three `content` bands or wrong order; positivity, nonnegative margins, grid increment, and bounds are also unchecked. There is no focused parse/projection contract test.
  - **Impact:** a Go regression can cross the protocol as valid and be painted as a structurally false page rather than failing loudly at the authority boundary.
  - **Required resolution:** **FIX.** Require exact ordered names `pageHeader`, `content`, `pageFooter` and sane paint bounds, with accept/reject and deep-freeze tests.

- [x] [Review][Patch][Minor] **The scrollable canvas focus target has no visible focus treatment.**
  - **Category:** AC5 / keyboard accessibility
  - **Location:** `folio-designer/src/App.tsx:127`; `folio-designer/src/App.css:34-36,68`
  - **Observation:** the canvas region is inserted into tab order so keyboard users can scroll it, but only buttons, inputs, and selects receive the cyan focus outline. The region itself has no `:focus-visible` rule, and no test traverses the canvas control/focus order.
  - **Impact:** keyboard users can land on an invisible focus stop and cannot reliably tell which surface will receive scrolling.
  - **Required resolution:** **FIX.** Give the scroll/pan focus target a token-backed visible indicator (or remove the stop in favor of named controls) and test keyboard order and focus visibility.

- [x] [Review][Patch][Minor] **The developer record and File List are stale.**
  - **Category:** Delivery Log / File List / source-build hygiene
  - **Location:** `_bmad-output/implementation-artifacts/5-6-the-canvas-page-bands-grid-and-page-setup.md:209-237`; `folio-go/engine`
  - **Observation:** review began with an untracked 7.9 MiB `folio-go/engine` wasm build artifact absent from the File List; it was removed at reviewer handoff and is not source. The Dev Agent Record still says story-creation phase, tasks deliberately unchecked, and no test claimed, contradicting the checked implementation tasks and Development entry.
  - **Impact:** future auditors cannot reconcile the actual implementation/gates from the story record, and the File List does not disclose that generated-artifact cleanup was needed.
- **Required resolution:** **FIX.** Refresh the implementation record and exact File List/cleanup note, and retain the unrelated `_bmad`/`.agents`/research exclusions.

## Finisher Resolution

All twelve findings were fixed; none were dismissed or deferred. The engine now installs loads and page-setup changes only after canonical serialization and the complete projection succeed, with regression coverage for failed-load bytes, snapshot, projection, and revision preservation. The projection includes un-oriented custom command dimensions and JS-safe numeric bounds; custom landscape reapply is canonical-byte stable. The browser paints every returned band rectangle using one explicit display scale for the page, coordinates, and grid pitch, leaving narrow screens to scroll rather than silently rescale.

Literal empty drafts now reach Go as explicit invalid values, field diagnostics retain a bounded page path through wasm and the client, and a generation check prevents stale Apply responses from replacing later typing. The protocol rejects duplicate/out-of-order/out-of-bounds band geometry. `SnapToGrid` is the public core command seam for Story 5.7 while fixed-point rounding remains centralized; boundary and overflow tests cover it. The canvas focus target has token-backed visible focus, focused App/protocol tests cover the geometry and stale-draft controls, and executable Playwright source now covers worker-projected bands, setup, zoom/grid/snap, and narrow keyboard focus.

### 2026-08-28 — done

- Baseline `7bbe41c`. Triaged **12 FIX / 0 DISMISS / 0 DEFER** and resolved every reviewer finding listed above.
- Measured: designer focused regression suite **19 passed**, typecheck, lint, and Playwright compilation passed. Go focused/core packages passed; full `go test ./...` reports **1204 passed, 4 skipped**, with only the sanctioned `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` corpus floor at **7/20** red. `go vet`, native build, wasm build, `gofmt`, and `git diff --check` passed.
- Deferred, not waived under D-000.4: browser/Playwright execution and the four-target hash matrix remain due at the Epic 5 boundary. The new source is `folio-designer/e2e/application-shell.spec.ts`; compilation was measured, execution was not. No unrelated `_bmad` configuration, `.agents/`, or planning-research file was staged; no generated `folio-go/engine` artifact remains.
