# Story 4.7: The golden report renders and hashes match

Status: done

Epic: 4 — Tables and pagination · **this story is the C4 gate**
Sprint key: `4-7-the-golden-report-renders-and-hashes-match`
Baseline: `df8cbcc` (Story 4.6 finisher), tree clean at creation.
Covers: S1, S2, S3, S4 · AD-21, AD-22, AD-1, AD-5, AD-14, AD-23
Governing decisions read: D-4.7.0, D-000.22, D-000.44, D-000.53, D-000.54, D-000.26 (refined),
D-000.41, D-000.85, D-000.86, D-000.88, D-000.89, D-000.4 (override criterion), D-4.5.5, D-000.87,
D-000.21, D-000.37, D-2.3.5, D-2.5.1, D-2.6.6, DW-11, DW-13, DW-14, DW-21

---

## In plain terms (read this first if you just want the gist)

This story ships the product's central proof: a real customer account statement — logo, customer and
account details, a five-column transaction table whose header repeats on every page and whose total
sits at the foot, and a page footer with a supplied date and a page number — rendered at four different
lengths. The 1-, 5-, 20- and 50-page results are byte-for-byte identical on every target platform, and
Panit Wechasil has visually inspected the rendered pages and signed the exact four current digests.

What happened after the code review is the part worth knowing. The money column — the one column the
project's whole "never do money arithmetic the way computers do fractions" rule exists for — was built
from amounts a computer doing it the wrong way would have got right by luck. The documents could not
have caught the very mistake they were created to catch. Every amount was changed and all four
documents re-recorded before anyone was asked to approve them, which is why the approval was
deliberately held back until that landed.

The adversarial review found several checks weaker than their descriptions. One claimed to prove three
scripts sat inside a single table cell and only checked the row. One would have accepted a Thai mark
nudged the wrong way. One pointed at a cross-check nobody had written. All sixteen findings were fixed,
with no dismissals or deferrals. A shared helper that quietly returns confident nonsense on documents
like this one now refuses to run rather than answering wrongly, and the signed approval states that the
reader inspected the rendered pages rather than extracted text.

What will look wrong later and is intentional: one corpus-floor test remains red by design and belongs
to the separately owned opaque-name work, while several slow tests run only when explicitly enabled.
The semantic sign-off gate now passes and will turn red again if any signed PDF changes. Two side
questions were settled rather than left hanging: compressing embedded fonts is not being adopted,
because of future toolchain-upgrade determinism rather than file size; and a long-standing worry about
a text-mapping limit is closed with a measurement and an explanation of what would actually trigger it.

---

## Story

As a solo builder,
I want the Customer Account Statement to render correctly and reproducibly at four page counts,
So that the MVP's central claim is demonstrated rather than asserted.

---

## Premise verification (D-000.81 / this story's creator as reviewer of its own brief)

Everything below was measured at `df8cbcc` with read-only probes; the tree is clean. **Seven premises
confirmed, one corrected, four extended — and two of the extensions are obligations the brief omits
entirely and that land inside this story by their own terms.**

### Confirmed

1. **The story key is `backlog`.** `sprint-status.yaml:97`. `4-4`, `4-5`, `4-6` are `done` (`:94`,
   `:95`, `:96`). No gap, no out-of-order creation. Explicit mode was correct.
2. **No committed golden contains a table.** `grep -l '"table"' fixtures/*/input.folio` returns
   **nothing**. Nine `expected.pdf` files exist, in exactly the nine directories D-4.7.0 names.
   D-4.7.0's central measurement holds as written.
3. **The circularity is real and it is structural, not rhetorical.** Each fixture is compared by a
   bespoke `*_fixture_test.go` that re-renders and byte-compares against the committed
   `expected.pdf`, cross-checked against `expected.json`'s `sha256` and against a third literal in
   `goldenDigestRecord` (`byte_neutrality_test.go:92`). Three sites, one value, all of it recorded by
   this story. Nothing in that construction can distinguish a right document from a wrong one.
4. **Cross-target identity proves determinism only.** `matrix_test.go` captures four targets and
   compares. `golden_structural_validity_test.go:11-45` already says this in the project's own words,
   from the multi-page incident: *"A deterministically wrong file is byte-identical to itself."*
5. **D-000.89's figures are exact at HEAD.** Reproduced below under *The measurement*.
6. **AD-22's concern applies.** Any change moving already-recorded bytes is a versioned behaviour
   change, not an implementation detail. This story is scoped so that no existing golden moves, and
   that is asserted (AC11).
7. **DW-21 is open and this story does not close it.** `deferred-work.md:183`. Its owner is the Epic 4
   boundary gate **and** the orchestrator's checklist, two addresses on purpose. Stated below.

### Corrected — the "46 table behaviour tests" figure is a mutation blast radius, not the size of the suite

D-4.7.0 says the 4.6 mutation *"reddened 46 top-level tests, all in the table behaviour suite"*, and
the brief adds *"plus 11 subtests"*. Both are descriptions of **what one mutation reddened**. Neither
is the size of the second producer this story is required to protect.

**Measured, at `df8cbcc`, counting top-level `func Test…` in package `folio`:**

| file | top-level tests | story |
|---|---|---|
| `folio-go/table_render_test.go` | 17 | 4.1 |
| `folio-go/table_render_row_test.go` | 12 | 4.2 |
| `folio-go/table_pagination_test.go` | 8 | 4.3 |
| `folio-go/table_header_repeat_test.go` | 14 | 4.4 |
| `folio-go/table_footer_test.go` | 28 | 4.5 |
| `folio-go/table_row_clip_test.go` | 11 | 4.6 |
| **total** | **90** | |

Plus 25 in supporting internal packages (`internal/layout/table_test.go` 5,
`internal/layout/paginate_group_test.go` 9, `internal/template/table_geometry_test.go` 6,
`internal/template/footer_test.go` 5).

**Why the correction matters rather than being pedantry.** D-4.7.0's obligation 2 is *"the table
behaviour tests are NOT superseded."* A guard written against "46" would protect a number that was
never the population. **The population is six named files, and the guard is written against the set of
names (D-2.5.1), never a count.**

### Extended — obligation 1 is not a design problem; D-000.22 and D-2.3.5 already specify it, and the mechanism is built and running twice

The brief asks this story to *design* a semantic acceptance step that is "performable and auditable"
and to avoid "a checkbox nobody can fail". **That work is done and ratified.** Per the standing note
against framing a fork where a ratified rule decides, this story **applies** the existing form rather
than inventing one:

- **D-000.22** (binding) — a golden's first recording must assert a semantic property read off the
  artifact, separately from the hash. *"A hash tells you nobody moved the furniture. It does not tell
  you the furniture was ever in the right room."*
- **D-000.44** — that duty is owed on every re-recording, not only the first.
- **D-000.53** (binding) — the golden is not accepted until **a reader this project did not write**
  resolves it into the objects it claims to contain, with the reader, its version and the **verbatim
  invocation** recorded in provenance.
- **D-2.3.5 / D-000.26 (refined)** — the irreducibly-human half is recorded as a **JSON record naming
  the reader, the date, what was examined, and the digest signed**, enforced by a **failing test**,
  and the digest binding makes a re-record invalidate it automatically.

**The shipped instance is `folio-go/shaped_signoff_matrix_test.go`** — a red gate whose failure message
is written for the person who must resolve it (D-000.37), whose schema rejects an empty field because
*"an unattributed, undated or unspecific sign-off is indistinguishable from no sign-off at all"*, and
whose sibling `assertSignOffMatchesFrozenHash` cancels the approval when the digest moves. A second
instance exists for break placement (`expected_breaks_signoff_matrix_test.go`,
`fixtures/expected-breaks/break-signoff.json`). **This story copies that mechanism; it does not
redesign it.** The one thing genuinely open is the record's *shape* for four goldens at once, framed
as a fork below.

### Extended — the brief omits DW-13, and DW-13 is scheduled to run INSIDE this story

