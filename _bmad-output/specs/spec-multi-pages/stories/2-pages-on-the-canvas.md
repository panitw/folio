---
title: 'Pages on the canvas'
type: 'feature'
created: '2026-09-14'
status: 'done'
route: 'dispatch'
baseline_commit: '0abf685093119097a9bb09a01a800523d3746244'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-multi-pages/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/multi-pages-decision-log.md'
  - '{project-root}/_bmad-output/specs/spec-multi-pages/stories/1-pages-in-the-engine-and-file-format.md'
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the frozen Intent below governs implementation.*

**The problem.** Story 1 taught the engine about pages, but the designer can't show or manage them. Opening a multi-page file draws page 2's content on top of page 1's sheets, and there is no way to add or remove a page.

**What this story builds.**
- **On the canvas:** each designed page is drawn as its own group of sheets, with a "Page N" label beside it. Clicking a page's empty space or its label selects that page, and a selected page is outlined.
- **In the canvas toolbar:** two icon buttons, **Add page** and **Delete page**, alongside the other canvas tools. Add page puts an empty page after the selected page (or at the end) and selects it. Delete page asks "Delete page N?" before removing the page and everything on it. Each is one undo step.
- **In Page Setup:** with a page selected, the panel gains a section for that page with a **Page Break** checkbox. It is disabled on page 1, with the reason shown.

**What it does not do.**
- You still can't drop new elements onto page 2 or later, or drag elements between pages (story 3).
- Editing the header from any page is story 4.
- Page Break off is saved but still prints as on until story 5.

**What done looks like.** Starting from a blank document, an author adds two pages, deletes the middle one after confirming, undoes that, and unticks Page Break on page 2. Every step shows correctly on the canvas and saves.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The designer draws a multi-page projection as one column, homing later pages' components onto page 1's sheets. It offers no way to select, add or delete a page, or to set a page's Page Break.

**Approach:**
- Add engine commands `addPage`, `deletePage` and `setPageBreak`, each one undo entry.
- Project each content component's designed page.
- Group the canvas sheet stack by page, with a selectable page label.
- Add Add page and Delete page (with an in-app confirmation) to the canvas controls toolbar.
- Add a selected-page section with a Page Break checkbox to Page Setup.

## Boundaries & Constraints

**Always:**
- **Toolbar** (owner decision D-2.3). Add page and Delete page sit in `canvas-tools` (never the document bar) as icon-only glyph buttons in the existing `tool-button` pattern.
  - Each has a new `ToolGlyph`, an `aria-label` ("Add page", "Delete page") and a `data-tip` tooltip.
  - The owner overrides control-vocabulary V2 here, so the pinned icon-only control list (R4) grows by these two, deliberately.
  - A disabled page button states its reason in its tooltip and accessible description, never a bare grey-out.
  - Existing toolbar controls keep their labels and order.
- **Add page.** It inserts an empty page with Page Break on after the selected page, or at the end when no page is selected. The new page becomes selected.
- **Delete page.** It opens an in-app `role="dialog"` confirmation reading "Delete page N?", with Delete page and Cancel.
  - Confirming sends one `deletePage` that removes the page and all its elements.
  - Cancel or Escape changes nothing.
  - With one page, the button is disabled with its reason, and the engine refuses the command.
- **Which page Delete page targets** (owner decision).
  - It targets the selected page. With no page selected, it targets the page holding the selected content elements, when all of them are on one page.
  - It is disabled with its reason when nothing is selected, when only header or footer elements or the section break are selected, or when the selected elements span pages.
- **The Delete and Backspace keys never delete a page** (owner decision). With only a page selected they do nothing, and element and section-break deletion is unchanged.
- **Shape changes (D-G.1).**
  - Going from one page to two moves `bands.content` elements and its section break into pages[0].
  - Going from two pages to one moves pages[0] back into `bands.content`, keeping its section break.
  - A deleted page's section break goes with it.
  - The saved bytes always follow the story-1 shape rules.
