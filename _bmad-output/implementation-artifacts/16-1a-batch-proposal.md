---
title: 'Story 16.1a — batch proposal (TASK 1 survey output)'
type: 'gate-artifact'
created: '2026-09-03'
status: 'awaiting-gate'
story: '_bmad-output/implementation-artifacts/16-1a-the-local-face-tier-covers-the-head-of-the-library.md'
---

# The local face tier batch — proposal awaiting the gate

**This document is the gate.** It is produced from committed artifacts alone, with **no network
request of any kind**. No upstream fetch, no licence classification and no byte has been attempted,
and none will be until the membership list below is agreed. The per-family question TASK 2 asks —
*does this family's own project release publish a static Regular, under which of the four admitted
identifiers?* — is deliberately unanswered here, because it is the only part that needs the network
and it is not worth asking about families that will not fit.

## Provenance of this measurement (D-8.4j.8)

| Field | Value |
|---|---|
| Command | the script under **Replaying this survey** below, run with `node` |
| Commit | `1632bbbb624050b4b115e614944ccddf2e5ed221` (story baseline `efd79bfc41cfb9ed45dd4a6223da38e83c00797b`; the intervening commits touch `_bmad-output/` only) |
| Tree state | One modified path — this story's own spec file (`status: in-progress`). No source, catalogue, index or build input is dirty. |
| Working directory | `/Users/panitw/Projects/folio/folio-designer` |
| Inputs | `font-index.json`, `font-catalogue.json`, `scripts/build-wasm.mjs`, `dist/offline-release-manifest.json` |
| Index snapshot | `font-index.json` `snapshotDate` **2026-09-03**, **1,811** families (last changed at `d6d51f1`, Story 16.1) |
| Release manifest | `dist/offline-release-manifest.json` — untracked build output of the baseline tree |
| Network | **None.** Every input is a file in the working tree. |

**Method, stated because one wrong way of doing it passed silently during planning.** Variable-only is
derived as **`axes.length > 0`**, never as `row.variable`: the snapshot does not store `variable` — it
is synthesised at emit time in `scripts/build-font-index.mjs:124` — so a survey reading `row.variable`
gets `undefined` for all 1,811 rows, reports zero variable-only families and passes without complaint.
`shippedFamilies` is **parsed out of `scripts/build-wasm.mjs`'s literal list**, not retyped, so a
rename there cannot silently unblock a collision here.

**Rank and `popularity` are two different numbers, and both are shown.** The snapshot's `popularity`
field is **not a dense rank** — 1,811 families carry only 1,100 distinct values, so ties are common and
they matter at the cut. *Rank* below is the position in the ascending sort; *popularity* is the field's
own value. The story's Code Map cites `Noto Sans` at "rank 12" and `IBM Plex Sans` at "rank 19", which
are positions; their `popularity` values are 8 and 15.

## The arithmetic

| Step | Count |
|---|---|
| Families in the committed index snapshot | 1,811 |
| Top 50 by `popularity`, variable-only (`axes.length > 0`) | **38** |
| − families the local face tier already holds (`Roboto`, `Inter`, `Noto Serif`, `Space Grotesk`) | 34 |
| − `shippedFamilies` collisions (`Noto Sans`, `IBM Plex Sans`) | **32 admissible** |
| Cache slots in use (`s1.assetCount`, of which **21** are `catalogue-*.ttf`) | **44** |
| `maximumCacheAssets` (`src/release-payload.ts:33`) | **64** |
| Free slots — the batch ceiling | **20** |

`44 + N ≤ 64`, therefore **N ≤ 20**, and **32 candidates want 20 slots**. The batch is
**ceiling-bound, not taste-bound**, and that gap is the single fact the gate exists to rule on.
Raising `maximumCacheAssets` is refused by the story's own Block If.

## The membership rule proposed for agreement

> **The most popular N admissible families that fit, ranked ascending by the index snapshot's
> `popularity` field, ties broken by family name ascending.**

It is a **ranking** rule, not a **threshold** rule, and that is deliberate: *"everything above
popularity P"* silently overflows the 64-slot ceiling the next time popularity moves, whereas *"the
most popular N"* cannot, because N is the constraint. Admissibility is the three subtractions above:
variable-only on the `google/fonts` mirror, not already in the local face tier, not a
`shippedFamilies` name.

