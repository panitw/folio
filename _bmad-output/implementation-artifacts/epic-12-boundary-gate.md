# Epic 12 — Boundary Gate

**Run 2026-09-06 at `0c0f3e9`.** Epic 12 is five stories, all `done`: 12.1 band heights, 12.2 locale and
UTC offset, 12.3 table header and alternating rows, 12.4 padding, 12.5 the band-boundary drag.

**VERDICT: CLOSE, with one carried failure that is not this epic's. This is a pass, and the sentence
below says exactly what it is a pass of.**

---

## Why this gate mattered more than most

**No run of either heavy suite had ever covered anything in Epic 12.** Story 12.4's late close established
it in the sharpest form: those suites *"were not run at `fd4da07` either, so no run of them has ever
covered this change."* Every one of the five stories deferred matrix and Playwright to this gate by the
run's end-of-epic cadence. **This is the first execution of either suite against any Epic 12 work, and it
produced two genuine findings that five story-level runs could not.**

---

## Leg 1 — Cross-target matrix: PASS

`cd folio-go && FOLIO_FONTGEN_PYTHON=<pinned> go test -tags=matrix -count=1 -v .` — **unfiltered.**

| | |
|---|---|
| Result | **709 PASS / 0 FAIL / 5 SKIP**, `ok`, rc 0, 44.3s |
| `TestCrossTargetByteIdentity` | **PASS (25.03s)** — the four-target byte-identity gate, which is the point of the suite |
| `TestFMAProbeDiverges` | **PASS (1.23s)** |
| `TestShippedFacesReproduceFromUpstream` | **PASS (8.57s)** |

**Five skips, named rather than counted:** `TestBrowserAuthoredRoundTripWitness`,
`TestXrefEntriesRejectsMalformedSubprocess`, `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`,
`TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`, `TestTwoTablesWithPageCountFooterRenderConsistently`.

### Deviation 1 — the gate's own false start, recorded because hiding it would be worse

**The first invocation was filtered and produced a green that asserted nothing.** I ran
`-run 'TestTargetRenderHash|TestTargetProbeHex'`, copying the test names out of `matrix.yml`. Both ran,
both passed, rc 0 — and both are **deliberate no-ops** without `FOLIO_MATRIX_TARGET`. They print a
paragraph saying so, naming the variable, naming the correct local tests, and explaining that CI never
reaches the path because every per-leg job sets the variable explicitly.

**CI's names are right for CI, where each job drives one target, and wrong for a laptop, where a
different pair of functions drives all four in one process.** Filed as D-000.28. **Standing rule: a
boundary gate runs its suite unfiltered** — the gate exists to discover what story-level runs could not,
and every filter is a pre-judgement of the answer.

### Deviation 2 — one failure on the corrected run, environmental, resolved rather than carried

`TestShippedFacesReproduceFromUpstream` failed with *"fontTools is not importable by this interpreter"*,
having picked up Homebrew's Python 3.14 while the derivation is pinned to **Python 3.12.13 / fontTools
4.63.0**. **The test refused to let a missing dependency read as a moved font** — it named the
interpreter, gave the exact install command, and said to check the toolchain first because fontTools is
not byte-deterministic across versions. Re-run with `FOLIO_FONTGEN_PYTHON` pointing at an interpreter at
exactly the pinned versions: **PASS**. Recorded as resolved, not as a could-not-look.

---

## Leg 2 — Playwright: PASS with one carried failure

`cd folio-designer && npx playwright test` — **unfiltered, no `--grep`.** 16 spec files, 36 tests.

| Run | Result |
|---|---|
| First (gate discovery) | 33 passed / **3 failed**, 8.5 min |
| After repair | **35 passed / 1 failed**, 7.9 min |

### Finding 1 — Story 12.5's own e2e spec failed on its first execution. It was the SPEC, not the feature.

`band-boundary-drag.spec.ts:85` — the footer boundary painted no proposal; `.band-boundary-proposal`
resolved to 0 elements, 64 times over 30 seconds. The header/content test passed doing structurally the
same thing.

**Diagnosed before repair, by measurement.** `playwright.config.ts` declares no viewport, so every run is
1280×720. The footer handle sits at **y=883 — 163px below the window**, and `page.mouse.move()` drives raw
viewport coordinates **without scrolling**, unlike `locator.click()`. `document.elementFromPoint` at the
press point returned `NONE`: **the press landed on nothing.** The header handle at y=213 is inside the
window, which is exactly why the identical test passed and this read as a footer-only product bug.

**The positive control is what makes the diagnosis trustworthy.** With `scrollIntoViewIfNeeded()` first,
the box moves 880 → 583 and the footer drag works end to end: proposal painted, readout `64`, committed
footer height **40 → 66** — a 24pt upward drag growing the footer, sign flip correct, snapped to the 6pt
grid. **The feature was never broken.**

