---
title: 'Story 17.1: The canvas follows the content field'
type: 'feature'
created: '2026-09-04'
status: 'in-review'
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

**RE-MEASURED AT `c13864c`, WHICH IS THE BASELINE YOU ARE BUILDING ON.** The frontmatter's `85f213f` is
the plan-gate commit; Stories 17.2, 17.4 and 17.3 landed after it and moved `App.tsx` by up to **+202
lines**. Every anchor below was re-verified by reading the file at `c13864c`. The old numbers are kept
beside them only so you can see the drift. `App.tsx` is a long-line file — several of these are one very
long line each.

| What | Spec said (`85f213f`) | **ACTUAL (`c13864c`)** |
|---|---|---|
| prose `onChange` — **the seam** | `:1837` | **`App.tsx:2039`** (the `<textarea>`) |
| `submit` (single-flight + reconcile) | `:1771-1777`, `:1778` | **`App.tsx:1868-1881`**; the reconcile is **`:1880`** |
| `commit` | `:1780` | **`App.tsx:1888`** |
| `shared` (`onBlur`, `readOnly`, **`disabled`**) | `:1835` | **`App.tsx:2037`** |
| `keyDown` (Enter guarded on `!prose`, Escape) | `:1796-1799` | **`App.tsx:1988-2001`**; Escape is **`:1990`** |
| `contentField`, `prose: true` | `:1628` | **`App.tsx:1649`** |
| `contentCommand` `{{`/`}}` routing | `:1634` | **`App.tsx:1655`**; `containsPlaceholder` **`:1654`** |
| `applyProperties` | `:749-766` | **`App.tsx:749-774`** (did not move) |
| `TextPaint` | `:2487-2490` | **`App.tsx:2689`** |
| `canonicalValue` | — | **`App.tsx:2042`** |
| `white-space: pre` | `App.css:106` | **`App.css:106`** (did not move) |
| `textPaint` in the protocol | `engine-protocol.ts:174` | **`engine-protocol.ts:179`** (in `components`), guarded at `:378-379` |
| the command bytes | `component-property-command.ts:19-26` | unchanged by this story |

### Three things I measured at `c13864c` that the Code Map above did not know

These are **measurements, not opinions** — each was taken by running code at the baseline, and each one
changes what you must write. Where a decision follows from one, the decision is stated here so you do not
have to make it yourself.

**1. `shared` CARRIES `disabled: pending` (`App.tsx:2037`), AND `submit`'s `disable` PARAMETER IS THE
ONLY THING THAT KEEPS A FOCUSED FIELD ALIVE.** `submit(intent, reconcileDraft, disable = true)` raises
`pending` for the duration of the round trip, which **disables the textarea the author is typing into**.
Story 17.4 already hit this and wrote the finding down at `App.tsx:1849-1857`: *"disabling a FOCUSED
input moves focus to the body and does not give it back when the input is re-enabled (measured in
Chromium 1217)"*. Its arrow step therefore passes `disable = false` (`App.tsx:1963`).

- **A debounced commit MUST pass `disable = false` for exactly the same reason.** With the default, every
  debounce would blank the author's focus mid-sentence and the acceptance criterion *"the canvas shows
  the text **without focus leaving the field**"* would be false in a real browser.
- ⚠ **jsdom WILL NOT CATCH THIS.** I measured it: in jsdom the textarea reports `disabled: true`
  mid-flight while `document.activeElement` stays the textarea. The browser run is the only place the
  focus theft is observable, so do not treat a green suite as evidence here.

**2. THERE ARE *THREE* CLOBBER SITES, NOT THE TWO THE INTENT NAMES — AND THE THIRD IS THE ONE THAT
ACTUALLY FIRES.** The intent names `submit`'s early return and its `setDraft(...canonicalValue...)`
reconcile. There is a third, one screen above them:

```
App.tsx:1845-1846
  const [lastCommitted, setLastCommitted] = useState(committed)
  if (lastCommitted !== committed) { setLastCommitted(committed); setDraft(inherited(committed)) }
```

When a commit is accepted the projection comes back, `committed` becomes the text that was sent, this
transition sees it change, and it **overwrites the draft with the engine's echo — whatever the author has
typed since.** It fires regardless of `reconcileDraft`, so fixing only `submit:1880` does not close the
hazard.

