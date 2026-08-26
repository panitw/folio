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

### DW-4 — Nobody owns cutting `folio-go/v0.1.0` — **owner decision when Epic 4 is planned**
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

### DW-17 — Surfacing a returned `Diagnostic` to a human is a presented-interface obligation, not a Go call-graph one
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
