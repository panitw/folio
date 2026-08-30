---
title: 'Story 7.2: Set the space between a paragraph''s lines'
type: 'feature'
created: '2026-08-30'
status: 'ready-for-dev'
baseline_commit: '02da139273bd9a4ce34874a64d6cadde826321c5'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-context.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
  - '{project-root}/_bmad-output/implementation-artifacts/2-5a-align-first-baseline-with-the-leading-model.md'
warnings: ['oversized'] # the ruling set D-7.2.1..D-7.2.6 is dense, and the format-version debt D-7.2.1 obliges this story to discharge is a second wide surface; both must be stated, not summarised
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative orientation. The contract below governs; where the two differ, the contract wins.*

A template author can now set how far apart a paragraph's lines sit. Set it wide and the lines
breathe; set it tight and they close up, as a filed contract is usually set.

The setting changes the distance from one line to the next, and nothing else — it does not
move the first line. The space above the first line comes from how tall the typeface reaches,
not from the spacing choice, and leaving it alone keeps a component's top edge where the author put
it. If spacing moved that too, every multi-line component would shift and every neighbour would
appear to jump. This project made that mistake once and deliberately undid it.

Tight spacing is allowed to be genuinely tight — tight enough that one line's letters reach into the
line below. That is what tight leading is, and the printed page has always drawn it. The preview
used to refuse it and blank out entirely; now it shows what the page shows.

A value outside the accepted range is refused at load, with a message naming the component that
carries it — never quietly rounded into range. A document that sets no spacing
renders to exactly the bytes it renders today, and a document now declares the format version its
content actually needs.

The test suite carries one standing, deliberate failure about fixture coverage; it stays red and is
not a regression.

Out of scope: justified edges, the designer's editing controls, and how pages break.

<intent-contract>

## Intent

**Problem:** A text element's baseline-to-baseline advance is fixed by the font chain, so an author
cannot set a house line spacing; `.folio` has no key for it, and the designer canvas actively refuses
any spacing tighter than the font's own leading by blanking the entire projection. Separately, the
format version has never moved: Epic 10 shipped `style.color` and bumped nothing, so colour-bearing
documents declare `1.0` while requiring `1.1`, and `versionForSave` is still the documented stub
reserved for "the day a future MINOR exists".

**Approach:** Add an optional `style.lineSpacing` — an exact decimal in `0.001`–`1000.0`, carried as
a whole number of thousandths — that scales the vertical model's `Advance` and nothing else,
threaded to every construction site through the existing cascade. Delete the canvas clause that
restates an engine invariant this feature dissolves. Mint a diagnostic code so the load error
actually reaches the author. Implement the reserved version raise path and retrofit `style.color`
onto it.

## Boundaries & Constraints

**Always:**
- `lineSpacing` scales **`Advance` only**. `FirstBaseline` and `LastDescent` are untouched, so a
  multi-line element's top edge does not move and no sibling appears to shift. This is the D-2.5a /
  DW-15 two-model split; Story 2.5a exists solely because the two were once conflated.
- Integer arithmetic throughout (AD-2, AD-3, AD-23): no `float64` under `internal/`, no per-line
  float multiply, no second exported scaling function in `internal/geom`.
- Absent `lineSpacing` produces byte-identical output to today, on all four targets. The existing
  corpus digests are **asserted, not assumed**.
- The load-time range is **representational and carries no typographic opinion** (D-7.2.3): a whole
  number of thousandths in `[1, 1000000]`. **No 1.0 minimum.** The `1000000` ceiling is a *stated
  sanity ceiling, not a derived safety bound* — its constant's comment must say so, and must record
  that the honest overflow-derived ceiling would be ~1023 thousandths, i.e. it would forbid
  `lineSpacing` above 1.0. That reductio is why the ceiling is a sanity bound and nothing more.
- **One validation function**, called from both the load path and the property-command path, so a
  value refused in a file is refused in the inspector for the same reason. Because `internal/template`
  may not import the module root while the root imports it, the function lives in `internal/template`
  and is exported from there.
- **A panic must never be reachable from authored input** (D-7.2.4). `geom.ScaleRound` panics on
  int64 overflow, and a Go panic aborts the package binary — every other test in `folio-go` then
  silently stops reporting, which is a suite-wide blindfold, not a crash.
- `lineSpacing` reaches table cells and `headerStyle` through the existing cascade — **every caller,
  no carve-out** (D-7.1.3). One rule for one property.
- Version is a property of the **document**, raised only by content (D-1.4.13): `lineSpacing`-only
  → `1.1`, neither `lineSpacing` nor `color` → `1.0`; both coexist. **A new document must declare the
  lowest version its content requires, never the library's ceiling.** Never lower a loaded version.

**Block If:**
- The corpus digests move for a document that declares no `lineSpacing`. That is a byte-neutrality
  regression, not a re-record.
- Honouring the range in `[1, 1000000]` turns out to require weakening any surviving canvas clause
  beyond the one D-7.2.2 names.
- The version raise path cannot express "raised by content" without changing what an existing
  `1.0` document serializes to.

