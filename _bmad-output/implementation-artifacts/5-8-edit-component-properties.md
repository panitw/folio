---
baseline_commit: 66e6bc6
story_key: 5-8-edit-component-properties
status: done
created: 2026-08-28
---

# Story 5.8: Edit component properties

**Epic:** 5 — A template author can lay out a report and see the real PDF  
**Covers:** FR5 · UX-DR10, UX-DR25 · AD-2, AD-9, AD-10, AD-13, AD-14, AD-15, AD-16, AD-17, AD-24

**Standing delivery decision:** D-000.4 requires focused unit/component/contract tests, lint, typecheck, Go tests/vet/build, and production build for this story. Designer browser/e2e execution and the four-target hash matrix (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) are **written/compiled, not run — D-000.4 Epic 5 boundary cadence**, due after Story 5.12. Property editing changes authoring commands and the approximate canvas, not canonical rendering, numeric emission, text measurement, or pagination; it is not a D-000.4 override. Compilation is not browser execution.

## In plain terms

This story turns the right rail from page setup into a truthful component-properties editor. Selecting one or more components exposes only values Folio can actually persist and render. An author may type an exact draft without modifying the document; only Blur or Enter sends one opaque command to Go. Go validates a fresh candidate, canonicalizes it, and returns a new paint-safe projection. A rejected edit leaves bytes, revision, dirty/save baseline, selection, and committed display unchanged while the literal draft remains visible with a located diagnostic.

The browser owns local selection, focus, field drafts, mixed-value display, and transient pending state. It does not own a `.folio` type, defaults, validation, expression parsing, exact geometry, style schema, or a second copy of committed properties. The Go engine remains the document and layout authority. The canvas must therefore continue to show structural labels/geometry only; Story 5.9 alone owns engine measurement and pre-broken text painting.

Binding discovery and table-column configuration are deliberately not built here: Story 6 owns the sample-data tree, bind picking, row scope, and table matrix. This story may expose only existing, directly editable persisted fields, and it must say when a selected type has no supported property instead of drawing a fake field.

## Story

**As a** template author,  
**I want** to set a component's exact position, size and typography,  
**So that** I can be precise where dragging is only approximate.

## Source-grounded acceptance criteria

### AC1 — selection-contextual, five-type property panel

**Given** no component is selected,  
**When** the properties rail renders,  
**Then** it retains the existing page-setup controls and explicitly says that component properties require a selection; it does not show invented component values.

**Given** one selected component,  
**When** its panel renders,  
**Then** it identifies the opaque element id and type and exposes the applicable persisted properties:

| Field group | Text | Image | Table | Line | Rectangle | Truth/constraint |
|---|---|---|---|---|---|---|
| Position `x`, `y` | yes | yes | yes | yes | yes | exact points, band-relative, Y-down; Go validates containment |
| Size `width`, `height` | yes | yes | **no** | yes | yes | table geometry is derived and never stored/resized (AD-13) |
| Text literal `value` | yes | no | no | no | no | Text may contain `{{ }}`; no browser expression schema/validation |
| Font family, font size, bold, italic, align, valign | yes | no | table-level style only | no | no | use only the format's existing `style` fields; unsupported type/value is absent, not disabled fiction |
| Border (width, colour, edges), padding, background | yes | yes | table-level style only | yes | yes | same existing `style` vocabulary; table header style/columns are not edited here |
| Visibility `visibleIf` | yes | yes | yes | yes | yes | bare expression, no `{{ }}` wrapper; Go/ParseTemplate validates it |
| Binding-related value | text value only | no | table bind is display-only/deferred | no | no | no sample-data path picker, table binding mutation, row scope, or column editor before Epic 6 |

**And** the panel never offers image asset import, a table width/height, table columns/header style, conditional formatting, arbitrary CSS, unimplemented font/colour choices, or a sixth component capability. It must render an explicit, terse “Not editable here” reason for a selected field/type that Folio cannot yet support rather than silently dropping a user edit.

**Proof / red proof.** Component tests assert the complete per-type field matrix, including table's derived-geometry absence and deferred binding/table-editor truth. Structural tests reject a TS `.folio`/element/style schema, client-side `nextId`, document JSON serialization, an arbitrary-style/CSS input, or fields shown for unsupported types.

