---
title: 'Story 12.2: The document declares its locale and its UTC offset'
type: 'feature'
created: '2026-09-05'
status: 'done'
review_loop_iteration: 0
baseline_commit: 'e2ff17a5bbb1ec933e852cf0abe55f8dd9cbae2c'
context: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs. Rewritten at close, to
describe what actually shipped.*

Every folio document declares a locale — English, Thai, Simplified Chinese or Japanese — and a fixed
UTC offset, which together decide how its dates and numbers are written. The engine always honoured
both; nothing in the designer could set either, so a Thai statement meant hand-editing the file. Page
setup now carries both, beside the margins: they show what the document says, send what the author
chose, and judge nothing themselves. No new rendering behaviour was added — the engine already knew
all four locales, Buddhist-era years included.

The larger thing this story found was in the file loader, which accepted time-zone offsets that are not
real times: ninety-nine hours and ninety-nine minutes passed, because only the shape was checked and
never the numbers. Such a document opened happily, then drew no dates at all. Rather than make the new
control stricter than the format, we repaired the format's own check, so the panel and the loader now
ask one question where two quietly disagreed. **No version number was increased, and that was
deliberate** — the format document now says why, and every test document still loads.

Three things look wrong on purpose. The loader refuses a bare Z where the expression evaluator accepts
it; they serve different things, and the reason is recorded at both. One long-standing failure in an
unrelated text corpus stays red. The heavier cross-platform and browser suites were not re-run here;
this project runs those once, at the end of the epic. Four smaller issues became follow-ups.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A document's `locale` and `utcOffset` have no writer anywhere in the product. The loader
requires both, the renderer consumes both, `formatDate`/`formatNumber` route every date and number
through `lookupLocale(fc.Locale)`, and no command in either door sets either one — so in a product
whose reason to exist is Thai and CJK statements, every author who wants Thai dates hand-edits the
file the designer just saved. The values are not even projected: `CanvasProjection`'s 21 keys carry
neither field, so the panel has nothing to show even if it had a control.

**Approach:** Widen `CanvasProjection` with `locale` and `utcOffset` — Go struct, the recorded wire
key list, and TypeScript's `hasOnly` mirror in one commit — and add two arms to
`ApplyComponentCommand`, `setDocumentLocale` and `setDocumentUTCOffset`, each writing one field and
each validating through the same predicate the loader uses. Two rows in the PAGE SETUP panel read the
projected values and send those commands on Apply. The panel enforces no bound of its own.

## Boundaries & Constraints

**Always:**
- **One spelling per side of AD-12's closed set.** Go already has exactly one (`LocaleEN`/`LocaleTH`/
  `LocaleZhHans`/`LocaleJA` and `LocaleTags`, built from those constants). TypeScript gets exactly one
  (`LOCALE_TAGS` in `engine-protocol.ts`), tied to Go's by a new mirror in `engine-bounds-mirror.test.ts`
  on the shipped `bandsCappingVertically` idiom. The guard, the panel's options and the factory's type
  all read that one array; no `<option>`, union or literal spells a tag a second time.
- **The command and the loader validate through ONE predicate each.** Export `template.IsLocale` and
  `template.IsUTCOffset` over the existing unexported `closedLocales` and `utcOffsetPattern` — the shape
  and the doc-comment reasoning `IsStyleAlign`/`IsTableStyleAlign` already ship for exactly this reason —
  and route `parse.go`'s own two checks through them, so there is one implementation, not two callers of
  two copies. The refusal messages derive their token list from `template.LocaleTags` and the `±HH:MM`
  phrase from one exported constant.
- **The projection field, the Go wire-key record and TypeScript's `hasOnly` entry move in ONE commit**
  (D-7.4.5). Neither new field carries `omitempty`.
- **Two arms, one field each.** A command names exactly what it changes (Story 15.2a). Refusals are
  `*ComponentCommandError` from `componentFailure` with an empty ElementID and DataPath `locale` or
  `utcOffset`, and leave the serialized document byte-identical.
- **THERE IS NO BACKSTOP HERE, AND THAT IS THE ASYMMETRY WITH 12.1.** `Canvas(t)` reads neither `Locale`
  nor `UTCOffset` — all seven of its refusal sites are geometry, component or font-chain checks — so the
  trailing `Canvas(t)` call that caught a content-window violation on its own in 12.1 (recorded in that
  story's red-proof 3) catches nothing at all here. **Each arm's own validation is the only refusal before
  the wasm-layer reparse.** Weigh it accordingly: a validation gap in these two arms is not a locatedness
  regression, it is an unrefused write.
- **The panel computes no bound and validates no value.** It sends; the engine refuses; the existing
  `role="alert"` path renders the engine's own sentence through `componentDiagnostic`, never
  `pageSetupDiagnostic`. Consistency with typing is the property asserted (D-12.B, Story 17.4 item 9).
- **Measurement discipline:** every count marked CLOSED or SAMPLED with its population; `/usr/bin/grep -a`
  by absolute path or `git grep --untracked`; exit status captured on the command's own line; red-proof
  by mutating the SUBJECT and echoing the mutated line back.

**Ask First:**
- **RULING B IS OUTSTANDING WITH THE ENGINEERING LEAD.** Whether `utcOffsetPattern` is tightened so the
  loader, the command and `internal/expr`'s `parseUTCOffsetMinutes` admit the same set (see Design Notes —
  `+99:99` loads and refuses to render) is a format-compatibility question and is not the builder's to
  settle. This spec's working default is the smallest scope: **the command reuses the loader's predicate
  exactly and nothing is tightened.** Three items below are marked `[HELD-B]` and must not be written until
  the ruling lands; everything else proceeds now. A ruling that tightens the pattern amends this frozen
  block, and only the human may make that amendment.
- Any change to `ApplyPageSetupCommand` or its `len(raw) != 7` arity.
- Minting a new `internal/diag` Code.
- Any fixture byte, golden sha256 or matrix hash moving.

**Never:**
- Never widen the `pageSetup` command with these fields, and never touch `page-setup-command.ts` or its
  ordered 7-key assertion.
- Never add a fifth locale, and never re-spell the four tags or the `±HH:MM` phrase.
- Never validate a locale or an offset in the browser, and never clamp or normalise a typed value.
- Never touch `folio-designer/e2e/browser-native-roundtrip.spec.ts`.
- Never `git add`, `commit`, `stash`, `checkout`, `reset`, `revert`, `restore`, branch, or push. The
  human commits. Leave every change unstaged in the working tree.
- Never delete or rename a designer `.tsx`, `.css` or `e2e/*.spec.ts` file — three population floors sit
  exactly on 8, 3 and 15.
- Never insert anything between `placementPoint`'s closing brace and `function pageStyle` in `App.tsx`.
- Never add an `@media` query, a hex/`rgb()`/`hsl()` literal, a bare `border-radius`, or a second
  `var(--type-display)` / `var(--type-numeric-lg)` to `App.css`.
- Never use `float64` in Go code or Go tests (`TestNoFloat64UnderModule` scans `_test.go` too).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Panel shows the engine's values | document declares `locale: "th"`, `utcOffset: "+07:00"` | the locale control reads `th` and the offset field reads `+07:00`, both seeded from the projection and from nothing else | N/A |
| Locale committed | `{"kind":"setDocumentLocale","version":1,"locale":"th"}` | `doc.Locale` becomes `th`; **`doc.UTCOffset` is unchanged**; a fresh projection is returned; the next render formats every `formatDate`/`formatNumber` under `th`, including the Buddhist-era year | N/A |
| Locale outside AD-12's set | `"locale":"fr"` | refused; serialized document byte-identical | `componentFailure("", "locale", …)`, the legal tags derived from `template.LocaleTags` |
| Locale malformed | `"locale":7` · key absent · `"locale":""` | refused; unchanged | located on `locale` |
| Offset committed | `{"kind":"setDocumentUTCOffset","version":1,"utcOffset":"+07:00"}` | `doc.UTCOffset` becomes `+07:00`; **`doc.Locale` is unchanged**; fresh projection | N/A |
| Offset not `±HH:MM` | `"+7:00"` · `"Z"` · `"07:00"` · `"+0700"` · `""` | refused; serialized document byte-identical | located on `utcOffset`, message carrying the one `±HH:MM` authority |
| Door gates | 4 keys instead of 3 · `locale` declared twice | refused before the handler runs | `componentFields` plain error; `refuseDuplicateCommandKeys` → `componentFailure("", "command", …)` |
| Value re-sent unchanged | current locale sent again | the door accepts and writes; canonical bytes are equal, so `wasm.Engine.Apply` short-circuits — no revision bump, no undo entry, no dirty flag | N/A |
| Row untouched at Apply | author edits only a margin | no `setDocumentLocale` and no `setDocumentUTCOffset` is sent at all | N/A |
| Projection widened without its mirror | Go emits `locale`; `hasOnly` does not list it | `isCanvas` false → `parseInbound` undefined → `EngineClient.#fail('PROTOCOL_INVALID')` terminates the worker; at startup `main.tsx`'s bare `catch` discards the error and the LoadScreen says only *"Local engine/template could not start"* | the red-proof; asserted, not narrated |
| Document never edited | open, then serialize | bytes identical; both keys still present and unmoved | N/A |