**Ties are real in this data and one of them straddles the cut** — `Google Sans`/`Roboto Mono` at
popularity 3, `Lora`/`Roboto Condensed` at 13, `Oswald`/`Plus Jakarta Sans` at 14,
`Fraunces`/`Heebo`/`Nunito Sans` at 19, and **`Instrument Sans`/`Rubik` at 22, which is exactly the
N = 20 boundary**. Without the name tie-break the batch's twentieth member would depend on JSON
iteration order, so the rule names it.

## The candidate table

Ranked in the rule's own order. The **slot column is cumulative**: it is the release's `assetCount`
if the batch were cut at that row. Everything from row 21 down is over the ceiling and is listed so
the gate can see what is being left out rather than being told it is nothing.

| # | Family | Rank | Popularity | Category | Subsets | Axes on the mirror | `44 + N` | Verdict |
|---|---|---|---|---|---|---|---|---|
| 1 | Open Sans | 1 | 2 | Sans Serif | cyrillic, cyrillic-ext, greek, greek-ext, hebrew, latin, latin-ext, math, symbols, vietnamese | `wdth,wght` | 45 ≤ 64 | **fits** |
| 2 | Google Sans | 3 | 3 | Sans Serif | armenian, bengali, canadian-aboriginal, cyrillic, cyrillic-ext, devanagari, ethiopic, georgian, greek, greek-ext, gujarati, gurmukhi, hebrew, khmer, lao, latin, latin-ext, malayalam, oriya, sinhala, symbols, tamil, telugu, thai, vietnamese | `GRAD,opsz,wght` | 46 ≤ 64 | **fits** |
| 3 | Roboto Mono | 4 | 3 | Monospace | cyrillic, cyrillic-ext, greek, latin, latin-ext, vietnamese | `wght` | 47 ≤ 64 | **fits** |
| 4 | DM Sans | 6 | 5 | Sans Serif | latin, latin-ext | `opsz,wght` | 48 ≤ 64 | **fits** |
| 5 | Montserrat | 8 | 6 | Sans Serif | cyrillic, cyrillic-ext, latin, latin-ext, vietnamese | `wght` | 49 ≤ 64 | **fits** |
| 6 | Arimo | 9 | 7 | Sans Serif | cyrillic, cyrillic-ext, greek, greek-ext, hebrew, latin, latin-ext, vietnamese | `wght` | 50 ≤ 64 | **fits** |
| 7 | Roboto Slab | 13 | 11 | Serif | cyrillic, cyrillic-ext, greek, greek-ext, latin, latin-ext, vietnamese | `wght` | 51 ≤ 64 | **fits** |
| 8 | Lora | 15 | 13 | Serif | cyrillic, cyrillic-ext, latin, latin-ext, math, symbols, vietnamese | `wght` | 52 ≤ 64 | **fits** |
| 9 | Roboto Condensed | 16 | 13 | Sans Serif | cyrillic, cyrillic-ext, greek, greek-ext, latin, latin-ext, vietnamese | `wght` | 53 ≤ 64 | **fits** |
| 10 | Oswald | 17 | 14 | Sans Serif | cyrillic, cyrillic-ext, latin, latin-ext, vietnamese | `wght` | 54 ≤ 64 | **fits** |
| 11 | Plus Jakarta Sans | 18 | 14 | Sans Serif | cyrillic-ext, latin, latin-ext, vietnamese | `wght` | 55 ≤ 64 | **fits** |
| 12 | Jost | 20 | 16 | Sans Serif | cyrillic, latin, latin-ext | `wght` | 56 ≤ 64 | **fits** |
| 13 | Raleway | 21 | 17 | Sans Serif | cyrillic, cyrillic-ext, latin, latin-ext, vietnamese | `wght` | 57 ≤ 64 | **fits** |
| 14 | Nunito | 24 | 18 | Sans Serif | cyrillic, cyrillic-ext, latin, latin-ext, vietnamese | `wght` | 58 ≤ 64 | **fits** |
| 15 | Fraunces | 25 | 19 | Serif | latin, latin-ext, vietnamese | `SOFT,WONK,opsz,wght` | 59 ≤ 64 | **fits** |
| 16 | Heebo | 26 | 19 | Sans Serif | hebrew, latin, latin-ext, math, symbols | `wght` | 60 ≤ 64 | **fits** |
| 17 | Nunito Sans | 27 | 19 | Sans Serif | cyrillic, cyrillic-ext, latin, latin-ext, vietnamese | `YTLC,opsz,wdth,wght` | 61 ≤ 64 | **fits** |
| 18 | Playfair Display | 30 | 20 | Serif | cyrillic, latin, latin-ext, vietnamese | `wght` | 62 ≤ 64 | **fits** |
| 19 | Libre Baskerville | 31 | 21 | Serif | latin, latin-ext | `wght` | 63 ≤ 64 | **fits** |
| 20 | Instrument Sans | 32 | 22 | Sans Serif | latin, latin-ext | `wdth,wght` | 64 ≤ 64 | **fits** |
| 21 | Rubik | 33 | 22 | Sans Serif | arabic, cyrillic, cyrillic-ext, hebrew, latin, latin-ext | `wght` | 65 > 64 | over ceiling |
| 22 | Dancing Script | 39 | 25 | Handwriting | latin, latin-ext, vietnamese | `wght` | 66 > 64 | over ceiling |
| 23 | Karla | 40 | 25 | Sans Serif | latin, latin-ext | `wght` | 67 > 64 | over ceiling |
| 24 | Manrope | 41 | 25 | Sans Serif | cyrillic, cyrillic-ext, greek, latin, latin-ext, vietnamese | `wght` | 68 > 64 | over ceiling |
| 25 | Josefin Sans | 42 | 26 | Sans Serif | latin, latin-ext, vietnamese | `wght` | 69 > 64 | over ceiling |
| 26 | Merriweather | 43 | 27 | Serif | cyrillic, cyrillic-ext, latin, latin-ext, vietnamese | `opsz,wdth,wght` | 70 > 64 | over ceiling |
| 27 | Public Sans | 44 | 27 | Sans Serif | latin, latin-ext, vietnamese | `wght` | 71 > 64 | over ceiling |
| 28 | Outfit | 45 | 28 | Sans Serif | latin, latin-ext | `wght` | 72 > 64 | over ceiling |
| 29 | Sora | 47 | 29 | Sans Serif | latin, latin-ext | `wght` | 73 > 64 | over ceiling |
| 30 | Syne | 48 | 29 | Sans Serif | greek, latin, latin-ext | `wght` | 74 > 64 | over ceiling |
| 31 | Work Sans | 49 | 29 | Sans Serif | latin, latin-ext, vietnamese | `wght` | 75 > 64 | over ceiling |
| 32 | Lexend | 50 | 30 | Sans Serif | latin, latin-ext, vietnamese | `wght` | 76 > 64 | over ceiling |

