# Story 3.1: Bind a repeating region to a collection with an explicit row scope

Status: **done**

**Covers:** FR16, FR17, FR21 · AD-11 · **first story in Epic 3** ("A Go developer can render
computed, parameterised documents")

**Baseline commit:** every measurement in this file was taken at `e7f3f9c` ("Epic 2 boundary gate:
run, verified, and closed") on a clean tree. **HEAD moved to `a20f60e` during story creation**
("Epic 3 opening: file the session-4 lead grounding refresh") — that commit changes
`folio-mvp-decision-log.md` **only** (`1 file changed, 61 insertions(+)`, no code, no tests, no
fixtures), so every figure and file:line reference below carries unchanged to `a20f60e`. Epics 1
and 2 are `done`; the Epic 2 boundary gate closed 2026-08-25 over one REQUIRED red (see *Baseline*
below).

> **Standing rule for this file.** If anything here looks like it contradicts an entry in
> `folio-mvp-decision-log.md`, the spine, or `folio-format.md` — **stop and surface it. Do not
> resolve it by choosing.** One decision is deliberately left open below (OD-1) and the developer
> must not write it.

> **Stop at `review`.** The final task is "story file, decision log, sprint status → `review`".
> **Do not commit, do not branch, do not set `done`.** Committing belongs to the finisher, after
> review.

---

## The epic's acceptance criteria, verbatim, cited BY ORDINAL

`_bmad-output/planning-artifacts/epics.md:957-986`. **Five `**Given**` blocks — counted at
creation, and the count is five.** They are reproduced here so the developer never works from a
paraphrase.

- **Source AC1** — *Given a repeating region declaring `{"bind": "transactions[]", "as":
  "transaction"}` / When its contents are evaluated for a row / Then `transaction.*` resolves to
  that row's fields*
- **Source AC2** — *Given a repeating region that omits `as` / When its contents are evaluated /
  Then the alias defaults to `row`*
- **Source AC3** — *Given an unqualified path inside a repeating region / When it is resolved /
  Then it resolves from the document root — a row never shadows the root*
- **Source AC4** — *Given a `params.` path anywhere, including inside a row scope / When it is
  resolved / Then it resolves from the parameter namespace and can be shadowed by nothing*
- **Source AC5** — *Given a collection binding whose path is not an array / When the document
  renders / Then the render fails with an error naming the path and the element id*

AD-11 (`ARCHITECTURE-SPINE.md:253-266`) is the governing invariant and is quoted in *Do not
re-open* below.

---

## Baseline, measured at creation

Measured at `e7f3f9c`, clean tree, on this machine, through
`env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` (D-000.12 — never through the
shell wrapper's pipes; the redirect is outside the proxied string):

| Scope | PASS | FAIL | SKIP |
|---|---|---|---|
| all-occurrence (`--- PASS` anywhere, subtests included) | **600** | **1** | **1** |
| top-level (`^--- PASS`) | **368** | **1** | **1** |

- The one FAIL is `TestCorpusMeetsP6ExerciseFloors` (`internal/text`), P6g 7 < 20. It is the
  **REQUIRED red** D-000.17 / D-2.1.14 mandate stays unmet and over which the Epic 2 gate closed
  (D-000.57). **It must still be red at handoff.** Turning it green is a regression, not a fix.
- The one SKIP is `TestXrefEntriesRejectsMalformedSubprocess`.
- Report your delta against **both** scopes, with flags and package scope named (a bare number is
  not a measurement).

---

## In plain terms (read this first if you just want the gist)

A statement has a table of transactions. The author places that table once and the engine repeats it
— one row per transaction. This story settled, in the running code, which transaction a field name
written inside that table refers to, and what happens to names that belong to the whole document
rather than to any one row.

The rule shipped as designed: the author names the row, or it is called "row" by default; a name
written inside the table means the current row and nothing else; a plain, unqualified name always
means the whole document, so a row can never quietly take over a name the rest of the document is
already using; the runtime parameters supplied by the calling program — the report date, the branch
name — keep a namespace that nothing can take over, not even a row; and if the author binds a table
to something that turns out not to be a list at all, the render stops and says which binding and
which element, rather than producing a plausible-looking document with the table silently missing.

An independent review then checked the work against its own tests, rather than taking the story's
account of them on trust. It found the naming rule itself was sound throughout, but one of its
tests — the one meant to prove that a row name colliding with a protected name is rejected and the
offending name is reported back correctly — had been written loosely enough that it would have
stayed green even if the reported name were wrong. That test has been tightened so it can actually
fail on the claim it makes, and it was then run against the exact mistake the review demonstrated,
to confirm it now catches it. Two smaller coverage gaps the review also flagged — one where a test
could not distinguish two related rules from each other, one where a new error case had shipped
with no test at all — were closed the same way, each checked against the specific defect it was
meant to catch. A cosmetic record-keeping nit, tracing to files that exist only on the machine that
did the work and never ship with the project, was investigated and left as is, with the reason
written down.

No table actually prints a row yet — that remains true, and is intentional rather than an
oversight. This story settles what a name means inside a table; producing rows on a page is a later
story's job.

---

## Story

**As a** template author,
**I want** to bind a region to a collection and address each row's fields unambiguously,
**So that** the rule for what a row-scoped field reference means is written down rather than
guessed.

---

## Do not re-open — settled rulings this story inherits

Each is reproduced with its rationale. **None of these is a question for this story.**

1. **AD-11 — Row scope is an explicit alias** (`ARCHITECTURE-SPINE.md:253-266`), verbatim:
   > a repeating region declares its alias: `{"bind": "transactions[]", "as": "transaction"}`,
   > defaulting to `row` when omitted. Inside the region, the alias resolves to the current row.
   > Unqualified paths always resolve from the document root — a row **never** shadows the root.
   > `params.` is always the parameter namespace and can be shadowed by nothing.

   **Prevents** (AD-11's own words): "the expression engine inferring a row alias by singularizing
   the collection name while the binding UI shows something else — two rules for the same
   `{{transaction.amount}}`."

2. **AD-11's aggregate clause is NOT this story's.** "Aggregates always take a root-relative
   collection path … always computed over the whole collection, never over the rows on the current
   page." That is Story 3.3 / 4.5. Do not build it, do not stub it.

3. **AD-23 / D-000.5 — no `float64` under `internal/` or the module root.** The row value this
   story carries is `bind.Value`, which preserves a number literal's own text
   (`internal/bind/value.go:53`, `Num string`, decoded through `json.Decoder.UseNumber`). **The row
   scope must carry `Value`, never a decoded numeric.** Story 3.3 needs exact decimal arithmetic
   over exactly these values; a row model that flattened them to a Go numeric type would foreclose
   it.

4. **D-1.7.4 — `params` is a namespace, a second resolution root, not a reserved token.**
   `page`/`pages` are reserved whole tokens resolved from neither root (AD-4, Story 2.7);
   `params` is a root. `internal/bind/text.go`'s `BindText` doc comment carries the fence verbatim:
   "conflating the two is how 'page' would eventually acquire a namespace, which AD-4 forbids
   forever." **A row alias is a root, like `params` — never a reserved token.**

5. **D-1.6.5 / D-1.6.8 — binding is scoped to text-element `value` interpolation.**
   `table.bind` and `columns[].bind` are *not* scanned by `bind.BindText`. This story reads
   `table.bind` for the first time, as a **collection path**, not as interpolated text. It does
   **not** start evaluating `columns[].bind` — those are expression-shaped and belong to 3.2.

6. **D-000.13 — a red-proof asserts on rule id and message, never on exit status.** Every negative
   assertion in this story asserts on the returned error's **text**, matched against the path and
   the element id.

7. **D-000.21 — assert on the produced thing.** Assert the resolved string, never "a lookup
   happened".

8. **D-000.9 / D-000.50 — a guard's all-clear must not be producible by the absent case, and before
   writing a guard, ask whether any subject can express the defect.** Applied per-AC below.

9. **Task breakdown ends at `review`.** No commit task, no branch, no `done`.

---

## Measured findings — read all of these before writing code

Every claim below was measured at `e7f3f9c` on a clean tree during story creation. File:line
references are at that commit.

### 1. Table elements are ALREADY parsed, round-trip, and are rendered by NOTHING

`internal/template/model.go:213-219` already carries `TableExt{Bind string; As Presence[string];
Columns []Column; HeaderHeight; AltRowBackground}`, populated by `decodeTableExt`
(`internal/template/parse_bands.go:260-300`). `As` is optional and **is not defaulted at load** —
`Presence[string]`, absent when `as` is omitted. Story 1.4 shipped this.

The render path ignores tables completely. `render.go` has exactly two element-type filters:

- `render.go:264` — `if el.Type != template.ElementImage { continue }` (`collectImageRuns`)
- `render.go:453` — `if el.Type != template.ElementText { continue }` (`collectBandTextRuns`)

There is no third. **No code anywhere reads `el.Table`.** So AC5's check is genuinely new wiring,
not a repair.

### 2. `table.bind` is stored RAW, and the `[]` suffix is NOT validated at load

`parse_bands.go:264-272` stores `bind` as whatever string was given. `parse_bands.go:291` computes
`collection := strings.TrimSuffix(bind, "[]")` for the `footerOf` derivation only — `TrimSuffix` is
lenient, so `"transactions"` (no suffix) loads today exactly as `"transactions[]"` does.
`grep HasSuffix parse_bands.go parse.go` returns **nothing**.

`folio-format.md:211` says `bind` is "The collection path, suffixed `[]`". **This story does not
add a load-time syntax rule** (see *Scope fence*) — source AC5 says "**When** the document
renders". The resolver strips one trailing `[]` if present and resolves the remainder as a bare
dotted path.

### 3. AC5's red-proof, MEASURED at `e7f3f9c` — all four non-array cases render SUCCESSFULLY today

Written into `folio-go/zz_scratch_ac5_test.go`, run, then removed; the tree was verified clean
afterwards (`git status --short` empty, HEAD `e7f3f9c`). The fixture is reproduced verbatim in
*The measurement fixture* below. Recorded output:

```
=== RUN   TestZZScratchAC5NonArrayBindFailsRender
=== RUN   TestZZScratchAC5NonArrayBindFailsRender/absent
    zz_scratch_ac5_test.go:44: AC5: render SUCCEEDED on a table bind "transactions[]" that is not an array (absent)
=== RUN   TestZZScratchAC5NonArrayBindFailsRender/scalar
    zz_scratch_ac5_test.go:44: AC5: render SUCCEEDED on a table bind "transactions[]" that is not an array (scalar)
=== RUN   TestZZScratchAC5NonArrayBindFailsRender/object
    zz_scratch_ac5_test.go:44: AC5: render SUCCEEDED on a table bind "transactions[]" that is not an array (object)
=== RUN   TestZZScratchAC5NonArrayBindFailsRender/null
    zz_scratch_ac5_test.go:44: AC5: render SUCCEEDED on a table bind "transactions[]" that is not an array (null)
--- FAIL: TestZZScratchAC5NonArrayBindFailsRender (0.00s)
```

Note also what this proves incidentally: a template with `"bind": "transactions[]"` **and** `"as":
"transaction"` **parses today with no change to the loader**. The story needs no load-side work.

### 4. THE COMMITTED VIOLATION — `TestRenderScopeFenceIgnoresTableBind` WILL go red, and it is the only one

D-000.50's pre-flight run in the "which committed artifacts already violate the constraint I am
about to enforce?" direction. The whole population of render-time table elements in the repository
is **one**:

| Artifact | Table? | Rendered? | Violates AC5? |
|---|---|---|---|
| `folio-go/render_bind_test.go:19` (`bindTestTemplateJSON`, element `e2`) | yes | **yes**, `Render` at line 69 | **YES** |
| `folio-go/testdata/template/golden/worked-example.json` (element `e2`) | yes | **no** — only round-tripped (`internal/template/roundtrip_test.go:19,55`) and fence-compared (`goldenfixture_test.go:19,191`) | no |
| `internal/template/fixtures_test.go:252`, `footer_test.go:16` | yes | no — loader tests | no |
| every `fixtures/*/expected.pdf` golden, `page-count-20`, the matrix documents | **no table at all** | yes | no |

`render_bind_test.go:19` declares `"bind": "{{transactions[]}}"` — braces and all — and
`render_bind_test.go:68` renders it against `{"customer": {"name": "Ada Lovelace"}}`, which has no
`transactions` key. `TestRenderScopeFenceIgnoresTableBind` **asserts the render SUCCEEDS**. It is
green at baseline (verified: `--- PASS: TestRenderScopeFenceIgnoresTableBind (0.00s)`).

Two things are wrong with that template under this story, and only one of them is about AC5:

- `"{{transactions[]}}"` is **not a collection path**. `folio-format.md:211` says `bind` is the
  collection path. D-2.4.1 already ruled the equivalent question for `unbreakableValues` — "one
  path convention in the format, not two" — and `internal/template/unbreakable_test.go:118`
  rejects `"{{customer.name}}"` and `"transactions[].payee"` at load for exactly that reason.
- Even corrected, the data supplies no `transactions`, which AC5 makes a render error.

**Required disposition** (this is settled by `folio-format.md:211` + source AC5 together, and is
not a developer choice): amend `bindTestTemplateJSON` so that
- `table.bind` becomes the bare collection path `"transactions[]"`, and
- the test's data supplies a real array for it,

**while keeping `columns[0].bind` expression-shaped exactly as it is today**
(`"{{formatNumber(transaction.amount, \"#,##0.00\")}}"`). The test's actual purpose —
D-1.6.8's field-scope fence, "a table's `columns[].bind` is never scanned by text binding" — is
**preserved and still non-vacuous**, because 3.1 does not evaluate column binds either. Update the
test's doc comment (`render_bind_test.go:53-62`) to say that the fence it proves is now the
**column** fence, and that `table.bind` is read as a collection path from this story on. Do not
delete the test and do not weaken its assertion to "no error is required".

### 5. THE STRUCTURAL OBSTACLE — a third resolution root trips `TestBindResolutionRootsAreClosed`, BY DESIGN

`internal/bind/text.go:88` declares `var declaredResolutionRoots = []string{"data", "params"}`.
`internal/bind/resolution_roots_arch_test.go` AST-scans every non-test `.go` file in package `bind`
for `lookupBound` call sites, extracts argument index 4 (`rootName`) as a **string literal**, and
compares the observed set against the declared set **both directions**. It has two teeth:

- a `rootName` that is **not a string literal** is a `t.Fatalf`, explicitly because "a computed root
  name is exactly how a third resolution root would be smuggled past a literal scan";
- an observed root outside the declared list fails with: *"a THIRD resolution root has appeared. If
  this is deliberate, it is a direction change under AD-4 … and needs the engineering lead's
  ruling, not a one-line edit to this list"*.

**Measured**, by adding a scratch non-test file to `internal/bind` containing one
`lookupBound(row, path[1:], path, elementID, "row", "the current row")` call and running the guard
(file removed afterwards; tree verified clean):

```
=== RUN   TestBindResolutionRootsAreClosed
    resolution_roots_arch_test.go:185: OBSERVED resolution root "row" (a lookupBound call site's rootName) is not in declaredResolutionRoots [data params] — a THIRD resolution root has appeared. …
    resolution_roots_arch_test.go:197: declared resolution roots: [data params]; observed at 3 lookupBound call site(s): [data params row]
--- FAIL: TestBindResolutionRootsAreClosed (0.00s)
```

**Two consequences, both binding:**

- **The alias must NEVER be the `rootName`.** The author's alias is document data and is not a
  literal; passing it as `rootName` trips the non-literal `Fatalf`, and "extend the declared list
  per document" is not a coherent thing to do to a closed set. Pass the **literal** `"row"` as the
  root-class name and carry the alias as ordinary data used for dispatch and for error text.
- **`declaredResolutionRoots` becomes `{"data", "params", "row"}` — one visible, argued line.**
  This is what AD-11 authorises: AD-4's fence is about `page`, and `page` remains a reserved
  **token** resolved from **neither** root. `TestBindResolutionRootsClosureRedProof` (which injects
  a `"page"` rootName into a scratch copy of `text.go`) must be **re-run after the edit and must
  still redden** — that is the proof AD-4's fence survived the widening. See *Ratification
  requested*.

### 6. AC4 is mostly ALREADY TRUE at baseline — retarget it at "prove it survived"

`internal/bind/text.go:277-292`: the `path[0] == "params"` branch is taken **before** any data
lookup, decided on the parsed path segment (so `{{ params.x }}` with whitespace behaves
identically, D-1.7.4/M-6), and a top-level `params` key in report data is unreachable ordinary JSON.
So "resolves from the parameter namespace and can be shadowed by nothing" holds today **against the
data root**. What does not exist at baseline is a row that could shadow it. **AC4's work is
therefore: keep the params branch first in the dispatch, and prove it stays unreachable from a row.**
Do not rewrite the params branch.

### 7. `internal/expr` and `internal/diag` MUST NOT be created by this story — two LIVE tripwires

Both verified present at `e7f3f9c`:

- `lint/internal/rules/absences.go:95-97` — rule `absence-expr-package`,
  `"folio-go/internal/expr/ must be absent until Story 3.2 derives columns[].footerOf"` (DW-5).
- `lint/internal/rules/absences.go:101` — rule `absence-diag-package` (DW-6), owned by whichever
  story first creates `internal/diag`, expected 3.6.

Creating either package reddens the production lint scan and pulls DW-5's or DW-6's whole
obligation into this story. **All of 3.1 lands inside `internal/bind` and `folio-go/render.go`.**

The session-4 grounding refresh (`folio-mvp-decision-log.md`, added at `a20f60e`) names this the
epic's "highest-probability quiet loss": all three absence tripwires (`absence-expr-package`,
`absence-diag-package`, `absence-source-date-epoch`) fire during Epic 3, and **in each case the
cheapest fix to the red is to delete the rule**, retiring the forcing function silently. DW-6
records the correct pattern — delete the rule *and* land the replacing positive assertion in the
same commit — but writes it down for `diag` only. **If this story ever finds itself deleting an
absence rule, that is the signal it has left its scope.** It has no business touching any of the
three.

The same refresh records four Epic 3 findings (the `math/big.Float` hole in AD-23's enforcement,
Story 3.2's "exactly eight entries" sequencing contradiction, AD-23's silence on accumulator
overflow, and DW-8's `Decimal` move being owned by 3.2 while 3.3 is what needs it). **None of them
lands on 3.1** — this story does no arithmetic and registers no functions — but they are why
inheritance item 3 above is not negotiable: the row scope carries `bind.Value`, literal text
intact, so 3.3 still has an exact-decimal option to take.

### 8. The render seam, named

`renderDocument` (`render.go:916`) currently: `documentBands` → `pageGeometryOf` →
`collectImageRuns` → PHASE A `collectBandTextRuns` (content) → `layout.Paginate` → PHASE B
(header, footer). `documentBands` (`render.go:222-232`) returns bands in the authored order
**0 pageHeader, 1 content, 2 pageFooter**.

AC5's check belongs **early — after `documentBands`, before `collectImageRuns`** — so that a
not-a-list binding fails before any font work, and so its error ordering is plain document order
across all three bands. Walk bands in `documentBands` order and elements in declaration order; the
**first** offending table element is the reported one (deterministic, never a map — D-1.3.5).

---

## Ratification requested (not blocking — the direction settles it, the list edit wants a record)

`declaredResolutionRoots` gains `"row"`. AD-11 makes the row scope's existence direction, so the
*whether* is settled; the guard's failure message demands that the **list edit** be a ruling rather
than a one-liner. Record it in the decision log as a D-3.1.x entry stating: (a) the third root is
the row class, authorised by AD-11; (b) the root-class name is the literal `"row"` and an author's
alias is **never** a `rootName`; (c) AD-4's fence is untouched — `page`/`pages` remain reserved
tokens resolved from neither root — and `TestBindResolutionRootsClosureRedProof` was re-run after
the widening and still reddens.

---

## DECISIONS NEEDED — escalate before writing the affected code

### OD-1 (the lead's) — an alias that collides with an unshadowable name

AD-11 says `params` "can be shadowed by nothing" and AD-4 says `page` never acquires a namespace.
Both hold under the dispatch order this story fixes (params first, reserved tokens first). But
nothing in the recorded direction says what a document declaring `"as": "params"` — or
`"as": "page"` — *is*.

- **Arm A — silently unreachable.** The alias is simply never matched; the region has a scope
  nobody can address. Cheapest, and consistent with "shadowed by nothing" taken literally.
- **Arm B — a load error.** D-2.4.1's precedent for `unbreakableValues` is directly analogous and
  its reasoning transfers word for word: an unenforced declaration "LOOKS honoured and silently is
  not" (`internal/template/unbreakable_test.go:108-113`). Arm B adds a load-time constraint, which
  this story otherwise deliberately does not do, and would need its own violation sweep (measured:
  **zero** committed documents declare such an alias, so the sweep is empty).

**The developer must not choose.** AC1–AC5 below are fully implementable while this is
adjudicated: no AC depends on the answer, and no test may assert either behaviour until it is
ruled. If it is ruled Arm B, that is a sixth AC, added by whoever rules it.

---

## Scope fence — what this story is NOT

- **Not row generation.** No repeated rows, no row heights, no cell layout, no table geometry.
  Source AC1 says "when its contents are **evaluated** for a row", not "rendered". FR22 and
  Story 4.2 own generation; Story 4.2's own third AC says cells "resolve within the table's row
  scope **per Story 3.1**" — this story supplies that, and only that.
- **Not the expression language.** `columns[].bind` stays unevaluated and expression-shaped. Story
  3.2 owns FR18, and DW-8 pre-commits that 3.2's parser **replaces** 1.6's path matcher. **Do not
  extend the path grammar** in `parseBindingPath` (`internal/bind/text.go:395`) — no parentheses,
  no brackets, no commas, no literals.
- **Not aggregates.** `sum`/`count`/`avg`, `columns[].footer`, `footerOf`, `footerFormat`: Story
  3.3 / 4.5, DW-5 / DW-7.
- **Not locale formatting** (3.4), **not `visibleIf`** (3.5), **not diagnostic codes** (3.6), **not
  the CLI** (3.7).
- **Not `internal/expr` and not `internal/diag`.** Finding 7. Both are live absence rules.
- **Not a new load-time rule.** `parse_bands.go` and `model.go` are unchanged by this story. In
  particular: `As` stays `Presence[string]` and is **not** defaulted at load — the default is a
  resolution-time rule (source AC2), and defaulting it in the parser would change the round-trip
  fixed point that `roundtrip_test.go` pins.
- **Not a change to any error message text on the existing paths.** The refactor in AC0 is
  byte-neutral in what it emits.
- **Not `page`/`pages`.** They remain reserved tokens, resolved from neither root. AD-4 stands.
- **Not any golden movement.** No committed golden contains a rendered table (finding 4). Byte
  identity is asserted, not assumed — AC7.
- **Not `epic-3: backlog`.** Flip it to `in-progress` per the tracker's own workflow note; do not
  touch `epic-3` retrospective or any later story key.

---

## Acceptance Criteria

**No AC may be satisfied by asserting on a table that renders nothing.** Every resolution AC below
is asserted on a **returned resolved string** or a **returned error's text**, through the row-scope
API, never on "a table element was visited". Nothing this story builds produces a visible glyph, so
"the render succeeded" is not evidence for AC1–AC4 and must not be used as such.

### AC0 — One resolver, reached through a scope; the existing call sites move onto it

**Given** `internal/bind`,
**When** a placeholder is resolved,
**Then** there is exactly **one** implementation of the dispatch, taking a scope value that carries
the data root, the params root, and — when a row scope is active — the current row and its alias;
**and** `BindText`/`BindTextSpans` remain as wrappers that construct a scope with **no** row,
**and** `collectBandTextRuns` (`render.go:453`) resolves through the same scoped entry point with
no row set.

The row scope must not be a second traversal. `BindTextSpans`' own doc comment already states the
discipline for exactly this reason ("there is exactly one implementation of the binding grammar and
the spans cannot drift"); the same applies to the roots.

**Red-proof:** none needed — this is structural. Its correctness is discharged by AC7 (byte
identity) plus the unchanged existing `internal/bind` suite. **Every existing error message must
remain byte-identical**; `internal/bind/text_test.go` asserts on message text and must pass
unedited.

### AC1 — The declared alias resolves to the current row's fields (source AC1)

**Given** a repeating region declaring `{"bind": "transactions[]", "as": "transaction"}` and a
scope whose current row is one element of that collection,
**When** a placeholder `{{transaction.<field>}}` in that region is resolved,
**Then** it returns **that row's** value for `<field>`,
**And** a field absent from the row is an error naming the row path and the element id (AD-14's
three cases, unchanged: absent → error, explicit null → empty and not an error, wrong kind →
error, never coerced).

**Subject requirement (D-000.50):** the collection must have **at least two rows with different
values for the same field**, and the assertion must be made against **both** rows. A single-row
fixture cannot distinguish "resolved the current row" from "resolved the only row".

**Red-proof (D-000.30 — capture before the change):** mutate the alias branch to resolve against
the data root instead of the row. AC1 must fail; state the observed failure text in the Delivery
Log.

### AC2 — An omitted `as` defaults the alias to `row` (source AC2)

**Given** a repeating region whose `as` is absent (`TableExt.As.Set == false`),
**When** its contents are resolved,
**Then** the alias is `row`, and `{{row.<field>}}` resolves to the current row's fields exactly as
AC1's declared alias does,
**And** the defaulting happens at resolution time — `TableExt.As` is still absent in the parsed
document and still absent after a round-trip.

**Red-proof:** remove the default (leave the alias empty when `As` is absent). `{{row.<field>}}`
then falls through to the data root and errors as an absent data path; AC2 reddens.

### AC3 — A row NEVER shadows the document root (source AC3)

**Given** a row scope is active, **and** the current row and the document root **both** carry a
key of the same name with **different** values,
**When** an unqualified path naming that key is resolved,
**Then** it returns the **document root's** value.

**This AC is vacuous unless both roots carry the colliding key.** A fixture where only the root has
the key passes under the reversed dispatch too and proves nothing. Assert the returned string
equals the root's value — not merely that it is not an error.

Cover the nested case as well: a two-segment unqualified path colliding at its first segment.

**Red-proof:** reorder the dispatch to try the row before the data root. AC3 must return the row's
value and fail. Record the two values and the failure text.

### AC4 — `params.` resolves from the parameter namespace and is shadowed by nothing, including a row (source AC4)

**Given** a row scope is active, **and** all three of report data, the current row, and params
carry `params.reportDate`-shaped content with **three different values** (data has a top-level
`params` object; the row has a `params` object; params itself has the key),
**When** `{{params.reportDate}}` is resolved inside the row scope,
**Then** it returns the value from the **params root**.

**Also assert the unchanged baseline case** — the same placeholder with **no** row scope active —
so the story proves it *preserved* the property rather than reimplemented it (finding 6). The
params branch stays first in the dispatch and is not rewritten.

**Red-proof:** delete the params branch so `params.reportDate` resolves from the data root. The
assertion must return the data root's value and fail. (Do **not** red-proof this by reordering the
alias branch ahead of params using an alias literally named `params` — that case is OD-1 and is
unruled.)

### AC5 — A collection binding whose path is not an array fails the render, naming the path and the element id (source AC5)

**Given** a document containing a table element whose `bind`, resolved against the report data root,
is **not** a JSON array,
**When** the document renders,
**Then** `Render` returns an error whose text contains **both** the bind path as authored **and**
the element id, and **no** PDF bytes are produced.

**All four cases are required**, and all four were measured as rendering **successfully** at
`e7f3f9c` (finding 3): the path is **absent**; the value is a **scalar**; the value is an
**object**; the value is explicit **null**.

**An empty array is NOT an error** — it is an array. Assert that a table bound to `[]` renders
successfully. Story 4.2 owns what an empty collection *renders as*.

**Resolution rule:** strip **one** trailing `[]` from `bind` if present, then resolve the remainder
as a bare dotted path against the **data root** (never params, never a row — `bind` is a
root-relative collection path, AD-11). The **authored** `bind` string appears in the error text, so
the author can find it in their document.

**Check placement:** early in `renderDocument`, after `documentBands`, before `collectImageRuns`
(finding 8). Bands in `documentBands` order, elements in declaration order, first offender
reported.

**Red-proof:** already captured — finding 3's four recorded failures at `e7f3f9c`, reproduced
verbatim in *The measurement fixture* below. Restore it as **task 1** and confirm it still
reproduces before changing any production code. If it does not reproduce, the baseline moved and
the evidence must be re-taken (do not proceed).

### AC6 — The resolution-root closure guard is widened deliberately, and AD-4's fence is re-proved

**Given** the row scope's `lookupBound` call site,
**When** `TestBindResolutionRootsAreClosed` runs,
**Then** it passes with `declaredResolutionRoots == {"data", "params", "row"}`, the root-class name
is a **string literal**, and the guard reports **three** literal call sites and zero dynamic ones;
**And** `TestBindResolutionRootsClosureRedProof` — which injects a `"page"` rootName into a scratch
copy of `text.go` — **still reddens** on the widened list, proving AD-4's fence survived (D-000.52:
a structural claim about a guard is worth exactly as much as its demonstration).

Record both runs' `t.Logf` output in the Delivery Log, not a summary of them (D-000.26).

**A guard that no longer forbids anything is not a guard:** if the widened list would accept any
root, say so and stop. It does not — `page` remains outside it, which is the whole point.

### AC7 — Byte identity of every committed golden, MEASURED

**Given** the implementation,
**When** every committed golden and the matrix documents are re-rendered,
**Then** every hash is **identical** to its value at `e7f3f9c`, and the measurement is **performed
and recorded**, not inferred from the scope fence.

No committed golden contains a rendered table (finding 4), so the expected delta is zero — but this
story edits `renderDocument`'s prologue and `internal/bind`'s dispatch, both on the emission path
for **every** document. Story 2.6a could argue non-movement from its scope fence because it edited
only test files; this story cannot.

### AC8 — The baseline is preserved and every delta is itemised

**Given** the delivered work,
**When** `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` runs,
**Then** the counts are reported against **600 / 1 / 1** all-occurrence and **368 / 1 / 1**
top-level at `e7f3f9c`, with **every** delta itemised test by test,
**And** `TestCorpusMeetsP6ExerciseFloors` is **still red** (D-000.17 / D-000.57 — it is REQUIRED to
be unmet; a green there is a regression to investigate, not a win),
**And** `TestRenderScopeFenceIgnoresTableBind`'s amendment (finding 4) is itemised **as an
amendment with its reason**, never folded into a count.

---

## The measurement fixture, verbatim

Restore this byte-for-byte as **task 1**, at `folio-go/zz_scratch_ac5_test.go`, run it, confirm it
reproduces finding 3's four failures, then proceed. It is a scratch file: it must **not** be in the
tree at handoff — its content is re-homed into the story's real AC5 test.

```go
package folio

import (
	"strings"
	"testing"
)

const zzScratchTableTpl = `{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e2", "type": "table", "x": 0, "y": 30, "bind": "transactions[]", "headerHeight": 20,
        "as": "transaction",
        "columns": [{"id": "e3", "label": "Amount", "width": 80, "bind": "{{transaction.amount}}"}]}
    ]},
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Roboto-Regular"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// AC5 red-proof: a collection binding whose path is not an array must
// fail the render, naming the path and the element id.
func TestZZScratchAC5NonArrayBindFailsRender(t *testing.T) {
	tpl, err := ParseTemplate([]byte(zzScratchTableTpl))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	for _, tc := range []struct{ name, data string }{
		{"absent", `{"customer": {"name": "Ada"}}`},
		{"scalar", `{"transactions": 7}`},
		{"object", `{"transactions": {"a": 1}}`},
		{"null", `{"transactions": null}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, rerr := Render(tpl, Data(tc.data), nil, testFontSet())
			if rerr == nil {
				t.Fatalf("AC5: render SUCCEEDED on a table bind %q that is not an array (%s)", "transactions[]", tc.name)
			}
			msg := rerr.Error()
			if !strings.Contains(msg, "transactions") || !strings.Contains(msg, "e2") {
				t.Fatalf("AC5: error must name the path and the element id; got %q", msg)
			}
			t.Logf("AC5 error: %s", msg)
		})
	}
}
```

Invocation, literally:

```
cd /Users/panitw/Projects/folio/folio-go && env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -run TestZZScratchAC5NonArrayBindFailsRender -v ."
```

---

## Tasks

1. [x] **Restore the measurement fixture** above, run it, confirm the four recorded failures reproduce
   at `e7f3f9c`. If they do not, stop and report — the baseline moved.
2. [x] **Re-measure the two guard runs** in finding 5 (add the scratch third-root file to
   `internal/bind`, run `TestBindResolutionRootsAreClosed`, remove the file, confirm clean) so the
   obstacle is observed and not merely read about.
3. [x] **AC0** — introduce the scope value in `internal/bind`, make it the single dispatch, and
   re-point `BindText`, `BindTextSpans` and `collectBandTextRuns` onto it with **no** row set.
   Run `internal/bind`'s existing suite unedited; every message must be byte-identical.
4. [x] **AC1, AC2** — the alias branch and its default, with the two-row subject; capture both
   red-proofs before the fix.
5. [x] **AC3** — the root-never-shadowed assertion with the colliding-key subject; capture its
   red-proof.
6. [x] **AC4** — the three-way collision assertion plus the no-row baseline case; capture its
   red-proof.
7. [x] **AC6** — widen `declaredResolutionRoots` to `{"data", "params", "row"}`; re-run both guard
   tests and record their `t.Logf` output.
8. [x] **AC5** — wire the collection-bind check into `renderDocument`'s prologue; re-home the scratch
   fixture into a real test with all four cases plus the empty-array case; delete the scratch file.
9. [x] **Finding 4** — amend `bindTestTemplateJSON` and `TestRenderScopeFenceIgnoresTableBind`'s doc
   comment as required; keep the column bind expression-shaped and keep the assertion strong.
10. [x] **AC7** — re-render and hash every committed golden and the matrix documents; record the
    comparison against `e7f3f9c`.
11. [x] **AC8** — full-suite run, delta itemised against both scopes, `TestCorpusMeetsP6ExerciseFloors`
    confirmed still red.
12. [x] **Record** the D-3.1.x ratification entry (*Ratification requested*) and any rulings received on
    OD-1 in `folio-mvp-decision-log.md`; update `deferred-work.md` if anything is newly deferred.
13. [x] **Story file, decision log, sprint status → `review`** (`epic-3: in-progress`,
    `3-1-…: review`).

**Stop here — do not commit, do not branch, do not set `done`.**

---

## Dev Agent Record

### Delivery Log

- **Task 1** — restored `folio-go/zz_scratch_ac5_test.go` byte-for-byte and ran it at the story's
  baseline (`a20f60e`, clean tree, D-3.1.1 already recorded but no code touched yet). All four
  cases (absent, scalar, object, null) reproduced the exact recorded failure text. Removed the
  scratch file afterwards (re-homed at task 8/9).
- **Task 2** — added a scratch file to `internal/bind` containing one
  `lookupBound(Value{}, nil, nil, "", "row", "the current row")` call, ran
  `TestBindResolutionRootsAreClosed`, and reproduced the exact recorded `t.Fatalf`/`t.Errorf`
  output (`declared resolution roots: [data params]; observed at 3 lookupBound call site(s): [data
  params row]`). Removed the scratch file, confirmed `git status --short internal/bind/` clean.
- **Task 3 (AC0)** — added `internal/bind/scope.go` (`Scope`, `NewScope`, `WithRow`). Renamed
  `BindTextSpans`'s body to `Resolve(text string, scope Scope, elementID string)` (exported: a
  future row-scoped caller, Story 4.2, needs it from package `folio`) and made `BindTextSpans` a
  thin wrapper: `return Resolve(text, NewScope(data, params), elementID)`. Ran
  `internal/bind`'s existing suite unedited — every test passed, every message byte-identical
  (verified by full `-v` run, no text_test.go edits).
- **Tasks 4–6 (AC1–AC4)** — added the row branch to `Resolve`'s dispatch, positioned after the
  `params` branch and before the final data-root lookup (D-3.1.1's mandated order): `if
  scope.rowSet && path[0] == scope.rowAlias { … lookupBound(scope.row, path[1:], path, elementID,
  "row", "the current row") … }`. The literal `"row"` is passed as `rootName`; `scope.rowAlias`
  (the author's own spelling) is carried only for the branch condition and left inside `fullPath`
  for error text. New tests in `internal/bind/scope_test.go` (package `bind`, calling `Resolve`
  directly) and `folio-go/row_scope_test.go` (package `folio`, exercising the real
  `resolvedRowAlias` default function through the exported `bind.Resolve`/`bind.NewScope` API).
- **Task 7 (AC6)** — widened `declaredResolutionRoots` to `{"data", "params", "row"}`. Re-ran both
  guards; `t.Logf` output captured verbatim below.
- **Task 8 (AC5)** — added `resolvedRowAlias` and `checkTableBindings` to `render.go`, wired into
  `renderDocument` immediately after `documentBands`, before `collectImageRuns`. Re-homed the
  scratch fixture into `folio-go/render_table_bind_test.go`
  (`TestRenderTableBindNonArrayFailsRender`, all four cases, plus
  `TestRenderTableBindEmptyArrayIsNotAnError`), then deleted `zz_scratch_ac5_test.go`. Also added
  `TestRenderTableRowAliasCollidesWithReservedNameFailsRender` there, proving D-3.1.1's ruling on
  the creator's OD-1 (Arm B: a located template error, data-free — the fixture supplies `{}` and
  the alias failure is still reported, never the collection-bind failure).
- **Task 9 (finding 4)** — amended `bindTestTemplateJSON` in `render_bind_test.go`:
  `table.bind` changed from `"{{transactions[]}}"` to the bare collection path `"transactions[]"`,
  data amended to supply a real `transactions` array; `columns[0].bind` left byte-for-byte
  expression-shaped. Amended `TestRenderScopeFenceIgnoresTableBind`'s doc comment to say the fence
  it now proves is the **column** fence, and that `table.bind` is read as a collection path from
  this story on. No other change to the test's assertions.
- **Task 10 (AC7)** — ran the full suite (`go test -count=1 -v ./...`) and confirmed every
  golden/digest/byte-comparison test passes: `TestRenderMatchesGoldenFixture`,
  `TestRenderMatchesFontTextGoldenFixture`, `TestRenderMatchesImageEmbedGoldenFixture`,
  `TestMultiScriptFallbackGoldenFixture`, `TestMultiPageGoldenFixtureMatchesTheInRepoTemplate`,
  `TestMultiPageGoldenMatchesTheCommittedArtifact`, `TestShapedTextGoldenFixture`,
  `TestThreeBandPageGoldenFixture`, `TestWrappedTextGoldenFixture`,
  `TestGoldenDigestAgreesAtEveryDeclaredSite`, `TestBoundaryGateDigestsAreWellFormed`,
  `TestEveryGoldenPDFResolvesItsPageTree` (all subtests), `TestSerializeIDMatchesReDerivedDigest`,
  `TestWorkedExampleMatchesGoldenFixture`, `TestAssetExampleMatchesGoldenFragment`,
  `TestFirstBaselineSemanticAcceptanceAcrossEveryReRecordedGolden` (all subtests). No committed
  golden contains a rendered table (finding 4 confirmed no artifact reaches the new render-time
  checks except `worked-example.json`, which is only round-tripped/fence-compared, never
  rendered) — every hash matched, zero delta, measured not inferred.
- **Task 11 (AC8)** — full-suite run reported below.
- **Task 12** — D-3.1.1 (the row-resolution-root ruling) and D-000.59 were already recorded in
  `folio-mvp-decision-log.md` by the orchestrator before this story's code was touched (visible in
  the working tree at story start). OD-1 is resolved as part of D-3.1.1 (Arm B). Nothing new was
  deferred by this story; `deferred-work.md` unchanged.
- **Task 13** — this update.

### Red-proofs captured

All captured by temporarily mutating `internal/bind/text.go` / `folio-go/render.go`, confirming
the test(s) redden with the exact reasoning the story specifies, then restoring the file from a
saved copy (never `git checkout`) and re-confirming green.

- **AC1** (`TestScopeRowAliasResolvesCurrentRow`): mutated the row branch to resolve against
  `scope.data` instead of `scope.row`. Observed failure: `row1: unexpected error: bind: element
  e2: row path "transaction.amount" is absent from the current row` (the mutated lookup ran
  against an empty `data` root that has no `transaction`/`amount` path).
- **AC2** (`TestRowAliasDefaultsToRowAtResolutionTime`): mutated `resolvedRowAlias` to return `""`
  instead of `"row"` for an absent `as`. Observed failure: `AC2: an absent "as" must default the
  alias to "row", got ""`.
- **AC3** (`TestScopeRowNeverShadowsDocumentRoot`): inserted a mutation immediately before the
  final data-root lookup that checks the row FIRST for any unqualified path (`if scope.rowSet { if
  _, p := scope.row.Lookup(path); p != Absent { … resolve from row … } }`), reproducing "a row
  shadows the root". Observed failure: `AC3: an unqualified colliding key must resolve to the
  DOCUMENT ROOT's value, got "FROM-ROW" (want "FROM-ROOT")`.
- **AC4** (`TestScopeParamsUnshadowableByRow`): disabled the params branch (`if false &&
  path[0] == "params"`), forcing `params.reportDate` to fall through to the data root. Observed
  failure on the UNCHANGED baseline (no row) case, as the story requires: `AC4 baseline:
  {{params.reportDate}} with no row active must still resolve from params, got "FROM-DATA"`.
- **AC5** (finding 3, inherited/reconfirmed): all four non-array cases (absent, scalar, object,
  null) reproduced rendering SUCCESSFULLY at baseline (task 1's log above); after AC5's wiring,
  all four fail with the correct located error (task 8).
- **AC6** (finding 5, inherited/reconfirmed): the scratch third-root file reproduced
  `TestBindResolutionRootsAreClosed`'s exact recorded failure at baseline (task 2's log above).

### AC6 guard runs, `t.Logf` output verbatim (post-widening)

```
=== RUN   TestBindResolutionRootsAreClosed
    resolution_roots_arch_test.go:197: declared resolution roots: [data params row]; observed at 3 lookupBound call site(s): [data params row]
--- PASS: TestBindResolutionRootsAreClosed (0.00s)
=== RUN   TestBindResolutionRootsClosureRedProof
    resolution_roots_arch_test.go:244: red-proof: injecting lookupBound(..., "page", ...) is observed as [params row data page], outside declared [data params row] — the guard reddens
--- PASS: TestBindResolutionRootsClosureRedProof (0.00s)
```

Three literal `lookupBound` call sites observed (`params`, `row`, `data`), zero dynamic — the
guard's non-literal-`rootName` `Fatalf` never fired. `TestBindResolutionRootsClosureRedProof`
still reddens on the widened list: `page` remains outside `{data, params, row}`, proving AD-4's
fence survived the widening.

### Deltas against baseline

Measured through `env CGO_ENABLED=0 GOWORK=off rtk proxy "go test -count=1 -v ./..."` inside
`folio-go`, package scope `./...`:

| Scope | Baseline (`e7f3f9c`) | This story | Delta |
|---|---|---|---|
| all-occurrence (`--- PASS` anywhere, subtests included) | 600 / 1 / 1 | **617 / 1 / 1** | +17 PASS |
| top-level (`^--- PASS`) | 368 / 1 / 1 | **378 / 1 / 1** | +10 PASS |

The one FAIL is still `TestCorpusMeetsP6ExerciseFloors` (REQUIRED red, D-000.17/D-000.57), with
**byte-identical** stats: `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — same
`corpus_test.go:189` message (`P6g (opaque names) floor not met: got 7, need >=20`). The one SKIP
is still `TestXrefEntriesRejectsMalformedSubprocess`, unchanged.

**Every delta, itemised** (10 new top-level tests, +7 more at all-occurrence from subtests = +17):

1. `TestRenderTableBindNonArrayFailsRender` — new — AC5. +1 top-level +4 subtests
   (`absent`/`scalar`/`object`/`null`) = 5.
2. `TestRenderTableBindEmptyArrayIsNotAnError` — new — AC5's empty-array carve-out. +1.
3. `TestRenderTableRowAliasCollidesWithReservedNameFailsRender` — new — D-3.1.1's OD-1 ruling. +1
   top-level +3 subtests (`params`/`page`/`pages`) = 4.
4. `TestScopeRowAliasResolvesCurrentRow` — new — AC1. +1.
5. `TestScopeRowAliasFieldAbsentIsLocatedError` — new — AC1's AD-14 clause. +1.
6. `TestScopeRowAliasDefaultsToRow` — new — AC2 (bind-package level). +1.
7. `TestScopeRowNeverShadowsDocumentRoot` — new — AC3. +1.
8. `TestScopeParamsUnshadowableByRow` — new — AC4. +1.
9. `TestRowAliasDefaultsToRowAtResolutionTime` — new — AC2 through the real production
   `resolvedRowAlias` default and the exported `bind.Resolve` API. +1.
10. `TestRowAliasHonoursDeclaredAs` — new — AC1's counterpart to #9. +1.

Total new: 10 top-level + 7 subtests = 17, matching the all-occurrence delta exactly.

**`TestRenderScopeFenceIgnoresTableBind` — AMENDMENT, not a new test** (finding 4, itemised
separately per AC8's own requirement, never folded into the count above): `table.bind` changed
from the expression-shaped `"{{transactions[]}}"` (not a collection path, and the test's data
supplied no `transactions`) to the bare collection path `"transactions[]"` with a real array
supplied; `columns[0].bind` stays byte-for-byte expression-shaped. The test's own doc comment now
says the fence it proves is the **column** fence. Still passes.

**`go vet ./...`**: clean, no findings.

**Lint suite** (`lint` module, `go test -count=1 ./...`): `TestManifestUpToDate`,
`TestResolveAssetsIncludesWordlist` and `TestFontsAssetsNoticeRemovalRedProof` fail — all three
pre-existing, unrelated to this story (`.font-sources` missing a `LICENSE*` file), confirmed
present identically before any of this story's code changes were applied. Every rule this story's
own scope fence touches is green: `TestAbsencesProductionScan` (confirms `internal/expr` and
`internal/diag` are still absent — this story created neither), `TestFloatTypedProductionScan`
(this story introduces no numeric type), `TestForbiddenImportsProductionScan`,
`TestStageRankProductionScan` (no new import-rank violation). D-000.4: this story is not a matrix
override; the cross-target matrix stays WRITTEN-not-run, due at Epic 3's close.

**Lint suite wording — corrected by the finisher (Review Finding 4).** The paragraph above calls
the three failures "pre-existing, unrelated to this story," which is true but incomplete in a way
that misleads a later reader checking a clean checkout. The reviewer traced them to `/.font-sources/`
— gitignored at `.gitignore:85`, three untracked font binaries with no `LICENSE*`/`NOTICE*` file,
present only on this working machine. The finisher independently re-verified: `.font-sources/` is
untracked (`git status --short --untracked-files=all -- .font-sources` empty of any tracked entry,
`git check-ignore -v` confirms `.gitignore:85`), and re-running the lint suite reproduces the exact
same three failures, all bottoming out in `.font-sources: contains a committed font binary but no
LICENSE* file (AC25, AD-26)`. A fresh clone or CI checkout carries no `.font-sources/` directory at
all, so the lint suite is green there — this is local machine state, not repository state, and no
code or repository change addresses it.

### Deltas against baseline (finisher re-measurement, post-fix)

Measured through `env CGO_ENABLED=0 GOWORK=off go test -count=1 -json ./...` inside `folio-go`,
package scope `./...`, parsed via `test2json` action fields (not a summary line):

| Scope | Baseline (`72cfc6a`) | Developer handoff | This finisher's fixes | Delta (finisher → baseline) |
|---|---|---|---|---|
| all-occurrence | 600 / 1 / 1 | 617 / 1 / 1 | **619 / 1 / 1** | +19 PASS |
| top-level | 368 / 1 / 1 | 378 / 1 / 1 | **380 / 1 / 1** | +12 PASS |

The finisher added exactly 2 new top-level tests, 0 subtests, accounting for the +2/+2 over the
developer's handoff numbers:

1. `TestScopeParamsUnshadowableEvenByRowAliasedParams` (`internal/bind/scope_test.go`) — Finding 2's
   fix. +1 top-level, +1 all-occurrence.
2. `TestScopeRowAliasBareIsLocatedError` (`internal/bind/scope_test.go`) — Finding 3's fix. +1
   top-level, +1 all-occurrence.

Finding 1's fix strengthened an existing test's assertion (`TestRenderTableRowAliasCollidesWithReservedNameFailsRender`)
without adding a test or subtest — 0 delta from that fix.

`TestCorpusMeetsP6ExerciseFloors` is still the single FAIL, re-confirmed **byte-identical**:
`P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` (D-000.17/D-000.57 — still REQUIRED
red). The one SKIP is still `TestXrefEntriesRejectsMalformedSubprocess`.

**`go build ./...` and `go vet ./...`**: both clean, re-run after the fixes.

**AC7 re-verified after the fixes**: `git status --short --untracked-files=all` on
`folio-go/testdata/` and `folio-go/fixtures/` is empty (no golden modified or added); every
golden/digest test (`TestRenderMatchesGoldenFixture`, `TestGoldenDigestAgreesAtEveryDeclaredSite`,
`TestBoundaryGateDigestsAreWellFormed`, `TestEveryGoldenPDFResolvesItsPageTree` and its 9 subtests,
and the rest of the AC7 set the developer named) passes. Zero byte movement, measured after the
finisher's edits — not merely inferred from "the fixes only touched test files."

**Frozen-function check re-verified**: `parseBindingPath` and `isValidIdent`, extracted from
`72cfc6a` and from the post-fix working tree, hash SHA-256 identical (`d421af844648f219…`,
`77e4d9548f2a63c3…`) — matching the reviewer's recorded values exactly. No production code was
touched by any of the three fixes; the finisher's diff is test-file-only.

**Cross-target matrix**: written-not-run for this story, due at Epic 3's close (D-000.4). The
finisher's fixes are test-only and not matrix-facing, so this status is unchanged by the finish
pass.

### File List

- `folio-go/internal/bind/scope.go` — new — `Scope`, `NewScope`, `WithRow`.
- `folio-go/internal/bind/scope_test.go` — new (dev), modified (finisher) — AC1-AC4 tests via
  `bind.Resolve` directly; finisher added `TestScopeParamsUnshadowableEvenByRowAliasedParams`
  (Finding 2) and `TestScopeRowAliasBareIsLocatedError` (Finding 3).
- `folio-go/internal/bind/text.go` — modified — `declaredResolutionRoots` widened to
  `{"data", "params", "row"}`; `BindTextSpans`'s body renamed/exported as `Resolve(text string,
  scope Scope, elementID string)`; `BindTextSpans` now a thin wrapper; row-scope dispatch branch
  added between the `params` branch and the final data-root lookup.
- `folio-go/render.go` — modified — `resolvedRowAlias` and `checkTableBindings` added; the latter
  wired into `renderDocument` right after `documentBands`, before `collectImageRuns`.
- `folio-go/render_table_bind_test.go` — new (dev), modified (finisher) — AC5 (re-homed from the
  scratch fixture) plus the empty-array carve-out and D-3.1.1's OD-1 alias-collision red-proof;
  finisher strengthened `TestRenderTableRowAliasCollidesWithReservedNameFailsRender`'s assertion to
  discriminate the interpolated alias from the message's static text (Finding 1).
- `folio-go/render_bind_test.go` — modified — `bindTestTemplateJSON`'s `table.bind` amended to a
  bare collection path with a real array supplied (finding 4);
  `TestRenderScopeFenceIgnoresTableBind`'s doc comment amended to describe the column fence.
- `folio-go/row_scope_test.go` — new — AC1/AC2 exercised through the real `resolvedRowAlias`
  default and the exported `bind.Resolve`/`bind.NewScope` API.
- `folio-go/zz_scratch_ac5_test.go` — added then removed (task 1's fixture, re-homed at task 8/9;
  not present at handoff).
- `_bmad-output/implementation-artifacts/3-1-bind-a-repeating-region-to-a-collection-with-an-explicit-row.md`
  — this file — task checkboxes, Status, Dev Agent Record, Review Findings, Finding Resolutions
  (finisher), plain-terms opener rewritten to state the result.
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — `epic-3: backlog` → `in-progress`
  (dev); `3-1-…: ready-for-dev` → `review` (dev) → `done` (finisher). `epic-3` left `in-progress`
  — the epic gate is not this story's to flip.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-3.1.1/D-000.59 (and
  three further Epic 3 rulings) already recorded by the orchestrator before this story's code was
  touched; unedited by this story's own work or by the finisher (no new decision was needed to
  dispose of any review finding).

### Change Log

- 2026-08-25 — Story 3.1 implemented: row-scope resolution (`internal/bind.Scope`) as a third
  resolution root (`"row"`, D-3.1.1), the declared-alias/default-`row` dispatch (AC1/AC2), the
  root-never-shadowed and params-unshadowable properties proved under a row scope (AC3/AC4), the
  collection-bind not-an-array render check (AC5), and D-3.1.1's ruling on the creator's OD-1 (an
  alias colliding with `params`/`page`/`pages` is a located template error). No row generation, no
  expression evaluation, no aggregates — all deferred to their owning stories per the scope fence.
- 2026-08-25 — Story 3.1 finished: adversarial review returned 1 Major, 2 Minor, 1 Nit. All four
  resolved — three fixed with a captured red-proof each (a test whose assertion could not
  discriminate a correct value from a wrong one, strengthened and confirmed to redden on the
  reviewer's own mutation; two coverage gaps closed with new tests, each verified against the
  specific defect it targets), one dismissed with independently re-verified evidence (a lint
  wording issue traced to gitignored local machine state, not repository state). No production
  code changed — every fix is test-file-only; `parseBindingPath`/`isValidIdent` re-verified
  SHA-256-identical to `72cfc6a`. Full suite, `go vet`, `go build`, and the lint suite re-run;
  `TestCorpusMeetsP6ExerciseFloors` re-confirmed the single REQUIRED red with byte-identical stats;
  AC7 byte-identity re-verified. Status → `done`.

---

## Review Findings

## Review Summary
- **Reviewed by:** bmad-code-reviewer (adversarial, context isolated from the developer's)
- **Date:** 2026-08-25
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 0
- **Majors:** 1
- **Minors:** 2
- **Nits:** 1

Every number below was re-measured by the reviewer, not read from the Dev Agent Record. The
working tree was mutated for red-proofs and restored **by hand** from saved copies (never
`git checkout`); `render.go`, `internal/bind/text.go` and `internal/bind/scope_test.go` were
re-verified SHA-256-identical afterwards and `git status --short --untracked-files=all` matches
the pre-review snapshot exactly.

### Independently verified — the reported measurements are accurate

- **Counts confirmed.** Parsed from `go test -json` (test2json), not from a summary line.
  Working tree: **617 PASS / 1 FAIL / 1 SKIP** all-occurrence, **378 / 1 / 1** top-level.
  Baseline re-run by the reviewer in a **clean `git worktree` at `72cfc6a`**: **600 / 1 / 1** and
  **368 / 1 / 1**. Delta **+17 / +10**, **zero tests removed**. The developer's itemisation of the
  10 new top-level tests matches the measured added-name set exactly.
- **CRITICAL INVARIANT HOLDS.** `TestCorpusMeetsP6ExerciseFloors` is **still the single FAIL**, with
  stats **byte-identical**: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`,
  same `corpus_test.go:189` message. No floor was filled. The single SKIP is still
  `TestXrefEntriesRejectsMalformedSubprocess`.
- **D-3.1.1's freeze holds.** `parseBindingPath` and `isValidIdent` extracted from `72cfc6a` and
  from the working tree and hashed: **SHA-256 identical** (`d421af844648f219`, `77e4d9548f2a63c3`).
  No new accepted character or shape; **no second path matcher** anywhere; **no** function-call
  special-case; `reservedPlaceholders` untouched (the alias was **not** added to it, per D-1.7.4);
  the `rootName` passed to `lookupBound` is the **literal** `"row"` and `scope.rowAlias` is used
  only for the branch condition and error text.
- **AC6 re-run by the reviewer.** `TestBindResolutionRootsAreClosed` passes with
  `[data params row]`, **3 literal call sites, 0 dynamic**;
  `TestBindResolutionRootsClosureRedProof` still reddens on an injected `"page"`. AD-4's fence
  survived the widening.
- **The known guard weakness was NOT walked into.** The Story 2.5 review (Blocker 2) recorded that
  `TestBindResolutionRootsAreClosed` can be defeated by a dispatch that early-returns without ever
  calling `lookupBound`. The row branch's main arm (`len(path) >= 2`) **does** call
  `lookupBound(scope.row, …, "row", …)`, and the guard demonstrably observes it. The closed-set
  guard is **not** green-while-blind. (One narrow sub-case does early-return — see Finding 3.)
- **AD-23.** No `float64`, `float32`, `big.Float` or `big.Rat` introduced in any changed or new
  file (checked by eye, since D-3.1a.1 records the guard cannot see `math/big.Float`).
- **`go vet ./...` clean; `go build ./...` clean** (both re-run).
- **AC7 substantiated.** No file under `folio-go/testdata/` or `folio-go/fixtures/` is modified or
  added (`git status --untracked-files=all` on those trees is empty), and every golden/digest test
  passes. Byte identity is measured, not inferred.
- **Check placement is correct and universal.** `checkTableBindings` sits immediately after
  `documentBands` and before `pageGeometryOf`/`collectImageRuns`, as finding 8 requires.
  `renderDocument` has exactly **one** call site (`render_entry.go:172`), through which **both**
  `Render` and `RenderTo` pass — the check runs on every path that renders.
- **The `render_bind_test.go` amendment did NOT weaken the fence.** Reviewer probe: feeding
  `columns[0].bind`'s value through `bind.BindTextSpans` still errors
  (`"formatNumber(transaction" is not a valid path segment`), so D-1.6.8's **column** fence retains
  its teeth. The table half of the old claim is genuinely gone, and the amended doc comment says
  so explicitly rather than papering over it. Assertion strength unchanged. No silent coverage loss.
- **Degenerate collection binds are safe.** Reviewer probe of `""`, `"[]"`, `"transactions[][]"`,
  `"transactions[].payee"`, `"."`, `"a..b"`: all produce located errors; none silently passes.

---

### Finding 1: The OD-1 alias-collision test cannot tell a correct alias from a wrong one
- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/render_table_bind_test.go:133-136` (assertion) against
  `folio-go/render.go:280-286` (message construction)
- **Observation**: The located error's **static** explanatory text enumerates all three reserved
  names at once — `"params" can be shadowed by nothing (AD-11) and "page"/"pages" never acquire a
  namespace (AD-4)`. The test asserts only `strings.Contains(msg, alias)`. Because every one of
  `params`, `page`, `pages` appears in the constant text regardless of what `%q` interpolates, the
  assertion is satisfied by the constant, not by the reported value.
  **Red-proof captured by the reviewer**: replacing the interpolated `alias` argument with the
  literal `"zzzWRONGzzz"` left **all three subtests PASSING**, logging
  `table's row alias "zzzWRONGzzz" collides with a reserved name — "params" can be shadowed by …`.
  A separate mutation disabling the alias check entirely **does** redden all three, so the check's
  existence and its precedence over the collection-bind check are genuinely proved — it is
  specifically the "names the colliding alias" claim that is unfalsifiable.
- **Impact**: The Delivery Log (task 8) and the Change Log present this test as the red-proof of
  D-3.1.1's OD-1 Arm B ruling. Arm B's literal wording is "a **located** template error naming the
  **element**", and the element half (`e2`, sourced only from `el.ID`) does hold. But the test
  advertises a stronger claim than it can support, and a future regression that reported the wrong
  alias — or a constant one — would ship green. This is the project's own "a test that cannot fail
  is not a test" rule applied to one of the two assertions in a new red-proof.
- **Suggested Resolution**: Assert on the *interpolated* region rather than on any substring of the
  whole message — e.g. require `msg` to contain `row alias "` + alias + `"` — or stop enumerating
  all three reserved names in the constant text. Then re-run the `"zzzWRONGzzz"` mutation above and
  confirm it reddens. **Not fixed by the reviewer.**
- **Related AC**: OD-1 / D-3.1.1 (Arm B); AC5's neighbouring check

### Finding 2: AC4's mandated third value is inert — the fixture cannot discriminate params-vs-row precedence
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/bind/scope_test.go:115-137`
  (`TestScopeParamsUnshadowableByRow`)
- **Observation**: AC4's subject requirement is that report data, the current row **and** params
  each carry a *different* `params.reportDate` value. The fixture supplies all three, but the row
  is registered under the alias `"transaction"`, so `{{params.reportDate}}` can never select the
  row branch under **any** dispatch order — `path[0]` is `"params"`, which never equals
  `"transaction"`. The row's `FROM-ROW` value is therefore unreachable by construction.
  **Red-proof captured by the reviewer**: replacing the row fixture with `{}` (no `params` key at
  all) left the test **PASSING**. The third value contributes nothing falsifiable.
- **Impact**: The property AC4 asserts *does* hold — a reviewer probe with the one input that could
  falsify it, `WithRow(row, "params")`, returned `FROM-PARAMS` — so this is a coverage gap, not a
  bug. But the test discriminates only params-vs-**data** (which is the pre-existing, pre-3.1
  behaviour), not params-vs-**row**, which is the clause this story added. The subject requirement
  is met in letter and not in substance. Worth noting that the alias ban lives in `render.go`'s
  `checkTableBindings` while `bind.Scope.WithRow` accepts any alias string, so inside package
  `bind` the branch **order** is the only thing protecting `params`.
- **Suggested Resolution**: Either add the discriminating case at bind level
  (`WithRow(row, "params")`, assert `FROM-PARAMS`) — the story forbade exactly this while OD-1 was
  unruled, but OD-1 has since been ruled Arm B, so that prohibition has lapsed and was never
  revisited — or record explicitly in the story that the row's value is inert and that the property
  rests on branch order plus the render-level alias ban. **Not fixed by the reviewer.**
- **Related AC**: AC4

### Finding 3: The new bare-row-alias error path has no test and no AC
- **Severity**: Minor
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/internal/bind/text.go:242-245`
- **Observation**: The row branch returns early when `len(path) == 1`, emitting
  `bind: element e2: "transaction" is a namespace, not a value` (reviewer-confirmed by probe). No
  test in the repository exercises it: the only test referencing that message
  (`internal/bind/text_test.go:296-307`) covers the **params** twin at `text.go:218`. The behaviour
  is not required by any of AC0–AC8 — the developer added it by analogy to AC17 — and it is not
  listed among the 10 itemised new tests.
- **Impact**: A new production error path ships unasserted. Separately, this sub-case is an early
  return that never reaches `lookupBound` — the same shape the Story 2.5 review flagged as the
  closed-set guard's blind spot. It does **not** defeat the guard here (the `len(path) >= 2` arm
  supplies the observed `"row"` root, independently confirmed), but the story should say that
  rather than let the guard's green imply the whole branch is covered.
- **Suggested Resolution**: Add a one-line assertion mirroring `text_test.go:296-307` for the row
  alias, or record the behaviour in the story as deliberate and knowingly unasserted. **Not fixed
  by the reviewer.**
- **Related AC**: AC1/AC2 (the row branch), AC6 (guard reach)

### Finding 4: Three lint failures are attributed to the repository but come from gitignored local state
- **Severity**: Nit
- **Category**: Maintainability (record accuracy)
- **Location**: Dev Agent Record, "Lint suite" paragraph
- **Observation**: The record says `TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist` and
  `TestFontsAssetsNoticeRemovalRedProof` fail and are "pre-existing, unrelated to this story". The
  "unrelated to this story" half is **correct and confirmed**. But the reviewer ran the lint module
  in a clean worktree at `72cfc6a`, where it passes **85/85, exit 0**. The three failures are
  produced by `/.font-sources/` — gitignored at `.gitignore:85`, three untracked font binaries with
  no `LICENSE*`/`NOTICE*` file, present only in this working copy and absent from any fresh clone
  or CI checkout.
- **Impact**: A later reader reconciling this record against a clean checkout will find a green
  lint suite and be unable to reproduce three "pre-existing" failures, or will wrongly conclude the
  repository carries three broken lint tests.
- **Suggested Resolution**: Reword to name the cause — untracked local `.font-sources/` — and note
  that the lint suite is green on a clean tree. **Not fixed by the reviewer.**
- **Related AC**: AC8 (delta reporting)

### AC-by-AC disposition

| AC | Disposition |
|---|---|
| **AC0** | **Satisfied.** One dispatch (`Resolve`); `BindText`/`BindTextSpans` are thin wrappers with no row; `internal/bind`'s existing suite passes **unedited** (`text_test.go` untouched — confirmed by diff) and every message is byte-identical. |
| **AC1** | **Satisfied.** Two rows, different values for the same field, asserted against **both**, plus an explicit `got1 == got2` guard that fails if the current row was not consulted. AD-14's absent-field clause asserted with a located error naming the **author's** alias. |
| **AC2** | **Satisfied.** Default applied at resolution time via the real production `resolvedRowAlias`, exercised with a genuinely absent `Presence`. `TableExt.As` is not defaulted at load; `parse_bands.go`/`model.go` unchanged; round-trip fixed point undisturbed. |
| **AC3** | **Satisfied.** Both roots carry the colliding key with different values; the assertion is on the returned string equalling the **root's** value, and the nested two-segment case is covered. Non-vacuous. |
| **AC4** | **Satisfied in substance, weakly proved — see Finding 2.** The property holds (reviewer-probed); the no-row baseline case is asserted as required; but the row's value in the fixture is inert. |
| **AC5** | **Satisfied.** All four non-array shapes plus the empty-array carve-out, asserted on the returned error's **text** against the bind *as authored* and the element id (D-000.13/D-000.21 respected — no bare exit status, no intermediate struct). Reviewer confirmed all four were successful renders at baseline and are located errors now. |
| **AC6** | **Satisfied.** Independently re-run: `[data params row]`, 3 literal sites, 0 dynamic; the `"page"` closure red-proof still reddens. |
| **AC7** | **Satisfied.** Zero byte movement, measured: no golden/fixture file modified or added, every golden/digest test green. |
| **AC8** | **Satisfied, with Finding 4 on the lint wording.** Counts, both scopes, the REQUIRED red and its byte-identical stats, the +17/+10 delta, and the `TestRenderScopeFenceIgnoresTableBind` amendment-not-a-new-test itemisation all independently reproduced. |

### Scope-fence compliance (all confirmed clean)

`internal/expr` and `internal/diag` were **not** created; **no absence rule was deleted or edited**
(`TestAbsencesProductionScan` green); the path grammar was not extended; no aggregates, no locale
formatting, no `visibleIf`, no CLI; no new load-time rule; `page`/`pages` remain reserved tokens
resolved from neither root (reviewer probe: `{{page}}` under an active row aliased `"page"` still
passes through literally as `{{page}}`); the row scope carries `bind.Value` with literal number
text intact, so Story 3.3's exact-decimal option is preserved. The cross-target matrix was **not**
run (written-not-run, due at Epic 3's close, per D-000.4); the new code is not matrix-facing.

---

## Finding Resolutions (finisher)

Triage and disposition of each finding in *Review Findings* above, applied under orchestrator
direction. All four are closed; nothing is deferred.

### Finding 1 (Major) — the OD-1 alias-collision test cannot tell a correct alias from a wrong one

**Decision: FIX.**

**Rationale:** The reviewer's own red-proof (substituting the literal `"zzzWRONGzzz"` for the
interpolated alias at `render.go:285` left all three subtests green) proves the assertion is
satisfied by the message's static enumeration, not by the interpolated value. This test is presented
in the Delivery Log as the red-proof of D-3.1.1's OD-1 ruling, which requires the message to name
**the colliding alias** (AD-14/AC5's located-error discipline) — a test offered as that proof must be
able to fail on that specific claim, per D-000.9.

**Fix applied:** `folio-go/render_table_bind_test.go`, `TestRenderTableRowAliasCollidesWithReservedNameFailsRender` —
the assertion now requires the exact substring `row alias "<alias>" collides`, which only a correct
interpolation of `alias` can produce; the static enumeration of `params`/`page`/`pages` never
contains this substring verbatim for a wrong value.

**Red-proof (re-run, D-000.30):** restored the reviewer's own mutation — hand-edited `render.go`'s
`checkTableBindings` to pass the literal `"zzzWRONGzzz"` instead of the interpolated `alias` as the
`%q` argument, ran the test, and confirmed all three subtests (`params`/`page`/`pages`) now **FAIL**:

```
render_table_bind_test.go:148: error must name the colliding alias via interpolation
  (want "row alias \"params\" collides");
  got "folio: Render: element e2: table's row alias \"zzzWRONGzzz\" collides with a reserved name — …"
```

`render.go` was then restored from a saved copy (never `git checkout`) and verified SHA-256
identical to its pre-mutation state; the tightened test passes green again on the restored file.

### Finding 2 (Minor) — AC4's mandated third value is inert

**Decision: FIX.**

**Rationale:** The reviewer's probe (`WithRow(row, "params")` → `FROM-PARAMS`) confirmed the
property AC4 asserts does hold; the gap is that the *existing* fixture (row aliased `"transaction"`)
can never select the row branch when resolving `{{params.reportDate}}`, so it discriminates only
params-vs-**data** (pre-existing, pre-3.1 behaviour), not params-vs-**row** (the clause this story
added. The story's own text forbade constructing the discriminating case while OD-1 was unruled;
OD-1 is now ruled (D-3.1.1, Arm B), so that prohibition has lapsed, exactly as the orchestrator's
direction notes.

**Fix applied:** `folio-go/internal/bind/scope_test.go`, new test
`TestScopeParamsUnshadowableEvenByRowAliasedParams` — constructed at bind level (below
`render.go`'s `checkTableBindings` alias ban, which does not exist inside package `bind`) with a row
aliased literally `"params"`, proving that dispatch **order** — not the render-level alias ban — is
what protects the params root from a row inside package `bind` itself.

**Red-proof:** hand-swapped the order of the `params` and row dispatch branches in
`internal/bind/text.go` (row checked before params) and ran the new test: it failed, returning
`"FROM-ROW"` where `"FROM-PARAMS"` was required — confirming the test discriminates precedence, not
merely presence. The existing `TestScopeParamsUnshadowableByRow` stayed green under the same
mutation, confirming *that* test genuinely cannot see this defect (as Finding 2 observed). `text.go`
was restored from a saved copy and verified SHA-256 identical to its pre-mutation state.

### Finding 3 (Minor) — the new bare-row-alias error path has no test and no AC

**Decision: FIX.**

**Rationale:** A new production error path (a bare row alias, e.g. `{{transaction}}` with no dot)
shipped unasserted. The reviewer additionally flagged that this is an early return that never
reaches `lookupBound` — the shape Story 2.5's review flagged as able to defeat
`TestBindResolutionRootsAreClosed`'s closed-set guard.

**Fix applied:** `folio-go/internal/bind/scope_test.go`, new test
`TestScopeRowAliasBareIsLocatedError`, mirroring the existing params twin
(`TestBindTextParamsBareIsLocatedError`). Its doc comment states explicitly, per the orchestrator's
direction, that this early return does **not** defeat the closed-set guard here: the row branch's
other arm (`len(path) >= 2`) independently supplies the literal `"row"` rootName to `lookupBound`,
so the guard's `[data params row]` observation does not depend on the bare-alias sub-case ever
executing — it was simply unasserted by any AC0–AC8 obligation, now closed.

**Red-proof:** disabled the bare-alias early return in `internal/bind/text.go` (`if false && len(path)
== 1`) and ran the new test: it failed — the call fell through to `lookupBound` on an empty
sub-path and produced a different (and still erroneous) message, `"…is a object, not a string —
text bindings are never coerced"`, confirming the assertion actually depends on the early return
existing. `text.go` was restored from a saved copy and verified SHA-256 identical to its
pre-mutation state.

### Finding 4 (Nit) — three lint failures attributed to the repository actually come from gitignored local state

**Decision: DISMISS.**

**Rationale:** Independently re-verified: `.font-sources/` is gitignored (`.gitignore:85`) and
`git status --short --untracked-files=all -- .font-sources` shows it untracked; running the lint
module here reproduces the reviewer's exact three failures
(`TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`, `TestFontsAssetsNoticeRemovalRedProof`),
all tracing to `.font-sources: contains a committed font binary but no LICENSE* file`. This is local
machine state, not repository state — a fresh clone or CI checkout will not carry `.font-sources/` at
all. The reviewer's own clean-worktree run at `72cfc6a` (85/85, exit 0) is taken as established; the
finisher did not re-run a second clean-worktree check, since the untracked-file evidence above is
independently sufficient to explain the divergence without one. No fix is applicable — "fixing" a
gitignored local directory would not change repository state and is out of scope. The Dev Agent
Record's "Lint suite" paragraph is corrected below to name the cause.

**Record correction:** see the "Lint suite wording — corrected by the finisher" paragraph in the
Dev Agent Record's *Deltas against baseline* section above, which replaces the "pre-existing,
unrelated to this story" wording with the untracked `.font-sources/` cause and the clean-tree
comparison.
