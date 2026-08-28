# Epic 5 Boundary Gate

**Date:** 2026-08-28  
**Baseline:** Story 5.12 commit `4bd6331d6d9bd0e0a83eaeea4a78d4658164772f`  
**Cadence:** D-000.4, per-epic boundary  
**Verdict:** **PASS**

This gate was run from the completed Story 5.12 code. It discharged the browser and cross-target
work deliberately compiled but deferred by Stories 5.1–5.12. Docker Desktop server `29.6.2` and
Node `v24.16.0` were available. No Epic 6 implementation was started before this verdict.

## 1. Real browser boundary

Exact final command, from `folio-designer/`:

```text
rtk env PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=/Users/panitw/Library/Caches/ms-playwright/chromium_headless_shell-1217/chrome-headless-shell-mac-arm64/chrome-headless-shell npm run test:e2e
```

The repository's Playwright 1.58.2 first requested browser revision 1208. A cache-only
`npx playwright install chromium` attempt produced no progress and was stopped after more than seven
minutes; it did not furnish evidence. The final run used the already installed compatible Chromium
headless-shell revision 1217 through an explicit executable path. The production web server built
the complete offline release before executing the suite.

Result: **14/14 scenarios passed** in 1.7 minutes, serially (`workers: 1`):

- application shell and responsive narrow canvas;
- current/stale production preview and page setup;
- component placement, drag/move/resize/delete, and mixed-property editing;
- real wasm worker page-setup command flow;
- complete offline reload;
- fallback open/download plus native-picker save and cancellation;
- successful waiting-worker update and failed-replacement recovery.

The first real-browser execution correctly exposed boundary defects that compiled tests could not.
They were fixed before the passing run:

- startup now remounts the application when the engine becomes ready instead of retaining the
  loading snapshot, and the footer receives the live offline state;
- file capability detection includes the browser's actual picker APIs;
- offline install timeout measures inactivity, resets on progress, clears on readiness, and does not
  force an update when no controlling worker exists;
- the service worker admits exact allowlisted same-origin static requests regardless of the browser's
  credential mode while precache fetches remain credential-independent;
- release payload validation accepts the current PDF.js-expanded 20-asset release under an explicit
  10–64 bound;
- the starter document supplies renderable font-family styles, and its specification example remains
  byte-identical to the golden;
- browser specs now use current accessible locators and exercise real generated service-worker/wasm
  replacement bytes instead of relying on interception that cannot affect worker fetches.

The final designer checks were also green: Vitest **23 files / 124 tests**, typecheck, e2e-source
compile, production build, and `git diff --check`. Lint passed with only the four pre-existing Fast
Refresh warnings.

## 2. Offline release gates

From `folio-designer/`, all three independent release witnesses passed:

| Command | Result |
|---|---|
| `npm run verify:offline` | **PASS** |
| `npm run verify:offline:red` | **PASS** |
| `npm run verify:offline:wasm` | **PASS** |

Together with the real-browser complete-offline reload and both update paths, these prove the
release manifest, cached bytes, wasm witness, mutation controls, installation lifecycle, and
replacement recovery rather than merely compiling their test source.

## 3. Four-target byte-identity matrix

Exact command, from `folio-go/`:

```text
rtk env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -tags=matrix -run TestCrossTargetByteIdentity -count=1 -v .
```

The real local `darwin/arm64` leg, Docker Linux `amd64` and `arm64` legs, and Node `js/wasm` leg
matched byte-for-byte for all 15 documents:

