---
title: 'Story 13.4: Preview runs without sample data, and an absent value is empty'
type: 'feature'
created: '2026-09-07'
status: 'done'
review_loop_iteration: 0
baseline_commit: 'b04767a'
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative — the frozen Intent below governs implementation. Rewritten at close to describe what
actually shipped.*

Before this story, opening Preview without sample data simply refused. It now renders the page from a
stand-in document generated for the template, so an author can see their layout before they have any
data to put in it.

The obvious approach — pick one value meaning "empty" and use it everywhere — turned out to be
impossible, and measuring that came before any design. Different formatting and conditional contexts
each reject a different set of values, and one of them silently deletes the element it guards rather
than complaining. So every referenced value is chosen by the context it appears in, and where two
contexts share no legal value at all the preview refuses that template outright, naming the value and
both contexts, rather than guessing.

The screen states plainly that what is on it was built from invented values, and it withholds every
exactness claim a real preview makes. That disclosure follows the bytes displayed rather than the
current state, and it travels onto the saved file too: exporting a stand-in preview names it as one,
because that export is the only place invented bytes outlive the session.

Three things a later reader should not read as oversights. Collections deliberately stand in as
empty, which is load-bearing for correctness rather than laziness. One internal refusal is honestly
untested, because nothing in the engine can currently reach it. And five follow-ups were filed rather
than fixed here.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Laying out a page is gated on inventing data the author does not have yet. Entering
Preview is not gated at all (D-000.8) — three gates downstream of it are: `renderPreview`'s early
return on `!sampleDataRef.current`, `disabled={!sampleData}` on **Render local PDF**, and the status
line's highest-priority branch `'Preview unavailable: no sample data loaded'`. The engine imposes no
such requirement: the CLI passes `Data("{}")` when `-data` is omitted.

**Approach (OWNER DECISION D-13.4.1, route (a)):** a new **read-only Go projection** emits a
*stand-in data document* for the current template — a real JSON document whose value at each
referenced path is chosen by the expression **wrapping** that path. A new wasm operation exposes it;
the designer sends those bytes on the existing `data` channel. `Render` is therefore called with
genuinely supplied data and its semantics are untouched **structurally, not by discipline**. The
screen then states plainly that it is a no-data preview built from stand-ins, and withholds every
production claim.

## Boundaries & Constraints

**Always:**
- `Render` is untouched. `BINDING_PATH_ABSENT` keeps its severity, the registry stays closed and
  additive (AD-14/I-8), the golden corpus is untouched, and `folio-go` rendering the same template
  with absent data still fails exactly as today (I-9; Epics 8–15 scope fence item 4).
- **Every collection stand-in is `[]`.** This is load-bearing, not incidental: with zero rows the
  entire row-scope surface (every `Column.Bind`) is never evaluated, which is why column bindings need
  no stand-in at all. State that reason wherever the code skips them.
- **The generator's coverage of the function set is exhaustive by construction.** Derive the name set
  from the exported `expr.LegalFunctionNames()`; never restate the names as a literal in the
  generator's test. Adding a function to the engine must red the check.
- **Byte-determinism.** `formatDate`'s stand-in is the fixed instant `"2024-01-15T12:00:00Z"`, never
  `time.Now()`. The same template must yield identical stand-in bytes on every run and every target.
- **No production claim on a stand-in render.** `EXACT LOCAL PRODUCTION PDF`, the viewer's
  `Current exact local production PDF…` label, and any cross-target or "matches native" wording must
  not appear for a no-data preview. Amber (`--color-bind`), never cyan — cyan is the authority colour
  that marks production output and the hash (DESIGN.md §Colors), which is the claim this state
  withholds.
- **The notice must name the fabricated condition, and only when there is one.** If the template
  contains at least one **fabricated condition — a `visibleIf` OR an `if()` condition** — the no-data
  notice must say that conditional content may be present or absent because its condition had no data
  to evaluate. If the template contains neither, that sentence must not appear — a disclosure shown
  unconditionally is noise everyone learns to skip.
  <!-- AMENDED BY THE ORCHESTRATOR AFTER APPROVAL, 2026-09-07 (D-13.4.2). The clause originally said
  `visibleIf` alone. That was MY error at the approval gate, not the builder's: `conditionalRule` also
  fabricates `if(cond, …)`'s first argument, so an `if()`-only template rendered a literal branch chosen
  by no data with the disclosure withheld. The builder refused to soften it to a patch on its own
  authority and brought it back as an intent gap, which is correct — only a human reopens this block. -->
  Both directions are matrix rows and both are tested. (It names *both* directions, not only hiding: a
  condition that resolved `true` from a stand-in shows an element that real data might hide, so the
  conditional layout is untrustworthy either way.)
- Voice is terse and technical: state the fact, name the location, offer no comfort (UX-DR24). Never
  apologize, never an exclamation mark.
- **Do not add a second literal `setPreviewStatus('stale')` to `App.tsx`.**
  `preview-authority-contract.test.ts`'s red proof uses `String.replace`, which substitutes only the
  first occurrence, so the assertion silently requires exactly one. There is exactly one today.
- Keep the freshness `<p>`'s opening tag byte-identical:
  `id="preview-freshness-status" className="preview-status" role="status"`. The ternary inside it is
  free to change.
- Every new control and notice is keyboard-reachable, labelled, with visible focus (UX-DR25).
- Params are **out of scope** by ruling, not by omission: `params.*` keeps today's behaviour and an
  absent param stays a located `BINDING_PATH_ABSENT`.

**Ask First:**
- Any change to `preview/pdf-viewer.tsx`, its render effect, or `PDFPreviewViewState` — DW-191 is
  Story 13.2's and its tear-down is live on the horizontal axis today.
- Any new `@media` rule in `App.css` — `canvas-authority-contract.test.ts` asserts the media-query
  list equals exactly `['prefers-reduced-motion: reduce']`.
- Any new or re-meant diagnostic code, and any new error **type** in a `folio` root file —
  `TestFolioMethodNamesAreInjective` forbids a second `Error`/`Unwrap` receiver there.
- Any change to the engine protocol's **shape** or version. Adding an operation name is expected;
  adding a request field or a response field is not.
- Any change to observable behaviour when sample data **is** loaded, to `.folio` open/save, or to
  `params`.
- Any new dependency, any file not named under **Execution**, any adjacent defect discovered.

