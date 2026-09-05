# Epic 16 — Boundary Gate

**Run 2026-09-05 at `bcb6ebb`, branch `main`, tree clean before and after. VERDICT: CLOSE, with one
carried failure. Not a pass.**

Go 1.26.0 darwin/arm64, node v24.16.0, `@playwright/test` **1.63.0**, Python 3.12.13 + fontTools 4.63.0 in
a pinned venv. Nothing was pushed or branched by the gate. The gate fixed no code.

**Stated deviation, first, because it conditions everything below.** This gate runs at `bcb6ebb`, a tree
that has moved **past** Epic 16 — it carries Stories 15.2a, 16.11, 12.4 and 12.1. Epic 16's own stories
were all `done` before any of those landed, and the gate was owed then and not run. So these measurements
certify the current tree, not Epic 16 in isolation. Where a finding is attributable to Epic 16 it is said
so; where it is pre-existing or later, that is said too. **A gate run late is weaker evidence than a gate
run on time, and this one says so rather than implying otherwise.**

## Why this gate was owed, and what skipping it cost

Under D-000.1(3) the heavy suites move to the epic boundary. **Epic 16 closed every story without one.**
The bill came due four days later:

- **Story 16.8 added Roboto to `fonts.Shipped()`** and renamed `starter.folio`'s only chain from `body` to
  `Roboto`, making a catalogue family both document-declared and the default typeface.