### AC2 — Go owns typed property commands, validation, serialization, and projection

**Given** a property draft is committed,  
**When** it crosses the worker boundary,  
**Then** TypeScript sends exactly one opaque, versioned `updateComponentProperties` intent through the existing `EngineClient` FIFO channel; it does not read/mutate canonical bytes or construct a document/element/style object.

**And** Go decodes a deliberately closed command vocabulary, validates its exact JSON fields with `UseNumber`, locates the opaque id, applies all requested changes to a fresh canonical candidate, reparses/canonicalizes, produces the new projection, and increments revision only after every step succeeds.

**And** command values are typed and Go-owned:

- Geometry and font size are canonical point literals parsed to `internal/geom.Length` millipoints; no `float64`, CSS pixels/DPI, browser rounding, tolerance, clamp, or silent repair decides persisted values. Respect the `MaxCanvasMillipoints` JS-safe projection bound and containment in the existing band.
- String and enum values use the format's closed validation: `align`, `valign`, border edges, colour spelling, font-family semantics, and field-presence/null semantics. Do not invent a TypeScript schema as a precondition for Go validation.
- `visibleIf` is a bare expression. Text `value` is the only literal content field here that may contain `{{ }}`. Go must validate both by the same `ParseTemplate`/expression route used on load; style and appearance strings containing `{{ }}` remain rejected, not treated as data-driven formatting.
- The command must distinguish **set**, **clear/absent**, and the format's intentional explicit `null` behavior wherever that field accepts it. Do not convert absence to zero, `false`, empty string, or an arbitrary default. A clear command must serialize canonically as an absent optional key where the format says absent means default/inherit.
- A request touching several fields is atomic: one accepted command produces one snapshot/revision; any invalid field rejects the whole request with no partial mutation. Preserve unknown extension fields and all unrelated component/document state.

**Projection extension rule.** The current `CanvasComponent` exposes only paint geometry/id/type/band/resizable. Extend the Go-owned canvas/wasm snapshot projection only with the minimum paint-safe, bounded, explicitly named committed values necessary to render the property panel and canvas honestly. Do not expose a generic style bag, raw `.folio`, arbitrary `Extra`, asset bytes, or a TypeScript mirror of `template.Element`. TypeScript must validate the closed projection shape, bounds, per-type coherence, unique ids/order/band containment, and optional-field bounds before painting it.

**Proof / red proof.** Go tests cover every legal field/type, clear vs explicit null/absence semantics, malformed/unknown/missing fields, unknown id, bad enum/colour/expression/length, table size rejection, containment, JS-safe edge/overflow, atomic multi-field rejection, canonical round trip, unchanged bytes/revision on every failure, and unknown-field preservation. Red controls must make a client-side acceptance, float decode, a partial mutation, an invalid style `{{ }}`, or a table size update fail.

### AC3 — literal local drafts and commit semantics

**Given** an editable property field is focused,  
**When** the author types, deletes to empty, or enters a temporarily invalid literal,  
**Then** the exact draft string remains UI-local and visible; no command, snapshot/revision change, dirty-state change, save serialization, or canvas committed-value change occurs before Enter or Blur.

**Given** Enter or Blur commits a non-pending draft,  
**When** Go accepts the command,  
**Then** the returned immutable snapshot replaces the committed value, the local draft reconciles to Go's canonical display spelling, and the existing dirty/save flow becomes dirty only from that returned revision.

**Given** Go rejects the commit, the worker fails, the selection changes, the document is opened/replaced, or an earlier async response arrives after a newer field/selection/document generation,  
**When** the response is handled,  
**Then** it cannot overwrite a newer committed snapshot, newer draft, selection, saved revision, or document lifecycle state. The failed draft remains literal for correction unless its owning document/selection was intentionally replaced; in that replacement case it is discarded atomically with Story 5.7's selection/placement/drag reset.

**And** a pending field cannot dispatch duplicate commits; blur after Enter for the same draft is idempotent. A selection-only click remains zero-command and zero-dirty. Save remains explicit/local and continues to serialize only engine-owned canonical bytes; it must neither commit an unblurred draft nor mark an in-flight/newer revision clean.

