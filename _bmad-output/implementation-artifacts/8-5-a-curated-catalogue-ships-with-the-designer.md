---
title: 'Story 8.5: A curated catalogue ships with the designer'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
baseline_revision: 'b0fae960cc0ace9f1a243aee0170e99e67ef2fb1'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized', 'multiple-goals']
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative — owner summary; the intent contract below governs implementation.*

Today an author can only use the handful of typefaces the designer was built with. Choosing anything
else means finding a font, working out whether its licence permits redistributing it inside the
documents you produce, and getting the file onto the machine — a research errand, not a design choice.

This story ships a curated list of freely-licensed typefaces with the designer itself, so choosing a
font becomes a search through a list somebody has already cleared. Every face travels with its own
licence text and copyright line, and the build now refuses to finish if any face carries a licence
outside a short permitted set — or one the tool cannot identify at all. Not knowing stops counting
as fine.

It deliberately does not add bold or italic. Those belong to later work that has not been decided,
so every face here is a single upright weight; a list that offers no way to embolden is intended,
not an oversight.

Done looks like this: at least twenty families ship inside the offline release, verified like every
other file the designer carries and reachable with no network; the build fails on an unacceptable or
unreadable licence; and the weight each face adds is measured, not estimated.

Two things will look wrong and are not. Picking a family still does nothing to the document — that
is the next story's job. And the release gets noticeably heavier, which is expected: the size limit
is set deliberately at the end of this run against the finished weight, never adjusted along the way
to whatever the build happens to weigh.

<intent-contract>

## Intent

**Problem:** The designer offers only the six faces it was built with, and nothing in the repository
checks whether a redistributed font's licence is *acceptable* — the asset gate checks only that a
`LICENSE*` and `NOTICE*` sit beside the binary, then records the classification as a label with
`"SEE NOTICE"` as fallback and returns nil. A GPL font would get a manifest row and a clean build.
There is also no scan anywhere for a forbidden font host, and per-asset compressed weight is
recorded nowhere.

**Approach:** Ship a curated catalogue of at least twenty permissively-licensed static faces inside
the designer's offline release, behind the same verified asset URLs as every other release asset;
make an unacceptable *or unclassifiable* asset licence fail the build the way the dependency ban
already does; widen the extension-class guard's population to match the checker it protects; add a
comment-stripping forbidden-host source scan; and record per-asset Brotli bytes for Story 8.4d to
consume. The catalogue is data plus binaries plus gates — no picking behaviour (Story 8.6).

## Boundaries & Constraints

