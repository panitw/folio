---
title: 'Story 7.1: Break a line where the author typed a break'
type: 'feature'
created: '2026-08-30'
status: 'done'
baseline_commit: '98cadf7fde2dcc69c29f7e8ae01e131a054a71f3'
baseline_revision: '57a4f8eb0a8ce1f24c3a8169172011a8939f73e4'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-context.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
warnings: ['oversized'] # ~5800 tokens: the ruling set (D-7.1.1..D-7.1.5) is dense and must be stated, not summarised
deferred:
  - summary: >-
      A canvas text element whose value carries 256 or more typed breaks now makes the whole
      canvas projection return a hard error, a failure mode that was unreachable before Story 7.1.
    evidence: |-
      page_setup.go's maxCanvasTextLines = 256 guard returns an error rather than degrading. Before
      7.1 a canvas text element's line count was bounded by wrapping, and was exactly 1 when width
      was unset, so the guard could not be reached from an element's own value. Typed breaks now
      set the line count directly. Story 7.1's intent contract forbids designer/editor surface work
      and forbids new diagnostic code, so clamping-and-warning is work this story cannot authorise;
      recorded for whoever owns the canvas surface next.
    location: >-
      folio-go/page_setup.go:27,456
    tracked_as: 'DW-25'
    severity: medium # filed low by the implementer; raised at closure and routed to the engineering lead
---

## In plain terms (read this first if you just want the gist)

*Non-normative orientation. The contract below governs; where the two differ, the contract wins.*

A line break typed into text, or arriving in the data a template is filled with, was treated as an
ordinary space — taken only if the line happened to be full. It now binds. Where the author put a
break, the text starts a new line however much room remained. Two breaks in a row give the empty line
between them, so a paragraph gap is finally expressible; a break at the start or the end gives its
empty line too, and a carriage return with a line feed counts once. A break also survives inside a
value the template declared must never be split: declaring a value unbreakable stops the engine
guessing at a break, it does not throw away one somebody supplied.

Every document already committed still renders to identical bytes on all four target platforms, and a
new one proves the behaviour by failing without it.

Two things will look wrong later and are not. One test stays red on purpose — a coverage
floor this project decided not to fill, which must never be repaired. And the new document clears the
font-coverage check for a reason unrelated to fonts: once a typed break is consumed, the character
stops being reported as uncovered.

Line spacing, justification, the designer's editing surface and the pagination model were left alone
deliberately. One thing is recorded rather than fixed: enough typed breaks in text on the design
canvas can now fail the whole canvas preview rather than just that item.

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

Two things the implementation had to settle that the task list did not name. Both are recorded here
rather than absorbed silently, because each changes something outside this story's enumerated edits.

**1. `TEXT_MISSING_GLYPH` no longer fires for U+000A, and it had to be MADE to stop.**
The Delivery Log obligation states "`TEXT_MISSING_GLYPH` stops firing for a literal `\n` once the
character is consumed". Measured, it did not: `shapeSegments` resolves every rune of the element's
text BEFORE the packer ever sees an opportunity, so no face covered U+000A, and each of the new
fixture's three elements returned a Warning saying the font chain was incomplete. That is a false
statement about the fonts, and it made the spec-mandated registration in
`TestCorpusFixturesProduceNoMissingGlyphWarnings` (which requires zero) unsatisfiable.

The fix is one guard in `render.go`'s `shapeSegments`, keyed on a new `lineBreakHandling` parameter
the CALLER supplies: where the caller hands the segments to `packLines` (`breaksAreConsumed` — the
text element, both table-cell paths, the canvas projection), a U+000A emits no missing-glyph
Diagnostic; where it does not (`breaksAreDrawn`), the Warning stands.

**The parameter is not decoration.** The suppression's justifying premise is "the breaker consumes
it", and that premise is FALSE on one production caller: `table_render.go:668` shapes a table COLUMN
LABEL and hands the whole rune range straight to `positionSegments`, never packing it. A line feed in
a label really is dropped — two words run together on one baseline — and the Warning is the only
signal it has. A global suppression would have removed that signal on the one path where it is true.
Caught in review; pinned by `TestLineFeedInAColumnLabelStillWarns`, which also pins that a label
stays on ONE baseline, because the intent contract's caller enumeration is closed and a label is not
a cell.