**The repair is not the scroll.** The test now **asserts its press point is inside the viewport height**,
by name. `toHaveCount(1)`, a non-null `boundingBox()` and even `isVisible()` are all true for an
off-screen element — so the test proved the handle existed and then pressed somewhere else, and reported
a mysterious zero. It now fails saying so. Applied to the header test too, which passes today only
because of where the page happens to sit.

### Finding 2 — a REGRESSION Story 12.5 caused, visible only to a real browser's strict mode

`image-asset.spec.ts:222` located a component's resize handle with `getByRole('button', { name: /^Resize / })`
— a prefix regex over the whole page. 12.5 added `Resize the page header` and `Resize the page footer`, so
**three elements matched and strict mode failed.** That test passed before this story.

12.5 had anticipated this collision class and mitigated it — but it surveyed
`getByRole('region', { name: …, exact: true })` in two specs. **An `exact: true` survey cannot see a
regex, and a `region` survey cannot see a `button`.** The reasoning was right and its scope was one
locator-shape too narrow.

**Fixed by scoping the locator to the component**, not by renaming the handles — the component id is
minted at runtime, so scope is the honest fix and exactness is unavailable. The locator asked for *any*
resize handle on the page while meaning *this component's*.

**The sweep that would have caught it, and now has:** all 16 `e2e/` files, **356 by-name/by-text locator
calls, 44 of them regex-named**, every regex compiled and executed against both handle strings. **Exactly
one offender.** Positive control fires (`/^Resize /` matches); discrimination control correctly silent
(`component-manipulation.spec.ts`'s `/Resize e/` does not match — after `"Resize "` the handle names carry
`t`, not `e`).

### The carried failure — not Epic 12's, established by ancestry rather than assertion

`browser-native-roundtrip.spec.ts:339` › *fresh authored sessions close exactly through admitted Preview
and native Folio*. It dies inside the font-family combobox, waiting for an option named `body` — nothing
to do with band boundaries.

- **Last modified 2026-08-30 in `a0cf8c2`**, which `git merge-base --is-ancestor` confirms is an ancestor
  of Epic 12's first commit `bcb6ebb` (2026-09-05), **314 commits earlier.**
- **No Epic 12 commit touches the file.**

**Its symptom changed and its cause did not.** Previously an open-ended hang (measured at 300s and 1500s
budgets, ~27 min wall at 9% CPU); now a timeout error at the 5-minute budget. **Flagged so nobody reads
the changed symptom as a new defect.** Tracked by DW-193 — no workflow runs this suite at all.

---

## What this gate could NOT look at

- **The e2e suite is executed by no workflow.** DW-193 is open. This gate ran it by hand; nothing runs it
  on a commit, so nothing prevents these three specs regressing tomorrow.
- **`tsc -p tsconfig.e2e.json --noEmit` proves the specs compile, never that they pass.** It is not
  coverage and is not reported as such anywhere in this epic's record.
- **The matrix's four legs ran from one process on darwin/arm64.** CI's per-target Docker legs
  (linux/amd64, linux/arm64) were not exercised here; `TestCrossTargetByteIdentity` is the local gate
  D-000.4 designates and is what passed.

---

## The status distinction this epic must not blur

Story 12.5's `band-boundary-drag.spec.ts` was reported at its story boundary as **written-not-run.** That
was correct then, and it was not coverage. **At this gate it is run, and 3 of 3 passing** — after its own
first execution found a defect *in itself*. Those are different claims about one file and the record
keeps them apart.

**The ruling to write an unrun spec is vindicated in the strongest available form.** It was written over
the reasonable objection that nothing executes it. Three hours later something did, and it immediately
produced the first real-browser proof that the footer drag works end to end — sign flip, proposal,
readout, snap, commit — *by way of* discovering that the test itself was lying. **A test nobody runs is
worthless until somebody runs it; the moment anybody does, part of what it buys is finding out the test
was wrong.**

---

## Owed after this gate, not blocking it

- **DW-226** `.page-seam` carries the identical specificity defect and has never rendered where placed.
  **Not cured by 12.5** — `.page-band > span` is byte-unchanged.
- **DW-227** an arrow nudge is inert whenever snapping is on, **for components as well as boundaries**;
  Shift steps 12pt, not the documented 10. Pre-existing product behaviour, not 12.5's.
- **DW-228** `isCanvas` pins contiguity but never anchors the band stack to the printable column.
- **DW-229** canvas gestures expose no accessible value semantics.
- **DW-193** nothing runs the e2e suite.
- `epic-12-retrospective` is `optional` per the tracker's own definitions.
