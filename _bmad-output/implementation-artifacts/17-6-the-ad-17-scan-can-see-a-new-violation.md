---
title: 'Story 17.6: The AD-17 scan can see a new violation'
type: 'refactor'
created: '2026-09-05'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: '3d2e3f7'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-11-14-decision-log.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

One rule in this product says the browser never measures — it draws what the engine tells it and asks the
engine for every text metric. A test scans the whole designer for the ways code could break that rule, and
it has been failing for weeks on a single old spec file.

A failing alarm cannot get louder. While that scan is red, a *new* violation anywhere in the designer
changes what the failure says and not whether it fails, and nothing reads what it says. Fifteen stories of
new designer work are about to be written straight into the area that scan is supposed to police.

This story makes the alarm work again. The one flagged spec is a browser test that reads the borders the
page actually painted — an instrument measuring the product's output, which is the opposite of the product
measuring itself. It gets a narrow, named exemption of exactly the kind three other places in this file
already have, and then the scan is proved to still catch a new violation by planting one.

What this story does not do is loosen the rule. The set of files scanned does not shrink, no folder is
waved through, and the exemption covers one spelling in one named block that the test itself checks is
still there.

<intent-contract>

## Intent

**Problem:** `canvas-authority-contract.test.ts`'s corpus scan has one standing violation —
`e2e/e9-5-border-no-ink.spec.ts`'s `getComputedStyle`. Because the assertion is
`expect(violations(...)).toEqual([])`, a second violation anywhere in `src/` or `e2e/` changes the
failure's **contents** and not its **status**, and **no gate reads contents** (DW-152). The CI job that
runs it is quarantined as `folio-designer-known-red` (DW-171's repair), which unblocked the six steps
behind it but explicitly did not fix the masking.

**Approach:** Grant the flagged assertion a named-owner exception in the shape this file already uses three
times, restore the scan to green, and prove by planting a violation that it can fail again.

## Boundaries & Constraints

**Always:**
- **THE EXCEPTION IS A SPELLING REWRITE OVER ONE NAMED OWNER, NOT A PASS.** Scoped to
  `e9-5-border-no-ink.spec.ts` by basename and to the one `page.evaluate` block that reads border ink;
  it rewrites **only** the `getComputedStyle` spelling. **Never a directory-wide `e2e/**` waiver, never an
  allowlist of files, never a change to `prohibited`.**
- **THE SCAN'S POPULATION MUST NOT SHRINK.** `production`, `tests` and `e2e` keep their existing
  constructions. The non-vacuity floors (`production.length > 10`, `tests.length > 10`, `e2e.length > 3`)
  stay and are the guard against a population that quietly empties.
- **THE EXCEPTION MUST ASSERT ITS OWN REASON.** Both sibling helpers `expect(...)` that the seam earning
  the carve-out is still present, so the exception dies with the thing it exempts. **Copy that.**
- **PROVE THE SCAN CAN FAIL AGAIN, BY PLANTING ONE.** A scan that has never been observed failing since
  the repair has not been shown to work. Plant a violation in **production** source, watch the scan red,
  restore by absolute path, watch it green. **Report where the failure landed.**
- **DELETE `folio-designer-known-red` FROM `ci.yml` WHEN THE SCAN IS GREEN**, and take
  `DESIGNER_KNOWN_RED` and the green step's `-t` filter with it, so `npm test` runs whole again.
  **`folio-go-known-red` is a DIFFERENT THING and is out of scope** — it reports an honestly unmet exercise
  floor (D-000.17 / D-2.1.14 / D-000.57) and is never to be "fixed".
- **DW-152 and DW-171 are both discharged by this story** and must be marked so, with this story named.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Any change to the `prohibited` array, or to how `scanned`/`violations` compose.
- Removing or rewriting the e2e assertion itself (see Never).
- Touching `folio-go-known-red`, or any other `ci.yml` job.

**Never:** rewriting the flagged assertion so it reads the app's own declaration back to itself — its value
is that it compares **resolved** border ink against an expected set, and making both sides move together
would let it pass through the very defect it was written for · a directory-wide exemption · a shrunk
population · a green claim not backed by a planted-violation proof.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| The scan, after the exception | Repo as committed | `violations([...])` is `[]` and the test passes | — |
| A NEW `getComputedStyle` in production | Planted | **Scan goes red** naming that file | The planted-violation proof |
| A NEW `getComputedStyle` in another e2e spec | Planted | **Scan goes red** — the exception is one file, not the folder | — |
| A second `getComputedStyle` in the exempt file, outside the named block | Planted | **Scan goes red** — the exception is one block, not the file | — |
| The exempt seam is renamed or deleted | Simulated | **The exception's own `expect` fails**, so the carve-out cannot outlive its reason | — |
| Every other prohibition inside the exempt file | e.g. `offsetWidth` planted there | **Still red** — only the one spelling is rewritten | — |
| The population | — | `production`, `tests`, `e2e` unchanged; the three non-vacuity floors still pass | — |
| `npm test` whole | No `-t` filter | Exit 0; **no test is excluded by name any more** | — |
| `ci.yml` | After the change | No `folio-designer-known-red` job; no `DESIGNER_KNOWN_RED`; `folio-go-known-red` untouched | — |

</intent-contract>

## Code Map

**Anchors are SYMBOL + COUNT per D-000.4; re-derive the line at implementation.** Measured at `3d2e3f7`.

**The scan**
- `canvas-authority-contract.test.ts` — `const prohibited` (**1**, ~:18), the pattern array. **Do not edit.**
- `function scanned` (**1**, ~:173) — composes the exceptions; `function violations` (**3** occurrences,
  definition ~:177) — applies `prohibited` to the scanned text.
- `function withoutComments` (**1**, ~:131) — comments are stripped before scanning, so prose about
  `getComputedStyle` is already safe.
- `withoutApprovedRuntimeFaceRegistration` (**2**) and `withoutApprovedLocalPointerInput` (**3**) — **the
  two shapes to copy.** Each is scoped by `path.basename` or a path segment, each `expect`s its seam is
  present, each rewrites only named spellings.
- `function violationsForFile` (**1**, ~:306) — *"The same scan a real file gets, addressed by NAME, so an
  exception that is scoped to a file can be proved to hold there and to hold nowhere else."* **This is the
  harness the matrix rows are written against; use it rather than inventing one.**

**The population** — three `readdirSync` constructions near the file head, for `production`, `e2e` and
`tests`. **D-000.6 records that the `e2e` arm is a construction default, not a recorded ruling**, which is
why narrowing by exception reopens no prior decision.

**The violation**
- `e2e/e9-5-border-no-ink.spec.ts` — one `getComputedStyle(box)` (**1**), inside the `page.evaluate` in
  `expect.poll` that maps `.canvas-box` to `owner :: side=colour`. **The seam to exempt and to assert.**

**CI**
- `.github/workflows/ci.yml` — `DESIGNER_KNOWN_RED` in `env`, the `folio-designer` job's green test step
  carrying the negative-lookahead `-t`, and the whole `folio-designer-known-red` job. **All three go.**

**The register**
- `deferred-work.md` — DW-152 (the masking) and DW-171 (the halted job). Both discharged here.

## Tasks & Acceptance

**Execution:**
- [ ] Add the named-owner exception, in the sibling shape, asserting its own seam.
- [ ] **Plant a violation and prove the scan reds**, in production and in a second e2e file; restore by
      absolute path. **Say where each landed.**
- [ ] Restore `npm test` to running whole: remove the `-t` filter, the env var and the quarantined job.
- [ ] Discharge DW-152 and DW-171, naming this story.
- [ ] Tests — one per matrix row (**9 rows**), driven through `violationsForFile`.

**Acceptance Criteria:**
- Given the repo as committed, when `npm test` runs **with no name filter**, then it exits 0.
- Given a `getComputedStyle` planted in production, when the scan runs, then it **fails naming that file**.
- Given one planted in a different e2e spec, when the scan runs, then it **fails** — the exception is one
  file.
- Given the exempt seam removed, when the scan runs, then the **exception's own assertion** fails.

## Design Notes

**Why an exception and not a repair.** AD-17's subject is the **product**: the canvas gets every text
metric from the engine. A Playwright assertion that reads what the browser actually painted is an
**instrument measuring the output** — the opposite of the failure AD-17 exists to prevent. Rewriting it to
read the projection's declared border instead would make both sides of the assertion move together, and it
would pass through the very defect it was written for.

**Why this is Wave A and not later.** The scan polices exactly the population Epics 13 and 14 are about to
add fifteen stories to. Every story landed under a red scan is a story whose AD-17 compliance was never
actually gated.

**The allowlist lesson applies here too.** `prohibited` is a denylist, and it is not this story's job to
fix that — but note in the report that a guard enumerating what is forbidden cannot see what nobody thought
to forbid. This run has two independent instances already (17.5's `border-top`/`box-shadow`; 11.3's implied
`font-weight` denylist that cannot see `transform: skewX()`).

## Verification

**Commands** — one per line, **exit codes read from `$?` immediately, never through a pipe**; zsh, so
`${PIPESTATUS[0]}` is wrong, and beware a `cd` earlier in a compound command silently re-rooting a later
relative path.
- `cd folio-designer && npm test` — **this story's whole point is that this now exits 0 with no filter.**
  The count must not fall.
- `npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`. **No Go is touched; say so rather than running it.**
- `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` — **ci.yml must still
  parse.** This file was once unparseable and therefore never ran at all (D-000.71).
- **Reproduce both designer CI jobs locally** as the workflow would run them, and report that the
  quarantined one no longer exists rather than that it passes.

**No browser run is required** — this story changes no rendered output.

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**9 rows**); state
the population beside every zero and pair every absence claim with a positive control; **a comment is not a
measurement**; **use `/usr/bin/grep -a`** — recursive `grep` returns false zeros here and may treat long
files as binary.
