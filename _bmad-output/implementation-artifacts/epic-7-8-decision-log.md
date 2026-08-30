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

---

## Story 7.1 — rulings

Story 7.1's plan dispatch halted `blocked` / `intent gap` on three forks. All five entries below are
the engineering lead's RULINGs, verified by the orchestrator where a ruling turned on a quoted fact,
and applied to the spec before re-dispatch. **No owner escalation was needed on any of them.**

### D-7.1.1 — A mandatory break survives inside a declared-unbreakable value, and AD-25 is NOT amended
**Lead ruling**, orchestrator-verified. This was the load-bearing fork and the one that looked like it
needed the owner.

**Verdict.** Exempt **mandatory** breaks from `spansCover`; every **optional** opportunity inside an
atomic span stays suppressed exactly as today. The exemption is keyed on the opportunity's **kind**, at
the filter site (`folio-go/internal/text/opportunity.go:170`) — never on "is this rune `\n`", and never
by shrinking or excluding the span. `atomicSpansFor` (`wrap.go:262`) is unchanged.

**The situation.** `atomicSpansFor` builds a span for every substituted value on a declared
`unbreakableValues` path — by construction, "whatever its script and whatever its content". `spansCover`
then drops any opportunity strictly interior to such a span. So the naive implementation **silently
deletes** the author's break, straight against Story 7.1's AC2. Verified by the orchestrator at
`opportunity.go:170` (`if spansCover(atomic, o) { continue }`) and `opportunity.go:182`.

The apparent conflict is inside 7.1's own `Covers: FR46 · AD-25` line. AD-25's third override says
"no break opportunity survives inside a substituted value from a listed path"; FR46 says a typed break
must survive.

**In simple terms.** A template can mark a field "never break this" — it exists so a Thai surname is not
split down the middle, because 83% of Thai surnames are two ordinary words joined and no algorithm can
tell the join from a word boundary. The question was whether that mark also throws away a line break
somebody deliberately put in the text. It does not, and the distinction is the ordinary one between
*guessing* and *being told*: the rule stops the engine inventing a break where it is unsure, and a line
feed sitting in the data is not the engine being unsure about anything.

**Why no amendment, and why no owner.** D-2.1.6 is an OWNER decision, so amending it would have been an
escalation. It does not need amending. Its Verdict sentence — read at
`folio-mvp-decision-log.md:4959` and confirmed verbatim by the orchestrator — is:

> "A template may declare that a bound value must never be broken. **The engine never proposes a break
> inside such a field.** Declarative, not inferred."

It binds what the engine **proposes**. A line feed present in the input is not the engine proposing
anything, so honouring it is outside what the owner ruled. AD-25's own text scopes the same way: its
three constraints "sit **under** whatever the dictionary proposes, and all override it", and it closes
"the engine's contract is **break opportunities**, not word segmentation" — the third override binds the
opportunity SET, and AC2 defines a mandatory break as expressly not a member of it. The two texts never
meet. Third support: the alternative is a **silent data mutation**, which AD-14 ("never silent") and
D-1.4.9 ("nothing is dropped, nothing is refused") treat as fatal.

**The correction that changes the fixture — this is the part worth the routing.** `bind.Substitution`'s
`Start, End` are rune indices covering **only what a placeholder substituted**
(`folio-go/internal/bind/text.go:37`, confirmed by the orchestrator). So **a line feed the template
author types into an element's literal text is never inside an atomic span at all** — the collision
cannot arise for it. The conflict exists only for a `\n` arriving **through data** on a declared path,
where the "author" is the data supplier. **A fixture that puts `\n` in template literal text proves
nothing and passes vacuously.** The discriminating fixture must supply the line feed through data on an
`unbreakableValues`-declared path.

**Consequences / guardrails.**
- One fixture red-proves both directions: a declared-unbreakable value containing **both** a `\n` and a
  space; assert the `\n` breaks and the space does not. Either assertion alone cannot distinguish
  "exemption works" from "span suppression broke".
- The exemption must be red-provable by flipping the kind test, not by deleting the span — deleting it
  reddens for the wrong reason.
- 7.1 carries a D-000.6 **clarifying** edit: AD-25's third override gains a scope clause saying it binds
  *inferred* break opportunities, not literal control characters in the input. **State the clause; do
  not change the rule.** Same sentence into `folio-format.md`'s `unbreakableValues` prose.
- D-2.1.6's disclosed cost is untouched: a Thai name in free-form text is still guarded only by the
  atomic-unknown-run rule.

**How we'd know it was wrong.** A declared-unbreakable value splitting at anything other than a literal
control character the caller supplied.

### D-7.1.2 — A leading or trailing line feed produces its empty line
**Lead ruling.**

**Verdict.** The separator model applies uniformly. `"a\n"` is two lines, the second empty and occupying
one full `Advance`; `"\na"` is two lines, the first empty. Achieved by **scoping** the two blockers, not
by special-casing: `opportunity.go:124`'s `i > 0 && j < n` guard stays as-is for whitespace and does not
extend to mandatory breaks, and `packLines`' `for start < totalRunes` loop gains the ability to emit a
final zero-length line when the input ends on a mandatory break.

**In simple terms.** If a pasted clause ends with a blank line, the author gets a blank line. Under the
alternative, a trailing blank line is *inexpressible* — you could never produce one — and the character
is thrown away without saying so. Under this rule, an author who wants no blank line just deletes the
newline. Both outcomes stay reachable, which is the test.

