---
title: 'Story 17.1: The canvas follows the content field'
type: 'feature'
created: '2026-09-04'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: '85f213f'
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Type into the Text box today and the canvas shows nothing until you click away. The text you are writing
and the page you are writing it onto disagree for as long as you keep typing.

After this story the canvas catches up shortly after you stop — about a fifth of a second — without you
leaving the field. Not on every keystroke: the browser is not allowed to lay text out itself, so every
update is a real round trip to the engine, and one per keystroke would be one command per keystroke.

<intent-contract>

## Intent

**Problem:** `App.tsx:1837`'s `onChange` calls `setDraft` and nothing else. The only commit is `onBlur`
(`:1835`), and for the prose field Enter inserts a line feed rather than committing (`:1797` guards on
`!prose`). So the canvas cannot reflect typed text until focus leaves the control.

**Approach:** After a pause in typing, commit the draft the same way blur does — through the existing
`submit` path, as a real command.

## Boundaries & Constraints

**Always:**
- **THE BROWSER STILL NEVER MEASURES OR RE-FLOWS TEXT (AD-17).** New line breaks exist only after the
  engine returns a new `textPaint`. `.canvas-text-*` keep `white-space: pre` (`App.css:106`), and
  `canvas-authority-contract.test.ts` must stay green with no new exception.
- **OWNER DECISION, TAKEN AT THE GATE: a debounce of ~200 ms sending a REAL command.** Three costlier
  options were declined with their costs stated. **The acceptance wording must therefore not promise
  per-keystroke painting** — it promises the canvas catches up shortly after typing stops.
- **TYPING MUST NEVER LOSE A CHARACTER, AND THIS IS THE STORY'S REAL RISK.** `submit` (`:1771-1777`)
  returns early when `pendingRef.current` is set, and on acceptance it calls
  `setDraft(canonicalValue(accepted, ids, field) ?? draft)` (`:1778`). A reconciliation landing while the
  author has typed further would overwrite what they typed with what the engine echoed. **The draft the
  author is holding wins; a late echo may not clobber it.**
- **A DEBOUNCED COMMIT THAT COLLIDES WITH AN IN-FLIGHT ONE MAY NOT BE SILENTLY DROPPED.** Today's early
  return is correct for a click; for a timer it means the last thing typed never reaches the engine.
- **Blur still commits, Escape still reverts, Enter still inserts a line feed.** No existing behaviour
  of this control changes.
- **`contentCommand` (`:1634`) routes to `value` or `expression` on whether the text holds `{{` or `}}`.**
  A half-typed `{{` now reaches the engine mid-word where it never did before. **Decide and test what a
  document does with a partial expression rather than discovering it in front of an author.**
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Any change to `engine-protocol.ts`, or any new engine operation.
- Applying the debounce to any field other than the prose content field.
- Changing the undo model, or coalescing commands.

**Never:** measuring text in the browser · a new `white-space`/wrap rule on a canvas class · a keystroke
that is lost or reordered · a command sent while the field is `readOnly` · a timer that survives unmount.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Type and pause | `Invoice` then stop | One command; canvas paints `Invoice` without leaving the field | — |
| Type continuously | 7 characters faster than the debounce | **One** command, not seven, carrying the final text | — |
| Keep typing during an in-flight commit | Text changes while a command is awaited | The later text still reaches the engine; **nothing is dropped** | Never silently discarded |
| Echo arrives late | Engine accepts older text after more typing | **The author's current draft survives**; the echo does not overwrite it | — |
| Blur while a debounce is pending | Click away mid-timer | Exactly one command for that text, not two | — |
| Escape mid-typing | Draft differs from committed | Reverts to committed and blurs; **no debounced command follows** | Timer cancelled |
| Unmount mid-timer | Selection changes | No command fires after unmount; no React state update warning | Timer cleared |
| Partial expression | Author has typed `{{cust` | Behaviour is defined and asserted, not incidental | — |
| Engine refuses | Command rejected | The existing error path renders; typing is not blocked | Existing `role="alert"` |

</intent-contract>

## Code Map

Anchors at `85f213f`. **Re-verify — `App.tsx` is a long-line file and these are one very long line each.**

- `App.tsx:1837` — the prose `onChange`, the single place a keystroke is observed. **The seam.**
- `App.tsx:1771-1780` — `submit` (single-flight via `pendingRef`, reconciles via `canonicalValue`) and
  `commit`. **Both hazards named above live here.**
- `App.tsx:1835` — `shared`, carrying `onBlur: () => void commit()` and `readOnly: live !== undefined`.
- `App.tsx:1796-1799` — `keyDown`; Enter is guarded on `!prose`, Escape reverts and blurs.
- `App.tsx:1628` — `contentField`, `prose: true`. `:1634` — `contentCommand`'s `{{`/`}}` routing.
- `App.tsx:749-766` — `applyProperties`, the `onCommit` this ends in; the returned snapshot replaces state.
- `component-property-command.ts:19-26` — the command bytes. **Unchanged by this story.**
- `App.tsx:2487-2490` `TextPaint`, `engine-protocol.ts:174` `textPaint` — the engine's projection, which
  is the ONLY source of painted text. `App.css:106` — `white-space: pre`.

## Tasks & Acceptance

**Execution:**
- [ ] Debounce the prose field's draft and commit through the existing `submit`, cancelling on blur,
      Escape and unmount.
- [ ] **Fix the two collision hazards** — a debounced commit must not be dropped by the in-flight guard,
      and a late echo must not overwrite a newer draft. **These are the story, not a detail of it.**
- [ ] Decide and assert the partial-expression behaviour.
- [ ] Tests — one per matrix row, driving the real path in a mounted designer with a fake timer.
- [ ] A browser run: type into the field without leaving it and photograph the canvas following.

**Acceptance Criteria:**
- Given typing then a pause, when the debounce elapses, then the canvas shows the text **without focus
  leaving the field**.
- Given seven characters typed faster than the debounce, when it elapses, then **exactly one** command
  was sent.
- Given continued typing across an in-flight commit, when everything settles, then the engine holds the
  final text and **no character was lost**.
- Given Escape mid-typing, when the timer would have fired, then **no command is sent**.

## Design Notes

**Why a debounce and not a preview.** The canvas paints only what the engine projects, so live text is a
round trip whatever we do. A non-committing preview op would give per-keystroke painting with an
untouched undo stack, and it was offered and declined for cost. **The consequence the owner accepted is
that the canvas lags typing by the debounce**, and this story must not quietly claim otherwise.

**The dangerous part is not the timer.** It is that this control's `submit` was written for events an
author generates one at a time — a click, a blur. A timer makes overlap ordinary. Both of its
single-author assumptions — the early return and the draft reconciliation — become wrong at that point,
and neither would fail loudly.

## Verification

**Nothing in Epic 16 was CI-verified and that repair has just landed (56d9413); the designer job now runs
all eight steps. It has still never run on a remote — nothing is pushed.** Say what you did not run.

**Commands** — one per line, **exit codes read from `$?` immediately, never through a pipe**; zsh, so
`${PIPESTATUS[0]}` is wrong. **A piped exit code reported a failing run as 0 in this repo today.**
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts`'s corpus
  scan (DW-152); **match it by NAME, not count.** The passing count must rise by the tests added.
- `npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`.
- **No Go is touched. Say so rather than running it.**

**A BROWSER RUN.** `chromium-1217` via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (`chromium-1208` is a 428K
stub, DW-180).

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**9 rows**);
state the population beside every zero and pair every absence claim with a positive control; **a comment
is not a measurement**; **use `/usr/bin/grep`** — recursive `grep` returns false zeros here.
