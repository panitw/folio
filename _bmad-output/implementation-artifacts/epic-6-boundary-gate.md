# Epic 6 Boundary Gate

**Date:** 2026-08-29  
**Baseline:** Story 6.7 commit `44121b8855f42eebaa7000fdd1bfaa9c7968401a`  
**Cadence:** D-000.4, per-epic boundary  
**Verdict:** **PASS** — *at the stated baseline only. See the amendment immediately below.*

## Amendment — 2026-08-29: this gate describes `44121b8`, not HEAD

**Owner decision, recorded as D-5.13.6.** Two commits landed on `designer-inspector-tabs` after this
gate passed, with no story key and no story file. This gate's evidence therefore no longer describes
the tree anyone will run. `epic-6` stays `done` — it was written correctly at the time — but the
delta below is open, named debt, and it is re-measured at the **second Epic 5 boundary gate** owed by
Story 5.13's reopening of Epic 5, rather than by re-running this gate now.

The post-gate delta, precisely:

- **`eef7fbb` changed canonical-byte-producing engine code.** `createComponentInBand`
  (`folio-go/component_commands.go`) now adopts the first declared non-empty font chain in sorted key
  order onto every palette-placed text element, so **the same author action serializes different
  `.folio` bytes than §3's matrix measured**. Any golden that records a palette-placed text element
  has moved; a hash change there must be investigated as an intended versioned change under AD-22,
  never regenerated on sight.
- **`eef7fbb` changed Story 6.6's diagnostic contract surface.** `folio-go/wasm/cmd/engine/main.go`
  replaced the fixed `ENGINE_REJECTED` / render-failure text with the engine's own bounded message.
- **`eef7fbb` carries in-progress owner dev-bypass work** in `offline-lifecycle.ts`, `main.tsx` and
  `verify-offline-release.mjs`, letting the dev server start the engine with **no release payload**.
  This is named here deliberately: it is a mechanism that stays invisible until it is load-bearing,
  and it currently exists only as a clause in a commit message. §5's offline-release evidence does
  not cover it.
- **`3a52ae4` edited five of the eleven Playwright specs** whose green §2 recorded, plus `App.tsx`,
  `App.css` and `DataPanel.tsx`.

Consequence for the next gate: a red second Epic 5 boundary needs a **two-way** bisect — across Story
5.13 and across these two unstoried commits — which is the attribution cost this option accepted.

This gate was run from the completed Story 6.7 code. It discharges the real-browser and four-target
work deliberately deferred by Stories 6.1–6.6 and independently repeats Story 6.7's integration-shaped
round-trip evidence. No release tag was cut.

## 1. Story and record audit

- Stories 6.1–6.7 are committed `done`, contain the required plain-terms opener, have completed task
  lists, and retain their creation, implementation/review, finisher and measured-verification records.
- All persisted review findings are resolved. Story 6.6's four findings were explicitly resolved by
  its finisher but their checkboxes remained stale; this boundary corrects those four record markers
  without changing product code.
- D-6.2.1 remains the sole Epic 6 engineering ruling and the shipped sample-independent binding command,
  browser discovery affordance and AD-14 runtime-kind diagnostics agree with it.
- DW-17 is discharged: the CLI, successful Preview diagnostics, failed-render presentation, and the
  Story 6.7 dynamic diagnostic census now cover its three named presentation owners. DW-11 remains open
  and visible as the sanctioned P6g corpus-floor red. DW-4's release checkpoint and DW-20's pre-tag
  obligation remain open; this gate therefore creates no `folio-go/v0.1.0` tag.
- The committed Story 6.7 evidence manifest remains the bounded durable record of exact raw inputs,
  hashes, authoring commands, and browser/native PDF identity. Disposable Playwright output and native
  binaries remain ignored and are removed after this gate.
- `sprint-status.yaml` keeps all seven stories `done`, moves `epic-6` to `done`, and leaves
  `epic-6-retrospective` optional.

## 2. Real Chromium boundary

Exact command, from `folio-designer/`:

