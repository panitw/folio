---
title: 'Story 8.5: A curated catalogue ships with the designer'
type: 'feature'
created: '2026-09-02'
status: 'done'
baseline_revision: '1a4cceaa81f65cd9899eb13efd2ef207d2394c3f'
review_loop_iteration: 0
followup_review_recommended: true
context: ['{project-root}/_bmad-output/implementation-artifacts/8-5-catalogue-procurement.md']
warnings: ['oversized']
deferred:
  - summary: >-
      The forbidden-host scan reads only git-TRACKED files, so a new untracked source
      file containing a forbidden host passes `npm run build` clean.
    evidence: |-
      `scannedPopulation` shells out to `git ls-files`, which reads the index. The guard's own
      framing — "the moment someone reaches for the cheap shape, the first thing they type is one
      of these two hosts" — describes exactly the pre-`git add` moment it cannot see. The remedy
      (`--others --exclude-standard`) is one flag, but it would also fail the build over a
      developer's un-ignored scratch file, so the noise/coverage trade is a decision rather than
      a drive-by patch.
    location: >-
      folio-designer/scripts/forbidden-font-hosts.mjs:197
    severity: medium
  - summary: >-
      Only the two Google Fonts hosts are forbidden; every other live font service would pass.
    evidence: |-
      `FORBIDDEN_FONT_HOSTS` lists `fonts.googleapis.com` and `fonts.gstatic.com`. `use.typekit.net`,
      `fonts.bunny.net`, `fonts.cdnfonts.com` and jsDelivr/unpkg Fontsource paths are all
      "reach for the live font service" shapes D-8.5.12 declined, and the scan waves them through.
      Widening the list is a policy call about what the gate is for, not a defect in this change.
    location: >-
      folio-designer/scripts/forbidden-font-hosts.mjs:41
    severity: medium
  - summary: >-
      `blankComments` approximates comment syntax for several languages it is applied to, while
      the file's prose claims precision.
    evidence: |-
      It treats `//` as a comment opener in `.css` and `.json`, where it is not, so an unquoted
      CSS `url(https://...)` blanks the rest of the line; Python triple-quoted strings and Go raw
      strings are likewise approximate. It fails SAFE — the scan reads raw text and only the
      EXEMPTION path is affected, so the worst case is a declaration that stops exempting — but
      the comment block asserts a fidelity the scanner does not have.
    location: >-
      folio-designer/scripts/forbidden-font-hosts.mjs:106
    severity: low
  - summary: >-
      `npm run build` now fails outright outside a git checkout, because the host scan is its
      first step and cannot obtain a population.
    evidence: |-
      Building from a source tarball, an export, or a Docker `COPY` without `.git` now fails at
      step one for a reason unrelated to the build. The fail-closed throw is CORRECT and must not
      become a warn-and-skip (that is the vacuous green Design Note 2 names), so the resolution is
      a documented build requirement or a non-git walk that preserves fail-closed — not a bypass.
    location: >-
      folio-designer/package.json:18
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Non-normative — owner summary; the intent contract below governs implementation.*

Twenty-one typefaces now ship inside the designer itself. They are not fetched, not downloaded on
first use, and not looked up anywhere: they travel in the release, and a machine with no network at
all can open the designer and find every one of them present and intact. Each face carries its own
licence text and its own copyright line, and the record beside it — where it came from, which
version, and a fingerprint of the exact bytes — is checked against the bytes rather than trusted.
Each face also states its licence inside the font file, and that statement is now read and matched
against what we claim about it.

The release is about two megabytes heavier compressed, and that is the expected price. The size
limit is still deliberately unset: it is fixed once at the end of this run against the finished
weight, never nudged along the way to whatever the build happens to weigh, so nothing here sets or
moves a threshold.

A family or two are absent, and the reason is narrower than it looks. Their licences are perfectly
acceptable and the fonts are genuinely free to use — the checker simply cannot yet read that
particular *wording*. Widening it later brings them straight back.

Picking one of these families still does nothing to a document. That is the next story's job; this
one only puts the faces there and proves they are there.

<intent-contract>

## Intent

**Problem:** The designer offers only the six faces it was built with. An author who wants any other
typeface has no route to one, and nothing records what each face costs the offline release.

**Approach:** Ship a curated catalogue of at least twenty additional permissively-licensed static
faces inside the designer's offline release, behind the same verified asset URLs as every other
release asset; record each face's provenance and its per-asset Brotli weight; and add a source scan
that fails the build if a forbidden font host appears in the scanned source. The catalogue is
binaries plus provenance plus a weight record — no picking behaviour (Story 8.6), and **no licence-gate
change of any kind**.

## Boundaries & Constraints

**Always:**
- **Bundled and precached, not fetched (D-8.5.12).** `SPEC.md` `## Non-goals` forbids a live font
  service, an arbitrary URL and "download on first use". The owner-brief clause asking for *"a shape
  where size is not paid for at first load"* is **declined on the contract** — recorded, priced
  (~15–18 MB against a `~9 MB` commitment) with its reversal named. **This story invents no middle
  tier.**
- **Vendored static route per face (D-8.5.4).** Provenance — pinned upstream version, NOTICE recording
  the upstream digest, the committed digest, byte size, fetch date, and the path inside the archive.
  Reproduction/derivation is **not** required and **not available**. Every face and its pinned tag
  comes from `8-5-catalogue-procurement.md`, which is normative for procurement.
- **Licences: the owner's four ids only** — `OFL-1.1`, `Apache-2.0`, `MIT`, `Ubuntu-font-1.0`
  (D-8.5.3). Admission is per-term; unclassifiable fails the build. **The gate already does this
  correctly — satisfy it, never touch it.**
- **No `WITH` expression in any procured licence (D-8.5.17).** Where a family is excluded for this,
  the record says **"pending a parser widening"** and never *"that licence is unacceptable"*.
- **The census pin stays a hand-written literal** derived from nothing under test (D-8.4j.2).
- **Byte identity.** No `.folio` change, no engine change, no `folio-go/fonts/` change. The 23 golden
  digests must be **byte-identical** to the pre-dispatch snapshot. `maximumCacheAssets` stays **64**.
- **Every guard added is mutation-proved**: delete it and a *named* test reddens on *its own* message.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- The catalogue cannot reach **≥20 new families** from `8-5-catalogue-procurement.md`'s pool and its
  reserve. **Do not widen the allowlist to fit the population** — that is the defect the gate exists
  to prevent.
- Any face requires **derivation** (a new `UPSTREAM` entry). The bootstrap gap is real.
- `maximumCacheAssets` would have to be raised, a golden digest moves, or
  `S1 CJK row is not the dominant font payload` fires.
- A catalogue face's row in `lint/MANIFEST.md` reads `SEE NOTICE`, or the build fails on a licence.
- **Seeding the compound-expression census case reddens any licence test other than the census
  itself.** Repairing that would be licence-gate work, which this story may not do — halt instead.
- *(Deliberately **not** a Block If: a licence-gate defect. See the `Never` list — it registers to
  Epic 15 and the story continues. This story is the one **most likely** to surface that class, and
  D-000.13 pre-committed the response before contact.)*

**Never:**
- **Any licence-gate change** — no edit to `classify.go`, `manifest.go`, the allowlists, the
  classifier tables, or any gate rule. **Flat prohibition (D-000.11), not a criterion.** A licence-gate
  defect registers to Epic 15's Story 15.3; it is never fixed, split out, or given an AC here.
  **Nor escalated** — D-000.13 closes that loophole in advance: escalation is for a finding that makes
  the epic's *stated goal unshippable*, not one that makes it imperfect, and *"an escalated
  licence-gate finding is the same failure wearing a different hat."* This story puts twenty `LICENSE*`
  files through the gate for the first time and is **the story most likely to surface this class**;
  **every such finding registers.** If a specific face is refused, **drop that face and take one from
  the reserve** — the remedy is procurement, never the gate.
- Bold, italic, oblique, synthetic emboldening, or variable-font axes. **Epic 11 (FR57) owns
  realize-vs-retire and the owner ruling has not been made (D-000.7).**
- CJK families in the catalogue; the shipped SC face remains the coverage fallback.
- A live font service, an arbitrary URL, or a "download on first use" (`SPEC.md` `## Non-goals`).
- Enumerating or reading host-installed fonts (`## Non-goals`, AD-8).
- Save-time subsetting, or any change to how the PDF producer subsets.
- A container format — the catalogue is committed files, never a zip-of-folders.
- Picking behaviour, chain proposal, or embedding on pick — **Story 8.6**.
- A byte threshold or first-load budget gate — **Story 8.4d owns it**, set last against finished
  weight. This story **records** weight and **must not** set or move a threshold. `epics.md`'s
  superseded `~9 MB` figure stays as-is.
