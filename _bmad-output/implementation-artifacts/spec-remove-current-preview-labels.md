---
title: 'Remove redundant current PDF preview labels'
type: 'bugfix'
created: '2026-09-11'
status: 'done'
route: 'one-shot'
---

# Remove redundant current PDF preview labels

## Intent

**Problem:** The production preview repeats its state in a heading and badge above the toolbar.

**Approach:** Remove the preview heading and visually hide the current status badge. Preserve the accessible PDF description and visible stale, rendering, and failure messages.

## Suggested Review Order

- Start the current preview with its toolbar and preserve the accessible status.
  [App.tsx:2938](../../folio-designer/src/App.tsx#L2938)
- Verify admission still gates current status and the redundant heading stays absent.
  [App.test.tsx:948](../../folio-designer/src/App.test.tsx#L948)
- Check the current preview and stale transition in a real browser.
  [application-shell.spec.ts:50](../../folio-designer/e2e/application-shell.spec.ts#L50)
