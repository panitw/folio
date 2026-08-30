---
title: 'Story 15.1: The golden report''s moved hash is explained'
type: 'bugfix'
created: '2026-08-30'
status: 'blocked'
baseline_revision: 'f91246319751dcaa38abdd2ccc0bad04552a4589'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-15-context.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
warnings: [multiple-goals, oversized]
deferred: ['DW-24 — style.valign renders but no fixture declares it, so textValignOffset has zero GOLDEN coverage; owner Story 7.1, trigger any story touching the vertical model or text_alignment.go']
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the intent contract below governs implementation.*

Folio's central promise is that the same template and data always produce the same PDF, byte for
byte. Four reference statements are committed to prove it. All four now render differently from the
copies on record — every page four bytes bigger — and nobody has written down why. Until someone
does, the reference suite has stopped being a test.

This story finds out what changed by opening both PDFs and comparing what the renderer actually
wrote, rather than by reading commit messages and guessing. It names the change in ordinary language
and writes that explanation where the next person will find it. It also corrects project records
that describe the problem wrongly, and actually runs the three larger statements the automated build
never runs, so their verification stops being an assumption.

The cause turns out to be small and the change turns out to be right. The footer of each statement
reads "Page 1 of 1" and the template has always asked for it to sit against the right-hand margin.
Until recently the renderer read that request, accepted it, and then drew the text at the left of its
box anyway, because nothing had ever been connected to act on it. That connection was made, the text
moved to where it was always meant to be, and the number describing its position grew four characters
longer. Nothing else about any of the documents changed.

The owner has ruled that this is wanted, so the reference copies are re-made to match. What this
story deliberately does not do is quietly refresh them: the explanation is written down first, and
the human confirmation that the reference documents still read correctly is asked for again rather
than carried over — the old confirmation was given against pages that placed the footer differently,
so it no longer describes what is on record.

Done looks like: a written explanation naming the cause, the three larger statements actually
rendered and measured rather than assumed, corrected project records, re-made references, and a fresh
human sign-off. Two things will look wrong along the way and are not. The sign-off check is expected
to be red until a person has genuinely re-read the pages. And one unrelated test about Thai name
coverage fails permanently by an earlier decision; fixing it is explicitly forbidden.

<intent-contract>

## Intent

**Problem:** All four golden statement fixtures render differently from their committed goldens —
exactly +4 bytes per page each (`statement-1` +4/1pg, `-5` +20/5pg, `-20` +80/20pg, `-50`
+200/50pg) — with no recorded explanation, so `TestStatementGoldenFixtures` is red and C6's
"regenerated rather than investigated" failure mode is one commit away.

**Approach:** Attribute the move to a named commit *and* a named emission behaviour by diffing the
produced page content streams against the recorded ones, record that attribution plus the heavy
5/20/50-page digests, correct the two project records that describe the move wrongly, and then HALT
for the owner's intended-or-not ruling before any golden byte is touched.

## Boundaries & Constraints

**Always:**
- Attribute by **content-stream diff of produced vs recorded PDF bytes**, never by commit-log
  inspection alone. The commit is already known to be `791ed00`; the deliverable is the *behaviour*.
- The named behaviour must explain **exactly 4 bytes, per page, uniformly across all four fixtures**.
  An attribution that cannot account for that arithmetic is not an attribution — keep diffing.
- Separate the three candidate classes explicitly before naming one: element-box paint (stroke/fill
  operators and edge operands), text ink (the colour bracket), and geometry (variable-width `Tm`
  coordinate operands that shift bytes without adding an operator).
- Record the attribution in the decision log **before** any golden file is touched (AC2 ordering).
- Reuse `splitPageContentStreams`; a second copy already exists, a third is a defect.
- Every measurement is stated as a number that was observed, with the command that produced it.
- Commit only on `main`.

**THE OWNER'S RULING — GIVEN 2026-08-30, recorded as D-R7.6. This contract now carries it.**

The move is **INTENDED**. `style.align` should render; the pre-`791ed00` output was the defect. Take
the **"Ruling = intended"** row of the matrix below. The `owner ruling required` halt is DISCHARGED —
do not halt on it, and do not re-open the question.