**Proof / red proof.** Tests prove one keystroke and invalid draft emit zero traffic; Enter/Blur emit exactly one command; successful response alone changes dirty; rejected input preserves bytes/revision/committed display and shows literal draft; an out-of-order/stale response cannot overwrite a newer draft/snapshot; open/start-blank clears scoped drafts; and save before/after property success preserves the existing saved-revision race rule. Mutate to commit on change, clear empty drafts, accept a stale response, or mark dirty before Go success; each test must red.

### AC4 — multi-selection is explicit, local, and safe

**Given** multiple components are selected through Story 5.7's UI-local selection semantics,  
**When** the panel renders,  
**Then** it exposes only the intersection of properties actually applicable to every selected type, and shows a stable mixed state for divergent committed values rather than choosing one arbitrarily.

**Given** a shared property is committed from a mixed selection,  
**When** Go accepts it,  
**Then** one atomic Go command applies that same field change to the explicit selected opaque ids, returning one snapshot/revision. No selection operation itself enters the document.

**Given** a multi-selection includes an incompatible type or table,  
**When** a non-shared field is considered,  
**Then** it is omitted with a truthful reason; `width`/`height` are not mass-applied across a table, text literal value is never mass-edited, and table structure/binding remains deferred. Multi-selection order is UI-local and must not become a document ordering rule.

**Given** any target id is stale, absent, invalid for the chosen field, or makes the candidate invalid,  
**When** Go processes the batch,  
**Then** the whole batch rejects transactionally with a bounded, located diagnostic and leaves all canonical bytes/revision/dirty/save baseline untouched.

**Proof / red proof.** Cover same-type and mixed-type selections, all-equal vs mixed display, an empty/no-common intersection, one atomic shared update, no command for selection/mixed display, and one bad target causing a complete rollback. Red controls: first-selected-value fallback, per-element browser command loop, partial batch application, or a table geometry field passing the intersection.

### AC5 — diagnostics, accessibility, and visual grammar

**Given** a property field or control,  
**When** navigated with a keyboard,  
**Then** it is reachable and operable in a logical panel order; Enter commits, Escape cancels/reverts the local draft to its last committed projection value, and visible focus uses the existing cyan `select` token. Native controls retain accessible labels; any icon-only control has an accessible name.

**Given** Go rejects an edit,  
**When** the error reaches the main thread,  
**Then** the existing bounded component-diagnostic protocol carries a stable code, message, optional opaque element id, and data path; the UI announces it once in an accessible live region and locates it to the field/selection without inventing a browser validation result. Bound `elementId`, `dataPath`, code, and message lengths at both Go/wasm and TypeScript protocol boundaries.

**And** the panel follows `DESIGN.md`: dense square property fields, dark chrome, token-backed text/rules, cyan for focus/selection/authority only, amber for actual data identity only, no gradients or raw style literals. It makes no server, account, cloud, autosave, exact-preview, or browser-layout-authority claim.

**Proof / red proof.** Component/a11y tests assert label/control association, tab order, Enter/Escape, select-token focus, mixed-state announcement, and bounded error live-region behavior. Design/ownership guards reject a raw colour/radius/gradient, amber selection/focus, missing accessible name, an unbounded protocol optional field, or a browser-only “valid” commit path.

## Tasks / subtasks

- [x] **1. Define the narrow Go-owned property command and projection seam** (AC: 1, 2, 4)
  - [x] Extend `folio-go/component_commands.go`; do not add a parallel command service or a TS document model. Make `wasm.Engine.Apply` explicitly route the new command kind while retaining page setup and Story 5.7 operations.
  - [x] Add typed, closed Go mutation helpers over `internal/template.Element`/`Style`/`Presence`; preserve extension fields and canonical parser/serializer ownership.
  - [x] Extend `folio.CanvasProjection`/wasm `Snapshot` minimally for panel-needed paint-safe values; update Go JSON transport and `folio-designer/src/engine-protocol.ts` closed inbound guard together.
  - [x] Carry `ComponentCommandError` field/id detail through `folio-go/wasm/cmd/engine/main.go` and worker protocol; retain bounded diagnostics on errors.