**Never:** commit, `git add`, stash, checkout, reset, revert, restore, push, create a branch, or open
a pull request — **never commit**; the human makes every commit in this run. Never touch
`fixtures/declared-variants/expected.pdf`, `input.folio` or `signoff.json` (human-attested). Never
build the thumbnail rail, the evidence rail, viewer navigation, or the chrome rework (13.2, 13.3,
13.5). Never widen the `src/preview/` canvas-authority exception. Never suppress an engine diagnostic
in preview.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Bindings-free template, no sample data | `sampleData === undefined` | stand-in document is `{}`; the page renders; the no-data notice is shown | N/A |
| Bound scalar path, no sample data | text `{{customer.name}}` | the path resolves to a stand-in that renders **empty**; the page is produced | N/A |
| `formatNumber(p, "…")` | p referenced only there | stand-in is `0` | N/A |
| `formatDate(p, "…")` | p referenced only there | stand-in is the fixed instant `"2024-01-15T12:00:00Z"` | N/A |
| `visibleIf: p` | p referenced only there | stand-in is `true` — the element is **shown**, because a layout preview exists to show the layout | N/A |
| p in BOTH a text binding and a `visibleIf` | e.g. `{{flags.vip}}` and `visibleIf: flags.vip` | stand-in is `null` — legal in both; the text renders empty and the element **hides**, with no diagnostic (D-3.2.3) | N/A |
| Template contains at least one fabricated condition — a `visibleIf` **or** an `if()` condition | any no-data preview of it | the notice **includes** the sentence that conditional content may be present or absent because its condition had no data to evaluate | N/A |
| Template contains **neither** a `visibleIf` **nor** an `if()` condition | any no-data preview of it | that sentence is **absent** from the notice — asserted, not merely unmentioned, so the disclosure cannot silently become unconditional. **The fixture MUST carry at least one real component with a non-conditional binding**: a zero-component fixture cannot distinguish an `if()`-only template from a condition-free one, which is what made the original negative row vacuous | N/A |
| p in contexts with **no common value** | e.g. `{{p}}` and `formatNumber(p, "…")` | **refuse to enter no-data preview for this template**, with a located message naming the path and both contexts | refused, never mangled |
| Bound table | `bind: "transactions[]"` | collection stand-in `[]`; header-only table, zero rows; `Column.Bind` paths never evaluated | N/A |
| `avg` footer over the stand-in collection | one column with `footer: "avg"` | the render succeeds and the engine's real `CodeEmptyAverage` **Warning** is shown as a diagnostic card | shown, not suppressed |
| Sample data is then loaded | `loadSample` | the identity key changes because `data` is hashed; the no-data preview is marked stale and re-renders on the exact-production path with no stand-ins | N/A |
| Path absent from data that WAS supplied | sample loaded, path missing | unchanged: the located `BINDING_PATH_ABSENT` Error Story 6.6 presents | announced |
| Sample data is cleared | `clearSampleData` | returns to a no-data preview rather than to an empty `'idle'` screen | N/A |
| Stand-in projection fails or is stale | revision moved mid-request | no preview is installed and no guessed empty document is substituted | surfaced as a local preview issue |

</frozen-after-approval>

## Code Map

Anchors re-derived by symbol at `b04767a` (D-000.4). Every Epic 13 anchor from the 2026-09-05
grounding report was SAMPLED and is re-derived here. HEAD moved twice under this plan gate
(`1091645` → `0a992a3` → `b04767a`); both commits touched only `_bmad-output/`, so the code baselines
below were re-measured at `b04767a` and stand.

**The three gates, and the thing that is not one.**
- `folio-designer/src/App.tsx` — `renderPreview` (`if (!sampleDataRef.current) { setPreviewStatus('idle'); return }`); `runPreview` (`const sample = sampleDataRef.current` … `if (!engine || !snapshotRef.current || !sample) return`, and `sample.bytes.slice(0)` is the sole source of the `data` payload); the preview-inputs JSX `disabled={!sampleData}` on **Render local PDF**; the preview-status `<p>` where `!sampleData` outranks every `previewStatus` branch.
- **Entering Preview is not gated.** The PREVIEW switch carries no `disabled` and `enterPreview` has
  no sample check. The epic's "Preview also refuses to run at all" prose is loose; D-000.8 is right.
- `loadParameterReferences` is the **shape to mirror** for a template-derived projection: capture
  `snapshotRef.current` and `documentGeneration`, bump a request counter, and admit the result only if
  the request counter, the document generation and **three** revisions all still agree — "never turn
  an unavailable projection into a guessed empty one".

**Where the stand-in generator goes, and why it needs no new exported engine surface.**
- `folio-go/parameter_references.go` — `ParameterReferences` / `collectParameterPaths`. The structural
  precedent: walks `PageHeader`/`Content`/`PageFooter`, reuses `expr.Parse` and
  `expr.ScanPlaceholders`, bounds its output, returns sorted results, and declares one new function on
  no new type. Copy this shape.
- `folio-go/internal/expr/table.go` — `functionTable` is `[8]funcEntry` (**sum, count, avg, formatDate,
  formatNumber, upper, lower, if**) and is **unexported**, as are `funcEntry`, `argKind` and
  `returnKind`. **`LegalFunctionNames()` is the only exported handle** — and it is sufficient, because
  `funcEntry.args` carries only *syntactic* constraints (`argAny`, `argNotLiteral`,
  `argStringLiteral`) and says nothing about the runtime value kind an argument needs. The value-kind
  rules live inside each `evalX`. **So no exported projection needs adding and the generator lives in
  `package folio`, not in `package expr`** — it derives the name set from `LegalFunctionNames()` and
  owns the value-kind rule itself.
- `folio-go/internal/expr/table_derivational_test.go` — `TestImplementedEntriesMatchEvalCallSwitch` is
  the exhaustiveness pattern to port: AST-extract one set, compare set-equality **in both directions**
  against the other, restate no names, and guard with a `len(...) == 0` presence precondition so an
  empty extraction cannot pass vacuously.

**Every site a data path can reach (the completeness surface — a missed site is the hard failure
this story exists to avoid).**
- `folio-go/internal/template/model.go` — `Bands` has exactly `PageHeader`, `Content`, `PageFooter`
  and deliberately carries no `Extra`, so the three-band walk cannot silently grow. `ElementType` is
  the closed five: `text`, `image`, `table`, `line`, `rect`.
- `Element.VisibleIf` is on **all five kinds**, is a **bare** expression (no `{{ }}`), and is
  evaluated in **document scope** (`render_visibility.go` `computeVisibility`, no `.WithRow`).
