---
title: 'Story 16.1a: The local face tier covers the head of the library'
type: 'feature'
created: '2026-09-03'
status: 'blocked'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: 'efd79bfc41cfb9ed45dd4a6223da38e83c00797b'
baseline_revision: '1632bbbb624050b4b115e614944ccddf2e5ed221'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
warnings: ['oversized']
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the intent contract below governs implementation.*

The font browser can offer an author almost two thousand typeface families, but not the ones they are
most likely to want. Roboto, Open Sans, Inter, Montserrat — thirty-eight of the fifty most popular —
are published on Google's mirror in a form this product deliberately does not accept, because it holds
every weight in one file and picking a weight out of it at the last moment would make the same
template print differently on different machines.

There is a cheaper answer than it first appeared. Those families publish perfectly ordinary
single-weight files from their *own* project releases; only the mirror lacks them. This product already
carries twenty-one typefaces exactly that way, each with its licence, its copyright and a record of the
precise file it came from. This story adds a batch of the most-wanted families the same way, so the
browser can offer them from the machine rather than refusing them.

**The batch cannot be as large as the want.** Measured this dispatch: thirty-two of those families are
addable in principle, and the offline release has room for **twenty more files, total**. So the batch is
limited by shelf space, not by taste, and the first thing this story does is say out loud how many fit
and which ones make the cut — and stop, for a human to agree the list, before a single byte is fetched.

What it deliberately does not do: it does not convert anything, it does not add a general mechanism for
getting statics, and it does not promise the whole library. It is a curated batch, and the story names
its size, its membership rule and who re-runs it — because a batch nobody owns is the failure this work
exists to prevent.

Done looks like: the families an author is most likely to type are offered and embed with no download
at all, each carrying its own terms.

<intent-contract>

## Intent

**Problem:** D-16.R.2 (owner) funded a bounded batch so the font browser does not refuse the head of the
popularity distribution — measured, **37 of the top 50 families are variable-only** on the `google/fonts`
mirror and therefore unaddable. D-16.R.2a then corrected the mechanism: the batch is **not** a
`tools/fontgen` derivation queue. Measured, `instance_faces.py` drives a hardcoded three-entry
`UPSTREAM` list of **engine** faces, and **none of the 21 designer catalogue faces is derived** — every
`NOTICE.md` says *"NO DERIVATION APPLIES."* This repository already ships **Roboto** and **Inter** as
byte-for-byte upstream statics from `googlefonts/roboto-classic` and `rsms/inter` v4.1.

**Approach:** Grow the **local face tier** by the mechanism this repository has executed 21 times: take
each family's static Regular from **that family's own upstream release**, commit it with its unmodified
`LICENSE*`, write its `NOTICE.md` recording both the upstream archive digest and the committed digest,
and add its `font-catalogue.json` row with `scripts` verified against the binary's own `cmap`.
Derivation stays available for a family whose own upstream publishes no static — now the exception, not
the plan.

## Boundaries & Constraints

**Always:**
- **The provenance regime is copied exactly, not approximated.** Per face: the unmodified upstream
  `LICENSE*`; a `NOTICE.md` recording the upstream project, the release, the **source** SHA-256 and the
  **committed** SHA-256; the `font-catalogue.json` row; and the copyright from the binary's own nameID 0.
  A face that cannot supply all of them is not added.
- **Each face passes the existing gates before it is proposed** — `font-catalogue.test.ts`'s nameID 13
  tie against its declared SPDX id, and its `scripts` declaration checked against its own `cmap` **in
  both directions** (a claimed script it cannot draw renders tofu; an unclaimed one it can draw staples
  a redundant fallback onto every chain that picks it).
- **The licence is one of D-8.5.3's four**, admitted by identifier, and an unclassifiable licence is a
  **build failure**, never a warning (D-8.5.2).
- **The batch's membership rule, size and OWNER are written down in this story.** D-16.5 left the batch
  unbounded and unowned and that is the defect D-16.R.2 exists to close. A batch nobody re-runs is a
  queue, and a queue was the thing the owner was told they were not funding.
