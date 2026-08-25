# Story 3.1a: Guard AD-23 against named binary-float types

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-1a-guard-ad-23-against-named-binary-float-types`
**Status:** `done`
**Covers:** no FR/NFR. **This story does not come from `epics.md`.** It was ruled into existence by
**D-3.1a.1** and sequenced between 3.1 and 3.2, on **D-000.25**'s precedent — which inserted 2.3a for
this same guard, on three reasons that recur verbatim. `epics.md` is **not** amended by this story and
must not be: there is no clause in it this story implements, weakens or corrects.
**Primary invariant:** **AD-23** (exact scaled-integer decimals; no binary floating point under
`folio-go/`).
**Adjacent invariants:** AD-1, AD-5 (stage rank), AD-14 (located diagnostics), AD-25 (offline oracles),
AD-26.
**Governing rulings:** **D-3.1a.1 (the charter)** · **D-000.50** · **D-000.59** · **D-000.22** ·
**D-000.53** · **D-000.23** · **D-000.13** · **D-000.38** · D-000.9 · D-000.14 · D-000.15 · D-000.21 ·
D-000.26 · D-000.30 · D-000.32 · D-000.42 · D-000.60 · D-1.3.3 (amended) · D-1.3.6 · D-1.3.11 ·
D-1.6.1 · D-1.6.6 · D-2.3.1 · D-3.2.1 (forward-looking; this story must not pre-empt it)
**Retires no DW item. Adds none.**

**Baseline measured in this run, at creation.** HEAD is **`b227dda`** — *"Record D-000.60: a
restriction taken out against an open question must be lifted when the question is answered"* — on
branch `main`, **working tree clean** (`git status --porcelain` empty, verified before every
measurement below and again afterwards).

Test state at baseline, stated with its scope and flags (D-000.26):

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting **every** `--- PASS` / `--- FAIL` / `--- SKIP` occurrence including subtests | **619 PASS · 1 FAIL · 1 SKIP** |
| `folio-go/` | the same invocation, counting only **top-level** results | **380 PASS · 1 FAIL · 1 SKIP** |
| `lint/` (**clean worktree**) | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, every occurrence | **85 PASS · 0 FAIL** |
| `lint/` (**clean worktree**) | the same invocation, top-level only | **47 PASS · 0 FAIL** |

The single `folio-go` failure is `internal/text`'s `TestCorpusMeetsP6ExerciseFloors` (P6g opaque-names
floor: got 7, need ≥ 20), with stats **`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`**. It is
**Story 2.1's deliberate, pre-stated shortfall (D-000.17 / D-2.1.14)** and Epic 2 was closed over it.
**It is REQUIRED to stay red, with those stats byte-identical.** Do not fix it, do not tune the corpus,
and do not count it when reporting this story's results. The one SKIP is
`TestXrefEntriesRejectsMalformedSubprocess`.

> **BASELINE CORRECTION THE DEVELOPER MUST READ BEFORE MEASURING `lint`.** Run in **this working
> checkout**, `lint` reports **82 PASS · 3 FAIL**, not 85 · 0. The three failures are
> `TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist` and
> `TestFontsAssetsNoticeRemovalRedProof`, and all three carry the same message:
> `ResolveAssets: .font-sources: contains a committed font binary but no LICENSE* file`.
>
> **`.font-sources/` is not in the repository.** It is untracked and gitignored (`.gitignore:85`,
> `/.font-sources/`) — a local scratch directory holding the three upstream **variable** faces from
> which the shipped static faces were generated. `git ls-files .font-sources` returns **zero** rows.
> Re-measured at `b227dda` in a detached worktree at `/tmp/folio-clean`, where the directory does not
> exist, `lint` is **85 PASS · 0 FAIL**. That is the true baseline and the figure in the table above.
>
> **Do not "fix" this**, and in particular do not add a LICENSE file to a gitignored directory, do not
> weaken `ResolveAssets`, and do not delete `.font-sources/` — it is the owner's, not ours. Measure
> `lint` in a clean worktree, or subtract these three known-local failures by name. That
> `ResolveAssets` walks a gitignored dot-directory at all is arguably a real latent defect; it is
> **out of scope here** and is noted for the record, not for fixing (see Flag F4).

Everything asserted below was measured at `b227dda`. The oracle arithmetic was computed with the
system `python3` (**3.12.13**) and independently cross-checked with `/usr/bin/bc`; both invocations
are reproduced verbatim in the findings.

---

## In plain terms (read this first if you just want the gist)

The engine has a standing rule: never use the kind of arithmetic that computers do quickly but
approximately. It is why a total can come out a fraction of a penny wrong, and why the same document
can add up differently on two machines. Money is about to arrive, so this rule now has real teeth: an
exact totalling and averaging routine, checked against answers worked out independently and confirmed
by two unrelated tools, plus a narrow ban on the two specific approximate number types that were
sneaking past the engine's older checks.

The instructive part is what review found in the ban itself. The whole reason this story exists is
that an approximate number type was hiding under an unfamiliar name. The first version of the new ban
still had exactly that blind spot, one layer down: a type given an alternate name for the same
underlying number sailed straight past it, and the check reported the tree clean while its own
instrumentation showed it had genuinely looked and found nothing. That is the worst shape a false
negative can take: not "we didn't check," but "we checked, and the check itself was wrong." Review
caught it, and the fix closes the gap directly. Review also found the ban was silently skipping test
files, an overflow bug that could let an extreme total wrap around instead of being rejected, a couple
of test assertions that could not actually fail no matter what they were checking, and a tripwire meant
to stop a future story from quietly reinventing this arithmetic that was narrower than its own
description claimed. All of it is now fixed, with each defect reproduced live before its fix and
re-verified after.

Two things still look intentionally odd and are not: one long-standing test stays failing on purpose,
inherited from earlier work, and some sample figures in the test data are there specifically because
they cannot catch the fault they exist to demonstrate — evidence, not oversight.

---

## Story

**As a** maintainer of an engine whose determinism and whose money path both rest on exact integer
arithmetic,
**I want** AD-23's ban enforced by a behavioural exactness oracle rather than by the spelling of two
type names, and the two named binary-float types that slip past both existing guards banned by
resolved type identity,
**So that** the story that builds aggregation is written against a guard that already exists, and the
developer writing `sum` is not also the developer deciding how strict the guard on `sum` is.

---

## Do not re-open — settled rulings this story inherits

Reproduced with their rationale so the developer does not re-litigate them.

1. [x] **NO TYPE-SHAPE CHECK CAN CATCH `math/big.Float`.** This is D-3.1a.1's recorded correction of
   the engineering lead's own grounding report, preserved **because the wrong version is the
   intuitive one**. **Verified independently in this run** against Go 1.26.0's
   `$(go env GOROOT)/src/math/big/float.go`:

   ```go
   type Float struct {
       prec uint32; mode RoundingMode; acc Accuracy
       form form; neg bool; mant nat; exp int32
   }
   ```

   `mant` is `nat`, and `nat.go:35` declares `type nat []Word` with `arith.go:19` declaring
   `type Word uint`. **There is no float field anywhere in the structure.** `big.Float` is binary
   floating point *semantically*, implemented over integers. A recursive float-containment check
   finds nothing and reports green with the hole wide open. **Do not build one.** If you find
   yourself writing a structural walk over type fields, stop and re-read this item.

2. [x] **The guard lives in `lint`, and NOT ALSO in `folio-go/internal/arch_test.go`.**
   `arch_test.go` is a pure AST walk with no type information (`findFloatOccurrences:49` matches
   `*ast.Ident` by `.Name`), so extending it could only ever be a **name** match on `big.Float` —
   the exact proxy mechanism this story exists to stop. Putting it in both places is D-000.38's
   *"two guards sharing a parser are one guard wearing two names"*.

3. [x] **Layer 3 is NOTHING.** Both existing guards stay exactly as they are, on their own axes.
   Do not edit `folio-go/internal/arch_test.go`. Do not edit
   `lint/internal/rules/floattyped.go`. This is a **scope prohibition, not an acceptance test** —
   and deliberately so: an AC asserting those files are byte-identical would pass vacuously
   whenever nobody edited them, which is every run in which the prohibition was already being
   obeyed. It has no teeth and it is not pretending to.

4. [x] **Layer 2 is a DENYLIST and must be labelled as one (D-000.23).** It covers `math/big.Float`
   and `math/big.Rat` — **those two types, not the class.** Its coverage witness must say so **in
   those words**. It is **not** counted as coverage for AD-23's property; **Layer 1 is.** A denylist
   entry is never coverage.

5. [x] **`big.Rat` is banned for a DIFFERENT reason from `big.Float`, and that is why Layer 1 alone
   is insufficient.** `big.Rat` is *exact*, so the behavioural oracle would **not** catch it. It is
   wrong because it carries an unrounded rational and thereby **dodges** AD-23's *"divides at a
   defined scale with round-half-to-even"* — and the scale is the **ruled** part, not an
   implementation detail. Do not "simplify" Layer 2 down to `big.Float` on the grounds that
   `big.Rat` is exact.

6. [x] **Accumulator overflow is a located Error, never a widening.** Accumulate in `big.Int`
   (already blessed, already imported at `decimal.go:20`, exact) and narrow **once**, at the end,
   through the existing `IsInt64()` pattern at `decimal.go:119-128`. **Do not write a checked-add
   helper** — there is one narrowing site and it already exists.

7. [x] **Do not create `folio-go/internal/expr/`.** `absence-expr-package` (`absences.go:96`)
   requires it absent until Story 3.2, and D-3.2.1 assigns that package's creation — and DW-8's
   `Decimal` move — to 3.2. Creating it here would fire a tripwire this story has no mandate to
   discharge. The kernel this story builds goes in `internal/bind`, beside `Decimal`, and **travels
   with `Decimal`** when 3.2 moves it.

8. [x] **3.1a is NOT a D-000.4 matrix override.** The criterion is a *new source of cross-target
   divergence*; a lint rule and an integer-only kernel introduce none. Cadence is per-epic:
   unit / vet / build / lint every story; the cross-target matrix is due at **Epic 3's close**.

---

## Findings — measured in this run, at `b227dda`

### F1. The hole is real and completely unoccupied

`big.Float` and `big.Rat` occur at **zero sites repo-wide**:

```
$ grep -rn 'big\.Float\|big\.Rat' --include='*.go' .    →  no matches (exit 1)
```

No use, no guard, **no mention**. `internal/bind` already imports `math/big` (`decimal.go:20`, for
`big.Int`), so the import is blessed and `big.Float` is one identifier away. Neither existing guard
can see it: `findFloatOccurrences` matches only the identifiers `float32`/`float64` and
`token.FLOAT` literals; `ScanFloatTypedValues` tests `tv.Type.Underlying().(*types.Basic)` for
`types.IsFloat`, and `big.Float`'s underlying type is a `*types.Struct`.

### F2. `bind.Decimal` has no arithmetic at all — verified

`folio-go/internal/bind/decimal.go` declares **exactly one function**:

```
$ grep -n 'func ' folio-go/internal/bind/decimal.go
folio-go/internal/bind/decimal.go:76:func NewDecimal(literal string) (Decimal, error) {
```

The only other route to a `Decimal` is `value.go:63`'s `func (v Value) AsDecimal() (Decimal, error)`,
which delegates to `NewDecimal`. **There is no Add, no Compare, no Div, and no accumulation
anywhere.** This is what forces the scope question resolved below.

### F3. D-000.50's population check — three candidate corpora, and only ONE can express the defect

**This is the story's real work, and it was done before any assertion was written.** The mutant used
throughout is the *honest* one: accumulate in `float64`, then round the result to the operation's
declared scale — because a mutant that skips the final rounding manufactures a divergence that the
real implementation would have rounded away.

| corpus | shape | `sum` reddens? | `avg` reddens? |
|---|---|---|---|
| **A** — `12345678901234.56` + 32 × `0.01` (n=33) | large balance, 16 significant digits, many small addends | **YES** | **YES** |
| **B** — `100.00`, `0.0001`, `3`, `0.000007`, `2.5` (n=5) | mixed scale; exercises alignment | no | no |
| **C** — `0.10 0.20 0.30 1.15 2.35 0.05 0.07` (n=7) | the intuitive small money corpus | no | no |

Measured values:

```
A: sum exact = 12345678901234.88     float64 mutant = 12345678901234.87    → REDDENS
   avg exact = 374111481855.602424   float64 mutant = 374111481855.602230  → REDDENS
B: sum exact = 105.500107            float64 mutant = 105.500107           → identical
   avg exact = 21.1000214000         float64 mutant = 21.1000214000        → identical
C: sum exact = 4.22                  float64 mutant = 4.22                 → identical
   avg exact = 0.602857              float64 mutant = 0.602857             → identical
```

**Corpus C is the finding that matters.** It is exactly what a reasonable person writes when asked for
a bank-statement fixture — seven small two-decimal amounts — and under an honest float64 accumulator it
produces a **byte-identical** total and a **byte-identical** average. An oracle built on it would have
been correct, would have read as sound in review, and would have caught nothing. That is D-000.50's
hazard, reproduced on this story's own subject.

**Corpus B is a second, subtler trap.** Before quantisation, B's `avg` *appears* to discriminate:
Python reports `21.1000214000` against float64's `21.1000214`, and it is tempting to record that as
teeth on the grounds that D-1.6.1's AC2 makes trailing zeros meaningful. **It is not teeth.** Rounding
the mutant's result to the declared scale restores the trailing zeros and the two agree exactly. The
apparent divergence was an artifact of measuring the mutant without its final rounding step. B stays in
the corpus because it exercises the **alignment** path — which A, whose operands all sit at scale 2,
does not touch at all — and it is labelled as non-discriminating.

**The mechanism that makes A work is magnitude absorption, not accumulated drift.** Summing plausible
two-decimal amounts drifts far too slowly to matter: corpus B-style accumulation of 100 random amounts
totalling ~467,610.75 was measured off by 2 × 10⁻¹⁰, which no money scale can see. A needs ~10⁷
operands to drift one cent. What actually reddens is a balance near float64's integer-exactness limit
(2⁵³ ≈ 9.007 × 10¹⁵) absorbing addends smaller than its ulp.

### F4. Exact addition is order-invariant; float64 addition is not — and this is the semantic acceptance step

```
$ echo "scale=2; 12345678901234.56+0.01+0.01+…(32 terms)" | bc     →  12345678901234.88
$ echo "scale=2; 0.01+0.01+…(32 terms)+12345678901234.56" | bc     →  12345678901234.88
   float64 forward = 12345678901234.873      reversed = 12345678901234.88      identical: False
```

Two things fall out, and both are load-bearing:

- **D-000.53's independent-reader step has an instrument.** `/usr/bin/bc` did not produce the golden
  (Python's `decimal` did) and this project did not write it. It resolves the recorded corpus into the
  total the golden claims. **Python cannot serve as its own acceptance reader** — that is the whole
  point of the independence.
- **D-000.22's semantic acceptance property is order-invariance.** A human can be wrong about it, a
  machine can check it, and it is read **off the artifact** rather than off the inputs: *the recorded
  total is unchanged when the corpus is summed in reverse order.* A float64-contaminated recording
  fails it. This also closes a real hole in the red-proof — note above that the **reversed** float64
  sum happens to land on the exact answer. A red-proof that mutated the accumulator and summed in one
  order only could, on a different corpus, pass **by luck**. Asserting order-invariance catches the
  mutant either way, because the mutant gives two different answers.

### F5. Layer 2's scope must be the folio-go **module root**, not `folio-go/internal/`

D-3.1a.1 phrases Layer 2 as forbidding the two types *"under `folio-go/internal/`"*. Taken literally
that is **strictly narrower than the guard it sits beside**: `ScanFloatTypedValues`' shipped production
caller (`floattyped_test.go:34`) points at `filepath.Join(root, "folio-go")` — the **module root** —
and the module root holds 60 `.go` files including `render.go`, `render_entry.go` and `page_number.go`,
i.e. **the public entry point**. A `big.Float` in `render.go` would sit outside a
`folio-go/internal/`-scoped rule.

**Resolved: scope Layer 2 to the folio-go module root**, which strictly contains `folio-go/internal/`.
Widening a ban is safe; narrowing one silently is not, and matching the neighbouring guard's scope
exactly is what keeps the two axes comparable. This is recorded as **Flag F1** below rather than
applied silently.

---

## The Layer-1 scope question, and how it is resolved

**The question.** D-3.1a.1 makes Layer 1 an exactness oracle over `sum`/`avg`. **`sum` and `avg` do
not exist** — 3.2 creates `internal/expr` and the function table, 3.3 builds all of `Decimal`'s
arithmetic from nothing, and F2 above verifies that `bind.Decimal` today has `NewDecimal` and nothing
else. A golden with nothing to compare against is vacuous, and under D-000.22 the first recording is
the **only** moment at which "is this right?" is answerable at all — recording it unconsumed means it
gets **ratified**, not checked.

**The resolution: 3.1a builds the exact arithmetic KERNEL, and Layer 1 is wired here, not deferred.**

The charter contains two sentences that must both be satisfiable in this story:

- Layer 1's subject is `sum`/`avg`, chosen so a float64 implementation demonstrably diverges; and
- *"accumulate in `big.Int` … and narrow **once**, at the end, through the existing `IsInt64()`
  pattern at `decimal.go:119-128`"*, with *"the alignment step … in scope"*.

The second sentence names line numbers in code that exists **today**. There is no reading under which
it is an obligation on 3.3. And the only way both sentences hold at once is that this story builds the
arithmetic those line numbers describe. So it does.

**What 3.1a builds:** an exact reduction kernel over `[]Decimal` in `internal/bind` — alignment to a
common exponent, accumulation in `big.Int`, one narrowing, one declared division scale.

**What 3.1a does NOT build, and 3.3 owns:** `sum`/`avg` as *expression-language functions* — the closed
eight-entry function table, arity and wrong-kind checking, collection resolution, row scope,
empty-collection semantics, and AD-14 diagnostic codes (which cannot exist before 3.6 mints them, per
`absence-diag-package`).

**Why this is not the erosion D-3.1a.1 guards against.** The charter's fourth reason for inserting this
story is that landing the guard in 3.3 *"makes the developer writing `sum` the same developer deciding
how strict the guard on `sum` is, under delivery pressure."* Fixing the exactness kernel and its oracle
**before** that developer arrives serves that reason rather than undercutting it: the strictness is
settled while nothing is being delivered under pressure.

**What 3.3 must discharge, and the tripwire that makes it unable to skip the wiring.** Per D-000.59's
pattern — an obligation asserted positively, never the event — 3.1a lands an **AST-counted set-equality
inventory** (D-000.14: when a count is load-bearing, count by AST, never by text):

> **the set of functions in the `folio-go` module that reduce a sequence of `Decimal` to a
> `(Decimal, error)` is exactly `{SumDecimals, AvgDecimals}`, and they are declared in one package.**

This is the same instrument D-3.2.1 ratified for `Decimal` itself — set-equality-plus-location, not
extinction, because the functions must exist somewhere. It **cannot be satisfied by adding a second,
differently-named top-level reducer function of this shape**, and it **cannot be satisfied by
duplicating**. If 3.3 needs a third reducer, it amends the inventory in the same diff and a human
reads it — the same deliberate friction `assertExactFindingSites` already documents for the float-site
register. Its red-proof is available **now**, before 3.3 exists (AC23), which matters because D-000.30
shuts that window the moment the obligation is wired.

**CORRECTION (QA review Finding 11, Minor; D-3.1a.4).** This paragraph and D-3.1a.3 originally claimed
the tripwire "cannot be satisfied by 3.3 writing a second accumulator inside `sum`". That overstated
what the instrument checks: nothing here (or anywhere) asserts that any code *calls*
`SumDecimals`/`AvgDecimals`, so an inline `big.Int` loop written directly inside a future
expression-language `sum()` is not a top-level `func([]Decimal) (Decimal, error)` declaration and would
never appear in the inventory at all — the tripwire would still report exactly `{SumDecimals,
AvgDecimals}` and pass. What it genuinely prevents is a second, differently-named top-level reducer
function of this exact shape — the wording above now says that and only that. Routing enforcement (did
3.3's `sum`/`avg` actually call the kernel) is an obligation Story 3.3 owns, not this story; see
D-3.1a.4 for the full correction and the follow-up it hands to 3.3.

---

## Acceptance Criteria

### Layer 1 — behavioural, PRIMARY, has teeth

- [x] **AC1.** An exact decimal reduction kernel exists in `folio-go/internal/bind`, exporting
  `SumDecimals([]Decimal) (Decimal, error)` and `AvgDecimals([]Decimal) (Decimal, error)`. These are
  arithmetic kernels, **not** expression-language functions: no arity checking, no collection
  resolution, no diagnostic codes, no function-table entry.
- [x] **AC2.** `SumDecimals` aligns operands to their common (minimum) exponent, accumulates in
  `big.Int`, and narrows **exactly once**, at the end, through the existing `IsInt64()` pattern of
  `decimal.go:119-128`. **No checked-add helper is written** (D-3.1a.1, verbatim). The single narrowing
  site is identified in a comment naming the pattern it reuses.
- [x] **AC3.** `SumDecimals`' result carries the **maximum operand scale** (the minimum operand
  exponent), with trailing zeros preserved — D-1.6.1's AC2 distinctness (`"1.50"` ≠ `"1.5"`) survives
  addition. Asserted on a corpus member whose operands differ in scale.
- [x] **AC4.** `AvgDecimals` divides at **maximum operand scale + 4**, **round-half-to-even**, declared
  **once** as a named constant that is a property of the operation and is **never derived from the
  data** and **never fitted to the shipped corpus** (D-000.32). A tie case is asserted, so the
  half-to-even half is exercised rather than assumed.
- [x] **AC5.** Coefficient overflow at the single narrowing site is a **located Error** naming the
  operation, the operand count and the bound breached — **never a widening, never a wrap, never a
  truncation**. Red-proved with a corpus that overflows `int64` only after accumulation, so the error
  is reached through the accumulator rather than through `NewDecimal`'s existing per-literal bound.
- [x] **AC6.** The **alignment step is bounded**: the spread between the maximum and minimum operand
  exponent is checked against `maxDecimalExponentMagnitude` (`decimal.go:65`) and a breach is an Error.
  **The check happens BEFORE any shift is computed** — computing a 200,000-digit shift in order to
  discover it was too large is itself the cost being defended against. Asserted by a test that would
  time out or allocate unboundedly if the check ran after the shift.
- [x] **AC7.** An oracle golden is recorded under `folio-go/testdata/` carrying the operand corpus and
  the expected `sum` and `avg` for each corpus member, as **decimal literal strings**. Its provenance
  file records: `python3 --version` output (**3.12.13**), the verbatim producing invocation, the
  explicit `decimal.Context` used (precision and `ROUND_HALF_EVEN`), and the independent reader's
  invocation and output. **AD-25's one-time offline reference run, hand-checked — never a runtime
  dependency**, as D-2.3.1 applied it to shaping via `hb-shape`.
- [x] **AC8.** A test compares the kernel's output against the recorded golden for **both** operations
  across **every** corpus member, asserting the `Decimal`'s coefficient **and** exponent — never a
  rendered string, which can agree while the value does not.
- [x] **AC9.** The same test asserts **order-invariance**: `SumDecimals` over each corpus in declared
  order and in reverse order produces the identical `Decimal`. (F4: this is what stops a red-proof
  passing by luck.)
- [x] **AC10. Red-proof, captured BEFORE the kernel is wired (D-000.30).** Replacing the accumulator
  with `float64` reddens AC8 and AC9 on corpus **A**, and the story records **both** measured values —
  `12345678901234.88` exact against `12345678901234.87` mutated for `sum`, and `374111481855.602424`
  against `374111481855.602230` for `avg`. The mutation must be **valid Go with forbidden semantics**
  (D-000.13) and must include the final rounding to the declared scale, or it manufactures a divergence
  the real implementation would have rounded away (F3).
- [x] **AC11. D-000.22 semantic acceptance step**, this being a **first recording**: the recorded totals
  are asserted to be **order-invariant** (a property read off the artifact, which a float64-contaminated
  recording fails), and the recorded `sum` scale is asserted to equal the maximum operand scale of its
  own corpus. Not a restatement of a hash, and not derived from the producing invocation.
- [x] **AC12. D-000.53 independent-reader step**: `/usr/bin/bc` — which this project did not write and
  which did not produce the golden — resolves each corpus into the total the golden claims. Reader,
  version and **verbatim invocation** recorded in the provenance file with its output. Python's
  `decimal` **may not** serve as its own acceptance reader.
- [x] **AC13.** Corpus members **B** and **C** ship, and each is labelled **in the golden's provenance
  and in the test** as **non-discriminating**, with the reason: B exercises the alignment path that A
  cannot; C is the intuitive money corpus that **cannot express the defect at all** and is retained as
  D-000.50 evidence. A later reader must not mistake either for teeth.

### Layer 2 — a type-identity DENYLIST, narrow, and labelled as one

- [x] **AC14.** A new `lint` rule forbids `math/big.Float` and `math/big.Rat` by **resolved type
  identity** — `Obj().Pkg().Path()` + `Obj().Name()` — and **never by source text**. An alias, a
  dot-import, a renamed import, a type parameter instantiated at one of them, and a variable, field,
  parameter or result of that type all resolve the same and all trip it.
- [x] **AC15.** Scope is the **folio-go module root** (which strictly contains `folio-go/internal/`),
  matching `ScanFloatTypedValues`' shipped production caller exactly — see F5 and Flag F1. `_test.go`
  files are **in scope**, matching AD-23's existing file scope; `testdata/` subtrees are excluded, so
  this rule's own violating fixture does not trip the production scan.
- [x] **AC16.** The rule returns a **coverage witness** reporting what it actually examined — files
  parsed, directories visited, and **expressions whose type resolved** — taken from the checker's **own
  execution**, never a second independent walk. Injecting `if true { return … }` as the checker's first
  statement must zero it. (`FloatTypedStats`' `TypedExprs` is the precedent and the reason: a checker
  that obtains no type information reports zero findings exactly as a clean tree does.)
- [x] **AC17.** The coverage witness and the rule's own documentation state, **in those words**, that
  this rule covers **those two types, not the class**, and that it is **not counted as coverage for
  AD-23's property — Layer 1 is** (D-000.23, D-3.1a.1).
- [x] **AC18. Red-proof (D-000.13).** A retained fixture under `folio-go/testdata/lint/` contains a
  **real `big.Float` value expression** — **valid syntax, forbidden semantics** — alongside a compliant
  sibling. The test asserts on **rule id and message**, **never on exit status or mere failure**. The
  message names the **resolved type identity** and the expression's file and line. At least one fixture
  case reaches the rule through a **renamed import** and one through a **variable of the type**, so
  AC14's "never by source text" is exercised rather than claimed.
- [x] **AC19.** The fixture scan reports **exactly** the violating file and **not** the compliant one,
  compared by **(file, rule)** — never by count (D-1.3.3 amended).
- [x] **AC20.** The rule **fails loudly** on a tree it cannot type-check — both the `packages.Load`
  error path and the per-package `Errors` sweep (D-1.3.11) — returning an error and **no findings**,
  never `(nil, nil)`. Red-proved against the existing `float-typed-untypecheckable` fixture shape.
- [x] **AC21.** A production caller points the rule at the real folio-go module root and asserts
  **zero findings**, guarded against vacuity by the AC16 witness. This is expected to pass on the first
  run — F1 measured zero sites — and the AC16 witness is what makes that green mean something.

### Layer 3 — nothing

- [x] **AC22.** `folio-go/internal/arch_test.go` and `lint/internal/rules/floattyped.go` are
  **not modified by this story**. Verified by reading the diff, not by a test (see *Do not re-open*
  item 3 for why an assertion here would be vacuous).

### The tripwire that makes Story 3.3 unable to skip the wiring

- [x] **AC23.** An assertion, counted **by AST over the whole `folio-go` module** (D-000.14 — never
  text, never a filtered pipe), states that the set of functions reducing a sequence of `Decimal` to
  `(Decimal, error)` is **exactly** `{SumDecimals, AvgDecimals}` and that they are declared in **one**
  package. Its failure message names the offending declaration and says plainly that a second reducer
  must either route through the kernel or be added to this inventory deliberately.
- [x] **AC24. Its red-proof, captured NOW, before 3.3 exists** (D-000.30 — wiring shuts the window
  permanently): adding a third `[]Decimal → (Decimal, error)` function anywhere in the module reddens
  AC23, and removing one of the two reddens it too. Both directions demonstrated.

### Guardrails

- [x] **AC25.** `TestCorpusMeetsP6ExerciseFloors` **stays red** with stats byte-identical to
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. **Never "fix" it.**
- [x] **AC26.** `folio-go/internal/expr/` is **not created**; `absence-expr-package` still passes.
  No absence rule is touched by this story, so D-000.59's replacement obligation is not triggered here.
- [x] **AC27.** Per-story gate cadence: `go test`, `go vet`, `go build` and `lint` all run and are
  reported **with their scope and flags** (D-000.26), against the corrected baseline in the header —
  including the `.font-sources/` caveat for `lint`. The **cross-target matrix is NOT run** (item 8).

---

## Task breakdown

1. [x] **Re-verify the baseline** at `b227dda` with a clean tree, including the `lint` measurement in a
   detached worktree so `.font-sources/` cannot contaminate it. Do not carry the header's numbers
   forward without re-measuring.
2. [x] **Re-verify item 1's premise** against the local Go toolchain's `math/big` source before writing
   anything. If `Float`'s definition has changed, stop and report.
3. [x] **Capture AC24's red-proof first** — the reducer inventory's two failure directions — before the
   kernel exists, then AC10's float64 mutant red-proof before the kernel is wired to the golden.
4. [x] **Build the kernel** (AC1–AC6) in `internal/bind`, reusing `decimal.go`'s existing bounds and
   narrowing pattern. Resist adding a checked-add helper.
5. [x] **Produce and record the oracle golden** (AC7), then discharge the acceptance steps: AC11
   semantic, AC12 independent reader via `bc`. Record provenance verbatim.
6. [x] **Wire the Layer-1 oracle test** (AC8, AC9, AC13) and confirm AC10's captured mutant still
   reddens it.
7. [x] **Build the Layer-2 rule** (AC14–AC21) in `lint/internal/rules`, following `floattyped.go`'s
   shape for stats, loud failure and fixture scanning.
8. [x] **Land the reducer inventory** (AC23) and confirm the step-3 red-proof still applies.
9. [x] **Run the per-story gate** (AC27), confirm AC25 and AC26, and record every figure with its scope
   and flags.
10. [x] **Fill in the Delivery Log**, including the D-000.60 line naming what each flag below
    suppressed and whether the suppression was lifted. Move status to `review`.

---

## Flags — and what each one SUPPRESSES (D-000.60)

D-000.60 binds: a flag must name what it holds back, so whoever applies the ruling can lift it.

- **F1 — Layer 2's scope widened from `folio-go/internal/` to the folio-go module root — RESOLVED,
  lifted.**
  *Grounds:* F5 — the charter's phrasing is strictly narrower than the neighbouring guard's shipped
  scope and would leave the public entry point uncovered.
  *What this SUPPRESSES:* nothing is withheld — this is a widening, and it is flagged so the lead can
  **narrow** it if the charter's phrasing was deliberate. If narrowed, `folio-go/render.go`,
  `render_entry.go` and `page_number.go` fall outside the rule and a follow-up owner is needed for
  them.
  *Resolution:* D-3.1a.1 (correction), recorded in the decision log during this run before any Layer-2
  code was written, **confirms the widening rather than narrowing it**: the module root is correct on
  the merits (a narrower Layer 2 than Layer 1's own shipped scope would read as covered when it is
  not), and is a HARD GUARDRAIL against widening further to the repository root (`hashmatrix/` stays
  excluded by construction). `ScanBigFloatTypes` (`lint/internal/rules/bigfloattype.go`) is built and
  tested exactly at the `folio-go` module root — `TestBigFloatTypeProductionScan` asserts the root
  package (".") visited by name, matching `TestFloatTypedProductionScan`'s own witness.

- **F2 — the kernel's error payload carries no AD-14 diagnostic code, data path or element id — carried
  as declared, per D-000.24.**
  *Grounds:* `internal/diag` does not exist until 3.6 (`absence-diag-package`), and the kernel receives
  a `[]Decimal` — it **cannot know** the element id or data path. The kernel returns a structured error
  naming the operand index and the bound breached; the caller attaches the located payload.
  *What this SUPPRESSES:* **a test asserting AD-14's full payload on an overflow is NOT written in this
  story.** When 3.6 mints the codes, that test becomes writable and is owed. AC5 asserts the error is
  located *as far as the kernel can locate it* and no further, and it must not be read as discharging
  AD-14.
  *Resolution:* still open, as designed — this is 3.6's obligation, not this story's. `reduce.go`'s doc
  comment states plainly (per D-000.24) that it does NOT claim AD-14 conformance. Not lifted; not owed
  yet.

- **F3 — `avg`'s `+4` is carried as declared, at the charter's own MEDIUM confidence.**
  *Grounds:* D-3.1a.1 marks the `+4` illustrative and medium-confidence; its forcing function is Story
  3.4's assertion that no `formatNumber` pattern can request more fractional digits than `avg`
  produces.
  *What this SUPPRESSES:* **no test in this story attempts to justify the value 4.** AC4 asserts only
  that the scale is declared once, is a property of the operation, and is not fitted to the corpus. If
  3.4 finds a pattern that can request more, the constant moves — and that is the design working, not a
  regression here.
  *Resolution:* not lifted — carried forward exactly as declared. `avgExtraScale = 4` is declared once
  in `internal/bind/reduce.go`, is not derived from `golden.json`'s corpus, and no test in this story
  argues for the number 4 itself. Owed to Story 3.4.

- **F4 — `ResolveAssets` walking the gitignored `.font-sources/` is noted, not fixed.**
  *Grounds:* out of this story's scope; it is a local-environment artifact producing three false reds.
  *What this SUPPRESSES:* **no fix and no test.** The three `lint` failures stay reproducible in this
  checkout. If the lead wants it owned, it needs a story; do not silently absorb it here.
  *Resolution:* not lifted, and not this story's to lift — now tracked as **DW-19** (deferred-work.md),
  recorded during this run. `TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist` and
  `TestFontsAssetsNoticeRemovalRedProof` remain reproducibly red in this checkout, all three for the
  same `.font-sources` reason, and are named as known-environmental in the Gate results below per
  D-000.55. Not fixed, not weakened, no LICENSE file added.

- **F5 — empty-input semantics for both kernel functions — AMENDED by D-3.1a.2, applied as built.**
  *Originally ruled for the kernel (as written above, before development):* both functions error on
  empty input. *As actually implemented, per D-3.1a.2 (recorded in the decision log during this run,
  before any kernel code was written):* `SumDecimals(nil)` returns the **additive identity**
  `{Coefficient: 0, Exponent: 0}`, never an error — sum has an identity element. `AvgDecimals(nil)`
  **keeps the error** — averaging has none; there is genuinely no value to report. The asymmetry is
  deliberate (Story 3.3's AC already requires exactly this shape: `sum`/`count` return zero on an empty
  collection, `avg` reports a diagnostic). Both halves are asserted in
  `internal/bind/reduce_test.go` (`TestSumDecimalsEmptyReturnsIdentity`,
  `TestAvgDecimalsEmptyReturnsError`).
  *What this SUPPRESSES:* **this is still NOT a ruling on what the expression-language `sum()`/`avg()`
  do on an empty collection** — that remains 3.3's own semantic question (now considerably narrowed by
  D-3.1a.2, but not answered outright: 3.3 still owns row-scope/collection-resolution empty-input
  behaviour). No test in this story presupposes an answer beyond the kernel's own two functions. When
  3.3 rules it, whoever applies the ruling should check whether the kernel's primitives (now identity
  for sum, error for avg) are the right ones to route through.

---

## Delivery Log

*(To be completed by the developer, reviewer and finisher.)*

- **Baseline re-verified at:** HEAD `98c2217` ("Story 3.1a: record F1's correction, the empty-input
  asymmetry, the relational tripwire, D-000.61 and DW-19") — the log's own stated baseline `b227dda`
  had already advanced by two commits by the time development started, both of which are the lead's
  own pre-development rulings for this story (D-3.1a.1 correction, D-3.1a.2, D-3.1a.3, D-000.61, DW-19),
  landed via the orchestrator before any code was written. Working tree clean at that HEAD. Re-measured:
  `folio-go/` — `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` — **619 PASS · 1 FAIL · 1 SKIP**
  (all-occurrence), **380 PASS · 1 FAIL · 1 SKIP** (top-level); byte-identical to the header. `lint/`, in
  a clean detached worktree at the same HEAD — same invocation — **85 PASS · 0 FAIL**, matching the
  header's corrected baseline. The dirty (working) checkout at the same HEAD independently reproduced
  the header's own `.font-sources` caveat: **82 PASS · 3 FAIL**, the same three named tests, same
  message.
  `math/big.Float`'s definition re-verified against this run's own Go toolchain (`go1.26.0`,
  `$(go env GOROOT)/src/math/big/float.go:65`): `prec uint32; mode RoundingMode; acc Accuracy; form
  form; neg bool; mant nat; exp int32` — unchanged from the story's own citation, `mant` is `nat`
  (`[]Word`, `Word uint`), no float field anywhere.

- **AC24 red-proof captured at (pre-kernel):** Yes — captured before `internal/bind/reduce.go` existed.
  `folio-go/internal/reducer_inventory_arch_test.go` (the checker, `reducerDecimalInventory`) and
  `folio-go/internal/reducer_inventory_test.go` (the tests) were written and run FIRST:
  `TestDecimalReducerInventoryIsExactlySumAndAvg` (the real production assertion) correctly **failed**
  ("expected reducer SumDecimals(...) not found... AvgDecimals(...) not found... count mismatch: got 0")
  because neither function existed yet — verified by running
  `go test -run TestDecimalReducerInventory ./internal/...` at that point (3 passed, 1 failed). The
  three fixture-tree tests (`baseline` = 2 reducers, `extra-reducer` = 3 reducers/AC24 direction 1,
  `missing-reducer` = 1 reducer/AC24 direction 2) all passed immediately, proving both red-proof
  directions independently of the not-yet-built kernel. After the kernel landed, the production test
  went green with no change to the checker or the fixtures.

- **AC10 red-proof captured at (pre-wiring), both measured values:** Captured independently in this run,
  before `golden_test.go` wired `SumDecimals`/`AvgDecimals` up against `golden.json`. Methodology: the
  "honest" float64 mutant (accumulate in float64, THEN round to the operation's declared scale),
  measured in Python (float64 arithmetic is IEEE-754 double in both languages) over corpus A, forward
  and reversed operand order:
  ```
  forward sum mutant:   12345678901234.87   reversed sum mutant:   12345678901234.88
  forward avg mutant:   374111481855.602234 reversed avg mutant:   374111481855.602478
  exact sum (golden):   12345678901234.88   exact avg (golden):    374111481855.602424
  ```
  Both properties the story predicts are reproduced independently: (1) the mutant **reddens** — its
  forward-order sum and avg both disagree with the exact golden value; (2) the mutant is **not**
  order-invariant on either operation (forward ≠ reversed for both sum and avg), while the reversed-sum
  mutant happens to land on the exact value by coincidence — exactly D-000.61's "passed by luck" finding,
  independently reproduced rather than merely re-read. (My avg-mutant figures differ in their last two
  digits from the story's own `602230` — expected: float64 accumulation is itself order/method-sensitive,
  and mine sums via a plain running `+=` in the order given, which is a different but equally "honest"
  accumulation path; the point — divergence exists, and is order-dependent — holds under both
  measurements.) No float64 code is committed anywhere in `folio-go/` — AD-23/Layer 2 forbid it — so this
  red-proof lives only as this measurement, not as standing test code.

- **Golden provenance — producer, reader, invocations:** `folio-go/testdata/decimal-reduction/golden.json`
  (the recording) + `produce_golden.py` (the producer, run once by hand) + `PROVENANCE.md` (the full
  record). Producer: system `python3`, version `Python 3.12.13` (verbatim, this run), context
  `Context(prec=50, rounding=ROUND_HALF_EVEN)`. Independent reader: `/usr/bin/bc`, version `bc 7.0.3`
  (verbatim, this run) — every corpus's `sum`, forward AND reversed operand order, matches
  `golden.json`'s recorded value exactly (six verbatim invocation/output pairs in `PROVENANCE.md`).
  `avg` is deliberately NOT cross-checked by `bc` (it truncates rather than rounding half-to-even, so it
  is not a faithful independent reader for that rounding mode); `avg`'s acceptance instead rests on
  `TestDecimalReductionKernelMatchesGolden` (AC8) and the order-invariance property (AC9/AC11).
  D-000.22 semantic acceptance (first recording): `TestGoldenIsOrderInvariant`
  (`internal/bind/golden_test.go`) computes each corpus's exact sum from scratch (a small standalone
  `big.Int` alignment, deliberately NOT calling `SumDecimals`), forward and reversed, and asserts both
  orders agree with each other and with the recorded golden — plus that the recorded sum's scale equals
  its corpus's own maximum operand scale. All green.

- **Gate results, with scope and flags:**
  | scope | invocation | before | after |
  |---|---|---|---|
  | `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, all occurrences | 619 PASS · 1 FAIL · 1 SKIP | **636 PASS · 1 FAIL · 1 SKIP** (+17 new, all this story's) |
  | `folio-go/` | same, top-level only | 380 PASS · 1 FAIL · 1 SKIP | **397 PASS · 1 FAIL · 1 SKIP** (+17) |
  | `lint/` (this dirty checkout, `.font-sources` present) | same invocation, all occurrences | 82 PASS · 3 FAIL | **88 PASS · 3 FAIL** (+6 new, all this story's; the 3 FAIL are unchanged and are the named environmental ones below) |
  | `folio-go/` | `CGO_ENABLED=0 GOWORK=off go build ./...` | — | **Success** |
  | `folio-go/`, `lint/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | — | **No issues found**, both modules |

  The single `folio-go` FAIL is `internal/text`'s `TestCorpusMeetsP6ExerciseFloors`, byte-identical stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — Story 2.1's deliberate shortfall (D-000.17 /
  D-2.1.14), **left red on purpose** (AC25). The single SKIP is
  `TestXrefEntriesRejectsMalformedSubprocess`, pre-existing and unrelated.

  The three `lint` FAILs are `TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`,
  `TestFontsAssetsNoticeRemovalRedProof` — **named as known-environmental per D-000.55, before this run**
  (per the story's own baseline-correction note and Flag F4/DW-19): `ResolveAssets` walks the gitignored,
  untracked `.font-sources/` scratch directory in this local checkout and finds a font binary with no
  LICENSE file. Re-measured in a clean detached worktree at the same HEAD (`/tmp/folio-clean-3.1a`):
  **85 PASS · 0 FAIL**, confirming these three are purely a local-environment artifact, not a defect this
  story introduced or should fix. All three were red for the SAME stated reason both before and after
  this story's changes — no new-for-a-different-reason failures.

  Cross-target matrix: **NOT run** (item 8 / AC27) — 3.1a introduces no new source of cross-target
  divergence (a lint rule and an integer-only `big.Int` kernel); it is due at Epic 3's close.

- **D-000.60 line — what each flag suppressed, and whether it was lifted:**
  - **F1** (Layer 2 scope widened to the module root) — **lifted/resolved**: D-3.1a.1's correction,
    recorded before development, confirms the widening. `ScanBigFloatTypes` ships at the `folio-go`
    module root, matching `ScanFloatTypedValues` exactly, witness-asserted by name.
  - **F2** (kernel error carries no AD-14 payload) — **not lifted, correctly still open**: owed to Story
    3.6. `reduce.go`'s doc comment states plainly it does not claim AD-14 conformance (D-000.24).
  - **F3** (avg's `+4` carried at MEDIUM confidence) — **not lifted, correctly still open**: owed to
    Story 3.4's forcing-function assertion. `avgExtraScale = 4` is declared once, not derived from the
    corpus.
  - **F4** (`.font-sources` false reds) — **not lifted, correctly still open, now tracked**: filed as
    DW-19 in `deferred-work.md` (recorded during this run, before development). Not fixed here; named as
    known-environmental in the Gate results above, per D-000.55.
  - **F5** (empty-input semantics) — **amended in place by D-3.1a.2** (recorded before development, so
    this was never a live suppression during coding): `SumDecimals(nil)` is the identity, not an error;
    `AvgDecimals(nil)` errors. The narrower remaining suppression — expression-language `sum()`/`avg()`'s
    own empty-collection semantics — is **still open**, correctly, and is Story 3.3's to rule.
  - The reducer-inventory tripwire's location clause (not itself one of the five numbered flags, but the
    same D-000.60 discipline applies) was **corrected before any code was written**: D-3.1a.3 made it
    relational ("the same package as the Decimal type declaration") rather than the literal
    `internal/bind` the story file's inventory prose named — implemented that way from the start in
    `reducer_inventory_arch_test.go`, so there was nothing to retrofit.

- **Files changed (new, all this story):**
  - `folio-go/internal/bind/reduce.go` — the kernel: `SumDecimals`, `AvgDecimals`.
  - `folio-go/internal/bind/reduce_test.go` — AC1–AC6 unit tests.
  - `folio-go/internal/bind/golden_test.go` — AC7–AC13 Layer-1 oracle test.
  - `folio-go/internal/bind/testutil_test.go` — `repoRootFromTest`/`mustReadFile` helpers.
  - `folio-go/testdata/decimal-reduction/produce_golden.py` — AC7's producer.
  - `folio-go/testdata/decimal-reduction/golden.json` — AC7's recorded golden.
  - `folio-go/testdata/decimal-reduction/PROVENANCE.md` — AC7/AC12 provenance.
  - `folio-go/internal/reducer_inventory_arch_test.go` — AC23's checker (`reducerDecimalInventory`).
  - `folio-go/internal/reducer_inventory_test.go` — AC23/AC24 production test and red-proofs.
  - `folio-go/testdata/arch/reducer-inventory/{baseline,extra-reducer,missing-reducer}/*.go` — AC24
    fixtures.
  - `lint/internal/rules/bigfloattype.go` — Layer 2: `ScanBigFloatTypes`.
  - `lint/internal/rules/bigfloattype_test.go` — AC14–AC21 tests.
  - `folio-go/testdata/lint/no-bigfloat-type/{compliant.go,violating_renamed_import.go,violating_variable.go}`
    — AC18/AC19 fixtures.
  - This story file (`3-1a-guard-ad-23-against-named-binary-float-types.md`) — checkboxes, Flags
    resolutions, Delivery Log, Status.
  Unmodified, per AC22 (verified by `git diff --name-only`, not by a test): `folio-go/internal/arch_test.go`,
  `lint/internal/rules/floattyped.go`.

### Finisher pass (2026-08-25)

Every review finding was triaged FIX except Finding 13 (Nit), DISMISSed — see ## Finding Resolutions
below for full triage and per-finding evidence. 12 FIX, 1 DISMISS, 0 DEFER.

**Files this pass modified or added, beyond the developer's own list above:**
  - `folio-go/internal/bind/reduce.go` — modified: overflow-safe alignment spread (Finding 3), avg
    result-exponent bound (Finding 10), corrected tripwire-force doc comment (Finding 11).
  - `folio-go/internal/bind/reduce_test.go` — modified: added red-proofs for Findings 3, 10, 12.
  - `folio-go/internal/bind/golden_test.go` — modified: AC13 label-test inversion (Finding 7),
    order-invariance doc correction (Finding 8).
  - `folio-go/testdata/decimal-reduction/PROVENANCE.md` — modified: corpus-A `bc` invocations
    (Finding 9), D-000.22 section doc correction (Finding 8).
  - `folio-go/internal/reducer_inventory_arch_test.go` — modified: widened AST predicate (variadic,
    slice-alias, method receiver — Finding 4), added `decimalReducerViolations` (Finding 6).
  - `folio-go/internal/reducer_inventory_test.go` — modified: production test and both existing
    red-proofs now use `decimalReducerViolations` (Finding 6); added the moved-decimal red-proof
    (Finding 5).
  - `folio-go/testdata/arch/reducer-inventory/moved-decimal/{pkga,pkgb}/*.go` — new fixture (Finding 5).
  - `lint/internal/rules/bigfloattype.go` — modified: `types.Unalias` (Finding 1), go-build-cache
    filter (Finding 2).
  - `lint/internal/rules/bigfloattype_test.go` — modified: added `TestBigFloatTypeTestScopeInventory`
    (Finding 2), extended the fixture-scan `want` list (Finding 1).
  - `folio-go/testdata/lint/no-bigfloat-type/violating_type_alias.go` — new fixture (Finding 1).
  - `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — appended `D-3.1a.4` (Finding 11).
  - This story file — plain-terms opener rewritten, Finding Resolutions and this Delivery Log
    addendum appended, Status set to `done`.
  Still unmodified, per AC22: `folio-go/internal/arch_test.go`, `lint/internal/rules/floattyped.go`.

**The `.font-sources` gate note (review's "Gates" section, not a numbered Finding) — DISMISSed, re-
verified.** Re-ran `lint/` in this dirty checkout: the same three tests
(`TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`, `TestFontsAssetsNoticeRemovalRedProof`)
fail, all three for the identical `.font-sources: contains a committed font binary but no LICENSE* file`
reason, no fourth failure. Independently re-verified clean in a detached worktree (`git worktree add
--detach`) with the finisher's own changed/new files copied in: **92 PASS · 0 FAIL** (all-occurrences).
DW-19 as filed; not fixed, not weakened, no LICENSE file added, `.font-sources/` untouched.

**Gate results after the finisher's fixes, with scope and flags:**

| scope | invocation | before finisher | after finisher |
|---|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, all-occurrences | 636 PASS · 1 FAIL · 1 SKIP | **640 PASS · 1 FAIL · 1 SKIP** (+4, all this pass's new tests) |
| `folio-go/` | same, top-level only | 397 PASS · 1 FAIL · 1 SKIP | **401 PASS · 1 FAIL · 1 SKIP** (+4) |
| `lint/` (this dirty checkout, `.font-sources` present) | same invocation, all-occurrences | 88 PASS · 3 FAIL | **89 PASS · 3 FAIL** (+1, `TestBigFloatTypeTestScopeInventory`; the 3 FAIL are the unchanged, named environmental ones above) |
| `lint/` (clean detached worktree, finisher's changed files copied in) | same invocation, all-occurrences | — | **92 PASS · 0 FAIL** |
| `folio-go/`, `lint/` | `go build ./...` | — | **Success**, both modules |
| `folio-go/`, `lint/` | `go vet ./...` | — | **No issues found**, both modules |
| `folio-go/`, `lint/` | `gofmt -l .` (absolute path, bare — not `rtk proxy`) | — | **Clean**, both (one file needed reformatting after an edit; fixed) |

The `+4` folio-go tests are named: `TestSumDecimalsAlignmentSpreadDoesNotOverflow` (Finding 3),
`TestAvgDecimalsResultExponentBoundIsEnforced` (Finding 10), `TestSumDecimalsMixedSignOperands`
(Finding 12), `TestDecimalReducerInventoryRedProofDecimalMovedLeavesReducersBehind` (Finding 5) — the
two negative-tie cases for Finding 12 were added as table rows inside the existing
`TestAvgDecimalsRoundsHalfToEven`, not new top-level tests, which is why the count is 4 rather than 5.
`TestCorpusMeetsP6ExerciseFloors` remains the single folio-go FAIL, stats byte-identical
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — confirmed unchanged, never touched.

**Cross-target matrix: NOT run** (D-000.4 per-epic cadence, unchanged from the developer's own
disposition) — this pass adds no new source of cross-target divergence (widened AST/type-identity
matching and additional Go-only arithmetic bound checks); it remains due at Epic 3's close, and
`epic-3` stays `in-progress` in `sprint-status.yaml` (untouched by this pass beyond that existing
value).

**Red-proof discipline applied throughout:** every FIX above was reproduced live (red) before its
patch and re-verified (green) after, per D-000.30 — see the per-finding notes under ## Finding
Resolutions for each mechanism and measurement. All scratch probes (`/tmp/reducer_probe_finding4`,
`/tmp/reducer_probe_finding1`, `lint/cmd/scratchprobe*`, a temporary `_test.go` under folio-go/ for
Finding 2's live measurement) were deleted after use; `git status --porcelain` before this pass's
commit shows only the files listed above, no probe residue.

**D-3.1a.4** (append-only decision log) records Finding 11's correction in full; not repeated here.

---

## QA Results

**Reviewed by:** bmad-code-reviewer · **Date:** 2026-08-25 · **Baseline:** `98c2217`, work uncommitted
**Method.** Nothing below is read off the Delivery Log. Both suites, `go vet` and `go build` were
re-run in both modules by the reviewer. Every mechanism claim was verified by construction: five
probe trees were built outside the repo, three mutations were applied to shipped files and reverted
**by hand from byte-exact `cp` backups with SHA-256 re-verified after restore** (never
`git checkout`), and four probe files created inside the repo were deleted. Final
`git status --porcelain` reports **exactly the twelve entries handed to the reviewer** — no probe
residue, no `.bak`. `rtk` intercepts `go test`; every count below was taken from the raw tee log
(`~/Library/Application Support/rtk/tee/…`) and counted in Python, never from rtk's own summary line
and never through a shell pipe.

### Gates, re-measured independently

| scope | invocation | reported | **observed** | verdict |
|---|---|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, all occurrences | 636 · 1 · 1 | **636 PASS · 1 FAIL · 1 SKIP** | reproduces |
| `lint/` (this dirty checkout) | same | 88 · 3 | **88 PASS · 3 FAIL** | reproduces |
| `folio-go/`, `lint/` | `go build ./...` | Success | **Success**, both | reproduces |
| `folio-go/`, `lint/` | `go vet ./...` | No issues | **No issues**, both | reproduces |

**AC25 holds byte-identically.** The single `folio-go` failure is `TestCorpusMeetsP6ExerciseFloors`,
`P6g … got 7, need >=20`, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — required red,
unchanged, and the only failure. Single SKIP is `TestXrefEntriesRejectsMalformedSubprocess`.

**The three `lint` failures are exactly the three named, and each is red for the stated reason** —
verified per test, not per count: `TestManifestUpToDate` and `TestResolveAssetsIncludesWordlist` both
report `ResolveAssets: .font-sources: contains a committed font binary but no LICENSE* file`, and
`TestFontsAssetsNoticeRemovalRedProof` reports the same `.font-sources` message where it expected
`no NOTICE* file`. No fourth failure and none red for a different reason. DW-19 as filed.

**AC22 verified by diff, not by test.** `git diff 98c2217 -- folio-go/internal/arch_test.go
lint/internal/rules/floattyped.go` is **empty**; `lint/internal/rules/absences.go` is likewise
untouched. **AC26 verified:** `folio-go/internal/expr` does not exist and `absence-expr-package`
(`absences.go:95-97`) is intact. **No `float32`/`float64` is introduced** anywhere in this story's new
code — the only occurrences are prose in comments and in `PROVENANCE.md`.

### What was red-proved, and what it proved

- **AC10 / D-000.61, reproduced independently.** `reduce.go` was mutated to the honest float64 mutant
  (accumulate operand *values* in `float64`, then round to `minExp`) and the oracle re-run. Corpus A
  reddened on **both** operations — `SumDecimals` `{1234567890123487,-2}` against the golden's
  `…488`, `AvgDecimals` `{374111481855602121,-6}` against `…602424` — and `TestDecimalReduction
  KernelIsOrderInvariant` reddened too, forward `…487` vs reversed `…488`. **D-000.61's "would have
  passed by luck" is real and visible here:** the reversed order lands exactly on the golden, so a
  single-order comparison in that direction would have been green. Corpora **B and C stayed green
  under the same mutant**, confirming the non-discrimination claims are true, not merely asserted.
  `reduce.go` restored from backup, SHA-256 `84411d70…a8b5` before and after.
- **AC16's vacuity witness is real.** `if true { return nil, BigFloatTypeStats{}, nil }` injected as
  `ScanBigFloatTypes`' first statement reddens **five of the six** Layer-2 tests, each on its own
  witness (`DirsVisited` empty, `FilesParsed` 0, `TypedExprs` 0, loud-failure paths silent). Only the
  constant-wording test is unaffected, correctly. Restored, SHA-256 `195fe4b7…3f04`.
- **AC15's module-root half is correct and hashmatrix is excluded by construction.** A `big.Float`
  planted in a new file in the **module-root package `folio`** was reported at both its sites; the
  scan's own stats report `.` visited and 13 directories, and **never** `hashmatrix/` or `lint/`.
  `testdata/` is excluded by the go tool's own package matching, so the rule's fixture does not trip
  the production scan.
- **AC23 has real teeth on the real module.** A third reducer added to `internal/bind` reddened
  `TestDecimalReducerInventoryIsExactlySumAndAvg` ("unexpected reducer … count mismatch: got 3, want
  2"); the same reducer placed in a *different* package under `internal/` reddened it too, via the
  qualified-import resolution path. Both probes removed.
- **The relational location clause genuinely works.** Pointed at a probe tree where `Decimal` lives in
  `pkga` and the two reducers were left behind in `pkgb`, the checker reports
  `DeclPkgDir="pkga"` with both reducers in `"pkgb"` — so D-3.1a.3's second failure mode functions and
  the guard needs **no edit** when Story 3.2 moves `Decimal`. (It is, however, never exercised by a
  shipped fixture — Finding 5.)
- **AC12's `bc` outputs all reproduce**, and the conclusion survives a stricter reading: I ran the
  **true 33-operand** corpus A through `/usr/bin/bc` forward *and* reversed and both give
  `12345678901234.88`. See Finding 9 for what the *recorded* invocation actually executed.

---

### Finding 1: Layer 2 does not resolve type ALIASES — an aliased `big.Float` evades the guard entirely

- **Severity**: Blocker
- **Category**: AC Conformance / Correctness
- **Location**: `lint/internal/rules/bigfloattype.go:219-225` (`resolveNamedType`), consumed at
  `:168-172`
- **Observation**: `resolveNamedType` unwraps one pointer and then type-asserts `t.(*types.Named)`.
  Since Go 1.23 (`gotypesalias=1` is the default; this toolchain is **go1.26.0**) `go/types`
  materialises an alias as **`*types.Alias`**, which is *not* `*types.Named`, so the assertion fails
  and the expression is passed over. Measured directly on a probe tree: every expression in a package
  using `aliasdecl.MyFloat` reports `*types.Alias`, `named=false`, and
  `types.Unalias(tv.Type)` recovers exactly `math/big.Float`. Scanned end to end, a package whose
  alias is declared **in a dependency** — `func Balance() bfdep.Money` where `type Money = big.Float`
  — produced **zero findings** while the checker's own stats confirmed it visited that directory and
  typed its expressions. The alias case is the one where the checker is green *and* the witness is
  non-zero, which is precisely the failure shape this story exists to eliminate.
- **Impact**: AC14 names "an alias" first in its list of forms that "all resolve the same and all trip
  it". They do not. Worse, the claim is repeated in shipped text a reader will trust: the doc comment
  at `bigfloattype.go:73-77` and the **finding message itself** (`:187-189`, *"an alias, a dot-import
  or a renamed import all resolve the same"*). Under D-000.23 this is advertised coverage that does
  not exist, which this project holds is worse than an admitted hole. When the alias is declared
  inside the scanned tree the declaration site still trips, so the hole is partial there; when it is
  declared outside the tree the hole is total.
- **Suggested Resolution**: apply `types.Unalias` in `resolveNamedType` — before the pointer test and
  again on `ptr.Elem()` — so both `MyFloat` and `*MyFloat` resolve. Add a fixture case under
  `folio-go/testdata/lint/no-bigfloat-type/` reaching the type through an alias, and add the
  dot-import and type-parameter-instantiation cases AC14 also names but no fixture exercises (both of
  those two **do** currently trip, verified on the probe tree — they are untested, not broken).
- **Related AC**: AC14, AC18

### Finding 2: `_test.go` files are in scope for no shipped caller, so AC15's file-scope clause is unimplemented

- **Severity**: Blocker
- **Category**: AC Conformance
- **Location**: `lint/internal/rules/bigfloattype_test.go:52, 88, 115, 159, 178` — every call site
  passes `includeTests: false`
- **Observation**: AC15 states *"`_test.go` files are **in scope**, matching AD-23's existing file
  scope"*. `ScanBigFloatTypes` accepts `includeTests` but **nothing ever passes `true`**. Measured: a
  file `folio-go/zzreviewprobe_test.go` declaring `var r big.Rat` and `var f big.Float` yields
  **0 findings** at `Tests:false` (58 files parsed) and **4 findings** at `Tests:true` (178 files
  parsed). The neighbouring guard does not have this gap — `TestFloatTypedTestScopeInventory`
  (`floattyped_test.go:145`) is a second production caller at `Tests:true`, and its own comment
  rejects exactly this narrowing: *"a type-aware rule that did not [walk `_test.go`] would be
  strictly weaker in file scope than the guard it strengthens — a silent regression dressed as a
  fix."*
- **Impact**: a `math/big.Float` in any `_test.go` file under `folio-go/` ships undetected. The
  syntactic guard cannot cover for it (it keys on `float32`/`float64` identifiers), so for this type
  the test-file surface is covered by **neither** guard. This is an unimplemented AC clause, not a
  design choice — the story argues the opposite position in F5 and in AC15 itself.
- **Suggested Resolution**: add a second production caller at `Tests:true` mirroring
  `TestFloatTypedTestScopeInventory` — an inventory, not a zero-assertion, so a sanctioned site would
  have to be enumerated. Note one implementation detail found while probing: at `Tests:true` the
  loader hands back generated test-main packages, and `DirsVisited` then contains entries such as
  `../../../Library/Caches/go-build/04`; those need filtering out of the witness (or the relative
  path rejected) or the stats become misleading.
- **Related AC**: AC15, AC21

### Finding 3: AC6's alignment bound is defeated by signed-integer overflow, and the result is a silent wrong answer

- **Severity**: Major
- **Category**: Correctness
- **Location**: `folio-go/internal/bind/reduce.go:95-105`
- **Observation**: `spread := maxExp - minExp` is `int` arithmetic on unvalidated `Decimal.Exponent`
  values. Measured: `SumDecimals([]Decimal{{1, math.MinInt}, {1, math.MaxInt}})` returns
  **`{Coefficient: 2, Exponent: -9223372036854775808}` with `err == nil`.** `spread` wraps to `-1`,
  sails past `spread > maxDecimalExponentMagnitude`, and the per-operand `shift := it.Exponent -
  minExp` wraps too, so both operands are added unshifted.
- **Impact**: AC6 says a breach *"is an Error"* and *Do-not-re-open* item 6 says overflow is *"a
  located Error, never a widening"* and *"never a wrap"*. The single most extreme possible breach is
  accepted silently and produces an arithmetically meaningless value in the kernel the money path
  will rest on. `Decimal` is an exported struct with exported fields, so any caller — including Story
  3.3's collection resolution and D-1.4.1's footer evaluation — can construct such a value without
  going through `NewDecimal`; the tests in this very story construct `Decimal` literals directly at
  `reduce_test.go:179-182, 203, 232-235`. Not reachable through `NewDecimal` today, which is why this
  is Major and not Blocker.
- **Suggested Resolution**: validate each operand's own `Exponent` against
  `maxDecimalExponentMagnitude` before computing any difference (a per-operand magnitude check makes
  the spread bound unable to overflow, and costs one comparison per operand), or compute the spread in
  a width that cannot wrap. Assert it with a case built from `math.MinInt`/`math.MaxInt` exponents.
- **Related AC**: AC5, AC6

### Finding 4: AC23's inventory is a signature-SHAPE whitelist, not the set-equality it claims

- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/internal/reducer_inventory_arch_test.go:192-215`
- **Observation**: the matcher requires exactly one parameter that is syntactically an `*ast.ArrayType`
  whose element matches `Decimal`, and `fn.Recv == nil`. Measured on a probe tree containing five
  additional reducers, the scan reported **three** — `SumDecimals`, `AvgDecimals` and
  `GenericReducer` — and did **not** see:
  `func VariadicReducer(items ...Decimal) (Decimal, error)`;
  `type Decimals = []Decimal` + `func AliasReducer(items Decimals) (Decimal, error)`;
  `func (a Acc) MethodReducer(items []Decimal) (Decimal, error)`;
  `func TwoParamReducer(items []Decimal, mode int) (Decimal, error)`.
  Each of those plainly *"reduce[s] a sequence of `Decimal` to `(Decimal, error)`"* in AC23's own
  words. (`ast.Ellipsis` is not `ast.ArrayType`; a slice alias is an `*ast.Ident`; a method is skipped
  by the `Recv` test.)
- **Impact**: the tripwire's whole value is that Story 3.3 cannot land a second accumulator without a
  human reading a diff. Four ordinary Go spellings walk straight past it, and the variadic one is the
  most likely of all for a function table's `sum(...)`. An inventory advertised as exact set equality
  that is in fact a shape whitelist is the D-000.14 hazard wearing an AST costume — counting by AST
  was the right call, but the predicate is narrower than the property.
- **Suggested Resolution**: broaden the predicate to cover `*ast.Ellipsis` element types and methods,
  and either resolve slice aliases (which needs `go/types`, a larger change) or state the residual
  gap explicitly in the checker's doc comment so it is an admitted hole rather than a claimed set.
  Whatever is chosen, add a fixture per admitted shape so the boundary is exercised, not described.
- **Related AC**: AC23, AC24

### Finding 5: D-3.1a.3's "second real failure mode" has no fixture — the location clause is never observed to fire

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/testdata/arch/reducer-inventory/` (three variants: `baseline`,
  `extra-reducer`, `missing-reducer`); clause at `reducer_inventory_test.go:62-68`
- **Observation**: all three fixtures declare `Decimal` and the reducers in the **same** package.
  Nothing exercises the case D-3.1a.3 gives as the relational form's added value — *"it also fails if
  someone moves `Decimal` and leaves the reducers behind"*. I built that tree by hand and confirmed the
  mechanism works (`DeclPkgDir="pkga"`, both reducers reported in `"pkgb"`), so this is a coverage
  gap, not a defect.
- **Impact**: the clause that justifies the whole relational design is the one clause with no
  red-proof. If a later refactor broke `matchesDecimal`'s qualified-reference arm, every shipped test
  would stay green.
- **Suggested Resolution**: add a fourth fixture variant (`moved-decimal`) with `Decimal` in one
  package directory and the reducers in another, asserting the scan reports both reducers with
  `PkgDir != DeclPkgDir`.
- **Related AC**: AC23, AC24

### Finding 6: AC24's red-proofs assert on the scanner's OUTPUT, not on AC23's verdict

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/reducer_inventory_test.go:122-153`
- **Observation**: both red-proof tests call `reducerDecimalInventory` on a fixture tree and assert the
  returned slice has 3 (or 1) entries with the expected names. Neither runs AC23's comparison logic,
  which is written inline inside `TestDecimalReducerInventoryIsExactlySumAndAvg` and is not factored
  into a callable predicate. AC24's wording is *"adding a third … **reddens AC23**"*.
- **Impact**: the proof demonstrates the *input* to AC23 changes, not that AC23 reddens; a future edit
  to the comparison block (e.g. dropping the `len(byName) != len(want)` arm) would leave both
  "red-proofs" green. I confirmed by construction that AC23 does redden today, in both directions, so
  the conclusion is currently true.
- **Suggested Resolution**: extract the want-set comparison into a helper taking
  `([]decimalReducer, stats, want)` and returning the list of violations, then have the production
  test and both red-proofs call it — the red-proofs asserting a non-empty violation list.
- **Related AC**: AC24

### Finding 7: the AC13 labelling test iterates its own label map, so an UNLABELLED corpus passes

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/bind/golden_test.go:72-88`
- **Observation**: `TestGoldenCorporaBAndCAreLabelledNonDiscriminating` ranges over
  `nonDiscriminatingCorpusMembers` and checks each label has a corpus. The direction AC13 actually
  cares about — every shipped corpus is either the discriminating one or carries a reason — is not
  checked. Delete `"C"` from the map and the test still passes; add a corpus `"D"` to `golden.json`
  with no label and the test still passes.
- **Impact**: the guard proves labels have corpora, not that corpora have labels. AC13's stated
  purpose is *"a later reader must not mistake either for teeth"*, and the failure it must prevent is
  an unlabelled member arriving later.
- **Suggested Resolution**: invert the iteration — range over `golden`, and require every key other
  than the enumerated discriminating member(s) to have a non-empty entry in
  `nonDiscriminatingCorpusMembers`.
- **Related AC**: AC13

### Finding 8: `TestGoldenIsOrderInvariant`'s order clause cannot fail — it tests `big.Int` commutativity

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/bind/golden_test.go:181-190`, helper at `:107-158`
- **Observation**: `exactBigIntSum` computes `maxScale` (order-independent) and then accumulates
  `big.Int` terms, which is commutative by construction. `forward != backward` is therefore
  unfalsifiable for any input. Confirmed empirically: this test **passed unchanged** under the float64
  mutation that reddened everything else.
- **Impact**: AC11 and D-000.22 ask for a semantic property read off the artifact. The clause that is
  named as that property is a tautology. The substantive half of the same test is real and does carry
  the acceptance — `forward != member.Sum` is an independent re-derivation, not a restated hash, and I
  confirmed it is genuinely independent of `SumDecimals` — so the step is discharged in substance;
  what is wrong is the label on which half discharges it.
- **Suggested Resolution**: keep the assertion (it is a cheap forward guard on the helper) but stop
  presenting it as the acceptance property in `golden_test.go:160-172` and in `PROVENANCE.md:101-111`;
  name the independent re-derivation plus the scale check as the acceptance step, which is what they
  are. Alternatively give the order clause teeth by re-deriving through a path that could differ
  (e.g. aligning to `maxScale` per operand in reverse), though this is not obviously worth it.
- **Related AC**: AC11

### Finding 9: AC12's recorded `bc` run never resolved the recorded corpus for member A

- **Severity**: Minor
- **Category**: AC Conformance / Maintainability
- **Location**: `folio-go/testdata/decimal-reduction/PROVENANCE.md:72-76`
- **Observation**: corpus A is recorded in `golden.json` as **33 operands** (one balance plus 32
  literal `"0.01"` entries). The independent-reader invocation is
  `echo "scale=2; 12345678901234.56+0.01*32" | /usr/bin/bc` — a **multiplication**, not 32 additions
  — and the invocation labelled the reversed order, `0.01*32+12345678901234.56`, is a two-term
  commutation of that same pre-collapsed product, not a reversal of the 33-operand sequence. The B and
  C invocations do resolve their corpora operand-by-operand, in both orders.
- **Impact**: D-000.53's reader is meant to resolve *the recorded corpus*; here it resolved an
  algebraic re-expression of it, so for the one member that actually discriminates, `bc` never
  performed the accumulation the corpus records — which is the exact operation whose exactness is
  under test. The conclusion is nonetheless sound: I ran the true 33-term forward and reversed forms
  through `/usr/bin/bc` and both return `12345678901234.88`. The record understates what was verified.
- **Suggested Resolution**: replace the two corpus-A invocations with the literal 33-term sums (both
  orders) and record their verbatim output. The line is long; that is the point of the step.
- **Related AC**: AC12

### Finding 10: `AvgDecimals` can return a `Decimal` whose exponent breaches `maxDecimalExponentMagnitude`

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/internal/bind/reduce.go:183`
- **Observation**: the result exponent is `total.Exponent - avgExtraScale` with no bound check.
  Measured: averaging two operands at `Exponent: -maxDecimalExponentMagnitude` returns
  `Exponent: -100004` against the package's declared bound of `100000`, with `err == nil` — a value
  `NewDecimal` would reject as out of bounds.
- **Impact**: the kernel can mint a `Decimal` the package's own constructor considers invalid, so the
  bound is not an invariant of the type, only of one entry point. Small in magnitude (always exactly
  `avgExtraScale` past), but *Do-not-re-open* item 6's principle is that a bound breach is an Error,
  never a silent widening.
- **Suggested Resolution**: check the computed result exponent against `maxDecimalExponentMagnitude`
  and return the same shape of located error `SumDecimals` uses for its spread breach.
- **Related AC**: AC4, AC6

### Finding 11: the tripwire is described as forcing 3.3 to route through the kernel; it cannot detect that

- **Severity**: Minor
- **Category**: AC Conformance / Maintainability
- **Location**: story §*"The tripwire that makes Story 3.3 unable to skip the wiring"*; mirrored in
  `folio-go/internal/bind/reduce.go:8-12` and in D-3.1a.3
- **Observation**: the story and D-3.1a.3 both say the inventory *"cannot be satisfied by 3.3 writing a
  second accumulator inside `sum`"*. It can. An inline `big.Int` loop inside an expression-language
  `sum()` — whose own signature will be something like `func(args []Value) (Value, error)` — is not a
  `[]Decimal → (Decimal, error)` function, so the inventory still reports exactly
  `{SumDecimals, AvgDecimals}` and passes. Nothing anywhere asserts that any code **calls**
  `SumDecimals`/`AvgDecimals`; verified — the only non-test references to either name in `folio-go/`
  are inside `reduce.go` itself.
- **Impact**: the instrument enforces *"there are exactly two functions of this shape, co-located with
  `Decimal`"*, which is what AC23's own text says and what it delivers. The **stated force** — that
  3.3 cannot skip the wiring — is not delivered by it. Recording an over-claim about a guard is the
  D-000.24 pattern applied to a tripwire rather than to an error payload, and the risk is that 3.3's
  reviewer reads the story's sentence and stops looking.
- **Suggested Resolution**: this is the lead's to settle, not the developer's — either narrow the
  prose in the story and D-3.1a.3 to what the instrument checks, or add the missing clause (an
  assertion that at least one call site outside `internal/bind` references each reducer once 3.3
  lands, which is itself a D-000.30 window that shuts then).
- **Related AC**: AC23

### Finding 12: no negative operand appears anywhere — the kernel's whole sign path is untested

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/bind/reduce.go:158-172, 188-192`;
  `folio-go/internal/bind/golden_test.go:118-133`; all three corpora in `golden.json`
- **Observation**: every operand in every corpus and every unit test is non-negative. That leaves
  untested: `roundQuotientAwayFromZero`'s `q.Sign() < 0` branch, `AvgDecimals`' `Abs(remainder)`
  handling of `QuoRem`'s sign-of-dividend remainder, the negative half of the half-to-even tie, and
  `exactBigIntSum`'s own `neg` path in the acceptance helper. I read the negative tie through by hand
  (`-30000/32` → `q=-937, r=-16`, tie, odd → `-938`) and believe the code is correct; it is simply
  unexercised.
- **Impact**: a money kernel with no negative operand in its oracle is a gap a bank statement will
  find first. AC4's tie case is asserted "both directions" — but both directions of *even/odd*, not of
  *sign*.
- **Suggested Resolution**: add a signed corpus member (labelled as discriminating or not, per the
  same D-000.50 measurement the others got) or, at minimum, a negative tie case to
  `TestAvgDecimalsRoundsHalfToEven` and a mixed-sign case to `TestSumDecimalsAlignsToMinimumExponent`.
- **Related AC**: AC4, AC8

### Finding 13: the D-000.23 statement is not carried by the coverage WITNESS, only by findings

- **Severity**: Nit
- **Category**: Convention
- **Location**: `lint/internal/rules/bigfloattype.go:22-25` (the const), `:49-61`
  (`BigFloatTypeStats`)
- **Observation**: AC17 says *"The **coverage witness** and the rule's own documentation state, in
  those words, …"*. The witness is `BigFloatTypeStats`, which carries `DirsVisited`, `FilesParsed`
  and `TypedExprs` and no labelling. The statement lives in a separate exported constant and reaches a
  reader only through a finding's `Message` — so on the expected clean tree, where there are no
  findings, the labelling surfaces nowhere the witness is read.
- **Impact**: cosmetic today; the wording requirement is genuinely met by an asserted exported constant
  (`TestBigFloatTypeCoverageStatementWording`), which is the stronger instrument. Noted only because
  AC17 names the witness specifically.
- **Suggested Resolution**: add a `Coverage string` field to `BigFloatTypeStats` set from
  `BigFloatTypeCoverageStatement`, or amend AC17's wording to name the constant.
- **Related AC**: AC17

---

### AC-by-AC disposition

**Satisfied, verified:** AC1, AC2 (one narrowing site, `reduce.go:113`, reusing `IsInt64()`; no
checked-add helper exists), AC3, AC4 (both tie directions genuinely falsifiable; `avgExtraScale`
declared once and not derived from `golden.json`), AC5, AC7, AC8 (red-proved), AC9 (red-proved),
AC10 (independently reproduced, both values), AC16 (red-proved), AC18, AC19, AC20, AC21, AC22
(diffed), AC25 (byte-identical), AC26, AC27 (all four gates re-run by the reviewer).
**Satisfied with a note:** AC17 (Finding 13).
**Partially satisfied:** AC11 (Finding 8), AC12 (Finding 9), AC13 (Finding 7), AC23 (Findings 4, 11),
AC24 (Findings 5, 6), AC4/AC6 also touched by Findings 10 and 12.
**Not satisfied:** AC6 (Finding 3), AC14 (Finding 1), AC15 (Finding 2).

---

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-25
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 2
- **Majors:** 2
- **Minors:** 8
- **Nits:** 1

**Rationale.** The load-bearing work of this story is sound and, where it matters most, better than
its own record: corpus A really does discriminate under an honest float64 mutant, the mutant really is
order-dependent in the way D-000.61 predicts (the reversed order lands exactly on the golden — the
luck is reproducible), the AC16 vacuity witness really does zero out under a dead checker, the
reducer inventory really does redden on the live module in both directions and through the qualified-
import path, and the relational location clause really does survive a simulated move of `Decimal`.
Both existing guards are byte-untouched, `internal/expr` was not created, and all four gates
reproduce with the required failure byte-identical.

The two Blockers are the same defect in two shapes: **Layer 2 is green while blind**, which is the
condition this story was written to end. An aliased `big.Float` evades the type-identity check because
`*types.Alias` is not `*types.Named` in Go 1.23+ — a package that never mentions `math/big` can hold
values of that type and the scan reports nothing with a non-zero witness. And `_test.go` files, which
AC15 puts in scope and which the neighbouring guard explicitly refused to drop, are in scope for no
caller that ships. Both are small fixes — `types.Unalias`, and a second production caller — and
neither is a design disagreement: AC14 and AC15 already say what should happen.

The two Majors are a silent wrap in the money kernel's own bound check (`SumDecimals` returns
`{2, MinInt}` with a nil error for extreme exponents, defeating AC6 by integer overflow) and an
inventory that four ordinary Go spellings walk past — variadic most importantly, since a function
table's `sum(...)` is exactly where that spelling shows up.

The eight Minors divide into coverage gaps that a probe closed for me but no shipped test closes
(Findings 5, 6, 12), assertions that cannot fail in the direction their name promises (Findings 7, 8),
and records that claim slightly more than was executed (Findings 9, 11) plus one bound the kernel can
step past (Finding 10). None of them weakens a conclusion I checked; each of them weakens the guard
that is supposed to keep the conclusion true after this story stops being read.

---

## Finding Resolutions (finisher)

Triage instruction received: fix every finding; the review was assessed as strong and every finding as
real, with the sole DISMISS being the pre-verified `.font-sources` nit. Disposition below; every FIX was
independently reproduced (red) before its patch and re-verified (green) after, per D-000.30.

| # | Severity | Decision | Files touched |
|---|---|---|---|
| 1 | Blocker | FIX | `lint/internal/rules/bigfloattype.go` (`resolveNamedType`), `.../bigfloattype_test.go`, new fixture `folio-go/testdata/lint/no-bigfloat-type/violating_type_alias.go` |
| 2 | Blocker | FIX | `lint/internal/rules/bigfloattype.go` (go-build-cache filter), `.../bigfloattype_test.go` (`TestBigFloatTypeTestScopeInventory`) |
| 3 | Major | FIX | `folio-go/internal/bind/reduce.go` (`SumDecimals`), `reduce_test.go` |
| 4 | Major | FIX | `folio-go/internal/reducer_inventory_arch_test.go` |
| 5 | Minor | FIX | `folio-go/testdata/arch/reducer-inventory/moved-decimal/`, `internal/reducer_inventory_test.go` |
| 6 | Minor | FIX | `folio-go/internal/reducer_inventory_arch_test.go` (`decimalReducerViolations`), `internal/reducer_inventory_test.go` |
| 7 | Minor | FIX | `folio-go/internal/bind/golden_test.go` |
| 8 | Minor | FIX (documentation) | `folio-go/internal/bind/golden_test.go`, `testdata/decimal-reduction/PROVENANCE.md` |
| 9 | Minor | FIX | `folio-go/testdata/decimal-reduction/PROVENANCE.md` |
| 10 | Minor | FIX | `folio-go/internal/bind/reduce.go` (`AvgDecimals`), `reduce_test.go` |
| 11 | Minor | FIX | this story file, `folio-go/internal/bind/reduce.go`, decision log `D-3.1a.4` |
| 12 | Minor | FIX | `folio-go/internal/bind/reduce_test.go` |
| 13 | Nit | DISMISS | none — see rationale below |

**1 (Blocker) — alias evades the denylist.** FIX. `resolveNamedType` now applies `types.Unalias`
before and after the pointer unwrap. Red-proved twice: (a) live, against the shipped folio-go module
root and the retained fixture (a new `type Money = big.Float` fixture, `AliasedFloat`, added to
`no-bigfloat-type/` and asserted in `TestBigFloatTypeFixtureScan`'s exact-set comparison); (b) against
the reviewer's own reproduction — a two-package probe tree (`bfdep` declaring the alias, `app`
referencing it only as `bfdep.Money`) — measured 1 finding pre-fix (only the alias's own declaration
site, a genuine `*types.Named` reference, resolved; the three usage sites did not) and 4 findings
post-fix, confirmed by reverting the fix in place, re-running, and restoring from a `cp` backup with
`/usr/bin/diff` confirming byte-identical restore. The probe tree lived only under `/tmp` and a
throwaway `lint/cmd/` binary, both deleted; nothing from it is committed.

**2 (Blocker) — `_test.go` files unreachable.** FIX. Added `TestBigFloatTypeTestScopeInventory`, a
second production caller at `includeTests:true` mirroring `TestFloatTypedTestScopeInventory`'s
inventory shape (zero sanctioned sites today, since F1 measured zero occurrences repo-wide including
tests). Also fixed the go-build-cache leak into the coverage witness the reviewer's own probing
surfaced (`DirsVisited` entries like `../../../Library/Caches/go-build/04` at `includeTests:true`):
`ScanBigFloatTypes` now skips any file relativizing to a path outside `root`, by construction. Measured
before the fix: 0 findings at `Tests:false`, 4 at `Tests:true` (matching the reviewer's own numbers
exactly); after the fix, the production module root is 0/0 clean at both settings and the new test
passes with vacuity guards intact.

**3 (Major) — AC6 spread overflow.** FIX. `SumDecimals`' alignment spread is now computed in
`math/big` (exact, cannot wrap) rather than plain `int`. Red-proved with the reviewer's exact input,
`SumDecimals([]Decimal{{1, math.MinInt}, {1, math.MaxInt}})`: returned `{Coefficient: 2, Exponent:
math.MinInt}` with `err == nil` before the fix (confirmed live), now returns a located "alignment
spread" error. The existing `TestSumDecimalsAlignmentSpreadIsBoundedBeforeShifting` (a legitimately
huge but non-overflowing spread) continues to pass unmodified, since the fix changes only how the
spread is computed, not the bound or the message shape.

**4 (Major) — shape whitelist.** FIX. Widened the AST predicate to accept a variadic parameter
(`*ast.Ellipsis`), a package-level slice type ALIAS of Decimal (`type Decimals = []Decimal`, resolved
the same way a qualified `pkg.Decimal` reference is — never by source text, and restricted to genuine
Go aliases via `ts.Assign.IsValid()`, not type definitions), and a method receiver (the `fn.Recv != nil`
exclusion was removed). Red-proved live against a five-reducer probe tree (`SumDecimals`, `AvgDecimals`,
a variadic reducer, an alias-parameter reducer, a method reducer): the scan found 3 before widening the
predicate and 5 after, matching the reviewer's own reproduction. The two-parameter shape
(`TwoParamReducer`) the reviewer also measured is deliberately NOT matched — AC23's own wording is a
single-parameter reduction, and folding in "any function that happens to touch a `[]Decimal`" would
widen the property past what AC23 states, not merely past what the old predicate caught. The checker's
doc comment now states its one remaining stated gap (an alias behind a second layer of indirection, or
behind a generic type parameter) explicitly, per D-000.23.

**5 (Minor) — no fixture for the relocation failure mode.** FIX. Added a fourth fixture variant,
`moved-decimal/` (`Decimal` in `pkga`, both reducers left behind in `pkgb`, reached only via
`pkga.Decimal`), and a new red-proof test asserting the scan reports both reducers with
`PkgDir="pkgb"` while `DeclPkgDir="pkga"`, reddening AC23 via `decimalReducerViolations`.

**6 (Minor) — red-proofs assert on raw output, not the verdict.** FIX. Factored AC23's want-set/
location comparison into `decimalReducerViolations`, called by the production test and by all three
(now four, with Finding 5's fixture) red-proofs; each red-proof now additionally asserts the violation
list itself is non-empty, in addition to the pre-existing shape assertion.

**7 (Minor) — AC13 label test iterates its own map.** FIX. Inverted the iteration to range over
`golden.json` and require every member other than the one discriminating corpus to carry a non-empty
label, so deleting a label or adding an unlabelled corpus now fails the test (verified by construction
during development of the fix, reverted before commit).

**8 (Minor) — order-invariance clause is a `big.Int` tautology.** FIX (documentation, not a behaviour
change). The reviewer's own suggested alternative — giving the clause real teeth — was assessed and
declined per the reviewer's own "not obviously worth it" note: `TestGoldenIsOrderInvariant`'s
`forward != member.Sum` clause (an independent re-derivation against the recorded golden) and the scale
check already discharge D-000.22/AC11 substantively. Retitled the doc comments in `golden_test.go` and
`PROVENANCE.md` to name the re-derivation and scale check, not the order clause, as what discharges the
acceptance step; the order clause stays as a cheap forward guard on the helper, documented as such.

**9 (Minor) — AC12's `bc` invocation didn't resolve the recorded corpus.** FIX. Replaced the two
corpus-A `bc` invocations (which multiplied `0.01*32` rather than summing the 33 recorded literal
addends) with the literal 33-term forward and reversed sums, re-run and re-recorded verbatim in
`PROVENANCE.md`; both give `12345678901234.88`, matching the golden.

**10 (Minor) — `AvgDecimals` can mint an out-of-bound exponent.** FIX. Added a result-exponent bound
check (`maxDecimalExponentMagnitude`) before returning, mirroring `SumDecimals`' own discipline. Red-
proved with the reviewer's exact input (two operands at `-maxDecimalExponentMagnitude`): minted
`Exponent: -100004` with `err == nil` before the fix, now returns a located error.

**11 (Minor) — tripwire over-claims what it enforces.** FIX, scoped exactly as instructed: narrowed the
prose in this story file and in `reduce.go`'s doc comment to state that the inventory is a
declaration-shape set-equality-plus-location check, not a call-graph one — it cannot by itself force a
future `sum()`/`avg()` to route through the kernel. No call-graph enforcement was built. Filed as
append-only decision-log entry **D-3.1a.4** (D-3.1a.3 itself is unedited, per this log's append-only
rule), which also hands Story 3.3 the explicit follow-up: if routing enforcement is wanted, it needs a
call-site assertion, itself a fresh D-000.30 window that only 3.3 can open (Decimal/sum/avg do not
exist yet in the expression language).

**12 (Minor) — no negative-operand coverage.** FIX. Added `TestSumDecimalsMixedSignOperands` and two
negative-tie table cases to `TestAvgDecimalsRoundsHalfToEven`, all independently hand-computed
(cross-checked against the reviewer's own `-30000/32 → q=-937, r=-16, tie, odd → -938` figure) rather
than re-derived from the production code.

**13 (Nit) — coverage statement doesn't live on the witness struct.** DISMISS. The reviewer's own
Impact assessment calls this "cosmetic today" and confirms AC17's wording requirement is genuinely met
by the exported `BigFloatTypeCoverageStatement` constant, asserted directly by
`TestBigFloatTypeCoverageStatementWording` — a stronger instrument than a witness field would be, since
it does not depend on a finding existing to surface it. Adding a duplicate `Coverage string` field to
`BigFloatTypeStats` would be pure duplication with no new guarantee. No change made.
