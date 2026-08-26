# Epic 3 — Boundary Gate

**Run:** 2026-08-26 · **HEAD:** `ba24e52` (Story 3.7 finisher) · **Working tree:** clean, verified
byte-identical before and after every mutation in this run.
**Cadence:** D-000.4 (per-epic). Epic 3 shipped eight stories — **3.1, 3.1a, 3.2, 3.3, 3.4, 3.5, 3.6,
3.7** — in twelve commits, `a20f60e`…`ba24e52`.

**Scope of this file.** The **cross-target hash matrix was run separately by the orchestrator and
PASSED**; it is not re-run here and is noted only where it discharges an inherited obligation. This
file carries the **four non-matrix audits** plus one item the standing-red rule (D-000.70) forced
into existence. Mirrors `epic-1-boundary-gate.md`'s five-item shape. Unlike
`epic-2-boundary-gate.md`, this is **not** an accumulator — no `epic-3-boundary-gate.md` existed
before this run (verified: `ls _bmad-output/implementation-artifacts/epic-3*` → no match).

## Verdict — **PASS, with three documentation falsehoods that must be corrected in this gate's commit**

The four audits come back **clean on the properties that matter**: no deferred item owned by an
Epic 3 story is still open; no spine invariant's text was invalidated in its rule body; the public
surface leaks nothing internal and the Epic 1 opaque-handle ruling has landed; every Epic 3 story
carries a conforming `## In plain terms` opener. Four Epic 3 deferral retirements were **re-proved by
mutation in this run** rather than accepted on their own record.

What is **not** clean is the prose: **three shipped documents now state things that Epic 3 made
false** — `ARCHITECTURE-SPINE.md` in two places, `folio-go/README.md` in three. These are the
narrated-vs-artifact drift class this program keeps finding, and they are the only findings I hold
`epic-3: done` for, because each is a one-line edit and each misdirects the next reader.

**Method note.** Every measurement below was written to a file and read back, not piped through a
grep (D-000.12, established at the Epic 1 gate). Every mutation ran in a `git worktree` at
`/private/tmp/folio-gate-wt`, removed and pruned at the end; `git status --porcelain` on the main
tree is **empty** and `git rev-parse HEAD` is still `ba24e52`.

---

## Item 1 — Deferred-work sweep · **PASS on Epic 3's own obligations; TWO owners are spent events**

`deferred-work.md` holds **20 entries**. Two sit under `## Done` (DW-1, DW-15). **Seven are marked
RETIRED / DISCHARGED in place but still sit physically under the `## Open` heading** — DW-5, DW-6,
DW-8, DW-9, DW-10, DW-18, DW-19. **Eleven are genuinely open**: DW-2, DW-3, DW-4, DW-7, DW-11, DW-12,
DW-13, DW-14, DW-16, DW-17, DW-20.

### The headline: **no open DW item is owned by an Epic 3 story.** Zero misses.

Every entry whose owner named an Epic 3 story was discharged inside that story:

| Entry | Owner named | Discharged at | Re-proved in this run |
|---|---|---|---|
| DW-5 | Story 3.2 (backstop 3.7) | Story 3.2 | **yes — M5** |
| DW-6 | whichever story creates `internal/diag` | Story 3.6 | **yes — M1, M2** |
| DW-8 | Story 3.2 | Story 3.2 | **yes — M3, M4** |
| DW-10 | Story 3.7 | Story 3.7 | **yes — M6** |
| DW-17 (3.7's third) | Story 3.7 | Story 3.7 | its own negative control, cited |
| DW-18 | whoever next touches `folio.Severity` | Story 3.6 | **yes — see Item 3** |
| DW-19 | whoever next touches `ResolveAssets` | Story 3.6 | `lint` 112/0, measured |

**The retirements were re-proved by mutation, not read.** Per D-000.9's extension — a record naming
tests that cannot fail is the same defect as a guard that cannot fail — four retirement claims were
put under a discriminating mutation in a throwaway worktree:

| # | mutation | result |
|---|---|---|
| M1 | drop `CodeTableFooterSourceUnresolved` from `allCodes` (`internal/diag/diag.go:150`) | **RED** — `TestRegistryIsAdditiveOnly` at `diag_test.go:70` and `:82` |
| M2 | `CodeTableFooterSourceForbidden` string → `…_FORBIDDEN_X` (`diag.go:85`) | **RED** — `diag_test.go:67`, `:70`, `:77` (three independent arms) |
| M3 | add a second `type Decimal` in `internal/geom` | **RED** — `TestExactlyOneDecimalDeclarationInTheModule`, `expr_arch_test.go:702`, naming both packages |
| M4 | add `parseBindingPath` / `isValidIdent` in `internal/geom` | **RED** — `TestParseBindingPathAndIsValidIdentAreAbsent`, `expr_arch_test.go:835`, naming file:line |
| M5 | `if !derivable {` → `if !derivable && false {` (`folio_expr_validate.go:325`) | **RED** — `TestParseTemplateRejectsNonDerivableFooterBind`, and **nothing else** in the module |
| M6 | `raw := getenv("SOURCE_DATE_EPOCH")` → `raw := ""` (`cmd/folio/main.go:362`) | **RED** — `TestRenderReadsSourceDateEpochFromEnvironment`, `TestRenderSourceDateEpochValueIsHonoured`, `TestSourceDateEpochFillsAbsentButNeverOverwritesSupplied` |

Six mutations, six reds, each naming the test its DW entry claims. `git status --porcelain` in the
worktree was empty after each restore. **DW-6's guard also passes D-000.68's anchor test**: `codePins`
(`internal/diag/diag_test.go`) is a **literal the test owns**, and `Code`'s string-ness is a
compile-time assertion (`diag.go:95`, `var _ Code = Code(...)`) rather than a test that cannot fail.

### FINDING 1.1 — **DW-14's owner is a spent event.** Carried forward, must be re-owned.

`deferred-work.md:661` reads: *"Owner: the Epic 2 boundary gate, or Story 2.4 if its corpora reach the
limit first."* **Both are past.** Story 2.4 measured it and did not trigger; the Epic 2 gate **ran**
(`e7f3f9c`), re-measured DW-14 at its section 8, wrote *"The fix stays the gate's"* — and closed
without naming **which** gate. DW-14 therefore has no live owner.

**Re-measured at HEAD with the project's own instrument** (the `(\d+)\s+beginbfchar(.*?)endbfchar`
regex `wrapped_text_fixture_test.go:377` uses, applied to all nine fixture PDFs):

```
multi-page 45 · wrapped-text 28,18,38 · page-count-20 32 · shaped-text 14,7,28
font-text 25 · three-band-page 17 · multi-script-fallback 4,1,1 · minimal-rect 0 · image-embed 0
```

