---
title: 'Story 8.4a — The canvas paints with the face the engine measured'
type: 'feature'
created: '2026-09-01'
status: 'done'
baseline_revision: 'dfe5129ae89fcc124d96ed4047e3a7fe6db3348f'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized']
deferred:
  - summary: >-
      A carried face is registered with no weight or style descriptors, so a bold or italic
      component gets browser-synthesized bold/italic over a 400/normal face.
    evidence: |-
      new FontFace(embeddedFaceFamily(assetKey), bytes) takes no descriptors, while TextPaint
      still emits --text-font-weight (700 when component.bold) and --text-font-style. Synthesis
      has different metrics from what the engine measured, which is the class of defect this
      story exists to remove. The projected chain entry already carries `style` and it is unused
      here. Pre-existing in kind for shipped faces too (build-wasm.mjs declares regular faces
      only), which is why this is deferred rather than patched — the fix is a chain/face-variant
      design question, not a local repair.
    location: >-
      folio-designer/src/embedded-face-registry.ts
    severity: medium
  - summary: >-
      The inline family replaces the declared stack rather than extending it, so a registered
      carried face missing a glyph falls to the browser default rather than App.css's stack.
    evidence: |-
      TextPaint emits `font-family: folio-carried-<key>` with no fallback list, and App.css:118 is
      unchanged. The implementation gates on registration so the fetch-failure row still degrades
      correctly, and argues the choice in a header comment — but for a face that IS registered and
      merely lacks a glyph, CSS's own font matching would have degraded per-glyph had the declared
      stack been appended. The intent does not settle this fork and no surface in the repo can
      observe the difference.
    location: >-
      folio-designer/src/App.tsx (TextPaint)
    severity: medium
  - summary: >-
      No concurrency cap on the carried-face asset fan-out, and release() does not abort
      in-flight requests.
    evidence: |-
      isCanvas admits up to MAX_ENGINE_FONT_FAMILIES (256) chains x MAX_ENGINE_FONT_CHAIN_ENTRIES
      (64) entries; registerCarriedFaces issues one engine.request('asset', ...) per distinct key
      in a bare loop. A document with many carried faces issues many concurrent full-font-byte
      requests down the single engine channel. release() flips `active` so nothing is added after
      supersession, but the fetches still complete and still transfer.
    location: >-
      folio-designer/src/embedded-face-registry.ts
    severity: medium
  - summary: >-
      The hand-written comment stripper does not skip regex literals, so a regex containing a
      double slash silently truncates that line for every scan routed through it.
    evidence: |-
      withoutComments tracks string and template-literal state but not regex-literal state, so
      /https:\/\// reads as a line comment. No designer source triggers it today, which is why
      it is latent rather than a live hole — but violations(), runtimeRegistrationSites and
      fontFamilyDeclarations all now route through it, so the first source that adds such a regex
      silently weakens three guards at once.
    location: >-
      folio-designer/src/canvas-authority-contract.test.ts
    severity: low
  - summary: >-
      The new test stubs install a page font set by a spelling the repaired prohibition does not
      match, and the file scans the test corpus.
    evidence: |-
      canvas-authority-contract.test.ts passes production, tests and e2e to violations().
      App.test.tsx and embedded-face-registry.test.ts both install a global face constructor and a
      document.fonts set, yet neither matches /\bdocument\.fonts\b/ or /\bnew FontFace\b/,
      because they write Object.defineProperty(document, 'fonts', ...) and class StubFace. A test
      stub is not runtime registration, so this is not a violation — but the rule this story
      revived is blind to the mechanism as this story itself writes it, and that is undisclosed.
    location: >-
      folio-designer/src/App.test.tsx
    severity: low
  - summary: >-
      Stub teardown deletes globalThis.FontFace and document.fonts rather than restoring the
      prior property descriptor.
    evidence: |-
      installStubFontSet's restore and embedded-face-registry.test.ts's afterEach both call
      Reflect.deleteProperty. jsdom provides neither today so there is no live impact, but in any
      environment that does (a jsdom upgrade, happy-dom) this permanently removes them for every
      later test in the worker. Capturing and restoring the prior descriptor is the correct shape.
    location: >-
      folio-designer/src/App.test.tsx
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Non-normative — a plain-language summary of what shipped. The intent contract below governs
implementation.*

A document can carry its own typeface inside it. The engine already drew printed pages with such a
typeface, but the design canvas in the browser had no name for it at all, so the canvas placed every run
of text at the engine's own positions while drawing it in a substitute face. The letters collided. Thai
is where a person first noticed it.

That is fixed. The engine now tells the canvas which of the document's carried assets it drew each run of
text with; the browser fetches those bytes over the channel images already travel on, registers the
typeface once for the whole document, and asks for it on exactly the runs it belongs to. If the bytes
never arrive, the canvas quietly keeps the stack it always used and carries on painting. Repeated sheets
get it too — the first attempt reached only the first sheet's copy, and review caught that.

Two things a later reader should not mistake for oversights. A document using only the typefaces the
product ships still has an older, separate mismatch; that one is a design-system decision, already ruled
and owned by a named follow-up story, and is deliberately left open. And nothing here can run a real
browser, so the tests prove the typeface is chosen, fetched, registered and asked for — not that a page
visibly renders with it. That check waits on browser end-to-end testing.

<intent-contract>

## Intent

**Problem:** Story 8.4 made the engine **measure and render** with a face the document carries in its
`assets` map, but the browser has **no CSS family for that face at all** — no `@font-face`, no name,
no bytes. `.canvas-text-fragment` (`App.css:118`) names a fixed three-family stack, so a carried face
falls straight through to generic `sans-serif`: the canvas rasterizes at the engine's x-positions with
the wrong metrics and glyphs collide. This is **DW-35 cause two**, created by 8.4 and split out of it
by **D-8.4.6**.

**Approach:** Carry each text fragment's **asset key** (present only when the engine resolved that
fragment to a carried face) through the canvas projection to the browser; fetch the face's bytes over
the **existing** media-type-agnostic `asset` operation; register them once per document as a
`FontFace` under a **name derived from the asset key**; and ask for that family on that fragment only.
Then widen — never weaken — the designer guards that were written to forbid exactly this mechanism,
and repair the `document.fonts` prohibition scan that is currently dead.

## Boundaries & Constraints

