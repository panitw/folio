---
baseline_commit: 2516b13
---

# Story 2.6: Render multi-page documents with headers and footers on every page

**Epic:** 2 — Text, shaping, breaking and page composition
**Story key:** `2-6-render-multi-page-documents-with-headers-and-footers-on-ever`
**Status:** `done`
**Charter:** `epics.md` §Story 2.6 (line 880). **Covers:** FR30 · NFR4 · AD-4.
**Primary invariant:** **AD-4** — *"Two passes, and the second one lays nothing out."* D-000.9's Epic 2
note names this story by name: *"AD-4 also gets its first real test at 2.5/2.6 — pagination is where a
second layout pass is most tempting."*
**Adjacent invariants:** **AD-24** (boxes are absolute, and nothing negotiates; bands are placed by
`internal/layout` alone) · AD-5 (the page model knows nothing about PDF) · AD-21 / AD-22 (every
feature ships its golden; a hash change is an intended, versioned event) · AD-13 (derived geometry,
one function) · AD-2 / AD-3 / AD-23 (one fixed-point unit, one number emitter, no float)
**Governing rulings:** **D-000.41** (do not request a scarce human sign-off while scheduled work may
move the artifact — *this story carries the obligation, and discharges it in AC1*) · **D-000.43** ·
**D-000.44** (a re-recording is a recording) · **D-000.45** (assert the computed value from a
declarative table, never a direction) · **D-000.33** (a conservation assertion cannot see a degenerate
partition) · **D-000.42** (redundant is a third category — label, never count) · **D-000.35** (name the
symbol) · **D-000.48** (a sweep that writes citations verifies each resolves) · **D-000.28** (never
write a claim before the event it asserts) · **D-000.49** (a record that overstates a risk is a defect)
· **D-2.5.1** (the gate-obligation guard reads a declared list) · **D-000.4 (override criterion)** ·
**D-000.22** · **D-000.21 (sharpened)** · **D-000.26 (refined)** · D-000.9 · D-000.23 · D-000.24 ·
D-000.30 · D-000.34 · D-000.36 · D-000.37 · D-000.39 (sharpened) · D-000.46 · D-000.47 · **D-2.4.2
(amended)** · **D-2.5.1** · **D-2.5a.1** · D-1.6.5 · D-1.8.10 · D-2.3.5 · D-2.4.3
**Deferred work touched:** **DW-11 owes a written answer here** — its owner window is *"Epic 2's later
stories"* and this is one (AC10). **DW-14 is re-measured, not discharged** (AC11). **DW-16 is not
touched.**

---

## FIRST — the sign-off measurement D-000.41 requires, performed at creation

**This gates a request to a human, so it is stated before anything else.**

### `fixtures/shaped-text/expected.pdf` — **Story 2.6 does NOT move it.**

Current digest `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370`, re-computed at
`2516b13` (not inherited). D-000.41's own branch table applies:

> *"2.6 does **not** move `shaped-text` → request the reading sign-off after 2.5a, as planned."*

