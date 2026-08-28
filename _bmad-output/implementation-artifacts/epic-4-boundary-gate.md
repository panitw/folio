# Epic 4 Boundary Gate

**Date:** 2026-08-28  
**Baseline:** Story 4.8 commit `0838db49021a3d409ee19b4f0f3431d549343d80`  
**Cadence:** D-000.4, per-epic boundary  
**Verdict:** **PASS**

This gate was run from the committed Story 4.8 code. Docker Desktop was available for the Linux
legs (`docker` server `29.6.2`), and Node was `v24.16.0`. No Epic 5 work was started.

## 1. Four-target matrix

Exact command, from `folio-go/`:

```text
rtk env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -tags=matrix -run 'TestCrossTargetByteIdentity' -count=1 -v .
```

The harness executed the real local `darwin/arm64` leg, Linux `amd64` and `arm64` legs in Docker,
and `js/wasm` under Node. All 15 registered documents matched on all four targets and matched
their committed expected digest:

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

The retained FMA discrimination probe also passed. Its repeated target values were:

```text
darwin/arm64  3c40a3d70a3d70a4
linux/amd64   3c50000000000000
linux/arm64   3c40a3d70a3d70a4
js/wasm       3c50000000000000
```

The arm64 and wasm values differ as expected; every target is stable across repeated execution.

## 2. DW-21 heavy/integration gate

Exact required command, from `folio-go/`:

```text
rtk env CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -count=1 -v -run 'TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable|TestTwoTablesWithPageCountFooterRenderConsistently|TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability' ./...
```

Results:

- `TestTableHeaderRepeatAcrossHundredsOfPagesIsByteStable`: **PASS** (0.04s)
- `TestTwoTablesWithPageCountFooterRenderConsistently`: **PASS** (0.01s)
- `TestFooterOrphanTieHoldsAcrossHundredsOfPagesWithByteStability`: **PASS** (0.08s)
- No `--- SKIP` test names were emitted.
- Nonmatching packages reported `no tests to run`, as expected from the exact `-run` expression;
  no heavy test was skipped.

DW-21 is discharged by this run, with all three names reported individually rather than as a count.

## 3. Ordinary gates

| Scope | Command/result |
|---|---|
| `folio-go` ordinary suite | `env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` — **PASS** |
| `lint` | `env GOPROXY=off CGO_ENABLED=0 GOWORK=off go test -count=1 ./...` — **PASS** |
| `hashmatrix` | `env CGO_ENABLED=0 GOWORK=off go test -count=1 ./...` — **PASS** |
| format/diff | `gofmt -l .` empty; `git diff --check` clean |
| ordinary build/vet | `go build ./...` and `go vet ./...` in `folio-go` — **PASS** |
| tagged build/vet | `go build -tags=matrix ./...` and `go vet -tags=matrix ./...` — **PASS** |

The sanctioned full-suite red was rechecked separately and is unchanged:
`TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` reports `got 7, need >=20`, with stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. This is the standing intentional red under
D-000.17, D-000.57, D-000.74, and D-2.1.14; it is not an Epic 4 failure.

## 4. Deferred integration/e2e census

The complete Epic 4 deferred execution set consisted of the three DW-21 tests above. Story logs
also contain historical references to their in-story skips and to the matrix deferral; those are
truthful story-time records, not additional unrun suites. No other Epic 4 integration/e2e work is
deferred beyond the full matrix and the three heavy tests, and all of those obligations ran here.

## 5. Contract, story, and tracker audit

- All eight Epic 4 story files (`4-1` through `4-8`) contain an `## In plain terms` opener. The
  openers are present at lines 275, 46, 50, 77, 105, 12, 15, and 17 respectively, and describe
  each story's delivered scope and intentionally deferred work in plain language. Statements that
  the matrix/heavy work was deferred are historical delivery-log statements and are superseded by
  this gate evidence.
- The architecture spine remains consistent: staged immutable pipeline, stage-rank direction, and
  the no-layout-to-PDF dependency are unchanged. No Epic 4 source-tree or invariant drift was found.
- Current-contract drift was corrected before this verdict: `folio-format.md` now records that
  aggregate functions and footer rendering are implemented; its expression section records all
  eight functions as implemented. `folio-go/README.md`, `SPEC.md`, and the PRD NFR1 assumption now
  identify Epic 4 as the latest 15-document four-target proof.
- `fixtures/statement-signoff.json` is valid and binds exactly the four live PDFs. The matrix-tagged
  `TestStatementSemanticSignOffIsRecorded` passed. Direct SHA-256 checks matched all four records:
  `statement-1` `ef58bbf6…e86611d`, `statement-5` `7f67b317…ff950f`, `statement-20`
  `be6f5e27…e3cd4ed`, and `statement-50` `9c5be7ba…c754c`.
- DW-4's owner decision is no longer unowned: D-000.78 records the project owner's choice to cut
  `folio-go/v0.1.0` after Epic 6. The deferred-work entry remains open only for the intermediate
  engineering-lead checkpoint required before that future tag. DW-13 remains correctly owned by
  the project owner; D-4.7.5 declined compression and defines a fixed-input compressor-determinism
  condition for re-entry. DW-11 remains the sanctioned P6g shortfall at 2 genuinely opaque sourced
  names; its Epic 4 signal fired and was recorded without inventing corpus data. Future owners for
  DW-2, DW-12, DW-16, DW-17, and DW-20 remain named and are not due in Epic 4.
- Pre-existing unrelated worktree churn was preserved and excluded: `_bmad/_config/files-manifest.csv`,
  `_bmad/_config/manifest.yaml`, `_bmad/bmm/config.yaml`, `_bmad/core/config.yaml`, and `.agents/`.

## 6. Disposition

Every Epic 4 boundary obligation passed. `epic-4` may transition to `done`; DW-21 is discharged;
the intentional P6g red remains explicitly disclosed; no Epic 5 work begins in this gate. This artifact,
the tracker/decision/deferred-record updates, and the current-contract corrections are the complete
scoped boundary change. The final local commit hash is reported in the boundary handoff because a
commit cannot contain its own hash before it exists.
