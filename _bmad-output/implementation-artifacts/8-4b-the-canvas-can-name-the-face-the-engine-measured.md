---
title: 'Story 8.4b: The canvas can name the face the engine measured'
type: 'feature'
created: '2026-09-01'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized']
deferred: []
---

<intent-contract>

## Intent

**Problem:** The engine measures shipped text with the faces `fonts.Shipped()` names — `Noto Sans`, `Noto Sans Thai`, `Noto Sans SC` — but the browser has no way to *name* those faces: the generator registers the three shipped files under IBM Plex family names, and `canvas-font-stack.test.ts` records the two vocabularies as deliberately disjoint. AD-17 makes the browser a rasterizer only, so it must rasterize with the face the engine measured with, and it cannot unless it can name that face. Today the canvas is correct **by accident** — `'IBM Plex Sans', 'IBM Plex Sans Thai', 'IBM Plex Mono'` happen to resolve to `NotoSans-Regular`, `NotoSansThai-Regular`, `NotoSansSC-Regular`, precisely the engine's three faces in the engine's own order — and Story 8.4c puts real IBM Plex bytes behind those names, at which point every fragment in every script rasterizes with a face the engine never measured with.

**Approach:** Register the three shipped files **additionally** under the engine's own face names — the same files, a second `@font-face` each, the family named from the `FontSet`'s own spelling — and point `.canvas-text-fragment` at those names. Not one chrome token is edited. Replace the disjointness assertion with its successor (the canvas stack *contains* the engine's face names, read from the engine's own authority rather than a hardcoded list), and build the guard nothing in this repository has: an assertion binding a family name to the file behind it, whose first job is to pin the deliberate two-names-one-file interval and go red on a "simplification".

## Boundaries & Constraints

**Always:**
- The three new `@font-face` families are the `FontSet`'s own spellings, verbatim: `Noto Sans`, `Noto Sans Thai`, `Noto Sans SC`. Each points at the **same already-registered file** as its IBM Plex counterpart — no new binary, no new `assets` slot, no copied byte.
- The canvas fragment stack stays **file-order preserving**: the new stack names the same three files in the same order the current one resolves to, so this story changes no rasterization today.
- Every generated rule keeps the exact literal spelling `@font-face { font-family: 'NAME'; src: url('./runtime/${SLOT}') format('truetype'); font-display: swap; }`, each rule's prefix contiguous in the generator source. `.canvas-text-fragment` stays a **single line** in `App.css` beginning at column 0, with every family a quoted literal and `sans-serif` last.
- Guards are **widened or replaced, never weakened**. Every assertion in `canvas-font-stack.test.ts` gets an explicit written fate before any of them is touched.
- Every new assertion carries a **non-vacuity floor** and a **red-proof that reddens on deletion or on a wrong binding**, not merely on an unrelated production change.
- Every acceptance criterion is stated at a layer the suite actually executes: a raw-text parse of a tracked source, a filesystem byte comparison, or a jsdom render.

