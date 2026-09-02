# Epic 8 — Boundary Gate

**Run 2026-09-02 at `f2db371`, branch `main`, tree clean before and after. VERDICT: PASS.**

Nothing was committed, pushed, branched or fixed by the gate. Go 1.26.0 darwin/arm64, node v24.16.0.

## The three owed suites — all three RAN

| Suite | Command | Working dir | Result |
|---|---|---|---|
| Four `FOLIO_MATRIX_TARGET` legs | `FOLIO_MATRIX_TARGET=<t> go test -count=1 -tags=matrix -run TestTargetRenderHash -v .` | `folio-go` | **PASS ×4** — darwin/arm64 1.40s, linux/amd64 9.45s, linux/arm64 6.37s, js/wasm 12.33s |
| `TestCrossTargetByteIdentity` | `go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` | `folio-go` | **PASS (23.78s)** |
| **Playwright** | `npm run test:e2e` | `folio-designer` | **EXECUTED AND GREEN — 24 passed (2.6m)** |

**Legs proved non-vacuous:** `grep -c "asserts NOTHING"` = **0** on all four. The unset control ran
separately and *did* log the no-op line — a control, not a fifth leg.

## Playwright — EXECUTED AND GREEN, with its deviation stated

**A real browser opened.** 24 tests across all 12 spec files, including the service-worker offline
reload and the full `browser-native-roundtrip` (14.9s). **This is the first execution in the
repository's history** (D-8.4.25d's assertion had been *owed, not attempted-and-passed*, since Epic 8's
middle).

**Why the pinned install cannot complete — measured, not inferred.** A **30-hour-hung** install from
Sep 1 (PID 94809, 1.53s CPU, zero network sockets) held `~/Library/Caches/ms-playwright/__dirlock`. A
fresh `npx playwright install chromium` **re-created the identical 428 KB stub** and stalled again
(8:49 elapsed, 0.78s CPU, no download child). **Root cause:**
`cdn.playwright.dev/.../chromium/1208/chromium-mac-arm64.zip` 307-redirects to
`playwright.download.prss.microsoft.com`, which returns **HTTP 400, body `GatewayExceptionResponse`,
24 bytes**, on both HEAD and ranged GET. **The download host refuses the asset.**

**The execution path:** `playwright.config.ts` carries an official hatch —
`launchOptions.executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH`. A genuine 344 MB
**Chrome for Testing 149.0.7827.55** (revision 1228) was already on disk and the suite ran against it.

**THE DEVIATION, stated rather than smoothed:** it executed against **Chrome 149 (rev 1228), not the
pinned 145.0.7632.6 (rev 1208)**, because 1208 is undownloadable. **The pinned-revision leg remains
undischarged, and behaviour under Chrome 145 is a COULD NOT LOOK.**

## Failures — three standing reds by identity, ZERO new

1. `folio-go` **1877 pass / 2 fail / 5 skip** — sole failure `TestCorpusMeetsP6ExerciseFloors` +
   `P6g_(opaque_names)`, *"floor not met: got 7, need >=20"*, mandated unmet by D-000.17 / D-2.1.14.
2. `gofmt -l folio-go lint` **from the repo root** → `lint/internal/rules/licencegraph_test.go`, that
   file only.
3. designer `npm run lint` → **exactly 4** `only-export-components`.

`lint` module `go test ./...` — 4 packages ok.

**`genmanifest` proved non-vacuously, and the trap fired once during the proof.** Run from inside
`lint/`, then `git diff --exit-code` from the **repo root**. The first attempt ran
`git ls-files --error-unmatch lint/MANIFEST.md` while still in `lint/`; it failed, and the `&&` chain
**short-circuited so the diff never ran**. Re-run from root: pathspec resolves, diff empty, exit 0. **A
real pass.**

## FINDING — a stale evidence artifact the never-run leg was hiding

The Playwright run left one tracked file modified: `evidence/story-6.7-roundtrip-manifest.json`,
**+182/−102**. Restored by the gate.

The spec **writes** this manifest and **never asserts against it** — its assertions are deliberately
loose (`toBeGreaterThanOrEqual(8)`), so **the drift cannot redden anything.**

**Not owned by 8.5 or 8.6, and that is PROVED.** The entire drift is 11 added `fontFamily` set
commands; `grep -icE "licence|license|embedFace|catalogue|typeface|fontAsset"` over the added lines =
**0**. `git blame` puts the `setFontFamily` helper at **44121b88 (2026-08-29, Story 6.7)** — days
before either story.

**The implication.** Commit `6c79d29` describes itself as *"the re-run 6.7 manifest"* — **but Playwright
had never executed, so that file cannot have come from a real run.** The first genuine execution in the
repository's history exposed it immediately. **This is precisely the class of rot a never-executed leg
conceals.**

**And a second consequence:** any future Playwright run **dirties the tree** by writing a tracked file,
which will fight commit hygiene and the dirty-tree halt at every subsequent dispatch.

## Goldens and matrix agreement

**All 23 golden digests present and verified.** **The matrix legs agree across all four targets** —
every one of the 21 matrix documents produced a byte-identical sha256 and byte length on darwin/arm64,
linux/amd64, linux/arm64 and js/wasm, each matching its golden (`embedded-font` `f533b04b…` 3225 B;
`thai-stacked-marks` `d5077f33…` 65257 B; `alignment-rounding` `986400a1…` 61346 B). The 23 reconcile
as 21 matrix documents − `hidden-image` (pins `expected.json`, not a PDF) + `font-text`, `image-embed`,
`minimal-rect`.

## Offline verification and the cache bound

`npm run build` (0), `verify:offline` (0), `verify:offline:red` (0), `verify:offline:wasm` (0).
**`maximumCacheAssets` = 64**; **44 assets** in the built release, 20 under the bound. Story 8.5's
faces visible as `catalogue-*.ttf`.

## COULD NOT LOOK

- **Chrome 145 (rev 1208) behaviour** — the download host refuses the archive.
- **Real linux/amd64, linux/arm64 and js/wasm hardware** — the legs are cross-compiled and
  hash-compared from a darwin/arm64 host, as designed.
- **Story 8.4d and the catalogue fetch tier** — moved to Epic 15; out of this gate's scope.

## Environment changes outside the repository (none in the working tree)

Killed a 30-hour-hung Playwright install process tree and removed the stale
`~/Library/Caches/ms-playwright/__dirlock` it left behind — the remedy Playwright itself prints. No
source, config or committed file modified.
