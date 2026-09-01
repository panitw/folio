---
title: 'Story 8.4c: The designer ships the typeface it specifies'
type: 'feature'
created: '2026-09-01'
status: 'done'
baseline_revision: 'a4bac02f4a162d751c33ca585595ea44a8b23cf1'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized', 'multiple-goals']
deferred: []
---

## Frontmatter warnings — what they mean here

**`oversized`.** Every folio story spec exceeds 1600 tokens. This one additionally carries a measured
procurement record — upstream release pins, artifact digests, and the Thai coverage evidence — that
the implementer cannot re-derive from the tree, so trimming it would cost more than it saves.

**`multiple-goals` — carried for honesty, NOT as an open question.** AC1 (the mono binary) is
genuinely independently shippable, and the engineering lead said so in as many words: it is *"the
cheapest end-to-end proof the IBM Plex pipeline works — one binary, one generated rule, one licence
entry."* The warning records that fact rather than raising it. **D-8.4.17(c) already refused the
split**, because splitting would put the licence-manifest edit in a different unit from the binaries
it describes. The adjudicated shape is **one story, two commits, AC1 first**. Do not split, do not
re-open, and do not route this warning to the lead as an unresolved fork.

## In plain terms (read this first if you just want the gist)

The designer's interface says it is drawn in IBM Plex. It is not. Three stylesheet rules give the
IBM Plex names to three Noto font files, and the repository contains no IBM Plex byte at all. The
worst of the three is the "mono" slot: the family called `IBM Plex Mono` is really Noto Sans SC, a
10.6 MB Chinese sans with no monospacing whatsoever, so every number, tab label and brand mark in
the chrome is drawn in a CJK face.

The owner decided to make the words true rather than change them: ship real IBM Plex. This story
adds three genuine IBM Plex binaries with their licences, points the three stylesheet rules at them,
and writes the first test in this repository that can tell whether the bytes behind a font name are
the font that name claims. The mono fix lands first, on its own, because the owner asked for it now
and because one binary is the cheapest possible proof the whole pipeline works.

Nothing about the document you are designing changes. The engine still measures with Noto, the 23
golden PDFs are untouched, and no `.folio` file moves a byte.

<intent-contract>

## Intent

**Problem:** The product **specifies** IBM Plex throughout its design source, **promises** IBM Plex
in the release licence manifest, and **ships no IBM Plex file at all** — `find -iname '*plex*'`
returns nothing, while three generated `@font-face` rules put IBM Plex family names over three Noto
files. The `IBM Plex Mono` family is bound to `noto-sans-cjk.ttf` (Noto Sans SC, a CJK sans, not a
monospace face), so `--type-mono`, `--type-mono-em`, `--type-numeric-lg`, `--type-brand`,
`--type-brand-load`, `--type-band-tab` and `--type-page-mono` all render chrome in a CJK face.

**Approach:** Ship real IBM Plex OFL binaries for the browser chrome (owner decision D-8.4.15), each
with its own licence text and NOTICE, and repoint the three generated `@font-face` rules at them. The
three chrome family NAMES do not change; only the files behind them. The mono binary lands as its own
commit, first (D-8.4.15, D-8.4.17(c)). The three Noto binaries stay in the bundle untouched — Story
8.4b needs them, NFR7 requires CJK coverage in the shipped set, and the release verifier requires
them by name.

## Boundaries & Constraints

**Always:**
- **`declared` stays exactly `['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai']`.** Three
  `@font-face` rules, three unchanged family names. Only the `src:` targets move.
- **All three Noto binaries stay in the bundle**, still fingerprinted, still copied, still present in
  `offline-assets.ts` and the release manifest. They lose only their `@font-face` declaration.
- **Font binaries are committed `.ttf`** — never `.woff`, `.woff2`, `.eot` or a variable font.
  `lint`'s asset licence scan recognises only `.ttf`/`.otf`/`.ttc`; a `.woff2` would ship with AD-26's
  asset half silently not applying.
- **Every committed font binary sits in its own directory beside a `LICENSE*` and a `NOTICE*` file**,
  the NOTICE carrying a line containing `Copyright`. This is enforced, not stylistic:
  `manifest.ResolveAssets` fails the build without both.
- **`lint/MANIFEST.md` is regenerated in the same commit as any binary addition** (`cd lint && go run
  ./cmd/genmanifest`), or `TestManifestUpToDate` reddens.
- **`build-wasm.mjs`'s `runtime-fonts.css` template literal stays ONE line, with its exact existing
  spacing.** `canvas-font-stack.test.ts` parses that line with the regex
  `/@font-face \{ font-family: '([^']+)'/`; reformatting it silently empties `declared`.
- **The generated stylesheet keeps `url('./runtime/…') format('truetype')`.** A `data:` or `blob:`
  source reddens `canvas-font-stack.test.ts`'s runtime-registration detector.
- Commit trailer on every commit: `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`.

**Block If:**
- The upstream IBM Plex face fetched does not match the recorded sha256 in this spec's Code Map, or
  its `name` table does not declare the family this spec says it does.
- `licence.ClassifyLicenceText` does not return the SPDX id `OFL-1.1` for a committed IBM Plex licence
  text, so a `lint/MANIFEST.md` row would read `SEE NOTICE` instead of `OFL-1.1`. The manifest must be
  true in the manifest's own terms; a fallback label is not "true".
- Any golden PDF digest moves, or `shasum -a 256 fixtures/*/expected.pdf` returns other than 23 lines.
- A third distinct test failure appears beyond the two standing reds recorded in Verification.
- The release verifier's `S1 CJK row is not the dominant font payload` assertion fails — that means a
  Noto binary was removed or replaced, which this story forbids.

**Never:**
- **Never touch `folio-go/**`.** Not `fonts/fonts.go`, not `Shipped()`, not `tools/fontgen/`, not
  `instance_faces.py`'s `UPSTREAM` table, not `shippedFaceSpecs`. IBM Plex is a **chrome** face and
  never enters the engine's `FontSet`. `TestShippedFacesReproduceFromUpstream`'s hardcoded
  `"derived and compared 3 of 3 faces"` stays 3.
- **Never add a Noto engine face name to `declared`.** That is Story 8.4b's territory, ruled on AD-17
  grounds by D-8.4.14. Do not register any shipped face under the engine's name here.

> **⚠ STALE AFTER STORY 8.4b — CORRECTED AT THE GATE, 2026-09-01 (D-8.4.22).** This spec was planned
> while 8.4c was sequenced **before** 8.4b. The lead **reversed** that at `6f0c095`: the order is now
> `8.4c` **after** `8.4b`. Three passages assumed the old order and are corrected in place, kept
> verbatim with their correction attached rather than rewritten. **Corrections at the plan gate are
> legal and cheap; the same corrections after implementation are neither.**

> **THIS CONSTRAINT IS INVERTED.** Story 8.4b has **already landed** (`90cdf8e`, closed at `5ae02d7`)
> and **added the three Noto engine face names to `declared`** — that was its whole subject. So
> `declared` now holds **six** families, not three, and the offline release emits **6 rules over the
> same 3 `.ttf` files**. **The live constraint is: do not REMOVE the engine-named half, and do not
> repoint it.** This story changes only the files behind the **three IBM Plex** names.

- **Never modify any `.folio` document, fixture, golden, or attestation record.** The corpus is 23
  documents and none may move. A moved PDF digest is a defect, not a re-record. No agent writes
  `reader`, `date` or `examined` into any `signoff.json`.
- **Never modify, move, delete or stage root `README.md`** (md5 `078d7d80d518d54af2fc04fb270d46b8`).
- **Never remove or replace a Noto binary**, and never delete the now-undeclared `sansCjk` entry from
  `build-wasm.mjs`'s `assets` object.
- **Never edit a planning artifact** — not `epics.md`, not `ARCHITECTURE-SPINE.md`, not `DESIGN.md`.
  All three already assert what this story makes true; none needs a word changed.
- **Never restructure `build-wasm.mjs`'s three-rule template literal into a loop**, and never expand
  to a per-weight or per-style face set. One static Regular per family, matching the Noto precedent
  and the current generator shape.
