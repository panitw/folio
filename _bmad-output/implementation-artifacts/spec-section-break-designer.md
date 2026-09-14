---
title: 'Section Break — designer'
type: 'feature'
created: '2026-09-14'
status: 'done'
route: 'dispatch'
baseline_commit: '3a645cfc0ab75f3254cd75e748772ffb89752b3a'
review_loop_iteration: 1
context:
  - '{project-root}/_bmad-output/specs/spec-section-break/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-section-break/brownfield.md'
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The engine supports a content-band `sectionBreak`, but an author can only set one by hand-editing the `.folio` file. The canvas neither shows the break nor stops an edit that would put an element across it.

**Approach:** In the designer, the Section Break behaves like an element: it is placed from the palette, selected, dragged, nudged, edited in Properties and deleted. In the file it stays the content band's `sectionBreak` key. The engine commands refuse any edit that would leave an element crossing the line. This story covers CAP-1 and CAP-6 of `spec-section-break`.

## Boundaries & Constraints

**Always:**
- **Placing (owner decision):**
  - A **Section Break** palette entry arms placement. A click inside the content band places the break at that height, snapped by the engine.
  - The entry is disabled while a break exists. This is a deliberate exception to the palette's closed-set and disabled-only-during-load rules.
  - No break exists by default.
  - Keyboard (owner decision, review pass 1): while armed, Enter or Space on the content band places the break at the middle of that sheet's content window (the sheet's origin plus half the content band's height), snapped by the engine. If that height is through an element, the engine refuses and names it.
- **After placing (owner decision):**
  - Clicking the line selects it with its own selection state, never a component id. It never joins bulk edits, group moves, copy, duplicate or select all.
  - Delete or Backspace removes it.
  - Properties shows an editable **Y** field (a typed value is not snapped).
  - Drag uses `clientY` travel with a proposal shown during the drag; arrows move 1pt, Shift 10pt, Escape aborts.
- **One undo entry:** place, drag, nudge, a Y edit and delete are each one engine command.
- **Look (owner decision):** a solid 1px cyan line with the band-boundary overhang and a mono **SECTION BREAK** tab. `DESIGN.md`'s fixed-meaning line table gains this row. No marking on below-line elements.
- **Drawing:**
  - The line is drawn once, on the sheet whose window holds its offset (`homeWindow`), with no echoes, and never in the PDF.
  - The below-line section is drawn only at its declared position; the canvas window count keeps plain `Paginate`.
- **Projection:** it carries the break offset (absent when unset) and each content component's section membership. Go and TypeScript wire pins change in the same commit.
- **Engine refusals** (located `COMPONENT_INVALID` naming the blocking element; document unchanged):
  - A move, resize, bounds edit, `x`/`y`/`height` property edit, create, drop, duplicate or table header-height edit that would leave an element across the break.
  - A group move with any member across; the whole group is refused.
  - A break placed, dragged or typed through an element.
  - A band-height or page-setup change that leaves the break at or below the content height; this refusal names the break.
- **Other rules:**
  - No DOM measurement and no new measurement exception.
  - The label **Section Break** is the owner's name; the EXPERIENCE rule against "section" is about softening *band* and does not apply.

**Never:**
- Engine pagination or format changes (story 1).
- A pushed copy of the section on the canvas.
- More than one break.
- A printed line.
- Changes to `PaletteKind` or the component create, drop, delete or duplicate commands' semantics.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Place | No break; palette Section Break, click content band at 400pt | Break at 400 (snapped), line drawn and palette entry disabled; one undo entry | Click through an element: refused alert, nothing placed |
| Keyboard place | Armed; Enter on the content band of sheet N | Break at that sheet's origin plus half the content height (snapped); one undo entry | Through an element: refused alert naming it |
| Wrong band | Armed; click page header or footer | Nothing placed | N/A |
| Select and delete | Click line, press Delete | Key removed, palette entry re-enabled; undo restores it | N/A |
| Y field | Selected; type 520 | Break at 520 exactly | Through an element or ≥ content height: alert, value reverts |
| Drag and nudge | Drag 40pt down; ArrowDown | One command each; Escape during drag sends nothing | Released onto an element: refused, line returns |
| Element move across | Text box `y` 700, `h` 40 moved over break 720 | Document unchanged | Alert naming the element |
| Group move | One member would cross | No member moves | Alert naming that member |
| Band height | Footer grows so content height ≤ break | Height unchanged | Alert naming the break |
| Multi-sheet | Offset in window 2 | Line drawn on sheet 2 only | N/A |
| Preview | Section-break statement fixture | Preview PDF bytes equal the native render | N/A |

</frozen-after-approval>

## Code Map

- **`folio-go/component_commands.go`:**
  - `:303` dispatch; `:2761-2924` the `setBandHeight` precedent (snap, echo, restore on refusal); `:33` `componentFailure`.
  - Straddle checks go at `:2200` (move), `:2237` (resize), `:2333` (bounds), `:1310` (property x/y/height), `:2097` (create/drop), `:2439`/`:2478` (duplicate/clone) and `~:3212` (header height).
  - Add `setSectionBreak`/`removeSectionBreak` and a shared `refuseSectionBreakStraddle`, reusing `sectionBreakDeclaredBox`/`declaredSectionBreak` (`section_break.go`).
