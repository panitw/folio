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

### DW-28 — a large class of ordinary Thai cannot be rendered at all: any glyph the shaper gives a non-zero vertical offset fails closed, and a code comment said the branch was unreachable

- **Deferred by:** Story 7.3 (2026-08-30). Pre-existing and **not justification's doing** — measured
  to fail identically under `align: left` — so it was out of 7.3's scope, but 7.3 is where it was
  found, and Epic 7 targets Thai legal documents.
- **Owner:** **Story 8.0** — *a stacked Thai mark reaches the page*, the opening story of Epic 8,
  named 2026-08-31. **The previous owner clause failed and is recorded here as evidence, not
  deleted:** it read *"the next story that touches `internal/pdf`'s glyph-positioning refusal, plus
  the Epic 7 and Epic 8 plan-gate checklists as a second standing address"*, and **Epic 7 ran eight
  plan gates without one of them picking it up.** That is the third checklist-as-owner failure in
  this run. **Do not re-file this as a checklist item** — a named story with a position in the
  sequence is the only owner that has worked on this project.
- **Severity:** **HIGH** — raised from MEDIUM 2026-08-31 by the engineering lead, on the stated
  criterion *blocks a supported use case, with no workaround, for a real user*. The shipped Thai
  face is the only Thai face; the document is the owner's real work; and "avoid the character
  sequence" is mutilating the document, not a workaround. **Explicitly NOT grounded on "the product
  lies to the author"** — AD-5 makes the page model blind to the emission stage, so the canvas
  drawing the text correctly is an invariant working as designed, not a defect. The severity comes
  from the outcome: no bytes at all.
- **Status:** **CLOSED 2026-08-31 by Story 8.0** (build `26e3ba1`). The characterization half had
  closed first (`folio-go/thai_mark_stacking_test.go`, landed separately at the owner's direction so
  the branch could not drift while the fix waited); the fix half shipped with it. Measured at close:
  the owner's clause renders exit 0 through the shipped CLI, its bytes are identical on all four
  AD-21 targets, and all 21 pre-existing goldens are unmoved.
  **THREE SUCCESSOR OBLIGATIONS WERE CREATED BY THE FIX AND ARE FILED SEPARATELY — this entry is
  closed, they are not:** **DW-56** (HIGH — the new golden's human reading sign-off, owner: the human
  reader), **DW-57** and **DW-60** (the two remaining sites of the falsified unreachability claim
  below). Closing this entry is a statement about *ordinary Thai reaching the page*, which is
  measured, and not about those three.

- **PREDICATE CORRECTED 2026-08-31 at Story 8.0's plan gate, in this entry's own heading.** It read
  *"any sequence stacking two marks over a base fails closed"*. Measured wrong, and over-broad: `ที่`,
  `ป้ำ` and `ปั` each stack two marks over one base and rendered exit 0 throughout, because the
  shipped face resolves that case by a GSUB lowered-form substitution **at zero offset**, and only
  the `ั`+tone case by a GPOS y-displacement. `ที่` appears in `fixtures/shaped-text`, in all four
  `statement-*` fixtures and in `justified-thai`. **The trigger is a non-zero `YOffset`, full stop** —
  an implementer taking the old predicate at face value would have built the wrong test and then read
  the shipped goldens as contradicting the story. Corrected in `epics.md` at the same gate.

**FOUND IN PRODUCTION, 2026-08-31.** The owner pasted a contractor-liability clause from a real Thai
contract into a text element. The canvas rendered it; the PDF preview failed with
`internal/pdf: face Noto Sans Thai: CID 27 carries a non-zero vertical offset (-2)…`, surfaced as
`Render failure · ENGINE_REJECTED`. Their clause ends `...รับผิดเป็นการส่วนตัวทั้งสิ้น`. Reproduced
through the shipped CLI at `31d6cc6`: the full clause fails with **CID 27, offset −2**; the single
word `ทั้งสิ้น` fails with **CID 3, offset −57**; the control `สัญญา` renders exit 0.

**TWO CORRECTIONS TO THIS ENTRY'S OWN TEXT (2026-08-31), because it overstated in one place and
understated in another.**

**(a) It is not untested in both directions.** This entry said the branch *"could start refusing
more, or stop refusing and draw the marks wrongly, and no test would notice either direction."* Too
strong: `internal/pdf/textdoc_test.go`'s `TestShapedRunFailsClosedOnYOffset` — **re-pointed by Story
8.0 to `TestShapedRunExpressesAYOffsetAsATextRise`, with the fail-closed half kept as
`TestShapedRunFailsClosedOnARiseThatRoundsAway`** (D-7.8.7: re-point, never delete) — exercises it
with a
synthetic run **and carries a non-vacuity leg** — the same run with `YOffset` zeroed must emit
cleanly — so the *stop refusing* direction has always reddened. What was genuinely unpinned was the
**message** and **reachability through a real document**, and both are now pinned.

**(b) A comment in the code asserted this branch was UNREACHABLE, and that is why nobody looked.**
`textdoc.go` read, verbatim: *"Measured at Story 2.3: YOffset is 0 for every glyph of every sample
across all three shipped faces, so this branch is UNREACHABLE through the render path with the
shipped set and cannot be red-proved through it."* **Both halves are false.** Story 2.3 measured
**its own samples** and reported on **the shipped set** — two different populations, the same
measure-one-report-wider error this run has now recorded three times. The comment was load-bearing:
it justified both the fail-closed choice and the absence of a render-path test, and it is what a
reader hitting the refusal would have checked first. **It protected itself.** Corrected in
`textdoc.go` and in `textdoc_test.go`'s parallel claim, 2026-08-31.

**THE FIX IS AVAILABLE AND THE FORMAT ALLOWS IT.** PDF's text-rise operator `Ts` expresses a
vertical offset directly, and it is **inside AD-6's pinned profile** — AD-6's exclusion list
(encryption, annotations, forms, transparency groups, shading, ICC, tagging) does not contain the
text-state operators. The refusal's own comment concedes the gap: *"the alternative … is not built
here."* `grep` for `Ts` across `internal/pdf` returns nothing. So this is an unbuilt capability, not
a limit of the format.

**IT DOES NOT JOIN D-7.8.3's BEFORE-THE-TAG SET**, and the reasoning is worth keeping because it
runs opposite to Stories 7.8 and 7.10. Emitting `Ts` for glyphs that currently **refuse** can move
no existing golden **by construction**: a document containing such a glyph produces no bytes today,
so no fixture can contain one. The change **widens** what renders rather than narrowing it, so a
consumer upgrading past `folio-go/v0.1.0` gets more documents rendering, never fewer. The tag is not
the constraint — the product is. **The one byte-identity guardrail:** the `Ts` path must be entered
**only** when `YOffset != 0`, with the 21 digests asserting the zero-offset corpus is unmoved.

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

**What discharges it — both halves, and one is already done.**

**The characterization: CLOSED 2026-08-31**, at the owner's direction, deliberately ahead of and
separate from the fix so the branch could not drift while the fix waited.
`folio-go/thai_mark_stacking_test.go` reaches `verticalOffsetError` through `ParseTemplate` +
`Render` on the shipped font set, pins the message **verbatim** (a reworded refusal is a changed
product and should have to be edited deliberately), asserts a refused render emits **zero** bytes,
and carries a same-script control — `สัญญา`, same face, size and box, no stacked marks — so the
refusal arm is evidence about **stacking** rather than about Thai. A third arm uses the owner's own
clause and deliberately does **not** pin the CID, because a longer document subsets more glyphs and
which CID reports first is an artefact of the subset. **Mutation-proved:** neutering the
`YOffset != 0` branch makes the refusal arms red and the document render **3,187 bytes with the
marks silently misplaced** — the "healthy output and broken output are the same bytes" outcome the
branch exists to prevent, now demonstrated rather than asserted.

**The fix: Story 8.0's.** The owner call this entry asked for has been made — it does **not** degrade
with a diagnostic. Marks are placed at the offset the shaper computed, using `Ts`; whatever remains
genuinely unexpressible still refuses, with its message pinned as it is today. The fail-closed
branch **narrows**, it does not disappear.

---

### DW-29 — `style.align: "justify"` on a table or its `headerStyle` loads, forces the document to 2.0, and renders every cell at the start edge with no diagnostic — **CLOSED by Story 7.8, 2026-08-31**

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
- **Status:** **CLOSED** by Story 7.8 (`7-8-refuse-a-justified-table-at-load-in-the-author-s-own-terms.md`),
  2026-08-31. See *DW-29 IS CLOSED* below for what shipped.

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

## DW-29 IS CLOSED — Story 7.8, 2026-08-31

**What shipped.** The alignment vocabulary is re-partitioned **by consumer, keyed on element type**,
which is the partition this entry's own root-cause note said was needed:

- A third closed set, `TableStyleAlignTokens` / `closedTableStyleAligns` — `left, center, right` —
  in `folio-go/internal/template/closedsets.go`, with an `IsTableStyleAlign` predicate beside
  `IsStyleAlign`.
- `decodeStyle` (`folio-go/internal/template/parse_bands.go`) takes an `ElementType` and selects the
  set by it. A table's `style.align` and `headerStyle.align` are now a **located** load error naming
  the element and the field, with a message derived from the set that rejected it — so it never
  names `justify` as legal for a table. A text element's `justify` is untouched and still raises the
  document to `2.0`.
- The property-command arm (`folio-go/component_commands.go`) selects the same way, discharging
  `IsStyleAlign`'s own obligation that the command path validate against the same single source the
  loader does. That was the one remaining route by which a table document could reach `2.0`.
- The format-version half needed no code at all on the file path and is asserted rather than
  implemented: `versionRequiredByContent` runs only at save, on a fully validated `*Document`, so a
  loader refusal closes the `2.0` raise by construction.

**The blocker this entry recorded is discharged, and by a different answer than it framed.** The
entry said a located error "cannot reach a designer author without a THIRD per-field style
diagnostic code", and that this was a lead call. The lead ruled (**D-7.8.1**, 2026-08-31) for the
**general** form instead: one code, `TEMPLATE_FIELD_INVALID`, supplied by `newLoadError` itself, so
every uncoded load-error site in `internal/template` became coded by construction. The
registry-policy rule it establishes is written into `internal/diag/diag.go` in place of the
reservation that asked the question: **the general code is the default; a specific code is minted
only when a named consumer must BRANCH on it to behave differently.** `TEMPLATE_MALFORMED` keeps
destroying its own messages, for the reason it was written — it now names the genuinely malformed
template rather than every located field error as well.

**Follow-on obligation, deliberately NOT done here.** Auditing `STYLE_COLOR_INVALID` and
`STYLE_LINE_SPACING_INVALID` against that rule is **D-7.8.2**, triggered by the `folio-go/v0.1.0`
tag, because AD-14 makes removing a code a breaking change and that is free exactly once.