`deferred-work.md:800-806`, verbatim: *"**The sizing runs during Story 4.7**, which is the first
document in the programme with real CJK content in volume (its own AC requires 'Latin, Thai and CJK
text in the same table' at 1, 5, 20 and 50 pages) — that is the measurement this entry has been asking
for, and 4.7 produces it as a by-product whether or not anyone asks."*

The brief does not mention DW-13. **It is this story's, by the deferred-work register's own words.**
Scope is precise and small: **measure**, do not adopt. The adoption decision is the project owner's and
is batched with the other Epic 4 close decisions. Adopting compression here would move every golden and
require retiring a lint rule that has been proved to fire (`lint/internal/rules/nocompressor.go`) — out
of scope on both counts.

### Extended — the brief omits DW-11, whose owner window names this story's work

`deferred-work.md:619-626`. Owner: *"Epic 2's later stories **and Epic 4's golden-report work**"*.
Its stated forgetting-signal is exact: *"S4 still carrying only 2 genuinely-uncoverable,
independently-attested opaque items **when Epic 4's golden report ships**."* Stories 2.4 and 2.6 each
answered in writing without adding items. **This story owes the same written answer**, and D-000.17
forbids inventing attestation to move the number.

Related and separate: `TestCorpusMeetsP6ExerciseFloors` is red by design at P6g 7/20. **This story does
not fix it** — it is the same shortfall DW-11 tracks, and filling it by invention is precisely what
D-000.17 bans.

### Extended — this story is a D-000.4 matrix override, and it must run its four legs in-story

`matrix_test.go:1120-1123` and `:1136-1139` both name the override list in their own comments:
*"1.2, 1.5, 1.8, 2.4 **and 4.7**"*. The 2026-08-26 refresh confirms it: *"**4.7 remains a named
override**"*. So the four Docker/Node legs are **this story's**, not the Epic 4 gate's. D-000.54
independently requires a newly-registered matrix document to run its native leg at registration.

---

## The measurement that shapes the implementation

**Exact invocation** (D-000.89), run at `df8cbcc` from `folio-go/`:

```
env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...
```

| population | pass | fail | skip | total |
|---|---|---|---|---|
| top-level | 711 | 1 | 4 | **716** |
| including subtests | 1088 | 2 | 4 | **1094** |

**Skips, as a set by name** (D-000.81 — never a count):

- `TestXrefEntriesRejectsMalformedSubprocess`
- `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`
- `TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`
- `TestTwoTablesWithPageCountFooterRenderConsistently`

**Red by design, named separately**: `TestCorpusMeetsP6ExerciseFloors` — subtest `P6g (opaque names)`,
`got 7, need >=20`. Full stats line, read off the run: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115
P6g:7}`. The other six floors pass.

**These figures match the brief exactly.** They are stated here so the developer re-derives them at its
own baseline rather than inheriting them.

### What exists, and what does not

Every mechanism this story needs is built. **None of the artifacts is.**

| needed | at HEAD |
|---|---|
| a `.folio` fixture containing a table | **none** — zero fixtures, in fifteen directories |
| repeated header / sum footer / row-atomic pagination / over-tall clip | built (4.3–4.6) |
| image embedding, `Page X of Y`, `params`-supplied values | built (1.8, 2.7, 1.7) |
| Latin+Thai+CJK fallback | built, and fixture-proven **outside** a table only |
| three scripts **inside one table** under one fallback stack | **no test and no fixture** — the table suites use one single-face `fontFamily` per subtest |
| frozen Thai break labels | `fixtures/expected-breaks/expected_breaks.json`, human-signed |
| data-driven page count | **no precedent** — `fixtures/page-count-{1,5,20,50}` are Story 2.7's all-Latin marker documents whose page count is a **geometric construction**, and only `page-count-20` carries a golden |
| clock ban | enforced statically — `lint/internal/rules/forbiddenimports.go` bans `time`; `grep -rn 'time.Now'` over the repo returns **one** hit, the lint rule's own red-proof literal |

**The `.folio` table shape**, from the serializer golden `folio-go/testdata/template/golden/worked-example.json`
(byte-pinned into `folio-format.md`): a `"type":"table"` element with `bind`, `as`, `headerHeight`,
`columns[]` (`id`/`label`/`width`/`align`/`bind`, optional `footer` ∈ `sum|count|avg`, `footerOf`,
`footerFormat`), and `altRowBackground` — **declared in the schema, applied by nobody; Story 4.8 owns
it.** This story must not set it.

### Recording a golden touches FIVE hand-maintained registries, and two of them redden in the ordinary suite

Measured at `df8cbcc`. **There is no `-update` flag, no env var and no make target** — `grep flag.Bool`
over `folio-go/` returns nothing, and re-recording is a deliberate, attributable act governed by the
`goldenDigestRemedy` constant (`byte_neutrality_test.go:490`): *"DO NOT UPDATE A DIGEST TO MAKE A TEST
GO GREEN."* In the order the guards fail:

| # | edit | guard that reddens until it is made |
|---|---|---|
| 1 | `fixtures/<slug>/input.folio` **and** a byte-identical Go constant | the per-fixture template-match test (`page-count-*`'s precedent) |
| 2 | a `capture<Name>Render` func + a `FOLIO_SUBPROCESS_<NAME>` seam (`render_test.go` `TestMain`, ~`:476-600`) | matrix capture |
| 3 | the semantic acceptance assertions, written **before** the digest is frozen | D-000.22 / D-000.44 |
| 4 | `qpdf --check` and `qpdf --show-npages`, output pasted verbatim into the README | D-000.53 |
| 5 | `expected.pdf` + `expected.json` (`folioGoVersion`, `goToolchain`, `sha256` — 64 lowercase hex, a string, never a per-target map) | `checkFixtureShape` |
| 6 | `README.md`; **if it quotes the digest, that is a declared `{kind:"readme"}` site** | the completeness walk |
| 7 | `goldenDigestRecord` entry with the **second literal** (`byte_neutrality_test.go:92`) | `TestGoldenDigestAgreesAtEveryDeclaredSite` |
| 8 | a per-fixture `Test<Name>GoldenFixture` | — |
| 9 | `matrixDocuments` entry (`matrix_test.go:1088`) | matrix legs |
| 10 | **`.github/workflows/matrix.yml`** — the `docs="…"` line **and four** `hash.<target>.<slug>.txt` upload paths | **`TestMatrixDocumentSlugsAreRegisteredInCI`** (`matrix_registration_test.go:40`) — **deliberately untagged, so it reddens in the ordinary `go test ./...` the moment step 9 lands** |
| 11 | **`declaredEpic2GateObligations`** (`byte_neutrality_test.go:532`) — one `"matrix-document: <slug>"` line with a story/ruling citation | **`TestEpic2GateObligationsMatchTheDeclaredSet`** — also ordinary-suite |

**Steps 10 and 11 are named here because a developer who does not expect them reads two unexplained
reds — one of them in an *Epic 2* gate test — and the tempting fix is to back the registration out.**
That would ship four goldens the matrix never compares, which `matrix_test.go` already calls *"a matrix
leg nobody compares, reported as green."*

**Two conveniences the developer gets for free, worth knowing before re-deriving them:**

- `goldenDigestSite.kind` already accepts **`"signoff"`**. The sign-off record is a first-class digest
  site, which is what makes the one-record fork arm mechanically cheap.
- `TestNoRealToUnicodeSectionExceedsTheCap` (`tounicode_corpus_test.go:48`) is **the only real
  directory walk** over `fixtures/*/expected.pdf`. It picks up the four new goldens automatically, so
  AC7's `/ToUnicode` measurement is mechanical, not a bespoke probe.

### DW-14's prediction is now testable and must be MEASURED, not inherited

DW-14's chunking landed at 4.2 (`internal/pdf/textdoc.go:548,626`, `internal/pdf/tounicode_chunk_test.go`).
Its own entry predicted that 4.7's document *"is the document this entry has been waiting for"* — the
first to exceed 100 `beginbfchar` entries in a section — and warned in the same breath:
*"`page-count-20` … repeats a template. **Do not read its green as headroom for 4.7's varied data.**"*

**This story is the first to be able to check that prediction.** It must report the produced section
counts as a measurement. If the cap is crossed, the chunker's boundary is live in a golden for the
first time and that is a semantic property worth asserting; if it is not crossed, say so plainly rather
than restating the prediction. Either answer is fine; an unmeasured one is not.

---

## Decision to surface — the Thai strings: reuse the frozen labels, or mint a third sign-off

**Pre-framed by the lead's own Epic 4 grounding** (`folio-mvp-decision-log.md:476`): *"Decision: does
4.7's Thai reuse strings already frozen in `fixtures/expected-breaks/expected_breaks.json`, or introduce
new ones? … **Lead's; will rule reuse.**"* That is a stated leaning in a grounding note, **not a ruling
in the log**, so it is framed here rather than treated as settled.

- **Arm A — reuse the frozen strings.** AC5 becomes a mechanical comparison against labels a human has
  already signed (`break-signoff.json`, reader named, digest bound). No new human attention is spent on
  break placement. Cost: the transaction descriptions are constrained to strings chosen for a different
  purpose, which may make the statement read as artificial.
- **Arm B — mint new Thai strings.** The statement reads naturally. Cost: a **third** human sign-off
  under D-000.26 (refined) and D-000.41 — and D-000.41 exists precisely because *"you are buying
  attributability with someone else's attention."*

**Recommendation: Arm A, with one qualification the lead's note does not cover.** Reuse the frozen
strings **for the cells whose breaks AC5 asserts**, and allow ordinary unasserted Latin/CJK text
elsewhere in the table. The frozen set must supply at least one string whose label carries an interior
seam (AC5's part (a) below depends on it) — if it does not, that is a finding to report, not a reason
to mint quietly.

**What changes under each:** under A, AC5 is a comparison and this story requests **one** human
sign-off (the reading one, AC8). Under B, it requests **two**, and D-000.41's sequencing question
reopens.

---

## Decision to surface — the shape of the reading sign-off across four goldens

D-2.3.5's mechanism binds one record to one digest. This story records **four** goldens at once.

- **Arm A — four records, one per fixture.** Literal application. Cost: four separate human looks at
  four documents that differ only in length; D-000.41's dilution — *"it trains the signer that
  re-approval is routine"* — arrives immediately.
- **Arm B — one record naming all four digests.** One look, one attestation, and re-recording **any**
  of the four invalidates the whole record and demands a fresh look.

**Recommendation: Arm B.** It is stricter than Arm A on the property that matters (invalidation is
all-or-nothing) and cheaper on the property D-000.41 protects. The record's `examined` field must say
what was actually done — the recommendation is: the 1-page and the 50-page documents read in full, and
for the 5- and 20-page documents the page-boundary pages (the last page before each break and the
first page after) examined specifically, because that is where table pagination can be wrong in a way
the interior pages cannot show.

**What changes under each:** under A the story ships four records and four gate tests; under B, one of
each, and `assertSignOffMatchesFrozenHash`'s single-digest helper needs a four-digest sibling.

**Not a fork, stated so it is not re-opened:** *whether* the gate test is `//go:build matrix` is settled
by D-2.3.5 — machine-checkable properties are asserted at recording with **no deferral available**; only
the irreducibly-human half may be pending, and only behind the matrix tag so the story commits green
while **the gate cannot pass**. This story follows that split exactly.

---

## D-000.41 check — is scheduled work known to move these goldens?

Required before requesting a human sign-off. Measured against the epic and the register:

- **Story 4.8 (alternating row styling)** is scheduled next and is *"optional, first to be cut."* It
  applies `altRowBackground`. **This story's fixtures must not set that key**, so 4.8 changes these
  documents only if it changes table rendering *defaults* — which its own AC does not require. Low
  risk; state it, do not assume it.
- **DW-13 adoption (font-stream compression)** would move **every** golden in the repository. It is the
  **owner's**, batched at Epic 4 close, i.e. **after** this story. If adopted, this story's sign-off is
  correctly invalidated and re-requested.
- **DW-4 (`folio-go/v0.1.0`)** is a release decision, not a byte-moving one.

**Conclusion, and it is not a clean "no".** There is one known candidate (DW-13) that could move these
bytes shortly after the sign-off is recorded. **The recommendation is still to request it in this
story**, because this story *is* the C4 gate and D-4.7.0 makes the sign-off the only non-circular check
on the recording — a gate that passes on a circular claim is worse than a sign-off that may be
re-requested. **The developer states this trade-off to the owner when making the request**, rather than
letting the owner discover it at Epic 4 close.

---

## Acceptance criteria

Every output-producing AC carries its **observable count** and a **named deletion-mutation per
observable** (D-000.85). Where an AC has no machine observable, it says so rather than inventing one —
D-000.88's declared-vs-audited delta is the anti-gaming metric, and it is raised, not lowered, by an
honest small denominator.

> **Declared total: 22 observables across AC1–AC12.**
> Per-AC: AC1 **2**, AC2 **4**, AC3 **3**, AC4 **1**, AC5 **2**, AC6 **3**, AC7 **2**, AC8 **2**,
> AC9 **1**, AC10 **1**, AC11 **1**, AC12 **0**.
> **Corrected at finish from 21 (Finding 6, a Major).** AC10 was declared **0** while shipping a live,
> separably-witnessed guard over the written record. D-000.88's delta is an anti-gaming metric and
> under-declaring games it in the FLATTERING direction — an unclaimed observable shrinks numerator and
> denominator together. The story warns against a fake zero and then shipped its mirror image. AC12's
> zero is genuinely honest and stays 0.
> The reviewer **audits this count** (D-000.88): the story declares the denominator, review finds the
> numerator, and the ledger records **three numbers — declared, audited, Class A — never one ratio.**

### AC1 — the four statement documents render at exactly 1, 5, 20 and 50 pages, and the page count is a consequence of the bound data

**Given** the Customer Account Statement fixture and four transaction collections
**When** each is rendered
**Then** the produced document carries exactly 1, 5, 20 and 50 `/Type /Page` objects respectively,
counted off the produced artifact
**And** the page count follows the **length of the bound collection**, not the geometric placement of
elements.

- **Observables: 2** — (i) each document's produced page count equals its declared count; (ii) the page
  count is data-dependent — removing rows removes pages.
- **Deletion (i):** change the paginator's content-window advance so one extra row fits per page. The
  declared-count assertion reddens; (ii)'s data-dependence assertion stays **green** (it is a relative
  property and survives a uniform shift). That green is the discrimination.
- **Deletion (ii):** replace the table's bound rows with fixed-position marker elements spaced one
  content window apart — `fixtures/page-count-20`'s construction. (i) stays **green** at all four
  counts; the data-dependence assertion reddens.
- **D-000.86 part (a), and it is the whole reason (ii) exists.** *"Renders at N pages"* is
  **accidentally true by construction** for any geometrically-built document, which is exactly how
  Story 2.7's `page-count-*` fixtures work and why their green says nothing about tables. The
  accidental cause is therefore **never introduced**: the fixture is data-driven from the start, and
  deletion (ii) is the standing demonstration that a geometric fixture would satisfy (i) — so (i) alone
  is not evidence, and this AC records that in writing.
- **D-000.68 / D-000.83:** assert the **set** `{page → row indices}` at each boundary, not only a page
  count. Two different wrong arrangements produce the same count.

### AC2 — the rendered content is read off the produced PDF and matches an independently-stated expectation

**Given** the four rendered statements
**When** their content streams are decoded through each document's own `/ToUnicode` CMap
**Then** the five column header labels appear on **every** page; the first and last transaction rows'
cell text appears with the values the data declares, formatted by the declared locale; the sum footer
appears **exactly once, on the last page only**, and equals a total computed independently of the
renderer; and the page footer's confidentiality text, `Page X of Y` and the `params`-supplied date
appear on every page.

- **Observables: 4** — (i) the header-label set on every page; (ii) the first and last row's cell text
  and formatting; (iii) the sum footer's single occurrence, its page, and its value; (iv) the page
  footer's three components.
- **Deletion (i):** suppress the header repeat on continuation pages. (ii)(iii)(iv) stay green; (i)
  reddens.
- **Deletion (ii):** drop the last row's text runs from the page assignment. (i)(iii)(iv) green; (ii)
  reddens.
- **Deletion (iii):** emit the footer aggregate on every page instead of the last. (i)(ii)(iv) green;
  (iii) reddens on both the count and the page index.
- **Deletion (iv):** drop the confidentiality literal from the page-footer band. (i)(ii)(iii) green;
  (iv) reddens.
- **D-000.21:** every one of these is read **off the produced artifact**, never off the input.
  `page_count_matrix_test.go`'s `splitPageContentStreams` / `mpParseToUnicode` / `mpExtractRuns` are the
  established instruments and are reused rather than re-derived.
- **D-000.86 part (a), two instances:**
  - **(i) is accidentally true on the 1-page document** — it has no continuation page, so "on every
    page" is satisfied by one page. **Part (a): assert (i) on the 5-, 20- and 50-page documents by
    name, and confirm that removing the 1-page document from the loop leaves the assertion passing.**
  - **(iii) is accidentally true if the extraction only reads page 1.** Part (a): extract from **every**
    page and assert the occurrence count over the whole document, then confirm the assertion still
    passes with the first page excluded.

### AC3 — Latin, Thai and CJK in the same table, at least one wrapping cell, and an embedded logo

**Given** the fixture data
**When** it is inspected and rendered
**Then** at least one table row draws glyphs from all three shipped faces under one fallback stack; at
least one cell occupies more than one line and its row's height grew accordingly; and the logo image is
embedded **once** and referenced from every page that shows it.

- **Observables: 3** — (i) the three-face row; (ii) the multi-line cell and the grown row height;
  (iii) the image XObject's definition count and its per-page references.
- **Deletion (i):** collapse the fixture's font stack to the Latin face alone. (ii)(iii) green; (i)
  reddens.
- **Deletion (ii):** widen the wrapping column so its content fits on one line. (i)(iii) green; (ii)
  reddens on both the line count and the row height.
- **Deletion (iii):** remove the logo element. (i)(ii) green; (iii) reddens.
- **D-4.5.5 — this is the AC most exposed to a fixture that cannot express the defect.** Measured at
  HEAD: **no test and no fixture puts all three scripts inside one table.** The table suites use a
  single-face `fontFamily` per subtest. So (i) is genuinely new coverage and there is nothing to
  inherit. For (ii), a cell whose content is *one* physical line longer than the column is not enough —
  choose content whose wrap point falls between two words so a wrong break is visible, and name the
  wrap point in the test.
- **D-000.86 part (a) on (iii), and it is load-bearing.** *"An image XObject is present"* is
  **accidentally true** of any document with an image — `fixtures/image-embed` already proves it, and
  the AC would pass while saying nothing new. **The discriminating property is one definition and many
  references across 50 pages.** Part (a): probe HEAD first with a two-page document carrying a
  header image and **record whether the definition count is already 1**. If it is, say so — the AC then
  pins an existing property at a new scale rather than creating one, which is legitimate but must be
  declared, not implied.

### AC4 — the generated date arrives through `params` and is never read from a clock

**Given** the statement fixture
**When** it is rendered with a `params` document supplying the generated date
**Then** the rendered date text equals the `params` value, and rendering with a different `params` value
changes that text and nothing else.

- **Observables: 1** — the rendered date's provenance is `params`.
- **Deletion:** make the date placeholder resolve from the **data** document instead of `params`. The
  provenance assertion reddens.
- **The second half is discharged by an existing guard, not by new code, and is declared as such rather
  than counted.** The clock ban is enforced statically by
  `lint/internal/rules/forbiddenimports.go` (`time` banned under `internal/` **and** in `_test.go`),
  red-proofed by `story34_forbiddenimports_test.go:98`, and measured at HEAD: **one** `time.Now`
  occurrence in the whole repository, the lint rule's own fixture literal. **Adding an observable here
  would be inflating the denominator for a property this story does not implement.**
- **D-000.86 part (a):** the rendered text is accidentally equal to the `params` value if the data
  document happens to carry the same string. Part (a): the data document must carry **no** date field,
  and the test must still pass with a `params` value that appears nowhere else in the fixture.

### AC5 — Thai line breaks match the frozen expected-break fixture, and Thai marks are positioned by `GPOS`

**Given** the Thai text in the fixture's table cells
**When** its line breaks are compared to `fixtures/expected-breaks/expected_breaks.json`
**Then** they match exactly
**And** at least one Thai mark's position differs from its unpositioned default, proving `GPOS` was
applied inside a table cell.

- **Observables: 2** — (i) the break positions of the Thai cell text; (ii) `GPOS` mark positioning
  inside a table cell.
- **Deletion (i):** bypass the dictionary breaker for table-cell text and fall back to per-character
  breaking. (ii) green; (i) reddens.
- **Deletion (ii):** drop `GPOS` application for the table's runs. (i) green; (ii) reddens.
- **D-4.5.5 / D-000.86 part (a), and this AC is vacuous without it.** A Thai cell that never wraps
  compares zero breaks and passes. **Part (a): the chosen Thai string must be one whose frozen label
  carries at least one interior seam, and the column width must fall between two seams**, so the
  comparison is a proper non-empty split. Name the string and the seam in the test. If the frozen set
  contains no such string at a usable width, **report it** — that is the Thai-strings fork's Arm A
  failing on measurement, not a reason to mint quietly.
  For (ii), a mark whose positioned offset happens to equal its default proves nothing; use a
  lowered/stacked case — `shaped-text`'s own `ปั ฟั ที่ ป้ำ` is the established one.
- Three of the frozen labels (`thai-007`, `thai-008`, `thai-009`) are **recorded divergences** where the
  human label and the engine deliberately disagree — read `break-signoff.json` before choosing, and do
  not assert against a divergent label.

### AC6 — four goldens are recorded, registered at every site, structurally valid, and independently resolved

**Given** the four rendered statements
**When** they are frozen
**Then** each `expected.pdf`'s own sha256 equals its `expected.json` digest **and** its
`goldenDigestRecord` literal; each page tree resolves under the hermetic oracle; and an independent
reader resolves each into its declared page count, with reader, version and **verbatim invocation**
recorded in the fixture's provenance.

- **Observables: 3** — (i) the three digest sites agree; (ii) the page tree resolves; (iii) an
  independent reader resolves it.
- **Deletion (i):** change one of the three digest sites. The three-site agreement reddens; (ii) stays
  green (the bytes are untouched).
- **Deletion (ii):** feed the oracle a **scratch copy** whose `/Kids` separator is removed — the exact
  multi-page defect. (ii) reddens; (i) stays green **because the committed file is not touched**. Do not
  mutate the committed golden; a mutation that moves the bytes cannot separate these two observables.
- **(iii) has no deletion and that is correct** — D-000.53's guardrail is explicit: *"this step is not
  discharged by an assertion this project wrote."* It is a recorded run of an external tool
  (`qpdf --check`, 12.4.0 is today's instrument), and the artifact is its verbatim output in provenance.
  **Declared as an observable and reported without a mutation, rather than dropped to make the count
  tidy.**
- **D-2.6.6:** structural validity does **not** become a gate obligation. It is asserted here, at
  recording, in the ordinary suite, via `TestEveryGoldenPDFResolvesItsPageTree` picking up the four new
  `goldenDigestRecord` entries.

### AC7 — byte identity across `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm`, for all four documents, run in this story

**Given** the four page counts
**When** each is rendered on all four targets
**Then** every hash is identical across all four targets
**And** each equals the recorded golden
**And** each reference render is confirmed non-blank and page-count-correct on every leg.

- **Observables: 2** — (i) pairwise identity across the four targets; (ii) agreement with the recorded
  golden.
- **Deletion (i):** perturb one target's captured bytes before comparison. Pairwise identity reddens;
  (ii) — read as "the other three agree with the golden" — stays green.
- **Deletion (ii):** perturb the recorded golden. (i) stays green (all four legs still agree with each
  other); the golden comparison reddens. **This is `matrix_test.go`'s own AC17 two-case split and it
  already exists — reuse it, do not re-derive it.**
- **D-000.4 override, D-000.54:** the legs are **run in this story**, not deferred. Register each
  document in `matrixDocuments` with an `extraGuard` asserting non-blank and page-count-correct, on the
  same basis as `requirePageCount20HasCorrectPageNumbers`.
- **D-000.86 part (a), and this one is the epic's central trap.** Cross-target identity is **already
  true for all ten registered documents.** Four more documents that exercise nothing new would pass
  vacuously and add a green that means nothing. **Part (a): name and measure what is new** — CJK
  subsetting at volume, and `/ToUnicode` section counts. **Report the produced section counts as a
  number.** DW-14 predicts the 100-entry cap is crossed for the first time here and warns explicitly
  that `page-count-20`'s green is **not** headroom. Confirm or refute by measurement; an inherited
  prediction is not a result.

### AC8 — the human semantic acceptance step, recorded and enforced

**Given** the four goldens
**When** the Epic 4 boundary gate runs
**Then** it fails unless a named human has recorded, in a machine-read record, what they examined and
what they saw, bound to the digests of the goldens they examined
**And** re-recording any of those goldens invalidates the record.

- **Observables: 2** — (i) the record exists and every field is non-empty and specific; (ii) the record
  is digest-bound, so a re-record cancels it.
- **Deletion (i):** delete the record file. The gate reddens with a message written for the person who
  must resolve it (D-000.37).
- **Deletion (ii):** leave the record in place and change one golden's recorded digest. (i) stays
  **green** — the file is still there and still fully populated; the digest-binding assertion reddens on
  its own. That green is the discrimination, and it is the whole anti-staleness property.
- **What the human attests to, stated so it is performable and so a rubber stamp is visible as one.**
  The record's `examined` field must name: which documents were opened; that the transaction rows carry
  the values the data declares; that the column header appears on continuation pages; that the total at
  the foot is the total of the rows above it; that the Thai reads correctly and its marks sit on the
  consonants they belong to; that the CJK is not the wrong weight or the wrong face; that the logo is
  present and not a black box; and that the page numbering and the supplied date are right. **A record
  saying "looked fine" is a defect and the schema's non-empty check is not sufficient to catch it —
  the reviewer reads this field.**
- **This AC is the only non-circular check on the first recording (D-4.7.0).** If it is weakened, the
  story ships a hash of an unexamined document, and every guard this project owns will then defend it
  faithfully and forever.

### AC9 — the golden does not supersede the table behaviour suite

**Given** four table goldens now exist
**When** future work changes table behaviour
**Then** the six table behaviour test files remain the second, independent producer of the same answers,
and their existence is asserted against a declared set of names, never a count.

- **Observables: 1** — the declared file-name set is present and non-empty.
- **Deletion:** remove one name from the declared set while leaving the file in place — the set-equality
  assertion reddens (D-2.5.1's shape: adding or removing a member is a one-line reviewable diff, not a
  rename).
- **Stated honestly, because a guard oversold is worse than none.** This catches **deletion**. It does
  **not** catch **rot** — tests kept in place but hollowed out. Nothing mechanical in this project
  catches that, and pretending otherwise would be the "checkbox nobody can fail" the brief warns
  against. **The real protection against rot is the per-observable deletion screen (D-000.85) applied by
  every future story that touches table code**, and the six files are named in this story so a later
  reader knows what population that screen must still redden.
- **Say it in the story, per D-4.7.0's obligation 2, and it is said here:** the goldens and the
  behaviour suite are **two independent producers of the same answers**. The golden pins bytes and
  cannot say whether they are right; the behaviour suite asserts properties and cannot say whether the
  bytes moved. **Neither replaces the other, and the pairing is the construction that has reliably
  worked all programme.** Do not retire, thin, or fold the behaviour suite into the goldens.

### AC10 — the coverage history is recorded

**Given** a future reader at a gate
**When** they ask what the golden corpus covered before this story
**Then** each new fixture's README and this story record, in writing, that **no committed golden
contained a table before this story**, and that Story 4.6's unconditional-clip mutation reddened **no
golden at all** while reddening the table behaviour suite.

- **Observables: 1** — the four fixture READMEs carry the coverage-history record, guarded by
  `TestStatementFixtureReadmesRecordTheCoverageHistory`.
- **Deletion:** re-spell the required sentence in one fixture's README (`reddened NO GOLDEN AT ALL` ->
  `reddened NOTHING WHATSOEVER`). That fixture's subtest reddens; the other three stay **green**. That
  green is the discrimination, and the reviewer performed it.
- **CORRECTED AT FINISH, from a declared 0 (Finding 6, a Major).** The original declaration read
  *"Observables: 0. This is a written record, not a mechanism"* — and then, in the same breath, §13
  recorded that the record IS guarded. A live, separably-witnessed machine observable by every
  criterion this story applies to the other twenty-one is an observable. **D-000.88's delta is an
  anti-gaming metric and under-declaring games it in the flattering direction**: an unclaimed
  observable shrinks numerator and denominator together and leaves the ratio looking better than it is.
  The story explicitly warns against a fake zero; this was its mirror image. AC12's zero survives audit
  and stays 0.
- What the guard does NOT do: it checks that the sentence is PRESENT (whitespace-normalised), not that
  it is true. AC10 is a written record and the presence check is the most a machine can offer it.
- Each new fixture README follows `fixtures/page-count-20/README.md`'s established sections: *why this
  fixture exists*, *what this document's content can express*, *what it cannot express*, and
  *independent-reader acceptance (D-000.53)*.

### AC11 — no existing recorded bytes move (AD-22)

**Given** this story adds fixtures and registrations
**When** the full suite runs
**Then** the ten existing golden digests are unchanged.

- **Observables: 1** — the existing digest set is unmoved.
- **Control mutation (not a deletion — the claim is negative):** make a byte-moving change to a shared
  emitter and confirm the existing byte-neutrality assertions **do** redden, so their green is evidence
  rather than absence. Story 4.6's AC7 is the precedent for this control.
- **AD-22, stated because it is a decision and not an implementation detail:** if anything in this story
  would move already-recorded bytes for a reason other than the four new goldens, **stop and surface
  it**. That is a versioned behaviour change for every downstream test suite, not a fixture edit.

### AC12 — DW-11 answered in writing; DW-13 measured, not adopted

**Given** DW-11's owner window names *"Epic 4's golden-report work"* and DW-13's sizing is scheduled
*"during Story 4.7"*
**When** this story completes
**Then** DW-11 carries a written answer in the same form Stories 2.4 and 2.6 gave it — whether any
genuinely-uncoverable, independently-attested opaque Thai name was found, and the load-bearing count
**And** DW-13 carries the measurement it has been asking for: the uncompressed `FontFile2` payload cost
of the 50-page statement with real CJK content, in bytes, alongside an estimate of what Flate would
save.

- **Observables: 0.** Both are recorded measurements and written answers, not assertions.
- **DW-11's constraint is binding: D-000.17 forbids inventing attestation to reach a number.** "None
  were found, and none were invented" is the correct answer if it is the true one. Do **not** attempt to
  turn `TestCorpusMeetsP6ExerciseFloors` green.
- **And note the guard cuts both ways.** `TestCorpusP6StatsMatchDeclaredBaseline`
  (`internal/text/corpus_test.go:238`) pins the exact literals `P6a:64 P6b:63 P6c:16 P6d:20 P6e:284
  P6f:115 P6g:7` and is **green in the ordinary suite**. It reddens on an **improvement** too. So if a
  genuinely-attested opaque name *is* found, adding it requires a deliberate baseline edit in the same
  commit — which is the mechanism working, not an obstacle. Either way, the answer is written down.
- **DW-13's scope is measurement only.** Adoption is the **owner's**, batched at Epic 4 close. Adopting
  it here would move every golden and require narrowing `lint/internal/rules/nocompressor.go`, a guard
  that has been proved to fire. If the 50-page CJK payload comes back in the tens of KB, the honest
  recommendation is *"leave it"* and DW-13 closes cheaply.

---

## Blast radius — what this story adds, and what it must not touch

**Adds:** a new fixture family (four directories), their `input.folio` / data / params documents, four
`expected.pdf` + `expected.json` pairs, four `goldenDigestRecord` entries, four `matrixDocuments`
entries with capture functions and an `extraGuard`, four `docs=` slugs and sixteen upload paths in
`.github/workflows/matrix.yml`, four `declaredEpic2GateObligations` lines, a fixture test file, a
sign-off record and its matrix-gated gate test, four READMEs, and written answers to DW-11 and DW-13.

**Two reds the developer must expect and must NOT resolve by backing out the registration:**
`TestMatrixDocumentSlugsAreRegisteredInCI` and `TestEpic2GateObligationsMatchTheDeclaredSet`. Both are
untagged, both fire in the ordinary suite, and both are satisfied by adding the declaration — see the
five-registry table above.

**Must not touch:** any existing `expected.pdf`, any existing digest, `altRowBackground` (Story 4.8's),
`lint/internal/rules/nocompressor.go` (DW-13's adoption is the owner's), the P6g floor (DW-11 /
D-000.17), or the four named skips (DW-21's, and the Epic 4 gate's).

**Does not close DW-21.** DW-21's owner is the Epic 4 boundary gate **and** the orchestrator's
checklist, deliberately two addresses because *"a gate has failed as an owner before"*. Its heavy
command must be run and recorded **at the Epic 4 gate**, not here — D-000.4's per-epic cadence. **This
story states the obligation so the gate does not have to remember it unaided** (D-000.87: a distant-event
owner has already absorbed four instances once this epic). The gate's command, verbatim from the
register:

```
env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -count=1 -v \
  -run 'TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable|TestTwoTablesWithPageCountFooterRenderConsistently|TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability' ./...
```

---

## Tasks

1. [x] **Re-measure the baseline before writing anything.** Reproduce the pass/fail/skip table and the
   skip set above at your own baseline, with the exact invocation, and confirm
   `grep -l '"table"' fixtures/*/input.folio` still returns nothing. If any figure differs, **stop and
   report** — do not build on this story's stated baseline without re-confirming it.
2. [x] **Probe HEAD for AC3(iii)'s accidental truth** (D-000.86 part (a)): render a throwaway two-page
   document with a header image and record how many image XObject **definitions** the produced PDF
   carries. Record the number; delete the probe.
3. [x] **Settle the two forks** — Thai strings (reuse vs mint) and sign-off shape (four records vs one).
   Both carry recommendations above. If the frozen break set contains no usable interior-seam string at
   a workable column width, that is a **finding**, not a licence to mint.
4. [x] **Author the Customer Account Statement fixture.** One `.folio` template; four data documents
   differing only in the length of the bound transaction collection; one params document supplying the
   generated date. **Page count follows the collection length** (AC1(ii)). Do **not** set
   `altRowBackground`. Include the logo, the customer and account blocks, the statement period, the
   five-column table with repeated header and sum footer, and the page footer carrying the
   confidentiality text, the params date and `Page X of Y`.
5. [x] **Write the discriminating content before writing any assertion** (D-4.5.5). The three-script
   row, the wrapping cell whose wrap point falls between two words, the Thai cell whose frozen label
   carries an interior seam, and the lowered/stacked mark case. **These are the fixtures without which
   AC3 and AC5 are vacuous.**
6. [x] **Assert semantic content off the produced artifact** (AC2), reusing
   `splitPageContentStreams` / `mpParseToUnicode` / `mpExtractRuns` rather than re-deriving them.
   Assert the `{page → row indices}` set at each boundary, never a page count alone.
7. [x] **Record the four goldens and register them at every declared site** — `expected.json`, the
   second literal in `goldenDigestRecord`, and the README if it quotes the digest — plus
   `TestEveryGoldenPDFResolvesItsPageTree` (AC6). Then run the independent reader on each
   (`qpdf --check` and `qpdf --show-npages`) and record reader, version and **verbatim invocation with
   its output** in each fixture's README (D-000.53). **Order matters: the semantic assertions of AC2–AC5
   are written and passing BEFORE any digest is frozen.**
8. [x] **Complete the registration across all five registries** — `goldenDigestRecord`,
   `matrixDocuments`, `.github/workflows/matrix.yml`'s `docs=` line and sixteen upload paths, and
   `declaredEpic2GateObligations` — then **RUN the four legs in this story** (D-000.4 override,
   D-000.54). Report every leg's hash. **Measure and report the `/ToUnicode` section counts** from
   `TestNoRealToUnicodeSectionExceedsTheCap`, which picks the new goldens up by directory walk — DW-14's
   prediction is confirmed or refuted by that number, not restated.
9. [x] **Build the sign-off record and its matrix-gated red gate**, copying
   `shaped_signoff_matrix_test.go`'s schema, non-empty field checks, digest binding and
   human-executable failure message. **Then make the request to the owner**, stating the D-000.41
   trade-off (DW-13 may move these bytes at Epic 4 close) rather than letting them discover it later.
10. [x] **Write the AC9 declared-set guard** over the six table behaviour files, by name (D-2.5.1), and
    write the two-producer statement into the fixture READMEs so a later reader does not have to
    re-derive it from this story file.
11. [x] **Answer DW-11 in writing and record DW-13's sizing measurement** (AC12). Neither is a code
    change. Do not touch the P6g floor.
12. [x] **Run the full suite plus the byte-neutrality and matrix suites.** Report **pass / fail / skip
    separately, with the exact invocation** (D-000.89). Report skips **by name** as a set (D-000.81) and
    the red-by-design test **separately by name**. Never a passed count labelled a total.
13. [x] **Report the observable ledger.** Declared count is **21**, per-AC as tabulated. Report Class A
    split by sub-cause — compound-observable / prose-claim / other (D-000.85). **The finisher PROPOSES a
    corrected denominator and does not apply it; the reviewer adjudicates** (D-000.88). This is **n=2**
    of the pre-registered series that began at 4.6 (*declared 15, audited 14, Class A 6*) — report the
    number that occurs, not a favourable one.
14. [x] **Do not commit.** This story ends at status `review`.

---

## Dev notes

- **AD-1.** `internal/` is render path: no `time`, no `os`, no map ranged into output. Nothing in this
  story needs any of them; if something appears to, that is a design error, not a carve-out.
- **AD-5.** `internal/layout` learns nothing about PDF. All four goldens are produced through the
  existing public `Render`.
- **AD-14.** If any of the four documents produces a `Warning`, it travels alongside the bytes and is
  asserted. A golden that silently carries an unasserted diagnostic is a recording of a document nobody
  read.
- **AD-23.** The statement's money column is the reason AD-23 exists. Amounts are exact scaled-integer
  decimals; the sum footer's independently-computed check value in AC2(iii) must be computed as an exact
  decimal too, or the check inherits the bug it exists to catch.
- **D-000.79.** Whoever names a mutation must not have written the code. Every mutation above was named
  at creation, before any implementation existed. If the implementation's shape makes one impossible to
  perform as written, **say so and name a replacement** — never silently substitute one the code
  happens to survive.
- **D-000.40.** A mutation asserts it applied by observing the artifact — a non-empty diff or a reported
  count — never by an exit code. And beware the portable-tooling hazards on darwin: `\b`, `-i`, `sort`,
  `grep -P`.
- **D-000.67 part 2.** Check every witness the recording mechanism reports, not only the ones these ACs
  made you think of. If a fixture's provenance grows a second population, enumerate both.
- **D-000.37.** Every failure message this story adds must be executable by the human who will read it,
  and any story that later changes the mechanism a message points at updates the message in the same
  commit.
- **The one-line reminder that outranks all of the above.** Once these four files are committed they
  stop being outputs and become inputs, and every guard this project owns will defend them faithfully
  and forever. **The only moment "is this right?" is answerable at all is before the commit.**


---

## Delivery Log

### 0. Baseline, re-measured before anything was written (Task 1)

**Exact invocation**, from `folio-go/`:

```
env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...
```

| population | pass | fail | skip | total |
|---|---|---|---|---|
| top-level | 711 | 1 | 4 | **716** |
| including subtests | 1088 | 2 | 4 | **1094** |

**Every figure matches the story's stated baseline exactly.** Skips, as a set by name:
`TestXrefEntriesRejectsMalformedSubprocess`,
`TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`,
`TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`,
`TestTwoTablesWithPageCountFooterRenderConsistently`. Red by design, named separately:
`TestCorpusMeetsP6ExerciseFloors`, subtest `P6g (opaque names)`, `got 7, need >=20`, stats line
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.

`grep -l '"table"' fixtures/*/input.folio` returned **nothing** — re-confirmed, not inherited.

### 1. The two held decisions, as ruled

**One sign-off record over four digests, invalidated in whole.** The record's **SHAPE** is built as
ruled — `statementSignOff` in `folio-go/statement_golden_fixture_test.go` declares `reader`, `date`,
`examined` and a `digests` object naming **all four** slugs. **The record itself does not exist, and
must not: its absence is what keeps the gate red.** `statementSignOffStaleness` rejects a record whose
named SET differs from the registered set, and reports **the whole record stale** if any one digest has
moved. Over-invalidating costs a re-read; under-invalidating ships an attestation nobody made.

*(Corrected at finish — Finding 10. As first written this paragraph described the schema in the PAST
TENSE as a thing that had been built, in the section a reader consults for what the held decisions
produced. §5, §13 and §18 all say the opposite, correctly. "The human sign-off has already been
obtained" is the single most consequential fact in this story to get backwards.)*

**Thai strings: Arm A, reuse.** Applied. The asserted break is `thai-001` (`ประเทศไทย`, labelled
`["ประเทศ","ไทย"]`, seam at rune 6), read **out of `fixtures/expected-breaks/expected_breaks.json` at
test time** rather than restated in Go. The lowered/stacked mark cell is `ปั ฟั ที่ ป้ำ`, verbatim from
`fixtures/shaped-text/input.folio`. **No new Thai string was minted and no third Thai sign-off is
requested.** CJK is ordinary text, deliberately NOT drawn from the frozen corpus, for the same reason
the ruling gives for Thai: a second apparent authority over the same rules is the cost being avoided.

**The width knob was not needed, and the fallback did not arise.** A 34pt Note column (28pt usable)
breaks `thai-001` at its labelled seam on the first setting.

**Two corrections to figures in the brief, both measured:**

- The brief says `expected_breaks.json` carries *"~40 `transaction_description` items"*. It carries
  **26 items** in total — 18 `thai-*` and 8 `cjk-*` — and no field named `transaction_description`.
  Arm A still holds, on `thai-001`.
- **`thai-002` (`เก็บเงิน`) cannot be asserted at any usable width. The CONCLUSION stands; the reason
  first recorded here was an inference, and is withdrawn (Finding 13; D-4.7.9).** The reason given was
  that its labelled words present **six advancing glyphs** — *"the same six as `thai-001`'s first word
  `ประเทศ`"*. The glyph **count** is right (8 runes less two zero-advance marks, re-verified); the
  argument is not. They are not the same six — `เ ก บ เ ง น` against `ป ร ะ เ ท ศ` — and line breaking
  is decided by **advance width**, not glyph count.

  **Measured at finish** by sweeping the Note column's declared width and reading the breaks and the
  `TEXT_CLIPPED_WIDTH` diagnostics off the render:

  | subject | advance |
  |---|---|
  | `thai-001`'s first labelled word `ประเทศ` | **23,704 mp** (from the clip diagnostic's own *"widest laid-out line"*) |
  | `thai-002` `เก็บเงิน`, whole string | fits at a content width of **23,400 mp**; breaks at **23,200 mp** |

  Breaking `thai-002` at its seam needs a content width below ~23,400 mp; `ประเทศ` does not fit under
  23,704 mp. **The two windows are disjoint by at least 304 mp**, so every width that breaks `thai-002`
  clips `thai-001`'s first word and raises a diagnostic AD-14's precondition turns into a fatal. AC5
  requires *at least one* interior seam and has one. `เก็บเงิน` is kept in the fixture as a second,
  unasserted Thai cell for the human reader; it is **not** counted as a second observable.

### 2. Task 2 — the AC3(iii) probe at HEAD (D-000.86 part (a))

A throwaway two-page document with a header image, rendered at `df8cbcc`:

```
PROBE: pages=2  imageXObjectDefinitions(/Subtype /Image)=1  Do-operators-per-page=[1 1]
PROBE /XObject line: << /Type /Page ... /XObject << /5a05ad0... 3 0 R >> ... >>   (page 1)
PROBE /XObject line: << /Type /Page ... /XObject << /5a05ad0... 3 0 R >> ... >>   (page 2)
```

**One definition, referenced by object number from both pages, one `Do` per page — the property was
ALREADY TRUE at HEAD.** So AC3(iii) **pins an existing property at a new scale** (50 pages, 50
references, one definition); it does not create one. That is declared, not implied. Probe deleted.

### 3. What was built

**Four fixtures** — `fixtures/statement-{1,5,20,50}/`, each carrying `input.folio`, `data.json`,
`params.json`, `expected.pdf`, `expected.json` and `README.md`. One template, one params document,
four data documents differing **only** in the length of the bound collection.

| fixture | rows | pages | bytes | sha256 |
|---|---|---|---|---|
| `statement-1` | 5 | 1 | 76,740 | `ef58bbf6dac1c3d4a5d679a77f9907a8d45f02ccd3f886c4d4e7cbdf9e86611d` |
| `statement-5` | 95 | 5 | 127,343 | `7f67b317c0a1925a404f8435bd4736b85e831a213f5a69fc2a2934a742ff950f` |
| `statement-20` | 425 | 20 | 269,804 | `be6f5e27af94e62e7c15a1814633cc48a2a91c5ee8686f5b76de5dc12e3cd4ed` |
| `statement-50` | 1085 | 50 | 555,629 | `9c5be7ba7b4f31c7d488c114a377058ec30cec5ffca082d9c76ee26f304c754c` |

**Row arithmetic — CORRECTED AT FINISH (Finding 5, a Major), and it is two facts, not one.** The band
table recorded here was off by one at both ends of every band, and `statement_fixture_test.go` was
separately wrong about the rows-per-page figure; each document was half right about a different half,
with nothing reconciling them. Both are now measured and both are recorded in both places.

- **Rows PLACED on page 1: nineteen.** Read off the artifact —
  `TestStatementDocumentsRenderTheirDeclaredPageCount` logs the partition, and for `statement-5` it is
  `[19 22 22 22 10]`. Every continuation page places **22**. `statementRowsOnFirstPage = 19` is right.
- **Where the page count STEPS: one row lower at both ends**, because the sum footer aggregate consumes
  a row slot on the FINAL page. The largest collection that still fits in P pages is `18 + 22*(P-1)`,
  not `19 + 22*(P-1)`.

Measured directly at each edge (rendering n on both sides of every step):

```
n =   18 -> 1 page      n =   19 -> 2 pages
n =   84 -> 4 pages     n =   85 -> 5 pages
n =  106 -> 5 pages     n =  107 -> 6 pages
n =  414 -> 19 pages    n =  415 -> 20 pages
n =  436 -> 20 pages    n =  437 -> 21 pages
n = 1074 -> 49 pages    n = 1075 -> 50 pages
n = 1096 -> 50 pages    n = 1097 -> 51 pages
```

**The true bands are 1..18 / 85..106 / 415..436 / 1075..1096.** Each declared length still sits in the
MIDDLE of its band — 10 above the lower edge and 11 below the upper — so the mid-band conclusion
survives and the footer aggregate's own orphan-avoidance cannot tip a document into an extra page. The
number a future reader would have used to judge whether a new statement length is safe was wrong, and
the next story to add one would have landed on a band edge.

**What each golden discriminates (D-4.5.5) — and the honest part first.**

- `statement-1` — the five discriminating rows and nothing else. It is the document on which *"the
  header appears on every page"* is **accidentally true**, and AC2(i)'s assertion **excludes it by
  name**, logging why.
- `statement-5` — the smallest document that can express a header that fails to repeat or a footer
  aggregate emitted off the last page.
- `statement-20` — **the largest document in the family a person can realistically read END TO END**
  at the sign-off or at a re-attestation, which is the role the sign-off's own `examined` instructions
  already give it. **That is its reason to exist.** It also crosses the page-9-to-page-10 digit-count
  boundary with a table present (`page-count-20` crosses the same boundary but is table-free and
  geometrically built) — but so does `statement-50`, recorded in the same commit, so the boundary alone
  would not justify keeping it, and a future reader deciding whether it may be retired must not be told
  that it does. *(Restated at finish — Finding 15.)*
- `statement-50` — the only document in the repository with CJK subsetting at any volume: **41
  distinct CJK glyphs** against `multi-script-fallback`'s **one**.

Shared: a page count that is a **consequence of the data**; three scripts in one **cell** under one
fallback stack (**no test and no fixture in the repository did this before**); a cell whose wrap point
falls between two words; a Thai break compared against the frozen S4 corpus; GPOS inside a cell, with
its **sign** asserted; a params-supplied date occurring nowhere else in the fixture; one image
definition, many references; and — **added at finish, and the reason all four goldens were re-recorded**
— **a money column whose values are not exactly representable in binary floating point**, so the
statement can express the defect AD-23 exists to prevent (Finding 1, the Blocker; D-4.7.8).

**`altRowBackground` is not set** on any of the four — Story 4.8 owns it.

### 4. An instrument correction, reported rather than performed silently (D-000.79)

Task 6 says to reuse `splitPageContentStreams` / `mpParseToUnicode` / `mpExtractRuns`.
`splitPageContentStreams` **is** reused, unchanged. The other two **are not usable on this document,
and the reason is measured, not stylistic:**

> **`mpParseToUnicode` merges EVERY embedded face's `beginbfchar` section into ONE cid → text map.**
> On a document with more than one face the CID spaces collide and the recovered text is garbage.

Measured on this fixture before the substitution: `"Customer: Ada Lovelace"` came back as
`"ไustomerะบdaบศovelace"`. Every prior caller is single-face, which is why the defect has never shown.

**The replacements are not newly derived either** — `toUnicodeForResources` and
`parseContentStreamRuns` are the **existing per-resource instruments** `shaped_fixture_test.go`
already uses, composed here rather than re-implemented.

**AND AT FINISH IT WAS MADE FAIL-CLOSED, per D-4.7.6, because a comment is not a mechanism.** The
defect was originally recorded and left in place. The engineering lead ruled otherwise during review,
and the reasoning is the failure MODE rather than the severity: `mpParseToUnicode` has **eight callers**
and returns **plausible-looking text**, not an error, so a future multi-face caller gets a string it may
well assert against. Counting font resources is cheap and fails closed. It now `t.Fatal`s when the
document embeds more than one face, naming the defect and pointing at the two replacements.

**Red-proved both ways, at finish:**

- With the guard in place, calling it on `statement-1` fatals:
  *"mpParseToUnicode was called on a document that embeds 3 font faces ([NotoSans NotoSansSC
  NotoSansThai]), and it CANNOT decode one correctly… Use the per-resource instruments instead…"*
- With the guard neutered (`&& false`) and the same call made, it returns silently and the run decodes
  as **`"ไustomerะบdaบศovelace"`** — reproducing the developer's recorded measurement character for
  character, from an independent run.

**Nothing was contaminated**, confirmed independently by the reviewer and again here: none of this
story's assertions routes through `mpParseToUnicode`, and `matrix_test.go:904`'s call is scoped to the
single-face `page-count-20`. All eight existing callers are single-face, so the new fatal reddens
nothing — measured: the full suite is unchanged by it.

### 5. Registration — all five registries, plus the two expected reds

Both reds the story warned about arrived, and **neither was resolved by backing the registration
out**:

- `TestMatrixDocumentSlugsAreRegisteredInCI` — satisfied by adding four slugs to
  `.github/workflows/matrix.yml`'s `docs=` line **and sixteen** `hash.<target>.<slug>.txt` upload
  paths (4 documents x 4 targets).
- `TestEpic2GateObligationsMatchTheDeclaredSet` — an **Epic 2** gate test firing on an Epic 4 story,
  exactly as predicted. Satisfied by adding five lines to `declaredEpic2GateObligations`: four
  `matrix-document:` lines and one `matrix-file: statement_signoff_matrix_test.go`.

The `signoff` digest site is **deliberately NOT declared yet** on the four `goldenDigestRecord`
entries — the record does not exist, and declaring the site would redden the ordinary suite, which
D-2.3.5 forbids ("the story commits green while the GATE cannot pass"). This follows
`fixtures/shaped-text` exactly, where the site was ADDED after the owner signed. The gate test's own
failure message instructs the reader to add it, and the completeness half of
`TestGoldenDigestAgreesAtEveryDeclaredSite` will **force** it: once the record exists, its digests are
an undeclared occurrence.

### 6. AC7 — the four matrix legs, RUN IN THIS STORY (D-000.4 override, D-000.54)

```
env CGO_ENABLED=0 GOWORK=off go test -tags matrix -count=1 -timeout 120m -run TestCrossTargetByteIdentity .
--- PASS: TestCrossTargetByteIdentity (18.57s)
```

All four documents, all four targets, every leg identical **and** equal to the recorded golden:

| document | darwin/arm64 | linux/amd64 | linux/arm64 | js/wasm |
|---|---|---|---|---|
| `statement-1` | `ef58bbf6…611d` | `ef58bbf6…611d` | `ef58bbf6…611d` | `ef58bbf6…611d` |
| `statement-5` | `7f67b317…950f` | `7f67b317…950f` | `7f67b317…950f` | `7f67b317…950f` |
| `statement-20` | `be6f5e27…d4ed` | `be6f5e27…d4ed` | `be6f5e27…d4ed` | `be6f5e27…d4ed` |
| `statement-50` | `9c5be7ba…c754` | `9c5be7ba…c754` | `9c5be7ba…c754` | `9c5be7ba…c754` |

The ten pre-existing documents' legs also ran and all still match their recorded goldens.

**D-000.86 part (a) — cross-target identity was ALREADY TRUE for all ten registered documents, so
what is new had to be named and measured, not asserted.** What is new is **CJK subsetting at volume**
(41 distinct CJK glyphs against one), and it is asserted per leg by
`requireStatementIsAWorkingStatement`, which counts the Noto Sans SC subset's `/ToUnicode` entries on
every target: 1 for `statement-1`, 41 for the other three.

### 7. DW-14 — MEASURED, and the prediction is REFUTED

DW-14 predicted 4.7's document would be *"the first to exceed 100 `beginbfchar` entries in a section"*
and warned that `page-count-20`'s green was **not** headroom. `TestNoRealToUnicodeSectionExceedsTheCap`
picked the four new goldens up by directory walk, in the ordinary suite:

```
DW-14 real-corpus witness: examined 13 fixture(s), 11 carrying at least one /ToUnicode section,
25 section(s) total; observed maximum section size = 55 (cap 100)
```

Per document, the three faces' section sizes are `[Noto Sans 55, Noto Sans SC 41, Noto Sans Thai 25]`
for `statement-5/20/50` and `[55, 1, 25]` for `statement-1`. **The corpus maximum rose from 45 to 55
and the 100-entry cap is NOT crossed.**

**Why, stated as the measurement rather than as a guess:** section size is driven by the number of
**distinct glyphs in a face's subset**, not by page count or row count. Fifty pages and 1085 rows add
no glyphs the first page did not already need.

**AND DW-14 IS DISCHARGED, NOT LEFT OPEN — the ruling at finish is D-4.7.4, and what it asks to be
recorded is the MECHANISM rather than the verdict.**

> **A `/ToUnicode` section's size follows the number of DISTINCT GLYPHS IN A SUBSET — not pages, not
> rows.**

That one sentence explains **both halves of DW-14's own warning**, which had been recorded as two
unconnected observations: *"twice the length in one face would reach the cap"* was right about the
direction and wrong about the axis (length is a poor proxy for distinct glyphs), and *"do not read
`page-count-20`'s green as headroom for 4.7's varied data"* was right — a repeated template has a small
glyph inventory — but 4.7's data turns out to vary in the wrong dimension: its **descriptions** vary,
its **alphabet** does not.

**What WOULD cross the cap**, so a future reader can recognise it: many distinct glyphs in ONE face.
Latin stops around 55 because ordinary business prose has about that many distinct characters. **A
CJK-heavy statement is the realistic case** — Chinese has no small alphabet, so genuinely varied Chinese
climbs one entry per distinct character. This family's SC subset is 41 from forty-one distinct glyphs;
two and a half times that vocabulary crosses 100.

**Log 55 as the observed maximum**, so the next movement is visible against a real number rather than a
prediction. And **the chunker is NOT dead code**: `internal/pdf`'s chunking (Story 4.2) is why a
document that does cross the cap emits a conformant file rather than a broken one, and it is exercised
by `internal/pdf/tounicode_chunk_test.go`. Discharging DW-14 discharges the WATCH, not the mechanism.
Recorded beside the entry in `deferred-work.md`.

### 8. DW-13 — measured, NOT adopted (Task 11, AC12)

Summing every `FontFile2` stream's `/Length1` in the committed goldens and compressing each program
with zlib level 9:

| document | file size | programs | uncompressed | Flate | saving |
|---|---|---|---|---|---|
| `statement-1` | 76,740 B | 3 | 65,740 B (85.7%) | 14,277 B | 51,463 B — 78.3% of program bytes, 67.1% of file |
| `statement-5` | 127,343 B | 3 | 77,452 B (60.8%) | 22,487 B | 54,965 B — 71.0% / 43.2% |
| `statement-20` | 269,804 B | 3 | 77,452 B (28.7%) | 22,487 B | 54,965 B — 71.0% / 20.4% |
| **`statement-50`** | **555,629 B** | **3** | **77,452 B (13.9%)** | **22,487 B** | **54,965 B — 71.0% / 9.9%** |

**The answer DW-13 asked for: 77,452 bytes of uncompressed `FontFile2` payload on the 50-page CJK
statement; Flate would save 54,965 of them, 9.9% of that document.**

**RE-MEASURED AT FINISH against the re-recorded goldens, and every figure is unchanged** — the
re-record moved digits, not glyphs, so the three `FontFile2` subsets are byte-identical to before.

**Recommendation: leave it. And D-4.7.5 rules that the reason to record is NOT the 9.9%.** The size
figures are real and they are all about size; recording them as the reason would leave a future reader
thinking a bigger saving would settle it. It would not.

> **The reason is byte-identity.** Compressing `FontFile2` makes every golden depend on **Go's
> `compress/flate` producing identical output across Go versions**, which is **not part of Go's
> compatibility promise** and has changed before.

**The risk is CROSS-UPGRADE, not cross-target, and that distinction decides what the matrix can tell
you.** The toolchain is pinned, so all four targets would compress identically and the matrix would
agree — green, for a reason unrelated to the property at stake. The failure arrives at the **next Go
bump**, silently moving every golden and presenting as a mysterious mass re-record rather than as a
toolchain event.

**RE-ENTRY IS A CONDITION, NOT A DATE.** Do not re-open on a schedule, and not because the size numbers
moved. Re-open when, and only when: *the compressed output of a FIXED input is pinned as a TEST-OWNED
LITERAL, compared against `compress/flate`'s live output, so a Go upgrade that changes the compressor
reddens a named test BEFORE it moves a single golden.* Without that pin the answer stays no regardless
of the size numbers.

The size argument, for completeness: DW-13's own threshold — *"tens of KB rather than hundreds"* — is
met at 76 KB. And a fact the entry could not have known: **the font payload is CONSTANT at 77,452 B across
the 5-, 20- and 50-page documents**, because the subsets are identical; the cost is per FACE, not per
page. So compression helps LEAST on the flagship document. Adoption would move every golden,
invalidate all three human sign-offs, and require narrowing a lint rule proved to fire. **Adoption
remains the owner's, batched at the Epic 4 close.** Nothing in `internal/pdf` or `lint/` was touched.
Recorded beside the entry in `deferred-work.md`.

### 9. DW-11 — answered in writing, and its forgetting-signal has FIRED (Task 11, AC12)

Written into `deferred-work.md` in 2.4's and 2.6's form. **No new items were added; the load-bearing
count remains 2; none were found and none were invented.** `corpus.json` and `cmd/gencorpus/main.go`
are byte-unchanged; `TestCorpusMeetsP6ExerciseFloors` still reports `got 7, need >=20` with the stats
line `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, character for character, in both
directions. The P6g floor was not touched.

**And the entry's own forgetting-signal is now true and is recorded as fired**, not quietly re-owned:
*"S4 still carrying only 2 genuinely-uncoverable, independently-attested opaque items when Epic 4's
golden report ships."* The golden report has shipped; S4 is still at 2.

This story adds **no** Thai judgment to S4 — Arm A reuses already-signed strings — so it adds no
attested name and no third sign-off.

### 10. AC8 — the sign-off is REQUESTED, and here is the request

**To the project owner.** Four table goldens are frozen and every machine-checkable property is
asserted. **The only non-circular check left is a person reading them**, and
`TestStatementSemanticSignOffIsRecorded` is red at the Epic 4 gate until that happens. Its failure
message carries the full instructions, the four file paths with page counts, the eight things to
confirm, and the JSON prefilled with the current digests.

**The trade-off, stated now rather than discovered later (D-000.41).** DW-13's adoption would move
every golden and invalidate this record. It is requested anyway, for the reasons the engineering lead
ruled: DW-13's adoption is a **decision, not a scheduled event**; 4.7 **is** the C4 gate, and *"frozen
with sign-off pending"* licenses a **transient** state, not the gate's **terminal** one.

**And the dilution worry inverts, which changes what is being asked for.** If DW-13 is adopted, the
re-record is a deliberate byte-moving change that **should not alter appearance** — so the second
attestation is the check that catches an unintended **visual** side-effect of it. **The second
signature answers a question the first cannot; the owner is not being asked to sign twice for the same
thing.** Recorded in the gate test's message and in all four fixture READMEs.

### 11. AC6(iii) — the independent reader (D-000.53)

`qpdf` **12.4.0** (`/opt/homebrew/bin/qpdf`). Verbatim invocation and output recorded in each
fixture's README. `qpdf --check` reports no syntax or stream encoding errors on all four;
`qpdf --show-npages` returns **1, 5, 20 and 50**, matching each document's declared page count exactly.

### 12. AC11 — no existing recorded bytes moved, and the CONTROL proves the green is evidence

The ten pre-existing golden digests are unchanged; every existing fixture test and
`TestGoldenDigestAgreesAtEveryDeclaredSite` are green, and all ten matrix documents still match.

**Control mutation** (`internal/pdf/builder.go`, `%PDF-1.7` → `%PDF-1.4` in the shared
template-render emitter). Reddened: `TestRenderMatchesFontTextGoldenFixture`,
`TestRenderMatchesImageEmbedGoldenFixture`, `TestMultiScriptFallbackGoldenFixture`,
`TestShapedTextGoldenFixture`, `TestThreeBandPageGoldenFixture`, `TestWrappedTextGoldenFixture`,
`TestStatementGoldenFixtures`. Stayed green: `TestRenderMatchesGoldenFixture` (minimal-rect renders
through `internal/pdf/document.go`, a different emitter — confirmed by mutating that one instead,
which reddened minimal-rect **alone**), `TestGoldenDigestAgreesAtEveryDeclaredSite` and
`TestEveryGoldenPDFResolvesItsPageTree` (both read the COMMITTED files, which an emitter change does
not touch — the correct division of labour).

**No AD-22 event occurred.** Nothing in this story moved an already-recorded byte.

### 13. The deletion screen (D-000.85)

**HEADING CORRECTED AT FINISH (Finding 2, a Major).** It read *"every observable, by name, reddened AND
stayed green"*, and that was **not true of AC1(ii)**: its declared deletion appears in no row of the
table below and among none of the replacements — it was neither performed nor declared unperformable.
D-000.79 is explicit that a mutation that cannot be performed as written is **said so and replaced by
name**, never silently dropped, and a heading claiming a red nobody observed is a false all-clear.
**The gap is closed below — the deletion has now been performed** — and the heading no longer claims
more than the table shows.

Every mutation was applied to a working copy and reverted with `cp` from a pristine mirror; the
committed goldens were never mutated. Each mutation asserted it applied by a reported diff or a
reported count, never by an exit code (D-000.40).

**Legend:** the AC column names the test; **R** = reddened, **·** = stayed green.

| mutation | AC1i | AC1ii | AC2i | AC2ii | AC2iii | AC2iv | AC3i | AC3ii | AC3iii | AC4 | AC5i | AC5ii | AC9 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| AC1(i) content-window advance +1 row | **R** | · | · | · | · | · | · | · | · | · | · | · | · |
| AC1(ii)‡ geometric substitution: the table replaced by spaced markers | · | **R** | — | — | — | — | — | — | — | — | — | — | — |
| AC2(i) suppress the header repeat | · | · | **R** | · | · | · | · | · | · | · | · | · | · |
| AC2(ii)* last row draws the previous row's Description | · | · | · | **R** | · | · | · | · | · | · | · | · | · |
| AC2(iii)* sum omits the collection's last element | · | · | · | · | **R** | · | · | · | · | · | · | · | · |
| AC2(iv) drop the confidentiality literal | · | · | · | · | · | **R** | · | · | · | **R** | · | · | · |
| AC3(i)* Thai+CJK removed from the three-script cell | · | · | · | · | · | · | **R** | · | · | · | · | · | · |
| AC3(ii)* shorten the wrapping cell to one line | · | · | · | · | · | · | · | **R** | · | · | · | · | · |
| AC3(iii) remove the logo element | · | · | · | · | · | · | · | · | **R** | · | · | · | · |
| AC4 date resolves from DATA not params | · | · | · | · | · | **R** | · | · | · | **R** | · | · | · |
| AC5(i)* Thai-only dictionary bypass † | · | · | · | · | · | · | · | · | · | · | **R** | · | · |
| AC5(ii) drop the GPOS x-offset | · | · | · | · | · | · | · | · | · | · | · | **R** | · |

`*` = **replacement mutation**, named below with the reason the story's own mutation could not be
performed as written (D-000.79 — never a silent substitution).
`†` = with one declared instrument relaxation, explained below.
`‡` = **performed at finish, and it is a FIXTURE SUBSTITUTION rather than a code mutation** — the
distinction the review asked to be recorded either way. `—` marks observables the substitution does not
address: it replaces the whole template, so the columns it does not touch are not evidence about it and
are not claimed as such.

**AC1(ii)'s declared deletion, PERFORMED (Finding 2).** The story names it: *"replace the table's bound
rows with fixed-position marker elements spaced one content window apart — `fixtures/page-count-20`'s
construction."* Performed as written, as a throwaway probe: for each document, a template carrying that
document's declared page count as markers at `y = i*682` (the statement's own 681.89pt content window),
the statement's own page geometry, page header and page footer, and the SAME data document. Measured:

```
AC1(i)  under the geometric substitution: statement-1   declared  1 page(s), produced  1  [PASS]
AC1(i)  under the geometric substitution: statement-5   declared  5 page(s), produced  5  [PASS]
AC1(i)  under the geometric substitution: statement-20  declared 20 page(s), produced 20  [PASS]
AC1(i)  under the geometric substitution: statement-50  declared 50 page(s), produced 50  [PASS]
AC1(ii) under the geometric substitution: statement-5    95 rows -> 5 page(s);  63 rows -> 5 page(s)  [FAIL]
AC1(ii) under the geometric substitution: statement-20  425 rows -> 20 page(s); 393 rows -> 20 page(s) [FAIL]
AC1(ii) under the geometric substitution: statement-50 1085 rows -> 50 page(s); 1053 rows -> 50 page(s) [FAIL]
RESULT: AC1(i) failures = 0 (it STAYED GREEN); AC1(ii) failures = 3 (it REDDENED)
```

**Exactly what the AC predicted**: a geometrically-built document satisfies *"renders at N pages"* at
all four counts while the data-dependence property fails outright, which is the whole reason AC1(ii)
exists. Probe deleted. **The reviewer's adjudication is not revised by this** — AC1(ii) was
**unwitnessed at review** and the ledger records it as such; what changed is that the gap is now
closed for the next audit rather than carried.

**AC6 (three observables).**
- **(i) three-site agreement** — mutated `fixtures/statement-20/expected.json`'s `sha256` to zeros.
  Reddened: `TestGoldenDigestAgreesAtEveryDeclaredSite`, `TestStatementGoldenFixtures`. **Stayed
  green: `TestEveryGoldenPDFResolvesItsPageTree`** — the bytes were untouched. That green is the
  discrimination.
- **(ii) the page tree resolves** — a **scratch copy** of `statement-20/expected.pdf` in `t.TempDir()`
  with the `/Kids` separator removed (`[8 0 R 10 0 R]` → `[8 0 R10 0 R]`, the exact multi-page
  defect). `folio.AssertPDFPageTreeResolves` **reddened**. **Stayed green:
  `TestGoldenDigestAgreesAtEveryDeclaredSite` and `TestEveryGoldenPDFResolvesItsPageTree`** — the
  committed golden was never touched. The temporary probe was deleted.
- **(iii) an independent reader resolves it** — **no deletion, by D-000.53's own guardrail** (*"this
  step is not discharged by an assertion this project wrote"*). Declared as an observable and reported
  without a mutation rather than dropped to make the count tidy. See §11.

**AC7 (two observables), the two-case split, run on the real matrix.**
- **(i) pairwise identity** — flipped one mid-file byte on the `js/wasm` leg (length-preserving, so
  `%%EOF` and the xref survive). Reddened with *"cross-target byte identity FAILED — targets disagree
  with EACH OTHER, not merely with the golden. This falsifies NFR1."*
  *(A first attempt appended a trailing byte and was caught earlier by `assertWellFormedPDF`'s
  `%%EOF` precondition — recorded because it is a real property of the harness, not a failed
  measurement.)*
- **(ii) agreement with the recorded golden** — perturbed `statement-20`'s recorded digest. Reddened
  with *"all four targets agree with EACH OTHER but NOT with the recorded golden"* — **pairwise
  identity stayed green.** The two cases are cleanly separated.

**AC8 (two observables).**
- **(i) the record exists and every field is populated** — **RESTATED AT FINISH (Finding 3, a Major),
  because the witness first claimed here was the wrong one.** The gate's current red is its
  `os.IsNotExist` branch, which `t.Fatalf`s and **returns**. Everything below it had therefore **never
  executed on any input**: the JSON unmarshal, the `reader`/`date`/`examined` non-empty loop, the
  `len(rec.Digests) == 0` check, and `assertStatementSignOffMatchesFrozenHashes`. The standing red
  witnesses that the FILE IS ABSENT, which is not the observable. A guard that has never run is not a
  guard — and had the wrapper's verdict been inverted (`if len(moved) == 0`), every test in the tree
  would still have been green. The asymmetry was unexplained: sibling AC8(ii) got a synthetic
  red-proof in the untagged file by exactly the technique that closes this.

  **Closed the same way.** `statementSignOffFieldProblems` and `statementSignOffGateVerdict` are now
  pure functions in the untagged `statement_golden_fixture_test.go`; the matrix gate calls them and
  prints what they return. `TestStatementSignOffFieldChecksAndGateWrapperRedProof` drives both over
  SYNTHETIC records (D-000.28 — nothing fabricates a human claim, and no file is written), control
  first:

  | case | expected |
  |---|---|
  | control: a fully populated record | **no problem reported** — without this, a check that rejected everything would satisfy the rest |
  | empty `reader` | exactly `empty reader` |
  | whitespace-only `date` | exactly `empty date` |
  | empty `examined` | exactly `empty examined` |
  | no digests at all | exactly `no digests` |
  | wrapper control: a current, in-scope record | **gate passes** |
  | wrapper: one moved digest of four | gate fails, message says `STALE`, names `statement-20`, says `1 of the 4 documents`, and carries D-4.7.7's LOOK-AT-THE-PAGES instruction |
  | wrapper: a record naming three of four | gate fails as a SCOPE MISMATCH, **not** as staleness |

  Each case asserts *exactly one* problem, so a check that fires on everything witnesses nothing.
  **Red-proved:** making `statementSignOffFieldProblems` return `nil` reddens four subtests with
  *"the field checks reported [] (0 problems), want exactly the one that was mutated"*; inverting the
  wrapper's `len(moved) != 0` to `== 0` reddens the wrapper control and the moved-digest case. Both
  mutations were reverted and the revert verified by `diff`.
- **(ii) the record is digest-bound** — witnessed permanently and **in the ordinary suite** by
  `TestStatementSignOffDigestBindingRedProof`, over synthetic records (D-000.28: nothing here
  fabricates a human claim). Four cases: a control that must be ACCEPTED; one moved digest of four
  invalidating the whole record; a record naming only three of four rejected outright; a record naming
  an unregistered document rejected outright. **The control is what makes the other three mean
  something** — without it, a function that rejected everything would pass.
  **(i) stays green under (ii)'s mutation** by construction: the pure comparison never looks at
  presence or field population.

**AC9 (one observable).** Removed `table_footer_test.go` from `declaredTableBehaviourSuite` while
leaving the file on disk. `TestTableBehaviourSuiteIsNotSupersededByTheGolden` **reddened**:
*"table_footer_test.go is a table behaviour test file on disk and is NOT in
declaredTableBehaviourSuite."* Everything else stayed green.
**This mutation only reddens because the guard was strengthened to a two-way SET EQUALITY during the
screen** — as first written it only checked that declared files were present, so shortening the list
would have made it check fewer things and still report green. The screen found that.

**AC10 — ONE observable, corrected at finish from a declared zero (Finding 6, a Major).** This
paragraph originally read *"AC10, AC12 — zero observables, and the zero is honest… No observable was
invented to fill them"* and then, in the same breath, recorded that AC10's written record **is** guarded
by `TestStatementFixtureReadmesRecordTheCoverageHistory` (whitespace-normalised, so a required sentence
falling across a line wrap still counts). It is a present, live, separably-witnessed machine observable
by every criterion applied to the other twenty-one, and the reviewer red-proved it: re-spelling the
required sentence in `fixtures/statement-20/README.md` (`reddened NO GOLDEN AT ALL` -> `reddened
NOTHING WHATSOEVER`) reddens the guard **on statement-20 alone**, the other three subtests green.
Declared total corrected to **22**.

**AC12 — zero observables, and this zero survives audit.** No machine guard exists over the DW-11 and
DW-13 written answers, and none was invented to fill the gap.

#### Replacement mutations, each with the reason (D-000.79)

- **AC2(ii)** — the story names *"drop the last row's text runs from the page assignment."* **Not
  performable: the renderer refuses it.** Applied at `layout.Paginate`, the render failed with
  *"element e8: a column item carries neither a text run, an image nor a table rect, so it would
  occupy column space while drawing nothing — an invisible item…"* That is a stronger result than the
  screen asked for: **the defect this mutation describes is structurally impossible at that seam.**
  Replacement: the last row draws the **previous** row's Description (an off-by-one in row
  projection). Isolated perfectly.
- **AC2(iii)** — the story names *"emit the footer aggregate on every page instead of the last."*
  Doing so requires duplicating the aggregate item per page, a substantial change to the aggregate's
  placement rather than a deletion. Replacement: **the sum aggregate omits the collection's last
  element** — a live money-path defect, exactly the shape AD-23 exists for. Reddened AC2(iii) alone.

  **AND THE REPLACEMENT MOVES TO A DIFFERENT SUB-CLAIM, which the reviewer adjudicated as Class A
  (Finding 7).** AC2(iii) is compound — *"the sum footer's single occurrence, its page, and its
  value"* — and the declared mutation was recorded as reddening *"on both the count and the page
  index"*. The replacement reddens the **value** comparison only. The assertion itself still covers all
  three limbs, so the exposure is bounded, but the screen no longer demonstrates that the *"exactly
  once, on the last page"* limb has teeth.

  **The reviewer's suggested cheaper alternative was ATTEMPTED AT FINISH and does not isolate.** Moving
  the aggregate item to the head of the column (`rowTop = tableBottom` before the footer row is built)
  was applied to a working copy and measured: on `statement-1` the footer is legitimately on its only
  page and AC2(iii) **passes**; on the three multi-page documents the render raises
  `TABLE_FOOTER_ORPHAN_SUPPRESSED` and AD-14's shared no-unasserted-diagnostic precondition fatals
  **before the placement assertion is reached** — and the same mutation also reddens AC1(i), AC2(i) and
  AC2(ii), because moving the aggregate to the table's head collapses the layout. So it is not a
  separating mutation either. **The measurement confirms the adjudication rather than overturning it:
  AC2(iii)'s placement limb is not separably witnessable with the tools this story has, and it is
  recorded as compound-observable Class A.** Mutation reverted; `git status` clean on
  `folio-go/table_render.go`.
- **AC3(i)** — the story names *"collapse the fixture's font stack to the Latin face alone."*
  Performed, and it reddens **eleven of twelve observables**, because the missing Thai and CJK
  coverage raises a diagnostic and AD-14's shared precondition fatals every render-based test.
  Replacement: **remove the Thai and CJK characters from the three-script row's Description cell** —
  the actual deletion of the observable. Isolated perfectly.
- **AC3(ii)** — the story names *"widen the wrapping column so its content fits on one line."*
  Performed, and it reddens AC2(ii), AC5(i) and AC5(ii) as well, because widening a column moves
  **every** column origin and the cell-locating assertions follow the geometry. Replacement:
  **shorten the wrapping cell's text so it fits on one line.** Isolated perfectly.
- **AC5(i)** — the story names *"bypass the dictionary breaker for table-cell text."* A global bypass
  changes Latin and CJK breaking too; scoped to Thai, the unbroken string **overflows the Note column
  and raises `TEXT_CLIPPED_WIDTH`**, which AD-14's shared "no unasserted diagnostic" precondition
  turns into a fatal on every render-based test. **† With that one shared precondition relaxed to a
  log — a declared instrument change, reverted immediately — AC5(i) reddens ALONE on its own break
  comparison, and AC5(ii) stays green**, exactly as the story predicted. Both readings are reported;
  neither is hidden.

#### Two couplings the screen found IN MY OWN ASSERTIONS, and fixed

The screen's value on this story was not confirming what was predicted; it was finding two places
where the observables were **not separably witnessable as first written**.

1. **AC1(ii) pinned the literal 22** ("removing one continuation page's rows costs exactly one page").
   The paginator mutation therefore reddened AC1(i) **and** AC1(ii), destroying the very
   discrimination AC1(ii) exists to provide. Rewritten to measure the rows the last two pages actually
   carried **off the artifact** and assert the page count strictly decreases — a genuinely relative
   property that survives a uniform capacity shift. It also had to stop asserting an exact drop of
   one, because Story 4.5's footer orphan-avoidance can legitimately hold a final page open for the
   total alone (measured: removing `statement-5`'s last 10 rows left it at 5 pages, the last carrying
   the aggregate and no rows — correct behaviour that the earlier assertion called a failure).
2. **AC4 asserted the whole confidentiality sentence**, so AC2(iv)'s mutation reddened it too. Rewritten
   to locate the run **by the date itself**. AC2(iv) owns the wording; AC4 owns the provenance.
   *(AC4's own mutation still reddens AC2(iv), and that is inherent: the mutation edits the very
   literal AC2(iv) asserts. The asymmetry is now in the right direction.)*

A third, smaller one: **AC3(iii) was pinned to the declared page count**, so the paginator mutation
reddened it. It now asserts one definition and one reference **per page of this render** — relative,
which is what "one definition, many references" actually claims.

### 13a. The FINISHER's own deletion screen — every guard added or changed at finish (D-000.85, D-000.86)

Same discipline as §13: applied to a working copy, reverted with `cp` from a pristine mirror, the
revert verified with `/usr/bin/diff` (never an exit code, and never `diff`'s wrapped form — see the
note at the end of this section), and each mutation confirmed applied by a printed diff before the run.
Every new test also carries **D-000.86 part (a)**: whether the property was already true at HEAD.

| # | guard | mutation | result |
|---|---|---|---|
| M1a | `TestStatementMoneyColumnDiscriminatesFloatArithmetic`, value-class limb | one filler cent part back to `25` | **RED** — *"amount \"20.25\" (row TXN-0010) is an exact multiple of 0.25, so it IS exactly representable in binary64…"* |
| M1b | …its CONTROL limb | the control's filler changed from `.25` to `.24` | **RED** — *"control: \"16.24\" is not a multiple of 0.25, but the value class this family was FIRST recorded with was entirely quarter-integral — the control does not reproduce the state it is meant to describe"* |
| M1c | …its pin's staleness half | `statement-1`'s pinned exact total `185200 -> 185201` | **RED** — *"the EXACT total is now 185200 minor units (1,852.00), but this pin records 185201 (1,852.01)"* |
| M1d | …its pin's discrimination half | `statement-50`'s pinned binary total set equal to the exact one | **RED** — *"the pinned binary64 total \"602,408.68\" is EQUAL to the exact total, so this document's 1085 amounts cannot tell an exact money path from a binary one"* |
| M2a | `TestStatementSignOffFieldChecksAndGateWrapperRedProof`, field half | `statementSignOffFieldProblems` returns `nil` | **RED** on four subtests — *"the field checks reported [] (0 problems), want exactly the one that was mutated"* |
| M2b | …its wrapper half | `if len(moved) != 0` inverted to `== 0` | **RED** on the wrapper control and the moved-digest case |
| M3 | `TestStatementSubprocessSelectorsCoverEveryFixture` | `statement-5`'s selector rows `95 -> 96` | **RED** — *"the subprocess selector renders a collection of 96 row(s) and statementFixtures records 95"* |
| M4 | AC3(i) restricted to the Description CELL | row 1's Thai+CJK moved out of Description into the Note cell | **RED** on all four — and **the pre-fix ROW-scoped assertion PASSES on the same mutated fixture**, which is the point |
| M5 | AC5(ii)'s SIGNED GPOS check | `internal/pdf/textdoc.go`'s emitter mirrored (`adjustments[i] += g.XOffset`, `[i+1] += -g.XOffset + …`) | **RED** on all four — *"first non-zero TJ adjustment is 21 at index 0 … emits as a BEFORE term of -21"* — and **the pre-fix MAGNITUDE-ONLY assertion PASSES on the same mirrored renderer**, producing `[21 -21]`, exactly as Finding 9 predicted |
| M6 | AC2(ii)'s new no-wrapping-row invariant | `TXN-0005`'s Description replaced by the wrapping text, so the LAST row wraps | **RED** — *"last row (TXN-0005) is the WRAPPING row… needs this assertion rewritten, not special-cased"* |
| M7 | `mpParseToUnicode`'s new multi-face fatal (D-4.7.6) | guard neutered with `&& false`, then called on `statement-1` | returns silently and decodes as **`"ไustomerะบdaบศovelace"`**; with the guard in place it fatals with the actionable message |

**D-000.86 part (a), per new test — was the property already true at HEAD?**

- **`TestStatementMoneyColumnDiscriminatesFloatArithmetic`: NO, and emphatically.** At HEAD every one of
  the 1085 amounts was a multiple of 0.25, so limb (1) would have reddened on all of them and the
  control reproduces the HEAD state exactly (`602,061.50` under both paths — the same figure the review
  measured). This guard was **false at HEAD and is true now**, which is the strongest form part (a) can
  take.
- **`TestStatementSignOffFieldChecksAndGateWrapperRedProof`: the functions did not exist at HEAD.** The
  checks existed as unreachable inline code below a returning `t.Fatalf`, so the property was neither
  true nor false — it was unassessed. That is the finding.
- **`TestStatementSubprocessSelectorsCoverEveryFixture`: YES, already true at HEAD** — the two lists did
  agree. Stated plainly rather than implied: this guard's green on the day it landed says nothing that
  was not already so. Its value is entirely in M3, which is what gives the comment above it something
  to point at.

**Not red-proved, and said so rather than left to look proved.** Finding 16 (the Amount cell's x-range
lower bound moved from `+31000` to the Note column's declared `+34000`) is a **latent** correction: Note
cells are left-aligned and their runs begin at the column origin, so nothing is drawn in the 3,000 mp
overlap today and no mutation short of changing the Note column's alignment can redden it. The only
evidence offered is that every Amount-cell assertion still passes with the tighter bound. Finding 12
(the `table_*_test.go` glob's scope) is a documentation change to a file header and adds no mechanism.

**One process note, because it bit and it is the kind of thing that reads as a clean run.** A revert
issued as `/bin/cp $SP/rp/textdoc.go folio-go/internal/pdf/textdoc.go` from a shell whose cwd had been
changed by an earlier `cd folio-go` in the same command **failed silently in the `&&` chain and left
BOTH the emitter and the assertion mutated**. It was caught by re-running the revert with absolute paths
and confirming with `git status --porcelain` on the file. Every revert in this section was verified by
`/usr/bin/diff` against the mirror, and the whole tree was re-verified with `git status` afterwards:
only the intended files are modified, and no file under `folio-go/internal/` is touched at all.

### 14. Observable ledger (D-000.88) — n=2 of the pre-registered series, ADJUDICATED

**THE FIGURES BELOW ARE THE REVIEWER'S ADJUDICATION, RECORDED VERBATIM. The finisher does not revise
them** — the measured party moving its own denominator is the gaming risk D-000.88 exists to close, and
that applies to the finisher exactly as it applies to the developer.

> **Declared 21 (UNDERSTATED BY 1 — Finding 6) / audited: 21 present, 20 separably witnessed /
> Class A 4.**

| sub-cause | n | ACs |
|---|---|---|
| prose-claim | 1 | AC6(iii) — as the developer proposed, and correct |
| compound-observable | 2 | AC8(i) (as proposed) **+ AC2(iii)**, whose replacement mutation moved to a different sub-claim (Finding 7) |
| unwitnessed — declared deletion never performed | 1 | **AC1(ii)** (Finding 2) |

**The developer proposed *declared 21 / audited 21 / Class A 2*. The reviewer adjudicated Class A 4 —
not the proposed 2 — and that difference IS D-000.88 working**; the audit is what makes this point in
the series comparable to the last.

**Series:** 4.6 declared 15 / audited 14 / Class A 6. **4.7 declared 21 / audited 21 present, 20
separably witnessed / Class A 4.**

**Note the direction of the declaration error.** The declared count was **understated**, not inflated —
which is the honest direction, and is still a miss: an unclaimed observable shrinks numerator and
denominator together and flatters the ratio. The corrected declaration in the AC section reads **22**.

**Two things the finisher did that do NOT change the figures above, recorded separately so they are not
mistaken for a revision of the audit:**

1. **AC1(ii)'s declared deletion was performed** (§13) and reddens as the AC predicted, with AC1(i)
   green. It was **unwitnessed at review** and the adjudication stands as written; the gap is closed
   for the NEXT audit, not retroactively.
2. **AC8(i)'s mechanical half now has a red-proof** (§13). Its **compound-observable** classification is
   unaffected and correct: the schema witnesses presence and non-emptiness, and **specificity cannot be
   machine-checked at all** — the reviewer reading the `examined` field is the only thing that catches
   *"looked fine"*.
3. **AC2(iii)'s placement limb was re-attempted and remains unwitnessable** (§13). The Class A
   classification is confirmed by measurement.

**And one observable was ADDED at finish, declared here rather than folded into the adjudicated
figures**: `TestStatementMoneyColumnDiscriminatesFloatArithmetic` (Finding 1's fix), which asserts the
money column's value class can express a decimal-arithmetic defect. It belongs to AC2(iii)/AD-23 and
carries its own red-proofs (§13a). **A future audit's denominator is 23, not 22.** It is not counted
into this story's adjudicated ledger, because the reviewer adjudicated what the reviewer saw.

### 15. Gates — pass / fail / skip reported SEPARATELY (D-000.89)

**`folio-go`**, `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` (from `folio-go/`):

| population | pass | fail | skip | total |
|---|---|---|---|---|
| top-level | 728 | 1 | 4 | **733** |
| including subtests | 1168 | 2 | 4 | **1174** |

Against the baseline's **716** / **1094**: **+17 top-level, +80 including subtests.**

**AT FINISH, same invocation, re-measured** (D-000.89 — pass / fail / skip reported separately, never a
passed count labelled a total):

| population | pass | fail | skip | total |
|---|---|---|---|---|
| top-level | **731** | **1** | **4** | **736** |
| including subtests | **1183** | **2** | **4** | **1189** |

**+3 top-level over the review state, attributed BY NAME rather than inferred from the totals** — the
two top-level test-name sets were differenced, and the difference is exactly:
`TestStatementMoneyColumnDiscriminatesFloatArithmetic`,
`TestStatementSignOffFieldChecksAndGateWrapperRedProof`,
`TestStatementSubprocessSelectorsCoverEveryFixture`. **Nothing was removed.**

**Skips at finish, as a SET by name — unchanged, no member added or lost:**
`TestXrefEntriesRejectsMalformedSubprocess`,
`TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`,
`TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`,
`TestTwoTablesWithPageCountFooterRenderConsistently`.

**Red by design at finish, named separately, and still the ONLY failure:**
`TestCorpusMeetsP6ExerciseFloors`, subtest `P6g (opaque names)`, `got 7, need >=20`, stats line
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — unchanged character for character. The P6g floor
was not touched.

**`lint` at finish**, same invocation from `lint/`: top-level **77** (77 / 0 / 0); including subtests
**115** (115 / 0 / 0) — unchanged. **`hashmatrix` at finish**, from `hashmatrix/`: top-level **3**
(3 / 0 / 0); including subtests **3** — unchanged.

**And two lint/arch guards were measured FIRING during the finisher pass, then satisfied rather than
worked around** (see D-4.7.8): `TestNoFloat64UnderModule` on a first draft of the money guard, and
`TestBigFloatTypeTestScopeInventory` on a second. Both are green in the figures above because the guard
was rewritten to assert the property in exact integer arithmetic, not because either was relaxed.

**Matrix at finish**, `env CGO_ENABLED=0 GOWORK=off go test -tags matrix -count=1 -timeout 120m -run TestCrossTargetByteIdentity .`
— `--- PASS: TestCrossTargetByteIdentity (18.62s)`, 14 documents x 4 targets, every statement leg equal
across all four targets **and** equal to the RE-RECORDED golden; the ten pre-existing documents still
match their own goldens. `go vet ./...` and `go vet -tags matrix ./...` both clean.

**Skips, as a SET by name — unchanged from baseline, no member added or lost:**
`TestXrefEntriesRejectsMalformedSubprocess`,
`TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`,
`TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`,
`TestTwoTablesWithPageCountFooterRenderConsistently`.

**Red by design, named separately, and it is the ONLY failure:**
`TestCorpusMeetsP6ExerciseFloors` — subtest `P6g (opaque names)`, `got 7, need >=20`, stats line
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, unchanged character for character. The other
six floors pass.

**`lint`**, same invocation from `lint/`: top-level **77** (77 / 0 / 0); including subtests **115**
(115 / 0 / 0).

**`hashmatrix`**, same invocation from `hashmatrix/`: top-level **3** (3 / 0 / 0); including subtests
**3**.

**Matrix suite**, `env CGO_ENABLED=0 GOWORK=off go test -tags matrix -count=1 -timeout 120m -run TestCrossTargetByteIdentity .`
— `--- PASS: TestCrossTargetByteIdentity (18.57s)`, 14 documents x 4 targets. `go vet -tags matrix ./...`
is clean.

**Heavy tests, BOTH directions** (they were not touched by this story, but the four skips are named in
the gate obligations and a skip nobody has ever seen pass is not a test):

- unset — `--- SKIP` for all three, in the run above.
- `env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -count=1 -v -timeout 120m -run 'TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable|TestTwoTablesWithPageCountFooterRenderConsistently|TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability' ./...`
  — `--- PASS` for all three. **None has an empty body under an unconditional skip.**

**`TestOverTallSingleRowStillOverflows` — green and unchanged** (D-4.5.4 as amended). Verified in the
final run; the file was not edited.

**DW-21 is NOT closed by this story.** Its heavy command belongs to the Epic 4 boundary gate **and**
the orchestrator's checklist, two addresses on purpose. It is stated here so the gate does not have to
remember it unaided; the run above is evidence the command works, **not** a discharge of DW-21.

### 16. Second-producer witness (AC9)

`TestTableBehaviourSuiteIsNotSupersededByTheGolden` reports, in the ordinary suite:

```
second-producer witness — 6 declared table behaviour files present, 90 top-level tests between them:
[table_footer_test.go=28 table_header_repeat_test.go=14 table_pagination_test.go=8
 table_render_row_test.go=12 table_render_test.go=17 table_row_clip_test.go=11]
```

**This independently reproduces the story's corrected table — 90, not 46 — from the tree rather than
from the story file.** The guard is written against the SET OF NAMES, in both directions, and never
against a count. It catches DELETION and it does **not** catch ROT; that limitation is stated in the
file itself rather than papered over.

### 17. Files changed

**Added** — `fixtures/statement-{1,5,20,50}/{input.folio,data.json,params.json,expected.pdf,expected.json,README.md}`;
`folio-go/statement_fixture_test.go`, `folio-go/statement_semantics_test.go`,
`folio-go/statement_golden_fixture_test.go`, `folio-go/statement_signoff_matrix_test.go`,
`folio-go/table_behaviour_suite_test.go`.

**Modified** — `folio-go/byte_neutrality_test.go` (four `goldenDigestRecord` entries, five
`declaredEpic2GateObligations` lines), `folio-go/matrix_test.go` (four `matrixDocuments` entries, the
parameterised capture and the per-document feature guard), `folio-go/render_test.go` (four env-gated
subprocess selectors), `.github/workflows/matrix.yml` (the `docs=` line and sixteen upload paths),
`_bmad-output/implementation-artifacts/deferred-work.md` (DW-11's answer, DW-13's measurement),
`_bmad-output/implementation-artifacts/sprint-status.yaml`.

**Changed at finish** — `folio-go/statement_fixture_test.go` (the amount value class and
`statementFillerCents`; the corrected row arithmetic; `statement-20`'s rationale; `thai-002`'s measured
reason), `folio-go/statement_semantics_test.go` (the money-discrimination guard; AC3(i) restricted to
the Description cell; AC5(ii)'s signed GPOS check; AC2(ii)'s no-wrapping-row invariant and the Amount
cell's derived left bound; `statementRowsOnFirstPage`'s cross-reference),
`folio-go/statement_golden_fixture_test.go` (`statementSignOffFieldProblems`,
`statementSignOffGateVerdict`, `statementSignOffReadingRemedy` and their red-proof),
`folio-go/statement_signoff_matrix_test.go` (the gate now calls those, and carries D-4.7.7's
instruction), `folio-go/render_test.go` (`TestStatementSubprocessSelectorsCoverEveryFixture`),
`folio-go/multi_page_fixture_test.go` (D-4.7.6's multi-face fatal),
`folio-go/table_behaviour_suite_test.go` (the glob's declared scope limitation),
`folio-go/byte_neutrality_test.go` (four re-recorded digests; `statement-20`'s rationale), all four
`fixtures/statement-*/{data.json,expected.pdf,expected.json,README.md}`, and
`_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` (D-4.7.8, D-4.7.9, appended).

**Not touched, deliberately** — any existing `expected.pdf` or digest, `altRowBackground`,
`lint/internal/rules/nocompressor.go`, the P6g floor, `corpus.json`, `cmd/gencorpus/main.go`, the four
named skips, and `TestOverTallSingleRowStillOverflows`.

### 18. Open, and named so nothing is quietly inherited

1. **The human sign-off is outstanding, and it is now being requested against the RE-RECORDED files.**
   `TestStatementSemanticSignOffIsRecorded` is red at the Epic 4 gate. **This is the only non-circular
   check on this recording.** Until it lands, four table goldens are frozen that no person has read.
   The four digests a signer will bind to are the finisher's (`ef58bbf6…`, `7f67b317…`, `be6f5e27…`,
   `9c5be7ba…`), not the review's — which is exactly why the owner was asked to HOLD the sign-off until
   the money-column Blocker landed. **No attestation was invalidated, because none existed.**
2. **`mpParseToUnicode` now FATALS on a multi-face document (D-4.7.6)** rather than returning garbage —
   §4. The underlying merge defect is still there; what changed is that it can no longer be reached
   silently. Fixing the merge itself is still out of this story's scope.
3. **DW-14 is DISCHARGED (D-4.7.4)**, with the mechanism recorded and **55** logged as the observed
   corpus maximum. The chunker's boundary is still not live in any committed golden, and the chunker is
   **not** dead code.
4. **DW-13 stays open, NOT adopted (D-4.7.5)** — with a measurement, a recommendation, and a **re-entry
   CONDITION rather than a date**: pin the compressed output of a fixed input as a test-owned literal
   first. Adoption remains the owner's.
5. **DW-11 stays open at 2, with its forgetting-signal FIRED.**
6. **DW-21 stays open**; its gate run is not discharged here.
7. **The `signoff` digest site is undeclared** on all four `goldenDigestRecord` entries until the
   record exists — by design, and forced by the completeness guard the moment it does.
8. **AC2(iii)'s "exactly once, on the last page" limb has no separating mutation**, and two candidates
   were tried (§13). It is recorded as compound-observable Class A rather than papered over.
9. **AC6(iii) remains a prose claim.** `qpdf`'s transcript in each README can go stale without anything
   reddening; D-000.53 makes that deliberate, and the ledger says so.
10. **The money column discriminates the scale-to-minor-units float path, not every float path.** A
    binary path that formats by rounding stays invisible at this magnitude, and no fixture here could
    change that. Recorded in all four READMEs' *what it cannot express*.

### 19. Finisher pass (2026-08-27)

**Triage: 16 findings — 16 FIX, 0 DISMISS, 0 DEFER.** 1 Blocker, 5 Majors, 9 Minors, 1 Nit. Every one
had a concrete, in-scope, cheaply-verifiable resolution; none required expanding the story's scope. Two
were resolved by MEASUREMENT going against the finding's own suggested remedy (Findings 7 and 13) and
are recorded that way rather than as clean agreements — see *Finding Resolutions* at the end of this
file.

**The Blocker first, because it outranked everything and forced the rest of the order.** All 1085
amounts had cent parts in `{0, 25, 50, 75}` — exactly representable in binary floating point — so all
four goldens would have been byte-identical under a float money path and AC2(iii) would still have
passed. The value class was changed, all four documents re-recorded, every digest site updated, the
matrix re-run, and the discrimination pinned by a new guard. Full reasoning, the AD-22 declaration, the
old-to-new digest table and the three-guards-forbid-writing-the-wrong-implementation finding are in
**D-4.7.8**, appended to the decision log.

**The float64 discrimination, demonstrated.** Two independent implementations of the naive money path
(CPython's `float`, and `math/big.Float` at 53 bits with `ToNearestEven`) agree to the digit:

| document | exact-decimal total | binary64 total | Amount cells that diverge |
|---|---|---|---|
| `statement-1` | `1,852.00` | **`1,851.99`** | 2 of 5 |
| `statement-5` | `7,348.39` | **`7,348.38`** | 5 of 95 |
| `statement-20` | `96,801.82` | **`96,801.81`** | 34 of 425 |
| `statement-50` | `602,408.68` | **`602,408.67`** | 78 of 1085 |
| **control — the ORIGINAL value class** | `602,061.50` | `602,061.50` | **0 of 1085** |

The control reproduces the reviewer's own figure exactly, which is what makes the rest of the table
mean something.

**Files this pass modified or added** — see §17's *Changed at finish*. **Nothing under
`folio-go/internal/` was changed**: `internal/pdf/textdoc.go` was mutated for one red-proof and
restored, verified byte-identical and confirmed clean by `git status --porcelain`.

**Gates, before and after** (D-000.89 — pass / fail / skip separately, with the exact invocation; never
a passed count as a total). `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` from each module:

| module | at review | at finish |
|---|---|---|
| `folio-go` top-level | 733 (728 / 1 / 4) | **736 (731 / 1 / 4)** |
| `folio-go` incl. subtests | 1174 (1168 / 2 / 4) | **1189 (1183 / 2 / 4)** |
| `lint` top-level | 77 (77 / 0 / 0) | **77 (77 / 0 / 0)** |
| `lint` incl. subtests | 115 | **115** |
| `hashmatrix` top-level | 3 (3 / 0 / 0) | **3 (3 / 0 / 0)** |

The **+3** top-level are named in §15, by differencing the two test-name sets rather than inferring from
totals. Skip set unchanged as a set. Sole failure remains the red-by-design
`TestCorpusMeetsP6ExerciseFloors`.

**Matrix RE-RUN in-story after the re-record** (D-000.4 override, D-000.54, and mandatory here because
the bytes moved): `--- PASS: TestCrossTargetByteIdentity (18.62s)`, 14 documents x 4 targets, all four
statement documents identical across `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm` and
equal to the re-recorded goldens. The ten pre-existing documents still match their own goldens.

**Heavy tests, BOTH directions, re-run at finish.** Unset: `--- SKIP` for all three, in the run above.
Set:

```
env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -count=1 -v -timeout 120m \
  -run 'TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable|TestTwoTablesWithPageCountFooterRenderConsistently|TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability' ./...