**Why this wins.** Expressibility: a reading that leaves an input inexpressible loses to one that does
not. `opportunity.go:124`'s guard is scoped by its own comment to a break that "gains nothing" — for a
mandatory break the empty side is the entire point, so the guard's rationale does not reach it, and
extending it would be the packer *declining* a mandatory break, which AC2 forbids in terms. Story 7.4
makes it concrete: pasted word-processor text routinely ends in a newline.

**Consequences.** Corpus-neutral by construction (no existing fixture value contains a line feed) —
assert that rather than assume it. The empty line occupies one `Advance` and does **not** move
`FirstBaseline`: D-2.5a/DW-15's two-model split must hold identically for empty lines. `textBlockHeight`
must count it — that is the number `textValignOffset` distributes slack against and the quiet path to a
wrong page break. A value that is a single `"\n"` (two empty lines) must not become a zero-line element
or a nil deref.

**How we'd know it was wrong.** An element whose rendered height disagrees with its line count, or a
`\n`-only value producing no line box at all.

### D-7.1.3 — The change lands in the shared packer, for every caller
**Lead ruling.**

**Verdict.** Every caller. One packer, one mandatory-break rule. A `\n` in a table cell's bound data
breaks that cell's line exactly as it does in a text element. No text-element-only flag, no second
packer.

**The question behind it.** Epic 7's header says it "changes nothing about pagination". Does that mean
"does not change the pagination MODEL", or "no template's page breaks may move"?

**Why the model reading wins, from inside the epic.** Story 7.2's fifth AC reads: "the wider line extents
feed `internal/layout` unchanged, **so page breaks follow the spacing rather than ignoring it**." Under
the strict reading that AC is unsatisfiable — an in-epic contradiction. The reading that leaves every
clause standing is that `paginate.go`'s four rules and D-2.6.1's window model are inputs and are
untouched; a page break moving because a line count changed is the window model **working**, not being
changed. Restricting by caller would also mean a second rule for one character, when
`table_render.go:878` already declares it is "the SAME packer text elements use" — a second source of
truth for what `\n` means.

**Consequences.** The spec states the consequence rather than discovering it: a `\n` in bound data
changes a cell's line count, hence its row height, hence where the table breaks. That is intended.
**Re-assert, do not assume, that a row stays atomic** — a `\n` must never become a way to split a row
across pages, and that property is currently held by code nobody is changing, which is exactly when it
breaks. A cell with enough line feeds to exceed the content window is Story 4.6's existing case:
`Pagination.Clipped`, a Warning beside the bytes, **no new diagnostic code**.

**How we'd know it was wrong.** A row splitting across a page boundary, or a new diagnostic code minted
for over-tall cells.

### D-7.1.4 — DW-24 is declined by 7.1, moves to Story 7.3 with two addresses, and its scope is amended now
**Lead ruling**, overriding the orchestrator's initial assignment of DW-24 to 7.1.

**Verdict.** 7.1 declines DW-24. New owner: **Story 7.3**, with the **orchestrator's gate checklist as a
second standing address** — explicitly **not** "Epic 7 close". 7.1 amends DW-24's text in its own record
to add the four `table_render.go` rounding sites.

**Why declined, on the criterion rather than the cost.** DW-24's stated hazard is the unexercised
**rounding** branch — `center`/`middle`, `geom.ScaleRound(slack, 1, 2)` at `text_alignment.go:56`/`:74`,
where a half-to-even tie is what the four-target matrix exists to catch. **7.1 touches neither the
rounding nor the population that reaches it**: no corpus document declares `center` or `valign`, and 7.1
adds none. 7.1 leaves the gap exactly as it found it, so closing it there discharges nothing 7.1
endangered. The `multiple-goals` cost is real but secondary — "a criterion that yields to a budget stops
being a criterion".

**Why not "Epic 7 close".** Precedent: DW-14's owner was "the Epic 2 boundary gate", which ran and closed
without re-owning it, and it survived a whole epic with nobody holding it — D-000.73's class, where an
owner that is an *event* stops existing the moment the event passes. DW-21 already fixed this by naming
two addresses. Use that shape.

**Why 7.3 is a real forcing function.** It extends the closed align set, must author an aligned fixture
anyway, and its slack-remainder rule is itself new integer rounding in the same neighbourhood. A
re-deferral owes a new trigger, and that is one.

**Consequences.** The **scope amendment lands in 7.1, not 7.3** — a deferral whose scope is known wrong
is worse than one merely open, because the literal closing fixture would satisfy DW-24's text while
missing `table_render.go:687/698/1017/1193` and the entry would be marked closed. A centred *text*
element does not cover the table sites: different code, same rounding. **At closure the enumeration must
be re-derived by grep and recorded, not read off the amended hand-list** — the hand-list is being amended
today precisely because it had already rotted once. 7.1's Delivery Log names DW-24 as
inspected-and-declined with this ground, so the next reader sees a decision rather than an omission.

### D-7.1.5 — The break-kind seam is two fields, and Story 7.1 lands both
**Lead ruling**, on an orchestrator finding.

**Verdict.** Two fields, not one. (1) `Opportunity` gains the break **kind** — 7.1 needs it, reads it,
and it is what makes AC2 true. (2) `wrappedLine` gains **the kind of break that ended this line** — 7.1
writes it, Story 7.3 reads it. Both land in 7.1. The last-line case stays **derived** from the index at
7.3's call site and is never stored, so there is one source for it.

