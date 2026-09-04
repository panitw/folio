---
title: 'Story 17.4: Arrow keys step a number field'
type: 'feature'
created: '2026-09-04'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: '85f213f'
context: []
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Every number box in the inspector ignores the up and down arrows. Nudging a width by a point means
selecting the text and retyping it.

After this story the arrows step the value, the way they do in every other design tool.

<intent-contract>

## Intent

**Problem:** `App.tsx:1796-1799`'s `keyDown` handles exactly two keys — `Enter` (single-line only) and
`Escape`. Measured: within `PropertyDraft` (`:1755-1839`) `event.key` occurs only on those two lines,
against a positive control of matches at `:1455` and `:2377` elsewhere in the file.

**Approach:** Add up/down stepping to the numeric fields, in the field's own unit.

## Boundaries & Constraints

**Always:**
- **THE STEP IS DERIVED FROM THE FIELD, NOT A CONSTANT.** The smallest representable increment is
  0.001 for every one of these — points are stored in **thousandths** (`App.tsx:2335` `points()` divides
  by 1000; `internal/template/decimal.go:30-58` refuses a fourth decimal place). **0.001 is the floor of
  the representation, not a useful step.** A point field steps by 1; leading is a dimensionless ratio and
  steps by 0.1. **State the chosen steps in the spec's own terms and test them.**
- **THE NUMERIC SET IS THE ONE THE CONTROL ALREADY KNOWS.** `App.tsx:1838` computes
  `inputMode={unit === 'pt' || unit === undefined && field === 'lineSpacing' ? 'decimal' : undefined}`.
  That is `x`, `y`, `width`, `height`, `fontSize`, `borderWidth`, `lineSpacing`. **Reuse that predicate;
  do not mint a second list that can drift from it.** Measured: there is **no `mm` unit anywhere** —
  `\bmm\b` returns 0 lines against 162 substring hits inside `command`/`commit`/`comment`, and a
  positive control of 4 for `unit: 'pt'`.
- **AN ARROW MAY NOT PRODUCE A VALUE THE ENGINE WILL REFUSE.** `width`, `height`, `fontSize` and
  `borderWidth` must stay `> 0` (`component_commands.go:1005-1007`); `lineSpacing` is bounded to
  0.001–1000 (`internal/template/linespacing.go:53-54`, `:61-71`). **Stepping down at the floor must
  stop, not underflow into a refusal the author did not ask for.**
- **THE TYPED LITERAL IS PASSED THROUGH UNQUOTED AND GO ALONE PARSES IT**
  (`component-property-command.ts:29-33`). A step must therefore produce a literal with **at most three
  decimal places** — floating-point arithmetic here can emit `13.999999999999998`, which Go refuses.
  **Do the arithmetic in integer thousandths.**
- **THE GEOMETRY FIELDS ARE `readOnly` DURING A DRAG** (`App.tsx:1835`, `readOnly: live !== undefined`).
  **An arrow must do nothing then**, exactly as typing does nothing.
- **An empty box has no value to step.** Define what an arrow does on an unset field and assert it.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Modifier behaviour (Shift for a coarse step, Alt for a fine one) — **not in scope unless asked for.**
- Applying stepping to any non-numeric field, or to the prose field.
- Changing `inputMode`, or introducing `<input type="number">`.

**Never:** a value with more than three decimals · a step past a bound into a refusal · an arrow that
fires while `readOnly` · a second list of numeric fields · float arithmetic on the value.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Up on a point field | `width` = `12` | Becomes `13` | — |
| Down on a point field | `width` = `12` | Becomes `11` | — |
| Up on leading | `lineSpacing` = `1` | Becomes `1.1`, **not `1.1000000000000001`** | — |
| Down at the floor | `width` = `1`, step would reach `0` | **Stops at the smallest legal value**; no refusal | Never sends an illegal value |
| Down at leading's floor | `lineSpacing` = `0.001` | Stops; no refusal | — |
| Arrow on an unset field | Box empty, placeholder showing | **Defined and asserted** behaviour | — |
| Arrow during a drag | `live` set, field `readOnly` | **Nothing happens**; no command | — |
| Arrow on a non-numeric field | `Text`, `Border colour` | Unhandled — the browser's own caret behaviour | — |
| Arrow on a mixed selection | Two components, different widths | **Defined and asserted** | — |
| Enter and Escape | — | **Unchanged** — still commit and revert | — |

</intent-contract>

## Code Map

Anchors at `85f213f`. **Re-verify — these are long lines.**

- `App.tsx:1796-1799` — `keyDown`. **The seam; two keys today.**
- `App.tsx:1838` — the `inputMode` predicate. **The numeric set, already computed once.**
- `App.tsx:1835` — `shared`, carrying `readOnly: live !== undefined` and `value: live ?? draft`.
- `App.tsx:1771-1780` — `submit` / `commit`; `commit` fires only when `draft !== committed`.
- `App.tsx:2335` `points()` — ÷1000, 3-dp, trailing zeros stripped. **The formatter to round-trip through.**
- `App.tsx:1616-1617` `x`/`y`/`width`/`height`; `:1646` `fontSize`; `:1651` `lineSpacing`; `:1661` `borderWidth`.
- `component-property-command.ts:6` `pointFields`, `:16` `ratioFields`, `:29-33` `rawNumberLiteral`
  ("Preserve the typed literal, unquoted; Go alone decides whether it is a valid number").
- `folio-go/internal/template/decimal.go:30-58` — the three-decimal refusal.
- `folio-go/component_commands.go:1005-1007` — the `> 0` rule.

## Tasks & Acceptance

**Execution:**
- [ ] `App.tsx` — step the draft on ArrowUp/ArrowDown for the numeric set, **in integer thousandths**,
      formatted back through `points()`.
- [ ] Clamp at each field's real bound rather than sending a value the engine refuses.
- [ ] Do nothing while `readOnly`; leave Enter and Escape untouched.
- [ ] Decide and assert the unset-field and mixed-selection behaviours.
- [ ] Tests — one per matrix row, **including a value that would expose float arithmetic** (leading from
      `1` upward is the one that does).
- [ ] A browser run: hold an arrow on a width and photograph the canvas moving.

**Acceptance Criteria:**
- Given `width` = `12`, when ArrowUp is pressed, then the box reads `13` and the committed value follows.
- Given `lineSpacing` = `1`, when ArrowUp is pressed, then the box reads exactly `1.1`.
- Given a field at its smallest legal value, when ArrowDown is pressed, then **no illegal value is sent**.
- Given a drag in progress, when an arrow is pressed, then **nothing happens and no command is sent**.

## Design Notes

**The trap here is arithmetic, not keys.** Every value in this control is a decimal string that Go parses
exactly and refuses beyond three places. `1 + 0.1` in IEEE doubles is not `1.1`, and the resulting literal
is refused by `decimal.go` — a failure that would look like a mysterious rejection of an arrow press.
**Stepping in integer thousandths and formatting through `points()` is the whole design.**

## Verification

**Commands** — one per line, **exit codes read from `$?` immediately, never through a pipe**; zsh, so
`${PIPESTATUS[0]}` is wrong.
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts`'s corpus
  scan (DW-152); **match it by NAME, not count.**
- `npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`. **No Go is touched; say so rather than running it.**

**A BROWSER RUN.** `chromium-1217` via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (DW-180).

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**10 rows**);
state the population beside every zero, pair every absence claim with a positive control; **a comment is
not a measurement**; **use `/usr/bin/grep`** — recursive `grep` returns false zeros here.
