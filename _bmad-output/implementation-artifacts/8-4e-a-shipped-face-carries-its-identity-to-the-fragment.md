---
title: 'Story 8.4e: A shipped face carries its identity to the fragment'
type: 'bugfix'
created: '2026-09-01'
status: 'done'
baseline_revision: 'd0f183b73940b1556b058ab3cbf5caf6428407e4'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: [oversized]
deferred:
  - summary: >-
      Closing DW-35 leaves the owed executed-browser font assertion recorded only in a closed
      entry's prose; DW-101 does not carry it.
    evidence: |-
      DW-35's Story 8.4e closing note defers the rasterization assertion to "the epic gate, behind
      CI wiring (DW-101)", but DW-101's own entry is about wiring CI and does not record the owed
      Epic 8 font assertion. Once DW-35 is closed, nothing in an OPEN entry names that obligation --
      the same ownerless-drift shape this epic already suffered once, to DW-35's own residual.
      Remedy touches DW-101, which this dispatch scopes to whoever wires CI.
    location: >-
      _bmad-output/implementation-artifacts/deferred-work.md (DW-35 closing note; DW-101 entry)
    severity: medium
  - summary: >-
      The shared 512 bound is measured in bytes on the Go side and UTF-16 code units on the
      TypeScript side, while comments on both sides call it "the same bound".
    evidence: |-
      folio-go/page_setup.go checks len(fragment.face) > maxCanvasPropertyString (bytes);
      engine-protocol.ts checks fragment.face.length <= MAX_CANVAS_PROPERTY_STRING (UTF-16 units).
      Go is the stricter of the two for any multi-byte name, so no wire breach follows and no
      current shipped face is near the bound -- but the claimed parity is false, and this is a
      pre-existing pattern shared with the chain entry's face rather than something 8.4e introduced.
    location: >-
      folio-go/page_setup.go:1528 and folio-designer/src/engine-protocol.ts
    severity: low
  - summary: >-
      DW-35's closure omits the "## DW-NN IS CLOSED" section marker that the register's four prior
      closed entries each carry.
    evidence: |-
      DW-24, DW-25, DW-29 and DW-36 each end with a "## DW-NN IS CLOSED -- Story X, date" marker
      delimiting the entry as it stood while open. DW-35's closure amended the heading, the Status
      bullet and appended a closing note (all of which AC4 required and which are present), but
      added no such marker. Adding one means deciding where a ~400-line entry ends, which is a
      mis-delimitation risk not worth taking unattended.
    location: >-
      _bmad-output/implementation-artifacts/deferred-work.md:2726
    severity: low
---

## Frontmatter warnings — what they mean here

**`oversized`.** Every folio story spec exceeds 1600 tokens. This one additionally carries the
enumerated fate of fifteen guard assertions in one test file and a re-measured Code Map, neither of
which the implementer can re-derive cheaply, and both of which this epic has already been burned by.

**`multiple-goals` is NOT carried, deliberately.** This is one mechanism — per-fragment face
attribution — extended from the carried population to the shipped one, across the one Go/TS seam that
must move in a single commit. The register amendments in Task 9 are not a second goal; AC4 requires
them.

## In plain terms (read this first if you just want the gist)

*Non-normative — a plain-language summary of the intent. The intent contract below governs
implementation.*

The canvas draws the page the way the engine will print it. To do that honestly it has to draw each
piece of text with the same typeface the engine measured that piece with. Since Story 8.4a it does
that for a face the document carries inside itself. For the three typefaces that ship with the
product it still does not: the engine says nothing about which one it used, and the browser is given
one fixed list of all three and takes whichever comes first that has the letter.

That is wrong whenever the document asks for them in a different order. All three of these typefaces
contain `A` and `5`, and they share hundreds of characters, so a document whose font list is just
Thai will have its English words measured in the Thai face by the engine and drawn in the Latin face
by the browser — right letters, wrong widths, so they creep out of position.

This story sends the typeface's name along with each piece of text and has the browser ask for that
one. The mechanism already exists; it gains a second kind of input.

One thing that will look like a gap and is not: nothing here can prove a letter was actually painted
in a given typeface. Nothing in this repository runs a browser as part of its checks. What is proved
is that the right name reaches the element and that the element asks for it. The rest waits on
continuous integration running the browser suite the project already has.

<intent-contract>

## Intent

**Problem:** A canvas text fragment drawn with a **shipped** face carries **no face identity on the
wire** — `CanvasTextFragment` is `{Text, X, AssetKey}` and `AssetKey` is populated only for a face the
document carries. The browser therefore falls through to `.canvas-text-fragment`'s **fixed stylesheet
constant**, which names all three shipped faces in one order regardless of what the document declared.
Measured at `f078b04` with the repository's own `.fontgen-venv` (fontTools 4.63.0) over
`folio-go/fonts/*`: the three faces map 3094 / 426 / 30890 code points, their pairwise cmap overlaps
are **339** (Sans×Thai) / **529** (Sans×SC) / **230** (Thai×SC), and **all three cover `A` (U+0041)
and `5` (U+0035)**. So for a chain of `["Noto Sans Thai"]` the engine measures Latin with Noto Sans
Thai while the Latin-first stack rasterizes it with Noto Sans — right glyphs, wrong advances. This is
the AD-17 violation Story 8.4b narrowed, surviving on the shipped-face arm, and it is DW-35's
attribution residual.

