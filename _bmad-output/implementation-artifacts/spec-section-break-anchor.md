---
title: 'Section Break — Anchor option'
type: 'feature'
created: '2026-09-14'
status: 'done'
route: 'dispatch'
baseline_commit: '341c7b53c166ff5134a26baa018a45848793f94d'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-section-break/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-section-break/brownfield.md'
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A crossed Section Break always sends the section to a new page at its declared y. An author who wants the legend directly after the rows, on the same page when there is room, cannot express that.

**Approach:** Add an optional content-band boolean `sectionBreakAnchor` (default true). When it is false and content above ends below the line, the section is pushed down to that end on the same page, or moves to a new page with the line at the window top. The designer gets an Anchor checkbox and an anchor icon in the tab. This covers CAP-7 and the Anchor clauses of CAP-3, CAP-5 and CAP-6 of `spec-section-break`.

## Boundaries & Constraints

**Always:**
- E is where the above-line content ends: its page-space bottom, as the anchored rule computes it today.
- Unanchored, E > line: the section is shifted by D = E − line on that page if E + (extent − offset) ≤ the content window's bottom. Extent is the lowest member bottom, counting a table's `y + minHeight` floor. Otherwise it starts a new page with D = contentTop − line. Unanchored with E ≤ line: behave exactly as anchored. Unanchored with a clipped last above-line page (E undefined): treat as no room.
- Members keep their offset from the line. A section taller than a window continues under the window rules.
- Anchored behaviour, every existing golden hash, and CAP-3 byte identity are unchanged for both settings.
- File: `sectionBreakAnchor` is written only when `false`, and an explicit `true` is dropped on save. Refused with `SECTION_BREAK_INVALID` naming the band: the key without `sectionBreak`, on the header or footer, duplicated, null, or not a boolean. The key stays within 4.1 with no bump.
- Commands: `setSectionBreakAnchor {anchor}` is one command and one undo entry. `removeSectionBreak` also clears the anchor.
- Projection: `sectionBreakAnchor` is present, and `false`, only when a break exists and is unanchored. It is an optional wire key, and the Go and TS pins change together.
- Designer: an **Anchor** checkbox in Section Break Properties, checked when anchored. The canvas tab shows an inline anchor SVG icon before SECTION BREAK when anchored, and none when unanchored. The section is still drawn at its declared position only. `DESIGN.md` Section Break row and paragraph mention the icon.
- Integer fixed-point only, and byte-identical on all four targets.

**Never:**
- Pulling an unanchored section up.
- Changing `layout.Paginate` for callers that do not opt in.
- A pushed copy of the section on the canvas.
- A printed line.
- A version bump beyond 4.1.
- Changes to the existing `section-break-statement` fixture or its hash.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Not crossed | Unanchored, rows end above the line on page 1 | Legend at declared y; hash equals anchored and no-key renders | N/A |
| Pushed, fits | Unanchored, rows end 120pt past the line on page 1; legend fits | Legend on page 1 at declared y + 120; 1 page; no overlap | N/A |
| Pushed, no room | Unanchored, rows end near the window bottom | Legend on page 2, line at window top, offset kept; page count includes it | N/A |
| Later page | Unanchored, rows fill page 1 and end past the line on page 2 | Pushed on page 2, or page 3 at window top | N/A |
| Floor crosses | Unanchored, `minHeight` floor bottom is past the line | Pushed from the floor bottom | N/A |
| Tall section | Unanchored, section holds a growing table, moved to a new page | Starts at window top and continues onto later pages | N/A |
| Invalid key | Anchor without break, on footer, null, `"no"`, duplicated | Not loaded | `SECTION_BREAK_INVALID` naming the band |
| Round-trip | `false` / explicit `true` | `false` byte-identical / `true` dropped | N/A |
| Toggle | Break selected; uncheck Anchor | One command; tab icon gone; undo restores icon and bytes | N/A |
| Remove | Unanchored break removed | Both keys gone; undo restores both | N/A |

</frozen-after-approval>

## Code Map

- **`folio-go/section_break.go`:**
  - `aboveLineEndsAtOrAbove` (:392): split into a helper returning `(end, clipped)`, keeping the bool wrapper.
  - `paginateWithSectionBreak` (:316): keep the fast path (:337) and the anchored path as they are.
  - Unanchored placement: move every `below` item's `Top`/`Bottom` by D before `planS`. Set `sectionShift[p] = planS.Pages[p].Shift − D`, and subtract D from each section page's `HeaderRepeats[].Shift` (render.go:3245 applies it to unmoved Y).
  - When pushed, merge with `mergePageAssignments` (:426); otherwise append.
  - Add a floor-aware extent helper, and do not reuse `sectionBreakDeclaredBox` (:108, the straddle rule).
  - `setSectionBreak`/`removeSectionBreak` (:176/:223) are the command pattern (`componentFields`, `commandBool`, `componentFailure`, restore on error). Register a case at `component_commands.go:305`.
  - `validateSectionBreak` (:79) covers the key without a break.
