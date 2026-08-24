---
baseline_commit: 213d98c3c61c0e20522e15c3a9a87bb5f278cdbe
---

# Story 2.5: Compose a page from three bands

**Epic:** 2 — Text, shaping, breaking and page composition
**Story key:** `2-5-compose-a-page-from-three-bands`
**Status:** `done`
**Covers:** FR6 · AD-5, AD-24 (`epics.md:846`–`:878`)
**Primary invariant:** **AD-5** (the page model knows nothing about PDF) — *its dependency arrow
becomes live for the first time in this story, and has never existed and never been guarded before
it* (D-2.0.4).
**Adjacent invariants:** AD-24 (boxes are absolute, and nothing negotiates) · AD-13 (derived
geometry, one function) · AD-4 (two passes, and the second one lays nothing out) · AD-2 (one
fixed-point unit) · AD-21 / AD-22 (every feature ships its golden; a hash change is a versioned
event) · AD-23 (no `float64`/`float32`) · AD-1 (import lints) · AD-10 (element ids)
**Governing rulings:** **D-2.0.4 (the spine of this story)** · **D-000.16 (the stage-rank import
guard — this is the story it lands in)** · **D-000.33** · **D-000.36** · **D-000.35** ·
**D-000.22** · **D-000.28** · **D-000.21 (sharpened)** · **D-000.23** · **D-000.24** ·
**D-000.26** · **D-000.29** · **D-000.30** · **D-000.32** · **D-000.34** · D-000.9 · D-000.4 ·
D-000.6 · D-000.17 · D-1.3.3 · D-1.5.5 · D-1.8.10 · D-1.7.2 · D-2.3.5 · D-2.4.2 · D-2.4.3 ·
D-2.4.6
**Deferred work touched:** none opened by scope, **two recorded**. **DW-15** (the baseline offset —
see *DECISIONS NEEDED* D-2.5-Q3) was proposed by this story and has since been recorded by the owner;
it is deliberately NOT landed here. **DW-16** was opened by the finisher from the reviewer's Finding 2
(`pagemodel.ShapedGlyph.CID` is not always a glyph id). No existing DW entry is owned or discharged
here.

---

## Baseline, measured at creation

HEAD is **`5968c5d`** — *"Story 2.4: Break and measure lines in all three scripts (finisher)"* — on
`main`.

**The working tree is NOT clean at this baseline, and that is reported rather than corrected.**
`git status --porcelain` reports one modified file:

```
 M _bmad-output/implementation-artifacts/folio-mvp-decision-log.md
```

It is a **44-line append** adding **`D-000.36`** ("A proposed remedy for a vacuity is itself subject
to vacuity") at the end of the log. Story 2.4's finisher committed the story but not this entry.
**This story neither reverts it nor commits it** — it is a decision-log entry, it is correct, this
story *cites it*, and quietly reverting a ruling would be the worst available outcome. Flagged for
the epic gate. **No `folio-go/` or `lint/` source file is dirty**, so every measurement below is
taken against `5968c5d`'s code exactly.

Every number is stated with its scope, its invocation and its counting rule (D-000.26):

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting **every** `--- PASS` / `--- FAIL` occurrence, subtests included | **464 PASS · 1 FAIL** |
| `folio-go/` | the same invocation, counting only **top-level** results (`^--- PASS` at column 0) | **293 PASS · 1 FAIL** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 ./...` | **4 packages `ok`, 0 FAIL** (81 PASS · 0 FAIL all-occurrences; 43 · 0 top-level, carried from Story 2.4 and unchanged — no `lint/` file moved in `5968c5d`) |
| `folio-go/` | `GOWORK=off go list -m all` | exactly **two** modules: `github.com/panitw/folio/folio-go`, `github.com/boxesandglue/textshape v0.0.15` |

**The single failure is `internal/text`'s `TestCorpusMeetsP6ExerciseFloors`** — *"P6g (opaque names)
floor not met: got 7, need >=20"*, with `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115
P6g:7}`. It is **Story 2.1's disclosed shortfall, held open by Story 2.4's AC5, and it must stay
red.** Do not close it, do not skip it, do not add corpus items to fill it (D-000.17: *a floor that
is not met is reported as unmet; it is never filled*). **A second failure is a regression.**

---

## In plain terms (read this first if you just want the gist)

A Folio report is built from three stacked strips of the page: a header at the top, a footer at the
bottom, and everything else between. Which strip something sits in decides whether it repeats on
every page later. Authors position things relative to their strip, not to the paper, so moving a
strip moves its contents and nothing is re-measured.

Since the font work the engine had done a rough version of this, as an admitted placeholder living
inside the code that writes the PDF. That placeholder is now the real thing, and two tangled jobs are
separated: deciding where everything goes, and writing a PDF. The deciding half now produces a plain
description of a finished page and knows nothing about PDF. A foundational rule forbidding that
separation from being crossed had never been checkable, because the piece it protects did not exist.
It exists now, and the guard shipped with it.

Delivered: a new reference document with content in all three strips, all six existing ones
byte-for-byte unchanged, and a long-standing test deleted because what it waited for arrived. The
strip measurements are now pinned by a check that previously could not tell a correct layout from a
badly wrong one — a deliberately wrong value used to pass everything, and now fails four tests.

Three things look wrong and are not. **One test fails on purpose**: the Thai corpus lacks enough
deliberately-confusing names, a gap reported rather than quietly filled. **The first line of text
sits inconsistently** with the spacing below it; correcting it would change every recorded reference
file, including one a human has already been asked to read, so it lands before that reading, not
here. **A glyph-labelling shortcut was recorded, not fixed**: in one uncommon case a
glyph is numbered in a way only a PDF writer can decode, narrowing the promise that this page
description is renderer-neutral. It dates from the shaping work and only became visible when the code
moved, so nothing broke — but it is written down, with its two possible fixes, before more code
depends on it.

Strips still never resize to fit their contents; overflow is visible, and handling it is a later
story.

---

## Story

As a template author,
I want my report composed of a Page Header, a Content region and a Page Footer,
So that which band I place something in decides whether it repeats on every page.

---

## Do not re-open — settled rulings this story inherits

Reproduced here with their rationale so nothing below has to be looked up. **If anything in this
file looks like it contradicts a decision-log entry, stop and surface it — do not resolve it by
choosing.**

1. **AD-5's arrow must be guarded in the same commit that creates `internal/layout`** (D-2.0.4,
   binding). *"The guard must exist from the package's first commit, not be added later — the window
   in which the arrow is easy to add and invisible is exactly the window before the guard exists."*
   Retained violating fixture, two-caller shape per D-1.3.3.

2. **The guard is a stage-rank table, not a single arrow** (D-000.16, binding for the rule,
   illustrative for the ranks). Each `internal/` package gets a rank; a package may import only
   **lower** ranks. This subsumes AD-5's arrow *and* D-1.6.1's `expr`↛`bind` pre-commitment, and
   pre-forbids the arrows nobody has thought of yet. **The ranks as published are illustrative and
   were explicitly left "for the implementing story to validate" — see Finding 1: they are wrong,
   and this story corrects them.**

3. **The signal rides on the value, never through an import** (D-000.16). *"Stages communicate by
   what they pass, not by what they import."* `internal/layout` receives what it needs as
   parameters.

4. **AD-24, verbatim, and it is directly load-bearing here.** *"An element's x/y is relative to its
   band's top-left corner, never to the page. Bands are placed on the page by `internal/layout`
   alone."* And: *"`PageModel` is top-left origin with Y increasing downward… The flip to PDF user
   space happens in exactly one function in `internal/pdf`, and nowhere else in the module inverts a
   coordinate."*

5. **Band placement is a translation, not an inversion** (D-2.0.4, illustrative). *"It must not add a
   second flip. Story 1.8's positive guard should catch it; 2.5 is its first real test against a
   genuinely different caller."*

6. **The content band's height is derived, by one function** (AD-13's sibling rule;
   `folio-format.md:94`–`:96`; D-2.0.4: *"Guard it positively at 2.5"*). Storing it would be a second
   source of truth. It is derived from **page geometry alone** — never from what the band contains
   (AD-24: nothing negotiates).

7. **Leading is `hhea` ascent − descent (+ lineGap), taken as the MAXIMUM over the declared chain,
   read from the `hhea` table** — never the vendor accessors, which substitute 800/−200/0
   (D-2.4.2 resolved, D-2.4.6). *"A chain declares what may appear in an element. Leading must
   accommodate what may appear — not what does appear."* **Do not re-open this. Do not introduce a
   `lineHeight` key.**

8. **Unbreakable fields are a document-level list of bare root-relative dotted data paths**
   (D-2.4.1), spelled as `columns[].footerOf` is. Not element-level, not binding syntax.

9. **Shaping and measurement share one path.** The breaker measures and slices the *same* glyphs
   that are drawn (`shapeSegments` + `positionSegments`, `folio-go/render.go:457`, `:522`).
   **Do not reintroduce a second advance path** — two paths disagreeing was Story 2.3's blocker.

10. **R4 stays shut** (D-2.0.4). Epic 1 emits `/FlateDecode` as a *filter name* on the image
    passthrough but never invokes a compressor. Nothing in this story requires compression.
    Introducing content-stream compression is a hash-changing, versioned event under AD-22 and
    **must not happen incidentally** — if it is wanted, that is a `DECISION NEEDED`.

11. **`{{page}}` / `{{pages}}` stay reserved and untouched until Story 2.7** (D-2.0.4). No `page`
    namespace, no late-bound slot type, nothing pagination-shaped is created here.

12. **A floor that is not met is reported as unmet. It is never filled** (D-000.17).

13. **The Epic 2 gate owes exactly three things** (Story 2.3a's gate note, verbatim): the outstanding
    four-target matrix legs (2.3 / 2.3a), **D-2.3.5's Thai *reading* sign-off**
    (`fixtures/shaped-text/thai-signoff.json`, bound to digest `5964aad0…c92e00f`, the sha256 of
    `expected.pdf`), and **D-2.4.3's Thai *break* sign-off** with its own digest.
    `TestStory23aAddedNoThirdEpic2GateObligation` catches a fourth. **See D-2.5-Q1: this story
    proposes one addition, explicitly and with the owner's ruling requested — it is not taken
    silently.**

---

## Corrections to inherited claims — verified at `5968c5d`

Every row was re-checked against the code at HEAD. **Three of the epic's five acceptance criteria
assert things that are already true**, and one live test's failure message contains a stale
instruction.

| # | Claim, and where it comes from | Verified status at `5968c5d` |
|---|---|---|
| C1 | `epics.md:856`–`:859`: *"it has exactly three bands… group headers, report headers and group footers are absent and unsupported"* | **ALREADY TRUE since Story 1.4.** `folio-go/internal/template/parse_bands.go:15` `decodeBands` requires exactly `content`, `pageFooter`, `pageHeader`, and a fourth key or a missing one is a **load error, never passthrough** (`:33`–`:35`). `internal/template/closedsets_test.go` and `fixtures_test.go` cover it. **This is a regression guard in this story, never evidence that it landed.** |
| C2 | `epics.md:862`: *"x and y are relative to that band's top-left corner, never to the page"* | **ALREADY TRUE in the document model.** `parse_bands.go:100`–`:120` decodes `x`/`y` per element with no page-level adjustment; `folio-format.md:106` states it. **The half that is NOT true is `epics.md:863`, *"bands are placed on the page by `internal/layout` alone"* — `internal/layout` does not exist.** Band placement lives in `folio-go/render.go:100` `documentBands`, in package `folio`, labelled `PROVISIONAL (AC28)` at `render.go:89`, `:223`. That is the real work. |
| C3 | `epics.md:871`: *"the flip… occurs in exactly one function in `internal/pdf` and nowhere else"* | **ALREADY TRUE and already guarded**, by name: `internal/pdf/flip.go:29` `flipY`, guarded by `internal/pdf/flip_test.go:173` `TestContentStreamYCoordinatesRouteThroughFlipY` with its own red-proof `TestFlipRoutingRedProof` (`:212`) covering three bypass shapes. Ratified as AD-24's "exactly one function" by **D-1.8.10**. **Regression guard; not new work.** (D-000.35: symbols named.) |
| C4 | `epics.md:869`: the page model *"names only geometry, glyph runs, images and vector primitives"* | **Nothing named `PageModel` exists.** The nearest thing is `internal/pdf`'s own `TextPage` / `TextRun` / `ImagePlacement` (`internal/pdf/textdoc.go:68`, `:149`), which live in the **wrong package** for AD-5. Also: **the epic omits `internal/pagemodel`, which AD-5 and AD-24 both name as the package that owns these types.** The spine is more specific than the epic; the spine governs. |
| C5 | `epics.md:877`: the content area *"is derived by one function as page height minus margins minus header height minus footer height"* | **The content band's height is not computed anywhere in the repository.** Measured: `usableHeight` (`render.go:105`) is `height − marginTop − marginBottom` and is used at exactly one site, the **footer** origin (`render.go:118`). `grep -rn "usableHeight\|contentHeight\|ContentHeight"` over `folio-go/` returns three hits, all of them those two lines plus one comment in `render_test.go:1090`. **This AC is genuinely new work.** |
| C6 | `render_test.go:1094` `TestProvisionalBandOriginIsPinned`'s failure message instructs that **two** provisional sites *"must be replaced"*, naming site (2) as `internal/pdf/textdoc.go`'s *"PROVISIONAL Y-flip into PDF's bottom-up user space"* | **Half stale.** The *flip mechanism* is no longer provisional — D-1.8.10 ratified `flipY` as AD-24's permanent one-and-only inverter, after that message was written. What remains genuinely provisional in `textdoc.go` is (a) the **stale comment** at `:69`–`:77` claiming `TextRun.X/Y` will become *band-relative* — they must not; under AD-24 layout resolves band origins so what reaches `internal/pdf` is **page-absolute**, and making `TextRun` band-relative would put band placement inside `internal/pdf`, violating AD-24 outright — and (b) the **baseline offset** `pdfY = flipY(page.Height, page.MarginTop, run.Y, run.FontSize)` at `:730`, which places the baseline `fontSize` below the run top. **(b) is a hash-changing question and is explicitly NOT in this story's scope — see D-2.5-Q3.** |
| C7 | D-000.16's ranking: `geom` 0 · `diag`/`pagemodel` 1 · `template` 2 · `expr` 3 · `bind` 4 · **`fontset` 5 · `text` 6** · `layout` 7 · `pdf` 8 | **Falsified by the current graph — see Finding 1.** `internal/fontset` imports `internal/text`. Applying the published ranks verbatim makes the new guard **red on its first commit against correct, shipped code.** D-000.16 anticipated this (*"for the implementing story to validate"*). |

---

## Scope fence — what this story is NOT

- **Not pagination.** One page only. Multi-page output, repeating the header/footer on every page,
  and the two-pass split under AD-4 are **Story 2.6**. Nothing here may create a page loop, a page
  break, or a "does it fit" test.
- **Not `Page X of Y`.** No `page` namespace, no late-bound slot, no `{{page}}`/`{{pages}}` —
  **Story 2.7**, ruled at D-2.0.4.
- **Not clipping or overflow diagnostics.** Content wider or taller than its box still overflows
  visibly, exactly as it does at `5968c5d`. That is **Story 2.8** (FR44, AD-14).
- **Not `visibleIf`.** AD-24's visibility clause needs an expression evaluator (`internal/expr`,
  Epic 3). `visibleIf` parses today and is ignored at render; it goes on being ignored.
- **Not vector primitives.** `line` and `rect` elements parse (`parse_bands.go:225`) and are drawn
  by nothing. AD-5's *"vector primitives"* is a statement about what the page model **may** name, not
  a requirement to declare a type with no producer and no red-proof. Table chrome (Epic 4) is what
  first produces them. **Do not pre-declare a vector-primitive type here** — writing the type before
  the thing that fills it is D-000.28's anticipatory shape, one level up.
- **Not a re-record of any existing golden.** See AC8. All six recorded digests must be
  byte-identical after this story.
- **Not a change to the leading rule, the breaking rule, the unbreakable-values rule, or the
  shaping path.** Those are Story 2.4's, ruled and welded.
- **Not compression** (D-2.0.4's R4 clause).
- **Not a new module.** `go list -m all` stays exactly two.

---

## Measured findings — read all of these before writing code

### Finding 1 — D-000.16's published ranking is falsified by the current import graph, and must be corrected before the guard is written

**Measured**, `CGO_ENABLED=0 GOWORK=off go list -f '{{.ImportPath}}:: {{join .Imports " "}}' ./...`
at `5968c5d`, first-party edges only (`@` = `github.com/panitw/folio/folio-go`):

```
@                    -> @/internal/bind @/internal/fontset @/internal/geom @/internal/pdf
                        @/internal/template @/internal/text
