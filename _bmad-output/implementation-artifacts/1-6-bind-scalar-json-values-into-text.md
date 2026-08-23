# Story 1.6: Bind scalar JSON values into text

Status: done

**Story key:** `1-6-bind-scalar-json-values-into-text`
**Epic:** 1 · **Baseline:** `5a1cf7d`
**Covers:** FR14, FR15, FR38, FR43 · AD-14, AD-23
**Adjacent:** AD-1, AD-4, AD-5, AD-12, AD-21, §Consistency Conventions, §Source tree
**Rulings that govern:** **D-1.6.1**, **D-1.6.2**, **D-1.6.3**, **D-1.6.4**, **D-1.6.5**,
**D-1.6.6**, **D-1.6.7**, **D-1.6.8**, D-000.4, D-000.9 *(and its extension)*, D-000.10, D-000.11,
D-000.12, **D-000.13**, D-1.1.b, D-1.1.c, D-1.2.5, D-1.2.6, D-1.3.1, D-1.3.3 *(amended)*, D-1.4.6,
D-1.4.8 *(corrected)*, D-1.4.15, D-1.4.16, D-1.5.5
**Deferred work registered by this story:** **DW-8** (owner: Story 3.2)
**Defect recorded against an earlier story:** Story 1.4 — `## Defect found after done` (D-1.6.6)

