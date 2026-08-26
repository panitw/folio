---
baseline_commit: b87558298cfc121543983e9d7a0ad0c7a685be49
---

# Story 3.3: Aggregate over a collection

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-3-aggregate-over-a-collection`
**Status:** `done`
**Covers:** **FR19** · **AD-11**, **AD-23** · touches **FR41**
**Primary invariant:** **AD-23** (report data numbers are exact decimals; all arithmetic in `sum`,
`avg` and comparison is exact decimal arithmetic; `float64` appears nowhere under `internal/`;
`avg` divides at a defined scale with round-half-to-even — ARCHITECTURE-SPINE.md:412).
**Co-primary invariant:** **AD-11** (aggregates always take a **root-relative collection path**, are
legal inside a row scope, and are **always computed over the whole collection — never over the rows
on the current page**; per-page subtotals are not in MVP and must not be introduced as a special
case of an aggregate — ARCHITECTURE-SPINE.md:253).
**Adjacent invariants:** AD-1 · AD-3 · **AD-4** (no `page` namespace in the expression language,
ever) · AD-9 · AD-13 · **AD-14** (one diagnostic type, closed code registry, never a panic).
**Governing rulings:** **D-3.1a.2** · **D-3.1a.1** + correction + **D-3.1a.3** + **D-3.1a.4** ·
**D-3.2.1** · **D-3.2.2** · **D-3.1.1** · **D-1.4.1** · **D-1.4.2** · **D-2.8.1** · **D-2.8.3** ·
**D-2.8.6** · **D-000.59** · **D-000.61** + **extension** · D-000.9 · D-000.17 · D-000.24 ·
D-000.26 · D-000.29 · D-000.30 · D-000.38 · D-000.45 · D-000.47 · D-000.49 · D-000.50 · D-000.57 ·
D-000.60 · D-1.6.1 · D-1.7.4 · D-2.1.14.

**Rulings issued against this story during creation, all folded in below:** the **collection seam**
(two methods, guard re-pointed — R1, R2) · the **per-element defect semantics** (R8, AC19) · the
**OWNER DECISION that an explicit `null` is a ZERO OBSERVATION**, resolving to the additive identity
(R7) · **`avg`'s divisor is `len(projected)`, forced by the reducer inventory's signature** (R10,
AC4) · the **four-state discrimination before the kernel** (R8) · the **`avg`-on-empty Warning**
(R9) · **AD-11 bound now**, with AC22's instrument replaced by set equality.

**The single most load-bearing new assertion is AC4's length invariant**, and its only *independent*
red-proof is **AC12a — the mutant named "implement option 1"**, the arm the owner rejected. `sum`
cannot see that mutation at all; only `avg` can.

**Retires no DW entry. DW-7 stays owned by Story 4.5 — but this story APPENDS to it** (see
*Cross-story obligation*): 3.3 lands the one implementation, and DW-7's *"anti-rot mechanism: none
possible"* becomes false for one of its two halves.

**One NEW escalation is open: `DECISION-5` (the Warning has no channel).** It is the only thing in
this file the developer must not start without. Everything else is ruled.

---

## Baseline, measured in this run at creation

HEAD **`b875582`** — *"docs: correct 'diagnostic' to 'error' in the expression reference"* — branch
`main`. Working tree carries one pre-existing modification, **`folio-mvp-decision-log.md` only**
(present before this run began, not produced by it); no `.go` file and no fixture is modified.

**The gate is two modules.** Both re-measured in this run, independently of the figures supplied
with the rulings — they agree exactly (D-000.26: scope and flags, or it is not a figure):

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 go test ./... -count=1 -v` | **740 pass · 1 fail · 1 skip**, 15 packages |
| `lint/` | `CGO_ENABLED=0 go test ./... -count=1` | **89 pass · 3 fail**, 4 packages |
| `folio-go/internal/expr`, `folio-go/internal/bind` | `go test ./internal/expr/... ./internal/bind/...` | ok · ok |

**The one `folio-go` FAIL is `TestCorpusMeetsP6ExerciseFloors`** — **REQUIRED red**, stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, "P6g (opaque names) floor not met: got 7,
need >=20". **D-000.17** ("a floor that is not met is reported as unmet — it is never filled"),
**D-2.1.14**, **D-000.57**. Nothing to do with this story. Never "fix" it.

**The three `lint` FAILs are one root cause — DW-19, and it is LOCAL-ONLY.**
`TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`, `TestFontsAssetsNoticeRemovalRedProof`,
all reporting *".font-sources: contains a committed font binary but no LICENSE…"*. `.font-sources/`
is **gitignored at `.gitignore:85`** and holds three variable-font `.ttf` with no `LICENSE`; **CI
never sees the directory and is green.** Any `lint` figure this story reports must name these three
or it is not a figure. **Nobody "fixes" DW-19 by weakening the rule.**

**Cadence (D-000.4):** unit / vet / build / lint every story, **both modules**. The cross-target hash
matrix is due at Epic 3's close, not here.

---

## In plain terms (read this first if you just want the gist)

A bank statement has to end with a total that adds up. The engine could already fetch one value at a
time from a customer's data, but not add a column of amounts, count them, or average them. This
story teaches it those three operations, on money terms.

Money needs care because computers ordinarily store fractional numbers in a form that cannot hold
one satang exactly, so thousands of additions drift. The engine instead keeps every amount as its own
written digits and adds those digits, exact whatever order the rows arrive in. A total always covers
the whole collection: there is no per-page subtotal, because the language cannot refer to the current
page.

Four situations that look alike are now told apart. Two are somebody's mistake and stop the document:
a list the data does not contain, and something present but not actually a list. The other two are
not mistakes: a list that exists and is empty, and a list (or an entry inside one) that is blank
rather than missing. A blank entry counts as one observation of zero, pulling an average down as a
real zero would, rather than being skipped. A column of blanks and a column of real zeros give
identical answers and cannot be told apart — that is the choice, not an oversight. A blank entry and
a genuinely missing one stay different: only the missing one stops the document.

An empty list and an all-blank list look inconsistent on purpose: an empty total prints as a plain
zero, while an all-blank average carries extra decimal places, and tidying either is a later story's
job. An empty list also cannot produce an average, since there is nothing to average, so instead of
failing the document it prints a quiet note explaining why that figure is missing, wherever it was
asked for, including a running total repeated on every page.

Not here: tables and pagination, the next epic, and turning a total into text like "1,234.56", the
following story. One test is expected to keep failing on purpose — a coverage marker left unmet, not
ours to close.

---

## Story

**As a** template author,
**I want** a statement total that adds up,
**So that** the sum footer on my transaction table is correct to the last satang.

---

## Do not re-open — settled rulings this story inherits or receives

1. **No third-party expression library. No CEL.** **D-3.2.2**, owner decision, on the **numeric
   model**: CEL has no decimal type and cannot express division at a declared scale with
   round-half-to-even. The "register a custom decimal type" steelman is already rejected there.
2. **No binary floating point.** `float32`, `float64` and — per **D-3.1a.1 (corrected)** —
   `math/big.Float`, banned by a two-layer guard scanning the **`folio-go` module root** by
   **resolved type identity**. A type-shape check is not an acceptable substitute.
3. **`SumDecimals([])` returns the identity; `AvgDecimals([])` errors.** **D-3.1a.2**. **The kernel
   is UNCHANGED by this story.** The asymmetry is the content.
4. **Two adjacent zeros are differently shaped, and both are correct.** `SumDecimals([])` returns
   exponent `0`, so an **empty-collection `sum` renders `0`, not `0.00`**; an **all-null `avg`
   renders `0.0000`** (maximum operand scale `0` plus `avgExtraScale`, R7.1). Presentation scale is
   Story 3.4's `formatNumber` / `footerFormat` job. **A developer who "corrects" either exponent, or
   who "harmonises" the two, has re-decided the scale at a call site — the precise failure D-3.1a.2
   exists to prevent.**
5. **`Decimal` has moved and is not moving again.** **D-3.2.1**/DW-8 done.
   `TestExactlyOneDecimalDeclarationInTheModule` and `reducer_inventory_arch_test.go` stay green
   **with zero edits**.
6. **The function table is closed at eight.** Three entries' `implemented` flag changes. No entry is
   added.
7. **The row alias may never become a `rootName`; root precedence is `params` → row alias → data.**
   **D-3.1.1**, unchanged and unchangeable by this story. Its guardrail list applies verbatim:
   `parseBindingPath`/`isValidIdent` stay **dead** (the extinction guard is not weakened to make
   room), no second path matcher, no alias in `reservedPlaceholders`, no non-literal `rootName`.
8. **No `page` namespace, ever.** **AD-4**. Not as a token, a root, an aggregate parameter, or a
   footer special case.
9. **`internal/expr` may not import `internal/bind`.** The two new methods are **declared** on
   `expr.Resolver` and **implemented** in `bind`.
10. **A float red-proof must vary operand ORDER, and its mutant must be HONEST.** **D-000.61** and
    **D-000.61 (extension)**. Both halves are written into AC9/AC10.
11. **An explicit `null` element is a ZERO OBSERVATION.** OWNER DECISION (below). Any AC asserting
    an error on a null element is removed. **See the dissolved fence in Flags.**

---

## Findings — measured in this run at `b875582`

### F1. THE COLLECTION SEAM DOES NOT EXIST. This is the story's real work, not three flag flips.

`expr.Value` (`internal/expr/ast.go`) is scalar **by construction** — its own doc comment says *"a
scalar, deliberately."* `expr.Resolver` has one scalar-returning method. `bind.Value.Lookup` *"walks
nested objects only"* and returns `Absent` the moment it steps into an array.

**Measured, by probe, on `{"transactions":[{"amount":"1"},{"amount":2}],"empty":[]}`:**

| probe | result |
|---|---|
| `Value.Lookup(["transactions","amount"])` | `Kind=null`, `Presence=Absent` |
| `Value.Lookup(["transactions"])` | `Kind=array`, `Presence=Present`, `len=2` |
| `exprResolver.Resolve(["transactions"])` | `data path "transactions" is a array, not a scalar value usable in an expression` |
| `exprResolver.Resolve(["transactions","amount"])` | `data path "transactions.amount" is absent from the report data` |
| `bind.Resolve("{{sum(transactions.amount)}}", …)` | `function "sum" is not yet implemented (coming in Story 3.3)` |

Row two is the sharp one: **the collection is reachable today and is rejected as "not a scalar value
usable in an expression."** *(Independently confirmed by the lead.)*

### F2. Flipping the three table entries is the small half, and it is already scaffolded

`evalCall` (`internal/expr/eval.go`) returns the located *"not yet implemented (coming in Story %s)"*
error **before evaluating any argument**, and its `switch entry.name` has an explicitly-unreachable
`default` arm that becomes reachable the moment a flag flips without a matching evaluator.

### F3. A NUMBER BOUND INTO TEXT IS AN ERROR — so `{{sum(...)}}` alone still does not print

`bind.Resolve` renders only `KindNull` (empty) and `KindString`; everything else hits *"resolved to a
%s, not a string — text bindings are never coerced."* Pinned deliberately at
`internal/bind/text_test.go:58` and `:287` (the second added by review specifically so *"a
well-formed number… must not become a backdoor"*).

**This story's observable "done" is at the evaluation layer, not at rendered text.** A total reaches
text only through Story 3.4's `formatNumber`.

**RECORD/CODE DIVERGENCE, confirmed and assigned here.** `docs/expression-reference.md:117-131`
presents all three aggregates as bare text bindings, as if they render. **The page becomes wrong in
a NEW way because of this story**, so the fix belongs here, and 3.3's edit and 3.4's are **different
edits**: 3.3 states *an aggregate must be formatted before it can appear in text; a bare
`{{sum(...)}}` is an Error*; 3.4 adds the worked `formatNumber` example. Deferring merges them and
loses the first. **Both sites** (`.md` and the `.html` twin) per D-000.47.

### F4. Accumulator overflow is already loud and located — the gaps are `%w` and the element id

Sharp edge investigated. `SumDecimals` (`internal/expr/reduce.go`) computes the alignment spread **in
`math/big`** and rejects a spread past `maxDecimalExponentMagnitude` **before any shift is
computed** (with a recorded correction noting the plain-`int` subtraction used to *wrap*);
accumulates in `big.Int` with no intermediate rounding; narrows **exactly once** through
`total.IsInt64()`. `AvgDecimals` additionally bounds its result exponent. The governing principle is
quoted in that file: *"a bound breach is a located Error, never a silent widening."*

**So AD-23's silence on accumulator overflow leaves nothing undetermined** — the required behaviour
is a loud, located failure, and the kernel produces one. **Two gaps this story must close:**

1. **Those errors are bare `fmt.Errorf` — no AD-14 code, no sentinel, no wrapped type.** They must be
   wrapped with **`%w`**, never re-formatted into a new message, or Story 3.6 must either re-plumb
   every site or match on message text — which AD-14 forbids callers from doing.
2. The element id and the collection path must be attached at the expression layer — `reduce.go`'s
   own F2 note says *"the eventual caller in `internal/expr` attaches the full AD-14 payload."*
   This story is that caller.

The trigger is realistic, not theoretical: **alignment moves every operand to the minimum exponent**,
so one high-precision value (`0.000000000001`) beside ordinary amounts multiplies every other
coefficient by 10^10 before the add. Overflow on a money path is reachable from caller data.

### F5. `avg`'s scale IS declared, and this story must not re-declare it

`avgExtraScale = 4` (`internal/expr/reduce.go`), declared **once**, *"as a property of the operation,
and never derived from a corpus… or from any caller's data."* `AvgDecimals` divides at **(maximum
operand scale + 4)** with round-half-to-even implemented by comparing `2·|remainder|` against
`|divisor|`. The value `4` carries a standing **MEDIUM-confidence** flag whose honesty-keeper is
**Story 3.4**. That flag carries forward unchanged (FLAG-1).

### F6. THE REDUCER-INVENTORY TRIPWIRE DOES NOT FORCE ROUTING — the two-accumulator hazard is live

**D-3.1a.4** corrected 3.1a's own over-claim and `reduce.go` repeats it: the inventory is *"a
declaration-shape set-equality check, not a call-graph one: it does NOT assert anything calls
`SumDecimals`/`AvgDecimals`."*

**A developer can implement `sum()` with an inline `big.Int` loop and every guard in the module stays
green.** *(Independently confirmed by the lead.)* The routing assertion is the only thing that stops
the second accumulator, it discharges **D-3.1a.4's follow-up owed to 3.3 by name**, and it is what
**DW-7 holds Story 4.5 to**. **Its red-proof must be captured BEFORE `sum` is wired** (D-000.30 —
the window shuts the moment 3.3 lands).

### F7. AD-11's "root-relative" IS the mechanism by which "never per-page" is achieved

Root dispatch selects the **row** root on first-segment equality with the declared alias, after
`params` and before `data`. An aggregate operand must **bypass** that and resolve from the data root
even when the alias would match. Testable today, with no table and no pagination.

### F8. The resolution-roots guard is keyed on a PROXY, and the proxy has a recorded blind spot

`TestBindResolutionRootsAreClosed` scans for `lookupBound` call sites and reads argument index 4.
Its own comments — and **D-3.1.1**, quoting Story 2.5's review Blocker 2 — record that it *"can be
defeated by an early-return dispatch that never calls `lookupBound` at all."* A new collection
dispatch is exactly that shape. **Ruled: re-point the guard, do not duplicate it** (AC7).

### F9. `count(collection)` — no projection — is the documented form

`docs/expression-reference.md:117` heads the section **`sum(collection.field)` ·
`count(collection)` · `avg(collection.field)`**. Corroborates the two-method seam and R5.

### F10. Pagination of collections does not exist yet

Nothing renders a table until Epic 4 (`4-1`…`4-5`, all `backlog`). **Story 4.2's own AC** —
*"Given a table bound to an empty collection … the header renders, no data rows are produced, and the
render succeeds"* — is verified present in `epics.md` and is what forces `avg`-on-empty to be a
Warning rather than an Error.

### F11. `TestThreeImplementedFunctions` / `TestFiveUnimplementedFunctions` are HARD-CODED NAME LISTS

Verified verbatim in `internal/expr/table_test.go:34` and `:43`: `[]string{"upper","lower","if"}`
and `[]string{"sum","count","avg","formatDate","formatNumber"}`. **This story edits both, and Story
3.4 edits both again** — exactly D-3.1a.3's *"a guard whose expected value must be edited is one that
gets edited wrongly."* Restate derivationally (AC29).

