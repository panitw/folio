---
title: 'Elements on a specific page'
type: 'feature'
created: '2026-09-14'
status: 'done'
route: 'dispatch'
baseline_commit: 'd2eb64a030182f47a31cc3d35a56634fd555ba3c'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-multi-pages/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/multi-pages-decision-log.md'
  - '{project-root}/_bmad-output/specs/spec-multi-pages/stories/2-pages-on-the-canvas.md'
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the frozen Intent below governs implementation.*

**The problem.** After story 2 you can add pages, but you can't put anything on them. Placing an element from the palette onto page 2 is refused, and dragging an element stops at the edge of its own page.

**What this story builds.**
- **Placing:** a palette element placed on any page's content lands on that page, where you placed it.
- **Moving:** dragging an element onto another page's sheet moves it to that page, at the drop point, as one undo step. Several selected elements from one page move together and keep their arrangement. While you drag, the preview follows the pointer onto the other page.
- **Header and footer:** placing into them works as before, from any sheet.

**What it does not do.**
- **Header and footer from any page** (story 4): editing them there isn't part of this story.
- **Section breaks on later pages** (story 5): not part of this story either.
- **Duplicate and paste:** they still put the copy on the same page as the original.
- **Select all:** it still selects across every page.

**What done looks like.** In a three-page contract, an author places a signature line on page 3. They drag a clause from page 1 onto page 2, then undo, and the clause is back on page 1. The saved file lists each element under the page it's on.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Palette placement onto a page after page 1 is blocked, because the engine creates every content element on page 1. A drag can't move an element to another page, because every move stays within the element's own page.

**Approach:**
- Add an optional target `page` to the engine's `createComponent`, `dropComponent` and `moveComponents` commands, each still one undo entry.
- The designer sends the page under the pointer when placing, and when a drag ends over another page's content band.
- The drag preview draws on the sheet under the pointer.

## Boundaries & Constraints

**Always:**
- **One-page documents.** When `page` is omitted, command bytes, field counts, results and refusal messages are unchanged. The designer omits `page` for page 1, so one-page documents send today's exact bytes.
- **Placement.** Placing a palette element on page N's content band creates it in page N's content at the page-local position placed. Story 2's `accepts` guard is removed. Placing into the page header or footer is unchanged.
- **Cross-page move.**
  - A drag whose pointer is released over another page's content band sends one `moveComponents` with that target `page`.
  - The moved elements leave their source page, join the target page's `Elements` at page-local positions under the drop point, keep their ids, and are one undo entry.
  - A drag released over its own page behaves exactly as today, window constraint included.
  - **Selected elements move together (D-3.1).** When every selected element is on one page and the drag ends over another page, all of them move to the target page and keep their offsets from one another. A selection spanning pages keeps today's per-page clamped move and never changes page.
- **Engine checks on the target page.**
  - The move is contained in the target page's content band.
  - The section-break straddle check applies only when the target is page 1.
  - A move that would leave a keep-together group's members on more than one page is refused as a whole (SPEC constraint). Moving every member of the group together is allowed.
  - A refused move leaves the document unchanged and names the element.
- **Drag preview.** While the pointer is over another page's content band, the moving elements draw on that sheet at the pointer. Otherwise they draw as today.
- **Unchanged scopes.** Duplicate and paste keep copies on the source's page. Select all keeps its current scope. Resize stays within the element's page.
- **Engine rules still hold.** The engine is the only layout authority, and the DOM-measurement ban holds.

