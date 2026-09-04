---
title: 'Story 17.5: The CONTENT box resizes from its bottom edge'
type: 'feature'
created: '2026-09-04'
status: 'backlog'
review_loop_iteration: 0
baseline_commit: '3bacfec'
context: []
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The CONTENT box can already be made taller — but only by finding the browser's own little diagonal
grip in its bottom-right corner, which sits inside the field's border looking like something that
escaped from a different application.

After this story the whole bottom edge of the box is the handle: hover it, the cursor says it can be
dragged, drag it and the box grows. The browser's grip is gone.

<intent-contract>

## Intent

**Problem:** `App.css:204` gives the CONTENT textarea `resize: vertical`, which is the *browser's*
resize affordance: a 16px diagonal grip painted by the user agent in the bottom-right corner, in the
user agent's own colours, unstyleable and unhinted. It is the only control in the inspector that is
not drawn by this application, and it reads as such (owner's report, 2026-09-04, from the running
designer). Measured: `resize` appears in `App.css` at exactly two places for real controls —
`:204` (this field) and `:329` (the raw-parameter textarea, **out of scope**); the canvas's own
handles at `:136-142` are hand-drawn, which is the house pattern this row is missing.

**Approach:** Turn the native grip off and make the field's own bottom edge the handle — a hit strip
along the full width of the box, with a resize cursor on hover and a pointer drag that changes the
box's height.

## Boundaries & Constraints

**Always:**
- **THE HEIGHT IS A VIEW STATE, NOT A DOCUMENT VALUE.** Dragging the box **sends no command**, marks
  nothing dirty, and changes no saved bytes. It is the panel's own presentation, exactly as the
  native grip was. **This is the property to assert, not merely to observe.**
- **THE DRAG MUST NOT DISTURB THE FIELD.** The author may be mid-clause with a caret in the text and
  an unflushed debounce from Story 17.1 pending. Pressing on the handle must not steal focus, must
  not move the caret, must not flush or cancel the debounce, and must not commit.
- **THE FLOOR STAYS 72px** — the four-row minimum the design asks for (`App.css:204`,
  `rows={4}` at `App.tsx:2231`). A drag may not produce a box shorter than the floor.
- **THE HANDLE IS THE BOTTOM EDGE OF THE VISIBLE BOX**, i.e. of `.property-field-prose`
  (`App.css:196`), which is the element carrying the border the author sees — not of the textarea
  inside its padding. A handle that is not on the line the eye reads as the edge is the same
  complaint again in a different place.
- **THE POINTER IS CAPTURED FOR THE DRAG**, so it survives leaving the strip; and the drag ends on
  pointerup **and on pointercancel**. The canvas handles (`App.tsx` selection/resize drags) are the
  precedent to follow rather than to reinvent.
- **`property-prose-height.test.ts` IS A GUARD TO CORRECT, NOT TO DELETE.** It asserts
  `resize: vertical` on `textarea.property-value-prose` (`:32-34`) precisely so this rule cannot
  drift onto the shared class. Its subject changes; its purpose does not. **Rewrite it to the new
  behaviour, keeping the shared-class exclusions and the non-vacuity test intact.**

**Never:** a document mutation from a resize · the native grip left on · a floor below 72px · a
drag that moves the caret or the focus · a height applied to the font-family input that shares
`.property-value-prose` · a second copy of the drag bookkeeping when the canvas already has one.

**Out of scope:**
- The raw-parameter textarea at `App.css:329` and the `textarea` base rule. Same defect, different
  surface; **name it in the report, do not fix it here.**
- Horizontal resize; the panel's width is not the author's to set.
- Persisting the height into the `.folio` file, or across a reload.
- Any other property row. Only CONTENT is multi-line.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Hover the bottom edge | Pointer over the handle strip | The resize cursor (`ns-resize`) | — |
| Drag down 40px | Box at the 72px floor | Box is 40px taller; the text reflows into it | — |
| Drag up past the floor | Box at floor, pointer moves up | **Stops at 72px**; no shorter box, no negative height | Clamped, not refused |
| Drag with a caret in the text | Caret mid-clause, field focused | Caret and focus **unmoved**; height changed | — |
| Drag with an unflushed 17.1 debounce | Text typed <200ms ago | The pending commit is **neither flushed nor cancelled** by the drag; it fires on its own schedule | — |
| Drag while a commit is IN FLIGHT | A command awaited, more text typed | The drag touches neither `queuedProse` nor `toldEngine`; the queued send still drains | — |
| Release outside the panel | Pointer leaves the window mid-drag | The drag ends cleanly; no stuck drag state | `pointercancel` ends it too |
| Native grip | Any state | **Absent** — `resize` is `none` on the textarea | — |
| The font-family input | Shares `.property-value-prose` | **Unchanged height, no handle** — the bug `property-prose-height.test.ts` exists to prevent | — |
| Selection changes while tall | Author selects another text component | **Defined and asserted** — see Design Notes | — |
| Keyboard only | No pointer | **No regression**: the native grip was not keyboard-reachable either. State this; do not silently drop it | — |

</intent-contract>

## Code Map

**Anchors re-measured at `3bacfec`. The previous set was taken at `c13864c`, before Story 17.1 landed,
and every `App.tsx` line below had moved. Re-verify anyway — this file has moved under three stories in
one day.**