**So D-2.3.5's Thai reading sign-off may be requested as soon as D-000.43's second condition is met**
(the gate's four matrix legs have run and agreed on that digest). This story does not send the
request — no story does — and it does not create `thai-signoff.json`.

**The measurement, not the reasoning.** Every fixture's content band was laid out with the production
functions (`pageGeometryOf`, `layout.ContentHeight`, `shapeSegments`, `packLines`,
`chainVerticalModel`) and its lowest content bottom compared against the derived content height:

| fixture | content height (millipoints) | lowest content bottom | headroom | occupancy | would paginate? |
|---|---|---|---|---|---|
| `shaped-text` | 729,890 | 193,760 (`e7`) | 536,130 | 26.5% | **no** |
| `wrapped-text` | 729,890 | 275,420 (`e4`) | 454,470 | 37.7% | **no** |
| `three-band-page` | 727,890 | 136,344 (`e2`) | 591,546 | 18.7% | **no** |
| `font-text` | 729,890 | 16,408 (`e1`) | 713,482 | 2.2% | **no** |
| `multi-script-fallback` | 729,890 | 22,540 (`e1`) | 707,350 | 3.1% | **no** |
| `image-embed` | 729,890 | 60,000 (`e1`) | 669,890 | 8.2% | **no** |

`minimal-rect` is a fixed four-object document with no template and no bands at all; it cannot
paginate by construction.

**The tightest fixture uses 37.7% of one content band.** No committed fixture is within 454pt of the
page boundary, so **no pagination rule that produces one page for content that fits can move any of
them** — and the one-page serialization path is not a new path: `pdf.SerializeTextDocument` already
takes `[]pagemodel.Page`, already reserves one page/content object pair per page, and already emits
`/Kids` and `/Count`. Today `render.go:723` hands it a one-element slice. **For a document that fits,
the byte sequence after this story must be the byte sequence before it.**

**That is a constraint on the story, not a property of the universe** — it holds *if and only if* 2.6
leaves the object-reservation order and the band arithmetic alone. **AC1 makes it machine-checkable.**

### `fixtures/expected-breaks/expected_breaks.json` — **Story 2.6 does NOT move it. The in-flight review sheet is safe.**

Current digest `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de`.

**This is a structural answer, not a probabilistic one.** Its own README states, and the repository
confirms: *"There is deliberately no generator for this directory. **Nothing in the repository can
write `expected_breaks.json`.**"* It is a hand-authored oracle over break **opportunities**, consumed
by `internal/text/s4_expected_test.go` against `text.Opportunities`. The only change that could
invalidate its labels is a change to `internal/text`'s break engine.

**Pagination cannot reach `internal/text`.** `packLines` takes no vertical quantity (`segs`, `ops`,
`totalRunes`, `fontSize`, `boxWidth`) and runs **before** the vertical model, so wrapping is
structurally incapable of shifting from a pagination change; and pagination consumes `packLines`'
output rather than feeding it. **D-000.26 (refined) already ruled that the break judgment binds to the
break-opportunity vector**, which this story does not compute, produce or consume.

**Nothing is in flight that this story endangers.** The review sheet already sent to the owner may
stand. AC1 pins it mechanically anyway.

---

## Baseline, measured at creation

HEAD is **`2516b13`** — *"Story 2.5a: Align the first baseline with the leading model (finisher)"* — on
`main`. Every number below was **re-measured at this commit**. None is inherited from the brief.

| scope | invocation (verbatim) | counting rule | result |
|---|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | every `--- PASS`/`--- FAIL` occurrence, subtests included | **504 PASS · 1 FAIL** |
| `folio-go/` | the same | top-level only | **312 PASS · 1 FAIL** |
| `lint/` | the same | all occurrences | **85 PASS · 0 FAIL** |
| `lint/` | the same | top-level only | **47 PASS · 0 FAIL** |
| `folio-go/` | `GOWORK=off go list -m all` | — | exactly **two**: `github.com/panitw/folio/folio-go`, `github.com/boxesandglue/textshape v0.0.15` |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | — | clean |

**The single failure is `internal/text`'s `TestCorpusMeetsP6ExerciseFloors`** — *"P6g (opaque names)
floor not met: got 7, need >=20"*, `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.
**Story 2.4's AC5 requires it to stay red and byte-identical.** Do not fix it, do not skip it, do not
add corpus items to fill it. **A second failure is a regression**, with the one named exception in AC6
(the new fixture's golden test is expected red *during* the story and green at its close).

> **Two measurement-instrument hazards, both hit while taking the numbers above** (D-000.40's family).
> First: `rtk` intercepts `go test` and rewrites its output — a filtered run reports
> *"text (47 passed, 1 failed)"* and **zero** `--- PASS` lines, so a naive `grep -c` returns 0 and
> reads as total collapse. Second: `rtk`'s own tee log is **`go test -json`**, so `grep -cE '^--- PASS'`
> against it also returns **0** — the top-level counts above were taken by matching
> `'"Output":"--- PASS'` (no leading spaces) versus `'"Output":"    --- PASS'`. **Take your counts the
> same way, or your gate table will be fiction.**

### The working tree is NOT clean at this baseline, and that is reported rather than corrected

`git status --porcelain` reports **one** modified file:

```
 M _bmad-output/implementation-artifacts/folio-mvp-decision-log.md
```

It is a **25-line append carrying D-000.49** (*"A record that overstates a risk is a defect, not
caution"*), raised by Story 2.5a's finisher and **never committed**. This story neither reverts it nor
commits it: it is a live ruling that arrived after the last commit. **It is treated as binding here**
and listed in *Governing rulings*. **No `folio-go/` or `lint/` source file is dirty**, so every code
measurement in this document is against `2516b13`'s code exactly. See *Flagged, not fixed* — item 2.

---

## In plain terms (read this first if you just want the gist)

Before this story, Folio drew exactly one sheet of paper no matter how much there was to draw. If a
report's content was taller than the space between the running header and the running footer, the
surplus was still drawn — simply past the bottom edge of the sheet, where no printer and no viewer
would ever show it. Nothing warned, nothing errored, and the file looked perfectly well-formed. A
statement that should have been three pages arrived as one page and two pages' worth of invisible ink.

That is fixed. A long document now becomes several sheets, and the running header and the running
footer appear on each of them, unchanged, so page thirty-four is as complete as page one. Nothing on a
page negotiates for space or shuffles to make room; growing means more sheets, never rearranged
furniture.

But the first multi-page document this feature ever produced did not actually open. It passed its own
fingerprint check and every test written to prove it correct, and it still would not render — a viewer
shown the file reported no pages at all. The team's own tests could not have found this, because every
one of them had been written when a document could only ever have one page, and each quietly assumed
that shape was permanent. The person who found it did so by trying to open the file, not by running a
test. That is now treated as the more serious defect: not the one broken byte, but the way an entire
family of checks could be satisfied by a file that plainly failed to work. The remedy the team put in
place was accordingly broader than a one-line fix: the broken construction is now physically impossible
to repeat, an outside program that has no stake in the outcome opens and reads every reference document
before it can be accepted, and the checks that had been quietly assuming "there is only one page" were
found and widened.

A second pass after that, checking the story's own safety net rather than the feature, found that the
very guard meant to protect this document from a repeat of the same defect had inherited the identical
blind spot — it read only the first page's worth of text and had never noticed the second existed. That
is fixed too, along with several smaller gaps the same pass surfaced in how the feature's tests count
and describe themselves. None of it changed what the feature does; all of it changed how confidently
that can be shown.

Two things this story deliberately still does not do. It does not fill in page numbers: the little
placeholders that will eventually read "page three of eleven" stay untouched and literal, because
resolving them belongs to the next story. And it does not trim anything that spills outside its own
box — that is the story after next.

One test keeps failing throughout and is expected to: an older, deliberately unmet coverage floor from
an earlier spike, left exactly as it was.

---

## Story

As an integrating Go developer,
I want a document longer than one page to carry its header and footer throughout,
So that page 34 of a statement is as complete as page 1.

---

## Do not re-open — settled rulings this story inherits

Each was ruled in the last two days and each was **verified against the code at `2516b13`**, not
taken from the brief.

1. **The vertical model is one rule with three maxima** (D-2.4.2 amended), each maximised
   **independently** over the declared chain: top→first baseline `max(A)`; baseline→baseline
   `max(A) + max(D) + max(gap)`; last baseline→bottom `max(D)`. Verified in `wrap.go`'s
   `verticalModel` and stated in `folio-format.md:271-273`. **Do not re-open. Do not re-derive.**
   A line's inter-baseline advance is `verticalMetrics.Advance` and nothing else.
2. **`internal/pagemodel` and `internal/layout` exist and are the seam.** `layout.ContentHeight`,
   `layout.Origins`, `layout.PlaceInBand`, `layout.ComposePage` are all present in
   `internal/layout/band.go`. `stage-rank` enforces `layout`(7) ↛ `pdf`(8).
3. **`ContentHeight` is geometry-only and cannot consult its own elements** (AD-24). Its only
   parameter is `layout.PageGeometry`, a closed struct of page setup. **Any negotiation or reflow is
   forbidden and this story does not add a field to that struct.**
4. **`splitByFace` is deleted** (D-000.46). Placement is `shapeSegments` + `positionSegments`.
   Verified: `9fa6f17` removed it and re-pointed its ten references. Do not resurrect it.
5. **`packLines` runs before the vertical model and takes no vertical quantity.** Verified:
   `wrap.go:179`, signature `packLines(segs, ops, totalRunes, fontSize, boxWidth)`. **Wrapping is
   structurally incapable of shifting from a leading or pagination change.**
6. **`{{page}}` / `{{pages}}` are RESERVED and owned by Story 2.7** (D-1.6.5 AC18). Verified live in
   `internal/bind/text.go`'s `reservedPlaceholders`: they pass through unchanged and produce no
   `Substitution`. **This story must not implement them** — see AC8 and *Corrections* item 3.
7. **The `pdf` serializer is already multi-page.** `pdf.SerializeTextDocument([]pagemodel.Page, …)`
   reserves a page object and a content object per page, writes every page id into `/Kids`, and
   writes `len(pages)` into `/Count`. **This story does not rewrite it; it stops handing it a
   one-element slice.**

---

## Corrections to inherited claims — verified at `2516b13`

**Four of Story 2.6's own epic ACs assert something about existing code, and three of them are
already true.** That is not a licence to skip them — it is a warning that written naively they will
ship as vacuous passes (D-000.42). Each is corrected below into a form with independent teeth.

### 1. `epics.md`'s *"pass one produces a complete page model and pass two serializes it"* is **already true**

`internal/pdf` performs no measurement, no line breaking and no pagination today. An AC asserting it
would pass on day zero. **What is NOT true is that anything prevents it from acquiring one.**
`stage-rank` forbids `layout`(7) → `pdf`(8); it says nothing about `pdf` → `layout`, which is a
*downward* import and therefore **currently legal**. A future `internal/pdf` that imported
`internal/layout` and re-split a page would violate AD-4 and break **no existing guard**.

**So AC5 is a new forward guard on a real, currently-open hole** — not a restatement. It is
constructible and it has a red-proof (mutate `internal/pdf` to split one page into two → it reddens).

### 2. `epics.md`'s *"never by displacing sibling components within a page"* is **already structural**

`layout.ContentHeight` receives only `PageGeometry`; `layout.Origins` derives from it alone;
`internal/bandcomposition_arch_test.go`'s `TestContentHeightDependsOnGeometryAlone` and
`TestNoBandOriginArithmeticInPackageFolio` already hold it. **Restating it as a fresh AC would be a
redundant guard counted as coverage** — exactly D-000.42's third category. AC2 therefore asserts the
**new** half only: that growth produces *more pages*, with the existing non-displacement guards named
as the mechanism rather than re-implemented.

### 3. `epics.md` §2.6 is **silent on page numbers, and that silence must be kept**

Nothing in 2.6's epic text implies `{{page}}`/`{{pages}}` resolution — checked line by line at
`epics.md:880-903`. **Story 2.7 at `epics.md:904` owns them, and the format spec at
`folio-format.md:446-448` says so normatively.** The hazard is not the epic text; it is that a
developer building pagination will have the page index in hand and it will be one line to substitute
it. **AC8 exists solely to make that one line fail a test.**

### 4. `epics.md`'s *"content exceeding one page"* has **no fixture that can express it**

Measured across all seven committed fixtures at `2516b13`:

| property | fixtures having it |
|---|---|
| more than one page of content | **0 of 7** |
| a **populated** `pageHeader` band | **1 of 7** (`three-band-page` only) |
| a **populated** `pageFooter` band | **2 of 7** (`three-band-page`, `font-text`) |
| pairwise-distinct margins and band heights | **1 of 7** (`three-band-page`) |

This is Story 2.5's finding one story on: *zero* tests detected a moved page-header band because
every fixture but one had an empty header and a symmetric page setup. **A correct assertion over
inputs that cannot express the defect is vacuous.** AC6's new fixture is specified against this
table, and AC6 states what its content can and cannot express.

---

## Scope fence — what this story is NOT

- **Not `Page X of Y`.** No substitution of a page index into any string. `reservedPlaceholders` is
  byte-unchanged. (Story 2.7.)
- **Not clipping, and not an overflow diagnostic.** Content that exceeds its own declared box is
  Story 2.8's (FR44, AD-14). This story changes *how many sheets exist*, not *what happens at a box
  edge*.
- **Not tables, and not table continuation headers.** Epic 4 (FR22–FR28) owns those; a table element
  is not laid out by this story at all.
- **Not streaming.** NFR4 is explicit that FR39's writer API is an ergonomic shape, **not** a
  constant-memory guarantee, precisely because `Page X of Y` forces two passes. `RenderTo` keeps its
  signature and keeps buffering.
- **Not a re-recording of any existing golden.** See AC1.
- **Not a change to `ContentHeight`, `Origins`, `PlaceInBand` or `PageGeometry`'s field set.**
- **Not the epic key.** `sprint-status.yaml`'s `epic-2: backlog` is the gate's to flip — see
  *Flagged, not fixed*.

---

## Measured findings — read all of these before writing code

### 1. THE RED-PROOF, captured at `2516b13` before any fix exists (D-000.30)

A document with a populated header (`height: 18`), a populated footer (`height: 24`), asymmetric
margins (top 30, bottom 42) and one content text element wrapped to a 150pt box over ~60 repetitions
of a Latin sentence, rendered through the public `folio.Render`:

| observable | value at `2516b13` |
|---|---|
| PDF bytes | 82,696 |
| `/Type /Page ` objects emitted | **1** |
| `/Count` in the page tree | **1** |
| text-placement (`Tm`) sites in the content stream | **146** |
| of those, sites with a **negative** PDF user-space Y | **91** |
| the most negative Y emitted | **−1000.727 pt** |

The paper is 841.89 pt tall. **91 of 146 lines — 62% of the document's text — are drawn between 0 and
roughly 1.19 page-heights *below the bottom edge of the sheet*.** The render succeeds. Nothing warns.

**This figure is the red-proof, not the shipped assertion.** D-000.45 forbids shipping a guard phrased
as a direction ("lines end up below the page"); AC4 ships a **declarative per-page line-index table**
instead. Reproduce this measurement at baseline before changing anything, and record it in the
Delivery Log. If it does not reproduce, stop and report — the story's premise has moved.

### 2. The one-element slice is the entire seam

`render.go:723`:

```go
return pdf.SerializeTextDocument([]pagemodel.Page{page}, embedded, pdfImages)
```

Everything upstream — `collectTextRuns`, `collectImageRuns`, `buildShapedPDFRuns`,
`layout.ComposePage` — produces one page's worth of page-absolute content. Everything downstream
already handles N. **The story is: make the middle produce N.**

### 3. Object numbering is order-sensitive and is what protects the six existing goldens

`SerializeTextDocument` reserves ids in this order: catalog, pages, **images (sorted by asset key)**,
**faces (sorted by name, with a `/CIDToGIDMap` id only for a face that needs one)**, then **one page
id and one content id per page, in page order**. A one-page document reserves exactly the ids it
reserves today. **Reserving anything new, or reserving per-page objects earlier, moves every existing
golden.** AC1 catches it; this note is so it is not discovered by a red hash.

### 4. `internal/pdf` → `internal/layout` is a legal import today

Confirmed by reading `lint`'s `stage-rank` rule direction and the two packages' declared ranks.
This is the open hole AC5 closes. **Do not close it by adding a rank; `pdf` legitimately sits above
`layout`.** Close it by asserting the *property* (pdf emits exactly `len(pages)` page objects and
references no layout or breaking symbol), which is the AD-4 clause itself.

### 5. DW-14 is re-measured, not discharged

DW-14 (`/ToUnicode` emits one unbounded `beginbfchar` section; the spec caps a section at 100) names
its owner as *"the Epic 2 boundary gate, or Story 2.4 if its corpora reach the limit first."* **A
multi-page fixture is the first document in the repository that can plausibly reach 100 distinct
`(glyph, text)` pairs in one face.** AC11 requires the number to be *measured and reported* for the
new fixture, in both directions (D-000.49), and DW-14 updated with the figure. It does **not** require
the fix — that stays the gate's.

---

## DECISIONS NEEDED — escalate before development starts

**Five. The first three are one design question at three granularities and should be ruled together.**

### DN-1 — What does "paginate" mean for absolutely-positioned content under AD-24?

`epics.md` §2.6 says content *"grows by producing more pages, never by displacing sibling components
within a page"* and says nothing further. The content band holds elements at **declared,
band-relative Y**. Under AD-24 nothing negotiates and no element moves relative to another. Three
candidate rules:

- **(A) The page is a window onto one unbounded content column.** The content band is a viewport of
  height `ContentHeight` onto a column of unbounded extent. Page *N* shows column rows
  `[N·H, (N+1)·H)`; an item's page-relative Y on page *N* is `columnY − N·H`. Page count is
  `ceil(lowestBottom / H)`, minimum 1. Nothing moves relative to anything; `ContentHeight` still sees
  only `PageGeometry`; the arithmetic is integer `geom.Length` division and comparison.
- **(B) Whole elements move to the next page when they do not fit.** **Ruled out by the epic's own
  third AC** — moving an element because its neighbour is tall *is* displacement. Recorded so nobody
  re-proposes it.
- **(C) Absolutely-positioned content does not paginate at all; only tables do (Epic 4).** Makes 2.6
  vacuous, contradicts FR30 and contradicts the epic's first AC. Recorded and rejected for the same
  reason.

**Recommendation: (A).** It is the only reading under which "growth produces more pages" and "nothing
negotiates" are both true, and it needs no new input to any existing function.

### DN-2 — At what granularity is content assigned to a page: element, or line?

Under (A), if the unit is the **whole element**, a text element taller than the content band still
runs off the bottom of its page and the story fixes nothing. If the unit is the **line**, a
paragraph's lines continue onto the next page — which is what a reader expects and is **not**
negotiation (no line moves horizontally, and no line moves relative to a sibling).

**Recommendation: the line**, with the boundary function being the line's **top edge** — i.e. the
`columnY` the page model already carries as `pagemodel.TextRun.Y`. Using the top rather than the
baseline avoids a second derivation of a vertical quantity, which D-2.4.2's whole amendment exists to
prevent.

**Note for whoever rules this: additivity cannot distinguish these.** "Every line appears on exactly
one page" holds for *any* monotone boundary function, including a degenerate one that puts everything
on page 1 (D-000.33). The ruling must be pinned by a **declared table of line index → page index**,
which is what AC4 does.

### DN-3 — What happens to an image that straddles a page boundary?

An image is an atomic drawn box, not a sequence of lines. Either (i) the whole placement goes to the
page containing its **top**, accepting that it may overhang that page's content band, or (ii) it is
clipped — but clipping is Story 2.8's, and 2.8 is scoped to *"content exceeding its declared bounds"*,
which a page boundary is not.

**Recommendation: (i)**, the same monotone top-edge boundary function as DN-2, so there is one rule
and not two.

### DN-4 — Does this story record a new golden, and does that make a **FIFTH** Epic 2 gate obligation?

**It does, and this is stated explicitly because the gate owes exactly four things today and
D-2.5.1 requires a ruling before a fifth is added.** The four, verified at `2516b13` against
`declaredEpic2GateObligations` and `matrixDocuments`:

| the gate's four obligations | this story's effect |
|---|---|
| the four-target matrix legs | **count unchanged by re-recording — but see below** |
| D-2.3.5's Thai **reading** sign-off | **untouched, and now RELEASED to be requested** (AC1) |
| D-2.4.3's Thai **break** sign-off | **untouched; its artifact is byte-unchanged** (AC1) |
| `three-band-page`'s deferred matrix legs | **still deferred, still declined under D-000.4** |

**The fifth, proposed:** `matrix-document: multi-page` — a new fixture directory, registered in
`matrixDocuments`, `.github/workflows/matrix.yml`, `goldenDigestRecord` and
`declaredEpic2GateObligations`, whose four legs are **deferred to the gate** on the same D-000.4
reasoning that deferred `three-band-page`'s.

**AD-21 binds every feature to ship its golden**, and a golden registered in one list and not the
other is *"a matrix leg nobody compares, reported as green."* **But it is still a fifth obligation on
a gate that owes four, and it must be ruled rather than assumed.** If it is refused, the alternative
is that FR30 ships with no cross-target artifact at all — state which.

### DN-5 — `folio-format.md` has no pagination clause, and the rule ruled in DN-1 is normative format behaviour

`folio-format.md` describes `bands`, the derived content height and the vertical model, and says
**nothing** about what happens when content exceeds one page. Under D-000.6's canonical-document
discipline the DN-1 rule belongs there, beside the content-height derivation it depends on.

**Recommendation:** amend `folio-format.md` §`bands` with the ruled pagination rule and *why*, in the
same shape as the vertical model's §`leading` (rule, worked example, and the wrong-looking alternative
named so it is not re-proposed). **Do not write the amendment before DN-1 is ruled** (D-000.28).

---

## Acceptance Criteria

> **Convention.** Every assertion below is on the **produced artifact** and states **what carries the
> property** (D-000.21 sharpened). No AC asserts a direction; each asserts a **computed value from a
> declarative table** (D-000.45). Any guard without independent teeth is **labelled redundant and not
> counted** (D-000.42). Every symbol named here was verified to exist at `2516b13` (D-000.35, D-000.48).

**AC1 — No signed-off-pending golden moves, and that is machine-checked, not argued.**
At close, all seven committed `expected.pdf` digests are byte-unchanged, asserted by the **existing**
`TestGoldenDigestAgreesAtEveryDeclaredSite` (`byte_neutrality_test.go:194`), which re-hashes each
artifact and cross-checks every declared recording site — **this guard is proven, not redundant: it
fires on a moved artifact.** Additionally `fixtures/expected-breaks/expected_breaks.json` is
byte-unchanged, and `git diff --stat -- fixtures/` at close lists **only** the new fixture directory —
a both-directions assertion (D-000.49): the claimed absence is shown absent *and* the claimed addition
shown present. The Delivery Log restates `shaped-text`'s digest `6c040ef7…c6c85370` and
`expected_breaks.json`'s `a545e042…9324de` as re-computed at close.

**AC2 — Content exceeding the content band produces more pages, and the page count is a computed
value from a declarative table.**
For each row of a declared table of `(page geometry, content extent) → expected page count`, the
rendered document's `/Count` and its number of `/Type /Page` objects both equal the declared integer.
The table includes at minimum: content that fits exactly (1), content one millipoint over (2), content
spanning three bands' worth (3), and a zero-content document (1). **Not** "more content yields more
pages." **Mechanism, not re-implementation:** non-displacement is already held by
`internal/bandcomposition_arch_test.go`'s `TestContentHeightDependsOnGeometryAlone` and
`TestNoBandOriginArithmeticInPackageFolio`; this AC names them and does not restate them.

**AC3 — The page header and page footer bands appear on every page, at identical band origins.**
On an N-page render of the AC6 fixture (N ≥ 2), for every page index `0..N-1` the content stream
carries the header element's glyph run at the **same** PDF user-space Y and the footer element's at
the **same** PDF user-space Y, both equal to declared literal values hand-derived from the page setup.
The header and footer strings are **distinct, identifiable literals** so a band mix-up shows in the
rendered text and not only in a coordinate. **Presence precondition:** the test first asserts both
bands are non-empty in the input, so the assertion cannot pass by finding nothing on either page.

**AC4 — The page partition is pinned by value, and both parts are non-empty.**
A declared table maps **line index → page index** for the AC6 fixture's straddling element,
hand-derived from the content height and the ruled advance, and the test asserts the observed
assignment equals it exactly. It **additionally** asserts every page in the partition is non-empty.
D-000.33, stated in the test's own comment: *a conservation law ("every line appears exactly once") is
satisfied by a degenerate partition, and additivity is preserved by any monotone boundary function* —
so the boundary index is asserted, never the sum.

**AC5 — Pass two lays nothing out (AD-4), enforced by a guard with independent teeth.**
A new structural guard asserts, positively: `internal/pdf` emits exactly `len(pages)` `/Type /Page`
objects for any input, and no file under `internal/pdf` imports `internal/layout`, `internal/text`, or
references any line-breaking or pagination symbol. **This is not redundant with any existing guard** —
`internal/bandcomposition_arch_test.go`'s `TestInternalLayoutDoesNotReachInternalPDF` asserts the
*opposite* arrow, and `stage-rank` permits `pdf` → `layout` because `pdf` legitimately outranks it
(*Measured findings* 4). **Red-proof required:** a mutation that makes `internal/pdf` split one input
page into two must redden this guard and the mutation must print its own preconditions
(D-000.40 sharpened). Record whether it fired.

**AC6 — A new multi-page golden fixture, with its semantic acceptance step (D-000.22, D-000.44).**
A new `fixtures/multi-page/` with `input.folio`, `expected.pdf`, `expected.json` and `README.md`,
plus its byte-identical in-repo template constant following `three_band_page_template.go`'s precedent.
Its content is specified against the *Corrections* item 4 table and it **must** carry: a populated
`pageHeader` **and** a populated `pageFooter`, each with a distinct literal string; **pairwise-distinct**
margins and band heights; at least one element whose lines straddle a page boundary; and enough content
for **N ≥ 2** and no more than **N = 3** (a golden a human must be able to read).
The fixture's `README.md` states, in its own words, **what defect its content can express and what it
cannot** — following `three_band_page_template.go`'s model. The **machine-checkable half of semantic
acceptance is non-deferrable** (D-2.3.5): page count, per-page header/footer presence and Y, the pinned
line→page partition, embedded face identity and byte-identity across two processes are all asserted **at
recording**. **No human sign-off obligation is created by this fixture** — it is all-Latin by
construction, creating no reading judgment (`three-band-page`'s precedent).

**AC7 — The new golden is registered everywhere a golden is registered, from one declared list.**
`matrixDocuments` (with a `slug`, a `capture` and an `extraGuard` that proves the document is genuinely
multi-page), `.github/workflows/matrix.yml`'s per-target artifact list and its `docs=` line,
`goldenDigestRecord` (with its `expected.json` site and its independent second literal, per D-000.47),
and `declaredEpic2GateObligations` as a **one-line addition naming this story and DN-4's ruling**
(D-2.5.1). `TestEpic2GateObligationsMatchTheDeclaredSet`,
`TestGoldenDigestAgreesAtEveryDeclaredSite` and `matrix_registration_test.go` must all pass.

**AC8 — `{{page}}` and `{{pages}}` are not implemented, and a test makes implementing them fail.**
`internal/bind/text.go`'s `reservedPlaceholders` is **byte-unchanged**. A multi-page render of a
document whose footer value contains `{{page}}` emits the **literal characters** `{{page}}` as glyphs
on **every** page — asserted off the produced content stream via the `/ToUnicode` extraction, not off
the input. No `page` namespace is added anywhere. **Presence precondition:** the test asserts the
document renders as N ≥ 2 pages first, so it cannot pass on a one-page document.

**AC9 — The Epic 2 gate's obligations, each named with its fate, and the fifth stated explicitly.**
The Delivery Log carries DN-4's table with the actual outcome for each of the four, plus an explicit
statement of whether the fifth was added and under which ruling. **No obligation is added silently**,
and `declaredEpic2GateObligations`' diff is exactly one line if DN-4 is granted, zero if refused.
`epic-2-boundary-gate.md` gains a new dated section — appended, never overwriting.

**AC10 — `TestCorpusMeetsP6ExerciseFloors` stays red and byte-identical, and DW-11 gets its written
answer.**
`fixtures/thai-break-corpus/corpus.json` and `cmd/gencorpus/main.go` are byte-unchanged, and the
failure still reports `P6g … got 7, need >=20` with `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284
P6f:115 P6g:7}` — the same stats line, character for character. **Separately: DW-11's owner window is
"Epic 2's later stories" and this is one**, so `deferred-work.md` gains this story's answer in writing,
following Story 2.4's AC18 precedent — *"none were found, and none were invented"* is an acceptable
answer; **silence is not**, and D-000.17 forbids manufacturing attestation to reach a number.

**AC11 — The standing constraints hold, asserted in both directions (D-000.49).**
`GOWORK=off go list -m all` under `folio-go/` is **exactly two** modules. No `.go` file under
`folio-go/` contains the literal `SOURCE_DATE_EPOCH`, **including inside error strings**
(`absence-source-date-epoch`). No compressor is imported anywhere — the claimed count is **0** and is
shown to be 0, not assumed. `fmt` is absent from `internal/pdf`. `no-float-typed-value` and the
syntactic float guard both report zero under `internal/` and the module root. **And DW-14 is
re-measured:** the new fixture's largest single-face `beginbfchar` section size is reported as a
number against the spec's cap of 100, and `deferred-work.md`'s DW-14 entry updated with it.

**AC12 — The gate table is re-measured with its scope, invocation and counting rule.**
`folio-go` and `lint`, all-occurrences **and** top-level, taken through `rtk proxy` on the raw stream,
against the baseline table above. The only permitted failure is AC10's. Any narration of a count is
**derived from the same computed table the assertion uses** (D-000.48's companion finding), so the
message and the guard cannot disagree.

**AC13 — Every citation this story writes resolves (D-000.48).**
Any comment sweep or doc amendment reports its swept list **with its count**, and every named symbol
is verified mechanically to resolve to a real declaration. An unresolvable citation is worse than none.

---

## Heavy-test cadence — proposed DECLINED, stated so it can be refused

**D-000.4's override criterion is a NEW SOURCE OF CROSS-TARGET DIVERGENCE — float arithmetic, a vendor
call, a compressor, a new dependency — not merely a moved or new golden.** Assessed honestly against
that:

| the criterion | this story |
|---|---|
| float arithmetic | **none.** Pagination is integer comparison and division on `geom.Length` (`int64` millipoints). No new call to `geom.ScaleRound` is required by the rule itself. |
| a vendor call | **none new.** The `textshape` entry points touched are a **subset** of those `collectTextRuns` already reaches; pagination consumes `packLines`' output and calls no shaper. |
| a compressor | **none.** Claimed count 0, and AC11 shows it to be 0. |
| a new dependency | **none.** `go list -m all` stays at exactly two modules. |

**Recommendation: DECLINE the per-story override**, exactly as Stories 2.5 and 2.5a declined it, and
defer the new fixture's four legs to the Epic 2 boundary gate alongside `three-band-page`'s.

**The honest counter-argument, stated so it can be weighed rather than hidden**: this is the first
document in the repository with more than one page, so it is the first that can expose a
target-dependent **page-object ordering** or **xref offset width**. Both are integer and both are
already exercised by the multi-object builder that every font-bearing fixture uses — which is why the
recommendation is still to decline. **If the lead disagrees, the override is cheap here and this
paragraph is the reason to grant it.**

`go vet -tags matrix ./...` must still compile every matrix-tagged file, including the new
`extraGuard`, so the deferral stays honest rather than rotting.

---

## Task breakdown

1. [x] **Re-measure the baseline** (gate table, both scopes, both counting rules, through `rtk proxy`) and
   **reproduce the red-proof of *Measured findings* 1** before touching any code. Both recorded in the
   Delivery Log.
2. [x] **Escalate DN-1…DN-5 and wait.** All five ruled before any pagination code was written
   (D-2.6.1…D-2.6.4, D-000.16 limitation, D-000.50), plus **D-2.6.1's residual arithmetic**, which the
   dev agent escalated and parked on rather than guessing — see *Delivery Log* §2.
3. [x] Implement the ruled pagination rule as a named function in `internal/layout`
   (`layout.Paginate`, `internal/layout/paginate.go`). No field added to `PageGeometry`. Integer
   arithmetic only.
4. [x] Replace `render.go:723`'s one-element slice with the paginated result
   (`paginateDocument`). Object-reservation order verified untouched — all seven pre-2.6 goldens are
   byte-identical.
5. [x] Emit the page header and page footer bands on every page at identical origins.
6. [x] Write AC5's AD-4 structural guard, and **capture its red-proof by mutation before** it goes
   green. Both halves red-proved; `stage-rank` measured green on the same mutation.
7. [x] Author `fixtures/multi-page/` and its in-repo template constant; write the semantic acceptance
   tests (AC3, AC4, AC6) **before** recording `expected.pdf`.
8. [x] Record the golden; register it in all four places (AC7).
9. [x] Write AC8's `{{page}}` pass-through test.
10. [x] Amend `folio-format.md` per D-2.6.3; append to `epic-2-boundary-gate.md`; update
    `deferred-work.md` for DW-11 (AC10) and DW-14 (AC11).
11. [x] Re-run the gate table; confirm AC1's digests and AC10's stats line; verify every citation
    written (AC13).
12. [x] Set the story status to `review`.

### Added mid-story, not in the original breakdown

13. [x] **STOP-AND-FIX: the recorded golden was not a PDF.** `/Kids [8 0 R10 0 R]` — no separator
    between two indirect references — left the page tree unresolvable. Red-proved with `qpdf --check`,
    fixed, re-recorded, guarded three ways. See *Delivery Log* §6, which is the most important section
    in this document.
14. [x] **D-000.53's four work items**: `assertWellFormedPDF` parameterised and called on the
    multi-page subject; `appendRefArray` added so the separator cannot be omitted; an artifact-level
    delimiter guard scoped to non-stream regions; the eight-fixture external-reader evidence recorded
    in every fixture's provenance.

## Flagged, not fixed

1. **`sprint-status.yaml` reads `epic-2: backlog` while `2-1`…`2-5a` are all `done`.** The epic key is
   the **gate's** to flip, not a story's. Flagged, unchanged.
2. **The decision log carries an uncommitted 25-line append (D-000.49).** Present in the working tree
   at `2516b13`, authored by Story 2.5a's finisher, never committed. **This story treats it as binding
   and neither reverts nor commits it.** Whoever commits next should decide deliberately whether it
   rides along — an uncommitted binding ruling is invisible to anyone who clones the repository.
3. **`internal/pdf` → `internal/layout` is a legal import today.** Closed as a *property* by AC5, not
   by a rank change. Recorded because a future reader will assume `stage-rank` already covers it.
4. **`fixtures/expected-breaks/expected_breaks.json` has no digest pin in the ordinary suite.**
   `goldenDigestRecord` covers only `expected.pdf` files; the break fixture's digest is named only by
   the matrix-tagged, deliberately-red sign-off test. It is *behaviourally* pinned by
   `internal/text/s4_expected_test.go`, which would redden on a label change — but a change that moved
   both the labels and the engine together would pass. Not this story's to fix; recorded so the gate
   can weigh it.

---

## Dev Agent Record — Delivery Log

*Written at close, after every event it asserts (D-000.28).*

---

### 1. The gate table, re-measured — not inherited

| scope | invocation (verbatim) | counting rule | baseline | at close |
|---|---|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | every `--- PASS`/`--- FAIL` occurrence | 504 · 1 | **551 · 1** |
| `folio-go/` | the same | top-level only | 312 · 1 | **336 · 1** |
| `lint/` | the same | all occurrences | 85 · 0 | **85 · 0** |
| `lint/` | the same | top-level only | 47 · 0 | **47 · 0** |
| `folio-go/` | `GOWORK=off go list -m all` | — | two | **two** |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | — | clean | **clean** |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet -tags matrix ./...` | — | — | **clean** |

**I re-measured the baseline myself before changing anything and it matched the story's table exactly.**

**The single permitted failure is AC10's**, and its stats line is byte-identical:
`P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.

Counts taken by matching `'"Output":"--- PASS'` versus `'"Output":"    --- PASS'` over the raw
`rtk proxy` stream, per the story's measurement-instrument warning. A naive `grep -c` returns 0 here
and reads as total collapse.

---

### 2. The one decision I refused to make, and why

D-2.6.1 settled the model and the boundary **invariant**, but not the **arithmetic** that satisfies the
invariant for a straddling line — and it adopted DN-1 option (A), whose stated arithmetic
(`[N·H, (N+1)·H)`, `columnY − N·H`, `ceil(lowestBottom/H)`) is a **rigid grid** under which a straddling
line has *no* page containing it entirely. Two readings survived. I **parked and escalated** rather than
guessing.

I eliminated a third by measurement rather than passing it up. On the Noto Sans 12pt chain
(advance 16,344mp, H = 727,890mp) line 44 spans 719,136..735,480 and straddles. A fixed grid with a
per-item bump puts line 44 at 0 and line 45 at 7,590 — **less than the advance, so the two lines
overlap.** A defect worse than the one the story fixes.

**Ruled: (C), the sliding window.** Option (A)'s arithmetic **withdrawn**. The rejected models and the
measurements that killed them are recorded in `internal/layout/paginate.go`'s package comment so neither
can be re-proposed without reading why.

---

### 3. THE RED-PROOF, captured at baseline before any fix existed (D-000.30)

Subject, cited: `folio.Render` over a document with a populated header (18) and footer (24), asymmetric
margins (top 30, bottom 42), one content text element wrapped to a 150pt box over **60 repetitions of
"The quick brown fox jumps over the lazy dog. "**.

| observable | at baseline | after the fix |
|---|---|---|
| PDF bytes | 77,062 | 77,947 |
| `/Type /Page ` objects | **1** | **3** |
| `/Count` | **1** | **3** |
| text placements (`Tm`) | 146 → measured **122** | **126** |
| of those, **negative** PDF Y | **72** | **0** |
| most negative Y | **−1163.874 pt** | **0** |

The sheet is 841.89 pt tall. **59% of the document's text was drawn below the bottom edge. The render
succeeded. Nothing warned.**

**I did not adopt the story's figures** (82,696 B / 146 / 91 / −1000.727). The story does not state
*which* Latin sentence it repeated, so its numbers are not reproducible; I cite my own subject instead.
The defect **class** reproduces exactly, which is what D-000.30 requires.

**The placement accounting is exact and worth more than the totals**: 122 → 126 is precisely the
header+footer pair repeated on two additional pages. Content placements are **conserved at 120**, so no
line was duplicated or dropped.

---

### 4. The pagination model as shipped

`internal/layout/paginate.go`. Window 0 begins at the content band's top; **window N+1 begins at the top
of the first item that did not fit in window N**; page-relative Y is `columnY − windowStart(N)`.

- **Line granularity is FORCED, not chosen** — a single text element can wrap past a page, so element
  granularity would make an element taller than a page unplaceable. Said in those words in the code.
- **No line is ever split.** A line is placed on the first page whose window contains it entirely, from
  `baseline − max(ascent)` to `baseline + max(descent)`.
- **The column is NEVER mutated**, so the shift is stored **once per page**, not per item — pagination
  structurally cannot displace one item relative to another. That is what makes AD-24's *"siblings never
  move, because nothing in a band ever reflows"* true here rather than argued.
- **Page count is not a closed form.** `ceil(lowestBottom/H)` is gone and
  `TestPaginateNeverProducesAnEmptyPage` names `51` in its failure message as the signature of its
  return.
- **Sweep runs in COLUMN order; emission runs in AUTHORED order.** That split is what keeps a one-page
  document byte-identical, and `TestPaginateEmitsContentInAuthoredOrder` pins it with an element
  declared first that sits below one declared second.

**Red-proof of the whole model**: disabling the window advance — which *is* the pre-2.6 degenerate
partition — reddens four tests including AC4's, while the "fits exactly" and "no content" rows correctly
stay green.

---

### 5. FR44's overflow disposition — a divergence I flagged rather than absorbed

D-2.6.1's amendment said *"clip at the window bottom"*. I implemented a **located error** instead and
**flagged it immediately rather than shipping it quietly**: `Render` returns `([]byte, error)` and has no
diagnostic channel, clipping is Story 2.8's machinery, and D-2.6.1's own image ruling says *"never a
silent clip … a template error with a located message."*

**D-2.6.5 ruled the divergence correct** and withdrew the clip disposition, on a stronger ground than
mine: both subjects are **declaration-level, not render-time** — leading is a function of the declared
font stack and size (D-2.4.2 constraint 1) and the window is page height minus declared margins and band
heights, so *"some line is taller than the window"* is decidable from template plus fontset **with no
data at all**, exactly as the image's declared box is. Two declaration-level impossibilities, one
disposition.

**Both subjects are exercised** through the public `folio.Render` on **synthetic templates**, asserting
the error's `ElementID` and `Kind` — never *"an error occurred"*, which a missing face or a malformed
asset also satisfies. Plus a **negative control** (`TestRenderAcceptsAnItemThatExactlyFitsTheWindow`),
because D-000.34 (extended) says a fix can destroy a negative control invisibly.

The `Kind` classifier derives `"image"` from `len(it.Images) > 0`, so **item exclusivity is asserted
rather than assumed** — `MixedItemError` rejects an item carrying both or neither, distinguishing the two
because they have different causes.

**Carried to Story 2.8**: `epics.md:938-950` requires its box-overflow case to clip **and** diagnose
**and** still return bytes. That is not expressible on `Render`'s current surface. **2.8 owns a
diagnostic-channel design decision on the public API.**

---

### 6. THE MOST IMPORTANT SECTION — a golden passed its hash and its acceptance step and was not a PDF

**Read this before the AC table.**

I recorded `fixtures/multi-page/expected.pdf`. It passed its golden hash. It passed the whole semantic
acceptance step I had written *before* recording, exactly as D-000.22 prescribes. **The owner opened it
and it would not render.**

```
$ qpdf --check fixtures/multi-page/expected.pdf
WARNING: (object 2 0, offset 99): unknown token while reading object; treating as string
WARNING: object 2 0 at offset 95: Pages tree includes non-dictionary object; ignoring
ERROR: file does not contain any pages
qpdf: errors detected                                                          exit=2
```

Emitted object 2: `<< /Type /Pages /Kids [8 0 R10 0 R] /Count 2 >>`. **No separator between the two
indirect references.** A PDF tokenizer reads `R10` as one unknown token, so **neither** kid resolves, the
page tree is empty, and a viewer shows *"page 0 of 2"*.

#### Why every check I wrote was satisfied by an unrenderable file

| the check | why it passed |
|---|---|
| `/Type /Page` objects == 2 | counted **object definitions**. Both page objects were emitted correctly and both were in the xref. What was broken was the **array pointing at them** — a different object. |
| `/Count 2` | read the integer. The integer was right. **`/Count` is a claim about the kids, not a fact derived from them.** |
| per-page header/footer Y, the pinned line→page partition, no negative Y | parsed the **content streams** directly, by splitting the file on stream boundaries. Content streams do not care whether anything references them. Independently confirmed: a length-preserving hand patch of the recorded bytes rendered correctly with **every one of these properties already right**. |
| byte-identity across two processes | says the output is **deterministic**. A deterministically wrong file is byte-identical to itself. |

**Every assertion measured PRESENCE. The property that mattered was REACHABILITY.** Every check asserted
a property of a **part**; the defect was in a **reference between** parts. No quantity of additional
per-part assertions would have found it.

**And it is worse than "the acceptance step was incomplete."** `folio-go/render_test.go`'s
`assertWellFormedPDF` — the module's one general well-formedness checker, with **18 call sites** —
hard-coded `pageCount != 1`. It would have **fatalled on a multi-page document on the very property that
made the document interesting.** So it was never called on the first multi-page golden. **The only
instrument that could have looked encoded the single-page shape as an invariant, and was therefore
silently never invoked** — while its 18 other call sites made the module look covered.

**A hash certifies that bytes have not changed since they were recorded. It is silent on whether they
were right when they were recorded.** The semantic acceptance step is the only thing standing in that
gap, and here it was assembled entirely out of checks a broken file satisfies.

#### Fixed, and guarded in four independent ways

1. **Prevention by construction** — `appendRefArray`/`writeRefArray` (`internal/pdf`) emit the array from
   the slice and **cannot** omit the separator. Leading separator, so the one-kid case stays `[8 0 R]`
   and no golden moved. Ruled in over my narrower scope: `appendRef` has ~13 call sites and there are
   two page-tree emitters; `/Kids` being the only ref array is a fact about today's feature set, not
   about the code.
2. **Token well-formedness at the artifact** — `assertIndirectRefsAreDelimited`, inside
   `assertWellFormedPDF` so it covers all 19 call sites. Keyed on the defect's true shape (D-000.15): a
   `\d+ 0 R` token whose following byte is not PDF white space or a delimiter. Scoped to **non-stream
   regions**, exactly — every stream declares `/Length` up front and `assertStreamLengthsAreExact` has
   already validated the framing, so the regions are known rather than guessed, and the false-positive
   hazard that would have forced the escape hatch does not arise. **Labelled a forward guard (D-000.24)**
   for every ref array other than `/Kids`, where no red-proof is available because no other ref array
   exists yet.
3. **Page-tree semantics over every golden** — `folio-go/golden_structural_validity_test.go`, hermetic
   (no shell-out, so it runs on `js-wasm` too), iterating `goldenDigestRecord`. It groups `/Kids` tokens
   **in threes**, which is what exposes a run-together pair: **a substring search for `"8 0 R"` and
   `"10 0 R"` finds both in `[8 0 R10 0 R]` and reports success.** It also requires each reference to
   resolve to a defined `/Type /Page` and no page object to be orphaned.
4. **The repaired checker, pointed at the subject** — `assertWellFormedPDF` now takes `wantPages`.
   `pageCount != 1` was **parameterised, not deleted** (D-000.34: on seven single-page fixtures that
   comparison is load-bearing). **All 20 call sites** pass a count; `matrixDocument` gained a
   `wantPages` field whose zero value reads as 1, so the seven pre-2.6 entries needed no edit.

These are **three properties and three mechanisms, not three coverages of one** (D-000.42): construction,
token well-formedness, page-tree semantics.

**Red-proofs**: the structural oracle reddened on the real broken bytes, naming `"8 0 R10"`, and **only
`multi-page` failed** — the other seven stayed green. The delimiter guard reddened when the separator was
removed from `appendRefArray`, with the single-page fixtures correctly staying green. Both restored,
verified by empty `git diff`.

#### The re-record, measured against the prediction

| | |
|---|---|
| old | 66,524 bytes, `83344565…a489c988` |
| new | **66,525 bytes**, **`66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b`** |
| delta | **+1 byte** |
| BT/ET blocks | old 33, new 33, **IDENTICAL** |
| bytes before object 2 | identical |
| bytes between object 2 and the xref | identical |

**No layout moved.** Only object 2 and the xref offsets that follow it.

#### D-000.53 — ratified from this finding

> No golden artifact is accepted — first recording **or re-recording** (D-000.44) — until a reader this
> project did not write parses it and resolves it into the semantic objects it claims to contain.

`qpdf` **12.4.0**, verbatim invocations, at `50ad6c8`. All eight recorded in each fixture's own
provenance section:

```
minimal-rect  font-text  image-embed  multi-script-fallback
shaped-text   wrapped-text  three-band-page      exit=0  pages=1
multi-page                                       exit=0  pages=2
```

Seven read **"settled, validated unchanged"** (they predate D-000.53 and were validated retroactively —
the row ends settled, not carried, D-000.29). The eighth names the breakage and the fix.

The external reader is the **acceptance instrument at recording time only** — never a runtime or CI
dependency (AD-25), and deliberately **not** gated to "the legs that have qpdf", because a check running
on some legs and not others reproduces D-000.9's failure — an "all clear" indistinguishable from "I could
not look" — one level up at the leg. The **in-repo hermetic checker is the standing every-leg guard**.

**This did not become a gate obligation.** D-2.6.2's criterion is *the only cross-target artifact for a
shipped FR*, and structural validity of committed bytes is not cross-target — the bytes are identical on
every leg by construction. **The Epic 2 gate still owes exactly five.**

---

### 7. Acceptance criteria — each with what carries the property

| AC | verdict | the artifact, and the measurement |
|---|---|---|
| **AC1** | **met** | All seven pre-2.6 `expected.pdf` digests **byte-unchanged**, re-hashed by the file route: `git status --porcelain fixtures/` lists **no `expected.pdf` as modified**, and `goldenDigestRecord` agrees 8/8. `expected_breaks.json` unchanged at `a545e042…9324de`. `git diff --stat -- fixtures/` at close lists the new directory **plus** the eight provenance appends required by D-000.53 — a both-directions statement (D-000.49): the claimed absence shown absent, the claimed additions shown present. `shaped-text` = `6c040ef7…c6c85370`, re-computed at close. |
| **AC2** | **met** | `TestPaginatePageCountFromADeclaredTable` — a declared table with rows for fits-exactly (1), one line over (2), **one millipoint** over (2), three windows (3) and zero content (1), plus `TestMultiPageRendersTheDeclaredNumberOfPages` at two independent byte sites. Non-displacement is **named, not restated** (D-000.42): `TestContentHeightDependsOnGeometryAlone` and `TestNoBandOriginArithmeticInPackageFolio`. |
| **AC3** | **met** | `TestMultiPageHeaderAndFooterAppearOnEveryPage` — header at **798.269 pt** and footer at **51.448 pt** on **both** pages, hand-derived (`841890 − 30000 − 4000 − 9621`; `841890 − 30000 − 751890 − 8552`). Distinct literals. **Presence precondition**: both bands asserted non-empty in the input first. |
| **AC4** | **met** | `TestMultiPageLineToPagePartitionIsPinnedByValue` + `TestPaginatePinsTheLineToPagePartitionByValue`. Declared partition **22 / 7** over 29 lines, hand-derived from `(i+1)·32688 ≤ 727890 → i ≤ 21`; every page asserted **non-empty**; the boundary index asserted, never the sum (D-000.33). |
| **AC5** | **met** | `TestInternalPDFReachesNoLayoutComputation` (transitive edge + a **2,318-identifier** sweep against 18 banned substrings) and `TestPassTwoEmitsExactlyOnePageObjectPerInputPage` (declared table, **three** independent byte sites). **Red-proved both ways**, and — the load-bearing measurement — with the forbidden import present, `stage-rank` stayed **fully green across all four of its tests** while the new guard reddened. |
| **AC6** | **met** | `fixtures/multi-page/` with all four files and a byte-identical in-repo constant. Acceptance tests written **before** recording. **See §6**: the machine-checkable half was necessary but not sufficient, and is now backed by an external reader and a hermetic structural oracle. |
| **AC7** | **met** | `matrixDocuments` (slug, capture, `extraGuard` proving genuine multi-pageness, `wantPages: 2`), `matrix.yml` (4 per-target artifact lines + the `docs=` line), `goldenDigestRecord`, and `declaredEpic2GateObligations` as a **one-line** addition. All three registration guards green. |
| **AC8** | **met** | `TestReservedPagePlaceholdersPassThroughOnEveryPage` — the literal `Page {{page}} of {{pages}}` recovered from the **drawn glyphs** via `/ToUnicode` on **every** page of an N=2 render, with an N≥2 presence precondition. `TestReservedPlaceholderSetIsUnchanged` pins the set **and** its negative half: a non-reserved placeholder must still error, or "everything passes through" would make the positive rows vacuous. |
| **AC9** | **met** | §6 above and `epic-2-boundary-gate.md`'s appended section carry the five-obligation table with each fate. `declaredEpic2GateObligations`' diff is **exactly one line**. |
| **AC10** | **met** | `corpus.json` and `cmd/gencorpus/main.go` byte-unchanged; the failure still reports `P6g … got 7, need >=20` with the stats line character for character. DW-11 answered in writing in `deferred-work.md`: **none were found, and none were invented** — pagination cannot reach `internal/text`. |
| **AC11** | **met** | `go list -m all` = **exactly two**. `SOURCE_DATE_EPOCH` in `.go` under `folio-go/`: **1**, and it is `testdata/lint/absences/violating/` — the rule's own violating fixture, which is the file that must contain it. Compressor imports: claimed **0**, **shown** 0. `fmt` in non-test `internal/pdf`: **0**. Float guards: zero. **DW-14 re-measured: 45** of a 100 cap — the largest section any fixture has recorded, from a two-page single-face document. Reported in **both directions** (D-000.49) in `deferred-work.md`. |
| **AC12** | **met** | §1. Both scopes, both counting rules, through `rtk proxy` on the raw stream. |
| **AC13** | **met** | Every symbol cited here was verified to resolve. **Two corrections found and reported** — see §8. |

#### The call-site enumeration, and a counting mistake I made twice

`assertWellFormedPDF` has **20 call sites** and one definition — 21 occurrences under `folio-go/`
excluding `testdata/`:

| file | call sites |
|---|---|
| `render_test.go` | 6 |
| `fixture_test.go` | 4 |
| `three_band_page_fixture_test.go` | 3 |
| `matrix_test.go` | 2 |
| `multi_page_fixture_test.go` | 2 |
| `first_baseline_acceptance_test.go`, `shaped_fixture_test.go`, `wrapped_text_fixture_test.go` | 1 each |

**I reported 19 twice and it was 20 both times** — the same class of error as reporting "both declared
sites" when D-000.47 names **four**. The root cause is the same in both cases and it generalises past
this story:

> **I counted what I intended to add, not what the artifact contains.** For the digest I hand-edited two
> sites and reported two — the recorder had already written the other two. For the checker I added one
> *test* and reported one *call site* — that test calls it **twice**, on the fresh render and on the
> committed bytes. Under D-000.47 the site list **is** the mechanism, so a count taken from intent
> rather than from the artifact is the same defect the mechanism exists to prevent, one level up.

**Enumerating rather than adjusting the number to match found a real defect.** `matrix_test.go` has
**two** leg-validation sites, not one: `:1020` in `TestCrossTargetByteIdentity` and `:1212` in the
per-target CI path. My mechanical substitution passed a literal `1` at the second, which loops over
**every registered document** — so it would have **fatalled on `multi-page` at the Epic 2 gate**, the
only place that matrix-tagged path runs and the last place anyone wants to find it. Both now use
`doc.pageCount()`. **That is the same defect class as the one that let the broken golden ship: a checker
encoding the single-page shape as an invariant.**

---

### 8. Citations corrected, and premises sharpened (D-000.48, D-000.49)

1. **AC2 cites `TestContentHeightDependsOnGeometryAlone` as living in
   `internal/bandcomposition_arch_test.go`.** It resolves, but in **`internal/layout/band_test.go:104`**.
   The other named guard, `TestNoBandOriginArithmeticInPackageFolio`, *is* in the cited file (`:498`).
2. **D-2.6.4 says `expected_breaks.json`'s "only digest pin is matrix-tagged today."** Measured: the
   literal occurs at **ZERO** sites. The matrix-tagged test **computes** the digest at run time and
   compares it against `fixtures/expected-breaks/break-signoff.json`, **which does not exist**. So the
   file was pinned **nowhere** and `expected_breaks_digest_test.go` is its **first** pin, not its second.
   The ordered remedy is unchanged; the premise was weaker than stated. **Two different things are
   explained by that one absent file**: the missing sign-off record making the matrix test fail is the
   pending-sign-off blocker **working as designed**; the unpinned fixture was the genuine hole.
3. **My own red-proof figures differ from the story's** (§3). The story does not state which sentence it
   repeated. I cite my own subject rather than reconcile to numbers I cannot reproduce.

**D-2.6.4's pin has a discriminating red-proof**: mutating a `gloss` field — which **no engine consults**
— was **invisible** to the pre-existing ordinary suite while the new pin caught it. Mutating a field the
engine *does* read would have proven nothing, because `internal/text/s4_expected_test.go` catches that
anyway.

---

### 9. Flagged, not fixed

1. **`sprint-status.yaml` reads `epic-2: backlog`.** The epic key is the **gate's** to flip. Unchanged.
2. **`expected_breaks_signoff_matrix_test.go:20`** carries a comment naming digest `5964aad0…c92e00f` as
   belonging to `fixtures/shaped-text/thai-signoff.json`. That file does not exist and `5964aad0` is the
   **superseded** digest of `shaped-text/expected.pdf` (now `6c040ef7…`). **Comment-only, no live
   assertion.** Routed to the reviewer; untouched here.
3. **DW-14 is the closest it has ever been to firing** — 45 of 100 from a two-page single-face document.
   A document roughly twice as long in one face reaches the cap, and that is an ordinary report. The fix
   stays the gate's: it moves the golden hash of every document over the cap.
4. **Story 2.8 inherits a public-API design decision**, not just clip machinery — see §5.
5. **The decision log's uncommitted append** (D-000.49) is still uncommitted. Treated as binding, neither
   reverted nor committed.

---

### 10. Files changed

**New** — `folio-go/internal/layout/paginate.go`, `paginate_test.go`; `folio-go/internal/passtwo_arch_test.go`;
`folio-go/internal/pdf/passtwo_pagecount_test.go`; `folio-go/multi_page_template.go`,
`multi_page_fixture_test.go`, `render_overflow_test.go`, `reserved_placeholders_test.go`,
`expected_breaks_digest_test.go`, `golden_structural_validity_test.go`; `fixtures/multi-page/` (4 files).

**Modified** — `folio-go/render.go` (the seam, run tagging, `paginateDocument`);
`folio-go/internal/pdf/textdoc.go` (`/Kids` via `appendRefArray`), `document.go` (`appendRefArray`),
`builder.go` (`writeRefArray`); `folio-go/render_test.go` (`assertWellFormedPDF` parameterised,
`assertIndirectRefsAreDelimited`, the 8th subprocess selector), `matrix_test.go`, `byte_neutrality_test.go`,
and the call sites passing `1`; `.github/workflows/matrix.yml`;
`_bmad-output/specs/spec-folio/folio-format.md` (the pagination clause);
`_bmad-output/implementation-artifacts/deferred-work.md`, `epic-2-boundary-gate.md`; the seven pre-2.6
fixture READMEs (provenance only — **no artifact touched**).

---

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-24
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 1
- **Majors:** 4
- **Minors:** 10
- **Nits:** 5
- **Finisher disposition (2026-08-24):** All Changes Requested addressed. 1 Blocker + 4 Majors + 12
  Minor/Nits FIXED and validated (see *Finisher's Delivery Log*, below the AC-by-AC disposition); 2
  Nit/Minor findings DISMISSED on their merits (Finding 14's premise is contradicted by this file's own
  ruled model, verified empirically; Finding 18 by the reviewer's own "no hole" admission); 1 Nit
  ACKNOWLEDGED with no action, per the reviewer's own suggested resolution. Zero findings deferred as new
  obligations. Status: **Done**.

### What I measured, and what I only read

**Measured (ran myself):**

| what | invocation (verbatim) | result |
|---|---|---|
| `folio-go` suite | `rtk proxy env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **551 PASS · 1 FAIL** all-occurrences; **336 · 1** top-level |
| build | `CGO_ENABLED=0 GOWORK=off go build ./...` | clean |
| vet (matrix) | `CGO_ENABLED=0 GOWORK=off go vet -tags matrix ./...` | clean |
| gofmt | `gofmt -l .` (bare, **not** through `rtk proxy` — it swallows list output) | empty |
| external reader | `qpdf --check` + `--show-npages` on all eight goldens | **8/8 exit 0**; `multi-page` = 2 pages, seven = 1 |
| sign-off artifacts | `shasum -a 256` | `shaped-text` = `6c040ef7…c6c85370` **UNMOVED**; `expected_breaks.json` = `a545e042…9324de` **UNMOVED** |
| `multi-page` golden | `shasum -a 256`, `ls -l` | `66ce0ee4…f895869b`, **66,525 bytes** |
| **the matrix leg** | `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash -v .` | **FAIL** — see Finding 1 |

The single failure is AC10's `TestCorpusMeetsP6ExerciseFloors`, stats line byte-identical:
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. **Not reported as a break.** No second failure in
the ordinary suite. Every number in the Delivery Log's §1 gate table reproduces exactly.

**Read only (not run):** the four-target matrix legs *as a cross-target comparison* — but the per-target
leg itself I **did** run natively on `darwin/arm64`, which is how Finding 1 was found. `.github/workflows/matrix.yml`
was read, not executed.

**Red-proofs I performed myself** (each mutation reverted by hand and the file's sha256 re-verified;
`fixtures/` and every tracked file byte-unchanged at close):

| mutation | result |
|---|---|
| separator removed from `appendRefArray` | **RED** — `assertIndirectRefsAreDelimited` fired at `multi_page_fixture_test.go:603`; `internal/pdf`'s `TestPagesTreeKidsAreSeparated` fired for 2, 3 and 7 pages. Seven single-page fixtures stayed **green**, confirming the leading-separator design moves no committed artifact. |
| identifier `columnWindowProbe` added to `internal/pdf` | **RED** — AC5's sweep fired: `AD-4 VIOLATED: internal/pdf/document.go declares or references "columnWindow…"`. |
| `TestEveryGoldenPDFResolvesItsPageTree` under the same emitter mutation | **stayed GREEN** — it reads *committed* bytes, not fresh renders. This is correct and is positive evidence for D-2.6.8: layers 1/2 guard **emission**, layer 3 guards **recorded artifacts**. They are genuinely different properties, not three coverages of one. |

**Independent re-derivations of the story's own arithmetic — all correct, no error found:**
`841890−30000−4000−9621 = 798269` (798.269 pt) · `841890−30000−751890−8552 = 51448` (51.448 pt) ·
`841890−30000−42000−18000−24000 = 727890` · `(i+1)·32688 ≤ 727890 → i ≤ 21` → **22 / 7** over 29 lines.
DW-14's **45** reproduced from the artifact's own `beginbfchar` header (and `wrapped-text`'s 38 and
`shaped-text`'s 28 cross-check).

### Binding-ruling compliance

| ruling | verdict |
|---|---|
| **D-000.53** | **SATISFIED.** All eight fixture READMEs carry a provenance block naming reader (`qpdf`), version (**12.4.0**), **verbatim invocation**, result and commit (`50ad6c8`, which resolves). I re-ran the oracle: 8/8 exit 0. |
| **D-2.6.6** | **SATISFIED.** `epic-2-boundary-gate.md:262` declares **FIVE**, and its table lists exactly five. The eight-fixture audit is recorded as evidence and the structural-oracle standing rule is explicitly *"PROPOSED … raised, not assumed"*. No sixth obligation. `declaredEpic2GateObligations`' diff is **exactly one line** (verified against `git diff`). |
| **D-2.6.7** | **SATISFIED.** No `qpdf` reference exists in any `.go`, `.yml`, `.sh` or Makefile except comments. The in-repo oracle is hermetic and runs over all 8 goldens on every leg. `assertWellFormedPDF` was **parameterised, not loosened**: `pageCount != wantPages` is kept and all seven single-page fixtures still pass a literal `1`. |
| **D-2.6.8** | **PARTIALLY SATISFIED.** The three layers are genuinely three properties (evidenced by the divergent red-proofs above). But layer 1 was applied to **one of the three** ref-array sites in the module — Finding 3. |
| **D-000.24 (forward guards)** | **NOT SATISFIED** — the forward-guard premise is factually false. Finding 3. |

### Answers to the two questions the developer asked to be looked at hardest

**1. `assertIndirectRefsAreDelimited`'s non-stream scoping — the scoping is EXACT, and the guard should
NOT be dropped.** I replicated `nonStreamRegions`' walk independently in Python and ran it over all
eight goldens: **stream-body bytes leaking into a scanned region = 0 on every fixture**, including the
`<< /Length N /Length1 M >>` font-program shape (the trickiest, because `/Length1 ` does not match the
`"/Length "` needle — verified) and the image-XObject shape. I then measured the false-positive
*potential*: if the scoping failed completely and every stream body were scanned, the number of
`\d+ 0 R`-followed-by-non-delimiter hits across all eight artifacts is **0**. So this is not a
false-positive generator, on today's corpus it could not be one even if broken, and D-000.15 does not
call for dropping it. The developer's argument from `assertStreamLengthsAreExact` is **sound as
written** — that function `t.Fatalf`s on every condition where `nonStreamRegions` silently `break`s,
and it runs immediately before, so the break branches are unreachable. The residual issue is the
undeclared lockstep coupling between two independent copies of the same walk — logged as Finding 8, a
Minor, not a reason to drop the guard.

**2. The forward half — the label is applied, but its stated premise is FALSE.** See Finding 3.

### The pattern I was asked to check independently: the story miscounts again

Re-derived from the artifact, not from the record:

| the story's claim | measured | verdict |
|---|---|---|
| `assertWellFormedPDF` has **20 call sites** | **20** | ✅ correct — and the per-file table (6/4/3/2/2/1/1/1) is correct row by row |
| four declared digest sites for `fixtures/multi-page` | 3 `sites` entries + the artifact = **4** | ✅ correct |
| "**21 occurrences** under `folio-go/` excluding `testdata/`" | **27** | ❌ wrong |
| "so it covers all **19 call sites**" (§6 item 2) | **20** (pre-story was **18**, measured at HEAD) | ❌ wrong |
| "a **2,318-identifier** sweep" | the test logs **2348** | ❌ wrong |

**Three further instances of the class the story diagnoses, two of them inside the very paragraph that
diagnoses it.** Logged as Finding 9.

**And the same-shape mechanical-edit sweep the story asked for came back positive** — not in
`matrix_test.go` (which is now clean: both `assertWellFormedPDF` sites use `doc.pageCount()`, and every
other literal `1` there is a slice offset), but in `requireMultiPageIsGenuinelyMultiPage`, which
encodes the single-page shape through a helper. That is Finding 1.

---

## Review Findings

### Finding 1: The `multi-page` matrix leg fails unconditionally on every target, and CI turns red on the next push to `main`
- **Severity**: **Blocker**
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/matrix_test.go:593` and `:606-618`; helper at `folio-go/shaped_fixture_test.go:382-400`; CI trigger at `.github/workflows/matrix.yml:25-29`
- **Observation**: `requireMultiPageIsGenuinelyMultiPage` obtains its runs from `readEmittedRuns`, which selects **exactly one** content stream and stops:
  ```go
  for _, obj := range objs {              // objs is a map[int][]byte
      s, ok := streamBody(obj)
      if !ok { continue }
      if bytes.Contains(s, []byte(" Tf\n")) { content = s; break }   // FIRST ONLY
  }
  ```
  `fixtures/multi-page/expected.pdf` has **two** content streams, one per page, each carrying exactly one header run and one footer run. The guard therefore counts 1 and 1 and fatals on `if headerRuns != 2 || footerRuns != 2`. I reproduced this natively, on the real CI entry point:
  ```
  $ FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash -v .
  matrix_test.go:1233: darwin/arm64: the multi-page leg carries 1 page-header run(s) and 1
    page-footer run(s) at their declared Y; a two-page document must carry exactly 2 of each
  --- FAIL: TestTargetRenderHash (0.42s)
  ```
  The other seven documents hashed and matched their goldens; `multi-page` fatals before its hash file is written. Secondarily, `pdfObjects` returns a **map**, so *which* page is read is nondeterministic run to run — even after the count is fixed, any per-run assertion built on this helper over a multi-page document is coin-flip coverage.
- **Impact**: `matrix.yml` is `on: push: branches:[main]` / `pull_request` with **no condition**; all four `render-*` jobs call `TestTargetRenderHash`. The next push to `main` turns **all four legs red** and `compare-render-hashes` never runs (`needs:` unsatisfied). The story's deferral honesty check was `go vet -tags matrix ./...` — which **compiles but does not run**, so it could not see this; I confirm vet is clean *and* the leg fails. AC7's verdict "**met** … `extraGuard` proving genuine multi-pageness" is false as recorded. **This is the story's own diagnosed defect class, committed fresh**: a helper encoding "one content stream per document" as an invariant, exactly as `assertWellFormedPDF` encoded `pageCount != 1`. The story fixed the instance it found and shipped another.
- **Suggested Resolution**: Make `readEmittedRuns` (or a multi-page-aware sibling) iterate **all** text-bearing streams in deterministic object-number order — `multi_page_fixture_test.go:304-325`'s `splitPageContentStreams` already exists for exactly this. Then re-run the native leg and record the result. Consider whether a "written but not run" matrix leg can honestly be declared deferred when a single native invocation would have caught it.
- **Related AC**: AC7 (and AC3, whose per-page property this guard claims to carry onto every leg)

### Finding 2: The `multi-page` extraGuard would have passed on the exact broken bytes the story exists to prevent
- **Severity**: Major
- **Category**: Tests / Correctness
- **Location**: `folio-go/matrix_test.go:586-591`
- **Observation**:
  ```go
  if got := bytes.Count(raw, []byte("<< /Type /Page /Parent ")); got != 2 { ... }
  if !bytes.Contains(raw, []byte("] /Count 2 ")) { ... }
  ```
  The shipped-broken golden emitted `<< /Type /Pages /Kids [8 0 R10 0 R] /Count 2 >>` and had **two** correct `/Type /Page` objects. That page-tree line contains the substring `] /Count 2 ` **verbatim**. Both assertions pass. Content geometry was correct (the defect was purely in the reference between objects), so the negative-Y sweep passes too, and header/footer would have been 2/2 had Finding 1's helper worked.
- **Impact**: The guard the record calls *"an `extraGuard` that proves the document is genuinely multi-page"* is blind to the only defect that has ever made this document not a document. It measures **presence** (`/Count`, object definitions) — precisely the category §6 identifies as insufficient — and adds no **reachability** property. What actually catches the defect on a leg is `assertWellFormedPDF` → `assertIndirectRefsAreDelimited` at `matrix_test.go:1020` and `:1212`, a different mechanism. The record overstates the guard (D-000.49 class). Also, `bytes.Count(raw, "<< /Type /Page /Parent ")` is strictly weaker and more formatting-brittle than `countPageObjects`, which `assertWellFormedPDF` ran two lines earlier with the same expected value.
- **Suggested Resolution**: Either give the extraGuard a genuine reachability assertion (resolve `/Kids` and require each kid to be a defined page object — i.e. reuse `golden_structural_validity_test.go`'s logic), or relabel it in AC7 and in the registry comment as what it is: a *page-count and geometry* guard whose well-formedness half is carried by `assertWellFormedPDF`.
- **Related AC**: AC7

### Finding 3: Layer 1 ("prevention by construction") was applied to one of the three ref-array sites the module has — and the forward-guard premise "no other ref array exists yet" is false
- **Severity**: Major
- **Category**: Correctness / Convention
- **Location**: `folio-go/internal/pdf/document.go:81-83` and `:220-224`; `folio-go/internal/pdf/textdoc.go` (`/DescendantFonts [`); `folio-go/render_test.go:730-734`; story §6 item 2 and line 819
- **Observation**: `appendRefArray`'s own comment states the reasoning it refuses to rely on:
  > *"The missing separator is not a fact about /Kids. `appendRef` has around thirteen call sites and there are **TWO page-tree emitters** (document.go's single-kid literal and textdoc.go's loop); /Kids being the only ref ARRAY today is a fact about today's feature set, not about this code."*

  Only `textdoc.go:192` was converted. The other two array-of-references sites are still hand-rolled with the identical `"[" + writeRef + "]"` shape that produced the bug:
  - `document.go:81-83` — the second page-tree emitter, named in the comment above.
  - `textdoc.go` — `" /Encoding /Identity-H /DescendantFonts ["` + `b.writeRef(ids.cidFont)` + `"] /ToUnicode "`.

  I enumerated every dict key whose value is an array containing an indirect reference, across all eight committed goldens: **`/Kids` and `/DescendantFonts`**. `/DescendantFonts` is live in **six of the eight** goldens. So `render_test.go`'s claim — *"For every OTHER ref array in this module there is no red-proof available, because no other ref array exists yet"* — and the story's identical sentence are **measurably false**.
- **Impact**: Both unconverted sites are safe **only** because they are always single-element — which is exactly the "fact about today's feature set" the design comment says it declines to depend on. The forward-guard label at `render_test.go:730-734` rests on a false premise, so a reader cannot tell which part of layer 1 is proven and which is aspiration (D-000.24's whole purpose). D-2.6.8's "prevention by construction" is therefore true of the page tree only.
- **Suggested Resolution**: Route `document.go:81-83` and `/DescendantFonts` through `appendRefArray`/`writeRefArray` (both are byte-neutral for a one-element array by the leading-separator design, so no golden moves — the red-proof above confirms single-page fixtures stay green). Then correct the premise sentence in `render_test.go`, `document.go` and the story: `/Kids` is not the only ref array; it is the only ref array that is ever **multi-element** today, which is a different and weaker statement.
- **Related AC**: AC5 / the §6 remedy (layer 1)

### Finding 4: The structural oracle's orphan check compares COUNTS, not SETS — a duplicated reference plus an orphaned page passes
- **Severity**: Major
- **Category**: Tests / Correctness
- **Location**: `folio-go/golden_structural_validity_test.go:166-170`
- **Observation**:
  ```go
  defined := bytes.Count(b, []byte("<< /Type /Page /Parent "))
  if defined != declaredCount { t.Errorf(...) }
  ```
  I replicated the oracle's exact four checks in isolation and probed them:

  | input | oracle verdict |
  |---|---|
  | healthy `[8 0 R 10 0 R]` | PASS |
  | the real defect `[8 0 R10 0 R]` | **FAIL** — `malformed ref "8 0 R10"` ✅ |
  | `[8 0 R 8 0 R]`, objects 8 and 10 both defined | **PASS** ❌ |

  In the third case `len(refs) == declaredCount == 2`; both references resolve to a defined `/Type /Page`; `defined == 2 == declaredCount`. Object **10** is a defined page object that nothing references, and page 2 is a duplicate of page 1.
- **Impact**: The test's own docblock claims check (4) means *"a page object that exists but is unreachable, **which is the general shape of this defect**, fails here too."* It does not. The oracle catches the specific run-together instance and misses the general class it advertises — the same presence-vs-reachability substitution §6 was written to end, one level up. A file in this state renders two identical pages and drops a page's content silently.
- **Suggested Resolution**: Compare the **set** of referenced object numbers against the set of defined `/Type /Page` object numbers, and additionally require the referenced numbers to be pairwise distinct. Then re-red-proof with a duplicate-reference mutation.
- **Related AC**: AC6 / the §6 remedy (layer 3)

### Finding 5: AC2's stated composition is asserted at exactly one row of its own table
- **Severity**: Major
- **Category**: AC Conformance
- **Location**: `folio-go/internal/layout/paginate_test.go:136`; `folio-go/internal/pdf/passtwo_pagecount_test.go:50-61`; `folio-go/multi_page_fixture_test.go:406-411`
- **Observation**: AC2 requires *"for each row of a declared table of (page geometry, content extent) → expected page count, the **rendered document's `/Count`** and **its number of `/Type /Page` objects** both equal the declared integer."* The chain is shipped in two halves that meet at one point:
  - `paginate_test.go:136` has the (geometry, content extent) table — rows for fits-exactly (1), one-line-over (2), **one-millipoint-over** (2), three windows (3) and zero content (1), all `wantPages` hand-written literals — but asserts only `len(plan.Pages)`, a Go value, never PDF bytes.
  - `passtwo_pagecount_test.go` asserts `/Count`, `/Type /Page` and `/Kids` for `{1,2,3,7}` — but its input is `inputPages int`, not (geometry, content extent).
  - `multi_page_fixture_test.go:409` asserts `/Count` on a rendered document for **N = 2 only**.

  A repo-wide sweep for `/Count \d` in non-test-data Go finds only `/Count 2`. **No 3-page rendered document exists anywhere in the tree.**
- **Impact**: The composition AC2 specifies — geometry+content all the way through to bytes — is verified at a single point. A defect in the seam between `Paginate` and `SerializeTextDocument` that manifests only at N=1, N=3 or N=0 is invisible. Both halves are individually good; the claim that AC2 is "met" over the declared table is stronger than what ships.
- **Suggested Resolution**: Render at least the 1-page, 3-page and zero-content rows through `folio.Render` and assert `/Count` and `/Type /Page` on the produced bytes, so the declared table drives the artifact end to end.
- **Related AC**: AC2

### Finding 6: `expected_breaks_signoff_matrix_test.go` cites a superseded digest and a file that does not exist
- **Severity**: Minor
- **Category**: Maintainability / Convention (D-000.48)
- **Location**: `folio-go/expected_breaks_signoff_matrix_test.go:19-20`
- **Observation**: Confirmed independently. The comment reads *"fixtures/shaped-text/thai-signoff.json and its digest 5964aad0…c92e00f are not touched, read, or extended by this file."* `thai-signoff.json` does not exist anywhere in the tree, and `5964aad0…` is the **superseded** digest of `shaped-text/expected.pdf`, moved to `6c040ef7…c6c85370` by Story 2.5a. I verified it is **comment-only**: `shaped_signoff_matrix_test.go:49` and `byte_neutrality_test.go:499` build the path from a constant, and no live assertion reads this literal. A repo-wide sweep for `5964aad0` finds it elsewhere **only** in closed, append-only story records (2-3, 2-3a, 2-4, 2-5, 2-5a), the decision log, and `epic-2-boundary-gate.md:146` — all of which are historical sections where it was accurate when written, and `epic-2-boundary-gate.md:153-155` explicitly records the move.
- **Impact**: Minor rather than Nit because this file is **matrix-tagged**, so it is read by the gate operator at exactly the moment the Thai sign-off is being handled — the one reader most likely to bind the sign-off to the wrong bytes. It is inert to the machine and misleading to the human.
- **Suggested Resolution**: Update the comment to `6c040ef7…c6c85370`, or drop the digest from it entirely and name the sign-off by test symbol instead (position- and value-bound citations restale, as Story 2.3a already recorded).
- **Related AC**: AC13

### Finding 7: AC4's "every page in the partition is non-empty" assertion is dead code
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/multi_page_fixture_test.go:523-533`
- **Observation**:
  ```go
  if len(contentTms) != wantPerPage[p] { t.Errorf(...); continue }
  if len(contentTms) == 0 { t.Errorf("page %d carries NO content at all — ...") }
  ```
  `wantPerPage := []int{mpLinesOnPage1, mpLinesOnPage2}` = `{22, 7}`. Reaching the second branch requires `len(contentTms) == wantPerPage[p] == 0`, which is unreachable.
- **Impact**: AC4 states the non-emptiness assertion as an explicit, additional requirement (*"It **additionally** asserts every page in the partition is non-empty"*). As written it is unfailable at the fixture level. The property is subsumed by the count check so there is no coverage hole — but the assertion AC4 names does not exist. The unit-level twin at `paginate_test.go:183` **is** live and correct.
- **Suggested Resolution**: Drop the dead branch and note that non-emptiness is carried by the count assertion, or hoist it above the count check so it can fire independently.
- **Related AC**: AC4

### Finding 8: `nonStreamRegions` is a second, independent copy of `assertStreamLengthsAreExact`'s walk, with an undeclared lockstep requirement
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/render_test.go:665-698` vs `:790-835`
- **Observation**: The two functions walk `/Length N` → `"stream\n"` → body → `endstream` with identical logic, except that `assertStreamLengthsAreExact` **`t.Fatalf`s** on each malformed condition where `nonStreamRegions` silently **`break`s**. The soundness argument depends entirely on (a) call order inside `assertWellFormedPDF` and (b) the two walks staying byte-for-byte equivalent. Neither is asserted. If a future edit relaxes one — say adds `stream\r\n` support or widens the 40-byte window — `nonStreamRegions` breaks early and the **entire remainder of the file, font programs included, becomes a "non-stream region"**.
- **Impact**: Forward-looking only. I measured both halves: region/body separation is **exact today** (0 bytes of stream body leak into a scanned region on any of the eight goldens, the `/Length1` font shape included), and even a *total* scoping failure would produce **0** false positives on today's corpus. So this is **not** a D-000.15 false-positive generator and the guard should not be dropped — but the coupling is invisible to the next editor.
- **Suggested Resolution**: Have `assertStreamLengthsAreExact` return the body ranges it validated and let `assertIndirectRefsAreDelimited` consume them, so one walk serves both and divergence is impossible. Failing that, state the coupling in both docblocks.
- **Related AC**: AC6 / the §6 remedy (layer 2)

### Finding 9: Three residual miscounts in the record, two of them inside the paragraph that diagnoses miscounting
- **Severity**: Minor
- **Category**: Convention (D-000.47, D-000.49)
- **Location**: story §7 "The call-site enumeration, and a counting mistake I made twice" (lines 901, 815) and §7 AC5 row (line 889)
- **Observation**: Re-derived from the artifact:

  | claim | measured | how |
  |---|---|---|
  | "20 call sites and one definition — **21 occurrences** under `folio-go/` excluding `testdata/`" | **27** occurrences (20 calls + 1 definition + **6 comment mentions**) | AST-free scan of every non-`testdata` `.go` file |
  | "inside `assertWellFormedPDF` so it covers all **19 call sites**" | **20** (pre-story tree at HEAD has **18**) | same scan, plus `git show HEAD:` over the pre-story tree |
  | "a **2,318-identifier** sweep against 18 banned substrings" | the test itself logs **2348** | `passtwo_arch_test.go:185` under `-v` |

  The 20-call-site figure and the per-file table are **correct**, as is the four-declared-digest-sites figure for `multi-page` (3 `sites` entries + the artifact).
- **Impact**: The story's own root-cause statement — *"I counted what I intended to add, not what the artifact contains"* — applies to three further figures it did not re-derive, including the one in the sentence announcing the fix. Under D-000.47 the enumeration **is** the mechanism, so a stated count taken from intent remains the defect the mechanism exists to prevent.
- **Suggested Resolution**: Re-derive all three from the artifact and correct them. The 2,318 figure in particular is trivially checkable — the test prints it.
- **Related AC**: AC13

### Finding 10: `paginateDocument`'s line-grouping loop has no progress guarantee
- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/render.go:825-850`
- **Observation**: If `runs[i].band` is ever a value other than `pageHeaderBandIndex`, `pageFooterBandIndex` or `contentBandIndex`, the inner loop's `runs[j].band == contentBandIndex` fails at `j == i`, so `j` stays `i`, `i = j - 1` decrements, and the outer `i++` restores it — an infinite loop appending an empty `ColumnItem` every iteration until OOM.
- **Impact**: Unreachable today because `documentBands` (`render.go:181-192`) returns exactly three bands. But the loop's termination rests on an unasserted enumeration invariant in another function rather than on its own structure, and Epic 4's tables are the kind of change that adds a band.
- **Suggested Resolution**: Treat `j == i` as an explicit internal error rather than relying on the band enumeration.
- **Related AC**: AC2

### Finding 11: `goldenDigestSearchScope` cannot see a live digest site outside `fixtures/`
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/byte_neutrality_test.go:384`; the undeclared occurrence at `_bmad-output/implementation-artifacts/epic-2-boundary-gate.md:291`
- **Observation**: The completeness half asserts *"the set of files inside the search scope that CONTAIN a declared digest must equal the set declared"* — a genuinely good mechanism. Its scope is `{"fixtures", "folio-go/byte_neutrality_test.go"}`, justified because `_bmad-output/` story files *"record digests as PAST-TENSE MEASUREMENTS … history and must NOT be rewritten."* That rationale is correct for story files. It does **not** cover `epic-2-boundary-gate.md`, which quotes `multi-page`'s digest at `:291` as a **live** gate binding, not as history.
- **Impact**: A fifth site exists that the completeness guard was built to make impossible, and it is in the document the gate operator reads. Small today (the gate doc is append-only and dated, so a re-record would append rather than stale a live line) — recorded because the guard's log message claims *"the search scope carries no undeclared occurrence"*, which is true only of the scope, not of the repository.
- **Suggested Resolution**: Either declare a `"gate-doc"` site kind for the gate document, or state in the scope's comment that live non-`fixtures/` records exist and are deliberately out of scope.
- **Related AC**: AC7

### Finding 12: `splitPageContentStreams` scans font-program bytes for an ASCII operator
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/multi_page_fixture_test.go:311-320`
- **Observation**: The helper splits on `">>\nstream\n"` across the **entire** document and keeps any body containing `"BT\n"`. The `multi-page` golden's four streams include a **56,692-byte `FontFile2`**; it happens not to contain the bytes `42 54 0A`.
- **Impact**: A future subset that does contain that byte triple produces a phantom "page", and `len(streams) != mpTotalPages` fires in AC3, AC4 and AC8 for a reason unrelated to pagination. Four tests depend on the helper. Same class as Finding 8's hazard, but here the guard against it is luck rather than scoping.
- **Suggested Resolution**: Select content streams by following the page objects' `/Contents` references rather than by scanning binary bytes for a text operator.
- **Related AC**: AC3, AC4, AC8

### Finding 13: AC2's two-page rows do not discriminate the ruled boundary function from the rejected one
- **Severity**: Minor
- **Category**: Tests (D-000.33)
- **Location**: `folio-go/internal/layout/paginate_test.go:109`, `:115`
- **Observation**: Both 2-page rows place the offending item's **top exactly on the window edge** (`stackedLines(11, 10000)` → line 10 spans `118000..128000` with the window ending at `118000`; and `Top: 118000, Bottom: 118001`). The rejected model — *"a line belongs to whichever page contains its top"* — yields 2 pages for both rows as well.
- **Impact**: The declared table alone cannot tell *"contains it entirely"* from *"contains its top"*, which is the distinction D-2.6.1's amendment exists to fix. The discriminating case **does** exist (`TestPaginateKeepsTheLeadingAcrossAWindowBoundary`, `:211`, line 3 spanning `108000..138000` with its top inside window 0, pinning `Shift == 90000`) — but it lives outside AC2's table, so the table is not the thing pinning the ruling.
- **Suggested Resolution**: Add one row with the item's top strictly *inside* the window and its bottom outside (e.g. `Top: 108001, Bottom: 118001`), putting the discrimination in the declared table itself.
- **Related AC**: AC2

### Finding 14: `PageAssignment.Shift`'s documented invariant is stronger than the code guarantees
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/internal/layout/paginate.go:213-214`
- **Observation**: The comment claims *"It is ZERO on page 0 of any document whose content fits one window."* Under the sliding-window rule, window 0 begins at the top of the **first item**, so content whose *extent* fits a window but is *positioned* below the band top (e.g. a single element at `y = 800pt`) gives page 0 a non-zero shift and pulls the element to the top of the content band. Per the ruled model this is intended behaviour; the comment is what is wrong. No test covers it.
- **Impact**: The pre-2.6 golden-stability argument leans on this invariant. It still holds, but only because the story's own measurement shows no committed fixture is within 454 pt of the boundary — a property of the corpus, not of the code, and the comment reads as though it were the latter.
- **Suggested Resolution**: Restate the invariant as *"zero when the first item's top is the content band's top"*, and add a test for the positioned-low case.
- **Related AC**: AC1, AC2

### Finding 15: `matrix.yml`'s artifact step name still says "ALL THREE documents"
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `.github/workflows/matrix.yml:160`
- **Observation**: The step is named for three documents; the list is now eight. Registration itself is **complete and correct** — `multi-page` appears at `:60`, `:88`, `:116`, `:144` (all four targets) and in the `docs=` line at `:203`, `if-no-files-found: error` is set on all eight upload steps, and the compare job uses explicit names with `-ne 4` against a literal rather than globbing a directory.
- **Impact**: A record that misstates its own scope.
- **Suggested Resolution**: Rename the step.
- **Related AC**: AC7

### Finding 16: `"] /Count 2 "` reintroduces the incidental-trailing-space pattern that Nit 26 removed
- **Severity**: Nit
- **Category**: Tests
- **Location**: `folio-go/multi_page_fixture_test.go:409-411`; `folio-go/matrix_test.go:589`
- **Observation**: `render_test.go:598-599` records Nit 26 as *"a page-count check that used to depend on an incidental trailing space rather than an actual dictionary boundary"* — and `countPageObjects` was rewritten to fix it. Both new `/Count` checks match `"] /Count 2 "`, which succeeds only because the emitter happens to write `/Count 2 >>` with a space. Additionally the needle hard-codes `2` while the failure message interpolates `mpTotalPages`, so changing the constant makes assertion and diagnostic disagree.
- **Impact**: Formatting coupling in a guard whose whole subject is a formatting defect.
- **Suggested Resolution**: Parse the integer after `/Count` and compare it to `mpTotalPages`.
- **Related AC**: AC2

### Finding 17: `formatMPForTm` produces malformed output for negative values
- **Severity**: Nit
- **Category**: Correctness
- **Location**: `folio-go/multi_page_fixture_test.go:549-561`
- **Observation**: For `v = -1500`: `whole = -1`, `frac = -500`, `itoaForTest(-500) = "-500"`, `TrimRight(…, "0") = "-5"` → `"-1.-5"`.
- **Impact**: Unreachable today — step (5) at `:414-430` separately forbids negative Y — but this is the exact number-spelling this file exists to police.
- **Suggested Resolution**: Take the absolute value of the fractional part.
- **Related AC**: AC4

### Finding 18: Several AC2/AC8 "presence preconditions" are compile-time tautologies over their own constants
- **Severity**: Nit
- **Category**: Tests (D-000.36)
- **Location**: `folio-go/multi_page_fixture_test.go:379-394`, `:396-402`; `folio-go/reserved_placeholders_test.go:87-90`
- **Observation**: The D-000.50 "the document actually overflows one content band" check computes `lowestBottom` and `windowBottom` entirely from `const`s declared in the same file (`965952 > 745890`) — decidable without rendering anything, and it does **not** confirm the document really wraps to 29 lines. Same shape for the pairwise-distinctness check over six literal constants and for `strings.Contains(<const template>, "{{page}}")`.
- **Impact**: These read as measurements of the artifact and are measurements of the file's own literals. No hole — the real evidence for 29 lines is AC4's 22/7 count assertion — but the preconditions do not add what their comments claim. Note the genuine ones for contrast: `:358-377` reads geometry back through `pageGeometryOf(tpl)`, `:442-447` reads `tpl.doc.Bands.*.Elements` (so **AC3's presence precondition is real and can fail**), and `reserved_placeholders_test.go:98`'s `len(streams) < 2` reads the render.
- **Suggested Resolution**: Derive the overflow precondition from the rendered/laid-out document rather than from constants, or relabel these as declarations rather than preconditions.
- **Related AC**: AC2, AC8

### Finding 19: "two inherited failures" — verified, and NOT a live claim
- **Severity**: Nit
- **Category**: Convention
- **Location**: `_bmad-output/implementation-artifacts/2-3-shape-latin-thai-and-cjk-text.md:723`, `:1700`, `:1995`; `2-3a-audit-the-vendor-boundary.md:783`, `:882`, `:1335`, `:1530`
- **Observation**: Swept for it as asked. Every surviving occurrence is in a **closed, append-only story record** where it was accurate at that story's close (`TestP2IndependentDPCrossCheck` was still red then; Story 2.4 fixed it per D-2.1.9). **No live document asserts it** — not this story, not `epic-2-boundary-gate.md`, not `deferred-work.md`, not the decision log. This story's own baseline table correctly states one permitted failure.
- **Impact**: None. Recorded to close the question rather than to manufacture a finding: D-1.6.6 forbids editing closed story text in place, so these are correct as they stand.
- **Suggested Resolution**: None. Do not edit the closed records.
- **Related AC**: AC12

---

### AC-by-AC disposition

| AC | disposition |
|---|---|
| **AC1** | **satisfied.** Both pending-sign-off artifacts re-hashed by me and **unmoved**. All seven pre-2.6 goldens byte-unchanged; `git diff --stat -- fixtures/` lists only the seven provenance appends, `git status` only `?? fixtures/multi-page/`. Both-directions statement holds. |
| **AC2** | **Finding 5** (composition asserted at one row), Finding 13, Finding 10. The declared table itself is well built — hand-written literals, zero-content and one-millipoint-over rows both present, nothing recomputed from `Paginate`. |
| **AC3** | **satisfied.** Hand-derivations independently reproduced (798.269 / 51.448). Presence precondition is real and can fail. Note: header/footer Y is appended unshifted at `render.go:880-894`, so "identical Y" and "present on every page" are one fact, not two. |
| **AC4** | **Finding 7.** Partition 22/7 independently re-derived and correct; the boundary index is asserted, not the sum. |
| **AC5** | **satisfied and red-proved by me** — the identifier sweep fired on an injected `columnWindowProbe`; 2348 identifiers over 6 files against 18 banned substrings. The rank-table limitation argument checks out. |
| **AC6** | **Finding 4**, Finding 12. Fixture is complete (4 files, byte-identical in-repo constant, README states what it can and cannot express, all-Latin so no new sign-off). |
| **AC7** | **NOT met — Finding 1.** Registration is complete and correct; the guard is not. |
| **AC8** | **satisfied.** Recovered from drawn glyphs via `/ToUnicode` on every page, with a real N≥2 precondition and a negative half. Finding 18 is cosmetic. |
| **AC9** | **satisfied.** Five obligations declared; `declaredEpic2GateObligations` diff is exactly one line, verified. |
| **AC10** | **satisfied.** Stats line byte-identical, re-measured. DW-11 answered in writing. |
| **AC11** | **satisfied.** Two modules; DW-14 = **45**, independently reproduced from the artifact and reported both directions. |
| **AC12** | **satisfied.** Every figure reproduced exactly. |
| **AC13** | **Finding 9**, Finding 6. Also: AC2's own text still cites `TestContentHeightDependsOnGeometryAlone` as living in `internal/bandcomposition_arch_test.go`; it is at `internal/layout/band_test.go:104`. Disclosed in §8 but not corrected in AC2. |

### Closing note on the framing question

The remedy is **substantially real and I could not falsify its core**. Layers 1–3 are three genuinely
different properties, the red-proofs reproduce under my own hand, `assertWellFormedPDF` was
parameterised rather than loosened, the external oracle is recording-time-only and covers all eight,
and the scoping the developer was most worried about measures exact.

What the review found is that the story's *own* diagnosis — **"what shape does this guard encode as an
invariant, and what future artifact will therefore skip it?"** — was not applied to the guard it wrote
last. `readEmittedRuns` encodes "one content stream per document"; the matrix guard inherited it; and
the honesty check chosen for the deferred legs (`go vet -tags matrix`) is a **compile** check, which is
structurally incapable of seeing a runtime fatal. One native invocation finds it. The same shape
recurs in Finding 3 (a construction fix applied to one of the three sites its own comment enumerates)
and Finding 4 (an orphan check that counts where it must compare sets).

---

## Finisher's Delivery Log

*Written by the story finisher, after every event it asserts (D-000.28). Every command was run from
`folio-go/` (or `lint/`) directly against the toolchain (`/opt/homebrew/bin/go`, `GOWORK=off` — this
worktree's `go.work` otherwise resolves back into the parent tree) — `rtk proxy`'s own filtering was
avoided for the exact commands D-000.12/D-000.26 require a raw stream for; `gofmt -l` was run bare, per
the story's own measurement-instrument warning.*

### 0. The Blocker, reproduced, fixed, and proven — at the mandated invocation

**Reproduced exactly as reported**, before any fix:

```
$ FOLIO_MATRIX_TARGET=darwin/arm64 CGO_ENABLED=0 GOWORK=off go test -tags=matrix -count=1 -run TestTargetRenderHash -v .
matrix_test.go:1233: darwin/arm64: the multi-page leg carries 1 page-header run(s) and 1
  page-footer run(s) at their declared Y; a two-page document must carry exactly 2 of each
--- FAIL: TestTargetRenderHash
```

**Root cause, confirmed by reading**: `shaped_fixture_test.go`'s `readEmittedRuns` selects the FIRST
text-bearing content stream out of `pdfObjects`' map and stops. `fixtures/multi-page/expected.pdf` has
two content streams, one per page; the guard read exactly one page's worth of runs.

**Fixed narrowly, per a mid-task scope directive** (a Story 2.6a has been ruled in to sweep the shared
PDF-reading helpers generally; this story's fix does not pre-empt it). `readEmittedRuns` and its other
ELEVEN call sites — `shaped_fixture_test.go` (5), `matrix_test.go` (3, excluding the one this fix
touches), `wrapped_text_fixture_test.go` (1), `segment_origin_test.go` (1), `first_baseline_acceptance_test.go`
(1) — are UNCHANGED: every one of them renders a single-page document today, so "first stream" and "the
only stream" are the same thing there, and none of them derives discriminating power FROM reading only
the first stream (checked: none asserts anything about a second page, because none of their subjects has
one). `readEmittedRuns`' per-block parser was extracted, byte-for-byte unchanged in behaviour, into
`parseContentStreamRuns` (`shaped_fixture_test.go`) so a new, narrowly-scoped sibling —
`readEmittedRunsAllPages` (`matrix_test.go`), used ONLY by `requireMultiPageIsGenuinelyMultiPage` — could
reuse it instead of a second ~70-line copy of the CID/adjustment parser. **Handoff to Story 2.6a,
recorded explicitly**: `readEmittedRuns`' eleven other call sites carry the identical single-page
assumption `readEmittedRuns`' own docblock now names, and are 2.6a's to sweep.

**The compounding half — Finding 2 — fixed in the same change**, because the orchestrator's own framing
made it one fix, not two: `requireMultiPageIsGenuinelyMultiPage`'s own checks (`bytes.Count("<< /Type
/Page /Parent ")`, `bytes.Contains("] /Count 2 ")`) measured PRESENCE and would have passed on the exact
broken bytes this story exists to prevent. A new `requirePageTreeResolves` (`matrix_test.go`) resolves
`/Kids` into the SET of object numbers it references and requires that set to equal both the declared
`/Count` and the set of objects actually defined as `/Type /Page` — REACHABILITY, applied to the FRESH
per-target render (a different subject from `golden_structural_validity_test.go`'s committed-bytes
oracle, per D-2.6.8's layer/subject distinction).

**Re-run at the mandated invocation, after the fix:**

```
$ FOLIO_MATRIX_TARGET=darwin/arm64 CGO_ENABLED=0 GOWORK=off go test -tags=matrix -count=1 -run TestTargetRenderHash -v .
    matrix_test.go:1382: target darwin/arm64 (minimal-rect (fontless)): sha256=0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c (547 bytes; ...)
    matrix_test.go:1382: target darwin/arm64 (font-text (template+font)): sha256=a69a665331e7f0d31619f48179b54c7b9cb7a90ae013ed9c7c79daa128612181 (22315 bytes; ...)
    matrix_test.go:1382: target darwin/arm64 (image-embed (template+image)): sha256=e5778eb872c98ec4a3c3c89466a8313cf52931b896701de8a43f3506abe689fc (995 bytes; ...)
    matrix_test.go:1382: target darwin/arm64 (multi-script-fallback (shipped faces, fallback chain)): sha256=4699c8d710724ea544cc26bb3ee2b96af7a333f3dddd4462c0c846f7790480b0 (55086 bytes; ...)
    matrix_test.go:1382: target darwin/arm64 (shaped-text (shaping, all three scripts)): sha256=6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370 (91059 bytes; ...)
    matrix_test.go:1382: target darwin/arm64 (wrapped-text (line breaking, all three scripts)): sha256=277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5 (72738 bytes; ...)
    matrix_test.go:1382: target darwin/arm64 (three-band-page (band composition, all three bands populated)): sha256=746efcbcfb5be30a06caaaefae25e3eaba1962c3fa47a74da10af6d0885372bf (54452 bytes; ...)
    matrix_test.go:1382: target darwin/arm64 (multi-page (pagination, running header and footer on every page)): sha256=66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b (66525 bytes; ...)
--- PASS: TestTargetRenderHash (0.93s)
PASS
```

All eight documents hashed and matched their recorded goldens on this native leg, including
`shaped-text` (`6c040ef7…c6c85370`, unmoved — the pending Thai reading sign-off's subject) and
`multi-page` (`66ce0ee4…5869b`, unmoved).

**Red-proof (D-000.30, D-000.40 — a mutation asserts it applied by a non-empty `diff`, never by an exit
code).** Reverted `internal/pdf/document.go`'s `appendRefArray` to the pre-2.6.7 hand-rolled loop (no
separator between references); `diff` against the pre-mutation copy was non-empty, confirming the
mutation applied. Re-ran the SAME invocation:

```
matrix_test.go:1351: darwin/arm64 (multi-page ...): the indirect reference "8 0 R" is followed by "1",
  which is neither white space nor a PDF delimiter.
matrix_test.go:1372: darwin/arm64: /Kids array "8 0 R10 0 R" does not tokenize into whole "N 0 R"
  triples — a run-together reference ... collapses two tokens into one unparseable token ...
--- FAIL: TestTargetRenderHash (1.15s)
```

Both the pre-existing `assertIndirectRefsAreDelimited` AND the new `requirePageTreeResolves` fired on
the exact broken bytes; the other seven legs stayed green. Restored by hand; `diff` against the saved
pre-mutation copy is empty; the passing re-run quoted above is that restored state, re-run once more
after restoration to confirm.

**What the native leg proves, and what it does NOT** (a ruling landed mid-finish and governs this
paragraph, not a choice made here). Registering a `matrixDocuments` entry now requires running that
document's leg on the HOST target once, before `review` — a sequencing fix (D-000.43's shape), not a
D-000.4 cadence override: the leg is a local `go test`, no Docker, nearly free, and the expensive thing
D-000.4 protects is the Docker arm64 boot. **It proves the leg EXECUTES and produces a hash on ONE
target** (`darwin/arm64`, this machine). **It proves NOTHING about cross-target agreement** — that
remains the Epic 2 gate's job; `multi-page`'s four-target matrix legs stay DEFERRED to it, unchanged
from the original story text (*Heavy-test cadence — proposed DECLINED*, above).

**`go vet -tags matrix` is demoted to a compile gate ONLY, per the same ruling.** It stays useful — it
is the only thing that can check a leg this host cannot run at all — but it is never again cited here as
assurance that a deferred leg WORKS. That gap (compiles ≠ runs correctly) is exactly where this Blocker
lived. Accordingly *§3* below states, separately, what was EXECUTED and what was NOT, by name.

### 1. Findings triage

| # | Sev | Decision | Rationale |
|---|---|---|---|
| **1** | Blocker | **FIX** | Root cause confirmed by reading and reproduced verbatim at the mandated invocation. Fixed via a narrowly-scoped sibling (`readEmittedRunsAllPages`), not by changing `readEmittedRuns`' contract for its other 11 callers. Red-proofed; see §0. |
| **2** | Major | **FIX** (bundled with 1, per the orchestrator's own framing) | `requireMultiPageIsGenuinelyMultiPage`'s presence-only checks would have passed on the exact broken bytes. `requirePageTreeResolves` adds genuine reachability (set equality between `/Kids` references and defined `/Type /Page` objects). Red-proofed on the same mutation as Finding 1; see §0. |
| **3** | Major | **FIX** | The "no other ref array exists yet" premise was measurably false (`/DescendantFonts` is live in 6/8 goldens). `document.go:81-83`'s page-tree literal and `textdoc.go`'s `/DescendantFonts` array are now routed through `appendRefArray`/`writeRefArray` (byte-identical for their always-single-element case — confirmed: all 8 goldens still `qpdf --check` clean at identical digests). The false premise sentences in `appendRefArray`'s docblock (`document.go`) and `assertIndirectRefsAreDelimited`'s docblock (`render_test.go`) are corrected to the narrower true claim: `/Kids` is the only ref array that is ever MULTI-element today, not the only ref array. |
| **4** | Major | **FIX** | `golden_structural_validity_test.go`'s orphan check compared COUNTS; a duplicated reference plus an orphaned page (`[8 0 R 8 0 R]` with objects 8 and 10 both defined) passed. Rewritten to compare SETS and to require referenced object numbers to be pairwise distinct. Red-proofed directly against the finding's own probe bytes (both the duplicate-reference and the orphan branch fired); reverted; the real 8-golden suite still passes. |
| **5** | Major | **FIX** | AC2's composition (geometry + content extent, through to bytes) was verified end-to-end at exactly N=2. Added `multi_page_composition_test.go`: a declared table rendering N=1 (content that fits), N=3 (spanning three windows) and N=0 (empty content band) through the public `folio.Render`, asserting `/Count` and `/Type /Page` object count at two independent sites on the produced bytes. Sentence counts were MEASURED (word-wrap depends on word boundaries, not just the per-line-advance arithmetic), chosen with headroom on both sides of their nearest boundary so an unrelated wrapping change cannot flip a row by accident. Red-proofed by mutating one row's declared value and confirming both assertion sites fail with the correct diagnostic; reverted. |
| **6** | Minor | **FIX** | `expected_breaks_signoff_matrix_test.go`'s comment named a superseded digest and a file that has never existed. Corrected to name the sign-off by pointer (its own artifact/provenance) rather than by a second copy of a value that restales the same way — the reviewer's own suggested alternative, chosen because a literal digest in a comment is exactly what went stale here. |
| **7** | Minor | **FIX** | The `len(contentTms) == 0` branch in `TestMultiPageLineToPagePartitionIsPinnedByValue` was unreachable dead code (`wantPerPage`'s declared values, 22 and 7, are never zero, and the presence precondition above already forbids a zero-valued declared partition). Removed, with a comment explaining that non-emptiness is carried by the count check immediately above rather than by a separate branch. |
| **8** | Minor | **FIX (declared, not restructured)** | Per the orchestrator's explicit instruction, the delimiter guard (`assertIndirectRefsAreDelimited`/`nonStreamRegions`) is KEPT — the reviewer's own independent replication measured 0 bytes of stream-body leakage across all 8 goldens including the `/Length1` font shape, and 0 false positives even under total scoping failure. The undeclared lockstep coupling between `nonStreamRegions` and `assertStreamLengthsAreExact` is now declared in BOTH docblocks (`render_test.go`), naming the hazard (an edit to one walk's shape silently un-scopes the other) without changing either walk's logic. |
| **9** | Minor | **FIX** | Re-derived from the artifact, AFTER every other finisher edit (counts drift with edits to the files being counted): `assertWellFormedPDF` now has **28** total occurrences (20 calls, 1 definition, 7 comment mentions — 1 more than the reviewer's 27, from this finisher's own Finding 8 comment addition) under `folio-go/` excluding `testdata/`; the identifier sweep now logs **2350** (2 more than the reviewer's 2348, from the 2 new identifier references `appendRefArray`/`writeRefArray` add at Finding 3's two call sites). Both re-derivations and their invocations are in §2 below. The call-site count (20) and the four-declared-digest-sites count (4) were already correct and are unchanged. |
| **10** | Minor | **FIX** | `paginateDocument`'s line-grouping loop (`render.go`) had no progress guarantee if `documentBands`' three-band enumeration were ever violated: an explicit `if runs[i].band != contentBandIndex` check now returns a named internal error instead. Red-proofed with a direct unit call to `paginateDocument` carrying an out-of-enumeration band value: pre-fix, `go test -timeout 3s` was killed by its own timeout with a goroutine stack trapped inside the loop at the exact line the finding named; post-fix, the same call returns a clean, named error in well under a second. |
| **11** | Minor | **FIX (comment only)** | `goldenDigestSearchScope`'s docblock (`byte_neutrality_test.go`) now names the boundary-gate-document exception explicitly — that `epic-2-boundary-gate.md` carries a LIVE digest citation outside the completeness guard's declared scope, and that it is covered separately, by shape, in `TestBoundaryGateDigestsAreWellFormed`. No logic changed. |
| **12** | Minor | **FIX** | `splitPageContentStreams` (`multi_page_fixture_test.go`) used to scan the WHOLE document for `">>\nstream\n"` boundaries and keep any chunk containing the ASCII bytes `"BT\n"` — including inside font-program streams, by luck rather than scoping. Rewritten to resolve the page tree's `/Kids` array (declared order) and follow each page object's OWN `/Contents` reference to its stream — the same reachability technique used at Findings 1/2/4, applied a third time to a third subject (per-page stream selection). All 4 dependent tests (AC3, AC4, AC8, plus the composition test) still pass. |
| **13** | Minor | **FIX** | Neither existing 2-page row in `paginate_test.go`'s AC2 table placed the offending item's TOP strictly inside the window with its BOTTOM outside — both rows the rejected "contains its top" model would also get right. Added the exact row the finding suggested (`Top: 108001, Bottom: 118001`); it passes under the ruled "contains it entirely" model and would fail under the rejected one. |
| **14** | Minor | **DISMISSED** — the finding's factual premise is false, verified empirically | The finding claims window 0 begins at "the top of the first item." This file's own package doc (rule 1, `paginate.go`, unchanged by this story) states window 0 begins at the CONTENT BAND'S OWN top, unconditionally; only window N+1 begins at the first item that did not fit in window N. Verified directly: `Paginate` on a single item positioned 80,000mp below the content band's top, comfortably inside one window, produces `Shift == 0` on page 0 (not the non-zero value the finding predicts) — the item renders exactly where it was declared, not pulled to the band's top. A permanent regression test, `TestPaginateShiftIsZeroOnPageZeroEvenForContentPositionedLow`, pins this and is confirmed to fail under the finding's (wrong) prediction. The original docblock is unchanged; a note records that it was checked, not merely trusted. |
| **15** | Nit | **FIX** | `.github/workflows/matrix.yml`'s artifact-comparison step, renamed from "for ALL THREE documents" to "for every registered document" (the list is 8, and the step's own name was the only thing out of date — registration itself was already complete and correct). |
| **16** | Nit | **FIX (both sites)** | Both cited `] /Count N ` substring matches — `matrix_test.go` (fixed as part of Finding 2's `requirePageTreeResolves`, which parses the integer via regex) and `multi_page_fixture_test.go` — now parse the integer (via the new shared `readDeclaredCount`, `multi_page_composition_test.go`) instead of matching a trailing-space-dependent literal. |
| **17** | Nit | **FIX** | `formatMPForTm` (`multi_page_fixture_test.go`) produced `"-1.-5"` for `v = -1500` (unreachable today, since step 5 of the same file separately forbids negative Y, but exactly the number-spelling this file exists to police). Fixed by taking the fractional part's absolute value before formatting. |
| **18** | Nit | **DISMISSED** | The reviewer's own text: *"No hole ... but the preconditions do not add what their comments claim."* A cosmetic relabeling of already-correct, non-misleading-in-practice preconditions, for a Nit-severity finding with an explicit "no hole" from the reviewer. Relabeling several presence checks across two files would be pure churn against the story's remaining scope; left as is. |
| **19** | Nit | **ACKNOWLEDGED, no action** | The reviewer's own suggested resolution is "None," and D-1.6.6 forbids editing the closed story records the finding names. Confirmed unchanged. |

### 2. Re-derived counts (Finding 9), with their invocations — taken AFTER every other finisher edit

| claim | value | invocation |
|---|---|---|
| `assertWellFormedPDF` total occurrences under `folio-go/`, excluding `testdata/` | **28** | `grep -rn "assertWellFormedPDF" --include="*.go" . \| grep -v "/testdata/" \| wc -l` |
| `assertWellFormedPDF` call sites (`assertWellFormedPDF(t,`) | **20** (unchanged from the review) | `grep -rn "assertWellFormedPDF(t," --include="*.go" . \| grep -v "/testdata/" \| wc -l` |
| the identifier sweep | **2350** | `go test -run TestInternalPDFReachesNoLayoutComputation -count=1 -v ./internal/` → `passtwo_arch_test.go:185` |

The +1 (27→28) and +2 (2348→2350) deltas from the review's own re-derivation are both attributable to
this finisher's own edits (a new comment mention at Finding 8's fix; two new identifier references at
Finding 3's `appendRefArray`/`writeRefArray` call sites) — re-derived rather than assumed, per the
orchestrator's explicit instruction not to adjust a number to match either party's prior figure.

### 3. The gate table, re-measured a second time — EXECUTED vs NOT RUN named separately

| scope | invocation (verbatim) | counting rule | at review | at finisher close |
|---|---|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | all occurrences | 551 · 1 | **557 · 1** |
| `folio-go/` | the same | top-level only | 336 · 1 | **338 · 1** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | all occurrences | 85 · 0 | **85 · 0** |
| `lint/` | the same | top-level only | 47 · 0 | **47 · 0** |
| `folio-go/` | `GOWORK=off go list -m all` | — | two | **two** |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go build ./...` | — | clean | **clean** |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | — | clean | **clean** |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet -tags matrix ./...` | — | clean | **clean (compile gate only — see §0)** |
| `folio-go/` | `gofmt -l .` (bare) | — | empty | **empty** |
| `folio-go/` (matrix) | `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .` | — | **FAIL (Blocker 1)** | **PASS, 8/8 hashes match — see §0** |

The +6 all-occurrence / +2 top-level delta from the review's own numbers is exactly the finisher's new
tests: `TestMultiPageComposedPageCountFromADeclaredTable` (1 top-level + 3 subtests), the new discriminating
row inside the pre-existing `TestPaginatePageCountFromADeclaredTable` (1 subtest, 0 new top-level), and
`TestPaginateShiftIsZeroOnPageZeroEvenForContentPositionedLow` (1 top-level) — 6 all-occurrence, 2
top-level, arithmetic confirmed. **The single permitted failure remains AC10's**, stats line
byte-identical: `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.

**EXECUTED, natively, by the finisher**: every row above, plus `qpdf --check` and `qpdf --show-npages`
on all 8 goldens (all exit 0; `multi-page` = 2 pages, seven = 1), plus `shasum -a 256` on the three
sign-off-pending / must-not-move artifacts (see §4). **NOT RUN**: the four-target Docker/Node
`TestCrossTargetByteIdentity` cross-check — unchanged from the review, and still the Epic 2 gate's to
run, per D-000.4's decline.

### 4. AC1's sign-off-pending and must-not-move artifacts, re-verified at close

| artifact | digest | status |
|---|---|---|
| `fixtures/shaped-text/expected.pdf` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` | **UNMOVED** |
| `fixtures/expected-breaks/expected_breaks.json` | `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de` | **UNMOVED** |
| `fixtures/multi-page/expected.pdf` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` (66,525 bytes) | **UNMOVED** |

`git diff --stat -- fixtures/` lists only the seven pre-existing fixture READMEs (provenance-only) and
the new `fixtures/multi-page/` directory. No `expected.pdf` or `expected_breaks.json` is modified.

### 5. AC7's disposition, corrected

**AC7 — met.** Was **NOT met** at review (Finding 1). `requireMultiPageIsGenuinelyMultiPage` now reads
every text-bearing content stream (via `readEmittedRunsAllPages`) and resolves the page tree's
reachability (via `requirePageTreeResolves`) before its content assertions run; the native leg passes at
the mandated invocation (§0); registration itself was already complete and was never in question.

**The Epic 2 gate still owes exactly five obligations** (D-2.6.6). Nothing in this finisher's fixes adds
a sixth: `TestEpic2GateObligationsMatchTheDeclaredSet` passes unchanged, `declaredEpic2GateObligations`'
diff against the pre-story tree is still exactly one line, and `epic-2: backlog` is unchanged (the
gate's to flip). The two Findings this finisher fixed inside `golden_structural_validity_test.go`
(Finding 4) and the matrix harness (Findings 1, 2, 12) are test-side corrections to how existing
obligations are CHECKED, not new obligations.

### 6. Files changed by the finisher

**New** — `folio-go/multi_page_composition_test.go` (Finding 5: end-to-end AC2 composition table).

**Modified** — `folio-go/shaped_fixture_test.go` (`readEmittedRuns` extracted into `parseContentStreamRuns`,
behaviour unchanged); `folio-go/matrix_test.go` (`readEmittedRunsAllPages`, `requirePageTreeResolves`,
`requireMultiPageIsGenuinelyMultiPage` rewired — Findings 1, 2, 16); `folio-go/golden_structural_validity_test.go`
(Finding 4: set-based orphan/duplicate check); `folio-go/internal/pdf/document.go` (Finding 3: page-tree
literal routed through `appendRefArray`; corrected premise comment); `folio-go/internal/pdf/textdoc.go`
(Finding 3: `/DescendantFonts` routed through `writeRefArray`); `folio-go/render_test.go` (Finding 3:
corrected premise comment; Finding 8: declared coupling in both docblocks); `folio-go/render.go`
(Finding 10: explicit internal error instead of an unguarded loop); `folio-go/internal/layout/paginate_test.go`
(Finding 13: discriminating table row; Finding 14: `TestPaginateShiftIsZeroOnPageZeroEvenForContentPositionedLow`);
`folio-go/internal/layout/paginate.go` (Finding 14: docblock left as written, with a note that it was
checked); `folio-go/multi_page_fixture_test.go` (Finding 7: dead branch removed; Finding 12:
`splitPageContentStreams` rewritten to follow `/Contents` references; Finding 16: `/Count` parsed, not
matched; Finding 17: `formatMPForTm` negative-fraction fix); `folio-go/byte_neutrality_test.go` (Finding
11: scope-exception comment); `folio-go/expected_breaks_signoff_matrix_test.go` (Finding 6: corrected
comment); `.github/workflows/matrix.yml` (Finding 15: step renamed).

### 7. Follow-ups (none deferred as new obligations)

No finding was DEFERRED as a follow-up ticket: every FIX was completed and validated in this pass, and
both DISMISSED findings (14, 18) were dismissed on their merits (a false factual premise, and the
reviewer's own "no hole" admission) rather than deferred for later. Story 2.6a (already ruled in,
outside this story's authority to schedule) inherits the sweep of `readEmittedRuns`' other eleven call
sites, as recorded in §0 and in that function's own docblock.