```

`--- PASS` for all three. **DW-21 is still NOT closed by this story** — its gate run belongs to the
Epic 4 boundary gate and the orchestrator's checklist, two addresses on purpose.

**AD-22, declared.** The four statement goldens moved, deliberately, for the reason in D-4.7.8. **No
OTHER recorded bytes moved:** `git diff df8cbcc -- fixtures/` touches only `fixtures/statement-*`, the
ten pre-existing digests are unchanged at all their declared sites, and
`TestGoldenDigestAgreesAtEveryDeclaredSite` is green.

**`TestOverTallSingleRowStillOverflows` is green and unedited** — `git diff df8cbcc --
folio-go/table_row_clip_test.go` is empty.

**Red-proof discipline.** Eleven mutations in §13a's table — two of them run TWICE, once against the
fixed assertion and once against the pre-fix one, because "the fix has teeth" is a different claim from
"something reddened" — plus AC1(ii)'s performed fixture substitution and AC2(iii)'s attempted placement
mutation. Each was applied to a working copy, confirmed applied by a printed diff before the run,
reverted, and the revert verified with `/usr/bin/diff` against a mirror — never by an exit code. Reddened and stayed-green are reported by name in §13a, including the
two mutations that did **not** produce the hoped-for result and are reported as failures to isolate
rather than quietly dropped.

**Two rulings were appended to the decision log rather than editing anything in place** (it is
append-only): **D-4.7.8** (the re-record, the AD-22 event, and the guards that forbid writing the wrong
implementation in-tree) and **D-4.7.9** (`thai-002`'s reason withdrawn and replaced by measured advance
widths, correcting the closing paragraph of D-000.93 without touching it).

### 20. Done — owner sign-off bound and final gates (2026-08-27)

Baseline remains `df8cbcc`. Panit Wechasil visually inspected the final rendered pages and confirmed
the eight required semantic points. `fixtures/statement-signoff.json` records that examination without
text extraction and binds all four current digests; the same file is now a declared `signoff` site for
all four statement entries. The first ordinary run caught one real finishing defect: the shared digest
site reader understood only a single top-level `sha256`, not D-4.7.1's one-record-over-four `digests`
shape. The reader now accepts both established forms and checks the map entry for the current fixture.

All nine Story 4.7 rulings were applied: D-4.7.1–D-4.7.9. Review triage is **16 FIX / 0 DISMISS /
0 DEFER**, with one explicit resolution per finding in *Finding Resolutions*. No new work was deferred.
Existing DW-11 remains owned by later sourced-corpus work, DW-13's adoption remains a project-owner
decision subject to D-4.7.5's re-entry condition, and DW-21 remains owned by the Epic 4 boundary gate;
none is silently claimed as closed here.

Measured final gates:

- Ordinary `folio-go` suite: top-level **731 pass / 1 fail / 4 skip / 736 total**; including subtests
  **1183 pass / 2 fail / 4 skip / 1189 total**. The sole designed failure is
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (`got 7, need >=20`); its top-level parent is the
  second failure entry in the inclusive count. No other test failed. The four named heavy tests were
  skipped under the ordinary invocation, as designed.
- `lint`: top-level **77 / 77 pass**; including subtests **115 / 115 pass**. `hashmatrix`: **3 / 3
  pass**.
- Human sign-off gate: `TestStatementSemanticSignOffIsRecorded` **PASS**, reader `Panit Wechasil`,
  date `2026-08-27`, four digests bound.
- Story 4.7 matrix override: `TestCrossTargetByteIdentity` **PASS (19.69s)**, fourteen documents across
  `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm`; every statement leg equals its signed
  golden and the ten pre-existing documents remain unchanged.
- Lint/build gates: `go vet ./...`, `go vet -tags matrix ./...`, and `go build ./...` pass for
  `folio-go`; `go vet ./...` and `go build ./...` pass for `lint`; the `hashmatrix` probe builds and its
  test-only `floatdiscrimination` package compiles as a test binary. `git diff --check` and `gofmt -d`
  are clean.

Commit: **PENDING** — replaced by the scoped local commit reported by the finisher; no second
prose-only commit will be created solely to rewrite this line.

---

## Review Findings

## Review Summary
- Reviewed by: bmad-code-reviewer
- Date: 2026-08-27
- Story Status Recommendation: **Changes Requested**
- Blockers: 1
- Majors: 5
- Minors: 9
- Nits: 1

**Observable ledger — ADJUDICATED (D-000.88).** The developer proposes *declared 21 / audited 21 /
Class A 2*. The reviewer's adjudication is **declared 21 (understated by 1 — see Finding 6) /
audited 21 present, 20 separably witnessed / Class A 4**:

| sub-cause | n | ACs |
|---|---|---|
| prose-claim | 1 | AC6(iii) — as proposed, and correct |
| compound-observable | 2 | AC8(i) (as proposed) **+ AC2(iii)** (Finding 7) |
| unwitnessed — declared deletion never performed | 1 | **AC1(ii)** (Finding 2) |

**Series:** 4.6 declared 15 / audited 14 / Class A 6. **4.7 declares 21 / audited 21 / Class A 4.**
The Class A figure is the reviewer's, not the developer's proposal, which is D-000.88 working.

### What was verified INDEPENDENTLY and reproduces exactly

Every figure below was re-derived by the reviewer, not read off the Delivery Log.

- **Gates**, `env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` — `folio-go` top-level
  **733 (728 pass / 1 fail / 4 skip)**, incl. subtests **1174 (1168 / 2 / 4)**; `lint` **77 (77/0/0)**,
  115 incl. subtests; `hashmatrix` **3 (3/0/0)**. Skip SET by name matches the baseline exactly, no
  member added or lost. Sole failure is the red-by-design `TestCorpusMeetsP6ExerciseFloors`
  (`P6g got 7, need >=20`).
- **Matrix legs re-run in review**: `go test -tags matrix -run TestCrossTargetByteIdentity` —
  `--- PASS (17.29s)`, **14 documents x 4 targets**, all four statement hashes identical on
  darwin/arm64, linux/amd64, linux/arm64 and js/wasm **and** equal to the recorded goldens; the ten
  pre-existing documents still match.
- **No AD-22 event.** `git diff df8cbcc -- fixtures/` is empty: not one pre-existing recorded byte moved.
- **AC11's control reproduces exactly.** `%PDF-1.7 -> %PDF-1.4` in `internal/pdf/builder.go` reddened
  precisely the seven named tests and left `TestRenderMatchesGoldenFixture`,
  `TestGoldenDigestAgreesAtEveryDeclaredSite` and `TestEveryGoldenPDFResolvesItsPageTree` green. The
  green is evidence, not absence.
- **AC1(ii)'s coupling fix is a real repair, not an un-reddening.** Applying AC1(i)'s named mutation
  (`ContentHeight +28880mp`) reddened `TestStatementDocumentsRenderTheirDeclaredPageCount` on
  statement-5/20/50 and left `TestStatementPageCountFollowsTheBoundCollectionLength` **PASS**. The
  claimed discrimination holds.
- **AC9's two-way set equality genuinely fails when the declared list is shortened.** Removing
  `table_footer_test.go` from `declaredTableBehaviourSuite` while leaving the file on disk reddened
  `TestTableBehaviourSuiteIsNotSupersededByTheGolden` with the named message.
- **The new per-leg matrix guard has teeth on every target.** Changing statement-50's declared CJK
  subset from 41 to 42 reddened `requireStatementIsAWorkingStatement` on all four legs.
- **Partial survival of the sign-off is structurally impossible.** `statementSignOffStaleness` returns
  a non-empty `moved` for one changed digest of four and the wrapper fatals the WHOLE record; a record
  naming three of four, or naming an unregistered document, is a scope mismatch rejected outright.
  The four-case red-proof (control + three) is green in the ordinary suite.
- **The sign-off gate's failure message is actionable by a person.** Run behind the matrix tag it
  emits the four file paths with page/transaction counts, all **eight** confirmation items, the JSON
  **prefilled with the four current digests**, and the follow-up instruction to add the
  `{kind:"signoff"}` site. No placeholder.
- **Thai Arm A is genuinely READ, not restated.** `statementFrozenBreakItem` opens
  `fixtures/expected-breaks/expected_breaks.json` at test time and fails if the item carries no
  interior seam. No new signed corpus was minted; `break-signoff.json` is untouched.
- **Both figure corrections are right.** `expected_breaks.json` carries **26 items (18 thai / 8 cjk)**;
  the string `transaction_description` does not occur in the file. `thai-002` is 8 runes of which two
  are zero-advance marks — six advancing glyphs, as claimed (but see Finding 13).
- **DW-14 is REFUTED by measurement.** Independently re-derived from the committed PDFs: corpus max
  `/ToUnicode` section **45 -> 55**, cap 100 not crossed; per-document sections `[55, 41, 25]` for
  statement-5/20/50 and `[55, 1, 25]` for statement-1. The mechanism claim holds decisively — the
  sections are **identical** across 95, 425 and 1085 rows and 5, 20 and 50 pages, so section size
  tracks distinct glyphs in a subset, not pages or rows.
- **DW-13's measurement reproduces to the byte.** `FontFile2` `/Length1` sums: statement-1 65,740 B;
  statement-5/20/50 **77,452 B each** (62,224 + 12,428 + 2,800). zlib level 9 -> 22,487 B, saving
  **54,965 B = 71.0% of program bytes, 9.9% of the 50-page file**. The developer's added fact is
  correct: the payload is **constant** across 5/20/50, so compression helps least on the flagship.
- **AC6(iii)'s transcript is not stale.** `qpdf 12.4.0` re-run in review: `--check` reports no syntax
  or stream encoding errors on all four; `--show-npages` returns 1, 5, 20, 50. Each README quotes its
  fixture's digest exactly once, so the declared `{kind:"readme"}` site is real.
- **The four fixtures differ ONLY in collection length.** `input.folio` and `params.json` are
  md5-identical across all four; `customer`, `account` and `period` are identical; each shorter
  `transactions` array is a strict **prefix** of statement-50's.
- **Mid-band placement is real.** Measured by rendering every n from 1 to 1120: the declared lengths
  sit 10 from the lower edge and 11 from the upper edge of their bands. A boundary-adjacent length
  would have made the page counts fragile; these are not.
- **`TestOverTallSingleRowStillOverflows` is green and the file is unedited** (`git diff df8cbcc --
  folio-go/table_row_clip_test.go` is empty).
- **`mpParseToUnicode` is genuinely out of scope and contaminates nothing.** No statement assertion
  routes through it: `statementPageRuns` and `requireStatementIsAWorkingStatement` both compose
  `toUnicodeForResources` + `parseContentStreamRuns`. The defect is real and correctly recorded, and
  no golden's evidence depends on it.
- **Tree integrity.** Every mutation above was applied by hand and reverted by hand; a SHA-256
  snapshot of 142 source and fixture files taken before the review is byte-identical afterwards.

---

### Finding 1: The flagship money column cannot express a decimal-arithmetic defect — every amount is a multiple of 0.25
- **Severity**: Blocker
- **Category**: AC Conformance
- **Location**: `folio-go/statement_fixture_test.go:194-198` (`statementDiscriminatingRows`) and `:290` (`Amount: fmt.Sprintf("%d.25", 10+i)`); recorded into `fixtures/statement-{1,5,20,50}/data.json`
- **Observation**: Measured over all 1085 amounts in `statement-50/data.json`, the set of distinct
  cent-parts is exactly `{0, 25, 50, 75}` — every amount is an exact multiple of 0.25. Consequently
  **every value is exactly representable in float64**, and a naive float64 accumulation of the whole
  collection returns `602061.5`, which formats to `602,061.50` — **identical to the exact-decimal sum**
  the test computes. The four discriminating rows are 1250.00, 480.25, 75.00, 12.50, 33.75; the 1080
  filler rows are all `.25`.
- **Impact**: This is the D-4.5.5 question the story sets itself, answered wrongly on the one column
  the story says is the reason AD-23 exists: *"The statement's money column is the reason AD-23
  exists... the sum footer's independently-computed check value must be computed as an exact decimal
  too, or the check inherits the bug it exists to catch."* The **check** is exact — but the **fixture**
  is not discriminating: all four goldens would be **byte-identical** under a renderer whose money path
  went through float64, and AC2(iii)'s assertion would still pass. This repeats Story 4.5's failure
  mode one notch up (4.5's defect survived because the only column reaching the default was integral;
  here the value class is quarter-integral, which is the same thing for binary floating point). The
  four documents are about to become permanently-defended inputs and be sent to a human for
  irrevocable sign-off; after the commit this costs a re-record plus a re-requested sign-off, and
  before the commit it costs one edit to a format string.
- **Suggested Resolution**: Give the amount generator a cent-part that is not a binary fraction —
  `.10`, `.33`, `.07` and similar — across enough rows that the exact and float64 totals diverge in
  the rendered digits, and state in the four READMEs' *what this document's content can express*
  section that the money column now discriminates exact-decimal arithmetic from float. Then re-record
  and re-run the deletion screen. Do not fix this by strengthening the assertion: the assertion is
  already exact; it is the value class that cannot express the defect.
- **Related AC**: AC2(iii), AD-23, D-4.5.5

### Finding 2: AC1(ii)'s declared deletion mutation was neither performed nor declared unperformable
- **Severity**: Major
- **Category**: Tests
- **Location**: Story Delivery Log §13, the mutation table (the `AC1ii` column is `·` in all eleven rows) and the *Replacement mutations* list, which names five replacements — none of them AC1(ii)'s
- **Observation**: The story declares deletion (ii) as *"replace the table's bound rows with fixed-position
  marker elements spaced one content window apart — `fixtures/page-count-20`'s construction."* No row of
  the screen table reddens `AC1ii`, and AC1(ii) does not appear among the five replacements. §13's own
  heading claims *"every observable, by name, reddened AND stayed green"*, which is not true of AC1(ii).
  The reviewer confirmed the half that **was** demonstrated (AC1(i)'s mutation leaves AC1(ii) green) but
  **no mutation has ever been shown to redden AC1(ii)**.
- **Impact**: AC1(ii) is the observable that exists *because* "renders at N pages" is accidentally true
  of any geometrically-built document. It is precisely the assertion whose teeth cannot be assumed, and
  it is the only declared observable in the story with no demonstrated red. D-000.79 is explicit:
  *"If the implementation's shape makes one impossible to perform as written, say so and name a
  replacement — never silently substitute one the code happens to survive."* Here neither happened.
- **Suggested Resolution**: Either perform the declared deletion (build a throwaway geometric template
  that yields 1/5/20/50 pages at the four declared collection lengths, confirm AC1(i)'s assertion passes
  on it and AC1(ii)'s reddens, and delete the probe), or record in §13 that it is a **fixture
  substitution rather than a code mutation**, name the replacement, and reclassify AC1(ii) as Class A in
  the ledger. Do not leave the screen's heading claiming a red that was never observed.
- **Related AC**: AC1(ii), D-000.85, D-000.79

### Finding 3: AC8(i)'s claimed witness does not exercise the mechanism it claims to witness
- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/statement_signoff_matrix_test.go:104-191`; Delivery Log §13 *AC8 (two observables)*, item (i)
- **Observation**: §13 records AC8(i) — *"the record exists and every field is non-empty and specific"* —
  as *"witnessed by the gate's current red."* The gate's current red is the `os.IsNotExist` branch at
  `:109-159`, which `t.Fatalf`s and returns. **Everything below it has never executed**: the JSON
  unmarshal, the `reader`/`date`/`examined` non-empty loop at `:169-181`, the `len(rec.Digests) == 0`
  check, and `assertStatementSignOffMatchesFrozenHashes` at `:189`. The only `statementSignOff` value
  constructed anywhere in the tree is the synthetic one at `statement_golden_fixture_test.go:236`, and
  it drives **only** the pure `statementSignOffStaleness` — not the field checks and not the wrapper.
