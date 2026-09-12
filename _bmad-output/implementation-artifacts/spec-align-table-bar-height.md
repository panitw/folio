---
title: 'Align the table bounds and gray header bar'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
route: 'one-shot'
baseline_commit: '6772c953b8dc1d352e800032c9b98059a95befce'
---

# Align the table bounds and gray header bar

## Intent

**Problem:** The gray table bar sized itself from fixed text and padding, extending below the 12-point table and selection box.

**Approach:** New tables start at 24 points. The gray bar fills the projected component height, and its metadata shrinks only when needed to fit a short bar. Existing saved header heights remain unchanged. Selection, placement, and render geometry continue to come from Go.

Verification passed: full native root and WASM packages, 51 canvas tests, browser test typecheck, and 11 production browser checks. Browser coverage compares gray bar, component and selection bounds at heights 12/24/36 and zoom 50/100/110/200 percent, including metadata containment. Updated the disabled-column-hit-test control to assert canvas fall-through after the grid moves below the correctly sized header.

Review: applied metadata-fit, independent free-box-height test, and fractional-zoom coverage improvements. Remaining review suggestions were outside this focused change or covered by existing invariants; no work deferred.

## Suggested Review Order

- Give newly placed tables a 24-point header.
  [component_commands.go:1652](../../folio-go/component_commands.go#L1652)

- Keep the gray bar and its metadata inside projected bounds.
  [App.css:419](../../folio-designer/src/App.css#L419)

- Verify default height, persisted geometry, and capped-band snapping.
  [component_commands_test.go:3520](../../folio-go/component_commands_test.go#L3520)

- Compare visible bounds and metadata containment across heights and zoom levels.
  [table-placement.spec.ts:68](../../folio-designer/e2e/table-placement.spec.ts#L68)

- Preserve the real-pointer negative control after the header is correctly sized.
  [table-column-binding.spec.ts:192](../../folio-designer/e2e/table-column-binding.spec.ts#L192)
