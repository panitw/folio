---
title: 'Story 17.2: Clicking beside the page keeps the selection'
type: 'refactor'
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

Click the grey area beside the page and whatever you had selected is dropped, and the inspector swaps
from the element's properties to the page's. It is easy to do by accident — reaching for a scrollbar,
or just missing.

After this story that click does nothing. Escape still deselects, and so does clicking the page itself.

<intent-contract>

## Intent

**Problem:** `App.tsx:1455`'s `onClick` on `<main className="canvas-region">` clears the selection when the
click lands on the `<main>` itself. That is the empty space around the page, where a click is far more
often a miss than an instruction.

**Approach:** Stop clearing there. Leave every other route to an empty selection alone.

## Boundaries & Constraints

**Always:**
- **THE SCOPE IS THE BACKDROP ONLY** — the `event.target === event.currentTarget` branch of the
  `canvas-region` click. **`App.tsx:1415`'s page-surface click stays**, and clicking the page still
  clears. That is a different surface and the owner asked about the space *around* the canvas.
- **ESCAPE MUST STILL DESELECT**, and after this story it is the deliberate way to. It is on the same
  line (`:1455`) and is **not** gated on the target test — leave it exactly as it is.
- **A TEST ASSERTING THE OLD CLEAR IS A GUARD TO CORRECT, NOT TO DELETE.** `App.test.tsx:804-808`
  asserts a click removes both resize handles and sends no engine request — **verify which surface it
  clicks before touching it**; it names the page, which this story does not change.
- **Selection is taken on pointerdown, not click** (`App.tsx:2365`), with `selectedByPointer` suppressing
  the duplicate. **Do not "simplify" that while here.**
- **No engine traffic either way.** Selection is designer-side state; this story sends no command and
  must be asserted to send none.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Changing `App.tsx:1415`, the page-surface clear.
- Removing or re-gating Escape-to-deselect.
- Any change to how a click reaches a component (`:2377`) or how `revokeTableEditor` is sequenced.

**Never:** an unreachable selection state · a click on the backdrop that sends a command · dropping the
table editor's revocation where it is still required · leaving no way to deselect.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Click the backdrop | One component selected | **Selection unchanged**; inspector still shows its properties | No command |
| Click the backdrop | Nothing selected | Nothing happens | No command |
| Click the page surface | One component selected | Selection cleared, as today | No command |
| Escape in the canvas region | One component selected | Selection cleared, as today | — |
| Click another component | `e1` selected, click `e2` | Selection becomes `e2`, as today | — |
| Shift-click a component | `e1` selected | Extends, as today | — |
| Backdrop click with the table editor open | Editor open | **Decide and assert** whether the editor still revokes | — |
| Backdrop click | — | The inspector does **not** swap to `PageSetup` (`:1464` keys off `selected.length`) | — |

</intent-contract>

## Code Map

Anchors at `85f213f`. **Re-verify — these are long lines.**

- `App.tsx:1455` — `<main className="canvas-region">`: the click handler
  `if (event.target === event.currentTarget) { revokeTableEditor(); setSelected([]) }`, **and on the same
  line** the `onKeyDown` whose Escape branch clears. **Only the click branch is in scope.**
- `App.tsx:1415` — `<section className="page-surface">` `onClick={() => { revokeTableEditor(); setSelected([]) }}`,
  unconditional. **Out of scope, and it is why "click the page" still deselects.**
- `App.tsx:1464` — renders `PageSetup` when `selected.length === 0`; this is what the author sees change.
- `App.tsx:2365`, `:2377` — pointerdown selects, click stops propagation.
- `App.tsx:134`, `:731`, `:732`, `:1227`, `:1343` — the other `setSelected` sites. **All out of scope**;
  listed so a reviewer can see the population is seven call sites and this story touches one.

## Tasks & Acceptance

**Execution:**
- [ ] `App.tsx:1455` — stop clearing the selection on a backdrop click.
- [ ] Decide the table editor's fate on that click and assert it either way.
- [ ] Tests — one per matrix row, **including the ones that must NOT change**, so the story proves it
      narrowed one branch and not the mechanism.
- [ ] **Red-prove the new guard** by restoring the clear, watching it fail, restoring by absolute path.
      **Say where the failure lands.**

**Acceptance Criteria:**
- Given a selected component, when the backdrop is clicked, then it is **still selected** and the
  inspector still shows its properties.
- Given a selected component, when the page surface is clicked, then it is deselected, **as before**.
- Given a selected component, when Escape is pressed in the canvas region, then it is deselected.
- Given any of the above, then **no engine request was made**.

## Design Notes

**The guard that used to justify the clear is what makes it wrong.** `event.target === event.currentTarget`
was written so a click on the toolbar or the sheet stack would not deselect. It succeeds at that and
therefore fires on precisely the region where a click is most likely to be a miss.

## Verification

**Commands** — one per line, **exit codes read from `$?` immediately, never through a pipe**; zsh, so
`${PIPESTATUS[0]}` is wrong.
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts`'s corpus
  scan (DW-152); **match it by NAME, not count.**
- `npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`. **No Go is touched; say so rather than running it.**

**A BROWSER RUN.** Select a component, click beside the page, photograph the inspector still showing its
properties. `chromium-1217` via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (`chromium-1208` is a stub, DW-180).

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**8 rows**);
state the population beside every zero, pair every absence claim with a positive control; **a comment is
not a measurement**; **use `/usr/bin/grep`** — recursive `grep` returns false zeros here.
