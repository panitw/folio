---
baseline_commit: 4ec1884
---

# Story 3.6: Fail with located, actionable diagnostics

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-6-fail-with-located-actionable-diagnostics`
**Status:** `done`
**Covers:** **FR41** · **AD-14**

**Primary invariant:** **AD-14**, verbatim (`ARCHITECTURE-SPINE.md:298-311`):

> One `Diagnostic` value carries `Severity` (`Error` aborts the render, `Warning` accompanies a
> successful one), a **stable string code** from a closed registry, an optional element id
> (AD-10), an optional data path, and a message. Every failure mode named in FR41 has a code and
> a test asserting it. Over-tall rows (FR25) and clipped content (FR44) are `Warning`s returned
> alongside PDF bytes, never silent and never fatal. Codes are additive only; changing a code's
> meaning is a breaking change.
> Three data cases that would otherwise each be decided twice: an **absent** path is an `Error`
> carrying the path; an explicit JSON **`null`** renders as empty and is not an error; a value of
> the **wrong kind** for its element is an `Error`, never a coercion.

Also binding: **AD-1** (the determinism/import boundary and its stage-rank table), **AD-8** (*"A
glyph covered by no font in the chain is a diagnostic (AD-14) with the element id and the offending
rune, never a blank box"*), **AD-22** (release tags are directory-prefixed — `folio-go/v0.1.0` — and
any behaviour change is a breaking change once tagged), **AD-23** (a literal too large for the
representation is an `Error`, never a silent narrowing).

---

## Baseline, measured in this run at creation (`4ec1884`, tree clean)

Run exactly as the gate specifies, in this run, not cited from another story (D-000.26). Note the
shell filters plain `gofmt` output — invoke it as `$(go env GOROOT)/bin/gofmt`.

| module | result |
|---|---|
| `folio-go` | build ✅ · vet ✅ · gofmt ✅ · **853 pass · 1 fail · 1 skip** |
| `lint` | **104 pass · 3 fail** |
| `hashmatrix` | **3 pass** |

**The one `folio-go` red is REQUIRED and must stay red.** `TestCorpusMeetsP6ExerciseFloors`
(`internal/text/corpus_test.go:189`): *"P6g (opaque names) floor not met: got 7, need >=20"*, with
stats **`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`** — byte-identical to Story 3.5's
figures. D-000.17 / D-2.1.14 mandate it stay unmet; the Epic 2 gate closed over it deliberately.
**Any change to these stats is a finding**, not a success.

**The three `lint` reds are DW-19, local-only, and expected.** Measured verbatim in this run:

- `rules/TestFontsAssetsNoticeRemovalRedProof` — `fontsassets_test.go:307`, *expected error to
  mention "no NOTICE* file", got: `.font-sources: contains…`*
- `manifest/TestManifestUpToDate` — *ResolveAssets: `.font-sources`: contains a committed font
  binary but no LICENS…*
- `manifest/TestResolveAssetsIncludesWordlist` — same cause

**DW-19's premise re-verified against the tree in this run**, not trusted: `git ls-files
.font-sources` returns **0 tracked files**; `.gitignore:85` is `/.font-sources/`. The resolver's own
message says *"contains a **committed** font binary"* about a directory with nothing committed in it
— which is the defect stated in the resolver's own words. **Never "fixed" by weakening the rule, by
adding files to `.font-sources/`, or by making the tests skip** (D-000.58).

---

## SEVEN divergences between the record and the shipped tree, all measured in this run

**Where the record and the code disagree, the code is the truth.** Each of these changes an
instruction in D-3.6.0's inventory or in the spine.

### 1. There is no exported error **sentinel** in `internal/expr`. Inventory item 3's word is wrong.

D-3.6.0 item 3 reads *"one exported `expr` sentinel"*, and D-3.3.6's verdict line says *"An exported
**sentinel in `expr`**"*. A reader takes that as `var ErrX = errors.New(…)`. **It does not exist.**
`grep -n 'Err[A-Z]' internal/expr/` returns nothing; there is no `errors.New` sentinel in the
package.

What actually shipped is D-3.3.6's own **binding 3** — *"It is a typed value, never an error"* — as
`internal/expr/caveat.go`: a `Caveat` struct (`Kind`, `Path`), a `CaveatKind` closed discriminant,
and its single member `CaveatEmptyAverage`, travelling on a **third return**. The verdict line and
binding 3 of the same entry disagree with each other; binding 3 is what was built, and the reason it
was built that way (*"every future caller's default `if err != nil { return err }` would turn a
Warning into a render abort"*) is still correct.

**Consequence for this story:** item 3's *"leave the sentinel in `expr`"* means **leave
`expr.Caveat`/`expr.CaveatKind`/`expr.CaveatEmptyAverage` exactly where they are**. It does not mean
"go find a sentinel". Nothing in `internal/expr` is edited by this story.

**The lead's own correction, folded in — now ratified as [[D-000.69]]** (*a ruling's verdict line is
what later briefs quote; keep mechanism words out of it, or the bindings are lost*): the word should
have read **"marker"**. In Go, *sentinel*
means `var ErrX = errors.New(…)` — precisely the fail-dangerous shape D-3.3.6's binding 3 rejected,
because a caller's default `if err != nil { return err }` would turn a Warning into a render abort.
`expr.Caveat` **is that ruling implemented, not departed from.**

### 2. `internal/diag` **cannot** hold the caveat-kind→code mapping. The stage-rank table forbids it.

Inventory item 3 instructs *"move the mapping into `internal/diag`"*. Measured:
`lint/internal/rules/stagerank.go:57-68` ranks **`diag` at 1** and **`expr` at 3**, and the rule is
*"a package may import only a **strictly lower** rank"*. The mapping is
`diagnosticFromCaveat` (`folio-go/render.go:518-548`) and it switches on `expr.CaveatKind`, so
placing it in `diag` requires `diag`(1) → `expr`(3), which `ScanStageRank` rejects. It also
constructs a `folio.Diagnostic` and reads `folio.DiagCodeEmptyAverage`, and no `internal/` package
may import the module root at all.

`diag`'s rank-1 position is deliberate — its own comment: *"Epic 3; ranked ahead of arrival so its
first commit is already guarded."* A rank-1 leaf is exactly right for a **code registry**:
`template`(2), `expr`(3), `bind`(4), `fontset`(6) and `layout`(7) — **all five packages that own an
AC4 condition** — may import it, so every stage can attach a code. It is exactly wrong for a mapping
that must know about the expression evaluator. **Ruled at D-3.6.4: the low rank is the
requirement, not a convenience** — every higher placement forbids exactly the imports AC4 needs.

### 3. The spine's source tree says `internal/diag/` holds *"Diagnostic, Severity, the code
registry"*. Two of those three shipped at the module root instead.

`ARCHITECTURE-SPINE.md:613`. But D-2.8.3 (the owner's decision) landed `Severity`, `Diagnostic` and
`Result` in the module-root `folio` package (`folio-go/diagnostic.go`), because `Result` is
`Render`'s return type and `Diagnostic` is a public API type. This is not a defect — it is a
ratified decision the spine's tree was never updated for. **3.6 must not re-decide it** (`Do not
re-open`, below). See **D-3.6.4** for what `internal/diag` gets instead.

### 4. `SeverityError` is declared and never constructed — re-measured, and the count is exactly three.

D-3.6.1's measurement re-verified independently in this run. `Diagnostic{` appears at exactly three
non-test sites — `render.go:527`, `render.go:542`, `render.go:721` — and `Severity:` at exactly three
non-test sites — `render.go:528`, `:543`, `:722` — **all three `SeverityWarning`**. Nothing in the
module constructs an error-severity `Diagnostic`. D-3.6.1's *"how we'd know it was wrong"* (an
error-severity `Diagnostic` turning up before 3.6 opens) has **not** fired. See **D-3.6.3**.

### 5. A `Diagnostic` with an **empty `Code`** can already reach a caller — AD-14 is violated today.

`diagnosticFromCaveat`'s `default:` arm (`render.go:539-547`) returns a `Diagnostic` with
`Severity`, `ElementID`, `DataPath` and `Message` set and **no `Code` field at all** — the empty
string. AD-14 requires *"a stable string code from a closed registry"* on every `Diagnostic`, and
this story's AC7 is what enforces it. The arm is labelled unreachable
(`expr.CaveatKind` has one member), but it is a live, returnable construction site whose output is
not a registry member. This is a subject for AC7, not a note.

### 6. Missing glyph is a **WARNING**, and today it is an aborting `error` with no element id. Two defects, not one.

**Corrected at creation review and ratified as [[D-3.6.2]] — the creator's first pass had this
wrong**, and the correction makes
the story's largest code change land on the wire that already exists rather than on new machinery.

**Missing glyph is a Warning.** Four sources, all verified in this run:
- **AD-8**, verbatim: *"A glyph covered by no font in the chain is a **diagnostic** (AD-14) with the
  element id and the offending rune, never a blank box."*
- **`EXPERIENCE.md:216`**, the State Patterns table, **Diagnostic** row: *"Render succeeded with
  caveats — clipped content, over-tall row, **missing glyph**. Non-blocking, dismissible, locates the
  offending element."* The adjacent **Error** row (*"Render failed"*) does not list it.
- **UX-DR22** (`epics.md:232`) groups *"clipped element, over-tall row, **glyph with no coverage**"*.
- **Story 5.12's first AC** (`epics.md:1687`): *"a render **returning warnings** — clipped content,
  an over-tall row, **a glyph with no coverage**."*
- **FR41's own sentence carries the tension and resolves it in its last clause**: *"Fail with a
  located, actionable error … for … missing glyphs … **Non-fatal diagnostics are reported without
  failing the render.**"*

**So FR41's five modes split 4 Errors + 1 Warning**, and AC2's *"`Error` aborts it"* has **four**
subjects.

**Defect (a) — it aborts.** `resolveRuneFace` (`render.go:923-936`) returns
`fmt.Errorf("no font in chain %v has a glyph for rune %U — a located failure, not a blank box
(AC4)", chain, r)`, which propagates out of `shapeSegments` and fails the render.
**Defect (b) — no element id.** It has the rune and the chain; `resolveRuneFace` is never passed an
element id, and its caller `shapeSegments` (`render.go:949`) is element-id-free either.

**Measured coverage of the current behaviour: zero.** `grep -rn "no font in chain"` over the repo
returns **exactly one hit, the production site**. No test asserts that an uncovered rune aborts, so
converting it to a Warning breaks no existing assertion.

**Where it lands.** On the **existing `Result.Diagnostics` channel** — D-2.8.3's wire, unchanged, and
the same route `diagnosticFromCaveat`'s caveats already travel. `shapeSegments` gains a collector or
a third return in the shape `BindTextSpans` already uses; no new machinery, no new public type, and
**nothing to do with D-3.6.3's error type.**

**One thing no source settles — see OPEN-1.** A Warning accompanies **successfully returned PDF
bytes**, so the render must emit *something* for the uncovered rune, and neither AD-8, EXPERIENCE.md,
UX-DR22 nor 5.12 says what. AD-8 forbids *"a blank box"*, which rules out `.notdef`.

### 7. `absenceChecks` holds **three** entries, and the grounding's 3 → 2 → 1 → 0 schedule is correct.

Verified against `lint/internal/rules/absences.go:81-124`: `absence-designer-project`,
`absence-diag-package`, `absence-source-date-epoch`. 3.6 discharges the second (→ **2**), 3.7 the
third (→ **1**), 5.1 the first (→ **0**). `absences_test.go:100-105` pins all three rule ids by a
test-owned literal in `TestAbsencesChecksIncludeAllThreeEntries`; **that test's name, comment and
literal all change in this story.** And `absences_test.go:56-74` proves the scheduled zero is real:
substituting an empty `absenceChecks` yields zero findings and `ChecksEvaluated == 0`, a green pass —
**the mechanism goes slack, not red, at zero.** That is D-000.67 part 1's collision, and it is
**Story 5.1's** to remove, not this story's to decrement toward silently. 3.6 records it forward.

---

## In plain terms (read this first if you just want the gist)

When a report fails to build, the person holding it needs two things at once: a sentence saying what
went wrong and where, and a small, unchanging label a program can recognise without reading English.
We had good sentences already. We did not have the label. Now we do, for each of the five ways a
report can go wrong: a template that will not load, missing data, a bad expression, an unfit-to-print
piece of content, and a character no available typeface can draw. Four of those are fatal. The fifth —
the missing character — is not: the report still comes out, with a complaint riding alongside it,
because a mostly-right report beats no report. This story also filled in two long-promised labels for
table totals, and closed a bookkeeping gap where a diagnostic that forgot to state its severity was
indistinguishable from a harmless one — dangerous the moment fatal ones exist, and fixable only before
release.

A second pass, once the first version looked finished, found that several of the guarantees above were
not actually being checked the way they claimed. Two labels could have quietly traded meanings with
nothing noticing. One of the five failure paths could have shipped an unrecognised label past both of
its supposed safety nets. And the claim that rewriting a message never changes its label had only ever
been tested against a copy the test made of itself, never a real message. All three now check the real
thing. A companion licensing-scan fix this story also touched gained one more safeguard: it now tells
"everything checked out" apart from "could not check," rather than letting the two look the same.

What it still does not do: it does not relabel every internal error, does not change how failures
leave the library, does not touch the expression engine, and builds nothing that shows a label to a
person — that is a later story's job. Left knowingly rough for that story: the same missing character
repeated many times in one place currently produces one complaint per occurrence, not one per place.

Two things look wrong on purpose. One test is *supposed* to fail — a long-standing, deliberately unmet
coverage floor from Epic 2 — and still does, unchanged. A licensing scan that used to fail for three
unrelated, machine-local reasons is genuinely fixed here, in a way chosen so nobody could close it the
easy way.

---

## Story

As an integrating Go developer,
I want every failure to tell me which element and which data path caused it,
So that debugging a template is reading one message rather than bisecting a document.

---

## Do not re-open — settled rulings this story inherits

1. **The channel out of `Render`.** D-2.8.1, D-2.8.2, **D-2.8.3**, D-2.8.5, D-2.8.6. `Render` returns
   `(Result, error)`; `Result{Bytes, Diagnostics}`; `RenderTo` returns `([]Diagnostic, error)`.
   `Diagnostics` carries **warnings only**, in document order, `nil` when empty. **3.6 does not add a
   third return, a sink parameter, or a sibling entry point.**
2. **3.3's wire.** D-3.3.6. The caveat travels on `BindTextSpans`/`Resolve`'s third return as a typed
   `expr.Caveat`, never as an `error`. **3.6 does not re-plumb it.** The only permitted edit to the
   3.3 wire is that `folio.DiagCodeEmptyAverage`'s *value* becomes a registry member (the constant's
   name, its string, and every existing assertion on it stay identical).
3. **`internal/expr` is not edited.** Item 3's *"leave the sentinel in `expr`"*, read against
   divergence 1: `caveat.go` is untouched. `expr.CaveatKind` remains `diag`'s key.
4. **The `avg`-over-empty disposition.** D-3.3.4 part 2: the aggregate resolves to empty and the
   render proceeds. Story 4.2's AC forces it. Not reopened.
5. **DW-17 is not an accepted gap against AD-14** (D-000.49, DW-17's own paragraph). 3.6 mints; 3.7's
   CLI is where a `Diagnostic` reaching a human becomes observable. **3.6 does not build a
   presentation surface and must not claim the presented-output property.**
6. **The stage-rank table's ranks are ratified as measured** (D-000.16, re-measured at Story 2.5).
   3.6 does not re-rank `diag`.
7. **D-000.4's per-epic heavy-test cadence.** Epic 3's boundary matrix run is due **after 3.7**, and
   now carries Story 3.5's outstanding obligation: the `hidden-image` matrix document's native leg
   ran; its other three targets are deferred to that gate. **3.6 does not run the matrix.**
8. **D-3.6.3, -2 and -3 are ruled** (see *Rulings received at creation*). AD-14 forces
   D-3.6.3's arm A; the "label `SeverityError` unexercised" arm is **not** an available fallback
   and does not go to the owner (D-000.2 — an escalation blocks on the owner's presence, and there
   is no reason on the table to amend AD-14). If arm A proves genuinely infeasible, that returns to
   the **lead** as a finding.
9. **Missing glyph is a Warning** (divergence 6, four sources). Not re-argued. The only open
   question is OPEN-1: what the render *emits* for the uncovered rune.

---

## R — design constraints derived from the record during creation

**R1 — `internal/diag` is a rank-1 leaf holding the REGISTRY, not the types and not the mapping.**
From divergences 2 and 3, ruled at D-3.6.4. It imports **no first-party package** (rank 1 may
import only rank 0 `geom`, and it needs nothing from geometry either), and **the zero-imports
property is a requirement, not an accident**: AC4's five conditions arise in `template`(2),
`expr`(3), `bind`(4), `fontset`(6) and `layout`(7) — five packages spanning the whole rank range —
and **only a rank-0/1 leaf is importable by all of them.** The first first-party import added to
`diag` quietly forecloses one of those five call sites, so it is asserted (AC1), not merely
observed. It holds: a `Code` defined type, the code constants, and the registry value itself. `folio.Diagnostic`, `folio.Severity` and `folio.Result`
stay at the module root, unmoved. `folio.DiagCodeTextClippedWidth` and `folio.DiagCodeEmptyAverage`
keep their names and their exact string values and become `= string(diag.CodeX)`, so the public API
is byte-identical and the two spellings can never drift.

**R2 — the registry is a CONSTRUCTED value, and membership is what gets asserted.**
D-1.4.2's own row (decision log `:9118`): *"Assert **membership in the registry as constructed**, not
the existence of two string constants — a constant nothing registers is not a code."* So a bare
`const` block is not a registry. There must be a value — a slice or map the package builds — with a
lookup, and the DW-6 replacement assertion queries **that value**.

**R3 — the additive-only invariant is enforced by pinning each code's STRING to a test-owned
literal, never by pinning the registry's SIZE.**
D-000.68's discriminator: *"Pin to a literal when the set is permanent; state it relationally when it
is scheduled to move."* The registry is **both**: each shipped code's *meaning and spelling* is
permanent (that IS the additive-only rule), while the *set* is scheduled to grow — FR25's over-tall
rows land in Epic 4, and Epics 5–6 add more. Applying one half of the discriminator to the whole
registry gives the wrong instrument either way: a pinned count reddens on every legitimate addition
(and D-000.68 already ruled *"a count is a lossy set"*), while a purely relational statement cannot
see a code's meaning being repurposed, which is the exact breaking change AC5 names.

The instrument that fits: a test-owned table of `{constant identifier → exact string literal}`
covering **every** currently-shipped code, asserted against the registry as constructed. Adding a new
code does not redden it. Changing a shipped code's string **does**. Adding a code without adding its
pin is caught by a second assertion in the same test: every registry member must appear in the
table. **The anchor is a literal the test owns** — D-000.68's third anchor, named explicitly.

**The pinned table lives in the TEST FILE and is never derived from, generated from, or imported out
of the registry source.** If both sides move together the instrument is **tautological** — it asks
the code whether it agrees with itself and reports the inevitable yes as coverage. That is D-3.4.6's
date-token failure exactly, the case that produced D-000.68.

**R4 — a second anchor, from the compiler: `Code` is a defined type.**
`type Code string` in `diag`, so a naked string literal cannot be registered by accident and a
mistyped code does not silently become a new one. This is D-000.68's *"the compiler"* anchor and
costs nothing. **`folio.Diagnostic.Code` stays `string`** — changing a public struct field's type is
a larger public-API change than this story needs, and R1's `= string(diag.CodeX)` bridge keeps them
in lockstep. Record that as a deliberate bound, not an oversight.

**R5 — mint where the `Diagnostic` is CONSTRUCTED, when the condition CAN OCCUR.**
D-000.65, verbatim. Both DW-6 conditions occur in real documents today (divergence 7's verification
plus the two live sites below), so 3.6 mints **late**, not early — D-000.65's own consequence
paragraph says so. Do not read DW-6 as evidence the codes were correctly withheld.

**R6 — discharge `absence-diag-package` by REPLACEMENT, in the SAME commit.**
D-000.59, and DW-6's corrected anti-rot paragraph. In one commit: (a) remove the
`absence-diag-package` row from `absenceChecks`; (b) land the positive assertion that the registry
**as constructed** contains `TABLE_FOOTER_SOURCE_UNRESOLVED` and `TABLE_FOOTER_SOURCE_FORBIDDEN`; (c)
update `TestAbsencesChecksIncludeAllThreeEntries` (name, comment and literal) to the remaining two.
**The cheapest fix to the red is to delete the rule; that is the hazard, and it is forbidden.**

**R7 — do NOT sweep the bare-error population, and the bound has a CRITERION rather than a "not
now".** D-000.67 part 2 says a targeted fix does not sweep *within a file*; it does not license
sweeping a module. Measured in this run: `internal/bind` and `internal/expr` alone carry **60+**
non-test `fmt.Errorf` sites, and inventory item 2 named only four categories within them — the
kernel's overflow, the alignment-spread breach, the digit-budget breach, and the per-element
absent/wrong-kind aggregate errors, i.e. **the Epic 3 arithmetic surface**, not every error in the
two packages.

**A bound without a reason gets re-litigated every story, so AD-14 supplies the reason. Ratified as
[[D-3.6.6]], and it goes in the code as a doc comment on the registry, not only here:**

> **A code exists for a failure mode a caller can act on — FR41's five, plus conditions a template
> author can cause and fix. An internal-invariant violation ("this should never happen") stays a
> plain error.**

Stated that way the remaining 60+ are **a different category**, not an accepted gap and not a
backlog (D-000.49), and a future story adding a code has a test to apply rather than a precedent to
argue from. The criterion also decides AC7's `default:` arm the other way — see R12.

**R8 — the two DW-6 conditions, located in this run.**
- `TABLE_FOOTER_SOURCE_UNRESOLVED` → `folio-go/folio_expr_validate.go:316-320`: a column with
  `footer: "sum"`/`"avg"`, `footerOf` omitted, and a `bind` that is not one of D-1.4.1's two
  derivable shapes.
- `TABLE_FOOTER_SOURCE_FORBIDDEN` → two sites in `folio-go/internal/template/parse_bands.go`, exactly
  matching D-1.4.2's parenthetical (*"`footerOf` with `count`; footer fields with no `footer`"*):
  `:408-410` (`footerOf`/`footerFormat` present with no `footer`) and `:419-421` (`footerOf`
  alongside `footer: "count"`). **One code, two sites** — the code names the condition, not the line.
  `internal/template` is rank 2 and may import `diag` at rank 1.

**R9 — the five FR41 subjects, each verified constructible in this run BEFORE the AC was written**
(D-000.50). None of these five ACs is written against a subject that does not exist.

**The split is 4 Errors + 1 Warning** (divergence 6). The severity column is not cosmetic: it decides
which wire each mode lands on, and the two wires are completely separate.

| FR41 mode | severity | subject, measured | today | wire |
|---|---|---|---|---|
| malformed template | **Error** | `internal/template.newLoadError` (`errors.go:29`), via `folio.LoadTemplate` (`folio.go:54`) | bare error, located, no code | D-3.6.3's error type |
| unresolvable binding | **Error** | `internal/bind/text.go:401` — *"%s path %q is absent from %s"* | bare error, located, no code | D-3.6.3's error type |
| invalid expression | **Error** | `folio_expr_validate.go:120` — *"element %s: %q is not a valid expression"* | bare error, located, no code | D-3.6.3's error type |
| unlayoutable content | **Error** | `internal/layout.OverflowError` (`paginate.go:143`, constructed `:344`) | typed error, located (element id, heights, kind), no code | D-3.6.3's error type |
| **missing glyph** | **Warning** | `render.go:934` `resolveRuneFace` | **aborts the render**; has the rune, has **no element id**; **zero test coverage** | **the existing `Result.Diagnostics` channel** |

`fontset`(6) is where `HasGlyph` lives (`fontset.go:555`) and `layout`(7) owns `OverflowError`, so
with `template`(2), `expr`(3) and `bind`(4) the five conditions span ranks 2–7 — R1's reason for
`diag` being a rank-1 leaf, restated as a measurement.

**R10 — DW-18's renumbering is NON-OPTIONAL in this story, and it is COUPLED to D-3.6.3's arm A.**
`SeverityWarning Severity = iota` makes `SeverityWarning == 0`, identical to the zero value of every
`Diagnostic{}` literal, so all three of `render_clip_diagnostic_test.go`'s `d.Severity !=
SeverityWarning` assertions compare a value to itself.

**Why it stops being deferrable here.** DW-18's own *"how we'd know it was forgotten"* reads: *"A
future `Diagnostic{…, Severity: SeverityError, …}` construction site ships with the `Severity:` field
accidentally omitted … and the render path silently returns it as a `Warning`."* **This story is that
future.** The moment arm A lands, a copy-paste from a Warning-shaped helper silently **downgrades an
Error to a Warning** — AD-14's disposition rule violated with nothing able to catch it. The failure
mode DW-18 was holding open arrives in the same commit that closes it.

The fix: an unexported `severityUnset Severity = iota` ahead of `SeverityWarning`, making
**`SeverityWarning` 1 and `SeverityError` 2**.
- **What it breaks:** any downstream that persisted, serialized or compared the *integer* value of a
  `folio.Severity`. Nothing in this repo does (all comparisons are against the named constants;
  verified across the three modules).
- **Why it is free now:** `folio-go/version.go` declares `Version = "0.0.0-dev"`, and `git tag`
  returns exactly one tag, `pre-email-rewrite` — **there is no `folio-go/v*` tag**. Nothing
  downstream can have pinned these values.
- **Why never again:** AD-22 — once `folio-go/v0.1.0` is cut, renumbering a public constant on the
  product's front door is a breaking change requiring `folio-go/v2`.
- `Severity.String()` gains a `severityUnset` arm (never a bare integer; `diagnostic.go:41-51`).

**R11 — the caller-side rule (AC4) is about code we do not own; label the bound (D-000.24).**
"Match on the code, never on message text" governs *callers*, and callers live outside the module.
What is assertable in-repo is two things, and only two:
- the **capability**: a caller can recover the code from every FR41 failure without reading
  `Message` — asserted positively, per failure mode;
- the **stability**: rewriting a `Message` does not change the code — asserted by mutation
  (rewrite the message text in the working tree, observe every code assertion stay green while a
  deliberately message-matching control assertion reddens).

What is **not** assertable is that any real caller obeys the rule. Say so in the story's own words
and in the doc comment; do not imply coverage over code we do not own. DW-17 already draws this exact
line for the presented-output half and explicitly declined an AST guard over call sites for a
related reason — do not re-litigate it here.

**R12 — the `default:`-arm defect (divergence 5) is a FALLBACK BUCKET, and a code alone does not fix
it.** Ratified as [[D-3.6.7]]. The arm conflates an **unhandled** case with a **handled** one and fails **open**: an unmapped
caveat kind emerges as a `Diagnostic` that is structurally indistinguishable from a real one except
for a field being blank, and blank is exactly what nobody notices. Three things together, not one:

1. **Give the arm a registered code naming an internal condition**, so an unmapped caveat is **loud
   rather than blank**. Note this does not contradict R7's criterion — the criterion governs whether
   an internal-invariant *error* gets a code; this is a `Diagnostic` that already exists and is
   already returned to a caller, and AD-14 requires every `Diagnostic` to carry a registry code. The
   choice is between a coded one and a codeless one, not between a coded one and a plain error.
2. **Land the positive assertion** that **no `Diagnostic` reachable from `Render`/`RenderTo` can
   carry an empty `Code`** — a property over every construction site present and future, not a
   repair of the one instance (D-000.67 part 2's unit-of-work rule applied to a *class*).
3. Keep the arm itself: AD-14 says **never a panic**.

AC1 then has teeth rather than a repaired instance. See AC7.

---

## Acceptance Criteria

Each AC names its **anchor** (D-000.68) and its **discriminating mutation**. An AC that passes
before implementation has been mis-written — fix the test, not the code.

---

**AC1 — `internal/diag` exists, is a rank-1 leaf with ZERO first-party imports, and holds a
constructed registry.** `folio-go/internal/diag/` contains a `Code` defined type (`type Code
string`), the code constants, R7's criterion as the registry's doc comment, and a **registry value**
the package constructs, with a lookup by `Code`. `ScanStageRank` passes over the new directory.
**The zero-first-party-imports property is asserted, not merely observed** (R1): it is what keeps
`diag` importable by all five of AC4's condition-owning packages, spanning ranks 2–7, and the first
import added forecloses one of them.
*Anchor:* the compiler (`Code` is a defined type; a bare string literal does not satisfy it) plus
`ScanStageRank`'s own table plus an AST assertion over `diag`'s import set.
*Mutation:* add `import ".../internal/expr"` to a `diag` file → **both** `ScanStageRank` and the new
import-set assertion must redden. Run it; a rank guard nobody defeated is not yet a guard.

**AC2 — the registry contains `TABLE_FOOTER_SOURCE_UNRESOLVED` and `TABLE_FOOTER_SOURCE_FORBIDDEN`,
asserted as MEMBERSHIP IN THE CONSTRUCTED REGISTRY, and `absence-diag-package` is removed in the
same commit.** (R6, DW-6, D-000.59, decision-log `:9118`.)
The `absenceChecks` list drops to **two** entries; `TestAbsencesChecksIncludeAllThreeEntries` is
renamed and its test-owned literal reduced to `absence-designer-project` and
`absence-source-date-epoch`, with its comment recording that 3.6 discharged the third **by
replacement**. The remaining rows' `desc` strings are re-read for staleness (D-000.37) — the
`absence-source-date-epoch` row names Story 3.7 by name and is still correct; confirm, do not assume.
*Anchor:* a literal the test owns.
*Mutation:* delete the two codes from the registry construction → the new positive assertion reddens.
A registry-lookup assertion that stays green with the codes removed is querying the constants, not
the registry.
*Forward record (D-000.67 part 1, NOT this story's fix):* `absences_test.go:56-74` proves an empty
`absenceChecks` yields zero findings and `ChecksEvaluated == 0` — **green, not red**. The mechanism
goes **slack** at zero, not loud. The schedule is 3 → **2** (this story) → 1 (3.7) → **0 (Story 5.1,
where the mechanism and its precondition must be REMOVED together, never decremented to zero)**. This
sentence must appear in `absences.go`'s own comment.

**AC3 — the two shipped codes are absorbed with their EXACT STRINGS, and 3.3's wire is not
re-plumbed.** `DiagCodeTextClippedWidth` and `DiagCodeEmptyAverage` keep their exported names and are
redefined as `= string(diag.CodeX)`. **`diagnosticFromCaveat` stays in `render.go`** (divergence 2 —
it cannot move), and `internal/expr/caveat.go` is not edited. Every existing assertion in
`render_clip_diagnostic_test.go` and `render_empty_average_diagnostic_test.go` passes unchanged.

**The exact string values are asserted as literals, not merely the bridge compiling.** AD-14 makes
changing a code's meaning a breaking change, and **a bridge that alters one byte is that breaking
change wearing a refactor** — invisible in a diff that looks like a tidy-up. Assert
`DiagCodeTextClippedWidth == "TEXT_CLIPPED_WIDTH"` and `DiagCodeEmptyAverage ==
"AGGREGATE_EMPTY_AVERAGE"` against literals the test owns.
*Anchor:* a literal the test owns, plus the compiler (one definition; the two spellings cannot
drift).
*Mutation, corrected per D-3.6.9 (Finding 8, QA review — the claim below was false and is corrected
to state only what holds):* change `diag.CodeEmptyAverage`'s string → **AC3's own literal assertion
(`diag_bridge_test.go`) and AC5's pin (`internal/diag`'s `TestRegistryIsAdditiveOnly`) both redden.**
The pre-existing 3.3 suites (`render_empty_average_diagnostic_test.go`) do **not**, and are not
expected to: they compare `d.Code` against the symbolic constant `DiagCodeEmptyAverage` itself, so
the comparison moves WITH the mutation — a wiring assertion (is the mapping picking the right code?),
not a string-value assertion. Per D-3.6.9: **AC5's pinned table owns the STRING-VALUE property; the
3.3 suites own the WIRING property. Neither is a substitute for the other.** Do not convert the 3.3
suites to literals to make them redden here — that would couple a wiring test to a string-value
decision it should not carry (D-3.6.9, D-000.15).

**AC4 — each of FR41's five failure modes produces a DISTINCT stable code, and a test asserts the
message LOCATES the problem. FOUR are Errors; ONE — missing glyph — is a WARNING.** Five subjects and
their severities exactly as measured in R9. For each: the code is a registry member, the five codes
are pairwise distinct, and the message names the element id (and, where the mode has one, the data
path or the offending rune).

**The four Error modes** (malformed template, unresolvable binding, invalid expression, unlayoutable
content) travel on D-3.6.3's `Diagnostic`-carrying error type, wrapped with `%w` so every existing
`errors.As` target keeps working.

**Missing glyph travels on the EXISTING `Result.Diagnostics` channel** (divergence 6) — D-2.8.3's
wire, unchanged, the same route the 3.3 caveats already take. `shapeSegments` gains a collector or a
third return in the shape `BindTextSpans` already uses. **No new machinery and no new public type for
this mode.** The element id is threaded into `resolveRuneFace`/`shapeSegments` so the diagnostic
carries it, per AD-8's *"the element id and the offending rune"*. **AD-8 is unmet today on both
counts** — it aborts, and it has no element id — **and this AC is what meets it.**
*Anchor:* a literal the test owns (the five expected code strings) plus, for the location half, an
assertion that the message contains the element id **as supplied by the test**, never as read back
from the error.
*Mutation, per mode:* remove the element id from the construction site → the location assertion
reddens.
*Baseline reds to observe first:* for missing glyph, **both** halves are impossible at `4ec1884` —
the render aborts (so no `Result` exists to inspect) and no element id exists at that site. Observe
both, and record that `grep -rn "no font in chain"` returns exactly one hit, the production site, so
**no existing assertion is broken by the abort→warning conversion.**

**AC5 — additive-only is enforced by a test-owned pin per code, and growth does not redden it.**
(R3.) A table of `{constant identifier → exact string literal}` covering every shipped code, asserted
**both** directions against the registry as constructed: every table entry is a registry member with
that exact string, and every registry member appears in the table. **The table lives in the test file
and is never derived from, generated from, or imported out of the registry source** — if both sides
move together the instrument is tautological, which is D-3.4.6's date-token failure.
*Anchor:* a literal the test owns. Named as such in the test's own comment, per D-000.68's
*"every new guard names its anchor and states which of the three it is."*
*Mutations, all three run and all three recorded:*
(a) change a shipped code's string → **reddens** — the breaking change AC5 exists to catch;
(b) add a code to the registry without adding its pin → **reddens** — the second direction;
(c) add a code **with** its pin → **must stay GREEN.**
**(c) is not optional and is not a formality.** It is the inverse of a red-proof: it proves the guard
discriminates *addition* from *alteration*, i.e. that it will not fire on legitimate work. A guard
that reddens on every correct addition is a guard that gets switched off — D-000.15's erosion
dynamic, guarded directly. If (c) reddens, the instrument is a size check and must be replaced.

**AC6 — `Severity`'s zero value stops being a valid severity, and a test can now prove the field was
set.** (DW-18, R10 — **non-optional in this story, and coupled to AC8**.) `severityUnset Severity =
iota` precedes `SeverityWarning`; `SeverityWarning` becomes **1** and `SeverityError` becomes **2**;
`Severity.String()` gains its arm. Every production construction site sets `Severity` explicitly.
`render_clip_diagnostic_test.go`'s three sites lose their recorded-limitation comments and gain real
assertions. DW-18 is **retired** in `deferred-work.md` with the numbers, the tag evidence, and the
"free now, never again" sentence.
*Anchor:* the type system (the zero value is no longer a member of the valid set).
*Mutation:* DW-18's own **M8** — delete `Severity: SeverityWarning,` from `render.go`'s clip
construction site. At `4ec1884` this is invisible; after AC6 it must redden. **Run M8 at baseline
first and record that it passes**, so the before/after is a measurement and not a claim.
*The coupling, stated in the test's comment:* AC8 constructs the first `SeverityError` values in the
module's history. Until AC6 lands, an omitted `Severity:` field on one of them reads as
`SeverityWarning` and **silently downgrades an Error to a Warning** — DW-18's own recorded failure
mode, arriving in the same commit that would otherwise have left it open.

**AC7 — no `Diagnostic` reachable from `Render`/`RenderTo` can carry an empty `Code`.**
(R12, divergence 5.) Three parts, because a code alone repairs an instance and leaves the class:
1. `diagnosticFromCaveat`'s `default:` arm (`render.go:539-547`), which today returns an **empty**
   `Code`, is **retained** (AD-14: never a panic) and given a **registered code naming an internal
   condition**, so an unmapped caveat kind is **loud rather than blank**.
2. A **positive assertion over the class**: no `Diagnostic` any `folio` entry point can return
   carries a `Code` outside the registry — asserted over every construction site, present and
   future, not over the three that exist today.
3. R7's criterion is recorded on the registry so the next reader knows why the arm gets a code while
   the 60+ internal errors do not: the arm produces a `Diagnostic` that is **already returned to a
   caller**, so the choice is coded-vs-codeless, not coded-vs-plain-error.
*Anchor:* a literal the test owns (the registry), applied to every construction site.
*Mutations:* set a construction site's `Code` to an unregistered string → reddens. Add a second
`expr.CaveatKind` member without a matching arm → the `default:` arm fires with its new code, and the
diagnostic names the unhandled kind rather than arriving blank.

**AC8 — `SeverityError` gets its first constructor: a public `Diagnostic`-carrying error type.**
(D-3.6.3, ruled arm A — **forced by AD-14, not preferred**.) A public error type carrying one
`Diagnostic` with `SeverityError`, implementing `error` and `Unwrap`, recoverable by `errors.As`,
wrapping each of AC4's **four Error** modes with `%w` so **no existing error type, message or
`errors.As` target changes** — `layout.OverflowError`, `internal/template`'s load error and
`expr.KernelOverflowError` all keep working, verified by their existing suites passing unchanged.
The render **aborts** (non-nil error, no bytes) and `Result.Diagnostics` remains **warnings-only**
(D-2.8.3 not reopened; D-3.3.6's wire not reopened).
*Anchor:* the type system (`errors.As` against a concrete type).
*Mutations:* (a) return the diagnostic on `Result.Diagnostics` instead of as an `error` → the
"aborts" assertion (bytes absent) must redden; (b) for each of the four, assert the pre-existing
`errors.As` target still resolves through the new wrapper — a wrapper that breaks
`errors.As(err, &overflowErr)` has replaced the error, not wrapped it.
**Do not mint codes onto a severity nothing produces** (D-3.6.1, D-000.9) — this AC is what stops
that being the outcome.

**AC9 — a caller matches on the code, never on message text — with the bound stated honestly.**
(R11, D-000.24.) Two positive assertions: (a) for each of AC4's five modes, the code is recoverable
without reading `Message` — via `errors.As` for the four Errors, via `Result.Diagnostics` for the
Warning; (b) a message-text mutation leaves every code assertion green.
Additionally: the doc comment on `Diagnostic.Message` (already: *"never parsed by this library"*) is
extended to state the caller contract, and the story records in writing that **no in-repo assertion
can constrain callers outside the module** — labelled as a bound, never counted as coverage. DW-17
already draws this line for the presented-output half and explicitly declined an AST guard over call
sites for a related reason; do not re-litigate it.
*Anchor:* a literal the test owns (the expected codes), with `Message` deliberately excluded from
every assertion.
*Mutation:* rewrite each of the five messages in the working tree → all code assertions stay green;
a deliberately message-matching **control** assertion reddens, proving the mutation was live and the
green is not vacuous.

**AC10 — DW-19 is fixed at its specified fix shape.** (D-3.6.5, ruled: fix it here.)
`ResolveAssets` (`lint/internal/manifest/manifest.go:142`) treats *"the asset root resolved to a path
with **zero tracked files**"* as a **scan error**, returned and assessed **before any findings**
(D-1.3.3 amended, D-000.58). The three named tests go green **without** weakening any rule,
**without** adding files to `.font-sources/`, and **without** a `t.Skip` — both of the latter trade a
loud environmental failure for a quiet one, which D-000.58 declined for `.fontgen-venv`. `lint`'s
fail count goes to **0** (Finding 9, QA review — corrected from a pinned absolute pass-count, which
went stale twice within this story alone as regression tests were added: 104/3 baseline → 109/0 at
review → 111/0 after the finisher's Finding 1/7 tests; D-000.68's discriminator applies — the fail
count is the permanent property, the pass count is scheduled to grow with every regression guard this
area gains). DW-19 is retired in `deferred-work.md` with its measurement.
*Anchor:* git's own index — a fact the scanned directory cannot move.
*Mutation:* point the resolver at a **tracked** directory carrying a real violation → it must still
report the violation. A new error path that swallows real findings has traded one blind spot for
another.

**AC11 — the spine's source tree is corrected in this commit.** (Divergence 3, D-000.6.)
`ARCHITECTURE-SPINE.md:613` currently reads *"`diag/` — Diagnostic, Severity, the code registry
(AD-14)"*. Two of those three were sited at the module root by **D-2.8.3, an owner decision**, and
the tree line was never updated. Corrected to: **`diag/` holds the closed code registry (AD-14);
`Diagnostic`, `Severity` and `Result` live in the root `folio` package per D-2.8.3.** The line is
amended, not the decision — moving the public types to satisfy a stale document would reopen an owner
decision, which is why D-3.6.4's option C was rejected.

**AC12 — the three-module gate, with the one required red unchanged.**
`folio-go`: build ✅ vet ✅ `$(go env GOROOT)/bin/gofmt -l .` empty ✅, and
`TestCorpusMeetsP6ExerciseFloors` **still red** with **byte-identical** stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. `lint`: **0 fail** (AC10; Finding 9, QA
review — the pass count is not pinned here for the reason stated at AC10, and is reported as
measured in the Delivery Log instead).
`hashmatrix`: 3 pass. Any movement in the P6 stats is a **finding**, not a success.

---

## Rulings received at creation — D-000.69, D-3.6.2 … D-3.6.7

All three creation decisions were **ruled by the engineering lead before development starts**. They
are recorded here with their grounds, because the grounds change what the developer may do when a
detail turns out harder than expected. **OPEN-1 is new, raised by the correction to divergence 6,
and is the only one still open.**

---

### D-3.6.3 — RULED: **arm A**, and it is FORCED by AD-14, not preferred

A public `Diagnostic`-carrying error type wrapping each of AC4's **four Error** modes with `%w`.

**The grounds are rule text, not judgement.** AD-14: *"one `Diagnostic` value carries `Severity` … a
stable string code from a closed registry … **Every failure mode named in FR41 has a code** and a
test asserting it."* A code is a **field of `Diagnostic`**; a bare `fmt.Errorf` has no fields. So
FR41's modes must produce `Diagnostic`s, and the aborting ones must carry `SeverityError`. **Arm A
reconciles AD-14 with Go's error-return convention rather than departing from either.**

**Arm B — "label `SeverityError` unexercised" — is not going to the owner, and the reason is worth
recording.** B needs a **scope cut** *and* an **AD-14 amendment**, and D-000.2 makes every escalation
a block on the owner's presence. Asking the owner to choose between following the architecture and
amending it, with no reason on the table to amend, is not a decision — it is an interruption.
B was also falsified on the merits, and **more thoroughly than this story's first draft argued**: the
draft claimed all five modes are Errors, which is wrong (divergence 6). The correct count is that
AC2's *"`Error` aborts it"* has **four** subjects — not zero, and not five.

**Arm C** (construct `SeverityError` but keep it caller-unreachable) stays rejected as D-000.9's
shape at the type level.

**If the developer finds arm A genuinely infeasible, that returns to the LEAD as a finding** — it
does not become a licence to fall back to B.

**Coupled consequence: DW-18 (AC6) is now non-optional.** See R10. Arm A constructs the first
`SeverityError` values in the module's history, which is the exact moment DW-18's recorded failure
mode — an omitted `Severity:` field silently downgrading an Error to a Warning — becomes live.

---

### D-3.6.4 — RULED: **arm A**, and the low rank is the REQUIREMENT, not a nice-to-have

`internal/diag` is a **rank-1 registry leaf**: `type Code string`, the constants, the constructed
registry, **zero first-party imports**. `Diagnostic`/`Severity`/`Result` stay in root `folio`. The
mapping stays in `render.go`. The two shipped constants bridge as `= string(diag.CodeX)`.

**The lead strengthened the creator's argument.** AC4 requires all five FR41 modes to name a code,
and those conditions arise in **`template`(2), `expr`(3), `bind`(4), `fontset`(6) and `layout`(7)** —
five packages spanning ranks 2 through 7. **Only a rank-0/1 leaf is importable by all of them; every
higher placement forbids exactly the imports AC4 needs.** So the rank is not a constraint the design
works around — it is the property that makes the design possible, and it is asserted (AC1) rather
than left to be observed.

**Two guardrails, both promoted to ACs:**
- **AC3** — the bridge must preserve the **exact string values**. Changing a code's meaning is a
  breaking change under AD-14, and **a bridge that alters one byte is that change wearing a
  refactor.** Assert the literals, not merely that the bridge compiles.
- **AC1** — assert `diag`'s **zero first-party imports**. That property is what keeps it importable
  by all five; the first import added quietly forecloses one of the call sites.

**Divergence 3 — amend the spine in this commit (AC11), under D-000.6.** `:613` becomes: `diag/`
holds **the closed code registry (AD-14)**; `Diagnostic`, `Severity` and `Result` live in the root
`folio` package per D-2.8.3. **Option C is upheld as rejected** — moving the public types to satisfy
a stale document would reopen an owner decision.

---

### D-3.6.5 — RULED: **fix DW-19 in this story** (AC10)

The lead's grounding already routed DW-19 to *"the first Epic 3 story that touches `lint`"*, and 3.6
is that story. **The `rules/` vs `manifest/` distinction is a technicality against a deferral whose
owner clause — *"whoever next touches `ResolveAssets`"* — names a category no story is scheduled to
enter**: the unowned-deferral shape already recorded against DW-16. D-000.55's attributability
argument decides it — 3.6 lands substantive new `lint` rules, and its own lint measurements are
unattributable while three unrelated reds stand.

**The fix shape is specified in DW-19 and must not drift:** *"the asset root resolved to a path with
zero tracked files"* is a **scan error**, returned and assessed **before any findings**. **Explicitly
not** adding files to `.font-sources/`, and **not `t.Skip`** — both trade a loud environmental
failure for a quiet one, which D-000.58 declined for `.fontgen-venv`.

---

### OPEN-1 — **STILL OPEN.** What does the render EMIT for an uncovered rune?

**Raised by the correction to divergence 6, and not settled by any source read during creation.**

Missing glyph is a **Warning**, so per AD-14 it *"accompanies successfully returned PDF bytes"* — the
render must produce a document. **What does that document contain where the uncovered rune was?** Not
one of the four sources that establish the severity answers this:
- **AD-8** forbids *"a blank box"*, which rules out emitting `.notdef` — the one thing the code would
  do by default.
- **EXPERIENCE.md:216** and **UX-DR22** say the diagnostic *"locates the offending element"*; neither
  describes the glyph.
- **Story 5.12's AC** presents the warning; it does not say what was rendered.

This is a **byte-output decision**: a document that fails today will produce bytes afterwards, and
whatever is emitted becomes part of a golden hash under AD-21/AD-22. It cannot be left to the
developer.

**Option A (recommended) — omit the rune entirely.** No glyph is drawn and no advance is contributed;
the Warning names the element id and the rune. Satisfies AD-8's letter (no box is drawn), keeps the
render successful, and is the only arm where the diagnostic is the *sole* record of the rune — which
is what makes surfacing it non-optional rather than cosmetic.
*Risk:* text silently shortens. That is precisely what the Warning exists to report, and 5.12's
"locate back to the element" is how a designer sees it.

**Option B — substitute a visible replacement (e.g. U+FFFD) from a face that covers it.** Makes the
gap visible in the artifact itself rather than only in the diagnostic channel. *Rejected as the
default:* it requires a guaranteed-coverage face in every chain (an AD-8 change), and a substituted
glyph is a silent content edit — the exact class AD-8's *"never a silent substitution"* language
elsewhere in `fontset.go:833` rejects.

**Option C — keep it an Error.** *Rejected:* contradicts four sources, and would leave Story 5.12's
first AC without a subject.

**Recommendation: A.** If the lead prefers B, AC4's missing-glyph half and AC12's golden expectations
both change, so **rule this before task 9**.

---

## Task breakdown

D-3.6.3, D-3.6.4 and D-3.6.5 are **already ruled** — do not re-open them. **OPEN-1 is open and blocks
task 9 only.**

1. [x] **Re-measure the baseline** — all three modules, before writing a line, with
   `$(go env GOROOT)/bin/gofmt`. Confirm 853/1/1, 104/3, 3/0 and the P6 stats byte-identical.
   Record in the Delivery Log (D-000.26).
2. [x] **Run the baseline mutations that must PASS now**, and record that they pass — a before/after
   never measured before is a claim, not a measurement:
   - DW-18's **M8**: delete `Severity: SeverityWarning,` from `render.go`'s clip site. Must be
     invisible at `4ec1884`.
   - The missing-glyph element-id assertion: must be impossible to satisfy (the render aborts, so
     there is no `Result` to inspect).
   - Confirm `grep -rn "no font in chain"` returns **exactly one hit, the production site** — so the
     abort→warning conversion breaks no existing assertion.
3. [x] **Raise OPEN-1 to the lead** (what the render emits for an uncovered rune). It blocks
   task 9 and AC12's golden expectations; everything else can proceed while it is open.
4. [x] **Write AC2, AC4, AC5, AC6, AC7 and AC9's tests FIRST and observe them fail.** Record each
   observed failure text. An AC in this set that passes before implementation is mis-written — fix
   the test, not the code.
5. [x] **Create `folio-go/internal/diag/`**: `Code`, the constants, the constructed registry, the
   lookup, and **R7's criterion as the registry's doc comment**. **No first-party imports.** Verify
   AC1's mutation defeats both `ScanStageRank` and the new import-set assertion.
6. [x] **Discharge `absence-diag-package` by replacement, in this same commit** (R6, AC2): remove
   the row, land the membership assertion, rename and re-literal
   `TestAbsencesChecksIncludeAllThreeEntries`, and add `absences.go`'s forward note on the
   3 → 2 → 1 → 0 schedule and the slack-at-zero property.
7. [x] **Bridge the two shipped codes** (AC3) — names unchanged, values asserted as **exact
   literals**, defined via `diag`. Confirm `render_clip_diagnostic_test.go` and
   `render_empty_average_diagnostic_test.go` pass **unchanged**.
8. [x] **Renumber `Severity`** (AC6, R10): `severityUnset` ahead of `SeverityWarning`, the `String()`
   arm, every production construction site explicit. Re-run M8 — it must now **redden**. Do this
   **before** task 11, so no `SeverityError` value is ever constructed while the zero value is still
   valid.
9. [x] **Convert missing glyph from an aborting error to a Warning on the EXISTING `Result` channel**
   (AC4), and thread the element id through `resolveRuneFace`/`shapeSegments`. **Blocked on
   OPEN-1** for what is emitted. Keep it a signature/collector change in the shape
   `BindTextSpans` already uses — **not** a refactor, and **not** D-3.6.3's error type.
10. [x] **Implement D-3.6.3's arm A** (AC8): the public `Diagnostic`-carrying error type, wrapping
    the **four** Error modes with `%w`. Verify each pre-existing `errors.As` target still resolves
    through the wrapper — a wrapper that breaks `errors.As(err, &overflowErr)` has replaced the
    error, not wrapped it.
11. [x] **Attach the codes**: AC4's five at the sites in R9, the two DW-6 codes at the sites in R8
    (one code, two sites for `FORBIDDEN`), and AC7's `default:`-arm code. **Do not sweep** the other
    60+ bare errors (R7) — and state in the Delivery Log how many sites were examined in each file
    touched (D-000.67 part 2).
12. [x] **Land AC7's class-level assertion** — no `Diagnostic` reachable from `Render`/`RenderTo`
    carries a `Code` outside the registry. Over every construction site, not the three that exist
    today.
13. [x] **Fix DW-19** (AC10) at its specified shape: zero tracked files → scan error, before any
    findings. Verify the resolver still reports a real violation in a **tracked** directory.
14. [x] **Amend `ARCHITECTURE-SPINE.md:613`** (AC11, D-000.6) — `diag/` holds the code registry; the
    types stay in root `folio` per D-2.8.3.
15. [x] **Run every discriminating mutation named in the ACs** and record each with its observed
    result: AC1's rank+import injection, AC2's registry deletion, AC3's byte-change (all three
    instruments must redden), AC5's three — **including (c) add-with-pin, which must stay GREEN** —
    AC6's M8, AC7's unregistered code and its added-caveat-kind probe, AC8's two, AC9's message
    rewrite with its control, AC10's real-violation probe. **A guard whose mutation was not run is
    not yet a guard** (D-000.68).
16. [x] **Retire DW-6, DW-18 and DW-19 in `deferred-work.md`** with their measurements; append to
    **DW-17** that 3.6 minted and 3.7 still owns the presented-output half.
17. [x] **Run the three-module gate** (AC12): the P6 red unchanged and byte-identical, `lint` at
    **107 pass · 0 fail**, `hashmatrix` 3 pass.
18. [x] **Set the story status to `review`.**

---

## Delivery Log

**Baseline re-measured, this run, before any code change (task 1):**
- `folio-go`: build ✅ vet ✅ `$(go env GOROOT)/bin/gofmt -l .` empty ✅. `853 pass · 1 fail · 1 skip`.
  `TestCorpusMeetsP6ExerciseFloors` red, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` —
  byte-identical to the story's recorded baseline.
- `lint`: `104 pass · 3 fail` — `rules/TestFontsAssetsNoticeRemovalRedProof`,
  `manifest/TestManifestUpToDate`, `manifest/TestResolveAssetsIncludesWordlist`. `git ls-files
  .font-sources` confirmed 0 tracked files.
- `hashmatrix`: `3 pass`.

**Task 2's pre-mutations, run and recorded before any implementation (both required to pass at
baseline, and both did):**
- DW-18's M8 (delete `Severity: SeverityWarning,` from `render.go`'s clip construction site): ran the
  full `render_clip_diagnostic_test.go`/`render_empty_average_diagnostic_test.go` suite (9 tests) with
  the field deleted — all 9 still passed. Confirmed INVISIBLE at baseline, exactly as DW-18 claimed —
  this is the direct evidence for why Task 8 (renumber `Severity`) must precede Task 10 (construct the
  first `SeverityError` values): the window DW-18 describes is real, not a hypothetical.
