---
title: 'Header and footer editable from any page'
type: 'feature'
created: '2026-09-14'
status: 'done'
route: 'dispatch'
baseline_commit: '54a6e7702813c391a03f47378f67cbb01ac384be'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-multi-pages/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/multi-pages-decision-log.md'
  - '{project-root}/_bmad-output/specs/spec-multi-pages/stories/3-elements-on-a-specific-page.md'
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the frozen Intent below governs implementation.*

**The problem.** The header and footer are drawn on every page, but you can only click them on page 1. On page 3 a click on the header passes straight through to the page underneath.

**What this story builds.**
- **Selecting:** click a header or footer element on any page to select it. That page becomes the current page, and the element's handles, name and keyboard focus move to that page's copy.
- **Editing:** move, resize or edit it in the properties panel there. Every page's copy updates together, because there is only one header and footer.
- **Current page:** the page you last selected something on. With nothing selected it is page 1.

**What it does not do.**
- **No per-page header or footer:** every page shares the same one.
- **Engine unchanged:** the engine already stores one header and footer and prints them on every page.
- **Page Break off and later section breaks:** both are story 5.

**What done looks like.** In a three-page contract, an author clicks the party name in page 3's header and changes it in the properties panel. Pages 1 to 3 show the new name. One undo reverts it everywhere.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Header and footer copies are interactive only on sheet 0 (`home: sheet.index === 0`), so an author can't select or edit the shared header or footer from page 2 onward.

**Approach:** The designer tracks a current page (D-G.2). The interactive, accessibly named copy of every header and footer component is drawn on the current page's first sheet. Every other copy stays an aria-hidden echo that responds to the pointer. Pressing an echo selects the component and makes that echo's page current. Designer-only; no engine command changes.

## Boundaries & Constraints

**Always:**
- **The current page (D-G.2).**
  - It is the page the author last selected something on.
    - Pressing a header or footer copy makes that copy's page current.
    - Selecting a content element makes its page current.
    - Selecting a page makes that page current.
  - With nothing selected and no page selected, it is page 1.
  - It is clamped when pages are deleted or undone away.
- **Exactly one interactive copy per header or footer component.**
  - It is drawn on the current page's first sheet (`pageStart`).
  - It carries the accessible name, the `data-component-id`, the tab stop and the selection handles.
  - Every other copy is a `ComponentEcho`: aria-hidden, no role, no id.
- **Echo presses.**
  - When not placing, pressing any header or footer echo, selected or not, behaves like pressing the component: a plain press selects it and can start a drag, and Shift toggles it.
  - It also makes the echo's page current.
  - Keyboard focus lands on the interactive copy on that page.
  - While placing, echoes stay pass-through, as today.
- **A press on the same page's continuation sheet** selects the component, but the interactive copy stays on that page's first sheet.
- **Content echoes are unchanged:** continuation-sheet copies of content stay inert as today.
- **One-page documents.** The current page is always page 1, so names, DOM and bytes are unchanged, except that header and footer echoes on overflow sheets now select their component.
- **Commands.** Every header and footer command is byte-identical from any page, and none carries `page`.
- **Accessible names count designed pages (D-4.1).**
  - **Scope:** this applies only in documents with more than one designed page.
  - **Sheet and band names:** they use the designed page number, "Report page 2 of 3 with Page Header, Content, and Page Footer" and "Page Header on page 2 of 3".
  - **Continuation sheets:** a sheet that isn't its page's first adds ", sheet K of M", counting every sheet, so every name stays unique.
  - **Column position notice:** it names the sheet, "on canvas sheet K of M", instead of a page number.
  - **One designed page:** documents with one designed page, overflow sheets included, keep today's names exactly.

