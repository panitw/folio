---
title: 'The shipped families gain a weighted and a sloped face'
type: 'feature'
created: '2026-09-06'
status: 'done'
baseline_commit: 'ffec48cdfd4ed2d61badfb1dbc595af556f4b5ed'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-11-context.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-11-14-decision-log.md'
---

## In plain terms (read this first if you just want the gist)

*Not normative, and rewritten after the fact: the frozen Intent below governs implementation, while
this section says what actually shipped.*

The product now ships seven new typeface cuts, so that asking for bold or italic finally has something
real to resolve to. Noto Sans and Roboto each gained a bold, an italic and a bold italic. Thai gained
bold only — the people who publish that typeface make no Thai italic at all, so the nine faces the epic
assumed were never possible.

Nothing here reads a weight or paints one. A document asking for bold still prints in book weight, and
that is deliberate: the following two stories do the resolving and the painting. What shipped is the raw
material and its paperwork — every cut carries its own licence and its own record of where its bytes came
from, and the build re-derives four of them from their sources on demand to prove none was hand-edited.

The cost is about three megabytes of new font data, close to what was predicted before anything was
fetched. The offline download now holds sixty-one cached pieces against a ceiling of sixty-four, so the
build has gained a warning that the margin has narrowed to three. That warning fires today by design; it is
not a failure.

Two Thai provenance records are still incomplete, though the missing value is now on hand. One
review finding was a genuine defect rather than a gap: a fallback rule could have made an entire
writing system print bold in every author's document. It was closed before anyone saw it.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** `Style.Bold` and `Style.Italic` are stored, projected and read by no producer on the render
path, because `folio-go/fonts/` embeds four faces and every one of them is Regular. A bold heading prints
in book weight and two documents differing only by `"bold": true` render to the same SHA-256. There is
nothing for a weight to resolve to.

**Approach:** Ship **seven new faces** so a later story has something to resolve to. Noto Sans and Roboto
each gain Bold, Italic and Bold Italic; Noto Sans Thai gains Bold only. Each cut is a **face of its own
under its own family name**, on both the Go and the browser side, carrying its own provenance record and
its own licence text. Nothing in this story reads a weight or paints one — 11.2 resolves, 11.3 paints.

## Boundaries & Constraints

**Always:**
- **Seven faces, and the count is measured, not assumed.** Noto Sans and Roboto get Bold / Italic /
  Bold Italic; Noto Sans Thai gets **Bold only**. D-A said "three families, nine new faces"; that half of
  D-A rests on a false premise. Measured at this plan gate: `folio-designer/font-index.json` records
  Noto Sans Thai's styles as `100`–`900` with **no italic variants at all**, and upstream publishes none.
  Seven is the whole realizable set.
- **Noto Sans SC stays Regular, deliberately (D-A).** State it in the NOTICEs and the Delivery Log as a
  ruling with its figure attached, never as a silent absence: its Regular alone is **10,595,932 bytes**,
  and three instances of it would take the offline payload from ~11 MB to ~45 MB. A reader who counts
  families must find the reason without asking.
- **A shipped face is committed output of a replayable derivation.** Derived cuts come from
  `tools/fontgen/instance_faces.py`, committed as output; `make fonts-verify` must reproduce every
  committed face byte-for-byte.
- **Each cut is its own family name, and that one string appears on three surfaces** (D-B, lead): the
  `fonts.Shipped()` key, the `@font-face` family the browser declares, and the family the canvas paints
  with. Confirmed key spelling: `Noto Sans Bold`, `Noto Sans Italic`, `Noto Sans Bold Italic`,
  `Noto Sans Thai Bold`, `Roboto Bold`, `Roboto Italic`, `Roboto Bold Italic`.
- **AD-26: every cut carries its own `NOTICE.md` and its own `LICENSE-OFL.txt`, in its own directory**,
  on both sides. Not negotiable by layout.
- **Byte identity (AD-21).** No corpus document declares bold or italic — verified CLOSED over 34 input
  documents under `fixtures/`, zero matches, positive control `fontSize` 57 lines. The corpus must hash
  identically on all four targets after this story.

**Ask First:**
- Any change to `folio-go/fontset.go`'s `FontSet` shape, or to `FontChainEntry`. **D-B forecloses both**;
  weight and slope reach a face through the chain entry in 11.2, and the public API does not change.
- Any post-processing of a derived face's name table, or any hand-edit of a font binary.
- Adding an eighth face, or giving Noto Sans SC a cut.
- Moving `maximumCacheAssets` (64) rather than adding the approach-warning below it.

**Never:**
- **Never parse the `Shipped()` key.** The key is chosen to be READABLE and nothing may derive a family
  from it. One careless `strings.TrimSuffix(key, " Bold")` reinstates the naming-convention weight carrier
  D-B foreclosed. The machine-readable family is `shippedFaceSpecs.Family`, which is sfnt name ID 1 and is
  asserted against the binary.
- Never synthesize bold or oblique, at emit time or on the canvas (I-2). `SetSyntheticBold` /
  `SetSyntheticSlant` are exposed by the vendor and called nowhere; keep it that way.
- Never resolve a weight, paint one, or touch `render.go`, `table_render.go` or the canvas. That is 11.2
  and 11.3. This story adds faces and their accounting, nothing else.
- Never give a new cut a **catalogue** entry. `font-catalogue.test.ts` asserts every catalogue face is
  upright Regular 400, so a bold cut cannot be one; the catalogue legitimately stays Regular-only.
- Never add a fetch step to the build. Sources are procured by a human, out of band, as they always have
  been.