- `Element.Value` — text only, `{{ }}` interpolation, document scope. Resolved **even when the element
  is hidden** (the bind runs before the visibility skip), so hiding does not spare its paths.
- `Element.Asset` is a literal `assets` map key, never bound. `line` and `rect` have no kind-specific
  fields at all. `Style.*`, `altRowBackground` and `headerStyle.*` are **negative space** — a
  placeholder there is a load error.
- Table fields are **flat on the element**, not nested under a `table` key: `bind`, `as`, `columns`,
  `headerHeight`. `TableExt.Bind` is a document-root collection path requiring an **array**;
  `Column.Bind` is `{{ }}` in **row scope**; `Column.FooterOf` is a document-root
  `<collection>.<field>`; `Column.Label`/`FooterFormat` are literals; a column-level `visibleIf` is a
  load error.
- **`folio-go/table_render.go` `footerCellExprText` SYNTHESIZES aggregate expressions that appear
  nowhere in the placeholder stream**: `count(TrimSuffix(bind, "[]"))`, and `sum|avg(footerOf)` where
  `footerOf` is the explicit `Column.FooterOf` or the one derived from `Column.Bind` by
  `expr.DeriveFooterOf`. They are evaluated at **document scope** (`bind.NewScope(data, params)`, no
  row scope). A walker that reads only template text misses all of them — cover `bind` and `footerOf`
  as first-class sources.
- `folio-go/render.go` `checkTableBindings` — absent/null/non-array is a hard error; **an empty array
  is explicitly not**. `folio-go/table_render.go` — the row block is guarded by
  `if len(items) > 0 || hasFooter`, and the row loop iterates zero times over `[]`, which is why every
  `Column.Bind` is unreachable under this story's `[]` stand-in.
- `Document.UnbreakableValues` is the one data-path site outside the bands, and it is **never
  resolved** — `wrap.go` `atomicSpansFor` does pure string equality against `Substitution.Path`, so an
  unmatched entry is silently inert. Emit nothing for it, deliberately.
- Excluded roots: `params` (always selects the params root — `selectRoot`), the reserved whole-tokens
  `page` and `pages` (`internal/expr/scan.go` `reservedPlaceholders`), and, inside a table only, that
  table's own row alias (`resolvedRowAlias`, default `"row"`).

**Where "empty" is decided (read-only).**
- `folio-go/internal/bind/value.go` — `Presence` is `Present | Absent | Null`; `Lookup` returns `Null`
  only for a leaf that is an explicit JSON null, so **nulls must sit at leaves**, never at an
  intermediate node.
- `folio-go/internal/bind/text.go` — `lookupBound`'s `case Null:` yields `KindNull`; the text-binding
  write switch accepts **only `KindNull` and `KindString`** and everything else is
  `"resolved to a %s, not a string — text bindings are never coerced"`; `CollectionLength` and
  `ProjectCollection` each have a `Null` arm.
- `folio-go/internal/expr/eval.go` — `evalUpperLower` rejects null; `ConditionValue` accepts only bool
  or null and returns `false, nil` for null (**owner ruling D-3.2.3, silent**).
  `internal/expr/numberformat.go` and `internal/expr/formatdate.go` carry the other two refusals.

**The protocol, which needs no shape change.**
- **A new operation name must join THREE lists, not two.** `folio-designer/src/engine-protocol.ts`
  carries two — the `EngineOperation` union and the runtime allowlist in `validateRequest` — and
  `folio-designer/src/engine-client.ts` carries the third, `matchesOperationPayload`, whose
  `default: return none` rejects any bytes-carrying response for an operation with no case. An
  operation missing from that third list does not fail loudly; it simply never works. **This inventory
  said "two" at the plan gate and was wrong**, in the same epic and two stories after D-000.8's
  "two hardcodings" was wrong when there were three. Count the lists by searching for the operation
  name, never by memory.
- `folio-designer/src/engine-protocol.ts` — the success envelope's
  `hasOnly([... 'bytes' ...])` **already permits a `bytes` response**, so a stand-in operation returning
  `{snapshot, bytes}` needs no new response field. `isRenderPayload`/`isIdentityPayload` use
  `hasExactKeys` and require `byteLength > 0` — a stand-in document is non-empty by construction, so
  no widening is needed and no extra key is tolerated.
- `folio-designer/src/engine.worker.ts` — byte decoding is operation-agnostic
  (`base64ToBytesBounded(..., operation === 'render' ? MAX_ENGINE_RENDER_PDF_BYTES : undefined)`), so a
  bytes-returning operation needs no worker change. Note the `undefined` bound: **the generator must
  bound its own output**, as `ParameterReferences` bounds at 128.
- `folio-go/wasm/engine.go` — `Engine.ParameterReferences` is the method shape;
  `Engine.Render`/`Engine.PreviewIdentity` each refuse zero-length inputs, re-enforced in
  `wasm/cmd/engine/main.go`'s render and identity arms. `folio-go/preview_identity.go` hashes
  `template ∥ data ∥ params ∥ version ∥ font faces` length-delimited, so **swapping stand-in bytes for
  real ones changes the key for free**.

**Freshness, which mostly already works.**
- `loadSample` calls `invalidatePreview()` then `schedulePreview()` — that already marks a no-data
  preview stale and re-renders, **provided the no-data render installed a real `PreviewRecord`**. If it
  is modelled without one, that call degrades to `'idle'` and nothing announces the change.
- `clearSampleData` calls `invalidatePreview(true)`, which clears to `'idle'`; after this story it must
  land on a no-data preview instead.
- `folio-designer/src/preview/freshness.ts` — `PreviewFreshness`, `StaleReason`, `staleCopy`,
  `canInstallPreview`. Read-only.

**Contract tests to keep green, and the with-data literals that must not move.** The production-claim
strings are asserted in **fifteen** places, every one of them on the with-data path
(`App.test.tsx` ×9, `e2e/pdf-export.spec.ts` ×2, `e2e/preview-parameters.spec.ts` ×2,
`e2e/browser-native-roundtrip.spec.ts` ×1, `e2e/application-shell.spec.ts` ×1). **Make the new wording
conditional on the no-data flag; do not rewrite the with-data wording.** Also:
`preview-authority-contract.test.ts` (13 forbidden tokens, the pinned freshness tag, the single-
occurrence `setPreviewStatus('stale')` red proof); `canvas-authority-contract.test.ts` (media-query
list, `prohibited` scan with comments stripped); `design-contract.test.ts` (no hex in `App.css`, radii
via `var(--radius…)`, `.preview-failure` pinned verbatim, exactly one `var(--type-display)`);
`folio-go/diagnostic_registry_census_test.go` (`CodeBindingPathAbsent` must stay producible);
`folio-go/render_arch_test.go` `TestFolioMethodNamesAreInjective`.