</frozen-after-approval>

## Code Map

Measured at `e2ff17a` (HEAD; the dispatch's `bcb6ebb` is `HEAD~1` — the intervening commit touches only
`_bmad-output/`). Counts marked CLOSED are fully enumerated. **Several line numbers carried in from the
survey were stale by 20–45 lines; the numbers below were re-derived at this tree.**

**Go — the model and its two authorities**
- `folio-go/internal/template/model.go:19,25-29` -- `Document.Locale string` and `Document.UTCOffset string`,
  top-level, **plain strings, not `Presence[T]`**, both **required** at load.
- `folio-go/internal/template/locale.go:21-46` -- the four constants, `LocaleTags` (exported, ordered,
  declared literal) and `closedLocales` (unexported, built from the constants). The file's own header
  states each tag is spelled EXACTLY ONCE in the module.
- `folio-go/internal/template/closedsets.go:177` -- `utcOffsetPattern = regexp.MustCompile("^[+-][0-9]{2}:[0-9]{2}$")`,
  **unexported**. `closedsets.go:138-153` -- `IsStyleAlign`/`IsTableStyleAlign`, the shipped precedent for
  exporting a predicate over an unexported closed set *for the command path*, with the doc comment that
  says so. `closedsets.go:160` -- `closedSetMessage(tokens)`, derived-never-hand-written.
- `folio-go/internal/template/parse.go:78-106` -- the two load-path checks. **`"must match ±HH:MM"` (`:101`)
  and `"not one of the closed set en, th, zh-Hans, ja (AD-12)"` (`:86`) are each spelled exactly once in the
  repo** (`git grep --untracked`, CLOSED) — so routing them through a shared authority moves no other test.
- `folio-go/internal/template/locale_test.go:13,34` -- `TestClosedLocalesMatchesLocaleTags` and
  `TestLocaleTagsExactOrder`; the shape a new predicate tie must copy, including its presence precondition.
- `folio-go/internal/template/serialize.go:116-126` -- `writeDocument`'s `[]kv` literal writes `locale` and
  `utcOffset` **unconditionally**. Canonical output is alphabetical, not declaration order.
  `internal/template/drift_test.go:73` pins the slice-literal shape.

**Go — the consumers (AC2 is preservation, not construction)**
- `folio-go/render.go:718` and `folio-go/render.go:2059` -- the **only two** non-test
  `expr.NewFormatContext(doc.Locale, doc.UTCOffset)` sites. CLOSED (13 hits total, 11 in `_test.go`).
- `folio-go/internal/expr/locale.go:127-136` -- `lookupLocale`; an unknown tag is a **hard error, never a
  fallback**, and its message is already derived from `template.LocaleTags`. `locale.go:101-107` --
  `LocaleTH` is the only row with `eraOffset: 543`; applied at `formatdate.go:330`.
- `folio-go/internal/expr/formatdate.go:298-301` and `numberformat.go:140-143` -- the two `lookupLocale`
  call sites. `formatdate.go:27-44` -- `parseUTCOffsetMinutes`, which **accepts `"Z"` and rejects `hh>23`
  or `mm>59`** — a set that is neither a superset nor a subset of the loader's regexp. See Design Notes.
- `folio-go/formatlocale_test.go:1` -- `package folio_test`, **no build tag** (verified: `go:build` count 0).
  `:48` `formatLocaleTemplateJSON(locale, utcOffset, chain, pattern)`; `:453`
  `TestFormatLocaleGoldensAllFourLocales` renders through the real `ParseTemplate`→`Render` path and decodes
  the drawn text back out of the PDF's `/ToUnicode` CMap, with a non-vacuity gate that en/th/zh-Hans differ.
  **Reuse this helper for AC2** — same package, so a sibling `_test.go` can call it.

**Go — the door**
- `folio-go/component_commands.go:202-280` -- `ApplyComponentCommand`. Pre-switch gates in order:
  `refuseDuplicateCommandKeys` (`:206`), the trailing-bytes decode (`:209-218`), the raw-bytes
  `version == 1` compare (`:219`), the `kind` string (`:222`). **25 `case` arms, CLOSED**, `setBandHeight`
  the 25th at `:275`. `default` at `:277` returns a **plain** error → unlocated `ENGINE_REJECTED`.
- `folio-go/component_commands.go:2141-2147` -- `bandHeightPath`; `:2367-2372` -- `fontChainPath`. Two bespoke
  builders, no generic document-level helper. Shared primitive: `truncateAtRuneBoundary` (`:2377`) against
  `maxComponentDataPathBytes = 256` (`:2350`).
- `folio-go/component_commands.go:2184-2315` -- **`setBandHeight`, the precedent to copy in full**:
  `componentFields(raw, 4)`; `commandString(raw, "band")` (`:1490`, rejects missing/non-string/empty);
  seven `componentFailure` sites, six with `""` as ElementID; write, `Canvas(t)`, and **restore the previous
  value on projection failure** (`:2305-2314`).
- `folio-go/component_commands.go:24-33` -- `ComponentCommandError{error; ElementID; DataPath; Message}` and
  `componentFailure(id, path, message)`. `:1483-1488` -- `componentFields`, an exact `len(raw) != want`.
  `:91-113` -- `refuseDuplicateCommandKeys`, routing to `componentFailure("", "command", …)`.
- `folio-go/wasm/cmd/engine/main.go:236-244` -- `errors.As(*ComponentCommandError)` → `COMPONENT_INVALID`
  with Message/ElementID/DataPath bounded at **512 / 128 / 256**. `:256-265` -- the page-setup fallback,
  which needs prefix `folio: page.` (dot). Nothing here needs changing.
- `folio-go/wasm/engine.go:227-232` -- the dispatch is a **binary if/else on the literal `"pageSetup"`**;
  every other kind reaches `ApplyComponentCommand`. **A new arm needs no edit here.** `:239-246` -- the
  canonical-bytes short-circuit: equal bytes means no revision, no dirty flag, no undo entry, and the door's
  projection is discarded. `:247-256` -- past it, the candidate is re-parsed, so the loader is a second net.

**Go — the projection**
- `folio-go/page_setup.go:339-518` -- `CanvasProjection`, **21 json-tagged fields, CLOSED, none `omitempty`,
  and neither `Locale` nor `UTCOffset` among them** (confirmed; negative control: `utcOffset` = 0 hits in
  `page_setup.go` while the same pattern hits `model.go:28-29`).
- `folio-go/page_setup.go:764-817` -- `Canvas`; the whole projection is one composite literal at **`:816`**.
  Its seven refusal sites are geometry/component/font-chain only — **`Canvas` never reads `Locale` or
  `UTCOffset`, so unlike 12.1 there is no trailing backstop; the arm's own validation is the only refusal.**
  `CanvasWithTextPaint` (`:819`) calls `Canvas` first and inherits new fields for free.
- `folio-go/canvas_projection_wire_test.go:47-69` -- `canvasProjectionWireKeys`, 21 sorted literals.
  `:175-189` `TestCanvasProjectionWireKeysAreTheRecordedSet` compares against `json.Marshal` **bytes**, for
  both a projected and a **zero** `CanvasProjection` — that second assertion is what forbids `omitempty`.
  `:346` `canvasGuardKeyList` regexp; `:373-402`
  `TestCanvasProjectionWireKeysAreTheOnesTheDesignerAccepts` reads `engine-protocol.ts` from disk and
  DeepEquals the extracted `hasOnly` list. **Sorted insertion: `locale` after `height`; `utcOffset` after
  `preset`.** This test runs in the ordinary Go gate, not in the designer suite.