- Missing-glyph element-id assertion: confirmed impossible at baseline — `resolveRuneFace` aborts the
  render (via `fmt.Errorf`), so no `Result` exists to inspect an element id on.
- `grep -rn "no font in chain"` returned exactly one hit (`render.go:936`, the production site) — no
  existing test asserts the current abort, so converting it to a Warning could not break any existing
  assertion. Confirmed.

**OPEN-1 raised to the lead (task 3) and ruled before task 9 began.** Verdict: omit the rune — no
glyph, no advance, no in-band marker of any kind. Three obligations attached and folded into task 9:
(a) the Warning message must name the element id, the rune as both `U+XXXX` and its literal form, and
the searched chain; (b) a corpus-wide assertion that rendering any committed fixture produces ZERO
missing-glyph Warnings, with the AC4 coverage subject built as a synthetic, test-only template (never
a committed fixture); (c) DW-17 amended in place to record this Warning's sharpened stake (it is the
first Warning whose only manifestation is the diagnostic — worse than FR44's clipped content, which at
least leaves a visible truncation cue).

**Divergence found and recorded (not in the original SEVEN, discovered during task 7):** AC3's
mutation instructs that changing `diag.CodeEmptyAverage`'s string must redden "AC3's literal
assertion, AC5's pin **and** the existing 3.3 suites." Measured: the pre-existing 3.3 suites
(`render_empty_average_diagnostic_test.go`) compare `d.Code` against the symbolic constant
`DiagCodeEmptyAverage` itself, never against a literal string — so they move WITH the mutation and stay
green under it (a self-referential comparison, not an independent witness).
`grep -n "AGGREGATE_EMPTY_AVERAGE" render_empty_average_diagnostic_test.go` returns zero hits — no
literal string appears anywhere in that file. Verified the mutation: changing the registry string
reddened the new bridge test (`diag_bridge_test.go`) and `internal/diag`'s own `TestRegistryIsAdditiveOnly`,
but left the 3.3 suites green, exactly as measured. This is not a defect introduced by this story — the
3.3 suites were written before AC3 existed and were never designed to catch a bridge-level breaking
change; AC3/AC5's own literal-owned instruments are what actually see it, and they do. Recorded here
per the story's instruction to flag drift between the record and the code rather than silently editing
around it; `render_empty_average_diagnostic_test.go` was NOT modified (AC3 requires its assertions
pass unchanged, which they do).