- **Page Break checkbox.** `setPageBreak` applies to page index ≥ 1 and is refused on page 1. The checkbox shows the selected page's value and is disabled on page 1 with its reason.
- **Page selection.** Selecting a page is designer state only, never saved. It is cleared by selecting an element or section break, by Escape, and by undo or redo removing that page.
- **Canvas drawing.**
  - Each content component is homed and echoed only on its own page's sheets.
  - The section break line draws only on page 1's sheets.
  - Rectangle selection tests components against their own page's windows.
- **One-page documents** keep today's canvas geometry, component and band accessible names, drop behaviour and seams. The only addition is the page label.
- **Engine rules still hold.** The engine is the only layout authority, and the DOM-measurement ban holds. `MAX_CANVAS_SHEETS` truncates the tail across all pages.

**Never:**
- No drop or create onto a page after page 1's sheets, and no drag between pages (story 3).
- No header or footer editing changes (story 4).
- No Page Break off pagination (story 5).
- No page reorder or duplicate.
- No `window.confirm`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Add to one-page | One-page doc with elements and a break; Add page | Two pages; old content and break in pages[0]; pages[1] empty with `pageBreak` true and selected; one undo restores the original bytes | N/A |
| Add after selected | 3 pages, page 2 selected; Add page | New empty page at position 3, selected; old page 3 becomes page 4 | N/A |
| Add with none selected | 2 pages, nothing selected | New page appended as page 3 | N/A |
| Delete confirmed | 3 pages, page 2 with elements selected; Delete page, confirm | Dialog reads "Delete page 2?"; page and its elements removed; one undo restores them | N/A |
| Delete cancelled | Same; Cancel or Escape | Nothing changes, no command sent | N/A |
| Delete to one page | 2 pages, break on page 1; delete page 2 | File returns to the one-page shape with the break in `bands.content` | N/A |
| Delete page 1 | 2 pages; delete page 1 | Old page 2 becomes the single page | N/A |
| Last page | 1 page | Delete page disabled with its reason | Engine `deletePage` refused |
| Page Break off | Page 2 selected; clear Page Break | Saved `"pageBreak": false`; one undo; checkbox reflects the value | N/A |
| Page 1 Page Break | Page 1 selected | Checkbox disabled with its reason | Engine `setPageBreak` on page 1 refused |
| Multi-page drawing | Multi-page fixture opened | Page 2's components are drawn only on page 2's sheets; labels "Page 1", "Page 2"; break line only on page 1 | N/A |
| Later-page drop | Palette item over page 2's content band | Not accepted as a drop target | N/A |
| Delete via element | 3 pages, one element on page 3 selected, no page selected; Delete page | Dialog reads "Delete page 3?"; confirming removes page 3 | N/A |
| Delete disabled | Nothing selected, a header element only, or elements on pages 1 and 2 | Delete page disabled with its reason | N/A |
| Delete key on a page | Page 2 selected; press Delete or Backspace | Nothing happens; no dialog, no command | N/A |

</frozen-after-approval>

## Code Map

- `folio-go/component_commands.go` -- `ApplyComponentCommand` switch (:218-328): add `addPage`, `deletePage` and `setPageBreak` cases next to `setSectionBreakAnchor` (:309). Helpers: `componentFailure` :33, `componentFields` :1789, `commandBool` :1808.
- `folio-go/section_break.go` -- `setSectionBreakAnchor` (:304-332) is the command precedent: field count, explicit null refusal, save, mutate, then `Canvas(t)` with rollback on error. `sectionBreakPath` is at :44.
- `folio-go/internal/template/model.go` -- `Document.Pages`, `ContentPage{Band, PageBreak}`, `PageCount`, `PageField` (:446). Invariant: `Pages` is nil for one page and ≥2 entries otherwise (`parse_pages.go:42`), and `Bands.Content` is empty beside `Pages`.
- `folio-go/content_pages.go` -- `contentPageIndex`, `firstContentBand`. The new page commands live beside these, in a new `pages_command.go`.
- `folio-go/wasm/engine.go` -- `Apply` (:298-354) clones the document and pushes one undo entry when the bytes change. Nothing to add per command.
- `folio-go/page_setup.go` -- `CanvasComponent` (:257-330), `canvasComponents` (:1935), and the `BelowSectionBreak` post-pass using `contentPageIndex` (:988-998). Add an always-present component `page` there. Mirror it in `canvas_projection_wire_test.go` (component keys ~:680-725, `canvasTableComponentEmittedKeys` :743) and in `folio-designer/src/engine-protocol.ts` (type :492, `hasOnly` and checks :846/:850).
- `folio-designer/src/section-break-command.ts` -- `commandBytes` precedent for the new `page-command.ts`. Test precedent: `section-break-command.test.ts:27`.
- `folio-designer/src/sheet-stack.ts`:
  - `homeWindow` :80 and `sheetStack` :105-135 (`homes` :114, `intersects` :126, `seam` :133) assume one column. Apply them per page, using `contentWindowPages` and the component `page`.
  - A seam never crosses a page boundary.
  - `stackYForColumn` :159, `columnForStackY` :171 and `columnEdgeAfterDrag` :182 stay within the component's page.
  - `MAX_CANVAS_SHEETS` :33 must not be computed as `contentWindowHeight * index` (`canvas-authority-contract.test.ts:68`).
