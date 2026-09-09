# Epic 13 Boundary Gate — the preview screen is the evidence screen

**Run 2026-09-09 by the orchestrator, at `d08ca04`.** Under owner ruling **D-000.33** this gate carries what
the per-story cadence no longer does: the full heavy suite, and **a wait for CI to come back green on the
pushed epic head**. An epic does not close on a red or an unfinished run.

## Verdict: PASS. Epic 13 closes.

## What ran, and what it said

**Local, all six stories at once** — the first full-suite run since the cadence changed:

| leg | result |
|---|---|
| `folio-go` `go test -count=1 ./...` | pass, **only the two standing reds by name** — `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)`, got 7 need ≥20. No third failure. |
| `folio-go` `-tags=matrix` | same two, no others |
| `lint` module | four packages ok |
| `hashmatrix` | vet clean, tests ok |
| `gofmt -l` over all three modules | empty |
| full build chain (scans, wasm, tsc, vite, offline, verify) | pass |
| `verify:offline:red`, `verify:offline:wasm` | pass |
| **`s1.assetCount`** | **62** of a maximum 64 — the release's own approach warning names the margin of 2 |
| browser suite | **43 passed, 4.0 min** |

**CI on the pushed head `d08ca04`** — run `34324585508`, **all seven jobs green**: `folio-go`,
`folio-go-matrix`, `folio-go-known-red`, `folio-designer`, `folio-designer-e2e`, `hashmatrix`, `lint`. The
`Cross-target byte identity` workflow (`34324585515`) also green.

## The finding this gate exists to have caught

The first browser run **failed**, on `e2e/preview-page-rail.spec.ts` — a spec written in Story 13.6 that,
per D-000.33, had never been executed by anything. That is the cadence working as designed: deferring the
heavy suites means a story can ship with unexecuted browser code, and the gate is where that comes due.

The failure was not what it looked like. `.document-name` stayed `Untitled template`, which reads as "the
file did not load" — but the file *was* delivered, and the engine *accepted* it (`wasm.Engine.Load` and
`.Serialize` both succeed; `folio validate` exits 0). The refusal is browser-side at
`engine-protocol.ts:552`: components in `pageHeader`/`pageFooter` must satisfy `y + height <= band.height`,
and the fixture's header element is `4 + 16 = 20` against a band of `18`. **Two points.**

**The orchestrator's hypothesis was falsified.** I nominated the fixture's five-page y-offsets. Those are
exactly the legal part — the `content` band is deliberately not vertically capped, so markers at
`y=2912000` are fine by design.

**Scope, measured rather than guessed.** Scanning all 26 `fixtures/*/input.folio` found **six** violating the
rule with byte-identical geometry — `multi-page`, `page-count-1`, `page-count-5`, `page-count-20`,
`page-count-50`, `three-band-page`, every one `y4 + h16 = 20 > band 18`. One ancestor copied five times.
Registered as **DW-323**, with a guard as part of what discharges it: nothing today prevents a seventh.

Fixed by giving the spec its own inline template with that one value corrected. **No assertion was touched** —
verified from the diff, zero changed `expect(` lines — and the shared fixture family was left alone, because
`multi-page` and `three-band-page` may carry pinned goldens that a two-point move would invalidate.

## What Epic 13 shipped

- **13.1** the export latch · **13.2** the viewer navigates like a PDF viewer · **13.4** a preview with no
  data that claims nothing about production · **13.3** the evidence rail · **13.5** the chrome tells the
  truth about the preview · **13.6** the page-thumbnail rail, on vendored pdf.js.

## What Epic 13 did NOT ship, and it is recorded rather than hidden

**The diagnostic map.** The epic's goal names "a page-thumbnail rail that doubles as a diagnostic map". The
rail shipped; the map did not, by owner ruling **D-13.6.3**, because there is no data to build it from —
`EngineDiagnostic` carries no page, `isDiagnostic`'s `hasExactKeys` rejects one, and five of seven
`Diagnostic{` sites in `folio-go/render.go` run *before pagination*. **DW-311** carries it and still needs an
epic home. A partial map marking two of seven sites was rejected by builder, orchestrator and owner alike:
an unmarked page would read as clean.

## Open items this gate hands forward

- **DW-324** (owner) — the engine renders templates the designer cannot open, reported only as "Could not
  open local file". All ten `containComponent` call sites are command paths; none is a load path.
- **DW-311** (owner) — the diagnostic map needs an epic.
- **DW-313** (owner) — two asset slots remain before Epic 14's ten stories of UI.
- **DW-305** — the contrast contract's five-pair list is not exhaustive.
- **DW-323** — the six-fixture family, plus the guard that would prevent a seventh.
