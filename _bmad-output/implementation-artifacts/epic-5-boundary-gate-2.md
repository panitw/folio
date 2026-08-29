# Epic 5 Boundary Gate — second run

**Date:** 2026-08-29
**Baseline:** `6c79d29` on `main` (Story 5.13 finisher `97d0d3f`, plus the working-tree commit)
**Cadence:** D-000.4, per-epic boundary — owed by Story 5.13's reopening of Epic 5 (D-5.13.6)
**Verdict:** **PASS**, with one pre-existing defect named as debt (DW-23)

This gate discharges two obligations at once: the four-target matrix that Story 5.13 deliberately
declined in-story (D-5.13.5), and the re-measurement of the post-Epic-6-gate delta that D-5.13.6
assigned here rather than to a re-run of the Epic 6 gate. No release tag was cut.

The first Epic 5 gate (`epic-5-boundary-gate.md`, PASS at `4bd6331`) stands unchanged as
baseline-anchored evidence for Stories 5.1–5.12. This document supersedes it as the description of
Epic 5 at HEAD.

## 1. Story and record audit

- Epic 5 carries **13** stories, all `done`. Story 5.13 completes FR4/FR5 for the Image component,
  which Stories 5.7 and 5.8 left without any way to choose the picture.
- Story 5.13's record is complete: plain-terms opener rewritten by the finisher to state what
  shipped, task list closed, `## QA Results` carrying 20 review findings, and a Finding Resolutions
  section recording 19 FIX / 0 DISMISS / 1 DEFER.
- Decisions **D-5.13.0 – D-5.13.9** are recorded in `folio-mvp-decision-log.md`, including two
  entries that correct earlier rulings rather than rewriting them (the D-5.13.2 amendment on key
  width, and D-5.13.7 overruling the substitution of Go tests for an AD-21 fixture).
- **DW-22** (per-element rather than per-distinct-asset paint fetching) is filed with this gate named
  as its owner. Reviewed here: it is a bounded cost on a path that already round-trips through the
  engine, not a correctness gap, and it stays open.
- **DW-11** remains open and visible as the sanctioned P6g corpus-floor red. **DW-4**'s release
  checkpoint and **DW-20**'s pre-tag obligation remain open, so this gate creates no
  `folio-go/v0.1.0` tag.
- Per **D-5.13.9**, the committed `evidence/story-6.7-roundtrip-manifest.json` post-dates the Epic 6
  gate that cites it. This gate treats it as input **re-derived** by §2's run, never as prior evidence
  to compare against.

## 2. Real Chromium boundary

Exact command, from `folio-designer/`:

```text
PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="$HOME/Library/Caches/ms-playwright/chromium_headless_shell-1217/chrome-headless-shell-mac-arm64/chrome-headless-shell" npm run test:e2e
```

Result: **23 scenarios passed, 0 failed, in 2.0 minutes**, serially (`workers: 1`), across **12**
spec files. The suite was run three times over the course of Story 5.13 and this gate, green each
time at 23/23.

Two scenarios are new since the first Epic 5 gate and are the reason D-5.13.5 granted a per-story
e2e override:

- **`image-asset.spec.ts`** exercises the File System Access tier end to end — set a local image
  through one committed command, paint it, save, reopen, undo/redo. Its picker mock brand-checks its
  receiver, so reverting `capability.ts`'s `.bind(browser)` fails the spec. That property was added
  by the finisher after review found the original mock was an arrow function that passed with
  `eef7fbb`'s defect reintroduced.
- **The image drag regression test**, added after the repository owner found in the real application
  that a placed image could not be dragged: the new `<img>` is natively draggable, so `dragstart`
  took pointer capture before `moveComponent` was reached. Red-proved by reverting
  `draggable={false}` and the two CSS rules, at which point the engine revision never advances past
  `REVISION 2`.

## 3. Four-target byte-identity matrix

Exact command, from `folio-go/`:

```text
CGO_ENABLED=0 GOWORK=off FOLIO_HEAVY=1 go test -tags=matrix -run TestCrossTargetByteIdentity -count=1 .
```

Result: **PASS in 22.751 seconds.** All **16** registered documents were byte-identical on local
`darwin/arm64`, Docker `linux/amd64`, Docker `linux/arm64`, and Node `js/wasm`. Docker Desktop server
`29.6.2`, Node `v24.16.0`.

The document count is **16** in all three registries that must agree — `matrixDocuments`'s slug
entries, `declaredEpic2GateObligations`, and `.github/workflows/matrix.yml`'s `docs=` list. (A naive
`grep -c "matrix-document: "` returns 18: two of those occurrences are a doc comment and a
string-construction line, not entries. Recorded because consuming that 18 as a population would be
the lossy-count defect of D-000.83 / D-000.89 / D-000.91 for a sixth time.)

The sixteenth document, **`component-asset-import`**, is Story 5.13's AD-21 fixture and its first
matrix run. It is not a render golden: its `input.folio` is the captured canonical output of one real
`setComponentAsset` command, and `TestComponentAssetImportCommandReproducesTheFixtureInput` re-runs
that command on every ordinary `go test ./...` to assert byte-for-byte reproduction — pinning the
authoring command's canonical-bytes behaviour (digest-as-key, 76-column wrapping, sorted keys, the
repoint), not merely a render of a document that already names an asset.

**This discharges D-5.13.5's declined override, and confirms its reasoning by measurement.** That
ruling declined the per-story matrix on the ground that Story 5.13 introduced no new *source* of
cross-target divergence — the digest is `crypto/sha256`, already executed on all four arms by
`decodeAssets` on every load; the wrapping is the existing `writeAssets`; the embedding is Story
1.8's untouched passthrough, so no compressor is invoked and R4 stays shut; the fit is integer
cross-multiply plus `ScaleRound`. The green result is consistent with that analysis. The ruling's
conditional — that the decline flips if any image decoder or re-encoder entered the module — was
verified as not triggered: no stdlib or third-party image package exists in `folio-go`.

