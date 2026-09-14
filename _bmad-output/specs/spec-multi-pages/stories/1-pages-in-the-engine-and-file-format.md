---
title: 'Pages in the engine and file format'
type: 'feature'
created: '2026-09-14'
status: 'done'
route: 'dispatch'
baseline_commit: '01a39d9f2d0ee23ad79c941ea015c80b95f29dba'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-multi-pages/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/multi-pages-decision-log.md'
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the frozen Intent below governs implementation.*

**The problem.** A Folio document has one content area. It can't yet say "these clauses are page 1 and these are page 2".

**What this story builds.** The engine and the file format learn about designed pages. A file can list several pages. Each page is its own block of content, and each renders starting on a fresh printed page after the previous page and anything that overflowed from it. The header and footer appear on every printed page, and "Page X of Y" counts all of them. The engine also tells the designer which canvas sheet belongs to which page.

**What it does not do.** There are no Add Page or Delete Page buttons and no page labels on the canvas (story 2). You can't place or move elements onto a chosen page (story 3), and you can't edit the header from any page (story 4). Section breaks on page 2 or later, and the "Page Break off" behaviour, are story 5. The Page Break value is saved and loaded now, but every page renders as if it were on.

**Why it's shaped this way.** One-page files must not change by a single byte. So the new `pages` list appears only in files with two or more pages.

**What done looks like.** A hand-written two-page file renders with page 2 starting after page 1's overflow, all pages numbered correctly, and saves back byte-for-byte.

**What will look wrong and isn't.** Opening a multi-page file in the designer won't crash, but it won't show pages properly until story 2.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The `.folio` model has one content column (`bands.content`), so a document cannot declare designed pages, and pagination, page count, and the canvas projection all assume a single column.

**Approach:**
- Add a top-level `pages` array. Each entry holds `elements`, an optional `sectionBreak`/`sectionBreakAnchor`, and `pageBreak`.
- Route every content-element reader through one accessor over all pages.
- Paginate each page's column separately and concatenate the output pages.
- Project a designed-page index for every canvas window.

## Boundaries & Constraints

**Always:**
- **File shape (D-G.1).** A file with two or more pages writes every page, page 1 included, in `pages`, and `bands.content` has no elements. A one-page document keeps today's shape and bytes, declared version, PDF hash and canvas projection (plus the new field only).
- **Save shape.** The saver chooses the shape by page count. The `pages` shape requires version `4.1`, with no new version rank.
- **Page Break saving.**
  - `pageBreak` is written explicitly as `true` or `false` on pages[1..].
  - When it is missing on pages[1..], it loads as `true`.
  - A `pageBreak` on pages[0] loads without error and is dropped on save.
  - In this story every page renders as if `pageBreak` is `true`.
- **Page order.** Each page starts a new output page after the previous page's last output page. An empty page yields one output page with only the header and footer.
- **Header, footer and page count.** The header and footer are drawn on every output page. `{{page}}`/`Page X of Y` totals are summed across pages.
- **Ids and keep-together.** Element ids are unique document-wide, using the one shared `parseCtx`. `keepTogether` is valid on page elements. A tag used on two pages is a load error naming both pages.
- **Existing machinery stays.**
  - `sectionBreak` on pages[0] keeps its section-break rules within that page's column.
  - Glyph subsetting, asset dedup, visibility, and PDF object reservation stay once per document.
  - Arithmetic stays integer fixed-point.
- **Diagnostics.** Every new code is registered, and every load error names its location as `pages[i]` (with key).
- **Commands by id.** Existing engine commands addressing an element by id work on multi-page documents: find, update, move within its page, delete. Commands that create or drop into `content` target pages[0]. Asset and font reference walks see every page.