**Never:**
- No per-page header or footer, override or hiding.
- No second accessibly named copy of a component.
- No engine change.
- No change to how content echoes behave.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Select from page 3 | 3 pages; header `h1`; press h1's copy on page 3 | h1 selected; page 3 current; the named copy and handles are on page 3's first sheet; focus on it | N/A |
| Edit from page 3 | Continue; change h1's text in properties | Today's `updateComponentProperties` bytes; every copy shows the new text; one undo | N/A |
| Drag from page 2 | Press h1's unselected copy on page 2 and drag 10px | Today's `moveComponents` bytes for h1, with no `page` | N/A |
| Content selection moves current | h1 current on page 3; click content `e1` on page 1 | Page 1 current; h1's named copy on page 1 | N/A |
| Page selection | Click the Page 2 label | Page 2 current; header named copies on page 2 | N/A |
| Nothing selected | Escape | Page 1 current | N/A |
| Deleted current page | Page 3 current; delete page 3 | Current clamps to the last page; one named copy | N/A |
| Continuation sheet | Page 1 spans sheets 1–2; press h1's echo on sheet 2 | h1 selected; named copy stays on sheet 1 | N/A |
| Shift on echo | h1 and e1 selected; Shift+press h1's page-2 echo | h1 removed from the selection | N/A |
| Placing | Place Rectangle armed; click header echo on page 2 | Places into the header as today; no selection change | N/A |
| One page, overflow | 1 page, 2 sheets; press h1 echo on sheet 2 | h1 selected; named copy on sheet 1; names unchanged | N/A |
| Names, multi-page | 2 pages; page 1 spans sheets 1–2, page 2 is sheet 3 | Bands: "Content on page 1 of 2", "Content on page 1 of 2, sheet 2 of 3", "Content on page 2 of 2" | N/A |

</frozen-after-approval>

## Code Map

- `folio-designer/src/App.tsx`:
  - **Where copies are decided:** `sheetSurface` at :3147 sets `home: sheet.index === 0` for header and footer occurrences; this becomes the current page's `pageStart` sheet. `paint` at :3149-3151 chooses `CanvasComponent` or `ComponentEcho`.
  - **Selection state:** `selectedPage`/`pageSelection` (:377, :621), `selectPage` (:1865), `installSelection` (:1231, which clears `selectedPage`) and `select` (:1587) have no page input yet. Add current-page state set by these paths and by echo presses.
  - **Focus by id:** `setPendingFocus` and the effect at :1663-1673 look a component up by `[data-component-id]`, and the move remounts the node (its key changes). Use pending focus when an echo press moves the named copy.
  - **Echo component:** `ComponentEcho` (:5880-5883) handles `onPointerDown` only when selected. Give header and footer echoes a press handler whether selected or not (select, current page, `beginSelectedGroup`), keyed by band.
  - **Must stay as is:** the band boundary handle stays on sheet 0 (:3181-3182).
  - **Names (D-4.1):** the page-surface `pageOf` (:3108, :3113), the band `aria-label` (:3182) and `canvasColumnPositionNotice(sheet.index + 1, sheets)` (:3150, text at :5865) use sheet numbers today.
  - **Name pins:** one-page multi-sheet names are pinned by `App.test.tsx` :6228-6349 and `e2e/component-manipulation.spec.ts` :155-190, and must not change. Multi-page names in `App.test.tsx` :11019, :11110-11128 and `e2e/pages.spec.ts` :98-108 follow the new rule.
  - **Where a copy's page comes from:** `componentPage` (sheet-stack.ts:85) is 0 for header and footer, so an echo's page comes from `sheet.page`.
- `folio-designer/src/App.css`:
  - :224-225 and :1344 make echoes `pointer-events: none`, except when selected and not placing. Add a repeating-band exception (header and footer echoes, not while `.canvas-region-placing`).
  - The table-column `:not(.canvas-component-echo *)` rule (:552, pinned by table-column-binding.test.tsx:581-610) must stay.
- Tests pinning today:
  - `App.test.tsx`: :6166-6183 (4 echoes, unique h1/f1; keep), :7322-7336 (handles on the home sheet), :11009-11025 (header h1 on `pages(2)`).
  - `e2e/selection-group-properties.spec.ts`: :342-371 (drag via a selected echo; Shift behaviour) and :435-448 (placing passes through), which must still pass.
  - `e2e/component-manipulation.spec.ts`: :155-190 (a one-page, two-sheet header or footer image), which must still pass.