- **From the family's own upstream, never from the `google/fonts` mirror.** The mirror is what lacks the
  statics; taking bytes from it is the failure this story routes around.
- **No derivation unless a family genuinely publishes no static**, and then it is `instance_faces.py`'s
  regime — committed output, replayable, hashes asserted in both directions — never a build-time or
  runtime instancing.
- Every recorded measurement carries **command, commit, tree state and working directory** (D-8.4j.8).
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **A face's licence cannot be classified against the four identifiers**, or its `LICENSE*` cannot be
  fetched from the same release as the binary.
- **A face's `cmap` disagrees with its declared `scripts`** in either direction.
- **The batch would be unbounded**, or would land without a named owner and a stated re-run trigger.
- **`maximumCacheAssets` (64) would be exceeded**, or the offline release contract would move. Measure
  the slot count before proposing a size; 44 of 64 were in use at Story 8.6.
- **Any of the 23 golden digests moves.** This story adds assets; it renders nothing differently.

**Never:** bytes from the `google/fonts` mirror · runtime or build-time instancing · a CJK family
(D-8.5.3's scope) · a variable face committed to the catalogue · a hand-copied licence text.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Family publishes a static Regular | `Roboto` at `googlefonts/roboto-classic` | Added: binary, `LICENSE*`, `NOTICE.md` with both digests, catalogue row | — |
| Family publishes VF only, everywhere | no static in any upstream release | **Out of the batch**, recorded with the evidence — not derived silently | Excluded, stated |
| Upstream ships several statics | full weight range published | **Regular only.** Bold and italic are out of scope by a standing Non-goal | — |
| Licence outside the four | e.g. a ShareAlike release | **Build failure**, never a warning | Fails the build |
| `cmap` disagrees with `scripts` | claims `thai`, cannot draw it | **Block** — the declaration is corrected or the face is dropped | Fails the build |
| Family already in the local tier | `Roboto`, `Inter` | No-op; not added twice | — |
| Slot budget | additions approach `maximumCacheAssets` | Measured before proposal; the batch is sized to fit | Halt rather than raise the cap |

</intent-contract>

## Code Map

**Every anchor below was re-measured at `efd79bf` this dispatch.** The previous draft's anchors were
taken at `4aa610a`; Story 16.1 landed 14 commits in between and moved several. Corrections are in the
Spec Change Log.

**The catalogue and its gates**
- `folio-designer/font-catalogue.json` — **21 rows**, every row exactly
  `{id, directory, file, family, licence, scripts}`. Untouched by 16.1. A new family is one new row.
- `folio-designer/public/fonts/<id>/` — **27 directories**, each holding exactly three files: one
  `.ttf`, one `LICENSE*`, one `NOTICE.md`. 21 are catalogue faces; the other 6 (`ibmplexmono`,
  `ibmplexsans`, `ibmplexsansthai`, `notosans`, `notosanssc`, `notosansthai`) are the **shipped**
  families — see the collision guard below. Licence filenames vary legally
  (`LICENSE-OFL.txt`, `LICENSE-OFL.md`, `LICENSE-UFL.txt`); exactly one `LICENSE*` is required.
- `folio-designer/public/fonts/roboto/NOTICE.md` — **the template; read it first.** Verified to carry
  all four fields D-16.R.12 names: `sha256 of the release archive` (with archive byte size),
  `sha256 of the SHIPPED file`, `| Size | 355,956 bytes |`, `| Fetched | 2026-09-02 |` — plus a
  distinct *source* digest, the licence file's own sha256, and the
  `> **NO DERIVATION APPLIES.**` clause that 24 of the 27 carry. Do not invent a shape; copy this one.
- `folio-designer/scripts/build-wasm.mjs:194-218` — the row validator every new row must satisfy:
  non-empty strings (`:198`); `scripts` a non-empty, duplicate-free subset of `scriptFallbacks`'
  closed vocabulary (`:200-204`); `id` matching `/^[a-z0-9]+$/` and unique (`:208-209`);
  `directory`/`file` single plain path segments (`:211`); `family` matching `familyShape` (`:213`);
  **family-collision check at `:214`**; `.ttf` extension (`:217`); then `fingerprint()` into
  `src/generated/runtime/` (`:218`). `:278` requires exactly one `LICENSE*` beside the binary.
  16.1 changed this file: it now delegates nameID 0 to `src/font-name-table.ts` and calls
  `emitFontIndexModule()` at `:356`.
- ⚠ **`build-wasm.mjs:153` / `:195` — `shippedFamilies` is seeded into `catalogueFamilies`, so a
  catalogue row naming `IBM Plex Sans`, `IBM Plex Mono`, `IBM Plex Sans Thai`, `Noto Sans`,
  `Noto Sans Thai` or `Noto Sans SC` throws at `:214`.** Mirrored as an assertion at
  `folio-designer/src/font-catalogue.test.ts:295`. Two top-50 candidates (`Noto Sans`, rank 12;
  `IBM Plex Sans`, rank 19) and one Thai candidate (`Noto Sans Thai`) are **structurally unaddable**
  by this guard. They are not a bug to route around — see Design Notes.
- `folio-designer/src/font-catalogue.test.ts` — the per-face gates. **nameID 13 tie now at `:344-356`**
  (the draft said `:355-366`; 16.1 shifted it −11 lines). `cmap` cross-check at `:434-467`, helper
  `cmapCoverage` at `:206`. Generated-face count tie at `:403`.

**The population floor — three sites, not one**
- `folio-designer/src/font-catalogue.test.ts:292` — `toBeGreaterThanOrEqual(20)`
- `folio-designer/src/font-index.test.ts:107` — `toBeGreaterThanOrEqual(20)`
- `folio-designer/src/font-name-table.test.ts:25` — `toBeGreaterThanOrEqual(20)`
  All three are `>= 20`, **not 21** as D-16.R.3 and D-16.R.12 both state. Two of the three were added
  by Story 16.1, after the previous draft was written. Nothing asserts an exact count, so a batch
  landing without raising them is silently unmeasured — exactly D-16.R.12's failure.

**The tier join — added by 16.1, absent from the previous draft, and load-bearing here**
- `folio-designer/src/font-index.ts:51-53` — `localByFamily` / `localTierHolds`, **exact `family`
  string equality**, no normalisation (D-16.R.3).
- `folio-designer/src/font-index.ts:80` — `addableFromTheWeb = (row) => !row.variable`; `:98` —
  `webFamilies = familyIndex.filter((row) => addableFromTheWeb(row) && !localTierHolds(row.family))`.
  **Adding a family locally is what makes it offered**; the web row was already filtered out for being
  variable, so the addable count rises by exactly the batch size.
- `folio-designer/src/font-index.ts:113-119` — `offeredFamilies`, local rows first.
  `folio-designer/src/App.tsx:681` — `if (source.tier === 'local')` reads the precached bundle asset
  and never reaches `fetchWebFamily`.
- `folio-designer/font-index.json` — the committed snapshot: `families` (1,811 rows), each with
  `family`, `category`, `subsets`, `axes`, `styles`, `popularity`. **`variable` is not stored** — it is
  derived at emit time as `axes.length > 0` (`scripts/build-font-index.mjs:124`). A survey reading
  `row.variable` off this JSON silently measures nothing.

**The slot ceiling**
- `folio-designer/src/release-payload.ts:33` — `const maximumCacheAssets = 64` (still line 33);
  `:32` `minimumCacheAssets = 10`. Read by regex at `scripts/offline-release-contract.mjs:66-70`.
- Enforced at `src/release-payload.ts:43` and `scripts/verify-offline-release.mjs:51` — **hard bound
  only. The margin is watched by nothing (DW-162).**
- Measured this dispatch from `dist/offline-release-manifest.json`: `s1.assetCount` **44**,
  and **all 21 `catalogue-*.ttf` are among the 44 cache assets**. One added family = one slot.

**The repo-wide asset licence gate — omitted by the previous draft entirely**
- `lint/internal/manifest/manifest.go:299` `ResolveAssets` walks the **whole repository** for font
  binaries; `:424` requires a `NOTICE*` file, `:429` requires a line starting with `Copyright`.
- `lint/internal/manifest/manifest_test.go:44` **`TestManifestUpToDate`** compares the committed
  `lint/MANIFEST.md` against that live walk. `lint/MANIFEST.md` carries **27 `public/fonts` rows**
  today. **Every added face reds `cd lint && go test ./...` until `go run ./cmd/genmanifest` is run
  and the regenerated `MANIFEST.md` is committed.**
- `manifest_test.go:904` `TestFontAssetLicenceAllowlistIsTheOwnersFourIds` — the four-id allowlist is
  enforced here, at the asset site. This is where the contract's "unclassifiable licence is a build
  failure" actually bites.

**Read-only / do not touch**
- `tools/fontgen/instance_faces.py:117-168` — `UPSTREAM` is still exactly **3 entries**
  (`notosans`, `notosansthai`, `notosanssc`), all writing to `folio-go/fonts/`. **No designer
  catalogue face is derived by it.** `folio-go/fontgen_matrix_test.go:117` hardcodes
  `"derived and compared 3 of 3 faces"`, so a fourth entry reds Go on the witness string. This story
  does not extend it.
- `folio-go/byte_neutrality_test.go:100` — `goldenDigestRecord`, **23 entries** (verified: 23
  `expected.pdf` files under `fixtures/`). The draft's "23 golden digests" is correct at HEAD.
- `_bmad-output/specs/spec-fonts/font-catalogue.md` — `## Selection criteria` (`:30`),
  `## The local face tier, and the tier beside it` (`:42`), `## Per-entry record` (`:73`).

## Tasks & Acceptance

**Execution:**

- **TASK 1 — the batch survey, offline, and it is a GATE.** Produce the proposal from the committed
  artifacts alone, with **no network access at all**: rank `folio-designer/font-index.json`'s
  `families` ascending by `popularity`, mark each row variable-only as `axes.length > 0`, subtract the
  families `font-catalogue.json` already holds, and subtract the six `shippedFamilies`. Record the
  method, the commit, the working directory and the tree state (D-8.4j.8). Write the result to
  `_bmad-output/implementation-artifacts/16-1a-batch-proposal.md` as a table with one row per
  candidate: family, popularity rank, subsets, and the **remaining slot arithmetic**
  (`44 + N <= 64`). Name the six Thai variable-only families explicitly and state each one's
  disposition. **Then HALT with status `blocked` and blocking condition `batch proposal awaiting
  gate`.** The per-family upstream-and-licence question — *does this family's own project release
  publish a static Regular, under which of the four identifiers?* — is the only part needing network,
  and it is not asked until the list is agreed. This is the contract's *"the batch would be
  unbounded"* Block If, discharged as a stop rather than as a judgement the build makes alone.

- **TASK 2 (post-gate, per admitted family) — `folio-designer/public/fonts/<id>/`** — fetch the static
  Regular and its `LICENSE*` from the **same** upstream release, from that family's own project and
  never the `google/fonts` mirror. Write `NOTICE.md` against `roboto/NOTICE.md`'s shape, carrying the
  release archive digest, the source digest, the committed digest, the byte size, the fetch date, the
  licence file's own digest, and the `NO DERIVATION APPLIES` clause. Confirm nameID 0 and nameID 13
  from the binary itself.

- **TASK 3 — `folio-designer/font-catalogue.json`** — one row per admitted family; `scripts` derived
  from the binary's own `cmap` rather than from the index's `subsets`, in both directions; `id`
  lower-case alphanumeric and unique; `family` byte-equal to the string the index snapshot uses, so
  `localTierHolds` suppresses the web row (`font-index.ts:98`).

- **TASK 4 — raise the population floor in all THREE sites, to the same new number, in one commit:**
  `src/font-catalogue.test.ts:292`, `src/font-index.test.ts:107`, `src/font-name-table.test.ts:25`.
  Add a comment at each naming the other two, so the next batch finds them. The floor is deliberate,
  not incidental: D-16.R.12 — *"a floor left at 21 while the tier grows to 30 is a floor that stops
  measuring the thing it was built to measure."* Note the sites read `>= 20` today, not 21.

- **TASK 5 — `lint/MANIFEST.md`** — regenerate with `cd lint && go run ./cmd/genmanifest` and commit
  it, so `TestManifestUpToDate` passes. Do not hand-edit the table.

- **TASK 6 — `folio-designer/src/font-index.test.ts`** — extend the offered-once assertion (`:141-150`,
  today `Roboto` and `Inter`) to cover added families, and add an assertion that the addable count
  rose by exactly the batch size. Red-prove it by removing one added row from the catalogue.

- **TASK 7 — `_bmad-output/specs/spec-fonts/font-catalogue.md`** — record the batch under
  `## The local face tier, and the tier beside it`: its membership rule, its measured size, its
  **named owner** and its **re-run trigger**.

- **TASK 8 — `_bmad-output/implementation-artifacts/deferred-work.md`** — register as **DW-166** the
  standing curation obligation (popularity moves; this batch is a snapshot of it), and record against
  **DW-162** that this story consumed N of the 20 remaining slots, with the new margin stated.

- **`source` stops being a URL, on BOTH tiers** (D-16.R.13, and this story carries it because it is
  already touching the field). Committed tier (`build-wasm.mjs`): **inline the pinned upstream release
  and the committed digest** instead of pointing at a `NOTICE.md` the recipient of a `.folio` does not
  have. Fetched tier (`src/font-source.ts`): drop the `.../main/...` branch URL; carry the **upstream
  project, the path within it, and the fetch date**. Never a resolvable-looking URL, never a branch
  name, and **never the SHA-256** — that is already the asset key, and duplicating it creates two
  authorities on one fact.
- **The tripwire, because the convention alone will not hold:** a test asserting `source` contains
  **no scheme and no host** for either tier.
- **Verify, do not assume, that no golden moves.** `source` is a `.folio` field; check whether any
  fixture feeding a golden embeds a face rather than reasoning it through.
- Record on **D-000.15's running list** for Story 15.3: format freedom spent on correcting `source`'s
  shape.

**Acceptance Criteria:**
- Given the committed index and catalogue, when the batch survey runs, then it completes with no
  network request, derives variable-only as `axes.length > 0`, excludes the six shipped families by
  name, states the slot arithmetic `44 + N <= 64`, and the run **halts for the gate** before any
  upstream fetch.
- Given the batch size N, when the release is built, then `s1.assetCount` equals `44 + N` and is at
  most 64, and the build never raises `maximumCacheAssets`.
- Given the three population-floor sites, when the batch lands, then all three declare the same raised
  floor and none is left at 20.
- Given a face added to `public/fonts/`, when `cd lint && go test ./...` runs, then
  `TestManifestUpToDate` passes because `lint/MANIFEST.md` was regenerated, and the face's licence
  classifies to one of the owner's four identifiers.
- Given an added family, when the font browser lists families, then it is offered **exactly once**,
  from the local tier, and the addable count rose by exactly the batch size.
- Given the 23 golden digests in `folio-go/byte_neutrality_test.go:100`, when this story lands, then
  none has moved.

## Spec Change Log

**2026-09-03 — re-plan at `efd79bf`. The intent contract block preserved byte-identically (md5
`8d6506096ad83b790779322b902b939b`, 4,978 bytes); every change below is outside it.**

1. **Anchors re-measured against 14 intervening commits.** The nameID 13 tie moved `:355-366` →
   `:344-356`. The Code Map now carries the `cmap` cross-check, the row validator and the collision
   guard by line.
2. **The population floor is three sites at `>= 20`, not one site at 21.** D-16.R.3 and D-16.R.12 both
   describe it as "the 21-face population floor"; measured, no site asserts 21 and two of the three
   sites did not exist when those rulings were written. The ruling's substance is unaffected — the
   floor must rise — but a build told to move it "in one place" would leave two behind.
3. **`font-index.ts` and the tier join added to the Code Map.** They are the mechanism by which this
   story's outcome occurs and were absent from the previous draft, which predates Story 16.1.
4. **The `shippedFamilies` collision guard recorded as a hard membership-rule exclusion**, removing
   `Noto Sans` and `IBM Plex Sans` from the top-50 candidate set and `Noto Sans Thai` from the Thai
   set. Not previously noted anywhere; it throws the build rather than warning.
5. **The `lint` module added to the Code Map and to Verification.** `TestManifestUpToDate` and the
   four-id asset allowlist are where this story's provenance regime is actually enforced, and the
   previous draft ran neither.
6. **The batch-proposal task converted from "get it agreed" to an explicit HALT.** In an unattended
   dispatch there is no party to agree with, so the previous phrasing was improvisable.
7. **The slot ceiling made a measured, binding input.** 44 of 64 in use, all 21 catalogue faces among
   them; headroom **20**. Recorded with the candidate count (32) so the constraint is visible as
   binding rather than theoretical.
8. **"37 of the top 50"** — the frozen contract's figure, from a live reading on 2026-09-02. Measured
   against the committed snapshot at HEAD it is **38**. Immaterial to the verdict, recorded so the
   difference is not later read as drift in the code.

## Review Triage Log

*Empty — no review pass has run. This dispatch halted after planning.*

## Design Notes

**The arithmetic, measured this dispatch at `efd79bf` (working directory `folio-designer`, tree
clean, offline, from committed artifacts plus the untracked build output).**

| step | count |
|---|---|
| top 50 by `popularity`, variable-only (`axes.length > 0`) | 38 |
| minus families the local tier already holds (Roboto, Inter, Noto Serif, Space Grotesk) | 34 |
| minus `shippedFamilies` collisions (`Noto Sans`, `IBM Plex Sans`) | **32 admissible** |
| free cache slots (`64 − 44`) | **20** |

**The batch is ceiling-bound, not taste-bound.** 32 candidates want 20 slots, and that gap is the
single most important fact for the gate. It also means the batch's membership rule must be a
*ranking* rule, not a *threshold* rule: "the most popular N that fit" rather than "everything above
popularity P". A threshold rule silently overflows the next time popularity moves.

**The slot ceiling is real, and D-16.R.2 underestimated it.** That decision reasoned that D-8.4d.1
moved catalogue faces to fetch-on-first-pick, so *"the binding constraints are committed repository
bytes and the per-face licence work, not `brotli.totalBytes`."* The payload half is correct. But all
21 `catalogue-*.ttf` assets **are** in `s1.cacheAssets`, so each face still consumes one of the 64
slots. The binding constraint is the slot count, and nothing in the build reports the margin
(DW-162). Raising `maximumCacheAssets` is refused by the contract's own Block If.

**`Noto Sans` and `IBM Plex Sans` are in the bundle and still not offerable, and that is not this
story's to fix.** Their bytes ship as chrome faces; the browser filters their index rows out for
being variable; and a catalogue row naming them throws at `build-wasm.mjs:214` because the collision
guard exists to stop two `@font-face` rules declaring one family. Making them pickable means giving
the shipped tier a browser presence, which is a different story. Register it, do not solve it here.

**The Thai set is three, not six.** The contract's task says *"the six Thai variable-only families are
in the candidate set by name"*, and there are exactly six — `Anuphan`, `Google Sans`, `Noto Sans
Thai`, `Noto Sans Thai Looped`, `Noto Serif Thai`, `Playpen Sans Thai`. But `Noto Sans Thai Looped`
and `Noto Serif Thai` are already local tier (a contract no-op), and `Noto Sans Thai` is a shipped
family. Three are genuine candidates, and only `Google Sans` (rank 3) is anywhere near the head.

**Why the survey must not read `row.variable`.** `font-index.json` stores `axes`; `variable` is
synthesised at emit time. A survey reading `row.variable` off the JSON gets `undefined` for every row,
reports zero variable-only families and passes without complaint. This was measured the wrong way once
during this plan gate before being caught, which is why it is written down.

**This is not a fontgen story.** D-16.5 asserted derivation was the remedy and D-16.R.2 was taken on
that framing; D-16.R.2a corrected it. Measured again at HEAD: `UPSTREAM` is three engine faces, no
designer catalogue face is derived, and a fourth entry would red
`folio-go/fontgen_matrix_test.go:117`'s hardcoded witness string. Derivation stays available only for
a family whose own upstream publishes no static.

**Why Regular only.** Bold and italic are out by a standing Non-goal (*"no synthetic bold or oblique,
and no variable-font axes. A weight is a face or it does not exist"*) and by Epic 11's placement.

## Verification

**Commands:**
- `cd folio-designer && npm run scan:font-hosts` — expected: exit 0. This story commits bytes rather
  than adding a host, so no new declaration should be needed; a new host line is a red flag.
- `cd folio-designer && npm run lint` — expected: clean.
- `cd folio-designer && npm run test` — expected: the suite green, including the three raised
  population floors, the `cmap` cross-check and the nameID 13 tie over the enlarged population.
- `cd folio-designer && npm run build` — expected: exit 0 through `build:wasm`, `tsc -b`, `vite
  build`, `build:offline` and `verify:offline`. Report the resulting `s1.assetCount` explicitly.
- `cd folio-designer && npm run test:e2e:compile` — expected: exit 0. Compile-only; **no browser run
  in this story** (D-16.R.1 cadence).
- `cd lint && go test ./...` — expected: green, **including `TestManifestUpToDate`**. This is the gate
  on the per-face provenance regime and the four-id licence allowlist.
- `cd folio-go && go test ./...` — expected: green except the two mandated `internal/text` P6g floor
  reds, which are pre-existing and are not this story's.

**Manual checks:**
- `shasum -a 256 fixtures/*/expected.pdf` — expected: 23 lines, every digest matching
  `folio-go/byte_neutrality_test.go:100`'s `goldenDigestRecord`. This story adds assets; it renders
  nothing differently.
- `folio-designer/src/release-payload.ts:33` still reads `const maximumCacheAssets = 64` — unchanged.
  State the measured `s1.assetCount` and the remaining margin in the Delivery Log.

**Known baseline reds at `efd79bf`, not attributable to this story:**
- `folio-designer`: `canvas-authority-contract.test.ts:190` (DW-152, Epic 9/10 lane).
- `folio-go`: two `internal/text` failures — the mandated P6g exercise-floor red.

## Auto Run Result

Status: blocked
Blocking condition: batch proposal awaiting gate
Directive: TASK 1 complete; halted at the designed gate before any upstream fetch
Baseline: `efd79bfc41cfb9ed45dd4a6223da38e83c00797b`; run at `1632bbb` (`1632bbbb624050b4b115e614944ccddf2e5ed221`)
Gate artifact: `_bmad-output/implementation-artifacts/16-1a-batch-proposal.md`
Intent contract: unchanged, byte-identical (md5 `8d6506096ad83b790779322b902b939b`, 4,978 bytes)

**What ran.** TASK 1 only, offline, from committed artifacts plus the untracked build output. Every
figure in Design Notes reproduced exactly: 38 variable-only in the top 50, 34 after the local tier,
**32 admissible**, `s1.assetCount` **44** with all 21 `catalogue-*.ttf` among the cache assets, so
`maximumCacheAssets` (64) leaves **20 free slots** and `N ≤ 20`.

**What did not run, and why.** TASKS 2–8 and the `source`-shape work (D-16.R.13) are all downstream of
the agreed membership list — TASK 2 needs the network the gate exists to withhold, and TASKS 4, 6 and 8
need N. Nothing was fetched, no font byte was written, no test floor was moved, and no source file was
touched. The tree carries exactly two changed paths: this spec and the new gate artifact.

**One correction the survey found, recorded so it is not read later as drift.** The snapshot's
`popularity` field is **not a dense rank** — 1,811 families carry 1,100 distinct values. The Code Map's
"`Noto Sans`, rank 12" and "`IBM Plex Sans`, rank 19" are **positions in the ascending sort**; the field
values are 8 and 15. Both numbers are carried in the gate artifact so neither reading can be mistaken
for the other. The ties matter: `Instrument Sans` and `Rubik` both carry popularity 22 and sit either
side of the `N = 20` cut, which is why the proposed membership rule names an explicit tie-break.

**What the gate must settle before this story can resume.** N (at most 20; filling the ceiling exactly
leaves zero margin against a bound nothing reports — DW-162), the membership rule as worded, and the
batch's **named owner** and **re-run trigger**, which the build cannot appoint.
