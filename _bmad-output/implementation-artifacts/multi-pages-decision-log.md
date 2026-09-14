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

## Owner decisions from story 2 planning

### D-2.1: Delete page targets the selected page, or the one page the selected elements are on
**Owner decision**, against the planner's recommendation to enable it only with a page selected.

**Verdict.** Delete page targets the selected page. When no page is selected, it targets the page holding the selected content elements, provided they all sit on one page. It is disabled, with its reason shown, when nothing is selected, when the selection is only header or footer elements or the section break, or when the selected elements span pages. The confirmation always names the target page.

**Situation.** SPEC CAP-2 says "select a page and delete it", but authors mostly have an element selected, not a page. The owner wanted the button to be usable from an element selection.

**In simple terms.** Click the signature line on page 3, press Delete page, and the dialog asks "Delete page 3?". Select a clause on page 1 and one on page 2 together, and the button is disabled, because there is no single page to name.

**Options considered.** Enabling the button only while a page is selected was the safer, more explicit option, and the owner chose convenience over it. The confirmation naming the page is what keeps the convenient option safe.

**Consequences.** The designer needs each component's page, which story 2 projects. "The page of the selection" is a narrower idea than story 4's "current page" (D-G.2) and must not be merged with it silently.

**How we'd know it was wrong.** Authors confirm deletes of pages they did not mean to delete.

### D-2.2: The Delete and Backspace keys never delete a page
**Owner decision**, against the planner's recommendation.

**Verdict.** With a page selected, Delete and Backspace do nothing. A page is deleted only from the toolbar's Delete page button. Deleting elements and the section break from the keyboard is unchanged.

**Situation.** The keys already delete the selected elements or section break. Extending them to pages would be consistent, but it risks starting a page delete by accident.

**In simple terms.** An author clicks empty space on page 2, which selects the page, and then presses Backspace out of habit. Nothing happens, not even a dialog.

**Consequences.** Keyboard routing (`keyboardDelete`) must ignore a page-only selection.

### D-2.3: Add page and Delete page are icon buttons
**Owner decision**, made at CHECKPOINT 1 and overriding the planned word buttons.

**Verdict.** Add page and Delete page are icon-only glyph buttons in the canvas controls toolbar, built the same way as Zoom, Grid, Snap, Duplicate and Delete: `tool-button`, a new `ToolGlyph`, an `aria-label` and a `data-tip` tooltip. A disabled button gives its reason in its tooltip and accessible description.

**Situation.** Control-vocabulary V2 (`14-1-one-button-vocabulary.md:407`) allows a glyph-only control only inside a segmented control. R4 pins the allowed icon-only controls to a fixed list of 11. On that basis the plan used words. The owner wants the page buttons to match the icon toolbar they sit in.

**Options considered.** Word buttons in their own group would follow V1–V3, but look out of place in an all-icon bar. The owner chose visual consistency with the existing tools.

**Consequences.**
- R4's icon-only list and `V2_CENSUS` grow by exactly these two controls, and the test pins are updated on purpose.
- Two glyphs join `toolbar-icons.tsx`.
- Because DESIGN.md's rule to state the reason next to a disabled control can't be shown inline on an icon, the reason goes in the tooltip.

**How we'd know it was wrong.** Authors can't tell the two page glyphs apart from Duplicate and Delete.

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

## D-3.1 — Selected elements move to another page together (story 3, owner, 2026-09-14)

**Question.** When several selected elements, all on one page, are dragged onto another page, do they all move there, or can only a single element change pages?

**Decision.** They all move together, keeping their spacing. A selection that spans pages keeps today's behaviour: each element stays clamped to its own page.

**Why.** SPEC.md refuses any move that splits a keep-together group across pages. If only single elements could change pages, a group could never move to another page at all.

**Example.** A signature block (a line and a name, grouped with keepTogether) sits on page 1. The author selects both and drags them onto page 3, and both land on page 3 in the same arrangement. Dragging only the line onto page 3 is refused, naming the group.

**Also decided.** The story 3 spec stays whole rather than being split. Placing and moving share one engine field and one way of detecting the page.

## D-3.2 — Paste lands on the page under the mouse (owner, 2026-09-14)

**Question.** When the author copies elements from page 1 and pastes, which page gets the copies?

**Decision.** The page whose sheet the mouse pointer is over when Ctrl/Cmd+V is pressed. This replaces story 3's "paste keeps copies on the source's page".
- **Position:** a copy pasted onto a different page keeps the original's position on its page. A paste onto the same page keeps today's small offset, so repeated pastes still stair-step.
- **Header and footer:** copies of header or footer elements stay in their band.
- **Fallbacks:** with the pointer off every page, or in a one-page document, paste behaves exactly as before. Duplicate is unchanged.

**Example.** The author selects a clause on page 1, copies it, moves the mouse over page 2 and presses Ctrl+V. The copy appears on page 2 at the same spot the clause occupies on page 1.

## D-4.1 — Screen-reader names count designed pages (story 4, owner, 2026-09-14)

**Question.** Sheet and band names count sheets ("Content on page 3 of 3"), while the visible page labels count designed pages ("Page 2"). Which numbering should the names use?

**Decision.** In documents with more than one designed page, names use designed page numbers.
- **Page and band names:** "Report page 2 of 3 with …" and "Page Header on page 2 of 3".
- **Continuation sheets:** a sheet that isn't its page's first adds ", sheet K of M", counting every sheet, so names stay unique.
- **Column position notice:** it names the sheet.
- **One designed page:** documents with a single designed page, including overflow sheets, keep today's names exactly.

**Why.** A screen reader should hear the same page number the author sees on the page label. Keeping one-page names unchanged protects the existing single-page behaviour and its tests.

**Example.** Page 1's clause table spills onto a second sheet, and page 2 follows. The three content bands are announced as "Content on page 1 of 2", "Content on page 1 of 2, sheet 2 of 3" and "Content on page 2 of 2".

**Also decided.** The story 4 spec stays whole rather than being split.
