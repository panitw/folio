# Multi-Pages delivery run — decision log

Program: SPEC-multi-pages (`_bmad-output/specs/spec-multi-pages/`), stories 1–5 from `stories.yaml`.
Started 2026-09-14 from baseline `b6b8b6f`.

## Lead Grounding

Filed 2026-09-14 from the lead's grounding report. Baseline b6b8b6f.

**Code verified.**
- `version.go`: SupportedVersion is "4.1", which covers sectionBreak only. There is no release tag, so multi-page joins 4.1.
- `parse_bands.go`: `bands` is a closed set of three keys.
- `parse.go`: unknown top-level keys pass through as Extra.
- `section_break.go` seams: `aboveLineEnd`, `sectionExtent`, `mergePageAssignments`.
- `page_setup.go`: `CanvasProjection` has a flat `ContentWindowOrigins`.
- `sheet-stack.ts`: MAX_CANVAS_SHEETS is 120.

**Invariants.**
- One-page files keep their bytes, version and PDF hash (AD-9, AD-21).
- The engine is the only authority on layout, undo and the canvas projection. Page selection is UI state only (AD-15). The DOM-measurement ban holds (AD-17).
- Page numbers stay late-bound slots. The total is summed across pages, and pass two lays nothing out (AD-4).
- Arithmetic is int64 millipoints, never float (AD-2, AD-23).
- Nothing reflows within a page (AD-24). Page Break off is a second rigid-block exception and must be written into Pagination.
- Every diagnostic is registered and names its page.
- There is one header definition and one footer definition.

**Under-determined, needs the owner.**
- **F1, the file shape for pages.** MINOR means optional keys only, and `bands` stays exactly three keys.
  - (a) `bands.content` stays page 1, and a top-level `pages` array holds pages 2..N.
  - (b) A multi-page file puts every page in `pages`, and `bands.content` is empty.
  - The lead leans (b): it gives CAP-7's zero-pages error and its page-1 Page Break rule something to attach to.
- **S4, which page is "current" after a header click.** The header belongs to no page, so clicking page 3's header copy would make page 1 current under SPEC's definition. The lead leans toward reading "current page" as the page whose copy was selected.

**Provisional leanings.**
- **S1:**
  - Page Break false is parsed and saved but renders as on until S5.
  - A sectionBreak on page 2 or later is refused, naming the page, until S5.
  - The projection groups windows per page rather than duplicating origins.
  - S1's fixture covers CAP-5. S5 adds its own fixture.
  - An empty designed page renders a page with only the header and footer, so "no page is ever blank" is qualified in the doc.
  - The Add Page and Delete Page engine commands belong to S2.
- **S2:**
  - The setPageBreak command and its checkbox are built here. The off rule lands in S5.
  - Add inserts after the selected page.
  - The 120-sheet cap truncates the tail of the stack.
- **S3:**
  - `moveComponent` gains a target page.
  - Possible escalation: a multi-selection spanning pages, and whether select-all, duplicate and paste act on the current page or the whole document. Keep today's behaviour unless forced.
- **S5:**
  - Page Break off mirrors unanchored rule 5, with a shift of E − 0.
  - The previous page's end E counts its below-line section.
  - Chained off-pages are allowed.
  - The per-page break reuses `sectionBreakOf` per column, and `refuseSectionBreakBeyondContent` iterates over pages.

## Owner decisions from the lead's grounding

### D-G.1: A multi-page file lists every page in a top-level `pages` array
**Owner decision**, taking the lead's recommendation.

**Verdict.** A document with more than one page writes all of its pages, page 1 included, as entries of a top-level `pages` array, and `bands.content` holds no elements. A one-page document keeps today's shape: elements in `bands.content`, no `pages` key. Each of the following is a registered load error: content in both `bands.content` and `pages`; a `pages` array with only one entry; a `pages` array with zero entries.

**Situation.** The owner ruled that multi-page is a MINOR change inside 4.1 (D-S.7). The format's MINOR rule allows only new optional keys, and `bands` is a closed set of three keys: `pageHeader`, `content` and `pageFooter`. So `bands.content` cannot be replaced with an array. The new data has to go in a new top-level key, which leaves the question of where page 1 lives.

**In simple terms.** Take a two-page contract.
- Under (a), page 1's clauses sit in `bands.content` and page 2's sit in `pages[0]`. Page 1 and page 2 live in different places and use different key names.
- Under (b), both pages sit side by side as `pages[0]` and `pages[1]`, and `bands.content` is left empty.

A reader of the file, or an AI agent editing it, sees one list and never has to learn that page 1 is special.

**Options considered.**
- **(a) `bands.content` is page 1 and `pages` holds pages 2..N.** It is closer to today's bytes. But SPEC CAP-7's "zero pages" load error and its rule that a page-1 Page Break value is ignored then have nothing to attach to, so the spec would need rewording. It also gives page 1 a different shape from every other page.
- **(b) Every page goes in `pages`.** This was chosen.

