---
baseline_commit: be7a688d5eae2e327ed60bc5da8749b7118a75a6
story_key: 5-9-a-canvas-the-browser-never-measures
status: in-progress
created: 2026-08-28
---

# Story 5.9: A canvas the browser never measures

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Covers:** NFR2 · NFR3 · AD-2, AD-15–17, AD-24 · UX-DR21

**Standing delivery decision:** Under **D-000.4**, run focused Go/wasm/designer unit and contract tests, typecheck, lint, build, vet, native/js-wasm builds, format/diff checks, and compile new browser e2e source. Designer browser/Playwright execution and the four-target hash matrix remain **written/compiled, not run — D-000.4 Epic 5 boundary cadence**, due after Story 5.12. This story adds a measurement projection but must not change canonical serialization, production PDF emission, or font assets; it is not itself a matrix override.

## In plain terms (read this first if you just want the gist)

The current design canvas can show a text component’s box, but it still uses a one-word placeholder. That is safe only as a temporary structural marker: a browser’s fonts and line-breaking rules can disagree with Folio about Thai words, CJK characters, kerning, and where a line fits. If the canvas were allowed to wrap text on its own, an author could place content based on a view that is systematically wrong about where the production engine will put it.

This story replaces that placeholder with an engine-produced paint plan. The engine decides the exact lines, their positions, and their metrics using the same shaping, font fallback, break opportunities, and vertical model that production uses. The browser receives that finished plan, maps its fixed document coordinates to pixels for display, and rasterizes it. It does not measure text, calculate a content height, infer a line break, or feed screen geometry back into the document.

This remains a design canvas, not the final PDF preview. It does not add data binding, finished table contents, a screen-reader model for the canvas, PDF rendering, responsive/mobile layouts, or a second rendering engine. Selection, focus, zoom, pointer position, and generous hit targets remain local interaction state. Done means the canvas shows pre-broken, engine-measured multi-line text at the right relative locations across viewport sizes, device-pixel ratios, and browser font availability, while changing any of those conditions leaves document bytes, revision, saved baseline, and engine geometry untouched.

## Story

**As a** template author,  
**I want** the canvas to wrap text where the engine would wrap it,  
**So that** the approximate view is approximate in scale, not wrong about where my content lands.

## Source-grounded acceptance criteria

### AC1 — the engine returns a bounded, paint-only text projection

**Given** a loaded template with a visible text component,  
**When** the worker produces its immutable snapshot,  
**Then** Go derives a closed text-paint projection from the engine’s existing shaping, font-chain resolution, `text.Opportunities`, `packLines`, and declared-chain vertical model—not from a browser callback or a second metric implementation.

**And** each text component’s projection contains only the bounded information required to paint and interact with it: opaque element id/band identity; component box geometry; ordered pre-broken lines; each line’s exact millipoint top/baseline/advance/width and paint text or shaped paint fragments; and an explicit overflow/clipping state where the existing engine’s line plan exceeds its declared width. It is not a generic template/style object, canonical `.folio` bytes, DOM/CSS instruction set, asset payload, or a TypeScript document schema.

**And** the projection follows production semantics:

- A text element uses the same declared font chain, shaping result, per-glyph advances, break opportunities, unbreakable-value behavior, width overflow rule, `FirstBaseline`, and `Advance` as production layout. It does not re-shape a line separately or use `hmtx`/per-rune metrics.
- Thai and CJK lines use the engine’s constrained opportunity rules. Unknown Thai runs stay atomic; clusters and declared unbreakable substitutions are not split merely because a browser could find a nicer-looking break.
- Geometry remains exact `geom.Length` millipoints, band-relative with top-left/Y-down coordinates (AD-2/AD-24). All values crossing to JS are safe integers within the existing canvas bounds; text/line counts and string/fragments are explicitly bounded before allocation/transport.
- The measurement path is read-only. Creating a new text projection does not mutate template bytes, revision, command history, dirty state, saved revision, selection, or Go engine ownership.

**Proof / red proof.** Go tests compare projection lines/widths/baselines against the production layout path for Latin kerning, Thai dictionary/unknown-run cases, CJK, mixed-face fallback, declared unbreakable substitution, empty text, overflow, and multiple lines. A mutation to use a browser-independent but raw advance path, re-shape each line, alter one break opportunity, or use font-size-as-baseline must make the parity test red. Tests must prove a measure/snapshot request preserves bytes and revision.

