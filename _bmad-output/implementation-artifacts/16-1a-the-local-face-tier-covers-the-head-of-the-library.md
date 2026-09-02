---
title: 'Story 16.1a: The local face tier covers the head of the library'
type: 'feature'
created: '2026-09-03'
status: 'done'
review_loop_iteration: 0
followup_review_recommended: true
baseline_commit: 'efd79bfc41cfb9ed45dd4a6223da38e83c00797b'
baseline_revision: '38005cdd89cb6adbbee0c089121810b3c401a331'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
warnings: ['oversized']
deferred:
  - summary: >-
      The catalogue's script vocabulary is closed at latin/thai/cjk, so faces covering Greek or
      Cyrillic cannot declare it and the cmap cross-check probes neither script.
    evidence: |-
      scriptFallbacks at build-wasm.mjs:177 is exactly {latin, thai, cjk} and :202 throws on any
      other value, so ["latin"] is the only legal declaration for Arimo (Liberation Sans metric
      compatible) and Lora (sourced from cyrealtype/Lora-Cyrillic). The unclaimed-direction half of
      the cmap check therefore probes only thai and cjk, and Greek/Cyrillic coverage is neither
      declared nor refuted.
    location: >-
      folio-designer/scripts/build-wasm.mjs:177
    severity: medium
  - summary: >-
      The source tripwire's host check recognises a host by an allowlist of nine TLDs, so a ccTLD
      host such as rsms.me would pass.
    evidence: |-
      hostShaped at font-provenance.test.ts:47 lists com|org|net|io|dev|co|app|sh|xyz. The narrowness
      is deliberate and documented (a broad TLD match reads .ttf and .md as hosts), so widening it
      naively regresses the guard; the gap is real but needs a positive shape assertion rather than a
      longer blocklist.
    location: >-
      folio-designer/src/font-provenance.test.ts:47
    severity: low
  - summary: >-
      The "exactly one writer per tier" check reads two named files, so a third module assigning
      source anywhere else is invisible to it.
    evidence: |-
      font-provenance.test.ts:110 scrapes font-source.ts and build-wasm.mjs by text. A new module in
      src/ writing its own source string satisfies the stated property's wording while defeating its
      purpose, and a trailing-comma or line-wrap reformat breaks the regex.
    location: >-
      folio-designer/src/font-provenance.test.ts:110
    severity: medium
  - summary: >-
      committedFaceSource is defined inline in build-wasm.mjs and is not importable, so its
      documented throw path has no test.
    evidence: |-
      The helper parses all 31 NOTICE.md files and its comment states that an unparseable NOTICE
      throws rather than emitting a vaguer string, but nothing red-proves that. Sibling script
      helpers were extracted for exactly this reason (scripts/offline-release-contract.mjs has
      scripts/offline-release-contract.test.mjs, and vitest already collects scripts/**/*.test.mjs).
      A whitespace edit to any NOTICE table breaks the build with no test naming why.
    location: >-
      folio-designer/scripts/build-wasm.mjs:312
    severity: medium
  - summary: >-
      Existing .folio documents keep the old source shape forever; no reader changed, no migration
      exists, and the tripwire only ever sees freshly generated values.
    evidence: |-
      The story changes what is written while SupportedMajor stays 2, so a document saved before this
      change retains a repository path or a .../main/... branch URL. Nothing asserts an old-shape
      source still loads, and DW-167 records the format spend without noting the installed base.
    location: >-
      folio-designer/src/font-source.ts
    severity: medium
  - summary: >-
      DW-168's preferred discharge (-count=1 wherever a story's verification is recorded) was applied
      by hand this run and is written into no script, Makefile target or CI config.
    evidence: |-
      The stale-cache class that produced a false green for TestLicenceSignalCensus in this very
      dispatch is therefore fully live for the next story. DW-168's owner field reads "whoever next
      adds a test that reads the tree at runtime", which is not an assignable owner.
    location: >-
      _bmad-output/implementation-artifacts/deferred-work.md
    severity: medium
  - summary: >-
      DW-166's first re-run trigger relies entirely on a human noticing that the index snapshot was
      regenerated.
    evidence: |-
      A cheap tripwire exists and was not added - assert that the committed font-index.json's last
      commit still equals the pinned d6d51f16988cddf20d1a28697cd556b3d0a63f62, or that the ten batch
      families are still the top-20 refused set. Conspicuous in a story that added a tripwire for
      source.
    location: >-
      _bmad-output/implementation-artifacts/deferred-work.md
    severity: medium
  - summary: >-
      The release build still reports no cache-slot margin, and this story halved it from 20 to 10.
    evidence: |-
      DW-162's stated discharge is the release build printing the margin and failing, or at minimum
      warning, at a declared threshold. The hard bound at release-payload.ts:43 and
      verify-offline-release.mjs:51 still fails only on exceeding 64, so the first signal remains a
      release that fails outright - now half as far away.
    location: >-
      folio-designer/src/release-payload.ts:33
    severity: medium
  - summary: >-
      licenceSignatures now holds three patterns matched against licence text with no assertion that
      a text matches exactly one.
    evidence: |-
      Adding Apache-2.0 alongside the OFL and Ubuntu patterns makes overlap newly possible; a
      dual-licensed or quoting text could satisfy two signatures, weakening the closed table's
      guarantee that admission is a decision somebody makes rather than a first match.
    location: >-
      folio-designer/src/font-catalogue.test.ts:197
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the intent contract below governs implementation.*

