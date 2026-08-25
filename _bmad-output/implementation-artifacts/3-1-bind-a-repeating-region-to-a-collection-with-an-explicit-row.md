# Story 3.1: Bind a repeating region to a collection with an explicit row scope

Status: **ready-for-dev**

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
— one row per transaction. The moment there are rows, a question appears that the product has never
answered in code: when the author writes a field name inside that region, which transaction does it
mean, and what happens to names belonging to the document as a whole rather than to any one row?

Guessing is the failure this story exists to prevent. A tool that infers the row's name from the
collection's name, behind the author's back, will eventually disagree with the designer showing the
author something else, and the same template will mean two things in two places. So the rule is
written down and made mechanical: the author names the row; if they do not, it is called "row".
Inside the region that name means the current row and nothing else. A plain,
unqualified name always means the document as a whole, so a row can never quietly take over a name
the rest of the document is using. Runtime parameters — the report date, the branch name, supplied
by the calling program rather than by the data — keep their own namespace that nothing can take
over, ever, including a row. And if the author binds a region to something that is not a list at
all, the render stops and says which binding and which element, rather than printing a
plausible-looking document with the table missing.

This story deliberately does **not** print any rows. Repeating a table, wrapping cells, breaking it
over pages and totalling a column are all later stories. This one settles the meaning of a name.
Done looks like: the naming rules hold under tests that would fail if any of them
were reversed, the not-a-list render fails by name, and every existing document produces
byte-identical output.

One thing will look wrong afterwards and is not: most of the new naming machinery has no visible
effect on any page yet, because nothing generates rows until a later story. That is the intended
shape, not an omission.

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

1. **Restore the measurement fixture** above, run it, confirm the four recorded failures reproduce
   at `e7f3f9c`. If they do not, stop and report — the baseline moved.
2. **Re-measure the two guard runs** in finding 5 (add the scratch third-root file to
   `internal/bind`, run `TestBindResolutionRootsAreClosed`, remove the file, confirm clean) so the
   obstacle is observed and not merely read about.
3. **AC0** — introduce the scope value in `internal/bind`, make it the single dispatch, and
   re-point `BindText`, `BindTextSpans` and `collectBandTextRuns` onto it with **no** row set.
   Run `internal/bind`'s existing suite unedited; every message must be byte-identical.
4. **AC1, AC2** — the alias branch and its default, with the two-row subject; capture both
   red-proofs before the fix.
5. **AC3** — the root-never-shadowed assertion with the colliding-key subject; capture its
   red-proof.
6. **AC4** — the three-way collision assertion plus the no-row baseline case; capture its
   red-proof.
7. **AC6** — widen `declaredResolutionRoots` to `{"data", "params", "row"}`; re-run both guard
   tests and record their `t.Logf` output.
8. **AC5** — wire the collection-bind check into `renderDocument`'s prologue; re-home the scratch
   fixture into a real test with all four cases plus the empty-array case; delete the scratch file.
9. **Finding 4** — amend `bindTestTemplateJSON` and `TestRenderScopeFenceIgnoresTableBind`'s doc
   comment as required; keep the column bind expression-shaped and keep the assertion strong.
10. **AC7** — re-render and hash every committed golden and the matrix documents; record the
    comparison against `e7f3f9c`.
11. **AC8** — full-suite run, delta itemised against both scopes, `TestCorpusMeetsP6ExerciseFloors`
    confirmed still red.
12. **Record** the D-3.1.x ratification entry (*Ratification requested*) and any rulings received on
    OD-1 in `folio-mvp-decision-log.md`; update `deferred-work.md` if anything is newly deferred.
13. **Story file, decision log, sprint status → `review`** (`epic-3: in-progress`,
    `3-1-…: review`).

**Stop here — do not commit, do not branch, do not set `done`.**

---

## Dev Agent Record

*(to be completed by the developer)*

### Delivery Log

### Red-proofs captured

### Deltas against baseline