### F12. Four stale comments — every one names a package that no longer declares the symbol. Verified.

| site | says | truth |
|---|---|---|
| `internal/reducer_inventory_arch_test.go:5` | *"internal/bind's SumDecimals/AvgDecimals"* | they live in `internal/expr` (D-3.2.1) |
| `internal/template/decimal.go:~81` | *"internal/bind.NewDecimal"* | `internal/expr.NewDecimal` |
| `internal/template/decimal.go:~121` | *"internal/bind.Decimal's own maxDecimalExponentMagnitude"* | `internal/expr` |
| `internal/template/decimal.go:~136` | *"Set to exactly internal/bind.maxDecimalExponentMagnitude"* | `internal/expr` |

**Plus three wrong line cites in `internal/expr/reduce.go:69-77`**, all re-measured in this run:
`decimal.go:20` → **`math/big` is imported at :12**; `decimal.go:119-128` → **the `IsInt64()`
narrowing is at :111-122**; `decimal.go:65` → **`maxDecimalExponentMagnitude` is declared at :57**.
**Prefer symbol references over line cites** when fixing these — a line cite is a stale comment
waiting to happen, and this is the second time the project has caught one.

### F13. THE EXPRESSION LAYER HAS NO DIAGNOSTIC CHANNEL AT ALL — this is `DECISION-5`

The ruling says `avg`-on-empty emits *"a `Warning` on the existing `Result` channel."* **Measured:
that channel does not reach the expression layer, in either direction.**

- `Diagnostic`, `Severity` and `DiagCodeTextClippedWidth` are declared in the **module-root `folio`
  package** (`folio-go/diagnostic.go`), and the only Warning that ships is **minted in `render.go`**
  (`:611`), inside `folio` itself.
- `expr.Eval` returns `(Value, error)`. `bind.Resolve` returns `(string, []Substitution, error)`.
  **Neither carries a diagnostic, and neither can import `folio`** — that is backwards in the rank
  table.
- `internal/diag` does not exist; **Story 3.6 owns it** and **D-1.4.2** forbids minting codes early.

So the Warning is ruled but **has no wire.** The developer cannot start this AC without a shape.
Recommendation and precedent in `DECISION-5` below.

---

## Rulings this story applies

### R1 — The collection seam is TWO methods on `expr.Resolver`, implemented in `bind`

```go
CollectionLength(path []string) (int, error)      // count
ProjectCollection(path []string) ([]Value, error) // sum, avg — exactly one Value per element, in data order
```

**Binding:** the two-operation split; the one-`Value`-per-element-in-data-order contract; `count`
being **structurally unable** to reach a projected value. **Illustrative:** the two spellings —
improve them if there is a better name, but record why in the delivery log.

**Why two, and it is forced by shipped code.** DW-5's discharge already draws this line: a column
requesting a `sum`/`avg` footer with `footerOf` omitted is a load error; a `count` footer is not.
The footer layer already says *sum and avg need a numeric source; count needs only the collection.*
A single `ResolveCollection` serving both would have to return `[]Value` for a bare collection path —
**putting the array kind back through the side door that keeping `expr.Value` scalar exists to
close.** Splitting makes `count`'s independence a **property of the seam**, not a discipline.

**Neither method may take a range, offset, index or limit.**

### R2 — Root selection is EXTRACTED and the guard is RE-POINTED, never duplicated

Do **not** land a second guard and amend the first with a coverage disclaimer: that trades a known
blind spot for advertised-coverage-that-does-not-exist (**D-000.38**) and leaves D-3.1.1's recorded
weakness permanently open while looking closed.

Extract root selection out of `lookupBound` into **one shared helper** that both the scalar dispatch
and the two collection methods call, and re-point `TestBindResolutionRootsAreClosed` at that helper.
Still **one** closed-set guard; its key moves from the **proxy** (`lookupBound` call sites) to the
**purpose** (the one place a resolution root is chosen).

**Fallback, if root selection proves inseparable from `lookupBound` without changing a signature
D-3.1.1 froze:** the two-guard remedy **with the uncovered surface named in the assertion's own
text** as D-000.24's labelled category — never a silent narrowing — and it **returns to the lead as
a finding, not a workaround.**

### R3 — The aggregate operand bypasses the row root and resolves root-relative

Verbatim AD-11, and it is the mechanism: the only thing that could narrow a collection is the scope
it resolves through, and it resolves through the document root.

### R4 — `sum` and `avg` route through `SumDecimals`/`AvgDecimals`; the kernel is unchanged

