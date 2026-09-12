---
title: 'Remove the redundant focus outline on selected canvas elements'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
route: 'one-shot'
---

# Remove the redundant focus outline on selected canvas elements

## Intent

**Problem:** Nudging a selected canvas element activates a thick two-pixel focus outline outside its existing selection border and handles.

**Approach:** Suppress the focus outline only when a canvas component is selected. Preserve actual keyboard focus, the selection layer, unselected-component focus indicators, and focus indicators on controls. Real-browser checks covered keyboard placement and pointer selection followed by uninterrupted one-point and Shift ten-point nudges, continued component focus, the one-pixel selection border, eight resize handles, and restored focus outlines after deselection. Before/after screenshots were inspected. Review refined selector specificity and the comment; the remaining suggestions did not require changes to this scoped visual fix.

## Suggested Review Order

- Suppress the redundant outline while retaining existing selection chrome and keyboard behavior.
  [App.css:507](../../folio-designer/src/App.css#L507)
