---
title: 'Make header image placement reliable'
type: 'bugfix'
created: '2026-09-11'
status: 'done'
baseline_commit: 3890ba850d1e447f511f9bcc553cc6f48276fa9c
context:
  - /Users/panitw/.codex/RTK.md
---

# Make header image placement reliable

## Intent

**Problem:** Image placement succeeds only near the top of the default header. A 48pt default image dropped at y=30 in a 60pt header fails containment. At the last three display pixels of the header, the boundary resize strip instead consumes the event without placing anything.

**Approach:** Keep the dropped image anchored at its requested position and reduce any overflowing placeholder width/height to fit the remaining header/footer space. Let palette placement pass through boundary resize handles. Preserve strict geometry validation for explicit creation, movement, and resizing.

## Boundaries & Constraints

**Always:** Engine owns document geometry and containment. Preserve default image dimensions where they fit; shrink only overflowing axes of the empty placeholder (not a loaded asset). Honor the snapped origin when possible, and keep it strictly within the target band if rounding would land on the boundary, leaving positive room. With snap off, preserve the exact in-band drop point. A header/footer smaller than the default image is now a successful resized drop. Do not grow bands. Preserve atomic rejection for invalid points, half-open band hit testing, and unrestricted content vertical placement. Apply identical image drop behavior in repeated header/footer sheet occurrences. User has authorized the bug fix; implement and verify without requesting another plan approval.

**Never:** Treat printable page margins as band content; alter unrelated component types or explicit bounds contracts; add browser layout measurements; change normal boundary resizing.

## Code Map

- `folio-go/component_commands.go`: dropComponent selects band and default image 96x48; containComponent caps header/footer but not content. Fit image placeholder width/height to remaining header/footer space after resolving the origin. Explicit create stays strict. createComponent is an explicit size contract and stays strict.
- `folio-go/component_commands_test.go`: existing drop hit-testing and containment tests; extend here for image placement regression.
- `folio-designer/src/App.tsx`: sheetSurface pointerup around 2771 routes first sheet to dropComponent and subsequent sheets to createComponent. Repeated headers/footer share their template band, so route their image placement consistently without making explicit create permissive. beginBoundaryDrag around 1500 has obsolete comment documenting the swallowed-drop bug.
- `folio-designer/src/App.css`: band-boundary-handle overlaps last 3px of header; disable its pointer events under canvas-region-placing, matching existing selection-handle guards.
- `folio-designer/e2e/component-manipulation.spec.ts`: use real palette pointer gesture and engine to verify header placements; band-boundary-drag.spec.ts already exercises normal resizing.
- Browser reproduction: on blank document header box x398.36 y153 h60 at 1440x1000; image drop y=.1 succeeds, .5/.8 fails geometry, .98 swallowed (palette remains armed).

## Tasks & Acceptance

- [x] Adjust engine image palette sizing at its drop origin, preserving explicit geometry contracts.
- [x] Ensure repeated header/footer image palette placements use the same engine drop semantics.
- [x] Disable boundary handle interception while placing; update stale explanatory comment.
- [x] Add and run Go regressions for header/footer, snap on/off, near-bottom/right edges, content, undersized bands (resize succeeds), and explicit out-of-bounds creation rejection.
- [x] Add and run browser regressions for header midpoint and just above header/content boundary, checking image created in header, selection and disarm. Run normal boundary resize browser checks.

**Acceptance Criteria:**
- Given any positive header/footer area at an in-band drop point, when the default image is larger than the remaining area, then the engine creates an image at that position with only overflowing dimensions reduced to fit.
- Given armed image placement, when released in the resize strip on the header side, then one header image is placed and the band height is unchanged.
- Given an unarmed palette, when dragging the boundary, then normal band resizing works.
- Given content or explicit component geometry commands, when submitted, then their existing containment semantics remain unchanged.

## Verification

- Targeted Go component/drop tests, designer TypeScript checks, browser placement regression and boundary resize tests.
- Development browser available at http://127.0.0.1:4174; use a temporary Playwright config with webServer undefined to avoid stale production builds. Rebuild wasm before browser verification of Go changes with npm run build:wasm. Remove temporary config when done.

## Spec Change Log

- Owner steering: resize the dropped placeholder if it is larger than available space. Replace origin pullback and undersized-band refusal with fitting placeholder dimensions to the remaining header/footer area. Keep the boundary-strip fix, repeated-header/footer routing, strict explicit bounds, and content behavior.

## Review corrections

- Empty repeating bands retain named creation/refusal instead of falling through page hit testing into Content.
- Added repeated-footer keyboard geometry checks, first-sheet footer-strip placement, and resizing after placement.
- Corrected three preview page-rail test selectors left over from the earlier toolbar move in this task.

## Verification Results

- 56 focused Go cases passed for sizing, snap, boundaries, undersized bands, and strict explicit geometry.
- Rebuilt WASM and app/e2e TypeScript checks passed.
- App and authority suite: 456 tests initially passed; three earlier toolbar selector regressions were corrected and all five affected page-rail tests passed on rerun. Both new zero-height keyboard guards also passed.
- All 12 placement and boundary browser tests passed, including repeated-band pointer/keyboard paths and resize restoration.
- Manual browser reproduction at header fractions 0.1/0.5/0.8/0.98 now produces y/height 6/48, 30/30, 48/12, 54/6 without errors. Default 40pt footer fits a 40pt image.

## Suggested Review Order

- Fit placeholder dimensions at the engine-owned drop origin.
  [component_commands.go:1678](../../folio-go/component_commands.go#L1678)
- Route repeated bands consistently and reject empty target bands by name.
  [App.tsx:2720](../../folio-designer/src/App.tsx#L2720)
- Allow palette drops through resize strips.
  [App.css:246](../../folio-designer/src/App.css#L246)
- Check snap, containment, and strict geometry invariants.
  [component_commands_test.go:642](../../folio-go/component_commands_test.go#L642)
- Exercise actual drop, keyboard, and resize gestures.
  [component-manipulation.spec.ts:87](../../folio-designer/e2e/component-manipulation.spec.ts#L87)