- `folio-designer/src/section-break.ts:38` -- `sectionBreakPlacement` searches page 0's windows only.
- `folio-designer/src/canvas-selection.ts:11-32` -- `enclosedComponents` uses the component's page windows.
- `folio-designer/src/App.tsx`:
  - **Selection:** state at :392 and the section-break flag at :372/:607. `installSelection` (:1206) clears a new page selection. `selectSectionBreak` (:1827) is the pattern for selecting a page.
  - **Clicking a page:** `sheetSurface` (:3035-3045) runs `installSelection([])` in the `page-surface` onClick; change it to select that sheet's page. The backdrop keeps having no handler (:3257).
  - **Drops:** `dropOnPage` (:3055) and `placeInBand` (:1522, later sheets) must not target a later page's content band.
  - **Sheets:** stack built at :3034 and mapped at :3277; the label renders at a page's first sheet. The header/footer `home` rule stays at :3062.
  - **Toolbar and panels:** toolbar at :3276, button pattern `tool-button`, `toolTip`. Right-panel ternary at :3323 and `PageSetup` at :3535; the page section is a new `section-label` block. Anchor checkbox precedent at :5367.
  - **Commands and keys:** `commitComponent` (:1187) and the `mutationInFlight` guard as in `deleteSectionBreak` (:1833). Keyboard routing at :1665 and :2914 (`keyboardDelete`, `canvasKeyAllowed`).
- `folio-designer/src/FontBrowser.tsx:276` -- the `role="dialog" aria-modal` focus-trap pattern for the confirmation.
- Tests pinning current UI:
  - `App.test.tsx:1203-1220`: the ordered toolbar button list (`aria-label`, `data-tip`, an `svg.tool-icon` and no text). The two page buttons join it as glyph buttons.
  - `folio-designer/src/toolbar-icons.tsx`: the closed `ToolGlyph` union; add the page glyphs on its 16px grid.
  - `App.test.tsx` `canvas sheet stack` (:6153-6382).
  - `control-vocabulary-contract.test.tsx` R1-R4 and `V2_CENSUS` (:433).
  - `sheet-stack.test.ts`, `section-break.test.ts`, `canvas-selection.test.ts`.