- **Never widen the canvas fragment fallback stack.** The counts in `canvas-font-stack.test.ts` move;
  **the stack does not.** Widening it changes what every unattributed fragment in every existing document
  falls back to — a silent rendering change to documents nobody edited, under a byte-determinism regime
  whose whole premise is that output moves only when input does. It would arrive disguised as a tidy-up.
- Never filter the matrix suite with `-run` to check a gate.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Derived cut, replayed | `make fonts-verify` with the pinned interpreter and `.font-sources/` populated | Every one of the four derived cuts reproduces to its recorded `out_sha256`; witness reads `derived and compared 7 of 7 faces` | Non-zero exit naming the diverging face |
| Static cut, accounted | Roboto Bold / Italic / Bold Italic | `static-upstream` route: recorded SOURCE digest == recorded SHIPPED digest == embedded digest | `t.Fatalf` naming the face and the disagreeing pair |
| Face with no NOTICE | a binary committed under `folio-go/fonts/<dir>/` its NOTICE does not name | Build fails: one binary per directory, one NOTICE per binary | `t.Fatalf` from `readNoticeRecords` |
| Engine face with no browser rule | a `Shipped()` key with no `@font-face` rule and no catalogue entry | Build fails — the mirror is half-populated | `familiesWithNoRule` returns it; `.toEqual([])` fails |
| Corpus render, all targets | any fixture, none declaring bold or italic | Bytes identical to before this story on all four targets | `TestCrossTargetByteIdentity` / golden digests fail |
| Cache-asset margin | release generated with the seven new assets | `s1.assetCount` 61 of `maximumCacheAssets` 64, and the build WARNS that the margin is 3 | Warning names the margin; hard failure only above 64 |
| Load screen total | a payload whose rows include an `embedded-in-engine` row | Displayed total equals Σ `cacheAssets`, never Σ `rows` | Test fails if the total is computed from rows |

</frozen-after-approval>

## Code Map

**Measured at this plan gate, at HEAD `1c0260d`. Cite by symbol; line numbers are not addresses here (D-000.4).**

**HEAD is now `ffec48c` (baseline for this build). Every measurement below still stands** — verified:
both commits since `1c0260d` (`46ce9e5`, `ffec48c`) touch `_bmad-output/` **only — decision log,
deferred-work, guard census; no code, no fixtures, no fonts**, so nothing in this map needs re-measuring.

### The seven faces, their routes, and their measured bytes

All seven derive/extract deterministically; the four derived cuts were produced twice at this gate and
hashed identically both times.

| face / `Shipped()` key | route | source | bytes | sha256 |
|---|---|---|---|---|
| `Noto Sans Bold` | derived | `NotoSans-VF.ttf` `wght=700 wdth=100` | 648,284 | `652b4b154d1c41f01de4c69b6d37d6a73a1c942e43bfcf4f95d4490b2fca6787` |
| `Noto Sans Italic` | derived | `NotoSans-Italic-VF.ttf` `wght=400 wdth=100` | 663,520 | `b3f9077e2c43979d1509bead91c46068742f16f3dd131b8005149451aaf40338` |
| `Noto Sans Bold Italic` | derived | `NotoSans-Italic-VF.ttf` `wght=700 wdth=100` | 665,508 | `1a1882aa2efca4388498a9db2ced6eca5dd6141d2cd989248e2bf1d189718df1` |
| `Noto Sans Thai Bold` | derived | `NotoSansThai-VF.ttf` `wght=700 wdth=100` | 47,800 | `fe60f91611714dc6a57d5facb1818292b08c22cf88a55c60084173ea92e2ddbd` |
| `Roboto Bold` | static-upstream | `Roboto_v3.016.zip` -> `android/static/Roboto-Bold.ttf` | 358,188 | `9aa793ad5e12c3b2486464ec6f047f16ac6d30bb06c8b4d162ba313459794968` |
| `Roboto Italic` | static-upstream | same archive -> `android/static/Roboto-Italic.ttf` | 375,320 | `1f4b29f2e9c648707620a0768f4030df7abb15c25383fb73d088fe2624724f46` |
| `Roboto Bold Italic` | static-upstream | same archive -> `android/static/Roboto-BoldItalic.ttf` | 378,148 | `8942bdf86849104971ccc41f038a91cc7e84a6b04a0df718fca6cac0a8c4e44a` |

**Total 3,136,768 bytes (2.99 MiB)** — shipped set 11,645,836 -> 14,782,604 raw, **+26.9%**. This lands on
D-A's "~3 MB" estimate; the owner's arithmetic held.

**Face metadata, measured — these are `shippedFaceSpecs` row values, not predictions:**

| face | name[1] | name[2] | name[6] | usWeightClass | fsSelection | macStyle | italicAngle |
|---|---|---|---|---|---|---|---|
| Noto Sans Bold | `Noto Sans` | `Bold` | `NotoSans-Bold` | 700 | `0x00a0` | `0x0001` | 0.0 |
| Noto Sans Italic | `Noto Sans` | `Italic` | **`NotoSansItalic-Italic`** | 400 | `0x0081` | `0x0002` | -12.02 |
| Noto Sans Bold Italic | `Noto Sans` | `Bold Italic` | **`NotoSansItalic-BoldItalic`** | 700 | `0x00a1` | `0x0003` | -12.02 |
| Noto Sans Thai Bold | `Noto Sans Thai` | `Bold` | `NotoSansThai-Bold` | 700 | `0x00a0` | `0x0001` | 0.0 |
| Roboto Bold | `Roboto` | `Bold` | `Roboto-Bold` | 700 | `0x0020` | `0x0001` | 0.0 |
| Roboto Italic | `Roboto` | `Italic` | `Roboto-Italic` | 400 | `0x0201` | `0x0002` | -12.0 |
| Roboto Bold Italic | `Roboto` | `Bold Italic` | `Roboto-BoldItalic` | 700 | `0x0221` | `0x0003` | -12.0 |