### AC2 — one worker protocol owns measurement and rejects malformed paint data

**Given** the existing singleton wasm worker and `EngineClient` FIFO channel,  
**When** a canvas needs a refreshed text plan after a successful load or committed command,  
**Then** the existing snapshot/projection route returns the Go-owned plan asynchronously; React neither calls wasm directly nor opens a second worker/wasm instance.

**And** the Go wasm bridge, TypeScript `EngineOperation`/inbound parser, worker admission, client deep-freeze, and tests are extended together as one closed protocol change. If a dedicated read-only measure operation is necessary, it accepts only a bounded opaque request and returns a snapshot-correlated paint result; it must not accept browser font metrics, CSS dimensions, viewport dimensions, device-pixel ratio, canvas contexts, a `.folio` fragment, or a claimed line break.

**And** TypeScript validates before paint: exact allowed keys, finite safe-integer millipoint fields, nonnegative/positive constraints, bounded ordered lines/fragments, unique opaque ids, known component/band order, line coordinates coherent with their component/band box, and per-type coherence. Unknown/surplus/malformed/unbounded/incoherent data is a protocol failure, never a best-effort browser repair.

**Proof / red proof.** Add Go/wasm/protocol tests for request closure, bounds, deep immutability, FIFO correlation, no stale response overwriting a newer snapshot/document lifecycle, and rejection of surplus fields, unsafe numbers, a duplicate id, a line outside the component, an invalid line order, an excessive string/list, and a fake browser-metrics field. Mutate one validator to accept surplus projection data and one worker request to accept a viewport/DPR/font input; each must red.

### AC3 — the browser rasterizes pre-broken text and cannot wrap or measure it

**Given** an engine-projected text component,  
**When** the canvas paints it,  
**Then** it renders the engine’s ordered lines exactly as supplied, with browser wrapping disabled: `white-space: pre`; no `text-wrap`, `overflow-wrap`, `word-break`, justification, flex/grid text sizing, ellipsis, line clamp, or content-derived box dimension may decide a line ending or component geometry.

**And** the Story 5.7 single-line `Text` placeholder is fully removed. No fallback code path may display text components by a single literal label, split on browser rules, or use the DOM to discover where a line ended. Non-text component structural painting remains exactly as it is until its owning stories add honest content.

**And** a repository-wide ownership guard covers production designer source, tests, and e2e source with narrowly documented false-positive exclusions. It rejects `CanvasRenderingContext2D.measureText`, `getBoundingClientRect`/`getClientRects`, `offset*`, `client*`, `scroll*` as canvas geometry sources, `ResizeObserver` used to feed document/paint geometry, computed-style/font inspection for metrics, `Range`/selection geometry, `document.fonts` measurement feedback, CSS wrapping/justification/clamp properties, and DOM-to-document coordinate persistence. UI chrome may use ordinary layout only where it cannot influence the canvas paint plan or a command.

**Proof / red proof.** React tests assert the literal projected line sequence and preformatted paint semantics; static tests scan the complete applicable source set, assert the guard’s nonempty population, and distinguish browser chrome layout from prohibited canvas authority. Mutate the painter to use `component.value`, add `white-space: normal`, add `measureText`, or add a DOM-rect-to-command conversion; each must red.

### AC4 — display mapping is deterministic, one-way, and DPI/font independent

**Given** a known immutable engine projection and a UI-local zoom,  
**When** it is painted at different viewport sizes, browser zooms, device-pixel ratios, or installed/fallback browser fonts,  
**Then** the mapping from projected millipoints to CSS pixels is one explicit, pure display transform shared by page, bands, components, text-line origins, selection outlines, handles, and pointer hit testing.

**And** the transform is based only on returned document coordinates and the explicit UI-local zoom. It has one documented display rounding rule; it never converts CSS pixels back into canonical geometry except the existing pointer delta/proposal path, which remains transient and is validated/snap/contained by Go. DPR may improve rasterization but is not an input to document geometry, line layout, hit identity, dirty state, save bytes, command payload, or engine measurement.

