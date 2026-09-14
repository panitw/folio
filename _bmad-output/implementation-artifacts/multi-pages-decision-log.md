# Multi-Pages delivery run — decision log

Program: SPEC-multi-pages (`_bmad-output/specs/spec-multi-pages/`), stories 1–5 from `stories.yaml`.
Started 2026-09-14 from baseline `b6b8b6f`.

## Lead Grounding

_Pending — filed by the orchestrator from the engineering lead's grounding report._

## Standing decisions

### D-S.1 — Build order is stories.yaml list order, 1 through 5
**Owner decision** (at breakdown). The stories are layered: 1 is the engine and file format, 2 draws pages on the canvas, 3 moves elements between pages, 4 lets the header and footer be edited from any page, and 5 adds the per-page Section Break and Page Break off. Each story builds on the committed code of the one before it, so they run one at a time. Stories 5 and 6 of the first breakdown were merged, because both reuse the section-break pagination code.

### D-S.2 — No spec review or pause after a story; the orchestrator and lead approve CHECKPOINT 1
**Owner decision.** `spec_checkpoint` and `done_checkpoint` are false on every story. The run continues on its own and pauses only when the lead escalates a design decision to the owner. The owner sees each plan through the plain-terms opener in the per-story report.

### D-S.3 — Owner questions are asked at the terminal
**Owner decision.** A Telegram config exists, but the owner chose the terminal as the answer channel.

### D-S.4 — Heavy-test cadence: end of run
**Owner decision.** The owner accepted that a failure found at the end could come from any of the five stories.
- **Every story:** the `## Verification` section runs, in folio-go, `go build ./...`, `go vet ./...`, the gofmt check, `go test -count=1 ./...` (skipping the CI `KNOWN_RED_TEST`), and a compile-only `go build -tags=matrix ./...`. In folio-designer it runs `npm run typecheck`, `npm run lint`, `npm test` and `npm run test:e2e:compile`.
- **Deferred to the end of the run:** `npm run test:e2e` (Playwright) and the matrix hash test `go test -tags=matrix -run TestTargetRenderHash .`. Locally that covers only the host platform; the four-platform comparison runs in CI's matrix workflow.
- **Tests are still written in their story.** Only running them moves to the end.

### D-S.5 — Commit to main, never branch, never push
**Owner decision** (standing memory). The builder commits locally on `main` and turns down step-05's offer to push or open a PR.

### D-S.6 — After any folio-go change, run `npm run build:wasm` before designer tests
**Owner decision** (standing memory, and `invoke_dev_with` on every story). `npm run dev` never rebuilds the engine. `npm test` and `npm run typecheck` already run it.

### D-S.7 — Spec decisions are settled and must not be re-opened
**Orchestrator decision.** The builder must not raise any of the following as Open Questions:
- everything in `SPEC.md` Constraints and Non-goals
- the memlog's resolved questions: a 4.x MINOR version; a Delete Page confirmation that names the page; Page Setup plus a page section holding Page Break; the current page's header and footer copy carrying the accessible name
- Page Break is saved as true or false on every page after page 1, a missing value loads as on, and a value on page 1 is ignored and dropped on save
- Page Break off follows directly after the previous page's content, like an unanchored section break