⚠ **The two bolded PostScript names are correct output and look like typos.** fontTools composes name[6]
from the *source* VF's variations-PostScript prefix, which for the italic VF is `NotoSansItalic`. The bold
cut escapes it because the roman VF's prefix is `NotoSans`. Hand-correcting a name table would put a
manual edit inside the byte-identity regime, which the replayable-script discipline exists to forbid.
**A comment recording this must sit at `shippedFaceSpecs` itself** — whoever is about to "fix" these two
rows is the person who needs the warning, and the spec is read once.

### Provenance inputs

- `.font-sources/` (gitignored, `.gitignore:123`) now holds four VFs; `NotoSans-Italic-VF.ttf` was added
  out of band for this story.
- `Roboto_v3.016.zip` — archive sha256 `1653dbe12f248da8fb0b9920db7b9496cd677ed3981154f6f15285c8bd4e334f`,
  29,162,959 bytes. **Already recorded** at `folio-designer/public/fonts/roboto/NOTICE.md`, and verified.
- **The three Roboto statics are already extracted and verified on disk** at
  `/private/tmp/claude-501/-Users-panitw-Projects-folio/c910e6ef-835c-4b93-9710-f11720b44f3a/scratchpad/upstream/x/`
  (`Roboto-Bold.ttf`, `Roboto-Italic.ttf`, `Roboto-BoldItalic.ttf`). Copy them from there. **Do not
  download anything** — procurement is done, and no build step may gain a network call.
- `NotoSans-v2.015.zip` — archive sha256
  `0c34df072a3fa7efbb7cbf34950e1f971a4447cffe365d3a359e2d4089b958f5`, 117,491,253 bytes. **Not previously
  recorded anywhere in the repo; record it in the new NOTICEs.** It was verified only **transitively** —
  the extracted roman VF hashes to the `src_sha256` already pinned in `UPSTREAM`. Say so in those terms;
  it is a weaker guarantee than Roboto's and must not be written as parity.

### Guards that read a file this story touches — as SOURCE TEXT or as a DIRECTORY LISTING

The existing census (`source-text-guard-inventory.md`) covers text-matchers only and **structurally cannot
see directory-listing guards**; three of this story's four hardest blockers are that second kind. It also
never names `accounting_test.go`, `folio-go/fonts/fonts.go`, or `NOTICE.md`. Treat this table as the
correction.

| file:symbol | reads | strips comments | fragility | what this story owes it |
|---|---|---|---|---|
| `folio-go/fonts/accounting_test.go:readNoticeRecords` | **dir listing** of `folio-go/fonts/` + each `NOTICE.md` as text | **no** | HIGH — four `(?m)^\|…\|$` whole-line regexes; `exactlyOneRow` fatals on 0 **and** 2+ | **Hardest blocker: one binary per directory.** Seven new sibling dirs, each with one `.ttf`, one `LICENSE-OFL.txt`, one `NOTICE.md` carrying all four rows and exactly one derivation statement. Untagged. |
| `folio-go/fonts/accounting_test.go:readFontgenManifestKeys` | `instance_faces.py` as text | **no** — Python `#` counts as live | HIGH — `(?s)\nUPSTREAM = \[\n(.*?)\n\]\n`, both brackets alone at column 0, non-greedy | New UPSTREAM entries must keep the block shape; keys byte-identical to the new `Shipped()` keys |
| `…:TestEveryShippedFaceIsAccountedForByExactlyOneRoute` | join of the above with `Shipped()` | n/a | no default arm | Each cut in **exactly one** route. Derived ⇒ UPSTREAM entry + DERIVATIVE blockquote. Static ⇒ `copied unmodified, no derivation` + SOURCE digest == SHIPPED digest == embedded digest |
| `…:TestEveryShippedFaceHasARecordedProvenance` | NOTICE rows vs disk vs embedded bytes | no | inherits row anchors | Every recorded sha256 and `Size` must match the new binary exactly |
| `folio-go/fontgen_matrix_test.go:fontgenManifestKeys` | `instance_faces.py` as text (duplicate reader, forced by the package boundary) | **no** | HIGH — identical regexes | same as above |
| `folio-go/fontgen_matrix_test.go:TestShippedFacesReproduceFromUpstream` | generator stdout + manifest + `Shipped()` | n/a | `witnessCount` needs `derived and compared N of M faces` | **`wantDerivedShippedFaces` 3 -> 7.** Its three-URL error message goes stale and must gain the new sources |
| `folio-go/shipped_faces_test.go:assertShippedFaceMatchesSpec` | spec table vs binaries | n/a | — | **`!strings.HasSuffix(names[6], "-Regular")` is a bare literal that reds on every new cut.** Parameterise it from the spec row; do not delete the check |
| `folio-go/shipped_faces_test.go:TestShippedSpecCoversEverythingShipped` | spec vs `testShippedFontSet()` | n/a | both directions | Seven spec rows + seven embeds/entries in `testfont_embed_test.go` |
| `folio-go/shipped_faces_ext_test.go` | `shippedPostScriptNames`, `shippedFaceFiles` vs `Shipped()` | n/a | cardinality fatals | A row in **both** maps per cut |
| `folio-go/fonts/fonts_test.go:TestShippedRobotoMatchesDesignerCatalogue` | Go bytes vs the designer's committed file | n/a | `minPlausibleBytes` floor 300,000 | **D-B: extend the one-cut discipline to the new cuts** — generalise from one hardcoded pair to a table, per-cut floor, negative control re-checked per cut |
| `folio-designer/src/font-binary-identity.test.ts:familiesWithNoRule` | `folio-go/fonts/fonts.go` **raw** + the browser generator | **NO** for `fonts.go` | HIGH — three regexes; the `//go:embed` one requires the `var` line **immediately** after | **THE MIRROR GUARD.** Every `Shipped()` key that is not a catalogue family must have an `@font-face` rule. This is what forces browser delivery in this story |
| same file | hardcoded counts | n/a | — | `toBe(4)` -> 11, `toBe(3)` -> 10, `perFile` `toBe(6)` -> 13 |
| `folio-designer/src/canvas-font-stack.test.ts` | `fonts.go` **raw** + `App.css` | NO for `fonts.go` | HIGH — same regexes, plus a stack **order** tie | `toBe(4)` / `toHaveLength(3)` move; check whether the fragment fallback stack must change (it should NOT — that would alter what every unattributed fragment falls back to) |
| `lint/internal/rules/fontsassets.go:expectedShippedFaces` | `fonts.go` line-by-line for `//go:embed ` | subject IS a comment | MEDIUM — **any pattern containing `*?[` is silently dropped** | Keep one literal directive per line. **A glob embed empties the expected set and fires `fonts-asset-missing`** |
| `lint/internal/rules/fontsassets.go:ScanFontsAssets` | **dir walk** of `folio-go/fonts` | n/a | — | Every committed cut needs its embed, or `fonts-asset-unaccounted`. **Note: hardcoded to `folio-go/fonts` — it never sees the designer side** |
| `lint/internal/manifest/manifest.go:ResolveAssets` | **whole-repo dir walk** | n/a | — | **This is the guard that covers the designer side.** Requires `LICENSE*` + `NOTICE*` + a `Copyright` line per directory; SPDX must be on the four-id font allowlist. Rows are **per file** |
| `lint/internal/manifest/manifest_test.go:TestManifestUpToDate` | committed `MANIFEST.md` vs live walk | n/a | whole-file compare | **+14 rows** (7 Go-side, 7 designer-side). Regenerate: `cd lint && go run ./cmd/genmanifest` — there is no root `go.mod`, so a repo-root invocation fails |
| `lint/internal/licence/licencecensus_test.go:pinnedCensus` | pinned LICENSE paths via `git ls-files` | n/a | exact path strings | **+14 rows**; a file with no pinned row is a `t.Errorf` |
| `folio-designer/scripts/build-wasm.mjs` | its own `shippedFamilies` + `shippedRules` | n/a | two build-time throws | `shippedFamilies` 6 -> 13; every member needs a rule; rule count must equal `shippedFamilies.length` |
| `folio-designer/src/font-catalogue.test.ts` | `font-catalogue.json`'s 31 entries | n/a | — | **Unchanged, and that is the point.** It iterates the catalogue, so the new cuts are outside it and the catalogue stays honestly Regular-only |

