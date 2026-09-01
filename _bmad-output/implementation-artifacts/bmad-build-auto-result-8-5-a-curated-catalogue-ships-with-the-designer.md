---
status: 'blocked'
---

# BMad Build Auto Result — Story 8.5: A curated catalogue ships with the designer

Status: blocked
Blocking condition: intent gap

Dispatch: classic intent, plan-only (`Halt after planning.`), at `afa33905260d5d5397d88199159d2dd76976f3c2`, tree clean, branch `main`.
No spec was written to `_bmad-output/implementation-artifacts/8-5-a-curated-catalogue-ships-with-the-designer.md`, and no code was written. The story-key path is deliberately left free so the next dispatch resumes at planning rather than matching a `blocked` file by name.

---

## The gap, in one line

**The catalogue's central content — which font families it contains and how many — is an explicitly
recorded OPEN QUESTION in the canonical contract that governs this story, and nothing in the epic,
the decision log, the register or the dispatch selects between the readings.**

This is not an omission I could route around. It is written down as open, in a list where every
question that *has* been settled is marked settled in place.

`_bmad-output/specs/spec-fonts/SPEC.md:130`, under `## Open Questions`, verbatim:

> - Which families make the shipped catalogue, how many, and who curates the list as it changes?

`_bmad-output/specs/spec-fonts/font-catalogue.md:26`, verbatim:

> Which families, and how many, is an open question in `SPEC.md`.

For contrast, the immediately preceding entry in that same list is struck through and marked
**SETTLED (Story 8.3)** with its derivation. The families question carries no such mark. The
absence is deliberate and recorded, not an oversight I should close by picking.

---

## Why picking a reading myself would be wrong, not merely bold

1. **The outcomes are observably different.** Different families mean different committed binaries,
   different bytes, different NOTICE records, different licence identifiers, and a different
   finished epic weight. D-8.4.27(d) rules that Story 8.4d sets the budget threshold **once,
   against the epic's finished weight** — so whatever 8.5 picks *is* the number 8.4d will gate on.
2. **It is an ownership question, not only a content one.** The open question's own third clause is
   *"and who curates the list as it changes"*. A build agent answering that would be inventing a
   curation owner.
3. **AD-26 is a licence boundary, and the asset half of it is NOT enforced by code.** Measured in
   `lint/internal/manifest/manifest.go`: `ResolveAssets` computes a licence family and then
   **discards it** — `if _, spdx := licence.ClassifyLicenceText(licenceText); spdx != ""` — so the
   label is published but never checked. A copyleft font licence classifies as `GPL-3.0`, is written
   into `lint/MANIFEST.md` as a row, and `ResolveAssets` returns **nil error**. Unclassifiable
   licence text soft-fails to the literal `"SEE NOTICE"` and also passes green. The only code that
   switches on `licence.FamilyCopyleft` is `rules.ScanLicenceGraph`, whose input is a Go module
   directory — it never sees an asset directory. **Nothing in this repository would refuse a
   badly-licensed catalogue face.** Choosing faces unattended against an unenforced boundary is
   precisely the decision the dispatch flagged as the one most able to violate AD-26 by accident.

---

## A SECOND, INDEPENDENT BLOCKER — AC1 is not executable here for any new family

Even given a settled family list, AC1's *"derived ahead of the build by the same replayable
derivation the shipped set uses"* cannot be carried out in this environment. Measured:

- `tools/fontgen/instance_faces.py` drives a hardcoded module-level `UPSTREAM` list of **exactly 3
  entries** (`"key": "Noto Sans"`, `"Noto Sans Thai"`, `"Noto Sans SC"` at lines 119/135/153). It
  does no directory walk and no globbing; each entry names a literal source basename.
- Its input directory `.font-sources/` is **gitignored** (`.gitignore:88`, `/.font-sources/`) and
  carries **0 git-tracked files**. It holds only the three Noto variable-font sources.
