---
title: 'Story 8.5: A curated catalogue ships with the designer'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
baseline_revision: '19959fa9d9188e9732e658c4e26f825f44ecdd24'
review_loop_iteration: 0
followup_review_recommended: false
context: ['{project-root}/_bmad-output/implementation-artifacts/8-5-catalogue-procurement.md']
warnings: ['oversized']
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative — owner summary; the intent contract below governs implementation.*

Today an author can only use the handful of typefaces the designer was built with. Choosing anything
else means finding a font, working out whether its licence permits redistributing it inside the
documents you produce, and getting the file onto the machine — a research errand, not a design choice.

The work that made the build refuse an unacceptable licence is finished and shipped. What is left is
the part you can actually see: this story puts the typefaces themselves into the designer, so choosing
a font becomes a search through a list somebody has already cleared. Every face travels with its own
licence text and copyright line.

It deliberately does not add bold or italic — those belong to later work that has not been decided, so
every face here is a single upright weight. And picking a family still does nothing to the document;
that is the next story's job.

A family or two are left out, and the reason is narrower than it looks: the checker cannot yet read
their particular licence *wording*. Not that the licence is unacceptable — the wording is fine and the
font is genuinely free to use. It is a known gap with a known price, and widening the checker later
brings them straight back.

The release gets noticeably heavier, and that is expected. The size limit is set deliberately at the
end of this run against the finished weight, never nudged along the way to whatever the build happens
to weigh.

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