@/internal/bind      -> @/internal/template
@/internal/fontset   -> @/internal/geom @/internal/text        <-- THE PROBLEM
@/internal/geom      -> (none)
@/internal/pdf       -> @/internal/geom
@/internal/template  -> @/internal/geom
@/internal/text      -> (none)
@/cmd/{genbreaks,gencorpus,gentrie} -> @/internal/text
```

**`internal/fontset` imports `internal/text`.** The site is real and load-bearing, not incidental:
`internal/fontset/fontset.go:25` imports it, `:74` declares `shaper *text.Shaper`, `:245` calls
`text.NewShaper(name, face)`, and `:445` exports `func (f *Font) Shaper() *text.Shaper`. That
wrapper is **Story 2.3a's vendor boundary** — *"the returned type is folio's own
`internal/text.Shaper`… so no vendor pointer crosses this"* (`:72`, `:439`). Removing it to satisfy a
rank table would undo a ruled vendor-containment decision.

Under D-000.16's published ranks (`fontset` 5, `text` 6) that edge points **upward** and is a
violation. **The graph is correct and the table is wrong.**

**The corrected ranking this story adopts and must validate against the real graph before writing
the guard:**

| rank | package | exists at `5968c5d`? |
|---|---|---|
| 0 | `geom` | yes |
| 1 | `diag`, `pagemodel` | `pagemodel` **created here**; `diag` not yet (Epic 3) |
| 2 | `template` | yes |
| 3 | `expr` | not yet (Epic 3) |
| 4 | `bind` | yes |
| 5 | **`text`** | yes |
| 6 | **`fontset`** | yes |
| 7 | `layout` | **created here** |
| 8 | `pdf` | yes |

Only `text` and `fontset` swap. Every published edge that D-000.16 exists to forbid still falls out
by construction: `layout`(7) cannot import `pdf`(8) — **AD-5**; `expr`(3) cannot import `bind`(4) —
**D-1.6.1**. Under the corrected table the graph above is **clean and forward-only, with zero
violations**, which is D-000.16's own stated precondition (*"green today"*).

> **This is a correction to a binding ruling's illustrative half.** It is recorded as
> `D-2.5-Q2` in *DECISIONS NEEDED* so the ruling's author confirms rather than has it reinterpreted
> — but the guard cannot be written against a table that reddens on shipped code, so the corrected
> ranks are what the ACs specify. If the ruling comes back differently, the ACs change; the developer
> does not choose.

### Finding 2 — Measured by injection: **no test in the repository can detect a wrong page-header band origin**

The three band origins are set at `render.go:116`–`:118`. Each was mutated in turn, the full suite
run, and the source restored. Invocation for each row: `CGO_ENABLED=0 GOWORK=off go test -count=1
./...` from `folio-go/`.

| mutation (at `render.go`) | result | what detected it |
|---|---|---|
| `{Bands.PageHeader, 0}` → `{Bands.PageHeader, 999000}` | **package `folio` still `ok`** — suite identical to baseline, only the pre-existing `TestCorpusMeetsP6ExerciseFloors` red | **nothing. Zero tests.** |
| `{Bands.Content, headerHeight}` → `headerHeight + 1000` | 5 failures | `TestRenderMatchesFontTextGoldenFixture`, `TestRenderMatchesImageEmbedGoldenFixture`, `TestMultiScriptFallbackGoldenFixture`, `TestShapedTextGoldenFixture`, `TestWrappedTextGoldenFixture` |
| `{Bands.PageFooter, usableHeight − footerHeight}` → `+ 1000` | 1 failure | `TestRenderMatchesFontTextGoldenFixture` **only** |

**The cause, and it is the D-000.36 corollary exactly — the subject matters as much as the
assertion.** Enumerated over all six fixture `input.folio` files:

| fixture | `pageHeader` elements | `content` elements | `pageFooter` elements |
|---|---|---|---|
| `minimal-rect` | *(no `input.folio`; hard-coded document)* | — | — |
| `font-text` | **0** | 1 | 1 |
| `image-embed` | **0** | 1 | 0 |
| `multi-script-fallback` | **0** | 1 | 0 |
| `shaped-text` | **0** | 7 | 0 |
| `wrapped-text` | **0** | 4 | 0 |

**Every fixture in the repository has an empty `pageHeader`.** Six goldens, twelve content elements,
one footer element, and **not one header element anywhere.** A band-origin assertion over these
inputs is a correct assertion over subjects that cannot express the defect.

Every fixture also uses the **same** page setup — A4, `margin` 36 on all four edges,
`pageHeader.height` 20, `pageFooter.height` 20 — so header height and footer height are
**indistinguishable**, top and bottom margins are **indistinguishable**, and a swap of any pair is
invisible in the bytes. That is what AC7's fixture must fix, and it is why AC7 specifies **distinct**
values for all four.

### Finding 3 — `TestProvisionalBandOriginIsPinned` fires on `mkdir internal/layout`

`render_test.go:1094`. Its body is `if _, err := os.Stat(filepath.Join("internal", "layout")); err
== nil { t.Fatal(...) }`. **The developer's very first structural action reddens the suite**, by
design (D-1.5.5's self-retiring-assertion pattern). This is expected, is not a regression, and the
test is **deleted** as part of AC9 — not skipped, not weakened, not left behind.

Its message names two sites. Site (1), `render.go`'s `collectTextRuns` band origins, is replaced
here. Site (2)'s instruction is half stale — see correction **C6** and **D-2.5-Q3**.

### Finding 4 — the "exactly one document-byte producer" guard constrains how this refactor may be shaped

`render_arch_test.go:145` `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt`
(D-1.7.2) asserts that **exactly one** top-level function declared in a non-test file directly under
`folio-go/` calls into `internal/pdf` **through any selector on the import alias**, and that both
`Render` and `RenderTo` reach it through package-`folio` direct calls. Today that function is
`renderDocument` (`render.go:553`).

**A refactor that adds a second package-`folio` function touching `internal/pdf` — for example a
`serializePages` helper split out beside `renderDocument` — reddens this guard for the right reason
under its own rules.** Keep exactly one. Anything that wants to be a second function belongs inside
`internal/layout` (which may not import `internal/pdf` at all) or inside `internal/pdf` itself
(where the guard does not look — it is scoped to the module-root package by construction, `:11`).

### Finding 5 — the flip guard is package-scoped and will not see `internal/layout`

`internal/pdf/flip_test.go:173` scans **`internal/pdf` only** — *"no function in this package other
than `flipY` may…"*. D-2.0.4 expects band placement to be *"a translation, not an inversion"* and
calls 2.5 *"its first real test against a genuinely different caller"*. Measured: that expectation is
**not automatically discharged**, because a coordinate inversion introduced inside `internal/layout`
is outside the guard's scan root entirely. AC6 therefore asserts the band-origin arithmetic
**positively and by value** in `internal/layout`'s own tests rather than relying on the
`internal/pdf` guard to notice.

### Finding 6 — `internal/pdf` does not import `internal/template`, and must not start

Measured above: `@/internal/pdf -> @/internal/geom`, nothing else first-party. That is the shape AD-5
protects. `internal/pdf` may gain an import of `internal/pagemodel` (rank 1 < 8, permitted); it must
not gain `internal/template`(2)… which the rank table also permits, so **AC1 states it as a property
of `pagemodel`'s content instead**: the page model carries resolved geometry and glyph runs, not
`template.Element` values. A page model that embedded `template` types would satisfy every rank check
and still put document-model concepts into the serializer.

### Finding 7 — `map-range` and the numeric guards apply to the new packages from their first line

`lint/internal/rules/maprange.go:14` bans `range` over a map (determinism); `no-float64` /
`no-float-typed-value` are **type-aware** since Story 2.3a and scan the whole module root
(`internal/arch_test.go`'s `TestNoFloat64UnderModule`); `forbidden-imports` (AD-1) applies to
render-path code. `internal/layout` is render-path code. Iterate bands and elements in **declared
order over slices**, never over a map, and keep every number in `geom.Length` millipoints.

---

## DECISIONS NEEDED — escalate before development starts

**Do not begin AC3 or AC7 until D-2.5-Q1 and D-2.5-Q2 are ruled.** The remaining ACs are unblocked.

### D-2.5-Q1 — This story adds a **fourth** item to the Epic 2 gate. Confirm it.

The Epic 2 gate currently owes exactly three things (Story 2.3a's note, and
`TestStory23aAddedNoThirdEpic2GateObligation` enforces the count). AC7 records a new golden,
`fixtures/three-band-page/`, and AD-21 plus `TestMatrixDocumentSlugsAreRegisteredInCI`
(`matrix_registration_test.go:32`) require every recorded document to be registered in **both**
`matrixDocuments` (`matrix_test.go`) and `.github/workflows/matrix.yml`'s `docs="…"` line, with four
`hash.<target>.<slug>.txt` upload paths each.

**That is a fourth gate obligation and it is stated explicitly here rather than taken silently, as
the constraint requires.** It is **not** optional: a new feature without a golden violates AD-21, and
a golden registered in one list and not the other is *"a matrix leg nobody compares, reported as
green"*.

- **Option A (recommended).** Register `three-band-page` in both lists now. The gate owes four
  documents' legs instead of three. The three existing obligations are untouched;
  `fixtures/shaped-text/` and its `5964aad0…c92e00f` digest are not read, not re-recorded, and not
  moved.
- **Option B.** Record the golden but do not register it, deferring registration to the gate.
  **Recommended against** — it is precisely the Story 2.2 drift `matrix_registration_test.go` was
  written to close, and that test would go red anyway if the slug were added to only one list.

**Also confirm:** does `TestStory23aAddedNoThirdEpic2GateObligation` need updating to expect four,
and is that update this story's to make?

### D-2.5-Q2 — Confirm the corrected stage-rank table (Finding 1)

D-000.16's published ranks put `fontset` 5 and `text` 6; `internal/fontset` imports `internal/text`
at `fontset.go:25`/`:245`/`:445`, so the guard would be **red on arrival** against shipped, ruled
code (Story 2.3a's vendor containment). The proposed correction swaps exactly those two ranks —
`text` 5, `fontset` 6 — and leaves every other rank as published. Both arrows D-000.16 exists to
enforce still hold by construction. **Requesting confirmation of the corrected table, and of whether
the correction is recorded as a new `D-2.5.x` entry or as an amendment to D-000.16.**

### D-2.5-Q3 — The baseline offset: `run.FontSize` versus the ruled ascent. **Do not act on this in this story.**

`internal/pdf/textdoc.go:730` places a text baseline at `flipY(page.Height, page.MarginTop, run.Y,
run.FontSize)` — i.e. `fontSize` below the run's top, described in its own comment (`:722`) as
*"approximately"*. Meanwhile D-2.4.2's ruled leading advances line-to-line by
`max(hhea ascent − descent + lineGap)` over the declared chain (`wrap.go:323` `lineAdvance`). The
first baseline's offset and the inter-line advance are therefore derived from **different
quantities**. A typographically conventional engine would place the first baseline at the same
**ascent** the leading is built from.

**Why it is not touched here:** changing it moves the Y of **every text run in every existing
golden**, which is a re-record of `font-text`, `multi-script-fallback`, `shaped-text` and
`wrapped-text`. Under AD-21/AD-22 that is a versioned, ruled event and never a developer's judgment —
and `fixtures/shaped-text/expected.pdf` carries digest **`5964aad0…c92e00f`**, to which **D-2.3.5's
pending human Thai reading sign-off is bound**. Re-recording it would silently invalidate a sign-off
that has not happened yet.

The offset is **content-independent**, so it violates nothing in AD-24 and is not a defect this
story must close. **Proposed disposition: open `DW-15` recording the question, its blast radius (4
goldens + 1 pending sign-off), and its natural owner** — a story that is already re-recording, or the
Epic 2 gate. Requesting the owner's ruling on opening DW-15 rather than resolving the question.

### D-2.5-Q4 — Heavy-test cadence: propose a **per-story matrix override** for 2.5 (D-000.4)

D-000.4 runs the four-target matrix **per epic**, with per-story overrides for hash-shaped stories —
naming **1.2, 1.5, 1.8, 2.4** and 4.7, and stating that further overrides are *"added as the run
reveals them"*.

**The argument that 2.5 qualifies, stated so it can be refused:** 2.4 was overridden because *"line
breaking feeds every measurement"*. Band composition feeds every measurement in the same sense and
one layer further out — **every** element's page Y on **every** golden from here is `band origin +
element y`, and Finding 2 measures that 5 of 6 existing goldens already move when a band origin
moves. Beyond that, this story **relocates the entire render path across a package boundary** and
introduces the first `internal/layout` arithmetic, which is the single highest-risk moment in Epic 2
for a cross-target divergence — and a divergence introduced here would be attributed to whichever
later story first ran the matrix.

**The argument against:** 2.5 introduces no new *numeric* rule (no new rounding, no new unit
conversion — the arithmetic is a relocation of `render.go:105`–`:118`), and AC8 requires all six
existing digests to be byte-identical, which is itself a strong local proof that nothing moved.

**Recommendation: override, restricted to the one new document `three-band-page` plus a re-run of
`wrapped-text`** (the most arithmetic-dense existing golden) — not all six. **Requesting a ruling;
this is not assumed either way.**

---

## Acceptance Criteria

Numbering is this story's own. **Every assertion group below states its presence precondition
(D-000.21 sharpened) and its red-proof (or is labelled a forward guard under D-000.24 or a regression
guard).** No AC may be satisfied by asserting on a fixture with an empty `pageHeader` — Finding 2
measures that such a subject cannot express the defect.

---

### AC1 — `internal/pagemodel` exists, and names no PDF concept

**Given** the spine's AD-5 (*"`internal/pagemodel` types name only geometry, glyph runs (font
identity + glyph ids + positions), images, and vector primitives"*)
**When** the package is inspected
**Then** `folio-go/internal/pagemodel/` exists and declares the page-model types
**And** those types carry **page-absolute** coordinates in `geom.Length` millipoints, top-left
origin, Y increasing downward
**And** the package imports **no** first-party package other than `internal/geom`
**And** no identifier in the package names a PDF object reference, a resource dictionary, a
content-stream operator, or a font *program* — glyph ids and a face **name** are page-model concepts;
an embedded, subsetted font program is not
**And** `internal/pdf`'s `TextPage`, `TextRun` and `ImagePlacement` (`internal/pdf/textdoc.go:68`,
`:149`, and the image struct) are **replaced by** the `pagemodel` types, **not duplicated beside
them** — two parallel type sets is D-1.4.14's drift hazard, and the whole point of AD-5 is that there
is one page model
**And** `EmbeddedFace` and `ImageXObject` stay in `internal/pdf` — they name PDF constructs and are
not page-model types

**Presence precondition.** The assertion that "the package names no PDF concept" must first assert
the package declares **at least one exported type with at least one field**, or it passes vacuously
on an empty package (D-000.9: an all-clear that a dead scan also produces).

**Red-proof.** Add a field of a `internal/pdf` type, or a field named for a resource dictionary, to a
`pagemodel` type in a scratch edit; the AC1 guard must fail naming the field. Record the message,
revert. *(The import half is subsumed by AC3's rank guard and needs no separate proof.)*

**Do not** declare a vector-primitive type here (scope fence; D-000.28).

---

### AC2 — `internal/layout` exists, and it alone places bands on the page

**Given** a loaded document and its bound data
**When** layout runs
**Then** `folio-go/internal/layout/` produces the page model of AC1
**And** the band-origin resolution currently at `render.go:100` `documentBands` lives in
`internal/layout`, and **no function in package `folio` computes a band origin any more**
**And** every element's page Y is `band origin + element.Y`, a **translation** — nothing in
`internal/layout` inverts, negates or subtracts a coordinate from a page height (D-2.0.4: *"a
translation, not an inversion"*)
**And** `internal/layout` receives what it needs as **parameters** — the atomic-span flags, the font
set, the bound values — never by importing a stage above it (D-000.16)

**Presence precondition.** The "no band origin in package `folio`" assertion must first confirm
package `folio` still declares `Render`, `RenderTo` and the single byte-producer of Finding 4 —
otherwise it passes on a deleted package.

**Red-proof.** Restore a band-origin computation in a package-`folio` function; the AC2 guard must
fail. Record and revert.

**Constraint from Finding 4.** After this AC, package `folio` must still contain **exactly one**
top-level non-test function that calls into `internal/pdf`.
`TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt` must be green **unmodified**.

---

### AC3 — The stage-rank import guard ships in the same commit that creates `internal/layout`

**Given** D-2.0.4 (*"the guard must exist from the package's first commit, not be added later"*) and
D-000.16
**When** the guard runs
**Then** a new lint rule with a stable rule id — `stage-rank`, alongside `forbidden-imports`,
`map-range`, `no-float-typed-value` in `lint/internal/rules/` — assigns each `folio-go/internal/`
package a rank and reports any import of an **equal or higher** rank as a finding carrying the rule
id, the importing file's path relative to the scanned root, and the line
**And** the ranks are exactly Finding 1's **corrected** table, subject to D-2.5-Q2
**And** an `internal/` package **not** in the table is a **finding, not a pass** — an unranked
package must be ranked deliberately (fail-safe default, the shape `mathAllowedCalls` already uses at
`forbiddenimports.go:82`)
**And** it follows D-1.3.3's **two-caller shape**: one pure checker taking a root directory and
returning `(findings, error)` with no `*testing.T` and no hard-coded root; a **production caller**
scanning the real `folio-go/` tree asserting **zero** findings; and a **fixture caller** scanning a
**retained violating fixture** asserting exactly the expected findings **by file and rule, never by
count** (`internal/arch_test.go`'s `assertExactFindings` is the precedent)
**And** the retained violating fixture lives under `folio-go/testdata/lint/stage-rank/` and contains
**at least** a `layout`-ranked file importing a `pdf`-ranked path — **AD-5's own arrow, named
explicitly**, because a guard written for a class must still be shown to cover the instance it exists
for (D-000.23, read the other way)
**And** the production caller carries a **vacuity guard reading the checker's own returned stats**
(not an independent re-walk — `internal/arch_test.go`'s Major 5 precedent): it must assert the
scanner reported visiting `layout`, `pagemodel`, `pdf`, `text` and `fontset` **by name**, and a
non-zero import count, so a dead scanner cannot produce the same zero-findings all-clear a healthy
one does (D-000.9)

**Red-proof, and it must actually be run (D-000.36 — a remedy for a vacuity inherits the vacuity's
hazard).** Three separate mutations, each recorded with its message:
1. Add `import ".../internal/pdf"` to a real `internal/layout` file → production caller fails naming
   the file and the arrow.
2. `if true { return nil, stats, nil }` as the checker's first statement → the **vacuity guard**
   fails (not merely "0 findings, pass").
3. Delete one expected finding from the fixture's expected set → fixture caller fails on the
   *unexpected reported* half; add a bogus one → fails on the *expected not reported* half. Both
   directions, per `assertExactFindings`.

**AC3 is why this story has a `DECISION NEEDED` gate.** Do not write the table until D-2.5-Q2 is
ruled.

---

### AC4 — AD-5's arrow, asserted by name as well as by rank

**Given** AC3's rank table makes `layout`(7) ↛ `pdf`(8) true by construction
**When** the arrow is asserted
**Then** there is **also** a named, single-purpose assertion that `internal/layout`'s transitive
first-party imports do not include `internal/pdf`, whose failure message quotes the grounding
sentence D-2.0.4 quotes: *"there is none from `internal/layout` to `internal/pdf`… it is precisely
the arrow a well-meaning commit will try to add"*

**Why both.** D-000.23: a guard written for a class is not evidence about the instance. A rank table
is a class guard; a future rank edit could permit this arrow without anyone noticing that AD-5 was
what the number meant. **Transitive**, not direct — an arrow laundered through a third package is
the same arrow.

**Presence precondition.** Assert `internal/layout` has at least one first-party import before
asserting `internal/pdf` is not among them; a package with no imports satisfies the assertion
vacuously.

**Red-proof.** Same mutation as AC3's (1); this guard must fail **as well**, and its message must be
the AD-5 one. If only one of the two fires, say which and why in the Delivery Log.

---

### AC5 — The content band's height is derived by exactly one function, from geometry alone

**Given** page dimensions, margins, and the two declared band heights
**When** the content area's height is computed
**Then** it is computed by **exactly one function** in `internal/layout`, as
`pageHeight − marginTop − marginBottom − pageHeaderHeight − pageFooterHeight`
**And** that function's inputs are **page geometry only** — it does not receive, and cannot consult,
the content band's elements or their measured sizes (AD-24: nothing negotiates; the content band
never grows to fit)
**And** "exactly one" is asserted **positively**, in the D-1.8.10 shape: every site that needs a
content height obtains it from that function, rather than a search for anyone who writes a
subtraction (*"not that 'nobody else writes a minus sign', which is unfalsifiable"*)
**And** the content band still declares **no** `height` key in `.folio` — `parse_bands.go:78`'s load
error for `bands.content.height` stays exactly as it is (regression; already true since Story 1.4)

**Presence precondition.** The positive guard must first assert the derivation function exists and is
called from at least one non-test site, or it passes on a function nobody calls.

**Red-proof.** (a) Compute a content height inline at a second site; the positive guard fails naming
it. (b) Drop `pageFooterHeight` from the derivation; **AC6's partition assertion** must fail — note
which, because a conservation-only assertion would not (see AC6).

---

### AC6 — The band partition is pinned by value, not by conservation (D-000.33, D-000.36)

**Given** AC7's fixture document, whose page geometry has **four pairwise-distinct** inputs
**When** the page model is inspected
**Then** these **four independent numbers** are asserted against **hand-derived literals**, stated in
the test source with the arithmetic that produced them:

| quantity | asserted as |
|---|---|
| page-header band origin | a literal (0 mp) |
| content band origin | a literal (= `pageHeaderHeight`) |
| page-footer band origin | a literal (= `pageHeight − marginTop − marginBottom − pageFooterHeight`) |
| content band **height** | a literal (AC5's derivation) |

**And** the assertion is explicitly **not** of the form `headerHeight + contentHeight + footerHeight
== usableHeight`, nor any other sum-to-whole
**And** the three origins are additionally asserted **pairwise distinct and strictly increasing**, so
a degenerate partition (any two bands sharing an origin, or a band collapsing to zero) is a failure
rather than a silent pass

**Why the sum is banned.** D-000.33: *"an additivity or conservation law is satisfied trivially by a
degenerate partition"*, and D-000.36 measured that the obvious remedy — a non-emptiness precondition
— **also stays green**, because *"additivity is preserved by any monotone boundary function"*. A
conservation assertion over band heights can detect a **saturating** error only. Four independent
literals over four distinct inputs can detect a **substitution** (footer height read where header
height belongs) and a **swap**, which is the error class this story can actually commit.

**Red-proof, all four required, each recorded, and run against the assertions as written (D-000.36 —
do not trust that these "obviously" close the hole):**
1. Swap `pageHeaderHeight` and `pageFooterHeight` in the derivation.
2. Swap `marginTop` and `marginBottom`.
3. Drop `pageFooterHeight` from the content-height derivation.
4. Set the page-header band origin to a non-zero value (the mutation Finding 2 measured **nothing**
   catches today).

**If any of the four does not redden, that is a finding, not a formality** (D-000.33, verbatim) —
report it and strengthen the subject, not only the assertion.

---

### AC7 — A new golden, `fixtures/three-band-page/`, with content in **all three** bands

**Given** Finding 2's measurement — every existing fixture has an empty `pageHeader`, and all six
share one page setup with `marginTop == marginBottom` and `pageHeader.height == pageFooter.height`
**When** the new fixture is recorded
**Then** `fixtures/three-band-page/` ships `input.folio`, `expected.pdf`, `expected.json` (the
`folioGoVersion` / `goToolchain` / `sha256` shape the other five use) and `README.md`
**And** its document has **at least one text element in each of the three bands**, each with a
**distinct, identifiable** bound or literal string so a band mix-up is visible in the rendered text
and not only in a coordinate
**And** its page setup makes all four geometric inputs **pairwise distinct** — `margin.top ≠
margin.bottom`, `pageHeader.height ≠ pageFooter.height`, and none of the four equal to another — so
no substitution among them is invisible in the bytes
**And** at least one element sits at a **non-zero band-relative `y`**, so "band origin" and "element
y" cannot be conflated
**And** the `README.md` states, in the shape `fixtures/wrapped-text/README.md` established: what each
element is for, **which defect the fixture's content is able to express** and which it is not, and
the AD-21/AD-22 rule that a hash change here is a defect until proven otherwise

**Semantic acceptance at first recording (D-000.22), machine-checkable half NEVER deferrable
(D-2.3.5).** All of the following are asserted at recording, in the story's own commit, none
deferred:
- the produced bytes are a well-formed PDF and are **not blank** (the `assertWellFormedPDF`
  precedent) — a hash over an empty page is Story 1.1's *"two empty files are byte-identical"*;
- exactly **three** distinct band origins are represented among the emitted runs, and each of the
  three bands contributed **at least one** run (the presence precondition — without it every
  band-origin assertion below is vacuous, which is exactly Finding 2's measured state);
- each element's emitted content-stream Y equals a **hand-computed literal**, shown with its
  arithmetic;
- the embedded faces are those the chain declares, and the leading between any wrapped element's
  lines is D-2.4.2's `max(hhea A − D + lineGap)` over the **declared** chain — carried forward, not
  re-derived;
- two independent process invocations produce byte-identical output.

**No human sign-off is created by this fixture, and none is needed.** Its text must be chosen so no
Thai *reading* judgment is required of it — the Thai reading obligation is D-2.3.5's, bound to
`fixtures/shaped-text`, and **this story must not create a second one**. If the developer believes a
Thai string is needed here, that is a `DECISION NEEDED`, not a judgment call.

**Registration** is governed by **D-2.5-Q1** and must not be done before that ruling.

---

### AC8 — **No existing golden moves.** All six recorded digests are byte-identical

**Given** the six recorded documents — `minimal-rect`, `font-text`, `image-embed`,
`multi-script-fallback`, `shaped-text`, `wrapped-text`
**When** the full suite runs after this story
**Then** every one of their `expected.json` `sha256` values is **unchanged**, and **no**
`expected.pdf` byte moves
**And** in particular `fixtures/shaped-text/expected.pdf`'s digest **`5964aad0…c92e00f`** is
untouched, because **D-2.3.5's pending human Thai reading sign-off is bound to it** and a re-record
would silently invalidate a sign-off that has not happened
**And** `fixtures/expected-breaks/expected_breaks.json` and D-2.4.3's pending break sign-off are
likewise untouched
**And** `fixtures/shaped-text/harfbuzz-oracle.json` is untouched

**This is the story's single strongest correctness signal.** Every band-composition change here is a
**relocation** of arithmetic, not a change to it; if a digest moves, layout changed a number, and
that is a defect until proven otherwise (AD-21). **Do not regenerate a golden to make a test pass.**

**If a digest genuinely must move**, that is a `DECISION NEEDED` under AD-21/AD-22 with the reason,
the affected documents, and the sign-off impact — **never a developer's judgment**, and never
resolved by re-recording first and explaining after.

**Presence precondition.** The digest-comparison assertion must confirm each `expected.pdf` is
non-empty and each `expected.json` parsed a non-empty `sha256` before comparing — comparing two
absent values passes.

---

### AC9 — The provisional pins retire, and their comments are corrected rather than deleted

**Given** `render_test.go:1094` `TestProvisionalBandOriginIsPinned`, which fires the moment
`internal/layout` exists (Finding 3)
**When** this story lands
**Then** that test is **deleted** — D-1.5.5's self-retiring-assertion pattern completing, not skipped
and not weakened
**And** site (1), `render.go`'s `PROVISIONAL (AC28)` band-origin convention (`:89`, `:223`,
`:116`–`:118`), is **replaced** by AC2's `internal/layout` composition, and its `PROVISIONAL` comments
go with it
**And** site (2), `internal/pdf/textdoc.go`, has its **stale comment corrected**: `TextRun.X/Y` are
**page-absolute** offsets from the page's top-left printable corner and that is **permanent, not
provisional** — under AD-24 band placement belongs to `internal/layout` alone, so making them
band-relative would move band placement into `internal/pdf` and violate AD-24 outright. The comments
at `textdoc.go:69`–`:77`, `:149`–`:151` and `:716`–`:727` must say this, and must **stop** citing a
deleted test
**And** the **baseline-offset** question at `textdoc.go:730` is recorded per **D-2.5-Q3** and
**not changed** — changing it re-records four goldens and invalidates a pending sign-off (AC8)
**And** `render_entry.go:141`–`:143`'s `PROVISIONAL` note is updated in the same sweep
**And** `render.go:26`'s `defaultFontSizePt` *"provisional"* label is **left alone** — it is a
different provisional (style defaults, Epic 3's `style` inheritance), and sweeping it here would be
scope creep dressed as tidying

**D-000.28 applies to every comment written in this AC.** Do not write *"the real band-relative
placement has arrived"* into any file before the code that makes it true is in the same commit.

**D-000.34 applies at the moment this lands.** `TestProvisionalBandOriginIsPinned` was one of the
tests whose discriminating power depended on `internal/layout` being **absent**. Sweep the suite for
any other assertion whose power depended on the pre-2.5 shape — in particular anything asserting on
`documentBands` from `layout_probe_test.go:54` — and report each as *re-pointed*, *retired* or
*still discriminating*, with which.

---

### AC10 — Regression guards that must stay green, **unmodified**, and are not evidence this story landed

Each is **already true at `5968c5d`** (see *Corrections*). They are listed so a break is caught, and
labelled so none is mistaken for new work (D-000.35: symbols named, checkable).

| # | property | the guard, by name |
|---|---|---|
| a | exactly three bands; a fourth or a missing one is a load error; group/report headers and group footers are absent and unsupported | `internal/template/parse_bands.go:15` `decodeBands`; `closedsets_test.go`, `fixtures_test.go` |
| b | element `x`/`y` are band-relative in the document model | `parse_bands.go:100`–`:120`; `folio-format.md:106` |
| c | the flip to PDF user space happens in exactly one function | `internal/pdf/flip.go:29` `flipY`; `flip_test.go:173` `TestContentStreamYCoordinatesRouteThroughFlipY` + `:212` `TestFlipRoutingRedProof` |
| d | exactly one document-byte producer; `Render` and `RenderTo` both route through it | `render_arch_test.go:145` (Finding 4 — this constrains the refactor's shape) |
| e | no `float64` / `float32`, type-aware, module-wide | `internal/arch_test.go` `TestNoFloat64UnderInternal`, `TestNoFloat64UnderModule`; `lint` `no-float-typed-value` |
| f | no `range` over a map | `lint` `map-range` |
| g | `go list -m all` is exactly two modules | `gomod_test.go`'s `wantModuleGraph` |
| h | no compressor import; no `SOURCE_DATE_EPOCH` literal anywhere under `folio-go/` **including inside error-message strings** | `lint` `no-compressor-import`; `absence-source-date-epoch` |
| i | matrix slugs registered in both lists, with four upload paths each | `matrix_registration_test.go:32` |
| j | the Epic 2 gate owes no fourth obligation | `TestStory23aAddedNoThirdEpic2GateObligation` — **see D-2.5-Q1** |

**Do not modify any of these tests to accommodate the refactor.** If one of them obstructs the
design, that is a finding about the design, and a `DECISION NEEDED`.

---

### AC11 — `TestCorpusMeetsP6ExerciseFloors` is still red, for the same reason, with the same numbers

**Given** the baseline's single failure
**When** the suite runs after this story
**Then** `internal/text`'s `TestCorpusMeetsP6ExerciseFloors` still fails with **`P6g: got 7, need
>=20`** and stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`
**And** the total is **465 PASS · 1 FAIL** or higher on the PASS side, all-occurrences — **one**
failure, that one
**And** nothing was added to the corpus, and no floor was lowered (D-000.17)

