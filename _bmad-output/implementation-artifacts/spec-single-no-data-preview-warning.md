---
title: 'Show one no-data preview warning'
type: 'bugfix'
created: '2026-09-11'
status: 'done'
route: 'one-shot'
---

# Show one no-data preview warning

## Intent

**Problem:** The PDF preview repeats the no-data disclosure in a heading, status badge, and warning box.

**Approach:** Keep the warning as the single visible disclosure for a current stand-in preview. Preserve the accessible PDF status and visible non-current freshness messages.

## Suggested Review Order

- Consolidate the visible disclosure while retaining freshness announcements.
  [App.tsx:2938](../../folio-designer/src/App.tsx#L2938)
- Update current no-data preview assertions.
  [App.test.tsx:7730](../../folio-designer/src/App.test.tsx#L7730)
- Align browser assertions and scroll measurements with the remaining warning.
  [preview-no-data.spec.ts:65](../../folio-designer/e2e/preview-no-data.spec.ts#L65)
  [preview-navigation.spec.ts:75](../../folio-designer/e2e/preview-navigation.spec.ts#L75)
