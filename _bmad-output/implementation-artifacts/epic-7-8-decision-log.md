# Decision log — Epic 7 & 8 delivery run (with Story 15.1 as story zero)

**Run started:** 2026-08-30
**Target:** Story 15.1, then Epic 7 (7.1–7.7), then Epic 8 (8.1–8.6) — 14 stories.
**Baseline at run start:** `09bb30e` on `main`.
**Orchestrator:** run-dev-cycle. **Per-story loop:** `bmad-build-auto` via `bmad-story-builder`.

This log records decisions that outlive the story that raised them. Per-story narrative lives in
each spec's `## Delivery Log`; findings consciously not fixed live in `deferred-work.md`. Project
decisions predating this run are in `folio-mvp-decision-log.md`, which this log does not duplicate.

---

## Lead Grounding

Filed by the orchestrator 2026-08-30 from the engineering lead's one-time grounding report, at
baseline `09bb30e`. The lead grounded from `ARCHITECTURE-SPINE.md` (AD-1…AD-26 in full), `epics.md`
(inventory, coverage map, Epics 7 and 8 in full, 9/10/11/15 read), both SPECs and their five
companions, this log, targeted greps of the 14k-line MVP decision log, and `deferred-work.md`.

**Correction to the run brief, recorded first because it changes Story 15.1.** There are no separate
ADR files. `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md`
is the only architecture document, and the `AD-N` invariants inside it *are* the ADRs. Cite `AD-N`
plus a SPEC clause or story AC — never an ADR path.

### The spine, as it constrains this run

- **Byte identity is the product, not a quality bar.** AD-2 (one fixed-point unit — `geom.Length`,
  int64 millipoints), AD-3 (numbers reach the PDF through one file in two representations), AD-23
  (no `float64` under `internal/`, geometry or data), AD-1 (determinism is a *directory* boundary),
  AD-21 (four-target matrix; a hash change is a defect until proven intended), AD-22 (any change to
  layout, subsetting, emission, the locale table or the toolchain is breaking, and ships as such).
  Counter-metric C6 — "golden hashes regenerated rather than investigated" — is what Story 15.1 exists
  to prevent.
- **One authority per fact,** recurring at every layer: one number→text emitter (AD-3), one
  coordinate-flip function (AD-24), one canonical serializer (AD-9), no TypeScript model of a
  document (AD-15), the browser never measures text (AD-17).
- **Command / projection** (AD-15, AD-16): the UI paints from an immutable snapshot and sends every
  committed mutation as one opaque command. This binds 7.4, 7.5, 7.6, 8.1, 8.2 and 8.6.
- **The window model** (D-2.6.1): pages are sliding windows onto one unbounded content column, which
  is never mutated. Epic 7's own header makes this an *input* to the epic, not a target.
- **The overflow scope line:** content exceeding its declared *box* is FR44/Story 2.8 — clip, diagnose,
  still return bytes. An item taller than the *window* is a located template error. Route every new
  overflow question through that distinction first.
- **AD-14's diagnostic contract:** one closed, additive SCREAMING_SNAKE registry; clipped content and
  over-tall rows are Warnings returned *alongside* PDF bytes, never fatal.
- **AD-9's P1** (`Parse(Serialize(d)) == d`) is forced, not chosen — the serializer preserves what it
  does not understand, orphaned assets included. This matters more in Epic 8 than it ever has.

### Risks the lead flagged before any story is specced

**Epic 7.** 7.1 and 7.2 are load-bearing: both change the vertical model everything downstream
measures against. 7.2's trap is in its own AC — `lineSpacing` scales `Advance` and *nothing else*;
`FirstBaseline` is unchanged, which is D-2.5a/DW-15's deliberate two-model split. 7.3 is where a spec
is most likely to drift: slack must be distributed in integer millipoints by one stated ordered rule
(a per-gap float divide is a second rounding site, banned by AD-2/AD-3), and a Thai line with no
interior break opportunities (AD-25) must fall back to its natural start edge — an underflow case the
AC does not cover. 7.3 also *depends* on 7.1: a line ended by a mandatory break must be
distinguishable at pack time. 7.5 is a designer-side command-validation change and must not touch
`internal/layout/paginate.go`. 7.7 extends `ItemGroup` past table rows for the first time, which is
exactly the co-extensiveness D-4.6.2 left a tripwire for.

