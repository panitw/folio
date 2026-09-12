---
title: 'Preserve line thickness during canvas resizing'
type: 'bugfix'
created: '2026-09-11'
status: 'done'
route: 'one-shot'
---

# Preserve line thickness during canvas resizing

## Intent

**Problem:** Lines exposed all eight box resize handles, allowing canvas drags to change Thickness.

**Approach:** Horizontal lines expose west/east endpoints; existing vertical lines expose north/south endpoints. Dragging changes only length and its associated origin. Thickness remains an inspector edit. The line cannot shrink past the orientation-preserving minimum. Rectangles and other components retain existing handles.

Line bounds honour the Snap toggle. The engine snaps position and length, excluding the line’s thickness axis. At band edges it adjusts position rather than reducing thickness. Near the minimum length it retains precise length if grid rounding would collapse the line or flip its orientation. Snap off keeps exact coordinates. This corrects the initial implementation, which disabled snapping for line resizing. No serialized format or command-schema changes.

## Suggested Review Order

- [Canvas gesture and endpoint handles](../../folio-designer/src/App.tsx#L4976): orientation comes from the captured drag geometry; only length handles are offered.
- [Bounds commit](../../folio-designer/src/App.tsx#L2732): the current Snap setting reaches the engine unchanged.
- [Engine bounds snapping](../../folio-go/component_commands.go#L1911): snap position and length while preserving thickness and containment.
- [Engine regression tests](../../folio-go/component_commands_test.go): real command execution covers both orientations, square and short lines, fractional thickness, Snap on/off, and band edges.
- [Minimum length](../../folio-designer/src/resize-anchor.ts#L42): optional minimum keeps resize gestures from flipping derived orientation.
- [UI regression coverage](../../folio-designer/src/line-rect-vocabulary.test.tsx#L710): handles, diagonal drags, committed geometry, cross-axis-only motion, and rectangle regression.
- [Anchor floor checks](../../folio-designer/src/resize-anchor.test.ts#L97): both axes and opposite endpoint preservation.

Initial validation: full unit suite passed (78 files / 1448 tests) before final review edits; focused unit tests and forced TypeScript check rerun after those edits. Lint passed with 8 existing Fast Refresh warnings; browser-test compilation passed. Suites that did NOT run: browser execution, Go suites, build, offline verification, matrix legs, font-host scans.

Snap correction validation: focused Go command/snapping tests and designer regression tests passed; the WebAssembly engine was rebuilt for the designer. Forced TypeScript verification also run.