### Payload surface

- One `fingerprint()` call = one dist asset = one cache slot. Verified by decomposing the current
  `s1.assetCount` **54** exactly: 31 catalogue + 9 app chunks + 6 hardcoded shipped slots + 4 pdfjs +
  engine wasm + wasm-exec.js + starter.folio + index.html.
- **Being inside the wasm exempts nothing** — the three Notos are `//go:embed`'d *and* hold three of the
  six hardcoded slots, because a CSS `@font-face` needs a URL and the wasm's copy has none.
- `folio-designer/src/release-payload.ts` — `minimumCacheAssets = 10`, `maximumCacheAssets = 64`, each on
  its own line, read by `readDeclaredConstant`'s `^const <name> = (\d+)$` (digits only, throws on 0 or 2+).
- S1 shape is pinned in **three** places: `verify-offline-release.mjs` (`s1Ids` and `semanticLabels` by
  ordered exact join equality, `cachedRows.length !== 4`, **positional `s1.rows[4]`**, `cjk-font` must be
  the `Math.max` of all `*font` rows), `release-payload.ts` (literal union + `cached.length !== 4` +
  `rows[4]`), and the generator.
- `LoadScreen.tsx` renders `row.label` and `formatMiB(row.bytes)` per row; its total is
  `payload.cachedBytes`, and `verified` reduces over `payload.cacheAssets` — **never over `rows`**.

## Tasks & Acceptance

**Execution:**

- [x] `tools/fontgen/instance_faces.py` -- add four `UPSTREAM` entries (Noto Sans Bold / Italic /
      Bold Italic, Noto Sans Thai Bold) with the measured `out_sha256` and `out_bytes` from the Code Map,
      and `src_url` recording the same archive with the new `->` path for the italic source -- keeps the
      derivation replayable and the block shape parseable. Preserve `UPSTREAM = [` and `]` alone at column 0.
- [x] `folio-go/fonts/<seven new dirs>/` -- create one directory per cut, each holding the binary, its
      `LICENSE-OFL.txt` and a `NOTICE.md` in the exact parsed shape (four line-anchored rows, exactly one
      derivation statement, the archive digest) -- one binary per directory is hard-enforced.
- [x] `folio-go/fonts/fonts.go` -- add seven `//go:embed` directives, each immediately above its `var`,
      and seven `Shipped()` entries -- one literal directive per line; never a glob.
- [x] `folio-go/testfont_embed_test.go` -- mirror the seven embeds and map entries -- `package folio`
      cannot import `fonts`, so this hand-copy is forced.
- [x] `folio-go/shipped_faces_test.go` -- add seven `shippedFaceSpecs` rows from the measured metadata
      table; **parameterise the `-Regular` suffix check** from the row rather than deleting it, and put one
      line at `assertShippedFaceMatchesSpec` saying WHY: it is a **correct assertion quantified over a
      population that just changed**, not a stale check — the same shape as the shipped-slot metadata gap
      this story also closes, and deleting it would be the third instance of that defect in one story and
      the only one we introduced ourselves. Add the comment recording why the two italic PostScript names
      are correct output and that "correcting" them reds the suite.
