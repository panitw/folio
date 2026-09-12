---
title: 'Fix arrow-key element nudging'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
baseline_commit: 'f22eb4aca4b1ea9ed67cfa19d0c044031fd30b71'
review_loop_iteration: 0
context:
  - '/Users/panitw/.codex/RTK.md'
---

<frozen-after-approval reason="user requested this bug fix; implementation proceeds under existing authorization">

## Intent

**Problem:** Arrow keys request a one-point element move while passing the enabled snap setting. The engine rounds that proposal onto the default six-point grid, leaving a grid-aligned element stationary. Existing mock tests check command dispatch without observing the engine's resulting geometry.

**Approach:** Treat keyboard nudges as precise relative movements: one point per arrow and ten points with Shift. Send these through the existing unsnapped move command and retain engine ownership of validation, history, and resulting geometry.

## Boundaries & Constraints

**Always:** Keep existing single-selection eligibility, editable-target protection, table-dialog protection, design-mode restrictions, and command history. Use canonical millipoints from the engine projection and the existing command encoder. Keep the Snap control's state intact during keyboard nudging.

**Ask First:** Expanding this fix into multi-selection keyboard movement or changing engine geometry contracts requires a separate scope decision.

**Never:** Globally disable snapping, change pointer placement/drag/resize behavior, introduce local authoritative geometry, or change the command protocol.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Default Snap on | Single selected element; arrow key | Exactly one point in requested direction; other coordinate preserved | Existing engine validation |
| Larger movement | Single selected element; Shift+arrow | Exactly ten points in requested direction | Existing engine validation |
| Snap off | Same gestures | Same precise relative movement | Existing engine validation |
| Off-grid origin | Repeated settled nudges | Each starts from returned geometry; no orthogonal-axis snapping | Existing engine validation |
| Inspector edit | Editable field focused | Field owns the key; no canvas move command | Existing editable guard |
| Dialog open | Table editor owns focus | No mutation of the selected element behind the dialog | Existing modal guard |
| History | Successful nudge then undo/redo | Original then nudged position restored | Existing history failure handling |

</frozen-after-approval>

## Code Map

- `folio-designer/src/App.tsx` — `nudgeSelection` adds millipoint deltas to current projected coordinates. The window shortcut listener maps arrows to 1,000 or 10,000 millipoints. Change only the nudge command's snap argument.
- `folio-designer/src/component-command.ts` — `moveComponentCommand` encodes millipoints as point literals and forwards the explicit snap boolean. Reuse unchanged.
- `folio-designer/src/App.test.tsx` — placement/nudge dispatch and table-modal shortcut protection already have coverage; update relevant command expectations and guard precise nudge dispatch.
- `folio-designer/e2e/placed-component-selection.spec.ts` — existing keyboard placement pattern and real-browser selection assertions to reuse.
- `folio-designer/e2e/component-nudging.spec.ts` — new real-WASM regression witness; compare actual inspector geometry and painted position, not just dispatched commands.
- `folio-designer/playwright.config.ts` — runs a production build, real browser, and real engine with one worker.
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md` — snapping is specified for dragging; every canvas mutation is undoable.
- `_bmad-output/implementation-artifacts/deferred-work.md`, DW-227 — previously recorded the same cause for components and band boundaries. This request resolves the component portion; band-boundary behavior stays outside this fix.

## Tasks & Acceptance

**Execution:**
- [x] `folio-designer/e2e/component-nudging.spec.ts` — reproduce the default-snap failure before changing production code, then cover all directions, Shift, Snap off, sequential movement, and undo/redo against the real engine.
- [x] `folio-designer/src/App.tsx` — disable snapping specifically for precise keyboard nudges and add a brief explanation of why.
- [x] `folio-designer/src/App.test.tsx` — update existing nudge expectations and verify the unsnapped intent while retaining shortcut isolation coverage.

**Acceptance Criteria:**
- Given a newly placed and selected element with default settings, when the author presses an arrow, then the engine's returned coordinates and the visible element both move by one point.
- Given successful keyboard movement, when the author undoes and redoes it, then geometry returns to the corresponding prior and subsequent positions.
- Given existing pointer and keyboard contexts, when the relevant regression suites execute, then pointer snapping and editable/dialog shortcut isolation remain intact.

## Spec Change Log

## Verification

**Commands:**
- `rtk npm exec -- vitest run src/App.test.tsx src/TableEditor.test.tsx src/component-command.test.ts` in `folio-designer` — targeted unit and integration suites pass.
- `rtk npm run typecheck` and `rtk npm run test:e2e:compile` in `folio-designer` — source and browser tests compile.
- `rtk npm run lint` in `folio-designer` — lint passes.
- `rtk npm exec -- playwright test e2e/component-nudging.spec.ts e2e/component-manipulation.spec.ts` in `folio-designer` — production build, real-engine nudges, and existing manipulation checks pass.

**Results (2026-09-12):** Original production code failed the real-browser first-arrow assertion: X remained `0`, expected `1`. Red trace: `/tmp/folio-component-nudging-red-2026-09-12/trace.zip`. After the fix, all 477 targeted Vitest tests and all 11 Playwright tests passed. Typecheck, browser-test compilation, lint, and whitespace checks passed; lint retains eight existing Fast Refresh warnings. History clears selection, so the browser test reselects the same component before inspecting restored geometry. Every matrix row is exercised by these passing suites.

## Review Outcome

Three independent review layers completed. No blocking issue or verification gap was found in the snapping change. Two confirmed pre-existing issues were deferred: overlapping keyboard commands can reuse an older projection, and inspector-tab arrow navigation can also reach the global nudge listener. Both exist with Snap off before this fix. Suggestions for further unrelated context/zoom/boundary coverage do not identify defects introduced by this change; engine validation, unit-level pointer snap assertions, and the existing isolation suites remain in place.

The complete baseline diff also included a concurrent, separately committed canvas-banner removal. Review and this commit cover only the nudge fix, its tests, and its tracking documents.

## Suggested Review Order

- Preserve exact keyboard steps while leaving pointer snapping under its existing control.
  [App.tsx:1522](../../folio-designer/src/App.tsx#L1522)

- Check unsnapped commands, fractional positions, and subsequent movement from the returned projection.
  [App.test.tsx:1669](../../folio-designer/src/App.test.tsx#L1669)

- Verify real-engine movement and paint, Snap settings, all directions, and undo/redo.
  [component-nudging.spec.ts:29](../../folio-designer/e2e/component-nudging.spec.ts#L29)

