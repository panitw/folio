---
title: 'Story 13.3: The preview screen is the evidence screen'
type: 'feature'
created: '2026-09-08'
status: 'done'
review_loop_iteration: 0
baseline_commit: '40427eea1062c45443b38bf23df2e5f6d3e42083'
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative — the frozen Intent below governs implementation. Rewritten at close to describe what
actually shipped.*

The preview screen shows you a PDF and, in small grey type at the bottom, a hash. The hash is the
whole point of the product: it is supposed to prove that the document on screen is the exact
document the engine produced. Today it is a footnote, and — worse — nobody ever checks it. The
browser is told a hash by the engine and prints it without ever computing the hash of the bytes it
is holding.

This story builds the evidence rail the design drew: a panel down the right-hand side carrying the
facts about the render (engine version, target, page count, how long it took, how big it is), the
hash in its own bordered block in monospace so a person can actually compare it by eye, and the
diagnostics with their counts and a legend. Re-render and Save PDF move here, next to the evidence,
instead of being buried in a tab.

And before any of that is shown, the browser now computes the SHA-256 of the PDF it is holding and
refuses to display a preview whose hash does not describe its own bytes. Asking someone to compare a
hash by eye that we never checked ourselves would be worse than not showing it at all.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The preview's central claim — *this is the exact production document, and here is the
hash that proves it* — survives on screen as one grey footnote line (`App.tsx:2157`,
`.preview-evidence`), and the hash it prints has never been checked against the bytes it claims to
cover (DW-270). Nothing in `folio-designer/src` recomputes SHA-256 over the PDF; the digest is
admitted by shape alone (`/^[a-f0-9]{64}$/`). The render's own facts — engine version, elapsed —
never cross the wasm boundary at all, and Re-render and Save PDF sit inside the Inspector's
`hidden={inspectorTab !== 'properties'}` tabpanel, where an author who left the Inspector on DATA
cannot reach them and a screen reader is not told they exist (DW-281, an owner request).

**Approach:** Build the evidence rail the design draws, in the Inspector's column in Preview mode:
RENDER facts, OUTPUT HASH as its own bordered mono block, DIAGNOSTICS with counts, zero state and
shape legend, and a paired Re-render / Save PDF action row at its foot — rendered outside any
`hidden` tabpanel. Extend the wasm render reply by exactly two fields (`elapsedMs`, `version`),
measured and read inside the engine. And close DW-270: recompute the digest browser-side at install
time and refuse to present a preview whose digest does not describe its own bytes.

## Boundaries & Constraints

**Always:**

