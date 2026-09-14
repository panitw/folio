# Existing integration points

Inspected on 2026-09-14. Paths are relative to the repository root, and line numbers may have moved since. Each row lists a starting point and the rule it must not break. None of them widens the scope.

## Engine

| Source | Current behaviour | Implication |
|---|---|---|
| `folio-go/internal/template/model.go` (`Document`, `Page`, `Bands`, `Band`, `Element`); `parse.go`, `parse_bands.go` | A document has one `Page` setup and three bands: `pageHeader`, `content`, `pageFooter`. Content-band element `X`/`Y` are relative to the band and describe one column of unbounded height. | Designed pages each own a content band's elements (and an optional section break). The header, the footer and the page setup stay single and at document level. A document with one page must parse and save exactly as it does today. |
| `folio-go/internal/layout/paginate.go`, `Paginate` | Paginates one content column into windows. Each window starts at the first item that did not fit. | Paginate each designed page's column separately, then join the results in page order. Page Break on starts a fresh output page. Page Break off shares the previous page's last output page when the block fits. |
| `folio-go/section_break.go`, unanchored path (`aboveLineEndsAtOrAbove` ~387 and the shared-page merge) | An unanchored section moves as a whole block directly below where the content above it ends, when it fits. Otherwise it starts at the top of a new page. | This is the precedent for Page Break off. Reuse its fit test and shared-page merge rather than inventing a second one. |
| `folio-go/render.go`, `paginateDocument`, `contentColumnItems`; header and footer repeat (~137); `headerFooterResolver` (~613, ~2487) | Builds the column and draws the header and footer on every output page. Page tokens are resolved after the page count is known. | The header and footer stay one definition drawn on every output page of every designed page. The page count is the total across all designed pages. |
| `folio-go/section_break.go` | One break per document, handled in `paginateWithSectionBreak`. | The break becomes per designed page. The section-break rules still apply, within that page's column only. |
| `folio-go/page_number.go` | Computes the total for `Page X of Y`. | It must sum across designed pages. |
| `folio-go/component_commands.go` (`moveComponent`, `updateComponentProperties`) | Commands address elements by band. | Commands must also address the target page. A move onto another page is one command. Group moves that would span pages are refused. |
| `folio-go/wasm/engine.go` | Undo and redo are snapshots of the whole document, up to 100, with one command producing one entry. | Add Page and Delete Page are engine commands. Deleting a page with its elements is one entry. |
| `folio-go/page_setup.go` (~433, `CanvasProjection`, `ContentWindowOrigins` ~556); `folio-designer/src/engine-protocol.ts:415` | Projects window origins for the single column. | Project windows per designed page, so the designer can tell which sheets belong to which page. The designer derives none of this itself. |

## Format

- `_bmad-output/specs/spec-folio/folio-format.md`, *Bands* (~331) and *Pagination* (~353): document how pages are represented, how a designed page starts a new output page, and that the header and footer are shared. Qualify "a single column" as "a single column per page".
- *Document* `version` derivation (~47): only documents with more than one page declare the new version. The version number is an open question in SPEC.md.
- Section break (Pagination, *Bands*): the break becomes a per-page key.

## Designer

- `folio-designer/src/App.tsx`, `<div className="canvas-tools" aria-label="Canvas controls">` (~3276): the toolbar above the canvas. It already holds zoom, grid, snap, duplicate and delete. Add Page and Delete Page go here, not in `<header className="document-bar">` (~3168).
- `folio-designer/src/sheet-stack.ts`: builds sheets from `contentWindowOrigins`, draws at most `MAX_CANVAS_SHEETS` = 120, and sets `home` so that only one occurrence of each component is interactive. Sheets must be grouped by designed page. Keep the 120-sheet cap across the whole stack.
- `App.tsx` `sheetSurface` (~3035): draws the header and footer on every sheet, but only one copy is clickable. For CAP-4, every page's copy must be selectable and must edit the same definition. The single-accessible-name rule (Ruling G) still applies, so the header and footer need an answer for which copy owns the accessible name.
- `App.tsx` selection (~392 element ids, ~607 section-break flag): clicking empty page space clears the selection and shows Page Setup. Page selection is new state beside these two. Page Setup (~3536) gains a section for the selected page containing the Page Break checkbox; `SectionBreakProperties`' Anchor checkbox is the command precedent for it.
- `folio-designer/src/section-break.ts`, `section-break-command.ts`, `SectionBreakProperties`: currently scoped to the document. Scope them to the selected page.
- `folio-designer/src/canvas-authority-contract.test.ts`: the ban on measuring the DOM applies to all new geometry.

## Verification locations

- Go tests beside `folio-go/table_pagination_test.go` for page order with overflow, page count across pages, and a per-page section break.
- A golden multi-page fixture under `fixtures/`, registered in `hashmatrix/`. The existing fixtures keep their hashes.
- A round-trip test and a load-error test for each new diagnostic.
- Designer e2e tests for add page, select page, delete page and undo, moving an element across pages, and editing the header from page 2.