**The finding.** `wrappedLine` is `{from, to, width}` (`wrap.go:152`, orchestrator-verified) and
`packLines` holds the winning opportunity in `op`/`chosen` then writes only those three fields
(`wrap.go:217-234`). The information is **destroyed after packing**, and 7.3's AC needs it: 7.3 must not
justify "the last line of a paragraph, **or a line ended by a mandatory break**".

**Why 7.1 carries a field it never reads.** Reconstructing it in 7.3 would be a second derivation of
where a break came from — the hazard `verticalMetrics`' own doc comment says that type exists to close.
It lands at the site that already knows. Precedent and its handling are established:
`verticalMetrics.LastDescent` shipped stating honestly in its comment that nothing in production consumed
it, and was asserted directly by a test over fabricated metrics in the meantime.

**Consequences / guardrails.** 7.1 asserts the `wrappedLine` field **directly** in a test over fabricated
input, and its comment names Story 7.3 / FR47 as the arriving consumer — an unread, unasserted field is
D-000.46's shape and is not authorised. A **named kind, not a bool**: 7.3's rule is two independent
conditions and a bare `!hardBreak` invites the third case to be re-derived wrongly. 7.1's spec states
that 7.3 depends on this field — D-R7.3's numeric order satisfies the dependency by accident, and relying
on an accident is how it gets reordered later.

### D-7.1.6 — Whitespace adjacent to a mandatory break is consumed with it
**Orchestrator decision** (routine — Story 7.1's AC5 settles it unambiguously, so it took the fast path
rather than a lead round-trip. Recorded here because it changes rendered output and the planner
explicitly invited pushback on it).

**Verdict.** A whitespace run containing one or more line feeds emits **mandatory** breaks *instead of*
rule 1's single optional opportunity, and inherits rule 1's consumption at the run's outer edges. So
`"a \n b"` renders as `a` / `b` with neither the space before nor the space after drawn on either line.

**Situation.** The three ruled forks were silent on whitespace *adjacent* to a line feed. Rule 1 in
`internal/text/opportunity.go` walks a whitespace run and emits one opportunity spanning the whole run,
so the run is drawn on neither line. A line feed is itself whitespace, so a run like `" \n "` is one run
today.

**Why this and not "consume only the line feed".** Story 7.1's AC5 says the break character "is consumed
and drawn on neither line, **exactly as a whitespace break already is**" — the comparison is the
instruction. `Opportunity`'s own doc comment already models this: `LineEnd` and `NextStart` "differ only
where the break CONSUMES text — today exactly one case, a run of whitespace", and it states that
modelling the consumed range explicitly is what keeps "the trailing space does not count toward the
line's width" a property of the break rather than a special case at the measuring site. Consuming only
the `\n` would leave a trailing space measured into the previous line — a new polarity nothing asked for,
and precisely the special-case-at-the-measuring-site that comment exists to prevent.

**Consequences.** No new consumption model: the existing `{LineEnd, NextStart}` pair already expresses
it. A run holding *k* line feeds emits *k* mandatory breaks, producing *k−1* empty lines between them,
which is what D-7.1.2's separator model requires.

**How we'd know it was wrong.** A justified or right-aligned line whose measured width includes a
trailing space — the polarity error this avoids.

### D-7.1.7 — Two silent-failure sites the rulings did not name, found at the plan gate
**Orchestrator finding**, verified in the code before accepting the spec. Recorded because both are
places where Story 7.1 would pass its tests while failing its ACs.

1. **A leading empty line is unreachable, and D-7.1.2 did not name the site.** The ruling scoped
   `opportunity.go:124`'s `i > 0 && j < n` guard and `packLines`' loop, which together cover `"a\n"`. They
   do **not** cover `"\na"`: the collection loop at `opportunity.go:165` is `for i := 1; i < n; i++`, so
   `LineEnd == 0` is unreachable and the leading empty line can never be emitted. Verified.
2. **`packLines` bypasses the opportunity list entirely when the remainder fits.** `wrap.go:190` reads
   `if w := measureRuneRange(segs, start, totalRunes, fontSize); w <= maxWidth { … break }`. Verified.
   **This is exactly where AC1 is won or lost** — AC1 requires the break be taken "regardless of how much
   width remained on the line before it", and this short-circuit is the path on which a mandatory break
   inside text that otherwise fits would be silently ignored. A test using text too wide for its box
   would never reach it and would pass green.

**Consequences.** Both are carried in the spec as explicit hazards and task steps. The AC1 test must use
a value that **fits** its declared width, so it exercises the short-circuit rather than the packing loop.

**How we'd know it was wrong.** A `\n` in a short value rendering as one line.

---

## Canvas projection bounds — rulings carried into Story 7.4's plan gate

Raised by the orchestrator at Story 7.1's close, after the implementer deferred `maxCanvasTextLines`
as `low`. The lead **widened the predicate** and found a bound that binds far earlier. Filed as
**DW-25**. All code claims below were verified in the tree by the orchestrator.

### D-7.4.1 — Story 7.4 owns the canvas bounds, and the obligation is not "raise 256"
**Lead ruling**, orchestrator-verified.