Scoped to the diagnostic only — the SEGMENTATION is untouched on every caller, so the rune still
claims its slot in the element-global rune index space, which is what keeps `internal/text`'s break
positions meaningful. A lone `\r` is deliberately NOT included: it gains no meaning in this story,
so nothing about it changes.

**2. The missing-glyph corpus table's coupling to `baselineAcceptanceFixtures` was widened from
EQUALITY to "superset, with the extras declared".**
`TestCorpusFixturesProduceNoMissingGlyphWarnings` asserted its fixture table and
`baselineAcceptanceFixtures` name the same set, and `baselineAcceptanceFixtures` is Story 2.5a's list
of THE FIVE GOLDENS THAT STORY RE-RECORDED — its own
`TestFirstBaselineSemanticAcceptanceAcrossEveryReRecordedGolden` asserts `len == 5`, and every entry
carries hand-derived baseline arithmetic for that re-recording. Adding `mandatory-break` there would
have been a false statement about what 2.5a did; not adding it made the spec-mandated registration
impossible. The relation is now: every `baselineAcceptanceFixtures` entry must appear in the corpus
table (unchanged — that is the direction the original Finding 11 was about), and every corpus entry
that is not one of them must be DECLARED in a new one-line `beyondBaselineAcceptance` map naming the
story that added it. A swap in either table still reddens, and so does an undeclared addition.

**Not a change, but worth recording: `Block If` #4 was checked and cleared.** All four
`table_render.go` rounding sites were confirmed by grep at their stated lines at baseline `98cadf7`:
`:687` (header `align: center`), `:698` (header `valign: middle`), `:1017` (body `align: center`),
`:1193` (footer `align: center`). The same grep also returns `render.go:482`/`:483`, which centre an
IMAGE in its box — a different subject from DW-24's text alignment, recorded in DW-24 only so the
next re-derivation is not surprised by them.

## Review Triage Log