- **Four things closed over the old state.** One failed loudly and was found in four days
  (`matrix_test.go`'s fixture guards, fixed at `28cd225`). One failed **silently** and nothing was going
  to find it (`fontgen`'s `3 of 3`). Two were **invisible**, because the suite that would have caught them
  is compiled in CI and never executed.
- **The `Cross-target byte identity` workflow was red on all four targets for four days**, and
  `ci.yml` had **never been green in its entire history** — 12 of 12 runs `failure` from 2026-08-26 —
  because a quarantine job took an expected-red test's exit code as its own verdict.

## The matrix leg — RAN UNFILTERED, and green

D-000.11 requires the **unfiltered** tagged suite, not CI's two name-filtered tests. The tag covers
**7 files and 12 test functions**; `matrix.yml` runs **2 of the 12** per leg. The other ten run in no gate
at all — and one of them is where the vacuous `3 of 3` hid.

| command | result |
|---|---|
| `go test -count=1 -tags=matrix ./...` (all 12) | rc 1 — **only** `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)` |
| `FOLIO_MATRIX_TARGET=darwin/arm64` render+probe | rc 0 |
| `FOLIO_MATRIX_TARGET=linux/amd64` render+probe | rc 0 |
| `FOLIO_MATRIX_TARGET=linux/arm64` render+probe | rc 0 |
| `FOLIO_MATRIX_TARGET=js/wasm` render+probe | rc 0 |
| `TestCrossTargetByteIdentity` | rc 0 |

`TestShippedFacesReproduceFromUpstream` — one of the ten CI never runs — now reports **`derived and
compared 3 of 4 shipped faces (Noto Sans, Noto Sans SC, Noto Sans Thai); 1 accounted static-upstream`**,
where before Story 16.11 it reported `3 of 3` from its own manifest while four faces shipped.

**Standing property of this gate, recorded so it is not rediscovered:** that test needs Python 3.12.13 and
fontTools 4.63.0 via `FOLIO_FONTGEN_PYTHON`. **CI does not have that toolchain and cannot acquire it.**
That is a second argument for the accounting half being toolchain-free, which it now is.

## The e2e leg — EXECUTED, 32 of 33

**This suite had never been executed in this programme.** DW-193 records that CI compiles it and never
runs it. Getting it to start took three attempts and roughly two hours of environment work — a pinned
Chromium build whose headless shell would not download and whose cached full browser was broken. The pin
was dropped on the owner's instruction (`@playwright/test` 1.58.2 → 1.63.0, committed deliberately at
`71627a5`), after which the headless shell fetched in under a minute.

| | |
|---|---|
| **PASS** | **32** |
| **FAIL** | **1** — `browser-native-roundtrip.spec.ts` › *fresh authored sessions close exactly through admitted Preview and native Folio* |

**Its first run found two failures that 804 passing unit tests and every CI gate were blind to**, both in
`font-embed-boundary.spec.ts`, both stale premises from 16.8's starter rename, both repaired by Story
16.11 and both green here. **Neither was a product defect** — an earlier report of mine that the default
typeface could not be chosen was wrong, and the record says so.

## The carried failure — CARRIED, not this gate's, and characterised

`browser-native-roundtrip.spec.ts` **hangs**. Measured rather than inferred:

| budget | result |
|---|---|
| 300 s (shipped) | FAIL, timeout |
| 1500 s (5×) | FAIL, timeout |

**The resource profile names the mechanism: 152 s user + 5 s system across 27 minutes of wall clock — 9%
CPU.** A slow test burns CPU; this one **waits**. The captured snapshot places it mid-authoring, in DESIGN
mode with unsaved changes, before the save/render comparison it exists to perform.

**What is NOT known, stated so the next reader does not over-read this: where it blocks.** The trace was
not decomposed to the failing action and no bisection of the authoring helpers was run. *"It hangs"* is
measured; *"it hangs at X"* is not.

**Why it matters more than one red test.** It is the heaviest assertion in the repository — it builds the
Go CLI, authors two full documents through the real UI, saves both, renders each in the browser **and**
natively, compares the PDFs byte-for-byte, and runs a Go witness test. **It is the only thing that proves
the browser and the native binary agree on a document a human actually authored.** While it hangs, that
guarantee is unverified.

**Not Epic 16's doing** — it is not in the blast radius of the shipped-set or starter changes, and it
predates them. **Carried, filed, and owed an owner.**

## The rest of the set, for the record

| working directory | command | result |
|---|---|---|
| `folio-go` | `go test -count=1 ./...` | rc 1 — 1996 pass / 2 fail / 5 skip, the two being the sanctioned red |
| `folio-go` | `go vet ./...` · `gofmt -l .` | rc 0 · empty |
| `folio-designer` | `npx vitest run` | rc 0 — 59 files / 819 tests |
| `folio-designer` | `npx tsc -b` and `tsc -b --force` | rc 0 · rc 0, **zero** errors |
| `folio-designer` | `npx oxlint` | rc 0, exactly 4 `only-export-components` |
| `lint` | `go build` · `go vet` · `gofmt -l` · `go test` | 0 · 0 · 0 · 0, four separate commands |

**`npx tsc --noEmit` is NOT in that table, deliberately.** It was the stated typecheck in every dispatch
of this run and it **measures nothing** — `tsconfig.json` is a solution file, so it checks zero files and
exits 0. Red-proved: with a deliberate type error planted, `--noEmit` returns 0 and `tsc -b` returns 2.
Filed as **DW-207**. A forced full rebuild confirms the designer is genuinely type-clean, so nothing
accumulated — `npm run build` runs `tsc -b` independently, which is the only reason.

## COULD NOT LOOK

- **Where the roundtrip test blocks.** Trace captured (42 MB) but not decomposed.
- **Whether the hang predates `@playwright/test` 1.63.0.** It reproduces standalone under 1.63 at both
  budgets; it was not re-run under 1.58, whose browser could not be fetched at all.
- **Epic 16 in isolation.** See the stated deviation.

## What this gate found that no per-story gate could

Every defect above is a **closure over something that moved** — and each was invisible to the story that
moved it, because the guard lived in a suite that story's gates did not run. That is the argument for the
boundary cadence, and the cost of skipping it here was four days of a red byte-identity workflow, a
vacuous reproducibility guard, two stale assertions, and a main CI workflow that had never once passed.

**Seven guards in this programme have now been found unable to distinguish a correct outcome from a
plausible wrong one** — the matrix fixture guards, `fontgen`'s manifest denominator, the CI quarantine
job, a denylist scan, Story 12.4's rotation-blind table assertion, Story 12.1's untested `pageFooter`
arm, and `tsc --noEmit`. Six were caught by review or by a later story; **one was caught by this gate.**

## VERDICT

**CLOSE Epic 16, with one carried failure. This is not a pass.**

The three findings the gate owed are discharged: the accounting guard was vacuous (**fixed**, 16.11), two
e2e assertions had gone stale (**fixed**, 16.11), and the deepest cross-boundary test in the repository is
hung (**characterised, carried, filed**). The repair the lead required before Epic 11's first face —
an untagged accounting guard total over `fonts.Shipped()` with no fallback arm — exists and is green.

Epic 11 is unblocked. The carried failure blocks nothing in Epics 11–15 and is owned by its register
entry.