**Rows 21–32 are excluded by the ceiling alone**, not by any judgement about the families. If the gate
admits fewer than 20 they move up; they never move up on merit.

## The Thai variable-only families, named and dispositioned

TASK 1 requires all six by name. There are exactly six in the whole snapshot, and **only three are
genuine candidates**:

| Family | Rank | Subsets | Disposition |
|---|---|---|---|
| `Google Sans` | 3 | …, thai, … (25 subsets) | **Candidate** — row 2 of the table above; the only Thai-covering family anywhere near the head |
| `Noto Sans Thai` | 207 | latin, latin-ext, thai | **Structurally unaddable** — a `shippedFamilies` name; a catalogue row for it throws at `build-wasm.mjs:214` |
| `Anuphan` | 501 | latin, latin-ext, thai, vietnamese | **Candidate in principle**, but far outside the top 50 and therefore outside this batch's admissible set |
| `Noto Serif Thai` | 635 | latin, latin-ext, thai | **Already local face tier** — a no-op by the contract's own matrix |
| `Noto Sans Thai Looped` | 1101 | latin, latin-ext, thai | **Already local face tier** — a no-op |
| `Playpen Sans Thai` | 1222 | emoji, latin, latin-ext, math, thai | **Candidate in principle**, outside the top 50 and outside this batch |