**Designer — the protocol**
- `folio-designer/src/engine-protocol.ts:150-200` -- the `CanvasProjection` type, 21 members.
  **`:278` `hasOnly` is a SUBSET check** (`Object.keys(value).every(k => keys.includes(k))`) — the file's own
  comment at `:322-325` says so. An unlisted extra key is rejected; a **missing** key passes `hasOnly` and is
  caught only by a separate typed clause. `:279` `hasExactKeys` is the equality variant.
- `folio-designer/src/engine-protocol.ts:295` -- `isCanvas`'s `hasOnly(value, [...21 keys...])`, on the first
  line of the body, **immediately followed on the same line by `['A4','Letter','custom'].includes(value.preset)`
  and the `orientation` pair** — the shipped precedent for validating a closed set inside the guard.
  `:296-297` the `integer()` sweep; `:336` the band-level `hasOnly`; **`:365`** the
  `BANDS_CAPPING_VERTICALLY` vertical clause. `:92` `BANDS_CAPPING_VERTICALLY`, `:111-112` `CappingBand` /
  `CAPPING_BANDS`.
- `folio-designer/src/engine-client.ts:86-87, 117-124` -- `parseInbound` undefined → `#fail('PROTOCOL_INVALID')`
  → `#detach()`, `worker.terminate()`, state `failed` **permanently**; no re-spawn exists.
  `main.tsx:32` catches and **discards the error object**, so the startup path shows only LoadScreen's
  *"Local engine/template could not start"* with a reload button. Mid-session the canvas **freezes on the last
  good projection** and a generic `role="alert"` appears — "blanks the canvas" is the startup case, not both.
- `folio-designer/src/engine-protocol.test.ts:9` -- a complete valid 21-key `canvas` fixture; `:156`
  `expect(projection({ ...canvas, extraProjectionKey: 1 })).toBeUndefined()` — the existing extra-key
  precedent. 59 `parseInbound` occurrences. **This is the only file that drives the real guard.**
- `folio-designer/src/App.test.tsx` -- `isCanvas` appears **once, inside a comment** (`:1413`); `parseInbound`
  **zero**; positive control `engine` **313**. The engine is a hand-rolled fake (`:154`) and the real
  `EngineClient` is imported type-only (`:9`). **A missing allowlist entry is invisible to all of App.test.tsx.**
- Projection literals that `tsc -b` will force to carry the two new members (found by
  `contentWindowCountIsExact`, CLOSED, binaries excluded): `engine-protocol.test.ts` (4),
  `App.test.tsx:152` (3), `sheet-stack.test.ts:15` (3), `App.font-store.test.tsx:40` (1),
  `DataPanel.test.tsx:23` (1), `sheet-stack.ts` (1, production), `engine-protocol.ts` (4).

**Designer — the command factories**
- `folio-designer/src/command-json.ts:115-116` -- `commandBytes(kind, fields)` emits `kind`, `version: 1`,
  then the fields in order. `jsonString` `:76`, `jsonNumber` `:88` (non-matching → `'null'`).
- `folio-designer/src/band-height-command.ts:37-39` -- the whole factory, 3 lines under a 30-line ruling
  header. **The template to copy.** `band-height-command.test.ts` -- 4 tests: byte-exact wire + `Object.keys`
  order, literal passthrough, emptied-draft `null`, and a **splice test** feeding
  `'80,"band":"pageFooter"'` and requiring one `"band"` key.
- `folio-designer/src/command-json-soleness.test.ts:88` -- `expect(commandJsonBuilders(sources())).toEqual(['command-json.ts'])`.
  **The named-factory list appears TWICE — `:95` and `:113` — as two independent 7-element literals.**
  A new factory must be added to **both**, or `:114`'s `\bString\(` check never looks at it.
- `folio-designer/src/engine-bounds-mirror.test.ts:43-51` -- `goSources` (three Go files);
  `:97` a `Record<GoSource, string>` literal that must gain any new key; `:110` `expect(pairs).toHaveLength(8)`
  and `:114` a set over **`pairs`' sources** — both untouched by a non-numeric tie. `:167-179`
  `goBandsCappingVertically` / `tsBandsCappingVertically`, which **resolve Go's named constants before
  comparing** — the exact idiom the locale tie must copy. `:237-267` the "no consumer holds its own copy"
  assertions, including `expect(app).not.toMatch(/\['pageHeader', 'pageFooter'\]/)`.

**Designer — the panel**
- `folio-designer/src/App.tsx:1682-1683` -- `PageSetup`, the entire component on one line. Row order:
  section label · honest-note · `Preset` `<select aria-label="Page preset">` · `Orientation`
  `<select aria-label="Page orientation">` · conditional `Width/Height (pt)` · four margins ·
  12.1's two conditional band-height `Field`s · `Apply page setup` · honest-note. Rendered from **`:1554`**
  when mode is edit and nothing is selected.