```text
rtk env PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH='/Users/panitw/Library/Caches/ms-playwright/chromium-1228/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing' npm run test:e2e
```

Result: **20/20 scenarios passed in 2.0 minutes**, serially (`workers: 1`). The suite covered the
application shell, exact/stale Preview, page setup, browser-authored/native round trip, scalar binding,
component placement/manipulation/properties, real Go worker, complete offline reload/update recovery,
local-file fallback and picker paths, parameter discovery and located failure, sample-data discovery,
and the keyboard-operable table matrix.

The round-trip scenario authored the Customer Account Statement and a structurally distinct template
in fresh browser sessions, admitted repeated exact PDFs through PDF.js, and compared the exact saved
template/data/parameter bytes through the native CLI and public Go library. The committed manifest's
golden PDF witness remains **73,705 bytes**, SHA-256
`2a07ca115ffa498969dafe36ab82d2253913b565fbf48d16a9ee135559314c98` for the first render,
repeated browser render, and native render.

## 3. Four-target byte-identity matrix

Exact command, from `folio-go/`:

```text
rtk env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -tags=matrix -run TestCrossTargetByteIdentity -count=1 -v .
```

Result: **PASS in 21.19 seconds**. All 15 documents were byte-identical on local `darwin/arm64`,
Docker `linux/amd64`, Docker `linux/arm64`, and Node `js/wasm`:

| Document | SHA-256 on all four targets | Bytes |
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

The independent FMA discrimination control also passed:

```text
rtk env CGO_ENABLED=0 GOWORK=off go test -tags=matrix -run TestFMAProbeDiverges -count=1 -v .

darwin/arm64  3c40a3d70a3d70a4
linux/amd64   3c50000000000000
linux/arm64   3c40a3d70a3d70a4
js/wasm       3c50000000000000
```

## 4. Go engine, static and sanctioned-red gates

The real deferred heavy tests ran inside the green suite:

```text
rtk env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...
```

Result: **PASS across all 18 packages**. Ordinary and matrix-tagged `go build ./...` and `go vet ./...`
passed with `CGO_ENABLED=0 GOWORK=off`; `gofmt -l .` returned no files.

The sanctioned red was then run directly rather than inferred from the skip:

```text
rtk env CGO_ENABLED=0 GOWORK=off go test -count=1 -run '^TestCorpusMeetsP6ExerciseFloors$' -v ./internal/text
```

P6a–P6f passed. Only `P6g_(opaque_names)` failed, exactly as authorised: **got 7, need >=20**.
The unchanged statistics were `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. The test was
not fixed, weakened, skipped silently, or hidden.

Additional static modules passed independently:

| Scope | Result |
|---|---|
| `hashmatrix` | `go test -count=1 ./...`: **3 tests / 2 packages**; vet and build passed |
| `lint` | `go test -count=1 ./...`: **117 tests / 4 packages**; vet and build passed |
| repository | `git diff --check`: **PASS** |

## 5. Designer and offline-release gates

| Command | Result |
|---|---|
| `npm test -- --run` | **27 files / 162 tests passed** |
| `npm run typecheck` | **PASS** |
| `npm run lint` | **PASS**, four pre-existing Fast Refresh warnings only |
| `npm run test:e2e:compile` | **PASS** |
| `npm run build` | **PASS**, including wasm, TypeScript, Vite and offline generation |
| `npm run verify:offline` | **PASS** |
| `npm run verify:offline:red` | **PASS** |
| `npm run verify:offline:wasm` | **PASS** |

## 6. Disposition

Every non-sanctioned Epic 6 boundary obligation passed. Epic 6 transitions to `done`; its retrospective
remains optional. The P6g red remains conspicuous and unchanged. The owner decision in D-000.78 / DW-4
authorises a release only after Epic 6 but also requires the intermediate lead checkpoint, while DW-20
remains an explicit pre-tag obligation; this boundary commit therefore creates no release tag. Existing
unrelated `_bmad` configuration and manifest changes, `.agents/`, and planning research remain excluded
from the boundary commit.
