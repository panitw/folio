---
title: 'A stacked Thai mark reaches the page'
type: 'bugfix'
created: '2026-08-31'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized']
deferred:
  - summary: "The new fixture's Thai mark placement needs a human sign-off record; no agent may write one."
    evidence: "Its subject is mark placement produced by a text-rise operator — a property nothing in the corpus has ever judged. fixtures/shaped-text/thai-signoff.json (D-2.3.5) attests GSUB lowered-form placement at zero offset, a different mechanism. Only the owner can commission a reader."
    location: 'fixtures/thai-stacked-marks/'
    severity: 'high'
  - summary: "shaping_expectations_test.go:33-40 still carries the falsified 2.3 unreachability claim (4th surviving instance)."
    evidence: "It states YOffset is 'a FORWARD GUARD WITH NO AVAILABLE RED-PROOF ... Do not manufacture a red-proof for it.' D-8.0.1 corrected the same claim in textdoc.go and textdoc_test.go but did not name this file. Adding a ทั้ง row to the frozen expectation table would red-prove it, but that table is bound to fixtures/shaped-text/harfbuzz-oracle.json and re-recording it is out of this story's scope."
    location: 'folio-go/shaping_expectations_test.go:33'
    severity: 'medium'
---

<intent-contract>

## Intent

**Problem:** `internal/pdf`'s `appendShapedRun` refuses any glyph carrying a non-zero `YOffset`, so ordinary Thai whose marks the shaper places by GPOS y-displacement (`ทั้งสิ้น`, `ครั้ง`, `ทั้งนี้`, `ตั้งแต่`) produces a hard `Render` error and zero bytes on the shipped Noto Sans Thai — the owner's real contract clause among them (DW-28, HIGH).

**Approach:** Express the offset with PDF's text-rise operator `Ts`, which is inside AD-6's pinned profile (settled, D-8.0.4), instead of refusing it. The `Ts` path is entered only when `YOffset != 0`, so every document in the existing corpus emits byte-identically.

## Boundaries & Constraints

**Always:**
- The `Ts` path is entered **only** when a glyph's `YOffset != 0`. A run whose every glyph carries zero offset emits the exact bytes it emits today (`<hex> Tj` or `[...] TJ`), with no `Ts` operator and no `q`/`Q` bracket added. This is the whole of the story's byte-identity risk and must be asserted, not assumed.
- All **21** `goldenDigestRecord` digests stay unmoved, on all four AD-21 targets.
- The rise reaches the output through `numbers.go`'s emitters (AD-3). CIDs keep their separate route — a big-endian hex pair inside a string literal, per D-1.1.b — and **must not be "unified"** with it.
- No `float64` anywhere (AD-23). The rise is `geom.Length` millipoints derived by `geom.ScaleRound`, whose round-half-to-even is what makes all four targets agree.
- Text rise is a persistent text-state parameter that survives `ET`. Every run therefore begins and ends at rise 0.

- **THE TRIGGER IS A NON-ZERO `YOffset`, NOT MARK STACKING.** The epic's original *"two stacked
  marks"* wording was measured over-broad at this gate and has been corrected in `epics.md`: `ที่`,
  `ป้ำ` and `ปั` each stack two marks and render **exit 0 today**, because Noto Sans Thai resolves
  the `ี`+tone case by a GSUB lowered-form substitution at zero offset and only the `ั`+tone case by
  a GPOS y-displacement. `ที่` already appears in `fixtures/shaped-text`, all four `statement-*`
  fixtures and `justified-thai`, so a two-marks predicate builds the wrong test and then reads the
  shipped goldens as contradicting this story.
- **"A message pinned as it is today" means PINNED BY A TEST, not byte-identical prose** — settled at
  this gate rather than left to the diff. The reason clause *"which a TJ array cannot express"*
  becomes **false** for the narrowed case: a TJ array is no longer why the glyph is refused, a rise
  that rounded to zero is. **Rewrite the reason clause to describe the narrowed condition.** Keep the
  second half — *"would place it wrongly with no observable difference in the output bytes, so this
  fails rather than degrades"* — **verbatim**, because it stays true and it is the sentence that
  explains why refusing beats degrading. Shipping a canonical statement that misdescribes its own
  condition is precisely the failure D-8.0.1 exists to stop.