**Thai coverage does not depend on this batch.** `Noto Sans Thai` ships as an engine face and is the
script's fallback; the batch is a palette, not coverage.

## Two families in the bundle that this batch cannot offer

`Noto Sans` (rank 12, popularity 8) and `IBM Plex Sans` (rank 19, popularity 15) are variable-only on
the mirror, are **not** in the local face tier, and their bytes **already ship** as chrome faces. They
are still not offerable: the browser filters their index rows out for being variable, and a catalogue
row naming either throws at `build-wasm.mjs:214` (mirrored as an assertion at
`font-catalogue.test.ts:295`), because the collision guard exists to stop two `@font-face` rules
declaring one family.

**This is not a bug to route around and not this story's to fix.** Making them pickable means giving
the shipped tier a browser presence, which is a different story. It should be registered, not solved
here.

## Two local faces with no index row, confirming the join

`Inter Display` and `Source Serif 4 Display` have **no row in the snapshot at all**, so 2 of the 21
local faces are unjoinable under any normalisation. Per D-16.R.3 that is correct behaviour, not a
defect, and it is why the join key is exact `family` string equality.

## What the gate is being asked to rule on

1. **N.** The ceiling permits at most 20. Filling it exactly leaves **zero** margin against
   `maximumCacheAssets`, and nothing in the build reports that margin (DW-162) — so the next asset
   added anywhere in the release, by any story, fails the offline-release contract with no warning
   beforehand. A smaller N buys headroom at the cost of families. **This is a judgement about risk
   appetite, not a measurement, which is why the build is not making it.**
2. **The membership rule** as worded above, including the name tie-break.
3. **The owner and the re-run trigger** that TASK 7 must write into
   `_bmad-output/specs/spec-fonts/font-catalogue.md`. The batch is a snapshot of a distribution that
   moves; D-16.R.2 states plainly that *"a batch nobody owns is the failure mode this decision exists
   to prevent"*. **No name is proposed here because the build cannot appoint one.**

**Every admitted family is still conditional.** TASK 2 may drop any of them on evidence — its own
upstream publishing no static Regular, a `LICENSE*` that will not classify to one of the four
identifiers (OFL-1.1, Apache-2.0, MIT, UFL), or a `cmap` that disagrees with its declared `scripts`
in either direction. Rows 21–32 are the **reserve**, taken in the same rank order, and a family
dropped for cause is recorded with its evidence rather than derived silently.

## Replaying this survey

Offline and self-contained. Run with `node` from anywhere; it reads the tree by absolute path.

```js
import { readFileSync } from 'node:fs'
const root = '<repo>/folio-designer'
const index = JSON.parse(readFileSync(`${root}/font-index.json`, 'utf8'))
const catalogue = JSON.parse(readFileSync(`${root}/font-catalogue.json`, 'utf8'))
const buildWasm = readFileSync(`${root}/scripts/build-wasm.mjs`, 'utf8')
// Parsed, never retyped: a rename in build-wasm.mjs must not silently unblock a collision here.
const shipped = new Set(JSON.parse('[' + /const shippedFamilies = \[([^\]]*)\]/.exec(buildWasm)[1].replace(/'/g, '"') + ']'))
const local = new Set(catalogue.map((r) => r.family))
// Deterministic order: popularity ascending, then family name ascending.
const families = [...index.families].sort((a, b) => a.popularity - b.popularity || a.family.localeCompare(b.family))
// Variable-only is axes.length > 0. The snapshot does not store `variable`.
const admissible = families.slice(0, 50)
  .filter((r) => r.axes.length > 0 && !local.has(r.family) && !shipped.has(r.family))
const manifest = JSON.parse(readFileSync(`${root}/dist/offline-release-manifest.json`, 'utf8'))
console.log(admissible.length, manifest.s1.assetCount, 64 - manifest.s1.assetCount)
```

Expected at this commit: `32 44 20`.

## Status

**HALT — `blocked`, blocking condition `batch proposal awaiting gate`.** No upstream fetch has been
made and none will be until this list is agreed.