**Why this is an AC and not a note.** Story 2.4's AC5 holds it open deliberately. A story that
"fixed" it would be hiding another story's disclosed shortfall, and a story that let a *second*
failure appear would be reporting a regression as "the known red".

---

## Task breakdown

Work in order. **Stop at the end — do not commit, do not branch, do not set `done`.**

1. [x] **Re-verify the baseline.** Run the four invocations in *Baseline, measured at creation* and
   confirm 464/1, 293/1, `lint` green, two modules. Confirm `git status --porcelain` shows the
   single expected decision-log modification and **nothing else**. If any number differs, the
   baseline moved — stop and report.
2. [x] **Escalate.** Confirm **D-2.5-Q1**, **D-2.5-Q2**, **D-2.5-Q3** and **D-2.5-Q4** are ruled before
   touching AC3 or AC7. Record each ruling verbatim in the Delivery Log.
3. [x] **Re-measure Finding 1's import graph yourself** and validate the corrected rank table against it.
   Do not write the table from this file without re-deriving it.
4. [x] **AC1** — create `internal/pagemodel`, move the page-model types out of `internal/pdf`, and point
   `internal/pdf` at them. Expect zero golden movement; run the suite and confirm AC8 holds after
   this step alone, before anything else changes.
5. [x] **AC3 + AC4** — write the `stage-rank` lint rule, its retained violating fixture, both callers and
   the vacuity guard. **Land this before `internal/layout` has a single line of production code**
   (D-2.0.4: *"the guard must exist from the package's first commit"*). Run all three red-proofs and
   record their messages.
