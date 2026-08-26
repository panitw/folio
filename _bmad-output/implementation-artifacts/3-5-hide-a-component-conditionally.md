---
baseline_commit: ae9e551
---

# Story 3.5: Hide a component conditionally

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-5-hide-a-component-conditionally`
**Status:** `done`
**Covers:** **FR20** · **AD-24**

**Primary invariant:** **AD-24**, *Visibility* clause, verbatim
(`ARCHITECTURE-SPINE.md:446-448`):

> **Visibility.** A condition (FR20) is evaluated during bind. A hidden element is **absent**
> from the `PageModel` and leaves no gap; siblings never move, because nothing in a band ever
> reflows. Visibility applies to elements only — a condition may **not** hide a table row,
> which would make pagination a function of data in a way FR25 does not define.

**Co-primary invariant:** **AD-4** (two passes; **no `page` namespace in the expression
language, ever** — `ARCHITECTURE-SPINE.md:151-164`). AD-24's *"evaluated during bind"* is what
keeps AD-4 intact: the moment visibility could depend on the page an element lands on,
evaluation would have to move after layout and the two-pass structure would collapse.

**Adjacent invariants:** **AD-14** (one `Diagnostic` type, error vs warning is a hard line,
never a panic) · **AD-1** (the determinism boundary is a directory boundary) · **AD-11** (row
alias, aggregates over the whole collection) · **AD-13** (a table declares `x` and `y` only) ·
**AD-5** (the page model knows nothing about PDF) · **AD-10** (opaque element ids).

**Governing rulings:** **D-3.2.3** (`if(null)` is silently FALSE; an **absent** condition path
is a located Error — this story reuses that pair verbatim and adds nothing to the axis) ·
**D-000.68** (a guard must be anchored to something the code under test cannot move; pin to a
literal when the set is permanent) · **D-000.67** (a targeted fix does not sweep, even within
one file; a presence precondition is itself population-keyed) · **D-000.65** (a diagnostic code
is minted by the story in which its condition can first occur in a real document, where the
`Diagnostic` is constructed) · **D-000.50** (*can any existing subject express this defect?*
comes **before** the assertion) · **D-000.24** (a guard with no constructible red-proof is
labelled as such and never credited with one) · **D-3.3.1** (the collection seam is exactly two
methods on `expr.Resolver`; **do not widen the interface**) · **D-3.4.4** (the
registered-but-unimplemented machinery was REMOVED at 3.4 — do not resurrect it) ·
**D-3.4.7** (dot-imports are banned outright under `internal/`) · **D-000.64** (the gate is
three modules) · **D-000.59** · **D-000.30** · **D-000.36** · **D-000.38** · **D-000.45** ·
**D-000.14** · **D-000.15** · **D-000.4** (heavy-test cadence is per-epic) · **D-000.17** /
**D-2.1.14** / **D-000.57** (the required red) · **D-000.9** · **D-000.26** (a figure without
its scope and flags is not a figure).

**Nothing in the decision log places an obligation on Story 3.5.** Grepped at creation across
all 282 entries: no prior ruling names this story. `deferred-work.md` likewise carries **no**
entry owned by 3.5, and **no** entry mentioning `visibleIf` or visibility at all (measured:
`grep -n "visibleIf\|visibility" deferred-work.md` → 0 matches). This story therefore inherits
no debt. It also **retires no DW entry**.

---

## FOUR divergences between the brief and the shipped tree, all measured in this run

**These change the shape of the story materially. Read them before anything else.**

**D1 — `visibleIf` already ships. This story does not add a field; it wires one that is already
there and inert.** The field exists end-to-end: modelled at
`folio-go/internal/template/model.go:198` (`VisibleIf Presence[string]`, a **common** field on
all five element kinds), parsed at `folio-go/internal/template/parse_bands.go:185-196`,
serialized at `folio-go/internal/template/serialize.go:237-243`, and **statically validated at
load** by `checkVisibleIfExpression` in `folio-go/folio_expr_validate.go:103-112`. It is
specified in the format reference at
`_bmad-output/specs/spec-folio/folio-format.md:184`. **Nothing reads it at render**: `grep -rn
"VisibleIf" folio-go | grep -v _test` returns only model, parse, serialize and the load check.
Story 2.5's file states the deferral in as many words — *"`visibleIf` parses today and is
ignored at render; it goes on being ignored"*
(`2-5-compose-a-page-from-three-bands.md:207-208`). And 3.2's own code comment names this story
as the collector: `folio_expr_validate.go:38` — *"Story 3.5 drives visibility from this exact
field."*

**D2 — tables DO exist in the schema. The brief's "tables do not exist until Epic 4" is true of
*rendering* and false of the *schema*.** `template.ElementTable` is one of the five closed
element kinds (`model.go:179`), with `TableExt` (`model.go:214`: `Bind`, `As`, `Columns`,
`HeaderHeight`, `AltRowBackground`) and `Column` (`model.go:223`: `ID`, `Label`, `Width`,
`Align`, `Bind`, `Footer`, `FooterOf`, `FooterFormat`, `Extra`). Tables load, validate,
round-trip and are bind-checked (`render.go:275-308`) today; only *layout* waits for Epic 4.
The brief's `grep -rl '"table"' fixtures/` → no matches **is** correct and reproduced at
creation, but it measures the wrong population: the table subjects live in
`internal/template/fixtures_test.go`, not in `fixtures/`.

**D3 — the decisive one. A column-level `visibleIf` LOADS CLEAN TODAY, and this was verified by
execution, not by inspection.** `decodeColumn` (`parse_bands.go:346`) seeds
`consumed := map[string]bool{"id": true, "label": true, "width": true, "bind": true}` plus the
four optional keys; anything else falls through `extraFields`
(`internal/template/decodehelpers.go:116-135`) into `Column.Extra` and is **preserved
opaquely** — accepted, round-tripped, and silently ignored. A throwaway probe at
`folio-go/probe_rowvis_test.go` (written, run, and **removed**; tree re-verified clean) put
`"visibleIf": "transaction.isVisible"` on column `e3` of a table element and `ParseTemplate`
returned **no error**.

> This falsifies sharp edge 3's expectation for AC3. **D-000.50's question — *can any existing
> subject express this defect?* — is answered YES, by measurement.** Row-level visibility is not
> unrepresentable; it is representable, silently accepted, and *documented as invalid* at
> `folio-format.md:184` (*"Not valid on a table column"*). AC3 is a real, red-provable
> load-time rejection closing a live spec/code divergence — not a forward guard, and it needs
> no D-000.24 label.

**D4 — the load check does NOT reject a literal condition, although `if()`'s does.**
`if()`'s condition slot is `argNotLiteral` (`internal/expr/table.go:99`), rejected statically at
`internal/expr/check.go:118-126`, because the grammar has no boolean literal so a bare literal
can never be a boolean. `checkVisibleIfExpression` runs only `expr.Parse` + `expr.Check` on the
whole expression, and a top-level `PathExpr`/`StringLit`/`NumberLit` is not a call, so
`argNotLiteral` never applies. Measured consequence: **`"visibleIf": "42"` and `"visibleIf":
"\"hello\""` load clean today** and would fail only once this story starts evaluating them —
precisely the outcome `folio_expr_validate.go:35-42` says must not happen. AC6 closes it.

---

## Baseline, measured in this run at creation

HEAD **`ae9e551`** — *"Story 3.4: Format dates and numbers by declared locale (finisher)"* —
branch `main`, working tree **clean** (verified before and after the AC3 probe).

**The gate is THREE modules (D-000.64). Every figure below was re-measured in this run, not
carried forward** (D-000.26: a figure without its scope and flags is not a figure).

| scope | invocation (verbatim) | measured at `ae9e551` |
|---|---|---|
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l . && go test ./... -count=1` | **822 pass · 1 fail · 1 skip** |
| `lint/` | `go test ./... -count=1` | **98 pass · 3 fail** |
| `hashmatrix/` | `go test ./... -count=1` | **3 pass** |

`go build`, `go vet` and `gofmt -l .` are all clean.