Forced by **D-1.4.1**/**DW-7** (one implementation, never two that drift) and **D-3.1a.2**. **Not**
forced by the reducer inventory (F6) — hence AC8's positive routing assertion in **D-000.59**'s
shape: assert the **obligation**, never the **event**.

### R5 — `count` counts the collection and never inspects a projected field

**Ratified as FORCED, not chosen.** The sentence that makes it read as a consequence:

> **`count` is a property of the collection; `sum` and `avg` are properties of a projection over it.**

This is the **third** independent place they separate for that one reason — **D-3.1a.2** (sum has an
identity, avg does not, count is neither question), **DW-5's discharge**, and now **the seam**.
`count` succeeding on a collection where `sum` fails is a consequence of that, not an inconsistency.

### R6 — Kernel errors are WRAPPED with `%w`, never replaced

R6 discharges `reduce.go`'s own F2 note. `%w`, plus element id and collection path. Never a
re-formatted message (F4).

### R7 — OWNER DECISION: an explicit `null` element is a ZERO OBSERVATION

A `null` element contributes **`0` to `sum`** and **counts 1 in `avg`'s divisor**.
**`avg([1, null, 3])` = 1.33. `count` = 3.**

Consistent with AD-14's null clause and with the owner's Story 3.2 ruling that `if(null)` is silently
false. **Four consequences that shape ACs:**

1. **A null element resolves to the additive identity `{Coefficient: 0, Exponent: 0}`** — the *same*
   value **D-3.1a.2** already rules `SumDecimals([])` returns. **Not** a zero carrying the column's
   scale; **not** a zero carrying the projection's maximum scale. Stated explicitly because the
   consequence **looks like a bug and is not**: an **all-null `avg` renders `0.0000`** (maximum
   operand scale `0`, plus `avgExtraScale`) while an **empty-collection `sum` renders `0`**. Two
   adjacent cases, two differently-shaped zeros, both correct. Both are **presentation's** problem
   at Story 3.4 / DW-5's discharge — the same posture as the exponent-`0` note in *Do not re-open*
   item 4. **A developer who "harmonises" the two zeros has re-decided the scale at a call site.**
2. **`[null, null, null]` and `[0, 0, 0]` are INDISTINGUISHABLE** in both `sum` and `avg`, and
   `count` does not disambiguate them either. **That is what the owner's decision means — it is the
   decision, not a defect.** Written down here so no test asserts against it and no reviewer files
   it as a finding.
3. **The all-null projection is not a fence — it has an answer.** `N` zero observations: `sum` = 0,
   `count` = N, `avg` = `0` at the declared scale. A *real value*, not the empty-collection Warning.
   **Zero-length and all-null are DIFFERENT outcomes** and are two distinct test subjects.
4. **`avg`'s divisor is `len(projected)`, and the length invariant is what makes that correct — see
   R10.**

**A null element must still NOT be conflated with a missing one.** That distinction is now
*arithmetically* load-bearing in a way it was not before, and AC19a is the fixture that proves it.

### R10 — `avg`'s divisor comes from `len(projected)`, NOT a second `CollectionLength` call. FORCED.

*(Placed here, out of numeric order, because it is R7.4's consequence and is unreadable apart from
it. R8 and R9 follow below.)*

**Not stylistic — forced by a shipped guard.** D-3.1a.3's reducer inventory
(`internal/reducer_inventory_arch_test.go`) asserts the module contains exactly
`{SumDecimals, AvgDecimals}` with the signature `func([]Decimal) (Decimal, error)`. **Adding a
divisor parameter to `AvgDecimals` changes that signature and trips the inventory.** So "the divisor
is always the collection length" **cannot** be implemented by passing the length in. It must be
`AvgDecimals(projected)` dividing by `len(projected)`, with

> **`len(ProjectCollection(p)) == CollectionLength(collectionOf(p))`**

as the **binding invariant** that makes those two the same number. A second resolver call for the
same collection would also be a second chance to disagree. **This invariant is the only thing
standing between the owner's ruling and a wrong average.**

### R8 — Four states are discriminated BEFORE the kernel is reached

D-3.1a.2 ruled the **kernel's slice argument**, and is safe only because that argument was already
proven to be a real collection that happens to be empty. At the expression layer,
`sum(transactions.amount)` has **four** states that all collapse to a zero-length projection if
nothing stops them:

| state of the collection path | disposition |
|---|---|
| `[]` — present, empty | `sum`→`0`, `count`→`0`, `avg`→ **Warning**, resolves empty |
| **absent from the data** | **located Error** (AD-14) |
| explicit **`null`** | per R7 — a zero observation |
| present, **not an array** | **located Error, never a coercion** (AD-14) |

**A `ProjectCollection` returning `([], nil)` for all four turns three of them into a printed zero
total on a bank statement** — F6's silent-zero hazard one layer above where 3.2 killed it, made
invisible by D-3.1a.2's identity return. The absent and wrong-kind arms must go red **without the
kernel being called at all** (**D-000.9**: the all-clear must not share a code path with the
could-not-look).

### R9 — `avg` over an empty collection is a **Warning**, not a render-aborting Error

Forced by **Story 4.2's AC** (verified in `epics.md`): *"Given a table bound to an empty collection …
the header renders, no data rows are produced, and the render succeeds."* An Error makes that
unsatisfiable.

- The **kernel** keeps its error unchanged (D-3.1a.2 stands).
- The **expression layer** catches it and emits a **`Warning`** naming the collection path and the
  element id, with the aggregate resolving to **empty**.
- **Do not read the Go `error` return as settling AD-14 severity.** Severity is the expression
  layer's to assign. `EXPERIENCE.md:145` defines the vocabulary: *errors (render cannot proceed)* vs
  *diagnostics (render proceeded, with a caveat)*.

**The wire for this does not exist — see F13 / `DECISION-5`.**

---

## Acceptance Criteria

### A. The collection seam (R1, R2)

- **AC1.** `expr.Resolver` declares exactly two new methods, `CollectionLength(path []string) (int,
  error)` and `ProjectCollection(path []string) ([]Value, error)`, implemented in `bind`.
  `expr.Value` **stays scalar** — no array kind is added. `internal/expr` does not import
  `internal/bind`.
- **AC2.** `ProjectCollection` returns **exactly one `Value` per element, in data order** — never
  sorted, deduplicated or filtered. Proven on an order-distinguishable fixture.
- **AC3.** **`count` is STRUCTURALLY unable to reach a projected value**: it calls
  `CollectionLength` and nothing else. Proven structurally, not by convention.
- **AC4.** **THE LENGTH INVARIANT — binding (R10).**
  `len(ProjectCollection(p)) == CollectionLength(collectionOf(p))` holds and is asserted across
  **every** fixture in AC12's table, including the all-null one. It is load-bearing: `avg` divides
  by `len(projected)` (it cannot take a divisor parameter without tripping the reducer inventory —
  R10), so this invariant is **the only thing standing between the owner's null ruling and a wrong
  average.** Its red-proof is AC12a, and no other mutant supplies its independent teeth (AC21a).
- **AC5.** **Neither method takes a range, offset, index or limit** — enforced by **AC22's**
  method-set equality assertion, not by a name list.
- **AC6.** The path's **array prefix** names the collection; remaining segments are the projection.
  `count` takes the **collection** path; `sum`/`avg` take a **projection** path (collection + at
  least one trailing segment). Both are arity 1, `argNotLiteral`. **A bare collection path passed to
  `sum`/`avg` is a located Error at EVALUATION, not at load** — whether a path names a collection is
  data-dependent, and Check runs without data.

### B. The guard, re-pointed (R2) — capture the red-proof FIRST

- **AC7.** In this story's own commit:
  1. **Root selection is one function.** Precedence unchanged and unchangeable: **`params` → row
     alias → data** (D-3.1.1). Both the scalar dispatch and the two collection methods call it.
  2. `rootName` values stay **string literals** — `{"data", "params", "row"}`. **No fourth root.**
  3. **The re-pointed guard's red-proof is captured BEFORE the collection methods are wired**
     (D-000.30): add a fourth literal root reachable **only** from the collection path, watch the
     re-pointed guard redden, remove it. **The window shuts once the extraction lands.**
  4. `TestBindResolutionRootsClosureRedProof` is **re-run after the re-point** — moving a closed
     set's fence is exactly when it needs re-proving, and it is **AD-4's `page` fence**.
  5. **D-3.1.1's *"a known weakness, recorded so a green guard is not over-read"* paragraph is now
     FALSE.** The decision log is append-only, so the story notes that it is **corrected by
     append**, never edited in place.

### C. `sum` is exact (epic Given 1 — AD-23)

- **AC8.** **Capture the inline-accumulator red-proof BEFORE `sum` is wired** (D-000.30, F6). Then
  flip the entry and route to `SumDecimals`. The positive routing assertion (R4/D-000.59) proves the
  aggregate's result equals the kernel's for the same operand sequence; the captured red-proof shows
  it reddening under an inline `big.Int` loop. **This discharges D-3.1a.4's follow-up, owed to 3.3 by
  name.**
- **AC9.** The hand-computed total is asserted against **D-000.61's discriminating corpus A** —
  `12345678901234.56` + 32 × `0.01` — **not** seven small 2-dp amounts (corpus C), which D-000.61
  measured as **byte-identical under `float64`**. The corpus and its measured discrimination are
  stated in the test's own comment. **Order-invariance:** original, reversed and shuffled order all
  yield the **identical** `Decimal`, asserted as a positive property.
- **AC10.** **The float red-proof is HONEST (D-000.61 extension).** The mutant accumulates the
  decimal **values** (`coefficient × 10^exponent`, where `0.01` is inexact in binary) and
  re-quantises — **not** coefficients in `float64`, which are exact below 2^53 and introduce the
  *type* without the *error*. **CORRECTED by [[D-3.3.7]] (finisher pass, Finding 4):** AD-23 bans a
  float **committed** anywhere under `folio-go/`, never one **applied** to the working tree,
  measured, and reverted — a mutation red-proof is a measurement, not an artifact, and [[D-000.61]]
  (extension) was itself produced exactly that way. The order-invariance assertion must redden
  **unconditionally** (forward order ≠ reversed order); the value assertion reddens in the corpus's
  **declared (forward) order** and is measured to **pass by luck in reverse** — record both, per
  D-3.3.7's table, rather than claiming both assertions redden in every order. The mutant is applied
  to the working tree, measured, and reverted — nothing float is committed under `folio-go`; the
  durable, dependency-free demonstration is pinned in `hashmatrix/floatdiscrimination/` (outside
  AD-23's scope by construction, same rationale as `hashmatrix/probe/`), with a `go test` step added
  to that module's CI job so the demonstration actually runs. **The test comment records that the
  coefficient-level mutant does NOT redden, and why** — losing that measurement re-opens the
  extension entry's false alarm; it is now an executing assertion in `hashmatrix/floatdiscrimination/`
  rather than a comment that can decay silently.
- **AC11.** AD-23's two-layer guard re-measured green: no `float32`, `float64` or `math/big.Float`
  introduced at the `folio-go` module root — **re-verified after the AC10 measurement's mutation was
  reverted** (D-3.3.7): `git diff`/`diff` against the pre-mutation copy confirmed byte-identical
  restoration, and the full gate was re-run green.

### D. `count`, `avg`, and the null-as-zero arithmetic (epic Given 2 — R5, R7)

- **AC12.** **One declarative table (D-000.45 — assert the computed value, never a direction)**
  covering, at minimum, these subjects as **distinct** rows:

  | subject | `sum` | `count` | `avg` |
  |---|---|---|---|
  | ordinary projection | hand-computed literal | N | hand-computed literal |
  | `[1, null, 3]` — **null is a zero observation** | `4` | `3` | **`1.33…` at the declared scale** |
  | **all-null**, N elements | `0` | `N` | **`0.0000`** — max operand scale `0` + `avgExtraScale` |
  | **zero-length** collection | **`0`**, exponent `0` | `0` | **Warning**, resolves empty |

  **All-null and zero-length are two different subjects with two different outcomes**, and the two
  zeros are **differently shaped on purpose** — `0.0000` for the all-null average, bare `0` for the
  empty sum (R7.1). Collapsing either pair is the defect this table exists to catch. **Every null
  element resolves to `{Coefficient: 0, Exponent: 0}`**, never to a scale borrowed from the column
  or from the projection's maximum.

- **AC12a.** **THE OPTION-1 RED-PROOF — the honest mutant for the owner's null ruling, named as
  such.** The mutation is *"implement option 1: `ProjectCollection` omits null elements"*, and it
  **must be named in those words in the test**, so the guard's teeth are visibly the owner's decision
  and any later drift back to option 1 reddens the build instead of silently changing every average
  in the product.

  **Why this mutant and no other (D-000.61 extension — introduce the ERROR, not merely a different
  mechanism):** omitting nulls returns `N−k` values, and **`sum` is byte-identical either way** —
  nulls contribute `0` under the owner's ruling and are absent under option 1, and adding zero
  changes nothing. So the `sum` assertion **cannot see the difference between the owner's ruling and
  the arm they rejected.** Only `avg` can. The mutant must therefore redden **the length invariant
  (AC4) and the `avg` assertion, and nothing else** — and the test states that expected blast radius,
  so a mutant that also reddens `sum` is a sign the fixture drifted.

  **Corpus rule (D-000.50 — check the population before writing the assertion):** *any* corpus
  exercising the null rule **must contain a null element AND assert `avg`.* A fixture that exercises
  nulls and checks only `sum` is **D-000.61's corpus C in a new costume** — it looks like it proves
  the null rule and proves nothing, because the two candidate implementations agree on every `sum` it
  can compute.
- **AC13.** `count` returns `Decimal{Coefficient: N, Exponent: 0}` and **never inspects the projected
  field** — an element missing that field still counts. The test comment carries R5's sentence
  verbatim so the asymmetry reads as a consequence.
- **AC14.** `avg` routes to `AvgDecimals` with the **same positive routing assertion and captured
  red-proof shape as AC8**. **`AvgDecimals`' signature is UNCHANGED** — it takes `[]Decimal` and
  nothing else, because a divisor parameter would change `func([]Decimal) (Decimal, error)` and
  **trip D-3.1a.3's reducer inventory** (R10). The divisor is therefore **`len(projected)`**, and
  **AC4's length invariant is what makes that equal the collection length.** A second
  `CollectionLength` call for the same collection is forbidden — it is a second chance to disagree.
- **AC15.** **The scale is DECLARED, not emergent:** `avg` divides at **(maximum operand scale +
  `avgExtraScale`)**, the single existing declaration in `internal/expr/reduce.go`. The expression
  layer **re-declares nothing** and derives nothing from data. Named in the AC test and on the
  reference page. FLAG-1 carries F5's MEDIUM-confidence flag on the value `4` forward to Story 3.4.
- **AC16.** **Round-half-to-even** proven at the declared scale by a table covering: an exact tie
  rounding **up** to even; an exact tie rounding **down** to even; a **negative** exact tie; a
  just-below-half case; a just-above-half case. **Red-proof:** replacing it with
  round-half-away-from-zero reddens the table. (A tie table with no tie in it passes under either
  rule — the red-proof is what proves it has one.)

### E. Four states, discriminated before the kernel (R8)

- **AC17.** Each of the four states in R8's table has its **own fixture** and its **own asserted
  outcome**. A zero-length return from `ProjectCollection` means **"present and empty" and nothing
  else**.
- **AC18.** The **absent** and **wrong-kind** arms go red **without the kernel being called at all**
  (**D-000.9**). Proven, not asserted — e.g. by a kernel call-counter or an equivalent instrument
  the test can observe.
- **AC19.** Per-element defects (R8 applies to the collection; this applies to its elements):
  - projected field **absent** on an element → **located Error**;
  - projected field present but **not a number** → **located Error, never a coercion**;
  - projected field **explicit `null`** → **a zero observation** (R7), never an error.

  Every located Error carries **the collection path, the zero-based element index, the projected
  field path, and the element id.** *(The element index is required: without it the message locates
  the template but not the datum, and FR41's "names which element and which data path" is only half
  met on a 400-row statement.)*
  **Guardrail:** do **not** report only the first defective element — **collect and report**, or
  **state explicitly in the delivery log that it is first-failure and why.**

- **AC19a.** **ONE fixture, not four — the null-vs-absent pair is the sharpest subject available.**
  A single projection holding **one `null` element and one field-absent element**, asserting a
  **number** for the first and a **located Error** for the second. This is the test that proves the
  seam discriminates the two states a sloppy resolver conflates, and it is the **only** place where
  getting it wrong is invisible: `{"amount": null}` and `{}` are shapes most JSON producers treat as
  interchangeable, and under R7 they now diverge into *a value* and *a render abort*. **Prefer this
  single fixture over four separate ones.**

### F. The whole collection, never the page (epic Given 3 — AD-11, AD-4)

- **AC20.** **Root-relative, proven against a shadowing alias.** With a row scope active whose
  declared alias is **literally the collection's own name** (`{"bind": "transactions[]", "as":
  "transactions"}`), `sum(transactions.amount)` still resolves the **whole collection from the data
  root**. **Red-proof:** routing the aggregate operand through the row root reddens it.
- **AC21.** **The substituted-window injection red-proof, reddening ON THE VALUE.** A resolver
  returning only a **contiguous slice** of the collection is injected; the test asserts the
  **hand-computed whole-collection total as a literal** (D-000.45 — a declarative table, never a
  direction) and the slice is **chosen so the sliced total provably differs.** *A slice whose total
  happens to equal the whole is D-000.61's corpus-C trap in a new costume.* State the sliced total
  in the comment, so the discrimination is measured rather than assumed.
- **AC21a.** **HONESTY NOTE — AC21's mutant and AC12a's are NOT independent evidence (D-000.38).**
  The contiguous-slice resolver **also** reddens the length invariant (AC4), so two reds here must
  not be read as two proofs. They test different properties and both are worth having, but **the
  length invariant's *independent* teeth come from the option-1 mutant specifically**, because that
  is the one that leaves `sum` green. State this in the test comments rather than letting the
  coverage be double-counted.
- **AC22.** **No per-page vocabulary — asserted by SET EQUALITY, not a name list.** *(A name list
  dies to the first synonym.)* An AST assertion over **`expr.Resolver`'s method set** — exactly
  `{Resolve, CollectionLength, ProjectCollection}` — in the **same instrument as
  `TestExprFunctionTableIsExactlyEight`**. A page-scoped variant then cannot be added under **any**
  spelling without reddening. AD-4's `page`/`pages` remain reserved tokens resolved from no root.
- **AC23.** **DW-7 is APPENDED TO in this story's commit** (the shape DW-11 was answered at 2.4 and
  2.6). The append names: **what landed** — the expression-layer routing assertion and its captured
  red-proof; **what remains** — the **footer** half, `columns[].footer` evaluating through the same
  path, still **Story 4.5**, still with no mechanism; and **corrects** *"anti-rot mechanism: none
  possible"* → *"none possible for the footer half."* Ownership is unchanged.

### G. Overflow, bounds, and the `%w` gap (F4)

- **AC24.** **Accumulator overflow is a loud, located failure, never a wrap.** An operand set whose
  aligned coefficient exceeds `int64` produces an error naming the **element id**, the **collection
  path**, the **operand count** and the **aligned exponent**. Red-proven **at the expression layer**,
  not only at the kernel — a kernel-layer proof does not prove the wiring.
- **AC25.** Same for the **alignment-spread** breach and `avg`'s **result-exponent** breach: located
  at the expression layer, each red-proven there.
- **AC26.** **Kernel errors are wrapped with `%w`**, never re-formatted. A test asserts both the
  kernel's own wording and the added element id / path are present, **and that the kernel error is
  recoverable by `errors.Is`/`errors.As` rather than by message matching** — so Story 3.6 attaches
  codes without re-plumbing every site (F4.1; AD-14 forbids callers matching on message text).

### H. The record, and the replacement discipline

- **AC27.** `docs/expression-reference.md` **and its `.html` twin** (D-000.47 — both sites recorded
  explicitly), in the same commit:
  1. the *Totals* heading loses *"(not yet implemented — Story 3.3)"*; the status table's Totals row
     changes accordingly;
  2. **the bare `{{sum(transactions.amount)}}` examples are corrected** to state that *an aggregate
     must be formatted before it can appear in text; a bare `{{sum(...)}}` is an Error* (F3). **This
     is 3.3's edit; 3.4's worked `formatNumber` example is a different edit and is not pre-empted
     here.**
  3. the declared `avg` scale (AC15) is stated; the null-as-zero-observation rule (R7) is stated;
  4. **`avg`-on-empty is described as a caveat the render survives, not a failure** (R9), and the
     wording stays **"an error"** where an error is meant — HEAD `b875582` is the commit that fixed
     that distinction, and it is not to be undone.
  5. **Reuse an existing doc-drift check if one fits; do NOT build one in 3.3.**
- **AC28.** `table.go`'s three entries flip to `implemented: true`.
  `TestExprFunctionTableIsExactlyEight`, `TestExprFunctionTableRedProofNinthEntry`,
  `TestExprTableHasNoExportedRegistrationPath` and `TestExprTableMutationRedProof` stay green.
- **AC29.** **`TestThreeImplementedFunctions` / `TestFiveUnimplementedFunctions` are restated
  DERIVATIONALLY** (F11, D-3.1a.3): *every entry with `implemented == false` names an
  `owningStory`*, and *every entry with `implemented == true` has an `evalCall` branch*. So stated,
  they survive **3.3's edit and 3.4's with none.** The test names may change to match.
- **AC30.** **D-000.59's shape applied to `eval_test.go`.** The unimplemented-function table drops
  `sum`/`count`/`avg` and keeps `formatDate`/`formatNumber` — a **replacement**: the positive
  aggregate assertions land in the **same commit** as the deletion, and the two remaining entries
  still prove the located-error arm (including the unselected-branch case and the "unimplemented ≠
  unknown-function" distinction at `eval_test.go:346`).
- **AC31.** **The four stale comments and three wrong line cites in F12 are fixed**, preferring
  **symbol references over line cites**. `TestExactlyOneDecimalDeclarationInTheModule` and
  `reducer_inventory_arch_test.go` pass **with zero functional edits** (comment-only changes are
  fine and are the point).
- **AC32.** **Cadence and gates, BOTH MODULES.** unit / vet / build / lint on `folio-go` **and**
  `lint`, each reported with scope and flags (D-000.26). `TestCorpusMeetsP6ExerciseFloors` stays red
  with **byte-identical** stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — required
  (D-000.17 / D-2.1.14 / D-000.57). The three `lint` FAILs are named as **DW-19, local-only,
  `.gitignore:85`, CI green** — or the figure is not a figure. **Nobody "fixes" DW-19 by weakening
  the rule.**

---

## `DECISION-5` — RESOLVED (ruled by the engineering lead mid-story; arm (a), as recommended below). Development record only — never edited in place; see the Delivery Log's own "DECISION-5 — how it resolved" for the shipped shape.

**Measured (F13):** `Diagnostic`/`Severity`/`DiagCodeTextClippedWidth` live in the **module-root
`folio` package**; the only shipped Warning is minted in `folio/render.go:611`. `expr.Eval` returns
`(Value, error)`; `bind.Resolve` returns `(string, []Substitution, error)`; **neither carries a
diagnostic and neither may import `folio`.** `internal/diag` is Story 3.6's and **D-1.4.2** forbids
minting codes early.

| arm | shape | cost |
|---|---|---|
| **(a) sentinel + caveat return, code minted in `folio`** *(recommended)* | `expr` exports a sentinel (`ErrEmptyAverage`) wrapping the kernel's error; the aggregate resolves to empty; `bind.Resolve` gains a third return carrying package-local caveats; `folio`'s render path converts each into a `Diagnostic` and mints one new code in `diagnostic.go`. | Widens two signatures and mints one code before 3.6. |
| **(b) create `internal/diag` now** | The structured channel arrives early. | **Rejected on sight:** it is Story 3.6's, and it trips the live `absence-diag-package` lint tripwire, whose discharge (DW-6) requires positive assertions this story cannot land. |
| **(c) return the error and let `avg`-on-empty abort the render** | No new plumbing. | **Rejected:** contradicts R9 and makes Story 4.2's AC unsatisfiable. |

**Recommendation: (a).** The precedent is direct and is not D-1.4.2's: **D-2.8.1 minted
`DiagCodeTextClippedWidth` in `folio/diagnostic.go` at Story 2.8**, when the condition shipped.
D-1.4.2's prohibition was against **Story 1.4** minting the two *footer* codes for a condition that
did not yet exist. Here the condition ships in this story, which is exactly when 2.8 minted. The
sentinel also gives Story 3.6 something to attach a code to without re-plumbing (AC26's `errors.Is`
requirement is the same mechanism).

**If the lead prefers otherwise, only AC12's fourth row and R9's wire change; nothing else in this
file moves.**

---

## Flags — and what each one SUPPRESSES (D-000.60)

- **FLAG-0 — DISSOLVED: the null fence, and its rejected arm is now a NAMED MUTANT.** The creation
  draft carried a fence proposing that an explicit `null` element be a located Error. **The owner
  ruled against it** (R7): a null element is a **zero observation** resolving to `{0, 0}`. The fence
  is **dissolved, not carried** — no AC asserts an error on a null element, and no residual "consider
  erroring on null" survives anywhere in this file. D-000.60 requires a dissolved fence to be visible
  as dissolved rather than simply absent.
  **The rejected sibling arm did not merely disappear — it became the guard.** *Option 1*
  (`ProjectCollection` omits nulls) is written into **AC12a as the named mutation**, so a later drift
  back to it reddens the build instead of silently changing every average in the product.
  *Suppresses:* re-litigating null semantics, and — separately — reading AC21's slice mutant as
  independent evidence for AC4 (see AC21a).
- **FLAG-1 — `avgExtraScale = 4` is ILLUSTRATIVE at MEDIUM confidence.** Carried from Story 3.1a's
  Flag F3, **not resolved here**; honesty-keeper is **Story 3.4**. *Suppresses:* any attempt to
  "confirm" 4. This story declares the scale; it does not ratify the number.
- **FLAG-2 — "Done" is at the evaluation layer, not at rendered text (F3).** *Suppresses:* an
  end-to-end render assertion for the first epic AC, and any relaxation of `bind.Resolve`'s
  never-coerced number arm to obtain one. That arm is pinned by two deliberate tests.
- **FLAG-3 — DISCHARGED by R2/AC7, not carried.** The creation draft flagged that the roots guard
  might not cover the new dispatch. The ruling **re-points the guard** instead, so the flag is
  answered rather than accepted. *Suppresses:* the two-guard remedy, except as R2's stated fallback,
  which returns to the lead as a finding.
- **FLAG-4 — DW-7's footer half still has no forcing function.** *Suppresses:* any claim that AC23
  gave it one. AC23 corrects the record for the half that **did** gain a mechanism and is explicit
  that the footer half did not.
- **FLAG-5 — `DECISION-5` is open.** *Suppresses:* starting AC12's zero-length row, or R9's wire, on
  a guess.

---

## Cross-story obligation — read this before choosing any shape

**DW-7 is live and owned by Story 4.5.** It requires that `columns[].footer`'s `sum`/`count`/`avg`
use the **same aggregate evaluation** as the `{{sum(...)}}` family — **one implementation, never two
that drift** (D-1.4.1, D-1.4.2).

**This story builds that one implementation. The shape 3.3 lands is what 4.5 will be held to.** The
developer must not build something 4.5 cannot reuse: the aggregate evaluation must be callable with
a collection and a projection **without** a text placeholder, **without** a parsed `{{ }}`
expression, and **without** a text-element context — because 4.5's caller holds a `columns[].footer`
declaration and a `footerOf` path (D-1.4.1), not a binding string. `DeriveFooterOf`
(`internal/expr/footer.go`) already resolves that path shape at load time; the evaluation entry point
this story lands is what 4.5 will hand it to. **R1's two-method split is what makes a `count` footer
reusable without a numeric source** — the same line DW-5's discharge already draws.

**Routed, not ruled — a referent for Epic 6, owed by nobody here.** R7 creates a cliff an author can
walk off without warning: a column bound to a field that is **`null`** in some rows shows a number,
while one bound to a field that is **absent** in some rows aborts the render. That is AD-14 working
as designed and **3.3 owes nothing**, but two later stories inherit it and should not rediscover it:
**Story 6.6** ("present a failed render honestly" — *names the element and the data path*) is where
the abort must read honestly, and **Stories 6.1/6.2** (the sample-data tree and pick-a-path binding)
are where an author would ideally have seen it coming. Recorded here so those stories inherit the
referent.

---

## Task breakdown

**All 15 tasks complete.** DECISION-5 was ruled by the engineering lead mid-story (sentinel +
caveat return, code minted in `folio`, per the message logged below); tasks 9–15, initially blocked,
proceeded once the ruling landed. See the Delivery Log for what each task actually produced.

[x] 1. **Re-measure both gates** at implementation HEAD (`folio-go` and `lint`), with scope and flags.
   Confirm `TestCorpusMeetsP6ExerciseFloors` red with byte-identical stats and the three `lint`
   failures as DW-19. Confirm the tree is clean apart from any pre-existing decision-log edit.
[x] 2. **Re-run this story's F1 probe** and the F12 citation audit against the tree as it stands; report
   any drift before writing code.
[x] 3. **`DECISION-5` must be resolved before task 9.** Everything else is ruled and may proceed.
[x] 4. **CAPTURE THE TWO D-000.30 RED-PROOFS FIRST, before any wiring** — both windows shut the moment
   3.3 lands:
   - the **fourth-root** proof against the re-pointed guard (AC7.3);
   - the **inline-accumulator** proof against the routing assertion (AC8, F6).
[x] 5. **Extract root selection and re-point the guard** (AC7): one helper, precedence `params` → row
   alias → data, literal root names, `TestBindResolutionRootsClosureRedProof` re-run after the
   re-point, and the note that D-3.1.1's weakness paragraph is corrected **by append**.
[x] 6. **Build the two-method seam** (AC1–AC6), including AC4's length invariant and AC3's structural
   proof that `count` cannot reach a projected value.
[x] 7. **Discriminate the four states before the kernel** (AC17–AC18), with the no-kernel-call
   instrument, then the per-element arms (AC19) with the element index and the collect-or-declare
   guardrail, and **AC19a's single null-plus-absent fixture** — one projection, one `null` element
   asserting a number and one field-absent element asserting an Error. Prefer it over four fixtures.
[x] 8. **Wire `sum`** (AC8–AC11): flip, route, land the routing assertion against the captured proof,
   then D-000.61's corpus, order-invariance, and the honest mutant with its recorded
   coefficient-level non-result.
[x] 9. **Wire `count` and `avg`** (AC12–AC16): the four-subject declarative table with its two
   differently-shaped zeros; null-as-`{0,0}` arithmetic; **`len(projected)` as the divisor with
   `AvgDecimals`' signature untouched** (R10); declared scale; tie table; round-half-away red-proof.
   Then **AC12a — the option-1 mutant, named in those words**, asserting it reddens **AC4 and `avg`
   and nothing else**, with the corpus rule enforced (a null fixture that asserts only `sum` proves
   nothing). **AC12's zero-length row waits on `DECISION-5`.**
[x] 10. **Bind AD-11** (AC20–AC22): shadowing-alias fixture with row-root red-proof; the
    substituted-window proof reddening **on a literal value** with the sliced total stated; the
    `expr.Resolver` method-set equality assertion; and **AC21a's honesty note** that this mutant and
    AC12a's are not independent evidence for AC4.
[x] 11. **Overflow, bounds and `%w`** (AC24–AC26), each red-proven at the expression layer, with
    `errors.Is`/`errors.As` recoverability asserted.
[x] 12. **Replacement discipline** (AC30): positives and deletion in one change.
[x] 13. **Update the record** (AC23, AC27–AC29, AC31): DW-7's append; both reference-page sites; the
    derivational restatement of the two table tests; the stale comments and line cites.
[x] 14. **Re-run both gates** (AC32) and write the Delivery Log: files touched, ACs mapped, gates with
    scope and flags, **every red-proof named with what it reddened under**, R1's spellings justified
    if changed, and every flag restated — including FLAG-0 as dissolved.
[x] 15. **Set the story status to `review`** in `sprint-status.yaml`.

---

## Delivery Log

### DECISION-5 — how it resolved

The engineering lead ruled DECISION-5 mid-story, before task 9. Arm (a), as recommended: `expr`
exports a typed, non-error `Caveat` (`internal/expr/caveat.go`, `CaveatKind`/`CaveatEmptyAverage`),
`Eval`'s signature widens to `(Value, []Caveat, error)`, and **both** `bind.Resolve` and
`bind.BindTextSpans` widen to a third return (the lead's own sequencing note: `BindTextSpans`, not
`Resolve`, is the production call site, `render.go:536` — both signatures moved together). `folio`
mints `DiagCodeEmptyAverage` in `diagnostic.go`, following `DiagCodeTextClippedWidth`'s own
precedent. The caveat is carried as a **separate return**, never folded into `bind.Substitution`
(the lead's own trap warning: `shiftSubstitutions` reconstructs that struct field-by-field and would
silently drop an added field). Sequencing followed exactly as instructed: the root-selection
extraction and re-point (task 5) and its red-proof (task 4) landed **before** the caveat return was
added — verified by running `TestBindResolutionRootsAreClosed`/`TestBindResolutionRootsClosureRedProof`
green immediately after the extraction, before touching `Eval`'s signature.

**D-3.3.6 binding 2's "recorded with its reason" (finisher pass, Finding 21):** the slice element type
(`Caveat`) lives in `internal/expr`, not `internal/bind`, and the reason — carried by `caveat.go`'s own
doc comment but missing from this record until now — is that `Caveat` is **not** a `folio.Diagnostic`:
`internal/expr` may not import the module-root `folio` package (that rank runs backwards), and Story
3.6 owns the general diagnostic-code registry. `expr` is the aggregate condition's own home — it mints
no code, decides no severity, and constructs nothing `folio`-shaped — so a higher-ranked caller
(`internal/bind`, then `folio`'s own render path) turns it into a `Diagnostic` without `expr` needing
to know how.

### Files touched

**`folio-go` — production**
- `internal/expr/ast.go` — `Resolver` gains `CollectionLength`/`ProjectCollection` (AC1).
- `internal/expr/caveat.go` (new) — `Caveat`/`CaveatKind` (DECISION-5).
- `internal/expr/eval.go` — `Eval`/`evalCall`/`evalIf` widen to carry `[]Caveat`; dispatch to
  `evalSum`/`evalCount`/`evalAvg`.
- `internal/expr/aggregate.go` (new) — `evalSum`/`evalCount`/`evalAvg`, `KernelOverflowError`,
  `decimalsFromProjection`, `aggregateOperandPath`.
- `internal/expr/table.go` — `sum`/`count`/`avg` flip to `implemented: true`.
- `internal/expr/doc.go` — package doc updated (F12-adjacent: no longer claims 3.3's own functions
  are unimplemented).
- `internal/expr/reduce.go` — F12: three wrong line cites replaced with symbol references (no
  functional change; kernel unchanged per D-3.1a.2).
- `internal/bind/text.go` — `selectRoot` extracted (R2); `exprResolver.Resolve` re-routed through it;
  `CollectionLength`/`ProjectCollection`/`splitCollectionPath`/`collectionSubPath` added; `Resolve`/
  `BindTextSpans`/`BindText` widen for the caveat return.
- `internal/template/decimal.go` — F12: four stale `internal/bind` comments corrected to
  `internal/expr`.
- `internal/reducer_inventory_arch_test.go` — F12: one stale `internal/bind` comment corrected.
- `render.go` — `BindTextSpans` call site widened; `diagnosticFromCaveat` added; caveats appended to
  `diags` before the element's own clip-width Warning (D-2.8.6 ordering).
- `diagnostic.go` — `DiagCodeEmptyAverage` minted.

**`folio-go` — tests**
- `internal/expr/eval_test.go` — `Eval` call sites updated for the 3-return signature;
  `mapResolver` gains `CollectionLength`/`ProjectCollection`; `TestUnimplementedFunctionsAreLocatedErrors`
  and `TestUnimplementedAndUnknownFunctionErrorsAreDistinguishable` restated per AC30 (formatDate
  replaces sum as the unimplemented example).
- `internal/expr/table_test.go` — the two hard-coded name-list tests removed (moved, derivationally).
- `internal/expr/table_derivational_test.go` (new) — AC29: `TestImplementedEntriesMatchEvalCallSwitch`
  (AST-derived), `TestExactlyTwoUnimplementedRemainAfterThisStory`.
- `internal/expr/routing_arch_test.go` (new) — AC8/AC14/F6/D-3.1a.4: positive routing assertions for
  `evalSum`→`SumDecimals` and `evalAvg`→`AvgDecimals`, plus the captured inline-accumulator red-proof.
- `internal/expr/aggregate_test.go` (new) — AC21/AC21a: `windowedResolver`, the substituted-window
  red-proof.
- `internal/expr_arch_test.go` — AC22: `TestExprResolverMethodSetIsClosed` and its red-proof.
- `internal/bind/aggregate_test.go` (new) — AC1–AC6, AC12/AC12a, AC13, AC17–AC19a, AC20.
- `internal/bind/aggregate_precision_test.go` (new) — AC9–AC11, AC16, AC24–AC26.
- `internal/bind/resolution_roots_arch_test.go` — R2: re-pointed at `selectRoot`'s return statements;
  red-proof mutation shape updated to match.
- `internal/bind/scope_test.go`, `internal/bind/spans_test.go`, `row_scope_test.go`,
  `layout_probe_test.go` — mechanical signature updates only (extra `_` for the new return values).
- `render_empty_average_diagnostic_test.go` (new) — DECISION-5/R9 end-to-end through `Render`,
  including the pipeline-stage ordering guarantee (bind-stage caveat before layout-stage clip).

**Docs and records**
- `docs/expression-reference.md`, `docs/expression-reference.html` — AC27: Totals section rewritten
  (no longer "not yet implemented"; bare-aggregate-is-an-error stated; null-as-zero-observation
  stated; avg-on-empty stated as a survivable caveat); status table row updated.
- `_bmad-output/implementation-artifacts/deferred-work.md` — AC23: DW-7 appended (what landed, what
  remains — the footer half — and the "none possible" → "none possible for the footer half"
  correction).
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — AC7.5: D-3.1.1's "known
  weakness" paragraph corrected by append (never edited in place).
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — status → `review`.

### AC map (selected; see task breakdown above for the full list)

- **AC1–AC6** (seam): `internal/bind/text.go` (`CollectionLength`/`ProjectCollection`), proven in
  `internal/bind/aggregate_test.go` (`TestProjectCollectionPreservesDataOrder`,
  `TestCollectionLengthNeverConsultsProjectedField`, `TestBareCollectionPathToSumIsLocatedErrorAtEvaluation`).
- **AC7**: `selectRoot` (`text.go`), re-pointed guard + red-proof
  (`internal/bind/resolution_roots_arch_test.go`).
- **AC8–AC11**: `evalSum`/`SumDecimals` routing + red-proof (`routing_arch_test.go`), D-000.61 corpus
  A + order-invariance (`internal/bind/aggregate_precision_test.go`).
- **AC12/AC12a**: `TestAggregateDeclarativeTable`, `TestOption1RedProofProjectCollectionOmitsNulls`
  (`internal/bind/aggregate_test.go`).
- **AC13–AC16**: same file, plus `TestAvgRoundHalfToEvenAtTheSeam`
  (`internal/bind/aggregate_precision_test.go`).
- **AC17–AC19a**: `TestFourCollectionStatesDiscriminated`, `TestPerElementDefectsAC19a`,
  `TestPerElementWrongKindIsLocatedErrorNeverCoerced` (`internal/bind/aggregate_test.go`).
- **AC20–AC22**: `TestAggregateBypassesRowRootEvenWhenAliasShadowsCollectionName`
  (`internal/bind/aggregate_test.go`); `TestSubstitutedWindowRedProofDiffersFromWholeCollectionTotal`
  (`internal/expr/aggregate_test.go`); `TestExprResolverMethodSetIsClosed` + red-proof
  (`internal/expr_arch_test.go`).
- **AC23**: `deferred-work.md`'s DW-7 append.
- **AC24–AC26**: `TestSumOverflowIsLocatedAtExpressionLayer`,
  `TestAvgResultExponentOverflowIsLocatedAtExpressionLayer` (`internal/bind/aggregate_precision_test.go`).
- **AC27**: docs, both sites.
- **AC28–AC30**: `table.go`, `table_derivational_test.go`, `eval_test.go`.
- **AC31**: `reduce.go`, `internal/template/decimal.go`, `internal/reducer_inventory_arch_test.go`.
- **AC32**: gates below.
- **DECISION-5/R9**: `render_empty_average_diagnostic_test.go`.

### Every red-proof, named with what it reddened under

1. **Fourth-root proof (AC7.3, D-000.30):** captured against the re-pointed `selectRoot`-keyed guard
   (task order: extraction/re-point landed first, per the DECISION-5 sequencing note) —
   `TestBindResolutionRootsClosureRedProof` injects a duplicate `selectRoot` returning `"page"`;
   reddens the closure check.
2. **Inline-accumulator proof (AC8, F6, D-3.1a.4):** `TestSumRoutingRedProofInlineAccumulator`
   replaces `evalSum`'s `SumDecimals(decimals)` call with an inline `big.Int` loop in a scratch copy
   of `aggregate.go`; `TestSumRoutesThroughSumDecimals`'s own AST scan no longer finds the call —
   reddens.
3. **Option-1 mutant (AC12a):** `TestOption1RedProofProjectCollectionOmitsNulls` — named in those
   words in the test. Reddens AC4's length invariant (2 vs 3) and the `avg` assertion; `sum` is
   byte-identical under both arms (measured in the same test).
4. **Substituted-window mutant (AC21):** `TestSubstitutedWindowRedProofDiffersFromWholeCollectionTotal`
   — a resolver returning only elements `[0:2)` of 4 produces `30.00`, provably different from the
   hand-computed whole total `100.00`.
5. **AC21a honesty note, measured, not just asserted:** the same test shows the windowed resolver is
   internally self-consistent (its own `CollectionLength` agrees with its own `ProjectCollection`
   length) — demonstrating it is invisible to the length invariant alone, which is exactly why the
   option-1 mutant (#3) is the invariant's only *independent* red-proof, not this one.
6. **AC22 fourth-method proof:** `TestExprResolverMethodSetRedProofFourthMethod` injects a fourth
   method onto a scratch copy of the `Resolver` interface; the method-set equality check catches it.
7. **AC9's discrimination check:** `TestSumIsExactOnD00061CorpusA` asserts the corpus's own exact
   answer; it does not re-derive the float64 comparison (AD-23's guard forbids `float64` anywhere
   under `internal/`, including inside a demonstrating test), so the discrimination is cited from
   D-000.61's own prior measurement rather than re-measured here.
8. **AC16's red-proof is inherited, not re-derived:** `TestAvgRoundHalfToEvenAtTheSeam` reuses
   `internal/expr/reduce_test.go`'s own constructed tie fixtures (`TestAvgDecimalsRoundsHalfToEven`)
   at the seam, proving the SAME rounding is reached through `avg()`; the round-half-away-from-zero
   red-proof for that rounding rule lives at the kernel (unchanged, D-3.1a.2) and is not re-proven
   here — re-deriving it would test the kernel a second time, not the seam.

### R1's spellings

`CollectionLength`/`ProjectCollection` kept as specified — no better name surfaced during
implementation.

### Flags, restated

- **FLAG-0 — DISSOLVED.** The null-error fence never existed in this implementation; its rejected
  arm is the option-1 mutant, named and red-proven (above).
- **FLAG-1 — `avgExtraScale = 4` remains ILLUSTRATIVE, MEDIUM confidence.** Not touched, not
  confirmed. Honesty-keeper: Story 3.4.
- **FLAG-2 — "Done" is at the evaluation layer.** Confirmed: the only end-to-end `Render` path this
  story can exercise is an aggregate that resolves to `KindNull` (an empty average); a non-empty
  `sum`/`avg`/`count` still cannot reach rendered text without Story 3.4's `formatNumber` (proven
  negatively by `TestAvgOverNonEmptyCollectionProducesNoCaveat`).
- **FLAG-3 — DISCHARGED by R2/AC7,** as planned; not carried forward.
- **FLAG-4 — DW-7's footer half still has no forcing function.** Unchanged; stated in the DW-7
  append.
- **FLAG-5 — DECISION-5 was open; now RESOLVED** (see above). No longer suppresses anything.

### Gates (AC32), measured at implementation HEAD

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 go test ./... -count=1 -v` | **776 pass · 1 fail · 1 skip**, 15 packages |
| `lint/` | `CGO_ENABLED=0 go test ./... -count=1` | **89 pass · 3 fail**, 4 packages |
| `folio-go/internal/expr`, `folio-go/internal/bind` | `go test ./internal/expr/... ./internal/bind/...` | **176 pass**, 2 packages |
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l .` | build ok · vet clean · gofmt clean |
| `lint/` | `go build ./... && go vet ./...` | build ok · vet clean |

The one `folio-go` FAIL is still `TestCorpusMeetsP6ExerciseFloors`, byte-identical stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — required red (D-000.17/D-2.1.14/D-000.57),
untouched by this story. The three `lint` FAILs are still DW-19 (`.font-sources/`, gitignored,
local-only, CI green) — untouched, root cause unchanged, no fourth failure introduced.

The pass count rose from the baseline 740/89 because this story adds tests; no baseline test was
deleted without a same-commit replacement (AC29/AC30's own replacement discipline; the two removed
hard-coded name-list tests are replaced by `table_derivational_test.go`).

### Finisher pass (2026-08-26)

**Triage: 1 Blocker + 10 Majors + 7 Minors + 3 Nits, all 21 FIX.** Zero DISMISS, zero DEFER — every
finding had a concrete, in-scope, cheaply-verifiable resolution; none required expanding scope beyond
the story's own ACs (Finding 4/AC10 required an engineering-lead ruling, recorded as [[D-3.3.7]], but
the resolution still landed inside this story's scope). Full per-finding disposition is in **Finding
Resolutions (finisher)** at the end of this file.

**The Blocker (Finding 1), by construction, not by a stronger scan.** `rootName`/`rootDesc` are now
one defined type, `rootKind` (`internal/bind/text.go`) — the reviewer's own injection,
`lookupBound(..., "page", "the current page")` translated to the new signature as
`lookupBound(..., "page")`, was built with `go build ./...` and produced, **verbatim**:
`internal/bind/zzscratch_redproof.go:6:60: cannot use "page" (untyped string constant) as rootKind
value in argument to lookupBound`. This is a compile-time property, recorded as one (D-000.24), never
dressed as a test. The scratch file was deleted immediately after and `go build ./...` re-run clean.
The SEPARATE property — is the declared root set closed — is re-pointed a second time, from scanning
`selectRoot`'s return statements to AST set-equality over every `rootKind` composite literal in the
package (`resolution_roots_arch_test.go`, fully rewritten), red-proven by injecting a fourth
`rootKind` composite literal into a scratch copy of `text.go`. Recorded at length in [[D-3.3.8]]
(append-only correction to [[D-3.1.1]]/[[D-3.3.1]], never edited in place, per [[D-000.49]]: this
entry does not repeat the "now closed" overclaim it corrects).

**AC10's escalation (Finding 4), ruled by the engineering lead as [[D-3.3.7]].** AD-23 bans a float
**committed** under `folio-go/`, never one **applied** to the working tree and reverted. `SumDecimals`
was mutated in place (working tree only, never committed) to accumulate each operand's own decimal
VALUE in `float64` and re-quantise; measured on D-000.61's corpus A: forward (declared) order gives
`{1234567890123487,-2}` (one satang short), reversed order gives `{1234567890123488,-2}` (exact, by
coincidence) — so the order-invariance assertion reddens unconditionally and the value assertion only
in the declared order, exactly as [[D-3.3.7]] states. The mutation was reverted and `diff`-confirmed
byte-identical; `go test ./internal/expr/... ./internal/bind/...` and the full AD-23 float guards were
re-run green afterward. The durable, dependency-free demonstration is pinned in
`hashmatrix/floatdiscrimination/` (three tests: the value-level mutant misses in declared order, its
order-invariance reddens, the coefficient-level negative control is exact both orders) — outside
AD-23's scope by construction, same rationale as `hashmatrix/probe/`. `hashmatrix`'s CI job gained a
`go test -count=1 ./...` step (it previously ran none).

**Finding 2 (AC5/AC22): a new, type-based lint rule, not an edit to the AST-only guard.** Per the
engineering lead's ruling, `lint/internal/rules/resolvermethodset.go` (new) loads `internal/expr` with
`go/types` and asserts `expr.Resolver`'s method set — names AND full resolved signatures — is exactly
the closed three, closing both evasions the review found: a same-named method WIDENED with an
offset/limit parameter, and a fourth method contributed through an EMBEDDED interface (invisible to
an AST walk over `ast.InterfaceType.Methods.List`, which has zero `Names` for an embedded field;
`go/types`' method-set expansion has no such blind spot). Two dedicated fixture trees under
`folio-go/testdata/lint/resolver-method-set/` (`widened-signature/`, `embedded-interface/`) plus a
`compliant/` negative control; both red-proofs pass. `folio-go/internal/expr_arch_test.go`'s own
AST-only test is unchanged — it still correctly closes the NAME set; it was never capable of the
signature/embedding property, and D-3.1a.1's own precedent is why that property belongs in `lint`.

**Findings 3, 5, 6, 8, 9, 10, 11, 12, 13, 14, 16, 17, 18, 19, 20, 21 — one line each, fixed in place:**
- **3 (AC18):** the fictional `noopCounter` comment is replaced by a real structural instrument,
  `TestKernelCallIsGuardedByProjectCollectionError` (`internal/expr/routing_arch_test.go`), AST-proving
  `evalSum`/`evalAvg`'s kernel calls are textually guarded by an immediate error-check-and-return on
  `ProjectCollection`'s result, with a captured red-proof removing that guard.
- **5 (AC25):** `TestSumAlignmentSpreadOverflowIsLocatedAtExpressionLayer`
  (`aggregate_precision_test.go`) exercises the alignment-spread breach at the expression layer (two
  individually-legal exponents, 50000 and -50001, whose spread of 100001 exceeds the bound), mirroring
  the existing overflow/result-exponent tests.
- **6 (AC29):** the hard-coded `if n != 2` count is replaced by
  `TestUnimplementedEntriesHaveNoEvalCallBranch`, the true derivational inverse of
  `TestImplementedEntriesMatchEvalCallSwitch`; the non-existent function name and the "count, not a
  name list" false survivability claim are both gone.
- **8 (AC27.2):** both doc twins' code samples reverted to the bare `{{sum(...)}}`/`{{avg(...)}}`
  (marked Error) / `{{count(...)}}` (marked its exception) forms; the `formatNumber` worked example is
  left to Story 3.4, with a sentence saying so.
- **9 (plain-terms opener, and its own echo in code):** the opener is rewritten (see below); the SAME
  false claim — "only the first is a legitimate zero" — is also corrected in `text.go`'s file-level
  comment on the collection seam, which had contradicted itself one paragraph later.
- **10 (AC12):** a fifth declarative-table row, `{"t":null}` (the collection path itself is null), now
  asserts `sum`/`count`/`avg` through the real aggregates (`sum=0, count=1, avg=0.0000`); both doc
  twins state the rule.
- **11 (D-3.3.6):** two new tests in `render_empty_average_diagnostic_test.go` build a genuinely
  3-page document (reusing `multi_page_composition_test.go`'s own measured geometry/sentence-count
  pairing) with the reviewer's exact page-footer trap case, and a page-header twin; both assert exactly
  one `AGGREGATE_EMPTY_AVERAGE` Warning across all pages, with the right `ElementID`/`DataPath`.
- **12 (AC8/AC14):** the vestigial-call bound is now STATED in `aggregate.go`'s own file header
  (D-000.24), not built against — a disproportionate call-graph/data-flow instrument was explicitly
  declined, mirroring D-3.1a.4's own ruling against one for the reducer inventory.
- **13:** `TestSumRoutingRedProofInlineAccumulator`'s comment no longer claims the injected mutant
  would compile and compute the right answer (`bigInt`/`bigIntFromInt64` never existed, and the
  original ignored `d.Exponent`); it is now honestly described as a textual routing-removal, with the
  claim that a real compiling exponent-aware accumulator ALSO reddens the assertion independently
  verified (working-tree mutation, measured green on D-000.61's corpus, reverted).
- **14 (AC14):** `TestAvgRoutingRedProofInlineAccumulator`, the missing mirror of the sum red-proof,
  added.
- **16 (AC16):** `TestAvgRoundHalfToEvenAtTheSeam`'s comment no longer claims the routing assertion
  proves anything about rounding; it now states plainly that the round-half-away red-proof is
  inherited from the kernel's own test, matching the Delivery Log's own honest wording.
- **17 (AC19):** the first-failure declaration this line always claimed lived "in the delivery log"
  now actually does (this sentence); `text.go`'s self-citation is corrected to point at it.
- **18 (AC15):** both doc twins now state the RULE (max operand scale plus a fixed extra), hedging only
  the CONSTANT, and the "declared scale" phrase used for the all-null example is resolved to match.
- **19:** the fifth stale `internal/bind` citation, in `testdata/lint/no-bigfloat-type/compliant.go`,
  corrected to `internal/expr`.
- **20:** [[D-3.3.9]] appends a correction to [[D-3.3.5]]'s heading wording ("amended in place" →
  the append it actually was); the shipped `deferred-work.md` append itself needed no change.
- **21 (DECISION-5 binding 2):** the reason for the `expr`-not-`bind` choice is now stated in this
  Delivery Log's own DECISION-5 section (see above), not only in `caveat.go`'s doc comment.

**Files this pass modified or added** (see the AC map and Files-touched list above for what the
developer already touched; this pass's own changes):
- `folio-go/internal/bind/text.go` — `rootKind` type (Finding 1/[[D-3.3.8]]); `selectRoot`/
  `lookupBound`/`collectionSubPath`/`CollectionLength`/`ProjectCollection` updated; two stale/
  self-contradicting comments corrected (Findings 9, 17).
- `folio-go/internal/bind/resolution_roots_arch_test.go` — fully rewritten for the `rootKind`-keyed
  scan (Finding 1/15).
- `folio-go/internal/bind/scope.go` — one comment corrected to match the type change.
- `folio-go/internal/bind/aggregate_test.go` — AC4 `t.Fatalf` → `t.Errorf` (Finding 7); the
  `noopCounter` comment corrected (Finding 3); a fifth declarative-table row (Finding 10).
- `folio-go/internal/bind/aggregate_precision_test.go` — AC9 comment corrected (Finding 4/[[D-3.3.7]]);
  the alignment-spread test added (Finding 5); the round-half-to-even comment corrected (Finding 16).
- `folio-go/internal/expr/aggregate.go` — the vestigial-call bound stated (Finding 12).
- `folio-go/internal/expr/routing_arch_test.go` — the dishonest mutant comment corrected (Finding 13);
  the avg routing red-proof added (Finding 14); the AC18 structural instrument and its red-proof added
  (Finding 3).
- `folio-go/internal/expr/table_derivational_test.go` — the hard-coded count test replaced (Finding 6).
- `folio-go/render_empty_average_diagnostic_test.go` — the page-footer/page-header multi-page tests
  added (Finding 11).
- `folio-go/testdata/lint/resolver-method-set/{compliant,widened-signature,embedded-interface}/` (new)
  — fixtures for the new lint rule (Finding 2).
- `folio-go/testdata/lint/no-bigfloat-type/compliant.go` — stale citation fixed (Finding 19).
- `lint/internal/rules/resolvermethodset.go`, `resolvermethodset_test.go` (new) — the AC5/AC22
  type-based rule and its tests (Finding 2).
- `hashmatrix/floatdiscrimination/floatdiscrimination_test.go` (new) — AC10's durable demonstration
  ([[D-3.3.7]]).
- `.github/workflows/ci.yml` — `hashmatrix` job gained a `go test` step ([[D-3.3.7]]).
- `docs/expression-reference.md`, `docs/expression-reference.html` — AC27.2 code samples reverted
  (Finding 8); the avg scale rule restated (Finding 18); the null-collection-path rule stated
  (Finding 10).
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — [[D-3.3.7]] (AC10 ruling),
  [[D-3.3.8]] (Finding 1 correction), [[D-3.3.9]] (Finding 20 correction), all appended.
- This story file — AC9/AC10 text corrected per [[D-3.3.7]]; plain-terms opener rewritten (Finding 9);
  DECISION-5 section's reason clause added (Finding 21); this subsection; **Finding Resolutions
  (finisher)** at the end; status → `done`.

**Red-proof discipline, stated once for this pass:** every mutation that touched a real production
file (`reduce.go`'s honest-float measurement, `aggregate.go`'s honest-mutant verification,
`text.go`'s option-1 measurement, and the blocker's own compile-time scratch file) was applied to a
BACKED-UP copy, measured with the real toolchain, and reverted with a byte-for-byte `diff` check
before moving on — none of these mutations are committed. Every mutation that is a PERMANENT red-proof
(the new/rewritten tests listed above) is a real assertion the suite runs on every future change, not
a one-time manual demonstration.

**Gates, re-measured after this pass (D-000.26 — scope and flags):**

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 go test ./... -count=1` | **783 pass · 1 fail · 1 skip**, 15 packages |
| `lint/` | `CGO_ENABLED=0 go test ./... -count=1` | **95 pass · 3 fail**, 4 packages |
| `hashmatrix/` | `go test -count=1 ./...` | **3 pass**, 1 package with tests (`floatdiscrimination`) |
| `folio-go/internal/expr`, `folio-go/internal/bind` | `go test ./internal/expr/... ./internal/bind/... -count=1` | **181 pass**, 2 packages |
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l .` | build ok · vet clean · gofmt clean |
| `lint/` | `go build ./... && go vet ./... && gofmt -l .` | build ok · vet clean · gofmt clean |
| `hashmatrix/` | `go build ./... && go vet ./... && gofmt -l .` | build ok · vet clean · gofmt clean |

The one `folio-go` FAIL is still `TestCorpusMeetsP6ExerciseFloors`, byte-identical stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — required red, untouched. The three `lint`
FAILs are still the same DW-19 root cause, untouched. `folio-go` rose 776→783 (+7: the alignment-
spread test, the two AC18 instrument tests, the avg routing red-proof, the null-collection-path
subtest, and the two multi-page caveat tests). `lint` rose 89→95 (+6: the new resolver-method-set
rule's tests). `hashmatrix` gained its first-ever test package (+3, previously 0 — it had a probe
binary and no tests at all).

**Cross-target hash matrix: not run — per-epic cadence (D-000.4).** This story is not hash-shaped and
this pass touched no production byte the matrix would need to re-certify (`text.go`'s and
`aggregate.go`'s changes are internal refactors behind the same public `Render` surface, already
covered by the full non-matrix suite above).

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer (adversarial, Story 3.3 by number — no auto-pick)
- **Date:** 2026-08-26
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 1
- **Majors:** 10
- **Minors:** 7
- **Nits:** 3

### Gates, re-measured independently by the reviewer (D-000.26 — scope and flags)

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 go test ./... -count=1 -json` | **776 pass · 1 fail · 1 skip**, 15 packages — matches the Delivery Log |
| `folio-go/` | `go build ./...` · `go vet ./...` · `gofmt -l .` | build ok · vet clean · **gofmt clean (empty output, verified with bare `gofmt`, not through the token filter)** |
| `lint/` | `CGO_ENABLED=0 go test ./... -count=1 -json` | **89 pass · 3 fail**, 4 packages — matches |
| `lint/` | `go build ./...` · `go vet ./...` | build ok · vet clean |
| `folio-go/internal/expr`, `folio-go/internal/bind` | `go test ./internal/expr/... ./internal/bind/... -count=1` | **176 pass · 0 fail** — matches |

The one `folio-go` FAIL is `TestCorpusMeetsP6ExerciseFloors`, **required red** (D-000.17 / D-2.1.14 /
D-000.57). Stats measured this run: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` —
**byte-identical to the declared baseline. No drift.**

The three `lint` FAILs are `TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`,
`TestFontsAssetsNoticeRemovalRedProof` — one root cause, **DW-19**, `.font-sources/` gitignored at
`.gitignore:85`, **local-only, CI green** (D-000.64). **No fourth failure; the three are unchanged.**

**Working tree restored after every probe.** SHA-256 verified byte-identical for
`internal/bind/text.go`, `internal/expr/aggregate.go`, `internal/expr/ast.go`,
`internal/bind/aggregate_test.go`, `internal/bind/resolution_roots_arch_test.go`; `git status`
file list identical to the pre-review snapshot; full suite re-run at 776/1/1 with the same P6 stats.
No probe file survives.

### Teeth-verification results (the three the reviewer was asked to measure)

**1 — the option-1 mutant (AC12a). TEETH CONFIRMED, but the recorded blast radius is wrong.**
`ProjectCollection` was really mutated to omit null elements (the owner's rejected arm) and the whole
suite re-run. Result: **771 pass / 6 fail**. `sum` is byte-identical under both arms — **correct, as
ruled**. The owner's decision **is** enforced. But see Major 7: the `avg` assertion never executes.

**3 — the routing assertion (AC8). TEETH CONFIRMED, genuinely.** `evalSum`'s `SumDecimals(decimals)`
call was replaced with a real, compiling, exponent-correct inline `big.Int` accumulator producing
identical numbers. `internal/bind` stayed **fully green (0 failures)** — exactly F6's hazard —
and `TestSumRoutesThroughSumDecimals` **reddened**. This closes D-3.1a.4's follow-up for real.
See Minor 1 for its one undisclosed bound.

**7 — the caveat wire (D-3.3.6). CORRECT, and verified end-to-end by a reviewer-written probe.**
The caveat is a typed value on a third return, never an error, and is not carried on
`bind.Substitution`. The reviewer constructed the exact trap case the lead named: a **page-footer**
element valued `"Page {{page}} of {{pages}} — avg {{avg(t.a)}}"` over `{"t":[]}`. Result: **render
succeeds**, and exactly one Diagnostic survives — `Severity=Warning`, `Code="AGGREGATE_EMPTY_AVERAGE"`,
`ElementID="e2"`, `DataPath="t.a"`. Repeated in the **page-header** band: **one** caveat, not one per
page. `shiftSubstitutions` (`page_number.go:320`) cannot drop it. **The design is right.** See
Major 11: nothing in the shipped suite pins it.

---

### Finding 1: The re-pointed resolution-roots guard LOST coverage it had at `b875582`, and its own written claim is measurably false

- **Severity**: **Blocker**
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/bind/text.go:322-350` (`selectRoot` doc comment);
  `folio-go/internal/bind/resolution_roots_arch_test.go:97-130`, `:160`;
  `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:9048-9062`
- **Observation**: `text.go`'s `selectRoot` doc comment states: *"The purpose the guard exists to
  protect (a THIRD resolution root — AD-4's `page` fence) has one place it can now be introduced,
  and this is it."* **Measured, and false.** The reviewer injected a new dispatch method that
  bypasses `selectRoot` entirely and passes a literal root straight to `lookupBound`:
  ```go
  func (r exprResolver) zzPageScopedLookup(path []string) (expr.Value, error) {
      return lookupBound(r.scope.data, path, path, r.elementID, "page", "the current page")
  }
  ```
  It **compiles**, and `TestBindResolutionRootsAreClosed` **PASSES** (`observed at 3 selectRoot
  return statement(s): [data params row]`). The identical injection applied to a clean `b875582`
  worktree **FAILED** the same-named test: *"OBSERVED resolution root \"page\" (a lookupBound call
  site's rootName) is not in declaredResolutionRoots"*. The re-point did not close the blind spot —
  it **moved** it, and traded away a case that was previously covered. The decision log's
  correction-by-append at `:9048` now asserts permanently that D-3.1.1's weakness paragraph *"is now
  FALSE"* and that *"the guard sees it anyway"*; that record is wrong as shipped.
- **Impact**: This is D-000.38 exactly — advertised coverage that does not exist — and it is the
  precise failure R2 was written to prevent (*"do not land a second guard and amend the first with a
  coverage disclaimer: that trades a known blind spot for advertised-coverage-that-does-not-exist"*).
  The fence in question is **AD-4**, a forever-invariant ("no `page` namespace, ever"), and the false
  claim is now recorded in an append-only decision log where it will be read as settled.
- **Suggested Resolution**: Keep the `selectRoot` scan, and **add a second, cheap assertion in the
  same test**: every `lookupBound` call site's `rootName` argument (index 4) must be a **non-literal**
  — the exact inverse of the pre-3.3 check. A literal there is then either a bypass of `selectRoot`
  or a smuggled root, and both redden. Capture a red-proof using the injection above. Then correct
  the `selectRoot` comment and append a further correction to the decision log. Do **not** simply
  soften the comment — R2 forbids the disclaimer route.
- **Related AC**: AC7 (R2, D-3.3.1, D-3.1.1, D-000.38, D-000.30)

### Finding 2: AC5's stated enforcement does not exist — the method-set instrument reads names only, and an embedded interface is fully invisible

- **Severity**: Major
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/expr_arch_test.go:334-360` (`extractResolverInterfaceMethods`),
  `:368-419` (`TestExprResolverMethodSetIsClosed`)
- **Observation**: AC22's instrument **is** AST set-equality over `expr.Resolver`'s method set
  (`{Resolve, CollectionLength, ProjectCollection}`), not a name list — **item 6 is satisfied as
  ruled**. But two evasions were measured:
  1. Widening the signature to `ProjectCollection(path []string, offset, limit int) ([]Value, error)`
     leaves `TestExprResolverMethodSetIsClosed` **PASSING**. AC5 says: *"Neither method takes a range,
     offset, index or limit — **enforced by AC22's method-set equality assertion**"*. It is not.
     (The red-proof `TestExprResolverMethodSetRedProofFourthMethod` incidentally fails, but with
     *"this red-proof's injection point is stale"* — a message that invites a maintainer to update
     the marker, not to see an AC5 breach.)
  2. Adding a fourth method as an **embedded interface** — `pageScopedResolver` on its own line in
     the `Resolver` declaration — leaves **BOTH** `TestExprResolverMethodSetIsClosed` **and** its own
     red-proof **PASSING, silently**. `extractResolverInterfaceMethods` iterates `m.Names`, which is
     nil for an embedded field.
  D-3.3.1's own "how we'd know it was wrong" names *"a collection method growing a positional
  argument"*; AC22's own text says *"A page-scoped variant then cannot be added under **any**
  spelling without reddening."* Both are falsified.
- **Impact**: An entire page-scoped method set can be attached to `expr.Resolver` under an embedded
  interface with no test reddening anywhere — the AD-11/AD-4 hole AC22 exists to close.
- **Suggested Resolution**: In `extractResolverInterfaceMethods`, treat an `*ast.Field` with
  `len(m.Names) == 0` as a **finding** (an embedded interface is an undeclared method-set widening),
  and additionally compare each method's **parameter and result type spelling** against a declared
  signature string so AC5 is actually enforced. Red-proof both new arms.
- **Related AC**: AC5, AC22 (D-3.3.5, D-3.3.1)

### Finding 3: AC18 ships an argument, not an instrument — the named `noopCounter` does not exist

- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/internal/bind/aggregate_test.go:346-356` (comment),
  `:357-396` (`TestFourCollectionStatesDiscriminated`)
- **Observation**: AC18 requires the absent and wrong-kind arms to go red *"without the kernel being
  called at all (D-000.9). **Proven, not asserted** — e.g. by a kernel call-counter or an equivalent
  instrument the test can observe."* Shipped instead: a comment headed `noopCounter` whose body
  argues *"checked here structurally by the simplest available instrument: the error path returns
  before any []Decimal is ever constructed, which the erroring return signature (nil/0, error)
  already guarantees by construction."* That is reasoning, not an observation. `grep -rn "noopCounter"`
  over the whole module returns **exactly one hit — this comment**. No counter, no spy resolver, no
  instrument of any kind exists, and the comment is attached to a function of a different name.
- **Impact**: AC18's explicit "proven, not asserted" is unmet. The behaviour is in fact correct
  (verified by reading `evalSum`/`evalAvg`: `ProjectCollection`'s error returns before
  `SumDecimals`/`AvgDecimals` is reached), so this is a proof gap, not a defect — but D-000.9's whole
  point is that the all-clear must be *observed* not to share a path with the could-not-look.
- **Suggested Resolution**: Add a `countingResolver` in `internal/expr` wrapping the real seam and
  incrementing a counter inside a test-local `SumDecimals`/`AvgDecimals` shim, or (simpler) assert at
  the `expr` layer with a resolver that returns the absent/wrong-kind error and a package-level test
  hook counting kernel entries. Delete or rewrite the `noopCounter` comment either way.
- **Related AC**: AC18 (R8, D-3.3.4, D-000.9)

### Finding 4: AC10 — the honest float red-proof — is entirely absent, and its impossibility was logged rather than escalated

- **Severity**: Major
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/bind/aggregate_precision_test.go:19-51`
  (`TestSumIsExactOnD00061CorpusA`); story Delivery Log item 7 (`:938-941`)
- **Observation**: AC10 requires **two** things: (a) a mutant that accumulates the decimal **values**
  (`coefficient × 10^exponent`) and re-quantises, under which **both** the value assertion and the
  order-invariance assertion must redden; and (b) a test comment **recording the measurement** that
  the coefficient-level mutant does *not* redden, and why. **Neither ships.** No float mutant exists
  anywhere in the change. The Delivery Log substitutes a citation: *"the discrimination is cited from
  D-000.61's own prior measurement rather than re-measured here"*, on the ground that AD-23 forbids
  `float64` under `internal/`. That ground is sound and probably makes AC10 **unsatisfiable as
  written** — but an unsatisfiable AC is a fence that belongs back with the lead (D-000.60), not a
  line in a delivery log. D-000.61 (extension)'s recorded false alarm is exactly what (b) exists to
  prevent recurring.
- **Impact**: AC9's corpus-A assertion is now a plain exactness check with no demonstration that the
  corpus discriminates at all — the property D-000.61 was written to protect is carried entirely by a
  cross-reference. Losing that measurement re-opens the extension entry's false alarm, which is what
  AC10 says in so many words.
- **Suggested Resolution**: Escalate as a fence: either (i) place the honest mutant in a location the
  float guards do not scan and state the location and why, (ii) restate AC10 as a citation obligation
  with the lead's ruling recorded, or (iii) get AC10 formally retired. Do not leave it as a log line.
- **Related AC**: AC9, AC10 (D-000.61 + extension, AD-23)

### Finding 5: AC25's alignment-spread half is never exercised at the expression layer

- **Severity**: Major
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/bind/aggregate_precision_test.go` (only
  `TestSumOverflowIsLocatedAtExpressionLayer` and
  `TestAvgResultExponentOverflowIsLocatedAtExpressionLayer` exist); the alignment-spread breach is
  covered only at the kernel, `folio-go/internal/expr/reduce_test.go:265-325`
- **Observation**: AC25: *"Same for the **alignment-spread** breach and `avg`'s **result-exponent**
  breach: located at the expression layer, **each** red-proven there."* Only the result-exponent
  breach has an expression-layer test. `grep -rn "alignment spread" internal/bind/` returns nothing.
  AC24 itself warns *"a kernel-layer proof does not prove the wiring"* — that warning applies
  verbatim to the half that was left at the kernel.
- **Impact**: `KernelOverflowError`'s wrapping of the alignment-spread error (element id, collection
  path, operand count, `%w` recoverability) is unproven for that error class. F4 named the
  alignment-spread breach as the **realistic** trigger — one high-precision operand beside ordinary
  amounts — so this is the arm most likely to fire on real caller data.
- **Suggested Resolution**: Add a sibling of `TestSumOverflowIsLocatedAtExpressionLayer` driving an
  alignment-spread breach through `sum(t.a)` (e.g. one operand at `1e-99999` beside an ordinary
  amount), asserting `errors.As(*expr.KernelOverflowError)`, the element id, the path, the operand
  count, and that the kernel's own *"alignment spread"* wording survives inside the wrapper.
- **Related AC**: AC24, AC25, AC26 (F4, R6)

### Finding 6: AC29's replacement re-introduces the hard-coded expected value it exists to remove, and its comment claims the opposite

- **Severity**: Major
- **Category**: Tests / Convention
- **Location**: `folio-go/internal/expr/table_derivational_test.go:133-147`
  (`TestExactlyTwoUnimplementedRemainAfterThisStory`)
- **Observation**: AC29 requires the two name-list tests be restated *"DERIVATIONALLY … So stated,
  they survive **3.3's edit and 3.4's with none**."* `TestImplementedEntriesMatchEvalCallSwitch` does
  satisfy this. Its sibling does not: it hard-codes `if n != 2`. Story 3.4 flipping `formatDate` and
  `formatNumber` to `implemented: true` makes `n == 0` and **this test fails**, forcing 3.4 to edit
  an expected value in the same diff as the thing being guarded — D-3.1a.3's named hazard, verbatim.
  The test's own comment asserts the reverse: *"a count, not a name list, **so 3.4's own edit needs
  no change here either**."* Measurably false. Two further slips in the same block: the doc comment
  names `TestExactlyThree​UnimplementedRemainAfterThisStory`, a function that does not exist; and
  D-2.5.1 — quoted two files away in `resolution_roots_arch_test.go` — says *"never a count, never in
  a test name"*, which `ExactlyTwo` breaches on both counts.
- **Impact**: AC29's stated purpose ("a guard whose expected value must be edited is one that gets
  edited wrongly") is defeated for exactly the story it was written to protect.
- **Suggested Resolution**: Derive it: assert every `implemented == false` entry names an
  `owningStory` **and** has **no** `evalCall` branch (the inverse of the sibling test), which holds
  for any table state. Drop the count, fix the function name in the comment, and delete the false
  survivability claim.
- **Related AC**: AC29 (F11, D-3.1a.3, D-2.5.1)

### Finding 7: AC12a — the mutant's measured blast radius is not the recorded one, and the `avg` assertion never runs

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/internal/bind/aggregate_test.go:213` (the AC4 `t.Fatalf`), `:236` (the
  `avg` assertion), `:245-330` (`TestOption1RedProofProjectCollectionOmitsNulls`); story Delivery
  Log red-proof item 3 (`:926-928`); `folio-mvp-decision-log.md` D-3.3.3 (addendum)
- **Observation**: The reviewer applied the real mutation (`ProjectCollection` omits null elements)
  to `internal/bind/text.go` and re-ran `./... -count=1`. Measured: **771 pass / 6 fail**. What
  reddened, verbatim:
  - `TestAggregateDeclarativeTable/null_is_a_zero_observation` — `aggregate_test.go:213: AC4 LENGTH INVARIANT VIOLATED: len(ProjectCollection)=2 != CollectionLength=3`
  - `TestAggregateDeclarativeTable/all-null,_N=3` — same, `0 != 3`
  - `TestOption1RedProofProjectCollectionOmitsNulls` — `aggregate_test.go:265: presence precondition: the REAL implementation must satisfy AC4 before this red-proof means anything (got 2, want 3)`
  - `TestPerElementDefectsAC19a` — `aggregate_test.go:423: element with explicit null field: got []expr.Value{}, want one KindNull value`

  Three divergences from the record. **(a)** The AC4 check at `:213` is a `t.Fatalf`, which aborts
  the subtest, so the **`avg` assertion at `:236` is never reached**. The Delivery Log claims the
  mutant *"Reddens AC4's length invariant (2 vs 3) **and the `avg` assertion**"*, and D-3.3.3
  (addendum) states the same. It does not. (The reviewer re-ran with `:213` downgraded to `t.Errorf`
  and confirmed the `avg` assertion **does** have teeth: `got {20000,-4}, want {13333,-4}` — so this
  is latent coverage, not absent coverage.) **(b)** The named red-proof test itself reddens only on
  its **own presence precondition** at `:265`; its three "blast radius" checks (sum-identical,
  avg-differs, length-differs) never execute under the mutant. **(c)** A fourth test outside the
  declared radius reddens, contradicting AC12a's *"and nothing else"*. **`sum` correctly does NOT
  redden** — the one thing AC12a most needed, and it holds.
- **Impact**: The owner's null ruling **is** enforced (the suite goes red), so this is not a
  correctness hole. But the story's own "single most load-bearing new assertion" is defended by a
  short-circuited `Fatalf` rather than by the `avg` assertion the ruling says is the only thing that
  can see the difference — and the record asserts an observation that was never made.
- **Suggested Resolution**: Change `:213` to `t.Errorf` so the `avg` assertion runs alongside the
  length invariant (the fixture is still usable after a length mismatch). Correct the Delivery Log's
  red-proof item 3 to the measured radius, including `TestPerElementDefectsAC19a`. Consider having
  `TestOption1RedProofProjectCollectionOmitsNulls` build its `trueProjected` from a literal fixture
  rather than from `ProjectCollection`, so its blast-radius checks survive the mutation they describe.
- **Related AC**: AC4, AC12, AC12a, AC21a (R7, R10, D-3.3.3 + addendum)

### Finding 8: AC27.2 pre-empts Story 3.4's worked `formatNumber` example, and the Delivery Log does not disclose it

- **Severity**: Major
- **Category**: AC Conformance / Maintainability
- **Location**: `docs/expression-reference.md:128-132`; `docs/expression-reference.html:407-409`;
  story Delivery Log docs entry (`:878-880`)
- **Observation**: AC27.2 is explicit: *"**This is 3.3's edit; 3.4's worked `formatNumber` example is
  a different edit and is not pre-empted here.**"* 3.3's required statement **did** land in both
  twins (`.md:123-125`, `.html:400-403`). But the code sample was **also** replaced with 3.4's
  example:
  ```
  {{formatNumber(sum(transactions.amount), "#,##0.00")}}
  {{count(transactions)}}
  {{formatNumber(avg(transactions.amount), "#,##0.00")}}
  ```
  The page now shows a reader how to call a function that does not exist yet (pending markers at
  `.md:126` / `.html:404` mitigate but do not cure this). The Delivery Log lists only the three 3.3
  items and never mentions the sample change.
- **Impact**: The exact merge AC27 was written to prevent — 3.3's edit and 3.4's collapsed into one —
  plus a documented call to an unimplemented function. Story 3.4 now has no distinct edit to make and
  no way to tell what it inherited.
- **Suggested Resolution**: Revert the code sample in both twins to the bare `{{sum(...)}}` /
  `{{count(...)}}` / `{{avg(...)}}` forms marked as errors-in-text, leaving 3.4 to add the worked
  `formatNumber` example. If the finisher judges the example worth keeping, say so explicitly in the
  Delivery Log and get it ruled, since AC27.2 forbids it on its face.
- **Related AC**: AC27.2 (F3, D-000.47)

### Finding 9: The plain-terms opener's "Only the first is a zero" is falsified by the shipped code

- **Severity**: Major
- **Category**: AC Conformance / Maintainability
- **Location**: story `:90-92`; `folio-go/internal/bind/text.go:514-517`, `:566-571`, `:412-425`
- **Observation**: The opener reads: *"Four situations that look alike are now told apart, because
  three would otherwise print a confident zero: a list that exists and is empty; a list the data does
  not contain; a list present but blank; and something that is not a list. **Only the first is a
  zero.**"* Measured against the shipped code: the third situation (an explicit `null` at the
  collection's own path) yields `sum = 0`, `count = 1`, `avg = 0.0000` — **it is a zero**; and the
  first situation's `avg` is a caveat, **not** a zero. Only two of the four are hard errors.
  `text.go:412-425` quotes this exact phrase (*"'a list present but blank' (the plain-language
  section's own phrase)"*) while implementing the opposite of what it says. The story's own R8 table
  (`:427`) agrees with the code, not with the opener.
- **Impact**: The one section written for a non-technical reader misdescribes the money-path
  behaviour, and it did so while the implementation was being written against it. Placement and
  jargon are otherwise clean (verified: no file paths, function names or identifiers in `:80-105`).
- **Suggested Resolution**: The finisher must refresh the opener to describe what actually happened —
  e.g. *"Two of the four are somebody's mistake and stop the document. A blank list counts as a
  single blank entry, so it reads as one observation of zero rather than as no observations at all;
  an empty list gives a zero total and a zero count, but no average at all."* Also drop the framing
  *"because three would otherwise print a confident zero"*.
- **Related AC**: plain-terms opener; R7, R8 (D-3.3.3, D-3.3.4)

### Finding 10: The null-COLLECTION-path semantics (`count` = 1) are undocumented and never asserted through an aggregate

- **Severity**: Major
- **Category**: AC Conformance / Tests / Maintainability
- **Location**: `folio-go/internal/bind/text.go:412-425`, `:514-517`, `:566-571`;
  `folio-go/internal/bind/aggregate_test.go:379-390`; `docs/expression-reference.md`,
  `docs/expression-reference.html`
- **Observation**: For `{"transactions": null}`, `CollectionLength` returns **1** and
  `ProjectCollection` returns one `KindNull` element, so `count(transactions)` reports **1** and
  `sum` reports **0** — one phantom row on a bank statement. This is a defensible reading of R8's
  *"explicit `null` → per R7 — a zero observation"* (singular) and of D-3.3.4's *"zero observations"*,
  and the reviewer is **not** filing the semantics themselves as a defect. What is missing is
  everything around them: (a) it is asserted **only** at the resolver level
  (`TestFourCollectionStatesDiscriminated/explicit null collection`) and **never** through
  `count`/`sum`/`avg` — AC12's declarative table has no null-collection row; (b) **neither doc twin
  says anything about it** — both document a null *field within a row* and an *all-null collection*,
  neither documents a null *collection*; (c) the Delivery Log never records that the developer
  extended R7 "one level up", although `text.go:412-425` says so at length. D-3.3.4's own "In simple
  terms" describes this state as *"the ledger says 'none recorded'"* and groups it with *"three of
  those four are somebody's mistake"* — a tension with its own disposition row that nobody resolved.
- **Impact**: A money-path outcome (`count` = 1 for a null collection) that no aggregate test pins,
  no document describes, and no ruling names at the collection level. A future refactor could flip it
  to 0 or to an Error and only one resolver-level subtest would notice.
- **Suggested Resolution**: Add a fifth row to AC12's declarative table (`{"t":null}` → `sum` 0,
  `count` 1, `avg` `0.0000`, no caveat) so it is asserted through the aggregates; state the case in
  both doc twins; and record in the Delivery Log that this was the developer's extension of R7, with
  its reason — or route it to the lead if the finisher reads D-3.3.4's "In simple terms" as
  contradicting its own table.
- **Related AC**: AC12, AC17, AC27 (R7, R8, D-3.3.3, D-3.3.4)

### Finding 11: D-3.3.6's own falsifier — "a caveat that arrives for body text but not for a page header" — has no test

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/render_empty_average_diagnostic_test.go` (all four tests use content-band
  elements with no page token); `folio-go/page_number.go:308-323` (`shiftSubstitutions`)
- **Observation**: D-3.3.6 states its own failure signal: *"A caveat that arrives for body text but
  not for a page header."* The shipped suite exercises **only** body text. The reviewer wrote the
  missing probe and **the behaviour is correct** (see the teeth-verification note above: a page-footer
  element with `{{page}}`, `{{pages}}` and an empty-collection `{{avg}}` renders successfully and
  emits exactly one `AGGREGATE_EMPTY_AVERAGE` Warning; likewise in the page-header band, once, not
  per page). But nothing in the repository pins it. The Delivery Log claims the trap was avoided *"the
  lead's own trap warning: `shiftSubstitutions` reconstructs that struct field-by-field"* — a design
  claim with no regression test behind it.
- **Impact**: A future refactor folding the caveat onto `bind.Substitution` would pass the **entire**
  shipped suite while silently dropping every diagnostic from page headers and footers — the exact
  bug D-3.3.6 was written to prevent, in the exact place it said to look.
- **Suggested Resolution**: Add the probe to `render_empty_average_diagnostic_test.go`: one
  `pageFooter` element valued `"Page {{page}} of {{pages}} — avg {{avg(t.a)}}"` over `{"t":[]}`,
  asserting the render succeeds and that exactly one `DiagCodeEmptyAverage` Warning with the right
  `ElementID`/`DataPath` survives. A header-band twin asserting a count of exactly 1 across a
  multi-page document is worth the extra ten lines.
- **Related AC**: DECISION-5, R9 (D-3.3.6)

### Finding 12: The routing assertion checks that a call is *present*, not that its result is *used* — and the bound is undisclosed

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/expr/routing_arch_test.go:38-70` (`funcCallsName`), `:72-84`;
  `folio-go/internal/expr/aggregate.go:11-16` (file header)
- **Observation**: Measured: inserting `_, _ = SumDecimals(decimals)` as a discarded vestigial call
  alongside an inline `big.Int` accumulator leaves `TestSumRoutesThroughSumDecimals` **PASSING** while
  `sum()` is computed by a second accumulator. `aggregate.go`'s header claims the AST scan *"is what
  makes the calls below an OBLIGATION, not a convention"*. It makes the *presence of a call* an
  obligation.
- **Impact**: Low — the bypass is deliberate, not accidental, and the assertion has real teeth
  against the realistic hazard (verified). But D-000.24 requires a bounded guard to name its bound,
  and this one does not.
- **Suggested Resolution**: Either tighten the scan to require the call's result to flow into the
  returned value, or (cheaper and honest) state the bound in the test's own comment: *"this asserts
  the call is present, not that its result is used; a deliberately vestigial call defeats it."*
- **Related AC**: AC8, AC14 (R4, D-000.59, D-000.24, F6)

### Finding 13: `TestSumRoutingRedProofInlineAccumulator`'s mutant is not what its comment claims

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/expr/routing_arch_test.go:118-133`
- **Observation**: The comment states the injected replacement is *"a SYNTACTICALLY VALID, honest
  inline `big.Int` accumulator … it would, if actually compiled and run, compute the SAME correct
  answer as `SumDecimals` for ordinary operands"*. Both claims are false. The injection uses
  `bigInt` and `bigIntFromInt64`, identifiers that **do not exist** in package `expr` (the real
  spellings are `big.Int` / `big.NewInt`), so it would not compile; and it ignores `d.Exponent`
  entirely, so it would **not** compute the same answer for any mixed-exponent operand set — including
  AC12's own `10.00 / 20.00 / 30.00` corpus. D-000.61 (extension)'s *"the mutant must be HONEST"* is
  the discipline the comment invokes and fails.
- **Impact**: Record-only. The **assertion** genuinely has teeth: the reviewer built a real,
  compiling, exponent-correct inline accumulator that left `internal/bind` fully green, and
  `TestSumRoutesThroughSumDecimals` reddened. So no coverage is lost — but the captured red-proof
  documents a weaker mutant than the one that matters.
- **Suggested Resolution**: Fix the injected text to a genuinely compiling, exponent-aware
  accumulator (align to the minimum exponent with `new(big.Int).Exp`), or delete the two false
  clauses from the comment and say plainly that the injection is a textual removal of the routing
  call.
- **Related AC**: AC8 (D-000.61 extension, D-000.30)

### Finding 14: AC14's captured red-proof for `avg` routing was never landed

- **Severity**: Minor
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/expr/routing_arch_test.go` (contains
  `TestSumRoutingRedProofInlineAccumulator` only); story Delivery Log red-proof item 2 (`:922-925`)
- **Observation**: AC14: *"`avg` routes to `AvgDecimals` with the **same positive routing assertion
  and captured red-proof shape as AC8**."* The positive assertion (`TestAvgRoutesThroughAvgDecimals`)
  ships; there is no `avg` counterpart to the captured red-proof, and the Delivery Log's red-proof
  list names only the `sum` one.
- **Impact**: Low — the two positive assertions share `funcCallsName`, so the `avg` assertion's teeth
  are inherited from the `sum` red-proof. But AC14 asked for both, and DW-7 holds Story 4.5 to this
  shape for `avg` footers as much as `sum` ones.
- **Suggested Resolution**: Add the four-line mirror of `TestSumRoutingRedProofInlineAccumulator` for
  `evalAvg` / `AvgDecimals`, or state in the Delivery Log that the `avg` red-proof is deliberately
  inherited from the `sum` one because both assertions share one helper.
- **Related AC**: AC14 (R4)

### Finding 15: AC7.3's red-proof does not match its own description

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/bind/resolution_roots_arch_test.go:229-243` (comment),
  `:285-300` (`injectThirdResolutionRoot`)
- **Observation**: AC7.3 required *"add a fourth literal root **reachable only from the collection
  path**, watch the re-pointed guard redden, remove it."* The shipped injection appends a **duplicate
  `selectRoot` declaration** returning `"page"`. That is reachable from nothing (it does not compile),
  and is in no way specific to the collection path. The test comment nonetheless asserts *"The
  injected literal is reachable ONLY through a path a collection-method caller would take"*.
- **Impact**: The guard does redden under the injection, so the red-proof is not worthless — but it
  proves the scan sees a second `selectRoot`, not that the collection path is covered. Combined with
  Finding 1, the specific claim that the collection dispatch is now covered rests on nothing measured.
- **Suggested Resolution**: Correct the comment to describe what the injection actually is, and
  consider an injection that adds a fourth `return "page", …` **inside** the real `selectRoot` under
  the `!allowRow` branch — genuinely collection-path-only, and still parse-only.
- **Related AC**: AC7.3 (R2, D-000.30)

### Finding 16: `TestAvgRoundHalfToEvenAtTheSeam` cites a red-proof that does not prove what it says

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/bind/aggregate_precision_test.go:98-104` (comment)
- **Observation**: The comment reads: *"Red-proof: `TestAvgRoutesThroughAvgDecimals`
  (routing_arch_test.go) already proves `avg()` cannot substitute a different rounding rule without
  reddening."* That test only checks that `evalAvg`'s body textually contains a call to
  `AvgDecimals`; it says nothing about rounding, and per Finding 12 a vestigial call satisfies it.
  AC16 asked for *"replacing it with round-half-away-from-zero reddens the table"*. The Delivery Log
  (item 8) discloses honestly that the red-proof is **inherited from the kernel**; the test comment
  makes a stronger, false claim than the log does. (The tie fixtures themselves were verified to be
  genuine exact ties: `0.01/32 = 0.0003125` → `312` even; `0.03/32 = 0.0009375` → `938` even.)
- **Impact**: Record-only; the rounding coverage is real via the kernel's own red-proof.
- **Suggested Resolution**: Replace the sentence with the Delivery Log's own honest wording — the
  round-half-away red-proof lives at the kernel and is not re-derived here — and drop the claim that
  the routing test proves anything about rounding.
- **Related AC**: AC16

### Finding 17: AC19/D-3.3.2's first-failure declaration is missing from the Delivery Log, and the code comment mis-cites it

- **Severity**: Minor
- **Category**: AC Conformance / Maintainability
- **Location**: `folio-go/internal/bind/text.go:531-546`; story Delivery Log (`:814-985`)
- **Observation**: AC19's guardrail: *"do **not** report only the first defective element — **collect
  and report**, or **state explicitly in the delivery log that it is first-failure and why**."*
  D-3.3.2 repeats it. `text.go`'s `ProjectCollection` comment does state it, and says *"this file
  states first-failure explicitly, **in the delivery log**, because a single-error return is the shape
  every other located error in this package already uses"*. But `grep -n "first-failure"` over the
  story file returns **one hit — AC19's own text at `:589`**. The Delivery Log says nothing.
- **Impact**: The chosen arm of a two-arm guardrail is declared only in a code comment that cites a
  record entry that does not exist.
- **Suggested Resolution**: Add one line to the Delivery Log stating first-failure and the reason,
  and fix the `text.go` comment's self-citation.
- **Related AC**: AC19 (D-3.3.2)

### Finding 18: The reference page hedges the declared `avg` scale into near-invisibility

- **Severity**: Minor
- **Category**: AC Conformance / Maintainability
- **Location**: `docs/expression-reference.md:137-141`; `docs/expression-reference.html:416-420`
- **Observation**: AC15 requires the declared scale — *"(maximum operand scale + `avgExtraScale`)"* —
  to be *"Named in the AC test and **on the reference page**"*. Both twins say: *"`avg` divides at a
  scale beyond every operand's own precision (four extra digits, illustrative — **the exact figure is
  not part of this contract**)"*. Hedging the **number** is right (FLAG-1 keeps `4` at MEDIUM
  confidence for Story 3.4), but the **rule** — that the scale is the maximum operand scale plus a
  fixed extra — is never stated in those terms, and the same section then refers to an all-null
  average as *"(zero, at **the declared scale**)"*, a phrase the page has just told the reader is not
  part of the contract.
- **Impact**: A reader cannot predict `avg`'s output scale from the page, which is what AC15 asks for.
- **Suggested Resolution**: State the rule and hedge only the constant: *"`avg` divides at the
  greatest number of decimal places any operand carries, plus a fixed number of extra digits (four
  today; the constant is illustrative, the rule is not)."* Resolve the "declared scale" phrasing to
  match.
- **Related AC**: AC15, AC27.3 (FLAG-1)

### Finding 19: A fifth instance of F12's stale-citation class survives

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/testdata/lint/no-bigfloat-type/compliant.go:11`
- **Observation**: Still reads *"…blesses (internal/bind's SumDecimals/AvgDecimals)…"*. `SumDecimals`
  and `AvgDecimals` live in `internal/expr` (D-3.2.1). AC31 names only F12's four sites and all four
  are fixed, so AC31 is met — but the class is not closed, and F12's framing (*"Four stale comments —
  every one names a package that no longer declares the symbol. **Verified.**"*) reads as an
  exhaustive enumeration that was not.
- **Impact**: Trivial; a lint-rule fixture comment.
- **Suggested Resolution**: One-word fix in the same commit, and drop "Verified" from F12's
  exhaustiveness claim, or note the fifth site.
- **Related AC**: AC31 (F12)

### Finding 20: D-3.3.5's heading says DW-7 is "amended in place" while its body and AC23 require an append

- **Severity**: Nit
- **Category**: Convention
- **Location**: `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:9959`
- **Observation**: The newly-appended ruling's heading reads *"DW-7 is amended **in place**"*, while
  its own body and AC23 both require an **append**. The shipped `deferred-work.md` edit is a correct
  append (20 added lines, 0 deleted, self-declared at `:273`, original *"none possible"* paragraph
  intact at `:263-267`, correction at `:288-291`, Story 4.5 named three times, ownership unchanged) —
  so this is a wording slip in the ruling, not a discipline breach.
- **Impact**: A future reader could cite the heading as licence to edit DW-7 in place.
- **Suggested Resolution**: Append a one-line correction to the decision log (never edit in place).
- **Related AC**: AC23 (D-3.3.5)

### Finding 21: D-3.3.6 binding 2 requires the caveat element-type choice to be "recorded with its reason"; only the choice is recorded

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: story Delivery Log "DECISION-5 — how it resolved" (`:816-829`);
  `folio-go/internal/expr/caveat.go:1-25`
- **Observation**: D-3.3.6 binding 2: *"Whether the slice element type lives in `expr` or `bind` is
  the developer's call, **recorded with its reason**."* The Delivery Log records the choice (`expr`)
  but not the reason. `caveat.go`'s doc comment does give the reason well, so nothing is lost — the
  record just does not carry it where the ruling asked.
- **Impact**: Negligible.
- **Suggested Resolution**: One clause in the Delivery Log, or a pointer to `caveat.go`'s comment.
- **Related AC**: DECISION-5 (D-3.3.6 binding 2)

---

### Explicitly checked and found SATISFIED (no finding filed)

- **AC1, AC2, AC3, AC6** — the two-method seam, data-order preservation on an order-distinguishable
  fixture, `count`'s structural inability to reach a projected value, and the bare-collection-path
  error at evaluation. `internal/expr` does not import `internal/bind`; `expr.Value` stays scalar.
- **AC4** — the length invariant is asserted on **every** row of AC12's table, including all-null,
  and it has measured teeth (Finding 7 concerns the recorded radius, not the invariant).
- **AC7.1, AC7.2, AC7.4, AC7.5** — precedence `params` → row alias → data preserved in `selectRoot`;
  root names stay literals `{"data","params","row"}`; no fourth root exists; the closure red-proof
  runs green against the re-pointed guard; D-3.1.1's paragraph is corrected **by append** with the
  original bytes intact (verified: the decision-log diff contains **zero** deletion lines).
- **AC8, AC9** — routing verified with real teeth (see teeth-verification 3); corpus A's exact total
  `12345678901234.88` and order-invariance over original/reversed/shuffled all assert as ruled.
- **AC11** — no `float32`, `float64`, `math/big.Float` or `math/big.Rat` introduced anywhere in the
  change; the only `float64` occurrences are four words inside comments in
  `aggregate_precision_test.go`. No `math/rand`, no `time`/`os`/`net` under `internal/`, no new
  dependency (`go.mod`/`go.sum` untouched). `go vet` clean in both modules.
- **AC12** (four subjects), **AC13**, **AC15**, **AC16** — the declarative table asserts computed
  literals for all four subjects; the two differently-shaped zeros (`0` exponent 0 for the empty sum,
  `0.0000` for the all-null avg) are **correct and were not filed**; `[null,null,null]` and `[0,0,0]`
  being indistinguishable is **the owner's decision and was not filed**; `count` never inspects the
  projected field, with R5's sentence carried verbatim; the scale is re-declared nowhere; the tie
  table's four cases are genuine exact ties.
- **AC17, AC19, AC19a** — all four collection states are discriminated **before** the kernel, and the
  absent/wrong-kind arms return without constructing a `[]Decimal` at all; the single
  null-plus-absent fixture ships as ruled, asserting a `KindNull` value for the first and a located
  Error naming index, field, element id and collection path for the second.
- **AC20, AC21, AC21a** — the shadowing-alias fixture proves root-relative resolution through
  `allowRow=false`; the windowed resolver produces `30.00` against a stated whole of `100.00`; the
  honesty note that this mutant is **not** independent evidence for AC4 is present and measured.
- **AC22** — the instrument is AST set-equality over the method set, exactly as D-3.3.5 ruled, **not**
  a forbidden-name list. (Finding 2 concerns the extractor's blind spots, not the instrument choice.)
- **AC23** — DW-7 appended correctly; the *"none possible"* → *"none possible for the footer half"*
  correction lands verbatim; Story 4.5 named; ownership unchanged; FLAG-4 restated honestly.
- **AC24, AC26** — `KernelOverflowError` carries element id, path and operand count; `Unwrap` is
  implemented; `errors.As`/`errors.Is` recoverability is asserted, and the kernel's own wording is
  asserted to survive inside the wrapper. Story 3.6 can attach codes without re-plumbing.
- **AC27.1, AC27.4, AC27.5** — Totals heading and status row updated in both twins; null-as-zero-
  observation stated in both; `avg`-on-empty described as a survivable caveat in both; **`b875582`'s
  "diagnostic" → "error" fix is NOT undone** (the word "diagnostic" appears nowhere on either page,
  and "error" is retained exactly where an error is meant).
- **AC28, AC30, AC31** — the three table entries flip; the four closed-table guards stay green; the
  unimplemented table drops `sum`/`count`/`avg` in the **same** change that adds the positives, and
  `formatDate` correctly replaces `sum` as the unimplemented example at `eval_test.go:348`; F12's four
  stale comments and three wrong line cites are fixed, preferring symbol references over line cites,
  with no functional edit to `reducer_inventory_arch_test.go`.
- **AC32** — both modules gated and reported with scope and flags; figures reproduced exactly.
- **DECISION-5 / D-000.65** — `DiagCodeEmptyAverage` is minted in `folio/diagnostic.go` where the
  `Diagnostic` is constructed, in the story where the condition first ships. Compliant.
- **sprint-status.yaml** — exactly two lines changed (`last_updated`, and this story → `review`).
  No other key touched.
- **Plain-terms opener** — present immediately before `## Story`, and jargon-free: no file paths, no
  function names, no identifiers. (Finding 9 concerns one factual claim inside it.)

### Could not verify

- **Sequencing (item 11 / D-000.30 / D-3.3.6's binding sequencing note).** The Delivery Log states
  the `selectRoot` extraction and its red-proof landed **before** the caveat return was added. The
  entire story is one uncommitted working tree with no intermediate commits, so **edit order is
  unfalsifiable from the repository**. What *can* be confirmed is that both red-proofs were
  **captured as permanent tests** rather than performed once and discarded, which is what D-000.30
  materially asks for. Recorded as unverified, not as a finding.
- **D-000.61's original float64 discrimination measurement for corpus A** was not re-derived (see
  Finding 4); the reviewer accepted the citation rather than introducing `float64` under `internal/`.
- The **cross-target hash matrix** was correctly not run (per-epic cadence, D-000.4); this story is
  not hash-shaped.

## Finding Resolutions (finisher)

| # | Severity | Decision | Files touched |
|---|---|---|---|
| 1 | Blocker | FIX | `internal/bind/text.go`, `internal/bind/resolution_roots_arch_test.go`, `internal/bind/scope.go` |
| 2 | Major | FIX | `lint/internal/rules/resolvermethodset.go`, `resolvermethodset_test.go`, `folio-go/testdata/lint/resolver-method-set/**` |
| 3 | Major | FIX | `internal/bind/aggregate_test.go`, `internal/expr/routing_arch_test.go` |
| 4 | Major | FIX (via engineering-lead ruling [[D-3.3.7]]) | `internal/bind/aggregate_precision_test.go`, `hashmatrix/floatdiscrimination/floatdiscrimination_test.go`, `.github/workflows/ci.yml`, decision log |
| 5 | Major | FIX | `internal/bind/aggregate_precision_test.go` |
| 6 | Major | FIX | `internal/expr/table_derivational_test.go` |
| 7 | Major | FIX | `internal/bind/aggregate_test.go`, decision log (D-3.3.3 addendum correction, below) |
| 8 | Major | FIX | `docs/expression-reference.md`, `docs/expression-reference.html` |
| 9 | Major | FIX | story plain-terms opener, `internal/bind/text.go` |
| 10 | Major | FIX | `internal/bind/aggregate_test.go`, both doc twins |
| 11 | Major | FIX | `render_empty_average_diagnostic_test.go` |
| 12 | Minor | FIX (bound stated, not built against, per ruling) | `internal/expr/aggregate.go` |
| 13 | Minor | FIX | `internal/expr/routing_arch_test.go` |
| 14 | Minor | FIX | `internal/expr/routing_arch_test.go` |
| 15 | Minor | FIX (superseded by Finding 1's rewrite) | `internal/bind/resolution_roots_arch_test.go` |
| 16 | Minor | FIX | `internal/bind/aggregate_precision_test.go` |
| 17 | Minor | FIX | story Delivery Log, `internal/bind/text.go` |
| 18 | Minor | FIX | both doc twins |
| 19 | Nit | FIX | `folio-go/testdata/lint/no-bigfloat-type/compliant.go` |
| 20 | Nit | FIX (append, per D-000.29) | decision log ([[D-3.3.9]]) |
| 21 | Nit | FIX | story Delivery Log (DECISION-5 section) |

**Finding 1 (Blocker).** FIX, per the engineering lead's explicit ruling (quoted in the finisher's
brief): `rootName`/`rootDesc` become one defined type, `rootKind`, closing "can a root be introduced
outside a declaration" by construction (a compile-time property, demonstrated with the compiler's
verbatim error, never dressed as a test) and re-pointing the closed-set scan to AST set-equality over
every `rootKind` composite literal in the package — strictly wider than the `selectRoot`-return-
statement proxy the review defeated. Recorded at length in [[D-3.3.8]] (append-only correction to
[[D-3.1.1]]/[[D-3.3.1]]).

**Finding 2 (Major).** FIX. The engineering lead ruled the AC5/AC22 enforcement belongs in `lint`,
using `go/types`' method-set expansion (which sees an embedded interface's contribution by
definition) and full signature comparison (which sees a widened parameter list). Landed as a new rule
with both evasions red-proven against dedicated fixture trees.

**Finding 3 (Major).** FIX. The `noopCounter` comment named an instrument that never existed. Replaced
with a real structural instrument at the layer where the property is actually decided
(`internal/expr`'s `evalSum`/`evalAvg`), proving the kernel call is textually unreachable without
first passing an error-check-and-return on `ProjectCollection`'s result, with a captured red-proof.

**Finding 4 (Major).** FIX, via an engineering-lead ruling ([[D-3.3.7]]) rather than a unilateral
call: the developer's premise (AD-23 forbids AC10 outright) was half right — AD-23 forbids a
COMMITTED float, not an APPLIED one. The measurement was performed live in this pass (mutated,
measured, reverted, byte-identical restore confirmed) and the durable demonstration was pinned in
`hashmatrix/floatdiscrimination/`, outside AD-23's scope by construction. AC10's own text is corrected
to state the true, order-dependent shape of the required demonstration rather than the "both
assertions redden unconditionally" claim the review found false.

**Finding 5 (Major).** FIX. `TestSumAlignmentSpreadOverflowIsLocatedAtExpressionLayer` added, mirroring
the existing overflow/result-exponent expression-layer tests for the one arm F4 named as the realistic
trigger and AC25 required but the story never landed.

**Finding 6 (Major).** FIX. The hard-coded `if n != 2` — reintroducing, inside the very guard AC29
exists to make edit-proof, the hazard D-3.1a.3 names — is replaced by a derivational inverse of the
sibling test, which holds for any table state.

**Finding 7 (Major).** FIX. `t.Fatalf` → `t.Errorf` at the AC4 length-invariant check, so the `avg`
assertion the null ruling's independent teeth actually depend on runs alongside it. Re-measured live:
the option-1 mutation now reddens `AC4 LENGTH INVARIANT VIOLATED: len(ProjectCollection)=2 !=
CollectionLength=3` AND `avg: got {Coefficient:20000 Exponent:-4}, want {Coefficient:13333
Exponent:-4}` in the same subtest run — the exact evidence the record previously claimed without
having produced it. The D-3.3.3 addendum's blast-radius overclaim is corrected via [[D-3.3.8]]'s
sibling reasoning being the wrong entry to touch — recorded instead as a plain statement in this
Delivery Log's Finisher pass subsection (D-3.3.3 itself is an OWNER DECISION entry and is not amended
for a measurement correction that belongs to the story's own record, not the ruling).

**Finding 8 (Major).** FIX. Both doc twins' code samples reverted to the bare, error-marked forms
AC27.2 requires; the `formatNumber` worked example is left for Story 3.4 to add, with a note saying so
explicitly rather than silently.

**Finding 9 (Major).** FIX. The plain-terms opener is rewritten to state what actually shipped: two of
the four states are errors, two are zeros — not "only the first is a zero." The identical false claim
in `text.go`'s own file-level comment (which contradicted itself one paragraph later) is corrected in
the same commit.

**Finding 10 (Major).** FIX. A fifth declarative-table row asserts the null-collection-path semantics
through `sum`/`count`/`avg` rather than only at the resolver level; both doc twins now state the rule.

**Finding 11 (Major).** FIX. Two new tests build a genuinely multi-page document (reusing this
codebase's own measured 3-page geometry) with the reviewer's exact page-footer trap case plus a
page-header twin, each asserting exactly one Warning survives across every page.

**Finding 12 (Minor).** FIX, per the ruling: the vestigial-call bound is stated in the assertion's own
text (D-000.24), not built against with a disproportionate call-graph instrument (D-3.1a.4's own
precedent, applied the same way here).

**Finding 13 (Minor).** FIX. The dishonest-mutant comment's two false claims are removed; the
red-proof is now honestly described as a textual routing-removal, with the stronger claim
independently re-verified by hand (a real compiling, exponent-aware accumulator was measured to also
redden `TestSumRoutesThroughSumDecimals`, then reverted) rather than asserted without evidence.

**Finding 14 (Minor).** FIX. `TestAvgRoutingRedProofInlineAccumulator` added, mirroring the sum
red-proof exactly as AC14 requires.

**Finding 15 (Minor).** FIX, superseded by Finding 1's rewrite: the re-pointed guard's red-proof is now
keyed on the same `rootKind` composite-literal mechanism the closure test itself uses, so the
"reachable only through a collection-path caller" claim the review found unsupported no longer
appears — the new red-proof makes no reachability claim at all, matching what it actually
demonstrates.

**Finding 16 (Minor).** FIX. The comment's false claim that the routing assertion proves anything about
rounding is removed and replaced with the Delivery Log's own honest wording: the rounding red-proof is
inherited from the kernel's own test, not re-derived at the seam.

**Finding 17 (Minor).** FIX. The first-failure declaration now actually lives in the Delivery Log (see
the Finisher pass subsection), and `text.go`'s self-citation is corrected to point at a record that
exists.

**Finding 18 (Minor).** FIX. Both doc twins now state the scale RULE (max operand scale plus a fixed
extra) and hedge only the CONSTANT, resolving the "declared scale" phrasing used for the all-null
example to match.

**Finding 19 (Nit).** FIX. The fifth stale `internal/bind` citation is corrected to `internal/expr`.

**Finding 20 (Nit).** FIX, by append ([[D-3.3.9]]): [[D-3.3.5]]'s heading said "amended in place" while
its body and AC23 required — and the shipped code correctly performed — an append; the wording slip is
corrected without touching the entry itself, and the shipped `deferred-work.md` change needed no fix.

**Finding 21 (Nit).** FIX. The reason for keeping `Caveat` in `internal/expr` (already present in
`caveat.go`'s own doc comment) is now also stated in this story's Delivery Log, where D-3.3.6 binding 2
asked for it.