**AC1's mutation (task 5), run and recorded:** added `import _
"github.com/panitw/folio/folio-go/internal/expr"` to `internal/diag/diag.go`. Both
`TestDiagPackageHasZeroFirstPartyImports` (`internal/diag_arch_test.go`, this story's new AST-scan
instrument) and `lint`'s `TestStageRankProductionScan` (`ScanStageRank`) reddened independently. File
restored; both green again.

**AC3's mutation (task 7), run and recorded:** changed `diag.CodeEmptyAverage`'s string to
`"AGGREGATE_EMPTY_AVERAGE_MUTATED"`. Reddened: `TestDiagCodeBridgePreservesExactStrings`
(`diag_bridge_test.go`) and `internal/diag`'s `TestRegistryIsAdditiveOnly`. Stayed green (see divergence
above): `render_empty_average_diagnostic_test.go`'s existing suites. File restored.

**AC5's three mutations (task 5/15), run and recorded, all in `internal/diag`:**
- (a) changed a shipped code's string (`CodeEmptyAverage`) → `TestRegistryIsAdditiveOnly` reddened on
  both directions (pinned string no longer a member; new string has no pin).
- (b) added a code to the registry without its pin (`CodeMutationProbeB`) → reddened (member with no
  pin entry; count mismatch).
- (c) added a code WITH its pin (`CodeMutationProbeC`, added to both `diag.go` and the test's own
  `wantCodes` table) → stayed GREEN, confirmed via `-v` output showing `PASS`. Files restored after
  each mutation; final state re-verified green.

**AC6's mutation, M8 re-run post-renumbering (task 8), recorded:** deleted `Severity:
SeverityWarning,` from `render.go`'s clip construction site again, after `severityUnset` landed.
`TestRenderAndRenderToDiagnosticsAgree` now REDDENS (`Severity:Severity(unset)` vs. want
`Severity:Warning`) — confirms DW-18 is genuinely retired, not merely renamed. File restored.

**Vertical-model reachability changed by this story — measured, not assumed (found while fixing
task 9).** `TestVerticalModelErrorPathsAreUnreachableThroughRender`
(`vertical_model_test.go`) previously asserted that a chain with NO member present in the FontSet
fails in `resolveRuneFace`, never in `verticalModel`. Converting missing-glyph to a Warning means
`resolveRuneFace` no longer errors for that condition, so when EVERY declared chain member is absent,
every rune in the element is now omitted and the element contributes zero face metrics — reaching
`verticalModel`'s previously-unreachable `len(metrics) == 0` path for the first time, through the
public `Render` entry point. This is the SAME located error as `TestVerticalModelRefusesAChainWithNoPresentFace`
already covered at the seam, now reachable one level higher because the layer that used to fail first
no longer does for this cause. Updated the test's own assertions and doc comment to record this rather
than treat it as a regression; all 9 `TestVerticalModel*` tests pass.

**AC4's coverage subject (task 9), rebuilt as the ruling required.** `TestMissingGlyphDiagnosticFiresOnUncoveredRune`
(`ac4_coverage_test.go`) previously asserted the render ABORTS for an uncovered rune (the old FR41
behaviour). Rewritten to assert the render SUCCEEDS with exactly one `DiagCodeTextMissingGlyph` Warning
naming the element id, `U+0E01`, the literal `ก`, and the chain `[Noto Sans]` — the same synthetic,
test-only template as before (never added to a committed fixture), satisfying OPEN-1's obligation (b).

**Corpus-wide zero-missing-glyph assertion (task 9), new (`missing_glyph_corpus_test.go`):**
`TestCorpusFixturesProduceNoMissingGlyphWarnings` re-renders all five committed acceptance fixtures
(the same five `first_baseline_acceptance_test.go`'s `baselineAcceptanceFixtures` walks) and counts
(derived, never a literal `len(...)==0`) `DiagCodeTextMissingGlyph` diagnostics across them — asserted
zero. `wrapped-text`'s pre-existing `DiagCodeTextClippedWidth` Warning is correctly NOT counted as a
violation, confirming the count is genuinely code-specific rather than a blanket
`len(Diagnostics)==0`.

**AC8's mutation (task 10), run and recorded:** temporarily made `ParseTemplate` swallow
`template.ParseDocument`'s error and proceed with an empty `*template.Document` (simulating a
Diagnostic emitted on a non-error channel instead of aborting). `TestFourErrorModesCarrySeverityErrorDiagnostics`'s
malformed-template subtest and `TestRenderErrorWrapsWithoutReplacing` both reddened as required (the
"aborts" presence precondition failed — no error was produced). File restored; suite re-verified
green.

**AC7's two mutations (task 12), run and recorded, both against the class-level assertions:**
- Omitted `Code:` from `diagnosticFromCaveat`'s `default:` arm → `TestNoDiagnosticCompositeLiteralOmitsCode`
  (AST scan over module-root production `.go` files) reddened, naming the exact file:line.
- Set `DiagCodeInternalUnhandledCaveat` to an unregistered string (`"NOT_IN_REGISTRY"`) →
  `TestAllProducedDiagnosticsCarryARegisteredCode` reddened (fired via the fabricated
  `expr.CaveatKind(99)` probe that exercises the `default:` arm directly, since it is otherwise
  unreachable through the closed `CaveatKind` set). Both files restored; suite re-verified green.

**R7 compliance (task 11), stated per file, per D-000.67 part 2's unit-of-work rule:**
- `internal/bind`: 0 additional sites coded beyond the one boundary wrap in `render.go` (the package
  itself is untouched — its errors are wrapped at the render.go call boundary, not inside `bind`).
- `internal/expr`: untouched, as required ("Do not re-open").
- `folio_expr_validate.go`: 3 sites coded (`checkVisibleIfExpression`'s call site,
  `checkTextExpressions`'s two call sites for element values and column binds, and the
  `TABLE_FOOTER_SOURCE_UNRESOLVED` `!derivable` branch). `checkStyleHasNoPlaceholders` and
  `checkStyleStringHasNoPlaceholder`'s call sites were examined and deliberately left uncoded — a
  style field illegally carrying a placeholder is a different condition from "invalid expression" and
  is not named by FR41's five nor DW-6's two.
- `internal/template`: 2 sites coded (`parse_bands.go`'s two `TABLE_FOOTER_SOURCE_FORBIDDEN` sites, via
  the new `newLoadErrorCoded` constructor). The other ~104 `newLoadError` call sites in this package
  were left uncoded — they fall back to `DiagCodeTemplateMalformed` generically at the
  `wrapTemplateError` boundary in `render_error.go`, so no other call site needed editing.
- `render.go`: 2 boundary sites wrapped (`bind.BindTextSpans`'s error, `layout.Paginate`'s error, both
  called twice for the latter — content pass and paginateDocument).
- No sweep of the 60+ bare `fmt.Errorf` sites in `internal/bind`/`internal/expr` beyond what R7/R8/R9
  named.

**DW-19 fix, one deliberate deviation from the story's literal wording, recorded per D-000.26 ("the
code is the truth" / flag drift rather than silently resolve it):** the story's task/AC text describes
the zero-tracked-files condition as returned as a hard error "before any findings." Implementing it
that way literally (a global pre-pass that aborts `ResolveAssets` the moment ANY directory has zero
tracked files, before any other directory's findings are assessed) reintroduced the exact masking
defect DW-19's own AC10 forbids: `.font-sources` sorts first alphabetically, so its scan error would
have fired before ever reaching a real, tracked-directory violation elsewhere — silently breaking
`TestFontsAssetsLicenceRemovalRedProof`, which (measured) was ALREADY vacuously passing at baseline for
exactly this reason (its expected substring "no LICENSE* file" happened to match `.font-sources`'s own
erroneous message, without the test ever reaching its real target, `notosansthai/`). The target pass
count in the story (`lint` 107 pass / 0 fail, meaning `TestManifestUpToDate` and
`TestResolveAssetsIncludesWordlist` — which call `ResolveAssets` on the real repo root, where
`.font-sources` is present locally — must succeed with NO error) also settles this: an untracked
directory cannot be a hard, propagated error in the general case, or those two tests could never pass
on this machine. The implemented fix instead EXCLUDES an untracked directory from `ResolveAssets`'
findings loop entirely (no row, no error for it) while leaving every tracked directory's real
violation fully live — verified by two new permanent tests
(`TestResolveAssetsExcludesUntrackedDirectoryWithoutError`,
`TestResolveAssetsStillReportsATrackedViolation`) against synthetic, temporary git repositories. This
satisfies AC10's own mutation instruction verbatim ("point the resolver at a tracked directory carrying
a real violation → it must still report the violation") and does not add files to `.font-sources/` and
does not use `t.Skip`, exactly as D-3.6.5 requires. `lint` result: **109 pass / 0 fail** (exceeds the
107/0 target by the two new permanent regression tests this fix required).

**Three-module gate (task 17), final run:**
- `folio-go`: build ✅ vet ✅ `$(go env GOROOT)/bin/gofmt -l .` empty ✅. `872 pass · 1 fail · 1 skip`.
  `TestCorpusMeetsP6ExerciseFloors` still red, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115
  P6g:7}` — byte-identical to the recorded baseline. (872 vs. the baseline's 853: 19 new tests added by
  this story, 0 removed, 0 broken.)