**The one `folio-go` FAIL is `TestCorpusMeetsP6ExerciseFloors` — a REQUIRED red.** Measured
stats, byte-identical to the orchestrator's declaration:
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` (P6g floor 20 not met). **D-000.17** — *a
floor that is not met is reported as unmet, never filled* — plus **D-2.1.14** and **D-000.57**.
It measures the Thai line-breaking corpus and has nothing to do with this story. **Never "fix"
it.** Its stats must be byte-identical at hand-off or D-000.57's third clause is breached.

**The three `lint` FAILs are one root cause — DW-19, and they are LOCAL-ONLY.** They are
`manifest.TestManifestUpToDate`, `manifest.TestResolveAssetsIncludesWordlist` and
`rules.TestFontsAssetsNoticeRemovalRedProof`, all from `.font-sources:` *"contains a committed
font binary but no LICENSE*"*. `.font-sources/` is gitignored at `.gitignore:85` and holds three
variable-font `.ttf` with no `LICENSE*`. **CI never sees the directory and is green.** **Fixing
them by weakening the rule is forbidden.** Any *fourth* `lint` failure is a regression owned by
this story.

**Cadence (D-000.4):** build / vet / gofmt / unit tests, **all three modules**, every story.
**The cross-target hash matrix is NOT run in this story.** Judged: **not hash-shaped.** This
story adds no font byte, no image byte, no glyph, no measurement, no new PDF operator class and
no new number-emitting path. It *removes* runs and placements that the already-pinned path would
otherwise have produced; every byte it does emit travels the identical shaping and serialization
code. Its determinism risk is a boolean decision made from decoded JSON, fully covered by AD-1's
existing import/selector lints. **Epic 3's boundary matrix run is due after 3.7 and will measure
everything this story changes in the render path.**

**Note for the developer:** `folio-go/.matrix-build/` exists in the working tree and is
gitignored. It holds stale cross-target test binaries that still contain the string `VisibleIf`.
Do not read them as evidence of a shipped consumer, and do not let them appear in a `File List`.

---

## In plain terms (read this first if you just want the gist)

A statement for a customer with no overdue balance should not print an empty "Overdue" box. The
file format has long carried a way to say "only show this if the data says it matters," but
nothing ever looked at it when the page was produced. This story is where the engine finally
reads it: a yes-or-no answer hides or shows a component, an unclear answer stops the render and
says why, and nothing is guessed at.

Here is the one thing that will look like a bug and is not. When a component is hidden, the
things around it do not slide up to fill the space — a gap is left, on purpose. If a component's
position depended on its neighbour being hidden, the same template would produce a different
document for every customer, with page breaks landing differently each time. A template's
geometry is decided by the template, not the data, and a gap is the price of that.

Two fences are held here rather than crossed, and both are now enforced, not just described.
Conditions turn components on and off; they never change how one looks, and a template that
tries to make a colour or font depend on the data is now refused at load, naming the offending
piece — including a table's own alternating-row colour, nearly missed and now covered too. And a
condition may sit on a component but never on a single table row, because rows decide where pages
break; that rule existed on paper from the start but was never checked, and is refused at load
now as well.

One further thing needed catching, and very nearly did not: a hidden picture was, for a while,
still being smuggled into the finished document even though the page correctly reported it as
absent. That is closed now — a hidden picture is genuinely gone, not merely invisible.

Two things will look wrong afterwards and are not. One coverage-marker test stays red on
purpose — it measures Thai line-breaking, unrelated to this work. And three licence-checker
warnings appear only on this developer's machine, from a folder the build server never sees.

## Story

**As a** template author,
**I want** to hide a component when the data says it is irrelevant,
**So that** an empty section does not print as an empty box.

---

## Do not re-open — settled rulings this story inherits

| # | Settled | Ruling | What that forbids here |
|---|---|---|---|
| S1 | Visibility is evaluated **during bind** — pass one, before pagination | **AD-24** verbatim | Do not evaluate at layout or serialization time. **Sharp edge 5 is CLOSED by the spine; nothing is open.** |
| S2 | Hidden means **absent**, leaving a **gap**; siblings never move | **AD-24** | Do not "close up" a band. Do not add a reflow mode, a flag, or a TODO proposing one. |
| S3 | Visibility applies to **elements only**, never a table row | **AD-24**, `folio-format.md:184` | Do not add `visibleIf` to `Column`. |
| S4 | A condition resolving to explicit JSON **null** is **silently false** | **D-3.2.3** (owner) | No diagnostic of any severity for the null case. |
| S5 | A condition path **absent** from the data is a **located Error** | **D-3.2.3**, `internal/bind/text.go:401` | Do not soften an absent path to "hidden". |
| S6 | **No truthiness.** A string or number condition is a hard error | `internal/expr/eval.go:164-168` | Do not treat `""`/`0`/`"false"` as false. |
| S7 | `expr.Resolver` is **exactly** `Resolve` + `CollectionLength` + `ProjectCollection` | **D-3.3.1**, enforced by `lint/internal/rules/resolvermethodset.go` (go/types, exact set + signature equality) | **Do not widen the interface.** Reuse `Resolve`. |
| S8 | There is **no `page` namespace** and none may be added | **AD-4** | A visibility condition may never consult page state. |
| S9 | The registered-but-unimplemented machinery was removed at 3.4 | **D-3.4.4** | Do not resurrect a "declared but not yet wired" registry for visibility. |
| S10 | Dot-imports are banned under `internal/` | **D-3.4.7** | — |

---

## R — design constraints derived from the record during creation

**R1 — Compute visibility for ALL THREE bands in ONE place, before any collection pass.** The
pipeline (`render.go:1046`, `renderDocument`) runs: `documentBands` → `checkTableBindings` →
`pageGeometryOf` → `collectImageRuns(t)` → **phase A** content-band text (`:1073`) →
`layout.Paginate` (`:1085`) → **phase B** header/footer text with `headerFooterResolver(pageCount)`
(`:1091`). Three consequences force a single early computation:

- `collectImageRuns(t)` at `:1068` runs **before** the band text loop and takes the template, not
  the resolved bands — so a per-band inline check would miss hidden images entirely.
- Content-band visibility must be decided **before** `layout.Paginate` at `:1085`, because
  hiding a content element changes which items exist and therefore the page count.
- **Phase B is the AD-4 hazard.** Header and footer text is bound *after* pagination, with a
  resolver that knows `pageCount`. Deciding a header element's visibility inside phase B would
  put the visibility decision downstream of the page count. Computing every element's verdict
  once, before phase A, from the data scope alone, makes visibility **structurally**
  page-independent rather than page-independent by comment — the D-000.68 distinction between an
  anchor and an assertion.

**R2 — Visibility suppresses OUTPUT, never VALIDATION.** This is the story's most dangerous
trap and it has a shipped precedent that reads as if written for it. `render.go:601-616` (QA
Finding 5 of Story 2.5, Major) records that the `fontFamily` chain must be validated **before**
the empty-text short-circuit, because the earlier ordering meant *"the SAME broken template
passing or failing depending on which report it was handed."* A visibility skip placed too early
reproduces that defect exactly: a hidden element with an unresolvable font chain, a missing image
asset, or a table bind that does not resolve to an array would stop erroring. **A template is
valid or invalid on its own terms; data decides only what is drawn.** Hidden elements still get
`checkTableBindings` (`render.go:275-308`), still get `fontChain` (`render.go:611`), still get
asset resolution.

**R3 — Drop hidden elements BEFORE `ColumnItem` construction.** `layout.MixedItemError`
(`internal/layout/paginate.go:178-198`) already refuses an empty column item, with the message
*"a column item carries neither a text run nor an image, so it would occupy column space while
drawing nothing — an invisible item can push visible content onto another page."* An
implementation that emits an empty item for a hidden element trips this and is wrong for exactly
the reason the message states.

**R4 — There is no exported bind entry point for a bare expression, and one is needed.** Every
`internal/bind` entry (`BindText:154`, `BindTextSpans:187`, `Resolve:206`) is text-span shaped
and coerces its result to a string (`text.go:265-278`: anything not string/null is a hard error).
`visibleIf` holds a **bare** expression with no `{{ }}` — that is why
`checkVisibleIfExpression` exists separately, and `folio_expr_validate.go:30-34` explains that
routing it through the text path *"would scan for `{{ }}` occurrences inside it and find none,
passing silently no matter what the string said — a vacuous check, not a real one."* The
adapter that would make `expr.Eval` reachable, `exprResolver` (`text.go:304-307`), is
**unexported**. So this story adds one exported function, shaped like
`func EvaluateCondition(src string, scope Scope, fc expr.FormatContext, elementID string)
(expr.Value, []expr.Caveat, error)`, wrapping `expr.Parse` / `expr.Check` / `expr.Eval` over
`exprResolver`. **This is a new function on `bind`, not a new method on `expr.Resolver` (S7).**

**R5 — This story mints NO diagnostic code, and that is a measured finding, not an omission.**
D-000.65 mints a code where the `Diagnostic` is constructed. Measured: the registry is exactly
two constants — `DiagCodeTextClippedWidth` (`diagnostic.go:87`) and `DiagCodeEmptyAverage`
(`diagnostic.go:109`) — and although `SeverityError` is declared (`diagnostic.go:36`), **nothing
in the module constructs an error-severity `Diagnostic`**; errors travel as ordinary Go errors
and `Render` returns `Result{}, err` (`render_entry.go:173-175`). Every condition this story
introduces is either a **load error** (AC3, AC6 — plain `error` from `internal/template` /
`folio`, exactly like the three existing column checks at `parse_bands.go:403-431`) or a **render
error** inherited unchanged from `if()`'s already-shipped paths (S5, S6). None constructs a
`Diagnostic`. **`internal/diag` remains Story 3.6's.** If the developer finds itself minting a
code, that is a signal the design has drifted — stop and raise it.

**R6 — Two different nulls, and conflating them is the subtlest available bug.** They are
distinct states with opposite meanings:

| state | JSON | meaning | evidence |
|---|---|---|---|
| no condition declared | field **absent** | **visible** | `Presence.Set == false` |
| no condition declared | `"visibleIf": null` | **visible** | `parse_bands.go:187-188` stores `presentNull`; `folio_expr_validate.go:43` deliberately skips validation for it (`el.VisibleIf.Set && !el.VisibleIf.Null`) — there is no expression to check because none was declared. Matches `"style": null` = inherit the documented defaults. |
| condition declared, **resolves** to null | `"visibleIf": "customer.flag"` with `"flag": null` in the data | **hidden**, silently | **D-3.2.3** (owner); `internal/expr/eval.go:162-163` |

A subject for the middle row already exists at `internal/template/fixtures_test.go:413`.

---

## Acceptance Criteria

Every AC names its **anchor** per **D-000.68** — the compiler, the type system, or a literal the
test owns — and every guard names the **mutation that reddens it**. An AC without a red-proof
says so in its own text (**D-000.24**); none here needs that label.

### AC1 — A false condition makes the element absent from the page model

**Given** an element carrying `visibleIf` whose condition evaluates to JSON `false`,
**When** the document is rendered,
**Then** the element contributes **zero** entries to `pagemodel.Page.Runs` and **zero** entries
to `pagemodel.Page.Images` (`internal/pagemodel/pagemodel.go:243-248` — there is no per-element
node in the page model; an element is evidenced only by the runs and images it produced).

**Instrument.** Render one template twice against two data sets, one making the condition
`true` and one `false`, and compare run/image counts and contents.

**Anchor: a literal the test owns** — the two data documents and the expected run count are
written by the test, not read from the template.

**Red-proof, available and REQUIRED.** At baseline the field is inert (D1), so **this AC fails
today**: both data sets produce identical, complete output. The developer runs this test
**before** writing the implementation and records the observed baseline failure in the Delivery
Log. Post-fix, the discriminating mutation is deleting the visibility skip: the two counts
become equal again.

### AC2 — Byte-identity with the element deleted, and siblings do not move

**Given** the same template and a data set hiding element `X`,
**When** the document is rendered,
**Then** the output bytes are **identical** to rendering a second template that is the first with
element `X` **deleted from the JSON**, against the same data —
**And** every other element's `TextRun` / `ImagePlacement` coordinates are identical between the
`X`-hidden and `X`-visible renders.

This is the AC that states AD-24's gap as a **deliberate design**. The first clause is what
"absent from the page model entirely" means, made observable at the strongest available
granularity. The second clause is the *"siblings do not move"* clause, and it is the one a
reader will misread: it asserts that the surviving elements keep the **same** declared
positions, i.e. that a **gap is left**. It is not asserting that the band closes up — asserting
that would contradict AD-24.

Positions are structurally incapable of reflowing, and the developer should verify this rather
than assume it: `layout.PlaceInBand` (`internal/layout/band.go:110-119`) is
`return bandOrigin + elementY`, and `layout.Origins` (`band.go:97-105`) derives band origins from
page setup alone through a deliberately closed `PageGeometry` struct — `band.go:41-45` states
there is *"no field here through which a measured element could reach the derivation, which is
how AD-24's 'nothing negotiates' is enforced structurally rather than by a comment."* **So the
second clause should pass at baseline and after; it is a regression guard, and the AC says so
rather than claiming a red-proof it does not have.** The first clause carries this AC's teeth
and does redden — see AC1's baseline failure.

**Anchor: a literal the test owns** — the second, element-deleted template is authored by the
test.

**One nuance the developer must NOT confuse with a reflow.** Pagination *is* a function of which
items exist. `layout.Paginate` (`internal/layout/paginate.go:257`) slides its window to the first
item that did not fit, and `paginate.go:56-58` states *"the window slides and the COLUMN IS NEVER
MUTATED. Every declared column Y is untouched."* Hiding a content element can therefore change
which **page** a later element lands on, and can change the total page count. That is fully
consistent with AD-24 — declared positions are untouched — and it is exactly what makes AC2's
byte-identity clause meaningful, because the deleted-element template must paginate identically.

**FINISHER CORRECTION (Blocker 1).** The shipped test exercised only the text subject. The
reviewer measured that a hidden **image** element's byte-identity clause FAILED:
`assetKeys`/`pdfImages` were built from the unfiltered `imageRuns` slice, so the `/XObject` was
embedded in the PDF regardless of visibility — hidden-image render 8686 bytes with
`/Subtype /Image` present, element-deleted render 8316 bytes with none. Fixed by keeping every
image run's validation unconditional (asset presence, decode — AC7(b) is unaffected, re-verified)
and building `pdfImages` from the asset keys reachable from **visible** image runs only
(`render.go`'s `buildPageModel`). `TestVisibleIfHiddenImageIsByteIdenticalToElementDeleted`
(`visibility_test.go`) is the image half of this AC, red-proved: reverting the fix (making every
asset key "visible") reproduces the exact byte-count divergence above; with the fix, hidden and
element-deleted renders are byte-identical. See Finding Resolutions, Finding 1.

### AC3 — A visibility condition on a table column is REJECTED at load

**Given** a template whose table `columns[]` entry carries a `visibleIf` key,
**When** the template is loaded,
**Then** `ParseTemplate` returns a **load error** naming the **column id** and stating that
visibility applies to elements only.

**Placement: `decodeColumn` in `folio-go/internal/template/parse_bands.go`**, as a **pure field
presence** check, alongside the three existing column checks at `parse_bands.go:403-431`
(AC43 checks 1–3: `footerOf` with `footer:"count"`, `footerOf`/`footerFormat` with no `footer`,
`footerOf` prefix). It belongs there and not in the root `folio` package for two reasons: it
needs no expression parsing, and placing it at the template layer makes the rejection independent
of whether expression validation runs at all.

**The reason goes in the message, because the reason is the whole point:** row-level visibility
would make pagination a function of data, which FR25 does not define (AD-24). A message that
merely says "unknown field" would be a weaker rejection than the one AD-24 asks for.

**Anchor: a literal the test owns** — the test writes the offending column JSON.

**Red-proof, MEASURED AT CREATION, not predicted.** A probe was written, run and removed at
`ae9e551`: a table column carrying `"visibleIf": "transaction.isVisible"` **loads clean today**,
absorbed into `Column.Extra` by `extraFields` (`decodehelpers.go:116-135`) and round-tripped.
**D-000.50's question is answered YES by execution.** The discriminating mutation post-fix is
deleting the check: the template loads clean again.

**This closes a live spec/code divergence.** `folio-format.md:184` has said *"Not valid on a
table column"* since the format was written; nothing has ever enforced it.

**Scope note the developer should not talk itself out of.** Tables do not *render* until Epic 4
(D2), but they *load* today, the subject is constructible today, and the AC is a **load-time**
assertion. There is nothing to defer.

### AC4 — Conditional formatting is unsupported, and the assertion is behavioural, not absential

**Given** a style field carrying what looks like an interpolation — e.g.
`"style": {"background": "{{if(customer.overdue, \"#FF0000\", \"#00FF00\")}}"` —
**When** the document is rendered against two data sets that would select **different** branches,
**Then** the output bytes are **identical**: no style field is ever interpolated, and styling is
never data-driven.

**This AC deliberately does not assert a rejection.** The source AC says conditional formatting
is *"unsupported and out of scope"*, not that it is refused. An AC asserting the mere **absence**
of a feature is the exact shape that passes vacuously (sharp edge 3), so this is phrased as a
**behavioural** claim with a mutation that reddens it: implement style interpolation and the two
renders diverge. That is a real anchor.

**Measured at creation, and it is why this phrasing was chosen over a rejection.**
`style.background` accepts **any** string with no validation whatsoever
(`parse_bands.go:479-490` — `decodeStringRaw` and nothing else; there is no hex-colour check
anywhere in the module), and `checkTextExpressions` runs only over a text element's `value` and
a table column's `bind` (`folio_expr_validate.go:22-64`) — never over a style field. So such a
template loads clean today and is inert. **A subject can express this defect, so D-000.24's
"no red-proof available" label does not apply and must not be used.**

**Anchor: a literal the test owns** — the two data sets and the expectation of byte-identity.

**See DECISION-1.** Whether to *additionally* reject `{{` inside a style string at load is a
scope question the lead should settle; the recommendation is not to, and this AC stands either
way.

**FINISHER CORRECTION (Finding 3 / Major, ruled D-3.5.4).** The shipped subject used `background`
as its render-consumer implication, but `style.Background` has **zero** render consumers outside
`internal/template`'s own parse/serialize and the load-time rejection check — disabling the check
entirely left the shipped subject green, so it could not carry the red-proof its own text claimed.
Measured population, non-test consumers outside `internal/template`, for all ten style fields:
`background`/`align`/`valign`/`border` — only the rejection check itself, nothing renders them;
`bold`/`italic`/`padding`/`altRowBackground` — zero references anywhere; `fontSize` —
type-impossible for this defect (`Presence[geom.Length]`, no string can inhabit it);
**`fontFamily`** — renders, and is a string: the **only** field both true of.

**Ruled: AC4's subject is `fontFamily`.** `TestStyleFieldIsNeverDataDriven` (`visibility_test.go`)
now renders one template whose `style.fontFamily` is a literal (`"body"`) against two data
documents differing only in `customer.overdue` — a field the template never binds — and asserts
byte-identity. **Red-proof, run and reverted:** temporarily making `fontChain` resolution consult
`data` (render.go) reddens this test (the `overdue: true` branch resolves a chain naming a face
absent from the FontSet and errors). **The bound is stated as a measured population, not an
apology** (D-000.24's labelled category used correctly): asserted for `fontFamily` (and `fontSize`
by type-impossibility); the remaining eight fields have no render consumer, measured, so the
property holds vacuously for them and cannot yet be asserted. **That obligation is inherited by
Story 4.1** (cell borders/padding) **and Story 4.8** (`altRowBackground`) — recorded against them
per D-3.5.4 so they inherit it rather than rediscover it.

### AC5 — Condition semantics are `if()`'s condition semantics, unchanged, and are asserted case by case

**Given** an element with `visibleIf`, **When** the condition is evaluated, **Then** all five
cases behave exactly as `if()`'s first argument does today
(`internal/expr/eval.go:148-175`), with **no** new rule on the axis:

| condition resolves to | outcome | inherited from |
|---|---|---|
| JSON `true` | element **visible** | `eval.go:156-159` |
| JSON `false` | element **hidden** | `eval.go:159-161` |
| explicit JSON `null` | element **hidden, silently** — no diagnostic of any severity | **D-3.2.3** (owner), `eval.go:162-163` |
| path **absent** from the data | **located Error**, render aborts, message names the path | **D-3.2.3**, `bind/text.go:401` |
| a **string** or a **number** (including `""`, `0`, `"false"`) | **located Error** — no truthiness | `eval.go:164-168`, AD-14 |

**Write this down explicitly and cite it, because the obvious wrong implementation is a
falsy-check and nothing in the tree currently forbids one.** Reusing `if()`'s ruled semantics is
the correct and cheap choice: the owner has already ruled this axis once (D-3.2.3), the
discriminator between the two nulls is already settled, and re-deciding it here would fork the
language's notion of truth across two features.

**Instrument: a table-driven test with one row per case above.** Each row is a data document and
an expected outcome, both owned by the test.

**Anchor: a literal the test owns.** The five cases are enumerated by the test, **not** derived
from `evalIf`'s switch — deriving them from the implementation is the D-000.68 case-4 defeat
(asking the code whether it agrees with itself).

**Red-proof.** Rows 1–3 fail at baseline (the field is inert; everything renders). Post-fix, the
discriminating mutation for row 5 is adding a truthiness fallback to the `default` arm; for row 3,
turning the null case into an error.

### AC6 — A literal `visibleIf` is rejected at LOAD, closing the asymmetry with `if()`

**Given** `"visibleIf": "42"` or `"visibleIf": "\"hello\""`,
**When** the template is loaded,
**Then** it is a **load error** naming the element — for the same reason `if()`'s condition slot
is `argNotLiteral`: the grammar has no boolean literal, so a bare literal can **never** be a
boolean, and this is decidable statically.

**Measured at creation (D4): both load clean today.** `checkVisibleIfExpression`
(`folio_expr_validate.go:103-112`) runs `expr.Parse` + `expr.Check`, and `argNotLiteral`
(`internal/expr/table.go:99`, checked at `internal/expr/check.go:118-126`) applies only to a
`CallExpr`'s arguments — a top-level `NumberLit` or `StringLit` is not a call, so nothing catches
it. Without this AC, `visibleIf: "42"` would load clean and fail only at render, which is exactly
what `folio_expr_validate.go:35-42` says must not happen: *"a malformed visibleIf … must fail at
load … not load clean and surface only when 3.5 starts evaluating it."*

**Note what is NOT decidable at load and must not be attempted:** whether a *path* yields a
boolean. `internal/expr/table.go:32-36` states that absent and explicit null are *"AD-14's OWN
distinct cases, decided at evaluation against real data, never here."* AC6 rejects **literals**
only.

**Anchor: a literal the test owns.** **Red-proof:** both subjects load clean at baseline; the
discriminating mutation post-fix is deleting the literal check.

### AC7 — Hiding suppresses output only; validation is unconditional

**Given** an element that is hidden by its condition **and** independently invalid — three
subjects: (a) a text element with an unresolvable `style.fontFamily` chain, (b) an image element
naming an absent asset key, (c) a table whose `bind` does not resolve to an array —
**When** the document is rendered,
**Then** each still produces its **existing** load/render error, unchanged in text and in
location, exactly as it does when the same element is visible.

**This is R2 made testable, and it is the AC most likely to be got wrong.** The precedent is
`render.go:601-616`, a Major QA finding on this exact failure mode: *"the SAME broken template
passing or failing depending on which report it was handed."* A visibility skip inserted at the
top of `collectBandTextRuns` / `collectImageRuns` — the naive placement — reproduces it.

**Anchor: the compiler and a literal the test owns** — the three broken templates are authored by
the test; the expected error text is pinned by the test, not read from the production string.

**Red-proof.** These pass at baseline (nothing is hidden yet, so nothing is skipped). Their teeth
are in the **discriminating mutation**, which the developer must run and record: move the
visibility skip **above** the validation calls in each of the three paths. Each subject must
redden. **An AC7 that does not redden under that mutation has not been implemented — it has been
placed somewhere harmless.**

### AC8 — A hidden element emits no diagnostics of its own

**Given** a text element that, when visible, emits `DiagCodeTextClippedWidth`
(`render.go:660-673`) because its content exceeds its declared `width`,
**When** the same element is hidden by its condition,
**Then** `Result.Diagnostics` contains **no** entry for that element, and the diagnostics of
every other element are unchanged in **content and in order**.

Ordering is document order — header, content, footer, then declaration order within a band
(`render.go:565-570`, `diagnostic.go:164-178`) — and removing an element from the middle must not
perturb the rest. A clip warning about a box nobody can see would be noise; more importantly,
`Diagnostics` is part of the public `Result` (`diagnostic.go:179-182`) and its determinism
guarantees are load-bearing.

**Anchor: a literal the test owns** — the expected diagnostic slice is written out in full by the
test.

**Red-proof.** Fails at baseline: today the element is not hidden and the clip warning is
emitted. Post-fix, the discriminating mutation is emitting diagnostics before the visibility
filter.

### AC9 — Visibility cannot depend on the page (AD-4)

**Given** a `visibleIf` on a **page-header** element and one on a **page-footer** element, in a
document that paginates to more than one page,
**When** the document is rendered,
**Then** each element's visibility verdict is **the same on every page** — it is decided once,
from the data scope alone, before any pagination.

**Why this needs its own AC.** Header and footer text is bound in **phase B**
(`render.go:1091-1106`), *after* `layout.Paginate` (`:1085`), with `headerFooterResolver(pageCount)`
in hand. An implementation that decides header visibility inside phase B has placed the decision
downstream of the page count — a latent AD-4 breach that is invisible until someone tries to make
it page-dependent. R1's single pre-phase-A computation makes it **structurally** impossible.

**Anchor: the type system / the compiler.** The strongest available form: the function computing
visibility verdicts takes the data scope and the bands and **does not receive `pageCount` or any
page-derived value in its signature** — so a page-dependent verdict is a compile error, not a
review catch. This is the same instrument `resolvePageRunForPage` uses on itself
(`page_number.go:437-443`: *"this function's own signature is the proof: it receives no FontSet,
no fontCache, nothing that could consult a font"*). Assert the signature as well as the
behaviour.

**Red-proof.** The behavioural half is not currently constructible in the "wrong" direction —
there is no way today to make visibility page-dependent, because `visibleIf` is inert. **The
signature half carries this AC's teeth and is fully anchored**; the behavioural half is a
multi-page regression guard. Stated plainly here per D-000.24 rather than credited with a
red-proof it does not have.

**FINISHER CORRECTION (Finding 2 / Major, ruled D-3.5.3).** What shipped was a **comment**
claiming the signature was an anchor — no assertion existed anywhere, and no type system forbids
adding a parameter to a function, so "cannot receive page-derived state" was never a constructible
compile-time property on its own. **Ruled instrument (stronger than what was asked for):** in
`lint`, using the `go/types` machinery from Story 3.3, `ScanVisibilityComputationSignature`
(`lint/internal/rules/visibilitysignature.go`) asserts `computeVisibility`'s resolved parameter
type list equals a literal the test owns — **any** input growth reddens it, page-derived or not,
because closing the input set outright is the real property AD-24/AD-4 need. Pinned to a literal
per D-000.68 (the set is permanent: AD-24 forbids row-level visibility outright, so nothing in
Epics 4–6 grows it). Red-proved (`visibilitysignature_test.go`): a widened 5-parameter signature,
and a same-count-different-type signature, both trip it; the real `computeVisibility` passes with
zero findings.

**The behavioural half was also blind, for a specific reason, not merely weak.**
`len(p.Runs) == 0` cannot discriminate, because the content element's runs land on every page
unaided — proved by mutation: making header/footer page-dependent (`if pageNum == 1`) left
`TestVisibilityIsConsistentAcrossPages` **green**. Fixed: the visible-direction assertion now
checks that BOTH the header's and the footer's run (by `SourceText`) are present on **every**
page — a set, not a count. Re-run against the same mutation: **reddens** on the first page missing
one of the two.

### AC10 — `expr.Resolver` is not widened

**Given** the implementation of every AC above,
**When** `lint/` runs,
**Then** `ScanResolverMethodSet` reports **zero** findings and
`folio-go/internal/expr_arch_test.go:371 TestExprResolverMethodSetIsClosed` passes.

The seam already exists and already suffices: visibility resolves a **path**, which is
`Resolve(path []string) (Value, error)` — the interface's first method. **D-3.3.1** closed this
set at three methods and `lint/internal/rules/resolvermethodset.go` enforces it with `go/types`
exact set-and-signature equality, expanding embedded interfaces (`:105-172`). The rule's own
message frames a new method as *"a page-scoped variant under ANY spelling … is a direction change
under AD-4."*

**Anchor: the type system**, via an already-shipped instrument. This AC adds no new guard; it
records that an existing one must stay green, and that satisfying R4 by adding a **method to the
interface** instead of a **function on `bind`** is the forbidden shortcut.

### AC11 — The three-module gate, with the two required reds unchanged

**Given** the completed implementation,
**When** the gate runs,
**Then**:

| scope | expectation |
|---|---|
| `folio-go/` | `go build ./...`, `go vet ./...`, `gofmt -l .` all clean; tests **≥ 822 pass**, **exactly 1 fail**, 1 skip |
| `lint/` | **98 pass · 3 fail** — the three DW-19 failures, unchanged. A **fourth** is a regression owned by this story |
| `hashmatrix/` | **3 pass** |

**The one `folio-go` failure must still be `TestCorpusMeetsP6ExerciseFloors`, with stats
byte-identical to `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.** D-000.57's third
clause. Never fill the floor; never weaken the lint rules.

---

## Decisions raised at creation

### DECISION-1 (for the lead) — should a `{{ }}` inside a style field also be REJECTED at load?

**The question.** AC4 asserts conditional formatting is inert and byte-neutral. Should the story
*additionally* make `"style": {"background": "{{...}}"}` a **load error**?

**Measured facts.** Style string fields are entirely unvalidated (`parse_bands.go:479-490`; no
colour-format check exists anywhere), and `checkTextExpressions` never sees them
(`folio_expr_validate.go:22-64`). Such a template loads clean today and renders inertly. Nothing
in the module consumes `style.Background` outside `internal/template` at HEAD — background
painting arrives with Epic 4/5.

**Options.**

- **(a) Behavioural inertness assertion only — AC4 as written.** Matches the source AC's wording
  exactly (*"unsupported and out of scope"*, not *"rejected"*), has a genuine red-proof, adds no
  new rejection surface.
- **(b) (a) plus a load-time rejection of `{{` in any style string field.** A harder fence, also
  red-provable. But it invents a rejection the source AC does not ask for, and it rejects on a
  **syntactic proxy** (the two-brace spelling) rather than on the property — the shape
  **D-000.15** warns against, and it would reject a legitimate literal string that merely
  contains braces.
- **(c) Defer the fence to Epic 5, when a designer UI could offer such a field.**

**Recommendation: (a).** It is what the AC says, it is honestly red-provable, and (b)'s anchor is
a spelling rather than a property. If the lead wants a hard fence, (b) is defensible — but it
should then be scoped to the style fields that will actually be *consumed* as colours, and
recorded as a scope widening rather than folded in silently.

**Either way AC4 stands unchanged**, so this does not block development; it only decides whether
a twelfth AC is added.

### DECISION-2 (for the lead, low stakes) — is AC6 in scope?

AC6 closes the load/eval asymmetry measured in D4 (`visibleIf: "42"` loads clean; `if(42, …)`
does not). It is not in the source AC list. **Recommendation: yes, in scope** — it is a defect in
*this story's own field*, `folio_expr_validate.go:35-42` explicitly demands the behaviour AC6
provides, it is a handful of lines beside an existing check, and leaving it means shipping a
field whose malformed values surface at render rather than at load. Raised rather than assumed
because it is a scope addition.

**Neither decision blocks the start of development.** AC1, AC2, AC3, AC5, AC7–AC11 are fully
determined by the record.

---

## Task breakdown

1. [x] **Re-measure the baseline** — all three modules, before writing a line. Confirm 822/1/1,
   98/3, 3/0 and the P6 stats. Record them in the Delivery Log (D-000.26).
2. [x] **Write AC1, AC3, AC5 (rows 1–3), AC6 and AC8 FIRST and observe them fail.** These are the
   ACs with baseline reds; record the observed failure text for each. An AC in this set that
   passes before implementation has been mis-written — stop and fix the test, not the code.
3. [x] **Add the bare-expression evaluator to `internal/bind`** (R4) — an exported function over the
   existing unexported `exprResolver`. **Do not touch `expr.Resolver`** (S7, AC10).
4. [x] **Add the single pre-phase-A visibility computation in `renderDocument`** (R1), with a
   signature that cannot receive page-derived state (AC9).
5. [x] **Consume the verdicts** in `collectImageRuns`, `collectBandTextRuns` and the content-column
   item construction — **after** each element's own validation (R2, AC7) and **before**
   `ColumnItem` construction (R3).
6. [x] **Add the column-level rejection** in `decodeColumn` (AC3), beside the existing three column
   checks.
7. [x] **Add the literal-condition rejection** to `checkVisibleIfExpression` (AC6), subject to
   DECISION-2.
8. [x] **Run every discriminating mutation named in the ACs** — especially AC7's (move the skip above
   validation; all three subjects must redden) and AC5's (add a truthiness fallback). Record each
   mutation and its observed result. A guard whose mutation was not run is not yet a guard
   (D-000.68).
9. [x] **Sweep, do not spot-fix** (D-000.67 part 2). Any comment in the tree asserting `visibleIf` is
   ignored at render is now false — sweep the enclosing files and **report the count of sites
   examined**, not merely that the named one was fixed. Known site:
   `folio_expr_validate.go:35-42` (forward-looking; re-read it, it may now be satisfied rather
   than stale).
10. [x] **Update `folio-format.md`** — `:184`'s *"Not valid on a table column"* becomes enforced
    rather than aspirational; say so, and document the two-nulls distinction (R6), which the
    field table does not currently draw.
11. [x] **Run the three-module gate** (AC11) and confirm the two required reds are unchanged, P6
    stats byte-identical.
12. [x] **Set the story status to `review`.**

---

## Delivery Log

**Baseline re-measured at `ae9e551` before writing a line (Task 1).** `folio-go`: build/vet/gofmt clean; **822 pass · 1 fail · 1 skip**, P6 stats byte-identical to `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. `lint`: **98 pass · 3 fail** (the three DW-19 `.font-sources` failures). `hashmatrix`: **3 pass**. All match the story's recorded figures exactly.

**Two decisions were sent to the engineering lead** (DECISION-1, DECISION-2) and both rulings arrived before implementation proceeded — see the story's own DECISION-1/DECISION-2 sections above for the questions; the rulings were: **DECISION-1 REJECTED** (a `{{ }}` placeholder in a style string field is now a load error, not merely behaviourally inert — anchored on `expr.ScanPlaceholders`, never a `strings.Contains` spelling match), and **DECISION-2 accepted, IN SCOPE** (AC6's literal-condition rejection, implemented by hoisting one shared predicate, `expr.IsLiteralExpr`, called by both `checkVisibleIfExpression` and `if()`'s own `checkArgKind` — not two independent literal checks).

**Task 2 — baseline reds observed before implementation.** AC1/AC2 (visibility inert: both data sets render identically, byte-for-byte, before any run was written), AC3 (a column-level `visibleIf` loaded clean, absorbed into `Column.Extra`), AC5 rows 1–3 (every condition value rendered, none hidden), AC6 (`visibleIf: "42"`/`"'x'"` loaded clean), AC8 (the clip diagnostic fired regardless of any would-be condition). All five failed at baseline as expected; none passed before implementation (no mis-written test found).

**Implementation (Tasks 3–7).**
- `internal/bind/condition.go` (new): `EvaluateCondition`, the exported bare-expression evaluator over the existing unexported `exprResolver` (R4). `expr.Resolver`'s method set is untouched (S7) — confirmed by `TestExprResolverMethodSetIsClosed` and `lint`'s `ScanResolverMethodSet`, both still green (AC10).
- `render_visibility.go` (new): `visibilityVerdicts`, `isVisible`, and `computeVisibility` — the single pre-phase-A visibility computation (R1). Its signature (`bands, data, params, fc` — no `pageCount`, no page-derived value) is AC9's compile-time anchor, the same instrument `resolvePageRunForPage` uses on itself.
- `render.go`: `renderDocument` split into `buildPageModel` (page-model construction) + a thin serialization wrapper, so a test can assert on `pagemodel.Page.Runs/Images` directly (AC1's own anchor) without decoding PDF bytes — same derivation, not a second one. `collectBandTextRuns` consumes verdicts AFTER `bind.BindTextSpans`, `fontChain` and `shapeSegments` all run unconditionally (R2/AC7) — the skip sits immediately after `shapeSegments` succeeds, before any packed-line/overflow/vertical-model/positioning work, and gates both the clip diagnostic and the caveat-derived diagnostics (AC8). `collectImageRuns` is DELIBERATELY left unfiltered (every image element, visible or not, still goes through width/height/asset-presence checks AND the later deduplicated asset-resolution pass) — visibility is instead consulted at PAGE-MODEL CONSTRUCTION, in `contentColumnItems` (phase-A) and `paginateDocument` (final), which is the first point after every image validation has run where a hidden image's placement would otherwise be built.
- `internal/expr/eval.go`: `ConditionValue` extracted from `evalIf`'s inline switch — the one place D-3.2.3's true/false/null-is-false/no-truthiness axis lives, shared verbatim by `if()` and visibility.
- `internal/expr/check.go`: `IsLiteralExpr` extracted from `checkArgKind`'s `argNotLiteral` case — the one predicate for "is this a literal", shared by `if()`'s condition slot and `checkVisibleIfExpression` (AC6/DECISION-2).
- `folio_expr_validate.go`: `checkVisibleIfExpression` gains the literal rejection (AC6) via `expr.IsLiteralExpr`; new `checkStyleHasNoPlaceholders` (DECISION-1) rejects a placeholder in any style string field (`align`, `background`, `fontFamily`, `valign`, `border.color`, `border.edges`) via `expr.ScanPlaceholders`, explicitly labelled as NOT general style validation.
- `internal/template/parse_bands.go`: `decodeColumn` gains AC3's pure field-presence rejection of `visibleIf`, beside the three existing AC43 column checks.

**Task 8 — discriminating mutations, run and reverted (not committed).**
1. AC7, text subject: moved the visibility skip in `collectBandTextRuns` to BEFORE `bind.BindTextSpans`/`fontChain`/`shapeSegments` — `TestHiddenElementStillFailsItsOwnValidation/unresolvable_fontFamily_chain` reddened (`expected a located render error, got nil`). Reverted; test green again.
2. AC7, image subject: temporarily filtered `collectImageRuns` by visibility BEFORE the width/height/asset-presence checks (the design this story deliberately does NOT ship) — `TestHiddenElementStillFailsItsOwnValidation/asset_key_absent_from_assets_map` reddened. This is exactly why visibility is consulted at page-model construction (`contentColumnItems`/`paginateDocument`) for images, never at collection. Reverted.
3. AC5 row 5: added a truthiness fallback (`KindString` → `v.Str != ""`) to `ConditionValue` — `TestVisibleIfConditionSemanticsMatchIf/string_is_a_located_error,_no_truthiness` reddened. Reverted.
4. AC5 row 3: turned `ConditionValue`'s `KindNull` case into an error — `TestVisibleIfConditionSemanticsMatchIf/explicit_null_is_hidden,_silently` reddened. Reverted.

Each mutation was built, run to confirm reddening, then restored from a pre-mutation copy (never `git checkout`) and re-verified green before continuing.

**Task 9 — sweep.** Searched every non-test `.go` file for `visibleIf`/"ignored at render"/"nothing reads it" claims: 5 files matched (`folio_expr_validate.go`, `render_visibility.go`, `internal/template/serialize.go`, `internal/template/parse_bands.go`, `internal/expr/eval.go`). The one forward-looking claim, `folio_expr_validate.go:35-42` ("a malformed visibleIf … must fail at load … not load clean and surface only when 3.5 starts evaluating it"), is now satisfied rather than stale — re-read and left as-is; no other site asserted `visibleIf` is ignored.

**Task 10 — `folio-format.md` updated.** The `visibleIf` field-table row now states load-time enforcement (literal rejection, column rejection) and the two-nulls distinction (R6: field absent / present-null both mean "visible, nothing to evaluate", vs. condition **resolving** to null, which hides silently). The `style` section gained a note on the new placeholder rejection, explicitly scoped as "not general style validation".

**Task 11 — final three-module gate, matching AC11 exactly:**
- `folio-go`: build/vet/gofmt clean; **847 pass · 1 fail · 1 skip** (25 new tests, all passing; the required red is still `TestCorpusMeetsP6ExerciseFloors`, stats byte-identical: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`).
- `lint`: **98 pass · 3 fail** — the same three DW-19 `.font-sources` failures, unchanged. No fourth failure.
- `hashmatrix`: **3 pass**, unchanged.

**Render-path note for the Epic 3 boundary matrix (due after 3.7):** this story touches the render path (`render.go`, `page_number.go`) — it removes runs/placements for hidden elements and adds one pre-phase-A pass over already-decoded JSON; it adds no font byte, no image byte, no new PDF operator class and emits nothing through a code path other than the already-pinned shaping/serialization path. Flagging per the launch brief's request rather than assuming it is out of scope.

---

## Finisher Delivery Log

**Reviewed by:** bmad-code-reviewer, findings triaged and resolved by the finisher (Story 3.5), 2026-08-26. **Disposition: all twelve findings FIX** — see "Finding Resolutions (finisher pass)" above for each, and the AC-by-AC disposition table for the resulting per-AC verdicts.

**Two items were with the engineering lead (AC9, AC4) plus the matrix-override question; all three rulings arrived and are implemented, not merely recorded**:
- **D-3.5.3 (AC9)** — `lint`'s `ScanVisibilityComputationSignature` pins `computeVisibility`'s parameter type list by set-equality to a literal; the behavioural half now asserts a per-page element-presence SET, not a run count. Both red-proved (widened signature; `if pageNum == 1` mutation), both reverted after confirming.
- **D-3.5.4 (AC4)** — subject corrected to `fontFamily` (the one style field that is both a string and has a render consumer); the eight-field population that cannot yet be asserted is stated explicitly and handed to Stories 4.1/4.8. Red-proved (font-chain resolution made to consult `data`), reverted.
- **D-3.5.5 (`table.altRowBackground`, Finding 4)** — the check now covers it, and a reflection-based test derives the covered population from the schema (`template.Style`/`Border`/`TableExt`/`Column`) rather than trusting a hand list, with a small, individually-reasoned exclusion list for the schema's non-style string fields. Red-proved twice (behavioural: placeholder in `altRowBackground` now rejected; completeness: an uncovered field added to the schema reddens the population test), both reverted.
- **D-3.5.6 (matrix override)** — declined, as ruled: building `pdfImages` from visible runs is integer/set work, no new source of cross-target divergence. Per the same ruling's correction, this story DOES register a new `matrixDocuments` entry (`hidden-image`, since Finding 1's fix is exactly the artifact the override question was about) and runs its **native leg** (host, darwin/arm64) under D-000.54 as a **sequencing** step — logged here, not as a D-000.4 override. Command and result: `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .` → **1 passed**, hash recorded at `fixtures/hidden-image/expected.json`. The other three targets' legs are deferred to the Epic 3 boundary gate.