6. [x] **AC2 + AC5** — create `internal/layout`; move `documentBands` and the band-origin resolution
   into it; add the content-height derivation function. Delete
   `TestProvisionalBandOriginIsPinned` (**AC9**) in the same step — the suite is red until you do.
   Keep exactly one package-`folio` function calling `internal/pdf` (Finding 4).
7. [x] **AC6** — write the four-literal partition assertions and run all four red-proofs. **If any does
   not redden, stop, report it as a finding, and strengthen the subject.**
8. [x] **AC7** — author `fixtures/three-band-page/input.folio` with all three bands populated and four
   distinct geometric inputs; record `expected.pdf` / `expected.json`; write the semantic acceptance
   assertions **before** freezing the hash; write the `README.md`. Register per D-2.5-Q1's ruling.
9. [x] **AC8** — verify all six existing digests byte-identical. Diff `fixtures/` and confirm the only
   additions are `three-band-page/`.
10. [x] **AC9** — correct the `internal/pdf` and `render_entry.go` comments; run D-000.34's sweep for
    tests that quietly lost their teeth, and report each disposition.
11. [x] **AC10 + AC11** — full suite. Confirm the regression table green and exactly one failure, the
    expected one, with the expected numbers.
12. [x] **Record.** Fill the Dev Agent Record and Delivery Log: every red-proof message verbatim, every
    ruling verbatim, the D-000.34 sweep, any deviation and why. Propose (do **not** write) any new
    `sprint-status.yaml` key or DW entry in the Delivery Log.
13. [x] **Story file, decision log, sprint status → `review`.**

**Stop here — do not commit, do not branch, do not set `done`.** Committing belongs to the finisher,
after review.

---

## Heavy-test cadence — what is deferred, and to which gate

The four-target hash matrix (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) runs **per
epic** under D-000.4, with per-story overrides for hash-shaped stories.

**This story proposes an override — see D-2.5-Q4 — and does not assume one.** The case for it is
that band composition feeds every subsequent measurement the way line breaking did (Finding 2
measures that 5 of 6 goldens already move when a band origin moves), and that this story relocates
the whole render path across a package boundary. The case against is that no new numeric rule is
introduced and AC8 already requires byte-identity of six documents.

**Until D-2.5-Q4 is ruled, the matrix legs for `three-band-page` are deferred to the Epic 2 boundary
gate**, and the story must say so plainly rather than implying they ran. **Do not write
"four-target agreement" into `fixtures/three-band-page/README.md` before the legs have actually
run** (D-000.28: an anticipatory claim is false from birth and reads identically to a true one). If
the override is granted, record the four digests and the byte lengths as `wrapped-text/README.md`
does; if it is refused, the README says *"registered; legs deferred to the Epic 2 gate"*, in those
terms.

**Housekeeping for the epic gate, flagged and not fixed:**
`_bmad-output/implementation-artifacts/sprint-status.yaml` still reads **`epic-2: backlog`** while
`2-1` through `2-4` are `done`. Per the tracker's own workflow note the epic should have moved to
`in-progress` at 2.1. Reported here for the gate; **this story does not edit it** beyond its own
key's status.

---

## Dev Agent Record — completion notes

**Baseline commit:** `213d98c` (`main`). The story file's *Baseline, measured at creation* names
`5968c5d`; the tree advanced by one commit before development began, and **every baseline number
re-measured identical** (below), so the measurements this story rests on still hold.

### Task 1 — baseline re-verified (D-000.26: scope, invocation, counting rule for each)

| scope | invocation (verbatim, `rtk proxy` first then redirect — D-000.12 corrected) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, every `--- PASS`/`--- FAIL` occurrence | **464 PASS · 1 FAIL** |
| `folio-go/` | same invocation, top-level only (`^--- PASS` at column 0) | **293 PASS · 1 FAIL** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 ./...` | **4 packages `ok`, 0 FAIL** |
| `folio-go/` | `GOWORK=off go list -m all` | exactly **two** modules |

The single failure was `internal/text`'s `TestCorpusMeetsP6ExerciseFloors` — *"P6g (opaque names)
floor not met: got 7, need >=20"*, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.

`git status --porcelain` showed **two** modified files, not one: the expected decision-log append,
**and** `sprint-status.yaml` (`2-5` moved `backlog` → `ready-for-dev` by the orchestrator). Neither
is a `folio-go/` or `lint/` source file, so every measurement is against `213d98c`'s code exactly.

### Task 2 — the four escalations were ruled before development; recorded verbatim

**D-2.5-Q1 → ruled as `D-2.5.1`** *(mechanism: binding)*: *"The fourth Epic 2 gate obligation is
sanctioned: a new golden that is not matrix-registered is exactly the drift Story 2.3 closed. But
`TestStory23aAddedNoThirdEpic2GateObligation` encodes a count in its name, and that name rots on
every future obligation. Replace it with an assertion that the registered obligation set equals an
explicit declared list. Adding one then becomes a one-line diff to that list — visible and
reviewable — instead of a rename."*

**D-2.5-Q2 → ruled as `D-000.16 (ranks corrected)`** *(mechanism: binding)*: *"This is not a
falsified ruling — it is the validation the ruling required… Swap ranks 5 and 6. Both arrows the
guard exists to enforce still hold: `layout`(7) ↛ `pdf`(8), `expr`(3) ↛ `bind`(4). **The table is
ratified as measured from here**, no longer illustrative."* With the reason: *"subsetting needs the
glyph set; the glyph set comes from shaping — AD-8: 'one subset per font per document over the union
of glyphs used.' **You cannot know the union without shaping first.** So **shape → collect →
subset** is the true pipeline order and `text` genuinely precedes `fontset`."*

**D-2.5-Q3 → ruled as `D-000.37`** *(mechanism: binding)*: *"`TestProvisionalBandOriginIsPinned`
fires correctly on `mkdir internal/layout` — a **true** positive — but **its message's site-(2)
instruction is half stale, and following it literally would silently move goldens**… The remedy text
is maintained as code, not as prose: any story that changes the mechanism a tripwire points at
**updates the message in the same commit.** Story 2.5 assigns the deletion to an AC rather than
leaving the stale instruction to be obeyed."* The baseline-offset defect itself is **DW-15**,
scheduled and **not touched here**.

**D-2.5-Q4 → ruled as `D-000.4 (override criterion)`, override DECLINED** *(mechanism: binding)*:
*"An override is warranted when a story introduces a new **source of cross-target divergence** —
float arithmetic, a vendor call, a compressor, a new dependency — **not merely because it records a
new golden**… **Story 2.5 is integer band arithmetic on `geom.Length`**: no floats, no new vendor
surface… If 'records a golden' sufficed, nearly every story would qualify."*

**Consequence, stated plainly rather than implied: the four-target matrix legs for
`three-band-page` were WRITTEN AND REGISTERED but NOT RUN.** No leg of any target executed for this
document in this story. They are the Epic 2 boundary gate's. `fixtures/three-band-page/README.md`
says *"Registered; legs deferred to the Epic 2 gate"* in those terms and claims no four-target
agreement (D-000.28).

### Task 3 — the import graph re-measured, and the corrected table validated against it

Measured here, not copied: `CGO_ENABLED=0 GOWORK=off go list -f '{{.ImportPath}} :: {{join .Imports
" "}} :: {{join .TestImports " "}}' ./internal/...`, first-party edges only:

```
internal            -> (none)                        (the test-only `arch` package)
internal/bind       -> internal/template
internal/fontset    -> internal/geom  internal/text  <-- confirms Finding 1
internal/geom       -> (none)
internal/pdf        -> internal/geom
internal/template   -> internal/geom
internal/text       -> (none)
```

`fontset -> text` is real and load-bearing (`fontset.go:25`, `:74`, `:245`, `:445` — Story 2.3a's
vendor containment). Under D-000.16's *published* ranks that edge points upward; under the corrected
ranks the whole graph is **clean and forward-only, zero violations**, which is D-000.16's own stated
precondition ("green today"). Verified by running the production caller: green.

**One row the story file did not anticipate, added deliberately.** The scan root itself
(`folio-go/internal/`, holding the test-only `arch` package) is a directory the checker visits and
therefore needs a rank, or the fail-safe "unranked is a finding" rule reddens on shipped code. It is
ranked **`rankNoStage` (-1)** — *below* every stage, not above — so it may import no first-party
internal package at all, which is exactly what it does today (measured: both its files import only
the standard library). Ranking an observer *last* would have licensed it to import everything.

### Task 4 — AC1