Largest section **45** against the spec cap of 100 — **byte-for-byte the Epic 2 gate's table.**
**Epic 3 did not move DW-14**, because Epic 3 added no PDF fixture: `fixtures/hidden-image/` (Story
3.5) holds `expected.json` only. So the risk is unchanged and the entry is a **recording** defect,
not a live one — but the owner line must be amended to a live gate or a numbered story.

### FINDING 1.2 — **DW-16's forcing function does not exist in the roadmap.** Carried forward, loudly.

`deferred-work.md:744-745`: *"the fix lands in **the first non-PDF renderer story**, which is its
natural forcing function."* **Measured:** `grep -n "PNG\|SVG\|non-PDF\|HTML" _bmad-output/planning-artifacts/epics.md`
→ **0 matches.** No story in Epics 1–6 is a non-PDF renderer. The trigger names a story class that
was never scheduled.

This is the shape the program has now hit three times (DW-16, DW-19, and DW-14 above), and
Story 3.6's own file already named it — `3-6-fail-with-located-actionable-diagnostics.md:701` says
*"the unowned-deferral shape already recorded against DW-16"* — **and nobody re-owned it.** DW-16's
own text explains why waiting is expensive: *"the window in which this is cheap to see closes as soon
as more producers write the field"*; there is still exactly one (`buildShapedPDFRuns`).

### FINDING 1.3 — DW-13 was never scheduled. Carried forward.

DW-13's owner is *"orchestrator to schedule the sizing"*. **Measured:** zero mentions of `DW-13` in
`epic-2-boundary-gate.md` or in any of the eight Epic 3 story files. Nothing schedules it and nothing
will trip on it. Same class as 1.1 and 1.2, lower stakes.

### FINDING 1.4 — the `folio-go/v0.1.0` tag now gates **three** open entries, and still nobody owns it.

Epic 1's gate flagged DW-4 as *"the only entry whose owner is a person rather than a story"*. Since
then the pile on that one unowned event has grown: **DW-3** (publish the licence manifest), **DW-4**
itself (plus its new `documentDate` reserved-key ledger, appended at Story 3.7 —
`deferred-work.md:220-224`, recorded correctly), and now **DW-20**. **Epic 4 planning is the moment
DW-4 names for the owner decision, and Epic 4 planning is next.** Verified the tag has not been cut:
`git tag -l` → `pre-email-rewrite` only; `folio-go/version.go:9` → `Version = "0.0.0-dev"`.

### DW-20 — **recorded correctly.** Verified, not assumed.

`deferred-work.md:1014-1039`. Every load-bearing clause checked:

- Both guards it names exist: `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt`
  (`render_arch_test.go:278`), `TestValidateNeverReachesRenderOrInternalPDF` (`:365`), and
  `buildFolioCallGraph` (`:115`).
- Its cost claim — *"measured, zero methods in package `folio` at HEAD call into `internal/pdf`"* —
  **holds.** Package `folio` declares exactly **seven** methods (`diagnostic.go:75`,
  `render_error.go:63`, `:68`, `render.go:934`, `wrap.go:37`, `:60`, `:99`); no body carries a `pdf.`
  selector. The merge-by-name over-approximation therefore changes no verdict today, exactly as
  stated.
- Its trigger is real and unreached (tag check above).
- **Its owner chains to DW-4**, which the entry itself concedes. That is honest, and it is Finding 1.4.

### Live forcing functions, confirmed