- Never push. Never create a branch. Never `git add -A` or `git add .` — explicit paths only.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Mono family resolved | `runtime-fonts.css` rule for `IBM Plex Mono` | `src` names the committed `IBMPlexMono-Regular.ttf`; that file's `name` table family is `IBM Plex Mono` and `post.isFixedPitch` is non-zero | No error expected |
| Sans family resolved | rule for `IBM Plex Sans` | `src` names `IBMPlexSans-Regular.ttf`; name-table family is `IBM Plex Sans` | No error expected |
| Thai family resolved | rule for `IBM Plex Sans Thai` | `src` names `IBMPlexSansThai-Regular.ttf`; name-table family is `IBM Plex Sans Thai`; its cmap maps every codepoint of `พระราชบัญญัติ` and `การทวงถามหนี้` | No error expected |
| A family is repointed back at a Noto file | generator edited so `IBM Plex Sans` names `noto-sans.ttf` | The new identity guard fails, naming the declared family and the family the file actually carries | Test failure, red |
| Declaration set changes size or names | a fourth `@font-face`, or a renamed family | `declared` no longer equals the three IBM Plex names; the identity guard has no source file for the new family and fails | Test failure, red |
| ~~CJK binary still shipped, no longer declared~~ **ROW CORRECTED AT THE GATE (D-8.4.22): `assets.sansCjk` IS declared after 8.4b, under the engine name `Noto Sans SC`.** The row's premise — absent from `runtime-fonts.css` — is false. | `assets.sansCjk` present **and declared** under `Noto Sans SC` | Still fingerprinted, copied and emitted; the `/noto-sans-cjk.` lookup still resolves; the `cjk-font` S1 row is still the dominant font payload — **and it now also backs a declared `@font-face`, so removing its declaration is a regression rather than a tidy-up** | No error expected |
| CJK binary deleted or replaced | `assets.sansCjk` removed | `generate-offline-release.mjs` throws `production build has no /noto-sans-cjk. runtime asset`; or the verifier fails `S1 CJK row is not the dominant font payload` | Build failure, halt |
| Font binary committed without licence artifacts | a `.ttf` in a directory with no `LICENSE*`/`NOTICE*` | `manifest.ResolveAssets` returns `<dir>: contains a committed font binary but no LICENSE* file (AC25, AD-26)` | Build failure, red |
| `lint/MANIFEST.md` not regenerated | binary added, manifest stale | `TestManifestUpToDate` fails with `lint/MANIFEST.md is out of date — run cd lint && go run ./cmd/genmanifest` | Test failure, red |

</intent-contract>

## Code Map

**Every anchor re-measured at `e66e8b3`, this dispatch's HEAD.** The plan gate and the build share
this commit, so there is no spec-baseline/HEAD divergence to reconcile — but re-verify before editing
(see [`spec-code-map-vs-dispatch-head`] discipline: anchors rot across story boundaries, so cite by
what a thing asserts, not by line, per D-8.4.13).

### The generator — the entire mutable surface

- `folio-designer/scripts/build-wasm.mjs` is **79 lines**. Two disjoint halves, and this matters:
  - **`:64-71` the `assets` object.** Keys are opaque slot names carrying **no family name**:
    ```js
    sans:     fingerprint(join(designerRoot,'public','fonts','notosans',    'NotoSans-Regular.ttf'),     'noto-sans.ttf'),
    sansCjk:  fingerprint(join(designerRoot,'public','fonts','notosanssc',  'NotoSansSC-Regular.ttf'),   'noto-sans-cjk.ttf'),
    sansThai: fingerprint(join(designerRoot,'public','fonts','notosansthai','NotoSansThai-Regular.ttf'), 'noto-sans-thai.ttf'),
    ```
    The **second argument is the emitted label**, and the emitted filename is
    `<stem>.<first 20 hex of sha256><ext>` (`fingerprint`, `:45-53`, which also `copyFileSync`s into
    `src/generated/runtime/`).
  - **`:79` the single-line template literal** that writes `src/generated/runtime-fonts.css` —
    three `@font-face` rules, family names hardcoded, keyed to `assets.sans` / `assets.sansCjk` /
    `assets.sansThai` **by hand**. There is no loop and no map. Each rule carries only
    `font-display: swap` — **no `font-weight`, no `font-style`, no `unicode-range`**, so every weight
    and italic in the chrome is browser-synthesised from one Regular face today and will remain so.
  - `:76-77` writes `offline-assets.ts` by iterating `Object.entries(assets)` **generically** — so a
    new `assets` key gets its `import …?url` and its `runtimeAssetUrls` member for free, which is what
    puts a new binary into Vite's asset graph and thence into the release manifest.
  - `:15-16` `rmSync(outputDir)` then recreate — the generated runtime tree is rebuilt every run.
- `folio-designer/src/generated/runtime-fonts.css` and `src/generated/runtime/` are **gitignored**
  (`.gitignore:68,70`), which is why the guards read the **generator**, not its output.

### The binaries as they stand

| Path | Bytes | Real identity |
|---|---|---|
| `folio-designer/public/fonts/notosans/NotoSans-Regular.ttf` | 646,160 | Noto Sans |
| `folio-designer/public/fonts/notosanssc/NotoSansSC-Regular.ttf` | 10,595,932 | **Noto Sans SC — the file behind `IBM Plex Mono`** |
| `folio-designer/public/fonts/notosansthai/NotoSansThai-Regular.ttf` | 47,788 | Noto Sans Thai |

Each directory already carries `LICENSE-OFL.txt` + `NOTICE.md`. `folio-go/fonts/` holds a
**byte-identical second copy** of all three (sha256 verified pairwise) with **nothing enforcing the
mirror** — `tools/fontgen/instance_faces.py:329` writes only into `folio-go/fonts/`. Out of scope
here; recorded because a reader will wonder.

### What the guards actually assert — `folio-designer/src/canvas-font-stack.test.ts` (443 lines)

- `:34-35` reads `scripts/build-wasm.mjs` and `src/App.css` **as tracked source text** (rationale at
  `:28-32`: the generated CSS is gitignored, and asserting against a file that may not exist is how a
  guard goes quietly vacuous).
- `:43-45` `declaredFamilies` — `[...generator.matchAll(/@font-face \{ font-family: '([^']+)'/g)]`.
  **This is why `:79` must stay one line with its exact spacing.**
- `:48-54` `requestedFamilies` — finds the single `App.css` line starting `.canvas-text-fragment {`
  and extracts quoted families in order. It **throws** if that rule is reformatted.
- `:301-308` the **engine/browser disjointness** assertion. `declared` is read dynamically; but
  `:304 engineFaces = ['Noto Sans','Noto Sans Thai','Noto Sans SC']` is **HARDCODED**, and `:305`
  pins the three IBM Plex names as an `arrayContaining` floor.
- **MEASURED, and it is the central verification fact of this story: NOT ONE ASSERTION IN THIS FILE
  MOVES UNDER A FILE-SWAP.** No assertion anywhere reads an `@font-face` `src`, a filename, a path, a
  hash, or a font byte. The dispatch's premise — that `declared` stays the three names and the
  disjointness assertion stays true throughout — is **CONFIRMED**, and its corollary is that **the
  repository currently has no way to observe this story succeeding or failing.** That net must be
  built (Task 3).
- Two tripwires to route around, both measured:
  - `:338` `registersAFaceAtRuntime(withoutComments(generator))` must stay `false`; its regex
    `/@font-face[\s\S]{0,400}?src\s*:[^;}]{0,200}?(?:data:|blob:)/` **reddens if the generator is
    changed to emit base64 `data:` sources.** Keep `url('./runtime/…')`.
  - `:409` `expect(positions).toEqual(['App.tsx: embeddedFaceFamily(fragment.assetKey)'])` — an
    **exact-equality census** of every `fontFamily:`/`font-family:` position in designer TypeScript.
    Adding any second one reddens it. This story adds none.
  - `:179`, `:331` `declared.length >= 3` and `:183`, `:298` `requested.length >= 3` are floors, not
    equalities — they tolerate the swap and would tolerate additions.

### The CSS that needs no edit — confirming the "names do not change" premise

- `folio-designer/src/tokens.css:11` — all three stacks on one line:
  `--font-sans: 'IBM Plex Sans', system-ui, sans-serif;`
  `--font-mono: 'IBM Plex Mono', ui-monospace, monospace;`
  `--font-page: 'IBM Plex Sans Thai', 'IBM Plex Sans', sans-serif;`
- `tokens.css:16` — the seven tokens resolving through `--font-mono` are `--type-mono`,
  `--type-mono-em`, `--type-numeric-lg`, `--type-brand`, `--type-brand-load`, `--type-band-tab` and
  **`--type-page-mono`**. The dispatch named four families of token; `--type-page-mono` is a seventh
  member it did not list. All seven are today drawn in Noto Sans SC.
