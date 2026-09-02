---
title: 'Story 8.4d: The size budget is a number something checks'
type: 'feature'
created: '2026-09-02'
status: 'blocked'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
warnings: ['oversized']
deferred: []
---

## In plain terms…

The designer has to keep working when the network does not, so the first time you open it, everything it needs comes down in one go and stays on the machine. That download has been growing. Adding typefaces this month made it grow a lot, and nothing anywhere has ever said how big is too big — so it has been free to grow quietly, one story at a time, with nobody watching the total.

This story sets that limit and makes the build refuse to cross it. From here, growing the download past the agreed size stops the build, and raising the limit means editing the number on purpose, in a place where the edit is visible.

The limit proposed is about 15.8 megabytes. An estimate written down early on said about 9 megabytes, so the new figure is noticeably larger. That earlier number was a guess made in advance, not a target we aimed at and missed — it was written before anyone knew what the typeface library would hold, and the single largest item in the download, the Chinese-character face, accounts for most of the gap on its own. The size is what it is. What has been missing is anyone agreeing to it, and anything checking it.

<intent-contract>

## Intent

**Problem:** The first-load payload weighs 15,729,262 Brotli bytes and no gate compares it to anything. `epics.md`'s NFR7 records `~9 MB` as an accepted budget; the build has been ~1.75× that since before Epic 8, and every story that added weight added it silently. The measurement exists — Story 8.5 built it (`brotli.totalBytes`) and deliberately set no threshold — so this story's whole job is to choose the number and make something check it.