The font browser could offer an author almost two thousand typeface families, but not the ones they are
most likely to want. Roboto, Open Sans, Montserrat, Lora — most of the fifty most popular — are published
on Google's mirror in a form this product deliberately does not accept, because it holds every weight in
one file and picking a weight out of it at the last moment would make the same template print differently
on different machines.

There was a cheaper answer than it first appeared. Those families publish perfectly ordinary single-weight
files from their **own** project releases; only the mirror lacks them. This product already carried
twenty-one typefaces exactly that way, each with its licence, its copyright and a record of the precise
file it came from. **This story added ten more the same way**, so the browser offers them from the machine
rather than refusing them: **Arimo, DM Sans, Lora, Montserrat, Open Sans, Oswald, Plus Jakarta Sans,
Roboto Condensed, Roboto Mono and Roboto Slab.** The tier is now thirty-one families.

**The batch was sized by goal, not by shelf space, and that turned out to matter.** An earlier plan asked
which of thirty-two candidates would fit in twenty free slots — a question that would have filled every
slot the release had. The rule that governs instead names a target: the refused families in the top twenty
by popularity, minus the ones already carried, minus the ones the product's own interface fonts collide
with, minus any whose own project publishes nothing usable. That yields ten, into twenty slots, and leaves
ten spare rather than none.

**Two families in that top twenty could not be taken, and both are recorded rather than quietly dropped.**
Google Sans publishes no static file anywhere. Jost publishes one, but the file calls itself `Jost*` —
with the asterisk — and this product refuses to publish a family name that the typeface's own bytes
contradict. Neither was replaced by the next family down: a goal-shaped batch shrinks when a member fails,
it does not backfill.

The story also fixed a smaller thing it was already touching. Every embedded typeface records where it
came from, and the two halves of the product wrote that differently: one pointed at a file inside this
repository that the recipient of a document does not have, the other wrote a web address on a branch that
can move tomorrow. Both now say the same three things — the project, the file within it, and the date it
was fetched — and a test fails the build if either ever writes an address again.

What it deliberately did not do: it converted nothing, it added no general mechanism for getting statics,
and it does not promise the whole library. It is a curated batch, and the rule, the size, the owner and
the trigger for re-running it are written down — because a batch nobody owns is the failure this work
exists to prevent.

Done looks like: the families an author is most likely to type are offered and embed with no download at
all, each carrying its own terms.

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
  artifacts alone, with **no network access at all**, **reproducible from `folio-designer/font-index.json`
  at commit `d6d51f16988cddf20d1a28697cd556b3d0a63f62`** — the file name alone does not pin a set, because a regeneration
  between the survey and the build would change the membership while every recorded step still read
  true *(clause tightened 2026-09-03 at implementation, after the sha was re-measured)*: rank
  `folio-designer/font-index.json`'s
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

- **THE BATCH, fixed and reproducible from the COMMITTED snapshot** (`folio-designer/font-index.json`
  at commit **`d6d51f16988cddf20d1a28697cd556b3d0a63f62`**, 1,811 families, `snapshotDate` 2026-09-03) — **not** from the live index, which is a different
  ordering and produced a wrong set once already (D-16.R.19). Top-20 positions by `popularity`
  (name-ascending tie-break), refused = `axes` non-empty, minus already-local, minus `shippedFamilies`,
  minus unobtainable: **Open Sans · Roboto Mono · DM Sans · Montserrat · Arimo · Roboto Slab · Lora ·
  Roboto Condensed · Oswald · Plus Jakarta Sans · Jost.** Eleven, into twenty free slots.
  **Reproduce it offline; if your reproduction differs, halt.**
- **Measure and record this story's first-load payload delta, and hand it to Story 15.2's budget gate**
  (D-16.R.21). Story 15.0 is `backlog`, so D-8.4d.1's fetch-on-first-pick is **policy with no
  implementation** and these 11 faces are **precached today**. The story that spends records what it
  spent; without it 15.2 meets a budget that moved for reasons its own record does not contain.

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

9. **The golden-digest manual check has no working directory, and every command beside it does.**
   Verification's six commands each begin `cd folio-designer`, `cd lint` or `cd folio-go`, but the
   manual check reads `shasum -a 256 fixtures/*/expected.pdf`. The 23 goldens live at the **repository
   root** `fixtures/`, not `folio-go/fixtures/`. Run from `folio-go` — the natural reading, since the
   digest record it ties to is `folio-go/byte_neutrality_test.go:100` — the glob matches nothing and
   the check reports zero lines while exiting cleanly, which reads as "no golden moved". Run it from
   the repository root, and confirm it prints **23** lines before comparing any digest.

**2026-09-03 — implementation dispatch. The intent contract is again byte-identical (md5
`8d6506096ad83b790779322b902b939b`, 4,978 bytes); every change below is outside it.**

10. **THE BATCH IS TEN, NOT ELEVEN, and the Tasks section's *"Eleven, into twenty free slots"* is now a
    statement of the set as it stood at plan time.** `Jost` — **position 20, the last family inside the
    top-20 cut** — was dropped at TASK 2 on evidence:
    `indestructible-type/Jost` publishes a static Regular whose own `name` table calls the family
    `Jost*`, which fails `build-wasm.mjs`'s `familyShape` and cannot be declared as `Jost` without
    publishing a name the bytes contradict. The other ten are exactly as listed. **No backfill from the
    reserve**, per D-16.R.16: a goal-bounded set shrinks when a member fails.