**Always:**
- **Bundled and precached, not fetched.** `spec-fonts/SPEC.md` `## Non-goals` — *"No live font
  service. No Google Fonts API, no arbitrary URL, no 'download on first use'"* — plus
  `font-catalogue.md` §"Why bundled rather than fetched" (*"Nothing is fetched. The bytes come from
  the bundle already on the machine"*) and the story's own AC (*"inside the bundle behind the same
  verified asset URLs … and the offline verification job covers them"*). See Design Note 1 for why
  the fetch-at-pick steer (D-8.5.1b) is not taken, why a lazily-cached second tier is foreclosed by
  the same clause, and the residual tension the plan gate should rule on.
- **Provenance per procurement route (D-8.5.4).** Reproduction (replayable derivation, committed
  output, both digests) only for a face **this project derives**; **provenance** (pinned upstream
  version + NOTICE recording upstream digest, committed digest, byte size, fetch date, path inside
  the archive) is sufficient for a **vendored static** face — the standard Story 8.4c already
  shipped IBM Plex on. Each face states which route it took. Never write an AC claiming a
  derivation that does not exist.
- **Licence: allowlist, fail-closed.** `OFL-1.1`, `Apache-2.0`, `MIT`, `UFL` only. Anything else,
  **and any licence text the classifier cannot identify**, fails the build (D-8.5.2, D-8.5.3).
  `"SEE NOTICE"` stops being a pass. `UFL` does not exist in `classify.go` today and must be added
  with a marker branch, an SPDX entry and a fixture — never by widening a list silently.
- **Byte identity.** No `.folio` format change, no engine change, no `folio-go/fonts/` change. The
  23 golden digests must be **byte-identical** to the pre-dispatch snapshot; a moved golden is a
  HALT, not a re-record. `maximumCacheAssets` stays **64** (measured headroom: 23 assets in use, 41
  free — a 20–40 family catalogue fits without raising it; see Design Note 2).
- **Guards are widened or replaced, never weakened**, and every guard added is mutation-proved:
  delete it and a **named** test or red proof reddens on **its own message**.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- The catalogue cannot reach ≥20 families without a face whose licence is outside the allowlist, or
  whose licence text `ClassifyLicenceText` cannot identify. (Do not widen the allowlist to fit the
  population — that is the defect this story exists to close.)
- Any face requires **derivation** (a new `UPSTREAM` entry) — the bootstrap gap is real
  (gitignored `.font-sources/`, nothing fetches upstream, `out_sha256` unknowable before the first
  run) and pulling it in-scope is a materially different story. Prefer vendored static faces.
- `maximumCacheAssets` would have to be raised, or a golden digest moves, or
  `S1 CJK row is not the dominant font payload` fires.
- A catalogue face's row in `lint/MANIFEST.md` reads `SEE NOTICE`.
- Widening the extension-class guard to the repo-wide population cannot be made green without
  excluding something `ResolveAssets` itself includes.

**Never:**
- Bold, italic, oblique, synthetic emboldening, variable-font axes, or any weighted/sloped face.
  **Epic 11 (FR57) owns realize-vs-retire in full and the owner ruling has not been made (D-000.7;
  `epics.md` Epic 8 header, amended 2026-09-02).** A weighted catalogue face is a collision with
  Epic 11, not a convenience.
- CJK families in the catalogue (a full SC face is 10.6 MB against 646 KB / 47 KB); the shipped SC
  face remains the coverage fallback.
- **A live font service, an arbitrary URL, or a "download on first use"** — `SPEC.md` `## Non-goals`.
- **Enumerating or reading host-installed fonts** on the authoring or rendering machine (`## Non-goals`,
  AD-8: fonts arrive as an explicit value and are never queried from the host).
- Save-time subsetting, or any change to how the PDF producer subsets (`## Non-goals`).
- A container format — the catalogue is committed files plus a data record, never a zip-of-folders.
  *(These five, with bold/italic/axes and CJK above, are the six constraints DW-108 dropped from the
  regenerated epic context; carried here sourced from `spec-fonts/SPEC.md` `## Non-goals` per the
  lead's plan-gate discharge.)*
- Picking behaviour, chain proposal, embedding on pick, or any family-control interaction — Story 8.6.
- A byte threshold or first-load budget gate — **Story 8.4d owns it**, set last against finished weight.
- Editing `epics.md`, `ARCHITECTURE-SPINE.md`, or `<intent-contract>` to match what was built.
- `.woff2`/`.woff`/`.ttc` catalogue faces: the engine decodes only `font/ttf` and `font/otf`, the
  licence gate is blind to those extensions, and the generator hardcodes `format('truetype')`.
- Claiming "no request leaves the machine". A source scan proves **no forbidden host appears in the
  scanned population** and nothing more (D-8.5.5). The literal offline proof stays at the epic gate.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Permissive catalogue face | Face dir with `LICENSE-OFL.txt` classifying `OFL-1.1` + `NOTICE.md` with a `Copyright` line | Manifest row labelled `OFL-1.1`; build green | No error expected |
| Copyleft asset licence | Face dir whose LICENSE text classifies GPL/AGPL/LGPL/SSPL | `ResolveAssets` returns a located error naming the directory and the classified id | Build fails, never warns |
| Unclassifiable asset licence | LICENSE text no marker matches (incl. a commercial EULA) | Located error: licence unresolvable, naming the directory | Build fails; `SEE NOTICE` never emitted |
| Permitted-but-new id | `Apache-2.0` / `MIT` / `UFL` font licence text | Classified to its SPDX id and accepted | No error expected |
| Font invisible to the gate | A `.woff2` (or any font-magic file) committed in a *new* directory outside `public/fonts` | Widened extension-class guard reports it by extension **and** by magic bytes | Test red, named |
| Forbidden host in source | `fonts.gstatic.com` in a scanned `.ts`/`.tsx`/`.css`/`.mjs` **string** | Scan reports the file and the pattern | Test red |
| Forbidden host in a comment | Same host inside `//` or `/* */` | **Not** reported — comments are stripped before matching | No error expected |
| Non-immutable asset, Brotli | `/index.html` (the only mutable asset; no `.br` sidecar) | Recorded with an explicit stated treatment, not silently skipped | Explicit, named |

</intent-contract>

## Code Map

**Licence gate (the allowlist lands here)**
- `lint/internal/manifest/manifest.go:298-301` — `licenceLabel := "SEE NOTICE"`, SPDX id substituted
  only when classification succeeds, `Family` discarded with `_`. **This is the site.** Duplicate
  fall-through for the wordlist at `:384-387`. `:110` `fontExtensions = []string{".ttf",".otf",".ttc"}`.
  `:165` `ResolveAssets(repoRoot)` walks the whole repo minus `.git` and `*/testdata/lint/**`, then
  intersects with `git ls-files` per directory (`:243-249`). `:287/:290/:295` are the existing
  located-error shapes to copy.
- `lint/internal/rules/licencegraph.go:28-62` — **the model.** `FamilyCopyleft` → forbidden-licence
  finding; `FamilyUnknown` → `"licence unresolvable — could not classify licence text"` (D-1.3.8);
  commercial EULA falls through to `FamilyUnknown` by design (`:51-55`).
- `lint/internal/licence/classify.go:96-106` `permissiveSPDX` (no `UFL`); `:121-187`
  `ClassifyLicenceText` marker chain (OFL branch `:137-162` requires the "VERSION 1.1" conjunct);
  `:85-95` the governing "loud miss vs rotting list" comment.
- Fixtures to copy: `lint/testdata/licence/{copyleft,permissive,unknown}/`.
- `lint/MANIFEST.md:280-301` asset table (12 rows today); regenerate with
  `cd lint && go run ./cmd/genmanifest`; `TestManifestUpToDate` (`manifest_test.go:41`) byte-compares.
- Red-proof precedent for asset licences: `lint/internal/rules/fontsassets_test.go:283,:322`.

**Extension-class guard (widen its population)**
- `folio-designer/src/font-binary-identity.test.ts:67` `designerFontsDir` — the single-directory
  population; call site `:635`; `:178-182` `licenceGateFontExtensions` parses `manifest.go`;
  `:217-222` `filesUnder`; `:206` `fontMagics`; `:616` the assertion, `:646` its discrimination proof.
  Widening needs: repo-root sweep + `ResolveAssets`' two exclusions + a `git ls-files` intersection
  (otherwise `dist/`, `src/generated/runtime/`, `.font-sources/`, `node_modules` turn it red).
- Exact-cardinality assertions that must **not** break: `:688`/`:705`/`:720`/`:577`/`:909` (six
  generator slots), `:806`/`:814` (three engine faces), `:880` (chrome faces carry OFL 1.1 in
  nameID 13), and `canvas-font-stack.test.ts:312` (`toBe(6)` `@font-face` rules). **Keep catalogue
  faces out of the `assets` map in `build-wasm.mjs` and out of `runtime-fonts.css`** — a catalogue
  face needs no `@font-face` rule; it reaches the browser as a carried face via
  `embedded-face-registry.ts:54` once Story 8.6 embeds it.

**Bundling the catalogue**
- `folio-designer/scripts/build-wasm.mjs:69-77` `fingerprint()`, `:88-98` the hand-written 9-slot
  `assets` map, `:103-104` `offline-assets.ts`, `:105` the `import.meta.glob('./runtime/pdfjs-*/**/*')`
  pattern — **the precedent for a many-file collection that is not a hand-listed slot**, and the
  shape a catalogue directory should follow. `:134` `runtime-fonts.css` (six rules, one line —
  do not restructure into a loop; `canvas-font-stack.test.ts:43-45`'s regex would silently empty).
- `folio-designer/vite.config.ts:32` `publicDir: false` — a file under `public/` is not an asset
  until something imports it. `:16` `assetsInlineLimit: 0`.
- `folio-designer/scripts/generate-offline-release.mjs:13-18` asset set (`immutable` = everything
  except `/index.html`); `:35-39` the `.br` loop over immutable assets only; `:40-49` `find(needle)`
  and the four hardcoded needles (`'.wasm'`, `'/noto-sans.'`, `'/noto-sans-thai.'`, `'/noto-sans-cjk.'`)
  — **a `/assets/noto-*` catalogue filename could shadow these**; `:53` `cacheAssets[]` records
  **uncompressed** bytes; `:79-83` the manifest write (`release.assets[]` carries no byte counts).
- `folio-designer/scripts/verify-offline-release.mjs:35-52` the 8.4f bound guard (`declaredCacheAssetBounds()`,
  message `over the declared maximum of`); `:78-99` the S1 block — **row pinning is READ-ONLY**;
  `:98-99` the CJK-dominance `Math.max` over ids ending `font`; `:107-119` immutable/Brotli sidecar
  determinism; `:222-227` `rewriteRelease`; `:229-240` `redProof(name, mutate, expected)`;
  `:242-307` `runRedProofs` (follow `asset-count-over-bound` `:269-279` for a population mutation).
- `folio-designer/src/release-payload.ts:32-33` `minimumCacheAssets = 10` / `maximumCacheAssets = 64`
  — **the sole authority; do not re-type, do not raise.** `scripts/offline-release-contract.mjs:56-71`
  derives them (exactly one live `^const <name> = <digits>$` match or it throws).

**Forbidden-host scan (new)**
- `folio-designer/src/canvas-authority-contract.test.ts:131-152` `withoutComments` — the character
  scanner to reuse (quotes checked before comment openers, so a `//` inside a URL string survives).
  `:6-16` the three corpora; `:173` `scanned`; `:186-191` non-vacuity floors; `:198-216` red
  direction; `:218-227` the green-on-prose direction using real comments from this repo.
- **Measured population risk:** `docs/expression-reference.html:2-4` carries live
  `fonts.googleapis.com` / `fonts.gstatic.com` `<link>` tags. Zero hits exist in
  `folio-designer/src`, `scripts`, `e2e`, `public`, `folio-go`, `lint`, `tools`. Define the scanned
  population explicitly and say what it excludes — do not silently exclude `docs/` and call it clean.

**`len(UPSTREAM)` de-hardcoding**
- `tools/fontgen/instance_faces.py:117-168` `UPSTREAM` (3 entries), `:366-367` prints
  `derived and compared {compared} of {total}` where `total = len(UPSTREAM)` — **already computed**.
- `folio-go/fontgen_matrix_test.go:117` `const wantWitness = "derived and compared 3 of 3 faces"`
  (`//go:build matrix`) — **the literal.** `:75-77` re-hardcodes the three source filenames and
  digests a second time. `UPSTREAM` is not readable from Go (no JSON/manifest intermediate,
  no `go:generate`) — so derive the count from the Python side's own output, not from a second literal.

**Continuity**
- `_bmad-output/implementation-artifacts/8-4c-…md` (font procurement: npm packages ship woff2 only —
  use GitHub release zips with `fonts/complete/ttf/`; NOTICE schema; a `SEE NOTICE` row is a Block If).
- `8-4f-…md` (the bound and its red proofs), `8-4g-…md` (build determinism; the wasm is still a
  function of the checkout path — DW-105 — so measure baselines in the **main checkout**).

## Tasks & Acceptance

**Execution:**

1. `lint/internal/licence/classify.go` — add `UFL` (SPDX entry + marker branch + `Family` mapping)
   so the allowlist's fourth member exists rather than being assumed. Add fixtures under
   `lint/testdata/licence/` for a UFL text and a non-classifiable font licence.
2. `lint/internal/licence/classify_test.go` — cover UFL positive, UFL-lookalike negative, and the
   unclassifiable case. Red-prove: removing the marker branch reddens a named test.
3. `lint/internal/manifest/manifest.go:298-301` — replace the `"SEE NOTICE"` fall-through with a
   hard, located error: an asset licence outside `{OFL-1.1, Apache-2.0, MIT, UFL}` **or**
   unclassifiable fails, naming the directory and the reason, in the `:287`/`:290`/`:295` voice.
   Apply the same at `:384-387`. Rationale: D-8.5.2/D-8.5.3 — enforce like the dependency ban.
4. `lint/internal/manifest/manifest_test.go` — add red proofs mirroring
   `fontsassets_test.go:283,:322`: a scratch dir with a copyleft LICENSE errors; one with an
   unreadable-by-classifier LICENSE errors; a permitted one does not. Assert **no** row anywhere in
   the generated manifest reads `SEE NOTICE`.
5. `folio-designer/public/font-catalogue/<family-slug>/` — commit ≥20 permissively-licensed static
   single-instance Latin (and where available Thai) faces, each with its unmodified upstream
   `LICENSE*` and a `NOTICE.md` recording upstream project, pinned version/tag, asset URL, path
   inside the archive, fetch date, upstream sha256, committed sha256 and byte size. **Vendored
   static route** per face (D-8.5.4); if any face would need derivation, HALT per Block If.
   A new directory — not `public/fonts` — so the six-slot chrome assertions stay exact.
6. `folio-designer/scripts/build-wasm.mjs` — fingerprint-copy every committed catalogue face into
   `src/generated/runtime/font-catalogue/` (the `fingerprint()` helper at `:69-77`), **iterating the
   directory rather than adding hand-written slots to the `assets` map at `:88-98`** — that map's
   nine entries are asserted by exact cardinality in five places (see Code Map). This step is
   required: `vite.config.ts:32` sets `publicDir: false`, so a file under `public/` is **not** in
   Vite's graph and an `import.meta.glob` over `public/` would emit nothing and silently ship an
   empty catalogue. Emit **no** `@font-face` rule for a catalogue face.
7. `folio-designer/src/catalogue/font-catalogue.ts` (new) — the catalogue record as data:
   `family`, `style` (always `Regular`), `licence` (SPDX id), `source` (upstream + version),
   `scripts` (declared coverage), and the asset URL — built from an
   `import.meta.glob('../generated/runtime/font-catalogue/**/*', { query: '?url' })`, the
   `build-wasm.mjs:105` pdfjs precedent, so a face becomes a release asset by being committed
   rather than by being hand-listed. **No UI consumer in this story.**
8. `folio-designer/src/catalogue/font-catalogue.test.ts` (new) — assert the catalogue has ≥20
   families, every entry's `style` is `Regular`, no entry declares a CJK script, every entry's
   declared family matches the `name` table read from its own bytes (reuse the sfnt reader in
   `font-binary-identity.test.ts:233-288`), and every entry's file is `.ttf`/`.otf` with an sfnt magic.
9. `folio-designer/src/font-binary-identity.test.ts:67,:217-222,:635` — widen the extension-class
   guard's population from `public/fonts` to `ResolveAssets`' own walk (repo root, minus `.git` and
   `*/testdata/lint/**`, intersected with `git ls-files`). Keep `:623`'s exact `['.ttf','.otf','.ttc']`
   mirror and `:646`'s discrimination proof; extend the latter to prove the *new* directory is now visible.
10. `folio-designer/src/forbidden-host.test.ts` (new) — comment-stripping scan for
   `fonts.google`, `fonts.gstatic.com`, `googleapis` over an explicitly declared population, using
   `withoutComments` from `canvas-authority-contract.test.ts:131-152`. Non-vacuity floors on each
   corpus; red-proved **both** directions (a host in a string is caught; the same host in a `//`
   and a `/* */` comment is not). State the excluded population in a comment, naming
   `docs/expression-reference.html` as a known live reference outside the scanned set.
11. `folio-designer/scripts/generate-offline-release.mjs` — record **per-asset Brotli bytes** for
    every asset in the manifest (a new field beside the existing uncompressed `cacheAssets[].bytes`,
    or a sibling collection), with an **explicit stated treatment** for the one non-immutable asset
    (`/index.html`, which has no `.br`) rather than a silent skip. Do **not** add a threshold.
12. `folio-designer/scripts/verify-offline-release.mjs` — assert every recorded Brotli figure equals
    its emitted `.br` sidecar (the `:93` pattern), and add a red proof (`brotli-record-drift`) with
    an `expected` message. Confirm `find(needle)` still resolves the four S1 needles unambiguously
    against the new filenames; if a catalogue filename could shadow one, tighten the needle and
    red-prove the tightening.
13. `folio-go/fontgen_matrix_test.go:117` — derive the expected witness count from the generator's
    own `len(UPSTREAM)` output instead of the literal `"3 of 3"`, so a new face fails on a byte
    divergence or not at all — never on a string (D-8.5.4). Also de-duplicate `:75-77`'s second copy
    of the source digests, or state why it must stay.
14. `lint/MANIFEST.md` — regenerate (`cd lint && go run ./cmd/genmanifest`); a second run must leave
    no diff. Every catalogue row must carry a real SPDX id and a real copyright line.
15. `_bmad-output/specs/spec-fonts/font-catalogue.md` — amend "Which families, and how many, is an
    open question" to record what shipped and the rule that admits a face; amend the Selection
    criteria table's Instance row to state **both** procurement routes (D-8.5.4). Mark the families
    half of `SPEC.md`'s Open Question settled, citing D-8.5.3. Do not touch the bold/italic entry.

**Acceptance Criteria:**

- **AC1 (procurement).** Given each catalogue face, when it is prepared, then it is a static
  single-instance face, never generated at build time, whose assurance is stated **per route** —
  reproduction for a face this project derives, provenance (pinned version + NOTICE with both
  digests and a byte size) for a vendored static face — and its `NOTICE.md` records which route applies.
- **AC2 (licence, fail-closed).** Given a committed font whose licence classifies outside
  `{OFL-1.1, Apache-2.0, MIT, UFL}`, or whose licence text cannot be classified at all, when the
  `lint` suite runs, then it **fails** with a located error naming the directory and the reason,
  and no manifest row anywhere reads `SEE NOTICE`. Red-proved in both directions from a scratch fixture.
- **AC3 (in the bundle, verified).** Given the offline release, when it is built and verified, then
  every catalogue face is present behind a content-addressed, immutable, `.br`-backed asset URL,
  is covered by `verify:offline`, and `release.assets.length` remains at or below the **unchanged**
  declared maximum of 64.
- **AC4 (forbidden-host scan).** Given the designer's declared source population, when it is
  scanned with comments stripped, then no `fonts.google`, `fonts.gstatic.com` or `googleapis`
  reference appears in it — proved red when such a host is placed in a string, and proved green
  when the same host sits in a comment. The criterion claims nothing about requests actually leaving
  the machine.
- **AC5 (weight, recorded not thresholded).** Given the built release, when its weight is recorded,
  then every asset carries a **per-asset Brotli** figure equal to its emitted sidecar, the one
  non-immutable asset's treatment is stated explicitly, and the figures are recorded with the exact
  invocation, commit and tree state — with **no threshold set** (Story 8.4d owns that).
- **AC6 (scope fence).** Given the catalogue, when it is assembled, then no entry is bold, italic,
  oblique, variable or CJK; every entry's `style` is `Regular`; and no family-control or
  pick behaviour is added (Story 8.6).
- **AC7 (gate visibility).** Given a font-magic file committed in **any** tracked directory the
  licence gate walks, when the extension-class guard runs, then it is reported if the gate cannot
  see it — proved by a fixture in a directory outside `folio-designer/public/fonts`.
- **AC8 (no drift elsewhere).** Given the whole change, when the gates run, then the 23 golden PDF
  digests are byte-identical to the pre-dispatch snapshot, `README.md`'s md5 is unchanged, and the
  fontgen witness count is derived rather than a literal.

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


## Review Triage Log

## Design Notes

**1. The delivery fork, resolved — and the one residual the plan gate should rule on.** D-8.5.1(b)
steers toward faces fetched at pick time ("zero build assets, zero S1 rows, zero first-load bytes")
and D-8.5.1 assigns the fork to this story because *"it IS the story's design"*. Resolved as
**bundled and precached**, on four pieces of contract text, three of which are decisive:

- `spec-fonts/SPEC.md` `## Non-goals`: *"**No live font service.** No Google Fonts API, no arbitrary
  URL, no 'download on first use'."* This forecloses **both** fetch-at-pick **and** the softer shape
  of shipping the faces in the release but caching them lazily — "download on first use" is named.
- `font-catalogue.md`: *"Nothing is fetched. The bytes come from the bundle already on the machine."*
- The story's own AC: *"inside the bundle behind the same verified asset URLs as every other release
  asset, and the offline verification job covers them."* Fetch-at-pick satisfies neither clause.
- Measured, there is nowhere to fetch **from**: AC4 bans the Google hosts, the product ships no
  server, and NFR7 promises the designer works offline. A face reachable only over a network is a
  palette that stops working on a plane — the exact failure `font-catalogue.md` was written against.

D-8.5.1 itself calls the bundled route *"legitimate and not a compromise"*, and — measured — it does
not even need the concession that ruling offered: at 23 of 64 assets, no bound is raised.

**The residual, stated rather than smoothed over.** `SPEC.md`'s newly-settled catalogue entry closes
with the owner's brief as the lead recorded it: *"the catalogue should be big and the engineering
should find a shape where size is not paid for at first load."* Under the Non-goals clause above,
**no such shape exists** for a catalogue that must also be available offline: precached costs
first-load bytes, and anything cheaper is a download on first use. This spec pays the bytes and
makes them **visible** (AC5's per-asset Brotli record is the instrument), rather than inventing a
tier the Non-goals forbid or quietly reporting a zero. That is a real, priced trade — the owner was
given the ceiling and the 37% overage when they chose 20+ families — but it is the one clause of the
brief this story cannot honour, so it is surfaced here for the gate rather than buried. **What would
reverse it:** a ruling that amends `SPEC.md` `## Non-goals` and `font-catalogue.md` together to
permit a deferred tier, with the offline consequence stated. Do not reverse it in the build.

**2. The ceiling is not the binding constraint; weight is.** Measured at `1f8e52b`:
`release.assets.length` = **23** against `maximumCacheAssets` = **64**, so **41 slots are free** — a
20–40 face catalogue fits with the bound untouched, which is what keeps 8.4f's *"confirm
`maximumCacheAssets` is still 64"* satisfiable. What the catalogue does cost is first-load Brotli
weight: the release currently emits **13,490,909** Brotli bytes across 22 sidecars, and ~20 Latin
faces at 45–226 KB Brotli each adds roughly 1.5–4.5 MB. That is a real, owner-priced cost
(D-8.5.3 put the 37% overage in the question) and **this story must not respond to it by setting or
moving a threshold** — 8.4d sets the figure last, against finished weight (D-8.4.24).

**3. Why the weight cannot be recorded where it looks like it belongs.** `s1.cacheAssets[].bytes` is
**uncompressed** (`generate-offline-release.mjs:53` stats the raw file), and `s1VisibleBytes` sums
four hardcoded filename needles and already misses **174,949** Brotli bytes of IBM Plex — so
recording the catalogue against either would produce a number reading *"this cost nothing"*. Brotli
bytes exist today only in `s1.rows[0..3].bytes` and only for four assets. Hence AC5's new per-asset
record. The `.br` loop filters `asset.immutable`, and `/index.html` is the sole non-immutable
asset — its treatment is stated, never skipped silently (D-8.5.6).

**4. Catalogue faces are deliberately not generator slots.** `build-wasm.mjs`'s `assets` map has
exactly nine hand-written slots, and five separate tests assert **six** font slots / six
`@font-face` rules / six declared families by exact equality. A catalogue face needs no
`@font-face` rule at all — it reaches the browser only after Story 8.6 embeds it, as a carried face
through `embedded-face-registry.ts`. Routing the catalogue through an `import.meta.glob`
(the pdfjs-cmaps precedent at `build-wasm.mjs:105`) keeps those six-way assertions exactly true
while still putting each face into Vite's graph and therefore into the release manifest.

**5. A source scan proves a weaker claim than "offline", and must say so.** Model:
`canvas-authority-contract.test.ts`'s character scanner, which checks quotes **before** comment
openers so a `//` inside a URL string is not mistaken for a comment. Story 8.4b measured a
chrome-token guard staying green over a live token parked in a CSS comment — comment-stripping is a
demonstrated failure mode, not a hypothesis. Equally: the scan's population is a choice, and
`docs/expression-reference.html` really does pull Google-hosted webfonts today. Declaring the
population and naming that exclusion is honest; quietly excluding it and reporting "clean" is the
overstatement this run keeps catching.

**6. UFL does not exist yet.** `permissiveSPDX` carries MIT, Apache-2.0, OFL-1.1, CC0-1.0, BSD, ISC,
0BSD, Unlicense, MIT-0, BlueOak, CC-BY-4.0 — and **no UFL**, no marker branch, no fixture. The
allowlist's fourth member has to be built before it can be enforced. Per `classify.go:85-95`, add it
loudly (entry + marker + fixture + test), never by widening a list to make a population pass.

## Verification

**Baseline: MEASURE at `b0fae960cc0ace9f1a243aee0170e99e67ef2fb1` in the MAIN CHECKOUT (never a
linked worktree — DW-105: the wasm is a function of the checkout path), clean tree, before the first
edit.** Record every figure with its command; do not carry any figure forward from a prior spec.
HEAD moved twice while this spec was being planned (`1f8e52b` → `f2d108c` → `b0fae96`); both
commits touch `_bmad-output/` only, so the **code tree is identical to `1f8e52b`**, where the
figures quoted in the Code Map and Design Notes were measured. Re-verify them at the gate that
consumes them (D-7.8.4: an anchor written at a plan gate is a claim with an expiry date).

**Cadence.** Per-epic (D-000.4): the four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity`
and the Playwright suite are **not** in this list — nothing engine-side changes, no `.folio` field is
added, and no golden can move. **The D-000.4 exception IS invoked** for the designer's release
integration commands (`npm run build`, `verify:offline{,:red,:wasm}`) and for the `lint` module
suite: this story puts new binaries into the offline bundle and changes the licence gate's verdict,
so its own correctness is bundle-shaped. The golden-digest diff is kept as the cheap byte-identity
witness that nothing leaked engine-side.

**Commands:**
- `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` — from REPO ROOT, **before the
  first edit**; expect **23** lines.
- `cd folio-go && go test -count=1 ./...` — expect the **one standing red only**:
  `TestCorpusMeetsP6ExerciseFloors` + its `P6g_(opaque_names)` subtest printing
  `P6g (opaque names) floor not met: got 7, need >=20` (two FAIL lines, one defect). Its drift twin
  `TestCorpusP6StatsMatchDeclaredBaseline` must stay green. **Never "fix" it.**
- `cd folio-go && go test -count=1 -tags=matrix ./...` — the same red plus the could-not-execute
  `TestShippedFacesReproduceFromUpstream` (`fontgen: fontTools is not importable by this interpreter`).
  **Exactly TWO standing reds by identity. Any third is a real failure.**
- `cd folio-go && FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python go test -count=1 -tags=matrix -run TestShippedFacesReproduceFromUpstream ./...`
  — sweep both ways; expect PASS non-vacuously, printing three `matches the recorded derivation`
  lines and the witness. **If it fails WITH the variable set, HALT.** Never commit the variable.
- `cd folio-go && go vet -tags=matrix ./...` — no output.
- `gofmt -l folio-go` — **from REPO ROOT**; no output. (An `lstat` line is a non-measurement.)
- `cd lint && go test -count=1 ./...` — four `ok`, no FAIL; `-count=1` always (rules walk dirs and
  Go's cache does not track `ReadDir`). **This story's licence gate lives here.** Additionally run
  `-run TestManifestUpToDate -v`, `-run TestResolveAssets -v` and `-run TestClassify -v` explicitly
  and **quote what they print**.
- `cd lint && go run ./cmd/genmanifest && git diff --stat -- lint/MANIFEST.md` — after the
  regeneration is committed, a second run leaves no diff.
- `cd folio-designer && npm run typecheck` — exit 0.
- `cd folio-designer && npm run lint` — exit 0 with **exactly 4** `react(only-export-components)`
  warnings. The count and rule name are the invariant; line numbers drift.
- `cd folio-designer && npm test` — project script only (`npx vitest run` from the repo root picks
  up the Playwright specs and gives a false mass failure). Baseline **40 files / 409 tests**; expect
  a higher count and **zero failures**.
- `cd folio-designer && npm run test:e2e:compile` — exit 0. This is `tsc -p tsconfig.e2e.json
  --noEmit`. **Do not report it as a run** (DW-101: CI executes no Playwright).
- `cd folio-designer && npm run build` — node **v24.16.0** exactly; runs `build:wasm`, `tsc -b`,
  `vite build`, `build:offline`, `verify:offline`. Record `release.assets.length`,
  `s1.assetCount`, `s1.cachedBytes`, and the **per-asset Brotli** figures with this exact command.
- `cd folio-designer && npm run verify:offline:red` — exit 0; all red proofs fire on their own
  messages, including the new `brotli-record-drift`. In particular
  `S1 CJK row is not the dominant font payload` must **NOT** fire, and
  `over the declared maximum of` must not fire.
- `cd folio-designer && npm run verify:offline:wasm` — exit 0.
- `shasum -a 256 fixtures/*/expected.pdf | diff <scratch>/digests.before -` — **empty diff**. A moved
  golden is a **HALT**, not a re-record.
- `md5 -q README.md` — `078d7d80d518d54af2fc04fb270d46b8`, unchanged.
- `grep -c 'maximumCacheAssets = 64' folio-designer/src/release-payload.ts` — expect `1`.

**Manual checks:**
- **Mutation-prove every guard added.** Delete the allowlist branch in `manifest.go` → a *named*
  `lint` test reddens on its own message. Delete the widened population in the extension-class
  guard → its discrimination proof reddens. Delete `withoutComments` from the forbidden-host scan →
  the comment-direction test reddens. Delete the Brotli-record assertion → `brotli-record-drift`
  reports `escaped verification`. A guard that survives its own deletion is decoration.
- **Prove each red proof fails for the right reason** by temporarily dropping its `expected`
  argument and confirming the message it actually carries; beware over-determination (a population
  mutation can trip `sameSet` first — assert *"fails on a message that is not the new guard's"* if so).
- **Read `lint/MANIFEST.md`'s `## Redistributed non-code assets` table by eye:** every catalogue row
  carries a real SPDX id and a real copyright line. **A `SEE NOTICE` label is a Block If.**
- **Verify each committed catalogue face against its own NOTICE:** `shasum -a 256` the binary equals
  the committed digest recorded there, and the byte size matches.
- Confirm the generated `src/generated/runtime-fonts.css` still has **exactly six** `@font-face`
  rules with the same six family names after `npm run build:wasm`.
- `git diff --stat` per commit must touch **no** `folio-go/fonts/`, no `fixtures/**`, no `.folio`,
  and no `_bmad-output/planning-artifacts/`. Explicit paths on every `git add` — never `-A`, never
  `.`; never push; never branch.

## Auto Run Result

Status: ready-for-dev
Blocking condition: none

Dispatch: classic intent, plan-only (`Halt after planning.`). Started at
`1f8e52bdd85235a5f5fb32c86837ed2407ed76b9`; HEAD advanced to
`b0fae960cc0ace9f1a243aee0170e99e67ef2fb1` mid-dispatch via `f2d108c` and `b0fae96`, both of which
touch `_bmad-output/` only — verified with `git diff --name-only 1f8e52b..b0fae96`, which returns no
path outside `_bmad-output/`. Working tree clean of tracked modifications throughout; branch `main`;
no code written, no commit created, nothing pushed, no branch created.

**The prior halt is discharged, in the document that gets read.** A previous 8.5 dispatch halted
`blocked` / `intent gap` on `spec-fonts/SPEC.md`'s open question *"Which families make the shipped
catalogue, how many, and who curates the list as it changes?"* That entry is now struck through and
answered in place (D-8.5.3, owner decision; propagated by D-000.10 in commit `f2d108c`): 20+
families, admitted by a named permissive allowlist, enforced fail-the-build. Planning therefore
resumed rather than re-blocking. The stale
`bmad-build-auto-result-8-5-a-curated-catalogue-ships-with-the-designer.md` is superseded by this
spec and should be removed or marked at close.

**DW-108 discharged at the plan gate:** the six scope constraints the regenerated `epic-8-context.md`
dropped are carried in `Never`, sourced from `spec-fonts/SPEC.md` `## Non-goals` rather than from the
cache.

**Note for the gate — `epic-8-context.md` is now a stale cache by its own rule.** `f2d108c` touched
`planning-artifacts/epics.md`, which is now newer than the cached context. It was NOT recompiled:
this is a plan-only dispatch, recompiling would leave a tracked modification with nothing to commit
it, and the constraints the cache is lossy about have been sourced from the SPEC kernel instead.