**And** browser fonts are cosmetic rasterizers only. The canvas may use a deterministic, project-owned CSS paint face/fallback declaration to make the pre-broken text legible, but it must not read font availability/metrics or adapt line positions/boxes to it. A missing browser face can look different; it cannot make a line reflow or move the following line. The exact visual artifact remains Story 5.10’s same-engine PDF preview.

**And** viewport changes are responsive only in the ordinary desktop-shell sense: scrolling/panning/available space may change, but the document page remains an absolute projected sheet. Remove the current mobile breakpoint behavior that collapses/hides designer rails if it claims or creates responsive canvas geometry; preserve a truthful desktop minimum/overflow experience instead. Do not add a mobile product mode or breakpoint-driven document layout.

**Proof / red proof.** Unit tests exercise the pure transform and hit-test inverse at fractional zooms, large safe values, page/band/component/text origins, and viewport/DPR/font-change simulations; they assert the same projected coordinates/line list and zero command/dirty/revision/serialization change. CSS/source checks reject a document-affecting media query, `window.devicePixelRatio` in geometry/command code, browser font readiness/measurement feedback, or separate transforms for paint and hit testing. A mutation that rounds text origins differently from component origins must red.

### AC5 — local hit testing, selection, accessibility, and dirty neutrality remain intact

**Given** a pointer or keyboard interaction over projected text,  
**When** the UI selects, drags, resizes, or focuses the component,  
**Then** hit testing derives from the same engine-projected component box and pure display mapping—not glyph bounds, DOM text bounds, `event.target` structure, or a layout query. Text paint may be nested/fragmented for rasterization without changing the opaque component target.

**And** selection, hover, focus, local zoom/pan/scroll, handle hit target, and any text-paint cache are UI state only. They send no command and do not alter bytes, revision, saved baseline, dirty indicator, or the next serialized file. A committed property/geometry/load replacement still replaces the authoritative snapshot, clears scoped transient state where prior stories require it, and invalidates/rebuilds only the derived local paint cache.

**And** the accessibility floor is preserved: component controls remain keyboard reachable with visible cyan focus and stable accessible names; a multi-line text component remains one control named from bounded projection metadata rather than one focus target per painted line; handles retain an enlarged named hit target. The known absence of a screen-reader semantic canvas model remains UX1—do not claim formal canvas/PDF accessibility or introduce a parallel DOM document as a workaround.

**Proof / red proof.** Component/a11y tests cover pointer text-line/blank-canvas selection, keyboard selection/delete continuity, focus/name stability across line changes, no line-level tab stops, enlarged handles, and local text paint/cache/viewport changes causing zero command/dirty effect. Lifecycle tests cover load/start-blank/property success/stale response without a prior plan repainting the new document. Mutate hit testing to read a DOM rect, make each line focusable, or set dirty on a measure/cache/resize event; each must red.

## Tasks / subtasks

- [x] **1. Extract one read-only production-parity text-paint seam in Go** (AC: 1)
  - Reuse the existing shaping, break, pack, declared-chain vertical-model, and position machinery. Factor only enough shared typed output to avoid a second layout algorithm; do not make the browser, PDF package, or a new Go renderer calculate its own text metrics.
  - Define bounded paint structs and an explicit per-snapshot derivation path. Preserve `Render` output and all canonical bytes/goldens; this projection is derived session output, never format state.
  - Cover production-parity cases and the no-mutation invariant before wiring React.

- [x] **2. Extend the singleton worker protocol as a closed paint boundary** (AC: 1, 2)
  - Update `folio-go/page_setup.go` projection types or a deliberately adjacent paint-only type, `folio-go/wasm/engine.go`, the wasm command bridge, `engine-protocol.ts`, worker admission/client tests, and deep-freeze behavior together.
  - Keep every request/result bounded and reject unknown fields on both sides. No TS `.folio` model, raw template, source text measurement request, font asset, browser environment input, or second wasm path.

- [x] **3. Replace placeholder text with projection-only multi-line DOM/SVG paint** (AC: 3, 4)
  - Evolve `folio-designer/src/App.tsx` and `App.css` from the current structural `Text` label to pre-broken engine lines. Use the one display transform and preformatted text; preserve current non-text rendering, selection, drag preview, resize handle, local file/dirty flow, token grammar, and dark chrome/light page.
  - Treat display font availability and viewport/DPR as raster/space conditions only. Do not add a responsive document mode; keep the desktop precision surface usable through bounded overflow/scroll.