- **Impact**: The standing red witnesses that the **file is absent**, which is not the observable. The
  non-empty field checks — the mechanical half of AC8(i), and the thing the story says makes an
  unattributed sign-off refusable — are shipped, unexecuted and unproved. So is the wrapper that turns
  the staleness verdict into the gate's failure: if it were inverted (`if len(moved) == 0`) every test
  in the tree would still be green. The asymmetry is unexplained: the sibling observable AC8(ii)
  **did** get a synthetic red-proof in the untagged file, by exactly the technique that would close
  this one.
- **Suggested Resolution**: Extend `TestStatementSignOffDigestBindingRedProof` (or add a sibling in the
  same untagged file) with synthetic records that exercise the field checks — empty `reader`, empty
  `examined`, whitespace-only `date`, empty `digests` — and a case that drives the wrapper's two failure
  paths over a synthetic `live` map. Then restate §13's AC8(i) witness accurately.
- **Related AC**: AC8(i), D-000.85

### Finding 4: `render_test.go` cites a guard that does not exist, leaving two independent lists unreconciled
- **Severity**: Major
- **Category**: Correctness
- **Location**: `folio-go/render_test.go:1971-1974` and `:1975-1986`
- **Observation**: The comment reads: *"The rows field is deliberately NOT read from statementFixtures'
  pages field: this list and statementFixtures are two independent statements of the same four
  collection lengths, and **TestStatementSubprocessSelectorsCoverEveryFixture below compares them**."*
  `grep -rn TestStatementSubprocessSelectorsCoverEveryFixture folio-go/` returns **one hit — the comment
  itself**. The test was never written. Relatedly, the `slug` field of `statementSubprocessSelectors` is
  never read anywhere (`grep -rn 'sel.slug'` returns nothing), which is the same half-wired shape.