- [x] **2. Implement transactional exact validation and expression/style fences** (AC: 2, 4)
  - [x] Reuse `lengthField`, `SnapToGrid` only where a documented property command calls for it, `containComponent`, `ParseTemplate`, `SerializeTemplate`, and `validateAndDeriveExpressions`; do not duplicate their rules in JS.
  - [x] Define explicit allowed property/type mapping and set/clear/null behavior in Go, including table derived geometry and non-support for asset/table-column/header-style editing.
  - [x] Require one candidate/serialization/projection/reparse transaction for single and multi-target updates; failures must not change engine state.
- [x] **3. Build the contextual React property panel from the immutable projection** (AC: 1, 3, 4, 5)
  - [x] Replace the current page-setup-only right rail with selection-contextual sections while preserving working page setup and its draft-generation stale-response behavior.
  - [x] Add a small UI-local draft controller keyed by document lifecycle generation + selected-id set + field. It may hold literal strings/mixed/pending/error state, never a committed element/style object.
  - [x] Encode only opaque property command intent in a dedicated command helper; use the sole `EngineClient` FIFO and update committed UI state only from a validated Go response.
  - [x] Implement exact single/multi-selection intersection and mixed display from the projection; do not infer absent values or silently use a first selection.
- [x] **4. Preserve dirty/save/lifecycle and accessibility contracts** (AC: 3, 5)
  - [x] Reuse `snapshotRef`, saved-revision comparison, and document replacement clearing in `App.tsx`; make property response generation-aware without regressing open, Start blank, Save/Save As, component manipulation, page setup, grid, or snap.
  - [x] Implement keyboard labels, logical focus, Enter/Blur/Escape behavior, live diagnostics, and token-only property styling. State unsupported/deferred fields honestly.
- [x] **5. Prove the contract and record truthful evidence** (AC: 1–5)
  - [x] Add Go command/engine/wasm tests, protocol rejects, command-helper tests, React interaction/a11y tests, ownership/style static tests, and an e2e source/compile scenario for one single and one mixed property edit.
  - [x] Run and restore named red-proof mutations; record exact commands/counts and distinguish executed gates from D-000.4 deferred browser/matrix gates. Do not claim e2e execution from compilation.

### Review Findings

- [x] [Review][Patch][High] Preserve newer property drafts across stale or unrelated successes — stable document/selection panel identity no longer remounts on revision; scoped response reconciliation is covered in `App.test.tsx`.
- [x] [Review][Patch][High] Carry and locate component diagnostics end to end — `elementId` now survives `EngineClient`, with a field-associated alert and accessible error relation.
- [x] [Review][Patch][High] Complete the five-type field and operation surface — border edges, display-only table bind, and explicit clear/null controls are present.
- [x] [Review][Patch][High] Validate colour and font-family semantics before accepting a command — invalid colours and undeclared/empty font chains reject before install.
- [x] [Review][Patch][High] Reject unsupported null operations instead of serializing empty strings — only documented `visibleIf` and `background` nulls remain legal.
- [x] [Review][Patch][High] Align Go projection bounds/coherence with the TypeScript boundary — bounded strings/lengths and type-filtered typography now agree at both ends.
- [x] [Review][Patch][Medium] Reject surplus fields in the inbound projection — canvas, bands, components, and diagnostics are closed shapes.
- [x] [Review][Patch][Medium] Make duplicate commits impossible for every control — synchronous in-flight gates cover draft, boolean, and border controls.
- [x] [Review][Patch][High] Add interaction, protocol, rollback, and mutation evidence — React/a11y, client diagnostics, protocol closure, engine rollback, command, Go property, and compiled e2e tests now cover the seam.

## Developer guardrails