Adversarial review of the delivered diff: 0 intent_gap, 0 bad_spec, 13 patch findings. All 13
applied; nothing inside `<intent-contract>` was edited, no existing golden hash moved, and
`TestCorpusMeetsP6ExerciseFloors` was not touched.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| P1 | HIGH | The U+000A missing-glyph suppression was global, but its premise is false on the table column-label path, which shapes without packing — the one place the rune really is dropped lost its only signal. | FIXED. `shapeSegments` now takes a `lineBreakHandling` parameter; the label site passes `breaksAreDrawn` and keeps warning. `TestLineFeedInAColumnLabelStillWarns` pins the Warning **and** that a label stays on one baseline (labels deliberately do not break — the contract's caller enumeration is closed). |
| P2 | MED | `firstMandatoryFrom` rescanned `ops` from 0 per line, for every document including the ~100% of the corpus with no line feed. | FIXED. Replaced by a cursor carried across iterations, sound because `ops` is ascending and `start` is monotonic. `TestPackLinesTakesEveryTypedBreakAtEveryBoxWidth` is the new invariant that would catch a cursor that loses the *second* break — which every hand-written case would still pass. Byte-identical. |
| P3 | MED | No test exercised the canvas projection, a caller the contract names, with a `\n`. | FIXED. `TestCanvasTextPaintHonoursATypedBreak`, deliberately on a NO-declared-box element (the canvas's own instance of the `maxWidth <= 0` hazard), with a space-instead-of-break negative control. |
| P4 | MED | Only the body table-cell path had `\n` coverage; no trailing-`\n` row-height case. | FIXED, and one half **corrected**: a footer cell's text is *always* an aggregate rendered through `footerFormat`, whose pattern is a closed `#`/`0`/`,`/`.` grammar, so a line feed cannot reach it. `TestFooterCellTextCannotCarryATypedBreak` asserts that by construction and is the tripwire if the grammar is ever widened. `TestTrailingBreakInACellGrowsTheRowByOneAdvance` covers the trailing case — and asserts that both documents emit the same number of line RUNS, which is exactly why a run-counting test is blind to it. |
| P5 | MED | `folio-format.md`'s Line breaking section still claimed a line may end after whitespace "and nowhere else", under a preamble promising "the **whole** rule". | FIXED. Split into *Inferred breaks* (the three script rules, unchanged) and a new *Mandatory breaks* section stating the separator model, the empty line's full Advance, CRLF folding, run consumption, and the `unbreakableValues` interaction. AD-25's rule and the `unbreakableValues` prose were left as already edited. |
| P6 | LOW | The Design Notes decline a precedence guard in `add` on the grounds that the story "asserts the outcome"; no such assertion existed. | FIXED. `TestMandatoryBreakIsNeverDisplacedByAnotherRule` puts a line feed hard against Han, kana and Thai, with a precondition proving the other rules are live at those positions (measured on the same string with its feeds turned into spaces). No guard added. |
| P7 | LOW | `beyondBaselineAcceptance`'s values were never read, so `{"foo": ""}` satisfied the documented convention. | FIXED — each value must be non-empty. |
| P8 | LOW | `requireMandatoryBreakIsBroken` asserted a baseline TOTAL and a gap at hardcoded indices, and duplicated the advance literal. | FIXED. The baselines are declared once, per element, in `mandatoryBreakBaselinesMP`, and both the ordinary suite and every matrix leg run the same `mandatoryBreakAssertBaselines` — element identity plus each element's own interval. The fixture test now also compares e2's interval to e1's rather than to a literal, so a uniform advance change cannot satisfy it. |
| P9 | LOW | DW-24's heading, closure sentence and Trigger block all contradicted the six-site body. | FIXED, entry left OPEN. Heading names six sites; the closure paragraph now says one document must reach every site the re-derived grep returns; the Trigger records that it already fired once (7.1 changed what feeds `textBlockHeight`) and the item was still declined on the criterion. Added *why* the pairs are uneven — verified in code: a body cell's valign is distributed in whole line slots (no remainder) and a footer cell has no vertical slack at all, so four `align` + two `valign` is the complete set. |
| P10 | LOW | Comments said the four CI legs were "DEFERRED" while the workflow wires them under `if-no-files-found: error`. | FIXED. Both comments now state that CI produces and compares all four legs, and that this story additionally ran them. |
| P11 | LOW | `multiRowTableDataWithBreak` panicked on a marshal error. | FIXED — takes `*testing.T`, calls `t.Fatalf`. |
| P12 | LOW | The `maxWidth <= 0` test asserted only line count and text. | FIXED — same `from == to && width == 0` guard as the packing path, plus the `endedBy` kinds. |
| P13 | LOW | No coverage of a typed break overflowing an element's HEIGHT. | FIXED, with a **correction**: there is no `TEXT_CLIPPED_HEIGHT` code — D-2.8.1 rules that a text element's declared height is not a clip bound. The reachable condition is Story 4.6's `TABLE_ROW_CLIPPED_HEIGHT`, which the contract's own I/O matrix names. `TestRowMadeTooTallByTypedBreaksIsTheExistingClipAndWarn` reaches it with typed breaks alone at an ordinary 8 pt, with a no-breaks negative control, and asserts no code outside the existing two is minted. |

### 2026-08-30 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 13: (high 1, medium 4, low 8)
- defer: 1: (high 0, medium 0, low 1)
- reject: 9: (high 0, medium 0, low 9)
- addressed_findings:
  - `[high]` `[patch]` The U+000A `TEXT_MISSING_GLYPH` suppression was global to `shapeSegments`, but its justifying premise ("the breaker consumes it") is false on `table_render.go:668`, the one production caller that shapes a column label without ever packing it — a `\n` there was dropped with NO diagnostic where it previously warned. Fixed by threading a `lineBreakHandling` parameter; the label site and the digit table pass `breaksAreDrawn` and keep warning, the five packing sites pass `breaksAreConsumed`. `TestLineFeedInAColumnLabelStillWarns` pins both the Warning and that a label stays on one baseline.
  - `[medium]` `[patch]` `firstMandatoryFrom` rescanned `ops` from index 0 once per line, making `packLines` O(lines x ops) for every document including the whole line-feed-free corpus. Replaced with a carried cursor; `TestPackLinesTakesEveryTypedBreakAtEveryBoxWidth` (24 input/box combinations) is the invariant that catches a cursor losing a later break.
  - `[medium]` `[patch]` The canvas projection, a caller the contract names explicitly, had no `\n` test. Added `TestCanvasTextPaintHonoursATypedBreak` on a no-declared-box element with a space-instead-of-break control.
  - `[medium]` `[patch]` Table-cell coverage was body-path only and used interior breaks only. Added `TestTrailingBreakInACellGrowsTheRowByOneAdvance` (+10,896 mp = exactly one Advance, and both documents emit the same run count, which is why run-counting tests are blind to it). The footer half was corrected on the facts: a footer cell's text is always an aggregate rendered through `footerFormat`'s closed `#`/`0`/`,`/`.` grammar, so a line feed cannot reach it; `TestFooterCellTextCannotCarryATypedBreak` pins that unreachability as a tripwire.
  - `[medium]` `[patch]` `folio-format.md`'s primary Line breaking section still stated the Latin rule as "a line may end after a run of whitespace, and nowhere else" under a preamble promising "the whole rule" — false from FR46 on. Split into Inferred breaks (unchanged) and a new Mandatory breaks section. AD-25's rule text untouched.
  - `[low]` `[patch]` The Design Notes decline a precedence guard in `add` because the story "asserts the outcome"; no such assertion existed. Added `TestMandatoryBreakIsNeverDisplacedByAnotherRule` (line feed against Han, kana and Thai, with a precondition proving the other rules are live there). No guard added.
  - `[low]` `[patch]` `beyondBaselineAcceptance`'s values were never read, so the documented convention was unenforced. Each value must now be non-empty.
  - `[low]` `[patch]` `requireMandatoryBreakIsBroken` asserted a baseline total and a gap at hardcoded indices without establishing element identity, and duplicated the advance literal. Baselines now declared once per element in `mandatoryBreakBaselinesMP`, shared by the ordinary suite and every matrix leg.
  - `[low]` `[patch]` DW-24's heading, closure sentence and Trigger block all contradicted the six-site body it gained. All three reconciled, the uneven align/valign pairs explained from verified code, entry left OPEN.
  - `[low]` `[patch]` Comments claimed the new fixture's four CI legs were "DEFERRED" while `.github/workflows/matrix.yml` wires them under `if-no-files-found: error`. Corrected to state CI runs all four and that this story ran them too.
  - `[low]` `[patch]` `multiRowTableDataWithBreak` panicked on a marshal error instead of taking `*testing.T`.
  - `[low]` `[patch]` `packMandatoryOnly`'s line widths were never asserted; it now carries the same `from == to && width == 0` guard as the packing path.
  - `[low]` `[patch]` No coverage of a typed break overflowing a declared height. Corrected on the facts — there is no `TEXT_CLIPPED_HEIGHT` (D-2.8.1: a text element's declared height is not a clip bound); the reachable condition is Story 4.6's `TABLE_ROW_CLIPPED_HEIGHT`, which the I/O matrix names. `TestRowMadeTooTallByTypedBreaksIsTheExistingClipAndWarn` reaches it with typed breaks alone, with a negative control, minting no new code.

**Deferred (1).** The canvas projection's `maxCanvasTextLines = 256` guard is newly reachable from an element's own value; recorded in frontmatter `deferred`, not fixed, because the contract forbids designer-surface work and new diagnostic code.

**Rejected (9), each on the authority of the intent contract.** Making U+2028/U+2029/U+000B/U+000C/U+0085 mandatory (the contract scopes to the line feed, with `\r\n` folded and a lone `\r` unchanged); adding a precedence guard to `add` (the Never clause forbids defensive code for a collision that cannot occur — the outcome is asserted instead, see the patch above); changing whitespace-run consumption so a NBSP adjacent to a break survives (D-7.1.6 fixes whole-run consumption and forbids a new consumption model); a no-progress guard in `packLines` (a mandatory opportunity's `NextStart` is strictly greater than its `LineEnd` by construction, so the loop cannot spin); the `if` -> `case` conversion of the `i > 0 && j < n` guard (the expression and its governance are preserved; only the statement form changed); rewording the doc scope clause away from "control characters" (that phrasing is what the task list dictated); clamping the canvas 256-line guard instead of erroring (deferred above rather than redesigned here); the `internal/text` package binary reporting FAIL and thereby masking package-level signal (that is the mandated permanent P6g red); and the observation that the designer's single-line input still cannot type a break (the Never clause excludes designer/editor surface work).

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


---

### Dispatch 3 — implement, review, commit (2026-08-30)

Baseline `57a4f8eb0a8ce1f24c3a8169172011a8939f73e4`. Terminal status: **done**.

**Implemented change.** A break opportunity now carries a named `text.BreakKind`. A whitespace run
holding at least one line feed emits one **mandatory** opportunity per line feed *instead of* the
single optional one, partitioned so k breaks yield k+1 lines; `\r\n` folds by counting only U+000A,
and a lone `\r` is untouched. The atomic-span filter exempts mandatory opportunities **by kind, at
the filter site** — `spansCover` and `atomicSpansFor` are unchanged. The collection loop was widened
from 1 to 0 so a leading break is reachable. `packLines` honours mandatory breaks at all three hazard
sites, and `wrappedLine` records the kind of break that ended it for Story 7.3 / FR47.

**Files changed.**
- `folio-go/internal/text/opportunity.go` — `BreakKind`, `Opportunity.Kind`, the mandatory partition, the kind-keyed exemption, collection from index 0.
- `folio-go/wrap.go` — `wrappedLine.endedBy`, `packMandatoryOnly`, the three hazard sites, cursor-based `firstMandatoryFrom`.
- `folio-go/render.go` — `shapeSegments` gains a `lineBreakHandling` parameter; U+000A emits no missing-glyph Diagnostic only where the caller actually consumes it.
- `folio-go/table_render.go`, `page_setup.go`, `page_number.go` — pass the new parameter; no packer changes (they already share it, which is D-7.1.3's point).
- `fixtures/mandatory-break/` + `mandatory_break_template.go` + `mandatory_break_fixture_test.go` — the new golden; registered in `goldenDigestRecord`, `matrixDocuments`, `declaredEpic2GateObligations`, the missing-glyph corpus, and `.github/workflows/matrix.yml`.
- `missing_glyph_corpus_test.go` — corpus/`baselineAcceptanceFixtures` relation widened from equality to superset-with-declared-extras.
- Tests in `opportunity_test.go`, `wrap_test.go`, `table_render_test.go`, `table_pagination_test.go`, `render_test.go`, `canvas_text_paint_test.go`, `matrix_test.go`, `byte_neutrality_test.go`.
- Docs: `ARCHITECTURE-SPINE.md` (AD-25 third override), `folio-format.md` (scope clause + new Mandatory breaks section), `deferred-work.md` (DW-24 amended, left OPEN).

**Review findings.** 13 patches applied (1 high, 4 medium, 8 low), 1 deferred, 9 rejected, 0 intent_gap, 0 bad_spec. Follow-up review recommended: **true** (a high-severity patched finding).

**Verification measured after the patches.**
- `cd folio-go && go test -count=1 ./...` — 13 packages ok; **exactly one** failure, `TestCorpusMeetsP6ExerciseFloors/P6g` (got 7, need >=20), the mandated permanent red. Untouched.
- `go vet -tags=matrix ./...` — clean. `gofmt -l .` — empty.
- `TestTargetRenderHash` ran on **all four legs** with `FOLIO_MATRIX_TARGET` set each time (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) — PASS on each, no "asserts NOTHING". All 17 documents byte-identical across the four.
- Golden digests unchanged: `statement-1` 76,744 `114df1d6…`; `-5` 127,363 `70dce051…`; `-20` 269,884 `56bfbbd9…`; `-50` 555,829 `5d090b0f…`. New `mandatory-break` 56,681 `7cf743de…`, identical on all four targets.
- `go test -tags=matrix -run TestCrossTargetByteIdentity .` — ok.
- `cd lint && go test -count=1 ./...` — 4 packages ok.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck ok; 4 pre-existing `only-export-components` warnings; 30 files / 213 tests pass.

**Residual risks.**
- The canvas `maxCanvasTextLines` guard is newly reachable from an element's own value (see frontmatter `deferred`).
- `TestShippedFacesReproduceFromUpstream` was not exercised: the full `-tags=matrix` suite was not run, only the two named matrix tests. It is known-environmental here (no `fontTools`).
- `mandatory-break` is the first document under `TestCorpusFixturesProduceNoMissingGlyphWarnings` to pass for a reason unrelated to glyph coverage; that green must not be read as coverage.
- Story 7.3 (FR47) depends on `wrappedLine.endedBy` landed here.

## Delivery Log

### 2026-08-30 — planned

Baseline `98cadf7`. Dispatch 1 halted `blocked` on three intent gaps the spec could not settle:
whether the atomic-span exemption is keyed on the rune or on the opportunity's kind, whether k breaks
yield k or k+1 lines, and which packer the change lands in. The engineering lead ruled all three
before any code existed — D-7.1.1 (exemption **by kind, at the filter site**, with `atomicSpansFor`
and `spansCover` untouched, so AD-25's rule is clarified rather than amended), D-7.1.2 (breaks are
**separators**, achieved by scoping rather than special-casing), D-7.1.3 (the **shared** packer, every
caller, no text-element-only flag) — with D-7.1.4 amending DW-24's scope and D-7.1.5 authorising the
two named fields. D-R7.1 put this story on the heavy tests despite Epic 7's per-epic cadence, because
its correctness is byte-identity-shaped. Dispatch 2 planned against the rulings and halted, as
directed, without writing code. Committed `62a6af1`; status flipped at `57a4f8e`.

### 2026-08-30 — built

Baseline `57a4f8e`, delivered at `33bd942` (28 files, +2850/−71). D-7.1.6 fixed whole-run consumption
and D-7.1.7 settled the two things the task list had not named: collect opportunities from index 0
(a loop starting at 1 silently drops a leading break while every trailing-break test passes), and
make the AC1 fixture element **fit** its box so it exercises the packer's short-circuit rather than a
break taken for want of room.

Review: 0 intent_gap, 0 bad_spec, 13 patched (1 high, 4 medium, 8 low), 1 deferred, 9 rejected. The
high finding is the one that mattered: the first implementation suppressed the U+000A missing-glyph
warning **globally**, and that suppression's premise — "the breaker consumes it" — is false on the
table column-label path, which shapes and positions without ever packing. A line feed there really is
dropped, and the warning was its only signal. The implementer also **corrected two of the reviewer's
own patch instructions on the facts**: a footer cell's text is always an aggregate through a closed
numeric format grammar, so a line feed cannot reach it (pinned as a tripwire rather than tested as a
behaviour); and there is no `TEXT_CLIPPED_HEIGHT` to overflow, D-2.8.1 having ruled that a text
element's declared height is not a clip bound. Two spec amendments were recorded rather than absorbed
silently, both in `## Spec Change Log`.

### 2026-08-30 — done

Baseline `57a4f8e`, story commit `33bd942`, closed on `main`. `followup_review_recommended` was
**cleared to false**: rather than a second builder review round, the high finding and the churn around
it were given an independent adversarial pass at closure, and the six things that could have been
wrong were each checked against the diff rather than against the build's report.

**The caller enumeration is complete.** `shapeSegments` has six production callers and three in
tests, all nine derived by grep and then checked against what each caller actually *does* rather than
against what its argument claims. The two `breaksAreDrawn` sites genuinely do not pack — the column
label hands its whole rune range to `positionSegments`, and the page-number digit table shapes the
literal `0123456789`, which no line feed can reach. The four `breaksAreConsumed` production sites —
the text element, both table-cell paths and the canvas projection — each call `packLines` a few lines
below. No caller is missed, so there is no path where a line feed is dropped without a diagnostic.

**Teeth, measured.** Neutering the kind-keyed exemption at the filter site reddens **five** distinct
tests: the unit case, the chain-level case, the fixture's layout and declaration-is-load-bearing
cases, and the golden hash. It does not redden for the wrong reason — with the exemption gone the
surviving opportunity list is `[{LineEnd:3 NextStart:4 Kind:0}]`, which is the space inside the
declared value still correctly suppressed, so the two failure directions really are distinguishable.
The fixture routes its line feed through **data** on a declared path, never through literal template
text, and both the unit and fixture tests carry `t.Fatalf` preconditions proving the feed is
*strictly interior* to the span — the exemption is not proved by an input that never needed it. The
`wrappedLine` break-kind field is asserted directly over fabricated input, naming Story 7.3 / FR47,
with a discrimination check that the typed and inferred kinds differ; it is not an unread field.
`internal/layout/paginate.go` is absent from the diff, and no new diagnostic code is minted — the
diff references only the three existing codes.

**One defect found and fixed at closure**, in this story's own test file: when the load-bearing
assertion regressed, `TestMandatoryBreakSemanticAcceptance` **panicked** on a six-way index after
reporting the wrong baseline count non-fatally, and a panic takes the whole `folio-go` test binary
down with it — every other test in the package stops reporting. That is DW-23's shape in miniature,
one signal swallowing another, and it would have turned a legible regression into an opaque one. A
length guard now stops at the report; re-measured under the same neutering, all five tests fail
individually and the rest of the package keeps reporting. Spot-checks of the nine rejections found
them sound, including the two worth doubting: the `if` → `case` conversion preserves the `i > 0 && j
< n` expression byte-for-byte and only changes the statement form, and the declined no-progress guard
is genuinely unreachable, since every mandatory opportunity's `NextStart` is strictly greater than its
`LineEnd` by construction.

**Gates measured at closure, independently, not carried forward.** `go test -count=1 ./...`: 13
packages ok, exactly one failure — `TestCorpusMeetsP6ExerciseFloors/P6g`, got 7 need >=20, the
mandated permanent red (D-000.17 / D-2.1.14 / DW-11), untouched. `go vet -tags=matrix ./...` clean;
`gofmt -l` empty. `TestTargetRenderHash` PASS on **all four legs** with `FOLIO_MATRIX_TARGET` set
each time (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) — no leg logged "asserts
NOTHING". `TestCrossTargetByteIdentity` PASS. `lint`: 4 packages ok. Designer: typecheck ok, 4
pre-existing `only-export-components` warnings, 30 files / 213 tests pass. Corpus neutrality was
confirmed twice over — the four statement goldens hash identically across all four legs **and** are
byte-identical to their pre-story bytes at `57a4f8e`: `statement-1` 76,744 `114df1d6…`, `-5` 127,363
`70dce051…`, `-20` 269,884 `56bfbbd9…`, `-50` 555,829 `5d090b0f…`. New `mandatory-break` 56,681
`7cf743de…`, identical on all four. Not run: `TestShippedFacesReproduceFromUpstream` (no `fontTools`
in this environment) and `lint/internal/rules/licencegraph_test.go`'s gofmt break (DW-23, owned by
Story 15.2) — both known-environmental.

**That the new fixture is green under `TestCorpusFixturesProduceNoMissingGlyphWarnings` is not a
statement about glyph coverage**, and must never be read as one. `TEXT_MISSING_GLYPH` stops firing
for a literal line feed precisely *because* the character is consumed by the breaker; this is the
first document in the corpus to pass that test for a reason unrelated to whether any face covers its
runes. A later reader treating that green as coverage evidence will be wrong.

**Deferrals.** DW-24 was **inspected and declined on the criterion rather than on the budget**: its
hazard is the unexercised *rounding* branch, and 7.1 touches neither the rounding nor the population
that reaches it — no corpus document declares `center` or `valign`, and this story adds none — so
closing it here would have discharged nothing 7.1 endangered. Its scope was amended instead (D-7.1.4)
and verified at closure: owner **Story 7.3 plus the orchestrator's gate checklist**, explicitly not
"Epic 7 close" since an owner that is an event stops existing when the event passes; six rounding
sites, not two; and the re-derive-by-grep-at-closure instruction present. The six sites were
re-derived by grep at `33bd942` — not merely at the baseline the entry cites — and all six answer at
their stated lines. The entry is OPEN.

**DW-25 is new**, and is the canvas line cap: `maxCanvasTextLines = 256` was unreachable from an
element's own value before this story, because a canvas element's line count was bounded by wrapping
and was exactly one when width was unset. Typed breaks set it directly, so the bound is now reachable
by pasting a long clause — and it returns an error that aborts the **whole** canvas projection rather
than degrading that one element. Filed at **medium** (the implementer's `low` was overridden at
closure on blast radius), owner **UNASSIGNED, awaiting the engineering lead's ruling** on owner and
shape, with the candidate shapes recorded so the ruling has options. It sits directly in Story 7.4's
path, whose whole subject is typing and pasting multi-paragraph clause text onto that surface. Not
fixed here: this story's contract forbids both designer-surface work and new diagnostic code.

**Story 7.3 (FR47) depends on the `wrappedLine` break-kind field landed here.** D-R7.3's numeric
order satisfies that dependency only by accident, so a reorder must not silently break it.