**Block If:**
- Closing this story would require editing `--font-sans`, `--font-mono`, `--font-page` or any `--type-*` token, adding a font binary, or amending `DESIGN.md`, `ARCHITECTURE-SPINE.md` or the release licence manifest.
- The three files the generator registers under IBM Plex names are found **not** to be byte-identical to the files `folio-go/fonts/fonts.go` embeds under the corresponding face names. The whole story rests on the browser's copy being the engine's face; if the mirror has drifted, naming it is a lie and the story needs a ruling, not an implementation. (Measured false at plan time — see Code Map — so this fires only on a change under the builder's feet.)
- A guard's replacement cannot be made to redden on deletion without also reddening at baseline.

**Never:**
- Never build a mapping table from IBM Plex names to Noto faces, in either direction, in any language. It is a second authority on which browser family corresponds to which engine face, maintained in lockstep with the shipped `FontSet`, and it is the alternative this verdict rejects by name.
- Never rename, remove or repoint an existing IBM Plex `@font-face` rule, and never make chrome ask for a Noto family. "The vocabularies now meet" is not licence to merge them; that is Story 8.4c's question, settled there and settled the other way.
- Never soften the disjointness assertion to an `arrayContaining` on the chrome half to keep it passing, and never delete it without relocating its scanning power.
- Never make the fragment stack take document input (a `var()`, an interpolation, a projected chain). It is a stylesheet constant and stays one.
- Never move a `.folio` document, fixture, golden or attestation record; never write `reader`, `date` or `examined` into a sign-off file. The corpus is 23 documents and a moved PDF digest is a defect, not a re-record.
- Out of scope: real IBM Plex binaries, the `IBM Plex Mono`→CJK defect and the licence-extension guard (Story 8.4c); the size budget gate (Story 8.4d); a canvas diagnostics channel (DW-93, ruled acceptable).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Engine face is nameable | the generated stylesheet | an `@font-face` exists for each of `Noto Sans`, `Noto Sans Thai`, `Noto Sans SC`, alongside the three unchanged IBM Plex rules — six rules, three files | No error expected |
| Canvas asks the engine's vocabulary | `.canvas-text-fragment` | its `font-family` names the three engine faces in engine order, then `sans-serif`; every name is one an `@font-face` declares | No error expected |
| The deliberate interval | `IBM Plex Sans` and `Noto Sans` (and the other two pairs) | both resolve, through the generator's own two halves, to one and the same source file | No error expected |
| A "simplification" of the interval | one of the six rules deleted, or a pair repointed to differing files | the family→file map no longer equals the pinned map; the failure names Story 8.4c as the successor that makes them diverge | Test failure, red |
| A file swap under an unchanged name | an engine-named family repointed at other bytes, or its file replaced in place | the engine-named family's file is no longer byte-identical to the file `fonts.Shipped()` embeds under that face name | Test failure, red |
| The engine's shipped set changes | a fourth face added to, or a face removed from, `fonts.Shipped()` | the engine-name set parsed from `fonts.go` no longer matches the engine-named families the generator declares | Test failure, red |
| Chrome is untouched | `tokens.css`, `design-tokens.ts`, `DESIGN.md` | no diff; the three IBM Plex families remain declared and remain what every `--type-*` token resolves through | No error expected |
| A guard's own parse breaks | the generator's rule spelling or `App.css`'s single line reformatted | the non-vacuity floors fail rather than the scans passing over an empty set | Test failure, red |

</intent-contract>

## Code Map

**Every anchor below measured at `475cb50`, this dispatch's HEAD.** Cite assertions by what they assert; the line numbers are a convenience and D-8.4.13 rules them stale across any boundary.

### The two halves of the join nothing performs

- `folio-designer/scripts/build-wasm.mjs:64-71` — the `assets` object. Maps **opaque slot names** to file paths and carries **no family name**:
  `sans` → `public/fonts/notosans/NotoSans-Regular.ttf`, `sansCjk` → `notosanssc/NotoSansSC-Regular.ttf`, `sansThai` → `notosansthai/NotoSansThai-Regular.ttf`. `fingerprint(source, label)` (`:45-53`) copies into `src/generated/runtime/` under the **label**, not the family.
- `folio-designer/scripts/build-wasm.mjs:79` — the **single-line template literal** writing `src/generated/runtime-fonts.css`, hardcoding three rules keyed to slots **by hand**. There is no loop. `'IBM Plex Sans'`→`${assets.sans}`, `'IBM Plex Mono'`→`${assets.sansCjk}`, `'IBM Plex Sans Thai'`→`${assets.sansThai}`.
- **`'IBM Plex Mono'` is bound to `NotoSansSC-Regular.ttf`** — a CJK face under a mono name. It is why the current canvas stack resolves correctly, and it is Story 8.4c's defect, not this story's.
- `:76-77` writes `offline-assets.ts` by iterating `Object.entries(assets)` **generically**. Adding a fourth `@font-face` would need a hand edit at `:79`; **this story adds no slot**, so `offline-assets.ts`, the emitted asset set and the release manifest's asset count are all untouched.
- `src/generated/runtime-fonts.css` and `src/generated/runtime/` are **gitignored** — which is why every guard reads the *generator source*, never its output.

### The resolution, verified — the load-bearing fact of the whole sequence

| `.canvas-text-fragment` asks | generator binds to slot | slot's file | engine `FontSet` name |
|---|---|---|---|
| `'IBM Plex Sans'` | `assets.sans` | `NotoSans-Regular.ttf` | `Noto Sans` |
| `'IBM Plex Sans Thai'` | `assets.sansThai` | `NotoSansThai-Regular.ttf` | `Noto Sans Thai` |
| `'IBM Plex Mono'` | `assets.sansCjk` | `NotoSansSC-Regular.ttf` | `Noto Sans SC` |

**The canvas is accidentally correct for every script today**, in the engine's own order. So the rewrite to `'Noto Sans', 'Noto Sans Thai', 'Noto Sans SC', sans-serif` is **order- and file-preserving and changes no rasterization**.

### The engine's authority for a face name

- `folio-go/fonts/fonts.go:57-62` — `func Shipped() folio.FontSet` returns exactly `{"Noto Sans": notoSans, "Noto Sans Thai": notoSansThai, "Noto Sans SC": notoSansSC}`. The three `//go:embed` directives are at `:41-48` (`notosans/NotoSans-Regular.ttf`, `notosansthai/NotoSansThai-Regular.ttf`, `notosanssc/NotoSansSC-Regular.ttf`). Face name → embed var → path is a two-hop join **within one 60-line file**.
- **There is no other machine-readable enumeration of the three face names.** Measured: no exported constant, no JSON registry, no generated artifact. The only other lists are `folio-designer/public/templates/starter.folio:9` (a document's chain, not a face registry), Go test fixtures, and `canvas-font-stack.test.ts:304`'s own hardcoded `engineFaces`.
- **The mirror is byte-identical, measured at `475cb50`** — `folio-designer/public/fonts/*` vs `folio-go/fonts/*`, sha256: `a4c81131…2d6c` (Noto Sans), `c94562c1…3caf` (Thai), `5ef5755b…0158` (SC). Nothing in the repository currently enforces this; `lint/internal/rules/fontsassets.go:47` scopes its scan to `folio-go/fonts` and never reaches the designer.

### The stylesheet

- `folio-designer/src/App.css:118` — the sole `font-family` declaration in the file:
  `.canvas-text-fragment { left: var(--text-fragment-x); top: 0; font-family: 'IBM Plex Sans', 'IBM Plex Sans Thai', 'IBM Plex Mono', sans-serif; }`
- `.canvas-text-fragment` appears in exactly one other rule, `:106`, which sets **no** font-family. **No cascade conflict.**
- `folio-designer/src/tokens.css:11` — the three chrome tokens. `--font-sans` feeds 7 `--type-*` tokens, `--font-mono` 7, `--font-page` 3. **`.canvas-text-fragment`'s families are referenced nowhere in `tokens.css`** — the fragment stack is a hardcoded literal, entirely disjoint from the token layer, which is exactly what makes "no chrome token is edited" satisfiable.

### The guards — every assertion in `canvas-font-stack.test.ts` (443 lines) with its fate

| asserts | fate |
|---|---|
| `declared.length >= 3` (vacuity) | **KEEP** — floor raised to `>= 6`, the true count after this story |
| `requested.length >= 3` (vacuity) | **KEEP** unchanged |
| (a) `requested ⊆ declared`; (b) the carried-face runtime tie via `paintedFragmentFamilies`; (c) the seam derivation; (d) no derived/declared family collision | **KEEP, all four.** (a) is what forces the two edits to land **together**; (d) strengthens for free (6 families now). Its comment block asserting the shipped half is *"FALSE, before this story and after it"* becomes false itself and must be rewritten |
| the fragment stack is a stylesheet constant — `not.toMatch(/var\(/)`, `requested.length >= 3` | **KEEP** unchanged. The DW-35 tripwire comment above it, calling cause one *"a design-system decision above a builder's authority"*, is **MEASURED-FALSE** per D-8.4.14 and must be rewritten |
| **`declared` contains the three IBM Plex names; `engineFaces ∩ declared = []` both ways** | **REPLACE.** The `arrayContaining` floor on the IBM Plex half **stays** (chrome is unchanged and must remain declared); the two intersection assertions **go red and are replaced** by containment |
| runtime registration in exactly one named seam; generator declares each `declared` family; generator registers nothing at runtime | **KEEP** unchanged — green with six rules, since the new rules use `url('./runtime/…') format('truetype')` |
| the site list's red-proof (second site red, quiet seam red, a mention is not a registration) | **KEEP** unchanged |
| the detector's own proof against 4 positive + 3 negative fixtures | **KEEP** unchanged |
| `positions` exactly `['App.tsx: embeddedFaceFamily(fragment.assetKey)']` | **KEEP** unchanged — and it is what closes the evasion route below |
| that guard's red-proof, both directions, prose left alone | **KEEP** unchanged |
| generic fallback last; every earlier entry starts `'` | **KEEP** unchanged — satisfied by the new stack |

- **Mechanical dependencies the edits must respect.** `declaredFamilies` regexes the generator for `/@font-face \{ font-family: '([^']+)'/g` — exact spelling, single space after `{`, single quotes; any drift **silently empties `declared`** and only the vacuity floors catch it. `requestedFamilies`, and two other assertions, locate the rule with `css.split('\n').find(line => line.startsWith('.canvas-text-fragment {'))` and **throw** if it is reformatted.
- **`designerSources()` walks `src/` recursively for `.ts`/`.tsx` only** — `scripts/build-wasm.mjs` is outside every designer scan corpus.

### What is NOT affected — measured, not assumed

- `design-contract.test.ts`: token-**name** equality only (`design-tokens.ts` carries no family string); the `@import './generated/runtime-fonts.css'` assertion reads `tokens.css:1`, unchanged; the no-hex/`rgb(`/`hsl(` scan over `App.css` is unaffected by the new family names; the two exact-count assertions count `--type-*` uses, and line 118 uses none.
- `canvas-authority-contract.test.ts`: 14 prohibitions, **none** naming a font family; its only CSS-shaped rule lists text-layout properties and `font-family` is not among them; its `@media` exact-list assertion over `App.css` is untouched. `scripts/*.mjs` is in **no** scanned corpus.
- `embedded-face-family.test.ts`'s shipped-collision fixture already lists all six family names and asserts the derived family equals none of them — unaffected.
- **Release path:** adding rules that reference the **same three files** emits no new asset. `s1Ids`, `semanticLabels`, `cachedRows.length !== 4`, `assetCount`, the `find('/noto-sans.')`/`/noto-sans-thai.`/`/noto-sans-cjk.` needles and the CJK-dominance check are all unchanged. The CSS bundle's content hash, filename, `sha256`, `.br` sidecar, `release.id` and `release.pageId` **do** change — every one is computed at generate time and recomputed at verify time, never compared to a literal. `release.assets.length < 2` is a floor.
- `lint/MANIFEST.md` has **no** IBM Plex row and is keyed on committed file paths. No binary is added, so no manifest change and `TestManifestUpToDate` stays green.

### Cross-language precedent

`folio-go/canvas_projection_wire_test.go:346` reads `folio-designer/src/engine-protocol.ts` to assert the wire shape declared in Go and in TS agree — a test in one language asserting about a source file in another is **established practice here**. `folio-go/wasm/cmd/engine/main_test.go:229` reads a designer data file. Nothing under `lint/` reads designer *source*.

### Baseline, measured at `475cb50` (not assumed)

- `go test -count=1 ./...` → **1811 pass / 2 fail / 5 skip**; the two are `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest (`got 7, need >=20`).
- `go test -count=1 -tags=matrix ./...` → **1822 pass / 3 fail / 5 skip**; the third is `TestShippedFacesReproduceFromUpstream`, a **could-not-execute, not a byte divergence**: `fontgen: fontTools is not importable by this interpreter (/opt/homebrew/opt/python@3.12/bin/python3.12)`. It never compared bytes (DW-86).
- **No third red.** `go vet -tags=matrix ./...` and `gofmt -l folio-go` both empty. `cd lint && go test -count=1 ./...` → four `ok`.
- All four AD-21 legs pass, **24 documents** hashed each; the unset control passes while asserting nothing (`matrix_test.go:2199`) — a control, never a fifth leg. `TestCrossTargetByteIdentity` passes.
- `shasum -a 256 fixtures/*/expected.pdf` → **23** lines. Root `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`.
- Designer: typecheck clean; oxlint **exactly 4** `only-export-components` warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1323,1330`); Vitest **37 files / 350 tests**; `test:e2e:compile` clean.
- Offline release: node `v24.16.0` (the pinned `RELEASE_RUNTIME`); `npm run build`, `verify:offline:red`, `verify:offline:wasm` all pass. Manifest `s1VisibleBytes` **12,423,974**.

## Tasks & Acceptance

**Execution:**

1. `folio-designer/scripts/build-wasm.mjs` -- extend the `runtime-fonts.css` template literal at `:79` with **three additional `@font-face` rules**, one per shipped face, families spelled exactly `Noto Sans`, `Noto Sans Thai`, `Noto Sans SC`, each interpolating the **same slot** its IBM Plex counterpart already uses (`assets.sans`, `assets.sansThai`, `assets.sansCjk` respectively) -- so the browser can name the face the engine measured, from the same bytes, with no new binary and no new slot. Leave the three existing rules **byte-identical**. Keep each rule's `@font-face { font-family: 'NAME'; src: url('./runtime/${SLOT}') format('truetype'); font-display: swap; }` spelling contiguous and exact; the guards parse it by regex and drift empties them silently.

2. `folio-designer/src/App.css` -- rewrite the `font-family` of the single-line `.canvas-text-fragment` rule at `:118` to `'Noto Sans', 'Noto Sans Thai', 'Noto Sans SC', sans-serif` -- the canvas vocabulary becomes the engine's by identity. Change nothing else on the line, keep it one line beginning at column 0, and touch no other rule and no token. This is file-order preserving: the same three files in the same order, so no rasterization changes today.

3. `folio-designer/src/canvas-font-stack.test.ts` -- **replace** the disjointness test. Keep the `arrayContaining` floor asserting the three IBM Plex families are still declared (chrome is unchanged and must stay registered). Delete the two intersection assertions and put in their place: the engine's face names, **parsed from `folio-go/fonts/fonts.go`'s `Shipped()` map rather than hardcoded**, are all present in `declared`, and the canvas fragment stack `requested` **contains** every one of them. Give the parse a non-vacuity floor (it must yield exactly three names, or fail saying it read nothing). Do **not** assert global disjointness, and do **not** soften the chrome half. Rewrite the test's name and the two now-false prose blocks (the "the shipped half is FALSE" paragraph above the tie, and the DW-35 tripwire calling cause one a decision above a builder's authority) to record what is now true and what remains open.

4. `folio-designer/src/canvas-font-stack.test.ts` -- raise the declaration vacuity floor from `>= 3` to `>= 6` -- six is the true count, and a floor that sits below the real number stops discriminating.

5. `folio-designer/src/font-binary-identity.test.ts` (new) -- build the guard nothing in this repository has. Perform the join the codebase never performs: parse the generator for **both** halves -- the `assets` slot→path map and the `@font-face` family→slot binding -- and compose family→**source file path**. Assert the whole six-entry map with an exact `toEqual`, so any added, removed or repointed rule reddens. Then pin the interval: for each of the three pairs, assert the IBM Plex family and the engine family resolve to **one and the same path**, with **Story 8.4c named in the assertion's own failure message** as the successor that makes them diverge. Include a red-proof over fixture generator text -- a pair bound to differing paths must make the same checker report a mismatch -- and a non-vacuity floor on both parsed halves. Named to match the file Story 8.4c's `ready-for-dev` spec already plans to extend, so its change is an edit, not a creation.

6. `folio-designer/src/font-binary-identity.test.ts` -- add the file-swap tie: for each face name in `fonts.Shipped()`, the file the generator binds that **engine-named** family to is **byte-identical** to the file `fonts.go` embeds under that same face name, and the two name sets are equal. Scope it to the engine-named half only -- that is where the claim is true, it stays true across Story 8.4c, and it is what makes "the face the engine measured" a checked fact rather than an assumption. Say in the test why the IBM Plex half is out of scope, citing 8.4c.

**Acceptance Criteria:**

- **Given** the shipped faces, **when** the designer's stylesheet is generated, **then** the generator declares six `@font-face` rules over three files, and each of `Noto Sans`, `Noto Sans Thai`, `Noto Sans SC` is declared with the `src` of the file `folio-go/fonts/fonts.go` embeds under that same face name.
- **Given** the canvas text fragment rule, **when** its `font-family` is read, **then** it names the three engine face names in the engine's own order followed by `sans-serif`, every name is one the generator declares an `@font-face` for, and no name is supplied by a `var()` or any document input.
- **Given** every chrome token — `--font-sans`, `--font-mono`, `--font-page`, every `--type-*` — **when** the story is complete, **then** none is edited, the three IBM Plex families are still declared, and `tokens.css`, `design-tokens.ts`, `DESIGN.md`, `ARCHITECTURE-SPINE.md` and the licence manifest carry no diff.
- **Given** the disjointness assertion, **when** it is updated, **then** its two intersection assertions are gone, its replacement asserts the fragment stack's families **contain** the engine's face names, the engine's names are read from `fonts.Shipped()` rather than a hardcoded array, and the chrome half is neither weakened nor removed.
- **Given** the interval before Story 8.4c, **when** an engine family and its IBM Plex counterpart are resolved through the generator's two halves, **then** they are asserted to be one and the same file, and the failure message names Story 8.4c as the successor that makes them diverge.
- **Given** the family→file guard, **when** any single assertion in it is **deleted**, or a family is bound to a different file, **then** a named test fails — proved by running each mutation, not argued.
- **Given** the whole change, **when** the corpus is rendered, **then** all 23 golden digests are unchanged and the four AD-21 legs still agree — this story emits no PDF byte.

## Spec Change Log

## Review Triage Log

## Design Notes

**Why the family is spelled in the generator and the tie is machine-checked.** The generator is a build script that writes CSS text; it cannot call Go. So the three engine names are literals there, exactly as the three IBM Plex names already are. What makes that safe is not the literal but the **tie**: Task 6 reads `fonts.Shipped()`'s own map and requires the generator's engine-named set to equal it, byte-for-byte on the files. That is one authority (`fonts.go`) with a checked mirror — **not** a mapping table, which would be a second authority asserting *which browser family corresponds to which engine face* and would have to be maintained in lockstep with the `FontSet`. The distinction is the whole verdict: here the browser family **is** the engine's name, so there is nothing to map.

**The hardcoded `engineFaces` array is why the replacement must not be a rename.** The assertion being replaced compares a dynamically-read `declared` against a literal `['Noto Sans','Noto Sans Thai','Noto Sans SC']`. Story 8.4c's spec flags it for whoever owns this story: if `folio-go` ever ships a different set, the engine half does not follow and the assertion keeps passing while having become false. Its successor must therefore not reintroduce the same hardcode — which is why the engine's names are parsed from `fonts.go` in both Task 3 and Task 6.

**Weakening by evasion, and what already closes it.** `requested` is read from `App.css` alone, so a future change that moved the fragment family into an inline `style` prop would leave this guard green while asserting nothing about the family actually asked for. That route is **already closed**, by the assertion that the designer's TypeScript contains exactly one `font-family` position and it is `App.tsx: embeddedFaceFamily(fragment.assetKey)`. The two guards are only jointly sufficient — neither alone is — so neither may be narrowed without the other.

**What this story does NOT close, stated rather than discovered later.** DW-35 cause one has two layers. This story closes the **vocabulary** layer: the browser can now name the engine's faces at all, which was the stated blocker and the ruled subject. The second layer remains open: the fragment stack is a **fixed constant**, not the document's chain, and a shipped-face fragment carries **no face identity** on the wire (only carried faces carry an `assetKey`; a shipped fragment paints with no inline family and falls to the stack). Since Epic 8 makes chains authorable, a document may declare `["Noto Sans Thai"]`, and the faces' coverage genuinely overlaps — measured at `475cb50` with fontTools: all three cover `A` and `5`; the pairwise cmap overlaps are 339 / 529 / 230 codepoints. So the engine can measure a Latin run with `Noto Sans Thai` while the browser's fixed Latin-first stack rasterizes it with `Noto Sans`. Closing that needs per-fragment shipped-face attribution on the wire — the shape 8.4a built for carried faces — which is a different story and is **not** in these ACs. **Recommendation: DW-35 cause one should be narrowed to this residual, not closed.**

**The limit these gates cannot pass.** Nothing in this repository executes a real font load, a `document.fonts.add`, or a rasterized glyph: `test:e2e:compile` is `tsc --noEmit`, Playwright is in no workflow, browser e2e is deferred by D-000.4, and jsdom applies no stylesheet and implements no font loading. So every AC here is written at a layer the suite observes — a parse of a tracked source, a filesystem byte comparison, a jsdom render. What is proved is that the browser is **asked** for the engine's face names and that a face under each of those names is declared from the engine's own bytes. That the glyphs are actually **rasterized** with those faces is unverifiable here and becomes checkable **when browser e2e arrives (D-000.4)**. Do not report a compile pass as a run.

**Coordination with Story 8.4c, whose `ready-for-dev` spec predates the reversal.** That spec was written assuming it ran first, and two of its statements are now stale: its "Never add a Noto engine face name to `declared` — that is Story 8.4b's territory" and its planned disclosure that `assets.sansCjk` is present with no `@font-face` naming it (after this story, the `Noto Sans SC` rule names it). Its Task 4 creates `font-binary-identity.test.ts`; after this story that file exists and its change becomes an edit. None of this is 8.4b's to amend — it is recorded here so the orchestrator can route it.

## Verification

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expect **exactly** the baseline reds: `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`. Any third red is a real failure. Report measured pass/fail/skip counts, not "green".
- `cd folio-go && go test -count=1 -tags=matrix ./...` -- expect those plus `TestShippedFacesReproduceFromUpstream`, which must be reported as a **could-not-execute** (`fontgen: fontTools is not importable`), never as a byte divergence (DW-86).
- `cd folio-go && go vet -tags=matrix ./...` -- expect empty.
- `cd /Users/panitw/Projects/folio && gofmt -l folio-go` -- expect empty.
- `cd folio-go && FOLIO_MATRIX_TARGET=<t> go test -count=1 -tags=matrix -run TestTargetRenderHash -v .` for each of `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` -- expect four passes, 24 documents hashed each. Run once unset as the control and report it **as a control**, not a fifth leg.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` -- expect pass.
- `cd lint && go test -count=1 ./...` -- **`-count=1` is mandatory**; the rules package walks a directory and Go's cache does not track `ReadDir`, so a cached `ok` is no measurement. Expect four `ok`.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expect typecheck clean; oxlint **exactly 4** `only-export-components` warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1323,1330`) and no errors; Vitest **at least 37 files / 350 tests** with the new tests added and all passing.
- `cd folio-designer && npm run test:e2e:compile` -- expect clean. **This compiles; it does not run.**
- `cd /Users/panitw/Projects/folio && shasum -a 256 fixtures/*/expected.pdf` -- expect **23** lines, every digest identical to the Code Map baseline.
- `cd folio-designer && npm run build && npm run verify:offline:red && npm run verify:offline:wasm` -- required because the generated stylesheet changes. Needs node exactly `v24.16.0`. Expect all three to pass and the emitted `.ttf` set to be the **same three files**.

**Manual checks:**
- Mutation-proof each replaced and each new assertion by **deleting it** and re-running: a guard that survives its own deletion is decoration. Record which named test failed for each deletion. Do this per assertion, not per file.
- Red-proof the family→file map by binding one family to the other pair's path in the fixture text and confirming the checker reports a mismatch. A checker that only ever passes has not been shown to discriminate.
- Confirm `git diff` touches **no** `tokens.css`, `design-tokens.ts`, `DESIGN.md`, `ARCHITECTURE-SPINE.md`, `epics.md` licence manifest, `lint/MANIFEST.md`, or any file under `fixtures/`.
- Confirm the three pre-existing IBM Plex `@font-face` rules are **byte-identical** in the diff, and that `assets` gained no key.
- Confirm root `README.md` md5 is still `078d7d80d518d54af2fc04fb270d46b8` and that no attestation record (`fixtures/statement-signoff.json`, `fixtures/thai-stacked-marks/signoff.json`, `fixtures/embedded-font/signoff.json`) is touched.
- Stage explicit paths only. Never `git add -A` or `git add .`. Never push, never branch.

## Auto Run Result

### 2026-09-01 — Plan-gate dispatch (terminal)

Status: ready-for-dev
Blocking condition: none

Dispatched classic-intent at HEAD `475cb50`, tree clean, with the directive `Halt after planning.`
The spec is written and no code was produced. Epic context `epic-8-context.md` was stale (the epics
file is newer) and was recompiled by the workflow's own step-1 subagent; that is this dispatch's only
tracked modification.

**Premises checked, all TRUE.** The generator's resolution — `IBM Plex Sans`→`NotoSans-Regular`,
`IBM Plex Sans Thai`→`NotoSansThai-Regular`, `IBM Plex Mono`→`NotoSansSC-Regular` — is exactly the
engine's three faces in the engine's own order, so the canvas is accidentally correct for every script
today and the rewrite is order- and file-preserving. The designer/engine font mirror is byte-identical
on all three faces. No chrome token needs editing. No mapping table is required.

**Two forks resolved rather than halted, with the selecting ground recorded.** (1) Whether the story
also makes the fragment stack follow the document's chain — selected against by the AC's own wording
and by the preserved "stylesheet constant with no document input" guard, which such a change would
redden; recorded instead as the open residual in Design Notes. (2) Where the family→file guard lives —
settled by Story 8.4c's own `ready-for-dev` spec, which already names `font-binary-identity.test.ts`
and records why it is a designer test rather than a `lint` rule.

**Baseline measured at `475cb50`, not assumed:** exactly two standing reds
(`TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`; and, under `-tags=matrix`,
`TestShippedFacesReproduceFromUpstream` as a could-not-execute). 23 golden digests. Vitest 37/350.
Full figures in the Code Map.