- `lint`: `109 pass · 0 fail` (exceeds the 107/0 target — see the DW-19 note above).
- `hashmatrix`: `3 pass`, unchanged.

**Finisher pass (QA review's 1 Blocker, 3 Majors, 13 Minors, 2 Nits — all 18 triaged; 17 FIXED, 1
DEFERRED). Full triage and per-finding resolution recorded in `## Finding Resolutions` below;
summary here is the re-measured gate only.**

Every discriminating mutation the QA review ran was re-run against the finisher's fixes and recorded
in the relevant Finding Resolution entry: Finding 1's all-untracked scan error, Finding 2's swap
(plus the original three), Finding 3's unregistered-code-on-the-missing-glyph-path (both AC7
instruments), and Finding 4's real production message rewrite. All four now redden as required, and
all four were restored and re-verified green afterward.

Three-module gate, re-measured after the finisher's fixes:
- `folio-go`: build ✅ vet ✅ `$(go env GOROOT)/bin/gofmt -l .` empty ✅. **876 pass · 1 fail · 1
  skip**, 16 packages. `TestCorpusMeetsP6ExerciseFloors` still red, stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — **byte-identical**, unchanged by this
  finisher pass (876 vs. the review's 872: 4 net additional test entries, from Finding 4's table
  gaining subtests and Finding 3/12's added assertions inside existing tests — no test removed, no
  test broken).
