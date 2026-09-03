---
title: 'Story 16.8: A new document opens in a typeface with a name'
type: 'feature'
created: '2026-09-04'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: '69bec37'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/16-4-the-family-control-names-three-sources.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

A new document opens with its text set in a chain called `body`, and `body` is what the font menu shows.
That is a piece of internal vocabulary standing where a typeface name belongs — nobody chooses a font
called `body`.

Roboto becomes a typeface the engine ships, beside the three Noto faces it already carries, and a new
document opens in it. Because the engine holds the bytes, no document has to carry them: the starter
template stays a fraction of a kilobyte and every existing `.folio` is untouched.

Thai and CJK still work. The chain keeps its Noto fallbacks behind Roboto, so a document that mixes scripts
renders all of them, exactly as it does today.

The cost is honest and it is paid once: everybody's download grows by about 348 KB whether they use Roboto
or not, because the engine now carries a fourth face.

<intent-contract>

## Intent

**Problem:** `starter.folio` declares `"fonts": {"body": [...]}`, so a new document's only typeface is named
after a role rather than a face, and the family control shows `body`. The engine ships only Noto Sans, Noto
Sans Thai and Noto Sans SC, so no widely-recognised Latin face can be named without a document carrying it.

**Approach:** Ship Roboto as a fourth engine face, and open new documents in a chain named for it.

## Boundaries & Constraints

**Always:**
- **THERE IS EXACTLY ONE ROBOTO.** The bytes the engine ships **must be byte-identical** to the designer's
  catalogue Roboto (`catalogue-roboto.*.ttf`, SHA-256 `e688a215e0841b6e…`, 347.6 KB). Two cuts under one
  name would make `"fontFamily": "Roboto"` render differently depending on which path produced the
  document — against this product's byte-identical guarantee. **`folio-go/testdata/fonts/Roboto-Regular.ttf`
  is a DIFFERENT cut (`79e851404657dac2…`, 167.7 KB) and is a test fixture; it is not the shipped face and
  is not touched.**
- **The licence travels with the bytes.** Roboto is **OFL-1.1** — an admitted identifier in both
  `lint/internal/licence/classify.go` and `font-licence.ts` — and ships with `LICENSE-OFL.txt` and
  `NOTICE.md` beside the face, exactly as each Noto face does.
- **Script coverage is not reduced.** The starter chain keeps `Noto Sans Thai` and `Noto Sans SC` behind
  Roboto. **A Thai or CJK document must render exactly as it does today.**
- **No existing document changes.** Adding a face to the shipped set may not alter the render of any
  document that does not name it. **The golden fixtures are the proof and must be unmoved.**
- **An engine-shipped face needs no embedding**, which is how the three Noto faces already work.
- The payload grows for every user. **State the measured delta; do not bury it.**
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- **Whether picking Roboto from `AVAILABLE LOCALLY` still embeds it.** Default taken here: **it does not** —
  a shipped face needs no copy, which is the rule the Noto faces already follow. This changes what picking
  Roboto does today and is the owner's to overturn.
- Removing Roboto from the designer's catalogue, or changing any other face's tier.
- Any change to the chain's fallback order.

**Never:** a second cut of Roboto anywhere in the shipped payload · reducing Thai or CJK coverage ·
changing the render of a document that does not name Roboto · embedding Roboto into `starter.folio`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| New document | Starter template | Opens in Roboto; the family control shows `Roboto`, never `body` | — |
| Latin text | Default chain | Rendered in Roboto by the engine, no embedded face | — |
| Thai text | Default chain | Falls through to Noto Sans Thai, unchanged from today | — |
| CJK text | Default chain | Falls through to Noto Sans SC, unchanged from today | — |
| Existing document naming `body` | A `.folio` saved before this story | Renders exactly as before; `body` is still a legal chain name | — |
| Golden corpus | The 23 fixtures | **Digests unmoved** — none of them names Roboto | Any movement is a defect, not a re-baseline |
| Document naming Roboto, opened elsewhere | Any build of this engine | Renders in Roboto with nothing carried | — |
| Offline payload | A release build | Grows by the face's measured size, once, for everyone | Stated in the report |

</intent-contract>

## Code Map

Anchors at `69bec37`. **Re-verify before editing.**

**The engine's shipped set**
- `folio-go/fonts/fonts.go:40-70` — three `//go:embed` directives and `Shipped()` returning a
  `folio.FontSet` **keyed by the exact names a document's chains reference.** Roboto joins here.
- `folio-go/fonts/notosans/` — **the pattern to copy exactly**: the `.ttf`, `LICENSE-OFL.txt`, `NOTICE.md`.
- `folio-go/internal/fontset/licencesignature.go` — the licence machinery an admitted face passes through.