**The attribution is already measured — VERIFY it, do not search for it.** Both the engineering lead
and the orchestrator reproduced this independently before the ruling. AC1 still requires you to
produce the content-stream diff yourself, but it is a confirmation step with a known answer, not an
investigation. The answer:

- The mover is **geometry**, not the box paint and not the text ink: one variable-width `Tm` x-operand.
- `statement-1`'s content stream `/Length` moves 4548 -> 4552. Exactly one byte run differs in the
  whole file, the same ten glyphs (`Page 1 of 1`) at a different position:
  - recorded: `1 0 0 1 436 53.88 Tm`
  - HEAD:     `1 0 0 1 522.474 53.88 Tm`
  - `"436"` (3 chars) -> `"522.474"` (7 chars) = **+4 bytes**.
- The element is `pageFooter` `e4` (`x=400`, `width=123`, `style.align="right"`). The page footer
  repeats on every page, so the delta is +4 **per page** and scales exactly across 1/5/20/50.
- The cause is `folio-go/text_alignment.go`, CREATED in `791ed00` with no story file, no AC and no
  FR, wired into `render.go`'s `collectBandTextRuns`. Epic 9's element-box paint contributes nothing
  for a document declaring neither background nor border, and the `rectdoc.go` `appendEdge`
  prefix->postfix fix is length-neutral. **Do not look for the defect there.**
- The arithmetic that proves the new output is correct: element coordinates are BAND-RELATIVE, so
  `e4`'s box spans 436 -> 559 absolute (36pt margin + x). 559 is the content right edge. The old
  position 436 was the box's LEFT edge — alignment ignored. The new 522.474 leaves 36.526pt for ten
  characters at 7pt, i.e. 0.52em each: correct for a proportional face.
- **Population, which closes the red set:** exactly four templates in the corpus declare a non-`left`
  `style.align` — `fixtures/statement-{1,5,20,50}/input.folio`. No other fixture could have moved.
  **No fixture anywhere declares `valign`**, so that half of `791ed00` shipped with zero coverage.

**Block If:**
- **The human sign-off.** After re-recording the goldens, `expected.json`, the READMEs and the
  `byte_neutrality_test.go` literal, **HALT** with status `blocked` and blocking condition
  `human sign-off required`. Leave `fixtures/statement-signoff.json` UNTOUCHED and its test red.
  Render the four PDFs to `_bmad-output/implementation-artifacts/evidence/15-1/` so the owner can
  open them. **You may not write, edit or synthesise the `reader`, `date` or `examined` fields.**
  That file is a human attestation that a named person opened the pages and looked at them; an agent
  filling it in is a fabricated record, and D-4.7.1 invalidated the existing one in whole. The
  orchestrator obtains the real attestation and writes it.
- The diff cannot account for the 4-bytes-per-page arithmetic with any named behaviour — HALT with
  blocking condition `attribution not reached`.
- Any fixture's move is found to be non-uniform (not exactly 4 bytes × page count) — the premise
  changed; HALT with blocking condition `move is not uniform`.

**Never:**
- **Never re-record, regenerate, or digest-patch any golden to obtain green.** `statementDigestRemedy`
  forbids it in the test's own failure text and AD-22 makes a moved digest a versioned behaviour
  change. There is no regeneration flag in this repo and none is to be added.
- Never patch `fixtures/statement-signoff.json` at all — not its digests, and above all not its
  `examined` prose. The sign-off is invalidated **in whole across all four documents** (D-4.7.1) and
  is re-attested by a human, not edited by this story. Re-run the semantic acceptance step
  (`statement_semantics_test.go`) as part of the evidence the owner's pass rests on.
- Never touch `TestCorpusMeetsP6ExerciseFloors` / the P6g floor. It is a mandated permanent red
  (D-000.17, D-2.1.14, DW-11) and filling it is forbidden.
- Never repeat the two wrong records rather than correcting them: only `statement-1` is named as
  moved (all four moved), and Epics 9 and 10 are named as joint suspects (only Epic 9 is implicated).