- E2E precedent: `folio-designer/e2e/section-break.spec.ts`.
- oxlint `only-export-components` baseline is 4, so helpers go in `.ts`.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/pages_command.go`, `folio-go/component_commands.go` -- `addPage{after?}`, `deletePage{page}`, `setPageBreak{page, pageBreak}` with shape normalization, refusals located at `pages[i]`, and rollback on a `Canvas` error -- engine commands
- [x] `folio-go/page_setup.go`, `canvas_projection_wire_test.go`, `folio-designer/src/engine-protocol.ts` -- component `page` field and its validation (0 for header/footer and one-page docs; must index an existing page) -- per-page homing
- [x] `folio-go/pages_command_test.go`, `folio-go/wasm/pages_test.go` -- every engine matrix row, including byte-identical undo and the shape round trips -- engine coverage
- [x] `folio-designer/src/page-command.ts` (+ test) -- command builders -- one JSON authority
- [x] `folio-designer/src/sheet-stack.ts`, `section-break.ts`, `canvas-selection.ts` (+ tests) -- per-page homing, echoes, seams, column mapping, break placement and rectangle selection -- correct drawing
- [x] `folio-designer/src/App.tsx`, `App.css`, `toolbar-icons.tsx` -- page labels, page selection and outline, Add page and Delete page glyph buttons with new glyphs and tooltip disabled reasons, confirmation dialog, Page Setup page section, later-page drop exclusion -- UI
- [x] `folio-designer/src/App.test.tsx`, `control-vocabulary-contract.test.tsx` -- UI matrix rows, the updated toolbar pin, the R4 icon-only list and `V2_CENSUS` extended by the two page buttons (D-2.3), and one-page names unchanged -- UI coverage
- [x] `folio-designer/e2e/pages.spec.ts` -- add, delete with confirm, undo, Page Break toggle and save (compiled; run deferred) -- e2e

**Acceptance Criteria:**
- Given any one-page document, when it opens, then every existing component, band and sheet accessible name and the sheet geometry match today's, and one "Page 1" label is added.
- Given a page is selected, when the author presses Escape or clicks an element, then the page selection clears and the panel follows the new selection.
- Given undo removes the selected page, when the snapshot updates, then no page is selected.

## Implementation Notes

**How it was built.**
- **Engine** (`folio-go/pages_command.go`):
  - `addPage`, `deletePage` and `setPageBreak` take 0-based page indexes on the wire, matching `pages[i]`.
  - `contentPagesForEdit` and `installContentPages` move content between the one-page and `pages` shapes whenever the page count crosses one.
  - A failed `Canvas` restores the saved shape.
  - Refusals are located at `pages`, `pages[i].pageBreak` or `bands.content`.
- **Projection:** `CanvasComponent.Page` is always sent. `CanvasProjection.PageBreaks` has one entry per page, and page 1's is always true. The Code Map did not list `PageBreaks`; it was added so the checkbox shows the engine's own value.
- **Designer drawing** (`sheet-stack.ts`):
  - `Sheet` gains `page` and `pageStart`, and `pageWindows` and `componentPage` home each component among its own page's windows.
  - Seams stop at page boundaries.
  - `stackYForColumn`, `columnForStackY` and `columnEdgeAfterDrag` take the component's page.
  - `sectionBreakPlacement` and `enclosedComponents` filter by page.
- **Designer UI** (`App.tsx`):
  - `selectedPage` state is cleared by `installSelection`, a document replace, undo and redo.
  - `deletePageTarget` works out D-2.1.
  - `DeletePageDialog` is an in-app `role="dialog"` that focuses Cancel, traps Tab and cancels on Escape.
  - `PageSection` sits below Page Setup.
  - A content band on a page after page 1 ignores placement (`accepts`).
- **Protocol guard:** it accepts a projection without `pageBreaks` or component `page` only when it has one page. Go always sends both.

**Carried forward.**
- **Story 3:** drops and creates onto later pages, and drags between pages.
- **Story 4:** header and footer copies are still interactive only on sheet 0.

## Spec Change Log

## Review Triage Log

Pass 1, 2026-09-14. Layers: blind-hunter (BH), edge-case-hunter (ECH), verification-gap (VG). OWNER is a defect the owner reported from the running designer during review.

| # | Finding | Verdict | Evidence | Route |
|---|---------|---------|----------|-------|
| BH1 | The designer's projection guard rejects a multi-page projection that carries a page-1 section break | high | Confirmed. `engine-protocol.ts:859` requires `belowSectionBreak` on every content component whenever `sectionBreak` is set, but `page_setup.go:1017` sends it only for page-0 content. The matrix row "Add to one-page" (a document with a break) produces exactly that projection. The App test used `initialSnapshot`, which skips the guard. | patch |
| OWNER | The "Page N" label is hidden under the PAGE HEADER band tab | medium | Confirmed. `.page-label` sits in the left gutter (`right: calc(100% + …)`), where the band tabs are already drawn (DESIGN.md band tabs outside the page's left edge). | patch |
| BH5 / ECH1 / ECH-claim2 | Delete page disabled by `fileBusy` alone is a bare grey-out | low | `disabled={'reason' in deletePageTarget \|\| fileBusy}`, but `data-tip` and `aria-describedby` only follow `deletePageTarget`. This breaks the frozen Always rule. | patch |
| BH6 / ECH7 | Focus is lost after a confirmed delete, and the dialog description is not linked | medium | `confirmDeletePage` does not restore focus, and the Delete page button is usually disabled afterwards, so focus falls to `<body>`. | patch |
| ECH5 | A backdrop press drops focus out of the dialog, and then Escape and Tab stop working | low | The keydown handler is capture-phase on the `section`, and `<body>` is outside it. The fix is a small refocus. | patch |
| BH9 | The Delete-key test cannot fail | medium | With only a page selected, `selected` is empty, so the key had nothing to delete before this change either. D-2.2's real case, an element on page 2 selected, is untested. | patch |
| VG1 | Resizing a later-page component through App is untested | medium | Pre-verified by VG. Dropping `componentPage(...)` at `App.tsx:3118` breaks no test. | patch |
| BH11a | `addPage` with `after` equal to the last index is untested, and the wasm test comment is misleading | low | The test and the comment are direct corrections. | patch |
| BH2 | A section break on a later page can't be seen or edited | false | The story-1 parser refuses `sectionBreak` on `pages[1..]`. `addPage` inserts at index ≥ 1 and `deletePage` removes a page's break with it, so no command moves a break off `pages[0]`. | reject |
| BH3 / ECH6 | Pages cut off by `MAX_CANVAS_SHEETS` can't be selected, and fall back to the wrong mapping | low | Reachable only past 120 sheets, and the frozen rule says the cap truncates the tail. The fix would add branches. | reject |
| BH4 | Accessible names mix sheet numbers ("Content on page 3 of 3") with designed page numbers ("Page 2") | medium | Real for multi-page documents with overflow. Renaming would change the multi-sheet names that the one-page constraint and existing tests pin. | defer |
| BH7 / ECH2 / ECH3 | Add page, or a confirmed delete, silently does nothing while another mutation is in flight | low | Only inside the short in-flight window, and the canvas visibly stays unchanged. The fix would add guards. | reject |
| BH8 | Later-page drops are ignored with no feedback, and the click selects the page | low | This is the deliberate temporary state until story 3, which replaces this code. | defer |
| BH10 | Add page ignores the selected element while Delete page uses it | false | The frozen rule says Add page inserts after the selected page, and D-2.1 applies only to Delete page. Changing it would edit this spec. | reject |
| BH11b | `savePageShape` makes a shallow copy, and the rollback is untested | false | The commands build new slices and never change `Elements` in place. Rollback runs only on a `Canvas` failure, which the matrix doesn't reach (low). | reject |
| BH12 | e2e not run, the Spec Change Log is empty, and the Code Map line numbers are stale | false | The e2e run is deferred by the owner's cadence (D-S.4). The Spec Change Log records review loopbacks only. Stale line anchors are not a defect. | reject |
| ECH4 | Page buttons are clickable while the table editor is open | false | `.table-editor-backdrop` is `position: fixed; inset: 0; z-index: 20` (`App.css:852`), covering the toolbar, and it traps keys. | reject |
| ECH-claim1 | A one-page click now outlines the page and shows a PAGE 1 section | false | The frozen rule says clicking a page's empty space selects it. AC1 constrains names and geometry, which are unchanged. | reject |
| ECH-claim3 | Refusals are located at `pages`, not `pages[i]` | low | An out-of-range index has no `pages[i]` to name. Page 1's Page Break refusal is at `pages[0].pageBreak`. | reject |

## Verification

**Commands:**
- `cd folio-go && go build ./... && go vet ./... && test -z "$(gofmt -l .)"` -- expected: clean
- `cd folio-go && go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` -- expected: pass
- `cd folio-go && go build -tags=matrix ./... && go vet -tags=matrix ./...` -- expected: clean (matrix hash run deferred to end of run)
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` -- expected: pass (Playwright deferred to end of run)
- `cd lint && go test ./...` -- expected: pass

**Manual checks (if no CLI):**
- `npm run dev` with the multi-page fixture: page labels show, page 2's content is drawn on page 2 only, and add, delete (with confirm) and undo behave as in the matrix.