- `.woff2`/`.woff`/`.ttc` faces: the engine decodes only `font/ttf` and `font/otf`.
- Editing `epics.md`, `ARCHITECTURE-SPINE.md`, or `<intent-contract>` to match what was built.
- **Claiming "no request leaves the machine."** A source scan proves no forbidden host appears in the
  **scanned population**, and nothing more (D-8.5.5).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Face admitted | Face dir with `LICENSE*` (`OFL-1.1`) + `NOTICE*` with a Copyright line | `ResolveAssets` returns a row; `MANIFEST.md` carries a real SPDX id | No error expected |
| Compound admitted | Licence line `MIT OR Apache-2.0` | Every term on the four-id allowlist → admitted | No error expected |
| Off-allowlist term | Licence line `OFL-1.1 OR CC0-1.0` | Build **fails**, naming the failing term | Build failure is correct |
| `WITH` form | `OFL-1.1 OR GPL-2.0-or-later WITH Font-exception-2.0` | Build **fails** closed | Correct; family is excluded pending a parser widening |
| Missing provenance | Face committed with no `NOTICE*` | Build **fails** | Correct |
| Forbidden host in source | `fonts.googleapis.com` in a `.ts` file | Scan **fails**, naming file and line | Build failure |
| Forbidden host in a comment | Same host inside a `//` comment | Scan **fails** — comments are stripped *from the exemption*, not from the scan | Build failure |
| Asset count | 23 existing + ~21 new faces | Under `maximumCacheAssets` (64) | Over → Block If |

</intent-contract>

## Code Map

**Designer bundle (all npm commands from `folio-designer/`):**
- `folio-designer/public/fonts/<family>/` -- where the 6 shipped faces live (`.ttf` + `LICENSE-OFL.txt`
  + `NOTICE.md`). New catalogue faces follow this exact shape. `vite.config.ts:32` sets
  `publicDir: false`, so `public/` is **never** published directly — fonts enter the bundle only via
  the generator below.
- `folio-designer/scripts/build-wasm.mjs:88-98` -- the `assets` map: **six hardcoded slots**, no
  manifest loop. `fingerprint()` at `:69-77` content-addresses each `.ttf` into `src/generated/runtime/`.
  Emits `src/generated/offline-assets.ts` (`:103-104`) and `src/generated/runtime-fonts.css` (`:134`).
  **This literal is the main structural change**: ~21 faces means a manifest-driven loop, not 21 more keys.
- `folio-designer/scripts/build-wasm.mjs:134` -- `@font-face` emitter, **hardcodes `format('truetype')`**
  and `font-display: swap`, with no `font-weight`/`font-style` descriptors. `.ttf` only avoids a branch here.
- `folio-designer/scripts/generate-offline-release.mjs:13-18` -- `assetsFromDist()` builds the precache
  list. `:35-39` writes a Brotli `.br` sidecar per immutable asset (quality 11, LGWIN 22). `:40-53`
  builds the S1 payload; `cachedRow()` at `:45` stats the `.br`.
- `folio-designer/src/release-payload.ts:32-33` -- `minimumCacheAssets = 10`, `maximumCacheAssets = 64`.
  **Single authority.** Must keep the exact `^const <name> = <digits>$` single-line form — the reader
  (`scripts/offline-release-contract.mjs`, `readDeclaredConstant()`) regex-matches it and throws on
  zero or two matches.
- `folio-designer/scripts/verify-offline-release.mjs:50-52` -- the count bound
  (`over the declared maximum of`). `:99` -- `S1 CJK row is not the dominant font payload`; it compares
  **only** the 3 S1 rows whose id ends `font`, so plain cache assets never enter it. `:107-117` --
  every immutable asset's sidecar must decompress to the original and **re-compress byte-identically**.
- `folio-designer/src/font-binary-identity.test.ts` -- opens each file's `name` table, ties each
  `@font-face` family to its own source, asserts one file per family and the NOTICE-recorded digests;
  repo-wide magic-byte sweep at `:764-769` with a >500-file floor at `:751`.
- **Measured headroom:** built manifest has **23** assets, `s1.cachedBytes` 38,460,833. 64 − 23 = **41
  free**; ~21 new faces fit. Brotli today: engine 7,226,258 / latin 226,026 / thai 24,872 / cjk 4,948,312.

**Licence side — READ-ONLY except the census pin (commands from `lint/`):**
- `lint/internal/manifest/manifest.go:299-518` -- `ResolveAssets`. Groups fonts by **directory**; a
  directory with 0 git-tracked files is silently skipped (`:381-383`). Requires `LICENSE*` (`:420-422`),
  `NOTICE*` (`:423-425`) and a Copyright line (`:427-430`). The gate is `:460-469`. **Do not edit.**
- `lint/internal/manifest/manifest.go:140` -- `fontAssetLicenceAllowlist` = the four ids. **Do not edit.**
- `lint/internal/manifest/manifest.go:109-110` -- `fontExtensions` = `.ttf .otf .ttc`, keyed on
  **extension**. D-1.8.11 forbids widening it; its miss is silent.
- `lint/internal/manifest/manifest.go:257-266` -- `assetServesLabel`. New designer fonts land in the
  `committed asset (<dir>)` bucket **automatically — no code change needed.**
- `lint/internal/licence/classify.go:51-79` -- `ClassifySPDXExpressionTerms`. Parenthesis or empty →
  error; even field count → error; odd indices must be `AND`/`OR`. **`WITH` is unsupported at either
  position** → `FamilyUnknown` → refused.
- `lint/internal/licence/classify.go:235` -- the SPDX line regex captures **the rest of the line**. A
  new face's LICENSE must not carry a comment terminator or parenthetical on its SPDX line
  (`SPDX-License-Identifier: MIT */` classifies unknown → build failure).
- `lint/internal/licence/classify.go:167-171` -- records that **nothing in this repo is
  Ubuntu-licensed today**, and that the analogue test cannot be written *"until Story 8.5 lands a face
  under it"*. Procuring Ubuntu Sans / Ubuntu Sans Mono is that face.
- `lint/internal/licence/licencecensus_test.go:46-82` -- `pinnedCensus`, **35 hand-written rows**,
  keyed on repo-relative path. Population A is *discovered* via `git ls-files`; an unpinned discovery
  errors at `:130-133`, an unfound pin at `:198-201`. **Contains no compound expression today.**
- `lint/cmd/genmanifest/main.go` -- regenerates `lint/MANIFEST.md`. **Runs from inside `lint/`**
  (`RELEASING.md:36-37`); from the repo root it fails with "cannot find main module".
- `lint/MANIFEST.md:299-310` -- existing asset rows. `SEE NOTICE` is now **unreachable** and two tests
  forbid its return.

**Provenance template:** `folio-designer/public/fonts/ibmplexsans/NOTICE.md` -- the
no-derivation shape every new face copies. (`folio-go/fonts/notosans/NOTICE.md` is the *derived*
variant with a toolchain table — **not** the template here.)

**Absent today:** there is **no forbidden-host source scan** anywhere, and **no per-asset Brotli
reporting** — both are net-new. `fonts.googleapis.com`/`fonts.gstatic.com` appear **zero times** in any
source file, so a naive scan passes vacuously on introduction. See Design Note 2.

## Tasks & Acceptance

**Execution:**
- `_bmad-output/implementation-artifacts/8-5-catalogue-procurement.md` -- procure each Tier A face at
  its pinned tag; record the actual fetched digests and dates back into this table -- the table is
  normative, and a placeholder digest is not provenance.
- `folio-designer/public/fonts/<family>/` -- one directory per face: the `.ttf`, the unmodified
  upstream `LICENSE*`, and a `NOTICE.md` on the IBM Plex Sans shape -- this is what the licence gate
  reads; a missing Copyright line fails the build.
- `folio-designer/scripts/build-wasm.mjs` -- replace the 6-key `assets` literal with a
  manifest-driven loop over the catalogue directories -- 27 hardcoded keys is not maintainable and
  each new face would otherwise be a hand edit in three places.
- `folio-designer/scripts/generate-offline-release.mjs` -- emit a **per-asset Brotli record** into the
  release manifest -- AC5; nothing records this today.
- `folio-designer/scripts/verify-offline-release.mjs` -- assert the Brotli record matches the emitted
  sidecars, with a `brotli-record-drift` red proof -- an unasserted record drifts silently.