- No feature work, no Epic 9 redesign, no CI restructuring (that is 15.2), no release (15.3).
- Never push. Never create a branch.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Diff a moved fixture | `fixtures/statement-1/expected.pdf` vs a live render of the same three inputs | Per-page stream pair whose byte delta is exactly 4, localized to one named emission site | If the delta is not 4/page, HALT `move is not uniform` |
| Candidate discrimination | The differing byte span | Classified as box-paint, text-ink, or geometry, with the operator or operand that changed | If it fits none, HALT `attribution not reached` |
| Heavy fixtures (AC5) | `statement-5`, `-20`, `-50` rendered in full | Each renders, page tree resolves, sha256 computed and recorded as measured evidence | A render failure or malformed page tree is a finding, not a skip |
| Ruling = intended | Owner ruling recorded in this contract | Four goldens + 4 `expected.json` + 4 READMEs + the `byte_neutrality_test.go` literal + signoff re-recorded in ONE commit; sign-off re-requested in whole | Matrix must be green on all four targets afterwards |
| Ruling = unintended | Owner ruling recorded in this contract | Epic 9's emission behaviour fixed; goldens untouched; a fixture ADDED that exercises the moved path | Goldens must return to their recorded digests unaided |

</intent-contract>

## Code Map

**The failing test and its standing prohibition**
- `folio-go/statement_golden_fixture_test.go:83` -- `TestStatementGoldenFixtures`; iterates the
  in-Go table, checks digest self-consistency (`:99`) then byte-compares render vs golden (`:103`).
- `folio-go/statement_golden_fixture_test.go:29` -- `statementDigestRemedy`, the constant that
  forbids updating a digest to go green. **Read it before touching anything.**
- `folio-go/byte_neutrality_test.go:603` -- `goldenDigestRemedy`, the repo-wide longer form.
- `folio-go/statement_fixture_test.go:348` -- `statementFixtures` table: slug/rows/pages for all four.

**The corpus and every place a digest lives**
- `fixtures/statement-{1,5,20,50}/` -- `expected.pdf`, `expected.json` (`{folioGoVersion,
  goToolchain,sha256}`), `input.folio`, `params.json`, `data.json`, `README.md` (digest at line 4).
- `folio-go/byte_neutrality_test.go:92` -- `goldenDigestRecord`; statement entries at `:231/:245/
  :264/:280`, each declaring FOUR sites (`expected.json`, `second-literal`, `readme`, `signoff`).
  `TestGoldenDigestAgreesAtEveryDeclaredSite:365` fails on any undeclared occurrence too.
- `fixtures/statement-signoff.json` -- one record, `digests` keyed by slug, all four; schema at
  `folio-go/statement_golden_fixture_test.go:163`; staleness is all-or-nothing (`:196`).
- `folio-go/statement_semantics_test.go:127` -- `renderStatement`, the live render seam; the file is
  the semantic acceptance step D-000.22 names.
- **There is NO regeneration path** — no flag, env var, `go:generate`, or Make target writes
  `fixtures/*/expected.pdf`. Recording is deliberate and manual, by design.

**Diff tooling that already exists — reuse, do not rewrite**
- `folio-go/multi_page_fixture_test.go:377` -- **`splitPageContentStreams(t, b) []string`**: resolves
  `/Kids`, follows each `/Contents`, returns per-page streams in page order. This is the tool.
- `folio-go/formatlocale_test.go:131` -- `flSplitPageContentStreams`, an existing duplicate. Do not
  add a third copy.
- `folio-go/shaped_fixture_test.go:42/:81/:627` -- `pdfObjects`, `streamBody`,
  `parseContentStreamRuns` (parses `BT…ET` into runs with `Tm` origins in millipoints).
- `folio-go/pdf_page_tree_check_test.go` -- `AssertPDFPageTreeResolves`, the structural oracle.
- **Content streams are UNCOMPRESSED** (`internal/pdf/textdoc.go:272-277`, raw `/Length`), enforced
  by lint rule `no-compressor-import` (`lint/internal/rules/nocompressor.go:19`). No decompressor
  is needed and none may be added.

