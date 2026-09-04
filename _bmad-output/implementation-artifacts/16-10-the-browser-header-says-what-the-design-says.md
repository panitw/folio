---
title: 'Story 16.10: The browser header says what the design says'
type: 'refactor'
created: '2026-09-04'
status: 'done'
review_loop_iteration: 0
baseline_commit: '0176415'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/16-9-the-typography-panel-offers-what-you-can-use.md'
warnings: []
deferred: ['DW-183', 'DW-184', 'DW-185']
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The font browser's header carries a three-clause paragraph explaining how many families exist, that the
list is a dated snapshot rather than a live feed, and that variable-only families are hidden. The design
draws no such paragraph. Its header is a single 46px row — title, search, sort, view, close — and nothing
below it.

The paragraph goes. The header becomes the one row the design draws.

The count is not lost with it. The results toolbar already prints `N of M families`, derived from the same
number the paragraph quoted, and that line stays exactly as it is.

<intent-contract>

## Intent

**Problem:** `FontBrowser.tsx:287` renders `<p className="font-browser-disclosure">{familyIndexDisclosure()}</p>`
below the header row. The design has no paragraph there. **This prose is an invention, not a transcription:
across the mockup file, "families you can add" occurs 0 times, "snapshot taken" 0 times and "variable file
are not shown" 0 times, against a positive control of `AVAILABLE LOCALLY`, which occurs 1 time.**

**Approach:** Delete the paragraph and everything that existed only to feed it, leaving the header the
single row the design draws.

## Boundaries & Constraints

**Always:**
- **THE DESIGN'S HEADER IS ONE ROW.** `Font Browser.dc.html:293-296` is a fixed 46px flex row holding the
  title, a metadata span, a divider, search, sort, view and close. There is no second row and no paragraph.
- **DO NOT MINT A REPLACEMENT SENTENCE.** The mockup's own metadata span reads *"web font library · 1,946
  families"* and **this codebase has already refused to draw it** — `FontBrowser.tsx:21-27` records why:
  1,946 is what the upstream source PUBLISHED on the snapshot date, not what this designer can add, and
  "library" reads as live when the list is a build-time snapshot. **That refusal stands. Removing our
  sentence is not licence to draw theirs, or a shorter one of our own.** Following the design strictly here
  means drawing nothing in that slot, not drawing a claim we have twice ruled false.
- **THE COUNT SURVIVES, IN THE PLACE THAT ALREADY OWNS IT.** `resultLine` (`font-browser-model.ts:317-321`)
  prints `N of M families` in the results toolbar from `addableFamilyCount`. **It is untouched by this
  story**, which is what keeps a count in the browser at all.
- **Deleted code is deleted** — the 16.9 rule, applied again. If `familyIndexDisclosure()` has no consumer
  left, the function, its tests and the orphaned CSS rule go with the paragraph. **Measure its consumers;
  do not assume this one is the last.**
- **ASSERTIONS THAT ARE NOT ABOUT THE SENTENCE MUST SURVIVE THE SENTENCE.** `font-index.test.ts:241-242`
  asserts `addableFamilyCount === webFamilies.length + catalogueFaces.length` and
  `addableFamilyCount < familyIndexPublishedFamilies`. **Those are facts about the count `resultLine` still
  reads, and they sit inside a test named for the disclosure.** Deleting the test would delete them.
  **Move them to a test that outlives the sentence.**
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Putting **any** text in the header's metadata slot.
- Any change to `resultLine`, `addableFamilyCount`, or the results toolbar.
- Removing `familyIndexSnapshotDate` or `familyIndexPublishedFamilies` from `generated/font-index.ts` —
  they are generated, and other tests read them.