- Every value in the RENDER block comes from the render that actually happened, or it is not shown.
  Five values ship: `engine` (the engine's own `folio.Version` constant, verbatim), `target`,
  `pages`, `elapsed`, `size`. **`rows` is deliberately omitted** — see Design Notes.
- `elapsed` is measured **inside the engine**, bracketing only the call to `folio.Render`, never
  around the worker round trip. A number that included `postMessage` and byte transfer would satisfy
  the words of the acceptance criterion while breaking its sentence.
- The engine version is printed **exactly as the constant reads** (`0.0.0-dev` today). Never a
  version number invented for display, and never the mockup's `folio-go v0.1`. On the one surface
  whose rule is *never print an affirmation you cannot earn*, a fabricated version is that defect.
- The byte-identity sentence is exactly: *Byte-identical across darwin/arm64, linux/amd64,
  linux/arm64 and js/wasm — proven by the build's cross-target matrix, not compared here.* No
  check-circle, no affirmation of a live comparison.
- That sentence is **withheld entirely for a stand-in (no-data) preview**. D-13.4.1 rules a no-data
  hash is not evidence of cross-target equality, and 13.4 shipped `Stand-in local digest` for exactly
  that reason. Every stand-in withholding 13.4 built is preserved.
- A preview whose recomputed SHA-256 does not equal the digest the engine reported is **never
  installed and never displayed**. The refusal is stated in the author's own terms.
- `Render`'s exported signature and `folio.Result`'s shape are untouched. The only Go files this
  story edits are under `folio-go/wasm/`.
- Both hand-enumerated protocol hops (`engine.worker.ts:134`, `engine-client.ts:111`) name both new
  fields, and a test reds if either site drops one.
- Two accents hold: cyan is structure/authority (the hash), amber is data only (diagnostics). Every
  machine-readable value is mono.

**Ask First:**

- **The Inspector's structure in Preview mode.** This spec renders the evidence rail as a sibling of
  the tabpanels — always present, never inside `hidden` — and leaves the INPUTS tab holding
  `ParameterEditor`, which is what "a collapsed section" already is. If the intent was a literal
  `<details>` inside the rail with the tablist removed in Preview, that removes DataPanel from
  Preview, which no acceptance criterion authorises. Confirm before implementing the alternative.
- Any change to `folio.Result`, `Render`'s signature, or `EngineDiagnostic`'s field set.
- Adding a design token, a `@media` rule, or a colour literal to `App.css`.
- Any need to run the Playwright suite (it triggers `npm run build` through `playwright.config.ts`).

**Never:**

- Never add `rows`, a page number on a diagnostic, or any field requiring a change to `Render` or to
  `EngineDiagnostic`. Both were ruled out by the owner on 2026-09-08.
- Never render the PAGES thumbnail rail or remove the component palette — that is the deferred goal A.
- Never claim the tab compared itself to a native render.
- Never present `dataPath` as though it is guaranteed to be a data binding; the engine also emits
  structural paths (`bands.content.e7`).
- Never regress 13.4: the stand-in notice, `NO-DATA LAYOUT PREVIEW`, the stand-in-qualified export
  label, and `Stand-in local digest` all stay.
- Never `git add`, commit, stash, checkout, reset, revert, restore, branch, or push.
- Never edit `sprint-status.yaml`, `deferred-work.md`, or anything under `fixtures/declared-variants/`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Current exact render | render reply with `pdfSha256` matching the bytes' real SHA-256 | rail shows five RENDER values, the hash block in mono across two lines, the byte-identity sentence, diagnostics counts | N/A |
| Digest mismatch | reply whose `pdfSha256` does not equal `SHA-256(bytes)` | preview is **not** installed; `previewStatus` becomes `error` with a stated reason naming the mismatch | refusal, not a silent display |
| Stand-in (no-data) preview | `preview.standIn === true` | hash block shows `Stand-in local digest`; the byte-identity sentence is **absent** | N/A |
| Stale preview | `previewStatus === 'stale'` | rail says stale; RENDER values are marked as describing the earlier render, never presented as current | N/A |
| Failed render | `currentFailure` set | rail's DIAGNOSTICS header reads `errors 1`; the square-solid legend row is the one that applies | existing `PreviewFailure` card, unchanged |
| Zero diagnostics | `diagnostics: []`, render proceeded | DIAGNOSTICS states zero explicitly plus the shape legend, so it reads *checked*, not *nothing here* | N/A |
| Diagnostic with `elementId` | element still present at the admitted revision | location line reads `dataPath · kind · band <name>`, joined locally against the canvas projection | join omitted when the element is absent |
| Diagnostic without `elementId` | `elementId === ''` | location line carries `dataPath` alone; no Locate control (existing behaviour) | N/A |
| Very fast render | engine elapsed rounds to 0 ms | `0 ms` is displayed and transmitted; the field is not `omitempty` on either Go struct | a dropped `0` would be a silent lie |
| Identity-only reply | `identity` operation | reply carries neither new field; admission still passes | all-or-nothing arm |

</frozen-after-approval>

## Code Map

Anchors re-derived by symbol at `40427ee` (D-000.4), each read directly rather than relayed; every
search used `grep -a` (DW-256). HEAD moved once under this plan gate, `631a9c9` → `40427ee`
("Close 13.2, and backfill the two rulings no decision log could find"), the orchestrator's own and
touching only `_bmad-output/`; every code anchor below was measured at `40427ee`.

**The Go side — the only two files this story edits under `folio-go/`.**

- `folio-go/wasm/engine.go:41-46` — `RenderResult{PDFSHA256, Identity, Revision, Diagnostics}`, json
  tags `pdfSha256`/`identity`/`revision`/`diagnostics`, **none `omitempty`**. Gains `ElapsedMs int64`
  and `Version string`. `:192` is `result, err := folio.Render(...)` — the single call to bracket, and
  the tightest scope that means *the render the engine actually performed*: it excludes
  `ParseTemplate` (`:188`), the SHA-256 (`:196`) and `PreviewIdentity` (`:197`). `:202` is the return
  literal. `:12` already imports the `folio` package, so `folio.Version` compiles here today.
- ⚠ **`time` is legal here and nowhere near here.** `lint/internal/rules/forbiddenimports.go:67-72`
  bans `time`/`os`/`math/rand`/`net`, enforced over two populations:
  `forbiddenimports_test.go:29-31` scans `folio-go/internal/` recursively, and `:258-276` scans the
  targets from `findFolioGoScanTargets` (`:325-348`), which is a **flat `os.ReadDir` over `folio-go/`
  with `if e.IsDir() … continue`** — verified by reading it. `folio-go/wasm/` is in neither. AD-1's
  own title is "The determinism boundary is a **directory** boundary".
- ⚠ **`Duration.Seconds()` returns `float64` and would red `folio-go/internal/arch_test.go:288-322`
  `TestNoFloat64UnderModule`**, which scans the whole module root recursively, build tags ignored,
  and `folio-go/wasm/` IS in that population. Use `.Milliseconds()` (`int64`).
- `folio-go/wasm/cmd/engine/main.go:36-50` — the `response` struct. Almost every field is
  `omitempty`; `Diagnostics` (`:49`) deliberately is not, and its comment at `:45-48` is the
  precedent: *"JavaScript treats [] as evidence, while a missing/null field is a protocol
  violation."* **`ElapsedMs` must follow `Diagnostics`, not its neighbours** — `omitempty` would drop
  a legitimate `0 ms`. `:211` is the `case "render"` return literal. `:332` `marshal` is the whole
  JS-facing path and needs no change.
- `folio-go/version.go:9` — `const Version = "0.0.0-dev"`. `folio-go/preview_identity.go:19` already
  folds `folio-version` into the AD-18 identity hash, so surfacing it makes explicit a value the
  preview key already depends on. `folio-go/diagnostic.go:49` records that no git tag names a release.
- **Read-only, do not change:** `folio-go/diagnostic.go:453` `type Result struct { Bytes []byte;
  Diagnostics []Diagnostic }` and `folio-go/render_entry.go:160` `func Render(...) (Result, error)`.

**The protocol — three hops, two of which drop unknown fields silently.**

- `folio-designer/src/engine-protocol.ts:173` — `PreviewEvidence`, and `:408` `isPreview`, whose
  `hasOnly(value, ['revision','identity','pdfSha256','diagnostics'])` is a **subset** check (`:399`).
  Both new fields go in the **render-only paired arm** alongside `pdfSha256`/`diagnostics`, because an
  `identity` reply carries only `{revision, identity}`. ⚠ If this list is not widened, the failure is
  not a stale preview: `parseInbound` returns `undefined` → `EngineClient` fails `PROTOCOL_INVALID`
  → **the worker is terminated and every pending request rejected** (`engine-client.ts:87`).
  The envelope's own `hasOnly` at `:697` needs no change — the fields go inside `preview`.
- ⚠ `folio-designer/src/engine.worker.ts:134` — builds `preview` by naming each member. A Go field
  not named here is **dropped before `parseInbound` ever sees it**.
- ⚠ `folio-designer/src/engine-client.ts:111` — rebuilds it again, member by member, inside
  `deepFreeze`. Same trap. `:10` `EngineResult` re-declares the preview shape structurally, a **third**
  site. The comment at `:113-128` documents this exact failure happening to Story 12.3's table
  projection: *"It passed `isTableColumns`, reached here, and was DROPPED — silently, before App.tsx
  ever saw it, with no protocol failure and nothing in the DOM to say so."* `:155-156`
  `matchesOperationPayload` is the natural second gate for render-vs-identity presence.

**App — `folio-designer/src/App.tsx` (4096 lines).**

- `:147` `PreviewRecord` — gains `elapsedMs`, `version`. The comment at `:140-146` explains why 13.4
  put `standIn` **on the record** rather than deriving it; the same argument applies here.
- `:548` `installPreview` — the single writer; it validates nothing, and is also called with
  `undefined` from `invalidatePreview`. It is synchronous, so the digest check cannot live inside it.
- `:587` `runPreview`; `:635` the validation guard; `:636` the `installPreview(...)` call. ⚠ **The
  digest check goes between `:635` and `:636`**, and because `crypto.subtle.digest` is a **new
  suspension point**, `current(identity)` must be re-checked after the await — every other await in
  this function already is.
- `:210` `preview` state, `:224` `previewPages` (Story 13.2's stored page count — the viewer's count
  via pdf.js, a different provenance from the engine-reported values; say so in the rail's markup
  comment), `:1831` `admittedPreview`, `:1984` `currentDiagnostics`, `:1985` `currentFailure`,
  `:1964` `noDataPreview`, `:1965` `standInNotice`.
- `:1872` `pdfExportStale`, `:1873` `pdfExportQualifier`, `:1874` `pdfExportLabel`, `:1875-1878`
  `pdfExportUnavailable`, `:1883` `exportPreviewPdf` — all move to the rail **unchanged**. ⚠ The
  `#preview-pdf-export-reason` note must travel **with** the button: the button's `aria-describedby`
  points at it by id, and splitting them silently breaks the association.
- `:2169` `<aside className="inspector-panel" aria-label="Inspector">`; `:2171` the properties
  tabpanel, `hidden={inspectorTab !== 'properties'}` — the `hidden` that is DW-281's defect.
  `:51` `inspectorTabs` maps `properties` → `PROPERTIES`/**`INPUTS`**, i.e. the third element is the
  Preview-mode label. `:2172` the DATA tabpanel — **untouched**.
- `:2157` the whole preview `<main>` on one line, including `.preview-evidence`, whose current text is
  `` `${preview.standIn ? 'Stand-in local digest' : 'Historical producer digest'} ${preview.digest}` ``
  + `` ` · ${preview.diagnostics.length} diagnostics retained` ``.

**Diagnostics — `folio-designer/src/preview/diagnostic-presenter.tsx` (26 lines).**

- `PreviewDiagnostics` / `PreviewFailure` / `DiagnosticLocation`. The dismissal key is
  `` `${index}:${severity}:${code}:${elementId}:${dataPath}:${message}` `` — index-prefixed and
  pinned verbatim by `diagnostic-presenter.test.tsx`. **Do not change the key.** The `Locate in
  Design` label becomes `Locate on canvas` per the AC; that string is pinned and its test updates.
- ⚠ `engine-protocol.ts:172` — `EngineDiagnostic = { severity: 'warning'; code; elementId; dataPath;
  message }`, admitted at `:407` by `hasExactKeys` (**exact**, so a sixth wire field is *rejected*,
  not ignored). `severity` is the literal `'warning'` — **there is no error severity on this type.**
  So the DIAGNOSTICS header's error count is `currentFailure ? 1 : 0`, not a count over the array.
  This is the honest mapping and it matches the legend exactly: triangle/dashed = render proceeded
  (the diagnostics array), square/solid = render failed (`PreviewFailure`).
- Element **kind** and **band** are recoverable by a local join `elementId` →
  `snapshotRef.current.canvas.components`, honest only while `admittedPreview` holds — the same guard
  `locateDiagnostic` (`:1833`) already applies.

**Styling — `folio-designer/src/App.css` (623 lines), `tokens.css`.**

- `App.css:46` `.workbench { … grid-template-columns: var(--palette-width) minmax(0, 1fr)
  var(--panel-width); }` — **unchanged by this story.** The rail lives inside the existing third
  column, which is why no grid edit and no breakpoint are needed.
- `tokens.css:16` `--panel-width: 300px`. The mockup's rail is 320px. **Reuse `--panel-width`** and
  accept 300: `design-contract.test.ts:21` pins `designTokenSets` to DESIGN.md's token names, so a new
  width token would require editing a planning artifact.
- `App.css:48` `.inspector-panel { … grid-template-rows: auto minmax(0, 1fr); overflow: hidden; }` —
  needs a Preview-mode modifier adding a row for the rail.
- ⚠ `canvas-authority-contract.test.ts:342-345` asserts `App.css`'s extracted `@media` list **equals**
  `['prefers-reduced-motion: reduce']`. D-000.8 records this as load-bearing for 13.3 by name: **no
  responsive breakpoint.** `App.css:540` is the only one.
- ⚠ `design-contract.test.ts:84-91` bans `#hex`/`rgb(`/`hsl(` and any `border-radius` not
  `var(--radius…)` in `App.css`. `:104-110` reserves the solid-danger card and square marker to
  `.preview-failure` and pins that rule's opening declarations verbatim — **do not mint a second
  danger style** for a digest mismatch. `:145-152` limits `--type-display` and `--type-numeric-lg` to
  **exactly one use each**, so the rail's stat values use `--type-mono`, not the large numeric token.
- Reusable: `.section-label` (`:56`, 9px/0.1em — the design's `OUTPUT HASH`/`DIAGNOSTICS` header
  grammar), `.honest-note` (`:61`), `.file-button` (`:38`), `.diagnostic-list`/`.diagnostic-card`
  (`:510-512`, `:531-534` — already `1px dashed var(--color-bind-edge-dash)` with a
  `3px dashed var(--color-bind)` left edge, i.e. the mockup's card modulo padding).
- Accent tokens: structure/authority `--color-select-bright` (the hash), data `--color-bind`,
  `--color-bind-text`, `--color-bind-muted`, `--color-bind-tint-warm`, `--color-bind-edge-dash`;
  `--color-ok`, `--color-danger`; surfaces `--color-panel`, `--color-base` (the recessed hash box),
  `--color-line-strong` (its border), inks `--color-ink-high`/`--color-ink`/`--color-ink-faint`/
  `--color-ink-ghost`.

**Guards that will bite.**

- ⚠ `folio-designer/src/preview/preview-authority-contract.test.ts:17-24` pins six App.tsx literals
  and three negatives: `'new PreviewWorkScheduler()'`, `'canInstallPreview('`,
  `'id="preview-freshness-status"'`, `"setPreviewStatus('stale')"`, and — attributes **adjacent and
  in this order** — `'id="preview-freshness-status" className="preview-status" role="status"'`; plus
  `not.toContain('<p role="alert">Local PDF render failed')`. **A digest-mismatch message must not be
  written in that forbidden shape**; route it through `previewIssue`/`previewStatus`.
- ⚠ `folio-designer/src/App.test.tsx:7911` — the self-count guard. It takes the **last**
  `// <ALLCAPS> tests.` comment before `describe('Story 17.1` (currently `TWENTY-THREE` at `:7312`,
  `describe` at `:7335`) and compares it to the count of `\n  it(` from there to EOF. **Put 13.3's
  block before `:7335`, and write no `// <ALLCAPS> tests.` comment before it.** Its companion at
  `:5974` pins each story header adjacent to its own describe — do not splice between a pair.
- ⚠ `canvas-authority-contract.test.ts` bans all DOM measurement repo-wide; the `src/preview/`
  carve-out is **not** a directory waiver — it is two names (`host.clientWidth`/`clientHeight`) inside
  one function in one file, with `:594` asserting **exactly two** occurrences. A new file under
  `src/preview/` gets no measurement waiver. This story needs none.
- ⚠ **20 existing fixture sites** carry `pdfSha256: 'a'.repeat(64)` over `new Uint8Array([9]).buffer`,
  which is not that buffer's digest: 15 in `App.test.tsx`, 3 in `engine-protocol.test.ts`, 1 in
  `engine-client.test.ts`, 1 in `DataPanel.test.tsx` (counted with `grep -ac`). Every one that must
  install a preview needs its digest computed from its own fixture bytes by a shared helper.
- `oxlint` baseline is exactly **4** `react(only-export-components)`, measured at `40427ee`:
  `preview/pdf-viewer.tsx:17:14`, `:18:14`, `App.tsx:3830:14`, `:3837:17`. A new `.tsx` under
  `src/preview/` exporting **only components and types** adds none (`diagnostic-presenter.tsx` is the
  proof). Any non-component value export adds a fifth — put helpers in a sibling `.ts`.
- `crypto.subtle` is available unstubbed in this suite: `vite.config.ts:33` sets `environment:
  'jsdom'` with a 5-line setup file that polyfills nothing, and `font-store.test.ts:114-121` already
  cross-checks `storedFaceKey` against Node's `createHash('sha256')` in jsdom. `font-store.ts:182` is
  the hex-encoding idiom to mirror (a `Uint8Array` **view**, lowercase hex, so the result satisfies
  `/^[a-f0-9]{64}$/`).

## Tasks & Acceptance

**Execution:**

- [ ] `folio-go/wasm/engine.go` -- add `ElapsedMs int64` (`json:"elapsedMs"`) and `Version string`
      (`json:"version"`) to `RenderResult`, neither `omitempty`; import `time`; bracket **only** the
      `folio.Render` call at `:192` and record `.Milliseconds()`; set `Version: folio.Version` in the
      return literal. Never `.Seconds()` — `float64` reds `TestNoFloat64UnderModule`.
- [ ] `folio-go/wasm/cmd/engine/main.go` -- add the two fields to `response` **without `omitempty`**,
      following `Diagnostics`' documented precedent, and populate them in the `case "render"` literal.
      A dropped `0 ms` would be a silent lie about a fast render.
- [ ] `folio-go/wasm/engine_test.go` -- extend the render test to assert `Version` equals
      `folio.Version` and `ElapsedMs >= 0`, and that a render reply carries both.
- [ ] `folio-designer/src/engine-protocol.ts` -- widen `PreviewEvidence` and `isPreview`'s `hasOnly`
      list with `elapsedMs`/`version`, placing both in the **render-only paired arm** with
      `pdfSha256`/`diagnostics`; bound them inline (`Number.isSafeInteger && >= 0`; a non-empty
      version string with a length cap). Introduce **no** exported `MAX_*` numeral — that would force
      a ninth pair into `engine-bounds-mirror.test.ts`, whose `pairs` length is hard-pinned at 8.
- [ ] `folio-designer/src/engine.worker.ts` -- name both new fields in the `preview` literal at `:134`
      and on `WasmResponse`.
- [ ] `folio-designer/src/engine-client.ts` -- name both fields in the `preview` rebuild at `:111`, in
      the `EngineResult` type at `:10`, and require their presence for `render` (and their absence for
      `identity`) in `matchesOperationPayload` at `:155-156`.
- [ ] `folio-designer/src/preview/pdf-digest.ts` -- **new**, a `.ts` so no lint warning: export an
      async `pdfDigest(bytes: ArrayBuffer): Promise<string>` mirroring `font-store.ts:182` (a
      `Uint8Array` view, lowercase hex). Do not reuse `storedFaceKey` — its name would lie.
- [ ] `folio-designer/src/App.tsx` -- add `elapsedMs`/`version` to `PreviewRecord`; between `:635` and
      `:636` await `pdfDigest(result.bytes)`, compare against `result.preview.pdfSha256`, re-check
      `current(identity)` after the await, and on mismatch refuse to install and set `previewStatus`
      to `error` with a stated reason. Render `PreviewEvidenceRail` in the Inspector column in Preview
      mode, **outside any `hidden` tabpanel**; move `Render local PDF` (as `Re-render`), the export
      button and its `#preview-pdf-export-reason` note into the rail's action row, unchanged.
- [ ] `folio-designer/src/preview/evidence-rail.tsx` -- **new**, components and types only: the RENDER
      block (five values), the OUTPUT HASH block, DIAGNOSTICS with counts, zero state and legend, and
      the action row. Withhold the byte-identity sentence when `standIn`.
- [ ] `folio-designer/src/preview/evidence-rail-facts.ts` -- **new**, a `.ts` for the non-component
      helpers: byte-size and elapsed formatting, the two-line hash split, and the `elementId` →
      component join yielding kind and band.
- [ ] `folio-designer/src/preview/diagnostic-presenter.tsx` -- re-dress only: rename `Locate in
      Design` to `Locate on canvas` and render the three-part location. Do **not** change the
      dismissal key or the announcement strings.
- [ ] `folio-designer/src/App.css` -- add the rail's rules and the `.inspector-panel` Preview-mode
      row. Tokens only: no hex, no `rgb(`/`hsl(`, no raw `border-radius`, no `@media`, and not
      `--type-numeric-lg`.
- [ ] `folio-designer/src/preview/evidence-rail.test.tsx` + `pdf-digest.test.ts` -- **new**, covering
      every I/O matrix row.
- [ ] `folio-designer/src/App.test.tsx`, `DataPanel.test.tsx`, `engine-client.test.ts`,
      `engine-protocol.test.ts` -- update the 20 fixture sites to carry their own bytes' real digest
      via one shared helper, and add the protocol arms below. Add 13.3's block **before line 7335**.
- [ ] `folio-designer/e2e/preview-evidence-rail.spec.ts` -- **new**, the browser witness for the
      layout half, which no unit test can see (jsdom applies no stylesheet — DW-289). Assert via
      `boundingBox()` that the rail occupies the Inspector column at its declared width, that the
      action row sits at the rail's foot, and that the hash block's full 64 characters are present and
      wrapped. ⚠ If it touches Save PDF it MUST `addInitScript` both pickers to `undefined` to force
      the fallback tier, or it will hang for 90s. Run the **full** suite; report 41/41.

**Acceptance Criteria:**

- Given a render reply whose `pdfSha256` does not equal the SHA-256 of the bytes it arrived with,
  when the preview is processed, then it is **not** installed, nothing is displayed, and the status
  states the mismatch. **Mutation-prove by corrupting one byte of the fixture** and confirming the
  test reds; restore by `cp` and report the restore digest. (DW-270)
- Given a correct reply, when the preview installs, then the recomputed digest equals the displayed
  one — and the assertion's two sides are independently derived, never one computed from the other
  (D-11.2.2, D-11.2.8).
- Given the two new protocol fields, when either `engine.worker.ts:134` or `engine-client.ts:111`
  omits one, then a test reds **naming the missing field**. **Mutation-prove by deleting one field
  from one site**, confirm it reds naming it, restore by `cp`, report the restore digest. This is the
  Story 12.3 failure shape and is the most dangerous thing in this story.
- Given a completed render, when the rail is shown, then RENDER carries exactly five values —
  `engine` reading the `folio.Version` constant verbatim, `target`, `pages`, `elapsed`, `size` — and
  `rows` appears nowhere.
- Given the OUTPUT HASH block, when it is shown, then the digest is mono, inside its own bordered
  block, wrapped across two lines, and the full 64 hex characters are present — the mockup's 32-char
  display is a mockup artifact, and a hash to be compared by eye must be the whole hash.
- Given a non-stand-in current preview, when the hash block is shown, then it carries exactly
  *Byte-identical across darwin/arm64, linux/amd64, linux/arm64 and js/wasm — proven by the build's
  cross-target matrix, not compared here.*, with no affirmation of a live match.
- Given `preview.standIn === true`, when the hash block is shown, then that sentence is **absent** and
  the digest is labelled as a stand-in digest. **Mutation-prove that the withholding is real** by
  flipping `standIn` in a fixture and confirming the sentence's presence changes.
- Given a render that proceeded with zero diagnostics, when DIAGNOSTICS is shown, then it states zero
  explicitly and carries both legend rows, so it reads *checked* rather than *nothing here*.
- Given a failed render, when the header counts are shown, then `errors` reads 1 — sourced from
  `currentFailure`, because `EngineDiagnostic.severity` is the literal `'warning'` and can never
  carry an error.
- Given a diagnostic whose element is present at the admitted revision, when its card is shown, then
  the location reads `dataPath · kind · band <name>`; given an absent `elementId`, then `dataPath`
  alone, and no Locate control.
- Given a stale preview, when the rail is shown, then it says so, and no RENDER value is presented as
  describing the current document (UX-DR14).
- Given the DATA tab is selected in Preview, when the rail is inspected, then Re-render, Save PDF and
  the export reason line are all still in the accessibility tree — the defect DW-281 records.
  **Mutation-prove by re-parenting the action row into the `hidden` tabpanel** and confirming the test
  reds.
- Given every new control, when it is used by keyboard alone, then it is reachable, labelled and
  operable (UX-DR25).
- Given the rail's CSS, which no unit test can observe because jsdom applies no stylesheet, when the
  browser witness runs, then it measures the rail's placement and width and the hash block's wrapped
  64 characters. **Mutation-prove that the witness can fail** by deleting the rail's width rule and
  confirming it reds; restore by `cp` and report the restore digest. A guard that cannot fail is
  worse than none (D-000.9), and DW-289 is the registered precedent for this exact blind spot.

## Spec Change Log

- **Plan gate, 2026-09-08 — step-01 scope split (owner decision).** Story 13.3 carried two
  independently shippable deliverables. **Goal A — the PAGES thumbnail rail** (thumbnails, numbering,
  current-page and diagnostic marking, click-to-navigate, `… 29 more` truncation, and not rendering
  the component palette in Preview) was **deferred**; this spec covers **goal B, the evidence rail**,
  only. Two decisions are recorded here so goal A's builder inherits them rather than rediscovering
  them: **(i)** whichever half ships first owns the diagnostic→page derivation — this spec does not
  build one, because the owner ruled out a page number on the diagnostic (below), so goal A must
  source page marking from its own thumbnail enumeration; **(ii)** `App.css:46`'s
  `grid-template-columns` line is shared and takes two commits — this story does not touch it.
- **Plan gate, 2026-09-08 — owner ruling, `rows` dropped (Q1).** The coordinator's summary to the
  owner said "add row count to the wasm reply". **That premise was false for this field**, and the
  coordinator corrected it rather than stretching the ruling. `elapsed` and `version` are wasm-reply
  changes; `rows` is not. The count exists only inside the renderer (`folio-go/table_render.go:961`),
  and `folio.Result` is `{Bytes, Diagnostics}` — surfacing it means changing `Render`'s exported
  contract, three files deep, the first such change in this epic, where 13.4's precedent explicitly
  preserved *"`Render` itself is still untouched"*. **KEEP on any re-derivation:** the second
  argument, which the owner also weighed — a scalar row count cannot distinguish *no tables* from *a
  table bound to an empty collection*, and reports `403` for a 3-row summary beside a 400-row ledger,
  a number describing no table in the document and changing for two unrelated reasons.
- **Plan gate, 2026-09-08 — owner ruling, no page on the diagnostic (Q2).** The card ships a
  three-part location. A page number is the only one of the design's four components that is a
  property of **layout** rather than of the element, and computing layout in the browser is forbidden
  outright by the canvas-authority contract; adding it to `EngineDiagnostic` means widening a
  `hasExactKeys` guard and a third engine extension. **KEEP:** `dataPath` is not guaranteed to be a
  data binding — the presenter's own tests carry structural paths like `bands.content.e7` — so the
  rail renders it without claiming it is one.
- **Plan gate, 2026-09-08 — lead ruling, the affirmation wording (Q3).** No numbered ruling existed;
  `epic-11-14-decision-log.md:385` registered it as open. Approved as written, and see Design Notes
  for the two asymmetries and the honest limit that must survive re-derivation.
- **Plan gate, 2026-09-08 — lead ruling, Inspector structure (Q4a).** The rail takes the Inspector's
  column in Preview; `ParameterEditor` stays reachable; the DATA tab is untouched; Re-render and Save
  PDF are paired in the rail, which discharges **DW-281 — an owner request**, verbatim *"Later this
  button should be moved to the preview area."* Cite it as an owner request, never as a finding
  (D-13.4.4 records that its attribution was once wrongly stripped).
- **CHECKPOINT 1, 2026-09-08 — a correction to the Q4(a) wording, not a deviation from it.** Q4(a) as
  delivered said both "the rail replaces the INPUTS tabpanel **content**" and "this discharges
  DW-281". **Those two pull apart.** DW-281's defect *is* that the control sits inside
  `hidden={inspectorTab !== 'properties'}` and vanishes from the accessibility tree when DATA is
  selected; a rail placed inside that same tabpanel inherits the defect and discharges nothing, and
  DW-281's own discharge clause says to render outside it. The lead confirmed the correction at
  approval. **The settled structure:** the rail is a **sibling of the tabpanels**, always present,
  never inside `hidden`; the tablist and the DATA tabpanel are untouched; `ParameterEditor` stays in
  the INPUTS tab, which is what "a collapsed section" already is. **KEEP on any re-derivation:** the
  rejected alternative — a literal `<details>` inside the rail with the tablist removed in Preview —
  would take DataPanel out of Preview entirely, a regression no acceptance criterion authorises.
- **CHECKPOINT 1, 2026-09-08 — browser run authorized as a condition, not a permission.** See
  `## Verification`. The reason is DW-289: Story 13.2 shipped with every CSS claim guarded by
  nothing, and the rail is this story's entire visual deliverable. Shipping it with the same blind
  spot, two stories after registering that exact gap, would be a choice rather than an oversight.
- **CHECKPOINT 1, 2026-09-08 — token gate answered [K] by standing ruling.** Measured ~9,719 tokens
  (38,877 chars ÷ 4), above the workflow's 1,600. The lead had already ruled "no further cut" on the
  one real seam (engine-extension-and-digest-verification versus rail-presentation), on the grounds
  that the digest verification exists *because* the rail promotes the hash to a compare-by-eye claim,
  and shipping the promotion without the verification is the one combination that makes the screen
  worse than today. Recorded rather than silently skipped. **KEEP:** the threshold is unreachable in
  this repo — no folio spec has ever met it and the Epic 8 peers ran 16,000–26,000 tokens — so
  thinning acceptance criteria to reach it would trade a real property for a metric.

## Design Notes

**Why the RENDER block has five values and not the mockup's six.** `rows` is omitted, and the
omission is a decision rather than an oversight — which is why it is stated on screen's own terms in
the spec and here. See the Spec Change Log entry for the two reasons: the exported-API cost, and the
fact that a single scalar has no defensible meaning across zero-table and many-table documents. An
omission with a stated reason is a decision; an omission without one reads as an oversight to
everyone downstream.

**The byte-identity sentence, and what it does not cover.** The approved wording is:

> Byte-identical across darwin/arm64, linux/amd64, linux/arm64 and js/wasm — proven by the build's
> cross-target matrix, not compared here.

Two asymmetries in the mockup made this necessary, and both should survive re-derivation:

- The mockup **under-claims coverage.** It names three targets. `.github/workflows/matrix.yml`'s
  `compare-render-hashes` job asserts byte-identical output across **four** — darwin/arm64,
  linux/amd64, linux/arm64 **and js/wasm** — over 25 documents, on every push and pull request.
- The mockup **over-claims liveness.** *"Matches native render"* beside a check-circle implies a
  comparison happening here. The tab cannot compare itself to a native render, and nothing in the
  browser does.

**The honest limit, recorded so a reader need not re-derive it.** The matrix's js/wasm leg runs Go
**test binaries under Node**, not the `folio.wasm` artifact the designer actually loads. What closes
that gap is `folio-designer/e2e/browser-native-roundtrip.spec.ts:408-409`, which drives the real
browser app, builds `folio-go/cmd/folio`, and asserts the native PDF is byte-equal to the
wasm-rendered one — **on linux/amd64 only**, in CI, not live in the tab. So the sentence's "proven by
the build's cross-target matrix" is exact, and its scope is: the engine source across four targets,
plus one browser-artifact-versus-native check on one of them.

**Why the version reads `0.0.0-dev` and that is correct.** `folio-go/version.go:9` is the constant;
`folio-go/diagnostic.go:49` records that no git tag names a release. Projecting it across the channel
makes it *available*; it does not make `folio-go v0.1` true. Printing a version the product cannot
substantiate, on the one surface whose rule is *never print an affirmation you cannot earn*, would be
the exact defect this story exists to prevent. If it reads oddly, that is the screen doing its job.

**Why the digest check sits where it does.** DW-270 was ruled out of Story 13.1 (Q3(a), 2026-09-06)
because an async crypto hop inside the activation-gated *save* gesture is the worst place for one.
Install time carries no such constraint. It matters most because of what this story does: it promotes
the hash from a grey footnote to a bordered block a person is **told to compare by eye**, and asking
someone to verify a claim we have not verified ourselves is worse than not showing it. Between the
engine and the screen the bytes cross four value-preserving copies — worker allocation, transfer,
`copyBytes` in `#settle`, `slice(0)` in `installPreview` — each correct today, none checked.

**Why the rail is 300px and that is not a divergence to "fix".** The mockup draws a 320px rail. This
story ships it at `--panel-width`, which is 300px, and **one decision satisfies three constraints at
once** — which is why a later reader who sees only a 20px gap and tries to close it will break two
things they were not looking at:

1. `design-contract.test.ts:21` pins `designTokenSets` to DESIGN.md's token names by exact equality,
   so minting a `--rail-width` would require editing a planning artifact — which also invalidates the
   epic-context cache every story in this epic loads.
2. Reusing the existing Inspector column means `App.css:46`'s `grid-template-columns` is **untouched**,
   which leaves that shared line free for goal A (the deferred PAGES rail), whose own change lands on
   the same declaration. Two stories editing one line is two commits; one story editing it is a
   conflict for the other.
3. A width that had to change per viewport would need a `@media` rule, and
   `canvas-authority-contract.test.ts:342-345` asserts `App.css`'s media-query list **equals**
   `['prefers-reduced-motion: reduce']` — a contract D-000.8 records as load-bearing for this story
   by name.

**The `errors` count has no source in the diagnostics array, and the coincidence is load-bearing.**
`EngineDiagnostic.severity` is the literal `'warning'` (`engine-protocol.ts:172`) — there is no error
severity on the type. An implementer counting that array for the header's error count would write a
number that can only ever be zero, and it would look correct forever. It comes from `currentFailure`.
That this **matches the shape legend exactly** — triangle/dashed = the diagnostics array (a render
that proceeded), square/solid = `PreviewFailure` (a render that failed) — is not a coincidence to
tidy away: the two counts and the two legend rows are the same partition, named twice. Do not
"simplify" the error count back onto the array.

**The props a test constructs are part of the claim it makes.** Story 13.2's DW-191 guard was
rigorous, mutation-proof, and about a viewer the application never renders, because it built its
callbacks once while `App` creates fresh ones every render. Mutation testing cannot catch that — the
mutation and the assertion agree with each other in a world the app never enters. Every fixture in
this story must supply what the real caller supplies: a render reply with **all four** render-arm
fields, bytes whose digest actually matches, and a canvas projection at the admitted revision when a
diagnostic join is under test.

## Verification

**Commands** (all paths absolute; never `cd` before a relative restore path). Baselines supplied by
the orchestrator at `631a9c9` and to be **re-measured first-hand** at `40427ee` before use — report
what they actually print, never an adjective:

- `cd /Users/panitw/Projects/folio/folio-designer && npm test` — baseline **66 files / 1043 tests**.
  Report a **diff of test names** (GONE / NEW), never a total; a total hides a swap.
- `cd /Users/panitw/Projects/folio/folio-designer && npx tsc -b --force` — expected: exit 0.
- `cd /Users/panitw/Projects/folio/folio-designer && npx oxlint` — expected: exactly **4**
  `react(only-export-components)` warnings. Measured at `40427ee`: `src/preview/pdf-viewer.tsx:17:14`,
  `:18:14`, `src/App.tsx:3830:14`, `:3837:17`. **The count and the rule are the invariant; the lines
  are not** — editing above them moves the anchors, so report the new ones rather than assuming these.
  A fifth warning means a non-component value was exported from a `.tsx`.
- `cd /Users/panitw/Projects/folio/folio-designer && npm run test:e2e:compile` — expected: exit 0.
- `cd /Users/panitw/Projects/folio/folio-go && go test -count=1 ./...` — baseline **2273 pass / 2 fail
  / 5 skip**, failing ONLY `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)`, checked by
  enumerated **NAME** and never by exit code. **A third distinct failure is a hard stop.**
- `cd /Users/panitw/Projects/folio/lint && go test -count=1 ./...` — expected: four `ok`. `-count=1`
  is mandatory: these tests `ReadDir` the `folio-go` tree, which the test cache does not track.
- `gofmt -l /Users/panitw/Projects/folio/folio-go /Users/panitw/Projects/folio/lint` — expected:
  **empty output**. Absolute paths only; judge by output **shape**, never by exit code.
- `git status --porcelain` — expected: only this story's files. **Never `git add -A`, never commit.**

**THE BROWSER SUITE IS AUTHORIZED FOR THIS STORY — as a condition, not a permission.**

- `cd /Users/panitw/Projects/folio/folio-designer && npm run test:e2e` — baseline **40/40** at
  `3dd6be9`; this story's witness makes **41**. Run the **full** suite, never this story's spec alone:
  a change that greens its own witness while breaking a neighbour is the failure shape this run keeps
  finding. Report the count.
- **Why it is a condition.** DW-289 records that Story 13.2 shipped with every CSS claim guarded by
  nothing — jsdom applies no stylesheet, and eight mutations including deleting a design token and
  `safe center` all left the suite green at 1043/1043. The rail is this story's **entire visual
  deliverable**. Shipping it with that same blind spot, two stories after registering the gap, would
  be a choice rather than an oversight. So the layout half needs a browser witness, and this is it.
- ⚠ **If the new witness waits on a `filechooser`** (the Save PDF control in the action row does),
  it MUST force the fallback tier with
  `addInitScript(() => Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }))`.
  Headless Chromium has the File System Access API and otherwise emits no chooser event at all: two
  specs shipped without this and timed out at 90s the first time anyone ran them.
- ⚠ **The run rewrites `folio-designer/evidence/story-6.7-roundtrip-manifest.json`** (DW-294). Leave
  that file alone, exclude it from this story's change list, and report it — the orchestrator reverts
  it.

**`npm run build` is still not to be run directly.** `playwright.config.ts:14-17`'s `webServer` runs
`npm run build && npm run preview` unconditionally (`reuseExistingServer: false`), so the authorized
`test:e2e` above triggers the build transitively; that is the only route by which it runs.

**Report the browser suite as LOCALLY EXECUTED, never as CI-executed.** `origin/main` is live, but no
CI run has been observed by anyone. A test that has never been executed is not coverage, whatever the
config says — and 13.2's spec asserted CI execution in wording that was false when written.

**Manual checks:**

- Confirm `--panel-width` is still 300px and the rail renders at that width; the 20px deviation from
  the mockup's 320px is deliberate (no new token) and should be reported, not silently absorbed.
- Confirm the full 64-character digest is present in the DOM, not a 32-character mockup-style excerpt.
- Confirm the byte-identity sentence is absent from a stand-in preview and present otherwise.
- Confirm Re-render, Save PDF and `#preview-pdf-export-reason` are in the accessibility tree with the
  DATA tab selected.

## Suggested Review Order

**The claim the screen makes about itself**

- The digest is recomputed browser-side before anything is shown; DW-270's discharge.
  [`App.tsx:644`](../../folio-designer/src/App.tsx#L644)

- The await, and the `current(identity)` re-check that makes the new suspension point safe.
  [`App.tsx:657`](../../folio-designer/src/App.tsx#L657)

- Lowercase-hex SHA-256 over a `Uint8Array` view, mirroring the `storedFaceKey` idiom.
  [`pdf-digest.ts:27`](../../folio-designer/src/preview/pdf-digest.ts#L27)

- The byte-identity sentence, withheld entirely for a stand-in preview.
  [`evidence-rail.tsx:109`](../../folio-designer/src/preview/evidence-rail.tsx#L109)

**Counts that can only describe a render that happened**

- `describesCurrent` gates the zero state, so a refused render can never read as a clean one.
  [`evidence-rail.tsx:128`](../../folio-designer/src/preview/evidence-rail.tsx#L128)

- Stale and dismissed have different causes, so they get different sentences.
  [`evidence-rail.tsx:147`](../../folio-designer/src/preview/evidence-rail.tsx#L147)

- `warningsOnScreen` is the same set the cards are, asked the same way.
  [`App.tsx:2230`](../../folio-designer/src/App.tsx#L2230)

- The error count comes from the failure, the only place an error can exist.
  [`App.tsx:2231`](../../folio-designer/src/App.tsx#L2231)

**Two new facts crossing the wasm boundary**

- The elapsed bracket wraps only `folio.Render` — not parse, digest or identity.
  [`engine.go:219`](../../folio-go/wasm/engine.go#L219)

- Both fields on the reply, neither `omitempty`, so a legitimate `0 ms` survives.
  [`engine.go:66`](../../folio-go/wasm/engine.go#L66)

- The host response, the one hop no Go test compiles (`//go:build js && wasm`).
  [`main.go:218`](../../folio-go/wasm/cmd/engine/main.go#L218)

- Admission widened on the render-only paired arm; an identity reply carries neither.
  [`engine-protocol.ts:414`](../../folio-designer/src/engine-protocol.ts#L414)

- Hand-enumeration hop one — a field unnamed here is dropped silently.
  [`engine.worker.ts:137`](../../folio-designer/src/engine.worker.ts#L137)

- Hand-enumeration hop two, plus the render/identity presence gate at `:159`.
  [`engine-client.ts:115`](../../folio-designer/src/engine-client.ts#L115)

**Placement, and the DW-281 discharge**

- The rail sits outside every `hidden` tabpanel, which is what discharges DW-281.
  [`App.tsx:2225`](../../folio-designer/src/App.tsx#L2225)

- `role="group"`, not a bare `<div>`, so the label is actually announced.
  [`evidence-rail.tsx:187`](../../folio-designer/src/preview/evidence-rail.tsx#L187)

- The whole 64-character digest, keyed positionally, wrapped across two lines.
  [`evidence-rail.tsx:105`](../../folio-designer/src/preview/evidence-rail.tsx#L105)

**Guards that had to be able to fail**

- The worker round trip: catches a cross-wire the source-text scan cannot see.
  [`engine-worker-boundary.test.ts`](../../folio-designer/src/engine-worker-boundary.test.ts)

- `ElapsedMs > 0` on the multi-page fixture; the old `>= 0` was measured unfailable.
  [`engine_test.go`](../../folio-go/wasm/engine_test.go)

- The race test: holds the digest, leaves Preview mid-flight, asserts nothing installed.
  [`App.test.tsx`](../../folio-designer/src/App.test.tsx)

- Browser witness for the layout half, which jsdom cannot observe at all.
  [`preview-evidence-rail.spec.ts`](../../folio-designer/e2e/preview-evidence-rail.spec.ts)