| Document | SHA-256 (all four targets) | Bytes |
|---|---|---:|
| `minimal-rect` | `0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c` | 547 |
| `font-text` | `a69a665331e7f0d31619f48179b54c7b9cb7a90ae013ed9c7c79daa128612181` | 22,315 |
| `image-embed` | `e5778eb872c98ec4a3c3c89466a8313cf52931b896701de8a43f3506abe689fc` | 995 |
| `multi-script-fallback` | `4699c8d710724ea544cc26bb3ee2b96af7a333f3dddd4462c0c846f7790480b0` | 55,086 |
| `shaped-text` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` | 91,059 |
| `wrapped-text` | `07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6` | 72,790 |
| `three-band-page` | `746efcbcfb5be30a06caaaefae25e3eaba1962c3fa47a74da10af6d0885372bf` | 54,452 |
| `multi-page` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` | 66,525 |
| `page-count-20` | `b32fa1c5babb8327b09b5c2bc0a11628b8c8885b9c5661c0262ec24920c5150f` | 72,659 |
| `hidden-image` | `902c4cdf66d88e0dc16f60a84a88b34e44dbf4dfbd9beafadbb6c564e758c9cd` | 10,381 |
| `statement-1` | `ef58bbf6dac1c3d4a5d679a77f9907a8d45f02ccd3f886c4d4e7cbdf9e86611d` | 76,740 |
| `statement-5` | `7f67b317c0a1925a404f8435bd4736b85e831a213f5a69fc2a2934a742ff950f` | 127,343 |
| `statement-20` | `be6f5e27af94e62e7c15a1814633cc48a2a91c5ee8686f5b76de5dc12e3cd4ed` | 269,804 |
| `statement-50` | `9c5be7ba7b4f31c7d488c114a377058ec30cec5ffca082d9c76ee26f304c754c` | 555,629 |
| `alternating-rows` | `e491d628ecd1dae9ad2d396341c014fb9dc5ce1e55c535a2f88bdae15b0e8bbd` | 55,734 |

`TestCrossTargetByteIdentity` passed in 25.34 seconds. The FMA discrimination control also passed:

```text
darwin/arm64  3c40a3d70a3d70a4
linux/amd64   3c50000000000000
linux/arm64   3c40a3d70a3d70a4
js/wasm       3c50000000000000
```

## 4. Ordinary gates

| Scope | Command/result |
|---|---|
| `folio-go` green suite | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` — **PASS** |
| ordinary build/vet | `go build ./...` and `go vet ./...` — **PASS** |
| tagged build/vet | `go build -tags=matrix ./...` and `go vet -tags=matrix ./...` — **PASS** |
| `hashmatrix` | test, vet, and build — **PASS** |
| `lint` | test, vet, and build — **PASS** |
| format/diff | scoped formatting clean; `git diff --check` clean |

The sanctioned corpus-floor red was rechecked independently and is unchanged:
`TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` reports `got 7, need >=20`; all statistics are
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. This remains the intentional D-000.17 /
D-000.57 / D-000.74 / D-2.1.14 red and is not an Epic 5 regression.

## 5. Story, contract, and deferred-work audit

- All twelve story files are `done` and now contain a plain-language opener. Story 5.11's missing
  opener was added during this audit; it changes no implementation contract.
- The tracker records every Story 5 item as done and this gate moves `epic-5` to `done`.
- The browser remains a projection/interaction layer: Go owns canonical `.folio` state, layout,
  measurement, diagnostics, revision, preview identity, and PDF bytes. The browser contributes
  local file transport, interaction, and rasterization only.
- Every Epic 5 Playwright deferral named under D-000.4 is discharged by the 14-scenario run above;
  the four-target deferral is discharged by the 15-document matrix. No additional Epic 5 deferred
  integration suite remains unrun.
- The worked example in `folio-format.md` and its golden JSON remain byte-identical and renderable
  through the production CLI with statement data/parameters.
- Pre-existing unrelated worktree changes were preserved and excluded: `_bmad/_config/files-manifest.csv`,
  `_bmad/_config/manifest.yaml`, `_bmad/bmm/config.yaml`, `_bmad/core/config.yaml`, `.agents/`, and
  `_bmad-output/planning-artifacts/research/`.

## 6. Disposition

Every Epic 5 boundary obligation passed. Epic 5 transitions to `done`; the intentional P6g red
remains explicitly disclosed; Epic 6 may begin. This artifact, tracker change, boundary-found fixes,
browser-test corrections, and worked-example repair are the complete scoped boundary change. The
final local commit hash is reported in the boundary handoff because a commit cannot contain its own
hash before it exists.