**Verdict.** Owner is **Story 7.4**, with the orchestrator's 7.4 plan-gate checklist as a second
standing address — not 7.6, not a new story, and explicitly not "Epic 7 close" (D-000.73: an owner that
is an *event* stops existing when the event passes). The obligation is: **no canvas projection bound may
abort the projection, and the bounds Epic 7's own input makes routinely reachable must admit that
input.** At minimum `page_setup.go:558` and `:456`.

**The finding that changed the question.** The orchestrator raised `maxCanvasTextLines = 256`
(`page_setup.go:27`, enforced `:456`). The lead found that **it is not the bound that stops Story 7.4**:

```go
// page_setup.go:557-560, in canvasComponents — a DIFFERENT function from the paint loop
if element.Type == template.ElementText && element.Value.Set && !element.Value.Null {
    if len(element.Value.Value) > maxCanvasPropertyString {
        return nil, fmt.Errorf("folio: component value exceeds the projection bound")
```

A text element's value is capped at **512 bytes**, and overrunning it returns `nil` for the **entire
component list** — the whole canvas with no components, worse than the `:456` case and firing far
earlier. Verified.

**In simple terms.** 512 bytes is about eighty English words — less than one numbered contract clause.
Story 7.4's promise is "type and paste a multi-paragraph clause". And because `len()` on a Go string
counts **bytes**, not characters, Thai and CJK cost three bytes per character, so the ceiling lands near
170 characters for exactly the two scripts NFR3 makes first-class. An author writing a Thai contract hits
the wall at a third of the length an English author does. To reach the 256-line cap with real prose you
must pass 512 bytes first, so in practice the value cap fires first for every realistic Epic 7 input; 256
is reachable only by a value that is almost entirely line feeds.

**The diagnosis underneath both.** `maxCanvasPropertyString` does **two different jobs**. At `:211`,
`:567`, `:573`, `:617`, `:642`, `:648`, `:663` it bounds **identifiers, colours and expressions** —
legitimately short. At `:558` it bounds **document body text**, which is not. One constant, two
categories. **The fix is to split it, not to change its value.**

