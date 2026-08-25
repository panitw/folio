# Story 3.1a: Guard AD-23 against named binary-float types

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-1a-guard-ad-23-against-named-binary-float-types`
**Status:** `ready-for-dev`
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
can add up differently on two machines. Money is about to arrive — the next stories teach the engine
to total and average the numbers in a report — so the rule matters more than ever.

The rule is real. The two checks enforcing it are not quite. Both look for the *name* of the
approximate kind of number, and a second, differently-named kind sits in a library the engine already
uses — the same arithmetic wearing a different label, one word away from being typed by accident.
Both checks look straight past it. Worse, the obvious fix does not work: that second kind is built out
of whole numbers underneath, so a check that inspects how it is put together finds nothing wrong and
reports all clear.

So this story stops checking names and starts checking answers. It builds the exact totalling and
averaging the money path will rest on, and pins it to expected results worked out independently, away
from this code, and confirmed by two unrelated tools. A total even a fraction off fails. A separate,
deliberately narrow check bans the two named culprits outright, labelled as what it is: a list of two,
not a principle.

What it does not do: build the report functions themselves, or touch either existing check. Done looks
like a total that cannot drift, a ban that cannot be side-stepped by renaming an import, and a
tripwire forcing the next story to route its totals through this arithmetic rather than reinvent it.

Two things will look wrong later and are not. One long-standing test stays failing on purpose — it
belongs to an earlier story and measures a shortfall left deliberately visible. And some sample
figures in the test data are there precisely because they *cannot* catch the fault; they are labelled,
and they are evidence, not oversight.

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

1. [ ] **NO TYPE-SHAPE CHECK CAN CATCH `math/big.Float`.** This is D-3.1a.1's recorded correction of
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

2. [ ] **The guard lives in `lint`, and NOT ALSO in `folio-go/internal/arch_test.go`.**
   `arch_test.go` is a pure AST walk with no type information (`findFloatOccurrences:49` matches
   `*ast.Ident` by `.Name`), so extending it could only ever be a **name** match on `big.Float` —
   the exact proxy mechanism this story exists to stop. Putting it in both places is D-000.38's
   *"two guards sharing a parser are one guard wearing two names"*.

3. [ ] **Layer 3 is NOTHING.** Both existing guards stay exactly as they are, on their own axes.
   Do not edit `folio-go/internal/arch_test.go`. Do not edit
   `lint/internal/rules/floattyped.go`. This is a **scope prohibition, not an acceptance test** —
   and deliberately so: an AC asserting those files are byte-identical would pass vacuously
   whenever nobody edited them, which is every run in which the prohibition was already being
   obeyed. It has no teeth and it is not pretending to.

4. [ ] **Layer 2 is a DENYLIST and must be labelled as one (D-000.23).** It covers `math/big.Float`
   and `math/big.Rat` — **those two types, not the class.** Its coverage witness must say so **in
   those words**. It is **not** counted as coverage for AD-23's property; **Layer 1 is.** A denylist
   entry is never coverage.

5. [ ] **`big.Rat` is banned for a DIFFERENT reason from `big.Float`, and that is why Layer 1 alone
   is insufficient.** `big.Rat` is *exact*, so the behavioural oracle would **not** catch it. It is
   wrong because it carries an unrounded rational and thereby **dodges** AD-23's *"divides at a
   defined scale with round-half-to-even"* — and the scale is the **ruled** part, not an
   implementation detail. Do not "simplify" Layer 2 down to `big.Float` on the grounds that
   `big.Rat` is exact.

6. [ ] **Accumulator overflow is a located Error, never a widening.** Accumulate in `big.Int`
   (already blessed, already imported at `decimal.go:20`, exact) and narrow **once**, at the end,
   through the existing `IsInt64()` pattern at `decimal.go:119-128`. **Do not write a checked-add
   helper** — there is one narrowing site and it already exists.

7. [ ] **Do not create `folio-go/internal/expr/`.** `absence-expr-package` (`absences.go:96`)
   requires it absent until Story 3.2, and D-3.2.1 assigns that package's creation — and DW-8's
   `Decimal` move — to 3.2. Creating it here would fire a tripwire this story has no mandate to
   discharge. The kernel this story builds goes in `internal/bind`, beside `Decimal`, and **travels
   with `Decimal`** when 3.2 moves it.

8. [ ] **3.1a is NOT a D-000.4 matrix override.** The criterion is a *new source of cross-target
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
extinction, because the functions must exist somewhere. It **cannot be satisfied by 3.3 writing a
second accumulator inside `sum`**, and it **cannot be satisfied by duplicating**. If 3.3 needs a third
reducer, it amends the inventory in the same diff and a human reads it — the same deliberate friction
`assertExactFindingSites` already documents for the float-site register. Its red-proof is available
**now**, before 3.3 exists (AC23), which matters because D-000.30 shuts that window the moment the
obligation is wired.

---

## Acceptance Criteria

### Layer 1 — behavioural, PRIMARY, has teeth

- [ ] **AC1.** An exact decimal reduction kernel exists in `folio-go/internal/bind`, exporting
  `SumDecimals([]Decimal) (Decimal, error)` and `AvgDecimals([]Decimal) (Decimal, error)`. These are
  arithmetic kernels, **not** expression-language functions: no arity checking, no collection
  resolution, no diagnostic codes, no function-table entry.
- [ ] **AC2.** `SumDecimals` aligns operands to their common (minimum) exponent, accumulates in
  `big.Int`, and narrows **exactly once**, at the end, through the existing `IsInt64()` pattern of
  `decimal.go:119-128`. **No checked-add helper is written** (D-3.1a.1, verbatim). The single narrowing
  site is identified in a comment naming the pattern it reuses.
- [ ] **AC3.** `SumDecimals`' result carries the **maximum operand scale** (the minimum operand
  exponent), with trailing zeros preserved — D-1.6.1's AC2 distinctness (`"1.50"` ≠ `"1.5"`) survives
  addition. Asserted on a corpus member whose operands differ in scale.
- [ ] **AC4.** `AvgDecimals` divides at **maximum operand scale + 4**, **round-half-to-even**, declared
  **once** as a named constant that is a property of the operation and is **never derived from the
  data** and **never fitted to the shipped corpus** (D-000.32). A tie case is asserted, so the
  half-to-even half is exercised rather than assumed.
- [ ] **AC5.** Coefficient overflow at the single narrowing site is a **located Error** naming the
  operation, the operand count and the bound breached — **never a widening, never a wrap, never a
  truncation**. Red-proved with a corpus that overflows `int64` only after accumulation, so the error
  is reached through the accumulator rather than through `NewDecimal`'s existing per-literal bound.
- [ ] **AC6.** The **alignment step is bounded**: the spread between the maximum and minimum operand
  exponent is checked against `maxDecimalExponentMagnitude` (`decimal.go:65`) and a breach is an Error.
  **The check happens BEFORE any shift is computed** — computing a 200,000-digit shift in order to
  discover it was too large is itself the cost being defended against. Asserted by a test that would
  time out or allocate unboundedly if the check ran after the shift.
- [ ] **AC7.** An oracle golden is recorded under `folio-go/testdata/` carrying the operand corpus and
  the expected `sum` and `avg` for each corpus member, as **decimal literal strings**. Its provenance
  file records: `python3 --version` output (**3.12.13**), the verbatim producing invocation, the
  explicit `decimal.Context` used (precision and `ROUND_HALF_EVEN`), and the independent reader's
  invocation and output. **AD-25's one-time offline reference run, hand-checked — never a runtime
  dependency**, as D-2.3.1 applied it to shaping via `hb-shape`.
- [ ] **AC8.** A test compares the kernel's output against the recorded golden for **both** operations
  across **every** corpus member, asserting the `Decimal`'s coefficient **and** exponent — never a
  rendered string, which can agree while the value does not.
- [ ] **AC9.** The same test asserts **order-invariance**: `SumDecimals` over each corpus in declared
  order and in reverse order produces the identical `Decimal`. (F4: this is what stops a red-proof
  passing by luck.)
- [ ] **AC10. Red-proof, captured BEFORE the kernel is wired (D-000.30).** Replacing the accumulator
  with `float64` reddens AC8 and AC9 on corpus **A**, and the story records **both** measured values —
  `12345678901234.88` exact against `12345678901234.87` mutated for `sum`, and `374111481855.602424`
  against `374111481855.602230` for `avg`. The mutation must be **valid Go with forbidden semantics**
  (D-000.13) and must include the final rounding to the declared scale, or it manufactures a divergence
  the real implementation would have rounded away (F3).
- [ ] **AC11. D-000.22 semantic acceptance step**, this being a **first recording**: the recorded totals
  are asserted to be **order-invariant** (a property read off the artifact, which a float64-contaminated
  recording fails), and the recorded `sum` scale is asserted to equal the maximum operand scale of its
  own corpus. Not a restatement of a hash, and not derived from the producing invocation.
- [ ] **AC12. D-000.53 independent-reader step**: `/usr/bin/bc` — which this project did not write and
  which did not produce the golden — resolves each corpus into the total the golden claims. Reader,
  version and **verbatim invocation** recorded in the provenance file with its output. Python's
  `decimal` **may not** serve as its own acceptance reader.
- [ ] **AC13.** Corpus members **B** and **C** ship, and each is labelled **in the golden's provenance
  and in the test** as **non-discriminating**, with the reason: B exercises the alignment path that A
  cannot; C is the intuitive money corpus that **cannot express the defect at all** and is retained as
  D-000.50 evidence. A later reader must not mistake either for teeth.

### Layer 2 — a type-identity DENYLIST, narrow, and labelled as one

- [ ] **AC14.** A new `lint` rule forbids `math/big.Float` and `math/big.Rat` by **resolved type
  identity** — `Obj().Pkg().Path()` + `Obj().Name()` — and **never by source text**. An alias, a
  dot-import, a renamed import, a type parameter instantiated at one of them, and a variable, field,
  parameter or result of that type all resolve the same and all trip it.
- [ ] **AC15.** Scope is the **folio-go module root** (which strictly contains `folio-go/internal/`),
  matching `ScanFloatTypedValues`' shipped production caller exactly — see F5 and Flag F1. `_test.go`
  files are **in scope**, matching AD-23's existing file scope; `testdata/` subtrees are excluded, so
  this rule's own violating fixture does not trip the production scan.
- [ ] **AC16.** The rule returns a **coverage witness** reporting what it actually examined — files
  parsed, directories visited, and **expressions whose type resolved** — taken from the checker's **own
  execution**, never a second independent walk. Injecting `if true { return … }` as the checker's first
  statement must zero it. (`FloatTypedStats`' `TypedExprs` is the precedent and the reason: a checker
  that obtains no type information reports zero findings exactly as a clean tree does.)
- [ ] **AC17.** The coverage witness and the rule's own documentation state, **in those words**, that
  this rule covers **those two types, not the class**, and that it is **not counted as coverage for
  AD-23's property — Layer 1 is** (D-000.23, D-3.1a.1).
- [ ] **AC18. Red-proof (D-000.13).** A retained fixture under `folio-go/testdata/lint/` contains a
  **real `big.Float` value expression** — **valid syntax, forbidden semantics** — alongside a compliant
  sibling. The test asserts on **rule id and message**, **never on exit status or mere failure**. The
  message names the **resolved type identity** and the expression's file and line. At least one fixture
  case reaches the rule through a **renamed import** and one through a **variable of the type**, so
  AC14's "never by source text" is exercised rather than claimed.
- [ ] **AC19.** The fixture scan reports **exactly** the violating file and **not** the compliant one,
  compared by **(file, rule)** — never by count (D-1.3.3 amended).
- [ ] **AC20.** The rule **fails loudly** on a tree it cannot type-check — both the `packages.Load`
  error path and the per-package `Errors` sweep (D-1.3.11) — returning an error and **no findings**,
  never `(nil, nil)`. Red-proved against the existing `float-typed-untypecheckable` fixture shape.
- [ ] **AC21.** A production caller points the rule at the real folio-go module root and asserts
  **zero findings**, guarded against vacuity by the AC16 witness. This is expected to pass on the first
  run — F1 measured zero sites — and the AC16 witness is what makes that green mean something.

### Layer 3 — nothing

- [ ] **AC22.** `folio-go/internal/arch_test.go` and `lint/internal/rules/floattyped.go` are
  **not modified by this story**. Verified by reading the diff, not by a test (see *Do not re-open*
  item 3 for why an assertion here would be vacuous).

### The tripwire that makes Story 3.3 unable to skip the wiring

- [ ] **AC23.** An assertion, counted **by AST over the whole `folio-go` module** (D-000.14 — never
  text, never a filtered pipe), states that the set of functions reducing a sequence of `Decimal` to
  `(Decimal, error)` is **exactly** `{SumDecimals, AvgDecimals}` and that they are declared in **one**
  package. Its failure message names the offending declaration and says plainly that a second reducer
  must either route through the kernel or be added to this inventory deliberately.
- [ ] **AC24. Its red-proof, captured NOW, before 3.3 exists** (D-000.30 — wiring shuts the window
  permanently): adding a third `[]Decimal → (Decimal, error)` function anywhere in the module reddens
  AC23, and removing one of the two reddens it too. Both directions demonstrated.

### Guardrails

- [ ] **AC25.** `TestCorpusMeetsP6ExerciseFloors` **stays red** with stats byte-identical to
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. **Never "fix" it.**
- [ ] **AC26.** `folio-go/internal/expr/` is **not created**; `absence-expr-package` still passes.
  No absence rule is touched by this story, so D-000.59's replacement obligation is not triggered here.
- [ ] **AC27.** Per-story gate cadence: `go test`, `go vet`, `go build` and `lint` all run and are
  reported **with their scope and flags** (D-000.26), against the corrected baseline in the header —
  including the `.font-sources/` caveat for `lint`. The **cross-target matrix is NOT run** (item 8).

---

## Task breakdown

1. [ ] **Re-verify the baseline** at `b227dda` with a clean tree, including the `lint` measurement in a
   detached worktree so `.font-sources/` cannot contaminate it. Do not carry the header's numbers
   forward without re-measuring.
2. [ ] **Re-verify item 1's premise** against the local Go toolchain's `math/big` source before writing
   anything. If `Float`'s definition has changed, stop and report.
3. [ ] **Capture AC24's red-proof first** — the reducer inventory's two failure directions — before the
   kernel exists, then AC10's float64 mutant red-proof before the kernel is wired to the golden.
4. [ ] **Build the kernel** (AC1–AC6) in `internal/bind`, reusing `decimal.go`'s existing bounds and
   narrowing pattern. Resist adding a checked-add helper.
5. [ ] **Produce and record the oracle golden** (AC7), then discharge the acceptance steps: AC11
   semantic, AC12 independent reader via `bc`. Record provenance verbatim.
6. [ ] **Wire the Layer-1 oracle test** (AC8, AC9, AC13) and confirm AC10's captured mutant still
   reddens it.
7. [ ] **Build the Layer-2 rule** (AC14–AC21) in `lint/internal/rules`, following `floattyped.go`'s
   shape for stats, loud failure and fixture scanning.
8. [ ] **Land the reducer inventory** (AC23) and confirm the step-3 red-proof still applies.
9. [ ] **Run the per-story gate** (AC27), confirm AC25 and AC26, and record every figure with its scope
   and flags.
10. [ ] **Fill in the Delivery Log**, including the D-000.60 line naming what each flag below
    suppressed and whether the suppression was lifted. Move status to `review`.

---

## Flags — and what each one SUPPRESSES (D-000.60)

D-000.60 binds: a flag must name what it holds back, so whoever applies the ruling can lift it.

- **F1 — Layer 2's scope widened from `folio-go/internal/` to the folio-go module root.**
  *Grounds:* F5 — the charter's phrasing is strictly narrower than the neighbouring guard's shipped
  scope and would leave the public entry point uncovered.
  *What this SUPPRESSES:* nothing is withheld — this is a widening, and it is flagged so the lead can
  **narrow** it if the charter's phrasing was deliberate. If narrowed, `folio-go/render.go`,
  `render_entry.go` and `page_number.go` fall outside the rule and a follow-up owner is needed for
  them.

- **F2 — the kernel's error payload carries no AD-14 diagnostic code, data path or element id.**
  *Grounds:* `internal/diag` does not exist until 3.6 (`absence-diag-package`), and the kernel receives
  a `[]Decimal` — it **cannot know** the element id or data path. The kernel returns a structured error
  naming the operand index and the bound breached; the caller attaches the located payload.
  *What this SUPPRESSES:* **a test asserting AD-14's full payload on an overflow is NOT written in this
  story.** When 3.6 mints the codes, that test becomes writable and is owed. AC5 asserts the error is
  located *as far as the kernel can locate it* and no further, and it must not be read as discharging
  AD-14.

- **F3 — `avg`'s `+4` is carried as declared, at the charter's own MEDIUM confidence.**
  *Grounds:* D-3.1a.1 marks the `+4` illustrative and medium-confidence; its forcing function is Story
  3.4's assertion that no `formatNumber` pattern can request more fractional digits than `avg`
  produces.
  *What this SUPPRESSES:* **no test in this story attempts to justify the value 4.** AC4 asserts only
  that the scale is declared once, is a property of the operation, and is not fitted to the corpus. If
  3.4 finds a pattern that can request more, the constant moves — and that is the design working, not a
  regression here.

- **F4 — `ResolveAssets` walking the gitignored `.font-sources/` is noted, not fixed.**
  *Grounds:* out of this story's scope; it is a local-environment artifact producing three false reds.
  *What this SUPPRESSES:* **no fix and no test.** The three `lint` failures stay reproducible in this
  checkout. If the lead wants it owned, it needs a story; do not silently absorb it here.

- **F5 — empty-input semantics for both kernel functions.**
  *Ruled for the kernel:* `SumDecimals(nil)` and `AvgDecimals(nil)` return an **error**. An empty sum
  has no operands and therefore no defined scale, and an empty average divides by zero.
  *What this SUPPRESSES:* **this is NOT a ruling on what the expression-language `sum()`/`avg()` do on
  an empty collection** — that is 3.3's semantic question and may legitimately resolve to a zero, an
  Error, or a hidden component. No test in this story presupposes an answer. When 3.3 rules it, whoever
  applies the ruling should check whether the kernel's error is still the right primitive.

---

## Delivery Log

*(To be completed by the developer, reviewer and finisher.)*

- **Baseline re-verified at:**
- **AC24 red-proof captured at (pre-kernel):**
- **AC10 red-proof captured at (pre-wiring), both measured values:**
- **Golden provenance — producer, reader, invocations:**
- **Gate results, with scope and flags:**
- **D-000.60 line — what each flag suppressed, and whether it was lifted:**