- [x] `folio-go/shipped_faces_ext_test.go` -- add each cut to `shippedPostScriptNames` and `shippedFaceFiles`.
- [x] `folio-go/fontgen_matrix_test.go` -- move `wantDerivedShippedFaces` 3 -> 7 and extend the
      source-missing message to name the new sources and digests.
- [x] `folio-go/fonts/fonts_test.go` -- generalise `TestShippedRobotoMatchesDesignerCatalogue` from one
      hardcoded pair to a table over all four Roboto cuts, with a per-cut plausibility floor and the
      negative control re-checked per cut (D-B).
- [x] `folio-designer/public/fonts/<seven new dirs>/` -- commit the same seven binaries with their own
      `LICENSE-OFL.txt` and `NOTICE.md` -- the one-cut discipline compares against the designer's file.
- [x] `folio-designer/scripts/build-wasm.mjs` -- add seven `fingerprint()` slots, seven `@font-face` rules
      under the new family names, and the seven names to `shippedFamilies` -- both build-time throws move
      together. No catalogue entries.
- [x] `folio-designer/scripts/generate-offline-release.mjs` -- add seven `cached-asset` S1 rows.
- [x] **Key every S1 row access by id, at all FIVE positional sites** -- not "change 4 to 11". An index
      that has moved once will move again in 11.3. The file's own better convention is already there and
      must be adopted rather than invented: `verify-offline-release.mjs` reads
      `s1.rows.find((row) => row.id === 'cjk-font')`. The five sites:
      (a) `verify-offline-release.mjs` `const dictionaryRow = s1.rows[4]`;
      (b) `verify-offline-release.mjs` `cachedRows.length !== 4` -> 11;
      (c) `release-payload.ts` -- ONE line carrying THREE couplings,
      `cached.length !== 4 || rows[4].delivery !== 'embedded-in-engine' || rows[4].assetUrl !== rows[0].assetUrl`
      -- note `rows[0]` is itself positional, means "the engine wasm row", and says so nowhere;
      (d) `verify-offline-release.mjs` `redProof('s1-delivery-fiction', ...)` which mutates `rows[4]`;
      (e) `verify-offline-release.mjs` `redProof('s1-cloud-label', ...)` which mutates `rows[0]`.
      **(d) and (e) are the ones to stop on: a falsifier must target its subject by IDENTITY.** An
      index-keyed red proof that lands on the wrong row after an insertion either fails to go red at all,
      or goes red for the WRONG REASON — and a red proof passing for the wrong reason is this run's
      dominant defect class wearing the costume of the thing meant to catch it.
- [x] `folio-designer/src/release-payload.ts` -- extend the literal union, `ids` and `labels`; add
      `warnCacheAssets` as a third declared constant on its own line in `^const <name> = (\d+)$` form.
- [x] `folio-designer/scripts/verify-offline-release.mjs` -- warn when `assetCount` reaches
      `warnCacheAssets`, reading it through `readDeclaredConstant` -- never a literal in the verifier.
- [x] `folio-designer/src/LoadScreen.test.tsx` -- assert the displayed total equals Σ `cacheAssets` and
      not Σ `rows`, red-provable by making the total a row sum.
- [x] `folio-designer/src/font-binary-identity.test.ts`, `src/canvas-font-stack.test.ts` -- move the
      hardcoded counts (4 -> 11, 3 -> 10, 6 -> 13) and add the per-cut engine/designer digest pairs.
- [x] `folio-designer/src/font-catalogue.test.ts` -- add a metadata assertion over the **hardcoded-slot**
      population (not the catalogue), reusing `instanceOfFile` to assert each shipped face's **intended**
      subfamily and weight class against its own bytes, not a blanket Regular -- these thirteen faces have
      no metadata assertion today, and this story is what makes that gap load-bearing by putting the first
      non-Regular face into that population.
- [x] `lint/MANIFEST.md` -- regenerate via `cd lint && go run ./cmd/genmanifest`.
- [x] `lint/internal/licence/licencecensus_test.go` -- add the fourteen new `LICENSE-OFL.txt` rows.

**Acceptance Criteria:**

- Given the extended derivation, when `make fonts-verify` runs with the pinned interpreter, then all
  seven committed faces reproduce byte-for-byte and the witness reads `derived and compared 7 of 7 faces`.
- Given each new face, when the build runs, then it carries its own `NOTICE.md` recording the upstream
  release URL, the archive sha256 and the source sha256, and its own OFL text, and `lint` accounts for it
  — a face shipping without a licence record fails the build.
- Given the enlarged shipped set, when the offline release is generated, then the manifest states each new
  face's measured byte cost, the load screen itemises each as its own row, and the build warns that the
  cache-asset margin has fallen to 3.
- Given the corpus, no document of which declares bold or italic, when it is rendered on all four targets,
  then the bytes are identical to before this story and every golden digest is unchanged.
- Given a `Shipped()` key, when any code needs the family, then it reads the face's own **sfnt name ID 1**
  and never parses the key. `shippedFaceSpecs.Family` is the TEST that asserts name ID 1 against the
  binary — it lives in `folio-go/shipped_faces_test.go` and no non-test package can import it.

## Spec Change Log

Empty. Append-only; populated by step-04 on a bad_spec loopback.

## Design Notes