`internal/pagemodel` created; `TextPage`→`Page`, `TextRun`, `ShapedGlyph` and `ImagePlacement`
**moved** out of `internal/pdf` (replaced, not duplicated — D-1.4.14's drift hazard). `EmbeddedFace`,
`ImageXObject` and `CIDText` stayed. **`ImagePlacement.ResourceName` was renamed `AssetKey`**: it
named a PDF *resource dictionary* from inside the page model, which is on AC1's own enumerated ban
list. The suite was run after this step alone, before anything else changed: **all six goldens
byte-identical, only the known red** — AC8 held at the type move, as the task required.

**One reading recorded rather than assumed.** `ShapedGlyph.CID` keeps its name. A CID is not on
AC1's enumerated list (object reference, resource dictionary, content-stream operator, font
program), and AD-5 names "glyph ids" as a page-model concept. Moving CID *allocation* out of package
`folio` into `internal/pdf` would be a re-architecture of `buildShapedPDFRuns`, which neither AC1's
"replaced by, not duplicated beside" clause nor Task 4's "move the types and point internal/pdf at
them" asks for. Flagged for review as the one judgment in AC1 that could be read the other way.

### Tasks 5–11

All ACs met; every red-proof run and recorded in the Delivery Log below. Full suite after the
story: **475 PASS · 1 FAIL** all-occurrences, **304 · 1** top-level, the one failure being
`TestCorpusMeetsP6ExerciseFloors` with byte-identical numbers. `lint/`: **84 · 0** all-occurrences,
**46 · 0** top-level, 4 packages `ok`. `go list -m all`: two modules.

---

## Delivery Log

### AC-by-AC

| AC | status | evidence |
|---|---|---|
| **AC1** | met | `folio-go/internal/pagemodel/pagemodel.go`; guard `TestPageModelNamesNoPDFConcept` (`internal/bandcomposition_arch_test.go`), red-proof RP-1 below |
| **AC2** | met | `internal/layout/band.go`; `render.go`'s `pageGeometryOf`/`documentBands` now *ask* rather than derive; guard `TestNoBandOriginArithmeticInPackageFolio`, red-proof RP-6 |
| **AC3** | met | `lint/internal/rules/stagerank.go` (rule id `stage-rank`), fixture `folio-go/testdata/lint/stage-rank/`, callers `stagerank_test.go`; three red-proofs RP-2/3/4 |
| **AC4** | met | `TestInternalLayoutDoesNotReachInternalPDF` — transitive, named, with the AD-5 quote; red-proofs RP-5 and RP-5b (2-hop) |
| **AC5** | met | `layout.ContentHeight`; guard `TestContentHeightIsDerivedByExactlyOneFunction` (declared once, called, parameter is `PageGeometry` only); red-proof RP-7 |
| **AC6** | met, **with a measured finding** | `TestThreeBandPagePartitionIsPinnedByValue`; four red-proofs RP-8..RP-11, plus RP-9b which **stays green** — reported below |
| **AC7** | met | `fixtures/three-band-page/`; `TestThreeBandPageSemanticAcceptance` run *before* the digest was frozen |
| **AC8** | met | `git status --porcelain fixtures/` reports **only** `?? fixtures/three-band-page/`. No existing `expected.pdf` or `expected.json` byte moved; `shaped-text` is still `5964aad0…c92e00f` |
| **AC9** | met | `TestProvisionalBandOriginIsPinned` **deleted**; comments corrected at `internal/pdf/textdoc.go`, `render.go`, `render_entry.go`; `defaultFontSizePt`'s label left alone |
| **AC10** | met | every listed guard green **and unmodified**, except `(j)` which D-2.5.1 ordered replaced |
| **AC11** | met | 475 · 1 all-occurrences; the one failure is `P6g: got 7, need >=20`, stats identical to baseline; nothing added to the corpus, no floor lowered |

### Red-proofs — every one RUN against the guard as written (D-000.36), message verbatim

**RP-1 (AC1) — a field named for a resource dictionary.** Added `ResourceDict string` to
`pagemodel.TextRun`:

> `pagemodel.go:71: internal/pagemodel's field "ResourceDict" names a PDF concept ("resource"). AD-5: "the page model knows nothing about PDF" — its types name only geometry, glyph runs (font identity + glyph ids + positions) and images. A face NAME and a glyph id belong here; a resource dictionary, an object reference, a content-stream operator and an embedded font PROGRAM do not (those are internal/pdf's EmbeddedFace and ImageXObject, and they stay there)`

(and a second finding for the substring `"dict"` — the substring list is deliberately overlapping).

**RP-2 (AC3.1) — `internal/layout` imports `internal/pdf`.** Production caller:

> `stagerank_test.go:58: stage-rank violations under folio-go/internal/ (D-000.16):`
> `layout/band.go:40: stage-rank violation: "layout" (rank 7) imports "pdf" (rank 8) — a HIGHER rank. The pipeline is strictly forward: a package may import only a STRICTLY LOWER rank. The signal rides on the VALUE, never through an import — stages communicate by what they pass, not by what they import (D-000.16), so pass what is needed as a parameter. Known ranks: geom=0 diag=1 pagemodel=1 template=2 expr=3 bind=4 text=5 fontset=6 layout=7 pdf=8 .=-1`

**RP-3 (AC3.2) — `if true { return nil, stats, nil }` as `ScanStageRank`'s first statement.** The
**vacuity guard** fired, not a "0 findings, pass":

> `stagerank_test.go:42: vacuity guard: the scanner's own stats do not report visiting package "layout" — a scan that never entered it reports the same zero findings a healthy scan does (D-000.9). Visited: []`
> `stagerank_test.go:96: vacuity guard: the scanner's own stats report 0 files parsed under …/testdata/lint/stage-rank — the retained fixture is missing, and an empty tree produces the same empty finding set a compliant one does`
> `stagerank_test.go:132: presence precondition: no finding reported for layout/violating_pdf_import.go — the message assertions below would pass vacuously on an empty string`

**RP-4 (AC3.3) — the fixture's expected set, BOTH directions.** Deleting one expected finding:

> `stagerank_test.go:104: unexpected finding reported: file=expr/violating_bind_import.go rule=stage-rank`
> `stagerank_test.go:104: distinct (file,rule) pair count mismatch: got 4, want 3`

Adding a bogus one (naming the *compliant* fixture file):

> `stagerank_test.go:106: expected finding not reported: file=template/compliant_lower_rank_import.go rule=stage-rank`
> `stagerank_test.go:106: distinct (file,rule) pair count mismatch: got 4, want 5`

**RP-5 (AC4) — the same mutation as RP-2. BOTH guards fire; AC4's carries the AD-5 sentence.**

> `bandcomposition_arch_test.go:340: AD-5 VIOLATED: internal/layout reaches internal/pdf (internal/layout -> internal/pdf).`
> `The spine's first line of grounding, verbatim: "One dependency arrow matters more than the rest: there is none from internal/layout to internal/pdf. That absence is what keeps PNG/SVG/HTML renderers possible later (AD-5), and it is precisely the arrow a well-meaning commit will try to add."`
> `The remedy is never to relax the rank table: pass what layout needs as a PARAMETER, or move the value into internal/pagemodel (rank 1), which both stages may see.`

**RP-5b (AC4, transitivity) — a 2-hop LAUNDERED arrow.** `layout` importing `text`, `text`
importing `pdf`. AC4 requires transitive, not direct; measured rather than assumed:

> `AD-5 VIOLATED: internal/layout reaches internal/pdf (internal/layout -> internal/text -> internal/pdf).`

**RP-6 (AC2) — band-origin arithmetic restored in a package-`folio` function.**

> `bandcomposition_arch_test.go:570: AC2 violation — package folio computes a band origin:`
> `render.go:144: function documentBands subtracts MarginTop`
> `render.go:145: function documentBands subtracts usableHeight`

**RP-7 (AC5a) — a second, inline content-height derivation inside `internal/layout`.**

> `bandcomposition_arch_test.go:477: AC5 violation — the content band's height must be derived by ContentHeight and nowhere else:`
> `band.go:122: function FooterOriginBypass subtracts band height PageHeaderHeight directly`

**RP-8 (AC6.1) — swap `pageHeaderHeight` and `pageFooterHeight` in the derivation.**

> `content band origin = 24000 mp, want 18000 mp (= pageHeader.height, 18000)`
> `page-footer band origin = 751890 mp, want 745890 mp (= 841890 - 30000 - 42000 - 24000)`

plus three content-stream Y literals and the digest. **Note which literal did NOT move: the content
band HEIGHT is unchanged (727890) under this swap**, because both terms are subtracted. A test that
pinned only the height would have passed. That is precisely why AC6 demands four *independent*
numbers.

**RP-9 (AC6.2) — swap `marginTop` and `marginBottom` at the reading site** (`pageGeometryOf`):

> `presence precondition: fixture's margin.top reads 42000 mp, want 30000 mp — the literals below were hand-derived from the fixture's declared setup and mean nothing against a different one`

plus **all four** content-stream Y literals and the digest.

**RP-9b (AC6.2, variant) — the SAME swap written INSIDE the derivation** (`ContentHeight`'s own
expression, `g.Height - g.MarginBottom - g.MarginTop - …`): **STAYS GREEN on all three tests and on
the digest.** Reported as required rather than glossed. **Diagnosis: it is not a defect.** Both
margins enter the partition only as a **sum**, and `a-b-c == a-c-b` exactly, so no number and no
byte changes — this is D-000.33's "additivity is preserved by any monotone boundary function"
restated one level up, and no assertion over derived quantities can ever see it. **The subject was
strengthened where strengthening is possible**: `margin.top` and `margin.bottom` are pinned
**individually** by literal at the input, so the substitution that *is* a defect (RP-9) reddens. The
`three_band_page_fixture_test.go` comment records this measurement rather than the claim originally
written there before it was run (D-000.28).

**RP-10 (AC6.3) — drop `pageFooterHeight` from the content-height derivation.**

> `page-footer band origin = 769890 mp, want 745890 mp (= 841890 - 30000 - 42000 - 24000)`
> `content band height = 751890 mp, want 727890 mp (= 841890 - 30000 - 42000 - 18000 - 24000)`

**RP-11 (AC6.4) — the page-header band origin set non-zero. THE mutation Finding 2 measured NOTHING
in the repository could catch.**

> `page-header band origin = 1000 mp, want 0 mp (the top of the printable column). Story 2.5's creator measured that a WRONG page-header origin was detected by ZERO tests in this repository before this fixture existed`
> `run 0 (e4, pageHeader band): content-stream placement is "1 0 0 1 36 797.89 Tm", want the hand-computed "1 0 0 1 36 798.89 Tm"`
> plus the digest.

**Finding 2 is closed.** A page-header band origin off by one millipoint now reddens three
independent assertions and the golden.