**Never:**
- No Add/Delete Page commands and no per-page drop targeting (stories 2–3).
- No designer drawing of pages beyond keeping the projection validator accepting multi-page projections.
- No Page Break off pagination, and no `sectionBreak` on pages[1..], which is refused as `SECTION_BREAK_INVALID` naming the page (story 5).
- Never add a fourth band or change the closed three-band set.
- Never silently turn a one-entry `pages` into the one-page shape.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| One-page unchanged | Every existing fixture | Same bytes on save, same version, same PDF hash | N/A |
| Two pages, page 1 overflows | Page 1 `transactions[]` spans 3 output pages; page 2 static | 4 PDF pages; page 2 content on output page 4 at declared y; `Page 1–4 of 4`; header and footer on all | N/A |
| Empty page | `pages[1].elements` is `[]` | One header- and footer-only output page | N/A |
| Round-trip | Multi-page file with `pageBreak` true and false | Byte-identical save | N/A |
| Missing pageBreak | pages[1] has no `pageBreak` | Loads as true; save writes `"pageBreak": true` | N/A |
| Page 1 pageBreak | pages[0] has `"pageBreak": false` | Loads; ignored; dropped on save | N/A |
| Both places | `pages` present and `bands.content.elements` non-empty | Refused | `PAGES_INVALID` naming `bands.content` |
| Short pages | `pages` has 0 or 1 entries | Refused | `PAGES_INVALID` naming `pages` |
| Bad entry | Unknown key in an entry, or non-boolean `pageBreak` | Refused | `PAGES_INVALID` naming `pages[i]` and the key |
| Dup id | Same id on pages[0] and pages[1] | Refused | Existing duplicate-id error, located at `pages[1]` |
| Tag spans pages | Same `keepTogether` tag on two pages | Refused | `PAGES_INVALID` naming both pages |
| Later-page break | `sectionBreak` on pages[1] | Refused | `SECTION_BREAK_INVALID` naming `pages[1]` |

</frozen-after-approval>

## Code Map

All engine paths are relative to `folio-go/`.

- `internal/template/model.go` -- `Document` (:19, `Extra` :79), `Bands` (:387, closed), `Band` (:396, SectionBreak :404/:410). Add `Pages` (a slice of page structs) to the Document.
- `internal/template/parse.go` -- `ParseDocument` (:53) consumes known keys. An unconsumed `pages` falls into `Extra` today, so consume it explicitly. `parseCtx.claimID` (:35) is shared for ids.
- `internal/template/parse_bands.go` -- `decodeBand` (:61) and the sectionBreak decode (:110-155). Extract that decode for reuse. The keepTogether gate at :348-365 is `bandField == "bands.content"`; widen it to page fields.
- `internal/template/serialize.go` -- `writeObject` sorts keys, so `pages` lands between `page` and `unbreakableValues`. Model `writeBand` (:324) conditional writes on it.
- `internal/template/version.go` -- `versionRequiredByContent` (:329). The sectionBreak check (:360) and element walk (:363) only see `Bands.Content`; add pages.
- `internal/diag/diag.go` (:411/:457/:503), `diagnostic.go` (:338) -- register `PAGES_INVALID` with its public bridge. Tests: `diagnostic_registry_census_test.go`, `diag_bridge_test.go`, `internal/diag/diag_test.go`.
- Content readers to reroute through one accessor returning every page's element list plus its page index:
  - `page_setup.go` (:1083, 1169, 1256, 1373, 1501, 1889, 2101)
  - `component_commands.go` (`findComponent` :2159 searches bands; `bandByName` :1847; `hitTestBand` :2137; :1249, :4729)
  - `render.go` (:325, :2864 `keepTogetherTags`)
  - `barcode_element.go` (:263, :338)
  - `group_movement.go:40`, `folio_expr_validate.go:24`, `parameter_references.go:34`, `stand_in_data.go:404`, `serialize_template.go:53`, `table_min_height.go:59`