**Epic 8.** 8.3 is load-bearing in a way the others are not — it changes the *format*, the one public
contract between two users of equal standing. Three findings, all pulled forward:

1. **The `.folio` version bump is a genuine owner fork.** D-1.4.12 (extending a closed set is MAJOR)
   points at MAJOR; but a higher MAJOR is a load error, and Story 15.3 is about to cut
   `folio-go/v0.1.0`, so a 2.0 document would be unreadable by the first tagged release forever.
   Materially cheaper to settle before the tag than after. Raised at 8.3's plan gate.
2. **8.3's "closed set of media types" contradicts D-1.8.1 (amended)**, which ruled `mediaType` must
   *not* be a closed set and an unrecognised one is never a load error — and whose own note predicted
   this recurrence "later for font formats". The lead will rule rather than escalate: unrecognised
   media type is preserved at load and errors at *render*; a *recognised* type whose bytes do not
   decode stays a load error, as does a chain entry naming an absent asset key.
3. **8.6's "unreferenced font assets are dropped on save" contradicts a forced invariant.**
   `folio-go/internal/template/serialize.go:420-427` preserves orphans unconditionally because P1
   forces it — its comment states there is "no policy latitude to drop one" and that collecting
   orphans is a designer *command*, never a serializer side effect. D-5.13.3's image-asset precedent
   already gives the correct shape. `spec-fonts/font-catalogue.md` needs amending in the same story.

**A data-loss trap for 8.1 and 8.6.** `assetKeyReferenced` in `folio-go/component_commands.go` matches
image elements only; a font asset is referenced from the `fonts` chain map, not from any element, so
that function returns false for *every* font asset. Any orphan check that reuses it deletes live
faces, with no compile error to announce it.

**Bold/italic is not unruled — this closes the run brief's open flag.** `epics.md` Epic 11 (FR57) owns
it with Stories 11.1–11.3. SPEC-fonts' open-question list predates Epic 11 and is stale, not open. No
Epic 8 story depends on the answer; the only residue is 8.5's catalogue shipping Regular-only, which
Epic 11 later widens. A weighted catalogue face proposed in 8.5 is a collision with Epic 11.

**Also noted:** 8.4 lands on DW-14 (unbounded `beginbfchar`) and DW-13 (uncompressed `FontFile2` — an
Epic 8 document pays for whole faces *and* an embedded subset). DW-21 gates the heavy statement legs
behind `FOLIO_HEAVY=1`; unset, they `--- SKIP` rather than silently pass, so Story 15.1 must set it or
its last AC passes on three suites that never ran.

---

## Standing decisions (settled at setup, 2026-08-30)

### D-R7.0 — Story 15.1 runs first, as story zero of this run
**Owner decision** (asked at setup; the owner chose "explain the move first" over two proceed-anyway
options).

**Verdict.** The run does not start at Story 7.1. It starts at Story 15.1 — explain the moved golden
hash — and Epic 7 does not begin until the golden corpus is green or its move has been deliberately
re-recorded with fresh human sign-off.

**Situation.** The four golden statement fixtures (`statement-1/5/20/50`) are the project's primary
acceptance evidence: each is a `.folio` template whose rendered PDF bytes are frozen and committed,
so any later change that moves those bytes is caught the moment it lands. At run start all four were
RED. Verified by the orchestrator before the run began, not taken from the sprint note:

| Commit | Golden corpus |
|---|---|
| `be40f24` (before Epic 9) | green |
| `791ed00` (Epic 9 — "a component's box prints") | **RED — all four, +4 bytes per page** |
| `304442f` (Epic 10 — text colour) | RED, byte counts unchanged from `791ed00` |

So **Epic 9 moved the corpus and Epic 10 did not.** This corrects two records: `sprint-status.yaml`
named only `statement-1` (all four moved) and named Epics 9 and 10 jointly as suspects (only 9 is
implicated). Epic 9's own sprint-status note claims "a document declaring no element background or
border must hash identically and the golden corpus stays a witness" — that claim is false as shipped.

The separate red, `TestCorpusMeetsP6ExerciseFloors` (P6g 7 < 20), is **not** part of this: it is the
permanent, mandated red D-000.17 / D-2.1.14 require to stay unmet. It is expected and stays red.