**Never:**
- Do not touch `internal/layout/paginate.go`. Pagination reaches line extents through item extents;
  wider lines feed it unchanged. Epic 7 changes nothing about the pagination model.
- Do not implement justification or add `justify` to any align set — Story 7.3, and a MAJOR bump.
- Do not add the inspector control, the field spec, or any designer editing surface for
  `lineSpacing` — Story 7.4. The engine-side property command is in scope; the UI that drives it is not.
- Do not bump `SupportedMajor`. That is Story 7.3's under D-7.3.1 and D-R7.9.
- Do not remove or weaken `paint.advance <= 0`, `paint.baseline < paint.top`, the
  `Number.isSafeInteger` checks, or `paint.top < priorTop + priorAdvance`.
- Do not add a non-panicking variant of `ScaleRound` (AD-2: scaling is one function; a second door
  drifts). Do not export a new function from `internal/geom`.
- Do not fold the zero-advance or overflow errors into the new diagnostic code to save a mint — they
  are different conditions at a different stage.
- Do not close `fontSize`'s missing range check. Record it as deferred; closing it is a
  format-domain decision on a second field and would earn `multiple-goals`.
- Do not "fix" `TestCorpusMeetsP6ExerciseFloors`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Absent (the whole existing corpus) | No `lineSpacing` on any style | `Advance` is byte-for-byte today's value; all recorded digests unchanged on all four targets | No error expected |
| Wider spacing | `"lineSpacing": 1.5`, 12pt chain | `Advance` scaled ×1500/1000; `FirstBaseline` and `LastDescent` unchanged; element top edge unmoved | No error expected |
| Tight leading | `"lineSpacing": 0.6` | Accepted. `FirstBaseline > Advance`; PDF draws overlapping line boxes; canvas projects the same numbers and the snapshot still validates | No error expected |
| Empty line under spacing | 7.1's two consecutive breaks plus `lineSpacing` | The empty line occupies one full **scaled** `Advance`; `FirstBaseline` still unmoved; block height counts it | No error expected |
| Table cell / header label | `lineSpacing` on `style`, on `headerStyle`, or cascading from `style` to header | Applies through the existing cascade at every construction site — no carve-out | No error expected |
| Below range | `"lineSpacing": 0` or `-1` | Refused at load | Located load error naming the element, code `STYLE_LINE_SPACING_INVALID` |
| Above range | `"lineSpacing": 1000.001` | Refused at load | Same code, same located shape |
| Not a whole thousandth | `"lineSpacing": 1.0005` | Refused at load | Same code; more than three decimal places |
| Resolved advance of zero | Tiny face × tiny ratio, e.g. scaling 400 millipoints by 1/1000 | Refused where both operands exist, in the leading model | Located error naming the element **and the resolved size**; a distinct condition, not the range code |
| Overflow at the scaling site | `Advance` × ratio exceeds int64 | Refused before `ScaleRound` is called | Located error; the panic is never reached |
| Inspector sets the value | Property command carrying `lineSpacing` | Same validation function, same refusal reason as the file path | Same rejection, surfaced through the command's error |
| Document raises its version | A document that introduces `lineSpacing` (or, retrofitted, `color`) | Serializes declaring `1.1` | No error expected |
| Document needs nothing new | A document using neither | Still serializes `1.0`; never raised to the library's ceiling | No error expected |

</intent-contract>

## Code Map

### The vertical model — the one-line change, and the threading

- `folio-go/wrap.go` -- `verticalMetrics` at **:513-533** (`FirstBaseline` :515, `Advance` :518,
  `LastDescent` :532). `verticalModel(chain, metrics, fontSize)` at **:584**; `chainVerticalModel`
  at **:635**; `lineAdvance` at **:646-652** (a projection, not a second derivation — it also calls
  `chainVerticalModel` and must thread the ratio). The three maxima are independent at **:600-611**
  (`maxDescent` negated at :605); the zero clamp *is* the zero initialisation at :600 — they are only
  ever raised by `>`. `units` sum + `units <= 0` guard at **:613-618**. The scale closure at
  **:620-622**. **THE SITE: `Advance: scale(units)` at :626**, inside the return literal :624-628
  next to `FirstBaseline: scale(maxAscent)` (:625) and `LastDescent: scale(maxDescent)` (:627).
  Applying the ratio at :626 and only there keeps the split correct *by construction*.
  Neither constructor takes a style, so the ratio must be threaded as a parameter exactly as
  `fontSize` already is.

### The four production construction sites (style is in scope at all four)

- `folio-go/render.go:848` -- text element. `fontSize` extracted at **:760-763** off `el.Style` — the
  threading precedent to copy verbatim.
- `folio-go/table_render.go:675` -- header labels. `resolvedHeaderStyle` struct at **:268-**,
  `fontSize` field **:272**, populated by `resolveHeaderStyle` at **:315-321** with the three-way
  cascade `headerStyle` → `style` → default. **The cleanest precedent: add a `lineSpacing` field with
  the same cascade.** Note this site reads `FirstBaseline`/`LastDescent` only and never `Advance`.