- `render.go` -- `predictDocument` (:2357) computes `pageCount` from one plan (:2480-2485). `paginateDocument` (:2979) paginates (:3191) and assembles pages (:3226-3344). Build per-page items with document-wide refs, call `paginateWithSectionBreak` per page, and concatenate. Offset `Suppressed`/`Clipped` page numbers as `section_break.go:471-477` does. Keep one `sectionPlan` per page for `shiftFor`. Keep `buildShapedPDFRuns`, assets and visibility global.
- `page_number.go:87` `contentColumnItems` -- filters on the content band index; give it a page filter.
- `section_break.go` -- `sectionBreakOf`/`declaredSectionBreak` (:56-94) read `Bands.Content`, and the diag paths are fixed at :37-40. Make them page-aware for pages[0] and refuse pages[1..]. Keep the one-page fast path at :400-405 intact.
- `page_setup.go` -- `CanvasProjection` (:433-639), `addCanvasWindowCount` (:1230). Add an always-present `ContentWindowPages []int`, one per window. Origins become page-local, restarting at 0 at each page start.
- `folio-designer/src/engine-protocol.ts` -- `isCanvas` (:752) uses a `hasOnly` key list, so add the key, a typed clause, and the relaxed origins check (strictly rising within a page, 0 at each page start). Mirror the key in `canvas_projection_wire_test.go` (:47-79, :383).
- Fixture precedent is commit `3a645cf`: `fixtures/<slug>/`, `matrix_test.go` `matrixDocuments`, `.github/workflows/matrix.yml` (4 upload paths plus the `docs=` list, guarded by `matrix_registration_test.go`), `byte_neutrality_test.go` `goldenDigestRecord`, `render_test.go` subprocess hook, and `*_template.go`/`*_fixture_test.go`. Also `internal/template/fixtures_test.go` `maximalFixture` and `drift_test.go`, which require new keys to be documented.
- `_bmad-output/specs/spec-folio/folio-format.md` -- version rule (:47), Bands (:331-348), Pagination (:353-408; qualify "No page is ever blank" at :408).

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/internal/template/{model,parse,parse_bands,serialize,version}.go` -- add the pages model, parse/validate (all matrix refusals), serialize by page count, and version rank -- the format contract
- [x] `folio-go/internal/diag/diag.go`, `folio-go/diagnostic.go` -- register `PAGES_INVALID` and its bridge -- registry tests
- [x] `folio-go/*.go` content readers in the Code Map -- reroute through one all-pages accessor; `findComponent` searches every page -- commands and reference walks keep working
- [x] `folio-go/render.go`, `page_number.go`, `section_break.go` -- per-page pagination and concatenation, summed page count, break refused on pages[1..] -- CAP-5
- [x] `folio-go/page_setup.go`, `folio-designer/src/engine-protocol.ts`, wire test -- add `contentWindowPages` and page-local origins -- engine side of CAP-8
- [x] `fixtures/multi-page-statement/` plus matrix, CI and byte-neutrality registration -- two-page golden fixture for the CAP-5 row
- [x] `folio-go/*_test.go`, `internal/template/*_test.go` -- a test per I/O matrix row, plus projection window pages for a multi-page doc -- coverage
- [x] `_bmad-output/specs/spec-folio/folio-format.md` -- document `pages`, `pageBreak`, the version rule, page-order pagination and the blank-page qualification -- drift tests

**Acceptance Criteria:**
- Given any existing fixture, when loaded, saved, rendered and projected, then bytes, hash and projection values are unchanged apart from the added `contentWindowPages` of all zeros.
- Given a multi-page doc, when an id-addressed update, move or delete command runs on a pages[1] element, then only that element changes and undo restores it.
- Given the designer loads a multi-page projection, when `isCanvas` validates it, then it passes.

## Implementation Notes

**How it was built.**
- **Model and saving:**
  - `Document.Pages []ContentPage` is nil for a one-page document.
  - `ContentBands()` and `ElementBands()` are the only ways to read content.
  - `parse_pages.go` decodes `pages`, and `writeContentPages` saves it.
  - The section-break key decode is now `decodeSectionBreakKeys`, shared by the content band and page entries.
- **Engine readers:**
  - They go through `content_pages.go` (`contentElements`, `contentPageIndex`, `firstContentBand`).
  - `findComponent` searches every page.
  - `bandByName` and `hitTestBand` return page 1.
- **Render:**
  - `documentBands` passes all pages' elements through the collectors as one content band, in page order.
  - `paginateContentPages` splits items by element id, paginates each page with `paginateWithSectionBreak`, and joins the results.
  - `outputShiftFor` maps an output page back to its page plan.
- **Canvas projection:** `ContentWindowPages` gives each window's page, and origins are page-local. Group movement filters candidate windows by the element's page.
- **Golden fixture:** `fixtures/multi-page-statement` is the golden, sha `4b0367f5…`, recorded on darwin/arm64. The four-target run is owed at the end of the run.

**Carried forward.**
- **Story 2:**
  - The projection does not yet say which page a component is on (only which page a window is on).
  - `bandsForSave` would write a single in-memory page in the one-page shape, so Delete Page must handle that case deliberately.
- **Story 5:** section-break commands (`setSectionBreak` and related) still act on page 1 and report `bands.content.*` paths.

## Spec Change Log

## Review Triage Log

Pass 1, 2026-09-14. Layers: blind-hunter (BH), edge-case-hunter (ECH), verification-gap (VG).

| # | Finding | Verdict | Evidence | Route |
|---|---------|---------|----------|-------|
| BH1 / ECH2 | `sheet-stack.ts` homes components across page-local origins, drawing multi-page docs wrongly | low | Real for multi-page projections only. The frozen Never excludes designer drawing of pages (story 2). | reject (intent-excluded) |
| BH2 | `sectionBreakPlacement` picks the wrong sheet on page-local origins | low | Same root as BH1. Designer drawing is excluded by the frozen Never. | reject (intent-excluded) |
| BH3 | `bandsForSave` collapses a one-entry in-memory `pages` | false | Load refuses a one-entry `pages`, so no loaded file reaches this. D-G.1 requires the one-page shape when a document is edited down to one page (story 2's Delete). | reject |
| BH4 | Drops and creates into content silently target page 1 | false | This is the spec's Always rule ("Commands that create or drop into `content` target pages[0]"). The fix would edit this spec. | reject |
| BH5 / ECH5 | `partition` sends an item with an unknown ElementID to page 0 | false | Every `ColumnItem` constructor sets `ElementID` from a content element (`page_number.go:91/128/150`, `render.go:3046/3143/3185`, `page_setup.go:1310/1720`). | reject |
| BH6 | `canvasWindowPage` returns 0 when the index is out of range | low | Unreachable: the projection builds the two slices together. The fix would add a guard. | reject |
| BH14 | Warnings from later pages are not renumbered | false | The merged diagnostics (`table_footer.go:328`, `section_break.go:367`) carry no page number. | reject |
| ECH1 / ECH6 | `refuseSectionBreakStraddle` applies page 1's break to later-page elements | medium | Confirmed at `section_break.go:191-201`. It uses `declaredSectionBreak` (page 1) for any content element, and it is reached from move, resize, update, duplicate and group move. | patch |
| BH10 / ECH3 | Section-break command failures name `bands.content.*` in a pages-shaped file | low | `setSectionBreak`, `removeSectionBreak`, `setSectionBreakAnchor` and `refuseSectionBreakBeyondContent` still use the fixed path constants. The fix is a direct substitution, and no multi-page command test exists. | patch |
| BH8 / ECH4 | Load errors name `sectionBreak` when only `sectionBreakAnchor` is present, and the format doc gives the location as `bands.content` | low | Visible in `parse_pages.go` `decodePages`, the `validateSectionBreak` later-page loop, and the Designed pages section of `folio-format.md`. | patch |
| BH9 | The format doc spells the page-key rows inconsistently (bare `pageBreak`) | low | Visible in the `folio-format.md` table. Direct correction. | patch |
| BH13 | The semantic-acceptance precondition is vacuous | low | `strings.Contains` on the constant passes whatever page 2 declares. | patch |
| BH15 | The "one page" load-error test builds its input through an unknown key | low | The `"unused": [` trick can pass for the wrong reason. Direct correction. | patch |
| BH7 | No render test uses `"pageBreak": false` | medium | Every render and golden uses `true`. The Always rule "renders as if true" has no test. | patch |
| VG1 / BH12 | Asset and font reference walks on later pages are untested | medium | Pre-verified by VG. No multi-page case in the command tests. | patch |
| VG2 / BH11 | The group-move page filter is untested | medium | Pre-verified by VG. `group_movement_test.go` has no multi-page case. | patch |
| VG3 | A page-1 section break in a multi-page doc is never rendered or projected | medium | Pre-verified by VG. The golden has no break, and the template test only parses. | patch |
| VG4 | Suppressed/Clipped renumbering on later pages is untested | low | Pre-verified by VG. It affects warning text only. VG filed it as defer, but the change caused it and the test is small. | patch |
| VG5 | The multi-page degrade path is untested | medium | Pre-verified by VG. A regression would blank the canvas. | patch |
| PATCH-FOUND | Renaming a font chain did not update elements on later pages | high | Surfaced by the VG1 test while the patch was being written. `fontChainBands` returned the concatenated copy from `contentElements`, so the rename wrote into a temporary slice. It now returns each band's own element slice. | patch (fixed with VG1) |
| VG6 | Rerouted read walks and `IsExact` are untested on later pages | medium | Pre-verified by VG. Parameter references, stand-in data, expression validation and `minHeight` have no multi-page cases. | patch |

## Verification

**Commands:**
- `cd folio-go && go build ./... && go vet ./... && test -z "$(gofmt -l .)"` -- expected: clean
- `cd folio-go && go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` -- expected: pass
- `cd folio-go && go build -tags=matrix ./... && go vet -tags=matrix ./...` -- expected: clean (matrix hash run deferred to end of run)
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` -- expected: pass (Playwright deferred to end of run)
- `cd lint && go test ./...` -- expected: pass

**Manual checks:**
- Open `fixtures/multi-page-statement/expected.pdf`: 4 pages, page 2's content starts on page 4, and footers read 1–4 of 4.