**Block If:**
- Any of the 21 existing digests moves. That means the zero-offset path changed, which is a defect — HALT, do not re-record.
- The work would move an existing golden's digest for any reason. HALT and say so; `fixtures/statement-signoff.json` is a human attestation and a moved digest invalidates it in whole (D-4.7.1).
- A human sign-off record is needed. State it and stop — **no agent writes `reader`, `date` or `examined`.**

**Never:**
- Never touch `folio-go/internal/layout/` — zero paths. `paginate.go`'s prohibition on a closed-form page count stands.
- Never touch `folio-designer/` — zero paths. AD-5 makes the page model blind to the emission stage; the canvas is blind by design, not defective (D-8.0.2).
- Never widen into Epic 8's font-embedding stories (8.1–8.6). Nothing is embedded here.
- Never change the shipped faces or `TestShippedFacesReproduceFromUpstream`.
- Never delete an arm of `thai_mark_stacking_test.go` — re-point it. A test deleted rather than re-pointed is a blocking defect (D-7.8.7).
- Never add a second exported scaling function to `internal/geom` (`scale_surface_test.go:26` pins the set to exactly `{ScaleRound}`).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Stacked mark renders | `ทั้งสิ้น` at `fontSize: 12`, shipped Noto Sans Thai (`YOffset = -57` on CID 3) | Renders; the run splits and emits `-0.684 Ts` before the offset glyph, then `0 Ts` before `ET` | No error expected |
| Zero-offset corpus unchanged | Any of the 21 goldens | Byte-identical to the committed `expected.pdf`; no ` Ts` substring anywhere in the file | No error expected |
| Same-script control | `สัญญา` (all glyphs zero offset) | Renders exactly as today, no `Ts` | No error expected |
| GSUB-resolved stack still renders | `ที่`, `ป้ำ`, `ปั` — two marks over one base, resolved by a lowered-form substitution at zero offset | Renders as today, no `Ts` — these never reached the refusal | No error expected |
| Mixed offsets in one run | Glyph sequence with offsets `0, -57, -57, 0` | Three segments in index order; `-0.684 Ts` before segment 2, `0 Ts` before segment 3 | No error expected |
| Rise rounds away | `ทั้งสิ้น` at `fontSize: 0.008` — `ScaleRound(8, -57, 1000) = 0` | **Still refuses**, zero bytes | `verticalOffsetError`, message describing the rounded-away rise, pinned verbatim by a test |
| Missing width | CID with no entry in `WidthForGlyph` | Refuses | `missingGlyphError`, unchanged |

</intent-contract>

## Code Map

- `folio-go/internal/pdf/textdoc.go:892-984` -- `appendShapedRun`, the whole change. Refusal predicate at `:899`, `verticalOffsetError` raised `:933`, the `Tj` fast path `:950-957`, the `TJ` array builder `:959-983`. Its doc comment `:858-891` states the TJ adjustment arithmetic and D-1.1.b's CID carve-out — update it, keep the carve-out.
- `folio-go/internal/pdf/textdoc.go:783-856` -- `buildTextContentStream`, the caller. Emits `BT / Tf / Tm`, then the body, then `ET`; `q`/`Q` brackets exist **only** for clipped or coloured runs (`:812-814, 833-835, 848-853`), so an unclipped uncoloured run has no bracket to restore state — the rise must be restored explicitly.
- `folio-go/internal/pdf/textdoc.go:1026-1039` -- `verticalOffsetError` (fields `face`, `cid`, `offset`; `Error()` built by hand-rolled `itoa` at `:1045`, never `fmt`).
- `folio-go/internal/pdf/numbers.go:39` -- `appendLength` (geom.Length millipoints → points, ÷1000 done by the spelling). `:90` `appendInt` (raw integers; TJ adjustments use this). The rise is a `geom.Length`, so it takes `appendLength`.
- `folio-go/internal/geom/scale.go:58` -- `ScaleRound(v Length, num, den int64) Length`, round-half-to-even, no float intermediate. Panics on overflow.
- `folio-go/internal/pagemodel/pagemodel.go:221-226` -- `ShapedGlyph{CID uint16; XAdvance, XOffset, YOffset int64}`. Units are **thousandths of an em** (`:200-201`), not millipoints. `TextRun.FontSize geom.Length` at `:69`.
- `folio-go/render.go:2638` -- the only producer of `YOffset`: `int64(geom.ScaleRound(geom.Length(int64(g.YOffset)), 1000, upem))`. No sign flip anywhere; the shaper's y-up convention survives intact.
- `folio-go/thai_mark_stacking_test.go` -- the characterization (landed `c3df718`). Arm A `:105`, control arm B `:127`, owner's clause arm C `:151`; pinned message const `:97-100`.
- `folio-go/internal/pdf/textdoc_test.go:140` -- `TestShapedRunFailsClosedOnYOffset` + its non-vacuity leg `:149-154`. `:162` `TestZeroAdjustmentRunEmitsTj` pins `"<00010002> Tj\n"` exactly. `:236` `TestColouredRunBracketsItsInk` — **the pattern to copy** (see Design Notes).
- `folio-go/byte_neutrality_test.go:92` -- `goldenDigestRecord`, 21 entries, hashes `fixtures/<dir>/expected.pdf` (`:562-574`). `:814` `declaredEpic2GateObligations`. `:527` `TestGoldenDigestAgreesAtEveryDeclaredSite`.
- `folio-go/shaped_fixture_test.go:627` -- `parseContentStreamRuns`; its `scanAdjustments` (`:658-670`) `ParseInt`s every field between hex strings and does **not** know `Ts`. ~18 call sites. Read-only hazard, see Design Notes.
- `folio-go/matrix_test.go:1477` `matrixDocuments`; `:69-74` `matrixTargets`; `:2064` `TestTargetRenderHash`; `:1887` `TestCrossTargetByteIdentity`. All `//go:build matrix`.
- Read-only evidence: no operator allow-list test exists anywhere (`Ts` needs no profile registration); `emit_source_test.go` is a negative `strconv`/`fmt` rule with no site registry; content streams are **uncompressed** (`document.go:3`), so a ` Ts` byte scan over a golden is valid.

