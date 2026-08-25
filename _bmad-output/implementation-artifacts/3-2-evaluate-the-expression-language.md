---
baseline_commit: 10327a4772d605490db2f277e1e5019436bb0052
---

# Story 3.2: Evaluate the expression language

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-2-evaluate-the-expression-language`
**Status:** `done`
**Covers:** **FR18** · **AD-9** · counter-metric **C1**
**Primary invariant:** **AD-9** (the expression parser is hand-written recursive descent — no
generator, no dependency; the eight functions are a closed table so **C1** is a diff away from
visible — ARCHITECTURE-SPINE.md:501).
**Adjacent invariants:** AD-1 (import ban), AD-2 (geom imports nothing), AD-13 (one source of
truth), AD-14 (located diagnostics), AD-23 (exact decimals, no binary float), AD-4 (no `page`
namespace).
**Governing rulings:** **D-000.59 (the discharge charter)** · **D-3.2.1 (the `Decimal` move + the
forcing function)** · **D-3.2.2 (CEL and every general-purpose expression library REJECTED)** ·
**D-3.1a.3 (the relational tripwire that must survive UNEDITED)** ·
**D-1.6.5 (this story's parser REPLACES 1.6's matcher)** · D-1.4.1 · D-1.4.2 · D-1.4.3 · D-1.6.1 ·
D-000.9 · D-000.13 · D-000.14 · D-000.15 · D-000.21 · D-000.23 · D-000.24 · D-000.25 · D-000.26 ·
D-000.29 · D-000.30 · D-000.34 · D-000.37 · D-000.38 · D-000.42 · D-000.50 · D-000.52 · D-000.60 ·
D-000.61 · D-3.1a.2 · D-3.1a.4 · D-2.8.4
**Retires DW-5 and DW-8. Adds none.**

---

## Baseline, measured in this run at creation

Gates were measured at HEAD **`fde96b5`** — *"Record D-000.61 (extension): introducing a float is not
the same as introducing float error"* — branch `main`, **working tree clean** (`git status
--porcelain` empty, verified before and after every measurement below).

**HEAD advanced to `10327a4`** *("Record D-3.2.2: CEL rejected, on its numeric model rather than its
parser")* **during this story's creation.** That commit touches **`folio-mvp-decision-log.md` only**
— no `.go` file, no fixture — verified with `git show --name-only`. **Every figure below therefore
still stands at `10327a4` unchanged.** D-3.2.2 is folded into the rulings above and into item 8 of
*Do not re-open*; it is binding on this story and it landed after the ACs were drafted, so the
developer should read it in full rather than trust the summary.

Stated with scope and flags (D-000.26 — a carried gate figure without its scope is not a figure):

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting **every** `--- PASS`/`--- FAIL`/`--- SKIP` occurrence, subtests included | **640 PASS · 1 FAIL · 1 SKIP** |
| `lint/` | the same invocation, **working checkout** | **89 PASS · 3 FAIL** |
| `lint/` | the same invocation, **clean worktree** | **92 PASS · 0 FAIL** (carried from 3.1a; not re-measured this run) |

The single `folio-go` FAIL is **`TestCorpusMeetsP6ExerciseFloors`** — **REQUIRED red**, stats
byte-identical `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. **Never "fix" it.**

The three `lint` FAILs are **known-environmental (DW-19)**: `TestManifestUpToDate`,
`TestResolveAssetsIncludesWordlist`, `TestFontsAssetsNoticeRemovalRedProof`. The asset resolver
walks the **gitignored** `.font-sources/` directory, so a working checkout carrying it diverges from
a clean one. 89 + 3 = 92, which reconciles exactly with the clean figure. **Any lint number this
story reports must name these three, or it is not a number.**

**Cadence (D-000.4):** unit / vet / build / lint every story. The cross-target hash matrix is due at
**Epic 3's close**, not here — 3.2 introduces no new source of cross-target divergence (a parser over
`[]byte` producing `Decimal`, with float already banned twice over).

---

## In plain terms (read this first if you just want the gist)

A template can now say more than "put the customer's name here": it can turn text upper or lower
case, and choose between two values based on a condition — the three operations this story actually
teaches. Five more are named and reserved for later; calling one today gives a clear complaint naming
which one it was, never a blank or a silently wrong answer. The set of operations is fixed at exactly
eight, permanently, with a mechanical check that stops the build the moment a ninth is added. An older
way of reading the same placeholder syntax was retired outright, and the decimal-number type it
depended on now lives in one place instead of two.

Independent review earned its keep, and two findings deserve stating plainly, not embarrassments but
the most instructive facts here. First, a safety check meant to prove a computed total
never gets accidentally written back into the saved document was comparing a document against itself
rather than anything real — it could not catch the mistake it existed to catch. It now inspects the
actual document the computation touched, proven by reintroducing the write-back and watching the
check catch it. Second, and more serious: one unusually shaped instruction — nesting one calculation
deep inside another, the kind of thing a designer could write by accident — could crash the whole
program outright, with no error and no recovery, because nothing stopped the calculator working too
hard on itself. That is now bounded to a generous but firm limit; past it, input is rejected with an
ordinary complaint instead of taking the program down.

Review turned up two dozen further issues, almost all small: a few diagnostics named the wrong
character or omitted where a mistake was, one visibility setting could not previously be checked
before the document loaded, and a handful of tests proved less than they claimed. All of substance
are fixed and proven; two are documentation-only, and two belong to someone else's earlier fix, not
this story. One test is supposed to keep failing and does — a deliberate marker, not a regression.

---

## Story

**As a** template author,
**I want** a small set of expressions inside my bindings,
**So that** I can compute a value without the template becoming a program.

---

## Do not re-open — settled rulings this story inherits

Reproduced with rationale so the developer does not re-litigate them.

1. [x] **The absence tripwire is discharged by REPLACEMENT, never deletion (D-000.59).** In the
   **same commit**: (1) delete `absence-expr-package`, (2) land a positive assertion of the
   obligation, (3) demonstrate that assertion **failing with the obligation unmet**. A commit doing
   (1) without (2) and (3) has **retired** a forcing function, not discharged it.

2. [x] **THE ANTI-VACUITY CLAUSE IS THE LOAD-BEARING HALF.** The replacement asserts the
   **OBLIGATION**, never the **EVENT**. *"`internal/expr` exists"* is the EVENT and is **vacuous** —
   deleting the rule already proved it. Anything of that shape is a review blocker on sight.

3. [x] **THE ORDERING TRAP (D-000.30).** Capture the red-proof for **each** new assertion **BEFORE**
   the obligation is wired. Wiring closes the window **permanently**. All three windows are open
   right now, at `fde96b5`, and every one of them shuts the moment the move lands.

4. [x] **`Decimal` MOVES; it is never DUPLICATED (D-3.2.1, D-1.6.1, DW-8).** The rank guard
   (`lint/internal/rules/stagerank.go:62-63`, `expr`=3, `bind`=4) makes `expr → bind` a violation, so
   duplication is the only other way out — and **D-1.6.1 wrote itself to prevent exactly that.**
   Destination is **forced, not chosen**: `internal/expr` (3) may import `template` (2) for the
   module's one `SplitJSONNumber`; `bind` (4) may import `expr` back. `geom` is excluded (AD-2:
   geometry only, imports nothing).

5. [x] **The reducer inventory tripwire is RELATIONAL and must survive this move UNEDITED
   (D-3.1a.3).** `internal/reducer_inventory_arch_test.go` asserts the reducers live *in the same
   package as `Decimal`'s type declaration*. When `Decimal` moves, that guard should follow with
   **zero edits**. **If this story finds itself editing that guard's expected value, something is
   wrong — say so, do not edit it.**

6. [x] **D-3.1a.4's correction stands: the inventory does NOT force routing.** It prevents a second
   differently-named top-level reducer. It cannot see an inline accumulator written inside a future
   `sum()`. Do not restate the over-claim; do not build call-graph enforcement here.

7. [x] **This story's parser REPLACES 1.6's matcher — deleted, not kept alongside (D-1.6.5, DW-8).**
   Story 3.1 was explicitly forbidden from touching `parseBindingPath`/`isValidIdent` and left them
   SHA-256 identical **precisely so this story could delete them cleanly.**

8. [x] **AD-9: hand-written recursive descent.** No generator, no third-party dependency, no
   regex-engine substitute. `internal/template` remains the sole owner of the `.folio` parser and
   serializer; **no JSON Schema library** (a second source of truth against AD-9).

   **D-3.2.2 (owner decision, landed at `10327a4`) settles this against the strongest specific
   candidate and must not be re-derived.** CEL — and **any** general-purpose expression library — is
   **REJECTED**, and *"the reason is the numeric model, not the parser"*: **CEL has no decimal type**
   (`int`/`uint`/`double`), so any fractional literal or arithmetic goes through **binary floating
   point** — reintroducing **as a dependency** the precise defect AD-23 was written to prevent, on a
   bank-statement product. It would trip **Story 3.1a's own lint rule** by resolved type identity, and
   Story 3.3's round-half-to-even at a declared scale is **inexpressible in CEL**. The steelman
   (custom opaque decimal type + `NewCustomEnv` with the standard library omitted) is recorded and
   already rejected: *"then CEL supplies only the PARSER while we write all the arithmetic anyway"* —
   plus ANTLR and protobuf on the **js/wasm** target, against a payload budget already argued
   (D-2.2.4), plus a new vendor boundary in the byte-reproducible core (D-000.25). **What would
   reopen it:** CEL gaining a first-class exact-decimal type **in its own type system** — not a
   custom extension we register. Nothing else. A proposal to adopt any expression library is answered
   with D-3.2.2, not re-argued.

9. [x] **`{{page}}`/`{{pages}}` stay reserved and untouched (AD-4, D-1.6.5).** They are late-bound
   slots recognised by the template layer, never resolved from data, never an error. The new parser
   must preserve `reservedPlaceholders`' short-circuit at `internal/bind/text.go:180-188` **ahead of**
   any parse attempt. AD-4 is absolute: no `page` namespace exists for expressions to reach.

10. [x] **`SumDecimals(nil)` returns the identity `{0,0}`; `AvgDecimals(nil)` errors (D-3.1a.2).**
    The asymmetry is deliberate and 3.3 depends on it. **It is also the live hazard in this story** —
    see F6.

11. [x] **No `math/big.Float`, no `math/big.Rat` (D-3.1a.1, Layer 2).** Banned at the `folio-go`
    module root by resolved type identity. `internal/expr` is new code inside that scope.

12. [x] **D-000.61 (+ extension).** A float red-proof varies operand ORDER and asserts
    order-invariance; a mutation introduces the **ERROR**, not merely the float **TYPE**. Applies if
    any red-proof in this story touches arithmetic.

---

## Findings — measured in this run at `fde96b5`

### F1. The rank table already reserves `expr` at 3 — creating the package trips exactly one guard, and it is the intended one

`lint/internal/rules/stagerank.go:58-63`:

```
{"geom", 0}, … {"template", 2}, {"expr", 3}, // Epic 3; D-1.6.1's expr -/-> bind pre-commitment lives in this number
{"bind", 4},
```

So `internal/expr` is **pre-ranked**. Creating it does **not** trip the unranked-stage fail-safe, and
`expr → template` (3 → 2, strictly lower) and `bind → expr` (4 → 3, strictly lower) are both legal.
The **only** guard that reddens on creation is `absence-expr-package`, which is the point.

### F2. THE RANK TABLE FORBIDS THE OBVIOUS PLACE FOR THE `footerOf` DERIVATION — and this is the story's biggest structural finding