11. **The population floor moves 20 → 31, not 20 → 32.** Same arithmetic, one fewer family.
12. **`Apache-2.0` was admitted to `font-catalogue.test.ts`'s `licenceSignatures`**, which the plan did
    not anticipate — `Roboto Slab`'s upstream is Apache-licensed, and that table is closed by design so
    an unadmitted id fails rather than skipping. Admission itself is unchanged; `Apache-2.0` was already
    on `lint`'s four-id allowlist.
13. **Five of the ten upstream projects publish no tagged release at all**, so their pin is a commit and
    their NOTICE says so explicitly, with a paragraph stating that the generated-archive digest is a
    fetch-time measurement while the file digests bind. The plan assumed a release archive per family.
14. **The `source` task bullet contradicts itself, and the contradiction is recorded rather than
    resolved by preference.** It asks the committed tier to inline *"the pinned upstream release and the
    committed digest"* and, four lines later, *"never the SHA-256 — that is already the asset key"*. For
    a committed face those are one value. The prohibition governs; see the Delivery Log and **DW-167**.

## Review Triage Log

**2026-09-03 — review pass over the implemented story. Twelve findings, all triaged `patch`; no
`intent_gap` and no `bad_spec`, so the code stood and these landed on top. Two further findings were
REJECTED on inspection and are recorded here so they are not re-litigated.**

