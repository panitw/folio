---
title: 'Story 17.5: The CONTENT box resizes from its bottom edge'
type: 'feature'
created: '2026-09-04'
status: 'in-review'
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


**RE-VERIFIED AT `1a56007` (the build baseline; `1a56007` is the only commit after `3bacfec`, and it
is artifact-only). Every anchor above is exact.** Two corrections and eleven measured facts follow.

- **Correction 1.** The Intent block's `rows={4}` at `App.tsx:2231` is STALE — `:2231` is a comment
  line inside the Escape/keyDown note. `rows={4}` is at `App.tsx:2289`, as the Code Map itself says.
  The frozen block is read-only, so this is recorded rather than fixed.
- **Correction 2.** `App.css:329` is the `textarea` BASE rule, not a rule for the raw-parameter
  textarea specifically. It still governs the CONTENT field: `textarea { resize: vertical }` (0,0,1)
  is outranked by `textarea.property-value-prose` (0,1,1), so `resize: none` at `:204` wins without
  touching `:329`. jsdom applies no stylesheet — that specificity is read by hand and guarded only by
  the source-text assertions in `property-prose-height.test.ts`.

- **NO DOM MEASUREMENT IS AVAILABLE.** `canvas-authority-contract.test.ts:18-45` prohibits
  `getBoundingClientRect`, `getClientRects`, `offsetWidth/Height/Left/Top/Parent`,
  `clientWidth/Height/Left/Top`, `scrollWidth/Height/Left/Top`, `getComputedStyle` and
  `ResizeObserver` across production, unit-test AND e2e sources. `clientY` is NOT in that list. So the
  height is component state, the drag is pure `clientY` arithmetic (the canvas drag's own shape,
  `App.tsx:2830`), and tests read the height off the INLINE style — never `getComputedStyle`.