- [x] **4. Make authority leakage mechanically difficult** (AC: 2–5)
  - Add a precise non-vacuous source guard for DOM/canvas text measurement, CSS wrapping/layout feedback, device/font/viewport feedback, and divergent paint/hit transforms. Keep exclusions named and justified; do not silence the guard with an empty candidate list or broad ignore.
  - Add pure mapping/hit-test tests and interaction/lifecycle/dirty-neutrality tests. Assert measurement/cache/resize events are not commands.

- [x] **5. Verify honestly under D-000.4** (AC: 1–5)
  - Run `npm run test`, `npm run typecheck`, `npm run lint`, `npm run build`, `npm run test:e2e:compile`; focused/full Go tests as appropriate, `go vet ./...`, native build, `GOOS=js GOARCH=wasm go build ./wasm/cmd/engine`, existing lint/hashmatrix static gates, gofmt, and `git diff --check`.
  - Execute and restore all named red proofs. Record counts, the sanctioned P6g corpus-floor red if it remains, and the unrun Playwright/four-target suites exactly; compilation is not browser execution.

### Review Findings

- [x] [Review][Patch] [High] Scoped the band-label rule to its direct child so it cannot cascade into text paint.
- [x] [Review][Patch] [High] Text uses horizontal-only clipping; vertical paint remains visible.
- [x] [Review][Patch] [High] Restored local pointer placement through the singular display mapper; it remains a transient Go-validated proposal.
- [x] [Review][Patch] [High] Closed request, response, lifecycle, and snapshot envelopes with operation/payload coherence.
- [x] [Review][Patch] [High] Reject incoherent text geometry: band/box escapes, non-advancing lines, invalid baselines, and fragment origins beyond supplied line width.
- [x] [Review][Patch] [High] Added an exact witness against the shipping shaped-run path: fragment text/origin, top, baseline, and per-glyph/per-face scaled line width must agree.
- [x] [Review][Patch] [High] Painter now consumes projected baseline and advance, with shared-mapper origin proof.
- [x] [Review][Patch] [High] Bound later-line, baseline, width, advance, and fragment derived coordinates before JSON transport.
- [x] [Review][Patch] [High] Authority guard scans production, unit tests, and e2e source; realistic measurement, range, style, justification, and coordinate mutations red-prove.
- [x] [Review][Patch] [Medium] Accessible names preserve fragment and line separation.

## Developer guardrails

1. **Projection is display data, not document authority.** AD-15/16 still mean one worker/wasm engine holds the template. Do not add TS `Element`, `Style`, `TextLine`, or `.folio` model types that can be serialized or mutated; a narrowly closed inbound paint DTO is not permission to mirror the format.
2. **Do not write a second layout engine.** Reuse the Go production shaping/break/packing/vertical calculations and arrange for both production and projection to consume the same result or an explicitly parity-tested shared seam. Do not duplicate line breaking in `App.tsx`, CSS, SVG, Canvas2D, or a Go helper with divergent arithmetic.
3. **No browser geometry feedback.** DOM/SVG/CSS/Canvas2D rasterize returned geometry only. Browser measurement APIs, `devicePixelRatio`, available width, loaded fonts, viewport, media queries, scroll dimensions, and text/glyph bounds never enter a measure request, command, saved coordinate, or line plan.
4. **Keep coordinate senses and rounding singular.** Document geometry is millipoints, top-left/Y-down, band-relative. One pure display mapper owns CSS-pixel conversion; one inverse/hit-test path reuses it. Do not use `Math.hypot`, CSS layout values, or per-widget rounding as a persisted coordinate source.
5. **Text scope is deliberately narrow.** Canvas text is the authoring template literal, not bound sample data or a production page. Table cells, binding values, final clipping/diagnostic overlays, and exact PDF visual fidelity remain owned by Stories 6, 5.12, and 5.10 respectively.
6. **Preserve prior interactions.** Selection is zero-command/zero-dirty; only accepted Go commands advance revision. Keep existing save/open/start-blank race protections, bound diagnostics, snap/containment in Go, table derived geometry, page setup, property drafts, local-only/offline behavior, and cyan/amber grammar.
7. **Do not turn accessibility into a hidden layout authority.** Accessible names/focus semantics may identify the component, but no hidden DOM text document, line-level control tree, or ARIA measurement substitute may define visual geometry. UX1 remains explicitly open.
8. **Preserve unrelated work.** Do not alter `.agents/`, `_bmad` configuration/manifest churn, research artifacts, fixtures/goldens/fonts, service worker release contents, deferred work, or later-story status. Do not commit, push, or execute Playwright/matrix suites in story creation.