**In simple terms.** The goldens are a tripwire: a set of documents whose output bytes were checked
by a human once and then frozen, so the build can shout if anything ever changes them. Somebody has
already tripped it, and nobody has said whether they meant to. Story 7.1's central promise is "the
whole existing fixture corpus re-renders with every hash unchanged" — but you cannot demonstrate that
nothing moved when the alarm is already sounding for a different reason. Build thirteen more stories
on top and the question stops being "what did Epic 9 change?" and becomes "which of fourteen stories
changed this?", with the four-target matrix re-run needed to answer it either way.

**Options considered.** (a) *Proceed and re-point every corpus AC at current HEAD bytes* — the guard
would still stop a new story moving bytes, but the Epic 9 delta stays unexplained and Story 15.1 gets
thirteen stories harder; also it silently ratifies a byte move no human signed off, which is exactly
what AD-22 and the golden test's own failure message forbid. (b) *Proceed with the ACs as written* —
every story's corpus check fails for a pre-existing reason, so the guard is off for the whole run and
a real regression hides inside the existing red. (c) *Explain it first* — chosen by the owner.

**Why this wins.** It is the only option under which Epic 7 and Epic 8's hash-identity criteria mean
anything when they are evaluated, and it matches the sprint sequence already agreed on 2026-08-30,
which named 15.1 "S1 UNBLOCK … doing this after building on top is strictly harder". The accepted
cost is one story of delay before any Epic 7 work starts.

**Consequences.** Epic 7 dispatch is gated on 15.1 reaching `done`. If 15.1 concludes the move was
*wanted*, re-recording the digests re-invalidates the human sign-off across all four documents in
whole — `statement-signoff.json` and the semantic acceptance step must be re-run, not patched. If it
concludes the move was a defect, the fix lands in Epic 9's code and the goldens are left untouched.
Either outcome is a decision this log records before Epic 7 begins.

**How we'd know it was wrong.** If 15.1 finds the +4 bytes/page are inseparable from a change Epic 7
is about to rewrite anyway, the sequencing cost bought nothing — visible as 15.1 halting on an intent
gap that only Epic 7 can settle.

### D-R7.1 — Heavy-test cadence is per-epic, overridden to every-story where correctness is byte-shaped
**Owner decision** (cadence), **orchestrator decision** (the override rule).

**Verdict.** Integration and e2e — Playwright and the four-target hash matrix — run at each epic
boundary: after Story 15.1, after Story 7.7, and after Story 8.6. Unit tests, lint and build run on
every story. **Overridden to every-story for any story whose own correctness is byte-identity-shaped**
— 15.1 by construction, and every story carrying a "the corpus hashes identically" criterion, which
is 7.1, 7.2, 7.3, 7.5, 7.7, 8.3 and 8.4.

**Why.** The override is not caution for its own sake: Epic 9 moved the corpus and it went undetected
across two epics precisely because the heavy check was deferred. A story that promises byte identity
must measure it, or the promise is unfalsifiable. Under a per-epic cadence every deferred suite must
still compile — a Go package that fails to build under its integration tag silently skips its tests,
so the catch-up run would otherwise open with a compile error from ten stories ago.

**Consequences.** Each story's Delivery Log names which suites were measured and which were deferred.
"Green" without that qualifier is a misreport. An epic is not marked `done` until its catch-up run
passes.

### D-R7.2 — Answer channel is the terminal; checkpoints are continuous
**Owner decision.** Questions are asked inline via `AskUserQuestion`, not Telegram, for this run. The
run does not pause at story or epic boundaries; it pauses only for a decision that genuinely needs the
owner. Per-story results are still reported as they land.

### D-R7.3 — Build order is numeric within each epic
**Orchestrator decision** (routine — numeric order already is the dependency order here).

**Verdict.** 15.1 → 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7 → 8.1, 8.2, 8.3, 8.4, 8.5, 8.6.

**Why.** Both epics' stories are written in dependency order. In Epic 7 the typography trio (7.1–7.3)
lands in the engine before 7.4 exposes it in the inspector, and 7.5 must lift the one-page bound before
7.6 can draw the resulting pages or 7.7 can keep a group inside one of them. In Epic 8, 8.1 makes the
chains editable before 8.2 gives them an editor, and 8.3 puts a face in the file before 8.4 renders
from it or 8.6 makes picking one do both. No re-ordering is needed.

