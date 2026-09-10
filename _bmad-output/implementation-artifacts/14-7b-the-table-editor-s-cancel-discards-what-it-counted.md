---
title: "Story 14.7b: The table editor's Cancel discards what it counted"
type: 'feature'
created: '2026-09-10'
status: 'done'
baseline_commit: '6c643ac18aeb616d2d099538d89eac804eb55e11'
review_loop_iteration: 0
context: []
---

# Story 14.7b: The table editor's Cancel discards what it counted

**In plain terms.** *Not normative — the frozen Intent below governs. Rewritten at close to say what
happened.*

The table editor now has a **Cancel**. Before this, every change was saved the moment you looked away
from the box, and the only way out was to undo each edit yourself, counting as you went. Cancel counts
for you and closes. **Done** closes and keeps your work, and so does Escape.

Two things will look like mistakes and are not. A change that did not really change anything is not
counted, so Cancel cannot reach back past the moment you opened the dialog. And Escape keeps rather
than discards, because Cancel switches off when there is more to unwind than the product can promise,
and a dialog you could not leave by keyboard would be worse.

The main window's keyboard shortcuts also no longer fire through the open dialog. They had been nudging
and duplicating the table you were editing.

The review found one real fault, which this story created, in two halves: leaving the dialog mid-unwind
could tear the unwind down part-way, and the recovered document was not being put back on screen. Both
were fixed before anything shipped.

Still open: the font browser keeps the same shortcut leak (`DW-371`), and the two dialogs' `Cancel`
buttons will mean different things until someone rules (`DW-370`). Three notes record where this work
is argued rather than tested (`DW-372`–`DW-374`), and the new shortcut guard was proved only in a
simulated browser, never a real one.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The table editor commits every cell on blur, so opening it to try something is a
commitment the author must unpick by hand — there is no discard. Separately, the global keyboard
shortcut fires from inside the open modal (`DW-368`): measured at `68aa91f`, `Cmd+Z` sends an `undo`
**and destroys the dialog**, while `Cmd+D` and the arrow-nudge send `duplicateComponent` and
`moveComponent` against **the very table the dialog is editing** and leave the dialog open — so the
dialog and the document disagree about what is being edited, and any edit counter would be
desynchronised by a history entry it never saw.

**Approach:** Per `D-14.7.1`, the footer's one `Close Table Editor` button becomes **`Cancel` /
`Done`**. Commit-on-blur is unchanged: no engine change, no new command, no `AD-15` change. The
application keeps an **integer count** of the dialog's commands that *actually changed the document*
(`revision !== priorRevision`); `Cancel` issues exactly that many `undo` operations and closes,
`Done` closes. **This is a compensating sequence replaying the engine's own byte snapshots — not a
transaction.** The global shortcut is made to ask *"is a modal open?"* rather than *"is the target
editable?"*, which is the question `DW-368` names as the root cause.

**Escape closes and keeps, like the font browser; `Cancel` is the only discard.** A modal whose
Escape is not its Cancel is correct here and still surprising, so it is stated rather than left to be
read as a bug. It is forced by AC5: if Escape meant Cancel, then the modal would stop being
keyboard-dismissible in exactly the state where Cancel is disabled.

## Boundaries & Constraints

**Always:**
- **Count only real mutations** (`revision !== priorRevision`), scoped to the current
  `tableEditorSession`. This is `D-14.7.1`'s guardrail 1 and the only guard against the single
  destructive failure mode: a no-op that consumed a Cancel step would make Cancel unwind edits from
  **before the dialog opened**.
- **The unwind runs while the dialog is still open.** AC2 says Cancel "issues … and closes"; AC6 says
  a failed undo makes the dialog "stop, state the actual position". A closed dialog can state
  nothing, so **closing is the success path only**.
- **The bound `N ≤ 100` is sound ONLY because the modal-open guard exists.** It assumes the dialog's
  own commands are the *only* source of undo entries while it is open. That is **false at `68aa91f`** —
  the arrow-nudge and `Cmd+D` push entries the dialog never counted — and the guard is what makes it
  true. The count is correct **because no other path can commit while the modal is open**, not
  because 100 is a large number. A bound stated without its precondition is a number, not a guard.
- **The modal-open condition keys on a state, never on a list of keys.** A key-list guard is the
  shape that produced this defect.

**Ask First:**
- Widening the modal-open guard to `fontBrowserOpen`. Ruled **out of this story** (Q1 = B′); the
  owner holds the widening as a separate non-blocking question. Proceed on the table editor alone
  regardless of how that comes back.
- Any change that makes the count come from the engine rather than from the application, or that adds
  a history-depth field to the wire. Ruled out: `D-14.7.1` forbids an engine change.
- Any change to `FontBrowser.tsx`, or any attempt to reconcile the two dialogs' `Cancel` semantics
  (`DW-370`). Filed, unassigned, **not this story's work**.

**Never:**
- Never build a local uncommitted buffer or a second document model. The application holds an
  **integer**, which is transient interaction state by `AD-15`'s own definition.
- Never implement the shortcut swallow inside `TableEditor.tsx`. `trapDialog` is a React synthetic
  handler and the shortcut is a native `window` listener; `TableEditor.tsx:141-148` records a live
  case — the last column removed — where focus sits on `document.body` and the trap sees no key at
  all. A swallow asserted only at the dialog is a guard never invoked.
- Never clear the engine's redo stack, and never claim a discard is permanent: `Undo()` pushes onto
  redo, so the discarded edits stay redoable **until the author's next committed edit**, which clears
  it (`e.redo = nil`).
- Never wrap the `Cancel` / `Done` pair in `role="group"`, and never edit the pinned string at
  `control-vocabulary-contract.test.tsx:885-889` to accommodate a wrapper.