**Coverage that does not exist yet.** Population: `folio-designer/src` and `folio-designer/e2e`,
searched with `grep -a`. `'Preview unavailable'` returns exactly **one** hit — the production literal
in `App.tsx` — and **zero** test files; no test asserts the disabled **Render local PDF** button
(`grep -an -E 'toBeDisabled|disabled' src/App.test.tsx` filtered for preview/sample/render terms
returns nothing, against an unfiltered control with many hits). Positive control for the instrument:
the same search style finds `'Binding unavailable: no sample data loaded.'` in three files. **The
refusal this story removes is guarded by nothing** — so no existing test needs deliberate editing, and
nothing would have caught a regression either.

**`folio-go/testdata/example/first-pdf.folio`** — one text element, `Hello, {{customer.name}}!`, no
table. The ready-made e2e fixture for a bound template previewed with no sample data.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/stand_in_data.go` (new) -- add a read-only projection, modelled on
  `ParameterReferences`, that walks the three bands and emits the stand-in JSON document for a
  template. Cover every site named in the Code Map's completeness surface: text `value` placeholders,
  `visibleIf` on all five kinds, table `bind`, explicit and derived `footerOf`, and `count`'s
  synthesized operand. Exclude `params`, the reserved `page`/`pages` tokens, and each table's own row
  alias. Every collection is `[]`; nulls sit only at leaves; the output is bounded and sorted so it is
  byte-identical run to run. Declare **no new error type** — `TestFolioMethodNamesAreInjective`
  forbids a second `Error`/`Unwrap` receiver in a `folio` root file.
- [x] `folio-go/stand_in_data_test.go` (new) -- the exhaustiveness gate, ported from
  `TestImplementedEntriesMatchEvalCallSwitch`: set-equality in **both** directions between the
  generator's per-function rules and `expr.LegalFunctionNames()`, restating no names, with a
  `len(...) == 0` presence precondition. Plus: one round-trip per function proving the stand-in is
  *accepted* (generate, render, assert success); the empty-intersection refusal; determinism (two
  generations byte-identical); and that no reserved or `params` key appears in the output.
- [x] `folio-go/wasm/engine.go` + `folio-go/wasm/cmd/engine/main.go` -- expose the projection as one
  new operation returning `{snapshot, bytes}` on the **existing** response envelope. Mirror
  `ParameterReferences`'s method and dispatch-arm shape. No new request field, no new response field,
  no protocol version change.
- [x] `folio-designer/src/engine-protocol.ts` -- add the operation name to `EngineOperation` and to
  the runtime allowlist in `validateRequest`. Nothing else: `bytes` is already an allowed response
  key and the payload guards already reject an empty `data`.
- [x] `folio-designer/src/App.tsx` -- drop the three gates. In `runPreview`, when there is no sample,
  request the stand-in document inside the same revision-guarded flow and use its bytes as `data`,
  admitting it only under the same request/generation/revision agreement `loadParameterReferences`
  uses; never substitute a guessed empty document. Carry a "built from stand-ins" flag on
  `PreviewRecord` so it survives into the stale state. Make the heading, the viewer label, the status
  line and the digest line conditional on that flag so no production claim is made, and render the
  no-data notice. Make `clearSampleData` land on a no-data preview. **Do not add a second literal
  `setPreviewStatus('stale')`.**
- [x] `folio-designer/src/App.css` -- style the notice in amber via existing tokens. No hex, no
  gradient, radii only via `var(--radius…)`, and **no new `@media`**.
- [x] `folio-designer/src/App.test.tsx` -- integration over the matrix rows: the bindings-free and the
  bound no-data renders, every withheld production claim, the notice, the enabled **Render local PDF**
  button, load-then-restale, clear-then-no-data, and the unchanged located Error for a path absent
  from data that *was* supplied. Mutation-prove each new guard by deleting the line it guards.
- [x] `folio-designer/e2e/preview-no-data.spec.ts` (new) -- open
  `folio-go/testdata/example/first-pdf.folio`, go straight to PREVIEW without loading JSON, assert a
  PDF renders and the notice is visible and no production claim is shown. Real coverage, executed in
  CI since `adf905a` — never described as compile-only. Use no identifier
  `canvas-authority-contract.test.ts` forbids; every `e2e/` file is auto-enrolled by directory walk.

**Acceptance Criteria:**
- Given the generator and the engine's own function registry, when a ninth function is added to
  `functionTable` without a matching generator rule, then the exhaustiveness test fails naming that
  function — and the test contains no literal list of function names. **Mutation-prove by deleting one
  function's rule from the generator**, confirm the test reds naming it, restore by `cp`, and report
  the restore digest.
- Given a template whose only bound path is a plain text binding, when it is previewed with no sample
  data, then the page is produced and that path renders as nothing at all — asserted against a literal
  expected string, never against a value read back out of the generator (D-11.2.2, D-11.2.8).
- Given the same template and the same engine, when `folio-go`'s own `Render` is called with absent
  data, then it still fails with the located `BINDING_PATH_ABSENT` Error exactly as it does today. This
  is the story's central promise and it gets its own pin.
- Given a template with a bound table, when it is previewed with no sample data, then the table renders
  header-only with zero rows. **Mutation-prove that `[]` is load-bearing**: change the collection
  stand-in to a one-element array, confirm the row-scope column bindings then fail the render, restore
  by `cp`, and report the digest.
- Given a path used in both a text binding and a `visibleIf`, when it is previewed with no sample data,
  then the render succeeds, the text is empty and the element is hidden — and given a path whose
  contexts share no legal value, the no-data preview is refused with a message naming the path and both
  contexts, never a mangled or failed render.
- Given two generations of the stand-in document for one template, when their bytes are compared, then
  they are identical; and given the generator's source, when it is searched for a clock, then there is
  none.
- Given a template carrying at least one `visibleIf`, when it is previewed with no sample data, then
  the notice states that conditional content may be present or absent because its condition had no data
  to evaluate — and given a template carrying none, that sentence is asserted **absent**. The negative
  direction is a real test, not an omission: it is the one that stops the disclosure quietly becoming
  unconditional, and it is mutation-proved by making the sentence unconditional and confirming the
  no-`visibleIf` case reds.
- Given a no-data preview on screen, when the screen is inspected, then no production claim appears —
  and given the same template with sample data loaded, then every one of those claims returns
  unchanged, with the fifteen existing with-data assertions still green.
- Given a no-data preview, when sample data is then loaded, then the preview is marked stale and
  re-renders on the exact-production path with no stand-ins.
- Given every guard added or changed by this story, when the line it guards is deleted, then that guard
  fails — restored by `cp` from a snapshot, with the restore digest reported.

## Spec Change Log

- **Plan gate, 2026-09-07 — owner ruling D-13.4.1 (not a review loopback).** The story's second AC is
  not deliverable as written: there is no single value that renders empty. Measured first-hand —
  `evalUpperLower` rejects null; `formatNumber` rejects null and `""`; `formatDate` rejects both and
  has no empty form; `ConditionValue` returns `false, nil` for null so a null-filled document silently
  deletes every `visibleIf` element (D-3.2.3); `checkTableBindings` refuses a null collection. The
  owner chose route (a) and **amended `epics.md`'s "This epic touches no engine byte" in place** to
  record the exception. **KEEP on any re-derivation:** the four refusals and the D-3.2.3 hiding are the
  evidence for the whole route; a re-derivation that drops them will re-propose the null-everywhere
  document that provably does not work.

- **Plan gate, 2026-09-07 — correction to the ruling's premise on the text ∥ `visibleIf` collision.**
  The condition attached to D-13.4.1 stated that a path used in both a text binding and a `visibleIf`
  has contradictory requirements and that **"no value satisfies both"**. Measured: the text-binding
  write switch accepts `KindNull` **or** `KindString`; `ConditionValue` accepts `KindBool` **or**
  `KindNull`. The intersection is **non-empty — `null` satisfies both.** The consequence is a hidden
  element, not a failure, so refusing the whole template for this case would be unnecessary. The real
  contradiction class is elsewhere and is genuinely empty (text ∩ `formatNumber`, `visibleIf` ∩
  `upper`, `upper` ∩ `formatNumber`). **The rule is therefore per-path intersection of every context's
  accepted kinds, refusing only when that intersection is empty** — one rule that covers both cases.
  Known-bad state avoided: refusing a whole template for a case that has a legal, non-failing answer,
  in a mode whose purpose is never to refuse.

- **Plan gate, 2026-09-07 — reachability of the empty-intersection refusal, measured.** Population: all
  **32** `.folio` templates in the repo excluding `node_modules` and `.git`. `visibleIf`: **0**
  templates. Explicit `footerOf`: **0**. Multi-context bare paths: **0**. The only function used
  anywhere is `formatNumber`, 4 occurrences, all inside a column `bind` in the four statement fixtures
  — row scope, therefore unreachable under the `[]` stand-in. **First measurement returned zero for
  four contexts and that was an instrument fault, not a finding**: table fields are flat on the element,
  not nested under a `table` key. Re-run corrected, the positive control now finds `tableBind` and
  `colBind` in 7 templates each. So the refusal arm is unreachable in today's corpus — **but the
  designer can author `visibleIf`** (`visibilityField`, `App.tsx`), so it is reachable through the
  product and must be implemented, not assumed away. **KEEP:** the corpus count with its instrument
  correction, so a later reader does not re-derive the false zero.

- **Plan gate, 2026-09-07 — the exhaustiveness seam, decided rather than discovered.** Compile-time
  exhaustiveness over function names is impossible: `funcEntry.name` is a bare `string` and dispatch is
  `switch entry.name`; `functionTable` being `[8]funcEntry` constrains only cardinality, and widening
  to `[9]` compiles. **`LegalFunctionNames()` is the only exported handle, and it is sufficient** —
  `funcEntry.args` carries only syntactic constraints (`argAny`/`argNotLiteral`/`argStringLiteral`) and
  no runtime value kinds, so no exported projection would help and none is added. **The generator lives
  in `package folio`**, derives its name set from `LegalFunctionNames()`, and owns the value-kind rule
  itself, gated by a both-directions set-equality test ported from
  `TestImplementedEntriesMatchEvalCallSwitch` including its `len(...) == 0` presence precondition.

- **Plan gate, 2026-09-07 — the synthesized-aggregate hole, and what closes it.** `footerCellExprText`
  synthesizes `count(…)` and `sum|avg(footerOf)` at render time from `TableExt.Bind` and
  `Column.FooterOf`; these appear nowhere in the placeholder stream, so a text-only walk misses them.
  Independently confirmed by a second reading of `table_render.go`. Closed by treating `bind` and
  `footerOf` as first-class sources. **A second, larger simplification falls out of the `[]` ruling:**
  with every collection empty the row block is skipped (`if len(items) > 0 || hasFooter`) and the row
  loop iterates zero times, so **no `Column.Bind` is ever evaluated** and row-scope paths need no
  stand-in at all. **KEEP:** `[]` is load-bearing for correctness, not a convenience — the AC
  mutation-proves it.

- **Step-03, 2026-09-07 — the operation-name inventory undercounted, and `engine-client.ts` is
  accepted outside the Execution list (patch, not a loopback).** The Code Map said a new operation name
  must join "the two lists"; there are **three**. The third is `matchesOperationPayload` in
  `folio-designer/src/engine-client.ts`, whose `default: return none` rejects any bytes-carrying
  response for an operation it has no case for — so the operation does not fail loudly, it simply never
  works. The implementer edited that file, which the Execution list did not name; ruled **accepted** by
  the coordinator under the standing carve-out for work with no alternative that makes a stated
  acceptance criterion pass. Known-bad state avoided: an operation wired into two of three lists, which
  presents as an inexplicable dead projection rather than as an error. **KEEP on any re-derivation:
  three lists, and count them by searching for the operation name rather than from memory** — this is
  the same failure shape as D-000.8's "two hardcodings" when there were three, in the same epic two
  stories earlier.

- **Step-03, 2026-09-07 — `schedulePreview()` deliberately NOT added to `clearSampleData` (measured,
  not overlooked).** The task text implied a call there. The implementer wrote it, mutation-proved it,
  and found deleting it reds nothing: both callers already end with
  `if (modeRef.current === 'preview') { … renderPreview() }`, and what used to make that tail land on
  `'idle'` was the sample gate this story removes. Shipping it would have been a guard that cannot fail
  (D-000.9). Ruled **accepted** by the coordinator. **KEEP: the comment recording the measurement must
  survive review** — the absence reads as an oversight to anyone who has not run the proof, and the
  next reader will otherwise "fix" it by adding dead reassurance.

- **Plan gate, 2026-09-07 — citations corrected (Q4, ruled).** The epic's "Covers: FR9, FR34" is wrong:
  **FR34 is the Design Canvas** ("an approximate visual representation"); the exact PDF preview is
  **FR35**. This spec cites **FR9 · FR35 · AD-18 · UX-DR13, UX-DR14, UX-DR21, UX-DR24**. And
  EXPERIENCE.md declares "Empty — no sample data" for **S3, the Binding Panel** — the Preview surface
  (S5) declares no states of its own — so this story **extends** a declared state to a new surface
  rather than implementing one.

- **Step-04 triage, 2026-09-07 — one intent gap, twelve patches, five defers, two rejects (NO loopback;
  `review_loop_iteration` stays 0).** Three review layers ran; the verification-gap layer ran ALONE
  (D-11.2.7) and its restoration was verified by `shasum -c` over all 11 touched files, not accepted on
  its word.
  **The intent gap:** the frozen disclosure clause keyed on `visibleIf`, but `conditionalRule`
  fabricates `if(cond, …)`'s first argument too, so an `if()`-only template showed a literal branch
  chosen by no data with the disclosure withheld — and the frozen matrix *required* that, via a row
  asserting the sentence is absent when no `visibleIf` exists. Root cause inside the frozen block, so it
  was escalated rather than patched even though the fix was one predicate. **The human amended the
  frozen block (D-13.4.2) and chose the cheap remedy**, and additionally required the negative fixture
  to carry a real bound component — the zero-component fixture could not distinguish an `if()`-only
  template from a condition-free one, so that row passed while asserting nothing.
  **KEEP on any re-derivation:** (1) the disclosure trigger is "any fabricated condition", never
  `visibleIf` alone; (2) the negative fixture must carry a real non-conditional binding, or the row goes
  vacuous again; (3) the `if(` detector must be run against the nearest legitimate spellings — it was,
  and it separates `{{if(…)}}` from `notif(`, `ifx(`, `customer.if(` and an `if (` outside any
  placeholder.

- **Step-04 triage, 2026-09-07 — Save PDF ruled IN SCOPE (patch).** The export named a stand-in PDF
  `Save PDF` and announced `Saved PDF of revision 1 as statement.pdf`, word-for-word a production
  export. Ruled in scope because the frozen **Intent** governs and the Boundaries' surface list is
  examples, not a schedule: the export is the one place these bytes leave the machine, and the file
  outlives the session with nothing on it saying otherwise. Staleness and stand-in now **compose**
  (`Save stale no-data PDF`) rather than choosing. **KEEP: they compose; a later change that makes them
  exclusive silently drops one of two true claims.**

- **Step-04 triage, 2026-09-07 — the second exhaustiveness axis, and what `[]` is holding up.** The
  function-name registry was gated, but `walk()` and `collectElement()` silently returned nil for an
  unrecognised AST node or element kind — the same silent-drop class on a different axis. Element kinds
  are now gated by AST-extracting `closedElementTypes` from `internal/template/closedsets.go` (derived,
  never restated) and unknown kinds are located refusals. **KEEP: the `walk` half is a located refusal
  but is NOT mutation-proved** — it is unreachable while `internal/expr` has exactly four node kinds,
  and proving it would mean adding a node type to the engine. Stated rather than implied.
  Separately, the band walk had two-thirds unverified (every fixture hardcoded empty header/footer
  arrays; reducing the loop to `Content` left the package green) — now covered in all three bands.

## Design Notes

**The owner's ruling, carried verbatim (D-13.4.1, 2026-09-07).**

> The owner chose route (a): a new read-only Go function, mirroring `ParameterReferences`'s three-band
> walk and built ON `expr.Parse`/`ScanPlaceholders`, emitting a document whose value at each referenced
> path is chosen by the expression WRAPPING it — `""` for string contexts, `0` for `formatNumber`, one
> fixed instant for `formatDate`, `true` for `visibleIf`, `[]` for a table bind. A wasm op exposes it
> and the designer sends the bytes on the existing `data` channel.
>
> What that buys, and it is the reason (a) beat the alternatives: `Render` is called with genuinely
> supplied data, so its semantics, error contract, codes and goldens are untouched **structurally
> rather than by discipline**. Staleness comes free — `data` is already hashed into the preview
> identity, so loading real data changes the key and `loadSample`'s existing `invalidatePreview()`
> marks the no-data preview stale with no new code.
>
> The generator's coverage must be exhaustive **by construction** — the operator list derived from the
> engine's own registry, never restated as a literal in the test, so adding an operator to the engine
> reds the check. Plus: the stand-in must be byte-deterministic (a fixed instant, never `time.Now()`),
> and the screen must withhold exactly what the mockup withholds — a page built from fabricated values
> is **not** production output and must never present as it.

**The per-context requirement table, measured.** This is the rule the generator implements, as an
intersection per path rather than a per-context assignment:

| Context | Accepts | Refuses |
|---|---|---|
| bare `{{p}}` text binding | string, null (null renders empty, AD-14) | everything else — "never coerced" |
| `visibleIf: p` / `if(p, …)` condition | boolean, null (**null ⇒ element hidden**, D-3.2.3) | string, number — no truthiness |
| `upper(p)` / `lower(p)` | string | null, number, boolean |
| `formatNumber(p, "…")` | number | null, string |
| `formatDate(p, "…")` | RFC 3339 string, integral epoch ms | null, `""` |
| `sum`/`avg` projected field | number, null | absent, other kinds |
| `count(c)` | bare collection path | a trailing field |
| table `bind` | array (**empty is legal**) | null, absent, non-array |

The preference within a non-empty intersection is **not** "emptiest" — emptiness is about what a
displayed value renders as, and it would wrongly hide every conditional element. The rule is: a path
that is *displayed* resolves to something that renders as nothing (`""`, or `null`, or `0`/the fixed
instant where a typed operator demands one); a path that *only gates visibility* resolves to `true`, so
a layout preview shows the layout. Where both apply, `null` is the only common member and the element
hides — the one degradation, and the notice says so. An empty intersection refuses the template with a
located message. `if`'s second and third arguments take their requirement from the context *enclosing*
the `if`, not from `if` itself.

**Why `"2024-01-15T12:00:00Z"` and not an edge instant.** `validateCivilRanges` and
`civilFromInstantMs` re-bound the **shifted** civil year to `[1, 9999]`, so an instant near either
bound can flip validity under the document's own `utcOffset`. A mid-range instant cannot.

**Determinism has a measured basis, not an asserted one.** Nothing in `internal/expr` imports `"time"`,
reads the clock, `os.Getenv`, `LC_*` or `math/rand` — an import scan whose positive control did find
`math/big`, `strconv` and `regexp`. So output is byte-deterministic **given the document's own locale
and `utcOffset`**, which is exactly what the determinism condition needs. Cite this measurement rather
than the claim.

**Why the no-data preview must still install a real `PreviewRecord`.** `loadSample`'s existing
`invalidatePreview()` sets `previewStatus` to `'stale'` only `if (previewRef.current)`; with no record
it degrades to `'idle'` and the transition from stand-ins to real data announces nothing. The freshness
AC is satisfied by existing machinery **only** on that condition.

**Two live defects this story must not touch.** DW-191 — `App.css` sets `overflow: auto` on both axes
with `canvas { max-width: none }`, so a zoomed page already scrolls horizontally and the render
effect's tear-down in `pdf-viewer.tsx` is live, not latent; it is 13.2's. DW-272/DW-273 — `fileStatus`
and `fileError` survive a mode switch, and a Design-mode template save still announces nowhere; both
are 13.1's leftovers and both are OPEN.

## Verification

All five baselines **re-measured first-hand at `b04767a`** (the two intervening commits touched only
`_bmad-output/`).

**Commands** (all paths absolute; never `cd` before a relative restore path):

- `cd /Users/panitw/Projects/folio/folio-designer && npm test` — baseline **65 files / 993 tests, 0
  failing**. Expect ≥ those totals; report the new numbers and name every added test file.
- `cd /Users/panitw/Projects/folio/folio-designer && npx tsc -b --force` — expected: exit 0 (clean at
  baseline).
- `cd /Users/panitw/Projects/folio/folio-designer && npx oxlint` — expected: **exactly 4**
  `only-export-components` warnings, 0 errors. Baseline sites at `b04767a`:
  `preview/pdf-viewer.tsx:16,17` and `App.tsx:3665,3672`. The count and the rule are the invariant; the
  lines are not.
- `cd /Users/panitw/Projects/folio/folio-designer && npm run test:e2e:compile` — expected: exit 0
  (clean at baseline).
- `cd /Users/panitw/Projects/folio/folio-go && go test -count=1 ./...` — baseline **2242 pass / 2 fail /
  5 skip**, failing ONLY `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest, checked
  by enumerated NAME and never by exit code. **A third distinct failure is a hard stop.**
- `gofmt -l /Users/panitw/Projects/folio/folio-go /Users/panitw/Projects/folio/lint` — expected: empty.
  ABSOLUTE paths only; a persisted `cd` makes this emit phantom `lstat` lines.
- `cd /Users/panitw/Projects/folio/lint && go test -count=1 ./...` — `-count=1` always: the rules
  package walks the `folio-go` tree with `ReadDir`, which Go's test cache does not track, so a **new Go
  file in `folio-go` is exactly the change a cached `ok` would hide**. This story adds one.
- Do **NOT** run `npm run build`.

**E2E is real coverage now.** Since `adf905a`, CI runs the unfiltered `-tags=matrix` suite and all
Playwright tests on every push (DW-268 discharged; cadence recorded in D-13.4 / D-13.1.4). The e2e
added here executes per-commit and is reported as executed-in-CI, never as a compile-only placeholder.

**Manual checks:**
- `git status --porcelain` after every mutation restore, before the next probe; report each restore
  digest.
- Confirm no file outside the **Execution** list is modified, and that `13-1-*.md`,
  `sprint-status.yaml`, `deferred-work.md` and `epics.md` are untouched — the coordinator edits those.

## Suggested Review Order

**The one idea the change turns on — a stand-in is chosen by the expression WRAPPING the path**

- Start here: the whole projection, and why `Render` never learns it happened.
  [`stand_in_data.go:379`](../../folio-go/stand_in_data.go#L379)

- The closed value set, in PREFERENCE order — displayed paths render as nothing, conditions show.
  [`stand_in_data.go:53`](../../folio-go/stand_in_data.go#L53)

- Per-path intersection of every context; an empty one refuses, naming both contexts.
  [`stand_in_data.go:567`](../../folio-go/stand_in_data.go#L567)

- The demands, each measured against an engine refusal rather than asserted.
  [`stand_in_data.go:160`](../../folio-go/stand_in_data.go#L160)

**Completeness — a missed path is a hard failure in the never-fail mode**

- `[]` for every collection, and why that means no `Column.Bind` is ever evaluated.
  [`stand_in_data.go:468`](../../folio-go/stand_in_data.go#L468)

- Function rules keyed by name; the key set is what the registry gate compares.
  [`stand_in_data.go:208`](../../folio-go/stand_in_data.go#L208)

- Element-kind rules — the second axis, refusing unknown kinds instead of dropping them.
  [`stand_in_data.go:302`](../../folio-go/stand_in_data.go#L302)

- Sorted placement makes a leaf/branch collision reportable rather than an overwrite.
  [`stand_in_data.go:619`](../../folio-go/stand_in_data.go#L619)

**The boundary — bytes on the channel that already existed**

- Returns bytes only; no revision, because the dispatch arm's snapshot carries it.
  [`engine.go:83`](../../folio-go/wasm/engine.go#L83)

- The silent third list: without this arm the operation cannot work at all.
  [`engine-client.ts:162`](../../folio-designer/src/engine-client.ts#L162)

- The only line that differs from a normal render: what goes on the data channel.
  [`App.tsx:607`](../../folio-designer/src/App.tsx#L607)

**Withholding the claim — the screen must describe the BYTES**

- Follows the installed record, falling back to screen state only before the first render.
  [`App.tsx:1947`](../../folio-designer/src/App.tsx#L1947)

- Stale and stand-in COMPOSE; the export is where these bytes leave the machine.
  [`App.tsx:1855`](../../folio-designer/src/App.tsx#L1855)

- Any fabricated condition, not just `visibleIf` — D-13.4.2's correction.
  [`App.tsx:1961`](../../folio-designer/src/App.tsx#L1961)

- Dashed, because the render proceeded; solid is the failed-render grammar.
  [`App.css:475`](../../folio-designer/src/App.css#L475)

**Tests**

- Exhaustiveness axis one: derived from `LegalFunctionNames()`, restating no names.
  [`stand_in_data_test.go:82`](../../folio-go/stand_in_data_test.go#L82)

- Axis two: AST-extracts `closedElementTypes` from the format's own file.
  [`stand_in_data_test.go:527`](../../folio-go/stand_in_data_test.go#L527)

- Both intersection outcomes: null shared, and a refusal that names both contexts.
  [`stand_in_data_test.go:258`](../../folio-go/stand_in_data_test.go#L258)

- The two-thirds of the band walk nothing exercised until step-04.
  [`stand_in_data_test.go:505`](../../folio-go/stand_in_data_test.go#L505)

- Eight designer behaviours, including both directions of the disclosure.
  [`App.test.tsx:6721`](../../folio-designer/src/App.test.tsx#L6721)

## Delivery Log

### 2026-09-07 — done

Baseline `b04767a`. Shipped at `ea7f31d` — 15 files, +2522/−17. Preview no longer refuses without
sample data: a new read-only Go projection emits a stand-in document for the current template, a wasm
op exposes it, and the designer sends those bytes on the `data` channel that already existed. `Render`
is therefore called with genuinely supplied data and its semantics are untouched **structurally, not
by discipline** — the property that makes the scope fence hold rather than merely be promised.

**The story's shape changed twice under measurement, before any design was done.** First: there is no
single value that renders empty. `upper`/`lower` reject null and never coerce, `formatNumber` rejects
null *and* `""`, `formatDate` accepts neither and has no empty form at all, and a null under a
condition returns false — silently *deleting* the element (D-3.2.3). That killed the obvious approach
and sent the route to the owner, who ruled D-13.4.1 route (a). Second: with every collection standing
in as `[]`, the row block is skipped and **no `Column.Bind` is ever evaluated**, so the entire
row-scope surface needs no stand-in at all. Both shrinks came from measurement, not from taste, and
the `[]` choice is mutation-proved load-bearing — a one-element array fails the render.

**Decisions applied:** D-13.4.1 (owner, route (a) and the `params.*` exclusion), D-13.4.2 (the intent
gap, below), D-3.2.3 (null under a condition deletes the element), D-000.4 (anchors re-derived by
symbol), D-000.9 (no guard that cannot fail), D-11.2.7 (the verification-gap layer runs alone).

**Triage: one intent gap, twelve patches, five defers, two rejects. No loopback;
`review_loop_iteration` stayed 0.** Three review layers ran.

**D-13.4.2 — the intent gap was the orchestrator's error, not the builder's.** The approval clause
keyed the conditional disclosure on `visibleIf` alone, but `if(cond, …)`'s first argument is fabricated
too — so an `if()`-only template rendered a literal branch chosen by no data with the disclosure
withheld, and the frozen matrix row made that *correct* behaviour a test failure. The builder refused
to soften it to a patch on its own authority even though the fix is a single predicate, because that
would have edited a frozen block to match code. Record the principle, not just the outcome: **who
decides is a different question from what gets decided, and cost asymmetry is an argument to present,
not an argument to decide with.** The human amended the block and chose the cheap remedy.

**The rename alone would have left a vacuous test.** The negative fixture had zero components and so
could not distinguish an `if()`-only template from a condition-free one — it was passing while
asserting nothing. The amendment additionally required that fixture to carry a real non-conditional
binding. A re-derivation that restores the rename without the fixture restores the vacuum.

**The most serious finding: the stand-in PDF could be exported under wording identical to a production
export** — `Saved PDF of revision 1 as statement.pdf`. All three review layers found it independently.
The frozen Boundaries enumerate the heading and the viewer label but *not* the export, and the builder
asked rather than assuming the fence. Ruled in scope: the Intent governs, the Boundaries are examples
rather than a schedule, and the export is the one place fabricated bytes outlive the session on disk
with nothing on the file to say what they are. Staleness and stand-in now **compose**; a later change
that makes them exclusive silently drops one of two true claims.

Review also caught the disclosure being affirmatively *wrong* — it keyed on current state while the
label and digest keyed on the bytes, so after clearing sample data one screen asserted "every bound
value is a stand-in" over real-data bytes beside a digest line correctly saying otherwise — and
"every bound value" was false whenever `params.*` is referenced, which D-13.4.1 excluded. Both fixed.

**One guard is honestly unproven, and this spec says so rather than implying otherwise:** `walk()`'s
unknown-AST-node refusal is unreachable while `internal/expr` has exactly four node kinds. Proving it
would mean adding a node type to the engine. The element-kind half of the same axis IS mutation-proved.

**A trap for the next designer story.** `App.test.tsx` carries a self-counting guard that reads its own
source and counts every `it(` from `describe('Story 17.1'…)` **to end of file**. Appending any new
`describe` block *after* that one silently breaks the count. This story's new block was safe only
because it sits above it. Add new designer describes before the Story 17.1 block, or re-pin the count.

**Deferred (five, all filed, all owner-unassigned, all LOW/severity-as-filed):** the stand-in date
spelled twice across two languages with nothing tying them; the no-data notice never announced to
assistive technology; the projection refetched on every no-data render though it is pure; the
"no document loaded" refusal path untested; and a wasm dispatch guard sitting in a `js/wasm` file no
CI job executes. Registered as DW-276…280 — **but see the numbering collision reported at close**: an
owner-raised deferral sourced to Story 13.1 landed in this same commit under the number DW-276, so two
distinct entries share it. Nothing in the repository cites DW-276 by number, so renumbering is free;
the coordinator places register entries and owns the fix.

**Measured gates, confirmed at `ea7f31d` (tree clean before and after):** designer `npm test` **65
files / 1002 tests, 0 failing** (baseline 993 — nine added, no new vitest file); `go test -count=1
./...` in `folio-go` **2273 pass / 2 fail / 5 skip**, failing only `TestCorpusMeetsP6ExerciseFloors`
and its `P6g_(opaque_names)` subtest by enumerated name, with no third failure; `lint` `go test
-count=1 ./...` **four packages ok, 227 passed** — `-count=1` mattered here, because this story adds a
Go file to the tree the rules package walks with `ReadDir`, which the test cache does not track;
`tsc -b --force` exit 0, empty; `oxlint` **exactly 4** `only-export-components` warnings, 0 errors, at
`preview/pdf-viewer.tsx:16,17` and `App.tsx:3745,3752`; `npm run test:e2e:compile` exit 0; `gofmt -l`
over absolute `folio-go` and `lint` paths empty. `npm run build` was deliberately not run.

**Not run locally:** the Playwright suite itself, including this story's new `preview-no-data.spec.ts`.
It is compile-checked here only. Since `adf905a` CI runs the unfiltered `-tags=matrix` suite and all
Playwright tests on every push (DW-268 discharged), so it executes per-commit there — reported as
executed-in-CI, never as a compile-only placeholder.