`absenceChecks` (`lint/internal/rules/absences.go:92-98`) now holds **exactly one** row —
`absence-designer-project`, DW-2's remaining JS half, owner **Story 5.1**, a story that exists. The
list shrank 5→4→3→2→1 across the epics and each removal was by **replacement** (D-000.59), never by
deletion alone. DW-7 (Story 4.5), DW-11 (Story 4.7's golden report — half its owner window, "Epic 2's
later stories", is spent) and DW-17 (Stories 5.12, 6.6) all name stories that exist.

### Carried-forward hygiene note

Seven retired entries sitting under `## Open` means a reader scanning for open work must read seven
entries' bodies to discover they are closed. The file's header says *"mark items done in place"*, so
this is within convention — but `## Done` holds only two of the nine closed entries, which makes the
section headings themselves misleading. One-line fix: retitle or re-file. **Not blocking.**

---

## Item 2 — Spine drift · **TWO Epic 3 falsehoods; D-000.6's constraint (4) is CLEAN**

### D-000.6 — clean, measured per commit

Only **two** Epic 3 commits touched `ARCHITECTURE-SPINE.md`. Counts taken from each blob
(`grep -cE '^### AD-'`, `'^- \*\*Binds:\*\*'`, `'^- \*\*Prevents:\*\*'`):

| Commit | AD- | Binds | Prevents | lines |
|---|---|---|---|---|
| `5968c5d` (Story 2.4, last pre-Epic-3) | 26 | 26 | 26 | 685 |
| `ae9e551` Story 3.4 | 26 | 26 | 26 | 690 |
| `c1c977f` Story 3.6 | 26 | 26 | 26 | 692 |

`git diff 5968c5d..HEAD -- <spine> | grep -E '^[+-].*\*\*(Binds|Prevents):\*\*'` → **no output.**
**No Epic 3 amendment added, removed or altered any invariant's `Binds` or `Prevents` line.** The
two amendments are exactly the two expected: AD-12's Japanese disclosure paragraph
(`ARCHITECTURE-SPINE.md:282-286`, +5/−0) and the `diag/` source-tree line (`:613-615`, +3/−1). Stories
3.1, 3.1a, 3.2, 3.3, 3.5 and 3.7 did not touch the spine at all.

### Invariant bodies AD-1…AD-26 — no rule was invalidated

Spot-checked against shipped code, the ones Epic 3 could plausibly have broken:

- **AD-1** — `grep -rn 'os\.Getenv\|"os"\|"time"\|"net\|math/rand'` over non-test `folio-go/internal`
  → **0 hits**. The only `os.Getenv` in the module is `cmd/folio/main.go:77`, which AD-1 permits and
  AD-7's `Binds` line already names.
- **AD-7** — text unchanged and now finally has a referent: `resolveDocumentDate`
  (`render.go:1272`), reserved key at `render.go:1302-1305`, emission via `internal/pdf/infodate.go`.
  The clause *"omitted unless a date arrives through `params`"* is exactly what shipped.
- **AD-5** — `internal/pagemodel` imports only `internal/geom`; `internal/layout` imports geom and
  pagemodel, never pdf. Seam intact.
- **AD-23** — `grep -rn 'float64' folio-go/internal` non-test → **10 hits, all inside comments**.
  Zero declarations.
- **AD-9** — `internal/expr/parser.go` is hand-written recursive descent; the closed function table
  is exactly 8 (`internal/expr/eval.go:64-78`).
- **AD-12** — the 4-row closed table with `LocaleTableVersion` mirrored to `folio-go/version.go:18`;
  the Story 3.4 amendment's text matches the code.

### FINDING 2.1 — BLOCKING. `ARCHITECTURE-SPINE.md:637-638` says "alone", and Epic 3 made that false.

> `hashmatrix/ # module … — holds the retained FMA contraction probe **alone**`

**Measured:** `ls hashmatrix/` → `floatdiscrimination/`, `probe/`. Story 3.3 (`c2b4fe2`) added
`hashmatrix/floatdiscrimination/`, which `.github/workflows/ci.yml:69-71` names explicitly
(*"Story 3.3, D-3.3.7: floatdiscrimination/ is a real guard only if this runs it"*). Epic 3 falsified
a word in the spine and did not amend it.

### FINDING 2.2 — BLOCKING. `docs/` exists and the spine's source tree does not name it.

`grep -n "docs/" <spine>` → **0 matches.** `ls docs/` → `README.md`, `expression-reference.md`,
`expression-reference.html`, `folio-mvp-plan.md`. `docs/expression-reference.md` is an **Epic 3
deliverable** (`5c94bae`, *"docs: add the template-author expression reference"*) and it is the
project's first shipped template-author documentation. A top-level directory absent from the tree
listing is precisely what the tree listing exists to prevent.

### FINDING 2.3 — BLOCKING (or a deliberate downgrade). The dependency-direction graph forbids three edges Epic 3 added.

`ARCHITECTURE-SPINE.md:58` states flatly: *"Arrows point at what a package may import. **Anything not
drawn is forbidden.**"* Measured against `go list -f '{{join .Imports "\n"}}'`, the real graph carries
**eight edges the diagram does not draw**, three of them **added by Epic 3**:

- `folio → internal/expr` and `folio → internal/bind` (Stories 3.2/3.3 — `folio_expr_validate.go`)
- `internal/expr → internal/template` (Story 3.4 — `internal/expr/locale.go:24` keys the locale table
  off `template`'s constants)

Five more predate Epic 3 (`folio →` fontset/geom/pagemodel/text; and `internal/fontset → internal/text`,
which the diagram draws **backwards** at `:75`). Underneath all of it, the **enforced** model is no
longer this diagram at all: it is D-000.16's **stage-rank table** (`lint/internal/rules/stagerank.go:57-68`),
which permits every one of the eight. The spine names neither "rank" nor D-000.16.

**Why I file this as blocking rather than carried forward:** the section does not merely omit edges,
it asserts a prohibition that CI contradicts. A reader who trusts `:58` will reject a legitimate
import in review. **The cheapest honest fix is one sentence** — say the mermaid graph is
illustrative and that `stagerank.go` is normative — not redrawing the graph.

### Carried forward from earlier epics (not Epic 3's to fix, recorded so it stops recurring)

- `:75` — `TXT --> FSET` points the wrong way (real direction `fontset → text`, ratified at Story
  2.5, encoded at `stagerank.go:58-68`).
- `:626` region — `cmd/gentrie`, `cmd/gencorpus`, `cmd/genbreaks` (Story 2.1) unlisted;
  `internal/text/data/`, `internal/text/wordlist/` unlisted.
- `:525` — the Stack table lists `go-text/typesetting v0.3.4` as a test-scope dependency. It was
  **rejected twice** at Story 2.3 (`folio-go/shaping_oracle_test.go:21-32`) and appears nowhere in
  `go.mod`. The Stack row names a dependency that does not exist.
- Locative drift, non-blocking, `Binds` lines correctly frozen: **AD-14** (`:298-311`) still reads as
  though `internal/diag` owns `Diagnostic`; it owns only the registry — `Diagnostic`, `Severity` and
  `Result` are in the root package (`diagnostic.go:24`, `:240`, `:302`), a fact recorded **only** in
  the source-tree footnote 300 lines away. **AD-12**'s `Binds: internal/expr` (`:269`) while the
  closed locale *set* is enforced in `internal/template`. **AD-3** (`:145`) does not anticipate
  `internal/pdf/infodate.go`'s date literal — the rule is not violated.

### "Future" language that has shipped — clean

27 forward-looking hits reviewed individually (`will`, `later`, `not yet`, `future`, `eventually`,
`planned`, `TBD`, `Epic 3`, `Story 3.`). **None describes as future anything Epic 3 landed.** The
`wasm/` and `folio-designer/` tree entries are legitimately future. The only forward-looking word
Epic 3 falsified is **"alone"** at `:637`, which no such grep would find — Finding 2.1.

---

## Item 3 — Exported surface · **CLEAN on leakage and on the zero-value property; the README is three-times false**

### The surface, enumerated

**31 top-level exported declarations plus 9 exported struct fields = 40 items.** Two independent
instruments agree exactly: `go doc -all .` (561 lines) and a `go/ast` enumerator over the non-test
root files; a `go/types` probe with `importer("source")` reports 28 package-scope objects (= 31 − 3
methods, which are not scope objects).

- **15 constants** — 11 `DiagCode*` (`diagnostic.go:128`–`:232`), `SeverityWarning` (`:60`),
  `SeverityError` (`:68`), `Version` (`version.go:9`), `LocaleTableVersion` (`version.go:18`).
- **8 types** — `Severity`, `Diagnostic`, `Result`, `Template`, `FontSet`, `Data`, `Params`,
  `RenderError`.
- **5 funcs** — `LoadTemplate`, `ParseTemplate`, `Render`, `RenderTo`, `Validate`.
- **3 methods** — `Severity.String`, `(*RenderError).Error`, `(*RenderError).Unwrap`.
- **9 exported fields** — `Diagnostic`×5, `Result`×2, `RenderError`×2. **No exported `var`.**

Epic 1's surface was nine declarations. It is now 31 — and **`LocaleTableVersion` (Story 3.4) is new
public surface that was not on this gate's brief**; it is intentional and documented
(`version.go:11-18` states why, and defines it as exactly `expr.LocaleTableVersion` so the two cannot
drift). Recorded here so the v0.1.0 inventory is complete.

### Epic 1's `Template` ruling **landed**, and is compiler-enforced

`folio.go:33` is now `type Template struct { doc *template.Document; derivedFooters map[...] }` — an
opaque handle, no `=`, no exported fields. Enforced not in prose:
`TestTemplateCompositeLiteralDoesNotTypeCheck` (`templateopaque_test.go:33`) builds a bad fixture by
explicit path and asserts it fails to type-check with *"unknown field"*. **Anchored to the compiler**
(D-000.68). The 32-declaration alias leak is closed.

### Zero internal leakage — measured, not inspected

A `go/types` probe walked all 40 surface items — every exported object, method signature and struct
field — matching against `/internal/`. **Zero hits.** The only internal types in the package are
`Template`'s two **unexported** fields.

### `Severity`'s zero value — the property holds, and DW-18's mutation still bites

`Severity` is the **only** exported enum-like named scalar with a constant group in package `folio`.
Measured, not assumed: `awk '/iota/' *.go` over non-test root files → **one hit, `diagnostic.go:53`**.
Nothing else qualifies — the 11 `DiagCode*` are plain `string`, and `Data`/`Params`/`FontSet` are
container types. **So the DW-18 generalisation this gate was asked to check has an empty remainder:
there is no second enum to check.**

Measured values: `SeverityWarning=1`, `SeverityError=2`, `Severity(0).String() == "Severity(unset)"`,
`severityUnset` unexported. **DW-18's own mutation re-run at every production construction site:**

| site | result |
|---|---|
| `render.go:755` clip warning | **RED** (`render_clip_diagnostic_test.go:120`, `:145`; `:405`) |
| `render.go:528` empty-average caveat | **RED** (`render_empty_average_diagnostic_test.go:67`, `:292`) |
| `render.go:1067` missing glyph | **RED** (`ac4_coverage_test.go:144`; `missing_glyph_coalesce_test.go:108`) |
| `render_error.go:75` `SeverityError` | **RED** (`TestFourErrorModesCarrySeverityErrorDiagnostics`, all four subtests) |
| **`render.go:556`** `DiagCodeInternalUnhandledCaveat` arm | **GREEN — 330 passed** |

Every red reports `Severity = Severity(unset)`, which is DW-18's renumbering visibly doing its job.

### FINDING 3.1 — BLOCKING (doc). `folio-go/README.md` states three things Epic 3 made false.

```
$ awk 'index($0,"Validate")||index($0,"RenderError")||index($0,"Severity")||index($0,"DiagCode")||index($0,"documentDate")' folio-go/README.md | wc -l
0
```

The README documents **none** of Epic 3's public surface, and its *"What's not here yet"* section is
now actively wrong:

- **`README.md:349-351`** — *"`{{page}}` and `{{pages}}` … **nothing resolves them yet**"*. False
  since Story 2.7: `page_number.go` and `TestReservedPagePlaceholdersResolveOnEveryPage`
  (`reserved_placeholders_test.go:99`).
- **`README.md:352-358`** — *"**No PDF metadata date is emitted by this version of the library at
  all** … Wiring `SOURCE_DATE_EPOCH` through the command-line tool arrives with `cmd/folio` **in a
  later story**."* False on both halves as of Story 3.7 (`render.go:1294-1298`;
  `cmd/folio/main.go:110-112`, `:374-393`).
- **`README.md:359-362`** — the matrix *"is verified at the **Epic 1** boundary"*. Two boundaries have
  since run.

And the one that matters most for DW-4: **`documentDate` — the first and only reserved name in the
`params` namespace, frozen at v0.1.0 (`render.go:1302-1304` says so in as many words) — appears in no
shipped document anywhere in the repository.** Its only user-reachable text is `cmd/folio`'s `-help`
output. A frozen caller-facing key documented only in a CLI help string is the finding I would hold
the tag for.

### Carried forward — cheap, pre-tag preferred

1. **Four `DiagCode*` constants render bare in godoc** — `diagnostic.go:194`, `:205`, `:206`, `:207`.
   Each is its own `GenDecl`, so the group's prose attaches only to the first sibling. `go doc` sorts
   alphabetically, so `DiagCodeBindingPathAbsent` is the **first** constant a reader meets and it is
   undocumented. One line each.
2. **`render.go:556` is a mutation-invisible `Severity` site.** `diagnosticFromCaveat`'s `default:`
   arm can drop its `Severity` field with the suite staying green — and it is **not** unreachable:
   `TestAllProducedDiagnosticsCarryARegisteredCode` calls it directly with a fabricated
   `expr.CaveatKind(99)` and asserts only on `Code`. The AST class-guard beside it,
   `TestNoDiagnosticCompositeLiteralOmitsCode` (`diag_no_empty_code_test.go:56`), enforces `Code` over
   the whole literal class but has **no `Severity` counterpart** — confirmed by deleting the adjacent
   `Code:` line, which reddens both tests at once. Not dangerous (the numbering means an omission
   yields `Severity(unset)`, not a false `Warning` — precisely what DW-18 bought) but it is the same
   class one field over, and the fix is symmetric and free.
3. **The public→internal code bridge is pinned for 2 of 11.** `diag_bridge_test.go` asserts exact
   literals for `DiagCodeTextClippedWidth` and `DiagCodeEmptyAverage` only.
   `internal/diag/diag_test.go:30-40` pins all 11 **internal** names, but nothing checks the
   **mapping**. Repointing `DiagCodeExpressionInvalid = string(diag.CodeTemplateMalformed)` compiles
   and passes the entire suite — silently changing a frozen public constant. That is exactly what
   `diag_bridge_test.go`'s own doc comment says it exists to prevent. **Extend the table to 11 rows.**
4. `DiagCodeTableFooterSourceUnresolved`, `DiagCodeTableFooterSourceForbidden` and
   `DiagCodeInternalUnhandledCaveat` have **zero mentions in any `_test.go`**; their conditions are
   exercised, the public constants are not. No test pins `Severity(0).String()` either — the
   `case severityUnset:` arm is only ever observed incidentally, inside other tests' failure messages.

---

## Item 4 — Story-file and tracker hygiene · **PASS on the opener; the Delivery Logs are uneven**

### `sprint-status.yaml` — **PASS.** No sub-agent narrative.

Every Epic 3 key carries a bare, valid status value (`done` ×8, `optional`, `in-progress` for the
epic key, which is this gate's to flip). **One comment block exists**, the three-line insertion
rationale for `3-1a` (`sprint-status.yaml:83-85`). **Attributed by measurement, not assumed:**
`git log -S "3-1a inserted between 3.1 and 3.2"` → **`72cfc6a`, "Epic 3: record the four
pre-Story-3.1 rulings; insert Story 3.1a"** — the orchestrator's own epic-planning commit, before any
story ran. It is form-identical to the Epic 2 precedent (`2-3a`, `2-5a`, `2-6a`) and is insertion
provenance, not story narrative. **Not the defect the rule names.**

### `## In plain terms` — **8/8 exist, 8/8 immediately precede `## Story`, 8/8 clean of forbidden content**

All eight use the identical heading `## In plain terms (read this first if you just want the gist)`
and sit as the heading directly before `## Story` (3.1 `:70`→`:104`, 3.1a `:65`→`:93`, 3.2
`:64`→`:93`, 3.3 `:78`→`:110`, 3.4 `:106`→`:141`, 3.5 `:158`→`:188`, 3.6 `:192`→`:225`, 3.7
`:179`→`:222`). Mechanically scanned each section's exact line range for backticks, `/`,
`.go`/`.md`/`.json`, `AD-`, `D-[0-9]`, `DW-`, `FR-`, `NFR-`, `AC-`, camelCase and snake_case:
**zero hits in all eight** — with a positive control confirming the pattern works (the same grep over
`3-2…md:1-30` returns 14 hits). The only near-misses are markdown emphasis and quoted output samples
(`"15 สิงหาคม 2569"`, `"1,234.56"`), neither of which is code.

**FINDING 4.1 — three openers describe the plan, not what shipped.** Carried forward, minor.

- **3.4 `:112`, `:136-137`** — the hard case. *"**This story builds them**, and closes a gap found
  along the way"*, then *"**Done looks like this**: the Thai statement shows …"* and *"Two things
  **will** look wrong afterwards and are not"*. A definition-of-done is a plan, not a delivery.
  Compounding it, the second of those two things — *"three licence-checker warnings appear only on
  this developer's machine"* — was **fixed at Story 3.6** (DW-19; `lint` is 112/0 today), so 3.4's
  opener now tells a reader something untrue about the repository.
- **3.5 `:162`, `:166`, `:184`** — *"This story is where the engine finally reads it"*, and two
  *"will look wrong"* sentences.
- **3.7 `:181-182`, `:214`** — *"Folio **is** a library today … **This story gives it** a
  command-line tool"*, and *"What **will** look wrong later"*. The rest of 3.7's opener from `:206`
  is correctly past-tense.

3.1, 3.1a, 3.2, 3.3 and 3.6 are clean: *"has been tightened"*, *"can now"*, *"are now told apart"*,
*"Now we do"*.

### Delivery Logs — **FINDING 4.2, and 3.7 is the one that matters**

The rule: a Delivery Log names the suites it measured **and** names any suite it did not run. Silence
about a suite is the defect. Repo gates: `folio-go` `go test`, `lint` `go test`, `hashmatrix`
`go test`, the `-tags=matrix` four-target matrix, `go vet`, `gofmt -l`.

| Story | matrix non-run stated? | unmentioned gates |
|---|---|---|
| 3.1 | **yes**, twice (`:828-829`, `:882`) | `gofmt` (zero occurrences in the file), `hashmatrix` |
| 3.1a | **yes**, twice (`:654-655`, `:757`) | `hashmatrix` |
| 3.2 | **yes**, twice (`:870`, `:970-974`) | `gofmt` (zero occurrences), `hashmatrix` |
| 3.3 | **yes** (`:1172-1175`) | **none — the only complete log of the eight** (3×3 grid at `:1156-1162`) |
| 3.4 | **yes**, twice (`:970-972`, `:1107-1108`) | vet/gofmt for `lint` and `hashmatrix` |
| 3.5 | **yes, and it ran the native leg** (`:759`) | vet/gofmt for `lint` and `hashmatrix` |
| 3.6 | **yes**, and it names 3.5's three outstanding legs too (`:1005-1007`) | vet/gofmt for `lint` and `hashmatrix` |
| **3.7** | **NO — silence** | the `-tags=matrix` matrix; `go vet` for `hashmatrix` |

**Story 3.7 — the last story of the epic, the one immediately before this gate — never states that
the matrix was not run.** The closest it comes is oblique (`:1066-1067`, `:1975-1976`: *"Epic 3's
boundary gate inherits ONLY Story 3.5's three outstanding `hidden-image` legs"*), which **implies**
it ran none but never says so. Six of the eight get this right explicitly; the one place it matters
most is the one that is silent. **Carried forward — the orchestrator's matrix run has since made the
consequence moot, but the record is the defect.**

`fixtures/hidden-image/`'s three deferred legs are recorded at `byte_neutrality_test.go:549` and
**were discharged by the orchestrator's matrix run**, which passed.

### FINDING 4.3 — D-000.9 extension, spot-checked across four Delivery Logs

The standard: a recorded measurement must name **(a)** the command, **(b)** the mutation it reddens
under, and **(c)** that the mutation was run.

**Exemplary — all three present:**

- **3.7 `:991-1038`** is the strongest record in the epic, and it is strong because it **red-proved a
  prior measurement as inert**. The originally-recorded AC14 command ran only three digest-agreement
  tests that re-hash committed `expected.pdf` files and never call a renderer; a one-byte trailer
  mutation in `builder.finish` left it at `rc 0, mutation UNDETECTED`, while `go test ./...` reddened
  seven genuine goldens. Both commands are quoted verbatim, and the corrected command is named as the
  one to cite going forward. **This is D-000.9's extension catching itself.**
- **3.6 `:831-835`** — DW-18's M8 run at baseline as a deliberate **green**-proof: the field deleted,
  all nine tests still passing, *"Confirmed INVISIBLE at baseline"*. Then `:887-890`, the same
  mutation post-renumbering, **RED** with the exact failure text quoted. Both directions, both run.
- **3.6 `:851-865`** — records a **negative** result rather than hiding it: the pre-existing 3.3
  suites compare `d.Code` against the symbolic constant, so they *move with* the mutation and stay
  green. `grep -n "AGGREGATE_EMPTY_AVERAGE" render_empty_average_diagnostic_test.go` → zero hits,
  quoted. A self-referential comparison identified as such — D-000.68's exact diagnosis, self-applied.
- **3.2 `:794-807`** — AC17's red-proof: source copied with `cp`, `evalCall` mutated, the precise
  `go test -run …` invocation given, the failure line quoted, restoration confirmed with `diff`.

**Below the standard — counterfactual, not run:**

- **3.2 `:767-772` (AC21)** — *"**would have failed** … **Verified by inspection**"*. No command, no
  executed mutation.
- **3.2 `:785-789` (AC23)** — the only command shown is a `git show | grep` that verifies a
  *location*, not a reddening. The guard was never run against the pre-change tree.
- **3.2 `:773-784` (AC22 mutation 1)** — *"would have failed … **reproduced structurally**"*. Mutation
  2 of the same pair is a standing test and is fine.
- **3.4 `:1004-1009` (AC12 Layer 1)** — *"measured by **integer bit-length reasoning** … never by
  executing `float64` arithmetic"*. Reasoning is not a measurement. **And the same log contradicts
  itself at `:1113`**, where AC12 Layer 1 is listed among the mutations that *were* run (the
  `tenPowLookup` mutation). Two different sub-claims wearing one AC label; the log should separate
  them.
- **3.4 `:1074-1077` (AC1 non-vacuity)** — a green re-run offered as proof of non-vacuity, with no
  mutation applied. A guard that passes is not thereby a guard that can fail.

**The dominant shape across all four logs is (b)+(c) without (a):** the mutation is named and its
execution asserted, but no invocation string is given, so the claim cannot be re-run from the log
alone. That is a **recoverable** defect and I do not hold the epic for it. The counterfactuals above
are not recoverable, and I checked three of them against the code: **M3, M4 and M5 in Item 1 are
precisely the mutations 3.2's AC21/AC22/AC23 records only reasoned about — all three redden.** The
guards are sound; the *records* were weaker than the guards.

One honest disclaimer worth naming as the correct form: **3.7 `:1056-1060`** — *"**Stated as a
construction argument, not measured.**"* That is what a non-measurement should look like.

---

## Item 5 — The one required red, and what sits behind it · **D-000.70**

**`lint` is fully green — measured, not carried:** `go test -count=1 -v ./...` in `lint/` →
**112 `--- PASS`, 0 `--- FAIL`**, exit 0. This confirms the orchestrator's 112/0.

**`folio-go` carries exactly one red**, the required one: `TestCorpusMeetsP6ExerciseFloors`
(`internal/text/corpus_test.go:169`), *"P6g (opaque names) floor not met: got 7, need >=20"*,
mandated unmet by D-000.17 / D-2.1.14 / D-000.57 and stable at
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — character-for-character the Epic 2 figure.

### In its own package: nothing is unreached. Measured.

`go test -count=1 -v ./internal/text/` written to a file and counted: **48 `=== RUN`, 47 PASS,
1 FAIL, 0 SKIP.** The failing test's siblings all run. And **inside** the test, all seven floors use
`t.Errorf`, not `t.Fatalf` (`corpus_test.go:185-187`), so every floor is evaluated and the final
`t.Logf` stats line executes. **Nothing downstream of the red is unreached at the package level.**

### FINDING 5.1 — but two CI steps sit behind it, and they have never run in CI.

`.github/workflows/ci.yml`'s `folio-go` job runs, in order: `go build` → `go vet` → `gofmt -l` →
**`go test -count=1 ./...`** → `go build -tags=matrix ./...` → `go vet -tags=matrix ./...`.

`go test -count=1 ./...` **exits 1 at HEAD** (measured). GitHub Actions aborts a job at the first
failing step absent `continue-on-error` or `if: always()`, neither of which is set. **So the two
matrix-tagged steps are unreachable, and have been since the corpus floor landed at Story 2.1.** They
are the *only* thing in `ci.yml` that compiles the `//go:build matrix` files — `matrix.yml` runs on
a per-epic cadence, not per push. A matrix-tagged file could stop compiling and no push would say so.

**Run locally to establish whether the red is currently hiding a second defect:**

```
go build -tags=matrix ./...   EXIT=0
go vet   -tags=matrix ./...   EXIT=0
go vet ./...                  EXIT=0
gofmt -l .                    (no output)
```

**Nothing is hiding behind it today.** The gap is latent, not live — which is the same shape as Epic
1's licence-manifest finding, and the same reason to fix it while it costs nothing. **Not blocking
for Epic 3** (it predates the epic and no defect is concealed) **but the entire `folio-go` CI job has
been standing-red for two epics**, which is D-000.15's erosion dynamic aimed at the repo's own
front-door signal. The fix is a job-level accommodation that makes the required red *expected*
rather than *fatal* — not a `t.Skip`, which the Epic 2 gate already declined for `.fontgen-venv`
(D-000.58).

---

## Cross-target matrix

**Run by the orchestrator before this gate and PASSED.** Not re-run here. It discharges Story 3.5's
three outstanding `hidden-image` legs (`byte_neutrality_test.go:549`), which were the only matrix
debt Epic 3 carried.

---

## What must be resolved before `epic-3: done`

**Three prose corrections, all one-line, all in this gate's own commit.** Nothing in the code blocks.

1. **`ARCHITECTURE-SPINE.md:637-638`** — delete "alone"; name `hashmatrix/floatdiscrimination/`
   (Story 3.3). *Finding 2.1.*
2. **`ARCHITECTURE-SPINE.md` source tree** — add `docs/`. It holds this epic's shipped
   template-author reference. *Finding 2.2.*
3. **`folio-go/README.md:349-362`** — three false claims: `{{page}}`/`{{pages}}` resolve (2.7), PDF
   metadata dates are emitted (3.7), and the matrix is verified at Epic 3's boundary, not Epic 1's.
   *Finding 3.1.*

And one that is a judgement call rather than a falsehood:

4. **`ARCHITECTURE-SPINE.md:58`** — *"Anything not drawn is forbidden"* is contradicted by eight real
   import edges, three added by Epic 3, all permitted by `stagerank.go`. One sentence naming
   `stagerank.go` as normative and the diagram as illustrative resolves it. *Finding 2.3.*

### Carried forward — **twelve items, with owners**

| # | Item | Owner |
|---|---|---|
| 1 | **DW-14's owner is a spent event** (the Epic 2 gate). Re-measured unmoved at 45/100. Re-own it. | this gate to re-assign |
| 2 | **DW-16's forcing function does not exist in the roadmap** (0 hits for a non-PDF renderer in `epics.md`). Third instance of this shape. | engineering lead |
| 3 | **DW-13 was never scheduled** by the orchestrator, as its own owner line requires. | orchestrator |
| 4 | **DW-3, DW-4 and DW-20 all gate on the unowned `folio-go/v0.1.0` tag**, and Epic 4 planning is the moment DW-4 names. | **the project owner, at Epic 4 planning** |
| 5 | **`documentDate` is a frozen reserved `params` key with no shipped documentation** — CLI `-help` only. | before the tag |
| 6 | The public→internal `DiagCode` bridge is pinned for **2 of 11**; a repointed constant compiles and passes. | before the tag |
| 7 | Four `DiagCode*` constants render bare in godoc (`diagnostic.go:194,205,206,207`). | before the tag |
| 8 | `render.go:556`'s `Severity` omission is mutation-invisible; the AST class-guard covers `Code`, not `Severity`. | next story touching `diagnosticFromCaveat` |
| 9 | The **`folio-go` CI job has been standing-red for two epics**, and `go build/vet -tags=matrix` sit behind the required red, unreached. Latent today, measured. | orchestrator |
| 10 | **3.7's Delivery Log never states the matrix was not run** — the last story before the gate. 3.1 and 3.2 never mention `gofmt`. | recorded; process |
| 11 | Three `## In plain terms` openers describe the plan (3.4 hardest — its `:137` also now misstates the repo state, fixed at 3.6). | recorded; process |
| 12 | Seven retired DW entries sit under `## Open`; `## Done` holds only two of nine. | cosmetic |

Plus the pre-Epic-3 spine items recorded under Item 2 (the reversed `TXT --> FSET` arrow, three
unlisted `cmd/` binaries, and the `go-text/typesetting` Stack row naming a twice-rejected dependency),
which are Epic 1/2 debt and are listed there rather than counted here.

---

## Orchestrator's addendum — what the gate run itself measured

Written by the orchestrator after reading the four audits above, verifying their blocking findings
independently, and applying the fixes. Three of the four blockers are now **closed in this gate's own
commit**; the fourth is with the engineering lead. One finding below is **new** and did not come from
the audits.

### Blockers 1–3 — closed, with the edit named

| # | Finding | Verified how | Fix applied |
|---|---|---|---|
| 1 | Spine said `hashmatrix/` holds the FMA probe **"alone"** | `ls hashmatrix/` → `probe/`, `floatdiscrimination/`, `go.mod`, `README.md` | The entry now lists **both** probes as sub-entries and states *why* both live outside `folio-go`: neither the AST guard nor `lint`'s float rule has an allowlist, so a landed, executing `float64` mutant has nowhere under `folio-go/` to live (D-000.24) |
| 2 | Spine source tree omitted `docs/` | `ls docs/` → `expression-reference.md`, `.html`, `folio-mvp-plan.md` | Added above `fixtures/`, carrying **D-3.4.3's caveat in the tree itself**: the page is hand-written, is not derived from the code, has been wrong, and is **never evidence for a decision about what the engine does** |
| 3 | `README.md` "What's not here yet" carried three false claims | `page_number.go` exists (2.7); `internal/pdf/infodate.go` + `cmd/folio/main.go:362` exist (3.7); matrix cadence is D-000.4 per-epic | Page-numbering bullet **deleted** (it shipped at 2.7 and no longer belongs in a "not here yet" list). Date bullet rewritten to state the shipped behaviour — including that the `SOURCE_DATE_EPOCH` route is **fill-only**. Matrix bullet now says *"at each epic boundary … most recently at the Epic 3 boundary, over ten documents on all four targets"* |

**Blocker 4** (`ARCHITECTURE-SPINE.md:58`, *"Anything not drawn is forbidden"*) is **not applied**, and
deliberately so — see *Finding O.2*.

### FINDING O.1 — **BLOCKING THE RECORD, NOT THE CODE. CI has never run. Not once, in fifty commits.**

The audit's Item 5 says the `folio-go` CI job *"has been standing-red for two epics."* That
understates it, and the difference matters.

```
gh auth status          → logged in as panitw, token scopes include 'workflow'
gh repo view            → {"defaultBranchRef":{"name":"main"},"isPrivate":false,"name":"folio"}
gh run list --limit 5   → []                       ← zero workflow runs, ever
git ls-remote --heads origin → main = f2aa8c0
git log --oneline origin/main..HEAD | wc -l → 50
```

**`origin/main` is fifty commits behind local `main`.** Every commit of Epics 1, 2 and 3 is local
only. `ci.yml` and `matrix.yml` have therefore **never executed a single time**. The job is not
standing-red; it is standing-**unrun**, which is D-000.9's own distinction — *"all clear" must never
be indistinguishable from "could not look"* — applied to the repo's front door.

This is a consequence of the run's own standing instruction (**never push**), not a defect anyone
introduced. But it has a consequence that reaches backwards into the record: **every claim in this
program of the form "CI is green" is an inference from local measurement, not an observation.** The
DW-19 disposition is the clearest instance — *"`.font-sources/` is gitignored, so the licence reds are
local-only; CI green."* The inference is **sound** (a fresh checkout genuinely does not contain those
files, and I re-confirmed the `.gitignore:85` entry), but it was written in the grammar of an
observation. Under D-000.9's extension, a construction argument must say it is one — as Story 3.7's
own log correctly did: *"Stated as a construction argument, not measured."*

**Nothing here is wrong in the code.** What is wrong is a class of sentence in the record.

### FINDING O.2 — the required red does not merely hide two steps; it makes CI structurally incapable of ever being green

Verified at HEAD: `go test -count=1 -run TestCorpusMeetsP6ExerciseFloors ./internal/text/` **fails** —
*"P6g (opaque names) floor not met: got 7, need >=20"* — with **no build tag on the test**. So on the
first push, `ci.yml`'s `folio-go` job fails at its `go test` step, permanently, by design.

D-2.1.14 chose that deliberately and refused to amend the floor, on a principle worth restating
because it is the reason not to "fix" this the easy way: *"post-hoc amendments always have good
arguments, which is precisely why the rule forbids them rather than weighing them."* The floor stays
unmet. **That is settled and is not reopened here.**

What was never decided is **where an intentionally-unmet floor gets reported.** A permanently-failing
test in the default job means CI can never distinguish "the known corpus shortfall" from "someone
broke the renderer" — the standing red would not be hiding a second defect, it would be hiding *every*
defect. That is D-000.70 at the scale of the whole signal rather than one package.

**Applied now, because it costs nothing and is independent of that decision:** the two matrix-tagged
steps have been **moved ahead of the test step** in `ci.yml`'s `folio-go` job, next to their untagged
siblings:

```
go build ./...            →  go vet ./...            →  go build -tags=matrix ./...
  →  go vet -tags=matrix ./...   →  gofmt -l .   →  go test -count=1 ./...
```

Both tagged steps verified green locally at HEAD (`EXIT=0`, `EXIT=0`). They are the only thing in
`ci.yml` that compiles the `//go:build matrix` files, and they no longer sit behind a step that is
guaranteed to fail. **This does not resolve the reporting question** — it only stops one specific
casualty. The reporting question is with the engineering lead.

### FINDING O.3 — the README documents neither the CLI nor page numbering

Falling out of blocker 3: `grep -nE 'cmd/folio|folio render|folio validate|page\}\}' folio-go/README.md`
returns **only the three lines I was correcting**. The project's first executable shipped yesterday
at Story 3.7 and `folio-go/README.md` does not mention it anywhere; `{{page}}`/`{{pages}}` shipped at
Story 2.7 and are named nowhere outside the bullet now deleted. Not blocking — but deleting a stale
bullet without adding a real section leaves the README quieter about page numbering than it was, so
this is recorded rather than absorbed. **Carried forward, owner: before the `folio-go/v0.1.0` tag**,
alongside carried item 5 (`documentDate` is likewise undocumented outside CLI `-help`).

### Blocker 4 — resolved by engineering-lead ruling **D-000.72**, and the finding was understated

Referred to the engineering lead rather than applied, because the auditor's one-sentence fix
("the diagram is illustrative, `stagerank.go` is normative") looked like it resolved the contradiction
by lowering what the document claims. **The lead disproved the orchestrator's stated reason and
reached the same refusal by a better one.**

`layout ↛ pdf` **is** enforced by executed mechanism, three ways over — `stagerank.go:66-67` (7 and 8,
under a strictly-lower comparison at `:239`), `TestStageRankProductionScan`'s tree scan with a vacuity
guard naming both packages, and a **retained violating fixture** at
`folio-go/testdata/lint/stage-rank/layout/violating_pdf_import.go` that quotes the spine's own sentence
and imports `internal/pdf` from package `layout`. All three run in CI's **`lint` job**, which declares
no `needs:` (`ci.yml:87-110`) and is therefore **not** gated behind the standing-red `folio-go` job —
so carried item 9 does not mask them. Verified independently by the orchestrator before applying.

**And this gate measured only half the drift.** The lead measured the other direction: of the **24
arrows the diagram drew, 11 named edges that do not exist** — `layout→bind`, `layout→text`,
`layout→diag`, `bind→template`, `bind→geom`, `text→fontset` (backwards), `text→geom`, `text→diag`,
`pdf→fontset`, `expr→geom`, `expr→diag`. The real internal graph is 12 edges; the diagram drew 24 and
got 13 of them wrong. **That is why "illustrative" was the wrong fix**: it excuses the omissions and
leaves eleven fictional dependencies standing, blessed as officially non-binding rather than removed.
A reader reasons from what is drawn.

**Applied, both arms in this gate's commit** (they must ship together, or Arm A's *"held in agreement
with that table by a test"* is false from birth — D-000.28):

- **Arm A** — the mermaid graph is **deleted**. The section now states the strictly-forward rank rule,
  names `stageRankTable` as the single declaration, and renders the ranks as a fenced ladder.
- **Arm B** — `TestSpineStageLadderMatchesStageRankTable` (`lint/internal/rules/stagerank_test.go`),
  anchored to the spine markdown file: authored outside both Go modules, unmovable by any code under
  test. Ordered equality, reported **both ways by name and rank**, never as a count. Two Fatal paths
  (missing file, missing/renamed fence) plus a <10-row shape guard, so a defeated extractor cannot
  read as agreement.

**Four red-proofs run, each mutation applied and the file restored:** delete a ladder row → RED
naming `{"bind", 4}`; change `layout`'s ladder rank to 9 → RED **both ways**; add a row to
`stageRankTable` only → RED; rename the `:begin` marker → **RED on the Fatal path**, not a silent
pass. `lint` moves 112 → **113 passing, 0 failing**; `gofmt -l` empty.

**Retired for free by this edit:** the carried-forward item *"`:75` — `TXT --> FSET` points the wrong
way"*. The arrow no longer exists. Carried-forward count drops from twelve to eleven, and the eleven
fictional arrows carry forward as nothing at all.

---

## Post-gate addendum — Item 5 and DW-16, both closed after `epic-3: done`

Recorded here rather than in Epic 4's gate because both are dispositions of **findings this gate
raised**, and the next reader of Item 5 should not have to follow a forward reference to learn it was
answered.

### Item 5's known red is now mechanised — **D-000.57 clause 3 is executed, not narrated**

The carried-forward item said the `folio-go` CI job had been standing-red for two epics with
`go build/vet -tags=matrix` unreached behind it, and Finding O.1 corrected that to standing-**unrun**.
The engineering lead ruled the disposition (**D-000.74**), and the reframe removed most of the work:
**D-000.57 already required the discriminator** — *"the test name stays red while what it measures
silently drifts"* — and already identified that the discriminator is the **numbers**, not the colour.
It was being checked by a human reading a log line once per epic. It is now a test.

| Change | State |
|---|---|
| Seven `t.Run` subtests, one verdict per floor | six named PASS, one named `…/P6g_(opaque_names)` FAIL. No floor value touched |
| `TestCorpusP6StatsMatchDeclaredBaseline` — green, in the green job | baseline transcribed from D-000.57 **and** verified against a live run; both agree. Red-proved at `7 → 6` |
| `ci.yml` split into `folio-go` and `folio-go-known-red` | green job **EXIT=0, 921 pass**; known-red job EXIT=1, its name the disclosure |
| The ratchet against 7 | **rejected** — it is D-2.1.14's amendment wearing a dynamic hat, and its only real motivation is already free from the drift detector |

**`KNOWN_RED_TEST` is a single workflow-level scalar, not a list**, so a second sanctioned red requires
changing the declaration's shape rather than appending a line. No `continue-on-error` on either job —
including the known-red one, because laundering the badge green would report an unmet floor as met on
the repo's most-read surface.

**One defect found in the ruling itself, by running the mutation it flagged as an assumption.** The
lead held that `-skip` fails safe under a rename. It does not: Go's `-run`/`-skip` are **unanchored**
regexps, so renaming the test to `…FloorsRenamed` and running the ruled command verbatim returns
**`ok`** — the quarantine survives the rename and the green job stays green over a red nobody is
watching. Anchoring to `^TestCorpusMeetsP6ExerciseFloors$` fixes it; verified both ways (unmutated →
921 pass; renamed → FAIL), and the anchors carry their reason at the declaration. **This was D-000.68
in a CI flag** — an accommodation anchored to a substring of the code's own spelling, which read as
safe to two readers and was caught only by executing it.

**Consequence for the next gate:** cite `TestCorpusP6StatsMatchDeclaredBaseline` as the discharge of
D-000.57's third clause instead of re-measuring the stats by hand. That is a reduction in gate work,
not an addition.

### DW-16's untriggerable owner is replaced by a mechanism — **D-000.73**

Carried item 2 said DW-16's forcing function does not exist in the roadmap. The lead confirmed it by
**measurement rather than by grep** — Story 5.10's preview consumes the real PDF, Story 5.9's canvas
paints pre-broken lines and never touches `TextRun.Glyphs` — and corrected a premise this gate had
also missed: DW-16's *"exactly one producer"* has been **false since Story 2.7**. There are two, and
the second is a copier, so option 1 is more expensive but not foreclosed. **Stories 4.1 and 4.2 do not
close the window**; the gate's assumption that 4.1 was a deadline was wrong in the direction of more
slack, not less.

`TestGlyphIdentifierCensus` now pins both censuses — producers at two, readers to `internal/pdf` plus
the copier — with field identity resolved through `go/types` rather than the name `CID`, because
`internal/text` declares its own `ShapedGlyph.CID` and a spelling-based instrument reports three
producers and eight consumers where the type checker reports two and two. **A read from anywhere else
reddens, and that red is the forcing function arriving.** Three red-proofs run; the field-rename case
fails **Fatal on the type-information path**, never as an empty census reported clean.

Placed in `lint` rather than `folio-go`, a deviation from the ruling that was surfaced to the lead
rather than left to be found: `lint` already type-checks the whole `folio-go` module, `folio-go`
cannot gain `x/tools` without breaking D-1.3.6's invariant (b), and D-3.7.9's lesson prefers the
lint-side instrument wherever one is affordable.

**Carried-forward count: eleven → nine.** Items 2 and 9 are closed. The option 1 / option 2 fork on
what the page model promises is **with the project owner**, batched into Epic 4 planning with DW-4 and
DW-13.