- Never change `Cmd+S`, which sits above the guard line deliberately.
- Never widen this story to the pre-existing ring-buffer eviction (see Design Notes) — disclose it.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Discard a session | Dialog open; 3 blurs each changing the document; `Cancel` | Exactly 3 `undo` operations sent, in sequence; dialog closes; document is byte-identical to dialog-open state; a status line states 3 edits discarded and that Redo restores them until the next edit | N/A |
| Keep a session | Dialog open; 3 changing edits; `Done` | **Zero** `undo` operations; dialog closes; edits stand | N/A |
| Escape | Dialog open; 3 changing edits; `Escape` | Identical to `Done` — closes and keeps, zero `undo` operations | N/A |
| A no-op is not counted | `Clear Header text colour` (`×`) clicked on an already-unset field, twice, plus 2 changing edits; `Cancel` | Engine returns an unchanged `revision` for the two clears, so the count is **2**, not 4; exactly 2 `undo` operations; document returns to dialog-open state, **not two edits earlier** | N/A |
| A rejected command is not counted | Dialog open; a command the engine refuses | Count unchanged; the existing error surface is unchanged | Existing `tableEditorError` path |
| Nothing to discard | Dialog open; no changing edit; `Cancel` | Zero `undo` operations; closes; no discard statement | N/A |
| Count reset per session | 2 changing edits, `Done`, reopen, `Cancel` | Zero `undo` operations — the count is a function of the session, not of the document | N/A |
| Over the history limit | Count is 101 | `Cancel` is **disabled** and a visible note states why (101 edits exceed the engine's 100-step history, so a discard would land part-way); `Done` and `Escape` stay available | N/A |
| At the history limit | Count is exactly 100 | `Cancel` is **enabled**; 100 `undo` operations land exactly at dialog-open state | N/A |
| An undo in the sequence fails | Count 4; the 3rd `undo` returns `UNDO_UNAVAILABLE` | The loop **stops**; the dialog **stays open** and re-projects; it states the real position — 2 of 4 discarded, the other 2 stand — and claims no completed discard | `ErrNoUndo` → `UNDO_UNAVAILABLE`; mapped to the dialog's own error surface, never thrown away |
| Cancel while a command is in flight | `tableEditorBusy` true | `Cancel` disabled, so the count cannot move under the loop | N/A |
| Undo/redo shortcut, dialog open | `Cmd+Z`, `Ctrl+Z`, `Shift+Cmd+Z`, `Ctrl+Y`, focus on a dialog **button** (not an input) | No `undo`/`redo` operation sent; document unchanged; dialog **stays open** | N/A |
| Duplicate shortcut, dialog open | `Cmd+D`, focus on a dialog button | No `duplicateComponent` command. Today it duplicates the table the dialog is editing, leaving the dialog projecting one table while the document holds two | N/A |
| Nudge shortcut, dialog open | `ArrowLeft`/`Right`/`Up`/`Down`, focus on a dialog button | No `moveComponent` command; the matrix's own roving navigation still moves focus | N/A |
| Snap and preview shortcuts, dialog open | `Alt+S`, `Alt+P`, focus on a dialog button | Both suppressed while the modal is open | N/A |
| Save shortcut, dialog open | `Cmd+S` | **Unchanged** — still saves. It sits above the guard and is not a mutation of what the modal edits | Existing save path |
| The guard does not over-reach | No dialog open; `Cmd+Z` from a canvas selection handle | Undo still reaches the window handler, exactly as today | N/A |

</frozen-after-approval>

## Code Map

Anchors re-measured by me at `68aa91f`; re-verified unmoved at `d298766` (`git diff --name-only 68aa91f..d298766`
returns nothing outside `_bmad-output/`, and Story 14.7b's `epics.md` section is byte-identical across the
range by `shasum`).
Nothing here was carried from the epic or from a register without checking.

**The count's seam — already built, needs one line**

- `folio-designer/src/App.tsx:971-995` `commitTableColumn` — **the only route by which the dialog
  mutates the document.** It already holds `const revision = snapshotRef.current?.revision` (`:975`)
  and already compares `committed.snapshot.revision !== revision` (`:987`) to decide
  `invalidatePreview()`. Guardrail 1 is an increment at that existing comparison, scoped by the
  `session` it already captured at `:977`.
- `folio-designer/src/App.tsx:392` `tableEditorSession = useRef(0)`. **Exactly three sites advance
  it** — `:950` (`++` in `openTableEditor`), `:960` (`++` in `revokeTableEditor`), and `:1968` (`++`
  inside `setCurrentSnapshot`'s `clearDocumentInteraction` branch). The count must be zeroed at all
  three; that is the enumerable invariant, and "a reopened dialog starts at zero" is its test.
- `folio-designer/src/App.tsx:1968` `setCurrentSnapshot` — one line. Its third argument
  `clearDocumentInteraction` runs `documentGeneration.current++`, `tableEditorSession.current++`,
  `setSelected([])`, `setTableEditor(undefined)`, `setTableEditorError(undefined)` and
  `setFontBrowserOpen(false)`. **This is why `Cmd+Z` destroys the dialog today, and why the Cancel
  loop must call it once at the end rather than once per iteration.**
- `folio-designer/src/App.tsx:1968` also carries **the storage precedent for the count**:
  `documentGeneration` is a ref with `setDocumentGenerationValue` as its state mirror, written
  together. The count needs the same shape — a ref the loop can trust, a mirror the footer can render.
- `folio-designer/src/App.tsx:963-970` `closeTableEditor` — `revokeTableEditor()` plus a
  `queueMicrotask` focus restore to `tableEditorInvoker.current` else `canvasRegionRef`. Cancel's
  success path owes the same focus restore.
- `folio-designer/src/App.tsx:2070-2086` `applyHistory` — **read it, do not reuse it as the loop
  body.** Its `undoAvailable` guard (`:2071`) is React state, so across a synchronous N-iteration loop
  it is stale-but-permissive; and it fires `invalidatePreview()` + `setCurrentSnapshot(…, true)` on
  **every** iteration, which would tear the dialog down after undo #1 and make AC6 unimplementable.
  Its `catch` (`:2081-2085`) **is** the right precedent for AC6: it maps `UNDO_UNAVAILABLE` to state
  rather than to a thrown failure.
- `folio-designer/src/App.tsx:2552` — the design-mode announcement region, already
  `role="status" aria-live="polite"`. The discard statement belongs here, in its own state, **not**
  overloaded onto `fileStatus`.

**The shortcut hole (AC4) — one condition, at one line**

- `folio-designer/src/App.tsx:2181-2205` — the `useEffect`. `:2182` opens the handler; **`:2203` is
  `window.addEventListener('keydown', shortcut)`** (verified verbatim; `DW-368`'s anchor is exact).
  `:2205` is `})` — **no dependency array**, so it re-attaches every render and there is no
  stale-closure risk in any fix.
- `folio-designer/src/App.tsx:2191` `if (editing) return` — **the guard line.** Below it, in order:
  undo `:2192`, redo `:2193`, `Cmd+D` duplicate `:2194`, arrow-nudge `:2195`, `Alt+S` snap `:2196`,
  `Alt+P` preview `:2197-2201`. Above it, deliberately: `Cmd+S` save `:2186-2190`.
- `folio-designer/src/App.tsx:4877-4880` `isEditableTarget` — `INPUT`, `TEXTAREA`, `SELECT`,
  contenteditable. **Exactly one call site, `:2183`** (verified with a positive control). Its shape is
  the root cause: it enumerates *editable* elements when the question is *"is a modal open?"*.
- `folio-designer/src/App.tsx:330` `tableEditor` state — truthy iff the dialog renders (`:2619`).
  This is the modal-open state B′ keys on. `fontBrowserOpen` (`:372`) is **out of scope**.
- **Why `Cmd+D` and the nudge are worse than an uncounted history entry** — both gate on
  `selectedRef.current.length === 1`, and `openTableEditor` (`:947`) *requires* that the single
  selection **be the table being edited**. Measured in my probe: the leaked commands were
  `moveComponent id:"e7" x:22.276` and `duplicateComponent id:"e7"` — `e7` is the table the dialog had
  open. So the nudge moves the subject of the dialog, and the duplicate leaves the dialog projecting
  one table while the document holds two.

**The dialog (AC1, AC5, AC6)**

- `folio-designer/src/TableEditor.tsx:476` — the footer bar, one line: an
  `<output aria-label="Column summary" aria-live="off">` carrying `{n} columns · {m} aggregates`, and
  one `<button className="file-button" onClick={onClose}>Close Table Editor</button>`. The summary
  stays; the button becomes the pair.
- `folio-designer/src/TableEditor.tsx:461-475` — the signed disclosure *"⚠ THIS BAR CARRIES NO Cancel /
  Done PAIR AND NO EDIT COUNTER"*, naming this story, `D-14.7.1` and `DW-368` as its owners.
  **It must go when the pair lands**, or it becomes a stale assertion of the kind this run keeps
  finding.
- `folio-designer/src/TableEditor.tsx:10` `Props` — one line, 8 data props and 11 `on*` callbacks, all
  `=> void`. The dialog **cannot observe a revision** and must not learn to: the count is the
  application's, and `TableEditor.tsx` never reads `projection.revision` today (verified with a
  positive control). Add `editCount` and `onCancel`; keep `onClose` as the Done/Escape act.
- `folio-designer/src/TableEditor.tsx:170-177` `trapDialog` — `:171` Escape → `onClose()`,
  ungated on `busy`. Escape keeps that binding (Escape = Done). `:173`'s focusable query is
  `'button:not([disabled]), input:not([disabled]), select:not([disabled])'`, so **a disabled `Cancel`
  drops out of the trap's list** and the wrap ends move again — the same deliberate re-ordering
  `D-12.3.2` and Story 14.7 each made once.
- **The reachable no-op, which is what AC3 must be driven through.** `TableEditor.tsx:271` and `:284`
  — the `×` `Clear {label}` buttons — are `disabled={busy}` only, **never disabled on "already
  unset"**. Clicking `Clear Header text colour` on an unset field sends a legal `clear` command that
  leaves canonical bytes unchanged, so `bytes.Equal` short-circuits and the revision does not move.
  This is a shipped gesture, not a constructed one. (The blur handlers at `:369`, `:380`, `:387`,
  `:393`, `:429`, `:434` and `commitStyleText` `:232` all compare old against new first, and align
  `:388` compares too since `D-14.7.3`, so the `×` buttons are the vector that remains.)
- `folio-designer/src/App.css:735-739` — `.table-editor-footer` is
  `display:flex; align-items:center; justify-content:space-between`, and its comment already names
  this story. Two buttons plus a note need an actions wrapper; `space-between` across four children
  would scatter them.

**The mirrored bound (AC5) — read-only Go evidence**

- `folio-go/wasm/engine.go:29` `const historyLimit = 100`. `appendBounded` (`:363-371`) **evicts the
  oldest entry** when full rather than refusing — so over-running the limit is silent.
  `historyLimit` appears in Go source at exactly two places, `:29` and `:365`, and **nowhere in
  TypeScript**. Read-only: do not edit any Go file.
- `folio-go/wasm/engine.go:296-298` — `if bytes.Equal(canonical, e.bytes) { return e.Snapshot(), nil }`,
  returning **before** `pushUndo` (`:311`), **before** `e.redo = nil` (`:312`) and **before**
  `install` (`:313`), which is the sole site of `e.revision++` (`:358`). **A no-op is not a history
  entry**, which is what guardrail 1 rests on.
- `folio-go/wasm/engine.go:321-329` `Undo()` — empty stack returns `ErrNoUndo` (`:31`); otherwise it
  calls `pushRedo(e.bytes)` (`:327`) **before** restoring, so N undos leave N redo entries and AC7 is
  structurally free. Revisions are **monotonic across undo** (`install` only increments), so the UI
  **cannot** detect "back where I started" by comparing them.
- `folio-go/wasm/cmd/engine/main.go:180-185` and `:246-248` — the `undo` op takes no payload and maps
  `ErrNoUndo` to `{ code: "UNDO_UNAVAILABLE", message: "Nothing to undo" }`. **The failure envelope
  carries no snapshot**, so a failed undo yields no revision at all.
- `folio-designer/src/engine-protocol.ts:5-13` — where the `MAX_ENGINE_*` constants live; the new one
  joins them. `:272-273` `canUndo?`/`canRedo?` are bare booleans, closed-shape validated at `:718`,
  which is why the limit cannot be read over the wire. **No protocol change is required for an
  application-side count.**
- `folio-designer/src/engine-bounds-mirror.test.ts` — 7 `describe` blocks; `goSources` (`:44-56`) has
  **five** entries and `folio-go/wasm/engine.go` is **not** among them. Each block carries four
  assertion kinds: a **non-vacuity `it` first**, agreement, a **`sites` regex** proving the TS
  constant is consumed at its validator, and an in-memory **red-proof** per side. Its own comment says
  why the `sites` column exists: without it "a hoisted constant could sit in the file unused while the
  validator kept a stale inline literal, and the tie would pass".

**Tests that will red, with why**

- `folio-designer/src/TableEditor.test.tsx:260` — `getByRole('button', { name: 'Close Table Editor' })`
  throws.
- `folio-designer/src/App.test.tsx:311-360` — the trap test. `:355` names `Close Table Editor`; `:356`
  asserts it is the trap's **last** focusable; `:357-360` assert **both wrap ends**. All move. Its
  comments at `:321` and `:349-350` name the old order and go stale with it.
- `folio-designer/src/App.test.tsx:384` — clicks `Close Table Editor` to close. This must become
  **`Done`**: a `Cancel` here would issue undos and change what the test measures.
- `folio-designer/e2e/table-editor.spec.ts:72` — asserts `Close Table Editor` visible; `:71` is its
  stale comment. `:73-74` (Escape closes) stays green **and stays true** under Escape = Done.
- `folio-designer/e2e/browser-native-roundtrip.spec.ts:275` — `Close Table Editor` inside
  `authorTableWithFooter`. **Must become `Done`.** A `Cancel` there would undo the five columns and the
  footer aggregate the round trip has just authored, and it would still pass as a rename while
  destroying the only end-to-end proof that authoring reaches the engine.
- `folio-designer/src/TableEditor.test.tsx:70-101` `tableEngine` — the mock returns
  `revision: ++state.revision` for **every** `command`, so it **cannot express a no-op** and AC3 is
  untestable against it as written.

**Tests that will NOT red, checked so nobody chases them**

- `control-vocabulary-contract.test.tsx` floors (`:591-594`: `PER_STATE_CONTROL_FLOOR 15`,
  `CONTROL_FLOOR 220`, `CLASS_FAMILY_FLOOR 18`, `GROUP_INSTANCE_FLOOR 33`) are **floors**, and
  `Close` → `Cancel` + `Done` is net **+1** control. Both are words in the existing `.file-button`
  family, so R1 (spelling uniformity) and R4 (the glyph census) are satisfied. ⚠ **But `:885-889`
  pins the exact string `'R0 the sweep visited 32 group instances, under the floor of 33'`. A
  `role="group"` around the pair takes that shrunk count to 33, clears the floor, and reds the
  `toEqual`.** Use a plain `<div>`; say so in a comment.
- `canvas-authority-contract.test.ts` — scans `src/**` production, `src/**/*.test.*` **and `e2e/**`**,
  waiving exactly one file (`e2e/e9-5-border-no-ink.spec.ts`). Every new test here asserts on roles
  and `request.mock.calls`, never on geometry, so it stays green — but a `getComputedStyle`,
  `scrollLeft` or `getBoundingClientRect` in a new test file would red `npx vitest run` immediately.
- `command-json-soleness.test.ts` — the count sends `undo` operations, not command JSON.
- `design-contract.test.ts` — only via CSS: a raw hex or a literal `border-radius` in new
  `.table-editor-footer` rules would red it. Use tokens.
- `App.test.tsx:1267-1276` — "does not route Undo through an editable %s". Green, and **the natural
  home** for the new modal-swallow guard: same describe, same idiom.
- ⚠ `App.test.tsx:6279-6287` — *"A MODIFIER IS THE APPLICATION'S. Undo must still reach the window
  handler from a focused handle."* Green, **and it is the counter-example the guard must not
  generalise over.** A swallow that suppresses undo whenever a `window` listener sees a modifier, or
  that keys on anything other than the modal being open, reds this. Do not weaken it.
- No e2e spec presses undo (verified with a positive control), so no browser spec covers AC4.

## Tasks & Acceptance

**Operational constraints for the implementer (`D-14.4.1` — these are binding):**

> **You must never commit.** You may not run `git commit`, `git add`, `git stash`, `git checkout`,
> `git reset`, `git revert`, `git restore`, `git clean`, `git branch`, `git merge`, `git rebase`,
> `git push`, or `git tag`. The orchestrator makes every commit in this run. Reading git state —
> `git log`, `git status`, `git show`, `git diff`, `git ls-files` — is permitted and encouraged. The
> single carve-out: `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`,
> removed in a `finally`, is permitted for a throwaway test harness.

- **Do not edit** `sprint-status.yaml`, `deferred-work.md`, `epics.md`, `DESIGN.md`, or anything under
  `fixtures/declared-variants/`. Do not run `code -r`. **Do not edit any file under `folio-go/`** —
  every Go anchor here is read-only evidence.
- Shell is **zsh**: `${PIPESTATUS[0]}` is empty; use `${pipestatus[1]}`, or `cmd > log 2>&1; echo $?`.
  Never `$?` after a pipe. Quote every glob and variable.
- `find` wraps `bfs` and is non-deterministic here — use `git ls-files`. A `find` miss is never
  evidence of absence, and neither is an empty `grep`: run a positive control beside every absence
  claim and report both.
- `App.tsx` holds two NUL bytes near line 3997: plain `diff` prints "Binary files differ" with zero
  changed lines. Use `diff -a` or `cmp`.
- Diff test names as a **multiset**, not a set — report GONE and NEW, never a delta.
- **Re-measure the vocabulary floors and `:885-889`'s expected list by EXECUTING the sweep** if you
  touch them at all. Do not hand-write either, and do not edit the pinned string to accommodate a
  wrapper — use a plain `<div>` instead.

**Execution:**

- [x] `folio-designer/src/engine-protocol.ts` -- add `MAX_ENGINE_HISTORY_ENTRIES = 100` beside the
      existing `MAX_ENGINE_*` constants, with a comment stating it is a **mirror of a Go constant**
      (`wasm/engine.go:29`), that the engine enforces it as a **ring buffer that silently evicts the
      oldest entry** rather than as a refusal, and that it is tied in `engine-bounds-mirror.test.ts`.
      No type, guard or validator changes — the count never crosses the worker boundary.
- [x] `folio-designer/src/App.tsx` -- hold the edit count: a ref the Cancel loop can trust plus a
      state mirror the footer renders, written together, following `documentGeneration` /
      `setDocumentGenerationValue` at `:1968`. **Increment only where `commitTableColumn` already
      compares `committed.snapshot.revision !== revision` (`:987`)**, scoped to the `session` captured
      at `:977`. **Zero it at all three sites that advance `tableEditorSession.current`** — `:950`,
      `:960`, `:1968`.
- [x] `folio-designer/src/App.tsx` -- add the Cancel routine. It must **not** reuse `applyHistory`
      (stale `undoAvailable` across iterations; `setCurrentSnapshot(…, true)` per iteration would
      unmount the dialog after undo #1). It awaits `engine.request('undo')` N times with the dialog
      still open, installs the reached snapshot **once** on success and then closes with the same
      focus restore `closeTableEditor` performs; refuses to start while `tableEditorBusy`, and refuses
      a count above `MAX_ENGINE_HISTORY_ENTRIES`. On a failed undo it **stops**, re-projects
      `table-columns`, and states the real position through the dialog's own error surface — mapping
      `UNDO_UNAVAILABLE` to state as `applyHistory:2081-2085` already does.
- [x] `folio-designer/src/App.tsx` -- state the discard after a successful Cancel, in a new state
      rendered by the existing `role="status" aria-live="polite"` region at `:2552`. It must name the
      number discarded **and** that Redo restores them **until the author's next edit** — the honest
      limit, because the next committed command clears the engine's redo stack.
- [x] `folio-designer/src/App.tsx:2191` -- the AC4 guard: `if (editing || <the table editor is open>)
      return`, keyed on the **open-modal state**, never on a list of keys. Do **not** reference
      `fontBrowserOpen`. `Cmd+S` stays above the line, untouched. Say in a comment that this answers
      *"is a modal open?"* where `isEditableTarget` answers *"is the target editable?"*, and that the
      count's `N ≤ 100` bound depends on this guard.
- [x] `folio-designer/src/TableEditor.tsx` -- replace the single `Close Table Editor` with the
      **`Cancel` / `Done`** pair inside a plain `<div>` actions wrapper (**no `role="group"`** — say
      why, naming `control-vocabulary-contract.test.tsx:885-889`). Both are words in the existing
      `.file-button` family. `Cancel` is disabled while `busy` and while `editCount` exceeds
      `MAX_ENGINE_HISTORY_ENTRIES`, and a **visible** note beside it states that reason — not a bare
      grey-out and not a `title`. Keep the `Column summary` `<output>` and keep Escape bound to
      `onClose`. **Delete the `:461-475` disclosure** and replace it with what the bar now does,
      including that Escape and `Done` are the same act and `Cancel` is the only discard.
- [x] `folio-designer/src/App.css` -- the actions wrapper and the disabled-reason note, in tokens
      only: no raw hex, no literal `border-radius`. Update the `:735-737` comment, which currently
      says this story is still pending.
- [x] `folio-designer/src/TableEditor.test.tsx` -- teach `tableEngine` (`:70-101`) to express a
      **no-op**: it must decide "changed" by comparing its own serialized state before and after
      `apply`, returning the **unchanged** revision when they agree, rather than incrementing on every
      command. Cite the rule at `App.test.tsx:3955-3957` — *the engine decides what is a mutation, not
      the UI.* Then prove: the counted-discard, the `Done`-keeps arm, Escape-keeps, the reachable
      no-op via `Clear Header text colour` on an unset field, count-resets-per-session, the
      disabled-Cancel reason, the failed-undo position, and the redo survival. Update `:260` for the
      new footer.
- [x] `folio-designer/src/App.test.tsx` -- update the trap test (`:311-360`): re-derive **both** wrap
      ends from the DOM as `:335` already does, and say in the test that the re-ordering is intended.
      Make `:384` click **`Done`**. Add the AC4 guard's proofs next to `:1267-1276` — one per
      suppressed shortcut (undo, redo, `Cmd+D`, each arrow, `Alt+S`, `Alt+P`), each asserting **no
      command reached the engine**, plus `Cmd+S` still saving. **Do not weaken `:6279-6287`.**
- [x] `folio-designer/src/engine-bounds-mirror.test.ts` -- a new describe tying
      `MAX_ENGINE_HISTORY_ENTRIES` to `wasm/engine.go`'s `historyLimit`, adding that file as a
      **sixth** `goSources` entry. All four assertion kinds: a **non-vacuity `it` first that fails if
      `wasm/engine.go` stops containing `historyLimit` at all** (a rename must red, not silently
      compare nothing); agreement; a **`sites` regex** pointing at the disable site that consumes the
      constant; and a red-proof that reds **by deleting the consumption**, not only by changing the
      number.
- [x] `folio-designer/e2e/table-editor.spec.ts` -- assert both `Cancel` and `Done` in the footer;
      fix the `:71` comment. Keep the Escape assertion and state that Escape is `Done`, not `Cancel`.
- [x] `folio-designer/e2e/browser-native-roundtrip.spec.ts:275` -- change to **`Done`**. Do not
      weaken the helper; a `Cancel` here would undo everything it just authored.

**Acceptance Criteria:**

- Given a dialog session with two changing edits and two `×` clears on already-unset fields, when
  `Cancel` is pressed, then exactly **two** `undo` operations are sent and the document is byte-equal
  to its dialog-open state — **proved by driving the clears through the real `×` buttons**, so the
  no-op arm is a shipped gesture rather than a mock-only construction, and by a control that shows the
  same test reds if the count increments on dispatch instead of on an observed revision change.
- Given the count, when the dialog is closed by `Done` and reopened, then `Cancel` issues **zero**
  undo operations — and no site that advances `tableEditorSession.current` leaves the count standing.
- Given a count of exactly `MAX_ENGINE_HISTORY_ENTRIES` and a count one above it, when the footer
  renders, then `Cancel` is enabled in the first case and disabled-with-a-visible-reason in the
  second, while `Done` and Escape remain available in both.
- Given a Cancel whose k-th undo returns `UNDO_UNAVAILABLE`, when the sequence stops, then the dialog
  is **still open**, its projection matches the document it actually reached, and its message states
  how many of how many were discarded — **proved by asserting the message names both numbers**, so a
  message that claims a completed discard cannot pass.
- Given a completed Cancel, when it has closed the dialog, then the engine reports `canRedo`, a redo
  restores the discarded edits, and the status line said so — including that the next committed edit
  ends it.
- Given the open dialog and focus on a **button** inside it, when undo, redo, `Cmd+D`, each of the four
  arrow keys, `Alt+S` and `Alt+P` are pressed, then **no operation and no command reaches the
  engine** and the dialog stays open — while `Cmd+S` still saves, and undo still reaches the window
  handler from a canvas selection handle with no dialog open.
- Given the matrix, when arrow keys are pressed on a matrix cell inside the open dialog, then the
  roving lattice still moves focus exactly as it does today — the guard suppresses the document
  mutation, not the dialog's own navigation.
- Given the rebuilt footer, when the dialog renders, then the focus trap still wraps at both ends over
  the re-ordered focusable list, **including when `Cancel` is disabled and therefore absent from that
  list**, and Escape still closes and restores the invoking control.
- Given the vocabulary sweep, when it runs, then `Cancel` and `Done` are swept as words in the
  `.file-button` family and `control-vocabulary-contract.test.tsx:885-889` is **unedited and green**.

## Spec Change Log

## Design Notes

**Why this is a compensating sequence and not a transaction, in one line each.** Nothing is buffered:
every edit is already in the document when the author sees it, so every engine refusal stays live and
located, which is Epic 14's whole subject. `Cancel` does not roll anything back — it asks the engine to
replay **its own byte snapshots** (`pushUndo(e.bytes)`), N of them. The UI holds an integer, which
`AD-15` names as permitted transient state, and a smaller footprint than the `draft` object
`PageSetup` already ships.

**The bound and its precondition are one fact, not two.** `N ≤ 100` is only sound because no other
path can commit while the modal is open. At exactly 100 the newest 100 undo entries are all the
dialog's own, so 100 undos land precisely at dialog-open state; at 101 the first has already been
evicted by `appendBounded`, so 100 undos land one edit short and the 101st returns `ErrNoUndo` — a
silent under-unwind, which is why AC5 disables rather than tries. **Before the AC4 guard, that
reasoning was false**: an arrow-nudge or `Cmd+D` pushes an entry the dialog never counted, so Cancel's
N undos would consume it and leave one of the dialog's own edits standing. Write the guard and the
bound as the same claim.

**The coupling claim is checkable, so here is the check rather than the assertion.** "No other path can
commit while the modal is open" has to survive an audit of every path, not just the shortcut. With the
AC4 guard in place the paths are: (1) the dialog's own controls, which all route through
`commitTableColumn` and are therefore counted; (2) the toolbar's `Undo`/`Redo` buttons (`App.tsx:2491`),
unreachable by pointer because `.table-editor-backdrop` is `position:fixed; inset:0; z-index:20`
(`App.css:647`) — a full-viewport overlay — and unreachable by Tab because `trapDialog` wraps at both
ends; and (3) the global shortcut, which this story closes. Path 2 has one known strand — the last
column removed leaves focus on `document.body`, where `trapDialog` sees nothing — and
`TableEditor.tsx:148-150` already repairs it by focusing the empty state's `Add column`. **Even if path
2 were reached, it would not desynchronise the count**: `applyHistory` passes
`clearDocumentInteraction`, which closes the dialog and destroys the count rather than corrupting it.
That asymmetry is the general shape worth carrying — **a path that tears the counter down is safe; a
path that leaves it standing is the dangerous one**, which is exactly why the nudge and the duplicate
mattered more than the undo.

**Disclosed, not fixed: the ring buffer can evict the author's pre-dialog history.** If a session is
already at 100 entries, the dialog's own commands push the oldest out. Cancel stays correct — it only
ever needs the newest N — but the author's ability to undo *past* the dialog is reduced by N. That is
a pre-existing property of `appendBounded`, not something this story introduces, and it is stated here
so it is not later discovered as a regression. **Do not widen the story to address it.**

**Why the swallow lives on the window handler and not in the dialog.** `trapDialog` is a React
synthetic `onKeyDownCapture` on the dialog element; the shortcut is a native `window` listener. Worse,
`TableEditor.tsx:141-148` records a live case — the last column removed — where focus sits on
`document.body` **inside** the open modal and `trapDialog` sees no key at all; that comment exists
because the case once broke Escape. A swallow asserted only at the dialog would be a guard never
invoked in exactly the state that already has a bug.

**Why `Cancel` disabled changes the focus trap, and why that is expected.** `trapDialog:173` selects
`button:not([disabled])`, so a disabled `Cancel` leaves the focusable list and the wrap ends move.
`D-12.3.2` re-ordered this list once and Story 14.7 re-ordered it again; this is the third, and like
both predecessors it must be re-derived from the DOM in the test and **declared intentional in
prose**, because an unexplained re-ordering reads as a regression to the next author.

**One vocabulary hazard is knowingly shipped, and it is recorded rather than solved.** After this
story the product has two modal `Cancel` buttons meaning different things — the font browser's closes
without undoing, this one discards. `D-14.7.1` rules the labels and is not reopened here; the
collision is filed as `DW-370`, unassigned. No contract test can see it: R1 checks spelling within a
class, never meaning.

## Verification

**Commands** — `D-000.33`'s per-story cadence, run from `folio-designer/`. Nothing heavier without an
explicit override.

- `npx vitest run` -- expected: **0 failures**. Baseline at `68aa91f`, CI green on all seven jobs:
  **76 files / 1320 tests / 0 failures**. Test names must be diffed as a **multiset** — report the
  GONE and NEW sets, never a delta.
- `npx tsc -b --force` -- expected: exit 0 and **0 bytes** of output. `--force` is mandatory.
- `npx oxlint` -- expected: exit 0, **0 errors**, and the warning **SET** below unchanged.

  **The baseline is a SET, never an integer** (`D-14.7.2`): `react(only-export-components)` in
  `src/App.tsx` ×2, `src/preview/pdf-viewer.tsx` ×2, `src/segmented-control.tsx` ×3.
  **Re-measure it; do not quote line numbers** — they move whenever `App.tsx` moves and were never
  the pin. **Explain any new key rather than re-baselining, and never change code to restore a
  count.** Nothing in this story adds a component export, so a new key is a finding.
- `npm run test:e2e:compile` -- expected: exit 0. **This is `tsc --noEmit` only. It is not a browser
  run and must never be reported as one.**

*Provenance:* the four baselines were measured at `68aa91f`. `05043a4`, `c92e673` and `d298766`
followed, and `git diff --name-only 68aa91f..d298766` returns nothing outside `_bmad-output/` — no code
moved, so the baselines stand. Re-measure anyway before trusting them.

**What these prove, and what they do not — state this plainly in the report:**

- **Proven by execution:** the counted discard and its no-op arm through the real `×` gesture; the
  per-session reset; both sides of the `MAX_ENGINE_HISTORY_ENTRIES` boundary; the failed-undo position
  message; redo survival; every suppressed shortcut and `Cmd+S`'s survival; the trap's wrap ends with
  `Cancel` present and absent; the Go↔TS tie with its non-vacuity control and its `sites` red-proof.
- **NOT proven, because jsdom applies no stylesheet and computes no layout:** that the four-child
  footer does not overflow the sheet, and that the disabled-reason note is legible beside `Cancel`.
  `e2e/table-matrix-layout.spec.ts:218-220` already measures the footer's box against the sheet and is
  **compiled here and executed only by CI**.
- **NOT proven at all:** no e2e spec presses undo (verified with a positive control), so **AC4 has no
  browser coverage** — its guard is proved only in jsdom, where the native `window` listener and React
  synthetic ordering are jsdom's, not Chromium's. Say so; do not call it covered.
- **Did not run, in these words:** the browser suite, the Go suites, the matrix legs, `npm run build`
  as a gate, the `verify:offline*` chain, and the font-host scans.

## Suggested Review Order

**The count, and why its bound is only sound with the guard**

- Start here: the increment rides an existing revision comparison, so a no-op cannot count.
  [`App.tsx:1068`](../../folio-designer/src/App.tsx#L1068)

- A ref for the loop, a state mirror for the footer, written together.
  [`App.tsx:415`](../../folio-designer/src/App.tsx#L415)

- The bound the dialog refuses at; consumed here, not inlined.
  [`TableEditor.tsx:176`](../../folio-designer/src/TableEditor.tsx#L176)

**The compensating sequence — not a transaction**

- The loop: N undos with the dialog open, installed once at the end.
  [`App.tsx:1104`](../../folio-designer/src/App.tsx#L1104)

- `discarding` shuts Done and Escape only while unwinding, and always comes down.
  [`TableEditor.tsx:191`](../../folio-designer/src/TableEditor.tsx#L191)

- The discard sentence is withdrawn wherever the document moves again.
  [`App.tsx:707`](../../folio-designer/src/App.tsx#L707)

**The shortcut hole (AC4, DW-368)**

- One condition, keyed on the open modal, not on a list of keys.
  [`App.tsx:2423`](../../folio-designer/src/App.tsx#L2423)

- Where the dialog receives the count and the file-busy latch.
  [`App.tsx:2854`](../../folio-designer/src/App.tsx#L2854)

**The footer**

- Cancel, Done, and the visible reason a disabled Cancel owes the author.
  [`TableEditor.tsx:533`](../../folio-designer/src/TableEditor.tsx#L533)

- A plain `div`, wrappable; no `role="group"`, deliberately.
  [`App.css:758`](../../folio-designer/src/App.css#L758)

**The mirrored bound (Go to TypeScript)**

- The mirror, cited by file and constant rather than by a line that rots.
  [`engine-protocol.ts:40`](../../folio-designer/src/engine-protocol.ts#L40)

- The tie: non-vacuity first, agreement, consumption sites, then the red-proofs.
  [`engine-bounds-mirror.test.ts:852`](../../folio-designer/src/engine-bounds-mirror.test.ts#L852)

- `wasm/engine.go` joins as the sixth Go source.
  [`engine-bounds-mirror.test.ts:74`](../../folio-designer/src/engine-bounds-mirror.test.ts#L74)

**Proofs (read last)**

- The destructive failure mode, with pre-dialog history as its positive control.
  [`TableEditor.test.tsx:793`](../../folio-designer/src/TableEditor.test.tsx#L793)

- A refused command reaches the engine and still does not count.
  [`TableEditor.test.tsx:832`](../../folio-designer/src/TableEditor.test.tsx#L832)

- Done and Escape cannot leave the engine ahead of the screen.
  [`TableEditor.test.tsx:1032`](../../folio-designer/src/TableEditor.test.tsx#L1032)

- 101 real edits prove the App's count reaches the footer.
  [`TableEditor.test.tsx:1114`](../../folio-designer/src/TableEditor.test.tsx#L1114)

- Ten shortcuts suppressed while the modal is open.
  [`App.test.tsx:1349`](../../folio-designer/src/App.test.tsx#L1349)

- Their positive controls — without these, the suppression asserts nothing.
  [`App.test.tsx:1419`](../../folio-designer/src/App.test.tsx#L1419)

- Compiled, not executed here: the footer's two buttons in a browser.
  [`table-editor.spec.ts:82`](../../folio-designer/e2e/table-editor.spec.ts#L82)

## Delivery Log

### 2026-09-10 — done

Baseline `6c643ac`. Shipped as specified: the table editor's single `Close Table Editor` became a
**`Cancel` / `Done`** pair; the application holds an integer count of the dialog's commands that
actually moved the revision, and `Cancel` issues exactly that many `undo` operations with the dialog
still open, installing the reached snapshot once and then closing. `Done` and Escape close and keep.
No engine change, no new command, no format field, no version increment — the count never crosses the
worker boundary. The second half landed as the same claim rather than an adjacent fix: the global
shortcut guard was re-keyed from *"is the target editable?"* to *"is a modal open?"*, closing `DW-368`
for the table editor, because the `N ≤ 100` bound is only sound while nothing else can commit behind
the open dialog.

**Decisions applied.** `D-14.7.1` (the `Cancel` / `Done` labels; the Cancel-issues-the-undos-it-counted
model; no engine change) and its guardrails 1–6, which the acceptance criteria carry in full — the
decision log's own word "six" is not the predicate, its AC enumeration is. `D-14.7.4` (a guard that can
only just pass is a guard that will red a true claim) drove the follow-up below. `D-14.4.1` (the
implementer commits nothing) held throughout; every commit here is the orchestrator's. `Q1 = B′` kept
`fontBrowserOpen` out of the guard, so the font browser's identical leak stands as `DW-371`.
`D-14.7.2` governs how the lint baseline is recorded below. `D-000.33` set the four-gate cadence,
`D-000.32` the wording of what did not run.

**Triage: 12 patched / 2 deferred / 4 rejected.** `review_loop_iteration` 0 — no `intent_gap`, no
`bad_spec`, no loopback. The triage tally is the build's; this close did not re-litigate it and did not
re-derive the population.

**What the review caught, and it was this story's own defect.** `Done` and Escape could tear the undo
sequence down mid-flight, and the loop returned before installing the reached snapshot — leaving the
engine k undos back while the canvas still painted the old document. Silently: the same
screen-disagrees-with-document divergence class the shortcut guard exists to close, arriving from
inside the feature that closes it. Fixed with a dedicated `discarding` flag rather than the general
`busy` flag, deliberately: `clearDocumentInteraction` never clears `busy`, so gating Escape on it would
have made the modal inescapable — a worse bug than the one being fixed.

**Gates, re-measured at this close on the tree at `cc19032`, from `folio-designer/`.** All four exit 0.

- `npx vitest run` — **76 files, 1368 tests, 0 failures.** The multiset diff against the `68aa91f`
  baseline of 1320 (**+48 new, 0 test titles removed anywhere**) is the build's measurement, carried
  here rather than re-derived: reproducing it needs a checkout of the baseline tree, which this close
  is not permitted to make. The 1368 total is this close's own.
- `npx tsc -b --force` — exit 0, **0 bytes** of output. `--force` was used; an incremental `tsc -b` can
  exit 0 without typechecking.
- `npx oxlint` — exit 0, **0 errors**. Warning baseline recorded as a per-file **SET**, never an
  integer (`D-14.7.2`), and re-measured, with no line numbers: `react(only-export-components)` in
  `src/App.tsx` ×2, `src/preview/pdf-viewer.tsx` ×2, `src/segmented-control.tsx` ×3. **No new key.**
  Nothing was changed to restore a count.
- `npm run test:e2e:compile` — exit 0. This is `tsc --noEmit` over the e2e sources. **It is not a
  browser run and is not reported as one.**

A third commit, `bd94ed4`, landed from a concurrent Story 14.8 planning agent while these gates were
running. `git diff cc19032..bd94ed4` returns `_bmad-output/planning-artifacts/epics.md` and nothing
else, so no code moved under the run and the figures above stand for both revisions.

**Did not run:** the browser suite, the Go suites, the matrix legs, `npm run build` as a gate, the
`verify:offline*` chain, and the font-host scans. CI was green on `cc19032` across all seven jobs plus
cross-target byte identity; that is CI's measurement, not this close's.

**AC4 has NO browser coverage, and must not be read as covered.** No e2e spec presses the undo, redo or
duplicate keyboard shortcut anywhere in the repo — re-verified at this close by enumerating every
`keyboard.press` / `.press(` call across all 24 e2e specs, which returns `Meta+S`/`Control+S`, `Enter`,
`Escape`, `Delete`, `Tab` and arrow keys and no `Meta+Z`, `Control+Z`, `Shift+Meta+Z`, `Control+Y` or
`Meta+D`; the same sweep returns many hits, so the absence is a real absence and not a broken search.
Two specs click the toolbar's `Undo` / `Redo` **buttons**, which is a different path and does not touch
the guard. The modal-open guard is therefore proved only in jsdom, where the ordering between the
native `window` listener and React's synthetic handlers is jsdom's and not Chromium's. Separately, the
footer's four-child layout and the legibility of the disabled-`Cancel` note are compiled here and
executed only by CI.

**Follow-up commit `cc19032`.** Two of the `Cancel discards what it counted` tests timed out on CI at
`482ea5d` while passing locally — the unit job red and the browser job green, the opposite of this
epic's usual failure, and nothing wrong with the product or the assertions. The timeout was raised on
the **whole describe block** rather than on the two that happened to fail, because four more tests in
it sit at 1.0–1.3s locally against a 5s default on a runner measured 3–5× slower under jsdom with
`userEvent` — patching only the two would have left a suite-level instance of the shape `D-14.7.4`
names. Red-proved rather than assumed: `describe` silently ignoring an unknown options object would
look identical to it working, so the block was set to 50ms and reds 21 tests with *"Test timed out in
50ms"* — the option is honoured and it reaches every test in the block.

**Deferred, all filed by the orchestrator at `482ea5d`, all unassigned.** `DW-372` — the ring-buffer
eviction the `N ≤ 100` bound rests on is asserted only as source text and no Go test executes it;
deleting one line from `appendBounded` leaves both pinning regexes matching, and a 100-step `Cancel`
would then land on the wrong document and report success. `DW-373` — the discard is proved only against
a mock that restates the engine's own rules, so a wrong reading would be wrong in both places and
agree. `DW-374` — the App-side limit guard is unreachable, so its off-by-one is pinned by nothing
executing. Carried forward unfixed and pre-existing: `DW-370` (two modal `Cancel` buttons will mean two
different things; no contract test can see it, since R1 checks spelling within a class and never
meaning) and `DW-371` (the font browser keeps the shortcut leak).

**Also disclosed, not fixed:** the ring buffer can evict the author's pre-dialog history. If a session
is already at 100 entries the dialog's own commands push the oldest out; `Cancel` stays correct because
it only ever needs the newest N, but the author's ability to undo *past* the dialog is reduced by N.
Pre-existing in `appendBounded`, deliberately not widened into.

**Housekeeping at this close.** Story 14.7's signed footer disclosure — *"⚠ THIS BAR CARRIES NO Cancel /
Done PAIR AND NO EDIT COUNTER"* — is gone, as its own handover required; verified by grep, not assumed.
Four `Suggested Review Order` anchors into `TableEditor.test.tsx` were re-derived: they were written at
`482ea5d` and `cc19032` inserted 17 lines above all of them, so `:776`, `:815`, `:1015` and `:1097`
rotted to `:793`, `:832`, `:1032` and `:1114`. Every one of the section's nineteen anchors was then
printed at `HEAD` and checked against the line it claims. Epic 14 has recorded this shape before: the
`epics.md` §14.8 block notes an anchor written as `TableEditor.tsx:427` on 2026-09-10 that rotted to
`:448` the same day when this very story moved the file by +22 lines, and says of it *re-measure it,
never quote it*. The same warning applies to the four corrected here.

Committed by the orchestrator at `482ea5d` (implementation and review patches, the register entries and
the tracker hop to `review`) and `cc19032` (the timeout). Both were already pushed before this close, so
nothing here was amended into them; this close's edits to the story file are left uncommitted for the
orchestrator.