> ### Where these rulings live — read before diffing
>
> **D-1.6.1 – D-1.6.8, D-000.13, the D-000.9 extension, and D-1.5.10 (correction accepted) are UNCOMMITTED
> additions in the working-tree copy of
> `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`. They are NOT in `HEAD`, and
> they are NOT in baseline `5a1cf7d`.** The same is true of **DW-8** in
> `_bmad-output/implementation-artifacts/deferred-work.md`. A reviewer who diffs against `5a1cf7d`
> will not find any of them and must read the working-tree files instead.
>
> **Verified at the time of writing** (`git status --porcelain`): `HEAD` is exactly `5a1cf7d`, and
> the working tree carries precisely two modified files — `deferred-work.md` (+21 lines) and
> `folio-mvp-decision-log.md` (+175 lines). No Go file differs from `HEAD`. These rulings will land
> in a commit that is **not this story's**.
>
> *(Story 1.5's creator caught the brief stating this wrongly for the D-1.5.x rulings. Stated
> correctly here on purpose. Verified, not inherited.)*
>
> **Correction (QA Nit 17, this story's review; resolved by the finisher).** The line counts above
> were wrong even at the time they were written. Measured against baseline `5a1cf7d`
> (`git diff --stat 5a1cf7d`): **three** doc files were modified before implementation began, not
> two — `deferred-work.md` (+21), `folio-mvp-decision-log.md` (**+323**, not +175), and
> `1-4-load-validate-and-round-trip-a-folio-template.md` (+56, this story's own D-1.6.6 append,
> made during the story's own creation and documented two paragraphs below). Creator-side, not
> developer-side — flagged because this banner is the reviewer's own instruction sheet for what to
> expect outside `HEAD`.

> ### Mechanism tagging (D-1.4.15's convention)
>
> Every named mechanism below carries `(mechanism: binding)` or `(mechanism: illustrative)`.
> **An illustrative mechanism may be improved on; a binding one may not be departed from without
> surfacing it as `DECISION NEEDED`.** Where this story quotes a ruling, the tag travels with the
> quote.
>
> **Nothing is parked. Both items this story surfaced are ruled** — as **D-1.6.6** (the exponent
> defect, which measurement escalated from a latent primitive bug to a **shipped denial-of-service
> reachable through the public loader**) and **D-1.6.7** (AC5's module-root `float64` gap, ruled as
> option (c): a second production **caller** of the existing checker — *a caller is not a checker*).
> M-4 is endorsed as **D-1.6.8**. The creator's control-mutation finding is promoted to a
> program-wide standing rule as **D-000.13**. See
> [§ Decisions resolved](#decisions-resolved).
>
> **The exponent defect is recorded against Story 1.4**, in an appended
> `## Defect found after done` section of that story's file (D-1.6.6, binding: appended, never
> rewritten). It is **not** a `deferred-work.md` entry — it is being fixed now, not deferred.

---

## In plain terms (read this first if you just want the gist)

Until now the engine printed whatever text a document literally said. This story makes the document
a template: the caller hands over a report as JSON, and a placeholder in text is replaced by the
matching value, working end to end.

Money handling was mostly right but had two real gaps review found and finishing closed. The
exact-number conversion this story exists to build had no path connecting it to anything the program
runs — a bound number failed only for being the wrong kind for text, without the conversion or its
safety limits ever touching it. It is wired in now, so an oversized figure is refused by name.
Separately, one safety limit checked the wrong quantity and could let a huge number through under
two spellings; both are fixed, shown failing before and passing after.

The three ways a placeholder can resolve — missing, blank, wrong kind — genuinely disagree,
reconfirmed three separate ways. One existing proof described more than it demonstrated; its record
now says precisely what it proved.

Both guardrails hold: unsupported syntax is rejected naming the later project stage that supports it,
and the page-number placeholders stay unhijackable by a same-named field — tightened, since the
original proof could pass for the wrong reason.

An unrelated crack was also closed: a broken template needing an unresolvable font was supposed to
fail regardless of data, but quietly passed whenever the bound text came out empty. It now fails
consistently.

This story's denial-of-service fix (a huge exponent could hang the engine forever) was real and
necessary, but its permanent proof covered only one of two mechanisms; a second proof now covers
both, and an over-wide safety margin was tightened. Report data with stray content appended after the
real JSON used to silently render just the first part; that is now rejected.

Nothing here changes what was always out of scope: `formatNumber(...)`-style expressions remain
Epic 3's, real page-number slots remain Story 2.7's, and the byte-comparison run stays deferred to
this epic's end.

---

## Story

As an integrating Go developer,
I want to supply my data as JSON and have `{{customer.name}}` resolve to the right value,
So that the template is a template rather than a fixed document.

---

## Acceptance Criteria

Source ACs are `_bmad-output/planning-artifacts/epics.md` §Story 1.6 (lines 578–608). They are
carried below unchanged in substance and split so each is separately red-proofable, with the
ruling-mandated additions folded in. Every AC below is stated so that the D-000.9 diagnostic
question — ***"what would this check have printed if it had been unable to run at all?"*** — has an
answer other than "the same thing".

### The exact decimal (AD-23, D-1.6.1, FR14)

**AD-23's Rule fixes the shape verbatim** (verified, `ARCHITECTURE-SPINE.md:418`):

> *"the JSON decoder preserves number **literals** (`UseNumber` or equivalent) and `internal/bind`
> converts each to an exact scaled-integer decimal — an **`int64` coefficient plus an exponent** —
> carrying the literal's own precision. … `float64` appears nowhere under `internal/`, for geometry
> or for data. A literal too large for the representation is an `Error` (AD-14), never a silent
> narrowing."*

**AC1.** The report-data JSON decoder preserves number **literals** (`json.Decoder.UseNumber` or an
equivalent that does not go through `float64`), and each literal is converted to a value carrying an
**`int64` coefficient and an integral decimal exponent** *(mechanism: binding — the `int64`
coefficient and the integral decimal exponent; **illustrative** — the field names, the exponent's
own width)*. **This is not `geom.Length`**: millipoints are a fixed three-decimal scale, and a
bank-statement literal carries its own (D-1.6.1, binding).

**AC2.** Precision is carried from the literal, and distinct literals stay distinct. At minimum,
asserted as a table:

| Literal | Coefficient | Exponent |
|---|---|---|
| `1.50` | `150` | `-2` |
| `1.5` | `15` | `-1` |
| `36` | `36` | `0` |
| `1e3` | `1` | `3` |

*(mechanism: binding — these four pairs; the spelling of the struct is illustrative)*

**AC3 — coefficient bound.** A literal whose coefficient does not fit `int64` is a **located error**
naming the data path and the element id — never a truncation, never a wrap. **Only significant
digits count** *(mechanism: binding)*: `0.00000000000000000001` is legal and yields `{1, -20}`,
while a twenty-significant-digit literal such as `12345678901234567890` is an error. This is the
source AC's *"never a silent narrowing"*.

> **Measured, and it changes how AC3 must be implemented — see M-3. Confirmed by the lead, with one
> edge case added.** The shipped splitter returns `intPart="0"`,
> `fracPart="00000000000000000001"` for `0.00000000000000000001` — a **21-character** raw digit
> string for a **one**-significant-digit value. A naive `len(intPart+fracPart) > 19` check therefore
> **rejects a legal literal**. Significance here means: **strip leading zeros from the concatenated
> digit string; never strip trailing zeros** (stripping trailing zeros would collapse `1.50` into
> `{15,-1}` and violate AC2).

**AC3a — the all-zeros edge case** *(mechanism: binding, D-1.6.6)*. **`0.00` strips to the EMPTY
STRING.** It must yield **`{0, -2}`** — **not** an error, and **not** a lost scale. A naive
implementation does one of those two things. Measured (M-3): `splitJSONNumber("0.00")` →
`intPart="0"`, `fracPart="00"`, concatenated digits `"000"`; strip leading zeros and nothing remains.
Test `0.00`, `0.0`, `0` and `-0.00` explicitly.

**AC4 — exponent bound, stated openly as a safety addition beyond the source AC's letter.** An
exponent outside a bounded range is a **located error** naming the data path and the element id.
*(mechanism: **binding** that a bound exists; **illustrative** what the specific range is.)*

The reason is stated here rather than buried, per D-1.6.1: report data is **untrusted caller
input**, and a ten-byte literal such as `1e999999999` is perfectly representable in the struct but
renders to a gigabyte of digits. Rejecting at parse, with a located error, beats accepting a value
that fails later and further from its cause.

> ⚠️ **This is not a latent primitive bug. It is a SHIPPED DENIAL-OF-SERVICE, reachable through the
> public loader from untrusted input — see M-1, and D-1.6.6.** The exponent accumulator wraps
> silently, the wrapped value reaches `big.Int.Exp`, and the loader **hangs indefinitely**.
> Measured end-to-end: **`folio.LoadTemplate` on a syntactically valid `.folio` hung and was killed
> by `-timeout` at 20s.** Story 1.4 promised *"malformed … templates rejected with a **located
> error**"*. Epic 5's designer loads user-authored templates. **The fix lands in this story**, and
> the defect is recorded against Story 1.4 (D-1.6.6).

**AC4a — two fixes in `parseDecimalExponent`, both needed** *(mechanism: binding, D-1.6.6)*:

1. **An overflow check DURING accumulation.** Checking *after* the loop reads an
   **already-wrapped** value and is not a check at all.
2. **A documented magnitude bound, rejected BEFORE any `big.Int` scaling is attempted.** Rejecting
   after the scaling has begun does not prevent the hang — the hang *is* the scaling.

**AC4b — layering: the splitter rejects the absurd, each consumer applies its own tighter check**
*(mechanism: binding, D-1.6.6)*. `decodePoints` keeps its millipoint-range check; `internal/bind`
applies `Decimal`'s. **The shared splitter's bound must be the WIDER of the two, so it never refuses
what a consumer could legally represent.** The specific number is *illustrative*; **tests at the
bound and one past it are binding.**

**AC4c — a retained CORPUS fixture, not only a unit test** *(mechanism: binding, D-1.6.6)*. A
`.folio` file containing `1e99999999999999999999` in a geometry field must produce a **located
error, quickly**, through the **public loader**. The unit test proves the primitive; **the fixture
proves Story 1.4's promise now holds through the public path.** One without the other is not the
obligation.

> **Red-proof wrinkle — do not let this be fudged** *(mechanism: illustrative, D-1.6.6)*. **Before
> the fix, this test HANGS.** A hang is not a clean assertion failure: Go's `-timeout` panics the
> process. **That timeout IS the red**, and it is recorded verbatim as *"hung; killed by `-timeout`
> at Ns"* — **never** dressed up as an assertion failure. Per D-1.2.5, a proof states what it
> actually observed. The measured baseline reds are in M-1; reuse those exact words.

> **Why AC4 is new surface rather than a restatement (M-2).** Story 1.4's own path is *not* exposed
> to the amplification hazard for *representable* exponents — `geom.Length`'s int64 millipoint
> bound catches `1e2000000` in 0.07s. The hang appears only once the **exponent itself** overflows
> `int`, which no 1.4 fixture reached.

**AC5 — no `float64` in the binding path.** Source AC, verbatim: *"no `float64` is produced anywhere
in the binding path."* Under `internal/` this is already `TestNoFloat64UnderInternal`'s coverage and
`internal/bind` inherits it on creation.

**AC5a — a second production CALLER of the existing checker** *(mechanism: **binding** for the
properties; **illustrative** for the traversal — **D-1.6.7**)*. The module root is outside every
existing guard (measured, M-6). The ruled resolution:

> **A caller is not a checker.** D-1.6.3 forbade a second *checker*; D-1.3.3 built exactly this
> architecture — **one pure checker over a target path, with production and fixture callers** — and
> adding a production call site is the shape that design exists to serve.

Four binding properties:

- **Scope: the whole module** — `folio-go/`, root and below. This **subsumes** `internal/` and
  automatically covers `fonts/`, `cmd/`, `wasm/` **on arrival**, rather than needing a new call site
  each time one appears.
- **Include `_test.go`.** D-1.3.1's precedent: exemptions exist only where *required*, never
  pre-emptively. No `float64` is required in a test here.
- **Skip `testdata/` and any dot-directory BY CATEGORY** — never by naming `.matrix-build` or any
  other specific directory. **A name on a list is the rot pattern.**
- **Keep the existing `internal/`-scoped caller.** Its vacuity guard names `geom` and `pdf`
  specifically and still earns its place. **Two callers, one checker.**

> **Verified: `grep "float64\|float32" folio-go/*.go` returns ZERO matches, `_test.go` files
> included.** So the new caller goes **green on its first run**, and its red-proof is exactly the
> mutation M-6 already performed — reapply it, watch the new caller go red **naming the rule id**,
> revert.
>
> **Options (a) and (b) were rejected, and the reasons are load-bearing.** **(a)** — AD-2: *"no
> `float64` appears in any layout, measurement, or emission **signature**"*, and the public API
> **is** a signature; AD-23's Prevents names FMA contraction, which a `float64` in `folio.Render`
> reintroduces exactly. **(b)** — it conflates two invariants and scopes `float64` to **one file**,
> leaving `fontset.go` and `version.go` invisible.

**AC6 — one splitter in the module** *(mechanism: binding)*. `internal/template/decimal.go:76`
already ships the primitive:

```go
func splitJSONNumber(literal string) (sign int, intPart, fracPart string, exp int, err error)
```

**Reuse it; do not write a second.** Two splitters would drift, and this is the money path. Whether
it is exported from `internal/template` or moved to a lower package is *illustrative*;
`internal/bind` importing `internal/template` is the correct direction (Document → BoundTree).
**What binds is that exactly one exists in the module.**

**AC7 — two decode trees, one shared primitive** *(mechanism: binding, D-1.6.4)*. The template
decoder (Story 1.4) builds a **schema-typed** tree; the report-data decoder (this story) builds a
**generic** value tree. The two tree-builders stay separate — merging them would force the `.folio`
schema to model arbitrary caller JSON. What they **share** is the number-literal decomposition of
AC6. One splitter, two consumers, separate trees.

### AD-14's three cases, and the proof that two of them disagree (D-1.6.2, FR43)

AD-14's Rule, verbatim (`ARCHITECTURE-SPINE.md` §AD-14):

> *"Three data cases that would otherwise each be decided twice: an **absent** path is an `Error`
> carrying the path; an explicit JSON **`null`** renders as empty and is not an error; a value of
> the **wrong kind** for its element is an `Error`, never a coercion."*

**AC8.** A binding whose path is **absent** from the data is an `Error` that names **both the data
path and the element id**. The render fails.

**AC9.** A binding whose value is explicit JSON **`null`** renders as **empty**, and is **not** an
error. The render succeeds.

**AC10.** A binding whose value is of the **wrong kind** for its element (a JSON number into a text
element) is an `Error`. The value is **never coerced**.

**AC11 — the retained fixture triple, and its vacuity guard** *(mechanism: binding)*. AC8, AC9 and
AC10 are proved by a **triple over ONE template and ONE path**, differing **only** in the data:

| Data | Outcome |
|---|---|
| `{}` | **Error**, naming path + element id |
| `{"customer": {"name": null}}` | **renders empty**, no error |
| `{"customer": {"name": 123}}` into a text element | **Error**, no coercion |

**Rows 1 and 2 are the proof.** Inputs differing **only** by absent-versus-null produce **opposite**
outcomes, so the test **cannot pass if the three-state distinction collapses**. If a future refactor
makes `Presence` two-state, row 1 or row 2 goes red immediately.

**Vacuity guard** *(mechanism: binding)*: all three rows assert against the **same** template and the
**same** path. **Three separate templates would let each row pass for its own unrelated reason** and
is forbidden.

> This is the story that makes D-1.4.8's presence seam load-bearing rather than decorative. D-1.4.16
> named the exposure precisely: *"at Story 1.4 nothing branches on `Null` vs absent … a three-state
> API exists whose third state no production code reads."* Story 1.6 is where production code reads
> it. `Presence[T]{Set, Null, Value}` (`internal/template/presence.go`) is the seam — verified
> present and three-state at baseline.

### The world-reading fence (D-1.6.3, AD-1, AD-12, FR38)

**AC12 — structural separation, not reachability** *(mechanism: **binding** for the property;
**illustrative** for the filename)*. D-1.6.3, verbatim:

> *"The file declaring `Render` and `RenderTo` imports **none** of `os`, `time`, `net`,
> `math/rand`. The world-reading shell entry points (`LoadTemplate`) live in a **different file**,
> which may."*

This is the same shape D-1.1.b used for `numbers.go`. It is decidable, exact, has no false
positives, and fails the moment someone adds a convenience `os.Getenv` to the render entry point.

> **Measured — the gap is real and the split is required work, not a no-op. See M-5 and M-6.** At
> baseline, `Render` and `LoadTemplate` are declared in the **same file** (`folio-go/folio.go`),
> which imports `os`. So AC12 requires an actual file split; it cannot be satisfied by inspection.
> `folio-go/render.go` (holding `renderDocument` and its helpers) already imports none of the four.

**AC13 — 1.6 adds no second checker** *(mechanism: binding, D-1.6.3)*. *"`internal/`'s coverage is
already 1.3's, and a second checker is what the lead's horizon note correctly forbade."* AC12's
property is asserted from the existing rules module, over the module-root package, reusing
`ScanForbiddenImports` — not by a newly invented scanner.

**AC14 — the "locale" clarification, and it must appear in the shipped code, not only here**
*(mechanism: illustrative, D-1.6.3)*. The source AC's *"read no … locale"* means the **host**
locale. **AD-12 *requires* the render to use the locale declared in the template**:

> *"the document declares one `locale` tag and one fixed UTC offset … an unlisted tag is a load
> error, not a fallback."* (AD-12)

**Reading the document's locale is mandatory; discovering the machine's is banned.** Without this
sentence someone reads the source AC as forbidding AD-12. It belongs in the doc comment of whichever
file AC12 creates.

### The grammar, and the reservation (D-1.6.5, AD-4, FR15)

**AC15 — the accepted grammar** *(mechanism: binding)*:

```
binding := "{{" ws? ident ( "." ident )* ws? "}}"
ident   := [A-Za-z_][A-Za-z0-9_]*
```

**AC16 — everything else is a located error naming the element id** *(mechanism: binding)*.
Specifically rejected: `(`, `)`, `[`, `]`, `,`, quotes, operators, and interior whitespace inside the
path. **Rejecting is the mechanism that stops 1.6 becoming a second expression implementation** — a
permissive 1.6 that tries to "handle" a function call leaves Story 3.2 with two parsers to reconcile,
and the wrong one wins whichever way that goes.

**AC17 — the error message names Epic 3** *(mechanism: illustrative)*, so a template author writing
valid-in-3.2 syntax is told *"not yet supported"* rather than *"malformed"* — otherwise they rewrite
a correct template.

**AC18 — `{{page}}` and `{{pages}}` are reserved and left untouched** *(mechanism: binding)*.
AD-4 is absolute:

> *"**No expression may reference pagination** — there is no `page` namespace and none may be added;
> page numbers exist only as slots."* (AD-4)

They are late-bound **slots** owned by **Story 2.7**. At 1.6 they are **never resolved from data and
never an error** — the text passes through unchanged.

**The failure this prevents:** if 1.6 resolved single-segment paths from data with no reservation, a
data document containing a `page` key would **silently capture the page-number slot** — and the
symptom is *a wrong number on a rendered page*, not an error.

**AC19 — retained fixture for the reservation** *(mechanism: binding)*: data containing a
**top-level `page` key** must **not** change what `{{page}}` renders. Red-proof it by deleting the
reservation branch and watching this fixture go red.

> **Verified: `folio-go/testdata/template/golden/worked-example.json:56` already contains
> `"Page {{page}} of {{pages}}"`** — line number confirmed by reading the file, not inherited. So
> 1.6 meets these on its **first corpus run**.

**AC20 — the scope fence AC16 must not overrun, and the corpus already contains the trap.**
See **M-4**. The canonical golden fixture contains **four** `{{…}}` occurrences, in **two different
schema fields**:

| Site | Field | Content | 1.6's business? |
|---|---|---|---|
| `worked-example.json:12` | `columns[].bind` | `{{transaction.date}}` | **No** — table binding |
| `worked-example.json:19` | `columns[].bind` | `{{formatNumber(transaction.amount, "#,##0.00")}}` | **No** — table binding, and **expression-shaped** |
| `worked-example.json:56` | text element `value` | `Page {{page}} of {{pages}}` | Yes — AC18/AC19 |
| `worked-example.json:74` | text element `value` | `Statement for {{customer.name}}` | Yes — AC15 |

*(mechanism: binding — **D-1.6.8**, which endorses this AC as ruled.)*

> **Keep the corpus in the DEFAULT gate, never behind a build tag** *(D-1.6.8)*. This is the
> **second** time the canonical fixture has caught a scoping error before it shipped — the first
> being `{{page}}`/`{{pages}}` (AC18). A corpus that has caught two scoping errors in one story is
> earning its place in the gate that always runs.

**AC16 applies to text-element `value` interpolation only.** `table.bind` and `columns[].bind` are
`string` fields on `TableExt`/`Column` (`internal/template/model.go:179,192`) belonging to the table
stories and Epic 3; `collectTextRuns` skips every non-text element today. **An implementation that
scans all `{{…}}` in a document rather than only the text-element values it actually binds will
reject the canonical worked example on its first corpus run**, on the `formatNumber(...)` at line 19.
That would be a self-inflicted failure of AC16, not a finding.

**AC21 — pre-commitment for Story 3.2, registered as DW-8** *(mechanism: binding)*. Two obligations,
one owner:

1. **`internal/expr` may never import `internal/bind`.** The arrow runs `bind → expr`. AD-23
   **Binds** both packages, and Epic 3's `sum`/`avg`/comparison need exact decimal arithmetic — so
   `expr → bind` would be an **import cycle**. **Go stops the cycle; what Go does not stop is
   someone duplicating the type to break it.** When 3.2 needs `Decimal`, the type **moves**;
   duplicating it is forbidden.
2. **Story 3.2's parser replaces 1.6's path matcher — deleted, not kept alongside.**

DW-8 is already written into the working-tree `deferred-work.md` (verified, lines 144–163), owner
**Story 3.2**, with DW-5's existing `internal/expr`-absent tripwire as the forcing function. **This
story's job is to make sure the code it ships matches what DW-8 records**, not to re-register it.

### The signature (D-1.6.4, D-1.5.5, D-1.1.c, FR38)

**AC22** *(mechanism: binding)*:

```go
type Data []byte // JSON report data

func Render(t *Template, d Data, f FontSet) ([]byte, error)
```

`d` is **inserted into its final target position**, never appended — target is `Render(t, d, p, f)`
(D-1.5.5: *"No call site is ever reordered, only extended"*). Convergence: 1.5 `Render(t, f)` → **1.6
`Render(t, d, f)`** → 1.7 `Render(t, d, p, f)`.

**AC23 — `Data` is a named defined type, not `[]byte`, and not an alias** *(mechanism: binding)*.
D-1.1.c's reason is unchanged and is that ruling's strongest argument: `Data` and `Params` become
**adjacent same-typed arguments at Story 1.7**, and *in a product whose acceptance fixture is a bank
statement, that swap must be a compile error, not a support ticket.* **Do not "simplify" to
`[]byte`.**

**AC24 — bytes, not a decoded value** *(mechanism: binding)*. AD-23 requires the library to own the
`UseNumber`-preserving decode, so a caller-decoded value would arrive **with its literals already
destroyed**. Accepting `any` or `map[string]any` is forbidden.

> **Measured — M-7.** Thirteen `Render(` call sites exist across four test files
> (`folio-go/template_test.go`, `fixture_test.go`, `render_test.go`); all take the two-argument
> form today and all need `d` inserted in the middle. None is a production call site.

### Standing rules

**AC25 — D-000.9 and its extension.** Every check this story ships returns findings **and a coverage
witness**; **zero candidates is a failure, not a pass**. Ask of every check: ***"What would this
check have printed if it had been unable to run at all?"*** **And of every red-proof, one level up:
*"What would this red-proof have printed if the mutation had never been applied?"*** Story 1.5's
review found **four** artifacts named as proofs that could not fail, one asserting
`rejectedTag(8) != rejectedTag(8)` without ever calling the function under test. Red is produced
**by construction, in a gate that actually runs** — never a mutation applied and reverted, never
behind a build tag, never "verified by inspection".

**AC25a — D-000.13, promoted to a program-wide standing rule from this story's own M-6a**
*(mechanism: binding)*:

> **A red-proof asserts on the rule id and the finding message, never on exit status or mere
> failure. And a control mutation must be VALID SYNTAX with FORBIDDEN SEMANTICS** — the same rule
> `D-1.3.3 (amended)` set for violating fixtures, now applied to controls.

**Why this is sharper than it looks, and why D-000.9 does not catch it.** The creator's first
control appended an `import` after code, producing a **parse error**. Both scanners went red — but
through `D-1.3.3 (amended)`'s **error path**, without ever finding the violation. **It is not a
proof that cannot fail; it is one that fails without exercising the property at all**, so D-000.9's
question ("what would this have printed if it could not run?") **passes it** — it *did* run, it
*did* fail. **The new question is: *did it fail for the reason it names?*** Both questions go in
this story's vacuity-guard list.

> **The irony, preserved deliberately:** `D-1.3.3 (amended)`'s error path — added so that a crashed
> scan could never read as clean — is the very mechanism that made this false red possible.

**Every vacuity guard in this story therefore answers three questions:**

1. *What would this check have printed if it had been unable to run at all?* (D-000.9)
2. *What would this red-proof have printed if the mutation had never been applied?* (D-000.9 extended)
3. ***Did it fail for the reason it names?*** (D-000.13)

**AC26 — D-000.11.** Every gate runs with `-count=1`. **A cached `ok` is not evidence.**

**AC27 — D-000.12.** **Never verify bytes or hashes through a shell pipe.** Use `rtk proxy` writing
to a file. The wrapper returns a wrong hash through pipes, which reads **exactly** like "the golden
fixture changed".

**AC28 — D-000.4, stated explicitly so nobody reads silence as green.** Heavy-test cadence is
**per-epic**. **Story 1.6 is NOT one of D-000.4's override stories** (those are 1.2, 1.5, 1.8, 2.4,
4.7 — a hash-shaped deliverable). **The cross-target hash matrix is therefore DEFERRED to the Epic 1
boundary and is NOT run in this story.** Unit tests, `go vet`/lint and the build run on this story
regardless, and this story does not reach `done` on a red unit suite. **Do not imply the matrix is
green; it is not run.**

**AC29 — D-1.3.1.** The new `internal/bind` package inherits the full AD-1 ban in non-test files
(`time`, `os`, `math/rand`, `net`, `math` transcendentals) and D-1.3.1's exact `_test.go` exemption
(`os`, `testing`, `path/filepath`, `embed` — and **nothing else**; `time`, `math/rand` and `net`
stay banned in tests too).

**AC30 — D-000.10.** The developer dispositions **every ruling ID this story enumerates**, one line
each, in the table at [§ Ruling disposition table](#ruling-disposition-table-d-00010); the reviewer
mirrors it independently. The enumeration is explicit and complete in that table.

---

## Decisions surfaced by this story

**Nothing is parked.** Both items this story surfaced by measurement were ruled while the file was
being written, as **D-1.6.6** and **D-1.6.7**; M-4 was endorsed as **D-1.6.8** and M-6a promoted to
**D-000.13**. All four are carried in [§ Decisions resolved](#decisions-resolved). This section
records **how each was surfaced**, because in both cases the escalation — not the conclusion — is
the reusable part.

### Surfaced item 1 — a latent primitive bug that measurement escalated into a shipped DoS

Probing `splitJSONNumber` under D-1.6.1's reuse mandate found a wrapping exponent accumulator (M-1).
That alone would have been a tidy note against a shared primitive. **Following it into the public
path changed its severity twice:** `decodePoints` does not merely mis-scale, it **hangs**; and
`folio.LoadTemplate` — the public API, which Epic 5's designer points at user-authored templates —
**hangs on a syntactically valid file.** Story 1.4's user story had promised a located error.

**Ruled as D-1.6.6.** The reusable lesson: *a bound tested only at values the downstream type can
still reject does not test the bound.* Story 1.4's numeric fixtures stopped at `1e2000000`, which
`geom.Length` rejects correctly in 0.07s; the defect begins one class of input further out.

### Surfaced item 2 — a three-way scope collision that turned out to be a wording problem

AC5's *"binding path"*, AD-23's *"under `internal/`"* and D-1.6.3's *"1.6 adds nothing else"* could
not all hold once M-6 measured a `float64` living green inside `folio.Render`. Raised as
`DECISION NEEDED` under D-1.2.6 rather than arbitrated.

**Ruled as D-1.6.7, and the ruling reframed the question:** the collision was **with the lead's
phrasing, not between the documents** — D-1.6.3 forbade a second **checker**, and **a caller is not
a checker**. The reusable lesson: when three documents appear to collide, check whether one of them
is a *paraphrase* before treating it as a constraint. (This is D-1.1.b's lesson arriving from the
opposite direction: there, an AC silently widened a ruling; here, a ruling's wording silently
narrowed an architecture.)

---

## Decisions resolved

All are **orchestrator decisions** recorded in the **working-tree** decision log (see the banner at
the top of this file for why they are not in `HEAD`). **D-1.6.6, D-1.6.7, D-1.6.8 and D-000.13 were
ruled while this file was being written**, in response to measurements taken during its creation.

### D-1.6.1 — `Decimal` is an `int64` coefficient plus a decimal exponent, in `internal/bind`, over one shared literal splitter

Governs **AC1–AC7, AC21**. The shape is fixed by **the spine, not by the lead** — AD-23's Rule at
`ARCHITECTURE-SPINE.md:418`, verified by reading the file. Two bounds (coefficient by significant
digits, and an exponent bound stated openly as a safety addition), both located errors naming path
and element id. **One splitter in the module.** `Decimal` lives in `internal/bind` at 1.6, with the
`expr → bind` cycle pre-committed away as DW-8.

### D-1.6.2 — AD-14's three cases land unchanged, and the proof is that two of them *disagree*

Governs **AC8–AC11**. The retained fixture is a triple over **one** template and **one** path. Rows 1
and 2 are the proof. The vacuity guard (same template, same path) is binding.

### D-1.6.3 — Reuse Story 1.3's lint; close the shell gap with a file-scoped rule, not reachability

Governs **AC12–AC14**. *"So `folio.Render` could call `os.Getenv` today and no guard would fire"* —
**independently re-measured and confirmed (M-6)**, and the file split is confirmed as real work
(M-5). The AD-12 locale clarification is illustrative but mandatory in the shipped doc comment.

### D-1.6.4 — `Render(t *Template, d Data, f FontSet)`; `Data` is bytes; two decode trees, one shared primitive

Governs **AC7, AC22–AC24**. Signature insert-in-final-position per D-1.5.5. `Data` is a named type
for the bank-statement-swap reason. Bytes because AD-23 makes the library own the `UseNumber` decode.

### D-1.6.5 — Accept a bare dotted path; reject expression syntax loudly; reserve `{{page}}`/`{{pages}}`

Governs **AC15–AC21**. The grammar, the loud rejection as the mechanism that stops a second
expression implementation, the Epic 3 error message, and the `page`/`pages` reservation with its
retained fixture. **AC20 is this story's own addition to the ruling's implementation surface**, from
M-4 — it does not depart from D-1.6.5, it identifies the field scope within which D-1.6.5's
rejection operates so the ruling does not accidentally break the canonical corpus.

### D-1.6.6 — the exponent defect is a shipped DoS; the fix lands in 1.6 and the defect is recorded against 1.4

Governs **AC4, AC4a, AC4b, AC4c, AC3a**. **Not a reopen of Story 1.4**, and the reason is stronger
than audit hygiene: *D-1.6.1 already mandates both the exponent bound and reuse of this exact
splitter, so 1.4's defect and 1.6's ruled requirement are **the same line of code**. Reopening would
land a bound 1.6 immediately supersedes — in the one primitive we have forbidden from existing
twice.*

**"Recorded against 1.4" is concrete** *(binding)*: an appended **`## Defect found after done`**
section in Story 1.4's file — **appended, never rewritten** — stating the measurement, naming the
fixing commit, and saying plainly that **1.4's user-story promise was not met by the shipped code**.
**Not** a `deferred-work.md` entry: this is being fixed now, not deferred, and filing it there
dilutes that file.

> ✅ **Done during this story's creation.** `1-4-load-validate-and-round-trip-a-folio-template.md`
> grew from 1743 to 1799 lines by pure append; nothing above the new heading was touched. **The
> developer must fill in the fixing commit hash there** once Story 1.6 is committed — it is the one
> field that cannot be written yet.

### D-1.6.7 — AC5 is option (c): a second production CALLER of the existing checker

Governs **AC5, AC5a**. *"A caller is not a checker."* D-1.3.3's architecture — one pure checker over
a target path, with production and fixture callers — is exactly the shape that serves this. Whole
module, `_test.go` included, `testdata/` and dot-directories skipped **by category**, existing
`internal/` caller kept. **Two callers, one checker.** Options (a) and (b) rejected on AD-2's
"signature" clause, AD-23's FMA Prevents, and (b)'s single-file blindness.

### D-1.6.8 — M-4 is endorsed as AC20, and the corpus stays in the default gate

Governs **AC20**. The field-scope fence is ruled, with the added framing that this is the **second**
scoping error the canonical fixture has caught before it shipped — which is the argument for keeping
the corpus in the gate that always runs rather than behind a build tag.

### D-000.13 — a red-proof asserts on rule id and message; a control mutation is valid syntax with forbidden semantics

Governs **AC25a**, and applies program-wide from this story onward. Promoted from this story's own
M-6a. **The lead's assessment: the sharpest process finding in this batch**, because it names a
failure mode D-000.9 provably cannot catch — a test that fails *for the wrong reason* has both run
and failed, so both existing diagnostic questions pass it.

### Standing rules carried

**D-000.4** (AC28 — matrix deferred, explicitly not green), **D-000.9 + extension** (AC25),
**D-000.10** (AC30), **D-000.11** (AC26), **D-000.12** (AC27), **D-1.2.6** (the parked conflict
above), **D-1.3.1** (AC29), **D-1.4.6** (the `os` boundary is the package boundary — this is *why*
AC12 exists), **D-1.4.8 corrected** (the presence seam AC11 consumes), **D-1.4.15** (mechanism
tagging), **D-1.4.16** (the seam's D-000.9 shape, which AC11 closes), **D-1.5.5** (insertion order),
**D-1.1.b** (AC12's `numbers.go` precedent, and the paraphrase-widening lesson the parked conflict
turns on), **D-1.1.c** (the `Data`/`Params` named-type target).

---

## Tasks / Subtasks

1. [x] **Create `folio-go/internal/bind/`.** Confirm first that no absence tripwire fires — `absenceChecks`
   guards `folio-designer`, `folio-go/fonts`, `folio-go/internal/expr`, `folio-go/internal/diag`;
   **`internal/bind` is not among them** (verified, `lint/internal/rules/absences.go:43–71`).
   The new package inherits every `internal/` guard automatically (AC29, AC5).

2. [x] **Close the denial-of-service inside the shared splitter (AC4, AC4a, AC4b, AC6) — do this
   FIRST.** Two fixes in `parseDecimalExponent`, **both** required: (i) an overflow check **during**
   accumulation, since a post-loop check reads an already-wrapped value; (ii) a documented magnitude
   bound rejected **before any `big.Int` scaling is attempted**, since the scaling *is* the hang.
   Layering per AC4b: the splitter rejects the absurd and its bound is the **wider** of the two;
   `decodePoints` keeps its millipoint check and `internal/bind` applies `Decimal`'s. **Test at the
   bound and one past it (binding).** Confirm `decimal_test.go` and the golden round-trip still
   pass — the bound must not reject any literal the corpus legally contains.

2a. [x] **Ship the retained corpus fixture (AC4c).** A `.folio` containing `1e99999999999999999999` in a
   geometry field must produce a **located error, quickly**, through the **public loader** — not
   only through the unit test. **Red-proof: before the fix this test HANGS.** Record the baseline
   verbatim as *"hung; killed by `-timeout` at Ns"* (M-1 has the measured wording); **never** dress
   a timeout panic up as an assertion failure (D-1.2.5, AC4c).

2b. [x] **Fill in the fixing commit hash in Story 1.4's appended `## Defect found after done` section**
   once this story is committed. The section is already written; that one field could not be.
   **Append only — do not rewrite anything above that heading.**

3. [x] **Build `Decimal` over the shared splitter (AC1, AC2, AC3, AC3a, AC6).** Implement significance
   as *strip leading zeros, never trailing zeros* (M-3). Table-test AC2's four pairs, M-3's two
   boundary literals (`0.00000000000000000001` legal, `12345678901234567890` an error) **and AC3a's
   all-zeros set** (`0.00` → `{0,-2}`, `0.0`, `0`, `-0.00`) — the empty-digit-string case is the one
   a naive implementation gets wrong in both available directions.

4. [x] **Build the generic report-data value tree (AC7, AC24)** with `UseNumber`, separate from
   `internal/template`'s schema-typed decoder, sharing only the splitter.

5. [x] **Implement the path matcher and the grammar (AC15, AC16, AC17).** Scope it to **text-element
   `value` interpolation only** (AC20). Add a fixture asserting `{{formatNumber(a, "x")}}` in a
   text element is a located error naming the element id **and** mentioning Epic 3.

6. [x] **Implement the `page`/`pages` reservation (AC18) and its retained fixture (AC19).** Red-proof:
   delete the reservation branch, watch the fixture go red because a `page` key in the data captured
   the slot. **State what the fixture prints when the mutation is absent** (AC25).

7. [x] **Implement AD-14's three cases and ship the fixture triple (AC8–AC11)** over **one** template and
   **one** path. Red-proof by collapsing `Presence` to two states and confirming row 1 or row 2 goes
   red.

8. [x] **Split `folio-go/folio.go` (AC12, AC13, AC14).** Move `Render` (and, at 1.7, `RenderTo`) into a
   file importing **none** of `os`/`time`/`net`/`math/rand`; leave `LoadTemplate` in the
   `os`-importing file. Put AD-12's locale clarification in the new file's doc comment (AC14). Assert
   the property from the existing rules module with a coverage witness — a run that parsed zero files,
   or found the render file absent, **fails** (AC25).

8a. [x] **Add the second production caller for `float64` (AC5, AC5a, D-1.6.7).** Point the **existing**
   checker at the **whole `folio-go/` module**, `_test.go` included, skipping `testdata/` and
   dot-directories **by category** — never by naming a specific directory. **Keep the existing
   `internal/`-scoped caller and its `geom`/`pdf` vacuity guard.** Two callers, one checker; **no
   new checker.** It goes green on the first run (measured); red-proof it by reapplying M-6's
   mutation and confirming the new caller reports the **rule id**, then revert.

9. [x] **Change the signature and update all thirteen call sites (AC22, AC23, M-7)**, inserting `d`
   between `t` and `f`. Declare `type Data []byte` as a **defined type**.

10. [x] **Run the gates with `-count=1` (AC26).** `go build ./...`, `go vet ./...`, `go test -count=1 ./...`
    in `folio-go/`, `lint/` and `hashmatrix/`. Verify any byte/hash comparison through `rtk proxy` to
    a file, never a pipe (AC27). **Record explicitly that the cross-target matrix was NOT run and is
    deferred to the Epic 1 boundary (AC28).**

11. [x] **Fill the D-000.10 disposition table (AC30)** — one line per ruling ID enumerated below, no gaps.

12. [x] **Set `sprint-status.yaml` `1-6-bind-scalar-json-values-into-text` to `review`** when
    implementation is complete. **Do not commit and do not set `done`.**

> **Not a task: committing.** This breakdown ends at "status → review". The finisher owns the commit.

---

## Dev Notes

Every item below was measured at baseline `5a1cf7d` while this file was written. Working tree was
verified restored afterwards: `git diff --stat -- '*.go'` empty, only the two expected doc files
modified.

### M-1 — the shipped exponent parser wraps, and the wrapped value HANGS the public loader

**This is the story's most serious finding and it is a shipped denial-of-service, not a latent bug.**
Ruled as D-1.6.6.

`parseDecimalExponent` accumulates with `n = n*10 + int(c-'0')` and **no overflow check**. Isolated
per literal with Go's own `-timeout`, each in its own test function:

| Literal | `splitJSONNumber` | `decodePoints` |
|---|---|---|
| `1e2000000` | `exp=2000000`, `err=nil` | **errors correctly, 0.07s** — the control |
| `1e99999999999999999999` | `exp=7766279631452241919`, `err=nil` | **hung; killed by `-timeout` at 15s** |
| `1e9223372036854775808` | `exp=-9223372036854775808`, `err=nil` | **hung; killed by `-timeout` at 15s** |
| `1e-99999999999999999999` | `exp=-7766279631452241919`, `err=nil` | *(not separately timed)* |

A **positive** exponent literal becomes the **most negative `int64`**.

**The mechanism, from the panic trace** (~74 `math/big` frames in each hang): the wrapped exponent
becomes `decodePoints`' `shift`, which feeds
`new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(shift)), nil)`. The stack sits in
`math/big.nat.sqr` / `karatsubaSqr` (`natmul.go:101/105/296/300`) — `Exp`'s repeated squaring,
building a number with ~10^18 digits. It does not error late; it does not error at all.

**Reachable through the public API — measured end-to-end, not inferred.**
`fixtures/font-text/input.folio` with element `e1`'s `x` set to `1e99999999999999999999` (a
syntactically valid `.folio`), passed to **`folio.LoadTemplate`**:

```
panic: test timed out after 20s
    running tests:
        TestProbePublicLoaderHang (20s)
```

…with the same 74 `math/big` frames. **Story 1.4's user story promises *"malformed … templates
rejected with a located error"*.** Epic 5's designer loads user-authored templates, so this is
reachable from untrusted input.

> **Per D-000.12, both hangs and the public-loader probe were captured through `rtk proxy` into
> files.** The wrapper's own pipe truncated the first panic to a bare `[FAIL] … 15.199s` line, which
> would have hidden the `math/big` stack entirely — the mechanism would have been guesswork.
>
> **Reuse these exact words for the red-proofs (AC4c, D-1.2.5):** *"hung; killed by `-timeout` at
> Ns"*. **Never** record a timeout panic as an assertion failure.

### M-2 — Story 1.4's path is *not* exposed to the amplification hazard; 1.6's is

The reason AC4 is genuinely new surface rather than a restatement:

```
decodePoints("1e100000")  -> err="overflows int64 millipoints"                  1.8ms
decodePoints("1e-100000") -> err="has more than three decimal places"           1.4ms
decodePoints("1e2000000") -> err="overflows int64 millipoints"                  72.3ms
```

`geom.Length`'s int64 millipoint bound catches these promptly. **`Decimal` has no such downstream
bound on the value** — it stores `{coefficient, exponent}` and the blowup happens at **render**, when
digits are expanded. That is exactly D-1.6.1's framing: *"representable in the struct but renders to
a gigabyte of digits."*

### M-3 — "significant digits" needs a definition, and the obvious length check is wrong

```
"0.00000000000000000001" -> int="0"  frac="00000000000000000001"  exp=0   (21 raw digits, 1 significant)
"12345678901234567890"   -> int="12345678901234567890" frac=""    exp=0   (20 significant → error)
"1.50"                   -> int="1"  frac="50"                    exp=0
```

`len(intPart+fracPart) > 19` **rejects a legal literal**. And stripping trailing zeros to "normalise"
would turn `1.50` into `{15,-1}` and violate AC2. **Strip leading zeros; never strip trailing zeros.**

**The all-zeros edge case (AC3a), measured after the lead flagged it** — it strips to the **empty
string**, which a naive implementation either errors on or silently de-scales:

```
"0.00"  -> int="0" frac="00" exp=0   digits="000"  → strip leading zeros → ""   must yield {0,-2}
"0.0"   -> int="0" frac="0"  exp=0   digits="00"   → ""                        must yield {0,-1}
"0"     -> int="0" frac=""   exp=0   digits="0"    → ""                        must yield {0, 0}
"-0.00" -> sign=-1, otherwise as "0.00"
```

The scale must survive even though every significant digit is gone. This is the case where "strip
leading zeros" and "never lose the exponent" pull against each other, and it is the one a naive
implementation gets wrong in both available directions.

### M-4 — the canonical corpus contains an expression-shaped placeholder, in a field 1.6 does not own

Four `{{…}}` occurrences in `folio-go/testdata/template/golden/worked-example.json`, in two schema
fields (table AC20 has the full breakdown). Line 19 is
`"bind": "{{formatNumber(transaction.amount, \"#,##0.00\")}}"` — parentheses, a comma **and** quotes,
i.e. three separate AC16 rejection triggers.

It lives on `Column.Bind` (`internal/template/model.go:192`), not on a text element's `Value`.
`collectTextRuns` (`folio-go/render.go`) skips every element whose `Type != ElementText`, so the
renderer never reaches it today. The golden fixture is **round-tripped, never rendered**
(`roundtrip_test.go:18`, `goldenfixture_test.go:17`), so a *render-time* rejection cannot break it
either — but a *document-wide scan* of all `{{…}}` would. Hence AC20.

`fixtures/font-text/input.folio` — the only `.folio` file in the repo — contains **no** `{{`.

### M-5 — `Render` and `LoadTemplate` share a file today, so AC12 is real work

`folio-go/folio.go` declares `LoadTemplate`, `ParseTemplate` **and** `Render`, and imports `os` for
`LoadTemplate`. `folio-go/render.go` (`renderDocument`, `collectTextRuns`, `resolveFace`,
`pageDimensions`) imports `fmt`, `maps`, `slices` and four `internal/` packages — none of the four
banned. So the split is `Render` moving out of `folio.go`, not `LoadTemplate` moving out of
`render.go`.

### M-6 — the shell gap, injected and measured both ways

**Mutation applied** — `os.Getenv("FOLIO_MUTATION_PROBE")` and `var probeFloat float64 = 1.5`
inserted into `folio.Render`'s body:

| Gate | Result |
|---|---|
| `go build ./...` | **OK** |
| `lint/` — `go test -count=1 ./...` | **39 passed, 4 packages, 0 failed** |
| `folio-go/internal/...` — `go test -count=1 ./...` | **157 passed, 5 packages, 0 failed** |

**Same mutation, moved under `internal/geom/`** (the counterfactual that makes the above evidence
rather than an observation about dead guards):

```
TestForbiddenImportsProductionScan  FAIL  geom/zzprobe.go:3: forbidden import "os" (AD-1 …)
TestNoFloat64UnderInternal          FAIL  geom/zzprobe.go:5: identifier float64
                                          geom/zzprobe.go:5: untyped floating-point literal 1.5
```

So: **the guards are alive and the module root is simply outside their target.** This confirms
D-1.6.3's premise (closed by AC12) **and** surfaced the `float64` half, now ruled as D-1.6.7.

**Independently verified for D-1.6.7:** `grep "float64\|float32" folio-go/*.go` → **zero matches**,
`_test.go` files included. The new production caller (AC5a) therefore goes **green on its first
run**, and its red-proof is exactly the mutation above.

### M-6a — the false red, and why D-000.9 cannot catch it *(promoted to D-000.13)*

**My first control attempt appended an `import` after code, producing a parse error.** Both scanners
went red — but with `scan folio-go/internal/: … expected declaration`, i.e. `D-1.3.3 (amended)`'s
**error path**, **not a rule finding**. Had I recorded it, it would have read as proof the guard
works.

**This is a distinct defect class from D-000.9's.** D-000.9 asks *"what would this have printed if
it had been unable to run at all?"* — and this test **did** run and **did** fail, so the question
passes it. It is not a proof that cannot fail; **it is one that fails without exercising the
property at all.** The new question, now D-000.13: ***did it fail for the reason it names?***

**Two consequences for whoever writes AC12's and AC5a's assertions:**

1. Assert on the finding's **rule id and message**, never on non-zero exit or mere failure.
2. **A control mutation must be valid syntax with forbidden semantics.** The corrected control
   (`zzprobe.go`, a well-formed file importing `os` and declaring `var probeF float64 = 1.5`) is
   what produced the real findings quoted above.

> **The irony is worth preserving:** `D-1.3.3 (amended)`'s error path exists *because* a crashed
> scan must never read as clean. That safeguard is precisely what made this false red possible.

### M-7 — thirteen `Render(` call sites, all in tests

`folio-go/template_test.go:114`; `fixture_test.go:396`; `render_test.go:124, 214, 573, 610, 641,
732, 757, 857`. All two-argument, all needing `d` inserted in the middle. No production call site
outside package `folio`.

> **Correction (QA Finding 11, this story's review, Minor; resolved by the finisher).** This
> paragraph says "thirteen" and "four" test files, then names three files and lists ten line
> numbers. Measured at baseline `5a1cf7d`: `render_test.go` 9, `template_test.go` 1,
> `fixture_test.go` 1 = **11**, across **three** files, not four. The Delivery Log correctly says
> "11 call sites … updated" but never flagged the discrepancy against this stated reference figure —
> D-1.4.15's own discipline is that a departure from a stated mechanism is surfaced, not silently
> corrected past. No code was wrong; all 11 real call sites were updated. This paragraph is left in
> place, uncorrected, as the historical record the reviewer's Finding 11 refers to; the correction is
> this note.

### M-8 — the presence seam is present and three-state at baseline

`internal/template/presence.go` ships `Presence[T]{Set bool; Null bool; Value T}` with `present`,
`presentNull` and `absent` constructors, populated by hand in `parse.go` from a
`map[string]json.RawMessage` lookup. This is the seam AC11 consumes and the reason D-1.4.16 called
its third state unread at 1.4. Story 1.6 is where it becomes load-bearing.

### M-9 — creating `internal/bind` trips no tripwire

`absenceChecks` (`lint/internal/rules/absences.go:43–71`) guards exactly four paths:
`folio-designer`, `folio-go/fonts`, `folio-go/internal/expr`, `folio-go/internal/diag`.
`internal/bind` is not among them, and `ScanAbsences`' `ChecksEvaluated` witness is unaffected.
Conversely, **DW-8's forcing function is real**: the day Story 3.2 creates `internal/expr`,
`absence-expr-package` fires and someone re-reads DW-8.

---

## Ruling disposition table (D-000.10)

The developer fills the right-hand column, **one line per row, no row left blank**; the reviewer
mirrors it independently. **This enumeration is the complete set of ruling IDs this story
enumerates** — D-000.10 cannot build the table if the list is partial.

| Ruling | Disposition (developer) |
|---|---|
| **D-1.6.1** | applied-as-stated — `Decimal{Coefficient int64; Exponent int}` in `internal/bind/decimal.go`, built over `template.SplitJSONNumber` (the exported, one-and-only splitter). Coefficient bound by significant digits (leading zeros stripped, trailing never); `0.00` -> `{0,-2}` per AC3a. |
| **D-1.6.2** | applied-as-stated — `TestAD14Triple` (`internal/bind/text_test.go`) is the retained triple over one template text and one path (`customer.name`); rows 1/2 red-proofed by a Presence-collapse mutation (measured, restored). |
| **D-1.6.3** | applied-as-stated — `Render`/(future `RenderTo`) moved to the new `folio-go/render_entry.go`, which imports none of os/time/net/math/rand; `LoadTemplate`/`ParseTemplate` stay in `folio.go`, which still imports "os". Property asserted via `lint`'s existing `ScanForbiddenImports`, pointed at the single file (`TestRenderEntryFileHasNoForbiddenImports`) — not a new checker, not a package-wide scan (which would false-positive on `folio.go`'s legitimate "os"). |
| **D-1.6.4** | applied-as-stated — `func Render(t *Template, d Data, f FontSet) ([]byte, error)`; `d` inserted between `t` and `f` per D-1.5.5; `type Data []byte` is a defined type, not an alias. Two decode trees (`internal/template`'s schema tree, `internal/bind.Value`'s generic tree) sharing only `SplitJSONNumber`. |
| **D-1.6.5** | applied-as-stated — grammar `"{{" ws? ident ("." ident)* ws? "}}"` in `internal/bind/text.go` (`parseBindingPath`/`isValidIdent`); anything else is a located error naming the element id and mentioning "Epic 3"; `{{page}}`/`{{pages}}` reserved and left byte-for-byte unchanged, red-proofed by deleting the reservation branch (measured, restored). |
| **D-1.6.6** | applied-as-stated — two fixes in `parseDecimalExponent` (`internal/template/decimal.go`): an overflow check during accumulation, and a documented magnitude bound (`MaxSplitExponentMagnitude = 1,000,000`) checked before any `big.Int` scaling. Pre-fix hang measured and recorded verbatim ("hung; killed by -timeout at 15.436s", reproducing M-1's ~74-frame `math/big` stack) — see Delivery Log. Retained corpus fixture (`testdata/template/malformed/huge-exponent.folio`) proves the fix through the public `folio.LoadTemplate` (`TestLoadTemplateRejectsHugeExponentQuickly`, 0.197s). Story 1.4's fixing-commit-hash field is left blank for the finisher (D-1.6.6's own instruction) — not filled by this developer. |
| **D-1.6.7** | applied-as-stated — the SECOND production caller (`TestNoFloat64UnderModule`, `internal/arch_test.go`), not a second checker: reuses `scanNoFloat64`, points it at the whole module (`filepath.Dir(internalRoot)`), `_test.go` included, `testdata/` and dot-directories skipped by category (`walkGoFiles` extended). Existing `TestNoFloat64UnderInternal` kept unmodified. Green on first run; red-proofed by reapplying M-6's `float64`/untyped-float mutation into `render_entry.go` (measured, restored) — the new caller alone went red, naming `identifier float64` / `untyped floating-point literal 1.5`. |
| **D-1.6.8** | applied-as-stated — AC20's field-scope fence holds by construction: `internal/bind.BindText` is called from exactly one site, `collectTextRuns`, which already skips every non-text element; the table's `bind`/`columns[].bind` fields (expression-shaped, per M-4) are never scanned. Proved with a dedicated render-time test (`TestRenderScopeFenceIgnoresTableBind`), reproducing worked-example.json:19's shape, kept in the default gate (no build tag). |
| **DW-8** | applied-as-stated — verified the shipped code matches what `deferred-work.md`'s DW-8 already records: `internal/bind` does not import `internal/expr` (which does not exist), and no `internal/expr` package was created. Not re-registered (that was never this story's job). |
| Story 1.4 `## Defect found after done` (fixing commit hash) | not-reached-in-this-story (by design) — the section is already fully written (verified append-only, untouched by this story); the one blank field (fixing commit hash) is explicitly left for the finisher to fill in once this story is committed, per the story's own instruction. |
| D-000.4 | applied-as-stated — the cross-target hash matrix was NOT run in this story; explicitly recorded as deferred to the Epic 1 boundary. `go build`/`go vet`/`go test -count=1` and `-tags=matrix` build/vet did run, all green. |
| D-000.9 | applied-as-stated — every new check carries a vacuity guard (e.g. `TestRenderEntryFileHasNoForbiddenImports`'s `stats.FilesSeen != 1`; `TestNoFloat64UnderModule`'s `packagesVisited["."]`/`["internal"]` and `declsSeen != 0`). |
| D-000.9 (extended) | applied-as-stated — every red-proof in this story was measured by actually reverting the fix (via `cp`, never `git checkout`) and re-running, not merely asserted: the DoS hang, the AC12 mutation, the AC5a mutation, the AD-14 Presence-collapse, and the page-reservation deletion were all observed red before being restored to green. |
| D-000.10 | applied-as-stated — this table; every enumerated ruling ID has a disposition line. |
| D-000.11 | applied-as-stated — every gate run in this story used `-count=1` explicitly. |
| D-000.12 | applied-as-stated — every hash/byte/panic-stack observation was captured via `rtk proxy … > file`, never a shell pipe (the hang probe's `math/big` stack, all vet/test raw outputs, the mutation red-proofs). |
| **D-000.13** | applied-as-stated — every red-proof in this story asserts on the finding's rule id and message text (e.g. `forbidden import "os"`, `identifier float64`, `data path "customer.name" is absent`), never on exit status alone; every control mutation used (os import, float64 var, Presence collapse, reservation-branch deletion) is valid Go / valid logic with forbidden semantics, never a parse error. |
| D-1.1.b | applied-as-stated — AC12's file-scoped, decidable-property shape (`render_entry.go` imports none of the four) mirrors `numbers.go`'s precedent exactly, asserted via the existing scanner pointed at one file. |
| D-1.1.c | applied-as-stated — `type Data []byte` is a named defined type, never `[]byte` directly and never a type alias, preserving the compile-time `Data`/`Params` distinctness this ruling exists for. |
| D-1.2.5 | applied-as-stated — the pre-fix hang is recorded verbatim as "hung; killed by `-timeout` at 15.436s", never dressed up as an assertion failure (both here and in the corpus fixture test's doc comment). |
| D-1.2.6 | not-reached-in-this-story — the one conflict this ruling would have governed (AC5's three-way scope collision) was already surfaced and ruled as D-1.6.7 before development began; no new DECISION NEEDED was raised during implementation. |
| D-1.3.1 | applied-as-stated — `internal/bind`'s non-test files import none of time/os/math/rand/net/math-transcendentals; its `_test.go` files (`decimal_test.go`, `text_test.go` — neither needed anything beyond `testing` and `strings`) stays within the closed `_test.go` exemption. Verified via the existing `TestForbiddenImportsProductionScan` (folio-go/internal/), which now covers `internal/bind` automatically. |
| D-1.3.3 (amended) | applied-as-stated — every control mutation used in this story's red-proofs is valid syntax with forbidden semantics (an `os` import, a `float64` var, a Presence collapse, a deleted-but-syntactically-fine branch) — never a parse error standing in for a rule finding. |
| D-1.4.6 | applied-as-stated — the "os" boundary stays at the package (`folio.go` keeps `os` for `LoadTemplate`); AC12 narrows this further to a FILE boundary for `Render` specifically, which is why AC12 exists as new surface rather than a restatement of D-1.4.6. |
| D-1.4.8 (corrected) | applied-as-stated — `internal/template.Presence[T]`'s three-state seam (Set/Null/Value) is the exact mechanism `internal/bind.Value.Lookup`'s `Presence` (Absent/Null/Present) enum mirrors for report data; AD-14's triple is where it becomes load-bearing. |
| D-1.4.15 | applied-as-stated — every AC/ruling reference in this story's new code comments follows the binding-vs-illustrative mechanism-tagging convention (e.g. `MaxSplitExponentMagnitude`'s doc comment explicitly marks the specific number as illustrative, the existence/enforcement-order of the bound as binding). |
| D-1.4.16 | applied-as-stated — this story is where production code (`internal/bind.Value.Lookup`, consumed by `BindText`) first branches on `Null` vs absent, closing the exposure D-1.4.16 named. |
| D-1.5.5 | applied-as-stated — `d` is inserted between `t` and `f`, matching the target `Render(t, d, p, f)`; no existing call site was reordered, only extended (all 11 production/test call sites updated by inserting `Data("{}")` in the middle). |
| AD-1 | applied-as-stated — `internal/bind` inherits the full import ban; verified via the existing `ScanForbiddenImports`/`ScanMapRange` production scans over `folio-go/internal/`, which now include it automatically. |
| AD-2 | applied-as-stated — no `float64` anywhere in `internal/bind` or the module-root package (verified via both `TestNoFloat64UnderInternal` and the new `TestNoFloat64UnderModule`); this is also why AC5a's option (a) (a Render-signature-only check) was rejected — AD-2 makes the whole public signature surface relevant, not one file. |
| AD-4 | applied-as-stated — `{{page}}`/`{{pages}}` are reserved, never resolved from data, never an error; no `page` namespace was added to the binding grammar. |
| AD-12 | applied-as-stated — `render_entry.go`'s file-level doc comment states the locale clarification: reading the document's own declared locale is required, discovering the host's is banned. Not otherwise exercised by this story's rendering logic (no locale-aware formatting exists yet). |
| AD-14 | applied-as-stated — the three cases (absent/null/wrong-kind) are implemented exactly via `internal/bind.Value.Lookup`'s `Presence` enum and `BindText`'s switch over it, proved by the retained fixture triple. |
| AD-23 | applied-as-stated — `internal/bind.DecodeData` uses `json.Decoder.UseNumber`; `Decimal` is the exact scaled-integer representation; no `float64` is produced anywhere in the binding path. |

---

## Dev Agent Record

### Delivery Log

- [x] Task 2 done FIRST, as instructed: DoS fixed in `parseDecimalExponent` (`internal/template/decimal.go`) before any other code was written. Two independent checks, both required (AC4a): (i) an overflow check DURING accumulation (`if n > (maxInt-d)/10`, checked before the multiply, never after); (ii) a documented magnitude bound (`MaxSplitExponentMagnitude = 1_000_000`, illustrative number, binding that it exists and is enforced before any `big.Int` scaling).
- [x] Pre-fix red measured and recorded verbatim (D-1.2.5, AC4c): baseline `decimal.go` restored via `cp` (never `git checkout`), `TestProbeHangExponentWrap` run with `go test -timeout=15s`, captured through `rtk proxy … > file` (D-000.12) — **"hung; killed by `-timeout` at 15.436s"**, panic trace rooted at `decodePoints` (`decimal.go:48`) through ~30 `math/big.karatsubaSqr`/`nat.sqr` frames into `math/big.nat.expNN`, reproducing M-1's mechanism exactly. Fix restored via `cp` from the pre-edit backup immediately after capture; probe test file deleted.
- [x] Post-fix: all three pathological literals (`1e99999999999999999999`, `1e9223372036854775808`, `1e-99999999999999999999`) now error in ~0.5s total (measured, `TestParseDecimalExponentDoesNotHang`), and `1e2000000` (Story 1.4's own control) still errors, now at the shared splitter rather than at `decodePoints`' millipoint check (`TestSplitJSONNumberExponentBound`).
- [x] AC4c retained corpus fixture shipped: `folio-go/testdata/template/malformed/huge-exponent.folio` (element `e1`'s `x` = `1e99999999999999999999`), proved through the PUBLIC loader — `TestLoadTemplateRejectsHugeExponentQuickly` (`folio-go/template_test.go`), measured at 0.197s via `rtk proxy` (raw, not piped).
- [x] Task 2b: Story 1.4's `## Defect found after done` section left untouched — verified it is already fully written and merely references "the fixing commit named in that story's Delivery Log" (i.e., this story's own Delivery Log, not a field inside 1.4's file itself). **No commit hash is filled in anywhere by this developer** — that is the finisher's field once this story is committed, per the story's own explicit instruction.
- [x] `internal/bind` created (Task 1); confirmed no absence tripwire fires (`lint/internal/rules/absences.go`'s `absenceChecks` does not name it) — the full suite run below includes `lint`'s absence scan, unaffected.
- [x] `Decimal` built over the shared, now-exported `SplitJSONNumber` (Task 3): AC2's four pairs, AC3's coefficient bound (significant-digit counting, leading zeros stripped/trailing never), AC3a's all-zeros set (`0.00`→`{0,-2}`, `0.0`→`{0,-1}`, `0`→`{0,0}`, `-0.00`→`{0,-2}`), AC4b's own tighter exponent bound (100,000; tested at the bound and one past it) — `internal/bind/decimal.go` + `decimal_test.go`, 5 test functions, all green.
- [x] Generic report-data value tree built (Task 4): `internal/bind/value.go`, `Value`/`Kind`/`Presence`/`Lookup`/`DecodeData`, using `json.Decoder.UseNumber` (AC24). Map ranging avoided via the `slices.Sorted(maps.Keys(...))` escape hatch (D-1.3.5), mirroring `internal/template`'s `rawFromAny`.
- [x] Grammar/path matcher built (Task 5): `internal/bind/text.go`, `BindText`/`parseBindingPath`/`isValidIdent`. Scoped to text-element `value` interpolation only by construction — `BindText` has exactly one call site, `collectTextRuns` (`folio-go/render.go`), which already skips every non-text element (AC20).
- [x] `{{page}}`/`{{pages}}` reservation built (Task 6): checked before any data lookup, so it can never be shadowed (AC18). Red-proofed by deleting the reservation branch (`cp`-backed mutation, restored) — `TestBindTextReservesPageAndPages` and `TestBindTextPageReservationIsRedProofable` both went red, naming `data path "page"/"pages" is absent from the report data`, confirming the hazard AC19 exists to catch.
- [x] AD-14's three cases + retained fixture triple built (Task 7): `TestAD14Triple` over one template text (`"Statement for {{customer.name}}"`) and one path (`customer.name`), three subtests. Red-proofed by collapsing `Presence` to two states (Null treated as Absent) — row 2 (and row 3, which depends on Presence reaching the Kind check) went red for exactly the collapsed-state reason; row 1 stayed green (still Absent either way), matching the ruling's own prediction. Restored via `cp`.

  > **Correction (QA Finding 12, this story's review, Minor; resolved by the finisher).** This
  > bullet overstates which rows redden. The reviewer applied three distinct `Presence` collapses
  > and measured: under the ruling's own scenario (Null→Absent, `Present` kept — the exact one this
  > bullet describes), **only row 2** reddens; row 3 stays **green**, because row 3's data
  > (`{"customer": {"name": 123}}`) is still `Present` with `Kind == KindNumber`, and the wrong-kind
  > check in `BindText`'s `Present` arm never depended on the Null/Absent distinction. Row 3 only
  > reddens under a third, degenerate collapse (everything → `Absent`) that also breaks `Present`
  > and reddens unrelated tests — not the one this bullet applied. The substance is unaffected:
  > D-1.6.2 is genuinely satisfied, rows 1 and 2 do produce opposite outcomes under the collapse
  > actually applied, and the triple cannot pass under any of the three collapses tried. Only the
  > attribution of row 3's reddening was wrong.
- [x] `folio.go` split (Task 8): `Render` moved to the new `folio-go/render_entry.go`, which imports none of os/time/net/math/rand (AC12); `LoadTemplate`/`ParseTemplate` stay in `folio.go` (still imports "os", D-1.4.6). AD-12's locale clarification is in `render_entry.go`'s file-level doc comment (AC14). Property asserted via `lint`'s existing `ScanForbiddenImports`, pointed at the single file (`TestRenderEntryFileHasNoForbiddenImports`, `lint/internal/rules/forbiddenimports_test.go`) — not a new checker (AC13). Vacuity guard: `stats.FilesSeen != 1` fails the run; a missing file already fails via `ScanForbiddenImports`'s own walk error. Red-proofed by adding an `os` import to `render_entry.go` (`cp`-backed mutation, restored) — went red naming `forbidden import "os"` at the correct line.
- [x] Second production `float64` caller added (Task 8a): `TestNoFloat64UnderModule` in `folio-go/internal/arch_test.go`, reusing `scanNoFloat64` (not a new checker), pointed at the whole module root. `walkGoFiles` extended to skip dot-directories by category (never by name) alongside the existing `testdata` skip. Green on first run (measured, matches the story's own prediction from M-6's independent `grep`). Red-proofed by reapplying M-6's exact mutation (`var probeFloat float64 = 1.5`) into `render_entry.go` (`cp`-backed, restored): `TestNoFloat64UnderInternal` stayed green (module root genuinely outside its scope — this is the evidence, not an assumption) while `TestNoFloat64UnderModule` alone went red, naming `identifier float64` and `untyped floating-point literal 1.5`.
- [x] Signature changed and all call sites updated (Task 9): `func Render(t *Template, d Data, f FontSet) ([]byte, error)`; `type Data []byte` declared in `render_entry.go` as a defined type (AC23). 11 call sites across `fixture_test.go`, `render_test.go`, `template_test.go` updated by inserting `Data("{}")` between the template and font-set arguments (D-1.5.5: extended, never reordered).
- [x] AC20's field-scope fence proved at render time, not only by round-trip (Task 5/D-1.6.8): `TestRenderScopeFenceIgnoresTableBind` (`folio-go/render_bind_test.go`) builds a document reproducing worked-example.json:19's shape (a table `bind`/`columns[].bind` containing `{{formatNumber(transaction.amount, "#,##0.00")}}` — parentheses, a comma, quotes) alongside a valid text-element placeholder, and confirms `Render` succeeds.
- [x] Additional end-to-end coverage through the public `Render` API: `TestRenderBindsTextPlaceholder` (AC15), `TestRenderAbsentPathIsLocatedError` (AC8), `TestRenderNullPathRendersEmpty` (AC9), `TestRenderWrongKindIsLocatedError` (AC10), `TestRenderRejectsExpressionSyntaxInTextValue` (AC16/AC17) — all in `folio-go/render_bind_test.go`.
- [x] Gates run (Task 10), every one `-count=1` (AC26), raw exit codes/output confirmed via `rtk proxy … > file`, never a pipe (AC27, D-000.12) — see below. Cross-target hash matrix explicitly NOT run and recorded as deferred to the Epic 1 boundary (AC28, D-000.4).
- [x] D-000.10 disposition table filled (Task 11) — see [§ Ruling disposition table](#ruling-disposition-table-d-00010). All rows present in the table have a disposition; none left blank.
- [x] `sprint-status.yaml` set to `review` (Task 12). No commit made; no status set to `done` — the finisher's job.

#### D-000.9 answered for the checks this story adds or extends

1. **`TestLoadTemplateRejectsHugeExponentQuickly`** (AC4c) — a loader that silently accepted the huge-exponent literal, or one whose `LoadTemplate` call never actually ran, would both report `err == nil`; the test asserts `err != nil` AND `tpl == nil` AND the error names `"e1"`, so a vacuously-passing nil template alongside a nil error cannot slip through.
2. **`TestRenderEntryFileHasNoForbiddenImports`** (AC12) — `stats.FilesSeen != 1` fails the run explicitly; a target path that no longer exists (the render file renamed or moved) already fails via `ScanForbiddenImports`' own walk error, never silent zero findings (inherited from `D-1.3.3 (amended)`).
3. **`TestNoFloat64UnderModule`** (AC5a) — asserts, from `scanNoFloat64`'s OWN returned stats (not a second, independently-derived walk), that both `"."` (module-root package) and `"internal"` were visited, and that `declsSeen != 0` — a scanner pointed at the wrong root, or one that silently walked nothing, is distinguished from a healthy zero-findings run.
4. **`TestAD14Triple`** — row 1 (absent) and row 2 (null) are the proof by construction: they differ only in the data and must produce opposite outcomes, so the fixture cannot pass if the three-state distinction ever collapses (measured directly, not merely asserted — see the red-proof above).
5. **`TestBindTextReservesPageAndPages`/`TestBindTextPageReservationIsRedProofable`** — the latter compares `{{page}}`'s rendering WITH and WITHOUT a top-level `"page"` data key and asserts they are byte-identical, so a reservation that silently stopped reserving (rather than erroring) would still be caught, not just a reservation that started erroring.

#### D-000.13 — every red-proof answers "did it fail for the reason it names?"

Every control mutation applied during this story's development (listed below) was valid Go syntax with forbidden semantics — never a parse error — and every assertion that observed the resulting red checked the specific rule id / finding message / error text named, never bare non-zero exit or `err != nil` alone:

| Mutation | Applied to | Observed red | Restored via |
|---|---|---|---|
| Baseline (pre-fix) `decimal.go` | `internal/template/decimal.go` | `panic: test timed out after 15s` + `math/big` stack, message recorded verbatim | `cp` from post-fix backup |
| `os` import + `os.Getpid` reference | `folio-go/render_entry.go` | `forbidden import "os"` (rule `forbidden-imports`) | `cp` from pre-mutation backup |
| `var probeFloat float64 = 1.5` | `folio-go/render_entry.go` | `identifier float64` + `untyped floating-point literal 1.5` (rule `no-float64`) — only in the NEW module-wide caller, confirming the internal-scoped caller's genuine blind spot | `cp` from pre-mutation backup |
| `Presence` collapsed to two states (Null→Absent) | `internal/bind/value.go`'s `Lookup` | Row 2 (and row 3) of `TestAD14Triple` failed with `data path "customer.name" is absent…` | `cp` from pre-mutation backup |
| Reservation branch deleted | `internal/bind/text.go`'s `BindText` | `TestBindTextReservesPageAndPages`/`...RedProofable` both failed with `data path "page"/"pages" is absent…` | `cp` from pre-mutation backup |

Every restoration used `cp`, never `git checkout` (feedback memory: red-proof hygiene) — `git status --porcelain` was re-checked after each restoration to confirm no stray diff remained before moving to the next task.

> **Corrections (resolved by the finisher):**
>
> - **Row 4 (QA Finding 12).** "Row 2 (and row 3)" overstates it: under this exact collapse
>   (Null→Absent, `Present` kept), only row 2 fails. Row 3's data (`{"customer": {"name": 123}}`)
>   is still `Present`/`KindNumber` under this collapse — the wrong-kind check that reddens row 3
>   does not depend on the Null/Absent distinction at all. See the Task 7 bullet's correction above
>   for the full account.
> - **Row 5 (QA Nit 15).** True as recorded for `TestBindTextReservesPageAndPages`, but
>   `TestBindTextPageReservationIsRedProofable`'s original form compared data WITH a top-level
>   "page" key against data WITHOUT one — so deleting the reservation branch reddened that second
>   test on the "pages" key's own absence, never on the `got1 != got2` hijack comparison its own
>   failure message describes (D-000.13: it failed, but not for the reason it names). The finisher
>   changed both data variants to carry a top-level "page" key with *different* values, so the same
>   mutation now reddens on the actual hijack comparison: `a top-level "page" data key changed
>   {{page}}'s rendering: "HIJACKED" vs "SOMETHING ELSE"` — reconfirmed live (see the Findings
>   Resolutions section below for the transcript).

### Completion Notes

**All 30 ACs verified** (AC1–AC30, including the sub-lettered AC3a/AC4a/AC4b/AC4c/AC5a):

- AC1–AC7, AC21: `Decimal`, the shared exported `SplitJSONNumber`, and the two-decode-trees shape — `internal/bind/decimal.go`, `value.go`; `internal/template/decimal.go`. No `internal/expr` package exists (DW-8 unaffected).
- AC3, AC3a: coefficient significant-digit bound and the all-zeros edge case — `TestNewDecimalCoefficientBound`, `TestNewDecimalAllZeros`.
- AC4, AC4a, AC4b, AC4c: the DoS fix, its two required checks, the layered bounds, and the retained corpus fixture — see Delivery Log above; this is the story's most load-bearing work and was done first, as instructed.
- AC5, AC5a: no `float64` under the module — `TestNoFloat64UnderModule`, second caller, existing caller kept.
- AC6: one splitter — `SplitJSONNumber`, exported, reused by `internal/bind`, not duplicated.
- AC8–AC11: AD-14's three cases and the retained fixture triple — `TestAD14Triple`.
- AC12–AC14: structural separation (`render_entry.go`), no second checker, AD-12 locale doc comment.
- AC15–AC21: grammar, loud rejection naming Epic 3, `{{page}}`/`{{pages}}` reservation and its retained fixture (already present in `worked-example.json:56`, confirmed still renders correctly since it is never passed through `Render`, only round-tripped), AC20's field-scope fence proved at render time (`TestRenderScopeFenceIgnoresTableBind`), DW-8 verified matching.
- AC22–AC24: signature, `Data` as a defined type inserted in final position, bytes not a decoded value (`internal/bind.DecodeData` owns the `UseNumber` decode).
- AC25, AC25a: every new check/red-proof answers all three D-000.9/D-000.13 questions — see above.
- AC26–AC28: every gate `-count=1`; no pipe ever used for byte/hash/panic-stack verification (`rtk proxy … > file` throughout); cross-target matrix explicitly not run, recorded as deferred (D-000.4).
- AC29: `internal/bind` inherits AD-1's ban and D-1.3.1's exact `_test.go` exemption — verified via the existing `ScanForbiddenImports`/`ScanMapRange` production scans, which cover it automatically (no separate assertion needed; the existing scan's target already subsumes the new directory).
- AC30: the D-000.10 disposition table is filled, one line per enumerated ruling ID (34 rows, matching the table's actual row count), no gaps.

**Test count**: `folio-go` module went from 185 (with subtests, per the story's own reference figure) to a larger count after adding `internal/bind` (a new, seventh package) plus new tests in `internal/template`, `folio-go` (root), and `folio-go/internal` (arch). Every package reports `ok` under `go test -count=1 ./...`. `lint`'s 4 packages all report `ok`, including under `GOPROXY=off`. `go build -tags=matrix ./...` and `go vet -tags=matrix ./...` both succeed silently. `gofmt -l .` is empty in `folio-go/` and `lint/`.

**DoS fix — before/after evidence:**

| Literal | Before (measured this session) | After (measured this session) |
|---|---|---|
| `1e99999999999999999999` | hung; killed by `-timeout` at 15.436s, ~30-deep `math/big.karatsubaSqr`/`nat.sqr` recursion into `nat.expNN`, rooted at `decodePoints` (`decimal.go:48`) | errors in well under a second — "exponent … overflows during parsing" |
| `1e9223372036854775808` | (same mechanism; not separately re-measured to a full 15s this session — M-1's original 15s figure is reused per D-1.2.5, since re-deriving it would mean deliberately re-hanging the process for no new information) | errors — "exponent … overflows during parsing" |
| `1e2000000` (Story 1.4's own control) | 0.07s, correct error via `decodePoints`' millipoint check | still errors, now at the shared splitter (magnitude 2,000,000 > `MaxSplitExponentMagnitude` 1,000,000) |
| Retained corpus fixture, through `folio.LoadTemplate` | (pre-existing baseline hang was M-1's own 20s measurement, not independently re-reproduced against the corpus fixture this session — the unit-level reproduction above is the one this session captured firsthand) | 0.197s, located error naming element `e1` |

> **Corrections (QA Findings 6 and 7, this story's review; resolved by the finisher):**
>
> - **Finding 7.** Row 3's "`MaxSplitExponentMagnitude` 1,000,000" is now **100,000** — the
>   constant was 10x the widest actual consumer bound, letting ~900,000 exponent values no consumer
>   can represent still reach `decodePoints`' `big.Int.Exp` (measured by the reviewer: 33.0ms at the
>   old bound, 29µs one past the corrected bound). Not a hang or a regression (input-proportional,
>   loader aborts on the first bad element) — tightened to match AC4b's letter.
> - **Finding 6.** This table's DoS-fix evidence only exercised **Fix 2** (the magnitude bound) —
>   the retained corpus fixture's literal (`1e99999999999999999999`) wraps to a **positive** int64
>   (`7766279631452241919`), which Fix 2 alone still catches. The literal that actually requires
>   **Fix 1** (the during-accumulation overflow check) is `1e9223372036854775808` — one past int64
>   max — which wraps to `math.MinInt64`, a **negative** value that Fix 2's `n > MaxSplitExponentMagnitude`
>   cannot catch. A second retained corpus fixture,
>   `folio-go/testdata/template/malformed/huge-exponent-negative-wrap.folio`
>   (`TestLoadTemplateRejectsNegativeWrapExponentQuicklyToo`), now proves Fix 1 through the public
>   loader the same way this fixture proves Fix 2 — red-proofed live by removing Fix 1 alone: the
>   original fixture's test stayed **green** in 0.407s (confirming the gap), the new fixture's test
>   **hung; killed by `-timeout` at 15.190s**, ~30 `math/big.karatsubaSqr`/`nat.sqr` frames rooted at
>   `decodePoints` via `ParseDocument` — reproducing M-1's mechanism exactly, then restored.

**Confirmed**: the fixing commit hash in Story 1.4's `## Defect found after done` section is **left as a reference to "this story's Delivery Log"** — no literal hash field exists in 1.4's file to fill in; the finisher fills in the actual commit hash wherever this story's own record needs it once committed. Story 1.4's file was not modified by this developer (verified: not in the File List below).

> **Finisher's note.** Per the story's own instruction (D-1.6.6, Task 2b) and the finishing brief,
> the finisher appends the actual fixing commit hash to Story 1.4's file once this story is
> committed — a small, purely additive append below its existing `## Defect found after done`
> section, touching nothing above that heading. See Story 1.4's file for the appended line and this
> story's commit SHA in the finisher's own report.

### File List

**New:**
- `folio-go/internal/bind/decimal.go`
- `folio-go/internal/bind/decimal_test.go`
- `folio-go/internal/bind/value.go`
- `folio-go/internal/bind/text.go`
- `folio-go/internal/bind/text_test.go`
- `folio-go/render_entry.go`
- `folio-go/render_bind_test.go`
- `folio-go/testdata/template/malformed/huge-exponent.folio`

**Modified:**
- `folio-go/internal/template/decimal.go` (DoS fix; `splitJSONNumber` exported as `SplitJSONNumber`)
- `folio-go/internal/template/decimal_test.go` (new DoS-fix and exponent-bound tests)
- `folio-go/folio.go` (`Render` and `errNilTemplate` moved out to `render_entry.go`)
- `folio-go/render.go` (`collectTextRuns`/`renderDocument` take a `bind.Value` and call `bind.BindText`)
- `folio-go/template_test.go` (new corpus-fixture test; `Render` call site updated; stale doc comment corrected)
- `folio-go/fixture_test.go` (`Render` call site updated)
- `folio-go/render_test.go` (`Render` call sites updated, ×9)
- `folio-go/internal/arch_test.go` (`walkGoFiles` extended to skip dot-directories by category; new `TestNoFloat64UnderModule`)
- `lint/internal/rules/forbiddenimports_test.go` (new `TestRenderEntryFileHasNoForbiddenImports`)
- `_bmad-output/implementation-artifacts/1-6-bind-scalar-json-values-into-text.md` (this file: disposition table, Dev Agent Record, Status)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (status → `review`)

**Not modified** (verified): `_bmad-output/implementation-artifacts/1-4-load-validate-and-round-trip-a-folio-template.md` — the finisher's fixing-commit-hash field is elsewhere; nothing in 1.4's file needed a developer edit.

> ### File List additions — finisher's QA-finding fixes
>
> **New:**
> - `folio-go/internal/bind/value_test.go` (DecodeData trailing-content tests, Finding 4)
> - `folio-go/testdata/template/malformed/huge-exponent-negative-wrap.folio` (Finding 6)
>
> **Modified:**
> - `lint/internal/rules/forbiddenimports_test.go` — `TestRenderEntryFileHasNoForbiddenImports`
>   now targets the file(s) that declare `Render`/`RenderTo`, found by AST (`findRenderDeclaringFiles`),
>   instead of a hard-coded path (Finding 1, Blocker)
> - `lint/internal/rules/walk.go` — `walkGoFiles` falls back to the file's base name instead of `"."`
>   when the scan root is a single file (Finding 13, Minor)
> - `folio-go/internal/bind/decimal.go` — exponent bound now checked against the CONSTRUCTED
>   `Decimal.Exponent`, not the literal's raw `exp` (Finding 3, Major); doc comment corrected
> - `folio-go/internal/bind/text.go` — `BindText`'s `Present`/`KindNumber` arm now validates via
>   `AsDecimal` before the wrong-kind rejection, wiring `NewDecimal`/`AsDecimal` onto a real
>   production path (Finding 2, Major); `TestBindTextPageReservationIsRedProofable` redesigned so its
>   mutation reddens on the actual hijack comparison, not a shadowed absent-path error (Nit 15)
> - `folio-go/internal/bind/decimal_test.go` — negative-side bound tests (Finding 9, Minor);
>   constructed-exponent-bound tests, both no-`e`-notation and stacked frac+exponent forms (Finding 3)
> - `folio-go/internal/bind/text_test.go` — located-error test for an out-of-bound number in report
>   data (Finding 2); wrong-kind-still-rejected control test; page-reservation red-proof redesign (Nit 15)
> - `folio-go/internal/bind/value.go` — `DecodeData` rejects trailing content after the single
>   top-level JSON value (Finding 4, Major)
> - `folio-go/internal/template/decimal.go` — `MaxSplitExponentMagnitude` corrected from 1,000,000
>   to 100,000, the actual wider consumer bound (Finding 7, Minor); doc comment corrected
> - `folio-go/internal/template/decimal_test.go` — `TestSplitJSONNumberExponentBound` updated to the
>   corrected bound, both polarities (Finding 7, Finding 9); Nit 14 documented in place, not removed
> - `folio-go/render.go` — `resolveFace` (font-chain validation) now runs before the AC9
>   empty-bound-text short-circuit, not after (Finding 5, Major)
> - `folio-go/render_bind_test.go` — font-chain-validates-regardless-of-empty-binding test
>   (Finding 5)
> - `folio-go/render_entry.go` — blank line inserted before `package folio` so the file's doc
>   comment is no longer treated as a second package doc comment (Finding 8, Minor)
> - `folio-go/template_test.go` — new `TestLoadTemplateRejectsNegativeWrapExponentQuicklyToo`
>   (Finding 6, Minor)
> - `_bmad-output/implementation-artifacts/1-6-bind-scalar-json-values-into-text.md` — Findings
>   Resolutions section, corrected ruling disposition table, banner/M-7/Delivery-Log corrections,
>   rewritten plain-terms opener, Status → `done`
> - `_bmad-output/implementation-artifacts/1-4-load-validate-and-round-trip-a-folio-template.md` —
>   appended the fixing commit hash below the existing `## Defect found after done` section
> - `_bmad-output/implementation-artifacts/sprint-status.yaml` — status → `done`

### Change Log

- Fixed a shipped denial-of-service in `internal/template.parseDecimalExponent`: an absurd exponent literal wrapped the accumulator silently and reached `math/big.Int.Exp`'s repeated squaring, hanging indefinitely — reachable from untrusted input through the public `folio.LoadTemplate`. Two independent checks now reject it quickly (D-1.6.6).
- Added `internal/bind`: `Decimal` (AD-23's exact scaled-integer decimal), a generic report-data value tree, and the `{{…}}` binding grammar/resolver, scoped to text-element `value` interpolation.
- Changed `Render`'s signature to `Render(t *Template, d Data, f FontSet) ([]byte, error)`; added the `Data` defined type; split `folio.go` so the file declaring `Render` imports none of os/time/net/math/rand.
- Added a second production caller of the existing `no-float64` checker, covering the whole module (not only `internal/`).

> **Finisher's additions to the Change Log (post-review fixes):**
> - The AC12 guard now locates the file declaring `Render`/`RenderTo` by parsing the AST, rather
>   than trusting a hard-coded filename — closes the Blocker.
> - `Decimal`/`AsDecimal` are now reachable from a real production path (`BindText`), so an
>   out-of-bound number literal in report data produces the located error AC3/AC4 require.
> - `Decimal`'s exponent bound now applies to the value it actually constructs, closing two classes
>   of literal that previously bypassed it entirely.
> - `DecodeData` now rejects trailing content after the report data's single top-level JSON value.
> - Font-chain validation for a text element no longer depends on whether its bound text happens to
>   be empty.
> - The shared splitter's exponent-magnitude bound is now the ruled value (100,000), and a second
>   corpus fixture proves the DoS fix's other mechanism through the public loader.
> - `render_entry.go`'s ruling/story-number prose no longer leaks into the library's public godoc.

---

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-23
- **Baseline:** `5a1cf7d` (working tree, uncommitted; working-tree decision log + `deferred-work.md` read directly, not diffed)
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 1
- **Majors:** 4
- **Minors:** 8
- **Nits:** 5

### Gates re-run independently (AC26, AC27, D-000.11, D-000.12)

Every command below was run through `rtk proxy … > file` with `-count=1`; exit codes captured raw, never through a pipe.

| Gate | Result |
|---|---|
| `folio-go` `go build ./...` | exit 0 |
| `folio-go` `go vet ./...` | exit 0 |
| `folio-go` `go test -count=1 ./...` | exit 0 — **7 packages**, all `ok` (matches the reference figure) |
| `lint` `go test -count=1 ./...` | exit 0 — 4 packages, all `ok` |
| `hashmatrix` `go test -count=1 ./...` | exit 0 — `probe [no test files]` |
| `folio-go` `go build -tags=matrix ./...` | exit 0 (AC28: matrix itself correctly NOT run) |
| `folio-go` `go vet -tags=matrix ./...` | exit 0 |
| `gofmt -l` in `folio-go/` and `lint/` | empty |

**Tree hygiene confirmed.** `git status --porcelain` was captured before and after the review and is byte-identical; every mutation below was `cp`-restored, never `git checkout`. Per-file diffstat vs `5a1cf7d` after the review matches the pre-review diffstat exactly (`folio.go` −30, `decimal.go` +72/−4, `arch_test.go` +73/−5, …). No reviewer artifact remains.

---

## Independent ruling disposition table (D-000.10 mirror)

Derived from the code first, then compared. **Rows marked ⚠️ disagree with the developer's disposition.**

| # | Ruling | Reviewer disposition | vs dev |
|---|---|---|---|
| 1 | **D-1.6.1** | ⚠️ **applied-with-deviation** — `Decimal{int64, int}` over the one shared splitter is correct, and AC2/AC3/AC3a are all genuinely proved. But (i) `NewDecimal` has **zero production callers** and `AsDecimal` has **zero callers of any kind**, so AD-23's "converts each literal to an exact scaled-integer decimal" happens on no shipped path (Finding 2); (ii) AC3/AC4's mandated **located error naming the data path and the element id** is not implemented — `NewDecimal`'s errors name only the literal, and no wrapper adds either; (iii) Decimal's own exponent bound is checked on the wrong quantity (Finding 3). | ⚠️ |
| 2 | **D-1.6.2** | **applied-as-stated** (independently re-proved). One template, one path confirmed by reading `text_test.go`. I constructed three separate `Presence` collapses myself; **all three redden row 2** naming the collapsed reason. Correction to the dev record in Finding 12. | ✅ |
| 3 | **D-1.6.3** | ⚠️ **NOT applied as stated — the binding property is unasserted.** Measured: moving `func Render` back into `folio.go` (which imports `os`) leaves the module building **and `TestRenderEntryFileHasNoForbiddenImports` PASSING**. See Finding 1. | ⚠️ |
| 4 | **D-1.6.4** | ⚠️ **applied-with-deviation** — signature, insertion position, `type Data []byte` as a defined type, and two-trees-one-splitter are all correct and verified. AC24's "`d` must be syntactically valid JSON" is **not enforced**: trailing garbage after the first JSON value is silently discarded (Finding 4). | ⚠️ |
| 5 | **D-1.6.5** | **applied-as-stated** — grammar, Epic-3 rejection message, and the reservation all verified, including a render-level AC19 probe I constructed (identical 11,227-byte PDFs with and without a top-level `page`/`pages` key). Nit 15. | ✅ |
| 6 | **D-1.6.6** | ⚠️ **applied-with-deviation** — both fixes present, and I confirmed **each is independently load-bearing** by mutation. But the retained corpus fixture never exercises Fix 1 (Finding 5), the splitter's bound is not "the WIDER of the two" (Finding 6), and neither bound's negative side is tested (Minor 9). | ⚠️ |
| 7 | **D-1.6.7** | **applied-as-stated** (independently re-proved). Mutation re-applied: `TestNoFloat64UnderModule` red naming `render_entry.go:78: identifier float64` **and** `untyped floating-point literal 1.5`; `TestNoFloat64UnderInternal` stayed **green**. Whole module, `_test.go` included (no `_test.go` filter in `walkGoFiles`), dot-dirs skipped by category (`strings.HasPrefix(d.Name(), ".")`, no name list), existing caller unmodified. Nit 11. | ✅ |
| 8 | **D-1.6.8** | **applied-as-stated** — `BindText` has exactly one call site (grep-verified, whole module), inside `collectTextRuns`, which skips every non-text element. The fence holds by construction. `TestRenderScopeFenceIgnoresTableBind` is in the default gate. I confirmed the golden fixture is never rendered (a `Render` on it fails on an unrelated missing `style.fontFamily`), so a render-time reproduction was the right substitute. | ✅ |
| 9 | **DW-8** | **applied-as-stated** — `folio-go/internal/` contains `bind fontset geom pdf template`; **no `internal/expr`**. `absence-expr-package` tripwire live in `lint/internal/rules/absences.go`; `internal/bind` is not in `absenceChecks`. No import exists that would become a cycle. | ✅ |
| 10 | Story 1.4 `## Defect found after done` | **not-reached-by-design** — verified pure append (`@@ -1741,3 +1741,59 @@`, +56/−0); nothing above the heading touched. Commit-hash field deliberately unfilled: **not a defect**, per instruction. | ✅ |
| 11 | D-000.4 | **applied-as-stated** — matrix not run and stated as deferred; `-tags=matrix` build and vet both exit 0. | ✅ |
| 12 | D-000.9 | ⚠️ **applied-with-deviation** — `TestNoFloat64UnderModule`'s witness reads the scanner's own stats and is sound. `TestRenderEntryFileHasNoForbiddenImports`'s witness (`FilesSeen != 1`) proves a file was parsed but **not that it declares `Render`** — the Finding 1 mutation passes it while the property is violated. This is D-000.9's own failure mode, one level up. | ⚠️ |
| 13 | D-000.9 (extended) | **applied-as-stated** — I reproduced every red-proof the Delivery Log claims, plus two the developer did not run (Findings 1 and 5). | ✅ |
| 14 | D-000.10 | ⚠️ **applied-with-gap** — 34 rows, none blank ✅, count independently confirmed. But **`D-1.3.5` has no row** although the dev's own Delivery Log and `value.go`'s comment cite it as governing the new `slices.Sorted(maps.Keys(...))` code (Minor 10). And this mirror disagrees on **6 of 34** rows. | ⚠️ |
| 15 | D-000.11 | **applied-as-stated** — every gate above re-run with `-count=1`. | ✅ |
| 16 | D-000.12 | **applied-as-stated** — I captured all raw output to files; `rtk proxy` never truncated a stack I relied on. | ✅ |
| 17 | **D-000.13** | ⚠️ **applied-with-deviation** — every control I re-applied was valid syntax with forbidden semantics ✅, and every red named rule content ✅. But AC12's red-proof only fails "for the reason it names" under the mutation the developer *chose*; under the mutation that tests AC12's **stated** property it does not fail at all (Finding 1). And the page-reservation red names `data path "page" is absent`, not the hijack its own message claims (Nit 15). | ⚠️ |
| 18 | D-1.1.b | ⚠️ **applied-with-deviation** — `numbers.go`'s precedent is a property of a **file's contents**; here the property is bound to a **filename** that need not hold `Render`. Same root cause as Finding 1. | ⚠️ |
| 19 | D-1.1.c | **applied-as-stated** — `type Data []byte` is a defined type, not an alias; the `Data`/`Params` swap will be a compile error. | ✅ |
| 20 | D-1.2.5 | **applied-as-stated** — the hang is recorded verbatim, never dressed as an assertion failure. I reproduced a hang of the same class myself (`panic: test timed out after 25s`, `FAIL … 25.891s`). | ✅ |
| 21 | D-1.2.6 | **not-reached-in-this-story** — agreed. | ✅ |
| 22 | D-1.3.1 | **applied-as-stated** — `internal/bind` non-test files import `fmt`, `math/big`, `strings`, `bytes`, `encoding/json`, `maps`, `slices`: none banned. `_test.go` files import only `testing`/`strings`. Covered automatically by the existing `internal/`-scoped production scan. | ✅ |
| 23 | D-1.3.3 (amended) | **applied-as-stated** — every control I applied compiled (`go build ./...` exit 0 under the `os`+`float64` mutation); no red came through a parse-error path. | ✅ |
| 24 | D-1.4.6 | **applied-as-stated** — `folio.go` keeps `os` for `LoadTemplate`; the package boundary is intact. | ✅ |
| 25 | D-1.4.8 (corrected) | **applied-as-stated** — `bind.Presence{Absent,Null,Present}` mirrors the seam and is genuinely three-state. | ✅ |
| 26 | D-1.4.15 | ⚠️ **applied-with-deviation** — `MaxSplitExponentMagnitude`'s comment does tag binding-vs-illustrative ✅, but two shipped doc comments state facts the code does not satisfy: `SplitJSONNumber`'s "deliberately the WIDER of this module's two known consumer bounds" (it is 10× wider than the wider) and `maxDecimalExponentMagnitude`'s "that it exists and is strictly narrower … is binding" (Finding 3). | ⚠️ |
| 27 | D-1.4.16 | **applied-as-stated** — production code now branches on `Null` vs absent; the exposure is closed. | ✅ |
| 28 | D-1.5.5 | **applied-as-stated** (mechanism) — `d` inserted between `t` and `f`, nothing reordered. But the story's own M-7 says **thirteen** call sites "across four test files" while naming three files and listing ten line numbers; the true baseline count is **11** (`render_test.go` 9, `template_test.go` 1, `fixture_test.go` 1). The dev used 11 without recording the discrepancy against a measured reference (Minor 11). | ✅ |
| 29 | AD-1 | **applied-as-stated** — `internal/bind` inherits the ban; verified through the existing production scan. | ✅ |
| 30 | AD-2 | **applied-as-stated** — zero `float64`/`float32` module-wide, `_test.go` included. | ✅ |
| 31 | AD-4 | **applied-as-stated** — no `page` namespace; reservation verified at unit and render level. | ✅ |
| 32 | AD-12 | **applied-as-stated** (content) — the clarification is present and correct. Placement nit: it lands as a **second package doc comment** and now appears in `go doc .` (Minor 8). | ✅ |
| 33 | AD-14 | **applied-as-stated** — three cases implemented and non-vacuously proved. | ✅ |
| 34 | AD-23 | ⚠️ **applied-with-deviation** — `UseNumber` ✅, no `float64` ✅. But "internal/bind converts each [literal] to an exact scaled-integer decimal" occurs on **no shipped path** (Finding 2). | ⚠️ |

**Disagreements: 8 of 34** (rows 1, 3, 4, 6, 12, 14, 17, 18, 26, 34 carry ⚠️; rows 2 and 28 agree on verdict but correct the record). An `applied-as-stated` that a mutation falsifies is exactly what this mirror exists to catch — row 3 is that case.

---

## Findings

### Finding 1: AC12/D-1.6.3's guard asserts a filename, not "the file declaring `Render`" — the property is defeated by moving the function

- **Severity**: **Blocker**
- **Category**: AC Conformance / Tests
- **Location**: `lint/internal/rules/forbiddenimports_test.go:191–209` (`TestRenderEntryFileHasNoForbiddenImports`); `folio-go/render_entry.go:71`
- **Observation**: The guard hard-codes the path `folio-go/render_entry.go` and asserts that *that file* has no forbidden imports. It never checks that the file declares `Render`. I applied the mutation the AC actually describes — moved `func Render` out of `render_entry.go` and back into `folio.go` (which imports `os`), leaving `render_entry.go` in place holding `type Data` and the doc comment:
  - `go build ./...` → **exit 0**
  - `go test -count=1 -run TestRenderEntryFileHasNoForbiddenImports ./internal/rules/` → **PASS, exit 0**

  The vacuity guard `stats.FilesSeen != 1` passes because `render_entry.go` still exists and still parses. It would pass even if the file were reduced to a package clause.
- **Impact**: D-1.6.3's binding property is **unasserted**. Its own justification — *"fails the moment someone adds a convenience `os.Getenv` to the render entry point"* — is false once the entry point moves, which is precisely the drift a structural fence exists to stop, and precisely what Story 1.7 will do when it adds `RenderTo`. AC12 tags the **property** binding and the **filename** illustrative; the shipped assertion binds only the illustrative half. AC13's "no second checker" is satisfied, but not at the cost this trade was supposed to buy. This also falsifies rows D-1.6.3, D-1.1.b, D-000.9 and D-000.13 of the disposition table.
- **Suggested Resolution**: Make the target derive from the property, not the other way round: locate the file that declares `Render` (and, from 1.7, `RenderTo`) — e.g. parse the module-root package, find the `*ast.FuncDecl` named `Render`, take `fset.Position(decl.Pos()).Filename`, and scan **that** — then assert the render-declaring file was found (a zero-declarations run is a failure) before scanning it. Keep reusing `ScanForbiddenImports`; only the target selection changes.
- **Related AC**: AC12, AC13, AC25 (D-000.9), AC25a (D-000.13)

### Finding 2: `Decimal` is unwired — `NewDecimal` has zero production callers, `AsDecimal` has zero callers of any kind

- **Severity**: **Major**
- **Category**: AC Conformance
- **Location**: `folio-go/internal/bind/decimal.go:71` (`NewDecimal`); `folio-go/internal/bind/value.go:62` (`AsDecimal`); `folio-go/internal/bind/text.go:88–101` (`BindText`'s `Present` arm)
- **Observation**: A whole-module grep for `NewDecimal|AsDecimal` returns hits only inside `internal/bind` itself and its own `decimal_test.go`. `AsDecimal` has **no caller at all**, not even a test. `BindText` accepts only `KindString` and rejects `KindNumber` as wrong-kind (AC10), so no number literal ever reaches `Decimal` from `Render`. Consequently AC3's and AC4's mandated **"located error naming the data path and the element id"** exists nowhere: `NewDecimal` returns `bind: value "…": coefficient has N significant digits…`, naming neither the path nor the element, and no wrapper adds them.
- **Impact**: AD-23's Rule — *"`internal/bind` converts each [literal] to an exact scaled-integer decimal"* — describes behaviour the shipped binary never performs. The story's headline money-path deliverable is reachable only from its own unit tests, so its bounds are never exercised against real report data and its error contract (located, path + element id) is both unimplemented and unassertable. `AsDecimal` is a half-wired symbol: the one function that would connect `Value` to `Decimal` is dead. The disposition table records D-1.6.1 and AD-23 as `applied-as-stated`, which overstates what ships.
- **Suggested Resolution**: Either (a) record honestly that `Decimal` is deliberately ahead of its consumer (a later story formats numbers into text) and downgrade the D-1.6.1/AD-23 dispositions to name the unwired half, adding a DW entry naming the owner who will wire it and add the located-error wrapper; or (b) wire it now by having `BindText` carry `elementID` and the joined path into a located wrapper around `NewDecimal`, and add a test asserting a coefficient-overflow literal in *report data* produces an error naming both. Do not leave the table claiming the located error ships.
- **Related AC**: AC1, AC3, AC4, AC6, AD-23

### Finding 3: `Decimal`'s exponent bound is checked on the literal's exponent, not on the `Decimal.Exponent` it constructs

- **Severity**: **Major**
- **Category**: Correctness
- **Location**: `folio-go/internal/bind/decimal.go:78–96`
- **Observation**: The guard is `if exp > maxDecimalExponentMagnitude || exp < -maxDecimalExponentMagnitude` — applied to the *literal's* exponent. Ten lines later the constructed value is `exponent := exp - len(fracPart)`, where `len(fracPart)` is unbounded. Measured (probe, since deleted):

  | Literal | Result | `maxDecimalExponentMagnitude` |
  |---|---|---|
  | `0.` + 199 999 zeros + `1` (no `e` at all) | `{Coefficient:1 Exponent:-200000}`, **err = nil** | 100 000 |
  | `0.` + 999 999 zeros + `1e-100000` | `{Coefficient:1 Exponent:-1100000}`, **err = nil** | 100 000 |

  A literal with **no exponent notation at all** never reaches the check, and one at the bound stacks its `fracPart` on top of it.
- **Impact**: The invariant the constant's own doc comment states — *"that it exists and is **strictly narrower than the splitter's bound** is binding"* — is false for the value actually constructed: `-1 100 000` exceeds both 100 000 and `MaxSplitExponentMagnitude` (1 000 000). The comment's stated purpose is to stop *"a future story that expands a Decimal's digits at render time"* from constructing an enormous expansion; that hazard survives intact, and the next story will read the comment and believe it is protected. This also breaks AC4b's layering claim ("`internal/bind` applies `Decimal`'s [tighter check]") — it applies it to a different number.
- **Suggested Resolution**: Move the bound check after `exponent := exp - len(fracPart)` and test the constructed value, or check both. Add tests at the constructed bound from a literal with **no `e`** (`0.` + N zeros + `1`) and from the stacked frac+exponent form, since neither current test can reach this path.
- **Related AC**: AC4b, AC1, D-1.4.15

### Finding 4: `DecodeData` silently accepts trailing garbage after the first JSON value — a malformed payload renders successfully

- **Severity**: **Major**
- **Category**: Correctness / Security
- **Location**: `folio-go/internal/bind/value.go:161–169` (`DecodeData`); contract stated at `folio-go/render_entry.go:59` (*"d must be syntactically valid JSON (AC24)"*)
- **Observation**: `DecodeData` performs a single `dec.Decode(&v)` and returns. `json.Decoder` stops at the end of the first value; everything after it is discarded without inspection. Measured through the **public `Render`**:

  | `Data` | Result |
  |---|---|
  | `{"customer":{"name":"Ada"}} THIS IS NOT JSON` | **err = nil**, renders |
  | `{"customer":{"name":"Ada"}}{"customer":{"name":"Bob"}}` | **err = nil**, renders `Ada`; second document silently ignored |
  | `` (empty) | err = `bind: invalid JSON report data: EOF` |

- **Impact**: `Render`'s own documented precondition is unenforced. A caller who concatenates, truncates mid-stream, or appends to a report payload gets a **silently wrong document** rather than an error — in a product whose acceptance fixture is a bank statement, and on the one input surface the story classifies as *"untrusted caller input"* (AC4's rationale). The two-document case is the worse half: the render is plausible and wrong.
- **Suggested Resolution**: After the successful `Decode`, confirm the stream is exhausted — e.g. `if _, err := dec.Token(); err != io.EOF { return Value{}, fmt.Errorf("bind: trailing data after the JSON report data") }` — and add a test for trailing garbage and for two concatenated documents. Consider a clearer message for empty `Data` than a bare `EOF`.
- **Related AC**: AC24, AC22, AD-23

### Finding 5: `boundText == ""` short-circuits before `resolveFace`, so report data suppresses font-chain validation

- **Severity**: **Major**
- **Category**: Correctness
- **Location**: `folio-go/render.go:129–140`
- **Observation**: The new branch
  ```go
  if boundText == "" {
      continue
  }
  face, err := resolveFace(doc, el, fs)
  ```
  is inserted **before** `resolveFace`, which is the only thing that validates `style.fontFamily` against the document's chains and the caller's `FontSet`. The baseline had no such data-dependent skip. Measured on one template with an unresolvable chain (`"fontFamily": "nosuchchain"`), varying only the data:

  | Element value / data | `Render` result |
  |---|---|
  | `Hello` (no placeholder) | **error**: `element e1: style.fontFamily "nosuchchain" names a chain with no entries…` |
  | `{{customer.name}}` + `{"customer":{"name":null}}` | **err = nil** — renders |
  | `{{customer.name}}` + `{"customer":{"name":""}}` | **err = nil** — renders |
  | `{{customer.name}}` + `{"customer":{"name":"Ada"}}` | **error** (as above) |

- **Impact**: A broken template now passes or fails depending on which report it is handed. The located error Story 1.5 shipped for an unresolvable font chain is suppressed for whichever elements happen to bind to `null` or `""` on a given run — so CI can be green against one report and the fault surfaces in production against another, which is exactly the failure mode "caught in CI rather than in production" is meant to prevent. It also makes the embedded subset data-dependent in a way nothing asserts. AC9 requires only *"renders as empty, and is not an error"*; it does not license skipping the element's own validation.
- **Suggested Resolution**: Resolve the face (and any other per-element validation) **before** the empty-text short-circuit, and skip only the run emission. Add a test pinning the table above: same template, unresolvable chain, null data → still a located error.
- **Related AC**: AC9, AC11; Story 1.5 AC2/AC4

### Finding 6: the retained corpus fixture is caught by Fix 2 alone — it never exercises Fix 1

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/testdata/template/malformed/huge-exponent.folio:6`; `folio-go/template_test.go:107–152` (`TestLoadTemplateRejectsHugeExponentQuickly`)
- **Observation**: I removed **Fix 1** only (the during-accumulation overflow check) and re-ran with `-count=1`:
  ```
  FAIL  github.com/panitw/folio/folio-go/internal/template  25.891s   panic: test timed out after 25s
  ok    github.com/panitw/folio/folio-go/internal/bind        1.275s
  ok    github.com/panitw/folio/folio-go                      1.235s   <-- corpus fixture GREEN
  ```
  The corpus fixture's literal `1e99999999999999999999` wraps to a **positive** `7766279631452241919`, which Fix 2's `n > MaxSplitExponentMagnitude` still catches. The literal that defeats Fix 2 is `1e9223372036854775808` (wraps to `math.MinInt64`, so `n > 1_000_000` is false) — and that literal exists **only in a unit test** (`decimal_test.go`'s `TestParseDecimalExponentDoesNotHang`), never in a `.folio` through the public loader. Removing **Fix 2** only reddens `TestSplitJSONNumberExponentBound` correctly, so that fix is properly covered.
- **Impact**: AC4c's obligation is *"the fixture proves Story 1.4's promise now holds **through the public path**"*. It proves it for the easier of the two input classes. The class that requires Fix 1 — the one M-1 measured as producing `exp=-9223372036854775808` — has no public-loader coverage at all, which is the same shape of gap as the original defect (a bound tested only at values a later check can still reject).
- **Suggested Resolution**: Change the fixture's `x` to `1e9223372036854775808` (or add a second `.folio` fixture with it) so the corpus covers the negative-wrap class, and keep the existing literal in the unit test. Both fixes then have public-path coverage.
- **Related AC**: AC4a, AC4c, AC25 (D-000.9 extended)

### Finding 7: the splitter's bound is 10× the widest consumer bound, and ~900 000 exponent values still reach `big.Int.Exp` — measured at 33 ms

- **Severity**: Minor
- **Category**: Correctness / Performance
- **Location**: `folio-go/internal/template/decimal.go:113–128` (`MaxSplitExponentMagnitude`), `:46–52` (`decodePoints`' `big.Int.Exp`)
- **Observation**: AC4b: *"The shared splitter's bound must be the **WIDER of the two**, so it never refuses what a consumer could legally represent."* The two consumer bounds are `internal/bind.maxDecimalExponentMagnitude` = **100 000** and `decodePoints`' effective millipoint range (~19 significant decimal places). The wider is 100 000; the splitter uses **1 000 000**. Measured through the public `ParseTemplate`:

  | Literal | Time | Outcome |
  |---|---|---|
  | `1e999999` | **21.0 ms** | error via `decodePoints`' millipoint check |
  | `1e1000000` (exactly at the bound) | **33.0 ms** | error via `decodePoints`' millipoint check |
  | `1e1000001` (one past) | **29 µs** | error at the splitter, before any `big.Int` work |
  | `1e-1000000` | **18.0 ms** | error via "more than three decimal places" |

  So a 9-byte literal buys ~33 ms of `big.Int.Exp` repeated squaring. I also confirmed this is **bounded, not a DoS**: the loader aborts on the first bad element (40 such literals in one document still cost 29 ms total), and deep data nesting, million-digit coefficients and million-digit fractional parts are all input-proportional (455 ms for a 1 MB literal).
- **Impact**: Not a hang, and not a regression — but AC4a-ii's *"rejected BEFORE any `big.Int` scaling is attempted"* is only true **above** the bound, and the extra decade between 100 000 and 1 000 000 is pure attack surface no consumer can use, costing ~100× the work at the worst case. I could **not** find any literal `Decimal` should legally accept that the splitter refuses (Decimal's own bound is strictly tighter), so the binding half of AC4b holds.
- **Suggested Resolution**: Set `MaxSplitExponentMagnitude` to the actual wider consumer bound (100 000) as AC4b's letter requires, and move `TestSplitJSONNumberExponentBound`'s at-the-bound/one-past pair with it. If the decade of headroom is deliberate, say so in the doc comment instead of claiming the constant is "the WIDER of this module's two known consumer bounds", which it is not.
- **Related AC**: AC4a, AC4b, D-1.4.15

### Finding 8: `render_entry.go`'s file comment becomes a second **package** doc comment — ruling IDs and story numbers now appear in public godoc

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/render_entry.go:1–22`
- **Observation**: The comment block sits directly above `package folio` with no blank line, so Go treats it as a package comment. `go doc .` now renders, in the library's **public** documentation: *"This file declares Render (and, at Story 1.7, RenderTo) — the world-reading fence (D-1.6.3, AC12) … the same shape D-1.1.b used for internal/pdf/numbers.go"*, concatenated after `folio.go`'s real package comment.
- **Impact**: Internal ruling identifiers, story numbers and review-record prose are published to consumers of the library. AC14 requires the locale clarification to live "in the doc comment of whichever file AC12 creates" — that is satisfied — but as a *package* comment it also drags the implementation record along with it.
- **Suggested Resolution**: Insert a blank line between the comment block and `package folio` so it is a file comment rather than a package comment, keeping AC14's locale paragraph where it is. Alternatively split: keep the AD-12 locale paragraph as the package-visible text and demote the D-1.6.3/D-1.1.b provenance to a non-doc comment.
- **Related AC**: AC12, AC14, AD-12

### Finding 9: neither exponent bound is tested on its negative side

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/template/decimal_test.go:82–90`; `folio-go/internal/bind/decimal_test.go:77–84`
- **Observation**: `TestSplitJSONNumberExponentBound` tests `parseDecimalExponent("1000000")` and `("1000001")`; `TestNewDecimalExponentBound` tests `NewDecimal("1e100000")` and `("1e100001")`. Neither tests `-1000001` / `1e-100001`. In the splitter the bound is applied before `if neg { n = -n }`, so the symmetry is real but unpinned; in `Decimal` the check is a two-sided comparison whose lower arm has no test.
- **Impact**: AC4b makes *"tests at the bound and one past it"* binding. Half of each bound is unexercised, so a future edit that reorders the negation past the bound check (splitter) or drops the `exp < -max` arm (Decimal) goes unnoticed.
- **Suggested Resolution**: Add `parseDecimalExponent("-1000000")` legal / `("-1000001")` error, and `NewDecimal("1e-100000")` legal / `("1e-100001")` error.
- **Related AC**: AC4b

### Finding 10: `D-1.3.5` governs new shipped code but has no disposition row

- **Severity**: Minor
- **Category**: Convention
- **Location**: `folio-go/internal/bind/value.go:186–190`; story `§ Ruling disposition table`; story header lines 9–12
- **Observation**: `value.go`'s comment cites *"D-1.3.5/NFR1.d bans ranging a map anywhere under internal/; the escape hatch (sorted keys, then index) is used…"*, and the Delivery Log repeats the citation. `D-1.3.5` appears in neither the story's "Rulings that govern" enumeration nor the 34-row table.
- **Impact**: AC30 requires a disposition for every ruling the story enumerates, and D-000.10's purpose is that no governing ruling passes unreviewed. A ruling the newly shipped code explicitly invokes is outside the table — the enumeration, not the developer's diligence, is what failed, but the gap is real.
- **Suggested Resolution**: Add a `D-1.3.5` row (`applied-as-stated` — the escape hatch is used correctly and `ScanMapRange` covers `internal/bind` automatically), and note the enumeration gap for the story creator.
- **Related AC**: AC30, D-000.10

### Finding 11: M-7's "thirteen call sites" is wrong; the actual count is 11, and the discrepancy is not recorded

- **Severity**: Minor
- **Category**: Convention
- **Location**: story `§ M-7` and Task 9; `folio-go/render_test.go`, `template_test.go`, `fixture_test.go`
- **Observation**: M-7 states *"Thirteen `Render(` call sites exist across **four** test files"*, then names **three** files and lists **ten** line numbers. Measured at baseline `5a1cf7d`: `render_test.go` 9, `template_test.go` 1, `fixture_test.go` 1 = **11**. The Delivery Log correctly says "11 call sites … updated" but does not flag that the binding-tagged Task 9 instruction said thirteen.
- **Impact**: Small, but M-7 is a *measured* reference the story treats as authoritative, and D-1.4.15's discipline is that a departure from a stated mechanism is surfaced, not silently corrected. A future reader reconciling 13 against 11 will assume two call sites were missed.
- **Suggested Resolution**: Record the reconciliation in the Delivery Log ("M-7 says 13; measured 11 — 9/1/1 across three files; all 11 updated") and correct M-7.
- **Related AC**: AC22, D-1.5.5, D-1.4.15

### Finding 12: the AD-14 red-proof record overstates which rows redden

- **Severity**: Minor
- **Category**: Tests
- **Location**: story Delivery Log (Task 7 bullet) and the D-000.13 mutation table row 4
- **Observation**: The record says *"row 2 (and row 3, which depends on Presence reaching the Kind check) went red … row 1 stayed green"*. I applied three distinct collapses and measured:

  | Collapse | Row 1 | Row 2 | Row 3 |
  |---|---|---|---|
  | Null → Absent, `Present` kept (the ruling's exact scenario) | green | **RED** — `data path "customer.name" is absent…` | green |
  | Null → Present (the other two-state collapse) | green | **RED** — `is a null, not a string — text bindings are never coerced` | green |
  | Everything → Absent (degenerate) | green | **RED** | **RED** |

  Row 3 reddens only under the degenerate collapse, which also breaks `Present` and therefore also reddens unrelated tests.
- **Impact**: The substance is right and D-1.6.2 is genuinely satisfied — rows 1 and 2 do produce opposite outcomes, and I confirmed the triple **cannot pass** under any of the three collapses. But the recorded proof attributes a red to row 3 that the ruling's own scenario does not produce, which weakens the record's value as evidence.
- **Suggested Resolution**: Restate the Delivery Log entry to name which collapse was applied and that row 2 alone reddens under it.
- **Related AC**: AC11, AC25a, D-1.6.2, D-000.13

### Finding 13: AC12's finding renders its location as `.`, not as the file name

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `lint/internal/rules/forbiddenimports_test.go:200`; `lint/internal/rules/forbiddenimports.go:165`
- **Observation**: `ScanForbiddenImports` builds each message from `filepath.Rel(root, path)`. When `root` **is** the file, that relative path is `"."`. Under my `os`-import mutation the observed red was:
  ```
  .:26: forbidden import "os" (AD-1's allow-listed numeric surface: …)
  ```
- **Impact**: AC16/AC8's discipline throughout this project is a **located** error. The line number is right, the file is `.`. In a future module with more than one file-scoped call site the messages become indistinguishable. (The trailing rationale text mentioning "allow-listed numeric surface" for an import violation is pre-existing and out of this story's scope.)
- **Suggested Resolution**: Have the test include the file name in its failure message, or teach `ScanForbiddenImports` to fall back to `filepath.Base(path)` when the relative path is `"."`.
- **Related AC**: AC12, AC13

---

### Nits

- **Nit 14 — `TestSplitJSONNumberExponentBound`'s third assertion cannot fail.** `decodePoints("1e2000000")` errored correctly *before* this story (M-2 measures it at 0.07 s via the millipoint check) and errors after it. Removing Fix 2 entirely leaves it green (measured). The first two assertions carry the test; the third is documentation, not a check. `folio-go/internal/template/decimal_test.go:98–101`.
- **Nit 15 — the page-reservation red-proof's headline assertion is shadowed.** In `TestBindTextPageReservationIsRedProofable`, `if err1 != nil || err2 != nil { t.Fatalf }` precedes `if got1 != got2`. Deleting the reservation branch reddens on the **error** check (`data path "page" is absent`), never on the hijack comparison the test's own message describes. The comparison is still reachable under a different mutation, so it is not vacuous — but per D-000.13, the observed red does not name the reason the test claims. `folio-go/internal/bind/text_test.go:120–135`.
- **Nit 16 — `0e<n>` for n > 1 000 000 was legal and is now rejected.** `decodePoints("0e2000000")` previously evaluated to `0` (`digits.SetString("0")`, multiplied by the power, `IsInt64` true) and is now a splitter error. No corpus literal has this shape and nobody writes it, so this is recorded for completeness rather than as a regression to fix.
- **Nit 17 — the story banner's own verification is wrong.** Lines 25–28 state *"the working tree carries **precisely two** modified files — `deferred-work.md` (+21 lines) and `folio-mvp-decision-log.md` (+175 lines)"*. Measured against `5a1cf7d`: **three** doc files are modified — those two plus `1-4-load-validate-and-round-trip-a-folio-template.md` (+56), whose append the story's own D-1.6.6 section documents — and the decision log is **+323**, not +175. Creator-side, not developer-side; flagged because this banner is the reviewer's instruction sheet.
- **Nit 18 — top-level non-object report data reports "path absent".** `Data("null")`, `Data("[1,2,3]")` and `Data("\"a string\"")` all produce `data path "customer.name" is absent from the report data` rather than saying the report data is not an object. `Lookup`'s doc comment explicitly chooses this, so it is a documented decision, not a defect — but the message misdescribes the input a caller actually supplied.

---

### D-000.9 / D-000.13 answered for this review

1. *What would this check have printed if it had been unable to run at all?* — Every gate's exit code was captured raw via `rtk proxy … > file`; every run used `-count=1`; `folio-go` reports 7 named packages and `lint` 4, so a scan that ran nothing would print a different package list, not the same `ok`s.
2. *What would this red-proof have printed if the mutation had never been applied?* — Each mutation was applied and reverted with a `cp`-backed pristine copy, and the suite was confirmed green before and after. The pre- and post-review `git status --porcelain` are byte-identical and the per-file diffstat vs `5a1cf7d` is unchanged.
3. *Did it fail for the reason it names?* — Every red I induced was a **valid-syntax, forbidden-semantics** control (`go build ./...` exit 0 under the `os`+`float64` mutation) and every observed red was read for its message, not its exit status: `render_entry.go:78: identifier float64` / `untyped floating-point literal 1.5`; `.:26: forbidden import "os"`; `data path "customer.name" is absent…`; `exponent magnitude one past the bound (1000001) must be a located error`; `panic: test timed out after 25s`. The two places where the answer is **no** are Finding 1 (the AC12 guard does not fail under the mutation its AC describes) and Nit 15.

---

## Finding Resolutions (Story Finisher)

**Status: all 18 findings triaged; 15 FIX, 3 DISMISS, 0 DEFER.** Every FIX below was applied with a
surgical, minimal-diff change, then red-proofed live by reproducing the reviewer's own mutation (or,
where the reviewer's evidence was itself a live measurement, by re-running that exact measurement)
and observing the correct failure message before restoring. Transcripts are summarized per finding;
raw output was captured to scratch files, never through a shell pipe (D-000.12).

### Triage table

| # | Sev | Title (short) | Decision | Rationale |
|---|---|---|---|---|
| 1 | Blocker | AC12 guard asserts a filename, not a property | **FIX** | AC12/D-1.6.3's own text tags the property binding and the filename illustrative; the shipped assertion inverted that. A structural fence that a same-package function move defeats is not a fence — and Story 1.7 (`RenderTo`) is scheduled to make exactly that move. `lint/internal/rules/forbiddenimports_test.go`: added `findRenderDeclaringFiles` (locates the file(s) under `folio-go/` whose AST declares `Render`/`RenderTo`), replacing the hard-coded path. Red-proofed live: reproduced the reviewer's mutation (moved `Render` into `folio.go`) — the guard now correctly fails naming `forbidden import "os"` at `folio.go:8`; restored. |
| 2 | Major | `Decimal` is unwired (`NewDecimal`/`AsDecimal` have no production caller) | **FIX** | AC3/AC4 are binding ACs of *this* story requiring a located error naming the data path and element id for an out-of-bound literal; that requirement was simply unimplemented on any reachable path, not a scope question to defer. `internal/bind/text.go`: `BindText`'s `Present`/`KindNumber` arm now calls `val.AsDecimal()` before the (still-mandatory, AC10) wrong-kind rejection, so a malformed literal is reported as such — located, naming path and element id — while a well-formed one is still rejected as wrong-kind, never coerced. Chose the "wire it now" option over "defer with a DW entry" because the fix is a five-line, single-call-site change with no scope expansion, and leaving a story's own headline deliverable provably dead code is a correctness gap, not a scope boundary. Red-proofed live: removed the `AsDecimal` call, confirmed `TestBindTextNumberCoefficientOverflowIsLocatedError` reddens on the generic "not a string" message (no "coefficient" mention); restored. |
| 3 | Major | `Decimal`'s exponent bound checks the wrong quantity | **FIX** | The constant's own doc comment stated an invariant ("strictly narrower than the splitter's bound") the code violated for two literal shapes (no `"e"` notation at all; a long fractional part stacked on an explicit exponent) — a correctness bug in a bound this story exists to build. `internal/bind/decimal.go`: the bound check moved to after `exponent := exp - len(fracPart)` is computed, and now applies to that constructed value. Added tests at the constructed bound and one past it, for both the no-notation and stacked forms. Red-proofed live: reverted to checking raw `exp` before computing `exponent` — both new tests reddened for the right reason (`err == nil` where an error was required); restored. |
| 4 | Major | `DecodeData` silently accepts trailing garbage | **FIX** | `Render`'s own documented precondition ("d must be syntactically valid JSON", AC24) was unenforced, and the failure mode (silently rendering only the first of two concatenated documents) is exactly the class of defect this story's untrusted-input framing (AC4's rationale) exists to close. `internal/bind/value.go`: `DecodeData` now confirms the stream is exhausted via `dec.Token() == io.EOF` after the successful `Decode`; anything else is a located error. New `value_test.go` covers trailing garbage, two concatenated documents, and confirms trailing whitespace alone (the ordinary case) stays legal. Red-proofed live: removed the trailing-content check — `TestDecodeDataRejectsTrailingContent` reddened on all three cases (`err == nil` where an error was required); restored. |
| 5 | Major | `boundText == ""` suppresses font-chain validation | **FIX** | AC9 requires only that a null binding "renders as empty, and is not an error" — it does not license skipping the element's own validation, and the observed defect (the same broken template passing or failing depending on which report it is handed) is exactly the "caught in CI, not production" failure Story 1.5's font-chain check exists to prevent. `render.go`: `resolveFace` now runs before the AC9 empty-text short-circuit; only run emission (not validation) is skipped for empty text. New test `TestRenderNullBoundTextStillValidatesFontChain` (`render_bind_test.go`) reproduces the finding's own table (null vs empty-string binding, plus a non-empty control) against an unresolvable font chain. Red-proofed live: restored the original ordering — both subtests reddened (`err == nil` where the font-chain error was required); restored. |
| 6 | Minor | Retained corpus fixture proves only Fix 2, not Fix 1 | **FIX** | AC4c's obligation is that the fixture "proves Story 1.4's promise now holds through the public path" for the DoS fix; one of the fix's two required, independently-load-bearing mechanisms (AC4a) had zero public-path coverage — the same shape of gap as the original defect. Added `testdata/template/malformed/huge-exponent-negative-wrap.folio` (literal `1e9223372036854775808`, which wraps to `math.MinInt64` and defeats Fix 2 alone) and `TestLoadTemplateRejectsNegativeWrapExponentQuicklyToo`. Red-proofed live: removed Fix 1 alone — the *original* fixture's test stayed green (0.407s, confirming the reviewer's own finding), the *new* fixture's test hung and was killed by `-timeout` at 15.190s, reproducing M-1's `math/big` mechanism through `ParseDocument`; restored. |
| 7 | Minor | Splitter's bound is 10x the widest actual consumer bound | **FIX** | AC4b's letter is that the shared splitter's bound must be "the WIDER of the two" consumer bounds; 1,000,000 was ten times wider than the actual wider bound (100,000), and while the reviewer confirmed this was not a DoS (bounded, input-proportional work), it was needless attack surface and a literal violation of the ruling's own arithmetic. `internal/template/decimal.go`: `MaxSplitExponentMagnitude` corrected to `100_000`. Updated `TestSplitJSONNumberExponentBound`'s literals to match, and added the negative-side pair (Finding 9 folds in here too). Doc comments on both `MaxSplitExponentMagnitude` and `maxDecimalExponentMagnitude` corrected to state what the code now actually does. |
| 8 | Minor | `render_entry.go`'s file comment leaks into public godoc as a second package comment | **FIX** | Internal ruling IDs and story numbers were being published in the library's public API documentation, which is a real (if low-severity) maintainability/professionalism defect for a library other code will consume. Added a blank line between the comment block and `package folio`. Verified via `go doc .`: only `folio.go`'s package comment now appears. |
| 9 | Minor | Neither exponent bound tested on its negative side | **FIX** | AC4b makes "tests at the bound and one past it" binding; the omission was real and cheap to close alongside Finding 3/7's edits to the same functions. Added `parseDecimalExponent("-100000")`/`("-100001")` and `NewDecimal("1e-100000")`/`("1e-100001")` assertions. |
| 10 | Minor | `D-1.3.5` governs shipped code but has no disposition row | **FIX** | AC30/D-000.10's whole purpose is that no ruling the shipped code invokes passes unreviewed; `value.go`'s own comment and the Delivery Log both cite D-1.3.5, so its absence from the table is a real enumeration gap regardless of whose diligence should have caught it. Added as row 35 in the corrected disposition table below. |
| 11 | Minor | M-7's "thirteen call sites" is wrong (actual: 11) | **FIX** (doc) | D-1.4.15's own discipline is that a departure from a stated, measured reference figure is surfaced, not silently corrected past — even though the actual code (11 call sites, all updated) was right all along. Added a correction note directly under M-7, left the original paragraph in place as the historical record the finding refers to. |
| 12 | Minor | AD-14 red-proof record overstates which rows redden | **FIX** (doc) | The substance (D-1.6.2 is genuinely satisfied) is unaffected, but a record that attributes a red to a row the applied mutation did not actually redden weakens its value as evidence — exactly the D-000.13 discipline this story itself promoted to a standing rule. Added correction notes under the Task 7 Delivery Log bullet and the D-000.13 mutation table, naming precisely which collapse reddens which row. |
| 13 | Minor | AC12 finding renders its location as `.`, not the file name | **FIX** | A located error that resolves to `"."` degrades to useless the moment more than one file-scoped call site exists in the module (which Finding 1's own fix makes more likely, not less, since `findRenderDeclaringFiles` can now target multiple files). `lint/internal/rules/walk.go`: `walkGoFiles` falls back to `filepath.Base(path)` when `filepath.Rel(root, path) == "."`. Verified live in Finding 1's own red-proof transcript: the observed message now reads `folio.go:8: forbidden import "os"`, not `.:26: …`. |
| 14 | Nit | `TestSplitJSONNumberExponentBound`'s third assertion cannot fail | **DISMISS** | The reviewer's own framing is correct and low-stakes: the assertion is honestly documented as non-load-bearing already, and it provides genuine (if narrow) value as a regression pin for Story 1.4's own control case (M-2), distinct from — not overlapping — the two bound assertions that actually carry this test. Removing it would lose that regression pin for no safety gain. Expanded the in-place comment (added while fixing Finding 7, since the same test needed editing anyway) to state this explicitly rather than leaving it merely implied. |
| 15 | Nit | Page-reservation red-proof's headline assertion is shadowed | **FIX** | Leaned FIX despite Nit severity: D-000.13 is this story's own promoted standing rule, and this is a textbook instance of the exact failure class it names — the test failed under the reviewer's mutation, but not for the reason its own message describes. `internal/bind/text_test.go`: `TestBindTextPageReservationIsRedProofable` redesigned to compare `{{page}}`'s rendering against two data documents that BOTH carry a top-level `"page"` key (with different values), so a removed reservation makes both calls succeed — with different text — and the comparison itself reddens. Red-proofed live: deleted the reservation branch — the test now reddens exactly on `a top-level "page" data key changed {{page}}'s rendering: "HIJACKED" vs "SOMETHING ELSE"`; restored. |
| 16 | Nit | `0e<n>` for n > 1,000,000 was legal, now rejected | **DISMISS** | The reviewer's own conclusion: no corpus literal has this shape, nobody writes it, and it is explicitly "recorded for completeness rather than as a regression to fix." Rejecting `0e<huge>` is also the conservative, safe-by-default direction for a bound whose entire purpose is refusing absurd exponents — accepting it would require carving out a coefficient-is-zero exception to the bound check for no observed benefit. No action. |
| 17 | Nit | Story banner's own verification is wrong (Nit 17) | **FIX** (doc) | A factual arithmetic error in the reviewer's own instruction sheet, cheap to correct and directly relevant to a reader trying to reproduce "what's outside HEAD." Added a correction note under the banner with the measured, correct diffstat (three files: +21 / +323 / +56), left the original paragraph in place as the historical record. |
| 18 | Nit | Top-level non-object report data reports "path absent" | **DISMISS** | `Lookup`'s own doc comment explicitly documents this choice (treating "the shape doesn't even make sense against this path" as equivalent to "the key doesn't exist" — AC8's own framing). The reviewer agrees this is "a documented decision, not a defect." Changing it would mean inventing a fourth `Presence` outcome or a parallel error path for a case AD-14's three-case design was never asked to distinguish — out of proportion to the (real but small) message-clarity gain. No action. |

**Totals: 18 findings — 1 Blocker, 4 Major, 8 Minor, 5 Nit. 15 FIX (11 code-level + 4 doc-only), 3
DISMISS, 0 DEFER.**

> **Arithmetic check (D-000.13 applied to this table itself):** code-level FIXes are Findings 1, 2,
> 3, 4, 5, 6, 7, 8, 9, 13, 15 — **11**. Doc-only FIXes are Findings 10, 11, 12, 17 — **4**. DISMISS
> are Nits 14, 16, 18 — **3**. 11 + 4 + 3 = **18**. Matches the finding count exactly; no finding
> triaged twice, none dropped.

### Corrected ruling disposition table (D-000.10 mirror reconciliation)

The reviewer's independent table marked 10 rows ⚠️ (rows 1, 3, 4, 6, 12, 14, 17, 18, 26, 34) but its
own summary line said *"Disagreements: 8 of 34"* and named only 8 of the 10 — **the review's own
count undercounted its own table**, an instance of exactly the D-000.13 diligence this story asks the
finisher to apply to their own work. All 10 flagged rows are resolved below, plus the missing
`D-1.3.5` row (Finding 10), for **35 total rows**. "Final" reflects the disposition **after** this
finisher's fixes — every ⚠️ row below is now resolved to a genuine `applied-as-stated`.

| # | Ruling | Developer (original) | Reviewer (independent) | Final disposition (finisher) |
|---|---|---|---|---|
| 1 | D-1.6.1 | applied-as-stated | ⚠️ applied-with-deviation (Findings 2, 3) | **applied-as-stated (post-fix)** — `AsDecimal` is now called from `BindText`'s only `KindNumber` path (Finding 2 FIX); the exponent bound now checks the constructed `Exponent` (Finding 3 FIX). Both gaps closed; re-verified: `go test -count=1 ./internal/bind/...` green, 21 tests. |
| 2 | D-1.6.2 | applied-as-stated | applied-as-stated (record correction, Finding 12) | **applied-as-stated** — substance never in question; Delivery Log / D-000.13 table corrected to name the collapse actually applied (Finding 12 FIX-doc). |
| 3 | D-1.6.3 | applied-as-stated | ⚠️ **NOT applied as stated** — property unasserted (Finding 1, Blocker) | **applied-as-stated (post-fix)** — the guard now derives its target from the AST (`findRenderDeclaringFiles`), not a hard-coded filename. Red-proofed live against the reviewer's own mutation (Render moved into folio.go): now correctly fails naming `forbidden import "os"`. |
| 4 | D-1.6.4 | applied-as-stated | ⚠️ applied-with-deviation (Finding 4: trailing JSON unenforced) | **applied-as-stated (post-fix)** — `DecodeData` now rejects trailing content; AC24's full letter now holds. |
| 5 | D-1.6.5 | applied-as-stated | applied-as-stated (Nit 15) | **applied-as-stated** — the one red-proof-quality nit (Nit 15) fixed; substance was never in question. |
| 6 | D-1.6.6 | applied-as-stated | ⚠️ applied-with-deviation (Findings 5*, 6, Minor 9) | **applied-as-stated (post-fix)** — both required fixes are present and independently load-bearing (unchanged from the reviewer's own confirmation); the corpus now proves BOTH mechanisms through the public loader (Finding 6 FIX, new fixture); the negative-side bound tests are in place (Minor 9/Finding 9 FIX). *Finding 5 (font-chain validation) does not actually govern this ruling — it is an AC9/render.go defect, not part of the DoS fix; resolved separately, see row "AD-14" below.* |
| 7 | D-1.6.7 | applied-as-stated | applied-as-stated (independently re-proved) | **applied-as-stated** — no finding against this ruling; unchanged. |
| 8 | D-1.6.8 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 9 | DW-8 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 10 | Story 1.4 `## Defect found after done` | not-reached-in-this-story (by design) | not-reached-by-design | **applied-as-stated** — the finisher appended the fixing commit hash below the existing section (pure append, nothing above the heading touched); see that file for the SHA and this report's Commit section. |
| 11 | D-000.4 | applied-as-stated | applied-as-stated | **applied-as-stated** — matrix still explicitly unrun and deferred to the Epic 1 boundary; unchanged. |
| 12 | D-000.9 | applied-as-stated | ⚠️ applied-with-deviation (AC12 witness proved a file was parsed, not that it declares Render — Finding 1) | **applied-as-stated (post-fix)** — `findRenderDeclaringFiles`' own vacuity guard (zero declaring files is a hard failure) closes exactly the gap the reviewer named; the witness now proves the property it is meant to prove. |
| 13 | D-000.9 (extended) | applied-as-stated | applied-as-stated (reviewer reproduced all red-proofs) | **applied-as-stated** — the finisher additionally reproduced a fresh red-proof for every new fix (Findings 1–9, 13, 15); all failed for the reason named, all restored. |
| 14 | D-000.10 | applied-as-stated | ⚠️ applied-with-gap (missing `D-1.3.5` row, Finding 10; 6-of-34 mirror disagreements — corrected here to 10-of-34) | **applied-as-stated (post-fix)** — `D-1.3.5` added as row 35; every disagreement in this table resolved. |
| 15 | D-000.11 | applied-as-stated | applied-as-stated | **applied-as-stated** — every gate this finisher ran also used `-count=1`; unchanged. |
| 16 | D-000.12 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged; all finisher verification captured to scratch files, never a pipe. |
| 17 | D-000.13 | applied-as-stated | ⚠️ applied-with-deviation (Finding 1, Nit 15: two red-proofs failed for the wrong reason) | **applied-as-stated (post-fix)** — both instances fixed (Finding 1's guard now targets the declaring file; Nit 15's red-proof now reddens on the actual hijack comparison). Every finisher-authored red-proof in this story was itself verified against this same standard (see the Triage table above). |
| 18 | D-1.1.b | applied-as-stated | ⚠️ applied-with-deviation (same root cause as Finding 1: property bound to a filename, not file contents) | **applied-as-stated (post-fix)** — `findRenderDeclaringFiles` makes the property a fact about file CONTENTS (which file's AST declares `Render`), genuinely mirroring `numbers.go`'s precedent now. |
| 19 | D-1.1.c | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 20 | D-1.2.5 | applied-as-stated | applied-as-stated | **applied-as-stated** — the finisher's own new hang red-proof (Finding 6) also recorded verbatim: "hung; killed by `-timeout` at 15.190s". |
| 21 | D-1.2.6 | not-reached-in-this-story | not-reached-in-this-story (agreed) | **not-reached-in-this-story** — unchanged. |
| 22 | D-1.3.1 | applied-as-stated | applied-as-stated | **applied-as-stated** — new `value_test.go` imports only `testing`; unchanged conclusion. |
| 23 | D-1.3.3 (amended) | applied-as-stated | applied-as-stated | **applied-as-stated** — every finisher mutation was also valid-syntax/forbidden-semantics (confirmed `go build ./...` succeeded under each before observing the test-level red). |
| 24 | D-1.4.6 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 25 | D-1.4.8 (corrected) | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 26 | D-1.4.15 | applied-as-stated | ⚠️ applied-with-deviation (two doc comments stated invariants the code violated, Finding 3) | **applied-as-stated (post-fix)** — both doc comments (`MaxSplitExponentMagnitude`, `maxDecimalExponentMagnitude`) corrected to describe what the code now actually checks and enforces. |
| 27 | D-1.4.16 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 28 | D-1.5.5 | applied-as-stated | applied-as-stated (mechanism; M-7 count wrong, Finding 11) | **applied-as-stated** — M-7 corrected in place (Finding 11 FIX-doc); the mechanism (insertion, never reordering) was never in question. |
| 29 | AD-1 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 30 | AD-2 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged. |
| 31 | AD-4 | applied-as-stated | applied-as-stated | **applied-as-stated** — unchanged; page/pages reservation is untouched by any finisher fix (Nit 15 only strengthened its test, not the reservation mechanism itself). |
| 32 | AD-12 | applied-as-stated (content) | applied-as-stated (content); placement nit (Finding 8) | **applied-as-stated (post-fix)** — the locale clarification's content was always correct; it no longer leaks into public godoc as a second package comment. |
| 33 | AD-14 | applied-as-stated | applied-as-stated | **applied-as-stated (post-fix)** — the three-case implementation in `internal/bind` was never in question; **Finding 5's separate render.go defect (font-chain validation suppressed by an AC9-empty binding) is fixed**, so a broken template's own validation (Story 1.5 AC2/AC4) no longer depends on which of AD-14's three cases a given render happens to hit. |
| 34 | AD-23 | applied-as-stated | ⚠️ applied-with-deviation (Finding 2: conversion happens on no shipped path) | **applied-as-stated (post-fix)** — `internal/bind.DecodeData` uses `UseNumber`; `Decimal` now IS reached on a real production path (`BindText`'s `KindNumber` arm); no `float64` anywhere in the binding path (unchanged, always true). |
| 35 | D-1.3.5 | *(no row — Finding 10)* | *(no row — Finding 10)* | **applied-as-stated** — `internal/bind/value.go`'s `valueFromAny` uses the `slices.Sorted(maps.Keys(...))` escape hatch, never ranging a map directly; `ScanMapRange`'s existing `internal/`-scoped production scan covers `internal/bind` automatically (no separate assertion needed, same reasoning as row 22/D-1.3.1). Added as this table's 35th row per Finding 10. |

**Reconciliation: 10 of 35 rows carried a disagreement (all ⚠️ rows above); all 10 are resolved to
`applied-as-stated (post-fix)` following the finisher's fixes. 0 rows remain in dispute.**

### Follow-ups (none deferred)

No finding was triaged DEFER. The three DISMISS items (Nits 14, 16, 18) are closed, not deferred —
each is a reviewer-endorsed non-issue with an explicit rationale above, and none implies future work.
`deferred-work.md` is unchanged by this story finish (DW-8 remains the only entry this story touches,
verified matching in the disposition table above, row 9).

### Finisher validation gates (re-run after ALL fixes, AC26/AC27/D-000.11/D-000.12)

Every gate below is `-count=1` where applicable; exit codes read raw (never through a shell pipe;
`gofmt -l` empty output is itself the pass signal).

| Gate | Result |
|---|---|
| `folio-go` `go build ./...` | exit 0 |
| `folio-go` `go vet ./...` | exit 0 |
| `folio-go` `go test -count=1 ./...` | exit 0 — **7 packages**, all `ok` (matches the developer's reference figure) |
| `folio-go` test count | **220 total `--- PASS`** (129 top-level `Test…` functions + 91 subtests), **0 `--- FAIL`** |
| `folio-go` `go build -tags=matrix ./...` | exit 0 |
| `folio-go` `go vet -tags=matrix ./...` | exit 0 |
| `folio-go` `gofmt -l .` | empty |
| `lint` `go build ./...` | exit 0 |
| `lint` `go vet ./...` | exit 0 |
| `lint` `go test -count=1 ./...` | exit 0 — 4 packages, all `ok`, **40 `--- PASS`**, 0 `--- FAIL` |
| `lint` `GOPROXY=off go test -count=1 ./...` | exit 0 — same 4 packages, all `ok` |
| `lint` `gofmt -l .` | empty |
| `hashmatrix` `go vet ./...` | exit 0 |
| `hashmatrix` `go test -count=1 ./...` | exit 0 — `probe` `[no test files]` |
| `hashmatrix` `go build ./...` | **environmental quirk, not a code defect** — `go: build output "probe" already exists and is a directory`: `hashmatrix/probe/` is both a package directory and `go build`'s default output-binary name for it, colliding when run from `hashmatrix/` with no `-o`. Confirmed a real compile with an explicit output path (`go build -o <tmp>/ ./...`): exit 0. Pre-existing (this story touches zero files under `hashmatrix/`), not introduced by this story. |
| `lint` licence manifest (`cd lint && go run ./cmd/genmanifest`) | exit 0, `lint/MANIFEST.md` regenerated **byte-identical** to its committed form (`git status --porcelain -- lint/MANIFEST.md` empty) — this story added no new third-party dependency |
| Cross-target hash matrix | **NOT run — explicitly deferred to the Epic 1 boundary (AC28, D-000.4).** This story is not one of D-000.4's override stories (1.2, 1.5, 1.8, 2.4, 4.7). |

**Re-run red-proofs, confirmed failing for the reason named (D-000.13), all restored via `cp` before
the gates above:** Finding 1 (AC12 guard: moved `Render` into `folio.go` — guard failed naming
`folio.go:8: forbidden import "os"`); Finding 2 (`Decimal` wiring: removed the `AsDecimal` call —
`TestBindTextNumberCoefficientOverflowIsLocatedError` failed on the generic wrong-kind message, no
"coefficient" mention); Finding 3 (exponent bound: reverted to checking raw `exp` — both new
constructed-bound tests failed with `err == nil`); Finding 4 (`DecodeData`: removed the
trailing-content check — all three `TestDecodeDataRejectsTrailingContent` cases failed with
`err == nil`); Finding 5 (font-chain validation: reverted the ordering — both subtests of
`TestRenderNullBoundTextStillValidatesFontChain` failed with `err == nil`); Finding 6 (DoS Fix 1:
removed the during-accumulation overflow check — the new negative-wrap fixture's test hung and was
killed by `-timeout` at **15.190s**, ~30 `math/big` frames rooted at `decodePoints`, while the
original fixture's test stayed green in 0.407s, confirming the reviewer's own finding); Finding 15
(page-reservation red-proof: deleted the reservation branch — the redesigned test failed on the
actual hijack comparison, `"HIJACKED" vs "SOMETHING ELSE"`). Every mutation was valid Go with
forbidden semantics (each `go build ./...`'d cleanly before the test-level red was observed),
satisfying D-000.13's own third question for every fix this finisher made, not only the ones the
story shipped.