- **`folio-go/internal/template/`:**
  - `model.go:404`: `SectionBreakAnchor Presence[bool]`.
  - `parse_bands.go:103-139`: mirror `sectionBreak`'s consumed, header/footer, duplicate and null refusals.
  - `serialize.go:331`: write it only when `false`.
  - `version.go:359,431`: probe the anchor with the 4.1 rank.
  - `fixtures_test.go:417`: add to `maximalFixture`.
  - `drift_test.go`: needs the row in `folio-format.md`.
  - Tests: `section_break_test.go`.
- `folio-go/page_setup.go:633,~967` -- `SectionBreakAnchor *bool` with omitempty. `canvas_projection_wire_test.go:79` optional keys.
- **Go tests to model on:**
  - `folio-go/section_break_test.go`: helpers `sectionBreakTestDoc`, `sectionBreakPages`, `legendOn`, `tableEnd`; tests at :144, :202, :269, :404, :475.
  - `section_break_command_test.go` and `wasm/section_break_test.go` (undo/redo bytes).
- **Golden registration, as for `section-break-statement`:**
  - `*_template.go` and `*_fixture_test.go`
  - `byte_neutrality_test.go:499,1074,1700`
  - `matrix_test.go:739-751,2037`
  - `render_test.go:883,1104`
  - `wasm/engine_test.go:326`
  - `.github/workflows/matrix.yml:80,128,176,224,283`
  - The hash is recorded by hand in `expected.json`, README and the digest record.
- **`folio-designer/src/`:**
  - `engine-protocol.ts:454,749,819` plus `engine-protocol.test.ts:62`.
  - `section-break-command.ts`: add `setSectionBreakAnchorCommand` via `commandBytes`/`jsonBoolean`, with its test. Do not add a new factory file (`command-json-soleness.test.ts`).
- **`folio-designer/src/App.tsx`:**
  - `SectionBreakProperties` (:5353): add `anchor`/`onAnchor`, using the checkbox pattern at :5301.
  - Render site :3317: put the anchor in `key`.
  - Tab :1901: the icon uses the `SectionBreakIcon` SVG conventions (:243).
