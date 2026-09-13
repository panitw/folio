---
title: 'Section Break — format and engine'
type: 'feature'
created: '2026-09-13'
status: 'done'
route: 'dispatch'
baseline_commit: '5a3a6c8f1bd54e9d9368a83e0ec1a69c53732756'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-section-break/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-section-break/brownfield.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A growing table draws over the static content placed below it in the content band, because Folio never moves an element when a neighbour grows (AD-24).

**Approach:** Add an optional content-band key `sectionBreak` (points from the content band's top). Content declared below the break is one section. It lands on the page where the above-line content ends, if that content ends at or above the line, or else on a new following page. In both cases each element keeps its declared `y`. This story covers CAP-2 to CAP-5 of `spec-section-break`: format, engine and fixture only.

## Boundaries & Constraints

**Always:**
- Section membership is decided by declared `y >= sectionBreak`, whatever the `x`.
- An element's declared box may not straddle the break. For a table, the box is `y` to `y + headerHeight`. Rows and the `minHeight` floor may run past the line.
- E is where the above-line content ends: the maximum page-space bottom on its last page, counting the table's `minHeight` floor (`TableSlice.Bottom`) and row displacement. If E ≤ break, the section renders on that page. Otherwise it renders on an added page. Either way it sits at its declared `y` with nothing drawn above it, and a section that doesn't fit continues under the existing rules.
- A document with a break, whose above-line content fits on page 1 and ends at or above the line, renders exactly as it does without the key.
- The page count, `{{pages}}` and `{{page}}` include an added page. The header and footer are drawn on it.
- A `keepTogether` group with members on both sides of the break is split at the line, and each side stays together. Every render emits a registered Warning `SECTION_BREAK_SPLITS_KEEP_TOGETHER` naming the group, and `-strict` fails on it.
- Registered load Errors, each naming the band or the element:
  - a break ≤ 0 or ≥ the content height
  - a duplicate `sectionBreak` key
  - `sectionBreak` on the page header or page footer
  - a straddling element
- The key raises the document to `4.1` (`SupportedVersion` 4.1, `SupportedMajor` 4). Documents without it keep their version and bytes, and every existing golden hash is unchanged.
- Integer fixed-point only, and byte-identical on all four targets.

**Never:**
- Designer UI, canvas drawing of the break, or commands that refuse moves or resizes across the line (story 2).
- A pushed section on the canvas projection. The canvas keeps plain `Paginate`.
- Changing `layout.Paginate`'s behaviour for callers that don't opt in.
- A printed break line.
- More than one break.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Not crossed | 5 rows end above the line on page 1 | Legend on page 1 at declared y; PDF hash equals the same document without the key | N/A |
| Crossed on last page | Last row passes the line on page 1 | Legend on page 2 at declared y; page 2 empty above it; 2 pages | N/A |
| Ends above line later | Rows fill page 1 and end above the line on page 2 | Legend on page 2 at declared y | N/A |
| Floor crosses | `minHeight` floor bottom passes the line | Treated as crossed | N/A |
| Empty section | Break set, nothing at or below it | Same as not crossed; no page added | N/A |
| Tall section | Section holds a growing table | Continues onto later pages under existing rules | N/A |
| Split group | `keepTogether: "sig"` with a member each side | Each side placed as its own group | Warning naming `sig`; `-strict` fails |
| Invalid break | `sectionBreak` ≤ 0, ≥ content height, duplicated, or on a header or footer | Not loaded | Located load Error |
| Straddle | Text box `y` 700, `height` 40, break 720 | Not loaded | Load Error naming the element |

</frozen-after-approval>

## Code Map

- `folio-go/internal/template/model.go:396` -- `Band`: add `SectionBreak Presence[geom.Length]`.
- `folio-go/internal/template/parse_bands.go:61-108` -- `decodeBand`: band keys are open (`extraFields` passes unknown keys through). Copy the `height` handling at `:86-100`. Refuse the key on the header and footer and refuse a duplicate key; don't let it fall into `Extra`.
- `folio-go/internal/template/serialize.go:316-334` -- `writeBand`: emit `sectionBreak` after `height`, before Extra. The drift tests (`drift_test.go:232,269,402`) need the key in `folio-format.md` and in `maximalFixture`.
- `folio-go/internal/template/version.go:110-113,149-177,321-398,407-437` -- `sectionBreakVersion = "4.1"`, a rank appended after `rankBarcode`, and a band probe beside the fonts probe (`:345`). Tests: `version_test.go:59,82,115`.
- `folio-go/folio.go:99`, `table_min_height.go:36-56` -- the pattern for `validateSectionBreak` (range and straddle need `layout.ContentHeight`, which package `template` cannot import).
- `folio-go/internal/diag/diag.go:246-266,410-436,453-479`, `diagnostic.go:386-391`, `diagnostic_registry_census_test.go:24,284-300` -- new codes, dispositions, aliases, triggers and warning witness.
- `folio-go/render.go:2477-2481` (pass A, `pageCount`), `:2943-3154` (pass B), `table_footer.go:208` (orphan-footer re-paginate) -- both passes must use one section-aware wrapper around `paginateWithFooterOrphanFix`.
- `folio-go/render.go:3215-3297` -- per-page transforms: Shift, `rowDisplacementFor`, `elementPushFor`, `applyTableFrames`. These are the inputs to E.
- `folio-go/internal/layout/paginate.go:664,839-840,1156-1175,1308` -- window 0 always has Shift 0, and there is no forced page start. Refs are global item indices. Any new capability goes behind an opt-in option.
- `folio-go/render.go:2831-2892` -- `keepTogetherTags`, `keepTogetherGroup`, `orKeepTogether`: split keys here so pass A, pass B and the canvas (`page_setup.go:1216,1435`) agree. Keep the key namespacing described at `:2757-2806`.
- `folio-go/page_setup.go:1018-1033,1193-1268` -- canvas window count on plain `Paginate`. Leave it unchanged.
- Test helpers: `table_pagination_test.go:29,76,136,904`, `keep_together_fixture_test.go:25,37,353,1194`, `table_ruled_form_test.go` (floor), `render_test.go:1808,1836` (`renderInSubprocess`, `firstDivergence`). Update `TestBothPaginationPassesAgree*`.
- Fixture registration, as for barcode: `byte_neutrality_test.go:494-506,1060`, `matrix_test.go:737-749,2016-2029`, `.github/workflows/matrix.yml:78,125,172,219,279`, `render_test.go:1086`, plus a `*_template.go` const and a `*_fixture_test.go`.
- `_bmad-output/specs/spec-folio/folio-format.md:47,331-346,348-378` -- the version chain, bands, and Pagination (AD-24 at `:369-372`, gap collapse at `:374-376`).

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/internal/template/{model,parse_bands,serialize,version}.go` and tests -- key, refusals, duplicate check, round-trip, 4.1 -- CAP-5.
- [x] `folio-go/internal/diag/diag.go`, `diagnostic.go`, `folio.go`, new `section_break.go` and census -- range and straddle load Errors, split Warning -- CAP-5.
- [x] `folio-go/render.go`, `table_footer.go` (and `internal/layout/paginate.go` only behind an opt-in option) -- the section-aware pagination wrapper used by both passes, and `keepTogether` splitting -- CAP-2, CAP-3, CAP-4.
- [x] `folio-go/section_break_test.go` -- every matrix row, page count and `Page X of Y` on an added page, pass A/B agreement, and no-break byte identity via subprocess.
- [x] `fixtures/section-break-statement/` and its registrations -- synthetic bilingual statement with 40 rows that crosses the line (no real customer data) -- CAP-2 golden.
- [x] `_bmad-output/specs/spec-folio/folio-format.md` -- `sectionBreak` key, the pagination rule, the narrowed AD-24 and gap-collapse sentences, and 4.1 in the version chain.

**Acceptance Criteria:**
- Given any existing fixture or golden document, when it is loaded, saved and rendered, then its bytes and hashes are unchanged.
- Given a document with `sectionBreak`, when it is saved, then it declares `4.1` and round-trips byte-identically.
- Given the section-break fixture, when rendered on the four matrix targets, then the hashes match the recorded digest.
- Given a design-canvas projection of a document with a break, when it is projected, then it matches the projection without the key, apart from any `keepTogether` split.

## Implementation Notes

- The engine half is in `folio-go/section_break.go`. `paginateWithSectionBreak` wraps `paginateWithFooterOrphanFix` for both passes. `layout.PaginateWithItemPages` is an opt-in that reports each item's page, and `layout.Paginate` is unchanged.
- Refusals use two codes. `SECTION_BREAK_INVALID` covers range, duplicate, null, and the key on a header or footer, and is located by data path `bands.<band>.sectionBreak`. `SECTION_BREAK_STRADDLED` covers the straddle case and names the element. The below-line side of a split group uses group index -3.
- Straddle is judged on the declared box. A text element with no declared height is treated as a point, so its wrapped lines can still cross the line at render; when that happens, the section moves to an added page.
- A `minHeight` floor push can still move section elements on the fast path, as it does today. No push applies to a section merged onto a later page.
- `section-break-statement` is added to `textRiseExemptGoldens` because its Thai tone marks are shaped with a rise.
- `TestCorpusMeetsP6ExerciseFloors` (`internal/text`, P6g 7 < 20) also fails on a clean baseline checkout at `5a3a6c8`. It fails independently of this change.
- Review pass 1 patches:
  - A non-numeric break is now coded `SECTION_BREAK_INVALID`.
  - `folio-format.md` qualifies the identity sentence, fixes rule 4, and restores the bands skeleton.
  - New tests: `TestSectionBreakFloorPushCountsTowardWhereContentEnds` and `TestSectionBreakSharedPageKeepsTheSectionsFramesAndPushes`, each confirmed by deleting the code it covers. No test covers the `ClippedRects` merge on a shared page; that needs an over-tall row clipped there.
- Not run here: the four-target matrix legs (owed at the next matrix gate) and a visual check of `expected.pdf`.

## Spec Change Log

## Review Triage Log

Pass 1 (blind-hunter, edge-case-hunter, verification-gap):

| # | Finding | Verdict | Evidence | Route |
|---|---|---|---|---|
| B1 | 4.1 is a MINOR, so a 4.0 reader draws over the section | low | Real: band keys are open passthrough. A MINOR for this key is the owner's frozen decision, and changing it edits this spec | reject |
| B2 | Canvas page count differs from the render; no designer authoring | false | Frozen intent: the canvas keeps plain `Paginate` and designer work is story 2 | reject |
| B3/V-other | Canvas projection changes for a split group, untested | low | The AC explicitly allows the split difference, and the fix only adds a test | reject |
| B4/E1/E2 | Non-numeric `sectionBreak` is uncoded, so it becomes TEMPLATE_MALFORMED with no data path | low | Confirmed: `decodePointsRaw` gives an uncoded LoadError, and `wrapTemplateError` (`render_error.go:109-121`) falls back and drops the field. The fix is a direct correction | patch |
| B5/E3/E4/E9 | "Renders exactly as without the key" is false when a keepTogether group is split | low | Confirmed: `keepTogetherTags` splits on every render with a break, including the fast path. It is only reachable with a straddling group that doesn't fit on page 1. The owner decided the break wins, so the fix qualifies the doc sentence | patch (doc) |
| B6 | Pagination rule 4 refers to its own table | low | Confirmed in the `folio-format.md` diff; direct correction | patch |
| B7 | Empty section, hidden elements and zero-height boxes are undocumented | low | The behaviour is correct and tested (empty section); only the prose is missing, and no user meets a defect | reject |
| B8 | Range check is skipped when `pageGeometryOf` errors | maybe-false | Copies `validateTableMinHeights`. It would need a document that loads while its page geometry is invalid, and even then it would only be low | reject |
| B9 | Images, rects, barcodes and QR codes below the break are untested | false | All rect sources carry `elementID` (`render.go:3025`), and images use `imageRuns[ref].elementID`, so every kind goes through `shiftFor`, like the tested text and table rects | reject |
| B10 | Duplicate-key refusal exists only for this key | false | The frozen intent requires refusing a duplicate `sectionBreak`; other keys are out of scope | reject |
| B11 | `render_error.go` passes the field only for one code | low | After the B4 patch, every section-break band error carries that code. The other band errors behaved this way before this change | reject |
| B12 | The canonical bands skeleton now shows an optional, page-dependent key | low | Confirmed in the `folio-format.md` diff; direct revert | patch |
| B13/E7 | Four-target matrix run and manual PDF check still owed | low | Recorded in Implementation Notes and the obligation note, as in the barcode precedent; no code defect | reject |
| B14 | `indexFrom` duplicates `strings.Index`; indirect side loop | low | Test-only direct simplification | patch |
| E5 | `sectionBreakWithout` panics on an absent id | low | Test helper only called with ids that exist; the fix adds a guard | reject |
| E6 | Task says the no-break byte identity is checked in a subprocess, but the test is in-process | low | Rendering is pure, and the fresh-process statement test covers the section path's process state. Adding a subprocess needs new TestMain plumbing | reject |
| E8 | maximalFixture bytes changed | false | It is an internal test fixture extended on purpose for the drift test, not an existing golden document | reject |
| V1 | Floor push term in `aboveLineEndsAtOrAbove` is untested | medium | Pre-verified by mutation | patch |
| V2 | Merged TableSlices/ElementPush/ClippedRects on a shared page are untested | medium | Pre-verified by mutation | patch |
| V3 | Section Suppressed/Clipped page re-basing and `diagsS` are untested | medium | Pre-verified by mutation. It affects warning text only, and the setup needs purpose-built geometry | defer |

## Design Notes

The byte-identity guarantee comes from a fast path. Paginate the above-line items alone first. If that yields one page and E ≤ break, run today's single `Paginate` over all items unchanged. Otherwise paginate the section items separately. Put them on the last above-line page (E ≤ break) or on a new page, with that page's section Shift set so that declared `y` maps to content origin + `y`. Then merge by global refs. On a shared page, preserve authored emission order.

## Verification

**Commands:**
- `cd folio-go && gofmt -l . && go vet ./... && go test -count=1 ./...` -- expected: no gofmt output, and all tests pass.
- `cd lint && go test -count=1 ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm test` -- expected: pass. Proves no projection wire drift.

**Manual checks:**
- Open `fixtures/section-break-statement/expected.pdf`. The legend is not overdrawn and appears once, at its designed position, on the last page.