**Consequences.** Stories run strictly one at a time; build-auto's previous-story continuity scan reads
the committed, `done` spec of the preceding story, so two in flight would halt the run.

### D-R7.4 — Dispatch mode is classic intent, not folder+id
**Orchestrator decision** (routine — determined by what is on disk).

**Verdict.** Stories are dispatched to `bmad-build-auto` as freeform intent naming the story number.

**Why.** Folder+id dispatch needs `{spec_folder}/stories.yaml`; neither `_bmad-output/specs/spec-folio`
nor `_bmad-output/specs/spec-fonts` has one — they carry `SPEC.md` and companions only. Build-auto
therefore resolves epic context from `epics.md` and writes the spec into
`_bmad-output/implementation-artifacts/`, which is where all 43 existing story files already live.

**Consequences.** Spec filenames follow build-auto's classic derivation (`spec-<slug>.md`), which
differs from the hand-made `<story-key>.md` convention of Epics 1–6. The orchestrator renames each spec
to the existing `<story-key>.md` convention at the plan gate so `sprint-status.yaml` keys and story
files continue to correspond one-to-one.

---

## Run decisions

### D-R7.5 — The corpus move is text alignment, not the element-box paint
**Orchestrator decision** (a correction of fact, not a choice — recorded because D-R7.0 asserts the
wrong mechanism and Story 15.1 would search the wrong file on that brief).

**Verdict.** The +4 bytes per page is **one `Tm` x-coordinate** on the page-footer element, caused by
`folio-go/text_alignment.go` — a file created in `791ed00` with no story, no AC and no FR. Epic 9's
element-box paint moved nothing. D-R7.0's phrase "Epic 9's claim that the corpus stays a witness is
false" is right in outcome and **wrong in mechanism**, and this entry supersedes it on that point.

**The measurement.** The lead attributed it and the orchestrator independently reproduced it — rendered
`statement-1` at HEAD and diffed the content stream against the committed `expected.pdf`. Content
stream `/Length` moves 4548 → 4552, and exactly one run of bytes differs in the whole file:

```
recorded: 1 0 0 1 436     53.88 Tm   <001b002200280026000100060001002e002700010006> Tj
HEAD:     1 0 0 1 522.474 53.88 Tm   <001b002200280026000100060001002e002700010006> Tj
```

Same ten glyphs — `Page 1 of 1` — moved, not changed. `"436"` (3 chars) → `"522.474"` (7 chars) is the
whole delta. The element is `pageFooter` `e4`: `x=400`, `width=123`, `style.align="right"`. The page
footer repeats on every page, which is why the delta is +4 **per page** and scales exactly across the
1-, 5-, 20- and 50-page fixtures.

**In simple terms.** The footer says `Page 1 of 1` and the template has always asked for it to sit
flush right. Until `791ed00` the engine read that request, validated it, round-tripped it, showed it in
the inspector — and then drew the text at the *left* edge of its box anyway, because nothing had ever
been wired to act on it. Epic 9 wired it up. The text moved to where the template always said it
should go, and the number describing its position got four characters longer. Nothing about the
document changed except that one instruction was finally obeyed.

**The arithmetic, which is the test of the attribution.** Element coordinates are *band-relative*; the
absolute position is the 36pt page margin plus `x`. So `e4`'s box spans 436 → 559 absolute, and 559 is
exactly the content right edge (A4's 595.276 less the 36pt right margin, to within the box's own
rounding). Old position 436 = the box's **left** edge — alignment ignored. New position 522.474 leaves
36.526pt of text before that right edge, which is 0.52em per character for ten characters at 7pt: the
correct width for a proportional face. **The new rendering is right and the old was the defect.**

**Population, which closes the red set.** Exactly four templates in the whole corpus declare a
non-`left` `style.align`: `fixtures/statement-{1,5,20,50}/input.folio`. That is precisely the set that
went red — no other fixture could have. And **no fixture anywhere declares `valign`**, so the vertical
half of the same commit shipped with zero corpus coverage. That is a second finding and it is Story
15.1's fourth AC in the literal: "a fixture is added that would have caught it — because no document in
the corpus exercised the path that moved."