**Approach:** Declare the budget as a single integer literal in `src/release-payload.ts`, read it in `verify-offline-release.mjs` through the existing `readDeclaredConstant` regex derivation (the same mechanism 8.4f's `maximumCacheAssets` uses), and fail the build when `brotli.totalBytes` exceeds it. Supersede NFR7's `~9 MB` in place, with its history.

## Boundaries & Constraints

**Always:**
- The threshold is a **literal the check cannot derive from the payload it measures**. It is declared in `src/release-payload.ts` as `^const <name> = <digits>$` on its own line, and read by `readDeclaredConstant`, which throws when the constant is renamed, deleted, reformatted or commented out. A budget computed from what the build happens to weigh is not a budget (D-8.4i.11).
- The gate **fails the build**, never warns. It runs inside `verifyOfflineRelease`, which `npm run build` invokes.
- The gate consumes `brotli.totalBytes` — the figure Story 8.5 already built (D-8.5.6). It does not compute a second total.
- **Headroom is stated explicitly, with its reasoning, rather than tuned to fit** (D-8.5.1).
- Every recorded figure carries its command, commit, tree state and working directory (D-8.4j.8).
- `maximumCacheAssets` stays **64** (44 used, 43 immutable).
- Every gate invocation runs `-count=1` (D-000.11, the `folio-mvp` entry).

**Block If:**
- **The threshold value is not settled.** D-8.5.12 pre-committed that a threshold at this magnitude is escalated to the owner at this story's plan gate. See `## Auto Run Result`. The three options are not interchangeable: two of them change what this story builds. Do not pick one.
- The build turns out to be nondeterministic at one commit — that finding outranks the budget gate entirely and goes to the engineering lead (D-8.4.27c).
- A golden digest moves, or `brotli.catalogue.totalBytes` changes from 2,227,609.

**Never:**
- **No licence-gate work, for any reason** (D-000.11, flat prohibition). Register to Epic 15.
- No `.folio` change, no engine change, nothing under `folio-go/fonts/`. The 23 golden digests stay byte-identical.
- No catalogue change, no picking-behaviour change, no bold/italic (Epic 11).
- **No deferred-fetch tier unless the owner reverses D-8.5.12.** A build proposing one is a halt, not a design choice.
- `epics.md` may be amended **only** for the `~9 MB` figure, and only if the owner's answer requires it.
- Never `git add -A`; never push; never branch.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Under budget | `brotli.totalBytes` ≤ declared budget | `verifyOfflineRelease` returns; `npm run build` exits 0 | No error expected |
| Over budget | payload genuinely enlarged past the budget, release regenerated | Build fails on the budget's **own** message naming both figures | `fail(...)`, non-zero exit |
| Constant deleted/renamed/commented | `src/release-payload.ts` no longer declares the constant on its own line | Verification throws; build fails | `readDeclaredConstant` throws "does not declare … as a single live constant" |
| Two live declarations | the constant declared twice | Verification throws — authority is ambiguous | Same helper, "a second live declaration makes the authority ambiguous" |
| Budget below the floor | declared budget < `minimumCacheAssets`-era sanity, or ≤ 0 | Verification throws at load | Inversion/sanity check beside `declaredCacheAssetBounds`' existing one |

</intent-contract>

## Code Map

Anchors measured at `6f94972`, working directory `/Users/panitw/Projects/folio`, tree clean.

- `folio-designer/src/release-payload.ts` — declares the budget. `minimumCacheAssets = 10` at **:32**, `maximumCacheAssets = 64` at **:33**; the format-mandating comment is **:26–31**. The new constant goes beside them, same `const <name> = <digits>` shape. **Digits only** — the regex rejects `9_000_000` and `16 * 1024 * 1024`. The module has **no exported constants**; these are module-private and read from source text, not imported.
- `folio-designer/scripts/offline-release-contract.mjs` — the derivation. `releasePayloadSource` **:50**; `readDeclaredConstant(source, name, label)` **:66–70**, regex `^const ${name} = (\d+)$` (`gm`), **requires exactly one match** and throws on 0 or 2+; `declaredCacheAssetBounds` **:72**, reads the file only when `source` is undefined (**:73**) so tests inject a fixture string; inversion throw **:80**. Add `declaredPayloadBudget(source, label)` here in the same shape.
- `folio-designer/scripts/verify-offline-release.mjs` — the gate. `fail()` **:14**; import site **:10**. The 8.4f count bound is **:50–52**, with its placement rationale at **:35–49**. The aggregate Brotli block is **:137–153**: `immutable` bound **:141**, `catalogue` **:142**, **`brotli.totalBytes` proved equal to the per-asset arithmetic at :146**, catalogue subtotal **:153**. **The budget check goes immediately after :146** — after the total is proved to be real arithmetic, so the gate is not bounding a number the manifest merely asserts.
- `folio-designer/scripts/generate-offline-release.mjs` — the measurement, already built. `brotli` object **:117–128**: `immutableAssetCount` **:119**, **`totalBytes` :120** (`immutableAssets.reduce(… + asset.brotliBytes, 0)`), `catalogue.totalBytes` **:126**. Per-asset `brotliBytes` assigned **:101** from sidecar `statSync` at **:51**. **Comments at :103–107 and :123–125 explicitly reserve the threshold for this story** — "It is a MEASUREMENT, not a budget, and nothing in this repository compares it to a threshold." Those comments must be updated when the budget lands, or the file will contradict itself.
- `folio-designer/scripts/offline-release-contract.test.mjs` — the unit-test precedent. Vitest collects `scripts/**/*.test.mjs` (`vite.config.ts` `test.include`). This is where both-ways proof of the derivation lives, via fixture-string injection. `verify-offline-release.mjs` itself has **no** test file.
- `folio-designer/scripts/verify-offline-release.mjs` `redProof(name, mutate, expected)` **:263–274** — `expected` is the third argument and is what stops a proof passing on someone else's failure. `rewriteRelease` **:256–261** recomputes `pageId`/`id` and rewrites `sw.js`. The `dev-bypass-shipped` proof **:328–338** is the precedent for *mutate then regenerate*, which is the shape a byte budget needs.
- `_bmad-output/planning-artifacts/epics.md:150` — NFR7, the `~9 MB` itemisation. The only line this story may amend there. (`:3910` mentions `~9 MB` in an Epic 11 context — **out of scope**, do not touch.)

**Read-only evidence:** `src/` imports nothing from `scripts/` (the boundary is one-directional by construction), so the constant cannot be shared as a symbol — deriving it from source text is the established answer, not a workaround.

## Tasks & Acceptance

**Execution:** *(gated — the threshold value is unsettled; see `## Auto Run Result`)*

- `folio-designer/src/release-payload.ts` — declare the budget constant beside `maximumCacheAssets`, digits only, own line, with a comment recording that it was **chosen deliberately after a 1.75× overage went unnoticed**, not observed — because AC2 requires the declaring artifact to say so.
- `folio-designer/scripts/offline-release-contract.mjs` — add `declaredPayloadBudget(source, label)` reusing `readDeclaredConstant`; reject a non-positive budget the way `declaredCacheAssetBounds` rejects an inverted pair.
- `folio-designer/scripts/verify-offline-release.mjs` — read the budget and compare against `brotli.totalBytes` immediately after `:146`; `fail` with a message naming the measured total, the declared budget and the overage.
- `folio-designer/scripts/generate-offline-release.mjs` — update the `:103–107` and `:123–125` comments, which currently assert no threshold exists.
- `folio-designer/scripts/offline-release-contract.test.mjs` — unit-test `declaredPayloadBudget` both ways over injected fixture strings: a well-formed declaration parses; renamed/deleted/commented/reformatted/duplicated each throw with its own message.
- `folio-designer/scripts/verify-offline-release.mjs` (red proofs) — register `payload-over-budget`: genuinely enlarge the payload on disk, re-run `generateOfflineRelease`, assert the failure carries the budget's **own** `expected` substring. A manifest-only inflation trips `brotli-record-drift` at `:129`/`:146` first and proves nothing.
- `_bmad-output/planning-artifacts/epics.md:150` — supersede `~9 MB` **in place with its history**, only if the owner's answer requires it.

**Acceptance Criteria:**
- Given the release manifest, when `npm run build` runs, then a gate compares `brotli.totalBytes` against the declared literal and **fails** when it is exceeded — the number living in one place, read by the check rather than by a human.
- Given the declared figure, when it is set, then the artifact declaring it records that it was chosen deliberately after an unnoticed overage — not derived from whatever the build happened to weigh.
- Given a change that grows the payload past the figure, when the build runs, then the gate reds, and raising the figure is a visible, deliberate one-line edit.
- Given the `~9 MB` itemisation in `epics.md`, when the new figure is declared, then the old one is superseded in place with its history, so a later reader sees `~9 MB` was the original commitment and what replaced it.
- Given the budget constant is renamed, deleted, commented out or declared twice, when verification runs, then it **throws** rather than defaulting — a reader that defaults on a miss rebuilds the silent defect inside its own fix.

## Spec Change Log

## Review Triage Log

## Design Notes

**The measurement, re-taken at HEAD.** `npm run build`, then reading `dist/offline-release-manifest.json`; working directory `/Users/panitw/Projects/folio/folio-designer`; commit `6f94972`; tree clean before and after; node `v24.16.0`; exit 0.

| Figure | At `9e2792d` (D-8.5.20) | At `6f94972` (this gate) | Δ |
|---|---|---|---|
| `brotli.totalBytes` | 15,719,224 | **15,729,262** | **+10,038** |
| `brotli.catalogue.totalBytes` | 2,227,609 | **2,227,609** | **0** |
| `brotli.immutableAssetCount` | — | **43** | — |
| `s1.assetCount` | 44 | **44** | 0 |
| catalogue share | 14.17% | **14.16%** | — |

**Disposition of the delta (D-8.4.27b: explained, not merely reproducible).** The +10,038 bytes are attributable to Story 8.6, the only work between the two commits: it changed `folio-go/component_commands.go` (+305), `internal/template/parse.go` (+82), `model.go` (+26) and `serialize.go` (+7) — which move the engine wasm — plus `src/App.tsx` (+101) and `src/font-chain-command.ts` (+55), which move the app bundle. **The catalogue subtotal is byte-identical**, confirming no font bytes moved. Since Story 8.4g the wasm build passes `-buildvcs=false` and is deterministic, so this is a real source delta, not the measurement drift DW-100 chased. **DW-100's three conflicting `s1VisibleBytes` readings were a pre-8.4g artefact and are closed by 8.4g, not by this story** — and note `s1VisibleBytes` sums four hardcoded needles, so it was never the instrument for a payload budget (D-8.5.6).

**43 vs 44 — two populations, do not conflate.** `brotli.immutableAssetCount` is 43; `s1.assetCount` is 44. The difference is `/index.html`, which is mutable, carries no Brotli sidecar, and is rejected if it ever declares `brotliBytes` (`verify-offline-release.mjs:133–134`). **The budget bounds the 43 immutable assets.** `/index.html` is ~a few KB and self-referentially sized by the generator's fixed-point loop, so excluding it is right — including it would make the budget depend on a value that converges rather than one that is measured.

**Headroom, stated rather than tuned.** Measured catalogue faces span **11,862** bytes (Noto Serif Thai) to **180,419** (Noto Serif), mean **106,077**. Story 8.6 — a whole feature story — moved the total by **10,038**. A proposed budget of **15,800,000** gives **70,738 bytes (0.45%)** of headroom, which is:
- **below the mean catalogue family, and below 15 of the 21 faces individually** (measured), so adding a typical family reds the gate;
- **~7× a full feature story's engine growth**, so ordinary work does not flap it.

Stated honestly: **six of the 21 faces are small enough to be added individually without tripping it** — 11,862 / 14,757 / 47,420 / 50,981 / 53,854 / 57,766 bytes. No single number does both jobs; a headroom tight enough to catch a 12 KB face would red on a Go toolchain bump. That limit is recorded rather than papered over, and it is the strongest argument for a *lower* budget if the owner wants one: the gate's sensitivity is set by the headroom, not by the total.

**Why the gate cannot derive its own bound.** `src/` imports nothing from `scripts/`, and `npm run build` does **not** run Vitest — so "duplicate the constant plus a tie test" leaves the build passing over drift, because the tie only reddens in a suite the build never consults. Deriving from source text and throwing on a miss is the idiom already in this file, and it is the reason the constant is a literal rather than a computed allowance.

## Verification

Cadence **per-epic** (D-000.4). Every invocation records its working directory. Every Go gate runs `-count=1` (D-000.11).

**Commands:**
- `npm run build` (wd `folio-designer`) — expect exit 0, budget gate green.
- `npm run verify:offline:red` (wd `folio-designer`) — expect exit 0; the new `payload-over-budget` proof must fail on **its own** message. Capture the exit code by redirect, never after a pipe.
- `npm run verify:offline:wasm`, `npm test`, `npm run typecheck`, `npm run lint`, `npm run test:e2e:compile` (wd `folio-designer`) — lint expect **exactly 4** `only-export-components`; Vitest expect ≥ 42 files / 432 tests.
- `go test -count=1 ./...` and `go vet ./...` (wd `folio-go`) — expect **1877 pass / 2 fail / 5 skip**; the two are `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`, standing reds by identity.
- `go test -count=1 ./...` (wd `lint`) — expect four `ok`. Cached `ok` is not evidence.
- `gofmt -l folio-go lint` **from the repo root** — expect exactly `lint/internal/rules/licencegraph_test.go`; **do not reformat it**. From `lint/` this prints `lstat` errors that read as clean.
- `shasum -a 256 fixtures/*/expected.pdf` **from the repo root** — expect **23** digests, byte-identical to baseline. A moved golden is a HALT.

**Manual checks:**
- `dist/offline-release-manifest.json`: `brotli.catalogue.totalBytes` still exactly `2,227,609`; `maximumCacheAssets` still `64`.
- Excluded by the epic boundary gate: the four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity`, Playwright.

## Auto Run Result

Status: blocked
Blocking condition: intent gap

**The threshold value is an owner decision that D-8.5.12 pre-committed to this gate**, and the three
options are not interchangeable — two of them change what this story builds. Nothing in the intent
selects between them, so the spec does not pick one.

**What the owner is being asked.** The finished payload is **15,729,262 Brotli bytes** — **1.75×** the
`~9 MB` NFR7 records. D-8.5.12 priced this at "roughly double" and said the owner had priced 37%, not
2×. Measured, it is 1.75×, not 2× — the estimate was high — but the substance of the trigger holds:
this is decisively past the 37% the owner priced when they chose 20+ families, so the escalation fires.

**Options, with costs.**

**(a) Accept the weight; set the budget at 15,800,000 with 70,738 bytes (0.45%) of stated headroom;
supersede `epics.md:150`'s `~9 MB` in place with its history.** Cost: the stated NFR7 commitment goes
from ~9 MB to ~15.8 MB, and that is a real change to a promise, not a documentation fix. Nothing else
moves; no story is re-opened; the gate lands this dispatch.

**(b) Reduce the catalogue.** The 21 faces are 2,227,609 bytes — **14.16%** of the payload. Deleting
*all* of them still leaves 13,501,653 bytes, **1.5× the ~9 MB**, because the engine wasm (7,231,165)
and the CJK face (4,948,312) are 77% of the payload on their own. Cost: real loss of author choice for
a change that cannot reach the stated commitment. **This option cannot achieve what it appears to
achieve**, and that is the load-bearing fact.

**(c) Reverse D-8.5.12 and add a deferred-fetch tier.** Buildable today — the mechanism was checked and
is available; it is foreclosed by policy, not physics. Reversing means amending `SPEC.md`
`## Non-goals` **and** `font-catalogue.md` together. Cost: a family not yet fetched cannot be picked
while offline. **The mitigant must be stated or this reads worse than it is: the catalogue is a
palette, not coverage** — the three shipped Noto faces are the coverage — so an unfetched family
degrades to "you cannot pick that family right now", never to a document that will not render. It
would take ~2.2 MB out of first load and re-opens Story 8.5's shipped shape.

**Recommendation: (a).** Option (b) cannot reach the target it exists to reach — the arithmetic above
is decisive, and cutting author choice for a 14% dent in a 75% overage is the worst trade of the three.
Option (c) is a genuine architectural choice, but it is Epic 8's *last* story and (c) re-opens the epic's
shipped shape for a gain smaller than the CJK face alone; it belongs in a release-scoped conversation
with the offline story intact, not as a threshold's side effect. **(a) does the thing the owner already
chose** at D-8.4.24 — pick a figure and make it enforceable — and it stops the drift, which is the
failure that actually occurred. The honest cost, stated plainly: the number written down goes up, and
this is the moment it stops being free to go up again.