- **Impact**: The deliberate duplication is sound design; the mechanism that makes it safe is missing. If
  `statementSubprocessSelectors[i].rows` drifted from `statementFixtures[i].rows`, **nothing in the
  ordinary suite would notice** — the matrix legs would catch it, but they are behind `//go:build matrix`
  and are not run per-commit. A reader who trusts the comment will believe a cross-check exists.
  D-000.37's second clause is directly on point: a message must point at a mechanism that is there.
- **Suggested Resolution**: Write `TestStatementSubprocessSelectorsCoverEveryFixture` in the untagged
  package — assert the two lists are the same length, in the same order, with equal `rows`, and that
  every `slug` matches its `statementFixtures` counterpart (which also gives the dead `slug` field a
  purpose). Alternatively delete the claim from the comment and the unused field — but the guard is
  three lines and the comment is right that it is wanted.
- **Related AC**: AC7, D-000.37

### Finding 5: The recorded row arithmetic does not reproduce, and the two in-tree records contradict each other
- **Severity**: Major
- **Category**: Convention
- **Location**: Story Delivery Log §3 (*"page 1 holds **19** rows... Each declared length sits in the MIDDLE of its band (1..19 / 86..107 / 416..437 / 1076..1097)"*); `folio-go/statement_fixture_test.go:37-39` and `:252-255` (*"Measured by rendering every n from 1 to 400: page 1 holds **18** rows"*, bands `1..18 / 85..106 / 415..436 / 1075..1096`); `folio-go/statement_semantics_test.go:77` (`statementRowsOnFirstPage = 19`)
- **Observation**: Re-measured by rendering every collection length from 1 to 1120: the true page-count
  bands are **1..18 / 85..106 / 415..436 / 1075..1096**. The story's §3 band table is **off by one at
  both ends of every band**. Separately, the artifact's real page-1 row partition is **19**
  (`partition sizes [19 22 22 22 10]`, logged by the passing test), so
  `statement_fixture_test.go`'s *"page 1 holds 18 rows"* is **also wrong** as a statement about rows
  per page — its band formula `pages(n) = 1 + ceil((n-18)/22)` happens to reproduce the true edges only
  because the footer aggregate consumes a row slot on the final page, which neither file mentions. Each
  document is half right about a different half, and they contradict each other with nothing
  reconciling them.