**Why this one wins.** It is the only reading that honours every CAP-7 clause, and it treats all pages the same way. The accepted cost is new refusal rules for content in both places and for a single-entry `pages` array. There are also two canonical shapes, since one-page files stay as they are, so the save path has to choose between them.

**Consequences.** This fixes the canonical bytes of multi-page files permanently.
- The saver writes the `pages` shape exactly when there are two or more pages.
- Deleting a page down to one page must convert the document back to the `bands.content` shape.
- Adding a page to a one-page file must move its content into `pages[0]`.
- `folio-format.md` documents both shapes. SPEC.md's format constraint is updated to match.

**How we'd know it was wrong.** A one-page fixture changes bytes, or hand-authored files regularly trip the both-places error.

### D-G.2: The page whose header or footer copy was clicked becomes the current page
**Owner decision**, taking the lead's recommendation. It changes SPEC CAP-4's definition of the current page.

**Verdict.** The current page, the page whose header and footer copy carries the accessible name, is the page the author last selected something on. Clicking page 3's header copy makes page 3 current. Otherwise the current page is the page of the selected element, or page 1 when nothing is selected.

**Situation.** The header belongs to no page. Under the earlier definition (the selected page, or the page of the selected element, or page 1), clicking page 3's header copy would select a component that has no page, so page 1 would become current. The accessible name and focus would then jump to page 1's copy, away from the one the author just clicked.

**In simple terms.** A screen-reader user tabs to page 3's header and presses Enter to edit the party name. Under the old rule their focus would teleport to page 1's header. Under this rule it stays where they are.

**Options considered.** Keeping the spec's rule as written was rejected, because focus leaves the clicked element.

**Consequences.** This is designer selection state and belongs to story 4. The engine is not involved.

**How we'd know it was wrong.** An e2e test finds focus anywhere other than on the copy that was clicked.

## Standing decisions

### D-S.1 — Build order is stories.yaml list order, 1 through 5
**Owner decision** (at breakdown). The stories are layered: 1 is the engine and file format, 2 draws pages on the canvas, 3 moves elements between pages, 4 lets the header and footer be edited from any page, and 5 adds the per-page Section Break and Page Break off. Each story builds on the committed code of the one before it, so they run one at a time. Stories 5 and 6 of the first breakdown were merged, because both reuse the section-break pagination code.

### D-S.2 — No spec review or pause after a story; the orchestrator and lead approve CHECKPOINT 1
**Owner decision.** `spec_checkpoint` and `done_checkpoint` are false on every story. The run continues on its own and pauses only when the lead escalates a design decision to the owner. The owner sees each plan through the plain-terms opener in the per-story report.

### D-S.3 — Owner questions are asked at the terminal
**Owner decision.** A Telegram config exists, but the owner chose the terminal as the answer channel.

### D-S.4 — Heavy-test cadence: end of run
**Owner decision.** The owner accepted that a failure found at the end could come from any of the five stories.
- **Every story:** the `## Verification` section runs, in folio-go, `go build ./...`, `go vet ./...`, the gofmt check, `go test -count=1 ./...` (skipping the CI `KNOWN_RED_TEST`), and a compile-only `go build -tags=matrix ./...`. In folio-designer it runs `npm run typecheck`, `npm run lint`, `npm test` and `npm run test:e2e:compile`.
- **Deferred to the end of the run:** `npm run test:e2e` (Playwright) and the matrix hash test `go test -tags=matrix -run TestTargetRenderHash .`. Locally that covers only the host platform; the four-platform comparison runs in CI's matrix workflow.
- **Tests are still written in their story.** Only running them moves to the end.

### D-S.5 — Commit to main, never branch, never push
**Owner decision** (standing memory). The builder commits locally on `main` and turns down step-05's offer to push or open a PR.

### D-S.6 — After any folio-go change, run `npm run build:wasm` before designer tests
**Owner decision** (standing memory, and `invoke_dev_with` on every story). `npm run dev` never rebuilds the engine. `npm test` and `npm run typecheck` already run it.

### D-S.7 — Spec decisions are settled and must not be re-opened
**Orchestrator decision.** The builder must not raise any of the following as Open Questions:
- everything in `SPEC.md` Constraints and Non-goals
- the memlog's resolved questions: a 4.x MINOR version; a Delete Page confirmation that names the page; Page Setup plus a page section holding Page Break; the current page's header and footer copy carrying the accessible name
- Page Break is saved as true or false on every page after page 1, a missing value loads as on, and a value on page 1 is ignored and dropped on save
- Page Break off follows directly after the previous page's content, like an unanchored section break
