---
title: 'Story 7.1: Break a line where the author typed a break'
type: 'feature'
created: '2026-08-30'
status: 'ready-for-dev'
baseline_commit: '98cadf7fde2dcc69c29f7e8ae01e131a054a71f3'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-context.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
warnings: ['oversized'] # ~5800 tokens: the ruling set (D-7.1.1..D-7.1.5) is dense and must be stated, not summarised
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative orientation. The contract below governs; where the two differ, the contract wins.*

Today a line break typed into a piece of text, or arriving inside the data a template is filled
with, is treated as an ordinary space. The engine may break there or may not, depending on how much
room is left on the line. A clause written as three paragraphs comes out as one run of prose, and a
paragraph gap is inexpressible.

This story makes a typed break binding. Where the author put a break, the text starts a new line,
however much width remained. Two breaks in a row give the empty line between them, so a
paragraph gap becomes expressible; a break at the very start or end gives its empty line too. A
carriage return and line feed together count as one break, never two. A break also survives inside a
value the template declared must never be split: declaring a customer's name unbreakable stops the
engine *guessing* at a break, it does not throw away a break somebody *supplied*. That distinction,
inferring versus being told, is the shape of the whole story.

It deliberately does not touch line spacing, justification, the designer's editing surface, or the
pagination model; those are later stories.

Done looks like: every existing document still renders to identical bytes on all four target
platforms, and the new behaviour is proved by documents that fail without it.

One test stays red throughout, on purpose: a coverage floor this project decided not to fill. That
red is expected and must never be repaired.

<intent-contract>

## Intent

**Problem:** A line feed in an element's text or bound data is classified as ordinary whitespace, so
the packer may decline it; a clause cannot hold more than one paragraph, a paragraph gap is
inexpressible, and a naive fix would be silently deleted inside a declared-unbreakable value.

**Approach:** Give a break opportunity a named *kind*. A whitespace run containing a line feed
yields **mandatory** opportunities — one per line feed, `\r\n` counted once — that the shared packer
always takes and that the atomic-span filter exempts. Record on each packed line the kind of break
that ended it, for Story 7.3.

## Boundaries & Constraints

**Always:**
- The exemption is keyed on the opportunity's **kind, at the atomic-span filter site**
  (`opportunity.go:170`) — never on "is this rune `\n`", never by shrinking or excluding a span.
  `atomicSpansFor` is unchanged. (D-7.1.1)
- A mandatory break consumes its whitespace run **exactly as an optional whitespace break already
  does** (AC5): the run's leading edge ends the previous line and its trailing edge begins the next.
  Only the *number* of breaks changes — one per line feed instead of one per run.
- Mandatory breaks are **separators**: *k* breaks yield *k+1* lines. `"a\n"` is two lines, the second
  empty; `"\na"` is two lines, the first empty; `"\n"` alone is two empty lines. Achieved by
  **scoping**, not special-casing: `opportunity.go:124`'s `i > 0 && j < n` guard stays exactly as-is
  for whitespace and simply does not reach mandatory breaks. (D-7.1.2)
- An empty line occupies one full `Advance`, is counted by `textBlockHeight`, and does **not** move
  `FirstBaseline` (D-2.5a / DW-15's two-model split holds identically for empty lines).
- The change lands in the **shared packer, for every caller** — text elements, both table-cell paths,
  and the canvas projection. No text-element-only flag, no second packer. (D-7.1.3)
- `Opportunity` gains a **named kind** (not a bool). `wrappedLine` gains the **named kind of break
  that ended that line**, written by this story and read by Story 7.3 / FR47. The last-line case
  stays derived at 7.3's call site and is never stored. (D-7.1.5)
- Every existing golden hashes identically on all four targets (AC6).
- `TestCorpusMeetsP6ExerciseFloors` (P6g: got 7, need >=20) is a **mandated permanent red**
  (D-000.17 / D-2.1.14 / DW-11). Never touch it, never fill it, never report it as a regression.

**Block If:**
- Any existing golden's hash changes. That is a defect, not a re-record — HALT.
- A table row splits across a page boundary at any point.
- Satisfying an AC would require changing AD-25's **rule** (as opposed to appending the clarifying
  scope clause D-000.6 authorises).
- The four `table_render.go` rounding sites named for the DW-24 amendment cannot be confirmed by
  grep at their stated lines.

**Never:**
- No line spacing, no justification, no designer/editor surface work, no change to the pagination
  model (`internal/layout/paginate.go`'s four rules and D-2.6.1's window model are inputs).