## Tasks & Acceptance

**Execution:**

1. `folio-go/internal/pdf/textdoc.go` -- Replace the `YOffset != 0` refusal in `appendShapedRun` with segment-wise `Ts` emission per the run-splitting rule in Design Notes. Compute `rise := geom.ScaleRound(run.FontSize, g.YOffset, 1000)`; emit via `appendLength` + `" Ts\n"`. -- This is the fix.
2. `folio-go/internal/pdf/textdoc.go` -- Narrow, do not delete, the fail-closed branch: refuse when `g.YOffset != 0 && rise == 0`, because emitting `0 Ts` would place the glyph wrongly with no observable difference in the output bytes — the exact hazard the existing message describes. Update `verticalOffsetError`'s message to state the rounded-away rise rather than TJ's inexpressiveness. -- AC3 of the epic; keeps a fail-closed posture that is red-provable through a real document.
3. `folio-go/internal/pdf/textdoc.go` -- Rewrite `appendShapedRun`'s doc comment (`:858-891`) and the FAIL CLOSED comment (`:904-932`) to describe what the code now does. **Keep D-1.1.b's CID carve-out verbatim.** State the population any negative claim was measured over (D-8.0.1). -- A comment asserting a negative carries a test's evidentiary burden.
4. `folio-go/internal/pdf/textdoc_test.go` -- Re-point `TestShapedRunFailsClosedOnYOffset`: the `YOffset: 37` leg must now assert a successful emit carrying the rise; keep the zeroed-YOffset non-vacuity leg. Add the byte-identity guard arm modelled on `TestColouredRunBracketsItsInk` (Design Notes). Correct the file's `:122-139` comment. -- The local expression of the byte-identity guardrail.
5. `folio-go/thai_mark_stacking_test.go` -- Re-point arm A and arm C from refusal to successful render (non-zero bytes, content stream carries the rise); leave control arm B **unchanged and green**. Move arm A's "a refused render emits zero bytes" property and the verbatim-message pin onto the narrowed refusal (a stacked-mark document at `fontSize: 0.008`), so `thaiStackedMarksMessage`'s discipline survives. Rename the two arms to match what they now assert. -- D-7.8.7: re-point, never delete.
6. `fixtures/thai-stacked-marks/` -- New fixture from the owner's contractor-liability clause (`deferred-work.md:2353`): `input.folio`, `expected.json` (`goToolchain` must be `go1.26.0`), `expected.pdf`, `README.md`. -- The epic requires the clause become a fixture.
7. `folio-go/thai_stacked_marks_template.go` -- New: template const kept byte-identical to `input.folio`, plus a data const. Follow `folio-go/justified_thai_template.go:65,92`. -- Repo convention.
8. `folio-go/thai_stacked_marks_fixture_test.go` -- New: render-vs-digest test plus the hand-sync drift check (`justified_thai_fixture_test.go:786` is the model), and an assertion that the emitted stream carries the rise operator. -- Without it the fixture is never re-rendered.
9. `folio-go/byte_neutrality_test.go` -- Add the fixture's entry to `goldenDigestRecord` (`:92`) and `"matrix-document: thai-stacked-marks"` to `declaredEpic2GateObligations` (`:814`). -- Both fail loudly otherwise.
10. `folio-go/matrix_test.go` -- Add the `matrixDocuments` entry (`:1477`) plus its `capture…Render` func. -- Four-target registration.
11. `folio-go/render_test.go` -- Add the subprocess selector const and its `TestMain` arm. -- Every matrix document renders in a fresh process.
12. `folio-go/missing_glyph_corpus_test.go` -- Add the fixture to the corpus table (`:41`) **and** to `beyondBaselineAcceptance` (`:210`) with a non-empty story-naming reason. Never add to `baselineAcceptanceFixtures`. -- The accounting identity at `:255` fatals if only one is updated.
13. `.github/workflows/matrix.yml` -- **Five** edits: the `docs=` list (`:259`) and one `hash.<target>.thai-stacked-marks.txt` line in each of the four upload blocks (`:53-74`, `:95-116`, `:137-158`, `:179-200`). -- `TestMatrixDocumentSlugsAreRegisteredInCI` is untagged and fails the ordinary suite if these disagree.
14. `folio-go/shaped_fixture_test.go` -- Teach `scanAdjustments` (`:658-670`) the `Ts` operator so a rise operand is never absorbed into `Adjustments`. -- See Design Notes; the helper is wrong in both directions today.
15. `folio-go/byte_neutrality_test.go` -- Add a test asserting no committed golden contains the substring ` Ts`. -- Cheap, non-vacuous (`Ts` appears nowhere today), and streams are uncompressed.