- `folio-designer/src/App.css:118` — the **only** `font-family` declaration in the file:
  `.canvas-text-fragment { … font-family: 'IBM Plex Sans', 'IBM Plex Sans Thai', 'IBM Plex Mono', sans-serif; }`
- **Neither file is edited by this story.** Because the family names are unchanged, the whole CSS
  surface is untouched — which is the premise D-8.4.17(c) rested the sequencing on, now confirmed.

### The licence machinery — measured reach, not assumed

- **AD-26 Rule** (`…/ARCHITECTURE-SPINE.md:512-518`) already says, verbatim: *"the shipped Noto and
  IBM Plex faces are **SIL OFL 1.1** and travel with their licence text and copyright lines."* Its
  Stack table `:553` already reads `| Designer fonts — IBM Plex Sans / Mono / Sans Thai | SIL OFL 1.1 |`.
  `epics.md:206` already promises the same. **All three statements become TRUE when the binaries
  land. Not one word of any planning artifact needs changing** — AC3 is satisfied by presence, not by
  an edit.
- `lint/internal/manifest/manifest.go:165` `ResolveAssets` — **walks the whole repository**
  (`filepath.WalkDir`, `:169`), not an allowlist. `:110` `fontExtensions = []string{".ttf",".otf",".ttc"}`.
  Excludes `.git` and `*/testdata/lint/**`; skips any directory with `gitTrackedFileCount == 0`
  (`:247-249`), which is why gitignored trees are invisible.
- Per candidate directory (`:262-296`), **non-recursive `os.ReadDir`**: a file whose name starts
  `LICENSE` and one starting `NOTICE` are both required, else a hard error naming AC25/AD-26.
  `extractCopyrightLine` (`:399-406`) accepts any line **containing** `Copyright`, trimming `*` — so
  the Noto NOTICE's `**Copyright 2022 …**` form works, and so will `**Copyright © 2017 IBM Corp. …**`.
- `:298-301` the licence label: `licenceLabel := "SEE NOTICE"`, replaced by the SPDX id **only if**
  `licence.ClassifyLicenceText` returns one. **This is the soft-failure to watch** — a licence text
  that does not classify does not fail the build, it just labels the row `SEE NOTICE`. The manifest
  would then be present but not true in its own terms, so this spec makes it a Block If.
- `licence.ClassifyLicenceText` returns `("permissive", "OFL-1.1")` for text carrying the OFL name
  **and** `Version 1.1` (`classify.go:139-162`); `TestCommittedOFLTextClassifiesAsOFL11`
  (`classify_test.go:128`) pins that against `folio-go/fonts/notosans/LICENSE-OFL.txt` and **only**
  that file.
- `lint/MANIFEST.md`'s `## Redistributed non-code assets` table, schema
  `| Path | Licence | Copyright | Serves |`. Designer paths get the `default:` label
  `"committed asset (<relDir>)"` (`assetServesLabel`, `:123-134`). Regenerate with
  `cd lint && go run ./cmd/genmanifest`; `TestManifestUpToDate` (`manifest_test.go:41`) byte-compares.
- **`lint` has no font-parsing dependency** — `lint/go.mod` requires only `golang.org/x/tools`
  (+ `x/mod`, `x/sync` indirect). `rules/fontsassets.go`'s `looksLikeSfnt` is a hand-rolled magic
  check, and `ScanFontsAssets` is scoped to `fontsAssetLocation = "folio-go/fonts"` — it does not and
  must not reach the designer.

### The release path — two hard couplings, both measured

- `folio-designer/scripts/generate-offline-release.mjs:40-53`:
  `const find = (needle) => { … if (!asset) throw new Error(\`production build has no ${needle} runtime asset\`) }`
  called with `'.wasm'`, `'/noto-sans.'`, `'/noto-sans-thai.'`, `'/noto-sans-cjk.'`. **All three Noto
  needles must keep resolving**, which they do as long as the three `assets` entries survive — the
  `@font-face` declaration is irrelevant to them. New Plex labels (`ibm-plex-*.ttf`) collide with no
  needle.
- `folio-designer/scripts/verify-offline-release.mjs:58-61`: `s1Ids = ['engine','latin-font',
  'thai-font','cjk-font','thai-dictionary']` and `semanticLabels = ['Engine','Latin font','Thai font',
  'CJK font','Thai dictionary']` — **exact ordered equality on both**, plus `cachedRows.length !== 4`.
- `verify-offline-release.mjs:78-79`: the `cjk-font` row's bytes must equal
  `Math.max(...)` over all rows whose id ends `font`. Noto Sans SC is ~4.95 MB Brotli against
  ~100-200 KB for any Plex face, so **this holds as long as the CJK binary is not replaced** — and
  fails immediately if it is. This is the mechanical reason Fork A is not a free choice.
- Neither script asserts anything about licences or about a size budget.

### Upstream procurement — MEASURED 2026-09-01, not recalled

**The npm packages are a trap and must not be used.** `@ibm/plex-sans@1.1.0`,
`@ibm/plex-mono@2.5.0` and `@ibm/plex-sans-thai@1.1.0` all declare `"license": "OFL-1.1"`, but every
one ships **zero `.ttf`/`.otf`/`.ttc`** — only `.woff`/`.woff2`. A `.woff2` is invisible to
`fontExtensions`, so AD-26's asset half would silently not apply.

**Use the GitHub release zips**, which carry `fonts/complete/ttf/`:

| Family | Release asset URL | Path inside zip | Bytes |
|---|---|---|---|
| IBM Plex Sans | `https://github.com/IBM/plex/releases/download/%40ibm/plex-sans%401.1.0/ibm-plex-sans.zip` | `ibm-plex-sans/fonts/complete/ttf/IBMPlexSans-Regular.ttf` | 200,500 |
| IBM Plex Mono | `https://github.com/IBM/plex/releases/download/%40ibm/plex-mono%402.5.0/ibm-plex-mono.zip` | `ibm-plex-mono/fonts/complete/ttf/IBMPlexMono-Regular.ttf` | 173,052 |
| IBM Plex Sans Thai | `https://github.com/IBM/plex/releases/download/%40ibm/plex-sans-thai%401.1.0/ibm-plex-sans-thai.zip` | `ibm-plex-sans-thai/fonts/complete/ttf/IBMPlexSansThai-Regular.ttf` | 116,728 |

**The recorded digests the Block If refers to.** All six measured 2026-09-01 from the fetched
artifacts. The **face digest is the one that gates**; the zip digest is recorded so a re-fetch can be
shown to have got the same release asset.

| Artifact | sha256 |
|---|---|
| `IBMPlexSans-Regular.ttf` | `975dcda37d80f038dcd143c22e33ca2d97a0cc5a929aace1c749153b0fe1afa5` |
| `IBMPlexMono-Regular.ttf` | `7c6fbddca4b700be918f5f6183d9bd4464fa427fe435f0b480d77fe2bb8c5a43` |
| `IBMPlexSansThai-Regular.ttf` | `83e1db8e8bad06bb760981f1dd528f5f209d20dfadebac12b21bf2f12453c8c6` |
| `ibm-plex-sans.zip` (9,921,777 B) | `fb365d910566e6d199cc2c15579a7dd9a267128e18431a394ed81f1970c69200` |
| `ibm-plex-mono.zip` (6,940,652 B) | `6d23f01257663d8cc49a0d64c22ced630b79e0e2a0ac08a0da86e9a38bbc481c` |
| `ibm-plex-sans-thai.zip` (1,985,960 B) | `d7203f43c20f9abd40487f845c48db4077d2056ea18632c8959591c6815d7fb9` |

Because upstream publishes **static** TTF, the committed file is byte-identical to the extracted
file: each NOTICE's "upstream sha256" and "committed sha256" are the **same** value, and that
identity is itself the provenance record — there is no derivation step to replay.

