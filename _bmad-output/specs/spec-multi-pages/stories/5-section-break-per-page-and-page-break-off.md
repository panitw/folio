---
title: 'Section Break per page and Page Break off'
type: 'feature'
created: '2026-09-15'
status: 'done'
route: 'dispatch'
baseline_commit: '8975ec04f95c92d517ae2b1faf600729f41417bc'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-multi-pages/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-section-break/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/multi-pages-decision-log.md'
  - '{project-root}/_bmad-output/specs/spec-multi-pages/stories/4-header-and-footer-editable-from-any-page.md'
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the frozen Intent below governs implementation.*

**The problem.** Only page 1 can have a Section Break. Unticking Page Break on a later page is saved, but the PDF still starts that page on a new printed page.

**What this story builds.**
- **A Section Break on any page:** each page can have its own break. It only moves that page's own content, following the same rules as today's break.
- **Placing, dragging and deleting:** you can do all three on any page. The palette's Place Section Break entry is disabled only when the current page already has a break.
- **Page Break off works:** a page with it off prints directly under the previous page's content when the whole page fits in the space left there. Otherwise it starts at the top of the next printed page. It never moves up onto an earlier printed page.
- **Proof:** a new golden fixture shows both cases, and the format doc describes them.

**What it does not do.**
- **Breaks:** no second break on one page, and no page break inside a page.
- **Canvas:** it still draws every page at its declared positions. Page Break off only changes the PDF.