- **Impact**: The mid-band conclusion survives (1085 measures 10 from the lower edge and 11 from the
  upper under the true bands, so the safety margin the story claims is real). But this is a story whose
  central discipline is that recorded measurements reproduce, and the number a future reader would use
  to judge whether a new collection length is safe is wrong in the story and misleading in the fixture
  file. The next story that adds a statement length will re-derive from one of these and land on a band
  edge.
- **Suggested Resolution**: Correct §3's band table to `1..18 / 85..106 / 415..436 / 1075..1096`, correct
  `statement_fixture_test.go`'s *"page 1 holds 18 rows"* to record both facts — **19 rows are placed on
  page 1, and the band edges sit one lower because the footer aggregate reserves a slot on the final
  page** — and cross-reference `statementRowsOnFirstPage = 19` so the two constants no longer read as a
  contradiction.
- **Related AC**: AC1, D-000.89

### Finding 6: AC10 declares zero observables while shipping a live, separably-witnessable one
- **Severity**: Major
- **Category**: Tests
- **Location**: Story AC10 (*"**Observables: 0.** This is a written record, not a mechanism"*); `folio-go/statement_golden_fixture_test.go:126-152` (`TestStatementFixtureReadmesRecordTheCoverageHistory`); Delivery Log §13 *AC10, AC12 — zero observables, and the zero is honest*
- **Observation**: AC10 declares 0 and §13 asserts *"No observable was invented to fill them"* — then, in
  the same paragraph, records that the written record **is** guarded by
  `TestStatementFixtureReadmesRecordTheCoverageHistory`. The reviewer red-proved that guard: re-spelling
  the required sentence in `fixtures/statement-20/README.md` (`reddened NO GOLDEN AT ALL` ->
  `reddened NOTHING WHATSOEVER`) reddens the guard **on statement-20 alone**, with the other three
  subtests green. That is a present, live, separably-witnessed machine observable by every criterion the
  story applies to the other twenty-one.