**Why the browser gets the bytes in this story, and Q2's ruling is superseded.** The orchestrator ruled
Arm A — commit the cuts here, bundle them in 11.3, and give them `embedded-in-engine` S1 rows — with an
explicit falsifier: flip to Arm B if a guard requires the Go-side face set and the browser side to agree.
**The falsifier fires.** `font-binary-identity.test.ts`'s *"gives every declared family a source file of
its own"* reads the `Shipped()` keys out of `fonts.go` and requires every non-catalogue key to have an
`@font-face` rule, resolving through a real `assets` slot. Arm A would leave that mirror half-populated
and red the suite. So the cuts are bundled here, the rows are **`cached-asset`**, and the earlier
`embedded-in-engine` ruling is superseded — measured at this plan gate, by the builder, under the
falsifier the orchestrator attached to the ruling.

**Consequences, stated rather than discovered later.** `s1.assetCount` 54 -> **61** against
`maximumCacheAssets` 64: the margin goes **10 -> 3**, in this story rather than in 11.3. That is why the
approach-warning ships here. These are predictions from the one-slot-per-fingerprint rule and must be
**measured, not assumed**, in the Delivery Log. `cjk-font` remains the `Math.max` of the `*font` rows
(CJK brotli 4,948,312 against a largest new cut of 665,508 raw), so that invariant survives untouched.

**The load-screen test survives the arm flip and is worth more after it.** The new rows are `cached-asset`
and therefore correctly additive; the `thai-dictionary` row remains the one `embedded-in-engine` row, and
its non-additivity is real but asserted nowhere at the surface an author reads. Three mechanisms make it
true today — `parseS1Payload` ties `cachedBytes` to Σ`cacheAssets`; `LoadScreen` reduces over
`cacheAssets`, not `rows`; the generator computes `s1VisibleBytes` before appending the dictionary — and
none of them would stop someone rewriting the displayed total as a row sum. With twelve rows instead of
five, that mistake gets worse, not better.

**Payload disclosure for 8.4d.** The seven cuts add **3,136,768 raw bytes twice** — once inside the engine
wasm and once as seven browser assets — because every shipped face is already duplicated that way. Record
`brotli.totalBytes` before and after. **Build the same arm twice first and quote that repeat as the
control**; `-buildvcs=false` is passed via `ENGINE_BUILD_FLAGS` and `assertNoVCSStamp` verifies no stamp
survives, so the tree-state input is closed and the figure should be stable. If the repeat is not
byte-stable, that is a finding, because something is supposed to have closed that input.

**AD-26 note for whoever touches designer assets next.** `fonts-asset-unaccounted` is hardcoded to
`folio-go/fonts` and never sees the designer side; `manifest.ResolveAssets` is the guard that actually
covers it, by walking the whole repo. Recorded here so the next story does not re-derive it.

**A builder forbidden to touch git cannot green a gate that reads `git ls-files`, and the failure can go
either way.** Two of this story's guards read the git index, so the fourteen new directories were
invisible to them until the orchestrator staged the files. The gate order for any file-adding story is
therefore: implement -> orchestrator stages -> licence and manifest gates.

The direction we hit was the **loud** one. `TestLicenceSignalCensus` compares a **hand-pinned** population
against a **walked** one, so the seventy-three rows written from intent disagreed with the fifty-nine the
walk could see, and it failed with a message naming both numbers.

The **silent** direction is the dangerous one and it is live for the next story. `TestManifestUpToDate`
compares the committed `MANIFEST.md` against a live walk, and **both sides derive from the same
git-scoped walk**. Regenerate before staging and the generated manifest omits the fourteen rows, the walk
omits them too, the comparison agrees, and the test goes **green while `MANIFEST.md` silently fails to
account for fourteen redistributed font binaries** — an AD-26 breach that passes.

So the loud failure we got was a property of how that particular guard is built, not of the mistake:
**a record compared only against itself cannot detect a blind spot it shares with its own source.**
Registered as its own deferred entry.

**A near-miss from the same family, pointing the other way.** Counting `//go:embed` in `fonts.go` with a
plain substring match returns **12**; the twelfth is prose reading `// estimated) - go:embed`. The real
count is eleven. It resolves clean only because `lint`'s `expectedShippedFaces` trims each line and
prefix-matches `"//go:embed "`, which that prose does not satisfy — **the repo's matcher was the strict
one and the ad-hoc check was the loose one**, which is the inverse of the usual finding and a concrete
reason the guard census tracks comment-stripping as a property.

## Verification

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected rc 1 with **exactly two** failures,
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child. Baseline **2113 pass / 2 fail / 5 skip**,
  measured at `1c0260d` and unchanged at `46ce9e5` (artifacts-only commit). A third failure is a hard stop — report before triaging.
- `cd folio-go && FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python go test -tags=matrix -count=1 ./...`
  -- **the FULL tagged suite, unfiltered, four targets, in this story.** Baseline **2125 / 2 / 5**.
  **DW-12 requires this per-story and forbids deferring it to the epic boundary**, and it composes with
  D-000.11: `TestShippedFacesReproduceFromUpstream` is among the ten tests CI never runs, and this story
  changes that guard's own subject. **Never a `-run` filter and never CI's subset** — a green from a
  narrower run discharges nothing. Without `FOLIO_FONTGEN_PYTHON` the test degrades to a
  could-not-execute rather than a failure, and a green from the degraded form discharges nothing either.
  The bare `python3` here is Homebrew 3.14 without fontTools; `make fonts` / `make fonts-verify` need
  `PYTHON=` overridden for the same reason.
- `cd folio-go && go vet -tags=matrix ./...` -- rc 0.
- `gofmt -l folio-go` -- **from the repo root**, empty. Run inside `folio-go/` it prints an `lstat` error
  that reads like clean.
- `cd lint && go test -count=1 ./...` -- four packages ok. **`-count=1` always**: the rules package walks
  directories and Go's test cache does not track that, so a cached `ok` is no measurement.