**The three candidate emission classes (unranked — the diff decides)**
- *Box paint:* `folio-go/element_box.go:52` `collectElementBoxRects` → `table_render.go:526`
  `buildCellRectWithBackgroundField` → `internal/pdf/rectdoc.go:27` `appendRectContentStream` (fill
  `:43-55`, stroke `:57-77`). `appendEdge` at `rectdoc.go:97`: `791ed00` flipped its operands
  prefix→postfix, **+1 fixed byte per call, ≤4 calls per stroked rect**. Gate: `r.HasStroke`, set
  only when a `style.border` is declared — and the statement templates declare none, so the diff
  must confirm reachability rather than assume it.
- *Text ink:* `internal/pdf/textdoc.go:829-834`, the `q\n<r> <g> <b> rg\n` bracket; emits **zero
  bytes** when `HasColor` is false — the reason Epic 10 (`304442f`) moved nothing.
- *Geometry:* `folio-go/text_alignment.go:47/:65` `textAlignOffset`/`textValignOffset` →
  `render.go:874/:877` → `TextRun.X/Y` → the `Tm` operands at `textdoc.go:841-843`, spelled by
  `internal/pdf/numbers.go:39` `appendLength` (variable width: sign, integer part, ≤3 trimmed
  fractional digits). **A coordinate change adds bytes without adding an operator.** The statement
  footer element `e4` carries the corpus's only element-level `align:"right"` and is drawn on every
  page — a per-page repeated element, which is the shape the +4/page signal has.

**What `791ed00` changed on the emission path** (`git show --stat 791ed00`): `internal/pdf/
rectdoc.go` (+17/−5, the only `internal/pdf` change), new `element_box.go`, new `text_alignment.go`,
`render.go` (element align/valign offsets; element rects prepended to `tableRects`). No fixture was
re-recorded in that commit — hence render-vs-golden divergence rather than a recorded move.

**Records that are WRONG and must be corrected, not repeated**
- `_bmad-output/implementation-artifacts/sprint-status.yaml:349-353` (only `statement-1`; Epics 9
  and 10 jointly), `:286-288` (same joint framing), `:193-195/:200` (Epic 9's "no golden moved"
  claim, false as shipped).
- `_bmad-output/planning-artifacts/epics.md:3865-3871` (same two errors).
- `_bmad-output/implementation-artifacts/epic-15-context.md:80-82` (same joint framing).

**Governing decisions and their formats**
- `folio-mvp-decision-log.md` -- D-000.17 (`:5143`) and D-2.1.14 (`:5226`) mandate the P6g red;
  D-000.22 (`:5544`); D-000.26-refined (`:7094`); D-4.7.1 (`:13672`, sign-off invalidated in whole).
  Entry format: `### D-<id> — <verdict sentence>`, bold attribution line, then `**Verdict.**` /
  `**Situation.**` / `**Options considered.**` / `**Consequences.**` / `**How we'd know it was
  wrong.**`, separated by `---`. Append-only.
- `epic-7-8-decision-log.md` -- D-R7.0 (the measured bisect), D-R7.1 (`:73`, the every-story heavy
  cadence override this story runs under), D-R7.4 (the orchestrator renames this spec at the gate).
- `architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md:405` AD-21, `:417` AD-22.
- `prds/prd-folio-2026-08-22/prd.md:522` C6.
- `deferred-work.md:674` DW-11 (the P6g floor, deliberately visible); entry format at `:1-10`.

**Matrix**
- `folio-go/matrix_test.go:1668` `TestTargetRenderHash` (`//go:build matrix`); **a no-op that asserts
  nothing unless `FOLIO_MATRIX_TARGET` is set** — targets `matrix_test.go:69`. The all-four-in-one-
  process local gate is `TestCrossTargetByteIdentity` (~`:1495`).
- Heavy render seam: `folio-go/render_test.go:2180-2189`, env
  `FOLIO_SUBPROCESS_RENDER_STATEMENT{1,5,20,50}`.

## Tasks & Acceptance

**Execution — Phase A: evidence, attribution, records (runs now)**