- No new diagnostic code. An over-tall cell is Story 4.6's existing `Pagination.Clipped` /
  `TABLE_ROW_CLIPPED_HEIGHT` path, unchanged.
- Do not change a lone `\r`'s behaviour — it stays an ordinary optional whitespace break.
- Do not close DW-24 (see Delivery Log obligations); do not bolt any new fixture onto a `statement-*`
  fixture (they carry a human sign-off, D-4.7.1).
- No defensive code for a collision that cannot occur (see Design Notes); no unread, unasserted field.
- Never `git add -A`, never branch, never push.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Break beats remaining width | `"a\nb"`, box wide enough for all of it | Two lines: `a`, `b` — the break is taken regardless of remaining width | No error |
| Paragraph gap | `"a\n\nb"` | Three lines: `a`, empty, `b`; the empty line occupies one `Advance` | No error |
| Trailing break | `"a\n"` | Two lines: `a`, empty | No error |
| Leading break | `"\na"` | Two lines: empty, `a` | No error |
| Break only | `"\n"` | Two empty lines; never zero lines, never a nil deref | No error |
| CRLF | `"a\r\nb"` | One break, never two | No error |
| Lone CR | `"a\rb"` | Unchanged: one ordinary optional whitespace break | No error |
| Space adjacent to break | `"a \n b"` | Two lines: `a`, `b` — the whole whitespace run is consumed, as today | No error |
| Declared-unbreakable value | data value `"first\nsecond word"` bound on an `unbreakableValues` path, box too narrow for `second word` | Breaks at the `\n`; does **not** break at the space; the second line overflows visibly | Existing clip-and-warn (AC11), no new code |
| No declared width | element with `maxWidth <= 0` and a `\n` in its value | The break is still honoured; lines are delimited by mandatory breaks alone | No error |
| Table cell | same value in a table cell | Same line count as the text element; row height grows; the row stays atomic across pages | Over-tall row → existing `TABLE_ROW_CLIPPED_HEIGHT` Warning |
| Existing corpus | every committed fixture (none contains a line feed) | Byte-identical on all four targets | Hash change = defect, HALT |

</intent-contract>

## Code Map

- `folio-go/internal/text/opportunity.go` -- `Opportunity{LineEnd, NextStart}` at **:26-29** (no kind
  today). `Opportunities` at **:94**. Whitespace rule at **:113-128**, with the `i > 0 && j < n` guard
  at **:124**. CJK rule **:130-135**; Thai rule **:137-154** (both skip positions adjacent to
  whitespace). The `add` collapse helper at **:106-111**. Collection loop **:164-174**, containing the
  `spansCover(atomic, o)` filter at **:170** — *this is the D-7.1.1 site*. `spansCover` at **:182-192**
  (strictly interior, both ends). **HAZARD:** the collection loop is `for i := 1; i < n; i++`, so
  `LineEnd == 0` is unreachable today — a leading empty line needs it collected.
