---
title: 'Place PDF navigation above the preview'
type: 'bugfix'
created: '2026-09-11'
status: 'done'
route: 'one-shot'
---

# Place PDF navigation above the preview

## Intent

**Problem:** Page and zoom controls in the bottom status bar are separated from the PDF.

**Approach:** Move the existing controls into a centered toolbar above the PDF, using the Designer canvas toolbar styling. Preserve the footer status information and existing navigation behavior.

## Suggested Review Order

- Reuse the Designer toolbar styling and move navigation into the preview region.
  [App.tsx:2819](../../folio-designer/src/App.tsx#L2819)
- Verify keyboard navigation and zoom controls in their new location.
  [App.test.tsx:8001](../../folio-designer/src/App.test.tsx#L8001)
- Verify centered placement, scroll behavior, and narrow desktop visibility.
  [preview-navigation.spec.ts:208](../../folio-designer/e2e/preview-navigation.spec.ts#L208)