**Acceptance Criteria:**
- Given a text element whose value contains a Thai base carrying a GPOS-displaced mark, when the document renders, then it produces a PDF with each mark at the vertical offset the shaper computed.
- Given the 21 shipped goldens, when they render, then all 21 digests are byte-identical on all four targets and no output file contains ` Ts`.
- Given a run whose glyphs all carry zero offset, when it is emitted, then the bytes are exactly those emitted before this story — proved by deleting the rise operators from a positive-case stream and asserting byte-equality with the negative case.
- Given the owner's clause, when it renders, then a fixture built from it joins the corpus with a **new** digest — new, not moved; no existing digest changes.
- Given a glyph whose non-zero offset scales to a rise of zero, when it is emitted, then it still refuses with a pinned message and zero bytes.

## Spec Change Log

## Review Triage Log

## Design Notes

**The run-splitting rule — deterministic, read off the glyph sequence.**
1. Partition `run.Glyphs` into maximal contiguous segments of equal `YOffset`, walking in index order. A boundary falls exactly where `Glyphs[i].YOffset != Glyphs[i-1].YOffset`.
2. Emit segments in index order. Track the currently-emitted rise, which is `0` on entry to every run.
3. Before a segment whose rise differs from the current rise, emit `<rise> Ts\n`; after the last segment, if the current rise is non-zero, emit `0 Ts\n`.
4. The adjustment term that sits *before* glyph `i` leads the segment that glyph `i` opens. The trailing term belongs to the final segment.
5. Each segment chooses `<hex> Tj` vs `[...] TJ` by the existing rule — `Tj` when all of that segment's terms are zero.

Every step is a pure function of the glyph sequence and of `ScaleRound`'s integer arithmetic, so two targets cannot split differently. A run with one distinct offset value (0) is one segment, zero `Ts` operators, today's bytes.

**Sign.** `YOffset` is the shaper's y-up delta in thousandths of an em and is never negated between `render.go:2638` and here; `Ts` is y-up in text space. So the rise passes through with no sign change. The shipped values are negative: `ScaleRound(12000, -57, 1000) = -684` → `-0.684 Ts`.

**"Two stacked marks" is the WRONG predicate — measured, and the epic's own prose is over-broad here.** `ที่`, `ป้ำ` and `ปั` each stack two marks over one base and render today at exit 0, because Noto Sans Thai resolves them with a GSUB lowered-form substitution at zero offset; `ทั้` refuses because that pair is resolved by a GPOS y-displacement of `-57`. `ที่` appears in `fixtures/shaped-text`, all four `statement-*` and `justified-thai` and renders fine. **The trigger is "a glyph the shaper gives a non-zero `YOffset`", full stop.** An implementer who writes a two-stacked-marks predicate will build the wrong test and think the existing goldens contradict the story.