- `folio-go/wrap.go` -- `wrappedLine{from, to, width}` at **:152-155** (D-7.1.5 field #2 lands here).
  `packLines` at **:179**. **THREE HAZARD SITES, all of which silently decline a mandatory break:**
  (1) the `maxWidth <= 0` early return **:183-185** returns one line spanning everything;
  (2) the "does everything that is left fit?" short-circuit **:190-194** bypasses the opportunity
  list entirely whenever the remainder fits — this is where AC1's "regardless of how much width
  remained" is won or lost; (3) the candidate loop **:198-213** skips `op.LineEnd <= start`, which
  excludes a valid leading break at position 0, and its greedy "last that fits" must be bounded by
  the first mandatory break. Only `LineEnd`/`NextStart` are read from the chosen op at **:217-234**;
  the kind is destroyed there today. `atomicSpansFor` at **:262-277** — *unchanged*.
- `folio-go/render.go:811,817` -- text-element caller. `positionSegments` at **:1307** handles
  `from == to` safely (**:1311**, **:1317**) — an empty line draws nothing. Widest-line scan **:972**.
- `folio-go/table_render.go:876,879` (body) and `:1120,1123` (footer) -- cell callers. **:877-878** is
  the "SAME packer text elements use" comment. Row height from cell line count at **:914**.
- `folio-go/page_setup.go:450,455` -- canvas-projection caller; the designer inherits the breaks here
  (Epic 5's canvas/PDF agreement). `canvasDerived` at **:516** accepts width 0.
- `folio-go/internal/layout/paginate.go:743-745` -- row atomicity: a group's page is decided once and
  later members ride along; internal-error guard at **:1108-1119**. Existing tests:
  `TestDataRowNeverSplitAcrossPages` (`table_pagination_test.go:174`),
  `TestPaginateRowGroupMovesWholeRegardlessOfAppendOrder` (`internal/layout/paginate_group_test.go:59`).
- `folio-go/text_alignment.go:84-88` -- `textBlockHeight(lines, vm)` =
  `(lines-1)*Advance + FirstBaseline + LastDescent`. An extra empty line adds exactly one `Advance`
  and cannot move `FirstBaseline` — *by construction*, which is what the direct assertion pins.
  `textValignOffset` at **:65**. Rounding sites **:56**, **:74**.
- `folio-go/internal/bind/text.go:37-46` -- `Substitution{Path, Start, End}`; populated at
  **:262-279**, where `from` is captured *after* the preceding literal is written. **This is the
  fixture trap:** the span brackets only the placeholder's own resolved text.
- Fixture precedent to copy: `fixtures/wrapped-text/` (`input.folio`, `expected.json`, `expected.pdf`,
  `README.md`), template const `folio-go/wrapped_text_template.go:23-43` kept byte-identical to
  `input.folio`, data const `wrappedTextDataJSON` at **:51**, declaring
  `"unbreakableValues": ["customer.name"]` with a deliberately narrow box on `e4`.
  Tests: `folio-go/wrapped_text_fixture_test.go` (semantic, layout, declaration-is-load-bearing,
  red-proof, golden). Registration: `goldenDigestRecord` at `folio-go/byte_neutrality_test.go:92`,
  `matrixDocuments` in `folio-go/matrix_test.go` (a new entry may defer its four CI legs; the local
  four-target gate is `TestCrossTargetByteIdentity`), and the fixture list inside
  `TestCorpusFixturesProduceNoMissingGlyphWarnings` (`folio-go/missing_glyph_corpus_test.go:29`).
- Existing test conventions: `internal/text/opportunity_test.go` — table-driven `cases` structs with
  an explicit `t.Fatalf("precondition: ...")` proving the property could fail
  (`TestAtomicSpansRemoveInteriorOpportunities` at **:190**); `folio-go/wrap_test.go:520`
  `TestAtomicSpansForIsDrivenByTheDeclarationSet`.
- Docs to amend: `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md`
  AD-25 at **:467**, third override **:482-488** (clause attaches to the sentence ending "a listed
  path." on **:483**); `_bmad-output/specs/spec-folio/folio-format.md` **:450-465** (attach to "Every
  value substituted from a listed path is kept on one line.", **:458**);
  `_bmad-output/implementation-artifacts/deferred-work.md` DW-24 at **:183-259** (prose `**Owner:**`
  style; DW-21 at **:295** is the two-addresses precedent).

## Tasks & Acceptance

**Execution:**

- `folio-go/internal/text/opportunity.go` -- Add a **named** break-kind type and field to
  `Opportunity` (optional/inferred vs mandatory), documenting that a mandatory break comes from a
  control character the caller supplied and is never inferred. Emit mandatory opportunities from the
  whitespace rule: a run containing at least one line feed (`\n`, with a preceding `\r` folded into
  the same unit) yields **one mandatory opportunity per line feed**, partitioning the run so the
  first ends the previous line at the run's start, each interior one produces a zero-length line, and
  the last begins the next line at the run's end. Such a run emits mandatory breaks *instead of* the
  single optional one. The `i > 0 && j < n` guard stays byte-for-byte as it is and simply does not
  gate the mandatory path. Let the collection loop reach `LineEnd == 0`. At **:170**, exempt
  mandatory opportunities from `spansCover` — keyed on kind, at the filter site, with the span and
  `atomicSpansFor` untouched. -- Rationale: D-7.1.1, D-7.1.2, D-7.1.5(1); AC1–AC5.
- `folio-go/wrap.go` -- Add the **named kind of break that ended this line** to `wrappedLine`, with a
  comment naming Story 7.3 / FR47 as the arriving consumer. Make `packLines` honour mandatory breaks
  at all three hazard sites named in the Code Map: the `maxWidth <= 0` path splits on mandatory
  breaks alone; the "everything left fits" short-circuit only fires when no mandatory break remains;
  the candidate set is bounded by the first mandatory break at or after `start` (a mandatory break at
  `start` itself is valid progress, since its `NextStart` advances), leaving the existing greedy
  "last that fits" and the AC11 overflow rule otherwise untouched. Emit a final zero-length line when
  the input ends on a mandatory break. -- Rationale: D-7.1.2, D-7.1.3, D-7.1.5(2); AC1, AC2.
- `folio-go/internal/text/opportunity_test.go` -- Add table-driven coverage for: break beats
  remaining width; `\n\n`; leading and trailing `\n`; `"\n"` alone; `\r\n` as one break; lone `\r`
  unchanged; whitespace adjacent to a break consumed as today. Add the **span-exemption** case: an
  atomic span covering a value holding **both** a `\n` and a space — assert the mandatory opportunity
  survives and the space's does not, with a `t.Fatalf("precondition: ...")` proving the span strictly
  contains the line feed's index. -- Rationale: red-proof by flipping the kind test, not by deleting
  the span (D-7.1.1).
- `folio-go/wrap_test.go` -- Assert the `wrappedLine` break-kind field **directly** over fabricated
  input, commenting that Story 7.3 / FR47 is the arriving consumer (precedent:
  `verticalMetrics.LastDescent`). Assert line counts for the empty-line cases and that
  `textBlockHeight` counts an empty line as one `Advance` while `FirstBaseline` is unmoved. Assert a
  `maxWidth <= 0` element honours a break. Add the chain-level discriminating case: build spans via
  `atomicSpansFor` from a declared path whose **data** value contains a `\n` and a space, with the
  same precondition assertion that the resulting span strictly contains the line feed. -- Rationale:
  D-7.1.2, D-7.1.5; the fixture trap.
- `folio-go/table_render_test.go` (or the nearest existing table test file) -- Assert a text element
  and a table cell holding the same `\n`-bearing value produce the **same** line count, with one side
  pinned to a literal expected number so the two live computations cannot drift together. Re-assert
  row atomicity with a `\n`-bearing cell: the row must never split across a page boundary. --
  Rationale: D-7.1.3.
- `fixtures/mandatory-break/` + `folio-go/mandatory_break_template.go` +
  `folio-go/mandatory_break_fixture_test.go` -- New golden fixture on the `wrapped-text` pattern: a
  template declaring `unbreakableValues` for a data path, an element binding that path in a
  deliberately narrow box, and a data const whose value contains **both** a line feed and a space, so
  the fixture red-proves both directions. Keep `input.folio` byte-identical to the Go const. Register
  in `goldenDigestRecord`, in `matrixDocuments`, and in
  `TestCorpusFixturesProduceNoMissingGlyphWarnings`'s fixture list. -- Rationale: the epic's "every
  feature ships its golden fixture"; AC6's guard is only falsifiable with a `\n`-bearing document.
- `ARCHITECTURE-SPINE.md` (AD-25, third override) and `folio-format.md` (`unbreakableValues` prose)
  -- Append the same scope clause to both: the constraint binds **inferred** break opportunities, not
  literal control characters present in the input. **State the clause; do not change the rule.** Note
  that D-2.1.6's disclosed cost (a Thai name in free-form literal text is still guarded only by the
  atomic-unknown-run rule) is untouched. -- Rationale: D-000.6 clarifying edit, in this story's commit.
- `_bmad-output/implementation-artifacts/deferred-work.md` (DW-24) -- Amend in place, in DW-24's own
  prose `**Owner:**` style: owner becomes **Story 7.3**, *and* the orchestrator's own gate checklist
  as a second standing address (DW-21's precedent) — explicitly **not** "Epic 7 close", because an
  owner that is an event stops existing when the event passes (DW-14 / D-000.73). Add the four
  `table_render.go` rounding sites `:687`, `:698`, `:1017`, `:1193` to the enumeration alongside
  `text_alignment.go:56`/`:74`, noting that a centred *text* element does not cover the table sites.
  Record that **at closure the enumeration must be re-derived by grep and recorded, never read off
  this hand-list** — the list is being amended today precisely because it rotted once. Do **not**
  close the entry. -- Rationale: D-7.1.4.

**Acceptance Criteria:**
- Given the whole committed fixture corpus, when it is re-rendered, then every existing golden hash
  is unchanged and `TestCrossTargetByteIdentity` passes; the four statement digests remain
  `statement-1` 76,744 `114df1d6…`, `-5` 127,363 `70dce051…`, `-20` 269,884 `56bfbbd9…`,
  `-50` 555,829 `5d090b0f…`.
- Given the new fixture, when the kind test at the atomic-span filter site is flipped, then the
  fixture goes red for the line feed; and when the atomic span is instead removed, then it goes red
  for the space — the two failures are distinguishable.
- Given the packer is forced to treat a mandatory opportunity as optional, when a `\n`-bearing value
  is packed, then the break is declined and a test fails — the kind field is load-bearing, not decorative.
- Given a `\n`-bearing table cell, when the document paginates, then its row is never split across a
  page boundary and no new diagnostic code exists in the registry.
- Given the amended AD-25 and `folio-format.md`, when they are read, then the scope clause is present
  and the surrounding rule text is otherwise unchanged.
- Given DW-24, when it is read after this story, then it names Story 7.3 plus the gate checklist,
  enumerates all six rounding sites, carries the re-derive-by-grep instruction, and is still open.

## Spec Change Log

## Review Triage Log

## Design Notes

**Why a whitespace *run* still consumes as a whole.** AC5 says the break character is consumed
"exactly as a whitespace break already is", and today a whitespace break consumes its entire run
(`Opportunity`'s doc comment at `opportunity.go:16-25`: modelling the consumed range is what keeps
"the trailing space does not count toward the line's width" a property of the break). Consuming only
the `\n` would leave a trailing space measured into the previous line and a leading space drawn on
the next — a new consumption polarity nothing asked for. So the mandatory rule inherits rule 1's
consumption at the run's outer edges and only changes the *count* of breaks inside it. `"a \n b"`
therefore gives `a` / `b`, exactly as `"a b"` would if it broke there.

**Worked partition.** `"a \n\n b"`, runes `a`(0) `␠`(1) `\n`(2) `\n`(3) `␠`(4) `b`(5); run is [1,5)
with two line feeds. Breaks: `{LineEnd:1, NextStart:3}` and `{LineEnd:3, NextStart:5}`. Lines: [0,1)
`a`, [3,3) empty, [5,6) `b`.

**Why no defensive precedence guard in `add`.** A mandatory break could only be displaced by another
opportunity at the same `LineEnd`. Rule 2 requires both neighbours to be Han/kana and rule 3 skips
any position adjacent to whitespace, so neither can propose an opportunity at a position inside or at
the head of a whitespace run; and a line-feed-bearing run emits mandatory breaks *instead of* rule 1's
optional one. The collision is unreachable by construction, so the story asserts the outcome rather
than adding a guard for a state no input can produce.

**`\r\n`.** Folded at detection: a `\n` preceded by `\r` is one line-feed unit. A lone `\r` contains
no `\n`, so its run stays optional — unchanged, as settled.

## Verification

7.1's correctness is byte-identity-shaped, so it carries the heavy tests despite the per-epic
cadence (D-R7.1). Report measured pass/fail counts, never "green".

**Commands:**
- `cd folio-go && go test ./...` -- expected: **exactly ONE** failure,
  `TestCorpusMeetsP6ExerciseFloors` (P6g got 7, need >=20), the mandated permanent red. Anything else
  red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- expected: run once per leg
  with `FOLIO_MATRIX_TARGET` set (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`; the list is
  `matrix_test.go`'s `matrixTargets`). **Unset, this test logs "asserts NOTHING" and a no-op is not a
  pass** — name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: pass; this is
  the all-four-in-one-process local gate.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: pass (4
  pre-existing `only-export-components` lint warnings are not a regression).

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` fails under
`-tags=matrix` because `fontTools` is not installed here; `lint/internal/rules/licencegraph_test.go`
is not gofmt-clean (DW-23, owned by Story 15.2).

**Delivery Log obligations (for the closer):**
- DW-24 was **inspected and declined** by 7.1 on the criterion, not the budget: 7.1 touches neither
  the rounding nor the population that reaches it. Its text was amended here; its owner is Story 7.3
  plus the gate checklist.
- `TEXT_MISSING_GLYPH` stops firing for a literal `\n` once the character is consumed. The new
  `\n`-bearing fixture is therefore the **first** document under
  `TestCorpusFixturesProduceNoMissingGlyphWarnings` to pass for a reason unrelated to glyph coverage —
  that green must not be read as coverage.
- Story 7.3 (FR47) **depends on** the `wrappedLine` break-kind field landed here. D-R7.3's numeric
  order satisfies the dependency only by accident; the dependency is stated so a reorder cannot
  silently break it.

## Auto Run Result

Status: ready-for-dev
Blocking condition: none — halted after planning, as the dispatch directed.

Dispatch 2 for Story 7.1 (classic intent, baseline `98cadf7fde2dcc69c29f7e8ae01e131a054a71f3`).
Dispatch 1's three intent-gap forks arrived pre-ruled (D-7.1.1 … D-7.1.5) and are applied above
without re-opening. No code was written and no commit was made. No verification command was run:
`Halt after planning.` stops the workflow before step-03, so `## Verification` is a plan, not a
measurement.