- Helpers: the multi-page describe `App.test.tsx:10900+` (`pages`, `text(id, page, y, band)`, which needs a `pageFooter` option, `open`, `sent`, `surfaces`, `label`).

## Tasks & Acceptance

**Execution:**
- [ ] `folio-designer/src/App.tsx` -- current-page state; named copy on the current page's first sheet; echo press selects, sets current, focuses -- UI
- [ ] `folio-designer/src/App.css` -- header and footer echoes respond to the pointer when not placing -- hit testing
- [ ] `folio-designer/src/App.test.tsx` -- a matrix row each; the existing pins still pass -- UI coverage
- [ ] `folio-designer/e2e/pages.spec.ts` -- real-browser: edit a header text from page 3 and see every copy change, then undo; echo hit-testing on page 2 (compiled; run deferred) -- e2e

**Acceptance Criteria:**
- Given any multi-page state, when the canvas renders, then each header and footer component has exactly one element with its accessible name and `data-component-id`.
- Given a one-page document, when nothing about pages is touched, then accessible names and every command byte are unchanged.

## Spec Change Log

## Review Triage Log

| # | Source | Finding | Verdict | Route | Evidence |
|---|--------|---------|---------|-------|----------|
| 1 | blind, edge-case | A Shift-press that removes the only selected header or footer still sets the current page, so the next press remounts the named copy on another page; the Shift branch also skips `preventDefault` | medium | patch | `pressRepeatingEcho` calls `setCurrentPage(page)` after `select(id, true)` empties the selection. Then `installSelection([h1])` keeps the stale ref because the selection is header-only. |
| 2 | blind, edge-case, verification-gap | Locate never clears `selectedPage`, so a page selection outranks the located element's page; the path is untested | medium | patch | `returnWithOptionalSelection` calls `setSelected([id])` directly, and `currentPage = pageSelection ?? …`. |
| 3 | edge-case | Escape (or another selection change) moves the named copy away from keyboard focus, which falls to `<body>` | medium | patch | `installSelection([])` sets page 1; the focused page-3 copy unmounts (its key changes to `id@sheet`). |
| 4 | verification-gap, blind | The branches that keep or move the current page, and footer echo presses, are unpinned | medium | patch | Pre-verified gap. No row clicks the named copy after the current page moves, selects content on two pages, or presses a footer echo. |
| 5 | blind | The "Deleted current page" matrix row expects the last page, but a real delete clears the selection, so page 1 is current | low | reject | The frozen constraint "with nothing selected and no page selected, it is page 1" governs. The clamp covers a selection that survives, which the row's fake answer exercises. The only effect is which sheet holds the named copy. |
| 6 | edge-case, blind | The clamp is applied only when reading; `currentPageRef` stays stale after the page count drops | low | reject | It needs a selection that survives a page-count drop (delete and undo both clear the selection), and the fix adds a sync effect. |
| 7 | blind | An unselected echo press doesn't pass `grabAt` | false | reject | `grabAt` returns `undefined` for non-content bands. |
| 8 | edge-case | Focus goes to an off-screen sheet-0 copy when the sheet cap truncates the current page | low | reject | Only at the `MAX_CANVAS_SHEETS` cap; the fallback copy is still the one named copy. |
| 9 | blind | The e2e test is thin and its fixture assumes the shape of `starter.folio` | low | reject | The Playwright run is deferred to the end of the run, where a fixture mismatch fails loudly. |
| 10 | blind | Matrix rows are tested by Enter, not click, and the placing row checks only "nothing sent" | low | reject | Select and Enter share the `select` path; placing through an echo is covered in a real browser by `selection-group-properties.spec.ts:435-448`. |
| 11 | blind | The decision behind D-4.1 is missing from the diff | false | reject | D-4.1 is in `multi-pages-decision-log.md`; the review diff covered only code directories. |

## Verification

**Commands:**
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` -- expected: pass (Playwright deferred to end of run)
- `cd folio-go && go build ./... && go vet ./... && go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` -- expected: pass (no engine change)

**Manual checks (if no CLI):**
- `npm run dev`: add two pages, click the header on page 3, edit its text, and confirm all three copies change; undo.
