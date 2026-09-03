---
title: 'Story 16.7: Every row shows the typeface it names'
type: 'feature'
created: '2026-09-03'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: 'cd7e683'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/16-4-the-family-control-names-three-sources.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The design draws each row of the font menu as a name on the left and a few letters set in that typeface on
the right — `Aa Bb 123` for a Latin face, `กขค Aa` for a Thai one. Today every row is the name followed by
a sentence saying where the typeface is, and nothing shows what it looks like. Choosing a typeface from a
list of names alone means picking blind.

The font browser already does this properly, including a rule worth keeping: it never sets a sample in a
substitute typeface while implying it is the real one. It says so in words instead. The menu inherits both
the mechanism and the rule.

The one thing it cannot do is show a specimen for a typeface that is not on this machine — the third group,
`AVAILABLE TO INSTALL`, is over a thousand families whose bytes are not here, and fetching them to draw a
menu is the thing this epic has spent its time refusing. Those rows keep saying what they say now.

<intent-contract>

## Intent

**Problem:** Every row in the family control is a name plus a sentence about where the typeface lives.
Nothing shows the typeface. The design draws a specimen on the right of every row, and the font browser
already renders specimens correctly — the menu is the surface that does not.

**Approach:** Give the family control its own preview-face registry and draw a specimen, right-aligned, on
every row whose bytes are on this machine. Reuse the browser's mechanism and its honesty rule unchanged.

## Boundaries & Constraints

**Always:**
- **A SPECIMEN IS NEVER SET IN A FALLBACK WHILE IMPLYING IT IS THE FAMILY.** This rule already exists in
  `FontBrowser.tsx` and is inherited verbatim. A row whose face is not registered shows **no specimen** —
  never the sample text in the panel's own typeface.
- **The specimen is decorative to assistive technology.** It carries `aria-hidden`, so **no option's
  accessible name changes.** The e2e specs and the browser roundtrip witness address these rows by name.
- **`AVAILABLE TO INSTALL` rows show no specimen and keep their existing note.** Their bytes are not on
  this machine and this story does not fetch them.
- **Sample text follows the face's own script**, from the `scripts` the row already carries — Thai faces
  get a Thai sample, as the design draws.
- **The dropdown's registry is its own instance.** `PreviewFaceRegistry.show()` releases every family it is
  not given, and `close()` documents the modal as its only caller — one shared instance would have the two
  surfaces releasing each other's faces.
- **Faces are registered only while the dropdown is open**, and released when it closes.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Fetching bytes for any `web`-tier row, on any trigger, including scroll.
- Removing the per-row note from `AVAILABLE TO INSTALL`.
- Any change to `previewFaceFamily`, `openPreviewFaceRegistry` or the browser's own specimen rendering.

**Never:** a specimen drawn in a substitute face · a specimen that changes an accessible name · measuring
text in the browser (AD-17) · a second copy of the registry's logic · blocking the dropdown on any read.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Local-tier row | `Arimo`, bytes shipped | Name left, `Aa Bb 123` right, set in Arimo | — |
| Thai-covering row | A face whose `scripts` include `thai` | Sample is Thai, `lang="th"` | — |
| Stored row | A face this machine downloaded | Specimen set in it, read from the store, no network | — |
| Web-tier row | One of ~1,273 not on this machine | **No specimen**, existing note unchanged | Never fetched |
| Face registered but bytes unreadable | Store read fails | **No specimen**, row still pickable | Silent, never a fallback sample |
| Declared chain row | `body` | Specimen set in the face the chain resolves to; none if it resolves to no carried face | — |
| Dropdown closed | — | Every face this control registered is released | — |
| Screen reader on any row | — | Accessible name is the family name and its note, **specimen absent from it** | — |

</intent-contract>

## Code Map

All anchors at `cd7e683`. **Re-verify before editing.**

**The mechanism to reuse, not reimplement**
- `src/preview-face-registry.ts:53` `PreviewFaceBytes`, `:62` `PreviewFaceStatus`
  (`preparing`/`ready`/`unavailable`), `:64-72` the registry — `show(families)` **releases every family it
  is not given**, `statusOf(family)`, `close()` whose comment says *"The modal closing is the only
  caller"*. **That comment stops being true; update it.**
- `src/preview-face-family.ts:9` `previewFaceFamily(family)` — the collision-proof CSS family name.
- `src/FontBrowser.tsx:195-207` — **the pattern to copy, including the honesty rule verbatim** and the
  `lang="th"` handling.

**The byte sources, per tier**
- `src/App.tsx:389-404` `browserSpecimenBytes` — local reads `source.face.url`; stored reads the store by
  key; **web calls `fetchWebFamily` and this story's version must not.**