- **Do not move authority to TypeScript.** AD-15/16: one wasm module/instance/dedicated Worker owns the template; TS stores immutable snapshots and sends opaque commands. Do not expose canonical bytes except the existing serialize/save route, create a `Component`/`Style` TS schema, parse `.folio` in the browser, or add a second worker/wasm instance.
- **Do not make the browser a layout authority.** DOM/CSS can only arrange controls and rasterize projection data. It must not determine line breaking, font measurement, pagination, persisted geometry, containment, or defaults. Story 5.9 owns engine measurement/pre-broken text; do not pre-empt it with `measureText`, DOM layout, `getBoundingClientRect` persistence, CSS wrapping, or browser fonts.
- **Respect format semantics.** Geometric values are exact points in command literals and `geom.Length` millipoints in Go. Coordinates are band-relative top-left/Y-down. Tables have `x`/`y` only; width derives from columns and height derives from rows. Optional style fields use `Presence`, not Go/JS zero values. The only component types remain text/image/table/line/rect.
- **Expressions are not style.** Text `value` may contain `{{ }}`. `visibleIf` is bare and must validate through the existing expression parser/checker. `style` string fields and `table.altRowBackground` must reject placeholder syntax: conditional formatting is out of scope. No JSON Schema or expression library.
- **Keep future ownership clean.** Story 5.9: measure/paint text; 5.10–5.11: exact preview/identity/staleness; 5.12: broad keyboard shortcuts/undo/redo/diagnostic UI; Epic 6: sample data, bind tree, path picking, table columns/matrix/aggregates. Do not implement their models or fake affordances.
- **Carry 5.7 repairs forward.** Selection is local and a no-op for revision/dirty; a real drag is the only move transaction; Go owns half-open band hit testing and exact points; command failures retain bounded id/path diagnostics; inbound projections must reject duplicate/noncanonical/out-of-band/incoherent values; successful open/start blank clears document-scoped interaction state; resize hit target exceeds visual mark.
- **Preserve unrelated work.** Do not modify current pre-existing `_bmad` configuration/manifest churn or `.agents/`. Do not regenerate engine artifacts by hand. No commit, push, or browser execution belongs to this story-creation task.

## Files and implementation seams

**Update (read before changing):**

- `folio-go/component_commands.go` — closed component command decoder/factories, exact lengths, transactions and bounded errors; extend, do not bypass.
- `folio-go/page_setup.go` — paint-safe `CanvasProjection` and exact geometry; only add narrowly required committed values with type/coherence checks.
- `folio-go/wasm/engine.go` — clone → apply → canonical serialize → reparse → projection → install transaction; preserve no-change-on-failure/revision semantics.
- `folio-go/wasm/cmd/engine/main.go` — wasm response diagnostic bridge.
- `folio-go/internal/template/model.go`, parser/serializer and `folio_expr_validate.go` — source of optional style/visibility/expression meaning; keep `internal/` boundaries and parser ownership intact.
- `folio-designer/src/App.tsx` — existing selection, page setup draft generation, file lifecycle, dirty/save baseline, and one `EngineClient`; evolve rather than replace.
- `folio-designer/src/component-command.ts` — opaque command encoding only; use a separate narrow property-intent helper, not model types.
- `folio-designer/src/engine-protocol.ts` and `engine-client.ts` — closed inbound validation/deep freeze/FIFO handling; extend bounds and projection shape synchronously with Go.
- `folio-designer/src/transient-interaction.ts` — existing explicit seam for local drafts; replace/extend only if it preserves no-document-access discipline.
- `folio-designer/src/App.css`, `tokens.css`, `design-contract.test.ts`, `engine-ownership-contract.test.ts` — token/a11y/ownership constraints, not a second style language.

**Likely new focused tests:** `folio-go/component_properties*_test.go`, `folio-go/wasm/*properties*_test.go`, `folio-designer/src/component-property-command.test.ts`, `folio-designer/src/properties-panel.test.tsx`, and a property-edit e2e spec. Follow existing colocated naming/tooling rather than adding a new test framework.

## Verification plan

Run the repository's current designer scripts from `folio-designer/`: `npm run test`, `npm run typecheck`, `npm run lint`, `npm run build`, and `npm run test:e2e:compile`; run appropriate focused and full Go tests, `go vet ./...`, native and `GOOS=js GOARCH=wasm` builds, format check, and `git diff --check`. Use the existing lint/hashmatrix gates as prescribed by the prior story’s delivery record.

Required red proofs (each restored): commit-on-keystroke; stale property response overwriting a newer draft/snapshot; mixed selection using first value; a table width update accepted; invalid `visibleIf`/style placeholder accepted; malformed/unknown property command fields; one bad multi-target producing a partial update; inbound duplicate/unbounded/incoherent property projection accepted; raw/amber focus or inaccessible property field. Record any sanctioned P6g corpus red exactly as it occurs; never call it green.