**Never:** a second sentence replacing the first · a count with two authorities · changing what
`resultLine` prints · touching the family dropdown, which lost this same sentence in an earlier story ·
deleting a count assertion because its surrounding test was about wording.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Open the font browser | Any state | Header is **one row**; **no disclosure paragraph** anywhere in the dialog | — |
| Read the header slot | — | **No sentence about counts, snapshots or variable files** — measured by absence, with a positive control | — |
| Results toolbar | No filters | `resultLine` prints `N of <addableFamilyCount> families`, **unchanged from before this story** | — |
| Results toolbar | Filtered to 12 of 340 | `12 of 340 matching families, out of <addableFamilyCount>`, unchanged | — |
| The count's own facts | — | `addableFamilyCount === webFamilies.length + catalogueFaces.length` and `< familyIndexPublishedFamilies` **still asserted somewhere** | — |
| `familyIndexDisclosure` | After removal | **Zero production consumers**; function, tests and CSS rule gone | — |
| Screen reader on the dialog | — | Dialog's accessible name is still `Font browser`; **no orphaned `aria-describedby` pointing at a removed node** | — |

</intent-contract>

## Code Map

Anchors at `0176415`. **Re-verify before editing** — and use `/usr/bin/grep` by absolute path: the shell's
`grep` in this repo returned **false zeros** three times during 16.9's verification.

**The paragraph**
- `folio-designer/src/FontBrowser.tsx:287` — `<p className="font-browser-disclosure">{familyIndexDisclosure()}</p>`.
- `folio-designer/src/FontBrowser.tsx:2` — the import of `familyIndexDisclosure`.
- `folio-designer/src/App.css:400-403` — the comment explaining why the region grows, and
  `.font-browser-disclosure`. **Both are orphaned by the removal.**
- `folio-designer/src/App.css:395` — `.font-browser-header { display: grid; gap: … }`. With one child left,
  **the grid and its gap are what still make it two rows.**

**The function and its tests**
- `folio-designer/src/font-index.ts:338-342` — `familyIndexDisclosure()`.
- `folio-designer/src/font-index.ts:86` `indexSnapshotDate` — **measured: its only reader is `:340`, inside
  the function being deleted.** `:87` `indexPublishedFamilies` appears to have no reader at all today;
  **measure both rather than trusting this line.**
- `folio-designer/src/font-index.test.ts:240-258` — two tests. **The first carries the two count assertions
  that must be rehomed.** The second is entirely about wording and goes.
- `folio-designer/src/App.test.tsx:1803` — a comment referring to the sentence still shipping in the
  browser. **It stops being true; correct it.** A comment is not a measurement, and a stale one is worse.

**What stays**
- `folio-designer/src/font-browser-model.ts:317-321` `resultLine`, and
  `font-browser-model.test.ts:285-288`. **Untouched.**

