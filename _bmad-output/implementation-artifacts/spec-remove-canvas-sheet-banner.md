---
title: 'Remove the canvas sheet-count banner'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
route: 'one-shot'
---

# Remove the canvas sheet-count banner

## Intent

**Problem:** The “Showing 1 sheet…” banner adds unwanted text above the canvas.

**Approach:** Remove the entire banner, its copy helper and unused styling. The user's explicit removal request supersedes Story 7.6's earlier canvas-wide disclosure requirement. Keep the sheet model, drawing limits and per-component accessible names.

## Suggested Review Order

- The canvas proceeds directly from its controls to the sheets.
  [App.tsx:2945](/Users/panitw/Projects/folio/folio-designer/src/App.tsx:2945)

- Canvas styling no longer includes the removed banner.
  [App.css:193](/Users/panitw/Projects/folio/folio-designer/src/App.css:193)

- The sheet model retains its metadata without promising a banner.
  [sheet-stack.ts:19](/Users/panitw/Projects/folio/folio-designer/src/sheet-stack.ts:19)

- Existing tests verify sheet rendering while expecting no banner.
  [App.test.tsx:5656](/Users/panitw/Projects/folio/folio-designer/src/App.test.tsx:5656)