**Always:**
- **An embedded face's CSS family name is derived from its ASSET KEY, never from `font.family`**
  (**D-8.4.1**, settled by the engineering lead on AD-8 and `format-changes.md`: `font.family` /
  `font.style` are display identity, *"never used to resolve or substitute a face — resolution is by
  asset key alone"*). Deriving from `font.family` would let an embedded "Inter" collide with a shipped
  "Inter" in the browser's font registry — AD-8's own hazard one layer down. **This is a premise, not
  a question. Do not re-open it.**
- **AD-17 is unchanged: the browser never measures text.** This story changes **rasterization only**.
  Every metric and line break still comes from the engine's measure API, and Story 8.4's pin
  (`folio-go/canvas_embedded_face_test.go`) must still hold, unmodified.
- **Each re-authored guard ends up asserting a STRICTLY STRONGER claim than it does today.** A guard
  re-authored to let a feature through is the cheapest way to lose one.
- Per-fragment attribution, never per-component: a fragment is exactly one face by construction, and
  per-component would be wrong for every mixed-script element.
- The corpus is **23** documents and **none may move**. This story changes browser rasterization; a
  moved PDF digest is a defect, not a re-record.
- `main` only. Never push, never branch. Never `git add -A` / `git add .` — explicit paths.
- Root `README.md` is the owner's file (md5 `078d7d80d518d54af2fc04fb270d46b8`) — never modified,
  moved, deleted or staged.
- `fixtures/statement-signoff.json`, `fixtures/thai-stacked-marks/signoff.json` and
  `fixtures/embedded-font/signoff.json` are attestation records. No agent writes `reader`, `date` or
  `examined`, and no agent adds those fields to the transferred record.

**Block If:**
- Any planned change would move a metric, an advance, or a line break into the browser, or would make
  the canvas await font readiness before laying out. That is AD-17 and this story going wrong.
- A guard in `canvas-font-stack.test.ts` cannot be widened without weakening it, or the widened form
  cannot be red-proved by deletion.
- Closing DW-35 **cause one** (shipped-face chains: the engine measures `Noto Sans Thai` while the
  browser asks `IBM Plex Sans`) turns out to be required to satisfy an AC. Cause one's fix is a
  **design-system decision above a builder's authority** and no ruling has made it — see Design Notes,
  "Cause one stays open".
- Any task turns out to require widening `page_setup.go`'s `errors.As`-scoped `shapeSegments` arm
  (**DW-92**, open with the engineering lead).
- A `//go:build matrix` sign-off gate is left red, or a golden digest other than none moves.

**Never:**
- Never derive a CSS family from `font.family`, from a chain entry's `face`, or from a chain name.
- Never spell the `"asset:"` prefix outside `embeddedFaceName` (`embedded_face.go:87`) — in Go **or**
  in TypeScript. Writing the prefix a second time is writing the derivation a second time.
- Never register a face per component. `ImagePaint`'s effect is the closest pattern and the **wrong**
  lifetime.
- Never touch `deriveTag`, the subsetting path, or any `expected.pdf` / `expected.json`.
- Never relax, delete or narrow a prohibition to clear a red this story creates.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Carried face paints | Chain `["Noto Sans", <embedded Thai>]`, Thai text | Fragment carries the embedded entry's asset key; browser registers a `FontFace` under the asset-key-derived family and the fragment asks for it | No error expected |
| Shipped face unchanged | Chain `["Noto Sans"]`, Latin text | Fragment carries **no** asset key; the existing declared stack paints it exactly as today | No error expected |
| Mixed-script element | One element, Latin + Thai, chain mixes shipped and carried | Each fragment attributed independently; only the carried-face fragments get the derived family | No error expected |
| Shared `font.family` | An embedded face and a shipped face both named "Inter" | **Distinct** CSS families: the embedded one is asset-key-derived, so no registry collision | No error expected |
| Two embedded entries in one chain | Chain with two carried faces, both covering | Each fragment names its **own** asset key; the two get distinct families | No error expected |
| Asset fetch fails | `asset` request rejects or returns no bytes | The fragment degrades to the existing declared stack; the canvas still paints and the session stays alive | No throw; no worker termination |
| Document replaced | Open a file / new template / undo across a font-chain edit | Faces re-registered for the new document; superseded faces released | No error expected |

</intent-contract>

## Code Map

**Every anchor below re-measured at `9e77c36`.** The dispatch's anchors were carried from Story 8.4's
plan (`1446b87`) and **all of them have moved**; the corrected ones are here. Where a handed claim was
wrong on substance rather than on line number, it is marked **MEASURED-FALSE**.

### The Go seam — where the face is known and thrown away

- `folio-go/page_setup.go:116-121` — `CanvasTextFragment`, **two fields only**:
  `Text string` with tag `json:"text"`, and `X int64` with tag `json:"x"`. No face, no size, no colour.
- `folio-go/page_setup.go:1381` — **the discard site**, and the one line that must learn a third field:
  `paintLine.Fragments = append(paintLine.Fragments, CanvasTextFragment{Text: fragment.text, X: int64(x)})`.
  (The dispatch and DW-35 both say `:1344`; that is now a `positionSegments` error return.
  **MEASURED-FALSE as an anchor.**)
- `fragment` is a `textRunSource` (`render.go:105-107`), whose `face string` is set at `render.go:1771`
  from `faceSegment.face`. For a carried face that string is `embeddedFaceName(assetKey)`.
- **One face per fragment — CONFIRMED.** `faceSegment.face` is a scalar `string` (`wrap.go:28-36`);
  `positionSegments` (`render.go:1755-1805`) emits at most one `textRunSource` per segment and
  **never merges** adjacent runs. Justification splits a line into more `positionSegments` calls,
  which only increases the run count.
- `folio-go/embedded_face.go:83-87` — the derivation, and its **one construction site**:
  ```go
  const embeddedFaceNamePrefix = "asset:"
  // embeddedFaceName is the ONE construction site for that name. A caller
  // that spells the prefix itself is writing the derivation a second time.
  func embeddedFaceName(assetKey string) string { return embeddedFaceNamePrefix + assetKey }
  ```
  Doc comment `:56-82` states the precedence rule and D-8.4.1's ground verbatim.
- `folio-go/render.go:1322-1325` `fontCache.isEmbedded`; `render.go:1357-1371` `fontCache.get` checks
  the embedded index **first**. `cache` is in scope at `page_setup.go:1381` (built `:1151`).
- **There is no reverse accessor.** `embeddedFaceName` is minted at `render.go:1204` and
  `embedded_face.go:229` and never inverted. `embeddedFaceSource.assetKey`
  (`embedded_face.go:90-93`) is an unexported field. **A reverse accessor must be added beside the
  forward mint** — not spelled at the call site, and never in TypeScript.

### The DW-92 arm — locate it so the plan can prove it is untouched

- `folio-go/page_setup.go:1229-1231`:
  ```go
  var unsupported *template.UnsupportedFontMediaTypeError
  if !errors.As(err, &unsupported) {
      return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
  }
  ```
  guarding `shapeSegments` at `:1208`, justification `:1209-1228`. Characterization test:
  `folio-go/canvas_embedded_face_test.go:307` `TestCanvasStillAbortsOnAnUnreadableCarriedFace`
  (doc `:273-306`), which records the behaviour and **explicitly declines to ratify it**.
- **No task in this plan touches it.** Every task is strictly downstream of `shapeSegments`' error
  path: the work begins at the fragment append (`:1381`), which is only reached when shaping already
  succeeded. See Design Notes, "DW-92 is not a dependency".
- D-7.4.2 anchors, for the degrade idiom if a new failure arm is ever needed: `page_setup.go:1152-1154`
  and `:1243`.

### The wire — a two-sided change with no Go-side test that catches a one-sided edit

- `folio-go/canvas_projection_wire_test.go:47` `canvasProjectionWireKeys` (top level),
  `:76` `canvasFontChainWireKeys = []string{"entries", "name"}`,
  `:95` `canvasFontChainEntryWireKeys = []string{"assetKey", "face", "family", "style"}`.
  Tests at `:122` (Go) and `:243` (TS mirror). **Nothing descends into `components`, `textPaint`,
  `lines` or `fragments` — CONFIRMED.** A new fragment field reddens this file **not at all**, which
  is exactly why a fragment-level record must be added.
- `folio-designer/src/engine-protocol.ts:173` — the `components` type, with the fragment shape spelled
  inline as `{ text: string; x: number }`.
- `folio-designer/src/engine-protocol.ts:453` —
  `hasOnly(fragment, ['text', 'x']) && …`. `hasOnly` (`:252`) rejects any key not on the list.
- **MEASURED-FALSE — the handed claim that an unlisted key "blanks the canvas with no diagnostic"**
  (DW-35 and 8.4's Design Notes both say so). Traced: `isCanvasTextPaint` → `isCanvasComponent`
  (`:325`) → `isCanvas` (`:269`) → `isSnapshot` (`:457`) → `parseInbound` (`:479`) returns `undefined`
  → `engine-client.ts:86-87` raises `PROTOCOL_INVALID`, which **terminates the worker**, rejects the
  ready promise and rejects every pending request. The session is **dead for good**, loudly — not a
  blank canvas. That is worse, not milder, and it makes the two-sided edit a hard ordering constraint.
- `folio-go/asset_bytes.go:19` — `func AssetBytes(t *Template, key string) ([]byte, string, error)`.
  **Media-type agnostic; serves font bytes as-is with no engine or protocol change.**
- `folio-go/wasm/engine.go:186-195` `Engine.AssetBytes`; note `raw, _, err := folio.AssetBytes(...)` at
  `:190` — **the media type is discarded at the wasm boundary.** JS dispatch:
  `wasm/cmd/engine/main.go:154-163`, `case "asset":`.

### The browser — how a fragment is painted today

- `folio-designer/src/App.tsx:1366-1369` `TextPaint`. The fragment span sets **one** inline custom
  property: `style={{ '--text-fragment-x': … }}`. Font-ish inline props (`--text-font-size`,
  `--text-font-weight`, `--text-font-style`, `--text-ink`) live one level up on `.canvas-text-paint`.
  **No family, ever.** Mounted from `CanvasComponent` (`:1300`) **and** `ComponentEcho` (`:1430`).
- `folio-designer/src/App.css:118` —
  `.canvas-text-fragment { left: var(--text-fragment-x); top: 0; font-family: 'IBM Plex Sans', 'IBM Plex Sans Thai', 'IBM Plex Mono', sans-serif; }`
- **MEASURED-FALSE — AC3's parenthetical.** `tokens.css:11`'s `--font-page` *is* Thai-first, but it
  **never reaches canvas text**: it feeds `--type-page-{title,body,fine}`, of which only
  `--type-page-body` is used, at `App.css:144` on `.file-message` (chrome). The canvas surfaces use
  `--type-page-mono`. `.canvas-text-fragment` uses a **hardcoded Latin-first literal** and **no
  `var()` at all**. So the hazard AC3 names does not exist on this path today. **AC3's requirement
  still stands and is still implementable** — see Design Notes, "AC3's false parenthetical".
- **The projection already carries font information — the browser already has every asset key.**
  `engine-protocol.ts:160-171`: `fontFamilies`, `defaultFontSize`, and
  `fontChains: …{ name, entries: …{ face, assetKey, family, style } }`. `FontChainEditor.tsx:43,:124`
  treat a non-empty `assetKey` as "embedded". **What is missing is not the bytes' address — it is
  which face a given FRAGMENT was resolved to.**
- `folio-designer/scripts/build-wasm.mjs:79` — the three build-time `@font-face` rules, written into
  the gitignored `src/generated/runtime-fonts.css`, imported at `tokens.css:1`. Families are
  `IBM Plex Sans` / `IBM Plex Mono` / `IBM Plex Sans Thai`; **`'IBM Plex Mono'` is actually Noto Sans
  SC** (`build-wasm.mjs:69`) — a CJK face under a mono name, a live trap for any widened engine↔browser
  tie.
- **No runtime font registration exists anywhere in the designer — MEASURED-TRUE.** No `new FontFace`,
  no `document.fonts.add`, no `insertRule`, no `CSSStyleSheet`/`adoptedStyleSheets`, no
  `document.createElement('style')`. The only occurrences of those spellings are detector patterns and
  string fixtures inside `canvas-font-stack.test.ts`.

### The asset-bytes precedent, and why its lifetime is wrong

- `folio-designer/src/App.tsx:1315-1352` `ImagePaint`; the effect is **`:1326-1340`** (not `:1341`),
  the image-specific consumer is **`:1335`**:
  `created = URL.createObjectURL(new Blob([result.bytes], { type: image.mediaType }))`.
  Deps `[engine, image?.assetKey, image?.mediaType, generation]`.
- **Lifetime is per mounted instance, and worse than DW-35 states.** `ImagePaint` mounts once per
  `CanvasComponent` **and** once per `ComponentEcho` (`:1430`, rendered per repeated sheet, `:775`),
  so N components × M sheets ⇒ N×M independent requests and N×M object URLs. No cache, no dedup.
- **Why that shape is fatal for fonts:** `createObjectURL` yields a *scoped handle*; `document.fonts`
  is a **global, name-keyed registry**. Copying the pattern gives duplicate adds under one family,
  one unmounting instance deleting the face another is still painting with, and an unmount-after-remount
  race that deletes the live face. Registration must be **document-scoped**.
- Call chain, React → wasm, all reusable unchanged: `App.tsx:1332` `engine.request('asset', …)` →
  `component-asset-command.ts:40-42` `assetBytesRequest` → `engine-client.ts:52-72` →
  `engine-protocol.ts:3,:465,:467` (`'asset'` is already allow-listed and already `needsPayload`) →
  `wasm/cmd/engine/main.go:154-163` → `wasm/engine.go:186-195` → `asset_bytes.go:19`.
- **Where a document-scoped effect belongs:** `App` (`App.tsx:83`) is the only document-level owner.
  `canvas` (and thus `canvas.fontChains`) is in scope at `:170`. `setCurrentSnapshot` (`:569`) is the
  sole document-replacement funnel; `documentGeneration.current++` fires there only when
  `clearDocumentInteraction` is true — at `:610` (open file), `:655` (new template), `:669`
  (undo/redo). **`documentGenerationValue` alone is an INSUFFICIENT dependency**: an ordinary font-chain
  command commits through `setCurrentSnapshot` *without* it (`:359,:406,:431,:476,:496,:524,:557`) yet
  can add or remove a carried entry. `FontChainEditor.tsx:100` already computes the correct stable
  listing key and is the pattern to copy.

### The guards — every anchor in the dispatch and the epic is stale; these are measured

`folio-designer/src/canvas-font-stack.test.ts` is **237 lines** at `9e77c36` (it was 189 at `1446b87`,
which is where the epic's `:100-106` / `:108-121` / `:123-132` come from). **All three epic anchors are
MEASURED-FALSE as anchors; two of the three assertions exist and one is misdescribed.** Corrected:

| test (line) | name | asserts | fate under 8.4a |
|---|---|---|---|
| `:72` | reads a non-empty declaration set | `declared.length >= 3` | green (vacuity guard) |
| `:76` | reads a non-empty request list | `requested.length >= 3` | green; **throws** if `App.css:118` is reformatted onto several lines |
| **`:80-83`** | **asks only for families an `@font-face` actually declares** | `:82` `expect(undeclared).toEqual([])` | **GUARD 1 of AC4** — the "tie". Green only because it reads `App.css` text; an inline family escapes it entirely, which is the *weakening* AC4 forbids |
| **`:126-139`** | records the fragment stack is a stylesheet constant with no document input | `:137` `expect(declaration as string).not.toMatch(/var\(/)`; `:138` `>= 3` | **the "no `var(`" guard the epic anchors at `:100-106`.** Red iff the family is routed through a custom property **in `App.css`** |
| `:141-147` | engine face names and browser family names do not intersect | `:143` `engineFaces = ['Noto Sans','Noto Sans Thai','Noto Sans SC']`; `:145-146` empty intersection both ways | survives mechanically (reads only the generator); **becomes false in spirit** — the "third guard" D-8.4.6 names |
| **`:169-188`** | **records that the browser has no family for a face the document carries (DW-35, Story 8.4a)** | `:179` generator registers nothing; `:184` **no non-test `src/**` `.ts`/`.tsx` registers a face at runtime**; `:187` requested ⊆ declared | **THE primary red.** A *disclosure of absence*: `:149-151` pre-authorises its **deletion** — *"the assertion that will have to be deleted by 8.4a rather than merely edited"* |
| `:194-210` | detects each of the three runtime-registration mechanisms it claims to scan for | calls `registersAFaceAtRuntime` on 4 positive + 3 negative fixtures | **green, and must be KEPT** — it is already the detector's own mutation proof |
| **`:212-225`** | **records that no designer source names a chain entry in a font-family declaration** | `:224` `sources.filter(s => /font-?[fF]amily['"\]]?\s*:\s*[^,;}\n]*(?:fontChains\|\bentries\b\|chain\.)/.test(s))` is `[]` | **GUARD 2 of AC4.** An asset-key-derived family does not match it, so it stays green *by luck*; AC4 requires it become "permits **only** an asset-key-derived name" |
| `:230-236` | keeps the generic fallback last | `:234` last entry `=== 'sans-serif'`; `:235` all others start `'` | red if a `var()` or unquoted family is inserted into the `App.css` stack |

- **Every guard in this file is a RAW-TEXT scan.** None parses CSS, none inspects a computed value,
  none calls a production function (except `:194-210`, which calls the file's own local helper).
  `:184` and `:224` scan **raw file text with no comment stripping**, so a *comment* in production
  source containing `new FontFace` or `fontFamily: chain.entries[0]` reddens them — the prose tax this
  repo has been bitten by before.
- The DW-35 tripwire comment (8.4's Task 14) is the block at **`:85-125`**, naming 8.4a at `:104-105`
  and pinning the inherited decision at `:111-115` (*"a carried face's CSS family name derives from its
  ASSET KEY, never from the asset's `font.family`"*). A second disclosure block sits at `:149-168`.

### The dead scan (AC6) — MEASURED-TRUE, with the mechanism

- `folio-designer/src/canvas-authority-contract.test.ts:24` — the prohibition: `/\bdocument\.fonts\b/,`
- `:145` — inside `violations()`, applied to **every** file before **every** pattern:
  `const source = withoutApprovedLocalPointerInput(file, fs.readFileSync(file, 'utf8')).replace(/document\.fonts\b/g, 'fontReadinessOnly')`
  Global, unconditional, and it rewrites the whole `document.fonts` **prefix** — so
  `document.fonts.add(face)` reads `fontReadinessOnly.add(face)` and `:24` **cannot match anything.**
- **Three independent confirmations that it is dead:** (1) `fontReadinessOnly` appears **nowhere else
  in the repo** — it is a throwaway token, not a real identifier; (2) the mutation block at `:199-211`
  proves **eleven** prohibitions and `document.fonts` is **not among them**, so deleting `:24` entirely
  leaves the suite green; (3) `git log -L 143,149` shows the `.replace(…)` was appended in `7bfb076`
  ("finish Story 6.1 sample data discovery") — a drive-by unblock, not a designed exception.
- **`new FontFace` is not in `prohibited` at all**, so this file would not catch a runtime registration
  by any route even if `:24` worked.
- **What the rewrite legitimately permits:** exactly one real usage,
  `folio-designer/e2e/engine-worker.spec.ts:23` `await document.fonts.ready` (waits for build-time
  faces so in-flight font requests are not counted as offline failures). The other three occurrences
  are prose/fixtures inside `canvas-font-stack.test.ts` (`:50`, `:158`, `:197`), which are in the
  `tests` corpus. **No occurrence in production `src/**`.**
- The file already has two precedents for a *scoped* exception, both non-vacuous:
  `withoutApprovedLocalPointerInput` at `:217-229`, whose `App.tsx` branch asserts
  `expect(source).toMatch(seam)` at `:227` so the carve-out cannot outlive its reason. `withoutComments`
  (`:113-133`) exists and is used by `refusalViolations()` (`:135-141`) but **not** by `violations()`
  (`:143-149`).

### Other designer tests that constrain the shape

- `design-contract.test.ts:22` — exact name equality between `src/design-tokens.ts` and
  `DESIGN.md`. **Adding a named design token reddens it** unless `DESIGN.md` changes in the same diff.
  A custom property set only from JS is not in `designTokenSets`, so it is safe; a new `typography`
  token is not.
- `design-contract.test.ts:33` — `expect(css).toContain("@import './generated/runtime-fonts.css'")`.
- `canvas-authority-contract.test.ts:196` — `@media` list in `App.css` must equal exactly
  `['prefers-reduced-motion: reduce']`.
- `design-contract.test.ts:55-56` — `App.css` must contain no hex/`rgb()`/`hsl()` literal. **Asset keys
  are 64 hex characters**; writing one into `App.css` would trip this.
- `App.tsx:1417` `componentAccessibleName` flattens `fragments.map(f => f.text)` — a new fragment field
  must not disturb it.
- `engine-protocol.test.ts` fixtures around `:349-420` (esp. `:410`) and `App.test.tsx:788` spell the
  fragment shape and will need the new field.
- `engine-ownership-contract.test.ts` is the repo's **only AST-based** contract test
  (`ts.createSourceFile`) — the model to copy if a widened guard should stop taxing prose.

### Baseline, measured at `9e77c36` (not assumed)

- `go test -count=1 ./...` → **one** red: `TestCorpusMeetsP6ExerciseFloors` (`internal/text`).
- `go test -count=1 -tags=matrix ./...` → **two** reds: `TestShippedFacesReproduceFromUpstream`
  (`folio-go`, `fontgen: fontTools is not importable`, DW-86) and
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (`internal/text`).
- **No third red at baseline.** The two standing reds and nothing else.
- `shasum -a 256 fixtures/*/expected.pdf` → **23** lines.
- `npm run typecheck` clean; oxlint **exactly 4** `only-export-components` warnings
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1263,1270`); Vitest **35 files / 325 tests, all passing**;
  `test:e2e:compile` clean. **Vitest is 325/35 at HEAD, not the 323/35 Story 8.4's spec recorded** —
  use 325/35 as this story's floor.

## Tasks & Acceptance

**Execution:**

1. `folio-go/embedded_face.go` — add the **reverse** accessor beside `embeddedFaceName` (e.g.
   `embeddedFaceAssetKey(name string) (string, bool)`, or a method on `embeddedFaceIndex`), so the
   `"asset:"` prefix is read in exactly the place it is written. Rationale: `:84-86` already forbids a
   second spelling of the derivation, and Task 2 needs the inverse.
2. `folio-go/page_setup.go:116-121` + `:1381` — add **one** optional field to `CanvasTextFragment`
   carrying the **asset key** of the face the engine resolved this fragment to, empty/omitted for a
   shipped face (`json:"assetKey,omitempty"`). Populate it at `:1381` from `fragment.face` via
   `cache` (in scope, built `:1151`) and Task 1's accessor. **Carry the asset key, not the minted
   `asset:` name** — see Design Notes, "Why the asset key and not the face name". Rationale: AC1, AC2;
   this is the attribution the projection currently discards.
3. `folio-go/canvas_projection_wire_test.go` — add a **fragment-level** wire record (a
   `canvasTextFragmentWireKeys` set) to both the Go test (`:122`) and the TS mirror (`:243`), which
   today descend no deeper than the top level and the two font-chain levels. Rationale: without it a
   one-sided protocol edit reddens nothing in Go and kills the worker in the browser.
4. `folio-designer/src/engine-protocol.ts:173` + `:453` — add the field to the inline fragment type and
   to the `hasOnly` allow-list, **in the same commit as Task 2**, with validation matching the engine's
   asset-key shape. Rationale: `hasOnly` rejects unlisted keys → `PROTOCOL_INVALID` →
   `engine-client.ts:87` terminates the worker and the session is unrecoverable.
5. `folio-designer/src/` (new module, e.g. `embedded-face-family.ts`) — the **named asset-key → CSS
   family derivation**, one exported function, with its own unit test including a collision case (two
   distinct asset keys ⇒ two distinct families) and a shared-`font.family` case. Rationale: D-8.4.1,
   AC2; DW-35 names this module by role.
6. `folio-designer/src/App.tsx` — a **document-scoped** effect that, for every carried entry the
   projection declares, fetches bytes via the existing `engine.request('asset', assetBytesRequest(key))`
   and registers `new FontFace(familyFor(key), bytes)` through `document.fonts.add`, releasing
   superseded faces when the document is replaced. Key it on `[engine, documentGenerationValue,
   <chain-entry listing>]` — copy `FontChainEditor.tsx:100`'s listing key, because
   `documentGenerationValue` alone does not change on a font-chain command. **Not inside `ImagePaint`,
   not per component.** Rationale: AC1; the `ImagePaint` lifetime is measurably wrong (N×M instances).
7. `folio-designer/src/App.tsx:1366-1369` — set the derived family **inline on the fragment span**,
   alongside the existing `--text-fragment-x`, **only** when the fragment carries an asset key; leave
   `App.css:118`'s declared stack untouched as the fallback and as the shipped-face path. Rationale:
   AC1 and AC3 require per-fragment attribution, and a single CSS rule cannot vary per fragment.
8. `folio-designer/src/canvas-font-stack.test.ts:80-83` — **widen GUARD 1**: the tie moves from "every
   family the rule asks for is declared" to "the families the rule asks for are the ones the engine
   measured with", **covering the inline families Task 7 introduces**, not only `App.css` text.
   Scope the engine↔browser equality to the **carried** case and state why in the test (see Design
   Notes, "Cause one stays open"). Mutation proof: deleting the widened assertion must redden a named
   test. Rationale: AC4 — an inline family that escapes the `App.css`-only scan is the weakening this
   AC exists to prevent.
9. `folio-designer/src/canvas-font-stack.test.ts:212-225` — **widen GUARD 2**: replace "no source names
   a chain entry in a font-family position" with "a font-family position may name **only** an
   asset-key-derived family", so it still catches `fontFamily: chain.entries[0]` **and** now catches any
   other document-supplied spelling. Mutation proof: a fixture spelling `fontFamily: entry.family` must
   redden it, and deleting the rule must redden a named test. Rationale: AC4.
10. `folio-designer/src/canvas-font-stack.test.ts:169-188` — **delete the disclosure of absence** (its
    own comment at `:149-151` pre-authorises exactly this) and **replace it with its positive twin**: a
    test asserting that runtime registration happens in **exactly one** named seam and nowhere else,
    reusing the `registersAFaceAtRuntime` detector. Keep `:194-210` untouched. Rationale: AC4 — the
    absence assertion cannot survive the feature, but its *teeth* must, relocated onto the seam.
11. `folio-designer/src/canvas-font-stack.test.ts:126-139` and `:230-236` — leave the `App.css`
    assertions **intact and green** by keeping the derived family out of `App.css` (Task 7). If Task 7's
    spelling forces a change here, that is a signal the spelling is wrong, not that the guard is.
    Rationale: AC4 "widened, never weakened"; these two are strictly about the stylesheet constant.
12. `folio-designer/src/canvas-authority-contract.test.ts:24` + `:145` — **repair the dead scan**:
    narrow the `:145` rewrite to `document.fonts.ready` only (or scope the exception by file the way
    `withoutApprovedLocalPointerInput` already does at `:217-229`, with a non-vacuity `toMatch`), and
    strengthen `:24` so `document.fonts.add`/`.delete`/`.load`/`.check` are caught while
    `e2e/engine-worker.spec.ts:23`'s `document.fonts.ready` stays legal. Route `violations()` through
    the existing `withoutComments` (`:113`). Add `new FontFace` to `prohibited` with a scoped carve-out
    for Task 6's single seam. Rationale: AC6.
13. `folio-designer/src/canvas-authority-contract.test.ts:199-211` — add the **missing mutation proofs**
    to the existing block: `document.fonts.add(face)` must be caught, `await document.fonts.ready` must
    not, and `new FontFace` must be caught outside the approved seam. Rationale: AC6 — today **no
    assertion in the file fails if `:24` is deleted**, which is the defect being repaired; repairing it
    without a mutation proof reproduces it.
14. `folio-go/canvas_embedded_face_test.go` — **leave unmodified** and re-run it as the AD-17 pin. Add
    no browser-side measurement. Rationale: AC5; Story 8.4 pinned the measurement path deliberately.
15. `folio-designer/src/engine-protocol.test.ts` (~`:349-420`, `:410`) and `App.test.tsx` fragment
    fixtures — extend for the new field, including a fragment **without** it (the shipped-face path).
    Rationale: the optional field must be proved optional.
16. `folio-designer/src/App.tsx` — degrade gracefully when the `asset` request fails or returns no
    bytes: the fragment falls back to the declared stack and the canvas keeps painting. Rationale: I/O
    matrix; the failure must not reach `#fail`/`PROTOCOL_INVALID`.
17. `_bmad-output/implementation-artifacts/deferred-work.md` — record **DW-35 cause two as closed by
    this story and cause one as explicitly still OPEN**, with the ground (a design-system decision
    above a builder's authority, never ruled) and what would discharge it. Rationale: DW-87 — the
    register is the audit trail, and a half-closed entry read as closed is how cause one disappears.

**Acceptance Criteria:**

- Given a component whose chain resolves a rune to a carried face, when the canvas paints it, then the
  face's own bytes reach the browser through the existing `asset` operation, are registered as a
  `FontFace` under an **asset-key-derived** family, and **that fragment's** rule asks for that family —
  no fallback, no host-installed font.
- Given an embedded face and a shipped face that share a `font.family`, when both are registered, then
  they get **distinct** CSS family names, because the embedded one derives from its asset key and never
  from `font.family`.
- Given an element whose text mixes scripts across a chain of shipped and carried faces, when it is
  painted, then each fragment is attributed **independently**, and only carried-face fragments carry a
  derived family.
- Given the two guards named in AC4, when they are re-authored, then each asserts a **strictly stronger**
  claim than before, and **each is red-proved by deletion** — removing the widened assertion reddens a
  named test.
- Given `canvas-font-stack.test.ts:194-210` (the detector's own proof), when this story closes, then it
  is **unmodified and green**.
- Given the browser never measures text (AD-17), when this story is complete, then no metric, advance or
  line break is computed in the browser, `folio-go/canvas_embedded_face_test.go` is **unmodified and
  green**, and no layout awaits font readiness.
- Given `canvas-authority-contract.test.ts`, when it scans for prohibited `document.fonts` usage, then
  the scan is **measured to actually run**: `document.fonts.add` is caught, `document.fonts.ready` in
  `e2e/engine-worker.spec.ts:23` is not, and a mutation proof fails if the rule is deleted.
- Given a projection whose fragment carries an unrecognised key, when the designer parses it, then Go
  and TypeScript agree — asserted by a new **fragment-level** record in `canvas_projection_wire_test.go`
  that reddens on a one-sided edit.
- Given the `asset` request for a carried face fails, when the canvas paints, then the fragment falls
  back to the declared stack, the canvas still renders, and the worker is **not** terminated.
- Given the whole corpus, when this story closes, then **all 23** `expected.pdf` digests are unmoved,
  and the only reds are the **two** standing ones measured at `9e77c36`.

## Spec Change Log

### 2026-09-01 — Code Map re-anchored from `9e77c36` to `dfe5129` (builder, step-03 preflight)

The Code Map above was measured at `9e77c36`. This dispatch's baseline is **`dfe5129`**, three commits
later (`e25381a`, `5d705b6`, `dfe5129`). Measured: the drift is **entirely Go-side** — `git diff --stat
9e77c36..dfe5129` touches no `folio-designer/**` file at all, so **every TypeScript and CSS anchor in
the Code Map is still exact**, including all nine `canvas-font-stack.test.ts` rows, the
`canvas-authority-contract.test.ts` `:24`/`:145`/`:199-211` anchors, `App.tsx:1366-1369`, `App.css:118`
and `engine-protocol.ts:173`/`:453`. No task changes. Corrections, per D-8.4.13 (cite by what a thing
asserts, not by line):

| Code Map claim (@`9e77c36`) | Measured at `dfe5129` |
|---|---|
| `page_setup.go:1381` — the fragment append / discard site | **`:1419`**. Text unchanged, still the ONE append site, still `{Text: fragment.text, X: int64(x)}`. |
| `page_setup.go:116-121` `CanvasTextFragment`, two fields only | **Unchanged and still exactly two fields** (`Text`/`json:"text"`, `X`/`json:"x"`), struct at `:118`. |
| `page_setup.go:1229-1231` — the DW-92 arm, `var unsupported *template.UnsupportedFontMediaTypeError` / `errors.As(err, &unsupported)` | **GONE. MEASURED-FALSE.** `grep "errors.As(err, &unsupported)"` returns nothing. `5d705b6`/`e25381a` rewrote the arm by *attributability*: it is now **`:1267-1268`**, `var carried *template.CarriedFaceError` / `if !errors.As(err, &carried)`, justified `:1236-1266`. `shapeSegments` is now called at **`:1218`** (not `:1208`). |
| `canvas_embedded_face_test.go:307` `TestCanvasStillAbortsOnAnUnreadableCarriedFace` — the characterization test | **GONE. MEASURED-FALSE.** That name no longer exists. The file grew +319 lines; the unreadable-carried-face case is now `TestCanvasDegradesRatherThanAbortingOnAnUnreadableCarriedFace` (**`:360`**), and a genuine abort is pinned by `TestCanvasStillAbortsOnAHostFontSetFaceThatWillNotParse` (`:453`). |
| `render.go:1204` — the second `embeddedFaceName` mint | **`:1205`**. |
| `embedded_face.go:83-87` — the derivation and its one construction site | **Unchanged** (`:83` prefix const, `:85-86` doc, `:87` func). |

**None of this changes a task, an AC, or the intent contract, and the contract is untouched.** The
Design Notes' structural claim *"DW-92 is not a dependency"* **still holds and is strengthened**: every
task remains strictly downstream of `shapeSegments` succeeding (the append at `:1419` is reached only
then), and the arm's rewrite happened *without* this story. Two Verification manual checks re-anchor
accordingly: the arm to keep absent from the diff is **`page_setup.go:1267-1268`**, and the
characterization test to leave unmodified is **`canvas_embedded_face_test.go:360`**
(`TestCanvasDegradesRatherThanAbortingOnAnUnreadableCarriedFace`); `canvas_embedded_face_test.go` stays
unmodified in whole, as AC5 already requires.

**Baseline re-measured at `dfe5129`, not assumed:** `go test -count=1 -tags=matrix ./...` → exactly
**two** reds, `TestShippedFacesReproduceFromUpstream` (`folio-go`) and
`TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (`internal/text`), 12 packages `ok`. **No third
red.** `shasum -a 256 fixtures/*/expected.pdf` → **23** lines. Root `README.md` md5
`078d7d80d518d54af2fc04fb270d46b8`.


## Review Triage Log

### 2026-09-01 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 7: (high 2, medium 4, low 1)
- defer: 7: (high 0, medium 4, low 3)
- reject: 6: (high 0, medium 1, low 5)
- addressed_findings:
  - `[high]` `[patch]` **The repeated-sheet echo had zero carried-face coverage.** `ComponentEcho` forwarded `carriedFaces` to `TextPaint` but nothing observed it: I independently reproduced the mutation `carriedFaces={carriedFaces}` -> `NO_CARRIED_FACES` at the echo's `TextPaint` and the whole designer suite stayed green (37 files / 344 tests). Every content window after the first is an echo, so a multi-sheet carried-face document would have painted sheets 2..N on the fallback stack at the engine's x-positions — this story's own defect — and shipped green. Fixed: new `carriedFaceEchoCanvas` fixture (3 content windows, one component spanning the seam) and test `paints a repeated sheet's echo with the carried face, not only the component's home`, asserting on `.canvas-component-echo .canvas-text-fragment`. **Re-verified by me after the patch: the same mutation now reddens exactly that one test (1 failed / 349 passed).**
  - `[high]` `[patch]` **The seam carve-out waived every prohibition, not the two font ones.** `withoutApprovedRuntimeFaceRegistration` did `readiness.replace(seam, '')`, deleting the entire `registerCarriedFaces` body from the scanned source, so `getComputedStyle`, `offsetWidth`, `ResizeObserver`, `devicePixelRatio` and the pagination-from-paint rules were all unenforced inside the one function permitted to touch fonts — an AD-17 hole, and a weakening-by-breadth in the very file this story was repairing. Its sibling carve-out for `canvas-font-stack.test.ts` already did it correctly (narrowed to two spellings, and says so). Fixed: the replacement now rewrites only `new FontFace` and `document.fonts` **inside** the matched seam; the non-vacuity `toMatch` is kept; the comment now states what is actually waived; and a new test `keeps every non-font prohibition live inside the approved seam` proves six non-font prohibitions are still caught there. Red-proved by restoring `replace(seam, '')`.
  - `[medium]` `[patch]` **The dependency the spec called out as load-bearing was unpinned.** The registration effect is keyed `[engine, documentGenerationValue, carriedFaceListing]` and the Code Map explicitly warned `documentGenerationValue` alone is insufficient. I reproduced: reducing the array to `[engine, documentGenerationValue]` left the suite green (344 tests). Fixed: `installStubFontSet` now records added/removed families, and a new test drives the author-reachable *Remove entry* control and asserts the face is released though the document was never replaced. Red-proved by dropping the listing from the deps.
  - `[medium]` `[patch]` **`page_setup.go` discarded the `ok` and could put a non-key on the wire.** `embeddedFaceAssetKey` was `strings.CutPrefix(...)`, which returns **the whole input name** with `false` on a miss; the append site wrote `carried, _ := cache.carriedAssetKey(...)`. Safe by construction today, but a face name reaching the wire meets the browser's `/^[a-f0-9]{64}$/` guard as `PROTOCOL_INVALID` — worker termination, the sharpest failure mode in this story. Fixed: the accessor now returns `"", false` on a miss, with `TestANonMintedFaceNameYieldsNoAssetKey` over seven non-minted names, and the discarded bool at the call site carries its justification.
  - `[medium]` `[patch]` **The chain-entry `assetKey` was unvalidated at the one place it becomes a CSS family.** `isFontChainEntry` only length-bounded it, while the fragment guard and the image guard both enforce 64-hex — yet the *chain-entry* key is what reaches `embeddedFaceFamily()` -> `new FontFace()` -> the inline `font-family`, so the module's "CSS `<custom-ident>` by construction" promise was unenforced at its only production caller (a string-injection path into an inline style). Fixed at the derivation rather than the protocol guard so it stays a degrade: `isCarriedFaceAssetKey` now filters `carriedFaceKeys`, with unit tests over twelve malformed shapes (including `..., serif` and `..."; }`) and an App test proving a malformed key yields no request, no family, no injected string, no alert, session alive.
  - `[medium]` `[patch]` **The degrade test did not discriminate.** It asserted `fontFamily === ['','']` on the line after `waitFor(request called)`, before any promise could settle — the identical assertion passes in the success case at that instant, so it proved "not yet registered", not "degraded". Fixed: two carried faces, the fetchable one's bytes withheld until the unfetchable one has already rejected, so `added === [family(fetchable)]` is a positive condition that cannot hold until the failure was handled.
  - `[low]` `[patch]` **DW-35's `Owner:` was stale.** Recorded per-cause ownership and what discharges cause one. **The finding's premise as I routed it was wrong and the implementer corrected it with evidence** — see the note below.

**Rejected findings, enumerated with the ground each was refused on (DW-87).** A rejection is sound only when it refutes the specific claim at the cited location; a true fact about nearby code is not a refutation.

1. *"No sprint-status, epics or Delivery Log entry accompanies the change"* (blind-hunter, on the diff as a whole). **Refused:** `bmad-build-auto` never writes `sprint-status.yaml` or a Delivery Log; those are the story-closer's step in this project's pipeline, run after this workflow halts. The absence is the workflow boundary, not an omission in this diff.
2. *"`expect(positions).toEqual(['App.tsx: embeddedFaceFamily(fragment.assetKey)'])` is brittle — it reddens if the local is renamed or a legitimate second font-family position is added"* (blind-hunter, `canvas-font-stack.test.ts` GUARD 2). **Refused:** that exactness is the guard's designed strength, not a defect. AC4 requires the position census to permit *only* an asset-key-derived family; a legitimate second position (a font-picker preview, say) *should* force a ruling rather than pass silently. The brittleness is the mechanism by which the guard cannot be widened by accident.
3. *"`fragment.assetKey !== undefined &&` is redundant because `carriedFaces.has(undefined as never)` is already false"* (blind-hunter, `TextPaint`). **Refused:** true as a fact about the expression, but it has no consequence for any consumer — the branch is a hot-path guard that also documents the optional field's absence as the shipped-face statement. Cosmetic; removing it trades clarity for nothing.
4. *"`NO_CARRIED_FACES` is declared between two type declarations and belongs with the module constants"* (blind-hunter). **Refused:** cosmetic placement, no consequence for correctness or comprehension.
5. *"`withoutComments` is duplicated into `canvas-font-stack.test.ts`; a helper module under `folio-designer/test/` satisfies both rejected alternatives"* (blind-hunter). **Refused:** the comment at the duplication site already reasons through the two alternatives that would break (importing registers the suite twice; hoisting into `src/` enters the scanned corpus). The proposed third directory is a real option but a style preference with no correctness consequence, and adding a new top-level test directory is a larger change than the duplication it removes.
6. *"`embedded-face-family.test.ts`'s collision list is hardcoded and will not follow the generator"* (intent-alignment). **Refused at the cited location:** the anchored version of that exact claim already exists — GUARD 1 part (d) loops over `declared`, read live from `scripts/build-wasm.mjs`, and asserts the derived family equals none of them. The module test's hardcoded list is a redundant restatement, not the only cover, so the "Shared `font.family`" row does not rest on an unanchored list.

**A finding I routed that was wrong on the facts, corrected by the implementer with evidence.** I routed the blind-hunter's claim that DW-35 cause one *"is left with no owner, no successor DW entry, and no recorded request for the ruling it needs."* That is **false**: `_bmad-output/planning-artifacts/epics.md:3129` carries **Story 8.4b — The canvas can name the face the engine measured**, whose Covers line names DW-35 cause one explicitly, and `epic-7-8-decision-log.md:3277` **D-8.4.14** (committed at `dfe5129`, before this story's implementation commit) rules cause one to 8.4b *and* rules the register's own stated blocker false. The implementer declined to write the invented ownerlessness and instead corrected DW-35 to the truth: per-cause owners (cause two -> 8.4a, delivered; cause one -> Story 8.4b, open and unstarted), a dated correction of the "OPEN and UNRULED" line, and what discharges cause one. **Cause one remains OPEN, as this story requires.** Consequence for the orchestrator: this spec's own Design Notes assertion that cause one's fix is *"a design-system decision no ruling has ever made"* was already stale when the spec was committed — D-8.4.14 predates it. The frozen contract was left untouched; the correction lives in the register.

## Design Notes

**Why the asset key and not the minted face name (Task 2).** `embedded_face.go:84-86` states that a
caller spelling the `"asset:"` prefix is *"writing the derivation a second time"*. Putting
`asset:<hex>` on the wire would force TypeScript to either strip the prefix (a second spelling, in a
second language, that no Go test can pin) or use it whole as a CSS family (embedding an engine-internal
namespace in a stylesheet). Carrying the **asset key** keeps the prefix wholly inside Go, gives the
browser the exact string the existing `asset` operation already takes as its payload
(`component-asset-command.ts:40-42`), matches the key already on the wire at
`canvasFontChainEntryWireKeys` (`:95` — `assetKey`), and makes D-8.4.1's derivation the browser's
**only** family rule. Absence of the field is then a precise, self-describing statement: *this fragment
is a shipped face*.

**AC3's false parenthetical — the requirement stands, the hazard does not.** AC3 says the family list
must follow the engine's chain order *"not `tokens.css`'s `--font-page` — which puts Thai first and
would hand Latin text Noto Sans Thai's Latin glyphs"*. Measured: `--font-page` **never reaches canvas
text** (it feeds three type tokens, only one of which is used, on chrome). `.canvas-text-fragment` uses
a hardcoded Latin-first literal with no `var()` at all — deliberately pinned by `:137`. So the named
hazard is not live on this path. **This is not an intent gap:** the AC's *requirement* — that what the
browser asks for follows the engine's resolution rather than a stylesheet constant — is well defined and
is satisfied **over-satisfied** by per-fragment attribution, because the engine has already applied
chain order when it resolved the rune. Implementing AC3 literally produces correct code; only its
justifying clause is stale. Recorded here rather than halted on, and reported to the orchestrator so the
epic text can be corrected the way D-8.4.2 corrected 8.4's AC3.

**Cause one stays open, and this is the one judgement the gate made.** DW-35 has **two** causes. Cause
two (a carried face has *no* family in the browser at all) is what Story 8.4 created and what D-8.4.6
split into this story. Cause one (a chain of **shipped** faces: the engine measures `Noto Sans Thai`
while the browser asks `IBM Plex Sans` first) is older, and DW-35 records its fix as *"a design-system
decision … above a builder's authority"* — rename the generated `@font-face` families, rippling into
`tokens.css`, its three type tokens and `design-contract.test.ts`, **or** generate a face-name → CSS
family map. **No ruling has ever made that decision**; D-8.4.1 settled the *embedded* derivation only.

The selecting evidence is DW-35's own *"What 8.4a owns"* enumeration, which D-8.4.6 adopted by re-owning
the entry to this story: every item on it — per-fragment attribution, the TS type and `hasOnly` guard,
the fragment-level wire record, the **asset-key → CSS-family** module, document-scoped `FontFace`
registration, the fragment `font-family` head, the two re-authors, the authority-contract narrowing — is
about the **carried** face. Not one mentions a shipped mapping. So this story closes cause two and
leaves cause one exactly where it was: open, unruled, and now explicitly recorded as such (Task 17).

**The consequence for AC4, and it is the trap.** AC4 words the widened tie as *"the families the rule
asks for are the ones the engine measured with"*. Asserted **universally**, that is **false today and
after this story** — for a shipped face the rule asks `'IBM Plex Sans'` while the engine measured
`'Noto Sans'`, which `:141-147` records as deliberately disjoint. A builder who writes the universal
assertion will find it red and will be tempted to weaken it back. **Scope the widened tie to the carried
case**, where it is true and newly checkable, and say in the test why the shipped half is out of scope,
citing DW-35 cause one. That is strictly stronger than today's guard — which ties no family to any
engine face at all — and it is the strongest claim that is true. Note also the aliasing trap for anyone
tempted to build the shipped half opportunistically: **`'IBM Plex Mono'` is Noto Sans SC**
(`build-wasm.mjs:69`), not a mono face.

**The two guards AC4 names are not quite the pair D-8.4.6 names.** D-8.4.6's split note names
{no `var(`, chain-entry-in-font-family}; AC4 names {the declared-families *tie*, chain-entry-in-font-
family}. Measured, those are three distinct assertions (`:137`, `:82`/`:187`, `:224`), plus `:235` and
the `:169-188` disclosure. The plan therefore **enumerates every assertion in the file and gives each an
explicit fate** (Tasks 8–11): widen `:82`, widen `:224`, delete-and-replace `:169-188` under its own
written authority, keep `:137`/`:235` green by construction, keep `:194-210` untouched. That
over-satisfies both readings and cannot be the wrong call.

**Deleting `:169-188` is not weakening — but only because its teeth are relocated.** It is a *disclosure
of absence*, and `:149-151` pre-authorises its deletion in its own words. An absence assertion cannot
survive the thing whose absence it asserts. What must survive is its **scanning power**: Task 10 keeps
the `registersAFaceAtRuntime` detector and re-points it from *"nowhere registers a face"* to *"exactly
one named seam registers a face"*. Deleting the replacement must redden a named test; if it does not,
the guard was lost.

**The prose tax, and why the widened guards should not inherit it.** `:184` and `:224` scan **raw file
text** with no comment stripping. A production comment containing `new FontFace` or
`fontFamily: chain.entries[0]` reddens them today. Widening is the moment to fix that: strip comments
(or use the AST approach `engine-ownership-contract.test.ts` already demonstrates) and carry a **negative
case** proving legitimate prose is not caught, alongside the positive one. A scan that only ever reddens
has not been shown to discriminate.

**DW-92 is not a dependency — measured, not assumed.** DW-92 concerns
`page_setup.go:1229-1231`, the `errors.As`-scoped arm guarding `shapeSegments` at `:1208`. Every task in
this plan is strictly **downstream** of that call: the work begins at the fragment append (`:1381`),
reached only when shaping already succeeded, and continues into the wire, TypeScript and CSS. **No task
reads, moves or widens that arm.** Nor does the ruling's outcome change this story's behaviour: DW-92
governs whether a projection *exists* for an unreadable carried face, not how an existing projection is
painted. **No task depends on how DW-92 resolves.** `TestCanvasStillAbortsOnAnUnreadableCarriedFace`
(`canvas_embedded_face_test.go:307`) is a characterization test and must be left exactly as it is —
it records the behaviour and explicitly declines to ratify it.

**The protocol edit is the sharpest failure mode in this story.** DW-35 and 8.4's Design Notes both say
an unlisted fragment key *"blanks the canvas with no diagnostic"*. Measured, that is **wrong and
optimistic**: it raises `PROTOCOL_INVALID` at `engine-client.ts:87`, which terminates the worker, rejects
the ready promise and rejects every pending request — the session is dead until reload, no edit, save,
undo or preview possible. Tasks 2, 3 and 4 must therefore land **together**, and Task 3's fragment-level
wire record is what makes a one-sided edit visible in Go rather than only in a browser nobody runs.

**What a browser-behaviour story can and cannot prove here.** The designer's only cross-boundary witness,
`e2e/browser-native-roundtrip.spec.ts`, is **compile-only**: `npm run test:e2e:compile` is `tsc --noEmit`,
and `npm run test:e2e` (Playwright) appears in no workflow. Browser e2e is deferred by D-000.4. So Vitest
+ jsdom can prove the *derivation*, the *registration call*, the *fragment attribute*, the *guards* and
the *protocol shape* — but **nothing in this repository can execute a real font load, a real
`document.fonts.add`, or a rasterized glyph**. jsdom applies no stylesheet and implements no font
loading. Any claim that the canvas *visibly* paints with the carried face is unverifiable by the gates
this story runs; say so plainly at close rather than letting a compile pass read as a run.

## Verification

**Commands** (from the repo root unless stated; `-count=1` on **every** Go gate — D-7.9.5, a warm cache
is not an anchor):

- **Before the first edit:** `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` —
  expected **23** lines. The baseline red set is **already measured at `9e77c36`** and recorded in the
  Code Map; re-confirm rather than re-derive, and never assume.
- `cd folio-go && go test -count=1 ./...` — expected: exactly one red,
  `TestCorpusMeetsP6ExerciseFloors`.
- `cd folio-go && go test -count=1 -tags=matrix ./...` — the **full** sweep, not a subset. Expected:
  exactly two reds, `TestShippedFacesReproduceFromUpstream` and
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`.
- `cd folio-go && go vet -tags=matrix ./...` — expected clean.
- `gofmt -l folio-go` — expected empty (`lint/…/licencegraph_test.go` is a known pre-existing offender,
  DW-23, outside this path).
- `cd folio-go && for T in darwin/arm64 linux/amd64 linux/arm64 js/wasm; do FOLIO_MATRIX_TARGET=$T go test -count=1 -tags=matrix -run TestTargetRenderHash -v .; done`
  — **all four AD-21 targets.** `TestTargetRenderHash` **asserts nothing unless `FOLIO_MATRIX_TARGET`
  is set**, so also run the **unset control** and report it as the deliberate no-op it is. Report all
  five legs with timings — the timing contrast is the evidence the legs actually asserted.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .`
- `cd lint && go test -count=1 ./...` — **`-count=1` is mandatory**: the rules package walks the
  `folio-go` tree with `ReadDir`, which Go's test cache does not track, so a cached `ok` is no
  measurement at all.
- `cd folio-go && GOOS=js GOARCH=wasm go test -count=1 -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./wasm/cmd/engine/`
  — this package is invisible to `go test ./...` (`//go:build js && wasm`), so "it compiles" is not
  evidence.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` —
  typecheck clean; oxlint **exactly 4** pre-existing `only-export-components` warnings
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1263,1270`); Vitest at or above the **325 tests / 35 files**
  baseline measured at `9e77c36` (report actual numbers); e2e **compiles only and executes nowhere by
  design**.
- `shasum -a 256 fixtures/*/expected.pdf | diff - <scratch>/digests.before` — **all 23 digests must be
  byte-identical.** This story changes browser rasterization; **any** movement is a defect, not a
  re-record.

**Mutation proofs required** (each states its anchor before editing, and reverts after):
- The widened GUARD 1 (`:80-83`) and GUARD 2 (`:212-225`): **deleting each widened assertion must redden
  a named test.** A guard that survives its own deletion was lost, not widened.
- GUARD 2's discrimination, **both directions**: a fixture spelling a document-supplied family must
  redden it, and legitimate prose in a comment must **not**.
- Task 10's replacement: removing the registration seam's allow-list entry must redden it, and adding a
  second registration site must redden it.
- The repaired `document.fonts` scan: `document.fonts.add(face)` **caught**, `await document.fonts.ready`
  **not caught**, and deleting the `:24` rule must redden a named test (today it does not — that is the
  defect).
- The fragment attribution: dropping the asset key at `page_setup.go:1381` must redden a named Go test,
  and the new fragment-level wire record must redden on a one-sided Go/TS edit.
- The derivation module: two distinct asset keys must yield two distinct families; an embedded and a
  shipped face sharing `font.family` must not collide.

**There are exactly TWO standing reds, measured at `9e77c36`. Neither may be hidden, skipped or "fixed".**
1. `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` — got 7, need ≥20, with its drift twin
   `TestCorpusP6StatsMatchDeclaredBaseline` (`internal/text/corpus_test.go:169,243`; `baselineP6g = 7`).
2. `TestShippedFacesReproduceFromUpstream` — under `-tags=matrix` **only**, failing with
   `fontgen: fontTools is not importable`. **DW-86 (D-8.3.4).** It **fails rather than skips
   deliberately** — *"the sources were not present"* must never read as *"the faces reproduce"*.

**Expect ONE transient red while the feature lands and before the guards are re-authored:**
`canvas-font-stack.test.ts:184` (no designer source registers a face at runtime). **That red is the
feature's own red-proof.** Never clear it by weakening the detector; clear it only via Task 10.

**Any third standing red is a real failure. Report it; do not absorb it.**

**Manual checks:**
- `git status --porcelain` before committing — only this story's paths. Stage **explicit paths only**;
  never `git add -A` / `git add .`.
- Root `README.md` untouched (md5 `078d7d80d518d54af2fc04fb270d46b8`).
- None of `fixtures/statement-signoff.json`, `fixtures/thai-stacked-marks/signoff.json`,
  `fixtures/embedded-font/signoff.json` appears in the diff. **The `embedded-font` record is a
  TRANSFERRED reading whose transfer LAPSES if `fixtures/thai-stacked-marks/expected.pdf` is ever
  re-recorded** (D-8.4.8) — this story must move neither.
- `folio-go/page_setup.go:1229-1231` (the DW-92 arm) is **absent from the diff**, and
  `TestCanvasStillAbortsOnAnUnreadableCarriedFace` is unmodified.
- `folio-go/canvas_embedded_face_test.go` is unmodified (the AD-17 measurement pin).
- Every commit message ends with the trailer
  `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`.
- **`## Review Triage Log`: every REJECTED finding is enumerated the way `addressed_findings` records a
  patch** — the claim, its cited location, and the ground on which it was refused. A bare `reject: N`
  cannot be audited; that is **DW-87**.

## Auto Run Result

### 2026-09-01 — Implement / review / commit dispatch (terminal)

Status: done
Blocking condition: none

**Implemented change.** The canvas now paints a carried face with that face. The engine attributes each
paint fragment to the asset it resolved that fragment to (`CanvasTextFragment.AssetKey`,
`json:"assetKey,omitempty"`, populated at the single fragment-append site from the chain-scoped
`fontCache`), the browser derives a CSS family from that asset key alone (`folio-carried-<key>` — a
namespace of the browser's own, never the engine's `asset:` prefix and never `font.family`), fetches the
bytes over the existing media-type-agnostic `asset` operation, registers one `FontFace` per key
**document-scoped** (not the `ImagePaint` N x M per-instance lifetime), and asks for that family inline on
that fragment only — and only once the face has actually registered, so a fetch failure degrades onto
`App.css`'s declared stack instead of onto the browser default. The Go and TypeScript sides of the
protocol changed in one commit, as the hard ordering constraint requires, with a new fragment-level wire
record (`canvasTextFragmentWireKeys`) tying `hasOnly(fragment, [...])` to the Go struct so a one-sided
edit reddens in Go rather than killing the worker in a browser nobody runs.

**Files changed.**
- `folio-go/embedded_face.go` — `embeddedFaceAssetKey`, the inverse of `embeddedFaceName`, beside it; the `asset:` prefix is read only where it is written, in Go and in no TypeScript.
- `folio-go/render.go` — `fontCache.carriedAssetKey`: embedded index first, prefix second, so a FontSet face that merely looks minted is not reported as carried.
- `folio-go/page_setup.go` — the optional `AssetKey` field and its population at the one append site.
- `folio-go/canvas_fragment_attribution_test.go` (new) — carried, shipped, mixed-script, non-minted-name and name/key round-trip, with vacuity preconditions.
- `folio-go/canvas_projection_wire_test.go` — the fragment-level wire record, both halves.
- `folio-designer/src/engine-protocol.ts` — fragment type, `hasOnly` allow-list, 64-hex validation.
- `folio-designer/src/embedded-face-family.ts` (new) — the one asset-key -> CSS-family derivation, plus `isCarriedFaceAssetKey`.
- `folio-designer/src/embedded-face-registry.ts` (new) — the only runtime registration seam; document-scoped, returns its own release.
- `folio-designer/src/App.tsx` — the document-scoped effect keyed `[engine, documentGenerationValue, carriedFaceListing]`; `carriedFaces` threaded to `CanvasComponent`, `ComponentEcho` and `TextPaint`.
- `folio-designer/src/canvas-font-stack.test.ts` — GUARD 1 widened (stylesheet half kept, plus a rendered-DOM tie, the seam's shared derivation, and a build-family collision check), GUARD 2 inverted to permit only an asset-key-derived family, the disclosure of absence deleted under its own pre-authorisation and replaced by its positive twin; `:194-210` untouched.
- `folio-designer/src/canvas-authority-contract.test.ts` — the dead `document.fonts` scan repaired and mutation-proved, `new FontFace` added, both carve-outs narrowed to spellings.
- `folio-designer/src/{embedded-face-family,embedded-face-registry}.test.ts` (new), `App.test.tsx`, `engine-protocol.test.ts` — coverage including the echo path, the chain-edit release, and the degrade path.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-35 cause two closed, cause one recorded OPEN and owned by Story 8.4b per D-8.4.14.

**Review findings breakdown.** 7 patched (2 high, 4 medium, 1 low), 7 deferred (4 medium, 3 low), 6 rejected — each rejection enumerated with its claim, cited location and refusing ground in the Review Triage Log above (DW-87). No intent gap and no bad_spec: the spec had flagged both high findings' surfaces (the `ComponentEcho` mount, and `documentGenerationValue`'s insufficiency); the implementation honoured both in code but left neither pinned, which is a patch, not a spec defect.

**Follow-up review recommendation: true.** Patched severities were high 2, medium 4, low 1; any high sets it true (the medium/low score is 3 x 4 + 1 = 13, also >= 5).

**Verification performed — measured at `dfe5129` before and re-run by the parent after the patches, not taken from a subagent's report.**
- `go test -count=1 ./...` — 16 `ok`, ONE red: `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (got 7, need >= 20).
- `go test -count=1 -tags=matrix ./...` — 13 `ok`, exactly TWO reds: `TestShippedFacesReproduceFromUpstream` (`fontgen: fontTools is not importable`, DW-86) and the P6g floor. **No third red.** Baseline at `dfe5129` measured independently first and was the same two.
- `go vet -tags=matrix ./...` exit 0, no output. `gofmt -l folio-go` empty.
- AD-21: darwin/arm64 PASS 0.70s, linux/amd64 PASS 6.36s, linux/arm64 PASS 4.93s, js/wasm PASS 11.12s; **unset control PASS 0.00s** — the deliberate no-op, and the timing contrast is the evidence the four legs asserted.
- `TestCrossTargetByteIdentity` PASS (24.3s). `cd lint && go test -count=1 ./...` 4 x `ok` (`-count=1` mandatory: the rules package walks the tree with `ReadDir`, which the test cache does not track).
- `GOOS=js GOARCH=wasm ... ./wasm/cmd/engine/` — `ok` 0.419s (invisible to `go test ./...`; run by hand).
- Designer: `typecheck` exit 0; oxlint exactly 4 pre-existing `only-export-components` warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1323,1330` — shifted from `1263,1270` by added comment lines, same two symbols); Vitest **37 files / 350 tests, all passing** (baseline 35/325, post-implementation 344); `test:e2e:compile` exit 0.
- **All 23 `fixtures/*/expected.pdf` digests byte-identical** to the pre-edit capture. Root `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8` unchanged and absent from the diff. No signoff record, fixture, `expected.pdf`/`expected.json`, `App.css`, `tokens.css`, `build-wasm.mjs` or `canvas_embedded_face_test.go` in the diff. The DW-92 successor arm is absent from the diff.
- Mutation proofs re-run by the parent: the echo mutation (`carriedFaces` -> `NO_CARRIED_FACES` at `ComponentEcho`'s `TextPaint`) left 344/344 green before the patch and now reddens exactly one named test (1 failed / 349 passed); the dep-array mutation left 344/344 green before the patch.

**Residual risks.**
- **Nothing in this repository can execute a real font load, a real `document.fonts.add`, or a rasterized glyph.** `test:e2e:compile` is `tsc --noEmit`, Playwright appears in no workflow, browser e2e is deferred by **D-000.4**, and jsdom applies no stylesheet and implements no font loading. The gates prove the derivation, the registration call, the fragment's rendered inline family, the guards and the protocol shape — **they do not prove the canvas visibly paints with the carried face.** The successor for this gap is named: when browser e2e arrives (D-000.4).
- Registration is proved against a hand-written `FontFace`/`document.fonts` stub, and the shipped-face fallback is proved as the *absence of an inline declaration* rather than as `.canvas-text-fragment`'s rule winning a cascade jsdom never runs.
- Seven deferred items are recorded in frontmatter, four of them medium: missing weight/style descriptors, no appended fallback list, no fan-out cap, and the Go half of the two-carried-faces matrix row (not patched deliberately — a second carried font fixture would create a new golden digest and a new human sign-off obligation).
- **The spec's own Design Notes claim that DW-35 cause one is "a design-system decision no ruling has ever made" was already stale when the spec was committed** — D-8.4.14 at `dfe5129` rules cause one to Story 8.4b and rules the register's stated blocker false. The frozen contract was left untouched; the correction lives in `deferred-work.md`. Cause one remains OPEN, as this story requires.

### 2026-09-01 — Plan-gate dispatch (superseded, kept for the record)

Status: ready-for-dev
Blocking condition: none

Dispatch: classic intent, `Halt after planning.` — the spec is written and **no code was written**.
Baseline measured at `9e77c36` (not assumed): `./...` one red, `-tags=matrix ./...` two reds — the two
standing reds and no third; 23 golden digests; Vitest 35 files / 325 tests passing; oxlint 4
pre-existing warnings; typecheck and e2e-compile clean.

Gate notes carried up to the orchestrator:
- **DW-92 is not a dependency.** No task reads, moves or widens `page_setup.go:1229-1231`; all work is
  downstream of `shapeSegments`' error path, and the ruling governs whether a projection exists, not
  how an existing one is painted.
- **AC3's parenthetical is measurably false** (`--font-page` never reaches canvas text). The AC's
  requirement stands and is implementable; only its justifying clause is stale.
- **Every guard anchor in the epic and the dispatch is stale** (the file grew 189 → 237 lines since
  `1446b87`). Corrected anchors are in the Code Map.
- **DW-35 cause one is left OPEN by design**, ruled at this gate on the register's own "What 8.4a owns"
  enumeration. This is the single judgement call the gate made.

## Delivery Log

### 2026-09-01 — done

Baseline `dfe5129`. Shipped in `c4cd60c` (implementation) and `51e38ac` (review patches); this closing
commit carries the record. **Every gate below was re-run by the closer at `9f04732` with `-count=1`;
none of it is the build's report relayed.**

**What shipped.** The engine attributes each canvas paint fragment to the document asset it resolved that
fragment's face to; the browser derives its own CSS family from that key alone, fetches the bytes over
the existing `asset` operation, registers one `FontFace` per key document-scoped, and sets that family
inline on exactly those fragments — and only once the face has actually registered, so a failed fetch
degrades onto the stylesheet's declared stack rather than off it. DW-35 **cause two closed**; **cause one
stays OPEN**, owned by Story 8.4b per **D-8.4.14**.

**Heavy-test cadence: every-story, and this project has no integration or e2e suite that executes.**
Measured and green: `go test -count=1 ./...` — 13 `ok`, one red; `-tags=matrix ./...` — 12 `ok`, exactly
two reds; `go vet -tags=matrix ./...` exit 0; `gofmt -l folio-go` empty; `lint` 4 × `ok`;
`GOOS=js GOARCH=wasm ./wasm/cmd/engine/` `ok` 0.475s (invisible to `go test ./...`, run by hand); AD-21
four legs PASS (darwin/arm64 0.74s, linux/amd64 6.98s, linux/arm64 6.30s, js/wasm 11.60s) plus the unset
control PASS 0.00s — the deliberate no-op, and the timing contrast is what shows the four legs asserted;
`TestCrossTargetByteIdentity` PASS 27.76s, all four targets on one hash; designer `typecheck` exit 0,
oxlint exactly 4 pre-existing `only-export-components` warnings, Vitest **37 files / 350 tests** all
passing, `test:e2e:compile` exit 0. **All 23 `expected.pdf` digests byte-identical** to a baseline
reconstructed **out of git** at `dfe5129`, not from any file this run wrote.

**Exactly two standing reds, no third.** `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (got 7,
need ≥ 20) and, under `-tags=matrix` only, `TestShippedFacesReproduceFromUpstream` (`fontgen: fontTools
is not importable`, **DW-86**, which fails rather than skips deliberately). Both suites were re-run after
the closer's own Go addition below and the red set was unchanged.

**The `followup_review_recommended` flag is cleared, on the closer's own measurements.** Both HIGH
patches were re-proved here rather than accepted:

- *The repeated-sheet echo.* Re-ran the build's mutation (`carriedFaces` → `NO_CARRIED_FACES` at the
  echo's `TextPaint`): **1 failed / 349 passed**, and the one red is
  `paints a repeated sheet's echo with the carried face, not only the component's home`. Read the test
  rather than counting it: it renders the real `App` over a three-content-window projection with a
  component spanning the seam, asserts non-vacuity (one home fragment, one echoed fragment), and asserts
  the **echoed node's own rendered `fontFamily`** equals the derived family. It is a rendered fact, not a
  proxy.
- *The authority-contract carve-out.* Verified the narrowing two ways. Restoring the broad
  `replace(seam, '')` reddens exactly `keeps every non-font prohibition live inside the approved seam`.
  And independently of that test, the seam replacement was reconstructed byte-wise: over the real seam
  body it produces **four changed regions, all inside `new FontFace` and `document.fonts`** — no other
  prohibition is silenced anywhere in the file. The readiness rewrite is scoped to
  `document.fonts.ready`, which can only ever shadow the `document.fonts` rule.

**Guard audit, all nine assertions against their stated fate (AC4).** The two that must not move —
the detector's own proof and *keeps the generic fallback last* — were checked **byte-identical** to
`dfe5129`, not merely present. `canvas_embedded_face_test.go` is absent from the diff entirely (AC5).
Teeth measured by mutating production, including the **weakening-by-evasion** case the AC exists for:

- Replacing the inline derived family with a bare `'IBM Plex Sans'` — a family `App.css` *does* declare,
  so the pre-8.4a stylesheet-only tie would have stayed green — reddens **both** widened guards.
- Deleting the inline family altogether reddens both widened guards and four App tests.
- Adding a second registration site to a production module reddens the Task 10 replacement **and** the
  repaired `document.fonts` prohibition **on the real corpus**. The build had only proved that rule
  against synthetic fixtures; this closes it against the files it actually scans.
- Go: dropping the asset key at the fragment append reddens three named tests
  (`TestACarriedFragmentIsAttributedToItsAssetKey`,
  `TestAMixedScriptElementIsAttributedFragmentByFragment`,
  `TestCanvasProjectionWireKeysAreTheRecordedSet`).

**A deferral disagreed with, and closed here.** The build deferred the I/O matrix row *"two embedded
entries in one chain"* at the Go attribution surface on the ground that a second carried font fixture
*"would create a new golden digest and a new human sign-off obligation"*. **Measured, that ground is
false for the surface the deferral names.** The carried-face document is *built in Go from bytes the
module already commits*, not read from `fixtures/`, and the attribution tests project it with
`CanvasWithTextPaint` — no page is rendered, so no digest and nothing to attest is produced. The closer
added `TestTwoCarriedFacesInOneChainAreAttributedToTheirOwnKeys`, which splices the shipped Latin face in
as a **second carried asset**, points the chain at two carried entries and no shipped one, and asserts
each fragment names its **own** key. **No new binary, no new fixture directory, no golden, no sign-off.**
Red-proved against the exact defect it names: attributing every carried fragment to the document's
first-indexed carried key leaves **every pre-existing test in the file green** and reddens only this one.
The row is now proved on all three surfaces and the deferral is struck from the frontmatter.

**Rejection audit (DW-87), all six spot-checked at their cited locations.** Five refusals are sound as
written (the workflow-boundary one, the GUARD 2 exactness one, the constant-placement one, the
`withoutComments` duplication one, and the collision-list one — GUARD 1 part (d) really does loop over
the families read live from the generator, so the module test's hardcoded list is redundant rather than
load-bearing). **One was refused on weaker ground than was available:** the claim that
`fragment.assetKey !== undefined &&` is redundant was refused as *cosmetic*. It is not cosmetic — the
set's `has` takes `string`, the field is `string | undefined`, and the project compiles `strict`.
Removing the clause was tried: **two `TS2345` errors**. The rejection reached the right outcome; its
stated ground understates why.

**Process record — commit provenance, instance two.** `c4cd60c` was created by the step-03 implementation
subagent, which step-03 does not authorize. The builder kept it (correctly scoped) and added `51e38ac`.
This is the **second occurrence in this run**; the first is at **D-8.4.9(c)**, recorded as a process
breach with the note that *re-measurement is a recovery, not a repeatable guarantee*. Audited here:
both commits contain **only this story's paths**, carry the required trailer, and neither touches root
`README.md` (md5 unchanged, `078d7d80d518d54af2fc04fb270d46b8`), any `signoff.json`, any fixture, any
`expected.pdf`/`expected.json`, `App.css`, `tokens.css`, `build-wasm.mjs` or `canvas_embedded_face_test.go`.
The recovery held again — and it is still a recovery.

**Frozen contract, verified rather than assumed.** The `<intent-contract>` slab is **byte-identical at
every revision of this spec** (`e25381a`, `c4cd60c`, `51e38ac`, worktree): 6004 bytes, md5
`e318a67c23df3e4cbb454c71144740c0`. Its Design Notes still claim cause one's fix is *"a design-system
decision no ruling has ever made"*, which **D-8.4.14 at `dfe5129` had already falsified before this spec
was committed**. That stale claim is **deliberately left standing** — a contract edited after the fact to
look correct is worth less than one showing what was believed when. The correction lives in
`deferred-work.md` under DW-35 and is restated here. Cause one remains **OPEN, owned by Story 8.4b**.

**The limit this story closes under, stated plainly.** Nothing in this repository can execute a real font
load, a real `document.fonts.add`, or a rasterized glyph. `test:e2e:compile` is `tsc --noEmit`; Playwright
appears in **no** workflow (confirmed — the string does not occur in `ci.yml`); browser e2e is deferred by
**D-000.4**; jsdom applies no stylesheet and implements no font loading. The gates above prove the
derivation, the registration **call**, the fragment's rendered inline family, the guards and the protocol
shape. **They do not prove the canvas visibly paints with the carried face.** This is a couldn't-look on
that one claim, not an all-clear, and the Design Notes carry the same disclosure with its named successor.

**Deferred work, now filed in the register rather than only in this file's frontmatter.** The build wrote
no `deferred-work.md` entry for any of its seven deferrals. Six survive and are registered as
**DW-94 … DW-99** (four medium: no weight/style descriptors on a registered carried face; the inline
family replacing rather than extending the declared stack; no concurrency cap or abort on the carried-face
fan-out; and — the two low ones — the comment stripper's blindness to regex literals and the test stubs
that install a font set by a spelling the repaired prohibition does not match, plus the stub teardown that
deletes rather than restores). Owners are named on each entry. The seventh is closed above.

**Two counts in the build's own report did not reproduce**, both benign and both recorded so the next
reader does not chase them: it reported `./...` as *16 ok* and `-tags=matrix ./...` as *13 ok*; measured
here they are **13** and **12** packages `ok` (the red sets and the red identities match exactly).

**Tracker.** `8-4a-the-canvas-paints-with-the-face-the-engine-measured: done`. `epic-8` deliberately left
`backlog` — the epic's heavy-test catch-up and its close are the orchestrator's. Remaining sequence
**8.4c → 8.4b → 8.5 → 8.6**, resequenced at `9f04732` per **D-8.4.17c**.
