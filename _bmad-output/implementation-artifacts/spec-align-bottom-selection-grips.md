---
title: 'Align the bottom selection grips with the selection border'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
route: 'one-shot'
---

# Align the bottom selection grips with the selection border

## Intent

**Problem:** The selection border pseudo-element used content-box sizing, adding two border pixels to its explicit projected height and drawing its bottom edge below the three grips.

**Approach:** Use border-box sizing on the selection border so its outer height matches the projected element height. Preserve grip positions, hit targets, and resize behavior. Real-browser measurements reproduced a 78px element with an 80px border box before the fix; afterward both match at 50%, 100%, and 150% zoom, with grip centers within half the one-pixel stroke. The resulting screenshot was visually inspected. Independent review found no actionable issues.

## Suggested Review Order

- Keep the selection border within its projected height without compensating in grip offsets.
  [App.css:1253](../../folio-designer/src/App.css#L1253)