Do not run Playwright/browser e2e or the four-target matrix at story scope absent a genuine D-000.4 override; compile their source and record that they remain due at the Epic 5 boundary.

## References

- `_bmad-output/planning-artifacts/epics.md` — Epic 5, Stories 5.7–5.12 and Story 5.8 ACs.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR5, FR20, FR41, NFR2/NFR5/NFR8.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AD-2, AD-9, AD-10, AD-13–17, AD-24; D-000.4.
- `_bmad-output/specs/spec-folio/folio-format.md` — element/style/visibility/table/expression semantics and canonical format.
- `_bmad-output/specs/spec-folio/SPEC.md` and `acceptance.md` — preservation and acceptance contracts.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` — property-field visual/token contract.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — property commit/mixed state, interaction and accessibility floor.
- `_bmad-output/implementation-artifacts/5-7-place-and-manipulate-the-five-components.md` — current command/projection, lifecycle, diagnostic, units, and D-000.4 lessons.

## Dev Agent Record

### Agent Model Used

GPT-5.6

### Debug Log References

`npm run test` (81/81); `go test . ./wasm` (583/583); full `go test ./...` (1,213 passed, 4 skipped, only sanctioned P6g floor red); `go vet ./...`; native and js/wasm build; lint and hashmatrix static suites; designer typecheck/lint/build/e2e TypeScript compilation; format and diff checks.

### Completion Notes List

Implemented Go-owned `updateComponentProperties` with a closed set/clear/null vocabulary, clone/serialize/reparse transaction, typed projection fields, transactional multi-target rejection, table derived-geometry fence, and expression/style revalidation. The React right rail now retains page setup without selection and otherwise uses local literal drafts, mixed values, Enter/Blur/Escape semantics, stale-scope protection, keyboard labels, and explicit deferred-property truth.

Executed gates: Vitest 81/81; focused Go/wasm 583/583; lint static 117/117; hashmatrix static 3/3; typecheck, oxlint, production build, e2e source TypeScript compilation, Go vet/native/js-wasm builds, gofmt and diff check all passed. Full Go tests remain red only for the pre-existing sanctioned P6g opaque-name corpus floor (7/20; parent plus subtest), never reported as green.

Restored red proofs: raw `border-radius: 0` made the design-contract test red; temporarily allowing table width made the table-geometry rollback test red. Browser Playwright execution and the four-target hash matrix were not run under D-000.4 and remain due at the Epic 5 boundary; compilation is not browser execution.

### File List

- `folio-go/component_commands.go`
- `folio-go/component_properties_test.go`
- `folio-go/page_setup.go`
- `folio-go/wasm/engine_test.go`
- `folio-designer/src/App.tsx`
- `folio-designer/src/App.css`
- `folio-designer/src/component-property-command.ts`
- `folio-designer/src/component-property-command.test.ts`
- `folio-designer/src/engine-protocol.ts`
- `folio-designer/e2e/component-properties.spec.ts`

## Change Log

- 2026-08-28: Implemented Story 5.8 component properties and recorded D-000.4-constrained verification evidence.

## Finisher verification (2026-08-28)

- Designer: `npm run typecheck`, `npm run test` (**86 passed**), `npm run lint`, `npm run build`, and `npm run test:e2e:compile` passed. Browser Playwright execution remains deferred to the Epic 5 D-000.4 boundary; the e2e scenario was compiled, not executed.
- Go: focused `go test . ./wasm` (**585 passed**), `go vet ./...`, native build, js/wasm build, gofmt and `git diff --check` passed. Full `go test ./...` reported **1,215 passed, 4 skipped**, plus only the sanctioned P6g corpus red (parent and `P6g_(opaque_names)`, **7/20**).
- Static gates: root `lint` (**117 passed**) and `hashmatrix` static suite (**3 passed**) passed. The four-target matrix was not run, per D-000.4.
- Restored mutation checks: accepting table width, accepting `not-a-colour`, and accepting a surplus projection field each made its focused test red; all three source mutations were restored before the final gate run.

## Status

done
