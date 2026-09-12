---
title: 'Draw thin lines at their declared thickness in the designer'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
route: 'one-shot'
---

# Draw thin lines at their declared thickness in the designer

## Intent

**Problem:** The designer's two-pixel interaction-box floor also enlarged line paint. At 100% zoom, lines with thickness 0.5pt, 1pt, and 2pt all painted as 2px, although their PDF output was correct.

**Approach:** Give the line's paint layer its projected width and height while retaining the outer interaction box and transparent padding. This covers horizontal and vertical line fill, including shared echo styles, without changing the engine or PDF renderer. Clarify existing helper and test terminology so the interaction floor is not confused with painted thickness. Keep DW-345 open for other component kinds and imported borders.

Real-browser verification covered both orientations, 0.5pt/1pt/2pt thickness, and 50%/100%/150% zoom: all 18 combinations painted their scaled thickness and remained selectable through the 12px padded hit area. The deselected 0.5pt screenshot was visually inspected. Relevant line, canvas-authority, and padding suites passed 82 tests; after the helper terminology cleanup, 18 targeted padding/authority checks passed again. Independent review prompted the terminology and deferred-work updates; broader rendering and interaction changes remain outside this fix.

## Suggested Review Order

- Paint line dimensions independently from the minimum interaction-box size.
  [App.css:496](../../folio-designer/src/App.css#L496)

- Preserve the existing twelve-pixel padded hit area using the interaction wrapper's dimensions.
  [App.tsx:5412](../../folio-designer/src/App.tsx#L5412)

- Retain padding assertions while explaining the distinction between line paint and hit area.
  [App.test.tsx:1709](../../folio-designer/src/App.test.tsx#L1709)