**And the correction that ruling needed (D-7.8.5).** D-7.8.1's stated ground — *"a `LoadError`'s
message never quotes the document"* — was **false**, and moving this population off
`TEMPLATE_MALFORMED` therefore switched off `reportableMessage`'s reflection guard for it. Measured
at this story's baseline: a well-formed document whose `style` key held 2048 Thai characters went
from a 35-character refusal to 512 bytes of the author's own file, cut mid-rune. The resolution
makes the premise **true** — `LoadError.Error()` bounds every author-supplied fragment as it renders
it into the sentence, in **runes**, with a visible `…`, while the struct fields stay **complete** for
a Go integrator's CI log. Four fragments were found and bounded, not one: the **value** (84 runes),
the **element id** (24), the **field path** (96) and the **reason** (256), each derived in
`errors.go`'s own comment. The same document now produces a **430-byte** message: 83 author runes
occupying 249 bytes, a 3-byte elision marker, and **178 bytes of the engine's own words**. (An
earlier version of this line said 347, having subtracted a *rune* count from a *byte* total;
corrected at the story's close.) The guarantee is that the value can never exceed **half the host's
512-byte window** and that no engine-authored word is ever truncated — not that the engine's words
outweigh the author's in every message. `reportableMessage`'s treatment of `TEMPLATE_MALFORMED` is
unchanged.

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
- **Owner:** **Story 15.2a** — *a component command means exactly what it names* — filed 2026-08-31
  in Epic 15 and **sequenced before Story 15.3 cuts the tag**. **15.2a MUST ALSO READ
  [DW-75](#dw-75), filed at Story 8.2's close:** two of the five encoders this entry consolidates do
  not merely spell escaping differently, they **corrupt non-BMP text** and bind to the wrong path.
  That is the same story's scope and was proved by execution, so the consolidation is a repair, not a
  tidy-up, and its acceptance must assert a non-BMP round trip explicitly. Not Epic 8: D-7.7.12 holds and Epic 8
  does not widen this defect, unlike Story 8.0's case. **Its Go half JOINS D-7.8.3's before-the-tag
  set**, which is therefore **two** items, not one — a duplicate-key refusal narrows the exported
  `ApplyComponentCommand` and has not shipped. **Both enforcement points are ONE subject and land in
  ONE commit**: without the Go half the only available test is *"the encoder produces well-formed
  JSON"*, which tests the fix rather than the property and goes green again the moment a future
  encoder regresses.
- **Severity:** **HIGH** — raised from medium 2026-08-31 at Story 8.2's plan gate, which measured the
  mechanism rather than the symptom. **This is not "a bad value reaches the document"; it is
  COMMAND-SHAPE INJECTION.** The designer's numeric encoder splices author text into command JSON
  **unquoted**, so a value of
  `0}},"ids":["other"],"changes":{"width":{"op":"set","value":10` yields **valid JSON with duplicate
  keys**. Go decodes into `map[string]json.RawMessage` where **last key wins**, while the arity check
  still counts four fields — so the command mutates **a different component's different property**.
  Escalation to another command `kind` is blocked only by an **arity coincidence**, not by a check.
  The register's previous claim that no bad value reaches the document is **false**.
- **Status:** OPEN — **but `component-property-command.ts` has MOVED since this entry was written.**
  Story 8.2 (2026-08-31) routed that file's `quote()` through `JSON.stringify`, minimally and by
  itself, because a chain name is author-typed and travels through the quoter: `quote()` escaped
  `\ " \n \r \t` and **nothing else**, while JSON requires all of U+0000–U+001F, so a pasted C0
  control produced **invalid JSON** and Go answered with a generic parse failure instead of the located
  refusal naming the field — engine-side name validation cannot substitute, because the bytes are
  malformed before the rule can run. **`rawNumberLiteral` and the other four encoders were NOT
  touched**, and neither was the shared authority: those remain wholly Story 15.2a's. **15.2a must
  RE-READ this file rather than assume the shape recorded below.**

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

### DW-35 — the canvas hard-codes ONE font stack regardless of the document's own `fonts` map, so a template naming a different chain still paints with these three families — **CLOSED by Story 8.4e, 2026-09-01. Cause two by Story 8.4a; cause one's vocabulary layer by Story 8.4b; cause one's attribution residual by Story 8.4e. All four things this entry named are now closed.**
- **Deferred by:** Story 7.4's close (2026-08-30), observed while fixing the owner-reported Thai canvas
  defect at `c6e4d03` and recorded there in the commit message
- **Owner — ONE PER CAUSE, because the causes were split. ALL OF THEM ARE NOW DELIVERED.**
  *(This heading read "…and only one of them is delivered" for as long as the attribution layer was
  open. It is corrected here at Story 8.4e's close rather than left standing beside the bullets below,
  which now record every cause closed; the historical reading is preserved in this note rather than
  erased.)* **Severity:** MEDIUM. Both assignments below are *rulings*, not recommendations — see the
  standing rule below.
  - **CAUSE TWO — Story 8.4a. DELIVERED AND CLOSED 2026-09-01 (`c4cd60c`); the story is finished and
    owns nothing further here.** RULED 2026-08-31 (D-8.4.1) to Story 8.4, then SPLIT to its named
    successor 2026-09-01 (D-8.4.6) when 8.4's plan gate returned `multiple-goals`. The ruling
    anticipated this: *"8.4, or the named successor story if the gate splits it."*
  - **CAUSE ONE — SPLIT INTO TWO LAYERS (narrowed 2026-09-01 at Story 8.4b's close), AND BOTH LAYERS
    ARE NOW DELIVERED.** *(The narrowing that split this cause recorded "ONLY ONE OF THEM IS
    DELIVERED", which was true of the state it was written in: the vocabulary layer had just landed
    and the attribution layer had not. Story 8.4e closed the second layer on 2026-09-01 (`21f93b4`),
    so the assertion of incompleteness is corrected here; the split itself, and the fact that it was
    made mid-flight, stand as recorded.)*
    - **VOCABULARY LAYER — Story 8.4b. DELIVERED AND CLOSED 2026-09-01 (`90cdf8e`).** Ruled into
      existence by D-8.4.14 and sequenced after 8.4a. The browser can now name the engine's faces at
      all: the three shipped files are declared a second time under `fonts.Shipped()`'s own spellings
      and the canvas fragment rule asks for those, with no chrome token edited, no binary added and
      no mapping table built. 8.4b owns nothing further here.
    - **ATTRIBUTION LAYER — Story 8.4e. DELIVERED AND CLOSED 2026-09-01 (`21f93b4`).** The fragment
      stack was a **fixed stylesheet constant**, not the document's chain, and a shipped-face fragment
      carried **no face identity on the wire** — only carried faces carried an `assetKey`, so a shipped
      fragment painted with no inline family and fell to the fixed Latin-first stack. Measured at
      8.4b's close: the three faces' pairwise cmap overlaps are **339 / 529 / 230** codepoints and all
      three cover `A` and `5`, so a document whose chain is `["Noto Sans Thai"]` had the engine
      measure a Latin run with one face while the browser rasterized it with another. Closing it needed
      **per-fragment shipped-face attribution on the projection** — the shape 8.4a built for carried
      faces, extended to shipped ones — and that is what 8.4e delivered; see the closing note at the
      end of this entry.
      **OWNERSHIP CORRECTION, RECORDED RATHER THAN OVERWRITTEN.** This bullet stood for one story as
      *"OPEN, AND OWNERLESS … escalated to the orchestrator for routing"*, and that was already
      superseded when this entry was last read: **D-8.4.26 assigned the residual to Story 8.4e**, and
      both `epics.md` and `sprint-status.yaml` carried the story. The stale passage is corrected here
      rather than left standing beside the truth. Before that it read **`Owner: Story 8.4a` alone**,
      and 8.4a's close left it that way — a finished story standing as the sole owner of a cause it
      explicitly did not close, which is how a surviving cause loses its owner without anyone deciding
      it should. **Twice now** on this one bullet, which is why the correction is written down.

  **The successor is sequenced IMMEDIATELY after 8.4, not "later in Epic 8"**, and Story 8.4
  discloses the canvas limitation as a **test** rather than a comment, so the gap between the two
  stories is asserted rather than described.

  **THIS ENTRY PREVIOUSLY CARRIED TWO CONTRADICTORY `Owner:` BULLETS** (Story 8.2, then Story 8.4)
  with a `Severity:` bullet wedged between them. That is a **defect in the entry**, not an ambiguity
  for a downstream reader to resolve — the entry could not have been read consistently by anyone.
  Collapsed to one bullet here; the superseded placement grounds are preserved below as history.

  **Standing rule set with this ruling:** *a "recommended owner" written by a story's spec is **not**
  an owner until a ruling or a decision-log entry adopts it.* DW-35's assignment to Story 8.4
  originated as "Recommended owner: Story 8.4" in **Story 8.2's spec**; `awk` over the run's decision
  log found **zero** occurrences of `DW-35`. It propagated for two stories as a recommendation that
  every downstream reader took for a decision. This was the **fourth** ownership-mechanism failure of
  the run.

  **The AC number was wrong here and in the epic, identically.** This entry said Story 8.4's **AC4**
  is DW-35 as an acceptance criterion. **AC4 is DW-83** (D-8.3.5); **DW-35 is AC5**. The epic text
  carried the same off-by-one, so cross-checking one against the other could never have caught it —
  **they are not independent witnesses, because one was derived from the other.**

  **Why 8.4 and not a neighbour: 8.4 does not widen this defect, it CREATES the condition.** Before
  8.4 an embedded face cannot render at all, so no author can reach the state. After it, the engine
  measures with the embedded face while the browser has **no CSS family for it at all** and falls
  through to `sans-serif` — the owner-reported defect fixed at `c6e4d03`, rebuilt for 8.4's own
  headline use case. By the rule set at Story 8.0, **a defect a story makes reachable is a
  precondition of that story.**

  **The blocker is removed: the design decision above a builder's authority has been MADE (D-8.4.1).**
  *An embedded face's CSS family name is derived from its **asset key**, never from `font.family`.*
  AD-8 — *"the asset key decides, even where an embedded face and a shipped face share a family
  name"*; and `font.family`/`font.style` are display identity, *"never used to resolve or substitute
  a face — resolution is by asset key alone."* Deriving the CSS family from `font.family` would let
  an embedded "Inter" collide with a shipped "Inter" in the browser's font registry — AD-8's own
  hazard, one layer down. If 8.4's plan gate sizes the paint half out, it goes to a **named successor
  sequenced immediately after 8.4**, and 8.4's record states the canvas limitation explicitly.

- **Superseded placement grounds, kept as history.** Owner was *"Epic 8's plan gate"*, then Story
  8.2 (at 8.1's close, per D-8.0.5 and 8.1's Design Notes R7, on reachability), then Story 8.4. The
  8.2 placement recorded a ground that was later **falsified**: it claimed the fix *"needs the
  projected per-component chain, which 8.2 needs anyway"* — both halves wrong, since 8.1 already
  landed `CanvasComponent.FontFamily` and `FontChains[].Entries`, and 8.2 edits the document-level
  map and never needs the per-component join. **The real obstacle was recorded nowhere:** the
  browser's `@font-face` families are `IBM Plex Sans` / `IBM Plex Mono` / `IBM Plex Sans Thai` (the
  shipped **Noto** files registered under IBM Plex names) while a chain's entries are `Noto Sans` /
  `Noto Sans Thai` / `Noto Sans SC`, so **a chain's entries cannot be used as CSS family names**.
  That is what D-8.4.1 now settles for the embedded case.

- **Status:** **CLOSED** by Story 8.4e's implementation commit `21f93b4`, 2026-09-01. This entry has
  TWO causes and cause one has TWO layers; **all four** things
  are now closed, and the layers are still listed separately because conflating them is how a
  residual disappears.
  **CAUSE TWO is CLOSED** by Story 8.4a (2026-09-01) — see the closing note at the end of this entry.
  **CAUSE ONE's VOCABULARY LAYER is CLOSED** by Story 8.4b (2026-09-01, `90cdf8e`) — see 8.4b's
  closing note at the end of this entry. **CAUSE ONE's ATTRIBUTION LAYER is CLOSED** by Story 8.4e
  (2026-09-01, `21f93b4`) — see 8.4e's closing note, which is the last one in this entry and states both what
  is proved and what is not.
  The design decision cause two inherited was made by D-8.4.1 (quoted below): *an embedded face's CSS
  family name is derived from its **asset key**, never from `font.family`.* **Cause one's equivalent
  is D-8.4.14 (2026-09-01)**, which ruled the register's stated blocker FALSE and gave cause one a
  named owner, **Story 8.4b**. *(This clause continued **"so cause one is open because the code is not
  written, not because the decision is missing"** until Story 8.4e's close, where it was found still
  asserting an OPEN cause inside a bullet whose own first word is **CLOSED**. The code is now written —
  the vocabulary layer at `90cdf8e`, the attribution layer at `21f93b4` — so the clause is corrected
  here rather than left standing. **This is the third self-contradiction found in this entry's closure**;
  the other two were two internal headers repaired at `24ca8f6`, and this one survived that sweep because
  it reads as prose about a decision rather than as a status claim.)* (This bullet previously read
  *"OPEN and UNRULED … no equivalent decision has ever been made"*; that was already untrue when it was
  written. See the correction in the closing note.)

  **CAUSE TWO — a face the DOCUMENT CARRIES has no family in the browser at all.** For cause one
  there is at least a shipped file behind the chain entry, registered under *some* family the browser
  knows; the defect is that the two vocabularies do not meet. For a carried face there is **nothing**
  — no `@font-face`, no family name, no bytes in the browser — so `.canvas-text-fragment` falls
  straight through to generic `sans-serif`. That is the `c6e4d03` defect rebuilt for the headline use
  case of the story that created the condition.

  **Story 8.4 delivered the MEASUREMENT half and disclosed the paint half, both by test.**
  `folio-go/canvas_embedded_face_test.go` pins that `CanvasWithTextPaint` over the embedded-face
  document produces fragment origins, advances and line widths **identical** to the PDF path's, from
  the render path's own `fontChain`/`shapeSegments`/`chainVerticalModel` — red-proved by building the
  canvas's `fontCache` without the document, which is the one thing the two paths do not share.
  `folio-designer/src/canvas-font-stack.test.ts` records the paint gap and names 8.4a, and its new
  assertion — that **no** designer source registers a font at runtime (`new FontFace`,
  `document.fonts.add`) — is written to redden when 8.4a arrives.

  **What 8.4a owns, recorded so it is disclosed rather than discovered** (measured at `15ca0dd`):
  per-fragment face attribution on `CanvasTextFragment` — the value is already in scope at
  `page_setup.go`'s fragment loop and **discarded**, and a fragment is exactly one face by
  construction, so per-component attribution would be wrong for any mixed-script element; its TS type
  and the `hasOnly` guard at `engine-protocol.ts` in the same commit (an unlisted key blanks the
  canvas with **no diagnostic**); a **new** fragment-level record in `canvas_projection_wire_test.ts`
  (the wire test records the top level and the two font-chain levels only, so a new field would
  redden nothing); a named asset-key → CSS-family derivation module; runtime `FontFace` registration
  at **document** scope, not per component (the `ImagePaint` effect in `App.tsx` is the closest
  pattern and is the *wrong* lifetime); the `.canvas-text-fragment` `font-family` var head; the
  `canvas-font-stack.test.ts` re-author, which must **keep** its shipped-face guarantees while adding
  the embedded exception with its own teeth; and a deliberate narrowing of
  `canvas-authority-contract.test.ts`, whose blanket `document.fonts` → `fontReadinessOnly` rewrite
  makes its own prohibition dead and would let a measurement call in unnoticed. The asset bytes
  themselves need no new transport: `AssetBytes` is media-type agnostic end to end and is usable for
  font bytes as-is; the only image-specific hop is its consumer in `App.tsx`.

  **What Story 8.2 made reachable is recorded by a TEST rather than by a comment.** 8.2 is the story
  that lets an author BUILD a chain, so it is the first at which a
  document can declare one whose first covering entry is not `Noto Sans` — `["Noto Sans Thai"]`, say —
  after which the engine measures with that face while the browser paints the fixed Latin-first stack.
  8.2 states this and does not fix it: the fix is a design-system decision (rename the generated
  `@font-face` families, rippling into `tokens.css`'s `--font-page` and its three type tokens and into
  `design-contract.test.ts`, or generate a face-name → CSS-family map) above a builder's authority.
  `folio-designer/src/canvas-font-stack.test.ts` now carries three tripwires: that the fragment stack
  is a stylesheet constant with no document input, that the engine's face names and the browser's
  family names **do not intersect at all**, and that no designer source turns a projected chain entry
  into a CSS family. A comment asserting a negative was carrying a test's evidentiary burden; it is a
  test now.

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

**MEASURED CORRECTION to the paragraph above (Story 8.4a, 2026-09-01).** The `--font-page` hazard it
names **is not live on this path**. `tokens.css:11`'s `--font-page` feeds `--type-page-{title,body,fine}`,
of which only `--type-page-body` is used, on `.file-message` — chrome, not canvas text.
`.canvas-text-fragment` uses a hardcoded Latin-first **literal** and no `var()` at all, deliberately
pinned by `canvas-font-stack.test.ts`. The ORDER constraint the paragraph states is still correct as a
requirement; only its justifying example is stale.

---

## DW-35 IS CLOSED — Story 8.4e, 2026-09-01

Everything ABOVE this line is the entry as it stood while it was open, kept verbatim with its
corrections marked in place. Everything BELOW is closure: three closing notes, one per layer, in the
order they landed — cause two (Story 8.4a), cause one's vocabulary layer (Story 8.4b), cause one's
attribution residual (Story 8.4e, which is the last of the four and closes the entry).

*(Added at Story 8.4e's close. The closure itself deferred this marker as a mis-delimitation risk — the
entry runs some four hundred lines and the build declined to guess where it ended unattended. The
boundary is not a guess: it is the line before the FIRST closing note, which is exactly the split
DW-24, DW-25, DW-29 and DW-36 each make. The four siblings carry this marker and DW-35 did not, so a
reader arriving at the middle of this entry had no way to tell a live claim from a historical one —
which is the same failure the three self-contradictions above came from.)*

---

**CLOSING NOTE — CAUSE TWO, closed by Story 8.4a (2026-09-01).**

**What was closed.** A face the document CARRIES now has a family in the browser. `CanvasTextFragment`
gained one optional field, `assetKey`, carrying the document asset the engine resolved **that fragment**
to (`page_setup.go`, populated at the fragment append from the chain-scoped `fontCache` through the new
`carriedAssetKey`/`embeddedFaceAssetKey` inverse of the mint in `embedded_face.go`). The browser derives
its own CSS family from that key — `folio-designer/src/embedded-face-family.ts`, the one site that makes
that decision — fetches the bytes over the **existing** media-type-agnostic `asset` operation, and
registers a `FontFace` **once per document** in `folio-designer/src/embedded-face-registry.ts`. The
fragment span asks for that family, and only once the face has actually registered; a fetch that fails
degrades to the stylesheet's declared stack with the canvas still painting and the worker still running.

**The derivation is from the ASSET KEY, per D-8.4.1, and the two sides derive independently.** The
engine's reserved `asset:` prefix is still spelled in one Go file and in no TypeScript at all: the wire
carries the KEY, so the browser's family is its own namespace rather than a second copy of an
engine-internal one.

**Guards, widened rather than weakened.** `canvas-font-stack.test.ts`'s declared-families tie now also
ties, for the carried case, the family the fragment ACTUALLY asks for (read off a rendered DOM node) to
the asset the engine attributed it to, and to the family the seam registers under. Its
chain-entry tripwire was inverted into an allow-list — a font-family position may name **only** an
asset-key-derived family — and it now strips comments, so it no longer taxes prose. Story 8.4's
*disclosure of absence* ("no designer source registers a face at runtime") was deleted under its own
written pre-authorisation and replaced by its positive twin: registration happens in **exactly one named
seam** and nowhere else, with the detector unchanged and still proved against its own fixtures.

**And the dead prohibition was repaired.** `canvas-authority-contract.test.ts`'s blanket
`document.fonts` → `fontReadinessOnly` rewrite (appended as a drive-by unblock in `7bfb076`) made its own
`/\bdocument\.fonts\b/` rule incapable of matching anything; deleting the rule left the suite green. The
rewrite is now scoped to `document.fonts.ready` alone, `new FontFace` was added to the prohibitions, the
whole scan is routed through the existing comment stripper, and the file now carries the mutation proofs
it never had — `document.fonts.add` caught, `e2e/engine-worker.spec.ts`'s `await document.fonts.ready`
not caught, and deleting either rule reddens a named test.

**WHAT IS NOT CLOSED, AND WHY NOT — CAUSE ONE.** A chain of **shipped** faces still disagrees across the
seam: the engine measures with `Noto Sans Thai` while the browser asks for `IBM Plex Sans` first,
because `scripts/build-wasm.mjs` registers the three shipped Noto files under IBM Plex family names while
a chain's entries are the ENGINE's face names. `canvas-font-stack.test.ts` records those two vocabularies
as **deliberately disjoint**. Story 8.4a's scope is the enumeration in "What 8.4a owns" above, adopted
by D-8.4.6, and **not one item on it mentions a shipped mapping**, so extending the tie to the shipped
half would have been a red assertion inviting a builder to weaken it back rather than a stronger guard.
Note the aliasing trap for whoever takes it: the generator's `'IBM Plex Mono'` is **Noto Sans SC**
(`build-wasm.mjs`), not a mono face.

**CORRECTION, RECORDED RATHER THAN LEFT STANDING (2026-09-01).** This closing note as first written said
cause one's fix is *"a design-system decision above a builder's authority, and no ruling has ever made
it"*, and the paragraphs above still frame the choice as *rename the generated `@font-face` families* vs
*generate a face-name → CSS-family map*. **Both statements were already stale when they were written.**
**D-8.4.14 ruled on 2026-09-01**, before 8.4a's implementation landed, and it ruled the register's own
blocker **FALSE**: renaming the generated families is *not* the decision and would be **wrong**, because
IBM Plex is the design system's specified typeface, named throughout `DESIGN.md` and promised in the
release **licence manifest** — the Noto files were the stand-in, not the label. The mapping table is
rejected too, as a second authority maintained in lockstep with the shipped `FontSet`. **The fork
dissolved instead of escalating:** DW-35 is about what the CANVAS paints with and says nothing about
chrome, so the shipped faces are registered **additionally** under the engine's own face names and the
canvas fragment rule points there — D-8.4.1 one case over, and grounded on AD-17.

**What would discharge cause one, and who owns it — SUPERSEDED 2026-09-01 by 8.4b's closing note at
the end of this entry, which records what 8.4b actually delivered and what it did not. Kept because the
prediction and the outcome differ in one load-bearing way: 8.4b closed the VOCABULARY layer only.**
**Story 8.4b — *the canvas can name the face the
engine measured*** (`epics.md`), ruled into existence by D-8.4.14 and sequenced after 8.4a. It (a)
registers each shipped face additionally under the engine's own `FontSet` spelling, (b) points the
canvas fragment rule at those names, editing **no** chrome token, and (c) **replaces rather than
weakens** `canvas-font-stack.test.ts`'s disjointness assertion — that assertion records the old state
and *should* go red; its successor asserts the canvas fragment stack's families **contain** the engine's
face names, and must not be softened to an `arrayContaining` to keep it passing. **No further ruling is
outstanding.** This entry stays **OPEN** until 8.4b lands, because the defect is still live on the
shipped path — not because the decision is still missing.

**What this repository still cannot prove, stated plainly.** Nothing here can execute a real font load,
a real `document.fonts.add`, or a rasterized glyph: jsdom applies no stylesheet and implements no font
loading, and the Playwright suite is compile-only (`test:e2e:compile` is `tsc --noEmit`; browser e2e is
deferred by D-000.4 — *a citation D-8.4.25(b) later declared void, the suite having already arrived;
what is deferred is CI EXECUTING it, DW-101*). Story 8.4a's gates prove the derivation, the registration call, the fragment's
rendered family, the guards and the protocol shape. The claim that the canvas **visibly** paints with the
carried face is unverified by anything in this repository.

**CLOSING NOTE — STORY 8.4b, 2026-09-01 (`90cdf8e`): CAUSE ONE'S VOCABULARY LAYER IS CLOSED, AND THE
ENTRY IS NARROWED TO WHAT SURVIVES.** The generator now emits **six `@font-face` rules over three
files**: the three IBM Plex chrome rules byte-identical, plus three declaring the **same files** under
`fonts.Shipped()`'s own spellings, and `.canvas-text-fragment` asks for those three names in the
engine's own order. The stack is order- and file-preserving, so **no rasterization changed on the day
it landed**. No chrome token was edited, no binary added, no `assets` slot created and **no mapping
table built in either direction** — the browser family now *is* the engine's name, so there is nothing
to map. D-8.4.14's ruling is implemented as ruled, and the fix descriptions this entry used to carry
(*rename the generated families* / *generate a face-name → CSS-family map*), both measured FALSE by
D-8.4.14, are gone from the live framing above rather than left standing beside the truth.

**A guard now exists where none did.** `font-binary-identity.test.ts` performs the family→file join
this codebase never performed, pins the whole six-entry map exactly, pins the deliberate
two-names-one-file interval with **Story 8.4c named in the assertion's own failure message**, and ties
each engine-named family to the bytes `folio-go/fonts/fonts.go` embeds under that face name. The
disjointness record was **replaced by its opposite, not weakened**: both intersection assertions are
gone, the chrome floor is intact, and the engine's names are **parsed from `Shipped()`** rather than
hardcoded — verified at close by renaming a face in `Shipped()` and watching two named tests redden,
which the hardcoded form could not have done.

**WHAT IS NOT CLOSED — THE ATTRIBUTION RESIDUAL, AND IT IS OWNERLESS.** The canvas can now *name* the
engine's faces; it still cannot say **which** face belongs to **which fragment**. The fragment stack is
a fixed stylesheet constant naming all three faces in one order, while a document may declare a chain
whose covering face is not the stack's first, and a shipped-face fragment carries **no face identity on
the wire**. Measured at 8.4b's close: pairwise cmap overlaps of **339 / 529 / 230** codepoints, all
three faces covering `A` and `5`. So the engine can measure a Latin run with `Noto Sans Thai` while the
browser's Latin-first stack rasterizes it with `Noto Sans` — the same shape of disagreement this entry
was opened for, one layer in. Closing it needs **per-fragment shipped-face attribution on the
projection**. **No story owns it**, and 8.4b's closer deliberately did not assign one: none of 8.4c,
8.4d, 8.5 or 8.6 covers it, and creating an owner is a ruling. **Escalated to the orchestrator.**

**And the limit is unchanged by 8.4b.** What is now proved is that the browser is **asked** for the
engine's face names and that a face under each is **declared from the engine's own bytes**. That any
glyph **rasterizes** with those faces remains unverifiable here, and becomes checkable **when CI
executes the Playwright suite (D-8.4.25(b))**. *(Trigger corrected 2026-09-01 at Story 8.4e's close.
This sentence used to say "when browser e2e arrives (D-000.4)". **D-8.4.25(b) declared that trigger
void**: browser e2e had already arrived — `folio-designer/e2e/` carries 12 executable Playwright specs
and a real `webServer` — so the named event could never fire again. What is missing is not the suite
but its EXECUTION: `.github/workflows/ci.yml` runs only `test:e2e:compile`, which is `tsc --noEmit`.
Wiring that is DW-101.)*

**CLOSING NOTE — STORY 8.4e, 2026-09-01 (`21f93b4`): THE ATTRIBUTION RESIDUAL IS CLOSED, AND WITH IT
THE WHOLE ENTRY.**

**What was closed, and it is the last of the four things this entry named.** A shipped-face fragment
now carries the engine's identity for the face it was measured with, exactly as a carried one has
carried its asset key since 8.4a. `CanvasTextFragment` gained one optional field, `Face`
(`json:"face,omitempty"`), populated at the fragment append in `folio-go/page_setup.go` from the value
that site already had in hand and **discarded** — `fragment.face`, which for a shipped face IS the
caller's `FontSet` key verbatim, because `chainFaceNames` mints a name only for an embedded entry and
copies `entry.Face` for every other. `folio-designer/src/shipped-face-family.ts` — one small module, one
exported derivation, its own test file, on `embedded-face-family.ts`'s model — turns that name into the
CSS `font-family` value the fragment asks for: the attributed face **quoted and first**, then the
declared stack as its tail. `TextPaint` sets it inline, at both of its call sites, so the repeated
sheets' `ComponentEcho` paints with the same face as the home occurrence.

**BOTH CAUSES, RESTATED AS ONE MECHANISM, because that is what the entry was always about.** The canvas
paints at the engine's own x-positions (AD-17), so it must rasterize with the face the engine MEASURED
with or every fragment's drawn width disagrees with the x of the fragment after it. Cause two was a face
the DOCUMENT carries, for which the browser had nothing at all — closed by 8.4a, which put the asset key
on the wire and registers a `FontFace` under a family derived from it. Cause one was a face the BUILD
ships, and it had two layers: the browser could not NAME the engine's faces (closed by 8.4b, which
declares the same three files a second time under `fonts.Shipped()`'s own spellings), and then it could
not say WHICH face belonged to WHICH fragment (closed here). The residual mattered because the three
shipped faces' cmaps genuinely overlap — **339** (Sans×Thai) / **529** (Sans×SC) / **230** (Thai×SC)
code points, and **all three cover `A` and `5`** — so a document whose chain is `["Noto Sans Thai"]` had
its Latin measured with Noto Sans Thai and rasterized with Noto Sans: right glyphs, wrong advances,
creeping out of position. That document is now the I/O matrix row a test asserts.

**NO MAPPING TABLE, NO RENAME, NO NEW BINARY, NO CHROME TOKEN.** The wire value is the `FontSet` name
**verbatim** — D-8.4.14's *"one rule for one question"* — and 8.4b had already made that name a family
the browser can resolve, so there was nothing left to map. The two alternatives this entry used to frame
the fix by, both measured FALSE by D-8.4.14, stayed rejected. `tokens.css` was not edited, and the guard
that would redden if it had been (`canvas-font-stack.test.ts`'s "no `--font-*` token names an engine
face") is untouched and green.

**THE OLD RECORD WAS REPLACED, NOT WEAKENED.** `canvas-font-stack.test.ts`'s *"the fragment stack is a
stylesheet constant with no document input"* was a **disclosure of absence** whose own prose header said
closing it *"is a different story"* — a written pre-authorisation of its own retirement, and the same
move 8.4a made on Story 8.4's registration disclosure. It is **gone**, replaced by its positive twin:
a shipped fragment's family DOES come from the projection, through exactly one named seam and nowhere
else. Both bounds the retired record also carried are kept in the replacement (no `var(` in the
declaration; the `requested.length >= 3` floor). The census guard was widened to a **closed set of
exactly two** approved expressions — never an `arrayContaining` — because the fragment's family has now
moved out of the stylesheet and into an inline style for BOTH populations, and an inline family string
would otherwise escape an `App.css`-only scan without anyone editing a guard. `requested.slice(0,3)`'s
order tie, D-8.4.14's delivered guard, is untouched: `.canvas-text-fragment` survives as the fallback
for an unattributed fragment, and three separate guards still read it.

**The tail has one authority.** The seam's inline value carries the declared stack behind the attributed
face — an inline declaration REPLACES the rule rather than extending it, so without a tail a codepoint
the attributed face does not cover would fall to the browser's default instead of to the other shipped
faces. That list is therefore spelled twice, and a guard reads **both sources** and asserts they are the
same sequence, entry for entry.

**WHAT IS PROVED, AND WHAT IS STILL NOT — stated so a compile pass is not read as a run.** Proved: the
engine's attribution (Go, over a shipped-only chain, a carried-only chain and a mixed-script element);
the wire shape in both directions as **exact key sets**, with `face` and `assetKey` mutually exclusive
on every emitted fragment; the browser's validation of the new key, bounded by the same 512 a chain
entry's `face` already uses; the derivation; and the family the rendered `.canvas-text-fragment` node
actually asks for, at the home fragment **and** at the `ComponentEcho`. **Not proved: that a glyph was
rasterized with that face.** jsdom applies no stylesheet and loads no font, and `test:e2e:compile` is
`tsc --noEmit`. Per **D-8.4.25(b)** that becomes checkable **when CI executes the Playwright suite**;
per D-8.4.25(d)+(e) the executed browser assertion is owed at the **epic gate**, behind CI wiring
(DW-101), and must not be added to a suite CI does not run. This entry closes on the mechanism, not on
a rasterization anyone watched.

---

### DW-36 — the designer's live drag/resize clamp still bounds Y by band height in ALL THREE bands, so Story 7.5's lift is reachable by command and not by hand — **CLOSED by Story 7.6, 2026-08-31**
- **Deferred by:** Story 7.5's build (2026-08-31); filed into this register at the story's close, where
  it was found recorded only in the spec's frontmatter
- **Owner:** **Story 7.6's plan gate** — a gate, never an event, per D-000.73. 7.6 is also the story that
  draws the later sheets a drag past page one would target
- **Severity:** MEDIUM — the engine and the protocol now accept the placement, so the product is
  internally inconsistent until this closes
- **Status:** **CLOSED** by Story 7.6's implementation commit `c834158`, 2026-08-31. See the closing
  note at the end of this entry.

**The gap.** `folio-designer/src/resize-anchor.ts:29` `proposedBounds` clamps `y` to
`limitHeight - originalHeight` (`:34-35`) and a south resize to `limitHeight` (`:52`).
`folio-designer/src/App.tsx:701` fills `DragLimit` from `{ width: band.width, height: band.height }` for
**every** band, and `DragLimit` carries **no band name**, so the helper cannot tell the content band from
a repeating one **even in principle**. `resize-anchor.test.ts:46,53` still pins the clamp and is
unmodified by Story 7.5.

**Why 7.5 did not close it.** The story is scoped to the command layer and the projection-admission
mirror, and the clamp's own header (`resize-anchor.ts:9-18`) declares it a UX affordance rather than a
second authority — so it is not an authority bug. Design Notes judgment 3 assigns it to 7.6 explicitly:
there are no later sheets to drag onto until 7.6 draws them.

**What this means for anyone reading 7.5 as shipped.** A component CAN be placed windows down the column
by command, survives the canonical bytes, and is admitted by the browser protocol. It CANNOT be dragged
there. **Do not read Story 7.5 as "done in the UI".** Closing this requires giving `DragLimit` the band's
identity and consulting the same `bandsCappingVertically` list the Go validator and `engine-protocol.ts`
now share — a third consumer of the tie, not a fourth spelling of it.

**Also for 7.6's plan gate, recorded here because 7.6 is the owner of both.** A content component
**taller than one window** — which Story 7.5 newly permits — makes `layout.Paginate` return
`*OverflowError`, which Ruling C degrades to a reported count of **one window**. 7.6 would therefore draw
**one sheet** for such a document. That is a floor, not a lie, and it is deliberate: the settled
alternative was a blank canvas with no attributable error. But 7.6 must know it before it draws, because
the symptom reads as a missing page rather than as a degradation.

---

## DW-36 IS CLOSED — Story 7.6, 2026-08-31

`folio-designer/src/resize-anchor.ts:1` now **imports** `BANDS_CAPPING_VERTICALLY` from
`engine-protocol.ts` and `:53` gates the ONE `limitHeight` that both vertical clamps consume — the
move's `y` and the south resize's `bottom` — on that list. `DragLimit` (`:39`) carries the band's
identity, and `App.tsx` fills it from the band it is already rendering. This is the **third consumer
of the tie, not a fourth spelling of it**, which is the wording this entry itself demanded:
`folio-go/component_commands.go:1795` declares `bandsCappingVertically` and consumes it at `:1820`
and `:1840`; `engine-protocol.ts:88` declares the browser's half and consumes it at `:258`; the drag
clamp reads the protocol's declaration rather than restating it.

The horizontal limit is deliberately **ungated** and unmoved: `resize-anchor.ts:56` and `:72` still
clamp width against the band in every band, and the diff touches neither. Nothing may hang off the
side of any band, in any band, ever.

`engine-bounds-mirror.test.ts` reads **all three sources**. Its `'consumes the list at the site it
governs, in all three consumers'` case pins the two Go sites, the protocol site, the drag clamp's
**import statement** and the drag clamp's gate expression; its drift proof rewrites the clamp back to
the unguarded `limit ? limit.height : Number.POSITIVE_INFINITY` and asserts the pin no longer matches.
The six numeral `pairs` and `toHaveLength(6)` are untouched — this tie is a **predicate** mirror, not
a numeral one.

**Verified at close by mutation, not by reading.** Restoring the unguarded clamp turns the content-band
move and south-resize cases red in `resize-anchor.test.ts`, both mirror assertions that read the drag
clamp as the third consumer, and the App drag that runs past one window's foot.

The second half of this entry — the note "for 7.6's plan gate" about a content component taller than
one window degrading to a reported count of one — was **carried, not lost**: it is the Ruling C
degradation, one of the three causes Story 7.6's `contentWindowCountIsFloor` now reports, so the
canvas says out loud that the number is a floor instead of silently drawing one sheet.

---

### DW-37 — `contentWindowCount` has no upper bound, and the non-text content extent is an unguarded sum
- **Deferred by:** Story 7.5's build (2026-08-31); filed into this register at the story's close, where
  it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 7's retrospective, or the plan gate of the next story that changes canvas geometry
  bounds (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** LOW — not reachable through any shipped template, and not reachable at all before Story
  7.5 lifted the cap
- **Status:** OPEN

**The gap.** After the lift, the only remaining ceiling on a content element's Y is
`MaxCanvasMillipoints`. Text line tops go through `canvasLineTop`'s JS-safe guard, but the **non-text**
branch of `addCanvasWindowCount` (`folio-go/page_setup.go`) builds `Bottom: element.Y + height` with a
**raw `+`**, and `folio-designer/src/engine-protocol.ts` requires only `Number.isSafeInteger` and `> 0`
of the resulting count. A component placed at the JavaScript-safe ceiling yields a count near 2.5e10 that
the browser guard admits and Story 7.6 would try to draw one sheet per window of.

**Why it is low, and why it is filed anyway.** Nothing else on the projection is uncapped, and no shipped
template comes near it. It is filed because Story 7.5 is what made the input reachable: the same shape as
DW-25's own lesson, where a bound that was safe only because nothing could reach it stopped being safe
when something could.

**This is the same axis as DW-26** (`style.fontSize` has no range check and is the other operand of the
one product that can overflow `geom.ScaleRound`). Whoever closes either should look at both: the question
"which canvas geometry inputs are bounded, and by what" has now been asked twice from two directions and
answered piecemeal each time.


---

### DW-38 — `createComponentCommand` hardcodes a 72x24 box for EVERY palette kind, so the same placement now yields a different box on a later sheet than on the first one — **NEEDS A RULING, NOT A PATCH**
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Owner** — this needs a ruling on *where per-kind placement defaults live* before anyone
  writes a line; due at **Epic 7's retrospective** at the latest, which is the gate that reviews what
  the epic claimed. Not "whoever touches the palette next" (D-000.73: an owner that is an event stops
  existing when the event passes).
- **Severity:** MEDIUM — silent, wrong-sized boxes on a path this story just opened to the author
- **Status:** OPEN

**The gap.** `folio-designer/src/component-command.ts:17` emits `"width":72,"height":24`
unconditionally, for every `PaletteKind`. Go's `dropComponent` path does not:
`folio-go/component_commands.go:1353-1360` declares `dropWidth/dropHeight` 72000/24000 but
`imageDropWidth/imageDropHeight` **96000/48000** for images and `lineDropHeight` **1000** for lines,
and applies them at `:1390-1395`. Story 7.6 put `createComponentCommand` on the **user path for the
first time** — before this story it was imported by tests only, and `App.tsx` did not import it — so
an image placed on sheet one gets a 96x48 box and the *same* image placed on sheet three gets 72x24.
The divergence is newly *visible*, not newly created, but the user-visible inconsistency is new.

**Why it was not patched here.** The fix has two shapes and the intent settles neither. Duplicating
Go's per-kind constants into TypeScript makes a **fourth spelling** of numbers the mirror discipline
exists to prevent drifting — and `engine-bounds-mirror.test.ts` freezes the numeral `pairs` at
**six** with an explicit `toHaveLength(6)`, so three more pairs is a deliberate widening of a fence
this story's own contract forbade it to touch. Moving the defaults into Go instead is a command-surface
decision the intent does not settle. Either way it is a ruling first.

**What the ruling must answer.** Whether per-kind placement defaults are the engine's (Go decides the
box, and `createComponent` stops carrying width/height at all) or the palette's (the browser decides,
and the numeral mirror widens to cover them). Whoever rules should note that the first shape removes
two numbers from the channel rather than adding three to the fence.

---

### DW-39 — a content component whose home window is past the drawing budget produces NO occurrence at all, so it is unreachable, and the disclosure never says components are hidden
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 7's retrospective, or the plan gate of the next story that changes the canvas
  drawing budget (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** MEDIUM in kind, LOW in reach — see below
- **Status:** OPEN

**The gap.** `folio-designer/src/sheet-stack.ts:114` iterates `index < drawn`, where `drawn` is
`Math.min(origins.length, MAX_CANVAS_SHEETS)`, while homes are computed over the **full** origins
list at `:112`. A component whose home index is `>= drawn` and whose extent intersects no drawn window
is therefore emitted nowhere: not selectable, not nameable, not deletable. The truncation disclosure
says only that sheets are truncated ("Showing the first N sheets of M"), never that **components** are
hidden — so the author is told the drawing is short and not told anything is missing from it.

**Why it is bounded in practice.** It requires a document of more than 120 windows, and Ruling J's own
derivation argues the paint budget (1920 body-text lines) cannot fill that many with text; past ~50
windows a window can only come from a *declared* placement gap. The budget test uses a component-free
projection, so nothing observes the case today.

**Adjacent to DW-34** (the canvas is unvirtualised), which `MAX_CANVAS_SHEETS` bounds rather than
closes. Whoever revisits the budget should decide both: what the budget is, and what the interface
owes an author whose component fell outside it.

---

### DW-40 — the floor flag's causes cover a bound TABLE but not a bound content TEXT element, so the one document that most needs the disclosure can be the one that does not get it
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 7's retrospective** — a projection-honesty scope decision, and the retrospective is
  the gate that audits what the epic claimed. A gate, never an event, per D-000.73.
- **Severity:** MEDIUM — a false *negative* on an honesty obligation is worse than a false positive
- **Status:** OPEN

**The gap.** `folio-go/page_setup.go:591` `canvasContentBandHasBoundTable` tests for a table with a
non-empty binding. The disclosure it drives generalises much further — *"A document whose length comes
from data prints more pages than are shown here."* A **bound content text element** whose bound value
is longer than its placeholder takes neither the table branch nor the font-chain branch, so such a
document under-counts its windows with `contentWindowCountIsFloor` **false**, and the interface
withholds the very sentence that describes it.

**Why it was not widened here.** Story 7.6's intent scopes the flag to **three named causes**, and
widening it is a projection-honesty decision about what the engine claims to know, not an
implementation choice. Recorded rather than taken, exactly as the contract required.

**Note for whoever rules.** The canvas breaks the **raw template string** with nil substitutions
(D-7.4.4), so it cannot measure a bound value's real length — which means the honest widening is
"a bound content text element exists" as a *floor* cause, not an attempt to count its lines.

---

### DW-41 — the origins refusal branch collapses a genuine multi-window document to one sheet with the floor flag set, and NO test reaches it
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 7's retrospective, or the plan gate of the next story that changes the window-origin
  projection (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** MEDIUM — the failure is indistinguishable from a legitimate degradation
- **Status:** OPEN

**The gap.** `folio-go/page_setup.go:565` `canvasWindowOrigins` refuses a `Shift` sequence that is
negative, above `MaxCanvasMillipoints`, non-zero at index 0, or non-increasing, and its caller then
degrades to one window with the floor flag set. Every fixture in `canvas_window_count_test.go`
produces a well-formed sequence, so the `!ok` path is **dead code as far as the suite knows**. Its
collapse is indistinguishable from the three legitimate floor causes, so a regression that made it
fire on ordinary documents would surface to the author as *"this document prints more pages than are
shown"* rather than as an error anyone could act on.

**Why the degradation shape is nonetheless right.** A sequence the browser's own validator would reject
must never be sent — `engine-protocol.ts` discards the WHOLE snapshot on a malformed projection and
blanks the canvas with no attribution. Discarding the number is cheaper than discarding the snapshot.
What is missing is a test that reaches the branch and, separately, a way to tell this collapse apart
from a real floor.

---

### DW-42 — `SHEET_STACK_GAP` and the `.sheet-stack` CSS gap are coupled by nothing executable, and every cross-seam drag is correct only while the laid-out pixel gap equals the constant
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 7's retrospective, or the plan gate of the next story that changes the sheet
  stack's display geometry (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** MEDIUM — **raised at close.** This is the SAME CLASS as the five mirrors Story 7.5's
  inventory already found, and Story 7.6's own high-severity defect came from exactly this seam.
- **Status:** OPEN

**The gap.** `folio-designer/src/sheet-stack.ts:40` declares `SHEET_STACK_GAP = 24` and `sheetPitch`
adds it to the page height to invert the stack's display geometry. `App.css:66` consumes it only
through the inline `--sheet-stack-gap` custom property `App.tsx` writes. `design-contract.test.ts`
reads `App.css` as text but asserts **nothing** about `.sheet-stack`, and jsdom applies no stylesheet,
so **changing or dropping the CSS gap keeps every test green and silently drifts every drag that
crosses a seam** by the difference. The only `.sheet-stack` assertion in the suite is that a
single-window canvas has none.

**Why this one deserves attention above its severity.** The repo already has the idiom for exactly
this tie — `engine-bounds-mirror.test.ts` reads sources as text and pins both the declaration and its
consumption site — and this story just added a third consumer to that mirror without adding this
fourth. The three surfaces where Story 7.6's correctness actually lives (the clipping, the seam
painting and real pointer travel) are CSS and layout, which jsdom cannot observe and the browser e2e
suite does not execute (D-000.4). A text-level mirror is the only executable proof available here.

---

### DW-43 — no shipped fixture draws an IN-SHEET seam, so the seam's rendering is covered only by hand-authored synthetic literals
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 7.7's plan gate** — 7.7 is keep-together across a page break, so a fixture whose
  column really breaks inside a sheet is its natural instrument. A gate, never an event, per D-000.73.
- **Severity:** LOW
- **Status:** OPEN

**The gap.** The `page-count-*` fixtures place elements a round 728000 apart in a 727890 window, so
`next - origin > contentWindowHeight` on **every** sheet and the no-marker branch is always taken. The
only coverage of the drawn seam is the hand-authored `prose` fixture in `sheet-stack.test.ts` and the
projection literals in `App.test.tsx`. Both are correct arithmetic over numbers no engine produced.

**Why it matters more than "low" suggests.** The 110-millipoint overshoot that makes these fixtures
such a sharp red proof for the *origins* (the closed form answers 727890 where the engine answers
728000) is the very thing that stops them exercising the *seam*. The two properties trade off in the
same fixture.

**Amended by Story 7.7 (2026-08-31), and left OPEN.** The instrument now exists:
`fixtures/keep-together/` is a shipped, four-target-registered document whose content column really
breaks inside a sheet — its signature block sits astride the first content window's ceiling
(729.890 pt) and the whole block moves to page 2, so the engine itself produces a break at a position
no `page-count-*` fixture reaches. What remains unanswered, and is why this entry is **not closed**,
is the half this story deliberately does not touch: **nobody has confirmed the canvas DRAWS that
seam.** The canvas builds its column items ungrouped (`page_setup.go`), so it does not preview
grouping at all, and Story 7.7 neither claims nor tests that its window count keeps the floor property
in the presence of a group. Whoever takes this entry should drive the seam's rendering from this
fixture rather than from a hand-authored synthetic literal.

---

### DW-44 — `MAX_ENGINE_CONTENT_WINDOWS` has no Go counterpart and no test on either side; exceeding it discards the whole snapshot and blanks the canvas with no attribution
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **The plan gate of the next story that changes canvas geometry bounds, or Epic 7's
  retrospective (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** LOW
- **Status:** OPEN

**The gap.** `folio-designer/src/engine-protocol.ts:25` declares `MAX_ENGINE_CONTENT_WINDOWS = 100_000`
and `:220` rejects an origins array longer than it. Go declares no maximum window count, which is the
stated ground for excluding this constant from the mirror's six numeral pairs — and is exactly why
nothing bounds the producible value. A projection that exceeded it would be rejected by `parseInbound`,
which discards the **entire** snapshot and renders "Waiting for Go page geometry" with no attribution.

**This is the same axis as DW-37 and DW-26.** "Which canvas geometry inputs are bounded, and by what"
has now been asked from three directions and answered piecemeal each time. Whoever closes any one of
the three should close all three, or say why not.

---

### DW-45 — the sheet stack is rebuilt on every render, including every pointermove during a drag, at O(sheets x components) with up to 120 sheets
- **Deferred by:** Story 7.6's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 7's retrospective, or the gate that closes DW-34 (whichever first)** — a gate,
  never an event, per D-000.73
- **Severity:** LOW
- **Status:** OPEN

**The gap.** `folio-designer/src/App.tsx:712` computes `const stack = canvas ? sheetStack(canvas) : undefined`
with no `useMemo`, and a drag calls `setDrag` on every pointermove. `stackYForColumn` additionally
allocates a fresh origins array per call, and `columnEdgeAfterDrag` calls it twice per move.

**Adjacent to DW-34** (the canvas is unvirtualised), which Story 7.6 bounds with `MAX_CANVAS_SHEETS`
and leaves open. The sheet count multiplies DW-34's cost, and this entry is the drag-time half of the
same problem. Filed as low because the arithmetic is cheap per sheet and the budget caps the product;
filed at all because the drag is the one interaction where a per-frame cost is felt.

---

### DW-46 — for a grouped document the canvas reports an EXACT window count that is wrong, and window ORIGINS that are wrong in a way no floor flag discloses
- **Deferred by:** Story 7.7's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 7.9** — *the canvas tells the truth about keep-together groups*. Ruled by the
  engineering lead 2026-08-31 (D-7.7.6, D-7.7.7, D-7.7.8): this is a **defect, not a shortfall**,
  because `keepTogetherTags` takes the Template and nothing else, so grouping is a pure template
  property the canvas already holds every input for. **No fourth floor cause is registered** — the
  canvas's `ColumnItem`s are tagged with the same groups the render path uses, and the true origins
  fall out of `Paginate` because 7.6 projected them from `Shift` rather than computing them
- **Severity:** HIGH
- **Status:** **CLOSED** 2026-08-31 by Story 7.9. Both canvas column-item arms carry the document's
  author-declared groups, taken from the render path's own `keepTogetherTags` index, and no fourth
  cause was registered. Verified at the close by independent measurement rather than from the build's
  report: for the grouped document the canvas answers **3 windows, origins `[0 700000 1440000]`**
  against a real `buildPageModel` render of **3 pages** with the identical origins; for its untagged
  twin, 2 and 2. Each arm was red-proved ALONE — neutralising the non-text arm gives "the canvas
  counts 2 windows and the render path 3", neutralising the text arm gives "the shipped pair begins
  its windows identically at `[0 734000]`". The vacuous oracle that let this ship
  (`renderPathWindows` passing `nil` for the keep-together index) was repaired and the repair
  red-proved: reverting that one argument reddens with canvas 3 against the crippled oracle's 2.
  Epic 7's boundary is released on this entry

**Measured at the close, with a control.** A document with a body element, a two-member `signature`
group at y 700 / y 740 and an untagged tail at y 1440 renders **3 pages**, while
`CanvasWithTextPaint` reports `ContentWindowCount = 2`, `ContentWindowCountIsFloor = **false**` and
`ContentWindowOrigins [0, 740000]`. The render's second window begins at the group's earliest top,
**700000**. The same document with the tags removed renders 2 pages and the canvas is right — that
control is what makes the failure the grouping's rather than the fixture's.

**Why this is worse than an inaccuracy.** `addCanvasWindowCount` (`folio-go/page_setup.go:627-702`)
builds its `layout.ColumnItems` with no `Group` at all, and grouping is not among the floor causes
computed at `:635`. So this is a **fourth floor cause Story 7.6 does not know about** — and 7.6
exists precisely to make the canvas honest about what it cannot know. A disclosure that is
confidently wrong (`IsFloor = false`, i.e. *this count is exact*) is a liability the pre-7.6 silence
was not.

**The wrong ORIGINS are a separate failure.** Story 7.6 projected origins so the browser would never
compute them. A floor flag says the count may be low; it says nothing at all about an origin that
points at the wrong column position, and there is no flag on the origins array. Fixing the count
alone would leave this half standing.

No test reads either value for a tagged document. Story 7.7 could not take it: the intent's `Never`
clause forbids making the canvas group-aware, and the spec's "Limits to state" explicitly forbids
adding a floor cause or touching the projection. **DW-40 and DW-41 are the neighbouring honesty
questions**; whoever rules on this should read all three together.

---

### DW-47 — an over-tall SINGLE-member keep-together group is clipped and warned, where the same element untagged is a fatal error — the contract's own matrix rows 3 and 5 collide — **CLOSED by Story 7.10, 2026-08-31**
- **Deferred by:** Story 7.7's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 7.10** — *an over-tall element is refused whether or not it is grouped*.
  Ruled 2026-08-31 (D-7.7.9): the discriminator is **what** is over-tall, not **whether** it is
  grouped. An over-tall individual element is fatal tagged or not; a group over-tall only **in
  aggregate** takes 4.6's clip-and-warn. Does **not** gate `epic-7: done` — nothing lies to the
  author here, both renders are self-consistent with true diagnostics — but it **gates the
  `folio-go/v0.1.0` tag**, because it narrows what renders (AD-22)
- **Severity:** MEDIUM
- **Status:** **CLOSED** by Story 7.10's implementation commit `f85da21`, 2026-08-31, verified at the
  story's close. A tagged, individually over-tall element is now a located fatal `OverflowError`
  naming that element, identical to the untagged refusal for this declared-box population — measured
  at the close from a throwaway module outside the repository, both arms reading *"element e1: box is
  taller than the content window (900000mp against a content height of 729890mp)"*. Matrix row 5 was
  tightened rather than left false: a group of one is a no-op only where the group adds nothing, and
  the equality is asserted for the declared-box population **only**, quarantined by a comment
  forbidding its generalisation (D-7.10.3). The **adjacent** finding recorded below — the refusal
  calling a non-table a "table" — is closed too, at the derivation rather than by special-casing the
  string (D-7.10.5): `overflowKind` gives an element box its own word, and a test asserts the wording
  rather than trusting it. Red-proved at the close: restoring the pre-7.10 derivation reddens both
  `box` assertions and the unit test; deleting the discriminator reddens 5 distinct test functions
  with the aggregate arm green

**The collision.** The story's I/O matrix says "Group taller than one window → clipped with a
Warning" (row 3) and "Single-member group → placement identical to the same element untagged"
(row 5). A single over-tall tagged element satisfies both rows and they disagree: tagging converts
D-2.6.1's located fatal into a clip-and-warn.

**Reproduced at the close.** An over-tall `rect` element (height 900 pt against a 729.890 pt content
window) renders **566 bytes plus one `TABLE_ROW_CLIPPED_HEIGHT` Warning** when tagged, and is a
**fatal `RenderError`** when untagged. The shipped behaviour is defensible under D-4.6.2 as amended
— a single-member group is still an author-declared group, and leniency follows authorship — but
row 5's "identical to untagged" is false in this corner and should be tightened or the behaviour
changed.

**Adjacent, and pre-existing, found while reproducing the above:** the untagged fatal's message reads
*"element e1: **table** is taller than the content window"* for an element that is not a table. The
symbol is outside this story's diff and untouched by it; whoever takes this entry is standing at that
line anyway.

---

### DW-48 — `duplicateComponent` copies a keep-together tag into a group the designer offers no way to see or clear
- **Deferred by:** Story 7.7's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 7.9**, riding with DW-46. Ruled 2026-08-31 (D-7.7.10): a duplicated component
  joins **no** group; drop the tag on copy, asserted rather than incidental. Designer-side group
  **authoring** stays deliberately out of Epic 7 — a stated scope boundary, not an accident — but
  creating state the author cannot reach or undo is refused, as it is everywhere else in this project
- **Severity:** MEDIUM
- **Status:** **CLOSED** 2026-08-31 by Story 7.9 — see the closing note below

**The gap.** `duplicateComponent`'s `clone := *element` copies the whole `Element`, `KeepTogether`
included. The designer has no grouping concept at all — `component-command.ts` has zero hits and
`component_commands.go`'s property surface does not accept the key — so a group can only be authored
by hand-editing the `.folio` file, yet duplicating a tagged signature element silently joins the copy
to that group, where it then constrains pagination invisibly.

Story 7.7 records the canvas **preview** limit (DW-46); this **authoring/duplication** limit is
recorded nowhere else. `unbreakableValues` is the shipped precedent for a format key with zero
designer references, so the absence of a control is not itself the defect — the silent, unclearable
join on duplicate is.

**Closed** 2026-08-31 by Story 7.9. `duplicateComponent` clears the copy's tag to the **zero
`Presence`**, never to an explicit null — `Set: true, Null: true` serializes back as
`"keepTogether": null`, which is still the key in the file and still raises the required format
version. Both halves are asserted directly: the copy carries no tag (`!Set && !Null && Value == ""`)
**and** the original element is unchanged. Red-proved at the close by removing the clear, which gives
"the duplicate carries keepTogether `Presence[string]{Set:true, Null:false, Value:\"signature\"}`".

---

### DW-49 — `ARCHITECTURE-SPINE.md` still scopes the over-tall clip carve-out to "rows" although a second population is now clipped — **FULLY CLOSED 2026-08-31** (half (a) Story 7.9, half (b) Story 7.10)
- **Deferred by:** Story 7.7's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **SPLIT 2026-08-31 (D-7.7.13), two edits with two homes, neither held for the other.**
  **(a)** Widening the carve-out from "rows" to rows **and author-declared groups** describes HEAD
  and has been stale since `ed485eb` — it lands with **Story 7.9**, as bookkeeping inside that
  story's record per D-000.6. **(b)** The discriminator clause — an individually over-tall element
  is fatal regardless of tagging — describes behaviour that does not exist yet and rides **Story
  7.10**. A spine running ahead of the code is the same defect as one lagging it
- **Severity:** LOW
- **Status:** **HALF (a) CLOSED** 2026-08-31 by Story 7.9; **half (b) still OPEN**, owned by Story
  7.10. Verified at the close: AD-14's Rule bullet now reads "Rows and author-declared keep-together
  groups too tall for one content window (FR25, FR51), and clipped content (FR44), are `Warning`s
  …", `FR51` was added to the `Binds:` line, the file is **still 722 lines** (so
  `internal/diag/diag.go`'s line citation is no worse — DW-65), and the **discriminator clause is
  confirmed ABSENT**: zero hits for "discriminator", "individually over-tall" or "regardless of
  tagging" anywhere in the file. A spine running ahead of the code is the same defect as one lagging
  it, so half (b) was deliberately not written
- **Half (b) CLOSED** by Story 7.10's implementation commit `f85da21`, 2026-08-31, in the **same
  commit as the behaviour** it describes (D-7.7.14). AD-14's Rule bullet now carries the
  discriminator clause: the leniency *"is scoped by **what is over-tall**, never by whether an item
  is grouped"*, a keep-together group being a `Warning` only **in aggregate** and a located `Error`
  when it holds a template element taller than a window, while a row stays a `Warning` whatever its
  shape. Appended inside the existing Rule bullet as this entry prescribed, so
  `table_row_clip_test.go`'s verbatim quotation of AD-14 is not falsified. Verified at the close: the
  `<!-- stage-rank-table:begin/end -->` markers are untouched at `:77`/`:91`, the spine still parses,
  and `cd lint && go test -count=1 ./...` is 4 packages `ok`. **The file is now 728 lines, not 722**
  — which no longer matters, because DW-65 closed the line citation in the same commit. **This entry
  is now fully closed**

**The gap.** `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md:319`
still reads "Over-tall **rows** (FR25) and clipped content (FR44) are …". D-4.6.2 was amended in
`folio-mvp-decision-log.md` for this story and quotes the spine as the authority for that noun, so
the two now disagree by one word. Verified still present at the close.

Cheap to fix and easy to leave rotting; filed so the next reader of D-4.6.2 does not conclude the
amendment was unauthorised.

---

### DW-50 — a tagged MULTI-LINE text element becomes atomic, so the matrix's "a single-member group changes nothing" holds only for single-line elements — **CLOSED by Story 7.10, 2026-08-31**
- **Deferred by:** Story 7.7's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 7.10**, with DW-47 — same contract row, second element kind, one two-arm
  fixture. Ruled 2026-08-31 (D-7.7.9)
- **Severity:** LOW
- **Status:** **CLOSED** by Story 7.10's implementation commit `f85da21`, 2026-08-31.
  **D-7.7.9 as written did not reach this entry, and that is the substance of the close.** Its
  mechanism — *every member fitting, the sum not* — reasons in ELEMENTS, while both pagination passes
  emit one column item per shaped LINE; read in items, every member of a tagged paragraph fits and
  only the union does not, so this document classified itself as merely aggregate and would have been
  left exactly as it was. The plan gate halted on that rather than guessing, and **D-7.10.1 ruled the
  member unit to be the TEMPLATE ELEMENT**, which is what reaches this case. A tagged multi-line text
  element is now refused, naming the element, with its OWN union extent reported rather than one
  line's. Red-proved at the close: regressing the member unit back to the column item reddens the
  text arm and the mixed-group arm **alone**, while the declared-box and tagged-image arms stay green
  — this entry reproduced exactly, under a suite that would otherwise have been green

**The gap.** `contentColumnItems` emits one `ColumnItem` per shaped line and the substitution stamps
the same key on all of them, so tagging one multi-line element makes its lines unbreakable and its
placement can change. `TestKeepTogetherSingleMemberChangesNothing` deliberately uses a single-line
element and its own doc comment concedes the multi-line case "would be a real change", while the
contract's row 5 is stated at the **element** level.

**Measured at the close, and the consequence is sharper than "placement can change":** a ~60-clause
text element that untagged flows cleanly across pages (71,374 bytes, **no diagnostic**) becomes,
tagged alone, a single over-tall group that is **clipped** — 66,636 bytes and one
`TABLE_ROW_CLIPPED_HEIGHT` Warning, with the overflow dropped from the document. So tagging a long
paragraph on its own does not merely move it; it can silently cost the author content. Arguably the
feature working as intended, which is exactly why it wants a ruling rather than a patch.

---

### DW-51 — no test combines non-contiguous membership with a union extent that crosses the window ceiling
- **Deferred by:** Story 7.7's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 7's retrospective, or the next story that touches `keep_together_fixture_test.go`
  (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** LOW
- **Status:** OPEN

**The gap.** `TestKeepTogetherMembersNeedNotBeContiguous` uses a group whose union **fits** window
one, so it exercises the ride-along but never the slide. The shipped fixture crosses the ceiling but
is contiguous. The case where the window must slide to the group's earliest `Top` **across an
intervening ungrouped item** — the two constraints together — is exercised by nothing.

Both constraints are individually covered and the shipped machinery is key-agnostic, so this is a
coverage gap rather than a suspected defect. Filed because the two-constraint case is the one a
future change to the slide would break first.

---

### DW-52 — `asLoadError` uses a bare type assertion while its comment claims `errors.As` semantics, so it fails on any wrapped `LoadError`
- **Deferred by:** Story 7.7's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **the story that next changes error wrapping in `folio-go/internal/template`, or Story
  15.3 before the `folio-go/v0.1.0` tag (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** LOW
- **Status:** OPEN — **pre-existing; the symbol does not appear in Story 7.7's diff**

**The gap.** The body is `err.(*LoadError)`, not `errors.As`, yet the same package wraps with
`fmt.Errorf("template: %s: %w", …)` in `decodeBand` / `decodeElement`. Either use `errors.As` or
correct the comment.

**Why it matters more than it reads.** Story 7.7's two load-refusal tests assert through
`asLoadError`, and they pass — so today's refusal path is unwrapped. The day a caller in that chain
starts wrapping, those assertions do not become wrong, they become **unreachable**: the test fails
loudly rather than silently, which is the good direction, but the diagnostic surface a
`TEMPLATE_MALFORMED` consumer sees would already have changed.

---

### DW-53 — `cmd/folio` prints a load error's full text to a terminal, so terminal-escape content in a `.folio` reaches the terminal
- **Deferred by:** Story 7.8 (2026-08-31), which **filed rather than fixed** it by explicit
  instruction in its own intent contract
- **Owner:** **the story that next changes `folio-go/cmd/folio`'s error reporting, or Story 15.3
  before the `folio-go/v0.1.0` tag (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** LOW
- **Status:** OPEN — **pre-existing; this story neither created it nor fixed it**

**The gap.** `cmd/folio` writes `err.Error()` straight to a terminal. A `.folio` whose field values
carry ANSI escape sequences therefore reaches a terminal that interprets them. Nothing in Story 7.8
introduced this: the CLI printed the full message before the story and prints it after.

**What Story 7.8 changed, and what it deliberately did not.** D-7.8.5 bounds every author-supplied
fragment of the message *in runes* at render (`internal/template/errors.go`), so the volume a
hostile document can push through the CLI is now capped where it was previously unbounded. That is a
side effect, not a fix: **bounding is not escaping**, and a short escape sequence passes a length
bound untouched. The remedy is a sanitizing writer in the shell, which is `cmd/folio`'s concern and
not the format's.

**Why it is filed and not urgent.** The residual is the user's own file rendered on the user's own
terminal, with no server and no third party anywhere in the product — the same position PRD §13
records for MVP, which deliberately has no threat model. **FR45's REST service is what reopens it**,
exactly as it reopens the reflection question D-7.8.5 ruled on.

---

### DW-54 — a `.folio` already carrying a table `justify` is permanently unloadable with no migration path, and `folio-format.md` states no rule for NARROWING a closed set
- **Deferred by:** Story 7.8 (2026-08-31); found recorded only in the spec's frontmatter at the
  story's close and filed into this register there
- **Owner:** **the engineering lead**, before the next narrowing lands — a gate, never an event, per
  D-000.73. It is **live now**, not at the tag: D-7.8.3's before-the-tag set already contains **two
  more narrowings** (Story 7.10's over-tall element, and D-7.8.2's code audit), so the rule will be
  re-litigated twice more unless it is written once
- **Severity:** MEDIUM
- **Status:** OPEN — **caused by this change** (the first half), **pre-existing** (the second half)

**The gap, first half.** Between Story 7.3 (which admitted `justify`) and Story 7.8 (which refuses
it), the designer's own engine could author a `.folio` whose table carried
`style.align`/`headerStyle.align: "justify"` — through the `component_commands.go` align arm this
story closes. Any file written in that window now fails `ParseDocument` **forever**. There is no
migration path, no repair mode and no diagnostic that says "this used to be legal". The story's I/O
matrix mandates the refusal and its Never list excludes new designer work, so a migration path was
correctly out of scope *there* — but the population is real and nothing currently owns it.

**The gap, second half, and it is the more general one.** D-1.4.12 states that **extending** a
closed set is MAJOR. `folio-format.md` says **nothing** about **removing** a member from one, so the
version a narrowing lands under is unstated. A format document that has a rule for extending a
closed set and none for narrowing one will have the question re-opened at every narrowing, and the
answer will be argued from the story's own convenience each time. The remedy is one stated rule in
`folio-format.md`, not a per-story ruling.

**Why the lead and not a story.** Both halves are format-policy calls above any single story's
scope: whether the orphaned population is owed a migration, and what version a narrowing costs.
Neither is answerable from a story's intent contract, which is exactly why this sat in a spec
frontmatter rather than in a plan.

---

### DW-55 — `wrapTemplateError` passes `LoadError.ElementID` unbounded into the Diagnostic, where the wasm host byte-cuts it at 128 and splits a rune
- **Deferred by:** Story 7.8 (2026-08-31); found recorded only in the spec's frontmatter at the
  story's close and filed into this register there
- **Owner:** **the story that next changes `folio-go/render_error.go`'s diagnostic construction, or
  Story 15.3 before the `folio-go/v0.1.0` tag (whichever first)** — a gate, never an event, per
  D-000.73
- **Severity:** LOW
- **Status:** OPEN — **pre-existing; the `elementId` field was populated this way before this story
  and is unchanged by it**

**The gap.** `folio-go/render_error.go` passes `le.ElementID` straight through into the Diagnostic,
and `wasm/cmd/engine/main.go` applies `bounded(elementId, 128)` — a raw `value[:max]` byte slice. A
multi-byte element id is therefore split mid-rune in the `elementId` **field**, producing invalid
UTF-8 for the designer to render.

**Why it is worth a number despite being LOW.** This is the same *"runes, never bytes"* property
D-7.8.5 ruled on, one field over. Story 7.8 bounded the **message** — `LoadError.Error()` now cuts
only on rune boundaries with a visible marker, and the assembled sentence fits the host's window so
`bounded(message, 512)` can no longer fire. The `elementId` field never went through that seam, so
the host's byte cut is still the only bound on it. The ruling's own reasoning applies unchanged; only
its scope did not reach here.

---

### DW-56 — the `thai-stacked-marks` golden has no human reading sign-off, and its subject is a placement mechanism nothing in the corpus has ever judged
- **Deferred by:** Story 8.0 (2026-08-31), recorded in the spec's frontmatter at the build and filed
  into this register at the story's close
- **Owner:** **the human reader.** Not a story, not a gate, not a checklist — **no agent may write
  `reader`, `date` or `examined`**, and the owner has already been asked to perform the visual pass
  out of band. This entry is discharged only by a sign-off record a person writes
- **Severity:** **HIGH**
- **Status:** **CLOSED 2026-08-31.** The owner performed the visual pass and signed off:
  *"The rendering at fixtures/thai-stacked-marks/expected.pdf looks ok."* Recorded verbatim in
  `fixtures/thai-stacked-marks/signoff.json` with reader, date and the fixture's frozen digest, and
  enforced going forward by `folio-go/thai_stacked_marks_signoff_matrix_test.go`, a `//go:build
  matrix` red gate following both shipped precedents — red-proved by removing the record.
  **The record carries the reader's own words and nothing else.** No agent wrote an observation on
  their behalf; the schema asks for detail *"specific enough that a later reader can tell it apart
  from 'looked fine'"*, and this record does not yet meet that bar. That is stated here rather than
  repaired by invention, because inventing it is the one failure this whole mechanism exists to
  prevent (D-000.22: this is the one claim no machine can make on someone else's behalf). A reader
  may strengthen it at any time; the digest binds it either way.
  pass had not been performed when the story closed

**The gap.** `fixtures/thai-stacked-marks/expected.pdf` is the first committed document whose Thai
mark placement is produced by a **text-rise operator**. Every machine gate this story ran proves the
bytes are *stable and identical across four targets* — none of them proves the marks are *at the
right height on the page*. Only a person reading the rendered document can say that.

**Why the existing attestation does not cover it.** `fixtures/shaped-text/thai-signoff.json`
(D-2.3.5) attests **GSUB lowered-form placement at zero offset** — a different mechanism, reached by
a different code path, and one whose correctness this story does not touch. Reading it as covering
vertical displacement would be the same measure-one-report-wider error D-8.0.1 exists to stop, one
artifact over.

**Why it was routed as a deferral and not a halt.** The spec's `Block If` says to state a needed
human sign-off and stop. On three measured grounds it did not engage: **no existing attestation was
invalidated** (all 21 pre-existing digests are unmoved, so D-4.7.1's whole-file invalidation is not
reached), **no sign-off gate is left red**, and **no agent wrote an attestation field**. The
distinction that matters is between *a new artifact awaiting a reader* and *an existing attestation
broken by a change* — only the second is a halt.

---

### DW-57 — the shaped-expectation table's YOffset guard has an available red-proof that is not built, because the table is bound to a frozen oracle
- **Deferred by:** Story 8.0 (2026-08-31)
- **Owner:** **the story that next re-records
  `fixtures/shaped-text/harfbuzz-oracle.json`, or Story 8.4 (rendering from an embedded face,
  which is the first work that can introduce a fourth face) — whichever first** — a gate, never an
  event, per D-000.73
- **Severity:** LOW — **the false half is already corrected; only the missing proof remains**
- **Status:** OPEN — **pre-existing** (the claim predates this story) — **the comment half is
  CLOSED** (`folio-go/shaping_expectations_test.go`, corrected at Story 8.0's close, 2026-08-31)

**What was corrected, and what is left.** The comment above `shapedExpectations` asserted that
YOffset is *"a FORWARD GUARD WITH NO AVAILABLE RED-PROOF"*, measured as *"0 for every glyph of every
row across all three shipped faces"*, and instructed the next reader **not to manufacture a red-proof
for it**. This was the **fourth** surviving instance of the population error D-8.0.1 names — measured
over **this table's sixteen rows**, reported over **the shipped faces** — and Story 8.0 disproves the
wider half outright: the shipped Noto Sans Thai gives `ั`+tone a y-displacement of −57. The comment
now scopes the measurement to the rows it was taken over, quotes the false claim rather than deleting
it, and names the red-proof as available-but-unbuilt.

**Why the proof itself was out of scope.** Red-proving YOffset means adding a `ทั้ง` row, and every
row of that table is bound to `fixtures/shaped-text/harfbuzz-oracle.json`. Re-recording the oracle is
a separate obligation with its own evidentiary burden, and Story 8.0's contract fences it out. Until
it lands, all 16 rows carry YOffset 0, which also means `shaping_oracle_test.go`'s HarfBuzz `Dy`
cross-check compares 0 against 0 for every row — the guard exists and asserts, but on a population
where it cannot fail.

---

### DW-58 — a negative `fontSize` now silently INVERTS the text rise, where the pre-change engine refused the glyph outright
- **Deferred by:** Story 8.0 (2026-08-31)
- **Owner:** **the story that next changes `folio-go/parse.go`'s `decodePoints`, or Story 15.3
  before the `folio-go/v0.1.0` tag (whichever first)** — a gate, never an event, per D-000.73. It
  belongs at **parse-time validation**, not in the emitter
- **Severity:** MEDIUM
- **Status:** OPEN — **the exposure is caused by this change; the missing validation is
  pre-existing.** Explicitly **NOT a spec deviation**: the implementation follows the narrowing rule
  the intent contract states, exactly

**The gap.** Measured at Story 8.0's dispatch: rendering the new fixture with `"fontSize": -12` exits
0 and emits `+0.024`, `+0.684`, `+0.708` rises instead of the negative operands — the marks are
placed **above** the baseline instead of below. Before this story any non-zero vertical offset was
refused outright, so a hard refusal has become silently wrong output.

**Why the emitter is the wrong place to fix it.** The narrowed refusal fires only when the offset is
non-zero **and** the rise rounds to zero; a negative font size produces a perfectly non-zero rise, so
the emitter has nothing to catch. Making the emitter reject a negative rise would put a document-level
validity rule inside the byte-emission stage, where it can neither name the offending element nor
reach the author in their own terms. `fontSize` has **no positivity floor at parse** — the same
absence that lets `0.008` through, which is what makes the narrowed refusal reachable at all. One
floor at parse closes both.

---

### DW-59 — a very large `fontSize` panics during wrapping instead of returning a located error
- **Deferred by:** Story 8.0 (2026-08-31)
- **Owner:** **the story that next changes `folio-go/wrap.go`'s `measureRuneRange`, or Story 15.3
  before the `folio-go/v0.1.0` tag (whichever first)** — a gate, never an event, per D-000.73
- **Severity:** MEDIUM
- **Status:** OPEN — **pre-existing and not caused by this change.** The panic is reached during
  wrapping, **before** the new emitter code runs

**The gap.** Measured at Story 8.0's dispatch: `fontSize` 9223372036854 panics with
`geom: ScaleRound: v*num overflows int64` at `folio-go/wrap.go:128`. It is the same defect class
`line_spacing_test.go:93` already names — **a panic must be a returned error** — and it shares
DW-58's root cause: `decodePoints` applies no bound to `fontSize` in either direction. A single
parse-time range check discharges both, which is why they are filed adjacently rather than merged:
the *fixes* coincide, the *symptoms* and severities do not.

---

### DW-60 — `internal/text/shape.go` carries the same falsified YOffset unreachability claim, and here the red-proof IS available
- **Deferred by:** Story 8.0 (2026-08-31)
- **Owner:** **the story that next changes `folio-go/internal/text/shape.go`, or Story 8.4
  (rendering from an embedded face) — whichever first** — a gate, never an event, per D-000.73
- **Severity:** MEDIUM — higher than DW-57 because this is **production source**, not a test comment,
  and because the proof is available with no blocker
- **Status:** **CLOSED** 2026-08-31 by the orchestrator, both halves. The comment is corrected in
  `internal/text/shape.go`, and the red-proof the entry called available is now written:
  `internal/text/shape_shipped_face_test.go` shapes ทั้ through the shipped Noto Sans Thai and finds
  a displaced glyph, with ที่ as a control that comes back at zero because the face resolves that
  pair by a GSUB lowered-form substitution. The two **mutually discriminate**: swapping the strings
  reddens both, so together they measure the substitution-versus-displacement distinction rather
  than "Thai". It is an EXTERNAL test package, reading the face from its own file rather than
  importing `fonts`, because AD-8 keeps font data outside `internal/` and `fonts` imports the root
  package.

**The gap.** `shape.go:18-21` states as measured fact that YOffset is *"zero for every glyph of every
sample across all three shipped faces today"* and is a forward guard with no available red-proof.
Story 8.0 disproves it: a reviewer confirmed by mutation that **zeroing YOffset in `shape.go` now
reddens five tests**. This is a **third site** of the D-8.0.1 error beyond the two the spec's own
frontmatter named, and unlike DW-57 nothing blocks the proof — the reddening tests already exist.

**Why it is worth a number rather than a comment fix folded into Story 8.0.** Story 8.0's contract
fences `internal/text` out, and the honest fix here is comment **and** red-proof together: a comment
correction alone would trade a false claim for an unproved one in production source, which is the
weaker of the two available outcomes when the proof is right there. D-8.0.1's lesson is that a
comment asserting a negative carries a test's evidentiary burden.

---

### DW-61 — the golden count 21 is hard-coded as a literal at six sites with nothing binding it to the record's length
- **Deferred by:** Story 8.0 (2026-08-31)
- **Owner:** **the story that next adds a golden to `goldenDigestRecord`** — a gate, never an event,
  per D-000.73. That is the exact moment every one of these literals becomes false
- **Severity:** LOW
- **Status:** OPEN — **pre-existing in kind, and this story added to the population** rather than
  creating it

**The gap.** The number 21 appears as a bare literal in `thai_stacked_marks_template.go`,
`byte_neutrality_test.go` (two sites), `thai_stacked_marks_fixture_test.go`,
`internal/pdf/textdoc_test.go` and the new fixture's README, with no assertion tying any of them to
`len(goldenDigestRecord)`. A 23rd golden makes each of them a false statement and **nothing goes
red** — the failure mode this project has repeatedly recorded as the expensive one, because a stale
count reads as a verified count. The remedy is one derivation, not six edits.

---

### DW-62 — an element whose visibility depends on data makes the canvas's window count wrong in EITHER direction, and it has been undisclosed since Story 7.5
- **Deferred by:** Story 7.9 (2026-08-31), found by measurement while tagging the canvas's column
  items; the halt it caused was correct, and the ruling that resolved it is D-7.9.1
- **Owner:** **Owner** — the project owner's call. Every remedy changes what the design view *is*:
  either the canvas is given data to evaluate against (which makes the preview a function of a data
  set rather than of the template), or the divergence stays permanently disclosed. Neither is a
  builder's decision, and the honesty half already shipped, so nothing is blocked on the answer
- **Severity:** MEDIUM
- **Status:** OPEN — **the DISCLOSURE is closed; the DIVERGENCE is not**

**This entry is NOT about grouping.** It is about **an element whose visibility depends on data**,
and it must not be filed under keep-together or a future reader will close it by fixing grouping.
`page_setup.go` only *projects* `visibleIf`, as a string; nothing on the canvas path evaluates it,
because evaluating it needs the data the canvas has never been given. The canvas therefore places
the element and the render may omit it, and AD-24 makes a hidden element **absent with no gap** — no
sibling moves up, the column is simply shorter, and the two counts differ.

**An UNGROUPED element carrying `visibleIf` has the identical problem.** Grouping is how this was
found, not what caused it: a conditional member makes a group's whole slide conditional, which is
loud enough to measure, but the divergence needs no group at all. It has therefore been live and
**UNDISCLOSED SINCE STORY 7.5** shipped the window count, and since 7.6 shipped the flag that was
supposed to be honest about exactly this.

**Measured** at `6a31a7f` on a group whose rect member carries `"visibleIf": "showRule"`: the canvas
answers **3 windows with no data at all**, while the real render answers **3** pages for
`{"showRule": true}` and **2** for `{"showRule": false}`. Canvas >= render — a **ceiling**, which is
why a boolean named `IsFloor` could not survive contact with it (D-7.9.6).

**What Story 7.9 did close.** It is now registered as cause (d) of `ContentWindowCountIsExact`, so a
document carrying one no longer claims an exact count. The disclosure is deliberately conservative:
ANY content-band `visibleIf` marks the count untrustworthy, including where the element could not
move a window boundary. That over-disclosure is intended — the flag must fail toward the honest
claim — and is not the defect this entry tracks.

---

### DW-63 — a content-band TEXT element declaring a background or border makes the canvas UNDER-count, and closing it ADDS a canvas item source rather than removing one
- **Deferred by:** Story 7.9's review (2026-08-31), measured, and left deliberately unclosed
- **Owner:** **the next story that changes what the canvas contributes as a content-column item** — a
  gate, never an event, per D-000.73. That is the exact moment a second item source becomes cheap and
  the four-row matrix has to be re-measured anyway
- **Severity:** MEDIUM
- **Status:** OPEN — **pre-existing in kind; Story 7.9 corrected the comment that overstated it**

**The gap.** `collectElementBoxRects` is eligible for four kinds *including text*, and
`buildPageModel` folds its output into the slice `contentColumnItems` turns into content-column
items. So a **styled** text element contributes a full declared-box column item on the render path,
while the canvas contributes only its shaped lines. **Measured:** a content band of a text at y 0
plus a styled text element at y 700 with height 200 and a short value projects canvas count **1**,
origins `[0]`, exact **TRUE**, against a real `buildPageModel` render of **2 pages**.

**Why Story 7.9 did not close it, stated so the omission is not read as an oversight.** RULING B's
subject is the unstyled NON-TEXT element — a **spurious** canvas item to remove. This is the mirror
case: closing it means **adding a NEW canvas column-item source**, which is the kind of growth the
spec's own `story materially larger than the ruling implies` Block If guards. Every text element in
every canvas window-count fixture is unstyled, so nothing reddens today. What Story 7.9 *did* fix is
the comment in `addCanvasWindowCount` that asserted the parity invariant unconditionally; the claim
is now scoped to non-text components with this exception written out, so nothing in-tree overstates
it.

---

### DW-64 — three sites still paraphrase or quote AD-14's SUPERSEDED over-tall wording, one of them as "not yet built" — **CLOSED by Story 7.10, 2026-08-31**
- **Deferred by:** Story 7.9 (2026-08-31); first surfaced in its Dispatch 2 review and carried
  unfixed through Dispatch 3
- **Owner:** **Story 7.10** — *an over-tall element is refused whether or not it is grouped*. It owns
  DW-49's half (b) and rewrites this exact clause of AD-14 anyway, so repairing the paraphrase there
  costs one line, and repairing it anywhere else risks two authorities disagreeing again
- **Severity:** LOW
- **Status:** **CLOSED** by Story 7.10's implementation commit `f85da21`, 2026-08-31 — the story that
  lifted the `internal/layout` fence, exactly the cheap moment this entry was filed to wait for. All
  three listed sites re-pointed, **plus a FOURTH the entry did not list**, found by the story's own
  review at `paginate.go:851` — four lines above the very narrowing that superseded it. Verified
  mechanically at the close: zero hits for `"over-tall rows (FR25)"` anywhere under
  `folio-go/internal/layout/`, and zero hits for `"not yet built"` in `folio-go/diagnostic.go`. The
  re-pointed sites now paraphrase AD-14 **by number rather than by quotation**, which is what stops a
  fifth copy accruing the next time the spine is amended

**The gap.** `folio-go/diagnostic.go:344` paraphrases AD-14 as "over-tall rows (FR25, not yet
built)". Story 7.7 **built** them at `ed485eb`, and Story 7.9 widened the spine clause this comment
paraphrases to "Rows and author-declared keep-together groups too tall for one content window (FR25,
FR51)". Both halves of the paraphrase are now wrong: the population is no longer only rows, and the
"not yet built" half has been false for two stories. A comment that describes shipped behaviour as
unbuilt reads as a verified statement about scope, which is the expensive failure mode this register
keeps recording.

**Two further sites, and why they are HERE rather than triaged away.** Story 7.9's review raised the
superseded AD-14 wording surviving inside `folio-go/internal/layout/`, and the build **rejected** it
— correctly, because that story's contract fenced **zero paths** under that directory, absolutely.
A scope-correct rejection is not a finding that stopped being true, so it is recorded here rather
than lost with the pass that refused it. Verified still present at Story 7.9's close:

- `internal/layout/paginate.go:224` — quotes AD-14 as *"over-tall rows (FR25) and …"*, and `:37`
  paraphrases the same clause;
- `internal/layout/paginate_group_test.go:331` — quotes the identical superseded sentence.

Both now describe a narrower population than the engine has, and both sit in a directory that the
story which widened the spine was forbidden to touch. Whoever takes DW-49's half (b) will be editing
that clause with the fence lifted, which is the only cheap moment to bring all three sites into
line at once.

---

### DW-65 — `folio-go/internal/diag/diag.go` cites `ARCHITECTURE-SPINE.md` by LINE NUMBER, and that line falls inside a mermaid ER diagram — **CLOSED by Story 7.10, 2026-08-31**
- **Deferred by:** Story 7.9 (2026-08-31); noted at the plan gate as already stale **before** the
  story touched anything, and carried by the spec's Design Notes rather than repaired
- **Owner:** **the next story that edits `ARCHITECTURE-SPINE.md`** — a gate, never an event, per
  D-000.73. A line citation into a file whose length nobody is holding constant will rot again the
  moment anyone forgets, so the fix is to cite the **AD number**, not to re-count the lines
- **Severity:** LOW
- **Status:** **CLOSED** by Story 7.10's implementation commit `f85da21`, 2026-08-31. The gate fired
  as specified — this is the next story that edited `ARCHITECTURE-SPINE.md`, for DW-49's half (b) —
  and the fix is the prescribed one: cite the **AD number**, never re-count the lines. `diag.go:10`
  now reads *"see `ARCHITECTURE-SPINE.md`'s AD-14, 'Errors and diagnostics are one type on one
  channel', as amended by Story 2.8's AC11"*, with the reason for citing by number recorded inline.
  Verified at the close: **zero** `ARCHITECTURE-SPINE.md:<line>` citations remain anywhere under
  `folio-go/` or `lint/`. This story then grew the spine from 722 to **728** lines — which is exactly
  the silent drift this entry predicted, and it is now harmless

**The gap.** `folio-go/internal/diag/diag.go:10` cites `ARCHITECTURE-SPINE.md:613`, and at HEAD line
613 falls inside a mermaid ER diagram — the citation was already pointing at the wrong thing before
this story began. Story 7.9's AD-14 widening deliberately **rewrapped** rather than grew, holding the
file at **722 lines** (verified at the close), so the citation is exactly as stale as it was and no
staler. That was the right call for a story fenced to FR51, and it is also why this entry exists: the
next edit to that file will not have the same reason to hold the line count, and a silent one-line
drift is all it takes.

---

### DW-66 — which over-tall member the refusal NAMES is decided by internal item order, not by the author's declaration order
- **Deferred by:** Story 7.10's build (2026-08-31); confirmed and re-measured at the story's close
- **Owner:** **Story 7.11 or the first story that touches `overTallGroupMember`** — and it wants a
  **ruling before a patch**, because naming a preferred subject is a product decision, not a
  builder's choice. Does **not** gate the `folio-go/v0.1.0` tag: the refusal is located, truthful and
  deterministic in every case, so nothing lies to the author. It narrows nothing and can land after
- **Severity:** MEDIUM
- **Status:** OPEN

**The gap.** When two or more template elements of one author-declared group are each individually
over-tall, `overTallGroupMember` (`folio-go/internal/layout/paginate.go`) returns the **first** such
element in `items` order. Package `folio` builds that slice by appending rects, then text, then
images, so an over-tall text element is named ahead of an over-tall rect **regardless of which the
author declared first**. Measured during the story's review: the text is named in both declaration
orders.

**The determinism is real and is not the complaint.** Members are accumulated into a slice in
first-appearance order and scanned in that order; no map is ranged, so the answer is a function of
the caller's item order rather than of a hash seed (R5 / D-1.3.5), and the doc comment says so. The
complaint is about the **subject** that determinism is anchored to: it is an implementation detail of
how the paginator assembles its column, and it reaches a diagnostic message an author reads.

**This is the same class of defect as DW-50**, which this very story exists to fix — a product
behaviour decided by an internal decomposition. It is filed rather than patched because the fix
requires deciding *which* member ought to be named (the author's first declaration? the tallest? all
of them?), and that is a ruling.

---

### DW-67 — the canvas degrades to an inexact window count for a document the renderer now refuses, and nothing asserts that pairing
- **Deferred by:** Story 7.10's build (2026-08-31); confirmed at the story's close
- **Owner:** **Story 7.11, with DW-62** — read the two together. DW-62 is the canvas's *other*
  undisclosed count divergence and its subject is adjacent; a single story that pins
  *renderer refuses ⇒ canvas reports inexact* alongside DW-62's `visibleIf` case is one fixture's
  worth of work, where two stories would build the same harness twice
- **Severity:** MEDIUM
- **Status:** OPEN — **the behaviour is CORRECT at HEAD; what is missing is the assertion**

**The gap.** `folio-go/page_setup.go` calls `layout.Paginate` directly and degrades on any error,
setting `ContentWindowCountIsExact` false. Because Story 7.10 put the discriminator **inside**
`Paginate` rather than in a pre-pass, a tagged over-tall element reaches the canvas automatically —
which is exactly the behaviour Story 7.9 closed the canvas/render divergence for, and it is right.
But no test pins the pairing, so a future change that moves the discriminator out of `Paginate`, or
that adds a second refusal path around it, could re-split the renderer from the canvas silently. The
canvas would then confidently draw a clipped layout for a document the renderer refuses.

Filed with DW-62 rather than folded into Story 7.10 because adding it means building a canvas
fixture for a document that produces no bytes, which is a fixture shape neither this story nor Story
7.9 needed.

---

### DW-68 — the AGGREGATE-ONLY over-tall group stays clipped on a REASON its own story disowned, and reopening it is cheap only before `folio-go/v0.1.0`
- **Deferred by:** Story 7.10 (2026-08-31), on D-7.10.4's explicit instruction to record the
  reopening condition **prominently rather than as a buried note**
- **Owner:** **Owner / engineering lead — a RULING, before the `folio-go/v0.1.0` tag.** Not a builder's
  call and not a defect report. It is filed here so that it is findable at the moment it becomes
  expensive, which is the moment the tag is cut
- **Severity:** MEDIUM — **low as a defect, HIGH as a deadline**
- **Status:** OPEN — **deliberately, and this entry exists so that "deliberately" stays honest**

**⚠ THE DEADLINE IS THE POINT OF THIS ENTRY.** Making this case fatal **narrows what renders**. Under
D-7.7.13, D-7.8.3 and AD-22 that is free before the `folio-go/v0.1.0` tag and ruinous after, when
every downstream suite pays for it. Whoever is about to cut that tag should read this entry first and
decide, rather than discover it afterwards.

**The gap.** A keep-together group whose template elements each fit a content window, but whose union
does not, is still clipped to a page of its own with a `TABLE_ROW_CLIPPED_HEIGHT` Warning — Story
4.6's treatment, shipped by Story 7.7 and untouched by Story 7.10.

**Why it is on the register even though it is working as shipped.** Story 7.10's whole ratio is
**fixability**: an author-created grouping is strict because the author typed the tag and can remove
it, while an engine-created grouping (a table row) is lenient because a row's height comes from data
the author cannot audit. Pushed all the way, that argument makes the aggregate-only case fatal too —
the author can untag it just the same. **Story 7.7 chose clip by importing Story 4.6's TREATMENT
without importing its REASON**, and the lead confirmed that at the time. D-7.10.4 left it as it
stands because it is shipped, deliberate and outside 7.10's subject — **not because it is obviously
right**, and it declined to dress it up as anything else.

**What reopens it:** a real document losing content that way, as DW-50 came from a real case. One
ruling would cover both halves of the question. **If anyone sees such a document, raise it now.**


---

### DW-69 — a pre-existing `.folio` with a chain over 64 entries, or a face name over 512 bytes, now fails to OPEN rather than merely to edit
- **Deferred by:** Story 8.1's review pass (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 8.3** — the epic's format-change story under D-R7.9, so a rule for a chain the
  format permits but the projection refuses belongs with it; **and the engineering lead** if the
  before-the-tag question below is disputed. A story and a role, never an event, per D-000.73
- **Severity:** MEDIUM — it narrows what the designer will open, on a bound this story invented
- **Status:** OPEN

**The gap.** `canvasFontChains`' two new refusals — entry count at `folio-go/page_setup.go:465`, face
length at `:469` — run inside `Canvas`, which `CanvasWithTextPaint` wraps and `Engine.load`
(`folio-go/wasm/engine.go:119`) calls. `decodeFonts` (`folio-go/internal/template/parse.go:313`) bounds
neither entry count nor face length, and `render.go`'s chain resolution never counted entries, so such a
document parsed and rendered before this story. `TestCanvasFontChainEntryCountIsBoundedOnALoadedDocument`
measures exactly this: the loader accepts a 65-entry chain and the projection refuses it.

**The bound itself is not the defect.** `maxCanvasFontChainEntries = 64` was directed by the spec's
Task 6 and mirrors the pre-existing `canvasFontFamilies` shape. What is unrecorded is the
**compatibility narrowing**, and it has no matrix row. Nothing in the repo declares more than three
entries, so nothing existing is affected — but an embedding story may want the number revisited.

**IT DOES NOT JOIN D-7.8.3's BEFORE-THE-TAG SET, and the measurement is the reason.** `Canvas` is
reachable only from the projection surface (`page_setup.go`, `component_commands.go`) and the wasm
`Engine`; **`cmd/folio` never calls it**, verified by grep at this closing revision. So the
`folio-go/v0.1.0` renderer accepts exactly what it accepted before, and the tag freezes nothing this
narrows — unlike Story 7.8's load refusal or Story 7.10's over-tall fatality, which both narrow the
renderer itself. The set therefore still holds **one** open item (D-7.8.2's code audit). This is
recorded rather than ruled: **membership of that set is the lead's call**, and this entry exists so the
question is asked rather than assumed either way.

---

### DW-70 — Go sorts projected chain names by BYTES while the browser guard checks ascending UTF-16 code units, and the two disagree on astral-plane names — **CLOSED by Story 8.2, 2026-08-31**
- **Deferred by:** Story 8.1's review pass (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 8.2** — it is the story that first lets an author type a chain name, which is what
  makes the divergence reachable through the product
- **Severity:** MEDIUM once 8.2 lands; **latent today**
- **Status:** **CLOSED** by Story 8.2's implementation commit, 2026-08-31 — the story that made it
  reachable is the story that closed it, which is the same rule Story 8.0 was decided under. The
  browser adopted Go's order; **Go's comparator was not touched.**

**How it was closed, and the guardrail that decided which side moved.** `engine-protocol.ts` now
compares the projected names with a `compareCodePoints` helper — code-point order, which for UTF-8 is
byte order — instead of `>=` on JavaScript strings. **Go's byte ordering is NORMATIVE and must stay
that way:** the `fonts` keys are sorted into the canonical `.folio` under AD-9, so Go's sort *is* the
byte order of the document, and changing the Go comparator would have moved golden bytes for any
document whose chain names cross the boundary. The cheaper-looking side was the wrong side.
`engine-protocol.test.ts` pins the measured pair `['\uE000', '\u{1F600}']` as ACCEPTED and its reverse
as still REJECTED, so the fix widened the accepted set rather than removing the check — proved in both
directions by restoring the UTF-16 comparator and by deleting the check outright.

**The gap.** `folio-designer/src/engine-protocol.ts`'s sorted/unique check compares names with `>=` on
JavaScript strings — ascending **UTF-16 code units** — while Go's `slices.Sorted(maps.Keys(...))` sorts
by **bytes**. The two orders disagree for names mixing astral-plane characters with U+E000–U+FFFF,
because a surrogate pair sorts below U+E000 in UTF-16 and above it in UTF-8. A disagreement makes
`isCanvas` false, which drops the whole snapshot and blanks the canvas with no attributable error.

**Pre-existing, and widened by one field.** The check predates this story and already applied to
`fontFamilies`; Story 8.1 makes it newly reachable through the Go command API, but the designer sends
no chain command until 8.2, so it is not reachable through the product yet. This is the same shape as
DW-44: a browser-side bound with no Go counterpart, where exceeding it discards everything.

---

### DW-71 — no host-boundary test dispatches a font-chain command, so the refusal messages are specified at the wasm wire and measured one layer below it
- **Deferred by:** Story 8.1's review pass (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 8 close** — due before that epic's key is marked `done`, since every remaining story
  in it adds command surface behind the same untested boundary
- **Severity:** MEDIUM — the layer that formats what the author reads is the one layer nothing executes
- **Status:** OPEN

**The gap.** `folio-go/wasm/cmd/engine` is `//go:build js && wasm`, so `go test ./...` never compiles it
and no CI job executes its host-boundary assertions. The intent contract specifies the refusal path at
that wire — `*ComponentCommandError` matched **before** `*RenderError` at
`folio-go/wasm/cmd/engine/main.go:236`, emitted with `bounded(msg, 512)` and no `reportableMessage`
filtering — and every assertion behind it is made one layer down, against `ApplyComponentCommand`.

**The package's own sibling test states the hazard verbatim:** *"Every Go-side assertion would still
have been green."* The package's dormancy is a standing condition, not introduced here. What this story
added is the most a compiled test can currently reach: a source-reading tripwire
(`TestComponentFailureBoundsMatchTheHostsOwnLiterals`) that ties the two hand-copied bound literals by
reading the host file, because no compiled test can see a `js && wasm` package.

---

### DW-72 — integer-index refusals reach the author without the `folio:` prefix every sibling refusal carries
- **Deferred by:** Story 8.1's review pass (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** the next story that adds an INTEGER-valued command field, **or** Epic 8 close (whichever
  first) — a story and a gate, never an event, per D-000.73
- **Severity:** LOW — cosmetic, but it is inconsistency in the one string the author actually reads
- **Status:** OPEN

**The gap.** `fontChainIndex` passes `commandInt`'s error through verbatim, and `commandInt`
(`folio-go/component_commands.go:510`) — unlike its `commandString` sibling at `:1271` — does not prefix
its messages. An author moving a chain entry sees `"index must be an integer"` beside
`"folio: name must be a non-empty string"` for the very same command. The inconsistency lives in
`commandInt` and predates this story; Story 8.1 is the first caller to put it in front of an author.

**Anchors re-derived at this story's closing revision.** The build's note cited `:509`, which is the
blank line above the declaration — the kind of one-line drift that comes of capturing anchors before a
review patch lands.

---

### DW-73 — `page-setup-command.ts` carries the SAME unquoted-number splice DW-32 names, and DW-32 does not name it, so closing DW-32 as written would leave the authority not sole
- **Deferred by:** Story 8.2's plan gate (2026-08-31, Design Note N4), measured while establishing which
  encoder an author-typed chain name actually travels through
- **Owner:** **Story 15.2a** — *a component command means exactly what it names*, the same story that
  owns DW-32. This is not a second story; it is a **second site DW-32's own acceptance must cover**,
  filed separately so it cannot be lost when that entry is read as naming one file.
- **Severity:** **HIGH**, for the same reason DW-32 is: it is the same mechanism at a second entry point
- **Status:** OPEN

**The gap.** DW-32 names `folio-designer/src/component-property-command.ts` and its `rawNumberLiteral`.
`folio-designer/src/page-setup-command.ts:6-7` performs the **identical** splice — the author's typed
width, height and margins are interpolated into the command JSON **unquoted** — and its inputs are the
free-text page-setup fields at `App.tsx:887`. So the injection DW-32 describes has **two** reachable
front doors, and a fix applied only to the file DW-32 names leaves the second one open while the entry
reads as closed.

**Measured, not inferred.** The designer has **five** command encoders giving **three** different
answers: `table-column-command.ts` uses `JSON.stringify` (the correct model); `component-command.ts`
and `component-asset-command.ts` carry byte-identical hand-rolled escapers **which are byte-identically
WRONG on non-BMP input — see [DW-75](#dw-75)**; `component-property-command.ts`
carried an incomplete one (its `quote()` half was fixed by Story 8.2 — see DW-32's status note); and
**numbers are not encoded at all**, in two files. That is D-8.1.3's exact shape: no single authority,
so no single place to be right.

**What closing it requires.** Whatever Story 15.2a builds must be routed through by
`page-setup-command.ts` as well, and its acceptance must exercise that file by name. Anything less
produces an authority that is not sole, which is the only property worth having.

---

### DW-74 — the Go/TypeScript wire test records the projection's TOP-LEVEL and `CanvasFontChain` key lists but NOT `CanvasComponent`'s, so a per-component projection field would blank the canvas with every test on both sides green
- **Deferred by:** Story 8.2's plan gate (2026-08-31, Design Note N5 v), measured while establishing
  that this story adds no projection field
- **Owner:** **Story 8.4** — the first story whose stated work (a per-component preview measured through
  the engine's own path) is likely to add a `CanvasComponent` field, and therefore the first to walk
  into this gap; **or Epic 8's close**, whichever comes first, per D-000.73
- **Severity:** MEDIUM — the symptom is the silent blank canvas the wire test exists to prevent, and the
  gap sits one level below where that test looks
- **Status:** OPEN

**The gap.** `folio-go/canvas_projection_wire_test.go` ties two key sets across the language boundary:
the projection's top-level keys (`:47`, against `engine-protocol.ts`'s `isCanvas` `hasOnly` list) and
the nested `CanvasFontChain`'s (`:76`, against `hasExactKeys(chain, ['name','entries'])`). It does
**not** read `engine-protocol.ts`'s `CanvasComponent` key list. That list is a `hasOnly` **subset**
check over roughly thirty keys, so a field Go adds to a projected component and the designer does not
list makes `isCanvas` return false, `parseInbound` discard the whole snapshot, `engine-client` raise
`PROTOCOL_INVALID` and **terminate the worker** — with the Go tests green (they read struct fields) and
the designer tests green (their canvas fixtures are hand-authored object literals that never see Go).

**Why it is reported and not fixed here.** Story 8.2 is designer-only and adds no projection field, so
it cannot make this reachable and a Go-side test change would be outside its diff. It is the gap DW-35's
eventual fix walks into, which is why the owner is pointed at the same story.

---

### DW-75 — `component-command.ts` and `component-asset-command.ts` CORRUPT non-BMP text: they iterate by code point but escape from `charCodeAt(0)`, so an astral character becomes a LONE SURROGATE and binds to a different path than the author typed
- **Deferred by:** Story 8.2's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 15.2a** — *a component command means exactly what it names* — the same owner as
  **DW-32**, and **sequenced before Story 15.3 cuts the tag**. This is not an additional candidate for
  that story; it is what that story's consolidation actually **fixes**, and 15.2a must know it is
  repairing a live corruption rather than tidying three spellings of one idea.
- **Severity:** **HIGH** — not "the encoders differ" but "two of them silently produce the wrong
  document path". Re-verified **by execution** at Story 8.2's close, not carried on report.
- **Status:** OPEN

**The defect, re-derived by executing the function at Story 8.2's close.** Both encoders read
`for (const character of value) { const code = character.charCodeAt(0); ... }`. `for...of` over a
string iterates **by code point**, so for U+1F600 the loop variable is the whole two-unit string.
`charCodeAt(0)` then reads only the **first** UTF-16 unit — the high surrogate `0xD83D` — which falls
in the `code >= 0xd800 && code <= 0xdfff` branch and is emitted as `\ud83d`. The loop has already
consumed the whole code point, so **the low surrogate is never emitted at all.**

Executed at close: `quote('a' + U+1F600 + 'b')` returns `"a\ud83db"`, which parses back to the three
code points `U+0061 U+D83D U+0062` — a **lone surrogate**, not the author's emoji. Go's
`encoding/json` substitutes **U+FFFD** for an unpaired surrogate, so the value the engine stores is
not the value the author typed.

**Why it matters beyond mojibake.** These two encoders carry **bind segments** and **asset keys** —
values that are *addresses*, not display text. A bind segment or asset key holding an astral character
therefore binds to a **different path than the author typed**, silently, and the engine's refusal names
a path the author never picked. There is no error anywhere on this route: the JSON is well-formed, the
arity is right, and both languages agree on a value neither of them received.

**Why every existing test misses it.** `component-command.test.ts`'s "complete JSON escaping" case
uses only BMP inputs. A lone surrogate reaching Go is invisible to any test written with BMP text,
which is why the register recorded these encoders as merely "unconverted" for as long as they were
read rather than run. **15.2a must assert the non-BMP case explicitly**; a consolidation that routes
both through `JSON.stringify` fixes this as a side effect, and a test that only proves "the payload is
valid JSON" would go green without ever proving the round trip.

**Why it is not fixed by Story 8.2.** D-8.2.6 settled these two encoders as Story 15.2a's, and 8.2's
contract forbids touching them by name (*"do not consolidate the five encoders"*). Story 8.2 fixed only
`component-property-command.ts`'s `quote()`, which is on its own path. Story 15.2a's scope in
`epics.md` was amended at `7692b50` to carry this finding; this entry is the register's half of it.

---

### DW-76 — nothing ties `font-chain-command.ts`'s six command kinds and field arities to the Go dispatch table they must match, so a rename in Go alone leaves every test in both languages green
- **Deferred by:** Story 8.2's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Epic 8 close** — due before that epic's key is marked `done`, alongside **DW-71**,
  which is the same boundary one layer up. Every remaining story in Epic 8 adds command surface behind
  this same untied seam.
- **Severity:** MEDIUM — **the same shape as DW-42 and the five untied Go/TS invariants** Story 7.5's
  inventory found: two sides of one contract, each asserted against its own independently hand-written
  literal, coupled by nothing executable. That class has already produced one high-severity defect in
  this project.
- **Status:** OPEN

**The gap.** `font-chain-command.ts` builds six payloads at the exact arities `componentFields` counts
(4/4/3/5/5/4) and with the field names Go's dispatch reads. Both sides then assert against
**independently hand-written literals**: `font-chain-command.test.ts` and `App.test.tsx` on the
TypeScript side, `component_commands_test.go` on the Go side. Renaming `from`/`to` on
`moveFontChainEntry`, or changing one `componentFields(raw, N)`, **in Go alone** leaves every test in
both languages passing while the designer dispatches payloads the engine refuses with an arity error
at every move — a feature that is green in CI and broken for every author.

`canvas_projection_wire_test.go` closes exactly this seam for the **Go to browser projection**
direction. The **browser to Go command** direction has no counterpart at all.

**The same shape, one layer down, is the refusal text itself.** The sentences an author reads are
coupled to Go **by transcription only** — no test reads both languages — so a reworded Go refusal
leaves the designer's fixtures asserting a string the engine no longer emits, green on both sides.
Story 8.2's own contract test forbids a TypeScript *copy of a rule*; it cannot detect a *stale
transcription of a message*. Both halves belong to whoever closes this entry.

**Scope.** Pre-existing for the five older encoders; Story 8.2 adds six more command kinds to the same
gap and so widens it without creating it. Reported rather than fixed because 8.2 is designer-only by
acceptance criterion and the fix is a Go-side test.

---

### DW-77 — the isolable half of `invalidatePreview()` on the chain path is unverified: deleting the call leaves every test green, because a second mechanism produces the same user-facing result
- **Deferred by:** Story 8.2's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **Story 8.4** — the first story to hold a render open across an edit (its preview is
  measured through the engine's own path), and therefore the first with the fixture this proof needs;
  **or Epic 8's close**, whichever comes first, per D-000.73
- **Severity:** MEDIUM — a coverage gap, not a live defect. The behaviour is correct today; nothing
  would catch it ceasing to be.
- **Status:** OPEN

**The gap, and the reason it is honest rather than hidden.** Story 8.2 added a test asserting that an
accepted chain command marks a rendered PDF stale, and it passes. But **deleting `invalidatePreview()`
leaves all tests green**: re-entering Preview re-requests identity, and `renderPreview`'s own
identity-mismatch branch sets the stale reason independently, so the user-facing property holds either
way. The passing test therefore measures the *sibling* mechanism, not the call it appears to cover.

This is the failure mode where a test passes because a **different** mechanism covers the case.
Isolating the token bump — the half that revokes an **in-flight** render — needs a render held open
**across** the chain command, which Story 8.2 did not build.

**Recorded as reported, not as covered.** The implementer found this by mutation and flagged it rather
than claiming the coverage, which is the reason it is a register entry instead of a future surprise.

---

### DW-78 — `FontChainEditor` reads its form values through `document.getElementById` on the GLOBAL document, so a second mount collides on ids and reads the wrong field
- **Deferred by:** Story 8.2's build (2026-08-31); filed into this register at the story's close,
  where it was found recorded only in the spec's frontmatter
- **Owner:** **the plan gate of the first story that mounts a second `App`, a portal, or a preview pane
  carrying the property panel** — a gate, never an event, per D-000.73
- **Severity:** LOW — unreachable today and correct today; a latent coupling, not a defect
- **Status:** OPEN

**The gap.** `typed()` resolves its ids against the **global** `document`, as does the focus-restore
pass in the settle effect. The ids are `useId`-scoped, so they are unique per mount — but the *lookup*
is not scoped to the component's own subtree. Two `App` instances in one test file, a future portal, or
a preview pane rendering the panel a second time would each resolve against whichever node the global
document returns first. A `useRef` map would be local, typed, and immune to this by construction.

**Why it holds today.** Exactly one `App` mounts, in production and in every test. This is filed so the
constraint is written down somewhere other than in the fact that nothing has violated it yet.

---

### DW-79 — the chain editor decides "the list moved" by a VALUE SIGNATURE rather than array identity, and nothing pins that choice: reverting it to a reference check goes unnoticed by all 319 tests
- **Deferred by:** Story 8.2's **close** (2026-08-31) — found by the closer's own mutation screen, not
  by the build's review pass
- **Owner:** **the plan gate of the next story that changes the chain editor's focus or settle
  behaviour**, or **Epic 8's close**, whichever comes first, per D-000.73
- **Severity:** LOW — the code is correct and strictly better than what it replaced; what is missing is
  the guard that keeps it that way
- **Status:** OPEN

**What was measured at close.** Story 8.2's review pass found and fixed a real defect: the editor
originally decided whether an edit had landed by comparing the `chains` **array identity**, which would
call an accepted no-op edit a move, and which was unfalsifiable in test because a fixture legitimately
reuses one array across snapshots. The fix compares a value signature over names and entries. **The fix
is real and is in the code**, and the mechanism it feeds has teeth in both directions — the closer
proved this by mutation:

- forcing `landed = false` reddens **4** tests (both keyboard-reorder focus tests and both
  add-field-clearing tests);
- forcing `landed = true` reddens **1** (`empties the add fields once the add has actually landed`).

**The gap.** Replacing the value signature with a faithful **array-identity** comparison — the exact
defect that was fixed — leaves **all 118 `App.test.tsx` tests green** (and the whole suite at 319/35
green). No test discriminates the value signature from the reference check, so the specific choice that
was the subject of the finding is unguarded, and a future refactor could reinstate the original defect
silently. Closing this needs a test in which two successive snapshots are **equal in value but distinct
objects** (or the reverse), which is precisely the fixture shape the original finding named.

---

### DW-80 — `assetKeyReferenced` returns false for EVERY font asset, so an orphan collector would delete fonts a document is using

- **Deferred by:** Story 8.3's plan gate (2026-08-31), which measured it while checking what of the
  image-asset mechanism is reused. **Not reachable today** — nothing collects orphans — so it is a
  trap laid for a later story rather than a live defect.
- **Owner:** **Story 8.6**, or whichever story first collects, prunes or garbage-collects unreferenced
  assets. A named story with a position in the sequence, per D-8.0.5 — **not** a checklist item; a
  checklist address was read past by eight consecutive plan gates in this run (DW-28).
- **Severity:** MEDIUM as filed, **HIGH the moment anything collects orphans** — it is silent data
  loss of an embedded face, and the document then fails to render for want of bytes it used to carry.
- **Status:** OPEN.

**The gap.** `assetKeyReferenced` answers *"does any element name this asset?"* by walking **image**
elements. A font asset is named from a **font chain entry** (`{"asset": "<key>"}`), which that walk
does not visit. So it returns **false for every font asset a document embeds**, however many chains
name it.

**Why it is filed now rather than found later.** Story 8.3 stores fonts as assets and adds no
collector, so nothing acts on the wrong answer today. The next story that prunes unreferenced assets
will call this function, receive a confident `false`, and delete a face the document is using. **The
function is not wrong about images and gives no sign that it is partial** — which is the shape this
run has now recorded eight times.

**What discharges it.** Extend the walk to font-chain entries **and** add a test that embeds a face,
names it from a chain, and asserts the answer is `true` — the arm that a `false` answer for images
would never reach.

---

### DW-81 — `linespacing_test.go`'s version-trigger list is hand-enumerated while its comment claims a derivation — **CLOSED by Story 8.3, 2026-08-31**

- **Deferred by:** Story 8.3's plan gate (2026-08-31).
- **Owner:** **Story 8.3**, if it mints a font-shaped version trigger; otherwise the next story that
  adds one.
- **Severity:** LOW as a defect, **MEDIUM as a trap** — a version trigger that is never checked is
  the failure mode the file exists to prevent.
- **Status:** **CLOSED.** Story 8.3 minted the font-shaped trigger, so the condition attached to the
  owner fired. Both halves were discharged in the same commit: the **third** builder loop was added
  (asserting an embedded-face entry requires `2.0` and a font asset alone does not), and the comment
  was rewritten to say the enumeration is hand-maintained and to state the obligation on whoever adds
  the fourth trigger. The correction was red-proved rather than asserted — removing the probe reddens
  the new loop, and mutating the probe to raise on a font ASSET reddens it in the other direction.

**The gap.** The comment claims the trigger set is **derived**; the code **hand-enumerates** builder
loops. So a new trigger — a font-shaped one, for instance — is **silently never checked** unless
someone remembers to add a loop. **A comment asserting a derivation the code does not perform carries
the same evidentiary burden as a test** (D-8.0.1), and this one is false.

**What discharges it.** Either derive the list, or correct the comment to say it is hand-maintained
and add the missing loop in the same commit as any new trigger.

---

### DW-82 — the Go/TypeScript wire test recorded the font chain's key names but nothing about an ENTRY, so changing an entry's TYPE blanked the canvas with every test on both sides green — **CLOSED by Story 8.3, 2026-08-31**

- **Deferred by:** Story 8.3 (2026-08-31), recorded on closing it.
- **Owner:** Story 8.3.
- **Severity:** HIGH while open — the symptom is a permanently blank canvas with no element id and
  nothing to attribute it to.
- **Status:** **CLOSED.**

**The gap, as measured.** `canvas_projection_wire_test.go` held two recorded key sets: the
projection's top-level keys and `CanvasFontChain`'s (`{"entries", "name"}`). Both record key **NAMES
only** — never a value's type, and never anything one level further down. So Story 8.3's change,
`CanvasFontChain.Entries` going from `[]string` to a slice of structs, moved **no** recorded name and
left both records green, while `engine-protocol.ts`'s entry clause (`typeof face === 'string'`)
rejected every object entry: `isCanvas` false → `parseInbound` undefined → `engine-client`
terminates the worker → the canvas is blank. This is DW-74's shape (which names `CanvasComponent`)
recurring one level down on a different object, and DW-74 stays OPEN on its own terms.

**What closed it.** A **third** recorded key set, `canvasFontChainEntryWireKeys`, plus a matching
TypeScript-side extraction anchored on `isFontChainEntry`'s own name — the guard's own list, never a
copy of it. Both halves were red-proved by making the one-sided change in each direction: a field
added to `CanvasFontChainEntry` in Go alone reddens the Go half; a key added to the browser guard
alone reddens the TypeScript half.

**What is still not covered, and is DW-74's, not this entry's.** The records still compare key
NAMES. A field whose **type** changed on both sides in the same edit — `face: string` becoming
`face: number` in Go and in the guard — would pass all three. Closing that needs a type-level record,
which is a larger change than this story's, and DW-74 is where it belongs.

---

### DW-83 — a chain entry may name a NON-FONT asset: correct to accept at load, but it errors nowhere at render either, so D-1.8.1's shape is half-built — **CLOSED by Story 8.4, 2026-09-01**

- **Deferred by:** Story 8.3 (2026-08-31), recorded on closing it. Filed by the build in the spec's
  frontmatter only; entered into this register at close.
- **Owner:** **Story 8.4** — the story that resolves an embedded entry to bytes, and therefore the
  first story in which a render surface exists for this to fail on. A named story with a position in
  the sequence, per D-8.0.5, not a checklist item.
- **Severity:** MEDIUM. Not reachable as data loss; reachable as a document that is accepted, drawn
  wrongly and never explained.
- **Status:** **CLOSED by Story 8.4 (2026-09-01)**, as an ACCEPTANCE CRITERION and not as a deferral
  note — which is what D-8.3.5 required, because the half most likely to be dropped is the third one
  below and a note is what drops it.

  **All four rows, and where each is asserted.** The first three are D-1.8.1's three parts; the
  fourth is the silence that bounds them, and it is in the same table because the placement of the
  decode is what makes both true at once. Every one is red-proved by reverting its own production
  expression, and each is asserted by a **separate named test**, so removing any one arm reddens a
  test of its own rather than being covered by a sibling:

  | part | asserted by | reddens if |
  |---|---|---|
  | Load **accepts** an entry naming a non-font asset | `TestNonFontAssetIsAcceptedAtLoad` (`folio-go/chain_face_names_test.go`) | anyone "fixes" this by tightening `decodeFontChainEntry` |
  | `Render` **errors, located** — naming the chain, the entry index and the asset key | `TestNonFontAssetDrawnErrorsAtRenderAndAtValidate` | the decode is removed from `fontCache.get`, or the error stops carrying `template.FontChainSite` |
  | `Validate` returns the **identical** error | the same test's second half, asserted by `verr.Error() != rerr.Error()` **and** a zero diagnostic count | the `Validate` arm is removed while the `Render` arm remains |
  | An entry **nothing draws from** stays silent | `TestNonFontAssetNeverDrawnRendersClean` | the decode is hoisted out of the point of use into an upfront pass |

  **`Validate`'s half is true BY CONSTRUCTION, not by a second rule system.** The resolution lands
  inside `predictDocument` — the single derivation `folio.Validate` calls directly and `Render`
  reaches through `renderDocument`/`buildPageModel` — so there is nothing to keep in step.
  `TestValidateNeverReachesRenderOrInternalPDF` makes a `Render`-only placement structurally visible,
  and the separate `Validate` assertion above is the second defence.

  **The load behaviour was NOT touched.** `decodeFontChainEntry` still checks only that the key is
  present in `assets`. D-1.8.1 as amended is now stated in `folio-format.md` for the wrong-kind case
  as well as the unrecognised-type case, so the rule is written down where an implementer reads it
  rather than only where it was decided.

**The gap, as measured.** The entry decoder checks only that the `{"asset": "<key>"}` key is
**present** in the assets map, never that the asset is a font. Refusing it at load would violate
**D-1.8.1 as amended** — an unrecognised or wrong-kind media type is preserved at load and errors at
render — **so the load behaviour is correct as shipped and must not be "fixed" by tightening the
loader.** But `chainFaceNames` drops every embedded entry before face resolution, so the render half
never fires either. The result is D-1.8.1's shape with one of three parts built.

**What discharged it, all three parts, in Story 8.4.** Load continues to **accept**; `Render`
**errors, located**, when something actually needs to draw from that entry; and **`Validate` predicts
what `Render` would do** rather than answering from a second rule system. The third is the one that
gets dropped, which is why it was an AC.

**"When something actually needs to draw from that entry" was given an operational meaning, and it
is worth recording because it decided where the decode is REFUSED.** The refusal is at **coverage
resolution** (`resolveRuneFace`) — the first moment a rune reaches the entry because no earlier entry
in the chain covers that rune.

**The vertical-model walk decodes too; what it does not do is refuse.** This paragraph used to say
the entry "is deliberately **not** decoded by the vertical-model walk (`chainLineMetrics`)", and that
mechanism claim was false. `fontCache.metricsFace` (`folio-go/render.go`) calls `declares()`, which
is true for any embedded name, and then `get()`, which base64-decodes the asset and parses it — so
the decode **is** attempted on every metrics walk, and a **readable** carried face is decoded there
and **does** contribute its metrics, which is what makes the canvas and the page agree on line
advance. What `metricsFace` does with an **unreadable** one is swallow the error: an entry that
cannot supply a face cannot appear in the element, so it does not constrain the model — the same
tolerance that walk already applied to a chain member the caller did not supply.

The paragraph's conclusion was right and is unchanged: the two answers are consistent rather than
contradictory — **if a render completes at all, coverage never reached the entry**, so its absence
from the vertical model is exactly right. `metricsFace` is where that one rule is written down, and
its doc comment now states the decode-attempted / failure-tolerated shape rather than the
never-decoded one.

---

### DW-84 — Story 8.3 narrowed what already-shipped 1.x documents load: an empty face name is now refused, and the narrowing is not version-gated — **RULED: KEEP IT (D-8.3.1 / D-8.3.2)**

- **Deferred by:** Story 8.3's build (2026-08-31), which flagged it rather than deciding it. Routed
  to the **engineering lead** as a correctness question, not ratified by the orchestrator.
- **Owner:** none, and none is invented. The ruling resolves it; what remains is the record.
- **Severity:** MEDIUM as filed — a load narrowing applied to documents declaring an older version,
  which **the 22-digest corpus cannot observe**, since that row measures rendered goldens only.
- **Status:** **RULED — keep the refusal, unconditionally, and record it** (`2fe1e59`, with the
  guardrail verification at `051ee4f`). The three dispositions put to the lead were: reverse it;
  keep it and version-gate it; keep it unconditionally and record it. **The third was chosen**, on three grounds and explicitly **not** on the build's proposed
  ground of precedent-by-habit.

**The two shapes, and the honest half.** Before Story 8.3 the chain decoder accepted `""` as an
entry, so a 1.0 document containing `{"fonts": {"body": [""]}}` loaded. There are **two** cases, not
one, and only the second is interesting:

- `{"body": [""]}` alone — no usable entry — reached the existing located error. **Never rendered.**
- `{"body": ["", "Noto Sans"]}` — the empty entry was **silently skipped** by face resolution, Noto
  Sans was used, and the document **rendered cleanly**.

*"It was always a latent bug"* is true of the first and **false of the second**, and the record says
so rather than letting the comfortable half stand for both.

**Why the refusal is right (D-8.3.2).** (1) **AD-8**: the chain is part of the FontSet's identity, so
the same template with a different chain is a different render, *not a silent substitution*. An
author declaring a two-entry chain and getting a one-entry render with nothing saying so is the
substitution AD-8 forbids by name — **the old behaviour was the defect.** (2) **D-1.8.1's
reader-independence test puts this at load**: `"Helvetica"` is reader-*dependent* (another library
may ship it) and is therefore skipped as a capability question, but `""` is not a face name under any
reader — reader-*independent* malformedness, which is a load error. (3) **D-1.4.9 is not the rule in
play**: it promises a higher-MINOR file loads in an older reader — forward compatibility of the
*version field* — and does not promise a library may never tighten its rejection of malformed input.

**Version-gating was rejected on coherence, and the reason is recorded because it will be proposed
again: malformedness is not versioned.** *"This file is malformed only if it claims to be new"* gives
two answers for the same bytes keyed on a number the author can edit — strictly worse than the
reader-dependence D-1.8.1 already rejected, where the discriminator was at least outside the
document.

**Before-the-tag set: unchanged, and it stays at two.** By **D-8.2.2's** test this does **not** join
it — the change has already shipped at `af4efde`, and every correction available from here is a
**widening**, the same shape as DW-69.

**Guardrails, and their state at close.** Both shapes are now asserted, with the mixed chain named as
the one that matters (added at close). The refusal names the chain **and** the entry index, verified
rather than assumed, and the index was red-proved with a non-first failing entry. The narrowing is
recorded above naming the second shape explicitly, so no later reader is told it affected only files
that never worked.

---

### DW-85 — an evidence test rests on a hand-maintained fixture count that no code derives

- **Deferred by:** Story 8.3 (2026-08-31), frontmatter only; entered into this register at close.
- **Owner:** the next story that adds or removes a fixture carrying a non-empty `assets` map, or any
  story that chooses to derive the population instead.
- **Severity:** LOW.
- **Status:** OPEN.

**The gap.** `TestTheFontRecordCostsAnExistingDocumentNothing` asserts `withAssets != 7`, and its own
failure message concedes the number must be edited by hand whenever any fixture gains or loses a
non-empty assets map. A fixture added without touching that line **reduces the population silently
rather than failing loudly** — the test still passes, over fewer documents than it claims to cover.

**Why it is the same family as DW-81**, which this story closed elsewhere: a stated coverage claim
that the code does not derive. It is filed rather than fixed because deriving the population is a
change to how fixtures are enumerated, not to this story's subject.

---

### DW-86 — `TestShippedFacesReproduceFromUpstream` is a standing red that nothing registers, which is the DW-23 shape starting again

- **Deferred by:** the engineering lead at Story 8.3's close (**D-8.3.4**, 2026-08-31).
- **Owner:** the next story that touches the gate definition, or whoever provisions the gate
  environment. **Not** Story 8.4 by default — it is not a font-rendering question.
- **Severity:** MEDIUM. The defect is not the red; it is the **unregistered** red.
- **Status:** OPEN.

**⚠ THIS ENTRY WAS WRONG, AND THE ENTRY — NOT THE TEST — IS WHERE THE DEFECT LIVED.
CORRECTED 2026-09-01 (D-8.4.34). The original text is kept below.**

**It is NOT a standing red. It is a misconfigured interpreter suppressing a working check.** The
pinned toolchain is **present**: `.fontgen-venv/bin/python` is **Python 3.12.13 with fontTools
4.63.0** — the exact pair the NOTICEs name — and `.font-sources/` holds all three upstream variable
fonts. `go test` invokes a bare `python3` that lacks fontTools. **With `FOLIO_FONTGEN_PYTHON` pointed
at the venv the test PASSES, non-vacuously** — `fontgen: derived and compared 3 of 3 faces`, with
real produced sizes (646160 / 47788 / 10595932 B) and digests; the Thai digest `c94562c1…73caf`
independently matches the value hashed for D-8.4.8's attestation transfer.

**THE TEST IS LARGELY INNOCENT, and this is the durable lesson.** It says in its own words *"IT DOES
NOT SKIP WHEN ITS INPUTS ARE ABSENT"* and carries **five distinct `t.Fatalf` sites**. It already
refuses to be quiet and **already distinguishes its causes in its messages**. The three states were
conflated **by this register entry**, which recorded *"pre-existing and environmental"* — **a GUESS AT
THE CAUSE** — instead of the assertion that actually fired. **Had the entry quoted the message
verbatim, the wrong interpreter would have been visible on the day it was written**, because the
message that fired was **not** the sources-absent one.

> **STANDING RULE (D-8.4.34): a standing red is registered by its failing assertion's message,
> VERBATIM — never by a category.** The verbatim text for this one is:
> `fontgen: fontTools is not importable by this interpreter`.

**And the three-state distinction this run keeps rediscovering:** an **all-clear** must differ from a
**couldn't-look**, and a couldn't-look must differ from a **looked-in-the-wrong-place**. The third is
the state that keeps appearing — this bare `python3`, the 428 KB Chromium stub, and
`s1VisibleBytes`' four-needle total. **All three are instruments pointed somewhere other than where
their reader believed.**

**RESIDUAL, recorded so nobody later reads this as covering more than it does: the check covers 3 of
the 6 shipped faces.** The three IBM Plex faces added by Story 8.4c are **vendored static files from
a pinned npm package** — their guarantee is a package version plus a NOTICE, which is a **provenance**
claim, not a **reproduction** one. **Two different kinds of assurance under one heading.**

---

**The original, wrong text follows.**

**Measured at close.** Under `-tags=matrix` the test **FAILS — it does not skip** — reporting
`fontTools is not importable by this interpreter`. Run at baseline `f51dd5e` in a detached worktree
**with the upstream sources supplied** (`FOLIO_FONT_SOURCES`), it fails at the same line with the
same message, so it is pre-existing and environmental; `folio-go/fonts/` is untouched by Story 8.3's
diff.

**That it fails rather than skips is CORRECT and must stay.** *"The sources were not present"* must
never read as *"the faces reproduce"* — the all-clear must differ from could-not-look.

**The failure is that it has gone unmeasured for several stories.** An unregistered standing red is
the **DW-23 shape**: a red nobody has decided about is the one that masks the next real failure. Two
acceptable resolutions — make `fontTools` available to the gate so the test actually runs, **or**
register it explicitly alongside the P6g floor as a known, named red with its reason. **This entry is
the second, taken now so that neither-of-the-two stops being the state.**

---

### DW-87 — the review triage record counts rejections without enumerating them, so a rejection cannot be audited at close

- **Deferred by:** Story 8.3's close (2026-08-31), found while auditing the build's own triage.
- **Owner:** whoever next amends the build loop's triage record format — a process defect, not a code
  one.
- **Severity:** LOW per story, **MEDIUM as a pattern** — it removes the one check that stands between
  an optimistic rejection and a shipped defect.
- **Status:** OPEN.

**The gap.** Story 8.3's `## Review Triage Log` records `reject: 6` and lists `addressed_findings`
in full, but **records nothing at all about the six rejections** — no summary, no location, no
refutation. Patched findings are legible and auditable; rejected ones are a number.

**Why it matters here specifically.** A rejection is sound only when it refutes the *specific claim*
at the *cited location*, and *"a true fact about nearby code"* is not a refutation. That test cannot
be applied to a count. Story 8.3's rejections were therefore **not** spot-checked at close, and the
Delivery Log says so rather than implying they were.

**What discharges it.** Record each rejection the way `addressed_findings` records a patch: the
claim, its location, and the ground on which it was refused.

---

### DW-88 — the corpus's only Thai-bearing golden with no human reading attestation — **RULED (D-8.4.8): the attestation TRANSFERS by asserted shaped-run equality**

- **Deferred by:** Story 8.4 (2026-09-01). Filed by the build in the spec's frontmatter as two
  entries; entered into this register at close as one, because they are one question — the
  attestation and the gate that would hold it open are the same obligation.
- **Owner:** **Engineering lead**, routed by the run orchestrator at Story 8.4's close and awaiting a
  ruling. The reading itself, if the ruling authorises the gate, is the **project owner's** — **no
  agent may write `reader`, `date` or `examined` in any sign-off file.**
- **Severity:** LOW as measured, **but it is the kind of gap D-2.3.5 calls "an obligation nobody
  trips over"**, which is why it is registered rather than left in a spec's frontmatter.
- **Status:** **RULED 2026-09-01 (D-8.4.8) — a FOURTH disposition, better than the three put to the
  lead and grounded on measurement rather than on any of them.** The transferred record is written at
  `fixtures/embedded-font/signoff.json`; the asserting test and its matrix gate are the follow-up.

  **The premise the orchestrator offered was REFUTED, and that matters more than the verdict.** The
  question routed was whether *"Ts-free"* makes the stacked-mark hazard structurally absent, which
  would have strengthened the do-nothing disposition. **It does not.** This repo's own
  `fixtures/thai-stacked-marks/README.md` records that `ที่`, `ป้ำ` and `ปั` stack **two** marks over
  one base and are resolved by a **GSUB lowered-form substitution at zero offset**. Ts-freeness means
  every glyph has `YOffset == 0` — it excludes only the **GPOS vertical-displacement** leg, and
  nothing about X-offsets, glyph selection, base-mark association or cluster order, which are exactly
  the **eye-only** failure modes. `สัญญา` carries **U+0E31 MAI HAN-AKAT** (`Mn`), whose eye-only
  failures — the mark over ญ instead of ส, a spacing glyph, a cluster reorder — all survive it.

  **Why the transfer is sound.** `fixtures/embedded-font/`'s `e1` and `thai-stacked-marks`' `e2` are
  identical in every input that determines a shaped run (same string, `body`, 12pt, width 400, x 0),
  differing only in `y`/`height` — **placement, not shaping** — and the face is the same **program**,
  measured rather than taken from the asset's prose `source`:
  `sha256(fonts/notosansthai/NotoSansThai-Regular.ttf)` = the embedded **asset key** = `sha256` of its
  decoded 47788 bytes = `c94562c1…73caf`. The PDFs differ (different subset, AD-7); the attested layer
  does not.

  **TWO CONDITIONS, and the first is a LAPSE CONDITION rather than a task.** (1) The equality is
  live-vs-live, so it is non-vacuous **only because** `fixtures/thai-stacked-marks/expected.pdf` is
  byte-pinned and the sign-off names its digest — **if that golden is ever re-recorded the transfer
  LAPSES and both fixtures need a real human reading.** (2) The transfer holds only while the embedded
  bytes are a shipped face's. **Checked before adding a tripwire, and none is needed:**
  `embeddedFontAssetBytes()` returns `testShippedNotoSansThai`, which is
  `//go:embed fonts/notosansthai/NotoSansThai-Regular.ttf` — the shipped file **directly**, so the
  condition is held by construction. **Standing rule:** a future fixture embedding a face that is
  **not** a shipped face's bytes owes a **real human reading** and must go **red rather than
  transfer**.

  **The anchor is SCOPELESS, and that limit is recorded rather than smoothed.** It reads, in full,
  *"The rendering at fixtures/thai-stacked-marks/expected.pdf looks ok."* — naming no scope, unlike
  `fixtures/statement-signoff.json`. **An attestation's scope is what transfers, so a scopeless one
  transfers scopelessly.** Accepted here because it is the owner's eye on a page whose entire subject
  is Thai mark placement. **It must not become the template — future readings name what was checked.**

  **No agent wrote `reader`, `date` or `examined`.** The transferred record carries neither field at
  all, plus an explicit `NOT_A_HUMAN_READING` statement, because a reconstructed record that does not
  say it is reconstructed is the failure this run has flagged before.

  **DISCHARGED 2026-09-01 at `43da56a`.** `TestEmbeddedFontShapedRunEqualsAttestedControl` compares
  at `shapeSegments`, whose glyphs are documented as *"the glyph index in the SOURCE face's numbering
  — not a subset glyph id and not a PDF CID"* — so **pre-subset was reachable and nothing weaker was
  substituted**. Five glyphs, identical GlyphID/XOffset/YOffset, with both fixtures read from their
  committed `input.folio` rather than the Go mirror constants. The test asserts the **difference
  first** — the control's faces in the `FontSet`, the carried face as `asset:c94562c1…` and **not** in
  it — before asserting the equality, so it cannot pass by the two sides being the same thing.

  **The lapse condition is asserted TWICE, independently** (`TestEmbeddedFontTransferredReadingHolds`,
  `//go:build matrix`): a frozen literal of the anchor digest in the gate file **and** a live re-hash
  of `fixtures/thai-stacked-marks/expected.pdf`. Appending one byte to that golden reddens the gate
  with **"THE TRANSFER HAS LAPSED"**. The frozen literal sits deliberately **outside
  `goldenDigestSearchScope`** — that scope exists to find sites a re-record must update *together*,
  and this is one that must **never** be updated. Adding a `"reader"` field reddens two tests.

  **Registration required a NEW SITE KIND, `transfer-anchor`** — the existing `signoff` case compares
  a record's *top-level* `sha256`, which is the wrong field for the anchor digest. So the omission at
  `4219a1b` (below) was not merely forgotten; it was **not expressible in the existing vocabulary**.

---

### DW-92 — the canvas-abort scoping rests on a premise `checkSfnt` does not hold, and the retained half was pinned by nothing

- **Deferred by:** Story 8.4's D-8.4.8 follow-up (2026-09-01), found while building the transfer
  assertion. **Routed to the engineering lead the same day; awaiting a ruling.**
- **Owner:** **Engineering lead.** The builder deliberately did **not** widen the arm: the scoping was
  a deliberate ruling, and **a ruling with a false factual premise is an intent gap, not a bug to
  patch.**
- **Severity:** MEDIUM — a document the designer cannot open, with an author repair available.
- **Status:** **RULED 2026-09-01 (D-8.4.12) — widen, on the ATTRIBUTABILITY axis rather than by
  extending an error-type allowlist.** A fault resolving a face the element's own chain names degrades
  that element whatever error type carries it; a fault after every needed face resolved keeps
  aborting. **The lead's own premise was the false one**, and it named the error as the same axis
  mistake as D-7.3.1. Six guardrails, two of which forbid the obvious implementations: **do not extend
  the allowlist** (mint one type at the single `fontCache` resolution door), and **do not pre-resolve
  the chain** (`resolveRuneFace` skips faces the runes do not need, so a pre-pass would degrade an
  element over an entry it never draws with). **The retained arm must gain an assertion — that, not
  the abort, is the finding.** Implementation is Story 8.4's follow-up.
- **CLOSED 2026-09-01 — implemented as ruled.** The allowlist is deleted, not extended. The
  discriminator is **one type stamped on EVERY error out of `fontCache.get`'s embedded arm**, the
  single door that resolves a face the document carries, so a future face-resolution failure type
  joins the degrade instead of silently rejoining the abort. The retained abort gained
  `TestCanvasStillAbortsOnAHostFontSetFaceThatWillNotParse` — guardrail 6's caller-supplied face,
  which is the case that **discriminates the two axes** and is deliberately **not absorbed**.
  `TestCanvasStillAbortsOnAnUnreadableCarriedFace` was **inverted** to
  `TestCanvasDegradesRatherThanAbortingOnAnUnreadableCarriedFace`, keeping its `ParseTemplate`
  precondition `Fatal` verbatim. `ContentWindowCountIsExact` cause (c) is amended.
- **ONE DEVIATION FROM THE RULED MECHANISM, and it moved the CARRIER, never the DOOR.** The stamp was
  first written as a package-local error type beside `fontCache` — the site guardrail 1 names — and
  that **reddened `TestFolioMethodNamesAreInjective`** (`render_arch_test.go`): a second root-file
  receiver type declaring `Error` and `Unwrap` makes `buildFolioCallGraph`'s name-keyed method map
  lossy. That guard's own two remedies are **rename the methods** (impossible for `error`) and **take
  DW-20 now** (a separate story), so neither was available. The type therefore lives in
  `internal/template` as **`CarriedFaceError`**, beside `UnsupportedFontMediaTypeError` — the
  document-attribution error it generalises and now wraps — while still being **minted at exactly one
  site**, in `get`'s embedded arm. The invariant guardrail 1 protects is intact; only the declaration
  site changed, and package `folio`'s root files gain no method at all. **DW-20's pressure is real
  and unrelieved**: the guard fired for the first time on a legitimate change, and the next
  package-level error type in `folio` will fire it again.

**The gap.** `internal/template/fontasset.go:checkSfnt` validates **the table directory only, never a
table's contents** — its own comment says so. So a carried face that is a **structurally valid sfnt
with unreadable contents LOADS**, and then aborts the **whole canvas projection** at `fontset.New`,
propagated by `wasm/engine.go:119,255,294`. Truncated and corrupt bytes **are** caught at load, so
this is the narrow surviving hole rather than a general one.

**Why it is the same shape Story 8.4 just fixed.** It is a **document property with an author repair,
aborting the one surface the author would use to make it** — the D-7.4.2 violation, surviving on the
neighbouring arm of the very branch 8.4 repaired.

**What makes it more than a corner: mutating that arm to degrade unconditionally reddened NOTHING in
the entire suite.** The half Story 8.4 changed was pinned by one test; the half it **retained** was
pinned by **none**. The scoping has been resting on a premise nothing measured, on either side.
`TestCanvasStillAbortsOnAnUnreadableCarriedFace` pinned it as a **characterization** test — it
recorded the measured behaviour and **explicitly declined to ratify it**, which was the correct call
and is what produced the ruling. It is now inverted, under the name
`TestCanvasDegradesRatherThanAbortingOnAnUnreadableCarriedFace`.

---

**The gap.** `fixtures/embedded-font/expected.pdf` ships from Story 8.4 onwards and its page is Thai
drawn from the face the document carries. Every OTHER Thai-bearing golden pairs a human reading
sign-off with a failing `//go:build matrix` gate and a `goldenDigestRecord` site of
`{kind: "signoff"}` — `fixtures/shaped-text/thai-signoff.json` under D-2.3.5,
`fixtures/thai-stacked-marks/signoff.json` under Story 8.0. This one has neither.

**Why Story 8.4 could not discharge it, and was right not to try.** The spec's `Block If` halts the
story if a `//go:build matrix` sign-off gate is left red, so minting the gate inside the story would
force the very halt the gate exists to prevent. The build filed it rather than minting; that was the
correct call and it is recorded here so the choice is visible.

**What is already attested, measured rather than assumed.** The drawn string is `สัญญา`, which is
byte-for-byte the string `fixtures/thai-stacked-marks/` draws as its **e2 zero-offset control**, at
the same 12pt, shaped by the **same** face — `folio-go/fonts/notosansthai/NotoSansThai-Regular.ttf`,
whose bytes this document merely carries a second copy of. So the shaping and the reading of those
five glyphs are attested by the owner (2026-08-31); what is unattested is **this subset and this
page**. Re-checked at close: the full `-tags=matrix` sweep leaves **no** sign-off gate red and
invalidates **no** existing attestation, so AC6's three grounds for closing `done` without a new
sign-off all held.

**What discharges it.** A ruling that authorises minting the gate together with scheduling the human
reading — the two must land together, or the gate is a red nobody has decided about, which is the
DW-23 shape.

---

### DW-89 — a readable embedded entry that nothing ever draws with still contributes its metrics, so line advance can move across the 8.4 boundary

- **Deferred by:** Story 8.4 (2026-09-01), recorded on closing it.
- **Owner:** **Engineering lead** to pin the intended answer; due at **Epic 8 close**. Nothing pins
  it either way today, so a later change in either direction is currently unfalsifiable.
- **Severity:** LOW. No committed golden can observe it.
- **Status:** OPEN.

**The gap.** `fontCache.metricsFace` treats any embedded name as declared and decodes it, so
`chainLineMetrics`/`chainVerticalModel` see a carried face even when no rune ever resolves to it.
Measured at close: only `fixtures/embedded-font/` has an embedded chain entry, so all 22 pre-existing
digests are unmoved and no golden can see the difference. But an integrator's document that carries a
face it never draws from could see its **line advance change across this version boundary**, with
nothing in the repo asserting whether that is right.

**What discharges it.** A ruling on the intended answer plus a test that pins it — either "a carried
face constrains the vertical model whether or not it draws" or "only faces that draw constrain it".

---

### DW-90 — a chain whose ONLY entry is an unreadable embedded one produces the right error by evaluation order, not by assertion

- **Deferred by:** Story 8.4 (2026-09-01), recorded on closing it.
- **Owner:** **Epic 8 close.**
- **Severity:** LOW.
- **Status:** OPEN.

**The gap.** Both new tests that exercise a non-font chain entry
(`TestNonFontAssetDrawnErrorsAtRenderAndAtValidate`,
`TestChainOfOnlyUnusableEntriesProducesTheExistingLocatedError`) place the failing entry **behind** a
shipped face that supplies the vertical model. A probe of the single-entry case returns the located
capability error rather than the generic empty-metrics error, so the behaviour is right today — but
which of the two errors wins depends on evaluation order that no test fixes. Confirmed still true at
close: the single-entry shape projects and refuses correctly, and nothing asserts which error it
must be.

**What discharges it.** A test that fixes the answer for a chain whose only entry is unreadable.

---

### DW-91 — two error branches the code itself documents as structurally unreachable are untested

- **Deferred by:** Story 8.4 (2026-09-01), recorded on closing it.
- **Owner:** **Epic 8 close.**
- **Severity:** LOW.
- **Status:** OPEN.

**The gap.** `predictDocument`'s `if cache.isEmbedded(name)` arm is commented *"Unreachable in
practice: a name only reaches this loop by having already been parsed"*, and
`embeddedFaceSource.decode`'s `if !s.present` arm is *"structurally unreachable today"*. Neither is
exercised. An untested unreachable error path is the one that is wrong when a later change makes it
fire.

**What discharges it.** This repo's own idiom elsewhere: an unreachability tripwire that reddens when
the branch becomes reachable, rather than a test that contrives the impossible state.

### DW-93 — the canvas cannot say WHY an element is blank, and D-8.4.12 added a second member to that silence

- **Deferred by:** the engineering lead with D-8.4.12 (2026-09-01), registered rather than left
  implicit, per its standing rule that **a silence ruled owes an escape hatch**.
- **Owner:** the first story that must distinguish *"no chain chosen"* from *"the chosen face will not
  load"* **on the canvas**. That is the **trigger**, and it is a new one — not a renewal of an
  existing deferral.
- **Severity:** LOW while the class has a reachable explanation elsewhere; re-price if it does not.
- **Status:** OPEN.

**The gap.** `CanvasProjection` carries **no diagnostics channel** — the struct was read, not assumed.
`FontChainDegraded` is internal, consumed only to compute `ContentWindowCountIsExact`, and never
projected. So a degraded element is **blank with no reason**, and D-8.4.12 put a **second** member
into that silence.

**Why it was ruled acceptable rather than escalated, on two narrow grounds.** D-7.4.2 already
established silent per-element degrade as the canvas's disposition for an unresolvable chain, so
D-8.4.12 **adds a member to an existing ruled class** rather than opening a new one; and the reason is
**not unreachable** — the same face still errors with a message on the **Render** path, so an author
who exports is told. The codebase's own metrics path already treats this condition as non-fatal, which
is **precedent**, not an argument against.

---

### DW-94 — a carried face is registered with no weight or style descriptors, so bold and italic components get browser-synthesized faces over a 400/normal one

- **Deferred by:** Story 8.4a (2026-09-01), recorded at close. Raised in review and deliberately not
  patched; filed here because the story's frontmatter is not the register.
- **Owner:** **Story 8.4b or the first story that gives a chain per-variant entries** — whichever
  reaches a face-variant decision first. This is a chain/face-variant **design** question, not a local
  repair.
- **Severity:** MEDIUM.
- **Status:** OPEN.

**The gap.** The registration seam constructs the face from bytes alone and passes no descriptors, while
the canvas still emits a 700 weight for a bold component and an italic style for an italic one. The
browser therefore **synthesizes** bold and italic over a regular face — different metrics from the ones
the engine measured with, which is precisely the class of defect Story 8.4a exists to remove. The
projected chain entry already carries a `style` and it is unused here. **Pre-existing in kind for shipped
faces too** (the build-time generator declares regular instances only), which is why this is a deferral
rather than a regression.

**What discharges it.** A ruling on how a chain expresses face variants, then descriptors on both the
build-time and the runtime registration paths.

---

### DW-95 — the canvas's inline carried-face family REPLACES the declared stack rather than extending it, so a registered face missing a glyph falls to the browser default

- **Deferred by:** Story 8.4a (2026-09-01), recorded at close.
- **Owner:** **Story 8.4b.** It is the story that re-points the canvas fragment stack, so it is the one
  place where "what the fragment falls back to" is already being decided.
- **Severity:** MEDIUM.
- **Status:** OPEN.

**The gap.** The fragment's inline `font-family` names one family and nothing after it. Story 8.4a gates
the inline declaration on the face having actually registered, so the **fetch-failure** row degrades
correctly — that path is closed. What is open is the face that IS registered and merely **lacks a
glyph**: CSS's own per-glyph font matching would have degraded onto the stylesheet's declared stack had
that stack been appended, and instead the glyph falls to the browser default. The intent contract does
not settle this fork and **no surface in this repository can observe the difference** (jsdom applies no
stylesheet and runs no font matching), which is the second reason it was not patched blind.

**What discharges it.** A decision on whether an inline carried family appends the declared stack, and —
because no unit surface can see it — a browser end-to-end check under **D-000.4**.

---

### DW-96 — no concurrency cap on the carried-face asset fan-out, and release does not abort requests already in flight

- **Deferred by:** Story 8.4a (2026-09-01), recorded at close.
- **Owner:** **Engineering lead** to price it against the real ceiling; due at **Epic 8 close**.
- **Severity:** MEDIUM.
- **Status:** OPEN.

**The gap.** The projection admits up to 256 chains of up to 64 entries, and the registration seam issues
one `asset` request per distinct carried key in a bare loop down the single engine channel. A document
with many carried faces therefore issues many concurrent full-font-byte transfers. Release flips the
seam's `active` flag so nothing is **added** after supersession — the correctness half is closed — but the
fetches still complete and still transfer their bytes.

**What discharges it.** Either a measured argument that the real ceiling is small enough to ignore, or a
bounded fan-out plus an abort signal threaded into the asset request.

---

### DW-97 — the hand-written comment stripper does not skip regex literals, so one regex containing a double slash silently truncates its line for every scan routed through it

- **Deferred by:** Story 8.4a (2026-09-01), recorded at close.
- **Owner:** **The first story that adds a regex literal containing `//` to designer source** — that is
  the trigger, and it is checkable by grep. Otherwise **Epic 8 close.**
- **Severity:** LOW while latent; **re-price to MEDIUM the moment the trigger fires**, because it weakens
  three guards at once and does so silently.
- **Status:** OPEN.

**The gap.** The stripper tracks string and template-literal state but not regex-literal state, so a
pattern such as `/https:\/\//` reads as the start of a line comment and the rest of that line is dropped
before any prohibition sees it. No designer source triggers it today. But the violation scan, the
runtime-registration site scan and the font-family position census **all** now route through it, so the
first source that adds such a regex quietly weakens all three.

**What discharges it.** Regex-literal state in the stripper, with a fixture proving a `//`-bearing regex
survives it — or the AST approach the repo's ownership contract test already demonstrates.

---

### DW-98 — the new test stubs install a page font set by a spelling the repaired prohibition does not match, and the prohibition file scans the test corpus

- **Deferred by:** Story 8.4a (2026-09-01), recorded at close.
- **Owner:** **Epic 8 close.**
- **Severity:** LOW. Not a live hole — a test stub is not runtime registration — but an **undisclosed
  blindness** in a rule this very story revived.
- **Status:** OPEN.

**The gap.** The authority contract scans production, tests **and** e2e. Story 8.4a's own new tests
install a global face constructor and a document font set, yet neither matches `\bdocument\.fonts\b` nor
`\bnew FontFace\b`, because they are written as a property definition and a locally declared stub class.
So the revived rule is blind to the mechanism **as this story itself writes it**. That is defensible —
stubbing is not registering — but it was undisclosed, and a real registration written in the same shape
would also pass.

**What discharges it.** Either a stated, tested exemption for the stub spelling, or a detector that
answers to the mechanism rather than to two literal spellings.

---

### DW-99 — stub teardown deletes the font-set globals rather than restoring the prior property descriptor

- **Deferred by:** Story 8.4a (2026-09-01), recorded at close.
- **Owner:** **The next test-environment upgrade** (a jsdom major, or a move to happy-dom) — that is the
  trigger. Otherwise **Epic 8 close.**
- **Severity:** LOW. No live impact: the current test environment provides neither global.
- **Status:** OPEN.

**The gap.** Both teardowns call a delete on the global face constructor and on the document's font set
rather than capturing and restoring the descriptor they replaced. In any environment that **does** provide
them, that permanently removes both for every later test in the same worker — a failure that would appear
as an unrelated test breaking, some distance away.

**What discharges it.** Capture the prior property descriptor and restore it, which is the correct shape
and is no larger than what is there.

---

### DW-100 — `s1VisibleBytes` is not yet a reproducible figure, and Story 8.4d is about to make it a gate

- **Deferred by:** Story 8.4b's close (2026-09-01), found while checking the spec's stale literal.
- **Owner:** **Story 8.4d — *the size budget is a number something checks*** (D-8.4.24 gives 8.4d the
  job of making this figure enforceable, which is precisely what an irreproducible input breaks).
- **Severity:** MEDIUM — it is LOW today, because no gate reads the figure; it becomes the gate's own
  flakiness the moment 8.4d lands, and a flaky size gate is worse than prose.
- **Status:** OPEN.

**Measured at 8.4b's close.** Building both arms in one environment — the story's two production files
reverted to `2ded2e3`, then restored — `s1VisibleBytes` read **12,423,049** at baseline and
**12,423,049** at HEAD. Story 8.4b moves it by **exactly zero**, and structurally so: the figure sums
the S1 **labelled rows**, and the CSS bundle is not one of them, so a story that only adds stylesheet
text cannot move it however much text it adds. `cachedBytes`, which *does* include the bundle, moved.

**The gap.** Three different figures are on record for the same quantity and none of the other two is
right: Story 8.4b's Code Map records **12,423,974** (+925 against the baseline it claims to record) and
its build report claims **12,426,422** both before and after (+3,373 against HEAD) — a correct
*conclusion* resting on a wrong *number*. A third read taken in the same session from a `dist` not
produced by a clean `npm run build` gave **12,423,167**. `verify:offline:red` was checked and does
**not** rewrite the manifest, so that variance is unexplained. A figure that reads three ways depending
on how `dist` got into its current state cannot be hard-gated as it stands.

**What discharges it.** Pin how the figure is produced before comparing it — a clean output directory,
a named command, and the figure read from that build alone — and state in the gate which rows the
number sums, so a later reader cannot mistake `cachedBytes` for `s1VisibleBytes`. A regression fixture
proving two consecutive clean builds agree would close it outright.

**UPDATE, Story 8.4c's close (2026-09-01) — the discharge condition is MET, and the mechanism on
record was WRONG.** DW-100 asks for "a regression fixture proving two consecutive clean builds agree".
**Two consecutive clean `npm run build` runs at `4f5925a` agree byte-for-byte**: identical engine wasm
sha256 `f1f195be93b0…`, identical raw 23,061,666, identical bytes on all five S1 rows, identical
`s1VisibleBytes` = **12,423,631**. **The figure is NOT irreproducible, and the variance is not Brotli
nondeterminism** — which is what Story 8.4c's build report attributed it to.

**The actual mechanism, shown rather than assumed.** `strings` on the built wasm:
`vcs.revision=4f5925af86abdebd7dc70b6cbadc51f2a9cadbff`, `vcs.time=…`, `vcs.modified=false`.
`build-wasm.mjs` runs `go build` with `-buildvcs` at its **default**, so the engine binary **embeds the
commit SHA, the commit timestamp and the working tree's dirty flag**. Proved at ONE commit with the Go
source held fixed: clean tree → `f1f195be93b0…` with `vcs.modified=false`; one tracked file touched →
**`570a1579faf1…` with `vcs.modified=true`**.

**Why this outranks the flakiness framing.** The engine row is **7,224,421 of 12,423,631 — 58%** of
`s1VisibleBytes`. So the figure moves **on every commit, deterministically, and again on whether the tree
was clean** — regardless of whether one byte of payload changed. Every reading on record is consistent
with this: all of them differ **only** in the engine row, the three Noto rows being constant at
226,026 / 24,872 / 4,948,312.

**What now discharges it, and it is concrete.** Build the wasm with **`-buildvcs=false`** (optionally
with `-trimpath`), or exclude the stamped bytes from the measure. Until then no threshold on
`s1VisibleBytes` can survive a single commit, and 8.4d must not pin one.

**A second, independent defect in the same figure, found at the same close.** `s1VisibleBytes` sums only
the four `delivery: "cached-asset"` S1 rows and **there is no IBM Plex row**, so it is *structurally
incapable* of seeing the 490,280 bytes of fonts Story 8.4c added. **A budget gate on it would have gone
green through that story.** Confirmed by exact arithmetic, not asserted. 8.4d must gate on a total that
includes every cached asset (`cachedBytes` does) or add rows for what it is meant to bound.

**UPDATE, Story 8.4g's close (2026-09-01) — the `-buildvcs=false` discharge condition above is MET, the
mechanism above is CORRECTED IN ONE HALF, and DW-100 STAYS OPEN on its second defect.**

**The mechanism, corrected.** The Story 8.4c UPDATE states the stamp as "the commit SHA, **the commit
timestamp** and the working tree's dirty flag". **The timestamp half is refuted on its premise** (D-8.5.7):
`vcs.time` is the timestamp of the **revision**, not of the build — Go stamps the commit's own time — so
two builds at one commit carry the **identical** value however far apart they run, and it can move
nothing. **The operative input is `vcs.modified`**, which Go derives from `git status`, where **a single
UNTRACKED file is enough**. That is the sharp part, because **this pipeline writes untracked files into
the tree** — halt files, result files, spec files — so a run that wrote an artifact between two
measurements **changed the stamp it was measuring**.

**The discharge condition is satisfied verbatim, and re-measured rather than assumed.** All six probes at
`873757f`, **in the main checkout** (`/Users/panitw/Projects/folio`), `cd folio-go && GOOS=js GOARCH=wasm
go build [-buildvcs=false] -o <tmp> ./wasm/cmd/engine`, with the tree state recorded against every figure:

| probe | tree state | `-buildvcs` | wasm sha256 | `vcs.` settings in the bytes |
|---|---|---|---|---|
| A | `git status --porcelain` **empty** | default | `68ee2569…c03117` | 3 — `vcs.modified=false` |
| B | **empty**, repeat of A | default | `68ee2569…c03117` — **identical** | 3 — `vcs.modified=false` |
| C | one **stray untracked** file | default | `1de602ca…3cc470` — **DIFFERS** | 3 — **`vcs.modified=true`** |
| D | **empty** | **`false`** | `ff324971…d90f732` | **0** |
| E | one **stray untracked** file | **`false`** | `ff324971…d90f732` — **IDENTICAL** | **0** |
| F | one **modified tracked** file (`README.md`) | **`false`** | `ff324971…d90f732` — **IDENTICAL** | **0** |

Full digests: A/B `68ee2569ded91b9856593beb6474a5a8a0e6d601906b1fbc028df36c39c03117`, C
`1de602ca69341c0767680417cbdccac9d398995d4f54396f38117eaf2e3cc470`, D/E/F
`ff324971091afd641151d1658020852ad0120687c225e5760b05888d4d90f732`. The probe file was removed and
`git status --porcelain` verified empty afterwards; `README.md` was restored (`md5` unchanged at
`078d7d80d518d54af2fc04fb270d46b8`).

**Read the stamp from the raw bytes, never from `go version -m`.** On a wasm binary that tool answers
`unrecognized file format`, so grepping its output for `vcs` prints nothing **because the tool failed** —
its silence is not evidence of absence, and a check built on it would have been vacuously green.

**Landed as** `folio-designer/scripts/build-wasm.mjs` passing `-buildvcs=false`, guarded at its point of
use by `assertNoVCSStamp` from the new `folio-designer/scripts/wasm-vcs-stamp.mjs` (red-proved: with the
flag removed, `npm run build:wasm` exits 1 naming all four settings and never publishes the fingerprinted
wasm) and covered by `folio-designer/scripts/wasm-vcs-stamp.test.mjs`. **THE REPRODUCIBILITY HALF OF
DW-100 — and only that half — is discharged by Story 8.4g's commit **`c985b9c`**. (The SHA is written
here at that story's CLOSE: the build could not record it, because a commit cannot self-reference its
own hash in its diff and this one was amended twice mid-dispatch besides.) **DW-100 ITSELF IS NOT CLOSED** — see the
final paragraph of this entry for what remains and who owns it.

**The provenance loss is a DELIBERATE TRADE, not a free win, and it is stated rather than assumed.**
`-buildvcs=false` removes `vcs.revision`, so the engine binary no longer self-identifies its commit. Two
things already cover that: the offline release manifest carries `releaseId` and `pageId` derived from
asset hashes, which identify the **bundle** more precisely than a commit does, and **AD-22** pins an exact
`toolchain` directive so the compiler behind these bytes is a release event rather than an ambient fact.

**DW-100 STAYS OPEN, and this is the half that is left.** Only the reproducibility defect is discharged.
The **second, independent defect recorded above is untouched**: `s1VisibleBytes` sums only the four
`delivery: "cached-asset"` S1 rows, **there is no IBM Plex row**, and the figure is therefore structurally
blind to the 490,280 bytes of fonts Story 8.4c added. That remains **Story 8.4d's**, and 8.4d must still
not pin a threshold on a figure that cannot see what it is meant to bound.

---

### DW-101 — twelve executable Playwright specs exist, run, and are never run by CI: a whole capability believed absent by everyone

- **Deferred by:** the run orchestrator (2026-09-01), on discovering it had told the engineering lead
  and two plan gates the opposite. **Ruled by the lead the same day (D-8.4.25e)** and **ranked ABOVE
  the font residual (DW-35)**.
- **Owner:** whoever wires CI. **It must land BEFORE, or in the same unit as, the executed browser
  assertion Epic 8 now owes** — see the ordering constraint below.
- **Severity:** **MEDIUM-HIGH.** The defect is not the missing coverage; it is that **the coverage
  exists and nobody runs it.**
- **Status:** OPEN.

**The gap, measured.** `folio-designer/e2e/` holds **12 Playwright spec files** — the application
shell, the engine worker, component binding and manipulation, component properties, image assets,
local file actions, offline update, preview parameters, sample data, the table editor and a
browser-native round trip. `package.json` has `"test:e2e": "playwright test"`. `playwright.config.ts`
has a real `webServer` running `npm run build && npm run preview`. `@playwright/test` is installed.
**`.github/workflows/ci.yml` runs only `test:e2e:compile`** — `tsc --noEmit`.

**It was RUN, not inferred.** The suite **built, started the preview server and executed all four
specs** in the file tried, failing only on
`browserType.launch: Executable doesn't exist … chromium_headless_shell-1208` — the browser cache
holding 1217/1223/1228 but not that build.

**This is a GUARD THAT DOES NOT RUN** — the same class as an unreferenced Makefile target or a
job-level default that excludes a module. **It is why an entire capability could be believed absent by
the orchestrator, by the engineering lead and by two plan gates**, and why a false statement of it
reached the Design Notes of two shipped stories.

**THE OBLIGATION THIS ENTRY NOW CARRIES, WRITTEN OUT RATHER THAN REFERRED TO (added at Story 8.4e's
close, 2026-09-01).** Until today this entry named *"the executed browser assertion Epic 8 now owes"*
twice and **nowhere said what it is**. Its content lived in **DW-35's prose** — and DW-35 closed today.
Once closed, nothing in an OPEN entry defined the obligation, only referred to one; a reader sweeping
open work would have found a dangling reference and no owed thing. **That is the ownerless-drift shape
this epic already suffered once, to DW-35's own residual, and it is recorded here so it does not
happen a second time by the same mechanism.** Stated in full, from **D-8.4.25(d)**:

> **Epic 8 does not close without ONE executed browser assertion covering carried-face rasterization.**

**Scope, as Epic 8 leaves it.** The assertion is owed for the **carried** population (D-8.4.25(d), the
face a document carries, Story 8.4a) and — since Story 8.4e put the engine's `FontSet` name on the wire
for the **shipped** population too — the same gap now exists on both arms of one mechanism. Neither is
observable today for the same single reason: **jsdom applies no stylesheet and loads no font, and
`test:e2e:compile` is `tsc --noEmit`.** What every gate in Epic 8 proves is that the right face name
reaches the element and the element asks for it. **That a glyph was rasterized with that face is
proved nowhere in this repository.**

**Owner:** this entry, and it is not dischargeable separately from the CI wiring above — an assertion
added to a suite CI does not run executes once, locally, and never again. **Due:** the Epic 8 boundary
gate, which is a gate and not an event (D-000.73).

**And it is ONE policy pass with DW-103 (D-8.4.34)**, which owes the same shape for
`TestShippedFacesReproduceFromUpstream`: *a real check exists, it works, and CI does not run it.*
Three obligations, one wiring change — do not fix them separately or they will produce three different
answers to one question.

**THE ORDERING CONSTRAINT IS THE POINT (D-8.4.25e).** **Wire CI to execute the suite BEFORE, or in the
same unit as, adding the executed browser assertion Epic 8 owes.** An executed assertion added to a
suite CI does not run **would execute once, locally, and never again** — **reproducing the exact
category error being corrected, one layer up.** Do not add the observer before arranging for anyone to
watch it.

---
### DW-102 — the chrome's monospace slot lost its CJK coverage, and the finding that said so was rejected on a location that was not where it was true

- **Deferred by:** Story 8.4c's close (2026-09-01), re-opening a review finding the build rejected.
- **Owner:** **the design system's font stack — THE OWNER'S.** NOT Story 8.4d: this is a coverage
  question, not a size-budget one, and 8.4d's remit under D-8.4.24 is the budget. **And NOT Story
  8.4e**, decided at 8.4e's close (2026-09-01) rather than left standing as a candidate on a story
  that has ruled itself out — the ownerless-drift failure this epic already had once, to DW-35's own
  residual. Five selectors, all one way: **(a) Scope** — D-8.4.26 scopes 8.4e to *per-fragment
  shipped-face attribution on the wire*, while this entry is about which face `--font-mono` resolves
  to; different subject, different surface. **(b) A standing ruling forbids the remedy on that path**
  — the fix is a chrome token edit, and D-8.4.14 ruled the canvas fix must *"edit no chrome token at
  all"*; `canvas-font-stack.test.ts`'s assertion that no `--font-*` token names an engine face
  **reddens** if one does, so folding this in would require breaking a guard 8.4e is otherwise obliged
  to keep. **(c) Measured separation** — this entry's live consumers are `.document-name` and
  `.property-value`, chrome elements; 8.4e touches `.canvas-text-fragment` only, and the two share no
  selector, no token and no code path. **(d) Reachability (the D-000.65 test)** — 8.4e does not make
  this condition newly reachable; **Story 8.4c did**, when the mono binary changed, and by the rule
  this run uses to place a defect it belongs where it became reachable. **(e) It carries a decision
  above a builder's authority** — the discharge is *either* ship a CJK fallback face, a bundle-weight
  commitment that collides with Story 8.4d's budget, *or* record a deliberate decision that chrome CJK
  falls to the system face. Both are the owner's call, which is what this bullet already half-said.
- **Severity:** **MEDIUM.** It degrades to a system face rather than to tofu, so nothing is unreadable;
  but it is a coverage regression this story caused, in the interface's most-used type token.
- **Status:** OPEN.

**Measured at close with a cmap read over both faces.** `IBM Plex Mono` maps **0** of the 20,992 code
points in U+4E00–U+9FFF. `Noto Sans SC` — the file the family name `IBM Plex Mono` resolved to before
Story 8.4c — maps **20,976**. Seven `--type-*` tokens resolve through `--font-mono`, and after 8.4c the
face behind them has no CJK at all.

**The rejection was sound where it looked, and that is the transferable part.** The review raised this
and rejected it, refuting it at `--type-page-mono`'s three consumers — correctly: verified at close, the
component box renders the text/image paints, the literal `Table`, or nothing; the image placeholder's
children are four hardcoded English strings; the dimension readout is `points(w) × points(h)`. **No
document-derived text reaches a `--type-page-mono` element.** But the finding's *headline* was about
`--font-mono`, and `--type-page-mono` is one of seven tokens that resolve through it. The claim is true
at consumers the rejection never examined: `.document-name` renders the open document's title
(`App.tsx:850`), and `.property-value` is the prose input an author types a text component's content
into (`App.tsx:1230-1231`), as are bare `input`/`select`/`textarea` (`App.css:293-294`) — all
`--type-mono`.

**The shape to remember, because it is a rejection pattern rather than a one-off.** *A refutation that
is airtight at the cited location can leave the finding's actual claim standing*, when the citation
names a narrower thing than the claim does. The reviewer cited one token; the claim was about the
variable seven tokens share. **Audit a rejection against the claim's own scope, not only against the
line it points at.**

**Why it is recorded and not fixed.** The remedy is adding a CJK fallback to the design system's mono
stack — an edit to the token file whose *untouched-ness* is the premise Story 8.4c's sequencing rested
on (the family names do not change, so no CSS surface moves). Folding it in would have made the story's
own premise false. It is also genuinely degraded-not-broken: the browser's own fallback finds a system
CJK face on any real OS, and this is the same class as the `--font-sans` narrowing in 8.4c's frontmatter
`deferred` (Greek, Cyrillic, Latin Extended-B).

**What discharges it.** Either extend `--font-mono` with a CJK fallback that is *shipped* rather than
assumed present on the reader's machine, or record a deliberate decision that chrome CJK falls to the
system face — with the reasoning written down, so the next reader does not re-derive it as a bug.

---

### DW-103 — a real check exists, it works, and CI does not run it: ONE policy pass, not two

- **Deferred by:** the engineering lead (2026-09-01, **D-8.4.34**), merging this with **DW-101**.
- **Owner:** whoever wires CI. **DW-101 and this are the SAME defect wearing different clothes** —
  *"a real check exists, it works, and CI does not run it."* **Fixing them separately will produce two
  different answers to one question.**
- **Severity:** MEDIUM-HIGH.
- **Status:** OPEN.

**The invariant, stated rather than the mechanism: a change to `folio-go/fonts/` must not be able to
merge without `TestShippedFacesReproduceFromUpstream` having run.**

**Two facts bound it.** The check's inputs are **gitignored and large**, so provisioning them on
**every** CI run is disproportionate for an artifact that has changed **twice in the project's life**
(`3373dac`, `4b797d4`). And **leaving it entirely unrun is how it went dark.** A **path-filtered**
job — provision sources and venv, run the check, **only** when a change touches `folio-go/fonts/` or
the generator — is proportionate.

**The other half is DW-101:** 12 executable Playwright specs that CI runs only `tsc --noEmit` over.

**Also owed here (D-8.4.34):** the interpreter-selection fix — preference order
`FOLIO_FONTGEN_PYTHON` → `.fontgen-venv/bin/python` (if it exists **and** imports fontTools) →
`python3`, so an explicit override still wins. **Auto-preference is safe for THIS test specifically
because it cannot go green by finding less:** green requires sources **and** a working interpreter
**and** three matching hashes **and** the `3 of 3` witness, and every other state is a loud, distinct
red. It cannot manufacture a pass — only stop suppressing a real run. **And the interpreter failure
must name its own remedy** — `.fontgen-venv/bin/python` and `FOLIO_FONTGEN_PYTHON` in the message
text. *The whole cost of this incident was that the next reader could not tell a wrong interpreter
from absent sources at a glance; the message is where that is fixed, not the register.*

---
### DW-104 — the `maxCanvasPropertyString` probe list is a hand-list whose count lives only in prose, and it has now missed a new site for the THIRD time

- **Deferred by:** Story 8.4e's close (2026-09-01), on measuring the guard rather than reading it.
- **Owner:** **the next story that adds a `maxCanvasPropertyString` site in `page_setup.go`** — a
  forcing function rather than a date, because that is the only moment the drift is cheap to see.
  Failing that, the Epic 8 boundary gate.
- **Severity:** LOW — no wire breach follows; every site still enforces the bound. What has failed is
  the guard's ability to say so.
- **Status:** OPEN.

**Measured at Story 8.4e's close, at `24ca8f6`.** `TestCanvasIdentifierBoundsStillRefuseAtFiveHundredAndTwelve`
in `folio-go/canvas_body_text_bounds_test.go` opens by declaring **"ALL EIGHT"** and defines the eight
as *"the sites `grep -n maxCanvasPropertyString folio-go/page_setup.go` reports, minus the constant's
own declaration and its doc comment."* Run that grep today and it reports **NINE** enforcement sites:

`:593` chain name · `:645` chain entry · **`:1547` a text fragment's shipped face name (Story 8.4e)** ·
`:1656` visibleIf · `:1662` table.bind · `:1707` style.fontFamily · `:1739` color · `:1745` background ·
`:1760` border.color

The probe table lists **eight**, and `:1547` is not among them. **The guard's own stated invariant is
therefore false in the file that states it** — which is precisely the drift its comment says a count
exists to catch. The comment already records this happening once before (*"the eighth is Story 8.1's
chain ENTRY: it was a site this list did not cover from the day it was added"*). Counting that, the
count in this comment, and this omission, the hand-list has now been wrong **three times**.

**Why `:1547` was not simply added as a ninth probe, and why that is the interesting half.** It is
**unreachable by construction** — `Canvas(t)` runs `canvasFontChains` → `projectFontChainEntry`, which
refuses the same string over the same bound at `:645`, and returns before `CanvasWithTextPaint` ever
calls `addCanvasTextPaint`. For a shipped face `fragment.face` **is** that entry's `Face`. **Measured
at Story 8.4e's close, independently of the build's own measurement: deleting `:1547` outright leaves
the Go suite at exactly 1815 pass / 2 fail / 5 skip with the same two failing identities — byte for
byte what it is unmutated.** So the site cannot be probed by a document, and this list's every entry is
a document-driven probe. That is a legitimate exception; **what is not legitimate is an exception
nothing records**, which is the state the story left it in and the state corrected in the comment at
this story's close.

**The remedy is to stop hand-counting.** A `grep`-derived assertion — enumerate the enforcement sites
out of `page_setup.go`, assert the probe table covers all of them **minus a NAMED exception list with
a stated reason per exception** — replaces a prose count with a derivation, which is the same move
`canvas-font-stack.test.ts` already made when it stopped restating `fonts.Shipped()`'s three names and
started reading them. **A count is lossy; a set difference is not** (Design Note 7 of Story 8.4e, and
D-8.4.34 one level up in the process). Deferred rather than done here because it is a **new guard** in
a story already reviewed and committed, and a new guard needs its own mutation proof.

---

### DW-105 — the engine wasm is still a function of the CHECKOUT PATH: same commit, same flags, both unstamped, different bytes

- **Deferred by:** Story 8.4g's close (2026-09-01), found while re-measuring the fix it landed — a
  **newly measured residual**, not a restatement of DW-100.
- **Owner:** **Story 8.4d — *the size budget is a number something checks*** (it inherits every input
  that can move the figure it is asked to gate), or whoever first needs two checkouts of one commit to
  agree.
- **Severity:** LOW-MEDIUM — it moves no shipped behaviour and no golden. It is exactly high enough to
  invalidate a **comparison**, which is the only thing byte figures are for.
- **Status:** OPEN.

**Measured against an ISOLATED POSITIVE CONTROL. Re-anchored at Story 8.4g's close to `c985b9c`,
the story's final commit:** the original measurement was taken at `4f8516a`, which was amended away
twice during the dispatch (`4f8516a` -> `77cd80e` -> `c985b9c`) and is therefore **unreachable from
`origin/main` and unresolvable in any other clone** — an anchor nobody else could check. The MAIN
digest below was **independently re-measured at `c985b9c` at that close and is unchanged**, which is
itself consistent with the finding: no `.go` path differs between those commits. The first attempt at this measurement
compared the main checkout to a **linked worktree**, and that comparison was CONFOUNDED: a linked
worktree differs in **two** ways at once — its absolute path **and** its `.git` being a file rather than
a directory. So it could not attribute the difference to the path, and a conclusion drawn from it would
have been inference dressed as measurement. Replaced by a control that varies **only the path**: a full
`rsync -a` copy of the checkout (excluding `node_modules`, `dist` and build caches) to a different
absolute path, with **`.git` a real DIRECTORY in both arms**, so `-buildvcs` behaves identically on each
side. Invocation, both arms: `cd <root>/folio-go && GOOS=js GOARCH=wasm go build -buildvcs=false -o <tmp>
./wasm/cmd/engine`. Tree state, **identical in both arms by construction** (the copy is an `rsync` of the
original): four modified tracked files — the story spec and this story's three designer scripts, none of
them Go source. Both binaries carry **ZERO** `vcs.` settings.

| arm | checkout root | wasm sha256 |
|---|---|---|
| MAIN | `/Users/panitw/Projects/folio` | `ff324971091afd641151d1658020852ad0120687c225e5760b05888d4d90f732` |
| COPY | a scratch path outside the repo | `8f8fe16b9786a7d77670716813f39fe9e38502bc5016924c885a653ffa5fe7ba` |

**Same commit, same flags, same tree state, `.git` a directory on both sides, both unstamped — DIFFERENT
BYTES.** The path is the only remaining difference, and it is therefore the cause.

**The mechanism was confirmed directly, not inferred from the digests.** Each binary embeds **87
occurrences** of **its own** checkout's absolute source-path prefix — e.g. `<root>/folio-go/internal/geom/scale.go`
— 87 in the MAIN binary naming the main root, 87 in the COPY binary naming the copy's root. Go embeds
absolute source paths, so the checkout's location is a genuine compile input.

**The digest of any non-main arm is itself path-dependent, which is the finding rather than an
inconsistency.** A review pass ran this same control at a **third** path and measured a **third** COPY
digest (`fbfe2bd3b06742e9ed0356f95ff468c841e590e061fe1558bdc8fa00abd0f02b`) against the same unchanged
MAIN digest. Three paths, three digests, one commit: **only the MAIN figure above is reproducible by
anyone else, and only by someone whose checkout is at that same path.**

**The consequence is about COMPARISON, not correctness.** No shipped behaviour moves and no golden moves.
But **a baseline measured in one checkout is not comparable to one measured in another**, so any process
that measures the two arms of a change in different directories is measuring the directory.

**A SEPARATE finding from the same investigation, recorded on its own terms because it is NOT the
evidence for the above.** In a **linked worktree `.git` is a FILE, not a directory**, and Go's default
`-buildvcs=auto` therefore **silently omits the VCS stamp entirely**. Measured: in a linked worktree the
**default** build and the `-buildvcs=false` build produce the **same** digest and **zero** `vcs.`
settings. A stray untracked file changes nothing there, so the tree-state discriminator run in a worktree
looks like a refutation **when it is a non-measurement.** Any future determinism probe must run in the
main checkout, with a positive control to prove the probe can move at all.

**What discharges it.** `-trimpath` is the candidate remedy. It was **deliberately out of scope** for
Story 8.4g: DW-100 lists it as optional, that story's AC named exactly one flag, and a second variable in
the same measurement would have made the re-measurement unattributable. Landing it needs its own
measure → change → **re-measure** unit, including a check that no golden PDF digest moves.

---

### DW-106 — the determinism guard watches ONE CAUSE (the VCS stamp), not THE PROPERTY (two builds of one commit agree)

- **Deferred by:** Story 8.4g's build, **filed into this register at that story's close** — the build
  recorded it in its spec frontmatter only, where nothing sweeps it.
- **Owner:** **Story 8.4f** — it owns `folio-designer/scripts/verify-offline-release.mjs` next, which
  is where this belongs and where it is cheap. See **DW-107**, which shares that surface exactly.
- **Severity:** MEDIUM — no shipped behaviour is wrong. What is missing is the ability to *notice* the
  next input of this kind.
- **Status:** **CLOSED by Story 8.4f (2026-09-02)**, commits `7a18079` and `68c548e`.

  **What discharged it.** `assertEngineWasmIsTreeIndependent()` in
  `folio-designer/scripts/verify-offline-release.mjs`, run inside the existing `wasmWitness` block —
  reached by `npm run verify:offline:wasm`, and deliberately **not** by `npm run build`, because
  `build:wasm` is a dependency of `typecheck`, `test` **and** `build` and a second engine build there
  would tax every designer gate. It builds the engine twice, writes one stray untracked file between the
  runs, and requires the two digests to agree, reporting both so a failure names what moved.

  **The first cut of it could not tell tree-independence from a probe that perturbed nothing**, because
  both arms build with `-buildvcs=false` — which closes the only measured tree→binary channel, so the
  stray file had no route into the output. That is this run's signature defect (an all-clear
  indistinguishable from a couldn't-look) reproduced inside the fix for it. Caught in review and patched
  in `68c548e`: the check now asserts the probe is **visible to `git status --porcelain`** before relying
  on the second build, and fails in the file's own voice if git cannot be run at all.

  **⚠ THE PROBE FILENAME MUST NEVER BE ADDED TO `.gitignore`.** Its visibility to git **IS** the
  perturbation — Go derives `vcs.modified` from `git status` — so ignoring it would make both builds see
  an identical tree and turn the whole check vacuous. The code carries this warning in place. See
  **DW-113**, whose obvious remedy is exactly that wrong one.

  **Verified independently at Story 8.4f's close, not taken on report.** With
  `.folio-tree-state-probe-*` appended to `.git/info/exclude` (never `.gitignore`),
  `npm run verify:offline:wasm` exits **1** with `the tree-state probe … is invisible to git, so both
  builds below see the same tree and their agreement would assert nothing`. The exclude file was restored
  byte-identically afterwards.

  **ITS LIMIT, RECORDED RATHER THAN OVERCLAIMED — and this is the half a later reader is most likely to
  get wrong.** This check **would NOT have caught DW-105.** It holds the checkout **path** fixed by
  construction, and DW-105 is **path dependence**: the same commit built from three different paths gives
  three different binaries, each embedding **87 occurrences of its own absolute root**. Both arms also
  still build with `-buildvcs=false`, so the one channel ever measured here is already closed and this
  check is *expected to be quiet*. **Its value is the NEXT tree-dependent input, not the last one.**
  **DW-105 remains OPEN.**

**The claim.** AC2's property is `sha256(build_clean) == sha256(build_stray)`. Every automated artifact
Story 8.4g added instead tests for the **absence of four ASCII needles** in the raw wasm. Nothing in the
repo compares two builds' digests, so a second tree-dependent input would leave every gate green while
the story's headline claim is false. Raised independently by the blind-hunter and intent-alignment
layers.

**This is the shape the lead recorded against its own ruling at D-8.4.30** — *"a guard keyed on a proxy
rather than its purpose is a defect; I wrote one into my own ruling."*

**Assessed at Story 8.4g's close, and the assessment cuts both ways.** It **matters**, and it is not
hypothetical: **DW-105**, filed by that same story, records a measured input of exactly this kind (the
checkout path) that the needle guard cannot see. **But the obvious remedy would not have caught DW-105
either** — two builds compared *within one checkout* agree today, because the path is held fixed by
construction. So the property check's real value is catching the **next** tree-dependent input, not the
one already known; it should be filed as that, not oversold as closing DW-105.

**Why it is not simply added where the guard is.** `build:wasm` is a dependency of `typecheck`, `test`
**and** `build`, so a second engine wasm build there would tax every designer gate. The property check
wants a home that runs once — the release verifier or a CI job — which is the same home DW-107 needs.

**What discharges it.** A check that builds the engine wasm twice at one commit, with the tree state
deliberately differing between the two runs, and requires the digests to agree — with the digests
reported, so a failure names what moved.

---

### DW-107 — both of Story 8.4g's red-proofs exist only as PROSE and cannot be re-run

- **Deferred by:** Story 8.4g's build, **filed into this register at that story's close** — spec
  frontmatter only, where nothing sweeps it.
- **Owner:** **Story 8.4f** — it lands its own assertions in `verify-offline-release.mjs` next and owns
  that surface. **Cheap if routed now, expensive once that story has closed.** Shares its home with
  **DW-106**; the two should be taken together.
- **Severity:** MEDIUM — the guard works today. The evidence that it works is the perishable part.
- **Status:** **CLOSED by Story 8.4f (2026-09-02)**, commit `7a18079`.

  **What discharged it.** `proveVCSStampGuardDiscriminates()` in
  `folio-designer/scripts/verify-offline-release.mjs`, called from `runRedProofs` and therefore carried by
  `npm run verify:offline:red`. It builds the engine with `-buildvcs=false` **dropped**, and requires the
  detector to report **every** setting Go then stamps in (`vcs.revision=`, `vcs.time=`, `vcs.modified=`
  and the bare `build\tvcs=` record) — mutate, observe a failure, hold it to the guard's **own** message,
  restore. It deliberately does **not** go through `redProof`, because that harness mutates `dist/` and
  re-runs `verifyOfflineRelease` while the guard under proof lives in `build-wasm.mjs`; the probe builds
  to a temp file, so `src/generated/` is never touched and there is nothing to restore by hand. It fails
  **loudly** rather than passing vacuously when the checkout has no `.git`.

  **The engine's compile argv moved to a single authority** (`ENGINE_BUILD_FLAGS` / `buildEngineWasm` in
  `folio-designer/scripts/wasm-vcs-stamp.mjs`, called by `build-wasm.mjs`), so neither this proof nor
  DW-106's re-types the argv it is proving something about — the same no-second-copy doctrine Story 8.4f
  imposes on the cache-asset bound.

  **Re-run at Story 8.4f's close:** `cd folio-designer && npm run verify:offline:red` → **exit 0** in
  183.33s, carrying this leg. The proof is no longer prose, and it can fail again.

**The claim.** `folio-designer/scripts/verify-offline-release.mjs:126` `redProof(name, mutate, expected)`
already backs `stale-wasm-byte`, `s1-total-mismatch`, `s1-delivery-fiction` and
`dictionary-witness-mismatch`, run by `npm run verify:offline:red`. It is the repo's established shape
for exactly this and gained nothing from Story 8.4g. That story's guard was red-proved **by hand**, and
recorded in prose, so it will rot silently.

**Measured at Story 8.4g's close: the proofs are real, which is precisely why they are worth keeping.**
Re-run by hand a third time at `c985b9c` — flag removed, `npm run build:wasm` exits 1 naming all four
settings; detector neutered, the synthetic stamped pair reddens; flag **and** in-script guard removed,
the build succeeds, emits a stamped wasm and **exactly** the real-population test reddens. Each of those
required a source mutation, a manual restore and an md5 check to undo. **A proof that costs that much to
re-run is one nobody will re-run.**

**What discharges it.** The flag-removed guard proof expressed as a `redProof` leg, so
`npm run verify:offline:red` carries it.

---

### DW-108 — Story 8.4g's dispatch regenerated `epic-8-context.md` and dropped scope constraints that the UNBUILT Stories 8.5 and 8.6 were to be built against

- **Deferred by:** Story 8.4g's build, **filed into this register at that story's close** — spec
  frontmatter only.
- **Owner:** **the orchestrator**, before dispatching Story 8.5 — that is the next story the loss could
  reach, and it has not been built.
- **Severity:** LOW — **downgraded at the close from the build's framing, because the authority was
  found.** See the correction below.
- **Status:** OPEN.

**What happened.** The workflow's step-1 regeneration rewrote `epic-8-context.md` wholesale inside a
build-flag commit. Dropped: the out-of-scope-by-decision list (bold/italic meaning, synthetic
emboldening/obliquing, variable-font axes, live font services, enumerating host-installed fonts, CJK
families in the embeddable catalogue), the catalogue procurement rules (static, single-instance,
prepared ahead of the build) and the chain-entry shape contract.

**CORRECTION MADE AT THE CLOSE — the conclusion held, the citation did not, and the citation was the
load-bearing half.** Story 8.4g's rejection of this finding, and its own deferral text, both assert that
every dropped constraint *"retains an authority in `epics.md`, the PRD or the architecture spine"*.
**Measured, and that is false for three of them:** variable-font axes, live font services / arbitrary
URLs, and enumerating host-installed fonts have **no** scope authority in those three files. (`epics.md`
does carry FR54's *"no host-installed font"*, but that is a **rendering** requirement, not the
constraint against **enumerating** host fonts in a picker — a different claim.) A reader checking the
three named sources would have concluded the constraints were lost.

**They are not lost. All six survive, verbatim, in a FOURTH source neither document names:**
`_bmad-output/specs/spec-fonts/SPEC.md` `## Non-goals` — *"No live font service. No Google Fonts API, no
arbitrary URL, no 'download on first use'."*, *"No host fonts. Faces installed on the authoring or
rendering machine are never enumerated or read."*, *"No synthetic bold or oblique, and no variable-font
axes."*, plus the container-format, save-time-subsetting and CJK-catalogue non-goals. The procurement
rule survives at `epics.md:3456` (*"static, single-instance"*) and the chain-entry shape contract at
`epics.md:2866` (`{"asset": "<key>"}`). **So the constraints sit in a spec kernel, which is a stronger
place than the regenerable cache they were dropped from.**

**Why it stays open anyway.** `epic-8-context.md` declares itself a regenerable cache and its compile
prompt instructs aggressive scoping, so this is designed lossiness over intact sources — but 8.5 and 8.6
have not been built, the next regeneration may or may not restore these lines, and the agent that builds
them reads the cache, not the kernel.

**What discharges it.** Confirm before dispatching Story 8.5 that its spec carries the catalogue
procurement rules and the out-of-scope list, **sourced from `spec-fonts/SPEC.md`** rather than from
`epic-8-context.md`.

---

### DW-109 — the ambient environment is an unrecorded input to the engine wasm build

- **Deferred by:** Story 8.4g's build, **filed into this register at that story's close** — spec
  frontmatter only.
- **Owner:** **Story 8.4d**, with **DW-105** — same question (what, besides the source, decides these
  bytes), and 8.4d inherits every input that can move the figure it must gate.
- **Severity:** LOW — no observed drift is attributed to it. It is a gap in the *documentation* of
  inputs, recorded so the set is not believed complete.
- **Status:** OPEN.

**The claim.** `folio-designer/scripts/build-wasm.mjs:44` passes `env: { ...process.env, GOOS, GOARCH }`,
forwarding `GOFLAGS`, `GOEXPERIMENT`, `CGO_ENABLED` and toolchain selection into the build. Story 8.4g
closed the tree-state input and **DW-105** records the checkout path, so the set of inputs that are *not
the source* is documented **incompletely** rather than wrongly.

**Mitigation already on record.** **AD-22** pins an exact `toolchain` directive and makes bumping it a
release event, which covers the compiler itself but not `GOFLAGS` or `GOEXPERIMENT` set in an operator's
shell.

**What discharges it.** Either record the environment alongside every byte figure, or pass an explicit
allow-list rather than `...process.env`. Best taken with `-trimpath` (**DW-105**) as one
measure -> change -> re-measure unit, since both change the same line for the same reason.

---

### DW-110 — `main.tsx`'s two startup decisions are still executed by no test AT THEIR CALL SITES: the gap was narrowed, not closed

- **Deferred by:** **Story 8.4f's close (2026-09-02)** — raised by the build's own review as a `medium`,
  patched, and **re-measured at close, where the patch was found to have narrowed the gap rather than
  closed it.** Filed here because the story's frontmatter is not the register.
- **Owner:** whoever lands the **executed browser assertion Epic 8 owes (D-8.4.25d)**, or the first story
  that gives this repo a harness able to import a module whose top level calls `void startObservation()`.
  **Not a local repair** — see below.
- **Severity:** MEDIUM. Nothing shipped is wrong. What is missing is the ability to notice when it stops
  being right.
- **Status:** OPEN.

**The gap, measured rather than argued.** Nothing in this repo imports `folio-designer/src/main.tsx`;
Vitest collects only `src/**/*.test.{ts,tsx}` and `scripts/**/*.test.mjs`; the 12 Playwright specs are
compiled and never run. So the two decisions that module makes about a parsed release payload — mapping a
rejected result to the `undefined` the lifecycle expects, and gating the dev-server bypass on the
absent-bootstrap reason alone — are executed by **no gate**.

**What Story 8.4f's review patch genuinely bought, and it is real.** Both decisions were extracted into
pure exported functions (`payloadForLifecycle`, `isDevBypassReason`) that Vitest now drives over the whole
rejection table, with the reason union held to a `Readonly<Record<S1PayloadRejection, true>>` so a future
unnamed reason is a typecheck failure. **Both have teeth, proved at close:** neutering the parser's
over-maximum check reddens **five** named tests, and re-widening the bypass predicate to `!result.ok`
reddens exactly the test that names the narrowing.

**What it did not buy, measured at Story 8.4f's close at `68c548e`.** The **call sites** in `main.tsx` are
still executed by nothing, and **the exact mutation the review demonstrated still ships green today**:

- Replacing the mapping with a bare `undefined` **and dropping the now-orphaned import** (which is what a
  real refactor would do) gives `typecheck` **exit 0**, `npm test` **exit 0 at 40 files / 409 tests**,
  `lint` **exit 0**, `test:e2e:compile` **exit 0**, `npm run build` **exit 0**, `verify:offline:red`
  **exit 0** — while shipping an app permanently reporting *"Offline cache unavailable"*.
- Reverting the bypass condition to the pre-story `import.meta.env.DEV && !payload` gives `typecheck`,
  `npm test` (40 / 409) and `npm run build` all **exit 0**.

**A trap worth recording for the next agent who tries this.** The *naive* mutation — neuter the
expression, leave the import — fails `tsc -b` with **TS6133 `'payloadForLifecycle' is declared but its
value is never read`**. That is an **incidental compile error, not a test detecting the defect**. A mutant
that will not compile proves nothing; the import must be dropped for the mutation to be faithful.

**Why it is a deferral and not a repair.** Closing it needs either a jsdom harness that can import a
module whose top level runs `void startObservation()` — a real design question, not a test file — or the
executed browser assertion Epic 8 already owes. Both are outside a story about a build-time bound.

**This is the THIRD consecutive story whose I/O-matrix row for this surface could not be tied to an
executed test.** The patch moved the untested surface from an expression to a call site. That is progress
and it is not closure.

**What discharges it.** An executed assertion — browser or harness — that the app hands the lifecycle the
parsed payload when the bootstrap is good, and that the dev bypass fires on an absent bootstrap and on
nothing else.

---

### DW-111 — the named rejection reason has no production consumer, so two causes still look identical to a user

- **Deferred by:** **Story 8.4f's close (2026-09-02)**. Raised **independently by THREE review layers** —
  the most-corroborated finding of that story — and **rejected on Design Note 5**, a note that rules it
  explicitly and **invites the disagreement in writing**. Filed here because a rejection can be
  scope-correct and still true, and the build has no way to carry that forward.
- **Owner:** the first story that gives the designer a diagnostic channel, or that is asked to
  distinguish these causes for a user or an operator. **Related to DW-93**, which is the same silence on
  the canvas rather than on the load screen.
- **Severity:** LOW as it stands, because Story 8.4f made the over-the-bound cause **unreachable in a
  shipped release** — the build now refuses one. Re-price if that gate is ever bypassed.
- **Status:** OPEN.

**The claim.** Story 8.4f gave every payload rejection its own name, and `asset-count-over-maximum` is
producible by that cause and no other. But **every rejection is then mapped to `undefined`** for the
lifecycle, and nothing logs, reports or renders the reason. In a production build a malformed bootstrap
and an over-bound bootstrap remain **observably identical** — same screen, same words, same retry button.
The named reason is currently observable only to tests and to the dev-server bypass predicate.

**Why the rejection was SCOPE-CORRECT, checked at close rather than accepted.** AC4 is phrased at the
**function-return** layer and is met there. The intent's **Block If** forbids changing what a user sees on
any path other than the over-the-bound path, and changing the load screen's copy is UI work in a story
about a build gate. And the "there is no diagnostic channel to write to" ground is **measured**: re-counted
at Story 8.4f's close, `folio-designer/src/` and `folio-designer/scripts/` contain **zero** non-test
`console.*` calls. The reason is also **not wholly discarded** — the dev-server bypass reads it and
changes behaviour on it, so it is load-bearing rather than decorative.

**Why it is still true, and therefore registered.** Design Note 5 states plainly that the two causes still
produce the same *screen*, and says the disagreement belongs with that paragraph. A finding three layers
raised, ruled on in a design note that invites contradiction, should live somewhere that is swept — not in
a rejected-findings list that nothing reads again.

**What discharges it.** A ruling on whether the designer gets a diagnostic channel at all, then either a
distinct message on the load screen or a recorded decision that the silence is intended, with the reason
named where an operator can reach it.

---

### DW-112 — the cache-asset bound is resolved from the WORKING TREE at verify time, not from the artifact

- **Deferred by:** **Story 8.4f's build**, recorded in that spec's `deferred:` frontmatter; **placed into
  this register at that story's close (2026-09-02)**, per the spec's own instruction that the reviewer
  places what the build files.
- **Owner:** the first story that verifies a release **detached from its source tree** — an extracted
  release, or a deploy-stage-only CI job. **Not reachable in this repo's current cadence**, where
  `verify:offline` always runs beside its own source tree.
- **Severity:** LOW.
- **Status:** OPEN.

**The claim.** `declaredCacheAssetBounds()` in `folio-designer/scripts/offline-release-contract.mjs` reads
`src/release-payload.ts` **from disk at the moment of verification**. Two consequences follow. Verifying a
`dist/` with no source tree beside it throws an **unframed Node `ENOENT`** rather than a stated
verification failure in the file's own `fail(…)` voice. And an **old `dist/` verified against a newer
`src/`** is silently checked against the **new** bounds, because the manifest carries no record of the
bound the release was actually built under.

**Why it was not fixed in Story 8.4f.** Reading the declaration rather than re-typing it is the whole point
of that story's AC5 — `npm run build` never runs Vitest, so a duplicated constant plus a tie test would
leave a drifted build green. The fix is not to stop deriving; it is to **stamp the resolved bound into the
manifest** so the artifact carries the number it was built under. That is a manifest-shape change, which
Story 8.4f was explicitly not scoped to make.

**What discharges it.** Record the resolved bound in the emitted manifest and have the verifier prefer the
stamped value, falling back to the source read with a stated message; and frame the missing-source case as
a `fail(…)` rather than letting `ENOENT` escape.

---

### DW-113 — the tree-state probe file survives a signal, and its obvious remedy is specifically WRONG

- **Deferred by:** **Story 8.4f's build**, recorded in that spec's `deferred:` frontmatter; **placed into
  this register at that story's close (2026-09-02)**.
- **Owner:** whoever next touches `folio-designer/scripts/verify-offline-release.mjs`, or the first run
  that actually trips it. **Cheap only there.**
- **Severity:** LOW — it costs a manual `rm`, but it trips this pipeline's own clean-tree gates, and a
  dirty tree **halts the next story's dispatch**.
- **Status:** OPEN.

**The claim.** DW-106's tree-independence check writes an untracked probe file into `folio-go/` between two
~40s engine builds and removes it in a `finally` block. **A signal does not run a `finally`.** A SIGINT or
SIGKILL landing in that ~40s window — an orchestrator kill, a Ctrl-C — leaves
`folio-go/.folio-tree-state-probe-<pid>` behind, and the next clean-tree gate fails on it.

**⚠ THE OBVIOUS REMEDY IS THE WRONG ONE, AND THIS IS THE WHOLE REASON THE ENTRY EXISTS.** Adding the
filename to `.gitignore` would silence the symptom and **make the check vacuous**: the probe's visibility to
git **IS** the perturbation being measured, because Go derives `vcs.modified` from `git status`. An ignored
probe means both builds see an identical tree and their agreement asserts nothing — which is exactly the
defect DW-106's review patch was raised to remove. The code now warns against it in place, and Story 8.4f's
close re-proved that the visibility assertion fires when the probe is hidden. **See DW-106.**

**The correct remedy** is a signal handler that removes the probe, which was out of proportion to Story
8.4f. Until then, a stray `.folio-tree-state-probe-*` in `folio-go/` after an interrupted
`verify:offline:wasm` is **safe to delete by hand** and must **never** be ignored.

**What discharges it.** A `SIGINT`/`SIGTERM` handler around the probe's lifetime, removing it on the way
out — with the `.gitignore` prohibition kept in the comment.

---

### DW-114 — the extension-class guard now sweeps the whole tracked repository, and today it reports nothing

- **Recorded by:** **Story 8.4h's build (2026-09-02)**, discharging **AC8** — the story's own obligation
  to state the widening's new coverage **by measurement** rather than to assume it inert.
- **Owner:** informational. Nothing to do; this is the baseline a later reader compares against.
- **Severity:** N/A — this is a MEASUREMENT RECORD, not an open gap.
- **Status:** RECORDED.

**What changed.** `folio-designer/src/font-binary-identity.test.ts`'s magic-byte sweep — the guard that
catches a file which IS a font by its own first four bytes but carries an extension
`manifest.go`'s `fontExtensions` does not recognise — ran over `folio-designer/public/fonts` only. It now
runs over the **git-tracked repository the licence gate itself walks** (repo root, minus `.git`, minus any
`lint` directory whose parent is `testdata`), per **D-8.5.2**.

**The measurement, taken at Story 8.4h's implementation.** So a later reader can tell an **inert** widening
from an **unmeasured** one:

| figure | value |
|---|---|
| files listed by `git ls-files` | 1438 |
| excluded by the gate's `*/testdata/lint` skip | 64 |
| population the widened guard asserts over | **1374** |
| population before the widening (`public/fonts`) | 18 |
| **newly covered** | **1356 files** |
| tracked files with a font-plausible extension | **11, all `.ttf`** — no `.woff`, `.woff2`, `.eot`, `.otc`, `.pfb`, `.pfa` or `.dfont` is tracked anywhere |
| tracked extensionless files | 19 (by `extensionOf`'s own rule, under which a dotfile counts as extensionless) |
| **newly REPORTED** | **none** |

**Measured AFTER this story's own fixtures landed**, not before. The first pass read 1435 / 1371 / 1353 / 18,
which was the tree without the three files this same commit adds
(`lint/testdata/licence/permissive/example.test/ufl-lib/{LICENSE,NOTICE.md,go.mod}`). A record whose whole
purpose is "measured, not assumed" has to be taken against the tree it ships in.

Top-level subtrees newly in reach: `.agents`, `.claude`, `.github`, `_bmad`, `_bmad-output`, `docs`,
`fixtures`, `folio-go`, `hashmatrix`, `lint`, `test-data`, `tools`, plus the rest of `folio-designer`.

**Two things that are NOT inert, recorded so the widening is not mistaken for cosmetic.** First, the
`git ls-files` intersection is **load-bearing**: a plain disk walk of the repo root would read the three
real variable TTFs in the gitignored `.font-sources/` (~20 MB) and everything under `node_modules/`,
`dist/` and `src/generated/`, none of which this repository redistributes. Second, **AC7 widened the
GUARD, not the licence gate** — `ResolveAssets`' own walk has been repo-wide since Story 3.6 and already
filters by extension, so **no non-font asset is newly subjected to the font allowlist by this story**. The
widening closes the composition D-8.5.2 names: a file invisible to the gate *because of its extension*, in
a directory invisible to the guard *because of its path*.

A latent defect was fixed on the way: `extensionOf` was `file.slice(file.lastIndexOf('.'))`, which for an
**extensionless** file returned the file's **last character**. Under `public/fonts` no such file existed;
repo-wide there are 19. It could never produce a false positive — the magic-byte check still gates every
report — but it would have printed nonsense in a failure message.

---

### DW-115 — `Ubuntu-font-1.0` is on the asset allowlist with no asset in the tree under it

- **Recorded by:** **Story 8.4h's build (2026-09-02)**, discharging the honesty obligation in that story's
  **Design Note 8** — *"the story must not claim a real-population witness it does not have."*
- **Owner:** **Story 8.5**, which is the story scoped to ship a catalogue face.
- **Severity:** LOW — the identifier is proved live; what it lacks is a real-artifact witness.
- **Status:** OPEN until Story 8.5 lands a face under it, or until it is established that none will.

**The claim.** Story 8.4h added the fourth member of the owner's four-id asset allowlist (D-8.5.3) —
`Ubuntu-font-1.0`, the Ubuntu Font Licence 1.0 — to `lint/internal/licence/classify.go` as both a
`permissiveSPDX` entry and a marker branch placed above the MIT case, and to `manifest.go`'s
`fontAssetLicenceAllowlist`. **Nothing in this repository is Ubuntu-licensed.** So the identifier is proved
by a classifier table, by an SPDX-line fixture module
(`lint/testdata/licence/permissive/example.test/ufl-lib/`) that the graph scan really reads, and by
synthetic scratch-repo fixtures in `manifest_test.go` — and by nothing else. There is no analogue of
`TestCommittedOFLTextClassifiesAsOFL11` to write, because there is no committed artifact to read.

That gap is **expected and deliberate**: D-8.5.13 scoped the member into 8.4h precisely so it exists before
8.5 needs it. It is recorded here so **the first real-population witness is attributable to Story 8.5**
rather than being assumed to have arrived with 8.4h.

**One decision inside it, flagged.** The owner's allowlist, as written in `epics.md`, `SPEC.md` and D-8.5.3,
names its fourth member **"UFL"** — the community abbreviation — while naming the other three as exact SPDX
identifiers. `UFL` and `UFL-1.0` are **not** SPDX identifiers (both 404 on spdx.org); `Ubuntu-font-1.0` is
(HTTP 200, *"Ubuntu Font Licence v1.0"*). 8.4h ruled, on the project's own convention, that "UFL" **denotes**
the licence and the code's obligation is to use its canonical id, and deliberately did **not** accept the
bare alias as a live map key — an exact-lookup entry no valid SPDX line can ever produce would be dead code
that looks live. **If the owner reads that as their call rather than the convention's, it is cheap to
reverse now and expensive after a face ships.**

**What discharges it.** Story 8.5 shipping a face under the Ubuntu Font Licence, with a committed-artifact
classification test on the real licence text — or a decision that no such face will ship, at which point
the allowlist member and its marker branch should be re-examined rather than left as detection with
nothing to detect.

---

### DW-116 — `gofmt -l` has a standing red that Story 8.4h's plan gate recorded as "no output"

- **Recorded by:** **Story 8.4h's close (2026-09-02)**, per **D-8.4h.5**.
- **Owner:** **unassigned — whichever story next touches `lint/internal/rules/`.** A one-line fix that
  must not be smuggled into an unrelated story's diff.
- **Severity:** LOW — cosmetic formatting; no behaviour, no gate.
- **Status:** OPEN.

`gofmt -l folio-go lint`, run from the repo root, prints exactly one line:
`lint/internal/rules/licencegraph_test.go`. Story 8.4h's `## Verification` recorded *"expected: no
output"*, which was a mis-measurement taken at its plan gate.

**Proven pre-existing, twice, at Story 8.4h's close:** the file's content extracted from `7aa283b`
is unformatted there too, and `git diff --name-only 80f46a0..HEAD` does not list it — Story 8.4h never
touched the file. It was NOT reformatted by 8.4h, deliberately: doing so would have put an unrelated
file in a commit whose whole claim is that it touches only the licence gate.

**What discharges it.** A story that legitimately edits `lint/internal/rules/licencegraph_test.go`
running `gofmt -w` on it as part of its own change. **Until then, `gofmt -l` printing a SECOND file is
a real failure and must not be waved through on this entry** — naming the red by identity is precisely
what stops that.

---

### DW-117 — the new Ubuntu-font marker branch is an unanchored substring conjunction, and correcting it corrects the OFL branch too

- **Recorded by:** **Story 8.4h's review (2026-09-02)**, deferral 1 of 7. Re-checked and confirmed at
  8.4h's close by direct probe.
- **Owner:** **ENGINEERING LEAD, to schedule as ITS OWN STORY** — an `8.4x`-series insertion on the
  D-8.5.1(a) / D-8.5.13 precedent. It is explicitly *not* assignable to Story 8.5.
- **Severity:** MEDIUM.
- **Status:** OPEN.

`strings.Contains(upper, "UBUNTU FONT LICENCE") && strings.Contains(upper, "VERSION 1.0")` in
`lint/internal/licence/classify.go` matches both conjuncts **anywhere in the text**, in either order,
with any distance between them. So an MIT licence text that merely *mentions* the Ubuntu Font Licence
and "version 1.0" in its body classifies as `(permissive, "Ubuntu-font-1.0")` rather than `MIT`.
**Measured at close, not inferred** — probed directly against the shipped classifier.

**Why it is its own story and not a patch here.** The pre-existing OFL-1.1 branch has the identical
unanchored shape (`Contains("SIL OPEN FONT LICENSE") && Contains("VERSION 1.1")`), and Story 8.4h's
own spec named that branch as *"the exact template the new branch must follow"*. Anchoring both
conjuncts to one title line changes the OFL branch, which classifies **ten of the eleven** committed
font directories — a change with real blast radius that 8.4h had no mandate for.

**Scope re-checked at close, and the deferral holds as scoped.** The misclassification is
permissive→permissive, and **both** ids are on the owner's four-id allowlist, so it is an
**attribution** defect, not a gate bypass: it cannot admit a forbidden licence. Confirmed by probe
that the copyleft branches sit **above** this one, so a GPL text that also mentions the Ubuntu Font
Licence and "version 1.0" still classifies `(copyleft, "GPL-3.0")` and is still refused. The
attribution still matters — the label is what lands in `lint/MANIFEST.md` and credits the asset.

**A crafted NOTICE cannot reach this branch.** `ResolveAssets` classifies only files whose basename
begins with `LICENSE`; `NOTICE*` content is read for the copyright line and never classified. The
exploit surface is a crafted or bundled `LICENSE*` file, not a `NOTICE`.

**What discharges it.** A story that anchors both branches' conjuncts to the licence's own title line,
with a red-proof per branch and a re-run of the committed-population witness.

---

### DW-118 — a second `LICENSE*` file in an asset directory silently decides the enforced verdict

- **Recorded by:** **Story 8.4h's review (2026-09-02)**, deferral 2 of 7.
- **Owner:** **Story 8.5**, the next story to add font asset directories.
- **Severity:** MEDIUM.
- **Status:** OPEN.

`ResolveAssets`'s `os.ReadDir` loop overwrites `licenceText` on every basename matching `LICENSE*`, so
the **last** entry in directory order wins. The loop pre-dates Story 8.4h; **what 8.4h changed is the
consequence.** Until 8.4h the arbitrarily chosen file only decided a manifest *label*; it now decides
whether the build **passes or fails**. A directory holding `LICENSE` (GPL-3.0) beside `LICENSE-MIT`
resolves to the MIT text and passes the fail-closed gate with a forbidden font.

**Not live today:** verified via `git ls-files` that every committed asset directory carries exactly
one `LICENSE*`, and no test covers the two-file case.

**What discharges it.** Refusing a directory that holds more than one `LICENSE*` file, with its own
located error and its own red-proof — the fail-closed reading, matching the story's own posture.

---

### DW-119 — the designer guard hand-copies the licence gate's walk exclusions instead of reading them

- **Recorded by:** **Story 8.4h's review (2026-09-02)**, deferral 3 of 7.
- **Owner:** **whichever story next touches `folio-designer/src/font-binary-identity.test.ts`.**
- **Severity:** LOW.
- **Status:** OPEN.

`insideTheLicenceGateWalk` restates the gate's two `SkipDir` rules — skip `.git`, skip any `lint`
directory whose parent is `testdata` — as TypeScript literals, with nothing tying them back to
`manifest.go`. This sits directly beside `licenceGateFontExtensions`, which parses `var fontExtensions`
**out of `manifest.go` source** precisely so the guard agrees with the gate rather than with a copy of
the gate, and whose header states that doctrine explicitly.

The silently harmful drift direction is narrow — the Go side narrowing its `*/testdata/lint` skip while
the TypeScript copy keeps skipping — but nothing compares the two.

**What discharges it.** Deriving the exclusions from `manifest.go` source the way the extension list is
derived, or an explicit assertion that the two agree.

---

### DW-120 — the owner's four-id font allowlist can be widened with NO test reddening (measured)

- **Recorded by:** **Story 8.4h's review (2026-09-02)**, deferral 4 of 7. **Measured by mutation at
  8.4h's close**, where it was previously an argument.
- **Owner:** **Story 8.5**, the story that first exercises the allowlist against a real face.
- **Severity:** MEDIUM. **Of the seven deferrals this is the one most worth promoting** — it is the
  only one that leaves an owner decision unguarded rather than an attribution imprecise.
- **Status:** OPEN.

Every test in `lint/internal/manifest` reads **into** `fontAssetLicenceAllowed[...]`; none asserts the
list's exact contents. **Measured at close, not argued:** appending `"GPL-3.0"` to
`fontAssetLicenceAllowlist` and running `cd lint && go test -count=1 ./...` leaves **all four packages
green** — the copyleft arm sits above the allowlist arm, so a copyleft id added to the owner's list is
not even caught by the copyleft refusal for the SPDX-line path that reaches it.

`fontAssetLicenceAllowlist` encodes a fixed **owner** decision (D-8.5.3), not a population that grows
by design, so **D-8.5.4's "no cardinality assertion" argument does not cover it** — the objection there
is to counting a set that legitimately grows. **D-8.5.13 forbids silent list-widening**, and this is
exactly the surface it forbids.

**What discharges it.** An exact `[]string` equality over `fontAssetLicenceAllowlist` naming D-8.5.3,
so that changing the owner's list requires changing a test that says whose decision it is.

---

### DW-121 — AC7's throws-proof depends on the temp directory not sitting inside a git repository

- **Recorded by:** **Story 8.4h's review (2026-09-02)**, deferral 5 of 7.
- **Owner:** **whichever story next touches `folio-designer/src/font-binary-identity.test.ts`.**
- **Severity:** LOW — machine-dependent vacuousness, not a wrong result.
- **Status:** OPEN.

The assertion is `expect(() => licenceGateTrackedFiles(notARepository)).toThrow(...)`. `git` searches
**upward** for a repository, so on a machine whose `TMPDIR` sits under a checkout the call succeeds and
the arm goes vacuous or asserts the wrong thing. It passes on this machine; it is not proved to pass on
every machine, which is what a throws-proof is for.

**What discharges it.** Passing `GIT_CEILING_DIRECTORIES`, or clearing `GIT_DIR`/`GIT_WORK_TREE`, so
the proof is machine-independent.

---

### DW-122 — the guard's tracked filter is per-FILE where the gate's is per-DIRECTORY, and a comment claims they are identical

- **Recorded by:** **Story 8.4h's review (2026-09-02)**, deferral 6 of 7.
- **Owner:** **whichever story next touches `folio-designer/src/font-binary-identity.test.ts`.**
- **Severity:** LOW — the stated equivalence is inaccurate; the divergence loses nothing today.
- **Status:** OPEN.

`manifest.go` skips a directory only when `gitTrackedFileCount` is **zero**, so an *untracked* font in
a directory holding any tracked file **is** assessed by the gate. The designer guard intersects
**file by file** with `git ls-files`, so it does not see that font. The guard's comment claims the two
exclude untracked files *"exactly"* the same way. All 18 files under `folio-designer/public/fonts` are
tracked, so the divergence is latent — but the comment is wrong, and a wrong comment about a
compliance boundary is what the next reader will trust.

**What discharges it.** Either matching the gate's per-directory rule, or correcting the comment to
state the difference and why it is acceptable.

---

### DW-123 — `looksLikeAFontBinary` ignores the `readSync` return count, newly reachable now that the guard is repo-wide

- **Recorded by:** **Story 8.4h's review (2026-09-02)**, deferral 7 of 7.
- **Owner:** **whichever story next touches `folio-designer/src/font-binary-identity.test.ts`.**
- **Severity:** LOW — latent false positive, not a false negative.
- **Status:** OPEN.

The helper compares a four-byte buffer without checking how many bytes `readSync` actually returned, so
a tracked file **shorter than four bytes** matches the zero-filled `00 01 00 00` sfnt magic and is
reported as a font. **Newly reachable because Story 8.4h took the population repo-wide** — under
`public/fonts` no such file could exist. Verified that no tracked file in the repository is currently
smaller than four bytes, so it is latent.

It fails **safe** (a false report, not a missed font), which is why it is deferred rather than fixed.

**What discharges it.** Checking the `readSync` return before comparing.

---

### DW-124 — CLOSER FINDING: the American spelling of the Ubuntu Font Licence is accepted SILENTLY as MIT, and the review rejection that said otherwise is FALSE

- **Recorded by:** **Story 8.4h's CLOSE (2026-09-02)** — a re-check of one of the review's seven
  rejections, measured by probe. **This is not a deferral the build recorded; it is a rejection the
  closer overturned.**
- **Owner:** **ENGINEERING LEAD**, to rule on scope. Naturally folds into **DW-117's** story, which
  already owns the same marker branch.
- **Severity:** MEDIUM.
- **Status:** OPEN.

Story 8.4h's review rejected the absence of a test for the American spelling `"UBUNTU FONT LICENSE"`
on the grounds that **"the miss is loud — `FamilyUnknown` fails the build — and D-2.1.3 makes a loud
miss fail-safe by design."** The classifier's own in-code comment makes the same claim.

**That is false after this change, measured by probe at close:**

| input | family | id | outcome at the font gate |
|---|---|---|---|
| UFL 1.0 text, British `LICENCE` + grant clause | permissive | `Ubuntu-font-1.0` | passes, correctly labelled |
| **UFL 1.0 text, American `LICENSE` + grant clause** | **permissive** | **`MIT`** | **passes, labelled `MIT`** |
| `"UBUNTU FONT LICENSE\nVersion 1.0\n"` (bare, no grant clause) | unknown | `""` | refused |

The rejection is true only of the **bare two-line synthetic** the story's own version-lookalike test
uses. Any **real** Ubuntu Font Licence text carries
`"Permission is hereby granted, free of charge, to any person obtaining a"` — Design Note 4 measured
this itself, from the SPDX-published text — so an American-spelled UFL file misses the new marker
branch, falls through to the MIT case immediately below it, and classifies `(permissive, "MIT")`.
`MIT` is on the owner's four-id allowlist, so the build **passes** with the asset **mis-attributed**.

**The miss is therefore SILENT, not loud** — which is character-for-character the defect Design Note 4
says the branch's placement above MIT exists to prevent. The placement fixes it for the British
spelling only. D-2.1.3's fail-safe argument does not apply to a miss that lands in a *neighbouring
permissive branch* rather than in `FamilyUnknown`.

**Not patched at close, deliberately:** the correct fix is a decision (accept both spellings, or anchor
the branch so a near-miss reaches `FamilyUnknown`), it touches the same branch as DW-117, and it
reopens a question the review already ruled on. That is a lead's call, not a closer's.

**What discharges it.** A ruling on whether the marker accepts both spellings, plus a test asserting
the American-spelled full text does **not** classify as `MIT` — the same shape as the existing
MIT-collision proof, which today covers only the British spelling.

---

### DW-125 — CLOSER FINDING: the first `SPDX-License-Identifier:` line anywhere in a LICENSE file wins outright, and that is now a GATE BYPASS

- **Recorded by:** **Story 8.4h's CLOSE (2026-09-02)**, measured by probe. **Named by no review layer.**
- **Owner:** **ENGINEERING LEAD**, to rule on scope — this is a compliance-boundary decision, not a
  cleanup.
- **Severity:** MEDIUM–HIGH on the compliance boundary; the mechanism is pre-existing, the
  **consequence is new with Story 8.4h**.
- **Status:** OPEN.

`ClassifyLicenceText` runs `spdxLineRE.FindStringSubmatch(text)` **first** and returns immediately on
the **first** match anywhere in the text, before any marker branch is consulted. So a licence file
carrying the **full GNU GPL v3 legal code** plus, anywhere below it, a line reading
`SPDX-License-Identifier: MIT` classifies as **`(permissive, "MIT")`**. Measured by probe at close.

Before Story 8.4h this produced a wrong *label* on a passing build. **After Story 8.4h it produces a
passing build for a forbidden licence** — the copyleft branches never run, because the SPDX path
returned before reaching them. This is the same "pre-existing shape, newly load-bearing" class as
DW-118, and it is the one shape that defeats the story's central claim that a copyleft text is
*"refused by name."*

**Realism.** Bundled/aggregated `LICENSE` files that list several licences with SPDX lines are common
in redistributed font packages, so this needs no adversary — only a real-world multi-licence file whose
first SPDX line is not the governing one.

**Not live today:** every committed asset `LICENSE*` was re-verified at close to resolve to an
allowlisted id, and `lint/MANIFEST.md` regenerates byte-identically.

**Not patched at close, deliberately:** changing the classifier's precedence order affects **every**
consumer of `ClassifyLicenceText`, including the dependency graph scan, and is well outside a closing
agent's scope fence.

**What discharges it.** A ruling on precedence — most likely that a copyleft **marker** in the body
outranks a permissive SPDX line, or that a text carrying more than one distinct SPDX identifier is
refused as unclassifiable — with a red-proof per arm.

---