## Files and implementation seams

**Update (read before changing):**

- `folio-go/wrap.go`, `folio-go/render.go`, `folio-go/internal/text/*`, `folio-go/internal/fontset/*` — source of shaped advances, break opportunities, packing, and vertical model. Extract/reuse rather than reimplement.
- `folio-go/page_setup.go` — current bounded `CanvasProjection`/`CanvasComponent`; extend only with the smallest paint-safe closed projection.
- `folio-go/wasm/engine.go`, `folio-go/wasm/cmd/engine/main.go` — snapshot lifecycle and one wasm transport boundary; read-only measurement must preserve state.
- `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`, `engine-worker-admission.ts` — protocol closure, FIFO and deep-freeze seams.
- `folio-designer/src/App.tsx`, `App.css` — current `displayPx`, page/band/component styling, single-line placeholder, drag handling, selection, dirty/save and lifecycle state. Evolve; do not replace the app state model.
- `folio-designer/src/{App,engine-protocol,engine-ownership-contract}.test.tsx` and `folio-designer/e2e/` — precedents for interaction, protocol and source-compiled browser scenarios. Add focused files only if clearer than extending these.

## Verification plan

Every-story gates are mandatory: Go production-parity/projection/wasm tests; designer Vitest component, mapping, protocol, ownership and accessibility tests; `npm run test`, typecheck, lint, production build, e2e TypeScript compilation; Go test/vet/native/js-wasm build; repository lint/hashmatrix static gates; format and diff hygiene.

Required restored red proofs: raw/unshaped advance or re-shaped line divergence; Thai/CJK break divergence; line baseline/width divergence; projection request changing revision/bytes; browser metric/viewport/DPR/font input accepted; surplus/unsafe/incoherent paint data accepted; `white-space: normal` or browser wrap accepted; `measureText`/DOM-rect text measurement admitted; distinct text/hit mapping; line-level focus targets; and dirty state changing after cache/viewport/measurement-only work.

