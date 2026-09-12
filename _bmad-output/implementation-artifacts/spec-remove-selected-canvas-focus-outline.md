---
title: 'Remove the redundant focus outline on canvas elements'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
route: 'one-shot'
---

# Remove the redundant focus outline on canvas elements

## Intent

**Problem:** Canvas elements draw an unwanted blue focus border. Clicking empty canvas clears selection without blurring the element, so hiding the border only while selected makes it reappear after deselection.

**Approach:** Suppress the outline whenever a canvas component is focused, independently of selection. Preserve actual keyboard focus, normal selection chrome, and focus indicators on toolbar and resize controls. Reproduced click-away leaving an unselected component focused with a two-pixel blue outline. After the correction, real-browser checks confirmed that the same retained-focus state has no outline, the image's dashed boundary remains, reselection and one-point/Shift ten-point nudges work, Escape and unselected focus do not restore the border, all eight selection handles remain, and toolbar focus styling is intact. The resulting screenshot was visually inspected. Review clarified the comment; no additional visual focus cue was introduced because the user explicitly requested removal of the canvas element focus indicator.

## Suggested Review Order

- Suppress the redundant outline while retaining existing selection chrome and keyboard behavior.
  [App.css:508](../../folio-designer/src/App.css#L508)
