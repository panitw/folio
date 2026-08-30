# Folio MVP — Deferred Work

Work consciously **not** done, with a named owner and the story that deferred it. This file exists
so a deferral is a tracked decision rather than a silent gap. Append-only; mark items done in place
with the commit that closed them.

Owners: **Story N.M** = the named story picks it up as part of its own scope. **Epic N close** = due
before that epic's key is marked `done`. **Owner** = needs the project owner's call before anyone acts.

---

## Done

### DW-1 — A fixture-path override so shape-check red-proofs never touch the real golden fixture — **DONE**
- **Deferred by:** Story 1.2 (reviewer Finding 7, finisher DEFER)
- **Owner:** **Story 1.3** — it is the guardrails story, and this is a guardrail's testability gap
- **Committed at:** `f9c27b3`
- **Closed by:** Story 1.3's finisher commit (this file's own commit; see that commit's message for
  the SHA — a commit cannot self-reference its own hash in its diff). `folio-go/matrix_test.go`'s
  `loadExpectedFixture` was split into a pure `checkFixtureShape(path string) (expectedFixture,
  error)` taking the fixture's location as a parameter, plus `TestFixtureShapeCheckRedProof`, which
  red-proves AC16's shape check against scratch copies (sha256 widened into a per-target object,
  **and** sha256 of the wrong length or case — Finding 12, this story's QA review) and asserts
  byte-identity of the real `fixtures/minimal-rect/expected.json` before and after (`bytes.Equal`)
  — the real fixture is never mutated.
- **Correction (Blocker 3, this story's QA review):** the developer's first pass placed
  `checkFixtureShape`, `loadExpectedFixture` and `TestFixtureShapeCheckRedProof` in
  `folio-go/matrix_test.go`, which carries the `//go:build matrix` tag — so the red-proof executed
  in **zero** gates (the tagged suite is deferred to the Epic 1 boundary, D-000.4). DW-1 was marked
  DONE on that strength anyway, which was the defect. The finisher moved all three into the
  untagged `folio-go/fixture_test.go` (package `folio`'s ordinary home for fixture-shape helpers,
  alongside `isSHA256HexString`); `go test ./...` now runs `TestFixtureShapeCheckRedProof` on every
  story, measured. DW-1 is marked done here only now that its proof runs in a gate that runs.

**What was deferred.** AC16's fixture-shape check (`sha256` must be a 64-character lower-case hex
JSON *string*) has no way to be red-proved against a scratch copy of `fixtures/minimal-rect/expected.json`,
because the check reads the real fixture by a hard-coded relative path. The finisher re-verified the
property by exercising `isSHA256HexString` directly through a temporary, uncommitted test, and
deliberately never mutated the real golden fixture to produce a red.

**Why it was deferred rather than fixed.** Adding a path override is a change to how every fixture
consumer resolves its path — a seam that AD-21 and D-000.5 both constrain (repo-root `fixtures/`,
read by relative path at test runtime, never `go:embed`ed). Doing it inside Story 1.2, whose
deliverable is the matrix harness, would have widened that story into fixture-resolution design
with no review coverage for it.

**Why it matters.** AC16 exists to stop a developer meeting a red matrix leg from converting `sha256`
into a per-target map (D-1.2.2). A guard that cannot be shown going red is a guard nobody can trust,
and this one protects the project's central reproducibility claim.

**How we'd know it was forgotten.** A future story adding a second fixture-shape assertion and
red-proving it the same indirect way, or not at all.

---

### DW-15 — First-baseline placement and inter-baseline spacing use two different models — **FIXED**

- **Raised at:** Story 2.5 (creator), ruled a defect by the lead at that story's DN-3.
- **Owner was:** engineering lead to schedule; *must land BEFORE the Thai reading sign-off is recorded.*
- **Closed by:** **Story 2.5a**, `2-5a-align-first-baseline-with-the-leading-model`, which exists for
  no other purpose. Commit: this file's own commit (a commit cannot self-reference its own hash).
- **Ends FIXED, not "partially addressed"** ([[D-000.29]]). **No open question is carried forward
  under this number.**

**The defect, as it was.** `internal/pdf`'s `buildTextContentStream` placed the first baseline with
`pdfY = flipY(..., run.FontSize)` — from the **point size**. [[D-2.4.2]] derived inter-baseline
spacing from the `hhea` metrics of the **declared chain**. So the *first* line was positioned on one
model and *every subsequent* line spaced on another, and the two agreed only by coincidence.

**Corrections to this entry's own text, measured at `17f5f7a` before the fix.** All three were wrong
as written, and each would have misdirected the work:

| this entry said | measured |
|---|---|
| the defect is at `internal/pdf/textdoc.go:730` | **`:689`**. `:730` is inside `appendShapedRun`, which places nothing. |
| correcting it re-records **four** goldens | **FIVE.** The fifth is `fixtures/font-text/`. |
| "a Thai or CJK chain diverges more than a Latin one" | True in magnitude, **incomplete in sign** — see below. |

**THE SIGN, which this entry got wrong by omission and which is the finding most worth keeping.** The
error is **not** a consistent downward drift. It is `max(hhea ascent) − 1000` units per em, and
`hhea` ascent is not always above the em: Roboto-Regular's is **928**. So `fixtures/font-text/`'s
baselines moved **UP** by 1.008 pt at 14 pt while every Noto-chain baseline moved **DOWN**. This is
now [[D-000.45]], binding: **assert the computed value from a declarative table, never a direction.**
A guard phrased *"the baseline sits lower than the font size implies"* is **false on a fixture that
ships.**

**PRE-FIX MEASUREMENT** ([[D-000.30]] — captured before any production file was edited, because the
window in which it is constructible closes permanently). Read from the `hhea` table via
`(*fontset.Font).LineMetrics()`, scaled to the 1000-unit em:

| face | A | D | gap | A − D + gap |
|---|---|---|---|---|
| Noto Sans | 1069 | −293 | 0 | 1362 |
| Noto Sans Thai | 1061 | **−450** | 0 | **1511** |
| Noto Sans SC | **1160** | −288 | 0 | 1448 |
| Roboto-Regular (test face) | **928** | −244 | 0 | 1172 |

Per shipped chain and size, in millipoints — `now` is `run.FontSize`, `aligned` is
`ScaleRound(max(A), size, 1000)`:

| document · element | chain | size | now | aligned | Δ |
|---|---|---|---|---|---|
| `font-text` e1 | Roboto | 14 pt | 14 000 | 12 992 | **−1 008 (UP)** |
| `font-text` e2 | Roboto | 9 pt | 9 000 | 8 352 | **−648 (UP)** |
| `multi-script-fallback` e1 | Noto ×3 | 14 pt | 14 000 | 16 240 | +2 240 |
| `shaped-text` e1…e7 | Noto ×3 | 16 pt | 16 000 | 18 560 | +2 560 |
| `three-band-page` e1, e2 | Noto Sans | 12 pt | 12 000 | 12 828 | +828 |
| `three-band-page` e4 | Noto Sans | 9 pt | 9 000 | 9 621 | +621 |
| `three-band-page` e3 | Noto Sans | 8 pt | 8 000 | 8 552 | +552 |
| `wrapped-text` e1…e4 | Noto ×3 | 11 pt | 11 000 | 12 760 | +1 760 |

**THE RULING TAKEN, and why it was not a judgment in the end.** This entry's owning story raised DN-1
— *whose* ascent places the first baseline — as a blocking choice between `max(ascent)` (1160) and
the ascent of the face that won the advance maximisation (1061). [[D-2.4.2]] **(amended)** dissolved
the question: the first span is the accommodate-what-may-appear argument asked about the ascent axis
alone, so its answer is `max(A)` by the same reasoning the other spans use. That the two candidates
coincided **on the shipped set only** is [[D-000.32]]'s fitted-to-the-sample hazard one ruling away
from where it was named.

**WHAT ELSE LANDED WITH IT, and why that is one cause rather than two** ([[D-2.5a.1]]). The story's
DN-3 reported that D-2.4.2's original `max(A − D + gap)` maximises the **worst single face** when the
constraint is the **worst adjacent pair** — a 99-unit shortfall on the shipped chain. That was ruled
**the same defect, not a second subject**: both are the vertical model using a **proxy** instead of
accommodating the declared chain. So D-2.4.2 was amended and both landed together:

| span | value |
|---|---|
| top → first baseline | `max(A)` |
| baseline → baseline | `max(A) + max(D) + max(gap)` |
| last baseline → bottom | `max(D)` |

**DN-3 therefore does NOT open a new DW entry.** It is fixed here. Neither does DN-4 (`splitByFace`,
zero call sites, six comments calling it the live placement path): [[D-000.46]] ruled that dead code
misdescribing the system is worse than dead code, and it was **deleted** with all nine of its
references corrected.

**Blast radius, measured rather than assumed.** The advance moves **only for a multi-face chain** —
for a single present face `max(A)+max(D)+max(gap)` is identically `A − D + gap` — and only
`wrapped-text` has multi-line elements on such a chain. Confirmed by mutation: reinstating the
superseded advance rule reddens `wrapped-text` and **nothing else**. Wrapping itself cannot shift:
`packLines` takes no vertical quantity and runs before the model is computed.

**Five goldens re-recorded** as one attributable movement under AD-21/AD-22, each with a D-000.22
semantic acceptance step read off the **new** artifact ([[D-000.44]]: a re-recording is a recording,
and three of the five had no such step at all before this story).

**The sequencing obligation was honoured.** The fix landed **before** the Thai reading sign-off was
recorded, and Story 2.5a **did not** request it — see `epic-2-boundary-gate.md`. The Thai *break*
sign-off is unaffected, and that is now **measured** rather than assumed: no break-related test moved.


---

### DW-3 — Publishing the third-party licence manifest as a release artifact — **RETIRED at Epic 4 planning (D-4.0.1)**

**Retired, not completed, and the distinction is the point.** AD-26's substance shipped at Story 1.3
and is guarded continuously: every module in the resolved graph carries a resolved licence, an
unresolvable one fails the build, and `TestManifestUpToDate` fails if the committed
`lint/MANIFEST.md` drifts from what the generator produces. What remained was *"attach it to a
release"* — **a line item in a procedure that did not exist**, not deferred engineering. Leaving it
under `## Open` asserted that something remained to be **built**. A record that overstates is a defect
even when it errs toward caution ([[D-000.49]]); this one spent attention at three consecutive gates.

**The two disagreeing owners were the tell.** *"Epic 4 close"* and *"the `folio-go/v0.1.0` tag"* were
one moment when written and are three epics apart since [[D-000.78]]. The lead declined to fix that by
picking one, because either choice keeps finished work in a backlog.

**Discharged by replacement** ([[D-000.59]]): the obligation now lives in **`RELEASING.md`** at the
repository root, item 1 — *where the story that cuts the tag must necessarily read it*, rather than in
a list that story has no reason to open. `TestReleasingDocNamesTheGuardedManifest`
(`lint/internal/manifest/releasing_test.go`) holds the two halves together: every `MANIFEST.md` path
the document names must equal `manifest.CommittedRelPath` — **the single declaration**, which
`TestManifestUpToDate` now also reads — and that file must exist. **Zero paths extracted is a Fatal**,
never a pass, so a document that quietly stops naming the manifest cannot silently un-retire this.

Red-proved three ways: the document naming a different path → red; the document naming no manifest at
all → **Fatal** on the vacuity path; moving the single declaration → **both** tests red.

**Owner: none. There is nothing left to own.**

## Open

### DW-26 — `style.fontSize` has no range check at all, and it is the other operand of the one product that can overflow `geom.ScaleRound`

- **Deferred by:** Story 7.2 (2026-08-30), whose own contract forbids it the fix — D-7.2.4 rules a
  bound on `fontSize` out of that story's scope explicitly, as a second field that would have earned
  the story a `multiple-goals` warning.
- **Owner:** the next story that bounds a numeric format field's legal range, plus the epic
  plan-gate checklist as a second standing address. Named this way on purpose, following DW-25:
  D-000.73 rules that an owner which is an EVENT ("Epic 7 close") stops existing when the event
  passes, and this file's own DW-24 is the worked example of what an unowned item costs — it
  survived two stories undischarged.
- **Severity:** MEDIUM. The consequence is not a crash but a suite-wide blindfold (see below).
- **Status:** OPEN. **Recorded, deliberately not closed** (D-7.2.4).

**What is missing.** `style.fontSize` is decoded through `decodePointsRaw` →
`decodePoints` (`folio-go/internal/template/decimal.go:30`), which enforces exactness (no more than
three decimal places) and refuses int64 millipoint overflow — and **nothing else**. There is no
minimum, no maximum, and no positivity check. `"fontSize": 0` loads. So does a font size of
9,000,000,000 points.

**Why it is recorded HERE and not merely as a nit. A PANIC IS STILL REACHABLE FROM A TEMPLATE
THROUGH THIS FIELD, AND THAT IS WHY THIS ITEM IS OPEN.**

`geom.ScaleRound` **panics** on int64 product overflow (`folio-go/internal/geom/scale.go:68`), and a
Go panic aborts the package binary, so every other test in `folio-go` silently stops reporting — a
suite-wide blindfold rather than a crash. `verticalModel` (`folio-go/wrap.go`) performs **two**
multiplications, and only one of them is guarded:

| multiply | operands | guarded? |
|---|---|---|
| the model's own `scale` closure — `ScaleRound(units, fontSize, 1000)` | font units × **`style.fontSize`** | **NO** |
| the leading ratio — `ScaleRound(ruledAdvance, lineSpacing, 1000)` | ruled advance × `style.lineSpacing` | **yes**, by `int64MulWouldOverflow` immediately before the call |