**The bytes**
- `folio-designer/src/generated/runtime/catalogue-roboto.e688a215e0841b6e4edb.ttf` — **347.6 KB, SHA-256
  `e688a215e0841b6e…`. This is the cut to ship.** Its licence text is already recorded in
  `src/generated/font-catalogue.ts` as `licence: "OFL-1.1"`.

**The starter**
- `folio-designer/public/templates/starter.folio` — **0.4 KB**, `"fonts": {"body": ["Noto Sans", "Noto Sans
  Thai", "Noto Sans SC"]}`, and **zero `fontFamily` occurrences**, so every element takes the default chain.
- Loaded by `folio-designer/src/startup-sequence.ts` and `src/main.tsx`.

**The designer's view of a shipped face**
- `folio-designer/src/shipped-face-family.ts:63` `canvasFragmentFallbackStack` — the three Noto names and
  `sans-serif`; `:67` `shippedFaceFamily`, `:75` `isShippedFaceName`.
- `folio-designer/src/font-index.ts:103` `localTierHolds` — the catalogue tier that currently owns Roboto.

**The guards that will notice**
- `fixtures/*/expected.json` — 23 golden digests, `sha256` per fixture. **0 of 116 fixture files reference
  `starter`; they carry their own `input.folio`, which is why the starter change alone cannot move them.**
- `folio-designer/scripts/verify-offline-release.mjs` — runs inside `npm run build`.

## Tasks & Acceptance

**Execution:**
- [ ] `folio-go/fonts/roboto/` — add the face, `LICENSE-OFL.txt` and `NOTICE.md`, **copying the catalogue
      cut byte-for-byte and asserting the digest**, not re-downloading it.
- [ ] `folio-go/fonts/fonts.go` — `//go:embed` it and add `"Roboto"` to `Shipped()`.
- [ ] `folio-go` — a test asserting the shipped Roboto's SHA-256 equals the designer catalogue's, so the
      one-Roboto rule is machine-checked rather than stated.
- [ ] `folio-designer/public/templates/starter.folio` — `"fonts": {"Roboto": ["Roboto", "Noto Sans Thai",
      "Noto Sans SC"]}`.
- [ ] `folio-designer/src/shipped-face-family.ts` — Roboto is a shipped face name; the canvas can paint it
      without an embedded copy.
- [ ] Tests — one per matrix row, driving the real path.
- [ ] Report the **measured** payload delta from the offline release manifest, before and after.

**Acceptance Criteria:**
- Given a new document, when the family control opens, then `IN THIS TEMPLATE` shows `Roboto` and no chain
  named `body` exists.
- Given a document containing Thai and CJK text, when it renders, then its output is **byte-identical to
  the same document rendered before this story**.
- Given the 23 golden fixtures, when they run, then **every digest is unmoved**.
- Given a document naming Roboto, when it is saved, then it carries **no embedded face** for it.

## Design Notes

**The one-Roboto rule is the whole risk of this story.** Two cuts already exist in this repository under
one name, differing by 180 KB. Shipping the wrong one would not fail a test — it would render, and every
document naming Roboto would quietly disagree with every other build. **That is why the digest assertion is
a task and not a note.**

**Why the payload cost is accepted rather than avoided.** The alternative that costs nothing is naming the
chain `Noto Sans` — the face already rendering today. The owner chose to ship Roboto instead, knowing the
download grows for everyone: **the default typeface is the one thing every author sees before they have
chosen anything**, and it is worth ~348 KB paid once rather than 348 KB carried by every document that
wants it.

## Verification

**Nothing in Epic 16 is CI-verified** (DW-171). Every gate below is a local measurement with no machine
watching. Say what you did not run.

**Commands** — one per line, exit codes from `$?` immediately, never through a pipe or wrapper; zsh, so
`${PIPESTATUS[0]}` is wrong. **Read lint from `npx oxlint` directly, never `npm run lint`.**
- `cd folio-go && go test -count=1 ./...` — **this story touches Go.** The failing leaf set must stay
  exactly `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`; **anything else moving is this story's.**
- `cd lint && go test -count=1 ./...` — `rc=0`, four packages `ok`.
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts:190`
  (DW-152); **match it by NAME, not count.**
- `cd folio-designer && npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` ·
  `npm run test:e2e:compile` — all `rc=0`.
- **The 23 golden digests, and this time they are the point:** run the fixture corpus and report
  **unmoved**, per fixture. A moved digest is a defect in this story, never a re-baseline.

**A BROWSER RUN.** Open a new document and photograph the family control: `IN THIS TEMPLATE` reads
`Roboto`, its specimen is set in Roboto, and no `body` row exists. Use `chromium-1217` via
`PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (`chromium-1208` is a 428K stub, DW-180).

**Standing rules — re-run, never cite:** the per-row matrix audit reports **N rows, N results** (this
matrix is **8 rows**); state the population beside every zero and pair every absence claim with a positive
control; **a reported digest names its algorithm and its byte range**; **a comment is not a measurement.**