- **THE RESTING BOX IS EXACTLY 72px, so a drag may start its arithmetic there.** `--type-body-em` is
  `400 11px/1.3` (`tokens.css:17`) → `rows={4}` gives an intrinsic 4 x 14.3 = **57.2px**, and
  `.property-value` sets `padding: 0; border: 0` (`App.css:181`). 57.2 < 72, so the `min-height`
  floor governs today and there is no first-drag jump. (72px is ~5 of those rows; "four-row minimum"
  is the design's number, not an arithmetic consequence of `rows={4}`.)
- **TEST PLACEMENT TRAP — the new block goes BEFORE the `// STORY 17.1:` header (`App.test.tsx:3628`),
  not at the end of the file.** `App.test.tsx:4229` reads the 17.1 header's spelled count back off the
  source and compares it with `(source.slice(start).match(/\n {2}it\(/g)).length`, where `start` is
  the index of that block's own `describe(` — and the 17.1 block runs to END OF FILE. A describe
  appended after it has every one of its `it(` counted into 17.1's total and reds that test. Two
  further constraints on the new block, both from the same test: it must not contain that describe
  title as a literal string, and it must not contain a comment matching `// WORD tests.` in capitals —
  that regex takes the LAST such match before the 17.1 describe as the declared count.
- **The 17.1 harness is block-scoped; a new block needs its own copies.** `proseEngine` (`:3672`,
  returns `{ sent: string[], request, hold, release, engineHolds }` — the only double that echoes a
  canvas back), `openEditor` (`:3718`), `type` (`:3726`), `elapse`/`settle` (`:3716-3717`),
  `changes(wire)` (`:3709`). The file-level engine double is `:154`. `PROSE_COMMIT_DEBOUNCE_MS` is
  exported from `./App` (`App.tsx:1667`, = 200). Fake timers: `vi.useFakeTimers()` in `beforeEach`,
  and **never `waitFor`** in a fake-timer block — it advances timers itself.
- **The CONTENT field is `screen.getByRole('textbox', { name: 'Text' })`** — the accessible name is
  `Text`, not `Content`. The font-family combobox is the other wearer of `.property-value-prose`.
- **Drag idiom (`App.test.tsx:747-767`): every pointer event fires on the SAME element.**
  `setPointerCapture?.()` (`App.tsx:2821`) is optional-called precisely because jsdom leaves it
  undefined. `fireEvent.pointerDown(...)` returns `false` when `preventDefault()` was called — that
  return value is the falsifiable proof of the preventDefault, **because jsdom does not move focus on
  pointerdown at all**, which makes "activeElement is unchanged" vacuous in jsdom. Assert both, and
  say which one is the real measurement; the focus and the caret are only proved in the browser run.
- **Selection-change idiom:** `fireEvent.click(screen.getByLabelText(/^text component e1/))` — a
  painted component's accessible name carries its text, so the anchored regex is required.
- **`documentGeneration` bumps ONLY in `setCurrentSnapshot(..., clearDocumentInteraction = true)`
  (`App.tsx:1227`)** — a document replacement, which also clears the selection. A property commit does
  not bump it. So the `documentGeneration:selection` key (`App.tsx:1478`) changes on a selection
  change or a new/opened document, and **never on the author's own typing**.
- **Handles in this codebase are `<span aria-hidden="true">` carrying the four pointer handlers**
  (`App.tsx:2846`); only the `se` canvas handle is a `<button>`, and a button would add a keyboard
  tab stop the matrix's last row says must not appear. Tests reach an aria-hidden handle with
  `view.container.querySelector(...)` — 37 existing uses in `App.test.tsx`.
- **`design-contract.test.ts:84` forbids colour literals and non-token `border-radius` in `App.css`**;
  `:93` requires the accent grammar. A transparent hit strip with a cursor trips neither.
- **BASELINE MEASURED AT `1a56007`:** `cd folio-designer && npm test` → 55 files, **733 tests, 1
  failed**. The one failure is `canvas-authority-contract.test.ts > canvas projection authority
  contract > scans a non-vacuous production, unit-test, and e2e corpus for browser measurement
  authority`, whose violation array is exactly one entry, the pre-existing `getComputedStyle` use in
  `e2e/e9-5-border-no-ink.spec.ts`. **That array must not grow**; the assertion is `toEqual([])`, so
  it reds identically either way and the array's CONTENTS are the only thing that distinguishes a
  regression from the standing red.

## Tasks & Acceptance

**Execution:**
- [x] `App.css` — `resize: none` on `textarea.property-value-prose`; keep the 72px floor; add the
      handle strip's rule (full width, `cursor: ns-resize`, a hit height in the 6–8px band, drawn as
      the box's own edge rather than as a new visual element).
- [x] `App.tsx` — the handle element on the prose row only, with a pointer-capture drag that sets the
      box height from the pointer delta, clamped at the floor. `onPointerDown` **must**
      `preventDefault()` so focus and caret stay put.
- [x] Decide and assert what happens to the height when the selection changes — the two honest
      options are "the author's height is the panel's until the app closes" and "each selection starts
      at the floor". Pick one, state why in Design Notes, and test it.
- [x] Rewrite `property-prose-height.test.ts` to the new contract; **do not delete it**.
- [x] Tests — one per matrix row (**11 rows**), plus the two the matrix cannot reach on its own: that
      a drag sends **no** command (assert against the command spy, with a positive control that the
      same field's typing does send one), and that the 17.1 debounce still fires after a drag.
- [x] A browser run: hover the edge and photograph the cursor, drag the box taller mid-typing, and show
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

**THE DECISION THE SPEC LEFT TO THE BUILD: an author's height does NOT survive a change of
selection. Each selection starts at the 72px floor.** Three reasons, in the order that decided it.

1. **It is what happens today, so the story changes the affordance and nothing else.** The native
   grip's height is user-agent state on the DOM element instance. `ComponentProperties` is keyed
   `${documentGenerationValue}:${selected.join(',')}` (`App.tsx:1478`), so selecting another
   component unmounts that whole subtree and mounts a *fresh* textarea — the grip's height cannot
   survive it. Making the new handle's height persist would be a second, unasked-for change smuggled
   in beside the one the owner asked for.
2. **The existing key gives it for free, and the alternative costs a second copy of the drag
   bookkeeping.** A `useState` height inside `PropertyDraft` resets on a remount that already exists
   and is already documented (`App.tsx:2060-2067` states it outright: "a change of selection or
   document REMOUNTS every `PropertyDraft` beneath it"). Persisting instead means lifting per-row
   view state above the keyed boundary and threading it through `ComponentProperties` — which the
   Boundaries forbid ("never … a second copy of the drag bookkeeping") — and it would apply one
   component's authored height to a different component's content, which is arbitrary rather than
   helpful.
3. **The failure mode that would have disqualified this option is ruled out by measurement, not by
   assumption.** If `documentGeneration` bumped on every commit, "starts at the floor" would in fact
   mean "the height evaporates while you type", which is indefensible. It does not: the counter is
   incremented only inside `setCurrentSnapshot(…, clearDocumentInteraction = true)` (`App.tsx:1227`),
   a document replacement, which also clears the selection. A property commit leaves the key alone.

**Tested as a pair, because the first half alone is not falsifiable.** (a) Drag taller, select
another text component, assert the CONTENT box is back at 72px. (b) Drag taller, type, let the 200ms
debounce fire and the command settle, assert the height is **unchanged** — this is what proves (a)
is a property of the *selection* key rather than of any state change at all.

**REVIEW PATCHES (iteration 1). Eight findings, all applied; two were MEASURED DEFECTS in the
first implementation, not opinions.** (1) A RIGHT-BUTTON press on the strip started a resize —
Chromium 1217, 122px after a 50px move, with the context menu then blocking the page mid-gesture.
`beginProseResize` now returns unless `event.button === 0`, before `preventDefault`, so a right-press
is still the browser's. (2) A missed `pointerup` left the drag live, so the next bare hover across the
strip resized the box with nothing held; a `event.buttons === 0` check now ends it on the first such
move. (3) The drag ref carried no pointer id, so a second pointer's press rebased the anchor and its
moves were read as the first pointer's travel; the ref now carries `pointerId`, a press from a
different pointer is refused while a gesture is live, and move/end both require the id to match.
(4) The anchor is now recorded BEFORE `setPointerCapture` is requested — that call is specified to
throw `NotFoundError` for an inactive pointer, and the old order would have let such a throw swallow
the press silently. (5) Pointer capture had ZERO executed coverage despite the Boundaries requiring it
in terms; it is now asserted through a `vi.fn()` stub, jsdom leaving `setPointerCapture` undefined
being why the production call is optional-called. (6) `touch-action: none` was asserted by nothing,
and dropping it makes a touch press a pan that fires `pointercancel`. (7) Three source-text guards did
not pin what they claimed: the "full-width strip on the bottom edge" row asserted only the cursor and
the two sides, so a strip pinned to the TOP stayed green; the "paints nothing" guard was a two-item
denylist that permitted `border-top`, `box-shadow` and `outline` and could not see a `::after` rule at
all, and is now an ALLOWLIST of the eight properties the rule may declare plus a no-pseudo-element
assertion; and the base-`textarea` row passed unchanged when `resize: vertical !important` was added
to the very rule it exists to be outranked by. (8) A false comment in `App.css` claimed the handle's
7px hit height was the canvas handles' idiom; measured, `.selection-handle` is 12px and
`.resize-handle` 24px — the CURSOR is shared, the size is this row's own 6-8px band. Every guard added
or changed was red-proved by mutating the production line.

**Two recording notes.** The build baseline is `1a56007`; the frontmatter keeps `baseline_commit:
3bacfec` because the workflow forbids overwriting an existing value, and `3bacfec` is the commit the
Code Map was measured at. The only commit between them is `1a56007` itself, which is artifact-only.
Separately: `App.css:329` — the `textarea` base rule, with the same `resize: vertical` defect on the
raw-parameter textarea — is **named and left alone**, as Out of scope directs.

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
