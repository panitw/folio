# Existing designer integration

Inspected on 2026-09-12, including the group-movement scope update. Paths are relative to the repository root; line anchors may move. These are integration constraints and starting points, not permission to broaden scope.

## Reuse and change points

| Source | Current behavior | Implication |
|---|---|---|
| `folio-designer/src/App.tsx`, `select` (~1397) | Selected IDs form an array; Shift-click toggles membership; selection clears column/editor state. | Reuse cleanup for set replacement/union. Do not implement Shift-drag by repeatedly toggling IDs. Preserve a group when pressing a selected body. |
| `App.tsx`, `sheetSurface` (~2700) | Page click clears selection; bands handle placement; occurrences distinguish interactive homes/echoes. | Add rectangle ownership and suppress post-drag clear/collapse clicks. Preserve logical selection across occurrences. |
| `App.tsx`, `CanvasComponent` (~4976) | `begin` always invokes `onSelect` before movement; `DragState` contains one ID; finish emits a singular move. | A selected-body press must retain the whole selection, including over table cells. Capture/preview multiple original positions without changing resize-handle ownership. |
| `folio-designer/src/sheet-stack.ts` | Owns display mapping, projected origins, sheet gaps, home/echo occurrences, drawing cap. | Reuse pure enclosure and pointer-to-document transforms. Never use page index multiplied by window height or independently convert a delta for every group member. |
| `folio-designer/src/canvas-authority-contract.test.ts` | Bans browser layout measurement as geometry authority; existing input-coordinate exceptions are narrow. | Use projected geometry and supported input coordinates. Do not weaken the contract for marquee/group-drag tests; scrolling must fit existing authority rules. |
| `App.tsx`, `ComponentProperties` (~3404) | Filters some typography/table dimensions, but mixed selections restore line borders and non-actionable binding content. | Make common-field/operation intersection testable while retaining authored kind rules. |
| `App.tsx`, `PropertyDraft` (~3542), specialized controls | Shared/mixed values exist; defaults and async drafts vary by control. | Prove untouched controls never write; scope all drafts/results to selection. Display group preview positions without competing edits. |
| `folio-designer/src/component-property-command.ts` | Encodes IDs and shared changes through `command-json.ts`. | Reuse for bulk property edits. Identical absolute X/Y assignments do not express relative group movement. |
| `folio-designer/src/component-command.ts`, `moveComponentCommand` | Encodes one ID plus absolute X/Y and snap flag. | Add engine-backed atomic group transport or equivalent transaction through the central encoder; preserve singular call behavior. |
| `folio-go/component_commands.go`, `moveComponent` (~1820) | Moves one element, snapping/clamping its origin independently. | Do not loop this command for a group. Apply one accepted delta and validate/install all members atomically. |
| `component_commands.go`, `updateComponentProperties` (~800) | Clone/validate/install is atomic; in-place handling (~1031) validates IDs and containment. | Reuse the transactional pattern; prove one group command produces one host history entry and one undo restores all members. |
| `component_commands.go`, `applyPropertyChanges` (~1137) | Defines fields per kind; tables have derived dimensions; multi-ID text value edits are rejected. | Keep property intersections aligned with engine legality and narrower inspector policy. |

## Existing UX and protocol alignment

The Property field section of `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` specifies shared properties and mixed divergent values. The user's clarified request agrees; no upstream UX rewrite is needed for that choice.

Story 14.2 keeps line-specific vocabulary on single selection and generic Width/Height on multi-selection. Preserve that boundary and the preceding line thickness/resize/snapping fix. Bulk content, asset, binding, and table-schema editing remain outside the existing main-inspector surface.

Rectangle selection and bulk properties can reuse existing selected-ID state and property commands. Group movement additionally requires an atomic relative/per-target geometry operation: the present singular move and shared absolute-property command do not provide it. Define and validate any additive command/host wiring during implementation without changing the document format or permitting independent member snapping.

## Suggested verification locations

- Extend `folio-designer/src/App.test.tsx` or focused selection/group-drag tests for gesture precedence, multiple previews, intersection, and draft lifecycle.
- Add pure helpers/tests for enclosure, repeated/split occurrences, shared translation, snap reference, and feasible group displacement; retain DOM-independent geometry.
- Extend `folio-designer/src/line-rect-vocabulary.test.tsx` for line-containing intersections and unchanged thickness during group movement.
- Use command encoder tests and `folio-go/component_commands_test.go` for transport, legality, atomic rejection, common-delta snapping, and boundary fallback. Verify undo/redo with the real host engine.
- Add browser tests under `folio-designer/e2e/` for actual rectangle/group drags, table-body hit targets, page gaps, capture/release, zoom, cancellation, and no accidental clear/collapse.

## Working-tree boundary

Source inspection includes the uncommitted line-resize/snapping fix from the preceding request. Keep those changes. The unrelated `spec-rendering-library-documentation` folder predates this spec. This turn updates specification artifacts only; no application implementation is requested yet.
