---
baseline_commit: 96a313b
---

# Story 2.7 — Render `Page X of Y`

**Epic**: 2 — Text, shaping and page composition
**Sprint status key**: `2-7-render-page-x-of-y`
**Status**: `done`
**Covers**: FR31 · AD-4 (`_bmad-output/planning-artifacts/epics.md:904-929`)
**Baseline commit**: `96a313b` — *"Story 2.6a: Sweep single-page assumptions from the shared PDF readers (finisher)"*
**Predecessor**: Story 2.6a, committed at `96a313b`

> **THIS STORY HAS ONE BLOCKING DECISION AND ONE CORRECTION TO ITS OWN BRIEF.** Both are stated
> before anything else, because AC2 cannot be written down until the first is ruled and the gate
> argument changes shape once the second is read. See **DECISIONS NEEDED** and
> **Correction — the epic has FOUR acceptance criteria, not three**.

---

## FIRST — the sign-off measurement D-000.41 requires, performed at creation

**Two of the three pinned digests have an owner sign-off pending against them. Moving either
silently invalidates a review the owner has already been asked to perform, so this is stated before
anything else.**

Re-computed at `96a313b` via `rtk proxy shasum -a 256 … > file` (D-000.12 as corrected — never
through the wrapper's pipes), not inherited from the brief:

| artifact | digest at `96a313b` | sign-off state |
|---|---|---|
| `fixtures/shaped-text/expected.pdf` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` | **Thai READING sign-off pending** (D-2.3.5) |
| `fixtures/expected-breaks/expected_breaks.json` | `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de` | **Thai BREAK sign-off pending** (D-2.4.3) |
| `fixtures/multi-page/expected.pdf` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` | recorded by Story 2.6 |

All three match the brief exactly.

**Why this story is different from 2.6a, and why the answer here is NOT structural.** Story 2.6a
edited test files only, so "no emitted byte moves" followed from its scope fence. **This story edits
the renderer.** It is therefore capable of moving every golden in the repository, and the
non-movement claim has to be *measured*, not derived. That is AC6, and it is the first thing to
measure after the implementation compiles, not the last.

**The one thing that makes the claim plausible in advance** — recorded as an expectation to be
checked, never as an assurance: **no committed fixture *document* contains `{{page}}` or
`{{pages}}`.** Measured at `96a313b`, and stated with its scope because the unscoped version of this
measurement is false:

```
rtk proxy grep -rl "{{page" --include='*.folio' --include='*.json' fixtures/ > /tmp/probe.txt
  → 0 files   (over the 7 `input.folio` documents and every `expected.json` in the 10 fixture dirs)

rtk proxy grep -rl "{{page" fixtures/ > /tmp/probe.txt
  → 1 file:  fixtures/multi-page/README.md
```

**The one hit is prose, not a document** — `multi-page`'s README quotes the placeholders while
explaining why they were kept out of its input. It renders nothing. Recorded rather than filtered
away, because "0 files" without the scope is the kind of true-sounding figure D-000.14 is about.

`fixtures/multi-page/input.folio`'s footer is the literal `FOOTER REPEATED ON EVERY PAGE`
(`folio-go/multi_page_template.go:76`), deliberately: the reserved-placeholder document was kept
**separate** from `fixtures/multi-page/` by Story 2.6 *"so that the committed golden is not
re-recorded when Story 2.7 implements these placeholders"*
(`folio-go/reserved_placeholders_test.go`, the `reservedPlaceholderFooterTemplate` comment). That
foresight is why the two pending sign-offs can survive this story. **It is a reason to expect the
digests to hold, not a substitute for measuring that they did.**

---

## Correction — the epic has FOUR acceptance criteria, not three

**The brief that commissioned this story lists three.** `epics.md:904-929` carries a fourth, and it
is the one that settles the gate question:

> **Given** documents of 1, 5, 20 and 50 pages
> **When** each renders
> **Then** `Page X of Y` is correct throughout and hashes match recorded goldens **on all four
> targets**

Read verbatim at `96a313b`, `epics.md:927-929`. This is not a detail: *"hashes match recorded
goldens on all four targets"* is a direct instruction to register a matrix document, and the brief's
framing (*"if this story registers a new `matrixDocuments` entry, the gate **may** legitimately
become six"*) is more tentative than the spec it is drawn from. It is carried below as **AC5**.

**Nothing else in the brief is contradicted.** The three ACs it names are `epics.md:914-925`
verbatim and are carried as AC1–AC3.

**Correction to this section's own citations (finisher, this story's review, Finding 6).** Every
`epics.md:NNN-NNN` line range in this document, and three in shipped source comments, were wrong —
including this section's own, and including the orchestrator's brief's. Re-measured by `grep -n`
against `epics.md` at the reviewed tree and corrected throughout: Story 2.7's section is `904-929`
(not `904-926`); AC1 is `914-916`; AC2 is `918-921`; AC3 is `923-925`; AC4 (this section's own
rescued fourth criterion) is `927-929` (not `923-926`, which both the developer and the
orchestrator's brief cited). The three code-comment citations were re-cited by AC ordinal rather
than line range, since a position-bound citation has already forced this project to re-cite by
name once (Finding 7, same review).

---

## Baseline, measured at creation

HEAD is **`96a313b`** on `main`. Working tree **clean** — `rtk proxy git status --porcelain`
produced no output, checked immediately before and after the probes below, and every probe file was
deleted. **No counted file had an uncommitted change** (D-2.6.9 correction).

**Full ordinary suite** — `rtk proxy go test ./... -count=1 -v` from `folio-go/`, output captured to
a file:

| measure | value |
|---|---|
| all-occurrences PASS (`--- PASS`) | **564** |
| all-occurrences FAIL (`--- FAIL`) | **1** |
| top-level PASS (`^--- PASS`) | **345** |
| top-level FAIL (`^--- FAIL`) | **1** |

**Exactly one expected failure**, and it is not this story's: `TestCorpusMeetsP6ExerciseFloors`, the
intentional Story 2.1 red, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — reproduced
character-for-character, not inherited.

**Scope note on the numbers (D-000.14, extended).** These are scoped figures: *whole module, no
build tags, `-count=1`, all-occurrences and top-level reported separately*. Any figure this story
reports later must carry the same three-part scope or it is not comparable (see
`[[carried-gate-figures-are-scoped]]` — a 56-vs-66 style mismatch is almost always two scopes, not a
regression).

---

## In plain terms (read this first if you just want the gist)

A printed statement should let whoever is holding it check that nothing has gone missing. Paper has
always done that with a line at the foot of each sheet saying which sheet this is and how many there
are in total. This story builds that: page numbering now works, on documents of one, five, twenty
and fifty pages, including the awkward case where the page count crosses from one digit to two
digits partway through the document.

Getting there needed the engine to reserve space for a number before it knows what that number will
be, since the total page count is not known until layout finishes. That design question was settled
before work started, so the hard part turned out to be closing what it opened up, not deciding it.

Before this was ready to ship, two mistakes were caught and fixed. The first: a footer that
mentioned the page number twice in one line — an unusual but perfectly reasonable thing for someone
to write — printed the right number once and a meaningless placeholder digit the second time, on
every single page, with nothing to say anything had gone wrong. That is now fixed and proven on a
real document that does exactly this. The second: a safeguard meant to guarantee that no future
change could quietly let page numbers leak into places they do not belong looked convincing but,
against one particular shape of mistake, would not actually have caught it. A second, independent
safeguard now closes that gap, checked against the exact shape of mistake that slipped past the
first one.

What will look wrong later and is not: only one of the four page-count documents has its exact bytes
pinned and checked across machines; the other three are checked for correct wording but not for
identical bytes. That is a deliberate, argued choice, not an oversight, and it is written down as
such rather than claimed as coverage it does not have.

---

## Story

As a template author,
I want a footer that says which page of how many the reader is holding,
So that a printed statement is verifiably complete.

---

## Do not re-open — settled rulings this story inherits

Apply these; do not re-litigate any of them.

| ruling | what it binds here |
|---|---|
| **AD-4** (`ARCHITECTURE-SPINE.md:151-166`) | Pass one produces a complete `PageModel`; pass two serializes it. Page-dependent text is a **late-bound slot that already holds its final measured box**, resolved **between the passes** by substituting **pre-measured glyphs**. `internal/pdf` performs no measurement, no line breaking, no pagination. **No expression may reference pagination — there is no `page` namespace and none may be added.** |
| **AD-5** | `internal/pagemodel` names only geometry, glyph runs and images. Whatever carries the slot lives there and must not name a PDF concept. |
| **AD-24** | Boxes are absolute; nothing negotiates; siblings never move. |
| **D-2.6.1 (amended)** | Sliding window; **no line is ever split**; page-relative Y = `columnY − windowStart(N)`; page count is **not** a closed form; **no page is ever empty**. |
| **D-2.6.5** | An item that fits in **no** window is a **located `Render`/`RenderTo` error** naming the element (`layout.OverflowError{ElementID, ItemHeight, ContentHeight, Kind}`). **No clip bound, no `q…W n…Q`, no public-signature change.** |
| **D-2.6.2** | A gate obligation is warranted when the artifact is **the only cross-target artifact for a shipped FR**. |
| **D-2.6.6** | Structural validity of committed bytes is **not** cross-target and is **not** a gate obligation. The gate owes five *today*. |
| **D-2.6.8** | Three layers, three properties: `writeRefArray`/`appendRefArray` by construction, token delimiter check, page-tree semantics. Ref arrays grow — *"Story 2.7 … add ref arrays"* is named in the ruling itself. |
| **D-2.6.9 + correction** | A behavioural predicate, never a grep. **Any enumeration names the commit it was taken at**, on a tree with no uncommitted changes in the counted files. |
| **D-000.53** | No golden accepted until a reader this project did not write resolves it into the objects it claims to contain — reader, version and **verbatim invocation** in the fixture's provenance. |
| **D-000.54** | A newly registered matrix document runs its **native leg once** before `review`. **Sequencing fix, NOT a D-000.4 cadence override**, and must not be logged as one. |
| **D-000.55** | `go vet -tags matrix` is a compile gate, not an honesty check. The phrase *"written, compiled and vetted, deliberately not run"* is **banned**. Name what was **executed, by target**; name the unrun legs **individually**. |
| **D-000.50** | Before writing a guard, ask **which subject can express the defect**. |
| **D-000.42 / D-000.24** | Do not count three coverages of one property; label a forward guard as one. |
| **D-000.21 / D-000.26 / D-000.28** | Assert on the artifact that carries the property; cite the subject; never write a claim before the event it asserts. |
| **D-000.30 / D-000.34 / D-000.40 / D-000.47 / D-000.48** | Do not silently disarm a detector; measure every declared site; a mechanical repair is verified mechanically. |

---

## Measured findings — read all of these before writing code

Every figure below was taken at **`96a313b`** on a **clean tree**. Probe files were written into the
module, run, and deleted; the tree was re-checked clean afterwards.

### 1. The slot mechanism is HALF-BUILT already, and the fence in `internal/bind/text.go` is deliberate

Four sites, all read:

- **`internal/bind/text.go:45-50`** — `reservedPlaceholders = {"page": true, "pages": true}`, commented
  *"AD-4's page-number slots (D-1.6.5, AC18): owned by Story 2.7, never resolved from data and never an
  error at 1.6… this reservation is checked BEFORE any data lookup is attempted, so it can never be
  shadowed."*
- **`internal/bind/text.go:136`** — the consumer. A reserved token is written back **byte-for-byte**
  (`{{` + `inner` + `}}`, preserving interior whitespace) and yields **no `Substitution`**.
- **`internal/bind/text.go`, line 76 at `96a313b`** (finisher note: shifted to line 96 by this
  story's own edits — cited by enclosing function, `BindText`'s own doc comment, from here on
  rather than by line number, this story's review, Finding 7) — **the fence.** Verbatim: *"This is deliberately NOT implemented by
  extending `reservedPlaceholders` below (AC12): page/pages are reserved whole **TOKENS**, resolved
  from neither root and owned by Story 2.7; params is a **NAMESPACE**, resolved from its own root —
  **conflating the two is how "page" would eventually acquire a namespace, which AD-4 forbids
  forever.**"*
- **`internal/bind/text.go:105-110`** — the span contract: reserved tokens *"name no data path, so
  there is nothing a document could declare about them"*.

**Read this before designing anything.** The fence is the whole of AC3's structural argument, already
written down by somebody who saw the hazard: the violation AC3 forbids is not *"someone adds a `page`
key to a map"* — it is *"`page` becomes a resolution **root**"*. That distinction is what the guard
must key on (D-000.15: key a guard on its purpose, never on a proxy).

**What the half-built mechanism does NOT yet do**: nothing anywhere carries a *measured box* for the
slot, and nothing resolves it. `bind` returns a string; measurement happens downstream in
`collectTextRuns`. So the slot type, its placement in `pagemodel`, and the between-passes resolution
step are all new.

### 2. Y is NOT circular — measured, and this is the difference between an easy story and a hard one

`internal/layout/band.go:61-77`, verbatim:

> `ContentHeight` … *"Its inputs are **PAGE GEOMETRY ONLY**. It does not receive, and cannot consult,
> the content band's elements or their measured sizes."*
>
> `ContentHeight(g) = g.Height − g.MarginTop − g.MarginBottom − g.PageHeaderHeight − g.PageFooterHeight`

And `PageGeometry` (`band.go:54-59`) carries `PageHeaderHeight` / `PageFooterHeight`, which come from
the template's **declared** `bands.pageFooter.height` (e.g. `24` in `multi_page_template.go:78`), never
from measurement.

`render.go:812-845` then routes **only content-band items** into `layout.Paginate`; page-header and
page-footer runs are collected into `header`/`footer` `BandContent` and **repeated verbatim on every
page** (`render.go:786-789`, `render.go:885-895`).

**Therefore, for a construct living in the page-header or page-footer band, the page count is
completely independent of the page-number text.** Y is simply *known* the moment `layout.Paginate`
returns. There is no circularity to solve.

**The circularity is real, and only real, if the construct is permitted in the CONTENT band** — where
the item's own line count is an input to pagination, so a wider number could re-wrap, add a line, add
a page, and change Y. That is a genuine fixed-point problem and this story must not accidentally
inherit it. **It is sub-question (c) of the decision below.**

### 3. All three shipped faces have tabular figures — measured independently of the coordinator's read

Measured through the project's **own** reader (`internal/fontset.New` → `Font.AdvanceForRune`,
`Font.UnitsPerEm`) from a throwaway `internal/fontset/zz_tabprobe_test.go`, run and deleted. This is a
**different route** from the coordinator's hand-rolled `cmap`/`hmtx` table read, which is the point of
re-deriving it.

| face | upem | advance of every digit `0`–`9` |
|---|---|---|
| `fonts/notosans/NotoSans-Regular.ttf` | 1000 | **572**, all ten |
| `fonts/notosansthai/NotoSansThai-Regular.ttf` | 1000 | **572**, all ten |
| `fonts/notosanssc/NotoSansSC-Regular.ttf` | 1000 | **555**, all ten |

**Agrees with the coordinator exactly**, by a second route.

Neighbouring literals, for the arithmetic below: in Noto Sans / Noto Sans Thai —
`P`=605, `a`=561, `g`=615, `e`=564, `o`=605, `f`=344, `SPACE`=**260**. In Noto Sans SC —
`P`=633, `a`=563, `g`=564, `e`=554, `o`=606, `f`=325, `SPACE`=**224**.

**What this settles**: the width of a rendered number depends only on **how many digits**, never on
**which** digits. `Page 19 of 20` and `Page 20 of 20` are byte-for-byte the same width.

**What it does not settle**: `Page 9 of 20` and `Page 19 of 20` are **not** the same width. Digit
*count* still varies across pages of one document, at every power-of-ten boundary.

**Do not write an AC requiring tabular figures.** It is already true of every shipped face, so such an
AC would be satisfied at birth — D-000.28's *"a claim written before the event it asserts"*, and
D-000.50's *"a guard no subject can falsify"*. If the property is to be enforced, it must be enforced
as a **forward guard against a future or caller-supplied face that lacks it** (D-000.24: labelled as
a forward guard, with no available red-proof from the shipped set), and the story must say which
subject could express the violation — a synthetic proportional-digit face is the only candidate, and
`internal/fontset` has no such fixture today.

### 4. FIGURE SPACE is NOT available across the shipped set — this kills the obvious padding option

Same probe, same route:

| rune | Noto Sans | Noto Sans Thai | Noto Sans SC |
|---|---|---|---|
| U+2007 FIGURE SPACE | **present, advance 572** (= the digit advance) | **ABSENT** | **ABSENT** |
| U+2008 PUNCTUATION SPACE | present, 268 | ABSENT | ABSENT |
| U+00A0 NBSP | present, 260 | present, 260 | present, 224 |
| U+0020 SPACE | 260 | 260 | 224 |

**Consequence, stated because it eliminates an option the brief's list implies is available.** "Pad the
short number to the reserved width with a figure space" works **only** in Noto Sans. In a chain led by
Noto Sans Thai or Noto Sans SC, U+2007 has no glyph, so `resolveRuneFace` (`render.go:476-490`, first
match over the chain by actual `cmap` coverage) would fall through to Noto Sans — introducing a
**face switch inside the run** and a **572-wide pad glyph adjacent to 555-wide SC digits**. And an
ordinary space is **not** the digit width in any face (260 vs 572; 224 vs 555), so padding with one
does not preserve the reservation either.

**There is no universally available equal-width pad glyph in the shipped set.** Any "nothing moves"
design must therefore reserve the width **positionally** — by advance arithmetic over already-measured
glyphs — not by inserting a filler character.

### 5. `align` is parsed and honoured by nothing

`internal/template/closedsets.go:29` defines `closedAligns` = `{left, center, right}`;
`internal/template/model.go:227,241` carry `Align Presence[string]` on `Column` and `Style`;
`parse_bands.go:368-377,457` decode and validate it.

**Measured at `96a313b`**: `rtk proxy grep -rn "Align" --include='*.go' .` from `folio-go/`, excluding
`_test.go` and `internal/template/`, returns **zero** hits. Nothing in `render.go`, `wrap.go`,
`internal/layout` or `internal/text` reads it.

**So today every element draws from its declared `x`, left-aligned, always.** A centred or
right-aligned page number is not expressible, which removes one whole family of consequences from the
width question — but it also means a change in digit count moves **every glyph after the slot within
the same run**, because `Page {{page}} of {{pages}}` is all-ASCII, resolves entirely to chain[0], and
is therefore **one face segment and one run**.

### 6. Digits never fall back within the shipped set

`resolveRuneFace` (`render.go:476-490`) walks the declared chain in order and returns the **first**
face both present in the `FontSet` **and** whose `cmap` actually contains the rune. All three shipped
faces cover `0`–`9` (finding 3). **Therefore a digit always resolves to `chain[0]`**, and the
572-vs-555 disagreement between faces is a **per-document constant**, not a per-page variable.

The hazard is narrow and worth naming rather than guarding: a **caller-supplied** `chain[0]` lacking
digits would push the number onto a later face while the surrounding literals stayed on the first —
a mid-run face switch. No shipped configuration can express it, so a guard on it is a **forward guard
with no available red-proof** (D-000.24) and must be labelled one if written at all.

### 7. The existing reservation tests are this story's to invert — and one half of them must survive

`folio-go/reserved_placeholders_test.go` holds two tests over one 2-page synthetic document
(`reservedPlaceholderFooterTemplate`, footer `"Page {{page}} of {{pages}}"`, `fontSize: 8`,
`width: 480`):

- `TestReservedPagePlaceholdersPassThroughOnEveryPage` — asserts the **literal** `Page {{page}} of
  {{pages}}` is drawn on every page, recovered through the document's own `/ToUnicode` CMap. **This
  story inverts it.** The file says so itself: *"at which point this document's rendering legitimately
  changes and this test legitimately becomes 2.7's to rewrite."*
- `TestReservedPlaceholderSetIsUnchanged` — pins the **set**, and carries a **negative half**: a
  non-reserved `{{notreserved}}` must still error, *"if every placeholder now passes through
  unchanged, the two assertions above are vacuous and data binding is silently disabled."*

**The finding, and it is a real hazard for this story.** Inverting the first test is correct.
**Deleting or weakening the negative half of the second is not**, and it is the easy accident: once
`{{page}}` resolves, the natural rewrite is *"assert the resolved text appears"*, which is green on a
renderer that has quietly stopped binding data at all. The negative half is the only thing standing in
that gap and it must survive the rewrite in a form that still reddens (D-000.34: a detector must not be
silently disarmed).

**A second, subtler one.** That test currently reaches the reservation *through behaviour* because
`internal/bind` is a different package. After this story, "the set is `{page, pages}`" is no longer the
property that matters — **"the set of resolution roots is `{data, params}`"** is. AC3 must move the
pin, not just keep it.

### 8. The subject population, per D-000.50 — what can express this story's defects

Measured at `96a313b`:

Scope of the population: **10 fixture directories under `fixtures/`**, of which **7 carry an
`input.folio`** (`expected-breaks`, `thai-break-corpus` and `multi-script-fallback`'s data are not
`.folio` documents in the same sense). Both denominators are given, because a single "of N" here
would be ambiguous.

| property a subject would need | fixtures having it |
|---|---|
| contains `{{page}}` or `{{pages}}` in a *document* | **0 of 7 documents / 0 of 10 dirs** |
| more than one page | **1 of 10** (`multi-page`, 2 pages) |
| a populated `pageFooter` band | 2 of 10 (`three-band-page`, `font-text`) |
| **≥ 10 pages** (needed for a digit-count change across pages) | **0 of 10** |

**No existing fixture can express this story's central defect** — a page number whose digit count
changes between page 9 and page 10. The `multi-page` fixture has 2 pages. The synthetic
`reservedPlaceholderFooterTemplate` has 2 pages. **The story must build its own subject**, and the
epic's fourth AC already names the shape: 1, 5, **20** and **50** pages — 20 and 50 both cross the
one-to-two-digit boundary, which is not a coincidence and should be stated as the reason they are in
the spec.

### 9. The shared PDF readers repaired by 2.6a — use them, and one observation

Located at `96a313b`:

| helper | file:line | shape after 2.6a |
|---|---|---|
| `readEmittedRuns` | `shaped_fixture_test.go:506` | sorts object numbers, reads **every** `Tf`-bearing stream; fatals on zero streams and zero runs |
| `resourceType0Objects` | `shaped_fixture_test.go:160` | resolves fonts from every page object |
| `assertXrefEntriesPointAtTheirObjects` | `render_test.go:879` | every subsection; fatals on a table declaring none |
| `assertWellFormedPDF` | `render_test.go:602` | takes `wantPages` (parameterised by Story 2.6) |
| `splitPageContentStreams` | `multi_page_fixture_test.go:332` | independent page-tree traversal via `/Kids` → each page's own `/Contents` |
| `mpParseToUnicode` / `mpExtractRuns` | `multi_page_fixture_test.go:162,220` | `/ToUnicode` CMap and per-stream run extraction |

**Do not write a new private reader.** `TestReadEmittedRunsMatchesThePerPageStreamSplitOnAMultiPageDocument`
already cross-checks the two independent traversals; adding a third family is exactly what 2.6a's
review found and penalised.

**Observation, flagged not fixed** (finding, per the brief's instruction): `splitPageContentStreams`
uses `mpKidsArrayRE.FindSubmatch` — **one** `/Kids` array — and `mpContentsRefRE.FindSubmatch` — **one**
`/Contents` reference per page. Both hold by construction today (folio emits a flat, single-level page
tree with one content stream per page) and neither is a defect at 50 pages. **They are recorded here so
the next reader does not have to re-derive that they are safe**, and so that if a future story emits a
`/Contents` array or a nested page tree, the assumption is already written down. This story must not
route around them, and must not silently repair them either.

**Also note the `mp` prefix now misnames them**: `reserved_placeholders_test.go` — a different subject —
already calls `mpParseToUnicode` and `mpExtractRuns`. Not this story's to rename; recorded.

### 10. D-2.6.8's own prediction lands here

D-2.6.8, verbatim: *"`/Kids` being the sole instance is a fact about today's feature set, not about the
code — **Story 2.7**, Epic 4's tables, and any future `/Annots` or name tree add ref arrays."*

If this story's implementation emits any new indirect-reference array, it must go through
`writeRefArray`/`appendRefArray` (layer 1, *by construction*), not a hand-rolled loop. **If it emits
none, say so explicitly in the Delivery Log** rather than leaving the prediction unaddressed — a
prediction that did not fire is a finding (`[[measure-predicted-escalation-against-its-precedent]]`).

---

## DECISIONS NEEDED — escalate before development starts

### DN-1 (BLOCKING) — What width is reserved for the slot, and what happens when the substituted digits do not fill it

**AC2 cannot be written down until this is ruled, and it is deliberately left open below.** This is
the design work of the story and the reason it is harder than it reads.

**The constraint, stated precisely.** AD-4 and `epics.md:918-921` both require the slot to be carried
through pass one **with its box already measured**, and resolved **between the passes** by substituting
**pre-measured glyphs**. Measurement happens in `collectTextRuns` (`render.go:282-380`), which runs
**before** `layout.Paginate` (`render.go:886`). So at measurement time neither X nor Y is known. At the
per-page assembly loop (`render.go:889-905`) **both** are known. The between-passes resolution point
therefore exists and is well-defined — but the **box** was fixed earlier, at a digit count nobody knew.

**What is settled and must not be re-litigated:**

- Digit *identity* is irrelevant. All shipped faces are tabular (finding 3). The width is an exact
  function of digit *count* alone, computable by integer advance arithmetic with no re-shaping.
- Y is not circular for a header/footer construct (finding 2).
- There is no universally available equal-width pad glyph (finding 4).
- Alignment is not honoured, so the only consequences of a width mismatch are **within-run glyph
  positions** and **line breaking against the element width** (finding 5).

**The options, with what measurement says about each:**

| option | standing after measurement |
|---|---|
| **(a) Measure the slot at the maximum digit count once Y's magnitude is known** | **Available and non-circular.** `ContentHeight` never consults the footer, so a "paginate first, then measure the footer slot at `digits(Y)`" ordering is legal within pass one. Residual: on a 99-page document, page 9's X is one digit in a two-digit reservation. **Where does the slack go?** — see the sub-options below. |
| **(b) Require tabular / monospaced figures in the face** | **Already true of every shipped face.** As an AC it is satisfied at birth (D-000.28) and no subject can falsify it (D-000.50). It is not a disposition; at most a forward guard (D-000.24). **It does not address digit count at all**, which is the actual variable. |
| **(c) Re-measure and re-justify the line after substitution** | **Excluded by AC2 as the epic words it** — the box is "already measured" before the slot is resolved. If the lead wants this, AC2's wording has to change, and that is the lead's call, not the developer's. |
| **(d) Forbid the construct where the face lacks equal-width digits** | Addresses only digit identity, which is not the problem. Would be a load-time error on a caller-supplied face; no shipped subject can trigger it. |
| **(e) NOT IN THE BRIEF'S LIST — reserve positionally, re-summing pre-measured advances** | The slot resolves to its exact digits and the glyphs **after** it in the same run have their `X` recomputed by summing already-measured advances. **This is arithmetic over pre-measured glyphs, not measurement** — no font consultation, no shaping, no line breaking — so it appears to satisfy AD-4's letter. It needs no pad glyph, which finding 4 shows is the binding constraint on every "nothing moves" design. **It is offered because the brief's four options do not contain it, not as a recommendation.** |

**If (a) is ruled, the slack disposition is a second, separate question** and must be ruled with it,
because it is what the reader sees:
  - **(a1)** left-align the number in the reservation and let the trailing literals sit further right
    on short-numbered pages (a visible ragged gap before "of");
  - **(a2)** zero-pad — `Page 09 of 99` — which is a **product judgment about what the reader is
    shown**, not an implementation detail;
  - **(a3)** pad with an equal-width space — **eliminated by finding 4** for any chain not led by Noto
    Sans.

**Sub-question — the failure mode when the resolved line no longer fits.** Whatever is ruled, a
resolved `Page X of Y` can be wider than the string that was measured (today's literal
`Page {{page}} of {{pages}}` is 26 characters; `Page 1 of 1` is 11 — the resolution makes it
**narrower**, but a template that measured a narrow placeholder and resolves to a wide number is
reachable in the general case). The line cannot be re-broken after the box is fixed. **What happens?**
D-2.6.5's precedent — a declaration-level impossibility becomes a **located error naming the element**,
never a clip, with no public-signature change — is the closest fit, and would compose. **But it is a new
disposition and this story will not pick it silently.**

**Sub-question (c) — is the construct permitted in the CONTENT band?** Finding 2 shows Y is
non-circular **only** for the two repeated bands. If `{{page}}` is allowed in flowed content, the page
count becomes a fixed point of the page-number text and this story becomes materially harder.
`epics.md:914` says *"a footer containing a `Page X of Y` construct"*. **Options**: (i) resolve only in
`pageHeader`/`pageFooter`, and make the construct in the content band a **located template error**;
(ii) resolve only in the repeated bands and leave it **literal** elsewhere (which re-creates today's
behaviour in one place only, and is confusing); (iii) permit it everywhere and solve the fixed point
(out of proportion to FR31). **Recommend (i); do not implement any of them before the ruling.**

### DN-2 — Confirm the gate becomes six, on D-2.6.2's criterion

**The argument, made explicitly as the brief requires — and it comes out in favour of six.**

D-2.6.2's criterion: *a gate obligation is warranted when the artifact is **the only cross-target
artifact for a shipped FR***.

1. **FR31 is shipped by this story.** Not disputed.
2. **Does FR31 have a cross-target artifact today?** **No.** Measured: zero of the ten committed
   fixtures contains `{{page}}` (finding, top of file). `fixtures/multi-page` covers **FR30**
   (pagination) and its footer is a fixed literal, kept deliberately free of the placeholders so this
   story would not re-record it. `matrixDocuments` has eight entries and none renders a page number.
3. **Therefore, without a new entry, FR31 ships with no cross-target artifact** — which is the exact
   sentence D-2.6.2 says *"settles it"*.
4. **And `epics.md:927-929` demands it in terms**: *"hashes match recorded goldens on all four
   targets."*

**Conclusion: the Epic 2 gate legitimately becomes SIX.** D-2.6.6's *"the gate still owes exactly
five"* is a statement about **structural validity of committed bytes** — which *"is not cross-target:
the bytes are identical on every leg by construction"* — and does **not** reach a new rendered
document, whose legs genuinely can diverge. The two rulings compose; neither is being re-opened.

**And the D-000.4 override is separately DECLINED.** Page-number substitution introduces **no new
source of cross-target divergence**: integer advance arithmetic on `geom.Length`, no float, no vendor
call, no compressor, no new dependency. D-000.4's own warning — that on a weaker trigger *"nearly every
story would qualify"* — applies exactly. **Obligation yes, override no**, which is the same
disposition D-2.6.2 reached for `multi-page`.

**What this story owes instead is D-000.54's native leg**, which is a **sequencing fix and must not be
logged as a cadence override**. It is AC5 and a task.

**Escalated rather than assumed** only because the brief asked for the argument either way and a
one-line addition to `declaredEpic2GateObligations` is, by D-2.5.1's design, a reviewable decision
point. **If the lead agrees, no re-ruling is needed** — D-2.6.2's criterion already authorises it and
the developer may proceed on this section as the authorisation.

---

## Scope fence — what this story is NOT

- **Not Story 2.8.** D-2.6.5 records that 2.8 requires clip **and** diagnostic **and** bytes-still-
  returned, and that *"is not expressible in the current public API"*, so **2.8 owns a
  diagnostic-channel design decision on `Render`'s surface**. This story must not presume an answer to
  it, must not add a non-fatal channel, and must not change `Render`/`RenderTo`'s signature.
- **Not the expression language.** Epic 3 mints it. This story only makes the `page` **namespace**
  structurally unavailable to it.
- **Not alignment.** `align` stays parsed-and-unhonoured (finding 5). Implementing centring to make a
  page number look nice is out of scope and would move goldens.
- **Not a re-recording of the three pinned digests.** A new fixture of this story's own is expected;
  moving `shaped-text`, `expected-breaks` or `multi-page` is not.
- **Not `epic-2: backlog`.** That line stays; the **gate** flips it, not this story.
- **Not a repair of the `mp*` helper naming or the single-`/Kids` assumption** (finding 9). Recorded,
  flagged, left alone.

---

## Acceptance Criteria

### AC1 — X is the current page and Y the document total, correct on EVERY page

`epics.md:914-916` verbatim. Given a footer containing a `Page X of Y` construct, when the document
renders, **X is the current page and Y the document total, correct on every page.**

**How it is asserted (D-000.21):** off the **produced content stream**, recovering the literal
characters through the document's own `/ToUnicode` CMap, using the repaired shared readers
(`splitPageContentStreams` + `mpParseToUnicode` + `mpExtractRuns`). **Asserting on the input, on the
page model, or on "a substitution occurred" does not carry the property** — the same argument
`reserved_placeholders_test.go` already makes for the inverse assertion.

**Presence precondition (mandatory):** the assertion is meaningless below N ≥ 2 pages, and the
digit-count property is meaningless below N ≥ 10. The test must **fatal** if its subject renders fewer
pages than the property needs, naming the count it got — the shape
`TestReservedPagePlaceholdersPassThroughOnEveryPage` already uses.

**Vacuity guard:** a subject where every page's X has the same digit count as Y cannot express the
central defect. At least one subject must span a power-of-ten boundary (page 9 → page 10).

### AC2 — The slot is late-bound, pre-measured, and resolved between the passes

`epics.md:918-921` verbatim: given the page model after pass one, when page-number text is inspected,
**it is carried as a late-bound slot whose box is already measured**, and **it is resolved between the
passes by substituting pre-measured glyphs**.

**Structural half — settled, implement it:**

1. The slot lives in `internal/pagemodel` and **names no PDF concept** (AD-5). `internal/arch`'s
   `TestPageModelNamesNoPDFConcept` is the existing subject that expresses a violation.
2. `internal/pdf` performs **no** measurement, **no** line breaking and **no** pagination in resolving
   it. `internal/passtwo_arch_test.go` is the existing subject; extend it, do not duplicate it
   (D-000.42).
3. Resolution happens **between** the passes — after `layout.Paginate` returns and before
   `pdf.SerializeTextDocument` is called — at a point where **both X and Y are known**
   (`render.go:889-905` is that point today).
4. The glyphs substituted are **pre-measured**: shaped during pass one, never re-shaped. A red-proof
   is available and required — reverting to a re-shape at resolution time must **redden** a test, or
   this AC is an intention rather than a guard.

**Width half — DELIBERATELY NOT WRITTEN. BLOCKED ON DN-1.**

> This AC is incomplete on purpose. The box is measured before Y's magnitude and before any page's X
> is known, and **what happens when the substituted digits are not the width the slot was measured
> for** — `Page 9 of 9` against `Page 10 of 99` — is a design decision the story creator declined to
> make silently. The options, and what measurement says about each, are in **DN-1**. **The developer
> must not choose one.** When the lead rules, this AC gains a clause stating the reserved width, the
> disposition of any slack, and the failure mode when a resolved line no longer fits its element —
> and only then does it become implementable.

### AC3 — No `page` namespace exists, and none can be added — enforced structurally, with a subject that can express the violation

`epics.md:923-925` verbatim: given the expression language, when an author attempts to reference
pagination, **no `page` namespace exists and none can be added.**

**This is a negative requirement, so D-000.50 applies first: which subject can express its
violation?** The expression language does not exist yet (Epic 3, Story 3.2), so *"grep Epic 3's
evaluator for `page`"* is inspection over an empty set — a guard that cannot fail, which is not
coverage.

**The subject that CAN express it, today, is the resolver's ROOT SET.** `internal/bind/text.go`'s
fence (`BindText`'s own doc comment) already names the exact mechanism of the violation: *"page/pages are reserved whole **TOKENS**,
resolved from neither root … params is a **NAMESPACE**, resolved from its own root — **conflating the
two is how "page" would eventually acquire a namespace**."* A `page` namespace comes into existence
precisely when `page` becomes a **third resolution root**.

**Precedent for enforcing an absence structurally rather than by inspection — follow it.**
D-2.6.8 layer 1: *"a `writeRefArray(ids []int64)` … that **cannot** omit the separator … This prevents
the class **by construction**, which outranks any guard over it."* The same move applies here: make the
root set a **closed, declared thing** — the shape `internal/template/closedsets.go` already uses for
`align`/`valign`, and `declaredEpic2GateObligations` uses for the gate — so that adding a third root is
a **visible one-line diff to a declared list** that a guard compares against the **observed** set, in
both directions. That is D-2.5.1's mechanism, and it is the project's established answer to "this
absence must stay absent".

**Requirements:**

1. The set of resolution roots is **declared** in one place and the guard asserts the **observed** set
   equals it exactly, both directions — never a count, never in a test name (D-2.5.1).
2. **Red-proof, executed and recorded**: adding a `page` root must make a named test fail. Per D-000.52
   — *a structural claim about a guard is worth exactly as much as its demonstration* — introduce the
   violation, observe which guard reddens, name the ones that stayed green, and record the invocation.
3. The **negative half survives** (finding 7): a non-reserved placeholder must still resolve-or-error.
   A rewrite that only asserts "the resolved text appears" is green on a renderer that has stopped
   binding data entirely, and that is a silent disarming (D-000.34).
4. The fence comment in `internal/bind/text.go` (`BindText`'s own doc comment) is **preserved or
   strengthened**, never deleted. It is the written record of why the distinction exists.

### AC4 — The inverted reservation tests, with nothing silently disarmed

`folio-go/reserved_placeholders_test.go`'s pass-through assertion is **inverted**: the same document
must now draw `Page 1 of 2` / `Page 2 of 2` instead of the literal. The file's own comment authorises
this.

**But (D-000.34, D-000.42):**

- Every assertion removed must be **named**, with the property it carried and where that property now
  lives. An assertion that carried no property elsewhere may not simply disappear.
- The set pin **moves** to AC3's root-set guard rather than being deleted.
- The negative half is **kept and still reddens**; demonstrate it.

### AC5 — A matrix document is registered, and the gate's obligation list gains exactly one line

Per **DN-2** and `epics.md:927-929`.

1. A new entry in `matrixDocuments` (`folio-go/matrix_test.go:1012`) with a non-empty `slug`, its
   `fixtureRelPath`, and `wantPages` set to the document's real page count (the field exists precisely
   because Story 2.6 added the first N > 1 document).
2. **Exactly one** new line in `declaredEpic2GateObligations` (`folio-go/byte_neutrality_test.go:466`),
   naming the story and the ruling that authorised it, in the established comment form.
   `TestEpic2GateObligationsMatchTheDeclaredSet` must be green with no rename and no count anywhere.
3. `.github/workflows/matrix.yml` registration kept in step —
   `matrix_registration_test.go` pins the slugs against it.
4. **D-000.54's native leg is RUN, once, before this story reaches `review`:**
   ```
   FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .
   ```
   **This is a sequencing fix, NOT a D-000.4 cadence override, and must not be logged as one.**
5. **D-000.54's guardrail, stated in the Delivery Log in both halves**: the native leg proves the leg
   **executes and produces a hash on one target, `darwin/arm64`**; it proves **nothing** about
   cross-target agreement, which remains the gate's job.
6. **D-000.55**: the Delivery Log names what was **executed, by target**, and names the unrun legs
   **individually** with the cadence clause deferring them. The phrase *"written, compiled and vetted,
   deliberately not run"* is **banned**.

**Note on the document's shape.** `epics.md` names 1, 5, 20 and 50 pages. **20 and 50 both cross the
one-to-two-digit boundary**, which is why they are in the spec — finding 8 shows no existing fixture
does. Whether all four become matrix documents or one document with the others as ordinary-suite
subjects is an implementation judgment; **at least one matrix-registered document must span the digit
boundary**, or AC5 covers FR31 without covering its hard case.

### AC6 — The three pinned digests do not move, and it is measured rather than argued

**This story edits the renderer, so non-movement is a measurement, not a consequence of scope.**

Re-compute at the story's own commit, by the file route (D-000.12 as corrected: `rtk proxy … > file`,
never through the wrapper's pipes), at **every declared site** (D-000.47's four-site list):

| artifact | must equal |
|---|---|
| `fixtures/shaped-text/expected.pdf` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` |
| `fixtures/expected-breaks/expected_breaks.json` | `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de` |
| `fixtures/multi-page/expected.pdf` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` |

The first two carry **pending owner sign-offs**; movement invalidates a review the owner has already
been asked to perform. **If any digest moves, stop and escalate — do not update a digest to make a
test go green** (AD-21, D-000.22).

**And the broader claim, measured the same way**: no *other* committed golden moves either. State the
count of goldens checked and the commit.

### AC7 — Every new golden passes D-000.53's independent-reader acceptance

No golden this story records — first recording or re-recording — is accepted until **a reader this
project did not write** parses it and resolves it into the objects it claims to contain: the page tree
resolves to N pages, every indirect reference in it names an object that exists, and N equals the count
the artifact declares.

**Recorded in the fixture's provenance**: the reader, **its version**, the **verbatim invocation**, and
its output. `qpdf` 12.4.0 at `/opt/homebrew/bin/qpdf` is today's instrument, named illustratively — the
rule is stated as an outcome, not a tool.

**This step is not discharged by an assertion this project wrote.** D-000.53's guardrail exists because
`assertWellFormedPDF` encoded the single-page shape and was therefore silently never invoked on the
first multi-page artifact. **A 20- or 50-page document is again a shape no existing guard was written
for** — check, do not assume, that each in-repo checker is actually reachable on this story's subject.

### AC8 — The baseline is preserved and any delta is itemised

The ordinary suite ends at **564 all-occurrences PASS / 1 FAIL** and **345 top-level PASS / 1 FAIL**,
plus whatever this story adds, with the **only** failure remaining `TestCorpusMeetsP6ExerciseFloors`
(the intentional Story 2.1 red, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`).

Report the closing figures with the **same three-part scope** as the baseline (*whole module, no build
tags, `-count=1`, all-occurrences and top-level separately*). **Itemise every delta**: a test added, a
test inverted, a test moved. A net figure that happens to match while an assertion was lost is exactly
what D-000.34 is about.

### AC9 — Every enumeration names its commit and its tree state

D-2.6.9 (correction). Any count, inventory or table this story records **names the commit it was taken
at**, and is taken against a tree with **no uncommitted changes in the counted files** — or says
plainly that it was not. **This has bitten three agents including the orchestrator.**

---

## Heavy-test cadence — proposed DECLINED, stated so it can be refused

**Per-epic cadence stands. Unit tests always.**

**A D-000.4 per-story matrix override is NOT requested**, and the argument is in **DN-2**: page-number
substitution introduces no new source of cross-target divergence — integer advance arithmetic on
`geom.Length`, no float, no vendor call, no compressor, no new dependency. D-000.4's named overrides
are 1.2, 1.5, 1.8, 2.4 and 4.7; this story is not among them and does not qualify.

**What IS owed is D-000.54's native leg** — one local `go test`, no Docker — which is a **sequencing
fix and must not be logged as an override**. It is AC5.4.

**D-000.55 applies to the write-up.** Name what was executed and on which target. Name the unrun legs
individually. Do not offer `go vet -tags matrix` as assurance that a deferred leg works.

---

## Task breakdown

1. [x] **Get DN-1 ruled.** Ruled: D-2.7.1 (box fixed), D-2.7.2 (reservation width/alignment). See
   `folio-mvp-decision-log.md`.
2. [x] **Confirm DN-2** with the lead. Ruled: D-2.7.4 (gate to six).
3. [x] Re-verify the baseline and the three digests at the story's own start commit; record both with
   the commit and the tree state (AC9). Verified at `ecd0056` (this story's baseline commit); see
   Completion Notes.
4. [x] Build the subject: documents of 1, 5, 20 and 50 pages, at least one spanning the page-9→page-10
   digit boundary. `fixtures/page-count-{1,5,20,50}/input.folio` byte-identical to
   `page_count_matrix_templates.go`'s constants (`TestPageCountFixturesMatchTheInRepoTemplates`).
5. [x] Implement AC3's root-set closure **first**. `internal/bind/declaredResolutionRoots` +
   `TestBindResolutionRootsAreClosed` + `TestBindResolutionRootsClosureRedProof` (executed, D-000.52).
6. [x] Implement the slot: `internal/pagemodel.PageNumberSlot`, pass-one carriage (`page_number.go`,
   `render.go`'s two-phase `collectBandTextRuns`), between-passes resolution at `paginateDocument`'s
   per-page assembly (`resolvePageRunForPage`). Width behaviour per D-2.7.1/D-2.7.2.
7. [x] Extended `internal/passtwo_arch_test.go`'s `layoutSymbolSubstrings` for AC2.2 rather than adding
   a parallel guard (D-000.42).
8. [x] Inverted `reserved_placeholders_test.go`
   (`TestReservedPagePlaceholdersResolveOnEveryPage`); moved the set pin to
   `internal/bind`'s structural guard; kept the negative half reddening (AC4).
9. [x] Recorded the `page-count-20` golden; ran D-000.53's independent reader (`qpdf` 12.4.0); reader,
   version and verbatim invocation in `fixtures/page-count-20/README.md` (AC7).
10. [x] Registered `page-count-20` in `matrixDocuments`; added **one** line to
    `declaredEpic2GateObligations`; `.github/workflows/matrix.yml` updated in step (AC5.1–3).
11. [x] **Ran D-000.54's native leg** (`darwin/arm64`) and recorded both halves of its guardrail
    (AC5.4–5) — see Completion Notes and `fixtures/page-count-20/README.md`.
12. [x] Re-measured the three pinned digests and every other golden by the file route at every declared
    site (AC6) — unmoved; see Completion Notes.
13. [x] Closed the baseline with scoped figures and an itemised delta (AC8) — see Completion Notes.
14. [x] Wrote the Delivery Log per D-000.55: executed vs unrun, by target, individually.

**There is no commit task.** This story ends at *status → review*.

---

## Flagged, not fixed

- `splitPageContentStreams` assumes exactly one `/Kids` array and one `/Contents` reference per page
  (finding 9). Safe by construction today, at any page count. Recorded so it is written down, not
  re-derived.
- `mpParseToUnicode` / `mpExtractRuns` are `mp`-prefixed but are already consumed by a non-multi-page
  subject. The prefix misnames them. Not this story's rename.
- `align` is parsed, validated against a closed set, and honoured by nothing (finding 5). A template
  author can write `"align": "center"` today and get left-aligned output with no diagnostic. **Out of
  scope here; worth a deferred-work entry.**
- The three shipped faces disagree on digit advance (572 / 572 / 555). Harmless today because digits
  never fall back (finding 6), but a caller-supplied `chain[0]` lacking digits would put the number on
  a different face from the surrounding literals. **Forward hazard with no available red-proof
  (D-000.24); do not write it as coverage.**

---

## Dev Agent Record

### Delivery Log

Per D-000.55: what was **executed**, by target, individually; unrun legs named individually. The
banned phrase *"written, compiled and vetted, deliberately not run"* does not appear below because
no leg is left in that state without saying so explicitly.

**Executed, this session, on `darwin/arm64` (the only target this environment can run natively):**

- `go build ./...` — success.
- `go vet ./...` — no issues.
- `go build -tags matrix ./...` — success (the matrix-tagged sources, including the new
  `page-count-20` registration and its capture/guard functions, compile).
- `go vet -tags matrix ./...` — no issues.
- `go test ./... -count=1` (ordinary suite, all packages, no build tags) — **579 all-occurrences
  PASS / 1 FAIL, 355 top-level PASS / 1 FAIL**; the one failure is the same intentional Story 2.1
  red, `TestCorpusMeetsP6ExerciseFloors`, reproduced character-for-character:
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.
- **D-000.54's native leg**, verbatim:
  `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .`
  — **PASS**. Every registered document's hash matched its recorded golden exactly, including
  **two of the three** pinned digests (`shaped-text`, `multi-page` — corrected, finisher, this
  story's review, Finding 14: `expected-breaks` is not a `matrixDocuments` entry and could not
  have been covered by `TestTargetRenderHash`; it is covered by AC6's own direct
  `shasum`-over-the-file measurement instead) and the new `page-count-20` entry
  (`b32fa1c5babb8327b09b5c2bc0a11628b8c8885b9c5661c0262ec24920c5150f`). This proves the leg
  **executes and produces a hash on one target, `darwin/arm64`**; it proves **nothing** about
  cross-target agreement, which remains the gate's job. This is a **sequencing fix per D-000.54,
  NOT a D-000.4 cadence override** — the override is separately and correctly **declined**
  (D-2.7.4: integer advance arithmetic on `geom.Length`, no float, vendor call, compressor or new
  dependency).
- `qpdf --check fixtures/page-count-20/expected.pdf` and `qpdf --show-npages
  fixtures/page-count-20/expected.pdf` (`qpdf` 12.4.0, `/opt/homebrew/bin/qpdf`) — D-000.53's
  independent reader: structurally valid, 20 pages, matching the declared count. Verbatim
  invocation and output recorded in `fixtures/page-count-20/README.md`.

**Unrun, individually named, deferred to the Epic 2 boundary gate (per D-000.4/D-2.7.4, not this
story's obligation):** the `linux-amd64` and `linux-arm64` Docker legs and the `js-wasm` leg of
`TestCrossTargetByteIdentity` / the gate's own matrix workflow. `go vet -tags matrix` was **run**,
not offered as a substitute for these — see D-000.55.

**Baseline re-verification (task 3, AC9).** At `ecd0056` (this story's `baseline_commit`), on a
tree carrying only this story's own uncommitted changes (no other uncommitted work): the three
pinned digests reproduced exactly —
`fixtures/shaped-text/expected.pdf` = `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370`,
`fixtures/expected-breaks/expected_breaks.json` = `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de`,
`fixtures/multi-page/expected.pdf` = `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b`
— via `rtk proxy shasum -a 256`, the file route (D-000.12 as corrected). `git status --porcelain
fixtures/` shows zero modification to any pre-existing fixture; only the four new
`fixtures/page-count-*/` directories are untracked additions (AC6).

**AC8's itemised delta — corrected (finisher, this story's review, Finding 5).** The developer's
original table below claimed "+17 top-level" reconciled against a measured "+10" by an invented
mechanism ("new sub-tests' internal assertions counting differently under `-v`'s `--- PASS` per
`t.Run`") that does not exist: `--- PASS` lines count subtests, never assertions. Re-measured
directly against the reviewed tree (developer's own delta, before this finisher's own additions
below):

| measure | developer's delta | accounting |
|---|---|---|
| new top-level test functions | **+10**, exactly | `TestBindResolutionRootsAreClosed`, `TestBindResolutionRootsClosureRedProof` (`internal/bind`); `TestResolvePageRunForPageUsesPreMeasuredGlyphsNotReshaping`, `TestBuildPageNumberSlotFailsClosedOnNonUniformDigitAdvances`, `TestPageSlotInContentBandIsALocatedError`, `TestPageSlotInContentBandPagesIsAlsoALocatedError` (`page_number_test.go`); `TestShippedFacesHaveUniformDigitAdvances`, `TestUniformDigitAdvanceGuardRedProof` (`tabular_digits_test.go`); `TestPageCountFixturesMatchTheInRepoTemplates`, `TestPageCountFixturesRenderCorrectPageNumbersEverywhere` (`page_count_matrix_test.go`) — 10 functions, nothing more |
| top-level PASS | 345 → 355 = **+10** | every new top-level test accounted for; **no top-level test was lost** |
| all-occurrences PASS | 564 → 579 = **+15**, i.e. **+5 subtests** | `t.Run` appears exactly once in the new files (`page_count_matrix_test.go`, four fixtures → 4 subtests) plus `TestEveryGoldenPDFResolvesItsPageTree/page-count-20`, produced by the new `goldenDigestRecord` entry flowing into an EXISTING loop (`golden_structural_validity_test.go`) — confirmed by grepping the `-v` output for subtests naming `page-count-20`, which returns exactly those two lines |
| tests inverted | 1 (no net count change) | `TestReservedPagePlaceholdersPassThroughOnEveryPage` → `TestReservedPagePlaceholdersResolveOnEveryPage` (same test, renamed and inverted per Story 2.6's own authorisation) |
| tests kept, narrowed | 1 (no net count change) | `TestReservedPlaceholderSetIsUnchanged` — positive half removed (superseded by the structural guard, D-000.34), negative half unchanged and still reddens |

**The true accounting is clean and simple: +10 top-level, +5 subtests, nothing removed.** The
"+17" the developer reported was a subtest-inclusive figure carrying a *top-level* label; I
verified nothing was actually lost (confirmed independently below), but the evidence offered for
that conclusion was wrong.

**This finisher's own additions, on top of the developer's 355/1 (top-level) and 579/1
(all-occurrences), same three-part scope**, addressing this review's findings (see Finding
Resolutions below for which finding each closes):

| test | top-level | subtests | file |
|---|---|---|---|
| `TestReservedPagePlaceholdersResolveTwoOccurrencesInOneElement` | +1 | — | `reserved_placeholders_test.go` (Blocker 1 red-proof) |
| `TestResolvePageRunForPageHandlesTwoSlotsInOneRun` | +1 | — | `page_number_test.go` (Blocker 1 red-proof) |
| `TestBindTextDottedPageAndPagesPathsAreOrdinaryDataPaths` | +1 | +4 (`t.Run` per case) | `internal/bind/text_test.go` (Blocker 2's primary guard) |
| `TestCollectTextRunsMatchesTheShippingBandComposition` | +1 | — | `collect_text_runs_composition_test.go` (Finding 8) |
| **total** | **+4** | **+4** | — |

**Closing figures**: top-level 355 + 4 = **359**; all-occurrences 579 + 8 (4 top-level + 4
subtests) = **587**. Measured and reported in the Delivery Log's finisher addendum below.

### Delivery Log — finisher addendum (D-000.55: executed vs unrun, by target, individually)

**Triaged and fixed 2 Blockers, 4 Majors, 6 Minors and 3 Nits** (see Finding Resolutions, below
the QA Results section, for the disposition and rationale of each). Every FIX was red-proved
before being credited (D-000.30/D-000.40): each mutation is a real edit to the production source
or test source, run, observed to redden, reverted by hand, and the restore confirmed by
`shasum -a 256` (never by an exit code, and never by a bare `diff`, per this project's own
precedent for both traps).

**Executed, this session, on `darwin/arm64`:**

- `go build ./...` and `go vet ./...` — clean, after every fix.
- `go build -tags matrix ./...` and `go vet -tags matrix ./...` — clean.
- `go test ./... -count=1 -v` (ordinary suite, no build tags) — **587 all-occurrences PASS / 1
  FAIL, 359 top-level PASS / 1 FAIL**; the sole failure is the same intentional Story 2.1 red,
  `TestCorpusMeetsP6ExerciseFloors`, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`,
  character-for-character unchanged from the baseline and the developer's own closing figures.
- `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .` —
  **PASS**, re-run after every fix. Every registered document's hash, including all three pinned
  digests and `page-count-20`'s own recorded hash
  (`b32fa1c5babb8327b09b5c2bc0a11628b8c8885b9c5661c0262ec24920c5150f`), matched its recorded golden
  exactly — **unmoved by any of this session's fixes**.
- `shasum -a 256` on the three pinned digests, by the file route, after every fix — all three
  reproduce exactly:
  `fixtures/shaped-text/expected.pdf` = `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370`,
  `fixtures/expected-breaks/expected_breaks.json` = `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de`,
  `fixtures/multi-page/expected.pdf` = `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b`.
  `git status --porcelain fixtures/` shows zero modification to any pre-existing fixture; only the
  four `fixtures/page-count-*/` directories are untracked additions.
- `qpdf --check` and `qpdf --show-npages` (`qpdf` 12.4.0, `/opt/homebrew/bin/qpdf`) on
  `fixtures/page-count-20/expected.pdf` — re-run: structurally valid, 20 pages, output
  byte-identical to what `fixtures/page-count-20/README.md` already recorded.
- `gofmt -l folio-go/` (`/opt/homebrew/Cellar/go/1.26.0/libexec/bin/gofmt`) — no output (clean).
- **Two live red-proofs of this finisher's own fixes**, each a real edit to committed production
  source, run, reverted by hand, restore confirmed by `shasum -a 256` comparison (not `diff`,
  not an exit code):
  1. **Blocker 1**: `folio-go/page_number.go`'s `resolvePageRunForPage`, temporarily made to
     resolve only the LAST `PageSlots` entry (reproducing the pre-fix scalar-overwrite behaviour).
     `TestReservedPagePlaceholdersResolveTwoOccurrencesInOneElement` and
     `TestResolvePageRunForPageHandlesTwoSlotsInOneRun` both **FAILED**, reproducing the
     reviewer's exact transcript (`"Page 0 of 2 / 1"`, `"Page 0 of 2 / 2"`). Reverted; `go build
     ./... && go vet ./...` clean afterward.
  2. **Blocker 2**: `folio-go/internal/bind/text.go`, the reviewer's exact mutation —
     `if path[0] == "page" { record(nil); continue }` inserted immediately above the params
     branch. `TestBindTextDottedPageAndPagesPathsAreOrdinaryDataPaths` **FAILED** on the
     `{{page.number}}`/`{{page.total}}` subtests, while `TestBindResolutionRootsAreClosed`
     (the pre-existing structural guard, widened but not replaced) stayed **green** —
     reproducing exactly the gap the reviewer's finding described. Reverted; `go build ./...
     && go vet ./...` clean afterward.

**Unrun, individually named, deferred to the Epic 2 boundary gate (per D-000.4/D-2.7.4, carried
unchanged from the developer's own submission, not this story's obligation):** the `linux-amd64`
and `linux-arm64` Docker legs and the `js-wasm` leg of `TestCrossTargetByteIdentity` / the gate's
own matrix workflow.

### File List

**New:**
- `folio-go/page_number.go` — the {{page}} slot mechanism (tokenizer, reservation, digit table,
  between-passes substitution).
- `folio-go/page_number_test.go` — AC2's red-proof, the non-uniform-advance fail-closed test, and
  D-2.7.3's content-band located-error red-proofs.
- `folio-go/tabular_digits_test.go` — D-2.7.2's declarative tabular-figures guard and its
  red-proof (D-000.52).
- `folio-go/page_count_matrix_templates.go` — the four page-count-N document templates (Go
  constants).
- `folio-go/page_count_matrix_test.go` — fixture/template parity and AC1's correctness assertions
  for all four page counts.
- `folio-go/internal/bind/resolution_roots_arch_test.go` — AC3's closed-resolution-root-set guard
  and its executed red-proof.
- `fixtures/page-count-1/input.folio`, `fixtures/page-count-5/input.folio`,
  `fixtures/page-count-20/{input.folio,expected.pdf,expected.json,README.md}`,
  `fixtures/page-count-50/input.folio`.
- `folio-go/collect_text_runs_composition_test.go` (**finisher**) — Finding 8: turns the
  "documentBands order == renderDocument's append order" coincidence into a checked invariant.

**Modified:**
- `folio-go/render.go` — two-phase text-run collection (content band first, to learn Y; then the
  two repeated bands), `positionSegments` extended with slot marking, `paginateDocument`'s per-page
  loop resolves page slots, `digitTableBandIndex` skip. **Finisher**: `textRunSource.pageSlots` is
  now a slice (Blocker 1); `positionSegments` returns an error instead of panicking on a straddling
  reservation (Finding 10); `firstReservedPageToken`/`resolvePageTokens` calls pass `subs` through
  (Finding 4); citation and ordering-disclosure comment fixes (Findings 6, 9).
- `folio-go/internal/pagemodel/pagemodel.go` — `PageNumberSlot` type, `TextRun.PageSlot` field.
  **Finisher**: `TextRun.PageSlot` renamed to `TextRun.PageSlots []PageNumberSlot` (Blocker 1).
- `folio-go/internal/bind/text.go` — `declaredResolutionRoots` (AC3's declared set). **Finisher**:
  fence citation de-positioned (Finding 7).
- `folio-go/internal/passtwo_arch_test.go` — extended `layoutSymbolSubstrings` for AC2.2.
- `folio-go/reserved_placeholders_test.go` — inverted per AC4. **Finisher**: added
  `TestReservedPagePlaceholdersResolveTwoOccurrencesInOneElement` (Blocker 1's full-render
  red-proof).
- `folio-go/matrix_test.go` — `page-count-20` registered in `matrixDocuments`; capture/extraGuard
  functions. **Finisher**: citation fix (Finding 6).
- `folio-go/render_test.go` — `subprocessPageCount20EnvVar` selector.
- `folio-go/byte_neutrality_test.go` — `declaredEpic2GateObligations` gains one line;
  `goldenDigestRecord` gains the `page-count-20` entry.
- `.github/workflows/matrix.yml` — `page-count-20` added to the `docs=` list and every target's
  `upload-artifact` path list.
- `_bmad-output/specs/spec-folio/folio-format.md` — D-2.7.3's outcome statement, beside the
  existing `{{page}}`/`{{pages}}` prose at line 494. **Finisher**: trimmed to end after "naming the
  element." per D-2.7.3's own instruction (Finding 3).
- `folio-go/page_number.go` (**finisher**) — `resolvePageRunForPage` rewritten for a run carrying
  more than one `PageSlots` entry (Blocker 1); `firstReservedPageToken`/`resolvePageTokens` refuse
  to treat a "{{page}}"/"{{pages}}"-shaped occurrence inside a data substitution's span as reserved
  (Finding 4); `digitTableRun`'s `positionSegments` call updated for its new error return.
- `folio-go/page_number_test.go` (**finisher**) — `PageSlot` → `PageSlots` field references
  updated; added `TestResolvePageRunForPageHandlesTwoSlotsInOneRun` (Blocker 1's unit-level
  red-proof); strengthened three weak assertions to check error identity, not just presence
  (Finding 11).
- `folio-go/internal/bind/text_test.go` (**finisher**) — added
  `TestBindTextDottedPageAndPagesPathsAreOrdinaryDataPaths`, AC3's primary behavioural guard
  (Blocker 2).
- `folio-go/internal/bind/resolution_roots_arch_test.go` (**finisher**) — widened the AST scan
  from one hardcoded filename (`text.go`) to every non-test `.go` file in the package (Finding 2's
  first evasion); fence citation de-positioned (Finding 7).
- `folio-go/tabular_digits_test.go` (**finisher**) — ranges `FontSet`'s own keys instead of a
  hardcoded three-face literal list, so an added (not just a removed) shipped face is checked
  (Finding 13).
- `_bmad-output/implementation-artifacts/2-7-render-page-x-of-y.md` (**finisher**) — this file:
  Finding Resolutions section added; AC8 delta table, Completion Notes, citations, plain-terms
  opener corrected; status → `done`.
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (**finisher**) —
  `2-7-render-page-x-of-y: review` → `done`.
- `fixtures/page-count-20/README.md` (**finisher**) — corrected citation, added `validated at` row
  (Finding 15).

### Change Log

- Implemented Story 2.7: `{{page}}`/`{{pages}}` resolve in the page-header/page-footer bands
  (D-2.7.1–D-2.7.3), the Epic 2 gate's obligation count moves to six (D-2.7.4).
- Inverted Story 2.6's negative-control test for the reservation's pass-through; kept its negative
  half (non-reserved placeholders still error) reddening.
- Added the closed-resolution-root-set structural guard in `internal/bind` (AC3), replacing the
  behavioural set-pin the inverted test used to carry.
- **Finisher, after code review**: fixed two Blockers (a run carrying more than one `{{page}}`
  occurrence resolved only the last one, silently; AC3's guard could not detect a `page`
  namespace added via an early-return dispatch shape) and four Majors (a spec doc leaked
  mechanism; a bound data value could inject a live `{{page}}` construct or break rendering; AC8's
  reported delta was wrong; every `epics.md` citation in the story and three in shipped source was
  wrong). See Finding Resolutions below for the full triage of all 15 findings.

### Completion Notes

**Updated by the finisher, after this story's review (see Finding Resolutions, below the QA
Results section).** The developer's original version of this section claimed all nine ACs met;
the review found AC1 and AC3 were not (Blockers 1 and 2), and AC8/AC9 carried inaccuracies
(Findings 5, 6). Both blockers are now fixed and red-proved; the text below states the ACs'
status as of this story's finishing commit, not as of the developer's original submission.

**All nine acceptance criteria are met:**

- **AC1** — X/Y correct on every page, asserted off the produced content stream through each
  document's own `/ToUnicode` CMap, for all four page counts (1/5/20/50), including the
  page-9-to-page-10 digit boundary (`page_count_matrix_test.go`,
  `TestPageCountFixturesRenderCorrectPageNumbersEverywhere`, and the matrix guard
  `requirePageCount20HasCorrectPageNumbers`). **Also now correct for more than one `{{page}}`
  occurrence in a single element** (Blocker 1, fixed by the finisher):
  `pagemodel.TextRun.PageSlots` is a slice, not a scalar field, and
  `TestReservedPagePlaceholdersResolveTwoOccurrencesInOneElement` red-proves it on a real template
  reproducing the reviewer's exact repro (`Page {{page}} of {{pages}} / {{page}}`).
- **AC2** — the slot lives in `internal/pagemodel` (`PageNumberSlot`), names no PDF concept
  (`TestPageModelNamesNoPDFConcept` passes); `internal/pdf` performs no measurement/pagination in
  resolving it (`internal/passtwo_arch_test.go` extended); resolution happens between the passes,
  in `paginateDocument`'s per-page assembly; the glyphs substituted are pre-measured, proven by
  `TestResolvePageRunForPageUsesPreMeasuredGlyphsNotReshaping` (D-000.52's executed red-proof: the
  resolution function's own signature carries no FontSet/cache, so it cannot shape). Width per
  D-2.7.1/D-2.7.2, as ruled. The straddling-reservation internal impossibility is now a **located
  error**, not a panic (Finding 10, fixed by the finisher: `positionSegments` returns an error;
  the public `Render` entry point can no longer have an internal panic cross it).
- **AC3** — no `page` namespace exists: `internal/bind`'s resolution roots are a declared, closed
  set (`{"data","params"}`), asserted against the OBSERVED set by AST-scanning `lookupBound` call
  sites (`TestBindResolutionRootsAreClosed`), with an EXECUTED red-proof
  (`TestBindResolutionRootsClosureRedProof`) demonstrating the guard catches a third root. The
  fence comment in `internal/bind/text.go` (`BindText`'s own doc comment) is preserved. The negative half of
  `TestReservedPlaceholderSetIsUnchanged` survives and still reddens. **AC3's own stated
  requirement — "a subject that can express the violation" — is now met on the NAMESPACE shape,
  not only the extra-root shape** (Blocker 2, fixed by the finisher):
  `TestBindTextDottedPageAndPagesPathsAreOrdinaryDataPaths` (`internal/bind/text_test.go`) asserts
  BEHAVIOURALLY that `{{page.number}}`/`{{page.total}}`/`{{pages.number}}`/`{{pages.total}}`
  resolve as ordinary absent-data-path errors — the property AC3 actually names — and reddens
  under the reviewer's exact mutation, where `TestBindResolutionRootsAreClosed` (kept, as a
  narrower secondary check, and widened to scan the whole package rather than one hardcoded
  filename) does not.
- **AC4** — `TestReservedPagePlaceholdersPassThroughOnEveryPage` inverted to
  `TestReservedPagePlaceholdersResolveOnEveryPage`; the set pin moved to AC3's structural guard;
  the negative half kept and demonstrated to still redden.
- **AC5** — `page-count-20` registered in `matrixDocuments` as the gate's sixth obligation (one
  line in `declaredEpic2GateObligations`, `TestEpic2GateObligationsMatchTheDeclaredSet` green with
  no rename/count); `.github/workflows/matrix.yml` kept in step
  (`TestMatrixDocumentSlugsAreRegisteredInCI` green); D-000.54's native leg run once on
  `darwin/arm64` and PASSED, both halves of the guardrail stated above (re-run and re-confirmed by
  the finisher, see the Delivery Log addendum). **Qualification, stated plainly rather than left
  implicit (Finding 12, D-000.17):** the epic's fourth AC's *"hashes match recorded goldens on all
  four targets"* clause is satisfied for `page-count-20` only — the smallest of the four sizes
  that spans the digit-count boundary, by design and with D-2.7.4's authorisation. The 1-, 5- and
  50-page documents have no recorded golden at all (only `input.folio`, pinned against the in-repo
  template constants); AC1's **correctness** property is verified for all four, on one target, but
  the epic's **hash** clause is inapplicable to three of them, not weakly satisfied. The judgment
  to register only one matrix document is sound (D-2.7.4's no-new-divergence finding: page-number
  substitution is integer advance arithmetic, the same mechanism on all four sizes), and is left
  as made; only the completeness of the claim above is corrected.
- **AC6** — the three pinned digests re-measured by the file route and unmoved; `git status
  --porcelain fixtures/` confirms zero other committed golden moved. Re-confirmed by the finisher
  after the Blocker 1/2 and Finding 4 fixes (see Delivery Log addendum): all three digests, and
  `page-count-20`'s own recorded hash, are unchanged.
- **AC7** — `page-count-20`'s golden accepted only after `qpdf` 12.4.0 (a reader this project did
  not write) resolved it: structurally valid, 20 pages. Reader, version and verbatim invocation
  recorded in `fixtures/page-count-20/README.md`, now with a `validated at` anchor (Finding 15).
- **AC8** — baseline preserved (same single expected failure, identical stats). The delta table
  above is corrected (Finding 5: the developer's "+17 top-level" was wrong, reconciled by a
  mechanism that does not exist; true developer delta is +10 top-level / +5 subtests, nothing
  lost) and extended with the finisher's own four new tests (+4 top-level / +4 subtests), closing
  at **359 top-level / 587 all-occurrences**, single expected failure unchanged.
- **AC9** — every enumeration in this record names `ecd0056` and the tree state it was taken
  against. Citations to `epics.md` (Finding 6) and to `internal/bind/text.go`'s fence (Finding 7)
  are corrected throughout this file and in the three shipped source comments that carried them.

**D-2.6.8's prediction (finding 10, story creation).** This story emits **no new indirect-reference
array** — the {{page}} slot is resolved entirely within existing `TextRun.Glyphs` slices (glyph
substitution, not a new PDF object class), so `writeRefArray`/`appendRefArray` gained no new
caller. The prediction did not fire; stated explicitly per its own instruction
(`[[measure-predicted-escalation-against-its-precedent]]`).

**Scope fence honoured.** No change to `Render`/`RenderTo`'s public signature; no diagnostic
channel added (Story 2.8's); `align` remains parsed-and-unhonoured; the three pinned digests were
not re-recorded; `epic-2: backlog` in `sprint-status.yaml` is untouched by this story (the gate
flips it, not this story).

---

## QA Results

## Review Summary

- **Reviewed by**: bmad-code-reviewer
- **Date**: 2026-08-25
- **Reviewed at**: working tree over baseline `ecd0056`, tree state as listed by `git status --porcelain`
  at review start and re-confirmed byte-identical at review end (every mutation below was reverted by
  hand and the restore verified by digest, never by `git checkout`).
- **Story Status Recommendation**: **Changes Requested**
- **Blockers**: 2
- **Majors**: 4
- **Minors**: 6
- **Nits**: 3

### What was MEASURED versus what was only READ (D-000.55's discipline, applied to this review)

**Executed, on `darwin/arm64`:**

- `go test ./... -count=1 -v` — **579 all-occurrences PASS / 1 FAIL, 355 top-level PASS / 1 FAIL**;
  the single failure is `TestCorpusMeetsP6ExerciseFloors` with stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, character-for-character the expected Story
  2.1 red. Run twice — once at review start, once after all mutations were reverted — with identical
  figures both times. **Confirms the orchestrator's independent measurement exactly.**
- `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .` — **PASS**.
- `shasum -a 256` on the three pinned digests, by the file route, twice — **all three unmoved**.
  `git status --porcelain fixtures/` shows no modification to any pre-existing fixture; only the four
  new `page-count-*` directories are added. **AC6 confirmed independently.**
- `qpdf --check` + `qpdf --show-npages` (qpdf 12.4.0) on **all nine** golden PDFs — all structurally
  clean; `page-count-20` resolves to 20 pages.
- **Nine behavioural probes** of input classes no committed fixture spans (empty content band, empty
  content text, two slots in one element, two slots in two elements, narrow-footer line-break
  reachability at six widths, data-injected tokens into both bands, header-only slot, whitespace-spelled
  tokens, adjacent tokens). Probe file written into the module, run, deleted; tree re-verified clean.
- **Four mutation red-proofs**, each applied to real production source, run, and reverted by hand with
  the restore confirmed by digest: (a) data binding silently disabled; (b) a `page` namespace added
  without a `lookupBound` call site; (c) `{{page.number}}`/`{{page.total}}` rendered against the
  mutated binder; (d) a `pageSlot`-named identifier planted in `internal/pdf`.

**Read, not executed**: the three deferred non-native matrix legs (`linux-amd64`, `linux-arm64`,
`js-wasm`). Read for teeth — `requirePageCount20HasCorrectPageNumbers` (`matrix_test.go:745`) reads
the literal `Page X of Y` off **every one of the 20 pages** through the document's own `/ToUnicode`
CMap and calls `requirePageTreeResolves(…, 20)`. It is a genuine feature guard, not a
`FontFile2`-presence check, and it can express a wrong page number. The legs are correctly deferred
per the cadence.

---

### Finding 1: Two `{{page}}` occurrences in one element — only the LAST resolves; the others render as the literal filler digits

- **Severity**: **Blocker**
- **Category**: Correctness / AC Conformance
- **Location**: `folio-go/render.go:735-752` (the slot loop in `positionSegments`); the single-slot
  field set it writes into is `folio-go/render.go:20-22` (`hasPageSlot`, `pageSlotGlyphLo/Hi`,
  `pageSlotDigitsY` on `textRunSource`); the matching single-append is `folio-go/render.go:518-524`.
- **Observation**: `resolvePageTokens` (`page_number.go:185`) correctly emits **one `pageSlotSpan` per
  `{{page}}` occurrence**. `positionSegments` then iterates them —
  `for _, sl := range slots { … run.hasPageSlot = true; run.pageSlotGlyphLo = … }` — but writes into a
  **single scalar field set on the run**. Every occurrence after the first overwrites the previous, so
  a run carries at most one slot. `collectBandTextRuns` likewise appends exactly one
  `pendingPageSlot` per run (`render.go:518`). Because `Page {{page}} of {{pages}} / {{page}}` is
  all-ASCII it is **one face segment on one line**, so both occurrences land in the same run and the
  first is silently lost.

  **Executed, not inferred.** Footer value `Page {{page}} of {{pages}} / {{page}}` on the existing
  two-page `reservedPlaceholderFooterTemplate`, read off the produced content stream through the
  document's own `/ToUnicode` CMap:

  ```
  page 1 drew: "Page 0 of 2 / 1"
  page 2 drew: "Page 0 of 2 / 2"
  ```

  The unresolved occurrence renders the reservation's **filler `0`** (`page_number.go:229`,
  `write(strings.Repeat("0", digitsY))`) — so it does not fail loudly, it prints a plausible-looking
  wrong page number, identically on every page.
- **Impact**: AC1's central property — *"X is the current page … correct on every page"* — is **false**
  for a template an author can plainly write. The failure is **silent**: no error, no diagnostic, and a
  document that reads `Page 0 of 2` on every sheet. On a document with `digits(Y) ≥ 2` it would read
  `Page 00 of 20`. For a project whose stated purpose is that *"a printed statement is verifiably
  complete"*, a footer that misreports the page number is the exact defect the feature exists to
  prevent. No test in the suite has two `{{page}}` occurrences in one element, so nothing catches it.
  A control confirms the neighbouring case is fine: two slots in **separate** footer elements resolve
  correctly (`Page 1 of 2` + `sheet 1`), which localises the defect precisely to the per-run field set.
- **Suggested Resolution**: carry slots **per run as a slice**, not as one scalar field set — i.e. make
  `textRunSource` hold `[]pendingPageSlot`-shaped data and let `pagemodel.TextRun` carry
  `[]*PageNumberSlot` (or have `resolvePageRunForPage` walk several slots), applying substitutions
  **right-to-left** so earlier glyph indices stay valid as the slice is rewritten. If multiple
  occurrences per element are instead to be **refused**, that is a disposition change and must be ruled,
  not chosen here — but it cannot stay a silent mis-render. Whichever way it goes, add the two-occurrence
  case as a red-proved subject; it is currently absent from the entire suite. **Do not fix by review.**
- **Related AC**: AC1 (and AC4, whose inverted test would have been the natural place to catch it).

---

### Finding 2: AC3's closure guard cannot express the violation AC3 names — a `page` NAMESPACE ships with the whole suite green

- **Severity**: **Blocker**
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/bind/resolution_roots_arch_test.go:44-62` (the AST scan),
  `:85` (`TestBindResolutionRootsAreClosed`), `:140` (`TestBindResolutionRootsClosureRedProof`);
  the declaration it defends is `folio-go/internal/bind/text.go:70`.
- **Observation**: The guard obtains its **observed** set by AST-scanning `text.go` for
  `lookupBound` call sites and reading argument index 4 as a string literal. That is a **proxy** for
  "the set of resolution roots", and D-000.15 forbids keying a guard on a proxy for its purpose.

  The story's own justification (`text.go:66-67`) asserts: *"A `page` NAMESPACE is precisely a THIRD
  entry here."* **That claim is false**, and the file two lines away shows why: `reservedPlaceholders`
  is handled at `text.go:153` by an **early return that calls no `lookupBound` at all**. A `page`
  namespace can be added in exactly that shape.

  **Executed red-proof.** I inserted into `BindTextSpans`, immediately above the `params` branch:

  ```go
  if path[0] == "page" {
      record(nil)
      continue
  }
  ```

  Result: `go test ./internal/bind/...` → **ok**. `go test ./... -count=1` → **the entire module green**,
  with the sole failure still the pre-existing Story 2.1 red. `TestBindResolutionRootsAreClosed` and
  `TestBindResolutionRootsClosureRedProof` both **passed**. A probe then confirmed the namespace is
  live, not cosmetic: a footer reading `N={{page.number}} T={{page.total}}` **rendered without error**,
  where the unmutated tree correctly rejects it. Mutation reverted; `text.go` restored to digest
  `498d719ac7a4de5487647638c6f30b5c20b579dcb73ff0cbb2a6727b516de73e`, byte-identical.

  Two further evasions, narrower but real: the scan hardcodes the filename `"text.go"`
  (`resolution_roots_arch_test.go:86`), so a third root introduced in any other file of package `bind`
  is invisible; and it hardcodes the function name `lookupBound`, so a differently-named helper trips
  nothing while the zero-call-sites precondition (`:97`) stays satisfied.
- **Impact**: AC3 is `epics.md:923-925` verbatim — *"no `page` namespace exists and none can be added"* —
  and the story's own AC3 requires *"a subject that can express the violation"* with an executed
  red-proof. The guard **does** redden for a third `lookupBound` **root**, which is what the shipped
  red-proof demonstrates; it does **not** redden for a `page` **namespace**, which is the thing AC3, the
  epic, and the fence comment all actually name. The story's Completion Notes state AC3 is met; on the
  AC's own wording it is met only for the narrower of the two violations. This is the same class the
  fence comment was written to prevent, and it is currently unguarded.
- **Suggested Resolution**: key the guard on the **property**, not the call sites — e.g. a behavioural
  assertion that a placeholder whose first path segment is `page` is rejected (today it is, incidentally,
  by the data-absence path), and/or a structural assertion that `BindTextSpans` contains exactly the
  declared set of first-segment dispatch branches. Then red-prove **the namespace shape above**, not
  only the extra-root shape, and record which guards stayed green. Widening the AST scan to the whole
  package and to any helper is worth doing but is not sufficient on its own — it would still miss the
  early-return shape. **Do not fix by review.**
- **Related AC**: AC3 (requirements 1 and 2).

---

### Finding 3: `folio-format.md`'s D-2.7.3 statement restates the MECHANISM after delivering the outcome

- **Severity**: **Major**
- **Category**: Convention (binding ruling)
- **Location**: `_bmad-output/specs/spec-folio/folio-format.md:498-501`
- **Observation**: D-2.7.3 (`folio-mvp-decision-log.md:7884-7887`) requires the statement be made *"as
  an **outcome** … and **not** as a mechanism — the correction this run has now made four times."*
  The first sentence complies and is close to the ruling's own parenthetical. Everything after the
  colon on line 499 does not:

  > *"…: the page number this construct needs is not independent of the content band's own layout there,
  > and this format does not negotiate a fixed point to resolve it (D-2.7.3)."*

  That is the ruling's mechanism paragraph restated — the layout dependency
  (`folio-mvp-decision-log.md:7873-7875`) and the fixed-point/AD-24-negotiation argument (`:7876-7877`).
  Placement is otherwise correct (beside the existing prose at `:494-496`) and nothing contradicts it.
- **Impact**: This is the **fifth** instance of a correction the log records as already made four times.
  The format specification is a reader-facing contract; carrying the engine's internal reasoning into it
  commits the project to that reasoning in a document that must survive an implementation change.
- **Suggested Resolution**: end the statement at *"naming the element."* and delete the remainder of the
  sentence. The `(D-2.7.3)` pointer may stay for provenance.
- **Related AC**: D-2.7.3's guardrail (carried under AC2/AC3's ruling set).

---

### Finding 4: Bound DATA values are re-scanned for `{{page}}`/`{{pages}}` — report data can inject a live construct, or break the render

- **Severity**: **Major**
- **Category**: Security / Correctness
- **Location**: `folio-go/render.go:397-402` (`collectBandTextRuns` calls `resolve` on `boundText`,
  i.e. on text **after** `bind.BindTextSpans` has already substituted data), consumed by
  `contentBandResolver` (`render.go:359`) and `headerFooterResolver` (`render.go:377`).
- **Observation**: `resolvePageTokens` and `firstReservedPageToken` scan the **post-binding** string. Any
  `{{page}}` that arrived *inside a substituted data value* is therefore indistinguishable from one the
  template author wrote. Executed, both directions:

  - Footer `{{label}}` with data `{"label":"Page {{page}} of {{pages}}"}` → renders **`Page 1 of 2`** /
    **`Page 2 of 2`**. A data value became a live pagination construct.
  - Content element `{{label}}` with data `{"label":"hello {{page}} world"}` → the whole document
    **fails to render**: `folio: Render: element e1: {{page}} resolves only in the page header and page
    footer bands (D-2.7.3) …`.

  The second is the more serious half: a *report data* value whose text happens to contain the literal
  `{{page}}` produces a **located template error**, blaming an element whose template is correct.
- **Impact**: `page_number.go:136-140` states the design assumption explicitly — *"text ALREADY PASSED
  THROUGH `bind.BindTextSpans`, so … the only tokens that can still appear are the two `internal/bind`
  leaves untouched."* That is false: bound values can reintroduce them. Consequences: (a) data controls
  template behaviour, which is the posture AD-14 exists to prevent for bound values; (b) a customer
  name or free-text memo containing `{{page}}` makes a statement unrenderable, with a diagnostic that
  names the wrong culprit; (c) `shiftSubstitutions`' stated invariant (`page_number.go:250-252`, *"A
  substitution can never OVERLAP a `{{page}}`/`{{pages}}` token"*) is also false in this case, so the
  substitution spans handed to `atomicSpansFor` can describe the wrong runes and mis-drive
  `unbreakableValues` line breaking.
- **Suggested Resolution**: resolve the page tokens against the **element's declared template text**,
  not against the post-binding string — i.e. have `internal/bind` report the reserved tokens' spans (it
  already knows them; it writes them back byte-for-byte at `text.go:153-159`) and let the resolver act
  only on those reported spans. That makes data-origin text structurally incapable of reaching the
  scanner and simultaneously repairs the `shiftSubstitutions` invariant. The disposition for a data
  value containing `{{page}}` then becomes the correct one: it is inert text.
- **Related AC**: AC1, AC2, and D-2.7.3's fence.

---

### Finding 5: AC8's itemised delta is wrong — "+17 top-level" against a measured +10, reconciled by a mechanism that does not exist

- **Severity**: **Major**
- **Category**: AC Conformance (AC8/AC9)
- **Location**: story file, Delivery Log delta table, lines 839-852
- **Observation**: The table's first row reads *"tests added | **+17 top-level**"*; its last row reads
  *"net top-level delta | **+10**"*; and the reconciling prose attributes the gap to *"the new sub-tests'
  internal assertions counting differently under `-v`'s `--- PASS` per `t.Run`"* and to *"2 fewer
  per-token sub-assertions"*.

  Measured directly, at the reviewed tree:

  - New top-level test functions across the four new test files: **exactly 10**
    (`page_number_test.go` 4, `tabular_digits_test.go` 2, `page_count_matrix_test.go` 2,
    `internal/bind/resolution_roots_arch_test.go` 2).
  - Top-level PASS: 345 → **355** = **+10**. Every new top-level test is accounted for; **no top-level
    test was lost.**
  - All-occurrences PASS: 564 → **579** = **+15**, i.e. **+5 subtests**. Measured attribution: `t.Run`
    appears exactly **once** in the new files (`page_count_matrix_test.go:77`, four fixtures → 4
    subtests), and the fifth is `TestEveryGoldenPDFResolvesItsPageTree/page-count-20`, produced by the
    new `goldenDigestRecord` entry flowing into `golden_structural_validity_test.go`'s existing loop —
    confirmed by grepping the `-v` output for subtests naming `page-count-20`, which returns exactly
    those two lines.

  So the true accounting is clean and simple: **+10 top-level, +5 subtests, nothing removed.** The
  reported "+17" is a subtest-inclusive figure carrying a *top-level* label, and the explanation is a
  category error — `--- PASS` lines count **subtests**, never assertions, so "fewer per-token
  sub-assertions" cannot move either figure.
- **Impact**: AC8 requires the delta be **itemised**, precisely because *"a net figure that happens to
  match while an assertion was lost is exactly what D-000.34 is about."* The itemisation is the
  deliverable and it is wrong; a reader reconciling +17 against +10 is sent looking for seven lost tests
  that do not exist. The record explains a real number with an invented mechanism, which is the failure
  mode D-000.49 names. I verified nothing was actually lost, so the **conclusion** stands — but the
  evidence offered for it does not.
- **Suggested Resolution**: replace the table with the measured accounting above, naming the commit and
  tree state per AC9, and drop the fabricated reconciliation sentence.
- **Related AC**: AC8, AC9.

---

### Finding 6: Every `epics.md` citation for this story is wrong, in the story file and in three new-code sites — and AC3's citation resolves to AC2's text

- **Severity**: **Major**
- **Category**: Convention / Maintainability
- **Location**: story file lines 10, 73, 81, 86, 421, 468, 488, 534, 553, 583, 634;
  `folio-go/page_count_matrix_test.go:3`; `folio-go/page_count_matrix_templates.go:3`;
  `folio-go/matrix_test.go:1177`
- **Observation**: Measured by `grep -n` against `_bmad-output/planning-artifacts/epics.md` at the
  reviewed tree. True line numbers:

  | element | true range | cited as |
  |---|---|---|
  | Story 2.7 section | **904-929** | `904-926` (story:10, story:73) |
  | AC1 (Given/When/Then) | **914-916** | `912-915` (story:534) |
  | AC2 (Given/When/Then/And) | **918-921** | `916-920` (story:553); `917-920` (story:421) |
  | AC3 (Given/When/Then) | **923-925** | `921-922` (story:583) |
  | AC4 (Given/When/Then) | **927-929** | `923-926` (story:81, 488, 634; and all three code sites) |
  | *"a footer containing a `Page X of Y` construct"* | **914** | `912` (story:468) |

  **Not one is correct.** The most damaging is AC3's: `epics.md:921-922` is *"**And** it is resolved
  between the passes by substituting pre-measured glyphs"* plus a blank line — **AC2's text, not AC3's**.
  A reader following that pointer to check AC3 verbatim lands on the wrong acceptance criterion.
  `epics.md:912` (cited at story:468) is the literal string `**Acceptance Criteria:**`.

  This also settles the citation error the brief flagged: `page_count_matrix_test.go:3` cites `923-926`
  and the orchestrator's brief cited `928-931`; **both are wrong** and the true range is **927-929**
  (`grep -n "documents of 1, 5, 20 and 50 pages"` → line 927).
- **Impact**: Position-bound citations to `epics.md` have already forced this project to re-cite by test
  name once. Three of these now live in **shipped source comments**, where they will be read long after
  the line numbers have moved again. The story's own correction section — which exists precisely because
  the brief truncated the AC list — cites the AC it rescued at the wrong lines.
- **Suggested Resolution**: re-cite by **quoted text or AC ordinal** ("Story 2.7's fourth acceptance
  criterion"), not by line range, in the three code comments at minimum; correct the story file's
  ranges to those measured above. Naming the commit alongside any range that must stay would satisfy
  D-2.6.9's correction, but the code comments are better served by a position-free citation.
- **Related AC**: AC9 (and the D-000.31 process note D-2.7.4 records).

---

### Finding 7: The `internal/bind/text.go:76` fence citation is stale in five story sites and wrong in two NEW code sites

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: story file lines 187, 197, 591, 616, 920; `folio-go/internal/bind/text.go:65`;
  `folio-go/internal/bind/resolution_roots_arch_test.go:12`
- **Observation**: The fence sentence (*"conflating the two is how 'page' would eventually acquire a
  namespace…"*) is at **`text.go:96`** in the reviewed tree. It was at 76 before this story; adding the
  `declaredResolutionRoots` block shifted it by ~20 lines. The story's five citations were therefore
  correct at creation and are now stale — including AC3.4's *"The fence comment at
  `internal/bind/text.go:76` is preserved"* and the Completion Notes' identical claim. More importantly,
  **two new code sites introduce the wrong pointer fresh**: `text.go:65` says *"the fence at line ~76
  below"* while standing 31 lines above a fence at 96, and `resolution_roots_arch_test.go:12` repeats it.
  The quoted text is verbatim-accurate in both; only the pointers are wrong.
- **Impact**: Same class as Finding 6, and self-inflicted: this story's own edit invalidated the pointer
  its own new comment writes. The fence is AC3's whole structural argument, so a pointer that lands 20
  lines short of it is the one citation most worth keeping resolvable.
- **Suggested Resolution**: in the two code comments, drop the line number and cite the fence by its
  quoted sentence or by the enclosing function (`BindText`'s doc comment). Update the story's five sites
  to 96, or likewise de-position them.
- **Related AC**: AC3.4, AC9.

---

### Finding 8: `collectTextRuns` and `passthroughResolver` have no production caller; two tests now exercise a path that does not ship

- **Severity**: Minor
- **Category**: Maintainability / Tests
- **Location**: `folio-go/render.go:313` (`collectTextRuns`), `folio-go/render.go:349`
  (`passthroughResolver`); the only remaining callers are `shaped_fixture_test.go:835` and
  `shaped_fixture_test.go:1415`
- **Observation**: The two-phase restructure replaced `renderDocument`'s single `collectTextRuns` call
  with three `collectBandTextRuns` calls. Grepped repo-wide: `collectTextRuns` is now invoked **only
  from tests**. `passthroughResolver` exists solely to serve it, and its doc comment
  (`render.go:345-348`) still describes it as *"BYTE-FOR-BYTE today's pre-Story-2.7 behaviour"* — true,
  but no longer the behaviour any document gets.
- **Impact**: Two fixture tests believe they are exercising the renderer's run-collection path; they are
  exercising an all-bands variant the renderer no longer uses. The orders happen to coincide today —
  `documentBands` indexes header=0, content=1, footer=2 (`render.go:1018-1020`) and `renderDocument`
  appends header, content, footer, so the run sequence is identical, which is exactly why the eight
  pre-existing goldens hold. That coincidence is now **load-bearing and unasserted**: if either order
  changed, the two tests would keep passing while every golden moved.
- **Suggested Resolution**: either re-point the two tests at the shipping composition, or keep
  `collectTextRuns` and assert explicitly that its output equals the header/content/footer concatenation
  `renderDocument` builds — which turns the coincidence into a checked invariant. Deleting it is also
  defensible, but not without giving the two tests a shipping subject.
- **Related AC**: AC6 (the mechanism that makes the non-movement claim true is untested).

---

### Finding 9: The phase split reordered error reporting for inputs no fixture covers

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/render.go:770-800` (`renderDocument`'s new prologue) versus the pre-story
  ordering visible in `git diff ecd0056 -- folio-go/render.go`
- **Observation**: Before this story, `renderDocument` collected **all** text runs first, then images,
  then geometry, then paginated. It now runs `documentBands` → `pageGeometryOf` → `collectImageRuns` →
  content-band text → `layout.Paginate` → header text → footer text. Four observable reorderings follow
  for documents with more than one defect:
  1. A geometry error now precedes any text error (previously the text error won).
  2. An image error now precedes any text error.
  3. A content-band text error now precedes a header- or footer-band text error (previously bands were
     reported in `documentBands` order, header first).
  4. **`layout.Paginate`'s `OverflowError` (D-2.6.5) now precedes header/footer text errors** — a
     document with both an overflowing content item and an unresolvable footer font chain reports the
     overflow where it used to report the font error.

  Additionally, `resolve` is now called **before** `fontChain` (`render.go:397` vs `:414`), so a
  content-band element with both a `{{page}}` token and a bad font chain reports D-2.7.3's error. I
  verified this does **not** regress the pre-existing QA Finding 5 ordering fix (font-chain validation
  still precedes the empty-text short-circuit at `render.go:414-431`).
- **Impact**: `Render` returns a different located error for the same input than it did at `ecd0056`.
  No golden covers a multi-defect document, so nothing measured it; all eight goldens are unmoved, which
  is a statement about single-outcome inputs only. This is behaviour change on the public API, currently
  undocumented and unasserted.
- **Suggested Resolution**: state the intended precedence in `renderDocument`'s doc comment and pin at
  least the Paginate-versus-band-text case with a test, so the ordering is a decision rather than a
  by-product of where the phase boundary fell. If any ordering is thought contractual, say which.
- **Related AC**: AC6's broader claim (bytes are unmoved; behaviour on multi-defect inputs is not).

---

### Finding 10: `positionSegments` panics on the public render path

- **Severity**: Minor
- **Category**: Maintainability / Correctness
- **Location**: `folio-go/render.go:739-746`
- **Observation**: A `{{page}}` reservation that straddles a face-segment or line boundary raises
  `panic(...)`, not an error. Everywhere else in this path a structural impossibility is a **located
  error with no public-signature change** (D-2.6.5's precedent; `digitTableRun` at `page_number.go:301`
  and `buildPageNumberSlot` at `page_number.go:343` both do exactly that, in the same file).
- **Impact**: `Render` is a library entry point; a panic crosses it uncaught. I **probed hard for
  reachability and could not reach it**: eleven renders across footer widths 1, 2, 3, 5, 6, 7, 8 and 10
  points at font sizes forcing 28-page documents (so `digits(Y) == 2`), both for
  `Page {{page}} of {{pages}}` and for a bare `{{page}}` element, all rendered cleanly — the breaker
  never splits the digit run, and the slot is all-ASCII so it cannot straddle a face segment. So the
  claim of unreachability holds against the shipped set, and this is a robustness issue rather than a
  live defect.
- **Suggested Resolution**: return the internal error instead of panicking, matching the two sibling
  checks in `page_number.go`. The function would need an error return, which is a local change — no
  public signature moves.
- **Related AC**: scope fence (*"no public-signature change"*) is honoured; the panic is the odd one out
  against this file's own posture.

---

### Finding 11: Two of the new tests assert only that *an* error occurred, not that the right one did

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/page_number_test.go:81`; `folio-go/page_number_test.go:189`; also
  `folio-go/page_number_test.go:64`
- **Observation**:
  - `TestPageSlotInContentBandPagesIsAlsoALocatedError` (`:74`) — the **entire** assertion is
    `if rerr == nil { t.Fatal(…) }`. Any render failure passes it: a parse error, a font error, a
    geometry error. It cannot distinguish D-2.7.3's fence from an unrelated fault.
  - `TestBuildPageNumberSlotFailsClosedOnNonUniformDigitAdvances` (`:178`) — asserts `err != nil` only,
    over a `pendingPageSlot` whose fields are arbitrary relative to the digit-table run it is given
    (`glyphLo: 5, glyphHi: 7` against a 10-glyph run). Correct **today**, because
    `buildPageNumberSlot` has exactly two error paths and the test satisfies the other one. If any
    bounds validation were added ahead of the uniformity loop, this test would stay green while the
    uniformity check it names silently stopped being exercised — D-000.34's shape.
  - `:64`'s `strings.Contains(rerr.Error(), "page")` is a near-free assertion in a codebase whose errors
    routinely say "page header", "page count", "page geometry".
- **Impact**: These are the only subjects for D-2.7.3's content-band fence and for D-2.7.2's fail-closed
  reservation. Their discriminating power is thinner than the properties they are credited with in the
  Completion Notes.
- **Suggested Resolution**: assert on the error's identity, not its existence — `Contains(err.Error(),
  "resolves only in the page header and page footer")` for the fence, `Contains(err.Error(), "advance")`
  for the uniformity check. Both are one-line changes.
- **Related AC**: AC2, and D-2.7.3's guardrail.

---

### Finding 12: The epic's fourth AC's *"hashes match recorded goldens"* clause is not satisfied for the 1-, 5- and 50-page documents — the judgment is sound, the claim is not

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `fixtures/page-count-1/`, `fixtures/page-count-5/`, `fixtures/page-count-50/` (each
  holds `input.folio` and nothing else); `folio-go/page_count_matrix_test.go:1-15`; story Completion
  Notes, AC1 bullet
- **Observation**: The brief asked this be **argued**, not assumed. Both halves:

  **The judgment is sound.** AC5 explicitly authorised the choice, and D-2.7.4 independently declined
  the D-000.4 override on the ground that page-number substitution is integer advance arithmetic on
  `geom.Length` with no float, vendor call, compressor or new dependency. If that reasoning holds — and
  I see no reason it does not — then all four documents exercise the **same** divergence-free mechanism,
  and three more matrix legs would be three more measurements of one property. `page-count-20` is
  correctly chosen: it is the smallest of the four that spans the page-9-to-page-10 boundary, which is
  the case D-2.7.2's reservation actually turns on. Registering four documents would cost gate time and
  buy nothing D-2.6.2's criterion asks for.

  **The claim is nonetheless false as written.** `epics.md:929` says *"hashes match recorded goldens on
  all four targets."* For three of the four documents **there is no recorded golden at all** — no
  `expected.pdf`, no `expected.json`, no digest in `goldenDigestRecord`. A clause about hashes matching
  goldens is not satisfied by a document that has no golden; it is not weakly satisfied or satisfied by
  a substitute, it is inapplicable. What `TestPageCountFixturesRenderCorrectPageNumbersEverywhere`
  verifies is a **different property**: that the rendered text is *correct today on one target*. That is
  the right property to check, but it is not byte-stability, and no single-target behavioural test can
  be. Concretely: a change to CID allocation order that altered `page-count-50`'s bytes without altering
  its recovered text would be caught for `page-count-20` and invisible for the other three.

  Second-order: the three `input.folio` files are pinned only against the in-repo constants
  (`TestPageCountFixturesMatchTheInRepoTemplates`), and the constants are what the tests actually
  render. The on-disk fixtures therefore carry no independent verification of anything — they are
  copies kept in step with their originals.
- **Impact**: Small in risk, real in record-keeping. The Completion Notes state *"All nine acceptance
  criteria are met"* and that AC1 is covered *"for all four page counts (1/5/20/50)"* without
  qualification. The precise statement is: AC1's **correctness** property is covered for all four; the
  epic's fourth AC's **hash** clause is covered for `page-count-20` only, by design and with
  authorisation.
- **Suggested Resolution**: say so plainly in the Completion Notes rather than leaving the unqualified
  claim, and record the reasoning (D-2.7.4's no-new-divergence finding) as the ground for why one
  document suffices. Consider whether the three golden-less fixture directories earn their place on
  disk at all, given the constants are the rendered subject.
- **Related AC**: AC5's note, and the epic's fourth AC.

---

### Finding 13: The tabular-digits guard iterates a hardcoded three-face list, not the FontSet

- **Severity**: Nit
- **Category**: Tests
- **Location**: `folio-go/tabular_digits_test.go:61-62`
- **Observation**: `fs := testShippedFontSet()` is obtained and then indexed by the literals
  `{"Noto Sans", "Noto Sans Thai", "Noto Sans SC"}`. Face **removal** is caught (`:29` fatals when a
  named face is absent); face **addition** is not — a fourth shipped face would never be checked for
  digit uniformity.
- **Impact**: Bounded. D-2.7.2's stated residual hazard is *"the face set is not frozen"*, and the one
  way the set changes without the guard noticing is by growing.
- **Suggested Resolution**: range over the FontSet's own keys, keeping a non-empty precondition.
- **Related AC**: AC2's width half / D-2.7.2.

**Recorded in this guard's favour, because it is the strongest artifact in the story and the brief asked
specifically:** `TestUniformDigitAdvanceGuardRedProof` (`tabular_digits_test.go:83`) is a **genuine,
executed red-proof, not a forward guard**. It shapes the real face to find `'3'`'s glyph id, byte-patches
that glyph's advance to 900 in a **scratch copy of the real embedded `hmtx` table** (`:94`, `:162`),
guards against a mis-aimed patch (`:108` fatals if the patch did not take), and then drives the **same**
`digitAdvancesForFace` + `allEqual` helpers the green test drives — with `fontset.New` confirmed to do no
name-keyed caching, so the patched bytes really are what is measured. D-2.7.2 said a red-proof was
available; one was built and it is the real thing. No false forward-guard credit here.

---

### Finding 14: The Delivery Log's native-leg entry says "the three pinned digests" and then names two

- **Severity**: Nit
- **Category**: Convention (AC9)
- **Location**: story file lines 808-810
- **Observation**: *"including the three pinned digests (`shaped-text`, `multi-page`)"* — two are named.
  The third, `fixtures/expected-breaks/expected_breaks.json`, is **not a matrix document** and could not
  have been covered by `TestTargetRenderHash`; `matrixDocuments` holds nine entries and
  `expected-breaks` is not among them.
- **Impact**: The claim over-reaches slightly: the native leg verified two of the three pinned digests,
  not three. AC6 is nonetheless satisfied — I re-measured all three by the file route and all are
  unmoved.
- **Suggested Resolution**: name the two, and note that `expected-breaks` is covered by AC6's direct
  measurement rather than by the matrix leg.
- **Related AC**: AC6, AC9.

---

### Finding 15: `page-count-20`'s provenance has no "validated at &lt;commit&gt;" anchor

- **Severity**: Nit
- **Category**: Convention
- **Location**: `fixtures/page-count-20/README.md:34-58`
- **Observation**: All four D-000.53 elements are present, specific, and **verified**: reader (`qpdf`,
  with absolute path), version (`12.4.0`), verbatim invocation (both commands), and output (both pasted
  in full). I re-ran both commands and the pasted output reproduces exactly; the digest in the README
  matches `expected.json` and `byte_neutrality_test.go:170`. The sibling `fixtures/multi-page/README.md`
  additionally carries a `| validated at | 50ad6c8 (Story 2.6) |` row; this one does not, so nothing in
  the file states which bytes the pasted output was taken against.
- **Impact**: None today. D-000.53 does not require the anchor. But D-000.44 makes every re-recording owe
  the step afresh, and the commit anchor is what makes a stale paste detectable.
- **Suggested Resolution**: add the row, matching the sibling fixture's shape.
- **Related AC**: AC7.

---

## AC-by-AC disposition

| AC | disposition |
|---|---|
| **AC1** — X/Y correct on every page | **NOT met.** Correct for a single `{{page}}` per element (verified on all four page counts, every page, off the content stream). **Finding 1**: a second occurrence in the same element renders the filler digit on every page. |
| **AC2** — late-bound, pre-measured, resolved between the passes | **Met, structurally.** Slot lives in `internal/pagemodel` and names no PDF concept (`CID` has precedent on `ShapedGlyph`); resolution is in `paginateDocument`'s per-page loop, between the passes; `internal/pdf` is guarded and **I red-proved the sweep fires** — planting a `pageSlot`-named identifier in `internal/pdf/textdoc.go` reddens `TestInternalPDFReachesNoLayoutComputation` with a named witness. Width behaviour per D-2.7.1/D-2.7.2 is **correct**: I verified from `internal/pdf/textdoc.go:799-800` that a run's consumed advance is `Σ XAdvance` independent of `XOffset`, so the slot consumes exactly `DigitsY × DigitAdvance` and the digits are genuinely right-aligned with the literals unmoved. **D-2.7.1 honoured** — no position derivation was smuggled into `internal/pdf`. See Findings 10, 11 for quality. |
| **AC3** — no `page` namespace, none can be added | **NOT met.** **Finding 2**, with an executed red-proof: a `page` namespace ships with the entire suite green. The declared/observed comparison is genuinely bidirectional and its vacuity guards are real; the defect is that its subject is `lookupBound` call sites rather than the property. |
| **AC4** — inverted tests, nothing silently disarmed | **Met.** The inversion is correct and asserts off the content stream. **The negative half survives and still reddens — verified by mutation, not by reading**: with data binding silently disabled (`lookupBound`'s error swallowed), `TestReservedPlaceholderSetIsUnchanged` **FAILED** while `TestReservedPagePlaceholdersResolveOnEveryPage` and `TestPageCountFixturesRenderCorrectPageNumbersEverywhere` both stayed **green** — which is precisely the gap D-000.34 predicted and the reason the negative half had to survive. Mutation reverted, `text.go` restored byte-identically. |
| **AC5** — matrix document registered, one gate line | **Met.** Exactly one line added to `declaredEpic2GateObligations` (`byte_neutrality_test.go:495`), naming Story 2.7 and D-2.7.4. `matrix.yml` updated at **all five** slug sites (four per-target upload paths + the `docs=` list). D-000.54's native leg **run and re-run by me — PASS**. Both halves of its guardrail stated. Logged as a **sequencing fix, explicitly not a cadence override**, correctly. See Finding 12 on the 1/5/50 judgment. |
| **AC6** — three pinned digests do not move | **Met, independently confirmed.** All three reproduce exactly; `git status --porcelain fixtures/` shows zero modification to any pre-existing fixture; all nine golden PDFs are `qpdf --check` clean. **No pending owner sign-off is invalidated.** |
| **AC7** — independent-reader acceptance | **Met.** All four D-000.53 elements present and re-verified by re-running both commands. See Finding 15 (Nit). |
| **AC8** — baseline preserved, delta itemised | **Partially met.** Baseline preserved: 579/1 and 355/1, single expected red with matching stats, reproduced twice. **Finding 5**: the itemisation is wrong and its reconciliation fabricated. |
| **AC9** — every enumeration names its commit | **Partially met.** The digest and suite enumerations name `ecd0056` and their tree state. **Findings 5, 6, 7, 14**: several enumerations and citations are inaccurate. |

**Binding rulings — compliance:**

- **D-2.7.1** — honoured. The box is fixed at pass one; option (e) was not implemented; no advance
  arithmetic landed in `internal/pdf` (red-proved).
- **D-2.7.2** — honoured. Reservation is `digits(Y)`, right-aligned, no zero-padding, only `{{page}}`
  reserved. The fail-closed non-uniform-advance path exists and the tabular guard is a **proven**
  red-proof, correctly keyed on purpose.
- **D-2.7.3** — **partially honoured.** The content-band fence works (verified, including the
  whitespace spelling `{{ page }}`) and is red-proved on a **real template**. **No DW entry was filed** —
  confirmed by searching the whole deferred-work register and the repo; `deferred-work.md` is untouched
  and there is no `DW-17`. The `align` item is only *proposed*, is unfiled, and is a distinct subject.
  **Finding 3**: the `folio-format.md` statement leaks mechanism.
- **D-2.7.4** — honoured. Gate to six, one line, override correctly declined, native leg owed and run.
- **D-000.55** — honoured. The banned phrase appears only where the story quotes the ruling itself,
  never as a claim. What was executed is named by target; the three unrun legs are named individually.
- **Scope fence** — honoured. `epic-2: backlog` is **still present and unchanged** at
  `sprint-status.yaml:49`; the only sprint-status changes are the date and this story's own status.

---

*Reviewer's note on tree state: four mutations were applied to production source during this review
(`internal/bind/text.go` twice, `internal/pdf/textdoc.go` once, plus two throwaway probe test files).
All were reverted by hand and the restoration verified by digest and by re-running the full suite to
the identical 579/1 and 355/1 figures. `git status --porcelain` at review end is identical to review
start. No production code, test, or fixture was changed by this review.*

---

## Finding Resolutions (finisher)

All 15 findings triaged: **15 FIX, 0 DISMISS, 0 DEFER.** Every finding named a concrete, in-scope,
cheaply-verifiable defect in this story's own delivery — none required expanding scope beyond the
story's own ACs, and none was a stylistic preference or a duplicate. Both Blockers were red-proved
live, in this session, against the reviewer's own exact mutations (see the Delivery Log addendum
above for the transcripts).

| # | Sev. | Title | Decision | Rationale | Files changed |
|---|---|---|---|---|---|
| 1 | Blocker | Two `{{page}}` occurrences in one element — only the last resolves | **FIX** | A live correctness bug with a silent, plausible-looking wrong output — the worst failure class for this feature. Disposition: **support N slots per run** (option (a) of the two the review offered), not rejection — nothing in AD-4/epics.md forbids more than one occurrence per element, and the semantics are perfectly well-defined (each occurrence independently shows the current page), so treating it as a located error would invent a restriction the epic never asked for. The adjacent panic (straddling reservation) is converted to a located error in the same pass, per D-2.6.5's precedent, closing Finding 10 at the same time. | `internal/pagemodel/pagemodel.go`, `render.go`, `page_number.go`, `page_number_test.go`, `reserved_placeholders_test.go` |
| 2 | Blocker | AC3's guard is keyed on a proxy (`lookupBound` call sites), not the property | **FIX** | D-000.15 exactly. Added a BEHAVIOURAL guard (`TestBindTextDottedPageAndPagesPathsAreOrdinaryDataPaths`) that asserts the actual observable property AC3 names — no code shape can evade an assertion on the OUTPUT. Kept the existing structural guard as a narrower, secondary check (real, just insufficient alone) and widened its file scope to close its first evasion (hardcoded `text.go`). | `internal/bind/text_test.go`, `internal/bind/resolution_roots_arch_test.go` |
| 3 | Major | `folio-format.md`'s D-2.7.3 statement leaks mechanism after delivering the outcome | **FIX** | D-2.7.3 explicitly requires the outcome framing and calls this "the correction this run has now made four times" — unambiguous, cheap, no judgment call. | `_bmad-output/specs/spec-folio/folio-format.md` |
| 4 | Major | Bound data values are re-scanned for `{{page}}`/`{{pages}}` | **FIX** | Real security/correctness gap (AD-14's "never a coercion" posture, and a wrong-culprit diagnostic on ordinary report data). Fixed by excluding any `{{page}}`/`{{pages}}`-shaped occurrence that overlaps a `bind.Substitution`'s own span from being treated as reserved — a genuine reserved token, by construction, never overlaps a Substitution (reserved tokens produce none), so this cannot false-negative on real template text. | `page_number.go`, `render.go` |
| 5 | Major | AC8's itemised delta is wrong ("+17" reconciled by an invented mechanism) | **FIX** | The developer's own conclusion (nothing was actually lost) was right; the evidence for it was fabricated. Replaced with the measured accounting (+10 top-level / +5 subtests) and extended with this finisher's own additions. | story file |
| 6 | Major | Every `epics.md` citation is wrong, including AC3's (which resolves to AC2's text), in the story and in three shipped source comments | **FIX** | Mechanical, unambiguous once re-measured by `grep -n`. Story-file prose citations corrected to the true ranges; the three shipped-source citations re-cited by AC ordinal rather than line range (per the finding's own suggested resolution and this project's prior experience re-citing by name after a position-bound citation restaled). | story file, `page_count_matrix_test.go`, `page_count_matrix_templates.go`, `matrix_test.go`, `fixtures/page-count-20/README.md` |
| 7 | Minor | `internal/bind/text.go:76`'s fence citation is stale in five story sites and wrong in two new code sites | **FIX** | Same class as Finding 6, self-inflicted by this story's own edit. De-positioned all seven sites (cite by enclosing function/doc comment, not line number) rather than hardcoding the new line (96), since a hardcoded line number is exactly what went stale the first time. | story file, `internal/bind/text.go`, `internal/bind/resolution_roots_arch_test.go` |
| 8 | Minor | `collectTextRuns`/`passthroughResolver` have no production caller; the order coincidence that keeps their two remaining test callers valid is unasserted | **FIX** | Chose "assert the coincidence explicitly" over "delete" or "re-point the two callers": deleting would lose the two shaped-fixture tests' actual coverage, and re-pointing them individually would not name the invariant the review actually flagged. One new test (`TestCollectTextRunsMatchesTheShippingBandComposition`) turns "documentBands order == renderDocument's append order" into a checked property. | `collect_text_runs_composition_test.go` |
| 9 | Minor | The phase split reordered error reporting on multi-defect inputs, undocumented | **FIX (documentation only)** | No AC and no golden fixes an ordering among simultaneous, unrelated defects, and none of the eight pre-existing goldens (single-outcome inputs) can observe it — so this is a genuine implementation detail, not a silently-broken contract. Declared explicitly non-contractual in `renderDocument`'s own doc comment, per the finding's own "if any ordering is thought contractual, say which" — the answer is: none is. A pinning test was considered and rejected as pinning behaviour the story never promised. | `render.go` |
| 10 | Minor | `positionSegments` panics on the public render path | **FIX** | Folded into Blocker 1's fix, since both touch the same function's slot loop. Converted to a located error naming the run's rune range, matching every sibling check in `page_number.go` and D-2.6.5's precedent; no public-signature change (the scope fence's own requirement). | `render.go`, `page_number.go` |
| 11 | Minor | Two new tests assert only that *an* error occurred, not the right one | **FIX** | Three one-line changes: assert on D-2.7.3's fence phrase (both content-band tests) and on "advance" (the non-uniform-advance test), rather than presence alone. | `page_number_test.go` |
| 12 | Minor | The epic's fourth AC's "hashes match recorded goldens" clause is unmet for three of four page-count documents; the Completion Notes claimed it unqualified | **FIX (disclosure only)** | The review's own judgment (one matrix document suffices, D-2.7.4's no-new-divergence argument) stands — re-verified, not re-litigated. Per D-000.17, the unmet portion of the clause is now stated plainly in the Completion Notes rather than folded into an unqualified "AC5 met." The three golden-less fixture directories are left in place (their `input.folio` still earns its keep as a pinned-against-template-constant, behaviourally-verified subject); deleting them was considered and rejected as unrelated churn on a finding about a claim's precision, not the fixtures' existence. | story file |
| 13 | Nit | Tabular-digits guard iterates a hardcoded three-face list, not the FontSet | **FIX** | One-line-class fix exactly as suggested: range `fs`'s own sorted keys (D-1.3.5/ScanMapRange), with a non-empty precondition. | `tabular_digits_test.go` |
| 14 | Nit | Delivery Log says "three pinned digests" and names two | **FIX** | Corrected the wording; noted `expected-breaks` is covered by AC6's own direct measurement, not the matrix leg (it is not a matrix document). | story file |
| 15 | Nit | `page-count-20`'s provenance has no "validated at" anchor | **FIX** | Added, matching `fixtures/multi-page/README.md`'s sibling row; cites `ecd0056` (the baseline commit both the review's and this finishing session's `qpdf` runs were taken against) rather than this story's own not-yet-existing commit hash (this project has been burned once by a story trying to cite its own commit before it existed). | `fixtures/page-count-20/README.md` |

**What was confirmed and deliberately left untouched** (per the review's own "do not disturb"
list): the tabular-figures guard's red-proof (`TestUniformDigitAdvanceGuardRedProof`), the
negative half of `TestReservedPlaceholderSetIsUnchanged`, `page-count-20`'s D-000.53 provenance,
D-000.54's native leg logged as a sequencing fix, D-000.55 compliance, the absence of a DW entry
for content-band pagination, `epic-2: backlog`, and all eight pre-existing goldens' byte-identity.
None of these needed a finding filed against them, and none was touched by this finishing pass.