**I reproduced this at `c13864c`, before any debounce exists.** Typing `Invoice`, blurring, typing
`Invoice 2026` while the command was still in flight, then releasing the engine, left the field reading
**`Invoice`** — the later text was destroyed. That is the story's stated risk, already live on the blur
path; the debounce only makes it ordinary. **Any fix must cover this site too, and the test that proves it
must let the engine echo an OLDER value while a NEWER draft is held.**

⚠ Note for whatever you write as a test helper: the `request` mock used by the existing prose tests
(`App.test.tsx:2209`) returns a snapshot with **no `canvas`**, so `applyProperties` returns `undefined`
and neither reconcile path is ever reached. Every existing test in that block is blind to all three
clobber sites. **You need a mock that echoes a canvas carrying the accepted value.**

**3. THE PARTIAL EXPRESSION IS *REFUSED* BY THE ENGINE. THE DECISION IS MADE; IMPLEMENT AND ASSERT IT.**
I ran the real Go command path (`ApplyComponentCommand`) at `c13864c` over a ladder of half-typed text.
The result, measured, not reasoned:

| Text typed so far | `contentCommand` routes to | Engine's answer |
|---|---|---|
| `{` | `value` | **accepted** |
| `{{` | `expression` | **REFUSED** — `component properties did not pass format validation` |
| `{{c` / `{{cust` | `expression` | **REFUSED** |
| `{{customer.name` | `expression` | **REFUSED** |
| `{{customer.name}}` | `expression` | **accepted** |
| `}}` | `expression` | **accepted** |

The browser and Go predicates agree exactly (`App.tsx:1654` is `includes('{{') || includes('}}')`;
`component_commands.go:1298` is the same two needles), so the routing is never wrong — but an author
typing an expression now crosses **four refusals** on the way to a valid one, where blur-only commit
crossed none.

**DECIDED BEHAVIOUR — send it, and let the existing refusal path render.** This is what the frozen matrix
already says (*"Engine refuses → the existing error path renders; typing is not blocked"*), and it is the
option that adds **no duplicated grammar** across the Go/TypeScript boundary. Suppressing the send would
mean re-implementing the engine's placeholder grammar in the browser, which is a mirrored invariant
needing its own ruling (D-7.4.5) and this story does not carry one.

What you must therefore assert: typing `{{cust`, letting the debounce fire, and finding the existing
`role="alert"` rendered; **the field still accepts typing** while it is shown; and on completing the text
to a valid `{{...}}` the next debounce succeeds and **the alert clears itself** (`applyProperties` calls
`setPropertyError(undefined)` on entry, `App.tsx:752`). The self-clearing half is the part that makes
this acceptable to an author, so it is not optional.

### One pre-existing defect this story's AC4 forces you to fix

**AC4 says "Given Escape mid-typing … then no command is sent." AT `c13864c`, ESCAPE ON A *FOCUSED* PROSE
FIELD SENDS A COMMAND AND DOES NOT REVERT.** Measured: focus the field, type, press Escape — one command
leaves carrying the unwanted draft, and the field ends up still showing it.

The cause is at `App.tsx:1990`: `revert()` schedules `setDraft`, then `event.currentTarget.blur()`
dispatches blur **synchronously**, and `shared`'s `onBlur: () => void commit()` runs against the
**pre-revert** `draft` from the current render — so `draft !== committed` is still true and it commits.

**The existing test does not catch this because it never focuses the field.** `App.test.tsx:2302`
(*"reverts and blurs on Escape"*) calls `fireEvent.change` and then `keyDown`; the element was never
focused, so `.blur()` is a no-op, no blur event fires, and `expect(sent).toHaveLength(0)` passes
vacuously. Do not cite it as cover.

So Escape must suppress the commit that its own blur triggers, in addition to cancelling the debounce
timer. This restores the invariant the frozen Boundaries already assert (*"Escape still reverts"*) rather
than changing it. **Add a focused Escape test; leave the existing unfocused one alone.**

## Tasks & Acceptance

**Execution:**
- [x] Debounce the prose field's draft and commit through the existing `submit`, cancelling on blur,
      Escape and unmount.
- [x] **Fix the two collision hazards** — a debounced commit must not be dropped by the in-flight guard,
      and a late echo must not overwrite a newer draft. **These are the story, not a detail of it.**
- [x] Decide and assert the partial-expression behaviour.
- [x] Tests — one per matrix row, driving the real path in a mounted designer with a fake timer.
- [x] A browser run: type into the field without leaving it and photograph the canvas following.

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