- **Impact**: D-000.88's delta is an anti-gaming metric, and it is gamed in the favourable direction by
  under-declaring: an unclaimed observable shrinks the denominator and the numerator together, leaving
  the ratio flattering. The story explicitly warns against the inverse (*"a fake zero"*) and then ships
  the mirror image. The correct declared total is **22**.
- **Suggested Resolution**: Raise AC10's declared observable count from 0 to 1, naming
  `TestStatementFixtureReadmesRecordTheCoverageHistory` and its red-proof, and restate the declared
  total as **22**. AC12's zero is genuinely honest — no machine guard was found over the DW-11/DW-13
  written answers — and should stay 0.
- **Related AC**: AC10, AC12, D-000.88

### Finding 7: AC2(iii)'s replacement mutation moves to a different sub-claim, leaving the named half unwitnessed
- **Severity**: Minor
- **Category**: Tests
- **Location**: Story Delivery Log §13, *Replacement mutations*, AC2(iii); `folio-go/statement_semantics_test.go:508-558`
- **Observation**: AC2(iii)'s declared observable is compound — *"the sum footer's single occurrence, its
  page, and its value"* — and its declared mutation (*emit the aggregate on every page*) is recorded as
  reddening *"on both the count and the page index"*. The replacement (*the sum aggregate omits the
  collection's last element*) reddens the **value** comparison only. The count-and-page half named in the
  story therefore has no mutation at all, while the value half — which the declared mutation did not
  reach — now has one.
- **Impact**: The replacement is a legitimate money-path defect and is a good test, but it is not
  *equivalent in discriminating power to the mutation it replaces*: it swaps which third of a compound
  observable is witnessed rather than preserving it. The assertion itself does cover all three claims,
  so the exposure is bounded — but the screen no longer demonstrates that the "exactly once, on the
  last page" limb has teeth.
- **Suggested Resolution**: Either add a second mutation reaching the placement limb (moving the
  aggregate item to the head of the column is cheaper than duplicating it per page and reddens the page
  index), or record AC2(iii) as compound-observable Class A in the ledger — the reviewer's adjudication
  above takes the latter reading.
- **Related AC**: AC2(iii), D-000.85, D-000.79

### Finding 8: AC3(i)'s assertion is at ROW granularity while the recorded claim is "three faces in ONE cell"
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/statement_semantics_test.go:611-621`; the claim at `folio-go/statement_fixture_test.go:158` (*"Row 1 three faces in ONE cell under ONE fallback stack"*) and in all four READMEs (*"Latin, Thai and CJK inside ONE table row under ONE fallback stack"*)
- **Observation**: The faces map is accumulated over **every run sharing row 1's baseline**, across all
  five columns — not restricted to the Description column's x-range. The logged witness shows the
  property does currently hold within one cell (`NotoSansSC:汉 NotoSansThai:ก`, both from the
  Description cell, since row 1's Note cell is empty), but the assertion would be satisfied just as well
  if the Thai were moved to the Note column and the CJK left in Description.
- **Impact**: The recorded claim — the one that makes this genuinely new coverage, since the table suites
  use a single-face `fontFamily` per subtest — is *one cell, one fallback stack*. The guard implements
  the weaker *one row*. A future fixture edit that spread the three scripts across cells would keep the
  test green while falsifying the sentence the READMEs and the story record.
- **Suggested Resolution**: Restrict the faces map to `run.XMP >= statementDescriptionColumnXMP &&
  run.XMP < statementNoteColumnXMP`, so the assertion says what the record says.
- **Related AC**: AC3(i), D-4.5.5

### Finding 9: AC5(ii)'s GPOS assertion is sign-agnostic, so a mirror-image mark offset would be recorded and defended
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/statement_semantics_test.go:1012-1028`
- **Observation**: The assertion takes the absolute value of every TJ adjustment and accepts the run if
  **any** magnitude equals the declared offset. The observed adjustments are `[-21 21]` against a
  declared GPOS x-offset of `+21`. A renderer that applied the offset with the wrong sign would produce
  `[21 -21]` and pass unchanged.
- **Impact**: On a first recording there is no prior golden to catch a sign error — the golden records
  whatever the engine did, and this assertion cannot tell the two apart. The only check on mark
  *direction* is the human sign-off item *"its marks sit on the consonants they belong to"*, so the
  machine-checkable half of AC5(ii) is weaker than the AC's wording (*"a Thai mark's position differs
  from its unpositioned default"* is satisfied either way, but the sign is what makes it correct).
- **Suggested Resolution**: Compare the signed adjustment against the signed declared offset (accounting
  for the TJ sign convention explicitly and documenting it), rather than comparing magnitudes.
- **Related AC**: AC5(ii)

### Finding 10: Delivery Log §1 states the sign-off record was built; it does not exist
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: Story Delivery Log §1, *"**One sign-off record over four digests, invalidated in whole.** Built as ruled. `fixtures/statement-signoff.json` names `reader`, `date`, `examined` and a `digests` object naming **all four** slugs."*
- **Observation**: `fixtures/statement-signoff.json` does not exist and is correctly *supposed* not to
  exist — §5, §13 and §18 all say so, and the whole gate design depends on its absence. §1's wording
  describes the record's schema in the past tense as a thing that was built.
- **Impact**: §1 is the section a reader consults for what the held decisions produced, and it reads as
  though the human sign-off has already been obtained. That is the single most consequential fact in the
  story to get backwards.
- **Suggested Resolution**: Reword to *"the record's SHAPE is built as ruled — `statementSignOff` in
  `statement_golden_fixture_test.go` declares `reader`, `date`, `examined` and a `digests` object naming
  all four slugs. The record itself does not exist and must not: its absence is what keeps the gate red."*
- **Related AC**: AC8

### Finding 11: The four READMEs' "what it cannot express" omits the money value-class limitation
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `fixtures/statement-{1,5,20,50}/README.md`, *## What it cannot express*
- **Observation**: The section lists only two limitations (no new element type or binding shape; no
  `altRowBackground` and no font-stream compression). It does not record that the money column's values
  are all multiples of 0.25 and therefore cannot discriminate exact-decimal from float arithmetic
  (Finding 1), nor that the CJK content is unasserted, nor that only one of the two Thai cells carries an
  asserted break.
- **Impact**: The *cannot express* section is the record a future reader at a gate relies on to know what
  a green golden does **not** prove. A disclosed limit that is narrower than the real one is worse than
  no disclosure, because it stops the reader looking.
- **Suggested Resolution**: Add the money value-class limitation (or remove it by fixing Finding 1 and
  recording the new capability instead), the unasserted CJK, and the unasserted `thai-002` cell.
- **Related AC**: AC10, D-4.5.5

### Finding 12: AC9's second direction is keyed on a filename convention, not on the population
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_behaviour_suite_test.go:121`
- **Observation**: The on-disk half of the set equality globs `table_*_test.go`. A table behaviour test
  file added under any other name — `statement_table_pagination_test.go`, `row_clip_test.go` — is
  invisible to the guard and can never trigger the "on disk and NOT in declaredTableBehaviourSuite" arm.
- **Impact**: Bounded: the guard's stated job is to catch deletion of the six named files, and it does
  that (red-proved above). But the second direction is advertised as set equality over the population,
  and it is set equality over a naming convention.
- **Suggested Resolution**: Either state the glob's scope as a limitation in the file header alongside the
  existing honest note about rot, or widen the observed set to any `_test.go` file importing the table
  fixtures. The former is proportionate.
- **Related AC**: AC9, D-2.5.1

### Finding 13: The `thai-002` finding's stated reason does not establish its conclusion
- **Severity**: Minor
- **Category**: Correctness
- **Location**: Story Delivery Log §1, second correction; `folio-go/statement_fixture_test.go:176-184`
- **Observation**: The recorded reason is that `เก็บเงิน` *"presents **six advancing glyphs** — the same
  six as `thai-001`'s first word `ประเทศ`"*, from which *"No column width breaks `thai-002` while leaving
  `thai-001`'s first word intact"* is concluded and labelled *"a measured finding, not a choice."* The
  glyph **count** is correct and was re-verified (8 runes minus two zero-advance marks). But the six
  glyphs are not the same six — `เ ก บ เ ง น` against `ป ร ะ เ ท ศ` — and line breaking is decided by
  **advance width**, not glyph count. Equal counts of different glyphs do not imply equal widths.
- **Impact**: No AC exposure: AC5 requires *at least one* interior seam, `thai-001` supplies it, and the
  reviewer confirmed that comparison is non-vacuous (the cell wraps to two lines and matches the frozen
  label at rune 6). The defect is in the record, which presents an inference as a measurement in a story
  whose discipline is that measurements are measured.
- **Suggested Resolution**: Either measure the two strings' shaped advance widths and record the numbers,
  or reword to *"not asserted; no usable width was found, and the reason was not measured to exhaustion"*.
- **Related AC**: AC5(i), D-000.17

### Finding 14: Dead branch in AC2(ii)'s row assertion
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/statement_semantics_test.go:475-478`
- **Observation**: `if strings.Contains(wantDesc, " regional office") { wantDesc = "Quarterly maintenance
  retainer for the" }` special-cases the wrapping row. `assertRow` is only ever called with `rows[0]`
  (always `TXN-0001`) and `rows[len(rows)-1]` (`TXN-0005`/`0095`/`0425`/`1085`, per the test's own log).
  The wrapping row is `TXN-0002` and can never be either.
- **Impact**: Dead code inside an assertion, which reads as though the first-or-last row might wrap and
  invites a future reader to preserve a case that does not exist.
- **Suggested Resolution**: Delete the branch, or add a comment recording that it guards against a future
  reordering of `statementDiscriminatingRows` and assert that invariant explicitly.
- **Related AC**: AC2(ii)

### Finding 15: `statement-20`'s stated discriminating rationale is subsumed by `statement-50`
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/statement_fixture_test.go:54-55`; `folio-go/byte_neutrality_test.go` (the statement-20 `goldenDigestRecord` comment); `fixtures/statement-20/README.md`
- **Observation**: statement-20 is justified as *"crosses the page-9-to-page-10 digit-count boundary
  (D-2.7.2) WITH a table present."* statement-50 crosses the same boundary, with the same table, and was
  recorded in the same commit.
- **Impact**: Not a correctness problem — statement-20 is a genuine additional length and the family's
  cost is already paid. But the recorded rationale is the thing a future reader would use to decide
  whether the fixture may be retired, and as written it is a property another fixture already has.
- **Suggested Resolution**: Restate statement-20's rationale as the property only it has — the mid-sized
  document that exercises the boundary at a scale a human can actually read end to end during a
  re-attestation, which is the role the sign-off's `examined` instructions already give it.
- **Related AC**: AC7, D-4.5.5

### Finding 16: The Amount cell's x-range lower bound overlaps the Note column
- **Severity**: Nit
- **Category**: Correctness
- **Location**: `folio-go/statement_semantics_test.go:483` — `cell(statementNoteColumnXMP+31000, statementAmountColumnRightMP+1)`
- **Observation**: The Note column spans 335,000..369,000 mp (34pt wide); the Amount cell's lower bound
  is 335,000 + 31,000 = **366,000 mp**, which sits 3,000 mp inside the Note column's tail. It is harmless
  today because Note cells are left-aligned and their runs begin at 335,000, so nothing is ever drawn in
  the overlap.
- **Impact**: Latent. Changing the Note column's alignment to `right` or `center` would silently fold Note
  text into the Amount comparison, producing a confusing failure rather than a clear one.
- **Suggested Resolution**: Use `statementNoteColumnXMP + 34000` (the column's declared width) as the
  boundary, or derive it from the same declared widths the other column-origin constants are derived from.
- **Related AC**: AC2(ii)

---

## Finding Resolutions

### Finding 1 — FIX

The statement amount generator now uses cent parts that are not binary fractions, and
`TestStatementMoneyColumnDiscriminatesFloatArithmetic` pins the exact-decimal totals, the naive
binary64 totals, and the number of individually divergent Amount cells. All four data files, PDFs,
metadata digests, README digest literals and `goldenDigestRecord` literals were re-recorded together.
Evidence: `folio-go/statement_fixture_test.go`, `folio-go/statement_semantics_test.go`,
`fixtures/statement-{1,5,20,50}/`, and D-4.7.8.

### Finding 2 — FIX

AC1(ii)'s declared geometric-fixture substitution was performed. It kept the absolute page-count
observable green and reddened the data-dependence observable, exactly separating the two claims. The
mutation transcript and the corrected Class A accounting are recorded in Delivery Log §§13–14.

### Finding 3 — FIX

The sign-off record's required-field checks and the wrapper's scope/staleness verdict were extracted
into pure helpers and red-proved with synthetic records covering empty `reader`, `date`, `examined`,
empty digests, scope mismatch, moved digests and the populated control. Evidence:
`statementSignOffFieldProblems`, `statementSignOffGateVerdict`, and
`TestStatementSignOffFieldChecksAndGateWrapperRedProof` in
`folio-go/statement_golden_fixture_test.go`.

### Finding 4 — FIX

`TestStatementSubprocessSelectorsCoverEveryFixture` now reconciles the untagged fixture declarations
with the subprocess selector list by length, order, slug and row count. The formerly unused `slug`
field is part of that check. Evidence: `folio-go/render_test.go`.

### Finding 5 — FIX

The record now distinguishes the two measured facts: 19 rows are placed on page 1, while the page-count
bands begin one row lower because the final aggregate reserves a slot. The true bands are recorded as
`1..18 / 85..106 / 415..436 / 1075..1096`, and the shared 19-row constant is cross-referenced.
Evidence: Delivery Log §3, `folio-go/statement_fixture_test.go`, and
`folio-go/statement_semantics_test.go`.

### Finding 6 — FIX

AC10's README-history guard is now declared as one separably witnessed observable. The declared total
was corrected from 21 to 22, while AC12 remains an honest zero; the post-review money discriminator is
named separately as the next audit's twenty-third observable. Evidence: AC10 and Delivery Log §§13–15.

### Finding 7 — FIX

The replacement money mutation is retained as evidence for AC2(iii)'s value limb, and an attempted
aggregate-placement mutation was measured rather than assumed. It did not isolate the occurrence/page
limb, so that compound limb is explicitly classified Class A instead of being claimed as witnessed.
Evidence: Delivery Log §§13–14.

### Finding 8 — FIX

The three-face assertion now restricts collected runs to the Description cell's x-range, so its
mechanical claim is one cell under one fallback stack rather than merely one row. Evidence:
`folio-go/statement_semantics_test.go`, with mutation M4 in Delivery Log §13a.

### Finding 9 — FIX

The GPOS assertion now checks the signed adjustment using the PDF `TJ` sign convention. A mirrored
renderer reddens the corrected assertion while passing the former magnitude-only assertion, proving
the fix detects the exact defect reported. Evidence: `folio-go/statement_semantics_test.go`, mutation
M5 in Delivery Log §13a.

### Finding 10 — FIX

Delivery Log §1 was corrected to say that the sign-off record's schema and gate shape were built while
the record itself was deliberately absent pending inspection. The record now exists only because Panit
Wechasil subsequently completed the visual examination captured in
`fixtures/statement-signoff.json`.

### Finding 11 — FIX

All four fixture READMEs now state the money-path boundary, the unasserted CJK reading, and the
unasserted `thai-002` seam. Because Finding 1 removed the quarter-integral value class, the money note
accurately says the fixtures discriminate truncating scale-to-minor-units float arithmetic but cannot
discriminate every possible rounded float implementation. Evidence:
`fixtures/statement-{1,5,20,50}/README.md`.

### Finding 12 — FIX

The AC9 guard's file header now says explicitly that its reverse scan covers the
`table_*_test.go` naming convention rather than every conceivable filename. The six-name deletion
guard remains two-way and red-proved within that declared population. Evidence:
`folio-go/table_behaviour_suite_test.go`.

### Finding 13 — FIX

The equal-glyph-count inference was withdrawn. The usable-width window was measured: `thai-002` fits
at 23,400 mp and breaks at 23,200 mp, while `thai-001`'s first labelled word needs 23,704 mp; therefore
every width that breaks the former clips the latter. Evidence: `folio-go/statement_fixture_test.go`,
Delivery Log §1, and D-4.7.9.

### Finding 14 — FIX

The unreachable wrapping-row special case was removed from the first/last-row assertion and replaced
with an explicit invariant that neither sampled boundary row is the wrapping row. Evidence:
`folio-go/statement_semantics_test.go`, mutation M8 in Delivery Log §13a.

### Finding 15 — FIX

`statement-20` is now justified as the mid-sized table document that crosses the page-number digit
boundary while remaining practical for a person to read end to end during re-attestation; it no longer
claims the boundary alone uniquely distinguishes it from `statement-50`. Evidence:
`folio-go/statement_fixture_test.go`, `folio-go/byte_neutrality_test.go`, and
`fixtures/statement-20/README.md`.

### Finding 16 — FIX

The Amount cell's lower bound is derived from the Note column origin plus its full declared width, so
the inspected x-range begins at the true Amount boundary and cannot absorb right- or center-aligned Note
text. Evidence: `folio-go/statement_semantics_test.go`.