Each zip carries a top-level `LICENSE.txt` (4,456 B) opening
`Copyright © 2017 IBM Corp. with Reserved Font Name "Plex"`, and each `ttf/` directory its own
`license.txt` (4,360 B). All three faces measured: **`fvar` absent (true static instances)**, `glyf`
outlines with no CFF (satisfying NFR7's "glyf/TrueType static build over CFF/OpenType"), and nameID
13 = `This Font Software is licensed under the SIL Open Font License, Version 1.1. This license is
available with a FAQ at: http://scripts.sil.org/OFL`, nameID 14 = `http://scripts.sil.org/OFL`.

### AC2's acceptance risk — DISCHARGED BY MEASUREMENT, 2026-09-01

The epic writes the Thai-coverage question in as *"the acceptance risk of the chosen option … written
here rather than discovered"*. It was measured with fontTools 4.63.0 against the shipped
`NotoSansThai-Regular.ttf`, and **IBM Plex Sans Thai passes on every axis**:

- **cmap parity is exact.** Both map **87 of 128** codepoints in U+0E00–U+0E7F. The set difference is
  **empty in both directions**. The 41 unmapped are Unicode-unassigned, so both cover 100% of
  assigned Thai.
- **Both defect strings are fully covered.** Every codepoint of `พระราชบัญญัติ` and `การทวงถามหนี้`
  — the strings from the shipped defect recorded at `canvas-font-stack.test.ts:18-26`, where Thai fell
  through to `sans-serif` and *"letters rendered on top of each other"* — is in Plex's cmap.
- **Mark positioning is strictly better, which is the mechanism that defect was about.** Under script
  tag `thai`, Plex reaches **MarkBasePos ×3 and MarkMarkPos ×2** (Noto: ×1 and ×2), plus PairPos ×3,
  ChainContextPos ×2, `ccmp` and `calt`; Noto has `ccmp` and no `calt`. Plex declares langsys `dflt`,
  `PAL`, `SAN`, `THA`; Noto only `dflt`.
- **The one measured difference:** Noto declares a `dist` feature under `thai` that Plex does not.
  Plex compensates with more mark lookups. `dist` is not the mechanism of the recorded defect (mark
  attachment is), so this is recorded rather than treated as a blocker.
- **IBM Plex Mono is genuinely monospaced**, confirmed two ways: `post.isFixedPitch = 1`, and every
  one of 1,149 non-zero advance widths is 600/1000 — exactly one distinct width font-wide.
- **IBM Plex Sans covers chrome Latin fully**: 95/95 ASCII, 96/96 Latin-1 Supplement, 128/128 Latin
  Extended-A — zero gaps across U+0020–U+017F. It is thinner beyond: Latin Ext-B 33 vs Noto's 208,
  Greek 73 vs 121, Cyrillic 192 vs 256. The chrome renders none of those; recorded for completeness.

### Bundle weight — the cost the owner accepted

- Added: 200,500 + 173,052 + 116,728 = **490,280 bytes raw**.
- Existing font weight retained: 646,160 + 10,595,932 + 47,788 = **11,289,880 bytes**.
- New total committed font weight: **11,780,160 bytes**, a **+4.34%** increase.
- Context to record but **not to act on**: NFR7's budget (`epics.md:150`, *"Accepted: ~9 MB first
  load"*) is **prose only — nothing enforces it**, and the release already measures
  `s1VisibleBytes = 12,372,693` Brotli bytes against it. **The budget was already exceeded before this
  story.** Report that up; do not edit the epic.

## Tasks & Acceptance

**Execution — COMMIT 1 (AC1, lands FIRST and alone): give the chrome's mono family a real monospace face.**

1. `folio-designer/public/fonts/ibmplexmono/IBMPlexMono-Regular.ttf` -- add the committed binary,
   fetched from the pinned release asset above and extracted from `fonts/complete/ttf/` --
   the owner asked for the mono defect fixed "now, separately", and one binary is the cheapest
   end-to-end proof the IBM Plex pipeline works.
2. `folio-designer/public/fonts/ibmplexmono/LICENSE-OFL.txt` and `.../NOTICE.md` -- add the
   unmodified upstream OFL text, and a NOTICE modelled on `folio-designer/public/fonts/notosansthai/NOTICE.md`
   carrying `**Copyright © 2017 IBM Corp. with Reserved Font Name "Plex"**`, the upstream project,
   release tag and asset URL, the path inside the zip, the fetch date, the **sha256 of the upstream
   file**, the **sha256 of the committed file**, and the byte size -- `ResolveAssets` fails the build
   without both files, and the Noto NOTICEs are this repo's provenance-record precedent.
   **Say explicitly that no derivation applies**: upstream publishes static TTF, so the committed file
   is byte-identical to the upstream file and the two digests are equal. Do not invent an instancer
   invocation; the Noto NOTICEs have one because those faces are derived and these are not.
3. `folio-designer/scripts/build-wasm.mjs` -- add one `assets` entry
   `mono: fingerprint(join(designerRoot,'public','fonts','ibmplexmono','IBMPlexMono-Regular.ttf'), 'ibm-plex-mono.ttf')`,
   and change **only** the `${assets.sansCjk}` interpolation inside the `'IBM Plex Mono'` rule on the
   single-line `runtime-fonts.css` literal to `${assets.mono}` -- **keep `sansCjk` in `assets`**, keep
   the literal on one line with its exact spacing, keep `url(…) format('truetype')`.
4. `folio-designer/src/font-binary-identity.test.ts` -- **THIS FILE ALREADY EXISTS. THIS TASK IS AN
   EDIT, NOT A CREATION (corrected at the gate, D-8.4.22).** Story 8.4b created it, because the
   reversal made 8.4b the story that first needed the net. **Extend the existing guard; do not
   recreate it, and do not rewrite what 8.4b pinned there** — its assertion that `IBM Plex Sans` and
   `Noto Sans` deliberately resolve to one file, naming *this story* in its own failure message, is
   the assertion **this story is expected to update when the two diverge.** That update is the
   small, obvious edit it was designed to be. Add the identity guard: parse
   `build-wasm.mjs` for both halves (the `assets` slot→path map and the `@font-face` family→slot
   binding on the emitted line), open each source `.ttf` with a small dependency-free TrueType
   `name`-table reader, and assert the file's family (nameID 16 falling back to nameID 1) **equals**
   the CSS family it is declared under. Scope this commit's assertion to `IBM Plex Mono`, and
   additionally assert `post.isFixedPitch != 0` for it. Include a **red-proof**: a fixture generator
   text binding `IBM Plex Mono` to the Noto CJK path must make the same checker report a mismatch --
   an identity guard that only ever passes has not been shown to discriminate.
5. **DELETED AT THE GATE — DO NOT WRITE THIS ASSERTION (corrected 2026-09-01, D-8.4.22). ITS PREMISE
   IS FALSE.** It planned a *disclosure-of-absence*: that `assets.sansCjk` is present while **no
   `@font-face` rule names its interpolation**. **Story 8.4b's new engine-named rules NAME IT**, so
   the absence it would disclose no longer exists. Note the shape: the task even anticipated 8.4b and
   instructed that *"8.4b should delete this assertion rather than edit it"* — under the reversed
   order there is nothing to delete, because it was never written. **An assertion of a negative is
   exactly the kind that survives as a false green when the negative stops holding**, which is why it
   is struck rather than adapted. The superseded text follows, kept for the record only:
   ~~in the same file, add the disclosure-of-absence assertion that `assets.sansCjk` is present in the
   generator while no `@font-face` rule names its interpolation,~~ with a comment recording that this state is deliberate,
   that the binary ships for the engine-name registration **Story 8.4b** adds, and that **8.4b should
   delete this assertion rather than edit it** -- without it a later reader "simplifies" away an
   `assets` entry nothing appears to use, and the release build then throws
   `production build has no /noto-sans-cjk. runtime asset`.
6. `lint/MANIFEST.md` -- regenerate with `cd lint && go run ./cmd/genmanifest` -- `TestManifestUpToDate`
   byte-compares it. **Verify the new row's Licence column reads `OFL-1.1`, not `SEE NOTICE`**; a
   `SEE NOTICE` label means `ClassifyLicenceText` rejected IBM's OFL text and is a Block If.
7. Commit with an explicit path list. Never `git add -A`.

**Execution — COMMIT 2 (FINAL): give the sans and Thai families real IBM Plex faces, and record the cost.**

8. `folio-designer/public/fonts/ibmplexsans/` and `folio-designer/public/fonts/ibmplexsansthai/` --
   add `IBMPlexSans-Regular.ttf` and `IBMPlexSansThai-Regular.ttf` with the same three-file shape and
   the same NOTICE content rules as Task 2.
9. `folio-designer/scripts/build-wasm.mjs` -- add the two remaining `assets` entries with labels
   `ibm-plex-sans.ttf` and `ibm-plex-sans-thai.ttf`, and repoint the `'IBM Plex Sans'` and
   `'IBM Plex Sans Thai'` interpolations. **Keep `sans` and `sansThai` in `assets`.** After this task
   `assets` has six font entries and `runtime-fonts.css` still has exactly three rules.
10. `folio-designer/src/font-binary-identity.test.ts` -- generalise the identity assertion to **every
    family in `declared`**, so all three are tied to their real binaries, and add the Thai coverage
    assertion: the file declared under `IBM Plex Sans Thai` maps every codepoint of `พระราชบัญญัติ`
    and `การทวงถามหนี้` -- these are the exact strings from the recorded defect, so the guard fails on
    the failure that actually happened. Extend the red-proof to cover the generalised form.
11. `lint/MANIFEST.md` -- regenerate again; three IBM Plex rows must now be present, each with
    `OFL-1.1` and the IBM copyright line. **This is the final commit's manifest update** — the release
    licence manifest's IBM Plex claim, and AD-26's and `epics.md`'s standing statements, all become
    true here and nowhere earlier.
12. Record the measured weight: the three added byte counts, their total (**490,280**), the retained
    Noto total (**11,289,880**), the new total (**11,780,160**), and the release build's
    `s1VisibleBytes` before and after, against NFR7's `~9 MB first load`. Put the numbers in the commit
    message and in this spec's `## Auto Run Result`. **Do not edit `epics.md`.**
13. Commit with an explicit path list.

**Acceptance Criteria:**

- Given the family named `IBM Plex Mono`, when its generated `@font-face` source is read, then it
  names a committed file whose `name` table declares the family `IBM Plex Mono` and whose
  `post.isFixedPitch` is non-zero — and this is asserted by a test that has been shown to go red when
  the family is bound back to `noto-sans-cjk.ttf`.
- Given the families `IBM Plex Sans` and `IBM Plex Sans Thai`, when the stylesheet is generated, then
  each names a committed IBM Plex OFL binary carrying the OFL nameID 13 string, and the Thai face
  maps every codepoint of `พระราชบัญญัติ` and `การทวงถามหนี้`.
- Given the release licence manifest, when `lint/MANIFEST.md` is regenerated, then it carries one row
  per IBM Plex binary with licence `OFL-1.1` and the IBM copyright line, `TestManifestUpToDate`
  passes, and `cd lint && go test -count=1 ./...` is green.
- Given the bundle size budget, when the IBM Plex binaries are added, then the added weight is
  measured and recorded in the final commit message and in this spec, stated against NFR7's accepted
  first-load figure — including the fact, if it remains true at measurement, that the figure was
  already exceeded before this story.
- Given the whole change, when the full verification cadence runs, then `declared` is still exactly
  `['IBM Plex Sans','IBM Plex Mono','IBM Plex Sans Thai']`, `canvas-font-stack.test.ts`'s
  engine/browser disjointness assertion is still true and still green, `shasum -a 256
  fixtures/*/expected.pdf` still returns the same 23 digests, and no test outside the two standing
  reds fails.
- Given Story 8.4b's later second registration, when both stories have landed, then chrome asks for
  real IBM Plex and the canvas asks for the engine's Noto face names, and the two vocabularies are
  separate by design rather than by accident.

## Spec Change Log

**2026-09-01 — Code Map re-anchored at the build dispatch's HEAD `a4bac02`.** The Code Map states
"every anchor re-measured at `e66e8b3`, this dispatch's HEAD", but the plan gate and the build are
**separate dispatches with seven commits between them** (`acfb68b`, `6f0c095`, `475cb50`, `2ded2e3`,
`90cdf8e`, `5ae02d7`, `a4bac02`). `git diff --stat e66e8b3..a4bac02` touches three of this story's
files. The intent contract is unchanged; these are corrections to rotted anchors only.

| Code Map anchor (as written, at `e66e8b3`) | Measured at `a4bac02` |
|---|---|
| `build-wasm.mjs` is **79 lines**; `:79` is the `runtime-fonts.css` template literal emitting **three** rules | **92 lines**. The literal is at **`:92`**, still one line, and emits **SIX** rules. `:79-91` is a new 13-line comment block. The `assets` object is still at `:64-71` and is unchanged. |
| `canvas-font-stack.test.ts` is **443 lines**; `declaredFamilies` `:43-45`, `requestedFamilies` `:48-54`, disjointness `:301-308`, `registersAFaceAtRuntime` `:338`, the `fontFamily:` census `:409` | **709 lines**. `declaredFamilies` `:78`, `requestedFamilies` `:105`, `registersAFaceAtRuntime` `:122`, the census `:675`. |
| `:304 engineFaces = ['Noto Sans','Noto Sans Thai','Noto Sans SC']` is **HARDCODED** — a stale-green risk flagged for 8.4b | **RESOLVED by 8.4b.** `:234` now reads `engineFaces = shippedFaceNames(...)` **dynamically from `folio-go/fonts/fonts.go`**, as D-8.4.14 required. The Design Note "A stale-green guard, noted for whoever owns 8.4b" is discharged. |
| `:301-308` the **engine/browser disjointness** assertion | **THIS ASSERTION NO LONGER EXISTS.** 8.4b replaced it with its logical opposite at `:540-574`: every engine face must be **declared** (`:554`) and **requested** (`:558`), in `fonts.Shipped()`'s own order (`:571-573`). The chrome half survives as an `arrayContaining` floor at `:548`. **Read AC5's "engine/browser disjointness assertion is still true and still green" as referring to the guard that replaced it** — the two vocabularies are now deliberately joined at the canvas, which is 8.4b's whole subject. Neither the swap nor any task in this story touches the assertions at `:540-574`. |
| `:179`/`:331` `declared.length >= 3` floors | `:246` is now `>= 6`; `:597` is still `>= 3`. Both tolerate this story's swap. |
| `font-binary-identity.test.ts` does not exist (Task 4 "create") | **EXISTS, 376 lines**, created by 8.4b. Task 4 is an EDIT — already corrected at the gate (D-8.4.22). |
| "**NOT ONE ASSERTION MOVES UNDER A FILE-SWAP** … the repository currently has no way to observe this story succeeding or failing" | **TRUE of `canvas-font-stack.test.ts`, and now FALSE of the repository.** `font-binary-identity.test.ts` opens files and compares sha256 digests (`:203`, `:370-374`). 8.4b built the beginning of the net; this story extends it rather than originating it. |
| Verification baseline "Designer … 37 files / 350 tests" | **38 files / 360 tests** at `a4bac02`. |
| Verification baseline "measured at `e66e8b3`" | **Re-measured at `a4bac02` by this dispatch** — see `## Auto Run Result`. The two standing reds are unchanged. |

**Upstream procurement re-verified before implementation:** `ibm-plex-mono.zip` fetched from the
pinned release URL has sha256 `6d23f01257663d8cc49a0d64c22ced630b79e0e2a0ac08a0da86e9a38bbc481c`,
**equal to the digest this spec recorded**. The pinned assets are reachable and unchanged.


## Review Triage Log

## Design Notes

**The four things a builder would otherwise get wrong, and what settles each.**

**1. The three Noto binaries STAY. This is not a preference; five independent things require it.**
(a) This story's own `Covers:` line is **NFR7**, whose text reads *"Latin + **CJK** + Thai coverage
requires a shipped font set embedded so the designer works fully offline"* — removing Noto Sans SC
removes CJK from the shipped set, so the story would violate the requirement it covers. (b) The
epic's own AC5 says that after 8.4b *"the canvas asks for the **engine's Noto face names**"*, which
requires those files to be in the designer bundle for 8.4b to register. (c) The dispatch settles it:
*"Story 8.4b (next) adds the Noto names to `declared`. This story does not."* (d)
`generate-offline-release.mjs:47-49` throws `production build has no /noto-sans-cjk. runtime asset`.
(e) `verify-offline-release.mjs:78-79` requires the `cjk-font` row to be the dominant font payload,
which is only true while a ~10.6 MB face is there. Five selectors pointing one way is a ruling, not a
judgement call.

**2. The intermediate state this creates, which D-8.4.17(c)'s own analysis did not name.** Between
this story and 8.4b, `noto-sans-cjk.ttf` **ships in the bundle with no `@font-face` family at all**.
Today CJK canvas text rasterizes correctly *by accident* — it falls through `App.css:118`'s stack to
`'IBM Plex Mono'`, which happens to be Noto Sans SC. After this story that slot is a real mono with
no CJK, so CJK canvas glyphs fall to the generic `sans-serif`. D-8.4.17(c) reasoned about the
intermediate state and named only one hazard — *"8.4b first would register the same Noto file under
two family names … reads as redundancy"* — and did **not** name this one.

This is reported, not halted on, for three measured reasons. It is **forced** by settled decisions
(`declared` stays three; declaring the CJK file under `Noto Sans SC` is 8.4b's ruled territory), so
there is no fork for a builder to pick. It is **one story long**, and 8.4b is the immediately next
story. And it is **unobserved by every executable gate**: the only designer file containing CJK
codepoints is `e2e/browser-native-roundtrip.spec.ts`, which is compile-only and never executed
(D-000.4), and jsdom rasterizes nothing. Task 5's disclosure assertion is what keeps the state
deliberate rather than accidental, and it pre-authorises its own deletion by 8.4b — the same shape
`canvas-font-stack.test.ts:149-151` used for 8.4a.

**3. `.ttf` from the GitHub release, never the npm package.** This is the single procurement decision
most likely to be got wrong, because `@ibm/plex-*` is the obvious route and it declares `OFL-1.1` in
its `package.json`. Measured: those tarballs contain **zero** `.ttf`/`.otf`/`.ttc`. `fontExtensions`
in `lint/internal/manifest/manifest.go:110` recognises only those three, so a `.woff2` face would
ship with **no LICENSE required, no NOTICE required, and no manifest row** — AD-26's asset half
silently not applying to the very binaries this story exists to bring under it. NFR7 independently
says take the glyf/TrueType static build. Two selectors, same answer.

**4. One static Regular per family — not a weight matrix.** `DESIGN.md` specifies `fontWeight` 400,
500 and 600 across its tokens, and the canvas sets `--text-font-weight: 700` and
`--text-font-style: italic`. None of that is satisfied by real files today either: the generator
emits one rule per family with **no `font-weight` or `font-style` descriptor**, so every weight and
italic is browser-synthesised from a single Regular. Shipping one Regular per family **preserves that
behaviour exactly** and changes only which face is synthesised from. Expanding to per-weight files is
work nothing asked for, would force `:79` to be restructured into a loop — which silently empties
`declaredFamilies`' regex and takes the guard file with it — and matches neither the AC's singular
*"an actual IBM Plex OFL binary"* nor the Noto precedent of static Regular-only instances.

**Why the identity guard is a new designer test and not an extension of `lint`.** The alternative was
`lint/internal/rules`, which already owns *"a font binary is what it claims to be"* (`looksLikeSfnt`)
and already reaches designer paths elsewhere (`ScanPDFJSNotice`). It was rejected on a measured
ground: **`lint/go.mod` has no font parser** — it requires only `golang.org/x/tools` — so the check
would need either a hand-rolled `name`-table parser in Go or a new module dependency, and `lint` is
the module whose own dependency graph `ScanLicenceGraph` audits. The designer side needs no
dependency at all (a `name`-table read is a short `DataView` walk), sits beside the file that already
owns the generator's declaration set, and reads the same tracked sources `canvas-font-stack.test.ts`
deliberately reads instead of the gitignored output. Recorded because it is a real fork with no
correctness stake either way; if the orchestrator prefers `lint`, the tasks move without changing an AC.

**Why AC3 needs no document edit, which will look like an omission.** `ARCHITECTURE-SPINE.md:512-518`
(AD-26's Rule), its Stack table at `:553`, and `epics.md:206` **already assert** that the shipped IBM
Plex faces are OFL 1.1 travelling with their licence text and copyright lines. Those statements are
false today and become true the moment the binaries land. AC3 asks for the statement to *be true*, not
for it to be written — so the correct diff touches no planning artifact, and a builder who "fixes the
docs" is amending architecture nobody authorised.

**What this story's gates can and cannot prove.** Vitest under jsdom **applies no stylesheet and loads
no font**. So the identity guard proves the *binding* — which family name resolves to which bytes, and
what those bytes declare themselves to be. **Nothing in this repository can execute a real font load
or rasterize a glyph**: `npm run test:e2e` (Playwright) appears in no workflow, and
`test:e2e:compile` is `tsc --noEmit`. Any claim that the chrome *visibly* renders in IBM Plex is
unverifiable by these gates. Say so at close rather than letting a compile pass read as a run.

**A stale-green guard, noted for whoever owns 8.4b.** `canvas-font-stack.test.ts:301-308` compares a
**dynamically read** `declared` against a **hardcoded** `engineFaces = ['Noto Sans','Noto Sans
Thai','Noto Sans SC']`. It is true and green throughout this story. But if a later story changes what
`folio-go` ships, the engine half does not follow, and the disjointness assertion would keep passing
while having become false. Out of scope here; 8.4b is the story that touches it, and D-8.4.14 already
requires it be **replaced, not weakened**.

## Verification

**Baseline, MEASURED at `e66e8b3` (this dispatch's HEAD) — not assumed.** Re-measure before the first
edit; the numbers below are what a clean tree produced today.

- `cd folio-go && go test -count=1 ./...` → **1811 pass / 2 fail / 5 skip**. The two failures are the
  one mandated red and its parent: `TestCorpusMeetsP6ExerciseFloors` and
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` in `internal/text`
  (`P6g floor not met: got 7, need >=20`). Never "fix" it.
- `cd folio-go && go test -count=1 -tags=matrix ./...` → **1822 pass / 3 fail / 5 skip**. The third is
  `TestShippedFacesReproduceFromUpstream` (`folio-go`), and it is a **could-not-execute, not a byte
  divergence**: `fontgen: fontTools is not importable by this interpreter`
  (`/opt/homebrew/opt/python@3.12/bin/python3.12`). Report it in those words (DW-86).
- **Exactly TWO standing reds. Any third is a real failure.**
- `cd lint && go test -count=1 ./...` → four `ok` lines, **no FAIL**. Use `-count=1` always: this
  module's rules walk directories and Go's test cache does not track `ReadDir`, so a new file in the
  scanned tree never invalidates a cached `ok`. **This story adds files to a scanned tree, so a cached
  `ok` here is no measurement at all.**
- `cd folio-go && go vet -tags=matrix ./...` → no output. `gofmt -l folio-go` → no paths.
- `shasum -a 256 fixtures/*/expected.pdf` → **23 lines**. `md5 README.md` →
  `078d7d80d518d54af2fc04fb270d46b8`.
- Designer: `npm run typecheck` clean; `npm run lint` → **exactly 4 warnings, 0 errors**, all
  `react(only-export-components)` at `src/preview/pdf-viewer.tsx:16,17` and `src/App.tsx:1323,1330`;
  `npm test` → **37 files / 350 tests, all passing**; `npm run test:e2e:compile` clean.
- AD-21 legs: env var `FOLIO_MATRIX_TARGET`, targets `darwin/arm64`, `linux/amd64`, `linux/arm64`,
  `js/wasm` (`matrix_test.go:69`, `wantMatrixLegs = 4`). All four asserted (16 documents hashed each,
  identical across targets). The **unset control passes while asserting nothing** and says so at
  `matrix_test.go:2199` — count it as a control, never as a fifth leg.
- `TestCrossTargetByteIdentity` → PASS.

**Commands — heavy cadence, EVERY story, `-count=1` on every Go gate:**

- `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` **before the first edit** --
  expected: 23 lines, matching the baseline.
- `cd folio-go && go test -count=1 ./...` -- expected: the two standing reds and nothing else.
- `cd folio-go && go test -count=1 -tags=matrix ./...` -- expected: the same two plus
  `TestShippedFacesReproduceFromUpstream`'s could-not-execute; **no fourth failure**.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: no output.
- `gofmt -l folio-go` -- expected: no paths.
- `cd lint && go test -count=1 ./...` -- expected: four `ok`, no FAIL. **This story's licence check
  lives here; it is not a formality for this one.** Run `-run TestManifestUpToDate -v` and
  `-run TestResolveAssets -v` explicitly and quote what they print.
- `cd lint && go run ./cmd/genmanifest && git diff --stat -- lint/MANIFEST.md` -- expected: after the
  regeneration is committed, a second run leaves no diff.
- For each of `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`:
  `cd folio-go && FOLIO_MATRIX_TARGET=<T> go test -count=1 -tags=matrix -run TestTargetRenderHash -v .`
  -- expected: PASS with 16 documents hashed. Plus the unset control, reported as a control.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` -- expected: PASS.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile`
  -- expected: typecheck clean; **exactly 4** oxlint warnings, the same four; Vitest ≥ 350 passing
  with the new identity tests added and **zero failures**; e2e compile clean.
- `cd folio-designer && npm run build` (which runs `build:wasm`, `tsc -b`, `vite build`,
  `build:offline`, `verify:offline`) -- expected: completes; **the bundle gains binaries, so this is
  not optional for this story**. Requires node `v24.16.0` exactly (`RELEASE_RUNTIME` in
  `offline-release-contract.mjs:3`) and a Go toolchain.
- `cd folio-designer && npm run verify:offline:red && npm run verify:offline:wasm` -- expected: both
  pass. In particular `S1 CJK row is not the dominant font payload` must **not** fire.
- `shasum -a 256 fixtures/*/expected.pdf | diff <scratch>/digests.before -` -- expected: **empty
  diff**. Any moved digest is a defect, not a re-record.

**Manual checks:**

- `md5 README.md` -- must still be `078d7d80d518d54af2fc04fb270d46b8`; `git status --porcelain` must
  never show `README.md` at any point.
- `git diff --stat` on each commit -- must touch **no** path under `folio-go/`, no `.folio` file, no
  `fixtures/**`, and no file under `_bmad-output/planning-artifacts/`.
- `git log -1 --stat` after each commit -- confirm the file list is exactly the intended one; this
  story's commits are file-additions and a two-line generator edit, so an unexpected path is a breach.
- Read `lint/MANIFEST.md`'s `## Redistributed non-code assets` table and confirm each IBM Plex row's
  Licence column reads **`OFL-1.1`** and its Copyright column carries the IBM line. A `SEE NOTICE`
  label is a Block If.
- Read the generated `folio-designer/src/generated/runtime-fonts.css` after `npm run build:wasm` and
  confirm it still has **exactly three** `@font-face` rules, with the same three family names, each
  `src` now naming an `ibm-plex-*` file.
- `shasum -a 256` each committed IBM Plex binary and confirm it equals the digest recorded in its
  NOTICE.md and the digest of the file extracted from the pinned release asset.
- Record the measured weight numbers (added total, retained total, new total, and the release's
  `s1VisibleBytes` before and after) in the final commit message and in `## Auto Run Result`.

## Auto Run Result

Status: done
Blocking condition: none

### Build dispatch — 2026-09-01, baseline/HEAD `a4bac02`, branch `main`, two commits, no push, no branch

**COMMIT 1 (AC1) `3d0eba1` — "Draw the chrome's mono type in a face that is actually monospaced".**
`folio-designer/public/fonts/ibmplexmono/` (`.ttf` + `LICENSE-OFL.txt` + `NOTICE.md`), one `assets`
slot `mono`, the `IBM Plex Mono` rule repointed off `sansCjk`, the identity guard added to
`src/font-binary-identity.test.ts`, `lint/MANIFEST.md` regenerated.

**COMMIT 2 (AC2–AC5) — `ibmplexsans/` and `ibmplexsansthai/` with the same three-file shape, slots
`plexSans` and `plexSansThai`, both remaining rules repointed, the identity guard generalised to all
three chrome families with the Thai coverage assertion, `lint/MANIFEST.md` regenerated.**

**Upstream procurement — RE-FETCHED AND RE-VERIFIED, not recalled.** All three release archives were
downloaded from the pinned URLs and all six digests in the Code Map matched exactly: the zips
(`6d23f012…`, `fb365d91…`, `d7203f43…`) and the three extracted faces (`7c6fbddc…`, `975dcda3…`,
`83e1db8e…`). Each committed file is byte-identical to the file extracted from the archive, so each
NOTICE records the same value for the upstream and the committed digest and says explicitly that no
derivation applies. `LICENSE-OFL.txt` in each directory is the archive's own top-level `LICENSE.txt`,
unmodified and identical across the three (sha256 `7e6b2818…`). No npm package was used.

**Every claim the spec carried was re-measured against the committed bytes, and every one held.**
IBM Plex Mono: `post.isFixedPitch = 1`, and all 1,149 non-zero advance widths are 600/1000 — one
distinct width font-wide. IBM Plex Sans: 95/95 ASCII, 96/96 Latin-1 Supplement, 128/128 Latin
Extended-A, zero gaps across U+0020–U+017F. IBM Plex Sans Thai: exact cmap parity with the shipped
Noto Sans Thai (87 of 128 in U+0E00–U+0E7F, empty set difference in both directions), both defect
strings `พระราชบัญญัติ` and `การทวงถามหนี้` fully covered, and under script `thai` langsys `dflt`
MarkBasePos ×3 / MarkMarkPos ×2 / PairPos ×3 / ChainContextPos ×2 against Noto's ×1 / ×2 / ×1 / ×2,
plus `calt` which Noto lacks. The one recorded difference stands: Noto declares a `thai` `dist`
feature that Plex does not. All three: `fvar` absent, `glyf` outlines, no `CFF`, nameID 13 the OFL
1.1 string, nameID 14 `http://scripts.sil.org/OFL`.

**Block Ifs, all cleared.** Digests matched; every `name` table declares the family the spec said;
each `lint/MANIFEST.md` row reads **`OFL-1.1`** (not `SEE NOTICE`) with
`Copyright © 2017 IBM Corp. with Reserved Font Name "Plex"`; `shasum -a 256 fixtures/*/expected.pdf`
is byte-identical to the pre-edit capture, 23 lines; no third distinct test failure appeared; and the
verifier's `S1 CJK row is not the dominant font payload` did not fire — the `cjk-font` row is
4,948,312 Brotli bytes against 226,026 for the next largest font row.

### The measured weight, and NFR7

| figure | before (`a4bac02`) | after | delta |
|---|---|---|---|
| Added IBM Plex raw bytes | — | 200,500 + 173,052 + 116,728 = **490,280** | +490,280 |
| Retained Noto raw bytes | 11,289,880 | 11,289,880 | 0 |
| Total committed designer font weight | 11,289,880 | **11,780,160** | **+4.34%** |
| Release `s1.cachedBytes` (raw precache total) | 37,967,930 | **38,458,923** | +490,993 |
| Release `s1.assetCount` | 20 | **23** | +3 |
| Release `s1VisibleBytes` (Brotli) | 12,424,172 | **12,424,282** | +110 |

**`s1VisibleBytes` does not measure this change, and saying otherwise would be false.** The figure
sums only four rows — the engine wasm and the three Noto faces (`generate-offline-release.mjs:50`).
The IBM Plex faces are cached assets but are not S1 rows, so **none** of the +110 is IBM Plex; that
delta is the engine wasm's own Brotli size moving between two builds of identical Go source
(7,224,962 → 7,225,072), which is exactly the non-reproducibility **DW-100** already records. The
figure that actually moves is `s1.cachedBytes`, +490,993 (the 490,280 font bytes plus 713 of
`index.html` bootstrap growth for three more cache entries).

**Against NFR7's accepted `~9 MB first load` (`epics.md:150`): the budget was already exceeded before
this story and remains so.** 12,424,172 Brotli bytes at baseline, 12,424,282 after. Nothing enforces
the figure — it is prose — and this spec forbids editing the epic, so it is reported, not acted on.

### Verification — full heavy cadence, run at the finished tree

- `cd folio-go && go test -count=1 ./...` → 13 `ok`, **exactly the two standing reds**:
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest (`got 7, need >=20`).
- `cd folio-go && go test -count=1 -tags=matrix ./...` → the same two plus
  `TestShippedFacesReproduceFromUpstream`, which is a **could-not-execute, not a byte divergence**:
  `fontgen: fontTools is not importable by this interpreter`
  (`/opt/homebrew/opt/python@3.12/bin/python3.12`) — DW-86, unchanged. **No fourth failure.**
- `cd folio-go && go vet -tags=matrix ./...` → no output. `gofmt -l folio-go` → no paths.
- `cd lint && go test -count=1 ./...` → four `ok`, no FAIL, run after each binary addition.
  `-run TestManifestUpToDate -v` → `--- PASS: TestManifestUpToDate (0.25s)`. `-run TestResolveAssets -v`
  → five PASS (`…IncludesWordlist`, `…ExcludesUntrackedDirectoryWithoutError`,
  `…StillReportsATrackedViolation`, `…AllDirectoriesUntrackedIsAScanError`,
  `…NoCandidateDirectoriesIsNotAScanError`).
- Designer: `npm run typecheck` clean; `npm run lint` → **exactly 4 warnings, 0 errors**, the same four
  `react(only-export-components)` at `pdf-viewer.tsx:16,17` and `App.tsx:1323,1330`; `npm test` →
  **38 files / 365 tests, all passing** (350 → 360 at 8.4b → 365 here, +5 from this story);
  `npm run test:e2e:compile` clean.
- `npm run build` completes on node `v24.16.0`; `npm run verify:offline`, `verify:offline:red` and
  `verify:offline:wasm` all pass.
- All four AD-21 legs PASS (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`); the unset control
  passes while asserting nothing and is counted as a control, not a fifth leg.
  `TestCrossTargetByteIdentity` PASS.
- `shasum -a 256 fixtures/*/expected.pdf | diff <before> -` → **empty**. `md5 README.md` →
  `078d7d80d518d54af2fc04fb270d46b8`, and `README.md` never appeared in `git status`.
- Generated `runtime-fonts.css` read by hand after `build:wasm`: still **six** rules (three chrome,
  three engine, per Story 8.4b), the three IBM Plex families now naming `ibm-plex-sans.*`,
  `ibm-plex-mono.*` and `ibm-plex-sans-thai.*`, the three engine names still naming the Noto files,
  every `src` still `url('./runtime/…') format('truetype')`.
- No commit touches `folio-go/**`, any `.folio` file, `fixtures/**`, or any planning artifact.

### The guard was shown to discriminate, twice, through the code the real assertion runs

Beyond the in-file red-proof fixtures, the **real generator was mutated** and the real assertions
reddened: binding `IBM Plex Mono` back at `assets.sansCjk` failed four assertions, naming
`"IBM Plex Mono: declared from public/fonts/notosanssc/NotoSansSC-Regular.ttf, whose name table says
'Noto Sans SC'"`; binding `IBM Plex Sans` and `IBM Plex Sans Thai` back at the Noto slots failed
three, naming both. Both mutations were reverted and the suite re-run green.

### Reported at close — three things, none of them halting

1. **What these gates cannot prove.** Vitest under jsdom applies no stylesheet and loads no font, and
   this repository has **no gate that executes a real font load or rasterizes a glyph** —
   `npm run test:e2e` appears in no workflow and `test:e2e:compile` is `tsc --noEmit`. The **binding**
   is proved: which family name resolves to which bytes, and what those bytes declare themselves to
   be. That a browser **visibly** draws the chrome in IBM Plex is not proved by anything here.
2. **The CJK rasterization gap the spec's Design Note 2 predicted did not open, because the story
   order was reversed.** Under the planned order, `IBM Plex Mono` ceasing to be Noto Sans SC would
   have left canvas CJK text falling through to the generic `sans-serif` until 8.4b landed. 8.4b
   landed **first** and registered `Noto Sans SC` under its own name, so the CJK file keeps a
   `@font-face` family throughout and the interval never existed. Task 5's disclosure-of-absence
   assertion was correctly struck at the gate (D-8.4.22) rather than written.
3. **The IBM Plex faces are narrower than the Noto faces outside the chrome's range** — 895 mapped
   code points for Plex Sans against Noto Sans's 3,094 (Latin Extended-B 33 vs 208, Greek 73 vs 121,
   Cyrillic 192 vs 256). The chrome draws none of those, and canvas text asks for the engine's Noto
   names, so nothing regresses; recorded because a reader will wonder.

### Plan dispatch (halt after planning), retained for the record


**Dispatch:** classic intent, plan-only (`Halt after planning.`). Baseline/HEAD `e66e8b3`, branch
`main`, working tree clean at entry. No code written, no commit created, no branch, no push.

**Premises checked against the tree, with results.**

- *"Under this option the chrome family NAMES do not change at all"* — **CONFIRMED.** `tokens.css:11`
  and `App.css:118` need no edit; `declared` is parsed from the generator's family names only, and the
  `src:` is never read by any assertion.
- *"`canvas-font-stack.test.ts`'s engine/browser disjointness assertion stays true throughout"* —
  **CONFIRMED**, at `:301-308`. Reported with a caveat: its engine half (`:304 engineFaces`) is
  **hardcoded**, so it is a stale-green risk for whoever changes what `folio-go` ships. 8.4b's
  territory; D-8.4.14 already requires it be replaced rather than weakened.
- *"Someone must confirm IBM Plex has the Thai coverage `--font-page` asks for"* — **DISCHARGED BY
  MEASUREMENT, not deferred.** IBM Plex Sans Thai 1.1.0 has exact cmap parity with the shipped Noto
  Sans Thai (87/128 in U+0E00–U+0E7F, empty set difference both directions), covers both strings from
  the recorded rendering defect, and reaches strictly more Thai-script GPOS mark positioning
  (MarkBasePos ×3 / MarkMarkPos ×2 vs ×1 / ×2). **No halt.** One measured difference recorded: Noto
  declares a `thai` `dist` feature that Plex does not.
- *"AD-26 — verify the licence check actually reaches the new binaries"* — **VERIFIED, and it does.**
  `manifest.ResolveAssets` walks the whole repository and requires a same-directory `LICENSE*` +
  `NOTICE*` with a `Copyright` line, failing the build without them. **But it recognises only
  `.ttf`/`.otf`/`.ttc`** — a `.woff2` face would be invisible to it, and the npm `@ibm/plex-*`
  packages ship **no** `.ttf` at all. That trap is now an Always constraint and a Design Note.
- *"The added bundle weight must be measured and recorded"* — measured: **+490,280 bytes** (200,500 +
  173,052 + 116,728), against 11,289,880 bytes of retained Noto faces, for a new total of 11,780,160
  (+4.34%).

**Two findings for the orchestrator that the spec cannot resolve on its own.**

1. **D-8.4.17(c)'s intermediate-state analysis did not name the CJK rasterization gap.** It reasoned
   about the 8.4c→8.4b ordering and named one hazard (*"the same Noto file under two family names …
   reads as redundancy"*). Measured: CJK canvas text today rasterizes *by accident*, falling through
   `App.css:118`'s stack to `'IBM Plex Mono'`, which is Noto Sans SC. After this story that slot is a
   real mono with no CJK, so CJK glyphs fall to generic `sans-serif` **until 8.4b lands**. Not halted
   on: the outcome is forced by settled decisions (so there is no fork to pick), it is one story long,
   and no executable gate observes it — the only designer file with CJK codepoints is the compile-only
   `e2e/browser-native-roundtrip.spec.ts`. Task 5 makes the state deliberate and pre-authorises 8.4b to
   delete the disclosure. **If the ordering ruling would change knowing this, that is the lead's call.**
2. **NFR7's bundle budget was already exceeded before this story.** `epics.md:150` accepts *"~9 MB
   first load"*; the committed release manifest measures `s1VisibleBytes = 12,372,693` Brotli bytes.
   Nothing enforces the figure — it is prose. The spec forbids editing the epic, so this is reported
   rather than acted on.

**Also reported:** no test in the repository can currently observe a font file swap — zero assertions
read any `@font-face` `src`, filename, path, or byte. The story therefore has to build its own
regression net (Task 4/10), which is why AC1 is phrased as an observation rather than an edit.

**Baseline measured at `e66e8b3`, not assumed.** `go test -count=1 ./...` → 1811 pass / 2 fail / 5
skip. `go test -count=1 -tags=matrix ./...` → 1822 pass / 3 fail / 5 skip. **Exactly two standing
reds:** `TestCorpusMeetsP6ExerciseFloors` with its `P6g_(opaque_names)` subtest (`got 7, need >=20`),
and — matrix only — `TestShippedFacesReproduceFromUpstream`, which is a **could-not-execute**
(`fontgen: fontTools is not importable by this interpreter`), not a byte divergence. No third red.
`go vet -tags=matrix ./...` and `gofmt -l folio-go` clean. `cd lint && go test -count=1 ./...` → four
`ok`, no FAIL. `shasum -a 256 fixtures/*/expected.pdf` → 23 lines. `md5 README.md` →
`078d7d80d518d54af2fc04fb270d46b8`. Designer: typecheck clean; oxlint exactly 4
`only-export-components` warnings; Vitest 37 files / 350 tests all passing; `test:e2e:compile` clean.
All four AD-21 legs asserted (16 documents each, identical hashes); the unset control passes while
asserting nothing and says so. `TestCrossTargetByteIdentity` PASS.

**Tree at halt:** ` M _bmad-output/implementation-artifacts/epic-8-context.md` (this workflow's own
step-1 recompile, the epics file having been touched after the cached context was written) and the
untracked spec. Nothing else. HEAD unmoved at `e66e8b3`.