- `App.css:204` — `textarea.property-value-prose { min-height: 72px; resize: vertical; }`. **The seam.**
- `App.css:196` — `.property-field-prose`, the bordered box the handle belongs to.
- `App.css:171` — `.property-field`, the shared row box; `position` is **not** set on it, and a handle
  positioned against the prose row will need one. **Scope that to the prose variant.**
- `App.css:136-142` — `.resize-handle` and the eight `.selection-handle-*` rules: the house idiom for
  a hand-drawn handle, its hit target and its cursor. `:140-142` already use `ns-resize`.
- `App.tsx:2288` — the `PropertyDraft` return and its `property-field-prose` wrapper; `:2289` the
  `<textarea … rows={4} …>` inside it, **whose `onChange` is now the debounce's only typing entry point**.
- `App.tsx:1868` `writeDraft`, `:1905` `holdDraft`, `:1941` `scheduleProseCommit`, `:2263` `pasteProse`
  (`:2279` its schedule call) — the Story 17.1 machinery **the drag must not touch**.
- `src/property-prose-height.test.ts:33` — the `resize: vertical` assertion to rewrite; `:22-27` the
  shared-class exclusions, `:38-43` the non-vacuity pairing and `:45-47` the body-type assertion, **all
  three to keep**.

## Tasks & Acceptance

**Execution:**
- [ ] `App.css` — `resize: none` on `textarea.property-value-prose`; keep the 72px floor; add the
      handle strip's rule (full width, `cursor: ns-resize`, a hit height in the 6–8px band, drawn as
      the box's own edge rather than as a new visual element).
- [ ] `App.tsx` — the handle element on the prose row only, with a pointer-capture drag that sets the
      box height from the pointer delta, clamped at the floor. `onPointerDown` **must**
      `preventDefault()` so focus and caret stay put.
- [ ] Decide and assert what happens to the height when the selection changes — the two honest
      options are "the author's height is the panel's until the app closes" and "each selection starts
      at the floor". Pick one, state why in Design Notes, and test it.
- [ ] Rewrite `property-prose-height.test.ts` to the new contract; **do not delete it**.
- [ ] Tests — one per matrix row (**11 rows**), plus the two the matrix cannot reach on its own: that
      a drag sends **no** command (assert against the command spy, with a positive control that the
      same field's typing does send one), and that the 17.1 debounce still fires after a drag.
- [ ] A browser run: hover the edge and photograph the cursor, drag the box taller mid-typing, and show
      the caret where it was left and the canvas still following the text.

**Acceptance Criteria:**
- Given the CONTENT box, when the pointer is over its bottom edge, then the cursor is `ns-resize` and
  no browser grip is painted anywhere in the box.
- Given a drag downward, when it is released, then the box is taller by the pointer's travel.
- Given a drag upward past the floor, when it is released, then the box is exactly 72px.
- Given a caret in the middle of the text, when the box is dragged, then **no command is sent** and the
  caret is where it was.

## Design Notes

**The risk here is not the drag; it is what the drag is standing on.** This row is the one control
Story 17.1 put a timer under: text typed into it commits ~200 ms after typing stops. A resize that
steals focus fires `blur`, and `blur` on this field is a commit path. **`preventDefault()` on
pointerdown is therefore not polish — it is what keeps a resize from becoming a write.**

**THAT CONCLUSION IS UNCHANGED; THE REASON BEHIND IT IS NOT, AND THIS PARAGRAPH USED TO STATE THE WRONG
ONE.** It was written while 17.1 was still in flight and said that `submit` returns early while a
command is in flight and that the draft reconciles from the engine's echo. **Those were 17.1's two
DEFECTS, and 17.1 fixed them** (`3bacfec`). Blur now QUEUES through `queuedProse` when its commit would
be swallowed, and reconciliation keys off `toldEngine` — what the engine was last told — rather than a
`committed` that goes stale mid-flight. So the danger a focus-stealing handle poses is no longer a LOST
EDIT; it is simply an unasked-for WRITE. Smaller, still disqualifying, and a different thing to assert:
the test to write is that a drag sends no command, not that a drag loses no text.

The second trap is the shared class. `.property-value-prose` is worn by both the textarea and the
font-family combobox, and the last time a height contract was written on the shared class the
combobox rendered at twice its size. Anything this story adds is scoped to the element or to
`.property-field-prose`, never to the shared class — which is exactly what
`property-prose-height.test.ts` was built to catch, and why it is rewritten rather than removed.

## Verification

**Commands** — one per line, **exit codes read from `$?` immediately, never through a pipe**; zsh, so
`${PIPESTATUS[0]}` is wrong.
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts`'s corpus
  scan (DW-152); **match it by NAME, not count.**
- `npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`. **No Go is touched; say so rather than running it.**

**A BROWSER RUN.** `chromium-1217` via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (DW-180). JSDOM has no
layout, so the drag's arithmetic is unit-testable but **the cursor and the absent grip are only real in
a browser** — photograph both.

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**11 rows**);
state the population beside every zero, pair every absence claim with a positive control; **a comment is
not a measurement**; **use `/usr/bin/grep`** — recursive `grep` returns false zeros here.