- `folio-go/table_render.go:800` -- body **and** footer share ONE model. `bodyFontSize` at
  **:782-785** is read directly off `el.Style`, *not* through `bs` (`resolveBodyStyle` at :786;
  its doc at :377 says fontFamily/fontSize are deliberately not carried in `bs`).
- `folio-go/page_setup.go:460` -- canvas projection. `fontSize` at **:442-445**, same pattern.

### The drift trap — `Advance` re-multiplied into a height in three longhand copies

- `folio-go/text_alignment.go:88` -- `textBlockHeight` = `(lines-1)*vm.Advance + FirstBaseline +
  LastDescent`. Only **two** callers: `render.go:873`, `page_setup.go:474`.
- `folio-go/table_render.go:914` -- data `rowHeight`, longhand, does **not** reuse the helper.
- `folio-go/table_render.go:1147` -- `footerRowHeight`, longhand, identical expression.
- **HAZARD:** `folio-go/table_render.go:692` declares a *local variable* `textBlockHeight :=
  vm.FirstBaseline + vm.LastDescent` that **shadows the package-level function** for the rest of that
  block (used :696, :698). It is the one-line case so it carries no `Advance` term, but any refactor
  that tries to call the helper inside that scope will not compile.
- Because the ratio is applied inside the model, all three copies inherit it automatically. **Assert
  that they do**, rather than assuming: a future reader must be unable to reintroduce the drift.

### Four `origin + i*Advance` stepping loops

- `folio-go/render.go:876`; `folio-go/table_render.go:998`; `folio-go/table_render.go:1175`;
  `folio-go/page_setup.go:476` via `canvasLineTop` (**:530-535**, which already carries its own
  overflow guard against `MaxCanvasMillipoints` at :531 and explicitly tolerates `advance == 0`).

### Pagination — reached through item extents; `paginate.go` needs NO edit

- `folio-go/render.go:892` `placed[j].itemTop = lineY`; **:893** `itemBottom = lineY + FirstBaseline
  + LastDescent` — **note there is no `Advance` term**: a line's own extent is ascent+descent, so
  spacing changes *where lines sit*, not how tall each item is. That is precisely how AC5 is
  satisfied without touching layout. Field decl :128-133. Flow: `render.go:2034-2035` and
  `page_number.go:93-94` → `internal/layout/paginate.go:105-125` (`Top, Bottom` at :111).
  `paginate.go` contains no reference to `Advance`, `lineAdvance` or `verticalMetrics`.
- Table extents: `table_render.go:717-718` (header), `:1030-1031` (data), `:1206-1207` (footer).

### Template model / parse / serialize

- `folio-go/internal/template/model.go` -- `Style` at **:250-268** (`FontSize` :263); `Element.Style`
  **:199**; `TableExt.HeaderStyle` **:229**. `Presence[T]` at `presence.go:16-20`, constructors
  `present` :23 / `presentNull` :28 / `absent` :34.
- `folio-go/internal/template/parse_bands.go` -- `decodeStyle` **:517-627**, signature
  `(elementID, raw, fieldPrefix)`. Numeric precedent (`fontSize`) at **:595-602** via
  `decodePointsRaw`. `consumed` map at :521 and `extraFields(obj, consumed)` at :620 — **a new key
  must be added to `consumed` or it round-trips opaquely through `Extra` and is silently ignored.**
  Call sites: element **:204** (`"style"`), headerStyle **:338** (`"headerStyle"`).
- `folio-go/internal/template/parse.go:427-438` -- `decodePointsRaw` → `decodeNumberRaw` →
  `decodePoints` (`decimal.go:30`), the exact-decimal path that rejects >3 decimal places and
  int64 overflow. This is the ≤3-decimal discipline `lineSpacing` inherits.
- `folio-go/internal/template/serialize.go` -- `writeStyle` **:333-377**; `writeObject` sorts all
  keys at **:42**, so emission order is irrelevant. `versionForSave` is called at **:118**.

### Overflow, scaling, zero advance

- `folio-go/internal/geom/scale.go` -- `ScaleRound` at **:59**, round-half-to-even at :87-108.
  **Panics:** `den == 0` :60-62, `den == minInt64` :63-65, **`v*num` overflow :68-70**, quotient
  overflow :73-75.
- **`int64MulOverflows` already exists at `internal/geom/scale.go:12` but is UNEXPORTED**, so the
  root package cannot call it. `internal/geom/scale_surface_test.go:26`
  (`TestExactlyOneExportedScalingFunction`, `want := []string{"ScaleRound"}` at **:68**) reddens on
  **any** new exported function in the package. Therefore: **do not export it.** Put the
  root-package overflow predicate next to the guard that uses it, in `wrap.go`. AD-2 forbids a
  second *scaling* door; a local overflow predicate is not one.
- **Zero-advance, verified by reading the code:** `ScaleRound(400, 1, 1000)` → `q = 0`, `r = 400`,
  `half = 600`, `400 > 600` false and `400 == 600` false → **returns 0**. Generally `ScaleRound`
  returns 0 whenever `v*num < den/2`. There is **no existing guard anywhere** against
  `vm.Advance == 0`.