**Why 7.4 and not elsewhere.** 7.4 is where the condition first becomes reachable through the product —
the same logic D-000.65 applies to minting. Before 7.4 the properties surface is single-line; 7.4 hands
the author the input, and its first AC is "the editor accepts and preserves multiple lines". An AC that
cannot be demonstrated without lifting these bounds is not a neighbouring concern, it is 7.4's own
acceptance. 7.6 is wrong — it draws pages and neither bound is a function of page count. 7.1's
implementer deferred correctly (7.1's contract forbade designer-surface work); 7.4 is a designer-surface
story, so that constraint is gone. **The severity was wrong, not the deferral.**

**Consequence.** 7.4's Delivery Log must state that it verified an **actual pasted multi-paragraph clause
reaches the canvas** — not that a constant was changed. A constant edit with no end-to-end demonstration
is the shape that lets this survive.

### D-7.4.2 — Degrade per element; never abort; mint no diagnostic code
**Lead ruling.**

**Verdict.** Shape **(b)** — degrade the offending element, keep the projection. The pattern already
exists in the tree eleven lines above the site in question and must be reused, not reinvented:
`page_setup.go:428-435` handles a `fontChain` failure by setting
`component.TextPaint = &CanvasTextPaint{Lines: []CanvasTextLine{}}` and `continue`ing, with the rationale
"Existing designer documents can be structurally valid while incomplete for production rendering … They
remain loadable; there is simply no honest measured paint to display yet." Verified; that reasoning
covers this case verbatim.

**Options rejected.** (a) *Raise the constant* — rejected on the criterion, not the cost: moving a
threshold so a symptom stops reporting is the twin of manufacturing data to meet a floor. **The cliff is
the defect; its position is not.** (c) *A document-level limit with a located diagnostic* — rejected
because there is no code to mint. The render path has **no such cap**: a 400-clause document renders
correctly to PDF, so there is no render-time condition and D-000.65 never fires. Worse, it would let the
**canvas refuse a document the engine renders**, inverting AD-15/AD-16 and the
canvas-approximate/preview-exact asymmetry that is the product concept. **A canvas projection bound must
never become a document validity rule.**

**Guardrails — the first two carry most of the weight.**
1. **Truncate the paint, never the value.** `CanvasTextPaint.Lines` is regenerated every projection and
   never written back, so bounding it is safe. `component.Value` is what the properties panel edits and
   **saves** — truncating it would write the truncation back into the document and destroy the author's
   text. So `:558` gets body text its own genuinely large bound (a channel-size concern, megabytes, not
   an identifier bound), and the degradation lives on the paint side alone.
2. **The degraded state must be distinguishable from the empty one.** As written, a 400-line element and
   an element with no text both project `Lines: []` — the all-clear wearing the face of could-not-look.
   Add a field beside the existing `Overflow bool` (already the precedent for a paint-plan flag naming a
   degraded state) and **paint the first N lines rather than none**, so the author sees their text and is
   told it is cut. A projection field, not a `diag.Diagnostic`; no registry entry.
3. **Split `maxCanvasPropertyString`.** It stays 512 for identifiers, colours and expressions (all seven
   other sites are correct at that size). Body text gets its own named constant.
4. **Whatever number replaces 256 must be derived and recorded, not chosen round**, with the criterion in
   the constant's comment. Epic 7's narrative target is "forty pages"; at ~45 lines per A4 page at 11pt,
   256 lines is about six pages — short of the epic's own stated target by most of an order of magnitude.
5. **Verify, do not assume, that truncating the paint leaves Story 7.5's window count intact.** 7.5's AC
   has the projection report "how many windows the column currently occupies"; that must come from layout,
   not from `CanvasTextPaint`. If a truncated paint can shorten the reported column, 7.6 draws the wrong
   number of sheets and **the canvas lies about pagination**. Assert the independence.
6. The other seven `maxCanvasPropertyString` sites keep aborting, and that is **recorded, not fixed**:
   Epic 7 makes none of them newly reachable. Naming the population is what stops the next reader
   believing the whole class was closed.

### D-7.4.3 — The three canvas constants are not a coordinated budget, and 256 protects nothing
**Lead ruling**, answering the orchestrator's question of whether raising one number was safe.

**Verdict.** `maxCanvasTextLines`, `maxCanvasPropertyString` and `maxCanvasTextFragments` are **three
independent defensive caps that happen to share a const block** — not a projection-size budget.
`maxCanvasTextLines` protects no invariant and may be raised or replaced on its own merits.

**Grounding.** They bound three different axes at three different granularities: a string length at eight
sites; fragments **per line** (`:499`); lines **per element** (`:456`). Their product, 256 × 512 × 512, is
not a coherent total anyone budgeted. The only representational invariant in the file — nothing
unrepresentable crosses the JS boundary (AD-15/AD-16) — is already enforced **per value** by
`canvasLineTop` and `canvasDerived` against `MaxCanvasMillipoints`, each with its own error. 256 lines at
a typical ~14,000 mp advance is ≈3.6e6 against 9.007e15: **ten orders of magnitude of headroom**. What 256
does bound is projection payload size across the worker channel — a real concern and a legitimate reason
to keep *a* cap, but a payload budget **degrades, it does not abort**, which is D-7.4.2.

### D-7.4.4 — Watch item: the canvas breaks the raw template string, not bound data
**Lead observation**, not a decision. Recorded so it is not discovered inside a story that assumed
otherwise.

`page_setup.go:449` calls `atomicSpansFor(t.doc.UnbreakableValues, nil)` — the canvas passes **nil**
substitutions, so **no atomic span ever exists on the canvas**, and the canvas breaks the raw template
string (`{{...}}` placeholders and all) rather than the bound data. Correct as designed, and untouched by
D-7.1.1, which has no canvas-side effect.

**Consequence.** Stories 7.4 and 7.6 both lean on canvas/PDF agreement. **If any story claims the canvas
shows where the engine will break a _bound_ value, that claim is false today.** To be raised at 7.6's plan
gate rather than decided now.

---

## The `.folio` format version — rulings and the owner's reframing

### D-R7.9 — Nothing is tagged, so the MAJOR bump is free and no release is scheduled
**Owner decision** (asked at the terminal 2026-08-30). **The owner REFRAMED the question rather than
choosing an option, and the reframing is better than any of the four offered.**

**Verdict, in the owner's words:** *"We haven't released anything to production yet so no need to tag
now."* So Epic 7 takes the MAJOR bump to `.folio` 2.0 freely at Story 7.3, Epic 8's format change joins
the same 2.0 at Story 8.3, and **no release is cut in or around this run**. Story 15.3 is not imminent
and must not be treated as a deadline.

**Why the reframing beat the options.** All four options I offered assumed a tag was coming and argued
about where to put it relative to the epics — the lead's recommendation was "land both epics, then tag",
and two of the remaining three were about paying to avoid a MAJOR that a tag would make permanent. The
owner removed the premise. **There is nothing to sequence against.** The load-bearing fact the lead
found — that a MAJOR costs per *released version*, not per feature — becomes unconditional rather than a
race: with no tag pending, the second format extension is free, and so is the third. That dissolves the
"cheaper before the tag" urgency that framed the whole question, and it removes the trade-off the lead's
own recommendation had to accept (a slipped tag).

**In simple terms.** A format MAJOR is expensive because it orphans everyone already using the old one.
Nobody is using this yet — no tag, no consumers, `version.go` still reads `0.0.0-dev`. So the change that
would be costly later is free now, and the right move is to make it properly rather than to design around
a cost that does not exist yet.

**Consequences.**
- Story 7.3 raises `SupportedMajor` to 2 without further escalation. Story 8.3 joins the same 2.0.
- **The build order does not change.** 15.3 was never in this run's scope; this only removes the pressure
  to reach it.
- Options 3 (`style.justified` as an additive key) and 4 (cut `justify`) are **rejected** and should not
  be re-proposed: both existed only to dodge a cost the owner has just said is not being incurred, and
  option 3 additionally reintroduces the silently-wrong render D-1.4.12 exists to prevent.
- When 15.3 is eventually planned it must re-make the release decision from scratch — its own AC already
  says the "cut after Epic 6" trigger has been overtaken by events. This entry is now part of what it
  re-makes.

**How we'd know it was wrong.** Someone outside the project starts depending on `folio-go` from `main`
before a tag exists — at which point the MAJOR stops being free and this decision needs revisiting.

### D-7.3.1 — `justify` extends a closed set named literally in D-1.4.12, so Story 7.3 is a MAJOR-bump story
**Lead ruling**, orchestrator-verified. **This also corrects the lead's own grounding report.**

**Verdict.** D-1.4.12 reaches `align` **by name**, with nothing to interpret. Its verdict is the mandatory
`folio-format.md` sentence and it enumerates the eight closed sets literally: element `type`, `locale`,
**`align`**, **`valign`**, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`.
Extending any of them is MAJOR "because every existing library validates those sets as load errors".
Verified in code: `closedAligns = {left, center, right}` at `internal/template/closedsets.go:30`,
enforced as a load error at `parse_bands.go:399` and `:531`. A 1.0 library hits `justify` and refuses the
file — D-1.4.12's stated mechanism exactly.

**The correction.** The orchestrator hypothesised a distinction between extending a **value** set and
extending a **shape** set, which would have let `align` through as MINOR. **It does not exist, and the
asymmetry runs the other way:** all eight named sets are value sets, so `align` is the textbook case,
while Epic 8's chain-entry *shape* set is not on the list and reaches D-1.4.12 only by analogy. The lead's
grounding report had framed 8.3 as the D-1.4.12 instance and missed that 7.3 is the cleaner one — the
lead owns that correction. **The real instance is four stories earlier than anyone thought.**

D-1.4.12 had already considered and rejected per-set treatment, with `align` named in the rejection: "on
inspection every one of the eight genuinely cannot be faked by an older library … an unknown `align`
would be silently wrong." And its own falsifier fires precisely here: "How we'd know it was wrong. A
MINOR release adding a closed-set value."

**Guardrails for Story 7.3, both concrete and both verified.**
1. **`closedAligns` is a single map SHARED between `style.align` (`parse_bands.go:531`) and
   `columns[].align` (`:399`).** Adding `justify` to it silently legalises `columns[].align: "justify"` —
   justified table cells, which Epic 7 never specified and nothing implements. **Split the set:**
   `columns[].align` keeps `{left, center, right}`; only the style set gains `justify`. Justified columns
   would be a separate scope decision, not a side effect of a map edit.
2. Both call sites carry a **hand-written** error string, "not one of the closed set left, center, right".
   Derive the message from the set, or 7.3 ships two error messages that lie about what is legal.

### D-7.2.1 — `style.lineSpacing` is MINOR, and Story 7.2 inherits an unrecorded version debt it must discharge
**Lead ruling**, orchestrator-verified.

**Verdict.** `style.lineSpacing` is a new optional key with no closed set, so D-1.4.12 does not reach it
and D-1.4.9's "may add new optional keys only" makes it **MINOR**. But 7.2 cannot merely declare that —
it is the first story able to discharge three things nobody has, and it must.

**The debt, measured in the tree.**
1. `SupportedVersion = "1.0"`, `SupportedMajor = 1` (`internal/template/version.go:24-25`), and **all 18
   fixtures declare `"1.0"`. The format version has never moved since Story 1.4.**
2. **Epic 10 shipped `style.color` and bumped nothing** — its own prose says "This epic adds one optional
   field, `style.color`". So documents using colour **declare 1.0 while requiring 1.1**. An older 1.0
   library loads them, ignores the key, and renders the text **black**. That is a silently-wrong render
   arriving through the back door — the exact failure D-1.4.12's rejected option (b) would have allowed.
3. **`versionForSave` is still a stub.** `version.go:88` returns `loaded` unchanged, and its own comment
   says why: "At Story 1.4 the raise path is unreachable — no 1.1 exists yet … so that the day a future
   MINOR exists, the raise path has exactly one place to be filled in." **That day is Story 7.2.**
   Bumping `SupportedVersion` alone changes nothing; without the raise path no document is ever raised and
   the constant is decorative.

**Consequences.** Story 7.2 owns: implement D-1.4.13's raise path at the one place reserved for it; raise
a document to 1.1 when it introduces `lineSpacing`; and **retrofit the same for `style.color`** so Epic
10's documents stop misdeclaring. Cheap now.

**A worry that dissolves — Epic 7 does NOT move the version twice.** Under D-1.4.13 `version` is a
property of the **document**, raised only by the content it actually contains: a document using only
`lineSpacing` is 1.1, one using `justify` is 2.0, one using neither stays 1.0, and all three coexist. What
moves twice is the *library's* `SupportedMajor`/`SupportedVersion`, which is a library fact, not a format
event.

**Two version systems, not to be conflated** (they were being, in this run and elsewhere):
- **AD-22 governs the LIBRARY version** (`folio-go`): any change to layout, subsetting, emission or the
  toolchain is breaking. Epic 9's alignment change was an AD-22 breaking change — it moved goldens and
  touched no format key.
- **D-1.4.9 / D-1.4.12 / D-1.4.13 govern the FORMAT version** (`.folio`). Epic 9 changed no format version
  and was right not to.

**Guardrail for whichever story bumps `SupportedMajor`.** Its doc comment calls it "the full MAJOR.MINOR
this library itself would author for a brand-new document". Once that reads `"2.0"`, a blank document
would declare 2.0 while using no 2.0 feature — needlessly orphaning it from every 1.x reader and
contradicting D-1.4.13's raise-only-by-content rule. **A new document must declare the lowest version its
content requires, never the library's ceiling.**

---

## Story 7.2 — rulings

Story 7.2's plan halted `intent gap`: AC7 rejects a `lineSpacing` "outside the supported range" and
**nothing anywhere says what that range is**. Four lead rulings, all orchestrator-verified in code.

### D-7.2.2 — The canvas clause that forbids tight leading is the defect; delete it
**Lead ruling**, orchestrator-verified. This one removes a guard, so it was checked before acting.

**Verdict.** In `folio-designer/src/engine-protocol.ts:207`, remove **`paint.baseline > paint.top +
paint.advance`** from the predicate. Everything else on that line stays.

**Situation.** The canvas emits `baseline = top + FirstBaseline` while `advance` is the scaled value, so
that clause reduces to **`FirstBaseline <= Advance`** — the unscaled relationship `wrap.go:600-612`
guarantees by clamping `maxDescent`/`maxLineGap` at zero. Verified at `page_setup.go:485-497`:
`baseline = canvasDerivedSum(top, vm.FirstBaseline)`, `advance = vm.Advance`. It is **not an independent
property**; it is a restatement of an engine invariant that `lineSpacing` deliberately dissolves, sitting
on the wrong side of the channel. The predicate lives inside `isTextPaint`, so one bad line fails
`isTextPaint` → `isCanvas` → `isSnapshot` and blanks the **whole projection**. Measured cliff on the
shipped chain at 12pt (`FirstBaseline: 11759`, `Advance: 14982`): **784 thousandths rejects, 785 passes.**

**In simple terms.** `FirstBaseline > scaledAdvance` means one line's baseline sits below the next line's
top — the line boxes overlap. **That is what tight leading is**, and it is exactly what the PDF will
draw. A guard refusing it is stating a typographic opinion, not checking a consistency property.

**Why removing it is the AD-17 fix, not an AD-17 breach.** AD-17 says the canvas takes every text metric
from the engine and the browser contributes rasterization only. This clause has the **browser refusing
the engine's own honest measurement** and blanking the canvas — the invariant inverted. Story 5.9's guard
is about *where metrics come from*, not about the browser adjudicating them.

**What replaces it: nothing new. The real invariants are already in the same predicate and all survive.**
`paint.advance <= 0` (genuinely required, unaffected); `paint.baseline < paint.top` (still holds —
`FirstBaseline` is `maxAscent` clamped at zero and `lineSpacing` scales only `Advance`); the
`Number.isSafeInteger` checks (the actual JS-boundary concern); and `paint.top < priorTop + priorAdvance`
— **orchestrator-verified not to become the next cliff**: `top = canvasLineTop(originY, i, vm.Advance)`
so `top_i = originY + i·A`, making the test `originY+i·A < originY+i·A`, false for any positive advance.

**How we'd know it was wrong.** If browser paint code ever derived a line's top from `baseline + …`
rather than from the engine's supplied `top`, overlap could corrupt layout. It does not — `top` is
supplied per line.

### D-7.2.3 — The load-time range is representational and encodes no typographic opinion
**Lead ruling.**

**Verdict.** `style.lineSpacing` is a whole number of thousandths in **[1, 1000000]** (0.001–1000.0
inclusive). Outside that, or not a whole number of thousandths, is a located load error naming the
element. **No lower bound at 1.0.**

**Why not a typographic bound.** A load-time check **cannot see the font size**, so it cannot express
one: the canvas cliff is 785 thousandths at 12pt on the shipped chain and *moves with face and size*. A
bound pretending to be typographic while blind to the input that determines it is wrong at every size but
one. So the load-time range is the only thing load time can see — the value's own domain. A 1.0 minimum
was also rejected on product grounds: Epic 7 exists to produce a filed contract's house style, and tight
leading is part of that. Asymmetry with `fontSize` (exactness and overflow checks, no range) is a reason
*not* to invent a bound, not a reason to add one. The 1000.0 ceiling is a **stated sanity ceiling, not a
derived safety bound**, and must say so in the constant's comment.

**Guardrail — the real lower-bound failure is not load-time checkable either.** `ScaleRound(400, 1, 1000)`
is **0**: a small face at `lineSpacing: 0.001` yields `advance = 0`, which the canvas correctly rejects
and which gives layout zero-height lines. Check it **where both operands exist**, in the leading model, as
a located error naming the element and the resolved size. Do not raise the minimum to prevent it — that
just moves the blindness. One validation function, called from both the load path and the property-command
path, so a value refused in a file is refused in the inspector for the same reason.

### D-7.2.4 — Overflow is a separate guard at the computation site, and it is not a number
**Lead ruling.**

**Verdict.** Not the same bound and not a load-time bound at all. Guard `ScaleRound`'s precondition
**where it is called**: check `int64MulOverflows(Advance, r)` before the call in the leading model and
return a located error. **A panic must never be reachable from authored input** — `geom.ScaleRound` panics
on int64 overflow (`internal/geom/scale.go:67-69`, verified).

**Why not at load.** The precondition cannot be discharged there: `Advance` is unknown and `fontSize` is
unbounded. Deriving a load-time ceiling honestly from the extreme case gives `r <= 1023` — i.e. it would
forbid `lineSpacing` above 1.0, which is absurd. That reductio is why D-7.2.3's ceiling is a sanity bound
and nothing more. This is **D-1.5.2's shape** (`unitsPerEm`): validate at the trust boundary, and add **no
non-panicking variant** of `ScaleRound` — AD-2 says scaling is one function and a second door drifts.

**Stakes.** Story 7.1's closer found that a Go panic aborts the package binary, so every other test in
`folio-go` silently stops reporting. **A panic reachable from a template is not a crash, it is a
suite-wide blindfold.**

**Guardrail.** `fontSize`'s missing range check is the pre-existing half of this. 7.2 **records it in
`deferred-work.md` rather than closing it** — closing it is a format-domain decision on a second field and
would earn `multiple-goals`. Name it, noting the overflow site is shared.

### D-7.2.5 — Mint `STYLE_LINE_SPACING_INVALID`
**Lead ruling** under D-000.65 (mint where the condition first occurs).

**Verdict.** Mint `STYLE_LINE_SPACING_INVALID` in `internal/diag`, `SeverityError`, with the public bridge
`folio.DiagCodeStyleLineSpacingInvalid`, raised from the single validation function both the load path and
the property-command path call.

**Why minting is what makes AC7 reachable at all.** `wasm/cmd/engine/main.go:271-280`'s
`reportableMessage` replaces the message with "The template could not be processed" **only** for
`DiagCodeTemplateMalformed`; every other code gets `bounded(message, 512)`. Every load-time style
rejection today is uncoded and becomes `TemplateMalformed`, so **an uncoded `lineSpacing` error is
destroyed before it reaches the author** and AC7's "a located load error naming the element" is
unsatisfiable through the designer. A dedicated code is not registry hygiene here — it is the mechanism
that delivers the AC. The destruction rule's own stated reason does not reach this case: it exists because
a malformed-template message "quotes the offending document back", and an engine-authored message naming
an element id and a numeric range quotes nothing back. `STYLE_COLOR_INVALID` is the standing precedent.

**Guardrails.** Assert the code is **not** `TemplateMalformed` and that its message survives
`reportableMessage` intact — without that, someone later routes it through the generic load-error path and
AC7 fails silently, invisibly from the Go side where the tests are. The code names **the field's value
being outside its declared domain** — the zero-advance (D-7.2.3) and overflow (D-7.2.4) errors are
different conditions at a different stage and **must not be folded into it to save a mint**.

**Forward note, not a decision.** `STYLE_COLOR_INVALID` plus this makes two per-field style codes. Before
a **third** is minted, someone must decide whether the general form is right or AD-14's closed registry
accretes one entry per style field forever. Raise at Epic 11 or 12 planning, whichever first proposes one.

### D-7.2.6 — DW-24 declined a second time, on a narrower ground, and is not deferrable a third
**Lead ruling**, correcting the orchestrator's and planner's reasoning while accepting the outcome.

**Verdict.** DW-24 is inspected-and-declined for 7.2. **But not for the reason given.** The planner argued
"a different call site, denominator and file". In fact **7.2 does change the input to the unexercised
rounding site**: `textBlockHeight` is built from `Advance` and feeds the slack that `ScaleRound(slack, 1,
2)` halves for `valign: middle`. The exposure is unchanged only because **no fixture declares `valign` at
all** — which is the same absence DW-24 exists to record. The decline holds on that narrower ground.

**Condition.** **DW-24 is not deferrable a third time.** Story 7.3 is its owner (D-7.1.4); 7.3's plan gate
treats it as an **acceptance criterion, not a deferred item**, and a decline there is an escalation to the
lead rather than another entry. A trigger that has failed to fire twice stops being a trigger.

### D-7.4.5 — The TypeScript mirror is three hand-copied constants, not one
**Lead ruling**, extending D-7.4.2 after an orchestrator observation.

**Verdict.** `engine-protocol.ts:201-212` mirrors **three** Go constants with no shared source:
`value.lines.length > 256`, `fragments <= 512`, and `fragment.text.length <= 512`. This is the drift
pattern in pure form — the Go side can be raised and the TS side will silently keep rejecting, blanking
the projection with **no error anyone can attribute**.

**Consequence, added to DW-25.** Whichever story changes a Go bound changes its TS mirror **in the same
commit**, and lands **an assertion tying the two** — a test that reads both, not a comment asking the next
person to remember.

### D-R7.10 — The run continues story-to-story without check-ins, through Epic 8
**Owner decision** (given 2026-08-30, mid-run): *"make sure the dev cycles continues after each user
story until finishing epic 8 unless there are important decisions I need to make."*

**Verdict.** The orchestrator chains plan → build → close for every remaining story (7.4–7.7, then
8.1–8.6) without pausing to report between them. It stops **only** for a decision that genuinely
requires the owner. This supersedes the "run continuously, pause only at design decisions" setting from
D-R7.2 by making the endpoint explicit: **Epic 8 complete.**

**Why it was given.** The orchestrator twice ended a turn saying it would continue to the next story and
did not dispatch it — the run stalled with nothing running. That is the failure this instruction closes.

**What still counts as worth stopping for**, so the instruction is not read as "never ask":
- A ruling that touches the owner's own artifacts — the golden sign-off is the standing example, since
  no agent may write a `reader`/`date`/`examined` attestation.
- A scope or release decision with lasting consequence (the D-R7.9 class).
- A defect that falsifies a decision already made.
- An `OWNER-DECISION-NEEDED` block the engineering lead declines to rule on.

Everything else — lead rulings, spec amendments, deferrals, halts the lead can settle — is applied and
logged without interrupting the owner, and reported when the run next surfaces.

**Consequence.** Reports become epic-boundary-shaped rather than story-shaped. The per-story record does
not degrade: each story still gets its `## Delivery Log` entry and its decision-log entries, so the audit
trail is unchanged — only the interruptions stop.