- `lint/internal/licence/licencecensus_test.go` -- add one `pinnedCensus` row per new `LICENSE*`
  **and** seed the single compound-expression case (Design Note 3) -- D-8.4j.2; keep the literal
  hand-written.
- `lint/MANIFEST.md` -- regenerate via `cd lint && go run ./cmd/genmanifest` and commit -- the
  committed output is pinned by `TestManifestUpToDate`.
- **new** forbidden-host source scan + its tests -- AC4; must strip comments and carry a population
  floor and a positive control (Design Note 2).

**Acceptance Criteria:**
- **AC1 —** Given a procured catalogue face, when its directory is inspected, then it carries the
  unmodified upstream `LICENSE*`, and a `NOTICE.md` recording the pinned upstream version, the upstream
  archive digest, the committed digest, the byte size, the fetch date and the path inside the archive;
  and `shasum -a 256` of the committed binary **equals** the digest its own NOTICE records.
- **AC3 —** Given the built offline release, when the release manifest is read, then it carries **at
  least 20 new families** beyond the 6 already shipped, each behind a verified content-addressed asset
  URL covered by the offline verification job, and the asset count remains **at or under 64**.
- **AC4 —** Given a source file containing a forbidden font host, when the scan runs, then the build
  fails naming the file and line; and given the same host appearing **inside a comment**, then the
  build still fails — the scan strips comments so a commented-out host cannot exempt itself. The scan
  reports the size of the population it scanned, and fails if that population is implausibly small.
- **AC5 —** Given the built release, when the release manifest is read, then **every** immutable asset
  carries its own recorded Brotli byte count, each equal to its emitted `.br` sidecar; and the
  catalogue's **total added Brotli bytes** is reported as **one number**, with its command, commit,
  tree state and working directory (D-8.4j.8). **No threshold is set or moved.**
- **AC6 —** Given every catalogue face, when its `name` table and `OS/2` fields are read, then each is
  a single upright static Regular instance — no bold, italic, oblique or variable axes — and the
  generated `@font-face` rules declare no `font-weight` or `font-style` variants.
- **AC8 —** Given the completed story, when the goldens and bounds are checked, then all **23** golden
  digests are byte-identical to the pre-dispatch snapshot, `maximumCacheAssets` is still **64**,
  and `git diff` touches no `folio-go/fonts/`, no `fixtures/**`, and no `.folio`.

## Spec Change Log

### 2026-09-02 — SUPERSEDED at the plan gate, pending re-plan after Story 8.4h (D-8.5.13)

This spec was planned as one story and the plan gate split it. **Do not dispatch it to build as
written.** The engineering lead ruled `multiple-goals` a real signal rather than a nuisance: the
licence gate's failure is live at HEAD *today*, with zero catalogue faces in the tree — measured at
`lint/internal/manifest/manifest.go:298-301`, a font whose licence text does not classify gets
`"SEE NOTICE"` and `ResolveAssets` returns nil, so a GPL font ships with a clean build. That is a
repair, and a repair does not depend on the extension that surfaced it.

**Moving out to Story 8.4h:** AC2 (both fall-through sites), the `UFL` addition to `classify.go`, and
AC7 (the widened extension-class guard population), with Execution tasks 1–4.

**Staying here:** AC1, AC3, AC4, AC5, AC6, AC8. **AC4 stays deliberately** — the forbidden-host scan
is the assertion that D-8.5.12's decline was honoured, so it belongs with the thing it constrains, not
with the licence gate.

**Design Note 1 is corrected by D-8.5.12.** Its claim that no cheaper shape exists overstates the
case: no shape exists *that the Non-goals clause permits*. A same-origin, cached-on-first-pick tier is
buildable today (a release may carry assets the service worker does not precache; 41 of 64 slots are
free). It is policy, not physics, and the record says so — the decline is recorded with its measured
price and its reversal, so the owner can revisit it. The story still invents no middle tier, and a
build proposing one is a halt.

**`oversized` is expected to resolve via the split.** If it does not at six ACs, the lever is moving
AC1's twenty-row per-face procurement table into a companion artifact this spec cites — **never**
thinning acceptance criteria, which would be moving the bar to fit the instrument.

### 2026-09-02 — RE-PLANNED to reduced scope after D-000.11 (owner course correction)

Rewritten at `19959fa`. Predecessors 8.4h/8.4i/8.4j are `done`: the licence gate now fails closed,
reads a whole SPDX line, and admits **per-term** against the owner's four ids. **AC2 and AC7 are gone
— they shipped in 8.4h.** Six ACs remain, unchanged in substance from the split.

**D-000.11 binds this spec.** No licence-gate work, for any reason — a flat prohibition, not a
criterion. A licence-gate defect found here registers to **Epic 15 (Story 15.3)** and the story
continues. The `## Code Map`'s licence-side anchors are therefore marked **read-only**: they are
present so the implementer can satisfy the gate without searching for it, never so it can be edited.

**`multiple-goals` cleared; `oversized` did NOT, and the honest figure is recorded rather than the
flattering one.** `multiple-goals` went with the split. The named lever *was* applied — the per-face
procurement table moved out to `8-5-catalogue-procurement.md`, exactly as the previous entry
predicted, and **no acceptance criterion was thinned** (AC4 and AC5 gained detail rather than losing
it). The spec fell from 40,568 to 29,546 characters, ~10,100 to ~7,400 estimated tokens — **27%
smaller**, and less than half the size of the next-smallest spec in this epic (8.3, ~16,300). **But
the flag's threshold is 1,600 tokens, and it still fires.** Every spec in Epic 8 trips it (8.4f is
~25,900), so `oversized` here marks a project-wide convention gap, not a defect in this spec.
Clearing it at six ACs would require thinning them, which the previous entry expressly forbade as
"moving the bar to fit the instrument". **The flag stays set.**

**The `<intent-contract>` was rewritten rather than preserved.** Step-02's draft-resume rule preserves
the contract verbatim; that was **not** applied, because the preserved contract carried the superseded
scope — AC2, AC7, the `UFL` addition to `classify.go` and the widened extension-class guard, all now
either shipped or prohibited. The replacement scope was supplied by the orchestrator in the dispatch,
which is the orchestrator amending its own contract. Recorded because a silent contract rewrite is
exactly what this log exists to catch.

**KEEP on re-derivation:** the plain-terms opener's four-part shape (what it does / what it
deliberately does not / the Libertine-shaped exclusion / the deliberate late budget); the Block If
against widening the allowlist to fit the population; and the D-8.5.5 prohibition on claiming "no
request leaves the machine".

## Review Triage Log

### 2026-09-02 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 7: (high 0, medium 5, low 2)
- defer: 4: (high 0, medium 2, low 2)
- reject: 8: (high 0, medium 2, low 6)
- addressed_findings:
  - `[medium]` `[patch]` **The collision guard's shipped-family half could never fire.** `build-wasm.mjs` seeded `catalogueFamilies` from `Object.keys(assets)` — measured as the SLOT keys `wasmExec, wasm, starter, sans, sansCjk, sansThai, mono, plexSans, plexSansThai`, not family names — so its error message promising "or over a family the six shipped rules already declare" was half false, and an entry declaring `Noto Sans` would have emitted a duplicate `@font-face`. Reseeded from a `shippedFamilies` literal of the six real family names, plus two new throws tying that literal to the hand-written stylesheet it describes. Proved independently: an entry with family `Noto Sans` now throws; before the fix it did not.
  - `[medium]` `[patch]` **`font-catalogue.json`'s `licence` field was dead data.** Nothing under `folio-designer` read `.licence` or `"licence"`; it duplicated a fact held in the NOTICE, `lint/MANIFEST.md` and `pinnedCensus` and could silently disagree with all three. It now has a consumer (below) and is in the required-field validation.
  - `[medium]` `[patch]` **No catalogue face was checked for its licence in its own bytes, though the three chrome faces are.** `font-binary-identity.test.ts` holds each chrome face's nameID 13 to the SIL OFL; `font-catalogue.test.ts` had adopted that file's digest check but not this one, so swapping a binary and its NOTICE together — same family, different terms — passed every check while `lint/MANIFEST.md` published a licence the bytes contradicted. Added a per-face nameID-13 assertion keyed off the `licence` field via a closed `licenceSignatures` table (an unrecognised id fails rather than skipping). Premise measured first: all 21 faces carry nameID 13; the 19 OFL faces match `/SIL Open Font License/i`, both Ubuntu faces match `/Ubuntu Font Licence/i`. Proved: flipping `ubuntusans` to `OFL-1.1` reds that face by name.
  - `[medium]` `[patch]` **`SCANNED_EXTENSIONS` omitted `.mts`/`.cts`, and this change added the repository's only `.mts` file.** `git ls-files` returns exactly one: `folio-designer/scripts/forbidden-font-hosts.d.mts` — the scanner's own sidecar sat outside the population it claims covers the product source. Both added; the reported population moved 578 → 579.
  - `[medium]` `[patch]` **`directory`, `file` and `family` were unvalidated where they become syntax.** `directory`/`file` are `join`ed into a filesystem path with no rejection of separators or `..`; `family` is interpolated unescaped into `font-family: '<name>'`. The `id` field was already held to `/^[a-z0-9]+$/` with a comment explaining why; the same reasoning now covers all three. Each rejection mutation-proved.
  - `[low]` `[patch]` **The scan's reach assertion covered three of six `SCANNED_ROOTS`.** `hashmatrix`, `tools` and `.github` could have dropped out of the walk with every test green, and `[...reached].sort()` fed to `expect.arrayContaining` made the sort a no-op. Now derived from `SCANNED_ROOTS` and exact.
  - `[low]` `[patch]` **One Brotli branch was not red-proved.** The mutable-entry arm (`asset.brotliBytes !== undefined`, reachable only for `/index.html`) was never exercised: the `brotli-record-drift` proof mutates an IMMUTABLE asset, so deleting that branch left every red proof green. Added `brotli-record-on-mutable-entry`, held to that branch's own message.