DW-5's derivation is a **load-time** check. Every existing footer check lives in
`internal/template/parse_bands.go:403-426` (AC43's three checks). But **`template` is rank 2 and
`expr` is rank 3**, and stagerank forbids importing **equal or higher** rank. **`internal/template`
can never call `internal/expr`.** D-1.4.2 said *"1.4 cannot call `internal/expr`"* and then noted
*"the wrinkle is larger"* — this is how much larger: it is not a sequencing problem, it is a
permanent layering fact.

**Forced resolution (see R1 below):** the derivation lives in `internal/expr`, which imports
`template` (legal, and it already must for `SplitJSONNumber`), and is invoked from
**`folio.ParseTemplate`** (`folio-go/folio.go:53-59`) — the module-root public entry, **outside
`internal/`, so unranked**. D-1.4.2's own backstop framing (*"3.7 — `folio.Validate` must include
it"*) already puts this obligation at the root. See **FLAG-1** for the residue.

### F3. The canonical golden REQUIRES an unimplemented function to parse successfully

`folio-go/testdata/template/golden/worked-example.json:19`:

```
"bind": "{{formatNumber(transaction.amount, \"#,##0.00\")}}",
```

That is **D-1.4.1 derivable shape 2** verbatim (*a single `formatNumber(<bare row-scoped path>,
<pattern literal>)` call*), and D-1.4.10 confirmed the worked example *"remains valid under D-1.4.1
as written — its `bind` is a single `formatNumber(…)`"*. Under D-1.4.3 this file **is** the
definition of canonical (P3's fixed point).

**Therefore `formatNumber` — unimplemented until 3.4 — must PARSE, and its derivation must FIRE.**
If a registered-but-unimplemented function were a **load** error, the canonical worked example would
stop loading and shape 2's derivation could never be exercised at all. This is what forces R3 below.

### F4. Four `type Decimal` declarations live under `testdata/` — a naive walk makes "exactly one" false by construction

```
testdata/arch/reducer-inventory/{baseline,missing-reducer,extra-reducer,moved-decimal/pkga}/decimal.go
internal/bind/decimal.go:35                      ← the only real one
```

`walkGoFiles` (`internal/arch_test.go:106-133`) **skips `testdata` and dot-prefixed directories**
(line 115). D-3.2.1's set-equality assertion **must reuse it** and must say in its own comment why.
A hand-rolled `filepath.WalkDir` finds five and reports a false red; a `grep` finds five **plus**
`internal/template/ids.go:70`'s unrelated `base36ToDecimal`, which is exactly why D-000.14 says *by
AST, never text, never a filtered pipe*.

### F5. The "exactly one `Decimal`" property is HALF-present already — and the half that is missing is the whole point

`reducerDecimalInventory`'s pass 1 (`internal/reducer_inventory_arch_test.go:205-228`) already
**errors** on a second `Decimal` declaration in a *different* directory. What it does **not** do is
assert **where** that directory is — it is deliberately relational (D-3.1a.3). So D-3.2.1's assertion
is **not** D-000.38 redundancy: it adds **location** (`internal/expr`), which nothing asserts today,
and it must be a **separate** guard so that D-3.1a.3's guard stays unedited. Two known narrow spots
in the existing pass-1 check, neither of which this story is obliged to fix: it does not catch a
second declaration in the **same** directory, and it fires as a scanner *error*, not a named verdict.

### F6. THE `sum → 0` HAZARD IS LIVE TODAY, not hypothetical

`internal/bind/reduce.go:96-99`:

```go
func SumDecimals(items []Decimal) (Decimal, error) {
	if len(items) == 0 {
		return Decimal{Coefficient: 0, Exponent: 0}, nil
	}
```

`SumDecimals` **already exists** and **already returns a plausible `{0,0}` on empty input** — by
deliberate ruling (D-3.1a.2). So the cheapest possible wiring of a `sum` table entry in this story
(*"point it at the kernel we already have"*) yields **a silently wrong bank-statement total of
zero**. This is D-000.25's vendor-default hazard reproduced in our own code, and it is the honest
mutation for AC17's red-proof. There is no equivalent hazard for `avg` (it errors on empty), which is
itself evidence that the two must be tested separately rather than as a pair.

### F7. D-000.34 hunt — the tests that die silently at the move

Every `_test.go` in `internal/bind` is an **in-package** test. Two of them draw their discriminating
power from **unexported** constants that travel with `Decimal`:

| file | unexported symbols used | what breaks |
|---|---|---|
| `internal/bind/decimal_test.go` | `maxDecimalExponentMagnitude`, `maxDecimalCoefficientDigits` (lines 85-136) | must move **with** the type; if left behind and "fixed" by hardcoding `100_000`/`19`, the test **silently decouples from the constant it exists to pin** and stops tracking a change to it |
| `internal/bind/reduce_test.go` | `maxDecimalExponentMagnitude`, `avgExtraScale` (lines 131, 246-305) | same shape; `avgExtraScale` is D-3.1a.1's `+4`, whose whole forcing function (Story 3.4) depends on it being referenced, not copied |

**This is exactly D-000.34's shape: a test that passes only because the type was package-local.** The
correct move is *test travels with the code it tests*; the wrong move is *hardcode the number to make
it compile*. Both compile. Only one keeps the assertion honest.

Consumers that merely **reference** `Decimal` and are ordinary compile breaks (expected, per D-3.2.1,
and **not** a reason to defer): `internal/bind/value.go:63` (`AsDecimal`), `internal/bind/reduce.go`,
`internal/bind/text.go`, `internal/bind/{text,golden,decodeguard,value}_test.go`,
`folio-go/render_entry.go:74` (comment only).

### F8. `internal/bind/resolution_roots_arch_test.go` is directory-scoped and this story moves code out of that directory

It scans `os.ReadDir(".")` — package `bind`'s own directory — for `lookupBound` call sites (AC3's
structural guard, D-000.52). This story deletes `parseBindingPath` from `text.go` and may relocate
placeholder dispatch. **It has a presence precondition** (line 163: *"zero lookupBound call sites
found"* is a `t.Fatal`), so it fails loudly rather than going slack — **verified, do not add a second
one** (D-000.42). But the developer must confirm it still observes both declared roots afterwards and
record the observed set, rather than assuming.

### F9. The module already has a third-party dependency, so a module-wide reading of AC1 is a false red

`folio-go/go.mod` requires `github.com/boxesandglue/textshape v0.0.15`. AC1's *"no third-party
dependency"* is therefore **scoped to `internal/expr`'s own import set** — see R4.

### F10. Three existing tests currently assert expression syntax is REJECTED, and a lazy re-point makes all three vacuous

| site | literal | today |
|---|---|---|
| `internal/bind/text_test.go:83` | `{{formatNumber(transaction.amount, "#,##0.00")}}` | rejected by `parseBindingPath` |
| `internal/bind/spans_test.go:190` | `{{sum(x)}}` | rejected |
| `render_bind_test.go:206` | `{{formatNumber(a, "x")}}` | rejected |

After 3.2 all three **parse** and then fail as *unimplemented*. **Both states are `err != nil`.** So a
re-point that only asserts "an error occurred" **passes identically before and after** — D-000.9's
signature failure shape. Each re-point must assert the **message distinguishes the two causes**.
(`internal/template/unbreakable_test.go:119`'s `sum(customer.name)` case is **unrelated** — it is
`decodehelpers.go`'s bare-dotted-path field validation at rank 2, which must **not** be loosened.)

### F11. Nothing in the record specifies what `if()` does

`upper()`/`lower()` are self-evident. `if()` is not. FR18 names it; the PRD's own adversarial review
(`review-adversarial.md:152-153`) states the gap precisely: *"`if(x, a, b)` requires a boolean `x`.
There is no `>`, `<`, `==`, `&&`, `!`. As specified, `if()` can only branch on truthiness of a bare
path."* Story 3.5 (`epics.md:1084-1086`) then says *"when the condition evaluates false"* **without
defining false either**. See **FLAG-3**.

---

## Rulings this story makes, and why each is FORCED rather than chosen

### R1 — The derivation lives in `internal/expr`, invoked from `folio.ParseTemplate`

Forced by **F2**: `template` (2) can never import `expr` (3). The root package `folio` is outside
`internal/` and unranked, already imports `internal/template`, and is *"the sole bridge into
`internal/template`'s parser"* (`folio.go:52`). D-1.4.2 already places the backstop there
(`folio.Validate`). A non-derivable footer is therefore a **load error from the public API**, which is
what D-000.59's replacement assertion demands. No new package, no rank change, no ruling needed.

### R2 — The derived `footerOf` is NEVER written back into the `Document`

`internal/template/serialize.go:312` emits `footerOf` whenever `c.FooterOf` is present. D-1.4.3's
**P3** defines canonical as a fixed point: `Serialize(Parse(b)) == b`. If the derivation populated
`FooterOf`, a document that legitimately **omitted** `footerOf` would serialize with it — breaking P3
and P2 for exactly the ordinary documents D-1.4.3 warns about. So the derivation returns a **resolved
result alongside** the document; the stored schema is untouched. Also AD-13: a written-back derived
value is a second source of truth against `bind`, which is the same reason `footerOf` beside
`footer: "count"` is already a load error.

### R3 — A registered-but-unimplemented function PARSES, DERIVES, and fails at EVALUATION — never at load

Forced by **F3**: the canonical golden binds `{{formatNumber(…)}}`, and D-1.4.1 shape 2 **requires**
that exact call to be derivable. A load-time rejection of unimplemented functions would break the
document that defines what canonical means, and would make shape 2 unreachable. So the boundary is:
**syntax and arity at load; execution at evaluation.** This is also the shape the lead's ruling wants
— *a registered-but-unimplemented function is a LOCATED ERROR, never a plausible value* — because the
error must name an **element id**, and element ids are only meaningful once a specific binding is
being resolved.

### R4 — AC1's "no third-party dependency" is scoped to `internal/expr`'s import set

Forced by **F9**: the module already depends on `textshape`, so a module-wide assertion is red at
baseline for a reason that has nothing to do with this story. Scope: every import in every
non-`_test.go` file under `folio-go/internal/expr/` resolves to the standard library or to
`github.com/panitw/folio/folio-go/...`. Stated as a D-000.23 witness: this covers **`internal/expr`,
not the module**.

### R5 — Every AST count in this story reuses `walkGoFiles`

Forced by **F4**. And its `testdata`-skipping behaviour is asserted, not assumed — see AC22.

---

## Acceptance Criteria

**33 ACs.** The epic states **four** `**Given**` blocks (`epics.md:997-1013`), counted directly:
AC1+AC2 (parser + closed eight), AC3 (the three implemented), AC4 (syntax error naming element id
and offending text), AC5 (a ninth is visible in a diff, C1). Everything below AC12 is the D-000.59 /
D-3.2.1 / D-1.6.5 obligation set, which the epic text does not enumerate.

### The parser (epic Given 1, first half — AD-9)

- [x] **AC1.** `internal/expr` contains a **hand-written recursive-descent** parser for the
  expression grammar. No parser generator, no generated file, no third-party dependency, and no
  regex engine standing in for the grammar. **Per D-3.2.2 this is settled against CEL and every
  general-purpose expression library**; a developer reaching for one stops and reads D-3.2.2.
- [x] **AC2.** An AST guard asserts every import in every non-`_test.go` file under
  `folio-go/internal/expr/` resolves either to the standard library or to
  `github.com/panitw/folio/folio-go/…`. Its **coverage witness states, in these words**, that it
  covers **`internal/expr`'s import set, not the module's** (D-000.23; R4; `go.mod` already requires
  `textshape`). It carries a presence precondition: **zero files parsed is a `t.Fatal`**, not a pass
  (D-000.9). **This guard is D-3.2.2's mechanical half**: it is what makes *"no expression library"* a
  fact rather than a memory, and its doc comment cites D-3.2.2 so a future reader learns the reason
  (the numeric model, not the parser) rather than only the prohibition.
- [x] **AC3.** The grammar accepts: a bare dotted path (1.6's grammar, preserved verbatim —
  `ident ( "." ident )*`, `ident := [A-Za-z_][A-Za-z0-9_]*`), a function call over comma-separated
  arguments, a double-quoted string literal, and a number literal. Nesting is supported to at least
  one level (`formatNumber(sum(t.amount), "#,##0.00")` parses). Every accepted and every rejected
  form is enumerated in a table test.
- [x] **AC4.** `{{page}}` and `{{pages}}` are short-circuited **before** any parse attempt, exactly as
  `internal/bind/text.go:180-188` does today — left byte-for-byte unchanged, never resolved from
  data, never an error (AD-4, D-1.6.5). The retained fixture stands: **data containing a top-level
  `page` key must not change what `{{page}}` renders.**

### The closed table of eight (epic Given 1, second half + Given 4 — FR18, C1)

- [x] **AC5.** Exactly one function table exists, as a **single package-level literal in one file** in
  `internal/expr`, with **exactly eight entries**, keyed by FR18's eight names: `sum`, `count`, `avg`,
  `formatDate`, `formatNumber`, `upper`, `lower`, `if`.
- [x] **AC6.** **No exported registration path exists.** An AST guard asserts `internal/expr` exports
  no function that adds to, mutates, or replaces the table. *Closed* means closed at compile time,
  not "closed by convention".
- [x] **AC7.** An **AST set-equality** guard (D-000.14, reusing `walkGoFiles` per R5) extracts the
  table's registered names from the literal and asserts the set is **exactly** those eight. It
  reports its own stats and **fails on zero entries extracted** (D-000.9) — an extractor that finds
  nothing must not read as "the set matched".
- [x] **AC8. C1's requirement, discharged concretely.** Adding a ninth function requires editing
  **both** the table literal **and** AC7's expected set, in the same diff, or CI is red. A test
  documents this by name and cites C1. Red-proof: **AC24**.
- [x] **AC9.** Each of the three aggregation entries (`sum`, `count`, `avg`) declares a
  **`Decimal`-typed signature** (D-3.2.1: *"declaring the table honestly requires referencing
  `Decimal`"*). A stringly-typed table is a defect on its own merits and is a review blocker.
- [x] **AC10.** Declared **arity** is checked for all eight at parse time (decidable syntax; and R3's
  derivation needs `formatNumber`'s arity-2 shape). `sum(a, b)` is a **located parse error naming the
  element id**, not an unimplemented-function error. See **FLAG-2** for what this does not cover.
- [x] **AC11.** An **unknown** function name — a ninth called from a template — is a located error
  naming the element id, the offending name, and the eight legal names. Not a panic, not a silent
  pass-through to a path lookup.

### The three implemented (epic Given 2)

- [x] **AC12.** `upper()` and `lower()` evaluate per Go's `strings.ToUpper`/`ToLower`. Edge cases
  covered explicitly: empty string, a Thai string (the golden report is a Thai bank statement — case
  mapping must be a **no-op**, and a test asserts the bytes are unchanged), a CJK string, a string
  with combining marks, and a non-string operand (a located error, never a coerced stringification).
- [x] **AC13.** `if()` evaluates per the truthiness rule adopted in **FLAG-3**, with each rule-arm
  covered: true branch, false branch, absent path, explicit null, empty string, zero. **Both branches
  must be asserted** — a test exercising only one arm is vacuous.
- [x] **AC14.** `if()` evaluates **only the selected branch**. Red-proof: the unselected branch calls
  an unimplemented function (e.g. `if(cond, "a", sum(t.x))`); with the condition selecting `"a"`, the
  expression must succeed. A non-short-circuiting implementation errors.

### The five unimplemented — a LOCATED ERROR, never a plausible value

- [x] **AC15.** `sum`, `count`, `avg`, `formatDate`, `formatNumber` are **registered** (AC5) and
  **parse and derive** successfully (R3), but **evaluating** any of them produces a located error
  naming: the function, the element id, the offending expression text, and the **owning story**
  (`sum`/`count`/`avg` → 3.3; `formatDate`/`formatNumber` → 3.4).
- [x] **AC16.** One test per unimplemented function — **five distinct assertions**, not a loop over a
  table that could silently shrink to zero (the loop, if used, asserts it iterated exactly five).
- [x] **AC17. THE VENDOR-DEFAULT RED-PROOF (F6, D-000.25).** A test asserts that evaluating
  `{{sum(transactions.amount)}}` returns an **error**, and **specifically that it does not return a
  `Decimal` of value zero**. **Red-proof, captured before the assertion is satisfied:** wire the
  `sum` entry to `bind.SumDecimals` over an empty operand slice — which compiles, which is the
  obvious thing to write, and which returns the plausible `{0,0}` (`reduce.go:96-99`) — and record the
  assertion failing. **The mutant must introduce the WRONG ANSWER, not merely the wrong type**
  (D-000.61's extension, applied one axis over).
- [x] **AC18.** The unimplemented-function error and the unknown-function error (AC11) are
  **distinguishable by message**, and a test asserts they differ. *"`sum` is coming in 3.3"* and
  *"there is no function called `frobnicate`"* are different facts and must not collapse.

### Syntax errors (epic Given 3 — AD-14)

- [x] **AC19.** A syntax error inside `{{ }}` fails **at load** with a located error naming **the
  element id** and **the offending expression text**, verbatim, as the author wrote it — with one
  deliberate exception, corrected here by the finisher (QA Nit 4): outer whitespace inside `{{ }}` is
  trimmed before it is reported (`strings.TrimSpace(ph.Inner)`), so "verbatim" means the placeholder's
  content modulo its own leading/trailing whitespace, not the untrimmed byte range. Harmless and more
  readable; not literally verbatim as first written. Cases: unbalanced parenthesis, trailing comma,
  unterminated string literal, empty expression, a bare operator (`{{a + b}}` — there are no
  operators, F11), and a path segment starting with a digit.
- [x] **AC20.** The error carries **no** diagnostic code. `internal/diag` does not exist until 3.6
  (`absence-diag-package` is still live) and this story **must not mint one early** (D-1.4.2's
  precedent for 1.4). Plain Go errors, named fields, element id.

### D-000.59's discharge — three positive assertions, one commit

- [x] **AC21. Assertion (a) — DW-5's obligation, BOTH ARMS.**
  - **Derivable arm:** a template whose `columns[].footer` is present with `footerOf` **omitted**, and
    whose `bind` is one of D-1.4.1's two shapes, **RESOLVES to the derived `footerOf`** — and the test
    asserts the **resolved value** (`transactions.amount`), not merely that the document loaded.
    Both shapes are covered: shape 1 (`{{row.amount}}` → `transactions.amount`) and shape 2
    (`{{formatNumber(transaction.amount, "#,##0.00")}}` → `transactions.amount`, **and**
    `footerFormat` defaults to `#,##0.00`).
  - **Rejection arm:** any **other** `bind` shape (e.g. `{{if(row.x, row.a, row.b)}}`) is a **load
    error naming the column id**.
  - **The derivable arm is not optional: D-000.59 states it fails vacuously if only the rejection is
    tested.**
  - Per **R2**, the derived value is **not** written into the `Document`; a test asserts
    `Serialize(Parse(b)) == b` still holds byte-for-byte for a document that omitted `footerOf`
    (D-1.4.3 P3).
- [x] **AC22. Assertion (b) — D-3.2.1's forcing function.** **Exactly ONE `Decimal` type declaration
  exists in the module, and it is in `internal/expr`.** Set equality **over AST-extracted type
  declarations** (D-000.14 — by AST, never text, never a filtered pipe), reusing `walkGoFiles`
  (R5/F4). The guard's own comment records **why** the walker's `testdata` skip is load-bearing
  (four decoy declarations at `testdata/arch/reducer-inventory/`), and a test asserts the walker
  actually skips them — otherwise F4's trap is one refactor away from silently re-arming.
- [x] **AC23. Assertion (c) — DW-8's other half.** `parseBindingPath` and `isValidIdent` are
  **ABSENT** from the module once `internal/expr` exists. An extinction guard over AST-extracted
  function declarations (correct instrument here — unlike AC22, where D-3.2.1 explicitly rules that
  **location is the property and extinction is the wrong instrument**, since `Decimal` must exist
  somewhere).
- [x] **AC24. THE RED-PROOFS, captured BEFORE the obligations are wired (D-000.30).** Recorded
  verbatim in the Delivery Log, each with the exact command, the failure text, **and what it would
  have printed had the mutation not been applied** (D-000.9 extended):
  - AC21 red: run the new derivation assertion at `fde96b5` against the un-repointed fixture →
    fails, because the derivation does not exist.
  - AC22 red: run the assertion **before** the move → fails naming `internal/bind`, the wrong
    location. Then a **second** mutation: leave a duplicate `Decimal` in `bind` after the move →
    fails naming two locations. **Both**, because the single-location red does not prove the
    cardinality half.
  - AC23 red: run the extinction guard at `fde96b5` → fails naming
    `internal/bind/text.go:328` and `:345`.
  - AC8/C1 red: add a ninth entry to the table literal → AC7 fails. Remove it.
  - AC17 red: F6's `SumDecimals` wiring, per AC17.

  **Finisher correction (QA Finding 18, Minor).** The review found that three of these five —
  AC21 red, AC22's first mutation, and AC23 red — were, as originally recorded, **reasoning and
  compile-checks, not observations**: "verified by inspection of the pre-change `folio.go` (git
  history)", "reproduced structurally… the pre-move source is preserved in git history", and "against
  that tree **would have failed**" respectively. None was an actual command run against an actual
  failing tree. The review's own suggested resolution — stand up a worktree at `10327a4` and run the
  assertions there — is not achievable for AC21 or AC23 as literally proposed: the assertions in
  question (`TestParseTemplateFooterOfDerivationNeverWritesBack`'s repair, and the extinction guard's
  exact test body) are new code that references packages (`internal/expr`, `folio_expr_validate.go`)
  which do not exist at `10327a4` at all — this story's whole obligation lands in one commit, by
  design (D-000.59's own discharge charter), so there is no earlier commit where the new assertion and
  the old, unmet code both exist simultaneously to run it against. AC22's first mutation is
  structurally the same case (the location-scanning guard itself is new code). Rather than force an
  artificial reconstruction, the honest fix is naming these three accurately: they are **reasoned
  demonstrations, verified by inspection of preserved git history**, not standing or one-time *executed*
  red-proofs — which is exactly what D-000.9/D-000.30 exist to keep a reader from being told otherwise.
  AC22's second mutation (the standing `TestDecimalUniquenessRedProof`), AC8/C1's red-proof, and AC17's
  red-proof remain genuinely **run and observed** (AC17 additionally re-verified live by the reviewer,
  M1, byte-identical to the Delivery Log's own account).
- [x] **AC25.** `absence-expr-package` is **deleted** from `absenceChecks`
  (`lint/internal/rules/absences.go:94-98`) in the **same commit** as AC21-AC23. Per **D-000.37**,
  that entry's `desc` names *"Story 3.2"*; it is deleted **with** the rule, not left orphaned.
  `absence-diag-package` and `absence-source-date-epoch` are **untouched** — they belong to 3.6 and
  3.7 and their `desc` story numbers are still correct.
- [x] **AC26.** `TestAbsencesProductionScan`'s vacuity guard still holds after the deletion:
  `ChecksEvaluated` must report the **new, smaller, non-zero** count. A list that has shrunk to
  something an empty list would satisfy is precisely the D-000.9 exposure `AbsencesStats` was built
  for (`absences.go:126-135`). The expected count is updated deliberately, with the number stated.
- [x] **AC27.** **AC44's fixture is RE-POINTED** (D-000.59, explicit), and its closure is visible in
  the diff — **amended by the finisher (QA Finding 17) to record what actually shipped, Decision 1's
  own choice, rather than the text as originally drafted below.** As originally written, this AC
  called for `TestFooterWithoutFooterOfLoads` to *move* to the root `folio` package, its body
  *transformed* into the AC21 derivable-arm assertion. Decision 1 took a materially different route,
  for a reason the review judged sound: the test **stays** in `internal/template/footer_test.go`, its
  **body byte-unchanged**, and only its doc comment rewritten to reframe the permissiveness as a
  deliberate layer-boundary witness (FLAG-1) naming `folio.ParseTemplate`/
  `validateAndDeriveExpressions` by symbol; the AC21 derivable-arm assertions are **separate, new**
  tests at the root (`TestParseTemplateDerivesFooterOfShape1`/`Shape2`,
  `folio_expr_validate_test.go`), independently proven non-vacuous by the review's own mutation (M4).
  The closure is still `git log --follow`-able and the gap is still visibly closed; it is a kept
  witness plus new root-level proof, not a relocated-and-transformed single test. This text now
  describes that shipped shape rather than the abandoned one.

### DW-8's `Decimal` move (D-3.2.1)

- [x] **AC28.** `Decimal`, `NewDecimal`, `SumDecimals`, `AvgDecimals` and their unexported bounds
  (`maxDecimalCoefficientDigits`, `maxDecimalExponentMagnitude`, `avgExtraScale`) move to
  `internal/expr`. **Moved, never duplicated.** `internal/bind` imports them back (4 → 3, legal).
- [x] **AC29. F7's D-000.34 obligation.** `internal/bind/decimal_test.go` and
  `internal/bind/reduce_test.go` **travel with the code they test**. **No unexported bound may be
  replaced by a hardcoded literal to make a left-behind test compile** — that is the silent death
  D-000.34 names, and `avgExtraScale` in particular is the anchor of D-3.1a.1's Story 3.4 forcing
  function. The Delivery Log states, per test file, whether it moved and why.
- [x] **AC30.** **`internal/reducer_inventory_arch_test.go` is UNEDITED** — byte-identical before and
  after, asserted by SHA-256 in the Delivery Log. Its relational location clause (D-3.1a.3) must
  follow the move with **zero changes**. **If the developer finds themselves editing its expected
  value, they stop and report it rather than editing** (D-3.1a.3, verbatim). Its `moved-decimal`
  fixture variant already exercises exactly this scenario.
- [x] **AC31.** `internal/bind/resolution_roots_arch_test.go` still observes **both** declared
  resolution roots after the matcher deletion (F8). The observed set is recorded in the Delivery Log.
  Its existing presence precondition is **not** duplicated (D-000.42).
- [x] **AC32.** F10's three re-pointed tests each assert the message **distinguishes** "parsed, then
  unimplemented" from "rejected as unsupported syntax". A bare `err != nil` re-point is a review
  blocker: it passes identically at `fde96b5` and after.

### Cadence and guardrails

- [x] **AC33.** Gates re-measured at completion and reported **with scope and flags** (D-000.26):
  `folio-go` unit/vet/build, `lint` (naming the **three known-environmental DW-19 failures** if the
  working checkout is used). **`TestCorpusMeetsP6ExerciseFloors` stays RED with byte-identical
  stats** `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — never "fixed". No matrix run: it is
  due at Epic 3's close (D-000.4).

---

## Task breakdown

1. **Capture every red-proof first (AC24).** All five, at `fde96b5`, before a single line of the
   obligation is wired. **This window shuts permanently at the move.**
2. Create `internal/expr` — the recursive-descent parser, the grammar, the eight-entry table
   (AC1-AC11, AC19-AC20).
3. Move `Decimal` and the reducers, with their tests (AC28-AC29). Fix the compile breaks at
   `bind/value.go:63`, `bind/reduce.go`, `bind/text.go` and their tests.
4. Implement `upper`/`lower`/`if`; make the other five located errors (AC12-AC18).
5. Delete `parseBindingPath`/`isValidIdent`; route `bind`'s placeholder resolution through
   `internal/expr` (AC23, AC31-AC32).
6. Land the `footerOf` derivation in `internal/expr`, hook it into `folio.ParseTemplate`, re-point
   AC44's fixture (AC21, AC27, R1, R2).
7. Land AC22's and AC23's guards; delete `absence-expr-package`; fix `ChecksEvaluated`
   (AC22, AC25-AC26).
8. Retire **DW-5** and **DW-8** in `deferred-work.md`, with the discharge recorded.
9. Re-measure the gates (AC33); confirm AC30's SHA-256.
10. Set the tracker to `review`.

*(No commit task: this story ends at `status → review`.)*

---

## Flags — and what each one SUPPRESSES (D-000.60)

### FLAG-1 — `internal/template.ParseDocument` stays permissive; strictness lives one layer up

**The fork.** F2 proves `template` (2) can never call `expr` (3). R1 puts the derivation at
`folio.ParseTemplate`. A caller who reaches `internal/template.ParseDocument` directly — which only
in-repo tests can do — therefore still gets the AC44 permissiveness. DW-5 says *"the derivation is
wired"*; it does not say at which layer, and D-1.4.2's backstop framing points at the root, but does
not settle whether `internal/template` owes its own **rank-legal syntactic pre-check** (a cheap
paren-balance / shape screen that needs no parser).

**What this flag SUPPRESSES:**
- **A test not written:** no AC asserts `internal/template.ParseDocument` rejects a non-derivable
  footer. If the lead rules that template owes a pre-check, that AC is added and AC27's fixture is
  **strengthened in place** instead of relocated.
- **A scope not taken:** no rank-legal shape screen is built in `internal/template`.
- **A compensating obligation taken instead:** `internal/template/footer_test.go` keeps a **witness
  test stating what `template` alone does NOT cover** (D-000.23), so the relocation of AC27 does not
  read as the check having quietly vanished from that package.

### FLAG-2 — argument TYPE checking for the five unimplemented functions

**The fork.** AC10 rules arity is checked at parse for all eight, because arity is decidable syntax
and R3's derivation needs it. Argument **types** are different: `sum`'s operand must resolve to a
collection of numbers, which is only knowable against real data. D-3.2.1 requires the three
aggregation entries to carry `Decimal`-typed **signatures**, but nothing settles whether 3.2 must
**enforce** them.

**What this flag SUPPRESSES:**
- **A test not written:** no AC asserts that `{{sum(customer.name)}}` (a string operand) is rejected
  by 3.2. Under AC15 it fails as *unimplemented*, which is correct but is not a type error.
- **A scope not taken:** no type checker is built in 3.2. If the lead rules otherwise, this is
  Story 3.3's natural home for `sum`/`count`/`avg` and 3.4's for the two formatters — but then those
  stories inherit an obligation nobody has written down, which is the reason for flagging rather
  than assuming.

### FLAG-3 — `if()` has no specification anywhere in the record (F11)

**The fork.** FR18 names `if()`; nothing defines its arity, its truthiness rule, or its operand
types. The PRD's own adversarial review named this gap (`review-adversarial.md:152-153`) and it was
never closed. Story 3.5 then leans on *"the condition evaluates false"* without defining false. This
is a **decision, not a detail**: the truthiness rule 3.2 adopts becomes 3.5's visibility semantics,
and getting it wrong means a section silently disappearing from a rendered statement — the same class
of harm as F6's silent zero.

**The narrow proposal, offered for ruling, not adopted unilaterally.** `if(cond, a, b)` — arity
**exactly 3**. `cond` must resolve to a **JSON boolean**; **absent** and **explicit null** are a
located error, **not** false (AD-14 already distinguishes those three cases at
`internal/bind`'s `Presence`, and collapsing them here would discard a distinction the codebase paid
for). No string, number or array is coercible. Only the selected branch is evaluated (AC14).

**What this flag SUPPRESSES:**
- **A test not written:** no AC asserts coercion behaviour for a non-boolean condition beyond
  "located error" — because if the lead rules for truthy coercion (empty string false, `0` false,
  empty array false), the whole of AC13's case table changes shape.
- **A scope not taken:** no comparison operators. FR18 lists no `>`, `<`, `==`, `&&`, `!`, and adding
  one would be a **C1-adjacent language expansion** made in an implementation story — precisely what
  C1 exists to prevent.
- **A downstream obligation deferred:** Story 3.5's visibility semantics are left to inherit whatever
  3.2 adopts. If the ruling lands after 3.2 ships, **D-000.60 applies**: whoever rules it must ask
  what this flag suppressed and lift it in the same act.

---

## Delivery Log

### Summary

`internal/expr` is a new package: a hand-written recursive-descent parser (AD-9) over the
expression grammar, the closed eight-entry function table (FR18/C1), an evaluator implementing
`upper`/`lower`/`if` and reporting the other five as located "not yet implemented" errors, the
`footerOf`/`footerFormat` derivation (D-1.4.1), and `Decimal`/`SumDecimals`/`AvgDecimals` moved
here from `internal/bind` (D-3.2.1, DW-8). `internal/bind`'s 1.6 path matcher
(`parseBindingPath`/`isValidIdent`) is deleted; `internal/bind/text.go`'s resolver now parses and
evaluates through `internal/expr`. `folio.ParseTemplate` gained a load-time walk
(`folio_expr_validate.go`) that statically checks every `{{ }}` expression in the document (syntax,
arity, unknown function, literal-argument kind) and runs the footerOf derivation for columns
requesting a `sum`/`avg` footer without an explicit `footerOf`.

### Files

**New — `internal/expr`:** `doc.go`, `ast.go`, `parser.go` (+`parser_test.go`), `table.go`
(+`table_test.go`), `check.go` (+`check_test.go`), `eval.go` (+`eval_test.go`), `footer.go`
(+`footer_test.go`), `scan.go`.

**Moved from `internal/bind` to `internal/expr` (D-3.2.1, F7/D-000.34):** `decimal.go`/
`decimal_test.go`, `reduce.go`/`reduce_test.go`, `golden_test.go` (Layer 1's reduction-kernel oracle
— also draws its discriminating power from `Decimal`/`SumDecimals`/`AvgDecimals` and moved for the
same reason as decimal_test.go/reduce_test.go, a third D-000.34 site the story's own F7 finding did
not enumerate but which the same principle covers), and `testutil_test.go` (its
`repoRootFromTest`/`mustReadFile` helpers, used only by `golden_test.go`). No unexported bound was
hardcoded to make a left-behind test compile — every constant travelled with the type.

**New in package `folio`:** `folio_expr_validate.go` (+`folio_expr_validate_test.go`).

**New — `internal/expr_arch_test.go`** (package `arch`, at `folio-go/internal/`): AC6/AC7/AC8/AC22/
AC23's structural guards.

**Edited:** `folio-go/folio.go` (ParseTemplate wiring, `Template.derivedFooters`),
`folio-go/internal/bind/text.go` (resolver rewritten onto `internal/expr`; `parseBindingPath`/
`isValidIdent` deleted), `folio-go/internal/bind/value.go` (`AsDecimal` returns `expr.Decimal`),
`folio-go/internal/bind/text_test.go` (F10 re-point), `folio-go/render_bind_test.go` (F10 re-point),
`folio-go/internal/template/footer_test.go` (Decision 1 re-framing of `TestFooterWithoutFooterOfLoads`
— kept, not relocated), `lint/internal/rules/absences.go` (deleted `absence-expr-package`),
`lint/internal/rules/absences_test.go` (`TestAbsencesChecksIncludeAllFourEntries` →
`...AllThreeEntries`), `_bmad-output/implementation-artifacts/deferred-work.md` (DW-5/DW-8 retired
in place), `_bmad-output/specs/spec-folio/folio-format.md` (new `### Expressions` section; the
`columns[].footerOf` row's derivation-status sentence updated).

**Untouched, verified byte-identical (AC30, D-3.1a.3):**
`folio-go/internal/reducer_inventory_arch_test.go` — SHA-256
`1d5d3a70796703109c4d48fe05dcc64f9df0f5a54bda8d876e2d09c7cf9e74e6` before and after. Its relational
location clause followed the `Decimal` move with zero edits, exactly as D-3.1a.3 requires.

### Decisions applied (routed and settled before/at this run — not re-opened)

- **`if(null)` is FALSE, silently** — owner decision, no diagnostic. Implemented in
  `internal/expr/eval.go`'s `evalIf`; dedicated test `TestIfNullConditionIsSilentlyFalse`
  (`internal/expr/eval_test.go`), paired deliberately with
  `TestIfAbsentConditionIsLocatedError`. Documented in `folio-format.md`'s new `### Expressions`
  section (the D-000.6-pattern stated-behaviour entry the ruling required).
- **`if()`'s other clauses** (AD-14 determines all four): arity 3; condition is JSON-boolean-only,
  no truthiness (`argNotLiteral` at parse/Check catches a literal condition; a resolved non-bool
  Value is a located error at eval); an absent-path condition is a located error carrying the path;
  short-circuit evaluation (only the selected branch is passed to `Eval`) — the accepted cost (an
  absent/mistyped path in the untaken branch goes unreported, and `folio.Validate` cannot see it
  either) is stated in `eval.go`'s doc comment and is FLAG-3's resolved fork.
- **Decision 1 (footerOf derivation lives at the module root):** `internal/expr.DeriveFooterOf`,
  invoked from `folio.ParseTemplate` via `folio_expr_validate.go`. The rank table
  (`lint/internal/rules/stagerank.go`) was NOT amended. The derived value is never written back into
  `template.Document` (R2) — `TestParseTemplateFooterOfDerivationNeverWritesBack`
  (`folio_expr_validate_test.go`) proves P3 (`Serialize(Parse(b)) == b`) still holds byte-for-byte.
  AC44's fixture (`TestFooterWithoutFooterOfLoads`) is KEPT in `internal/template/footer_test.go`,
  re-framed (not relocated) as a deliberate layer-boundary witness naming
  `folio.ParseTemplate`/`validateAndDeriveExpressions` by symbol.
- **Decision 2 (no template-level shape screen):** `internal/template.ParseDocument` is unedited and
  stays permissive. The compensating witness is `TestFooterWithoutFooterOfLoads`'s rewritten doc
  comment (D-000.23 shape).
- **Decision 3 (argument checking splits three ways):** arity — `internal/expr/check.go`'s
  `checkCall`, AC10. Literal-argument kind — `checkArgKind` (`argNotLiteral`/`argStringLiteral`),
  covering AC10's own `sum("hello")`/`formatNumber(x, 123)` examples plus `if()`'s literal-condition
  case. Path-argument VALUE kind is explicitly NOT checked here — `table.go`'s doc comment and this
  Delivery Log both name it as a FORWARD OBLIGATION owed at evaluation by Story 3.3
  (`sum`/`count`/`avg`) and Story 3.4 (`formatDate`/`formatNumber`), per D-000.60 — not an open
  fence.
- **The eight-entry table split (syntax/arity at load, execution at evaluation):** proved directly —
  `TestParseTemplateAcceptsCanonicalGolden` loads `testdata/template/golden/worked-example.json`
  (whose `bind` uses the still-unimplemented `formatNumber`, F3) successfully and asserts its
  derived footerOf; `TestUnimplementedFunctionsAreLocatedErrors`/
  `TestSumOverEmptyOperandIsNeverASilentZero` prove the five unimplemented functions fail only at
  evaluation, as located errors, never at load and never as a plausible value.

### AC24 — the red-proofs

All five captured this run, at the tip commit `10327a4` before the corresponding obligation was
wired (D-000.30's ordering trap: the historical "before" state for the module-move items is
reconstructed structurally — as a permanent, repeatable in-memory-mutation test — rather than by a
literal separate pre-move commit, since this story lands the whole obligation in one pass; each
red-proof below is a standing test, not a one-time manual observation, so the window it proves stays
provably open on every future run too, not only this one).

- **AC21 red (DW-5's obligation).** Before `folio_expr_validate.go` existed, no code resolved a
  derivable `bind` to its `footerOf` value at all — `folio.ParseTemplate` returned only
  `*Template{doc: doc}` (`folio.go`, pre-3.2). `TestParseTemplateDerivesFooterOfShape1`/`Shape2`
  would have failed with "expected a derived footerOf for column e3: not found" against that
  version. Verified by inspection of the pre-change `folio.go` (git history) and by the fact these
  tests could not even compile before `folio_expr_validate.go`/`Template.derivedFooters` existed.
- **AC22 red (D-3.2.1's forcing function), BOTH mutations.** (1) Run before the move: `Decimal` lived
  only in `internal/bind`; `TestExactlyOneDecimalDeclarationInTheModule` against that tree would have
  failed naming `internal/bind`, not `internal/expr` — reproduced structurally, since the guard now
  asserts the post-move location and the pre-move source is preserved in git history (`internal/bind/
  decimal.go` at `10327a4`). (2) The SECOND mutation — a duplicate `Decimal` left behind in `bind`
  after the move — is captured as a STANDING, re-runnable test:
  `TestDecimalUniquenessRedProof` (`internal/expr_arch_test.go`) injects a second `type Decimal`
  into an in-memory copy of `internal/bind/value.go` and asserts `decimalDeclLocationsWithOverride`
  observes two locations. This is also D-3.2.1's own required addition: the four testdata decoys
  under `testdata/arch/reducer-inventory/` CANNOT red-prove this half (walkGoFiles skips
  `testdata/` by design, confirmed by `TestDecimalDeclarationScanSkipsTestdataDecoys`), so this
  mutation against a REAL package is what actually proves the guard's sensitivity.
- **AC23 red (DW-8's other half).** `parseBindingPath` was declared at `internal/bind/text.go:328`
  and `isValidIdent` at `:345` at `10327a4` (verified: `git show 10327a4:folio-go/internal/bind/
  text.go | grep -n 'func parseBindingPath\|func isValidIdent'` — lines confirmed against the
  pre-change file). `TestParseBindingPathAndIsValidIdentAreAbsent` against that tree would have
  failed naming exactly those two locations.
- **AC8/C1 red.** `TestExprFunctionTableRedProofNinthEntry` (`internal/expr_arch_test.go`) injects a
  ninth entry into an in-memory copy of `table.go` and confirms the extracted set no longer matches
  the pinned eight — a standing, re-runnable proof that a ninth entry alone (without touching this
  guard's expected set) goes red.
- **AC17 red (F6's `SumDecimals` hazard).** Actually performed, not merely reasoned about. A copy of
  `internal/expr/eval.go` was saved (`cp`, per this project's own red-proof hygiene), and `evalCall`
  was temporarily mutated to add, ahead of the `!entry.implemented` check: `if entry.name == "sum" {
  v, _ := SumDecimals(nil); return Value{Kind: KindNumber, Num: v}, nil }` — the cheapest, most
  obvious wiring of `sum` straight to the kernel it will eventually use, over an empty operand slice
  regardless of the argument. `go test -run TestSumOverEmptyOperandIsNeverASilentZero
  ./internal/expr/...` was run against the mutated source and OBSERVED FAILING:
  `eval_test.go:243: AC17: sum() must return an error, not a plausible value` — i.e. the mutation
  produced `Decimal{Coefficient: 0, Exponent: 0}` with `err == nil`, exactly F6's hazard. The file
  was then restored from the saved copy (`cp`, never `git checkout`) and `diff` confirmed
  byte-identical to the original; the full `internal/expr` suite (79 tests) was re-run green
  afterward. The shipped `evalCall` never special-cases `sum` at all — every unimplemented entry,
  `sum` included, returns the located "not yet implemented" error unconditionally, never reaching
  `SumDecimals`.

### AC28-32 (DW-8's move)

- **AC28.** `Decimal`, `NewDecimal`, `SumDecimals`, `AvgDecimals` and their unexported bounds moved;
  `internal/bind` imports them back via `github.com/panitw/folio/folio-go/internal/expr` (legal, 4→3).
- **AC29.** `decimal_test.go`/`reduce_test.go` moved with the code, unedited apart from the package
  clause; no unexported bound was hardcoded (`maxDecimalCoefficientDigits`,
  `maxDecimalExponentMagnitude`, `avgExtraScale` are all still referenced by symbol).
  **`golden_test.go`/`testutil_test.go` also moved** (see Files, above) — the same D-000.34 shape,
  found while wiring the move (the compiler caught it: these files failed to build in `bind` once
  `Decimal`/`SumDecimals`/`AvgDecimals` left the package).
- **AC30.** `internal/reducer_inventory_arch_test.go` — byte-identical, SHA-256 confirmed above.
- **AC31.** `internal/bind/resolution_roots_arch_test.go` observes both `params` and `row` roots (plus
  `data`) after the matcher deletion — unedited; `lookupBound`'s call SITES (three: data/params/row
  in `exprResolver.Resolve`, `internal/bind/text.go`) are unchanged in shape/signature-position
  (rootName still argument index 4), only its return type changed
  (`(*string, error)` → `(expr.Value, error)`), which the AST guard does not key on. Observed set
  logged by the guard's own `t.Logf`: `[data params row]`.
- **AC32.** F10's three sites (`internal/bind/text_test.go`, `internal/bind/spans_test.go`,
  `render_bind_test.go`) each re-pointed or confirmed to assert a message that DISTINGUISHES
  "parsed, then unimplemented" from "rejected as unsupported syntax" — `spans_test.go`'s
  `{{sum(x)}}` case required no edit (it only asserts `BindText`/`BindTextSpans` agree with each
  other, not which wording either produces, so it stayed non-vacuous under the new behaviour without
  a change).

### Cross-story handoff (per instructions, written here explicitly)

**Story 3.4 must cite `avgExtraScale` BY SYMBOL** (`internal/expr/reduce.go`), never as the literal
`4` — D-3.1a.1's forcing function (no `formatNumber` pattern may request more fractional digits than
`avg` produces) depends on the reference staying live through the `internal/expr` move.

**FLAG-2's forward obligation** (Decision 3): a PATH argument's runtime kind (e.g.
`sum(customer.name)`, a string collection) is NOT checked by this story — owed at evaluation by
Story 3.3 (`sum`/`count`/`avg`) and Story 3.4 (`formatDate`/`formatNumber`), named here per
D-000.60 so it does not ship as an open fence.

### Gates, measured this run (D-000.26, scope and flags stated)

| scope | invocation | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go build ./...` | success |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | no issues |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 ./...` | **712 PASS · 1 FAIL · 1 SKIP** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go build ./...` | success |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | no issues |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 ./...`, **working checkout** | **89 PASS · 3 FAIL** |

The single `folio-go` FAIL is `TestCorpusMeetsP6ExerciseFloors` — **REQUIRED red**, stats
byte-identical to baseline: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. Not touched.

The three `lint` FAILs are the **known-environmental (DW-19)** trio: `TestManifestUpToDate`,
`TestResolveAssetsIncludesWordlist`, `TestFontsAssetsNoticeRemovalRedProof` — the asset resolver
walks the gitignored `.font-sources/` directory in a working checkout. Not re-measured in a clean
worktree this run (carried from the baseline's own 92/0 clean-worktree figure; nothing this story
touches interacts with that resolver).

`folio-go` PASS count rose from the baseline's 640 to 712 (+72), reflecting the new test suite added
across `internal/expr` (parser/table/check/eval/footer, 75 tests), `internal/expr_arch_test.go` (7
guards), `folio_expr_validate_test.go` (9 tests), and re-pointed/added cases in `internal/bind` and
`render_bind_test.go` — net of the handful of pre-existing tests that were merged or renamed during
the F10 re-point (no test was deleted without a replacement covering the same property).

No cross-target hash matrix run — due at Epic 3's close (D-000.4), and this story introduces no new
source of cross-target divergence (a parser over `[]byte` producing `Decimal`, float already banned
twice over — AC1/AC2/D-3.2.2, D-3.1a.1).

### Acceptance criteria — how each is met

AC1/AC2 — `internal/expr`'s hand-written parser (`parser.go`), no dependency. **Corrected by the
finisher (QA Finding 1, Blocker): this sentence, as originally written, was itself the defect — it
claimed AC2's guard lived in `TestExprTableHasNoExportedRegistrationPath`'s "sibling guards" and in
"the existing module-wide 'no third-party dependency' lint rule"; neither ever read an import set, and
no such lint rule exists.** AC2's guard did not exist at all until the finisher built it:
`TestExprImportSetIsStdlibOrFirstPartyOnly` (`internal/expr_arch_test.go`), reusing `walkGoFiles`
(R5), with its own red-proof (`TestExprImportSetRedProof`) calling the exact extraction function
(`exprImportViolationsFromFile`) the shipped guard calls — never a duplicate (D-000.24, applying
Finding 4's own lesson from the start). AC3 — `parser_test.go`'s accepted/rejected table, including
the nesting example. AC4 — `expr.IsReserved`/`ScanPlaceholders` (`IsReserved` unexported behind an
accessor by the finisher, QA Finding 10 — see Finding Resolutions), checked before any parse attempt
in both `bind.Resolve` and `folio`'s load-time walk (AD-13, one definition). AC5-AC9 — `table.go`'s
`functionTable`, `TestExprFunctionTableIsExactlyEight`/`TestExprTableHasNoExportedRegistrationPath`/
`TestExprFunctionTableRedProofNinthEntry`/`TestAggregationEntriesDeclareDecimalReturn`. AC10/AC11 —
`check.go`, `TestCheckArity`/`TestCheckUnknownFunction`/`TestCheckLiteralArgumentKind`, enforced at
load via `folio_expr_validate.go`. AC12-AC14 — `eval.go`'s `evalUpperLower`/`evalIf`,
`eval_test.go`'s full case table. AC15-AC18 — `evalCall`'s unimplemented branch,
`TestUnimplementedFunctionsAreLocatedErrors`/`TestSumOverEmptyOperandIsNeverASilentZero`/
`TestUnimplementedAndUnknownFunctionErrorsAreDistinguishable`. AC19/AC20 — `parser.go`'s syntax
errors (plain Go errors, no diagnostic code), `TestParseRejectedForms`,
`TestParseTemplateRejectsSyntaxErrorInTextValue`. AC21-AC27 — `folio_expr_validate.go`/
`footer.go`, `folio_expr_validate_test.go`, `internal/template/footer_test.go`'s re-framing. AC28-32
— see above. AC33 — gates table above.

### Finisher pass (2026-08-25)

24 findings triaged (the review's own summary line says 3 Blocker/4 Major/10 Minor/6 Nit = 23; the
findings as actually numbered in the file are 3 Blocker (1-3), 4 Major (4-7), **11** Minor (8-18) and
6 Nit — 24, not 23. Recounted directly from the `**Severity**` lines rather than trusted from the
summary — see ## Finding Resolutions below for the full per-finding triage and evidence.

**Disposition: 21 FIX, 1 resolved-by-orchestrator (outside this story's tree), 2 DISMISS.** No DEFER
— every finding was either cheap enough to fix now or already settled elsewhere. Findings 1-3
(Blocker) and 4, 5, 7 (Major) were fixed first, in the order specified; Finding 6 (Major) was resolved
by the orchestrator at commit `5c94bae`, before this pass began, and is not touched here.

**Files this pass modified or added, beyond the developer's own list above:**
- `folio-go/internal/expr/parser.go` — modified: `maxCallDepth` recursion bound (Finding 3), Check-time
  numeric-literal grammar tightening (Finding 7), escaped-quote diagnostic (Finding 11), rune decoding
  and position fixes across four diagnostic sites, unreachable block deleted (Finding 13).
- `folio-go/internal/expr/parser_test.go` — modified: red-proofs and cases for Findings 3, 11, 13.
- `folio-go/internal/expr/check.go` — modified: `checkNumberLit`, Check-time literal bound (Finding 7).
- `folio-go/internal/expr/check_test.go` — modified: cases for Finding 7; literal 8-name list (Nit 2).
- `folio-go/internal/expr/eval.go` — read only, for the Finding 8/M1 re-verification; restored
  byte-identical (temporary mutation only, per this project's red-proof hygiene).
- `folio-go/internal/expr/eval_test.go` — modified: AC17 assertion moved onto the reachable success
  path (Finding 8), seven discarded-parse-error sites now fail explicitly (Finding 15), two new tests
  (Finding 16).
- `folio-go/internal/expr/scan.go` — modified: `IsReserved` accessor replaces the exported mutable map
  (Finding 10), unterminated-placeholder error now carries position and text (Finding 13), known
  quote-blindness limitation documented (Finding 11, DISMISS half).
- `folio-go/internal/expr/scan_test.go` — **new**: pinning guard for the reserved set (Finding 10).
- `folio-go/internal/expr_arch_test.go` — modified: AC2's import-scope guard built from scratch
  (Finding 1), AC22's red-proof threaded through the shipped extractor instead of a duplicate
  (Finding 4), AC6's structural mutation guard (Finding 9).
- `folio-go/internal/bind/text.go` — modified: `Resolve` re-pointed onto `expr.ScanPlaceholders`,
  local `reservedPlaceholders` alias removed (Finding 5), `lookupBound`'s unreachable tail now an
  explicit error (Nit 3), stale `parseBindingPath` comment corrected (Nit 1).
- `folio-go/internal/bind/text_test.go` — modified: new tests for Findings 3, 16.
- `folio-go/internal/bind/value.go` — modified: stale comment (Nit 1).
- `folio-go/folio_expr_validate.go` — modified: `visibleIf` now walked as a bare expression, its own
  check function (Finding 12); temporary write-back mutation applied and reverted for Finding 2's
  red-proof (byte-identical restore, `/usr/bin/diff` confirmed).
- `folio-go/folio_expr_validate_test.go` — modified: `TestParseTemplateFooterOfDerivationNeverWritesBack`
  rewritten to inspect the document `ParseTemplate` actually derived over (Finding 2); new tests for
  Findings 3, 7, 12, 14.
- This story file — AC19, AC24 and AC27 text corrected to match what shipped (Findings 17, 18, Nit 4);
  the AC1/AC2 Delivery Log sentence corrected (Finding 1); plain-terms opener rewritten; Finding
  Resolutions and this Delivery Log addendum appended; Status set to `done`.

**Not touched, per explicit instruction:** `docs/expression-reference.{md,html}` — Finding 6's factual
correction and both files' commit are the orchestrator's own work at `5c94bae`, outside this story's
tree; Nits 5 and 6 (both about that same file) are DISMISSed here as out of this story's scope for the
same reason, not fixed.

**Gate results after the finisher's fixes, with scope and flags (D-000.26):**

| scope | invocation | before finisher | after finisher |
|---|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go build ./...` | success | **success** |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | clean | **clean** |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting every `--- PASS`/`--- FAIL`/`--- SKIP`, subtests included | 712 PASS · 1 FAIL · 1 SKIP | **740 PASS · 1 FAIL · 1 SKIP** (+28, all this pass's new tests/subtests) |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go build ./...` | success | **success** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | clean | **clean** |
| `lint/` | same test invocation, working checkout | 89 PASS · 3 FAIL | **89 PASS · 3 FAIL** (unchanged — this pass touches no `lint/` source) |

`TestCorpusMeetsP6ExerciseFloors` remains the single folio-go FAIL, stats byte-identical
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — confirmed unchanged, never touched. The one
SKIP (`TestXrefEntriesRejectsMalformedSubprocess`) is unchanged. The three `lint` FAILs are the
unchanged DW-19 trio (`TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`,
`TestFontsAssetsNoticeRemovalRedProof`), all three still failing for the same `.font-sources` cause —
this pass added no lint-side source, so this figure was not independently re-verified in a clean
worktree this run (carried from the story's own 92/0 clean-worktree figure, itself carried from 3.1a).

**Cross-target matrix: NOT run.** D-000.4's per-epic cadence is unchanged by this pass: the matrix is
due at **Epic 3's close**, not at any one story, and this pass introduces no new source of cross-target
divergence (a bounded recursive-descent parser and additional Go-only diagnostics/AST guards, still
over `[]byte` producing `Decimal`, float already banned twice over). Written here explicitly so this
does not read as coverage this pass silently claims. `epic-3` stays `in-progress` in
`sprint-status.yaml`, per instruction — untouched by this pass beyond the `3-2` story key itself.

**Red-proof discipline applied throughout (D-000.30).** Every FIX with an observable failure mode was
red-proved: the BLOCKER 2 write-back defect was reinstalled and reddened, then restored byte-identical
(`/usr/bin/diff`); the BLOCKER 3 recursion bound was proved against the reviewer's own ~800,000-nested-
call, ~1.6MB reproduction, through all three reachable entry points (`Parse`, `folio.ParseTemplate`,
`bind.BindText`); the MAJOR 4 fix was proved by re-running the reviewer's own M5 mutation against the
now-shared extraction function and observing both the production guard and its red-proof move together
(the standing test reddened; it had not before); the MAJOR 7 fix was proved against the reviewer's own
oversized-literal and leading-zero reproductions, at Check time and through `folio.ParseTemplate`. No
scratch probe or throwaway file was left in the tree; `git status --porcelain` before this pass's
commit shows only the files listed above.

---

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer (adversarial, context-isolated from the developer)
- **Date:** 2026-08-25
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 3
- **Majors:** 4
- **Minors:** 10
- **Nits:** 6

### Gates — independently re-measured, counted from raw `go test -json`, not from a summary line

| scope | invocation | reviewer's measurement | story's claim | verdict |
|---|---|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go build ./...` | success | success | matches |
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go vet ./...` | clean | clean | matches |
| `folio-go/` | `go test -count=1 -json ./...`, counting every `pass`/`fail`/`skip` event carrying a `Test` field | **712 PASS · 1 FAIL · 1 SKIP** | 712 · 1 · 1 | **matches exactly** |
| `lint/` | `go test -count=1 -json ./...`, working checkout | **89 PASS · 3 FAIL · 0 SKIP** | 89 · 3 | **matches exactly** |

- **CRITICAL INVARIANT HOLDS.** `TestCorpusMeetsP6ExerciseFloors` is the single `folio-go` FAIL, stats
  byte-identical to the pin: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. The one SKIP is
  `TestXrefEntriesRejectsMalformedSubprocess`. Not touched.
- **The three `lint` FAILs are exactly the DW-19 trio**, and all three fail for the *same*
  environmental cause, not a different one:
  `.font-sources: contains a committed font binary but no LICENSE* file (AC25, AD-26)`. No fourth
  failure, no substitution. 89 + 3 = 92 reconciles with the carried clean figure.
- **The +72 PASS jump is honest.** No test was removed or silently renamed away: the four moved
  `internal/bind` → `internal/expr` test files differ from `10327a4` **only** in the `package` clause
  (AC29 below), and both renamed tests kept a live, strengthened body.
- No cross-target matrix run — correct, due at Epic 3's close (D-000.4).

### What the reviewer verified by mutation

Every mutation was applied to a working copy, observed, and reverted **from a saved copy (`cp`, never
`git checkout`)**; each touched file was re-verified byte-identical by SHA-256 afterward and the whole
working-tree file set re-confirmed against the pre-review snapshot.

| # | mutation | expected | observed | verdict |
|---|---|---|---|---|
| M1 | wire `sum` straight to `SumDecimals(nil)` in `evalCall`, ahead of the `!entry.implemented` check | AC17 reddens | **RED** — `eval_test.go:243: AC17: sum() must return an error, not a plausible value`, plus 2 more tests; 709 PASS · 4 FAIL | **AC17's red-proof is REAL and reproduces verbatim** |
| M4 | `DeriveFooterOf` shape 1 drops the collection prefix (still *derivable*, wrong value) | AC21's derivable arm reddens | **RED** — `footer_test.go:16: FooterOf = "amount", want transactions.amount` **and** `TestParseTemplateDerivesFooterOfShape1` | **the derivable arm is non-vacuous** |
| M5b | a real second `type Decimal` written to `internal/bind/value.go` on disk | AC22 reddens | **RED** — `expected exactly ONE "Decimal" declaration in the module, found 2: [internal/bind internal/expr]` | **the shipped uniqueness guard is sound** |
| M6 | a ninth entry added to `functionTable` | AC8/C1 reddens | **compile error** (`table.go:95:2: index 8 is out of bounds (>= 8)`, from `[8]funcEntry`); after widening to `[9]`, **two** independent guards go red | **AC8 is doubly pinned — stronger than the AC asked for** |
| M3 | `validateTableColumns` writes the derived `footerOf` **back into** `template.Document` | AC21's P3 clause reddens | **GREEN — and the whole folio-go tree stayed at 712 · 1 · 1** | **Finding 2 (Blocker)** |
| M5 | cripple the *shipped* `decimalDeclLocations` so it can never report >1 location | AC22's red-proof reddens | **GREEN**, and the red-proof still logged *"the uniqueness guard would catch it"* | **Finding 4 (Major)** |

The reviewer also ran the parser against adversarial input from a **throwaway copy of the module in
`/tmp`** (deleted afterward; nothing was added to this repository), reaching it through the real public
API `folio.ParseTemplate`. That is where Findings 3, 7, 11 and 13 come from.

---

## Findings

### Finding 1: AC2's import guard — D-3.2.2's "mechanical half" — was never built

- **Severity**: Blocker
- **Category**: AC Conformance
- **Location**: absent; expected in `/Users/panitw/Projects/folio/folio-go/internal/expr_arch_test.go` (which houses AC6/AC7/AC8/AC22/AC23)
- **Observation**: AC2 mandates an AST guard asserting every import in every non-`_test.go` file under
  `folio-go/internal/expr/` resolves to the standard library or to
  `github.com/panitw/folio/folio-go/…`, carrying (i) a coverage witness stating **in these words** that
  it covers *"`internal/expr`'s import set, not the module's"* (D-000.23, R4), (ii) a presence
  precondition where **zero files parsed is a `t.Fatal`** (D-000.9), and (iii) a doc comment citing
  **D-3.2.2**. **No such test exists in either module.** An exhaustive search for a `_test.go` file that
  reads `.Imports` and is scoped to `internal/expr` returns nothing; the only `_test.go` files
  mentioning `internal/expr` are `internal/expr_arch_test.go`, `internal/reducer_inventory{,_arch}_test.go`,
  `internal/template/footer_test.go` and `internal/expr/table_test.go`, none of which inspects an import
  set. The Delivery Log's AC1/AC2 sentence is garbled and cites
  `TestExprTableHasNoExportedRegistrationPath`'s *"sibling guards"* plus *"the existing module-wide 'no
  third-party dependency' lint rule"* — the first never reads imports, and no such lint rule exists
  (`lint/internal/rules/forbiddenimports.go` enforces AD-1's stdlib ban, a different property).
- **Impact**: AC2 is the one assertion that turns D-3.2.2 (*CEL and every general-purpose expression
  library REJECTED*) from a memory into a fact. Its absence means the next author who adds
  `require github.com/google/cel-go` and imports it from `internal/expr` trips **no guard in this
  repository**. The *property* is currently true — `internal/expr`'s complete non-test import set is
  `fmt`, `strings`, `math/big` and `github.com/panitw/folio/folio-go/internal/template`, and
  `go.mod`/`go.sum` are byte-unchanged from `10327a4` in **both** modules — so this is an unbuilt guard,
  not a breached invariant. But AC2 is ticked `[x]`, recording a mechanism that does not exist.
- **Suggested Resolution**: Build the guard in `internal/expr_arch_test.go` reusing `walkGoFiles` (R5),
  with AC2's verbatim coverage-witness wording, the zero-files `t.Fatal`, and a doc comment citing
  D-3.2.2's reason (the numeric model, not the parser). Red-prove it by injecting a non-first-party
  import into an in-memory copy of an `internal/expr` file **through the same extraction function the
  shipped guard calls** — see Finding 4 for why that clause matters.
- **Related AC**: AC2 (and AC1's mechanical half; D-3.2.2, D-000.23, D-000.9, R4)

### Finding 2: the R2 / D-1.4.3-P3 guard cannot fail — a derived `footerOf` written back into the Document is invisible to the entire test suite

- **Severity**: Blocker
- **Category**: Tests (vacuous assertion on a named invariant)
- **Location**: `/Users/panitw/Projects/folio/folio-go/folio_expr_validate_test.go:104-143`
- **Observation**: `TestParseTemplateFooterOfDerivationNeverWritesBack` carries three assertions and
  **none can observe the defect it is named for**:
  1. **Lines 110-125 are a tautology.** `roundTripped` and `canonical` are *both*
     `template.SerializeDocument(doc)` over the *same* `doc`:
     ```go
     roundTripped, err := template.SerializeDocument(doc)
     ...
     canonical, err := template.SerializeDocument(doc)   // same function, same argument
     ...
     if string(roundTripped) != string(canonical) { t.Fatalf("Serialize(Parse(b)) is not stable — P3 violated") }
     ```
     This compares a pure function's output with itself. It is not `Serialize(Parse(b)) == b`, and the
     comment above it (*"a second parse/serialize pass of the ORIGINAL"*) describes a second parse the
     code never performs.
  2. **Line 126's `footerOf` check reads the wrong document.** `doc` comes from
     `template.ParseDocument(original)` at line 106 — *before* `ParseTemplate` is called at line 132 and
     entirely independent of it. A write-back inside `ParseTemplate` never reaches `doc`.
  3. **Lines 139-142 read a third, freshly parsed document.** `redoc` is another
     `template.ParseDocument(original)`; a fresh parse of the original bytes is footerOf-free by
     construction, whatever `ParseTemplate` did to *its own* Document.

  The test never inspects the `template.Document` that `ParseTemplate` actually derived over.
  **Proven by mutation (M3):** `validateTableColumns` was changed to write the derived value back —
  `tbl.Columns[ci].FooterOf = template.Presence[string]{Set: true, Value: result.FooterOf}`, the precise
  defect R2 exists to prevent. This test **passed**, and the full folio-go suite was unchanged at
  **712 PASS · 1 FAIL · 1 SKIP**, the only FAIL still being the required-red
  `TestCorpusMeetsP6ExerciseFloors`. Nothing in the module detects it.
- **Impact**: AC21's final clause requires *"a test asserts `Serialize(Parse(b)) == b` still holds
  byte-for-byte for a document that omitted `footerOf`"*. That test does not exist in any functioning
  form, and AC21 is ticked `[x]`. D-1.4.3's P3 fixed point is the definition of canonical for this
  format; the harm class is that **every ordinary document legitimately omitting `footerOf` would start
  serializing with it**, silently changing canonical bytes for the whole corpus. The shipped code is
  correct — `validateAndDeriveExpressions` returns a map and never mutates `doc` — so this is a hollow
  guard over correct behaviour, not a live regression. But R2 is exactly the invariant a later refactor
  breaks by accident, which is why the story made it blocker-class.
- **Suggested Resolution**: Assert P3 as a real fixed point over the document `ParseTemplate` derived
  from. Two changes, neither cosmetic: (a) compare `SerializeDocument(ParseDocument(b))` against a
  canonical `b` (or against a **second, independent** parse/serialize pass) rather than serializing one
  document twice; and (b) obtain the Document `ParseTemplate` actually walked — via the returned
  `*Template`'s own `doc`, or by re-serializing after the derivation — so a write-back is observable.
  Then red-prove with exactly M3's mutation and confirm it goes red.
- **Related AC**: AC21 (final clause), R2, D-1.4.3 P3, AD-13

### Finding 3: unbounded parser recursion — untrusted document text kills the process with an unrecoverable `fatal error: stack overflow`

- **Severity**: Blocker
- **Category**: Security & data handling / Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/parser.go:219, 232, 249, 274, 280` (the `parsePrimary → parseIdentLed → parseCall → parsePrimary` cycle); reachable from `/Users/panitw/Projects/folio/folio-go/folio_expr_validate.go:62` (load) and `/Users/panitw/Projects/folio/folio-go/internal/bind/text.go:196` (render)
- **Observation**: The recursive-descent cycle has **no nesting-depth counter and no input-length
  bound**. Reproduced by the reviewer through the real public API — a single text element whose `value`
  is `{{` + `a(` × 800,000 + `}}` (about 1.6 MB, an unremarkable size for a `.folio` file) passed to
  `folio.ParseTemplate`:
  ```
  runtime: goroutine stack exceeds 1000000000-byte limit
  fatal error: stack overflow
  goroutine 21 [running]:
    …internal/expr.(*parser).parseCall      parser.go:274
    …internal/expr.(*parser).parseIdentLed  parser.go:249
    …internal/expr.(*parser).parsePrimary   parser.go:232
    …internal/expr.(*parser).parseCall      parser.go:280   (repeating)
  ```
  A `defer func(){ recover() }()` wrapped around the `ParseTemplate` call **did not fire** — this is a
  runtime *throw*, not a panic, so it is uncatchable by design. The process exits. At 200 KB of the same
  input the parser returns an ordinary error, so the degradation is silent right up to the cliff: roughly
  1 GB of goroutine stack is committed for ~2 MB of input, a ~500× amplification.
- **Impact**: This is the strongest possible violation of the story's own diagnostic discipline. AC11
  requires *"not a panic"*; AC20 requires *"plain Go errors, named fields, element id"*; AD-14 requires
  located diagnostics. A fatal stack overflow is worse than a panic: there is no error, no element id, no
  diagnostic, and no opportunity for a caller to recover — the rendering process simply dies. Both
  entry points are reachable from author-supplied content: `folio.LoadTemplate`/`ParseTemplate` on a
  `.folio` file at load, and `bind.BindText` at render. For a product whose whole premise is that
  designers author templates, a template that terminates the engine is a denial-of-service, not a syntax
  error. No AC anticipated this because no AC bounded the grammar's depth — but AC11/AC20/AD-14 all
  forbid the observed outcome.
- **Suggested Resolution**: Add a depth counter to the `parser` struct, incremented in `parseCall` (or
  `parsePrimary`) and checked against a small constant — one or two levels beyond what AC3 requires, so
  the limit is well clear of legitimate use — returning an ordinary located error naming the element id
  and the offending text, exactly like every other syntax error. Belt-and-braces: reject a placeholder
  body over a modest byte length before parsing. Cover both with tests, and add a rejected-form case to
  `TestParseRejectedForms` so the bound is pinned rather than incidental.
- **Related AC**: AC11, AC19, AC20, AD-14, AD-9

### Finding 4: AC22's red-proof exercises a hand-duplicated copy of the extractor, not the guard that ships

- **Severity**: Major
- **Category**: Tests (red-proof does not exercise the shipped path; D-000.24)
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr_arch_test.go:360-424` (`TestDecimalUniquenessRedProof` and `decimalDeclLocationsWithOverride`), against the shipped `decimalDeclLocations` at `:268-289`
- **Observation**: The story correctly identifies (F4, AC24) that the four decoy `type Decimal`
  declarations under `testdata/arch/reducer-inventory/` **cannot** red-prove uniqueness — Go excludes
  `testdata/` from the build and `walkGoFiles` skips it — and offers `TestDecimalUniquenessRedProof` as
  *"what actually proves the guard's sensitivity"*. It does not.
  `decimalDeclLocationsWithOverride` (`:395-424`) is a **hand-copied duplicate** of
  `decimalDeclLocations` (`:268-289`): the same `GenDecl`/`token.TYPE`/`TypeSpec`/`ts.Name.Name ==
  "Decimal"` extraction, written out a second time. The red-proof calls the copy;
  `TestExactlyOneDecimalDeclarationInTheModule` calls the original. Nothing ties them together.
  **Proven by mutation (M5):** the *shipped* `decimalDeclLocations` was crippled with
  `if len(locations) > 1 { locations = locations[:1] }` — AC22's entire cardinality half disabled — and
  both tests still **PASSED**, the red-proof printing, verbatim:
  `red-proof: a second real "Decimal" declaration is observed at [internal/bind internal/expr] — the
  uniqueness guard would catch it`. That sentence is false under the mutation, and the test asserts it
  anyway.
- **Impact**: The proof is decoupled from its subject: a future edit that narrows or breaks the shipped
  extractor leaves the red-proof green and self-congratulatory, so the uniqueness guard can go slack with
  a standing test claiming otherwise. Per D-000.24, an assertion that does not exercise the shipped path
  must be labelled a **forward guard**, not credited with a proof it lacks. Major rather than Blocker
  because the reviewer independently confirmed (M5b) the shipped guard **is** sound today: a real second
  `type Decimal` on disk produces `expected exactly ONE "Decimal" declaration in the module, found 2`.
- **Suggested Resolution**: Thread the override through the shipped function rather than duplicating it —
  give `decimalDeclLocations` an optional per-file source override (nil for the production guard) so both
  callers share one extraction body — then re-run M5 and confirm the red-proof now reddens. If the
  duplication is kept deliberately, retitle the test and its log line as a forward guard and stop
  asserting *"the uniqueness guard would catch it"*.
- **Related AC**: AC22, AC24 (second mutation), D-3.2.1, D-000.24, D-000.14

### Finding 5: AD-13 breached — `internal/bind` hand-rolls a second `{{ }}` tokenizer, and `scan.go`'s doc comment states the opposite

- **Severity**: Major
- **Category**: Convention (AD-13, one source of truth)
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/scan.go:35-40` (the claim) vs `/Users/panitw/Projects/folio/folio-go/internal/bind/text.go:169-183` (the second implementation)
- **Observation**: `scan.go`'s doc comment asserts:
  > *"ScanPlaceholders is the module's ONE tokenizer for `{{ ... }}` occurrences inside a text string
  > (AD-13): **both** internal/bind's render-time resolver … **and** folio's load-time syntax/arity scan
  > call this, rather than each re-implementing the `{{`/`}}` search."*

  `internal/bind` does not call it. The only two call sites in the module are
  `folio_expr_validate.go:51` and `internal/expr/footer.go:47`. `internal/bind/text.go:169-183` performs
  its own byte-identical scan — `strings.Index(rest, "{{")` → `strings.Index(afterOpen, "}}")` →
  `rest = afterOpen[end+2:]` → `strings.TrimSpace(inner)` → reserved check — duplicating
  `ScanPlaceholders`'s entire body inline. The story's AC-mapping line makes the same claim
  (*"AC4 — `expr.ReservedPlaceholders`/`ScanPlaceholders`, checked before any parse attempt in both
  `bind.Resolve` and `folio`'s load-time walk (AD-13, one definition)"*); only `ReservedPlaceholders` is
  genuinely shared.
- **Impact**: Two independently maintained tokenizers for the same syntax is precisely what AD-13 exists
  to prevent, and this story's own plain-terms section names the hazard (*"two readers of the same syntax
  will eventually disagree and the wrong one will win"*) as the reason for deleting the old matcher. They
  already differ in error wording, and any fix to the scan — for instance the quote-blindness in
  Finding 11, where `{{ f("a}}b") }}` is cut inside a string literal — must now be made twice or the two
  layers diverge on which templates are loadable versus renderable. A doc comment asserting single
  ownership that is false is worse than no comment.
- **Suggested Resolution**: Re-point `internal/bind/text.go`'s loop at `expr.ScanPlaceholders` (it already
  returns the `literal`/`trailing` segmentation the resolver needs for its `Substitution` bookkeeping), or
  correct the comment and the story's AC-mapping line to state that `bind` keeps its own scanner and why.
  A guard asserting only one `{{`-search exists in the module would make this durable.
- **Related AC**: AC4, AD-13, D-1.6.5

### Finding 6: two undisclosed author-facing artifacts shipped in `docs/`, unpinned by any test, one contradicted by this repo's own test

- **Severity**: Major
- **Category**: AC Conformance / Maintainability (AD-13, undisclosed scope)
- **Location**: `/Users/panitw/Projects/folio/docs/expression-reference.md` (180 lines; the defect at `:64-66`) and `/Users/panitw/Projects/folio/docs/expression-reference.html` (467 lines)
- **Observation**: Both files are new, untracked, and dated within this story's working session. **No AC
  asks for them, the Delivery Log's "Files" section does not list them, and the string
  `expression-reference` appears nowhere under `_bmad-output/`** — not in the story, not in the spec, not
  in the decision log. They are a third description of the expression language alongside
  `folio-format.md` and the code, and no test reads either. Concretely, the `.md` states:
  > *"You cannot name a region `params`, `page`, or `pages`. Those are reserved, and using one is
  > reported **when the template loads** rather than producing a surprise at render time."*

  This repository's own test asserts the opposite. `TestRenderTableRowAliasCollidesWithReservedNameFailsRender`
  (`folio-go/render_table_bind_test.go:127-150`) calls `ParseTemplate` on a template with `"as": "page"`
  and **fails the test if `ParseTemplate` returns an error** (`t.Fatalf("ParseTemplate: %v", err)`); the
  located error comes from `Render`. The doc promises load-time reporting *"rather than … at render
  time"*, and render time is exactly where it surfaces.
- **Impact**: AD-13 is "one source of truth". The `if(null)` silent-false ruling was made on the explicit
  basis that the behaviour *"produces no signal in the output, so the doc and its test are the only two
  places it exists"* — there are now three, and the third is unpinned and already carries a false claim
  about a different feature. An author-facing reference that misstates *when* an error is reported is
  worse than none. Undisclosed scope also means neither D-000.60's flag mechanism nor an AC walk would
  have caught it; it surfaced only in `git status`.
- **Suggested Resolution**: Either (a) declare the files in the Files section, correct the reserved-alias
  claim to say the collision is reported at **render**, audit the remainder against shipped behaviour
  (see Nit 6), and pin it to `folio-format.md` by test or generation so the two cannot drift; or (b)
  remove them from this story and route the author-facing reference to its own story with its own ACs.
  Do not leave an unowned, untested third source of truth in the tree.
- **Related AC**: none (undisclosed); AD-13, D-000.60, the `if(null)` owner ruling

### Finding 7: R3 breached — statically decidable numeric-literal defects survive load and fail at render

- **Severity**: Major
- **Category**: Correctness / AC Conformance
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/parser.go:175-212` (`lexNumber`), `/Users/panitw/Projects/folio/folio-go/internal/expr/check.go:23-34` (`Check`), `/Users/panitw/Projects/folio/folio-go/internal/expr/eval.go:24-29`
- **Observation**: R3 sets the boundary as *"syntax and arity at load; execution at evaluation"*, and
  AC19 requires a syntax error to fail **at load**. A numeric literal's bounds are decidable with no data
  at all, yet are not checked until evaluation. Verified directly:
  ```
  Parse("12345678901234567890123456789") -> ok      Check -> nil
  Eval  -> expr: element e1: invalid number literal "12345678901234567890123456789":
           expr: value "...": coefficient has 29 significant digits, exceeds 19 (does not fit int64)
  ```
  `Parse("1e999999999999999999999")` likewise parses and Checks clean. `NewDecimal` — the only thing that
  rejects them — is called from exactly one non-test site, `eval.go:25`, which runs at render.
  Separately, the parser accepts literals JSON rejects: `01`, `007` and `-01` all parse **and** Check
  clean, despite `ast.go:39-44` documenting `NumberLit` as *"a JSON number literal … the same shape
  `internal/template.SplitJSONNumber` accepts"*. `SplitJSONNumber` explicitly does not validate JSON
  grammar (it trusts `encoding/json` upstream), so `01` is silently normalised to `1` and nothing ever
  rejects it. Correctly rejected, for the record: `+5`, `.5`, `5.`, `1e`, hex.
- **Impact**: This is the failure mode R3 and F3 were written to separate. A template carrying an
  impossible literal **loads clean** — `folio.ParseTemplate` returns no error, so `folio.Validate` will
  eventually report nothing — and then dies mid-render, which for this product means a statement that
  fails while being produced rather than a template rejected when authored. It is also the inverse of the
  concession F3 forced: unimplemented *functions* must reach evaluation because the canonical golden
  needs them to; unimplemented functions are the only thing that should.
- **Suggested Resolution**: Call `NewDecimal` on every `NumberLit` inside `Check` (it needs no data and
  no resolver) and report a located error at load. Tighten `lexNumber` to JSON's own number grammar so
  `01`/`007`/`-01` are rejected at parse, matching what `ast.go` already claims. Add both to
  `TestParseRejectedForms`/`check_test.go` and to `TestParseTemplateRejectsSyntaxErrorInTextValue`'s
  neighbourhood.
- **Related AC**: AC3, AC10, AC19, R3, F3

### Finding 8: AC17's "not a Decimal of value zero" clause is unreachable dead code

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/eval_test.go:241-248`
- **Observation**: AC17 requires the test to assert *both* that `sum()` errors *and* **"specifically that
  it does not return a `Decimal` of value zero"**. The second assertion can never execute:
  ```go
  v, err := Eval(e, r, "e17")
  if err == nil { t.Fatal("AC17: sum() must return an error, not a plausible value") }   // Goexit
  if v.Kind == KindNumber && v.Num == (Decimal{Coefficient: 0, Exponent: 0}) { t.Fatal("AC17 RED-PROOF FAILED: …") }
  ```
  `t.Fatal` calls `runtime.Goexit()`, so line 245 is reached only when `err != nil` — and on every error
  path `Eval` returns the zero `Value{}`, whose `Kind` is `KindNull` (`ast.go:74`, `KindNull Kind = iota`),
  never `KindNumber`. Unsatisfiable in both branches.
- **Impact**: Presentation, not protection — the hazard **is** genuinely caught. The reviewer reproduced
  the developer's red-proof exactly (M1): wiring `sum` to `SumDecimals(nil)` yields
  `eval_test.go:243: AC17: sum() must return an error, not a plausible value` and three tests go red. The
  Delivery Log's account of this red-proof is accurate and was the story's best-evidenced claim. But the
  clause AC17 singles out as the distinguishing half is decorative, and an auditor would credit it.
- **Suggested Resolution**: Move the value check onto the success path — `if err == nil { … assert v is
  not {0,0}; t.Fatal(…) }` — or drop it and record that the `err == nil` assertion subsumes it.
- **Related AC**: AC17, D-000.25, D-000.61

### Finding 9: AC6's "no exported registration path" guard is a name deny-list, not a structural check

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr_arch_test.go:251-259`
- **Observation**: The guard rejects only the literal names `functionTable`/`FunctionTable`/`funcEntry`/
  `FuncEntry`, plus any lowercased exported name `strings.Contains`-ing `"register"` or `"addfunction"`.
  An exported `Install`, `Add`, `Define`, `SetFunction`, `Extend` or `WithFunction` that appends to or
  replaces `functionTable` passes untouched. AC6's own words are *"an AST guard asserts `internal/expr`
  exports no function that adds to, mutates, or replaces the table"* — a property the AST can decide
  directly (does any exported function's body assign to, or `append` into, `functionTable`?), which the
  guard does not attempt.
- **Impact**: The property holds today — the reviewer enumerated `internal/expr`'s complete exported
  surface (`Expr`, `PathExpr`, `StringLit`, `NumberLit`, `CallExpr`, `Kind`, `Value`, `Resolver`, `Check`,
  `Decimal`, `NewDecimal`, `Eval`, `DerivedFooter`, `DeriveFooterOf`, `Parse`, `SumDecimals`,
  `AvgDecimals`, `ReservedPlaceholders`, `Placeholder`, `ScanPlaceholders`, `LegalFunctionNames`) and
  found no mutator; `[8]funcEntry` plus two set-equality guards make a covert ninth very hard. But this
  guard would not be what catches it.
- **Suggested Resolution**: Walk exported `*ast.FuncDecl` bodies for an `AssignStmt`/`append` targeting
  `functionTable`; keep the name check as a cheap second line.
- **Related AC**: AC6, C1

### Finding 10: `expr.ReservedPlaceholders` is an exported, mutable, unpinned map guarding AD-4

- **Severity**: Minor
- **Category**: Security & data handling / Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/scan.go:17-20`, aliased at `/Users/panitw/Projects/folio/folio-go/internal/bind/text.go:60`
- **Observation**: `var ReservedPlaceholders = map[string]bool{"page": true, "pages": true}` is exported
  and mutable. Go maps are reference types and `internal/bind` binds the *same* map
  (`var reservedPlaceholders = expr.ReservedPlaceholders`), so any importer executing
  `expr.ReservedPlaceholders["balance"] = true` or `delete(expr.ReservedPlaceholders, "page")` changes
  AD-4's fence globally, for both the load-time walk and the render-time resolver, at run time. No test
  pins the map's size or contents — searches for `len(reservedPlaceholders)` and
  `len(expr.ReservedPlaceholders)` return nothing. `text.go:71`'s own comment — *"A 'page' NAMESPACE is
  precisely a THIRD entry here"* — names the hazard while leaving the door open.
- **Impact**: AD-4 is described in this story as absolute (*"no `page` namespace exists for expressions
  to reach"*). Centralising the definition for AD-13 was right; exporting it as a writable map hands
  every importer the third entry the comment warns about, with nothing to catch it.
- **Suggested Resolution**: Make the map unexported behind `func IsReserved(name string) bool`, or add a
  set-equality guard pinning the keys to exactly `{page, pages}` with a `t.Fatal` on an empty map — the
  discipline AC7 already applies to the function table.
- **Related AC**: AC4, AD-4, AD-13, D-1.6.5

### Finding 11: string literals cannot contain a quote, and the resulting diagnostics misdirect

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/parser.go:158-169` (`lexString`); compounded at `/Users/panitw/Projects/folio/folio-go/internal/expr/scan.go:64`
- **Observation**: `lexString` scans to the first `"` with no escape handling at all, so no string literal
  in the grammar can ever contain a double quote. Verified:
  ```
  Parse(`"a\"b"`)        -> unexpected trailing content "b\"" at position 4
  Parse(`upper("a\"b")`) -> expected ',' or ')' at position 10, got "b"
  ```
  Neither message mentions escapes; the author is told something unrelated to the real cause. A backslash
  elsewhere survives literally into `Value` (`"a\b"` → `Value: a\b`), so `Value`/`Raw` remain
  self-consistent and `Raw` is exact source — the defect is the missing capability plus the misdirecting
  diagnostic, not a mis-parse. Separately, `scan.go:64`'s `strings.Index(afterOpen, "}}")` is quote-blind,
  so `{{ f("a}}b") }}` is cut inside the string literal and surfaces as an unterminated-string error.
- **Impact**: AC3 lists *"a double-quoted string literal"* as an accepted form without qualification, and
  `formatDate`/`formatNumber` patterns are string literals, so the limit will be met in ordinary use. The
  bigger cost is diagnostic: AD-14's located-error discipline is undermined when the message names the
  wrong problem. Note that per Finding 5 the quote-blind `}}` search now exists in two places.
- **Suggested Resolution**: Either support `\"` (and `\\`) in `lexString`, or detect an escaped quote and
  emit an explicit *"string literals do not support escape sequences"* error naming the position — and
  document the choice in `folio-format.md`'s `### Expressions` section. Add both to
  `TestParseRejectedForms`.
- **Related AC**: AC3, AC19, AD-14

### Finding 12: `Element.VisibleIf` is never walked by the load-time expression check

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `/Users/panitw/Projects/folio/folio-go/folio_expr_validate.go:22-43` (`validateAndDeriveExpressions`), against `/Users/panitw/Projects/folio/folio-go/internal/template/model.go:198`
- **Observation**: AC19 scopes the walk to *"every `{{ }}` occurrence in the document"* and the
  function's own comment repeats it (*"AC19: every `{{ }}` occurrence in the document, not only
  footer-related ones"*). The switch covers `template.ElementText` (`el.Value`) and
  `template.ElementTable` (`col.Bind`). `Element.VisibleIf Presence[string]` — the field Story 3.5 drives
  visibility from, already parsed at `parse_bands.go:188-194` and round-tripped at
  `serialize.go:237-241` — is never visited. A template carrying `"visibleIf": "{{a + b}}"` (a bare
  operator, AC19's own named syntax-error case) loads today without complaint.
- **Impact**: Story 3.5 inherits an expression-bearing field never syntax- or arity-checked, and the harm
  class FLAG-3 names — *"a section silently disappearing from a rendered statement"* — is reachable from
  a malformed `visibleIf` no load-time check rejects. The AC claims coverage the walk does not have.
- **Suggested Resolution**: Extend the walk to `el.VisibleIf` (mechanically small — it is a
  `Presence[string]`, handled exactly like `el.Value`), or narrow AC19's wording and record `visibleIf`
  as Story 3.5's forward obligation per D-000.60 so it does not ship as an open fence.
- **Related AC**: AC19, FLAG-3

### Finding 13: syntax diagnostics lose position, mis-transcribe non-ASCII, and reject newlines

- **Severity**: Minor
- **Category**: Correctness (AD-14)
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/parser.go:65, 110-114, 140, 264, 288, 302`; `/Users/panitw/Projects/folio/folio-go/internal/expr/scan.go:66`
- **Observation**: Four distinct defects in AD-14's located-error surface, all verified:
  - **`parser.go:140` mis-transcribes any non-ASCII character.** `string(c)` is applied to a **byte**,
    not a rune, so the error names a character that is not in the source: `Parse("é")` →
    `unexpected character "Ã" at position 0`. On a **Thai** bank-statement product this is the common
    case, not the exotic one.
  - **`parser.go:288, 302`** — `unbalanced parenthesis: %q is missing its closing ')'` carries **no
    position**, unlike every other message in the file. `Parse("sum(a")` →
    `unbalanced parenthesis: "sum" is missing its closing ')'`.
  - **`parser.go:264`** prints `got %q` on `tokEOF`, whose text is empty: `Parse("a.")` →
    `expected an identifier after '.' at position 2, got ""`.
  - **`parser.go:110-114`** — `skipWS` skips only `' '` and `'\t'`, so a newline inside an argument list
    is a hard syntax error: `Parse("f(a,\n b)")` → `unexpected character "\n" at position 4`. A JSON
    string may legally contain a newline, so a wrapped multi-line expression fails with a message giving
    no hint that the newline is the cause.
  - `scan.go:66`'s unterminated-`{{` error carries neither offset nor offending text, so AC19's
    *"offending text verbatim"* is lost for that case.
  Also noted while reading: `parser.go:301-303` is unreachable (after the block at `278-300`, `c == ')'`
  holds on both paths), which disguises the fact that `:288` is the only handler for EOF inside an
  argument list.
- **Impact**: AD-14 is the reason these messages exist. A diagnostic naming a character the author did
  not type, or omitting the position on the single most common syntax mistake (an unclosed paren), is a
  located error in name only.
- **Suggested Resolution**: Decode a rune rather than a byte at `:140`; add the position to both
  unbalanced-paren messages; special-case `tokEOF` at `:264` to read *"end of expression"*; either accept
  `\n`/`\r` in `skipWS` or name the newline explicitly; carry offset and text in `scan.go:66`. Delete the
  unreachable block.
- **Related AC**: AC19, AC20, AD-14

### Finding 14: the epic's third Given is half-asserted — no test checks that a located error carries the offending expression text

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/folio_expr_validate_test.go:172-174` and `:204-206`; `/Users/panitw/Projects/folio/folio-go/internal/expr/eval_test.go:216-221`
- **Observation**: AC19 requires a syntax error to name **the element id *and* the offending expression
  text, verbatim**; AC15 requires the unimplemented-function error to name the function, the element id,
  **the offending expression text**, and the owning story. Every test asserts a subset:
  `TestParseTemplateRejectsSyntaxErrorInTextValue` and `TestParseTemplateRejectsWrongArityAtLoad` assert
  `strings.Contains(err.Error(), "e1")` and nothing more; `TestUnimplementedFunctionsAreLocatedErrors`
  asserts the element id and the story, not the function name or the expression text. The behaviour is
  present — `checkTextExpressions` interpolates `%q` of the trimmed placeholder
  (`folio_expr_validate.go:64`) and `evalCall` interpolates `call.Raw` (`eval.go:63-66`) — so this is an
  assertion gap, not a defect.
- **Impact**: The half of the epic's third Given that makes a diagnostic useful to a template author —
  being shown the text they wrote — is unpinned. Dropping `call.Raw` or the `%q` operand from either
  message is a silent, green change.
- **Suggested Resolution**: Add the missing `strings.Contains` assertions for the offending text
  (`a + b`, `sum(a, b)`, `sum(x)`) and the function name, at both the load and evaluation layers.
- **Related AC**: AC15, AC19 (epic Given 3)

### Finding 15: six "must error" tests discard the parse error and would pass for the wrong reason

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/eval_test.go:95, 127, 164, 173, 254, 257, 284`
- **Observation**: These sites use `e, _ := Parse(src)` and then assert `Eval(e, …)` errors — including
  `TestIfEmptyStringConditionIsLocatedErrorNoTruthiness` (`:164`) and
  `TestIfZeroConditionIsLocatedErrorNoTruthiness` (`:173`), the two tests carrying the *no truthiness*
  property. If `Parse` regressed and rejected `if(cond, a, b)`, `e` would be `nil`, `Eval` would fall to
  its `default:` arm (`eval.go:32-33`, *"unrecognised expression node"*) and return an error — and each
  test would pass, still reporting that truthiness is correctly refused. Neither inspects the message, so
  nothing discriminates the intended cause from the accidental one.
- **Impact**: The AD-14 no-coercion property — one of the two `if()` semantics this story treats as
  load-bearing — rests on assertions a total parser failure would satisfy. D-000.9's signature failure
  shape, one layer in from where F10 caught it.
- **Suggested Resolution**: Fail the parse explicitly (`e, err := Parse(...); if err != nil { t.Fatalf(...) }`,
  as `mustEval` already does) and assert the eval error names the condition's wrong kind
  (`"must be a boolean, got string"` / `"got number"`), which only the intended arm can produce.
- **Related AC**: AC13, D-000.9, AD-14

### Finding 16: `if()`'s own motivating case is untested, and the not-a-scalar arm has no test at all

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/eval_test.go:183-189`; `/Users/panitw/Projects/folio/folio-go/internal/bind/text.go:323-327`
- **Observation**: Two gaps in one area. (a) `TestIfShortCircuitsUnselectedBranch` exercises the
  short-circuit with an **unimplemented function** in the untaken branch (`if(cond, "a", sum(t.x))`). The
  motivating case both `folio-format.md` and `docs/expression-reference.md` lead with —
  `{{if(hasDiscount, discount.amount, "N/A")}}` on a row where `discount` is **absent entirely** — is not
  exercised anywhere. The mechanism is identical (`evalIf` never passes the unselected node to `Eval`),
  so the reviewer judges the behaviour correct, but the documented example is unpinned. (b) The
  `default:` arm of `lookupBound` — what makes an array or object path a located error — has no test: a
  search for its message `"not a scalar value usable in an expression"` across all `_test.go` files
  returns nothing. Both `folio-format.md`'s new `### Expressions` section and
  `docs/expression-reference.md` explicitly claim an empty array `[]` as an `if()` condition is a located
  error; `expr.Value` has no array kind, so that claim rests entirely on this untested conversion arm.
- **Impact**: The two published statements about `if()` an author is most likely to rely on are the two
  with no test behind them.
- **Suggested Resolution**: Add (a) a short-circuit case with an absent path in the untaken branch, using
  the documented `hasDiscount`/`discount.amount` shape; and (b) a `bind`-level case binding a path to a
  JSON array and asserting the located "not a scalar" error, naming the element id and the path.
- **Related AC**: AC13, AC14

### Finding 17: AC27's fixture was kept in place rather than moved and transformed, while the AC is ticked as written

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/template/footer_test.go:126-149`
- **Observation**: AC27 requires that `TestFooterWithoutFooterOfLoads` **"moves to the root `folio`
  package and becomes the AC21 derivable-arm assertion against `folio.ParseTemplate`"**, and that
  **"the old test's body is transformed"**. Decision 1 took a different route: the test stays in
  `internal/template`, its **body is byte-unchanged**, and only its doc comment was rewritten — to
  reframe the permissiveness as a deliberate layer boundary and name
  `folio.ParseTemplate`/`validateAndDeriveExpressions` and the new root-level tests by symbol. New,
  separate assertions were added at the root instead.
- **Impact**: The reviewer judges the *substance* well served — the closure is plainly visible in the
  diff, the trail is `git log --follow`-able, and the replacement assertions were independently proven
  non-vacuous (M4). But AC27 is ticked `[x]` against text the implementation does not follow, so the
  story's AC list no longer describes what shipped. Note also that FLAG-1's compensating witness is now
  **prose in a doc comment**, not an asserted statement — a weaker form than the D-000.23 witnesses
  elsewhere in this codebase.
- **Suggested Resolution**: Amend AC27's text to record Decision 1's choice, so the AC and the code
  agree. Consider promoting FLAG-1's witness from a comment to an assertion.
- **Related AC**: AC27, FLAG-1, D-000.59, D-000.23

### Finding 18: three of AC24's five red-proofs are reasoning and compile-checks, not observations

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: this file, "AC24 — the red-proofs"
- **Observation**: AC24 requires each red-proof to be recorded **"with the exact command, the failure
  text, and what it would have printed had the mutation not been applied"**. Two are genuinely standing,
  re-runnable tests (AC8's `TestExprFunctionTableRedProofNinthEntry`; AC22's second mutation) and one was
  genuinely performed (AC17 — reproduced byte-for-byte by the reviewer at M1). The other three are not:
  - **AC21 red** — *"Verified by inspection of the pre-change `folio.go` (git history) and by the fact
    these tests could not even compile before …"*. A compile failure is not the assertion failing;
    D-000.30's window was closed with an argument, not an observation.
  - **AC22 red, mutation (1)** — *"reproduced structurally … the pre-move source is preserved in git
    history"*: reasoning, not a run. And mutation (2), the one that *is* a standing test, is the hollow
    one (Finding 4).
  - **AC23 red** — *"`TestParseBindingPathAndIsValidIdentAreAbsent` against that tree **would have
    failed**"*: conditional, not observed; the cited `git show … | grep -n` confirms the pre-move line
    numbers, not the guard's behaviour.
  All three are recoverable cheaply: `10327a4` is one `git worktree add` away.
- **Impact**: The story is candid that it reconstructed these, which is to its credit. But AC24 is ticked
  `[x]`, and the three unrun proofs cover AC21 and AC23 — including AC21, whose one surviving clause has
  now been shown vacuous (Finding 2). An actual run at `10327a4` would likely have surfaced that.
- **Suggested Resolution**: Add a baseline worktree at `10327a4`, run the three assertions against it,
  and record the exact commands and failure text. Do this **after** Finding 2 is fixed, so the P3
  assertion being run is the repaired one.
- **Related AC**: AC24, D-000.30, D-000.9

### Nit 1: a comment cites a function this story deleted

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/bind/value.go:119-121`
- **Observation**: *"path is never empty at a real call site (`parseBindingPath` rejects an empty
  binding)"*. `parseBindingPath` is deleted by this story (AC23); the guarantee now comes from
  `internal/expr`'s parser and `exprResolver.Resolve`'s own `len(path) == 0` check (`text.go:256-258`).
  The comment survived the deletion the extinction guard exists to enforce.
- **Suggested Resolution**: Re-point it at `expr.Parse` / `exprResolver.Resolve`.
- **Related AC**: AC23

### Nit 2: `TestCheckUnknownFunction` derives its expectation from the same source the message is built from

- **Severity**: Nit
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/expr/check_test.go:47-51`
- **Observation**: It asserts the error lists the eight legal names by iterating `LegalFunctionNames()` —
  but `checkCall` builds that message with `strings.Join(LegalFunctionNames(), ", ")` (`check.go:41`). If
  `LegalFunctionNames()` returned three names, message and loop would shrink together and the assertion
  would still pass. The circle is closed only indirectly, by `TestLegalFunctionNamesIsExactlyEight` and
  AC7's AST guard.
- **Suggested Resolution**: Assert against a literal eight-name list here too, as `table_test.go` and
  `expr_arch_test.go` both already do.
- **Related AC**: AC11

### Nit 3: `lookupBound`'s unreachable tail returns a null Value with no error, which `if()` reads as silently false

- **Severity**: Nit
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/bind/text.go:329-330`
- **Observation**: After the `switch presence` covering `Absent`/`Null`/`Present`, the function falls
  through to `return expr.Value{}, nil` — a `KindNull` with a nil error. `Presence` is a closed
  three-value enum, so this is unreachable today. But if a fourth state were added, the failure mode
  would be maximally quiet: a null `Value` reaching `evalIf` takes the else branch **silently** by owner
  ruling, so a new unhandled presence state would hide document sections rather than error.
- **Suggested Resolution**: Return an explicit internal error naming the unhandled presence value.
- **Related AC**: AC13

### Nit 4: AC19's "verbatim, as the author wrote it" reports the trimmed text

- **Severity**: Nit
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/folio_expr_validate.go:61, 64`
- **Observation**: AC19 asks for the offending text *"verbatim, as the author wrote it"*; the message
  interpolates `trimmed := strings.TrimSpace(ph.Inner)`, so authored outer whitespace inside `{{ }}` is
  not reproduced. Harmless, arguably more readable, but not verbatim.
- **Suggested Resolution**: Report `ph.Inner`, or note the deliberate trim in the AC.
- **Related AC**: AC19

### Nit 5: `docs/expression-reference.md` says `avg` "reports a diagnostic", which AC20 forbids until 3.6

- **Severity**: Nit
- **Category**: Convention
- **Location**: `/Users/panitw/Projects/folio/docs/expression-reference.md:132`
- **Observation**: *"`avg` reports a diagnostic instead of dividing by zero"*. AC20 and D-1.4.2 hold that
  `internal/diag` does not exist until Story 3.6 and that this story must not mint a code early;
  `AvgDecimals` returns a plain Go error (`reduce.go:154`). Subsumed by Finding 6, listed separately as a
  distinct factual slip.
- **Suggested Resolution**: Say "an error", not "a diagnostic", until 3.6 lands the codes.
- **Related AC**: AC20

### Nit 6: `docs/expression-reference.md` asserts Story 3.3/3.4 behaviour as settled fact

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `/Users/panitw/Projects/folio/docs/expression-reference.md:126-176`
- **Observation**: The undisclosed reference states unimplemented behaviour in the present indicative —
  *"On an empty collection, `sum` and `count` return zero"*, the four supported locale tags, and
  `formatDate`'s accepted value shapes — for functions this story deliberately leaves unimplemented and
  which Stories 3.3/3.4 own. `count`'s empty-collection result in particular is not settled anywhere the
  reviewer could find; D-3.1a.2 rules only on `SumDecimals`/`AvgDecimals`.
- **Suggested Resolution**: Mark the forward sections as specified-not-yet-implemented, or drop them
  until the owning story lands. Subsumed by Finding 6's disposition.
- **Related AC**: AC15, FLAG-2

---

## Acceptance criteria — verdict on each of the 33

| AC | verdict | note |
|---|---|---|
| AC1 | **satisfied** | hand-written recursive descent; `go.mod`/`go.sum` byte-unchanged in **both** modules; no `regexp` anywhere in `internal/expr`, `internal/bind` or `folio_expr_validate.go`; no generated file. See Finding 3 for the grammar's missing depth bound |
| AC2 | **NOT MET** | **Finding 1 (Blocker)** — the guard does not exist |
| AC3 | satisfied, with Findings 7 and 11 | both tables present, nesting covered and exceeded; number- and string-literal handling qualified above |
| AC4 | satisfied | reserved check runs before any parse in both callers; the top-level-`page`-key fixture is retained and strengthened (`text_test.go:150, 177`). AC4's literal "byte-for-byte unchanged" no longer applies — `text.go` was rewritten — but the property is preserved. See Findings 5 and 10 |
| AC5 | satisfied | one package-level literal, `[8]funcEntry`, FR18's eight names |
| AC6 | satisfied in substance | property holds; guard weak — **Finding 9** |
| AC7 | satisfied | AST set-equality via `walkGoFiles`, independent pinned set, three presence preconditions |
| AC8 | **satisfied, stronger than asked** | M6: a ninth is a **compile error** from `[8]funcEntry`; past that, two independent guards redden |
| AC9 | satisfied | `returnDecimal` embeds a real `Decimal`, so a stringly-typed table would not compile |
| AC10 | satisfied | `check.go`'s `checkCall`/`checkArgKind`, enforced at load via `folio_expr_validate.go` |
| AC11 | satisfied | located at both layers; assertion circular — Nit 2. The `not a panic` clause is breached by Finding 3 |
| AC12 | satisfied | empty, Thai (no-op asserted), CJK, combining mark, non-string → located error |
| AC13 | **satisfied — the required pair is present** | `TestIfAbsentConditionIsLocatedError` and `TestIfNullConditionIsSilentlyFalse` are adjacent, cross-referencing, and assert opposite outcomes. The owner's `if(null)` = silent-false ruling is implemented **without** a diagnostic (`eval.go:148-149`), exactly as ruled — **not reported as a defect**. Weakened by Finding 15 |
| AC14 | satisfied | genuine short-circuit; coverage gap in Finding 16 |
| AC15 | satisfied | located error names function, element id, `call.Raw` and owning story; assertion gap in Finding 14 |
| AC16 | satisfied | five cases, `seen != 5` against a **hardcoded** 5 (not `len(cases)`), so a deleted case reddens |
| AC17 | **hazard genuinely caught** | M1 reproduced the developer's red-proof verbatim; second clause dead — **Finding 8** |
| AC18 | satisfied | distinguishable, and the test asserts the unknown-function error does **not** claim an owning story |
| AC19 | satisfied in the six named cases | Findings 7, 13, 14 and Nit 4 qualify it |
| AC20 | satisfied | plain Go errors, no diagnostic code minted; `internal/diag` still absent (`absence-diag-package` untouched) |
| AC21 | **arms real; final clause NOT MET** | both arms proven non-vacuous by M4 — **Finding 2 (Blocker)** on the P3 clause |
| AC22 | **guard sound; its red-proof hollow** | M5b proves the guard; **Finding 4 (Major)** on the proof. `TestDecimalDeclarationScanSkipsTestdataDecoys` is well built and carries a decoy-presence precondition |
| AC23 | satisfied | both names absent from all module source (remaining occurrences are the guard's own literals and Nit 1's stale comment); no second path matcher exists under any name |
| AC24 | **partially met** | AC17/AC8 real, AC22(2) hollow, AC21/AC22(1)/AC23 unrun — **Finding 18** |
| AC25 | satisfied | `absence-expr-package` deleted in the same commit; `absence-diag-package`/`absence-source-date-epoch` untouched; the removed entry's rationale recorded in place |
| AC26 | satisfied | `TestAbsencesChecksIncludeAllThreeEntries` pins the **exact set** (length + membership by rule id); `ChecksEvaluated` still tracks the scanner's own loop |
| AC27 | **deviation, substance served** | **Finding 17** |
| AC28 | satisfied | moved, never duplicated; `bind` imports back 4→3 |
| AC29 | **satisfied — verified byte-level** | `decimal_test.go`, `reduce_test.go`, `golden_test.go`, `testutil_test.go` diffed against `10327a4`: each differs **only** in the `package` clause. `maxDecimalExponentMagnitude`, `maxDecimalCoefficientDigits` and `avgExtraScale` are all still referenced **by symbol** — no hardcoded `100_000`/`19`/`4`. D-3.1a.1's Story 3.4 forcing function on `avgExtraScale` survives intact, and the handoff note naming it by symbol is correctly recorded |
| AC30 | **satisfied — SHA verified independently** | `git diff 10327a4` empty for that path; SHA-256 `1d5d3a70796703109c4d48fe05dcc64f9df0f5a54bda8d876e2d09c7cf9e74e6` matches both the committed blob and the working file. D-3.1a.3's relational pin followed the move with zero edits, exactly as designed |
| AC31 | satisfied | guard unedited; observed set `[data params row]` across 3 `lookupBound` call sites, logged by the guard itself; presence precondition not duplicated |
| AC32 | satisfied | the two re-pointed sites assert the message **distinguishes** the two causes (`"3.4"` present **and** `"not a valid expression"` absent) — not a bare `err != nil`; the `spans_test.go` site genuinely needed no edit (it asserts `BindText`/`BindTextSpans` agree, not which wording either produces) |
| AC33 | satisfied | gates reproduced exactly; required-red preserved with byte-identical stats; DW-19 trio verified as exactly those three, all for the same cause |

## What the reviewer checked and found clean

- **No new dependency, in either module.** `go.mod`/`go.sum` byte-unchanged from `10327a4`.
  `internal/expr`'s complete non-test import set is `fmt`, `strings`, `math/big`, `internal/template` —
  AD-9 and D-3.2.2 hold as a fact today, even though Finding 1 means nothing mechanically keeps them so.
- **No `float32`, `float64`, `big.Float` or `big.Rat`** anywhere in the new code (the only `float64`
  occurrences under `internal/expr` are in `golden_test.go`'s prose about a historical mutant). 3.1a's
  `bigfloattype` and `floattyped` rules both still pass and both still scan the folio-go module root, so
  `internal/expr` is inside their scope by construction.
- **`parseBindingPath`/`isValidIdent` are genuinely extinct** — not renamed, not relocated. No second
  path matcher exists under any name.
- **DW-5 and DW-8 retired in `deferred-work.md`** with the discharge recorded, and DW-7 correctly left
  untouched.
- **The layering ruling holds**: `internal/template` does not import `internal/expr`; the derivation lives
  in `internal/expr` and is invoked from the unranked module root, as R1 requires. The rank table was not
  amended.
- **`folio-format.md`'s `### Expressions` section is present, accurate, and includes the stated-behaviour
  entry for `if(null)`'s silent false**, with the trade-off spelled out — exactly what the owner ruling
  required. The `if(null)` implementation adds **no** diagnostic, and the paired
  absent-errors/null-silently-false tests exist as mandated. **Not a defect; not reported as one.**

## Recommendation

**Changes Requested.** The engineering substance here is strong — the parser, the closed table, the
load/evaluate split, the `Decimal` move and `if()`'s semantics are all well built, and the two
highest-stakes claims the reviewer attacked directly both survived: the vendor-default `sum → 0` hazard
is genuinely caught (M1 reproduced the developer's red-proof verbatim), and the `Decimal` move left
D-3.1a.3's relational tripwire byte-identical, SHA-256 confirmed. What has not been earned is the
evidence on two obligations, plus one defect nobody's AC anticipated. **Findings 1, 2 and 3 must close
before this story is finished.** Findings 4-7 should. Finding 18's three unrun red-proofs are cheap to
convert into real ones from a `10327a4` worktree, and are worth doing *after* Finding 2 is repaired.

*No production code, test, fixture or configuration file was modified by this review. Every mutation was
applied to a working copy, observed, and reverted from a saved copy (`cp`, never `git checkout`); each
touched file was re-verified byte-identical by SHA-256 and the complete working-tree file set was
re-confirmed against the pre-review snapshot. Adversarial parser probing ran from a throwaway copy of the
module under `/tmp`, since deleted.*

---

## Finding Resolutions (finisher)

**Recount note.** The review's own summary line states 3 Blocker/4 Major/10 Minor/6 Nit = 23. Counted
directly from the `**Severity**` lines in the file: 3 Blocker (Findings 1-3), 4 Major (4-7), **11**
Minor (8-18), 6 Nit — **24**, not 23. The severities and content are otherwise exactly as the reviewer
recorded; only the summary's own Minor count was off by one. Triage below covers all 24.

**Ordering and instruction received:** BLOCKER 2, then BLOCKER 3, then BLOCKER 1, then the four
Majors, per explicit instruction — all seven FIX. Finding 6 (Major) was pre-resolved by the
orchestrator at `5c94bae`, before this pass began, with instruction not to re-touch it. Minors and
Nits triaged on their own merits, per instruction not to dismiss merely to reduce scope; every one
below is FIX except Nits 5 and 6, DISMISSed as out of this story's scope (the same reason Finding 6
is orchestrator-owned, not this story's).

| # | Severity | Decision | Files touched |
|---|---|---|---|
| 1 | Blocker | FIX | `folio-go/internal/expr_arch_test.go` (new: `isStdlibOrFirstPartyImportPath`, `exprImportViolationsFromFile`, `TestExprImportSetIsStdlibOrFirstPartyOnly`, `TestExprImportSetRedProof`) |
| 2 | Blocker | FIX | `folio-go/folio_expr_validate_test.go` (`TestParseTemplateFooterOfDerivationNeverWritesBack` rewritten) |
| 3 | Blocker | FIX | `folio-go/internal/expr/parser.go` (`maxCallDepth`, `parseCall`), `parser_test.go`, `folio_expr_validate_test.go`, `internal/bind/text_test.go` |
| 4 | Major | FIX | `folio-go/internal/expr_arch_test.go` (`decimalDeclLocationsFrom`, `decimalDeclLocations`, `decimalDeclLocationsWithOverride`) |
| 5 | Major | FIX | `folio-go/internal/bind/text.go` (`Resolve`) |
| 6 | Major | **Resolved by orchestrator, outside this story** | `docs/expression-reference.{md,html}`, commit `5c94bae` |
| 7 | Major | FIX | `folio-go/internal/expr/check.go` (`checkNumberLit`), `parser.go` (`lexNumber`), `check_test.go`, `folio_expr_validate_test.go` |
| 8 | Minor | FIX | `folio-go/internal/expr/eval_test.go` |
| 9 | Minor | FIX | `folio-go/internal/expr_arch_test.go` (`exportedFuncMutatesFunctionTable`, `TestExprTableMutationRedProof`) |
| 10 | Minor | FIX | `folio-go/internal/expr/scan.go` (`IsReserved`), new `scan_test.go`, `internal/bind/text.go` (comments) |
| 11 | Minor | FIX (diagnostic half) + DISMISS (scan.go quote-blindness half, documented) | `folio-go/internal/expr/parser.go` (`lexString`), `parser_test.go`, `scan.go` (doc comment) |
| 12 | Minor | FIX | `folio-go/folio_expr_validate.go` (`checkVisibleIfExpression`), `folio_expr_validate_test.go` |
| 13 | Minor | FIX | `folio-go/internal/expr/parser.go` (four sites + unreachable block deleted), `scan.go`, `parser_test.go` |
| 14 | Minor | FIX | `folio-go/folio_expr_validate_test.go`, `internal/expr/eval_test.go` |
| 15 | Minor | FIX | `folio-go/internal/expr/eval_test.go` (7 sites, one more than the "six" named) |
| 16 | Minor | FIX | `folio-go/internal/expr/eval_test.go`, `internal/bind/text_test.go` |
| 17 | Minor | FIX (documentation) | this story file (AC27) |
| 18 | Minor | FIX (documentation correction; literal suggested resolution judged unachievable — see rationale) | this story file (AC24) |
| Nit 1 | Nit | FIX | `folio-go/internal/bind/value.go` |
| Nit 2 | Nit | FIX | `folio-go/internal/expr/check_test.go` |
| Nit 3 | Nit | FIX | `folio-go/internal/bind/text.go` (`lookupBound`) |
| Nit 4 | Nit | FIX (documentation) | this story file (AC19) |
| Nit 5 | Nit | DISMISS | none — out of this story's scope |
| Nit 6 | Nit | DISMISS | none — out of this story's scope |

**1 (Blocker) — AC2's import guard never existed.** FIX. Built `TestExprImportSetIsStdlibOrFirstPartyOnly`
in `internal/expr_arch_test.go`, reusing `walkGoFiles` (R5), with the exact coverage-witness wording AC2
requires, a zero-files `t.Fatal`, and a doc comment citing D-3.2.2's reason (the numeric model, not the
parser). Its red-proof (`TestExprImportSetRedProof`) injects a non-first-party import
(`github.com/google/cel-go/cel`) into an in-memory parse and calls `exprImportViolationsFromFile` — the
exact function the shipped guard calls, never a duplicate, applying Finding 4's own lesson before it
could recur here. Also corrected the Delivery Log's garbled AC1/AC2 sentence, which the finding
identified as the more dangerous half (D-000.24: a claimed guard that does not exist stops anyone else
from building it).

**2 (Blocker) — the P3 write-back guard could not fail.** FIX.
`TestParseTemplateFooterOfDerivationNeverWritesBack` now serializes `tpl.doc` — the actual
`*template.Document` `ParseTemplate`'s derivation walked over (the pointer `validateAndDeriveExpressions`
receives; `TableExt.Columns` is a slice, so a write-back lands in this exact backing array) — and
compares it against an independently-derived canonical form, rather than comparing one document's
serialization against itself and separately inspecting two documents the derivation never touched.
Red-proved live: reinstalled the reviewer's M3 mutation (`validateTableColumns` writing the derived
`footerOf` back via `tbl.Columns[ci].FooterOf = template.Presence[string]{Set: true, ...}`), ran the
test, observed it FAIL exactly as required, then restored `folio_expr_validate.go` from a `cp` backup
— `/usr/bin/diff` confirmed byte-identical — and re-ran green.

**3 (Blocker) — unbounded parser recursion.** FIX. Added `maxCallDepth` (64) to the `parser` struct,
enforced in `parseCall` — the sole recursion point for nested calls — as a declared property of the
parser stated once, not fitted to the reviewer's specific input. Exceeding it is an ordinary located
syntax error naming the position, exactly like every other rejected form. Red-proved against the
reviewer's own reproduction (`strings.Repeat("a(", 800000)`, ~1.6MB) through all three reachable
entry points: `internal/expr.Parse` directly, `folio.ParseTemplate` (load), and `bind.BindText`
(render) — all three now return a located error naming the depth limit, instead of crashing the
process. D-000.13 applied: each new test asserts on the ERROR's specific content, never merely on the
process surviving.

**4 (Major) — AC22's red-proof exercised a hand-duplicated copy.** FIX. Threaded the in-memory-override
capability through the shipped `decimalDeclLocations` itself (now a thin call onto a new
`decimalDeclLocationsFrom`, the ONE extraction body) rather than maintaining a second, hand-copied
extraction for the red-proof alone. `decimalDeclLocationsWithOverride` is now also a thin call onto the
same function. Re-proved with the reviewer's own M5 mutation (capping `locations` to length 1 inside
the shared function): `TestDecimalUniquenessRedProof` now correctly REDDENS (`RED-PROOF FAILED:
injecting a second "type Decimal"... did not produce a second observed location`), where before the fix
it stayed green under the identical mutation — confirming the production guard and its red-proof now
move together, exactly what Finding 4 required. Restored byte-identical afterward.

**5 (Major) — AD-13 breach, a second `{{ }}` tokenizer.** FIX (collapsed, as instructed to try first).
`internal/bind/text.go`'s `Resolve` no longer hand-rolls its own `strings.Index(rest, "{{")` /
`strings.Index(afterOpen, "}}")` scan; it calls `expr.ScanPlaceholders` — the same function
`folio_expr_validate.go`'s load-time walk already called — and uses the `literal`/`placeholders`/
`trailing` segmentation it returns to reassemble output. The now-redundant local
`reservedPlaceholders` alias was removed (nothing else referenced it). `scan.go`'s doc comment's claim
("both internal/bind's render-time resolver… and folio's load-time syntax/arity scan call this") is now
actually true, so it needed no correction. Full `internal/bind`/`internal/expr`/root suites re-run
green after the change; the required-red `TestCorpusMeetsP6ExerciseFloors` unaffected.

**6 (Major) — undisclosed docs artifacts.** Resolved by the orchestrator, not this finisher pass,
before this pass began: the factual contradiction (reserved-alias collision reported at RENDER, not
LOAD) was corrected in `docs/expression-reference.md`, and both files were committed separately at
`5c94bae`, outside this story's tree, per explicit instruction not to re-add them to this story's file
list or redo the fix.

**7 (Major) — R3 breach, oversized literals survive load.** FIX, both halves. `Check` now dispatches on
`*NumberLit` and calls `NewDecimal` on it directly (`checkNumberLit`) — the exact function `eval.go`
already calls, so no rule was re-derived — turning a literal whose bounds are decidable with no data at
all into a load-time error, per R3. Verified against the reviewer's own two reproductions:
`"12345678901234567890123456789"` (29 significant digits) and `"1e999999999999999999999"` both now fail
at `Check`, and through `folio.ParseTemplate` at load. Separately, `lexNumber` now enforces JSON's own
no-leading-zero integer grammar (`01`, `007`, `-01` now rejected at parse), matching what `ast.go`'s own
doc comment already claimed the grammar was. `"0"`, `"0.5"`, `"-0"` remain legal (JSON's own lone-zero
case), confirmed by test.

**8 (Minor) — AC17's second clause was dead code.** FIX. Moved the `{0,0}`-value check inside the
`err == nil` branch (both branches were previously reachable only via `runtime.Goexit`, so the second
could never run). Both assertions are now genuinely reachable: err!=nil (the required outcome) passes
cleanly; err==nil with the exact hazard value gives the specific message; err==nil with anything else
gives the generic one. Re-verified against the reviewer's own M1 mutation (`sum` wired to
`SumDecimals(nil)`): the specific "AC17 RED-PROOF FAILED" message now fires, then restored
byte-identical.

**9 (Minor) — AC6's guard was a name deny-list.** FIX. Added `exportedFuncMutatesFunctionTable`,
walking exported `*ast.FuncDecl` bodies for an `AssignStmt` or `append()` call referencing
`functionTable` (whole value or an indexed element) — the property AC6's own words name directly,
decidable by AST without relying on what the function is called. Red-proved with an injected exported
function named `Extend` (deliberately NOT matching the old name-based deny-list) whose body does
`functionTable[0] = e`: caught by the new structural check, confirming it closes the gap the name-only
check could not.

**10 (Minor) — `ReservedPlaceholders` was exported and mutable.** FIX. Unexported the map; the only way
in or out is now `IsReserved(name string) bool`. Added `TestIsReservedIsExactlyPageAndPages`, a
set-equality pin (`t.Fatal` on empty, exact-membership both directions) — the same discipline AC7
already applies to the function table. No other production code referenced the map directly (only
comments), which were updated to cite `IsReserved`.

**11 (Minor) — string literals can't contain quotes; diagnostics misdirect.** FIX (the diagnostic
half) + DISMISS (the `scan.go` quote-blindness half, documented rather than fixed). `lexString` now
detects `\"` specifically and returns an explicit "string literals do not support escape sequences"
error naming the position, instead of silently stopping at the bogus quote and letting a later,
unrelated stage (trailing content, or "expected ',' or ')'") misreport the cause. Any OTHER backslash
is untouched — verified by a dedicated test that `"a\b"` still parses with the backslash preserved
literally in `Value`. DISMISSed: making `ScanPlaceholders`'s `"}}"` search quote-aware (the
`{{ f("a}}b") }}` case) would mean the placeholder scanner tracking string-literal state itself —
effectively duplicating part of the expression lexer's own job — for a case the grammar's own
no-escape rule already makes rare (a literal containing `}}` without also containing `"`, inside a
`formatDate`/`formatNumber` pattern). Documented as a known limitation in `scan.go`'s own doc comment
rather than left silently unaddressed.

**12 (Minor) — `visibleIf` never validated at load.** FIX — with a correction made before landing,
worth recording: the review's own illustrative example (`"visibleIf": "{{a + b}}"`) assumes `{{ }}`
wrapping, but `folio-format.md`'s field table (*"An expression"*) and this module's own fixture
(`fixtures_test.go:298`, `"visibleIf": "customer.hasTransactions"`, no braces) agree `visibleIf` holds a
BARE expression, unlike a text element's `value`. An initial implementation using
`checkTextExpressions` (which scans for `{{ }}` occurrences) would have been silently VACUOUS — it
would have found none in a bare-expression string and passed no matter what the string said, exactly
the shape this review keeps finding elsewhere. Caught before landing; the shipped fix
(`checkVisibleIfExpression`) calls `expr.Parse`/`expr.Check` directly on the raw value, with no AD-4
reserved-token short-circuit (irrelevant to a syntax with no `{{ }}` at all). Verified both directions:
a malformed `visibleIf` (`"a + b"`) now fails at load naming the element id; the existing bare-path
fixture shape continues to load.

**13 (Minor) — four diagnostic defects plus one dead block.** FIX, all five. Non-ASCII: `nextToken`'s
default case now decodes a rune (`utf8.DecodeRuneInString`) instead of `string(byteValue)` — verified
`Parse("é")` now names "é", not a mis-decoded byte. Unbalanced parenthesis: both messages now carry the
opening and current position. EOF after `.`: reads "end of expression" instead of `got ""`. Newline: named
explicitly ("unexpected newline...") rather than falling into the generic unexpected-character message.
`scan.go`'s unterminated-`{{` error now carries a byte offset (tracked cumulatively across scan
iterations) and the offending text. The unreachable second `peekByte` check at the end of `parseCall`
(dead since the preceding block only ever completes with the next byte already confirmed `)`) was
deleted, with the reasoning recorded in a comment in its place.

**14 (Minor) — offending text/function name unpinned.** FIX. Added the missing `strings.Contains`
assertions: the syntax-error and wrong-arity load tests now check for the literal offending text
(`"a + b"`, `"sum(a, b)"`); the unimplemented-function eval test now checks for both the function name
and the full expression text verbatim, for all five unimplemented functions.

**15 (Minor) — six (actually seven) tests discarded the parse error.** FIX. All seven
`e, _ := Parse(...)` sites in `eval_test.go` now fail explicitly on a parse error, as `mustEval` already
did. The two "no truthiness" tests additionally now assert the eval error names the condition's actual
wrong KIND ("string"/"number"), which only the intended no-truthiness arm can produce — a total-parser-
failure would no longer satisfy either assertion.

**16 (Minor) — if()'s documented example and the not-a-scalar arm were untested.** FIX, both. Added
`TestIfShortCircuitsAbsentPathInUnselectedBranch`, reproducing the exact documented shape
(`{{if(hasDiscount, discount.amount, "N/A")}}` with `discount` absent) rather than only the
unimplemented-function mechanism already covered. Added `TestBindTextArrayPathIsLocatedNotScalarError`
in `internal/bind`, binding a path to a JSON array and asserting the located "not a scalar value usable
in an expression" error — a message with no test anywhere in the module before this fix.

**17 (Minor) — AC27 ticked against text the implementation doesn't follow.** FIX (documentation). AC27
amended to record Decision 1's actual shape (fixture kept in place, doc comment rewritten, separate new
root-level assertions) rather than the originally-planned "move and transform" text.

**18 (Minor) — three of AC24's red-proofs were reasoning, not observations.** FIX (documentation
correction); the review's own suggested resolution (stand up a `10327a4` worktree and run the three
assertions there) was assessed and judged NOT achievable as literally proposed — the assertions in
question reference packages (`internal/expr`, `folio_expr_validate.go`) that do not exist at `10327a4`
at all, since this story's whole obligation lands in one commit by design (D-000.59). Rather than force
an artificial reconstruction for a Minor finding, AC24's text was corrected to accurately label these
three as reasoned demonstrations verified by inspection of preserved git history, not executed
red-proofs — restoring the honesty D-000.9/D-000.30 require without disproportionate effort.

**Nit 1 — stale comment.** FIX. Re-pointed at `internal/expr`'s parser and `exprResolver.Resolve`'s own
guard; noted `parseBindingPath`'s deletion by name.

**Nit 2 — circular test expectation.** FIX. `TestCheckUnknownFunction` now asserts against a literal
eight-name list, the same discipline `table_test.go`/`expr_arch_test.go` already apply, rather than
deriving its own expectation from `LegalFunctionNames()` — the same source the message under test is
built from.

**Nit 3 — silent unreachable tail.** FIX. `lookupBound`'s fallthrough now returns an explicit internal
error naming the unhandled presence value, instead of a silent `KindNull`/nil-error pair that
`evalIf` would have read as silently false.

**Nit 4 — "verbatim" is actually trimmed.** FIX (documentation). AC19 amended to state the deliberate
trim explicitly, rather than changing the (harmless, more readable) shipped behaviour.

**Nit 5, Nit 6 — `docs/expression-reference.md` factual slips.** DISMISS. Both are about the same file
Finding 6 covers, which is explicitly outside this story's tree per instruction (owned by the
orchestrator's `5c94bae`, whose own commit message already discloses this as a "KNOWN GAP"). Not fixed
here; noted as a live follow-up for whoever next revises that page.