**What done looks like.** A three-page contract has a break on pages 1 and 2, and Page Break off on page 3. The signature page prints right under the last clause when it fits, or at the top of the next page when it doesn't. Every page still carries the header, the footer and the right "Page X of Y".

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The engine, the file format and the canvas allow one section break (page 1's). Pagination ignores `pageBreak: false`.

**Approach:**
- Every designed page's content band holds its own optional break. Each break is loaded, validated, paginated, projected and edited per page, reusing the section-break rules unchanged.
- Pagination honours Page Break off with the SPEC rule, reusing the unanchored "fits under the end, else next window" mechanics for a whole page.
- The designer places, draws, selects and edits a break per page.

## Boundaries & Constraints

**Always:**
- **Load and save.**
  - `pages[i].sectionBreak` and `pages[i].sectionBreakAnchor` load on any page.
  - Range, anchor-without-break and straddle checks apply per page, against that page's own elements, located at `pages[i].<key>` (`SECTION_BREAK_INVALID` / `SECTION_BREAK_STRADDLED`).
  - Save writes each page's keys. Version rules are unchanged (4.1).
- **Commands.**
  - `setSectionBreak`, `removeSectionBreak` and `setSectionBreakAnchor` take an optional 0-based `page`.
  - Omitted means page 1, with today's field counts, bytes and results. The designer omits it for page 1.
  - An unknown page is refused at `pages`. Straddle checks judge the page the element sits on or lands on.
  - Each command is one undo entry.
- **Pagination per page.**
  - Each designed page paginates its own content with its own break, anchored or not, per `spec-section-break`.
  - `SECTION_BREAK_SPLITS_KEEP_TOGETHER` is reported per page.
  - A page without a break renders exactly as today.
- **Page Break off (CAP-9).** For page N > 1 with `pageBreak: false`:
  - **E:** the bottom of page N−1's content on its final output page, below-line section included, by the same measure as the section-break "where content ends".
  - **The block:** page N's content, paginated alone with its own break.
  - **Fits:** the block fits when it paginates to exactly one output page with nothing clipped, and E + its extent below the content window top ≤ the content window bottom. Then every item keeps its page-relative offset, moved so the window top sits at E, and joins page N−1's last output page.
  - **Doesn't fit:** otherwise page N starts on a new output page exactly as with Page Break on.
  - **Never upward:** the block never lands on an earlier output page. If page N−1's last output page is clipped (no E), it starts a new page.
  - **No overflow, no pull:** the rule applies only when page N−1's content occupies more than one output page. When page N−1 fits on a single output page, page N starts a new output page, as with Page Break on (SPEC non-goal; CAP-9 "No overflow").
  - **Empty page:** a page with Page Break off and no content, after a page that overflows, adds no output page.
- **Unchanged output.** Page counts, `Page X of Y`, header and footer cover every output page. Every existing fixture keeps its bytes and PDF hash, `multi-page-statement` included.
- **Golden fixture.** A new multi-page golden fixture proves CAP-6 and CAP-9 (a fits case and a doesn't-fit case, and a break on page 2). It joins the four-platform matrix registration, and its hash run is deferred to the end of the run.
- **Projection.**
  - One-page projections keep `sectionBreak`/`sectionBreakAnchor` exactly.
  - A multi-page projection carries `sectionBreaks` (per page: offset or null) and `sectionBreakAnchors` (per page: `false` only where unanchored, else `true`) instead.
  - `belowSectionBreak` is present exactly on content components whose own page has a break.
- **Canvas per page.**
  - Each page's break line is drawn among its own page's sheets.
  - The handle is named "Section Break" in a one-page document and "Section Break on page N" otherwise.
  - Selecting a break selects that page's break and makes its page current. Drag, arrow nudge, Delete, the Y and Anchor properties, and pending focus act on that page.
- **Palette (D-5.1).**
  - Place Section Break is disabled when the current page already has a break.
  - Its note reads "This page already has its Section Break." (one-page documents keep "This document already has its one Section Break.").
  - While armed, a content band on a page with a break is not a target.
- **Format doc.** `folio-format.md` documents the section break on any page and both rules for how one page follows another (CAP-7). The "not yet honoured" note is removed.

**Never:**
- No second break on one page.
- No break in the header or footer.
- No reflow: the Page Break off block moves as one rigid block.
- No Page Break off effect on the canvas.
- No change to one-page bytes, versions or hashes.
- No DOM measurement.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Later-page break loads | `pages[1].sectionBreak: 300` | Loads; round-trips byte-identically | N/A |
| Break out of range on page 2 | `pages[1].sectionBreak` ≥ content height | Load error | `SECTION_BREAK_INVALID` at `pages[1].sectionBreak` |
| Straddle on page 2 | Page 2 element crosses page 2's break | Load error / command refused | `SECTION_BREAK_STRADDLED` naming page 2's path |
| Set break on page 2 | `setSectionBreak{offset:200,snap:false,page:1}` | Page 2 has the break; page 1 unchanged; one undo | N/A |
| Page 1 command | No `page` | Today's bytes and behaviour | N/A |
| Unknown page | `removeSectionBreak{page:7}` | Refused, document unchanged | Located at `pages` |
| CAP-6 isolation | Breaks on pages 1 and 2; page 1's table crosses its break | Page 1 follows section-break rules; page 2's output equals rendering page 2 alone | N/A |
| CAP-9 fits | Page 1's rows end 200pt above output page 3's window bottom; page 2 off, 150pt tall | Page 2 drawn on output page 3 from the rows' end; PDF has 3 pages | N/A |
| CAP-9 doesn't fit | Same, page 2 250pt tall | Page 2 at the top of output page 4's window | N/A |
| CAP-9 no overflow | Page 1 fits one output page; page 2 off | Page 2 renders as output page 2 at its declared positions | N/A |
| Empty off page | Page 2 off, no elements | No extra output page | N/A |
| Palette | Current page 2 has a break, page 1 none | Entry disabled; select page 1: entry enabled; armed, page 2's content band is not a target | N/A |
| Canvas break per page | Breaks on pages 1 and 2 | One line and one handle per page, on that page's sheets; names "Section Break on page 1", "… on page 2" | N/A |

</frozen-after-approval>

## Code Map

- **Engine: load and save**
  - `folio-go/internal/template/parse_pages.go`:
    - `decodeContentPage` (:70-124) refuses a break on `i > 0` (:108-116). Lift that, and decode it with `decodeSectionBreakKeys` as page 0 is (:118).
    - `decodePages` (:36-67) checks shapes.
  - `folio-go/internal/template/serialize.go` `writeContentPages` (:180-206) already writes every page's keys.
- **Engine: section-break helpers** (`folio-go/section_break.go`):
  - **Page-1 load checks:** `validateSectionBreak` (:107-155) duplicates the later-page refusal (:114-124), and range, anchor and straddle only look at `firstContentBand`.
  - **Page-1 helpers:** `sectionBreakPath` (:44-46) is page 0 only. `declaredSectionBreak` (:81-90), `sectionBreakAnchored` (:94-100), `sectionBreakOf` (:62-78) and `sectionBreakSplitTags`/`sectionBreakSplitDiagnostics` (:344-393) read page 1.
  - **Straddle:** `refuseSectionBreakStraddleOnPage` (:208-218) skips `page != 0`.
  - **Commands:** `setSectionBreak` (:242-285), `removeSectionBreak` (:289-306) and `setSectionBreakAnchor` (:311-339).
  - **Beyond content:** `refuseSectionBreakBeyondContent` (:223-236) is called from `component_commands.go:3043` and `page_setup.go:2541`.
  - **Dead constants:** `sectionBreakDataPath`/`sectionBreakAnchorDataPath` (:37-40) look unused; check them.
- **Engine: other readers**
  - `folio-go/render.go`: `keepTogetherTags` (:2871-2885) applies page 1's offset to all pages, so make it per page. Split diagnostics are appended at :2534. Page count feeds `headerFooterResolver` at :2488-2509. Pass B `paginateDocument` (:3200) applies `outputShiftFor` at :3269/:3294/:3334.
  - `folio-go/pages_command.go`: `pageIndexField` (:80-92) for the optional `page`. `contentPagesForEdit`/`installContentPages` (:42-66) already carry a break with its page. `setPageBreak` (:167-197) already exists.
- **Engine: pagination** (`folio-go/content_pages.go`)
  - `contentPagesOf` (:28-29) fills only `breaks[0]`.
  - `paginateContentPages` (:73-99) always appends and never reads `PageBreak`.
  - `contentPagesPlan`/`outputShiftFor` (:56-68) assume one designed page per output page, so they must resolve shifts when two pages share an output page.
  - Reuse `section_break.go`: `paginateWithSectionBreak` (:417-518; the unanchored fit and move at :458-480), `aboveLineEnd` (:575-602), `sectionExtent` (:537-543), `unshiftSectionPages` (:550-570), `mergePageAssignments` (:609-620), `sectionPlan.shiftFor` (:399-411).
- **Engine: projection** (`folio-go/page_setup.go`): `SectionBreak`/`SectionBreakAnchor` (:652-657, set at :1004-1012), `BelowSectionBreak` (:330, page 0 only at :1013-1022), `PageBreaks` (:575, :991-996). The wire pair test is `canvas_projection_wire_test.go`.
- **Engine: fixture and matrix registration**
  - Model on `fixtures/multi-page-statement` and `multi_page_statement_template.go` + `multi_page_statement_fixture_test.go`.
  - Register in `matrix_test.go` (:753-768), `.github/workflows/matrix.yml` (:82, :132, :182, ~:232) and `byte_neutrality_test.go` (:508-517, ledger :1100-1102).
  - The registration checks are `matrix_registration_test.go` and `matrixdocs_source_test.go`.
- **Engine tests that flip:** `pages_test.go` :206 (later-page break load error), :128-141 `TestAPageBreakOffStillStartsANewOutputPageInThisStory` and :245-278; `pages_command_test.go:365`.
- **Designer**
  - `folio-designer/src/engine-protocol.ts`: `isCanvas` :762 (allowed keys), :848-851 (break fields) and :861-862 (`firstPageContent` membership); tests in `engine-protocol.test.ts` :67-85 and :262-271.
  - `folio-designer/src/section-break.ts`: `sectionBreakPlacement(canvas)` (:35-43) needs a page. The helpers `pageWindows`/`homeWindow` are in `sheet-stack.ts:88-116`. Tests are in `section-break.test.ts` :30-53.
  - `folio-designer/src/section-break-command.ts` (:16-26) gets an optional `page` through `commandBytes`.
  - `folio-designer/src/App.tsx`:
    - **Break state and handlers:** `sectionBreakSelected` (:373) becomes a page. `breakSelected` (:631). The handlers (:1879-2012) are `selectSectionBreak`, `armSectionBreak`, `placeSectionBreak`, `deleteSectionBreak`, `sendSectionBreak`, `nudgeSectionBreak`, the drag functions and `keySectionBreak`.
    - **Marker:** `sectionBreakMarker` (:2004-2012) and pending break focus `.section-break-handle` (:1699-1706).
    - **Keys and toolbar:** `keyboardDelete` (:1723), the window arrow arm (:3018) and the toolbar Delete (:3414).
    - **Placement:** `placeInBand` (:1575-1592; remove the `page !== undefined` refusal), `breakAt` (:3147) and the marker on the sheet (:3234), the `accepts` guard (:3174-3180).
    - **Palette:** the entry and note (:3388).
    - **Properties:** `SectionBreakProperties` wiring (:3462).
  - Tests: `App.test.tsx` section-break describe (:10710-10917; page-1 bytes stay) and the multi-page describe :10945-10955 and :11133-11139 (flips to a page-2 `setSectionBreak`); `e2e/section-break.spec.ts` (one page, stays).
- **Docs:** `_bmad-output/specs/spec-folio/folio-format.md` :348-349, :376 (later-page break load error), :420-440 ("at most one break"), :442-448 (Page Break off not honoured).

## Tasks & Acceptance

**Execution:**
- [ ] `folio-go/internal/template/parse_pages.go`, `folio-go/section_break.go` -- load and validate a break on every page, located per page -- format
- [ ] `folio-go/section_break.go`, `folio-go/render.go` -- commands with optional `page`; straddle, beyond-content and split tags per page -- commands
- [ ] `folio-go/content_pages.go` (+ `section_break.go` helpers) -- per-page breaks in pagination; the Page Break off rule; output-shift mapping for shared output pages -- pagination
- [ ] `folio-go/page_setup.go` (+ `canvas_projection_wire_test.go`) -- per-page break projection -- projection
- [ ] `fixtures/<new multi-page flow fixture>`, its template/test, `matrix_test.go`, `.github/workflows/matrix.yml`, `byte_neutrality_test.go` -- CAP-6/CAP-9 golden -- evidence
- [ ] `folio-go/pages_test.go`, `pages_command_test.go`, `section_break*_test.go`, `wasm/pages_test.go` -- matrix rows; flip the story-1 pins; one-page byte identity -- engine coverage
- [ ] `folio-designer/src/engine-protocol.ts`, `section-break.ts`, `section-break-command.ts`, `App.tsx` (+ tests) -- per-page break on the canvas, palette rule D-5.1 -- UI
- [ ] `folio-designer/e2e/pages.spec.ts` -- place a break on page 2, see the palette entry disabled there and enabled on page 1, save shows `pages[1].sectionBreak` (compiled; run deferred) -- e2e
- [ ] `_bmad-output/specs/spec-folio/folio-format.md` -- break on any page; Page Break on and off rules -- docs

**Acceptance Criteria:**
- Given every existing fixture and one-page document, when rendered and saved, then bytes, versions and PDF hashes are unchanged.
- Given a multi-page document with a break on page 2, when page 2 renders, then its output equals rendering page 2's content alone as a one-page document with that break.

## Spec Change Log

## Review Triage Log

| # | Source | Finding | Verdict | Route | Evidence |
|---|--------|---------|---------|-------|----------|
| 1 | blind, edge-case | A single-element duplicate judges the copy against page 1's break | high | patch | `duplicateComponent` calls `refuseSectionBreakStraddle` with the clone, whose new id is absent from `contentPageIndex`, so page 0 is used. A page-2 copy straddling page 2's break is accepted and the saved file no longer loads. |
| 2 | blind, edge-case | Undo and redo leave `sectionBreakSelected` (a page index) set, so after pages shift it names another page's break | medium | patch | `setCurrentSnapshot`'s clearDocumentInteraction branch doesn't clear it; `selectedBreak` only checks that some break exists at that index. |
| 3 | edge-case | The clipped check runs before the empty-block check, so an empty Page Break off page after a clipped overflow page adds a blank output page | low | patch | `pageBreakOffShift` returns false on `clipped` before `len(items) == 0`. Reordering is a direct correction. |
| 4 | verification-gap, blind | Page Break off is not tested with clipping, nor with a moved block that has its own unanchored break | medium | patch | Pre-verified gap: every fitting block in the tests is a probe rect or text. |
| 5 | verification-gap | Toolbar Delete and the window-level arrow nudge of a later page's break are untested | medium | patch | Pre-verified gap: the tests press keys only on the handle. |
| 6 | blind | A test title still says "the break only on page 1" | low | patch | Direct rename. |
| 7 | blind | The golden's page assertion never checks the "Acknowledged" text | low | patch | Direct assertion. |
| 8 | blind | The byte-neutrality obligation uses a different deadline wording from its siblings | low | patch | Direct wording correction. |
| 9 | verification-gap, blind | A framed or floored table in a fitting Page Break off block is untested | medium | defer | Pre-verified with a defer disposition: a narrow case that belongs in the next table-rules or multi-page pass. |
| 10 | blind | A chain of Page Break off pages stops after one merge | false | reject | The frozen constraint "applies only when page N−1's content occupies more than one output page" and the SPEC non-goal make a merged single-output page followed by a new page correct. |
| 11 | blind | `outputShiftFor` falls back to `slots[0]` silently | low | reject | Every item on an output page belongs to one of its slots' pages; the fallback is unreachable from `paginateContentPages`. |
| 12 | blind | Command field-count errors are located at page 1, and an explicit `"page":0` is untested | low | reject | Cosmetic location on a malformed command; `optionalPageField` is shared and tested. |
| 13 | blind | The 420896 constant in the golden test is unexplained | low | reject | Pinned by the fixture's hash; the implementation report explains it (8pt line). |
| 14 | blind | Decision log and story file are missing from the diff | false | reject | Both exist; the diff was limited to code, fixtures and format docs. |
| 15 | blind | A placement on a page that already has a break gives no feedback | low | reject | Spec-settled (D-5.1): the band is not a target and the entry's note names the page. |

## Verification

**Commands:**
- `cd folio-go && go build ./... && go vet ./... && test -z "$(gofmt -l .)"` -- expected: clean
- `cd folio-go && go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` -- expected: pass
- `cd folio-go && go build -tags=matrix ./... && go vet -tags=matrix ./...` -- expected: clean (matrix hash run deferred to end of run)
- `cd folio-designer && npm run build:wasm && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` -- expected: pass (Playwright deferred to end of run)
- `cd lint && go test ./...` -- expected: pass