**RP-12 (D-2.5.1's replacement guard) — both directions.** A fifth `//go:build matrix` file:

> `"matrix-file: zz_fifth_matrix_test.go" is an Epic 2 gate obligation and is NOT in declaredEpic2GateObligations.`
> `An obligation may not be added without a ruling that says so explicitly. Record it as a ONE-LINE addition to that list, naming the story and the decision that authorised it (D-2.5.1). Do not rename this test, and do not encode a count anywhere.`

A declared obligation disappearing from the tree:

> `declared Epic 2 gate obligation "matrix-document: three-band-page" is NOT present in the tree — obligations must not disappear silently either. Either restore it, or remove its line from declaredEpic2GateObligations with the ruling that discharged it.`

A control run after every restore was green in each case (mutated files restored with `cp` from a
scratch copy, never `git checkout` — `internal/layout/band.go` is untracked at this commit and a
checkout would have deleted it).

### A defect the new guard found in shipped code

The stage-rank production caller reddened on its first real run against
`internal/fontset/vendorboundary_test.go:17` — Go's **external-test-package idiom**
(`package fontset_test` importing `fontset`), reported as `"fontset" (rank 6) imports "fontset"
(rank 6) — an EQUAL rank`. A package talking to itself is one stage, not an arrow between stages.
`ScanStageRank` now skips `target == stage` **and does not count it as a first-party edge examined**
— counting it would have let the vacuity guard's `FirstPartyImports` be satisfied entirely by
self-imports.

### D-000.34 sweep — tests whose discriminating power depended on the pre-2.5 shape

| test | disposition | why |
|---|---|---|
| `render_test.go` `TestProvisionalBandOriginIsPinned` | **RETIRED (deleted)** | its whole power was `os.Stat("internal/layout")` failing; D-1.5.5's self-retiring pattern completing. Its stale site-(2) remedy was **not** obeyed (D-000.37) |
| `byte_neutrality_test.go` `TestStory23aAddedNoThirdEpic2GateObligation` | **RETIRED, replaced and strengthened** | D-2.5.1. `TestEpic2GateObligationsMatchTheDeclaredSet` subsumes it and additionally pins the registered matrix **documents**, which the old inventory never saw |
| `layout_probe_test.go:54` (`documentBands`) | **STILL DISCRIMINATING, re-pointed** | it consumes band origins to build element layouts and never asserted on how they were derived; `documentBands` still exists with the same signature, now answered by `internal/layout` |
| `internal/pdf/flip_test.go` `TestContentStreamYCoordinatesRouteThroughFlipY` + `TestFlipRoutingRedProof` | **STILL DISCRIMINATING, unmodified** | unchanged scan root, unchanged property. Finding 5's gap — it cannot see an inversion inside `internal/layout` — is now covered positively by `internal/layout/band_test.go`'s `TestPlaceInBandIsATranslationNotAnInversion`, not by widening this one |
| `render_arch_test.go` `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt` | **STILL DISCRIMINATING, unmodified** | green without modification: `renderDocument` remains the only package-`folio` function that *calls* through the `internal/pdf` alias. `buildShapedPDFRuns` names `pdf.EmbeddedFace` as a type but calls nothing |
| `internal/pdf/textdoc_test.go` | **STILL DISCRIMINATING, re-pointed** | mechanically re-pointed at `pagemodel.Page` / `pagemodel.TextRun`; assertions unchanged |
| `byte_neutrality_test.go` `TestStory23aMovedNoGoldenDigest` | **STILL DISCRIMINATING, unmodified** | five independent digest literals, all still matching. `three-band-page` was deliberately **not** added: AC10 says do not modify these, and AC8's own evidence covers the sixth |

### Proposed — NOT written (the story reserves these)

*Finisher sweep: item 1 is CLOSED and rewritten in place (it had gone stale — reviewer Finding 4);
items 2 and 3 remain accurate and open, with item 2's parenthetical corrected for the finisher's own
status edit. `DW-16` is NOT listed here — unlike these, it was actually written (`deferred-work.md`).*

1. **`DW-15` — CLOSED as an action; the citation resolves.** *(Corrected by the finisher; reviewer
   Finding 4, a Nit.)* This item was written when `deferred-work.md`'s last entry was `DW-14`, and
   it was true then. **DW-15 now exists**, recorded by the owner at `deferred-work.md:493` after the
   developer flagged the dangling citation, so `internal/pdf/textdoc.go`'s comment resolves to a
   real entry. The recorded entry is materially the one proposed here — first-baseline placement
   from `run.FontSize` against D-2.4.2's `max(hhea ascent − descent + lineGap)` leading, two models
   for one typographic quantity — and it carries a **binding sequencing obligation this story did
   not propose**: the fix must land *before* D-2.3.5's Thai reading sign-off is recorded, not after,
   because that sign-off binds to the rendered image and a baseline shift would ask the human twice.
   Nothing further is owed by this story; the baseline offset itself is untouched here (AC9).
2. **`sprint-status.yaml` still reads `epic-2: backlog`** while `2-1`…`2-4` are `done` — and now
   `2-5` as well. **Still flagged for the gate, and deliberately NOT changed here**: the epic key is
   the boundary gate's to flip, and flipping it from inside a story would assert an epic-level
   judgment no story is entitled to make. This story edited only its own key
   (`ready-for-dev` → `review` by the developer, → `done` by the finisher).
3. **No decision-log entry was written by this story**, and the finisher wrote none either. Every
   ruling applied was already recorded by its author (`D-2.5.1`, `D-000.16 (ranks corrected)`,
   `D-000.37`, `D-000.4 (override criterion)`, `D-000.26 (refined)`), and the finisher's own work
   applied `D-000.36` and `D-000.37` as already ruled rather than establishing anything new. The
   RP-9b finding — that a `marginTop`/`marginBottom` swap inside `ContentHeight` is a provable
   *semantic no-op* rather than an uncovered defect, since integer subtraction of a commutative sum
   reassociates exactly and the produced digest is byte-identical — remains the strongest candidate
   for recording, because it is a general rule about when a surviving mutation is evidence of
   equivalence rather than of a gap. Item 2 is the other. **Both are for the gate, not for this
   story.**

### Measured gates, from real runs

| scope | invocation | before | after |
|---|---|---|---|
| `folio-go/` all-occurrences | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | 464 · 1 | **475 · 1** |
| `folio-go/` top-level | same, `^--- PASS` at column 0 | 293 · 1 | **304 · 1** |
| `lint/` all-occurrences | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | 81 · 0 | **84 · 0** |
| `lint/` top-level | same | 43 · 0 | **46 · 0** |
| modules | `GOWORK=off go list -m all` | 2 | **2** |
| `gofmt -l folio-go lint` (excluding `testdata/`) | — | clean | **clean** |
| `SOURCE_DATE_EPOCH` literal under `folio-go/**/*.go` | — | 0 | **0** |

*This table measures the DEVELOPER's handover, at the tree the reviewer received. It is left
unaltered as that record. The finisher's own runs moved three of these rows and are recorded
separately below, under "Finisher — measured gates".*

The +11 net top-level tests account exactly: 4 new architectural guards, 4 new
`internal/layout` tests, 4 new `three-band-page` tests, minus the 1 deleted pin. `lint/` gains 3
(the stage-rank production, fixture and message callers). **The one remaining failure is
`TestCorpusMeetsP6ExerciseFloors`**, with `P6g: got 7, need >=20` and stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — byte-identical to the baseline. Nothing was
added to the corpus and no floor was lowered (D-000.17).

### File List

**New**

- `folio-go/internal/pagemodel/pagemodel.go`
- `folio-go/internal/layout/band.go`
- `folio-go/internal/layout/band_test.go`
- `folio-go/internal/bandcomposition_arch_test.go`
- `folio-go/three_band_page_template.go`
- `folio-go/three_band_page_fixture_test.go`
- `folio-go/testdata/lint/stage-rank/layout/violating_pdf_import.go`
- `folio-go/testdata/lint/stage-rank/expr/violating_bind_import.go`
- `folio-go/testdata/lint/stage-rank/pagemodel/violating_equal_rank.go`
- `folio-go/testdata/lint/stage-rank/unrankedstage/violating_unranked.go`
- `folio-go/testdata/lint/stage-rank/template/compliant_lower_rank_import.go`
- `folio-go/testdata/lint/stage-rank/pdf/compliant_pagemodel_import.go`
- `folio-go/testdata/lint/stage-rank/compliant_root_no_first_party.go`
- `lint/internal/rules/stagerank.go`
- `lint/internal/rules/stagerank_test.go`
- `fixtures/three-band-page/input.folio`
- `fixtures/three-band-page/expected.pdf`
- `fixtures/three-band-page/expected.json`
- `fixtures/three-band-page/README.md`

**Modified**

- `folio-go/render.go`
- `folio-go/render_entry.go`
- `folio-go/render_test.go`
- `folio-go/matrix_test.go`
- `folio-go/byte_neutrality_test.go`
- `folio-go/internal/pdf/textdoc.go`
- `folio-go/internal/pdf/textdoc_test.go`
- `folio-go/internal/pdf/imagedoc.go`
- `.github/workflows/matrix.yml`
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (this story's key only)
- `_bmad-output/implementation-artifacts/2-5-compose-a-page-from-three-bands.md`

**Deleted**

- `folio-go/render_test.go`'s `TestProvisionalBandOriginIsPinned` (AC9)
- `internal/pdf`'s `TextPage`, `TextRun`, `ShapedGlyph`, `ImagePlacement` (moved to
  `internal/pagemodel`, not duplicated)

### Change Log

| date | change |
|---|---|
| 2026-08-24 | Story 2.5 implemented against baseline `213d98c`: `internal/pagemodel` and `internal/layout` created, the `stage-rank` lint rule landed in the same change as `internal/layout`'s first line, `fixtures/three-band-page/` recorded and registered (legs deferred), `TestProvisionalBandOriginIsPinned` retired, the Epic 2 gate-obligation guard converted to declared-set equality per D-2.5.1. All six existing goldens byte-identical. Status → `review`. |

---

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-24
- **Baseline:** `213d98c`; work uncommitted. Every mutation below was applied by hand and reverted by
  hand (`cp` from a pre-review `rsync` snapshot — never `git checkout`, never `git stash`, because
  `internal/layout/`, `internal/pagemodel/`, `testdata/lint/stage-rank/` and both `lint/` rule files
  are **untracked** at this commit and a checkout would have deleted them). A final `diff -rq`
  against that snapshot reports **0 differences** across `folio-go/`, `lint/`, `fixtures/` and
  `.github/`, and `git status --porcelain` is **byte-identical** to the tree handed over.
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 0 · **Majors:** 2 · **Minors:** 1 · **Nits:** 1

### Gates re-measured by the reviewer (not read from the record)

| scope | invocation | measured |
|---|---|---|
| `folio-go/` all-occurrences | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **475 PASS · 1 FAIL** ✓ |
| `folio-go/` top-level (`^--- PASS` at column 0) | same | **304 · 1** ✓ |
| `lint/` all-occurrences | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` | **84 · 0** ✓ |
| `lint/` top-level | same | **46 · 0** ✓ |
| modules | `GOWORK=off go list -m all` | **2** ✓ |
| `gofmt -l` (both modules, excl. `testdata/`) | — | clean ✓ |
| `go vet ./...` (both modules) | — | clean ✓ |

**The single failure is the sanctioned one.** `TestCorpusMeetsP6ExerciseFloors` —
`P6g (opaque names) floor not met: got 7, need >=20`, stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — **byte-identical to baseline**, satisfying
Story 2.4's AC5 and D-000.17. **No second failure at any point.**

---

### Verified by construction — the six areas attacked hardest

**1. The population fix (AC6/AC7) — the story's central premise. CLOSED, and I re-ran the
mutation.** `layout.Origins`' `PageHeader: 0` → `1000` — the injection Finding 2 measured was caught
by **ZERO** tests before this story — now reddens **four** independent assertions plus the golden:
`TestOriginsPartitionThePrintableColumnExactly`, `TestThreeBandPagePartitionIsPinnedByValue`,
`TestThreeBandPageSemanticAcceptance` (`"1 0 0 1 36 797.89 Tm"` vs the hand-computed `798.89`) and
`TestThreeBandPageGoldenFixture`. All four AC6 red-proofs were re-run independently and **all four
fire**: header/footer-height swap in `Origins` (content origin 24000 ≠ 18000, footer 751890 ≠
745890, three Tm literals, digest); margin swap at the reading site; `PageFooterHeight` dropped from
`ContentHeight` (751890 ≠ 727890, plus `font-text`'s golden). The fixture's four geometric inputs
are genuinely pairwise distinct (30000/42000/18000/24000), all three bands carry an element with a
distinct string, and `e2` sits at a non-zero band-relative `y` (120). AC6's ban on the conservation
form holds — no sum-to-whole appears anywhere in the assertions.

**2. The developer's RP-9b disclosure — CONFIRMED CORRECT AND COMPLETE, not a defect.** I applied
the swap inside `ContentHeight`'s own expression
(`g.Height - g.MarginBottom - g.MarginTop - …`). It stays green — **and the produced digest is
byte-identical**, which is strictly stronger evidence than green tests: the mutation is a *semantic
no-op*, not an uncovered defect. Integer subtraction of a commutative sum is exact reassociation
(`a-b-c == a-c-b`) at these magnitudes with no overflow, so the two programs are extensionally
equal and **no observation of behaviour can distinguish them**; only a source-shape guard could,
and that would be an AST whitelist over a change that alters nothing. The developer did **not** stop
one step early. I probed the adjacent mutation that *is* observable — `ComposePage` carrying
`g.MarginBottom` into `Page.MarginTop` — and it reddens `TestComposePageCarriesPageAbsoluteContentUnchanged`,
all four Tm literals and the digest. The margins are individually observable where it matters
(`MarginTop` reaches the bytes through `flipY`; `MarginBottom` fixes the footer's distance from the
paper edge), and both directions are covered.

**3. `stage-rank` and D-000.16's corrected table — SOUND.** Ranks 5/6 swapped (`text` 5, `fontset`
6); the production caller measures **27 first-party edges over 9 packages, 92 files, 0 findings**.
Both arrows the guard exists for hold and both have retained violating fixtures:
`layout`(7) ↛ `pdf`(8) (`layout/violating_pdf_import.go`) and `expr`(3) ↛ `bind`(4)
(`expr/violating_bind_import.go`), plus an equal-rank fixture and an unranked-package fixture. All
three AC3 red-proofs fire with the recorded messages, both directions of the exact-set check
included. **The `target == stage` exclusion did not blind the guard**: with the skip disabled the
production scan reports **exactly one** additional finding —
`fontset/vendorboundary_test.go:17`, the external-test-package idiom — and nothing else
(FirstPartyImports 28 vs 27, so the vacuity guard's non-zero condition rests on 27 genuine
cross-stage edges, never on self-imports). I also probed the laundering risk directly: a
`package layout_test` file inside `internal/layout/` importing `internal/pdf` **is still caught**,
by both the lint rule and AC4's named guard. AC4's transitive BFS and its AD-5 quotation fire on
the same mutation.

**4. D-2.5.1's declared-list guard — the one-line-diff property HOLDS. Constructed.** A fifth
`//go:build matrix` file reddens `TestEpic2GateObligationsMatchTheDeclaredSet`; adding the single
line `"matrix-file: zz_fifth_matrix_test.go",` to `declaredEpic2GateObligations` returns it to
green; renaming a declared document reddens **both** directions ("declared … is NOT present in the
tree" *and* "is NOT in declaredEpic2GateObligations") plus `TestMatrixDocumentSlugsAreRegisteredInCI`.
The count-in-the-name is genuinely gone. **One evasion found — Finding 1 below.**

**5. AD-24 — `ContentHeight` is geometry-only and cannot consult its own elements. VERIFIED
STRUCTURALLY. No Blocker.** `PageGeometry` is a closed struct of eight `geom.Length` scalars with no
field through which an element or a measurement could arrive; `pageGeometryOf` reads only
`doc.doc.Page.*` and the two **declared** band heights. I probed the clause directly by giving
`ContentHeight` a second parameter (`elements []pagemodel.TextRun`) that changes its result — the
guard fails with *"ContentHeight must take exactly one parameter … AD-24's 'nothing negotiates' is
enforced by what it CANNOT be handed."* A second inline derivation inside `internal/layout` reddens
(RP-7); band-origin arithmetic restored in a package-`folio` function reddens (RP-6, naming
`documentBands` and both operands). `pageDimensions` now has exactly one caller
(`pageGeometryOf`), matching its comment; no `usableHeight` survives in package `folio`.
**No path exists by which content influences band geometry.**

**6. D-000.37's stale tripwire — DELETED, not obeyed.** `TestProvisionalBandOriginIsPinned` is gone
from `render_test.go`; the half-stale site-(2) instruction that would have moved goldens was not
followed. `textdoc.go`'s comments now state `TextRun.X/Y` are **page-absolute and permanent**, cite
`flipY` and D-1.8.10 rather than a deleted test, and name **DW-15** for the baseline offset. A
repo-wide sweep for references to deleted/moved mechanisms (`TestProvisionalBandOriginIsPinned`,
`pdf.TextPage`, `pdf.TextRun`, `pdf.ShapedGlyph`, `pdf.ImagePlacement`) returns **zero** live hits.
The only surviving `AC28` mentions are `render.go:29`'s `defaultFontSizePt` label (which AC9
explicitly rules out of scope) and two unrelated `internal/template` uses of a different story's AC
numbering. **One newly-written remedy message does point at a symbol that does not exist — Finding 3.**

### Verified, and confirmed NOT findings

- **Types MOVED, not duplicated.** `TextPage`/`TextRun`/`ShapedGlyph`/`ImagePlacement` have exactly
  one declaration each, in `internal/pagemodel`; `grep` for `TextPage`, `pdf.TextRun` and
  `ResourceName` returns no live definition or use. `ImagePlacement.ResourceName` → `AssetKey` is
  carried through `imagedoc.go` including the error type. `EmbeddedFace`, `ImageXObject` and
  `CIDText` correctly stayed in `internal/pdf`. AC1's guard fires on a planted `ResourceDict` field
  (twice — the substring list overlaps deliberately) and its presence precondition counts exported
  types *with fields*.
- **Matrix legs written and registered but NOT run — correctly declared.** `three-band-page` is in
  **both** `matrixDocuments` and `matrix.yml`'s `docs="…"` line, with four
  `hash.<target>.three-band-page.txt` upload paths. `go vet -tags matrix` compiles the legs. **The
  legs have teeth**: `requireThreeBandPageUsesAllThreeBands` reads the produced bytes, demands
  exactly four distinct baselines, and requires each of the three bands to contribute at least one
  run against hand-derived PDF-user-space boundaries — a collapsed band fails every leg *before* any
  byte comparison. The README says *"Registered; legs deferred to the Epic 2 gate"* in those terms,
  claims no four-target agreement, and correctly labels the two-process check a determinism witness
  on one target (D-000.28 respected).
- **`fixtures/three-band-page/`** — sha256 `2315855a…ada6d04f` ✓, **54,445 bytes** ✓,
  `expected.json` self-consistent ✓. Its D-000.22 semantic acceptance step is present and its
  machine-checkable half is **not deferred**: well-formedness, non-blankness, `/FontFile2` presence,
  four Tm literals each paired with its band's text decoded back through the document's *own*
  `/ToUnicode` CMap, exactly-three-bands-contributing, and a two-process byte-identity check — all
  run **before** the digest comparison.
- **`git status --porcelain fixtures/`** reports **only** `?? fixtures/three-band-page/`. All six
  pre-existing goldens re-hashed and match their `expected.json`;
  `fixtures/shaped-text/expected.pdf` is still **`5964aad0…c92e00f`**, so D-2.3.5's pending human
  sign-off is intact. **No Blocker.**
- **DW-15** is present in `deferred-work.md`, `textdoc.go:672`'s comment resolves to it, and the
  baseline offset itself (`flipY(page.Height, page.MarginTop, run.Y, run.FontSize)`,
  `textdoc.go:689`) is **untouched**.
- **The gate owes four things; there is no fifth.** Observed obligation set equals
  `declaredEpic2GateObligations` exactly (4 matrix-tagged files + 7 matrix documents), and the
  semantic buckets remain: the four-target legs, D-2.3.5's reading sign-off, D-2.4.3's break
  sign-off, and `three-band-page`'s deferred legs.
- **AC10's regression guards are green *and unmodified*.** `render_arch_test.go`,
  `internal/pdf/flip_test.go`, `internal/arch_test.go`, `matrix_registration_test.go`,
  `gomod_test.go` and `internal/template/`'s tests carry no diff. `byte_neutrality_test.go`'s
  `TestStory23aMovedNoGoldenDigest` is unchanged (the diff touches only the (j) guard, which D-2.5.1
  ordered replaced). `SOURCE_DATE_EPOCH` appears only inside the retained violating fixture.
- **No lint registry was bypassed.** `lint/` has no rule-dispatch registry — `map-range`,
  `forbidden-imports` and `no-float-typed-value` are all invoked solely by their own production and
  fixture callers, and `stage-rank` follows that precedent exactly. `MANIFEST.md` is a licence
  manifest, not a rule list.

---

### Finding 1: A new matrix document can be registered invisibly to *both* obligation guards

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/byte_neutrality_test.go:255` (the `slug:` regex) and
  `folio-go/matrix_registration_test.go:39` (the same regex, pre-existing)
- **Observation**: `TestEpic2GateObligationsMatchTheDeclaredSet` derives the *observed* document
  half of the obligation set by regex over `matrix_test.go`'s **source text**:
  `(?m)^\s*slug:\s*"([^"]+)"`. It requires `slug:` to begin a line. **Constructed and confirmed**: a
  seventh entry appended to `matrixDocuments` as a single-line composite literal —
  `{label: "evader", slug: "zz-evader", capture: captureThreeBandRender, fixtureRelPath: […], requireFontFile2: true},`
  — is **gofmt-clean** (`gofmt -l` reports nothing), **compiles under `-tags matrix`**
  (`go vet -tags matrix` passes), and is **invisible to both guards**: the obligation test passes
  with no line added to `declaredEpic2GateObligations`, *and*
  `TestMatrixDocumentSlugsAreRegisteredInCI` passes with no entry in `matrix.yml`. Both tests were
  green.