- `folio-go/` (read-only investigation) -- render each of the four fixtures live from its committed
  `input.folio`/`data.json`/`params.json` and capture the bytes -- establishes the produced side of
  the diff without mutating anything.
- `folio-go/` (throwaway diff harness, `t.TempDir()` only) -- using `splitPageContentStreams`,
  diff produced vs recorded page streams for `statement-1` first, then confirm the same span on a
  multi-page fixture -- locates the differing bytes. **Reuse the existing splitter.**
- `folio-go/internal/pdf/rectdoc.go`, `folio-go/internal/pdf/textdoc.go`,
  `folio-go/internal/pdf/numbers.go` (read-only) -- classify the differing span against the three
  candidate emission classes by reading the exact emission site the diff points at, and show the
  arithmetic reaching exactly 4 bytes per page -- this is AC1's deliverable; a class name without
  the arithmetic does not close it.
- `_bmad-output/implementation-artifacts/evidence/` -- record the diff evidence: the page-stream
  excerpt before and after, the named behaviour, and the byte accounting -- so the attribution is
  reproducible by someone who was not here.
- `fixtures/statement-{5,20,50}` (AC5) -- actually run the three heavy fixtures, confirm each page
  tree resolves, and record each rendered sha256 and byte count beside its recorded digest -- CI
  never runs these; the four page counts must be verified rather than assumed from the one-page case.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` -- append a decision entry in the
  established format naming the attributed commit (`791ed00`) and behaviour, the four-fixture
  measurement, and that the intended-or-not call is **pending the owner** -- AC2 requires the log to
  carry the attribution before any golden is touched.
- `_bmad-output/implementation-artifacts/sprint-status.yaml` -- correct lines 349-353, 286-288 and
  193-195/200: all four fixtures moved, +4 bytes/page; Epic 9 alone is implicated; Epic 9's
  "no golden moved" note is false as shipped -- a wrong record is worse than no record.
- `_bmad-output/planning-artifacts/epics.md:3865-3871` -- correct the same two errors in the Epic 15
  prose -- prose only, outside any story's ACs.
- `_bmad-output/implementation-artifacts/epic-15-context.md:80-82` -- narrow the suspect framing to
  Epic 9 -- keeps the compiled context from re-seeding the wrong premise.
- **HALT** with blocking condition `owner ruling required: golden move intended or unintended`.

**Execution — Phase B: remediation (ONLY after this contract carries the ruling)**

- *If ruled INTENDED:* re-record in ONE commit -- 4 `expected.pdf`, 4 `expected.json`, 4 `README.md`
  digest lines, the `byte_neutrality_test.go` second literal, and `fixtures/statement-signoff.json`
  -- then re-run the semantic acceptance step and **re-request the human sign-off in whole across all
  four documents**; the commit message says what moved and why; matrix green on all four targets.
- *If ruled UNINTENDED:* fix the behaviour at whichever site the diff named --
  `folio-go/internal/pdf/rectdoc.go` (`appendEdge`), `folio-go/text_alignment.go` /
  `folio-go/render.go` (the alignment offsets), or `folio-go/element_box.go` -- leave every golden
  untouched, and **add a fixture under `fixtures/` that exercises the moved path**, registered in
  `folio-go/byte_neutrality_test.go`'s `goldenDigestRecord` -- no document in the corpus exercised
  it, which is why nothing caught it.

**Acceptance Criteria:**

- Given `statement-1`'s new digest, when it is investigated, then the change is attributed to a named
  commit AND a named emission behaviour, established by diffing produced against recorded PDF content
  streams, and the attribution accounts arithmetically for exactly 4 bytes per page.
- Given the attribution, when it is recorded, then the decision log carries it, and it names whether
  the change is intended — and no golden file has been touched before that naming.
- Given the three wrong project records, when this story closes, then they state that all four
  fixtures moved and that Epic 9 alone is implicated.
- Given the heavy 5-, 20- and 50-page fixtures, when this story closes, then they have actually been
  run and their digests recorded, so all four page counts are verified rather than assumed.
- Given an intended change, when the goldens are re-recorded, then every affected fixture moves in one
  commit, the sign-off is re-requested in whole rather than patched, and the matrix is green on all
  four targets.
- Given an unintended change, when it is found, then it is fixed rather than absorbed, the goldens are
  never re-recorded, and a fixture is added that would have caught it.
- Given the whole story, when the suite is run, then `TestCorpusMeetsP6ExerciseFloors` is still red
  and untouched.

## Spec Change Log

## Review Triage Log

## Design Notes

**Why this story halts instead of deciding.** The two remediations are opposite and both expensive to
undo: one rewrites four goldens and destroys a human sign-off that took a person reading four
rendered documents; the other reverts shipped behaviour. Nothing in the evidence distinguishes
"wanted" from "regression" — that is a product judgement about whether element boxes *should* have
changed the emission, and it belongs to the owner. AC2 encodes the ordering deliberately: name
intended-or-not first, touch bytes second.

**Why the arithmetic is the real test of the attribution.** +4 bytes on a 1-page document and +200 on
a 50-page document is the same finding: one per-page site gained 4 bytes. `appendEdge`'s
prefix→postfix flip is +1 byte per call and four calls per stroked rect, which multiplies to 4 —
seductively exact. But its gate is `HasStroke`, and the statement templates declare no border, so
the diff must prove the path is reached before that story is told. The competing shape is a
per-page-repeated element whose `Tm` coordinate operands changed width under the new alignment
offsets — the footer element is drawn on every page and carries the corpus's only element-level
`align`. Both produce +4/page. Only the bytes distinguish them.

**Uncompressed streams are a gift here.** Because a lint rule bans compressor imports, both PDFs can
be diffed as written, with no decode step and no new tooling.

## Verification

This story is byte-identity-shaped, so it carries the heavy tests despite the run's per-epic cadence
(orchestrator override, D-R7.1). Report measured counts, never "green".

**Commands:**
- `cd folio-go && go test ./...` -- expected: exactly ONE failing test,
  `TestCorpusMeetsP6ExerciseFloors` (P6g, got 7 need >=20 — the mandated permanent red under
  D-000.17/D-2.1.14). `TestStatementGoldenFixtures` MUST be green when this story is done.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean; proves the matrix-tagged code still
  compiles.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- expected: the cross-target
  hash matrix. **Note: this test asserts nothing unless `FOLIO_MATRIX_TARGET` is set** — record which
  target was measured, or record that it ran as a no-op. A no-op is not a pass.
- `cd lint && go test ./...` -- expected: all pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: all pass.

**Manual checks:**
- The four heavy digests recorded in Phase A are stated as observed sha256 values beside their
  recorded counterparts, with the command that produced each.
- The owner ruling is recorded in this contract (it is, above), and no `fixtures/` file is modified
  before the attribution is written to the decision log.
- `fixtures/statement-signoff.json` is byte-identical to its state at `09bb30e` when this dispatch
  halts. Its test is EXPECTED to be red at that halt; that red is the story working, not failing.
- A `valign` fixture is added, or its absence is deferred with a named owner and trigger — AC4's
  "a fixture is added that would have caught it" applies to the uncovered half of `791ed00`.

## Delivery Log

**Dispatch 2 (implementation), 2026-08-30, baseline `f912463` on `main`.** Phase A and Phase B both
ran; the ruling in the intent contract (D-R7.6) had already discharged the Phase A halt. Full
evidence, with the command that produced every number, is
`_bmad-output/implementation-artifacts/evidence/15-1/attribution.md`, and the four rendered PDFs sit
beside it for the owner's visual pass.

**AC1 — attribution by content-stream diff.** Rendered each fixture live through the same
`renderStatement` seam the golden test uses, resolved both the produced and the recorded PDF with the
EXISTING `splitPageContentStreams` (no third copy was written), and diffed the per-page streams. On
every page of all four fixtures the stream grew by exactly 4 bytes and **exactly one line differed**,
out of 230 lines on a `statement-1` page and 485–505 on the others:
`1 0 0 1 436 53.88 Tm` → `1 0 0 1 522.474 53.88 Tm` (`statement-1`, `-5`) /
`1 0 0 1 514.466 53.88 Tm` (`-20`, `-50`). The `Tj` glyph string on the next line is byte-identical.

The three candidate classes were separated **mechanically** before one was named: an operator census
(`re f S rg m l Tm Tj TJ q Q w`) over the joined page streams is IDENTICAL recorded vs produced for
all four fixtures. That eliminates box paint (would add `re`/`f`/`S`/`m`/`l`; `appendEdge`'s
`HasStroke` gate is never reached — confirmed, not assumed) and text ink (would add `rg`; zero bytes
when `HasColor` is false, which is why Epic 10 moved nothing). Geometry is the only class that adds
bytes without adding an operator, and it is what the diff shows.

**The arithmetic.** `appendLength` spells `436` in 3 characters and `522.474`/`514.466` in 7 — **+4
bytes, once per page**, on `pageFooter` element `e4`, the corpus's only element-level non-`left`
`style.align`, drawn on every page. 4 × {1,5,20,50} = {4,20,80,200}, matching the file deltas exactly,
with no residue. And the new position is provably the right one: `36 + 400 + 123 = 559` is the box's
absolute right edge, and `522.474 + 36.526 = 559.000`, `514.466 + 44.534 = 559.000`. The recorded
`436.000` was the box's LEFT edge on every page of every fixture.

**A causal proof beyond the diff, which is the strongest thing in this story.** `render.go`'s two
alignment offsets were temporarily multiplied by zero and the golden test re-run: **all four fixtures
went green byte for byte** against their pre-existing recorded goldens. Edit reverted; `git diff`
clean. Nothing else in the tree contributes a byte to this move.

**AC5 — the heavy fixtures were actually run, not assumed.** All four rendered in full, each page
tree resolved through `AssertPDFPageTreeResolves`, and each resolved page count compared against the
declared one: 1/1, 5/5, 20/20, 50/50. Measured bytes and digests: 76,744 `114df1d6…`; 127,363
`70dce051…`; 269,884 `56bfbbd9…`; 555,829 `5d090b0f…`. These are exactly the four counts D-R7.5
predicted, so its falsification condition did not fire.

**AC2 — ordering held.** `D-15.1.1` was appended to `folio-mvp-decision-log.md`, and the three wrong
records were corrected, BEFORE any file under `fixtures/` was touched.

**AC3 — the three wrong records corrected**, not repeated: `sprint-status.yaml` (Epic 9's false "no
golden moved" note, the joint Epic 9/10 suspect framing in the readiness gate, and the Epic 15 block
that named only `statement-1`), `epics.md`'s Epic 15 prose, and `epic-15-context.md`'s cross-story
dependency. All now say all four fixtures moved, +4 bytes per page, Epic 9 alone implicated.

**Phase B — re-recorded in this one commit:** 4 `expected.pdf`, 4 `expected.json`, 4 `README.md`
(digest line, byte count, and a new section recording what moved and why), and the four
`goldenDigestRecord` second literals in `byte_neutrality_test.go`. The independent-reader acceptance
each README claims was **re-run on the new bytes** — `qpdf 12.4.0 --check` / `--show-npages`, same
version, identical output, correct page count on all four.

**`fixtures/statement-signoff.json` was NOT touched** — verified `git diff 09bb30e` on that path is
empty. No `reader`, `date` or `examined` field was written, edited or synthesised.

**AC4 / the `valign` obligation — deferred with a named owner and trigger, on a measurement.** The
same experiment was run on the vertical half: multiplying `textValignOffset` alone by zero reddens
`TestAlignedTextElementsMoveInsideTheirDeclaredBox` and
`TestCanvasPaintMatchesTheShippingRunPathUnderAlignment` — the behaviour suite does cover it — and
**no golden anywhere in the repository**. Filed as **DW-24** in `deferred-work.md`: owner **Story
7.1**, trigger any story touching the vertical model or `text_alignment.go`. A fifth golden was not
added here because it would put an unsigned new artifact into a corpus whose sign-off is mid-
re-attestation and enlarge the diff the owner must read before signing.

**AC7 — `TestCorpusMeetsP6ExerciseFloors` is still red and untouched** (P6g, got 7 need >= 20).

### Measured verification (never "green")

| Command | Result |
|---|---|
| `cd folio-go && go test ./...` | **2 failing tests**, both expected: `TestGoldenDigestAgreesAtEveryDeclaredSite` (the untouched sign-off names the four old digests — the designed red) and `TestCorpusMeetsP6ExerciseFloors` (P6g, the mandated permanent red). `TestStatementGoldenFixtures` **PASS** on all four. |
| `cd folio-go && go vet -tags=matrix ./...` | clean, exit 0 |
| `cd folio-go && FOLIO_MATRIX_TARGET=<t> go test -tags=matrix -run TestTargetRenderHash -v .` | run for **all four targets** — `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` — each PASS, each producing the four re-recorded digests exactly. Not a no-op: the target was set every time. |
| `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity -v .` | PASS, 17.4s — all four targets agree on all documents |
| `cd folio-go && go test -tags=matrix ./...` | same two expected reds, plus `TestStatementSemanticSignOffIsRecorded` (the matrix-gated sign-off gate — the designed red), plus `TestShippedFacesReproduceFromUpstream`, which is **environmental and pre-existing**: `fontTools` is not importable by this machine's Python 3.12. Unrelated to this story; no font byte was touched. |
| `cd folio-go && go test -run 'TestStatement' -v .` | 19 tests PASS — the semantic acceptance step (`statement_semantics_test.go`, D-000.22) re-run against the re-recorded bytes |
| `cd lint && go test ./...` | all 4 packages PASS |
| `cd folio-designer && npm run typecheck && npm run lint && npm test` | typecheck clean; lint clean (4 pre-existing `only-export-components` warnings); **30 files / 213 tests PASS** |
| `qpdf --check` + `--show-npages` on all four re-recorded goldens | no syntax or stream encoding errors; 1 / 5 / 20 / 50 pages |
| `gofmt -l folio-go lint` | one hit, `lint/internal/rules/licencegraph_test.go` — pre-existing and already tracked as **DW-23**; untouched by this story |

## Auto Run Result

Status: blocked
Blocking condition: `human sign-off required`

**This is the designed halt, not a defect.** Every executable obligation of this story is discharged.
What remains cannot be done by an agent: a named person must open the four re-recorded PDFs in
`_bmad-output/implementation-artifacts/evidence/15-1/` — in a viewer, reading the drawn pages, never
through text extraction (D-4.7.7) — and then write `fixtures/statement-signoff.json` with their own
`reader`, `date` and `examined`, plus the four new digests:

```
statement-1   114df1d6508981d4eb162c585ff6f01eedf2a75393a5a2a9b649809e8ac968db
statement-5   70dce051495cf68daa71fe8185aa2467acfd82d10fb195439a4d71bcf41944d0
statement-20  56bfbbd9a7d20a2a9404fc931dfbe70da9d25979eec17cc8027c0f1063f84b9e
statement-50  5d090b0f01ddb5072636caded9feec2cad24cb16297a1afbba301b2a4802f171
```

The sign-off is invalidated **in whole across all four documents** (D-4.7.1), so it is re-attested,
never patched. Until it is, `TestGoldenDigestAgreesAtEveryDeclaredSite` and the matrix-gated
`TestStatementSemanticSignOffIsRecorded` stay red, **and that red is the story working**.

**What the reader should expect to see, so a genuine difference stands out:** the ONLY visible change
from the previously-signed pages is the footer's `Page N of M` moving about 86pt to the right, to sit
flush against the right margin. Anything else different falsifies the attribution and the re-record
must be undone (D-15.1.1's own falsification condition).

Dispatch: classic intent (no `stories.yaml` in this project). Epic context compiled to
`_bmad-output/implementation-artifacts/epic-15-context.md`.

Warnings: `multiple-goals` (investigation/attribution, the AC5 heavy-fixture verification, and the
ruling-gated remediation are independently shippable — carried forward per step-01 item 4, not
split), `oversized` (the spec exceeds the 1600-token guide; the Code Map is deliberately dense
because it is the implementer's diff map and the alternative is re-deriving it at dispatch time).