Story 7.2 discharged the precondition it introduced, and red-proved it by neutering the guard and
observing the panic. It did **not** discharge the pre-existing one, because doing so means bounding
`fontSize` — a format-domain decision about a second field's legal values, which D-7.2.4 rules out
of that story's scope. **Measured:** `verticalModel([]string{"probe"}, notoSansMetrics,
geom.Length(1<<62), 1000)` panics with *"geom: ScaleRound: v\*num overflows int64"* — **with
`lineSpacing` absent entirely**, i.e. through nothing but an authored `fontSize`, on a document the
7.2 feature never touches.

So the honest statement of this item is: the ratio multiply is safe; the font-size multiply on the
same function is not; and a template can still reach a panic through an unbounded `fontSize`. The
shared site is named here so that whoever eventually bounds `fontSize` knows the ratio guard already
standing beside it, and does not remove it as redundant — and so that the panic route is not
mistaken for closed.

**Related, and equally unbounded:** an absurd `fontSize` also reaches `measureRuneRange`'s glyph
scaling and the canvas projection's `MaxCanvasMillipoints` bounds. Those refuse or clamp rather than
panic. **No route through `fontSize` is closed; one route through `lineSpacing` is.**

**Why not raise `lineSpacing`'s ceiling instead.** Deriving a load-time `lineSpacing` ceiling
honestly from the overflow case gives roughly **1023 thousandths** — it would forbid `lineSpacing`
above 1.0. That reductio is exactly why D-7.2.3's `1000000` ceiling is a stated sanity bound and
nothing more, and why the honest remedy is a bound on `fontSize`, not a tighter one on the ratio.

**Trigger:** any story that bounds a numeric format field's range, any story that changes
`decodePoints` or `decodePointsRaw`, and any story that touches `geom.ScaleRound`'s preconditions or
the leading model's guards in `folio-go/wrap.go`.

**How we'd know it was CLOSED.** `verticalModel`'s `scale` closure can no longer be driven to a
panic from an authored `fontSize` — either because the field is bounded at load, or because that
multiply is guarded the way the ratio multiply is. Until one of those is true, this entry stands and
the panic above is reproducible.

**This item is OPEN.**

---

### DW-24 — the SIX rounded alignment branches — `align: "center"` and `valign: "middle"`, in `text_alignment.go` and in all three table-cell paths — are declared by no fixture, so every rounding site in the feature has zero golden coverage — **CLOSED by Story 7.3, 2026-08-30**

> **STATUS: CLOSED.** Everything from here to "DW-24 IS CLOSED" below is the entry as it stood while
> it was open, kept verbatim because the closing note reconciles against it — the hand-list it
> contains had rotted twice, and deleting the wrong list would delete the evidence of that. The
> closure, the re-derived enumeration and the per-site falsifier results are at the end of the entry.

**Owner:** **Story 7.3**, *and* the **orchestrator's own gate checklist** as a second standing
address. Two addresses, on DW-21's precedent, because one of them is a person's checklist that
outlives any single story.

Story 7.1 **inspected this item and declined it, on the criterion rather than on the budget**
(D-7.1.4). DW-24's stated hazard is the unexercised **rounding** branch, and 7.1 touches neither the
rounding nor the population that reaches it: no corpus document declares `center` or `valign`, and
7.1 adds none. It leaves the gap exactly as it found it, so closing it there would discharge nothing
7.1 endangered. 7.1 did amend this entry's scope — see the enumeration below — because a deferral
whose scope is known wrong is worse than one merely open.

Story 7.3 is a real forcing function rather than a second guess: it extends the closed align set with
`justify`, must author an aligned fixture anyway, and its slack-remainder rule is itself new integer
rounding in the same neighbourhood.

**NOT "Epic 7 close", and that is deliberate** (D-7.1.4, D-000.73). DW-14's owner was "the Epic 2
boundary gate", which ran and closed without re-owning it; the item then survived a whole epic with
nobody holding it. An owner that is an *event* stops existing the moment the event passes.

**Trigger (any one fires it):** any story that touches the vertical model — `verticalMetrics`,
`FirstBaseline`, `Advance`, `textBlockHeight`, or `textValignOffset` (7.2 and 7.7 both do) — any
change to `folio-go/text_alignment.go` or to the table-cell alignment switches in
`folio-go/table_render.go`, and any change to `geom.ScaleRound` or its rounding rule.

**The trigger has now fired TWICE, and the item was declined both times. IT IS NOT DEFERRABLE A
THIRD TIME** (D-7.2.6).

*First firing — Story 7.1.* It changed what FEEDS `textBlockHeight` — an empty line now contributes
a full `Advance` — without touching the rounding or the population that reaches it. See the Owner
note above: the criterion is the unexercised rounding branch, and a story that leaves that gap
exactly as it found it discharges nothing by closing it.

*Second firing — Story 7.2, declined on a NARROWER ground than the one first offered.* The reasoning
matters more than the outcome here, because the reasoning first proposed was wrong. It was that 7.2
touches "a different call site, denominator and file". **It does not.** Story 7.2 genuinely DOES
change the input to the unexercised rounding site: `textBlockHeight` is built from `Advance`, and
`style.lineSpacing` scales `Advance`, so the slack that `ScaleRound(slack, 1, 2)` halves for
`valign: "middle"` is a number this story moves. The exposure is unchanged for one reason only —
**no fixture declares `valign` at all**, which is the very absence this entry exists to record.
That, and not "a different call site", is the ground on which 7.2's decline stands.

*Why there is no third time.* Story 7.3 is this item's owner (D-7.1.4). **7.3's plan gate treats it
as an ACCEPTANCE CRITERION, not a deferred item**, and a decline there is an escalation to the
engineering lead rather than another entry in this file. A trigger that has failed to fire twice
stops being a trigger.

**Raised at:** Story 15.1, from the population census that closed its attribution (D-15.1.1, D-R7.5),
and **widened in that story's own review** from a `valign`-only item to this one.

**What the corpus actually declares.** `grep -oh '"align"[^,}]*' fixtures/*/input.folio` yields
**16 `"left"` and 8 `"right"`. No `"center"`. And `grep -rn 'valign' fixtures/` returns nothing at
all.** So of `text_alignment.go`'s six branches, the corpus exercises two — `left` (the no-op) and
`right`.

**Why those two specific absences matter more than the others, and it is not symmetry.** `center` and
`middle` are the **only** branches that round: both return `geom.ScaleRound(slack, 1, 2)`
(`folio-go/text_alignment.go:56` and `:74`), and that is the only rounding in that file. `right`
returns `slack` unchanged and `left`/`top` return zero — neither can express a rounding defect. So the
one construct in this feature where a cross-target byte divergence could plausibly appear — a
half-to-even tie broken differently on a different target, which is what AD-2/AD-3 exist to prevent
and what the four-target matrix exists to catch — is declared by **no document the matrix renders**.
An exact-half slack in millipoints is not exotic; it is what a symmetric box gives you.

**The enumeration is SIX sites, not two — amended by Story 7.1 (D-7.1.4).** This entry shipped naming
only `text_alignment.go`'s pair, which is the rounding a centred *text element* reaches. A table's
own cells round in the same way, in different code, and a centred text element does not cover them:

| site | branch |
|---|---|
| `folio-go/text_alignment.go:56` | text element, `align: "center"` |
| `folio-go/text_alignment.go:74` | text element, `valign: "middle"` |
| `folio-go/table_render.go:687` | table **header** cell, `align: "center"` |
| `folio-go/table_render.go:698` | table **header** cell, `valign: "middle"` |
| `folio-go/table_render.go:1017` | table **body** cell, `align: "center"` |
| `folio-go/table_render.go:1193` | table **footer** cell, `align: "center"` |

All six were confirmed by grep at those exact lines at Story 7.1's baseline (`98cadf7`).

**The pairs are deliberately uneven, and that is a fact about the code rather than a gap in this
list.** A text element and a table HEADER cell each round on both axes; a table BODY cell rounds only
horizontally, because a body cell's vertical slack is distributed in whole line SLOTS rather than in
millipoints (a row's height is already an exact multiple of `vm.Advance`), so there is no remainder
to round; and a FOOTER cell has no vertical slack of its own at all. So four `align` sites and two
`valign` sites is the complete set, not four of an expected six pairs.

The same grep also surfaces `folio-go/render.go:482`/`:483`, which centre an *image* inside its box —
a different subject from this entry's text alignment, recorded here only so the next re-derivation is
not surprised by them.

**AT CLOSURE THE ENUMERATION MUST BE RE-DERIVED BY GREP AND RECORDED — NEVER READ OFF THIS
HAND-LIST.** The list above is being amended today precisely because it had already rotted once: the
closing fixture would have satisfied this entry's literal text while missing four of the six sites,
and the entry would then have been marked closed. Run
`grep -rn 'ScaleRound(.*1, 2)' folio-go --include='*.go' | grep -v _test`, record what it returns in
the closing note, and close against **that**, not against this table.

**Measured, in both directions, so this is a number rather than a worry.** Story 15.1 ran the
zero-it-out experiment on the wiring. Multiplying `textAlignOffset` by zero in `render.go` turned all
four statement goldens green against their PRE-`791ed00` bytes — the corpus detects the `right`
branch exactly. Multiplying `textValignOffset` by zero reddened
`TestAlignedTextElementsMoveInsideTheirDeclaredBox` and
`TestCanvasPaintMatchesTheShippingRunPathUnderAlignment` and **not one golden anywhere in the
repository**; every `expected.pdf` stayed byte-identical. `center` cannot even be reached by that
experiment, because no fixture selects it.

So the honest statement of the gap is narrower than "untested" and still worth an entry: the
behaviour suite covers all six branches, and the **golden corpus covers two**. That matters here
specifically because the golden corpus is the guard Epic 7's stories bind themselves to — 7.1, 7.2,
7.3, 7.5, 7.7, 8.3 and 8.4 each carry a "the corpus hashes identically" criterion (D-R7.1), and for
the rounded branches that criterion is currently unfalsifiable: `center` or `middle` can move by any
amount, or diverge between targets, and every one of those criteria still passes. It is also the
exact shape of the defect `791ed00` fixed — alignment parsed, validated and displayed while emitting
nothing — which no golden caught for the whole life of the feature.

**What closing it looks like.** One fixture under `fixtures/` whose template declares **both**
uncovered branches: a text element with `style.align: "center"` and a second with
`style.valign: "middle"` (and ideally `"bottom"`, which is unrounded but also undeclared), each with
a declared `width`/`height` leaving slack, and **at least one of them with an ODD slack in
millipoints** so the half-to-even tie is actually taken rather than avoided. Recorded as a golden,
registered in `folio-go/byte_neutrality_test.go`'s `goldenDigestRecord`, and — because the rounding
is the point — added to `matrixDocuments` so all four targets render it. It needs no second face and
no fifty pages; what it needs is slack, a declared box, a half-slack tie, and **one document reaching
every site the re-derived enumeration returns** — which means a centred text element AND a table with
a centred column (header, body and footer cells round in `table_render.go`, in different code from
`text_alignment.go`, and a centred text element does not cover them).

It should **not** be bolted onto a statement fixture: those four carry a human sign-off whose
re-attestation costs a person reading four documents (D-4.7.1), and widening them makes every future
alignment change cost that same re-read.

**Why Story 15.1 did not simply add it.** 15.1's ruling is that the alignment move is *intended*
(D-R7.6), so its remediation is a re-record, and its own halt is on a human sign-off. Adding a fifth
golden in the same commit would put an unsigned new artifact into a corpus whose sign-off is
mid-re-attestation, and it would enlarge the diff the owner has to read before signing. The gap is
recorded here, with an owner and a trigger, exactly as that story's acceptance allows.

**How we'd know it was still wrong.** A change to the vertical model, or to the rounding rule,
landing with the corpus green while `center` or `middle` visibly moves. The cheap check has two
halves and **both** must be run — running only the first is what let this item ship as a
`valign`-only note:

1. Multiply `textValignOffset` by zero in `render.go` and run the whole golden suite.
2. Change the `center` branch's `geom.ScaleRound(slack, 1, 2)` to `slack/2` (or to `slack`
   outright) and run the whole golden suite. Do this at **every** site the re-derived enumeration
   returns, not only at `text_alignment.go`'s pair.

While this item is open, the suite stays green under both. When it is closed, **both** must go red.

---

## DW-24 IS CLOSED — Story 7.3, 2026-08-30

Closed by `fixtures/alignment-rounding/`, registered at every surface a golden is registered at and
rendered on all four targets. Story 7.1 amended this entry's scope and re-addressed it; Story 7.2
re-addressed it on the corrected ground above and fixed the deadline; 7.3 owned it as an acceptance
criterion (D-7.1.4, D-7.2.6) and discharged it.

### The enumeration, RE-DERIVED BY GREP, not read off the hand-list

**The grep is recorded TWICE, at two named revisions, and the difference between them is stated
rather than left to be noticed.** This story's own edits to `folio-go/text_alignment.go` moved two of
the lines the grep returns, so a closing note that published only the baseline anchors would ship
stale line numbers — which is exactly the rot that has already made the hand-list above wrong twice.
Neither block below is a summary of the other; both are verbatim command output.

**(a) AT THE BASELINE, `20ccefa`** — before this story's own edits moved any line. This is the
revision the reconciliation against the hand-list is computed from, and that reconciliation stays
valid at it: it is a statement about how far the entry's hand-written anchors had drifted by the time
7.3 picked the item up, and it is not restated at the closing revision.

```
$ grep -rn 'ScaleRound(.*1, 2)' folio-go --include='*.go' | grep -v _test
folio-go/render.go:486:// (ScaleRound(box-drawn, 1, 2)) rather than "/2" — D-1.8.4: "(bw-dw)/2
folio-go/render.go:505:	offsetX := geom.ScaleRound(bw-drawW, 1, 2)
folio-go/render.go:506:	offsetY := geom.ScaleRound(bh-drawH, 1, 2)
folio-go/text_alignment.go:56:		return geom.ScaleRound(slack, 1, 2)
folio-go/text_alignment.go:81:		return geom.ScaleRound(slack, 1, 2)
folio-go/table_render.go:705:				textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
folio-go/table_render.go:716:				lineTopY = contentY + geom.ScaleRound(contentH-textBlockHeight, 1, 2)
folio-go/table_render.go:1038:							textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
folio-go/table_render.go:1214:							textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
```

**(b) AT THE CLOSING REVISION — the working tree as Story 7.3 delivers it (`20ccefa` plus this
story's own edits).** This is the re-derivation the spec's Manual check mandates, and **this is the
enumeration DW-24 is closed against.** The anchors published here are the ones a future reader should
open a file at:

```
$ grep -rn 'ScaleRound(.*1, 2)' folio-go --include='*.go' | grep -v _test
folio-go/render.go:486:// (ScaleRound(box-drawn, 1, 2)) rather than "/2" — D-1.8.4: "(bw-dw)/2
folio-go/render.go:505:	offsetX := geom.ScaleRound(bw-drawW, 1, 2)
folio-go/render.go:506:	offsetY := geom.ScaleRound(bh-drawH, 1, 2)
folio-go/text_alignment.go:85:		return geom.ScaleRound(slack, 1, 2)
folio-go/text_alignment.go:110:		return geom.ScaleRound(slack, 1, 2)
folio-go/table_render.go:705:				textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
folio-go/table_render.go:724:				lineTopY = contentY + geom.ScaleRound(contentH-textBlockHeight, 1, 2)
folio-go/table_render.go:1046:							textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
folio-go/table_render.go:1230:							textX = contentX + geom.ScaleRound(contentW-measured, 1, 2)
```

> **CORRECTED AT STORY CLOSE, 2026-08-30 — and this correction is itself the point of the entry.**
> The block above as first written published `table_render.go:716`, `:1038` and `:1214` — the
> *baseline* anchors — and asserted below that the `table_render.go` anchors were "unmoved". They had
> moved. One of this story's own review-pass patches corrected the three cell-align `default:` arm
> comments in `table_render.go`, inserting lines *above* three of the four sites, and the grep was
> never re-run afterwards. So DW-24's closing note shipped stale anchors **in the very commit that
> published them** — the entry's hand-list rotting a third time, inside its own closure. Re-derived
> at `3f99e7f` (`table_render.go` is byte-identical at `ff6d565`, `9898845` and `HEAD`, so these
> anchors hold across all three) and corrected in place. **The lesson is the entry's own: an anchor
> is only true at the revision it was measured at, and a line number captured before the last patch
> of a story is not a closing-revision anchor.** Publish anchors last, or name sites by function.

**THE SET OF SITES IS IDENTICAL AT THE TWO REVISIONS. No site appeared, none disappeared, and the
subject of this entry did not change.** Five line numbers moved, all for the same mundane
reason — this story inserted text *above* them. Two are in `text_alignment.go` (the justification
rule and the extended file doc comment sit above both arms); three are in `table_render.go` (the
corrected cell-align `default:` arm comments sit above the header-`valign`, body-`align` and
footer-`align` sites). **Nothing about any of the branches themselves changed.**

| site (named by what is stable — the function) | at baseline `20ccefa` | at the closing revision |
|---|---|---|
| `textAlignOffset`, the `align: center` arm | `text_alignment.go:56` | **`text_alignment.go:85`** |
| `textValignOffset`, the `valign: middle` arm | `text_alignment.go:81` | **`text_alignment.go:110`** |
| header cell `align` / `valign`, body cell `align`, footer cell `align` | `table_render.go:705`, `:716`, `:1038`, `:1214` | `:705` unmoved; **`:716`→`:724`, `:1038`→`:1046`, `:1214`→`:1230`** — moved by this story's own correction of the three cell-align `default:` arm comments |
| image centring (out of subject) | `render.go:505`, `:506` | unmoved — same two anchors |
| the line-slot split (not a `ScaleRound`, found separately) | `table_render.go:966` | **`table_render.go:974`** — moved by the same comment patch |

**HOW TO READ EVERY ANCHOR BELOW.** The hand-list reconciliation immediately following is stated at
the **baseline**, and its bare `text_alignment.go:56` / `:81` are baseline anchors — they are what the
hand-list was compared against, not where the code is now. Everywhere after it, the two
`text_alignment.go` sites are named by **function**, which is stable across both revisions, and the
`table_render.go` anchors are the same at both.

**RECONCILED AGAINST THE HAND-LIST ABOVE, AT THE BASELINE `20ccefa`: it had rotted a SECOND time.** The entry claims all six
anchors "were confirmed by grep at those exact lines" at `98cadf7`. Five of the six had drifted, and
a seventh site the entry never listed exists:

| hand-list, as the entry wrote it | actual at baseline `20ccefa` | branch | verdict |
|---|---|---|---|
| `text_alignment.go:56` | `text_alignment.go:56` | text element `align: center` | unchanged at the baseline — but **`:85` at the closing revision** |
| `text_alignment.go:74` | **`text_alignment.go:81`** | text element `valign: middle` | drifted — and **`:110` at the closing revision** |
| `table_render.go:687` | **`table_render.go:705`** | header cell `align: center` | drifted |
| `table_render.go:698` | **`table_render.go:716`** | header cell `valign: middle` | drifted |
| `table_render.go:1017` | **`table_render.go:1038`** | body cell `align: center` | drifted |
| `table_render.go:1193` | **`table_render.go:1214`** | footer cell `align: center` | drifted |
| *(absent)* | **`table_render.go:966`** `slack / 2` | table `valign: middle`, a body row's spare LINE SLOTS | **never listed** |

`table_render.go:966` at the baseline (**`:974` at the closing revision**) is an integer LINE-COUNT split, not a `geom.ScaleRound` and not a cross-target
float hazard — its own comment argues it out of that scope, and correctly. It is nevertheless a
`middle`-only branch with zero golden coverage, which is exactly the absence this entry exists to
record, so it is closed here with the same fixture rather than left to be rediscovered.

`render.go:505`/`:506` centre an IMAGE inside its box. They are **unconditional on every image
element**, so they are a different subject from this entry's declared alignment, exactly as `:338`
already noted — and they are recorded again below with a measured result, so the next re-derivation
is not surprised by them a third time. `render.go:486` is a comment, not a site.

### The falsifier, run PER SITE, at the closing revision

DW-24's own two-part cheap check, plus the truncation mutation at **every** site the closing-revision
grep — block (b) above — returned. The command in each case is
`go test -count=1 -run 'Golden|golden' .` in `folio-go`, and the baseline for it is **no golden
failures at all**.

**Anchors in this table are closing-revision anchors, corrected at story close (2026-08-30).** The
two `text_alignment.go` sites are named by function rather than by line precisely because those are
lines this story moved (`:56`→`:85`, `:81`→`:110`). The `table_render.go` anchors were *also* moved
by this story — by its own review-pass comment patch, after this table was first written — and are
cited here at their corrected closing-revision values (`:716`→`:724`, `:966`→`:973`, `:1038`→`:1046`,
`:1214`→`:1230`). The mutations were **re-run at story close against these anchors**, and the results
below are that re-run, not the original one.

| # | site | mutation | result |
|---|---|---|---|
| 1 | `text_alignment.go` `textAlignOffset`, `align: center` | `ScaleRound(slack,1,2)` → `slack/2` | **RED** — `TestAlignmentRoundingGoldenFixture` |
| 2 | `text_alignment.go` `textValignOffset`, `valign: middle` | `ScaleRound(slack,1,2)` → `slack/2` | **RED** — `TestAlignmentRoundingGoldenFixture` |
| 3 | `text_alignment.go` `textValignOffset` | `→ 0` (the entry's own half 1: "multiply `textValignOffset` by zero") | **RED** — `TestAlignmentRoundingGoldenFixture` |
| 4 | `table_render.go:705` header cell `align: center` | → truncation | **RED** — `TestAlignmentRoundingGoldenFixture` |
| 5 | `table_render.go:724` header cell `valign: middle` | → truncation | **RED** — `TestAlignmentRoundingGoldenFixture` |
| 6 | `table_render.go:974` body row `valign: middle` line slots | `slack/2` → `0` (top) | **RED** — `TestAlignmentRoundingGoldenFixture` |
| 7 | `table_render.go:1046` body cell `align: center` | → truncation | **RED** — `TestAlignmentRoundingGoldenFixture` |
| 8 | `table_render.go:1230` footer cell `align: center` | → truncation | **RED** — `TestAlignmentRoundingGoldenFixture` |
| — | `render.go:505` image centring (**out of subject**) | → truncation | GREEN — no golden moved. Recorded, not closed: this entry's subject is TEXT alignment, and the image site is unconditional on every image element rather than selected by a declared value. It is the one rounding the corpus still does not pin, and it is now a measured fact rather than an assumption. |

**While this item was open the suite stayed green under all of the above. It is now red under every
one of them.** That is the closure condition, met per site.

### Why an odd slack was not enough, and what the fixture actually declares

`geom.ScaleRound(slack, 1, 2)` and a truncating `slack / 2` agree on every **even** slack — so a
centred fixture with even slack would have satisfied this entry's literal text and detected nothing.
An odd slack takes the tie, but **half of all odd slacks round DOWN to even**, which truncation also
produces. The discriminating condition is `slack ≡ 3 (mod 4)`.

Every one of the fixture's seven slacks is `3 (mod 4)`, and
`TestAlignmentRoundingSlacksAreOdd` asserts it rather than leaving it to luck. Three boxes are
declared in thousandths of a point to make it so — `height: 40.001`, `headerHeight: 24.001` and the
centred column's `width: 60.003`:

| site | slack | halves to | truncation gives |
|---|---|---|---|
| `e1` `align: center` | 158,783 | 79,392 | 79,391 |
| `e2` `valign: middle` | 25,019 | 12,510 | 12,509 |
| header `Qty` centred | 41,831 | 20,916 | 20,915 |
| header `valign: middle` | 9,019 | 4,510 | 4,509 |
| body `3`/`7` centred | 53,711 | 26,856 | 26,855 |
| body `12` centred | 47,419 | 23,710 | 23,709 |
| footer count `3` centred | 53,711 | 26,856 | 26,855 |

The line-slot split is not a millipoint round: row 1's clause wraps to four lines while its qty cell
is one, so `middle` takes slot `3 / 2 = 1` — neither the first nor the last, which is what makes
mutation 6 above red.

**THE GENERAL RULE, stated because it is not specific to halving and will be needed again.** A
red-proof against a *rounding mode* must be built on a value at which the two modes **actually
disagree** — which is a strictly stronger condition than "a value the mode does not divide exactly".
The two are easy to confuse, and this entry confused them for three stories.

Worked here: `geom.ScaleRound(slack, 1, 2)` is round-half-to-even. Against a truncating `slack / 2`
it agrees on every even slack, so "make the slack odd" looks like the discriminating choice — but it
is not, because round-half-to-even sends `slack ≡ 1 (mod 4)` to the *same* answer truncation gives
(`4m+1 → 2m`, since 2m is the even side of the tie). Half of all odd slacks are therefore a coin
flip that lands on truncation's answer. Only `slack ≡ 3 (mod 4)` discriminates: `4m+3 → 2m+2` under
half-to-even, `2m+1` under truncation. "Odd" is the *exactness* condition; `≡ 3 (mod 4)` is the
*disagreement* condition, and only the second one falsifies anything.

Generalise before reusing this: for any two rounding modes, enumerate the residue class where their
outputs differ and pick the fixture value from **that** class — never from the wider class of values
the operation merely fails to divide exactly. A fixture chosen on exactness alone can be green under
the mutation it was built to catch, and will report that as success.

### Registration

`fixtures/alignment-rounding/` (`README.md`, `input.folio`, `expected.json`, `expected.pdf`), plus
`folio-go/alignment_rounding_template.go`, `folio-go/alignment_rounding_fixture_test.go`,
`goldenDigestRecord`, `matrixDocuments` (with its own per-leg feature guard
`requireAlignmentRoundingRounds`), the missing-glyph corpus table and `beyondBaselineAcceptance`,
`declaredEpic2GateObligations`, `render_test.go`'s subprocess selector, and
`.github/workflows/matrix.yml`'s `docs=` list plus an upload path for each of the four targets under
`if-no-files-found: error`. The four legs were RUN in-story: `TestTargetRenderHash` once per
`FOLIO_MATRIX_TARGET` and `TestCrossTargetByteIdentity`; all four agree on
`986400a1c8bb1ff84d868bb8df70479c5e7e7a2ad5e867634efb810a47327087` (61,346 bytes).

It was **not** bolted onto a statement fixture (D-4.7.1), and no recorded digest moved.

**THIS ITEM IS CLOSED.**

---

### DW-22 — `ImagePaint` fetches paintable bytes once per ELEMENT with no cache, not once per distinct asset

**Owner:** **the second Epic 5 boundary gate** (already owed by Story 5.13's reopening, D-5.13.6) —
a legitimate, already-scheduled owner for this deferral, not a new one invented for it.

**Raised at:** Story 5.13's finisher pass, Finding 18 of that review.

**What.** D-5.13.2 accepted, as the cost of keeping the canvas snapshot lossy (AD-17), "an extra
request per **distinct** asset and its cache/lifetime handling in the browser." What shipped
(`folio-designer/src/App.tsx`'s `ImagePaint`) is a request per **element instance**, with no cache
at all: N image components sharing one asset key issue N separate `'asset'` worker round-trips,
each carrying the full decoded bytes back through the single FIFO-correlated worker, and the same
N requests fire again on every `generation` change (Open, Start blank, undo/redo).

**Why deferred rather than fixed now.** On a document with a header logo repeated across a few
bands this is a small, unmeasured cost; on a document with many placed images it serializes the
worker behind repeated multi-megabyte transfers, but no story to date has produced such a document,
so there is no measured regression to fix against — only the accepted-but-unbuilt design point
D-5.13.2 named. Building a keyed cache (one in-flight request and one object URL per distinct
asset key, revoked when the last referent unmounts) is a real, non-trivial addition to `ImagePaint`
that the mandatory Blockers and Majors already fixed in this story's finisher pass did not touch,
and scope discipline (the finisher's "minimal change that resolves the concern" heuristic) argues
against expanding this commit to build it speculatively.

**The trigger.** Either (a) a story places or measures a document with enough image elements/shared
assets that the N-requests-no-cache shape becomes an observed cost, or (b) the second Epic 5
boundary gate's own measurement pass surfaces it, whichever comes first.

**Retire when:** `ImagePaint` (or its replacement) keys the per-key bytes fetch by asset key at the
canvas level — one in-flight request and one object URL per distinct key, revoked when the last
referent unmounts — or the accepted cost is re-confirmed explicitly with a measurement showing it
does not matter in practice.

### DW-21 — Three heavy tests exist but run ONLY under `FOLIO_HEAVY=1`; the Epic 4 boundary gate must set it

**Owner:** **the Epic 4 boundary gate run**, and — because a gate has failed as an owner before
([[DW-14]]'s owner was *"the Epic 2 boundary gate"*, which ran and closed without re-owning it —
[[D-000.73]]) — **also the orchestrator's own gate checklist.** Two addresses deliberately, for a
one-line obligation whose whole failure mode is being forgotten.

**Raised at:** Story 4.4's finisher pass, closing that review's Blocker 1.

**What.** Story 4.4 shipped two genuine heavy/integration tests, gated on an ordinary env var so the
routine per-story gate never pays their cost (D-000.4's per-epic cadence):

```
env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -count=1 -v   -run 'TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable|TestTwoTablesWithPageCountFooterRenderConsistently|TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability' ./...
```

**Verified independently by the orchestrator before this entry was written**, both directions: unset →
both report `--- SKIP`, never a silent pass; `FOLIO_HEAVY=1` → both `--- PASS`. So the gate genuinely
has something to turn on, which is exactly what the story's first attempt did not have.

**Why an env var and not a build tag.** A new `//go:build matrix` file would itself register as an
**unauthorised Epic 2 gate obligation** — `TestEpic2GateObligationsMatchTheDeclaredSet` scans for the
matrix build constraint specifically and correctly refused one in Story 4.3. An env-gated ordinary
test stays outside that obligation set by construction.

**Why this entry exists at all rather than living only in the story.** Story 4.4's review found the
first attempt at these tests was **two empty bodies under unconditional `t.Skip`s** — *"written, not
run"* had shipped neither, and the skip count was misreported as 1 when it was 3. The lesson is that
a deferred-execution obligation recorded **only** in the story that deferred it is invisible at the
moment it comes due. **If the Epic 4 gate closes without this command having been run, that is the
same defect a second time.**

**Amended 2026-08-27 ([[D-000.81]]) — part of the concealment was a ruling's, not a relay error's,
and the remedy is a reporting rule rather than a checklist item.** [[D-000.74]] quarantined the
known-red corpus test by running the green CI job with `go test -skip "$KNOWN_RED_TEST"`. `-skip`
**excludes** a test rather than printing it as skipped — so that ruling silently **re-based the
denominator of the skip figure**, and nobody re-derived the baseline against its new meaning. Story
4.4's two empty-bodied skips then landed inside the slack that created, and the arithmetic reconciled
cleanly. Tidy reconciliation is what should have raised suspicion.

**Standing remedy, now in force for every gate report: report skips BY NAME, never by count.**
`go test -v` emits `--- SKIP: TestName` lines; the gate reports **the set**. A new name appearing in a
set is visible; a new name appearing inside an integer is not. The red-by-design test is **named** too,
never folded into arithmetic. A count is a lossy set.

**Retire when:** the Epic 4 boundary gate records the command, its output, and the pass/fail — or an
epic gate adopts a standing "run every env-gated heavy suite" step, which would make this entry
unnecessary rather than merely discharged.

**Amended at Story 4.5:** a third heavy test joins the set, same pattern (real body, `FOLIO_HEAVY=1`,
never a build tag) — `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`
(`folio-go/table_footer_test.go`): a 500-row footer table through the public `Render()`, confirming
the footer's sum appears exactly once, on the last page only, byte-stable across two renders, bounded
time. Confirmed both directions in this story's own run (unset → `--- SKIP`; `FOLIO_HEAVY=1` →
`--- PASS`; see the story's Delivery Log for the exact output). The recorded command above and
`table_header_repeat_test.go`'s own doc comment are both updated in the same change. **Not run as
part of this story's routine gate** — stated explicitly per D-000.4.

**AMENDED at Epic 4 boundary (2026-08-28, DW-21 discharge):** all three heavy tests named above
were run with `CGO_ENABLED=0`, `GOWORK=off`, and `FOLIO_HEAVY=1` via the exact command recorded in
the Epic 4 boundary artifact. Each passed. No test emitted `--- SKIP`; the selected tests were not
skipped, and nonmatching packages reported only `no tests to run` because the command's `-run`
regular expression selects these three names.


### DW-2 — The licence check's JS half: `folio-designer/`'s lockfile
- **Deferred by:** Story 1.3 (ruling D-1.3.4)
- **Owner:** **Story 5.1** — the story that creates `folio-designer/`
- **Anti-rot mechanism:** Story 1.3 ships an assertion that the lockfile is **absent**. The day Epic 5
  creates it, that assertion goes **red** and the build stays red until the JS half is wired.

AD-26 requires the licence check to cover the whole Go module graph **and** `folio-designer/`'s
lockfile at any depth. The Go half ships complete at 1.3; the JS half has nothing to check yet.
A conditional "check it if present" was rejected because it starts silently passing the moment the
directory arrives — the guard would report success exactly when it stopped covering anything.

Same treatment, same story, same mechanism: the OFL 1.1 text for shipped faces (fonts arrive at
**Story 2.2**) and the Apache-2.0 NOTICE for `pdfjs-dist` (**Epic 5**) are asserted absent now, so
landing either breaks the build until the manifest covers it.

**Confidence on the owner:** medium. If Epic 5's shape shifts, the owner moves with whichever story
first creates the lockfile.

**Fonts half retired (Story 2.2, AC5).** Story 2.2 shipped the three production faces at
`folio-go/fonts/` (Noto Sans, Noto Sans Thai, Noto Sans SC — AC1, AC9), each with its OFL-1.1
`LICENSE-OFL.txt` and `NOTICE.md`, so the `absence-fonts-dir` tripwire's job is done: it fired the
day the directory landed, exactly as designed. It has been **removed** from
`lint/internal/rules/absences.go`'s `absenceChecks` and replaced by a fail-closed guard with the
opposite polarity — `ScanFontsAssets` (`lint/internal/rules/fontsassets.go`, rule ids
`fonts-asset-unaccounted` / `fonts-asset-missing`), which now REQUIRES `folio-go/fonts/` to exist
and to hold only recognised shapes (a font file, a `LICENSE*`/`NOTICE*` pair, or a `.go` source
file), red-proved both ways (a stray file at the real location, and — via
`lint/internal/manifest.ResolveAssets`, AC25 — a missing `NOTICE*` file). **The JS/lockfile half of
this entry is unaffected and remains open, owned by Story 5.1** — only the fonts half is retired
here; DW-2 as a whole is not closed.

**Correction (Finding 8, this story's QA review, Major).** The three absence checks originally keyed
on exact guessed filenames (`folio-designer/package-lock.json`, `folio-go/fonts/OFL.txt`,
`folio-designer/third-party-notices/pdfjs-dist/NOTICE`) — a `pnpm-lock.yaml` or an `OFL-1.1.txt`
would have satisfied none of them and gone build-green, D-1.3.4's own rejected hazard arriving
through a side door. The finisher re-keyed both checks on the **directory** instead
(`folio-designer/` absent; `folio-go/fonts/` absent) — Story 5.1 cannot create the project without
creating `folio-designer/` first, and Story 2.2 cannot ship a face without creating
`folio-go/fonts/` first, so neither check depends on guessing a filename. The directory-level
`folio-designer/` check is strictly broader than the two checks it replaces (package-lock.json's
and the pdfjs-dist NOTICE's): both artifacts live inside `folio-designer/`, so any artifact landing
there — regardless of name — now trips the same finding. This still fully implements this
paragraph's "same treatment, same story, same mechanism" for the OFL text and the pdfjs-dist
NOTICE; it does not narrow anything DW-2 promises.

**DISCHARGED — Story 5.1 (2026-08-28, finisher corrected).** `folio-designer/package-lock.json` is covered by the positive, fail-closed lockfile-only `licence.ResolveNPMGraph` / `rules.ScanNPMGraph` AD-26 path and generated `lint/MANIFEST.md` rows labelled `folio-designer`; neither reads `node_modules`. Missing, unknown, optional, and prohibited transitive records red the guard. A zero-tenant absence scanner cannot return because `ScanAbsences`, `AbsencesStats`, `absenceChecks`, its tests, and fixture-only rule were removed and the structural guard scans the live lint tree. `pdfjs-dist` remains absent, so no fabricated notice is committed; when it appears its Apache-2.0 lock record and `folio-designer/third-party-notices/pdfjs-dist/NOTICE` are both required by `ScanPDFJSNotice`.

### DW-4 — Nobody owns cutting `folio-go/v0.1.0` — **owner decision recorded at Epic 4 planning; release checkpoint remains open**
- **Raised by:** the engineering lead during Story 1.3 rulings
- **Owner:** **the project owner**, at Epic 4 planning

No story in Epics 1–6 owns the tagging event itself. "Who cuts `folio-go/v0.1.0`, and what ships with
it" is a release-process decision with licence-compliance consequences (DW-3 depends on it), and
D-1.1.c fixes the public API at that tag — so it is also the moment the medium-confidence
argument-packaging question becomes irreversible. Not due now; **must not evaporate.** If it is still
unowned when Epic 4 is planned, it goes to the owner rather than being absorbed into a story.

**Ledger entry (Story 3.7, D-3.7.2):** `documentDate` is now a reserved top-level `params` key
(RFC 3339 string, setting both `/CreationDate` and `/ModDate`) — public contract, frozen at
`folio-go/v0.1.0` alongside the API signatures this entry already tracks. The `params` namespace now
has one reserved name in it; a future story adding a second one should append its own line here
rather than letting this ledger go stale.

**AMENDED again 2026-08-30 — the release timing decision has been overtaken by events, and Story 15.3
re-makes it rather than assuming it.** The owner's recorded trigger was "cut after Epic 6"; Epics 7–15
now all postdate that decision without amending it, and Epics 9 and 10 have already added public
surface and format fields since. Story 15.3 therefore re-affirms or amends the trigger explicitly,
naming which epics are inside v0.1.0, and re-measures the API surface census against what Epics 5–14
actually left rather than the 40 items measured at the Epic 3 boundary. The engineering-lead
checkpoint this entry keeps open is discharged by that story, not before it.

**AMENDED at Epic 4 boundary (2026-08-28, D-000.78 audit):** the project owner's decision is
recorded in `folio-mvp-decision-log.md` as **cut `folio-go/v0.1.0` after Epic 6**, so the original
"owner decision when Epic 4 is planned" trigger has fired and the release timing is no longer
unowned. The entry remains open only for the intermediate engineering-lead checkpoint explicitly
owed by D-000.78 before the post-Epic-6 tag; this boundary does not discharge that future checkpoint.

### DW-5 — Derivation validation of `columns[].footerOf` from `bind` — **RETIRED at Story 3.2**
- **Deferred by:** Story 1.4 (ruling D-1.4.2, AC43/AC44)
- **Owner:** **Story 3.2**, backstop **Story 3.7** (`folio.Validate` must include it)
- **Retired (D-000.59, discharge by replacement, not deletion alone):** the derivation landed at
  `internal/expr.DeriveFooterOf` (`folio-go/internal/expr/footer.go`), invoked from
  `folio.ParseTemplate` (`folio-go/folio_expr_validate.go`) — forced up to the module root by F2:
  `internal/template` (stage rank 2) can never import `internal/expr` (rank 3). Both derivable D-1.4.1
  shapes resolve to the derived `footerOf` (and, for shape 2, `footerFormat`); any other `bind` shape
  on a column requesting a `sum`/`avg` footer with `footerOf` omitted is now a load error naming the
  column id. The derived value is resolved ALONGSIDE the document, never written back into it (R2 —
  writing it back would break D-1.4.3's P3 fixed point for every document that legitimately omits
  `footerOf`). The `absence-expr-package` lint tripwire that stood in for this obligation is DELETED
  in the same commit, replaced by the positive assertions above plus D-3.2.1's own guards
  (`folio-go/internal/expr_arch_test.go`). **The aggregate itself is still not computed — DW-7 is the
  entry that tracks that, and it is untouched.**

D-1.1's derivation rule (a bare row-scoped path, or a single `formatNumber(...)` call over one) needed
the parsed expression tree to check `bind`'s shape — machinery Story 1.4 deliberately did not build
(no `internal/expr` package existed yet). Before Story 3.2, a `footer` with no `footerOf` simply
loaded (AC44's known, fixture-pinned gap) rather than being derived or rejected.

### DW-6 — The two footer diagnostic codes: `TABLE_FOOTER_SOURCE_UNRESOLVED` / `TABLE_FOOTER_SOURCE_FORBIDDEN` — **RETIRED by Story 3.6 (R6, R8, AC2)**
- **Retired at:** Story 3.6, by replacement, in the same commit (R6, D-000.59): `absence-diag-package`
  was deleted from `absenceChecks` (`lint/internal/rules/absences.go`), and `internal/diag`'s own
  `TestRegistryIsAdditiveOnly` (`folio-go/internal/diag/diag_test.go`) lands the positive assertion
  that the registry as constructed contains both codes, each pinned to its exact string:
  `TABLE_FOOTER_SOURCE_UNRESOLVED` (attached at `folio-go/folio_expr_validate.go`'s
  `validateTableColumns`, the `!derivable` branch) and `TABLE_FOOTER_SOURCE_FORBIDDEN` (attached at
  `folio-go/internal/template/parse_bands.go`'s two sites — `newLoadErrorCoded`, one code, two sites,
  because the code names the condition, not the line). Both travel wrapped in `*folio.RenderError`
  (D-3.6.3), never merely as a bare error. `TestAbsencesChecksIncludeTheRemainingEntry`
  (`absences_test.go`, renamed again at Story 3.7 when the list shrank to one entry) pins the
  remaining row.
- **Deferred by:** Story 1.4 (ruling D-1.4.2)
- **Owner:** **whichever story first creates `folio-go/internal/diag`** — expected to be Story 3.6,
  but the obligation attaches to the condition, not the story number (D-2.8.4)
- **Anti-rot mechanism (corrected, D-2.8.4):** this is **not a test**. It is the live production lint
  rule `absence-diag-package` (`lint/internal/rules/absences.go:100-104`), registered in
  `absenceChecks`, executed by `TestAbsencesProductionScan`. The day any story creates
  `folio-go/internal/diag`, that scan goes red naming `absence-diag-package` — and the **real hazard
  is the inverse of what this entry used to say**: the cheapest fix to the red is to **delete the
  rule**, which would retire this forcing function silently and permanently. Whichever story creates
  the package must, in the **same commit**: (1) delete `absence-diag-package` AND (2) land the
  **positive** assertion that replaces it — that the code registry contains
  `TABLE_FOOTER_SOURCE_UNRESOLVED` and `TABLE_FOOTER_SOURCE_FORBIDDEN`. Replace, never merely delete.
  Also update the rule's `desc` (it currently names "Story 3.6" by name — D-000.37) in that same
  commit if it turns out not to be 3.6.

Story 1.4's load failures are plain Go errors (D-1.4.2: *"1.4 must not mint them early"*).
`internal/diag` does not exist yet; AD-14 lands with Story 1.6 and the codes with Story 3.6.

### DW-7 — Footer evaluation sameness with `{{sum(...)}}` / `{{avg(...)}}` / `{{count(...)}}`
- **Deferred by:** Story 1.4 (ruling D-1.4.2)
- **Owner:** **Story 4.5**, by name
- **Anti-rot mechanism:** **none possible.** D-1.4.2's own honesty: *"no package tripwire exists, so
  `deferred-work.md` is the only trigger — flagged as the weakest of the three."* Nothing renders a
  table (and so nothing evaluates an aggregate) until Story 4.5, so there is no absent-package
  structural seam to key a red-proof on the way DW-5 and DW-6 have. This entry itself is the only
  thing keeping the requirement visible until then.

`columns[].footer`'s `sum`/`count`/`avg` must eventually use the *same* aggregate evaluation as the
`{{sum(...)}}` family of expression functions — a single implementation, not two that can drift.
Story 1.4 builds neither; nothing renders a table until Story 4.5.

**APPENDED at Story 3.3** (append-only, per D-000.29/D-3.1.1's own discipline — the paragraph above
is never edited in place):

- **What landed:** Story 3.3 builds the one aggregate evaluation `columns[].footer` must eventually
  reuse — `internal/expr`'s `evalSum`/`evalCount`/`evalAvg` (`aggregate.go`), routing through
  `SumDecimals`/`AvgDecimals` via a **positive routing assertion**, not merely the reducer inventory's
  declaration-shape check (`TestSumRoutesThroughSumDecimals`/`TestAvgRoutesThroughAvgDecimals`,
  `internal/expr/routing_arch_test.go`), with a **captured red-proof**
  (`TestSumRoutingRedProofInlineAccumulator`) showing an inline `big.Int` accumulator — which passed
  every guard that existed before this story (D-3.1a.4) — reddens this new assertion. This discharges
  D-3.1a.4's own follow-up, owed to Story 3.3 by name.
- **What remains:** the **footer half** is untouched — `columns[].footer` does not exist yet
  (Story 4.5's own field), and nothing calls `evalSum`/`evalCount`/`evalAvg` except the
  `{{sum(...)}}` family of expression functions. Story 4.5 still owns wiring the footer to this same
  evaluation, and still owns proving it did.
- **Correction:** *"Anti-rot mechanism: none possible"* is now **"none possible for the footer
  half."** The `{{...}}` half gained a real anti-rot mechanism this story (the routing assertion
  above); the footer half gained none, because there is still nothing to key one on until Story 4.5
  gives `columns[].footer` a shape. Ownership of the footer half is unchanged: **Story 4.5**, by name.

**APPENDED at Story 4.5 (RETIRED by replacement — R6/D-000.59's discipline: replace, never merely
delete):**

- **What landed:** the footer half is wired to the SAME evaluation as the `{{...}}` half — not by a
  second evaluator, but STRUCTURALLY: `table_render.go`'s `footerCellExprText` synthesises the exact
  `"{{sum(<footerOf>)}}"` / `"{{count(<collection>)}}"` / `"{{avg(<footerOf>)}}"` text (wrapped in
  `formatNumber` per AC2) an author would write, and hands it to `bind.Resolve` — the one
  display-text function every other cell already uses, which parses it with `expr.Parse` and
  evaluates it with the SAME `expr.Eval` → `evalSum`/`evalCount`/`evalAvg` → `SumDecimals`/
  `AvgDecimals` path the `{{...}}` half already reaches. This makes "routes through the same
  evaluation" a structural fact (the call graph is identical) rather than an argued one.
- **Positive routing assertion, the footer half's own, in the shape Story 3.3 shipped for the
  `{{...}}` half:** `TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression`
  (`folio-go/table_footer_test.go`) asserts BEHAVIOURAL equality — the footer's rendered sum/count/avg
  strings are byte-identical to an independently-evaluated, author-written `{{sum(...)}}`/
  `{{count(...)}}`/`{{avg(...)}}` expression over the SAME data, through `bind.Resolve` directly —
  which is AC3's own "equal, exactly" requirement, extended to be this AC's routing witness too.
- **Captured red-proof (the mutation this AC's own text names):** replacing the footer's route into
  the shared aggregate evaluation with a hand-rolled, ignore-the-exponent inline accumulator (adding
  Decimal coefficients directly, without `SumDecimals`'s alignment step) reddened
  `TestDecimalReducerInventoryIsExactlySumAndAvg` (`folio-go/internal/reducer_inventory_test.go`) the
  moment the accumulator was given the reducer inventory's own tripwire shape (`func(...[]expr.Decimal)
  (expr.Decimal, error)`) — a THIRD reducer where the inventory expects exactly two — proving R3/
  D-3.1a.3's "one aggregate implementation" guard reaches the footer half too. See the story's
  Delivery Log for the exact mutation, command and reddened test name.
- **Correction:** DW-7 is now discharged in full. Both halves — `{{...}}` (Story 3.3) and
  `columns[].footer` (Story 4.5) — share one evaluation, each with its own positive routing assertion
  and its own captured red-proof. No further owner is due.

**APPENDED at Story 4.5, FINISHER PASS (append-only, D-000.29/D-3.1.1 — the block above is not
edited in place; it is CORRECTED here, because part of what it claimed was not true):**

- **The claim that was wrong.** The block above records
  `TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression` as "the footer half's own
  positive routing assertion" and cites a captured red-proof that reddened
  `TestDecimalReducerInventoryIsExactlySumAndAvg`. That test is **Story 3.1a's module-wide reducer
  inventory** — it fires for a third reducer *anywhere* in the module and has no relation to the
  footer. The story's code review re-ran the mutation the AC actually names (the footer's `sum` route
  replaced by a literal `"13,500.00"` that never touches `SumDecimals`) and the **whole suite stayed
  green**: the behavioural equality was a single-dataset value comparison, so any rival that agreed on
  that one fixture passed. **DW-7 was, for one commit, marked discharged against another story's
  instrument** — the [[D-4.2.4]] defect, a deferral discharged by something that was not measuring it.
- **What discharges it now, and it is two things, not one:**
  1. **Structural, footer-side:** `TestFooterCellExpressionNamesTheSharedAggregateFunctions`
     (`folio-go/table_footer_test.go`) parses the expression `footerCellExprText` actually hands to
     `bind.Resolve` and asserts, **by AST**, that it is a `formatNumber()` call whose operand is a
     call to the shared `sum`/`count`/`avg` over the column's own resolved source path. This is the
     footer-side analogue of `internal/expr/routing_arch_test.go`'s
     `TestSumRoutesThroughSumDecimals` / `TestAvgRoutesThroughAvgDecimals`, which cover the `{{...}}`
     half and only that half.
  2. **Behavioural, over MORE THAN ONE DATASET:**
     `TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression` now renders at two row
     counts and asserts up front that their aggregates differ, so a constant cannot agree with both.
     A footer value is a function of the data, and a function is not witnessed by one point.
- **Captured red-proof, the review's own mutation, re-run against these two:** inserting
  `if col.Footer.Value == "sum" && !collectionEmpty { return "13,500.00", nil }` at the head of
  `footerCellExprText` reddens **`TestFooterCellExpressionNamesTheSharedAggregateFunctions`** and
  **`TestFooterRoutesThroughTheSameAggregateEvaluationAsAnOrdinaryExpression`** (whole suite,
  `-count=1`, not behind `-run`; 4 top-level failures in all). Recorded by NAME rather than by count,
  per D-000.81.
- **Correction:** DW-7 **is** discharged, but on the evidence in this block, not on the evidence in
  the block above it. No further owner is due.

### DW-8 — `Decimal` moves to `internal/expr` (or a leaf) and 1.6's path matcher is deleted — **RETIRED at Story 3.2**
- **Deferred by:** Story 1.6 (rulings D-1.6.1, D-1.6.5)
- **Owner:** **Story 3.2** — the expression-language story
- **Forcing function:** DW-5's existing `internal/expr`-absent tripwire reddened the moment that
  package was created, which is what made someone re-read this entry.
- **Retired.** Both obligations discharged in the same story:
  1. `Decimal`, `NewDecimal`, `SumDecimals`, `AvgDecimals` and their unexported bounds
     (`maxDecimalCoefficientDigits`, `maxDecimalExponentMagnitude`, `avgExtraScale`) MOVED from
     `internal/bind` to `internal/expr` — never duplicated. `TestExactlyOneDecimalDeclarationInTheModule`
     (`folio-go/internal/expr_arch_test.go`) asserts, by AST set-equality, that exactly one `Decimal`
     type declaration exists in the module and it is in `internal/expr`. D-3.1a.3's reducer-inventory
     tripwire (`folio-go/internal/reducer_inventory_arch_test.go`) — relational by design — followed
     the move with ZERO edits, exactly as that ruling required.
  2. `parseBindingPath` and `isValidIdent` (`internal/bind/text.go`) are DELETED — the expression
     parser in `internal/expr` (AD-9: hand-written recursive descent, no generator, no third-party
     dependency) is now the module's one grammar for `{{ }}` content.
     `TestParseBindingPathAndIsValidIdentAreAbsent` (`folio-go/internal/expr_arch_test.go`) is an
     extinction guard confirming both names are absent from the module.

**Two obligations, one owner.**

1. **The `Decimal` type moves; it is never duplicated.** AD-23 **Binds** both `internal/bind` and
   `internal/expr`, and Epic 3's `sum`/`avg`/comparison need exact decimal arithmetic. But
   `internal/bind` imports `internal/expr`, so `expr → bind` would be an **import cycle** — a hard
   compile error. **The dangerous resolution is not the cycle (Go stops that) but duplicating the
   type to break it.** Pre-committed: `internal/expr` may never import `internal/bind`; when 3.2
   needs `Decimal`, the type **moves**.
2. **Story 3.2's parser replaces 1.6's path matcher — deleted, not kept alongside.** 1.6 accepts only
   a bare dotted path and rejects everything expression-shaped precisely so that two parsers never
   coexist. If 3.2 leaves the matcher in place, the wrong one eventually wins.

**How we'd know it was forgotten.** A second `Decimal` type anywhere in the module, or a dotted-path
matcher surviving alongside the expression parser after 3.2.

### DW-9 — Re-test AC4's "nothing ceremonial" claim once a shipped font set exists — **RETIRED at Story 2.2**
- **Deferred by:** Story 1.7 (ruling D-1.7.1)
- **Owner:** **Story 2.2** — "The shipped font set and its fallback chain"
- **Retired:** the re-test ran. `folio-go/example_test.go` carries a compiled, EXECUTED
  `func Example()` (`go test -run Example` passes with its `// Output:` comment), rewritten against
  the shipped set: `folio.LoadTemplate`, then `fonts.Shipped()` — the `FontSet` obtained in ONE
  expression taking NO arguments — then `folio.Render`. `folio-go/README.md`'s "Your first PDF"
  section now shows this Example verbatim — and "verbatim" is mechanically checked, not asserted:
  `TestREADMEExampleBlockMatchesSource` byte-compares the fenced block against `example_test.go`,
  after the two were found to have drifted (the README kept the dead `err == nil &&` conjunct that
  the .go file had genuinely dropped). The "this step is ceremony Story 2.2 REMOVES" comment is
  removed (it is no longer true), and a new subsection explains why `folio`/`folio/fonts` stay two
  separate imports: a root re-export would embed **~11.3 MB raw** into every caller's binary,
  including the wasm build. *(That figure was `~9 MB` here and in the README until Story 2.2's
  finisher — `go:embed` stores RAW bytes, so the binary-size argument takes the uncompressed
  measurement; ~9 MB is NFR7's COMPRESSED download budget, against which the shipped faces measure
  5.07 MB at `brotli -q 11`.) **Verdict: the claim held — the re-test did not surface a
  `DECISION NEEDED`.**

Story 1.7's AC4 requires that producing a first PDF takes *"a load call and a render call, and nothing
ceremonial"*, and D-1.1.c named the README as the test of whether five positional arguments read as
ceremony. **That test cannot be run fairly at 1.7.** Verified: there is no README in the repo (1.7
writes the first), and `folio-go/fonts/` does not exist — the shipped faces arrive at 2.2. So 1.7's
first-PDF example must show a caller assembling a `FontSet` from their own bytes, **which is ceremony
that Story 2.2 removes, not ceremony the signature causes.**

**The packaging decision itself is closed** (D-1.7.1, on AD-8: an options struct would make `FontSet`
omittable at compile time, turning an AD-8 violation from a compile error into a runtime one). What is
deferred is only the **ceremony judgement** on the README example.

**At 2.2:** rewrite the first-PDF example using the shipped default font set and re-read it. If it
still reads as ceremony, that is a `DECISION NEEDED` — and it must be raised **before the
`folio-go/v0.1.0` tag** (Epic 4 close, DW-3/DW-4), after which the signature is fixed.

**How we'd know it was forgotten.** A `v0.1.0` tag cut without anyone re-reading the first-PDF example
against a shipped font set.

### DW-10 — `/CreationDate` and `/ModDate` wiring — **ALREADY OWNED, not newly deferred**
- **Raised by:** Story 1.7's creator (the clause became reachable when `params` first existed)
- **Owner:** **Story 3.7** — "Validate a template and render it from the command line"
- **Status:** already scheduled with acceptance criteria written; this entry exists only so the state
  *"`params` exists and nothing reads it for `/Info`"* is not silently forgotten.

AD-7 (`ARCHITECTURE-SPINE.md:201`) says `/CreationDate` and `/ModDate` are *"omitted **unless a date
arrives through `params`**"*. Until Story 1.7 `params` did not exist, so the condition was
**unreachable**. Story 1.7 creates the trigger and deliberately does not wire it.

**Story 3.7 already owns it** (`epics.md:1074–1100`): its user story says *"pin document dates
reproducibly"*, its **Covers** line names **AD-7**, and it carries criteria for `SOURCE_DATE_EPOCH`
being read by the CLI and passed in as a parameter — with the library core still reading no
environment variable — plus the negative case (*"no date supplied by any route → omitted"*).

**Forcing function — RE-KEYED at Story 2.1 (D-2.1.5).** Story 1.7 originally added an absence
tripwire keyed on the PATH `folio-go/cmd/` existing at all, as a proxy for "the CLI that reads
`SOURCE_DATE_EPOCH` has arrived." That key was **broader than the purpose**: `cmd/` has more than one
legitimate tenant, and Story 2.1's own build-time tooling (`cmd/gentrie`, `cmd/gencorpus`,
`cmd/genbreaks`) tripped it despite having nothing to do with AD-7 or params-date wiring — a measured
false positive, confirmed independently (`TestAbsencesProductionScan` failing, naming
`absence-cmd-dir`, before this re-key).

**The row is now keyed on its trigger, not on a path:** `SOURCE_DATE_EPOCH` must not appear in any Go
source under `folio-go/` until this is settled. Implemented as a new check KIND
(`absenceKindContent`) in `lint/internal/rules/absences.go`, rule id **`absence-source-date-epoch`**
(was `absence-cmd-dir`) — it scans `.go` files under `folio-go/` (excluding `testdata/`) for the
literal string `SOURCE_DATE_EPOCH`, rather than checking whether a directory exists. Red-proofed by
injection at the real repo location (a scratch reference added under `folio-go/`, observed
`TestAbsencesProductionScan` fail naming the new rule, then removed) and by a permanent fixture
(`folio-go/testdata/lint/absences/violating/folio-go/internal/paramsdate/placeholder.go`, replacing
the old `.../folio-go/cmd/placeholder.go`). The coverage witness (`AbsencesStats.ChecksEvaluated`)
was verified to still count this row: it increments once per entry in `absenceChecks` regardless of
which check kind that entry is, and `TestAbsencesChecksIncludeAllFourEntries` still pins the rule ids
by name (now including `absence-source-date-epoch`), so a silently shrunk list still fails loudly
either way. *(Renamed and reduced from five to four by Story 2.2, which retired `absence-fonts-dir`
when it shipped the faces that tripwire existed to force — see DW-2.)*

**The general rule this produced** (recorded in Story 2.1's Dev Notes too): key a guard on its
purpose, not on a proxy for its purpose. Where the key is broader than the purpose, the gap is where
false positives live — and a false positive in a guard invites exactly the workaround (weakening the
guard) that erodes it fastest. `cmd/folio` — the CLI itself — will still trip this the moment Story
3.7 writes `os.Getenv("SOURCE_DATE_EPOCH")` anywhere under `folio-go/`, regardless of what its `cmd/`
subpackage is named.

**Trade-off, named explicitly (this story's code review, Finding 14) so 3.7's reviewer checks for
it**: purpose-keying traded an UNEVADABLE predicate for an EVADABLE one. The old path key
(`folio-go/cmd/` existing) could not be dodged by spelling. The new content key
(`strings.Contains(source, "SOURCE_DATE_EPOCH")`) does not fire on `"SOURCE_DATE_" + "EPOCH"`, on a
constant defined elsewhere and referenced by name, or on a value read from a variable — Story 3.7's
developer meeting a red build now has a cheaper workaround than before existed. This is an accepted
trade (a guard keyed on a real proxy that can occasionally be evaded beats one keyed on the wrong
thing that never fires falsely but also never fires correctly for a legitimate second tenant of the
path), but it is a trade, not a strict improvement, and 3.7's reviewer should specifically check for
`os.Getenv` calls under `folio-go/` (already banned outside `_test.go` by AD-1) rather than trusting
this content match alone.

Story 1.7 also re-scopes `TestRenderHasNoCreationOrModDate` from an unconditional assertion to
*"params carrying no date"*, so 3.7's developer does not meet a red test whose cheapest resolution is
to weaken it — this part is unchanged by the Story 2.1 re-key.

**Blast radius, measured smaller than feared:** only fixtures that **supply a date** would move. A
params-carrying render with **no** date is byte-identical to today, so 3.7's impact is new fixtures
plus any existing fixture that opts in — **not the corpus**.

**How we'd know it was forgotten.** `cmd/folio` existing while `/CreationDate` is still emitted
unconditionally-absent with no params date path.

**DISCHARGED at Story 3.7 (D-000.59, AC13).** All three of D-000.59's parts are now positively
asserted, and `absence-source-date-epoch` — the forcing function above — was removed by
REPLACEMENT in the same commit, together with the ENTIRE content-check mechanism it was the sole
tenant of (D-000.67 part 1: that mechanism carried a second presence precondition,
`AbsencesStats.ContentFilesScanned`, that the roadmap's own 3→2→1→0 schedule never tracked — see
`lint/internal/rules/absences.go`'s doc comment for the full account). Where each part now lives:

- **(a)** `cmd/folio render` reads `SOURCE_DATE_EPOCH` and passes it in as the `documentDate`
  parameter — asserted through the params path, via a genuine subprocess with the env var set in
  the child's own environment, reading the formatted date off the produced PDF bytes:
  `folio-go/cmd/folio/main_subprocess_test.go`'s `TestRenderReadsSourceDateEpochFromEnvironment` and
  `TestRenderSourceDateEpochValueIsHonoured`.
- **(b)** the library core still reads no environment variable — cited, not re-implemented:
  `TestForbiddenImportsProductionScan` (unchanged) and `TestRenderEntryFileHasNoForbiddenImports`,
  widened at Story 3.7 to a test-owned literal set of pure entry points, `{Render, RenderTo,
  Validate}` (`lint/internal/rules/forbiddenimports_test.go`'s `pureEntryPointNames`), with its own
  non-firing control, `TestFindRenderDeclaringFilesExcludesFolioGo`.
- **(c)** with no date supplied by any route, `/CreationDate` and `/ModDate` are absent from the
  produced bytes — `folio-go/render_test.go`'s `TestRenderWithNoDateInParamsOmitsCreationAndModDate`
  (unweakened, all three original cases plus its full forbidden-key list) with a fourth case added
  at Story 3.7 (the CLI run with `SOURCE_DATE_EPOCH` unset, byte-identical to no params at all):
  `folio-go/cmd/folio/main_subprocess_test.go`'s
  `TestRenderWithSourceDateEpochUnsetIsByteIdenticalToNoParams`.

The stale test name this entry's own prose carried (`TestAbsencesChecksIncludeAllFourEntries`) was
already renamed twice more by the time Story 3.7 opened it (five → four → three → two entries); it
is `TestAbsencesChecksIncludeTheRemainingEntry` as of this discharge, pinning the ONE entry
`absenceChecks` now holds (`absence-designer-project`) — DW-2's own remaining artifact, unrelated to
this entry.

### DW-11 — S4's opaque-name coverage is thin: 2 genuinely-uncoverable sourced items on its most fragile path
- **Raised by:** Story 2.1's re-measurement under D-000.17 (a floor reported unmet, not filled)
- **Corrected by:** Story 2.1's finisher, per its second QA review (Minor 1 and Major 5) — the load-bearing
  count this entry originally carried (8) conflated two different properties. See *Corrected count*
  below.
- **Owner:** **Epic 2's later stories and Epic 4's golden-report work** — add genuinely-opaque sourced
  Thai personal names as they are found
- **Status:** open, and deliberately visible

P6g's floor asked for **≥20** genuinely opaque (zero-interior-break) sourced Thai personal names.
The generator's own P6g count (its literal criterion — "the unconstrained matcher proposes no interior
break") is **7** (was reported as 8 before this correction; see below). Per D-000.17 this was
**reported unmet, not filled.**

**Corrected count — the honest load-bearing figure is 2, not 7 and not the original 8.** The second QA
review (Minor 1) measured that P6g's criterion is satisfied by two structurally different populations
that must not be conflated:

| Surname | Whole dictionary entry? | Genuinely uncoverable (the hard path P2 fails on)? |
|---|---|---|
| `ดอเลาะ` | no | ✅ yes — independently attested (Thai-Malay/Muslim regional surname) |
| `แนแซ` | no | ✅ yes — independently attested (Thai-Malay/Muslim regional surname) |
| `ชินวัตร`, `จิราธิวัฒน์`, `หวั่งหลี`, `ประยูรวงศ์`, `ทวีสิน` | **yes — all five** | ❌ no — these exercise the OPPOSITE path (whole-word match, nothing to override) |

An eighth item, `ฉั่วสมบูรณ์` ("a plausible Sino-Thai family name" per its own original comment), was
**removed from this bucket entirely** by the finisher (D-000.17's "may not invent items to reach a
number" applies to attestation, not just to obsolete-character padding): "plausible" is not sourced,
so rather than retroactively claim it was attested, `cmd/gencorpus/main.go` now labels it
`synthetic_probe` and excludes it from every genuine floor, exactly like the 38 obsolete-consonant
probes. This is why the generator's own P6g figure is 7, not 8.

**So the true load-bearing count for S4's most fragile path (the one P2 demonstrably fails on) is 2,
not 7 and not 8.** The five whole-dictionary-entry names satisfy P6g's literal wording but exercise the
*other* polarity P6g exists to guarantee (nothing proposed, nothing to override) — they are real and
correctly counted under P6g's criterion, but they carry none of the risk this deferred item is about.

**Why the shortfall is not free — measured.** `ฅ (U+0E05)` appears in **2 of 62,107** dictionary words,
so the 38 synthetic obsolete-character strings are **near-trivially uncoverable**: nothing can
partially match inside them, and the atomic-run rule succeeds **the easy way**. The real opaque names
that produced violations (`ดอเลาะ`, `แนแซ`) are built from ordinary characters appearing in **thousands**
of words — which is **why** they violated, via the resume scan landing on a spurious short match inside
a run already declared uncoverable. **That is the hard path, and the path P2 demonstrably fails on.**

**So the 2 genuinely-uncoverable, independently-attested sourced opaque names are load-bearing for
S4** — currently the **only** items covering the path where P2 breaks. **The 38 synthetics (plus the
one reclassified name) cover the easy path and must not be counted as substitutes.**

**Context that explains the shortfall rather than excusing it:** 115 of 122 sourced names (**94%**)
decompose into recognisable morphemes, because Thai naming convention favours composing from meaningful
words. **Genuinely opaque real names appear to be a real minority of the language, not a sourcing-effort
gap.** That measured fact is now available to whoever specifies S4's adequacy criteria.

**How we'd know it was forgotten.** S4 still carrying only 2 genuinely-uncoverable, independently-attested
opaque items when Epic 4's golden report ships.

#### Story 2.4's discharge — AC18. **No new items were added. The load-bearing count remains 2.**

Story 2.4 falls inside DW-11's stated owner window ("Epic 2's later stories"), so it owes an answer in
writing rather than silence. The answer is: **none were found, and none were invented.**

The dev agent had no access to a sourced, independently-attested register of Thai personal names, and
D-000.17 — reinforced by D-2.1.15 Major 5's own precedent, where "a plausible surname" was demoted to
`synthetic_probe` rather than counted — forbids manufacturing attestation to reach a number. Adding a
name that *looks* opaque would move the figure from 2 to 3 while moving the evidence not at all, which
is the failure this entry exists to keep visible. **`ดอเลาะ` and `แนแซ` remain the only two.**
`corpus.json` and `cmd/gencorpus/main.go` are byte-unchanged by Story 2.4.

**What did change, and why it is not a discharge.** Story 2.4 closed the P2 defect these two items were
the load-bearing witnesses *for* (26 violations across 17 corpus items to 0, D-2.1.9). That removes the
**symptom** DW-11 was tracking the risk of, and it does not remove the **thinness**: coverage of the
hard path still rests on two attested items, and a future regression in that path would still be
detected by only those two. Story 2.4's new fixture `fixtures/expected-breaks/` exercises the same path
from a second direction — `ดอเลาะ`, `แนแซ`, `ชัยวัฒน์` and `ฉั่วสมบูรณ์` all appear there as
zero-break labels — but that is a **conformance fixture, not the corpus**, it contributes nothing to
P6g, and it adds no attested name. **DW-11 stays open, at 2.**

#### Story 2.6's answer — AC10. **No new items were added. The load-bearing count remains 2.**

Story 2.6 also falls inside DW-11's owner window (*"Epic 2's later stories"*), so it owes an answer in
writing rather than silence. The answer is the same as 2.4's, and for the same reason: **none were
found, and none were invented.**

Nothing in Story 2.6 goes near the corpus. Pagination consumes `packLines`' output; it does not compute,
produce or consume a break-opportunity vector, and `packLines` itself takes no vertical quantity and
runs **before** the vertical model — so there is no route by which a pagination change could either
require a corpus item or make one newly available. `fixtures/thai-break-corpus/corpus.json` and
`cmd/gencorpus/main.go` are **byte-unchanged**, and `TestCorpusMeetsP6ExerciseFloors` still reports
`P6g … got 7, need >=20` with the stats line `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`,
character for character (Story 2.4's AC5 requires exactly that).

The dev agent had no more access to a sourced, independently-attested register of Thai personal names
than 2.4's did, and D-000.17 forbids manufacturing attestation to reach a number. **`ดอเลาะ` and
`แนแซ` remain the only two. DW-11 stays open, at 2.**

The one thing worth adding for the gate: Story 2.6's new fixture `fixtures/multi-page/` is **all-Latin
by construction** and creates **no Thai judgment of any kind**, so it neither contributes to P6g nor
adds a third sign-off obligation. That was a deliberate choice recorded in the fixture's README, not an
accident of what the document happened to contain.

#### Story 4.7's answer — AC12. **No new items were added. The load-bearing count remains 2, and this entry's own forgetting-signal has now FIRED.**

Story 4.7 is the second address in DW-11's owner line (*"Epic 2's later stories **and Epic 4's
golden-report work**"*), and it is also the event this entry named as its forgetting-signal:
*"S4 still carrying only 2 genuinely-uncoverable, independently-attested opaque items **when Epic 4's
golden report ships**."* **The golden report has now shipped, and S4 is still carrying 2. The signal
has fired, and it is recorded as fired rather than quietly re-owned to a later story.**

The answer to the question itself is the same as 2.4's and 2.6's, and for the same reason: **none were
found, and none were invented.** The dev agent had no access to a sourced, independently-attested
register of Thai personal names, and D-000.17 forbids manufacturing attestation to reach a number.
Adding a name that *looks* opaque would move the figure from 2 to 3 while moving the evidence not at
all. **`ดอเลาะ` and `แนแซ` remain the only two. DW-11 stays open, at 2.**

Measured at Story 4.7's baseline (`df8cbcc`) and again after its work, with
`env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` from `folio-go/`:
`TestCorpusMeetsP6ExerciseFloors` reports `P6g (opaque names) floor not met: got 7, need >=20` with the
stats line `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, character for character — unchanged
in both directions. `fixtures/thai-break-corpus/corpus.json` and `cmd/gencorpus/main.go` are
**byte-unchanged** by this story. The P6g floor was **not** touched: turning it green by invention is
exactly what D-000.17 bans, and the guard cuts both ways —
`TestCorpusP6StatsMatchDeclaredBaseline` pins those literals and reddens on an **improvement** too, so
a genuinely-attested addition would require a deliberate baseline edit in the same commit.

**What Story 4.7 adds to the record, and it is not a discharge.** The four statement goldens reuse
Thai strings **already frozen** in `fixtures/expected-breaks/expected_breaks.json` — the engineering
lead's ruling on this story, on the ground that minting new Thai strings would create a **second
signed Thai-break corpus**, and two signed authorities over the same rules can disagree. So this
story adds **no** Thai judgment to S4, **no** new attested name, and **no** third Thai sign-off. Two
of the four names DW-11 discusses (`ดอเลาะ`, `แนแซ`) do not appear in the statement fixtures at all;
the string this story does assert against is `thai-001` (`ประเทศไทย`), which is an ordinary
two-headword sequence and carries no opacity claim.

**A note for whoever specifies S4's adequacy criteria next, since the signal has fired.** The
measured context in this entry has not changed: 115 of 122 sourced names (94%) decompose into
recognisable morphemes, so genuinely opaque real names appear to be a **real minority of the
language, not a sourcing-effort gap**. A floor of 20 may be asking the corpus for a population the
language does not contain. That is a question about the FLOOR, not about the corpus, and it is the
owner's — this story neither answers it nor uses it as a reason to stop reporting the shortfall.

#### Story 4.7's DW-13 measurement — recorded here beside the entry that asked for it

DW-13's sizing was scheduled by the Epic 3 boundary gate to run **during Story 4.7** (see DW-13's own
entry). It ran. Measured on the committed goldens, summing every `FontFile2` stream's `/Length1` and
compressing each program with zlib at level 9:

| document | file size | `FontFile2` programs | uncompressed | Flate | saving |
|---|---|---|---|---|---|
| `statement-1` | 76,740 B | 3 | 65,740 B (85.7% of file) | 14,277 B | 51,463 B — 78.3% of program bytes, 67.1% of file |
| `statement-5` | 127,343 B | 3 | 77,452 B (60.8%) | 22,487 B | 54,965 B — 71.0% of program bytes, 43.2% of file |
| `statement-20` | 269,804 B | 3 | 77,452 B (28.7%) | 22,487 B | 54,965 B — 71.0% of program bytes, 20.4% of file |
| **`statement-50`** | **555,629 B** | **3** | **77,452 B (13.9%)** | **22,487 B** | **54,965 B — 71.0% of program bytes, 9.9% of file** |

**The answer DW-13 asked for: the 50-page statement with real CJK content carries 77,452 bytes of
uncompressed `FontFile2` payload, and Flate would save 54,965 of them — 9.9% of that document's
total size.**

**Recommendation: leave it. DW-13 closes cheaply.** Three measured reasons, in order of weight:

1. DW-13's own stated threshold is met on the "leave it" side: *"If 4.7's 50-page CJK payload comes
   back in the tens of KB rather than the hundreds, the honest recommendation is 'leave it'."* It came
   back at **76 KB** — tens, not hundreds.
2. **The font payload does not scale with the document.** It is *constant* at 77,452 B across the 5-,
   20- and 50-page statements, because the subsets are identical — the cost is per FACE, not per page.
   So the relative benefit of compressing it **falls** as documents grow: 43% of the 5-page file, but
   under 10% of the 50-page one. The flagship document is the case where compression helps least.
3. Adoption would move **every golden in the repository**, invalidate **all three** human sign-offs
   (`shaped-text`, `expected-breaks` and this story's four-document statement record), and require
   narrowing `lint/internal/rules/nocompressor.go` — a guard that has been proved to fire.

**Adoption remains the project owner's decision, batched with the other Epic 4 close decisions.**
This story measured and did not adopt; nothing in `internal/pdf` or `lint/` was touched.

#### D-4.7.5 — DW-13 is NOT adopted, and the reason to record is NOT the 9.9%

The three measured reasons above are real, and they are all about SIZE. **They are not the reason the
answer is no, and recording them as though they were would leave a future reader thinking a bigger
saving would settle it.** It would not.

**The reason is byte-identity.** Compressing `FontFile2` makes every golden's bytes depend on **Go's
`compress/flate` producing identical output for identical input across Go versions** — and that is
**not part of Go's compatibility promise**. Flate's implementation has changed between Go releases
before, and it may change again. This is exactly the carried risk R4 the `no-compressor-import` rule
(D-1.8.1) exists to keep closed: *"compressor output is stable by observation, not by contract."*

**The risk is CROSS-UPGRADE, not cross-target, and the distinction matters because it decides what the
matrix can tell you.** The toolchain is pinned (`go.mod`'s `toolchain go1.26.0`), so all four targets
would compress identically and the four-target matrix would agree — green, and green for a reason that
has nothing to do with the property at stake. The failure arrives at the **next Go bump**, silently
moving every golden in the repository, invalidating every human sign-off, and presenting as a
mysterious mass re-record rather than as a toolchain event.

**RE-ENTRY IS A CONDITION, NOT A DATE.** Do not re-open this on a schedule and do not re-open it
because the size numbers moved. Re-open it when, and only when, this is true:

> The compressed output of a FIXED input is pinned as a TEST-OWNED LITERAL — a byte string committed in
> this repository, compared against `compress/flate`'s live output, so that a Go upgrade that changes
> the compressor reddens a named test **before** it moves a single golden.

With that pin in place the risk becomes visible and the size argument can be had on its merits.
Without it the answer stays no regardless of what the saving turns out to be.


### DW-12 — Every later pinned instance inherits AC7's golden + matrix obligation
- **Deferred by:** Story 2.2 (ruling D-2.2.1)
- **Owner:** **whichever story next adds a pinned instance of a shipped variable face** (on current
  planning, a candidate is Bold — a `wght`-axis instance — should it arrive; not committed to any
  numbered story here)

Story 2.2 ships exactly one pinned instance per shipped face — each face's DEFAULT instance (Bold
and other non-default named instances are explicitly out of this story's scope). D-2.2.1's binding
standing condition: **value-dependence is the whole hazard** — a clean four-target result on one
(face, pinned instance) pair says nothing about a different pair of the SAME face, because
instancing arithmetic (gvar/avar interpolation) is value-dependent (D-2.2.0's own measured limits).

**So:** any later story that adds a NEW pinned instance of a face already shipped (e.g. a Bold
weight) must (a) add its own golden — the embedded, instanced program's own digest, the same shape
`folio-go/fixture_test.go`'s `TestMultiScriptFallbackGoldenFixture` already records for Story 2.2's
three default instances — and (b) re-run the full four-target matrix in that story, not defer it to
an epic boundary. This is a standing obligation this story registers so a later reader does not have
to rediscover why (D-000.4's per-story override list already gets amended each time a story is
identified as one; this is the reasoning that identifies the NEXT one before it is drafted).

**How we'd know it was forgotten.** A story ships a new pinned instance of an existing shipped face
with no new golden recorded for it, and no four-target matrix re-run logged in its Delivery Log.

---

### DW-13 — Size the uncompressed-`FontFile2` payload cost against real CJK content, then put it to the owner

**Owner:** orchestrator to schedule the sizing; **the adoption decision is the project owner's.**
**Raised at:** Story 2.2, while verifying the embedded programs' table sets.

**The observation.** Folio's `FontFile2` streams ship **uncompressed** — `/Length` equals `/Length1`,
no `/Filter` — while the project uses `/FlateDecode` with `/Predictor 15` elsewhere.

**It is deliberate and mechanically enforced, not drift.** `lint/internal/rules/nocompressor.go`
(`no-compressor-import`) forbids any file under `folio-go/` importing `compress/flate`,
`compress/zlib` or `compress/gzip`, with a retained violating fixture at
`testdata/lint/no-compressor/violating-compressor/bad.go`. Per D-1.8.1, *"no compressor is invoked"*
is **the mechanism that keeps R4 closed** — `acceptance.md:83`: *"compressor output is stable by
observation, not by contract."* The image route embeds each file's **own already-compressed bytes**;
it never invokes a compressor. **Compressing font streams is therefore impossible without retiring or
narrowing a guard that has been proved to fire.** Nothing to change in Story 2.2; its report cites the
rule id and D-1.8.1 and stops.

**Why it still deserves sizing.** The fixtures hide the magnitude — the CJK subset is **732 bytes** —
and that is exactly why the question read as invisible. Rough bound: the static CJK face is 10.6 MB
over ~20k glyphs, so roughly 300–800 bytes of `glyf` per character; a Chinese bank statement plausibly
uses 300–800 distinct characters, giving **~150–400 KB embedded per PDF**, against which Flate on
`glyf` typically saves about half. **Honest expectation: tens to low-hundreds of KB per document —
real, bounded, and quite possibly not worth its cost.** Measuring may well return "leave it", which is
a good outcome and cheap to reach.

**Two constraints to settle before anyone proposes a clever middle path:**

1. **The image passthrough precedent does not extend here.** Images work because the file arrives
   already compressed and we embed its own bytes. A font subset is **synthesised at render time**, so
   there are no pre-existing compressed bytes to pass through — and WOFF2 is not available, since
   `FontFile2` requires raw TTF. **Along that axis the choice is binary:** invoke a compressor or do
   not. There is no third way, and someone will propose one.
2. **There is a genuine third option on a different axis** — a **vendored, version-pinned DEFLATE
   implementation** instead of `compress/flate`. That converts R4 from *"we cannot control the
   compressor"* into *"we pin the compressor exactly as we pin the toolchain"*, which is
   philosophically identical to AD-22's move and arguably **more** stable than stdlib, whose flate has
   been tuned across releases while a pinned vendored copy would not be. Costs: a new dependency
   through AD-26's licence check and D-1.5.1's allowlist, plus narrowing the guard to permit exactly
   that one import — which keeps the decision **visible** rather than dissolving it.

**Framing for the owner, not to be over-stated:** adopting compression **widens** R4 rather than
creating a new risk. Golden hashes are *already* toolchain-sensitive — AD-22 makes a toolchain bump a
release event requiring re-measurement. Compression makes **more** of the output sensitive, and
sensitive to a component historically tuned across releases. That is a difference of **degree**, and
the owner should weigh it as one.

**Also do, cheaply and now:** state the **consequence** beside `acceptance.md:83`'s R4 note — *font
programs and content streams ship uncompressed, so Folio's PDFs are deliberately larger than a typical
producer's.* The document already records the **reason**; what is missing is what follows from it,
which is the sentence someone will otherwise re-derive — or "fix".

**SCHEDULED at the Epic 3 boundary gate.** The Epic 3 gate found this entry had never been scheduled
at all, despite naming the orchestrator as the owner of scheduling it — an owner line that names a
role but no moment is the same failure DW-14 hit with an owner that named an event. **The sizing runs
during Story 4.7**, which is the first document in the programme with real CJK content in volume (its
own AC requires *"Latin, Thai and CJK text in the same table"* at 1, 5, 20 and 50 pages) — that is the
measurement this entry has been asking for, and 4.7 produces it as a by-product whether or not anyone
asks. **The adoption decision then goes to the project owner batched with the other Epic 4 owner
decisions, not as a separate interruption.** If 4.7's 50-page CJK payload comes back in the tens of KB
rather than the hundreds, the honest recommendation is "leave it", and this entry closes cheaply.


---

### DW-14 — `/ToUnicode` emits one unbounded `beginbfchar` section; the spec caps a section at 100

**Raised by:** Story 2.3's QA review (Finding 7, Nit). **Deferred by:** Story 2.3's finisher.
**Owner:** ~~the Epic 2 boundary gate, or Story 2.4 if its corpora reach the limit first~~ —
**RE-OWNED at the Epic 3 boundary gate to Story 4.2.** Both original owners are spent events: Story
2.4 measured and did not trigger (below), and the Epic 2 boundary gate **ran and closed without
re-owning this entry**, which is how it survived a whole epic with nobody holding it. Recorded as
D-000.71's neighbour class: an owner that is an event, rather than a person or a story, stops existing
the moment the event passes, and nothing notices.

**What.** `internal/pdf.buildToUnicodeCMap` emits the whole CMap as a single
`N beginbfchar … endbfchar` block with `N = len(face.ToUnicode)`. The ToUnicode CMap specification
limits one `beginbfchar`/`endbfchar` section to **100 entries**. A document whose face needs more than
100 distinct CIDs would emit a section a strict validator rejects.

**Why it is deferred rather than fixed in 2.3.** It is **pre-existing, not a 2.3 regression** — the
pre-2.3 derivation had the same unbounded shape, so the defect is inherited and charging it to this
story would be wrong on the record. It is also not reachable by anything folio ships today. Measured
across every text fixture at this commit, every section is well under the cap:

```
font-text              : 25
multi-script-fallback  : 4, 1, 1
shaped-text            : 14, 7, 28
```

**Story 2.4's measurement — AC17. NOT TRIGGERED; this entry stays open.** DW-14 named 2.4 as its owner
*"if its corpora reach the limit first"*, and a wrapped Thai paragraph was the plausible input that
would. It does not. Measured on the produced bytes of the new `fixtures/wrapped-text/` render (four
text elements, three scripts, all three shipped faces embedded):

```
wrapped-text           : 28, 18, 38
```

Largest section **38**, against a cap of 100. The fixture was **not sized to duck the cap** — its box
widths were chosen to force wrapping (the numbers are in the story's AC15 record) and the section sizes
are simply what that produced; the largest is comparable to `shaped-text`'s existing 28.

`folio-go/wrapped_text_fixture_test.go`'s `assertToUnicodeSectionsUnderCap` now measures this on every
run and **fails loudly if any section exceeds 100**, with a message saying stop-and-escalate rather
than fix-in-place — because the remedy (chunking into ≤100-entry sections) moves the golden hash of
every document over the cap, and this entry's own text asks that it *"land with a deliberate re-record
rather than as a drive-by"*. So the trigger is now a test rather than a reader's vigilance.

**Story 2.6's measurement — AC11. NOT TRIGGERED; this entry stays open.** DW-14's owner is *"the Epic
2 boundary gate, or Story 2.4 if its corpora reach the limit first"*, and Story 2.6's creator flagged a
new plausible trigger: **a multi-page fixture is the first document in the repository that could
plausibly reach 100 distinct `(glyph, text)` pairs in one face**, because the entry source is now
per-(glyph, cluster text) rather than per-glyph and a longer document has more clusters. Measured on
the produced bytes of `fixtures/multi-page/expected.pdf`:

```
multi-page             : 45          (one section, one face)
```

Largest section **45**, against a cap of 100. **Reported in both directions (D-000.49), because the
figure understates the headroom in one way and overstates it in another and both matter:**

- It is the **largest single section recorded so far** — above `wrapped-text`'s 38 and `shaped-text`'s
  28 — and it comes from a document that is only **two pages** and **single-face**. The trend DW-14
  tracks is confirmed, not contradicted.
- But it is **one face and all-Latin**. `wrapped-text`'s 38 is the largest of *three* sections across
  three faces, so per-face this document is not the outlier the raw number suggests. A three-script
  multi-page document would be the real test, and none exists.

**The honest reading**: 45 of 100 on a 29-line, single-face, two-page document means a document of
roughly **twice the length in one face** would reach the cap — which is an ordinary report, not a
pathological one. This is the closest any artifact has come, and the trigger is now foreseeable rather
than hypothetical. **The fix stays the gate's**, per this entry's own request that it *"land with a
deliberate re-record rather than as a drive-by"*: chunking into ≤100-entry sections moves the golden
hash of every document over the cap.

**DW-14 stays open, and its risk is now higher than when it was written.**

**Why it is worth a standing entry rather than a passing mention.** Story 2.3 is the story that both
**rewrote this function's entry source** (CIDs are now allocated per (glyph, cluster text), so the
entry count is no longer bounded by the glyph count) **and produced the largest section to date (28)**.
Story 2.4's larger Thai corpora push it further in the same direction. The trend and the rewrite are
both this story's; only the fix is not.

**Fix when taken:** chunk into sections of at most 100 entries. It is a local change to one emitter
with no effect on any document currently under the cap — but it **will move every golden hash of any
document that exceeds it**, so it wants to land with a deliberate re-record rather than as a drive-by.

**Epic 3 boundary gate — the trigger is no longer foreseeable, it is scheduled.** Story 4.7 renders the
golden Customer Account Statement at **1, 5, 20 and 50 pages**, with *"Latin, Thai and CJK text in the
same table"* (4.7's own AC), on data that varies row by row. Entries are allocated per
**(glyph, cluster text)** pair since Story 2.3, so distinct content — not page count — is what fills a
section; a 50-page statement of varying transaction descriptions across three faces is the document
this entry has been waiting for since 2.3.

**This is why the owner is 4.2 and not 4.7.** Chunking is byte-identical for every document already
under the cap, so landing it **before any golden fixture is recorded costs zero re-record**. Landing it
at 4.7 instead means re-recording the golden report's hashes at four page counts across four targets —
inside the story that **is the C4 gate**, where a hash re-record is the single most expensive edit in
the programme. The fix does not get cheaper by waiting and it is about to get much more expensive:
**land it at 4.2, with the existing `assertToUnicodeSectionsUnderCap` as the witness that nothing
moved.**

**One thing 4.2 must measure rather than assume:** `fixtures/page-count-20/` is a 20-page document
already in the matrix's ten and its assert is green, so page count alone plainly does not fill a
section — that fixture repeats a template. Do not read its green as headroom for 4.7's varied data.

#### Story 4.7's measurement — DW-14 is DISCHARGED (D-4.7.4)

`TestNoRealToUnicodeSectionExceedsTheCap` picks every `fixtures/*/expected.pdf` up by directory walk,
in the ordinary suite. Run over the corpus with the four statement goldens in it:

```
DW-14 real-corpus witness: examined 13 fixture(s), 11 carrying at least one /ToUnicode section,
25 section(s) total; observed maximum section size = 55 (cap 100)
```

Per document, the three faces' section sizes are `[Noto Sans 55, Noto Sans SC 41, Noto Sans Thai 25]`
for `statement-5/20/50` and `[55, 1, 25]` for `statement-1`. **The corpus maximum rose from 45 to 55.
The cap was not crossed. 55 is the number to measure the next movement against** — record it, so a
future story sees a real figure rather than a prediction.

**THE MECHANISM, WHICH IS WHAT THIS ENTRY IS ACTUALLY FOR (D-4.7.4). A `/ToUnicode` section's size
follows the number of DISTINCT GLYPHS IN A FACE'S SUBSET — not pages, and not rows.** The three
sections are **identical** across 95, 425 and 1085 rows and across 5, 20 and 50 pages: fifty pages of
transactions add no glyph the first page did not already need.

That single sentence explains **both halves of this entry's own warning**, which had until now been
recorded as two separate observations:

- *"a document of roughly twice the length in one face would reach the cap"* — right about the
  direction, wrong about the axis. Length is a proxy for distinct glyphs and a poor one; a long
  document that repeats vocabulary adds nothing.
- *"`page-count-20` repeats a template — do not read its green as headroom for 4.7's varied data"* —
  right, and now explicable: a repeated template has a small glyph inventory, so its green really did
  say nothing. But 4.7's varied data turns out to be varied in the wrong dimension. Its descriptions
  vary; its ALPHABET does not.

**What WOULD cross the cap, stated so the next reader can recognise it:** many DISTINCT GLYPHS IN ONE
FACE. Latin reaches ~55 and stops, because Latin has about that many distinct characters in ordinary
business prose. **A CJK-heavy statement is the realistic case** — Chinese has no small alphabet, so a
document whose Chinese content is genuinely varied rather than drawn from a ten-phrase cycle climbs
one entry per distinct character. This family's Noto Sans SC subset is 41 from forty-one distinct
glyphs; two and a half times that vocabulary crosses 100.

**The chunker is NOT dead code, and this entry must not be read as saying so.** `internal/pdf`'s
chunking (Story 4.2, `textdoc.go:548,626`) is the reason a document that does cross the cap emits a
conformant file instead of a broken one. It has never fired in a committed golden; it is
`internal/pdf/tounicode_chunk_test.go`'s subject and is exercised there. Discharging DW-14 discharges
the WATCH, not the mechanism.

**Do not** "fix" this by capping the number of CIDs; the CID allocation is D-2.3.2 and is correct.

---

### DW-16 — `pagemodel.ShapedGlyph.CID` is not always a glyph id, and the table needed to interpret it is not in the page model

**Owner:** ~~the first non-PDF renderer story, which is its natural forcing function~~ —
**RE-OWNED at the Epic 3 boundary gate (D-000.73).** That owner is a story that does not exist and is
not scheduled: measured through Epic 6, not merely grepped for. The owner is now the **guard** at
`folio-go/glyphid_arch_test.go`, which fires on the real condition. **The shape ruling landed**
(D-000.73); the **option 1 / option 2 fork is with the project owner**, batched into Epic 4 planning
with DW-4 and DW-13.
**Raised at:** Story 2.5 (reviewer Finding 2, a Major; finisher DEFER — recorded, not re-architected).
**Nothing regressed.** The allocation is Story 2.3's [[D-2.3.2]], unchanged. Story 2.5's type move
relocated it into `internal/pagemodel` and thereby made it visible.

**The property, precisely.** `pagemodel.ShapedGlyph.CID` carries **two different kinds of value**,
and the field's name and doc comment describe only the first:

- In `buildShapedPDFRuns`' **base block** (`internal/pdf` is not involved; the site is `render.go`'s
  `cid = newGID`, guarded by `state.baseClaimed`), `CID` **is** the subset glyph id. A renderer
  holding the subset font can look the glyph up directly. This is AD-5's "glyph ids", legitimately.
- In the **`default` block**, a **second, synthetic identifier is minted** for a glyph that carries a
  *different source text* — `cid = uint16(sub.NumGlyphs + len(state.extras))`. That value is **not**
  a glyph id in any font. It is an index into `state.extras`, and its ceiling is stated by the code
  itself: *"exceeds Identity-H's two-byte CID ceiling of 65535."*

That branch exists **only** because PDF's `/ToUnicode` CMap maps one CID to one text. So the field's
value range and its second meaning are both **defined by a PDF encoding**.

**What a non-PDF consumer cannot do today.** The `CID → GID` map that resolves the synthetic values —
`state.extras`, and the `pdf.CIDText` entries built from it — is deliberately kept **out** of the page
model, correctly, because it is a PDF construct. The consequence is that a PNG/SVG/HTML renderer
handed a `pagemodel.TextRun` **cannot resolve `Glyphs` back to glyphs**: for any glyph whose text
differs from its first-seen text, `CID` indexes a table the renderer was not given. It cannot detect
which case it is holding either — the two kinds of value are indistinguishable at the type.

**Why this matters more than the field.** AD-5's stated purpose for keeping PDF out of the page model
is *"that absence is what keeps PNG/SVG/HTML renderers possible later."* The page model now carries
a field that partially defeats that, and AC1's substring guard **cannot see it**: `"cid"` is not in
`pdfConceptSubstrings`, and adding it would be the **wrong fix**, because the base-block value is a
perfectly legitimate subset glyph id. There is no lint that closes this; only a ruling does.

**What would fix it** — the ruling needed, not a decision this entry may take:

1. **Renderer-neutral.** `ShapedGlyph` carries a true `GlyphID` plus a separate text association, and
   the PDF writer performs the CID allocation on its own side of the boundary, from those two.
   `internal/pdf` already owns `CIDText`; this moves the *allocation* to sit beside it. Strictly more
   faithful to AD-5, and strictly more work.
2. **Amend AD-5.** Admit an encoding-scoped identifier explicitly, rename the field to say so, and
   record that a non-PDF renderer needs an accompanying table which the page model does not carry.
   Cheaper, and honest, but it narrows what the page model promises.

**Why it is worth an entry rather than a passing mention.** The window in which this is cheap to see
closes as soon as **more producers write the field**. ~~Today there is exactly one
(`buildShapedPDFRuns`), so option 1 is a local change.~~ Every additional producer makes both options
more expensive, and makes the field's dual meaning harder to establish from the code.

**CORRECTION, Epic 3 boundary gate (D-000.6 — the false clause is amended in place, the rest stands).
"Exactly one producer" was true when written at Story 2.5 and has been FALSE since Story 2.7.** There
are **two** non-test construction sites of `pagemodel.ShapedGlyph`, measured:

- `folio-go/render.go:1954` — `buildShapedPDFRuns`, **the allocator.** It mints both kinds: the base
  value (`cid = newGID`, `:1917`) and the synthetic one (`sub.NumGlyphs + len(state.extras)`, `:1936`).
- `folio-go/page_number.go:520` — `resolvePageRunForPage`, Story 2.7's page-number substitution.

**Two groundings, a boundary gate and the engineering lead's own memory all carried the "exactly one"
sentence forward unmeasured** — [[D-000.66]]'s shape in a sentence that announces itself, a dated
measurement re-read as a standing fact.

**The closure is mild, and that changes the pricing rather than the answer.** The 2.7 site is a
**copier, not an allocator**: `buildPageNumberSlot` (`page_number.go:452-453`) reads
`cids[d] = dt.Glyphs[d].CID` out of a digit-table run the allocator already produced. Under option 1
it needs a `GlyphID` plus text instead — and digits are the one case where the text association is
trivially known (*"the digit d"*), so `DigitCID [10]uint16` becomes `DigitGID [10]uint16` and the text
comes free. **Option 1 got slightly more expensive in the most benign way available. It is not
foreclosed.**

**The near-miss worth recording:** `internal/text/shape.go:161` also constructs a `ShapedGlyph{`, and
it is **not** one of the two — it is `text.ShapedGlyph`, a different type in a different package.
`render.go:1993` likewise matches `.CID` but on `pdf.CIDText`. A name-matching instrument reports
three producers and eight consumers here; a `go/types` one reports two and six. This is why the guard
below resolves field identity through the type checker rather than through the spelling `CID`
([[D-000.68]]).

**The forcing function does not exist under any name — measured, not grepped.** The gate's audit
grepped `epics.md` for PNG/SVG/HTML/non-PDF and found zero hits, which only establishes that no story
is *named* that. The lead read the two roadmap stories that could be a non-PDF renderer under another
name: **Story 5.10**'s preview is a controlled `pdfjs-dist` canvas that consumes **the real PDF** and
must hash-match a native render; **Story 5.9**'s canvas paints **pre-broken lines of text** from
engine metrics and *"the browser contributes rasterization only"* — it never touches `TextRun.Glyphs`.
**Through Epic 6 there is no consumer that can be harmed by this defect.** That is what a retirement
would have to rest on — but it is conditional on the landing plan rather than on the design, so it is
recorded as a **falsifier**, not as a closure.

**Stories 4.1 and 4.2 do NOT close the window.** Both add *callers* of the existing shaper→pagemodel
bridge, not new construction sites; a new site would be a deliberate second bridge and neither story's
ACs ask for one. 4.1's borders and padding are rects, not glyphs. There is more slack here than the
gate assumed, not less.

**Do not** "fix" this by adding `cid` to `pdfConceptSubstrings`, and do not re-architect
`buildShapedPDFRuns` opportunistically — neither Story 2.5's AC1 nor its Task 4 asked for it, and the
developer was right not to attempt it.

### DW-17 — Surfacing a returned `Diagnostic` to a human is a presented-interface obligation, not a Go call-graph one — **DISCHARGED by Stories 3.7, 5.12, and 6.6**
- **Deferred by:** Story 2.8 (ruling D-2.8.5)
- **Owners:** **Story 3.7** (the CLI must print the diagnostics it receives), **Story 5.12**
  ("Diagnostics that locate and an interface that can be driven"), **Story 6.6** ("Present a failed
  render honestly")
- **Anti-rot mechanism:** none possible, and none is owed. D-2.8.5 declined an AST guard asserting
  that every `Render`/`RenderTo` call site also reads `.Diagnostics`/its returned slice: the call-site
  population is overwhelmingly tests that render a fixture and hash bytes, which legitimately do not
  care, so the guard would fire on scores of correct sites and its remedy (`_ = res.Diagnostics`)
  would be ceremony training the codebase to discard rather than to check. This entry is the only
  thing keeping the obligation visible until one of its three owners lands it.

**This is not an accepted gap against AD-14 — do not record it as one (D-000.49).** AD-14's
**Prevents** clause is about type fragmentation (*"each area inventing its own error type… CI cannot
assert that every FR41 case is covered"*), not about caller discipline. `folio.Result` satisfies AD-14
completely: one `Diagnostic` type, one channel, returned alongside the bytes, never dropped, never
fatal. What remains open is a **presented-interface** property — *"a clipped-content warning reaches a
human"* — which is not carried by a Go expression and cannot be asserted from the Go call graph
(D-000.21: the property belongs to the artifact that carries it). Story 3.7's CLI, Story 5.12's
located-diagnostics interface and Story 6.6's honest-failure presentation are where it becomes
observable, and each owes an assertion on ITS OWN presented output once built.

**How we'd know it was forgotten.** Story 3.7, 5.12 or 6.6 ships without a test that a `Diagnostic`
folio's render path returns actually appears in what a human sees (CLI stdout, the driven interface,
the failed-render presentation) — a case where `Render`/`RenderTo` returned a non-empty
`Diagnostics`/warning slice and nothing downstream of it printed, logged or displayed any part of it.

**Amended by Story 3.6 (OPEN-1's ruling): this obligation's weight just increased, for one specific
code.** `DiagCodeTextMissingGlyph` (FR41's fifth mode) is minted this story with a render-side
disposition ruled by the engineering lead: the uncovered rune is OMITTED — no glyph, no advance, and
no in-band marker of any kind (never `.notdef`, never a substituted replacement glyph, per AD-8 and
the ruling's three grounds: the chain is document-declared and not guaranteed to cover a substitute;
substitution is the "silent content edit" class AD-8 already rejects; and omission — unlike
substitution — keeps `/ToUnicode` extraction honest). **This makes DiagCodeTextMissingGlyph's case
WORSE than FR44's clipped content**, which this entry's own "not novel" framing (below) was written
against: a reader can at least SEE that clipped content was truncated; here there is nothing at all on
the page — the Warning is the ONLY record the rune ever existed. If DW-17's three owners ship without
surfacing this specific Warning, the defect is not merely unreported, it is INVISIBLE — there is no
artifact-level clue for a human to even suspect something is missing. Story 3.7's CLI, 5.12's
located-diagnostics interface and 6.6's honest-failure presentation each now carry this sharpened
stake explicitly, not only the general "some Diagnostic went unprinted" case this entry originally
named.

**One framing this amendment preserves, so a future reader does not mistake it for a novel hazard:**
FR44's clipped content already gives no in-band page signal either (D-2.8.1 ruled it "clipped at the
box's left/right edges, never reflowed and never dropped" — no marker drawn) — "the defect lives in the
diagnostics, not in the artifact" is the engine's EXISTING posture, not something this story
introduced. What changed is only that missing-glyph's case has NO visible truncation cue at all, where
clipping at least leaves something a reader can notice went wrong.

**The reversal cost, recorded because it decides what "urgent" means here.** The two arms are
asymmetric: omitting now, then substituting later (if a future story wants an in-band marker) would
move the BYTES of any document containing an uncovered rune — but Story 3.6's own corpus-wide
assertion (`folio-go/missing_glyph_corpus_test.go`,
`TestCorpusFixturesProduceNoMissingGlyphWarnings`) guarantees there are NONE in the committed corpus
today, so that reversal costs nothing beyond the AD-8 amendment it would need anyway (an implicit
terminal, guaranteed-coverage face in every chain — which also changes every existing golden hash,
since the chain is part of a `FontSet`'s identity). This is not urgent in the sense of "code debt
accruing interest" — it is urgent only in the sense that a document with the defect exists silently
until one of DW-17's three owners makes it visible.

**Amended by Story 3.6's finisher (Finding 13, QA review): the Warning is emitted PER RUNE OCCURRENCE,
knowingly, as a bound for DW-17's owners to inherit.** `shapeSegments` (`folio-go/render.go`) appends
one `Diagnostic` for every uncovered rune before adjacent uncovered runes are merged into a display
segment — an element whose declared chain covers none of its script produces one Warning per
character (e.g. 200 near-identical Warnings for a 200-character string). This is left as-is by this
story rather than coalesced, for two reasons: (1) no AC or ruling specifies a batching shape, and
inventing one now risks a shape 3.7/5.12's actual presentation needs would not have chosen anyway —
D-2.6.5's guardrail against building presentation machinery a consuming story doesn't yet name; (2)
`Result.Diagnostics` is otherwise unordered-count-sensitive nowhere else in the module, so this is a
presentation concern, squarely DW-17's own territory, not a `folio-go` correctness one. **Whichever of
DW-17's three owners is first to present this Warning to a human must decide the granularity**
(per-rune, per-distinct-rune, or per-element) as part of that presentation design — do not inherit
per-occurrence silently by copying the raw diagnostic list.

**Amended at Story 3.7 (D-3.7.3, this story's own DECISION-3, OVERRULING its creator's
presentation-layer recommendation): the granularity question above is now ANSWERED, in the ENGINE,
as a BEHAVIOUR PRECEDENT — not amortised across three separate presentation-layer implementations.**
`shapeSegments` (`folio-go/render.go`) now coalesces to ONE `Diagnostic` per (element, distinct
rune), in FIRST-OCCURRENCE order (a slice with a linear scan, never a map — AD-1, D-2.8.6's
determinism guarantee on the diagnostics slice's order). No count field, no public type change: the
actionable unit was always the font chain, not the occurrence count. The reasoning for landing this
in the engine rather than leaving it to each of DW-17's three presenters: three implementations of
one rule is precisely the drift hazard this very entry's header names, and the moment to fix it is
NOW, before any presenter exists — after three presenters have each independently coalesced, it is
four places and three conventions to reconcile instead of one. **Story 3.7's OWN presented-output
half of DW-17 (the CLI printing every `Diagnostic` it receives, asserted on stdout/stderr content)
is discharged**: `folio-go/cmd/folio/main.go`'s `printDiagnostics`, tested in
`folio-go/cmd/folio/main_test.go`'s `TestDiagnosticsPrintedOnStderr` (including the missing-glyph
case and its negative control — a clean render prints nothing). **Story 5.12 and Story 6.6 inherit
the coalesced, first-occurrence-ordered form as a property of the `Diagnostic` slice itself** — they
owe only their OWN presentation of it, never a re-decision of its granularity.

**DISCHARGED at the Epic 6 boundary (2026-08-29).** All three presentation owners have landed and
the boundary re-ran their executable evidence. The CLI prints returned diagnostics on stderr;
Preview presents successful-render warnings from the engine's closed diagnostic array with code,
message, element/path location, live announcement, dismiss and locate behavior; and rejected producer
renders use the distinct failure card with preserved code/message/location, Retry, and return-to-canvas
behavior. Viewer, identity, serialization, lifecycle and cancellation-adjacent failures are explicitly
excluded from producer-render provenance. Story 6.7's dynamic registry census executes every registered
warning and error through its production path. The designer unit suite exercises warning presentation,
and the real Chromium suite exercises the located `BINDING_PATH_ABSENT` render failure. No presenter
re-coalesces or reclassifies the engine's diagnostic list.

### DW-18 — `Severity`'s zero value is a VALID severity, so no test can prove the field was ever explicitly set — **RETIRED by Story 3.6 (AC6, R10)**
- **Retired at:** Story 3.6. `severityUnset Severity = iota` now precedes `SeverityWarning` in
  `folio-go/diagnostic.go`, so `SeverityWarning` is **1** and `SeverityError` is **2** — the zero value
  is no longer a member of the valid set. `Severity.String()` gained a `severityUnset` arm. Every
  production `Diagnostic{...}` construction site (render.go, render_error.go) sets `Severity`
  explicitly. `render_clip_diagnostic_test.go`'s three sites lost their "known limitation" comments and
  now carry real coverage. **M8, re-run**: at `4ec1884` (baseline), deleting `Severity:
  SeverityWarning,` from `render.go`'s clip construction site was confirmed INVISIBLE (all clip
  diagnostic tests still passed) — the concrete demonstration that the defect this entry names was
  real. After AC6 landed, the same deletion was re-run and confirmed to REDDEN
  (`TestRenderAndRenderToDiagnosticsAgree` failed with `Severity:Severity(unset)`).
  **Free now, never again**: `folio-go/version.go` declares `Version = "0.0.0-dev"` and `git tag`
  named no `folio-go/v*` tag at the time this story ran, so nothing downstream could have pinned the
  previous integer values (AD-22 — once `folio-go/v0.1.0` is cut, renumbering a public constant here
  becomes a breaking change requiring `folio-go/v2`).
- **Why Task 8 (renumber) had to precede Task 10 (construct the first `SeverityError` values,
  D-3.6.3).** Not diligence — an INSTRUMENT. The hazard this entry names is a WINDOW IN TIME: if
  `SeverityError` values existed while the zero value was still a valid `SeverityWarning`, a
  copy-paste omitting the `Severity:` field would silently downgrade an Error to a Warning with
  nothing able to catch it. Ordering the tasks so the zero value stops being valid BEFORE any
  `SeverityError` is ever constructed means that failure mode has NO INTERVAL in which to occur —
  the same move as making an invalid call fail to compile rather than testing that nobody writes it.
- **Deferred by:** Story 2.8 finisher (review Finding 2, `render_clip_diagnostic_test.go`)
- **Owner:** whoever next touches `folio.Severity` (`diagnostic.go`) — plausibly Story 3.6, which
  mints the next `Diagnostic`-carrying codes and is the first natural place a `SeverityError` value
  gets constructed for real.
- **The defect, measured.** `SeverityWarning Severity = iota` makes `SeverityWarning == 0`, which is
  also the zero value every `Diagnostic{}` literal starts from. The code-review's mutation M8 —
  deleting `Severity: SeverityWarning,` from the one production construction site
  (`render.go:532-533`) — leaves the field unset, which is bit-for-bit identical to a field correctly
  set to `SeverityWarning`. Every assertion in `render_clip_diagnostic_test.go` that reads
  `d.Severity != SeverityWarning` (three sites) is therefore comparing a value to itself under either
  outcome; none can fail no matter which of the two states produced it.
- **Why this story does not fix it.** The only fix that closes the gap is a change to `Severity`'s
  zero-value semantics — e.g. an unexported `severityUnset Severity = iota` ahead of `SeverityWarning`
  so the zero value stops being a valid severity and `SeverityWarning` becomes `1`. That changes the
  numeric value of a public constant on a public exported type, which is the product's front door
  (AD-14) and an owner/engineering-lead call, not a finisher's to make unilaterally while closing out
  a test-quality review. No production code was touched for this finding; `render_clip_diagnostic_test.go`
  gained a comment recording the limitation at each of the three assertion sites instead of a false
  claim of coverage.
- **Anti-rot mechanism:** none exists and none is owed by this story. This entry is what keeps the
  gap visible until `Severity` is touched again.
- **How we'd know it was forgotten.** A future `Diagnostic{..., Severity: SeverityError, ...}`
  construction site ships with the `Severity:` field accidentally omitted (e.g. a copy-paste from a
  `Warning`-only helper), and the render path silently returns it as a `Warning` — AD-14's
  disposition rule ("Error aborts the render, Warning accompanies a successful one") violated with no
  test catching it, because the zero value still reads as a valid, unremarkable `Warning`.

### DW-19 — The lint asset resolver walks a GITIGNORED directory, so it fails in BOTH directions — **RETIRED by Story 3.6 (AC10, D-3.6.5)**
- **Retired at:** Story 3.6, at a mechanism the engineering lead ratified as a deviation from the fix
  shape this entry originally specified — see the D-3.6.5 amendment (decision log), which found the
  literal pre-pass shape implemented and measured strictly worse (104/3 → 103/4, AC10's 0-fail target
  unreachable) and ratified the exclusion mechanism actually shipped instead. `ResolveAssets`
  (`lint/internal/manifest/manifest.go`) now checks, per discovered asset directory, whether git
  actually tracks any file under it (`gitTrackedFileCount`, shelling out to `git ls-files`) BEFORE
  evaluating that directory's LICENSE/NOTICE findings. A directory with zero tracked files (measured:
  `.font-sources/`, still gitignored, still holding real font files on disk, still zero tracked files
  at this story's run) is EXCLUDED from the findings loop entirely — no row, no error — because it is
  not a redistributed asset (AD-26) at all, never having been committed. A TRACKED directory's real
  violation is untouched by this and still fails loudly (verified:
  `TestResolveAssetsStillReportsATrackedViolation`,
  `TestFontsAssetsNoticeRemovalRedProof`/`LicenceRemovalRedProof` — of those two, only
  `TestFontsAssetsLicenceRemovalRedProof` was SILENTLY MASKED before this fix (D-000.70; corrected
  here from an earlier draft that claimed both), since `.font-sources` sorted first alphabetically
  and its own erroneous "no LICENSE* file" message satisfied that test's substring check by
  coincidence, without the test ever reaching its real target directory —
  `TestFontsAssetsNoticeRemovalRedProof` was already failing loudly at baseline for its own reason
  (this story's own baseline table). `lint` moved from
  **104 pass / 3 fail** to **109 pass / 0 fail** (the three named tests plus two new permanent
  regression guards, `TestResolveAssetsExcludesUntrackedDirectoryWithoutError` and
  `TestResolveAssetsStillReportsATrackedViolation`). Not fixed by adding files to `.font-sources/` and
  not by `t.Skip`, exactly as this entry required.
- **Deferred by:** Story 3.1a creator (finding F4), ruled out of scope for 3.1a by the engineering lead
- **Owner:** whoever next touches `lint`'s asset resolution (`ResolveAssets`). Not Story 3.1a — it is
  already building a kernel, a corpus, an oracle and a lint rule, and this is a different subject
  ([[D-000.25]]'s reason for not folding the vendor audit into 2.4).
- **Status:** retired (Story 3.6), at the ratified deviation recorded above, not at the literal fix
  shape stated below — see the D-3.6.5 amendment.

**The defect, measured.** `ResolveAssets` walks `.font-sources/` — **gitignored** (`.gitignore:85`),
**zero tracked files**, the owner's local variable-font scratch directory. Three lint tests
(`TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`, `TestFontsAssetsNoticeRemovalRedProof`)
**fail in this checkout and pass 85/85 in a clean detached worktree** at `b227dda`. Confirmed by
running both.

**Why this is [[D-000.9]] verbatim, at the infrastructure level.** *"The sources were not present"* and
*"the assets are fine"* produce **the same signal**. And it fails in **both directions**, which is the
worst property an instrument can have:

- **Fails RED in a working checkout** — three lint tests red for an environmental reason. A developer
  will correctly diagnose them as noise **and then learn to discount lint reds generally.** That is
  [[D-000.15]]'s erosion dynamic aimed at the whole lint module.
- **Fails GREEN in the dangerous direction** — because behaviour depends on an **untracked** directory's
  contents, **anyone can make these tests pass by putting files there, and nothing in the repository
  records that they did.**

**The fix shape, stated so the entry cannot be closed vacuously.** The resolver must treat *"the asset
root resolved to a path with **zero tracked files**"* as a **scan error** — returned and assessed
**before any findings** ([[D-1.3.3]] (amended)'s shape, and [[D-000.58]]'s rule that a procedure
depending on an environment existing only on one machine is not a procedure).

**Explicitly NOT the fix:** adding files to `.font-sources/`, or making the three tests **skip**. Both
trade a **loud** environmental failure for a **quieter** one — exactly what the Epic 2 gate declined to
do for `.fontgen-venv` ([[D-000.58]]).

**Superseded by the D-3.6.5 amendment.** The literal "scan error before any findings" shape above was
implemented and measured at Story 3.6: it strictly worsens `lint` (104/3 → 103/4) and makes AC10's
0-fail target unreachable, because `.font-sources` sorts before the real violation directories and a
global pre-pass would abort before ever reaching them. The engineering lead ratified the shipped
alternative instead — EXCLUDE an untracked directory from the findings loop entirely (no row, no
error) rather than treat it as a blocking scan error — which preserves the property this entry exists
for (a tracked directory's real violation still fails loudly) without the literal mechanism's
regression. See D-3.6.5 amendment (decision log) for the full grounds, and Story 3.6's own Finding 1
(QA review) for a further, narrower "Required" floor the amendment adds on top of this shape (all
candidate directories untracked is its own scan error), which this story implements separately.

**Story 3.1a's one-line obligation, already discharged in its prompt** ([[D-000.55]]): its Delivery Log
names these three tests as failing for a known environmental reason **before** its run, with the
reason — so 3.1a's red-proof figures stay attributable and nobody later reads three unexplained reds as
evidence about the new denylist rule. **If any of the three is still red for a DIFFERENT reason after
that, that is a finding.**

---

### DW-20 — `folio-go/render_arch_test.go`'s call-graph walker resolves methods and func-typed vars by AST NAME-MATCHING only; a `go/types`-precise version is owed before the `folio-go/v0.1.0` tag

- **Raised at:** Story 3.7, finisher pass, resolving the engineering lead's ruling D-3.7.9(a) on this
  story's review Finding 2.
- **The residual, stated precisely.** `buildFolioCallGraph` (`render_arch_test.go`) now treats every
  method declaration in package `folio` as a graph node keyed by method NAME ALONE, merged across every
  receiver type declaring that name, and resolves a selector call `x.Foo()` to every method named `Foo`
  regardless of `x`'s static or dynamic type. This is a deliberate, safe over-approximation — a spurious
  edge only makes `TestValidateNeverReachesRenderOrInternalPDF` and
  `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt` STRICTER, never looser — but it
  is not the precise property AC1's doc comment describes. A `go/types`-checked version would resolve
  `x`'s actual type and only add the edge that is really there.
- **Why deferred rather than built now.** At Story 3.7, `buildFolioCallGraph` is a `_test.go`-only tool
  scoped to package `folio`'s own root files (D-000.42: no second call-graph builder), and the
  over-approximation costs nothing today — measured, zero methods in package `folio` at HEAD call into
  `internal/pdf`, so the merge-by-name behaviour changes no test's verdict. Building a `go/types` version
  now would be effort spent before the property it protects (`Validate`'s public contract) has frozen.
- **The real trigger, not a vague "eventually."** **Before `folio-go/v0.1.0` is tagged** (AD-22:
  `version.go`'s `Version = "0.0.0-dev"`, no `git tag` naming `folio-go/v*` yet) — because that tag is
  the point at which `Validate`'s public contract (and everything reachable from it) freezes, and a
  `lint` rule over `go/types` is the complete, non-over-approximating version the lead named as the
  eventual replacement (D-3.7.9's own words: *"both guards that actually held this story live in
  `lint`, which type-checks the module; both that leaked are in-module AST scans"*).
- **Owner:** ~~whoever cuts `folio-go/v0.1.0`~~ — **RE-OWNED at Epic 4 planning (D-4.0.2) to
  `TestFolioMethodNamesAreInjective`** (`folio-go/render_arch_test.go`, beside the walker whose
  precondition it asserts). [[D-000.78]] moved the tag to after Epic 6, which would have left this
  keyed to an event three epics away — the fourth instance of the owner shape [[D-000.73]] ruled
  against. `RELEASING.md` item 3 carries it as a **backstop**, never as the trigger.
- **Status:** open, with a mechanical trigger.

**AMENDED at Epic 4 planning — the cost argument above rests on the wrong fact, and the right one was
free all along.** Measured over package `folio`'s non-test root files: **seven methods, and every name
is distinct** — `Severity.String`, `(*RenderError).Error`, `(*RenderError).Unwrap`,
`(*fontCache).get`, `faceSegment.segmentLocal`, `faceSegment.glyphRangeForRunes`,
`faceSegment.advance1000`. So the name→receiver map is **injective**, the merge-by-name is
**lossless**, and `buildFolioCallGraph` **over-approximates nothing at HEAD**. It is not
"safe but loose." It is exact.

That matters because the two facts are not equally durable:

| fact | what it buys | when it expires |
|---|---|---|
| *"zero methods call into `internal/pdf`"* (the entry's original argument) | the imprecision is **unobservable** — no edges, so nothing merges wrongly | **the first time any method touches `pdf`** |
| **injectivity** (the amendment) | the imprecision is **absent** — nothing could merge wrongly | only when two receiver types share a method name — and it **keeps holding after** methods start reaching `pdf` |

The original is also a **dated measurement of the current tree**, which is precisely the shape
DW-16's *"exactly one producer"* had when it went stale for three epics unnoticed. **Injectivity is
the condition; the zero is not, and it comes out of this entry as the cost argument.**

**The hazard framing above is also corrected.** *"A spurious edge only makes the tests STRICTER,
never looser"* is true and is **not** the reassurance it reads as. The failure mode is not a missed
defect — it is a **legitimate commit blocked by an edge that is not really there**, then "fixed" by
someone loosening `TestValidateNeverReachesRenderOrInternalPDF`. **The safe direction is the dangerous
one here**, and the pin is what stops it arriving unannounced.

**The guard's anchor is structural, not a name list** ([[D-000.68]]): it asserts a **property** of the
map — injectivity — so it cannot rot as methods are added, removed or renamed, and it reddens only on
the condition that re-prices this entry. That is deliberately the opposite choice from this
programme's other two censuses, which pin literal sets because those sets are frozen by design; this
set is expected to grow, so a pinned member list would be [[D-3.1a.3]]'s relational case handled
wrongly. Vacuity is covered separately by a floor on the walk's own method count, because an empty map
is trivially injective and reports the same all-clear a healthy one does.

**Red-proved:** a second `String()` on another receiver type → **red**, naming the method and both
receivers with their files; the census floor raised above the true count → **Fatal** on the vacuity
path.

**The replacement is pre-priced, so the next reader need not re-derive it as greenfield.** `lint`
reaches across the module boundary with `packages.Load` today — [[D-000.73]]'s census and the
type-checking rules of [[D-000.75]] — so when this trigger fires, a `go/types` walker is a **marginal
cost on working infrastructure**, exactly as [[D-3.7.9]] anticipated: *"both guards that actually held
this story live in `lint`, which type-checks the module; both that leaked are in-module AST scans."*

### DW-23 — `lint`'s gofmt break has been red since Story 5.10, and CI's permanently-red workflow hid it from two boundary gates
- **Deferred by:** the second Epic 5 boundary gate (2026-08-29), which measured it rather than fixing it
- **Owner:** whoever next touches `lint/` — or the next boundary gate, whichever comes first
- **Status:** OPEN — **now owned by Story 15.2** (`_bmad-output/planning-artifacts/epics.md`, Epic 15),
  added 2026-08-30. That story takes BOTH halves this entry names: the gofmt fix, and the structural
  remedy — the known-red job moves to its own workflow so the guardrails conclusion is readable, and
  the boundary-gate procedure runs gofmt in all three modules and reads per-job conclusions rather
  than the badge. Fixing only the first half was explicitly rejected here and stays rejected there.

**The defect.** `lint/internal/rules/licencegraph_test.go` is not gofmt-clean. The committed file is
byte-identical to its version at `5dddbea` (Story 5.10, "finish exact PDF preview"), so the break has
been in the tree since that story. It is one collapsed single-line `for` loop:

```
for _, name := range []string{...} { if err := os.WriteFile(...); err != nil { t.Fatal(err) } }
```

`gofmt -w lint/internal/rules/licencegraph_test.go` fixes it. It is unrelated to Story 5.13, which
touched no file under `lint/`.

**Why it is worth an entry rather than a one-line fix in passing.** It survived the Epic 5 **and**
Epic 6 boundary gates, and the reason is structural:

1. **The local gate procedure ran `gofmt` in `folio-go` only.** CI runs that step three times, under
   `working-directory:` `folio-go`, `hashmatrix` and `lint` (`.github/workflows/ci.yml`). A gate that
   measures fewer modules than CI cannot certify what CI will report.
2. **CI's workflow status cannot distinguish a real failure from the sanctioned one.** The
   `folio-go-known-red` job is red by design so DW-11's unmet floor stays visible — its own comment
   reads "this job going GREEN is the surprising event, not this job going red." Because that job
   lives in the "Build, vet, and guardrails" workflow, the workflow is permanently red, and a genuine
   `lint` failure beside it is camouflaged. `gh run view` shows `lint` and `folio-go-known-red` both
   failing on the last two pushes to `main`, indistinguishable from the badge.

This is D-000.38's shape — two conditions sharing one signal, so neither can be read alone — applied
to CI job status rather than to a parser. The quarantine design is right; what is missing is that
nothing reads the **per-job** conclusions.

**The fix has two halves, and the second is the one that matters.** Run `gofmt -w` on the file; and
change the boundary-gate procedure to (a) run `gofmt -l` in all three Go modules and (b) read CI's
per-job conclusions rather than the workflow badge. Fixing only the first half leaves the blind spot
that produced it, and the next such break will hide for exactly as long.

**How we'd know it was still wrong.** Another genuine failure sitting green-adjacent in that workflow
for multiple stories — or a boundary gate reporting "CI clean" on the strength of a badge that is red
by design.

---

### DW-25 — two canvas projection bounds abort the WHOLE projection instead of degrading, and Epic 7's own input reaches both; the 512-BYTE value cap binds first — **CLOSED by Story 7.4, 2026-08-30**

- **Deferred by:** Story 7.1 (2026-08-30), which created the reachability and whose own contract
  forbids it the fix — no designer/editor surface work, and no new diagnostic code.
- **Owner:** **Story 7.4**, plus the orchestrator's 7.4 plan-gate checklist as a second standing
  address. Ruled 2026-08-30 — see `epic-7-8-decision-log.md` D-7.4.1 / D-7.4.2 / D-7.4.3. Explicitly
  NOT "Epic 7 close" (D-000.73: an owner that is an event stops existing when the event passes).
- **Severity:** **MEDIUM**, not the `low` Story 7.1's implementer filed it at. The lead went further:
  it is **Story 7.4's own acceptance criterion**, since 7.4's first AC ("the editor accepts and
  preserves multiple lines") cannot be demonstrated until these bounds lift.
- **Status:** **CLOSED** by Story 7.4, 2026-08-30. See the closing note at the end of this entry.

**AMENDED 2026-08-30 by the engineering lead's ruling — the bound named in the original title is NOT
the one that binds.** `page_setup.go:557-560`, in `canvasComponents` (a DIFFERENT function from the
paint loop below), caps a text element's **value** at `maxCanvasPropertyString` = **512 bytes** and
`return nil`s the ENTIRE component list — the whole canvas with no components at all. That is ~80
English words, less than one numbered contract clause. `len()` on a Go string counts BYTES, so Thai and
CJK cost 3 bytes per character and the ceiling lands near 170 characters — for exactly the two scripts
NFR3 makes first-class. Reaching the 256-line cap with real prose requires passing 512 bytes first, so
the value cap fires first for every realistic Epic 7 input; 256 is reachable only by a value that is
almost entirely line feeds.

**AMENDED 2026-08-30 at Story 7.4's plan gate — this entry UNDERCOUNTED on three axes, and one of
them would have made a correct fix inert.**

1. **`maxCanvasPropertyString` has NINE sites, not eight, and TWO of them are body text** — not one.
   `page_setup.go:581` (an element's value) **and `:522` (a fragment's text)**, which nothing had named.
   A value large enough to pass `:581` would still abort the projection at `:522`. Both move to the new
   body-text constant; the other seven (`:211`, `:590`, `:596`, `:640`, `:665`, `:671`, `:686`) are
   identifiers, colours and expressions and stay at 512.
2. **The TypeScript mirrors are FOUR, not the three D-7.4.5 recorded.** The fourth is
   `engine-protocol.ts:152-154`'s `optionalString`, which caps `value` at 512 alongside seven identifier
   keys — `maxCanvasPropertyString`'s exact conflation, reproduced on the browser side. **Without
   splitting it too, a correct Go-side fix changes nothing observable**: the browser rejects the
   projection regardless.
3. **The fragment peak is 256, not ~249**, and the "~73 justified lines" coupling figure holds only for
   a ~240pt column — at full A4 width it is **~31 lines at 11pt**. The geometry-free statement, which is
   what belongs in a record: cumulative fragments ≈ the value's **word count**, so the browser's 512 is
   crossed at roughly **512 words**, at any column width.

**Also found:** `CanvasTextPaint.Overflow` sets a CSS class for which **no rule exists anywhere**, so the
existing degradation flag is invisible to the author. That is why the new truncation flag must state
itself in text rather than only setting a class.

**Root cause:** `maxCanvasPropertyString` does two jobs. At `:211`, `:567`, `:573`, `:617`, `:642`,
`:648`, `:663` it bounds identifiers, colours and expressions — legitimately short. At `:558` it bounds
document body text, which is not. **The fix is to SPLIT it, not to change its value.**

**Ruled shape (D-7.4.2):** degrade per element, never abort, mint no diagnostic code. The precedent is
eleven lines above the `:456` site — `page_setup.go:428-435`'s `fontChain` failure path already sets
`TextPaint = &CanvasTextPaint{Lines: []}` and continues. Two guardrails carry the weight: **truncate
the paint, never the value** (the properties panel writes the value back, so truncating it destroys the
author's text), and **the degraded state must differ from the empty one** (today a 400-line element and
an empty element both project `Lines: []`) — add a field beside `Overflow bool` and paint the first N
lines rather than none. Raising the constant was rejected on the criterion: the cliff is the defect, its
position is not.

**The defect.** `folio-go/page_setup.go:27` declares `const maxCanvasTextLines = 256`, enforced at
`:456` inside `addCanvasTextPaint`, which projects **every** text element in **every** band. The
guard `return`s an error rather than clamping, and that error is the function's own return, so one
oversized element does not lose its own paint — it aborts the entire canvas projection. The designer
gets no canvas at all, for a document that renders to a perfectly good PDF.

**Why Story 7.1 is the story that has to record it.** Before 7.1 a canvas text element's line count
was bounded by *wrapping*: a declared width produced as many lines as the text needed, and an element
with no declared width produced exactly **one**. Neither path could reach 256 from an element's own
value. Typed breaks now set the line count **directly** — 256 line feeds in a value is 257 lines
whatever the box is, and on the no-declared-width path the breaks are the *only* delimiter. So the
guard moved from unreachable-by-construction to reachable by pasting a long clause.

**Why it is medium and not low.** Three things compound:

1. **The blast radius is the whole projection, not the element.** Every other canvas bound in this
   file — `maxCanvasPropertyString`, `maxCanvasTextFragments` — sits beside a projection that either
   degrades or is bounded by construction. This one is the only reachable hard abort of the canvas
   BOUNDS. **Amended 2026-08-30 by Story 7.2:** it is no longer the only hard abort reachable through
   `addCanvasTextPaint`. That function now derives the leading model per element and returns its
   error, so an authored, LOAD-LEGAL `style.lineSpacing` that resolves to a zero advance or an
   overflowing product also aborts the whole projection — a second route in, from the format side
   rather than the bounds side. This does not widen DW-25's own remedy (7.2's contract forbids it any
   designer-surface work), but whoever lifts these bounds under Story 7.4 must know the function has
   two classes of caller-visible failure to degrade, not one.
2. **It is directly in Story 7.4's path.** 7.4 is "author body text in the designer", whose whole
   point is typing and pasting multi-paragraph clause text into exactly this element on exactly this
   surface. A pasted contract clause is not an adversarial input.
3. **The failure is opaque where it lands.** The message names the element and "the line projection
   bound"; what the author sees is that the canvas stopped working after a paste.

**What it is not.** It is not a correctness or byte-identity defect: the PDF path has no such bound
and renders the document correctly. It is a designer-surface availability defect.

**The shapes the lead is being asked to choose between** (recorded so the ruling has options, not to
pre-empt it): clamp the projection at the bound and report the truncation through Story 4.6's
existing clip-and-warn vocabulary rather than minting a code; keep the error but scope it to the
element so the rest of the canvas still projects; or raise the bound and re-derive it from the
`Number.MAX_SAFE_INTEGER` argument the other canvas bounds are drawn from. Each is designer-surface
work, which is why 7.1 could take none of them.

**How we'd know it was still wrong.** Story 7.4 lands, someone pastes a real multi-paragraph clause
into a canvas text element, and the canvas goes blank instead of showing the clause.

**SCOPE AMENDED 2026-08-30 by Story 7.3, which discovered the interaction and measured it rather
than guessing.** The amendment lands here, in the story that found it, on D-7.1.4's own precedent: a
deferral known to be wrong is worse than one merely open.

Story 7.3 makes a justified line project as one fragment **per word-piece** instead of one per face
segment, which multiplies a component's cumulative fragment count by words-per-line. That matters
here because the two fragment caps are enforced against **different quantities**:
`folio-go/page_setup.go:28`'s `maxCanvasTextFragments = 512` is **per line**, while
`folio-designer/src/engine-protocol.ts`'s `fragments <= 512` is **cumulative across the whole
component** — and a validator failure there drops the entire response. The cumulative one is
therefore the cliff a long justified paragraph would reach first, and the reachability was measured
before anything was concluded from it:

| document | value bytes | lines | cumulative fragments | max per line |
|---|---|---|---|---|
| realistic clause prose, 200 pt column, 9 pt | 380 | 9 | **64** | 9 |
| the same prose, 523 pt full content width, 9 pt | 380 | 4 | **68** | 23 |
| **adversarial**: 256 one-letter words (511 B), 40 pt column, 6 pt | 511 | 32 | **249** | 8 |
| **adversarial**: the same value in a 200 pt column | 511 | 7 | **241** | 40 |

**The cliff does NOT open, and the reason is a bound that was already here.** A justified component's
total fragment count is `Σ(pieces per line) = (interior gaps) + (lines) ≤ (break opportunities) + 1`,
and the component's **value** is itself capped at `maxCanvasPropertyString` = 512 **bytes** — the
very cap this entry's amended title already names as the one that binds first. The largest number of
break opportunities 512 bytes of text can carry is ~255 (alternating one-letter words), so the
cumulative fragment count cannot exceed ~257: **half the browser-side cap of 512**, measured worst
case 249. The per-line Go cap of 512 is even further away; the widest measured line held 40.

So **no bound needed widening for justification**, and Story 7.3 widened none — its HALT condition
"canvas projection bound must widen" was not triggered. What changes for Story 7.4 is the ORDER OF
CONSEQUENCES once the 512-byte value cap is split as D-7.4.2 rules: today the value cap makes the
fragment cap unreachable. **The moment 7.4 lets a component carry more than 512 bytes of body text,
the cumulative fragment cap becomes reachable for the first time — and it becomes reachable through
`justify` long before it would through any other alignment**, because a ragged component projects
roughly one fragment per line (≤ 256) while a justified one projects one per word. At the realistic
rate measured above (~7 fragments per line of a 200 pt clause column), the browser-side cumulative
cap of 512 is hit at about **73 justified lines** — roughly one and a half pages of a clause column,
which is not an adversarial input. Whoever lifts the value cap must lift or re-derive the cumulative
fragment cap in the same change, or 7.4 will trade a blank canvas at 512 bytes for a blank canvas at
73 justified lines. Owner is unchanged: **Story 7.4**.

---

## DW-25 IS CLOSED — Story 7.4, 2026-08-30

Closed against a **grep re-run at the closing revision**, not against the hand-list above: this
entry's own enumeration was both stale and incomplete twice over, which is exactly how DW-24's first
closure went wrong.

### The enumeration, re-derived. `grep -n 'maxCanvasPropertyString\|maxCanvasBodyText' folio-go/page_setup.go`

**Re-run at the story's CLOSING revision.** The block first written here was produced mid-dispatch and
its anchors had already rotted by roughly twenty-two lines before the commit landed — the review pass
that corrected two constant comments grew the const block underneath them. The site IDENTITIES were
right; only the numbers drifted. That is this entry's own lesson biting inside its own closing note,
so the block below is the one taken at the revision that closes it:

```
38:const maxCanvasPropertyString = 512
70:const maxCanvasBodyText = 1048576
310:  if len(name) > maxCanvasPropertyString {            # font-family name
645:  if len(fragment.text) > maxCanvasBodyText {         # BODY TEXT (was 512)
748:  if len(element.Value.Value) > maxCanvasBodyText {   # BODY TEXT (was 512)
757:  if len(element.VisibleIf.Value) > maxCanvasPropertyString {
763:  if len(element.Table.Value.Bind) > maxCanvasPropertyString {
808:  if len(style.FontFamily.Value) > maxCanvasPropertyString {
840:  if len(style.Color.Value) > maxCanvasPropertyString {
846:  if len(style.Background.Value) > maxCanvasPropertyString {
861:  if len(border.Color.Value) > maxCanvasPropertyString {
```

**NINE sites, two of them body text, exactly as the plan gate re-derived.** The two body-text sites
moved together — a value large enough to pass the first would otherwise have aborted at the second —
and the **seven identifier, colour and expression sites keep 512 and keep aborting**. That residue is
**recorded, not fixed** (D-7.4.2 §6): Epic 7 makes none of them newly reachable, and
`TestCanvasIdentifierBoundsStillRefuseAtFiveHundredAndTwelve` (`folio-go/canvas_body_text_bounds_test.go`)
is where "recorded" is executable — it asserts **all seven** still refuse at 513 bytes, with the
identifier-bound refusal's own message rather than merely a non-nil error, **and** that a 513-byte
clause no longer does.

**One half of that last claim was unpinned until the close, and is now pinned.** The two body-text
sites do not fail the same way any more: the VALUE site still aborts, so `err == nil` is a complete
assertion for it — but the FRAGMENT site now **degrades**, so repointing it back at the identifier
bound raises no error at all. It quietly paints a truncation notice over a clause well inside the
documented bound, and an error-only check stays green straight through it. Measured with that revert
applied at the closing revision: `Truncated=true, len(Lines)==0`. The test now asserts the fragment
half on its own terms — one line, one fragment, all 513 bytes, `Truncated` false — and that assertion
was red-proved by the revert before this entry was closed.

### What the fix actually is

`maxCanvasPropertyString` was **split at the declaration**, not raised. Three new constants, each
carrying its criterion and its arithmetic in its own doc comment:

| Constant | Value | Derivation |
|---|---|---|
| `maxCanvasBodyTextLines` | 1920 | 40 pages × ⌊729890 mp content-band height ÷ 14982 mp advance at 11pt on the shipped chain⌋ = 40 × 48 |
| `maxCanvasBodyTextFragments` | 65536 | cumulative per element; the same forty-page document justified at full A4 width, **18.05 fragments/line measured** (see the re-measurement below) → 1920 × 18.05 = 34 656, to the next power of two |
| `maxCanvasBodyText` | 1048576 bytes | channel-representability backstop, a full power of two above 1920 lines × ~90 chars × 3 bytes = 518 400 bytes, so it cannot bind before the paint bounds |

The two **paint** bounds now **degrade per element**: the element paints its first N lines, sets a new
`CanvasTextPaint.Truncated` flag beside `Overflow`, and the projection continues — reusing verbatim
the `fontChain` disposition eleven lines above the site, not a second one. `maxCanvasTextLines = 256`
is **gone**, not renamed.

**`maxCanvasBodyText` is the one site that still refuses, and that is deliberate.** Degradation lives
on the paint side alone (D-7.4.2 §1): `component.Value` is what the properties panel edits and
**saves**, so truncating it would write the truncation into the author's document. It is a
channel-representability backstop at megabyte scale that Epic 7's input cannot reach, and its comment
says so — following D-7.2.3's precedent for a stated sanity ceiling.

### The fourth mirror, which is what made the Go-side fix observable

`engine-protocol.ts:152-154`'s `optionalString` capped an element's `value` at 512 alongside seven
identifier keys — this entry's own two-jobs conflation, reproduced exactly on the browser side.
Splitting Go without splitting it would have changed **nothing observable**: the browser would have
gone on dropping the whole response. All four mirrors are now hoisted to named constants
(`MAX_CANVAS_BODY_TEXT`, `MAX_CANVAS_BODY_TEXT_LINES`, `MAX_CANVAS_BODY_TEXT_FRAGMENTS`,
`MAX_CANVAS_PROPERTY_STRING`) and tied to the Go declarations by
`folio-designer/src/engine-bounds-mirror.test.ts`, which reads **both files**, asserts all four pairs
non-vacuously, asserts each constant is consumed at the validator site it bounds, and red-proofs a
one-sided edit in both directions.

**The unit mismatch is recorded, not "fixed".** Go counts BYTES (`len()`); TypeScript counts UTF-16
CODE UNITS (`.length`). For non-ASCII the browser is the more permissive of the pair, so Go refuses
first and nothing unrepresentable crosses. The tie compares **literals, not quantities**, and its own
comment says so.

### The re-measurement, at the closing revision, with the value cap lifted

**Method.** Justified English contract prose, 11pt, the **shipped `["Noto Sans"]` chain** — the same
face and size `maxCanvasBodyTextLines` is derived from, so the two bounds are measured against one
document — projected through `CanvasWithTextPaint` and counted off the returned
`CanvasTextPaint.Lines`:

| column | words | lines | cumulative fragments | max per line | fragments/line |
|---|---|---|---|---|---|
| 523.276 pt (full A4 content width) | 912 | 51 | 911 | 21 | 17.86 |
| 523.276 pt | 1824 | 101 | 1823 | 21 | **18.05** |
| 240 pt (clause column) | 912 | 113 | 910 | 10 | 8.05 |
| 240 pt | 1824 | 225 | 1822 | 10 | 8.10 |
| 523.276 pt, SHORT-WORD worst case | 6400 | 207 | 6387 | 31 | 30.86 |

The figure rises with sample length and settles at the words-per-line asymptote: a justified block's
LAST line is drawn at its natural edge, so a short sample's average is dragged down by it. **18.05
over 101 lines is the figure the constant is derived from**, and it supersedes two earlier ones that
were not measured this way: the original briefing's **16.72** was a thirteen-line sample, and this
entry's own first re-measurement of **19.35** used the `Roboto-Regular` TEST face rather than the
shipped chain. `folio-go/page_setup.go` and `epic-7-8-decision-log.md` carry 18.05 too.

**The geometry-free law holds and is the thing worth recording:** cumulative fragments ≈ the value's
**word count**, at any column width (911/912, 1823/1824 — the last line contributes one).
Consequences against the new bounds:

- 1920 × 18.05 = **34 656**, which is ABOVE 32 768: the bound this entry first shipped did **not**
  cover its own forty-page criterion, and binding at ~1 815 lines it fell short by a page and a half.
  Raised to the next power of two, **65 536**.
- 65 536 also clears the **short-word worst case**: 1920 × 30.86 = 59 251. The criterion therefore
  holds for prose denser than a contract's, not only for the corpus it was measured on.
- At a **240 pt clause column** the LINE bound binds first: 1 920 lines × 8.1 ≈ 15 600 fragments,
  under a quarter of the cumulative allowance.
- The **peak a single element can now emit** is `min(65 536, 1920 × fragments-per-line)` by
  construction — the old figure of 256 (this entry's ~249 was itself an undercount) is retired along
  with the 512-byte value cap that produced it.

**Correction to this entry's own SCOPE AMENDMENT.** The "~73 justified lines" crossing point was a
240 pt-column figure quoted as a general one; at full A4 content width the browser's old cumulative
512 was crossed at ~28 lines (measured 18.05 fragments/line), not 73. The general statement is the
word-count law above, not any lines figure.

### One residue found while closing, noted and NOT fixed

`CanvasTextPaint.Overflow` sets the CSS class `canvas-component-text-overflow` and **there is no rule
for that class anywhere in `App.css`** — the existing degradation flag has always been invisible to
the author. Story 7.4 did not fix it (out of contract) but deliberately did not repeat it: the new
`Truncated` flag states its reason **in words** at the component and in the same sentence the
component's accessible name carries. Whoever owns the overflow presentation next has a precedent to
copy and a bug to close.

### Verification at closure

- `folio-go/canvas_body_text_bounds_test.go` — degradation past the line bound (prefix painted, flag
  set, **every other component projects normally**, the document's own value untouched); truncated vs
  empty distinguishable in one projection; the per-line/cumulative asymmetry; degradation past the
  per-line guard through a real projection; `Paginate`'s page count independent of paint truncation.
- `folio-designer/src/canvas-authority-contract.test.ts` — a new prohibited pattern banning any
  height or window derivation from `textPaint…lines.length`, with its own red-proof.
- All twenty recorded golden digests **measured** unchanged; four-target matrix legs run individually
  and the cross-target gate passed.

---

### DW-27 — `fixtures/justified-thai/` is the first golden anywhere to insert visible space between Thai words, and no Thai reader has looked at it

- **Deferred by:** Story 7.3 (2026-08-30), at story close. The story's own contract could not
  discharge it: **no agent may write a `reader`/`date`/`examined` sign-off record**, because doing so
  would be fabricating an attestation about a human reading that did not happen.
- **Owner:** **the project owner**, who must commission a Thai reader. Named as `Owner` in this
  file's own vocabulary — needs the project owner's call before anyone acts. The orchestrator has
  surfaced it to the owner; this entry is the standing address so it does not depend on that message
  being remembered.
- **Severity:** MEDIUM.
- **Status:** OPEN.

**What is missing.** `fixtures/justified-thai/` justifies **432 Thai characters** by inserting gaps
of up to **3,528 millipoints** between Thai words. Thai normally writes those word boundaries with no
space at all. Every machine-checkable property of that page is pinned — the right edge lands on the
declared width exactly, the pieces concatenate back to the control element's line, the digest is
identical on all four targets — and **not one of those properties answers whether the page reads
correctly to someone who reads Thai.** That is the irreducibly-human half D-2.3.5 exists for.

**Why no existing record binds.** D-000.26 binds a sign-off to the artifact expressing the property
judged. `fixtures/shaped-text/thai-signoff.json` (D-2.3.5) judges Thai **mark placement**;
`fixtures/expected-breaks/break-signoff.json` (D-2.4.3) judges that **every marked seam falls between
words and never inside one**. Neither judges inter-word *spacing*, which is this fixture's whole
subject. Story 4.7 set the precedent of creating a third record rather than stretching an existing
one.

**And the precedent is NOT "every Thai-bearing golden gets a sign-off."** Measured at this revision,
`fixtures/wrapped-text/` carries 47 Thai characters with no record and `multi-script-fallback` one.
Those fixtures only BREAK Thai at seams `expected-breaks` has already signed off. This one is the
first to put visible space *into* the run.

**How we'd know it was still wrong.** A Thai reader opens `fixtures/justified-thai/expected.pdf` and
says the spacing reads as broken words rather than as justified text. No test in the tree can reach
that verdict, and none attempts to.

**What discharges it.** A sign-off record beside the fixture, following the two existing ones in
shape, written by or on behalf of a human who read the page — enforced by a `//go:build matrix` red
gate as both precedents are. **Not** an agent-authored record.

---

### DW-28 — a large class of ordinary Thai cannot be rendered at all: any sequence stacking two marks over a base fails closed, and the refusal itself has no test

- **Deferred by:** Story 7.3 (2026-08-30). Pre-existing and **not justification's doing** — measured
  to fail identically under `align: left` — so it was out of 7.3's scope, but 7.3 is where it was
  found, and Epic 7 targets Thai legal documents.
- **Owner:** **the next story that touches `internal/pdf`'s glyph-positioning refusal**, plus the
  Epic 7 and Epic 8 plan-gate checklists as a second standing address (D-000.73: not "Epic 7 close",
  which stops existing when the event passes).
- **Severity:** MEDIUM for the limit; LOW for the untested refusal. Recorded as one entry because the
  second is why the first could regress in either direction unnoticed.
- **Status:** OPEN.

**The limit.** Every Thai codepoint renders in isolation — measured **91/91** over U+0E01..U+0E5B.
But **any sequence stacking two marks over a base** fails closed with
`face Noto Sans Thai: CID N carries a non-zero vertical offset (-57), which a TJ array cannot
express`. This is a hard `Render` **error**, not a diagnostic: the document does not render at all.

**These are not exotic strings.** `ครั้ง` ("time/occasion"), `ทั้งนี้` ("in this regard" — a stock
phrase in Thai legal prose) and `ตั้งแต่` ("since/from") are all unrenderable. A bisect over a
natural Thai sentence isolated the four-rune window `าทั้`.

**Re-measured independently at story close (2026-08-30), not relayed.** Rendering
`ทั้งนี้ ครั้ง ตั้งแต่ คู่สัญญา` through the shipped `folio render` CLI fails with
`internal/pdf: face Noto Sans Thai: CID 9 carries a non-zero vertical offset (-57)…` and exit 1
under `align: "justify"` **and, byte-for-byte the same error, with no `align` declared at all**. The
shipped `fixtures/justified-thai/input.folio` renders exit 0 in the same run as the control. **So the
limit is confirmed independent of justification and is not Story 7.3's doing.**

**It is a deliberate, correct refusal** — `internal/pdf`'s AC6 fail-closed branch raised at
`folio-go/internal/pdf/textdoc.go:914`, typed at `:1007-1019` — and refusing is better than drawing
the marks wrongly.
The problem is not the branch; it is that **a document author cannot be expected to know that mark
stacking is what broke their document**, and that the fixture text for any Thai document in this
repository has to be chosen around it (which is exactly what Story 7.3 had to do).

**And the refusal is unguarded.** Grepping `verticalOffsetError|non-zero vertical offset` across
`*.go` returns **only five hits, all inside `textdoc.go` itself** (`:914` the raise site, `:1007`
`:1009` the type and its doc, `:1015` `:1017` its `Error()` method). **Nothing anywhere constructs a document that reaches it, and nothing pins the
message.** So the branch that refuses a large class of ordinary Thai could start refusing more, or
stop refusing and draw the marks wrongly, and no test would notice either direction.

**What discharges it.** At minimum, a test that reaches `verticalOffsetError` and pins its message.
Beyond that, an owner call on whether a hard render error is the right outcome for ordinary Thai
words, or whether this should degrade with a located diagnostic.

---

### DW-29 — `style.align: "justify"` on a table or its `headerStyle` loads, forces the document to 2.0, and renders every cell at the start edge with no diagnostic — ROUTED TO STORY 7.8

- **Deferred by:** Story 7.3 (2026-08-30). The diff is **contract-correct**: 7.3's intent contract
  explicitly directed that `headerStyle.align: "justify"` load and raise the document to 2.0, while
  its Never list forbade implementing justified table cells. The residue is the product question the
  contract deliberately did not settle.
- **Owner:** **Story 7.8**, a named story now written into `epics.md`. Originally Story 7.4, as an
  explicit acceptance criterion — the engineering lead ruled (2026-08-30) that the value must be
  **rejected at load**. **Story 7.4's plan gate judged the addition `multiple-goals` and exercised
  this entry's own escalation clause**, which provides for exactly that outcome. **This is not a
  further deferral**: the ruling stands unchanged, and the work now has a numbered home with its
  inheritances written down.
- **Severity:** MEDIUM.
- **Status:** OPEN, ruled, **owned by Story 7.8**.

**AMENDED 2026-08-30 at Story 7.4's close — what 7.4 discharged and what it did not.**

7.4 discharged the **product** half and nothing else: the inspector's align control offers `justify`
only when every selected component is text, and offers three segments for a table or a mixed
text+table selection (`folio-designer/src/App.tsx`, pinned by *"offers justify for text alone, and
never for a table or a mixed selection"* in `App.test.tsx`). So the designer can no longer author the
defective document.

The **format** half — the load-time refusal — is untouched, and 7.4's Never list forbade it. The
reachability argument that put DW-25 into 7.4 runs the other way here: 7.4 does **not** make DW-29's
condition newly reachable, precisely because the panel no longer offers the value. After 7.4 the
defective document is reachable only by hand-editing a `.folio`, exactly as before. Three further
reasons are recorded in `7-4-author-body-text-in-the-designer.md`'s Design Notes, the load-bearing one
being that a located error naming the element and the field **cannot reach a designer author without
a THIRD per-field style diagnostic code**, which `internal/diag/diag.go:249-252` explicitly reserves
for a deliberate decision. That is a lead call, not a builder's.

**The behaviour today.** A table element's `style.align: "justify"`, or its `headerStyle.align:
"justify"`, loads without error and raises the document to format **2.0** — making it unreadable to
every 1.x reader — and then renders **identically to `align: left`**, with no diagnostic. The author
pays the whole cost of the MAJOR and receives nothing for it. Story 7.3 pinned that fallback with
`TestTableCellsCascadedJustifyIsDrawnAtTheStartEdge` and corrected the three cell-align `default:`
arm comments that had wrongly claimed the load-time check already rejected such a value.

**The lead's own root-cause note, recorded because it is the reusable half.** D-7.3.1's guardrail
split the alignment closed set **by JSON key location** (`style`/`headerStyle` versus `columns[]`)
rather than **by consumer**. Those are different partitions. A table's `style.align` and
`headerStyle.align` are read into `r.alignFallback` (`folio-go/table_render.go:373-376`, `:440-441`)
and consumed at **the same site** as `columns[].align` — so the guardrail that was supposed to make
justified table cells impossible by construction let the value in through the other door. **When
splitting a closed set, partition it by the code that consumes the value, not by where the value is
written in the document.**

**Record correction this entry carries.** 7.3's intent-contract paragraph directing that
`headerStyle.align: "justify"` load and raise to 2.0 is **superseded** by the lead's ruling and has
been annotated as such in `7-3-justify-a-paragraph-s-edges.md`. The shipped behaviour matches the
contract as written at the time; the direction has since changed.

---

### DW-30 — AC-TH2's wording in the owner's scope amendment is wrong on the facts, and a reader checking it the obvious way concludes the test is broken

- **Deferred by:** Story 7.3 (2026-08-30). The build dispatch reported it rather than editing it,
  correctly: **the amendment text is the owner's**, not the workflow's to rewrite.
- **Owner:** **the project owner** (the amendment's author), for a one-sentence correction. No code
  is owed — the delivered behaviour is right.
- **Severity:** LOW. Documentation only, but the failure mode is a future reader concluding a correct
  test is broken.
- **Status:** OPEN.

**What is wrong.** AC-TH2 directs the implementer to "pick a Thai run **the dictionary does not
cover**". `กานต์` — a suffix of the fixture's `ณัฐกานต์` — **is** in the shipped dictionary, at
`folio-go/internal/text/wordlist/words_th.txt:3084`, and the greedy matcher **does** propose a break
at rune 3.

**Why the test is nevertheless right.** The opportunity is withdrawn by D-2.1.9's both-sides-coverable
filter (`folio-go/internal/text/tileable.go`), because the preceding `ณัฐ` cannot be tiled by
dictionary entries at all. The AC-TH2 test never relied on the wrong reason: its precondition is
**computed from production `text.Opportunities(text.Dictionary(), …)`** and asserted with a
`t.Fatalf`, not read off a literal or a wordlist grep. Verified at story close: the precondition is
live and non-vacuous.

**The corrected wording, for whoever amends it:**

> pick a Thai run **the segmenter proposes no interior opportunity inside**

**Why it matters.** "The dictionary does not cover it" and "the segmenter proposes no interior
opportunity inside it" are different claims, and only the second is the precondition the test
asserts. A reader checking AC-TH2 the obvious way — grepping `words_th.txt` — finds the word present
and concludes the fixture is mis-chosen and the test broken. It is neither.

### DW-31 — `render.go`'s image-centring rounding is the one `ScaleRound` halving DW-24 deliberately left uncovered
- **Deferred by:** Story 7.3's close (2026-08-30), which measured it rather than absorbing it
- **Owner:** the next story that touches image placement, **or** Epic 8's plan gate (whichever first) —
  a role and a gate, never an event, per D-000.73
- **Severity:** LOW — one site, no author-facing control reaches it today
- **Status:** OPEN

**The gap.** DW-24 closed over the eight rounding sites that alignment reaches. Story 7.3's closure
mutation-tested a ninth, `render.go:505`'s image-centring `geom.ScaleRound`, and measured it **GREEN** —
i.e. no golden in the corpus exercises it, so a change to its rounding mode would move no recorded byte
and nothing would notice. It was correctly recorded **out-of-subject** for DW-24, whose subject is the
text/valign alignment rounding; this entry exists so "out of subject" does not quietly become "covered".

**Why it is low and not medium.** The image-centring path is reached only by an image whose declared box
exceeds its intrinsic size, and no author-facing control in the designer produces that today. It becomes
reachable the moment image fit/placement is authorable.

**What closing it requires.** A fixture whose image centres with a slack **≡ 3 (mod 4)** in millipoints —
per Story 7.3's generalised lesson, an odd slack is a coin flip because round-half-to-even and truncation
agree on `slack ≡ 1 (mod 4)`; only `≡ 3 (mod 4)` discriminates. Registered at every golden surface, with
the enumeration **re-derived by grep at closure**, never read off a hand-list: DW-24's anchors rotted
three times, the third time *inside the commit that closed it*.

---

### DW-32 — the property-command encoder splices the author's typed text into the command JSON unquoted, so a non-numeric entry produces malformed bytes instead of a located engine error
- **Deferred by:** Story 7.4's review pass (2026-08-30); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** the next story that adds a NUMERIC property control, **or** Epic 8's plan gate (whichever
  first) — a role and a gate, never an event, per D-000.73
- **Severity:** MEDIUM — the author sees a generic refusal instead of the field-located message the
  panel is built around, but no bad value reaches the document
- **Status:** OPEN

**The gap.** `folio-designer/src/component-property-command.ts` routes `pointFields` and Story 7.4's new
`ratioFields` through `rawNumberLiteral`, which returns the typed string **verbatim**. Typing `abc` into
line spacing emits `{"op":"set","value":abc}` — bytes that fail JSON parsing on the engine side and yield
a generic refusal rather than the located `STYLE_LINE_SPACING_INVALID` message the panel is built to show.

**Why it is deferred and not fixed here.** The pattern is **pre-existing** for the point fields and the
spec directed Story 7.4 to follow it, so the story did not cause it. What the story did do is **widen the
set of fields on that path** by one, which is why it is filed rather than merely noted.

**What closing it requires.** A refusal on the encoder side before the command is built — the panel
already has the field identity it needs to locate the message — rather than letting unparseable bytes
travel and be diagnosed by their absence of structure.

---

### DW-33 — a text element whose FIRST packed line already exceeds the per-line fragment guard paints ZERO lines, and its only signal is the truncation notice
- **Deferred by:** Story 7.4's review pass (2026-08-30); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** Epic 7's retrospective, **or** the plan gate of the next story that changes canvas paint
  degradation (whichever first) — a gate, never an event, per D-000.73
- **Severity:** MEDIUM — conformant with the contract as written, but it is the degradation path that
  degrades furthest and the author is given the least to go on
- **Status:** OPEN

**The gap.** Measured during Story 7.4's review: when the very first packed line already carries more than
`maxCanvasTextFragments` fragments, the paint loop breaks before appending anything, and the element
projects `Truncated=true, len(Lines)==0`. The author sees an empty box carrying the truncation notice.

**Why production behaviour was deliberately NOT changed.** This is **conformant** with the contract Story
7.4 was written against: painting "stops at the last whole line that fits", and here no whole line fits.
The degraded state also remains distinguishable from the empty one — `Truncated` is the flag that
distinguishes them, and the notice states the reason in words — which is the invariant the contract
actually protects. Story 7.4's own review made the vacuity visible rather than papering over it:
`assertWithinBrowserFragmentBounds` now requires a caller-stated expected line count, so the zero is
**explained** at its call site rather than certified by accident.

**What closing it requires.** A ruling, not a patch. Whether such a line should paint a **partial
prefix** is a design question the contract does not settle, and it trades one honesty against another:
a partial line shows the author a sentence the document does not contain. Whoever closes this decides
that first.

---

### DW-34 — the canvas renders one DOM span per fragment for every projected line, with no virtualisation, and Story 7.4's bounds made the reachable ceiling two orders of magnitude higher
- **Deferred by:** Story 7.4's review pass (2026-08-30); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** Story 7.6's plan gate (the canvas draws every page a document produces) — a gate, never an
  event, per D-000.73
- **Severity:** LOW — not a correctness defect and not newly introduced; only the scale at which it can
  be reached is new
- **Status:** OPEN

**The gap.** `folio-designer/src/App.tsx`'s `textPaint` painting path maps every projected line and every
fragment within it unconditionally to a DOM node. Story 7.4 raised the projectable ceiling from 256 lines
/ 512 fragments to **1920 lines / 65536 fragments**, so a document at the new bounds can build tens of
thousands of spans for a single component.

**Why it is low.** Nothing about correctness changes, and the pre-existing code is unchanged. Forty pages
of clause text is also not the common case. It is filed because the ceiling moved, not because the code
did.

**Why Story 7.6 is the owner.** 7.6 draws every page rather than one, which multiplies the same
unvirtualised path by the page count. The two questions are the same question and should be answered once.

---

### DW-35 — the canvas hard-codes ONE font stack regardless of the document's own `fonts` map, so a template naming a different chain still paints with these three families
- **Deferred by:** Story 7.4's close (2026-08-30), observed while fixing the owner-reported Thai canvas
  defect at `c6e4d03` and recorded there in the commit message
- **Owner:** **Epic 8's plan gate** — a gate, never an event, per D-000.73. This is exactly the class
  D-7.4.4 belongs to: a limit to STATE, not to fix, until the product makes it reachable
- **Severity:** MEDIUM once Epic 8 lands; **latent today**
- **Status:** OPEN

**The gap.** `folio-designer/src/App.css`'s `.canvas-text-fragment` rule names a **fixed** three-family
stack — the three faces `scripts/build-wasm.mjs` ships and declares an `@font-face` for. The engine, by
contrast, measures with the chain the **document's own `fonts` map** names. When those two disagree, the
browser rasterizes at the engine's x-positions with the wrong metrics, and fragments collide.

**This is the same mechanism as the defect fixed at `c6e4d03`**, one level up. That fix corrected a **name
mismatch** — `App.css` asked for the faces under Noto names that no `@font-face` declares, so the browser
fell through to generic `sans-serif` and Thai overlapped at every space. `canvas-font-stack.test.ts` now
ties the two tracked sources so a name can never drift again. But the tie pins the stack to the
**generator's** three families, not to the **document's** chain, and nothing checks the latter.

**Why it is latent today.** Every document the product can currently produce resolves to the shipped
chain, so the hard-coded stack happens to be right. **Epic 8 makes chains editable and author-chosen**
(8-1 through 8-5), which turns "happens to be right" into "wrong for any document that says otherwise" —
and the symptom is overlapping glyphs, which reads as a rendering bug rather than a font-resolution one.

**What closing it requires.** The canvas fragment rule must derive its family list from the projected
component's own resolved chain rather than from a stylesheet constant, and the `canvas-font-stack.test.ts`
tie must be widened from "every family the rule asks for is declared" to "the families the rule asks for
are the ones the engine measured with". Note the ORDER constraint that fix already discovered: the stack
must follow the engine's chain order rather than borrow `tokens.css`'s `--font-page`, which puts Thai
first and would hand Latin text Noto Sans Thai's Latin glyphs.