- **Impact**: A recorded document ships with **no matrix.yml registration and no gate obligation** —
  precisely *"a matrix leg nobody compares, reported as green"*, the drift both guards name in their
  own failure text. Under **D-000.36** (ruled for this story: *a remedy for a vacuity is itself
  subject to vacuity*) this is the hazard D-2.5.1's replacement guard inherited from the extractor
  it reused. It also sits against the repo's own stated discipline, quoted verbatim inside
  `stagerank.go`: *"never a regex, never the literal text."*
- **Suggested Resolution**: Replace the regex with a `go/parser` read of `matrixDocuments`'
  composite literal (`matrix_test.go` is `//go:build matrix`, so source parsing is still required —
  but an AST walk over the `KeyValueExpr` for `slug` is spelling-independent). Add a vacuity check
  that the parsed entry count equals the number of `matrixDocument` literals found, so a shape the
  walker cannot read fails rather than shrinking the set. Fix `matrix_registration_test.go` in the
  same change, or the two guards remain jointly evadable.
- **Related AC**: AC10(i), AC10(j), AC7 (Registration), D-2.5.1

### Finding 2: `pagemodel.ShapedGlyph.CID` is not always a glyph id — its second-meaning allocation is defined by a PDF encoding

- **Severity**: Major
- **Category**: AC Conformance
- **Location**: `folio-go/internal/pagemodel/pagemodel.go:79`–`:85`; allocation at
  `folio-go/render.go:833`–`:872`