**Never:**
- No move of an element into or out of the page header or footer.
- No reordering of pages.
- No new command kind when an optional field on the existing command does the job.
- No change to header and footer interaction (story 4).
- No section break on later pages (story 5).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Place on page 2 | 2 pages; Text placed at (100, 50) on page 2's content band | `createComponent` or `dropComponent` with `page: 1`; element in `pages[1]` at y 50; page 1 unchanged | N/A |
| Place on page 1 | Any doc; place on page 1 | Today's bytes (no `page` field) | N/A |
| Place by keyboard | Page 2's content band focused, Enter while armed | Created on page 2 at the band's top-left | N/A |
| Drag to another page | e1 on page 1 dragged, released over page 2's content at page-local y 200 | One `moveComponents` with `page: 1`; e1 now in `pages[1]` at y 200; one undo restores the original bytes | N/A |
| Drag within page | e1 on page 2 dragged within page 2 | Today's `moveComponents` bytes and window clamp | N/A |
| Target outside band | Drop point leaves the element beyond the target content band's width | Refused, document unchanged | Located refusal naming the element |
| Straddle on page 1 | Element moved onto page 1 across page 1's section break | Refused | Existing straddle refusal |
| Target is page 2 with a break on page 1 | Element moved onto page 2 at page 1's break offset | Accepted | N/A |
| Group to another page | e1 and e2 selected on page 1 (sharing a keepTogether tag), dragged over page 2 | One `moveComponents` with `page: 1`; both in `pages[1]` with their relative offset kept | N/A |
| Selection spans pages | e1 on page 1 and e3 on page 2 selected, dragged | Today's per-page clamped move; no `page` field | N/A |
| Split a group | e1 and e2 share a keepTogether tag on page 1; only e1 is moved to page 2 | Refused, document unchanged | Located refusal naming the group |
| Unknown page | `page: 5` on a 2-page doc | Refused | Located at `pages` |
| One-page doc with page | `page: 1` on a 1-page doc | Refused | Located at `pages` |
| Preview | Pointer over page 2 while dragging a page-1 element | Preview drawn on page 2's sheet at the pointer | N/A |

</frozen-after-approval>

## Code Map

- `folio-go/component_commands.go`:
  - `createComponent` (:1875, 9 fields), `dropComponent` (:1961, 6 fields), `createComponentInBand` (:2040), `bandByName` (:1853) and `hitTestBand` (:2144) route content to `firstContentBand`. Add an optional `page` and resolve `t.doc.ContentBands()[page]`.
  - `containComponent` (:2705) checks x/width on content.
  - `findComponent` (:2167) finds the source band, and `deleteComponent` (:2400) is the splice precedent.
  - Refusals go through `componentFailure`, and optional-field counting follows `addPage`'s `after` (`pages_command.go:97`). `pageIndexField` (`pages_command.go:80`) reads a 0-based page index and refuses out-of-range values.
- `folio-go/group_movement.go` -- `moveComponents` (8 fields, +1 with `constrainToWindow`; ~:136/:214), `groupMemberIndex` (:29, page-filtered windows), `solveComponentMove` (:93). Add an optional target `page`: for the cross-page case, solve positions in the target page's column, splice members across pages, then run the straddle check against the target.
- `folio-go/section_break.go:195` -- `refuseSectionBreakStraddle` uses the page in `contentPageIndex`, which is the source page. It needs to judge the target page.
- `folio-go/internal/template/parse_pages.go:129` -- `refuseKeepTogetherAcrossPages`, a load-time rule a move must never break. `wasm/engine.go:298-345` `Apply` reparses canonical bytes, but moves must refuse before mutating.
- `folio-designer/src/component-command.ts` (:34-41, :82-84) with its test `component-command.test.ts` -- add an optional `page` to `createComponentCommand`, `dropComponentCommand` and `moveComponentsCommand`, written last and only when given.
- `folio-designer/src/App.tsx`:
  - **Placement:** `sheetSurface` has the `accepts` guard (:3112-3116); remove it. Placement goes through `place` and `placeInBand` (:1523-1560) and `placementPoint` (:5496). On a later page the drop sends `page`, and y is page-local (`origin/1000 + offset`, where `sheet.origin` is page-local).
  - **Body drag:** it goes through `beginSelectedGroup` (:1245-1258) into `canvasSelection.beginGroup`. The `moveComponentCommand` arm at :3126 is reached only by resize.
  - **Preview:** `translatedCanvas` at :3080. The `.band-window` clip is lifted by `band-window-open` (:3128-3130).