**Consequences.** Story 15.1 is briefed with the two byte strings above so it verifies rather than
searches. It must still perform its own diff (AC1 forbids attribution from the commit log alone), but
it must not go looking for a defect in the box paint or in `rectdoc.go` — the `appendEdge` prefix→postfix
fix is length-neutral, and no corpus document strokes an edge. The intended-or-not ruling is the owner's
and is recorded in D-R7.6.

**How we'd know it was wrong.** If re-recording the four goldens does not produce exactly the byte
counts already measured (76,744 / 127,363 / 269,884 / 555,829), something other than alignment is also
moving and the attribution is incomplete.

### D-R7.6 — The alignment change is intended; the goldens are re-recorded and the sign-off re-attested
**Owner decision** (asked at the terminal 2026-08-30, on the evidence in [[D-R7.5]]; the owner chose
"intended, re-record now" over reverting, and over deferring the re-land into Story 7.3).

**Verdict.** `style.align` should render. The pre-`791ed00` output — text drawn at its box's left edge
while the template asked for right — was the defect, and the current output is correct. All four
golden statements are re-recorded to the bytes the engine now produces, in one commit, together with
their `expected.json`, their READMEs and the digest literal in `byte_neutrality_test.go`. The human
sign-off is **re-attested by the owner**, not edited.

**Situation.** Re-recording a golden is the one operation the golden test's own failure text forbids
("DO NOT UPDATE A DIGEST TO MAKE A TEST GO GREEN"), because a digest records that a human once looked
at these pages and accepted them. AD-22 makes a moved digest a versioned behaviour change and C6 makes
an unexplained one a defect. The prohibition is against regenerating *instead of* investigating — it is
not a prohibition on ever moving a golden. The investigation is done, so the question became whether
the new bytes are the ones that should be on record.

**In simple terms.** The reference PDFs are photographs of what the product should look like, and each
one has a signature on the back saying a person checked it. The product has changed in one small way —
a page number that used to sit in the wrong place now sits in the right one — so the photographs are
out of date. You can either take new photographs or undo the improvement. Taking new photographs is
correct, but the old signature was given against the old pictures, so it does not transfer: somebody
has to look again and sign again. What you must never do is take the new photograph and move the old
signature across, because then the signature stops meaning anything.

**Options considered.** (a) *Unintended — revert the alignment rendering.* Rejected: the new output is
demonstrably correct (the arithmetic in D-R7.5 shows the text now ends exactly at the content right
edge), and reverting would make Story 7.3's premise false — 7.3 is written assuming `align` already
renders and adds `justify` to that same closed set, so Epic 7 would inherit an alignment feature it
currently gets for free. (b) *Intended, but revert now and re-land inside Story 7.3*, keeping the
goldens green in the interim. Rejected by the owner: it costs a revert and a re-land to buy a delay,
and it defers the sign-off past the point where `folio-go/v0.1.0` gets cut. (c) *Intended, re-record
now* — chosen.

**Why this wins.** The cost of re-recording is lowest right now and rises sharply: `version.go` still
reads `0.0.0-dev` and `folio-go/v0.1.0` has never been cut, so this is not yet an AD-22 release event.
Once Story 15.3 tags v0.1.0, the same move becomes a breaking change against a published artifact. The
accepted cost is one genuine visual pass over four documents by the owner, which cannot be delegated,
automated, or carried forward.

**Consequences.**
- The sign-off is invalidated **in whole across all four documents** (D-4.7.1), not per-document. The
  story re-records the goldens and then **HALTS** with blocking condition `human sign-off required`,
  leaving `fixtures/statement-signoff.json` byte-identical to `09bb30e` and its test red. That red is
  the story working. No agent may write, edit or synthesise the `reader`, `date` or `examined` fields —
  a filled-in attestation nobody performed is a fabricated record, and the orchestrator obtains the
  real one.
- The four PDFs are rendered to `_bmad-output/implementation-artifacts/evidence/15-1/` for the owner to
  open. The only visible change is the footer's `Page N of M` moving ~86pt right to sit flush with the
  right margin.
- **`valign` still has zero corpus coverage.** The same commit shipped `textValignOffset` and no fixture
  anywhere declares `valign`, so nothing would catch a regression in it. AC4 obliges either a fixture or
  a deferral with a named owner and trigger.