Do **not** run Playwright/browser e2e or the four-target matrix at this story scope absent a newly demonstrated D-000.4 override. Compile the e2e source and record that the browser suite and matrix remain due at Epic 5 close.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5 and Story 5.9 acceptance criteria.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — NFR2/NFR3.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-2, AD-15–17, AD-24.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/{EXPERIENCE.md,DESIGN.md}` — UX-DR21, desktop/no-responsive boundary, accessibility floor and token rules.
- `_bmad-output/implementation-artifacts/5-6-the-canvas-page-bands-grid-and-page-setup.md` through `5-8-edit-component-properties.md` — current projection, coordinate, interaction, lifecycle, dirty/save and protocol contracts.
- `_bmad-output/implementation-artifacts/{folio-mvp-decision-log.md,sprint-status.yaml}` — D-000.4 and tracker state.

## Delivery Log

### Story creation — 2026-08-28

- Created from the complete Epic 5 requirements; shipped Stories 5.1–5.8 and commits through `be7a688`; PRD; architecture spine; UX EXPERIENCE/DESIGN; the Folio format/SPEC/acceptance material; D-000.4; current Go shaping/wrapping/render/geometry/wasm code; current worker/protocol/client/App source; tracker; decision log; and dirty-worktree inspection.
- Confirmed the existing canvas projection is Go-owned, safe-integer millipoints, ordered bands/components, and only a temporary single-line `Text` placeholder. Current production line packing already has the necessary shaping, `text.Opportunities`, `packLines`, overflow, and declared-chain vertical-model semantics, but no browser-facing measure projection exists. This story must factor/reuse that source, not duplicate it.
- Confirmed current UI local state includes zoom, selection, hover, drag, resize, snap, file/save baseline and property drafts. The current `displayPx`/pointer conversion is the appropriate narrow seam to consolidate into one display mapping; DOM/layout feedback must not become a document or measure authority.
- D-000.4 exact deferrals recorded: designer browser/Playwright execution and four-target matrix (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) are written/compiled, not run — Epic 5 boundary cadence, due after Story 5.12. No override is justified because the story must be byte-neutral and does not change canonical rendering.
- No implementation, test execution, commit, push, decision/deferred-work mutation, or change to pre-existing unrelated `_bmad`/`.agents` work was performed during creation.

### Development — 2026-08-28

- Added `CanvasWithTextPaint`, a read-only, bounded projection that reuses Go's shaped segments, break opportunities, line packer, positioned fragments, overflow decision, and declared-chain vertical model. The wasm singleton now returns it with each authoritative snapshot; canonical bytes and revision are unchanged by projection derivation.
- Extended the closed inbound TypeScript DTO/validator and deep-freeze path for ordered text paint lines and fragments. The validator rejects surplus fields, unsafe/incoherent geometry, oversized lists/text, duplicate component ids, and browser-metric-shaped payloads.
- Replaced the `Text` canvas placeholder with pre-broken, absolute projected fragments; all page/band/component/text display coordinates use `canvasDisplay`. The desktop surface now preserves rails and uses overflow rather than a breakpoint-driven mobile document mode. Text remains one keyboard control with its bounded projected content as its accessible name.
- Added Go projection byte-neutrality/overflow tests plus designer paint, protocol, and source-authority contract tests. Designer Vitest: 91 passing. Go: 1217 passing; the required P6g corpus-floor red remains (`7 < 20`) as recorded by D-000.4. Typecheck, lint (one pre-existing Fast Refresh advisory), e2e TypeScript compilation, vet, native build, js/wasm build, and diff hygiene passed. Browser Playwright and the four-target matrix were not run, per D-000.4.

### Code review — 2026-08-28

- Adversarial single-reviewer pass completed under the explicit no-subagent constraint: 10 patch findings (9 high, 1 medium), 0 decision-needed, 0 deferred, 0 dismissed.
- Designer tests passed 91/91. Go reported 1217 passing, 4 skipped, and only the sanctioned P6g corpus-floor failure (`7 < 20`; represented by the parent test and its P6g subtest).
- Typecheck, e2e source compilation, lint (one Fast Refresh advisory), production build, offline release verification, wasm-witness verification, offline red controls, Go vet/native/js-wasm builds, lint-module tests (117), gofmt, and diff hygiene passed. The offline stale-wasm red proof failed as intended.
- Playwright/browser execution and the four-target matrix remained unrun under D-000.4. No implementation fixes, commit, push, or next-story work was performed.

### Finisher — 2026-08-28

- Resolved all ten persisted findings; none were waived. The display path now uses direct-child band labels, width-only text clipping, baseline/advance paint coordinates, and a single local pointer proposal mapper. The mapper receives the pointer's local event coordinate only and neither measures DOM layout nor owns document geometry.
- The TypeScript boundary now rejects surplus request/response/lifecycle/snapshot fields, incorrect payload-operation combinations, browser-environment inputs, non-advancing lines, invalid baselines, and fragment origins incoherent with their engine-supplied line width. The source guard covers applicable production, unit, and e2e source and has red controls for Canvas2D measurement, offsets outside the isolated placement seam, computed style, DOM ranges, selection geometry, wrapping, and justification.
- The Go projection now proves its output against the shipping shaped-run collection: exact fragments, origins, baseline, and the same per-glyph/per-face rounding sequence are compared. Derived origins, baseline, advance, width, and fragments are range-checked before crossing JSON; explicit overflow red proof covers later-line/baseline arithmetic.
- Re-measured: designer Vitest **94/94**; typecheck and e2e TypeScript compilation passed; lint passed with only the existing Fast Refresh advisory; production build, offline red controls, and emitted-wasm witness passed. Go green path passed **1213** with the one sanctioned P6g test skipped; full Go reports only the expected P6g corpus-floor red (`7 < 20`); vet, native/js-wasm, matrix-tag compile/vet, lint-module (**117**), hashmatrix (**3**), gofmt, and diff hygiene passed.
- Per D-000.4, Playwright/browser execution and the four-target hash matrix remain written/compiled but unrun until Epic 5 close; no new override was demonstrated.

## Status

done