**Approach:** Extend the per-fragment attribution Story 8.4a built for carried faces to the shipped
population. The fragment carries the engine's **`FontSet` face name** — the engine's identity for a
shipped face, exactly as the asset key is its identity for a carried one (D-8.4.14: *"One rule for one
question"*) — on a new optional wire key beside `assetKey`. The browser derives a CSS family from that
name through one named seam and asks for it on the fragment, falling back to the existing declared
stack. Story 8.4b already declares an `@font-face` under each of those three names, so nothing new is
registered, no binary is added, and no mapping table is built.

## Boundaries & Constraints

**Always:**
- **The wire value is the engine's `FontSet` face name, verbatim** — `"Noto Sans"`, `"Noto Sans Thai"`,
  `"Noto Sans SC"` — never a derived, mapped or re-spelled name. D-8.4.14 rejects a mapping table by
  name, as *"a second authority maintained in lockstep with the shipped `FontSet`."*
- **`AssetKey` stays carried-only and stays empty for a shipped face.** The shipped name goes on a
  **new** key. `TestAShippedFaceCarriesNoAssetKey` must remain green **unmodified**.
- **The two keys are mutually exclusive, and the implementer must prove it rather than assume it.**
  Exactly one of the face-name key and `assetKey` is non-empty on every emitted fragment. Probe it over
  a shipped-only chain, a carried-only chain, and a mixed-script element before relying on it.
- **The Go field and its TypeScript validation land in ONE commit**, together with the
  `canvasTextFragmentWireKeys` record. An unlisted key fails `hasOnly` → `isCanvas` false →
  `PROTOCOL_INVALID` → **the worker terminates and the session is dead until reload** (D-8.4.13(b)).
  This is a hard ordering constraint, not a cosmetic one.
- **Reuse `maxCanvasPropertyString` / `MAX_CANVAS_PROPERTY_STRING` (512) for the new string.** It is
  already the bound `engine-protocol.ts` applies to a chain entry's `face`. Introduce **no new
  numeral**, so `engine-bounds-mirror.test.ts` needs no new row.
- **`.canvas-text-fragment` keeps its single-line rule beginning exactly `.canvas-text-fragment {`,
  keeps a `font-family:` declaration, keeps the three engine face names as quoted literals in
  `fonts.Shipped()`'s order, and keeps `sans-serif` last.** Three separate guards read that rule by
  parsing it, one of them at `describe` scope where a parse failure kills five tests. It survives as
  the fallback for an unattributed fragment.
- **The fallback tail has exactly ONE authority.** If the inline value and the stylesheet both spell
  the fallback list, a guard must read **both sources** and assert they agree — the tie idiom
  `engine-bounds-mirror.test.ts` and the declared/requested tie already use here.
- **Every guard this story widens is proved in BOTH directions**: the newly-legal form is accepted and
  the previously-illegal form still reddens. Every guard this story adds is mutation-proved by
  **deleting it** and confirming a named test goes red.
- Commit trailer: `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`.

**Block If:**
- A fragment is found carrying **neither** a face name nor an asset key, or **both** — the mutual
  exclusivity above is the design's premise and a counterexample invalidates the wire shape.
- Closing this requires editing a **chrome** type token (`--font-*` / `--type-*` in `tokens.css`) or
  any `DESIGN.md` statement. D-8.4.14 ruled the canvas fix must *"edit no chrome token at all"*; a
  chrome edit means the approach has drifted out of the story's ruling.
- Any golden PDF digest moves, or `shasum -a 256 fixtures/*/expected.pdf` returns other than **23**
  lines.
- A **third** distinct test failure appears beyond the two standing reds recorded in `## Verification`.
- Making an AC observable would require adding an assertion to the Playwright suite. D-8.4.25(d)+(e)
  place the executed browser assertion at the **epic gate**, behind CI wiring (DW-101), and require CI
  to execute the suite **before or in the same unit as** the assertion. Do not add an observer to a
  suite nothing runs.

**Never:**
- **Never touch `folio-go/fonts/` or the engine `FontSet`.** No face added, renamed or removed;
  `TestShippedFacesReproduceFromUpstream`'s `"3 of 3 faces"` stays **3**.
- **Never modify any `.folio` document, fixture, golden or attestation record.** The corpus is 23
  documents and none may move. A moved PDF digest is a **defect**, not a re-record. No agent writes
  `reader`, `date` or `examined` into any `signoff.json`.
- **Never modify, move, delete or stage root `README.md`** (md5 `078d7d80d518d54af2fc04fb270d46b8`).
- **Never build the size-budget metric or set a threshold** — Story 8.4d, D-8.4.24. Never build a
  canvas diagnostics channel — DW-93, ruled acceptable. Never wire CI — DW-101, not this story's.
- **Never soften a guard to keep it passing.** Specifically: never relax the census `toEqual` to
  `arrayContaining`, never drop the `requested.slice(0,3)` order tie, never weaken the engine-face
  membership assertions Story 8.4b delivered under D-8.4.14.
- **Never restructure `build-wasm.mjs`'s one-line `runtime-fonts.css` template literal**, and never
  add or remove an `@font-face` rule. This story adds no font and declares no family.
- **Never derive the browser family from a chain entry's `family` / `style` display fields.** AD-8 and
  D-8.4.1: resolution is by the engine's identity alone.
- Never push. Never create a branch. Never `git add -A` or `git add .` — explicit paths only.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Shipped face attributed | element whose chain resolves `"Noto Sans Thai"` | the fragment emits the face-name key with `"Noto Sans Thai"`; `assetKey` absent | No error expected |
| Latin through a Thai-first chain | chain `["Noto Sans Thai"]`, text `A5` | engine attributes the fragment to `"Noto Sans Thai"`; the rendered fragment's `style.fontFamily` names `Noto Sans Thai` **first** | No error expected |
| Carried face | chain naming a font asset | `assetKey` emitted as today, face-name key absent; the fragment's family is the asset-key-derived one | No error expected |
| Mixed-script element | Latin + Thai in one element through a multi-entry chain | each fragment attributed independently to the face the engine used for it | No error expected |
| Unattributed fragment | a fragment carrying neither key | falls back to the stylesheet's declared stack; the canvas still paints | No throw, no worker termination |
| Go emits the key, TS does not accept it | fragment key absent from `hasOnly`'s list | `isCanvas` false → `PROTOCOL_INVALID` → **worker terminated, session dead** | Build failure — the two sides must land together |
| Face name over the bound | a face name longer than 512 | refused by the same bound a chain entry's `face` already uses, on both sides | Projection refused with a stated reason |
| Face name the browser has no face for | attributed name not among the declared families | the inline value falls through its own tail to the declared stack and then `sans-serif` | No error expected; nothing blank |

</intent-contract>

## Code Map

**Every anchor below was re-measured at `f078b04`, this dispatch's HEAD.** Per **D-8.4.13** line
anchors rot across story boundaries and **assertions are named by what they assert**; the line numbers
are navigation aids only. If the build dispatch's HEAD differs from `f078b04`, run
`git diff --stat f078b04..<HEAD>` first and record the drift in `## Spec Change Log`.

### The Go seam

- `folio-go/page_setup.go` — **module root, `package folio`. There is no
  `folio-go/internal/template/page_setup.go`; do not look for one.**
  - `:116-149` `CanvasTextFragment` — three fields today: `Text` `json:"text"`, `X` `json:"x"`,
    `AssetKey` `json:"assetKey,omitempty"` (8.4a's addition). Its doc comment argues, at length,
    against putting the minted `"asset:"+key` name on the wire and against per-component attribution.
    **Read it before adding a field; the new field's comment must answer the same two questions.**
  - `:151-165` `CanvasTextLine{Top,Baseline,Advance,Width,Fragments}`.
  - `:1434-1464` the fragment build loop inside `addCanvasTextPaint`. `:1462` is
    `carried, _ := cache.carriedAssetKey(fragment.face)`; `:1463` is the append. **`fragment.face` is
    consumed only by that one call and then discarded — that discard is what this story closes.**
  - In scope at the append site and currently unused there: `chain` (from `fontChain(t, element)`,
    `:1240`) and the chain-scoped `cache` (`:1236`).
  - `:40` `maxCanvasPropertyString = 512` — the bound to reuse.
- `folio-go/render.go` — `:106` `textRunSource` (its `face string` is the only face information at the
  append site); `:1844` `positionSegments` fills `placed`; `:1859-1860` copies `face` verbatim from
  `faceSegment.face`; `:1798-1806` `shapeSegments`; `:1557-1577` `resolveRuneFace` **returns an element
  of `chain []string`**; `:1205-1213` `chainFaceNames` — the function that makes a shipped entry's
  `Face` string and a carried entry's minted name into one list. **Consequence: for a shipped face
  `fragment.face` IS the `FontSet` key verbatim**, because `fontCache.get` cannot resolve anything
  else. `:1345-1350` `fontCache.carriedAssetKey` — asks the index first, reads the prefix second.
- `folio-go/embedded_face.go:84-88` `embeddedFaceName` (the mint); `:114-120` `embeddedFaceAssetKey`
  (empties on miss, deliberately, because the value reaches the wire).
- `folio-go/fonts/fonts.go` — `Shipped()` returns a `folio.FontSet` map with exactly
  `"Noto Sans"`, `"Noto Sans Thai"`, `"Noto Sans SC"`. **It is a map, so it carries names, not order**;
  the order the guards use comes from the source text of the literal.

### The wire tripwire — it already descends to fragments

- `folio-go/canvas_projection_wire_test.go`
  - `:122` `canvasTextFragmentWireKeys = []string{"assetKey", "text", "x"}` — the **accepted** set, not
    the emitted one.
  - `:149` / `:181-197` pins the emission in **both** directions: a carried fragment must marshal to
    **exactly** `{assetKey,text,x}`; a shipped fragment to **exactly** `{text,x}`. **A new
    always-emitted key reddens both; a new `omitempty` key reddens the shipped assertion.**
  - `:339` the extraction regex `hasOnly\(fragment, \[(.*?)\]\)`; `:345` / `:391-399` reads
    `folio-designer/src/engine-protocol.ts` from disk and compares to the same list. **This is the
    cross-language gate — it `t.Fatal`s rather than skipping if the TS file is absent.**
- `folio-go/canvas_fragment_attribution_test.go` — 8.4a's attribution tests, including
  `TestAShippedFaceCarriesNoAssetKey` (`:89`), which asserts `fragment.AssetKey == ""` for a shipped
  face. **These read struct fields, so they are blind to a new field — extend them, do not rely on
  them to catch this story.**

### The TypeScript protocol

- `folio-designer/src/engine-protocol.ts`
  - `:252` `hasOnly` — **a SUBSET check**: every present key must be listed; a listed key may be absent.
    `:253` `hasExactKeys` is the two-directional sibling.
  - `:453` the fragment guard, inside `isTextPaint`: `hasOnly(fragment, ['text','x','assetKey'])` plus
    per-field checks, `assetKey` validated as `/^[a-f0-9]{64}$/`.
  - `:173` the declared `CanvasProjection` fragment type
    `{ text: string; x: number; assetKey?: string }`. **⚠ This type line is pinned by nothing** — the
    Go tripwire regex-extracts only the guard's key list, so type and guard can drift apart silently.
    Move both.
  - `:219-225` `isFontChainEntry` — **the discriminated-pair precedent to copy**:
    `hasExactKeys(value,['face','assetKey','family','style'])`, then
    `if ((face.length > 0) === (assetKey.length > 0)) return false`, with the 512 bound applied at
    `:223`. `:59` `MAX_CANVAS_PROPERTY_STRING = 512`.
- `folio-designer/src/engine-bounds-mirror.test.ts` — reads both `page_setup.go` and
  `engine-protocol.ts` and asserts the duplicated numerals agree. Reusing 512 means no new row.

### The designer paint half

- `folio-designer/src/App.tsx`
  - `:1426-1449` the 8.4a comment block, `:1450-1452` `TextPaint` (exported for testability). The
    fragment `<span className="canvas-text-fragment">`'s inline style is a **conditional spread** gated
    on `fragment.assetKey !== undefined && carriedFaces.has(fragment.assetKey)`. **A shipped fragment
    sets no `fontFamily` at all and falls to the stylesheet rule — that is the defect, in one
    expression.**
  - `:29` imports `embeddedFaceFamily`, `isCarriedFaceAssetKey`; `:159` `carriedFaces` state; `:219`
    derives it from `canvas.fontChains` entries; `:220-231` the document-scoped registration effect.
  - `:1360` and `:1514` are `TextPaint`'s two call sites — the live component **and the
    `ComponentEcho`** (repeated sheets). Any fix must hold at both; a one-window fixture proves nothing
    about the echo.
- `folio-designer/src/embedded-face-family.ts` — `:47` `embeddedFaceFamilyPrefix = 'folio-carried-'`,
  `:49` the 64-hex pattern, `:55-57` `embeddedFaceFamily`. **The structural model for the shipped-side
  seam**: one tiny module, one exported derivation, its own test file.
- `folio-designer/src/embedded-face-registry.ts:54-79` — the single runtime registration seam (`:71`
  `new FontFace(...)`). **Shipped faces need none of this**: 8.4b already declares them in the
  generated stylesheet.
- `folio-designer/src/App.css:118` — verbatim:
  `.canvas-text-fragment { left: var(--text-fragment-x); top: 0; font-family: 'Noto Sans', 'Noto Sans Thai', 'Noto Sans SC', sans-serif; }`
  The only `font-family` declaration in the file. `:106`/`:116`/`:117` are the sibling paint rules.

### `canvas-font-stack.test.ts` — 709 lines, 15 tests, and the story's main collision surface

Suite scope reads `build-wasm.mjs`, `App.css`, and `folio-go/fonts/fonts.go` at `:227-234`.

**Parsers, and what breaks them.** `declaredFamilies` (`:78-80`,
`/@font-face \{ font-family: '([^']+)'/g` over the generator) · `wellFormedRuleFamilies` (`:89-91`,
whole-rule) · **`requestedFamilies` (`:104-111`) — parses `App.css` by finding the line starting
exactly `.canvas-text-fragment {`, then `font-family:([^;]+);`, then `'([^']+)'`, and THROWS if either
is missing. It runs at `describe` scope, so a throw kills tests 4, 7, 8, 9 and 15** ·
`shippedFaceNames` (`:56-60`, parses `Shipped()` — **dynamic, not hardcoded**, per D-8.4.14) ·
`fontTokenValues` (`:100-102`) · `withoutComments` (`:143-163`, applied to the generator, `tokens.css`
and every designer source — **not** to `App.css`).

**The fifteen assertions, by what they assert, with the fate this story gives each.** *(KEEP = must
stay green untouched; WIDEN = must accept the new form and still reject the old; REPLACE = records the
old state and is expected to go red.)*

| # | Line | Asserts | Fate |
|---|---|---|---|
| 1 | 245 | `declared.length >= 6` (vacuity floor) | KEEP |
| 2 | 268 | exactly 6 well-formed rules; `declared` equals the well-formed set | KEEP |
| 3 | 279 | red-proof of #2 over `good`/`stray` fixtures | KEEP |
| 4 | 290 | `requested.length >= 3` (vacuity floor) | KEEP |
| 5 | 304 | a commented-out `@font-face` is not counted; the unstripped parse would have counted it | KEEP |
| 6 | 323 | each `--font-*` token names an IBM Plex family; **no token names an engine face** | KEEP — and it is the guard that reddens if this story drifts into chrome |
| 7 | 411 | `requested ⊆ declared`; three DOM assertions on painted fragment families; the seam/import ties; the derived family is not a declared family | **WIDEN** — its DOM half must now also cover a shipped fragment |
| 8 | **507** | **"the fragment stack is a stylesheet constant with no document input"** — effectively only `declaration` has no `var(` plus a re-floor of `requested.length >= 3` | **REPLACE** — see Design Note 3 |
| 9 | 540 | `engineFaces ⊆ declared`, `⊆ requested`, and **`requested.slice(0,3) toEqual engineFaces`** (the order tie) | KEEP — D-8.4.14's delivered guard; **never weaken** |
| 10 | 593 | runtime registration happens in exactly one named seam, `embedded-face-registry.ts`, and nowhere else | KEEP — this story registers nothing |
| 11 | 613 | red-proof of #10 (second site red; seam that stops registering red) | KEEP |
| 12 | 635 | the registration detector works, over 4 positive + 3 benign fixtures | KEEP |
| 13 | **666** | the census: `positions toEqual ['App.tsx: embeddedFaceFamily(fragment.assetKey)']` — **an exact `toEqual` over every `font-family`/`fontFamily` position in every non-test designer source**, with the approved value matched by `assetKeyDerivedFamily` (`:203`) | **WIDEN to a closed set of exactly two** |
| 14 | 678 | red-proof of #13 over 5 red + 6 green fixtures, incl. "prose describing one is left alone" | **WIDEN** — add a fixture per new approved form and per newly-rejected near-miss |
| 15 | 702 | the last stack entry is exactly `sans-serif`; all others are quoted | KEEP |

`fontFamilyDeclarations` (`:197-199`) is the census scanner, applied over `withoutComments(source)`;
`designerSources()` (`:166-170`) walks non-test `.ts`/`.tsx` under `src/` recursively.

### Where AC2 is observable

- `folio-designer/src/App.test.tsx` (2514 lines) — the richest per-fragment face assertions, all
  reading `node.style.fontFamily` off a `.canvas-text-fragment` node: `:858-860`, `:891-897` (home
  **and echo** fragments), `:949-952`, `:983-989` (injection guard), `:1024-1037` (unregister → both
  fragments fall back to `''`). **This is the layer AC2 must be written at.**

### Read-only evidence (do not modify)

- `folio-designer/src/font-binary-identity.test.ts` (1144 lines, 18 tests) — opens font bytes, pins the
  six-family→six-file map and each face's sha256 against its own `NOTICE.md`. **Blind to this story's
  change surface**; it reddens only if the generator's rule set moves, which it must not.
- `folio-designer/scripts/build-wasm.mjs` (110 lines) — `assets` at `:64-74` (nine slots), the one-line
  `runtime-fonts.css` literal at `:110` emitting **six** rules over **six distinct files**.
- `folio-designer/e2e/` — **12 executable Playwright specs**, a real `webServer`, and
  `"test:e2e": "playwright test"`. `.github/workflows/ci.yml:218` runs only `test:e2e:compile`
  (`tsc --noEmit`). **DW-101; not this story's to wire.**

## Tasks & Acceptance

**Execution:**

1. `folio-go/page_setup.go` — add one optional face-name field to `CanvasTextFragment`
   (`json:"...,omitempty"`), populated at the fragment append (`:1463`) from `fragment.face` when
   `cache.carriedAssetKey` reports the face is **not** carried, bounded by `maxCanvasPropertyString`.
   Write its doc comment to answer the two questions the `AssetKey` comment answers: **why the
   `FontSet` name and not something derived** (D-8.4.14: the engine's identity for a shipped face), and
   **why per fragment and never per component** (`faceSegment.face` is scalar; a fragment is exactly
   one face by construction). Rationale: the value is already in scope and discarded.
2. `folio-go/canvas_projection_wire_test.go` — extend `canvasTextFragmentWireKeys` and **both**
   direction-pinned assertions: a carried fragment emits its exact set, a shipped fragment emits its
   exact set including the new key. Rationale: this is the only thing in Go that would notice a
   one-sided edit, and a one-sided edit kills the worker.
3. `folio-go/canvas_fragment_attribution_test.go` — extend with the shipped-face attribution
   assertions, including a **mixed-script element attributed fragment by fragment** and the
   mutual-exclusivity probe the contract requires. Keep `TestAShippedFaceCarriesNoAssetKey` unmodified.
   Rationale: the existing tests read struct fields and cannot see a new one.
4. `folio-designer/src/engine-protocol.ts` — add the key to the fragment `hasOnly` list, add its
   per-field validation (string, non-empty, `<= MAX_CANVAS_PROPERTY_STRING`), add the mutual-exclusion
   check on the `isFontChainEntry` model, **and update the `CanvasProjection` fragment type at `:173`
   in the same edit**. Rationale: the type is pinned by nothing and will otherwise drift.
5. `folio-designer/src/` — add the shipped-side derivation seam (a small module beside
   `embedded-face-family.ts`, with its own test file) that turns an engine face name into a CSS
   `font-family` value: the attributed face **quoted and first**, then the fallback tail. It must
   reject a name that is not usable as a CSS family rather than interpolating it blind. Rationale: one
   named seam per derivation is this epic's established shape and is what the census guard can name.
6. `folio-designer/src/App.tsx` — in `TextPaint`, set the fragment's inline `fontFamily` from the new
   seam when the fragment carries a face name, leaving the carried branch exactly as it is. Rationale:
   one expression, two populations; it must hold at **both** `TextPaint` call sites, including
   `ComponentEcho`.
7. `folio-designer/src/canvas-font-stack.test.ts` — apply the fate table above: **replace** assertion 8
   with its positive twin, **widen** 7, 13 and 14, keep the rest green untouched. Relocate assertion
   8's teeth rather than deleting them (Design Note 3). Add the two-source tie for the fallback tail if
   the tail is spelled twice. Rationale: the story's own AC3.
8. `folio-designer/src/App.test.tsx` — add the rendered-DOM assertions for AC2, at both the home and
   echo fragments, mirroring the carried-face tests already there. Rationale: this is the only layer
   that executes.
9. `_bmad-output/implementation-artifacts/deferred-work.md` — **DW-35: close it.** Amend the heading to
   record `CLOSED by Story 8.4e`, set `**Status:** **CLOSED.**` in the file's own closure format, and
   append a closing note covering **both causes and the residual**. In the same edit correct two things
   the entry is now stale on: it still says the residual *"has no named owner … Escalated to the
   orchestrator"* (superseded by **D-8.4.26**), and it still cites the **void** trigger *"when browser
   e2e arrives (D-000.4)"* (replaced by **D-8.4.25(b)**: *"when CI executes the Playwright suite"*).
   **DW-102: re-own it** — amend its `Owner:` bullet to remove Story 8.4e and record the ground
   (Design Note 5). Rationale: AC4, and the ownerless-drift failure this epic has already had once.
10. Verify the whole `## Verification` cadence and record measured counts, not adjectives.

**Acceptance Criteria:**

- **AC1 — identity on the wire.** Given a component whose chain resolves a shipped face, when the
  canvas projection is built, then the fragment's marshalled JSON carries the engine's `FontSet` face
  name for **that** face and omits `assetKey`; and given a carried face, then it carries `assetKey`
  and omits the face name — both asserted as **exact key sets**, in both directions.
- **AC2 — the fragment asks for that face, observed where something executes.** Given a document whose
  chain is `["Noto Sans Thai"]` and Latin text drawn through it, when the canvas renders in the test
  environment, then the engine attributes the fragment to `"Noto Sans Thai"` **and** the rendered
  `.canvas-text-fragment` node's `style.fontFamily` names `Noto Sans Thai` **before** any other family
  — at the home fragment **and** at the `ComponentEcho` fragment. *(Phrased at the rendered-DOM layer
  deliberately: jsdom applies no stylesheet and loads no font, so "rasterizes with" is not observable
  here. See Design Note 6.)*
- **AC3 — the old record is replaced, not weakened.** Given the guard asserting *"the fragment stack is
  a stylesheet constant with no document input"*, when per-fragment attribution lands, then that
  assertion is **gone** — not softened — its detector is relocated onto the new seam, and deleting the
  replacement reddens a **named** test; and the census guard admits the new derivation as a **closed
  set** while a document-supplied family in a `font-family` position still reddens.
- **AC4 — DW-35 closes.** Given DW-35, when this story is complete, then its register entry records it
  **CLOSED** — both causes and the residual — in the file's own closure format, with its two stale
  passages corrected.
- **AC5 — nothing else moved.** Given the corpus, when the story is complete, then
  `shasum -a 256 fixtures/*/expected.pdf` is **23** lines with unchanged digests, the four AD-21 legs
  and `TestCrossTargetByteIdentity` pass, and the test suites show exactly the two standing reds
  recorded in `## Verification` and no third.

## Spec Change Log

**Build dispatch, HEAD `d0f183b73940b1556b058ab3cbf5caf6428407e4`. Code Map drift from the spec's
measurement commit `f078b04`: NONE affecting any anchor.** `git log --oneline f078b04..d0f183b` is a
single commit, `d0f183b` *"Plan Story 8.4e, and find that a standing red was never a red"*, and
`git diff --stat f078b04..d0f183b` touches exactly two files, both under
`_bmad-output/implementation-artifacts/`: this spec (created, 553 lines) and `epic-8-context.md`.
**No `folio-go`, `folio-designer` or `lint` source file changed**, so every line anchor in the Code
Map, every guard row in the fifteen-assertion fate table, and every `## Verification` expectation is
exact as written. Recorded per the Code Map's own instruction. The intent contract is untouched:
8410 bytes, md5 `17a5bd706ba1010eefaca282283f54b6`, re-verified unchanged after implementation, after review patching, and at finalize.

## Review Triage Log

### 2026-09-01 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 4: (high 0, medium 2, low 2)
- defer: 3: (high 0, medium 1, low 2)
- reject: 16
- addressed_findings:
  - `[medium]` `[patch]` **The new browser-side face-name predicate was tied to nothing the engine actually ships.** `shipped-face-family.ts:65` gates attribution on a regex admitting only ASCII words; `canvas-font-stack.test.ts` reads the real `fonts.Shipped()` keys into `engineFaces` and already tied them to `declared` (@font-face) and `requested` (App.css) plus the order tie, but never to this new third authority. A future shipped face named e.g. `IBM_Plex_Sans`, `Noto Sans 2.0` or a CJK name would pass all three existing ties, be declined by the predicate, set no inline family, and silently fall back to the fixed Latin-first stack — reintroducing this story's own defect with nothing red. Found independently by three of the four review layers. Fixed by asserting, as SET DIFFERENCES read from `fonts.go` (never literals), that every shipped name satisfies `isShippedFaceName` and that the seam names each one FIRST. Mutation-proved by defeating the predicate: the `fonts.go`-derived test reddens.
  - `[medium]` `[patch]` **The closed-set census admitted the exact evasion it exists to catch.** `shippedFaceDerivedFamily` was `/^shippedFaceFamily\([A-Za-z][A-Za-z0-9_]*\.face\)$/` — any bare identifier `.face`, so a per-component `shippedFaceFamily(entry.face)` sailed through the guard Design Note 7 designates against weakening by evasion. Anchored to `/^shippedFaceFamily\(fragment\.face\)$/`; the pre-existing `assetKeyDerivedFamily` left alone. Assertion 14 gained the three cross-wiring near-misses its own comment claimed but omitted: `shippedFaceFamily(fragment.assetKey)`, `embeddedFaceFamily(fragment.face)`, `shippedFaceFamily(entry.face)`.
  - `[low]` `[patch]` **The new Go bound check is unreachable and therefore carries no mutation proof.** Measured: deleting `page_setup.go:1528` entirely leaves the Go suite at 1815 pass / 2 fail / 5 skip — byte-identical to unmutated. `canvasFontChains` (`:764`) runs before `addCanvasTextPaint` (`:785`), and `projectFontChainEntry` already refuses the same string over the same bound, so the earlier refusal always fires first. The check was KEPT as defence in depth at a wire boundary, and its comment now records the unreachability and names the site that refuses first, so no later reader mistakes it for a proved guard. The contract requires added guards to be mutation-proved by deletion; this one cannot be, and says so rather than pretending otherwise.
  - `[low]` `[patch]` **DW-35's closure contradicted itself and omitted the implementing commit.** Two internal headers still asserted partial delivery (`Owner — ONE PER CAUSE, because the causes were split and only one of them is delivered`; `CAUSE ONE — SPLIT INTO TWO LAYERS, AND ONLY ONE OF THEM IS DELIVERED`) directly above bullets saying all four things are closed. Both re-worded to record delivery while preserving the historical narrowing. Commit `21f93b4` added to the Status bullet and the 8.4e closing-note header, matching the file's own convention (cf. DW-36's `c834158`, DW-35's own 8.4b note `90cdf8e`).

**Rejected findings, each with the authority it was tested against** (DW-87 — the population enumerated is *every* finding returned by all four review layers: blind-hunter 18, edge-case-hunter 3, verification-gap 3, intent-alignment 4 divergences, deduplicated to 25 distinct claims):

1. *Silent declension path — no warning or diagnostic when a name is declined.* Refuted: DW-93 (no canvas diagnostics channel) is ruled acceptable and is named out of scope by this dispatch.
2. *`shipped-face-family.test.ts` admits the undeclared `'Noto Serif'` and puts it first, contradicting "asks only for families the browser has a face for".* Refuted at the cited location by I/O matrix row 8, which REQUIRES exactly that: an attributed name not among the declared families "falls through its own tail to the declared stack and then `sans-serif`; nothing blank".
3. *The family-parsing test helper is spelled three times and the copies disagree on the empty string.* Test-only; no assertion in any of the three files depends on the empty-string case. Consequence for the artifact's consumer: none.
4. *Rejecting a both-identities fragment terminates the worker, an unargued asymmetry against the tolerated neither-case.* Refuted by the contract's own Block If, which states a fragment carrying both "invalidates the wire shape". The TS refusal is contract-aligned; the asymmetry is the contract's deliberate choice.
5. *`shippedFaceFamily` is evaluated twice per fragment per render.* At most three distinct inputs, trivial string work, and the duplication is forced by the census guard's anchored regex. Cosmetic.
6. *`reflect.DeepEqual` where `slices.Equal` would do; O(n²) dedup.* Style only, in test code, over a three-element list.
7. *`expect(callers).toEqual(['App.tsx'])` is order-sensitive on `readdirSync`.* Single-element list, so order is moot; and a second caller SHOULD redden — that is the closed-set census working as designed, not a false positive.
8. *DW-35 is marked CLOSED while the story is `in-review` and `sprint-status.yaml` still says `backlog`.* The register is amended in the same commit that delivers the story, which is the file's established practice; `sprint-status.yaml` is explicitly the closer's file and `bmad-build-auto` never touches it.
9. *DW-102's new owner is "the owner's" with no decision id, sprint row or escalation.* Adding routing machinery for DW-102 would be absorbing it, which this dispatch forbids by name. Task 9's requirement — remove Story 8.4e from its `Owner:` bullet and record the ground — is satisfied.
10. *The Spec Change Log claimed a close-time check before any close happened.* Correct as to timing; the wording was mine and is now amended to state exactly when each check ran. Not a code finding.
11. *The carried branch sets an inline family with no fallback tail, so a missing codepoint hits the browser default.* Refuted by Task 6, which requires "leaving the carried branch exactly as it is"; this is pre-existing Story 8.4a behaviour, unchanged here.
12. *The browser-level half of the fix has no executing home.* Disclosed, not defective: Design Note 6 and D-8.4.25(b)/(d)/(e) place the executed browser assertion at the epic gate behind DW-101, and the contract's Block If forbids adding an observer to a suite CI does not run.
13. *"Wrong advances" — the defect the intent names — is never asserted anywhere.* Same authority as 12: AC2 is deliberately phrased at the rendered-DOM layer because jsdom applies no stylesheet and loads no font. Asserting advances requires the executed browser the contract defers.
14. *The predicate's decline path is a fourth browser behaviour the matrix does not enumerate.* Its outcome is identical to matrix row 5's (falls back to the stylesheet's declared stack, canvas still paints), and it is unreachable from `fonts.Shipped()`'s three keys. No new consumer-visible behaviour.
15. *Record-surface edits beyond the intent's ask (DW-102 reassignment, the D-000.4 citation rewrite).* Both are required by Task 9 by name, with the ground recorded in Design Note 5.
16. *The mutation proofs are prose in comments, unverifiable from the diff.* Answered by measurement rather than by an edit: this dispatch re-ran the core mutation independently — removing the shipped branch from `TextPaint` reddens four named tests — plus the predicate and census mutations, and the Go bound deletion which reddened nothing (finding 3 above).

## Design Notes

**1. Why the wire value is the `FontSet` name, and why that is not a re-litigation.** D-8.4.14 settled
it in as many words: *"a carried face's browser family derives from the engine's identity for it (the
asset key); a shipped face's from the engine's identity for it (the `FontSet` name). One rule for one
question."* It also rejected the two alternatives by name — renaming the generated `@font-face`
families (wrong: IBM Plex is the design system's specified typeface) and a face-name → CSS-family
mapping table (wrong: a second authority maintained in lockstep with `Shipped()`). Story 8.4b then
made the name usable by declaring an `@font-face` under each. **The browser family now *is* the
engine's name, so there is nothing to map** — which is exactly why this story is small.

**2. Why `fragment.face` is safe to put on the wire for a shipped face.** `resolveRuneFace` returns an
element of `chainFaceNames(chain)`, and a chain entry whose face is absent from the supplied `FontSet`
is silently skipped — so a face can only be attributed if the engine actually loaded and measured with
it. The value is therefore drawn from `Shipped()`'s three names at runtime, not from arbitrary
document text. **Do not let that reasoning become the browser's security model**: the TS validator
must still bound and shape-check the string, because the guard's job is to hold when the Go side is
wrong.

**3. Assertion 8 is a disclosure of absence, and this epic already has the precedent for retiring one.**
Its measured content is thin — no `var(` in the declaration, plus a floor duplicated at assertion 4 —
while its weight sits in a prose header that names *"CAUSE ONE, ATTRIBUTION LAYER (STILL OPEN)"* and
says closing it *"is a different story"*. That is a written pre-authorisation of its own deletion. The
precedent to copy is recorded in DW-35's own 8.4a closing note: Story 8.4's disclosure of absence
*"was deleted under its own written pre-authorisation and replaced by its positive twin"*, detector
unchanged and still proved against its fixtures. Do the same here: the twin asserts that a shipped
fragment's family **does** come from the projection, through exactly one named seam and nowhere else.
**Two bounds must survive the swap** or they are silently dropped: the `no var(` clause (keep an
equivalent, so the fallback tail cannot become an indirection that empties `requestedFamilies`) and
the `requested.length >= 3` floor (already duplicated at assertion 4 — confirm, do not assume).

**4. Why `.canvas-text-fragment` stays, when the AC's phrasing invites deleting it.** Three independent
selectors, all pointing the same way. (a) Assertion 9's `requested.slice(0,3) toEqual engineFaces` is
D-8.4.14's delivered guard, and it is computed **from that rule** — deleting the rule destroys the
guard, which is the weakening the epic's own convention forbids. (b) `requestedFamilies` **throws at
`describe` scope**, so removing or reformatting the rule takes five tests down with it, and a suite
that cannot run is not a suite that passed. (c) 8.4a's carried path already keeps the stylesheet stack
as its degrade path — *"a fetch that fails degrades to the stylesheet's declared stack with the canvas
still painting"* — and this story extends that mechanism rather than replacing it. The rule stops
being the **authority** and becomes the **fallback**; that is what "replaced, not weakened" means
here.

**5. DW-102 is NOT this story's, and here is the ground.** *(The dispatch delegated this decision
explicitly; it is recorded rather than escalated.)* Five selectors, all one way:
- **Scope.** D-8.4.26 scopes 8.4e to *"per-fragment shipped-face attribution on the wire"*. DW-102 is
  about which face `--font-mono` resolves to. Different subject, different surface.
- **A standing ruling forbids the remedy here.** DW-102's fix is a chrome token edit. D-8.4.14 ruled
  the canvas fix must *"edit no chrome token at all"*, and assertion 6 of `canvas-font-stack.test.ts`
  **reddens** if an engine face name reaches a `--font-*` token. Folding DW-102 in would require
  breaking a guard this story is otherwise obliged to keep.
- **Measured separation.** DW-102's live consumers are `.document-name` and `.property-value` — chrome
  elements. This story touches `.canvas-text-fragment` only. They share no selector, no token and no
  code path.
- **Reachability (the D-000.65 test).** 8.4e does not make DW-102's condition newly reachable; Story
  8.4c did, when the mono binary changed. By the rule this run uses to place a defect, it is not 8.4e's.
- **It carries a decision above a builder's authority.** DW-102's own discharge is *either* ship a CJK
  fallback face — a bundle-weight commitment that collides with Story 8.4d's budget — *or* record a
  deliberate decision that chrome CJK falls to the system face. Both are the owner's call, and DW-102's
  Owner bullet already says *"Story 8.4e **or the owner**"*.

  **Verdict: the owner's.** Task 9 amends DW-102's `Owner:` bullet to say so with this ground, so it
  does not sit as a candidate on a story that has ruled itself out — the ownerless-drift failure that
  already happened once this epic, to DW-35's own residual.

**6. What this story's gates can and cannot prove, stated before any code exists.** jsdom applies no
stylesheet and loads no font; `test:e2e:compile` is `tsc --noEmit`. So these gates prove the
attribution, the wire shape, the derivation, the element's requested family and the guards. **They do
not prove a glyph was rasterized with that face.** Per **D-8.4.25(b)** the trigger for that is *"when
CI executes the Playwright suite"* — **not** the void *"when browser e2e arrives (D-000.4)"*, which
named an event that had already happened. Per D-8.4.25(d)+(e) the executed browser assertion is owed at
the **epic gate**, behind CI wiring (DW-101), and must not be added to a suite CI does not run. **Say
this at close rather than letting a compile pass read as a run.**

**7. The trap that has caught three guards in this epic already.** A raw-text scan stayed green over
**commented-out** `@font-face` rules; a chrome-token guard stayed green over a live token parked in a
**CSS comment**; and a `declared` check with a floor and **no ceiling** let a seventh rule with a
literal remote `src` through while 19 tests passed. Assume the same of every guard written here until
it is mutated. **A count is lossy — assert the set difference in both directions.** And watch
**weakening by evasion**: an inline family string escapes an `App.css`-only scan without anyone editing
the guard, which is precisely the shape this story introduces, since the fragment's family is moving
**from the stylesheet into an inline style**. Assertion 13's census is the guard that must catch that,
and widening it to a **closed set of exactly two** approved expressions — never to `arrayContaining` —
is what keeps it able to.

## Verification

**Baseline MEASURED at `f078b04` (this dispatch's HEAD), clean tree, in place — not assumed.**
Re-measure before the first edit and record any drift in `## Spec Change Log`.

**Commands:**

- `cd folio-go && go test -count=1 ./...` — expected: **1811 pass / 2 fail / 5 skip**. The two are
  `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)` in `internal/text`
  (`P6g floor not met: got 7, need >=20`). **Mandated permanent red; never "fix" it.**
- `cd folio-go && go test -count=1 -tags=matrix ./...` — expected: **1822 pass / 3 fail / 5 skip**. The
  third is `TestShippedFacesReproduceFromUpstream`, and it is a **could-not-execute, not a byte
  divergence** — verbatim: `fontgen: fontTools is not importable by this interpreter`
  (`/opt/homebrew/opt/python@3.12/bin/python3.12`). **It never compared bytes.** Report it in those
  words per **DW-86**.
- **Exactly TWO standing reds. Any third is a real failure.**
- `cd folio-go && go vet -tags=matrix ./...` — expected: no output. *(Measured empty.)*
- `gofmt -l folio-go` — **run from the REPO ROOT**, expected: no output. Running it after a
  `cd folio-go` in the same compound command prints an `lstat` error that reads like success.
- `cd lint && go test -count=1 ./...` — expected: four `ok` lines, no FAIL. **`-count=1` always**: this
  module's rules walk directories and Go's cache does not track `ReadDir`, so a cached `ok` is not a
  measurement.
- `cd folio-go && FOLIO_MATRIX_TARGET=<t> go test -count=1 -tags=matrix -run TestTargetRenderHash -v .`
  for each of `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` — expected: PASS, **24 documents
  hashed per leg** (measured). Then the same command with the variable **unset** — expected: PASS in
  **0.00s** while asserting **nothing**, saying so at `matrix_test.go:2199`. **It is a control, never a
  fifth leg.**
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` — expected: PASS
  (~23s measured).
- `cd folio-designer && npm run typecheck` — expected: clean.
- `cd folio-designer && npm run lint` — expected: **exactly 4** `only-export-components` warnings, at
  `preview/pdf-viewer.tsx:16,17` and `App.tsx:1323,1330`. **These line numbers move when `App.tsx`
  changes; the count of 4 and the rule name are the invariant.**
- `cd folio-designer && npm test` — baseline **38 files / 372 tests, all passing**. Use the project's
  own script; `--reporter=basic` is not a valid reporter here and exits 1, which reads like a failure.
- `cd folio-designer && npm run test:e2e:compile` — expected: clean. **This is `tsc --noEmit`. Do not
  report it as a run.**
- `shasum -a 256 fixtures/*/expected.pdf` — expected: **23** lines, digests unchanged.
- `md5 -q README.md` — expected: `078d7d80d518d54af2fc04fb270d46b8`, unchanged.
- Offline release, required because this story touches designer sources that reach the bundle:
  `cd folio-designer && npm run build`, then `npm run verify:offline` / `npm run verify:offline:red` /
  `npm run verify:offline:wasm` — expected: pass, node **v24.16.0**. **Do not relay any
  `s1VisibleBytes` figure without re-measuring it and quoting the command** (DW-100, D-8.4.27b); it
  drifts between builds because the engine wasm embeds its commit stamp.

**Manual checks:**

- **Mutation-prove the new attribution, per site and in both directions.** Emit the face name for a
  carried face too → the wire test's carried assertion must redden. Stop emitting it for a shipped face
  → the shipped assertion and AC2's DOM assertion must redden. A green suite over correct code proves
  nothing.
- **Delete each newly-added guard and confirm a NAMED test goes red.** A guard that survives its own
  deletion is decoration.
- **Widened guards, both directions.** For the census: the new approved expression is accepted, and a
  document-supplied family in a `font-family` position still reddens. For assertion 7's DOM half: the
  shipped fragment's family is asserted, and an unattributed fragment still falls back.
- **Confirm the fallback tail has one authority.** If it is spelled in both the seam and `App.css`,
  change one and confirm a named test reddens.
- **Confirm the echo path.** Repeated sheets mount `TextPaint` a second time; a one-window fixture will
  not notice a fix that misses `ComponentEcho`.
- **Confirm the corpus did not move** before claiming AC5: a changed PDF digest is a defect, not a
  re-record.

## Auto Run Result

Status: done
Blocking condition: none

**Dispatch:** classic-intent, implement/review/commit against the existing spec (no `Halt after planning.`). `baseline_revision` `d0f183b73940b1556b058ab3cbf5caf6428407e4`, tree clean at start. **Code Map drift from the spec's measurement commit `f078b04`: none affecting any anchor** — the single intervening commit touched only this spec and `epic-8-context.md` (see `## Spec Change Log`). `<intent-contract>` never edited: 8410 bytes, md5 `17a5bd706ba1010eefaca282283f54b6`, verified after implementation, after review patching, and at finalize.

**Implemented change.** A canvas text fragment drawn with a **shipped** face now carries the engine's own `FontSet` face name on the wire, and the browser asks for that face instead of falling through to `.canvas-text-fragment`'s fixed Latin-first stylesheet constant. Story 8.4a's carried-face attribution mechanism gains a second population; nothing is mapped, renamed or registered, because Story 8.4b already declares an `@font-face` under each of the three engine names. The stylesheet rule survives as the fallback for an unattributed fragment — "replaced, not weakened" applies to the guard, not to the stylesheet.

**Files changed.**
- `folio-go/page_setup.go` — `CanvasTextFragment` gains a `Face` field carrying the JSON key `face` with `omitempty`, populated at the fragment append from `fragment.face` when the face is not carried; bounded by the existing `maxCanvasPropertyString` (512) with a stated refusal, never silently emptied. Doc comment answers the two questions the `AssetKey` comment answers, and now also records that the bound check is unreachable by construction.
- `folio-go/canvas_projection_wire_test.go` — the accepted key record widened to `{assetKey, face, text, x}`; both emission directions pinned as exact key sets, plus an exclusivity check.
- `folio-go/canvas_fragment_attribution_test.go` — four new tests: FontSet-name attribution, Latin through a Thai-first chain, mixed-script per-fragment naming, and the mutual-exclusivity probe. `TestAShippedFaceCarriesNoAssetKey` unmodified and still green.
- `folio-designer/src/engine-protocol.ts` — `face` added to the fragment `hasOnly` list **and** to the `CanvasProjection` fragment type in the same edit (the type was pinned by nothing); validated non-empty and `<= MAX_CANVAS_PROPERTY_STRING`; both-identities rejected, neither-identity still legal.
- `folio-designer/src/shipped-face-family.ts` + `.test.ts` (new) — the one named derivation seam: attributed face quoted and first, then the declared tail; declines a name not usable as a CSS family rather than interpolating it.
- `folio-designer/src/App.tsx` — one ternary arm in `TextPaint`, so it holds at both call sites including `ComponentEcho`; carried branch untouched.
- `folio-designer/src/canvas-font-stack.test.ts` — assertion 8 replaced by its positive twin with **both** of its bounds preserved (no `var(`, and the `>= 3` floor); 7, 13 and 14 widened; census is a closed set of exactly two, anchored to `fragment.face`; a new tie reads the fallback tail from **both** the seam and `App.css` and asserts they agree; and the shipped names read from `fonts.go` are now tied to the browser-side predicate in both directions.
- `folio-designer/src/App.test.tsx`, `engine-protocol.test.ts` — AC2 at the rendered DOM (home **and** echo, plus the unattributed-fallback direction) and the protocol-guard cases.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-35 CLOSED (both causes and the residual), its two stale passages corrected (the ownerless/"escalated to the orchestrator" claim, superseded by D-8.4.26; the void `when browser e2e arrives (D-000.4)` trigger, replaced by D-8.4.25(b)), self-contradicting headers repaired, commit `21f93b4` recorded. DW-102 re-owned to the owner with its five-selector ground; Story 8.4e removed as a candidate.

**Review findings breakdown.** 4 patches applied (2 medium, 2 low), 3 items deferred (1 medium, 2 low), 16 rejected — each with the authority it was tested against, enumerated in `## Review Triage Log`. 0 intent gaps, 0 bad-spec loopbacks; `review_loop_iteration` stayed 0.

**Follow-up review recommendation: `true`.** Patched findings only: high 0, medium 2, low 2 → `3x2 + 1x2 = 8`, which is >= 5.

**Verification performed — measured, in this dispatch, after the patches.**
- `cd folio-go && go test -count=1 ./...` → **1815 pass / 2 fail / 5 skip** (baseline 1811/2/5; +4 new Go tests). The two are `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest, `P6g floor not met: got 7, need >=20` — the mandated permanent red.
- `cd folio-go && go test -count=1 -tags=matrix ./...` → **1826 pass / 3 fail / 5 skip** (baseline 1822/3/5). The third is `TestShippedFacesReproduceFromUpstream`, a **could-not-execute, not a byte divergence** — verbatim `fontgen: fontTools is not importable by this interpreter.` It never compared bytes (DW-86). **No third distinct failure.**
- **Both ways, as this dispatch required.** With `FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python` the same test **PASSES non-vacuously** — `fontgen: derived and compared 3 of 3 faces`, real source and produced sha256s, all three committed faces reproducing from their recorded derivation. So there is no divergence in the shipped faces. The variable was set only on the command line; it is in no committed file, and the test itself was not edited.
- `go vet -tags=matrix ./...` empty. `gofmt -l folio-go` **run from the repo root** empty. `cd lint && go test -count=1 ./...` four `ok`, no FAIL.
- Four AD-21 legs (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) **PASS, 24 documents hashed each**. The unset control **PASSES in 0.00s asserting nothing** (`matrix_test.go:2199`) — a control, never a fifth leg. `TestCrossTargetByteIdentity` **PASS (~23s)**.
- Designer: `typecheck` exit 0; oxlint **exactly 4** `only-export-components` warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1324,1331` — the App.tsx lines moved from 1323/1330 because `App.tsx` changed; the count and rule name are the invariant). `npm test` → **39 files / 383 tests, all passing** (baseline 38/372). `npm run test:e2e:compile` clean — **this is `tsc --noEmit`, not a run.**
- `shasum -a 256 fixtures/*/expected.pdf` → **23 lines, byte-identical to the pre-dispatch snapshot** (diffed, not eyeballed). `md5 -q README.md` → `078d7d80d518d54af2fc04fb270d46b8`, unchanged. **23 golden digests but 24 AD-21 documents per leg — different populations, not conflated.**
- Offline release on node **v24.16.0**: `npm run build`, `verify:offline`, `verify:offline:red`, `verify:offline:wasm` all exit 0. Measured after `cd folio-designer && npm run build`, reading `dist/offline-release-manifest.json`: `s1VisibleBytes` **12,427,899**, `s1.cachedBytes` **38,460,399** over **23** assets. Quoted with its command because the figure drifts between builds of identical source; it is **not** a metric and no threshold was set (Story 8.4d, D-8.4.24, stayed out of scope).

**Matrix test audit.** All eight I/O matrix rows have a covering test that ran and passed: rows 1-4 by the four new Go attribution tests (the exclusivity probe walks a real population in 0.63s, not vacuously); row 5 by the unattributed-fallback DOM test; row 6 by `TestCanvasProjectionWireKeysAreTheOnesTheDesignerAccepts`, the cross-language gate that reads `engine-protocol.ts` from disk; row 7 by the pre-existing chain-entry bound probe the row itself names, plus the TS bound test; row 8 by the seam's undeclared-name test.

**Mutation proofs run independently by this dispatch** (not merely relayed): removing the shipped branch from `TextPaint` reddens **four named tests** — AC2 at home and echo, the AC3 positive twin, the widened assertion 7, and the census. Defeating the face-name predicate reddens the `fonts.go`-derived tie. Deleting the new Go bound check reddens **nothing** — recorded as finding 3 rather than papered over.

**Residual risks.**
1. **Nothing here proves a glyph was rasterized with the attributed face.** jsdom applies no stylesheet and loads no font; `test:e2e:compile` is `tsc --noEmit`. What is proved is that the right name reaches the element and the element asks for it. The repository does contain 12 executable Playwright specs and a working config, but CI runs only the compile step (DW-101). Per D-8.4.25(b) the trigger is *"when CI executes the Playwright suite"* — the old *"when browser e2e arrives"* trigger is void. **A compile pass is not a run.**
2. The three deferred items in frontmatter, chiefly that closing DW-35 leaves the owed executed-browser assertion recorded only in a closed entry's prose.
3. `TestShippedFacesReproduceFromUpstream`'s disposition remains with the engineering lead. It was neither edited nor accommodated here.