- `folio-go/group_movement.go:194-225` -- a per-member check before install.
- `folio-go/page_setup.go:2316-2425` (page-setup refusal after `refuseStrandedFloor`, then `restorePage`) and `:250`/`:426`/`:939-952` (projection). `wasm/engine.go:298-351` already reparses and records one undo entry per changed document.
- `folio-go/canvas_projection_wire_test.go:47,676,732` holds the wire pins. `section_break_test.go:547`: update the projection-unchanged test on purpose; the pagination-derived fields stay equal.
- `folio-designer/src/engine-protocol.ts:415,477-478,744,~802,~816` -- types, `hasOnly`, and an explicit check for the optional offset. Tests in `engine-protocol.test.ts:10-72`.
- `band-boundary.ts`, `band-height-command.ts`, `command-json.ts:115` -- patterns for a pure `section-break.ts` and `section-break-command.ts`.
- **`folio-designer/src/App.tsx`:**
  - Palette: `:122`/`:230`/`:3107` (separate entry, `disabled` when `canvas.sectionBreak` is set). Widen `placing` to `PaletteKind | 'sectionBreak'` (`:354`), branch early in `placeInBand` (`:1496`), content band only.
  - Selection: `sectionBreakSelected` state, cleared by `installSelection` (`:1180`); selecting it clears `selected`.
  - Handle: model it on `button.band-boundary-handle` (`:2953`, drag and nudge at `:1668-1772`, `stopPropagation`). Draw it on `homeWindow(origins, offset)`, outside `.band-window`. The Delete guard is at `:1611-1627`/`:2774`.
  - Properties: a `SectionBreakProperties` branch at `:3180`, copying the `PropertyDraft` (`:3791`) commit-on-blur/Enter pattern. Focus after placing mirrors `selectPlaced`/`pendingFocus` (`:1552-1607`).