**The design**
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Font Browser.dc.html:293-296`.

## Tasks & Acceptance

**Execution:**
- [x] `FontBrowser.tsx` — delete the paragraph and its import.
- [x] **Measure `familyIndexDisclosure`'s remaining consumers, stating the population.** If zero, delete the
      function; if not, stop and report what still reads it.
- [x] `font-index.ts` — delete the function and any export orphaned by it, **each justified by a measured
      consumer count, not by inspection.**
- [x] `App.css` — delete `.font-browser-disclosure` and its comment, and make `.font-browser-header` the
      single row the design draws.
- [x] `font-index.test.ts` — **rehome the two count assertions first**, then delete the wording tests.
- [x] `App.test.tsx:1803` — correct the now-false comment.
- [x] **An absence guard: the browser renders no disclosure paragraph.** Red-prove it by restoring the
      paragraph, watching the new assertion fail, restoring the file by absolute path, confirming green.
      **Say where the failure landed.** 16.9's first report claimed guards it did not have.
- [x] A browser run: photograph the open browser's header.

**Acceptance Criteria:**
- Given the font browser open, when its header is read, then it is one row and contains no paragraph.
- Given the results toolbar, when it renders, then `resultLine` prints exactly what it printed before.
- Given the suite, when it runs, then the two count assertions still exist and still pass.
- Given `familyIndexDisclosure`, when the repository is searched, then it has zero consumers and does not
  exist.

## Design Notes

**This is the sentence's last home.** It was cut from the family dropdown earlier in this epic; the browser
header is the only other place it rendered. That is why removing it deletes a function rather than a line.

**The trap in this story is the rehoming, not the deletion.** Two assertions about `addableFamilyCount` —
the number `resultLine` still prints — live inside a test named for the disclosure. Delete the test and the
count silently loses its guard while every gate stays green. **A deletion story's characteristic failure is
taking a live assertion out with a dead subject.**

**One registered defect is discharged and one is NOT, and the difference matters.** The disclosure said
*"31 already on this machine"* while `AVAILABLE LOCALLY` counted 31 plus every stored face — two authorities
on one count. That divergence stops existing with the sentence. **The `resultLine` defect survives
untouched:** it can print `1274 of 1273` when a stored family has been dropped from the snapshot, because
`addableFamilyCount` never counts `orphanedStored`. **Do not record that one as closed.**

## Verification

**Nothing in Epic 16 is CI-verified** (DW-171); every gate is a local measurement with no machine watching.
Say what you did not run.

**Commands** — one per line, exit codes from `$?` immediately, never through a pipe or wrapper; zsh, so
`${PIPESTATUS[0]}` is wrong. **Read lint from `npx oxlint` directly, never `npm run lint`.**
- `cd folio-designer && npm test` — the count **falls** as the wording tests retire. The one permanent red
  is `canvas-authority-contract.test.ts:190` (DW-152); **match it by NAME, not count.**
- `cd folio-designer && npm run typecheck` — **the gate that catches the orphaned export.**
- `cd folio-designer && npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`.
- No Go is touched; say so rather than re-running it.

**A BROWSER RUN.** Photograph the open font browser's header and **report its measured height against the
design's 46px row.** Use `chromium-1217` via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (`chromium-1208` is a
428K stub, DW-180).

**Standing rules — re-run, never cite:** the per-row matrix audit reports **N rows, N results** (this matrix
is **7 rows**); state the population beside every zero and pair every absence claim with a positive control;
**a comment is not a measurement**; **use `/usr/bin/grep`** — the shell's `grep` gave false zeros here.

## Suggested Review Order

**The refusal, and what replaced it (nothing)**

- The refusal to draw the mockup's sentence stands; what is new is that ours goes too.
  [`FontBrowser.tsx:21`](../../folio-designer/src/FontBrowser.tsx#L21)

- The header is now the five children the design draws, and no sixth.
  [`FontBrowser.tsx:279`](../../folio-designer/src/FontBrowser.tsx#L279)

- One flex row at the design's own 46px, `padding: 0 14px`, `gap: 14px`.
  [`App.css:401`](../../folio-designer/src/App.css#L401)

**The deletion, and the export it orphaned**

- `familyIndexDisclosure` is gone; the comment records why `indexSnapshotDate` went with it.
  [`font-index.ts:85`](../../folio-designer/src/font-index.ts#L85)

**The rehoming — the point of the story**

- The two count facts, out of a test named for the sentence and into one named for the count.
  [`font-index.test.ts:253`](../../folio-designer/src/font-index.test.ts#L253)

**The guards**

- The closed set of header children: wording-free, so a NEW short sentence reds it too.
  [`FontBrowser.test.tsx:102`](../../folio-designer/src/FontBrowser.test.tsx#L102)

- The single-row claim, pinned where a gate actually runs — jsdom applies no stylesheet.
  [`design-contract.test.ts:127`](../../folio-designer/src/design-contract.test.ts#L127)

- The 46px measurement itself, in the suite only a browser run executes.
  [`font-browser.spec.ts:78`](../../folio-designer/e2e/font-browser.spec.ts#L78)

**Peripheral**

- A comment that had gone false: the subject no longer exists, not merely this surface.
  [`App.test.tsx:1801`](../../folio-designer/src/App.test.tsx#L1801)