- `folio-designer/src/use-canvas-selection.ts` -- `finishGroup` (commit :77) and preview (:96) send `dx/dy` with `constrainToWindow`. Detect the target page with `columnForStackY(stack, canvas, zoom, stackY).window` → `stack.sheets[window].page` (`sheet-stack.ts:219`), using the stack-y construction from `beginRectangle` (`App.tsx:1243`).
- `folio-designer/src/canvas-selection.ts` -- `translatedCanvas`, which the preview uses. Extend it so the moving ids draw on the target page.
- Tests pinning current behaviour:
  - `App.test.tsx`: :6313-6326 (body drag → `moveComponents` constrained), :11100-11108 (later-page drop refused; rewrite it), :11018-11024 (select all across pages).
  - `use-canvas-selection.test.ts:66-74`.
  - e2e: `drag-area-boundary.spec.ts:89` and `selection-group-properties.spec.ts:141` (page-two movement stops at its own edges, which must still hold for a drop on its own page).
- Go tests to extend: `component_commands_test.go` (:827 drop hit-testing, :1199 move), `group_movement_test.go` (:264, :344), `pages_command_test.go`, `wasm/pages_test.go` (:15, :113).

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/component_commands.go`, `folio-go/group_movement.go`, `folio-go/section_break.go` -- optional `page` on create, drop and group move; cross-page splice as one mutation; target-page containment and straddle; located refusals -- engine
- [x] `folio-go/component_commands_test.go`, `folio-go/group_movement_test.go`, `folio-go/pages_command_test.go`, `folio-go/wasm/pages_test.go` -- engine matrix rows, byte-identical one-page commands, and a single undo restoring bytes after a cross-page move -- engine coverage
- [x] `folio-designer/src/component-command.ts` (+ test) -- optional `page` builders -- one JSON authority
- [x] `folio-designer/src/App.tsx`, `use-canvas-selection.ts`, `canvas-selection.ts` -- remove the `accepts` guard, send the target page on placement and cross-page drag, draw the preview on the target sheet -- UI
- [x] `folio-designer/src/App.test.tsx`, `use-canvas-selection.test.ts`, `canvas-selection.test.ts` -- UI matrix rows; rewrite the story-2 later-page-drop row -- UI coverage
- [x] `folio-designer/e2e/pages.spec.ts` -- place on page 2, drag a page-1 element onto page 2, undo, save shows it under `pages[1]` (compiled; run deferred) -- e2e

**Acceptance Criteria:**
- Given a one-page document, when the author places, drags or resizes, then every command sent is byte-identical to today's.
- Given a cross-page move was refused, when the snapshot updates, then nothing moved, the preview is cleared and the refusal is shown.

## Implementation Notes

**How it was built.**
- **Engine:**
  - `optionalPageField` (`pages_command.go`) reads the optional 0-based `page`, counted as a field only when present.
  - `createComponent` and `dropComponent` pass it to `createComponentInBand`, which creates into `ContentBands()[page]`.
  - A `page` on a header or footer target is refused at `pages`.
  - `refuseSectionBreakStraddleOnPage` judges the page the element will sit on. The old `refuseSectionBreakStraddle` delegates to it with the element's current page.
- **Group move** (`group_movement.go`):
  - With `page`, every member must be content and on one page, otherwise it is refused at `pages`. When that page is the target, the move is today's.
  - Otherwise it is cross-page. Snapping still applies, but the band and window clamps do not, so the preview answers wherever the pointer is.
  - `moveComponentsToPage` checks each member before splicing: target containment, safe geometry, the page-1 straddle, and a keep-together group split (refused at `pages`, naming the element and the tag). It then splices the members, in document order, onto the end of the target page's elements.
  - Install goes through canonical bytes (`installMovedComponents`), so it is one undo entry.
- **Designer:**
  - The builders write `page` last, only when given.
  - Placement on a later page's content band sends `createComponent` with `page`, using page-local y. The keyboard path does the same.
  - The narrowed guard only stops a section break from being placed on a later page (story 5).
  - Body presses on content pass a grab point down the stack: the occurrence's top plus the press offset when the component itself is the event target.
  - `useCanvasSelection` finds the page under the pointer with `contentPageAt`. When that page's content band differs from the selection's single page, it sends `page`, with `dy` measured in the target column from the grab.
  - A cross-page commit is sent even at a zero delta.
  - `translatedCanvas` re-pages the moving components, so the preview draws on the target sheet.
- **Tests:**
  - Go: `pages_command_test.go` (matrix rows) and `wasm/pages_test.go` (preview, and single undo for a cross-page move and a paged create).
  - Designer: `component-command.test.ts`, `canvas-selection.test.ts`, `use-canvas-selection.test.ts`, and `App.test.tsx`, where the story-2 later-page drop row was rewritten.
  - E2E: `e2e/pages.spec.ts`.

## Spec Change Log

## Review Triage Log

| # | Source | Finding | Verdict | Route | Evidence |
|---|--------|---------|---------|-------|----------|
| 1 | verification-gap, blind | No test places on a later page with a pointer release; only Enter is covered | medium | patch | Pre-verified gap. Only the `threeWindows` one-page test (App.test.tsx:6262) releases a pointer, and it expects no `page` field. |
| 2 | blind | Wasm test calls `GroupMovePreview` without checking its result | low | patch | Direct test correction. The preview's dx/dy are never compared with the commit. |
| 3 | blind, edge-case | The grab offset falls back to 0 when a child receives the press, so the page is judged from the element's top | low | reject | Real, but only table heading and cell spans take pointer events (App.css, story 14.10), and they sit near the table's top, so the error is at most about a row. A fix needs a new non-measuring offset source. |
| 4 | blind | A cross-page preview can hang past the target band and is refused on release | low | reject | The frozen matrix row "Target outside band → Refused" settles this; the fix would edit the spec. |
| 5 | blind | Releasing over a header, footer or gap does a same-page clamped move | low | reject | It matches "otherwise as today". A release there is not over another page's content band. |
| 6 | blind | `contentPageAt` ignores x | low | reject | The x follows the pointer unclamped in a cross-page move, so a release beside the sheet is refused with a named error, not silently re-paged. |
| 7 | blind | Moved elements are appended to the end of the target page, on top of the stack | low | reject | Consistent with create, which appends. No ordering promise exists to break. |
| 8 | blind | The one-page byte promise has no new Go test; the task list names untouched test files | false | reject | The existing component_commands and group_movement tests send commands without `page` and still pass unchanged, which is the promise. |
| 9 | blind | The optional `page` count is duplicated three times | low | reject | Cosmetic; no drift exists today. |
| 10 | blind | Placement on later sheets uses `createComponent` at a fixed 72×24 instead of a clamped `dropComponent` | low | defer | Pre-existing for page 1's continuation sheets; story 3 only extends the same path. |
| 11 | blind | Changing an element's page is pointer-only, with no keyboard path or announcement | medium | defer | Not in the intent (CAP-3 names dragging); needs an owner decision. |
| 12 | blind | E2E doesn't assert the moved element's y or re-check the file after redo | low | reject | The Playwright run is deferred and pixel drags snap; the unit and wasm tests pin the bytes. |

## Verification

**Commands:**
- `cd folio-go && go build ./... && go vet ./... && test -z "$(gofmt -l .)"` -- expected: clean
- `cd folio-go && go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` -- expected: pass
- `cd folio-go && go build -tags=matrix ./... && go vet -tags=matrix ./...` -- expected: clean (matrix hash run deferred to end of run)
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` -- expected: pass (Playwright deferred to end of run)
- `cd lint && go test ./...` -- expected: pass

**Manual checks (if no CLI):**
- `npm run dev`: add a page, place Text on page 2, drag it to page 1 and back with the preview following the pointer, undo, and save.