- **Observation**: This is the reviewer's ruling on the one judgment the developer flagged. **The
  name passes AC1 as written, and I do not fault the developer's reading**: AC1's ban list is closed
  and enumerated ("a PDF object reference, a resource dictionary, a content-stream operator, or a
  font *program*"), a CID is on none of them, and AD-5 names "glyph ids" as a page-model concept.
  **But the placement carries a leak the developer's reading did not reach.** `CID` is only
  *sometimes* a glyph id. In the base block `cid = newGID`, the subset glyph id — legitimate. In the
  `default` block a **second, synthetic CID is minted** for a glyph that carries a different source
  text, and its ceiling is stated in the code itself: *"Identity-H's CID is TWO BYTES … exceeds
  Identity-H's two-byte CID ceiling of 65535"*. That branch exists **only** because PDF's
  `/ToUnicode` CMap maps one CID to one text. The `CID → GID` map that would let anyone interpret
  those extras (`state.extras`, `pdf.CIDText`) is deliberately kept **out** of the page model. So a
  PNG/SVG/HTML renderer handed a `pagemodel.TextRun` **cannot resolve `CID` back to a glyph** — which
  is exactly the capability AD-5 exists to preserve ("that absence is what keeps PNG/SVG/HTML
  renderers possible later").
- **Impact**: AD-5 is this story's **primary invariant**, and the story's headline claim is that the
  page model knows nothing about PDF. It now contains a field whose value range and second-meaning
  semantics are defined by a PDF encoding, and whose interpretation requires a table that is not in
  the page model. AC1's substring guard cannot see this — `"cid"` is not in `pdfConceptSubstrings`,
  and adding it would be the wrong fix, since the base-block value *is* a legitimate subset glyph id.
  **Nothing regressed**: the allocation is Story 2.3's D-2.3.2, unchanged; the type move merely
  relocated it into `pagemodel` and thereby exposed it.
- **Suggested Resolution**: **Do not re-architect `buildShapedPDFRuns` in this story** — neither AC1
  nor Task 4 asks for it, and the developer was right not to. Record it as a deferred-work entry
  (sibling to DW-15) stating the property precisely: *`pagemodel.ShapedGlyph.CID` is a subset glyph
  id in the base block and a PDF-`/ToUnicode`-driven synthetic identifier in the extras block; the
  `CID→GID` map required to interpret the latter is not carried in the page model, so a non-PDF
  renderer cannot consume `Glyphs` today.* Then obtain a ruling on the shape: either the page model
  carries `GlyphID` plus a separate text association (renderer-neutral), or AD-5's "glyph ids"
  clause is amended to admit an encoding-scoped identifier. **The first non-PDF renderer is the
  natural forcing story** — but the record must exist before then, because the window in which this
  is cheap to see closes as soon as more producers write the field.
- **Related AC**: AC1, AD-5

### Finding 3: The `stage-rank` unranked-package remedy names a symbol that does not exist

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `lint/internal/rules/stagerank.go:200`
- **Observation**: The fail-safe finding instructs the reader to *"rank it deliberately in
  `lint/internal/rules/stagerank.go`'s **stageRanks** table"*. The identifier is **`stageRankTable`**
  (declared at `:57`, read at `:98` and `:269`). `stageRanks` appears nowhere in the repository.
- **Impact**: **D-000.37 is ruled for this story** and its rule is that *"a tripwire's failure
  message is executable by a human"* and *"the remedy text is maintained as code, not as prose."*
  This message is newly written in this same commit and its remedy names a nonexistent symbol on
  first delivery. The file path is correct, so the cost is small — but this is the exact failure
  class the story's own governing ruling exists to close, and `stagerank_test.go`'s
  `TestStageRankMessageNamesAD5sArrow` asserts message content for the `layout → pdf` finding while
  the unranked branch's message has no such assertion.
- **Suggested Resolution**: One-word fix: `stageRanks` → `stageRankTable`. Consider extending
  `TestStageRankMessageNamesAD5sArrow` (or adding a sibling) to assert the unranked-package
  message names `stageRankTable`, so the remedy pointer is held by a test the way the `layout → pdf`
  message already is.
- **Related AC**: AC3, D-000.37

### Finding 4: The Delivery Log's "Proposed — NOT written" item 1 is now stale

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `_bmad-output/implementation-artifacts/2-5-compose-a-page-from-three-bands.md:1117`–`:1126`
- **Observation**: Item 1 states *"`DW-15` is cited by this story but is not present in
  `deferred-work.md` (its last entry is `DW-14`)"* and proposes the entry text. **DW-15 now exists**
  (`deferred-work.md:493`), added by the reviewing owner after the developer correctly flagged the
  dangling citation. The developer's statement was true when written; it is no longer.
- **Impact**: A record that reads as an open action when the action is closed. No code effect.
- **Suggested Resolution**: The finisher should reword item 1 to *"DW-15 was proposed by this story
  and has since been recorded by the owner at `deferred-work.md:493`; the `textdoc.go:672` citation
  resolves."* Items 2 and 3 remain accurate and open.
- **Related AC**: AC9, D-2.5-Q3 / D-000.37

---

### AC-by-AC disposition (reviewer's own verification)

| AC | verdict | how verified |
|---|---|---|
| AC1 | **satisfied**, with Finding 2 raised against the field it cannot see | single declarations confirmed; `ResourceDict` red-proof re-run (fires twice); presence precondition counts exported types *with fields*; `EmbeddedFace`/`ImageXObject`/`CIDText` stayed |
| AC2 | **satisfied** | RP-6 re-run; `documentBands` asks rather than derives; `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt` green **unmodified** |
| AC3 | **satisfied** | all three red-proofs re-run with recorded messages; exact-set both directions; vacuity guard reads the checker's own stats and fires on injection |
| AC4 | **satisfied** | fires on the same mutation as AC3(1), carries the AD-5 sentence; transitive BFS; laundering through `package layout_test` still caught |
| AC5 | **satisfied** | RP-7 re-run; my own two-parameter probe rejected by the guard; declared once, called from a non-test site |
| AC6 | **satisfied**; RP-9b independently confirmed a semantic no-op, **not** a gap | all four red-proofs re-run; conservation form absent; strict-increase asserted with a degenerate control |
| AC7 | **satisfied** | four distinct inputs, three bands populated, non-zero band-relative `y`, README complete, semantic acceptance non-deferred |
| AC8 | **satisfied** | all six digests re-hashed; `shaped-text` still `5964aad0…c92e00f`; `git status --porcelain fixtures/` = `?? fixtures/three-band-page/` only |
| AC9 | **satisfied** | pin deleted not obeyed; comments corrected; DW-15 resolves; baseline offset untouched; `defaultFontSizePt` left alone; zero stale mechanism references |
| AC10 | **satisfied** | (a)–(i) green and carry no diff; (j) replaced per D-2.5.1 and re-proved in both directions — see Finding 1 for its one evasion |
| AC11 | **satisfied** | 475·1 / 304·1; the one failure is `P6g: got 7, need >=20` with baseline-identical stats; no corpus change, no floor lowered |

---

## Finisher — finding resolutions and close-out

**Finished by:** bmad-story-finisher · **Date:** 2026-08-24 · **Baseline:** `213d98c`, work uncommitted
**Triage:** **3 FIX · 0 DISMISS · 1 DEFER** over 0 Blockers / 2 Majors / 1 Minor / 1 Nit.

Mutations were applied by hand against a pre-work `rsync` snapshot and reverted by `cp` from it —
never `git checkout`, never `git stash`, because `internal/layout/`, `internal/pagemodel/`,
`testdata/lint/stage-rank/` and both `lint/` rule files are untracked at this commit and a checkout
would have deleted them. Every revert was verified with `/usr/bin/diff` against the snapshot, not by
a green re-run.

### Triage

| # | severity | decision | rationale |
|---|---|---|---|
| 1 — a new matrix document is invisible to both obligation guards | Major | **FIX** | Reproduced before touching anything: the single-line literal is `gofmt`-clean, passes `go vet -tags matrix`, and **both guards reported PASS**. Root cause fixed as stated — the shared regex, not the regex's pattern |
| 2 — `pagemodel.ShapedGlyph.CID` is not always a glyph id | Major | **DEFER → DW-16** | Confirmed at the source; nothing regressed (Story 2.3's D-2.3.2, merely relocated). Neither AC1 nor Task 4 asks for the re-architecture, and the ruling needed is AD-5's, not a finisher's. Recorded with a named owner; the overstating doc comment corrected in place |
| 3 — the unranked-package remedy names `stageRanks` | Minor | **FIX** | `stageRanks` occurs exactly once in the repository, at the site itself. D-000.37 is ruled for this story, so a remedy naming a symbol that does not exist is not cosmetic |
| 4 — the Delivery Log's "Proposed" item 1 is stale | Nit | **FIX** | `DW-15` exists at `deferred-work.md:493`. Rewritten in place, and the sweep widened to items 2 and 3 |

**Nothing was dismissed.** All four findings were reproduced from the tree before being acted on;
none rested on a claim that did not hold.

### Finding 1 — the fix, and why it is not a better regex

The reviewer's diagnosis is adopted whole: `TestEpic2GateObligationsMatchTheDeclaredSet` and
`TestMatrixDocumentSlugsAreRegisteredInCI` shared one `(?m)^\s*slug:\s*"([^"]+)"`, so they were
**one guard wearing two names**. Widening the pattern would have left exactly that structure standing.

**`folio-go/matrixdocs_source_test.go`** now declares `MatrixDocumentSlugsFromSource`, the single
reader both guards call. It parses `matrix_test.go` with `go/parser` and walks the `matrixDocuments`
composite literal, so line breaks, alignment and field order are invisible to it *because they are
invisible to the language*. `matrix_test.go` is `//go:build matrix`, so reading source is still
required — but the Go grammar, unlike a regex, admits no second spelling of a composite literal.
This is the discipline `stagerank.go` already states in its own words — *"never a regex, never the
literal text"* — applied to the one place in the module that had not adopted it.

Per **D-000.36** — *a remedy for a vacuity is itself subject to vacuity* — an entry the walker cannot
read is an **error, never a skip**. Skipping would have moved the evader from *"not matched by the
regex"* to *"not understood by the walker"* with both guards green again, which is the same defect
one spelling further along. Both call sites additionally assert their own **N-of-N witness**
(`len(slugs) == elements`) rather than taking the reader's word for its completeness, and both now
log it: *"7 matrix documents read from 7 matrixDocuments literal entries"*.

The reader is exported from `package folio` because `matrix_registration_test.go` is an internal test
file while `byte_neutrality_test.go` is `package folio_test`; it is test-only source, so the module's
public API is unchanged (`go list -m all` still reports exactly two modules).

#### Red-proofs — run against the remedy, the single-line form FIRST

Each was applied by hand, run, and reverted with a verified `/usr/bin/diff`. Every shape below is
`gofmt`-clean and compiles under `-tags matrix` (`go vet -tags matrix` exit 0) except RP-M1e-parse,
so none is caught by the toolchain.

| # | mutation | before the fix | after the fix |
|---|---|---|---|
| **RP-M1a** | **the exact single-line literal that escaped** — `{label: "evader", slug: "zz-evader", capture: captureThreeBandRender, fixtureRelPath: […], requireFontFile2: true},` | **BOTH guards PASS** (measured) | **BOTH FAIL.** Registration: *"matrix_test.go: […] zz-evader / matrix.yml: […]"*; obligation: *"`matrix-document: zz-evader` is an Epic 2 gate obligation and is NOT in declaredEpic2GateObligations"*. The witness moved **7 → 8**, so the entry was genuinely read |
| RP-M1b | an entry with **no `slug:` field** | — | both FAIL: *"matrixDocuments[7] … declares no non-empty `slug:` field"* |
| RP-M1c | a **computed** slug (`"zz-" + "evader"`) | — | both FAIL: *"… is not a string literal — a slug computed at run time cannot be compared against .github/workflows/matrix.yml, which is text"* |
| RP-M1d | **two entries under one slug** | — | both FAIL: *"repeats slug \"three-band-page\", already used by entry 6 — … the second is a leg nobody compares"* |
| RP-M1e | an element that is an **identifier**, not a literal | — | both FAIL: *"matrixDocuments[7] … is a \*ast.Ident, not a composite literal — a document it cannot read must FAIL rather than vanish from the set (D-000.36)"* |

RP-M1a is the one that matters: it is the shape the review constructed, and it is the shape the
remedy was proved against.

**A first attempt at RP-M1e was invalid and is recorded as such.** The injected Go was
syntactically malformed, so it exercised the reader's *parse-error* path, not its identifier path —
`go vet -tags matrix` exited 1, which is the tell. It was rewritten as valid Go (a package-level
`var sneakyDoc = matrixDocument{…}` referenced as an element), which compiles cleanly, and only then
did it prove the intended branch.

#### The remedy's own red-proofs are held by a test, not by this transcript

`TestMatrixDocumentSlugsFromSourceReadsShapesTheRegexMissed` (11 subtests) keeps every shape above
running in the ordinary suite, plus an empty list, an absent declaration and a doubled declaration.
The single-line case asserts **by name and in order** (`[kept-slug zz-evader]`), not by count — a
count of 2 would have been satisfied by the multi-line entry alone.

That test was then itself mutated, **one mutation per mechanism**:

- **R1 — mechanism 1 neutered** (per-element slug error → silent skip): **2 subtests redden.** The
  failure text shows mechanism 2 catching it (*"read 1 slugs from 2 matrixDocuments entries"*), so
  the set never silently shrinks. Defence in depth, demonstrated rather than assumed.
- **R2 — mechanism 2 neutered alone** (the final `len(slugs) != elements` check): **nothing
  reddens — 0 subtests.** Reported as measured: that check has **no independent teeth** while
  mechanism 1 stands. It is belt-and-braces, and is not credited with a red-proof it does not have.
- **R3 — both neutered:** the reader returns a **shorter list with no error**, and the subtests fail
  on *"read 1 slugs from 2 entries with NO error — an unreadable entry that shrinks the set is the
  defect this reader exists to prevent"*. The pair is load-bearing, and the test's "NO error" leg is
  reachable.

Post-revert, the suite is green again and the file is byte-identical to its pre-mutation state.

**D-000.34 sweep.** No test's discriminating power depended on the regex. The regex had exactly two
readers, both replaced; `regexp` is now unused in `byte_neutrality_test.go` and its import removed,
while `matrix_registration_test.go` retains it for `matrix.yml`'s `docs="…"` line — deliberately, and
the file's header comment now says why: **matrix.yml is YAML holding a shell line, so there is no
grammar to parse**, and a missing match there is reported as a failure rather than as an empty list.

### Finding 2 — deferred as DW-16, with the doc comment corrected

`DW-16` is recorded in `_bmad-output/implementation-artifacts/deferred-work.md`. **Owner: the
engineering lead to rule on the shape; the fix lands in the first non-PDF renderer story**, its
natural forcing function.

**What a non-PDF consumer cannot do today** (the entry states this precisely): in
`buildShapedPDFRuns`' base block `CID` **is** the subset glyph id; in the `default` block a
**synthetic** identifier is minted for a glyph carrying a different source text
(`cid = uint16(sub.NumGlyphs + len(state.extras))`), which is an index into `state.extras` and not a
glyph id in any font — its ceiling stated by the code as *"Identity-H's two-byte CID ceiling of
65535"*. The `CID → GID` table that resolves it (`state.extras`, `pdf.CIDText`) is correctly kept
out of the page model. So a PNG/SVG/HTML renderer handed a `pagemodel.TextRun` **cannot resolve
`Glyphs` back to glyphs, and cannot tell at the type which of the two kinds of value it holds.**

**What would fix it** — a ruling, not a finisher's choice: either `ShapedGlyph` carries a true
`GlyphID` plus a separate text association with the allocation moved beside `pdf.CIDText`
(renderer-neutral, strictly more work), or AD-5's "glyph ids" clause is amended to admit an
encoding-scoped identifier and the field renamed to say so. The entry records that the window closes
as **more producers write the field** — today there is exactly one.

**Not re-architected**, as instructed. One change was made beyond the record: `ShapedGlyph.CID`'s doc
comment asserted *"It is a glyph identifier — AD-5's 'glyph ids'"*, which is **false for the extras
branch**. Leaving a false statement compiled into the package while filing a note elsewhere is the
failure D-000.28 names, so the comment now states both kinds of value and points at DW-16. This is a
comment-only edit: no identifier changed, so AC1's guard — which walks type names, field names and
field types, never comments — is unaffected, and no byte of any golden moves.

`cid` was **not** added to `pdfConceptSubstrings`; DW-16 records that as the wrong fix, since the
base-block value is a legitimate subset glyph id.

### Finding 3 — fixed, and the remedy pointer is now held by a test

`lint/internal/rules/stagerank.go`: `stageRanks` → **`stageRankTable`**. Verified before the edit
that `stageRanks` occurred exactly once repo-wide (the message itself), and that `stageRankTable` is
the declared name.

The reviewer's optional half is taken up, because the finding's own cause was that this branch had
**no message assertion at all** while the `layout → pdf` branch did.
`TestStageRankUnrankedMessageNamesASymbolThatExists` asserts the unranked message names the package,
the file, `stageRankTable`, *"never a pass"* and the known-ranks list — **and then reads
`stagerank.go` and requires that the identifier it names is actually declared there.** A string
assertion alone would only have moved the stale name from the message into the test.

Red-proved, **one mutation per mechanism**, both reverted byte-identically:

- **RP-M3a** (message reverted to `stageRanks`): FAILS — *"unranked-package finding message does not
  contain \"stageRankTable\""*.
- **RP-M3b** (table renamed to `stageRankRoster`, message left naming `stageRankTable`): FAILS —
  *"the remedy names \"stageRankTable\", but …/stagerank.go declares no such variable — a tripwire
  whose remedy points at a symbol that does not exist is not executable by a human (D-000.37)"*.

**RP-M3b's first attempt silently did not apply** and is recorded as such: BSD `sed` on macOS does
not support `\b`, so the rename matched nothing and the test's green was measuring an unmutated file.
It was caught by printing the mutation's own preconditions (`var stageRankTable declared? 1`, want 0)
rather than by trusting the exit code, and redone in Python. A red-proof whose mutation never landed
is not a red-proof.

### Finding 4 — corrected, and the sweep widened

Item 1 rewritten in place; it also now records the **binding sequencing obligation** the recorded
DW-15 carries and the proposal did not — the fix lands *before* D-2.3.5's Thai reading sign-off, not
after. Items 2 and 3 were re-read rather than assumed still true: item 2's parenthetical was stale
against the finisher's own `review → done` edit and is corrected; item 3 is widened to state that the
finisher wrote no decision-log entry either, and names RP-9b as the strongest remaining candidate.

### Not reopened

Verified present rather than re-litigated: `git status --porcelain fixtures/` reports **only**
`?? fixtures/three-band-page/`; `fixtures/shaped-text/expected.pdf` still hashes
**`5964aad0e696010c6e3f34a48d0775af6ae527a6cbe2f5c6319158f43c92e00f`**, so D-2.3.5's pending human
sign-off is intact; the baseline offset in `internal/pdf/textdoc.go` is untouched (DW-15, not this
story); `go list -m all` is exactly two modules; no `SOURCE_DATE_EPOCH` literal occurs in any `.go`
under `folio-go/` outside `testdata/`. RP-9b, the `stage-rank` self-import skip, the type move, the
`AssetKey` rename and the fixture's digest and size were accepted as the review measured them.

### The Epic 2 gate — still exactly four, and one flag

The gate owes **four** things, unchanged by this close-out: the four-target legs, D-2.3.5's reading
sign-off, D-2.4.3's break sign-off, and `three-band-page`'s deferred legs. No fifth was added —
`declaredEpic2GateObligations` is untouched, and the guard reading it now reports a 7-of-7 witness.

**`three-band-page`'s matrix legs are WRITTEN AND REGISTERED, BUT NOT RUN.** It is in both
`matrixDocuments` and `matrix.yml`'s `docs="…"` line with four `hash.<target>.three-band-page.txt`
upload paths; the legs compile under `-tags matrix` and have teeth
(`requireThreeBandPageUsesAllThreeBands` demands four distinct baselines and a contribution from each
of the three bands before any byte comparison). **The D-000.4 per-story override was DECLINED, and
the decline reason is the criterion itself:** an override is warranted when a story introduces a new
**source of cross-target divergence** — float arithmetic, a vendor call, a compressor, a new
dependency — *not* merely because it records a new golden. Story 2.5 is **integer `geom.Length` band
arithmetic** and introduces no such source. The four legs are the Epic 2 boundary gate's.

**Flagged for the gate, not changed here:** `sprint-status.yaml` still reads **`epic-2: backlog`**
while `2-1`…`2-5` are all `done`. The epic key is the gate's to flip; this story edited only its own.

### Finisher — measured gates

Every row from a real run at the finished tree. Invocation, verbatim, from each module's directory:

```
CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...
```

with `go` = go1.26.0 darwin/arm64. All-occurrences counts are `grep -c -- '--- PASS'` /
`'--- FAIL'`; top-level counts are `grep -c '^--- PASS'` / `'^--- FAIL'` (column 0).

| scope | at handover (reviewer) | **at close-out** |
|---|---|---|
| `folio-go/` all-occurrences | 475 · 1 | **487 · 1** |
| `folio-go/` top-level | 304 · 1 | **305 · 1** |
| `lint/` all-occurrences | 84 · 0 | **85 · 0** |
| `lint/` top-level | 46 · 0 | **47 · 0** |
| modules (`GOWORK=off go list -m all`) | 2 | **2** |
| `go vet ./...` (both modules) | clean | **clean** |
| `go vet -tags matrix ./...` | clean | **clean** |
| `gofmt -l` (both modules, excl. `testdata/`) | clean | **clean** |
| `SOURCE_DATE_EPOCH` literal in `.go` under `folio-go/` (excl. `testdata/`) | 0 | **0** |
| `git status --porcelain fixtures/` | `?? fixtures/three-band-page/` | **`?? fixtures/three-band-page/`** |

**The delta accounts exactly, and is entirely the two new guards:**

- `folio-go` **+12 all-occurrences / +1 top-level** — `TestMatrixDocumentSlugsFromSourceReadsShapesTheRegexMissed`:
  1 top-level parent + **11** subtests (2 positive-read cases, 8 unreadable-shape cases, 1 read of
  the real `matrix_test.go`).
- `lint` **+1 / +1** — `TestStageRankUnrankedMessageNamesASymbolThatExists`, which has no subtests.

**The single failure is the sanctioned one**, and it is unchanged: `TestCorpusMeetsP6ExerciseFloors`,
*"P6g (opaque names) floor not met: got 7, need >=20"*, stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — **byte-identical to the `213d98c` baseline
measured before any finisher edit**, satisfying Story 2.4's AC5 and D-000.17. **No second failure at
any point.** Nothing was added to the corpus and no floor was lowered.

### File List — finisher delta

**New**

- `folio-go/matrixdocs_source_test.go` — `MatrixDocumentSlugsFromSource` and its 11-subtest guard

**Modified**

- `folio-go/byte_neutrality_test.go` — calls the shared reader; `regexp` import removed; N-of-N witness
- `folio-go/matrix_registration_test.go` — calls the shared reader; header comment corrected; witness widened
- `folio-go/internal/pagemodel/pagemodel.go` — `ShapedGlyph.CID` doc comment corrected, cites DW-16 (comment only)
- `lint/internal/rules/stagerank.go` — `stageRanks` → `stageRankTable` in the unranked-package remedy
- `lint/internal/rules/stagerank_test.go` — `TestStageRankUnrankedMessageNamesASymbolThatExists`; `os` import
- `_bmad-output/implementation-artifacts/deferred-work.md` — **DW-16** recorded
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — `2-5` → `done` (this story's key only)
- `_bmad-output/implementation-artifacts/2-5-compose-a-page-from-three-bands.md` — opener rewritten to
  what shipped; status → `done`; Finding 4 corrected and the sweep widened; this section

**Unchanged, deliberately:** `folio-go/matrix_test.go` (every mutation reverted, `/usr/bin/diff`
verified), `declaredEpic2GateObligations`, all six pre-existing goldens, `internal/pdf/textdoc.go`'s
baseline offset, and `sprint-status.yaml`'s `epic-2` key.

### Change Log

| date | change |
|---|---|
| 2026-08-24 | Finisher close-out. Review findings triaged **3 FIX / 0 DISMISS / 1 DEFER**. Major 1: both matrix-obligation guards re-pointed at a single `go/parser` reader of `matrixDocuments`, replacing the shared regex that made them one guard wearing two names — red-proved against the single-line literal specifically, plus four further evasion shapes, and the remedy's own two mechanisms mutated separately. Major 2: deferred as **DW-16** with a named owner; `ShapedGlyph.CID`'s overstating doc comment corrected in place, no re-architecture. Minor: `stageRanks` → `stageRankTable`, with a new test asserting the remedy names a symbol that exists. Nit: the stale DW-15 item rewritten and the sweep widened. `folio-go` 487·1 (305·1 top-level), `lint` 85·0 (47·0); the one failure is `TestCorpusMeetsP6ExerciseFloors`, baseline-identical. Status → `done`. |