- `App.css:309` (tab inline-flex plus gap), `App.test.tsx:10687` (describe block, :10821 test to update), `e2e/section-break.spec.ts`.
- `_bmad-output/specs/spec-folio/folio-format.md:47,347,385-403` -- version prose, bands row, Pagination rules. `DESIGN.md:508,535-537`.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/internal/template/{model,parse_bands,serialize,version}.go` and tests -- key, refusals, omit-when-true, 4.1 -- CAP-5.
- [x] `folio-go/section_break.go`, `section_break_test.go` -- unanchored pagination and every matrix row, including page count and pass A/B agreement -- CAP-7, CAP-3, CAP-4.
- [x] `folio-go/section_break.go`, `component_commands.go`, `page_setup.go` and tests -- `setSectionBreakAnchor`, remove clears the anchor, projection field and wire pins -- CAP-6.
- [x] `fixtures/section-break-unanchored/` and all registrations -- a synthetic statement pushed on page 1 -- golden.
- [x] `folio-designer/src/{engine-protocol,section-break-command}.ts`, `App.tsx`, `App.css` and tests, plus `e2e/section-break.spec.ts` -- checkbox, tab icon, undo.
- [x] `folio-format.md`, `DESIGN.md` -- key row, unanchored pagination rule, icon.

**Acceptance Criteria:**
- Given any existing fixture or golden, when loaded, saved and rendered, then bytes and hashes are unchanged.
- Given the unanchored fixture, when rendered on the four targets, then the hashes match the recorded digest and preview equals native.
- Given any accepted anchor command, when Undo then Redo run, then document bytes return exactly to before and after.

## Implementation Notes

- The extent helper (`sectionExtent`) measures the section as `aboveLineEnd` measures the content above: it paginates the section's items at their declared offset and takes its page-space end, counting slices (the `minHeight` floor), floor pushes and row displacement. A section that needs more than one window, or clips, has no room.
- Besides `HeaderRepeats[].Shift`, `ClippedRects[].Bottom` is also a column-space quantity the render applies to unmoved coordinates, so `unshiftSectionPages` subtracts D from both. `TableSlices` are page space and need nothing. `TestSectionBreakUnanchoredClippedRowIsCutAtTheContentBottom` pins it.
- `setSectionBreakAnchor` refuses `anchor: null` explicitly: `commandBool` (shared) reads JSON null as false.
- The parser refuses the Anchor key without `sectionBreak`; `validateSectionBreak` repeats the rule defensively.
- Golden `fixtures/section-break-unanchored/`: break at 390 (20pt above the legend), anchor off, the first 35 of section-break-statement's transactions; the rows end ~112pt past the line and the legend is pushed on page 1. sha256 `3ae4e8d50eccbc10f0aa055490a1184a8b9601496e2ac0d58613319b61b5186b` (recorded by hand in expected.json, README and goldenDigestRecord). The four-target matrix run is owed at the next matrix gate.
- Review pass 1 patches: B1 (rule 5 now defines extent as laid out at the declared offset), B3 (panel key no longer carries the anchor, so the checkbox keeps focus), B4 (checkbox beside its label), B5 (accurate Properties notes), V1 (repeated-header position asserted on continuation pages), V2 (`TestSectionBreakUnanchoredClippedAboveLineMovesToANewPage`). V1 and V2 were each confirmed by mutation.
- Since B3 the checkbox keeps focus after a toggle. The window shortcuts ignore keys from any `INPUT` (`isEditableTarget`), so Cmd+Z pressed on the focused Anchor checkbox does nothing, the same as the existing Border edges checkboxes. The e2e test moves focus to the line's handle before undo. Changing that app-wide rule is out of scope.

## Spec Change Log

## Review Triage Log

Pass 1 (blind-hunter, edge-case-hunter, verification-gap):

| # | Finding | Verdict | Evidence | Route |
|---|---|---|---|---|
| B1 | `folio-format.md` rule 5 says extent is "as declared", but `sectionExtent` measures the section as laid out at its declared offset | low | Confirmed: rule 5 wording against `sectionExtent`. The code matches the owner's "no room means next page, top of content", so the doc is wrong | patch (doc) |
| B2 | A null Anchor presence reads as anchored but saves as `false` | false | No writer sets `Null`: the parser refuses null, the commands set only `Set`/`Value`, and `internal/template` cannot be imported from outside the module | reject |
| B3 | `SectionBreakProperties` key includes the anchor, so toggling remounts the panel and drops keyboard focus | low | Confirmed: key `…:${canvas.sectionBreakAnchor ?? true}` remounts on the engine's answer. The checkbox is already controlled by `anchor`; direct correction | patch |
| B4 | The Anchor label has no CSS, so the global `label { display: grid }` stacks the checkbox above its text | low | Confirmed: `App.css:1066` is global grid, and only `.property-edges`/`.table-header-edges` give their checkboxes flex; direct CSS rule | patch |
| B5 | Properties notes misdescribe unanchored: "follows directly after it" ignores the kept gap and the new-page window top | low | Confirmed against rule 5 and the fixture (legend 20pt below the rows); text correction | patch |
| B6 | Anchored state is not exposed to screen readers on the canvas | low | The Properties checkbox exposes the state whenever the break is selected. Changing the handle's name touches e2e selectors and adds surface | reject |
| B7 | DESIGN.md does not document the Properties checkbox or its notes | low | Frozen intent requires only the row and paragraph to mention the icon; documentation completeness | reject |
| B8 | No exact-boundary test for `<=` against the content bottom | low | Real, but the oracle orders pushed before moved crossings. An exact-millipoint fixture adds test machinery for a rare slip | reject |
| B9 | `validateSectionBreak` anchor-without-break branch untested; command null check unproven | low | The branch is defensive and unreachable through the parser. The null refusal is proven: without it `commandBool` accepts null and `sectionBreakRefusal` fails | reject |
| B10/E1 | Tests discard `SerializeTemplate` errors and index `Rects[0]` unguarded | low | Test-only; a panic still fails the test | reject |
| B11 | The section is paginated twice on the unanchored path | low | Real cost, bounded by section size; reuse adds branching for no observed harm | reject |
| B12/E3 | The four-target matrix run for the new golden has not run | low | Same as the story-1 precedent: legs are wired and the run is owed at the next matrix gate, per the obligation note; no code defect | reject |
| B13 | Spec file absent from the review diff | false | Deliberate: the spec goes to the edge-case layer only | reject |
| E2 | Re-paginating a pushed section could spill onto an added page | maybe-false | Pagination is translation-invariant inside a window and the fit test uses the same extent plus D. Even if it spilled, it continues under window rules (low) | reject |
| V1 | Repeated header shift in a moved unanchored section is untested; a mutant draws headers 75pt low | medium | Pre-verified by mutation | patch |
| V2 | Clipped above-line content with an unanchored break is untested; a mutant draws the legend above the content window | medium | Pre-verified by mutation | patch |

## Design Notes

Shifting item coordinates by D before paginating `planS` reuses the window rules unchanged. D > 0 pushes the section down on the shared page. D = contentTop − line lifts the section's first window to the page top, and because every member's y ≥ line, no item goes above contentTop. The render subtracts `sectionShift` from the declared y, so the net page position is declared y + D − Shift.

## Verification

**Commands:**
- `cd folio-go && gofmt -l . && go vet ./... && go test -count=1 ./...` -- expected: clean, apart from the known baseline failure `TestCorpusMeetsP6ExerciseFloors`.
- `cd lint && go test -count=1 ./...` -- expected: pass.
- `cd folio-designer && npm run lint && npm run typecheck && npm test && npm run test:e2e` -- expected: pass.

**Manual checks:**
- Open `fixtures/section-break-unanchored/expected.pdf`. The legend follows the last row, is not overdrawn, and appears once.