**Asserting "only when `YOffset != 0`" — copy `TestColouredRunBracketsItsInk` (`textdoc_test.go:236`).** It is the same claim for the colour operator, in the same file, and its fourth move is the one that matters: build the stream with the condition absent, flip one field on the *same* run, then delete the new bytes from the positive output and assert byte-equality with the negative output. Transposed: `bytes.ReplaceAll(withRise, []byte("-0.684 Ts\n"), nil)` with the restoring `0 Ts\n` also removed must equal the zero-offset stream exactly.

**The test-helper hazard (task 14).** `scanAdjustments` replaces `[`, `]`, `TJ`, `Tj` with spaces and `ParseInt`s every remaining field. A rise that is a whole number of points (`-12`) parses as an integer and is silently recorded as a TJ adjustment; a fractional one (`-0.684`) fails `ParseInt` and is silently dropped. Wrong in both directions, and ~18 call sites depend on it. Teach it `Ts` explicitly rather than relying on the operand happening to be fractional.

**Why the branch narrows rather than vanishes.** With `Ts` in place, all four `ShapedGlyph` fields are expressible, so the refusal has nothing left to catch *except* the rounding boundary. That boundary is real and reachable: `fontSize` has no positivity floor at parse (`parse.go:428` → `decodePoints`), and a stacked-mark document at `fontSize: 0.008` loads and reaches the emitter today (measured through the shipped CLI). There, `ScaleRound(8, -57, 1000) = 0`, so emitting the rise would drop the offset silently — "the healthy output and the broken output would be the same bytes", which is exactly the posture the existing comment defends. Ordinary sizes are unaffected: at any `fontSize >= 1pt`, `|rise| >= |YOffset|` millipoints, non-zero.

## Verification

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: exactly ONE failure, `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (got 7, need >=20) — the mandated permanent red. Never touch it or its drift twin `TestCorpusP6StatsMatchDeclaredBaseline`. 4 skips are normal (`TestBrowserAuthoredRoundTripWitness`, `TestXrefEntriesRejectsMalformedSubprocess`, and the toolchain/GOOS-conditional ones).
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `gofmt -l folio-go` (from repo root) -- expected: no output. (`lint/…/licencegraph_test.go` has a known separate finding, DW-23.)
- `shasum -a 256 fixtures/*/expected.pdf` -- expected: exactly 21 files, every digest identical to the pre-change value; plus one new line for `fixtures/thai-stacked-marks/expected.pdf`. **Any moved digest is a HALT, not a re-record.**
- `cd folio-go && FOLIO_MATRIX_TARGET=darwin/arm64 go test -count=1 -tags=matrix -run TestTargetRenderHash .` -- and once each for `linux/amd64`, `linux/arm64`, `js/wasm`. Expected: pass with per-document hashes written. **Also run it once with `FOLIO_MATRIX_TARGET` unset** and confirm it logs "asserts NOTHING and is a deliberate no-op" — that control proves the four legs were not no-ops.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: all four targets agree with each other and with the goldens.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` -- expected: unchanged (284 tests / 34 files at this baseline; 4 pre-existing `only-export-components` lint warnings). This story touches zero designer paths, so any movement here is a signal, not a pass.
- `cd folio-go && go run ./cmd/folio render -o /tmp/clause.pdf ../fixtures/thai-stacked-marks/input.folio` -- expected: exit 0 and a non-empty PDF. Today the same command fails at CID 27, offset −2, and writes no file. (Flags must precede the positional argument.)

**Manual checks:**
- Open `fixtures/thai-stacked-marks/expected.pdf` and confirm the content stream carries `Ts` operators and that each is restored to `0 Ts` before its `ET`.
- `TestShippedFacesReproduceFromUpstream` fails without fontTools installed — environmental, not a regression. Say so explicitly rather than reporting it green.
- Note the coverage gap: `page-count-20` has no ordinary-suite render-vs-golden test; its render is only re-derived under `-tags=matrix`. The matrix legs are what cover it.

## Auto Run Result

Status: ready-for-dev
Blocking condition: none
Dispatch: plan-only ("Halt after planning."). No implementation code was written and no commit was created.
Baseline: aad3f72 on main.