- `src/generated/font-catalogue.ts:2+` — catalogue faces are hashed local URLs, so local-tier bytes are
  offline and cheap. **There are no `@font-face` rules for them: measured 0 of 2 tracked stylesheets,
  against a positive control of 3 files matching `font-family`.** The registry is the only route.

**The rows**
- `src/App.tsx:1936` `declared` — group 1 is the document's declared **family names**, which is why `body`
  appears. Resolving one to a face goes through the document's font map, the way the canvas paints.
- `src/App.tsx:2096-2108` the listbox and its option rows; `familySourceNote` supplies today's per-row text.
- `src/App.css` — `.property-option` is the row; the specimen needs a right-hand column.

## Tasks & Acceptance

**Execution:**
- [ ] `folio-designer/src/App.tsx` — open a preview-face registry for the family control, scoped to the
      dropdown being open, and release it on close.
- [ ] `folio-designer/src/App.tsx` — a bytes reader for the control: local from the catalogue URL, stored
      from the store, **`web` returns `undefined` without fetching.**
- [ ] `folio-designer/src/App.tsx` — render the specimen on each row, right-aligned, `aria-hidden`, set in
      `previewFaceFamily(...)`, **omitted entirely unless `statusOf` is `ready`.**
- [ ] `folio-designer/src/App.tsx` — resolve a declared chain to the face it paints with; no specimen when
      it resolves to none.
- [ ] `folio-designer/src/preview-face-registry.ts` — correct the `close()` comment: the modal is no longer
      its only caller.
- [ ] `folio-designer/src/App.css` — the row becomes name left, specimen right, without changing the row's
      height enough to alter the 168px scroller's row count materially.
- [ ] Tests — one per matrix row, driving the real path in a mounted designer.

**Acceptance Criteria:**
- Given a local-tier row, when the dropdown opens, then its specimen is set in that face and the row's
  accessible name is unchanged.
- Given a web-tier row, when the dropdown opens, then no specimen is drawn and **no fetch is made** —
  asserted with a fetch spy that must record zero calls.
- Given a face whose bytes cannot be read, when its row draws, then it shows no specimen and never the
  sample in a substitute face.
- Given the dropdown closes, when the registry is inspected, then every face it registered is released.

## Design Notes

**Three decisions are taken here as defaults and are the owner's to overturn at the gate.**
**(1) The web group gets no specimen** — the alternative is fetching to draw a menu, and a pick already
blocks up to 30 s on a stall and 180 s against a slow host. **(2) The specimen replaces the per-row note in
the first two groups**, because the design draws name and specimen only and the group heading already says
where the bytes are. **This narrows D-16.R.72's three-carrier ruling** — heading, note, and the row's move
between groups — to two carriers for those groups; the note survives where it does the most work, on the
group that is not on this machine. **(3) A declared chain shows the face it paints with**, so the menu and
the canvas agree.

**The honesty rule is the reason this story is not just CSS.** `FontBrowser.tsx:195-201` refuses to set a
sample in a substitute face and says so in words instead. A dropdown row has no room for that sentence, so
its version of the rule is to draw nothing — which is why `statusOf` gates the render rather than a
`try`/fallback.

## Verification

**Nothing in Epic 16 is CI-verified** (DW-171). Every gate is a local measurement with no machine watching.
Say what you did not run.

**Commands** — one per line, exit codes from `$?` immediately, never through a pipe or wrapper; zsh, so
`${PIPESTATUS[0]}` is wrong. **Read lint from `npx oxlint` directly, never `npm run lint`.**
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts:190`
  (DW-152); **match it by NAME, not count.**
- `cd folio-designer && npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` ·
  `npm run test:e2e:compile` — all `rc=0`.
- `cd lint && go test -count=1 ./...` and `cd folio-go && go test -count=1 ./...` — no Go here; the failing
  leaf set must stay exactly `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`.

**A BROWSER RUN IS REQUIRED, AND A SCREENSHOT IS THE EVIDENCE.** This story is typography: jsdom does no
layout, `getComputedStyle` is banned outside the canvas, and **no test in this repository can see whether a
specimen actually rendered in its own face.** Use `chromium-1217` (Chrome for Testing 147.0.7727.15) via
`PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH`; `chromium-1208` is a 428K stub (DW-180). Photograph the open
dropdown and report the bounding boxes.

**Standing rules — re-run, never cite:** the per-row matrix audit reports **N rows, N results** (this
matrix is **8 rows**), a module-level-only row is a flagged PARTIAL; state the population beside every
zero and pair every absence claim with a positive control; **a comment is not a measurement** — this spec
exists partly because a comment about build-time `@font-face` rules was read as one and measured false.