## 4. Go engine, static and sanctioned-red gates

| Gate | folio-go | lint | hashmatrix |
|---|---|---|---|
| `go build ./...` | clean | clean | clean |
| `go vet ./...` | clean | clean | clean |
| `gofmt -l .` | clean | **DIRTY** | clean |

- `go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` — **GREEN**. This is the gate
  that answers "is anything actually broken?", and nothing is.
- `go test -count=1 -run "^TestCorpusMeetsP6ExerciseFloors$" ./internal/text/` — **FAILS**, as
  designed. P6g reports `got 7, need >=20`, `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115
  P6g:7}`, identical to the figures measured at baseline `3a52ae4` before Story 5.13 began. Under
  D-000.17 an unmet floor is reported unmet, never filled.
- `go build -tags matrix ./...` and `go vet -tags matrix ./...` — clean, so the matrix leg cannot
  open with a stale compile error.

### Finding: `lint`'s gofmt break, and why two gates missed it

`lint/internal/rules/licencegraph_test.go` is not gofmt-clean, and has not been since **Story 5.10**
(`5dddbea`) — the committed file is byte-identical to that commit's version. It is a single collapsed
one-line `for` loop; `gofmt -w` fixes it, and it is unrelated to Story 5.13, which touched no file
under `lint/`.

It survived the Epic 5 and Epic 6 boundary gates for two compounding reasons, and the second is the
more important:

1. **The local gate procedure only ran `gofmt` in `folio-go`.** CI checks all three Go modules
   (`ci.yml` runs the step under `working-directory:` `folio-go`, `hashmatrix` and `lint`
   separately). A gate that measures fewer modules than CI cannot certify what CI will say.
2. **CI's signal cannot distinguish this from the sanctioned red.** The `folio-go-known-red` job is
   red *by design* — its own comment states "this job going GREEN is the surprising event, not this
   job going red" — and it is deliberately quarantined rather than silenced so DW-11's unmet floor is
   reported in three places at once. The consequence is that the **workflow** is permanently red, so
   a genuine failure in the sibling `lint` job is camouflaged. `gh run view` confirms both `lint` and
   `folio-go-known-red` failing on the last two pushes to `main`, indistinguishable from the badge.

This is the D-000.38 shape — two conditions sharing one signal, so neither can be read alone — applied
to CI job status rather than to a parser. **Filed as DW-23**, owned by whoever next touches `lint`.
Not fixed here: this gate measures and reports; changing a file mid-gate would muddy what was
measured, and D-000.17's discipline is to report an unmet condition rather than quietly fill it.

## 5. Designer and offline-release gates

From `folio-designer/`:

- `npm run typecheck` — **PASS**.
- `npx vitest run` — **29 files, 190 tests, all passing** (166 at the Story 5.13 baseline; +24).
- `npm run lint` — no errors. Three `react(only-export-components)` warnings, all pre-existing
  (`pdf-viewer.tsx:13`, `App.tsx:947`, `App.tsx:954`).
- `npm run build` — clean, including `build:wasm`, `tsc -b`, `vite build`, `generate-offline-release`
  and `verify:offline`. The offline release verification is the check that covers the release-payload
  dev bypass named in §6.
- `npm run test:e2e:compile` — clean.

## 6. The post-Epic-6-gate delta (D-5.13.6's obligation)

D-5.13.6 recorded that two commits landed after the Epic 6 gate with no story key, and assigned their
re-measurement here. Measured:

- **`eef7fbb`'s `createComponentInBand` change** (palette-placed text adopts the first declared
  non-empty font chain) — the Epic 6 amendment warned that any golden recording a palette-placed text
  element would have moved. **Neither post-gate commit touched `fixtures/` or any golden at all**, and
  §3's matrix is green across all 16 documents, so nothing moved. The warning is discharged. The
  change is also not untested: `eef7fbb` shipped `component_properties_test.go` (+38) alongside it.
- **`eef7fbb`'s render-failure / `ENGINE_REJECTED` message policy change** in
  `wasm/cmd/engine/main.go` — Story 6.6's diagnostic contract surface. Shipped with
  `main_test.go` (+55), and §2's suite exercises the located-failure path green.
- **`eef7fbb`'s release-payload dev bypass** in `offline-lifecycle.ts` / `main.tsx` /
  `verify-offline-release.mjs` — covered by §5's `verify:offline`, which passes.
- **`3a52ae4`'s five edited Playwright specs** — §2 re-ran the full suite at 23/23.

What these commits lacked was a story record, not test coverage. That record gap is closed by the
Epic 6 gate's amendment and by D-5.13.6; nothing further is owed.

## 7. Disposition

- **PASS.** `sprint-status.yaml` moves `epic-5` to `done` and leaves `epic-5-retrospective` optional.
  All 13 Epic 5 stories remain `done`.
- **DW-23** is filed for `lint`'s gofmt break and the two-gate blind spot that hid it. **DW-22**,
  **DW-11**, **DW-4** and **DW-20** remain open. No release tag is cut.
- Recommended, and not done here because it is a process change rather than a measurement: the local
  gate procedure should run `gofmt -l` in **all three** Go modules, and the boundary checklist should
  read CI's **per-job** conclusions rather than the workflow badge, which is permanently red by
  design.
- Disposable Playwright output (`folio-designer/test-results/`) is gitignored and left in place; it
  is regenerated by every run.