### Diagnostics — what minting costs, from the `STYLE_COLOR_INVALID` precedent

- Production: `folio-go/internal/diag/diag.go` — const **:149**, `allCodes` entry **:240**,
  `dispositions` entry **:273** (`DispositionError`). Public bridge:
  `folio-go/diagnostic.go:243`. `SeverityError` lives in the **root** package at `diagnostic.go:68`.
- Test pins (both mandatory): `internal/diag/diag_test.go:41` (`codePins`, consumed by
  `TestRegistryIsAdditiveOnly` :49) and `folio-go/diag_bridge_test.go:43` — the bridge test parses
  the package's own sources, so a bridge const with no pin fails.
- **Census:** `folio-go/diagnostic_registry_census_test.go`, `TestDiagnosticRegistryErrorCensus` :17.
  A new Error code needs a **real production trigger** in the `triggers` map (:24-77) — the arity
  check at **:82-84** makes a code without a trigger fatal — and the trigger must fail, transport as
  `*RenderError`, carry `SeverityError`, match the code, have a non-blank message, and be **located**
  (`ElementID != "" || DataPath != ""`, **:109-111**). The only located-ness exemptions are
  `CodeTemplateMalformed` and `CodeDocumentDateInvalid`; the new code gets neither.
- **Why the mint is what delivers AC7:** `folio-go/internal/template/errors.go:52-54` `newLoadError`
  sets **no** `Code`; `folio-go/render_error.go:97-106` maps an uncoded `LoadError` to
  `DiagCodeTemplateMalformed`; and `folio-go/wasm/cmd/engine/main.go:276-281` `reportableMessage`
  replaces the message with "The template could not be processed" **only** for
  `DiagCodeTemplateMalformed` — everything else gets `bounded(message, 512)`. So an uncoded
  `lineSpacing` error is destroyed before the author sees it. **Use `newLoadErrorCoded`
  (`errors.go:56`)**, the coded variant that already exists (3 call sites, all footer-source).

### Format version

- `folio-go/internal/template/version.go` -- `SupportedMajor = 1` **:24**, `SupportedVersion = "1.0"`
  **:25**; `parseVersion` :32-47; `checkVersionLoadable` :53-64 (higher MAJOR errors, **higher MINOR
  loads**); **`versionForSave` doc :66-84, body :85-87 returning `loaded` unchanged**, its own comment
  naming the reserved raise path. **Exactly one caller: `serialize.go:118`.**
- `folio-go/internal/template/version_test.go:49` and **:77** pin "version must not be lowered to %s
  on save" — the raise path must keep that true.
- **All 18 fixtures declare `"1.0"`** (`fixtures/*/input.folio`); also
  `folio-go/testdata/template/golden/worked-example.json` and `testdata/example/first-pdf.folio`.

### The property-command path (engine side only — the UI is 7.4's)

- `folio-go/component_commands.go` -- `applyPropertyChanges` **:894-1157**. `propertyOrder` **:896**
  and `propertyPath` **:836-845** must both gain the key. Text/table allow-list arm **:911-913** (the
  `style.color` precedent). Unknown-key rejection :915-921. Numeric arm **:934-991** with
  `propertyLength` :944 and the `length <= 0` positivity check :948-950 — note `propertyLength` →
  `lengthField` (`page_setup.go:845-869`) is a **points** decoder bounded by `MaxCanvasMillipoints`
  and is **not** suitable as-is for a ratio. `styleFor` :886-891. **`cleanupEmptyStyle` :1168-** must
  learn about a `lineSpacing`-only style block or it will strand an empty `style`.
- Shared-predicate precedent: `validPropertyColor` **:1159-1162**. Mirror the *shape*, not the
  location: `style.color` is validated at render in the root package, which a load-time-validated
  key cannot copy.

### The designer canvas — the clause to delete and the test that pins it

- `folio-designer/src/engine-protocol.ts` -- `isTextPaint` **:199-216**. **Line :208 carries EIGHT
  clauses**, including `paint.baseline > paint.top + paint.advance`. Deleting it is a **sub-line
  edit**; the seven survivors on that line stay byte-for-byte. Chain: **:162** (`isCanvas` rejects
  the component) → :113 (`isCanvas`) → **:217** (`isSnapshot`) → :239 (`parseInbound` drops the whole
  response). One bad line blanks the **entire** projection.
- Go side confirms the reduction: `page_setup.go:474` `originY`, **:476** `top = originY + i*Advance`
  via `canvasLineTop`, **:485** `baseline = top + FirstBaseline`, **:489** `advance = vm.Advance`.
  So the clause reduces to `FirstBaseline <= Advance` — an engine invariant restated on the browser's
  side of the channel. `paint.top < priorTop + priorAdvance` evaluates to `originY+i·A < originY+i·A`,
  **false for any positive advance**, so it does not become the next cliff.
- **The test that must be updated:** `folio-designer/src/engine-protocol.test.ts:74`
  `it('rejects non-advancing or out-of-box text paint geometry')` — **:81** asserts
  `{...line, baseline: 13}` (with `top: 0, advance: 12`) is rejected. That assertion is exactly the
  behaviour being removed; it must become the inverse — tight leading is **accepted** — and the
  other rejections on :80, :82, :83 must stay.