- **Nothing in the repository fetches an upstream font source.** No script, no CI step, no Make
  target. The only record of provenance is prose — `UPSTREAM`'s `src_url` field and each
  `NOTICE.md`'s "Provenance — the source" table.
- Therefore a NEW family's upstream source must be obtained **over the network, by hand, once**, and
  neither `.font-sources/` nor the pinned `.fontgen-venv/` is committed or reconstructible from the
  repo.
- The bootstrap is also undocumented and self-contradictory on first run: `UPSTREAM` requires an
  `out_sha256` and `out_bytes` that cannot be known until the derivation has been run once, so the
  first run of a new face necessarily fails its own assertion.

**This is a dependency gap under the READY standard's "Sufficient" criterion**, independent of which
families are chosen. It needs a ruling on how catalogue sources are provisioned and recorded before
any implementer can execute AC1.

---

## Findings the plan gate should carry regardless of how the gap is ruled

These were measured this dispatch and are worth keeping whoever resolves the above.

**1. Adding a 4th derived face breaks the reproduction check on a hardcoded string, not on bytes.**
`folio-go/fontgen_matrix_test.go:117` reads `const wantWitness = "derived and compared 3 of 3 faces"`
while `tools/fontgen/instance_faces.py:367` prints `f"...{compared} of {total} faces"`. A 4th
`UPSTREAM` entry makes the script print `4 of 4` and the **Go test fails on the witness while the
derivation is perfectly green**. The literal must move in lockstep, or the witness should be derived.

**2. Two kinds of assurance currently sit under one heading, and the catalogue must not inherit the
confusion.** The 3 Noto faces are *reproduced* (replayable derivation, verified). The 3 IBM Plex
faces are *provenance-only* — vendored static upstream files whose source and shipped sha256 are the
same value, which is the record. This matches DW-86's residual. **The spec must say which kind each
catalogue face gets and must not let one heading cover both.**

**3. Where the faces land is unruled AND load-bearing — D-8.4.30's residual is live.** Measured: the
magic-byte guard `committedFontsTheLicenceGateCannotSee` in
`folio-designer/src/font-binary-identity.test.ts` walks **`folio-designer/public/fonts` and nothing
else**. `manifest.ResolveAssets` does walk the whole repo, but only for `.ttf`/`.otf`/`.ttc` and only
for git-tracked directories. So a catalogue face in a NEW directory: `ResolveAssets` **would** force
its `LICENSE*` + `NOTICE*` and add a `MANIFEST.md` row; the extension-class guard **would not** see
it at all. A `.woff2` in a new directory is invisible to every guard in the repo.

**4. There is exactly ONE route into the bundle, and it is not `public/`.** `vite.config.ts:32` sets
`publicDir: false` with the comment *"Never publish the mutable source files from public/ as
application URLs."* Measured: `dist/` contains zero `public/fonts/*` files. The only path to a
content-addressed ("verified") URL is `build-wasm.mjs`'s `fingerprint()` → `offline-assets.ts`
`?url` → Vite `assets/[name]-[hash][extname]`. AC2 is nearly free along that route and unsatisfiable
along any other.

**5. Do NOT give catalogue faces S1 rows.** `verify-offline-release.mjs` pins `s1Ids` and
`semanticLabels` by **ordered exact equality**, requires `cachedRows.length === 4`, reads
`s1.rows[4]` **positionally**, and requires `cjk-font` to be the `Math.max` of all `*font` rows.
`src/release-payload.ts` additionally types the row ids as a literal union. Added as ordinary assets
they are covered generically and break nothing.

**6. A hard ceiling nobody's build script checks.** `src/release-payload.ts` declares
`maximumCacheAssets = 64` against a measured `assetCount` of 23 — so **at most 41 further assets**
before `parseS1Payload` returns `undefined` and the first-run load screen silently loses its payload.
No build script catches this; only a Vitest test and runtime do.