**Rejections, enumerated individually (D-8.4j.3 — each names the caller, path or population verified).**

1. *"The 44 ≤ 64 asset count is not written down anywhere."* — Rejected: it is **enforced**, which the finding treats as merely unwritten. `asset-count-over-bound` and `asset-count-under-bound` are registered red proofs (`verify-offline-release.mjs:303`, `:314`) and fired inside an exit-0 `verify:offline:red`; the bound is read from `src/release-payload.ts:33` by `readDeclaredConstant`. Population verified: the built `dist/offline-release-manifest.json`, 44 assets.
2. *"The Brotli record is a measurement with no owner; add a marker that reds if 8.4d never lands."* — Rejected against the intent contract, whose `Never` list reads "A byte threshold or first-load budget gate — **Story 8.4d owns it** ... This story **records** weight and **must not** set or move a threshold." A failing placeholder is setting one.
3. *"Make the scan warn-and-skip (`process.exit(0)`) outside a git checkout."* — Rejected: the proposed remedy inverts the guard. `forbidden-font-hosts.mjs:198-201` throws "could not look ... must never read as all-clear" deliberately, and an exit-0 warn is precisely the vacuous green Design Note 2 names as the trap. The underlying observation (the build now needs a checkout) is **deferred**, not rejected.
4. *"Under a 'distinct upstream design' reading, the catalogue is ~16 families and AC3's Block If would fire."* — Rejected as re-litigating a settled decision. AC3's text is "at least 20 **new families**"; the dispatch settles the count at 21 clearing by one. Population verified by reading the `name` table of each committed binary, not filenames: **21 distinct family strings across 21 `.ttf` files in 21 directories**.
5. *"'21 catalogue faces' is hardcoded in 21 NOTICEs and will drift."* — Rejected: verified no code path parses that sentence (`grep` across `folio-designer` and `lint` for the phrase finds only the NOTICE prose). The NOTICEs are provenance records of a fixed procurement event, and a count inside a historical record is not a live claim.
6. *"`cousine`'s 'Copyright 2026' is the shape of a transcription slip."* — Rejected on measurement: the binary's own nameID 0 reads `Copyright 2026 The Cousine Project Authors (https://github.com/googlefonts/cousine)`. The NOTICE and the `MANIFEST.md` row transcribe it verbatim. Upstream really says 2026.
7. *"`geistmono` records its upstream URL with a `.git` suffix while `geist`, from the same archive, does not."* — Rejected as factually wrong. Both NOTICEs record the identical `Download URL` (`.../v1.7.2/geist-font-v1.7.2.zip`) and the identical archive digest `7fc800d2...` at 8,207,303 bytes. There is no `.git` suffix in either file.
8. *Hardening for symlinked paths, truncated sfnt tables, and `name` table format 1.* — Rejected: the population is 21 committed binaries that I parsed successfully for nameIDs 0/1/2/13/16/17, `OS/2`, `head` and `post`; none is a symlink, truncated, or format 1. These guard inputs that cannot reach the site without a separate commit that would fail AC1's digest check first.


## Design Notes

**1 — The decline is policy, not physics, and stays recorded.** A same-origin, cached-on-first-pick
tier is buildable today: `immutable` and the S1 row set are separate mechanisms, and 41 of 64 slots
are free. It is foreclosed by *"no download on first use"*, which names it literally. The price is
recorded (~15–18 MB against `~9 MB`) so the owner can reverse it cheaply — the precache set is a
**designer** artifact, so moving catalogue faces to a deferred tier later breaks no integrator. **This
story invents no middle tier; a build proposing one is a halt.**

**2 — AC4's scan would pass vacuously, and that is the trap.** The forbidden hosts appear **zero
times** in source today, so a scan introduced now is green before it is correct and would stay green
if it scanned nothing at all. Three things are therefore required, not optional: a **positive control**
(a fixture that *does* contain the host, asserted to fail), a **population floor** (the scan reports
how many files it examined and fails if that number is implausibly small — the pattern
`font-binary-identity.test.ts:751` already uses), and a **comment-direction test** (the host inside a
comment must still fail; deleting the comment-stripping step must red that test on its own message).
A scan whose only evidence is "it passed" is decoration. And per D-8.5.5, what it proves is bounded:
no forbidden host appears in the **scanned population** — never that no request leaves the machine.

**3 — The compound-expression census case cannot come from a font, and the domain says so.** D-8.4j.2
requires the census to pin at least one compound case so the compound-line fix stays fixed. Measuring
the domain first (D-000.12): **no procurable family carries a compound expression admissible under the
four-id allowlist.** The two compound font licences found — Hack (MIT + Bitstream Vera) and Public Sans
(OFL-1.1 + CC0-1.0) — must both *fail* the gate, correctly. So the case must be seeded as a fixture,
following the pattern already in the census: `lint/testdata/licence/permissive/example.test/<name>-lib/LICENSE`
carrying `SPDX-License-Identifier: MIT OR Apache-2.0`, plus its one pinned row. **This is the only
route consistent with all three constraints** (census must pin a compound case; pins must correspond
to real committed files; the literal stays hand-written), which is why it is not an intent gap. It is
also the boundary of what this story may touch on the licence side: **if seeding it reddens any licence
test other than the census, halt** — repairing that would be licence-gate work.

**4 — `build-wasm.mjs` is the real structural change, not the fonts.** The `assets` map is six
hardcoded keys emitting a fixed six-rule CSS string. Twenty-one more faces by hand would mean 27 keys
maintained in three places. A manifest-driven loop over the catalogue directories is the change; keep
the emitted CSS shape identical (one static Regular per family, `format('truetype')`, no weight or
style descriptors) so AC6 stays observable from the generated file.

**5 — `.ttf` only, and the margin is one family.** The Tier A pool yields **21 new `.ttf` families**
against AC3's ≥20. That is thin, and stated rather than smoothed: the reserve list and the `.otf`
route (which needs only a `format()` branch) exist for procurement failures. `.otf` is *permitted* by
the engine — it is held back to avoid the branch, not because it is unsafe.

## Verification

**Baseline: MEASURE at `19959fa9d9188e9732e658c4e26f825f44ecdd24` in the MAIN CHECKOUT** (never a
linked worktree — DW-105: the wasm is a function of the checkout path), clean tree, **before the first
edit**. Record every figure with its command, commit, tree state **and working directory** (D-8.4j.8).
Do not carry any figure forward from a prior spec. An anchor written at a plan gate is a claim with an
expiry date (D-7.8.4) — re-verify at the gate that consumes it.

**HEAD moved once while this spec was being planned** (`19959fa` → `20002e4`, adding D-000.13). That
commit touches `_bmad-output/` only, so the **code tree is identical to `19959fa`**, where every
figure in the Code Map was measured. The baseline stays `19959fa` as dispatched. **Re-measure HEAD at
the build gate rather than trusting this line** — it is a claim about a moment that has passed.