- `folio-designer/src/App.tsx:2973` -- `type Draft`, **every member a `string`**, with 12.1's comment
  explaining why the two band keys are named for the bands (no key-to-band map to rotate).
  `:2981` `Field` -- `<label>{label}<input aria-label={label} inputMode="decimal" …/></label>`, **no
  className** (so `property-prose-height.test.ts`'s counts cannot see it) and **`inputMode` hardcoded**.
  `:2983` `points(value: number)`; `:2984` `draftFor(canvas?)`; `:1304` `updateDraft`; `:139`/`:1303` the
  seeding sites, where `keepNewerDraft` gates the reseed and `setPreset`/`setOrientation` do **not**.
- `folio-designer/src/App.tsx:780-823` -- `applyPageSetup`. The `for (const band of CAPPING_BANDS)` loop at
  `:787`, the three-way skip at `:794` (`typed === points(projected)`), the bare `return` in the catch at
  `:799` that stops the whole sequence, and `pageSetupCommand(…, draft)` at `:815`. **`page-setup-command.ts:13`
  types its last parameter as a 4-key record and reads only those four, so extra `Draft` keys never reach
  the wire** — already true of `pageHeader`/`pageFooter`.
- `folio-designer/src/App.tsx:3000` `pageSetupDiagnostic` (**discards the engine's message unless
  `PAGE_SETUP_INVALID`**) vs `:3002` `componentDiagnostic` (always keeps it, prefixed by elementId ?? dataPath).
- `folio-designer/src/App.test.tsx:1366-1539` -- 12.1's eight band-height tests. **The shape to copy**:
  `getByRole('textbox', { name: '<exact label>' })`, `toHaveValue` for the projected value, and the wire
  pinned as full decoded strings in order (`expect(sent[0]).toBe('{"kind":"setBandHeight",…}')`) with the
  comment explaining that *"a band-height command was sent"* passes just as happily with the rows crossed.
  The refusal test (`:1428-1449`) asserts the engine's own sentence **and** `not.toHaveTextContent` of the
  fixed page-setup sentence, **and** `toHaveBeenCalledOnce()` for the stop, **and** that the field still
  shows the refused value.
- `folio-designer/src/App.css:353-361` -- `label`, `input, select`, and the focus ring are **bare-element**
  rules. **A `<select>` row and a text `<input>` row need no new CSS and no class.**

**Source-text contracts — re-derived at this tree, not inherited**
- `property-prose-height.test.ts:143-148` counts `<input[^>]*property-value-prose` and the `<textarea>`
  twin at exactly 1 each; `:118-125` forbids `position:` inside `.property-field`. **A class-free panel row
  fires neither.**
- `canvas-authority-contract.test.ts` -- `withoutComments` at `:131-153`, quote-aware, applied at **both**
  scan sites (`:155` and `:184`) **and** at `:484`. **A comment cannot trip the prohibited-token list**
  (D-12.A.1 is wrong; D-12.A.2 corrects it, and this tree agrees). Two real raw-text hazards:
  `:238-241` App.css may contain **exactly one** `@media`, and `:535-546` `withoutApprovedLocalPointerInput`
  runs on raw App.tsx and **hard-asserts** that `placementPoint`'s closing `}` is immediately followed by
  `function pageStyle` (verified adjacent at this tree). Population floors `:431-435`: production ≥58 (**60**),
  tests ≥51 (**56**), e2e ≥15 (**15, on the floor**), `.tsx` ≥8 (**8, on the floor**), `.css` ≥3 (**3, on the floor**).
- `design-contract.test.ts` -- **reads App.css and never App.tsx** (grep rc=1; positive control `App.css` = 10
  lines). `:87` no hex/`rgb(`/`hsl(` anywhere in App.css; `:88` `border-radius` must be `var(--radius…)`;
  `:147-148` exactly one `var(--type-display)` and one `var(--type-numeric-lg)`.
- `engine-ownership-contract.test.ts:11,25-26,65` -- any TS type or object literal carrying **≥2** of
  `{version, page, bands, elements, assets}` is a schema mirror and fails. `{locale, utcOffset}` carries
  **zero**; but note `CanvasProjection` already has `bands`, so never add a `page` or `version` sibling to it.
- The complete set of raw App.tsx/App.css readers, CLOSED: `property-prose-height.test.ts`,
  `canvas-authority-contract.test.ts`, `canvas-font-stack.test.ts` (**SAMPLED** — reads App.tsx at `:616`,
  `:751` and App.css at `:44`; check before touching font-adjacent CSS), `engine-bounds-mirror.test.ts`,
  `engine-ownership-contract.test.ts`, and `App.test.tsx:4030`.

**What will NOT move, measured**
- Every `fixtures/*/expected.json` sha256, every golden PDF and every matrix hash. `Canvas()` is called only
  from command handlers — never from the render path — and no `.folio` byte changes. The corpus (28 `.folio`
  files, CLOSED) is `locale` 24×`en` / 4×`th` and `utcOffset` 21×`+00:00` / 7×`+07:00`; `zh-Hans` and `ja`
  appear in **no** fixture.
- `engine-bounds-mirror.test.ts`'s eight numeric pairs; `page-setup-command.ts` and its ordered 7-key test;
  `folio-go/page_setup.go:1961`'s `len(raw) != 7`; `internal/diag`'s 17 Codes; the starter template's bytes.
- No Go census of command kinds exists (measured with controls) — a 26th and 27th arm join no list.
- `lint/` is untouched: a new exported identifier in `internal/template` trips nothing, and no new root-level
  `folio-go/*.go` file is created (the arms go beside `setBandHeight`, so `forbiddenimports`' auto-fenced
  population is unchanged).

**Read-only / do not touch**
- `folio-designer/e2e/browser-native-roundtrip.spec.ts` — hung, pre-existing (D-000.19), out of scope. It
  carries no key list; nothing here requires editing it.
- `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — the mandated permanent red.
- `e2e/` is **not** typechecked by `npx tsc -b` (`tsconfig.e2e.json` is unreferenced); CI compiles it in a
  separate step and `playwright test` runs in no workflow. Put every real assertion in vitest, not e2e.

## Tasks & Acceptance

**Execution:**

**Sequencing under ruling B — HISTORICAL, SUPERSEDED; see the paragraph that follows it.** While the
ruling was outstanding, three items below carried `[HELD-B]` and were not to be written until the
engineering lead's ruling on `utcOffsetPattern` lands. Two further pieces *depend* on those three and are
therefore held with them, not by a marker of their own: the `setDocumentUTCOffset` arm (it calls
`template.IsUTCOffset`, so it cannot compile before the predicate exists) and that arm's rows in
`document_settings_command_test.go`. **Do the whole locale half, the projection widening and every
TypeScript task first** — all of them are offset-agnostic apart from carrying the field through as an
opaque string — and land the offset half last, in one pass, once the ruling is in hand. Do not stub
`IsUTCOffset` to unblock yourself: a stub written on the wrong semantics is exactly what the hold exists
to prevent. If you reach the end of the unheld work and the ruling has not arrived, stop and report that,
rather than guessing.

**RULING B LANDED (D-12.C). THE PARAGRAPH ABOVE IS HISTORY, NOT INSTRUCTION.** The hold is lifted, the
three `[HELD-B]` markers below no longer bind, and the sequence is free — implement the whole spec,
offset half included. **The frozen block's Ask First clause is likewise superseded**: it records the
working default as *"the command reuses the loader's predicate exactly and nothing is tightened"*, and
the ruling reverses exactly that. The frozen text is deliberately left unamended so the record shows what
was believed when the block was locked; do not follow it. Read D-12.C for the ruling itself.
- [x] `folio-go/internal/template/locale.go` -- export `IsLocale(s string) bool` over `closedLocales`, with a
  doc comment on `IsStyleAlign`'s model naming the command path as the reason. -- One predicate, two doors.
- [x] `folio-go/internal/template/closedsets.go` -- **RULED (D-12.C): repair `utcOffsetPattern` so it
  enforces `±HH:MM`'s actual range and refuses `+99:99` and `+24:00`** — the regexp was a loose rendering of
  a spec (`folio-format.md:49`, *"Fixed offset, `±HH:MM`"*) that was never loose. Then export
  `IsUTCOffset(s string) bool` over the repaired pattern, and one exported constant carrying the `±HH:MM`
  phrase, beside it. -- The
  offset's syntax and the sentence that describes it get one authority each. **The export and the constant
  are needed under every option B can rule; what ruling B decides is whether `utcOffsetPattern` ITSELF
  changes underneath them. Write this only once B has landed, so the predicate is written once.**
- [x] `folio-go/internal/template/parse.go` -- route `:85` and `:100` through the two new predicates and the
  new constant. **The `locale` message bytes must not change. The `utcOffset` message bytes need not change
  either — D-12.C repairs the PATTERN, not the sentence — but the set the sentence describes now genuinely
  matches it, which it did not before.** **The emitted message bytes must not change** (both strings are spelled exactly once in the
  repo, so nothing else pins them). -- Otherwise the command is a second caller of a second copy.
- [x] `folio-go/internal/template/locale_test.go` and `closedsets_test.go` -- tie each predicate to the set it
  reads, on `TestClosedLocalesMatchesLocaleTags`' shape including its presence precondition: `IsLocale` admits
  every member of `LocaleTags` and refuses a non-member and the empty string; `IsUTCOffset` agrees with the
  loader on a table that includes `+07:00`, `-00:00`, `+00:00` (admit) and `Z`, `+7:00`, `+0700`, `07:00`, `""`
  (refuse). **RULED (D-12.C): `+99:99` and `+24:00` go in the REFUSE column**, for `IsUTCOffset` and for the
  loader alike — they now agree by construction, because both call the one repaired predicate. -- A predicate nothing ties can drift from the set it claims to read.
- [x] `folio-go/page_setup.go` -- add `Locale` and `UTCOffset` to `CanvasProjection` (json tags `locale`,
  `utcOffset`, **no `omitempty`**) and populate both from `t.doc.Locale`/`t.doc.UTCOffset` in `Canvas`'s
  composite literal at `:816`. -- AC1's projection half; the zero-value wire test forbids `omitempty`.
- [x] `folio-go/canvas_projection_wire_test.go` -- insert `"locale"` and `"utcOffset"` into
  `canvasProjectionWireKeys` in **sorted position**. -- The three-way record; it reddens if only one side moves.
- [x] `folio-go/component_commands.go` -- add `setDocumentLocale` and `setDocumentUTCOffset` arms (26th and
  27th) to the switch, and their handlers beside `setBandHeight`. Each: `componentFields(raw, 3)`;
  `commandString` for its one field; validate with the new predicate; refuse with
  `componentFailure("", "locale"|"utcOffset", …)` where the locale message derives its token list from
  `template.LocaleTags` (as `updateComponentProperties` already does with `StyleAlignTokens` at `:1318`) and
  the offset message carries the exported `±HH:MM` constant; **write exactly one field**; then `Canvas(t)`,
  restoring the previous value if the projection fails. Add the two DataPath constants beside
  `componentCommandPath`. -- The writer. Note `Canvas` gives no backstop here, unlike 12.1.
- [x] `folio-go/document_settings_command_test.go` -- new, on `band_height_command_test.go`'s shape: a
  `documentSettingsRefusal` helper that asserts non-nil error **and** byte-identical serialization **and**
  `errors.As(*ComponentCommandError)` in one call; a row per Matrix line; located-field compares on
  `ElementID`/`DataPath`; door-gate rows kept separate because `componentFields` and
  `refuseDuplicateCommandKeys` return plain errors. **Every accepting test must assert the OTHER field is
  unchanged in the serialized bytes** — read the raw JSON literals, never decode into `any`. -- The rotation
  that shipped green twice in this run is `setDocumentLocale` writing `UTCOffset`; only a both-fields
  assertion can see it.
- [x] `folio-go/document_settings_command_test.go` -- add the red-proofs as a numbered comment block, each
  naming the production line mutated and echoing it: delete the locale predicate call; swap the two arms'
  write targets; delete the offset predicate call; make each arm write nothing. Record which tests went red
  and which stayed green. -- D-000.14: mutate the subject, and prove the arm is REACHED, not just ordered.
- [x] `folio-go/formatlocale_command_test.go` -- new, `package folio_test`, reusing
  `formatLocaleTemplateJSON` and the `/ToUnicode` decode from `formatlocale_test.go`: build an `en` document
  carrying a `formatDate` and a `formatNumber`, apply `setDocumentLocale` with `th`, serialize, re-parse,
  render, and assert the drawn text is the Thai Buddhist-era form — and that the same document rendered
  before the command was the `en` form. -- AC2's only test that can fail; it also catches the field rotation
  end to end.
- [x] `folio-go/internal/template/parse_test.go` (or the closest existing load-path test) -- **the red-proof
  is the REACHABILITY, not the regexp.** Assert that a document declaring `"utcOffset": "+99:99"` is now
  **refused at load with a located message**, where before this story it loaded and every `formatDate` then
  failed at render. A test asserting only that the pattern rejects the string proves the regexp changed, not
  that the defect closed. -- D-12.C's carried condition.
- [x] `folio-go/internal/expr/formatdate.go` -- update `parseUTCOffsetMinutes`' doc comment. It claims the
  document field is *"already syntax-checked at load"*; that is **false before this story and true only
  because of it**, and it is the sentence that made the gap invisible. **Do not otherwise touch the function
  — `Z` stays, and the document-path call at `:313` stays** (no `Canvas` backstop; defence-in-depth inside
  one package is cheap). -- D-000.10 applies to accurate-*sounding* comments, not only to stale rulings.
- [x] `folio-go/internal/template/` -- add the anti-divergence guard: assert loader-admits <=> evaluator-parses
  across a candidate offset set, **with `Z` declared as the one expected asymmetry and its reason stated
  inside the assertion** (RFC 3339 data spelling, required by `formatdate.go:196` and `:237`; not a document
  spelling). A guard that silently filters `Z` out of its candidate set hides the fact D-12.C uncovered. Mind
  the stage-rank wall: `template` cannot import `expr`, so site this where both are reachable —
  `internal/expr` is the only side where both predicates are in scope, because `expr` already imports
  `template` and the reverse is forbidden. **State that reason in the test file's own comment, not only in
  the Delivery Log:** the next person who tries to move this test needs the constraint at the file, where
  they will be standing when they try. -- The tie
  that stops the two doors drifting again.
- [x] `_bmad-output/specs/spec-folio/folio-format.md` -- **narrate the non-increment.** Add one more `> `
  note to the version-rule blockquote cluster that ends at `:93` (continue the same blockquote: a bare `>`
  separator line, then the note), in that cluster's voice. It must say that the `utcOffset` loader was
  tightened to the `±HH:MM` that **`:49` already specifies**, that **no version increment was taken**, and
  why. The three grounds, in order of force: (i) `utcOffset` is **not one of the nine closed sets the rule
  enumerates at `:68-70`** — it is a pattern-constrained string, so the closed-set rule never reached it;
  (ii) D-7.3.1's pre-reader test (`:575`) asks what an OLDER reader does with a NEWER document, and
  **narrowing produces no new documents** — it produces a stricter reader, a direction the rule is silent
  in; (iii) **this is the same reader-strictness principle the document already records at `:585-587`** —
  *"Making the reader stricter about `2.0` documents is not a version trigger … version describes the
  document, never the writer"* — so cross-reference it rather than re-deriving it. Carry the corpus number:
  **28 `.folio` files, 21× `+00:00`, 7× `+07:00`, zero excluded**; the only documents excluded are ones
  whose every `formatDate` already fails at render. Note too that a bump would have restamped every
  fixture and moved its goldens for a reason unrelated to time zones — the same objection `:589-591`
  raises. Cite D-12.C / D-12.C.1. -- A file whose convention is to narrate version events makes an
  unnarrated one a false signal; the next person tightening a validator needs the precedent.
- [x] Search the repo for any existing test pinning the OLD looseness -- if anything asserts that `+99:99` or
  another out-of-range offset loads successfully, correct it **and report it explicitly**, because such a
  test would mean the looseness was once observed and accepted. Use `/usr/bin/grep -a` or
  `git grep --untracked`, state the population, and run a control that fires. -- D-12.C's carried condition;
  a silent correction here would bury the more interesting finding.
- [x] `folio-designer/src/engine-protocol.ts` -- add `locale` and `utcOffset` to the `CanvasProjection` type;
  add both to `isCanvas`'s `hasOnly` array; add typed clauses (`locale` against a new exported `LOCALE_TAGS`,
  in the same style as the shipped `preset`/`orientation` checks; `utcOffset` a non-empty string bounded by
  `MAX_CANVAS_PROPERTY_STRING`). **The typed clauses are load-bearing: `hasOnly` is a subset check and cannot
  see a key Go failed to send.** -- The mirror half of the protocol change, in the same commit.
- [x] `folio-designer/src/engine-protocol.test.ts` -- extend the `:9` fixture; add a case that an unlisted
  extra key still drops the snapshot, and one case per new field for **absence** (`delete canvas.locale`) and
  for an illegal value (`locale: 'fr'`, `utcOffset: ''`). -- The absence cases are the only ones `hasOnly`
  cannot catch, and they are the failure this story could otherwise ship silently.
- [x] `folio-designer/src/document-settings-command.ts` -- new factory, two exports on
  `band-height-command.ts:37-39`'s shape, both through `commandBytes` and `jsonString`, the locale one typed
  by the union derived from `LOCALE_TAGS`. -- The encoder.
- [x] `folio-designer/src/document-settings-command.test.ts` -- new, on `band-height-command.test.ts`'s four-
  test shape including the **splice test** (feed `'th","utcOffset":"+09:00'` and require one key). -- Pins the
  wire and proves the escaping.
- [x] `folio-designer/src/command-json-soleness.test.ts` -- add `document-settings-command.ts` to the named
  factory literal at **`:95` AND at `:113`**. and leave a one-line comment at each
  site saying the list is spelled twice and both must move together. -- Two independent literals; adding to
  one leaves the file scanned-but-never-named and the `\bString(` check blind to it. Without the comment the
  next person adds to one again.
- [x] `folio-designer/src/engine-bounds-mirror.test.ts` -- add `folio-go/internal/template/locale.go` to
  `goSources` and to the `Record<GoSource, string>` literal at `:97`, and add a `locale tag mirror` describe
  block on the `bandsCappingVertically` idiom: resolve Go's named constants, read `LOCALE_TAGS` out of
  `engine-protocol.ts`, assert non-vacuity on both sides first, assert equality, and assert that `App.tsx`
  and the factory hold **no copy** of their own (`expect(app).not.toMatch(/'zh-Hans'/)`). Do not add a
  numeric pair — `:110`'s `toHaveLength(8)` and `:114`'s source set are derived from `pairs` and must not move.
  -- Kills the fourth spelling of AD-12's set with the repo's own mechanism.
- [x] `folio-designer/src/App.tsx` -- extend `Draft` with `locale` and `utcOffset` (seeded in `draftFor` from
  the projection, `''` in the no-canvas branch); add a `Locale` `<select aria-label="Document locale">` row
  whose options are **mapped from the imported `LOCALE_TAGS`**, with each option's value AND its visible text
  both the tag itself (no display-name map — a fifth artifact keyed by the tag set would need its own tie, and
  the `.folio` file says `zh-Hans`); and a `<Field label="UTC offset (±HH:MM)">` text row (the format goes IN the label: every other unit-bearing row in this panel carries its unit there, and without it the only route to the syntax is to be refused — which also aborts the rest of the gesture) (give `Field` an optional
  `inputMode` prop rather than duplicating it — the existing rows must keep `decimal`, and this one takes
  none); in `applyPageSetup`, before the band loop,
  send `setDocumentLocale` and then `setDocumentUTCOffset` **only when the draft value differs from the
  projected value**, stopping at the first refusal and surfacing it with `componentDiagnostic`. No clamp, no
  browser-side validation. -- AC1 and AC3. Do not touch the `placementPoint`/`pageStyle` adjacency.
- [x] `folio-designer/src/App.test.tsx`, `sheet-stack.test.ts`, `App.font-store.test.tsx`,
  `DataPanel.test.tsx`, `sheet-stack.ts` -- add the two members to every projection literal (`tsc -b` will
  name them all). In `App.test.tsx`, add tests on 12.1's shape, finding the rows by
  `getByRole('combobox', { name: 'Document locale' })` and `getByRole('textbox', { name: 'UTC offset (±HH:MM)' })`:
  each row shows the projected value; the locale
  select offers exactly the four tags and no more; changing only the locale sends **exactly**
  `{"kind":"setDocumentLocale","version":1,"locale":"th"}` and **no** offset command; changing only the offset
  sends only the offset command; an unchanged row sends nothing; a refusal renders the engine's own located
  sentence **and not** the fixed page-setup sentence, stops the sequence, and leaves the typed value in place.
  -- Assert what would have to change for each to fail: a bare "a command was sent" cannot see the two
  commands crossed.

**Acceptance Criteria:**
- Given a document declaring `locale: "th"` and `utcOffset: "+07:00"`, when the PAGE SETUP panel is shown,
  then both rows read those values from the projection — and searching `folio-designer/src/` finds no default,
  no fallback and no second copy of the tag list outside `LOCALE_TAGS`.
- Given the locale row changed and Apply pressed, when the command is accepted, then the next render formats
  every `formatDate` and `formatNumber` under the new locale, and `utcOffset` is byte-unchanged in the file.
- Given `setDocumentLocale` and `setDocumentUTCOffset` at this commit, when each arm's write is swapped for
  the other's field, then at least one test in the Go suite fails for each swap — recorded in the red-proof
  block with the mutated line echoed.
- Given the designer at this commit, when `folio-designer/src/` is searched for a browser-side locale or
  offset validator, then there is none: the panel proposes values the engine may refuse, exactly as typing
  already does.
- Given a document whose locale and offset are never edited, when it is opened and serialized, then the bytes
  are identical.
- Given `'locale'` removed from `isCanvas`'s `hasOnly` array, when the designer suite and the Go suite run,
  then both `engine-protocol.test.ts` and `TestCanvasProjectionWireKeysAreTheOnesTheDesignerAccepts` fail —
  the omission is visible on both sides of the channel, not only in the one that ships it.

## Spec Change Log

**2026-09-05 — review pass 1. `review_loop_iteration` stays 0: NO loopback.** Three review layers ran
(blind-hunter, edge-case-hunter, verification-gap). **No `intent_gap` and no `bad_spec`** — the frozen
intent held and the non-frozen sections did not mislead the implementer — so the implementation was not
reverted and no section was re-derived. Every finding routed to `patch`, `defer` or `reject`.

Two non-frozen amendments were made as part of that triage, recorded here because the code and the spec
must not drift: (1) the UTC offset row's label became `UTC offset (±HH:MM)`, because the panel's own
convention puts the unit in the label and the row as first specified gave the author no way to learn the
syntax except by being refused; (2) `App.tsx` gained a disabled `Not set` placeholder option for the
empty-locale draft, ruled a required pre-commit fix — without it the browser paints the first tag while
React's value is `''`, so the control asserts `en` for a document that has said nothing. Behaviour was
already safe (Apply disabled, `applyPageSetup` returns early), which is exactly why no command assertion
could see it; a display test was added and red-proved by deleting the placeholder.

**KEEP on any future re-derivation:** the two arms stay separate with fixed DataPaths (no discriminator to
rotate); `IsLocale`/`IsUTCOffset` stay the single predicate each door calls; the projection field, the Go
wire-key record and TypeScript's `hasOnly` entry stay in one commit; `Z` stays admitted by the evaluator
and refused by the loader, named as the one declared asymmetry with its reason inside the assertion; and
the anti-divergence guard stays in `internal/expr`, which is the only side of the stage-rank wall where
both predicates are in scope.

## Design Notes

**Why two arms, not one.** `setBandHeight` carries a `band` discriminator because a band is one field on
three interchangeable structures. `locale` and `utcOffset` are two independent top-level fields with no
shared shape and no shared validation, so a single arm would have to branch its DataPath and could refuse a
good locale because of a bad offset. Two arms give each refusal a fixed DataPath and no discriminator to
rotate. The seven font-chain arms are the shipped precedent for one document-level structure served by
several narrowly-named commands.

**The `±HH:MM` sets do not agree, and RULING D-12.C fixes it here: (c) then (a).** Measured at this tree:

| value | loader (`utcOffsetPattern`) | evaluator (`parseUTCOffsetMinutes`) |
|---|---|---|
| `+07:00`, `-00:00` | admit | admit |
| `+99:99` | **admit** | **refuse** (`hh > 23`) |
| `Z` | **refuse** | **admit** |

So an author can commit `+99:99`, save, reopen — and every `formatDate` in the document then fails to render
with `expr: invalid UTC offset "+99:99"`. Pre-existing, but this story is the first thing that makes it
reachable without hand-editing. **Ruled (D-12.C): repair `utcOffsetPattern` to enforce the range, then have
the command reuse the repaired predicate.** The fork dissolves — one spelling of the document rule, in one
place, command and loader agreeing by construction. The predicate body changes; the arm does not.

**`Z` is NOT a divergence and must not be touched.** `parseUTCOffsetMinutes` serves two populations with one
function: the document's `utcOffset` and offsets embedded in RFC 3339 timestamps arriving in report *data*.
Three call sites — `formatdate.go:196` and `:237` are the data path, `:313` is the document field. `Z` is
RFC 3339's canonical UTC spelling and the data caller requires it. The loader refuses it on ONE ground and
needs no other: `folio-format.md:49` says the field is a *"Fixed offset, `±HH:MM`"*, and `Z` is not that —
excluding it implements the format, exactly as excluding `+99:99` does. **An earlier draft of this spec
justified the refusal by calling a second byte-spelling of `+00:00` a canonicalization hazard. That
argument is WITHDRAWN (D-12.C.4): it proves too much — the loader admits `-00:00`, which is exactly a
second spelling of `+00:00`, both before this story and after it, so the argument condemns the very
pattern D-12.C ruled correct. It never bit because values travel verbatim: the serializer does not
normalise an offset, `-00:00` round-trips as `-00:00`, and byte identity holds without anything choosing
between spellings. A canonicalization hazard requires a canonicalizer, and there is none.** The asymmetry
is still disclosed — the loader admits `-00:00` and refuses `Z` — but the reason is only that `-00:00`
IS `±HH:MM` and `Z` is not. The anti-divergence guard must therefore **name `Z` as the one expected asymmetry with its reason
stated inside the assertion** — a guard that silently filters it out hides the fact this decision uncovered.

**MY OWN CITATION WAS WRONG, AND SO IS THE CORRECTION FILED AGAINST IT.** This spec originally justified
option (a) by claiming *"tightening `utcOffsetPattern` narrows what the format accepts, which
folio-format.md:68's compatibility clause makes an owner's call."* That was citation-without-definition: the
line exists but does not say that. D-12.C then over-corrected, recording that `folio-format.md` has **no**
compatibility clause on a `compatib|backward|will always load|never reject` query returning 0 (rc 1).
Re-measured here: that query is genuinely 0, **but the clause is real and sits at `folio-format.md:67-72`**,
written in MINOR/MAJOR/increment vocabulary (10 hits in that file) that the query never contained — a false
zero from a vocabulary-based absence test. Its actual content: *"A MINOR increment may add new optional keys
only … it may **not** extend a closed set of legal values (element `type`, `locale`, …). Extending a closed
set is a **MAJOR** change, because every existing library validates those sets as load errors."* **It governs
EXTENDING a closed set, not NARROWING a loose regexp**, so it never governed this question — (c) is right
because the clause is inapplicable, not because it is absent. **Consequence for this story: the `Never add a
fifth locale` boundary is positively grounded at `folio-format.md:67-72`, not merely prudent.**

**A charter phrased against the mechanism can be met by a fix that misses the purpose.** Epic 12's charter —
*"every value it makes authorable is one the loader already accepts"* — was what made (a) look correct.
Under (a) alone that charter holds in **letter** while the panel ships a value that destroys the document's
rendering; it only appeared satisfiable because the loader was doing too little. Under (c)+(a) it is true in
**substance**. Worth remembering the next time a charter is quoted to bound a fix.

**The mockup does not draw this panel.** `epics.md:4107-4108` points Story 12.2 at
`ux-designs/ux-folio-2026-08-23/mockups/Main.dc.html` — *"the inspector's PAGE SETUP panel"*. That file
contains **zero** occurrences of "page setup", "locale" or "UTC" (population 383 lines; positive controls
fired: 14 uppercase section labels extracted, `div` on 206 lines), and its inspector draws only the
component-selected state — POSITION, TYPOGRAPHY, BINDING. There is no visual precedent to follow. The shipped
`PageSetup` component is the only authority for this panel's shape, and the two existing `<select>`s in it are
the pattern for the locale control.

**Why the locale set gets a mirror and `preset`/`orientation` did not.** Both existing selects hardcode their
options with no tie to Go (measured: zero test anywhere renders or reads either select; `engine-bounds-mirror.test.ts`
deliberately covers neither). Copying that would make AD-12's four tags a fourth un-tied spelling, in the one
story whose dispatch names second-authority spellings as the defect it keeps catching. The mirror costs one
`goSources` entry and one describe block, and the idiom already exists three lines away.

**What the failure actually looks like, so the red-proof asserts the right thing.** A projection key Go emits
and `hasOnly` does not list makes `isCanvas` false, `parseInbound` undefined, and `#fail('PROTOCOL_INVALID')`
call `worker.terminate()` with no restart. On the **startup** path — the one this story hits, because the
starter loads first — `main.tsx:32`'s bare `catch` discards the error entirely and the user sees only *"Local
engine/template could not start"* and a reload button that fails identically. Mid-session the canvas freezes
on the last good projection and a generic alert appears. Assert the guard's return value, not "the canvas
blanks".

## Verification

Run each command with its exit status captured on the command's **own** line (`cmd; rc=$?` — `$?` is clobbered
by any intervening command, `echo` included), and never `&&`-chain them. A `cd` earlier in a compound command
silently re-roots every later relative path.

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: exit **1** with **exactly two** failures,
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child (measured at baseline `e2ff17a`:
  14 packages ok, `internal/text` FAIL, `P6g` got 7 need >=20). **A third failure is a real regression.**
- `cd folio-designer && npx vitest run` -- expected: exit 0. Baseline **59 files / 819 tests**; report the new
  totals and the diff of test-file names, not just the count.
- `cd folio-designer && npx tsc -b` -- expected: exit 0. **Not `--noEmit`**: `folio-designer/tsconfig.json` is a
  solution file (`"files": []`, two references), so `npx tsc --noEmit` type-checks zero files and exits 0 (DW-207).
- `cd folio-designer && npx oxlint` -- expected: exit 0 with **exactly 4** `only-export-components` warnings
  (baseline: `preview/pdf-viewer.tsx:16,17`, `App.tsx:2988,2995`). No fifth.
- `cd lint && go build ./...` -- expected: exit 0.
- `cd lint && go vet ./...` -- expected: exit 0.
- `cd lint && test -z "$(gofmt -l .)"` -- expected: exit 0.
- `cd lint && go test -count=1 ./...` -- expected: exit 0 (4 packages ok at baseline).
- `cd folio-designer && npm run scan:font-hosts` -- expected: exit 0 and
  `forbidden font host scan: 0 occurrence(s) in N tracked source files under .. (floor 400)`.
  Measured mid-flight: **0 in 625 tracked files, rc 0**.
  **CAVEAT, AND IT IS THIS STORY'S OWN BLIND SPOT.** The scan discovers its population with
  `git ls-files -z` (`folio-designer/scripts/forbidden-font-hosts.mjs:253`) — **tracked files only**; the
  script's own header at `:11` names an untracked file as a known hole. This story adds four NEW files,
  and staging is prohibited, so **they are outside the scanned population and the green says nothing about
  them.** A pass here must not be reported as covering this story's new work.
- **Compensating check, run explicitly and reported as its own measurement** — grep each new file added by
  this story for `fonts.googleapis.com`, `fonts.gstatic.com` and the `DECLARED_ONLY_FONT_HOSTS` entries
  (`forbidden-font-hosts.mjs:41-44` and just below), with `/usr/bin/grep -a` by absolute path. Expected: 0
  in each, with the file list stated and a positive control that fires. Without this, the story's new files
  are unscanned and nothing says so.

No matrix suite and no Playwright run — end-of-epic cadence.

**Manual checks:**
- `git status --porcelain` -- expected: only the files this story names. Nothing staged, nothing committed.
- Confirm `folio-designer/e2e/browser-native-roundtrip.spec.ts` is untouched.

## Suggested Review Order

**Start here — the one authority the whole story turns on**

- Both doors now ask one predicate; the repaired pattern is the story in one line.
  [`closedsets.go:237`](../../folio-go/internal/template/closedsets.go#L237)

- The range repair itself: `+99:99` and `+24:00` stop loading. Read the `Z` rationale above it.
  [`closedsets.go:226`](../../folio-go/internal/template/closedsets.go#L226)

- The locale twin, over the unexported closed map, on `IsStyleAlign`'s shipped precedent.
  [`locale.go:61`](../../folio-go/internal/template/locale.go#L61)

- The loader now calls the predicates rather than re-spelling them; message bytes unchanged.
  [`parse.go:85`](../../folio-go/internal/template/parse.go#L85)

**The write path — two arms, one field each, no discriminator to rotate**

- The 26th and 27th arms. Separate kinds so a bad offset cannot refuse a good locale.
  [`component_commands.go:289`](../../folio-go/component_commands.go#L289)

- Validates, writes one field, re-projects, restores on failure. Note: `Canvas` gives no backstop here.
  [`component_commands.go:2365`](../../folio-go/component_commands.go#L2365)

- The offset arm; its refusal is located on `utcOffset`, ElementID empty.
  [`component_commands.go:2408`](../../folio-go/component_commands.go#L2408)

**The protocol seam — three edits that must never travel apart**

- The projection gains two keys, no `omitempty`; the zero-value wire test forbids it.
  [`page_setup.go:354`](../../folio-go/page_setup.go#L354)

- Populated from the document in `Canvas`'s single composite literal.
  [`page_setup.go:830`](../../folio-go/page_setup.go#L830)

- The recorded key list, sorted. Reds if either side of the channel moves alone.
  [`canvas_projection_wire_test.go:62`](../../folio-go/canvas_projection_wire_test.go#L62)

- The mirror: `hasOnly` is a SUBSET check, so each new key also needs its own typed clause.
  [`engine-protocol.ts:329`](../../folio-designer/src/engine-protocol.ts#L329)

- AD-12's set, spelled once on this side; the panel and the factory both read it.
  [`engine-protocol.ts:118`](../../folio-designer/src/engine-protocol.ts#L118)

**The panel — proposes, never judges**

- The two rows. The disabled `Not set` option stops the control asserting `en` for no document.
  [`App.tsx:1730`](../../folio-designer/src/App.tsx#L1730)

- Sent only when the draft differs from the projection, before the band loop, stopping at the first refusal.
  [`App.tsx:817`](../../folio-designer/src/App.tsx#L817)

- The encoder, through `commandBytes`; the locale parameter is typed by the tag union.
  [`document-settings-command.ts:28`](../../folio-designer/src/document-settings-command.ts#L28)

**Where the two doors are held together**

- Loader-admits vs evaluator-parses, with `Z` declared as the one asymmetry and its reason in the assertion.
  [`offset_divergence_test.go`](../../folio-go/internal/expr/offset_divergence_test.go)

- The shared probe table; the command test reads it from source and re-asks the authority per row.
  [`closedsets_test.go:584`](../../folio-go/internal/template/closedsets_test.go#L584)

- TypeScript's tag list tied to Go's, plus a `src/`-wide census for copies — it caught one.
  [`engine-bounds-mirror.test.ts:480`](../../folio-designer/src/engine-bounds-mirror.test.ts#L480)

**Peripherals**

- A command reaching the renderer, proved by drawn glyphs for both locale and offset.
  [`formatlocale_command_test.go`](../../folio-go/formatlocale_command_test.go)

- Matrix rows, located refusals, byte-identity after every refusal, and the red-proof block.
  [`document_settings_command_test.go`](../../folio-go/document_settings_command_test.go)

- The panel's behaviour, including the three guards this story would otherwise have shipped unfalsifiable.
  [`App.test.tsx:1556`](../../folio-designer/src/App.test.tsx#L1556)

- The non-increment, narrated where the file already narrates version events.
  [`folio-format.md:98`](../../_bmad-output/specs/spec-folio/folio-format.md#L98)

## Delivery Log

### 2026-09-05 — done

Baseline `e2ff17a`. Shipped as one commit, `52d0509` on `main`, 27 files. Closed by the closer; the
frozen intent block was not touched (verified byte-identical by sha256 before and after this entry).

**What shipped, and the one thing that was not the panel.** Two rows in PAGE SETUP, two command arms
(`setDocumentLocale`, `setDocumentUTCOffset`), the projection widened with `locale` and `utcOffset`
across the Go struct, the recorded wire-key list and TypeScript's `hasOnly` mirror in a single commit.
The interesting half was the loader. `setDocumentUTCOffset` was specified to accept whatever the loader
accepts, and the loader's pattern admitted `+99:99` — a value that loaded and then failed to draw every
date in the document. Of the three ways out, the ruling took the third: repair the pattern to the range
`folio-format.md`'s `utcOffset` row always specified, and export `IsUTCOffset` as the **single**
predicate both doors ask. `IsLocale` is its twin over the closed locale map. Command and loader now
agree by construction rather than by care.

**The red-proof of that is REACHABILITY, not the regexp** — a test asserting only that the pattern
rejects a string proves the pattern changed, not that the defect closed. Reverting `utcOffsetPattern`
to the pre-ruling `^[+-][0-9]{2}:[0-9]{2}$` reddens `TestUTCOffsetLoadRefusalIsReachableAndLocated`,
`TestIsUTCOffsetMatchesTheLoader` and `internal/expr`'s
`TestLoaderAdmitsNothingTheEvaluatorCannotParse`, while
`TestSetDocumentUTCOffsetAgreesWithTheLoader` stays **green** — and that is the point, not a gap: the
command test is written against the loader, so the two move together because they are one predicate.
Eight numbered red-proofs are recorded in `document_settings_command_test.go`, each naming the
production line mutated. The rotation proof is the load-bearing one: swapping the two arms' write
targets reds eight tests, and only because every accepting test asserts the OTHER field is unchanged —
a "some byte moved" assertion would have stayed green, which is how that same rotation shipped green
twice earlier in this run.

**`Z` stays in the evaluator and out of the loader, on ONE ground.** `folio-format.md`'s `utcOffset`
row says *"Fixed offset, `±HH:MM`"*, and `Z` is not that; excluding it implements the format exactly as
excluding `+99:99` does. `parseUTCOffsetMinutes` keeps admitting it because it serves a second
population — offsets inside RFC 3339 timestamps arriving in report *data*, where `Z` is the canonical
UTC spelling and is required. **The canonicalization argument that was first given for refusing `Z` is
WITHDRAWN (D-12.C.4) and must not be restated as live**: the loader already admitted `-00:00`, which is
exactly a second byte spelling of `+00:00`, so the argument condemned the very pattern D-12.C ruled
correct — and it never bit, because values travel verbatim and there is no canonicalizer. A tombstone
sits at the authority site in `closedsets.go`, because a reversed ruling whose only trace is its
absence gets re-derived. The anti-divergence guard names `Z` as the one expected asymmetry with the
reason inside the assertion, and lives in `internal/expr` because the stage-rank wall forbids
`template` importing `expr`; that constraint is stated in the test file itself, where the next person
to move it will be standing.

**The non-increment is narrated, not silent.** `folio-format.md` gained a note in its version-rule
blockquote cluster: the loader was tightened to the `±HH:MM` the field table already specifies and **no
version increment was taken**, on three grounds — `utcOffset` is not one of the nine closed sets the
MINOR rule enumerates; D-7.3.1's pre-reader test asks what an older reader does with a newer document
and narrowing produces no new documents, only a stricter reader; and the file already records this
reader-strictness principle. A bump would also have restamped every fixture and moved its goldens for a
reason unrelated to time zones. The corpus figure was re-measured at close and reproduces exactly: 31
`.folio` files under `git ls-files --others --cached`, 24× `+00:00` and 7× `+07:00`, **zero excluded**
by the repair. (Under `--exclude-standard` it is 29/22/7 — the two extra are gitignored build copies of
the same starter document. The comment's smaller 28/21/7 figure is the tracked-only population, and it
says so.)

**The widened census caught this story's own copy.** The mirror test's "the designer holds no copy of
the tag list" check was widened from a remembered file list to a walk of all of `src/`, and it
immediately found a hardcoded `'zh-Hans'` in one of **this story's own new test files**. The copy was
fixed rather than the file exempted — the census would otherwise have shipped a fourth spelling of
AD-12's set in the very story whose dispatch names second-authority spellings as the defect it keeps
catching.

**The stale-citation finding, and the sweep that answered it.** Three line-number citations this story
wrote were **correct when written and false on arrival** — moved by the comment block that cited them.
The response was not to re-anchor the three: every line-number citation the story wrote became a symbol
reference, including the ones verified still accurate, because a citation that is accurate today and
one that rotted are indistinguishable to the next reader. Closer's check: all 19 remaining
`#L`-anchored links in `## Suggested Review Order` were re-resolved against HEAD and every one lands on
its intended symbol (`App.tsx:1730` is a 2,155-character single-line JSX return that contains all three
of `Document locale`, `Not set` and `UTC offset (±HH:MM)`).

**Decisions applied by ID.** D-12.C (repair the pattern, then reuse it) with carried conditions
D-12.C.1 (narrate the non-increment), D-12.C.2/D-12.C.3 (corpus measurement), D-12.C.4 (withdraw the
canonicalization argument). Ruling B is discharged by D-12.C; the frozen block's Ask First clause
recording the opposite working default is deliberately left unamended as a record of what was believed
when the block was locked. D-7.4.5 (projection, wire record and mirror in one commit), D-000.10,
D-000.14, D-000.21, AD-12, AD-17.

**Findings triage.** Review pass 1 ran three layers (blind-hunter, edge-case-hunter, verification-gap)
with `review_loop_iteration` staying 0 — **no `intent_gap`, no `bad_spec`, no loopback**, so nothing was
reverted and no section re-derived. Two non-frozen amendments were applied as patches and recorded in
the Spec Change Log: the offset row's label became `UTC offset (±HH:MM)` (the panel's convention puts
the unit in the label, and without it the only route to the syntax was to be refused), and `App.tsx`
gained a disabled `Not set` placeholder for the empty-locale draft — ruled a required pre-commit fix,
because otherwise the browser paints the first tag while React's value is `''`, so the control asserts
`en` for a document that has said nothing; behaviour was already safe, which is exactly why no command
assertion could see it, and the display test added for it was red-proved by deleting the placeholder.
**No numeric patched/deferred/rejected tally exists anywhere in this story's record, and the closer did
not invent one** — what the record carries is the two amendments above plus the five register entries
below.

**Deferred, with owners.** DW-209 (`parseUTCOffsetMinutes` admits a signed hour or minute field —
recorded rather than repaired), DW-210 (the command path is proved to reach the renderer for `en` and
`th` only, never `ja` or `zh-Hans`), DW-211 (nothing observes that a document command marks the preview
stale) and DW-212 (`tsc -b` does not typecheck `e2e/` either, so that suite is neither compiled locally
nor ever executed) were filed by the builder, and DW-206 was re-priced LOW→MEDIUM ("Apply page setup"
became a multi-command gesture, so it can apply partially). All five are in `deferred-work.md` inside
`52d0509`; none is re-filed here.

**Measured gates, re-run by the closer at `52d0509` on the committed tree.** `folio-go`
`go test -count=1 ./...` rc 1 — 14 packages ok and **exactly two** failures,
`TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child, the mandated permanent red; no
third. `gofmt -l .` empty, `go vet ./...` rc 0. `folio-designer` `npx vitest run` rc 0, **60 files /
844 tests** passed. `npx tsc -b` rc 0 and `npx tsc -b --force` rc 0 (forced, because the incremental
build can exit 0 having checked nothing). `npx oxlint` rc 0 with **exactly 4** `only-export-components`
warnings, freshly measured at `preview/pdf-viewer.tsx:16,17` and `App.tsx:3061,3068` — the App.tsx pair
moved from the spec's recorded `2988,2995` because this story's rows sit above them; no fifth warning.
`lint` module: `go build ./...`, `go vet ./...`, `test -z "$(gofmt -l .)"` and `go test -count=1 ./...`
all rc 0, 4 packages ok. `npm run scan:font-hosts` rc 0, **0 occurrences in 630 tracked source files**
(floor 400).

**The story's own declared blind spot is now CLOSED, and that is a measurement.** The scan discovers
its population with `git ls-files`, so while the work was unstaged this story's five new source files
were outside it and the green said nothing about them. They are committed now: the population rose
625 → 630, and the diff of the two populations at `e2ff17a` and HEAD is **exactly** those five files
and nothing else — `document-settings-command.ts`, `document-settings-command.test.ts`,
`document_settings_command_test.go`, `formatlocale_command_test.go`, `offset_divergence_test.go`. The
compensating grep was re-run anyway: 0 forbidden-host occurrences in each of the five under
`/usr/bin/grep -a` by absolute path, with a positive control firing on every one.

**Not run, and owed.** The cross-target matrix and the Playwright suite were not run — end-of-epic
cadence, due at the Epic 12 boundary gate. `folio-designer/e2e/browser-native-roundtrip.spec.ts` is
untouched, confirmed absent from the commit's file list.