- `lint`: **111 pass · 0 fail** (+2 over the review's 109, from Finding 1's two new tests;
  AC10/AC12's text no longer pins this number — see Finding 9).
- `hashmatrix`: **3 pass**, unchanged.

Not run in this pass, and not owed by it (D-000.4: the boundary matrix runs per-epic, due after
3.7): Epic 3's cross-target matrix, including the three legs Story 3.5's `hidden-image` document
still owes. Recorded here again so the finisher's own gate does not read as implying full coverage.

One line worth carrying into the Delivery Log, from the engineering lead's ratification of the
DW-19 deviation ([[D-000.70]], [[D-000.58]]): the masking defect behind Finding 1's blocker was
found only because a standing red (DW-19's three `lint` failures) was fixed rather than skipped or
allowlisted — either of the cheap options would have left the hidden "all clear looks like could not
look" defect in place too.

## File List

**New files:**
- `folio-go/internal/diag/diag.go` — the closed code registry (AC1, AC5, R1–R4, R7).
- `folio-go/internal/diag/diag_test.go` — `TestRegistryIsAdditiveOnly`, `TestCodeIsADefinedType`.
- `folio-go/internal/diag_arch_test.go` — `TestDiagPackageHasZeroFirstPartyImports` (AC1's second
  anchor).
- `folio-go/render_error.go` — `RenderError` (D-3.6.3 arm A, AC8), `wrapTemplateError`,
  `wrapOverflowError`.
- `folio-go/render_error_test.go` — `TestFourErrorModesCarrySeverityErrorDiagnostics`,
  `TestRenderErrorWrapsWithoutReplacing`, `TestMessageRewriteDoesNotAffectCodeRecovery` (AC4, AC8, AC9).
- `folio-go/diag_bridge_test.go` — `TestDiagCodeBridgePreservesExactStrings` (AC3).
- `folio-go/diag_no_empty_code_test.go` — `TestNoDiagnosticCompositeLiteralOmitsCode`,
  `TestAllProducedDiagnosticsCarryARegisteredCode` (AC7).
- `folio-go/missing_glyph_corpus_test.go` — `TestCorpusFixturesProduceNoMissingGlyphWarnings` (OPEN-1's
  corpus obligation).

**Modified files:**
- `folio-go/diagnostic.go` — `Severity` renumbered (`severityUnset`, AC6); new
  `DiagCodeTableFooterSourceUnresolved`, `DiagCodeTableFooterSourceForbidden`,
  `DiagCodeTemplateMalformed`, `DiagCodeBindingPathAbsent`, `DiagCodeExpressionInvalid`,
  `DiagCodeContentUnlayoutable`, `DiagCodeTextMissingGlyph`, `DiagCodeInternalUnhandledCaveat`
  constants, all bridged via `internal/diag`; `DiagCodeTextClippedWidth`/`DiagCodeEmptyAverage`
  re-bridged (AC3).
- `folio-go/render.go` — `resolveRuneFace` returns `(name, found, err)` instead of erroring on no
  coverage; `shapeSegments` takes `elementID`, returns `(segs, diags, err)`, omits uncovered runes
  (OPEN-1); `formatFontChain`/`missingGlyphMessage` helpers; `diagnosticFromCaveat`'s `default:` arm
  coded (AC7/R12); bind/layout error sites wrapped in `*RenderError` (AC4/AC8).
- `folio-go/folio.go` — `ParseTemplate`'s `template.ParseDocument` error wrapped via
  `wrapTemplateError` (AC4/AC8).
- `folio-go/folio_expr_validate.go` — `checkVisibleIfExpression`/`checkTextExpressions`/
  `!derivable` call sites wrapped with `DiagCodeExpressionInvalid`/`DiagCodeTableFooterSourceUnresolved`
  (AC4/AC8, R8/R9).
- `folio-go/page_number.go` — `digitTableRun` updated for `shapeSegments`' new signature.
- `folio-go/internal/template/errors.go` — `LoadError` gains an optional `Code diag.Code` field;
  `newLoadErrorCoded` constructor added (R8).
- `folio-go/internal/template/parse_bands.go` — two `TABLE_FOOTER_SOURCE_FORBIDDEN` sites use
  `newLoadErrorCoded` (R8).
- `folio-go/ac4_coverage_test.go` — `TestMissingGlyphDiagnosticFiresOnUncoveredRune` rewritten for the
  Warning disposition (OPEN-1).
- `folio-go/render_clip_diagnostic_test.go` — three "known limitation" (DW-18) comments replaced with
  real coverage statements (AC6).
- `folio-go/vertical_model_test.go` — `TestVerticalModelErrorPathsAreUnreachableThroughRender` updated
  for the new reachability of `verticalModel`'s `len(metrics)==0` path.
- `folio-go/layout_probe_test.go`, `folio-go/tabular_digits_test.go`, `folio-go/wrap_test.go` — updated
  for `shapeSegments`' new signature.
- `lint/internal/manifest/manifest.go` — `ResolveAssets` excludes untracked asset directories
  (`gitTrackedFileCount`) instead of treating them as licensing violations (AC10, DW-19).
- `lint/internal/manifest/manifest_test.go` — two new tests,
  `TestResolveAssetsExcludesUntrackedDirectoryWithoutError`,
  `TestResolveAssetsStillReportsATrackedViolation`.
- `lint/internal/rules/absences.go` — `absence-diag-package` row removed; forward schedule note added
  (AC2, R6).
- `lint/internal/rules/absences_test.go` — `TestAbsencesChecksIncludeAllThreeEntries` renamed to
  `TestAbsencesChecksIncludeBothRemainingEntries`, re-literalled to the two remaining rule ids (AC2).
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` —
  `:613`'s `diag/` line amended (AC11, D-000.6).
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-6, DW-18, DW-19 retired; DW-17 amended
  in place (task 16).

**Modified by the finisher pass (QA review findings — see `## Finding Resolutions`):**
- `lint/internal/manifest/manifest.go` — Finding 1: the D-3.6.5 amendment's "Required" scan-error
  floor (`sawTrackedDirectory`).
- `lint/internal/manifest/manifest_test.go` — Finding 1: two new tests
  (`TestResolveAssetsAllDirectoriesUntrackedIsAScanError`,
  `TestResolveAssetsNoCandidateDirectoriesIsNotAScanError`); Finding 7:
  `TestResolveAssetsStillReportsATrackedViolation` rebuilt with a co-present untracked directory;
  `TestResolveAssetsExcludesUntrackedDirectoryWithoutError` updated with a co-present tracked,
  compliant directory (required by Finding 1's floor).
- `folio-go/internal/diag/diag_test.go` — Finding 2: `wantCodes` replaced by `codePins`, binding each
  named constant to its pin; Finding 18: `TestCodeIsADefinedType` replaced by a compile-time
  assertion.
- `folio-go/diag_no_empty_code_test.go` — Finding 3: `missingGlyphTemplateJSON` added; the missing-glyph
  path added to `TestAllProducedDiagnosticsCarryARegisteredCode`; the AST scan extended to check
  registry membership of literal `Code` values.
- `folio-go/render_error_test.go` — Finding 4: `TestMessageRewriteDoesNotAffectCodeRecovery` rebuilt as
  a 5-mode table with production-text controls; Finding 12: pairwise distinctness extended to the
  missing-glyph code.
- `folio-go/ac4_coverage_test.go` — Finding 17: stale filename cross-reference corrected.
- `folio-go/missing_glyph_corpus_test.go` — Finding 11: fixture-table comparison changed from a length
  check to a two-way name-set comparison.
- `lint/internal/rules/absences.go` — Finding 10: forward-schedule comment corrected to name the test
  that actually exists.
- `lint/internal/rules/stagerank_test.go` — Finding 14: `"diag"` added to the vacuity guard's scanned
  package list.
- `_bmad-output/implementation-artifacts/deferred-work.md` — Finding 5: DW-19's `Retired at`/`Status`
  corrected; Finding 16: DW-19's "the latter two were masked" corrected to the singular; Finding 13:
  DW-17 amended in place with the per-occurrence-emission bound.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — Finding 6: a forward-pointer
  paragraph added to the original D-3.6.5 entry; Finding 15: a new `D-000.6 amendment (Story 3.6,
  finisher pass)` entry added.
- This story file — Finding 8: AC3's mutation clause corrected; Finding 9: AC10/AC12 restated
  relationally; plain-terms opener refreshed; `## Finding Resolutions` and this addendum added;
  `Status` set to `done`.

## Change Log

- Created `internal/diag` as a rank-1, zero-first-party-import registry leaf holding `Code`, the
  code constants, and the constructed registry value (AC1, R1–R4).
- Discharged `absence-diag-package` by replacement in the same commit: removed the row, landed the
  positive registry-membership assertion, renamed and re-literalled the covering test (AC2, R6).
- Bridged `DiagCodeTextClippedWidth`/`DiagCodeEmptyAverage` to `internal/diag` with byte-identical
  exact strings; every existing 3.3/2.8 assertion passes unchanged (AC3).
- Renumbered `Severity` (`severityUnset` ahead of `SeverityWarning`), retiring DW-18 non-optionally
  ahead of minting the first `SeverityError` values (AC6, R10).
- Converted the missing-glyph condition from an aborting, uncoded error to FR41's fifth mode: a
  Warning on the existing `Result.Diagnostics` channel, naming the element id, the rune (`U+XXXX` and
  literal), and the searched chain; the render OMITS the uncovered rune (no glyph, no advance), per the
  engineering lead's OPEN-1 ruling (AC4, divergence 6).
- Added a corpus-wide assertion that no committed fixture ever produces a missing-glyph Warning, per
  OPEN-1's attached obligation.
- Implemented `RenderError` (D-3.6.3, arm A): the first publicly constructible `SeverityError` value,
  wrapping each of FR41's four Error modes (malformed template, unresolvable binding, invalid
  expression, unlayoutable content) without replacing any pre-existing error type, message, or
  `errors.As` target (AC4, AC8, AC9).
- Minted `TABLE_FOOTER_SOURCE_UNRESOLVED`/`TABLE_FOOTER_SOURCE_FORBIDDEN` (DW-6) at their named sites.
- Gave `diagnosticFromCaveat`'s `default:` arm a registered code and landed the class-level assertion
  that no `Diagnostic` reachable from `Render`/`RenderTo` can carry an unregistered `Code` (AC7, R12).
- Fixed DW-19: `ResolveAssets` now excludes an untracked asset directory from its findings instead of
  reporting it as a licensing violation, while still reporting a real, tracked-directory violation
  (AC10) — `lint` moved from 104/3 to 109/0.
- Amended `ARCHITECTURE-SPINE.md:613` to reflect D-2.8.3's already-ratified placement of `Diagnostic`/
  `Severity`/`Result` at the module root (AC11).
- Retired DW-6, DW-18, DW-19 in `deferred-work.md`; amended DW-17 in place per the OPEN-1 ruling's
  third obligation.

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer (adversarial, Story 3.6 by number; no story auto-picked)
- **Date:** 2026-08-26
- **Baseline:** `4ec1884`, changes uncommitted. All mutations were run in throwaway `git worktree`s;
  the main working tree was never mutated (verified byte-for-byte before and after).
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 1
- **Majors:** 3
- **Minors:** 13
- **Nits:** 2

### Three-module gate, re-measured independently in this review

| module | measured here | story claim | verdict |
|---|---|---|---|
| `folio-go` | build ✅ vet ✅ `$(go env GOROOT)/bin/gofmt -l .` empty ✅ · **872 pass · 1 fail · 1 skip** · **16 packages** | 872/1/1 | ✅ matches |
| `folio-go` required red | `TestCorpusMeetsP6ExerciseFloors`, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` | byte-identical | ✅ **byte-identical**, still red |
| `lint` | **109 pass · 0 fail** | 109/0 | ✅ matches (AC10/AC12 text still says 107 — see Finding 9) |
| `hashmatrix` | **3 pass** | 3 | ✅ matches |

### AC-by-AC disposition

| AC | verdict |
|---|---|
| AC1 | **Satisfied, teeth verified.** Injecting `import _ ".../internal/expr"` into `internal/diag` reddened **both** `TestDiagPackageHasZeroFirstPartyImports` and lint's `TestStageRankProductionScan`, independently. `Diagnostic`/`Severity`/`Result` confirmed still in root `folio` (`diagnostic.go:24,229,291`); `diagnosticFromCaveat` confirmed still in `render.go:524`; no such type declared under `internal/`. |
| AC2 | Satisfied. `absenceChecks` at two entries, test renamed and re-literalled, replacement assertion landed in the same commit. Minor defect: Finding 10. |
| AC3 | Satisfied in substance; the AC's own mutation text is now known false and was not corrected — Finding 8. |
| AC4 | Satisfied for the four Error modes and the Warning; location asserted against test-owned literals. Gaps: Findings 13, 14. |
| AC5 | **Partially satisfied — Finding 2 (Major).** All three named mutations behave as specified, but the instrument cannot see repurposing. |
| AC6 | **Satisfied, teeth verified.** See Finding-free verification note below. |
| AC7 | **Partially satisfied — Finding 3 (Major).** Parts 1 and 3 land; part 2 does not cover the class it claims. |
| AC8 | Satisfied. `errors.As` targets verified to resolve through the wrapper; deleting `Severity: SeverityError` reddens all four modes. |
| AC9 | **Not satisfied as specified — Finding 4 (Major).** |
| AC10 | Satisfied at a **deviated** shape that is verified sound (see below), but the ratifying amendment's own "Required" clause is unimplemented — Finding 1 (Blocker). |
| AC11 | Satisfied. Spine `:613` amended correctly and minimally (only line touched in the whole spine). Findings 15, 17. |
| AC12 | Satisfied on measurement; stale numbers in the AC text — Finding 9. |

---

## The two self-reported divergences — both independently verified

### Divergence 1 — the DW-19 deviation from D-3.6.5's literal wording: **SOUND. Not a finding.**

All three of the developer's claims were reproduced from scratch at `4ec1884` in an isolated
worktree with `.font-sources/` recreated (3 font binaries, 0 tracked, `.gitignore:85`).

1. **Was `TestFontsAssetsLicenceRemovalRedProof` genuinely vacuous at baseline? — YES, proven.**
   With `folio-go/fonts/notosansthai/LICENSE-OFL.txt` **fully intact and unmutated**,
   `ResolveAssets` at `4ec1884` already returns:
   `.font-sources: contains a committed font binary but no LICENSE* file (AC25, AD-26)`
   That string contains the test's asserted substring `"no LICENSE* file"` and **never mentions
   `notosansthai`**. The red-proof passed without its own mutation having any effect.
2. **Does the literal reading actually mask? — YES, proven by implementation.** I implemented
   D-3.6.5 literally at baseline (a pre-pass returning a scan error before any findings). Result:
   `lint` goes **104 pass / 3 fail → 103 pass / 4 fail** — strictly worse. It newly breaks
   `TestFontsAssetsLicenceRemovalRedProof` and leaves `TestManifestUpToDate` and
   `TestResolveAssetsIncludesWordlist` still failing. AC10's 0-fail target is **unreachable** under
   the literal mechanism.
3. **Does the shipped alternative preserve the property? — YES, proven by construction.** With
   `.font-sources` present (it sorts first, `.` &lt; `f`) I created a **tracked** directory
   `folio-go/fonts/zzprobeface/` holding a committed font binary and no LICENSE file. The shipped
   resolver still reports it:
   `folio-go/fonts/zzprobeface: contains a committed font binary but no LICENSE* file (AC25, AD-26)`
   — the untracked directory neither masks it nor is masked by it.
4. **Bonus, unreported by the developer and worth crediting:** the shipped fix **de-vacuates** the
   red-proof. Under it, `ResolveAssets` returns `nil` with all licences intact, so
   `TestFontsAssetsLicenceRemovalRedProof` and `TestFontsAssetsNoticeRemovalRedProof` now both fail
   for their own reasons and pass for their own reasons. The deviation did not weaken the check —
   it is the first time either red-proof has been non-vacuous.

**Verdict: the deviation is sound and strictly strengthens the guard.** The engineering lead's
ratification (D-3.6.5 amendment) matches the facts. The separate blocker below concerns a clause
*inside* that ratification, not the deviation itself.

### Divergence 2 — D-3.6.4's bridge guardrail: **accurately self-reported. Verified.**

Mutating `diag.CodeEmptyAverage`'s string to `"AGGREGATE_EMPTY_AVERAGE_MUTATED"` and running the
whole `folio-go` suite (`-count=1`) reddens **exactly two** tests beyond the required P6 red:

- `TestDiagCodeBridgePreservesExactStrings` (AC3's literal assertion) — **reddens** ✅
- `internal/diag.TestRegistryIsAdditiveOnly` (AC5's pin) — **reddens** ✅
- the existing 3.3 suites (`render_empty_average_diagnostic_test.go`) — **stay green**, exactly as
  reported. Confirmed cause: the file contains **zero** literal code strings, and asserts
  `d.Code != DiagCodeEmptyAverage` (`:69`, `:282`, `:325`) — the symbol moves with the mutation.

**Constant-compared-against-itself census (D-000.68 tautology class): 9 sites.** Every comparison
of a produced `.Code` against a `DiagCode*` symbol, none of which can see a code-string change:

| # | site | expression |
|---|---|---|
| 1 | `render_empty_average_diagnostic_test.go:69` | `d.Code != DiagCodeEmptyAverage` |
| 2 | `render_empty_average_diagnostic_test.go:282` | `d.Code == DiagCodeEmptyAverage` |
| 3 | `render_empty_average_diagnostic_test.go:325` | `d.Code == DiagCodeEmptyAverage` |
| 4 | `render_clip_diagnostic_test.go:122` | `d.Code != DiagCodeTextClippedWidth` |
| 5 | `render_clip_diagnostic_test.go:144` | `d.Code != DiagCodeTextClippedWidth \|\| …` |
| 6 | `visibility_test.go:708` | `d.Code != DiagCodeTextClippedWidth` |
| 7 | `ac4_coverage_test.go:146` | `d.Code != folio.DiagCodeTextMissingGlyph` |
| 8 | `render_error_test.go:220` | `renderErr.Diagnostic.Code != DiagCodeExpressionInvalid` |
| 9 | `missing_glyph_corpus_test.go:104` | `d.Code == DiagCodeTextMissingGlyph` |

Only **2** assertions in the whole tree pin a public code to a test-owned literal
(`diag_bridge_test.go:16,19`), plus the 10 pins inside `internal/diag/diag_test.go`.
**None of the nine is a defect in itself** — their job is dispatch correctness, not string pinning,
and that is legitimate. They are listed because AC3's mutation text asserts three instruments will
redden when only two can, and because Finding 3 depends on site 7 being the *only* thing standing
behind the missing-glyph path.

---

## Verification log — mutations re-run by this review (not accepted as reported)

Every mutation below was executed by the reviewer, full suite, `-count=1`, in a worktree replica
verified faithful (872/1/1, same single required red). Files restored and re-verified by SHA-256
after each.

| mutation | required | observed | verdict |
|---|---|---|---|
| AC1: `import _ ".../internal/expr"` into `internal/diag` | both guards redden | `TestDiagPackageHasZeroFirstPartyImports` **red**; lint `TestStageRankProductionScan` **red** | ✅ |
| AC3/D-3.6.4: change `diag.CodeEmptyAverage` string | AC3 + AC5 red, 3.3 green | exactly that | ✅ |
| **AC5 (a)** change a shipped code's string | reddens | `TestRegistryIsAdditiveOnly` red, both directions | ✅ |
| **AC5 (b)** add `CodeProbeB` without a pin | reddens | red: *"registry member \"PROBE_B\" has no entry in this test's pin table"* + count mismatch | ✅ |
| **AC5 (c)** add `CodeProbeB` **with** its pin | **stays GREEN** | `--- PASS: TestRegistryIsAdditiveOnly` | ✅ **guard discriminates addition from alteration** |
| **AC5 extra (reviewer):** *swap* two shipped codes' strings | should redden | **NOTHING reddened** (872/1) | ❌ **Finding 2** |
| reviewer: re-point a public bridge to another registry member | — | caught, but only via pairwise-distinctness in `TestFourErrorModesCarrySeverityErrorDiagnostics` | ⚠️ incidental |
| **AC6/DW-18 M8** at baseline `4ec1884` | must be invisible | **291 pass, 0 fail** — invisible | ✅ |
| **AC6/DW-18 M8** after renumbering | must redden | **4 tests red** (`TestRenderAndRenderToDiagnosticsAgree`, `TestWidthOverflowDetectedPerElement` + 2 subtests) | ✅ **window closed** |
| reviewer: delete `Severity: SeverityError` from `newRenderError` | should redden | **all four Error modes red** | ✅ the dangerous site is covered |
| AC7: omit `Code` from the `default:` arm | reddens | **both** AC7 instruments red | ✅ |
| **AC7 extra (reviewer):** unregistered code on the **missing-glyph** path | should redden | **both AC7 instruments GREEN** | ❌ **Finding 3** |
| AC9: rewrite production message text in the working tree | codes green **and a control reddens** | codes green (872/1) but **no control fired anywhere** | ❌ **Finding 4** |
| AC10: tracked directory with a real violation, `.font-sources` co-present | still reports | reports `zzprobeface` correctly | ✅ |
| reviewer: all directories untracked, `dirOrder` non-empty | scan error (D-3.6.5 amendment) | **0 rows, `nil` error** | ❌ **Finding 1** |

### D-3.6.8 (the uncoverable rune) — verified at the byte level

Rendering the element value `"กX"` produces a PDF **byte-identical** to rendering `"X"`. That single
measurement discharges the whole ruling at once:

- **no glyph drawn** ✅
- **no advance contributed** ✅ (not merely no glyph — identical bytes means `X` sits at the
  identical coordinate)
- **no CID allocated** ✅ (an allocated CID would change the font subset and the bytes)
- **`/ToUnicode` never maps it** ✅ — inflating every Flate stream finds no `0E01`/`0e01` and no
  literal `ก`; extracted text therefore matches visible text
- the Warning names the element id (`e1`), the rune as `U+0E01` **and** its literal form `ก`, **and**
  the searched chain (`Noto Sans`) — all asserted against literals the test owns
  (`ac4_coverage_test.go:140-160`)

### Other items confirmed clean

- **AD-1 / AD-23:** no `float32`/`float64`/`math/big.Float`/`math/big.Rat` in new or changed non-test
  code; nothing new under `internal/` importing `time`, `os`, `math/rand`, `net` or `math`
  transcendentals; no dot-imports. `internal/diag`'s two package-level vars are unexported,
  never written after init, and `All()` returns a copy — the mandated constructed-registry shape,
  not mutable state. `registry` is never ranged; only `allCodes` (a slice) is, so no map iteration
  reaches output.
- **No sweep (R7/D-3.6.6):** `fmt.Errorf` counts in `internal/bind` (**31**) and `internal/expr`
  (**79**) are **identical** to baseline, per file. `internal/expr` and `internal/bind` are
  completely untouched by the diff.
- **D-000.69 / `caveat.go`:** `internal/expr/caveat.go` is byte-identical to `4ec1884` and still
  declares `Caveat`, `CaveatKind`, `CaveatEmptyAverage`. `grep 'Err[A-Z]'` over `internal/expr`
  returns **zero hits** — no sentinel, as required.
- **D-3.6.6's criterion is in the registry's doc comment**, not only in the story
  (`internal/diag/diag.go`, the `const` block's doc comment) ✅.
- **`internal/template` coding is targeted, not swept:** exactly 2 sites use `newLoadErrorCoded`;
  101 `newLoadError` call sites left uncoded. `template`(2) → `diag`(1) is legal (strictly lower).
- **DW-17 amended in place**, not retired; still open; still names Stories 3.7 / 5.12 / 6.6; the
  amendment records that this is the first Warning whose only manifestation is the diagnostic ✅.
- **DW-6 / DW-18 / DW-19 retired, none deleted or renumbered** (19 DW ids before and after,
  identical membership and order). DW-18 carries the numbers, the `git tag` evidence and the
  "free now, never again" sentence; independently re-verified (`version.go` = `0.0.0-dev`;
  `git tag` names only `pre-email-rewrite`).
- **Plain-terms opener** explains why a stable code matters more than a good message, in prose,
  with no identifiers ✅.
- **`digitTableRun`'s discarded diagnostics** (`page_number.go:383`) are safe: its
  `len(segs) != 1 || len(segs[0].glyphs) != 10` guard fires loudly for any uncovered digit
  (an all-missing run yields one segment with zero glyphs). Claim verified, not a finding.

---

## Findings

### Finding 1: D-3.6.5 (amendment)'s "Required" scan-error floor is unimplemented, untested, and untracked
- **Severity:** Blocker
- **Category:** AC Conformance
- **Location:** `lint/internal/manifest/manifest.go:82-83` and `:164`; ruling at
  `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:11038-11040`
- **Observation:** The same commit's ratification of the DW-19 deviation attaches a hard
  requirement: *"**Required: if `dirOrder` is non-empty but EVERY directory resolves to zero tracked
  files, that is a scan error** — returned and assessed before findings."* It is not implemented.
  `ResolveAssets` does `if tracked == 0 { continue }` and then `return rows, nil` with no
  `len(rows) == 0 && len(dirOrder) > 0` floor. I proved it reachable with a synthetic git repository
  holding two on-disk font directories, neither tracked: **`rows=0`, `err=<nil>`**, and
  `Render` prints *"No redistributed non-code assets are committed at this commit."* Neither new
  test covers this case, and a repo-wide grep finds the requirement **only** inside the decision-log
  amendment — it appears in no DW entry, no AC, and not in the Delivery Log, Change Log or File
  List. Its natural owner, DW-19, is marked **RETIRED**.
- **Impact:** Self-declared D-000.9 exposure — *"all clear" and "could not look" produce identical
  output* — reachable by an ordinary directory rename, a wrong `repoRoot`, or a layout refactor.
  A ruling issued in this commit and marked "Required" has no implementation, no test and no open
  record, so nothing will resurface it.
- **Suggested Resolution:** Either implement the floor (a post-loop scan error when `dirOrder` is
  non-empty and every directory resolved to zero tracked files) with a red-proof, or open a new DW
  entry owning it explicitly and cite that entry from D-3.6.5's amendment. Do not close the story
  with the requirement in neither state.
- **Related AC:** AC10

### Finding 2: AC5's pin cannot detect code repurposing — the {identifier → string} binding is never asserted
- **Severity:** Major
- **Category:** Tests
- **Location:** `folio-go/internal/diag/diag_test.go:22-33` (the `wantCodes` table) and `:41-61`
  (`TestRegistryIsAdditiveOnly`)
- **Observation:** AC5/R3 specify *"a test-owned table of `{constant identifier → exact string
  literal}`"*. The shipped table has that **shape**, but its **keys are never used in any
  assertion** — they appear only inside `t.Errorf` format arguments. All three assertions operate on
  the table's *values* as a set: every pinned value is registered; every registry member is pinned;
  the two lengths match. A set comparison is invariant under permutation. I swapped two shipped
  codes' strings in `diag.go` — `CodeTemplateMalformed Code = "BINDING_PATH_ABSENT"` and
  `CodeBindingPathAbsent Code = "TEMPLATE_MALFORMED"` — and ran the **whole** `folio-go` suite:
  **872 pass / 1 fail**, the fail being only the required P6 red. **Nothing reddened.**
- **Impact:** Swapping two codes' meanings is precisely AD-14's *"changing a code's meaning is a
  breaking change"*, and it is the breaking change AC5 names as its reason to exist. The guard is
  blind to it. Because `folio.DiagCode*` are `= string(diag.CodeX)`, the public API's two constants
  would silently trade meanings with no test in either module objecting.
- **Suggested Resolution:** Assert the map key against the constant it names, so the
  identifier→string binding is checked rather than decorative — e.g. a table of
  `{Code constant → literal}` pairs compared elementwise (`wantCodes["CodeTemplateMalformed"] ==
  CodeTemplateMalformed`), which restores the anchor AC5 actually specifies. Then re-run the swap as
  a fourth recorded mutation.
- **Related AC:** AC5

### Finding 3: AC7 part 2 does not cover the class it claims — the missing-glyph path is outside both instruments
- **Severity:** Major
- **Category:** Tests
- **Location:** `folio-go/diag_no_empty_code_test.go:110-150`
  (`TestAllProducedDiagnosticsCarryARegisteredCode`) and `:30-97`
  (`TestNoDiagnosticCompositeLiteralOmitsCode`); production site `folio-go/render.go:1046-1051`
- **Observation:** AC7 part 2 requires *"a property over every construction site, present and
  future, not a repair of the one instance"*. I set the missing-glyph construction site's `Code` to
  the unregistered literal `"MADE_UP_UNREGISTERED"` and ran **both** AC7 instruments: both **PASS**.
  Two independent reasons. (a) The runtime test enumerates four paths — clip, empty-average,
  `diagnosticFromCaveat` matched, `diagnosticFromCaveat` default — and **omits the missing-glyph
  path**, which is the one this story added and one of only five production Diagnostic producers.
  (b) The AST scan only checks that a `Code` key is present and not the literal empty string; it
  never checks **registry membership**, so any non-empty literal satisfies it. The mutation was
  caught only by `ac4_coverage_test.go:146`, which is AC4's dispatch test and is itself a
  constant-vs-itself comparison (census site 7).
- **Impact:** AC7's stated guarantee — no `Diagnostic` reachable from `Render`/`RenderTo` carries a
  code outside the registry — does not hold over the class. A future story adding a Diagnostic on a
  path the runtime test does not enumerate ships an unregistered code with both AC7 guards green,
  which is the exact fallback-bucket failure D-3.6.7 was written to close.
- **Suggested Resolution:** Add the missing-glyph path to the runtime enumeration, and extend the
  AST scan to resolve each literal `Code` value against `diag.Registered` rather than testing only
  for non-emptiness. Re-run the mutation above as the red-proof.
- **Related AC:** AC7

### Finding 4: AC9's shipped mutation is not AC9's mutation, and the instrument is structurally unfailable
- **Severity:** Major
- **Category:** Tests
- **Location:** `folio-go/render_error_test.go:200-232`
  (`TestMessageRewriteDoesNotAffectCodeRecovery`)
- **Observation:** AC9's mutation reads: *"rewrite each of the five messages in the working tree →
  all code assertions stay green; a deliberately message-matching **control** assertion reddens,
  proving the mutation was live and the green is not vacuous."* What shipped instead assigns to a
  field of an already-recovered local value — `renderErr.Diagnostic.Message = "a completely
  rewritten, unrelated sentence"` — and then asserts `renderErr.Diagnostic.Code` is unchanged. That
  `Code` survives an assignment to `Message` is guaranteed by Go's type system: the assertion cannot
  fail for the reason AC9 names. Its "control" (`Message == originalMessage`) merely asserts that
  the assignment on the line above took effect — also unfailable. It covers one of five modes.
  I then ran the **real** mutation: rewriting the production message text in
  `folio_expr_validate.go:127,147` (both occurrences of *"is not a valid expression"*). Result:
  **872 pass / 1 fail** — every code assertion green, and **no control assertion reddened anywhere
  in the tree**.
- **Impact:** The stability property AC9 claims does in fact hold — I verified it — but nothing in
  the suite witnesses it. The green under a live message rewrite is exactly the vacuity AC9's
  control clause exists to rule out, and the shipped control cannot detect it because it never
  touches the working tree.
- **Suggested Resolution:** Add a control assertion that matches on production message text (for at
  least one mode) so that a real message rewrite reddens something, and state in the test comment
  that the control is the non-vacuity witness. Optionally extend the code-stability assertion to all
  five modes rather than one.
- **Related AC:** AC9

### Finding 5: DW-19's retirement is internally contradictory — `Status: open` under a RETIRED heading
- **Severity:** Minor
- **Category:** Convention
- **Location:** `_bmad-output/implementation-artifacts/deferred-work.md:869`, `:870`, `:891`,
  `:910-913`
- **Observation:** `:869` marks DW-19 **RETIRED**; `:891` still reads verbatim
  `- **Status:** open, with the fix shape SPECIFIED so it cannot be discharged vacuously`. Worse,
  `:870` claims retirement *"**at the fix shape this entry itself specified**"* — which is false:
  the shipped fix departs from that shape (correctly, and ratified at decision-log `:11023`). The
  contradicting fix-shape paragraph survives at `:910-913`, and DW-19 never mentions the amendment.
- **Impact:** A reader re-grounding from DW-19 gets three mutually inconsistent answers about
  whether the item is closed and what shape closed it — and is pointed at a mechanism the lead has
  since ruled would have been wrong.
- **Suggested Resolution:** Set `Status:` to retired, correct `:870` to say the fix was applied with
  a different mechanism, and cite D-3.6.5 (amendment).
- **Related AC:** AC10

### Finding 6: D-3.6.5 carries no forward pointer to its own amendment — the exact failure D-000.69 forbids
- **Severity:** Minor
- **Category:** Convention
- **Location:** `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:10938` (original)
  vs `:11023` (amendment)
- **Observation:** The original D-3.6.5 verdict line still reads *"the fix is a SCAN ERROR, never a
  skip"* with no qualifier. Its amendment 85 lines later says *"The original entry currently forbids
  the thing that was correct."* D-000.69 — added in this same commit — rules that where a verdict
  line and its bindings diverge, the fix is *"to **append a clause saying which half governs** —
  never to leave a reader to choose."*
- **Impact:** A lead re-grounding by reading verdict lines at scale (D-000.8) reads the superseded
  mechanism as current. This is the precise propagation failure D-000.69 was minted to stop,
  occurring in the commit that mints it.
- **Suggested Resolution:** Append a one-line pointer to D-3.6.5 naming the amendment and which half
  governs.
- **Related AC:** —

### Finding 7: `TestResolveAssetsStillReportsATrackedViolation` omits the co-present untracked directory
- **Severity:** Minor
- **Category:** Tests
- **Location:** `lint/internal/manifest/manifest_test.go:167-190`
- **Observation:** AC10's mutation is *"point the resolver at a **tracked** directory carrying a
  real violation → it must still report the violation"*, and its whole purpose is proving the new
  exclusion cannot mask a real finding. The shipped test builds a synthetic repo containing **only**
  the violating tracked directory — no untracked directory is present, and certainly none sorting
  ahead of it. The masking scenario is therefore never exercised. (I verified the property holds by
  constructing the co-present case by hand; the shipped regression guard does not.)
- **Impact:** The permanent guard does not cover the condition it was added for. A future change to
  the exclusion's position in the loop could reintroduce masking with this test still green.
- **Suggested Resolution:** Add an untracked font directory sorting alphabetically ahead of the
  violating tracked one to the same synthetic repo, and assert the tracked violation is still the
  reported error.
- **Related AC:** AC10

### Finding 8: AC3's mutation text still carries the claim D-3.6.9 ruled false and ordered corrected
- **Severity:** Minor
- **Category:** AC Conformance
- **Location:** story `:509-512` (AC3's *Mutation* clause); ruling at decision log `:11067`
- **Observation:** AC3 still reads *"change `diag.CodeEmptyAverage`'s string → AC3's literal
  assertion, AC5's pin **and** the existing 3.3 suites must all redden. If any one of the three
  stays green, that instrument is not seeing the bridge."* D-3.6.9, added in this commit, rules the
  claim *"false and must be corrected to claim only what holds."* The Delivery Log records the
  correction; the AC text does not.
- **Impact:** The AC as written declares the shipped tree non-conformant. A later reader running the
  AC's mutation literally will conclude an instrument is broken.
- **Suggested Resolution:** Amend AC3's mutation clause to name the two instruments that do redden
  and record why the 3.3 suites cannot.
- **Related AC:** AC3

### Finding 9: AC10 and AC12 still specify `lint` 107 pass / 0 fail; actual is 109 / 0
- **Severity:** Minor
- **Category:** AC Conformance
- **Location:** story `:612` (AC10) and `:629` (AC12)
- **Observation:** Both ACs state the target as *"**107 pass · 0 fail**"*. The measured result is
  **109 / 0** — the two new permanent regression tests the fix required. Disclosed in the Delivery
  Log (`:980-981`), not corrected in the AC text.
- **Impact:** A gate stated in absolute numbers that no longer matches the tree invites a future
  reader to treat a correct result as a failure, or to "fix" the count by deleting tests.
- **Suggested Resolution:** Update both AC numbers to 109 / 0, or restate them relationally.
- **Related AC:** AC10, AC12

### Finding 10: `absences.go`'s AC2-mandated comment names a test that does not exist
- **Severity:** Minor
- **Category:** Correctness
- **Location:** `lint/internal/rules/absences.go:105`
- **Observation:** AC2 requires the 3 → 2 → 1 → 0 schedule sentence appear in `absences.go`'s own
  comment. It does, but it cites `TestAbsencesToleratesAnEmptyCheckList` as the proof that the
  mechanism goes slack at zero. No such test exists; the real one is `TestAbsencesZeroWitnessIsCaught`
  (`lint/internal/rules/absences_test.go:59`).
- **Impact:** The forward record for Story 5.1 points at nothing. The whole value of the comment is
  that a future reader can find the evidence.
- **Suggested Resolution:** Correct the name to `TestAbsencesZeroWitnessIsCaught`.
- **Related AC:** AC2

### Finding 11: the corpus assertion guards its duplicated fixture table by COUNT, not by SET
- **Severity:** Minor
- **Category:** Tests
- **Location:** `folio-go/missing_glyph_corpus_test.go:44-87` (the hand-duplicated table) and
  `:88-90` (the guard)
- **Observation:** The test re-declares all five committed fixtures inline and guards the
  duplication with `if len(fixtures) != len(baselineAcceptanceFixtures)`. That is a length check.
  If a later story **replaces** one fixture with another (count unchanged at five), this test
  silently keeps rendering the old five and never sees the new one, while its own doc comment claims
  *"the two must name the same committed corpus"*. (The count itself is separately pinned at
  `first_baseline_acceptance_test.go:258`.) The zero-warning count is correctly **derived**, not a
  literal `len(res.Diagnostics) == 0` — that half of AC's requirement is satisfied, and task 9's
  subject is correctly a synthetic inline template, never a committed fixture.
- **Impact:** The corpus-wide guarantee silently narrows to a stale subset on any fixture swap —
  exactly the drift the "same committed corpus" comment promises to prevent.
- **Suggested Resolution:** Compare the fixture **names** against `baselineAcceptanceFixtures`'
  names as sets, or drive the loop from `baselineAcceptanceFixtures` directly instead of duplicating
  it.
- **Related AC:** AC4 / OPEN-1 obligation (b)

### Finding 12: no assertion covers all FIVE FR41 codes being pairwise distinct
- **Severity:** Minor
- **Category:** Tests
- **Location:** `folio-go/render_error_test.go:127`, `:159-166`
- **Observation:** AC4 requires *"the five codes are pairwise distinct"*. `seenCodes` in
  `TestFourErrorModesCarrySeverityErrorDiagnostics` covers only the **four** Error modes; the fifth
  (missing glyph) is asserted in a different test and never compared against the other four.
- **Impact:** The distinctness property is asserted over 4 of the 5 subjects AC4 names.
- **Suggested Resolution:** Include `DiagCodeTextMissingGlyph` in the distinctness set, or add a
  small five-element table asserting pairwise distinctness directly.
- **Related AC:** AC4

### Finding 13: the missing-glyph Warning is emitted once per RUNE OCCURRENCE
- **Severity:** Minor
- **Category:** Maintainability
- **Location:** `folio-go/render.go:1044-1051` (inside `for _, r := range elementText`)
- **Observation:** A Diagnostic is appended for every uncovered rune, before adjacent uncovered
  runes are merged into a single segment. An element whose declared chain covers none of its script
  yields one Warning per character — a 200-character Thai string in a Latin-only chain produces 200
  near-identical Warnings on `Result.Diagnostics`.
- **Impact:** D-2.8.3's channel is the same one Story 5.12 must present to a human, and DW-17's
  amendment records that for this mode the diagnostic is the *only* record. A flood is a
  presentation problem for 3.7 and 5.12 that is cheaper to shape here.
- **Suggested Resolution:** Consider one Warning per (element, distinct rune) or per element naming
  the distinct uncovered runes. If per-occurrence is deliberate, record it as a bound so 3.7/5.12
  inherit it knowingly.
- **Related AC:** AC4, OPEN-1 / D-3.6.8

### Finding 14: lint's half of AC1's double anchor has no vacuity witness for `diag`
- **Severity:** Minor
- **Category:** Tests
- **Location:** `lint/internal/rules/stagerank_test.go:41`
- **Observation:** The vacuity guard pins the scanned set to
  `[]string{"layout", "pagemodel", "pdf", "text", "fontset"}` — `diag` is absent. `ScanStageRank`
  does walk every directory, so a violation there is reported (I proved it reddens), but nothing
  asserts the scanner actually **entered** `internal/diag`. AC1 requires the property rest on two
  independent guards; the lint one is unwitnessed for this package specifically.
- **Impact:** If a future change caused the scan to skip `internal/diag`, the lint half would go
  slack silently, leaving the property on one guard.
- **Suggested Resolution:** Add `"diag"` to that slice.
- **Related AC:** AC1

### Finding 15: no `D-000.6 amendment` decision-log entry for the spine edit, breaking a four-instance convention
- **Severity:** Minor
- **Category:** Convention
- **Location:** `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` (absent); prior
  instances at `:1542`, `:2144`, `:2986`, `:10704`
- **Observation:** Every previous canonical-document amendment under D-000.6 got a standalone entry
  quoting verbatim before/after — Story 1.2, 1.3, 1.4 (*"verbatim before/after"*) and 3.4's finisher.
  Story 3.6 has none. The before/after appears only inside D-3.6.4's Divergence-3 paragraph, and the
  text **as actually shipped** (which adds a provenance parenthetical) is quoted nowhere.
- **Impact:** The spine's amendment history becomes unreconstructable from the log alone, which is
  what D-000.6 exists to preserve.
- **Suggested Resolution:** Add a `### D-000.6 amendment (Story 3.6)` entry quoting the shipped
  before/after.
- **Related AC:** AC11

### Finding 16: DW-19 says both red-proofs were "silently masked"; only one was
- **Severity:** Minor
- **Category:** Correctness
- **Location:** `_bmad-output/implementation-artifacts/deferred-work.md:878-880`
- **Observation:** The retirement says `TestFontsAssetsNoticeRemovalRedProof`/`LicenceRemovalRedProof`
  *"the latter two were SILENTLY MASKED before this fix"*. Measured: only
  `LicenceRemovalRedProof` was vacuously green. `NoticeRemovalRedProof` was one of the three
  baseline **reds** — the story's own baseline table says so. D-000.70 (`:11056`) states it
  correctly in the singular.
- **Impact:** Overstates the defect and misdescribes the baseline the retirement is measured against.
- **Suggested Resolution:** Correct to the singular, matching D-000.70.
- **Related AC:** AC10

### Finding 17: stale cross-reference to a test file that does not exist
- **Severity:** Nit
- **Category:** Maintainability
- **Location:** `folio-go/ac4_coverage_test.go:98`
- **Observation:** The doc comment cites the corpus assertion as living in
  `zero-glyph_warning_test.go`. The file is `missing_glyph_corpus_test.go`. The two tests
  deliberately cross-reference each other so the apparent conflict between them is explained; one
  direction points at nothing.
- **Impact:** Cosmetic, but it breaks the explanatory pairing on purpose-built comments.
- **Suggested Resolution:** Correct the filename.
- **Related AC:** AC4

### Finding 18: `TestCodeIsADefinedType` is a tautology, and its own comment says so
- **Severity:** Nit
- **Category:** Tests
- **Location:** `folio-go/internal/diag/diag_test.go:74-80`
- **Observation:** The body is `var c Code = "PROBE"; if string(c) != "PROBE" { … }` — a constant
  compared against itself. It cannot fail while it compiles, which the comment concedes
  (*"it exists to keep that property visible to a reader … not because the compiler needs help"*).
- **Impact:** Counts toward the suite's pass total without discriminating anything. Harmless in
  isolation; noted because AC1 lists the compiler as an anchor and this is the only test that
  purports to exercise it.
- **Suggested Resolution:** Either delete it (the compiler already enforces the property at every
  use site) or convert it to a compile-time assertion with an explanatory comment.
- **Related AC:** AC1

---

## Could not verify

- **Epic 3's boundary matrix run** is correctly out of scope here (due after 3.7, per D-000.4), so
  the three cross-target legs Story 3.5's `hidden-image` document owes were not exercised. Recorded,
  not assessed.
- **AC9's "no in-repo assertion can constrain callers outside the module"** is a bound, not a
  property; it is correctly labelled as such in the story and cannot be verified by construction.
- **The engineering lead's ruling on the DW-19 deviation** is explicitly out of this review's remit.
  The facts are reported above; the deviation is sound on all three claims tested.

---

## Finding Resolutions (finisher pass)

All 18 findings triaged and resolved below. 17 FIXED, 1 DEFERRED (with an explicit owner and a
recorded bound). None dismissed — every finding named a real gap, either in behaviour or in the
record.

### Finding 1 (Blocker, AC10) — FIX
The D-3.6.5 amendment's "Required" floor is implemented: `ResolveAssets`
(`lint/internal/manifest/manifest.go`) now tracks whether ANY directory in `dirOrder` had a non-zero
tracked-file count (`sawTrackedDirectory`); if `dirOrder` is non-empty and no directory was tracked,
it returns a scan error naming every candidate directory, before the wordlist asset is resolved. The
precision the amendment demands is preserved: an empty `dirOrder` (no font-extensioned file anywhere
on disk) stays legitimately empty, never an error. Red-proved against a synthetic repo using the same
`initGitRepo`/`gitAdd` harness the two existing DW-19 tests use
(`TestResolveAssetsAllDirectoriesUntrackedIsAScanError`); the precision boundary itself is pinned by a
second new test (`TestResolveAssetsNoCandidateDirectoriesIsNotAScanError`). This required updating
`TestResolveAssetsExcludesUntrackedDirectoryWithoutError`, whose original synthetic repo held only one,
untracked, candidate directory — under the new floor that shape IS the all-untracked scan-error case,
so a second, tracked, fully-compliant font directory was added alongside it (matching the real
repository's own shape: `.font-sources` untracked, `folio-go/fonts/*` tracked).
**Files:** `lint/internal/manifest/manifest.go`, `lint/internal/manifest/manifest_test.go`.

### Finding 2 (Major, AC5) — FIX
`internal/diag/diag_test.go`'s pin table (`wantCodes`, keyed by identifier STRING) is replaced by
`codePins`, a slice binding each named CONSTANT (read live) to its pinned literal, compared
elementwise (`p.constant != p.literal`). This makes the identifier→string binding itself the subject
of the assertion, rather than the pinned values' set membership. Re-ran all three original mutations
((a) change a string, (b) add without a pin, (c) add with a pin — (c) confirmed still GREEN) plus the
reviewer's swap mutation as a fourth, recorded case: swapping `CodeTemplateMalformed`'s and
`CodeBindingPathAbsent`'s strings now reddens (`CodeTemplateMalformed = "BINDING_PATH_ABSENT", want
pinned literal "TEMPLATE_MALFORMED"` and the mirror image). The table remains hand-typed in the test
file, never derived from the registry source. **Files:** `folio-go/internal/diag/diag_test.go`.

### Finding 3 (Major, AC7) — FIX
Both instruments extended. `TestAllProducedDiagnosticsCarryARegisteredCode` now also renders a
synthetic missing-glyph template (`missingGlyphTemplateJSON`, new) through the public `Render` path
and checks its Warning's Code — the path the test's own doc comment already claimed to cover but
never exercised. `TestNoDiagnosticCompositeLiteralOmitsCode`'s AST scan now resolves any LITERAL
string `Code:` value against `diag.Registered`, not merely non-emptiness (a symbolic `DiagCode*`
reference is registered by construction via the AC3 bridge and is not, and cannot be, resolved
statically). Red-proved by setting the missing-glyph construction site's `Code` to
`"MADE_UP_UNREGISTERED"`: both `TestNoDiagnosticCompositeLiteralOmitsCode` and
`TestAllProducedDiagnosticsCarryARegisteredCode` now redden (previously both stayed green). Restored
and re-verified green. **Files:** `folio-go/diag_no_empty_code_test.go`.

### Finding 4 (Major, AC9) — FIX
`TestMessageRewriteDoesNotAffectCodeRecovery` rebuilt as a table over all five FR41 modes (the
optional extension in the finding's suggested resolution): each case reads Code and Message
independently from a real construction site, asserts Code against its expected registry value, and
asserts Message contains a REAL, production-owned substring (e.g. `"is not a valid expression"` for
invalid-expression, drawn from the actual running text, not invented) as the non-vacuity control.
Red-proved by rewriting the real production wording at `folio_expr_validate.go:127,147` (`"is not a
valid expression"` → `"is not a legal expression"`, the exact site and edit the reviewer used): the
"invalid expression" subtest's control reddened while every other subtest's Code assertion stayed
green. Restored and re-verified green (6/6 pass). Finding 12's pairwise-distinctness fix (below) was
implemented alongside this one since both concerned the same five-mode enumeration.
**Files:** `folio-go/render_error_test.go`.

### Finding 5 (Minor, AC10) — FIX
`deferred-work.md`'s DW-19 entry corrected: `Retired at:` now states the ratified deviation (not "the
fix shape this entry itself specified," which was false — the shipped mechanism excludes rather than
hard-errors); `Status:` changed from `open` to `retired`, both citing the D-3.6.5 amendment.
**Files:** `_bmad-output/implementation-artifacts/deferred-work.md`.

### Finding 6 (Minor) — FIX
The original D-3.6.5 decision-log entry gets a forward-pointer paragraph, per D-000.69's own rule
("append a clause saying which half governs — never leave a reader to choose"): the verdict's
mechanism is superseded by the amendment; its outcome/ownership clauses still stand.
**Files:** `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`.

### Finding 7 (Minor, AC10) — FIX
`TestResolveAssetsStillReportsATrackedViolation` rebuilt to co-locate an untracked directory
(`aaa-scratch-fonts`) that sorts alphabetically ahead of the tracked, violating directory
(`committed-fonts`) — exercising the exact masking scenario AC10's mutation exists to rule out, which
the original version (violating directory alone) never did. **Files:**
`lint/internal/manifest/manifest_test.go`.

### Finding 8 (Minor, AC3) — FIX
AC3's mutation clause rewritten to state only what D-3.6.9 confirms holds: changing
`diag.CodeEmptyAverage`'s string reddens AC3's own literal assertion and AC5's pin; the pre-existing
3.3 suites do not, and are not expected to (D-3.6.9's wiring-vs-string-value layering, quoted in
place). **Files:** `_bmad-output/implementation-artifacts/3-6-fail-with-located-actionable-diagnostics.md`
(this file, AC3).

### Finding 9 (Minor, AC10/AC12) — FIX
AC10 and AC12 no longer pin `lint`'s absolute pass count (which had already gone stale twice within
this story: 104/3 baseline → 109/0 at review → 111/0 after this finisher's own Finding 1/7 tests).
Restated relationally per D-000.68's discriminator: the FAIL count (0) is the permanent property; the
pass count is reported as measured, in the Delivery Log, since it is scheduled to keep growing.
**Files:** this story file (AC10, AC12).

### Finding 10 (Minor, AC2) — FIX
`absences.go`'s forward-schedule comment corrected to cite the test that actually exists,
`TestAbsencesZeroWitnessIsCaught`, replacing the non-existent `TestAbsencesToleratesAnEmptyCheckList`.
**Files:** `lint/internal/rules/absences.go`.

### Finding 11 (Minor, AC4/OPEN-1) — FIX
`missing_glyph_corpus_test.go`'s presence precondition now compares the test's own fixture table
against `baselineAcceptanceFixtures` by NAME, as two sets (both directions), in addition to the
existing length check — a fixture swap (count unchanged, membership changed) now reddens, which the
length check alone could not see. **Files:** `folio-go/missing_glyph_corpus_test.go`.

### Finding 12 (Minor, AC4) — FIX
`TestFourErrorModesCarrySeverityErrorDiagnostics` extended: after its four-mode loop, it separately
renders the missing-glyph synthetic template and checks that Warning's code against the SAME
`seenCodes` set the four Error modes built, so a registry-code collision between the Warning and any
Error mode is now caught, not just among the four Errors. **Files:** `folio-go/render_error_test.go`.

### Finding 13 (Minor, AC4/D-3.6.8) — DEFER
Per-rune-occurrence Warning emission is a presentation-shaping question, not a `folio-go` correctness
gap: no AC or ruling specifies a batching granularity, and DW-17 (the presented-interface obligation)
already owns exactly this territory, assigned to Stories 3.7/5.12/6.6. Inventing a coalescing shape
in this finisher pass risks choosing a granularity those stories' actual presentation design would not
have picked, which D-2.6.5's guardrail (don't build presentation machinery a consuming story doesn't
yet name) counsels against. Recorded as an explicit, knowing bound in DW-17's own entry so none of its
three owners inherits per-occurrence emission silently by copying the raw diagnostic list.
**Owner:** whichever of Stories 3.7 / 5.12 / 6.6 first presents this Warning to a human.
**Files:** `_bmad-output/implementation-artifacts/deferred-work.md` (DW-17, amended in place).

### Finding 14 (Minor, AC1) — FIX
`internal/rules/stagerank_test.go`'s vacuity-guard package list gains `"diag"`, so the lint half of
AC1's double anchor now asserts the scanner actually entered `internal/diag`, not merely that a
violation there would be reported. **Files:** `lint/internal/rules/stagerank_test.go`.

### Finding 15 (Minor, AC11) — FIX
A standalone `D-000.6 amendment (Story 3.6, finisher pass)` decision-log entry added, quoting the
spine's `diag/` line verbatim before and after — matching the convention every prior canonical-document
amendment under D-000.6 followed (Stories 1.2, 1.3, 1.4, 3.4's finisher). **Files:**
`_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`.

### Finding 16 (Minor, AC10) — FIX
`deferred-work.md`'s DW-19 entry corrected from "the latter two were SILENTLY MASKED" to name only
`TestFontsAssetsLicenceRemovalRedProof` — `TestFontsAssetsNoticeRemovalRedProof` was already failing
loudly at baseline for its own reason, matching D-000.70's own singular phrasing. **Files:**
`_bmad-output/implementation-artifacts/deferred-work.md`.

### Finding 17 (Nit, AC4) — FIX
`ac4_coverage_test.go`'s stale cross-reference to `zero-glyph_warning_test.go` corrected to the file's
actual name, `missing_glyph_corpus_test.go`. **Files:** `folio-go/ac4_coverage_test.go`.

### Finding 18 (Nit, AC1) — FIX
`TestCodeIsADefinedType` (a runtime test whose own comment conceded it could not fail while the
package compiled) replaced with a package-level compile-time assertion,
`var _ Code = Code("compile-time anchor: Code must stay string-based (R4)")`, which fails to COMPILE
rather than passing-but-toothless if `Code` ever stops being string-based. **Files:**
`folio-go/internal/diag/diag_test.go`.