**7. AC4 has an existing home for per-entry weight, in the wrong unit.** `s1.cacheAssets[]` already
carries `{assetUrl, bytes}` for every asset — **uncompressed**. Per-asset Brotli exists only as `.br`
files on disk and is recorded nowhere. And `s1VisibleBytes` is structurally incapable of seeing
catalogue fonts: it sums four hardcoded needles and already misses the 174,949 Brotli bytes of IBM
Plex. Recording the catalogue's weight against it would be a **false zero**. Per D-8.4.27(d), 8.5
records the figure with its exact invocation, commit and tree state; **8.4d owns the threshold**.

**8. AC3 cannot be phrased at the only layer that could literally observe it.** Measured: there is
**no forbidden-host scan anywhere** — zero hits for `gstatic|googleapis|fonts.google` in any scanned
source population. The only literal "no request leaves the machine" proof is Playwright
(`e2e/engine-worker.spec.ts` does `context.setOffline(true)`), and CI contains **zero** occurrences
of `playwright`. The reusable model is `canvas-authority-contract.test.ts`, which scans the whole
designer with a hand-written **comment-stripping character scanner** and red-proves both directions.
Any disclosure must use the live trigger *"when CI executes the Playwright suite"* (D-8.4.25(b)).

**9. Rulings that are already settled and must NOT be re-opened as part of this gap.** The catalogue
ships **Regular-only** (weighted faces collide with Epic 11 — lead's pre-story grounding). The
per-entry record is **settled** in `font-catalogue.md`: `family`, `style`, `licence`, `source`,
`bytes`, `scripts`. **CJK is excluded.** 8.5 = *which faces an author may pick*; 8.6 = *picking one
writing the chain* (D-8.4.26). 8.5 **may add binaries** (D-8.4.27(d)).

---

## The unanswered questions, for the engineering lead or the owner

1. **Which families make the catalogue, and how many?** (SPEC open question, verbatim above.)
   Constrained by, but not determined by: OFL-1.1 or equivalently redistribution-permitting; Latin
   and Thai scale; no CJK; Regular-only.
2. **Who curates the list as it changes?** (Third clause of the same open question.)
3. **How are catalogue upstream sources provisioned and recorded**, given `.font-sources/` is
   gitignored, nothing fetches it, and the bootstrap first-run is undocumented? Does the catalogue
   extend `UPSTREAM` (reproduction assurance) or vendor static upstream files (provenance
   assurance, the IBM Plex precedent)?
4. **Where do catalogue faces land on disk**, given D-8.4.30's residual leaves the extension-class
   guard place-keyed to `folio-designer/public/fonts/`?

---

## Verification measured this dispatch

Baseline probes only — no code was written, so no suite could regress.

- `cd folio-go && go test -count=1 -tags=matrix -run TestShippedFacesReproduceFromUpstream ./...`
  → **FAIL**, verbatim `fontgen: fontTools is not importable by this interpreter`. This is DW-86's
  registered could-not-execute identity, **not** a byte divergence. It never compared bytes.
- Same with `FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python`
  → **PASS in 8.25s, non-vacuously**: three `matches the recorded derivation` lines,
  `fontgen: derived and compared 3 of 3 faces`, `fontgen: OK`.
  **No divergence in the shipped faces. No golden is at risk on this ground.**
- `cd lint && go test -count=1 ./...` → **4 packages ok, 0 FAIL**, uncached.
- `cd folio-designer && npm test` → **39 files / 383 passed**. `npm run typecheck` → clean, exit 0.
  `npm run lint` → exit 0, **exactly 4** `only-export-components` warnings
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1324,1331`). `npm run test:e2e:compile` → exit 0
  (`tsc -p tsconfig.e2e.json --noEmit` — a COMPILE, not a run).

The full Go suite, the matrix legs, `TestCrossTargetByteIdentity`, the golden digest diff and the
offline release build were **not** run this dispatch: no code changed, and the story halted before
any implementation existed to verify. They are the implementer's cadence, not the plan gate's.