- `cd folio-designer && npm test` -- baseline **63 files / 938 tests**, all passing; expect growth.
- `cd folio-designer && npx tsc -b --force` -- rc 0. (`npx tsc --noEmit` typechecks zero files at the root.)
- `cd folio-designer && npx oxlint` -- **exactly 4** `react(only-export-components)` warnings and no other
  rule. **Prove the SET, not the count**: `preview/pdf-viewer.tsx:16` and `:17`, `App.tsx:3240` and `:3247`.
  The lines move; the file set, rule and count are the invariant.
- `cd folio-designer && npm run scan:font-hosts` (638 files, 0 occurrences) and `npm run scan:host-fonts`
  (152 files, 0 occurrences).
- `cd folio-designer && npm run build && npm run verify:offline:red && npm run verify:offline:wasm` --
  node must be exactly `v24.16.0`. Record `brotli.totalBytes`, `s1.assetCount` and the margin warning.

**Manual checks:**
- Confirm `git status --porcelain` shows no unintended file, and that `.font-sources/` stays gitignored.
- Confirm the Delivery Log records: the before/after `brotli.totalBytes` with the same-arm repeat control,
  the measured `s1.assetCount` 54 -> 61 and margin 10 -> 3, the +2.99 MiB / +26.9% shipped-set delta
  against D-A's "~3 MB" estimate, and that Noto Sans SC's exclusion is stated with its 10,595,932-byte
  figure.

## Suggested Review Order

**The derivation, and what makes a face reproducible**