- `folio-designer/src/App.css:230,258,282-300` -- mirror the seam, rule, handle, proposal and readout styles.
- **Tests to update or model on:**
  - `App.test.tsx:1947,8996` (exact palette lists), `:1984,2023,2491,6982,7152`
  - `canvas-authority-contract.test.ts`, `control-vocabulary-contract.test.tsx`
  - `folio-go/band_height_command_test.go`, `group_movement_test.go:109`
  - `e2e/band-boundary-drag.spec.ts`, `placed-component-selection.spec.ts`
  - `folio-go/wasm/engine_test.go:317` (add `section-break-statement`)

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/component_commands.go`, `group_movement.go`, `page_setup.go`, `section_break.go` and tests -- set/remove commands, straddle refusals on every geometry command, band-height and page-setup refusals.
- [x] `folio-go/page_setup.go`, `canvas_projection_wire_test.go`, `section_break_test.go` -- projection offset and membership, plus wire pins.
- [x] `folio-go/wasm/engine_test.go` -- preview/native byte parity for `section-break-statement`.
- [x] `folio-designer/src/{engine-protocol,section-break,section-break-command}.ts` and tests -- protocol, pure maths, command bytes.
- [x] `folio-designer/src/App.tsx`, `App.css`, `App.test.tsx` -- palette entry, placement, selection, handle drag and keyboard, Delete, Properties Y field, alerts, home-sheet-only drawing.
- [x] `folio-designer/e2e/section-break.spec.ts` -- place, select, drag, undo, delete and a refused drag through the real worker.
- [x] `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` -- a fixed-meaning table row for the solid cyan Section Break line.

**Acceptance Criteria:**
- Given a document without a break, when it is projected, then it has no `sectionBreak` and every existing designer and engine test passes, apart from the deliberately updated palette-list and projection pins.
- Given any accepted section-break command, when Undo and then Redo run, then the document bytes return exactly to the before and after states.

## Implementation Notes

- **Engine commands** (`folio-go/section_break.go`):
  - `setSectionBreak {offset, snap}` snaps before any check. `removeSectionBreak` clears the key, never writing null.
  - `refuseSectionBreakStraddle` runs on move, resize, bounds, property edits, create, drop, both duplicates, table header height and group move. `refuseSectionBreakBeyondContent` runs on band-height and page-setup changes; its refusal carries data path `bands.content.sectionBreak` and names the break.
- **Projection:** `sectionBreak` (`*int64`, omitempty) and `belowSectionBreak` on content components, present only when a break exists. `sectionBreak` is recorded as an optional wire key (`canvasProjectionOptionalWireKeys`). The projection test is renamed `TestSectionBreakLeavesTheCanvasPaginationUnchanged`.
- **Designer:**
  - Pure maths in `section-break.ts` and command bytes in `section-break-command.ts`; `homeWindow` is now exported from `sheet-stack.ts`.
  - In `App.tsx`: a separate palette entry, `sectionBreakSelected` state, the handle modelled on the band-boundary handle, and `SectionBreakProperties`.
  - The drag proposal line is solid too, because dashed cyan means a band boundary.
- **Known limits:**
  - jsdom has no pointer offsets, so App tests place the break by keyboard Enter at offset 0. Only `e2e/section-break.spec.ts` proves a click places it at a real point.
  - Keyboard Enter-to-place on the content band is refused by the engine (offset 0). *Superseded by review pass 1 below.*
- **Review pass 1 patches:**
  - Keyboard placement uses the middle of the focused sheet's content window, proven against the real engine by a new e2e test.
  - Drag and nudge now stop 1 millipoint inside the band.
  - An armed entry places nothing once a break exists.
  - New App tests: toolbar and Properties Delete, and the window arrow nudge. The select-all clause is dropped from a test title.

## Spec Change Log

- Review pass 1, B1/E1 (intent_gap): keyboard placement sent offset 0, which the engine always refuses, and the intent defined click placement only. The owner chose placement at the middle of the focused sheet's content window, and to keep the implementation and patch it rather than revert and re-derive. Amended: a frozen Placing bullet and a Keyboard place matrix row. Known-bad state avoided: a keyboard author who can never add a break. KEEP: engine commands and refusals, projection and wire pins, palette entry and selection state, handle drag and nudge, Properties Y field, e2e spec, preview parity test, and undo/redo test.

## Review Triage Log

Pass 1 (blind-hunter, edge-case-hunter, verification-gap):

| # | Finding | Verdict | Evidence | Route |
|---|---|---|---|---|
| B1/E1 | Keyboard Enter/Space placement sends offset 0, which the engine always refuses | medium | Confirmed: the band `onKeyDown` calls `placeInBand(band.name, 0, origin/1000)`, and `setSectionBreak` refuses `proposed <= 0`. The App test mocks acceptance. The intent defines click placement only, so the keyboard default is not settled | intent_gap |
| B2/E3 | Drag or nudge clamps to `[0, contentHeight]`, and both ends are always refused | low | Confirmed in `proposedSectionBreak`; the engine refuses `<= 0` and `>= height`. Direct correction of the bounds | patch |
| B3/E4 | Rapid or held arrow nudges reuse a stale offset | low | Real, but repeated identical offsets are engine no-ops with no history entry, as with the band-boundary nudge. The fix adds an in-flight guard | reject |
| B4 | Entry stays armed if a break appears while armed; the next click moves the existing break | low | Confirmed: `armSectionBreak` checks only when arming. Fix is an early return in `placeInBand` | patch |
| B5 | Drag readout shows the unsnapped value | low | Same as the band-boundary readout; the fix adds snapping logic to the browser, which the plan forbids | reject |
| B6 | Proposal may draw past the sheet when dragging across a seam | maybe-false | Would need a multi-sheet drag, and is cosmetic during the drag only; low even if true | reject |
| B7/V-o2 | `belowSectionBreak` has no designer consumer | false | The frozen intent requires the projection to carry membership | reject |
| B8 | `removeSectionBreak` accepts a null break | false | A null `sectionBreak` is a load error (story 1), so `Set && Null` is unreachable | reject |
| B9/E5 | `refuseSectionBreakBeyondContent` ignores a geometry error | false | The `Canvas(t)` that follows needs the same geometry, errors, and restores the prior state | reject |
| B10 | A refused typed Y is reported in the canvas alert, not in Properties | low | The matrix row requires an alert and a revert, and both happen | reject |
| B11a | Toolbar and Properties Delete for a selected break are untested | medium | The verification-gap review pre-verified this (mutation survives) | patch |
| B11b | Escape reverting Y; undo then redo leaves the break selected | low | Cosmetic; no document harm | reject |
| B12 | DESIGN.md omits the proposal, focus ring and disabled-entry note | low | Documentation completeness only | reject |
| B13 | The spec is not in the review diff | false | Deliberate: the spec goes to the edge-case layer only | reject |
| E2 | Undo mid-drag unmounts the handle and leaves the drag ref set | low | Needs Cmd+Z with the pointer held; Escape on the canvas region clears it through `clearInteraction` | reject |
| E6 | Typed "400.0" sends a no-op command | low | The engine records no history for unchanged bytes; only a round trip | reject |
| V1 | Toolbar Delete with the break selected is unverified | medium | Pre-verified | patch (with B11a) |
| V2 | Window arrow nudge with the break selected is unverified | medium | Pre-verified | patch |
| V-o1 | Test title claims select-all but never runs it | low | Direct correction | patch |

## Verification

**Commands:**
- `cd folio-go && gofmt -l . && go vet ./... && go test -count=1 ./...` -- expected: clean, apart from the known baseline failure `TestCorpusMeetsP6ExerciseFloors`.
- `cd folio-designer && npm run lint && npm run typecheck && npm test && npm run test:e2e` -- expected: pass.