**Blocker 1's golden-fixture obligation, measured not assumed:** `git show ae9e551:folio-go/render_test.go | grep -c "visibleIf"` → **0**. At baseline exactly one image element existed in the module's own template literals and it never carried `visibleIf`. No existing golden fixture contains a hidden element; none moved; nothing required an owner decision before committing.

**Every discriminating mutation this pass added was run and reverted from a pre-mutation copy** (never `git checkout`), the same discipline the story's own Task 8 used: Finding 1 (pdfImages visible-filter reverted → reproduces the reviewer's exact byte counts), Finding 2's signature pin (widened to 5 params; same-count-different-type), Finding 2's behavioural half (`if pageNum == 1`), Finding 3 (font-chain made data-dependent), Finding 4 behavioural (`altRowBackground` placeholder) and completeness (extra schema field). `git status --short` immediately after each mutation-and-revert cycle was byte-for-byte the pre-mutation status.

**Final three-module gate, re-measured after all finisher fixes:**
- `folio-go`: `go build ./...`, `go vet ./...`, `$(go env GOROOT)/bin/gofmt -l .` all clean. **853 pass · 1 fail · 1 skip**, 15 packages (6 new tests beyond the developer's 847: the image byte-identity test, the altRowBackground pair, the schema-population test, plus the two `internal/bind` `EvaluateCondition` literal tests). The one failure is still `TestCorpusMeetsP6ExerciseFloors`, stats byte-identical: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.
- `lint`: `go test ./... -count=1` → **104 pass · 3 fail** (6 new tests: `ScanVisibilityComputationSignature`'s production scan, coverage-statement wording, compliant fixture, and two red-proofs, plus the not-found guard). The three failures are the same DW-19 `.font-sources` cases, unchanged. No fourth failure.
- `hashmatrix`: **3 pass**, unchanged.

**`go build -tags=matrix ./...` and `go vet -tags=matrix ./...`** both clean, confirming the new matrix-tagged code (`captureHiddenImageRender`, `requireNoImageXObject`, the `hidden-image` `matrixDocuments` entry) compiles and vets under the tag the CI legs actually use.

## File List

Developer's files (unchanged from the review):
- `folio-go/internal/bind/condition.go` (new)
- `folio-go/render_visibility.go` (new)
- `folio-go/visibility_test.go` (new)
- `folio-go/render.go`
- `folio-go/page_number.go`
- `folio-go/collect_text_runs_composition_test.go`
- `folio-go/folio_expr_validate.go`
- `folio-go/folio_expr_validate_test.go`
- `folio-go/internal/expr/eval.go`
- `folio-go/internal/expr/check.go`
- `folio-go/internal/template/parse_bands.go`
- `_bmad-output/specs/spec-folio/folio-format.md`
- `_bmad-output/implementation-artifacts/3-5-hide-a-component-conditionally.md` (this file)
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

Finisher's additional files (Finding 1, 2, 3, 4, 7, 9, 10, 11, 12 — see Finding Resolutions):
- `lint/internal/rules/visibilitysignature.go` (new — AC9's real anchor, Finding 2)
- `lint/internal/rules/visibilitysignature_test.go` (new)
- `folio-go/internal/bind/condition_test.go` (new — Finding 10's direct unit coverage)
- `folio-go/matrix_test.go` (new `hidden-image` `matrixDocuments` entry, Finding 1)
- `folio-go/render_test.go` (new `hiddenImageTestTemplateJSON` / `FOLIO_SUBPROCESS_RENDER_HIDDENIMAGE` selector)
- `folio-go/byte_neutrality_test.go` (`declaredEpic2GateObligations` +1 line for `hidden-image`)
- `fixtures/hidden-image/expected.json` (new — the registered document's recorded golden, host leg measured)
- `.github/workflows/matrix.yml` (four target artifact lists + the compare-hashes `docs` list, +`hidden-image`)
- `folio-go/render.go` (Finding 1: `pdfImages` built from visible-only asset keys; Finding 7: element id in the "asset not present" error)
- `folio-go/page_number.go` (Finding 6: doc-comment naming unconsumed kinds)
- `folio-go/render_visibility.go` (Finding 6: doc-comment; Finding 12: caveat-discard comment)
- `folio-go/folio_expr_validate.go` (Finding 4: `table.altRowBackground` check + shared helper + corrected doc comment)
- `folio-go/folio_expr_validate_test.go` (Finding 4: two behavioural tests + the schema-reflection completeness test)
- `folio-go/internal/bind/condition.go` (Finding 10: `IsLiteralExpr` check inside `EvaluateCondition`)
- `folio-go/visibility_test.go` (Finding 1: image byte-identity test; Finding 2: AC9 signature-pin test comment + per-page set fix; Finding 3: doc comment corrected; Finding 5: regression-guard label; Finding 7: assertion strengthened; Finding 9: rebuilt middle/cross-band test)
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` (Finding 11 — the developer's uncommitted D-3.5.1/D-3.5.2/D-3.6.1/D-000.68-correction entries, plus the engineering lead's D-3.5.3/D-3.5.4/D-3.5.5/D-3.5.6 rulings this pass implements; append-only, not otherwise edited)

## Change Log

| Date | Change |
|---|---|
| 2026-08-26 | Story created at `ae9e551`. Four brief/tree divergences measured and folded in (D1–D4). AC3's subject verified constructible by execution. Two decisions raised, neither blocking. |
| 2026-08-26 | Implemented. Both raised decisions ruled by the engineering lead (DECISION-1 rejected/hardened; DECISION-2 accepted, in scope, implemented via a single shared predicate). All 12 tasks and all 11 ACs complete; four discriminating mutations run and reverted; three-module gate matches AC11 exactly (847/1/1, 98/3, 3/0). Status set to `review`. |
| 2026-08-26 | Reviewed (bmad-code-reviewer): 1 Blocker, 3 Majors, 6 Minors, 2 Nits. |
| 2026-08-26 | Finished. All 12 findings FIX (see Finding Resolutions). Blocker 1 (hidden image byte-embedded) fixed by building `pdfImages` from visible-only asset keys; AC9 (D-3.5.3) and AC4 (D-3.5.4) rulings implemented; `table.altRowBackground` (D-3.5.5) covered via a schema-derived completeness test; a new `hidden-image` matrix document registered with its native leg run (D-000.54, D-3.5.6). Golden-fixture check measured: N=0 existing goldens moved. Three-module gate: 853/1/1 · 104/3 · 3/0, no regression. Status set to `done`. |

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer (adversarial, Story 3.5, explicit by number — no auto-pick)
- **Date:** 2026-08-26
- **Baseline:** `ae9e551`, uncommitted working tree
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 1
- **Majors:** 3
- **Minors:** 6
- **Nits:** 2

### Gate, re-measured in this review (not carried forward — D-000.26)

| scope | invocation | measured |
|---|---|---|
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l . && go test ./... -count=1` | build **clean**, vet **clean**, `gofmt -l .` **clean** (run as `$(go env GOROOT)/bin/gofmt` to defeat the rtk list filter), **847 pass · 1 fail · 1 skip**, **15 packages** |
| `lint/` | `go test ./... -count=1` | **98 pass · 3 fail** |
| `hashmatrix/` | `go test ./... -count=1` | **3 pass** |

The one `folio-go` failure is `internal/text TestCorpusMeetsP6ExerciseFloors` — the **required
red**. Stats verified **byte-identical**: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`
(P6g 7, floor 20). No drift. The three `lint` failures are the same three DW-19
`.font-sources` cases (`manifest.TestManifestUpToDate`,
`manifest.TestResolveAssetsIncludesWordlist`,
`rules.TestFontsAssetsNoticeRemovalRedProof`) — **no fourth**. AC11 is met as stated.

`TestExprResolverMethodSetIsClosed` **passes** and `lint`'s `ScanResolverMethodSet` reports zero
findings — **AC10 satisfied**, the seam is still exactly `{Resolve, CollectionLength,
ProjectCollection}`. `EvaluateCondition` correctly lands as a function on `bind`, not a method on
the interface.

### Method, and how the mutations were run

**Nothing in the repository was written to during this review.** Every mutation and probe ran in
an isolated `rsync` copy of the `folio-go` module under the scratchpad, restored from a pristine
copy between mutations (never `git checkout`). `git status --short` at the end of the review is
byte-for-byte the status at the start.

**Every discriminating mutation the Delivery Log claims was re-run independently, and all four
redden as reported** — a mutation reported is not a mutation verified, so these were rebuilt from
scratch rather than taken on the record:

| # | mutation | expected red | observed |
|---|---|---|---|
| M1 | visibility skip moved to immediately after `elVisible := isVisible(...)`, i.e. **above** `bind.BindTextSpans` / `fontChain` / `shapeSegments` | AC7 text subject | **RED** — `TestHiddenElementStillFailsItsOwnValidation/unresolvable_fontFamily_chain` |
| M2 | `imageRuns` filtered by visibility immediately after `collectImageRuns`, before the asset-resolution pass | AC7 image subject | **RED** — `.../asset_key_absent_from_assets_map` |
| M3 | truthiness fallback (`KindString → v.Str != ""`) added to `expr.ConditionValue` | AC5 row 5 | **RED** — `TestVisibleIfConditionSemanticsMatchIf/string_is_a_located_error,_no_truthiness` |
| M4 | `expr.ConditionValue`'s `KindNull` arm turned into an error | AC5 row 3 | **RED** — `.../explicit_null_is_hidden,_silently` **and** `TestVisibleIfNullConditionEmitsNoDiagnostic` |

Four further mutations were run that the story did not claim:

| # | mutation | observed |
|---|---|---|
| M5 | **`expr.IsLiteralExpr` neutered to `return false` — ONE change** | **RED on all three**: `TestParseTemplateRejectsLiteralVisibleIf/number_literal`, `/string_literal`, **and** `TestIfConditionStillRejectsLiteralAfterVisibleIfSharesThePredicate` |
| M6 | `decodeColumn`'s `visibleIf` presence check deleted | **RED** — `TestParseTemplateRejectsVisibleIfOnTableColumn` |
| M7 | the text visibility skip in `collectBandTextRuns` deleted | **RED** — AC1, AC2, AC5 rows 2–3, AC8 |
| M8 | header/footer runs made **page-dependent** (emitted only on page 1) | **GREEN** — see Finding 2 |

### Teeth results on the three questions this review was pointed at

**1. AC7/AC8 — "absent" is a claim about the PageModel, not about the pipeline. VERIFIED, by
construction, and the design holds.** Four independent subjects were built from scratch and each
still fails while hidden, with the identical error text it produces while visible:

- hidden text element with an unresolvable font chain → `folio: Render: element e1: no font in chain … has a glyph for rune U+0E01` / `nonexistentChain`
- hidden text element with a **bad binding** → `folio: Render: bind: element e1: data path "customer.nope" is absent from the report data`
- hidden text element with a **malformed expression** → rejected at **load**: `folio: ParseTemplate: element e1: unknown function "nosuchfunc" …`
- hidden **image** element naming an absent asset key → `folio: Render: an image element references asset … which is not present in the document's assets map`

M1 and M2 then confirm the placements carry teeth in both paths. The developer's account is
accurate: `collectBandTextRuns` consumes the verdict strictly after `BindTextSpans`, `fontChain`
and `shapeSegments`, and leaving `collectImageRuns` unfiltered is genuinely load-bearing —
M2 reproduces exactly the `render.go:601-616` defect for images. **No route was found by which a
hidden element's defects can be smuggled past validation.** The QA-Finding-5 blind spot is not
repeated.

**2. D-3.5.1's hoisted predicate — VERIFIED as a genuine single implementation.** M5 changes one
function body and reddens `"visibleIf": "42"`, `"visibleIf": "\"hello\""` **and** `if(42, …)`.
One predicate, two call sites, no second copy. `expr.IsLiteralExpr` is extracted from
`checkArgKind`'s `argNotLiteral` arm and `argNotLiteral` now delegates to it. D-000.38's failure
mode is avoided. D-3.5.2's mechanism is likewise correct: `checkStyleHasNoPlaceholders` runs
`expr.ScanPlaceholders`; a repo-wide grep finds **no** `strings.Contains("{{")` brace match
anywhere in the new code (the only hits are the doc comment saying it is not used, and two
pre-existing `_test.go` sites unrelated to this story). The check is labelled as **not** style
validation, and `TestParseTemplateAcceptsOrdinaryStyleValues` pins that hex colours remain
unvalidated. **See Finding 4 for where the ruled scope was not fully reached.**

**3. AC9's compile-time anchor — NOT an anchor.** See Finding 2. No live smuggling route exists
today (verified: `bandWithOrigin` is `{template.Band, geom.Length}`; `bind.Value` is a plain
decoded-JSON struct, not an interface and not a closure; `expr.FormatContext` is
`{Locale, UTCOffset string}`; and `computeVisibility` is called at `render.go:1152`, **before**
`pageCount` is declared at `:1178`, so referencing it there is a compile error today). But the
anchor is an unenforced property of where the call happens to sit, plus a comment — not an
instrument. AC9's own text mandated asserting the signature and that was not done.

### Verified and correct — recorded so the finisher does not re-litigate them

- **AC1** — text and image halves both hold; hidden contributes exactly zero `Page.Runs` / `Page.Images`.
- **AC2, second clause / "the gap"** — verified independently at four positions the shipped test does not cover: hidden **first**, **last**, **middle**, and **sole occupant** of the content band. Surviving siblings keep identical `(X, Y)` in every case (e.g. `Text e3` at `(0, 220000)` whether `e1` is visible or hidden); the sole-occupant case still yields exactly one page and no error. No reflow in either direction.
- **AC3** — `decodeColumn`'s check fires, names the column (`e3`) and states the reason ("elements only"); M6 confirms the red-proof. Spec and code now agree — `folio-format.md:184` is enforced for the first time.
- **AC5** — all six condition kinds probed independently: `true` → visible; `false`, explicit `null` → hidden with **zero** diagnostics; absent path → located error naming the path; `0`, `""`, `"false"` → `expr: element e1: visibleIf must be a boolean, got number|string (no truthiness — AD-14)`; `[]` and `{}` → located `not a scalar value usable in an expression`. **Nothing coerces.**
- **AC6** — both literal spellings rejected at load, naming the element.
- **AC10** — resolver method set untouched; both instruments green.
- **Item 7 (no new diagnostic code)** — confirmed: `git diff ae9e551 -- folio-go | grep '^+.*DiagCode'` returns nothing. Every condition is a load error or an inherited `if()` render error. `internal/diag` remains 3.6's. The design did not drift.
- **Item 9 (the `renderDocument` → `buildPageModel` split)** — behaviour-preserving. `renderDocument` is now a wrapper calling `buildPageModel` then `pdf.SerializeTextDocument` in the identical order; every golden test passes (`TestRenderMatchesGoldenFixture`, `TestMultiPageGoldenMatchesTheCommittedArtifact`, `TestThreeBandPageGoldenFixture`, `TestGoldenDigestAgreesAtEveryDeclaredSite`, and the rest). **No existing golden carries a `visibleIf`** (`grep -rn visibleIf fixtures/ folio-go/testdata/` → 0), so no pre-existing document changes behaviour under the new evaluation. No output bytes moved for any existing golden. **But see Finding 1** — this diff *does* move bytes for a new class of document, and Epic 3's boundary matrix run after 3.7 will measure it.
- **AD-1 / AD-23 (item 10)** — clean. The added lines contain no `float32`/`float64`/`math/big.Float`/`math/big.Rat`, no `time`/`os`/`math/rand`/`net`/`math`, no dot-import, no package-level mutable state. `visibilityVerdicts` is a map but is only ever **looked up**, never ranged over, so no output-reaching map iteration. `computeVisibility` walks `bands` and `band.Elements` slices in authored order.
- **Item 11 (the spec edit)** — the `visibleIf` row is correct on every clause I checked against behaviour (bare expression, evaluated during bind, `if()` semantics verbatim, the two-nulls distinction, literal rejection, column rejection). It is long but not padded. The style-section paragraph is accurate about *what it checks*; **Finding 4** is about what it invites a reader to infer.
- **The plain-terms opener** — meets its obligation. It explains why hiding leaves a gap rather than closing up, in prose, with no identifiers, and gives the real reason rather than a restatement: *"if the position of a component depended on whether its neighbour happened to be hidden, then the same template would produce a different-looking document for every customer, and page breaks would land in different places for every customer."* That is the one thing a reader would call a bug, and it is answered. See **Finding 8** for the one thing it now omits.

---

### Finding 1: A hidden image is absent from the page model but still embedded in the PDF — AC2's byte-identity fails for every image element

- **Severity**: **Blocker**
- **Category**: AC Conformance
- **Location**: `folio-go/render.go:1308-1345` (`assetKeys` / `firstElementIDByAssetKey` / `pdfImages` construction inside `buildPageModel`); test gap at `folio-go/visibility_test.go:161` (`TestVisibleIfHiddenIsByteIdenticalToElementDeleted`)
- **Observation**: `assetKeys` is derived from the **unfiltered** `imageRuns` slice. Visibility is consulted only later, at `contentColumnItems` (`page_number.go:84`) and `paginateDocument` (`render.go:1518`), which drop the *placement*. The `/XObject` itself is already in `pdfImages` by then and is serialized regardless. Measured by construction, in an isolated copy of the module:

  ```
  hidden-image render:    8686 bytes,  /Subtype /Image present=true,  /XObject present=true
  element-deleted render: 8316 bytes,  /Subtype /Image present=false, /XObject present=false
  ```

  Both templates carry the identical `assets` map and the identical always-visible text sibling; the only difference is whether the image element is declared-and-hidden or deleted. AC2 says *"the output bytes are **identical** to rendering a second template that is the first with element `X` **deleted from the JSON**"* — unconditionally, with no carve-out for element kind, and the story itself treats images as in scope for this pair (AC1 has an explicit image half, `TestVisibleIfFalseImageContributesZeroImages`). The shipped AC2 test exercises **only** the text subject (`visTextTemplateJSON`); the image subject is never rendered through `Render` and compared.
- **Impact**: A shipped acceptance criterion fails for one of the two element kinds that can reach the page model, and nothing detects it. Three consequences beyond the AC: (i) the output grows by the full compressed size of every hidden image, unbounded in the number of hidden images; (ii) content the page model declares **absent** is nevertheless recoverable from the file — a statement that hides a signature block, a watermark or a customer photo still ships those bytes to the recipient; (iii) this is a render-path byte change in a class of document Epic 3's boundary matrix run (due after 3.7) will measure, and it is currently unrecorded as such.
- **Suggested Resolution**: Do **not** fix this by filtering `collectImageRuns` or `imageRuns` earlier — mutation M2 proves that breaks AC7 subject (b), which is precisely why the current design exists. Instead keep the validation pass over **all** image runs exactly as it is (asset presence, `DecodeAssetBytes`, `DecodeImageForRender` — every image element still errors when broken, hidden or not), and build the **`pdfImages` map** from the asset keys reachable from **visible** image runs only. Validate the union; embed the visible subset. Then add the image half of `TestVisibleIfHiddenIsByteIdenticalToElementDeleted` and re-run M2 to confirm AC7(b) still reddens under the earlier-filter mutation. Record the byte-movement note for the boundary matrix.
- **Related AC**: AC2 (primary), AC1 (image half), AC7 (the constraint the fix must not break)

### Finding 2: AC9 ships with no instrument — the signature "anchor" is a comment, and the behavioural guard cannot detect a page-dependent verdict

- **Severity**: **Major**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/render_visibility.go:72` (`computeVisibility`'s signature); `folio-go/visibility_test.go:552-621` (`TestVisibilityIsConsistentAcrossPages`), specifically the visible-direction loop at `:600-604`
- **Observation**: Two halves, both hollow.

  **(a) The signature is asserted nowhere.** `grep -rn computeVisibility folio-go` returns exactly four sites: the call at `render.go:1152`, the definition at `render_visibility.go:72`, two of its own doc-comment lines, and one **comment** in `visibility_test.go:554` claiming *"computeVisibility's SIGNATURE … is AC9's real anchor"*. AC9's own text says **"Assert the signature as well as the behaviour."** No assertion exists. Under D-000.68 this is case 4 — the guard is anchored to the code's own declaration, which the code under test can move in the same edit that breaks the property. (The cited precedent, `resolvePageRunForPage`, has the same weakness at `page_number_test.go:110`; that is an argument for fixing both, not for repeating it.)

  **(b) The behavioural half cannot fail in the direction that matters.** Mutation M8 made header and footer runs literally page-dependent — emitted on page 1 only:

  ```go
  if pageNum == 1 {
      for _, ref := range header.Runs { … }
  }
  ```

  `TestVisibilityIsConsistentAcrossPages` stayed **GREEN**. The visible-direction assertion is `if len(p.Runs) == 0`, and the content element (`strings.Repeat("word ", 400)`) puts runs on every page unaided, so the loop cannot discriminate whether the header/footer runs are there. Only the hidden direction (which scans `r.SourceText` for `"HEADER"`/`"FOOTER"`) has teeth, and it catches only a verdict flipping *to visible*; a verdict flipping *to hidden* on a later page is invisible.
- **Impact**: The AC-4 co-primary invariant — the one the story names as the reason AD-24's *"evaluated during bind"* matters at all — is protected by prose and by a half-blind assertion. The story's Delivery Log states the signature "is AC9's compile-time anchor"; it is not an anchor in D-000.68's sense, and the sentence should not stand unqualified. No live breach exists today (I confirmed no page-derived state can reach `computeVisibility` through `bands`, `data`, `params` or `fc`, and the call precedes `pageCount`'s declaration) — the defect is that nothing would catch one.
- **Suggested Resolution**: Two changes, both cheap. (i) Add a test-owned compile-time assertion in `visibility_test.go`, so widening the signature breaks a file the implementer must consciously edit:
  `var _ func([]bandWithOrigin, bind.Value, bind.Value, expr.FormatContext) (visibilityVerdicts, error) = computeVisibility`.
  (ii) Replace the `len(p.Runs) == 0` loop with a per-page count of runs whose `SourceText` is `"HEADER"` and `"FOOTER"` (exactly one of each, on every page). Then re-run M8 and confirm it reddens.
- **Related AC**: AC9

### Finding 3: AC4's replacement subject cannot redden — the guard is inert, and the AC still claims a red-proof it no longer has

- **Severity**: **Major**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/visibility_test.go:349-385` (`TestStyleFieldIsNeverDataDriven`); story text at the AC4 section
- **Observation**: D-3.5.2 correctly foresaw that its own ruling makes AC4's drafted subject unconstructible, and correctly required AC4 to be **kept** with a placeholder-free subject proving a different property. The subject that shipped proves almost nothing. Three measurements:

  1. The shipped subject renders one template against `{"customer": {"name": "Ada Lovelace", "overdue": true}}` and `…"overdue": false`, and asserts byte-identity. **The template never binds `customer.overdue` anywhere.** Byte-identity is implied by the setup, not by any property of styling.
  2. Disabling `checkStyleHasNoPlaceholders` entirely (mutation M9) leaves this test **GREEN** — only `TestParseTemplateRejectsPlaceholderInStyleField` reddens.
  3. `style.Background` has **no consumer outside `internal/template`'s parse/serialize and the new load check** (`grep -rn '\.Background' --include='*.go'` → `folio_expr_validate.go:210`, `serialize.go:329`, `parse_bands.go:497`). No style colour reaches the render path at all. The mutation AC4's own text names — *"implement style interpolation and the two renders diverge"* — cannot apply to a subject whose `background` is the literal `#FF0000` with no placeholder in it.

  AC4's own text still reads: *"A subject can express this defect, so D-000.24's 'no red-proof available' label does not apply and **must not be used**."* That sentence was true of the pre-ruling subject and is false of the shipped one.
- **Impact**: One of eleven ACs ships as an assertion that no plausible defect can redden, and the story records it as behaviourally red-provable. This is the exact shape sharp edge 3 and D-000.24 exist to prevent, arriving through the side door of a mid-story ruling that invalidated the subject. Worse than an admitted hole, because the AC's own text forbids admitting it.
- **Suggested Resolution**: Pick one and say which. Either (a) name and **run** a mutation that reddens the shipped subject — if none exists, that is the answer; or (b) relabel AC4 under D-000.24 as a forward regression guard with no constructible red-proof at HEAD, **and** strike the now-false "must not be used" sentence from the AC. Do not leave the AC asserting a red-proof the shipped subject cannot carry.
- **Related AC**: AC4 (and D-3.5.2's second guardrail, which anticipated this collision)

### Finding 4: D-3.5.2 implemented narrower than ruled — `table.altRowBackground` accepts a `{{ }}` placeholder at load, while the code comment and the spec claim full coverage

- **Severity**: **Major**
- **Category**: Convention / AC Conformance
- **Location**: `folio-go/folio_expr_validate.go:184-232` (`checkStyleHasNoPlaceholders`, and its doc comment's ruled-scope paragraph); `folio-go/folio_expr_validate.go:56-60` (the single call site, `el.Style` only); `_bmad-output/specs/spec-folio/folio-format.md:298-307`
- **Observation**: D-3.5.2's verdict is *"A placeholder in **any style string field** is a located load error."* The implementation covers the six fields of the element-level `style` object only. Measured by construction:

  ```
  table.altRowBackground with placeholder -> <nil>          (loads clean)
  columns[].align       with placeholder -> rejected, but by the pre-existing closed-set check
  ```

  `"altRowBackground": "{{if(customer.overdue, \"#FF0000\", \"#00FF00\")}}"` on a table element returns **no error** from `ParseTemplate`. `altRowBackground` is documented at `folio-format.md:225` as *"Colour for alternating rows (FR28)"* — it is the format's only colour field outside `style`, it is declarable today, and D-3.5.2's own worked example (*"An author writes a colour as 'whatever `if(overdue, red, black)` says'"*) describes it exactly.

  The overclaim is the sharper half. `checkStyleHasNoPlaceholders`' doc comment states: *"every style field a document can declare **TODAY** is checked (align, background, fontFamily, valign, border.color, border.edges)"*. That is false. And the spec paragraph sits directly beneath *"Colours are `#RRGGBB`. There is no colour-by-data: conditional visibility is in scope, conditional formatting is not"*, so a reader arrives at the new Story-3.5 sentence already primed to read it as covering colours. D-3.5.2's **first** named guardrail is D-000.24: *"a check that reads broader than it is, is worse than an admitted hole."* The check reads broader than it is.

  (`columns[].align` is incidentally refused, but by `template: field align (element e2): not one of the closed set left, center, right` — a message that says nothing about conditional formatting and would not survive `align` ever becoming open-ended.)
- **Impact**: The ruled fence has a hole in the one field most likely to attract a conditional colour, and both the code and the shipped spec assert the fence is complete. A future reader — including Epic 4/5, which is when `altRowBackground` acquires a consumer — will reasonably conclude the population is closed.
- **Suggested Resolution**: Extend the check to `table.altRowBackground`, having first re-derived the population from `template.Style`, `template.Border`, `template.TableExt` and `template.Column` rather than from the existing enumeration (D-000.67: a presence precondition is itself population-keyed). If the finisher judges the extension out of scope, then the code comment and the spec sentence must both be corrected to name the exact six fields and to state that `altRowBackground` is **not** covered — silence is not available here, because the ruling's own guardrail forbids it.
- **Related AC**: AC4 / DECISION-1 / D-3.5.2

### Finding 5: AC7's third subject has no discriminating mutation, none was run, and it is not labelled as a regression guard

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/visibility_test.go:461-497` (`TestHiddenElementStillFailsItsOwnValidation`, subtest `table_bind_does_not_resolve_to_an_array`); story text at AC7
- **Observation**: AC7 says *"the discriminating mutation, which the developer must run and record: move the visibility skip **above** the validation calls in **each of the three paths**. **Each subject must redden.** An AC7 that does not redden under that mutation has not been implemented — it has been placed somewhere harmless."* Under both M1 (text) and M2 (image), the table subtest **passed**. It cannot redden under any mutation of this story's code: `checkTableBindings` runs at `render.go:1136`, sixteen lines **before** `computeVisibility` at `:1152`, and no visibility verdict is ever consulted for a table element anywhere. There is no third path.

  The Delivery Log is honest — Task 8 records exactly two AC7 mutations, text and image. It is AC7's own text that overstates, and the AC is credited with three red-proved subjects when it has two.
- **Impact**: A subject presented as red-proved is a regression guard. Small in itself; it matters because AC7 is the AC whose text most loudly demands proof, and an unlabelled soft subject inside it erodes the signal the rest of the AC earns.
- **Suggested Resolution**: Label subject (c) in AC7's text as a regression guard with no constructible red-proof at HEAD (D-000.24), noting *why* — table bind validation runs before visibility is computed and tables never consult a verdict — or move it out of AC7 into a plain regression note. Keep the test either way; it is worth having.
- **Related AC**: AC7

### Finding 6: `visibleIf` verdicts are computed for `table`, `line` and `rect` elements and consumed by nothing, with no guard for Epic 4

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/render_visibility.go:75-96` (`computeVisibility` records a verdict for every element of all five kinds); consumers at `folio-go/render.go:679-693` (text only), `folio-go/page_number.go:84-91` and `folio-go/render.go:1518-1526` (image only)
- **Observation**: `template.ElementType` is a closed set of five (`model.go:177-181`). `computeVisibility` walks every element and records a verdict for each; only `ElementText` and `ElementImage` verdicts are ever read. Measured: a `line` element and a `rect` element carrying `visibleIf: "c.f"` render identically under `flag=true` and `flag=false` — no observable defect today, because neither kind reaches the page model at all yet, and a table only reaches `checkTableBindings`.
- **Impact**: Latent. The verdict map is already correct and already populated for all five kinds, so Epic 4's table/line/rect placement work will find the data waiting — but nothing forces it to consult it, nothing at the consumption sites names the kinds still unwired, and no test would fail if Epic 4 shipped table rendering that ignored `visibleIf` entirely. AD-24 says visibility applies to elements, all five kinds of them.
- **Suggested Resolution**: A test pinning the set of element kinds whose verdicts are actually consumed, written so that adding a kind to the render path without wiring visibility fails it — or, at minimum, a comment at each of the three consumption sites naming `table`/`line`/`rect` as deliberately unconsumed-because-unrendered, so the next author reads it where they will be working.
- **Related AC**: AC1, AD-24

### Finding 7: AC7's image subject asserts on the substring `"asset"`, which cannot discriminate the error AC7 names

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/visibility_test.go:469` (`errContains: "asset"`)
- **Observation**: AC7 requires each subject to produce *"its **existing** load/render error, **unchanged in text and in location**"*. The assertion is `strings.Contains(err.Error(), "asset")`. The word `asset` appears in several unrelated render errors on the same path — `folio: Render: asset %q: %w` from `DecodeAssetBytes`, and `DecodeImageForRender`'s own text — so the assertion would be satisfied by a *different* asset failure than the one the subject constructs. The subtest does redden under M2, so it is not vacuous; it is under-specified.
- **Impact**: If the fix for Finding 1 reorders or re-locates the asset-resolution pass, this subtest can keep passing on the wrong error. Given that Finding 1's fix touches exactly this code, that matters now rather than hypothetically.
- **Suggested Resolution**: Assert the distinctive substring `is not present in the document's assets map`, and add the element id.
- **Related AC**: AC7

### Finding 8: The plain-terms opener does not mention the new load-time style rejection, which arrived after the opener was written

- **Severity**: Minor
- **Category**: Maintainability (story artifact)
- **Location**: story section *"In plain terms (read this first if you just want the gist)"*, the *"Two fences are held here"* and *"Done looks like this"* paragraphs
- **Observation**: The opener meets its main obligation well — it explains, in prose and with no identifiers, why hiding leaves a gap rather than closing up, and gives the actual reason rather than restating the rule. But it was written at creation, before D-3.5.2 was ruled, and it still describes the styling fence as a thing the engine simply does not do (*"Conditions turn components on and off; they never change how a component looks"*). It never says that an author who tries anyway now gets a **refusal when the document loads**. The column fence, by contrast, is spelled out (*"It is refused from now on, when the document loads, naming the column"*), and *"Done looks like this"* enumerates the column rejection but not the style one.
- **Impact**: The one user-visible breaking change this story introduces beyond visibility itself — a template that loaded yesterday and fails to load today — is absent from the section whose whole purpose is to tell a non-implementer what changed.
- **Suggested Resolution**: One sentence in the *"Two fences"* paragraph and one clause in *"Done looks like this"*, in the same register as the column-fence sentences already there.
- **Related AC**: AC4 / DECISION-1

### Finding 9: AC8's "removing an element from the middle must not perturb the rest" is not exercised

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/visibility_test.go:499-551` (`TestHiddenElementEmitsNoDiagnosticOfItsOwn`)
- **Observation**: AC8 states *"Ordering is document order — header, content, footer, then declaration order within a band … and **removing an element from the middle must not perturb the rest**."* The subject has exactly two elements, both in the content band, and hides the **first**. There is no middle to remove, and no cross-band case: `Diagnostics` ordering across the header/content/footer concatenation is never asserted with anything hidden.
- **Impact**: The clause AC8 singles out as the interesting one is the clause not tested. The mechanism (a boolean gate over an append in a loop that already walks in order) makes an ordering defect unlikely, but "unlikely" is what the AC's own wording was written to stop being sufficient.
- **Suggested Resolution**: Add a third clipping element so the hidden one is genuinely in the middle, and put one clipping element in the page header and one in the page footer, then assert the whole `Diagnostics` slice — element ids and codes, in order — against a literal the test owns.
- **Related AC**: AC8

### Finding 10: `bind.EvaluateCondition` is a second condition slot that does not route through the hoisted predicate

- **Severity**: Minor
- **Category**: Convention
- **Location**: `folio-go/internal/bind/condition.go:39-51`
- **Observation**: D-3.5.1's own tripwire is stated as: *"How we'd know it was wrong. A third condition slot appearing that does not route through the hoisted predicate."* `EvaluateCondition` is exactly a condition slot — it exists to evaluate a bare condition expression — and it runs `expr.Parse` + `expr.Check` but **not** `expr.IsLiteralExpr`. D-3.5.1 is enforced only by `checkVisibleIfExpression`, at load. Unreachable through the public API today, because every `*Template` comes from `ParseTemplate` and a literal `visibleIf` never survives it — I could not construct a bypass.
- **Impact**: Latent, and precisely the shape the ruling asked to be watched for. If a future caller (Story 3.7's `folio.Validate`, a designer-side evaluator, a programmatically-built document) reaches `EvaluateCondition` without going through `checkVisibleIfExpression`, the literal rejection silently stops applying and a guaranteed render-time failure loads clean again — the exact defect D-3.5.1 was written to close.
- **Suggested Resolution**: Either call `expr.IsLiteralExpr` inside `EvaluateCondition` (cheap; it is one type switch on an already-parsed AST), or add a doc-comment line naming `checkVisibleIfExpression` as the sole enforcement point and stating that any second caller must apply the predicate itself. The comment is acceptable; the current silence is not, given the ruling names this scenario by hand.
- **Related AC**: AC6 / D-3.5.1

### Finding 11: The decision-log edit is missing from the File List

- **Severity**: Nit
- **Category**: Maintainability (story artifact)
- **Location**: story *File List*; `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` (+87 lines)
- **Observation**: The working tree carries an 87-line addition to the decision log — D-3.5.1, D-3.5.2, D-3.6.1 (advance notice) and a self-reported correction to §3 of the session-4 grounding refresh. Every other modified or added file in `git status` appears in the File List; this one does not. Diffing the declared list against the actual file set is the only way to notice it, and an undisclosed artifact carries claims nobody checked against the diff.
- **Impact**: Minimal in substance — I read the three rulings and they match what shipped, apart from the scope gap in Finding 4 — but the File List is what the finisher's commit is built from.
- **Suggested Resolution**: Add the decision log to the File List.
- **Related AC**: —

### Finding 12: `EvaluateCondition`'s caveat slice is discarded without comment

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/render_visibility.go:86` — `val, _, everr := bind.EvaluateCondition(...)`
- **Observation**: The `[]expr.Caveat` return is dropped. It is currently unreachable: the only caveat producer is `avg()`-on-empty, whose value is a number, so any expression that could emit one is a `ConditionValue` "must be a boolean" error before the caveat could matter — verified against the eight-function table. The discard nonetheless reads as a silent loss next to `collectBandTextRuns`, which turns caveats into diagnostics ten lines away.
- **Impact**: None today. A reader has to re-derive the reachability argument to know it is safe.
- **Suggested Resolution**: Name the reason in one clause on that line, so the next reader does not have to redo the analysis — and so that if a boolean-returning caveat producer is ever added, the comment is the thing that turns out to be wrong.
- **Related AC**: AC8

---

## Finding Resolutions (finisher pass)

Every finding below was triaged FIX / DISMISS / DEFER. **All twelve are FIX.** AC9's and AC4's
rulings arrived from the engineering lead as D-3.5.3/D-3.5.4 (recorded in
`folio-mvp-decision-log.md`); the `altRowBackground` enumeration ruling is D-3.5.5; the matrix
override question is D-3.5.6. Every other finding was judged directly against its own suggested
resolution.

### Finding 1 (Blocker) — hidden image byte-embedded in the PDF

**Decision: FIX.** Legitimate AC2 failure, unconditional clause, measured by construction. Per the
reviewer's own instruction, `collectImageRuns`/`imageRuns` stays unfiltered (AC7(b) depends on
it — re-verified, not re-litigated). Fixed at the one place the reviewer named: `pdfImages` (the
embedded-XObject map) is now built from asset keys reachable from **visible** image runs only,
while every asset key (visible or not) is still validated exactly as before.
**Files**: `folio-go/render.go` (`buildPageModel`'s `pdfImages`/`assetKeys` construction),
`folio-go/visibility_test.go` (new `TestVisibleIfHiddenImageIsByteIdenticalToElementDeleted`).
**Red-proof**: reverting the visible-only filter (temporarily marking every asset key visible)
reproduces the exact divergence the reviewer measured (hidden render carries `/Subtype /Image`;
element-deleted does not); with the fix, byte-identical. Re-verified after the fix.
**Golden-fixture check (blocker obligation 2), measured not assumed**: `git show
ae9e551:folio-go/render_test.go | grep -c "visibleIf"` → **0**. At baseline, exactly one image
element existed anywhere in the module's own template literals (`image-embed`'s
`imageTestTemplateJSON`) and it never carried `visibleIf`. **No existing golden fixture contains a
hidden element; none moved.** This is a measured zero, not an expectation (D-3.5.6's own
distinction) — no golden digest needed re-recording, and there was nothing to bring to the user
before committing.
**Cross-target artifact (D-3.5.6):** this story registers a new `matrixDocuments` entry,
`hidden-image` (`folio-go/matrix_test.go`, `folio-go/render_test.go`), because the fix moves
render-path bytes for a document class no existing golden covers (AD-21). Per D-000.54 this is a
**sequencing** step, not a D-000.4 override: its **native leg (host, darwin/arm64) was run by this
story** — `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run
TestTargetRenderHash .` — and passes (`fixtures/hidden-image/expected.json` records the measured
SHA-256). The other three targets' legs are **deferred** to the Epic 3 boundary gate (due after
3.7), on the same integer/set-work reasoning D-000.4 already accepted for three-band-page/
multi-page/page-count-20. `.github/workflows/matrix.yml` and the two mechanical obligation guards
(`byte_neutrality_test.go`'s `declaredEpic2GateObligations`, `matrix_registration_test.go`) are
updated to keep `hidden-image` in sync everywhere the other nine documents are.

### Finding 2 (Major) — AC9's anchor was a comment; behavioural half blind

**Decision: FIX**, per the engineering lead's ruling (D-3.5.3). See AC9's updated text above.
**Files**: `lint/internal/rules/visibilitysignature.go` (new), `lint/internal/rules/
visibilitysignature_test.go` (new), `folio-go/visibility_test.go`
(`TestVisibilityIsConsistentAcrossPages`'s visible-direction assertion, count → set).

### Finding 3 (Major) — AC4's replacement subject was vacuous

**Decision: FIX**, per the engineering lead's ruling (D-3.5.4). See AC4's updated text above.
**Files**: `folio-go/visibility_test.go` (`TestStyleFieldIsNeverDataDriven`'s doc comment
corrected to name `fontFamily` as the subject and state the measured population; the test's own
template/assertions were already shaped correctly and needed no functional change — the defect
was in the doc comment's claim, not the code). The obligation for the eight inert fields is
recorded against Stories 4.1 and 4.8 above (AC4 section); the engineering lead is carrying the
same referent into their briefs.

### Finding 4 (Major) — `table.altRowBackground` unchecked; comment overclaimed coverage

**Decision: FIX**, per the engineering lead's ruling (D-3.5.5): enumerate from the schema, not a
hand list — and this is the OPPOSITE instrument from Finding 2/AC9's, because the style-field set
is scheduled to grow (Epic 4 tables, 4.8) where AC9's set is permanent. Implemented as: (a) the
production check now also covers `table.altRowBackground`, via the same
`checkStyleStringHasNoPlaceholder` mechanism `style.*` already uses (no second copy — D-000.38);
(b) a reflection-based completeness test, `TestStyleStringFieldPopulationMatchesSchema`
(`folio_expr_validate_test.go`), walks `template.Style`/`Border`/`TableExt`/`Column` for every
`Presence[string]`/`Presence[[]string]` field and asserts the covered set equals the schema-derived
set minus a small, individually-reasoned exclusion list (`styleStringFieldExclusions` — five
fields that are structural/closed-enum, not appearance properties: `TableExt.As`, `Column.Align`,
`Column.Footer`, `Column.FooterOf`, `Column.FooterFormat`). A future eleventh
`Presence[string]` field on any of the four structs is neither checked nor excluded by name and
fails this test loudly.
**Files**: `folio-go/folio_expr_validate.go` (`checkStyleStringHasNoPlaceholder` factored out;
table.altRowBackground call added in `validateAndDeriveExpressions`; doc comment corrected),
`folio-go/folio_expr_validate_test.go` (two behavioural tests + the reflective completeness test).
**Red-proof**: (i) behavioural — `TestParseTemplateRejectsPlaceholderInAltRowBackground`, run
against the pre-fix tree, fails (loads clean); with the fix, rejects, naming the element and
field. (ii) completeness test — temporarily adding an uncovered `Presence[string]` field to
`template.Style` reddens `TestStyleStringFieldPopulationMatchesSchema` naming the field; reverted.

### Finding 5 (Minor) — AC7 subject (c) has no discriminating mutation

**Decision: FIX.** Labelled as a regression guard with no constructible red-proof at HEAD
(D-000.24), in place, with the reason (`checkTableBindings` runs before `computeVisibility` even
exists in the pipeline, and no table element ever consults a verdict). **File**:
`folio-go/visibility_test.go` (comment on the "table bind does not resolve to an array" subtest).

### Finding 6 (Minor) — table/line/rect verdicts computed, consumed by nothing

**Decision: FIX**, the cheaper of the two suggested options (a comment at each consumption site
naming the unwired kinds), given this is Minor and the property is genuinely latent rather than
broken. **Files**: `folio-go/render_visibility.go` (`computeVisibility`'s doc comment),
`folio-go/page_number.go` (`contentColumnItems`'s doc comment).

### Finding 7 (Minor) — AC7 image subject's `errContains: "asset"` under-specifies

**Decision: FIX.** Strengthened to the distinctive substring, and the production error message
itself was improved to name the element id (it did not before, even though
`firstElementIDByAssetKey` was already computed at that point) — directly relevant because
Finding 1's fix touches this exact code block. **Files**: `folio-go/render.go` (the "not present
in the document's assets map" error now names the element), `folio-go/visibility_test.go`
(assertion strengthened to match).

### Finding 8 (Minor) — plain-terms opener omits the new load-time style rejection

**Decision: FIX**, folded into the opener refresh below (see "In plain terms", rewritten for this
finish pass as required regardless of this finding).

### Finding 9 (Minor) — AC8's "removed from the middle" clause unexercised

**Decision: FIX.** Rebuilt `TestHiddenElementEmitsNoDiagnosticOfItsOwn` with one clipping element
in each band (header `e4`, footer `e5`) and three in content (`e1`, `e2` — hidden, `e3`), so the
hidden element is genuinely in the middle of its own band; the assertion now compares the WHOLE
`Diagnostics` slice (element ids, in document order) against a literal the test owns, both
directions, plus per-diagnostic byte-identity for every surviving element. **File**:
`folio-go/visibility_test.go`.

### Finding 10 (Minor) — `bind.EvaluateCondition` bypasses the hoisted literal predicate

**Decision: FIX**, the cheaper of the two options declined in favour of the safer one:
`EvaluateCondition` now calls `expr.IsLiteralExpr` itself (one type switch on an already-parsed
AST), so a future second caller reaching it directly cannot silently bypass D-3.5.1. **Files**:
`folio-go/internal/bind/condition.go`, `folio-go/internal/bind/condition_test.go` (new — direct
unit coverage, since the path is unreachable through the public API today).

### Finding 11 (Nit) — decision-log edit missing from File List

**Decision: FIX.** Added below.

### Finding 12 (Nit) — `EvaluateCondition`'s caveat slice discarded without comment

**Decision: FIX.** One-clause comment added naming the reachability argument. **File**:
`folio-go/render_visibility.go`.

---

### AC-by-AC disposition (finisher pass — supersedes the review's column above)

| AC | Verdict |
|---|---|
| AC1 | **Satisfied.** Text and image halves both verified; M7 confirms the red-proof. |
| AC2 | **Satisfied, both element kinds.** Text half as reviewed; image half fixed (Finding 1 / Blocker) — `pdfImages` now built from visible-only asset keys, red-proved. |
| AC3 | **Satisfied.** M6 red-proof re-run; spec/code divergence closed. |
| AC4 | **Satisfied, subject corrected to `fontFamily`** (Finding 3 / Major, D-3.5.4). Population of eight unprovable fields stated explicitly and handed to Stories 4.1/4.8. |
| AC5 | **Satisfied.** All six condition kinds probed; M3 and M4 red-proofs re-run. |
| AC6 | **Satisfied.** M5 proves one predicate serves both slots. |
| AC7 | **Satisfied for all three subjects.** Subject (c) now labelled a regression guard with no constructible red-proof (Finding 5). Assertion strength fixed (Finding 7), production error now names the element. |
| AC8 | **Satisfied**, instrument completed: the "removed from the middle," cross-band case is now exercised (Finding 9). |
| AC9 | **Satisfied — real instrument.** `lint`'s `ScanVisibilityComputationSignature` pins the parameter type list by set-equality (Finding 2 / Major, D-3.5.3); the behavioural half now asserts a per-page set, not a count, and reddens under the same mutation that left it green before. |
| AC10 | **Satisfied.** Both instruments green; `EvaluateCondition` correctly a function on `bind`. |
| AC11 | **Satisfied.** See the finisher's own gate measurement in the Delivery Log — 853/1/1 (folio-go), 104/3 (lint), 3/0 (hashmatrix); no fourth failure in either module; P6 stats byte-identical. |

### Could not verify (review-time) — resolved by the finisher

- **The Epic 3 cross-target boundary matrix** is still not due until after 3.7 (D-000.4) — unchanged. What the review flagged as missing (the render-path note not mentioning Finding 1's byte movement) is now addressed: this story registers a new `matrixDocuments` entry, `hidden-image`, runs its native (host) leg per D-000.54 (sequencing, not a D-000.4 override — ruled D-3.5.6), and defers the other three targets' legs to the Epic 3 gate exactly as three-band-page/multi-page/page-count-20 already do. The golden-fixture question the review raised ("does any existing golden move?") is answered by measurement, not assumption: **N = 0** — see Finding 1's resolution above for the exact command and result.
- **`.matrix-build/`** was excluded from every scan, per the story's own instruction; it contains stale binaries carrying the string `VisibleIf`.