*(The line this section carried until now — "Empty — no review pass has run. This dispatch halted after
planning." — was true of the planning dispatch and is preserved in that sentence rather than deleted.)*

**The two medium findings that were real defects rather than tidying.**

- **The fetched tier's REAL write path had no provenance assertion, and a mutation proved it.**
  `font-provenance.test.ts` called `webFaceSource` with an **explicit** date, so nothing ever observed
  `fetchWebFamily`'s default `today` expression — the only place a fetched pick's date actually comes
  from. Deleting the `.slice(0, 10)` from that default makes every fetched pick publish
  `…, fetched 2026-09-03T…Z`, and the suite stayed **green**. A new case in `font-source.test.ts` now
  drives `fetchWebFamily` with **no `today` argument** against the kanit stub and applies the shared
  predicate to the outcome. **Red-proved with that exact mutation**, then restored from a `cp` snapshot.
- **The census gap's twin, one level up: I added a FOURTH population floor while the other three
  comments still said "ONE OF THREE".** D-16.R.18's own rule — *a floor that exists in N files is N
  floors* — broken in the act of citing it. All four now say four and enumerate the other three.

**The rest, in one line each.** The `source` predicate moved to `src/test/provenance-shape.ts`, shared
by both suites rather than copied, because the defect it guards *is* two copies of a rule drifting ·
its project-half check asserted only non-emptiness, which `split(' — ')` makes nearly vacuous when the
separator is absent — the separator is now required and the half is shaped · `branchShaped` knew only
`main|master|HEAD` with no `i` flag, so `develop`, `trunk` and `Main` passed; widened, because a guard
fitted to the instance rather than the class lets the identical defect back in under another spelling ·
`licenceSignatures`' `Apache-2.0` pattern did not pin the **version**, so an Apache 1.1 name table would
have satisfied the admitted id — tightened to `folio-go`'s own `Apache License,?\s+Version 2\.0`, and
`RobotoSlab-Regular.ttf`'s nameID 13 re-checked against it · two stale slot counts in `build-wasm.mjs`
and one stale "THE 21 COMMITTED FACES" comment corrected · the Delivery Log's blanket *"exactly one
unmodified `LICENSE*` from the same pinned artifact"* now names the **two** faces for which the text is
a second asset of the same release rather than a file inside the archive · three different names for one
heading reconciled · `Jost`'s upstream stated **identically** in all three records, since it is the sole
evidence for a member's removal · the exclusion list now carries **both** units, position and
popularity, on every entry · and the *"minus CJK"* clause is recorded as having fired **zero** times,
measured.

**REJECTED, with the reason, so they are not raised again.**

- *"The delta test reads `row.variable`, which the story forbids."* The prohibition is on the raw
  `font-index.json` **snapshot**, which stores `axes` and no `variable` key. `familyIndex` rows are the
  emitted module's rows and genuinely carry `variable` — asserted at `font-index.test.ts`, and
  production's own `addableFromTheWeb` is literally `!row.variable`. Reading it there is correct.
- *"`Arimo` and `Lora` cover Cyrillic and Greek, so `scripts: ["latin"]` is under-inclusive."*
  `scriptFallbacks` is a **closed** `{latin, thai, cjk}` vocabulary and `build-wasm.mjs` throws on
  anything outside it, so `["latin"]` is the only legal declaration these faces can carry. The
  vocabulary's coarseness is a real observation and was deferred as such; the rows are correct as
  written.

### 2026-09-03 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 12: (high 0, medium 4, low 8)
- defer: 9: (high 0, medium 7, low 2)
- reject: 6: (high 0, medium 0, low 6)
- addressed_findings:
  - `[medium]` `[patch]` The fetched tier's real write path had no provenance assertion — `fetchWebFamily`'s default `today` was never observed; mutation-proven green. A case now drives `fetchWebFamily` with no `today` and asserts the shared shape predicate over the produced `source`; red-proved with that mutation.
  - `[medium]` `[patch]` A fourth population floor was added at `font-provenance.test.ts` while the other three comments still said "ONE OF THREE" — D-16.R.18's own rule broken in the act of citing it. All four now say four and enumerate the other three.
  - `[medium]` `[patch]` `licenceSignatures`' `Apache-2.0` pattern did not pin the version, so an Apache 1.1 name table satisfied the admitted id. Tightened to `folio-go`'s `Apache License,?\s+Version 2\.0` and re-checked against `RobotoSlab-Regular.ttf`'s nameID 13.
  - `[medium]` `[patch]` The Delivery Log's blanket "exactly one unmodified `LICENSE*` from the same pinned artifact" was contradicted by two NOTICEs; `Lora` and `Roboto Condensed` are now named as the faces whose licence text is a second asset of the same release.
  - `[low]` `[patch]` The project-half assertion was near-vacuous — `split(' — ')` returns the whole string when the separator is absent. The separator is now required and the half is shaped.
  - `[low]` `[patch]` `branchShaped` knew only `main|master|HEAD` with no `i` flag; widened so `develop`, `trunk` and `Main` cannot pass.
  - `[low]` `[patch]` Two stale slot counts in `build-wasm.mjs` (":272" and ":340") still said the release cache stays at 44; it is 54.
  - `[low]` `[patch]` `font-index.test.ts` still said "THE 21 COMMITTED FACES SURVIVE EPIC 16 UNCHANGED" above an assertion requiring 31.
  - `[low]` `[patch]` Three different names were given for one heading across the Delivery Log, DW-166 and `font-catalogue.md`; reconciled to the real heading.
  - `[low]` `[patch]` `Jost`'s upstream was stated two ways; now stated identically in all three records, since it is the sole evidence for a member's removal.
  - `[low]` `[patch]` The exclusion list mixed `popularity` values with sort positions in one sentence; every entry now carries both units, labelled.
  - `[low]` `[patch]` The "minus CJK" clause was unmarked despite excluding nobody; recorded as firing zero times, measured (`publishedFamilies` 1,946 − `excludedCjkFamilies` 135 = 1,811; no CJK family present, and no row declares a Chinese, Japanese or Korean subset).
  - `[low]` `[reject]` "The delta test uses `row.variable`, which the story forbids" — the prohibition is on the raw `font-index.json` snapshot; `familyIndex` rows genuinely carry `variable` (asserted at `font-index.test.ts:117`) and production `addableFromTheWeb` is literally `!row.variable`.
  - `[low]` `[reject]` "`scripts: [\"latin\"]` is under-inclusive for Arimo and Lora" — `scriptFallbacks` is a closed `{latin, thai, cjk}` and `build-wasm.mjs:202` throws on anything else, so `["latin"]` is the only legal declaration. Deferred as a vocabulary limitation instead.
  - `[low]` `[reject]` "37 of the top 50 versus the measured 38" — already recorded as Spec Change Log entry 8.
  - `[low]` `[reject]` Line-wrap drift and a stray blank line after a `---`; cosmetic.
  - `[low]` `[reject]` The gate artifact's `status: settled` is not in the spec-status vocabulary; it is a gate artifact, not a spec.
  - `[low]` `[reject]` The "PINNED TO THE SNAPSHOT COMMIT" paragraph appears in three documents; the duplication is deliberate so each record stands alone.

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

Status: done
Blocking condition: none
Directive: TASKS 1–8 and the `source`-shape work complete; the gate the previous dispatch halted at was
settled by D-16.R.16 (criterion) and D-16.R.19 (membership) before this dispatch began
Baseline: `efd79bfc41cfb9ed45dd4a6223da38e83c00797b`; run at `38005cdd89cb6adbbee0c089121810b3c401a331`
Gate artifact: `_bmad-output/implementation-artifacts/16-1a-batch-proposal.md` — status `settled`
Intent contract: unchanged, byte-identical (md5 `8d6506096ad83b790779322b902b939b`, 4,978 bytes)

**Review findings breakdown.** 12 patched (0 high, 4 medium, 8 low), 9 deferred (frontmatter
`deferred:`), 6 rejected. Follow-up review **recommended**: `3 x 4 medium + 1 x 8 low = 20`, at or
above the threshold of 5. No `intent_gap` and no `bad_spec`, so no loopback ran and
`review_loop_iteration` stays 0.

**Verification measured by the driving workflow at `84507cf`, not taken from the implementation
subagent's report.** Every Go command was re-run with `-count=1` after a cached `ok` was found to be
hiding a real failure earlier in this dispatch.

- `npm run test` — **515 passed, 1 failed** of 516; 47 of 48 files. The single failure is
  `canvas-authority-contract.test.ts:190` (DW-152). Both files behind it — `e2e/e9-5-border-no-ink.spec.ts`
  and the test itself — are untouched since `baseline_commit` `efd79bf`, verified by
  `git diff --name-only`.
- `npm run build` — exit 0 through `verify:offline`. **`s1.assetCount` 54** (44 + 10), **31**
  `catalogue-*.ttf` among the cache assets, **margin 10** against `maximumCacheAssets` 64, which
  `src/release-payload.ts:33` still declares unchanged.
- `npm run lint` — exit 0 (4 pre-existing `react(only-export-components)` warnings).
- `npm run scan:font-hosts` — exit 0, **0 occurrences in 600 tracked source files** (floor 400).
- `npm run test:e2e:compile` — exit 0.
- `cd lint && go test -count=1 ./...` — exit 0, all 4 packages `ok`, including `TestManifestUpToDate`
  and `TestLicenceSignalCensus`.
- `cd folio-go && go test -count=1 ./...` — **13 packages `ok`, 1 failed**:
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`, "floor not met: got 7, need >=20". No
  `folio-go/` file changed since `efd79bf`, verified by `git diff --name-only`.
- `shasum -a 256 fixtures/*/expected.pdf` **from the repository root** — **23 lines**, 23 pinned in
  `goldenDigestRecord`, **0 mismatches, 0 unpinned**. Run from `folio-go` it errors on the glob rather
  than printing zero lines, which is the trap Spec Change Log entry 9 exists for; it was hit once
  during this dispatch and re-run from the root.

**Matrix Test Audit.** 5 of 7 rows COVERED by tests that ran and passed, with no vacuous or uncollected
covering test found. Rows 2 (*"publishes VF only, everywhere"* -> out of the batch, recorded with the
evidence) and 7 (*"the batch is sized to fit"* / halt rather than raise the cap) are **process
obligations rather than code behaviours** and have no covering test. Both have partial mechanical
backing: `familyShape` at `build-wasm.mjs:162` is what actually threw on `Jost*` and forced row 2's
drop, and `scripts/verify-offline-release.mjs:51` fails the build above 64 and ran green at 54. Row 7's
absence is DW-162 restated, and is deferred above.

**Residual risks.**
- The batch is **ten, not the eleven the Tasks section names**: `Jost` was admitted by the rule at
  position 20 and dropped at TASK 2 because its own binary's nameID 1 is `Jost*`, which
  `familyShape` refuses. Per D-16.R.16 a goal-bounded set shrinks rather than backfilling, so no
  reserve family was promoted.
- The **`source` workstream reaches further than the intent contract does.** The contract says nothing
  about `source`, the `.folio` wire field, or the fetched tier; that work is sourced from D-16.R.13.
  It changes the `source` string of the 21 pre-existing committed faces as well as the 10 new ones.
  The 23 goldens were re-measured rather than reasoned about, and none moved.
- **D-16.R.13 contradicts itself** on the committed tier: it asks for the committed digest inline and
  forbids restating the SHA-256, which for a committed face is one value. The prohibition was taken as
  governing and the conflict is registered as DW-167.
- The batch's **owner was appointed by the story** ("the engineering lead", flagged overridable), which
  the gate artifact had said the build could not do.
- **Five of the ten upstreams publish no tagged release**, so their pin is a commit; their NOTICEs say
  so and record that the generated-archive digest is a fetch-time measurement while the file digests
  bind.

## Delivery Log

**2026-09-03 — the batch landed at TEN, not eleven, and the one drop is the interesting part.**

**TASK 1 — the survey, reproduced offline before anything was fetched.** From
`folio-designer/font-index.json` **at commit `d6d51f16988cddf20d1a28697cd556b3d0a63f62`** (`snapshotDate` 2026-09-03,
1,811 families), sorted by `popularity`
ascending with family name ascending as the tie-break, variable-only derived as `axes.length > 0` and
never as `row.variable`, with `shippedFamilies` parsed out of `scripts/build-wasm.mjs` rather than
retyped. The top-20 candidate set came back **identical to D-16.R.19's twelve**, in the same order:
`Open Sans · Google Sans · Roboto Mono · DM Sans · Montserrat · Arimo · Roboto Slab · Lora ·
Roboto Condensed · Oswald · Plus Jakarta Sans · Jost`. No network was used for this step.

**PINNED TO THE SNAPSHOT COMMIT, not merely to the file** — `folio-designer/font-index.json` at
**`d6d51f16988cddf20d1a28697cd556b3d0a63f62`**. The file name alone does not pin a set: the index is regenerated by
`scripts/build-font-index.mjs`, and a regeneration landing between the survey and the build would
silently change the membership while every recorded step still read true. Worse, a later reader could
not tell whether the list was **wrong** or the **file moved underneath it** — two very different
failures with one appearance. Re-measured at implementation time rather than taken on report:
`git log -1 -- folio-designer/font-index.json` is exactly that commit, and this story's whole diff
against its baseline touches the file **not at all**.

**THE "MINUS CJK" CLAUSE FIRED ZERO TIMES, and that is recorded rather than left to be inferred.** A
reader who sees no CJK family in the batch cannot otherwise tell whether one was dropped by that clause
or whether none was ever a candidate — the ambiguity D-16.R.16's *"exclusions itemised"* guardrail
closes. **Measured on the committed snapshot: it removed nobody, because there is no CJK family in the
snapshot at all.** The index build excludes them upstream — `publishedFamilies` **1,946**,
`excludedCjkFamilies` **135**, and `1,946 − 135 = 1,811`, the family count exactly. Probed directly:
`Noto Sans JP`, `Noto Sans SC`, `Noto Sans TC`, `Noto Serif SC` and `Noto Sans KR` are **absent**, and
**zero** of the 1,811 rows declares a Chinese, Japanese or Korean subset. D-16.R.19 already caught this
as an exclusion with no referent; it is now measured and stated at the batch itself rather than only in
the correction that noticed it.

**The boundary of the cut, itemised.** Positions are the committed snapshot's own order (`popularity`
ascending, family name ascending as the tie-break), computed from `folio-designer/font-index.json` at
`d6d51f1` and from nothing live: **Jost 20** (pop 16) is the last family *inside* the cut; immediately outside sit
**Raleway 21** (17), **Share Tech 22** (17, static upstream and never refused), **Bebas Neue 23** (18,
static upstream), **Nunito 24** (18), **Fraunces 25** (19) and **Heebo 26** (19). `Raleway`, `Nunito`
and `Fraunces` are named because they are the three D-16.R.19 removed when it recomputed from the
committed snapshot — without this line they would appear in neither the batch nor the exclusions, which
is precisely what D-16.R.16's *"exclusions itemised"* guardrail forbids. They are out for sitting
**outside the top-20 cut**, not on merit and not on the ceiling.

**`Jost` (position 20) dropped for cause at TASK 2 — the batch shrank rather than backfilled.** It is
inside the cut and admissible on every other stated criterion. `indestructible-type/Jost`
publishes `fonts/ttf/Jost-400-Book.ttf` at release `3.5` and at HEAD; both call the family **`Jost*`** in
nameID 1. The asterisk is outside `build-wasm.mjs`'s `familyShape` — that string is interpolated
unescaped into `font-family: '<name>'` — and `src/font-catalogue.test.ts` ties the manifest's `family` to
nameID 1 byte-for-byte, so declaring it as `Jost` (the only spelling that would join the index) would
publish a family name the face's own bytes contradict. **No admissible static exists, so it is out by
D-16.R.16's own clause**, and the next family down was *not* promoted: a goal-bounded set shrinks when a
member fails, and backfilling would have re-introduced the margin-bounded criterion the lead removed.
`Google Sans` had already dropped on the same clause at D-16.R.16.

**TASK 2 — ten faces, each with the full committed regime.** Every binary is byte-for-byte its own
project's upstream static; **not one byte came from the `google/fonts` mirror**. Every directory carries
the binary, exactly one unmodified `LICENSE*` — **from the same pinned release, though for two of the
ten not from the same FILE, and both say so in their own NOTICE**: `Lora`'s zip carries no licence file,
so its `OFL.txt` is a **second asset of the same release `v3.021`**, and `roboto-3-classic`'s archive
carries none either, so `Roboto Condensed`'s is **`OFL.txt` at tag `v3.016`** — the same exception, and
the same wording, `roboto/NOTICE.md` has recorded since Story 8.5. Neither text is hand-copied and
neither comes from a different release. The other eight are inside the pinned artifact itself. Plus a
`NOTICE.md` written
against `roboto/NOTICE.md`'s shape — upstream project, pinned release, download URL, path inside the
archive, fetch date, archive digest with byte size, source digest, shipped digest, size, declared family,
`NO DERIVATION APPLIES`, and the licence file's own digest.

| family | id | upstream, pinned | licence | shipped bytes |
|---|---|---|---|---|
| Open Sans | `opensans` | `googlefonts/opensans` @ `bd7e376…` | OFL-1.1 | 147,528 |
| Roboto Mono | `robotomono` | `googlefonts/RobotoMono` @ `v3.001` | OFL-1.1 | 125,748 |
| DM Sans | `dmsans` | `googlefonts/dm-fonts` @ `4412393…` | OFL-1.1 | 78,256 |
| Montserrat | `montserrat` | `JulietaUla/Montserrat` @ `v9.000` | OFL-1.1 | 445,928 |
| Arimo | `arimo` | `googlefonts/Arimo` @ `4a6255f…` | OFL-1.1 | 478,712 |
| Roboto Slab | `robotoslab` | `googlefonts/robotoslab` @ `67af3ce…` | **Apache-2.0** | 171,376 |
| Lora | `lora` | `cyrealtype/Lora-Cyrillic` @ `v3.021` | OFL-1.1 | 198,380 |
| Roboto Condensed | `robotocondensed` | `googlefonts/roboto-3-classic` @ `v3.016` | OFL-1.1 | 355,076 |
| Oswald | `oswald` | `googlefonts/OswaldFont` @ `8979526…` | OFL-1.1 | 105,572 |
| Plus Jakarta Sans | `plusjakartasans` | `tokotype/PlusJakartaSans` @ `2.7.1` | OFL-1.1 | 132,040 |

**Five of the ten pin a COMMIT rather than a tag, and the NOTICE says so rather than implying a release
that does not exist.** `googlefonts/opensans`, `googlefonts/Arimo`, `googlefonts/robotoslab`,
`googlefonts/OswaldFont` and `googlefonts/dm-fonts` publish **no tagged release at all** (dm-fonts' own
releases are DeepMind faces, a different family). Each of those five, and the two whose tags carry no
attached asset (`RobotoMono` v3.001, `Montserrat` v9.000), is pinned to GitHub's source archive for the
commit or tag, and each NOTICE carries an explicit paragraph saying **the archive digest is a fetch-time
measurement and the FILE digests are the binding ones** — because a generated tarball's compression is
GitHub's, while the commit and the file bytes are upstream's. The three release-asset families
(`Lora`, `PlusJakartaSans`, `roboto-3-classic`) record the published archive directly.

**One archive digest confirms the whole regime by accident.** `Roboto_v3.016.zip` re-downloaded to
`1653dbe1…`, byte-identical to the digest `public/fonts/roboto/NOTICE.md` has recorded since Story 8.5,
and its `OFL.txt` at tag `v3.016` to `06140232…`, likewise. `Roboto Condensed` is taken from that same
archive, from `android/static/` beside `Roboto-Regular.ttf`.

**`Roboto Slab` is the batch's one non-OFL face, and it required a decision at a named place.**
`font-catalogue.test.ts`'s `licenceSignatures` is a deliberately closed table — *"admitting a licence to
the catalogue is a decision somebody makes here rather than a silent hole"* — so `'Apache-2.0':
/Apache License/i` was added there, with the reasoning inline. Admission itself is unchanged:
`Apache-2.0` is the second of the owner's four ids (D-8.5.3) and was already on `lint`'s
`fontAssetLicenceAllowlist`; what is new is that a catalogue face now exercises it. `lint/MANIFEST.md`
classifies it `Apache-2.0` from the committed text.

**TASK 3 — ten catalogue rows, `scripts` read from each binary's own `cmap` in both directions.** All ten
declare `["latin"]`: each maps all six Latin probes and **none** of the four Thai or four CJK probes, so
neither half of the cross-check is vacuous. `family` is byte-equal to the index snapshot's spelling for
all ten, which is what makes `localTierHolds` suppress the web row.

**TASK 4 — all THREE population floors raised 20 → 31 in one commit**, each carrying a comment naming the
other two (`src/font-catalogue.test.ts:301`, `src/font-index.test.ts:117`, `src/font-name-table.test.ts:35`).
D-16.R.18's correction is quoted at each site: *"a floor that exists in three files is three floors."*

**TASK 5 — `lint/MANIFEST.md` regenerated** with `cd lint && go run ./cmd/genmanifest`, not hand-edited;
ten rows added. Note for the next batch: `ResolveAssets` resolves **git-tracked** files, so the new
directories had to be staged before the regeneration produced anything at all — an unstaged batch
regenerates to a no-op that looks like success.

**TASK 6 — the offered-once assertion extended, and the count assertion added and red-proved.**
`src/font-index.test.ts` now names the batch explicitly (deriving it from the catalogue would be a
tautology), asserts each added family is offered **exactly once and from the local tier**, and asserts the
addable count **rose by exactly ten** — recomputed from the index and the catalogue-minus-the-batch rather
than compared against a number typed in from a previous run. **Red-proof, run:** removing the `lora` row
from `font-catalogue.json` and rebuilding reds **four** assertions — the floor, the variable-upstream
membership, the offered-once loop, and the delta — each with its own sentence.

**TASK 7 — the batch recorded in `_bmad-output/specs/spec-fonts/font-catalogue.md`** under the heading
*"The Story 16.1a batch — the tier grows to 31, by rule, with an owner and a re-run trigger"*, which
sits inside the section the task names (*"The local face tier, and the tier beside it"*): the membership
rule as a **ranking** rule with its halt-on-overflow
clause, the itemised exclusions, the measured size, the **owner** and the **re-run trigger**.

**The owner was APPOINTED BY THIS STORY rather than received from the gate, and that is flagged.** The
gate artifact said plainly that the build cannot appoint one; TASK 7 requires a named owner. The name
written is **the engineering lead**, as the standing holder of this epic's font-tier decisions and the
same authority that replaced the criterion at D-16.R.16 — with the appointment marked in
`font-catalogue.md` as overridable by naming someone else there.

**TASK 8 — DW-166 registered** (the standing curation obligation, with its three triggers), and
**DW-162 updated** with what this story consumed: `s1.assetCount` **44 → 54**, margin against
`maximumCacheAssets` (64) **20 → 10**, cap not moved.

**`source` stops being a URL, on BOTH tiers (D-16.R.13).** Committed tier: `build-wasm.mjs`'s new
`committedFaceSource` **parses each face's own NOTICE.md** — never retypes it — and emits
`googlefonts/roboto-3-classic@v3.016 — android/static/Roboto-Regular.ttf, fetched 2026-09-02`. Fetched
tier: `webFaceSource` emits `google/fonts — ofl/kanit/Kanit-Regular.ttf, fetched 2026-09-03`, replacing
the bare `…/main/…` branch URL. The tripwire is `src/font-provenance.test.ts`: over **both** tiers, no
scheme, no host, no branch name, no SHA-256 — plus an assertion that each tier has exactly **one**
writer, since a second assignment site is how the branch URL survived a suite that already had opinions
about this field. Registered for Story 15.3 as **DW-167** on D-000.15's running list.

**One clause of D-16.R.13 was NOT implemented, and it is raised rather than resolved silently.** The
ruling — and this story's own task bullet — say the committed tier should inline *"the pinned upstream
release **and the committed digest**"*, while the same ruling lists *"not the SHA-256"* among the three
things `source` must never carry, because duplicating the asset key creates two authorities on one fact.
For a committed-tier face those are the same value: the face is stored under the SHA-256 of its own
bytes. **The prohibition was taken as governing** — it is the clause carrying the reasoning and it is
stated absolutely in both documents — and the substantive half (stop pointing at a `NOTICE.md` the
recipient does not have) is implemented in full. Recorded in DW-167 with the one-line change that would
reverse it, if the lead meant the digest to be inlined anyway.

**Payload delta, measured and handed to Story 15.2's budget gate (D-16.R.21).** Story 15.0 is `backlog`,
so D-8.4d.1's fetch-on-first-pick is policy with no implementation and these ten faces are **precached
today**. `brotli.totalBytes` **15,757,356 → 16,681,316**: **+923,960 first-load Brotli bytes**, the sum of
the ten new immutable assets at quality 11 (Open Sans 72,925 · Roboto Mono 64,357 · DM Sans 34,059 ·
Montserrat 138,954 · Arimo 209,197 · Roboto Slab 74,024 · Lora 78,263 · Roboto Condensed 159,452 ·
Oswald 44,489 · Plus Jakarta Sans 48,240). Raw cached bytes rise 2,238,616; `brotli.catalogue.totalBytes`
is 3,151,569 across 31 families.

### Verification, as run

| command | working directory | result |
|---|---|---|
| `npm run scan:font-hosts` | `folio-designer` | **exit 0**, 0 occurrences in 599 tracked files. No new host: this story commits bytes. *(It reds first — twice — on my own PROSE: both new comments spelled the fetch host while describing the URL they were removing. Reworded. The scan reads raw text and does not care that the string is a comment, which is the point.)* |
| `npm run lint` | `folio-designer` | clean (4 pre-existing `react(only-export-components)` warnings) |
| `npm run test` | `folio-designer` | **514 passed, 1 failed** — `canvas-authority-contract.test.ts:190` only, the known DW-152 Epic 9/10 baseline red |
| `npm run build` | `folio-designer` | **exit 0** through `build:wasm`, `tsc -b`, `vite build`, `build:offline`, `verify:offline`. **`s1.assetCount` = 54**, `maximumCacheAssets` 64, margin **10**; all 31 `catalogue-*.ttf` among the 54 |
| `npm run test:e2e:compile` | `folio-designer` | **exit 0**, compile-only; no browser run (D-16.R.1 cadence) |
| `go test -count=1 ./...` | `lint` | **green**, including `TestManifestUpToDate`, `TestFontAssetLicenceAllowlistIsTheOwnersFourIds` and `TestLicenceSignalCensus` |
| `go test -count=1 ./...` | `folio-go` | green except `internal/text`'s `TestCorpusMeetsP6ExerciseFloors/P6g` — the mandated pre-existing exercise-floor red |

**`-count=1` IS NOT OPTIONAL HERE, AND THE FIRST RUN OF THIS STORY PROVED IT.** `cd lint && go test ./...`
reported green on every package while `internal/licence` was in fact **red**: it printed `(cached)`,
because Go caches on the package's own inputs and `TestLicenceSignalCensus` walks the **filesystem** for
`LICENSE*` texts. Ten new licence texts landed, the test's verdict changed, and the cache served a PASS
recorded before they existed. **A batch that adds redistributed licences and reads a cached green has
measured nothing.** The stale report is corrected here rather than quietly overwritten, and the reason is
written into the census table's own header comment so the next batch meets it before the trap does —
**and registered as DW-168**, stated as the general shape rather than as a Go anecdote, because the
people who need the warning are precisely the ones who never open that file.

**What the census actually required (AD-26).** All ten texts classify correctly and none was
unclassifiable — the failure was that their verdicts were **not RECORDED**. Ten rows were added to
`pinnedCensus` in path order: nine `LICENSE-OFL.txt` at `(permissive, "OFL-1.1")` and
`robotoslab/LICENSE-APACHE.txt` at `(permissive, "Apache-2.0")`, **the first Apache-2.0 font asset this
repository has ever carried** — the first row exercising the second of the owner's four ids from the
asset side rather than from a dependency or a fixture. Nothing was deleted from the population and
nothing was narrowed. Census after the fix: **67 licence texts measured (58 committed files + 9
dependency licences), all matching their pinned verdicts** — the same 67 the failing run reported, since
the population is read off disk and already held the ten; what changed is that all 58 committed files now
have a written verdict. **Red-proved:** removing the `robotoslab/LICENSE-APACHE.txt` row reds
`TestLicenceSignalCensus` naming that exact file, and restoring it greens.

**Manual checks.**

- `shasum -a 256 fixtures/*/expected.pdf` **run from the repository root** — printed **23 lines** before
  any comparison, per the Spec Change Log's item 9, and every digest matched
  `folio-go/byte_neutrality_test.go:100`'s `goldenDigestRecord` (23 recorded, 23 measured, 0 unmatched).
- `folio-designer/src/release-payload.ts:33` still reads `const maximumCacheAssets = 64` — unchanged, and
  it must not be moved anywhere in this epic (D-16.R.16).
- **The golden question was verified, not reasoned through** (D-16.R.13's guardrail). One fixture feeding
  a golden **does** embed a face — `fixtures/embedded-font/input.folio` — and its `source` is
  hand-authored JSON that neither writer produces, so no golden reaches either changed code path.

**What is NOT done, and is deliberate.**

- **No browser run.** Compile-only e2e by D-16.R.1's cadence; nothing in this story changes a rendered
  pixel, and the browser witness belongs to Story 16.3's dispatch.
- **`Noto Sans`'s predicate defect (D-16.R.17) is untouched.** It is a real defect against Story 16.1's
  shipped AC and it costs zero slots, but it is not in this story's Tasks and fixing it here would be the
  going-looking D-000.15 forbids. It stays where D-16.R.17 put it.
- **`IBM Plex Sans` stays unofferable**, per the same ruling: promoting a chrome asset to a document
  asset is a different story with its own licence obligations.