- Entry point: four new entries turn a 3-face manifest into 7, same replayable script.
  [`instance_faces.py:128`](../../tools/fontgen/instance_faces.py#L128)

- The pin that fails loudly if the derived set and the shipped set disagree.
  [`fontgen_matrix_test.go:264`](../../folio-go/fontgen_matrix_test.go#L264)

- The provenance shape every new face must match; four rows parsed, exactly one each.
  [`notosans-bold/NOTICE.md:44`](../../folio-go/fonts/notosans-bold/NOTICE.md#L44)

**The engine face set, and the naming discipline 11.2 inherits**

- Eleven keys, seven new; the key is readable and nothing may parse it.
  [`fonts.go:150`](../../folio-go/fonts/fonts.go#L150)

- Why the family comes from sfnt name ID 1 and never from the key.
  [`fonts.go:88`](../../folio-go/fonts/fonts.go#L88)

- Seven spec rows carrying measured metadata, not requested metadata.
  [`shipped_faces_test.go:143`](../../folio-go/shipped_faces_test.go#L143)

- The `-Regular` literal parameterised from the row rather than deleted.
  [`shipped_faces_test.go:249`](../../folio-go/shipped_faces_test.go#L249)

- One-cut discipline generalised to all four Roboto cuts (D-B).
  [`fonts_test.go:50`](../../folio-go/fonts/fonts_test.go#L50)

**Payload disclosure — the surface that was pinned three ways**

- Seven new `cached-asset` rows; per-face cost, not an aggregate.
  [`generate-offline-release.mjs:102`](../../folio-designer/scripts/generate-offline-release.mjs#L102)

- Twelve ordered ids; cached-row count now derived, not a literal.
  [`verify-offline-release.mjs:140`](../../folio-designer/scripts/verify-offline-release.mjs#L140)

- Row access keyed by id at every site, including both red-proof falsifiers.
  [`release-payload.ts:104`](../../folio-designer/src/release-payload.ts#L104)

- The margin warning, declared as a constant so the verifier cannot restate it.
  [`release-payload.ts:57`](../../folio-designer/src/release-payload.ts#L57)

- Extracted so a test can execute the warning rather than only its reader.
  [`verify-offline-release.mjs:53`](../../folio-designer/scripts/verify-offline-release.mjs#L53)

**The browser declaration, and the mirror that forced it**

- Thirteen families; each cut is its own `@font-face`, never a weight axis in CSS.
  [`build-wasm.mjs:191`](../../folio-designer/scripts/build-wasm.mjs#L191)

- Script fallbacks validated against the six Regulars, so none can resolve to a bold cut.
  [`build-wasm.mjs:253`](../../folio-designer/scripts/build-wasm.mjs#L253)

- The guard that made Arm A impossible: every engine key needs a browser rule.
  [`font-binary-identity.test.ts:995`](../../folio-designer/src/font-binary-identity.test.ts#L995)

**Peripherals — the guards this story had to move or add**

- Per-face intended subfamily, weight and oblique over the previously unchecked slots.
  [`font-catalogue.test.ts:64`](../../folio-designer/src/font-catalogue.test.ts#L64)

- The displayed total must sum cacheAssets, never rows.
  [`LoadScreen.test.tsx:75`](../../folio-designer/src/LoadScreen.test.tsx#L75)

## Delivery Log

### 2026-09-06 — done

Baseline `ffec48c`. Shipped in **`429cb1a`** on `main` — 66 files, +4538 / −232 — local and unpushed at
this close. Decisions D-11.1.1 through D-11.1.25 live in
[`epic-11-14-decision-log.md`](./epic-11-14-decision-log.md) and are not restated here. This closer
touched only this story file, `sprint-status.yaml` and `deferred-work.md`, and staged nothing.

**Seven faces, and seven is the whole realizable set.** Noto Sans and Roboto each gained Bold, Italic and
Bold Italic; Noto Sans Thai gained **Bold only**, because upstream publishes no Thai italic at all. D-A's
"nine new faces" was arithmetically impossible, not merely unambitious — that half of D-A rested on a
false premise and the story measured it away at the plan gate rather than inheriting it. **Noto Sans SC
stays Regular deliberately**: its Regular alone is **10,595,932 bytes**, and three instances of it would
take the offline payload from roughly 11 MB to roughly 45 MB. Anyone counting families must be able to
find that reason without asking, which is why it sits in the NOTICEs as well as here.

**The payload cost, and an estimate that earned its keep.** The seven cuts add **3,136,768 B — 2.99 MiB,
+26.9%** — taking the shipped set **11,645,836 → 14,782,604** raw. Re-measured at this close directly
from the eleven committed binaries: the seven sum to 3,136,768 exactly and the eleven to 14,782,604
exactly. **This lands almost exactly on D-A's "~3 MB" estimate**, and that is worth saying out loud: an
estimate checked against the outcome and found good is how the *next* estimate earns the right to be
believed. Most of this run's figures have been corrections; this one is a confirmation.

**The cache-asset margin moved in this story, not in 11.3.** `s1.assetCount` **54 → 61** against
`maximumCacheAssets` 64, so the margin goes **10 → 3**; twelve S1 rows, `cachedBytes` **53,939,356**. The
approach warning ships here and **fires**, naming its own threshold (`warnCacheAssets` = 56), its file,
and the fact that nothing fails until the maximum is exceeded. Re-read at this close from the release
manifest the orchestrator built at this tree: 61 assets, 12 rows, 53,939,356 cached bytes, and
`brotli.totalBytes` **19,012,573**.

**Non-additivity is now empirical rather than argued.** Σ `rows` = **17,238,271** against Σ `cacheAssets`
= **53,939,356**, a factor of three apart — so a row-sum total could never have masqueraded as correct,
and the load-screen assertion pins the one surface where the mistake would otherwise go unseen. Both
figures re-derived at this close from the manifest, independently of the build's own report.

**The control, and the limit that is part of it.** Two consecutive clean builds were byte-identical on
`assetCount`, `cachedBytes`, every row and every asset. **Both ran at a FIXED TREE (HEAD `ffec48c`)** —
this is reproducibility at one tree and **not** a cross-commit result; it must not be read as one.
Separately and more durably, `strings` over the emitted wasm returns **zero** vcs markers, so the
commit-hash input is closed *in the artifact* rather than merely declared in a flag. Note that the
before/after `brotli.totalBytes` pair the Verification section asks for was **superseded by D-11.1.11**,
which replaced the cross-arm delta with this same-arm control; only the after figure exists, recorded
above. No baseline release build was run at this close.

**Procurement: five upstream releases priced, cost zero.** Both sources were already inside archives the
repo names, and the one new source file is a sibling path in an archive `UPSTREAM` already pins. Archive
digests: Roboto `1653dbe1…` **directly verified**; Noto Sans `0c34df07…` and Noto Sans Thai `af889cc6…`
(`NotoSansThai-v2.002.zip`, 4,720,990 B) both **transitively verified**, via the extracted VF matching the
pinned `src_sha256`. **Transitive is a weaker guarantee than Roboto's and the two are not parity** — the
extracted file is confirmed, the archive around it is inferred. Stated so no later reader flattens the
three into one standard.

**For whoever next opens a font NOTICE.** Noto Sans **Thai Regular**'s NOTICEs still carry **no archive
digest**, and it is now backfillable from `af889cc6…` — same archive, same release, **no re-fetch
needed**. That is a two-minute job sitting behind a value that already exists in the tree, and it will
stop being obvious the moment this story scrolls out of view.

**Review triage: 13 patched, 5 deferred, 2 escalated and both ruled, 0 loopbacks**
(`review_loop_iteration` 0). The review's own best catch was **not a guard gap but a shipped rendering
defect**: script fallbacks were validated against all thirteen families, so a fallback could have named a
bold cut and rendered an entire script bold in every author's document. Three further patches were the
run's dominant defect class found in guards *this story had just introduced* — a coverage witness that
could not fire by construction, an acceptance criterion's sole realization that no test executed, and a
falsifier that consumed the signal it existed to prove.

**A register debt this close paid.** Of the five deferrals, one was written up as **DW-231** and the other
four had been appended to `deferred-work.md` as raw, unnumbered `source_spec` blocks lodged inside
DW-231's body — filed in substance, unfindable in practice. They are now **DW-233 through DW-236**, with
owners, severities and discharge conditions, and the raw dump has been removed from DW-231. **DW-232** was
filed separately by the same review. Nothing was invented and no evidence text was dropped; the four
entries carry their original wording.

**Measured gates at this close**, re-run rather than carried forward:

| Gate | Result |
|---|---|
| `folio-go go test -count=1 ./...` | **2124 pass / 2 fail / 5 skip**, rc 1 |
| — the two failures | `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — **pre-existing baseline reds**, mandated unmet by D-000.17 / D-2.1.14 |
| `lint go test -count=1 ./...` | **227 pass**, four packages ok, rc 0 |
| `folio-designer npm test` | **64 files / 953 tests**, all pass, rc 0 |
| `folio-designer npx tsc -b --force` | rc 0, clean |
| `make fonts-verify` (pinned interpreter) | `derived and compared 7 of 7 faces`, `fontgen: OK` |

**Not re-run at this close, and green when the orchestrator ran them at commit time:** the full unfiltered
`-tags=matrix ./...` four-target suite (green but for the same two baseline reds, with
`TestShippedFacesReproduceFromUpstream` **PASS** on witness `derived and compared 7 of 7 faces` — 7
derived + 4 static-upstream = 11), and the offline release build. Neither is re-measured here; both are
the orchestrator's figures, not mine.

**What this story deliberately did not do.** It resolves no weight and paints none — 11.2 resolves, 11.3
paints — so a bold document still renders in book weight at this commit and every golden digest is
unchanged. No cut got a catalogue entry, because the catalogue asserts every face is upright Regular 400.
The canvas fragment fallback stack did not move, only the counts around it.