- Hand-mirrored Go constants (D-7.4.5 / DW-25, unchanged by this story):
  `engine-protocol.ts:201` ↔ `page_setup.go:27`; `:213` ↔ `page_setup.go:28` and `:26`.

### Test fences a new format key trips

- `folio-go/internal/template/numeric_classification_test.go` -- kinds at **:16-19** (only
  `numericKindPoints` and `numericKindPlainInt`; the comment at :11-15 says "There is no third
  kind"). `numericFieldRegistry` **:27-39**. `extractNumericDocKeys` **:64-71** derives the key list
  from `folio-format.md` fences **at test time**, and its comment names `"lineSpacing"` as the exact
  hypothetical twelfth key. **The moment `lineSpacing` appears with a numeric literal in any fence,
  `TestNumericFieldClassificationInventory` (:100-135, unclassified fatal at :122) goes red.**
- `folio-go/internal/template/drift_test.go` -- `TestDriftGoToDoc` **:232-263** and `TestDriftDocToGo`
  **:269-300**: bidirectional, **no allowlist**. `normaliseDocToken` **:216-225** reduces paths to
  the bare last segment. `TestDriftASTMatchesRuntimeEmission` **:402-450** is what actually enforces
  `maximalFixture` coverage (failure text at :445).
- `folio-go/internal/template/fixtures_test.go` -- `maximalFixture` **:236-**, style blocks at
  :294-316, :334-336, :349-352, :367-370. **`style.color` escaped these fences only because
  `border.color` (:277, :299) already emitted the bare key `color`.** `lineSpacing` has **no such
  shadow** — no other key ends in `lineSpacing` — so it will be caught by every fence and must be
  added explicitly.
- `folio-go/internal/template/goldenfixture_test.go:16-25` -- `TestWorkedExampleMatchesGoldenFixture`;
  `extractWorkedExampleFence` :30-47 takes the **last** ```json fence under `## Worked example`
  (`folio-format.md:512`). Golden: `folio-go/testdata/template/golden/worked-example.json`, which
  **contains three `style` blocks** (:28, :52, :71) and declares `"version": "1.0"` at **:106**.
  Deliberately **left untouched** by this story — but if anything does edit that fence, fence and
  golden move together in one commit, and the declared version moves with the content.
- `_bmad-output/specs/spec-folio/folio-format.md` -- style keys **:293-342** (defaults table
  :312-325); version rules **:67-71**; sorted-keys rule :64-65.

### Golden fixture registration (four places) and the corpus baseline

- `folio-go/byte_neutrality_test.go:92-346` `goldenDigestRecord`; `folio-go/matrix_test.go:1281-`
  `matrixDocuments`; `folio-go/missing_glyph_corpus_test.go:41` corpus table **plus** the second
  declaration in `beyondBaselineAcceptance` at **:130-132**; and
  `.github/workflows/matrix.yml:239` (`docs=` list) plus the four per-target upload paths at
  :69-70, :106-107, :143-144, :180-181 (`if-no-files-found: error`).
- **Do not add to** `baselineAcceptanceFixtures` (`first_baseline_acceptance_test.go:100`), which is
  hard-pinned to exactly five at :258-259.
- Precedent to copy: `fixtures/mandatory-break/` (Story 7.1) with
  `folio-go/mandatory_break_template.go:24` kept byte-identical to `input.folio`, asserted by
  `mandatory_break_fixture_test.go:406-421`.
- **Recorded baseline digests to assert unchanged** — `folio-go/byte_neutrality_test.go`:
  `statement-1` :232 `114df1d6…` (76,744 B); `statement-5` :246 `70dce051…` (127,363 B);
  `statement-20` :265 `56bfbbd9…` (269,884 B); `statement-50` :281 `5d090b0f…` (555,829 B);
  `mandatory-break` :339 `7cf743de…` (56,681 B). The record "is invalidated IN WHOLE if any one
  moves" (:224-225).
- `folio-go/matrix_test.go:69-74` `matrixTargets`; `TestTargetRenderHash` **:1758-** with its
  `FOLIO_MATRIX_TARGET` no-op guard at **:1759-1770**; `TestCrossTargetByteIdentity` **:1581-**.

### Deferred-work records

- `_bmad-output/implementation-artifacts/deferred-work.md` -- `## Open` at :181. **DW-24 at
  :183-321**, whose trigger list at **:204-207 names Story 7.2 by name**. DW-25 at :1707-1777.
  DW-15 at :56-153 (the entry that unified the two models — read it). **Highest number in use is
  DW-25, so a new entry is DW-26.** Neither `fontSize`'s missing range check nor `ScaleRound`
  overflow is recorded anywhere in the file today.

## Tasks & Acceptance

**Execution:**

- `folio-go/internal/template/model.go` -- Add `LineSpacing` to `Style` as a `Presence` of a whole
  number of thousandths (an int64 count, **not** a `geom.Length` — it is dimensionless). -- Rationale:
  the key must exist on both attachment points, which it does automatically via `Style`.
- `folio-go/internal/template/linespacing.go` (new) -- Add the **one** exported validation function
  for a thousandths ratio (range `[1, 1000000]`) with named bound constants. It lives in
  `internal/template` because that package may not import the module root while the root imports it,
  so this is the only direction in which both the load path and the property-command path can reach
  the same function. Write the ceiling's comment to say it is a **stated sanity ceiling, not a
  derived safety bound**, recording that the honest overflow-derived ceiling would be ~1023
  thousandths and would therefore forbid `lineSpacing` above 1.0. -- Rationale: D-7.2.3; the comment
  is what stops the ceiling being re-proposed as typography.
- `folio-go/internal/diag/diag.go` + `folio-go/diagnostic.go` -- Mint `STYLE_LINE_SPACING_INVALID`
  (const, `allCodes`, `dispositions` as `DispositionError`) and the bridge
  `DiagCodeStyleLineSpacingInvalid`. -- Rationale: D-7.2.5; this is the mechanism that delivers AC7.
- `folio-go/internal/template/parse_bands.go` -- Decode `lineSpacing` in `decodeStyle` through the
  existing exact-decimal path (≤3 decimal places, `decodePointsRaw`'s discipline), convert to
  thousandths, call the shared validator, and add the key to `consumed`. On refusal return
  `newLoadErrorCoded` with the new code, located via `fieldPrefix+".lineSpacing"` and the element id —
  so the `headerStyle` case locates as `headerStyle.lineSpacing`, not `style.lineSpacing`. --
  Rationale: AC7; the `consumed` entry is what stops silent `Extra` passthrough.
- `folio-go/internal/template/serialize.go` -- Emit `lineSpacing` from `writeStyle` in the same
  exact-decimal form it was authored in. -- Rationale: AD-9 round-trip; also required by
  `TestDriftGoToDoc`.
- `folio-go/internal/template/version.go` -- Implement the raise path in `versionForSave`: derive the
  version the **document's content** requires (`1.1` if any style anywhere sets `lineSpacing` **or**
  `color`; `1.0` otherwise), return the higher of that and `loaded`, and never lower. Update
  `SupportedVersion` to the library's new ceiling and **correct its doc comment**, which currently
  claims it is what the library would author for a brand-new document — that framing is exactly what
  D-1.4.13's raise-only-by-content rule forbids. Leave `SupportedMajor = 1`. -- Rationale: D-7.2.1;
  this discharges the `style.color` retrofit in the same function.
- `folio-go/wrap.go` -- Thread the ratio into `verticalModel`, `chainVerticalModel` and `lineAdvance`
  as a parameter, exactly as `fontSize` is threaded. Apply it at **:626 only**. Before calling
  `ScaleRound`, guard `int64MulOverflows(Advance, ratio)` with a **root-package, unexported**
  predicate (do not export anything from `internal/geom`) and return a located error. Guard a
  resolved `Advance == 0` where both operands exist, as a located error naming the element **and the
  resolved size**. -- Rationale: D-7.2.4, D-7.2.3; a panic must never be reachable from authored input.
- `folio-go/render.go` -- Extract `lineSpacing` off `el.Style` beside the existing `fontSize`
  extraction at :760-763 and pass it to the model at :848. -- Rationale: text-element site.
- `folio-go/table_render.go` -- Add a `lineSpacing` field to `resolvedHeaderStyle` with the same
  three-way cascade `resolveHeaderStyle` already uses for `fontSize` (:315-321); pass it at :675.
  Extract the body value off `el.Style` beside `bodyFontSize` (:782-785) and pass it at :800, which
  serves body **and** footer. -- Rationale: D-7.1.3, every caller, no carve-out.
- `folio-go/page_setup.go` -- Extract `lineSpacing` beside the `fontSize` pattern at :442-445 and
  pass it at :460, so the canvas consumes the identical advance the renderer does. -- Rationale: AC6,
  the Story 5.9 invariant.
- `folio-go/component_commands.go` -- Add `lineSpacing` to `propertyOrder` and `propertyPath`, admit
  it on the text/table arm, decode and validate it through the **same** shared validator (not
  `propertyLength`, which is a points decoder bounded by `MaxCanvasMillipoints`), and teach
  `cleanupEmptyStyle` about a `lineSpacing`-only style block. -- Rationale: D-7.2.3's "refused in a
  file is refused in the inspector for the same reason".
- `folio-designer/src/engine-protocol.ts` -- Delete **only** `paint.baseline > paint.top +
  paint.advance` from :208, leaving the other seven clauses on that line byte-for-byte. -- Rationale:
  D-7.2.2; keeping it is the AD-17 violation.
- `folio-designer/src/engine-protocol.test.ts` -- Invert the :81 assertion so tight leading is
  **accepted**, keeping :80, :82 and :83 as rejections, and add a case proving a whole snapshot with a
  tight-leading line still validates. -- Rationale: the deleted clause must be pinned as deliberately
  gone, not silently untested.
- `_bmad-output/specs/spec-folio/folio-format.md` -- Document `lineSpacing` in the style section
  (:293-342): put it in the **JSON fence at :297-310** with a numeric literal, and in the defaults
  table at :312-325, with its range, its absent-default and its exact-decimal rule. That one edit
  satisfies both drift tests and puts the key in front of the numeric registry. **Do not add
  `lineSpacing` to the worked example** at :512 — its golden declares `"version": "1.0"`
  (`worked-example.json:106`), so adding the key would force the fence and the golden to `1.1` in
  lockstep for no test benefit. -- Rationale: the fences are satisfied by the style section alone.
- `folio-go/internal/template/numeric_classification_test.go` -- Register `lineSpacing`. Neither
  existing kind honestly describes a dimensionless ratio, so **add a third kind** and amend the
  :11-15 comment rather than mislabel it as points. -- Rationale: the registry's value is that its
  claims are true; :122 is fatal for an unclassified key.
- `folio-go/internal/template/fixtures_test.go` -- Add `lineSpacing` to `maximalFixture`'s style
  blocks. -- Rationale: it has no `border.color`-style shadow, so it will be caught.
- `folio-go/diagnostic_registry_census_test.go` -- Add the census trigger for the new code. --
  Rationale: the arity check at :82-84 makes this mandatory, and located-ness is asserted.
- `fixtures/line-spacing/` + its Go template const + `folio-go/..._fixture_test.go` -- Ship the
  story's golden fixture following `fixtures/mandatory-break/`, and register it in **all four**
  places (`goldenDigestRecord`, `matrixDocuments`, the missing-glyph corpus table **and**
  `beyondBaselineAcceptance`, and `.github/workflows/matrix.yml`'s `docs=` list plus its four upload
  paths). Do **not** add it to `baselineAcceptanceFixtures`. -- Rationale: Epic 7's "every feature
  ships its golden fixture".
- `folio-go/` tests -- Add direct assertions for: `FirstBaseline` and `LastDescent` **unchanged**
  while `Advance` scales; a multi-line element's top edge unmoved; 7.1's **empty line occupying one
  full scaled advance** and `textBlockHeight` counting it, which is what `textValignOffset`
  distributes slack against; all **three** longhand block-height copies inheriting the ratio; the
  canvas projecting the same advance the renderer uses; the new code being **not**
  `TEMPLATE_MALFORMED` and its message surviving `reportableMessage` intact; and the corpus digests
  unchanged. -- Rationale: each is a guard the story would otherwise only satisfy by construction.
- `_bmad-output/implementation-artifacts/deferred-work.md` -- Add **DW-26** recording `fontSize`'s
  missing range check, naming the shared overflow site, in the file's existing entry style. Amend
  **DW-24** to record that Story 7.2 fired its trigger and declined on the narrow ground that **no
  fixture declares `valign` at all** — not "a different call site" — and that it is **not deferrable
  a third time**: Story 7.3 treats it as an acceptance criterion. -- Rationale: D-7.2.4, D-7.2.6.

**Acceptance Criteria:**

- Given a document with no `lineSpacing`, when the whole corpus is re-rendered on all four targets,
  then every recorded digest is unchanged — asserted against the five recorded values above, not
  assumed.
- Given `lineSpacing` on an element's style, when the vertical model is computed, then `Advance`
  scales and `FirstBaseline` and `LastDescent` are bit-identical to the unscaled model.
- Given a multi-line element, when `lineSpacing` changes, then the element's top edge does not move.
- Given `lineSpacing` below 1, above 1000000, or carrying more than three decimal places, when the
  template is loaded, then loading fails with a located error naming the element and carrying
  `STYLE_LINE_SPACING_INVALID`, and that error's message reaches the designer unreplaced.
- Given a font size and ratio whose product would overflow int64, when the leading model runs, then a
  located error is returned and `ScaleRound` is never called — no panic is reachable from a template.
- Given a resolved advance of zero, when the leading model runs, then a located error names the
  element and the resolved size, and it is **not** the range code.
- Given the same value arriving through the property command, when it is applied, then it is refused
  for the same reason and by the same function as the load path.
- Given a document that introduces `lineSpacing` or `color`, when it is serialized, then it declares
  `1.1`; given one that uses neither, then it still declares `1.0`; and no load ever lowers a version.
- Given a canvas line whose baseline sits below the next line's top, when the snapshot is validated,
  then it is accepted and the projection is not blanked.

## Spec Change Log

## Review Triage Log

## Design Notes

**Why the ratio goes in at exactly one line.** `verticalModel` builds `FirstBaseline`, `Advance` and
`LastDescent` from three independent maxima and scales each separately at `wrap.go:625-627`. Putting
the ratio on `:626` alone makes "scales `Advance` and nothing else" true *by construction* rather
than by discipline. The risk is entirely downstream: `Advance` is re-multiplied into a height in
three longhand copies (`text_alignment.go:88`, `table_render.go:914`, `:1147`) and stepped in four
independent loops. They all inherit the ratio automatically **because** it is applied inside the
model — which is the argument for applying it there and asserting the inheritance, rather than
scaling at any consumer.

**Why tight leading must be allowed.** `FirstBaseline > Advance` means one line's baseline sits below
the next line's top: the line boxes overlap. That *is* tight leading, and it is what the PDF draws.
The canvas clause forbidding it was the browser refusing the engine's own honest measurement and
blanking the whole projection — the AD-17 invariant inverted. Measured on the shipped chain at 12pt
(`FirstBaseline: 11759`, `Advance: 14982`) the cliff is 784 thousandths rejects / 785 passes, and it
**moves with face and size**, which is exactly why it can never be a load-time bound.

**Why the load-time range carries no typographic opinion.** A load-time check cannot see the font
size, so a bound pretending to be typographic while blind to the input that determines it is wrong at
every size but one. The range is the value's own domain and nothing more. The two failures that *are*
typographic — zero resolved advance and int64 overflow — are checked where both operands exist, in
the leading model, as separate conditions with their own errors. Raising the minimum to prevent the
zero case would only move the blindness.

**Worked range.** Authored `1.5` → 1500 thousandths → `Advance × 1500 / 1000`. Authored `0.001` → 1,
the floor. Authored `1.0005` → refused, four decimal places. Authored `1000.0` → 1000000, the stated
sanity ceiling.

**DW-24, declined on the correct ground.** Story 7.2 *does* change the input to the unexercised
rounding site: `textBlockHeight` is built from `Advance` and feeds the slack that
`ScaleRound(slack, 1, 2)` halves for `valign: middle`. The exposure is unchanged only because **no
fixture declares `valign` at all** — the very absence DW-24 exists to record. That, not "a different
call site", is the ground for declining.

## Verification

7.2's correctness is byte-identity-shaped, so it carries the heavy tests regardless of the per-epic
cadence. Report measured pass/fail counts, never "green".

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: **exactly ONE** failure,
  `TestCorpusMeetsP6ExerciseFloors` (P6g got 7, need >=20), the mandated permanent red. Never touch
  it. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `gofmt -l folio-go` -- expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- expected: run **once per
  leg** with `FOLIO_MATRIX_TARGET` set (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`; the
  list is `matrix_test.go:69-74`). **Unset, this test logs "asserts NOTHING" and a no-op is not a
  pass** — name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: pass; the
  all-four-in-one-process local gate.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: pass (4
  pre-existing `only-export-components` lint warnings are not a regression).

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` fails under
`-tags=matrix` because `fontTools` is not installed here; `lint/internal/rules/licencegraph_test.go`
is not gofmt-clean (DW-23, owned by Story 15.2).

**Delivery Log obligations (for the closer):**
- The five corpus digests must be reported as **measured unchanged**, not assumed.
- DW-24 fired its trigger here and was declined on the narrow ground above; it is **not deferrable a
  third time**.
- The `style.color` version retrofit changes what colour-bearing documents serialize as. **Measured
  at planning: no fixture under `fixtures/` declares `style.color`, and `worked-example.json`
  declares none either**, so the retrofit moves nothing in the corpus today. Re-measure rather than
  trusting this line; the hazard it guards is a round-trip or golden file that declares a `version`
  and would now serialize a raised one, not the rendered PDF digests (the version string is not
  drawn).

## Auto Run Result

### Dispatch 1 — plan only (2026-08-30)

Status: ready-for-dev
Blocking condition: none

Dispatched with `Halt after planning.` after the previous dispatch halted `intent gap` on the
`lineSpacing` range. That gap is ruled: D-7.2.1 through D-7.2.6 (plus D-R7.9 and D-7.3.1) in
`_bmad-output/implementation-artifacts/epic-7-8-decision-log.md`. All six rulings are applied in the
contract above and none was re-opened.

Planned at baseline `02da139273bd9a4ce34874a64d6cadde826321c5`. No code was written and no commit was
created. Epic 7 context was recompiled during this dispatch because
`planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` is newer than
the cached `epic-7-context.md`, which invalidated the cache.

**Three forks surfaced during investigation that the rulings did not name. None is an intent gap; all
three were resolved by a principle already stated in the intent, and each is recorded in the contract
rather than left open:**

1. `int64MulOverflows` already exists (`internal/geom/scale.go:12`) but is **unexported**, and
   `scale_surface_test.go:68` asserts `internal/geom`'s exported surface is exactly `{ScaleRound}`.
   Exporting it would redden that test; AD-2 and D-1.5.2 select the root-package predicate instead.
2. The single shared validator cannot live in the root package: `internal/template` may not import
   the module root while the root imports it, so the `parseHexColor` precedent (a render-time,
   root-package predicate) cannot be copied for a load-time-validated key. It lives in
   `internal/template` and is exported from there.
3. `numericFieldRegistry` has exactly two kinds and its comment states "There is no third kind"; a
   dimensionless thousandths ratio is honestly neither. The registry's value is that its claims are
   true, so a third kind is added rather than mislabelling the key as points.

**Measured at planning, so the implementer does not have to rediscover it:** no fixture under
`fixtures/` declares `style.color`, and `worked-example.json` declares none, so the D-7.2.1 retrofit
moves nothing in the corpus today.