**Cadence.** Per-epic (D-000.4): the four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity`
and the Playwright suite are **excluded** — Epic 8's boundary gate owns them, nothing engine-side
changes, no `.folio` field is added, and no golden can move. **The D-000.4 exception IS explicitly
invoked** for the designer's release integration set (`npm run build`, `verify:offline`,
`verify:offline:red`, `verify:offline:wasm`) and for the **whole `lint` module suite**, because this
story puts new binaries into the offline bundle — its correctness is bundle-shaped, and the cheap
per-story set would not see it.

**Three standing reds, by identity. Any fourth is a real failure.**
1. `folio-go`: `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)` — expect **1815 pass / 2 fail /
   5 skip**. Two standing decisions forbid filling it. **Never "fix" it.**
2. `gofmt -l folio-go lint` → `lint/internal/rules/licencegraph_test.go` (DW-116). **Do not reformat.**
3. `folio-designer` `npm run lint` → **exactly 4** `only-export-components` warnings.

**Commands:**
- `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` — **REPO ROOT**, before the first
  edit; expect **23** lines.
- `cd folio-go && go test -count=1 ./...` — standing red 1 only.
- `cd folio-go && go vet ./...` — no output.
- `gofmt -l folio-go lint` — **REPO ROOT** (from `lint/` it prints `lstat` errors that read as clean —
  a non-measurement). Expect standing red 2 and nothing else.
- `cd lint && go test -count=1 ./...` — no FAIL. `-count=1` always (rules walk dirs; Go's cache does
  not track `ReadDir`). Run `-run TestLicenceSignalCensus -v` explicitly and **quote what it prints**.
- `cd lint && go run ./cmd/genmanifest` — **from inside `lint/`**; then, from the **REPO ROOT**,
  `git diff --exit-code -- lint/MANIFEST.md` after the regeneration is committed.
- `cd folio-designer && npm run typecheck` — exit 0.
- `cd folio-designer && npm run lint` — exit 0 with standing red 3.
- `cd folio-designer && npm test` — project script only (`npx vitest run` from the repo root picks up
  the Playwright specs and gives a false mass failure). Baseline **40 files / 411 tests**; expect a
  higher count and **zero failures**.
- `cd folio-designer && npm run test:e2e:compile` — exit 0. This is `tsc -p tsconfig.e2e.json
  --noEmit`. **Do not report it as a run** (DW-101: CI executes no Playwright).
- `cd folio-designer && npm run build` — node **v24.16.0** exactly. Record `release.assets.length`,
  `s1.assetCount`, `s1.cachedBytes`, and the **per-asset Brotli** figures with this exact command.
- `cd folio-designer && npm run verify:offline:red` — exit 0; every red proof fires on **its own**
  message, including the new `brotli-record-drift`. `S1 CJK row is not the dominant font payload` must
  **NOT** fire, and `over the declared maximum of` must not fire.
- `cd folio-designer && npm run verify:offline:wasm` — exit 0.
- `shasum -a 256 fixtures/*/expected.pdf | diff <scratch>/digests.before -` — **empty diff**. A moved
  golden is a **HALT**, not a re-record.
- `grep -c 'maximumCacheAssets = 64' folio-designer/src/release-payload.ts` — expect `1`.

**Manual checks:**
- **Mutation-prove every guard added, by deletion.** Delete the comment-stripping step in the
  forbidden-host scan → the comment-direction test reddens on its own message. Delete the population
  floor → the vacuity test reddens. Delete the Brotli-record assertion → `brotli-record-drift` reports
  `escaped verification`. **A guard that survives its own deletion is decoration.** Prove each red
  proof fails for the *right reason*, and beware over-determination — if a population mutation trips
  an earlier check first, assert *"fails on a message that is not this guard's"*.
- **Read `lint/MANIFEST.md`'s asset table by eye:** every catalogue row carries a real SPDX id and a
  real copyright line. A `SEE NOTICE` row is a Block If.
- **Verify each committed face against its own NOTICE:** `shasum -a 256` equals the recorded digest,
  and the byte size matches. This is AC1's observable.
- **Confirm the total added Brotli bytes is stated as ONE number** in the Delivery Log with its
  command, commit, tree state and working directory — so Story 8.4d does not inherit twenty rows
  nobody added up.
- `git diff --stat` per commit must touch **no** `folio-go/fonts/`, no `fixtures/**`, no `.folio`, no
  `_bmad-output/planning-artifacts/`, and **no licence-gate source**. Explicit paths on every
  `git add` — never `-A`, never `.`; never push; never branch.

## Auto Run Result
### Dispatch 3 — 2026-09-02, implement + review + commit

Status: **done**. Blocking condition: none.
Baseline: `1a4cceaa81f65cd9899eb13efd2ef207d2394c3f` (the spec's planning baseline `19959fa` differs
only under `_bmad-output/`; `git diff 19959fa..1a4ccea -- . ':(exclude)_bmad-output'` is **empty**,
so every Code Map figure was measured against this code tree).

**Implemented change.** Twenty-one permissively-licensed static Regular faces now ship inside the
designer's offline release, each behind a content-addressed, digest-verified asset URL, with its own
upstream `LICENSE*`, a `NOTICE.md` recording full provenance, a row in `lint/MANIFEST.md`, and a
recorded per-asset Brotli weight. All Tier A rows procured — no reserve draw, no `.otf` route, no
derivation, no allowlist widening.

**Files changed (one line each).**
- `folio-designer/public/fonts/<family>/` ×21 — the `.ttf`, the unmodified upstream `LICENSE*`, and a
  `NOTICE.md` on the IBM Plex Sans shape (pinned tag, download URL, in-archive path, archive digest
  and size, committed digest, byte size, fetch date, "copied unmodified, no derivation").
- `folio-designer/font-catalogue.json` — **new**; the single declaration of the catalogue.
- `folio-designer/scripts/build-wasm.mjs` — the six hardcoded asset slots keep their named consumers;
  the catalogue is a manifest-driven loop emitting one `@font-face` per face, no weight/style
  descriptor. Now validates `id`, `directory`, `file`, `family`, `licence`, and refuses a family the
  six shipped rules already declare.
- `folio-designer/scripts/generate-offline-release.mjs` — writes `brotliBytes` per immutable asset and
  a `release.brotli` summary carrying the catalogue subtotal.
- `folio-designer/scripts/verify-offline-release.mjs` — holds every recorded Brotli figure to its
  emitted sidecar and to the rows' own arithmetic; two red proofs (`brotli-record-drift`,
  `brotli-record-on-mutable-entry`).
- `folio-designer/scripts/forbidden-font-hosts.mjs` (+ `.d.mts`) — **new**; the AC4 source scan, wired
  as the first step of `npm run build`.
- `folio-designer/src/font-catalogue.test.ts` — **new**; AC1 and AC6 read out of each committed
  binary's own tables, including the licence the bytes declare (nameID 13).
- `folio-designer/src/forbidden-font-hosts.test.ts` — **new**; positive control, population floor,
  comment direction, exact root reach.
- `folio-designer/src/canvas-font-stack.test.ts` — widened, not weakened: admits the catalogue rule
  shape and adds a discrimination case proving a catalogue-shaped rule pointed at a font service reds.
- `lint/internal/licence/licencecensus_test.go` — 21 new pins plus the compound-expression case.
- `lint/testdata/licence/permissive/example.test/compound-lib/` — **new**; the `MIT OR Apache-2.0`
  fixture (Design Note 3 — no procurable font carries an admissible compound expression).
- `lint/internal/manifest/manifest_test.go` — the matrix audit's missing `WITH` row (see below).
- `lint/MANIFEST.md` — regenerated from inside `lint/`.

**Review findings breakdown.** 7 patches applied (5 medium, 2 low), 4 items deferred, 8 items
rejected — every rejection enumerated individually in the Review Triage Log above with the caller,
path or population verified. 0 intent gaps, 0 bad-spec loopbacks.

**Follow-up review recommendation: `true`.** Patched this pass: high 0, medium 5, low 2 →
`3 x 5 + 1 x 2 = 17`, which is ≥ 5. No patched finding was high severity; the score is driven by
breadth, not by any single intolerable defect.

**Verification performed** — every invocation with its working directory (D-8.4j.8).
- `[repo root]` `shasum -a 256 fixtures/*/expected.pdf | diff digests.before -` → **empty. 23/23
  golden digests byte-identical** to the pre-dispatch snapshot, re-checked after the patch pass.
- `[repo root]` `gofmt -l folio-go lint` → `lint/internal/rules/licencegraph_test.go` only (standing
  red 2, DW-116, not reformatted).
- `[repo root]` `grep -c 'maximumCacheAssets = 64' folio-designer/src/release-payload.ts` → `1`.
- `[folio-go]` `go test -count=1 ./...` → **1815 pass / 2 fail / 5 skip**; the two failures are
  `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)` by identity (standing red 1). No fourth red.
- `[folio-go]` `go vet ./...` → exit 0, no output.
- `[lint]` `go test -count=1 ./...` → exit 0, **0 FAIL**, all four packages ok.
- `[lint]` `go test -run TestLicenceSignalCensus -v` → PASS, printing *"CENSUS: 57 licence texts
  measured (48 committed files + 9 dependency licences), all matching their pinned verdicts"*, and
  pinning `.../compound-lib/LICENSE (permissive, "MIT OR Apache-2.0")` — D-8.4j.2's compound case.
- `[lint]` `go run ./cmd/genmanifest` → exit 0; then `[repo root]` `git diff --exit-code --
  lint/MANIFEST.md` → exit 0 (regeneration is a no-op against the committed output).
- `[folio-designer]` `npm run typecheck` → exit 0. `npm run lint` → exit 0 with **exactly 4**
  `only-export-components` warnings (standing red 3). `npm test` → **42 files / 424 tests, 0
  failures** (baseline 40 / 411). `npm run test:e2e:compile` → exit 0 (compile only; CI executes no
  Playwright, DW-101).
- `[folio-designer]` `npm run build` (node **v24.16.0**) → exit 0. `release.assets.length` **44**,
  `s1.assetCount` **44**, `s1.cachedBytes` **44,693,796**. The scan reports *"0 occurrence(s) in 579
  tracked source files ... (floor 400). This bounds the SCANNED POPULATION only; it is not a claim
  that no request leaves the machine."*
- `[folio-designer]` `npm run verify:offline:red` → exit 0; `S1 CJK row is not the dominant font
  payload` did **not** fire and `over the declared maximum of` did **not** fire.
  `npm run verify:offline:wasm` → exit 0.

**Per-asset Brotli, and the ONE number Story 8.4d inherits (AC5, D-8.4j.8).**
`brotli.immutableAssetCount` 43, `brotli.totalBytes` 15,719,224.
**The catalogue's total added Brotli weight is 2,227,609 bytes (2.12 MiB) across its 21 faces** —
14.2% of the immutable payload. Command `npm run build`, reading `brotli.catalogue.totalBytes` from
`dist/offline-release-manifest.json`; working directory `/Users/panitw/Projects/folio/folio-designer`;
commit `9e2792d`; tree clean at measurement; node v24.16.0. **No threshold was set or moved**;
`epics.md`'s superseded `~9 MB` figure is untouched.

**Manual checks.**
- `lint/MANIFEST.md` asset table read by eye: every catalogue row carries a real SPDX id and a real
  copyright line. **Zero `SEE NOTICE` rows.** 19 rows `OFL-1.1`, 2 rows `Ubuntu-font-1.0`.
- AC1 re-verified independently across all 21 faces: `shasum -a 256` of each committed binary equals
  the digest its own `NOTICE.md` records, the byte size matches, and a `LICENSE*` and a `Copyright`
  line are present. **21/21, zero problems.**
- AC6: `src/generated/runtime-fonts.css` carries **27** `@font-face` rules, **27** `format('truetype')`,
  and **zero** `font-weight`/`font-style` descriptors.
- Guards mutation-proved **by deletion**, each on its own message: comment-stripping → the
  comment-direction test reds alone (1 failed / 8 passed); population floor → the vacuity test reds
  alone (1 failed / 8 passed); the per-asset Brotli assertion → `red proof brotli-record-drift
  escaped verification`, exit 1; the mutable-entry branch → `red proof
  brotli-record-on-mutable-entry escaped verification`; the shipped-family collision guard → an entry
  declaring `Noto Sans` throws (before the fix it did not); the nameID-13 licence check → flipping
  `ubuntusans` to `OFL-1.1` reds that face by name.
- `git diff` touches **no** `folio-go/fonts/`, **no** `fixtures/**`, **no** `.folio`, **no**
  `_bmad-output/planning-artifacts/`, and **no licence-gate source** (`classify.go`, `manifest.go`,
  the allowlists and the classifier tables are untouched; the lint-side diff is `MANIFEST.md`, two
  `_test.go` files and testdata).

**Matrix Test Audit.** Seven of the eight I/O matrix rows had a covering test that ran and passed.
One did not: *`WITH` form → build fails closed* had **no covering test anywhere in the repository** —
`TestResolveAssetsRefusesAnUnenumerableSPDXExpressionAsUnclassifiable` covers `(MIT)`, `MIT()` and
`MIT XOR Apache-2.0`, and the only `WITH` in any test file was prose in a comment. That row is
load-bearing: it is the sole ground for excluding Linux Libertine, and for excluding it on **parser
scope, pending a parser widening**, rather than on licence policy. Measured first — the gate refuses
the Linux Libertine expression on the *unclassifiable* arm, not the copyleft arm — then pinned by
`TestResolveAssetsRefusesAWITHExpressionAsUnclassifiableNotAsCopyleft`, which asserts the **arm**
rather than merely the failure and carries a second case with no copyleft term at all. Proved
falsifiable both ways. **No gate behaviour was added or changed**; the test records what
`classify.go` already does, and no licence test's verdict moved.

**Residual risks.**
1. **AC3 clears by a margin of one** — 21 new families against ≥20. Losing any single face for any
   reason puts the catalogue under the bar; the reserve list and the `.otf` route in
   `8-5-catalogue-procurement.md` are the named remedies, and neither was needed here.
2. **Thai coverage is adequate but not diverse** — it comes from two vendors only (Noto and IBM Plex),
   because the Cadson Demak families are tagless and therefore procurement-blocked. Not a licence problem.
3. **The cache-asset ceiling is now materially consumed**: 44 of 64 slots used, 20 free. Story 8.6 and
   any further catalogue growth spend them.
4. **The Brotli subtotal has no owner until Story 8.4d lands.** The catalogue adds ~2.1 MiB compressed
   to a precached first load and nothing in the repository compares it to a threshold — deliberately
   (8.4d owns the budget), but the number can grow with every gate green until 8.4d sets one.
5. **The scan's claim is bounded and must stay bounded (D-8.5.5).** It proves no forbidden host appears
   in the 579-file scanned population — never that no request leaves the machine. Four scan-scope
   limitations are recorded in `deferred`.

**Licence-gate findings registered rather than fixed (D-000.11).** This story put 21 `LICENSE*` files
through the gate for the first time and the gate handled all of them correctly: 21/21 classified,
zero `SEE NOTICE`, zero build failures, no allowlist pressure. **No licence-gate defect was found, so
none was registered to Epic 15.** One licence-gate *coverage* gap was found and closed additively
without touching gate code — the `WITH` row above. The pre-existing `WITH` parser limitation is not a
new finding: it is already registered as D-8.5.16 / Epic 15 and is the reason Linux Libertine is out.

**Resolved by judgement rather than a recorded ruling.**
1. **Adding `TestResolveAssetsRefusesAWITHExpressionAsUnclassifiableNotAsCopyleft`.** The workflow's
   Matrix Test Audit mandates a covering test per matrix row and halts otherwise; D-000.11 prohibits
   licence-gate *changes*. I read a test that records existing behaviour as outside that prohibition —
   the spec itself tasks this story with editing `licencecensus_test.go`, a licence-side test file —
   and chose it over halting the story the run was course-corrected to reach. **No gate source was
   touched and no licence test's verdict moved.** Reversible by deleting one test function.
2. **Correcting the scanner's recorded population figure.** The comment claimed 1,058 files; the scan
   reads 578 (579 after `.mts`/`.cts`), and the repository-wide extension count is 704, so the figure
   described nothing. Corrected to the measurement with the correction stated rather than silently
   overwritten. The floor of 400 was never wrong.
3. **`baseline_revision` was advanced** from the spec's planning value `19959fa` to the dispatch HEAD
   `1a4ccea`, per step-03. Safe because the two code trees are byte-identical (verified above); the
   spec's `## Verification` note anticipated exactly this and asked for HEAD to be re-measured at the
   build gate rather than trusted.
4. **The six hardcoded asset slots were kept** rather than folded into the catalogue loop. Three
   separate consumers name them (`runtimeAssetUrls` imports, the release generator's URL-substring
   finds, and the family→file identity join), so the loop covers the catalogue only.


### Dispatch 2 — 2026-09-02, re-plan (halt after planning)

Status: ready-for-dev
Blocking condition: none
Baseline: `19959fa9d9188e9732e658c4e26f825f44ecdd24` (HEAD moved to `20002e4` mid-dispatch;
`_bmad-output/` only, code tree identical).

**No code written. No commit made.** Working tree carries this spec (modified) and
`8-5-catalogue-procurement.md` (new) — nothing else.

**Warnings:** `multiple-goals` cleared via the 8.4h split. **`oversized` remains** — 31,230 chars
(~7,800 est. tokens) against the flag's 1,600-token threshold. Reduced 27% from the superseded
version by the named lever (procurement table moved to the companion); no AC thinned. Every spec in
Epic 8 trips this flag (8.4f ~25,900 tokens), so it marks a convention gap, not a defect here.

**Domain measured before planning (D-000.12):** 21 new `.ttf` families verified procurable under the
four-id allowlist with no `WITH` form — licence read from each project's own text, tag pinned,
in-archive path confirmed against a real archive listing. **AC3 clears by a margin of one**; the
reserve and the `.otf` route exist for procurement failures.

**Resolved by judgement, not by a recorded ruling** — flagged for the plan gate:
1. The `<intent-contract>` was rewritten rather than preserved verbatim (see Spec Change Log).
2. The compound-expression census case is seeded as a fixture, because no procurable font carries an
   admissible compound expression (Design Note 3). Bounded by its own Block If.
3. AC3 read as **≥20 *new*** families (the stronger reading), which satisfies the literal text under
   either interpretation.

### Dispatch 3 — 2026-09-02, implementation

Status: done
Blocking condition: none
Baseline: `1a4cceaa81f65cd9899eb13efd2ef207d2394c3f` (HEAD at dispatch; the code tree is
identical to `19959fa` — every commit since touches `_bmad-output/` only, re-measured at the build
gate rather than trusted from the plan gate's line).

**All 21 Tier A faces procured; no reserve draw, no `.otf` route, no derivation.** AC3 clears at 21
new families against a bar of 20 — the margin of one the plan gate predicted, unchanged. Every face
is a single upright static Regular by its own `OS/2`/`head`/`post` tables, with no `fvar`, no `gvar`
and no `CFF` (AC6, asserted per face in `folio-designer/src/font-catalogue.test.ts`).

**Two Ubuntu-font-1.0 assets land**, closing the gap `classify.go:167-171` records in as many words.
`lint/MANIFEST.md` now carries `Ubuntu-font-1.0` on a real committed binary rather than on a fixture.

**THE ONE NUMBER (AC5, D-8.4j.8).** The catalogue adds **2,227,609 Brotli bytes** (2.12 MiB) to the
offline release, across 21 assets.

| item | value |
|---|---|
| Command | `cd folio-designer && npm run build`, then `node -e` over `dist/offline-release-manifest.json` reading `brotli.catalogue.totalBytes` |
| Commit | `1a4cceaa81f65cd9899eb13efd2ef207d2394c3f` plus this story's uncommitted working tree |
| Tree state | dirty — the catalogue and its guards staged, nothing else |
| Working directory | `/Users/panitw/Projects/folio/folio-designer` (the MAIN CHECKOUT, never a linked worktree — DW-105) |
| Node | v24.16.0 |

The whole immutable set is **15,719,224 Brotli bytes** across 43 assets, so the catalogue is **14.2%**
of the compressed payload; the engine wasm remains the dominant term. **No threshold was set or
moved** — Story 8.4d owns that and sets it last, against finished weight.

**Measured against the bounds.** `release.assets.length` **44** (was 23), `s1.assetCount` **44**,
`s1.cachedBytes` **44,693,796** (was 38,460,833). `maximumCacheAssets` stays **64**; 44 is 20 under
it. The four S1 rows are byte-for-byte what the Code Map recorded — engine 7,226,258 / latin 226,026
/ thai 24,872 / cjk 4,948,312 — so `S1 CJK row is not the dominant font payload` cannot fire and did
not. All **23** golden digests are byte-identical to the pre-dispatch snapshot.

**The generator change is a manifest-driven loop, and the six named slots stay.** `font-catalogue.json`
is the single declaration; `build-wasm.mjs` loops it. The six hardcoded slots were deliberately NOT
folded in: they are a vocabulary, not a list — application code imports them by name out of
`runtimeAssetUrls`, `generate-offline-release.mjs` finds three of them by URL substring, and
`font-binary-identity.test.ts` joins them family by family to a source file. Twenty-seven keys is the
shape Design Note 4 refuses; six names plus one loop is what it asks for.

**AC4's scan found something, and the population is a decision on the record.** The forbidden hosts
appear **zero** times in the product source trees (`folio-designer`, `folio-go`, `lint`, `hashmatrix`,
`tools`, `.github` — 573 tracked files). They appear **18 times outside them**: 15 in `_bmad-output/`
(archived UX mockups and the story artifacts that quote the hosts in order to forbid them) and **3 in
`docs/expression-reference.html`, which really does link a Google Fonts stylesheet**. That third one
is a PRE-EXISTING fact about a documentation page, predates this story, and is **out of scope here**;
it is reported rather than swept into the scan's exclusion list in silence. Per D-8.5.5 the claim is
bounded to the scanned population and is never "no request leaves the machine".

**Three guards mutation-proved by deletion**, each reddening on its own message and nothing else's:

| mutation | result |
|---|---|
| `exemptLineNumbers` reads the raw text instead of the comment-blanked text | 1 failed / 8 passed — *"a declaration written in a comment declares nothing — comments are stripped FROM THE EXEMPTION: expected Set{ 1 } to deeply equal Set{}"* |
| the population floor check is deleted | 1 failed / 8 passed — *"refuses to report a clean population over an implausibly small one: expected [Function] to throw an error"* |
| the per-asset Brotli assertions are deleted | `offline release verification failed: red proof brotli-record-drift escaped verification` |
| one NOTICE's recorded SHIPPED digest is altered by one nibble | 1 failed / 3 passed, naming `inter` and both digests |

**Verification.** `lint` **all green** (`TestLicenceSignalCensus`: *"CENSUS: 57 licence texts measured
(48 committed files + 9 dependency licences), all matching their pinned verdicts"*). `folio-go`
1815 pass / 2 fail / 5 skip — standing red 1 exactly. `gofmt -l folio-go lint` prints
`lint/internal/rules/licencegraph_test.go` and nothing else — standing red 2 exactly.
`npm run lint` exits 0 with exactly 4 `only-export-components` warnings — standing red 3 exactly.
`npm test` **42 files / 424 tests, zero failures** (baseline 40/411). `npm run typecheck`,
`npm run test:e2e:compile`, `npm run build`, `verify:offline:red` and `verify:offline:wasm` all exit 0.
**No fourth red.**

**Resolved by judgement, flagged for review:**
1. **The six named slots were kept** rather than folded into the loop (reasoning above). The Execution
   line reads "replace the 6-key `assets` literal with a manifest-driven loop over the catalogue
   directories"; what shipped is the loop over the catalogue, with the six kept because three separate
   consumers name them.
2. **`canvas-font-stack.test.ts`'s well-formed-rule parse was widened** to admit the catalogue's
   `${face.filename}` interpolation beside `${assets.<slot>}`, and its rule-spelling count moved
   6 → 7. The guard's actual property — no `@font-face` may carry a LITERAL `src`, i.e. an arbitrary
   URL outside the offline asset graph — is unchanged and gained a second discrimination case
   (a catalogue-shaped rule pointed at a font service is still refused).
3. **The scan's population is the product source trees, not the repository**, for the reason recorded
   above. Stated in the scanner's own source with the measured counts of what is excluded.
4. **The compound census fixture was wired into the permissive fixture graph** (go.mod, require,
   replace) rather than committed as a bare LICENSE, following `ufl-lib`'s precedent. Both its terms
   are permissive, so `TestLicenceGraphFixtureScan`'s permissive subtest still expects zero findings —
   **no licence test other than the census changed verdict**, which was Design Note 3's Block If.

**No licence-gate change of any kind.** `classify.go`, `licencesignals.go`, `manifest.go`, the
allowlists and the classifier tables are untouched; `git diff` over `lint/internal/` shows only
`licencecensus_test.go`. No licence-gate defect was found to register.

## Delivery Log

### 2026-09-02 — planned

Planned twice. The first plan gate split the story: `multiple-goals` was a real signal, the licence
gate was live-broken at HEAD, and AC2/AC7 left for Story 8.4h (D-8.5.13). The 8.4h line then ran to
8.4i and 8.4j, and the owner stopped it (**D-000.11**) — four stories of licence-checker work, zero
lines advancing the epic's stated goal. The story was **re-planned to six ACs at `19959fa`** under a
flat prohibition on any licence-gate change. `oversized` did not clear and was **left set** rather
than cleared by thinning acceptance criteria; the named lever moved the twenty-one-row procurement
table into `8-5-catalogue-procurement.md`, which is normative for procurement.

### 2026-09-02 — built

Dispatched at baseline `1a4ccea` (code tree byte-identical to the plan baseline `19959fa`).
Three commits: **`048f662`** — the 21 faces, `font-catalogue.json`, the manifest-driven generator,
the per-asset Brotli record, the forbidden-host scan and the census pins; **`9e2792d`** — the
`WITH`-clause refusal pinned to its arm, and the scanner's recorded population corrected to what it
actually reads; **`efa32f7`** — the review's seven patches.

Triage: **7 patched / 4 deferred / 8 rejected**, 0 intent gaps, 0 bad-spec loopbacks.
**Patched (5 medium, 2 low):** the collision guard's shipped-family half seeded from slot keys rather
than family names, so it could never fire; `font-catalogue.json`'s `licence` field was dead data;
no catalogue face was checked for the licence in its own bytes; `SCANNED_EXTENSIONS` omitted `.mts`
and this change added the repo's only `.mts` file; `directory`/`file`/`family` unvalidated where they
become a path or CSS syntax; the scan's reach assertion covered three of six roots; the mutable-entry
Brotli branch was never red-proved.
**Deferred (2 medium, 2 low)** — now **DW-134** (scan reads only tracked files), **DW-135** (only the
two Google hosts forbidden), **DW-136** (`blankComments` approximates comment syntax), **DW-137**
(`npm run build` needs a git checkout), all owned by the **engineering lead**.
**Rejected (2 medium, 6 low)**, each enumerated with what it verified: (1) the 44 ≤ 64 count is
enforced, not merely unwritten; (2) a failing threshold placeholder is 8.4d's job, not this story's;
(3) warn-and-skip outside a checkout inverts the guard — the observation was deferred as DW-137
instead; (4) "~16 families under a distinct-design reading" re-litigates a settled count; (5) the
"21 catalogue faces" sentence in each NOTICE is a historical record, parsed by nothing; (6) cousine's
"Copyright 2026" is what the binary says; (7) geist/geistmono record identical archive URLs; (8)
symlink/truncated-sfnt/name-format-1 hardening guards inputs that cannot reach the site.

One boundary call: the matrix audit found the `WITH` row uncovered and the build added **one test**
recording existing gate behaviour. Kept under **D-8.5.18**, and named there as a narrowing reading of
a flat prohibition — *if a second such reading appears, the answer is no.*

### 2026-09-02 — done

Baseline `1a4ccea`. Closed at `main`, nothing pushed (`origin/main` still `c985b9c`).

**Gates re-measured at HEAD, not relayed** — every invocation with its working directory (D-8.4j.8).
`[folio-go]` `go test -count=1 ./...` → **1815 pass / 2 fail / 5 skip**, the two failures
`TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)` by identity; `go vet ./...` exit 0, silent.
`[repo root]` `gofmt -l folio-go lint` → `lint/internal/rules/licencegraph_test.go` and nothing else.
`[lint]` `go test -count=1 ./...` → all four packages ok, 0 FAIL; `TestLicenceSignalCensus` prints
*"CENSUS: 57 licence texts measured (48 committed files + 9 dependency licences), all matching their
pinned verdicts"* and pins the compound case `MIT OR Apache-2.0`. `[lint]` `go run ./cmd/genmanifest`
exit 0, then `[repo root]` `git diff --exit-code -- lint/MANIFEST.md` exit 0 — regeneration is a no-op.
`[folio-designer]` (node v24.16.0) `npm run typecheck` exit 0; `npm run lint` exit 0 with **exactly 4**
`only-export-components`; `npm test` **42 files / 424 tests, 0 failures** (baseline 40/411);
`npm run test:e2e:compile` exit 0 — compile only, DW-101; `npm run build` exit 0 with
`release.assets.length` **44**, `s1.assetCount` **44**, `s1.cachedBytes` **44,693,796**, and the scan
reporting *0 occurrences in 579 tracked source files (floor 400)* — 579 re-derived independently from
`git ls-files` over the six roots. `npm run verify:offline` exit 0, `verify:offline:red` exit 0 with
neither `S1 CJK row is not the dominant font payload` nor `over the declared maximum of` firing,
`verify:offline:wasm` exit 0. `[repo root]` `grep -c 'maximumCacheAssets = 64'` → `1`. Golden digests
diffed against a **worktree at `1a4ccea`**: **23/23 byte-identical**, empty diff.

**THE ONE NUMBER (AC5).** The catalogue adds **2,227,609 Brotli bytes** — 2.12 MiB — across its 21
faces, against an immutable payload of **15,719,224** bytes across 43 assets, i.e. **14.17%**.
Command `npm run build` then `node -e` over `dist/offline-release-manifest.json`; working directory
`/Users/panitw/Projects/folio/folio-designer` (main checkout, never a linked worktree — DW-105);
commit `c63fb29`; tree clean at measurement; node v24.16.0. Derived **three independent ways** at the
close and equal in all three: the recorded `brotli.catalogue.totalBytes`; the sum of the 21 assets'
own `brotliBytes`; and `stat` over the 21 emitted `.br` sidecars on disk. **This is the number Story
8.4d inherits. No threshold was set or moved** — `epics.md`'s `~9 MB` is untouched by all three
commits, and the only threshold-shaped lines the diff adds are comments disclaiming one.

**Teeth, proved at the close rather than accepted.** Deleting a catalogue face from the built bundle
makes `verify:offline` **fail, exit 1** — the offline claim is not decoration. Two directions name the
face by URL (`missing Brotli sidecar /assets/catalogue-literata…`, and a swapped binary reds
`/assets/catalogue-literata…`); the *whole face removed* direction reds on the earlier generic
`manifest and production runtime output are not an exact set`, because `sameSet` shadows the
`missing manifest asset <url>` guard that would have named it. Loud in every direction, specific in
two of three. Separately, declaring an unrecognised SPDX id for a face **fails rather than skips** —
mutating one face to `MIT` reds on *"which no entry in licenceSignatures recognises"*, and nameID 13
was read out of two binaries by hand and matched (`Ubuntu Font Licence 1.0` / `SIL Open Font
License`). Procurement sampled on **3 faces drawn at random** (notoserifthai, ubuntusansmono,
literata): committed sha256 equals the digest its own NOTICE records, byte size matches, `LICENSE*`
and a `Copyright` line present — 3/3.

**`followup_review_recommended: true` discharged per D-8.5.19**, without a second review dispatch:
zero high, zero intent gaps, zero bad-spec loopbacks, all 8 rejections enumerated with what each
verified, all 7 patches mutation-proved by deletion. The closer carried the scrutiny — rejections
(1), (4), (6) and (7) were re-checked against the bytes and the built manifest and all four hold.
**No licence-gate defect was found, so none was registered to Epic 15** (D-8.5.20): 21 `LICENSE*`
files went through the gate for the first time and it handled all 21 — zero `SEE NOTICE` rows, zero
build failures, no allowlist pressure.

**Heavy tests: written and compiling, NOT RUN.** Under the per-epic cadence (D-000.4) the four
`FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity` and the Playwright suite were **not run
here**; `npm run test:e2e:compile` proves the e2e specs still typecheck, and nothing engine-side, no
`.folio` field and no golden moved. **They come due at Epic 8's boundary gate**, which owns them.

**Residual risks, carried forward.** AC3 clears **by one** — 21 new families against a bar of 20, so
losing any single face puts the catalogue under it; Thai coverage comes from **two vendors** only
(Noto and IBM Plex), the Cadson Demak families being tagless and therefore procurement-blocked, not
licence-blocked; **44 of 64** cache slots are now used, and Story 8.6 plus any further catalogue
growth spend the remaining 20; and the forbidden-host scan's claim is bounded to its **579-file
population** — never *"no request leaves the machine"* (D-8.5.5). One pre-existing fact reported
rather than swept up: `docs/expression-reference.html` really does link a Google Fonts stylesheet.
It sits outside the scanned roots, predates this story, and is out of scope here.

**Epic scope check (D-000.11).** This is the **first story of the run to ship something a user would
notice** — 21 typefaces, in the bundle, reachable offline. Epic 8 stays `in-progress`: 8.6 and 8.4d
remain, and 8.4k is `deferred-to-epic-15`.
