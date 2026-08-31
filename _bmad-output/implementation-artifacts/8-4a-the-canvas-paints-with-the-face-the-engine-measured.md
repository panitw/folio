---
title: 'Story 8.4a — The canvas paints with the face the engine measured'
type: 'feature'
created: '2026-09-01'
status: 'in-progress'
baseline_revision: 'dfe5129ae89fcc124d96ed4047e3a7fe6db3348f'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized']
deferred: []
---

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
