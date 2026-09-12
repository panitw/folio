---
title: 'Dragging stays inside the current area'
type: bugfix
created: '2026-09-12'
status: done
baseline_commit: c74df9b7a7444647850f43fa8e1b3a4790e16207
context: []
---

<frozen-after-approval reason="User explicitly requested fixing the bug and confirmed stop at the current area's edge">

## Intent

**Problem:** A content element can be dragged into the footer, where the later footer hit target makes it unselectable. The user explicitly chose “Stop at the current area's edge” over dragging into another page's content.

**Approach:** Constrain pointer movement to the starting content window (or repeating band), accounting for the component's entire box and snapping. Apply the same rule to single-element and selected-group body drags. Keep engine-owned continuous content coordinates and existing documents valid.

## Boundaries & Constraints

**Always:** Preserve existing uncommitted work. Shell commands use `rtk` per `/Users/panitw/.codex/RTK.md`. Geometry comes from engine projection, never browser measurements. Go must keep snapping and acceptance authoritative. Header/footer membership stays unchanged. Pointer previews and committed movement share the same constraint. A group retains a common delta and one Undo entry. Existing long content and numeric property editing remain supported.

**Ask First:** Changes to saved file vocabulary or destructive changes to existing work.

**Never:** Globally cap authored content Y to one page, add cross-band transfers, change pagination, commit/push, or alter unrelated properties or resize behavior.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Content boundary | Single content body dragged far toward header/footer | Full box stops within its starting content window and remains selectable after release | No refusal or jump back |
| Snap | Non-grid-aligned content foot with Snap on/off | Accepted position stays inside bounds | Choose legal snapped position or feasible edge |
| Group | Several selected boxes, possibly different bands/windows | Shared delta stops when any member reaches its own starting boundary | Atomic commit; one Undo |
| Later page | Content body starts on a later displayed sheet | Its own window constrains movement; it does not jump to page one | Preserve column coordinates |
| Existing overflowing content | Component too tall for a window or omitted from pagination | Preserve authored dimensions; allow horizontal movement and avoid increasing existing vertical overflow; orphan home remains accessible inside content | No destructive normalization on load |
| Cancellation | Escape, lost capture, zero travel, stale response | Preserve prior geometry/history | Existing gesture fencing survives |

</frozen-after-approval>

## Code Map

- `folio-designer/src/App.tsx`: `beginSelectedGroup` currently admits only existing multi-selection. Single movement uses local `CanvasComponent` previews. Single sheet lacks band-window clip. Home ignores occurrence.y. Reuse the existing engine-backed group gesture for single bodies if practical, preserving selection and table/Shift behavior.
- `folio-designer/src/use-canvas-selection.ts`: owns engine previews and final moveComponents commands, currently maps deltas across page seams. Pointer drags should request current-window constraints and use direct displacement, without crossing page chrome.
- `folio-designer/src/component-command.ts`: opaque movement command encoding. Prefer an optional `constrainToWindow` boolean on moveComponents, omitted by existing non-drag callers, retaining public continuous movement compatibility.
- `folio-go/group_movement.go`: shared preview/commit solver intersects member ranges and chooses feasible snap. Can derive starting window from Canvas contentWindowOrigins and member Y. Add drag-only constraints here without changing general containComponent.
- `folio-designer/src/sheet-stack.ts`: orphan occurrence fallback currently puts top exactly at foot. Place its selectable footprint inside the content window and honor mapped Y in home painting without changing authored Y; keep paint offsets aligned.
- `folio-designer/src/resize-anchor.ts`: resize behavior is outside scope. Legacy local movement path must not remain an unconstrained active body path.
- Existing native group, hook, App, sheet-stack and authority tests verify invariants. Existing browser tests expecting group previews in footer/gaps or cross-page body drags must be updated to the user's new explicit rule, not silently retained.

## Tasks & Acceptance

**Execution:**
- [x] Implement bounded pointer movement through shared engine preview/commit, including single bodies, group snapping, later windows and existing tall components.
- [x] Keep already stranded/orphan content reachable through correct display mapping; do not rewrite loaded geometry.
- [x] Add/update native and designer tests for the matrix and run affected tests/type checks.
- [x] Verify actual browser pointer hit testing and Undo at the footer boundary (parent owns independent browser regression work).

**Acceptance Criteria:**
- Given a normal content text box, when dragged beyond either vertical edge and released, then its full box stays in the original content area and can be deselected, reselected, and dragged again.
- Given selected members, when any reaches its starting window edge, then all retain relative offsets and Undo restores them together.
- Given content on page two, when dragged upward or downward beyond its current content window, then it stays on page two.
- Given a loaded document, when merely displayed, then authored coordinates, pagination, and history are unchanged.

## Spec Change Log

## Verification

- Native group command/preview tests; designer hook, App, sheet-stack, authority tests; TypeScript checks.
- Parent will build production assets and execute browser regressions for single/group edge behavior, selection hit tests, Snap and Undo. Coordinate browser test edits with parent; parent owns a new `folio-designer/e2e/drag-area-boundary.spec.ts`.

- Implementation verification: `go test . ./wasm` passed 1,443 tests; affected designer suites passed 505 tests, protocol/client/selection suites passed 82 tests, and TypeScript passed. Full designer suite passed: 80 files / 1,513 tests. Parent owns final browser/build verification.
- Full native `go test ./...` also ran: unrelated existing `internal/text` P6g corpus floor failed (7 opaque-name exercises, minimum 20); the affected native packages passed.


## Independent review

Three independent review passes covered edge cases, general regressions, and verification gaps. All actionable findings were patched:

- Adjacent content windows can overlap. The engine now limits movement at the next origin as well as the window foot; a short-component native regression verifies it cannot acquire a later-page home.
- Single-body capture prevented native focus. Body press now explicitly restores focus without scrolling; App and real-browser tests verify inspector-to-body focus followed by Delete and Undo.
- Containment assertions alone allowed premature stopping. Native and browser tests now assert exact maximal feasible downward movement, including Snap on/off and group limits.

The original uncommitted workspace was preserved. Review compared against a pre-change source snapshot, not all unrelated changes since HEAD. No commit or push was made.

## Final verification

- Production `npm run build` passed, including WASM, TypeScript, Vite, font scans and offline verification. Log: `/tmp/folio-drag-boundary-build.log`. Existing chunk-size and offline-cache-margin warnings remain.
- Production Chromium integration: **40 passed**, covering the six boundary tests, 19 selection tests, component manipulation, images and table column binding. Log: `/tmp/folio-drag-boundary-browser.log`.
- Browser invocation used `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=/Applications/Google Chrome.app/Contents/MacOS/Google Chrome` and a temporary config spreading `playwright.config.ts`, with only its webServer changed to `npm run preview -- --host 127.0.0.1 --port 4173` over the completed production build. Temporary config removed. The normal config remains the reproducible build-and-test entry point.
- Exact-bottom Snap on/off, group maximal travel, second-page top/bottom, stranded text re-selection without load normalization, and inspector focus → Delete → Undo all passed. Visually inspected `folio-designer/test-results/drag-area-boundary-single--cfd4b-emains-selectable-Snap-true/footer-boundary.png`: selected text stays immediately above Content's foot.
- E2E TypeScript compilation and lint passed; lint retains eight existing Fast Refresh warnings. Diff whitespace check passed.
- Full designer suite: **1,513 passed**. Affected native command and WASM packages: **1,443 passed**, plus strengthened exact-edge tests. The full native run retains the existing unrelated P6g corpus-floor failure (7 entries versus 20 required).

## Suggested Review Order

**Movement and snapping**

- Follow single and group body presses into the shared engine gesture.
  [App.tsx:1111](../../folio-designer/src/App.tsx#L1111)
- Preview and commit request the same current-window constraint.
  [use-canvas-selection.ts:75](../../folio-designer/src/use-canvas-selection.ts#L75)
- Intersect starting-window bounds and select the closest feasible snapped displacement.
  [group_movement.go:86](../../folio-go/group_movement.go#L86)

**Reachability and evidence**

- Keep previously stranded home occurrences inside the visible content band.
  [sheet-stack.ts:130](../../folio-designer/src/sheet-stack.ts#L130)
- Exercise edge reach, actual pointer hit testing, selection, keyboard actions, and Undo.
  [drag-area-boundary.spec.ts:40](../../folio-designer/e2e/drag-area-boundary.spec.ts#L40)


## Follow-up: selection chrome at band edges

The user reported selection handles and the width × height label being clipped when an element rests at the content top/left edge. Keep the established movement bounds and content clipping, while painting and hit-testing selection chrome outside the clipped content layer. Do not expose overflowing content or duplicate component identities. Verify real pixels for the label, hit testing outside the band for the resize corner, resize/Undo, and the existing bounded body-drag cases.


### Selection chrome implementation and review

Selection outline, dimensions and resize handles now render in a pointer-transparent per-sheet overlay, outside the content paint clip and band stacking contexts. Projected band/component coordinates position it. Original component IDs and accessible names remain unique; React portal event ancestry preserves resizing and keyboard behavior. Armed placement disables overlay handle hit targets.

Review caught two long-element cases: hiding a resizing endpoint would lose its pointer capture, so the active anchor remains mounted until release; an invisible authored-height overlay could enlarge the scroll area, so the chrome container is bounded to the home window while visible midpoint controls retain authored offsets. Off-window idle endpoints are hidden, and the clipped outline omits a false bottom border.

The original screenshot regression was reproduced by comparing the rendered size-label pixels with and without the band clip: the old label was cropped. The fixed test passes at 50%, 100% and 150%, asserts the NW handle receives a hit beyond both top and left boundaries, and resizes/undoes through that handle. A 4000pt unresolved text fixture verifies selection does not add scrollable space or expose off-page endpoints. Product code still does not measure browser geometry; the test exercises scrolling and observes output with Playwright.

Production build passed, including font scans, WASM, TypeScript, Vite and offline release verification; existing size/cache warnings remain. Log: `/tmp/folio-selection-chrome-build.log`. E2E compilation passed. Final browser and full designer suite outcomes are recorded below when complete.


### Follow-up final verification

- All **1,515 designer tests across 80 files** pass; TypeScript and lint pass with the eight existing Fast Refresh warnings.
- Production browser batch exercised 49 cases: 48 passed initially; the remaining image-resize test used obsolete descendant lookup for a handle now in the sheet overlay. Updated it to the exact component-ID accessible name and reran all three image tests successfully. No runtime change was needed after the completed production build. Logs: `/tmp/folio-selection-chrome-production-browser.log`, `/tmp/folio-selection-chrome-image-browser.log`.
- The 10 boundary cases include exact drag caps, legacy orphan recovery, keyboard focus, full label pixels and outside-edge resize hit testing at three zooms, and oversized selection scroll behavior. The other production cases cover group gestures, placement arbitration, image paint/resize, and table column binding.
- E2E TypeScript compilation and diff checks pass. Temporary preview-only Playwright config removed. The normal Playwright config builds production assets before running the same specs.
- Visually inspected the top-left selection screenshot: full label, outline, and handles are visible above/left of the content boundary. No commit or push performed.