- Epic 7's stories may now rely on `style.align` rendering. Story 7.3 extends the closed align set with
  `justify` rather than introducing alignment.

**How we'd know it was wrong.** If the owner's visual pass finds anything wrong on the re-recorded
pages beyond the footer's new position, the attribution was incomplete and the re-record must stop —
D-R7.5 predicts exactly four moved goldens with byte counts 76,744 / 127,363 / 269,884 / 555,829, and
any other difference falsifies it.

### D-R7.7 — The sign-off gate is unfalsifiable, and this run does not fix it
**Orchestrator decision** (routing, not remediation — the fix belongs to a story this run does not own).

**Verdict.** Story 15.1's review found, and demonstrated, that
`fixtures/statement-signoff.json`'s gate cannot distinguish a genuine re-attestation from someone
pasting four new digests over the old prose. It is **deferred**, not fixed, and this run does not
touch it. Its existence is disclosed to the owner **before** the attestation is requested, which is the
only mitigation available inside this run.

**Situation.** `statementSignOffFieldProblems` (`folio-go/statement_golden_fixture_test.go:237-252`)
checks only that `reader`, `date` and `examined` are **non-empty**. `statementSignOffStaleness` and
`byte_neutrality_test.go` compare only digest equality. A reviewer edited nothing but the four digest
strings — leaving `reader "Panit Wechasil"`, `date 2026-08-27`, and prose describing pages whose footer
sat at the box's *left* edge — and **both gates went green**. The file was restored. Nothing asserts the
record is new: no check that the date advanced, that `examined` or `reader` changed, or that the record
names what it supersedes.

**In simple terms.** The signature on the back of the photograph is checked for being *present*, never
for being *recent* or *about this photograph*. So the cheapest way past the gate is to keep the old
signature and swap the picture — which is exactly the failure the gate exists to prevent, and exactly
what an agent under pressure to turn a test green would reach for. The gate has been this way since it
was written; Story 15.1 is simply the first time anything load-bearing has rested on it.

**Why it is not fixed here.** The intent contract's `Never` assigns CI and gate restructuring to Story
15.2, and this is pre-existing code untouched by 15.1's diff. Widening 15.1 to fix it would mean editing
the very gate whose verdict 15.1 is currently waiting on — a conflict of interest in the literal sense.
The accepted cost is that the gate stays weak until 15.2, and that the integrity of *this* attestation
rests on process rather than on a check.

**Consequences.**
- **No agent in this run may write `reader`, `date` or `examined`.** Recorded in the intent contract as a
  `Never`, and honoured: `git diff 09bb30e HEAD -- fixtures/statement-signoff.json` is 0 lines.
- The orchestrator disclosed the weakness to the owner in the same message that requested the pass, so
  the owner knew the record rests on their actually opening the files.
- Story 15.2 inherits it. It should assert the record is *newer* than the goldens it covers and names
  what it supersedes — not merely that three strings are non-empty.

**How we'd know it was wrong.** A future sign-off whose `date` is older than the goldens it attests to,
or whose `examined` prose describes a rendering the current bytes no longer produce.

### D-R7.8 — The "second permanent red" deferral is transient and closes with the sign-off
**Orchestrator decision** (a correction to a finding, recorded so the closer does not carry it forward).

**Verdict.** Story 15.1's second high deferral — that the story leaves a second permanently-expected red
which `.github/workflows/ci.yml`'s single-scalar `KNOWN_RED_TEST` rule forbids appending to — describes
the **halt state, not the closed state**. `TestGoldenDigestAgreesAtEveryDeclaredSite` is red only while
the sign-off is outstanding. A genuine attestation turns it green, and no second permanent red exists.
The reviewer was right about the halt and wrong about the consequence.

**Consequences.** The closer must verify this empirically — the suite must show **exactly one** red
(`TestCorpusMeetsP6ExerciseFloors`, the mandated P6g floor) once the sign-off lands — and must not carry
the deferral into `deferred-work.md` as an open item against Story 15.2. `KNOWN_RED_TEST` stays a single
scalar and is not touched.

**How we'd know it was wrong.** The sign-off lands and `TestGoldenDigestAgreesAtEveryDeclaredSite` is
still red — which would mean the digests, not the attestation, are the problem.
